package cacheengine

import (
	"context"
	"time"
)

// EvictExpired removes all expired items from the cache.
func (c *Cache) EvictExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	count := 0
	for k, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, k)
			c.stats.RecordDel()
			count++
		}
	}
	return count
}

// StartBackgroundEvictor starts an eviction loop.
// DEFECT: Goroutine leak - ticker loop runs indefinitely with no stop channel or context to terminate.
func (c *Cache) StartBackgroundEvictor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			c.EvictExpired()
		}
	}()
}

// StartManagedEvictor starts an eviction loop with proper lifecycle termination (clean implementation).
func (c *Cache) StartManagedEvictor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.EvictExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}

