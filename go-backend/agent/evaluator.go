package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Standard error definitions for candidate evaluation.
var (
	ErrSandboxUnavailable   = errors.New("sandbox execution unavailable: secure container runtime not detected or daemon not running")
	ErrExecutionTimeout     = errors.New("candidate execution timed out")
	ErrSecurityViolation    = errors.New("candidate code violates execution security policy")
	ErrInvalidRequest       = errors.New("invalid candidate evaluation request")
	ErrHostExecutionBlocked = errors.New("untrusted candidate execution on host is strictly forbidden")
)

// EvaluationOutcome represents the structured classification of a candidate run.
type EvaluationOutcome string

const (
	OutcomePassed                EvaluationOutcome = "passed"
	OutcomeTestFailed            EvaluationOutcome = "tests_failed"
	OutcomeTimeout               EvaluationOutcome = "timeout"
	OutcomeCompilationFailed     EvaluationOutcome = "compilation_failed"
	OutcomeSecurityRejected      EvaluationOutcome = "security_rejected"
	OutcomeInfrastructureFailure EvaluationOutcome = "infrastructure_failure"
)

// ResourceLimits specifies runtime isolation boundaries for candidate evaluation.
type ResourceLimits struct {
	Timeout        time.Duration `json:"timeout"`
	MaxMemoryMB    int           `json:"max_memory_mb"`
	MaxCPUs        float64       `json:"max_cpus"`
	MaxPids        int           `json:"max_pids"`
	MaxOutputBytes int64         `json:"max_output_bytes"`
}

// DefaultResourceLimits returns hardened default bounds.
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		Timeout:        60 * time.Second,
		MaxMemoryMB:    1024,
		MaxCPUs:        2.0,
		MaxPids:        128,
		MaxOutputBytes: 64 * 1024, // 64 KiB
	}
}

// EvaluationRequest contains the code and context needed to test a mutation safely.
type EvaluationRequest struct {
	RepoPath      string         `json:"repo_path"`
	FilePath      string         `json:"file_path"`
	OriginalBytes []byte         `json:"-"`
	MutatedCode   string         `json:"mutated_code"`
	StartByte     int            `json:"start_byte"`
	EndByte       int            `json:"end_byte"`
	Limits        ResourceLimits `json:"limits"`
}

// EvaluationResult contains the outcome, captured logs, and duration of an evaluation.
type EvaluationResult struct {
	TestsPassed      bool              `json:"tests_passed"`
	Outcome          EvaluationOutcome `json:"outcome"`
	Output           string            `json:"output"`
	Duration         time.Duration     `json:"duration"`
	UpdatedFileBytes []byte            `json:"-"`
	Error            error             `json:"error,omitempty"`
}

// CandidateEvaluator defines the boundary through which all candidate code must be evaluated.
type CandidateEvaluator interface {
	EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error)
}

// DockerEvaluatorConfig holds settings for the Docker sandbox evaluator.
type DockerEvaluatorConfig struct {
	Image          string
	DockerBinary   string
	ResourceLimits ResourceLimits
	RunnerChecker  func(ctx context.Context) error
}

// DockerEvaluator implements CandidateEvaluator using an ephemeral, unprivileged Docker container.
type DockerEvaluator struct {
	image          string
	dockerBinary   string
	limits         ResourceLimits
	runnerChecker  func(ctx context.Context) error
}

// NewDockerEvaluator constructs a DockerEvaluator with sensible defaults.
func NewDockerEvaluator(cfg DockerEvaluatorConfig) *DockerEvaluator {
	img := cfg.Image
	if strings.TrimSpace(img) == "" {
		img = os.Getenv("SANDBOX_RUNNER_IMAGE")
	}
	if strings.TrimSpace(img) == "" {
		img = "ucs503-sandbox-runner:latest"
	}
	bin := cfg.DockerBinary
	if strings.TrimSpace(bin) == "" {
		bin = "docker"
	}
	limits := cfg.ResourceLimits
	def := DefaultResourceLimits()
	if limits.Timeout <= 0 {
		limits.Timeout = def.Timeout
	}
	if limits.MaxMemoryMB <= 0 {
		limits.MaxMemoryMB = def.MaxMemoryMB
	}
	if limits.MaxCPUs <= 0 {
		limits.MaxCPUs = def.MaxCPUs
	}
	if limits.MaxPids <= 0 {
		limits.MaxPids = def.MaxPids
	}
	if limits.MaxOutputBytes <= 0 {
		limits.MaxOutputBytes = def.MaxOutputBytes
	}

	return &DockerEvaluator{
		image:         img,
		dockerBinary:  bin,
		limits:        limits,
		runnerChecker: cfg.RunnerChecker,
	}
}

