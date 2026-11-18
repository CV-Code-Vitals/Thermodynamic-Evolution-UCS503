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

	// 2. Investigation Only (Heuristic low confidence or unverified)
	if f.ValidationState == diagnostic.StateRequiresHumanReview || (f.CalibratedConfidence > 0 && f.CalibratedConfidence < 0.60) {
		return StrategyInvestigationOnly, ClassInvestigationRequired, true,
			"Heuristic finding requiring manual code review before any modification plan can be approved"
	}

	// 2.5 Semantic Engine Findings (Milestone 3.5 & 3.6)
	if strings.Contains(f.FindingID, "SEM-SEC-SQL") || strings.Contains(f.FindingID, "SQL_INJECTION") {
		return StrategyASTTransformation, ClassAutomationWithReview, false,
			"Parameterized SQL query rewriting guarded by compiler syntax checks and test suite verification"
	}

	if strings.Contains(f.FindingID, "SEM-SEC-CMD") || strings.Contains(f.FindingID, "COMMAND_INJECTION") {
		return StrategyASTTransformation, ClassHumanApprovalRequired, true,
			"Shell interpreter refactoring alters execution semantics and environment variables; human authorization required"
	}

	if strings.Contains(f.FindingID, "SEM-REL-RES") || strings.Contains(f.FindingID, "RESOURCE_LEAK") {
		// If concurrent ownership or escape is present, strictly enforce Human Approval
		if strings.Contains(f.ReadinessRationale, "concurrency") ||
			strings.Contains(f.ReadinessRationale, "goroutine") ||
			strings.Contains(f.ReadinessRationale, "channel") ||
			strings.Contains(f.Description, "loop") ||
			f.RemediationReadiness == diagnostic.ReadinessHumanApprovalRequired {
			return StrategyHumanGuided, ClassHumanApprovalRequired, true,
				"Resource handle with concurrent escape, loop accumulation, or uncertain ownership requires human authorization"
		}
		return StrategyASTTransformation, ClassAutomationReady, false,
			"Deterministic deferred resource cleanup (defer Close()) along verified local execution exits"
	}

	if strings.Contains(f.FindingID, "SEM-REL-ERR") || strings.Contains(f.FindingID, "UNHANDLED_ERROR") {
		if f.RemediationReadiness == diagnostic.ReadinessHumanApprovalRequired {
			return StrategyHumanGuided, ClassHumanApprovalRequired, true,
				"Partial error handling or cleanup error discard requires developer architectural review"
		}
		return StrategyASTTransformation, ClassAutomationReady, false,
			"Deterministic error check injection (if err != nil) with safe return propagation"
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
