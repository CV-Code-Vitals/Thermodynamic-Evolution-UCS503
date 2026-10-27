package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestValidateGitHubRepositoryURL_AcceptsPublicRepository(t *testing.T) {
	for _, testCase := range []struct {
		url  string
		repo string
	}{
		{url: "https://github.com/octocat/Spoon-Knife", repo: "Spoon-Knife"},
		{url: "https://github.com/octocat/Spoon-Knife.git", repo: "Spoon-Knife"},
	} {
		owner, repo, err := ValidateGitHubRepositoryURL(testCase.url)
		if err != nil {
			t.Fatalf("ValidateGitHubRepositoryURL(%q) returned error: %v", testCase.url, err)
		}
		if owner != "octocat" || repo != testCase.repo {
			t.Fatalf("ValidateGitHubRepositoryURL(%q) = (%q, %q), want (%q, %q)", testCase.url, owner, repo, "octocat", testCase.repo)
		}
	}
}

func TestValidateGitHubRepositoryURL_RejectsUntrustedURL(t *testing.T) {
	for _, rawURL := range []string{
		"http://github.com/octocat/repo",
		"https://user:pass@github.com/octocat/repo",
		"https://github.com.evil.example/octocat/repo",
		"https://127.0.0.1/octocat/repo",
		"https://github.com/octocat/repo/extra",
		"https://github.com/octocat/repo?redirect=https://evil.example",
		"https://github.com/octocat/repo#fragment",
		"https://github.com:8443/octocat/repo",
	} {
		t.Run(rawURL, func(t *testing.T) {
			if owner, repo, err := ValidateGitHubRepositoryURL(rawURL); err == nil {
				t.Fatalf("expected URL to be rejected, got owner=%q repo=%q", owner, repo)
			}
		})
	}
}

func TestStageRepository_RejectsUnsafeRedirect(t *testing.T) {
	roundTripper := &sequenceRoundTripper{responses: []*http.Response{
		jsonResponse(`{"default_branch":"main"}`),
		{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{"https://attacker.example/archive.zip"}},
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}}
	store := &fakeSnapshotStore{}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    store,
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("140.82.112.5")}}},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil {
		t.Fatal("expected a redirect outside the GitHub archive host allowlist to fail")
	}
	if roundTripper.calls != 2 {
		t.Fatalf("unsafe redirect target was requested; got %d HTTP calls, want 2", roundTripper.calls)
	}
	if len(store.keys) != 0 {
		t.Fatalf("unsafe redirect created snapshots: %#v", store.keys)
	}
}

func TestStageRepository_RejectsRedirectToNonHTTPSPort(t *testing.T) {
	roundTripper := &sequenceRoundTripper{responses: []*http.Response{
		jsonResponse(`{"default_branch":"main"}`),
		{
			StatusCode: http.StatusFound,
			Header: http.Header{
				"Location": []string{"https://api.github.com:8443/repos/octocat/repo/zipball/main"},
			},
			Body: io.NopCloser(strings.NewReader("")),
		},
		binaryResponse(http.StatusOK, makeZIP(t, []zipEntry{{name: "repo/main.go", content: "package main"}})),
	}}
	store := &fakeSnapshotStore{}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    store,
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("140.82.112.5")}}},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil {
		t.Fatal("expected redirect to a nonstandard HTTPS port to be rejected")
	}
	if roundTripper.calls != 2 {
		t.Fatalf("nonstandard redirect target was requested; got %d HTTP calls, want 2", roundTripper.calls)
	}
	if len(store.keys) != 0 {
		t.Fatalf("nonstandard redirect created snapshots: %#v", store.keys)
	}
}

func TestStageRepository_RejectsPrivateAddress(t *testing.T) {
	roundTripper := &sequenceRoundTripper{}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    &fakeSnapshotStore{},
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil {
		t.Fatal("expected private-address DNS resolution to fail")
	}
	if roundTripper.calls != 0 {
		t.Fatalf("private-address host was requested before rejection: %d calls", roundTripper.calls)
	}
}

