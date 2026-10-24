package agent

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const testBrokerWorkspaceID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestSandboxRunnerIsolationAndCIPinsImage(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve broker test source path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
	readRepoFile := func(relativePath string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		return string(content)
	}

	dockerfile := readRepoFile(filepath.Join("docker", "sandbox-runner", "Dockerfile"))
	if !strings.Contains(dockerfile, "USER 10001:10001") {
		t.Error("sandbox runner Dockerfile must use a non-root user")
	}

	evaluatorSource := readRepoFile(filepath.Join("go-backend", "agent", "evaluator.go"))
	for _, required := range []string{
		`"--network", "none"`,
		`"--read-only"`,
		`"--cap-drop", "ALL"`,
		`"--security-opt", "no-new-privileges"`,
		`"--memory"`,
		`"--cpus"`,
		`"--pids-limit"`,
		"outBuf.maxBytes = limits.MaxOutputBytes",
		`limitedBuffer`,
	} {
		if !strings.Contains(evaluatorSource, required) {
			t.Errorf("sandbox evaluator is missing required isolation control %q", required)
		}
	}

	ciWorkflow := readRepoFile(filepath.Join(".github", "workflows", "ci-go.yml"))
	if !strings.Contains(ciWorkflow, "SANDBOX_RUNNER_IMAGE: ucs503-sandbox-runner:${{ github.sha }}") {
		t.Error("mandatory Docker integration tests must use a commit-SHA-tagged runner image")
	}
	if !strings.Contains(ciWorkflow, `docker build -t "$SANDBOX_RUNNER_IMAGE"`) {
		t.Error("CI must build the same pinned image used by the integration tests")
	}
	if strings.Contains(ciWorkflow, "ucs503-sandbox-runner:latest") {
		t.Error("CI must not use a mutable latest sandbox runner image")
	}
	if !strings.Contains(ciWorkflow, "TestDockerEvaluator_LiveExecution") ||
		!strings.Contains(ciWorkflow, "--- PASS: TestDockerEvaluator_LiveExecution") ||
		!strings.Contains(ciWorkflow, "--- SKIP: TestDockerEvaluator_LiveExecution") {
		t.Error("CI must preserve the mandatory live Docker test and its pass/skip checks")
	}

	publishWorkflow := readRepoFile(filepath.Join(".github", "workflows", "publish-api-image.yml"))
	if !strings.Contains(publishWorkflow, "ghcr.io/cv-code-vitals/ucs503-analysis:${{ github.sha }}") ||
		!strings.Contains(publishWorkflow, "ghcr.io/cv-code-vitals/ucs503-sandbox-runner:${{ github.sha }}") {
		t.Error("image publishing must tag analysis and runner images with the commit SHA")
	}
	if strings.Contains(publishWorkflow, ":latest") {
		t.Error("published deployment images must not use the mutable latest tag")
	}
	if strings.Contains(publishWorkflow, "ucs503-api") {
		t.Error("image publishing must not publish the obsolete standalone archive API")
	}
}

// TestRemoteBrokerEvaluator_MissingOrInvalidConfig verifies that empty or invalid configuration
// causes safe, immediate failures.
func TestRemoteBrokerEvaluator_MissingOrInvalidConfig(t *testing.T) {
	evaluator := NewRemoteBrokerEvaluator("", "")

	ctx := context.Background()
	err := evaluator.CheckAvailable(ctx)
	if err == nil || !errors.Is(err, ErrSandboxUnavailable) {
		t.Fatalf("expected ErrSandboxUnavailable for empty broker URL, got: %v", err)
	}

	_, err = evaluator.EvaluateCandidate(ctx, EvaluationRequest{
		RepoPath: "",
		FilePath: "",
	})
	if err == nil || !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected ErrInvalidRequest for empty paths, got: %v", err)
	}
}

