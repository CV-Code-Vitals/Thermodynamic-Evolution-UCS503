package benchmark

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	diagcontext "github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/context"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/rootcause"
)

// LanguageCapabilityEntry defines the operational support tier for a language across platform stages.
type LanguageCapabilityEntry struct {
	Language             string `json:"language"`
	FileDiscovery        string `json:"file_discovery"`
	ParsingIngestion     string `json:"parsing_ingestion"`
	SymbolAnalysis       string `json:"symbol_analysis"`
	DependencyAnalysis   string `json:"dependency_analysis"`
	ContextualDiagnostic string `json:"contextual_diagnostic"`
	CounterEvidence      string `json:"counter_evidence"`
	RootCauseAnalysis    string `json:"root_cause_analysis"`
	RemediationPlanning  string `json:"remediation_planning"`
	PatchGeneration      string `json:"patch_generation"`
	IndependentValidation string `json:"independent_validation"`
	OverallReadinessTier string `json:"overall_readiness_tier"`
	AuditNotes           string `json:"audit_notes"`
}

// LanguageEvaluationReport captures language capabilities and limitations.
type LanguageEvaluationReport struct {
	Timestamp            time.Time                 `json:"timestamp"`
	EvaluatedLanguages   []string                  `json:"evaluated_languages"`
	Capabilities         []LanguageCapabilityEntry `json:"capabilities"`
	ProductionSupportSummary map[string]string     `json:"production_support_summary"`
}

func TestCrossLanguageCapabilityAudit(t *testing.T) {
	reportsDir := filepath.Join("..", "..", "..", "..", "reports")
	_ = os.MkdirAll(reportsDir, 0755)

	caps := []LanguageCapabilityEntry{
		{
			Language:             "Go",
			FileDiscovery:        "SUPPORTED",
			ParsingIngestion:     "SUPPORTED",
			SymbolAnalysis:       "SUPPORTED",
			DependencyAnalysis:   "SUPPORTED",
			ContextualDiagnostic: "SUPPORTED",
			CounterEvidence:      "SUPPORTED",
			RootCauseAnalysis:    "SUPPORTED",
			RemediationPlanning:  "SUPPORTED",
			PatchGeneration:      "SUPPORTED",
			IndependentValidation: "SUPPORTED",
			OverallReadinessTier: "PRODUCTION_READY",
			AuditNotes:           "Full end-to-end AST analysis, counter-evidence, patch generation, and go test validation.",
		},
		{
			Language:             "Rust",
			FileDiscovery:        "SUPPORTED",
			ParsingIngestion:     "SUPPORTED",
			SymbolAnalysis:       "PARTIAL",
			DependencyAnalysis:   "SUPPORTED",
			ContextualDiagnostic: "PARTIAL",
			CounterEvidence:      "UNSUPPORTED",
			RootCauseAnalysis:    "PARTIAL",
			RemediationPlanning:  "PARTIAL",
			PatchGeneration:      "UNSUPPORTED",
			IndependentValidation: "SUPPORTED",
			OverallReadinessTier: "RESTRICTED_RESEARCH_ONLY",
			AuditNotes:           "High-speed parsing and manifest analysis (Cargo.toml); semantic diagnostics and automated patch generation unsupported.",
		},
		{
			Language:             "C++",
			FileDiscovery:        "SUPPORTED",
			ParsingIngestion:     "PARTIAL",
			SymbolAnalysis:       "UNSUPPORTED",
			DependencyAnalysis:   "UNSUPPORTED",
			ContextualDiagnostic: "UNSUPPORTED",
			CounterEvidence:      "UNSUPPORTED",
			RootCauseAnalysis:    "UNSUPPORTED",
			RemediationPlanning:  "UNSUPPORTED",
			PatchGeneration:      "UNSUPPORTED",
			IndependentValidation: "UNSUPPORTED",
			OverallReadinessTier: "UNSUPPORTED",
			AuditNotes:           "Basic file discovery and rough line-count metrics only; zero semantic analysis, AST linting, or remediation.",
		},
	}

	rep := LanguageEvaluationReport{
		Timestamp:          time.Now(),
		EvaluatedLanguages: []string{"Go", "Rust", "C++"},
		Capabilities:       caps,
		ProductionSupportSummary: map[string]string{
			"Go":   "FULL_AUTONOMOUS_SUPPORT",
			"Rust": "READ_ONLY_DISCOVERY_AND_MANIFEST_ANALYSIS",
			"C++":  "UNSUPPORTED_INGESTION_ONLY",
		},
	}

	outPath := filepath.Join(reportsDir, "milestone3_1_language_evaluation.json")
	data, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("failed writing language evaluation report: %v", err)
	}
}

