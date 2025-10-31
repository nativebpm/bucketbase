package pocketbase

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupHooks(app *pocketbase.PocketBase) {
	config := GetConfig()

	if config.Profile == "docker" {
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			// Add checkpoint endpoint for schema changes
			e.Router.POST("/api/checkpoint", func(c *core.RequestEvent) error {
				if err := ForceCheckpoint(); err != nil {
					return c.JSON(500, map[string]string{"error": err.Error()})
				}
				return c.JSON(200, map[string]string{"status": "checkpoint completed"})
			})

			// Litestream optimizations
			if _, err := app.DB().NewQuery("PRAGMA busy_timeout = 5000").Execute(); err != nil {
				slog.Warn("Failed to set busy_timeout", "error", err)
			}
			if _, err := app.DB().NewQuery("PRAGMA synchronous = NORMAL").Execute(); err != nil {
				slog.Warn("Failed to set synchronous", "error", err)
			}
			if _, err := app.DB().NewQuery("PRAGMA wal_autocheckpoint = 0").Execute(); err != nil {
				slog.Warn("Failed to disable wal_autocheckpoint", "error", err)
			}

			cmd := exec.Command("/pocketbase", "superuser", "upsert", config.PocketbaseAdminEmail, config.PocketbaseAdminPassword)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				slog.Error("Superuser upsert failed", "error", err)
			}
			return e.Next()
		})

		// Add automatic checkpoint after each request that modifies data
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			// This will run after the server starts
			go func() {
				// Simple approach: checkpoint every 30 seconds
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for range ticker.C {
					if err := ForceCheckpoint(); err != nil {
						slog.Warn("Failed to periodic checkpoint", "error", err)
					} else {
						slog.Debug("Periodic checkpoint completed")
					}
				}
			}()
			return e.Next()
		})

		// Add hooks for automatic checkpoint on data changes
		app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record creation
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record create", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record create", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record update
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record update", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record update", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
			// Run checkpoint after record deletion
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after record delete", "error", err, "collection", e.Record.Collection().Name, "id", e.Record.Id)
				} else {
					slog.Debug("Checkpoint completed after record delete", "collection", e.Record.Collection().Name, "id", e.Record.Id)
				}
			}()
			return e.Next()
		})

		app.OnCollectionCreate().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection creation (schema change)
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection create", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection create", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})

		app.OnCollectionUpdate().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection update (schema change)
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection update", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection update", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})

		app.OnCollectionDelete().BindFunc(func(e *core.CollectionEvent) error {
			// Run checkpoint after collection deletion (schema change)
			go func() {
				if err := ForceCheckpoint(); err != nil {
					slog.Warn("Failed to checkpoint after collection delete", "error", err, "collection", e.Collection.Name)
				} else {
					slog.Debug("Checkpoint completed after collection delete", "collection", e.Collection.Name)
				}
			}()
			return e.Next()
		})
	}
}
