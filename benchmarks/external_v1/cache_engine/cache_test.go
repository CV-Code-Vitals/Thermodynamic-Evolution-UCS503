package cacheengine

import (
	"context"
	"testing"
	"time"
)

func TestCacheBasic(t *testing.T) {
	c := New()
	c.Set("foo", "bar", 1*time.Minute)

	val, found := c.Get("foo")
	if !found {
		t.Fatalf("expected key 'foo' to be found")
	}
	if val.(string) != "bar" {
		t.Fatalf("expected value 'bar', got %v", val)
	}

	c.Delete("foo")
	_, found = c.Get("foo")
	if found {
		t.Fatalf("expected key 'foo' to be deleted")
	}
}

func TestCacheExpiration(t *testing.T) {
	c := New()
	c.Set("quick", "vanish", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	_, found := c.Get("quick")
	if found {
		t.Fatalf("expected expired key 'quick' to be gone")
	}
}

func TestShardedCache(t *testing.T) {
	sc := NewShardedCache(8)
	sc.Set("key1", "val1", 0)
	sc.Set("key2", "val2", 0)

	v1, ok := sc.Get("key1")
	if !ok || v1 != "val1" {
		t.Fatalf("sharded get key1 failed: %v", v1)
	}
}

func TestManagedEvictor(t *testing.T) {
	c := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.StartManagedEvictor(ctx, 10*time.Millisecond)
	c.Set("temp", "val", 5*time.Millisecond)

	time.Sleep(30 * time.Millisecond)
	c.mu.RLock()
	_, exists := c.items["temp"]
	c.mu.RUnlock()

	if exists {
		t.Fatalf("expected temp to be evicted by managed evictor")
	}
}

