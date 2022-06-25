package multiplex

import (
	"testing"

	"github.com/m3dev/dsps/server/domain"
	. "github.com/m3dev/dsps/server/testing"
)

func TestMalformedHandle(t *testing.T) {
	_, err := decodeMultiplexAckHandle(domain.AckHandle{
		SubscriberLocator: domain.SubscriberLocator{
			ChannelID:    "ch1",
			SubscriberID: "s1",
		},
		Handle: "xxx",
	})
	IsError(t, domain.ErrMalformedAckHandle, err)
}
