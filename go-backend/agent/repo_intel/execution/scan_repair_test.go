package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for relPath, content := range files {
		absPath := filepath.Join(dir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			t.Fatalf("write file failed: %v", err)
		}
	}
	return dir
}

func TestScanToRepairWorkflow_Success(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testapp\n\ngo 1.21\n",
		"server/logger.go": `package server

import (
	"io"
	"os"
)

func DumpLogs(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID := "ws-test-success"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	// 1. List targets for workspace
	targets := svc.ListTargets(wsID)
	for _, trg := range targets {
		t.Logf("Found target: ID=%s Stratum=%s Desc=%s", trg.TargetID, trg.Stratum, trg.Description)
	}
	if len(targets) == 0 {
		t.Fatalf("Expected targets for workspace %s, got 0", wsID)
	}

	var scannedTarget *MVPTarget
	var demoTargetCount int
	for _, target := range targets {
		if target.IsDemoTarget {
			demoTargetCount++
			if !strings.HasPrefix(target.Description, "[DEMO]") {
				t.Errorf("Demo target %s description missing [DEMO] prefix: %s", target.TargetID, target.Description)
			}
		} else if target.Stratum == "SCANNED_REPOSITORY" && (strings.Contains(target.TargetFile, "logger.go") || strings.Contains(target.TargetID, "DumpLogs")) {
			targetCopy := target
			scannedTarget = &targetCopy
		}
	}

	if scannedTarget == nil {
		t.Fatalf("Expected scanned target for server/logger.go, but found none")
	}
	if demoTargetCount == 0 {
		t.Errorf("Expected preconfigured demo targets to be present")
	}

	// 2. Analyze Target
	analysis, err := svc.AnalyzeTarget(scannedTarget.TargetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}
	if !analysis.AutomaticEligibility {
		t.Errorf("Expected intra-procedural resource leak to be automatically eligible, got false (rationale: %s)", analysis.EligibilityRationale)
	}
	if analysis.UncertaintyScore > 0.10 {
		t.Errorf("Expected low uncertainty for intra-procedural leak, got %.2f", analysis.UncertaintyScore)
	}
	if !strings.Contains(analysis.CodeSnippet, "os.Open") {
		t.Errorf("Expected code snippet to contain os.Open, got:\n%s", analysis.CodeSnippet)
	}

	// 3. Propose Repair
	proposal, err := svc.ProposeRepair(analysis.AnalysisID)
	if err != nil {
		t.Fatalf("ProposeRepair failed: %v", err)
	}
	if proposal.Status != "PROPOSED" {
		t.Fatalf("Expected status PROPOSED, got %s (reason: %s)", proposal.Status, proposal.AbstainReason)
	}
	if !strings.Contains(proposal.UnifiedDiff, "defer f.Close()") {
		t.Errorf("Expected patch to contain 'defer f.Close()', got:\n%s", proposal.UnifiedDiff)
	}

	// 4. Approve Repair
	approval, err := svc.ApproveRepair(proposal.ProposalID, "lead-engineer@code-vitals.internal", "Reliability Lead")
	if err != nil {
		t.Fatalf("ApproveRepair failed: %v", err)
	}
	if !approval.Valid || approval.TokenID == "" {
		t.Fatalf("Expected valid approval token, got %+v", approval)
	}

	// 5. Execute Repair in Sandbox
	execRes, err := svc.ExecuteRepair(approval.TokenID)
	if err != nil {
		t.Fatalf("ExecuteRepair failed: %v", err)
	}
	if !execRes.AppliedInSandbox {
		t.Errorf("Expected applied in sandbox true")
	}
	if !execRes.TokenConsumed {
		t.Errorf("Expected token consumed true")
	}
	if !execRes.OriginalRepoImmutable {
		t.Fatalf("CRITICAL: Original repository was modified!")
	}
	if !execRes.AllOraclesPassed {
		t.Errorf("Expected all oracles to pass, verdict: %s", execRes.VerificationVerdict)
	}

	// 6. Verify Single-Use Replay Protection
	_, replayErr := svc.ExecuteRepair(approval.TokenID)
	if replayErr == nil {
		t.Fatalf("Expected replay of consumed token to fail, but it succeeded")
	}
}

func TestScanToRepairWorkflow_UnsupportedCommandInjection(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testcmd\n\ngo 1.21\n",
		"runner/cmd.go": `package runner

import "os/exec"

func ExecuteHook(hook string) error {
	c := exec.Command("sh", "-c", "echo "+hook)
	return c.Run()
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID := "ws-test-cmd"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	targets := svc.ListTargets(wsID)
	var cmdTarget *MVPTarget
	for _, target := range targets {
		if target.DefectCategory == "COMMAND_INJECTION" && (strings.Contains(target.TargetFile, "cmd.go") || strings.Contains(target.TargetID, "ExecuteHook")) {
			targetCopy := target
			cmdTarget = &targetCopy
			break
		}
	}

	if cmdTarget == nil {
		t.Fatalf("Expected COMMAND_INJECTION target for runner/cmd.go")
	}

	// Analyze
	analysis, err := svc.AnalyzeTarget(cmdTarget.TargetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}
	if analysis.AutomaticEligibility {
		t.Errorf("Command injection must not be automatically eligible")
	}
	if analysis.UncertaintyScore < 0.90 {
		t.Errorf("Expected high uncertainty (>= 0.90) for command injection, got %.2f", analysis.UncertaintyScore)
	}
	if !strings.Contains(analysis.EligibilityRationale, "Conservative fail-closed") &&
		!strings.Contains(analysis.EligibilityRationale, "fail-closed") {
		t.Errorf("Expected fail-closed rationale, got: %s", analysis.EligibilityRationale)
	}

	// Propose Repair must safely abstain
	proposal, err := svc.ProposeRepair(analysis.AnalysisID)
	if err != nil {
		t.Fatalf("ProposeRepair returned error: %v", err)
	}
	if proposal.Status != "ABSTAINED" {
		t.Errorf("Expected status ABSTAINED for command injection, got %s", proposal.Status)
	}
	if proposal.AbstainCategory == "" {
		t.Errorf("Expected non-empty abstain category")
	}
}

func TestScanToRepairWorkflow_EscapingResourceAbstention(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testescape\n\ngo 1.21\n",
		"storage/store.go": `package storage

import "os"

func OpenStore(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID := "ws-test-escape"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	targets := svc.ListTargets(wsID)
	var escTarget *MVPTarget
	for _, target := range targets {
		if strings.Contains(target.TargetID, "OpenStore") {
			targetCopy := target
			escTarget = &targetCopy
			break
		}
	}

	if escTarget == nil {
		t.Fatalf("Expected target for OpenStore")
	}

	analysis, err := svc.AnalyzeTarget(escTarget.TargetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}
	if analysis.AutomaticEligibility {
		t.Errorf("Escaping resource handle must not be automatically eligible")
	}
	if analysis.UncertaintyScore < 0.85 {
		t.Errorf("Expected high uncertainty for escaping handle, got %.2f", analysis.UncertaintyScore)
	}
	if !strings.Contains(analysis.EligibilityRationale, "escapes") {
		t.Errorf("Expected rationale to explain escaping handle, got: %s", analysis.EligibilityRationale)
	}
}

func TestScanToRepairWorkflow_VerificationFailure_FailClosed(t *testing.T) {
	wsDir := createTestWorkspace(t, map[string]string{
		"go.mod": "module testfail\n\ngo 1.21\n",
		"service/broker.go": `package service

import (
	"io"
	"os"
)

func ReadConfig(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}
`,
	})

	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID := "ws-test-fail"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	targets := svc.ListTargets(wsID)
	var trg *MVPTarget
	for _, target := range targets {
		if target.Stratum == "SCANNED_REPOSITORY" && strings.Contains(target.TargetFile, "broker.go") {
			targetCopy := target
			trg = &targetCopy
			break
		}
	}
	if trg == nil {
		t.Fatalf("Expected target for broker.go")
	}

	analysis, err := svc.AnalyzeTarget(trg.TargetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}

	proposal, err := svc.ProposeRepair(analysis.AnalysisID)
	if err != nil {
		t.Fatalf("ProposeRepair failed: %v", err)
	}

	// Deliberately corrupt the modified source in proposal to trigger verification failure
	proposal.ModifiedSource = "package service\n\nfunc BrokenSyntax({{{ illegal"

	approval, err := svc.ApproveRepair(proposal.ProposalID, "audit-lead@code-vitals.internal", "Verification Lead")
	if err != nil {
		t.Fatalf("ApproveRepair failed: %v", err)
	}

	execRes, err := svc.ExecuteRepair(approval.TokenID)
	if err != nil {
		t.Fatalf("ExecuteRepair returned error instead of fail-closed verdict: %v", err)
	}

	if execRes.AllOraclesPassed {
		t.Errorf("Expected verification failure for broken syntax, but oracles reported passed")
	}
	if execRes.VerificationVerdict != "FAILED" {
		t.Errorf("Expected verdict 'FAILED', got %s", execRes.VerificationVerdict)
	}
	if !execRes.OriginalRepoImmutable {
		t.Fatalf("CRITICAL: Original repository modified during failing verification!")
	}
	if !execRes.TokenConsumed {
		t.Errorf("Token must be consumed even on verification failure to prevent replay")
	}
}

func TestScanToRepairWorkflow_MissingWorkspace_FailClosed(t *testing.T) {
	svc, err := NewMVPService("", "")
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	// Create a dummy dynamic target pointing to a non-existent workspace
	targetID := "SCAN-ORPHAN-1"
	orphanTarget := &MVPTarget{
		TargetID:       targetID,
		Repository:     "orphan_repo",
		TargetFile:     "main.go",
		DefectCategory: "RESOURCE_LEAK",
		Stratum:        "SCANNED_REPOSITORY",
		ExpectedAction: "REPAIR",
		Description:    "Orphan Target",
		WorkspaceID:    "non-existent-workspace-id",
		LineNumber:     10,
	}
	svc.mu.Lock()
	svc.dynamicTargets[targetID] = orphanTarget
	svc.mu.Unlock()

	analysis, err := svc.AnalyzeTarget(targetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}
	if analysis.AutomaticEligibility {
		t.Errorf("Missing workspace target must not be automatically eligible")
	}
	if analysis.UncertaintyScore < 0.95 {
		t.Errorf("Expected uncertainty >= 0.95 for missing workspace, got %.2f", analysis.UncertaintyScore)
	}
	if !strings.Contains(analysis.EligibilityRationale, "fail-closed") {
		t.Errorf("Expected fail-closed rationale, got: %s", analysis.EligibilityRationale)
	}
}
