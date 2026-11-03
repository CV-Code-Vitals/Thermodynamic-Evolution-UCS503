package diagnostic

import (
	"encoding/json"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ValidationState classifies whether a finding's evidence, location, and metadata are verified.
type ValidationState string

const (
	StateValidated           ValidationState = "VALIDATED"
	StatePartiallyValidated  ValidationState = "PARTIALLY_VALIDATED"
	StateStale               ValidationState = "STALE"
	StateInvalid             ValidationState = "INVALID"
	StateRequiresHumanReview ValidationState = "REQUIRES_HUMAN_REVIEW"
)

// RemediationReadiness classifies a finding's suitability for future autonomous remediation.
type RemediationReadiness string

const (
	ReadinessAutomationReady        RemediationReadiness = "AUTOMATION_READY"
	ReadinessAutomationWithReview   RemediationReadiness = "AUTOMATION_WITH_REVIEW"
	ReadinessHumanApprovalRequired  RemediationReadiness = "HUMAN_APPROVAL_REQUIRED"
	ReadinessInvestigationRequired  RemediationReadiness = "INVESTIGATION_REQUIRED"
	ReadinessInformationalOnly      RemediationReadiness = "INFORMATIONAL_ONLY"
)

// CorrelationType characterizes how findings relate to each other.
type CorrelationType string

const (
	CorrelationDuplicate             CorrelationType = "DUPLICATE"
	CorrelationParentChild           CorrelationType = "PARENT_CHILD"
	CorrelationSharedCause           CorrelationType = "SHARED_CAUSE"
	CorrelationCoordinatedResolution CorrelationType = "COORDINATED_RESOLUTION"
	CorrelationConflict              CorrelationType = "POTENTIAL_CONFLICT"
)

// FindingCorrelation links two or more findings under a structured relationship.
type FindingCorrelation struct {
	CorrelationID     string          `json:"correlation_id"`
	Type              CorrelationType `json:"type"`
	PrimaryFindingID  string          `json:"primary_finding_id"`
	RelatedFindingIDs []string        `json:"related_finding_ids"`
	Rationale         string          `json:"rationale"`
}

// ValidatedFinding represents a finding enriched with validation, independent confidence,
// calibrated severity, root-cause analysis, and remediation readiness.
type ValidatedFinding struct {
	FindingID              string                        `json:"finding_id"`
	SourceFindingID        string                        `json:"source_finding_id"`
	ValidationState        ValidationState               `json:"validation_state"`
	ValidationReason       string                        `json:"validation_reason,omitempty"`
	Category               repo_intel.Category           `json:"category"`
	Severity               repo_intel.Severity           `json:"severity"`
	SeverityRationale      string                        `json:"severity_rationale"`
	CalibratedConfidence   float64                       `json:"calibrated_confidence"`
	ConfidenceRationale    string                        `json:"confidence_rationale"`
	Title                  string                        `json:"title"`
	Description            string                        `json:"description"`
	RootCause              string                        `json:"root_cause"`
	RootCauseConfidence    float64                       `json:"root_cause_confidence"`
	AffectedComponents     []string                      `json:"affected_components"`
	ImpactScope            string                        `json:"impact_scope"`
	PotentialConsequences  []string                      `json:"potential_consequences"`
	ExpectedBenefit        string                        `json:"expected_benefit"`
	RemediationComplexity  string                        `json:"remediation_complexity"` // "LOW", "MEDIUM", "HIGH"
	RegressionRisk         repo_intel.Severity           `json:"regression_risk"`
	RemediationReadiness   RemediationReadiness          `json:"remediation_readiness"`
	ReadinessRationale     string                        `json:"readiness_rationale"`
	Evidence               []repo_intel.EvidenceLocation `json:"evidence"`
	RecommendedAction      string                        `json:"recommended_action"`
	ValidationRequirements []string                      `json:"validation_requirements"`
	RelatedFindingIDs      []string                      `json:"related_finding_ids,omitempty"`
	DetectionMethod        string                        `json:"detection_method"`
}

// ValidationSummary tallies counts across validation states.
type ValidationSummary struct {
	TotalFindings          int `json:"total_findings"`
	Validated              int `json:"validated"`
	PartiallyValidated     int `json:"partially_validated"`
	Stale                  int `json:"stale"`
	Invalid                int `json:"invalid"`
	RequiresHumanReview    int `json:"requires_human_review"`
}

// AnalysisMetadata stores telemetry and execution parameters for the diagnostic run.
type AnalysisMetadata struct {
	Timestamp            time.Time `json:"timestamp"`
	FilesDiscovered      int       `json:"files_discovered"`
	FilesAnalysed        int       `json:"files_analysed"`
	AnalysisCoverage     float64   `json:"analysis_coverage"`
	AnalysisWarnings     []string  `json:"analysis_warnings,omitempty"`
	ExecutionDurationMs  int64     `json:"execution_duration_ms"`
}

// DiagnosticReport is the master contract for Phase 3 autonomous remediation.
type DiagnosticReport struct {
	ReportSchemaVersion   string                               `json:"report_schema_version"`
	ReportID              string                               `json:"report_id"`
	Repository            repo_intel.RepositoryMetadataRegistry `json:"repository"`
	Analysis              AnalysisMetadata                     `json:"analysis"`
	Summary               ValidationSummary                    `json:"summary"`
	Findings              []ValidatedFinding                   `json:"findings"`
	Correlations          []FindingCorrelation                 `json:"correlations"`
	GlobalRisks           []string                             `json:"global_risks"`
	RecommendedPriorities []string                             `json:"recommended_priorities"`
	Limitations           []string                             `json:"limitations"`
}

// FindingValidationResult stores the output of Module A.
type FindingValidationResult struct {
	Summary  ValidationSummary  `json:"summary"`
	Findings []ValidatedFinding `json:"findings"`
}

// ToJSON serializes DiagnosticReport to pretty JSON.
func (dr *DiagnosticReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(dr, "", "  ")
}
