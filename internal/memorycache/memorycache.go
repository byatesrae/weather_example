// Package memorycache contains a very simple in-memory cache.
package memorycache

import (
	"context"
	"sync"
	"time"
)

// cacheEntry represents a value cached.
type cacheEntry struct {
	val    interface{}
	expiry time.Time
}

// MemoryCache is a very simple in-memory cache that is safe for concurrent access.
type MemoryCache struct {
	values map[interface{}]*cacheEntry

	m sync.Mutex
}

// New creates a new MemoryCache.
func New() *MemoryCache {
	return &MemoryCache{
		values: make(map[interface{}]*cacheEntry),
	}
}

// Get retrieves a value from cache as well as the time it expires.
func (m *MemoryCache) Get(ctx context.Context, key interface{}) (interface{}, time.Time, error) {
	m.m.Lock()
	defer m.m.Unlock()

	entry, ok := m.values[key]
	if !ok {
		return nil, time.Time{}, nil
	}

	return entry.val, entry.expiry, nil
}

// Set will set a value to be cached as well as an expiry.
// Cache entries are not automatically evicted based on their expiry.
func (m *MemoryCache) Set(ctx context.Context, key, val interface{}, expiry time.Time) error {
	m.m.Lock()
	defer m.m.Unlock()

	entry := cacheEntry{val: val, expiry: expiry}
	m.values[key] = &entry

	return nil
}
