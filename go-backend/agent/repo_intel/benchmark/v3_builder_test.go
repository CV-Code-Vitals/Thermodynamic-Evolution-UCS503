package benchmark

import (
	"path/filepath"
	"testing"
)

func TestGenerateBenchmarkSuiteV3(t *testing.T) {
	tempBase := t.TempDir()
	tempManifest := filepath.Join(tempBase, "ground_truth_manifest_v3.json")

	manifest, err := GenerateBenchmarkSuiteV3(tempBase, tempManifest)
	if err != nil {
		t.Fatalf("GenerateBenchmarkSuiteV3 failed: %v", err)
	}

	if manifest.TotalFixtures != 144 {
		t.Errorf("Expected 144 total fixtures, got %d", manifest.TotalFixtures)
	}
	if len(manifest.DefectFixtures) != 96 {
		t.Errorf("Expected 96 defect fixtures, got %d", len(manifest.DefectFixtures))
	}
	if len(manifest.CleanFixtures) != 48 {
		t.Errorf("Expected 48 clean fixtures, got %d", len(manifest.CleanFixtures))
	}
}
