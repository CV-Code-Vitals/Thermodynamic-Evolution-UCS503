package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Helper to create a disposable Go package inside t.TempDir()
func setupTestRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()

	modContent := "module ucs503-sample\n\ngo 1.25\n"
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte(modContent), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	srcContent := `package sample

func Add(a, b int) int {
	return a + b
}
`
	srcPath := filepath.Join(repoDir, "calc.go")
	if err := os.WriteFile(srcPath, []byte(srcContent), 0o600); err != nil {
		t.Fatalf("write calc.go: %v", err)
	}

	testContent := `package sample

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Fatalf("expected 5, got %d", Add(2, 3))
	}
}
`
	if err := os.WriteFile(filepath.Join(repoDir, "calc_test.go"), []byte(testContent), 0o600); err != nil {
		t.Fatalf("write calc_test.go: %v", err)
	}

	return repoDir, "calc.go"
}

// ── Test 1 & 2 & 10: Annealer flow with MockEvaluator ─────────────────────────

type mockMutator struct {
	mutations []string
	callCount int
}

func (m *mockMutator) ProposeMutation(_ context.Context, _ string, _ string) (string, error) {
	if m.callCount < len(m.mutations) {
		res := m.mutations[m.callCount]
		m.callCount++
		return res, nil
	}
	return "return a + b", nil
}

func TestAnnealer_MockEvaluator_SuccessfulOptimization(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	// Function body is "return a + b\n" between bytes 37 and 52
	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
			// Simulate passed tests
			return &EvaluationResult{
				TestsPassed:      true,
				Outcome:          OutcomePassed,
				UpdatedFileBytes: append(append([]byte{}, req.OriginalBytes[:req.StartByte]...), []byte(req.MutatedCode)...),
			}, nil
		},
	}

	mutator := &mockMutator{
		mutations: []string{
			"// optimized body\n\treturn a + b",
		},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40.0},
	}

	cfg := AnnealingConfig{
		InitialTemp: 100.0,
		CoolingRate: 0.85,
		MaxSteps:    3,
	}

	ch := make(chan StepTelemetry, 10)
	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 50.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if res.TotalSteps != 3 {
		t.Errorf("expected 3 steps, got %d", res.TotalSteps)
	}

	// Verify telemetry events were generated
	count := 0
	for len(ch) > 0 {
		telem := <-ch
		count++
		if !telem.TestsPassed {
			t.Errorf("step %d expected TestsPassed=true", telem.Step)
		}
	}
	if count != 3 {
		t.Errorf("expected 3 telemetry events, got %d", count)
	}
}

func TestAnnealer_CandidateTestFailure(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, _ EvaluationRequest) (*EvaluationResult, error) {
			// Simulate test failure
			return &EvaluationResult{
				TestsPassed: false,
				Outcome:     OutcomeTestFailed,
				Output:      "--- FAIL: TestAdd (0.00s)",
			}, nil
		},
	}

	mutator := &mockMutator{
		mutations: []string{
			"return a - b", // Broken logic
		},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40.0},
	}

	cfg := AnnealingConfig{
		InitialTemp: 10.0,
		CoolingRate: 0.5,
		MaxSteps:    1,
	}

	ch := make(chan StepTelemetry, 5)
	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 50.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	// Verify the broken code was rejected
	if res.OptimizedCode != "return a + b" {
		t.Errorf("expected original code to be preserved, got %q", res.OptimizedCode)
	}

	telem := <-ch
	if telem.TestsPassed {
		t.Error("expected TestsPassed=false on failing test")
	}
	if telem.Accepted {
		t.Error("expected Accepted=false on failing test")
	}

	// Verify original file on host was NEVER altered
	postBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(postBytes) != string(originalBytes) {
		t.Error("host file was modified despite candidate test failure!")
	}
}

// ── Test 3: Timeout & Cancellation Handling ───────────────────────────────────

func TestAnnealer_ExecutionTimeout(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, _ EvaluationRequest) (*EvaluationResult, error) {
			return &EvaluationResult{
				TestsPassed: false,
				Outcome:     OutcomeTimeout,
				Output:      "execution timed out",
				Error:       ErrExecutionTimeout,
			}, nil
		},
	}

	mutator := &mockMutator{mutations: []string{"for {}"}} // Infinite loop
	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40.0},
	}

	cfg := AnnealingConfig{InitialTemp: 10.0, CoolingRate: 0.5, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)
	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 50.0, cfg, ch)
	if err != nil {
		t.Fatalf("expected graceful step rejection on timeout, got error: %v", err)
	}

	if res.OptimizedCode != "return a + b" {
		t.Errorf("expected original code preserved on timeout, got %q", res.OptimizedCode)
	}

	telem := <-ch
	if telem.Reason != string(OutcomeTimeout) {
		t.Errorf("expected reason '%s', got '%s'", OutcomeTimeout, telem.Reason)
	}
}

