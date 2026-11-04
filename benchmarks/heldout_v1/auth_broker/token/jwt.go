package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// Neg_Control_VerifySignature purely functional crypto verification
func VerifySignature(message, signature, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(message); err != nil {
		return false
	}
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(signature, expectedMAC)
}

// Neg_Control_ParseTokenSafe parses and validates with full error checks
func ParseTokenSafe(raw string, secret []byte) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding: %w", err)
	}
	if len(data) < 32 {
		return nil, fmt.Errorf("token too short")
	}
	return data, nil
}
