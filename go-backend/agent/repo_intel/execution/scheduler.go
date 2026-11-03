package execution

import (
	"fmt"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// ExecutionSchedule holds ordered batches of tasks ready for execution.
type ExecutionSchedule struct {
	OrderedTasks []*planning.RemediationTask
	Dependencies map[string][]string // task_id -> prerequisite task_ids
}

// BuildExecutionSchedule creates a dependency-ordered list of tasks to execute.
func BuildExecutionSchedule(plan *planning.RemediationPlan, targetTaskIDs []string, batchLimit int) (*ExecutionSchedule, error) {
	if plan == nil {
		return nil, fmt.Errorf("remediation plan is nil")
	}

	taskMap := make(map[string]*planning.RemediationTask)
	for i := range plan.Tasks {
		taskMap[plan.Tasks[i].TaskID] = &plan.Tasks[i]
	}

	// Filter tasks if specific targets are provided
	var selectedTasks []*planning.RemediationTask
	if len(targetTaskIDs) > 0 {
		targetSet := make(map[string]bool)
		for _, id := range targetTaskIDs {
			targetSet[id] = true
		}
		for i := range plan.Tasks {
			if targetSet[plan.Tasks[i].TaskID] {
				selectedTasks = append(selectedTasks, &plan.Tasks[i])
			}
		}
	} else {
		for i := range plan.Tasks {
			selectedTasks = append(selectedTasks, &plan.Tasks[i])
		}
	}

	// Apply batch limit
	if batchLimit > 0 && len(selectedTasks) > batchLimit {
		selectedTasks = selectedTasks[:batchLimit]
	}

	// Build dependency graph for selected tasks
	deps := make(map[string][]string)
	for _, task := range selectedTasks {
		deps[task.TaskID] = task.Dependencies
	}

	// Topologically sort tasks (ensuring prerequisites come first)
	visited := make(map[string]bool)
	tempMark := make(map[string]bool)
	var ordered []*planning.RemediationTask

	var visit func(task *planning.RemediationTask) error
	visit = func(t *planning.RemediationTask) error {
		if tempMark[t.TaskID] {
			return fmt.Errorf("circular dependency detected involving task %s", t.TaskID)
		}
		if !visited[t.TaskID] {
			tempMark[t.TaskID] = true
			for _, depID := range t.Dependencies {
				if depTask, exists := taskMap[depID]; exists {
					if err := visit(depTask); err != nil {
						return err
					}
				}
			}
			tempMark[t.TaskID] = false
			visited[t.TaskID] = true
			ordered = append(ordered, t)
		}
		return nil
	}

	for _, task := range selectedTasks {
		if !visited[task.TaskID] {
			if err := visit(task); err != nil {
				return nil, err
			}
		}
	}

	return &ExecutionSchedule{
		OrderedTasks: ordered,
		Dependencies: deps,
	}, nil
}