// CheckAvailable verifies whether the Docker sandbox runtime is accessible.
func (d *DockerEvaluator) CheckAvailable(ctx context.Context) error {
	if d.runnerChecker != nil {
		return d.runnerChecker(ctx)
	}

	if _, err := exec.LookPath(d.dockerBinary); err != nil {
		return fmt.Errorf("%w: '%s' executable not found in PATH", ErrSandboxUnavailable, d.dockerBinary)
	}

	checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(checkCtx, d.dockerBinary, "info", "--format", "{{.ServerVersion}}")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: docker info failed: %v (output: %s)", ErrSandboxUnavailable, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// EvaluateCandidate isolates, validates, mounts, and runs tests for a candidate patch.
func (d *DockerEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	start := time.Now()

	limits := req.Limits
	if limits.Timeout <= 0 {
		limits = d.limits
	}

	// 1. Validate request and parameters
	if strings.TrimSpace(req.RepoPath) == "" || strings.TrimSpace(req.FilePath) == "" {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: repository path and file path are required", ErrInvalidRequest),
		}, ErrInvalidRequest
	}

	// Reject path traversal and absolute/rooted paths
	cleanInput := filepath.ToSlash(req.FilePath)
	cleanedRel := filepath.Clean(filepath.FromSlash(req.FilePath))
	if filepath.IsAbs(cleanedRel) || strings.HasPrefix(cleanInput, "/") || strings.HasPrefix(cleanInput, "\\") ||
		strings.HasPrefix(cleanedRel, "..") || filepath.VolumeName(cleanedRel) != "" {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: path traversal detected in file path '%s'", ErrSecurityViolation, req.FilePath),
		}, ErrSecurityViolation
	}

	if req.StartByte < 0 || req.EndByte < req.StartByte || req.StartByte > len(req.OriginalBytes) || req.EndByte > len(req.OriginalBytes) {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: byte range [%d:%d] out of bounds for file of size %d", ErrInvalidRequest, req.StartByte, req.EndByte, len(req.OriginalBytes)),
		}, ErrInvalidRequest
	}

	// 2. Splice candidate code into buffer
	newContent := make([]byte, 0, req.StartByte+len(req.MutatedCode)+(len(req.OriginalBytes)-req.EndByte))
	newContent = append(newContent, req.OriginalBytes[:req.StartByte]...)
	newContent = append(newContent, []byte(req.MutatedCode)...)
	newContent = append(newContent, req.OriginalBytes[req.EndByte:]...)

	// 3. Static security & syntax pre-checks (Fast-fail before container overhead)
	if err := validateCandidateSource(req.FilePath, req.MutatedCode, newContent); err != nil {
		outcome := OutcomeCompilationFailed
		if errors.Is(err, ErrSecurityViolation) {
			outcome = OutcomeSecurityRejected
		}
		return &EvaluationResult{
			TestsPassed:      false,
			Outcome:          outcome,
			Output:           err.Error(),
			Duration:         time.Since(start),
			UpdatedFileBytes: newContent,
			Error:            err,
		}, nil
	}

	// 4. Ensure sandbox runtime is available (FAIL-CLOSED: NEVER fall back to host!)
	if err := d.CheckAvailable(ctx); err != nil {
		return &EvaluationResult{
			TestsPassed:      false,
			Outcome:          OutcomeInfrastructureFailure,
			Duration:         time.Since(start),
			UpdatedFileBytes: newContent,
			Error:            err,
		}, err
	}

	// 5. Prepare isolated ephemeral workspace directory
	tempWorkspace, err := os.MkdirTemp("", "ucs503-sandbox-ws-*")
	if err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create temporary workspace: %w", err),
		}, err
	}
	defer os.RemoveAll(tempWorkspace) // Guarantee full cleanup on all exit paths
	_ = os.Chmod(tempWorkspace, 0o777)

	if err := copyProjectFiles(req.RepoPath, tempWorkspace); err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to populate sandbox workspace: %w", err),
		}, err
	}

	// Inject mutated file into the ephemeral workspace
	targetDest := filepath.Join(tempWorkspace, cleanedRel)
	relCheck, err := filepath.Rel(tempWorkspace, targetDest)
	if err != nil || strings.HasPrefix(relCheck, "..") || filepath.IsAbs(relCheck) {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: path escape detected in target destination", ErrSecurityViolation),
		}, ErrSecurityViolation
	}
	if err := os.MkdirAll(filepath.Dir(targetDest), 0o777); err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create directory for mutated file: %w", err),
		}, err
	}
	if err := os.WriteFile(targetDest, newContent, 0o666); err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to write candidate into workspace: %w", err),
		}, err
	}

	// 6. Execute in Docker sandbox with strict containment
	containerName := fmt.Sprintf("ucs503-run-%s", uuid.New().String()[:12])
	cmdCtx, cmdCancel := context.WithTimeout(ctx, limits.Timeout)
	defer cmdCancel()

	// Convert workspace path to slashes for Docker compatibility across platforms
	workspaceMount := filepath.ToSlash(tempWorkspace)

	tmpfsSizeMB := limits.MaxMemoryMB / 2
	if tmpfsSizeMB < 256 {
		tmpfsSizeMB = 256
	}
	if tmpfsSizeMB > 512 {
		tmpfsSizeMB = 512
	}

	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", "none", // NO network access
		"--read-only", // Read-only root FS
		"--tmpfs", fmt.Sprintf("/tmp:rw,exec,nosuid,nodev,mode=1777,size=%dm", tmpfsSizeMB),
		"-e", "GOCACHE=/tmp/gocache",
		"-e", "CGO_ENABLED=0",
		"--user", "10001:10001", // Non-root
		"--cap-drop", "ALL", // Drop all Linux capabilities
		"--security-opt", "no-new-privileges",
		"--memory", fmt.Sprintf("%dm", limits.MaxMemoryMB),
		"--cpus", fmt.Sprintf("%.2f", limits.MaxCPUs),
		"--pids-limit", fmt.Sprintf("%d", limits.MaxPids),
		"-v", fmt.Sprintf("%s:/workspace:rw", workspaceMount),
		"-w", "/workspace",
		d.image,
		"go", "test", "-v", "./...",
	}

	cmd := exec.CommandContext(cmdCtx, d.dockerBinary, args...)

	var outBuf limitedBuffer
	outBuf.maxBytes = limits.MaxOutputBytes
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	runErr := cmd.Run()
	duration := time.Since(start)

	// Ensure container is forcefully cleaned up if timed out or killed
	if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
		killCtx, killCancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = exec.CommandContext(killCtx, d.dockerBinary, "rm", "-f", containerName).Run()
		killCancel()

		return &EvaluationResult{
			TestsPassed:      false,
			Outcome:          OutcomeTimeout,
			Output:           outBuf.String(),
			Duration:         duration,
			UpdatedFileBytes: newContent,
			Error:            ErrExecutionTimeout,
		}, nil
	}

	if runErr != nil {
		output := outBuf.String()
		outcome := OutcomeTestFailed
		if strings.Contains(output, "syntax error") || strings.Contains(output, "build failed") {
			outcome = OutcomeCompilationFailed
		}

		return &EvaluationResult{
			TestsPassed:      false,
			Outcome:          outcome,
			Output:           output,
			Duration:         duration,
			UpdatedFileBytes: newContent,
			Error:            runErr,
		}, nil
	}

	return &EvaluationResult{
		TestsPassed:      true,
		Outcome:          OutcomePassed,
		Output:           outBuf.String(),
		Duration:         duration,
		UpdatedFileBytes: newContent,
		Error:            nil,
	}, nil
}

