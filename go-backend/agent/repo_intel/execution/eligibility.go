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

