package storage

import (
	"context"
	"log/slog"
	"net/url"
	"os"

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
	s3Endpoint, err := url.Parse(os.Getenv("S3_ENDPOINT"))
	if err != nil {
		slog.Error("Failed to parse S3 endpoint", "error", err)
		os.Exit(1)
	}

	useSSL := os.Getenv("S3_USE_SSL") == "true"

	if useSSL {
		s3Endpoint.Scheme = "https"
	} else {
		s3Endpoint.Scheme = "http"
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	return S3Config{
		Enabled:         os.Getenv("S3_ENABLED") == "true",
		UseSSL:          useSSL,
		Bucket:          os.Getenv("S3_BUCKET"),
		Region:          region,
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
		Buckets: []string{
			os.Getenv("S3_BUCKET"),
			os.Getenv("LITESTREAM_BUCKET"),
		},
		Endpoint: s3Endpoint,
	}
}

func MakeBucket() {
	cfg := GetS3Config()

	ctx := context.Background()

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		os.Exit(1)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint.String())
		o.UsePathStyle = true
	})

	for _, bucketName := range cfg.Buckets {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucketName,
		})
		if err != nil {
			_, errHead := client.HeadBucket(ctx, &s3.HeadBucketInput{
				Bucket: &bucketName,
			})
			if errHead == nil {
				slog.Info("Bucket already exists", "bucket", bucketName)
			} else {
				slog.Error("Failed to create bucket", "bucket", bucketName, "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("Bucket created successfully", "bucket", bucketName)
		}
	}

}
