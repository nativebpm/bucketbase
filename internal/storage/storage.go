package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Enabled         bool
	UseSSL          bool
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Buckets         []string
	Endpoint        *url.URL
}

func GetS3Config() S3Config {
	rawEndpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	useSSL := os.Getenv("S3_USE_SSL") == "true"
	var s3Endpoint *url.URL

	if rawEndpoint != "" {
		if !strings.HasPrefix(rawEndpoint, "http://") && !strings.HasPrefix(rawEndpoint, "https://") {
			if useSSL {
				rawEndpoint = "https://" + rawEndpoint
			} else {
				rawEndpoint = "http://" + rawEndpoint
			}
		}

		var err error
		s3Endpoint, err = url.Parse(rawEndpoint)
		if err != nil && os.Getenv("S3_ENABLED") == "true" {
			slog.Error("Failed to parse S3 endpoint", "error", err, "endpoint", rawEndpoint)
		} else if s3Endpoint != nil {
			if useSSL {
				s3Endpoint.Scheme = "https"
			} else {
				s3Endpoint.Scheme = "http"
			}
		}
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Filter, clean, and deduplicate bucket names
	var buckets []string
	seen := make(map[string]bool)
	for _, b := range []string{os.Getenv("S3_BUCKET"), os.Getenv("LITESTREAM_BUCKET")} {
		b = strings.TrimSpace(b)
		b = strings.TrimPrefix(b, "s3://")
		b = strings.Trim(b, "/")
		if b != "" && !seen[b] {
			buckets = append(buckets, b)
			seen[b] = true
		}
	}

	cleanBucket := strings.Trim(strings.TrimPrefix(strings.TrimSpace(os.Getenv("S3_BUCKET")), "s3://"), "/")

	return S3Config{
		Enabled:         os.Getenv("S3_ENABLED") == "true",
		UseSSL:          useSSL,
		Bucket:          cleanBucket,
		Region:          region,
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
		Buckets:         buckets,
		Endpoint:        s3Endpoint,
	}
}

func MakeBucket() error {
	cfg := GetS3Config()

	if !cfg.Enabled || len(cfg.Buckets) == 0 {
		return nil
	}

	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != nil && cfg.Endpoint.Host != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint.String())
			o.UsePathStyle = true
		}
	})

	for _, bucketName := range cfg.Buckets {
		var success bool
		for i := 0; i < 15; i++ {
			_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: &bucketName,
			})
			if err != nil {
				_, errHead := client.HeadBucket(ctx, &s3.HeadBucketInput{
					Bucket: &bucketName,
				})
				if errHead == nil {
					slog.Info("Bucket already exists", "bucket", bucketName)
					success = true
					break
				} else {
					slog.Warn("Failed to create bucket, retrying...", "bucket", bucketName, "error", err, "attempt", i+1)
					time.Sleep(2 * time.Second)
				}
			} else {
				slog.Info("Bucket created successfully", "bucket", bucketName)
				success = true
				break
			}
		}
		if !success {
			slog.Error("Failed to create bucket after retries", "bucket", bucketName)
			return fmt.Errorf("failed to create bucket %q after retries", bucketName)
		}
	}

	return nil
}

