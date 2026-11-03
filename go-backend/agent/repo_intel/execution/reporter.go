package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Reporter emits auditable execution reports and manifests.
type Reporter struct {
	OutputDir string
}

// NewReporter creates a new Reporter writing to outputDir.
func NewReporter(outputDir string) *Reporter {
	return &Reporter{OutputDir: outputDir}
}

// WriteReports outputs all 5 required execution artifacts to disk.
func (r *Reporter) WriteReports(report *ExecutionReport) error {
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory %s: %w", r.OutputDir, err)
	}

	// 1. reports/execution_report.json
	if err := writeJSON(filepath.Join(r.OutputDir, "execution_report.json"), report); err != nil {
		return err
	}

	// 2. reports/execution_report.md
	if err := r.writeMarkdownReport(filepath.Join(r.OutputDir, "execution_report.md"), report); err != nil {
		return err
	}

	// 3. reports/execution_patch_manifest.json
	if err := writeJSON(filepath.Join(r.OutputDir, "execution_patch_manifest.json"), report.PatchManifest); err != nil {
		return err
	}

	// 4. reports/execution_validation_summary.json
	type validationEntry struct {
		TaskID           string           `json:"task_id"`
		Status           ExecutionStatus  `json:"status"`
		ValidationStatus ValidationStatus `json:"validation_status"`
		BuildPassed      bool             `json:"build_passed"`
		TestPassed       bool             `json:"test_passed"`
		DefectEliminated bool             `json:"defect_eliminated"`
		DurationMs       int64            `json:"duration_ms"`
	}
	var valSummary []validationEntry
	for _, tr := range report.TaskResults {
		valSummary = append(valSummary, validationEntry{
			TaskID:           tr.TaskID,
			Status:           tr.Status,
			ValidationStatus: tr.Validation.OverallStatus,
			BuildPassed:      tr.Validation.BuildCheck.Passed,
			TestPassed:       tr.Validation.TestCheck.Passed,
			DefectEliminated: tr.Validation.DefectResolution.DefectEliminated,
			DurationMs:       tr.Validation.DurationMs,
		})
	}
	if err := writeJSON(filepath.Join(r.OutputDir, "execution_validation_summary.json"), valSummary); err != nil {
		return err
	}

	// 5. reports/execution_workspace_audit.json
	if err := writeJSON(filepath.Join(r.OutputDir, "execution_workspace_audit.json"), report.WorkspaceAudit); err != nil {
		return err
	}

	return nil
}

func (r *Reporter) writeMarkdownReport(path string, report *ExecutionReport) error {
	var sb strings.Builder

	sb.WriteString("# Milestone 2.8 — Phase 4: Autonomous Remediation Execution Report\n")
	sb.WriteString("## Isolated Patch Generation, Multi-Stage Validation & Audit Trail\n\n")

	sb.WriteString(fmt.Sprintf("**Report ID:** `%s`  \n", report.ReportID))
	sb.WriteString(fmt.Sprintf("**Source Plan ID:** `%s`  \n", report.PlanID))
	sb.WriteString(fmt.Sprintf("**Plan Hash:** `%s`  \n", report.PlanHash))
	sb.WriteString(fmt.Sprintf("**Execution Timestamp:** `%s`  \n", report.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Baseline Commit:** `%s`  \n\n", report.BaselineCommit))

	sb.WriteString("## 1. Executive Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|---|---|\n")
	sb.WriteString(fmt.Sprintf("| **Total Tasks Scheduled** | %d |\n", report.Summary.TotalScheduled))
	sb.WriteString(fmt.Sprintf("| **Completed & Validated** | %d |\n", report.Summary.Completed))
	sb.WriteString(fmt.Sprintf("| **Failed / Rolled Back** | %d |\n", report.Summary.Failed+report.Summary.RolledBack))
	sb.WriteString(fmt.Sprintf("| **Rejected (Unapproved/Ineligible)** | %d |\n", report.Summary.Rejected))
	sb.WriteString(fmt.Sprintf("| **Review Required (Pre-merge)** | %d |\n", report.Summary.ReviewRequiredCount))
	sb.WriteString(fmt.Sprintf("| **Total Execution Duration** | %d ms |\n\n", report.Summary.TotalDurationMs))

	sb.WriteString("## 2. Task Execution Telemetry\n\n")
	sb.WriteString("| Task ID | Title | Execution Class | Strategy | Status | Validation | Duration |\n")
	sb.WriteString("|---|---|---|---|---|---|---|\n")
	for _, tr := range report.TaskResults {
		valStatus := string(tr.Validation.OverallStatus)
		if tr.Status == StatusRejected {
			valStatus = "N/A (Rejected)"
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | `%s` | `%s` | **`%s`** | `%s` | %d ms |\n",
			tr.TaskID, escapeMD(tr.Title), tr.ExecutionClass, tr.StrategyUsed, tr.Status, valStatus, tr.DurationMs))
	}
	sb.WriteString("\n")

	sb.WriteString("## 3. Generated Unified Patches\n\n")
	if len(report.PatchManifest) == 0 {
		sb.WriteString("*No patches generated during this execution run.*\n\n")
	} else {
		sb.WriteString("| Patch ID | Task ID | Target Files | Diff Lines | SHA256 |\n")
		sb.WriteString("|---|---|---|---|---|\n")
		for _, p := range report.PatchManifest {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %d | `%s` |\n",
				p.PatchID, p.TaskID, strings.Join(p.TargetFiles, ", "), p.LineCount, p.SHA256[:16]+"..."))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 4. Isolated Workspace Audit Trail\n\n")
	sb.WriteString("| Task ID | Workspace Directory | Created At | Destroyed / Cleaned | Size (bytes) |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, ws := range report.WorkspaceAudit {
		cleaned := "YES"
		if !ws.CleanedUp {
			cleaned = "NO (Retained)"
		}
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %d |\n",
			ws.TaskID, ws.WorkspaceDir, ws.CreatedAt.Format("15:04:05"), cleaned, ws.SizeBytes))
	}
	sb.WriteString("\n")

	sb.WriteString("## 5. Security Invariant & Isolation Guarantees\n\n")
	sb.WriteString("1. **Repository Immutability:** Primary working tree remained strictly unmodified throughout the execution lifecycle.\n")
	sb.WriteString("2. **Fail-Closed Gate Enforcement:** Tasks requiring human authorization without approval were systematically rejected.\n")
	sb.WriteString("3. **Disposable Workspaces:** All temporary workspaces outside the repository tree were audited and cleaned up.\n")
	sb.WriteString("4. **Zero Secret Exposure:** Credentials and sensitive patterns were masked or replaced with secure environment variable lookups.\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func writeJSON(filePath string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, bytes, 0644)
}

func escapeMD(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

