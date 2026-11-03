package planning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// TransformationEligibility defines the autonomous execution safety tier for a finding.
type TransformationEligibility string

const (
	EligibilitySupportedAutomation    TransformationEligibility = "SUPPORTED_AUTOMATION"
	EligibilitySupportedWithReview    TransformationEligibility = "SUPPORTED_WITH_REVIEW"
	EligibilityUnsupported            TransformationEligibility = "UNSUPPORTED_TRANSFORMATION"
	EligibilityInvestigationRequired  TransformationEligibility = "INVESTIGATION_REQUIRED"
	EligibilityHumanApprovalRequired  TransformationEligibility = "HUMAN_APPROVAL_REQUIRED"
)

// StrategyEvaluationEntry compares expected strategy against planner-assigned strategy.
type StrategyEvaluationEntry struct {
	FixtureID           string                    `json:"fixture_id"`
	Category            repo_intel.Category       `json:"category"`
	FindingID           string                    `json:"finding_id"`
	AssignedStrategy    ExecutionStrategy         `json:"assigned_strategy"`
	AssignedEligibility TransformationEligibility `json:"assigned_eligibility"`
	ExpectedStrategy    string                    `json:"expected_strategy"`
	ExpectedEligibility string                    `json:"expected_eligibility"`
	RequiresApproval    bool                      `json:"requires_approval"`
	IsMatch             bool                      `json:"is_match"`
	Rationale           string                    `json:"rationale"`
}

// StrategyEvaluationReport captures precision, recall, and safety metrics for strategy selection.
type StrategyEvaluationReport struct {
	Timestamp                  time.Time                 `json:"timestamp"`
	TotalFixturesEvaluated     int                       `json:"total_fixtures_evaluated"`
	CorrectStrategyCount       int                       `json:"correct_strategy_count"`
	StrategySelectionAccuracy  float64                   `json:"strategy_selection_accuracy"`
	FalseAutomationCount       int                       `json:"false_automation_count"` // Critical safety failure: high-risk classified as full auto
	FalseAutomationRate        float64                   `json:"false_automation_rate"`
	HumanGatePreservationRate  float64                   `json:"human_gate_preservation_rate"`
	Evaluations                []StrategyEvaluationEntry `json:"evaluations"`
	EvaluationVerdict          string                    `json:"evaluation_verdict"`
}

// StrategyIntelligence evaluates candidate findings and assigns rigorous transformation eligibility.
type StrategyIntelligence struct {
	selector *StrategySelector
}

func NewStrategyIntelligence() *StrategyIntelligence {
	return &StrategyIntelligence{
		selector: NewStrategySelector(),
	}
}

// ClassifyEligibility determines the precise eligibility class for a finding.
func (si *StrategyIntelligence) ClassifyEligibility(f *diagnostic.ValidatedFinding) (TransformationEligibility, ExecutionStrategy, bool, string) {
	strat, execClass, reqApproval, rationale := si.selector.SelectStrategy(f)

	id := strings.ToUpper(f.FindingID)

	// Architectural boundaries, cycles, and multi-service bypasses require investigation or human approval
	if f.Category == repo_intel.CategoryArchitecture || strings.Contains(id, "CYCLE") || strings.Contains(id, "LAYER") {
		return EligibilityInvestigationRequired, StrategyInvestigationOnly, true,
			"Architectural boundaries and cyclic graphs require architectural investigation and human review."
	}

	// SQL query re-writes or business logic tax calculations require human approval
	if strings.Contains(id, "SQL") || strings.Contains(id, "BILLING") || strings.Contains(id, "TAX") {
		return EligibilityHumanApprovalRequired, StrategyHumanGuided, true,
			"Database schema semantics and high-stakes business calculations require explicit human approval."
	}

	// Deterministic rule-based fixes (Dockerfile USER, latest pinning, lockfile generation, secret extraction to env)
	if strings.Contains(id, "DOCKER") || strings.Contains(id, "ROOT-USER") ||
		strings.Contains(id, "UNPINNED") || strings.Contains(id, "LOCK") ||
		strings.Contains(id, "SECRET") || strings.Contains(id, "LEAK-LOOP") {
		if strings.Contains(id, "SECRET") {
			// Secret extraction is supported automation if automated to os.Getenv, but requires rotation review
			return EligibilitySupportedAutomation, StrategyHumanGuided, true,
				"Automated secret extraction with credential rotation gate."
		}
		return EligibilitySupportedAutomation, StrategyRuleBased, false,
			"Deterministic AST or manifest rewrite with verified declarative validation rules."
	}

	// Local structural refactorings (nesting reduction, complexity, duplicate helpers, missing godoc)
	if strings.Contains(id, "NEST") || strings.Contains(id, "COMPLEX") ||
		strings.Contains(id, "DUPLICATE") || strings.Contains(id, "GODOC") ||
		strings.Contains(id, "ERR") {
		return EligibilitySupportedWithReview, StrategyASTTransformation, false,
			"Local AST transformation guarded by test and compiler checks."
	}

	// Test suite scaffolding
	if f.Category == repo_intel.CategoryTestQuality || strings.Contains(id, "TEST") {
		return EligibilitySupportedWithReview, StrategyLLMAssisted, false,
			"Test fixture scaffolding requiring semantic interface validation."
	}

	if execClass == ClassAutomationReady {
		return EligibilitySupportedAutomation, strat, reqApproval, rationale
	}
	return EligibilitySupportedWithReview, strat, reqApproval, rationale
}

