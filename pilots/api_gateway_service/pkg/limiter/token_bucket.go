package limiter

import (
	"sync"
	"time"
)

// TokenBucketRateLimiter implements in-memory token rate limiting.
type TokenBucketRateLimiter struct {
	mu         sync.Mutex
	capacity   int
	tokens     int
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket instance.
func NewTokenBucket(capacity int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is permitted under quota.
func (tb *TokenBucketRateLimiter) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}
