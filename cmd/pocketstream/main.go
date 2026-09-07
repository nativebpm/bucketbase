package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func findLitestreamBin() string {
	candidates := []string{
		"/litestream",
		"/usr/local/bin/litestream",
		"/usr/bin/litestream",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if path, err := exec.LookPath("litestream"); err == nil {
		return path
	}
	return "/litestream"
}

func findPocketbaseBin() string {
	candidates := []string{
		"/pocketbase",
		"./pocketbase",
		"/usr/local/bin/pocketbase",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if path, err := exec.LookPath("pocketbase"); err == nil {
		return path
	}
	return "/pocketbase"
}

func databaseRestore() error {
	cfg, err := litestream.Config()
	if err != nil {
		return fmt.Errorf("failed to generate litestream config: %w", err)
	}

	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for database: %w", err)
		}
		
		bin := findLitestreamBin()
		slog.Info("Database file not found, attempting restore", "path", cfg.DBPath, "litestream", bin)
		cmd := exec.Command(bin, "restore", "-config", cfg.ConfigPath, cfg.DBPath)
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

	bin := findPocketbaseBin()
	err := syscall.Exec(bin,
		[]string{"pocketbase", "serve", "--http", ":8090"}, os.Environ())
	if err != nil {
		slog.Error("Failed to exec pocketbase", "error", err, "path", bin)
		os.Exit(1)
	}
}
