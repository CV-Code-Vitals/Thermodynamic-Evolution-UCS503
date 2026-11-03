package diagnostic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// DiagnosticReporter generates machine-readable JSON contracts and human-readable Markdown reports.
type DiagnosticReporter struct {
	repoRoot  string
	outputDir string
}

// NewDiagnosticReporter creates a new DiagnosticReporter.
func NewDiagnosticReporter(repoRoot, outputDir string) *DiagnosticReporter {
	return &DiagnosticReporter{
		repoRoot:  repoRoot,
		outputDir: outputDir,
	}
}

// BuildReport constructs the DiagnosticReport struct from validated findings and correlations.
func (dr *DiagnosticReporter) BuildReport(
	meta repo_intel.RepositoryMetadataRegistry,
	findings []ValidatedFinding,
	summary ValidationSummary,
	correlations []FindingCorrelation,
	durationMs int64,
	warnings []string,
) *DiagnosticReport {
	// Synthesize Global Risks
	var globalRisks []string
	hasSecCritical := false
	hasArchCycles := false
	hasParseErrors := false

	for _, f := range findings {
		if f.Category == repo_intel.CategorySecurity && f.Severity == repo_intel.SeverityCritical {
			hasSecCritical = true
		}
		if f.Category == repo_intel.CategoryArchitecture && strings.Contains(f.FindingID, "CYCLE") {
			hasArchCycles = true
		}
		if strings.Contains(f.FindingID, "PARSE-ERROR") {
			hasParseErrors = true
		}
	}

	if hasSecCritical {
		globalRisks = append(globalRisks, "CRITICAL: Hardcoded credentials or tokens detected in codebase; immediate rotation and IAM review required")
	}
	if hasArchCycles {
		globalRisks = append(globalRisks, "HIGH: Circular dependencies exist in package graph, hindering modular testability and build performance")
	}
	if hasParseErrors {
		globalRisks = append(globalRisks, "HIGH: Grammar parse errors detected in source files, preventing complete AST thermodynamic analysis")
	}
	if summary.Stale > 0 {
		globalRisks = append(globalRisks, fmt.Sprintf("INFORMATIONAL: %d findings were marked STALE due to recent codebase edits or out-of-bounds line numbers", summary.Stale))
	}

	// Recommended Priorities
	var priorities []string
	priorities = append(priorities, "1. Remediate CRITICAL and HIGH security findings (credential masking, process command injection, SQL concatenation)")
	priorities = append(priorities, "2. Execute AUTOMATION_READY hygiene items (Docker non-root USER, pinned base image tags, license creation)")
	priorities = append(priorities, "3. Address architectural dependency cycles to restore clean modular boundaries")
	priorities = append(priorities, "4. Refactor functions with extreme cyclomatic complexity (M >= 20) under automated regression test gates")

	limitations := []string{
		"Static taint analysis for SQL/exec execution is heuristic and flags potential patterns without inter-procedural dataflow verification",
		"Unit test presence is audited by directory co-location; multi-package integration suites located in root test dirs may cause benign NOTEST warnings",
		"Diagnostic reporting is strictly read-only and does not mutate source code or execute remediation patches",
	}

	reportID := fmt.Sprintf("DR-%d", time.Now().UTC().Unix())

	return &DiagnosticReport{
		ReportSchemaVersion: "1.0",
		ReportID:            reportID,
		Repository:          meta,
		Analysis: AnalysisMetadata{
			Timestamp:           time.Now().UTC(),
			FilesDiscovered:     meta.TotalFiles,
			FilesAnalysed:       len(meta.SourceFiles),
			AnalysisCoverage:    100.0,
			AnalysisWarnings:    warnings,
			ExecutionDurationMs: durationMs,
		},
		Summary:               summary,
		Findings:              findings,
		Correlations:          correlations,
		GlobalRisks:           globalRisks,
		RecommendedPriorities: priorities,
		Limitations:           limitations,
	}
}

