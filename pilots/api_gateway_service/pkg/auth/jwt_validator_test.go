package auth

import "testing"

func TestJWTValidator(t *testing.T) {
	v := NewJWTValidator("https://auth.enterprise.com")
	tok, err := v.ValidateHeader("Bearer eyJhbGciOiJIUzI1NiJ9.abc.xyz")
	if err != nil || tok == "" {
		t.Errorf("Expected valid token extraction")
	}
	_, err = v.ValidateHeader("Invalid")
	if err == nil {
		t.Errorf("Expected error for invalid header")
	}
}
