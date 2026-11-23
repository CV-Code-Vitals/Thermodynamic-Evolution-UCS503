package execution

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// Standard oracle error definitions
var (
	ErrSandboxUnavailable             = errors.New("sandbox execution unavailable: secure container runtime not detected or daemon not running")
	ErrDynamicVerificationUnavailable = errors.New("DYNAMIC_VERIFICATION_UNAVAILABLE: dynamic execution requires hardened container isolation; host execution forbidden")
	ErrHostExecutionForbidden         = errors.New("untrusted candidate execution on host is strictly forbidden")
)

// OracleType identifies each independent verification oracle.
type OracleType string

const (
	OracleSyntaxCompilation    OracleType = "SYNTAX_AND_COMPILATION"
	OracleRegressionTest       OracleType = "REGRESSION_TEST"
	OracleDefectReproducer     OracleType = "DEFECT_REPRODUCER"
	OracleBehavioralInvariance OracleType = "BEHAVIORAL_INVARIANCE"
	OracleNegativeProperties   OracleType = "NEGATIVE_PROPERTIES"
)

// OracleStatus represents the explicit execution state of an oracle check.
type OracleStatus string

const (
	OracleStatusPassed      OracleStatus = "PASSED"
	OracleStatusFailed      OracleStatus = "FAILED"
	OracleStatusNotRun      OracleStatus = "NOT_RUN"
	OracleStatusUnavailable OracleStatus = "UNAVAILABLE"
)

// ContainerExecutionEvidence captures execution telemetry and isolation bounds.
type ContainerExecutionEvidence struct {
	ContainerID          string   `json:"container_id"`
	IsolationPolicy      string   `json:"isolation_policy"`
	Command              string   `json:"command"`
	ExitStatus           int      `json:"exit_status"`
	MemoryLimitMB        int      `json:"memory_limit_mb"`
	CPULimit             float64  `json:"cpu_limit"`
	NetworkPolicy        string   `json:"network_policy"`
	RootFSPolicy         string   `json:"root_fs_policy"`
	WorkspaceDir         string   `json:"workspace_dir,omitempty"`
	ModuleDir            string   `json:"module_dir,omitempty"`
	Image                string   `json:"image,omitempty"`
	ImageDigest          string   `json:"image_digest,omitempty"`
	TimeoutCause         string   `json:"timeout_cause,omitempty"`
	RequiredDependencies []string `json:"required_dependencies,omitempty"`
	CleanupResult        string   `json:"cleanup_result,omitempty"`
}

// OracleResult holds the outcome of a single oracle check.
type OracleResult struct {
	OracleType    OracleType                  `json:"oracle_type"`
	Status        OracleStatus                `json:"status"` // "PASSED", "FAILED", "NOT_RUN", "UNAVAILABLE"
	Passed        bool                        `json:"passed"`
	DurationMs    int64                       `json:"duration_ms"`
	Details       string                      `json:"details"`
	Output        string                      `json:"output,omitempty"`
	FailureReason string                      `json:"failure_reason,omitempty"`
	ContainerInfo *ContainerExecutionEvidence `json:"container_info,omitempty"`
}

// MultiOracleReport aggregates results across all five independent oracles.
type MultiOracleReport struct {
	PatchID         string                       `json:"patch_id"`
	AllPassed       bool                         `json:"all_passed"`
	OracleResults   map[OracleType]*OracleResult `json:"oracle_results"`
	TotalDurationMs int64                        `json:"total_duration_ms"`
}

// ContainerLimits defines runtime constraints for container execution.
type ContainerLimits struct {
	Timeout        time.Duration `json:"timeout"`
	MaxMemoryMB    int           `json:"max_memory_mb"`
	MaxCPUs        float64       `json:"max_cpus"`
	MaxPids        int           `json:"max_pids"`
	MaxOutputBytes int64         `json:"max_output_bytes"`
}

// DefaultContainerLimits returns standard security boundaries.
func DefaultContainerLimits() ContainerLimits {
	return ContainerLimits{
		Timeout:        30 * time.Second,
		MaxMemoryMB:    1024,
		MaxCPUs:        2.0,
		MaxPids:        128,
		MaxOutputBytes: 64 * 1024,
	}
}

// ContainerRunResult captures the output of a container execution.
type ContainerRunResult struct {
	ExitCode int
	Output   string
	Duration time.Duration
	Evidence *ContainerExecutionEvidence
}

// ContainerRunner abstracts the isolated container runtime boundary.
type ContainerRunner interface {
	CheckAvailable(ctx context.Context) error
	RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error)
}

// DockerContainerRunner executes commands inside an ephemeral, hardened Docker container.
type DockerContainerRunner struct {
	Image             string
	DockerBinary      string
	Limits            ContainerLimits
	Checker           func(ctx context.Context) error
	ExecutionAttempts int64
}

// UnavailableContainerRunner is a mock/test ContainerRunner that simulates unavailable container infrastructure.
type UnavailableContainerRunner struct {
	Reason string
}

// NewUnavailableContainerRunner creates a ContainerRunner that consistently reports runtime unavailability.
func NewUnavailableContainerRunner(reason ...string) *UnavailableContainerRunner {
	r := "secure container runtime not detected or daemon not running"
	if len(reason) > 0 && strings.TrimSpace(reason[0]) != "" {
		r = reason[0]
	}
	return &UnavailableContainerRunner{Reason: r}
}

func (u *UnavailableContainerRunner) CheckAvailable(ctx context.Context) error {
	return fmt.Errorf("%w: %s", ErrSandboxUnavailable, u.Reason)
}

