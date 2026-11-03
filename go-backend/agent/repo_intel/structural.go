package repo_intel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent"
)

// StructuralAnalyzer wraps the compiled Rust Tree-sitter thermodynamic engine
// to extract repository-wide AST metrics, hotspots, and structural issues.
type StructuralAnalyzer struct {
	engineBinary string
	repoRoot     string
}

func NewStructuralAnalyzer(engineBinary, repoRoot string) *StructuralAnalyzer {
	return &StructuralAnalyzer{
		engineBinary: engineBinary,
		repoRoot:     repoRoot,
	}
}

// AnalyzeStructural runs the Rust engine and translates AST metrics into repository findings.
func (sa *StructuralAnalyzer) AnalyzeStructural(ctx context.Context, timeout time.Duration) (*agent.ThermodynamicReport, []Finding, error) {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	tempDir, err := os.MkdirTemp("", "ucs503-struct-audit-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp dir for structural report: %w", err)
	}
	defer os.RemoveAll(tempDir)

	reportPath := filepath.Join(tempDir, "thermodynamic_report.json")

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, sa.engineBinary, sa.repoRoot, "--output", reportPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("thermodynamic-ast-engine execution failed (%v): %s", err, string(out))
	}

	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read thermodynamic report: %w", err)
	}

	var report agent.ThermodynamicReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		return nil, nil, fmt.Errorf("unmarshal thermodynamic report: %w", err)
	}

	findings := sa.convertReportToFindings(&report)
	return &report, findings, nil
}

