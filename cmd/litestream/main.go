package main

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/nativebpm/pocketstream/internal/litestream"
)

func findLitestreamBin() string {
	candidates := []string{
		"/usr/local/bin/litestream",
		"/litestream",
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
	return "/usr/local/bin/litestream"
}

func main() {
	cfg, err := litestream.Config()
	if err != nil {
		slog.Error("Failed to generate litestream config", "error", err)
		os.Exit(1)
	}

	bin := findLitestreamBin()
	err = syscall.Exec(bin,
		[]string{"litestream", "replicate", "-config", cfg.ConfigPath}, os.Environ())
	if err != nil {
		slog.Error("Failed to exec litestream replicate", "error", err, "path", bin)
		os.Exit(1)
	}
}
