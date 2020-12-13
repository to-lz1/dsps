package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/saiya/dsps/server/config"
	. "github.com/saiya/dsps/server/testing"
)

func TestChannelWebhookDefaultConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	webhooks:
		-
			url: 'http://localhost:3001/you-got-message/room/{{.regex.id}}'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	webhook := cfg.Webhooks[0]
	assert.Equal(t, "http://localhost:3001/you-got-message/room/{{.regex.id}}", webhook.URL.String())
	assert.Equal(t, MakeDurationPtr("30s"), webhook.Timeout)
	assert.Equal(t, MakeIntPtr(3), webhook.Retry.Count)
	assert.Equal(t, MakeDurationPtr("3s"), webhook.Retry.Interval)
	assert.Equal(t, 1.5, *webhook.Retry.IntervalMultiplier)
	assert.Equal(t, MakeDurationPtr("1.5s"), webhook.Retry.IntervalJitter)
	assert.Equal(t, 0, len(webhook.Headers))
}

func TestWebhookFullConfig(t *testing.T) {
	configYaml := strings.ReplaceAll(`
channels:
-
	regex: 'chat-room-(?P<id>\d+)'
	webhooks:
		-
			url: 'http://localhost:3001/you-got-message/room/{{.regex.id}}'
			timeout: 61s
			retry:
				count: 4
				interval: 3.5s
				intervalMultiplier: 3.1
				intervalJitter: 2s500ms
			headers:
				User-Agent: my DSPS server
				X-Chat-Room-ID: '{{.regex.id}}'
`, "\t", "  ")
	config, err := ParseConfig(Overrides{}, configYaml)
	if err != nil {
		t.Error(err)
		return
	}

	cfg := config.Channels[0]
	assert.Equal(t, "chat-room-(?P<id>\\d+)", cfg.Regex.String())

	webhook := cfg.Webhooks[0]
	assert.Equal(t, "http://localhost:3001/you-got-message/room/{{.regex.id}}", webhook.URL.String())
	assert.Equal(t, MakeDurationPtr("61s"), webhook.Timeout)
	assert.Equal(t, MakeIntPtr(4), webhook.Retry.Count)
	assert.Equal(t, MakeDurationPtr("3.5s"), webhook.Retry.Interval)
	assert.Equal(t, 3.1, *webhook.Retry.IntervalMultiplier)
	assert.Equal(t, MakeDurationPtr("2.5s"), webhook.Retry.IntervalJitter)
	assert.Equal(t, "my DSPS server", webhook.Headers["User-Agent"].String())
	assert.Equal(t, "{{.regex.id}}", webhook.Headers["X-Chat-Room-ID"].String())
}

func TestInvalidWebhookConfig(t *testing.T) {
	_, err := ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", retry: { count: 0 } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: retry.count must not be negative nor zero`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", retry: { interval: "0s" } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: retry.interval must not be negative nor zero`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", retry: { intervalMultiplier: 0.9 } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: retry.intervalMultipler must be equal to or larger than 1.0`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", retry: { intervalJitter: 0 } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: retry.intervalJitter must not be negative nor zero`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", connection: { max: 0 } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: connection.max must not be negative nor zero`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", connection: { maxIdleTime: "0s" } ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: connection.maxIdleTime must not be negative nor zero`, err.Error())

	_, err = ParseConfig(Overrides{}, `channels: [ { regex: '.+', webhooks: [ url: "http://localhost:3000", timeout: "0s" ] } ]`)
	assert.Regexp(t, `error on webhooks\[1\]: timeout must not be negative nor zero`, err.Error())
}
