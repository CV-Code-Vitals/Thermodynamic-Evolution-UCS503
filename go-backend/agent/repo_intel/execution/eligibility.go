package execution

import (
	"fmt"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// CheckTaskEligibility validates whether a task can be executed under the current
// approval and execution class policies.
func CheckTaskEligibility(task *planning.RemediationTask, approvedIDs map[string]bool) (bool, bool, string) {
	if task == nil {
		return false, false, "task is nil"
	}

	// 1. Check if human approval is explicitly required
	isHumanApprovalRequired := task.HumanApprovalRequired || task.ExecutionClass == planning.ClassHumanApprovalRequired

	if isHumanApprovalRequired {
		if approvedIDs != nil && approvedIDs[task.TaskID] {
			// Approved by human operator
			return true, false, ""
		}
		// Fails closed by default
		return false, false, fmt.Sprintf("EXEC_REJECTED_UNAPPROVED: Task %s requires explicit human approval (%s)", task.TaskID, task.ApprovalRationale)
	}

	// 2. Check execution class
	switch task.ExecutionClass {
	case planning.ClassAutomationReady:
		// Fully autonomous: patch generation and validation permitted
		return true, false, ""

	case planning.ClassAutomationWithReview:
		// Autonomous generation & validation permitted, review required for repository merge
		return true, true, ""

	case planning.ClassInvestigationRequired, planning.ClassInformationalOnly:
		return false, false, fmt.Sprintf("EXEC_REJECTED: Task %s is classified as %s and cannot be executed automatically", task.TaskID, task.ExecutionClass)

	default:
		// Unknown or blocked class
		return false, false, fmt.Sprintf("EXEC_REJECTED: Task %s has unrecognized execution class %s", task.TaskID, task.ExecutionClass)
	}
}

// CheckTaskEligibilitySecure enforces cryptographic token authenticity, bounded expiry,
// scope binding (commit, file, patch, workspace), and atomic replay prevention.
func CheckTaskEligibilitySecure(
	task *planning.RemediationTask,
	token *ApprovalToken,
	authority *ApprovalAuthority,
	currentCommit string,
	currentFileHashes map[string]string,
	currentPatchDiffHash string,
	currentWorkspace string,
) (bool, bool, string) {
	if task == nil {
		return false, false, "task is nil"
	}

	isHumanApprovalRequired := task.HumanApprovalRequired || task.ExecutionClass == planning.ClassHumanApprovalRequired

	if isHumanApprovalRequired {
		if token == nil || authority == nil {
			return false, false, fmt.Sprintf("EXEC_REJECTED_UNAPPROVED: Task %s requires cryptographically signed ApprovalToken (%s)",
				task.TaskID, task.ApprovalRationale)
		}

		if token.TaskID != task.TaskID {
			return false, false, fmt.Sprintf("EXEC_REJECTED_UNAPPROVED: ApprovalToken task ID %s does not match task %s", token.TaskID, task.TaskID)
		}

		if err := authority.VerifyAndConsumeToken(token, currentCommit, currentFileHashes, currentPatchDiffHash, currentWorkspace); err != nil {
			return false, false, fmt.Sprintf("EXEC_REJECTED_UNAPPROVED: %v", err)
		}

		return true, false, ""
	}

	switch task.ExecutionClass {
	case planning.ClassAutomationReady:
		return true, false, ""

	case planning.ClassAutomationWithReview:
		return true, true, ""

	case planning.ClassInvestigationRequired, planning.ClassInformationalOnly:
		return false, false, fmt.Sprintf("EXEC_REJECTED: Task %s is classified as %s and cannot be executed automatically", task.TaskID, task.ExecutionClass)

	default:
		return false, false, fmt.Sprintf("EXEC_REJECTED: Task %s has unrecognized execution class %s", task.TaskID, task.ExecutionClass)
	}
}

