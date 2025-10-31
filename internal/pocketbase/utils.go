package pocketbase

import (
	"os"
	"os/exec"
	"regexp"
)

func ValidateEncryptionKey(key string) bool {
	if len(key) != 32 {
		return false
	}
	matched, err := regexp.MatchString(`^[0-9a-fA-F]{32}$`, key)
	return err == nil && matched
}

func ForceCheckpoint() error {
	// Force WAL checkpoint to ensure schema changes are written to main database file
	if _, err := os.Stat("/pb_data/data.db"); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("sqlite3", "/pb_data/data.db", "PRAGMA wal_checkpoint;")
	return cmd.Run()
}
