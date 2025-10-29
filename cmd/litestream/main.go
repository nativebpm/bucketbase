package main

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"
)

const configPath = "/tmp/litestream.yml"

type LitestreamYml struct {
	AccessKeyID     string           `yaml:"access-key-id,omitempty"`
	SecretAccessKey string           `yaml:"secret-access-key,omitempty"`
	Addr            string           `yaml:"addr,omitempty"`
	MCPAddr         string           `yaml:"mcp-addr,omitempty"`
	Logging         LoggingConfig    `yaml:"logging,omitempty"`
	Levels          []LevelConfig    `yaml:"levels,omitempty"`
	Snapshot        SnapshotConfig   `yaml:"snapshot,omitempty"`
	Dbs             []DatabaseConfig `yaml:"dbs"`
}

type LevelConfig struct {
	Interval string `yaml:"interval"`
}

type SnapshotConfig struct {
	Interval  string `yaml:"interval,omitempty"`
	Retention string `yaml:"retention,omitempty"`
}

type DatabaseConfig struct {
	Path                   string        `yaml:"path"`
	MetaPath               string        `yaml:"meta-path,omitempty"`
	MonitorInterval        string        `yaml:"monitor-interval,omitempty"`
	CheckpointInterval     string        `yaml:"checkpoint-interval,omitempty"`
	BusyTimeout            string        `yaml:"busy-timeout,omitempty"`
	MinCheckpointPageCount int           `yaml:"min-checkpoint-page-count,omitempty"`
	MaxCheckpointPageCount int           `yaml:"max-checkpoint-page-count,omitempty"`
	Replica                ReplicaConfig `yaml:"replica"`
}

type ReplicaConfig struct {
	// Common fields
	Name                   string    `yaml:"name,omitempty"`
	Type                   string    `yaml:"type,omitempty"`
	URL                    string    `yaml:"url,omitempty"`
	Retention              string    `yaml:"retention,omitempty"`
	RetentionCheckInterval string    `yaml:"retention-check-interval,omitempty"`
	SnapshotInterval       string    `yaml:"snapshot-interval,omitempty"`
	ValidationInterval     string    `yaml:"validation-interval,omitempty"`
	SyncInterval           string    `yaml:"sync-interval,omitempty"`
	Generation             string    `yaml:"generation,omitempty"`
	GenerationInterval     string    `yaml:"generation-interval,omitempty"`
	ValidationOnRestore    bool      `yaml:"validation-on-restore,omitempty"`
	MaxWALBytes            string    `yaml:"max-wal-bytes,omitempty"`
	Compress               string    `yaml:"compress,omitempty"`
	Age                    AgeConfig `yaml:"age,omitempty"`

	// S3 specific
	Bucket          string `yaml:"bucket,omitempty"`
	Path            string `yaml:"path,omitempty"`
	Region          string `yaml:"region,omitempty"`
	Endpoint        string `yaml:"endpoint,omitempty"`
	AccessKeyID     string `yaml:"access-key-id,omitempty"`
	SecretAccessKey string `yaml:"secret-access-key,omitempty"`
	SkipVerify      bool   `yaml:"skip-verify,omitempty"`
	ForcePathStyle  bool   `yaml:"force-path-style,omitempty"`
	SSE             string `yaml:"sse,omitempty"`
	SSEKMSKeyID     string `yaml:"sse-kms-key-id,omitempty"`

	// File specific (path already above)

	// GS specific (bucket, path already above)

	// ABS specific
	AccountName string `yaml:"account-name,omitempty"`
	AccountKey  string `yaml:"account-key,omitempty"`
	// bucket, path, endpoint already above

	// SFTP specific
	Host    string `yaml:"host,omitempty"`
	User    string `yaml:"user,omitempty"`
	KeyPath string `yaml:"key-path,omitempty"`
	// path already above

	// NATS specific
	// bucket already above
	Username      string   `yaml:"username,omitempty"`
	Password      string   `yaml:"password,omitempty"`
	JWT           string   `yaml:"jwt,omitempty"`
	Seed          string   `yaml:"seed,omitempty"`
	Creds         string   `yaml:"creds,omitempty"`
	NKey          string   `yaml:"nkey,omitempty"`
	Token         string   `yaml:"token,omitempty"`
	TLS           bool     `yaml:"tls,omitempty"`
	RootCAs       []string `yaml:"root-cas,omitempty"`
	ClientCert    string   `yaml:"client-cert,omitempty"`
	ClientKey     string   `yaml:"client-key,omitempty"`
	MaxReconnects int      `yaml:"max-reconnects,omitempty"`
	ReconnectWait string   `yaml:"reconnect-wait,omitempty"`
	Timeout       string   `yaml:"timeout,omitempty"`
}

