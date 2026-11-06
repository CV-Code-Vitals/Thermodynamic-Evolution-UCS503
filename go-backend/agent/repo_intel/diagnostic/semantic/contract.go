package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// AnalysisCompleteness classifies whether the AST, CFG, and call graph analysis was fully bounded or truncated.
type AnalysisCompleteness string

const (
	CompletenessFull        AnalysisCompleteness = "COMPLETE"
	CompletenessPartialCall AnalysisCompleteness = "PARTIAL_CALL_GRAPH_TRUNCATED"
	CompletenessIncomplete  AnalysisCompleteness = "INCOMPLETE_DEPENDENCIES"
)

// UncertaintyClassification explicitly models semantic confidence boundaries.
type UncertaintyClassification string

const (
	UncertaintyDeterministic UncertaintyClassification = "DETERMINISTIC"
	UncertaintyBounded       UncertaintyClassification = "BOUNDED_UNCERTAINTY"
	UncertaintyPossibleEscape UncertaintyClassification = "POSSIBLE_ESCAPE"
	UncertaintyConcurrent   UncertaintyClassification = "UNKNOWN_CONCURRENT_OWNERSHIP"
)

// SemanticEvidenceTrace preserves the rich interprocedural and structural context of a finding.
type SemanticEvidenceTrace struct {
	CallPath         []string `json:"call_path,omitempty"`
	CFGNodeIDs       []int    `json:"cfg_node_ids,omitempty"`
	DataflowHops     int      `json:"dataflow_hops"`
	AccessPath       string   `json:"access_path,omitempty"`
	WrapperContract  string   `json:"wrapper_contract,omitempty"`
	Aliases          []string `json:"aliases,omitempty"`
	AliasHopDepth    int      `json:"alias_hop_depth"`
	ConcurrencyState string   `json:"concurrency_state,omitempty"`
}

// SemanticFindingContract is the versioned (2.0-semantic) specification bridging semantic diagnostics and planning.
type SemanticFindingContract struct {
	SchemaVersion          string                    `json:"schema_version"` // "2.0-semantic"
	FindingID              string                    `json:"finding_id"`
	RuleID                 string                    `json:"rule_id"`
	SourceCommit           string                    `json:"source_commit"`
	FileSHA256             string                    `json:"file_sha256"`
	TargetFile             string                    `json:"target_file"`
	LineStart              int                       `json:"line_start"`
	LineEnd                int                       `json:"line_end"`
	Category               repo_intel.Category       `json:"category"`
	Severity               repo_intel.Severity       `json:"severity"`
	Confidence             float64                   `json:"confidence"`
	AnalysisCompleteness   AnalysisCompleteness      `json:"analysis_completeness"`
	UncertaintyClass       UncertaintyClassification `json:"uncertainty_class"`
	SemanticTrace          SemanticEvidenceTrace     `json:"semantic_trace"`
	RemediationEligibility RemediationEligibility    `json:"remediation_eligibility"`
	HumanApprovalRequired  bool                      `json:"human_approval_required"`
	ApprovalRationale      string                    `json:"approval_rationale,omitempty"`
	ValidationRequirements []string                  `json:"validation_requirements"`
	CounterEvidences       []string                  `json:"counter_evidences,omitempty"`
	Title                  string                    `json:"title"`
	Description            string                    `json:"description"`
	RootCause              string                    `json:"root_cause"`
	RecommendedAction      string                    `json:"recommended_action"`
	GeneratedAt            time.Time                 `json:"generated_at"`
}

// ComputeFileSHA256 computes the bitwise SHA256 hash of a file on disk.
func ComputeFileSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// ValidateIntegrity verifies file hash against disk and enforces fail-closed uncertainty invariants.
func (sfc *SemanticFindingContract) ValidateIntegrity(repoRoot string) (bool, string) {
	// 1. Verify file hash on disk
	fullPath := filepath.Join(repoRoot, filepath.FromSlash(sfc.TargetFile))
	currentHash, err := ComputeFileSHA256(fullPath)
	if err != nil {
		return false, fmt.Sprintf("FAIL_CLOSED: cannot read target file %s: %v", sfc.TargetFile, err)
	}
	if sfc.FileSHA256 != "" && currentHash != sfc.FileSHA256 {
		return false, fmt.Sprintf("FAIL_CLOSED: stale finding for %s (hash mismatch: expected %.8s, got %.8s)",
			sfc.TargetFile, sfc.FileSHA256, currentHash)
	}

	// 2. Fail closed on uncertain concurrency
	if sfc.UncertaintyClass == UncertaintyPossibleEscape || sfc.UncertaintyClass == UncertaintyConcurrent {
		if sfc.RemediationEligibility == EligibleAutonomous {
			return false, fmt.Sprintf("FAIL_CLOSED: finding %s has uncertain concurrency (%s) but was labeled autonomous",
				sfc.FindingID, sfc.UncertaintyClass)
		}
	}

	// 3. Fail closed on truncated call graphs
	if sfc.AnalysisCompleteness != CompletenessFull {
		if sfc.RemediationEligibility == EligibleAutonomous {
			return false, fmt.Sprintf("FAIL_CLOSED: finding %s has incomplete analysis (%s) but was labeled autonomous",
				sfc.FindingID, sfc.AnalysisCompleteness)
		}
	}

	return true, ""
}

