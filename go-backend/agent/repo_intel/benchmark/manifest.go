package benchmark

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// LoadGroundTruthManifest loads and cryptographically validates the ground-truth benchmark manifest.
func LoadGroundTruthManifest(filePath string) (*GroundTruthManifest, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read ground truth manifest from %s: %w", filePath, err)
	}

	hashBytes := sha256.Sum256(data)
	computedHash := hex.EncodeToString(hashBytes[:])

	var manifest GroundTruthManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, computedHash, fmt.Errorf("invalid ground truth manifest JSON: %w", err)
	}

	manifest.ManifestSHA256 = computedHash

	// Basic validation
	if manifest.BenchmarkName == "" {
		return nil, computedHash, fmt.Errorf("missing benchmark_name in manifest")
	}
	if len(manifest.DefectFixtures) == 0 {
		return nil, computedHash, fmt.Errorf("manifest contains 0 defect fixtures")
	}
	if len(manifest.CleanFixtures) == 0 {
		return nil, computedHash, fmt.Errorf("manifest contains 0 clean control fixtures")
	}

	// Validate fixture counts and ensure uniqueness
	seenIDs := make(map[string]bool)
	for _, f := range manifest.DefectFixtures {
		if f.FixtureID == "" {
			return nil, computedHash, fmt.Errorf("defect fixture has empty fixture_id")
		}
		if seenIDs[f.FixtureID] {
			return nil, computedHash, fmt.Errorf("duplicate fixture_id detected: %s", f.FixtureID)
		}
		seenIDs[f.FixtureID] = true
	}

	for _, c := range manifest.CleanFixtures {
		if c.FixtureID == "" {
			return nil, computedHash, fmt.Errorf("clean fixture has empty fixture_id")
		}
		if seenIDs[c.FixtureID] {
			return nil, computedHash, fmt.Errorf("duplicate fixture_id detected: %s", c.FixtureID)
		}
		seenIDs[c.FixtureID] = true
	}

	return &manifest, computedHash, nil
}

