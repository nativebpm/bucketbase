package litestream_test

import (
	"os"
	"strings"
	"testing"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func TestConfig_DefaultDBPath(t *testing.T) {
	os.Unsetenv("LITESTREAM_DB_PATH")
	os.Unsetenv("LITESTREAM_META_PATH")

	cfg, err := litestream.Config()
	if err != nil {
		t.Fatalf("Config() failed: %v", err)
	}

	expectedDBPath := "/pb_data/data.db"
	expectedMetaPath := "/pb_data/data.db-litestream"

	if len(cfg.Dbs) == 0 {
		t.Fatalf("expected at least 1 database config, got 0")
	}

	if cfg.Dbs[0].Path != expectedDBPath {
		t.Errorf("expected Dbs[0].Path %q, got %q", expectedDBPath, cfg.Dbs[0].Path)
	}

	if cfg.Dbs[0].MetaPath != expectedMetaPath {
		t.Errorf("expected Dbs[0].MetaPath %q, got %q", expectedMetaPath, cfg.Dbs[0].MetaPath)
	}

	if cfg.DBPath != expectedDBPath {
		t.Errorf("expected DBPath %q, got %q", expectedDBPath, cfg.DBPath)
	}

	// Verify the written YAML file content
	content, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		t.Fatalf("failed to read generated config file %s: %v", cfg.ConfigPath, err)
	}

	if !strings.Contains(string(content), "path: /pb_data/data.db") {
		t.Errorf("generated config file missing 'path: /pb_data/data.db', content:\n%s", string(content))
	}
}

func TestConfig_CustomDBPath(t *testing.T) {
	customPath := "/custom/database/mydb.sqlite"
	os.Setenv("LITESTREAM_DB_PATH", customPath)
	defer os.Unsetenv("LITESTREAM_DB_PATH")

	cfg, err := litestream.Config()
	if err != nil {
		t.Fatalf("Config() failed: %v", err)
	}

	if cfg.Dbs[0].Path != customPath {
		t.Errorf("expected Dbs[0].Path %q, got %q", customPath, cfg.Dbs[0].Path)
	}

	expectedMetaPath := customPath + "-litestream"
	if cfg.Dbs[0].MetaPath != expectedMetaPath {
		t.Errorf("expected Dbs[0].MetaPath %q, got %q", expectedMetaPath, cfg.Dbs[0].MetaPath)
	}
}

func TestConfig_FileReplicaDefaultBackupPath(t *testing.T) {
	os.Setenv("LITESTREAM_REPLICA_TYPE", "file")
	os.Unsetenv("LITESTREAM_BACKUP_PATH")
	defer os.Unsetenv("LITESTREAM_REPLICA_TYPE")

	cfg, err := litestream.Config()
	if err != nil {
		t.Fatalf("Config() failed: %v", err)
	}

	expectedBackup := "/pb_backup"
	if cfg.Dbs[0].Replica.Path != expectedBackup {
		t.Errorf("expected file replica Path %q, got %q", expectedBackup, cfg.Dbs[0].Replica.Path)
	}
}

func TestConfig_S3Replica(t *testing.T) {
	os.Setenv("LITESTREAM_REPLICA_TYPE", "s3")
	os.Setenv("LITESTREAM_BUCKET", "my-test-backup")
	defer func() {
		os.Unsetenv("LITESTREAM_REPLICA_TYPE")
		os.Unsetenv("LITESTREAM_BUCKET")
	}()

	cfg, err := litestream.Config()
	if err != nil {
		t.Fatalf("Config() failed: %v", err)
	}

	expectedURL := "s3://my-test-backup/"
	if cfg.Dbs[0].Replica.URL != expectedURL {
		t.Errorf("expected S3 replica URL %q, got %q", expectedURL, cfg.Dbs[0].Replica.URL)
	}
}

func TestConfig_UnsupportedReplicaType(t *testing.T) {
	os.Setenv("LITESTREAM_REPLICA_TYPE", "invalid_type_xyz")
	defer os.Unsetenv("LITESTREAM_REPLICA_TYPE")

	_, err := litestream.Config()
	if err == nil {
		t.Fatalf("expected error for unsupported replica type, got nil")
	}
}
