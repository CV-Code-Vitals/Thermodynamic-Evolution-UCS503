package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

func main() {
	repoPath := flag.String("repo", ".", "Path to repository root")
	intelReportPath := flag.String("input", "reports/repository_intelligence.json", "Path to input repository_intelligence.json")
	outputDir := flag.String("output", "reports", "Directory to write output diagnostic reports")
	flag.Parse()

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Invalid repo path %q: %v", *repoPath, err)
	}

	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Invalid output dir %q: %v", *outputDir, err)
	}

	absInput, err := filepath.Abs(*intelReportPath)
	if err != nil {
		log.Fatalf("Invalid input path %q: %v", *intelReportPath, err)
	}

	log.Printf("Starting Diagnostic Intelligence Engine (Phase 2)")
	log.Printf("Repository root: %s", absRepo)
	log.Printf("Input report:    %s", absInput)
	log.Printf("Output dir:      %s", absOut)

	engine := diagnostic.NewDiagnosticEngine(absRepo, absOut)
	report, err := engine.ProcessReportFile(absInput)
	if err != nil {
		log.Fatalf("Diagnostic processing failed: %v", err)
	}

	// Run controlled quality evaluation
	fixtureDir := filepath.Join(os.TempDir(), "diagnostic_quality_fixtures")
	evalReportPath := filepath.Join(absOut, "diagnostic_quality_evaluation.md")
	metrics, _, err := diagnostic.RunControlledQualityEvaluation(fixtureDir, evalReportPath)
	if err != nil {
		log.Fatalf("Controlled quality evaluation failed: %v", err)
	}

	fmt.Println("\n=======================================================")
	fmt.Printf(" Milestone 2.8 Phase 2 Diagnostic Intelligence Complete\n")
	fmt.Printf(" Total Ingested Findings: %d\n", report.Summary.TotalFindings)
	fmt.Printf(" - Validated:             %d\n", report.Summary.Validated)
	fmt.Printf(" - Partially Validated:   %d\n", report.Summary.PartiallyValidated)
	fmt.Printf(" - Stale:                 %d\n", report.Summary.Stale)
	fmt.Printf(" - Invalid:               %d\n", report.Summary.Invalid)
	fmt.Printf(" - Requires Review:       %d\n", report.Summary.RequiresHumanReview)
	fmt.Printf(" Total Correlations:      %d\n", len(report.Correlations))
	fmt.Printf(" Fixture Evaluation:\n")
	fmt.Printf(" - Precision:             %.2f%%\n", metrics.Precision*100)
	fmt.Printf(" - Recall:                %.2f%%\n", metrics.Recall*100)
	fmt.Printf(" - Evidence Accuracy:     %.2f%%\n", metrics.EvidenceAccuracyPercent)
	fmt.Printf(" Reports Written to:      %s\n", absOut)
	fmt.Println("=======================================================")

	os.Exit(0)
}

