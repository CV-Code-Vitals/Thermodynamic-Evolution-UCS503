package repo_intel

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Aggregator consolidates, deduplicates, and sorts findings across all analysis modules,
// producing comprehensive JSON and Markdown reports.
type Aggregator struct{}

// NewAggregator creates a new Aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// SeverityWeight maps severities to sorting weights (higher = more critical).
var SeverityWeight = map[Severity]int{
	SeverityCritical:      5,
	SeverityHigh:          4,
	SeverityMedium:        3,
	SeverityLow:           2,
	SeverityInformational: 1,
}

// Consolidate merges findings, deduplicates them, and populates the IntelligenceReport.
func (a *Aggregator) Consolidate(
	registry RepositoryMetadataRegistry,
	graph DependencyGraph,
	findings []Finding,
	coverage AnalysisCoverage,
	metrics PerformanceMetrics,
	warnings []string,
) *IntelligenceReport {
	// Deduplicate findings by FindingID
	seenIDs := make(map[string]bool)
	var deduped []Finding

	for _, f := range findings {
		if f.FindingID == "" {
			f.FindingID = fmt.Sprintf("RI-%s-%s", string(f.Category), sanitizeID(f.Title))
		}
		if seenIDs[f.FindingID] {
			continue
		}
		seenIDs[f.FindingID] = true
		deduped = append(deduped, f)
	}

	// Sort deterministically: Severity desc, Category asc, Title asc, FindingID asc
	sort.Slice(deduped, func(i, j int) bool {
		wi, wj := SeverityWeight[deduped[i].Severity], SeverityWeight[deduped[j].Severity]
		if wi != wj {
			return wi > wj
		}
		if deduped[i].Category != deduped[j].Category {
			return deduped[i].Category < deduped[j].Category
		}
		if deduped[i].Title != deduped[j].Title {
			return deduped[i].Title < deduped[j].Title
		}
		return deduped[i].FindingID < deduped[j].FindingID
	})

	return &IntelligenceReport{
		Version:            "1.0.0",
		GeneratedAt:        time.Now().UTC(),
		RepositoryMetadata: registry,
		DependencyGraph:    graph,
		Findings:           deduped,
		AnalysisCoverage:   coverage,
		PerformanceMetrics: metrics,
		Warnings:           warnings,
	}
}

