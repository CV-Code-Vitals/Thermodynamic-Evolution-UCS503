package consumer

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/external-bench/queue-worker/internal/producer"
)

func TestConsumerProcessing(t *testing.T) {
	var processed int32
	c := NewConsumer(2, func(ctx context.Context, msg producer.Message) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	in := make(chan producer.Message, 5)
	for i := 0; i < 5; i++ {
		in <- producer.Message{ID: "msg"}
	}
	close(in)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c.Start(ctx, in)

	if atomic.LoadInt32(&processed) != 5 {
		t.Fatalf("expected 5 messages processed, got %d", processed)
	}
}

func TestSafeParseRawPayload(t *testing.T) {
	valid := []byte(`{"status":"ok"}`)
	res, err := SafeParseRawPayload(valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["status"] != "ok" {
		t.Errorf("expected status ok, got %v", res["status"])
	}

	invalid := []byte(`{not json}`)
	_, err = SafeParseRawPayload(invalid)
	if err == nil {
		t.Fatalf("expected error for invalid json, got nil")
	}
}

