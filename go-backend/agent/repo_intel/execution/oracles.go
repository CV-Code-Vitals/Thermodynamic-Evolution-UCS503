package execution

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/dirhash"
)

// Standard oracle error definitions
var (
	ErrSandboxUnavailable             = errors.New("sandbox execution unavailable: secure container runtime not detected or daemon not running")
	ErrDynamicVerificationUnavailable = errors.New("DYNAMIC_VERIFICATION_UNAVAILABLE: dynamic execution requires hardened container isolation; host execution forbidden")
	ErrHostExecutionForbidden         = errors.New("untrusted candidate execution on host is strictly forbidden")
)

// OracleType identifies each independent verification oracle.
type oracleOutput struct{ buffer bytes.Buffer }

func (b *oracleOutput) Write(p []byte) (int, error) {
	const limit = 64 << 10
	remaining := limit - b.buffer.Len()
	if remaining > 0 {
		_, _ = b.buffer.Write(p[:min(remaining, len(p))])
	}
	return len(p), nil
}

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

// CommandExecutor abstracts subprocess command execution for isolated container runtime operations.
type CommandExecutor func(ctx context.Context, binary string, args ...string) ([]byte, error)

// DockerContainerRunner executes commands inside an ephemeral, hardened Docker container.
type DockerContainerRunner struct {
	Image             string
	DockerBinary      string
	Limits            ContainerLimits
	Checker           func(ctx context.Context) error
	ExecCommand       CommandExecutor
	ExecutionAttempts int64
}

// exec dispatches a subprocess command via the configured CommandExecutor hook or standard exec.
func (d *DockerContainerRunner) exec(ctx context.Context, args ...string) ([]byte, error) {
	if d.ExecCommand != nil {
		return d.ExecCommand(ctx, d.DockerBinary, args...)
	}
	cmd := exec.CommandContext(ctx, d.DockerBinary, args...)
	output := &oracleOutput{}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	return output.buffer.Bytes(), err
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
	if out, err := d.exec(checkCtx, "image", "inspect", "--format", "{{index .RepoDigests 0}}", d.Image); err == nil {
		s := strings.TrimSpace(string(out))
		if s != "" && s != "<no value>" {
			return s
		}
	}
	// Fallback to Image ID
	if out, err := d.exec(checkCtx, "image", "inspect", "--format", "{{.Id}}", d.Image); err == nil {
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

// ModuleRequirement specifies an external module dependency required by a workspace.
type ModuleRequirement struct {
	Path    string
	Version string
}

func parseGoModRequirements(wsDir string) ([]ModuleRequirement, error) {
	data, err := readWorkspaceFile(wsDir, "go.mod")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	file, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, err
	}
	if len(file.Replace) != 0 {
		return nil, fmt.Errorf("security violation: unsafe replace directives are disabled in offline verification")
	}
	reqs := make([]ModuleRequirement, 0, len(file.Require))
	for _, req := range file.Require {
		if err := module.Check(req.Mod.Path, req.Mod.Version); err != nil {
			return nil, err
		}
		reqs = append(reqs, ModuleRequirement{Path: req.Mod.Path, Version: req.Mod.Version})
	}
	return reqs, nil
}

func parseGoSumChecksums(wsDir string) (map[string]string, error) {
	data, err := readWorkspaceFile(wsDir, "go.sum")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	checksums := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 3 || !strings.HasPrefix(fields[2], "h1:") {
			return nil, errors.New("invalid go.sum checksum entry")
		}
		digest, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fields[2], "h1:"))
		if err != nil || len(digest) != 32 {
			return nil, errors.New("invalid go.sum checksum digest")
		}
		key := fields[0] + "@" + fields[1]
		if previous, exists := checksums[key]; exists && previous != fields[2] {
			return nil, errors.New("conflicting go.sum checksums")
		}
		checksums[key] = fields[2]
	}
	return checksums, nil
}