// GenerateJSON serializes the IntelligenceReport to pretty-printed JSON.
func (a *Aggregator) GenerateJSON(report *IntelligenceReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// GenerateMarkdown creates an executive and technical Markdown report.
func (a *Aggregator) GenerateMarkdown(report *IntelligenceReport) string {
	var sb strings.Builder

	// Calculate counts
	bySeverity := make(map[Severity]int)
	byCategory := make(map[Category]int)
	byStatus := make(map[VerificationStatus]int)

	for _, f := range report.Findings {
		bySeverity[f.Severity]++
		byCategory[f.Category]++
		byStatus[f.VerificationStatus]++
	}

	sb.WriteString("# Repository Intelligence Diagnostic Report\n\n")
	sb.WriteString(fmt.Sprintf("**Version:** `%s`  \n", report.Version))
	sb.WriteString(fmt.Sprintf("**Generated At:** %s  \n", report.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Repository Root:** `%s`  \n", report.RepositoryMetadata.RepoRoot))
	sb.WriteString(fmt.Sprintf("**Git Branch:** `%s` (Commit: `%s`, Clean: `%t`)  \n", report.RepositoryMetadata.GitBranch, report.RepositoryMetadata.HeadCommit, report.RepositoryMetadata.WorkingTreeClean))
	sb.WriteString(fmt.Sprintf("**Total Files Discovered:** %d  \n", report.RepositoryMetadata.TotalFiles))
	sb.WriteString(fmt.Sprintf("**Total Duration:** %d ms  \n\n", report.PerformanceMetrics.TotalDurationMs))

	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("The autonomous Repository Intelligence Agent completed whole-repository discovery and multi-dimensional analysis. A total of **%d findings** were identified across %d files.\n\n", len(report.Findings), report.AnalysisCoverage.StructuralFilesCount))

	sb.WriteString("| Severity | Count |\n|---|---|\n")
	sb.WriteString(fmt.Sprintf("| Critical | %d |\n", bySeverity[SeverityCritical]))
	sb.WriteString(fmt.Sprintf("| High | %d |\n", bySeverity[SeverityHigh]))
	sb.WriteString(fmt.Sprintf("| Medium | %d |\n", bySeverity[SeverityMedium]))
	sb.WriteString(fmt.Sprintf("| Low | %d |\n", bySeverity[SeverityLow]))
	sb.WriteString(fmt.Sprintf("| Informational | %d |\n\n", bySeverity[SeverityInformational]))

	sb.WriteString("### Findings by Category\n\n")
	sb.WriteString("| Category | Count |\n|---|---|\n")
	for cat, count := range byCategory {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
	}
	sb.WriteString("\n")

	sb.WriteString("### Detected Languages & Code Footprint\n\n")
	sb.WriteString("| Language | Files | Code Lines | Total Size | Parsed / Errors |\n|---|---|---|---|---|\n")
	var langs []string
	for l := range report.RepositoryMetadata.Languages {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	for _, l := range langs {
		st := report.RepositoryMetadata.Languages[l]
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %.1f KB | %d / %d |\n", l, st.FileCount, st.TotalLines, float64(st.TotalBytes)/1024.0, st.ParsedCount, st.ParseErrorsCount))
	}
	sb.WriteString("\n")

	sb.WriteString("### Dependency Manifests & Ecosystems\n\n")
	if len(report.RepositoryMetadata.Manifests) == 0 {
		sb.WriteString("No dependency manifests detected.\n\n")
	} else {
		sb.WriteString("| Manifest File | Ecosystem | Dependencies | Lockfile Present |\n|---|---|---|---|\n")
		for _, m := range report.RepositoryMetadata.Manifests {
			lockStr := "No"
			if m.HasLockfile {
				lockStr = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %d | %s |\n", m.FilePath, m.ManifestType, len(m.Dependencies), lockStr))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Architecture & Dependency Graph\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Nodes (Files):** %d\n", len(report.DependencyGraph.Nodes)))
	sb.WriteString(fmt.Sprintf("- **Total Directed Edges (Dependencies):** %d\n", len(report.DependencyGraph.Edges)))
	sb.WriteString(fmt.Sprintf("- **Discovered Service Boundaries:** %d\n\n", len(report.RepositoryMetadata.ServiceBoundaries)))

	sb.WriteString("## Detailed Findings\n\n")
	if len(report.Findings) == 0 {
		sb.WriteString("No structural, architectural, or security defects identified.\n")
	} else {
		for i, f := range report.Findings {
			sb.WriteString(fmt.Sprintf("### %d. [%s] %s (`%s`)\n\n", i+1, f.Severity, f.Title, f.FindingID))
			sb.WriteString(fmt.Sprintf("- **Category:** `%s`\n", f.Category))
			sb.WriteString(fmt.Sprintf("- **Verification Status:** `%s` (Confidence: `%.2f`)\n", f.VerificationStatus, f.Confidence))
			sb.WriteString(fmt.Sprintf("- **Detection Method:** `%s`\n", f.DetectionMethod))
			sb.WriteString(fmt.Sprintf("- **Root Cause:** %s\n\n", f.RootCause))
			sb.WriteString(fmt.Sprintf("%s\n\n", f.Description))

			if len(f.Evidence) > 0 {
				sb.WriteString("**Evidence Locations:**\n\n")
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
	}

	sb.WriteString("## Performance Telemetry\n\n")
	sb.WriteString(fmt.Sprintf("- **Discovery Duration:** %d ms\n", report.PerformanceMetrics.DiscoveryDurationMs))
	sb.WriteString(fmt.Sprintf("- **Manifest Analysis Duration:** %d ms\n", report.PerformanceMetrics.ManifestDurationMs))
	sb.WriteString(fmt.Sprintf("- **Structural AST Duration:** %d ms\n", report.PerformanceMetrics.StructuralDurationMs))
	sb.WriteString(fmt.Sprintf("- **Architecture Graph Duration:** %d ms\n", report.PerformanceMetrics.ArchitecturalDurationMs))
	sb.WriteString(fmt.Sprintf("- **Security Duration:** %d ms\n", report.PerformanceMetrics.SecurityDurationMs))
	sb.WriteString(fmt.Sprintf("- **Quality Duration:** %d ms\n", report.PerformanceMetrics.QualityDurationMs))
	sb.WriteString(fmt.Sprintf("- **Total Pipeline Runtime:** %d ms\n", report.PerformanceMetrics.TotalDurationMs))

	return sb.String()
}
