package benchmark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

func TestGroundTruthManifestLoader(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "..", "..", "benchmarks", "ground_truth_manifest.json")
	manifest, hash, err := LoadGroundTruthManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed loading ground truth manifest: %v", err)
	}

	if manifest.BenchmarkName == "" {
		t.Errorf("Expected non-empty benchmark name")
	}
	if len(hash) != 64 {
		t.Errorf("Expected 64-char SHA256 hash, got %d", len(hash))
	}
	if len(manifest.DefectFixtures) != 15 {
		t.Errorf("Expected 15 defect fixtures, got %d", len(manifest.DefectFixtures))
	}
	if len(manifest.CleanFixtures) != 6 {
		t.Errorf("Expected 6 clean fixtures, got %d", len(manifest.CleanFixtures))
	}
}

func TestGroundTruthEvaluationMetrics(t *testing.T) {
	manifest := &GroundTruthManifest{
		DefectFixtures: []DefectFixture{
			{
				FixtureID:  "FIX-1",
				Category:   repo_intel.CategoryStructural,
				TargetFile: "pkg/a.go",
			},
			{
				FixtureID:  "FIX-2",
				Category:   repo_intel.CategorySecurity,
				TargetFile: "pkg/b.go",
			},
		},
		CleanFixtures: []CleanFixture{
			{
				FixtureID:  "CLEAN-1",
				Category:   repo_intel.CategorySecurity,
				TargetFile: "pkg/clean.go",
			},
		},
	}

	// Case 1: Perfect detection (2 TP, 0 FP, 0 FN, 1 TN)
	findingsPerfect := []diagnostic.ValidatedFinding{
		{
			FindingID: "F-1",
			Category:  repo_intel.CategoryStructural,
			Evidence: []repo_intel.EvidenceLocation{
				{File: "pkg/a.go"},
			},
		},
		{
			FindingID: "F-2",
			Category:  repo_intel.CategorySecurity,
			Evidence: []repo_intel.EvidenceLocation{
				{File: "pkg/b.go"},
			},
		},
	}

	mPerfect := EvaluateGroundTruth(findingsPerfect, manifest)
	if mPerfect.TruePositives != 2 || mPerfect.FalsePositives != 0 || mPerfect.FalseNegatives != 0 || mPerfect.TrueNegatives != 1 {
		t.Errorf("Expected TP=2, FP=0, FN=0, TN=1, got %+v", mPerfect)
	}
	if mPerfect.Precision != 1.0 || mPerfect.Recall != 1.0 || mPerfect.F1Score != 1.0 {
		t.Errorf("Expected 1.0 for Precision, Recall, F1, got P=%.2f, R=%.2f, F1=%.2f",
			mPerfect.Precision, mPerfect.Recall, mPerfect.F1Score)
	}

	// Case 2: 1 TP, 1 FN, 1 FP (flagged clean fixture)
	findingsWithFalsePositive := []diagnostic.ValidatedFinding{
		{
			FindingID: "F-1",
			Category:  repo_intel.CategoryStructural,
			Evidence: []repo_intel.EvidenceLocation{
				{File: "pkg/a.go"},
			},
		},
		{
			FindingID: "F-FP",
			Category:  repo_intel.CategorySecurity,
			Evidence: []repo_intel.EvidenceLocation{
				{File: "pkg/clean.go"},
			},
		},
	}

	mWithFP := EvaluateGroundTruth(findingsWithFalsePositive, manifest)
	if mWithFP.TruePositives != 1 || mWithFP.FalseNegatives != 1 || mWithFP.FalsePositives != 1 || mWithFP.TrueNegatives != 0 {
		t.Errorf("Expected TP=1, FN=1, FP=1, TN=0, got %+v", mWithFP)
	}
	if mWithFP.Precision != 0.5 || mWithFP.Recall != 0.5 || mWithFP.F1Score != 0.5 {
		t.Errorf("Expected 0.5 metrics, got P=%.2f, R=%.2f, F1=%.2f",
			mWithFP.Precision, mWithFP.Recall, mWithFP.F1Score)
	}
}

