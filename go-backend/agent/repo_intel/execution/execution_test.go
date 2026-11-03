package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

func TestPlanLoader(t *testing.T) {
	// 1. Valid plan loading
	planPath := filepath.Join("..", "..", "..", "..", "reports", "remediation_plan.json")
	plan, hash, err := LoadRemediationPlan(planPath)
	if err != nil {
		t.Fatalf("Failed to load valid remediation plan: %v", err)
	}
	if plan.PlanID == "" {
		t.Errorf("Expected non-empty plan ID")
	}
	if len(hash) != 64 {
		t.Errorf("Expected 64-char SHA256 hash, got %d", len(hash))
	}
	if len(plan.Tasks) == 0 {
		t.Errorf("Expected tasks in plan, got 0")
	}

	// 2. Invalid schema rejection
	tmpFile := filepath.Join(t.TempDir(), "bad_schema.json")
	_ = os.WriteFile(tmpFile, []byte(`{"plan_schema_version":"2.0","plan_id":"TEST"}`), 0644)
	_, _, err = LoadRemediationPlan(tmpFile)
	if err == nil || !strings.Contains(err.Error(), "unsupported plan schema version") {
		t.Errorf("Expected schema version error, got %v", err)
	}
}

func TestProvenanceVerification(t *testing.T) {
	repoRoot, _ := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	plan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-PROV-TEST",
		Tasks: []planning.RemediationTask{
			{
				TaskID:      "RP-REAL",
				TargetFiles: []string{"go-backend/Dockerfile"},
			},
			{
				TaskID:      "RP-MISSING",
				TargetFiles: []string{"nonexistent/file/path.go"},
			},
		},
	}

	prov, err := VerifyPlanProvenance(plan, "test-hash", repoRoot)
	if err != nil {
		t.Fatalf("Unexpected provenance error: %v", err)
	}
	if len(prov.StaleTasks) != 1 || prov.StaleTasks[0] != "RP-MISSING" {
		t.Errorf("Expected RP-MISSING to be flagged stale, got %v", prov.StaleTasks)
	}
}

func TestEligibilityAndHumanApprovalGate(t *testing.T) {
	// Case 1: AUTOMATION_READY
	taskAuto := &planning.RemediationTask{
		TaskID:         "TASK-AUTO",
		ExecutionClass: planning.ClassAutomationReady,
	}
	eligible, reviewReq, reason := CheckTaskEligibility(taskAuto, nil)
	if !eligible || reviewReq || reason != "" {
		t.Errorf("Expected AUTOMATION_READY to be eligible without review, got eligible=%v, reviewReq=%v, reason=%s", eligible, reviewReq, reason)
	}

	// Case 2: AUTOMATION_WITH_REVIEW
	taskReview := &planning.RemediationTask{
		TaskID:         "TASK-REVIEW",
		ExecutionClass: planning.ClassAutomationWithReview,
	}
	eligible, reviewReq, reason = CheckTaskEligibility(taskReview, nil)
	if !eligible || !reviewReq || reason != "" {
		t.Errorf("Expected AUTOMATION_WITH_REVIEW to be eligible with reviewReq=true, got eligible=%v, reviewReq=%v", eligible, reviewReq)
	}

	// Case 3: HUMAN_APPROVAL_REQUIRED without approval (Fail Closed)
	taskApproval := &planning.RemediationTask{
		TaskID:                "TASK-SEC",
		ExecutionClass:        planning.ClassHumanApprovalRequired,
		HumanApprovalRequired: true,
		ApprovalRationale:     "Secret rotation",
	}
	eligible, _, reason = CheckTaskEligibility(taskApproval, nil)
	if eligible || !strings.Contains(reason, "EXEC_REJECTED_UNAPPROVED") {
		t.Errorf("Expected unapproved task to fail closed, got eligible=%v, reason=%s", eligible, reason)
	}

	// Case 4: HUMAN_APPROVAL_REQUIRED with approval
	approved := map[string]bool{"TASK-SEC": true}
	eligible, _, reason = CheckTaskEligibility(taskApproval, approved)
	if !eligible || reason != "" {
		t.Errorf("Expected approved task to be eligible, got eligible=%v, reason=%s", eligible, reason)
	}

	// Case 5: INVESTIGATION_REQUIRED
	taskInv := &planning.RemediationTask{
		TaskID:         "TASK-INV",
		ExecutionClass: planning.ClassInvestigationRequired,
	}
	eligible, _, reason = CheckTaskEligibility(taskInv, nil)
	if eligible || !strings.Contains(reason, "EXEC_REJECTED") {
		t.Errorf("Expected INVESTIGATION_REQUIRED to be rejected, got eligible=%v, reason=%s", eligible, reason)
	}
}

