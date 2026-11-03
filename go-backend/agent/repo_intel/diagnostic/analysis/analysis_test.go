package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnosticErrorAnalysis(t *testing.T) {
	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_0_diagnostic_error_analysis.md")
	report, err := RunDiagnosticErrorAnalysis(outPath)
	if err != nil {
		t.Fatalf("Failed running error analysis: %v", err)
	}

	if report.TotalErrorsAnalyzed != 11 {
		t.Errorf("Expected 11 errors analyzed (4 FP + 7 FN), got %d", report.TotalErrorsAnalyzed)
	}
	if report.FalsePositiveCount != 4 {
		t.Errorf("Expected 4 FPs, got %d", report.FalsePositiveCount)
	}
	if report.FalseNegativeCount != 7 {
		t.Errorf("Expected 7 FNs, got %d", report.FalseNegativeCount)
	}

	// Verify report file was created
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("Expected report file %s to exist", outPath)
	}
}
