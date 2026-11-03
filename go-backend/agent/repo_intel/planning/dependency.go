package planning

import (
	"fmt"
	"sort"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// DependencyResolver builds the directed task dependency graph and computes a deterministic topological ordering.
type DependencyResolver struct{}

// NewDependencyResolver creates a new DependencyResolver.
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{}
}

// ResolveDependencies infers prerequisites between tasks and computes topological execution order.
func (dr *DependencyResolver) ResolveDependencies(tasks []RemediationTask) (map[string][]string, []string, []string, []RemediationTask) {
	depGraph := make(map[string][]string) // TaskID -> list of prerequisite TaskIDs
	taskMap := make(map[string]RemediationTask)
	fileToTasks := make(map[string][]string)

	for _, t := range tasks {
		taskMap[t.TaskID] = t
		depGraph[t.TaskID] = []string{}
		for _, f := range t.TargetFiles {
			fileToTasks[f] = append(fileToTasks[f], t.TaskID)
		}
	}

	// 1. Dependency Rule: Config / Lockfile tasks precede code refactorings in the same modules
	for _, t := range tasks {
		if t.Category == repo_intel.CategoryDependency || strings.Contains(t.TaskID, "DOCKER") {
			// Other tasks in the same directory depend on dependency/lockfile tasks
			for _, tf := range t.TargetFiles {
				for _, otherID := range fileToTasks[tf] {
					if otherID != t.TaskID && taskMap[otherID].Category == repo_intel.CategoryStructural {
						depGraph[otherID] = uniqueStrings(append(depGraph[otherID], t.TaskID))
					}
				}
			}
		}
	}

	// 2. Dependency Rule: Architectural Cycle Breaking precedes internal function refactorings
	var cycleTaskIDs []string
	for _, t := range tasks {
		if t.Category == repo_intel.CategoryArchitecture && strings.Contains(t.Title, "Circular") {
			cycleTaskIDs = append(cycleTaskIDs, t.TaskID)
		}
	}
	for _, cycleID := range cycleTaskIDs {
		cycleTask := taskMap[cycleID]
		for _, cf := range cycleTask.TargetFiles {
			for _, otherID := range fileToTasks[cf] {
				if otherID != cycleID && taskMap[otherID].Category == repo_intel.CategoryStructural {
					depGraph[otherID] = uniqueStrings(append(depGraph[otherID], cycleID))
				}
			}
		}
	}

	// 3. Topological Sort (Kahn's algorithm)
	inDegree := make(map[string]int)
	for id := range depGraph {
		inDegree[id] = len(depGraph[id])
	}

	// Track outgoing edges: prerequisite -> dependents
	dependents := make(map[string][]string)
	for id, prereqs := range depGraph {
		for _, p := range prereqs {
			dependents[p] = append(dependents[p], id)
		}
	}

	var zeroQueue []string
	for id, deg := range inDegree {
		if deg == 0 {
			zeroQueue = append(zeroQueue, id)
		}
	}

	// Deterministic sorting of initial queue: PriorityScore desc, TaskID asc
	sort.Slice(zeroQueue, func(i, j int) bool {
		ti, tj := taskMap[zeroQueue[i]], taskMap[zeroQueue[j]]
		if ti.PriorityScore != tj.PriorityScore {
			return ti.PriorityScore > tj.PriorityScore
		}
		return ti.TaskID < tj.TaskID
	})

	var executionOrder []string
	visitedCount := 0

	for len(zeroQueue) > 0 {
		curr := zeroQueue[0]
		zeroQueue = zeroQueue[1:]
		executionOrder = append(executionOrder, curr)
		visitedCount++

		for _, depID := range dependents[curr] {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				zeroQueue = append(zeroQueue, depID)
				// Re-sort queue deterministically
				sort.Slice(zeroQueue, func(i, j int) bool {
					ti, tj := taskMap[zeroQueue[i]], taskMap[zeroQueue[j]]
					if ti.PriorityScore != tj.PriorityScore {
						return ti.PriorityScore > tj.PriorityScore
					}
					return ti.TaskID < tj.TaskID
				})
			}
		}
	}

	var blockedGroups []string
	updatedTasks := make([]RemediationTask, len(tasks))
	copy(updatedTasks, tasks)

	// Check if a cycle blocked any tasks
	if visitedCount < len(tasks) {
		for i := range updatedTasks {
			t := &updatedTasks[i]
			if inDegree[t.TaskID] > 0 {
				t.Status = "BLOCKED"
				blockedGroups = append(blockedGroups, fmt.Sprintf("Task %s is blocked due to circular task dependency involving %s", t.TaskID, strings.Join(depGraph[t.TaskID], ", ")))
			}
		}
	}

	// Populate dependencies field on each task
	for i := range updatedTasks {
		t := &updatedTasks[i]
		t.Dependencies = depGraph[t.TaskID]
	}

	return depGraph, executionOrder, blockedGroups, updatedTasks
}