// TestRemoteBrokerEvaluator_BrokerUnavailable_FailsClosed verifies that when the sandbox broker
// is down, unreachable, or returns a 5xx error, candidate evaluation fails closed.
func TestRemoteBrokerEvaluator_BrokerUnavailable_FailsClosed(t *testing.T) {
	// Point to an unreachable port
	evaluator := NewRemoteBrokerEvaluator("https://127.0.0.1:54321", "test-token-123456")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := evaluator.EvaluateCandidate(ctx, EvaluationRequest{
		WorkspaceID: testBrokerWorkspaceID,
		FilePath:    "main.go",
		MutatedCode: "package main",
	})

	if err == nil {
		t.Fatal("expected error when broker is unreachable, got nil")
	}
	if !errors.Is(err, ErrSandboxUnavailable) {
		t.Fatalf("expected ErrSandboxUnavailable, got: %v", err)
	}
	if res.TestsPassed {
		t.Fatal("expected TestsPassed to be false when broker is unavailable")
	}
	if res.Outcome != OutcomeInfrastructureFailure {
		t.Fatalf("expected OutcomeInfrastructureFailure, got %s", res.Outcome)
	}
}

// TestRemoteBrokerEvaluator_AuthenticationFailure verifies that mismatched or missing tokens
// between the analysis service and sandbox broker cause explicit security rejections.
func TestRemoteBrokerEvaluator_AuthenticationFailure(t *testing.T) {
	expectedToken := "correct-secret-token-32byteslong!"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Sandbox-Token")
		if token != expectedToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// 1. Wrong token
	badEvaluator := newTestRemoteBrokerEvaluator(t, server, "wrong-token-value-here")
	ctx := context.Background()

	err := badEvaluator.CheckAvailable(ctx)
	if err == nil || !errors.Is(err, ErrBrokerAuthenticationFailed) {
		t.Fatalf("expected ErrBrokerAuthenticationFailed, got: %v", err)
	}

	res, err := badEvaluator.EvaluateCandidate(ctx, EvaluationRequest{
		WorkspaceID: testBrokerWorkspaceID,
		FilePath:    "calc.go",
		MutatedCode: "package calc",
	})
	if err == nil || !errors.Is(err, ErrBrokerAuthenticationFailed) {
		t.Fatalf("expected ErrBrokerAuthenticationFailed from EvaluateCandidate, got: %v", err)
	}
	if res.TestsPassed {
		t.Fatal("expected TestsPassed to be false on auth failure")
	}
	if res.Outcome != OutcomeSecurityRejected {
		t.Fatalf("expected OutcomeSecurityRejected, got %s", res.Outcome)
	}

	// 2. Correct token
	goodEvaluator := newTestRemoteBrokerEvaluator(t, server, expectedToken)
	if err := goodEvaluator.CheckAvailable(ctx); err != nil {
		t.Fatalf("expected CheckAvailable to succeed with valid token, got: %v", err)
	}
}

// TestRemoteBrokerEvaluator_TimeoutPropagation verifies that request deadlines and context
// cancellations are propagated and result in safe timeout classification.
func TestRemoteBrokerEvaluator_TimeoutPropagation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		// Simulate long-running execution
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tests_passed":true,"outcome":"passed"}`))
	}))
	defer server.Close()

	evaluator := newTestRemoteBrokerEvaluator(t, server, "valid-token-1234567890")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	res, err := evaluator.EvaluateCandidate(ctx, EvaluationRequest{
		WorkspaceID: testBrokerWorkspaceID,
		FilePath:    "main.go",
		MutatedCode: "package main",
	})

	if err != nil && !errors.Is(err, ErrExecutionTimeout) {
		t.Fatalf("expected ErrExecutionTimeout or nil with Timeout outcome, got: %v", err)
	}
	if res.Outcome != OutcomeTimeout {
		t.Fatalf("expected OutcomeTimeout, got: %s", res.Outcome)
	}
	if res.TestsPassed {
		t.Fatal("expected TestsPassed to be false on timeout")
	}
}

