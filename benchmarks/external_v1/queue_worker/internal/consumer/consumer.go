package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/external-bench/queue-worker/internal/producer"
)

// HandlerFunc processes an individual queue message.
type HandlerFunc func(ctx context.Context, msg producer.Message) error

// Consumer runs a pool of workers consuming messages from a channel.
type Consumer struct {
	concurrency int
	handler     HandlerFunc
}

// NewConsumer creates a consumer with a fixed worker count.
func NewConsumer(concurrency int, handler HandlerFunc) *Consumer {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Consumer{
		concurrency: concurrency,
		handler:     handler,
	}
}

// Start begins processing messages from in channel until context cancelled (clean implementation).
func (c *Consumer) Start(ctx context.Context, in <-chan producer.Message) {
	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-in:
					if !ok {
						return
					}
					_ = c.handler(ctx, msg)
				}
			}
		}()
	}
	wg.Wait()
}

// ParseRawPayload parses raw JSON byte slices into target map.
// DEFECT: json.Unmarshal error is ignored without error propagation or logging.
func ParseRawPayload(data []byte) map[string]interface{} {
	var res map[string]interface{}
	// DEFECT: unhandled error on json.Unmarshal
	json.Unmarshal(data, &res)
	return res
}

// SafeParseRawPayload parses raw JSON with proper error handling (clean implementation).
func SafeParseRawPayload(data []byte) (map[string]interface{}, error) {
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("failed to parse message payload: %w", err)
	}
	return res, nil
}

