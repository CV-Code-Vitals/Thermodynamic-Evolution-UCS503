package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestEvolutionAdmissionController_CapacityAndRelease(t *testing.T) {
	limiter := NewEvolutionAdmissionController(2, 50*time.Millisecond)

	ctx := context.Background()

	// Acquire slot 1
	rel1, err := limiter.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire slot 1 failed: %v", err)
	}
	if limiter.CurrentLoad() != 1 {
		t.Errorf("expected load 1, got %d", limiter.CurrentLoad())
	}

	// Acquire slot 2
	rel2, err := limiter.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire slot 2 failed: %v", err)
	}
	if limiter.CurrentLoad() != 2 {
		t.Errorf("expected load 2, got %d", limiter.CurrentLoad())
	}

	// Slot 3 should time out with ErrOverloaded
	_, err = limiter.Acquire(ctx)
	if !errors.Is(err, ErrOverloaded) {
		t.Fatalf("expected ErrOverloaded on capacity exhaustion, got: %v", err)
	}

	// Release slot 1
	rel1()
	if limiter.CurrentLoad() != 1 {
		t.Errorf("expected load 1 after rel1, got %d", limiter.CurrentLoad())
	}

	// Now slot 3 should succeed
	rel3, err := limiter.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire slot 3 failed after release: %v", err)
	}
	if limiter.CurrentLoad() != 2 {
		t.Errorf("expected load 2, got %d", limiter.CurrentLoad())
	}

	rel2()
	rel3()
	if limiter.CurrentLoad() != 0 {
		t.Errorf("expected load 0 after all releases, got %d", limiter.CurrentLoad())
	}
}

func TestEvolutionAdmissionController_ContextCancellationWhileWaiting(t *testing.T) {
	limiter := NewEvolutionAdmissionController(1, 5*time.Second)

	// Occupy sole slot
	rel1, err := limiter.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	defer rel1()

	// Attempt second acquire with context that cancels after 20ms
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = limiter.Acquire(ctx)
	if !errors.Is(err, ErrLimiterCancelled) {
		t.Fatalf("expected ErrLimiterCancelled, got: %v", err)
	}
}

func TestEvolutionAdmissionController_IdempotentRelease(t *testing.T) {
	limiter := NewEvolutionAdmissionController(2, 50*time.Millisecond)

	rel, err := limiter.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}

	// Multiple release calls must not underflow capacity
	rel()
	rel()
	rel()

	if limiter.CurrentLoad() != 0 {
		t.Errorf("expected load 0, got %d", limiter.CurrentLoad())
	}
}

func TestEvolutionAdmissionController_ConcurrentAccess(t *testing.T) {
	capacity := 4
	limiter := NewEvolutionAdmissionController(capacity, 500*time.Millisecond)

	workers := 16
	var wg sync.WaitGroup
	wg.Add(workers)

	acquiredCount := 0
	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			rel, err := limiter.Acquire(context.Background())
			if err == nil {
				mu.Lock()
				acquiredCount++
				mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				rel()
			}
		}()
	}

	wg.Wait()

	if limiter.CurrentLoad() != 0 {
		t.Errorf("expected all slots freed, got active load %d", limiter.CurrentLoad())
	}
	if acquiredCount == 0 {
		t.Error("expected at least some workers to acquire capacity")
	}
}
