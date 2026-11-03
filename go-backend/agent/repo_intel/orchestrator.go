package repo_intel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent"
)

// OrchestratorConfig holds options for the Repository Intelligence pipeline.
type OrchestratorConfig struct {
	RepoRoot       string
	RustEnginePath string
	OutputDir      string
	CustomExcludes []string
	MaxFileSize    int64
	Timeout        time.Duration
}

// Orchestrator coordinates the end-to-end repository intelligence pipeline.
type Orchestrator struct {
	cfg OrchestratorConfig
}

// NewOrchestrator creates a new Orchestrator with the given configuration.
func NewOrchestrator(cfg OrchestratorConfig) *Orchestrator {
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 10 << 20 // 10MB default
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}
	return &Orchestrator{cfg: cfg}
}

// Run executes the complete repository intelligence analysis.
func (o *Orchestrator) Run(ctx context.Context) (*IntelligenceReport, error) {
	totalStart := time.Now()
	var metrics PerformanceMetrics
	var allFindings []Finding
	var warnings []string

	// Ensure RepoRoot is absolute and clean
	absRoot, err := filepath.Abs(o.cfg.RepoRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid repository root: %w", err)
	}
	o.cfg.RepoRoot = absRoot

	// 1. Module A: Discovery
	t0 := time.Now()
	discOpts := DiscoveryOptions{
		MaxFiles:         50000,
		MaxFileSizeBytes: o.cfg.MaxFileSize,
		Timeout:          o.cfg.Timeout,
		CustomExcludes:   o.cfg.CustomExcludes,
	}
	discResult, err := DiscoverRepository(ctx, o.cfg.RepoRoot, discOpts)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}
	metrics.DiscoveryDurationMs = time.Since(t0).Milliseconds()
	warnings = append(warnings, discResult.Warnings...)

	// Populate RepositoryMetadataRegistry from DiscoveryResult
	var sourceFilePaths []string
	for _, f := range discResult.Files {
		if f.IsSource {
			sourceFilePaths = append(sourceFilePaths, f.RelativePath)
		}
	}

	langStatsMap := make(map[string]LanguageStats)
	for lang, count := range discResult.LanguageCounts {
		langStatsMap[lang] = LanguageStats{
			Language:  lang,
			FileCount: count,
		}
	}
	for _, f := range discResult.Files {
		ext := strings.ToLower(f.Extension)
		if lang, ok := extensionLanguageMap[ext]; ok {
			st := langStatsMap[lang]
			st.TotalBytes += f.Size
			langStatsMap[lang] = st
		}
	}

	metadataRegistry := RepositoryMetadataRegistry{
		RepoRoot:          o.cfg.RepoRoot,
		GitBranch:         detectGitBranch(o.cfg.RepoRoot),
		HeadCommit:        detectHeadCommit(o.cfg.RepoRoot),
		WorkingTreeClean:  detectWorkingTreeClean(o.cfg.RepoRoot),
		Languages:         langStatsMap,
		TotalFiles:        len(discResult.Files),
		SourceFiles:       sourceFilePaths,
		IgnoredFilesCount: discResult.IgnoredCount,
		TestDirectories:   discResult.TestDirectories,
		CIWorkflows:       discResult.CIWorkflows,
	}

	// 2. Module B: Manifests
	t0 = time.Now()
	manifestAnalyzer := NewManifestAnalyzer(o.cfg.RepoRoot)
	manifestInfos, svcBoundaries, manifestFindings := manifestAnalyzer.AnalyzeManifests(discResult.ManifestFiles)
	allFindings = append(allFindings, manifestFindings...)
	metadataRegistry.Manifests = manifestInfos
	metadataRegistry.ServiceBoundaries = svcBoundaries
	metrics.ManifestDurationMs = time.Since(t0).Milliseconds()

	// 3. Module C: Structural AST Analysis
	t0 = time.Now()
	rustPath := o.resolveRustEnginePath()
	var astReport *agent.ThermodynamicReport
	structuralFilesCount := 0

	if rustPath != "" {
		structuralAnalyzer := NewStructuralAnalyzer(rustPath, o.cfg.RepoRoot)
		rep, strFindings, strErr := structuralAnalyzer.AnalyzeStructural(ctx, 30*time.Second)
		if strErr != nil {
			warnings = append(warnings, fmt.Sprintf("structural AST engine error: %v", strErr))
		} else {
			astReport = rep
			allFindings = append(allFindings, strFindings...)
			if rep != nil {
				structuralFilesCount = len(rep.FileReports)
				// Update language stats with parsed counts
				for _, fr := range rep.FileReports {
					for lang, st := range metadataRegistry.Languages {
						if strings.EqualFold(st.Language, fr.Language) {
							st.TotalLines += fr.LinesScanned
							st.ParsedCount++
							if fr.ParseErrors {
								st.ParseErrorsCount++
							}
							metadataRegistry.Languages[lang] = st
						}
					}
				}
			}
		}
	} else {
		warnings = append(warnings, "thermodynamic-ast-engine binary not found; skipping native AST parsing")
	}
	metrics.StructuralDurationMs = time.Since(t0).Milliseconds()

	// 4. Module D: Architecture & Dependency Graph
	t0 = time.Now()
	archAnalyzer := NewArchitecturalAnalyzer(o.cfg.RepoRoot)
	depGraph, archFindings := archAnalyzer.AnalyzeArchitecture(astReport, metadataRegistry.ServiceBoundaries)
	allFindings = append(allFindings, archFindings...)
	metrics.ArchitecturalDurationMs = time.Since(t0).Milliseconds()

	// 5. Module E: Reliability & Security
	t0 = time.Now()
	secAnalyzer := NewReliabilitySecurityAnalyzer(o.cfg.RepoRoot)
	secFindings := secAnalyzer.AnalyzeReliabilityAndSecurity(discResult.Files)
	allFindings = append(allFindings, secFindings...)
	metrics.SecurityDurationMs = time.Since(t0).Milliseconds()

	// 6. Module F: Quality
	t0 = time.Now()
	qualAnalyzer := NewQualityAnalyzer(o.cfg.RepoRoot)
	qualFindings, qualCoverageStr := qualAnalyzer.AnalyzeQuality(discResult.Files, discResult.CIWorkflows)
	allFindings = append(allFindings, qualFindings...)
	metrics.QualityDurationMs = time.Since(t0).Milliseconds()

	metrics.TotalDurationMs = time.Since(totalStart).Milliseconds()

	// Coverage metrics
	coverage := AnalysisCoverage{
		DiscoveryComplete:    true,
		ManifestsAnalyzed:    len(manifestInfos),
		StructuralFilesCount: structuralFilesCount,
		ArchitecturalModules: len(metadataRegistry.ServiceBoundaries),
		ReliabilityCoverage:  fmt.Sprintf("Scanned %d files for unhandled errors and SQL/exec anti-patterns", len(discResult.Files)),
		SecurityCoverage:     fmt.Sprintf("Scanned %d files for hardcoded credentials and token leakage", len(discResult.Files)),
		QualityCoverage:      qualCoverageStr,
	}

	// 7. Module G: Aggregation & Reports
	agg := NewAggregator()
	report := agg.Consolidate(metadataRegistry, depGraph, allFindings, coverage, metrics, warnings)

	// 8. Write output artifacts if OutputDir specified
	if o.cfg.OutputDir != "" {
		if err := os.MkdirAll(o.cfg.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output dir: %w", err)
		}

		// JSON report
		jsonBytes, err := agg.GenerateJSON(report)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize JSON report: %w", err)
		}
		jsonPath := filepath.Join(o.cfg.OutputDir, "repository_intelligence.json")
		if err := os.WriteFile(jsonPath, jsonBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", jsonPath, err)
		}

		// Markdown report
		mdStr := agg.GenerateMarkdown(report)
		mdPath := filepath.Join(o.cfg.OutputDir, "repository_intelligence.md")
		if err := os.WriteFile(mdPath, []byte(mdStr), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", mdPath, err)
		}

		// Dependency graph JSON
		graphBytes, err := json.MarshalIndent(depGraph, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize dependency graph: %w", err)
		}
		graphPath := filepath.Join(o.cfg.OutputDir, "repository_dependency_graph.json")
		if err := os.WriteFile(graphPath, graphBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", graphPath, err)
		}
	}

	return report, nil
}

