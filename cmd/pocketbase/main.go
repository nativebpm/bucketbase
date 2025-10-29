package main

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/nativebpm/pocketstream/internal/storage"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type Config struct {
	Profile                 string
	PocketbaseAdminEmail    string
	PocketbaseAdminPassword string
}

func getConfig() Config {
	return Config{
		Profile:                 os.Getenv("PROFILE"),
		PocketbaseAdminEmail:    os.Getenv("POCKETBASE_ADMIN_EMAIL"),
		PocketbaseAdminPassword: os.Getenv("POCKETBASE_ADMIN_PASSWORD"),
	}
}

func main() {
	app := pocketbase.New()

	config := getConfig()
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
