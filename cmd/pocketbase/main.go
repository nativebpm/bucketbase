package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"regexp"

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
	s3Config := storage.GetS3Config()

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
			cmd := exec.Command(os.Args[0], "superuser", "upsert", config.PocketbaseAdminEmail, config.PocketbaseAdminPassword)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				slog.Error("Superuser upsert failed", "error", err)
			}
			return e.Next()
		})
	}

	if err := app.Start(); err != nil {
		slog.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
}