func TestAnnealer_ContextCancellation(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}

	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	evaluator := &MockEvaluator{}
	mutator := &mockMutator{}
	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40.0},
	}

	cfg := AnnealingConfig{InitialTemp: 100.0, CoolingRate: 0.8, MaxSteps: 5}
	ch := make(chan StepTelemetry, 5)

	_, err = annealer.Evolve(ctx, repoDir, relPath, "Add", startByte, endByte, 50.0, cfg, ch)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

// ── Test 4 & 5: Missing Sandbox & Fail-Closed Protection ───────────────────────

func TestDockerEvaluator_MissingSandboxFailsClosed(t *testing.T) {
	evaluator := NewDockerEvaluator(DockerEvaluatorConfig{
		DockerBinary: "non_existent_docker_binary_xyz_123",
	})

	ctx := context.Background()
	err := evaluator.CheckAvailable(ctx)
	if !errors.Is(err, ErrSandboxUnavailable) {
		t.Fatalf("expected ErrSandboxUnavailable, got: %v", err)
	}

	orig := []byte("package main\n\nfunc main() {}\n")
	startByte := strings.Index(string(orig), "func main() {}")
	endByte := startByte + len("func main() {}")

	req := EvaluationRequest{
		RepoPath:      t.TempDir(),
		FilePath:      "main.go",
		OriginalBytes: orig,
		MutatedCode:   "func main() {}",
		StartByte:     startByte,
		EndByte:       endByte,
	}

	res, err := evaluator.EvaluateCandidate(ctx, req)
	if !errors.Is(err, ErrSandboxUnavailable) {
		t.Fatalf("expected ErrSandboxUnavailable from EvaluateCandidate, got: %v", err)
	}
	if res.Outcome != OutcomeInfrastructureFailure {
		t.Errorf("expected OutcomeInfrastructureFailure, got %s", res.Outcome)
	}
	if res.TestsPassed {
		t.Error("TestsPassed must be false when sandbox is unavailable")
	}
}

func TestAnnealer_FailsClosedWhenEvaluatorMissing(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	mutator := &mockMutator{}

	// Explicitly nil Evaluator
	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       nil,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40.0},
	}

	cfg := AnnealingConfig{InitialTemp: 10.0, CoolingRate: 0.5, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)

	_, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", 0, 10, 50.0, cfg, ch)
	if !errors.Is(err, ErrHostExecutionBlocked) {
		t.Fatalf("expected ErrHostExecutionBlocked when Evaluator is nil, got: %v", err)
	}
}

func TestAnnealer_FailsClosedWhenEnergyEvaluatorMissing(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	mutator := &mockMutator{}
	evaluator := &MockEvaluator{}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: nil,
	}

	cfg := AnnealingConfig{InitialTemp: 10.0, CoolingRate: 0.5, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)

	_, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", 0, 10, 50.0, cfg, ch)
	if err == nil || !strings.Contains(err.Error(), "EnergyEvaluator") {
		t.Fatalf("expected error when EnergyEvaluator is nil, got: %v", err)
	}
}

func TestFailClosedEvaluator(t *testing.T) {
	evaluator := &FailClosedEvaluator{Reason: "maintenance mode"}
	res, err := evaluator.EvaluateCandidate(context.Background(), EvaluationRequest{})
	if !errors.Is(err, ErrHostExecutionBlocked) {
		t.Fatalf("expected ErrHostExecutionBlocked, got %v", err)
	}
	if res.TestsPassed {
		t.Error("TestsPassed must be false")
	}
	if res.Outcome != OutcomeSecurityRejected {
		t.Errorf("expected OutcomeSecurityRejected, got %s", res.Outcome)
	}
}

// ── Test 6: Bounded Output Handling ──────────────────────────────────────────

func TestLimitedBuffer_BoundingAndTruncation(t *testing.T) {
	var buf limitedBuffer
	buf.maxBytes = 256

	// Write 1000 bytes
	payload := strings.Repeat("A", 1000)
	n, err := buf.Write([]byte(payload))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != 1000 {
		t.Errorf("expected write to report full consumption, got %d", n)
	}

	str := buf.String()
	if !strings.Contains(str, "[...OUTPUT TRUNCATED: byte limit reached...]") {
		t.Error("expected truncation indicator in output string")
	}

	// Verify raw buffer content does not exceed maxBytes
	if int64(buf.buf.Len()) != 256 {
		t.Errorf("expected buffer len 256, got %d", buf.buf.Len())
	}
}

// ── Test 7: Temporary Workspace Cleanup ───────────────────────────────────────

