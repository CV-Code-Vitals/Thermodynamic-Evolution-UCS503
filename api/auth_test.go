package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"testing"
	"time"
)

type testTokenRenewer interface {
	Renew(token string, now time.Time) (string, time.Time, error)
}

func tokenRenewerForTest(t *testing.T, auth *HMACTokenAuthenticator) testTokenRenewer {
	t.Helper()
	renewer, ok := any(auth).(testTokenRenewer)
	if !ok {
		t.Fatal("HMACTokenAuthenticator does not support token renewal")
	}
	return renewer
}

func signedTokenForTest(auth *HMACTokenAuthenticator, payload []byte) string {
	mac := hmac.New(sha256.New, auth.signingKey)
	_, _ = mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestHMACTokenAuthenticatorIssuesAndVerifiesToken(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}

	token, expiresAt, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatalf("Issue returned an error for a valid passkey: %v", err)
	}
	if token == "" {
		t.Fatal("Issue returned an empty token")
	}
	payloadPart := strings.Split(token, ".")[0]
	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		t.Fatalf("Issue returned an invalid payload: %v", err)
	}
	if len(payload) != 49 || payload[0] != 1 {
		t.Fatalf("Issue payload has version/length %d/%d, want 1/49", payload[0], len(payload))
	}
	if want := 15 * time.Minute; time.Until(expiresAt) > want || time.Until(expiresAt) <= want-time.Second {
		t.Fatalf("token expiry is not approximately 15 minutes from now: %s", expiresAt)
	}
	if !auth.Verify(token, time.Now()) {
		t.Fatal("Verify rejected a newly issued token")
	}
}

func TestHMACTokenAuthenticatorRenewsNewTokens(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, originalExpiry, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Add(time.Minute)
	renewedToken, expiresAt, err := tokenRenewerForTest(t, auth).Renew(token, now)
	if err != nil {
		t.Fatalf("Renew rejected an active new-format token: %v", err)
	}
	if renewedToken == token {
		t.Fatal("Renew returned the original token instead of rotating it")
	}
	if !expiresAt.After(originalExpiry) || expiresAt.After(now.Add(tokenLifetime)) {
		t.Fatalf("Renew returned unexpected expiry %s", expiresAt)
	}
	if !auth.Verify(renewedToken, now) {
		t.Fatal("Verify rejected the renewed token")
	}
}

func TestHMACTokenAuthenticatorRejectsInvalidAndExpiredRenewals(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, expiresAt, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}
	renewer := tokenRenewerForTest(t, auth)

	if _, _, err := renewer.Renew(token+"x", time.Now()); err == nil {
		t.Fatal("Renew accepted a tampered token")
	}
	if _, _, err := renewer.Renew(token, expiresAt); err == nil {
		t.Fatal("Renew accepted an expired token")
	}
}

func TestHMACTokenAuthenticatorVerifiesButDoesNotRenewLegacyTokens(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	payload := make([]byte, 40)
	binary.BigEndian.PutUint64(payload[:8], uint64(time.Now().Add(time.Minute).UnixNano()))
	legacyToken := signedTokenForTest(auth, payload)

	if !auth.Verify(legacyToken, time.Now()) {
		t.Fatal("Verify rejected a valid legacy token")
	}
	if _, _, err := tokenRenewerForTest(t, auth).Renew(legacyToken, time.Now()); err == nil {
		t.Fatal("Renew accepted a legacy token")
	}
}

func TestHMACTokenAuthenticatorRejectsRenewalAfterAbsoluteLifetime(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	payload := make([]byte, 49)
	payload[0] = 1
	binary.BigEndian.PutUint64(payload[1:9], uint64(now.Add(-8*time.Hour-time.Minute).UnixNano()))
	binary.BigEndian.PutUint64(payload[9:17], uint64(now.Add(time.Minute).UnixNano()))
	token := signedTokenForTest(auth, payload)
	if !auth.Verify(token, now) {
		t.Fatal("test token should be unexpired so the absolute-age check is exercised")
	}

	if _, _, err := tokenRenewerForTest(t, auth).Renew(token, now); err == nil {
		t.Fatal("Renew accepted a token older than the eight-hour absolute lifetime")
	}
}

func TestHMACTokenAuthenticatorCapsRenewalAtAbsoluteLifetime(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	issuedAt := now.Add(-absoluteTokenLifetime + time.Minute)
	expiresAt := now.Add(time.Minute)
	payload := make([]byte, versionedTokenSize)
	payload[0] = versionedToken
	binary.BigEndian.PutUint64(payload[1:9], uint64(issuedAt.UnixNano()))
	binary.BigEndian.PutUint64(payload[9:17], uint64(expiresAt.UnixNano()))
	token := signedTokenForTest(auth, payload)

	renewedToken, renewedExpiry, err := tokenRenewerForTest(t, auth).Renew(token, now)
	if err != nil {
		t.Fatalf("Renew rejected a token before its absolute expiry: %v", err)
	}
	if !renewedExpiry.Equal(issuedAt.Add(absoluteTokenLifetime)) {
		t.Fatalf("Renew expiry = %s, want absolute limit %s", renewedExpiry, issuedAt.Add(absoluteTokenLifetime))
	}
	if !auth.Verify(renewedToken, now) {
		t.Fatal("Renew returned an invalid token at the absolute lifetime boundary")
	}
	if _, _, err := tokenRenewerForTest(t, auth).Renew(renewedToken, renewedExpiry); err == nil {
		t.Fatal("Renew accepted a token at the absolute expiry boundary")
	}
}

func TestHMACTokenAuthenticatorRejectsInvalidPasskey(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := auth.Issue("incorrect-passkey"); err == nil {
		t.Fatal("Issue accepted an invalid passkey")
	}
}

func TestHMACTokenAuthenticatorRejectsExpiredAndTamperedTokens(t *testing.T) {
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}

	token, expiresAt, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}
	if auth.Verify(token, expiresAt) {
		t.Fatal("Verify accepted a token at its expiry time")
	}
	if auth.Verify(token+"x", time.Now()) {
		t.Fatal("Verify accepted a tampered token")
	}
}
