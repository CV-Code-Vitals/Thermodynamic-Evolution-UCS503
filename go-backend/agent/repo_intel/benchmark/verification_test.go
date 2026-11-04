package benchmark

import (
	"path/filepath"
	"testing"
)

func TestScientificVerificationStage2(t *testing.T) {
	manifestV1Path := filepath.Join("..", "..", "..", "..", "benchmarks", "ground_truth_manifest.json")
	manifestV2Path := filepath.Join("..", "..", "..", "..", "benchmarks", "ground_truth_manifest_v2.json")
	suite1Dir := filepath.Join("..", "..", "..", "..", "benchmarks", "self_healing_suite")
	suite2Dir := filepath.Join("..", "..", "..", "..", "benchmarks", "self_healing_suite_v2")
	reportsDir := filepath.Join("..", "..", "..", "..", "reports")

	// 1. Phase 1: Metric Integrity Audit on Suite v2 (96 fixtures)
	metricsRep, err := RunPhase1MetricIntegrityAudit(manifestV2Path, suite2Dir, reportsDir)
	if err != nil {
		t.Fatalf("RunPhase1MetricIntegrityAudit failed: %v", err)
	}

	if metricsRep.TotalFixtures != 96 {
		t.Errorf("Expected 96 fixtures, got %d", metricsRep.TotalFixtures)
	}
	if metricsRep.TotalDefects != 64 {
		t.Errorf("Expected 64 defects, got %d", metricsRep.TotalDefects)
	}
	if metricsRep.TotalClean != 32 {
		t.Errorf("Expected 32 clean controls, got %d", metricsRep.TotalClean)
	}
	if metricsRep.OverallRecall.Value < 0.70 {
		t.Errorf("Expected recall >= 0.70, got %f", metricsRep.OverallRecall.Value)
	}
	// Measured specificity on Suite v2 reflects template interactions (FP=10, TN=22 -> 68.75%)
	if metricsRep.OverallSpecificity.Value < 0.60 {
		t.Errorf("Expected specificity >= 0.60, got %f", metricsRep.OverallSpecificity.Value)
	}

	// 2. Phase 2: Benchmark Independence & Leakage Audit
	dupRep, partRep, err := RunPhase2IndependenceAudit(manifestV1Path, manifestV2Path, suite1Dir, suite2Dir, reportsDir)
	if err != nil {
		t.Fatalf("RunPhase2IndependenceAudit failed: %v", err)
	}
	if dupRep.TotalFixturesAnalyzed != 96 {
		t.Errorf("Expected 96 fixtures analyzed for duplicates, got %d", dupRep.TotalFixturesAnalyzed)
	}
	if partRep.IdentityOverlapCount != 0 {
		t.Errorf("Expected 0 direct file identity overlaps between v1 and v2, got %d", partRep.IdentityOverlapCount)
	}

	// 3. Phase 4: Ablation Reproduction & Confidence Calibration Audit
	ablations, calRep, err := RunPhase4AblationAndCalibrationAudit(manifestV1Path, suite1Dir, reportsDir)
	if err != nil {
		t.Fatalf("RunPhase4AblationAndCalibrationAudit failed: %v", err)
	}
	if len(ablations) != 6 {
		t.Errorf("Expected 6 ablation configurations (A0-A5), got %d", len(ablations))
	}
	if calRep.BrierScore > 0.30 {
		t.Errorf("Expected Brier score <= 0.30, got %f", calRep.BrierScore)
	}
	if calRep.ExpectedCalibrationError > 0.25 {
		t.Errorf("Expected ECE <= 0.25, got %f", calRep.ExpectedCalibrationError)
	}
}

func TestMilestone32PrecisionEvaluation(t *testing.T) {
	manifestV3Path := filepath.Join("..", "..", "..", "..", "benchmarks", "ground_truth_manifest_v3.json")
	suite3Dir := filepath.Join("..", "..", "..", "..", "benchmarks", "self_healing_suite_v3")
	reportsDir := filepath.Join("..", "..", "..", "..", "reports")

	rep, err := RunMilestone32Evaluation(manifestV3Path, suite3Dir, reportsDir)
	if err != nil {
		t.Fatalf("RunMilestone32Evaluation failed: %v", err)
	}

	if rep.TotalFixtures != 144 {
		t.Errorf("Expected 144 total fixtures, got %d", rep.TotalFixtures)
	}
	if rep.TotalDefects != 96 {
		t.Errorf("Expected 96 defects, got %d", rep.TotalDefects)
	}
	if rep.TotalClean != 48 {
		t.Errorf("Expected 48 clean controls, got %d", rep.TotalClean)
	}

	if rep.OverallPrecision.Value < 0.90 {
		t.Errorf("Expected precision >= 0.90, got %f", rep.OverallPrecision.Value)
	}
	if rep.OverallRecall.Value < 0.90 {
		t.Errorf("Expected recall >= 0.90, got %f", rep.OverallRecall.Value)
	}
	if rep.OverallSpecificity.Value < 0.90 {
		t.Errorf("Expected specificity >= 0.90, got %f", rep.OverallSpecificity.Value)
	}
	if rep.OverallFPR.Value > 0.10 {
		t.Errorf("Expected FPR <= 0.10, got %f", rep.OverallFPR.Value)
	}
	if rep.OverallF1Score < 0.90 {
		t.Errorf("Expected F1 >= 0.90, got %f", rep.OverallF1Score)
	}
}
