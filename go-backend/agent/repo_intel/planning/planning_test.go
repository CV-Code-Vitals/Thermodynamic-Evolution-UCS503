package planning

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

func TestIngester_PathTraversalRejection(t *testing.T) {
	ingester := NewReportIngester("/repo")
	report := &diagnostic.DiagnosticReport{
		ReportSchemaVersion: "1.0",
		ReportID:            "DR-TRAVERSAL",
		Repository: repo_intel.RepositoryMetadataRegistry{
			RepoRoot: "/repo",
		},
		Findings: []diagnostic.ValidatedFinding{
			{
				FindingID: "TRAVERSAL-FINDING",
				Evidence: []repo_intel.EvidenceLocation{
					{File: "../../etc/shadow"},
				},
			},
		},
	}

	_, err := ingester.ValidateReport(report, "hash", "")
	if err == nil {
		t.Fatalf("expected path traversal error, got nil")
	}
}

func TestOpportunity_ActionabilityFilter(t *testing.T) {
	oa := NewOpportunityAnalyzer()

	// Invalid -> RejectedAsInvalid
	fInvalid := &diagnostic.ValidatedFinding{
		ValidationState: diagnostic.StateInvalid,
	}
	state, _ := oa.EvaluateActionability(fInvalid)
	if state != RejectedAsInvalid {
		t.Errorf("expected RejectedAsInvalid, got %s", state)
	}

	// Secret -> HumanApprovalRequired
	fSecret := &diagnostic.ValidatedFinding{
		FindingID:            "RI-SEC-HARDCODED-SECRET",
		Category:             repo_intel.CategorySecurity,
		RemediationReadiness: diagnostic.ReadinessHumanApprovalRequired,
	}
	state, _ = oa.EvaluateActionability(fSecret)
	if state != HumanApprovalRequired {
		t.Errorf("expected HumanApprovalRequired, got %s", state)
	}

	// Docker root -> Actionable
	fDocker := &diagnostic.ValidatedFinding{
		FindingID:            "RI-SEC-DOCKER-ROOT-USER",
		ValidationState:      diagnostic.StateValidated,
		RemediationReadiness: diagnostic.ReadinessAutomationReady,
		CalibratedConfidence: 0.90,
	}
	state, _ = oa.EvaluateActionability(fDocker)
	if state != Actionable {
		t.Errorf("expected Actionable, got %s", state)
	}
}

func TestTaskGrouper_CorrelationConsolidation(t *testing.T) {
	tg := NewTaskGrouper()

	findings := []diagnostic.ValidatedFinding{
		{
			FindingID:          "RI-STR-COMPLEXITY",
			AffectedComponents: []string{"srv.go"},
		},
		{
			FindingID:          "RI-STR-NESTING",
			AffectedComponents: []string{"srv.go"},
		},
	}

	correlations := []diagnostic.FindingCorrelation{
		{
			CorrelationID:     "CORR-01",
			Type:              diagnostic.CorrelationParentChild,
			PrimaryFindingID:  "RI-STR-COMPLEXITY",
			RelatedFindingIDs: []string{"RI-STR-NESTING"},
			Rationale:         "Compound complexity",
		},
	}

	groups := tg.GroupFindings(findings, correlations)
	if len(groups) != 1 {
		t.Fatalf("expected 1 consolidated group, got %d", len(groups))
	}
	if len(groups[0].RelatedFindings) != 1 {
		t.Errorf("expected 1 related finding in group, got %d", len(groups[0].RelatedFindings))
	}
}

func TestConflictDetector_HardAndSoftConflicts(t *testing.T) {
	cd := NewConflictDetector()

	tasks := []RemediationTask{
		{
			TaskID:        "RP-0001",
			TargetFiles:   []string{"app.go"},
			TargetSymbols: []string{"HandleRequest"},
		},
		{
			TaskID:        "RP-0002",
			TargetFiles:   []string{"app.go"},
			TargetSymbols: []string{"HandleRequest"}, // Same symbol -> HARD CONFLICT
		},
		{
			TaskID:        "RP-0003",
			TargetFiles:   []string{"app.go"},
			TargetSymbols: []string{"HelperFunc"}, // Different symbol, same file -> SOFT CONFLICT
		},
	}

	conflicts, updated := cd.DetectConflicts(tasks)
	if len(conflicts) < 2 {
		t.Fatalf("expected at least 2 conflicts, got %d", len(conflicts))
	}

	foundHard := false
	foundSoft := false
	for _, c := range conflicts {
		if c.Type == ConflictHard {
			foundHard = true
		}
		if c.Type == ConflictSoft {
			foundSoft = true
		}
	}

	if !foundHard {
		t.Errorf("expected hard conflict for same symbol")
	}
	if !foundSoft {
		t.Errorf("expected soft conflict for same file")
	}
	_ = updated
}