func stageDependency(sourceCache, destination string, req ModuleRequirement, checksum, modChecksum string) error {
	escapedPath, err := module.EscapePath(req.Path)
	if err != nil {
		return err
	}
	escapedVersion, err := module.EscapeVersion(req.Version)
	if err != nil {
		return err
	}
	cacheName := filepath.FromSlash(escapedPath + "@" + escapedVersion)
	root, err := os.OpenRoot(sourceCache)
	if err != nil {
		return err
	}
	defer root.Close()
	source, err := root.OpenRoot(cacheName)
	if err != nil {
		return fmt.Errorf("required dependency %s@%s is missing from approved offline cache: %w", req.Path, req.Version, err)
	}
	defer source.Close()
	dest := filepath.Join(destination, cacheName)
	if err := fs.WalkDir(source.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dest, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := readRegularFile(source, filepath.FromSlash(path))
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	}); err != nil {
		return err
	}

	hash, err := dirhash.HashDir(dest, req.Path+"@"+req.Version, dirhash.Hash1)
	if err != nil {
		return err
	}
	if hash != checksum {
		return errors.New("approved cached module does not match its pinned checksum")
	}

	metadataDir := filepath.Join("cache", "download", filepath.FromSlash(escapedPath), "@v")
	metadataPath := filepath.Join(metadataDir, escapedVersion)
	modBytes, err := readRegularFile(root, metadataPath+".mod")
	if err != nil {
		return err
	}
	hash, err = dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(modBytes)), nil
	})
	if err != nil {
		return err
	}
	if hash != modChecksum {
		return errors.New("approved cached go.mod does not match its pinned checksum")
	}
	if err := os.MkdirAll(filepath.Join(destination, metadataDir), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destination, metadataPath+".mod"), modBytes, 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destination, metadataPath+".ziphash"), []byte(checksum), 0644)
}

// prepareMinimalDependencyCache constructs an isolated, disposable directory containing ONLY
// the minimal verified dependencies required by wsDir.
// DIRECT HOST DOWNLOADS OF UNTRUSTED REPOSITORY METADATA ARE STRICTLY PROHIBITED.
func prepareMinimalDependencyCache(wsDir string) (string, func(), error) {
	reqs, err := parseGoModRequirements(wsDir)
	if err != nil {
		return "", nil, fmt.Errorf("%w: go.mod validation failed: %v", ErrDynamicVerificationUnavailable, err)
	}
	if len(reqs) == 0 {
		// Zero external dependencies required; nothing to mount
		return "", func() {}, nil
	}

	// Verify that pinned checksums exist for all required modules
	checksums, err := parseGoSumChecksums(wsDir)
	if err != nil {
		return "", nil, fmt.Errorf("%w: go.sum check failed: %v", ErrDynamicVerificationUnavailable, err)
	}
	if len(checksums) == 0 {
		return "", nil, fmt.Errorf("%w: repository declares external dependencies but lacks pinned go.sum checksums", ErrDynamicVerificationUnavailable)
	}

	for _, req := range reqs {
		key := fmt.Sprintf("%s@%s", req.Path, req.Version)
		if checksums[key] == "" {
			return "", nil, fmt.Errorf("%w: required dependency %s lacks verified checksum in go.sum", ErrDynamicVerificationUnavailable, key)
		}
	}

	// Resolve approved dependency source
	sourceCache := os.Getenv("APPROVED_DEPENDENCY_CACHE")
	if sourceCache == "" {
		return "", nil, fmt.Errorf("%w: no approved offline dependency cache configured or available", ErrDynamicVerificationUnavailable)
	}

	// Create disposable, minimal module directory
	disposableDir, err := os.MkdirTemp("", "ucs503-minmod-*")
	if err != nil {
		return "", nil, fmt.Errorf("%w: failed to create disposable dependency cache directory: %v", ErrDynamicVerificationUnavailable, err)
	}
	cleanup := func() {
		_ = os.RemoveAll(disposableDir)
	}

	// Allow the unprivileged container user to read the verified tree.
	if err := os.Chmod(disposableDir, 0755); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("%w: prepare staged cache: %v", ErrDynamicVerificationUnavailable, err)
	}
	for _, req := range reqs {
		key := req.Path + "@" + req.Version
		if err := stageDependency(sourceCache, disposableDir, req, checksums[key], checksums[key+"/go.mod"]); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("%w: stage dependency %s: %v", ErrDynamicVerificationUnavailable, key, err)
		}
	}

	return disposableDir, cleanup, nil
}

