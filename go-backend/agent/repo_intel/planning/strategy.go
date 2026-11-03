package planning

import (
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// StrategySelector determines the execution strategy and approval requirements for a task.
type StrategySelector struct{}

// NewStrategySelector creates a new StrategySelector.
func NewStrategySelector() *StrategySelector {
	return &StrategySelector{}
}

// SelectStrategy assigns conservative execution strategy and execution class.
func (ss *StrategySelector) SelectStrategy(f *diagnostic.ValidatedFinding) (ExecutionStrategy, ExecutionClass, bool, string) {
	// 1. Human-Guided (Security secrets & architectural cycles)
	if f.Category == repo_intel.CategorySecurity && strings.Contains(f.FindingID, "SECRET") {
		return StrategyHumanGuided, ClassHumanApprovalRequired, true,
			"Credential revocation, cloud secret rotation, and IAM policy provisioning require human authorization"
	}
	if f.Category == repo_intel.CategoryArchitecture && strings.Contains(f.FindingID, "CYCLE") {
		return StrategyHumanGuided, ClassHumanApprovalRequired, true,
			"Resolving architectural circular dependencies requires structural package interface redesign and human approval"
	}

	// 2. Investigation Only
	if f.ValidationState == diagnostic.StateRequiresHumanReview || (f.CalibratedConfidence > 0 && f.CalibratedConfidence < 0.60) {
		return StrategyInvestigationOnly, ClassInvestigationRequired, true,
			"Heuristic finding requiring manual code review before any modification plan can be approved"
	}

	// 3. Rule-Based (Deterministic declarative modifications)
	if strings.Contains(f.FindingID, "UNPINNED-BASE") ||
		strings.Contains(f.FindingID, "ROOT-USER") ||
		strings.Contains(f.FindingID, "MISSING-LICENSE") ||
		strings.Contains(f.FindingID, "GOSUM-MISSING") {
		return StrategyRuleBased, ClassAutomationReady, false,
			"Safe, deterministic file modification with verified validation rules"
	}

	// 4. AST Transformation (Local structural complexity)
	if strings.Contains(f.FindingID, "HIGH-NESTING") || strings.Contains(f.FindingID, "HIGH-COMPLEXITY") {
		return StrategyASTTransformation, ClassAutomationWithReview, false,
			"Local AST refactoring using guard clauses and helper function extraction under regression test gates"
	}

	// 5. LLM-Assisted (Test generation, contextual refactoring)
	if strings.Contains(f.FindingID, "NOTEST") {
		return StrategyLLMAssisted, ClassAutomationWithReview, false,
			"Contextual test suite scaffolding requiring semantic understanding of package interfaces"
	}

	// Fallback
	return StrategyASTTransformation, ClassAutomationWithReview, false,
		"Automated code modification guarded by compiler and test verification"
}
