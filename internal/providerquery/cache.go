package providerquery

import (
	"context"
	"time"

	"github.com/byatesrae/weather"
)

// resultCacheKey is used as a key to cache resultCacheEntry.
type resultCacheKey struct{}

// resultCacheEntry wraps a weather summary to be cached.
type resultCacheEntry struct {
	result    *weather.Summary
	createdAt time.Time
}

// Cache is used to store & retrieve responses.
type Cache interface {
	Get(ctx context.Context, key interface{}) (interface{}, time.Time, error)
	Set(ctx context.Context, key, val interface{}, expiry time.Time) error
}
