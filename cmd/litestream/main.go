package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func main() {
	if cfg, err := litestream.Config(); err != nil {
		slog.Error("Failed to load litestream config", "error", err)
		os.Exit(1)
	} else {
		cmd := exec.Command("/litestream", "restore", "-if-replica-exists", "-config", cfg.ConfigPath, cfg.DBPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			slog.Warn("Failed to restore database, starting fresh", "error", err)
		}
		err := syscall.Exec("/litestream", []string{"/litestream", "replicate", "-config", cfg.ConfigPath, "-exec", "/pocketbase serve --http :8090"}, os.Environ())
		if err != nil {
			slog.Error("Failed to exec litestream", "error", err)
			os.Exit(1)
		}
	}
}