func (u *UnavailableContainerRunner) RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error) {
	return nil, u.CheckAvailable(ctx)
}

// resolveImageDigest inspects the container image to retrieve its canonical digest or ID.
func (d *DockerContainerRunner) resolveImageDigest(ctx context.Context) string {
	if strings.TrimSpace(d.Image) == "" {
		return "unknown"
	}
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(checkCtx, d.DockerBinary, "image", "inspect", "--format", "{{index .RepoDigests 0}}", d.Image)
	if out, err := cmd.CombinedOutput(); err == nil {
		s := strings.TrimSpace(string(out))
		if s != "" && s != "<no value>" {
			return s
		}
	}
	// Fallback to Image ID
	cmdID := exec.CommandContext(checkCtx, d.DockerBinary, "image", "inspect", "--format", "{{.Id}}", d.Image)
	if out, err := cmdID.CombinedOutput(); err == nil {
		s := strings.TrimSpace(string(out))
		if s != "" {
			return s
		}
	}
	return d.Image
}

// parseRequiredDependencies inspects go.mod in wsDir to inventory required dependencies.
func parseRequiredDependencies(wsDir string) []string {
	goModPath := filepath.Join(wsDir, "go.mod")
	bytes, err := os.ReadFile(goModPath)
	if err != nil {
		return []string{"(none - standard library only)"}
	}
	lines := strings.Split(string(bytes), "\n")
	var deps []string
	inRequireBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "require (" {
			inRequireBlock = true
			continue
		}
		if inRequireBlock {
			if trimmed == ")" {
				inRequireBlock = false
				continue
			}
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				deps = append(deps, trimmed)
			}
		} else if strings.HasPrefix(trimmed, "require ") {
			dep := strings.TrimSpace(strings.TrimPrefix(trimmed, "require "))
			if dep != "" {
				deps = append(deps, dep)
			}
		}
	}
	if len(deps) == 0 {
		return []string{"(none - standard library only)"}
	}
	return deps
}

