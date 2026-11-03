package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// LoadRemediationPlan loads a Phase 3 remediation plan, computes its SHA256 digest,
// and validates its schema and integrity.
func LoadRemediationPlan(filePath string) (*planning.RemediationPlan, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read remediation plan from %s: %w", filePath, err)
	}

	hashBytes := sha256.Sum256(data)
	computedHash := hex.EncodeToString(hashBytes[:])

	var plan planning.RemediationPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, computedHash, fmt.Errorf("invalid remediation plan JSON: %w", err)
	}

	// Validate basic plan schema
	if plan.PlanSchemaVersion != "1.0" {
		return nil, computedHash, fmt.Errorf("unsupported plan schema version: %q, expected '1.0'", plan.PlanSchemaVersion)
	}
	if plan.PlanID == "" {
		return nil, computedHash, fmt.Errorf("missing plan_id in remediation plan")
	}
	if len(plan.Tasks) == 0 {
		return nil, computedHash, fmt.Errorf("remediation plan contains zero tasks")
	}

	// Validate task structure and check duplicate IDs
	seenIDs := make(map[string]bool)
	for i, task := range plan.Tasks {
		if task.TaskID == "" {
			return nil, computedHash, fmt.Errorf("task at index %d has empty task_id", i)
		}
		if seenIDs[task.TaskID] {
			return nil, computedHash, fmt.Errorf("duplicate task_id detected: %s", task.TaskID)
		}
		seenIDs[task.TaskID] = true
	}

	return &plan, computedHash, nil
}

