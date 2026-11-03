package benchmark

import (
	"path/filepath"
	"testing"
)

func TestGenerateBenchmarkSuiteV2(t *testing.T) {
	baseDir := filepath.Join("..", "..", "..", "..", "benchmarks", "self_healing_suite_v2")
	manifestPath := filepath.Join("..", "..", "..", "..", "benchmarks", "ground_truth_manifest_v2.json")

	manifest, err := GenerateBenchmarkSuiteV2(baseDir, manifestPath)
	if err != nil {
		t.Fatalf("Failed to generate suite v2: %v", err)
	}

	if manifest.DefectFixturesCount < 60 {
		t.Errorf("Expected at least 60 defect fixtures, got %d", manifest.DefectFixturesCount)
	}
	if manifest.CleanFixturesCount < 30 {
		t.Errorf("Expected at least 30 clean controls, got %d", manifest.CleanFixturesCount)
	}
	if manifest.TotalFixtures != manifest.DefectFixturesCount+manifest.CleanFixturesCount {
		t.Errorf("Total fixtures mismatch: %d != %d + %d", manifest.TotalFixtures, manifest.DefectFixturesCount, manifest.CleanFixturesCount)
	}
}