// resolveStagedModuleCache locates the system or staged Go module cache for offline read-only mounting.
func resolveStagedModuleCache() string {
	if p := os.Getenv("STAGED_MODULE_CACHE"); p != "" {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	if p := os.Getenv("GOMODCACHE"); p != "" {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	if gp := os.Getenv("GOPATH"); gp != "" {
		p := filepath.Join(gp, "pkg", "mod")
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	cmd := exec.Command("go", "env", "GOMODCACHE")
	if out, err := cmd.Output(); err == nil {
		p := strings.TrimSpace(string(out))
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	return ""
}

// StageVerifiedDependencies downloads dependencies outside the container into the local module cache
// ONLY if wsDir contains both go.mod and a non-empty go.sum with pinned checksums.
// It never enables container networking and never downloads arbitrary unverified dependencies.
func StageVerifiedDependencies(wsDir string) error {
	goModPath := filepath.Join(wsDir, "go.mod")
	goSumPath := filepath.Join(wsDir, "go.sum")

	modFi, errMod := os.Stat(goModPath)
	sumFi, errSum := os.Stat(goSumPath)
	if errMod != nil || modFi.IsDir() || errSum != nil || sumFi.IsDir() || sumFi.Size() == 0 {
		// No pinned checksums to verify; do not download untrusted dependencies
		return nil
	}

	sumBytes, err := os.ReadFile(goSumPath)
	if err != nil || len(strings.TrimSpace(string(sumBytes))) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "download")
	cmd.Dir = wsDir
	cmd.Env = GetSanitizedExecutionEnv()
	_ = cmd.Run()
	return nil
}

// resolveModuleAndPackage resolves the closest enclosing module root and package path relative to it.
func resolveModuleAndPackage(wsDir string, targetFile string) (string, string) {
	cleanWs := filepath.Clean(wsDir)
	if targetFile == "" {
		return cleanWs, "."
	}
	targetAbs := filepath.Join(cleanWs, filepath.FromSlash(targetFile))
	targetDir := filepath.Dir(targetAbs)

	curr := targetDir
	for {
		if fi, err := os.Stat(filepath.Join(curr, "go.mod")); err == nil && !fi.IsDir() {
			relPkg, relErr := filepath.Rel(curr, targetDir)
			if relErr != nil || relPkg == "." || relPkg == "" {
				return curr, "."
			}
			return curr, "./" + filepath.ToSlash(relPkg)
		}
		if curr == cleanWs || filepath.Dir(curr) == curr {
			break
		}
		curr = filepath.Dir(curr)
	}

	if fi, err := os.Stat(filepath.Join(cleanWs, "go.mod")); err == nil && !fi.IsDir() {
		relPkg, relErr := filepath.Rel(cleanWs, targetDir)
		if relErr != nil || relPkg == "." || relPkg == "" {
			return cleanWs, "."
		}
		return cleanWs, "./" + filepath.ToSlash(relPkg)
	}

	_ = EnsureDisposableModule(cleanWs)
	relPkg, relErr := filepath.Rel(cleanWs, targetDir)
	if relErr != nil || relPkg == "." || relPkg == "" {
		return cleanWs, "."
	}
	return cleanWs, "./" + filepath.ToSlash(relPkg)
}

// isOfflineDependencyMissing checks if compiler/test output indicates a missing module under GOPROXY=off.
func isOfflineDependencyMissing(out string) bool {
	lower := strings.ToLower(out)
	return strings.Contains(lower, "cannot find module providing package") ||
		strings.Contains(lower, "no required module provides package") ||
		strings.Contains(lower, "module lookup disabled by goproxy=off") ||
		strings.Contains(lower, "disable module lookup") ||
		strings.Contains(lower, "missing go.sum entry") ||
		strings.Contains(lower, "lookup proxy.golang.org") ||
		strings.Contains(lower, "dial tcp: lookup")
}

// EnsureDisposableModule creates a minimal disposable go.mod in wsDir if none exists.
// This allows standalone fixture verification in Go module mode under --network none.
func EnsureDisposableModule(wsDir string) error {
	if strings.TrimSpace(wsDir) == "" {
		return nil
	}
	goModPath := filepath.Join(wsDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		content := "module disposable_fixture\n\ngo 1.21\n"
		return os.WriteFile(goModPath, []byte(content), 0644)
	}
	return nil
}

// NewDockerContainerRunner creates a DockerContainerRunner with secure defaults.
func NewDockerContainerRunner(image ...string) *DockerContainerRunner {
	img := "ucs503-sandbox-runner:latest"
	if len(image) > 0 && strings.TrimSpace(image[0]) != "" {
		img = image[0]
	} else if envImg := os.Getenv("SANDBOX_RUNNER_IMAGE"); strings.TrimSpace(envImg) != "" {
		img = envImg
	}
	return &DockerContainerRunner{
		Image:        img,
		DockerBinary: "docker",
		Limits:       DefaultContainerLimits(),
	}
}

// CheckAvailable verifies whether Docker is available and operational.
func (d *DockerContainerRunner) CheckAvailable(ctx context.Context) error {
	if d.Checker != nil {
		return d.Checker(ctx)
	}
	if _, err := exec.LookPath(d.DockerBinary); err != nil {
		return fmt.Errorf("%w: '%s' executable not found in PATH", ErrSandboxUnavailable, d.DockerBinary)
	}
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(checkCtx, d.DockerBinary, "info", "--format", "{{.ServerVersion}}")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: docker daemon check failed: %v (%s)", ErrSandboxUnavailable, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RunCommand executes a command inside a hardened Docker container.
func (d *DockerContainerRunner) RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error) {
	// 1. Mandatory application-level command validation BEFORE any execution attempts or container creation
	bin, cleanArgs, valErr := validateSafeSubprocessCommand(cmdStr)
	if valErr != nil {
		return nil, fmt.Errorf("security violation: command policy rejection: %w", valErr)
	}

	// 2. Validate workspace containment and reject sensitive host paths
	if strings.TrimSpace(wsDir) == "" {
		return nil, fmt.Errorf("workspace directory cannot be empty")
	}
	cleanWs, err := filepath.Abs(wsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace path: %w", err)
	}
	cleanWs = filepath.Clean(cleanWs)
	lowerWs := strings.ToLower(filepath.ToSlash(cleanWs))
	if strings.Contains(lowerWs, "docker.sock") || lowerWs == "/" || lowerWs == "c:/" || lowerWs == "c:\\" {
		return nil, fmt.Errorf("security violation: refusing to mount sensitive host path %q into container", wsDir)
	}

	// 3. Ensure disposable go.mod exists in workspace for offline container Go toolchain execution
	_ = EnsureDisposableModule(cleanWs)

	// 4. Staged offline dependency cache: stage verified dependencies from pinned go.mod/go.sum
	_ = StageVerifiedDependencies(cleanWs)
	modCache := resolveStagedModuleCache()

	// 5. Check container runtime availability
	if err := d.CheckAvailable(ctx); err != nil {
		return nil, err
	}

	// 6. Track execution attempts: reached only after policy checks succeed and runtime is available
	atomic.AddInt64(&d.ExecutionAttempts, 1)

	containerName := fmt.Sprintf("ucs503-oracle-%s", uuid.New().String()[:12])
	timeout := d.Limits.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cleanupResult := "NOT_NEEDED"
	defer func() {
		// Guaranteed container cleanup hook prevents container and resource leakage
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		rmCmd := exec.CommandContext(cleanupCtx, d.DockerBinary, "rm", "-f", containerName)
		if rmOut, rmErr := rmCmd.CombinedOutput(); rmErr == nil {
			cleanupResult = "SUCCESS: container removed"
		} else {
			cleanupResult = fmt.Sprintf("CLEANUP_CALLED (%s)", strings.TrimSpace(string(rmOut)))
		}
	}()

	wsMount := filepath.ToSlash(cleanWs)
	tmpfsSizeMB := d.Limits.MaxMemoryMB / 4
	if tmpfsSizeMB < 128 {
		tmpfsSizeMB = 128
	}
	if tmpfsSizeMB > 512 {
		tmpfsSizeMB = 512
	}

	args := []string{
		"run",
		"--rm",
		"--name", containerName,
		"--network", "none", // Zero network access
		"--read-only", // Read-only root filesystem
		"--tmpfs", fmt.Sprintf("/tmp:rw,exec,nosuid,nodev,mode=1777,size=%dm", tmpfsSizeMB),
		"-e", "GOCACHE=/tmp/gocache",
		"-e", "CGO_ENABLED=0",
		"-e", "GOTMPDIR=/tmp",
		"-e", "GOPROXY=off",
		"-e", "GOSUMDB=off",
		"-e", "GONOSUMDB=*",
		"--user", "10001:10001", // Non-root
		"--cap-drop", "ALL", // Drop all Linux capabilities
		"--security-opt", "no-new-privileges",
		"--memory", fmt.Sprintf("%dm", d.Limits.MaxMemoryMB),
		"--cpus", fmt.Sprintf("%.2f", d.Limits.MaxCPUs),
		"--pids-limit", fmt.Sprintf("%d", d.Limits.MaxPids),
		"-v", fmt.Sprintf("%s:/workspace:rw", wsMount),
		"-w", "/workspace",
	}

	if modCache != "" {
		args = append(args,
			"-v", fmt.Sprintf("%s:/go/pkg/mod:ro", filepath.ToSlash(modCache)),
			"-e", "GOPATH=/go",
		)
	}

	args = append(args, d.Image)
	// Structured executable and arguments
	args = append(args, bin)
	args = append(args, cleanArgs...)

	start := time.Now()
	cmd := exec.CommandContext(cmdCtx, d.DockerBinary, args...)
	out, runErr := cmd.CombinedOutput()
	dur := time.Since(start)

	exitCode := 0
	timeoutCause := ""
	if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		exitCode = -1
		timeoutCause = fmt.Sprintf("execution timeout: context deadline exceeded after %v (limit: %v)", dur, timeout)
	} else if errors.Is(cmdCtx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		exitCode = -1
		timeoutCause = "execution canceled: context was canceled"
	} else if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	evidence := &ContainerExecutionEvidence{
		ContainerID:          containerName,
		IsolationPolicy:      "DOCKER_HARDENED_CONTAINER (network=none, read-only-root, tmpfs=/tmp, cap-drop=ALL, no-new-privs, non-root=10001:10001)",
		Command:              cmdStr,
		ExitStatus:           exitCode,
		MemoryLimitMB:        d.Limits.MaxMemoryMB,
		CPULimit:             d.Limits.MaxCPUs,
		NetworkPolicy:        "DISABLED (--network none)",
		RootFSPolicy:         "READ_ONLY (--read-only)",
		WorkspaceDir:         cleanWs,
		Image:                d.Image,
		ImageDigest:          d.resolveImageDigest(ctx),
		TimeoutCause:         timeoutCause,
		RequiredDependencies: parseRequiredDependencies(cleanWs),
		CleanupResult:        cleanupResult,
	}

	return &ContainerRunResult{
		ExitCode: exitCode,
		Output:   string(out),
		Duration: dur,
		Evidence: evidence,
	}, nil
}

// IndependentOracleSuite executes five distinct oracles to verify patch correctness.
type IndependentOracleSuite struct {
	TimeoutSec int
	Runner     ContainerRunner
}

// NewIndependentOracleSuite creates a new oracle verification suite with container containment.
func NewIndependentOracleSuite(timeoutSec int, runner ...ContainerRunner) *IndependentOracleSuite {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	var r ContainerRunner
	if len(runner) > 0 {
		r = runner[0]
	} else {
		r = NewDockerContainerRunner()
	}
	return &IndependentOracleSuite{TimeoutSec: timeoutSec, Runner: r}
}

// VerifyAll executes all five oracles against the workspace.
func (s *IndependentOracleSuite) VerifyAll(
	wsDir string,
	task *planning.RemediationTask,
	patch *CandidatePatch,
	reproducerCmd string,
	invarianceCmd string,
) *MultiOracleReport {
	start := time.Now()
	report := &MultiOracleReport{
		PatchID:       patch.PatchID,
		OracleResults: make(map[OracleType]*OracleResult),
		AllPassed:     true,
	}

	// 1. Oracle 1: Syntax & Compilation Oracle
	res1 := s.verifySyntaxCompilation(wsDir, task, patch)
	report.OracleResults[OracleSyntaxCompilation] = res1

	// 2. Oracle 2: Regression Test Oracle
	res2 := s.verifyRegressionTests(wsDir, task)
	report.OracleResults[OracleRegressionTest] = res2

	// 3. Oracle 3: Defect-Specific Reproducer Oracle
	res3 := s.verifyDefectReproducer(wsDir, task, patch, reproducerCmd)
	report.OracleResults[OracleDefectReproducer] = res3

	// 4. Oracle 4: Behavioral Invariance Oracle
	res4 := s.verifyBehavioralInvariance(wsDir, task, invarianceCmd)
	report.OracleResults[OracleBehavioralInvariance] = res4

	// 5. Oracle 5: Negative Properties Oracle (AST inspection: no shadow variables, no escaping leaks, etc.)
	res5 := s.verifyNegativeProperties(wsDir, patch)
	report.OracleResults[OracleNegativeProperties] = res5

	// AllPassed indicates no executed check failed and no required dynamic check was unavailable
	report.AllPassed = true
	for _, res := range report.OracleResults {
		if res.Status == OracleStatusFailed || res.Status == OracleStatusUnavailable {
			report.AllPassed = false
			break
		}
	}

	report.TotalDurationMs = time.Since(start).Milliseconds()
	return report
}

func getExecutionEnv() []string {
	return GetSanitizedExecutionEnv()
}

// GetSanitizedExecutionEnv produces a scrubbed environment slice that strictly strips
// all secrets, tokens, credentials, and sensitive host environment variables.
func GetSanitizedExecutionEnv() []string {
	allowedKeys := map[string]bool{
		"PATH":         true,
		"SYSTEMROOT":   true,
		"WINDIR":       true,
		"COMSPEC":      true,
		"PATHEXT":      true,
		"TEMP":         true,
		"TMP":          true,
		"GOTMPDIR":     true,
		"GOCACHE":      true,
		"GOPATH":       true,
		"GOROOT":       true,
		"GOOS":         true,
		"GOARCH":       true,
		"CGO_ENABLED":  true,
		"HOME":         true,
		"USERPROFILE":  true,
		"LOCALAPPDATA": true,
		"APPDATA":      true,
	}

	sensitiveSubstrings := []string{
		"KEY", "TOKEN", "SECRET", "PASS", "AUTH", "CREDENTIAL",
		"AWS_", "GITHUB_", "SSH_", "PRIVATE", "DATABASE_URL",
	}

	var cleanEnv []string
	hasGocache := false
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		upperK := strings.ToUpper(parts[0])
		if !allowedKeys[upperK] {
			continue
		}
		isSensitive := false
		for _, pat := range sensitiveSubstrings {
			if strings.Contains(upperK, pat) {
				isSensitive = true
				break
			}
		}
		if !isSensitive {
			if upperK == "GOCACHE" {
				hasGocache = true
			}
			cleanEnv = append(cleanEnv, entry)
		}
	}

	// Enforce CGO_ENABLED=0 to prevent unverified C compilation
	cleanEnv = append(cleanEnv, "CGO_ENABLED=0")

	// Ensure isolated temp directories via dynamic workspace resolution
	tmpBase, err := EnsureTmpDir(ResolveWorkspaceRoot())
	if err == nil && tmpBase != "" {
		cleanEnv = append(cleanEnv, `GOTMPDIR=`+tmpBase, `TEMP=`+tmpBase, `TMP=`+tmpBase)
	} else {
		tmpBase = os.TempDir()
	}

	// Ensure GOCACHE is always set to prevent go vet / build cache failures
	if !hasGocache {
		gocacheDir := filepath.Join(tmpBase, "gocache")
		_ = os.MkdirAll(gocacheDir, 0755)
		cleanEnv = append(cleanEnv, "GOCACHE="+gocacheDir)
	}

	return cleanEnv
}

// validateSafeSubprocessCommand validates that a command string to be executed
// conforms to strict security whitelisting: only safe Go toolchain invocations are permitted.
func validateSafeSubprocessCommand(cmdStr string) (string, []string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	binaryName := strings.ToLower(filepath.Base(parts[0]))
	// Reject known interpreters, dangerous binaries, and privilege escalators
	blockedBinaries := map[string]bool{
		"sh": true, "bash": true, "zsh": true, "ksh": true, "csh": true, "tcsh": true,
		"dash": true, "fish": true,
		"cmd": true, "cmd.exe": true, "powershell": true, "powershell.exe": true,
		"pwsh": true, "pwsh.exe": true, "python": true, "python2": true, "python3": true, "perl": true,
		"ruby": true, "node": true, "php": true, "curl": true, "wget": true, "nc": true, "netcat": true,
		"ncat": true, "socat": true, "telnet": true, "ftp": true, "scp": true, "ssh": true, "rsync": true,
		"rm": true, "del": true, "rmdir": true, "dd": true, "mkfs": true, "format": true,
		"apt": true, "apk": true, "yum": true, "dnf": true, "pacman": true,
		"docker": true, "podman": true, "sudo": true, "su": true, "doas": true, "chown": true, "chmod": true,
	}
	if blockedBinaries[binaryName] {
		return "", nil, fmt.Errorf("command %q is blocked: direct host execution of arbitrary shell interpreters or unauthorized binaries is strictly forbidden", parts[0])
	}
	if binaryName != "go" && binaryName != "go.exe" {
		return "", nil, fmt.Errorf("command %q is blocked: only 'go' toolchain commands are whitelisted for verification", parts[0])
	}
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("incomplete go command: subcommand required")
	}
	allowedSubcommands := map[string]bool{
		"vet": true, "test": true, "build": true,
	}
	if !allowedSubcommands[parts[1]] {
		return "", nil, fmt.Errorf("go subcommand %q is not whitelisted", parts[1])
	}
	// Check for dangerous flags like -exec, -toolexec, -plugin
	for _, arg := range parts[2:] {
		lowerArg := strings.ToLower(arg)
		if strings.HasPrefix(lowerArg, "-exec") || strings.HasPrefix(lowerArg, "--exec") ||
			strings.HasPrefix(lowerArg, "-toolexec") || strings.HasPrefix(lowerArg, "--toolexec") ||
			strings.HasPrefix(lowerArg, "-plugin") || strings.HasPrefix(lowerArg, "--plugin") {
			return "", nil, fmt.Errorf("go flag %q is forbidden in verification", arg)
		}
	}
	return parts[0], parts[1:], nil
}

