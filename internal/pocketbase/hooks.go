package pocketbase

import (
	"log/slog"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func SetupHooks(app *pocketbase.PocketBase) {
	config := GetConfig()

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Litestream optimizations - apply in all environments
		if _, err := app.DB().NewQuery("PRAGMA busy_timeout = 5000").Execute(); err != nil {
			slog.Warn("Failed to set busy_timeout", "error", err)
		}
		if _, err := app.DB().NewQuery("PRAGMA synchronous = NORMAL").Execute(); err != nil {
			slog.Warn("Failed to set synchronous", "error", err)
		}

		// Native Superuser Upsert (only if credentials are provided)
		if config.PocketbaseAdminEmail != "" && config.PocketbaseAdminPassword != "" {
			admin, err := app.FindAuthRecordByEmail("_superusers", config.PocketbaseAdminEmail)
			if err != nil {
				// Admin not found, create a new one
				superusers, err := app.FindCollectionByNameOrId("_superusers")
				if err != nil {
					slog.Error("Failed to find _superusers collection", "error", err)
				} else {
					admin = core.NewRecord(superusers)
					admin.SetEmail(config.PocketbaseAdminEmail)
					admin.SetPassword(config.PocketbaseAdminPassword)
					if err := app.Save(admin); err != nil {
						slog.Error("Failed to create superuser", "error", err)
					} else {
						slog.Info("Superuser created successfully", "email", config.PocketbaseAdminEmail)
					}
				}
			} else {
				// Admin exists, update password if changed
				if !admin.ValidatePassword(config.PocketbaseAdminPassword) {
					admin.SetPassword(config.PocketbaseAdminPassword)
					if err := app.Save(admin); err != nil {
						slog.Error("Failed to update superuser password", "error", err)
					} else {
						slog.Info("Superuser password updated successfully", "email", config.PocketbaseAdminEmail)
					}
				}
			}
		}

		return e.Next()
	})
}
