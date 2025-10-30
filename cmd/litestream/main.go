package main

import (
	"log/slog"
	"os"
	"syscall"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func main() {
	cfg, err := litestream.Config()
	if err != nil {
		slog.Error("Failed to generate litestream config", "error", err)
		os.Exit(1)
	}

	err = syscall.Exec("/usr/local/bin/litestream",
		[]string{"litestream", "replicate", "-config", cfg.ConfigPath}, os.Environ())
	if err != nil {
		slog.Error("Failed to exec litestream replicate", "error", err)
		os.Exit(1)
	}
}