func TestRemediationMetricsCalculation(t *testing.T) {
	execReport := &execution.ExecutionReport{
		Summary: execution.ExecutionSummary{
			TotalScheduled: 4,
			Completed:      3,
			Failed:         0,
			RolledBack:     1,
		},
	}

	remediatedIDs := map[string]bool{
		"TASK-1": true,
		"TASK-2": true,
	}

	metrics := CalculateRemediationMetrics(
		execReport,
		3, // 3 eligible defects
		remediatedIDs,
		0, // 0 regressions
		2, // 2 integrated
		2, // 2 approved
	)

	if metrics.VerifiedRemediations != 2 || metrics.EligibleGroundTruthDefects != 3 {
		t.Errorf("Incorrect verified remediation counts: %+v", metrics)
	}
	expectedRate := 2.0 / 3.0
	if metrics.VerifiedRemediationRate < expectedRate-0.01 || metrics.VerifiedRemediationRate > expectedRate+0.01 {
		t.Errorf("Expected rate ~%.2f, got %.2f", expectedRate, metrics.VerifiedRemediationRate)
	}
	if metrics.RegressionRate != 0.0 {
		t.Errorf("Expected 0.0 regression rate, got %.2f", metrics.RegressionRate)
	}
	if metrics.IntegrationSuccessRate != 1.0 {
		t.Errorf("Expected 1.0 integration success rate, got %.2f", metrics.IntegrationSuccessRate)
	}
}

func TestBeforeAfterComparison(t *testing.T) {
	before := &diagnostic.DiagnosticReport{
		Findings: []diagnostic.ValidatedFinding{
			{FindingID: "FIND-1", Severity: repo_intel.SeverityHigh},
			{FindingID: "FIND-2", Severity: repo_intel.SeverityMedium},
			{FindingID: "FIND-3", Severity: repo_intel.SeverityLow},
		},
	}

	after := &diagnostic.DiagnosticReport{
		Findings: []diagnostic.ValidatedFinding{
			{FindingID: "FIND-2", Severity: repo_intel.SeverityMedium}, // Still unresolved
			{FindingID: "FIND-NEW", Severity: repo_intel.SeverityLow},  // Newly introduced
		},
	}

	comp := CompareBeforeAndAfter(before, after)
	if comp.BeforeFindingsCount != 3 || comp.AfterFindingsCount != 2 {
		t.Errorf("Incorrect counts: before=%d, after=%d", comp.BeforeFindingsCount, comp.AfterFindingsCount)
	}
	if comp.NetFindingsDelta != -1 {
		t.Errorf("Expected delta -1, got %d", comp.NetFindingsDelta)
	}
	if len(comp.ResolvedFindingIDs) != 2 {
		t.Errorf("Expected 2 resolved findings (FIND-1, FIND-3), got %v", comp.ResolvedFindingIDs)
	}
	if len(comp.NewlyIntroducedFindings) != 1 || comp.NewlyIntroducedFindings[0] != "FIND-NEW" {
		t.Errorf("Expected 1 newly introduced finding, got %v", comp.NewlyIntroducedFindings)
	}
}

type failSyntaxStrategy struct{}

func (s *failSyntaxStrategy) Name() string { return "FAIL_SYNTAX" }

func (s *failSyntaxStrategy) Execute(wsDir string, task *planning.RemediationTask) ([]string, error) {
	fullPath := filepath.Join(wsDir, filepath.FromSlash(task.TargetFiles[0]))
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, _ = f.WriteString("\nthis is a deliberate syntax error !!!\n")
	return task.TargetFiles, nil
}

func TestControlledFailureRollback(t *testing.T) {
	repoRoot, _ := filepath.Abs(filepath.Join("..", "..", "..", "..", "benchmarks", "self_healing_suite"))
	execCfg := execution.ExecutionConfig{
		RepoRoot:       repoRoot,
		KeepWorkspaces: false,
	}

	execEngine, err := execution.NewRemediationExecutor(execCfg)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	execEngine.StrategyReg.Register("FAIL_SYNTAX", &failSyntaxStrategy{})

	// Task that triggers syntax and compilation failure
	task := &planning.RemediationTask{
		TaskID:            "TASK-FAIL-TEST",
		ExecutionStrategy: "FAIL_SYNTAX",
		ExecutionClass:    planning.ClassAutomationReady,
		TargetFiles:       []string{"structural/nesting.go"},
	}

	res, _, err := execEngine.ExecuteTask(task, nil)
	if err != nil {
		t.Fatalf("ExecuteTask error: %v", err)
	}
	if res.Status != execution.StatusRolledBack {
		t.Errorf("Expected status ROLLED_BACK, got %s", res.Status)
	}
	if !res.RollbackOccurred {
		t.Errorf("Expected RollbackOccurred = true")
	}
	if !strings.Contains(res.RollbackReason, "validation checks failed") {
		t.Errorf("Expected validation failure rollback reason, got %s", res.RollbackReason)
	}
}
