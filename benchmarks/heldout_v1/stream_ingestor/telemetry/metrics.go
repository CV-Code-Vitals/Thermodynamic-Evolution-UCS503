package telemetry

import (
	"fmt"
	"sync"
)

type Collector struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewCollector() *Collector {
	return &Collector{counts: make(map[string]int)}
}

// Neg_Control_IncrementRecord handles thread-safe counter update
func (c *Collector) IncrementRecord(metric string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[metric]++
}

// Neg_Control_FormatStats cleanly formats stats with no defect
func (c *Collector) FormatStats() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("total_keys=%d", len(c.counts))
}
