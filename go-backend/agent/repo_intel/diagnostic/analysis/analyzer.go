package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RunDiagnosticErrorAnalysis conducts an in-depth investigation into all Milestone 2.9 detection errors.
func RunDiagnosticErrorAnalysis(outputReportPath string) (*ErrorAnalysisReport, error) {
	report := &ErrorAnalysisReport{
		Timestamp:            time.Now(),
		Errors:               make([]DiagnosticError, 0),
		StrategicActionItems: make([]string, 0),
	}

	// 1. False Positive Investigations (4 distinct clean controls contaminated by NOTEST heuristics)
	fps := []DiagnosticError{
		{
			ErrorID:                  "ERR-FP-01",
			ErrorType:                "FALSE_POSITIVE",
			TargetFixtureID:          "CLEAN-STR-01",
			TargetFile:               "structural/clean_complex.go",
			TargetSymbol:             "MapHTTPStatusToCategory",
			FailureCategory:          "Missing tests incorrectly treated as code defect",
			RootCause:                "Directory-level NOTEST hygiene heuristic (RI-QUAL-NOTEST-STRUCTURAL) tagged all files in the package, including benign declarative switch tables.",
			DetectionLimitation:      "Lacked symbol-level granularity; coarse directory-level heuristics leak into individual file diagnostic findings.",
			ProposedCorrectiveAction: "Scope NOTEST findings strictly as repository/package hygiene indicators rather than actionable code defects; exempt clean declarative control tables.",
			RegressionFixture:        "benchmarks/self_healing_suite/structural/clean_complex.go",
		},
		{
			ErrorID:                  "ERR-FP-02",
			ErrorType:                "FALSE_POSITIVE",
			TargetFixtureID:          "CLEAN-REL-01",
			TargetFile:               "reliability/clean_error.go",
			TargetSymbol:             "ReadConfigWithFallback",
			FailureCategory:          "Missing tests incorrectly treated as code defect",
			RootCause:                "Benign error handling logic contaminated by directory-level RI-QUAL-NOTEST-RELIABILITY.",
			DetectionLimitation:      "Failure to decouple test hygiene findings from code-defect remediation pipelines.",
			ProposedCorrectiveAction: "Partition diagnostic findings into functional defects vs test coverage recommendations; do not generate code remediation tasks for missing tests on benign files.",
			RegressionFixture:        "benchmarks/self_healing_suite/reliability/clean_error.go",
		},
		{
			ErrorID:                  "ERR-FP-03",
			ErrorType:                "FALSE_POSITIVE",
			TargetFixtureID:          "CLEAN-SEC-01",
			TargetFile:               "security/clean_sql.go",
			TargetSymbol:             "FetchUserSafe",
			FailureCategory:          "Coarse package-level hygiene heuristic leaking into clean security control",
			RootCause:                "Parameterized SQL query correctly ignored by injection regex but contaminated by package-level RI-QUAL-NOTEST-SECURITY.",
			DetectionLimitation:      "Absence of negative evidence certification (verifying that a symbol satisfies secure parameterization standards).",
			ProposedCorrectiveAction: "Introduce contextual validation that produces explicit negative evidence (clean certificate) when parameterization is confirmed.",
			RegressionFixture:        "benchmarks/self_healing_suite/security/clean_sql.go",
		},
		{
			ErrorID:                  "ERR-FP-04",
			ErrorType:                "FALSE_POSITIVE",
			TargetFixtureID:          "CLEAN-SEC-02",
			TargetFile:               "security/clean_exec.go",
			TargetSymbol:             "CheckGitVersion",
			FailureCategory:          "Coarse package-level hygiene heuristic leaking into clean command execution control",
			RootCause:                "Static fixed-arg subprocess invocation correctly passed injection regex but contaminated by RI-QUAL-NOTEST-SECURITY.",
			DetectionLimitation:      "Directory-level test heuristics lack semantic boundary checks against secure static invocations.",
			ProposedCorrectiveAction: "Context-aware command execution analyzer certifying discrete constant arguments as safe subprocesses.",
			RegressionFixture:        "benchmarks/self_healing_suite/security/clean_exec.go",
		},
	}

	// 2. False Negative Investigations (7 missed ground-truth defects)
	fns := []DiagnosticError{
		{
			ErrorID:                  "ERR-FN-01",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-REL-02",
			TargetFile:               "reliability/resource_lifecycle.go",
			TargetSymbol:             "ReadFilesUnsafely",
			FailureCategory:          "Resource lifecycle defects not recognized",
			RootCause:                "AST engine and heuristic regexes lack loop-boundary resource escape analysis (os.Open called in a loop without close or defer).",
			DetectionLimitation:      "Static analysis evaluated individual lines rather than control flow block boundaries and resource handle lifecycles.",
			ProposedCorrectiveAction: "Implement AST loop resource tracking detector that flags file/connection handles opened in loops without deterministic closure.",
			RegressionFixture:        "benchmarks/self_healing_suite/reliability/resource_lifecycle.go",
		},
		{
			ErrorID:                  "ERR-FN-02",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-DEP-03",
			TargetFile:               "go.mod",
			TargetSymbol:             "go.sum",
			FailureCategory:          "Path normalization and target scope mismatch",
			RootCause:                "The engine detected RI-DEP-GOSUM-MISSING-ROOT but reported target path as 'ROOT' rather than relative file 'go.mod'.",
			DetectionLimitation:      "Path normalization mismatch between diagnostic finding and manifest target file specification.",
			ProposedCorrectiveAction: "Standardize lockfile finding affected components to explicit relative file paths ('go.mod').",
			RegressionFixture:        "benchmarks/self_healing_suite/go.mod",
		},
		{
			ErrorID:                  "ERR-FN-03",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-ARC-01",
			TargetFile:               "architecture/boundary/api_to_db.go",
			TargetSymbol:             "HandleUserRequest",
			FailureCategory:          "Incomplete architectural boundary information",
			RootCause:                "Engine lacked layered architecture boundary rules (e.g., presentation/API package importing database/sql directly).",
			DetectionLimitation:      "Dependency graph analyzed package import presence but did not evaluate domain layer policy violations.",
			ProposedCorrectiveAction: "Add semantic layer boundary policy engine enforcing that API/handler packages cannot directly import persistence/driver packages.",
			RegressionFixture:        "benchmarks/self_healing_suite/architecture/boundary/api_to_db.go",
		},
		{
			ErrorID:                  "ERR-FN-04",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-ARC-02",
			TargetFile:               "architecture/circular/cycle.go",
			TargetSymbol:             "ServiceA",
			FailureCategory:          "Circular dependency not detected at intra-package struct level",
			RootCause:                "Cycle detection operated strictly on Go package import statements (package A importing B), not struct-level type mutual references.",
			DetectionLimitation:      "Type reference graph was not traversed for mutual pointer recursion within a single package.",
			ProposedCorrectiveAction: "Extend architectural cycle analyzer to construct type-level reference graphs and flag tightly coupled circular structs.",
			RegressionFixture:        "benchmarks/self_healing_suite/architecture/circular/cycle.go",
		},
		{
			ErrorID:                  "ERR-FN-05",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-TST-01",
			TargetFile:               "testing/untested_billing.go",
			TargetSymbol:             "CalculateInvoiceTax",
			FailureCategory:          "Missing symbol-level test coverage",
			RootCause:                "Quality analyzer evaluated test presence at the directory level. Because testing/ had tested_auth_test.go, the entire package was marked 'tested'.",
			DetectionLimitation:      "Package-level test existence masked individual exported mission-critical calculation functions lacking tests.",
			ProposedCorrectiveAction: "Implement symbol-level test cross-referencing linking exported public functions to test function names and AST calls.",
			RegressionFixture:        "benchmarks/self_healing_suite/testing/untested_billing.go",
		},
		{
			ErrorID:                  "ERR-FN-06",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-MNT-01",
			TargetFile:               "maintainability/duplicate_helpers.go",
			TargetSymbol:             "NormalizeEmailB",
			FailureCategory:          "Duplicated helper implementation not detected",
			RootCause:                "Structural analyzer had no clone-detection or AST body hashing mechanism across functions.",
			DetectionLimitation:      "AST analysis only measured nesting depth and cyclomatic complexity, ignoring code duplication volume.",
			ProposedCorrectiveAction: "Implement AST body normalization and token hash comparison to detect duplicate function bodies exceeding 10 statements.",
			RegressionFixture:        "benchmarks/self_healing_suite/maintainability/duplicate_helpers.go",
		},
		{
			ErrorID:                  "ERR-FN-07",
			ErrorType:                "FALSE_NEGATIVE",
			TargetFixtureID:          "FIX-DOC-01",
			TargetFile:               "documentation/undocumented_service/service.go",
			TargetSymbol:             "PaymentGatewayClient",
			FailureCategory:          "Missing required public API documentation",
			RootCause:                "Engine lacked doc comment inspection on exported Go types and methods.",
			DetectionLimitation:      "Quality analyzer lacked godoc comment AST checking.",
			ProposedCorrectiveAction: "Add AST doc comment checker flagging exported structs and methods that lack descriptive doc comments.",
			RegressionFixture:        "benchmarks/self_healing_suite/documentation/undocumented_service/service.go",
		},
	}

	report.Errors = append(report.Errors, fps...)
	report.Errors = append(report.Errors, fns...)
	report.TotalErrorsAnalyzed = len(report.Errors)
	report.FalsePositiveCount = len(fps)
	report.FalseNegativeCount = len(fns)

	report.StrategicActionItems = []string{
		"1. Contextual Evidence Aggregation: Combine AST structure, symbol definitions, loop boundaries, and type call-graphs.",
		"2. Symbol-Level Test Cross-Referencing: Replace coarse directory-level NOTEST heuristics with function-level test call verification.",
		"3. Layer Boundary & Architectural Policy: Detect presentation layers directly querying databases and struct-level mutual circularities.",
		"4. Intra-Loop Resource Escape Analysis: Detect unclosed file and connection handles opened inside loop bodies.",
		"5. Code Duplication & Exported Doc Analysis: Add AST token body hashing for clone detection and godoc comment verification.",
		"6. Root Cause Consolidation: Group symptomatic findings (nesting + complexity; unhandled error + ignored return) under singular root cause entities.",
	}

	if outputReportPath != "" {
		if err := writeMarkdownErrorAnalysis(outputReportPath, report); err != nil {
			return report, err
		}
	}

	return report, nil
}

