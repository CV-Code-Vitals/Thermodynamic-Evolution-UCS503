package retry

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry tuning parameters.
type Config struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultConfig returns reasonable defaults.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		InitialWait: 50 * time.Millisecond,
		MaxWait:     500 * time.Millisecond,
	}
}

// Retry executes fn with exponential backoff (clean implementation).
func Retry(ctx context.Context, cfg Config, fn func() error) error {
	wait := cfg.InitialWait
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during retry: %w", err)
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < cfg.MaxAttempts {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return ctx.Err()
			}
			wait *= 2
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
	}

	return fmt.Errorf("exceeded %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// RetryWithTimerLeak executes fn with a timer created per attempt.
// DEFECT: Timer resource leak - time.NewTimer is allocated on each loop iteration without deferring or calling timer.Stop().
func RetryWithTimerLeak(attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		timer := time.NewTimer(20 * time.Millisecond)
		// Missing: timer.Stop() before discarding or looping
		<-timer.C
	}
	return err
}

