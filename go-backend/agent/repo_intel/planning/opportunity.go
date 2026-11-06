package planning

import (
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// OpportunityAnalyzer assesses whether a diagnostic finding qualifies as an actionable remediation item.
type OpportunityAnalyzer struct{}

// NewOpportunityAnalyzer creates a new OpportunityAnalyzer.
func NewOpportunityAnalyzer() *OpportunityAnalyzer {
	return &OpportunityAnalyzer{}
}

// EvaluateActionability maps a ValidatedFinding into an ActionabilityState with clear justification.
func (oa *OpportunityAnalyzer) EvaluateActionability(f *diagnostic.ValidatedFinding) (ActionabilityState, string) {
	// 1. Rejections
	if f.ValidationState == diagnostic.StateInvalid {
		return RejectedAsInvalid, "Finding rejected: invalid metadata or evidence could not be validated on disk"
	}
	if f.ValidationState == diagnostic.StateStale {
		return RejectedAsStale, "Finding rejected: referenced code or line numbers are stale relative to repository state"
	}

	// 2. Human Approval Gate
	if f.RemediationReadiness == diagnostic.ReadinessHumanApprovalRequired ||
		(f.Category == repo_intel.CategorySecurity && strings.Contains(f.FindingID, "SECRET")) ||
		(f.Category == repo_intel.CategorySecurity && (strings.Contains(f.FindingID, "CMD") || strings.Contains(f.FindingID, "COMMAND"))) ||
		(f.Category == repo_intel.CategoryArchitecture && strings.Contains(f.FindingID, "CYCLE")) ||
		strings.Contains(f.ReadinessRationale, "concurrency") ||
		strings.Contains(f.ReadinessRationale, "goroutine") ||
		strings.Contains(f.ReadinessRationale, "channel") {
		return HumanApprovalRequired, "High-impact security modification, concurrent ownership ambiguity, or architectural cyclic dependency requires explicit human review and authorization"
	}

	// 3. Investigation Required
	if f.ValidationState == diagnostic.StateRequiresHumanReview ||
		f.RemediationReadiness == diagnostic.ReadinessInvestigationRequired ||
		(f.CalibratedConfidence > 0 && f.CalibratedConfidence < 0.60) {
		return InvestigationRequired, "Heuristic finding with confidence below automated remediation threshold; manual code review required"
	}

	// 4. Actionable
	if f.RemediationReadiness == diagnostic.ReadinessAutomationReady && f.ValidationState == diagnostic.StateValidated {
		return Actionable, "Validated, deterministic finding suitable for direct automated remediation"
	}

	// 5. Actionable with Review
	if f.RemediationReadiness == diagnostic.ReadinessAutomationWithReview && f.ValidationState == diagnostic.StateValidated {
		return ActionableWithReview, "Validated structural refactoring or test scaffolding suitable for automation guarded by regression test gates"
	}

	return Informational, "Non-actionable observation or informational metric"
}
