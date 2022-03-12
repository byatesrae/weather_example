package memorycache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheGet(t *testing.T) {
	t.Parallel()

	t.Run("no_value_for_key", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		m := New()
		val, _, err := m.Get(ctx, "Test123")
		assert.Nil(t, val)
		assert.Nil(t, err)
	})

	t.Run("value_retrieved", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		expectedExpiry := time.Now()

		m := New()
		err := m.Set(ctx, "Test123", 456, expectedExpiry)
		assert.NoError(t, err)

		actualValue, actualExpiry, err := m.Get(ctx, "Test123")
		assert.Equal(t, 456, actualValue)
		assert.Equal(t, expectedExpiry, actualExpiry)
		assert.Nil(t, err)
	})
}