func (o *Orchestrator) resolveRustEnginePath() string {
	if o.cfg.RustEnginePath != "" {
		if _, err := os.Stat(o.cfg.RustEnginePath); err == nil {
			return o.cfg.RustEnginePath
		}
	}

	// Probe common locations relative to repo root or cwd
	candidates := []string{
		filepath.Join(o.cfg.RepoRoot, "thermodynamic-ast-engine", "target", "release", "thermodynamic-ast-engine.exe"),
		filepath.Join(o.cfg.RepoRoot, "thermodynamic-ast-engine", "target", "release", "thermodynamic-ast-engine"),
		filepath.Join(o.cfg.RepoRoot, "..", "thermodynamic-ast-engine", "target", "release", "thermodynamic-ast-engine.exe"),
		filepath.Join(o.cfg.RepoRoot, "..", "thermodynamic-ast-engine", "target", "release", "thermodynamic-ast-engine"),
		"thermodynamic-ast-engine.exe",
		"thermodynamic-ast-engine",
	}

	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			abs, err := filepath.Abs(cand)
			if err == nil {
				return abs
			}
			return cand
		}
	}
	return ""
}

func detectGitBranch(repoRoot string) string {
	headPath := filepath.Join(repoRoot, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}
	return content
}

func detectHeadCommit(repoRoot string) string {
	headPath := filepath.Join(repoRoot, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: ") {
		refPath := filepath.Join(repoRoot, ".git", filepath.FromSlash(strings.TrimPrefix(content, "ref: ")))
		refData, err := os.ReadFile(refPath)
		if err == nil {
			return strings.TrimSpace(string(refData))
		}
	}
	return content
}

func detectWorkingTreeClean(repoRoot string) bool {
	// Simple check: if git index exists
	indexPath := filepath.Join(repoRoot, ".git", "index")
	_, err := os.Stat(indexPath)
	return err == nil
}
