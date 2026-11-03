package rootcause

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

func TestRootCauseClustering(t *testing.T) {
	findings := []repo_intel.Finding{
		{
			FindingID:     "F1-NEST",
			Category:      repo_intel.CategoryStructural,
			Confidence:    0.90,
			Title:         "Excessive nesting depth",
			AffectedFiles: []string{"pkg/process.go"},
		},
		{
			FindingID:     "F2-COMPLEX",
			Category:      repo_intel.CategoryStructural,
			Confidence:    0.85,
			Title:         "High cyclomatic complexity",
			AffectedFiles: []string{"pkg/process.go"},
		},
		{
			FindingID:     "F3-SECRET",
			Category:      repo_intel.CategorySecurity,
			Confidence:    0.95,
			Title:         "Hardcoded secret",
			AffectedFiles: []string{"config/secrets.go"},
		},
		{
			FindingID:     "F4-RESOURCE",
			Category:      repo_intel.CategoryReliability,
			Confidence:    0.92,
			Title:         "Leaked file handle in loop",
			AffectedFiles: []string{"io/files.go"},
		},
	}

	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_0_root_cause_evaluation.json")
	eval := NewRootCauseEvaluator()
	rep, err := eval.EvaluateAndReport(findings, outPath)
	if err != nil {
		t.Fatalf("EvaluateAndReport failed: %v", err)
	}

	if rep.TotalFindings != 4 {
		t.Errorf("Expected 4 findings, got %d", rep.TotalFindings)
	}
	if rep.RootCausesIdentified != 3 {
		t.Errorf("Expected 3 root causes (F1+F2 clustered), got %d", rep.RootCausesIdentified)
	}
	if rep.CompressionRatio <= 1.0 {
		t.Errorf("Expected compression ratio > 1.0, got %f", rep.CompressionRatio)
	}

	// Verify report file was written
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Expected report file %s to exist", outPath)
	}
}