// TestRemoteBrokerEvaluator_SuccessfulEvaluation verifies end-to-end payload serialization,
// execution result reception, and byte preservation across the broker boundary.
func TestRemoteBrokerEvaluator_SuccessfulEvaluation(t *testing.T) {
	origContent := []byte("package calc\nfunc Add(a, b int) int { return 0 }\n")
	expectedOutput := "=== RUN TestAdd\n--- PASS: TestAdd (0.01s)\nPASS"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		if r.URL.Path == "/v1/evaluate" {
			var req EvaluationRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.FilePath != "calc.go" {
				http.Error(w, "unexpected file path", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID != testBrokerWorkspaceID || req.RepoPath != "" {
				http.Error(w, "unexpected workspace fields", http.StatusBadRequest)
				return
			}
			if string(req.OriginalBytes) != string(origContent) {
				http.Error(w, "original bytes mismatch", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tests_passed":       true,
				"outcome":            OutcomePassed,
				"output":             expectedOutput,
				"duration":           45 * time.Millisecond,
				"updated_file_bytes": []byte("updated source"),
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	evaluator := newTestRemoteBrokerEvaluator(t, server, "auth-token-1234567890")

	res, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{
		WorkspaceID:   testBrokerWorkspaceID,
		FilePath:      "calc.go",
		OriginalBytes: origContent,
		MutatedCode:   "return a + b",
		StartByte:     32,
		EndByte:       40,
	})

	if err != nil {
		t.Fatalf("unexpected error from EvaluateCandidate: %v", err)
	}
	if !res.TestsPassed {
		t.Fatal("expected TestsPassed to be true")
	}
	if res.Outcome != OutcomePassed {
		t.Fatalf("expected OutcomePassed, got: %s", res.Outcome)
	}
	if res.Output != expectedOutput {
		t.Fatalf("expected output %q, got %q", expectedOutput, res.Output)
	}
	if string(res.UpdatedFileBytes) != "updated source" {
		t.Fatalf("UpdatedFileBytes = %q, want updated source", res.UpdatedFileBytes)
	}
}

func TestRemoteBrokerEvaluator_DecodesEvaluationErrorWithoutFailingTransport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tests_passed":       false,
			"outcome":            OutcomeCompilationFailed,
			"updated_file_bytes": []byte("invalid source"),
			"error":              "candidate did not compile",
		})
	}))
	defer server.Close()

	evaluator := newTestRemoteBrokerEvaluator(t, server, "auth-token-1234567890")
	result, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{
		WorkspaceID: testBrokerWorkspaceID,
		FilePath:    "main.go",
		MutatedCode: "invalid source",
	})
	if err != nil {
		t.Fatalf("EvaluateCandidate() transport error = %v, want nil", err)
	}
	if result.Outcome != OutcomeCompilationFailed || result.TestsPassed {
		t.Fatalf("evaluation outcome = %q, passed = %t", result.Outcome, result.TestsPassed)
	}
	if result.Error == nil || result.Error.Error() != "candidate did not compile" {
		t.Fatalf("evaluation error = %v, want candidate did not compile", result.Error)
	}
	if string(result.UpdatedFileBytes) != "invalid source" {
		t.Fatalf("updated bytes = %q, want invalid source", result.UpdatedFileBytes)
	}
}

// TestNewEvaluator_Resolution verifies that NewEvaluator selects RemoteBrokerEvaluator
// when SANDBOX_SERVICE_URL is set, and DockerEvaluator when unset.
func TestNewEvaluator_Resolution(t *testing.T) {
	// 1. Unset: returns *DockerEvaluator
	t.Setenv("SANDBOX_SERVICE_URL", "")
	eval1 := NewEvaluator(DockerEvaluatorConfig{})
	if _, ok := eval1.(*DockerEvaluator); !ok {
		t.Fatalf("expected *DockerEvaluator when SANDBOX_SERVICE_URL is empty, got %T", eval1)
	}

	// 2. Set: returns *RemoteBrokerEvaluator
	t.Setenv("SANDBOX_SERVICE_URL", "https://sandbox-broker:8082")
	t.Setenv("SANDBOX_SERVICE_TOKEN", "super-secret-token")
	eval2 := NewEvaluator(DockerEvaluatorConfig{})
	remoteEval, ok := eval2.(*RemoteBrokerEvaluator)
	if !ok {
		t.Fatalf("expected *RemoteBrokerEvaluator when SANDBOX_SERVICE_URL is set, got %T", eval2)
	}
	if remoteEval.brokerURL != "https://sandbox-broker:8082" {
		t.Fatalf("expected brokerURL 'https://sandbox-broker:8082', got %q", remoteEval.brokerURL)
	}
	if remoteEval.token != "super-secret-token" {
		t.Fatalf("expected token 'super-secret-token', got %q", remoteEval.token)
	}
}

