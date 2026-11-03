package repo_intel

import (
	"fmt"
	"path/filepath"
	"strings"
)

// QualityAnalyzer inspects repository test coverage presence, CI coverage,
// and engineering hygiene configurations.
type QualityAnalyzer struct {
	repoRoot string
}

// NewQualityAnalyzer creates a new QualityAnalyzer.
func NewQualityAnalyzer(repoRoot string) *QualityAnalyzer {
	return &QualityAnalyzer{repoRoot: repoRoot}
}

// AnalyzeQuality evaluates test presence, CI workflows, and hygiene configs.
func (qa *QualityAnalyzer) AnalyzeQuality(files []DiscoveredFile, ciWorkflows []string) ([]Finding, string) {
	var findings []Finding

	// 1. Analyze Test Presence per directory / module
	dirSourceFiles := make(map[string][]string)
	dirTestFiles := make(map[string][]string)
	hasGo := false
	hasRust := false

	for _, file := range files {
		ext := strings.ToLower(file.Extension)
		dir := filepath.Dir(file.RelativePath)

		if ext == ".go" {
			hasGo = true
		}
		if ext == ".rs" {
			hasRust = true
		}

		isTest := isTestFilePath(file.RelativePath, ext)
		if isTest {
			dirTestFiles[dir] = append(dirTestFiles[dir], file.RelativePath)
		} else if file.IsSource {
			dirSourceFiles[dir] = append(dirSourceFiles[dir], file.RelativePath)
		}
	}

	// Flag directories with substantive source code but zero test files
	for dir, srcFiles := range dirSourceFiles {
		// Ignore top-level root, scripts, docs, assets, build, deploy, reports, security_audit
		normDir := filepath.ToSlash(dir)
		if normDir == "." || strings.HasPrefix(normDir, "docs") || strings.HasPrefix(normDir, "scripts") ||
			strings.HasPrefix(normDir, "build") || strings.HasPrefix(normDir, "deploy") ||
			strings.HasPrefix(normDir, "reports") || strings.HasPrefix(normDir, "security_audit") {
			continue
		}

		tests := dirTestFiles[dir]
		if len(tests) == 0 && len(srcFiles) >= 2 {
			var evidenceLocs []EvidenceLocation
			for _, sf := range srcFiles {
				evidenceLocs = append(evidenceLocs, EvidenceLocation{
					File:        sf,
					LineStart:   1,
					LineEnd:     1,
					Description: "Source file without co-located unit tests",
				})
			}

			findingID := fmt.Sprintf("RI-QUAL-NOTEST-%s", sanitizeID(normDir))
			findings = append(findings, Finding{
				FindingID:   findingID,
				Category:    CategoryTestQuality,
				Severity:    SeverityLow,
				Confidence:  0.85,
				Title:       fmt.Sprintf("Directory %s lacks unit tests for %d source files", normDir, len(srcFiles)),
				Description: fmt.Sprintf("Directory contains %d production source files (%s) but no co-located unit or integration test files.", len(srcFiles), strings.Join(evidenceLocsBasenames(srcFiles), ", ")),
				RootCause:   "Modules authored without corresponding unit test suites.",
				AffectedFiles: srcFiles,
				Evidence:      evidenceLocs,
				RecommendedAction: fmt.Sprintf("Add unit tests for modules in %s to prevent regressions during automated evolution.", normDir),
				ExpectedBenefit:   "Provides safety gates and regression catchers during autonomous refactoring.",
				RegressionRisk:    SeverityLow,
				ValidationRequirements: []string{"Unit tests authored and passing"},
				DetectionMethod:   "FILE_CORRELATION_HEURISTIC",
				VerificationStatus: StatusConfirmed,
			})
		}
	}

	// 2. Check CI Workflow completeness for detected primary languages
	if len(ciWorkflows) == 0 {
		findings = append(findings, Finding{
			FindingID:   "RI-QUAL-NOCI-WORKFLOWS",
			Category:    CategoryTestQuality,
			Severity:    SeverityHigh,
			Confidence:  0.95,
			Title:       "No continuous integration (CI) workflows detected",
			Description: "The repository contains no CI workflows in .github/workflows, leaving code changes unverified before merge.",
			RootCause:   "Absence of continuous integration test runners.",
			AffectedFiles: []string{".github/workflows"},
			Evidence: []EvidenceLocation{
				{
					File:        ".github/workflows",
					Description: "Directory empty or missing",
				},
			},
			RecommendedAction: "Create automated CI workflows in .github/workflows for test execution and linting.",
			ExpectedBenefit:   "Continuous gate verification on every commit and pull request.",
			RegressionRisk:    SeverityLow,
			ValidationRequirements: []string{"Workflow runs on push/pull_request and executes test suites"},
			DetectionMethod:   "CI_WORKFLOW_AUDIT",
			VerificationStatus: StatusConfirmed,
		})
	} else {
		var workflowNames []string
		for _, cw := range ciWorkflows {
			workflowNames = append(workflowNames, strings.ToLower(filepath.Base(cw)))
		}
		allWF := strings.Join(workflowNames, " ")

		if hasRust && !strings.Contains(allWF, "rust") && !strings.Contains(allWF, "cargo") {
			findings = append(findings, Finding{
				FindingID:   "RI-QUAL-CI-MISSING-RUST",
				Category:    CategoryTestQuality,
				Severity:    SeverityMedium,
				Confidence:  0.80,
				Title:       "Rust codebase detected without dedicated Rust CI workflow",
				Description: "Repository contains Rust source files but no CI workflow explicitly dedicated to cargo test/check.",
				RootCause:   "CI coverage gap for Rust toolchain.",
				AffectedFiles: ciWorkflows,
				Evidence: []EvidenceLocation{
					{
						File:        ciWorkflows[0],
						Description: "Existing CI workflows omit cargo test/check steps",
					},
				},
				RecommendedAction: "Ensure Rust compilation and cargo test are included in CI workflow steps.",
				ExpectedBenefit:   "Validates that native Rust engine passes all unit tests.",
				RegressionRisk:    SeverityLow,
				ValidationRequirements: []string{"cargo test executed in CI"},
				DetectionMethod:   "CI_COVERAGE_GAP",
				VerificationStatus: StatusProbable,
			})
		}

		if hasGo && !strings.Contains(allWF, "go") {
			findings = append(findings, Finding{
				FindingID:   "RI-QUAL-CI-MISSING-GO",
				Category:    CategoryTestQuality,
				Severity:    SeverityMedium,
				Confidence:  0.80,
				Title:       "Go codebase detected without dedicated Go CI workflow",
				Description: "Repository contains Go source files but no CI workflow explicitly named for Go test/check.",
				RootCause:   "CI coverage gap for Go toolchain.",
				AffectedFiles: ciWorkflows,
				Evidence: []EvidenceLocation{
					{
						File:        ciWorkflows[0],
						Description: "Existing CI workflows omit Go test steps",
					},
				},
				RecommendedAction: "Ensure Go test suite is executed in continuous integration.",
				ExpectedBenefit:   "Validates Go backend behavior in CI.",
				RegressionRisk:    SeverityLow,
				ValidationRequirements: []string{"go test ./... executed in CI"},
				DetectionMethod:   "CI_COVERAGE_GAP",
				VerificationStatus: StatusProbable,
			})
		}
	}

	// 3. Repository Documentation & Hygiene
	hasReadme := false
	hasLicense := false
	for _, f := range files {
		base := strings.ToUpper(filepath.Base(f.RelativePath))
		if strings.HasPrefix(base, "README") {
			hasReadme = true
		}
		if strings.HasPrefix(base, "LICENSE") || strings.HasPrefix(base, "COPYING") {
			hasLicense = true
		}
	}

	if !hasReadme {
		findings = append(findings, Finding{
			FindingID:   "RI-HYG-MISSING-README",
			Category:    CategoryReliability,
			Severity:    SeverityLow,
			Confidence:  1.0,
			Title:       "Missing README documentation",
			Description: "No root or top-level README file was detected, hindering developer onboarding and automated repository context ingestion.",
			RootCause:   "Documentation omitted.",
			AffectedFiles: []string{"README.md"},
			RecommendedAction: "Create a README.md documenting architecture, build steps, and prerequisites.",
			ExpectedBenefit:   "Provides clear onboarding and automated documentation ingestion context.",
			RegressionRisk:    SeverityLow,
			ValidationRequirements: []string{"README.md exists and is populated"},
			DetectionMethod:   "DOCUMENTATION_AUDIT",
			VerificationStatus: StatusConfirmed,
		})
	}

	if !hasLicense {
		findings = append(findings, Finding{
			FindingID:   "RI-HYG-MISSING-LICENSE",
			Category:    CategoryReliability,
			Severity:    SeverityLow,
			Confidence:  1.0,
			Title:       "Missing repository LICENSE file",
			Description: "No software license file (LICENSE or COPYING) was detected in repository root.",
			RootCause:   "License file omitted.",
			AffectedFiles: []string{"LICENSE"},
			RecommendedAction: "Add an explicit LICENSE file (e.g. Apache-2.0 or MIT) to govern code redistribution and usage.",
			ExpectedBenefit:   "Clarifies legal rights and license boundaries.",
			RegressionRisk:    SeverityLow,
			ValidationRequirements: []string{"LICENSE file exists"},
			DetectionMethod:   "LICENSE_AUDIT",
			VerificationStatus: StatusConfirmed,
		})
	}

	coverageStr := fmt.Sprintf("Scanned %d files across %d directories for test co-location and %d CI workflows", len(files), len(dirSourceFiles), len(ciWorkflows))
	return findings, coverageStr
}

func isTestFilePath(relPath, ext string) bool {
	base := strings.ToLower(filepath.Base(relPath))
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	if strings.HasSuffix(base, ".test.ts") || strings.HasSuffix(base, ".test.js") ||
		strings.HasSuffix(base, ".spec.ts") || strings.HasSuffix(base, ".spec.js") {
		return true
	}
	if strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") {
		return true
	}
	norm := filepath.ToSlash(relPath)
	if strings.Contains(norm, "/tests/") || strings.Contains(norm, "/test/") {
		return true
	}
	return false
}

func evidenceLocsBasenames(paths []string) []string {
	var names []string
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	return names
}

