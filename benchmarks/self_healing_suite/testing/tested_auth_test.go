package testing

import (
	"testing"
)

func TestValidateToken(t *testing.T) {
	if ValidateToken("short") {
		t.Errorf("expected false for short token")
	}
	if ValidateToken("Basic 1234567890abcdef") {
		t.Errorf("expected false for non-bearer token")
	}
	if !ValidateToken("Bearer 1234567890abcdef") {
		t.Errorf("expected true for valid bearer token")
	}
}

