package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type sandboxLifecycleStub struct {
	events      *[]string
	acquireErr  error
	deadline    time.Time
	hasDeadline bool
}

func (s *sandboxLifecycleStub) Acquire(ctx context.Context) (func() error, error) {
	if s.acquireErr != nil {
		return nil, s.acquireErr
	}
	s.deadline, s.hasDeadline = ctx.Deadline()
	*s.events = append(*s.events, "acquire")
	return func() error {
		*s.events = append(*s.events, "release")
		return nil
	}, nil
}

func TestSandboxLifecycleMiddleware_HoldsLeaseForRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	events := []string{}
	controller := &sandboxLifecycleStub{events: &events}
	router := gin.New()
	router.Use(sandboxLifecycleMiddleware(controller))
	router.POST("/evolve", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok || !controller.hasDeadline || !deadline.Equal(controller.deadline) {
			t.Errorf("Evolution request deadline = %v (%t), lifecycle deadline = %v (%t)", deadline, ok, controller.deadline, controller.hasDeadline)
		}
		events = append(events, "evolve")
		c.Status(http.StatusOK)
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/evolve", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if want := []string{"acquire", "evolve", "release"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSandboxLifecycleMiddleware_FailsClosedOnAdmissionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sandboxLifecycleMiddleware(&sandboxLifecycleStub{
		events:     new([]string),
		acquireErr: errors.New("EC2 denied"),
	}))
	router.POST("/evolve", func(c *gin.Context) {
		t.Fatal("Evolution handler ran without a sandbox lease")
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/evolve", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
