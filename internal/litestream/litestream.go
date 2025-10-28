package litestream

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ReplicateDatabase() error {
	dbFile := os.Getenv("LITESTREAM_DB_PATH")
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		if err := restoreDatabase(dbFile); err != nil {
			return fmt.Errorf("failed to restore database: %w", err)
		}
	}

	if err := checkDatabaseIntegrity(dbFile); err != nil {
		log.Printf("[WARN] Database integrity check failed: %v", err)
	}

	replicaType := os.Getenv("LITESTREAM_REPLICA_TYPE")
	if replicaType == "" {
		replicaType = "file" // default to file
	}
	config, err := CreateLitestreamConfig(replicaType)
	if err != nil {
		return err
	}
	if configPath, err := createLitestreamConfig(config); err != nil {
		return err
	} else {
		cmd := exec.Command("/litestream", "replicate", "-config", configPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start replication: %w", err)
		}
	}

	return nil
}

func createLitestreamConfig(configData []byte) (string, error) {
	if err := os.WriteFile("/tmp/litestream.yml", configData, 0644); err != nil {
		return "", fmt.Errorf("failed to create litestream config: %w", err)
	}
	return "/tmp/litestream.yml", nil
}

func restoreDatabase(dbPath string) error {
	replicaType := os.Getenv("LITESTREAM_REPLICA_TYPE")
	if replicaType == "" {
		replicaType = "file"
	}
	config, err := CreateLitestreamConfig(replicaType)
	if err != nil {
		return err
	}
	if configPath, err := createLitestreamConfig(config); err != nil {
		return err
	} else {
		cmd := exec.Command("/litestream", "restore", "-if-replica-exists", "-config", configPath, dbPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}
	}

	return nil
}

func checkDatabaseIntegrity(dbPath string) error {
	cmd := exec.Command("sqlite3", dbPath, "PRAGMA integrity_check;")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("integrity check command failed: %w", err)
	}

	if !strings.Contains(string(output), "ok") {
		return fmt.Errorf("database integrity check failed: %s", string(output))
	}

	return nil
}

func CheckReplicationHealth() error {
	replicaType := os.Getenv("LITESTREAM_REPLICA_TYPE")
	if replicaType == "" {
		replicaType = "file" // default to file
	}
	configPath := getLitestreamConfigPath(replicaType)
	if _, err := os.Stat(configPath); err == nil {
		cmd := exec.Command("/litestream", "db", "list", "-config", configPath)
		output, err := cmd.Output()
		if err == nil {
			outputStr := string(output)
			if !strings.Contains(outputStr, "error") && strings.Contains(outputStr, "ok") {
				if replicaType == "s3" {
					log.Println("[OK] S3 replication health check passed")
				} else {
					log.Println("[OK] File replication health check passed")
				}
				return nil
			}
		}
	}
	return fmt.Errorf("replication health check failed")
}
