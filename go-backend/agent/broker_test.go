package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	evaluator := NewRemoteBrokerEvaluator("http://127.0.0.1:54321", "test-token-123456")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := evaluator.EvaluateCandidate(ctx, EvaluationRequest{
		RepoPath:    t.TempDir(),
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	badEvaluator := NewRemoteBrokerEvaluator(server.URL, "wrong-token-value-here")
	ctx := context.Background()

	err := badEvaluator.CheckAvailable(ctx)
	if err == nil || !errors.Is(err, ErrBrokerAuthenticationFailed) {
		t.Fatalf("expected ErrBrokerAuthenticationFailed, got: %v", err)
	}

	res, err := badEvaluator.EvaluateCandidate(ctx, EvaluationRequest{
		RepoPath:    t.TempDir(),
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
	goodEvaluator := NewRemoteBrokerEvaluator(server.URL, expectedToken)
	if err := goodEvaluator.CheckAvailable(ctx); err != nil {
		t.Fatalf("expected CheckAvailable to succeed with valid token, got: %v", err)
	}
}

// TestRemoteBrokerEvaluator_TimeoutPropagation verifies that request deadlines and context
// cancellations are propagated and result in safe timeout classification.
func TestRemoteBrokerEvaluator_TimeoutPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	evaluator := NewRemoteBrokerEvaluator(server.URL, "valid-token")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	res, err := evaluator.EvaluateCandidate(ctx, EvaluationRequest{
		RepoPath:    t.TempDir(),
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			if string(req.OriginalBytes) != string(origContent) {
				http.Error(w, "original bytes mismatch", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(EvaluationResult{
				TestsPassed: true,
				Outcome:     OutcomePassed,
				Output:      expectedOutput,
				Duration:    45 * time.Millisecond,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	evaluator := NewRemoteBrokerEvaluator(server.URL, "auth-token-1234")

	res, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{
		RepoPath:      t.TempDir(),
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
	t.Setenv("SANDBOX_SERVICE_URL", "http://sandbox-broker:8082")
	t.Setenv("SANDBOX_SERVICE_TOKEN", "super-secret-token")
	eval2 := NewEvaluator(DockerEvaluatorConfig{})
	remoteEval, ok := eval2.(*RemoteBrokerEvaluator)
	if !ok {
		t.Fatalf("expected *RemoteBrokerEvaluator when SANDBOX_SERVICE_URL is set, got %T", eval2)
	}
	if remoteEval.brokerURL != "http://sandbox-broker:8082" {
		t.Fatalf("expected brokerURL 'http://sandbox-broker:8082', got %q", remoteEval.brokerURL)
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
