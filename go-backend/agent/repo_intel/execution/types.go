package execution

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// ExecutionStatus represents the execution lifecycle state of a remediation task.
type ExecutionStatus string

const (
	StatusPending            ExecutionStatus = "PENDING"
	StatusPreparingWorkspace ExecutionStatus = "PREPARING_WORKSPACE"
	StatusExecuting          ExecutionStatus = "EXECUTING"
	StatusValidating         ExecutionStatus = "VALIDATING"
	StatusCompleted          ExecutionStatus = "COMPLETED"
	StatusFailed             ExecutionStatus = "FAILED"
	StatusRejected           ExecutionStatus = "REJECTED"
	StatusRolledBack         ExecutionStatus = "ROLLED_BACK"
	StatusSkipped            ExecutionStatus = "SKIPPED"
)

// ValidationStatus represents the outcome of post-patch validation.
type ValidationStatus string

const (
	ValPassed  ValidationStatus = "PASSED"
	ValFailed  ValidationStatus = "FAILED"
	ValSkipped ValidationStatus = "SKIPPED"
	ValWarning ValidationStatus = "WARNING"
)

// ExecutionConfig holds the configuration for the execution subsystem.
type ExecutionConfig struct {
	RepoRoot         string   `json:"repo_root"`
	PlanPath         string   `json:"plan_path"`
	DryRun           bool     `json:"dry_run"`
	WorkspaceBaseDir string   `json:"workspace_base_dir"`
	ApprovedTaskIDs  []string `json:"approved_task_ids"`
	TargetTaskIDs    []string `json:"target_task_ids"`
	BatchLimit       int      `json:"batch_limit"`
	KeepWorkspaces   bool     `json:"keep_workspaces"`
	TimeoutSec       int      `json:"timeout_sec"`
	Verbose          bool     `json:"verbose"`
}

// UnifiedPatch represents an atomic, reversible diff generated in an isolated workspace.
type UnifiedPatch struct {
	PatchID     string    `json:"patch_id"`
	TaskID      string    `json:"task_id"`
	TargetFiles []string  `json:"target_files"`
	DiffContent string    `json:"diff_content"`
	SHA256      string    `json:"sha256"`
	LineCount   int       `json:"line_count"`
	BytesCount  int       `json:"bytes_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// BuildCheckResult records build/compiler verification telemetry.
type BuildCheckResult struct {
	Passed     bool   `json:"passed"`
	Command    string `json:"command"`
	Output     string `json:"output,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// TestCheckResult records unit and regression test suite execution telemetry.
type TestCheckResult struct {
	Passed      bool   `json:"passed"`
	Command     string `json:"command"`
	Output      string `json:"output,omitempty"`
	TestsPassed int    `json:"tests_passed"`
	TestsFailed int    `json:"tests_failed"`
	DurationMs  int64  `json:"duration_ms"`
}

// DefectResolutionResult records verification that the source diagnostic finding is resolved.
type DefectResolutionResult struct {
	DefectEliminated bool   `json:"defect_eliminated"`
	Details          string `json:"details"`
}

// RegressionCheckResult records checks ensuring no new architectural/structural violations.
type RegressionCheckResult struct {
	NoRegressions bool   `json:"no_regressions"`
	Details       string `json:"details"`
}

// ValidationReport aggregates all independent validation checks for a generated patch.
type ValidationReport struct {
	OverallStatus    ValidationStatus       `json:"overall_status"`
	BuildCheck       BuildCheckResult       `json:"build_check"`
	TestCheck        TestCheckResult        `json:"test_check"`
	DefectResolution DefectResolutionResult `json:"defect_resolution"`
	RegressionCheck  RegressionCheckResult  `json:"regression_check"`
	DurationMs       int64                  `json:"duration_ms"`
}

// TaskExecutionResult records the complete outcome of executing a single remediation task.
type TaskExecutionResult struct {
	TaskID                 string              `json:"task_id"`
	Title                  string              `json:"title"`
	Status                 ExecutionStatus     `json:"status"`
	ExecutionClass         string              `json:"execution_class"`
	StrategyUsed           string              `json:"strategy_used"`
	StartTime              time.Time           `json:"start_time"`
	EndTime                time.Time           `json:"end_time"`
	DurationMs             int64               `json:"duration_ms"`
	WorkspacePath          string              `json:"workspace_path,omitempty"`
	TargetFiles            []string            `json:"target_files"`
	Patch                  *UnifiedPatch       `json:"patch,omitempty"`
	Validation             ValidationReport    `json:"validation"`
	RollbackOccurred       bool                `json:"rollback_occurred"`
	RollbackReason         string              `json:"rollback_reason,omitempty"`
	ErrorMessage           string              `json:"error_message,omitempty"`
	ReviewRequired         bool                `json:"review_required"`
	HumanApprovalVerified  bool                `json:"human_approval_verified"`
	TaskRef                *planning.RemediationTask `json:"-"`
}

// WorkspaceAuditRecord documents isolated workspace lifecycle for audit trails.
type WorkspaceAuditRecord struct {
	TaskID       string    `json:"task_id"`
	WorkspaceDir string    `json:"workspace_dir"`
	CreatedAt    time.Time `json:"created_at"`
	DestroyedAt  time.Time `json:"destroyed_at,omitempty"`
	Retained     bool      `json:"retained"`
	CleanedUp    bool      `json:"cleaned_up"`
	SizeBytes    int64     `json:"size_bytes"`
}

// ExecutionSummary summarizes the overall results of an execution batch.
type ExecutionSummary struct {
	TotalScheduled      int   `json:"total_scheduled"`
	Completed           int   `json:"completed"`
	Failed              int   `json:"failed"`
	Rejected            int   `json:"rejected"`
	RolledBack          int   `json:"rolled_back"`
	ReviewRequiredCount int   `json:"review_required_count"`
	TotalDurationMs     int64 `json:"total_duration_ms"`
}

// ExecutionReport is the complete auditable record of an execution run.
type ExecutionReport struct {
	ReportID        string                 `json:"report_id"`
	PlanID          string                 `json:"plan_id"`
	PlanHash        string                 `json:"plan_hash"`
	GeneratedAt     time.Time              `json:"generated_at"`
	BaselineCommit  string                 `json:"baseline_commit"`
	Summary         ExecutionSummary       `json:"summary"`
	TaskResults     []TaskExecutionResult  `json:"task_results"`
	PatchManifest   []UnifiedPatch         `json:"patch_manifest"`
	WorkspaceAudit  []WorkspaceAuditRecord `json:"workspace_audit"`
}

// ProvenanceCheckResult records the result of verifying the input plan's integrity and freshness.
type ProvenanceCheckResult struct {
	PlanID            string   `json:"plan_id"`
	PlanHash          string   `json:"plan_hash"`
	Valid             bool     `json:"valid"`
	BaselineCommit    string   `json:"baseline_commit"`
	CurrentHeadCommit string   `json:"current_head_commit"`
	CleanWorkingTree  bool     `json:"clean_working_tree"`
	StaleTasks        []string `json:"stale_tasks"`
	Issues            []string `json:"issues"`
}

