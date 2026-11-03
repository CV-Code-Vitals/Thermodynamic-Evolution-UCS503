package repo_intel

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent"
)

// setupMockRepository creates an isolated temporary directory simulating a multi-ecosystem repository.
func setupMockRepository(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// 1. Create .gitignore
	gitignore := "target/\nnode_modules/\n*.secret.txt\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	// 2. Create ignored folder and file
	if err := os.MkdirAll(filepath.Join(dir, "target"), 0755); err != nil {
		t.Fatalf("failed to mkdir target: %v", err)
	}
	_ = os.WriteFile(filepath.Join(dir, "target", "ignored.o"), []byte("binary data"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "test.secret.txt"), []byte("super_secret"), 0644)

	// 3. Create go.mod
	goMod := "module example.com/testmod\n\ngo 1.22\n\nrequire (\n\tgithub.com/gin-gonic/gin v1.9.1\n)\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	// Missing go.sum intentionally to trigger finding

	// 4. Create package.json without package-lock.json
	pkgJSON := "{\n  \"name\": \"frontend\",\n  \"dependencies\": {\n    \"react\": \"^18.2.0\"\n  }\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// 5. Create Dockerfile with unpinned base and missing USER
	dockerfile := "FROM golang:latest\nWORKDIR /app\nCOPY . .\nRUN go build -o app .\nCMD [\"./app\"]\n"
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		t.Fatalf("failed to write Dockerfile: %v", err)
	}

	// 6. Create Go source file with hardcoded secret and command injection
	backendDir := filepath.Join(dir, "backend")
	if err := os.MkdirAll(backendDir, 0755); err != nil {
		t.Fatalf("failed to mkdir backend: %v", err)
	}
	badGo := `package backend

import (
	"fmt"
	"os/exec"
)

func BadHandler(input string) {
	aws_access_key_id := "AKIAIOSFODNN7EXAMPLE"
	cmd := exec.Command("sh", fmt.Sprintf("-c %s", input))
	_ = cmd.Run()
	_ = aws_access_key_id
}
`
	if err := os.WriteFile(filepath.Join(backendDir, "handler.go"), []byte(badGo), 0644); err != nil {
		t.Fatalf("failed to write handler.go: %v", err)
	}
	// Note: backend folder has no _test.go file, so it should trigger RI-QUAL-NOTEST

	// 7. Create CI workflow in .github/workflows
	wfDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatalf("failed to mkdir workflows: %v", err)
	}
	ciWf := "name: CI\non: [push]\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo CI\n"
	if err := os.WriteFile(filepath.Join(wfDir, "ci-go.yml"), []byte(ciWf), 0644); err != nil {
		t.Fatalf("failed to write ci-go.yml: %v", err)
	}

	return dir
}

func TestDiscovery_GitIgnoreAndFilter(t *testing.T) {
	repoRoot := setupMockRepository(t)
	ctx := context.Background()

	opts := DefaultDiscoveryOptions()
	res, err := DiscoverRepository(ctx, repoRoot, opts)
	if err != nil {
		t.Fatalf("DiscoverRepository failed: %v", err)
	}

	if res.RepoRoot != repoRoot {
		t.Errorf("expected RepoRoot %s, got %s", repoRoot, res.RepoRoot)
	}

	foundIgnored := false
	foundSecretTxt := false
	foundHandler := false
	foundGoMod := false

	for _, f := range res.Files {
		if strings.Contains(f.RelativePath, "target") {
			foundIgnored = true
		}
		if strings.HasSuffix(f.RelativePath, ".secret.txt") {
			foundSecretTxt = true
		}
		if f.RelativePath == "backend/handler.go" || f.RelativePath == filepath.Join("backend", "handler.go") {
			foundHandler = true
		}
		if f.RelativePath == "go.mod" {
			foundGoMod = true
		}
	}

	if foundIgnored {
		t.Errorf("discovery included target/ directory which was in .gitignore")
	}
	if foundSecretTxt {
		t.Errorf("discovery included *.secret.txt which was in .gitignore")
	}
	if !foundHandler {
		t.Errorf("discovery failed to find backend/handler.go")
	}
	if !foundGoMod {
		t.Errorf("discovery failed to find go.mod")
	}
	if res.LanguageCounts["Go"] < 1 {
		t.Errorf("expected at least 1 Go file, got %d", res.LanguageCounts["Go"])
	}
}