// TestSandboxWorkspace_CustomBaseDir verifies that SANDBOX_WORKSPACE_DIR is respected
// and temporary workspaces are cleaned up after evaluation.
func TestSandboxWorkspace_CustomBaseDir(t *testing.T) {
	customBase := filepath.Join(t.TempDir(), "custom-sandboxes")
	t.Setenv("SANDBOX_WORKSPACE_DIR", customBase)

	evaluator := NewDockerEvaluator(DockerEvaluatorConfig{
		RunnerChecker: func(ctx context.Context) error {
			return errors.New("simulated runtime unavailable")
		},
	})

	// When runtime is unavailable, workspace creation should never even occur or leak
	_, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{
		RepoPath:    t.TempDir(),
		FilePath:    "main.go",
		MutatedCode: "package main",
	})

	if err == nil {
		t.Fatal("expected simulated failure")
	}

	// Verify no stray workspace directory left behind
	entries, _ := os.ReadDir(customBase)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "ucs503-sandbox-ws-") {
			t.Errorf("found leaked workspace directory: %s", e.Name())
		}
	}
}

// TestEvaluateCandidate_RestrictedRepoPaths verifies that attempts to evaluate against sensitive
// system root locations or escape paths are rejected immediately with ErrSecurityViolation.
func TestEvaluateCandidate_RestrictedRepoPaths(t *testing.T) {
	evaluator := NewDockerEvaluator(DockerEvaluatorConfig{})

	sensitivePaths := []string{
		"/",
		"/etc",
		"/etc/ssl",
		"/root",
		"/proc",
		"/sys",
		"/dev",
		"../outside",
		"C:/Windows",
		"C:/Windows/System32",
	}

	for _, p := range sensitivePaths {
		t.Run(p, func(t *testing.T) {
			res, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{
				RepoPath:    p,
				FilePath:    "main.go",
				MutatedCode: "package main",
			})

			if err == nil {
				t.Fatalf("expected ErrSecurityViolation for RepoPath %q, got nil", p)
			}
			if !errors.Is(err, ErrSecurityViolation) {
				t.Fatalf("expected ErrSecurityViolation for RepoPath %q, got: %v", p, err)
			}
			if res.Outcome != OutcomeSecurityRejected {
				t.Fatalf("expected OutcomeSecurityRejected, got: %s", res.Outcome)
			}
		})
	}
}

func TestRemoteBrokerEvaluator_PrepareWorkspace(t *testing.T) {
	var archiveEntries []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/workspaces" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Sandbox-Token") != "workspace-token-1234567890" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/zip" {
			http.Error(w, "unexpected content type", http.StatusUnsupportedMediaType)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, entry := range reader.File {
			archiveEntries = append(archiveEntries, entry.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"workspace_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
	}))
	defer server.Close()

	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main"), 0o600); err != nil {
		t.Fatal(err)
	}
	evaluator := newTestRemoteBrokerEvaluator(t, server, "workspace-token-1234567890")
	workspaceID, err := evaluator.PrepareWorkspace(context.Background(), repo)
	if err != nil {
		t.Fatalf("PrepareWorkspace() error = %v", err)
	}
	if workspaceID != testBrokerWorkspaceID {
		t.Fatalf("workspace ID = %q, want %s", workspaceID, testBrokerWorkspaceID)
	}
	if len(archiveEntries) != 1 || archiveEntries[0] != "main.go" {
		t.Fatalf("uploaded archive entries = %v, want [main.go]", archiveEntries)
	}
}

