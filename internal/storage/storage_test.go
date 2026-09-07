package storage_test

import (
	"os"
	"testing"

	"github.com/nativebpm/pocketstream/internal/storage"
)

func TestGetS3Config_EndpointNormalization(t *testing.T) {
	tests := []struct {
		name             string
		envEndpoint      string
		envUseSSL        string
		envEnabled       string
		expectedNil      bool
		expectedScheme   string
		expectedHost     string
		expectedEndpoint string
	}{
		{
			name:        "empty endpoint (native AWS S3)",
			envEndpoint: "",
			envUseSSL:   "",
			envEnabled:  "true",
			expectedNil: true,
		},
		{
			name:             "endpoint without scheme (http default)",
			envEndpoint:      "minio:9000",
			envUseSSL:        "false",
			envEnabled:       "true",
			expectedNil:      false,
			expectedScheme:   "http",
			expectedHost:     "minio:9000",
			expectedEndpoint: "http://minio:9000",
		},
		{
			name:             "endpoint without scheme with SSL",
			envEndpoint:      "s3.example.com",
			envUseSSL:        "true",
			envEnabled:       "true",
			expectedNil:      false,
			expectedScheme:   "https",
			expectedHost:     "s3.example.com",
			expectedEndpoint: "https://s3.example.com",
		},
		{
			name:             "endpoint with http scheme",
			envEndpoint:      "http://localhost:8334",
			envUseSSL:        "false",
			envEnabled:       "true",
			expectedNil:      false,
			expectedScheme:   "http",
			expectedHost:     "localhost:8334",
			expectedEndpoint: "http://localhost:8334",
		},
		{
			name:             "endpoint with https scheme",
			envEndpoint:      "https://custom-s3.domain.com",
			envUseSSL:        "true",
			envEnabled:       "true",
			expectedNil:      false,
			expectedScheme:   "https",
			expectedHost:     "custom-s3.domain.com",
			expectedEndpoint: "https://custom-s3.domain.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("S3_ENDPOINT", tt.envEndpoint)
			os.Setenv("S3_USE_SSL", tt.envUseSSL)
			os.Setenv("S3_ENABLED", tt.envEnabled)
			defer func() {
				os.Unsetenv("S3_ENDPOINT")
				os.Unsetenv("S3_USE_SSL")
				os.Unsetenv("S3_ENABLED")
			}()

			cfg := storage.GetS3Config()

			if tt.expectedNil {
				if cfg.Endpoint != nil {
					t.Fatalf("expected Endpoint to be nil, got %v", cfg.Endpoint)
				}
				return
			}

			if cfg.Endpoint == nil {
				t.Fatalf("expected Endpoint to be non-nil")
			}
			if cfg.Endpoint.Scheme != tt.expectedScheme {
				t.Errorf("expected Scheme %q, got %q", tt.expectedScheme, cfg.Endpoint.Scheme)
			}
			if cfg.Endpoint.Host != tt.expectedHost {
				t.Errorf("expected Host %q, got %q", tt.expectedHost, cfg.Endpoint.Host)
			}
			if cfg.Endpoint.String() != tt.expectedEndpoint {
				t.Errorf("expected Endpoint string %q, got %q", tt.expectedEndpoint, cfg.Endpoint.String())
			}
		})
	}
}

func TestGetS3Config_BucketSanitizationAndDeduplication(t *testing.T) {
	tests := []struct {
		name            string
		s3Bucket        string
		litestreamBucket string
		expectedBucket  string
		expectedBuckets []string
	}{
		{
			name:             "normal distinct buckets",
			s3Bucket:         "app-storage",
			litestreamBucket: "db-backups",
			expectedBucket:   "app-storage",
			expectedBuckets: []string{"app-storage", "db-backups"},
		},
		{
			name:             "buckets with s3 prefix and trailing slashes",
			s3Bucket:         "s3://app-storage/",
			litestreamBucket: "s3://db-backups/",
			expectedBucket:   "app-storage",
			expectedBuckets: []string{"app-storage", "db-backups"},
		},
		{
			name:             "identical buckets should be deduplicated",
			s3Bucket:         "shared-bucket",
			litestreamBucket: "shared-bucket",
			expectedBucket:   "shared-bucket",
			expectedBuckets: []string{"shared-bucket"},
		},
		{
			name:             "empty litestream bucket should not produce empty entry",
			s3Bucket:         "app-storage",
			litestreamBucket: "",
			expectedBucket:   "app-storage",
			expectedBuckets: []string{"app-storage"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("S3_BUCKET", tt.s3Bucket)
			os.Setenv("LITESTREAM_BUCKET", tt.litestreamBucket)
			defer func() {
				os.Unsetenv("S3_BUCKET")
				os.Unsetenv("LITESTREAM_BUCKET")
			}()

			cfg := storage.GetS3Config()

			if cfg.Bucket != tt.expectedBucket {
				t.Errorf("expected Bucket %q, got %q", tt.expectedBucket, cfg.Bucket)
			}

			if len(cfg.Buckets) != len(tt.expectedBuckets) {
				t.Fatalf("expected %d buckets, got %d (%v)", len(tt.expectedBuckets), len(cfg.Buckets), cfg.Buckets)
			}

			for i, b := range cfg.Buckets {
				if b != tt.expectedBuckets[i] {
					t.Errorf("bucket[%d]: expected %q, got %q", i, tt.expectedBuckets[i], b)
				}
			}
		})
	}
}

func TestMakeBucket_DisabledReturnsNil(t *testing.T) {
	os.Setenv("S3_ENABLED", "false")
	defer os.Unsetenv("S3_ENABLED")

	err := storage.MakeBucket()
	if err != nil {
		t.Fatalf("expected nil error when S3 is disabled, got %v", err)
	}
}
