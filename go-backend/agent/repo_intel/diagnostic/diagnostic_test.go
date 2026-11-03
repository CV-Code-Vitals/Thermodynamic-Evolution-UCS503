package diagnostic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

func TestValidator_PathTraversalDefense(t *testing.T) {
	tmpDir := t.TempDir()
	v := NewFindingValidator(tmpDir)

	f := repo_intel.Finding{
		FindingID: "TRAVERSAL-TEST",
		Category:  repo_intel.CategorySecurity,
		Severity:  repo_intel.SeverityHigh,
		Title:     "Path Traversal Test",
		Evidence: []repo_intel.EvidenceLocation{
			{File: "../../etc/passwd", LineStart: 1, LineEnd: 1},
		},
	}

	vf, summary := v.ValidateFindings([]repo_intel.Finding{f})
	if len(vf) != 1 {
		t.Fatalf("expected 1 validated finding, got %d", len(vf))
	}

	if vf[0].ValidationState != StateInvalid {
		t.Errorf("expected StateInvalid for path traversal attempt, got %s", vf[0].ValidationState)
	}
	if summary.Invalid != 1 {
		t.Errorf("expected 1 invalid in summary, got %d", summary.Invalid)
	}
	if !strings.Contains(vf[0].ValidationReason, "resolves outside repository root") {
		t.Errorf("expected path traversal validation reason, got: %s", vf[0].ValidationReason)
	}
}

func TestValidator_StaleLineBoundDetection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "short.go")
	_ = os.WriteFile(filePath, []byte("package short\nfunc A() {}\n"), 0644) // 2 lines

	v := NewFindingValidator(tmpDir)

	f := repo_intel.Finding{
		FindingID: "STALE-LINE-TEST",
		Category:  repo_intel.CategoryStructural,
		Severity:  repo_intel.SeverityMedium,
		Title:     "Stale line number test",
		Evidence: []repo_intel.EvidenceLocation{
			{File: "short.go", LineStart: 100, LineEnd: 100}, // 100 > 2
		},
	}

	vf, summary := v.ValidateFindings([]repo_intel.Finding{f})
	if len(vf) != 1 {
		t.Fatalf("expected 1 validated finding, got %d", len(vf))
	}

	if vf[0].ValidationState != StateStale {
		t.Errorf("expected StateStale for out of bounds line, got %s", vf[0].ValidationState)
	}
	if summary.Stale != 1 {
		t.Errorf("expected 1 stale in summary, got %d", summary.Stale)
	}
}

func TestCalibrator_IndependenceFromSeverity(t *testing.T) {
	c := NewConfidenceCalibrator()

	// High severity but heuristic detection -> moderate confidence
	fHeuristic := &ValidatedFinding{
		FindingID:       "HEURISTIC-HIGH",
		Category:        repo_intel.CategorySecurity,
		Severity:        repo_intel.SeverityHigh,
		DetectionMethod: "STATIC_TAINT_HEURISTIC",
		ValidationState: StateRequiresHumanReview,
	}
	confH, _ := c.CalibrateConfidence(fHeuristic)

	// Low severity but deterministic Tree-sitter AST -> high confidence
	fDeterministic := &ValidatedFinding{
		FindingID:       "AST-LOW",
		Category:        repo_intel.CategoryStructural,
		Severity:        repo_intel.SeverityLow,
		DetectionMethod: "TREE_SITTER_AST",
		ValidationState: StateValidated,
	}
	confD, _ := c.CalibrateConfidence(fDeterministic)

	if confH >= confD {
		t.Errorf("expected deterministic AST confidence (%.2f) to exceed heuristic confidence (%.2f)", confD, confH)
	}
}

func TestSeverity_ImpactAssessment(t *testing.T) {
	sa := NewSeverityAssessor()

	// Secret credential -> CRITICAL
	fSecret := &ValidatedFinding{
		FindingID: "RI-SEC-HARDCODED-SECRET",
		Category:  repo_intel.CategorySecurity,
	}
	sevSecret, _, _, _ := sa.AssessSeverityAndImpact(fSecret)
	if sevSecret != repo_intel.SeverityCritical {
		t.Errorf("expected Critical severity for secret finding, got %s", sevSecret)
	}

	// Unpinned Docker base -> LOW
	fDocker := &ValidatedFinding{
		FindingID: "RI-SEC-DOCKER-UNPINNED-BASE-L1",
		Category:  repo_intel.CategorySecurity,
	}
	sevDocker, _, _, _ := sa.AssessSeverityAndImpact(fDocker)
	if sevDocker != repo_intel.SeverityLow {
		t.Errorf("expected Low severity for unpinned base image, got %s", sevDocker)
	}
}

func TestCorrelation_ClusteringAndDuplicates(t *testing.T) {
	fc := NewFindingCorrelator()

	findings := []ValidatedFinding{
		{
			FindingID:          "FINDING-DUP-1",
			Category:           repo_intel.CategoryStructural,
			AffectedComponents: []string{"backend/srv.go"},
			Evidence: []repo_intel.EvidenceLocation{
				{File: "backend/srv.go", LineStart: 10},
			},
		},
		{
			FindingID:          "FINDING-DUP-2",
			Category:           repo_intel.CategoryStructural,
			AffectedComponents: []string{"backend/srv.go"},
			Evidence: []repo_intel.EvidenceLocation{
				{File: "backend/srv.go", LineStart: 10}, // exact duplicate location & category
			},
		},
	}

	corrs, relatedMap := fc.CorrelateFindings(findings)
	if len(corrs) == 0 {
		t.Fatalf("expected at least 1 correlation for duplicate findings")
	}

	if corrs[0].Type != CorrelationDuplicate {
		t.Errorf("expected CorrelationDuplicate, got %s", corrs[0].Type)
	}

	if len(relatedMap["FINDING-DUP-1"]) == 0 {
		t.Errorf("expected related map to link FINDING-DUP-1 to FINDING-DUP-2")
	}
}

