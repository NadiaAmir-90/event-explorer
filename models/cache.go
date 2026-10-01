package models

import "sync"

// Cache is a thread-safe in-memory cache.
type Cache struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

// NewCache creates a new empty cache.
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]interface{}),
	}
}

// GetOrFetch returns a cached value if it exists.
// On cache hit:
//   - returns the cached value
//   - keeps the value in the cache
// On cache miss:
//   - calls fetch()
//   - stores the returned value
//   - returns the value
func (c *Cache) GetOrFetch(
	key string,
	fetch func() interface{},
) interface{} {

	// Check whether the value is already cached.
	c.mu.RLock()
	value, exists := c.items[key]
	c.mu.RUnlock()

	if exists {
		return value
	}

	// Cache miss.
	value = fetch()

	// Store the fetched value.
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()

	return value
}

// Clear removes all values from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]interface{})
}
