// =============================================================================
// go-backend/agent/agent_test.go
//
// Unit tests for the Agentic Thermodynamic Audit Engine.
// =============================================================================

package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func makeTestReport() *ThermodynamicReport {
	return &ThermodynamicReport{
		EngineVersion:    "2.0.0",
		GeneratedAt:      "2026-10-05T00:00:00Z",
		ScannedDirectory: "/tmp/test_code",
		FilesAnalyzed:    2,
		TotalHotspots:    4,
		GlobalEntropy:    42.5,
		FileReports: []FileReport{
			{
				FilePath:     "/tmp/test_code/crawler.go",
				Language:     "Go",
				LinesScanned: 150,
				TotalEntropy: 30.0,
				Hotspots: []Hotspot{
					{
						FunctionName:      "CrawlBatch",
						LineNumber:        39,
						ByteRange:         [2]int{500, 1200},
						SourceSnippet:     "func (c *Crawler) CrawlBatch(urls []string) []CrawlResult {",
						EntropyScore:      12.5,
						VulnerabilityType: "BlockingIO",
						Description:       "Blocking I/O call(s) inside control flow",
						Metrics: FunctionMetrics{
							CyclomaticComplexity: 8,
							MaxNestingDepth:      3,
							ASTNodeCount:         95,
							ConcurrencyPrims:     2,
							Energy:               12.5,
						},
					},
					{
						FunctionName:      "BuildIndex",
						LineNumber:        79,
						ByteRange:         [2]int{1300, 1800},
						SourceSnippet:     "func BuildIndex(results []CrawlResult) map[string][]string {",
						EntropyScore:      10.2,
						VulnerabilityType: "DeepNesting",
						Description:       "Nesting depth 3 — O(n^3) complexity risk",
						Metrics: FunctionMetrics{
							CyclomaticComplexity: 6,
							MaxNestingDepth:      4,
							ASTNodeCount:         72,
							ConcurrencyPrims:     0,
							Energy:               10.2,
						},
					},
				},
			},
			{
				FilePath:     "/tmp/test_code/data_pipeline.py",
				Language:     "Python",
				LinesScanned: 100,
				TotalEntropy: 12.5,
				Hotspots: []Hotspot{
					{
						FunctionName:      "recursive_flatten",
						LineNumber:        58,
						ByteRange:         [2]int{2000, 2400},
						SourceSnippet:     "def recursive_flatten(data, depth=0):",
						EntropyScore:      9.8,
						VulnerabilityType: "RecursiveCall",
						Description:       "Recursive call — stack-depth risk",
						Metrics: FunctionMetrics{
							CyclomaticComplexity: 5,
							MaxNestingDepth:      3,
							ASTNodeCount:         45,
							ConcurrencyPrims:     0,
							Energy:               9.8,
						},
					},
					{
						FunctionName:      "process_batch",
						LineNumber:        79,
						ByteRange:         [2]int{2500, 3200},
						SourceSnippet:     "def process_batch(self, batch_ids):",
						EntropyScore:      14.0,
						VulnerabilityType: "SyncContention",
						Description:       "Lock contention at nesting depth 3",
						Metrics: FunctionMetrics{
							CyclomaticComplexity: 10,
							MaxNestingDepth:      4,
							ASTNodeCount:         110,
							ConcurrencyPrims:     3,
							Energy:               14.0,
						},
					},
				},
			},
		},
	}
}

// ── Phase 1 Tests ────────────────────────────────────────────────────────────

func TestRunAudit_Phase1_TemperatureClassification(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	if audit.SystemTemperature == "" {
		t.Fatal("SystemTemperature should not be empty")
	}
	t.Logf("System Temperature: %s", audit.SystemTemperature)
	t.Logf("Average Energy: %.2f", audit.AverageFuncEnergy)
}

func TestRunAudit_Phase1_VolatilityRanking(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	if len(audit.VolatilityRanking) == 0 {
		t.Fatal("VolatilityRanking should not be empty")
	}

	// Verify descending energy order
	for i := 1; i < len(audit.VolatilityRanking); i++ {
		if audit.VolatilityRanking[i].Energy > audit.VolatilityRanking[i-1].Energy {
			t.Errorf("Volatility ranking not sorted: rank %d (E=%.2f) > rank %d (E=%.2f)",
				i+1, audit.VolatilityRanking[i].Energy,
				i, audit.VolatilityRanking[i-1].Energy)
		}
	}

	// Top entry should be process_batch (E=14.0)
	top := audit.VolatilityRanking[0]
	if top.FunctionName != "process_batch" {
		t.Errorf("Expected top hotspot 'process_batch', got %q", top.FunctionName)
	}
	t.Logf("Top hotspot: %s (E=%.2f)", top.FunctionName, top.Energy)
}

