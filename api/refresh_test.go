package main

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRefreshHandlerRenewsActiveBearerToken(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	request.Header.Set("Origin", "https://portal.example")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("refresh returned %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload struct {
		Status      string    `json:"status"`
		Token       string    `json:"token"`
		ExpiresAt   time.Time `json:"expiresAt"`
		ExpiresInMs int64     `json:"expiresInMs"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Status != "authenticated" || payload.Token == "" || payload.Token == token || payload.ExpiresAt.IsZero() ||
		payload.ExpiresInMs <= 0 || payload.ExpiresInMs > tokenLifetime.Milliseconds() {
		t.Fatalf("refresh returned invalid authentication payload: %#v", payload)
	}
	if !auth.Verify(payload.Token, time.Now()) {
		t.Fatal("refresh returned a token that fails verification")
	}
}

func TestLoginHandlerReportsServerRelativeSessionLifetime(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"passkey":"correct-passkey"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://portal.example")
	response := httptest.NewRecorder()
	newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("login returned %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload struct {
		Status      string    `json:"status"`
		Token       string    `json:"token"`
		ExpiresAt   time.Time `json:"expiresAt"`
		ExpiresInMs int64     `json:"expiresInMs"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Status != "authenticated" || payload.Token == "" || payload.ExpiresAt.IsZero() ||
		payload.ExpiresInMs <= 0 || payload.ExpiresInMs > tokenLifetime.Milliseconds() {
		t.Fatalf("login returned invalid authentication payload: %#v", payload)
	}
}

func TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}
	handler := newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}})
	const requestCount = 12
	responses := make([]*httptest.ResponseRecorder, requestCount)
	var wait sync.WaitGroup
	for i := range responses {
		wait.Add(1)
		go func(i int) {
			defer wait.Done()
			request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
			request.Header.Set("Origin", "https://portal.example")
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			responses[i] = response
		}(i)
	}
	wait.Wait()

	for i, response := range responses {
		if response.Code != http.StatusOK {
			t.Fatalf("parallel refresh %d returned %d: %s", i, response.Code, response.Body.String())
		}
		var payload struct {
			Status string `json:"status"`
			Token  string `json:"token"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Status != "authenticated" || payload.Token == "" || !auth.Verify(payload.Token, time.Now()) {
			t.Fatalf("parallel refresh %d returned an invalid token", i)
		}
		if i > 0 {
			for previous := 0; previous < i; previous++ {
				var earlier struct {
					Token string `json:"token"`
				}
				if err := json.Unmarshal(responses[previous].Body.Bytes(), &earlier); err != nil {
					t.Fatal(err)
				}
				if payload.Token == earlier.Token {
					t.Fatalf("parallel refresh %d reused another response token", i)
				}
			}
		}
	}
}

func TestRefreshHandlerRejectsMissingBearerEvenWithPasskey(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/refresh", strings.NewReader(`{"passkey":"correct-passkey"}`))
	request.Header.Set("Origin", "https://portal.example")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("refresh with a passkey body returned %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestRefreshHandlerRejectsPasskeyBody(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.Issue("correct-passkey")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/refresh", strings.NewReader(`{"passkey":"correct-passkey"}`))
	request.Header.Set("Origin", "https://portal.example")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("refresh accepted a passkey body with status %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestRefreshHandlerRejectsInvalidLegacyAndDuplicateBearerTokens(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	legacyPayload := make([]byte, 40)
	binary.BigEndian.PutUint64(legacyPayload[:8], uint64(time.Now().Add(time.Minute).UnixNano()))
	legacyToken := signedTokenForTest(auth, legacyPayload)

	for name, headers := range map[string][]string{
		"invalid":   {"Bearer invalid"},
		"legacy":    {"Bearer " + legacyToken},
		"duplicate": {"Bearer first", "Bearer second"},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
			request.Header.Set("Origin", "https://portal.example")
			for _, value := range headers {
				request.Header.Add("Authorization", value)
			}
			response := httptest.NewRecorder()
			newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("refresh returned %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRefreshHandlerRequiresPostAndTrustedOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	auth, err := NewHMACTokenAuthenticator("correct-passkey", strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{name: "method", method: http.MethodGet, origin: "https://portal.example", want: http.StatusMethodNotAllowed},
		{name: "missing origin", method: http.MethodPost, want: http.StatusForbidden},
		{name: "untrusted origin", method: http.MethodPost, origin: "https://attacker.example", want: http.StatusForbidden},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(testCase.method, "/api/refresh", nil)
			if testCase.origin != "" {
				request.Header.Set("Origin", testCase.origin)
			}
			response := httptest.NewRecorder()
			newRouter(Dependencies{Auth: auth, Archives: &testArchiveService{}}).ServeHTTP(response, request)
			if response.Code != testCase.want {
				t.Fatalf("refresh returned %d, want %d", response.Code, testCase.want)
			}
		})
	}
}
