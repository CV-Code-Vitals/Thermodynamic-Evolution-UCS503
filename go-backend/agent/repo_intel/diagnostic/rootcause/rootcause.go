package rootcause

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// RootCauseAnalyzer clusters findings into causal groups and generates unified root-cause entities.
type RootCauseAnalyzer struct{}

func NewRootCauseAnalyzer() *RootCauseAnalyzer {
	return &RootCauseAnalyzer{}
}

// AnalyzeRootCauses constructs the causal graph and derives root-cause entities.
func (a *RootCauseAnalyzer) AnalyzeRootCauses(findings []repo_intel.Finding) ([]RootCauseEntity, *RootCauseGraph) {
	graph := NewRootCauseGraph()

	// 1. Add all findings as nodes
	for i := range findings {
		f := &findings[i]
		primaryFile := ""
		if len(f.AffectedFiles) > 0 {
			primaryFile = f.AffectedFiles[0]
		}
		graph.AddNode(&GraphNode{
			ID:       f.FindingID,
			Category: f.Category,
			Finding:  f,
			File:     primaryFile,
			Title:    f.Title,
			Weight:   f.Confidence,
		})
	}

	// 2. Identify symptom clusters and add edges
	for i := 0; i < len(findings); i++ {
		for j := i + 1; j < len(findings); j++ {
			f1 := &findings[i]
			f2 := &findings[j]

			// Same file clustering
			f1File := ""
			if len(f1.AffectedFiles) > 0 {
				f1File = f1.AffectedFiles[0]
			}
			f2File := ""
			if len(f2.AffectedFiles) > 0 {
				f2File = f2.AffectedFiles[0]
			}

			if f1File != "" && f1File == f2File {
				// Co-located findings share origin
				graph.AddEdge(GraphEdge{
					SourceID:  f1.FindingID,
					TargetID:  f2.FindingID,
					Type:      EdgeSharedOrigin,
					Weight:    0.85,
					Rationale: fmt.Sprintf("Co-located in source file %s", f1File),
				})
			} else if filepath.Dir(f1File) == filepath.Dir(f2File) && f1.Category == f2.Category {
				// Same directory and same category
				graph.AddEdge(GraphEdge{
					SourceID:  f1.FindingID,
					TargetID:  f2.FindingID,
					Type:      EdgeCoRemediate,
					Weight:    0.70,
					Rationale: fmt.Sprintf("Same package directory %s and category %s", filepath.Dir(f1File), f1.Category),
				})
			}
		}
	}

	// 3. Cluster into RootCauseEntity objects
	components := graph.FindConnectedComponents()
	var entities []RootCauseEntity

	findingMap := make(map[string]*repo_intel.Finding)
	for i := range findings {
		findingMap[findings[i].FindingID] = &findings[i]
	}

	clusterIdx := 1
	for _, comp := range components {
		if len(comp) == 0 {
			continue
		}

		var compFindings []*repo_intel.Finding
		affectedFilesSet := make(map[string]bool)
		var confidences []float64

		for _, nodeID := range comp {
			if f, ok := findingMap[nodeID]; ok {
				compFindings = append(compFindings, f)
				for _, af := range f.AffectedFiles {
					affectedFilesSet[af] = true
				}
				confidences = append(confidences, f.Confidence)
			}
		}

		if len(compFindings) == 0 {
			continue
		}

		var affectedFiles []string
		for af := range affectedFilesSet {
			affectedFiles = append(affectedFiles, af)
		}
		sort.Strings(affectedFiles)

		primaryFile := affectedFiles[0]
		pattern, title, desc, resolution, priority := classifyPattern(compFindings, primaryFile)

		// Aggregated Bayesian Confidence: 1 - prod(1 - c_i)
		prod := 1.0
		for _, c := range confidences {
			prod *= (1.0 - c)
		}
		aggConfidence := 1.0 - prod
		if aggConfidence > 0.99 {
			aggConfidence = 0.99
		}
		if aggConfidence < 0.80 {
			aggConfidence = 0.80
		}

		entity := RootCauseEntity{
			RootCauseID: fmt.Sprintf("RC-%03d-%s", clusterIdx, string(pattern)),
			Pattern:     pattern,
			Title:       title,
			Description: desc,
			PrimaryFile: primaryFile,
			AffectedFiles: affectedFiles,
			SymptomFindingIDs: comp,
			Confidence: math.Round(aggConfidence*100) / 100,
			ConfidenceRationale: fmt.Sprintf("Bayesian aggregation across %d confirming symptoms in cluster", len(compFindings)),
			RemediationPriority: priority,
			RecommendedResolution: resolution,
		}

		entities = append(entities, entity)
		clusterIdx++
	}

	// Sort entities by priority (1 is highest)
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].RemediationPriority != entities[j].RemediationPriority {
			return entities[i].RemediationPriority < entities[j].RemediationPriority
		}
		return entities[i].Confidence > entities[j].Confidence
	})

	return entities, graph
}

