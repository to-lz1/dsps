package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChannelAtomStringer(t *testing.T) {
	assert.Equal(t, `chat-room-(?P<id>\d+)`, newChannelAtomByYaml(t, `{ regex: 'chat-room-(?P<id>\d+)' }`, true).String())
}

func TestChannelAtomMatching(t *testing.T) {
	assert.True(t, newChannelAtomByYaml(t, `{ regex: 'chat-room-(?P<id>\d+)' }`, true).IsMatch("chat-room-123"))
	assert.False(t, newChannelAtomByYaml(t, `{ regex: 'chat-room-(?P<id>\d+)' }`, true).IsMatch("Xchat-room-123"))
}

func TestChannelAtomTemplateValidation(t *testing.T) {
	var testdata = []struct {
		errorRegex string
		yaml       string
	}{
		{
			"",
			`
regex: 'chat-room-(?P<id>\d+)'
webhooks:
	-
		url: 'http://localhost:3001/you-got-message/room/{{.regex.id}}'
		headers:
			User-Agent: "{{.regex.id}}"
jwt:
	iss: [ "http://example.com" ]
	keys: RS256: [ "../jwt/testdata/RS256-2048bit-public.pem" ]
	claims:
		chatroom: '{{.regex.id}}'`,
		},
		{
			`invalid template found on webhooks\[0\].url:.*map has no entry for key`,
			`
regex: 'chat-room-(?P<id>\d+)'
webhooks:
	# Invalid template key "idX"
	- url: 'http://localhost:3001/you-got-message/room/{{.regex.idX}}'`,
		},
		{
			`invalid template found on webhooks\[0\].headers.User-Agent:.*map has no entry for key`,
			`
regex: 'chat-room-(?P<id>\d+)'
webhooks:
	-
		url: 'http://localhost:3001/you-got-message/room/{{.regex.id}}'
		headers:
			User-Agent: "{{.regex.idX}}"`,
		},
		{
			`invalid template found on jwt.claims.chatroom:.*map has no entry for key`,
			`
regex: 'chat-room-(?P<id>\d+)'
jwt:
	iss: [ "http://example.com" ]
	keys: RS256: [ "../jwt/testdata/RS256-2048bit-public.pem" ]
	claims:
		chatroom: '{{.regex.idX}}'`,
		},
	}
	for _, tt := range testdata {
		err := newChannelAtomByYaml(t, tt.yaml, false).validate()
		if tt.errorRegex == "" {
			assert.NoError(t, err)
		} else {
			assert.Regexp(t, tt.errorRegex, err)
		}
	}
}
