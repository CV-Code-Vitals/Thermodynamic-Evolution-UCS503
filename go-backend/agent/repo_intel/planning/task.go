package planning

import (
	"fmt"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// TaskFactory constructs structured RemediationTask objects from FindingGroups.
type TaskFactory struct {
	strategySelector *StrategySelector
	priorityEngine   *PriorityEngine
	criteriaGen      *CriteriaGenerator
}

// NewTaskFactory creates a new TaskFactory.
func NewTaskFactory() *TaskFactory {
	return &TaskFactory{
		strategySelector: NewStrategySelector(),
		priorityEngine:   NewPriorityEngine(),
		criteriaGen:      NewCriteriaGenerator(),
	}
}

// CreateTask builds an enriched RemediationTask from a FindingGroup.
func (tf *TaskFactory) CreateTask(taskIndex int, group FindingGroup) RemediationTask {
	primary := group.PrimaryFinding
	taskID := fmt.Sprintf("RP-%04d", taskIndex)

	// Collect all source finding IDs
	sourceIDs := []string{primary.FindingID}
	for _, rel := range group.RelatedFindings {
		sourceIDs = append(sourceIDs, rel.FindingID)
	}

	// Collect target files and target symbols
	targetFiles := uniqueStrings(primary.AffectedComponents)
	var targetSymbols []string
	if sym := extractSymbolFromFindingID(primary.FindingID); sym != "" {
		targetSymbols = append(targetSymbols, sym)
	}

	for _, rel := range group.RelatedFindings {
		targetFiles = uniqueStrings(append(targetFiles, rel.AffectedComponents...))
		if sym := extractSymbolFromFindingID(rel.FindingID); sym != "" {
			targetSymbols = uniqueStrings(append(targetSymbols, sym))
		}
	}

	// Strategy selection
	strategy, execClass, approvalReq, approvalRationale := tf.strategySelector.SelectStrategy(&primary)

	// Priority evaluation
	priority, prioScore, prioRationale := tf.priorityEngine.EvaluatePriority(&primary)

	// Acceptance criteria, validation commands, rollback conditions
	criteria, commands, rollback := tf.criteriaGen.GenerateCriteria(&primary, targetFiles)

	title := primary.Title
	if len(group.RelatedFindings) > 0 {
		title = fmt.Sprintf("%s (+ %d related findings)", primary.Title, len(group.RelatedFindings))
	}

	status := "PLANNED"
	if approvalReq {
		status = "REQUIRES_APPROVAL"
	}

	return RemediationTask{
		TaskID:                taskID,
		Title:                 title,
		SourceFindingIDs:      sourceIDs,
		Category:              primary.Category,
		Description:           primary.Description,
		TargetFiles:           targetFiles,
		TargetSymbols:         targetSymbols,
		ExecutionStrategy:     strategy,
		ExecutionClass:        execClass,
		Priority:              priority,
		PriorityScore:         prioScore,
		PriorityRationale:     prioRationale,
		ExpectedBenefit:       primary.ExpectedBenefit,
		RegressionRisk:        primary.RegressionRisk,
		Dependencies:          []string{},
		Conflicts:             []TaskConflict{},
		AcceptanceCriteria:    criteria,
		ValidationCommands:    commands,
		RollbackConditions:    rollback,
		HumanApprovalRequired: approvalReq,
		ApprovalRationale:     approvalRationale,
		Status:                status,
	}
}

func extractSymbolFromFindingID(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if (p == "NEST" || p == "COMPLEX" || p == "VOLUME" || p == "HIGH_NESTING" || p == "HIGH_COMPLEXITY") && i+2 < len(parts) {
			return parts[i+2]
		}
	}
	return ""
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var res []string
	for _, s := range slice {
		if s != "" && !seen[s] {
			seen[s] = true
			res = append(res, s)
		}
	}
	return res
}

func init() {
	_ = repo_intel.CategoryStructural
	_ = diagnostic.StateValidated
}