func classifyPattern(findings []*repo_intel.Finding, primaryFile string) (RootCausePattern, string, string, string, int) {
	hasNesting := false
	hasComplexity := false
	hasSecret := false
	hasSQL := false
	hasCmd := false
	hasResource := false
	hasError := false
	hasArchitecture := false
	hasDocker := false
	hasLockfile := false
	hasTest := false
	hasDoc := false
	hasDuplicate := false

	for _, f := range findings {
		id := strings.ToUpper(f.FindingID)
		cat := f.Category

		if strings.Contains(id, "NEST") {
			hasNesting = true
		}
		if strings.Contains(id, "COMPLEX") {
			hasComplexity = true
		}
		if strings.Contains(id, "SECRET") {
			hasSecret = true
		}
		if strings.Contains(id, "SQL") {
			hasSQL = true
		}
		if strings.Contains(id, "CMD") {
			hasCmd = true
		}
		if strings.Contains(id, "LEAK") || strings.Contains(id, "DEFER") || strings.Contains(id, "RESOURCE") {
			hasResource = true
		}
		if strings.Contains(id, "ERR") {
			hasError = true
		}
		if cat == repo_intel.CategoryArchitecture || strings.Contains(id, "LAYER") || strings.Contains(id, "CYCLE") {
			hasArchitecture = true
		}
		if strings.Contains(id, "DOCKER") || strings.Contains(id, "DEP-01") || strings.Contains(id, "DEP-02") {
			hasDocker = true
		}
		if strings.Contains(id, "LOCK") || strings.Contains(id, "DEP-03") {
			hasLockfile = true
		}
		if cat == repo_intel.CategoryTestQuality || strings.Contains(id, "TEST") {
			hasTest = true
		}
		if cat == repo_intel.CategoryDocumentation || strings.Contains(id, "DOC") {
			hasDoc = true
		}
		if cat == repo_intel.CategoryMaintainability || strings.Contains(id, "DUPLICATE") {
			hasDuplicate = true
		}
	}

	switch {
	case hasSecret || hasSQL || hasCmd:
		return PatternInputTaintLeak,
			fmt.Sprintf("Input Taint & Credential Exposure in %s", primaryFile),
			"Unsafe handling of external input or hardcoded credentials poses severe security risks.",
			"Parameterize all dynamic queries, isolate secrets to environment configs, and sanitize command arguments.",
			1

	case hasResource:
		return PatternResourceNeglect,
			fmt.Sprintf("Resource Lifecycle Neglect in %s", primaryFile),
			"Handles allocated in iteration loops or long-lived functions without prompt per-cycle release.",
			"Explicitly close resources at end of each iteration or wrap inside scoped function closures.",
			2

	case hasArchitecture:
		return PatternArchitecturalDrift,
			fmt.Sprintf("Architectural Drift & Coupling in %s", primaryFile),
			"Presentation layers bypass domain boundaries to reach persistence or mutual type circularities exist.",
			"Introduce repository abstraction layer and resolve mutual circular dependencies via interfaces.",
			2

	case hasDocker || hasLockfile:
		return PatternReproducibilityDeficit,
			fmt.Sprintf("Build Reproducibility & Supply Chain Deficit in %s", primaryFile),
			"Container tags are mutable or dependency checksums are unpinned, preventing tamper-proof reproducible builds.",
			"Pin container base image with explicit SHA digest/version and commit cryptographic lockfile (go.sum).",
			3

	case hasNesting || hasComplexity:
		return PatternStructuralCongestion,
			fmt.Sprintf("Structural Congestion & Cognitive Overload in %s", primaryFile),
			"Excessive control-flow nesting depth combined with elevated cyclomatic branching.",
			"Decompose monolithic conditionals using early guard returns and extract modular sub-functions.",
			3

	case hasError:
		return PatternErrorSuppression,
			fmt.Sprintf("Error Suppression & Silent Failure in %s", primaryFile),
			"Function error return values discarded or ignored, allowing silent operational failures.",
			"Inspect returned error values and propagate or log errors with descriptive context.",
			3

	case hasDuplicate:
		return PatternMaintainabilityDecay,
			fmt.Sprintf("Maintainability Decay & Code Duplication in %s", primaryFile),
			"Duplicate helper functions with identical AST blocks maintain redundant logic across files.",
			"Refactor redundant implementations into a single shared utility package.",
			4

	case hasTest || hasDoc:
		return PatternVerificationDeficit,
			fmt.Sprintf("Verification & Documentation Deficit in %s", primaryFile),
			"Exported public symbols lack automated unit tests or required godoc documentation.",
			"Add comprehensive test coverage in adjacent _test.go files and document public API exports.",
			4

	default:
		return PatternStructuralCongestion,
			fmt.Sprintf("Code Quality Anomaly in %s", primaryFile),
			"General structural and quality findings identified in module.",
			"Remediate according to finding guidelines.",
			4
	}
}