func TestManifestAnalyzer(t *testing.T) {
	repoRoot := setupMockRepository(t)
	manifestAnalyzer := NewManifestAnalyzer(repoRoot)

	manifests := []string{"go.mod", "package.json", "Dockerfile"}
	infos, svcs, findings := manifestAnalyzer.AnalyzeManifests(manifests)

	if len(infos) != 3 {
		t.Fatalf("expected 3 manifest infos, got %d", len(infos))
	}
	_ = svcs

	foundMissingGoSum := false
	foundMissingPkgLock := false
	foundUnpinnedBase := false
	foundMissingUser := false

	for _, f := range findings {
		switch f.FindingID {
		case "RI-DEP-GOSUM-MISSING-ROOT":
			foundMissingGoSum = true
		case "RI-DEP-LOCKFILE-MISSING-ROOT":
			foundMissingPkgLock = true
		case "RI-SEC-DOCKER-UNPINNED-BASE-DOCKERFILE-L1":
			foundUnpinnedBase = true
		case "RI-SEC-DOCKER-ROOT-USER-DOCKERFILE":
			foundMissingUser = true
		}
	}

	if !foundMissingGoSum {
		t.Errorf("expected RI-DEP-GOSUM-MISSING-ROOT finding")
	}
	if !foundMissingPkgLock {
		t.Errorf("expected RI-DEP-LOCKFILE-MISSING-ROOT finding")
	}
	if !foundUnpinnedBase {
		t.Errorf("expected RI-SEC-DOCKER-UNPINNED-BASE-DOCKERFILE-L1 finding")
	}
	if !foundMissingUser {
		t.Errorf("expected RI-SEC-DOCKER-ROOT-USER-DOCKERFILE finding")
	}
}

func TestReliabilitySecurityAnalyzer_RedactionAndDetection(t *testing.T) {
	repoRoot := setupMockRepository(t)
	secAnalyzer := NewReliabilitySecurityAnalyzer(repoRoot)

	files := []DiscoveredFile{
		{
			RelativePath: filepath.Join("backend", "handler.go"),
			Extension:    ".go",
			IsSource:     true,
		},
	}

	findings := secAnalyzer.AnalyzeReliabilityAndSecurity(files)
	if len(findings) == 0 {
		t.Fatalf("expected security findings, got 0")
	}

	foundSecret := false
	foundCmdInject := false

	for _, f := range findings {
		if strings.Contains(f.FindingID, "SECRET") {
			foundSecret = true
			if len(f.Evidence) > 0 {
				desc := f.Evidence[0].Description
				// Check redaction: must NOT contain the full secret "AKIAIOSFODNN7EXAMPLE"
				if strings.Contains(desc, "AKIAIOSFODNN7EXAMPLE") {
					t.Errorf("secret was NOT redacted in evidence: %s", desc)
				}
				if !strings.Contains(desc, "AK****LE") {
					t.Errorf("expected redacted format AK****LE, got: %s", desc)
				}
			}
		}
		if strings.Contains(f.FindingID, "CMD-INJECT") {
			foundCmdInject = true
		}
	}

	if !foundSecret {
		t.Errorf("expected AWS secret finding")
	}
	if !foundCmdInject {
		t.Errorf("expected command injection finding")
	}
}

func TestArchitecturalCycleDetection(t *testing.T) {
	arch := NewArchitecturalAnalyzer("/fake/root")

	// Create a synthetic report with circular dependency A -> B -> C -> A
	mockReport := &agent.ThermodynamicReport{
		FileReports: []agent.FileReport{
			{FilePath: "pkg/a.go"},
			{FilePath: "pkg/b.go"},
			{FilePath: "pkg/c.go"},
			{FilePath: "pkg/d.go"},
		},
		Relationships: []agent.Relationship{
			{Kind: "import", SourceFile: "pkg/a.go", TargetFile: "pkg/b.go"},
			{Kind: "import", SourceFile: "pkg/b.go", TargetFile: "pkg/c.go"},
			{Kind: "import", SourceFile: "pkg/c.go", TargetFile: "pkg/a.go"}, // cycle!
			{Kind: "import", SourceFile: "pkg/d.go", TargetFile: "pkg/a.go"}, // safe
		},
	}

	depGraph, findings := arch.AnalyzeArchitecture(mockReport, nil)

	if len(depGraph.Nodes) != 4 {
		t.Fatalf("expected 4 nodes in dependency graph, got %d", len(depGraph.Nodes))
	}
	if len(depGraph.Edges) != 4 {
		t.Fatalf("expected 4 edges in dependency graph, got %d", len(depGraph.Edges))
	}

	foundCycleFinding := false
	for _, f := range findings {
		if strings.HasPrefix(f.FindingID, "RI-ARCH-CYCLE") {
			foundCycleFinding = true
			if f.Severity != SeverityHigh {
				t.Errorf("expected SeverityHigh for cycle finding, got %s", f.Severity)
			}
		}
	}

	if !foundCycleFinding {
		t.Errorf("expected cycle detection finding for a -> b -> c -> a")
	}
}

func TestStructuralAnalyzer_HotspotConversion(t *testing.T) {
	sa := NewStructuralAnalyzer("dummy-binary", "/fake/root")

	// Mock report with high nesting and complexity
	mockReport := &agent.ThermodynamicReport{
		FileReports: []agent.FileReport{
			{
				FilePath:     "backend/complex.go",
				Language:     "Go",
				LinesScanned: 150,
				ParseErrors:  true,
				Hotspots: []agent.Hotspot{
					{
						FunctionName: "DeeplyNestedFunc",
						LineNumber:   42,
						Metrics: agent.FunctionMetrics{
							CyclomaticComplexity: 16,
							MaxNestingDepth:      6,
							ASTNodeCount:         450,
						},
					},
				},
			},
		},
	}

	findings := sa.convertReportToFindings(mockReport)

	foundParseErr := false
	foundNesting := false
	foundComplexity := false
	foundVolume := false

	for _, f := range findings {
		switch {
		case strings.Contains(f.FindingID, "PARSE-ERROR"):
			foundParseErr = true
		case strings.Contains(f.FindingID, "HIGH-NESTING"):
			foundNesting = true
		case strings.Contains(f.FindingID, "HIGH-COMPLEXITY"):
			foundComplexity = true
		case strings.Contains(f.FindingID, "LARGE-AST-VOLUME"):
			foundVolume = true
		}
	}

	if !foundParseErr {
		t.Errorf("expected parse error finding")
	}
	if !foundNesting {
		t.Errorf("expected high nesting finding")
	}
	if !foundComplexity {
		t.Errorf("expected high complexity finding")
	}
	if !foundVolume {
		t.Errorf("expected large AST volume finding")
	}
}

