package deps

import (
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/telemetry"
)

// StorageDeps contains objects required by storage implementations.
type StorageDeps struct {
	Telemetry *telemetry.Telemetry
	Sentry    sentry.Sentry
}
