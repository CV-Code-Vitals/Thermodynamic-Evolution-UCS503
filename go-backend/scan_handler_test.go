package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestScanRepoHandler_UsesStagedSnapshot(t *testing.T) {
	workspace := t.TempDir()
	provider := &fakeSnapshotWorkspaceProvider{path: workspace}
	var engineWorkspace string
	handler := scanRepoHandlerWithRunner(
		"engine",
		time.Second,
		provider,
		NewPublicScanController(time.Minute, 1),
		func(_ context.Context, _, codeDir, reportPath string) (string, error) {
			engineWorkspace = codeDir
			if err := os.WriteFile(reportPath, []byte("{}"), 0o600); err != nil {
				return "", err
			}
			return "", nil
		},
	)
	response := serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, "8.8.8.8", "")

	if response.Code != http.StatusOK {
		t.Fatalf("scan status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if provider.calls != 1 || provider.snapshot != testSnapshotID {
		t.Fatalf("snapshot provider calls = %d, ID = %q", provider.calls, provider.snapshot)
	}
	if engineWorkspace != workspace {
		t.Fatalf("engine workspace = %q, want extracted snapshot path %q", engineWorkspace, workspace)
	}
	if !provider.cleanedUp {
		t.Fatal("snapshot workspace was not cleaned after scan")
	}
}

func TestPublicScanController_RateLimitsClient(t *testing.T) {
	provider := &fakeSnapshotWorkspaceProvider{path: t.TempDir()}
	engineCalls := 0
	handler := scanRepoHandlerWithRunner(
		"engine",
		time.Second,
		provider,
		NewPublicScanController(time.Minute, 1),
		func(_ context.Context, _, _, reportPath string) (string, error) {
			engineCalls++
			return "", os.WriteFile(reportPath, []byte("{}"), 0o600)
		},
	)

	first := serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, "8.8.8.8", "")
	second := serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, "8.8.8.8", "")

	if first.Code != http.StatusOK {
		t.Fatalf("first scan status = %d, want %d", first.Code, http.StatusOK)
	}
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second scan status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
	if engineCalls != 1 {
		t.Fatalf("engine calls = %d, want 1", engineCalls)
	}
}

func TestPublicScanController_RejectsConcurrentScan(t *testing.T) {
	provider := &fakeSnapshotWorkspaceProvider{path: t.TempDir()}
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	handler := scanRepoHandlerWithRunner(
		"engine",
		time.Second,
		provider,
		NewPublicScanController(time.Minute, 1),
		func(_ context.Context, _, _, reportPath string) (string, error) {
			select {
			case entered <- struct{}{}:
			default:
			}
			<-release
			return "", os.WriteFile(reportPath, []byte("{}"), 0o600)
		},
	)
	firstDone := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		firstDone <- serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, "8.8.8.8", "")
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first scan did not enter the engine")
	}
	second := serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, "9.9.9.9", "")
	close(release)
	first := <-firstDone

	if first.Code != http.StatusOK {
		t.Fatalf("first scan status = %d, want %d", first.Code, http.StatusOK)
	}
	if second.Code != http.StatusServiceUnavailable {
		t.Fatalf("concurrent scan status = %d, want %d", second.Code, http.StatusServiceUnavailable)
	}
}

func TestScanRepoHandler_IgnoresClientForwardedIP(t *testing.T) {
	provider := &fakeSnapshotWorkspaceProvider{path: t.TempDir()}
	handler := scanRepoHandlerWithRunner(
		"engine",
		time.Second,
		provider,
		NewPublicScanController(time.Minute, 1),
		func(_ context.Context, _, _, reportPath string) (string, error) {
			return "", os.WriteFile(reportPath, []byte("{}"), 0o600)
		},
	)
	for _, trustedIP := range []string{"8.8.8.8", "9.9.9.9"} {
		response := serveScanRequest(handler, `{"snapshot_id":"`+testSnapshotID+`"}`, trustedIP, "1.1.1.1")
		if response.Code != http.StatusOK {
			t.Fatalf("scan for trusted client %s returned %d; X-Forwarded-For must not control admission", trustedIP, response.Code)
		}
	}
}

func TestScanRepoHandler_CleansTemporaryWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		cancelRequest bool
		runner        func(context.Context, string, string, string) (string, error)
	}{
		{
			name: "success",
			runner: func(_ context.Context, _, _, reportPath string) (string, error) {
				return "", os.WriteFile(reportPath, []byte("{}"), 0o600)
			},
		},
		{
			name: "failure",
			runner: func(context.Context, string, string, string) (string, error) {
				return "", errors.New("engine failed")
			},
		},
		{
			name:    "timeout",
			timeout: 10 * time.Millisecond,
			runner: func(ctx context.Context, _, _, _ string) (string, error) {
				<-ctx.Done()
				return "", ctx.Err()
			},
		},
		{
			name:          "cancellation",
			cancelRequest: true,
			runner: func(ctx context.Context, _, _, _ string) (string, error) {
				<-ctx.Done()
				return "", ctx.Err()
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			workspace := t.TempDir()
			provider := &fakeSnapshotWorkspaceProvider{path: workspace}
			timeout := testCase.timeout
			if timeout == 0 {
				timeout = time.Second
			}
			handler := scanRepoHandlerWithRunner(
				"engine",
				timeout,
				provider,
				NewPublicScanController(time.Minute, 1),
				testCase.runner,
			)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request := httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(`{"snapshot_id":"`+testSnapshotID+`"}`)).WithContext(ctx)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Trusted-Client-IP", "8.8.8.8")
			response := httptest.NewRecorder()
			done := make(chan struct{})
			go func() {
				newScanRouter(handler).ServeHTTP(response, request)
				close(done)
			}()

			if testCase.cancelRequest {
				time.Sleep(10 * time.Millisecond)
				cancel()
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("scan handler did not finish")
			}
			if !provider.cleanedUp {
				t.Fatal("snapshot workspace was not cleaned")
			}
		})
	}
}

func serveScanRequest(handler gin.HandlerFunc, body, trustedIP, forwardedFor string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/scan", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Trusted-Client-IP", trustedIP)
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	response := httptest.NewRecorder()
	newScanRouter(handler).ServeHTTP(response, request)
	return response
}

func newScanRouter(handler gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.POST("/api/scan", handler)
	return router
}
