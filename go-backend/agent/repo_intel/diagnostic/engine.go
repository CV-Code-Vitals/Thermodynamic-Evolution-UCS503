package diagnostic

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// DiagnosticEngine orchestrates the full Phase 2 diagnostic pipeline.
type DiagnosticEngine struct {
	repoRoot   string
	outputDir  string
	validator  *FindingValidator
	calibrator *ConfidenceCalibrator
	assessor   *SeverityAssessor
	correlator *FindingCorrelator
	classifier *ReadinessClassifier
	reporter   *DiagnosticReporter
}

// NewDiagnosticEngine initializes a new DiagnosticEngine.
func NewDiagnosticEngine(repoRoot, outputDir string) *DiagnosticEngine {
	return &DiagnosticEngine{
		repoRoot:   repoRoot,
		outputDir:  outputDir,
		validator:  NewFindingValidator(repoRoot),
		calibrator: NewConfidenceCalibrator(),
		assessor:   NewSeverityAssessor(),
		correlator: NewFindingCorrelator(),
		classifier: NewReadinessClassifier(),
		reporter:   NewDiagnosticReporter(repoRoot, outputDir),
	}
}

// ProcessReportFile reads an existing Phase 1 repository_intelligence.json report and runs the diagnostic pipeline.
func (de *DiagnosticEngine) ProcessReportFile(reportPath string) (*DiagnosticReport, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("read intelligence report: %w", err)
	}

	var intelReport repo_intel.IntelligenceReport
	if err := json.Unmarshal(data, &intelReport); err != nil {
		return nil, fmt.Errorf("unmarshal intelligence report: %w", err)
	}

	return de.ProcessIntelligenceReport(&intelReport)
}

// ProcessIntelligenceReport executes diagnostic analysis on an in-memory Phase 1 report.
func (de *DiagnosticEngine) ProcessIntelligenceReport(intelReport *repo_intel.IntelligenceReport) (*DiagnosticReport, error) {
	t0 := time.Now()

	// 1. Module A: Finding Validation
	validatedFindings, summary := de.validator.ValidateFindings(intelReport.Findings)

	// 2. Module D: Correlation and Deduplication (linking related findings)
	correlations, relatedMap := de.correlator.CorrelateFindings(validatedFindings)

	// 3. Modules B, C, E, F: Calibration, Severity, Impact, and Readiness
	for i := range validatedFindings {
		vf := &validatedFindings[i]

		// Module B: Confidence Calibration
		conf, confRationale := de.calibrator.CalibrateConfidence(vf)
		vf.CalibratedConfidence = conf
		vf.ConfidenceRationale = confRationale

		// Module C: Severity and Impact Assessment
		sev, sevRationale, impactScope, consequences := de.assessor.AssessSeverityAndImpact(vf)
		vf.Severity = sev
		vf.SeverityRationale = sevRationale
		vf.ImpactScope = impactScope
		vf.PotentialConsequences = consequences

		// Assign related findings from correlation
		if rels, ok := relatedMap[vf.FindingID]; ok {
			vf.RelatedFindingIDs = uniqueStrings(append(vf.RelatedFindingIDs, rels...))
		}

		// Module E: Root Cause Confidence
		if vf.ValidationState == StateValidated {
			vf.RootCauseConfidence = vf.CalibratedConfidence
		} else {
			vf.RootCauseConfidence = vf.CalibratedConfidence * 0.8
		}

		// Module F: Remediation Readiness
		readiness, complexity, readinessRationale := de.classifier.ClassifyRemediationReadiness(vf)
		vf.RemediationReadiness = readiness
		vf.RemediationComplexity = complexity
		vf.ReadinessRationale = readinessRationale
	}

	durationMs := time.Since(t0).Milliseconds()

	// 4. Module G: Build and Write Diagnostic Report
	report := de.reporter.BuildReport(
		intelReport.RepositoryMetadata,
		validatedFindings,
		summary,
		correlations,
		durationMs,
		intelReport.Warnings,
	)

	if de.outputDir != "" {
		if err := de.reporter.WriteOutputs(report); err != nil {
			return nil, fmt.Errorf("write diagnostic reports: %w", err)
		}
	}

	return report, nil
}
