package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

func main() {
	repoPath := flag.String("repo", ".", "Path to repository root")
	diagReportPath := flag.String("input", "reports/diagnostic_report.json", "Path to input diagnostic_report.json")
	outputDir := flag.String("output", "reports", "Directory to write output remediation plan")
	currentCommit := flag.String("commit", "", "Current repository HEAD commit")
	flag.Parse()

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Invalid repo path %q: %v", *repoPath, err)
	}

	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Invalid output dir %q: %v", *outputDir, err)
	}

	absInput, err := filepath.Abs(*diagReportPath)
	if err != nil {
		log.Fatalf("Invalid input path %q: %v", *diagReportPath, err)
	}

	log.Printf("Starting Remediation Planning Engine (Phase 3)")
	log.Printf("Repository root: %s", absRepo)
	log.Printf("Input report:    %s", absInput)
	log.Printf("Output dir:      %s", absOut)

	cfg := planning.RemediationPlannerConfig{
		RepoRoot:        absRepo,
		OutputDirectory: absOut,
		ReportPath:      absInput,
		CurrentCommit:   *currentCommit,
	}

	planner := planning.NewRemediationPlanner(cfg)
	plan, assessment, err := planner.GeneratePlan()
	if err != nil {
		log.Fatalf("Planning failed: %v", err)
	}

	// Run controlled quality evaluation
	evalReportPath := filepath.Join(absOut, "remediation_planning_evaluation.md")
	metrics, _, err := planning.RunControlledPlanningEvaluation(evalReportPath)
	if err != nil {
		log.Fatalf("Controlled planning evaluation failed: %v", err)
	}

	fmt.Println("\n=======================================================")
	fmt.Printf(" Milestone 2.8 Phase 3 Remediation Planning Complete\n")
	fmt.Printf(" Ingested Report ID:      %s\n", assessment.ReportID)
	fmt.Printf(" Provenance Assessment:   %s\n", assessment.AssessmentRationale)
	fmt.Printf(" Total Findings Evaluated: %d\n", plan.Summary.TotalFindingsConsidered)
	fmt.Printf(" Actionable Findings:     %d\n", plan.Summary.ActionableFindings)
	fmt.Printf(" Generated Tasks:         %d\n", plan.Summary.TotalTasksGenerated)
	fmt.Printf(" - Automation Ready:      %d\n", plan.Summary.AutomationReadyTasks)
	fmt.Printf(" - Automation With Review: %d\n", plan.Summary.AutomationWithReviewTasks)
	fmt.Printf(" - Human Approval:        %d\n", plan.Summary.HumanApprovalTasks)
	fmt.Printf(" - Investigation Only:    %d\n", plan.Summary.InvestigationTasks)
	fmt.Printf(" - Blocked:               %d\n", plan.Summary.BlockedTasks)
	fmt.Printf(" Total Conflicts:         %d\n", plan.Summary.TotalConflictsDetected)
	fmt.Printf(" Fixture Evaluation:\n")
	fmt.Printf(" - Scenarios Passed:      %d / %d (%.2f%%)\n", metrics.ScenariosPassed, metrics.TotalScenarios, metrics.SuccessRatePercent)
	fmt.Printf(" Reports Written to:      %s\n", absOut)
	fmt.Println("=======================================================")

	os.Exit(0)
}

