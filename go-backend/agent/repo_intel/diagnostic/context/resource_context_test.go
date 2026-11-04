package context

import (
	"path/filepath"
	"testing"
)

func TestResourceContextAnalyzerOnFixtures(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	analyzer := NewResourceContextAnalyzer(repoRoot)

	fixtureRelPath := filepath.Join("agent", "repo_intel", "diagnostic", "semantic", "fixtures", "resource_fixtures.go")
	eval, err := analyzer.EvaluateFile(fixtureRelPath)
	if err != nil {
		t.Fatalf("EvaluateFile failed: %v", err)
	}

	if len(eval.Findings) == 0 {
		t.Fatalf("expected resource findings, got 0")
	}

	if len(eval.CounterEvidences) == 0 {
		t.Fatalf("expected safe counter-evidences, got 0")
	}

	// Verify that confirmed leaks and counter-evidences are discovered
	hasLeakFinding := false
	for _, f := range eval.Findings {
		if f.Category == "RELIABILITY" && f.Severity == "HIGH" {
			hasLeakFinding = true
			break
		}
	}

	if !hasLeakFinding {
		t.Errorf("expected high severity reliability resource leak findings")
	}

	hasCounterEv := false
	for _, ce := range eval.CounterEvidences {
		if ce.RuleID == "SAFE_RESOURCE_LIFECYCLE" {
			hasCounterEv = true
			break
		}
	}

	if !hasCounterEv {
		t.Errorf("expected SAFE_RESOURCE_LIFECYCLE counter-evidences")
	}
}

