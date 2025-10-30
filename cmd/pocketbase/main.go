package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/nativebpm/pocketstream/internal/storage"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Config struct {
	Profile                 string
	PocketbaseAdminEmail    string
	PocketbaseAdminPassword string
	GOMEMLIMIT              string
	RateLimitEnabled        string
	RateLimitRules          string
	PocketbaseEncryptionKey string
}

func getConfig() Config {
	return Config{
		Profile:                 os.Getenv("PROFILE"),
		PocketbaseAdminEmail:    os.Getenv("POCKETBASE_ADMIN_EMAIL"),
		PocketbaseAdminPassword: os.Getenv("POCKETBASE_ADMIN_PASSWORD"),
		GOMEMLIMIT:              os.Getenv("GOMEMLIMIT"),
		RateLimitEnabled:        os.Getenv("RATE_LIMIT_ENABLED"),
		RateLimitRules:          os.Getenv("RATE_LIMIT_RULES"),
		PocketbaseEncryptionKey: os.Getenv("POCKETBASE_ENCRYPTION_KEY"),
	}
}

func validateEncryptionKey(key string) bool {
	if len(key) != 32 {
		return false
	}
	matched, err := regexp.MatchString(`^[0-9a-fA-F]{32}$`, key)
	return err == nil && matched
}

func forceCheckpoint() error {
	// Force WAL checkpoint to ensure schema changes are written to main database file
	if _, err := os.Stat("/pb_data/data.db"); os.IsNotExist(err) {
		return nil // Database doesn't exist yet
	}

	cmd := exec.Command("sqlite3", "/pb_data/data.db", "PRAGMA wal_checkpoint;")
	return cmd.Run()
}

func main() {
	pocketbaseConfig := pocketbase.Config{}
	config := getConfig()

	if config.PocketbaseEncryptionKey != "" {
		if !validateEncryptionKey(config.PocketbaseEncryptionKey) {
			slog.Error("POCKETBASE_ENCRYPTION_KEY must be a 32-character hexadecimal string (generated with 'openssl rand -hex 16')")
			os.Exit(1)
		}
		pocketbaseConfig.DefaultEncryptionEnv = "POCKETBASE_ENCRYPTION_KEY"
	}

	app := pocketbase.NewWithConfig(pocketbaseConfig)

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		settings := app.Settings()
		settings.RateLimits.Enabled = config.RateLimitEnabled == "true"
		if settings.RateLimits.Enabled {
			if config.RateLimitRules != "" {
				var rules []core.RateLimitRule
				if err := json.Unmarshal([]byte(config.RateLimitRules), &rules); err != nil {
					slog.Warn("Failed to parse RATE_LIMIT_RULES, using defaults", "error", err)
				} else {
					settings.RateLimits.Rules = rules
				}
			}
		}
		return e.Next()
	})

	s3Config := storage.GetS3Config()
	if s3Config.Enabled {

		storage.MakeBucket()

		app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
			settings := app.Settings()
			settings.S3.Enabled = true
			settings.S3.Bucket = s3Config.Bucket
			settings.S3.Region = s3Config.Region
			settings.S3.Endpoint = s3Config.Endpoint.String()
			settings.S3.AccessKey = s3Config.AccessKeyID
			settings.S3.Secret = s3Config.SecretAccessKey
			settings.S3.ForcePathStyle = true
			return e.Next()
		})
	}

	if config.Profile == "docker" {
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			// Add checkpoint endpoint for schema changes
			e.Router.POST("/api/checkpoint", func(c *core.RequestEvent) error {
				if err := forceCheckpoint(); err != nil {
					return c.JSON(500, map[string]string{"error": err.Error()})
				}
				return c.JSON(200, map[string]string{"status": "checkpoint completed"})
			})

			// Litestream optimizations
			if _, err := app.DB().NewQuery("PRAGMA busy_timeout = 5000").Execute(); err != nil {
				slog.Warn("Failed to set busy_timeout", "error", err)
			}
			if _, err := app.DB().NewQuery("PRAGMA synchronous = NORMAL").Execute(); err != nil {
				slog.Warn("Failed to set synchronous", "error", err)
			}
			if _, err := app.DB().NewQuery("PRAGMA wal_autocheckpoint = 0").Execute(); err != nil {
				slog.Warn("Failed to disable wal_autocheckpoint", "error", err)
			}

			cmd := exec.Command("/pocketbase", "superuser", "upsert", config.PocketbaseAdminEmail, config.PocketbaseAdminPassword)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				slog.Error("Superuser upsert failed", "error", err)
			}
			return e.Next()
		})

		// Add automatic checkpoint after each request that modifies data
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			// This will run after the server starts
			go func() {
				// Simple approach: checkpoint every 30 seconds
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					if err := forceCheckpoint(); err != nil {
						slog.Warn("Failed to periodic checkpoint", "error", err)
					} else {
						slog.Debug("Periodic checkpoint completed")
					}
				}
			}()
			return e.Next()
		})

		// Add hooks for automatic checkpoint on data changes
		app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record creation
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record create", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record create", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record update
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record update", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record update", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record deletion
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record delete", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record delete", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnCollectionCreate().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection creation (schema change)
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection create", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection create", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})

		app.OnCollectionUpdate().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection update (schema change)
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection update", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection update", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})

		app.OnCollectionDelete().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection deletion (schema change)
			go func() {
				if err := forceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection delete", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection delete", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})
	}

	if err := app.Start(); err != nil {
		slog.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
}
