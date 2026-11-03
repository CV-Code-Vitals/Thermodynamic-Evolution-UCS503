package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/benchmark"
)

func main() {
	repoPath := flag.String("repo", "benchmarks/self_healing_suite", "Path to benchmark target repository")
	manifestPath := flag.String("manifest", "benchmarks/ground_truth_manifest.json", "Path to ground truth manifest JSON")
	outputDir := flag.String("output", "reports", "Directory to write benchmark reports")
	approveFlag := flag.String("approve", "", "Comma-separated list of approved task IDs")
	branchName := flag.String("branch", "benchmark/milestone-2.9-integration", "Target git branch for controlled integration")
	timeoutSec := flag.Int("timeout", 300, "Stage timeout in seconds")

	flag.Parse()

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid repo path %s: %v\n", *repoPath, err)
		os.Exit(1)
	}

	absManifest, err := filepath.Abs(*manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid manifest path %s: %v\n", *manifestPath, err)
		os.Exit(1)
	}

	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid output dir %s: %v\n", *outputDir, err)
		os.Exit(1)
	}

	var approvedTasks []string
	if *approveFlag != "" {
		for _, a := range strings.Split(*approveFlag, ",") {
			trimmed := strings.TrimSpace(a)
			if trimmed != "" {
				approvedTasks = append(approvedTasks, trimmed)
			}
		}
	}

	fmt.Println("================================================================================")
	fmt.Println("Milestone 2.9: Autonomous Self-Healing Pipeline Benchmark Orchestrator")
	fmt.Printf("Benchmark Target:  %s\n", absRepo)
	fmt.Printf("Ground Truth:      %s\n", absManifest)
	fmt.Printf("Report Directory:  %s\n", absOut)
	fmt.Println("================================================================================")

	cfg := benchmark.BenchmarkRunConfig{
		BenchmarkRepoPath:       absRepo,
		GroundTruthManifestPath: absManifest,
		OutputDirectory:         absOut,
		ApprovedTaskIDs:         approvedTasks,
		CreateIntegrationBranch: true,
		IntegrationBranchName:   *branchName,
		TimeoutSec:              *timeoutSec,
	}

	pipeline := benchmark.NewPipeline(cfg)
	report, err := pipeline.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline execution error: %v\n", err)
	}

	if report != nil {
		reporter := benchmark.NewBenchmarkReporter(absOut)
		if err := reporter.WriteAllReports(report); err != nil {
			fmt.Fprintf(os.Stderr, "Failed writing reports: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------------------------------")
		fmt.Println("Benchmark Evaluation Summary:")
		fmt.Printf("  Run ID:              %s\n", report.RunID)
		fmt.Printf("  Overall Verdict:     %s\n", report.OverallVerdict)
		fmt.Printf("  True Positives (TP): %d\n", report.DetectionMetrics.TruePositives)
		fmt.Printf("  False Positives(FP): %d\n", report.DetectionMetrics.FalsePositives)
		fmt.Printf("  False Negatives(FN): %d\n", report.DetectionMetrics.FalseNegatives)
		fmt.Printf("  True Negatives (TN): %d\n", report.DetectionMetrics.TrueNegatives)
		fmt.Printf("  Precision:           %.2f%%\n", report.DetectionMetrics.Precision*100)
		fmt.Printf("  Recall:              %.2f%%\n", report.DetectionMetrics.Recall*100)
		fmt.Printf("  F1 Score:            %.2f%%\n", report.DetectionMetrics.F1Score*100)
		fmt.Printf("  Verified Remediations: %d\n", report.RemediationMetrics.VerifiedRemediations)
		fmt.Printf("  Remediation Rate:    %.2f%%\n", report.RemediationMetrics.VerifiedRemediationRate*100)
		fmt.Printf("  Regressions:         %d (Rate: %.2f%%)\n", report.RemediationMetrics.IntroducedRegressions, report.RemediationMetrics.RegressionRate*100)
		fmt.Printf("  Duration:            %d ms\n", report.TotalDurationMs)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("[+] Scientific reports and artifacts written to: %s/\n", absOut)
	}

	if err != nil {
		os.Exit(1)
	}
}