func TestStageRepository_EnforcesArchiveLimits(t *testing.T) {
	archive := makeZIP(t, []zipEntry{{name: "octocat-repo/main.go", content: "123456789"}})
	roundTripper := &sequenceRoundTripper{responses: []*http.Response{
		jsonResponse(`{"default_branch":"main"}`),
		binaryResponse(http.StatusOK, archive),
	}}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    &fakeSnapshotStore{},
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("140.82.112.5")}}},
		Limits: SourceLimits{
			MaxArchiveBytes:  8,
			MaxExpandedBytes: 32,
			MaxEntries:       10,
			MaxFileBytes:     16,
		},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil {
		t.Fatal("expected oversized compressed archive to fail before storage")
	}
}

func TestValidateArchiveAllowsFileLargerThanLegacyLimit(t *testing.T) {
	archive := makeZIP(t, []zipEntry{{
		name:    "repo/data.npz",
		content: strings.Repeat("x", 17<<20),
	}})
	stager := NewRepositoryStager(RepositoryStagerConfig{Limits: defaultSourceLimits()})

	if err := stager.validateArchive(archive); err != nil {
		t.Fatalf("validateArchive() rejected a file within the 32 MiB per-file limit: %v", err)
	}
}

func TestStageRepository_RejectsUnsafeArchiveEntries(t *testing.T) {
	archive := makeZIP(t, []zipEntry{{name: "../escape.txt", content: "blocked"}})
	roundTripper := &sequenceRoundTripper{responses: []*http.Response{
		jsonResponse(`{"default_branch":"main"}`),
		binaryResponse(http.StatusOK, archive),
	}}
	store := &fakeSnapshotStore{}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    store,
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("140.82.112.5")}}},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil {
		t.Fatal("expected archive traversal entry to be rejected")
	}
	if len(store.keys) != 0 {
		t.Fatalf("unsafe archive was stored: %#v", store.keys)
	}
}

func TestStageRepository_PropagatesS3Failure(t *testing.T) {
	archive := makeZIP(t, []zipEntry{{name: "octocat-repo/main.go", content: "package main"}})
	roundTripper := &sequenceRoundTripper{responses: []*http.Response{
		jsonResponse(`{"default_branch":"main"}`),
		binaryResponse(http.StatusOK, archive),
	}}
	store := &fakeSnapshotStore{err: errors.New("S3 unavailable")}
	stager := NewRepositoryStager(RepositoryStagerConfig{
		Client:   &http.Client{Transport: roundTripper},
		Store:    store,
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("140.82.112.5")}}},
	})

	if _, err := stager.StageRepository(context.Background(), "https://github.com/octocat/repo"); err == nil ||
		!strings.Contains(err.Error(), "S3 unavailable") {
		t.Fatalf("expected S3 failure to be surfaced, got %v", err)
	}
	if len(store.keys) != 1 {
		t.Fatalf("expected one attempted snapshot write, got %#v", store.keys)
	}
	if !strings.HasPrefix(store.keys[0], "snapshots/") || !strings.HasSuffix(store.keys[0], "/source.zip") {
		t.Fatalf("unexpected generated snapshot key: %q", store.keys[0])
	}
}

type zipEntry struct {
	name    string
	content string
}

func makeZIP(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		file, err := writer.Create(entry.name)
		if err != nil {
			t.Fatalf("create ZIP entry: %v", err)
		}
		if _, err := file.Write([]byte(entry.content)); err != nil {
			t.Fatalf("write ZIP entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP writer: %v", err)
	}
	return buffer.Bytes()
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func binaryResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/zip"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

type sequenceRoundTripper struct {
	responses []*http.Response
	calls     int
}

func (r *sequenceRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	if r.calls >= len(r.responses) {
		return nil, errors.New("unexpected HTTP request")
	}
	response := r.responses[r.calls]
	r.calls++
	return response, nil
}

type staticResolver struct {
	ips []net.IPAddr
	err error
}

func (r staticResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return r.ips, r.err
}

type fakeSnapshotStore struct {
	keys  []string
	err   error
	bytes [][]byte
}

func (s *fakeSnapshotStore) PutSnapshot(_ context.Context, key string, body []byte) error {
	s.keys = append(s.keys, key)
	s.bytes = append(s.bytes, append([]byte(nil), body...))
	return s.err
}
