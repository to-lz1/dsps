package testing

import (
	"testing"

	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/storage/deps"
	"github.com/m3dev/dsps/server/telemetry"
)

// EmptyDeps fills stub objects
func EmptyDeps(t *testing.T) deps.StorageDeps {
	return deps.StorageDeps{
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}
}
