package main

import (
	"log/slog"
	"os"

	"github.com/nativebpm/pocketstream/internal/pocketbase"
)

func main() {
	app := pocketbase.New()

	pocketbase.SetupHooks(app)

	if err := app.Start(); err != nil {
		slog.Error("Failed to start application", "error", err)
		os.Exit(1)
	}
}