// validateCandidateSource inspects code for forbidden directives and basic Go syntax validity.
func validateCandidateSource(filePath, candidateSnippet string, fullContent []byte) error {
	// 1. Prohibit CGO and unsafe directives
	forbiddenDirectives := []string{
		"import \"C\"",
		"import `C`",
		"//go:cgo",
		"//go:generate",
		"//go:linkname",
	}
	for _, directive := range forbiddenDirectives {
		if strings.Contains(candidateSnippet, directive) || strings.Contains(string(fullContent), directive) {
			return fmt.Errorf("%w: forbidden directive '%s' detected", ErrSecurityViolation, directive)
		}
	}

	// 2. Syntax validation
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, filePath, fullContent, parser.AllErrors)
	if parseErr != nil {
		return fmt.Errorf("candidate syntax error: %w", parseErr)
	}

	return nil
}

// copyProjectFiles copies source files (.go, go.mod, go.sum, test data) while filtering secrets.
func copyProjectFiles(srcDir, destDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		base := info.Name()

		// Skip hidden directories and VCS
		if info.IsDir() {
			if strings.HasPrefix(base, ".") || base == "uploads" || base == "node_modules" || base == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files, environment files, credentials, and executables
		if strings.HasPrefix(base, ".") || strings.HasPrefix(base, ".env") ||
			strings.HasSuffix(base, ".exe") || strings.HasSuffix(base, ".bin") ||
			strings.HasSuffix(base, ".dll") || strings.HasSuffix(base, ".so") ||
			strings.HasSuffix(base, ".zip") {
			return nil
		}

		// Explicitly skip symlinks to prevent out-of-tree file exposure
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		// Only copy Go source files, module definitions, and text fixtures
		ext := strings.ToLower(filepath.Ext(base))
		isGoSource := ext == ".go" || ext == ".mod" || ext == ".sum" || ext == ".json" || ext == ".txt" || ext == ".yaml" || ext == ".yml"
		if !isGoSource {
			return nil
		}

		cleanRel := filepath.Clean(rel)
		if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) || filepath.VolumeName(cleanRel) != "" {
			return nil
		}

		targetPath := filepath.Join(destDir, cleanRel)
		relCheck, err := filepath.Rel(destDir, targetPath)
		if err != nil || strings.HasPrefix(relCheck, "..") || filepath.IsAbs(relCheck) {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o777); err != nil {
			return err
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// limitedBuffer captures up to maxBytes without unbounded memory consumption.
type limitedBuffer struct {
	maxBytes int64
	buf      bytes.Buffer
	overflow bool
}

func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	rem := b.maxBytes - int64(b.buf.Len())
	if rem <= 0 {
		b.overflow = true
		return len(p), nil // Consume and discard
	}
	if int64(len(p)) > rem {
		b.buf.Write(p[:rem])
		b.overflow = true
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) String() string {
	s := b.buf.String()
	if b.overflow {
		s += "\n[...OUTPUT TRUNCATED: byte limit reached...]"
	}
	return s
}

// FailClosedEvaluator unconditionally blocks execution, ensuring fail-closed safety.
type FailClosedEvaluator struct {
	Reason string
}

func (f *FailClosedEvaluator) EvaluateCandidate(_ context.Context, _ EvaluationRequest) (*EvaluationResult, error) {
	reason := f.Reason
	if reason == "" {
		reason = "execution of untrusted candidate code on host is forbidden"
	}
	return &EvaluationResult{
		TestsPassed: false,
		Outcome:     OutcomeSecurityRejected,
		Error:       fmt.Errorf("%w: %s", ErrHostExecutionBlocked, reason),
	}, ErrHostExecutionBlocked
}

// MockEvaluator enables deterministic unit tests without requiring a live container runtime.
type MockEvaluator struct {
	EvaluateFunc func(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error)
}

func (m *MockEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	if m.EvaluateFunc != nil {
		return m.EvaluateFunc(ctx, req)
	}
	return &EvaluationResult{
		TestsPassed: true,
		Outcome:     OutcomePassed,
	}, nil
}
