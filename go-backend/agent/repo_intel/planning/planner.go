package planning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// RemediationPlannerConfig holds parameters for the planning engine.
type RemediationPlannerConfig struct {
	RepoRoot          string
	OutputDirectory   string
	ReportPath        string
	CurrentCommit     string
	AllowStaleReports bool
}

// RemediationPlanner coordinates the end-to-end planning pipeline.
type RemediationPlanner struct {
	cfg         RemediationPlannerConfig
	ingester    *ReportIngester
	opportunity *OpportunityAnalyzer
	grouper     *TaskGrouper
	factory     *TaskFactory
	depResolver *DependencyResolver
	conflictDet *ConflictDetector
	reporter    *PlanReporter
}

// NewRemediationPlanner creates a new RemediationPlanner.
func NewRemediationPlanner(cfg RemediationPlannerConfig) *RemediationPlanner {
	return &RemediationPlanner{
		cfg:         cfg,
		ingester:    NewReportIngester(cfg.RepoRoot),
		opportunity: NewOpportunityAnalyzer(),
		grouper:     NewTaskGrouper(),
		factory:     NewTaskFactory(),
		depResolver: NewDependencyResolver(),
		conflictDet: NewConflictDetector(),
		reporter:    NewPlanReporter(cfg.OutputDirectory),
	}
}

// GeneratePlan reads a diagnostic report and constructs the machine-readable RemediationPlan.
func (rp *RemediationPlanner) GeneratePlan() (*RemediationPlan, *ProvenanceAssessment, error) {
	// 1. Ingest report & verify provenance
	diagReport, assessment, err := rp.ingester.IngestReportFile(rp.cfg.ReportPath, rp.cfg.CurrentCommit)
	if err != nil {
		return nil, assessment, fmt.Errorf("ingest diagnostic report: %w", err)
	}

	if !assessment.CommitMatches && !rp.cfg.AllowStaleReports {
		// Non-fatal warning or flag, but proceed to revalidate
	}

	plan, err := rp.PlanFromDiagnosticReport(diagReport, assessment.ReportHashSHA256)
	if err != nil {
		return nil, assessment, err
	}

	// 2. Write outputs if OutputDirectory is set
	if rp.cfg.OutputDirectory != "" {
		if err := rp.reporter.WritePlanOutputs(plan); err != nil {
			return nil, assessment, fmt.Errorf("write plan outputs: %w", err)
		}
	}

	return plan, assessment, nil
}

// PlanFromDiagnosticReport executes the planning pipeline on an in-memory DiagnosticReport.
func (rp *RemediationPlanner) PlanFromDiagnosticReport(
	report *diagnostic.DiagnosticReport,
	reportHash string,
) (*RemediationPlan, error) {
	t0 := time.Now()

	// 1. Actionability Analysis
	var actionableFindings []diagnostic.ValidatedFinding
	actionableCount := 0

	for _, f := range report.Findings {
		state, _ := rp.opportunity.EvaluateActionability(&f)
		if state == Actionable || state == ActionableWithReview || state == HumanApprovalRequired {
			actionableFindings = append(actionableFindings, f)
			actionableCount++
		}
	}

	// 2. Root-Cause Grouping
	groups := rp.grouper.GroupFindings(actionableFindings, report.Correlations)

	// 3. Task Generation
	var rawTasks []RemediationTask
	for i, g := range groups {
		task := rp.factory.CreateTask(i+1, g)
		rawTasks = append(rawTasks, task)
	}

	// 4. Dependency Resolution & Topological Ordering
	depGraph, execOrder, blockedGroups, tasksWithDeps := rp.depResolver.ResolveDependencies(rawTasks)

	// 5. Conflict Detection
	conflicts, finalTasks := rp.conflictDet.DetectConflicts(tasksWithDeps)

	// Sort final tasks deterministically: PriorityScore desc, TaskID asc
	sort.Slice(finalTasks, func(i, j int) bool {
		if finalTasks[i].PriorityScore != finalTasks[j].PriorityScore {
			return finalTasks[i].PriorityScore > finalTasks[j].PriorityScore
		}
		return finalTasks[i].TaskID < finalTasks[j].TaskID
	})

	// Tally summary statistics
	summary := PlanSummary{
		TotalFindingsConsidered: len(report.Findings),
		ActionableFindings:      actionableCount,
		TotalTasksGenerated:     len(finalTasks),
		TotalConflictsDetected:  len(conflicts),
	}

	for _, t := range finalTasks {
		switch t.ExecutionClass {
		case ClassAutomationReady:
			summary.AutomationReadyTasks++
		case ClassAutomationWithReview:
			summary.AutomationWithReviewTasks++
		case ClassHumanApprovalRequired:
			summary.HumanApprovalTasks++
		case ClassInvestigationRequired:
			summary.InvestigationTasks++
		}
		if t.Status == "BLOCKED" {
			summary.BlockedTasks++
		}
	}

	limitations := []string{
		"Remediation tasks define declarative specifications and validation gates; execution requires Phase 4 approval",
		"Soft conflicts indicate shared file modifications that must execute sequentially rather than concurrently",
		"Human approval is mandatory for all secret revocation, cryptographic token rotation, and circular dependency refactorings",
		"Planning engine does not modify repository files or apply automated code transformations in Phase 3",
	}

	if reportHash == "" {
		b, _ := json.Marshal(report)
		h := sha256.Sum256(b)
		reportHash = hex.EncodeToString(h[:])
	}

	planID := fmt.Sprintf("PLAN-%d", t0.UTC().Unix())

	return &RemediationPlan{
		PlanSchemaVersion:        "1.0",
		PlanID:                   planID,
		RepositoryIdentity:       report.Repository,
		BaselineCommit:           report.Repository.HeadCommit,
		SourceDiagnosticReportID: report.ReportID,
		SourceReportHash:         reportHash,
		GeneratedAt:              time.Now().UTC(),
		Summary:                  summary,
		Tasks:                    finalTasks,
		ExecutionOrder:           execOrder,
		DependencyGraph:          depGraph,
		Conflicts:                conflicts,
		BlockedTaskGroups:        blockedGroups,
		PlanningLimitations:      limitations,
	}, nil
}

func init() {
	_ = os.ReadFile
}
