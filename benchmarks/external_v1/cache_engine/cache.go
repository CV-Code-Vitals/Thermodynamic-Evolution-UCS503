package cacheengine

import (
	"sync"
	"time"
)

// Item represents a cached value with optional expiration.
type Item struct {
	Value      interface{}
	Expiration int64
}

// IsExpired returns true if the item has expired.
func (i *Item) IsExpired() bool {
	if i.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.Expiration
}

// Cache is a thread-safe in-memory key-value store.
type Cache struct {
	items map[string]*Item
	mu    sync.RWMutex
	stats CacheStats
}

// New creates an initialized Cache.
func New() *Cache {
	return &Cache{
		items: make(map[string]*Item),
	}
}

// Set stores a key-value pair with a TTL duration.
func (c *Cache) Set(key string, val interface{}, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value:      val,
		Expiration: exp,
	}
	c.stats.RecordSet()
}

// Get retrieves a key from the cache.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		c.stats.RecordMiss()
		return nil, false
	}

	if item.IsExpired() {
		c.Delete(key)
		c.stats.RecordMiss()
		return nil, false
	}

	c.stats.RecordHit()
	return item.Value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.items[key]; found {
		delete(c.items, key)
		c.stats.RecordDel()
	}
}

// UnsafeKeys returns all keys currently present in the cache.
// DEFECT: Concurrency data race - reads from shared items map without acquiring RLock.
func (c *Cache) UnsafeKeys() []string {
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// SafeKeys returns all keys with proper locking (clean implementation).
func (c *Cache) SafeKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

