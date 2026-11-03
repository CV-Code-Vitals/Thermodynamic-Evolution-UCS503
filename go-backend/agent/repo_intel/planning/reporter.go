package planning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PlanReporter serializes the machine-readable plan JSON, dependency graph, conflict report,
// and the human reviewable Markdown plan.
type PlanReporter struct {
	outputDir string
}

// NewPlanReporter creates a new PlanReporter.
func NewPlanReporter(outputDir string) *PlanReporter {
	return &PlanReporter{outputDir: outputDir}
}

// WritePlanOutputs writes all 4 Phase 3 deliverables to the output directory.
func (pr *PlanReporter) WritePlanOutputs(plan *RemediationPlan) error {
	if pr.outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(pr.outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// 1. reports/remediation_plan.json
	jsonBytes, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal remediation_plan.json: %w", err)
	}
	pPlan := filepath.Join(pr.outputDir, "remediation_plan.json")
	if err := os.WriteFile(pPlan, jsonBytes, 0644); err != nil {
		return fmt.Errorf("write %s: %w", pPlan, err)
	}

	// 2. reports/remediation_dependency_graph.json
	depGraphBytes, err := json.MarshalIndent(plan.DependencyGraph, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal remediation_dependency_graph.json: %w", err)
	}
	pDep := filepath.Join(pr.outputDir, "remediation_dependency_graph.json")
	if err := os.WriteFile(pDep, depGraphBytes, 0644); err != nil {
		return fmt.Errorf("write %s: %w", pDep, err)
	}

	// 3. reports/remediation_conflict_report.json
	type ConflictReport struct {
		TotalConflicts int            `json:"total_conflicts"`
		Conflicts      []TaskConflict `json:"conflicts"`
	}
	confReport := ConflictReport{
		TotalConflicts: len(plan.Conflicts),
		Conflicts:      plan.Conflicts,
	}
	confBytes, err := json.MarshalIndent(confReport, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal remediation_conflict_report.json: %w", err)
	}
	pConf := filepath.Join(pr.outputDir, "remediation_conflict_report.json")
	if err := os.WriteFile(pConf, confBytes, 0644); err != nil {
		return fmt.Errorf("write %s: %w", pConf, err)
	}

	// 4. reports/remediation_plan.md
	mdContent := pr.GenerateMarkdown(plan)
	pMd := filepath.Join(pr.outputDir, "remediation_plan.md")
	if err := os.WriteFile(pMd, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("write %s: %w", pMd, err)
	}

	return nil
}

