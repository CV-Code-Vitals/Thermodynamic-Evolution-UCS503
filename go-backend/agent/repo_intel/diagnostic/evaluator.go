package diagnostic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// FixtureTestCase defines a controlled test case with known ground truth.
type FixtureTestCase struct {
	Name                    string
	Description             string
	Finding                 repo_intel.Finding
	ExpectedValidationState ValidationState
	ExpectedReadiness       RemediationReadiness
	ExpectedMinConfidence   float64
	ExpectedMaxConfidence   float64
	ExpectedSeverity        repo_intel.Severity
	IsGroundTruthDefect     bool
}

// EvaluationMetrics summarizes quality evaluation results across fixtures.
type EvaluationMetrics struct {
	TotalFixtures           int
	ValidationStateMatches  int
	ReadinessMatches        int
	ConfidenceMatches       int
	SeverityMatches         int
	TruePositives           int
	FalsePositives          int
	TrueNegatives           int
	FalseNegatives          int
	Precision               float64
	Recall                  float64
	F1Score                 float64
	EvidenceAccuracyPercent float64
	DeduplicationAccuracy   float64
}

// RunControlledQualityEvaluation executes the diagnostic engine on controlled fixtures and emits reports/diagnostic_quality_evaluation.md.
func RunControlledQualityEvaluation(fixtureDir string, outputReportPath string) (*EvaluationMetrics, string, error) {
	if err := os.MkdirAll(fixtureDir, 0755); err != nil {
		return nil, "", fmt.Errorf("create fixture dir: %w", err)
	}

	// 1. Create controlled source files
	goSource := `package fixture

import (
	"fmt"
	"os/exec"
)

// ConfirmedGenuineCmdInject concatenates input directly into shell
func ConfirmedGenuineCmdInject(input string) {
	cmd := exec.Command("sh", fmt.Sprintf("-c %s", input))
	_ = cmd.Run()
}

// HarmlessExecCommand passes arguments safely as individual tokens
func HarmlessExecCommand(safeArg string) {
	cmd := exec.Command("ls", "-l", safeArg)
	_ = cmd.Run()
}

// HighComplexityMonolithicFunc contains deep nesting and branching
func HighComplexityMonolithicFunc(a, b, c, d, e int) int {
	if a > 0 {
		if b > 0 {
			if c > 0 {
				if d > 0 {
					return e * 2
				}
			}
		}
	}
	return 0
}
`
	if err := os.WriteFile(filepath.Join(fixtureDir, "sample.go"), []byte(goSource), 0644); err != nil {
		return nil, "", err
	}

	shortFile := "package fixture\nfunc Short() {}\n"
	if err := os.WriteFile(filepath.Join(fixtureDir, "short.go"), []byte(shortFile), 0644); err != nil {
		return nil, "", err
	}

	testFile := "package fixture\nimport \"testing\"\nfunc TestShort(t *testing.T) {}\n"
	if err := os.WriteFile(filepath.Join(fixtureDir, "short_test.go"), []byte(testFile), 0644); err != nil {
		return nil, "", err
	}

	// Untested module
	untestedDir := filepath.Join(fixtureDir, "untested")
	if err := os.MkdirAll(untestedDir, 0755); err != nil {
		return nil, "", err
	}
	_ = os.WriteFile(filepath.Join(untestedDir, "mod1.go"), []byte("package untested\nfunc M1() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(untestedDir, "mod2.go"), []byte("package untested\nfunc M2() {}\n"), 0644)

	cases := []FixtureTestCase{
		{
			Name:        "1. Confirmed Architectural Cycle",
			Description: "Circular import between packages A and B",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-ARCH-CYCLE-PKG-A-PKG-B",
				Category:        repo_intel.CategoryArchitecture,
				Severity:        repo_intel.SeverityHigh,
				Confidence:      0.98,
				Title:           "Circular dependency detected between pkg A and pkg B",
				DetectionMethod: "GRAPH_CYCLE_DETECTION",
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 3, LineEnd: 6, Description: "Circular import path"},
				},
				RootCause: "Direct cyclic reference between packages",
			},
			ExpectedValidationState: StateValidated,
			ExpectedReadiness:       ReadinessHumanApprovalRequired,
			ExpectedMinConfidence:   0.90,
			ExpectedMaxConfidence:   1.00,
			ExpectedSeverity:        repo_intel.SeverityHigh,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "2. Confirmed High-Complexity Monolithic Function",
			Description: "High cyclomatic complexity with deep nesting",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-STR-HIGH-COMPLEXITY-SAMPLE-L21",
				Category:        repo_intel.CategoryStructural,
				Severity:        repo_intel.SeverityMedium,
				Confidence:      0.95,
				Title:           "Excessive cyclomatic complexity (M=22)",
				DetectionMethod: "TREE_SITTER_AST",
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 21, LineEnd: 33, Description: "Branching logic"},
				},
				RootCause: "Accumulation of branching logic in a single function",
			},
			ExpectedValidationState: StateValidated,
			ExpectedReadiness:       ReadinessAutomationWithReview,
			ExpectedMinConfidence:   0.90,
			ExpectedMaxConfidence:   1.00,
			ExpectedSeverity:        repo_intel.SeverityMedium,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "3. Genuine Command Injection Risk",
			Description: "Direct string concatenation into exec.Command argument",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-SEC-CMD-INJECT-SAMPLE-L10",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityHigh,
				Confidence:      0.88,
				Title:           "Potential command injection risk in exec.Command",
				DetectionMethod: "STATIC_TAINT_HEURISTIC",
				VerificationStatus: repo_intel.StatusProbable,
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 10, LineEnd: 10, Description: `cmd := exec.Command("sh", fmt.Sprintf("-c %s", input))`},
				},
				RootCause: "String concatenation into process execution arguments.",
			},
			ExpectedValidationState: StateRequiresHumanReview,
			ExpectedReadiness:       ReadinessInvestigationRequired,
			ExpectedMinConfidence:   0.40,
			ExpectedMaxConfidence:   0.60,
			ExpectedSeverity:        repo_intel.SeverityHigh,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "4. Confirmed Hardcoded Secret Credential",
			Description: "Hardcoded AWS token with automatic redaction",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-SEC-HARDCODED-SECRET-AWS-KEY",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityCritical,
				Confidence:      0.95,
				Title:           "Potential hardcoded secret (AWS Access Key ID)",
				DetectionMethod: "REGEX_PATTERN_MATCHING",
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 5, LineEnd: 5, Description: "AWS pattern matched: AK****LE"},
				},
				RootCause: "Embedding secret tokens in source code",
			},
			ExpectedValidationState: StateRequiresHumanReview,
			ExpectedReadiness:       ReadinessHumanApprovalRequired,
			ExpectedMinConfidence:   0.50,
			ExpectedMaxConfidence:   0.70,
			ExpectedSeverity:        repo_intel.SeverityCritical,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "5. Confirmed Excessive Nesting Depth",
			Description: "Control flow nesting >= 4 in sample function",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-STR-HIGH-NESTING-SAMPLE-L21",
				Category:        repo_intel.CategoryStructural,
				Severity:        repo_intel.SeverityMedium,
				Confidence:      0.95,
				Title:           "Excessive control flow nesting depth (4)",
				DetectionMethod: "TREE_SITTER_AST",
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 21, LineEnd: 27, Description: "Nesting depth 4"},
				},
				RootCause: "Deeply nested if blocks without guard clauses",
			},
			ExpectedValidationState: StateValidated,
			ExpectedReadiness:       ReadinessAutomationWithReview,
			ExpectedMinConfidence:   0.90,
			ExpectedMaxConfidence:   1.00,
			ExpectedSeverity:        repo_intel.SeverityMedium,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "6. Confirmed Missing Test Coverage",
			Description: "Production directory without co-located unit tests",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-QUAL-NOTEST-UNTESTED",
				Category:        repo_intel.CategoryTestQuality,
				Severity:        repo_intel.SeverityLow,
				Confidence:      0.85,
				Title:           "Directory untested lacks unit tests",
				DetectionMethod: "FILE_CORRELATION_HEURISTIC",
				AffectedFiles:   []string{"untested/mod1.go", "untested/mod2.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "untested/mod1.go", LineStart: 1, LineEnd: 1, Description: "Source file without unit tests"},
				},
				RootCause: "Modules authored without corresponding unit test suites.",
			},
			ExpectedValidationState: StateRequiresHumanReview,
			ExpectedReadiness:       ReadinessInvestigationRequired,
			ExpectedMinConfidence:   0.50,
			ExpectedMaxConfidence:   0.70,
			ExpectedSeverity:        repo_intel.SeverityLow,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "7. Stale Finding with Out-of-Bounds Line Number",
			Description: "Finding pointing to non-existent line 9999 in short.go",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-STR-STALE-LINE-SHORT",
				Category:        repo_intel.CategoryStructural,
				Severity:        repo_intel.SeverityLow,
				Confidence:      0.90,
				Title:           "Stale line number in short.go",
				DetectionMethod: "TREE_SITTER_AST",
				AffectedFiles:   []string{"short.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "short.go", LineStart: 9999, LineEnd: 9999, Description: "Stale line"},
				},
			},
			ExpectedValidationState: StateStale,
			ExpectedReadiness:       ReadinessInformationalOnly,
			ExpectedMinConfidence:   0.00,
			ExpectedMaxConfidence:   0.20,
			ExpectedSeverity:        repo_intel.SeverityLow,
			IsGroundTruthDefect:     false,
		},
		{
			Name:        "8. Stale Finding on Non-Existent Deleted File",
			Description: "Finding targeting file that was deleted",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-STR-DELETED-FILE",
				Category:        repo_intel.CategoryStructural,
				Severity:        repo_intel.SeverityLow,
				Confidence:      0.90,
				Title:           "Finding on non-existent file",
				DetectionMethod: "TREE_SITTER_AST",
				AffectedFiles:   []string{"deleted_module.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "deleted_module.go", LineStart: 5, LineEnd: 5, Description: "Deleted"},
				},
			},
			ExpectedValidationState: StateStale,
			ExpectedReadiness:       ReadinessInformationalOnly,
			ExpectedMinConfidence:   0.00,
			ExpectedMaxConfidence:   0.20,
			ExpectedSeverity:        repo_intel.SeverityLow,
			IsGroundTruthDefect:     false,
		},
		{
			Name:        "9. Malformed / Invalid Finding Metadata",
			Description: "Finding with empty category and missing title",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-INVALID-METADATA",
				Category:        "", // empty
				Severity:        repo_intel.SeverityLow,
				Title:           "", // empty
				DetectionMethod: "UNKNOWN",
			},
			ExpectedValidationState: StateInvalid,
			ExpectedReadiness:       ReadinessInformationalOnly,
			ExpectedMinConfidence:   0.00,
			ExpectedMaxConfidence:   0.20,
			ExpectedSeverity:        repo_intel.SeverityLow,
			IsGroundTruthDefect:     false,
		},
		{
			Name:        "10. Path Traversal Attempt Escaping Root",
			Description: "Finding attempting path traversal escape via ../../",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-SEC-TRAVERSAL-ATTEMPT",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityLow,
				Title:           "Traversal path",
				DetectionMethod: "STATIC_TAINT_HEURISTIC",
				Evidence: []repo_intel.EvidenceLocation{
					{File: "../../etc/shadow", LineStart: 1, LineEnd: 1, Description: "Escaped path"},
				},
			},
			ExpectedValidationState: StateInvalid,
			ExpectedReadiness:       ReadinessInformationalOnly,
			ExpectedMinConfidence:   0.00,
			ExpectedMaxConfidence:   0.20,
			ExpectedSeverity:        repo_intel.SeverityLow,
			IsGroundTruthDefect:     false,
		},
		{
			Name:        "11. Primary Finding for Deduplication Check",
			Description: "First scanner report for command injection at line 10",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-SEC-CMD-INJECT-DUP-PRIMARY-L10",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityHigh,
				Confidence:      0.88,
				Title:           "Potential command injection (Scanner A)",
				DetectionMethod: "STATIC_TAINT_HEURISTIC",
				VerificationStatus: repo_intel.StatusProbable,
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 10, LineEnd: 10, Description: "Line 10"},
				},
			},
			ExpectedValidationState: StateRequiresHumanReview,
			ExpectedReadiness:       ReadinessInvestigationRequired,
			ExpectedMinConfidence:   0.40,
			ExpectedMaxConfidence:   0.60,
			ExpectedSeverity:        repo_intel.SeverityHigh,
			IsGroundTruthDefect:     true,
		},
		{
			Name:        "12. Duplicate Finding for Deduplication Check",
			Description: "Second scanner report targeting identical location and category",
			Finding: repo_intel.Finding{
				FindingID:       "FIXTURE-RI-SEC-CMD-INJECT-DUP-SECONDARY-L10",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityHigh,
				Confidence:      0.88,
				Title:           "Potential command injection (Scanner B)",
				DetectionMethod: "STATIC_TAINT_HEURISTIC",
				VerificationStatus: repo_intel.StatusProbable,
				AffectedFiles:   []string{"sample.go"},
				Evidence: []repo_intel.EvidenceLocation{
					{File: "sample.go", LineStart: 10, LineEnd: 10, Description: "Line 10"},
				},
			},
			ExpectedValidationState: StateRequiresHumanReview,
			ExpectedReadiness:       ReadinessInvestigationRequired,
			ExpectedMinConfidence:   0.40,
			ExpectedMaxConfidence:   0.60,
			ExpectedSeverity:        repo_intel.SeverityHigh,
			IsGroundTruthDefect:     true,
		},
	}

	engine := NewDiagnosticEngine(fixtureDir, "")

	var rawFindings []repo_intel.Finding
	for _, tc := range cases {
		rawFindings = append(rawFindings, tc.Finding)
	}

	intelReport := &repo_intel.IntelligenceReport{
		RepositoryMetadata: repo_intel.RepositoryMetadataRegistry{
			RepoRoot: fixtureDir,
		},
		Findings: rawFindings,
	}

	diagReport, err := engine.ProcessIntelligenceReport(intelReport)
	if err != nil {
		return nil, "", fmt.Errorf("process fixtures: %w", err)
	}

	metrics := EvaluationMetrics{
		TotalFixtures: len(cases),
	}

	for i, tc := range cases {
		vf := diagReport.Findings[i]

		valMatch := vf.ValidationState == tc.ExpectedValidationState
		readMatch := vf.RemediationReadiness == tc.ExpectedReadiness
		confMatch := vf.CalibratedConfidence >= tc.ExpectedMinConfidence && vf.CalibratedConfidence <= tc.ExpectedMaxConfidence
		sevMatch := vf.Severity == tc.ExpectedSeverity

		if valMatch {
			metrics.ValidationStateMatches++
		}
		if readMatch {
			metrics.ReadinessMatches++
		}
		if confMatch {
			metrics.ConfidenceMatches++
		}
		if sevMatch {
			metrics.SeverityMatches++
		}

		// Ground truth matrix:
		// A finding is considered a predicted active defect if it is VALIDATED or REQUIRES_HUMAN_REVIEW,
		// has calibrated confidence >= 0.40, and is not classified as INFORMATIONAL_ONLY.
		isPredictedDefect := (vf.ValidationState == StateValidated || vf.ValidationState == StateRequiresHumanReview) &&
			vf.CalibratedConfidence >= 0.40 &&
			vf.RemediationReadiness != ReadinessInformationalOnly

		if tc.IsGroundTruthDefect {
			if isPredictedDefect {
				metrics.TruePositives++
			} else {
				metrics.FalseNegatives++
			}
		} else {
			if isPredictedDefect {
				metrics.FalsePositives++
			} else {
				metrics.TrueNegatives++
			}
		}
	}

	if metrics.TruePositives+metrics.FalsePositives > 0 {
		metrics.Precision = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalsePositives)
	}
	if metrics.TruePositives+metrics.FalseNegatives > 0 {
		metrics.Recall = float64(metrics.TruePositives) / float64(metrics.TruePositives+metrics.FalseNegatives)
	}
	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
	}
	metrics.EvidenceAccuracyPercent = float64(metrics.ValidationStateMatches) / float64(metrics.TotalFixtures) * 100.0

	// Check correlation / deduplication
	for _, c := range diagReport.Correlations {
		if c.Type == CorrelationDuplicate {
			metrics.DeduplicationAccuracy = 1.0
			break
		}
	}

	// Generate diagnostic_quality_evaluation.md
	var sb strings.Builder
	sb.WriteString("# Diagnostic Quality Evaluation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Evaluation Date:** %s  \n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Evaluated Fixtures:** %d controlled test cases with known ground truth  \n\n", metrics.TotalFixtures))

	sb.WriteString("## Evaluation Metrics Summary\n\n")
	sb.WriteString("| Metric | Value | Target Threshold | Status |\n|---|---|---|---|\n")
	sb.WriteString(fmt.Sprintf("| **Precision** | %.2f%% | >= 80.0%% | %s |\n", metrics.Precision*100, passFail(metrics.Precision >= 0.80)))
	sb.WriteString(fmt.Sprintf("| **Recall** | %.2f%% | >= 80.0%% | %s |\n", metrics.Recall*100, passFail(metrics.Recall >= 0.80)))
	sb.WriteString(fmt.Sprintf("| **F1-Score** | %.2f | >= 0.80 | %s |\n", metrics.F1Score, passFail(metrics.F1Score >= 0.80)))
	sb.WriteString(fmt.Sprintf("| **Evidence Accuracy** | %.1f%% | >= 90.0%% | %s |\n", metrics.EvidenceAccuracyPercent, passFail(metrics.EvidenceAccuracyPercent >= 90.0)))
	sb.WriteString(fmt.Sprintf("| **Deduplication Accuracy** | %.1f%% | 100.0%% | %s |\n\n", metrics.DeduplicationAccuracy*100, passFail(metrics.DeduplicationAccuracy == 1.0)))

	sb.WriteString("## Fixture Test Results Matrix\n\n")
	sb.WriteString("| Fixture Name | Expected Validation | Actual Validation | Expected Readiness | Actual Readiness | Confidence | Severity | Result |\n|---|---|---|---|---|---|---|---|\n")
	for i, tc := range cases {
		vf := diagReport.Findings[i]
		passed := vf.ValidationState == tc.ExpectedValidationState &&
			vf.RemediationReadiness == tc.ExpectedReadiness &&
			vf.Severity == tc.ExpectedSeverity
		statusStr := "PASS"
		if !passed {
			statusStr = "FAIL"
		}
		sb.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | `%s` | `%s` | `%.2f` | `%s` | **%s** |\n",
			tc.Name, tc.ExpectedValidationState, vf.ValidationState, tc.ExpectedReadiness, vf.RemediationReadiness, vf.CalibratedConfidence, vf.Severity, statusStr))
	}
	sb.WriteString("\n")

	sb.WriteString("## Deduplication & Correlation Telemetry\n\n")
	sb.WriteString(fmt.Sprintf("- **Total correlations generated:** %d\n", len(diagReport.Correlations)))
	for _, c := range diagReport.Correlations {
		sb.WriteString(fmt.Sprintf("- **[%s] %s**: Primary finding `%s` linked with %s (*%s*)\n",
			c.Type, c.CorrelationID, c.PrimaryFindingID, strings.Join(c.RelatedFindingIDs, ", "), c.Rationale))
	}
	sb.WriteString("\n")

	mdContent := sb.String()
	if outputReportPath != "" {
		_ = os.WriteFile(outputReportPath, []byte(mdContent), 0644)
	}

	return &metrics, mdContent, nil
}

func passFail(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
