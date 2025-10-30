package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func databaseRestore() error {
	cfg, err := litestream.Config()
	if err != nil {
		return fmt.Errorf("failed to generate litestream config: %w", err)
	}

	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		slog.Info("Database file not found, attempting restore", "path", cfg.DBPath)
		cmd := exec.Command("/litestream",
			"restore", "-config", cfg.ConfigPath, "-o", cfg.DBPath, "-if-replica-exists")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if restoreErr := cmd.Run(); restoreErr != nil {
			return fmt.Errorf("failed to restore database: %w", restoreErr)
		}
		slog.Info("Database restored successfully", "path", cfg.DBPath)
	}

	return nil
}

func main() {
	if err := databaseRestore(); err != nil {
		slog.Error("Database restore failed", "error", err)
	}

	err := syscall.Exec("/pocketbase",
		[]string{"pocketbase", "serve", "--http", ":8090"}, os.Environ())
	if err != nil {
		slog.Error("Failed to exec pocketbase", "error", err)
		os.Exit(1)
	}
}
