package diagnostic

import (
	"fmt"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ConfidenceCalibrator computes an objective, evidence-backed confidence score
// independent of the finding's severity.
type ConfidenceCalibrator struct{}

// NewConfidenceCalibrator creates a new ConfidenceCalibrator.
func NewConfidenceCalibrator() *ConfidenceCalibrator {
	return &ConfidenceCalibrator{}
}

// CalibrateConfidence evaluates the detection method, verification state, and reproducibility.
func (c *ConfidenceCalibrator) CalibrateConfidence(f *ValidatedFinding) (float64, string) {
	// Base confidence by detection method
	var baseScore float64
	var methodRationale string

	switch f.DetectionMethod {
	case "TREE_SITTER_AST":
		baseScore = 0.95
		methodRationale = "Deterministic Tree-sitter AST syntax and complexity extraction"
	case "GRAPH_CYCLE_DETECTION":
		baseScore = 0.98
		methodRationale = "Deterministic Tarjan/DFS topological cycle proof"
	case "MANIFEST_ANALYSIS":
		baseScore = 0.92
		methodRationale = "Deterministic package manifest parser"
	case "FILE_CORRELATION_HEURISTIC":
		baseScore = 0.70
		methodRationale = "Filesystem directory correlation heuristic"
	case "STATIC_TAINT_HEURISTIC":
		baseScore = 0.55
		methodRationale = "Lexical pattern matching with potential false-positive context"
	case "REGEX_PATTERN_MATCHING":
		baseScore = 0.65
		methodRationale = "Regular expression token match"
	default:
		baseScore = 0.50
		methodRationale = "Generic heuristic detection"
	}

	// Adjust based on ValidationState
	switch f.ValidationState {
	case StateValidated:
		// No discount
	case StatePartiallyValidated:
		baseScore -= 0.20
		methodRationale += "; discounted due to partial evidence verification"
	case StateStale:
		baseScore = 0.15
		methodRationale += "; severely discounted: finding references stale source state"
	case StateInvalid:
		baseScore = 0.05
		methodRationale += "; invalidated: evidence or metadata invalid"
	case StateRequiresHumanReview:
		baseScore -= 0.10
		methodRationale += "; discounted pending human verification"
	}

	// Specific heuristic discounts
	if f.Category == repo_intel.CategorySecurity && strings.Contains(f.FindingID, "SQL") {
		// Dynamic string concatenation without semantic taint proof is a heuristic
		baseScore = 0.45
		methodRationale = "Static SQL string concatenation without dataflow taint proof; requires runtime query inspection"
	}

	if strings.Contains(f.FindingID, "NOTEST") {
		// Non-co-located tests might exist in an integration/test suite directory
		baseScore = 0.60
		methodRationale = "Absence of sibling unit tests; integration tests may reside in root or e2e suites"
	}

	// Clamp between 0.0 and 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	rationale := fmt.Sprintf("%.2f: %s", baseScore, methodRationale)
	return baseScore, rationale
}