func TestReadiness_Classification(t *testing.T) {
	rc := NewReadinessClassifier()

	// Docker root USER -> AUTOMATION_READY
	fDocker := &ValidatedFinding{
		FindingID:            "RI-SEC-DOCKER-ROOT-USER",
		Category:             repo_intel.CategorySecurity,
		ValidationState:      StateValidated,
		CalibratedConfidence: 0.90,
	}
	readDocker, compDocker, _ := rc.ClassifyRemediationReadiness(fDocker)
	if readDocker != ReadinessAutomationReady {
		t.Errorf("expected ReadinessAutomationReady for Docker root user, got %s", readDocker)
	}
	if compDocker != "LOW" {
		t.Errorf("expected LOW complexity, got %s", compDocker)
	}

	// Hardcoded secret -> HUMAN_APPROVAL_REQUIRED
	fSecret := &ValidatedFinding{
		FindingID:            "RI-SEC-HARDCODED-SECRET",
		Category:             repo_intel.CategorySecurity,
		ValidationState:      StateRequiresHumanReview,
		CalibratedConfidence: 0.95,
	}
	readSecret, compSecret, _ := rc.ClassifyRemediationReadiness(fSecret)
	if readSecret != ReadinessHumanApprovalRequired {
		t.Errorf("expected ReadinessHumanApprovalRequired for secret, got %s", readSecret)
	}
	if compSecret != "HIGH" {
		t.Errorf("expected HIGH complexity, got %s", compSecret)
	}
}

func TestDiagnosticEngine_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "reports")

	// Create test file
	_ = os.WriteFile(filepath.Join(tmpDir, "app.go"), []byte("package main\nfunc Run() {}\n"), 0644)

	intelReport := &repo_intel.IntelligenceReport{
		RepositoryMetadata: repo_intel.RepositoryMetadataRegistry{
			RepoRoot: tmpDir,
		},
		Findings: []repo_intel.Finding{
			{
				FindingID:       "RI-DOCKER-ROOT-USER",
				Category:        repo_intel.CategorySecurity,
				Severity:        repo_intel.SeverityMedium,
				Confidence:      0.85,
				Title:           "Dockerfile lacks USER",
				DetectionMethod: "MANIFEST_ANALYSIS",
				Evidence: []repo_intel.EvidenceLocation{
					{File: "app.go", LineStart: 1, LineEnd: 1},
				},
			},
		},
	}

	engine := NewDiagnosticEngine(tmpDir, outDir)
	diagReport, err := engine.ProcessIntelligenceReport(intelReport)
	if err != nil {
		t.Fatalf("engine run failed: %v", err)
	}

	if diagReport.ReportSchemaVersion != "1.0" {
		t.Errorf("expected schema version 1.0, got %s", diagReport.ReportSchemaVersion)
	}
	if len(diagReport.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(diagReport.Findings))
	}

	// Verify outputs written to disk
	pJson := filepath.Join(outDir, "diagnostic_report.json")
	pVal := filepath.Join(outDir, "finding_validation.json")
	pMd := filepath.Join(outDir, "diagnostic_report.md")

	if _, err := os.Stat(pJson); err != nil {
		t.Errorf("missing %s", pJson)
	}
	if _, err := os.Stat(pVal); err != nil {
		t.Errorf("missing %s", pVal)
	}
	if _, err := os.Stat(pMd); err != nil {
		t.Errorf("missing %s", pMd)
	}

	// Unmarshal JSON
	data, _ := os.ReadFile(pJson)
	var unmarshaled DiagnosticReport
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal generated json: %v", err)
	}
	if unmarshaled.Summary.TotalFindings != 1 {
		t.Errorf("expected summary total 1, got %d", unmarshaled.Summary.TotalFindings)
	}
}

func TestControlledQualityEvaluationHarness(t *testing.T) {
	tmpDir := t.TempDir()
	fixtureDir := filepath.Join(tmpDir, "fixtures")
	reportPath := filepath.Join(tmpDir, "eval_report.md")

	metrics, mdContent, err := RunControlledQualityEvaluation(fixtureDir, reportPath)
	if err != nil {
		t.Fatalf("RunControlledQualityEvaluation failed: %v", err)
	}

	if metrics.TotalFixtures < 5 {
		t.Errorf("expected at least 5 fixtures, got %d", metrics.TotalFixtures)
	}
	if metrics.Precision < 0.70 {
		t.Errorf("expected precision >= 0.70, got %.2f", metrics.Precision)
	}
	if metrics.EvidenceAccuracyPercent < 80.0 {
		t.Errorf("expected evidence accuracy >= 80%%, got %.1f%%", metrics.EvidenceAccuracyPercent)
	}
	if mdContent == "" {
		t.Errorf("expected non-empty evaluation report content")
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Errorf("expected report file to be written to %s", reportPath)
	}
}
