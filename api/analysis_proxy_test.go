package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/evolve" || r.Header.Get("X-API-Token") != "private-service-token" {
			t.Errorf("unexpected upstream path or authorization")
		}
		if r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" || r.Header.Get("Authorization") != "" {
			t.Error("browser credentials leaked to upstream")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"workspace_id":"example"}` {
			t.Errorf("unexpected request body: %s", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "event: complete\ndata: {\"final_energy\":1}\n\n")
	}))
	defer upstream.Close()
	proxy, err := newAnalysisProxy(upstream.URL, "private-service-token")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := NewHMACTokenAuthenticator("local-passkey", strings.Repeat("k", 32))
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := auth.Issue("local-passkey")
	if err != nil {
		t.Fatal(err)
	}
	router := newRouter(Dependencies{Auth: auth, Analysis: proxy})
	request := httptest.NewRequest(http.MethodPost, "/api/analysis/evolve", strings.NewReader(`{"workspace_id":"example"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Cookie", "browser=secret")
	request.Header.Set("X-API-Token", "untrusted")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "event: complete\ndata: {\"final_energy\":1}\n\n" {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body)
	}
	for _, tc := range []struct {
		path, method, origin, bearer string
		status                       int
	}{
		{"/api/analysis/health", "GET", "", "", 401},
		{"/api/analysis/evolve", "POST", "https://untrusted.example", token, 403},
		{"/api/analysis/anything", "POST", "http://localhost:5173", token, 404},
	} {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		r.Header.Set("Authorization", "Bearer "+tc.bearer)
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Errorf("%s: got %d, want %d", tc.path, w.Code, tc.status)
		}
	}
}

func TestAnalysisProxyFlushesBeforeUpstreamCompletion(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "event: step\ndata: {\"step\":1}\n\n")
		w.(http.Flusher).Flush()
		<-release
	}))
	proxy, err := newAnalysisProxy(upstream.URL, "private-token")
	if err != nil {
		t.Fatal(err)
	}
	gateway := httptest.NewServer(newRouter(Dependencies{Analysis: proxy, Auth: &testAuthenticator{validToken: "portal-token"}}))
	defer func() { close(release); gateway.Close(); upstream.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, gateway.URL+"/api/analysis/evolve", strings.NewReader(`{"workspace_id":"example"}`))
	req.Header.Set("Authorization", "Bearer portal-token")
	req.Header.Set("Origin", "https://portal.example")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stream headers were buffered: %v", err)
	}
	defer response.Body.Close()
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil || line != "event: step\n" {
		t.Fatalf("first event not delivered while upstream was running: %q %v", line, err)
	}
}

func TestPortalAnalysisRestrictsEvolutionToScannedWorkspaces(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://portal.example")
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			RetainWorkspace bool `json:"retain_workspace"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || !payload.RetainWorkspace {
			t.Error("public scan was forwarded without retaining a restricted workspace")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	router := newRouter(Dependencies{Analysis: upstream, Auth: &testAuthenticator{validToken: "portal-token"}})
	for _, tc := range []struct {
		path, body string
		want       int
	}{
		{"scan", `{"repo_url":"https://github.com/example/project","retain_workspace":false}`, 204},
		{"evolve", `{"repo_path":"/server/source"}`, 400},
		{"evolve", `{"workspace_id":"valid","repo_path":"/server/source"}`, 400},
		{"evolve", `{"workspace_id":""}`, 400},
		{"scan", `{broken}`, 400},
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/analysis/"+tc.path, strings.NewReader(tc.body))
		req.Header.Set("Authorization", "Bearer portal-token")
		req.Header.Set("Origin", "https://portal.example")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.body, response.Code, tc.want)
		}
	}
}

func TestAnalysisProxyConfiguration(t *testing.T) {
	for _, address := range []string{"", "/api", "file:///tmp", "https://user:pass@example.com", "https://example.com/api"} {
		if _, err := newAnalysisProxy(address, "token"); err == nil {
			t.Errorf("accepted invalid origin %q", address)
		}
	}
	if proxy, err := newAnalysisProxy("", ""); proxy != nil || err != nil {
		t.Fatal("optional analysis should remain disabled")
	}
}
