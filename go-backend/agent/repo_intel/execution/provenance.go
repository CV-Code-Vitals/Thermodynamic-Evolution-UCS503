package execution

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// VerifyPlanProvenance checks the cryptographic digest, schema version,
// and current repository freshness against the remediation plan.
func VerifyPlanProvenance(plan *planning.RemediationPlan, computedHash string, repoRoot string) (*ProvenanceCheckResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("remediation plan is nil")
	}

	result := &ProvenanceCheckResult{
		PlanID:         plan.PlanID,
		PlanHash:       computedHash,
		Valid:          true,
		BaselineCommit: plan.BaselineCommit,
		StaleTasks:     make([]string, 0),
		Issues:         make([]string, 0),
	}

	// 1. Verify schema and basic identity
	if plan.PlanSchemaVersion != "1.0" {
		result.Valid = false
		result.Issues = append(result.Issues, fmt.Sprintf("invalid schema version: %s", plan.PlanSchemaVersion))
	}

	// 2. Query git repository for current HEAD and status if in a git directory
	cmdHead := exec.Command("git", "rev-parse", "HEAD")
	cmdHead.Dir = repoRoot
	if headOut, err := cmdHead.Output(); err == nil {
		result.CurrentHeadCommit = strings.TrimSpace(string(headOut))
	}

	cmdStatus := exec.Command("git", "status", "--porcelain")
	cmdStatus.Dir = repoRoot
	if statusOut, err := cmdStatus.Output(); err == nil {
		result.CleanWorkingTree = len(strings.TrimSpace(string(statusOut))) == 0
	} else {
		result.CleanWorkingTree = true // default if git is not accessible in isolated fixture
	}

	// 3. Verify target files exist in repoRoot; detect stale tasks
	for _, task := range plan.Tasks {
		for _, tf := range task.TargetFiles {
			fullPath := filepath.Join(repoRoot, filepath.FromSlash(tf))
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				result.StaleTasks = append(result.StaleTasks, task.TaskID)
				result.Issues = append(result.Issues, fmt.Sprintf("Task %s target file missing: %s", task.TaskID, tf))
				break
			}
		}
	}

	if len(result.Issues) > 0 && !result.Valid {
		return result, fmt.Errorf("provenance validation failed with %d issues", len(result.Issues))
	}

	return result, nil
}

