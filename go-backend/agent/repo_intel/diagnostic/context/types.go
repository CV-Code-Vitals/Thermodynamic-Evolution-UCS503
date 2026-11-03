package context

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// EvidenceTrace captures a line-by-line or token-level contextual audit trail.
type EvidenceTrace struct {
	Location    repo_intel.EvidenceLocation `json:"location"`
	ASTNodeType string                      `json:"ast_node_type"`
	ContextType string                      `json:"context_type"` // e.g. "LOOP_BODY", "FUNCTION_DECL", "CALL_ARG", "STRUCT_FIELD"
	CodeSnippet string                      `json:"code_snippet"`
	Rationale   string                      `json:"rationale"`
}

// CounterEvidence represents verified facts proving a candidate finding is benign or false positive.
type CounterEvidence struct {
	RuleID      string    `json:"rule_id"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	VerifiedAt  time.Time `json:"verified_at"`
}

// FindingState defines the explicit probabilistic lifecycle state of a diagnostic finding.
type FindingState string

const (
	StateConfirmed              FindingState = "CONFIRMED"
	StateProbable               FindingState = "PROBABLE"
	StateInsufficientEvidence   FindingState = "INSUFFICIENT_EVIDENCE"
	StateCounterEvidencePresent FindingState = "COUNTER_EVIDENCE_PRESENT"
	StateHumanInvestigationReq  FindingState = "HUMAN_INVESTIGATION_REQUIRED"
)

// SuppressionRecord captures traceable audit evidence when a finding is suppressed or demoted.
type SuppressionRecord struct {
	FindingID        string       `json:"finding_id"`
	TargetFile       string       `json:"target_file"`
	SuppressionRule  string       `json:"suppression_rule"`
	Rationale        string       `json:"rationale"`
	OriginalState    FindingState `json:"original_state"`
	FinalState       FindingState `json:"final_state"`
	ConfidenceBefore float64      `json:"confidence_before"`
	ConfidenceAfter  float64      `json:"confidence_after"`
	Timestamp        time.Time    `json:"timestamp"`
}

// ContextFinding represents an enriched or newly discovered finding with contextual depth.
type ContextFinding struct {
	FindingID           string                      `json:"finding_id"`
	Category            repo_intel.Category         `json:"category"`
	Severity            repo_intel.Severity         `json:"severity"`
	Confidence          float64                     `json:"confidence"`
	ConfidenceRationale string                      `json:"confidence_rationale"`
	State               FindingState                `json:"state"`
	Title               string                      `json:"title"`
	Description         string                      `json:"description"`
	RootCause           string                      `json:"root_cause"`
	AffectedFiles       []string                    `json:"affected_files"`
	EvidenceTraces      []EvidenceTrace             `json:"evidence_traces"`
	CounterEvidence     []CounterEvidence           `json:"counter_evidence,omitempty"`
	IsFalsePositive     bool                        `json:"is_false_positive"`
	DetectionMethod     string                      `json:"detection_method"`
	RecommendedAction   string                      `json:"recommended_action"`
}

// ContextAnalysisSummary holds execution metrics for the contextual analysis engine.
type ContextAnalysisSummary struct {
	TotalFilesScanned        int      `json:"total_files_scanned"`
	ContextualFindingsFound  int      `json:"contextual_findings_found"`
	FalsePositivesEliminated int      `json:"false_positives_eliminated"`
	DurationMs               int64    `json:"duration_ms"`
	AnalyzersRun             []string `json:"analyzers_run"`
}

