package pocketbase

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/nativebpm/pocketstream/internal/storage"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func New() *pocketbase.PocketBase {
	pocketbaseConfig := pocketbase.Config{}
	config := GetConfig()

	if config.PocketbaseEncryptionKey != "" {
		if !ValidateEncryptionKey(config.PocketbaseEncryptionKey) {
			slog.Error("POCKETBASE_ENCRYPTION_KEY must be a 32-character hexadecimal string (generated with 'openssl rand -hex 16')")
			os.Exit(1)
		}
		pocketbaseConfig.DefaultEncryptionEnv = "POCKETBASE_ENCRYPTION_KEY"
	}

	app := pocketbase.NewWithConfig(pocketbaseConfig)

	// Bootstrap hook: run default bootstrap first (DB init, migration, settings reload),
	// then apply and persist environment-configured overrides (S3, RateLimits).
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		settings := app.Settings()
		var needsSave bool

		if config.RateLimitEnabled == "true" {
			settings.RateLimits.Enabled = true
			if config.RateLimitRules != "" {
				var rules []core.RateLimitRule
				if err := json.Unmarshal([]byte(config.RateLimitRules), &rules); err == nil {
					settings.RateLimits.Rules = rules
				} else {
					slog.Warn("Failed to parse RATE_LIMIT_RULES, using defaults", "error", err)
				}
			}
			needsSave = true
		}

		s3Config := storage.GetS3Config()
		if s3Config.Enabled {
			if err := storage.MakeBucket(); err != nil {
				slog.Error("Failed to initialize S3 storage buckets", "error", err)
				return err
			}

			settings.S3.Enabled = true
			settings.S3.Bucket = s3Config.Bucket
			settings.S3.Region = s3Config.Region
			if s3Config.Endpoint != nil && s3Config.Endpoint.Host != "" {
				settings.S3.Endpoint = s3Config.Endpoint.String()
				settings.S3.ForcePathStyle = true
			} else {
				settings.S3.Endpoint = ""
				settings.S3.ForcePathStyle = false
			}
			settings.S3.AccessKey = s3Config.AccessKeyID
			settings.S3.Secret = s3Config.SecretAccessKey
			needsSave = true
		}

		if needsSave {
			if err := app.Save(settings); err != nil {
				slog.Error("Failed to persist bootstrap settings", "error", err)
				return err
			}
			slog.Info("Successfully applied and persisted bootstrap settings", "s3_enabled", s3Config.Enabled)
		}

		return nil
	})

	return app
}
