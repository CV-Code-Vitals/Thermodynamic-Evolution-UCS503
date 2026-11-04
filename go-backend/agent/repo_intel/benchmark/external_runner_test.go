package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
	diagcontext "github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/context"
)

type ExternalRepoExecutionResult struct {
	RepoID           string                          `json:"repo_id"`
	RepoName         string                          `json:"repo_name"`
	RepoPath         string                          `json:"repo_path"`
	DiscoveredFiles  int                             `json:"discovered_files"`
	RawFindings      []repo_intel.Finding            `json:"raw_findings"`
	FilteredFindings []repo_intel.Finding            `json:"filtered_findings"`
	DiagnosticReport *diagnostic.DiagnosticReport    `json:"diagnostic_report"`
	SuppressedCount  int                             `json:"suppressed_count"`
	Suppressions     []diagcontext.SuppressionRecord `json:"suppressions"`
	DurationMs       int64                           `json:"duration_ms"`
}

func TestRunExternalRepositoriesDiagnostics(t *testing.T) {
	repos := []struct {
		ID   string
		Name string
		Path string
	}{
		{"ext-repo-01", "http_router", filepath.Join("..", "..", "..", "..", "benchmarks", "external_v1", "http_router")},
		{"ext-repo-02", "sql_service", filepath.Join("..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")},
		{"ext-repo-03", "cache_engine", filepath.Join("..", "..", "..", "..", "benchmarks", "external_v1", "cache_engine")},
		{"ext-repo-04", "cli_tooling", filepath.Join("..", "..", "..", "..", "benchmarks", "external_v1", "cli_tooling")},
		{"ext-repo-05", "queue_worker", filepath.Join("..", "..", "..", "..", "benchmarks", "external_v1", "queue_worker")},
	}

	ctx := context.Background()
	results := make([]ExternalRepoExecutionResult, 0)
	reportsDir := filepath.Join("..", "..", "..", "..", "reports")

	for _, r := range repos {
		absPath, err := filepath.Abs(r.Path)
		if err != nil {
			t.Fatalf("failed to get abs path for %s: %v", r.Name, err)
		}

		t0 := time.Now()
		discOpts := repo_intel.DiscoveryOptions{
			MaxFiles:         5000,
			MaxFileSizeBytes: 10 << 20,
			Timeout:          60 * time.Second,
		}
		disc, err := repo_intel.DiscoverRepository(ctx, absPath, discOpts)
		if err != nil {
			t.Fatalf("discovery failed for %s: %v", r.Name, err)
		}

		var filePaths []string
		for _, f := range disc.Files {
			filePaths = append(filePaths, f.RelativePath)
		}

		secAnalyzer := repo_intel.NewReliabilitySecurityAnalyzer(absPath)
		secFindings := secAnalyzer.AnalyzeReliabilityAndSecurity(disc.Files)

		qualAnalyzer := repo_intel.NewQualityAnalyzer(absPath)
		qualFindings, _ := qualAnalyzer.AnalyzeQuality(disc.Files, disc.CIWorkflows)

		manAnalyzer := repo_intel.NewManifestAnalyzer(absPath)
		_, _, manFindings := manAnalyzer.AnalyzeManifests(disc.ManifestFiles)

		var allFindings []repo_intel.Finding
		allFindings = append(allFindings, secFindings...)
		allFindings = append(allFindings, qualFindings...)
		allFindings = append(allFindings, manFindings...)

		ce := diagcontext.NewContextEngine(absPath)
		filtered, _, err := ce.FilterAndEnrich(allFindings, filePaths)
		if err != nil {
			t.Fatalf("context engine failed for %s: %v", r.Name, err)
		}

		supps := ce.GetSuppressionAudit()

		intelReport := &repo_intel.IntelligenceReport{
			Version:     "1.0.0",
			GeneratedAt: time.Now(),
			Findings:    filtered,
		}
		diagEngine := diagnostic.NewDiagnosticEngine(absPath, filepath.Join(reportsDir, "external_diag_"+r.Name))
		diagRep, err := diagEngine.ProcessIntelligenceReport(intelReport)
		if err != nil {
			t.Fatalf("diagnostic engine failed for %s: %v", r.Name, err)
		}

		dur := time.Since(t0).Milliseconds()

		res := ExternalRepoExecutionResult{
			RepoID:           r.ID,
			RepoName:         r.Name,
			RepoPath:         r.Path,
			DiscoveredFiles:  len(disc.Files),
			RawFindings:      allFindings,
			FilteredFindings: filtered,
			DiagnosticReport: diagRep,
			SuppressedCount:  len(allFindings) - len(filtered),
			Suppressions:     supps,
			DurationMs:       dur,
		}
		results = append(results, res)

		fmt.Printf("Repo %s: %d raw findings, %d active findings, %d suppressed in %d ms\n",
			r.Name, len(allFindings), len(filtered), res.SuppressedCount, dur)
		for _, f := range filtered {
			loc := ""
			if len(f.Evidence) > 0 {
				loc = fmt.Sprintf("%s:%d", f.Evidence[0].File, f.Evidence[0].LineStart)
			}
			fmt.Printf("  [ACTIVE] %s | %s | %s | %s\n", f.FindingID, loc, f.Category, f.Title)
		}
		for _, s := range supps {
			fmt.Printf("  [SUPPRESSED] %s | File: %s | Rule: %s | Reason: %s\n", s.FindingID, s.TargetFile, s.SuppressionRule, s.Rationale)
		}
	}

	// Write raw output to reports directory
	outBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	outFile := filepath.Join(reportsDir, "milestone3_3_external_predictions.json")
	if err := os.WriteFile(outFile, outBytes, 0644); err != nil {
		t.Fatalf("write file error: %v", err)
	}
	fmt.Printf("Wrote raw predictions to %s\n", outFile)
}
