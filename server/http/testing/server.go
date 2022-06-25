package testing

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	"github.com/m3dev/dsps/server/domain/channel"
	"github.com/m3dev/dsps/server/http"
	httplifecycle "github.com/m3dev/dsps/server/http/lifecycle"
	"github.com/m3dev/dsps/server/logger"
	"github.com/m3dev/dsps/server/sentry"
	"github.com/m3dev/dsps/server/storage"
	"github.com/m3dev/dsps/server/storage/deps"
	"github.com/m3dev/dsps/server/telemetry"
)

// WithServerDeps runs given test function with ServerDependencies
func WithServerDeps(t *testing.T, configYaml string, f func(*http.ServerDependencies)) {
	cfg, err := config.ParseConfig(context.Background(), config.Overrides{}, strings.ReplaceAll(configYaml, "\t", "  "))
	if !assert.NoError(t, err) {
		return
	}

	logFilter, err := logger.InitLogger(cfg.Logging)
	assert.NoError(t, err)

	ctx := context.Background()
	clock := domain.RealSystemClock
	sentry, err := sentry.NewSentry(cfg.Sentry)
	assert.NoError(t, err)
	telemetry, err := telemetry.InitTelemetry(cfg.Telemetry)
	assert.NoError(t, err)
	channelProvider, err := channel.NewChannelProvider(ctx, &cfg, channel.ProviderDeps{
		Clock:     clock,
		Telemetry: telemetry,
		Sentry:    sentry,
	})
	assert.NoError(t, err)
	storage, err := storage.NewStorage(ctx, &cfg.Storages, clock, channelProvider, deps.StorageDeps{
		Telemetry: telemetry,
		Sentry:    sentry,
	})
	assert.NoError(t, err)
	serverClose := httplifecycle.NewServerClose()
	defer serverClose.Close()

	f(&http.ServerDependencies{
		Config:          &cfg,
		ChannelProvider: channelProvider,
		Storage:         storage,

		LogFilter:   logFilter,
		Telemetry:   telemetry,
		Sentry:      sentry,
		ServerClose: serverClose,
	})
}

// WithServer runs given test function with HTTP server
func WithServer(t *testing.T, configYaml string, setup func(deps *http.ServerDependencies), f func(deps *http.ServerDependencies, baseURL string)) {
	WithServerDeps(t, configYaml, func(deps *http.ServerDependencies) {
		setup(deps)

		server := http.CreateServer(context.Background(), deps)
		ts := httptest.NewServer(server)
		defer ts.Close()
		defer deps.ServerClose.Close() // Close requests first

		f(deps, strings.TrimSuffix(strings.Join([]string{
			strings.TrimSuffix(ts.URL, "/"),
			strings.TrimPrefix(strings.TrimSuffix(deps.Config.HTTPServer.PathPrefix, "/"), "/"),
		}, "/"), "/"))
	})
}