func TestRemoteBrokerEvaluator_CleanupWorkspace(t *testing.T) {
	var deletedPath string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Sandbox-Token") != "workspace-token-1234567890" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		deletedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	evaluator := newTestRemoteBrokerEvaluator(t, server, "workspace-token-1234567890")
	if err := evaluator.CleanupWorkspace(context.Background(), testBrokerWorkspaceID); err != nil {
		t.Fatalf("CleanupWorkspace() error = %v", err)
	}
	if deletedPath != "/v1/workspaces/"+testBrokerWorkspaceID {
		t.Fatalf("deleted path = %q", deletedPath)
	}
}

func TestRemoteBrokerEvaluator_RequiresHTTPS(t *testing.T) {
	evaluator := NewRemoteBrokerEvaluator("http://sandbox-broker:8082", "workspace-token-1234567890")
	if err := evaluator.CheckAvailable(context.Background()); err == nil {
		t.Fatal("CheckAvailable() accepted a plaintext broker URL")
	}
}

func TestRemoteBrokerEvaluator_RequiresTLS12OrNewer(t *testing.T) {
	client, err := newRemoteBrokerHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("HTTP transport has type %T, want *http.Transport", client.Transport)
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("TLS minimum = %d, want %d", transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}
}

func TestAnnealer_UsesOneRemoteWorkspaceForEvolution(t *testing.T) {
	repo, relPath := setupTestRepo(t)
	fileBytes, err := os.ReadFile(filepath.Join(repo, relPath))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(fileBytes), "return a + b")
	evaluator := &remoteWorkspaceMock{workspaceID: testBrokerWorkspaceID}
	evaluator.EvaluateFunc = func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
		if req.WorkspaceID != evaluator.workspaceID {
			t.Fatalf("evaluation workspace ID = %q, want %q", req.WorkspaceID, evaluator.workspaceID)
		}
		return &EvaluationResult{TestsPassed: true, Outcome: OutcomePassed}, nil
	}
	annealer := &Annealer{
		Mutator:         &mockMutator{mutations: []string{"return a + b"}},
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40},
	}
	_, err = annealer.Evolve(context.Background(), repo, relPath, "Add", start, start+len("return a + b"), 50, AnnealingConfig{
		InitialTemp: 10,
		CoolingRate: 0.5,
		MaxSteps:    1,
	}, make(chan StepTelemetry, 1))
	if err != nil {
		t.Fatalf("Evolve() error = %v", err)
	}
	if evaluator.prepareCalls != 1 || evaluator.cleanupCalls != 1 || evaluator.evaluateCalls != 1 {
		t.Fatalf("workspace calls: prepare=%d evaluate=%d cleanup=%d", evaluator.prepareCalls, evaluator.evaluateCalls, evaluator.cleanupCalls)
	}
}

type remoteWorkspaceMock struct {
	MockEvaluator
	workspaceID   string
	prepareCalls  int
	evaluateCalls int
	cleanupCalls  int
}

func (m *remoteWorkspaceMock) PrepareWorkspace(context.Context, string) (string, error) {
	m.prepareCalls++
	return m.workspaceID, nil
}

func (m *remoteWorkspaceMock) CleanupWorkspace(context.Context, string) error {
	m.cleanupCalls++
	return nil
}

func (m *remoteWorkspaceMock) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	m.evaluateCalls++
	return m.EvaluateFunc(ctx, req)
}

func newTestRemoteBrokerEvaluator(t *testing.T, server *httptest.Server, token string) *RemoteBrokerEvaluator {
	t.Helper()
	certPath := filepath.Join(t.TempDir(), "broker-ca.pem")
	certificate := server.Certificate()
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("write test broker CA: %v", err)
	}
	t.Setenv("SANDBOX_TLS_CA_FILE", certPath)
	return NewRemoteBrokerEvaluator(server.URL, token)
}
