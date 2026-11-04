package cacheengine

import (
	"hash/fnv"
	"time"
)

// ShardedCache partitions keys across multiple Cache shards to reduce lock contention.
type ShardedCache struct {
	shards    []*Cache
	shardMask uint32
}

// NewShardedCache creates a cache with 2^n shards.
func NewShardedCache(shardCount int) *ShardedCache {
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		shardCount = 16
	}
	shards := make([]*Cache, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = New()
	}
	return &ShardedCache{
		shards:    shards,
		shardMask: uint32(shardCount - 1),
	}
}

func (sc *ShardedCache) getShard(key string) *Cache {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := h.Sum32() & sc.shardMask
	return sc.shards[idx]
}

// Set stores a key in the appropriate shard.
func (sc *ShardedCache) Set(key string, val interface{}, ttl time.Duration) {
	sc.getShard(key).Set(key, val, ttl)
}

// Get retrieves a key from the appropriate shard.
func (sc *ShardedCache) Get(key string) (interface{}, bool) {
	return sc.getShard(key).Get(key)
}

// Delete removes a key from the appropriate shard.
func (sc *ShardedCache) Delete(key string) {
	sc.getShard(key).Delete(key)
}

