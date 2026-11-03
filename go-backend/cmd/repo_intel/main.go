package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

func main() {
	repoPath := flag.String("repo", ".", "Path to repository to analyze")
	outputDir := flag.String("output", "reports", "Directory to write output reports")
	enginePath := flag.String("engine", "", "Path to thermodynamic-ast-engine binary (optional)")
	timeout := flag.Duration("timeout", 5*time.Minute, "Execution timeout for repository intelligence")
	flag.Parse()

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatalf("Invalid repo path %q: %v", *repoPath, err)
	}

	absOut, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Invalid output dir %q: %v", *outputDir, err)
	}

	log.Printf("Starting Repository Intelligence Agent on %s", absRepo)
	log.Printf("Output directory: %s", absOut)

	cfg := repo_intel.OrchestratorConfig{
		RepoRoot:       absRepo,
		RustEnginePath: *enginePath,
		OutputDir:      absOut,
		Timeout:        *timeout,
	}

	orchestrator := repo_intel.NewOrchestrator(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	report, err := orchestrator.Run(ctx)
	if err != nil {
		log.Fatalf("Repository intelligence analysis failed: %v", err)
	}

	fmt.Println("\n=======================================================")
	fmt.Printf(" Repository Intelligence Analysis Complete\n")
	fmt.Printf(" Total Files Scanned: %d\n", report.RepositoryMetadata.TotalFiles)
	fmt.Printf(" Total Findings:      %d\n", len(report.Findings))
	fmt.Printf(" Analysis Runtime:    %d ms\n", report.PerformanceMetrics.TotalDurationMs)
	fmt.Printf(" Reports Written to:  %s\n", absOut)
	fmt.Println("=======================================================")

	os.Exit(0)
}