// ── Phase 2 Tests ────────────────────────────────────────────────────────────

func TestRunAudit_Phase2_StressFailures(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	if len(audit.StressFailures) == 0 {
		t.Fatal("StressFailures should not be empty for volatile functions")
	}

	for _, sf := range audit.StressFailures {
		if sf.FailureMode == "" {
			t.Error("FailureMode should not be empty")
		}
		if sf.Mechanism == "" {
			t.Error("Mechanism should not be empty")
		}
		if sf.Severity == "" {
			t.Error("Severity should not be empty")
		}
		t.Logf("Failure: %s — %s (%s)", sf.Component, sf.FailureMode, sf.Severity)
	}
}

// ── Phase 3 Tests ────────────────────────────────────────────────────────────

func TestRunAudit_Phase3_Directives(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	if len(audit.Directives) == 0 {
		t.Fatal("Directives should not be empty")
	}
	if len(audit.Directives) > 3 {
		t.Errorf("Expected at most 3 directives, got %d", len(audit.Directives))
	}

	for _, d := range audit.Directives {
		if d.DeltaEnergy >= 0 {
			t.Errorf("DeltaEnergy for %s should be negative (improvement), got %.2f",
				d.TargetFunction, d.DeltaEnergy)
		}
		if d.CodePatch == "" {
			t.Errorf("CodePatch for %s should not be empty", d.TargetFunction)
		}
		t.Logf("Directive: %s — E: %.2f → %.2f (ΔE=%.2f)",
			d.TargetFunction, d.OriginalEnergy, d.ProjectedEnergy, d.DeltaEnergy)
	}
}

// ── Invariants Test ──────────────────────────────────────────────────────────

func TestRunAudit_Invariants(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	if len(audit.Invariants) == 0 {
		t.Fatal("Invariants should not be empty")
	}
	t.Logf("Generated %d stability invariants", len(audit.Invariants))
}

// ── Output serialization tests ──────────────────────────────────────────────

func TestWriteAuditJSON(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	tmp := t.TempDir()
	jsonPath := filepath.Join(tmp, "audit.json")

	if err := WriteAuditJSON(audit, jsonPath); err != nil {
		t.Fatalf("WriteAuditJSON: %v", err)
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Read audit JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Audit JSON should not be empty")
	}
	t.Logf("Audit JSON: %d bytes", len(data))
}

func TestWriteAuditMarkdown(t *testing.T) {
	report := makeTestReport()
	audit := RunAudit(report, 8.0)

	tmp := t.TempDir()
	mdPath := filepath.Join(tmp, "audit.md")

	if err := WriteAuditMarkdown(audit, mdPath); err != nil {
		t.Fatalf("WriteAuditMarkdown: %v", err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("Read audit MD: %v", err)
	}

	content := string(data)
	if !contains(content, "Executive Entropy Summary") {
		t.Error("Markdown should contain 'Executive Entropy Summary'")
	}
	if !contains(content, "Volatility Ranking") {
		t.Error("Markdown should contain 'Volatility Ranking'")
	}
	if !contains(content, "Stress Testing") {
		t.Error("Markdown should contain 'Stress Testing'")
	}
	if !contains(content, "Rectification") {
		t.Error("Markdown should contain 'Rectification'")
	}
	if !contains(content, "Stability Guardrails") {
		t.Error("Markdown should contain 'Stability Guardrails'")
	}
	t.Logf("Audit Markdown: %d bytes", len(data))
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ── Energy computation tests ─────────────────────────────────────────────────

func TestComputeProjectedEnergy(t *testing.T) {
	// E = 0.45*10 + 0.25*4 + 0.15*log2(128) + 0.15*2
	// E = 4.5 + 1.0 + 1.05 + 0.3 = 6.85
	e := computeProjectedEnergy(10, 4, 128, 2)
	if e < 6.5 || e > 7.2 {
		t.Errorf("Expected energy ~6.85, got %.2f", e)
	}
}

func TestClassifyTemperature(t *testing.T) {
	tests := []struct {
		avg      float64
		hot      int
		total    int
		expected SystemTemperature
	}{
		{2.0, 1, 20, TempCool},
		{5.0, 3, 20, TempStable},
		{9.0, 6, 20, TempVolatile},
		{15.0, 10, 20, TempCritical},
	}

	for _, tt := range tests {
		got := classifyTemperature(tt.avg, tt.hot, tt.total)
		if got != tt.expected {
			t.Errorf("classifyTemperature(%.1f, %d, %d) = %s, want %s",
				tt.avg, tt.hot, tt.total, got, tt.expected)
		}
	}
}
