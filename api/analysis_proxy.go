package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

func newAnalysisProxy(address, token string) (http.Handler, error) {
	if address == "" && token == "" {
		return nil, nil
	}
	target, err := url.Parse(address)
	if err != nil || target == nil || target.Host == "" ||
		(target.Scheme != "http" && target.Scheme != "https") || target.User != nil ||
		(target.Path != "" && target.Path != "/") || target.RawQuery != "" || target.Fragment != "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("analysis requires ANALYSIS_SERVICE_URL as an HTTP(S) origin and ANALYSIS_API_TOKEN")
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(target)
			request.Out.URL.Path = "/api/" + strings.TrimPrefix(request.In.URL.Path, "/api/analysis/")
			request.Out.Header.Set("X-API-Token", token)
			request.Out.Header.Del("Authorization")
			request.Out.Header.Del("Cookie")
			request.Out.Header.Del("Origin")
		},
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			writeJSONError(w, http.StatusBadGateway, "analysis_unavailable")
		},
	}
	return proxy, nil
}

func registerAnalysisRoutes(mux *http.ServeMux, deps Dependencies) {
	publicScans := make(chan struct{}, 1)
	repositoryPath := regexp.MustCompile(`^/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/?$`)
	mux.HandleFunc("/api/analysis/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/analysis/")
		publicScan := path == "scan" && r.Method == http.MethodPost
		authorized := routeAuthorized(r, deps.Auth)
		if !publicScan && !authorized {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !((path == "health" && r.Method == http.MethodGet) ||
			((path == "scan" || path == "evolve") && r.Method == http.MethodPost)) {
			writeJSONError(w, http.StatusNotFound, "analysis_route_not_found")
			return
		}
		if r.Method == http.MethodPost && !requireTrustedOrigin(w, r) {
			return
		}
		if deps.Analysis == nil {
			writeJSONError(w, http.StatusServiceUnavailable, "analysis_not_configured")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "analysis_request_too_large")
				return
			}
			var payload map[string]json.RawMessage
			if json.Unmarshal(body, &payload) != nil || payload == nil {
				writeJSONError(w, http.StatusBadRequest, "invalid_analysis_request")
				return
			}
			if path == "scan" {
				var repository string
				if json.Unmarshal(payload["repo_url"], &repository) != nil {
					writeJSONError(w, http.StatusBadRequest, "invalid_repository")
					return
				}
				parsed, err := url.Parse(repository)
				if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") ||
					parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
					!repositoryPath.MatchString(parsed.Path) || parsed.EscapedPath() != parsed.Path {
					writeJSONError(w, http.StatusBadRequest, "invalid_repository")
					return
				}
				for _, part := range strings.Split(strings.Trim(parsed.Path, "/"), "/") {
					if part == "." || part == ".." {
						writeJSONError(w, http.StatusBadRequest, "invalid_repository")
						return
					}
				}
				retain, _ := json.Marshal(authorized)
				payload = map[string]json.RawMessage{"repo_url": payload["repo_url"], "retain_workspace": retain}
				if !authorized {
					select {
					case publicScans <- struct{}{}:
						defer func() { <-publicScans }()
					default:
						writeJSONError(w, http.StatusTooManyRequests, "scan_capacity_reached")
						return
					}
				}
			} else {
				var workspace string
				_, hasServerPath := payload["repo_path"]
				if json.Unmarshal(payload["workspace_id"], &workspace) != nil || strings.TrimSpace(workspace) == "" || hasServerPath {
					writeJSONError(w, http.StatusBadRequest, "scan_workspace_required")
					return
				}
			}
			body, _ = json.Marshal(payload)
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
		}
		deps.Analysis.ServeHTTP(w, r)
	})
}
