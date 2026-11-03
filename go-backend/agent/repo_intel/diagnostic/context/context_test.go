package context

import (
	"path/filepath"
	"testing"
)

func TestContextEngineOnSelfHealingSuite(t *testing.T) {
	suitePath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "self_healing_suite")

	ce := NewContextEngine(suitePath)

	files := []string{
		"architecture/boundary/api_to_db.go",
		"architecture/circular/cycle.go",
		"documentation/undocumented_service/service.go",
		"maintainability/duplicate_helpers.go",
		"reliability/resource_lifecycle.go",
		"security/clean_sql.go",
		"security/clean_exec.go",
		"security/sql_injection.go",
		"testing/untested_billing.go",
		"testing/tested_auth.go",
		"testing/tested_auth_test.go",
		"go.mod",
	}

	summary, findings, counterEvs, err := ce.AnalyzeRepository(files)
	if err != nil {
		t.Fatalf("AnalyzeRepository failed: %v", err)
	}

	if summary.TotalFilesScanned != len(files) {
		t.Errorf("Expected %d files scanned, got %d", len(files), summary.TotalFilesScanned)
	}

	// Verify counter-evidence generated for clean_sql.go and clean_exec.go
	if len(counterEvs["security/clean_sql.go"]) == 0 {
		t.Errorf("Expected counter-evidence for clean_sql.go, got 0")
	}
	if len(counterEvs["security/clean_exec.go"]) == 0 {
		t.Errorf("Expected counter-evidence for clean_exec.go, got 0")
	}

	// Verify that contextual findings were detected
	foundResourceLeak := false
	foundDuplicate := false
	foundLayerBypass := false
	foundMutualCycle := false
	foundUntestedSym := false
	foundMissingDoc := false
	foundNoLock := false

	for _, f := range findings {
		switch {
		case f.FindingID == "CTX-REL-LEAK-LOOP-reliability-resource_lifecycle_go-L13":
			foundResourceLeak = true
		case f.Category == "MAINTAINABILITY" && len(f.EvidenceTraces) >= 2:
			foundDuplicate = true
		case f.Category == "ARCHITECTURE" && f.FindingID == "CTX-ARC-LAYER-BYPASS-architecture-boundary-api_to_db_go":
			foundLayerBypass = true
		case f.Category == "ARCHITECTURE" && len(f.EvidenceTraces) == 2:
			foundMutualCycle = true
		case f.Category == "TEST_QUALITY" && f.FindingID == "CTX-TST-UNTESTED-SYM-testing-untested_billing_go":
			foundUntestedSym = true
		case f.Category == "DOCUMENTATION" && f.FindingID == "CTX-DOC-MISSING-GODOC-documentation-undocumented_service-service_go":
			foundMissingDoc = true
		case f.Category == "DEPENDENCY" && f.FindingID == "CTX-DEP-NOLOCK-go_mod":
			foundNoLock = true
		}
	}

	if !foundResourceLeak {
		t.Errorf("Expected resource leak finding in resource_lifecycle.go")
	}
	if !foundDuplicate {
		t.Errorf("Expected duplicate function finding in duplicate_helpers.go")
	}
	if !foundLayerBypass {
		t.Errorf("Expected layer bypass finding in api_to_db.go")
	}
	if !foundMutualCycle {
		t.Errorf("Expected mutual cycle finding in cycle.go")
	}
	if !foundUntestedSym {
		t.Errorf("Expected untested symbol finding in untested_billing.go")
	}
	if !foundMissingDoc {
		t.Errorf("Expected missing godoc finding in service.go")
	}
	if !foundNoLock {
		t.Errorf("Expected missing lockfile finding in go.mod")
	}
}
