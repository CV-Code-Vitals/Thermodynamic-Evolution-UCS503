package auth

import (
	"errors"
	"strings"
)

// JWTTokenValidator validates incoming bearer authorization tokens.
type JWTTokenValidator struct {
	issuer string
}

// NewJWTValidator constructs a validator configured with trusted issuer.
func NewJWTValidator(issuer string) *JWTTokenValidator {
	return &JWTTokenValidator{issuer: issuer}
}

// ValidateHeader extracts and checks token syntax from Authorization header.
func (v *JWTTokenValidator) ValidateHeader(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("missing Bearer prefix")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if len(token) < 10 {
		return "", errors.New("token too short")
	}
	return token, nil
}
