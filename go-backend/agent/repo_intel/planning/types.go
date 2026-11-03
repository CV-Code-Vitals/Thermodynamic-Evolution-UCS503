package planning

import (
	"encoding/json"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ExecutionStrategy indicates the specific mechanism to execute a remediation task.
type ExecutionStrategy string

const (
	StrategyRuleBased         ExecutionStrategy = "RULE_BASED"
	StrategyASTTransformation ExecutionStrategy = "AST_TRANSFORMATION"
	StrategyLLMAssisted       ExecutionStrategy = "LLM_ASSISTED"
	StrategyHumanGuided       ExecutionStrategy = "HUMAN_GUIDED"
	StrategyInvestigationOnly ExecutionStrategy = "INVESTIGATION_ONLY"
)

// ExecutionClass categorizes authorization and review boundaries.
type ExecutionClass string

const (
	ClassAutomationReady        ExecutionClass = "AUTOMATION_READY"
	ClassAutomationWithReview   ExecutionClass = "AUTOMATION_WITH_REVIEW"
	ClassHumanApprovalRequired  ExecutionClass = "HUMAN_APPROVAL_REQUIRED"
	ClassInvestigationRequired  ExecutionClass = "INVESTIGATION_REQUIRED"
	ClassInformationalOnly      ExecutionClass = "INFORMATIONAL_ONLY"
)

// TaskPriority ranks tasks for execution order.
type TaskPriority string

const (
	PriorityUrgent TaskPriority = "URGENT"
	PriorityHigh   TaskPriority = "HIGH"
	PriorityMedium TaskPriority = "MEDIUM"
	PriorityLow    TaskPriority = "LOW"
)

// ActionabilityState evaluates finding suitability for conversion into tasks.
type ActionabilityState string

const (
	Actionable             ActionabilityState = "ACTIONABLE"
	ActionableWithReview   ActionabilityState = "ACTIONABLE_WITH_REVIEW"
	HumanApprovalRequired  ActionabilityState = "HUMAN_APPROVAL_REQUIRED"
	InvestigationRequired  ActionabilityState = "INVESTIGATION_REQUIRED"
	Informational          ActionabilityState = "INFORMATIONAL"
	RejectedAsStale        ActionabilityState = "REJECTED_AS_STALE"
	RejectedAsInvalid      ActionabilityState = "REJECTED_AS_INVALID"
)

// ConflictType classifies conflicts between tasks.
type ConflictType string

const (
	ConflictHard                 ConflictType = "HARD_CONFLICT"
	ConflictSoft                 ConflictType = "SOFT_CONFLICT"
	ConflictCoordinationRequired ConflictType = "COORDINATION_REQUIRED"
)

// TaskConflict records a detected conflict between remediation tasks.
type TaskConflict struct {
	ConflictID           string       `json:"conflict_id"`
	Type                 ConflictType `json:"type"`
	TaskID               string       `json:"task_id"`
	ConflictingTaskID    string       `json:"conflicting_task_id"`
	ConflictingResource  string       `json:"conflicting_resource"`
	Explanation          string       `json:"explanation"`
	ResolutionSuggestion string       `json:"resolution_suggestion"`
}

// RemediationTask is an executable, dependency-aware work item.
type RemediationTask struct {
	TaskID                 string              `json:"task_id"`
	Title                  string              `json:"title"`
	SourceFindingIDs       []string            `json:"source_finding_ids"`
	Category               repo_intel.Category `json:"category"`
	Description            string              `json:"description"`
	TargetFiles            []string            `json:"target_files"`
	TargetSymbols          []string            `json:"target_symbols,omitempty"`
	ExecutionStrategy      ExecutionStrategy   `json:"execution_strategy"`
	ExecutionClass         ExecutionClass      `json:"execution_class"`
	Priority               TaskPriority        `json:"priority"`
	PriorityScore          float64             `json:"priority_score"`
	PriorityRationale      string              `json:"priority_rationale"`
	ExpectedBenefit        string              `json:"expected_benefit"`
	RegressionRisk         repo_intel.Severity `json:"regression_risk"`
	Dependencies           []string            `json:"dependencies"` // Prerequisites TaskIDs
	Conflicts              []TaskConflict      `json:"conflicts,omitempty"`
	AcceptanceCriteria     []string            `json:"acceptance_criteria"`
	ValidationCommands     []string            `json:"validation_commands,omitempty"`
	RollbackConditions     []string            `json:"rollback_conditions"`
	HumanApprovalRequired  bool                `json:"human_approval_required"`
	ApprovalRationale      string              `json:"approval_rationale,omitempty"`
	Status                 string              `json:"status"` // "PLANNED", "BLOCKED", "DEFERRED"
}

// PlanSummary summarizes planning outcomes.
type PlanSummary struct {
	TotalFindingsConsidered   int `json:"total_findings_considered"`
	ActionableFindings         int `json:"actionable_findings"`
	TotalTasksGenerated       int `json:"total_tasks_generated"`
	AutomationReadyTasks      int `json:"automation_ready_tasks"`
	AutomationWithReviewTasks int `json:"automation_with_review_tasks"`
	HumanApprovalTasks        int `json:"human_approval_tasks"`
	InvestigationTasks        int `json:"investigation_tasks"`
	BlockedTasks              int `json:"blocked_tasks"`
	TotalConflictsDetected    int `json:"total_conflicts_detected"`
}

// RemediationPlan is the complete deliverable consumed by Phase 4.
type RemediationPlan struct {
	PlanSchemaVersion        string                                `json:"plan_schema_version"`
	PlanID                   string                                `json:"plan_id"`
	RepositoryIdentity       repo_intel.RepositoryMetadataRegistry `json:"repository_identity"`
	BaselineCommit           string                                `json:"baseline_commit"`
	SourceDiagnosticReportID string                                `json:"source_diagnostic_report_id"`
	SourceReportHash         string                                `json:"source_report_hash"`
	GeneratedAt              time.Time                             `json:"generated_at"`
	Summary                  PlanSummary                           `json:"summary"`
	Tasks                    []RemediationTask                     `json:"tasks"`
	ExecutionOrder           []string                              `json:"execution_order"` // Ordered TaskIDs
	DependencyGraph          map[string][]string                   `json:"dependency_graph"`
	Conflicts                []TaskConflict                        `json:"conflicts"`
	BlockedTaskGroups        []string                              `json:"blocked_task_groups,omitempty"`
	PlanningLimitations      []string                              `json:"planning_limitations"`
}

// ToJSON serializes RemediationPlan to pretty JSON.
func (rp *RemediationPlan) ToJSON() ([]byte, error) {
	return json.MarshalIndent(rp, "", "  ")
}