func TestScheduler(t *testing.T) {
	plan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-SCHED",
		Tasks: []planning.RemediationTask{
			{TaskID: "RP-1", Dependencies: []string{"RP-0"}},
			{TaskID: "RP-0", Dependencies: []string{}},
			{TaskID: "RP-2", Dependencies: []string{"RP-1"}},
		},
	}

	sched, err := BuildExecutionSchedule(plan, nil, 0)
	if err != nil {
		t.Fatalf("Failed to build schedule: %v", err)
	}
	if len(sched.OrderedTasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(sched.OrderedTasks))
	}
	if sched.OrderedTasks[0].TaskID != "RP-0" || sched.OrderedTasks[1].TaskID != "RP-1" || sched.OrderedTasks[2].TaskID != "RP-2" {
		t.Errorf("Incorrect topological order: %s, %s, %s",
			sched.OrderedTasks[0].TaskID, sched.OrderedTasks[1].TaskID, sched.OrderedTasks[2].TaskID)
	}

	// Test batch limit
	batchSched, err := BuildExecutionSchedule(plan, nil, 2)
	if err != nil || len(batchSched.OrderedTasks) != 2 {
		t.Errorf("Expected batch limit 2, got %d", len(batchSched.OrderedTasks))
	}
}

func TestWorkspacePathTraversalProtection(t *testing.T) {
	repoRoot, _ := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	wm, err := NewWorkspaceManager(repoRoot, t.TempDir(), false)
	if err != nil {
		t.Fatalf("Failed to init workspace manager: %v", err)
	}

	wsDir := filepath.Join(wm.BaseDir, "ws_test")
	_ = os.MkdirAll(wsDir, 0755)

	// Relative escape
	_, err = wm.ValidateContainedPath(wsDir, "../../etc/passwd")
	if err == nil {
		t.Errorf("Expected path traversal error for ../../etc/passwd, got nil")
	}

	// Absolute path escape
	_, err = wm.ValidateContainedPath(wsDir, "C:\\Windows\\System32")
	if err == nil {
		t.Errorf("Expected path traversal error for absolute path, got nil")
	}
}

func TestPatchGenerationAndScope(t *testing.T) {
	tmpOrig := t.TempDir()
	tmpMod := t.TempDir()

	file1 := "app.txt"
	_ = os.WriteFile(filepath.Join(tmpOrig, file1), []byte("alpha\nbeta\ngamma\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpMod, file1), []byte("alpha\nbeta modified\ngamma\n"), 0644)

	patch, err := GenerateUnifiedPatch("TEST-TASK", tmpOrig, tmpMod, []string{file1})
	if err != nil {
		t.Fatalf("Failed generating patch: %v", err)
	}
	if patch == nil {
		t.Fatalf("Expected non-nil patch")
	}
	if patch.LineCount == 0 || patch.SHA256 == "" {
		t.Errorf("Invalid patch metadata: %v", patch)
	}
	if !strings.Contains(patch.DiffContent, "--- a/app.txt") || !strings.Contains(patch.DiffContent, "+++ b/app.txt") {
		t.Errorf("Missing unified diff headers in diff: %s", patch.DiffContent)
	}

	// Scope validation: allowed file passes
	if err := ValidatePatchScope(patch, []string{file1}); err != nil {
		t.Errorf("Allowed file should pass scope validation: %v", err)
	}

	// Scope validation: disallowed file fails
	if err := ValidatePatchScope(patch, []string{"other.txt"}); err == nil {
		t.Errorf("Unpermitted file should fail scope validation")
	}
}