// ToValidatedFinding converts the SemanticFindingContract into a diagnostic.ValidatedFinding,
// bridging the semantic engine to the RemediationPlanner without loss of governance or provenance.
func (sfc *SemanticFindingContract) ToValidatedFinding() diagnostic.ValidatedFinding {
	// Map validation state
	valState := diagnostic.StateValidated
	if sfc.AnalysisCompleteness != CompletenessFull {
		valState = diagnostic.StatePartiallyValidated
	}
	if sfc.HumanApprovalRequired {
		valState = diagnostic.StateRequiresHumanReview
	}

	// Map remediation readiness
	readiness := diagnostic.ReadinessAutomationReady
	switch sfc.RemediationEligibility {
	case EligibleAutonomous:
		readiness = diagnostic.ReadinessAutomationReady
	case EligibleHumanApproval:
		readiness = diagnostic.ReadinessHumanApprovalRequired
	case EligibleQuarantined:
		readiness = diagnostic.ReadinessInvestigationRequired
	case EligibleBenign:
		readiness = diagnostic.ReadinessInformationalOnly
	}

	// Encode semantic trace into description / location metadata
	traceDesc := fmt.Sprintf("Rule: %s | Completeness: %s | Uncertainty: %s | AccessPath: %s | Concurrency: %s",
		sfc.RuleID, sfc.AnalysisCompleteness, sfc.UncertaintyClass, sfc.SemanticTrace.AccessPath, sfc.SemanticTrace.ConcurrencyState)

	return diagnostic.ValidatedFinding{
		FindingID:             sfc.FindingID,
		SourceFindingID:       sfc.FindingID,
		ValidationState:       valState,
		ValidationReason:      fmt.Sprintf("Semantic analysis verified under contract %s", sfc.SchemaVersion),
		Category:              sfc.Category,
		Severity:              sfc.Severity,
		SeverityRationale:     fmt.Sprintf("Calculated by %s analyzer", sfc.RuleID),
		CalibratedConfidence:  sfc.Confidence,
		ConfidenceRationale:   fmt.Sprintf("Derived from semantic AST/CFG evidence under %s certainty", sfc.UncertaintyClass),
		Title:                 sfc.Title,
		Description:           sfc.Description,
		RootCause:             sfc.RootCause,
		RootCauseConfidence:   sfc.Confidence,
		AffectedComponents:    []string{sfc.TargetFile},
		ImpactScope:           "FUNCTION_LOCAL",
		RemediationComplexity: "MEDIUM",
		RegressionRisk:        repo_intel.SeverityLow,
		RemediationReadiness:  readiness,
		ReadinessRationale:    sfc.ApprovalRationale,
		Evidence: []repo_intel.EvidenceLocation{
			{
				File:        sfc.TargetFile,
				LineStart:   sfc.LineStart,
				LineEnd:     sfc.LineEnd,
				Description: traceDesc,
			},
		},
		RecommendedAction:      sfc.RecommendedAction,
		ValidationRequirements: sfc.ValidationRequirements,
		DetectionMethod:        "SEMANTIC_INTERPROCEDURAL",
	}
}

// BuildDiagnosticReportFromContracts converts a collection of SemanticFindingContracts into a diagnostic.DiagnosticReport.
func BuildDiagnosticReportFromContracts(
	repoRoot string,
	repoCommit string,
	contracts []SemanticFindingContract,
) *diagnostic.DiagnosticReport {
	var findings []diagnostic.ValidatedFinding
	valCount := 0
	reviewCount := 0

	for _, c := range contracts {
		vf := c.ToValidatedFinding()
		findings = append(findings, vf)
		if vf.RemediationReadiness == diagnostic.ReadinessHumanApprovalRequired {
			reviewCount++
		} else if vf.RemediationReadiness == diagnostic.ReadinessAutomationReady {
			valCount++
		}
	}

	return &diagnostic.DiagnosticReport{
		ReportSchemaVersion: "1.0",
		ReportID:            fmt.Sprintf("DIAG-SEM-%d", time.Now().UnixNano()),
		Repository: repo_intel.RepositoryMetadataRegistry{
			RepoRoot:   repoRoot,
			HeadCommit: repoCommit,
		},
		Analysis: diagnostic.AnalysisMetadata{
			Timestamp:        time.Now(),
			FilesDiscovered:  len(findings),
			FilesAnalysed:    len(findings),
			AnalysisCoverage: 1.0,
		},
		Summary: diagnostic.ValidationSummary{
			TotalFindings:       len(findings),
			Validated:           valCount,
			RequiresHumanReview: reviewCount,
		},
		Findings: findings,
	}
}
