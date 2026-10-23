package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

const tokenLifetime = 15 * time.Minute

func randomOpaqueArchiveID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

type Authenticator interface {
	Issue(passkey string) (token string, expiresAt time.Time, err error)
	Verify(token string, now time.Time) bool
}

type HMACTokenAuthenticator struct {
	passkey    []byte
	signingKey []byte
}

func NewHMACTokenAuthenticator(passkey, signingKey string) (*HMACTokenAuthenticator, error) {
	if strings.TrimSpace(passkey) == "" {
		return nil, errors.New("admin passkey is required")
	}
	if len(signingKey) < 32 {
		return nil, errors.New("token signing key must be at least 32 bytes")
	}
	return &HMACTokenAuthenticator{
		passkey:    []byte(passkey),
		signingKey: []byte(signingKey),
	}, nil
}

func (a *HMACTokenAuthenticator) Issue(passkey string) (string, time.Time, error) {
	if a == nil || !constantTimeSecretEqual(passkey, string(a.passkey)) {
		return "", time.Time{}, errors.New("invalid passkey")
	}

	expiresAt := time.Now().UTC().Add(tokenLifetime)
	payload := make([]byte, 8+32)
	binary.BigEndian.PutUint64(payload[:8], uint64(expiresAt.UnixNano()))
	if _, err := rand.Read(payload[8:]); err != nil {
		return "", time.Time{}, err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, a.signingKey)
	_, _ = mac.Write(payload)
	token := encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return token, expiresAt, nil
}

func (a *HMACTokenAuthenticator) Verify(token string, now time.Time) bool {
	if a == nil || token == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) != 40 {
		return false
	}
	providedMAC, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(providedMAC) != sha256.Size {
		return false
	}

	mac := hmac.New(sha256.New, a.signingKey)
	_, _ = mac.Write(payload)
	if !hmac.Equal(providedMAC, mac.Sum(nil)) {
		return false
	}
	expiresAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[:8])))
	return expiresAt.After(now)
}

func constantTimeSecretEqual(candidate, expected string) bool {
	candidateHash := sha256.Sum256([]byte(candidate))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(candidateHash[:], expectedHash[:]) == 1
}
