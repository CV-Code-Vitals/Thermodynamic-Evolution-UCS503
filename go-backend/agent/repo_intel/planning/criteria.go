package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// CriteriaGenerator creates objective, measurable acceptance criteria, validation commands,
// and rollback conditions for a remediation task.
type CriteriaGenerator struct{}

// NewCriteriaGenerator creates a new CriteriaGenerator.
func NewCriteriaGenerator() *CriteriaGenerator {
	return &CriteriaGenerator{}
}

// GenerateCriteria produces acceptance criteria, validation commands, and rollback conditions.
func (cg *CriteriaGenerator) GenerateCriteria(f *diagnostic.ValidatedFinding, targetFiles []string) ([]string, []string, []string) {
	var criteria []string
	var commands []string
	var rollback []string

	// General rollback safety invariant
	rollback = append(rollback, "Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress")

	switch f.Category {
	case repo_intel.CategorySecurity:
		if strings.Contains(f.FindingID, "SECRET") {
			criteria = append(criteria,
				"Hardcoded credential is completely removed from source code and git tracking",
				"Credential is verified revoked and rotated with cloud provider",
				"Application successfully loads replacement credential via environment variable or secret manager",
			)
			commands = append(commands, "git grep -n -i 'AKIA' || true")
		} else if strings.Contains(f.FindingID, "CMD-INJECT") {
			criteria = append(criteria,
				"Process execution arguments are passed as isolated tokens without shell string interpolation",
				"Execution unit test passes with malicious input without shell escape",
			)
			commands = append(commands, "go test -v ./...")
		} else if strings.Contains(f.FindingID, "DOCKER") {
			criteria = append(criteria,
				"Dockerfile specifies an unprivileged non-root USER instruction",
				"Container build succeeds and process runs as unprivileged user ID",
			)
			if len(targetFiles) > 0 {
				commands = append(commands, fmt.Sprintf("docker build -f %s .", targetFiles[0]))
			}
		}

	case repo_intel.CategoryArchitecture:
		criteria = append(criteria,
			"Circular dependency between packages is broken via interface extraction or package decomposition",
			"Topological dependency graph contains zero directed cycles",
			"All package unit tests pass cleanly after refactoring",
		)
		commands = append(commands, "go test ./...", "go build ./...")

	case repo_intel.CategoryStructural:
		criteria = append(criteria,
			"Target function cyclomatic complexity is reduced below threshold (M < 12)",
			"Control flow nesting depth is reduced to <= 3",
			"Existing package unit tests pass with zero regressions",
		)
		if len(targetFiles) > 0 {
			dir := filepath.Dir(targetFiles[0])
			commands = append(commands, fmt.Sprintf("go test -v ./%s", filepath.ToSlash(dir)))
		} else {
			commands = append(commands, "go test ./...")
		}

	case repo_intel.CategoryDependency:
		criteria = append(criteria,
			"All dependencies are pinned to immutable cryptographic digests or explicit versions",
			"Lockfile (go.sum, Cargo.lock, or package-lock.json) exists and is verified valid",
		)
		commands = append(commands, "go mod verify")

	case repo_intel.CategoryTestQuality:
		criteria = append(criteria,
			"Unit test file is added co-located with target production module",
			"Newly authored unit tests execute and pass cleanly in CI test runner",
		)
		if len(targetFiles) > 0 {
			dir := filepath.Dir(targetFiles[0])
			commands = append(commands, fmt.Sprintf("go test -v ./%s", filepath.ToSlash(dir)))
		}

	case repo_intel.CategoryReliability:
		criteria = append(criteria,
			"Error return values are handled or propagated rather than discarded",
			"No silent failure pathways exist in error-handling branch",
		)
		commands = append(commands, "go test ./...")
	}

	if len(criteria) == 0 {
		criteria = append(criteria,
			"Reported defect condition is eliminated from source file",
			"Module builds and all existing automated tests pass",
		)
		commands = append(commands, "go test ./...")
	}

	return criteria, commands, rollback
}
