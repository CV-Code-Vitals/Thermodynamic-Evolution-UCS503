package semantic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLAnalyzerOnFixtures(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "sql_fixtures.go")
	sa := NewSemanticAnalyzer(".")

	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on fixtures: %v", err)
	}

	sqlAnalyzer := NewSQLAnalyzer(".")

	// Expected classifications for each fixture function
	expectedResults := map[string]struct {
		ExpectedClassification SQLClassification
		ShouldBeVulnerable     bool
		MinConfidence          float64
	}{
		"Pos_SQL_Sprintf":              {ClassificationUntrustedSQLStructure, true, 0.95},
		"Pos_SQL_Concat":               {ClassificationUntrustedSQLStructure, true, 0.95},
		"Pos_SQL_PropagatedLocal":      {ClassificationUntrustedSQLStructure, true, 0.95},
		"Pos_SQL_ParamAsQuery":         {ClassificationUntrustedSQLStructure, true, 0.95},
		"Pos_SQL_ContextSink":          {ClassificationUntrustedSQLStructure, true, 0.95},
		"Neg_SQL_StaticQuery":          {ClassificationStaticSQL, false, 0.98},
		"Neg_SQL_ParameterizedQuery":   {ClassificationSafeParameterizedSQL, false, 0.98},
		"Neg_SQL_ParameterizedContext": {ClassificationSafeParameterizedSQL, false, 0.98},
		"Neg_SQL_PreparedStatement":    {ClassificationSafeParameterizedSQL, false, 0.90},
	}

	allEvidence := make(map[string][]SQLEvidence)

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		if err != nil {
			t.Errorf("AnalyzeFunction error on %s: %v", fnName, err)
		}
		allEvidence[fnName] = evs
	}

	// Verify non-sink negative fixtures produce 0 sink findings
	nonSinkFixtures := []string{"Neg_SQL_SprintfNonSQL", "Neg_SQL_DynamicUnused"}
	for _, nsf := range nonSinkFixtures {
		evs := allEvidence[nsf]
		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedSQLStructure {
				t.Errorf("fixture %s produced unexpected SQL injection finding: %+v", nsf, ev)
			}
		}
	}

	// Verify expected classifications on tested fixtures
	for fnName, expected := range expectedResults {
		evs, ok := allEvidence[fnName]
		if !ok || len(evs) == 0 {
			t.Errorf("fixture %s produced no SQL evidence records", fnName)
			continue
		}

		matched := false
		for _, ev := range evs {
			if ev.Classification == expected.ExpectedClassification {
				matched = true
				if ev.Confidence < expected.MinConfidence {
					t.Errorf("fixture %s: confidence %.2f lower than expected %.2f", fnName, ev.Confidence, expected.MinConfidence)
				}
				if expected.ShouldBeVulnerable && ev.LifecycleState != StateConfirmed {
					t.Errorf("fixture %s: expected StateConfirmed, got %v", fnName, ev.LifecycleState)
				}
				if !expected.ShouldBeVulnerable && len(ev.CounterEvidence) == 0 {
					t.Errorf("fixture %s: expected counter-evidence for safe control, got none", fnName)
				}
			}
		}

		if !matched {
			var got []SQLClassification
			for _, e := range evs {
				got = append(got, e.Classification)
			}
			t.Errorf("fixture %s: expected classification %s, got %v", fnName, expected.ExpectedClassification, got)
		}
	}
}

func TestSQLAnalyzerOnExternalSQLService(t *testing.T) {
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")
	targetFile := "repository.go"

	sa := NewSemanticAnalyzer(repoPath)
	model, _, err := sa.AnalyzeFile(targetFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on sql_service/repository.go: %v", err)
	}

	sqlAnalyzer := NewSQLAnalyzer(repoPath)
	foundDefSQL01 := false
	foundSafeGetUser := false

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, targetFile)
		if err != nil {
			t.Fatalf("AnalyzeFunction failed on %s: %v", fnName, err)
		}

		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedSQLStructure {
				if fnName == "UserRepository.FindUserByNameRaw" || fnName == "FindUserByNameRaw" {
					foundDefSQL01 = true
					if len(ev.UntrustedInputs) == 0 {
						t.Errorf("expected untrusted input trace in FindUserByNameRaw")
					}
					if ev.Confidence < 0.95 {
						t.Errorf("expected confidence >= 0.95, got %.2f", ev.Confidence)
					}
				}
			}

			if ev.Classification == ClassificationSafeParameterizedSQL {
				if fnName == "UserRepository.GetUserByID" || fnName == "GetUserByID" ||
					fnName == "UserRepository.CreateUser" || fnName == "CreateUser" {
					foundSafeGetUser = true
				}
			}
		}
	}

	if !foundDefSQL01 {
		t.Errorf("FAILED to detect DEF-SQL-01 in sql_service/repository.go (FindUserByNameRaw fmt.Sprintf SQL injection)")
	}

	if !foundSafeGetUser {
		t.Errorf("FAILED to certify safe parameterized queries in sql_service/repository.go")
	}
}

func TestExportSQLEvidenceReport(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "sql_fixtures.go")
	sa := NewSemanticAnalyzer(".")
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	sqlAnalyzer := NewSQLAnalyzer(".")
	var allFixtureEvs []SQLEvidence
	tp, fp, tn, fn := 0, 0, 0, 0

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, _ := sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		allFixtureEvs = append(allFixtureEvs, evs...)

		isPositiveFixture := len(fnName) >= 3 && fnName[:3] == "Pos"
		hasVulnerableEv := false
		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedSQLStructure {
				hasVulnerableEv = true
			}
		}

		if isPositiveFixture {
			if hasVulnerableEv {
				tp++
			} else {
				fn++
			}
		} else {
			if hasVulnerableEv {
				fp++
			} else {
				tn++
			}
		}
	}

	// External evaluation on sql_service
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")
	saExt := NewSemanticAnalyzer(repoPath)
	extModel, _, err := saExt.AnalyzeFile("repository.go")
	var extEvs []SQLEvidence
	if err == nil && extModel != nil {
		sqlExt := NewSQLAnalyzer(repoPath)
		for fnName, cfg := range extModel.CFGs {
			df := extModel.Dataflow[fnName]
			evs, _ := sqlExt.AnalyzeFunction(cfg.FuncDecl, extModel.FileSet, cfg, df, "repository.go")
			extEvs = append(extEvs, evs...)
		}
	}

	precision := 0.0
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	recall := 0.0
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	f1 := 0.0
	if precision+recall > 0 {
		f1 = 2 * (precision * recall) / (precision + recall)
	}

	report := map[string]interface{}{
		"report_type": "Milestone 3.4 Phase 2 SQL Expression Evidence",
		"timestamp":   time.Now().Format(time.RFC3339),
		"analyzer":    "semantic.SQLAnalyzer",
		"confusion_matrix": map[string]int{
			"true_positives":  tp,
			"false_positives": fp,
			"true_negatives":  tn,
			"false_negatives": fn,
		},
		"metrics": map[string]float64{
			"precision":   precision,
			"recall":      recall,
			"f1_score":    f1,
			"specificity": float64(tn) / float64(tn+fp),
		},
		"fixtures_evaluated": allFixtureEvs,
		"external_findings":  extEvs,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_4_sql_expression_evidence.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("write report error: %v", err)
	}
}
