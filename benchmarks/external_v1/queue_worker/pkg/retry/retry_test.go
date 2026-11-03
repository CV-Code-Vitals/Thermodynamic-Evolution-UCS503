package retry

import (
	"context"
	"errors"
	"testing"
)

func TestRetrySuccess(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), DefaultConfig(), func() error {
		attempts++
		if attempts < 2 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected retry to succeed: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := Retry(ctx, DefaultConfig(), func() error {
		return errors.New("should not succeed")
	})

	if err == nil {
		t.Fatalf("expected context cancellation error, got nil")
	}
}
