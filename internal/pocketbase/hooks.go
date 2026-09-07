package pocketbase

import (
	"log/slog"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupHooks(app *pocketbase.PocketBase) {
	config := GetConfig()

	if config.Profile == "docker" {
		app.OnServe().BindFunc(func(e *core.ServeEvent) error {
			// Litestream optimizations
			if _, err := app.DB().NewQuery("PRAGMA busy_timeout = 5000").Execute(); err != nil {
				slog.Warn("Failed to set busy_timeout", "error", err)
			}
			if _, err := app.DB().NewQuery("PRAGMA synchronous = NORMAL").Execute(); err != nil {
				slog.Warn("Failed to set synchronous", "error", err)
			}
			// Let Litestream manage wal_autocheckpoint natively
			// We no longer set wal_autocheckpoint = 0

			// Native Superuser Upsert
			admin, err := app.FindAuthRecordByEmail("_superusers", config.PocketbaseAdminEmail)
			if err != nil {
				// Admin not found, create a new one
				superusers, err := app.FindCollectionByNameOrId("_superusers")
				if err != nil {
					slog.Error("Failed to find _superusers collection", "error", err)
				} else {
					admin = core.NewRecord(superusers)
					admin.Set("email", config.PocketbaseAdminEmail)
					admin.SetPassword(config.PocketbaseAdminPassword)
					if err := app.Save(admin); err != nil {
						slog.Error("Failed to create superuser", "error", err)
					} else {
						slog.Info("Superuser created successfully")
					}
				}
			} else {
				// Admin exists, update password
				if !admin.ValidatePassword(config.PocketbaseAdminPassword) {
					admin.SetPassword(config.PocketbaseAdminPassword)
					if err := app.Save(admin); err != nil {
						slog.Error("Failed to update superuser password", "error", err)
					} else {
						slog.Info("Superuser password updated successfully")
					}
				}
			}

			return e.Next()
		})
	}
}
