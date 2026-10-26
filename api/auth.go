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
const absoluteTokenLifetime = 8 * time.Hour

const (
	legacyTokenPayloadSize = 40
	versionedTokenSize     = 49
	versionedToken         = 1
)

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

type TokenRenewer interface {
	Renew(token string, now time.Time) (newToken string, expiresAt time.Time, err error)
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

	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(tokenLifetime)
	payload, err := a.newVersionedPayload(issuedAt, expiresAt)
	if err != nil {
		return "", time.Time{}, err
	}
	return a.signPayload(payload), expiresAt, nil
}

func (a *HMACTokenAuthenticator) Renew(token string, now time.Time) (string, time.Time, error) {
	payload, err := a.decodeAndVerify(token)
	if err != nil || len(payload) != versionedTokenSize || payload[0] != versionedToken {
		return "", time.Time{}, errors.New("token cannot be renewed")
	}

	issuedAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[1:9]))).UTC()
	expiresAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[9:17]))).UTC()
	now = now.UTC()
	if issuedAt.After(now) || !expiresAt.After(now) || !now.Before(issuedAt.Add(absoluteTokenLifetime)) {
		return "", time.Time{}, errors.New("token cannot be renewed")
	}

	nextExpiry := now.Add(tokenLifetime)
	absoluteExpiry := issuedAt.Add(absoluteTokenLifetime)
	if nextExpiry.After(absoluteExpiry) {
		nextExpiry = absoluteExpiry
	}
	payload, err = a.newVersionedPayload(issuedAt, nextExpiry)
	if err != nil {
		return "", time.Time{}, err
	}
	return a.signPayload(payload), nextExpiry, nil
}

func (a *HMACTokenAuthenticator) newVersionedPayload(issuedAt, expiresAt time.Time) ([]byte, error) {
	payload := make([]byte, versionedTokenSize)
	payload[0] = versionedToken
	binary.BigEndian.PutUint64(payload[1:9], uint64(issuedAt.UnixNano()))
	binary.BigEndian.PutUint64(payload[9:17], uint64(expiresAt.UnixNano()))
	if _, err := rand.Read(payload[17:]); err != nil {
		return nil, err
	}
	return payload, nil
}

func (a *HMACTokenAuthenticator) signPayload(payload []byte) string {
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, a.signingKey)
	_, _ = mac.Write(payload)
	return encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *HMACTokenAuthenticator) Verify(token string, now time.Time) bool {
	payload, err := a.decodeAndVerify(token)
	if err != nil {
		return false
	}
	expiresAt, issuedAt, ok := tokenTimes(payload)
	if !ok || (issuedAt != nil && issuedAt.After(now)) {
		return false
	}
	return expiresAt.After(now)
}

func (a *HMACTokenAuthenticator) decodeAndVerify(token string) ([]byte, error) {
	if a == nil || token == "" {
		return nil, errors.New("invalid token")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || (len(payload) != legacyTokenPayloadSize && len(payload) != versionedTokenSize) {
		return nil, errors.New("invalid token")
	}
	if len(payload) == versionedTokenSize && payload[0] != versionedToken {
		return nil, errors.New("invalid token")
	}
	providedMAC, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(providedMAC) != sha256.Size {
		return nil, errors.New("invalid token")
	}

	mac := hmac.New(sha256.New, a.signingKey)
	_, _ = mac.Write(payload)
	if !hmac.Equal(providedMAC, mac.Sum(nil)) {
		return nil, errors.New("invalid token")
	}
	return payload, nil
}

func tokenTimes(payload []byte) (time.Time, *time.Time, bool) {
	switch len(payload) {
	case legacyTokenPayloadSize:
		return time.Unix(0, int64(binary.BigEndian.Uint64(payload[:8]))), nil, true
	case versionedTokenSize:
		if payload[0] != versionedToken {
			return time.Time{}, nil, false
		}
		issuedAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[1:9])))
		expiresAt := time.Unix(0, int64(binary.BigEndian.Uint64(payload[9:17])))
		if !expiresAt.After(issuedAt) {
			return time.Time{}, nil, false
		}
		return expiresAt, &issuedAt, true
	default:
		return time.Time{}, nil, false
	}
}

func constantTimeSecretEqual(candidate, expected string) bool {
	candidateHash := sha256.Sum256([]byte(candidate))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(candidateHash[:], expectedHash[:]) == 1
}
