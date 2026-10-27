package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const githubSessionLifetime = time.Hour

var accountTokenPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var browserChallengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
var githubLoginPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,39}$`)
var savedPathPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

type githubAccountService struct {
	store                                      accountStore
	clientID, secret, callbackURL, frontendURL string
	authorizeURL, tokenURL, userURL            string
	client                                     *http.Client
}

func accountServiceFromEnvironment(store accountStore, secret string) (*githubAccountService, error) {
	id := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	if id == "" {
		if secret != "" || os.Getenv("GITHUB_CLIENT_SECRET_PARAMETER") != "" {
			return nil, errors.New("GITHUB_CLIENT_ID is required when GitHub authentication is configured")
		}
		return nil, nil
	}
	callback, frontend := os.Getenv("GITHUB_CALLBACK_URL"), os.Getenv("GITHUB_FRONTEND_URL")
	if secret == "" || store == nil {
		return nil, errors.New("GitHub authentication requires a client secret and account storage")
	}
	for _, value := range []string{callback, frontend} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, errors.New("GitHub callback and frontend URLs must be absolute URLs without credentials, query, or fragment")
		}
		local := parsed.Scheme == "http" && (parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "localhost" || parsed.Hostname() == "::1") && os.Getenv("APP_ENV") != "production"
		if parsed.Scheme != "https" && !local {
			return nil, errors.New("GitHub URLs require HTTPS except local development loopback addresses")
		}
	}
	parsedCallback, _ := url.Parse(callback)
	parsedFrontend, _ := url.Parse(frontend)
	if !strings.HasSuffix(parsedCallback.Path, "/api/auth/github/callback") || !isAllowedOrigin(parsedFrontend.Scheme+"://"+parsedFrontend.Host) {
		return nil, errors.New("GitHub callback path or frontend origin is not allowed")
	}
	return &githubAccountService{store: store, clientID: id, secret: secret, callbackURL: callback, frontendURL: frontend,
		authorizeURL: "https://github.com/login/oauth/authorize", tokenURL: "https://github.com/login/oauth/access_token", userURL: "https://api.github.com/user",
		client: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}, nil
}

func accountKey(kind, token string) string {
	hash := sha256.Sum256([]byte(token))
	return kind + ":" + hex.EncodeToString(hash[:])
}

func accountJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func accountDecode(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if decoder.Decode(value) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeJSONError(w, http.StatusBadRequest, "invalid_account_request")
		return false
	}
	return true
}

func registerAccountRoutes(mux *http.ServeMux, accounts *githubAccountService) {
	mux.Handle("/api/account/", accounts)
	mux.Handle("/api/auth/github/", accounts)
}

func (a *githubAccountService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.URL.Path == "/api/account/config" && r.Method == http.MethodGet {
		accountJSON(w, http.StatusOK, map[string]any{"github_enabled": a != nil, "repository_limit": savedRepositoryLimit})
		return
	}
	if a == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "github_signin_not_configured")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 24*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	switch {
	case r.URL.Path == "/api/auth/github/start" && r.Method == http.MethodGet:
		a.start(w, r)
	case r.URL.Path == "/api/auth/github/callback" && r.Method == http.MethodGet:
		a.callback(w, r)
	case r.URL.Path == "/api/account/session" && r.Method == http.MethodPost:
		if !requireTrustedOrigin(w, r) {
			return
		}
		var body struct {
			Ticket   string `json:"ticket"`
			Verifier string `json:"verifier"`
		}
		if !accountDecode(w, r, &body) {
			return
		}
		if !accountTokenPattern.MatchString(body.Ticket) || !accountTokenPattern.MatchString(body.Verifier) {
			writeJSONError(w, http.StatusUnauthorized, "invalid_signin_ticket")
			return
		}
		record, err := a.store.GetTransient(ctx, accountKey("ticket", body.Ticket))
		if err != nil || record.ExpiresAt <= time.Now().Unix() {
			a.recordError(w, err, "invalid_signin_ticket")
			return
		}
		if !constantTimeSecretEqual(record.BrowserChallenge, browserHandoffChallenge(body.Verifier)) {
			writeJSONError(w, http.StatusUnauthorized, "invalid_signin_ticket")
			return
		}
		record, err = a.store.TakeTransient(ctx, accountKey("ticket", body.Ticket))
		if err != nil {
			a.recordError(w, err, "invalid_signin_ticket")
			return
		}
		token, err := randomSessionToken()
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "account_unavailable")
			return
		}
		record.ExpiresAt = time.Now().Add(githubSessionLifetime).Unix()
		record.BrowserChallenge = ""
		if err := a.store.PutTransient(ctx, accountKey("session", token), record); err != nil {
			a.recordError(w, err, "invalid_signin_ticket")
			return
		}
		a.sessionResponse(w, token, record)
	default:
		a.accountRequest(w, r)
	}
}

func (a *githubAccountService) recordError(w http.ResponseWriter, err error, missing string) {
	if err == nil || errors.Is(err, errAccountRecordMissing) {
		writeJSONError(w, http.StatusUnauthorized, missing)
	} else {
		writeJSONError(w, http.StatusServiceUnavailable, "account_unavailable")
	}
}

func (a *githubAccountService) stateCookie(value string, maxAge int) *http.Cookie {
	callback, _ := url.Parse(a.callbackURL)
	return &http.Cookie{Name: "ucs503_github_state", Value: value, Path: strings.TrimSuffix(callback.Path, "/callback"), MaxAge: maxAge, HttpOnly: true, Secure: callback.Scheme == "https", SameSite: http.SameSiteLaxMode}
}

func (a *githubAccountService) start(w http.ResponseWriter, r *http.Request) {
	proof := r.URL.Query().Get("handoff_challenge")
	if !browserChallengePattern.MatchString(proof) {
		writeJSONError(w, http.StatusBadRequest, "invalid_signin_request")
		return
	}
	state, err := randomSessionToken()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "account_unavailable")
		return
	}
	verifier, err := randomSessionToken()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "account_unavailable")
		return
	}
	if err := a.store.PutTransient(r.Context(), accountKey("state", state), accountTransient{Verifier: verifier, BrowserChallenge: proof, ExpiresAt: time.Now().Add(10 * time.Minute).Unix()}); err != nil {
		a.recordError(w, err, "invalid_oauth_state")
		return
	}
	challenge := sha256.Sum256([]byte(verifier))
	destination, _ := url.Parse(a.authorizeURL)
	query := url.Values{"client_id": {a.clientID}, "redirect_uri": {a.callbackURL}, "state": {state}, "scope": {""}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}}
	destination.RawQuery = query.Encode()
	http.SetCookie(w, a.stateCookie(state, 600))
	http.Redirect(w, r, destination.String(), http.StatusFound)
}

func (a *githubAccountService) callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("ucs503_github_state")
	if err != nil || !accountTokenPattern.MatchString(state) || !constantTimeSecretEqual(cookie.Value, state) {
		writeJSONError(w, http.StatusBadRequest, "invalid_oauth_state")
		return
	}
	http.SetCookie(w, a.stateCookie("", -1))
	record, err := a.store.TakeTransient(r.Context(), accountKey("state", state))
	if err != nil || record.ExpiresAt <= time.Now().Unix() {
		a.recordError(w, err, "invalid_oauth_state")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" || len(code) > 512 || r.URL.Query().Get("error") != "" {
		http.Redirect(w, r, a.frontendURL+"#auth_error=github_signin_cancelled", http.StatusFound)
		return
	}
	user, err := a.exchangeGitHubCode(r.Context(), code, record.Verifier)
	if err != nil {
		http.Redirect(w, r, a.frontendURL+"#auth_error=github_signin_failed", http.StatusFound)
		return
	}
	ticket, err := randomSessionToken()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "account_unavailable")
		return
	}
	if err := a.store.PutTransient(r.Context(), accountKey("ticket", ticket), accountTransient{User: user, BrowserChallenge: record.BrowserChallenge, ExpiresAt: time.Now().Add(time.Minute).Unix()}); err != nil {
		a.recordError(w, err, "invalid_signin_ticket")
		return
	}
	http.Redirect(w, r, a.frontendURL+"#auth_ticket="+ticket, http.StatusFound)
}

func browserHandoffChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func (a *githubAccountService) exchangeGitHubCode(ctx context.Context, code, verifier string) (GitHubUser, error) {
	body := url.Values{"client_id": {a.clientID}, "client_secret": {a.secret}, "code": {code}, "redirect_uri": {a.callbackURL}, "code_verifier": {verifier}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return GitHubUser{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	response, err := a.client.Do(request)
	if err != nil {
		return GitHubUser{}, err
	}
	defer response.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if response.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&token) != nil || token.AccessToken == "" || !strings.EqualFold(token.TokenType, "bearer") {
		return GitHubUser{}, errors.New("GitHub token exchange failed")
	}
	request, err = http.NewRequestWithContext(ctx, http.MethodGet, a.userURL, nil)
	if err != nil {
		return GitHubUser{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err = a.client.Do(request)
	if err != nil {
		return GitHubUser{}, err
	}
	defer response.Body.Close()
	var profile struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}
	if response.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&profile) != nil || profile.ID <= 0 || !githubLoginPattern.MatchString(profile.Login) {
		return GitHubUser{}, errors.New("GitHub identity lookup failed")
	}
	// The GitHub access token is discarded after identity verification.
	return GitHubUser{ID: strconv.FormatInt(profile.ID, 10), Login: profile.Login}, nil
}

func (a *githubAccountService) sessionResponse(w http.ResponseWriter, token string, record accountTransient) {
	accountJSON(w, http.StatusOK, map[string]any{"token": token, "user": record.User, "expires_at": time.Unix(record.ExpiresAt, 0).UTC(), "expires_in_ms": time.Until(time.Unix(record.ExpiresAt, 0)).Milliseconds()})
}

func (a *githubAccountService) accountRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	known := (path == "/api/account/me" && r.Method == http.MethodGet) || (path == "/api/account/logout" && r.Method == http.MethodPost) || (path == "/api/account/repositories" && (r.Method == http.MethodGet || r.Method == http.MethodPost)) || (strings.HasPrefix(path, "/api/account/repositories/") && r.Method == http.MethodDelete)
	if !known {
		writeJSONError(w, http.StatusNotFound, "account_route_not_found")
		return
	}
	if r.Method != http.MethodGet && !requireTrustedOrigin(w, r) {
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") || !accountTokenPattern.MatchString(token) {
		writeJSONError(w, http.StatusUnauthorized, "user_signin_required")
		return
	}
	record, err := a.store.GetTransient(r.Context(), accountKey("session", token))
	if err != nil || record.ExpiresAt <= time.Now().Unix() {
		a.recordError(w, err, "user_signin_required")
		return
	}
	switch {
	case path == "/api/account/me":
		a.sessionResponse(w, "", record)
	case path == "/api/account/logout":
		_, err := a.store.TakeTransient(r.Context(), accountKey("session", token))
		if err != nil {
			a.recordError(w, err, "user_signin_required")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case strings.HasPrefix(path, "/api/account/repositories/"):
		id := strings.TrimPrefix(path, "/api/account/repositories/")
		if !accountTokenPattern.MatchString(id) {
			writeJSONError(w, http.StatusBadRequest, "invalid_repository")
			return
		}
		if err := a.store.DeleteRepository(r.Context(), record.User.ID, id); err != nil {
			a.recordError(w, err, "user_signin_required")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPost:
		var body struct {
			URL string `json:"url"`
		}
		if !accountDecode(w, r, &body) {
			return
		}
		repository, err := parseSavedRepository(body.URL)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid_repository")
			return
		}
		if err := a.store.SaveRepository(r.Context(), record.User.ID, repository); err != nil {
			if errors.Is(err, errSavedRepositoryLimit) {
				writeJSONError(w, http.StatusConflict, "saved_repository_limit")
				return
			}
			a.recordError(w, err, "user_signin_required")
			return
		}
		accountJSON(w, http.StatusOK, repository)
	default:
		repositories, err := a.store.ListRepositories(r.Context(), record.User.ID)
		if err != nil {
			a.recordError(w, err, "user_signin_required")
			return
		}
		accountJSON(w, http.StatusOK, repositories)
	}
}

func parseSavedRepository(raw string) (SavedRepository, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || len(raw) > 300 || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.EscapedPath() != parsed.Path {
		return SavedRepository{}, errors.New("invalid GitHub repository")
	}
	path := strings.TrimSuffix(strings.Trim(parsed.Path, "/"), ".git")
	if !savedPathPattern.MatchString(path) {
		return SavedRepository{}, errors.New("invalid GitHub repository path")
	}
	for _, part := range strings.Split(path, "/") {
		if part == "." || part == ".." {
			return SavedRepository{}, errors.New("invalid GitHub repository path")
		}
	}
	hash := sha256.Sum256([]byte(strings.ToLower(path)))
	return SavedRepository{ID: hex.EncodeToString(hash[:]), URL: "https://github.com/" + path, Name: path, SavedAt: time.Now().UTC()}, nil
}