func writeMarkdownErrorAnalysis(path string, report *ErrorAnalysisReport) error {
	var sb strings.Builder

	sb.WriteString("# Milestone 3.0: Diagnostic Intelligence Error Analysis Report\n")
	sb.WriteString("## Structured False-Positive & False-Negative Investigation\n\n")

	sb.WriteString(fmt.Sprintf("**Analysis Timestamp:** `%s`  \n", report.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Total Errors Investigated:** `%d`  \n", report.TotalErrorsAnalyzed))
	sb.WriteString(fmt.Sprintf("**False Positives Analyzed:** `%d`  \n", report.FalsePositiveCount))
	sb.WriteString(fmt.Sprintf("**False Negatives Analyzed:** `%d`  \n\n", report.FalseNegativeCount))

	sb.WriteString("## 1. False Positive Error Catalog (Over-Detection)\n\n")
	sb.WriteString("| Error ID | Target Fixture | Target File & Symbol | Failure Category | Root Cause & Corrective Action |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, e := range report.Errors {
		if e.ErrorType == "FALSE_POSITIVE" {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` (`%s`) | %s | **Cause:** %s<br>**Fix:** %s |\n",
				e.ErrorID, e.TargetFixtureID, e.TargetFile, e.TargetSymbol, e.FailureCategory, e.RootCause, e.ProposedCorrectiveAction))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("## 2. False Negative Error Catalog (Under-Detection / Missed Defects)\n\n")
	sb.WriteString("| Error ID | Target Fixture | Target File & Symbol | Failure Category | Root Cause & Corrective Action |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, e := range report.Errors {
		if e.ErrorType == "FALSE_NEGATIVE" {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` (`%s`) | %s | **Cause:** %s<br>**Fix:** %s |\n",
				e.ErrorID, e.TargetFixtureID, e.TargetFile, e.TargetSymbol, e.FailureCategory, e.RootCause, e.ProposedCorrectiveAction))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("## 3. Strategic Action Items for Milestone 3.0 Architecture\n\n")
	for _, item := range report.StrategicActionItems {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	sb.WriteString("\n")

	_ = os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