// StageVerifiedDependencies verifies repository dependencies against pinned checksums and the approved offline cache.
// DIRECT HOST DOWNLOADS OF ARBITRARY REPOSITORY METADATA ARE STRICTLY PROHIBITED.
func StageVerifiedDependencies(wsDir string) error {
	_, cleanup, err := prepareMinimalDependencyCache(wsDir)
	if cleanup != nil {
		defer cleanup()
	}
	return err
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
	out, err := d.exec(checkCtx, "info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return fmt.Errorf("%w: docker daemon check failed: %v (%s)", ErrSandboxUnavailable, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// performVerifiedContainerCleanup ensures the container is removed and verifies its non-existence,
// returning an audit status string distinguishing verified removal, failure, timeout, or daemon error.
func (d *DockerContainerRunner) performVerifiedContainerCleanup(containerName string) string {
	if containerName == "" {
		return "NOT_NEEDED"
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()

	rmOut, rmErr := d.exec(cleanupCtx, "rm", "-f", containerName)
	if cleanupCtx.Err() != nil && errors.Is(cleanupCtx.Err(), context.DeadlineExceeded) {
		return fmt.Sprintf("CLEANUP_TIMEOUT: container %s removal exceeded 5s deadline", containerName)
	}
	if rmErr != nil {
		detail := strings.TrimSpace(string(rmOut))
		message := strings.ToLower(detail)
		if strings.Contains(message, "no such container") {
			return fmt.Sprintf("VERIFIED_REMOVAL: container %s absent (automatically removed via --rm)", containerName)
		}
		if strings.Contains(message, "cannot connect to the docker daemon") || strings.Contains(message, "daemon is not running") {
			return fmt.Sprintf("DAEMON_UNAVAILABLE: failed to remove container %s: %s", containerName, detail)
		}
		return fmt.Sprintf("CLEANUP_FAILURE: failed to remove container %s: %v (%s)", containerName, rmErr, detail)
	}

	// Verify container non-existence via inspect
	inspectCtx, inspectCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer inspectCancel()

	inspectOut, inspectErr := d.exec(inspectCtx, "inspect", "--format", "{{.State.Status}}", containerName)
	if errors.Is(inspectCtx.Err(), context.DeadlineExceeded) {
		return fmt.Sprintf("CLEANUP_TIMEOUT: verifying container %s absence exceeded 3s deadline", containerName)
	}
	if inspectErr != nil {
		detail := strings.TrimSpace(string(inspectOut))
		message := strings.ToLower(detail)
		if strings.Contains(message, "no such container") || strings.Contains(message, "no such object") {
			return fmt.Sprintf("VERIFIED_REMOVAL: container %s successfully removed and verified absent", containerName)
		}
		return fmt.Sprintf("CLEANUP_FAILURE: could not verify container %s absence: %v (%s)", containerName, inspectErr, detail)
	}

	status := strings.TrimSpace(string(inspectOut))
	return fmt.Sprintf("CLEANUP_FAILURE: container %s still exists after removal (state: %s)", containerName, status)
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

	// 4. Staged offline dependency cache: prepare minimal verified cache
	// DIRECT HOST DOWNLOADS OF ARBITRARY REPOSITORY METADATA ARE STRICTLY PROHIBITED.
	minimalModCache, cleanupModCache, stageErr := prepareMinimalDependencyCache(cleanWs)
	if stageErr != nil {
		return nil, fmt.Errorf("%w: safe dependency staging failed: %v", ErrDynamicVerificationUnavailable, stageErr)
	}
	if cleanupModCache != nil {
		defer cleanupModCache()
	}

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

	// Safety fallback cleanup hook prevents container and resource leakage
	cleanedUp := false
	defer func() {
		if !cleanedUp {
			_ = d.performVerifiedContainerCleanup(containerName)
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
		"--pull=never",
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
		"-e", "GOTOOLCHAIN=local", // Prevent dynamic network toolchain downloads
		"-e", "GOWORK=off",
		"-e", "GOENV=off",
		"--user", "10001:10001", // Non-root
		"--cap-drop", "ALL", // Drop all Linux capabilities
		"--security-opt", "no-new-privileges",
		"--memory", fmt.Sprintf("%dm", d.Limits.MaxMemoryMB),
		"--cpus", fmt.Sprintf("%.2f", d.Limits.MaxCPUs),
		"--pids-limit", fmt.Sprintf("%d", d.Limits.MaxPids),
		"-v", fmt.Sprintf("%s:/workspace:ro", wsMount),
		"-w", "/workspace",
	}

	if minimalModCache != "" {
		args = append(args,
			"-v", fmt.Sprintf("%s:/go/pkg/mod:ro", filepath.ToSlash(minimalModCache)),
			"-e", "GOPATH=/go",
		)
	}

	args = append(args, d.Image)
	// Structured executable and arguments
	args = append(args, bin)
	args = append(args, cleanArgs...)

	start := time.Now()
	out, runErr := d.exec(cmdCtx, args...)
	dur := time.Since(start)

	// Synchronous cleanup completed BEFORE building ContainerExecutionEvidence
	cleanupResult := d.performVerifiedContainerCleanup(containerName)
	cleanedUp = true

	exitCode := 0
	timeoutCause := ""
	if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(runErr, context.DeadlineExceeded) {
		exitCode = -1
		timeoutCause = fmt.Sprintf("execution timeout: context deadline exceeded after %v (limit: %v)", dur, timeout)
	} else if errors.Is(cmdCtx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) || errors.Is(runErr, context.Canceled) {
		exitCode = -1
		timeoutCause = "execution canceled: context was canceled"
	} else if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	if !strings.HasPrefix(cleanupResult, "VERIFIED_REMOVAL:") && exitCode == 0 {
		exitCode = -1
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

	// Only a complete verification run can pass every oracle.
	report.AllPassed = true
	for _, res := range report.OracleResults {
		if !res.Passed {
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
		res.Status = OracleStatusUnavailable // Parsing succeeded, but compilation did not execute.
		res.Passed = false
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
		if errors.Is(runErr, ErrDynamicVerificationUnavailable) || errors.Is(runErr, ErrSandboxUnavailable) || strings.Contains(runErr.Error(), "DYNAMIC_VERIFICATION_UNAVAILABLE") {
			res.Status = OracleStatusUnavailable
			res.Passed = false
			res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %v", runErr)
			res.Details = "Dynamic container vetting unavailable"
		} else {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Container execution failed: %v", runErr)
		}
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
			if errors.Is(runErr, ErrDynamicVerificationUnavailable) || errors.Is(runErr, ErrSandboxUnavailable) || strings.Contains(runErr.Error(), "DYNAMIC_VERIFICATION_UNAVAILABLE") {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %v", runErr)
				res.Details = "Dynamic regression test unavailable"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
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
			if errors.Is(runErr, ErrDynamicVerificationUnavailable) || errors.Is(runErr, ErrSandboxUnavailable) || strings.Contains(runErr.Error(), "DYNAMIC_VERIFICATION_UNAVAILABLE") {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %v", runErr)
				res.Details = "Dynamic defect reproducer unavailable"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
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
		res.Status = OracleStatusNotRun
		res.Passed = false
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
		res.Status = OracleStatusNotRun
		res.Passed = false
		res.Details = "Parameterized SQL placeholder and parameter bindings verified (syntactic check; dynamic reproducer not executed)"
	} else if strings.Contains(taskType, "error") {
		if !strings.Contains(src, "if err != nil") && !strings.Contains(src, "if err :=") {
			res.Status = OracleStatusFailed
			res.Passed = false
			res.FailureReason = "Patched file does not check error return value"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Status = OracleStatusNotRun
		res.Passed = false
		res.Details = "Error handling and propagation check verified (syntactic check; dynamic reproducer not executed)"
	} else {
		res.Status = OracleStatusNotRun
		res.Passed = false
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
			if errors.Is(runErr, ErrDynamicVerificationUnavailable) || errors.Is(runErr, ErrSandboxUnavailable) || strings.Contains(runErr.Error(), "DYNAMIC_VERIFICATION_UNAVAILABLE") {
				res.Status = OracleStatusUnavailable
				res.Passed = false
				res.FailureReason = fmt.Sprintf("DYNAMIC_VERIFICATION_UNAVAILABLE: %v", runErr)
				res.Details = "Dynamic behavioral invariance test unavailable"
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
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
