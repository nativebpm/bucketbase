package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
			// Create or update superuser (shell command: pocketbase superuser upsert <email> <password>)
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

		s3URL, err := url.Parse(os.Getenv("S3_ENDPOINT"))
		if err != nil {
			log.Fatalln(err)
		}

		useSSL := os.Getenv("S3_USE_SSL") == "true"

		makeBucket(s3URL)

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
}

func makeBucket(s3URL *url.URL) {
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	buckets := []string{
		os.Getenv("S3_BUCKET"),
		os.Getenv("LITESTREAM_BUCKET"),
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalln(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3URL.String())
		o.UsePathStyle = true
	})

	for _, bucketName := range buckets {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucketName,
		})
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)
			_, errHead := client.HeadBucket(ctx, &s3.HeadBucketInput{
				Bucket: &bucketName,
			})
			if errHead == nil {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		} else {
			log.Printf("Successfully created %s\n", bucketName)
		}
	}

}
