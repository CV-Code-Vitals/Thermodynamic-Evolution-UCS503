package main

import (
	"strings"
	"testing"
	"time"
)

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
	if want := 15 * time.Minute; time.Until(expiresAt) > want || time.Until(expiresAt) <= want-time.Second {
		t.Fatalf("token expiry is not approximately 15 minutes from now: %s", expiresAt)
	}
	if !auth.Verify(token, time.Now()) {
		t.Fatal("Verify rejected a newly issued token")
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
