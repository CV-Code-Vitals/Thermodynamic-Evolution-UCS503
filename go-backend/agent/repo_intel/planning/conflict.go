package planning

import (
	"fmt"
	"strings"
)

// ConflictDetector identifies resource contention, overlapping symbols, and semantic incompatibilities between tasks.
type ConflictDetector struct{}

// NewConflictDetector creates a new ConflictDetector.
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{}
}

// DetectConflicts scans a set of remediation tasks and identifies potential hard and soft conflicts.
func (cd *ConflictDetector) DetectConflicts(tasks []RemediationTask) ([]TaskConflict, []RemediationTask) {
	var conflicts []TaskConflict
	confIdx := 1

	// File and symbol maps
	fileToTasks := make(map[string][]string)
	type symKey struct {
		file   string
		symbol string
	}
	symbolToTasks := make(map[symKey][]string)

	for _, t := range tasks {
		for _, f := range t.TargetFiles {
			fileToTasks[f] = append(fileToTasks[f], t.TaskID)
			for _, s := range t.TargetSymbols {
				sk := symKey{file: f, symbol: s}
				symbolToTasks[sk] = append(symbolToTasks[sk], t.TaskID)
			}
		}
	}

	taskConflictMap := make(map[string][]TaskConflict)

	// 1. HARD CONFLICT: Multiple tasks modifying the exact same function symbol in the same file
	for sk, tIDs := range symbolToTasks {
		deduped := uniqueStrings(tIDs)
		if len(deduped) > 1 {
			for i := 0; i < len(deduped); i++ {
				for j := i + 1; j < len(deduped); j++ {
					tA, tB := deduped[i], deduped[j]
					c := TaskConflict{
						ConflictID:           fmt.Sprintf("CONF-HARD-%04d", confIdx),
						Type:                 ConflictHard,
						TaskID:               tA,
						ConflictingTaskID:    tB,
						ConflictingResource:  fmt.Sprintf("%s::%s", sk.file, sk.symbol),
						Explanation:          fmt.Sprintf("Tasks %s and %s both modify function %q in %s; concurrent modification would cause merge conflicts or overwrites", tA, tB, sk.symbol, sk.file),
						ResolutionSuggestion: "Consolidate into a single task or sequence them strictly using dependency order",
					}
					conflicts = append(conflicts, c)
					confIdx++
					taskConflictMap[tA] = append(taskConflictMap[tA], c)
					taskConflictMap[tB] = append(taskConflictMap[tB], c)
				}
			}
		}
	}

	// 2. COORDINATION REQUIRED: Multiple tasks modifying the same critical config file (Dockerfile, go.mod)
	for f, tIDs := range fileToTasks {
		deduped := uniqueStrings(tIDs)
		if len(deduped) > 1 && (strings.HasSuffix(f, "Dockerfile") || strings.HasSuffix(f, "go.mod") || strings.HasSuffix(f, "package.json")) {
			for i := 0; i < len(deduped); i++ {
				for j := i + 1; j < len(deduped); j++ {
					tA, tB := deduped[i], deduped[j]
					c := TaskConflict{
						ConflictID:           fmt.Sprintf("CONF-COORD-%04d", confIdx),
						Type:                 ConflictCoordinationRequired,
						TaskID:               tA,
						ConflictingTaskID:    tB,
						ConflictingResource:  f,
						Explanation:          fmt.Sprintf("Tasks %s and %s modify shared configuration/manifest file %s", tA, tB, f),
						ResolutionSuggestion: "Execute sequentially in the same worktree and run manifest validation",
					}
					conflicts = append(conflicts, c)
					confIdx++
					taskConflictMap[tA] = append(taskConflictMap[tA], c)
					taskConflictMap[tB] = append(taskConflictMap[tB], c)
				}
			}
		}
	}

	// 3. SOFT CONFLICT: Multiple tasks modifying different symbols in the same source file
	for f, tIDs := range fileToTasks {
		deduped := uniqueStrings(tIDs)
		if len(deduped) > 1 && !strings.HasSuffix(f, "Dockerfile") && !strings.HasSuffix(f, "go.mod") && !strings.HasSuffix(f, "package.json") {
			for i := 0; i < len(deduped); i++ {
				for j := i + 1; j < len(deduped); j++ {
					tA, tB := deduped[i], deduped[j]
					// If already flagged as HARD conflict, skip
					alreadyHard := false
					for _, existing := range taskConflictMap[tA] {
						if existing.Type == ConflictHard && existing.ConflictingTaskID == tB {
							alreadyHard = true
							break
						}
					}
					if !alreadyHard {
						c := TaskConflict{
							ConflictID:           fmt.Sprintf("CONF-SOFT-%04d", confIdx),
							Type:                 ConflictSoft,
							TaskID:               tA,
							ConflictingTaskID:    tB,
							ConflictingResource:  f,
							Explanation:          fmt.Sprintf("Tasks %s and %s touch the same source file %s but target different functions", tA, tB, f),
							ResolutionSuggestion: "Safe to execute sequentially; verify line offset stability",
						}
						conflicts = append(conflicts, c)
						confIdx++
						taskConflictMap[tA] = append(taskConflictMap[tA], c)
						taskConflictMap[tB] = append(taskConflictMap[tB], c)
					}
				}
			}
		}
	}

	// Attach conflicts to tasks
	updatedTasks := make([]RemediationTask, len(tasks))
	copy(updatedTasks, tasks)
	for i := range updatedTasks {
		t := &updatedTasks[i]
		t.Conflicts = taskConflictMap[t.TaskID]
	}

	return conflicts, updatedTasks
}