func TestCopyProjectFiles_SanitizesSecretsAndBins(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Legitimate files
	_ = os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0o600)
	_ = os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module test"), 0o600)

	// Files that MUST be filtered out
	_ = os.WriteFile(filepath.Join(srcDir, ".env"), []byte("SECRET=123"), 0o600)
	_ = os.WriteFile(filepath.Join(srcDir, ".env.production"), []byte("AWS_SECRET=xyz"), 0o600)
	_ = os.WriteFile(filepath.Join(srcDir, "app.exe"), []byte("MZbinary"), 0o600)
	_ = os.WriteFile(filepath.Join(srcDir, "test.zip"), []byte("PKzip"), 0o600)

	gitDir := filepath.Join(srcDir, ".git")
	_ = os.MkdirAll(gitDir, 0o750)
	_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0o600)

	if err := copyProjectFiles(srcDir, destDir); err != nil {
		t.Fatalf("copyProjectFiles failed: %v", err)
	}

	// Verify legitimate files copied
	if _, err := os.Stat(filepath.Join(destDir, "main.go")); err != nil {
		t.Error("main.go was not copied")
	}
	if _, err := os.Stat(filepath.Join(destDir, "go.mod")); err != nil {
		t.Error("go.mod was not copied")
	}

	// Verify secrets and binaries were strictly excluded
	if _, err := os.Stat(filepath.Join(destDir, ".env")); !os.IsNotExist(err) {
		t.Error(".env was copied to sandbox workspace!")
	}
	if _, err := os.Stat(filepath.Join(destDir, "app.exe")); !os.IsNotExist(err) {
		t.Error("app.exe was copied to sandbox workspace!")
	}
	if _, err := os.Stat(filepath.Join(destDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory was copied to sandbox workspace!")
	}
}

// ── Test 8: Malformed & Rejected Candidates (Syntax & Security) ───────────────

func TestValidateCandidateSource_SyntaxError(t *testing.T) {
	err := validateCandidateSource("main.go", "func broken(", []byte("package main\nfunc broken(\n"))
	if err == nil {
		t.Fatal("expected syntax error for broken Go syntax")
	}
	if !strings.Contains(err.Error(), "syntax error") {
		t.Errorf("expected 'syntax error' in message, got: %v", err)
	}
}

func TestValidateCandidateSource_SecurityDirectivesRejected(t *testing.T) {
	tests := []struct {
		name    string
		snippet string
	}{
		{"cgo import", `import "C"`},
		{"cgo compiler directive", `//go:cgo_import_dynamic`},
		{"generate directive", `//go:generate evil.sh`},
		{"linkname directive", `//go:linkname secretSymbol`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullContent := fmt.Sprintf("package main\n%s\nfunc F() {}\n", tt.snippet)
			err := validateCandidateSource("main.go", tt.snippet, []byte(fullContent))
			if !errors.Is(err, ErrSecurityViolation) {
				t.Fatalf("expected ErrSecurityViolation for %s, got: %v", tt.name, err)
			}
		})
	}
}

// ── Test 9: Prevention of Insecure Host-Execution Fallback ─────────────────────

func TestDockerEvaluator_NeverFallsBackToHost(t *testing.T) {
	// Provide a runner checker that always errors out
	evaluator := NewDockerEvaluator(DockerEvaluatorConfig{
		RunnerChecker: func(_ context.Context) error {
			return errors.New("docker daemon is offline")
		},
	})

	orig := []byte("package calc\n\nfunc Add() {}\n")
	startByte := strings.Index(string(orig), "func Add() {}")
	endByte := startByte + len("func Add() {}")

	req := EvaluationRequest{
		RepoPath:      t.TempDir(),
		FilePath:      "calc.go",
		OriginalBytes: orig,
		MutatedCode:   "func Add() {}",
		StartByte:     startByte,
		EndByte:       endByte,
	}

	res, err := evaluator.EvaluateCandidate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error when container daemon is offline")
	}

	// Must fail closed, never return success
	if res.TestsPassed {
		t.Fatal("TestsPassed must NEVER be true when Docker is offline")
	}

	if res.Outcome != OutcomeInfrastructureFailure {
		t.Errorf("expected OutcomeInfrastructureFailure, got %s", res.Outcome)
	}
}

// ── Test 10: Docker Integration Test (detects live daemon) ────────────────────

func TestDockerEvaluator_LiveExecution(t *testing.T) {
	evaluator := NewDockerEvaluator(DockerEvaluatorConfig{
		Image: os.Getenv("SANDBOX_RUNNER_IMAGE"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := evaluator.CheckAvailable(ctx); err != nil {
		t.Skipf("Skipping live Docker sandbox integration test: Docker daemon is not available on this host: %v", err)
	}

	// If Docker IS available, run a live evaluation
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	req := EvaluationRequest{
		RepoPath:      repoDir,
		FilePath:      relPath,
		OriginalBytes: bytes,
		MutatedCode:   "return a + b",
		StartByte:     startByte,
		EndByte:       endByte,
	}

	res, err := evaluator.EvaluateCandidate(context.Background(), req)
	if err != nil {
		t.Fatalf("Live Docker evaluation failed: %v", err)
	}

	if !res.TestsPassed {
		t.Errorf("expected live test to pass in container, outcome: %s, duration: %v, err: %v, output:\n%s", res.Outcome, res.Duration, res.Error, res.Output)
	}
}
