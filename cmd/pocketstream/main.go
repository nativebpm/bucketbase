package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/nativebpm/pocketstream/internal/litestream"
	"github.com/nativebpm/pocketstream/internal/storage"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	s3Enabled := os.Getenv("S3_ENABLED") == "true"
	profile := os.Getenv("PROFILE")

	if s3Enabled {
		s3URL, err := url.Parse(os.Getenv("S3_ENDPOINT"))
		if err != nil {
			log.Fatalln(err)
		}

		useSSL := os.Getenv("S3_USE_SSL") == "true"

		storage.MakeBucket(s3URL)

		if useSSL {
			s3URL.Scheme = "https"
		} else {
			s3URL.Scheme = "http"
		}

		app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
			settings := app.Settings()
			settings.S3.Enabled = true
			settings.S3.Bucket = os.Getenv("S3_BUCKET")
			settings.S3.Region = os.Getenv("S3_REGION")
			settings.S3.Endpoint = s3URL.String()
			settings.S3.AccessKey = os.Getenv("S3_ACCESS_KEY")
			settings.S3.Secret = os.Getenv("S3_SECRET_KEY")
			settings.S3.ForcePathStyle = true
			return e.Next()
		})
	}

	if profile == "docker" {
		email := os.Getenv("POCKETBASE_ADMIN_EMAIL")

		password := os.Getenv("POCKETBASE_ADMIN_PASSWORD")

		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			cmd := exec.Command(os.Args[0], "superuser", "upsert", email, password)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("Superuser upsert failed: %v", err)
			}
			return e.Next()
		})
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		collections, err := app.FindAllCollections()
		if err != nil {
			log.Printf("Failed to get collections: %v", err)
		} else {
			for _, collection := range collections {
				if strings.HasPrefix(collection.Name, "_") {
					continue
				}
				app.OnRecordCreateRequest(collection.Name).BindFunc(checkReplication)
				app.OnRecordUpdateRequest(collection.Name).BindFunc(checkReplication)
				app.OnRecordDeleteRequest(collection.Name).BindFunc(checkReplication)
			}
		}
		return e.Next()
	})

	if err := litestream.ReplicateDatabase(); err != nil {
		log.Fatal("Database replication failed:", err)
	}

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func checkReplication(e *core.RecordRequestEvent) error {
	if err := litestream.CheckReplicationHealth(); err != nil {
		log.Printf("[REPLICATION] Write operation blocked: %v", err)
		return e.Error(http.StatusServiceUnavailable, "Replication inactive, write operations disabled", nil)
	}
	return e.Next()
}