func TestStrategySelector_ConservativeAssignment(t *testing.T) {
	ss := NewStrategySelector()

	// Hardcoded secret -> StrategyHumanGuided & ClassHumanApprovalRequired
	fSecret := &diagnostic.ValidatedFinding{
		FindingID: "RI-SEC-HARDCODED-SECRET",
		Category:  repo_intel.CategorySecurity,
	}
	strat, class, approval, _ := ss.SelectStrategy(fSecret)
	if strat != StrategyHumanGuided || class != ClassHumanApprovalRequired || !approval {
		t.Errorf("expected HumanGuided / ApprovalRequired for secret, got %s / %s / %v", strat, class, approval)
	}

	// Docker unpinned base -> StrategyRuleBased & ClassAutomationReady
	fDocker := &diagnostic.ValidatedFinding{
		FindingID:            "RI-SEC-DOCKER-UNPINNED-BASE",
		Category:             repo_intel.CategorySecurity,
		CalibratedConfidence: 0.90,
	}
	strat, class, approval, _ = ss.SelectStrategy(fDocker)
	if strat != StrategyRuleBased || class != ClassAutomationReady || approval {
		t.Errorf("expected RuleBased / AutomationReady for docker unpinned, got %s / %s / %v", strat, class, approval)
	}
}

func TestRemediationPlanner_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "reports")

	report := &diagnostic.DiagnosticReport{
		ReportSchemaVersion: "1.0",
		ReportID:            "DR-TEST-E2E",
		Repository: repo_intel.RepositoryMetadataRegistry{
			RepoRoot:   tmpDir,
			HeadCommit: "commit_abc123",
		},
		Findings: []diagnostic.ValidatedFinding{
			{
				FindingID:            "RI-SEC-DOCKER-UNPINNED-BASE-L1",
				Category:             repo_intel.CategorySecurity,
				Severity:             repo_intel.SeverityLow,
				ValidationState:      diagnostic.StateValidated,
				RemediationReadiness: diagnostic.ReadinessAutomationReady,
				CalibratedConfidence: 0.90,
				Title:                "Unpinned Docker base image",
				AffectedComponents:   []string{"Dockerfile"},
			},
		},
	}

	planner := NewRemediationPlanner(RemediationPlannerConfig{
		RepoRoot:        tmpDir,
		OutputDirectory: outDir,
		CurrentCommit:   "commit_abc123",
	})

	plan, err := planner.PlanFromDiagnosticReport(report, "test_hash")
	if err != nil {
		t.Fatalf("PlanFromDiagnosticReport failed: %v", err)
	}

	if plan.PlanSchemaVersion != "1.0" {
		t.Errorf("expected plan schema version 1.0, got %s", plan.PlanSchemaVersion)
	}
	if len(plan.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(plan.Tasks))
	}
	if plan.Tasks[0].TaskID != "RP-0001" {
		t.Errorf("expected TaskID RP-0001, got %s", plan.Tasks[0].TaskID)
	}

	// Test writing outputs
	err = planner.reporter.WritePlanOutputs(plan)
	if err != nil {
		t.Fatalf("WritePlanOutputs failed: %v", err)
	}

	pPlan := filepath.Join(outDir, "remediation_plan.json")
	pMd := filepath.Join(outDir, "remediation_plan.md")
	pDep := filepath.Join(outDir, "remediation_dependency_graph.json")
	pConf := filepath.Join(outDir, "remediation_conflict_report.json")

	for _, p := range []string{pPlan, pMd, pDep, pConf} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing expected output %s", p)
		}
	}

	// Check JSON unmarshaling
	data, _ := os.ReadFile(pPlan)
	var unmarshaled RemediationPlan
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal remediation_plan.json: %v", err)
	}
	if unmarshaled.Summary.TotalTasksGenerated != 1 {
		t.Errorf("expected 1 generated task, got %d", unmarshaled.Summary.TotalTasksGenerated)
	}
}

func TestPlanningEvaluationHarness(t *testing.T) {
	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "eval_report.md")

	metrics, mdContent, err := RunControlledPlanningEvaluation(reportPath)
	if err != nil {
		t.Fatalf("RunControlledPlanningEvaluation failed: %v", err)
	}

	if metrics.TotalScenarios != 15 {
		t.Errorf("expected 15 scenarios, got %d", metrics.TotalScenarios)
	}
	if metrics.SuccessRatePercent != 100.0 {
		t.Errorf("expected 100.0%% success rate, got %.2f%%", metrics.SuccessRatePercent)
	}
	if mdContent == "" {
		t.Errorf("expected non-empty evaluation report content")
	}
	if _, err := os.Stat(reportPath); err != nil {
		t.Errorf("expected output report at %s", reportPath)
	}
}