type AgeConfig struct {
	Identities []string `yaml:"identities,omitempty"`
	Recipients []string `yaml:"recipients,omitempty"`
}

type LoggingConfig struct {
	Level  string `yaml:"level,omitempty"`
	Type   string `yaml:"type,omitempty"`
	Stderr bool   `yaml:"stderr,omitempty"`
}

func CreateLitestreamConfig(replicaType string) ([]byte, error) {
	var replica ReplicaConfig
	switch replicaType {
	case "s3":
		replica = ReplicaConfig{
			Type:             "s3",
			Bucket:           os.Getenv("LITESTREAM_BUCKET"),
			AccessKeyID:      os.Getenv("LITESTREAM_ACCESS_KEY_ID"),
			SecretAccessKey:  os.Getenv("LITESTREAM_SECRET_ACCESS_KEY"),
			Region:           os.Getenv("LITESTREAM_REGION"),
			Endpoint:         os.Getenv("LITESTREAM_ENDPOINT"),
			SkipVerify:       os.Getenv("LITESTREAM_SKIP_VERIFY") == "true",
			SyncInterval:     getEnvOrDefault("LITESTREAM_SYNC_INTERVAL", "5m"),
			SnapshotInterval: getEnvOrDefault("LITESTREAM_SNAPSHOT_INTERVAL", "6h"),
			Retention:        getEnvOrDefault("LITESTREAM_RETENTION", "168h"),
			MaxWALBytes:      getEnvOrDefault("LITESTREAM_MAX_WAL_BYTES", "512MB"),
			Compress:         getEnvOrDefault("LITESTREAM_COMPRESS", "gzip"),
		}
	case "file":
		replica = ReplicaConfig{
			Type:             "file",
			Path:             os.Getenv("LITESTREAM_BACKUP_PATH"),
			SyncInterval:     getEnvOrDefault("LITESTREAM_SYNC_INTERVAL", "5m"),
			SnapshotInterval: getEnvOrDefault("LITESTREAM_SNAPSHOT_INTERVAL", "6h"),
			Retention:        getEnvOrDefault("LITESTREAM_RETENTION", "168h"),
			MaxWALBytes:      getEnvOrDefault("LITESTREAM_MAX_WAL_BYTES", "512MB"),
			Compress:         getEnvOrDefault("LITESTREAM_COMPRESS", "gzip"),
		}
	default:
		return nil, fmt.Errorf("unsupported replica type: %s", replicaType)
	}

	config := LitestreamYml{
		Levels: []LevelConfig{
			{Interval: "5m"},
			{Interval: "1h"},
			{Interval: "24h"},
		},
		Snapshot: SnapshotConfig{
			Interval:  getEnvOrDefault("LITESTREAM_SNAPSHOT_INTERVAL", "6h"),
			Retention: getEnvOrDefault("LITESTREAM_RETENTION", "168h"),
		},
		Dbs: []DatabaseConfig{
			{
				Path:     os.Getenv("LITESTREAM_DB_PATH"),
				MetaPath: os.Getenv("LITESTREAM_BACKUP_PATH"),
				Replica:  replica,
			},
		},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal litestream config: %w", err)
	}

	return data, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func CreateLitestreamConfigFile(configData []byte) error {
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to create litestream config: %w", err)
	}
	return nil
}

func init() {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		replicaType := os.Getenv("LITESTREAM_REPLICA_TYPE")
		if replicaType == "" {
			replicaType = "file"
		}
		configData, err := CreateLitestreamConfig(replicaType)
		if err != nil {
			slog.Error("Failed to create litestream config", "error", err)
			os.Exit(1)
		}
		if err := CreateLitestreamConfigFile(configData); err != nil {
			slog.Error("Failed to write litestream config", "error", err)
			os.Exit(1)
		}
	}
}

func main() {
	err := syscall.Exec("/litestream", []string{"/litestream", "replicate", "-config", configPath, "-exec", "/pocketbase serve --http :8090"}, os.Environ())
	if err != nil {
		slog.Error("Failed to exec litestream", "error", err)
		os.Exit(1)
	}
}
