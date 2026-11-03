package benchmark

import (
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
)

// CalculateRemediationMetrics calculates scientific remediation metrics based on execution results.
func CalculateRemediationMetrics(
	execReport *execution.ExecutionReport,
	eligibleDefectsCount int,
	remediatedDefectIDs map[string]bool,
	regressionsDetected int,
	integratedCount int,
	approvedCount int,
) RemediationMetrics {
	metrics := RemediationMetrics{
		EligibleGroundTruthDefects: eligibleDefectsCount,
		VerifiedRemediations:       len(remediatedDefectIDs),
		IntroducedRegressions:      regressionsDetected,
		AcceptedPatches:            execReport.Summary.Completed,
		IntegratedPatches:          integratedCount,
	}

	// 1. Verified Remediation Rate: Verified / Eligible
	if metrics.EligibleGroundTruthDefects > 0 {
		metrics.VerifiedRemediationRate = float64(metrics.VerifiedRemediations) / float64(metrics.EligibleGroundTruthDefects)
	}

	// 2. Regression Rate: Regressions / IntegratedPatches
	if metrics.IntegratedPatches > 0 {
		metrics.RegressionRate = float64(metrics.IntroducedRegressions) / float64(metrics.IntegratedPatches)
	}

	// 3. False Remediation Rate: Rolled back or invalid patches / Scheduled
	if execReport.Summary.TotalScheduled > 0 {
		metrics.FalseRemediations = execReport.Summary.RolledBack + execReport.Summary.Failed
		metrics.FalseRemediationRate = float64(metrics.FalseRemediations) / float64(execReport.Summary.TotalScheduled)
	}

	// 4. Integration Success Rate: Integrated / Approved
	if approvedCount > 0 {
		metrics.IntegrationSuccessRate = float64(metrics.IntegratedPatches) / float64(approvedCount)
	} else if metrics.AcceptedPatches > 0 {
		metrics.IntegrationSuccessRate = float64(metrics.IntegratedPatches) / float64(metrics.AcceptedPatches)
	}

	return metrics
}

// CompareBeforeAndAfter compares diagnostic reports before and after remediation.
func CompareBeforeAndAfter(before, after *diagnostic.DiagnosticReport) BeforeAfterComparison {
	comp := BeforeAfterComparison{
		BeforeBySeverity:        make(map[string]int),
		AfterBySeverity:         make(map[string]int),
		ResolvedFindingIDs:      make([]string, 0),
		UnresolvedFindingIDs:    make([]string, 0),
		NewlyIntroducedFindings: make([]string, 0),
	}

	if before != nil {
		comp.BeforeFindingsCount = len(before.Findings)
		for _, d := range before.Findings {
			sev := string(d.Severity)
			comp.BeforeBySeverity[sev]++
		}
	}

	afterIDs := make(map[string]bool)
	if after != nil {
		comp.AfterFindingsCount = len(after.Findings)
		for _, d := range after.Findings {
			sev := string(d.Severity)
			comp.AfterBySeverity[sev]++
			afterIDs[d.FindingID] = true
		}
	}

	comp.NetFindingsDelta = comp.AfterFindingsCount - comp.BeforeFindingsCount

	beforeIDs := make(map[string]bool)
	if before != nil {
		for _, d := range before.Findings {
			beforeIDs[d.FindingID] = true
			if !afterIDs[d.FindingID] {
				comp.ResolvedFindingIDs = append(comp.ResolvedFindingIDs, d.FindingID)
			} else {
				comp.UnresolvedFindingIDs = append(comp.UnresolvedFindingIDs, d.FindingID)
			}
		}
	}

	if after != nil {
		for _, d := range after.Findings {
			if !beforeIDs[d.FindingID] {
				comp.NewlyIntroducedFindings = append(comp.NewlyIntroducedFindings, d.FindingID)
			}
		}
	}

	return comp
}
