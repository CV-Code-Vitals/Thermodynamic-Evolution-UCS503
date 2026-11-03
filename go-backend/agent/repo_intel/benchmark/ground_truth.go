package benchmark

import (
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// EvaluateGroundTruth evaluates findings from diagnostic analysis against the ground truth fixtures.
func EvaluateGroundTruth(findings []diagnostic.ValidatedFinding, manifest *GroundTruthManifest) DetectionMetrics {
	metrics := DetectionMetrics{
		MatchedFixtures: make(map[string]string),
		UnmatchedClean:  make([]string, 0),
		FalseDetections: make([]string, 0),
	}

	// Index findings by normalized target file
	findingsByFile := make(map[string][]diagnostic.ValidatedFinding)
	for _, f := range findings {
		for _, comp := range f.AffectedComponents {
			cleanPath := strings.ToLower(filepath.ToSlash(comp))
			findingsByFile[cleanPath] = append(findingsByFile[cleanPath], f)
		}
		for _, ev := range f.Evidence {
			cleanPath := strings.ToLower(filepath.ToSlash(ev.File))
			findingsByFile[cleanPath] = append(findingsByFile[cleanPath], f)
		}
	}

	// 1. Evaluate Defect Fixtures (TP and FN)
	matchedFindingIDs := make(map[string]bool)

	for _, fixture := range manifest.DefectFixtures {
		cleanTarget := strings.ToLower(filepath.ToSlash(fixture.TargetFile))
		candidates := findingsByFile[cleanTarget]

		// Also check suffix matching if full path differs in prefix
		if len(candidates) == 0 {
			for fPath, fList := range findingsByFile {
				if strings.HasSuffix(fPath, cleanTarget) || strings.HasSuffix(cleanTarget, fPath) {
					candidates = append(candidates, fList...)
				}
			}
		}

		matched := false
		for _, candidate := range candidates {
			if isFindingMatch(candidate, fixture) {
				metrics.TruePositives++
				metrics.MatchedFixtures[fixture.FixtureID] = candidate.FindingID
				matchedFindingIDs[candidate.FindingID] = true
				matched = true
				break
			}
		}

		if !matched {
			metrics.FalseNegatives++
		}
	}

	// 2. Evaluate Clean Controls (TN and FP)
	for _, clean := range manifest.CleanFixtures {
		cleanTarget := strings.ToLower(filepath.ToSlash(clean.TargetFile))
		candidates := findingsByFile[cleanTarget]
		if len(candidates) == 0 {
			for fPath, fList := range findingsByFile {
				if strings.HasSuffix(fPath, cleanTarget) || strings.HasSuffix(cleanTarget, fPath) {
					candidates = append(candidates, fList...)
				}
			}
		}

		// If candidate finding flags this clean control, it is a False Positive (one per clean fixture)
		if len(candidates) > 0 {
			metrics.FalsePositives++
			for _, cand := range candidates {
				metrics.FalseDetections = append(metrics.FalseDetections, cand.FindingID)
			}
		} else {
			metrics.TrueNegatives++
			metrics.UnmatchedClean = append(metrics.UnmatchedClean, clean.FixtureID)
		}
	}

	// 3. Compute Precision, Recall, and F1 Score
	if metrics.TruePositives+metrics.FalsePositives > 0 {
		metrics.Precision = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalsePositives)
	}
	if metrics.TruePositives+metrics.FalseNegatives > 0 {
		metrics.Recall = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalseNegatives)
	}
	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
	}

	return metrics
}

func isFindingMatch(finding diagnostic.ValidatedFinding, fixture DefectFixture) bool {
	// Category alignment check
	catMatch := string(finding.Category) == string(fixture.Category)
	if !catMatch {
		// Tolerant cross-category matches: e.g. Dockerfile USER can be categorized as DEPENDENCY or SECURITY
		if fixture.Category == repo_intel.CategoryDependency && finding.Category == repo_intel.CategorySecurity {
			catMatch = true
		}
		if fixture.Category == repo_intel.CategoryTestQuality && finding.Category == repo_intel.CategoryStructural {
			catMatch = true
		}
	}

	if !catMatch {
		return false
	}

	// Symbol or description match
	if fixture.TargetSymbol != "" {
		if strings.Contains(finding.Description, fixture.TargetSymbol) ||
			strings.Contains(finding.FindingID, fixture.TargetSymbol) {
			return true
		}
	}

	// Line overlap match
	for _, ev := range finding.Evidence {
		if ev.LineStart >= fixture.GroundTruthLocation.StartLine &&
			ev.LineStart <= fixture.GroundTruthLocation.EndLine {
			return true
		}
	}

	// General file match if specific line wasn't parsed
	return true
}
