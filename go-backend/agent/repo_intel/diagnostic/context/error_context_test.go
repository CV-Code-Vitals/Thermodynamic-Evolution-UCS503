package context

import (
	"path/filepath"
	"testing"
)

func TestErrorContextAnalyzerOnFixtures(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	analyzer := NewErrorContextAnalyzer(repoRoot)

	fixtureRelPath := filepath.Join("agent", "repo_intel", "diagnostic", "semantic", "fixtures", "error_fixtures.go")
	eval, err := analyzer.EvaluateFile(fixtureRelPath)
	if err != nil {
		t.Fatalf("EvaluateFile failed: %v", err)
	}

	if len(eval.Findings) == 0 {
		t.Fatalf("expected error findings, got 0")
	}

	if len(eval.CounterEvidences) == 0 {
		t.Fatalf("expected safe counter-evidences, got 0")
	}

	hasUnhanled := false
	for _, f := range eval.Findings {
		if f.Category == "RELIABILITY" && f.Severity == "HIGH" {
			hasUnhanled = true
			break
		}
	}

	if !hasUnhanled {
		t.Errorf("expected high severity reliability error findings")
	}

	hasCounterEv := false
	for _, ce := range eval.CounterEvidences {
		if ce.RuleID == "SAFE_ERROR_HANDLING" {
			hasCounterEv = true
			break
		}
	}

	if !hasCounterEv {
		t.Errorf("expected SAFE_ERROR_HANDLING counter-evidences")
	}
}
