package proxy

import (
	"testing"
	"github.com/enterprise/api-gateway/pkg/auth"
	"github.com/enterprise/api-gateway/pkg/limiter"
)

func TestReverseProxyRoute(t *testing.T) {
	lim := limiter.NewTokenBucket(5)
	val := auth.NewJWTValidator("https://auth.com")
	route := NewReverseProxyRoute(lim, val)
	ok, _ := route.RouteRequest("Bearer eyJhbGciOiJIUzI1NiJ9.1234567890", "/api/v1/orders")
	if !ok {
		t.Errorf("Expected successful routing")
	}
}
