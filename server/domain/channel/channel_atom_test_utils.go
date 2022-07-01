package channel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/telemetry"
)

func newChannelAtomByYaml(t *testing.T, yaml string, validate bool) *channelAtom {
	yaml = fmt.Sprintf("channels:\n  - %s", strings.ReplaceAll(strings.ReplaceAll(yaml, "\t", "  "), "\n", "\n    "))
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, yaml)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cfg.Channels))

	atom, err := newChannelAtom(context.Background(), &cfg.Channels[0], ProviderDeps{
		Clock:     domain.RealSystemClock,
		Telemetry: telemetry.NewEmptyTelemetry(t),
		Sentry:    sentry.NewEmptySentry(),
	}, validate)
	assert.NoError(t, err)
	return atom
}