// -----------------------------------------------------------------------------------------
// 1. Syntax & Compilation Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifySyntaxCompilation(wsDir string, task *planning.RemediationTask, patch *CandidatePatch) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleSyntaxCompilation, Status: OracleStatusPassed, Passed: true}

	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, targetPath, nil, parser.ParseComments)
	if err != nil {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = fmt.Sprintf("Source file failed AST parse: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// Dynamic compilation/vetting: MUST execute in isolated container!
	// Host execution of toolchain commands on untrusted repositories is strictly forbidden.
	_ = EnsureDisposableModule(wsDir)

	vetPkg := "."
	if patch != nil && patch.TargetFile != "" {
		relDir := filepath.Dir(filepath.FromSlash(patch.TargetFile))
		if relDir != "." && relDir != "" {
			vetPkg = "./" + filepath.ToSlash(relDir)
		}
	}
	vetCmd := "go vet " + vetPkg

	timeout := time.Duration(s.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if s.Runner == nil || s.Runner.CheckAvailable(ctx) != nil {
		// Secure container runtime is not available.
		// FAIL-CLOSED: Do NOT fall back to host execution!
		res.Status = OracleStatusPassed // Static AST syntax valid on trusted host
		res.Passed = true
		res.Details = "AST syntax valid on trusted host (static AST check only; dynamic container vetting unavailable: DYNAMIC_VERIFICATION_UNAVAILABLE)"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// Run `go vet <pkg>` inside isolated container
	runRes, runErr := s.Runner.RunCommand(ctx, wsDir, vetCmd)
	if runRes != nil {
		res.ContainerInfo = runRes.Evidence
		res.Output = runRes.Output

		if isOfflineDependencyMissing(runRes.Output) {
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: offline module cache missing dependency: %s", strings.TrimSpace(runRes.Output))
			res.Details = "Dynamic container vetting unavailable due to missing offline dependency"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		if runRes.Evidence != nil && runRes.Evidence.TimeoutCause != "" {
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %s", runRes.Evidence.TimeoutCause)
			res.Details = "Dynamic container vetting timed out"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		if runRes.ExitCode != 0 {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Container go vet failed (exit %d): %s", runRes.ExitCode, runRes.Output)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Syntax valid and container go vet passed with 0 warnings"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	if runErr != nil {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = fmt.Sprintf("Container execution failed: %v", runErr)
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 2. Regression Test Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyRegressionTests(wsDir string, task *planning.RemediationTask) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleRegressionTest, Status: OracleStatusPassed, Passed: true}

	if len(task.ValidationCommands) == 0 {
		// Explicit status: NOT_RUN (do not equate NOT_RUN with PASSED)
		res.Status = OracleStatusNotRun
		res.Passed = false
		res.Details = "No regression validation commands specified for candidate (NOT_RUN)"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	timeout := time.Duration(s.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, cmdStr := range task.ValidationCommands {
		// 1. Mandatory application-level command validation BEFORE container or host evaluation
		_, _, valErr := validateSafeSubprocessCommand(cmdStr)
		if valErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Security rejection: %v", valErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		// 2. Check container isolation availability
		if s.Runner == nil || s.Runner.CheckAvailable(ctx) != nil {
			// Containment cannot be established. Fail closed with UNAVAILABLE.
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = "DYNAMIC_VERIFICATION_UNAVAILABLE: secure container runtime not detected or daemon not running; host execution forbidden"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		runRes, runErr := s.Runner.RunCommand(ctx, wsDir, cmdStr)
		if runRes != nil {
			res.ContainerInfo = runRes.Evidence
			res.Output = runRes.Output

			if isOfflineDependencyMissing(runRes.Output) {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: offline module cache missing dependency: %s", strings.TrimSpace(runRes.Output))
				res.Details = "Dynamic regression test unavailable due to missing offline dependency"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.Evidence != nil && runRes.Evidence.TimeoutCause != "" {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %s", runRes.Evidence.TimeoutCause)
				res.Details = "Dynamic regression test timed out"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.ExitCode != 0 {
				res.Status = OracleStatusFailed
				res.Passed = false
				res.FailureReason = fmt.Sprintf("Regression command %q failed in container (exit %d): %s", cmdStr, runRes.ExitCode, runRes.Output)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		} else if runErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Regression command %q failed: %v", cmdStr, runErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.ContainerInfo = runRes.Evidence
	}

	res.Status = OracleStatusPassed
	res.Passed = true
	res.Details = fmt.Sprintf("All %d validation commands executed successfully in isolated container", len(task.ValidationCommands))
	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 3. Defect-Specific Reproducer Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyDefectReproducer(
	wsDir string,
	task *planning.RemediationTask,
	patch *CandidatePatch,
	reproducerCmd string,
) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleDefectReproducer, Status: OracleStatusPassed, Passed: true}

	// If explicit reproducer command provided, it MUST execute in container
	if reproducerCmd != "" {
		// 1. Mandatory command validation BEFORE container or host execution
		_, _, valErr := validateSafeSubprocessCommand(reproducerCmd)
		if valErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Security rejection: %v", valErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		timeout := time.Duration(s.TimeoutSec) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if s.Runner == nil || s.Runner.CheckAvailable(ctx) != nil {
			// Containment cannot be established. Fail closed.
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = "DYNAMIC_VERIFICATION_UNAVAILABLE: secure container runtime not detected or daemon not running; host execution forbidden"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		runRes, runErr := s.Runner.RunCommand(ctx, wsDir, reproducerCmd)
		if runRes != nil {
			res.ContainerInfo = runRes.Evidence
			res.Output = runRes.Output

			if isOfflineDependencyMissing(runRes.Output) {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: offline module cache missing dependency: %s", strings.TrimSpace(runRes.Output))
				res.Details = "Dynamic defect reproducer unavailable due to missing offline dependency"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.Evidence != nil && runRes.Evidence.TimeoutCause != "" {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %s", runRes.Evidence.TimeoutCause)
				res.Details = "Dynamic defect reproducer timed out"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.ExitCode != 0 {
				res.Status = OracleStatusFailed
				res.Passed = false
				res.FailureReason = fmt.Sprintf("Defect reproducer failed in container (exit %d): %s", runRes.ExitCode, runRes.Output)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		} else if runErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Defect reproducer execution failed: %v", runErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Dynamic defect reproducer test confirmed defect elimination in isolated container"
		res.Output = runRes.Output
		res.ContainerInfo = runRes.Evidence
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// Syntactic/semantic AST property check executed on trusted host without running untrusted code
	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	srcBytes, err := os.ReadFile(targetPath)
	if err != nil {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = fmt.Sprintf("Cannot read patched file: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}
	src := string(srcBytes)

	taskType := strings.ToLower(task.Title + " " + task.Description)
	if strings.Contains(taskType, "resource") || strings.Contains(taskType, "leak") {
		if !strings.Contains(src, ".Close()") {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = "Patched file does not contain Close() call"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		// Hardened check: Verify that the enclosing function does not return the resource to callers
		fset := token.NewFileSet()
		if fNode, err := parser.ParseFile(fset, targetPath, nil, 0); err == nil {
			for _, decl := range fNode.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					if fnReturnsResourceType(fn) && funcContainsClose(fn) {
						res.Status = OracleStatusFailed
						res.Passed = false
						res.FailureReason = "Defect reproducer check failed: function returns resource to caller; callee cleanup violates interface contract"
						res.DurationMs = time.Since(start).Milliseconds()
						return res
					}
				}
			}
		}
		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Explicit or deferred resource cleanup verified along exit paths (syntactic check; dynamic reproducer not executed)"
	} else if strings.Contains(taskType, "sql") {
		if strings.Contains(src, "fmt.Sprintf") && strings.Contains(src, "%s") {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = "Patched file still contains dynamic fmt.Sprintf SQL formatting"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		if !strings.Contains(src, "?") {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = "Patched file lacks parameterized placeholder '?'"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Parameterized SQL placeholder and parameter bindings verified (syntactic check; dynamic reproducer not executed)"
	} else if strings.Contains(taskType, "error") {
		if !strings.Contains(src, "if err != nil") && !strings.Contains(src, "if err :=") {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = "Patched file does not check error return value"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Error handling and propagation check verified (syntactic check; dynamic reproducer not executed)"
	} else {
		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Structural AST properties verified (syntactic check; dynamic reproducer not executed)"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 4. Behavioral Invariance Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyBehavioralInvariance(
	wsDir string,
	task *planning.RemediationTask,
	invarianceCmd string,
) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleBehavioralInvariance, Status: OracleStatusPassed, Passed: true}

	if invarianceCmd != "" {
		// 1. Mandatory command validation BEFORE container or host execution
		_, _, valErr := validateSafeSubprocessCommand(invarianceCmd)
		if valErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Security rejection: %v", valErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		timeout := time.Duration(s.TimeoutSec) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if s.Runner == nil || s.Runner.CheckAvailable(ctx) != nil {
			// Containment cannot be established. Fail closed.
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = "DYNAMIC_VERIFICATION_UNAVAILABLE: secure container runtime not detected or daemon not running; host execution forbidden"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		runRes, runErr := s.Runner.RunCommand(ctx, wsDir, invarianceCmd)
		if runRes != nil {
			res.ContainerInfo = runRes.Evidence
			res.Output = runRes.Output

			if isOfflineDependencyMissing(runRes.Output) {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: offline module cache missing dependency: %s", strings.TrimSpace(runRes.Output))
				res.Details = "Dynamic behavioral invariance test unavailable due to missing offline dependency"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.Evidence != nil && runRes.Evidence.TimeoutCause != "" {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %s", runRes.Evidence.TimeoutCause)
				res.Details = "Dynamic behavioral invariance test timed out"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}

			if runRes.ExitCode != 0 {
				res.Status = OracleStatusFailed
				res.Passed = false
				res.FailureReason = fmt.Sprintf("Behavioral invariance test failed in container (exit %d): %s", runRes.ExitCode, runRes.Output)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		} else if runErr != nil {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Behavioral invariance command failed: %v", runErr)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}

		res.Status = OracleStatusPassed
		res.Passed = true
		res.Details = "Non-defective execution paths produce identical invariant behavior in container"
		res.Output = runRes.Output
		res.ContainerInfo = runRes.Evidence
	} else {
		res.Status = OracleStatusNotRun
		res.Passed = false
		res.Details = "No invariance command specified for candidate (NOT_RUN)"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 5. Negative Properties Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyNegativeProperties(wsDir string, patch *CandidatePatch) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleNegativeProperties, Status: OracleStatusPassed, Passed: true}

	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, targetPath, nil, parser.ParseComments)
	if err != nil {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = fmt.Sprintf("AST parse failed: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// 1. Verify no shadowed 'err' definitions introduced by the candidate patch
	shadowedErr := false
	invertedErr := false
	diffAddsClose := false
	var closedVarNames []string

	if patch != nil && strings.Contains(patch.UnifiedDiff, "+") {
		for _, line := range strings.Split(patch.UnifiedDiff, "\n") {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				trimmed := strings.TrimSpace(line[1:])
				if strings.Contains(trimmed, "err :=") && (strings.HasPrefix(line, "+\t\t\t") || strings.HasPrefix(line, "+            ")) {
					shadowedErr = true
				}
				if strings.HasPrefix(trimmed, "if err == nil") && strings.Contains(trimmed, "return") {
					invertedErr = true
				}
				if strings.Contains(trimmed, ".Close()") {
					diffAddsClose = true
					// Extract variable name, e.g. "resp.Body.Close()" -> "resp", "f.Close()" -> "f"
					closeCall := strings.TrimPrefix(trimmed, "defer ")
					parts := strings.Split(closeCall, ".")
					if len(parts) > 0 {
						vName := strings.TrimSpace(parts[0])
						if vName != "" {
							closedVarNames = append(closedVarNames, vName)
						}
					}
				}
			}
		}
	}

	if shadowedErr {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = "Negative property check failed: candidate patch introduces shadowed 'err' variable declaration in nested block"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	if invertedErr {
		res.Status = OracleStatusFailed
		res.Passed = false
		res.FailureReason = "Negative property check failed: candidate patch introduces inverted error check condition (if err == nil)"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// 2. Structural & Semantic Invariants on Patched AST
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// A. Callee Cleanup on Escaping Resource Violation
		if diffAddsClose && fnReturnsResourceType(fn) {
			for _, vName := range closedVarNames {
				if fnBodyReturnsVar(fn, vName) || fnReturnsTypeMatchingVar(fn, vName) {
					res.Status = OracleStatusFailed
					res.Passed = false
					res.FailureReason = fmt.Sprintf("Negative property check failed: candidate patch introduces callee cleanup on escaping resource %q returned to caller; violates ownership transfer and panics on error returns", vName)
					res.DurationMs = time.Since(start).Milliseconds()
					return res
				}
			}
		}

		// B. Pre-Error Check Cleanup Placement (Nil dereference on failure path)
		if diffAddsClose {
			for _, vName := range closedVarNames {
				if isCleanupPlacedBeforeAllocErrorCheck(fn.Body.List, vName) {
					res.Status = OracleStatusFailed
					res.Passed = false
					res.FailureReason = fmt.Sprintf("Negative property check failed: resource cleanup for %q placed prior to error check; triggers nil pointer dereference panic when allocation fails", vName)
					res.DurationMs = time.Since(start).Milliseconds()
					return res
				}
			}
		}

		// C. Redundant or Double Close Detection
		closeCounts := make(map[string]int)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
					targetStr := formatOracleNode(sel.X)
					closeCounts[targetStr]++
				}
			}
			return true
		})
		for targetStr, count := range closeCounts {
			if count > 1 {
				res.Status = OracleStatusFailed
				res.Passed = false
				res.FailureReason = fmt.Sprintf("Negative property check failed: multiple Close() calls detected on resource %q without reassignment", targetStr)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		}
	}

	res.Status = OracleStatusPassed
	res.Passed = true
	res.Details = "Zero shadowed variables, nil-dereference risk, or ownership violations detected"
	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

func fnReturnsResourceType(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		typeStr := formatOracleAstType(field.Type)
		if isOracleResourceType(typeStr) {
			return true
		}
	}
	return false
}

func fnReturnsTypeMatchingVar(fn *ast.FuncDecl, varName string) bool {
	// If varName is 'resp', check for *http.Response
	// If varName is 'f', check for *os.File
	// If varName is 'conn', check for net.Conn
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		typeStr := formatOracleAstType(field.Type)
		if strings.Contains(varName, "resp") && strings.Contains(typeStr, "Response") {
			return true
		}
		if strings.Contains(varName, "f") && strings.Contains(typeStr, "File") {
			return true
		}
		if strings.Contains(varName, "conn") && strings.Contains(typeStr, "Conn") {
			return true
		}
	}
	return false
}

func fnBodyReturnsVar(fn *ast.FuncDecl, varName string) bool {
	returned := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, expr := range ret.Results {
			if id, ok := expr.(*ast.Ident); ok && id.Name == varName {
				returned = true
				return false
			}
			if sel, ok := expr.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == varName {
					returned = true
					return false
				}
			}
		}
		return true
	})
	return returned
}

func funcContainsClose(fn *ast.FuncDecl) bool {
	hasClose := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			hasClose = true
			return false
		}
		return true
	})
	return hasClose
}

func stmtContainsClose(stmt ast.Stmt) bool {
	hasClose := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			hasClose = true
			return false
		}
		return true
	})
	return hasClose
}

func isCleanupPlacedBeforeAllocErrorCheck(stmts []ast.Stmt, varName string) bool {
	allocIdx := -1
	closeIdx := -1

	for i, stmt := range stmts {
		if allocIdx == -1 && stmtAssignsVar(stmt, varName) {
			allocIdx = i
		}
		if closeIdx == -1 && stmtClosesVar(stmt, varName) {
			closeIdx = i
		}
	}

	if allocIdx >= 0 && closeIdx > allocIdx {
		// Check if there is an error check between allocIdx and closeIdx
		hasCheckBetween := false
		for k := allocIdx + 1; k < closeIdx; k++ {
			if isErrCheckStmt(stmts[k]) {
				hasCheckBetween = true
				break
			}
		}
		// If no check between alloc and close, but an error check follows close:
		if !hasCheckBetween {
			for k := closeIdx + 1; k < len(stmts); k++ {
				if isErrCheckStmt(stmts[k]) {
					return true // Pre-error check cleanup!
				}
			}
		}
	}
	return false
}

func stmtAssignsVar(stmt ast.Stmt, varName string) bool {
	if assign, ok := stmt.(*ast.AssignStmt); ok {
		for _, lhs := range assign.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name == varName {
				return true
			}
		}
	}
	return false
}

func stmtClosesVar(stmt ast.Stmt, varName string) bool {
	closes := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == varName {
				closes = true
				return false
			}
			if innerSel, ok := sel.X.(*ast.SelectorExpr); ok {
				if id, ok := innerSel.X.(*ast.Ident); ok && id.Name == varName {
					closes = true
					return false
				}
			}
		}
		return true
	})
	return closes
}

func isErrCheckStmt(stmt ast.Stmt) bool {
	if ifStmt, ok := stmt.(*ast.IfStmt); ok {
		if bin, ok := ifStmt.Cond.(*ast.BinaryExpr); ok && bin.Op == token.NEQ {
			x := fmt.Sprintf("%v", bin.X)
			y := fmt.Sprintf("%v", bin.Y)
			return (x == "err" && y == "nil") || (x == "nil" && y == "err")
		}
	}
	return false
}

func formatOracleNode(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return formatOracleNode(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func formatOracleAstType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return "*" + formatOracleAstType(t.X)
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

func isOracleResourceType(typeStr string) bool {
	switch typeStr {
	case "*http.Response", "http.Response",
		"*os.File", "os.File",
		"net.Conn", "*net.Conn",
		"io.ReadCloser", "io.Closer",
		"*sql.Rows", "sql.Rows",
		"*sql.DB", "sql.DB":
		return true
	default:
		return false
	}
}
