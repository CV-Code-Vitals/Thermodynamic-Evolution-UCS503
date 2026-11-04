package cacheengine

import "sync/atomic"

// CacheStats tracks hit/miss and operation metrics.
type CacheStats struct {
	Hits   int64
	Misses int64
	Sets   int64
	Dels   int64
}

// RecordHit atomically increments hits.
func (s *CacheStats) RecordHit() {
	atomic.AddInt64(&s.Hits, 1)
}

// RecordMiss atomically increments misses.
func (s *CacheStats) RecordMiss() {
	atomic.AddInt64(&s.Misses, 1)
}

// RecordSet atomically increments sets.
func (s *CacheStats) RecordSet() {
	atomic.AddInt64(&s.Sets, 1)
}

// RecordDel atomically increments dels.
func (s *CacheStats) RecordDel() {
	atomic.AddInt64(&s.Dels, 1)
}

// Snapshot returns a copy of current metrics.
func (s *CacheStats) Snapshot() (hits, misses, sets, dels int64) {
	return atomic.LoadInt64(&s.Hits),
		atomic.LoadInt64(&s.Misses),
		atomic.LoadInt64(&s.Sets),
		atomic.LoadInt64(&s.Dels)
}

