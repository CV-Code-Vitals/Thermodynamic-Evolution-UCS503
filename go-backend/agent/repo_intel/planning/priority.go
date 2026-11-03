package planning

import (
	"fmt"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// PriorityEngine assigns deterministic priority rankings and numerical priority scores.
type PriorityEngine struct{}

// NewPriorityEngine creates a new PriorityEngine.
func NewPriorityEngine() *PriorityEngine {
	return &PriorityEngine{}
}

// EvaluatePriority computes a deterministic Priority ranking, numeric score, and rationale.
func (pe *PriorityEngine) EvaluatePriority(f *diagnostic.ValidatedFinding) (TaskPriority, float64, string) {
	// Base points from Severity
	var sevPoints float64
	var rank TaskPriority

	switch f.Severity {
	case repo_intel.SeverityCritical:
		sevPoints = 60.0
		rank = PriorityUrgent
	case repo_intel.SeverityHigh:
		sevPoints = 40.0
		rank = PriorityHigh
	case repo_intel.SeverityMedium:
		sevPoints = 25.0
		rank = PriorityMedium
	case repo_intel.SeverityLow:
		sevPoints = 10.0
		rank = PriorityLow
	default:
		sevPoints = 5.0
		rank = PriorityLow
	}

	// Confidence points (up to 20.0)
	confPoints := f.CalibratedConfidence * 20.0

	// Special overrides
	var rationale string
	if f.Category == repo_intel.CategorySecurity && strings.Contains(f.FindingID, "SECRET") {
		rank = PriorityUrgent
		sevPoints = 70.0
		rationale = "CRITICAL security exposure: exposed credentials demand urgent remediation"
	} else if f.Category == repo_intel.CategoryArchitecture && strings.Contains(f.FindingID, "CYCLE") {
		rank = PriorityHigh
		sevPoints = 50.0
		rationale = "HIGH architectural blocker: circular dependencies prevent modular testing and builds"
	} else if strings.Contains(f.FindingID, "UNPINNED-BASE") || strings.Contains(f.FindingID, "ROOT-USER") {
		rank = PriorityMedium
		rationale = "MEDIUM container hygiene: automated fix with low regression risk"
	} else {
		rationale = fmt.Sprintf("%s priority based on severity (%s) and calibrated confidence (%.2f)", rank, f.Severity, f.CalibratedConfidence)
	}

	totalScore := sevPoints + confPoints
	return rank, totalScore, rationale
}