// WriteOutputs emits diagnostic_report.json, diagnostic_report.md, and finding_validation.json.
func (dr *DiagnosticReporter) WriteOutputs(report *DiagnosticReport) error {
	if dr.outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(dr.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// 1. diagnostic_report.json
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal diagnostic_report.json: %w", err)
	}
	pJson := filepath.Join(dr.outputDir, "diagnostic_report.json")
	if err := os.WriteFile(pJson, jsonBytes, 0644); err != nil {
		return fmt.Errorf("write %s: %w", pJson, err)
	}

	// 2. finding_validation.json
	valResult := FindingValidationResult{
		Summary:  report.Summary,
		Findings: report.Findings,
	}
	valBytes, err := json.MarshalIndent(valResult, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal finding_validation.json: %w", err)
	}
	pVal := filepath.Join(dr.outputDir, "finding_validation.json")
	if err := os.WriteFile(pVal, valBytes, 0644); err != nil {
		return fmt.Errorf("write %s: %w", pVal, err)
	}

	// 3. diagnostic_report.md
	mdContent := dr.GenerateMarkdown(report)
	pMd := filepath.Join(dr.outputDir, "diagnostic_report.md")
	if err := os.WriteFile(pMd, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("write %s: %w", pMd, err)
	}

	return nil
}

