package planning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

func TestStrategyIntelligence(t *testing.T) {
	fixtures := []GroundTruthEvaluationFixture{
		{
			FixtureID:           "FIX-STR-01",
			FindingID:           "RI-STR-HIGH-NESTING-L10",
			Category:            repo_intel.CategoryStructural,
			Severity:            repo_intel.SeverityMedium,
			Confidence:          0.90,
			Title:               "Excessive nesting depth",
			ExpectedStrategy:    "AST_TRANSFORMATION",
			ExpectedEligibility: "SUPPORTED_WITH_REVIEW",
			ExpectedApproval:    false,
		},
		{
			FixtureID:           "FIX-SEC-01",
			FindingID:           "RI-SEC-SECRET-L12",
			Category:            repo_intel.CategorySecurity,
			Severity:            repo_intel.SeverityCritical,
			Confidence:          0.95,
			Title:               "Hardcoded secret",
			ExpectedStrategy:    "HUMAN_GUIDED",
			ExpectedEligibility: "SUPPORTED_AUTOMATION",
			ExpectedApproval:    true,
		},
		{
			FixtureID:           "FIX-DEP-01",
			FindingID:           "RI-DEP-UNPINNED-BASE",
			Category:            repo_intel.CategoryDependency,
			Severity:            repo_intel.SeverityMedium,
			Confidence:          0.95,
			Title:               "Unpinned Dockerfile base image",
			ExpectedStrategy:    "RULE_BASED",
			ExpectedEligibility: "SUPPORTED_AUTOMATION",
			ExpectedApproval:    false,
		},
		{
			FixtureID:           "FIX-ARC-01",
			FindingID:           "CTX-ARC-LAYER-BYPASS",
			Category:            repo_intel.CategoryArchitecture,
			Severity:            repo_intel.SeverityMedium,
			Confidence:          0.93,
			Title:               "Layer boundary violation",
			ExpectedStrategy:    "INVESTIGATION_ONLY",
			ExpectedEligibility: "INVESTIGATION_REQUIRED",
			ExpectedApproval:    true,
		},
		{
			FixtureID:           "FIX-SEC-02",
			FindingID:           "CTX-SEC-SQL-INJECT",
			Category:            repo_intel.CategorySecurity,
			Severity:            repo_intel.SeverityHigh,
			Confidence:          0.96,
			Title:               "SQL query concatenation",
			ExpectedStrategy:    "HUMAN_GUIDED",
			ExpectedEligibility: "HUMAN_APPROVAL_REQUIRED",
			ExpectedApproval:    true,
		},
	}

	outPath := filepath.Join("..", "..", "..", "..", "reports", "milestone3_0_strategy_evaluation.json")
	si := NewStrategyIntelligence()
	report, err := si.EvaluateStrategies(fixtures, outPath)
	if err != nil {
		t.Fatalf("EvaluateStrategies failed: %v", err)
	}

	if report.TotalFixturesEvaluated != len(fixtures) {
		t.Errorf("Expected %d fixtures evaluated, got %d", len(fixtures), report.TotalFixturesEvaluated)
	}
	if report.FalseAutomationCount > 0 {
		t.Errorf("Critical safety failure: FalseAutomationCount = %d (must be 0)", report.FalseAutomationCount)
	}
	if report.HumanGatePreservationRate < 1.0 {
		t.Errorf("HumanGatePreservationRate must be 1.0, got %f", report.HumanGatePreservationRate)
	}
	if report.StrategySelectionAccuracy < 0.80 {
		t.Errorf("StrategySelectionAccuracy must be >= 0.80, got %f", report.StrategySelectionAccuracy)
	}

	// Verify report file on disk
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Expected report file %s to exist", outPath)
	}
}

