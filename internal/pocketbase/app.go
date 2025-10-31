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

	// Bootstrap hooks
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

	return app
}