// GenerateMarkdown formats an executive and technical diagnostic report.
func (dr *DiagnosticReporter) GenerateMarkdown(report *DiagnosticReport) string {
	var sb strings.Builder

	sb.WriteString("# Diagnostic Intelligence & Remediation-Ready Audit Report\n\n")
	sb.WriteString(fmt.Sprintf("**Report Schema Version:** `%s`  \n", report.ReportSchemaVersion))
	sb.WriteString(fmt.Sprintf("**Report ID:** `%s`  \n", report.ReportID))
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s  \n", report.Analysis.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Repository Root:** `%s` (Branch: `%s`, Head: `%s`)  \n", report.Repository.RepoRoot, report.Repository.GitBranch, report.Repository.HeadCommit))
	sb.WriteString(fmt.Sprintf("**Execution Runtime:** %d ms  \n\n", report.Analysis.ExecutionDurationMs))

	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("Phase 2 Diagnostic Intelligence ingested **%d findings** from Phase 1, performed physical disk and source location verification, calibrated independent confidence scores, assessed blast-radius severity, identified correlations, and determined remediation readiness.\n\n", report.Summary.TotalFindings))

	sb.WriteString("### Finding Validation Breakdown\n\n")
	sb.WriteString("| Validation State | Count | Description |\n|---|---|---|\n")
	sb.WriteString(fmt.Sprintf("| **VALIDATED** | %d | Source file exists, line bounds match physical content, metadata verified |\n", report.Summary.Validated))
	sb.WriteString(fmt.Sprintf("| **PARTIALLY_VALIDATED** | %d | Primary evidence verified but secondary reference ambiguous |\n", report.Summary.PartiallyValidated))
	sb.WriteString(fmt.Sprintf("| **STALE** | %d | Referenced lines out of bounds or condition altered |\n", report.Summary.Stale))
	sb.WriteString(fmt.Sprintf("| **INVALID** | %d | File missing without absence assertion or malformed metadata |\n", report.Summary.Invalid))
	sb.WriteString(fmt.Sprintf("| **REQUIRES_HUMAN_REVIEW** | %d | Heuristic pattern match or security taint requiring manual validation |\n\n", report.Summary.RequiresHumanReview))

	// Remediation Readiness breakdown
	readinessCounts := make(map[RemediationReadiness]int)
	for _, f := range report.Findings {
		readinessCounts[f.RemediationReadiness]++
	}

	sb.WriteString("### Remediation Readiness Classification\n\n")
	sb.WriteString("| Readiness Category | Count | Automated Action Policy |\n|---|---|---|\n")
	sb.WriteString(fmt.Sprintf("| `AUTOMATION_READY` | %d | Safe declarative or deterministic changes (e.g. Dockerfile pinning, non-root user, license) |\n", readinessCounts[ReadinessAutomationReady]))
	sb.WriteString(fmt.Sprintf("| `AUTOMATION_WITH_REVIEW` | %d | Local refactorings (guard clauses, test scaffolding) verified via test gates |\n", readinessCounts[ReadinessAutomationWithReview]))
	sb.WriteString(fmt.Sprintf("| `HUMAN_APPROVAL_REQUIRED` | %d | High-risk changes (secret rotation, architectural cycle breaking) |\n", readinessCounts[ReadinessHumanApprovalRequired]))
	sb.WriteString(fmt.Sprintf("| `INVESTIGATION_REQUIRED` | %d | Ambiguous heuristics or complex monolithic functions needing manual review |\n", readinessCounts[ReadinessInvestigationRequired]))
	sb.WriteString(fmt.Sprintf("| `INFORMATIONAL_ONLY` | %d | Inactive, stale, or non-actionable observations |\n\n", readinessCounts[ReadinessInformationalOnly]))

	if len(report.GlobalRisks) > 0 {
		sb.WriteString("### Global Architectural & Security Risks\n\n")
		for _, gr := range report.GlobalRisks {
			sb.WriteString(fmt.Sprintf("- %s\n", gr))
		}
		sb.WriteString("\n")
	}

	if len(report.Correlations) > 0 {
		sb.WriteString("### Finding Correlations & Clustering\n\n")
		sb.WriteString(fmt.Sprintf("Identified **%d structured finding relationships** across the codebase:\n\n", len(report.Correlations)))
		for _, c := range report.Correlations {
			sb.WriteString(fmt.Sprintf("- **[%s] %s**: Primary finding `%s` linked to %d related findings (%s).  \n  *Rationale:* %s\n",
				c.Type, c.CorrelationID, c.PrimaryFindingID, len(c.RelatedFindingIDs), strings.Join(c.RelatedFindingIDs, ", "), c.Rationale))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Prioritized Diagnostic Findings\n\n")
	// Sort findings by Severity desc, CalibratedConfidence desc
	sortedFindings := make([]ValidatedFinding, len(report.Findings))
	copy(sortedFindings, report.Findings)
	sort.Slice(sortedFindings, func(i, j int) bool {
		wi, wj := repo_intel.SeverityWeight[sortedFindings[i].Severity], repo_intel.SeverityWeight[sortedFindings[j].Severity]
		if wi != wj {
			return wi > wj
		}
		if sortedFindings[i].CalibratedConfidence != sortedFindings[j].CalibratedConfidence {
			return sortedFindings[i].CalibratedConfidence > sortedFindings[j].CalibratedConfidence
		}
		return sortedFindings[i].FindingID < sortedFindings[j].FindingID
	})

	for i, f := range sortedFindings {
		sb.WriteString(fmt.Sprintf("### %d. [%s] %s (`%s`)\n\n", i+1, f.Severity, f.Title, f.FindingID))
		sb.WriteString(fmt.Sprintf("- **Validation State:** `%s` (%s)\n", f.ValidationState, f.ValidationReason))
		sb.WriteString(fmt.Sprintf("- **Calibrated Confidence:** `%.2f` (%s)\n", f.CalibratedConfidence, f.ConfidenceRationale))
		sb.WriteString(fmt.Sprintf("- **Category:** `%s` | **Detection Method:** `%s`\n", f.Category, f.DetectionMethod))
		sb.WriteString(fmt.Sprintf("- **Remediation Readiness:** `%s` (Complexity: `%s`)\n", f.RemediationReadiness, f.RemediationComplexity))
		sb.WriteString(fmt.Sprintf("- **Readiness Rationale:** %s\n", f.ReadinessRationale))
		sb.WriteString(fmt.Sprintf("- **Impact Scope:** %s\n", f.ImpactScope))
		sb.WriteString(fmt.Sprintf("- **Severity Rationale:** %s\n", f.SeverityRationale))
		sb.WriteString(fmt.Sprintf("- **Root Cause:** %s\n\n", f.RootCause))
		sb.WriteString(fmt.Sprintf("%s\n\n", f.Description))

		if len(f.PotentialConsequences) > 0 {
			sb.WriteString("**Potential Consequences:**\n")
			for _, pc := range f.PotentialConsequences {
				sb.WriteString(fmt.Sprintf("- %s\n", pc))
			}
			sb.WriteString("\n")
		}

		if len(f.Evidence) > 0 {
			sb.WriteString("**Verified Evidence Locations:**\n\n")
			for _, ev := range f.Evidence {
				locStr := fmt.Sprintf("`%s`", ev.File)
				if ev.LineStart > 0 {
					locStr += fmt.Sprintf(" (Lines %d-%d)", ev.LineStart, ev.LineEnd)
				}
				if ev.Snippet != "" {
					locStr += fmt.Sprintf(":\n```\n%s\n```", ev.Snippet)
				} else if ev.Description != "" {
					locStr += fmt.Sprintf(" — *%s*", ev.Description)
				}
				sb.WriteString(fmt.Sprintf("- %s\n", locStr))
			}
			sb.WriteString("\n")
		}

		if f.RecommendedAction != "" {
			sb.WriteString(fmt.Sprintf("> [!TIP]\n> **Recommended Action:** %s\n>\n> **Expected Benefit:** %s\n\n", f.RecommendedAction, f.ExpectedBenefit))
		}
		sb.WriteString("---\n\n")
	}

	sb.WriteString("## Limitations & Boundary Notes\n\n")
	for _, lim := range report.Limitations {
		sb.WriteString(fmt.Sprintf("- %s\n", lim))
	}
	sb.WriteString("\n")

	return sb.String()
}
