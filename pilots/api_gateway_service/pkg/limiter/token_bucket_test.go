package limiter

import "testing"

func TestTokenBucket(t *testing.T) {
	tb := NewTokenBucket(2)
	if !tb.Allow() {
		t.Errorf("Expected first token allowed")
	}
	if !tb.Allow() {
		t.Errorf("Expected second token allowed")
	}
	if tb.Allow() {
		t.Errorf("Expected third token rejected")
	}
}
