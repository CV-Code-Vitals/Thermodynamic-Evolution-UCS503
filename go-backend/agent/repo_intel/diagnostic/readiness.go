package diagnostic

import (
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ReadinessClassifier assesses a finding's suitability for future autonomous remediation.
type ReadinessClassifier struct{}

// NewReadinessClassifier creates a new ReadinessClassifier.
func NewReadinessClassifier() *ReadinessClassifier {
	return &ReadinessClassifier{}
}

// ClassifyRemediationReadiness assigns an advisory automation readiness state and complexity.
func (rc *ReadinessClassifier) ClassifyRemediationReadiness(f *ValidatedFinding) (RemediationReadiness, string, string) {
	// If the finding was marked STALE or INVALID during validation, it is not actionable
	if f.ValidationState == StateInvalid || f.ValidationState == StateStale {
		return ReadinessInformationalOnly, "LOW", "Finding is invalid or stale; no remediation required"
	}

	// 1. Human Approval Required
	if f.Category == repo_intel.CategorySecurity && (strings.Contains(f.FindingID, "SECRET") || strings.Contains(f.FindingID, "KEY")) {
		return ReadinessHumanApprovalRequired, "HIGH", "Hardcoded credential revocation, cloud secret rotation, and IAM policy provisioning require human authorization"
	}
	if f.Category == repo_intel.CategoryArchitecture && strings.Contains(f.FindingID, "CYCLE") {
		return ReadinessHumanApprovalRequired, "HIGH", "Breaking cross-package architectural circular imports requires structural interface redesign and human approval"
	}

	// 2. Investigation Required
	if f.ValidationState == StateRequiresHumanReview || (f.CalibratedConfidence > 0 && f.CalibratedConfidence < 0.60) {
		return ReadinessInvestigationRequired, "MEDIUM", "Heuristic finding with confidence below automation threshold; manual code review required to verify true positive"
	}
	if strings.Contains(f.FindingID, "HIGH-COMPLEXITY") && f.Severity == repo_intel.SeverityHigh {
		return ReadinessInvestigationRequired, "HIGH", "Monolithic function with cyclomatic complexity >= 20 contains tangled branching logic requiring architectural inspection before decomposition"
	}

	// 3. Automation Ready
	if strings.Contains(f.FindingID, "UNPINNED-BASE") {
		return ReadinessAutomationReady, "LOW", "Deterministic Dockerfile edit: pin base image to immutable tag or digest"
	}
	if strings.Contains(f.FindingID, "ROOT-USER") {
		return ReadinessAutomationReady, "LOW", "Deterministic Dockerfile edit: append unprivileged USER and GROUP instructions"
	}
	if strings.Contains(f.FindingID, "MISSING-LICENSE") {
		return ReadinessAutomationReady, "LOW", "Safe declarative file addition: generate standard LICENSE template"
	}
	if strings.Contains(f.FindingID, "GOSUM-MISSING") {
		return ReadinessAutomationReady, "LOW", "Safe dependency lockfile generation: execute go mod download"
	}

	// 4. Automation With Review
	if strings.Contains(f.FindingID, "HIGH-NESTING") {
		return ReadinessAutomationWithReview, "MEDIUM", "Local refactoring via guard clauses and early returns; automated test verification required"
	}
	if strings.Contains(f.FindingID, "NOTEST") {
		return ReadinessAutomationWithReview, "MEDIUM", "Scaffold unit test suite template for unverified package; human review needed for test assertions"
	}
	if strings.Contains(f.FindingID, "UNHANDLED-ERR") {
		return ReadinessAutomationWithReview, "LOW", "Replace blank identifier error discard (_ = err) with error propagation or logging"
	}

	// Default fallback
	return ReadinessAutomationWithReview, "MEDIUM", "Automated code modification feasible with automated regression test gate verification"
}