func (sa *StructuralAnalyzer) convertReportToFindings(report *agent.ThermodynamicReport) []Finding {
	var findings []Finding

	for _, fr := range report.FileReports {
		relPath := filepath.ToSlash(fr.FilePath)
		rootSlash := filepath.ToSlash(sa.repoRoot)
		if strings.HasPrefix(strings.ToLower(relPath), strings.ToLower(rootSlash)) {
			relPath = strings.TrimPrefix(relPath, rootSlash)
			relPath = strings.TrimPrefix(relPath, "/")
		}

		// Parse errors
		if fr.ParseErrors {
			findings = append(findings, Finding{
				FindingID:   fmt.Sprintf("RI-STR-PARSE-ERROR-%s", sanitizeID(relPath)),
				Category:    CategoryStructural,
				Severity:    SeverityMedium,
				Confidence:  1.0,
				Title:       "AST parse error encountered",
				Description: fmt.Sprintf("Source file %s contains syntax or grammar parse errors preventing complete AST analysis.", relPath),
				RootCause:   "Syntax error or unsupported grammar construct.",
				AffectedFiles: []string{relPath},
				Evidence: []EvidenceLocation{
					{File: relPath, Description: "Tree-sitter failed to parse file completely"},
				},
				RecommendedAction:      "Inspect file syntax and ensure compatibility with standard language grammars.",
				ExpectedBenefit:        "Enables full structural AST analysis and prevents compiler failures.",
				RegressionRisk:         SeverityLow,
				ValidationRequirements: []string{"File parses cleanly with zero parse errors"},
				DetectionMethod:        "TREE_SITTER_AST",
				VerificationStatus:     StatusConfirmed,
			})
		}

		// Evaluate Hotspots
		for _, hs := range fr.Hotspots {
			// Deep Nesting
			if hs.Metrics.MaxNestingDepth >= 4 {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-STR-HIGH-NESTING-%s-%s-L%d", sanitizeID(relPath), hs.FunctionName, hs.LineNumber),
					Category:    CategoryStructural,
					Severity:    SeverityMedium,
					Confidence:  0.95,
					Title:       fmt.Sprintf("Excessive control flow nesting depth (%d)", hs.Metrics.MaxNestingDepth),
					Description: fmt.Sprintf("Function %q in %s reaches nesting depth %d (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.", hs.FunctionName, relPath, hs.Metrics.MaxNestingDepth),
					RootCause:   "Deeply nested if/for/switch blocks without early returns or helper decomposition.",
					AffectedFiles: []string{relPath},
					AffectedSymbols: []string{hs.FunctionName},
					Evidence: []EvidenceLocation{
						{
							File:        relPath,
							LineStart:   hs.LineNumber,
							LineEnd:     hs.LineNumber,
							ByteStart:   hs.ByteRange[0],
							ByteEnd:     hs.ByteRange[1],
							Description: fmt.Sprintf("Nesting depth %d, Cyclomatic complexity %d", hs.Metrics.MaxNestingDepth, hs.Metrics.CyclomaticComplexity),
							Snippet:     hs.SourceSnippet,
						},
					},
					RecommendedAction:      "Refactor with guard clauses, early returns, or decompose into helper functions.",
					ExpectedBenefit:        "Lowers thermodynamic structural energy and simplifies test branch coverage.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"Function nesting depth reduced to <= 3", "Unit tests pass"},
					DetectionMethod:        "TREE_SITTER_AST",
					VerificationStatus:     StatusConfirmed,
				})
			}

			// High Cyclomatic Complexity
			if hs.Metrics.CyclomaticComplexity >= 12 {
				sev := SeverityMedium
				if hs.Metrics.CyclomaticComplexity >= 20 {
					sev = SeverityHigh
				}
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-STR-HIGH-COMPLEXITY-%s-%s-L%d", sanitizeID(relPath), hs.FunctionName, hs.LineNumber),
					Category:    CategoryStructural,
					Severity:    sev,
					Confidence:  0.95,
					Title:       fmt.Sprintf("Excessive cyclomatic complexity (M=%d)", hs.Metrics.CyclomaticComplexity),
					Description: fmt.Sprintf("Function %q has cyclomatic complexity of %d, requiring at least %d branch test cases to achieve full coverage.", hs.FunctionName, hs.Metrics.CyclomaticComplexity, hs.Metrics.CyclomaticComplexity),
					RootCause:   "Accumulation of branching logic and conditional checks in a single function.",
					AffectedFiles: []string{relPath},
					AffectedSymbols: []string{hs.FunctionName},
					Evidence: []EvidenceLocation{
						{
							File:        relPath,
							LineStart:   hs.LineNumber,
							LineEnd:     hs.LineNumber,
							Description: fmt.Sprintf("Cyclomatic complexity: %d, AST nodes: %d", hs.Metrics.CyclomaticComplexity, hs.Metrics.ASTNodeCount),
							Snippet:     hs.SourceSnippet,
						},
					},
					RecommendedAction:      "Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.",
					ExpectedBenefit:        "Improves maintainability and reduces likelihood of latent edge-case bugs.",
					RegressionRisk:         SeverityMedium,
					ValidationRequirements: []string{"Cyclomatic complexity reduced below threshold", "Regression test suite passes"},
					DetectionMethod:        "TREE_SITTER_AST",
					VerificationStatus:     StatusConfirmed,
				})
			}

			// Large AST node volume
			if hs.Metrics.ASTNodeCount >= 400 {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-STR-LARGE-AST-VOLUME-%s-%s-L%d", sanitizeID(relPath), hs.FunctionName, hs.LineNumber),
					Category:    CategoryStructural,
					Severity:    SeverityLow,
					Confidence:  0.90,
					Title:       fmt.Sprintf("Large function AST node volume (%d nodes)", hs.Metrics.ASTNodeCount),
					Description: fmt.Sprintf("Function %q contains %d AST nodes, indicating monolithic implementation.", hs.FunctionName, hs.Metrics.ASTNodeCount),
					RootCause:   "Function fulfills multiple responsibilities violating Single Responsibility Principle.",
					AffectedFiles: []string{relPath},
					AffectedSymbols: []string{hs.FunctionName},
					Evidence: []EvidenceLocation{
						{
							File:        relPath,
							LineStart:   hs.LineNumber,
							LineEnd:     hs.LineNumber,
							Description: fmt.Sprintf("AST node count: %d", hs.Metrics.ASTNodeCount),
						},
					},
					RecommendedAction:      "Split function into smaller single-purpose procedures.",
					ExpectedBenefit:        "Reduces mental overhead and facilitates focused unit testing.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"AST node count reduced below 300", "Existing behavior preserved"},
					DetectionMethod:        "TREE_SITTER_AST",
					VerificationStatus:     StatusProbable,
				})
			}
		}
	}

	return findings
}

