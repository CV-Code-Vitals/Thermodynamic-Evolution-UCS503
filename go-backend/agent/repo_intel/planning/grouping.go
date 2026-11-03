package planning

import (
	"fmt"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// FindingGroup clusters related findings that share an underlying root cause or target.
type FindingGroup struct {
	GroupID          string
	PrimaryFinding   diagnostic.ValidatedFinding
	RelatedFindings  []diagnostic.ValidatedFinding
	SharedTargetFile string
	SharedSymbol     string
	GroupingReason   string
}

// TaskGrouper organizes actionable findings into logical groups based on correlations and symbols.
type TaskGrouper struct{}

// NewTaskGrouper creates a new TaskGrouper.
func NewTaskGrouper() *TaskGrouper {
	return &TaskGrouper{}
}

// GroupFindings clusters findings using correlations and shared symbols/files.
func (tg *TaskGrouper) GroupFindings(
	findings []diagnostic.ValidatedFinding,
	correlations []diagnostic.FindingCorrelation,
) []FindingGroup {
	// Index correlations by PrimaryFindingID
	corrMap := make(map[string]diagnostic.FindingCorrelation)
	for _, c := range correlations {
		corrMap[c.PrimaryFindingID] = c
	}

	findingByID := make(map[string]diagnostic.ValidatedFinding)
	for _, f := range findings {
		findingByID[f.FindingID] = f
	}

	consumed := make(map[string]bool)
	var groups []FindingGroup
	groupIdx := 1

	// 1. Group by correlations first (Duplicates, Parent-Child, Coordinated Resolution)
	for _, c := range correlations {
		primary, ok := findingByID[c.PrimaryFindingID]
		if !ok || consumed[primary.FindingID] {
			continue
		}

		group := FindingGroup{
			GroupID:        fmt.Sprintf("GRP-%04d", groupIdx),
			PrimaryFinding: primary,
			GroupingReason: c.Rationale,
		}
		if len(primary.AffectedComponents) > 0 {
			group.SharedTargetFile = primary.AffectedComponents[0]
		}
		consumed[primary.FindingID] = true

		for _, relID := range c.RelatedFindingIDs {
			if relFinding, exists := findingByID[relID]; exists && !consumed[relID] {
				group.RelatedFindings = append(group.RelatedFindings, relFinding)
				consumed[relID] = true
			}
		}

		groups = append(groups, group)
		groupIdx++
	}

	// 2. Remaining individual findings become standalone groups
	for _, f := range findings {
		if consumed[f.FindingID] {
			continue
		}

		group := FindingGroup{
			GroupID:        fmt.Sprintf("GRP-%04d", groupIdx),
			PrimaryFinding: f,
			GroupingReason: "Independent standalone finding",
		}
		if len(f.AffectedComponents) > 0 {
			group.SharedTargetFile = f.AffectedComponents[0]
		}
		consumed[f.FindingID] = true
		groups = append(groups, group)
		groupIdx++
	}

	_ = strings.Contains
	return groups
}
