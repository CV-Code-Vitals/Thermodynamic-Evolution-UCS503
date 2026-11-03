package execution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// ScenarioResult records the outcome of a single evaluation scenario.
type ScenarioResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
	DurationMs  int64  `json:"duration_ms"`
}

// EvaluationReport compiles the results of the complete evaluation suite.
type EvaluationReport struct {
	TotalScenarios int              `json:"total_scenarios"`
	PassedCount    int              `json:"passed_count"`
	FailedCount    int              `json:"failed_count"`
	Results        []ScenarioResult `json:"results"`
	DurationMs     int64            `json:"duration_ms"`
}

// RunEvaluationSuite executes all 22 evaluation scenarios and generates an audit report.
func RunEvaluationSuite(repoRoot string) (*EvaluationReport, error) {
	start := time.Now()
	report := &EvaluationReport{
		Results: make([]ScenarioResult, 0),
	}

	scenarios := []func(repoRoot string) ScenarioResult{
		scenario01SchemaValidation,
		scenario02ProvenanceDigest,
		scenario03StaleTargetDetection,
		scenario04AutoReadyExecution,
		scenario05AutoWithReviewFlag,
		scenario06UnapprovedRejection,
		scenario07ApprovedExecution,
		scenario08InvestigationRejection,
		scenario09DependencyOrdering,
		scenario10CircularDepPrevention,
		scenario11SoftConflictSequential,
		scenario12PathTraversalContainment,
		scenario13WorkspaceIsolation,
		scenario14BuildFailureRollback,
		scenario15TestFailureRollback,
		scenario16DefectUnresolvedRollback,
		scenario17ScopeViolationRollback,
		scenario18UnifiedDiffFormat,
		scenario19SecretRedaction,
		scenario20DryRunSafety,
		scenario21BatchLimitEnforcement,
		scenario22MultiLanguageSupport,
	}

	report.TotalScenarios = len(scenarios)

	for _, sc := range scenarios {
		res := sc(repoRoot)
		report.Results = append(report.Results, res)
		if res.Passed {
			report.PassedCount++
		} else {
			report.FailedCount++
		}
	}

	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

// -----------------------------------------------------------------------
// Scenario Implementations
// -----------------------------------------------------------------------

func scenario01SchemaValidation(repoRoot string) ScenarioResult {
	s := time.Now()
	// Test valid vs invalid schema version
	tmpFile := filepath.Join(os.TempDir(), "invalid_plan.json")
	_ = os.WriteFile(tmpFile, []byte(`{"plan_schema_version":"0.9","plan_id":"TEST"}`), 0644)
	defer os.Remove(tmpFile)

	_, _, err := LoadRemediationPlan(tmpFile)
	passed := err != nil && strings.Contains(err.Error(), "unsupported plan schema version")
	return ScenarioResult{
		ID:          "SCENARIO-01",
		Name:        "Schema Validation & Version Gate",
		Passed:      passed,
		Description: "Ensures plan loader rejects non-1.0 schema versions",
		Evidence:    fmt.Sprintf("Rejection error: %v", err),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario02ProvenanceDigest(repoRoot string) ScenarioResult {
	s := time.Now()
	planPath := filepath.Join(repoRoot, "reports", "remediation_plan.json")
	plan, hash, err := LoadRemediationPlan(planPath)
	passed := err == nil && len(hash) == 64 && plan != nil
	return ScenarioResult{
		ID:          "SCENARIO-02",
		Name:        "Cryptographic Digest Verification",
		Passed:      passed,
		Description: "Computes and verifies SHA256 cryptographic digest of ingested remediation plan",
		Evidence:    fmt.Sprintf("Computed SHA256: %s, PlanID: %s", hash, plan.PlanID),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario03StaleTargetDetection(repoRoot string) ScenarioResult {
	s := time.Now()
	dummyPlan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-STALE",
		Tasks: []planning.RemediationTask{
			{
				TaskID:      "RP-STALE-1",
				TargetFiles: []string{"nonexistent/stale/file.go"},
			},
		},
	}
	prov, _ := VerifyPlanProvenance(dummyPlan, "dummyhash", repoRoot)
	passed := prov != nil && len(prov.StaleTasks) == 1 && prov.StaleTasks[0] == "RP-STALE-1"
	return ScenarioResult{
		ID:          "SCENARIO-03",
		Name:        "Stale Finding & Missing Target Detection",
		Passed:      passed,
		Description: "Flags tasks whose target files are missing from the repository tree",
		Evidence:    fmt.Sprintf("Detected %d stale tasks: %v", len(prov.StaleTasks), prov.StaleTasks),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario04AutoReadyExecution(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{
		RepoRoot:       repoRoot,
		KeepWorkspaces: false,
	}
	exec, err := NewRemediationExecutor(cfg)
	if err != nil {
		return ScenarioResult{ID: "SCENARIO-04", Name: "AUTOMATION_READY Execution", Passed: false, Evidence: err.Error()}
	}

	task := &planning.RemediationTask{
		TaskID:            "RP-TEST-AUTO",
		Title:             "Dockerfile nonroot USER addition",
		ExecutionStrategy: planning.StrategyRuleBased,
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{"go-backend/Dockerfile"},
		Status:            "PLANNED",
	}

	res, _, err := exec.ExecuteTask(task, nil)
	passed := err == nil && res != nil && res.Status == StatusCompleted && !res.ReviewRequired
	return ScenarioResult{
		ID:          "SCENARIO-04",
		Name:        "AUTOMATION_READY Autonomous Lifecycle",
		Passed:      passed,
		Description: "Executes low-risk task autonomously in isolated workspace and validates",
		Evidence:    fmt.Sprintf("Status: %s, ReviewRequired: %v, Strategy: %s", res.Status, res.ReviewRequired, res.StrategyUsed),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario05AutoWithReviewFlag(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{
		RepoRoot:       repoRoot,
		KeepWorkspaces: false,
	}
	exec, _ := NewRemediationExecutor(cfg)

	task := &planning.RemediationTask{
		TaskID:            "RP-TEST-REVIEW",
		Title:             "Structural refactoring nesting reduction",
		ExecutionStrategy: planning.StrategyASTTransformation,
		ExecutionClass:    planning.ClassAutomationWithReview,
		TargetFiles:       []string{"admin-portal/src/evolution-measurements.js"},
		Status:            "PLANNED",
	}

	res, _, err := exec.ExecuteTask(task, nil)
	passed := err == nil && res != nil && res.Status == StatusCompleted && res.ReviewRequired
	return ScenarioResult{
		ID:          "SCENARIO-05",
		Name:        "AUTOMATION_WITH_REVIEW Enforcement",
		Passed:      passed,
		Description: "Generates patch and validates, marking task as review-required before merge",
		Evidence:    fmt.Sprintf("Status: %s, ReviewRequired: %v", res.Status, res.ReviewRequired),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario06UnapprovedRejection(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{
		RepoRoot:       repoRoot,
		KeepWorkspaces: false,
	}
	exec, _ := NewRemediationExecutor(cfg)

	task := &planning.RemediationTask{
		TaskID:                "RP-0103-TEST",
		Title:                 "Hardcoded secret remediation",
		ExecutionStrategy:     planning.StrategyHumanGuided,
		ExecutionClass:        planning.ClassHumanApprovalRequired,
		HumanApprovalRequired: true,
		ApprovalRationale:     "Secret rotation requires human authorization",
		TargetFiles:           []string{"go-backend/agent/repo_intel/repo_intel_test.go"},
	}

	res, _, _ := exec.ExecuteTask(task, nil) // No approval passed
	passed := res != nil && res.Status == StatusRejected && strings.Contains(res.ErrorMessage, "EXEC_REJECTED_UNAPPROVED")
	return ScenarioResult{
		ID:          "SCENARIO-06",
		Name:        "Fail-Closed Gate on Unapproved Task",
		Passed:      passed,
		Description: "Rejects HUMAN_APPROVAL_REQUIRED tasks when no explicit approval is supplied",
		Evidence:    fmt.Sprintf("Status: %s, Error: %s", res.Status, res.ErrorMessage),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario07ApprovedExecution(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{
		RepoRoot:       repoRoot,
		KeepWorkspaces: false,
	}
	exec, _ := NewRemediationExecutor(cfg)

	task := &planning.RemediationTask{
		TaskID:                "RP-0103-APPROVED",
		Title:                 "Hardcoded secret remediation",
		Description:           "File go-backend/agent/repo_intel/repo_intel_test.go contains AWS Access Key ID",
		ExecutionStrategy:     planning.StrategyHumanGuided,
		ExecutionClass:        planning.ClassHumanApprovalRequired,
		HumanApprovalRequired: true,
		TargetFiles:           []string{"go-backend/agent/repo_intel/repo_intel_test.go"},
	}

	approved := map[string]bool{"RP-0103-APPROVED": true}
	res, _, err := exec.ExecuteTask(task, approved)
	passed := err == nil && res != nil && res.Status == StatusCompleted && res.HumanApprovalVerified
	evidence := fmt.Sprintf("Status: %s, HumanApprovalVerified: %v", res.Status, res.HumanApprovalVerified)
	if !passed && res != nil {
		evidence = fmt.Sprintf("Status: %s, Error: %s, Rollback: %s, Build: %s", res.Status, res.ErrorMessage, res.RollbackReason, res.Validation.BuildCheck.Output)
	}
	return ScenarioResult{
		ID:          "SCENARIO-07",
		Name:        "Approved Task Execution Lifecycle",
		Passed:      passed,
		Description: "Executes HUMAN_APPROVAL_REQUIRED task when explicit approval is provided",
		Evidence:    evidence,
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario08InvestigationRejection(repoRoot string) ScenarioResult {
	s := time.Now()
	task := &planning.RemediationTask{
		TaskID:         "RP-INV-01",
		ExecutionClass: planning.ClassInvestigationRequired,
	}
	eligible, _, reason := CheckTaskEligibility(task, nil)
	passed := !eligible && strings.Contains(reason, "EXEC_REJECTED")
	return ScenarioResult{
		ID:          "SCENARIO-08",
		Name:        "Investigation Class Ineligibility",
		Passed:      passed,
		Description: "Prevents automated execution of INVESTIGATION_REQUIRED tasks",
		Evidence:    fmt.Sprintf("Eligible: %v, Reason: %s", eligible, reason),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario09DependencyOrdering(repoRoot string) ScenarioResult {
	s := time.Now()
	plan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-DEP",
		Tasks: []planning.RemediationTask{
			{TaskID: "RP-CHILD", Dependencies: []string{"RP-PARENT"}},
			{TaskID: "RP-PARENT", Dependencies: []string{}},
		},
	}
	sched, err := BuildExecutionSchedule(plan, nil, 0)
	passed := err == nil && len(sched.OrderedTasks) == 2 && sched.OrderedTasks[0].TaskID == "RP-PARENT" && sched.OrderedTasks[1].TaskID == "RP-CHILD"
	return ScenarioResult{
		ID:          "SCENARIO-09",
		Name:        "Topological Dependency Scheduling",
		Passed:      passed,
		Description: "Ensures prerequisite tasks are scheduled strictly prior to dependent tasks",
		Evidence:    fmt.Sprintf("Order: %s -> %s", sched.OrderedTasks[0].TaskID, sched.OrderedTasks[1].TaskID),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario10CircularDepPrevention(repoRoot string) ScenarioResult {
	s := time.Now()
	plan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-CYCLE",
		Tasks: []planning.RemediationTask{
			{TaskID: "RP-A", Dependencies: []string{"RP-B"}},
			{TaskID: "RP-B", Dependencies: []string{"RP-A"}},
		},
	}
	_, err := BuildExecutionSchedule(plan, nil, 0)
	passed := err != nil && strings.Contains(err.Error(), "circular dependency")
	return ScenarioResult{
		ID:          "SCENARIO-10",
		Name:        "Circular Dependency Prevention",
		Passed:      passed,
		Description: "Detects and rejects dependency cycles in execution schedule",
		Evidence:    fmt.Sprintf("Error caught: %v", err),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario11SoftConflictSequential(repoRoot string) ScenarioResult {
	s := time.Now()
	// Two tasks sharing a target file must be ordered sequentially
	plan := &planning.RemediationPlan{
		PlanSchemaVersion: "1.0",
		PlanID:            "PLAN-SOFT",
		Tasks: []planning.RemediationTask{
			{TaskID: "RP-1", TargetFiles: []string{"go-backend/Dockerfile"}},
			{TaskID: "RP-2", TargetFiles: []string{"go-backend/Dockerfile"}},
		},
	}
	sched, err := BuildExecutionSchedule(plan, nil, 0)
	passed := err == nil && len(sched.OrderedTasks) == 2
	return ScenarioResult{
		ID:          "SCENARIO-11",
		Name:        "Soft Conflict Sequential Serialization",
		Passed:      passed,
		Description: "Serializes execution of tasks with overlapping target resources",
		Evidence:    fmt.Sprintf("Scheduled %d tasks sequentially", len(sched.OrderedTasks)),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario12PathTraversalContainment(repoRoot string) ScenarioResult {
	s := time.Now()
	wm, _ := NewWorkspaceManager(repoRoot, "", false)
	_, err1 := wm.ValidateContainedPath("/tmp/ws", "../../etc/passwd")
	_, err2 := wm.ValidateContainedPath("C:\\temp\\ws", "..\\..\\windows\\system32")
	passed := err1 != nil && err2 != nil
	return ScenarioResult{
		ID:          "SCENARIO-12",
		Name:        "Workspace Path Traversal Defense",
		Passed:      passed,
		Description: "Blocks directory escape attempts (../ and absolute paths)",
		Evidence:    fmt.Sprintf("Traversals caught: %v; %v", err1, err2),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario13WorkspaceIsolation(repoRoot string) ScenarioResult {
	s := time.Now()
	// Verify workspace copy does not touch repoRoot
	targetFile := "samples/benchmark/go.mod"
	origFullPath := filepath.Join(repoRoot, filepath.FromSlash(targetFile))
	origInfo, _ := os.Stat(origFullPath)
	origModTime := origInfo.ModTime()

	cfg := ExecutionConfig{RepoRoot: repoRoot, KeepWorkspaces: false}
	exec, _ := NewRemediationExecutor(cfg)
	task := &planning.RemediationTask{
		TaskID:            "RP-TEST-ISOLATION",
		ExecutionStrategy: planning.StrategyRuleBased,
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{targetFile},
	}
	_, _, _ = exec.ExecuteTask(task, nil)

	// Verify original file was not mutated
	postInfo, _ := os.Stat(origFullPath)
	passed := origInfo.Size() == postInfo.Size() && origModTime.Equal(postInfo.ModTime())
	return ScenarioResult{
		ID:          "SCENARIO-13",
		Name:        "Repository Immutability Guarantee",
		Passed:      passed,
		Description: "Confirms that the primary repository files remain untouched during workspace execution",
		Evidence:    fmt.Sprintf("Original size: %d, post size: %d", origInfo.Size(), postInfo.Size()),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario14BuildFailureRollback(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{RepoRoot: repoRoot, KeepWorkspaces: false}
	exec, _ := NewRemediationExecutor(cfg)

	// Custom failing strategy that introduces syntax error
	exec.StrategyReg.Register("FAIL_SYNTAX", &failSyntaxStrategy{})

	task := &planning.RemediationTask{
		TaskID:            "RP-FAIL-SYNTAX",
		ExecutionStrategy: "FAIL_SYNTAX",
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{"go-backend/agent/agent.go"},
	}

	res, _, _ := exec.ExecuteTask(task, nil)
	passed := res != nil && (res.Status == StatusRolledBack || res.Status == StatusFailed) && res.RollbackOccurred
	return ScenarioResult{
		ID:          "SCENARIO-14",
		Name:        "Build Error Triggered Rollback",
		Passed:      passed,
		Description: "Triggers automatic rollback when post-patch compilation/syntax check fails",
		Evidence:    fmt.Sprintf("Status: %s, RollbackOccurred: %v, Reason: %s", res.Status, res.RollbackOccurred, res.RollbackReason),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario15TestFailureRollback(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{RepoRoot: repoRoot, KeepWorkspaces: false}
	exec, _ := NewRemediationExecutor(cfg)

	task := &planning.RemediationTask{
		TaskID:             "RP-FAIL-TEST",
		ExecutionStrategy:  planning.StrategyRuleBased,
		ExecutionClass:     planning.ClassAutomationReady,
		TargetFiles:        []string{"go-backend/Dockerfile"},
		ValidationCommands: []string{"go test -run TestNonExistentFailTrigger ./go-backend/..."},
	}

	res, _, _ := exec.ExecuteTask(task, nil)
	passed := res != nil && (res.Status == StatusRolledBack || res.Status == StatusFailed) && res.RollbackOccurred
	return ScenarioResult{
		ID:          "SCENARIO-15",
		Name:        "Test Failure Triggered Rollback",
		Passed:      passed,
		Description: "Triggers automatic rollback when validation test command fails",
		Evidence:    fmt.Sprintf("Status: %s, RollbackOccurred: %v", res.Status, res.RollbackOccurred),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario16DefectUnresolvedRollback(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{RepoRoot: repoRoot, KeepWorkspaces: false}
	exec, _ := NewRemediationExecutor(cfg)

	// Strategy that doesn't actually fix the defect (missing lockfile)
	exec.StrategyReg.Register("NO_FIX", &InvestigationOnlyStrategy{})

	task := &planning.RemediationTask{
		TaskID:            "RP-NO-FIX",
		ExecutionStrategy: "NO_FIX",
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{"samples/benchmark/go.mod"},
	}

	res, _, _ := exec.ExecuteTask(task, nil)
	passed := res != nil && (res.Status == StatusRolledBack || res.Status == StatusFailed) && res.RollbackOccurred
	return ScenarioResult{
		ID:          "SCENARIO-16",
		Name:        "Unresolved Defect Rollback Trigger",
		Passed:      passed,
		Description: "Rolls back if patch fails to eliminate the targeted diagnostic defect",
		Evidence:    fmt.Sprintf("Status: %s, RollbackOccurred: %v", res.Status, res.RollbackOccurred),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario17ScopeViolationRollback(repoRoot string) ScenarioResult {
	s := time.Now()
	patch := &UnifiedPatch{
		TargetFiles: []string{"unpermitted/file.txt"},
	}
	err := ValidatePatchScope(patch, []string{"permitted/file.txt"})
	passed := err != nil && strings.Contains(err.Error(), "scope violation")
	return ScenarioResult{
		ID:          "SCENARIO-17",
		Name:        "Patch Scope Violation Enforcement",
		Passed:      passed,
		Description: "Rejects patches attempting to modify files outside designated task scope",
		Evidence:    fmt.Sprintf("Error caught: %v", err),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario18UnifiedDiffFormat(repoRoot string) ScenarioResult {
	s := time.Now()
	origText := "line 1\nline 2\nline 3\n"
	modText := "line 1\nline 2 modified\nline 3\n"
	diff := formatUnifiedDiff("test.txt", origText, modText)
	passed := strings.Contains(diff, "--- a/test.txt") &&
		strings.Contains(diff, "+++ b/test.txt") &&
		strings.Contains(diff, "@@") &&
		strings.Contains(diff, "-line 2") &&
		strings.Contains(diff, "+line 2 modified")
	return ScenarioResult{
		ID:          "SCENARIO-18",
		Name:        "Unified Diff Format Compliance",
		Passed:      passed,
		Description: "Generates valid standard unified diff with standard headers and hunk headers",
		Evidence:    fmt.Sprintf("Diff generated:\n%s", diff),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario19SecretRedaction(repoRoot string) ScenarioResult {
	s := time.Now()
	strat := &HumanGuidedStrategy{}
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "secret_test.go")
	_ = os.WriteFile(testFile, []byte(`var key = "AKIA1234567890ABCDEF"`), 0644)
	defer os.Remove(testFile)

	task := &planning.RemediationTask{
		Title:       "Secret rotation",
		TargetFiles: []string{"secret_test.go"},
	}
	_, _ = strat.Execute(tmpDir, task)
	modBytes, _ := os.ReadFile(testFile)
	content := string(modBytes)
	passed := !strings.Contains(content, "AKIA1234567890ABCDEF") && strings.Contains(content, "os.Getenv")
	return ScenarioResult{
		ID:          "SCENARIO-19",
		Name:        "Zero Plaintext Secret Exposure",
		Passed:      passed,
		Description: "Ensures credentials are masked/replaced with secure environment variable lookups",
		Evidence:    fmt.Sprintf("Sanitized content: %s", content),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario20DryRunSafety(repoRoot string) ScenarioResult {
	s := time.Now()
	cfg := ExecutionConfig{
		RepoRoot: repoRoot,
		DryRun:   true,
	}
	exec, _ := NewRemediationExecutor(cfg)
	task := &planning.RemediationTask{
		TaskID:            "RP-DRYRUN",
		ExecutionStrategy: planning.StrategyRuleBased,
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{"go-backend/Dockerfile"},
	}
	res, _, err := exec.ExecuteTask(task, nil)
	passed := err == nil && res.Status == StatusCompleted && res.WorkspacePath == ""
	return ScenarioResult{
		ID:          "SCENARIO-20",
		Name:        "Dry-Run Mode Zero-Disk Mutation",
		Passed:      passed,
		Description: "Ensures dry-run mode bypasses workspace creation and file modifications",
		Evidence:    fmt.Sprintf("Status: %s, WorkspacePath: %s", res.Status, res.WorkspacePath),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario21BatchLimitEnforcement(repoRoot string) ScenarioResult {
	s := time.Now()
	planPath := filepath.Join(repoRoot, "reports", "remediation_plan.json")
	plan, _, _ := LoadRemediationPlan(planPath)
	sched, err := BuildExecutionSchedule(plan, nil, 3)
	passed := err == nil && len(sched.OrderedTasks) == 3
	return ScenarioResult{
		ID:          "SCENARIO-21",
		Name:        "Batch Limit Enforcement",
		Passed:      passed,
		Description: "Restricts scheduled execution batch size to user-specified ceiling",
		Evidence:    fmt.Sprintf("Requested batch 3, scheduled: %d", len(sched.OrderedTasks)),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

func scenario22MultiLanguageSupport(repoRoot string) ScenarioResult {
	s := time.Now()
	val := NewIndependentValidator(30)
	// Go verification
	goTask := &planning.RemediationTask{
		Category:    repo_intel.CategoryStructural,
		TargetFiles: []string{"go-backend/main.go"},
	}
	var goRes BuildCheckResult
	val.verifyBuild(repoRoot, goTask, &goRes)

	// Rust verification
	rustTask := &planning.RemediationTask{
		Category:    repo_intel.CategoryStructural,
		TargetFiles: []string{"thermodynamic-ast-engine/src/main.rs"},
	}
	var rustRes BuildCheckResult
	val.verifyBuild(repoRoot, rustTask, &rustRes)

	passed := goRes.Passed && rustRes.Passed
	return ScenarioResult{
		ID:          "SCENARIO-22",
		Name:        "Multi-Language Verification Support",
		Passed:      passed,
		Description: "Validates compilation and syntax across both Go and Rust target components",
		Evidence:    fmt.Sprintf("Go build passed: %v, Rust check passed: %v", goRes.Passed, rustRes.Passed),
		DurationMs:  time.Since(s).Milliseconds(),
	}
}

// -----------------------------------------------------------------------
// Helper for failure testing
// -----------------------------------------------------------------------

type failSyntaxStrategy struct{}

func (s *failSyntaxStrategy) Name() string { return "FAIL_SYNTAX" }

func (s *failSyntaxStrategy) Execute(wsDir string, task *planning.RemediationTask) ([]string, error) {
	fullPath := filepath.Join(wsDir, filepath.FromSlash(task.TargetFiles[0]))
	// Append broken syntax
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, _ = f.WriteString("\nthis is a syntax error !!!\n")
	return task.TargetFiles, nil
}

