package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	ErrOverloaded       = errors.New("evolution admission controller: server is at capacity")
	ErrLimiterCancelled = errors.New("evolution admission controller: request cancelled while waiting")
)

// AdmissionController manages capacity for expensive background evolution workflows.
type AdmissionController interface {
	Acquire(ctx context.Context) (func(), error)
	CurrentLoad() int
	Capacity() int
}

// EvolutionAdmissionController manages bounded concurrent evolution requests across HTTP endpoints.
type EvolutionAdmissionController struct {
	sem      chan struct{}
	capacity int
	timeout  time.Duration
	mu       sync.Mutex
	active   int
}

// NewEvolutionAdmissionController constructs an admission controller with specified capacity and timeout.
func NewEvolutionAdmissionController(capacity int, acquireTimeout time.Duration) *EvolutionAdmissionController {
	if capacity <= 0 {
		capacity = 2 // Conservative default: bounds memory, CPU, and container execution
	}
	if acquireTimeout <= 0 {
		acquireTimeout = 3 * time.Second
	}

	return &EvolutionAdmissionController{
		sem:      make(chan struct{}, capacity),
		capacity: capacity,
		timeout:  acquireTimeout,
	}
}

// NewEvolutionAdmissionControllerFromEnv reads EVOLVE_MAX_CONCURRENCY and EVOLVE_ACQUIRE_TIMEOUT_SEC.
func NewEvolutionAdmissionControllerFromEnv() *EvolutionAdmissionController {
	capVal := 2
	if env := os.Getenv("EVOLVE_MAX_CONCURRENCY"); env != "" {
		if c, err := strconv.Atoi(env); err == nil && c > 0 {
			capVal = c
		}
	}

	timeoutVal := 3 * time.Second
	if env := os.Getenv("EVOLVE_ACQUIRE_TIMEOUT_SEC"); env != "" {
		if t, err := strconv.Atoi(env); err == nil && t > 0 {
			timeoutVal = time.Duration(t) * time.Second
		}
	}

	return NewEvolutionAdmissionController(capVal, timeoutVal)
}

// Acquire requests an execution slot with context cancellation and bounded waiting.
func (c *EvolutionAdmissionController) Acquire(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLimiterCancelled, err)
	}

	timer := time.NewTimer(c.timeout)
	defer timer.Stop()

	select {
	case c.sem <- struct{}{}:
		c.mu.Lock()
		c.active++
		c.mu.Unlock()

		var once sync.Once
		release := func() {
			once.Do(func() {
				<-c.sem
				c.mu.Lock()
				c.active--
				c.mu.Unlock()
			})
		}
		return release, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrLimiterCancelled, ctx.Err())

	case <-timer.C:
		return nil, ErrOverloaded
	}
}

func (c *EvolutionAdmissionController) CurrentLoad() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}

func (c *EvolutionAdmissionController) Capacity() int {
	return c.capacity
}

var (
	ErrPublicScanRateLimited  = errors.New("public scan rate limit exceeded")
	ErrPublicScanBusy         = errors.New("public scan capacity is full")
	ErrInvalidTrustedClientIP = errors.New("trusted client IP is invalid")
)

const maxTrackedPublicScanClients = 10_000

type PublicScanController struct {
	mu            sync.Mutex
	lastRequest   map[string]time.Time
	active        int
	maxConcurrent int
	rateInterval  time.Duration
}

func NewPublicScanController(rateInterval time.Duration, maxConcurrent int) *PublicScanController {
	if rateInterval <= 0 {
		rateInterval = time.Minute
	}
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}
	return &PublicScanController{
		lastRequest:   make(map[string]time.Time),
		maxConcurrent: maxConcurrent,
		rateInterval:  rateInterval,
	}
}

func (c *PublicScanController) Acquire(ctx context.Context, clientIP string) (func(), error) {
	if ctx == nil {
		return nil, errors.New("public scan request context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLimiterCancelled, err)
	}
	if net.ParseIP(clientIP) == nil {
		return nil, ErrInvalidTrustedClientIP
	}

	now := time.Now()
	c.mu.Lock()
	for ip, last := range c.lastRequest {
		if now.Sub(last) >= c.rateInterval {
			delete(c.lastRequest, ip)
		}
	}
	if last, ok := c.lastRequest[clientIP]; ok && now.Sub(last) < c.rateInterval {
		c.mu.Unlock()
		return nil, ErrPublicScanRateLimited
	}
	if c.active >= c.maxConcurrent {
		c.mu.Unlock()
		return nil, ErrPublicScanBusy
	}
	if len(c.lastRequest) >= maxTrackedPublicScanClients {
		c.mu.Unlock()
		return nil, ErrPublicScanRateLimited
	}
	c.lastRequest[clientIP] = now
	c.active++
	c.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			c.active--
			c.mu.Unlock()
		})
	}, nil
}
