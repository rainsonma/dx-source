package helpers

import (
	"os"

	"github.com/goravel/framework/facades"
)

// StoragePath returns the configured storage root.
// Prefers the STORAGE_PATH env var so tests can t.Setenv without booting Goravel;
// falls back to the framework config. Production and test resolve to the same
// value because Goravel seeds config from env at boot.
func StoragePath() string {
	if v := os.Getenv("STORAGE_PATH"); v != "" {
		return v
	}
	return facades.Config().Env("STORAGE_PATH", "storage/app").(string)
}
