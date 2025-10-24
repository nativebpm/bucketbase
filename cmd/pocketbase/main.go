package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/exec"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "fs"
	}
	profile := os.Getenv("APP_PROFILE")
	if profile == "" {
		profile = "local"
	}

	log.Printf("Parsed mode=%s, profile=%s", mode, profile)

	app := pocketbase.New()
	appMode(mode, app)
	appProfile(profile, app)

	if err := app.Start(); err != nil {
		log.Fatal("PocketBase serve failed:", err)
	}
}

func appProfile(profile string, app *pocketbase.PocketBase) {
	if profile == "docker" {
		email := os.Getenv("POCKETBASE_ADMIN_EMAIL")
		if email == "" {
			email = "admin@example.com"
		}

		password := os.Getenv("POCKETBASE_ADMIN_PASSWORD")
		if password == "" {
			password = "admin123"
		}

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
}

func appMode(mode string, app *pocketbase.PocketBase) {
	if mode == "s3" {

		minioURL, err := url.Parse(os.Getenv("MINIO_ENDPOINT"))
		if err != nil {
			log.Fatalln(err)
		}

		makeBucket(minioURL)

		if minioURL.Scheme == "" {
			minioURL.Scheme = "http"
		}

		app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
			settings := app.Settings()
			settings.S3.Enabled = true
			settings.S3.Bucket = os.Getenv("MINIO_BUCKET")
			settings.S3.Region = os.Getenv("MINIO_REGION")
			settings.S3.Endpoint = minioURL.String()
			settings.S3.AccessKey = os.Getenv("MINIO_ACCESS_KEY")
			settings.S3.Secret = os.Getenv("MINIO_SECRET_KEY")
			settings.S3.ForcePathStyle = true
			return e.Next()
		})
	}
}

func makeBucket(minioURL *url.URL) {
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	buckets := []string{
		os.Getenv("MINIO_BUCKET"),
		os.Getenv("LITESTREAM_BUCKET"),
	}

	ctx := context.Background()

	// Initialize minio client object.
	minioClient, err := minio.New(minioURL.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	for _, bucketName := range buckets {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)
			exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
			if errBucketExists == nil && exists {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Successfully created %s\n", bucketName)
		}
	}

}
