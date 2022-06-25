package onmemory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m3dev/dsps/server/config"
	"github.com/m3dev/dsps/server/domain"
	. "github.com/m3dev/dsps/server/storage/deps/testing"
	. "github.com/m3dev/dsps/server/storage/testing"
)

func makeRawStorage(t *testing.T) *onmemoryStorage {
	s, err := NewOnmemoryStorage(context.Background(), &config.OnmemoryStorageConfig{}, domain.RealSystemClock, StubChannelProvider, EmptyDeps(t))
	if !assert.NoError(t, err) {
		return nil
	}

	raw, ok := s.(*onmemoryStorage)
	if !ok {
		assert.FailNow(t, "cast failed")
	}
	return raw
}