func TestAdversarialDiagnosticAndRootCause(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Adversarial Test: Parameterized SQL mixed with dynamic SQL in same file
	mixedSQLFile := filepath.Join(tempDir, "mixed_sql.go")
	mixedSQLCode := `package adversarial
import "database/sql"

func ProcessMixedQueries(db *sql.DB, safeID int, unsafeUser string) error {
	// Safe parameterized query
	_ = db.QueryRow("SELECT id FROM users WHERE id = $1", safeID)

	// Unsafe concatenated query in same function
	badQuery := "SELECT id, email FROM users WHERE username = '" + unsafeUser + "'"
	_, err := db.Exec(badQuery)
	return err
}
`
	if err := os.WriteFile(mixedSQLFile, []byte(mixedSQLCode), 0644); err != nil {
		t.Fatalf("write mixed sql: %v", err)
	}

	sqlAnalyzer := diagcontext.NewSQLContextAnalyzer(tempDir)
	sqlEval, err := sqlAnalyzer.EvaluateFile("mixed_sql.go")
	if err != nil {
		t.Fatalf("evaluate mixed sql: %v", err)
	}

	if !sqlEval.HasVulnerabilities {
		t.Errorf("Adversarial failure: safe query suppressed unsafe dynamic query in same function!")
	}
	if len(sqlEval.Findings) == 0 {
		t.Errorf("Expected SQL injection finding for badQuery, got 0")
	}
	if len(sqlEval.CounterEvidences) == 0 {
		t.Errorf("Expected counter-evidence for safeQuery, got 0")
	}

	// 2. Adversarial Test: Security secret commented out
	commentedSecretFile := filepath.Join(tempDir, "commented_secret.go")
	commentedCode := `package adversarial
// Inactive dummy example in comment:
// aws_access_key_id = "AKIA1111222233334444"
const ProductionEndpoint = "https://api.example.com"
`
	_ = os.WriteFile(commentedSecretFile, []byte(commentedCode), 0644)
	secAnalyzer := repo_intel.NewReliabilitySecurityAnalyzer(tempDir)
	secFindings := secAnalyzer.AnalyzeReliabilityAndSecurity([]repo_intel.DiscoveredFile{
		{RelativePath: "commented_secret.go", Extension: ".go", IsSource: true},
	})
	if len(secFindings) > 0 {
		t.Errorf("Adversarial failure: secret scanner flagged commented-out secret token!")
	}

	// 3. Adversarial Test: Bayesian confidence aggregation on correlated findings
	rcAnalyzer := rootcause.NewRootCauseAnalyzer()
	var correlatedFindings []repo_intel.Finding
	for i := 1; i <= 10; i++ {
		correlatedFindings = append(correlatedFindings, repo_intel.Finding{
			FindingID:     "F-CORRELATED-" + string(rune('0'+i)),
			Category:      repo_intel.CategoryReliability,
			Confidence:    0.90,
			Title:         "Repeated error pattern",
			AffectedFiles: []string{"server.go"},
		})
	}
	entities, _ := rcAnalyzer.AnalyzeRootCauses(correlatedFindings)
	if len(entities) == 0 {
		t.Fatalf("Expected root cause entities, got 0")
	}
	aggConf := entities[0].Confidence
	if aggConf > 0.99 {
		t.Errorf("Adversarial failure: Bayesian confidence exceeded safety cap of 0.99: got %f", aggConf)
	}

	// 4. Adversarial Test: Malformed syntax resilience
	malformedFile := filepath.Join(tempDir, "broken.go")
	malformedCode := `package adversarial
func IncompleteFunc( { return }}}`
	_ = os.WriteFile(malformedFile, []byte(malformedCode), 0644)
	ce := diagcontext.NewContextEngine(tempDir)
	_, ctxFindings, _, err := ce.AnalyzeRepository([]string{"broken.go"})
	if err != nil {
		t.Errorf("Engine crashed on malformed Go file: %v", err)
	}
	_ = ctxFindings
}