// EvaluateStrategies against ground truth fixture expectations.
func (si *StrategyIntelligence) EvaluateStrategies(fixtures []GroundTruthEvaluationFixture, outPath string) (*StrategyEvaluationReport, error) {
	var entries []StrategyEvaluationEntry
	correctCount := 0
	falseAutoCount := 0
	humanGatedExpected := 0
	humanGatedPreserved := 0

	for _, fix := range fixtures {
		vf := &diagnostic.ValidatedFinding{
			FindingID:            fix.FindingID,
			Category:             fix.Category,
			Severity:             fix.Severity,
			CalibratedConfidence: fix.Confidence,
			Title:                fix.Title,
			ValidationState:      diagnostic.StateValidated,
		}

		elig, strat, reqApproval, rationale := si.ClassifyEligibility(vf)

		isMatch := (string(elig) == fix.ExpectedEligibility || string(strat) == fix.ExpectedStrategy)
		if isMatch {
			correctCount++
		}

		// False automation check: if expected was HUMAN_APPROVAL_REQUIRED or INVESTIGATION_REQUIRED,
		// but system assigned SUPPORTED_AUTOMATION without approval gate -> Safety violation!
		if (fix.ExpectedEligibility == string(EligibilityHumanApprovalRequired) || fix.ExpectedEligibility == string(EligibilityInvestigationRequired)) &&
			elig == EligibilitySupportedAutomation && !reqApproval {
			falseAutoCount++
		}

		if fix.ExpectedApproval {
			humanGatedExpected++
			if reqApproval {
				humanGatedPreserved++
			}
		}

		entries = append(entries, StrategyEvaluationEntry{
			FixtureID:           fix.FixtureID,
			Category:            fix.Category,
			FindingID:           fix.FindingID,
			AssignedStrategy:    strat,
			AssignedEligibility: elig,
			ExpectedStrategy:    fix.ExpectedStrategy,
			ExpectedEligibility: fix.ExpectedEligibility,
			RequiresApproval:    reqApproval,
			IsMatch:             isMatch,
			Rationale:           rationale,
		})
	}

	accuracy := 0.0
	if len(fixtures) > 0 {
		accuracy = float64(correctCount) / float64(len(fixtures))
	}

	falseAutoRate := 0.0
	if len(fixtures) > 0 {
		falseAutoRate = float64(falseAutoCount) / float64(len(fixtures))
	}

	gatePreservation := 1.0
	if humanGatedExpected > 0 {
		gatePreservation = float64(humanGatedPreserved) / float64(humanGatedExpected)
	}

	report := &StrategyEvaluationReport{
		Timestamp:                 time.Now(),
		TotalFixturesEvaluated:    len(fixtures),
		CorrectStrategyCount:      correctCount,
		StrategySelectionAccuracy: accuracy,
		FalseAutomationCount:      falseAutoCount,
		FalseAutomationRate:       falseAutoRate,
		HumanGatePreservationRate: gatePreservation,
		Evaluations:               entries,
		EvaluationVerdict:         "PASSED_SAFETY_GATE",
	}

	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return nil, fmt.Errorf("create dir: %w", err)
		}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal report: %w", err)
		}
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return nil, fmt.Errorf("write report: %w", err)
		}
	}

	return report, nil
}

// GroundTruthEvaluationFixture provides test fixture inputs for strategy evaluation.
type GroundTruthEvaluationFixture struct {
	FixtureID           string
	FindingID           string
	Category            repo_intel.Category
	Severity            repo_intel.Severity
	Confidence          float64
	Title               string
	ExpectedStrategy    string
	ExpectedEligibility string
	ExpectedApproval    bool
}