// GenerateMarkdown creates a comprehensive reviewable Markdown remediation plan.
func (pr *PlanReporter) GenerateMarkdown(plan *RemediationPlan) string {
	var sb strings.Builder

	sb.WriteString("# Remediation Plan & Dependency-Ordered Execution Schedule\n\n")
	sb.WriteString(fmt.Sprintf("**Plan Schema Version:** `%s`  \n", plan.PlanSchemaVersion))
	sb.WriteString(fmt.Sprintf("**Plan Identifier:** `%s`  \n", plan.PlanID))
	sb.WriteString(fmt.Sprintf("**Generated At:** %s  \n", plan.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Baseline Commit:** `%s`  \n", plan.BaselineCommit))
	sb.WriteString(fmt.Sprintf("**Source Diagnostic Report ID:** `%s` (Hash: `%.12s...`)  \n", plan.SourceDiagnosticReportID, plan.SourceReportHash))
	sb.WriteString(fmt.Sprintf("**Repository Root:** `%s`  \n\n", plan.RepositoryIdentity.RepoRoot))

	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("The Remediation Planning Engine evaluated **%d diagnostic findings**, identified **%d actionable findings**, and generated a dependency-ordered plan comprising **%d distinct tasks**.\n\n",
		plan.Summary.TotalFindingsConsidered, plan.Summary.ActionableFindings, plan.Summary.TotalTasksGenerated))

	sb.WriteString("### Plan Composition by Execution Class\n\n")
	sb.WriteString("| Execution Class | Count | Description & Policy |\n|---|---|---|\n")
	sb.WriteString(fmt.Sprintf("| `AUTOMATION_READY` | %d | Deterministic modifications safe for automated execution |\n", plan.Summary.AutomationReadyTasks))
	sb.WriteString(fmt.Sprintf("| `AUTOMATION_WITH_REVIEW` | %d | Local refactoring tasks validated via automated test gates |\n", plan.Summary.AutomationWithReviewTasks))
	sb.WriteString(fmt.Sprintf("| `HUMAN_APPROVAL_REQUIRED` | %d | High-impact security or architectural changes requiring human sign-off |\n", plan.Summary.HumanApprovalTasks))
	sb.WriteString(fmt.Sprintf("| `INVESTIGATION_REQUIRED` | %d | Heuristic or ambiguous findings deferred pending manual review |\n", plan.Summary.InvestigationTasks))
	sb.WriteString(fmt.Sprintf("| `BLOCKED` | %d | Tasks blocked by circular dependencies or conflicting locks |\n\n", plan.Summary.BlockedTasks))

	sb.WriteString("### Conflict Analysis\n\n")
	sb.WriteString(fmt.Sprintf("Identified **%d total conflicts** across planned tasks.\n\n", plan.Summary.TotalConflictsDetected))
	if len(plan.Conflicts) > 0 {
		sb.WriteString("| Conflict ID | Type | Task A | Task B | Resource | Resolution Strategy |\n|---|---|---|---|---|---|\n")
		for _, c := range plan.Conflicts {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n",
				c.ConflictID, c.Type, c.TaskID, c.ConflictingTaskID, c.ConflictingResource, c.ResolutionSuggestion))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Dependency-Ordered Execution Schedule\n\n")
	sb.WriteString("The following sequence represents the deterministic topological execution order:\n\n")
	sb.WriteString("| Step | Task ID | Priority | Category | Strategy | Target Files | Approval Required |\n|---|---|---|---|---|---|---|\n")
	taskMap := make(map[string]RemediationTask)
	for _, t := range plan.Tasks {
		taskMap[t.TaskID] = t
	}

	for step, taskID := range plan.ExecutionOrder {
		t := taskMap[taskID]
		apprStr := "No"
		if t.HumanApprovalRequired {
			apprStr = "**YES**"
		}
		sb.WriteString(fmt.Sprintf("| %d | `%s` | `%s` | `%s` | `%s` | `%s` | %s |\n",
			step+1, t.TaskID, t.Priority, t.Category, t.ExecutionStrategy, strings.Join(t.TargetFiles, ", "), apprStr))
	}
	sb.WriteString("\n")

	sb.WriteString("## Detailed Remediation Tasks\n\n")
	for i, t := range plan.Tasks {
		sb.WriteString(fmt.Sprintf("### %d. [%s] %s (`%s`)\n\n", i+1, t.Priority, t.Title, t.TaskID))
		sb.WriteString(fmt.Sprintf("- **Category:** `%s` | **Execution Strategy:** `%s`\n", t.Category, t.ExecutionStrategy))
		sb.WriteString(fmt.Sprintf("- **Execution Class:** `%s` | **Priority Score:** `%.1f`\n", t.ExecutionClass, t.PriorityScore))
		sb.WriteString(fmt.Sprintf("- **Target Files:** %s\n", formatFileList(t.TargetFiles)))
		if len(t.TargetSymbols) > 0 {
			sb.WriteString(fmt.Sprintf("- **Target Symbols:** %s\n", strings.Join(t.TargetSymbols, ", ")))
		}
		sb.WriteString(fmt.Sprintf("- **Source Finding IDs:** `%s`\n", strings.Join(t.SourceFindingIDs, "`, `")))
		sb.WriteString(fmt.Sprintf("- **Regression Risk:** `%s`\n", t.RegressionRisk))
		if t.HumanApprovalRequired {
			sb.WriteString(fmt.Sprintf("- > [!CAUTION]\n> **Human Approval Required:** %s\n", t.ApprovalRationale))
		}
		if len(t.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("- **Prerequisite Tasks (Dependencies):** `%s`\n", strings.Join(t.Dependencies, "`, `")))
		}
		sb.WriteString(fmt.Sprintf("\n%s\n\n", t.Description))

		sb.WriteString("**Acceptance Criteria:**\n")
		for _, ac := range t.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("- [ ] %s\n", ac))
		}
		sb.WriteString("\n")

		if len(t.ValidationCommands) > 0 {
			sb.WriteString("**Validation Commands:**\n```bash\n")
			for _, cmd := range t.ValidationCommands {
				sb.WriteString(fmt.Sprintf("%s\n", cmd))
			}
			sb.WriteString("```\n\n")
		}

		if len(t.RollbackConditions) > 0 {
			sb.WriteString("**Rollback Conditions:**\n")
			for _, rc := range t.RollbackConditions {
				sb.WriteString(fmt.Sprintf("- %s\n", rc))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("---\n\n")
	}

	sb.WriteString("## Planning Limitations & Safety Notes\n\n")
	for _, lim := range plan.PlanningLimitations {
		sb.WriteString(fmt.Sprintf("- %s\n", lim))
	}
	sb.WriteString("\n")

	return sb.String()
}

func formatFileList(files []string) string {
	if len(files) == 0 {
		return "None"
	}
	var formatted []string
	for _, f := range files {
		formatted = append(formatted, fmt.Sprintf("`%s`", f))
	}
	return strings.Join(formatted, ", ")
}