func TestQualityAnalyzer(t *testing.T) {
	repoRoot := setupMockRepository(t)
	qa := NewQualityAnalyzer(repoRoot)

	files := []DiscoveredFile{
		{RelativePath: "backend/srv1.go", Extension: ".go", IsSource: true},
		{RelativePath: "backend/srv2.go", Extension: ".go", IsSource: true},
		// No test file in backend/
	}

	findings, covStr := qa.AnalyzeQuality(files, []string{".github/workflows/ci-go.yml"})
	if covStr == "" {
		t.Errorf("expected non-empty coverage string")
	}

	foundNoTest := false
	for _, f := range findings {
		if strings.HasPrefix(f.FindingID, "RI-QUAL-NOTEST") {
			foundNoTest = true
			if f.Category != CategoryTestQuality {
				t.Errorf("expected CategoryTestQuality, got %s", f.Category)
			}
		}
	}

	if !foundNoTest {
		t.Errorf("expected RI-QUAL-NOTEST finding for backend/ lacking unit tests")
	}
}

func TestAggregator_DeduplicationAndSorting(t *testing.T) {
	agg := NewAggregator()

	findings := []Finding{
		{FindingID: "FINDING-LOW", Severity: SeverityLow, Title: "Low Issue"},
		{FindingID: "FINDING-CRIT", Severity: SeverityCritical, Title: "Critical Issue"},
		{FindingID: "FINDING-MED", Severity: SeverityMedium, Title: "Medium Issue"},
		{FindingID: "FINDING-LOW", Severity: SeverityLow, Title: "Duplicate Low Issue"}, // duplicate
	}

	report := agg.Consolidate(
		RepositoryMetadataRegistry{TotalFiles: 10},
		DependencyGraph{},
		findings,
		AnalysisCoverage{StructuralFilesCount: 10},
		PerformanceMetrics{TotalDurationMs: 42},
		nil,
	)

	if len(report.Findings) != 3 {
		t.Fatalf("expected 3 deduped findings, got %d", len(report.Findings))
	}

	// First finding must be CRITICAL
	if report.Findings[0].Severity != SeverityCritical {
		t.Errorf("expected first finding to be Critical, got %s", report.Findings[0].Severity)
	}
	// Second must be MEDIUM
	if report.Findings[1].Severity != SeverityMedium {
		t.Errorf("expected second finding to be Medium, got %s", report.Findings[1].Severity)
	}
	// Third must be LOW
	if report.Findings[2].Severity != SeverityLow {
		t.Errorf("expected third finding to be Low, got %s", report.Findings[2].Severity)
	}

	// Test Markdown generation
	md := agg.GenerateMarkdown(report)
	if !strings.Contains(md, "# Repository Intelligence Diagnostic Report") {
		t.Errorf("markdown report missing header")
	}
	if !strings.Contains(md, "Critical Issue") {
		t.Errorf("markdown report missing Critical Issue")
	}

	// Test JSON serialization
	jsonBytes, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	var unmarshaled IntelligenceReport
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if len(unmarshaled.Findings) != 3 {
		t.Errorf("expected 3 findings in unmarshaled JSON, got %d", len(unmarshaled.Findings))
	}
}

func TestOrchestrator_EndToEndMock(t *testing.T) {
	repoRoot := setupMockRepository(t)
	outDir := filepath.Join(repoRoot, "reports")

	orch := NewOrchestrator(OrchestratorConfig{
		RepoRoot:  repoRoot,
		OutputDir: outDir,
		Timeout:   10 * time.Second,
	})

	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("Orchestrator Run failed: %v", err)
	}

	if report == nil {
		t.Fatalf("expected non-nil report")
	}
	if len(report.Findings) == 0 {
		t.Errorf("expected findings in mock repo, got 0")
	}

	// Verify emitted files exist
	jsonPath := filepath.Join(outDir, "repository_intelligence.json")
	mdPath := filepath.Join(outDir, "repository_intelligence.md")
	graphPath := filepath.Join(outDir, "repository_dependency_graph.json")

	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("missing output %s", jsonPath)
	}
	if _, err := os.Stat(mdPath); err != nil {
		t.Errorf("missing output %s", mdPath)
	}
	if _, err := os.Stat(graphPath); err != nil {
		t.Errorf("missing output %s", graphPath)
	}
}
