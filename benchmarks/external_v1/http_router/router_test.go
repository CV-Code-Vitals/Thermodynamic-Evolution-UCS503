package httprouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRouterBasicRoute(t *testing.T) {
	r := New()
	called := false
	r.GET("/hello", func(c *Context) {
		called = true
		c.Status(http.StatusOK)
		c.Writer.Write([]byte("world"))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if !called {
		t.Fatalf("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "world" {
		t.Errorf("expected 'world', got '%s'", w.Body.String())
	}
}

func TestRouterParams(t *testing.T) {
	r := New()
	var userID string
	r.GET("/users/:id", func(c *Context) {
		userID = c.Param("id")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if userID != "42" {
		t.Errorf("expected param id=42, got '%s'", userID)
	}
}

func TestContextTimeoutExecution(t *testing.T) {
	c := NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	err := c.ExecuteWithTimeout(50*time.Millisecond, func(ctx context.Context) error {
		select {
		case <-time.After(10 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err != nil {
		t.Errorf("expected timeout execution to succeed: %v", err)
	}
}

