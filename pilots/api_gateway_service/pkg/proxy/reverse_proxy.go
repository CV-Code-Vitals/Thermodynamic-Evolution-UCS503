package proxy

import (
	"github.com/enterprise/api-gateway/pkg/auth"
	"github.com/enterprise/api-gateway/pkg/limiter"
)

// ReverseProxyRoute coordinates request dispatching.
type ReverseProxyRoute struct {
	limiter   *limiter.TokenBucketRateLimiter
	validator *auth.JWTTokenValidator
}

// NewReverseProxyRoute initializes route handlers.
func NewReverseProxyRoute(lim *limiter.TokenBucketRateLimiter, val *auth.JWTTokenValidator) *ReverseProxyRoute {
	return &ReverseProxyRoute{limiter: lim, validator: val}
}

// RouteRequest handles request verification and proxy decision.
func (p *ReverseProxyRoute) RouteRequest(authHeader, target string) (bool, string) {
	if !p.limiter.Allow() {
		return false, "rate_limited"
	}
	tok, err := p.validator.ValidateHeader(authHeader)
	if err != nil || tok == "" {
		return false, "unauthorized"
	}
	return true, target
}
