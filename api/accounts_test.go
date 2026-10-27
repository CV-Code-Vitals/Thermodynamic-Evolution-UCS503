package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const accountTestOrigin = "http://127.0.0.1:18100"

func testAccountService(t *testing.T) (*githubAccountService, http.Handler) {
	t.Helper()
	t.Setenv("APP_ENV", "development")
	t.Setenv("CORS_ALLOWED_ORIGINS", accountTestOrigin)
	t.Setenv("GITHUB_CLIENT_ID", "fixture-client")
	t.Setenv("GITHUB_CALLBACK_URL", accountTestOrigin+"/api/auth/github/callback")
	t.Setenv("GITHUB_FRONTEND_URL", accountTestOrigin+"/graph-visualizer/")
	service, err := accountServiceFromEnvironment(&localAccountStore{path: filepath.Join(t.TempDir(), "accounts.json")}, "fixture-secret")
	if err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if r.Method != http.MethodPost {
				t.Error("OAuth code exchange must use POST")
			}
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.PostForm.Get("client_secret") != "fixture-secret" || r.PostForm.Get("redirect_uri") != service.callbackURL || len(r.PostForm.Get("code_verifier")) != 64 {
				t.Error("missing server secret, redirect binding, or PKCE verifier")
			}
			accountJSON(w, http.StatusOK, map[string]string{"access_token": "provider-token-" + r.PostForm.Get("code"), "token_type": "bearer"})
		case "/user":
			if r.Header.Get("Authorization") == "Bearer provider-token-bob" {
				accountJSON(w, http.StatusOK, map[string]any{"id": 202, "login": "bob"})
			} else {
				accountJSON(w, http.StatusOK, map[string]any{"id": 101, "login": "alice"})
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(provider.Close)
	service.tokenURL, service.userURL = provider.URL+"/token", provider.URL+"/user"
	admin, err := NewHMACTokenAuthenticator("fixture-admin", strings.Repeat("k", 32))
	if err != nil {
		t.Fatal(err)
	}
	return service, newRouter(Dependencies{Accounts: service, Auth: admin})
}

func accountCall(handler http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if method != http.MethodGet {
		request.Header.Set("Origin", accountTestOrigin)
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func beginAccountSignIn(t *testing.T, handler http.Handler) (string, *http.Cookie, string) {
	t.Helper()
	response := accountCall(handler, http.MethodGet, "/api/auth/github/start?handoff_challenge="+browserHandoffChallenge(strings.Repeat("b", 64)), "", "")
	if response.Code != http.StatusFound {
		t.Fatalf("start: %d %s", response.Code, response.Body.String())
	}
	destination, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if destination.Host != "github.com" || destination.Query().Get("scope") != "" || destination.Query().Get("code_challenge_method") != "S256" || len(destination.Query().Get("code_challenge")) != 43 {
		t.Fatalf("unexpected authorization request: %s", destination)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].MaxAge != 600 {
		t.Fatalf("invalid state cookie: %+v", cookies)
	}
	return destination.Query().Get("state"), cookies[0], destination.Query().Get("code_challenge")
}

func finishAccountSignIn(t *testing.T, handler http.Handler, state, code string, cookie *http.Cookie) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state="+state+"&code="+code, nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("callback: %d %s", response.Code, response.Body.String())
	}
	destination, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	fragment, _ := url.ParseQuery(destination.Fragment)
	ticket := fragment.Get("auth_ticket")
	if !accountTokenPattern.MatchString(ticket) || strings.Contains(response.Header().Get("Location"), "provider-token") {
		t.Fatal("callback must redirect with a one-time ticket only")
	}
	return ticket
}

func signInAccount(t *testing.T, handler http.Handler, code string) string {
	t.Helper()
	state, cookie, _ := beginAccountSignIn(t, handler)
	ticket := finishAccountSignIn(t, handler, state, code, cookie)
	wrong := accountCall(handler, http.MethodPost, "/api/account/session", "", `{"ticket":"`+ticket+`","verifier":"`+strings.Repeat("c", 64)+`"}`)
	if wrong.Code != http.StatusUnauthorized {
		t.Fatal("another browser could redeem the sign-in ticket")
	}
	response := accountCall(handler, http.MethodPost, "/api/account/session", "", `{"ticket":"`+ticket+`","verifier":"`+strings.Repeat("b", 64)+`"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("session: %d %s", response.Code, response.Body.String())
	}
	var session struct {
		Token string     `json:"token"`
		User  GitHubUser `json:"user"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if !accountTokenPattern.MatchString(session.Token) || session.User.ID == "" {
		t.Fatal("invalid personal session")
	}
	if repeated := accountCall(handler, http.MethodPost, "/api/account/session", "", `{"ticket":"`+ticket+`","verifier":"`+strings.Repeat("b", 64)+`"}`); repeated.Code != http.StatusUnauthorized {
		t.Fatal("sign-in ticket was reusable")
	}
	return session.Token
}

func TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart(t *testing.T) {
	service, handler := testAccountService(t)
	alice, bob := signInAccount(t, handler, "alice"), signInAccount(t, handler, "bob")
	response := accountCall(handler, http.MethodPost, "/api/account/repositories", alice, `{"url":"https://github.com/Zylus08/SleepApnea.git"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("save: %d %s", response.Code, response.Body.String())
	}
	var saved SavedRepository
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.URL != "https://github.com/Zylus08/SleepApnea" || saved.Name != "Zylus08/SleepApnea" {
		t.Fatalf("unexpected saved URL: %+v", saved)
	}
	response = accountCall(handler, http.MethodPost, "/api/account/repositories", alice, `{"url":"https://github.com/zylus08/sleepapnea/"}`)
	if response.Code != http.StatusOK {
		t.Fatal(response.Body.String())
	}
	if body := accountCall(handler, http.MethodGet, "/api/account/repositories", bob, "").Body.String(); strings.TrimSpace(body) != "[]" {
		t.Fatalf("another account saw saved links: %s", body)
	}
	if response := accountCall(handler, http.MethodDelete, "/api/account/repositories/"+saved.ID, bob, ""); response.Code != http.StatusNoContent {
		t.Fatal(response.Body.String())
	}
	var repositories []SavedRepository
	if err := json.Unmarshal(accountCall(handler, http.MethodGet, "/api/account/repositories", alice, "").Body.Bytes(), &repositories); err != nil || len(repositories) != 1 {
		t.Fatal("cross-account deletion or duplicate saves corrupted Alice's list")
	}
	oldStore := service.store.(*localAccountStore)
	service.store = &localAccountStore{path: oldStore.path}
	if response := accountCall(handler, http.MethodGet, "/api/account/me", alice, ""); response.Code != http.StatusOK {
		t.Fatal("session did not survive API restart")
	}
	info, err := os.Stat(oldStore.path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("account file must be private")
	}
	contents, _ := os.ReadFile(oldStore.path)
	if strings.Contains(string(contents), alice) || strings.Contains(string(contents), "provider-token") || strings.Contains(string(contents), "fixture-secret") {
		t.Fatal("raw bearer or GitHub credentials persisted to disk")
	}
	if response := accountCall(handler, http.MethodPost, "/api/account/logout", alice, ""); response.Code != http.StatusNoContent {
		t.Fatal(response.Body.String())
	}
	if response := accountCall(handler, http.MethodGet, "/api/account/me", alice, ""); response.Code != http.StatusUnauthorized {
		t.Fatal("logged-out token remained valid")
	}
	alice = signInAccount(t, handler, "alice")
	if body := accountCall(handler, http.MethodGet, "/api/account/repositories", alice, "").Body.String(); !strings.Contains(body, "sleepapnea") {
		t.Fatal("saved links did not survive sign-out and reauthentication")
	}
}

func TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions(t *testing.T) {
	service, handler := testAccountService(t)
	state, cookie, challenge := beginAccountSignIn(t, handler)
	record, err := service.store.GetTransient(context.Background(), accountKey("state", state))
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(record.Verifier))
	if base64.RawURLEncoding.EncodeToString(hash[:]) != challenge {
		t.Fatal("PKCE challenge is not bound to stored verifier")
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state="+state+"&code=alice", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatal("OAuth callback without browser state cookie was accepted")
	}
	finishAccountSignIn(t, handler, state, "alice", cookie)
	request.AddCookie(cookie)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatal("OAuth state was reusable")
	}
	token := signInAccount(t, handler, "alice")
	if response := accountCall(handler, http.MethodPost, "/api/upload", token, `{}`); response.Code != http.StatusUnauthorized {
		t.Fatalf("personal session granted administrator access: %d", response.Code)
	}
	admin, _ := NewHMACTokenAuthenticator("fixture-admin", strings.Repeat("k", 32))
	adminToken, _, _ := admin.Issue("fixture-admin")
	if response := accountCall(handler, http.MethodGet, "/api/account/me", adminToken, ""); response.Code != http.StatusUnauthorized {
		t.Fatal("admin token was accepted as a personal identity")
	}
	request = httptest.NewRequest(http.MethodPost, "/api/account/repositories", strings.NewReader(`{"url":"https://github.com/a/b"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Origin", "https://untrusted.example")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatal("untrusted origin mutated a personal account")
	}
	if _, err := service.store.TakeTransient(context.Background(), accountKey("session", token)); err != nil {
		t.Fatal(err)
	}
	if err := service.store.PutTransient(context.Background(), accountKey("session", token), accountTransient{User: GitHubUser{ID: "101", Login: "alice"}, ExpiresAt: time.Now().Add(-time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	if response := accountCall(handler, http.MethodGet, "/api/account/me", token, ""); response.Code != http.StatusUnauthorized {
		t.Fatal("expired session was accepted before database TTL deletion")
	}
}

func TestSavedRepositoryValidationAndDisabledGitHubSignIn(t *testing.T) {
	for _, raw := range []string{"http://github.com/a/b", "https://example.com/a/b", "https://user:secret@github.com/a/b", "https://github.com/a/b?token=secret", "https://github.com/a/b#fragment", "https://github.com/a/b/tree/main", "https://github.com/a/%2fb", "https://github.com/../b", "https://github.com/a/.git"} {
		if _, err := parseSavedRepository(raw); err == nil {
			t.Errorf("accepted invalid repository: %s", raw)
		}
	}
	handler := newRouter(Dependencies{})
	response := accountCall(handler, http.MethodGet, "/api/account/config", "", "")
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"github_enabled":false`) {
		t.Fatal("disabled account feature not reported")
	}
	if response := accountCall(handler, http.MethodGet, "/api/auth/github/start", "", ""); response.Code != http.StatusServiceUnavailable {
		t.Fatal("unconfigured GitHub sign-in was available")
	}
}

func TestGitHubAccountConfigurationBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, environment, callback, frontend, secret string
		valid                                         bool
	}{
		{"local", "development", accountTestOrigin + "/api/auth/github/callback", accountTestOrigin + "/graph-visualizer/", "fixture-secret", true},
		{"production HTTPS", "production", "https://api.example.com/api/auth/github/callback", "https://portal.example.com/graph-visualizer/", "fixture-secret", true},
		{"production HTTP", "production", accountTestOrigin + "/api/auth/github/callback", accountTestOrigin + "/graph-visualizer/", "fixture-secret", false},
		{"untrusted return origin", "development", accountTestOrigin + "/api/auth/github/callback", "https://untrusted.example/", "fixture-secret", false},
		{"wrong callback", "development", accountTestOrigin + "/callback", accountTestOrigin + "/graph-visualizer/", "fixture-secret", false},
		{"return query", "development", accountTestOrigin + "/api/auth/github/callback", accountTestOrigin + "/?redirect=other", "fixture-secret", false},
		{"missing secret", "development", accountTestOrigin + "/api/auth/github/callback", accountTestOrigin + "/graph-visualizer/", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("APP_ENV", test.environment)
			t.Setenv("CORS_ALLOWED_ORIGINS", accountTestOrigin+",https://portal.example.com")
			t.Setenv("GITHUB_CLIENT_ID", "fixture-client")
			t.Setenv("GITHUB_CALLBACK_URL", test.callback)
			t.Setenv("GITHUB_FRONTEND_URL", test.frontend)
			service, err := accountServiceFromEnvironment(&localAccountStore{path: filepath.Join(t.TempDir(), "accounts.json")}, test.secret)
			if (err == nil) != test.valid || (service != nil) != test.valid {
				t.Fatalf("configuration valid=%v, service=%v, error=%v", test.valid, service != nil, err)
			}
		})
	}
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET_PARAMETER", "")
	if service, err := accountServiceFromEnvironment(nil, ""); service != nil || err != nil {
		t.Fatal("unconfigured accounts should be disabled")
	}
	if _, err := accountServiceFromEnvironment(nil, "fixture-secret"); err == nil {
		t.Fatal("secret without client ID should fail configuration")
	}
}

func exerciseAccountStore(t *testing.T, store accountStore) {
	t.Helper()
	ctx := context.Background()
	record := accountTransient{User: GitHubUser{ID: "101", Login: "alice"}, ExpiresAt: time.Now().Add(time.Hour).Unix()}
	if err := store.PutTransient(ctx, "ticket:test", record); err != nil {
		t.Fatal(err)
	}
	var consumed atomic.Int32
	var group sync.WaitGroup
	for range 12 {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := store.TakeTransient(ctx, "ticket:test"); err == nil {
				consumed.Add(1)
			}
		}()
	}
	group.Wait()
	if consumed.Load() != 1 {
		t.Fatalf("single-use ticket consumed %d times", consumed.Load())
	}
	for i := range 40 {
		repository, _ := parseSavedRepository(fmt.Sprintf("https://github.com/a/repo%d", i))
		if err := store.SaveRepository(ctx, "101", repository); err != nil {
			t.Fatal(err)
		}
	}
	var added atomic.Int32
	for i := range 20 {
		group.Add(1)
		go func() {
			defer group.Done()
			repository, _ := parseSavedRepository(fmt.Sprintf("https://github.com/a/repo%d", i+40))
			err := store.SaveRepository(ctx, "101", repository)
			if err == nil {
				added.Add(1)
			} else if err != errSavedRepositoryLimit {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	if added.Load() != 10 {
		t.Fatalf("concurrent saves exceeded the account quota: %d additions", added.Load())
	}
	repository, _ := parseSavedRepository("https://github.com/a/extra")
	if err := store.SaveRepository(ctx, "101", repository); err != errSavedRepositoryLimit {
		t.Fatalf("expected repository limit, got %v", err)
	}
	repository, _ = parseSavedRepository("https://github.com/a/repo0")
	if err := store.SaveRepository(ctx, "101", repository); err != nil {
		t.Fatal("updating a saved repository should work at the limit")
	}
	if err := store.DeleteRepository(ctx, "202", repository.ID); err != nil {
		t.Fatal(err)
	}
	repositories, err := store.ListRepositories(ctx, "101")
	if err != nil || len(repositories) != savedRepositoryLimit {
		t.Fatalf("wrong account repository list: %d %v", len(repositories), err)
	}
}

func TestLocalAccountStoreAtomicTicketsAndRepositoryLimit(t *testing.T) {
	exerciseAccountStore(t, &localAccountStore{path: filepath.Join(t.TempDir(), "accounts.json")})
}
