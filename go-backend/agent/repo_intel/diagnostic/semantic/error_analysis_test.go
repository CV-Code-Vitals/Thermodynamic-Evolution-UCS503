package semantic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestErrorAnalyzerOnFixtures(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "error_fixtures.go")
	sa := NewSemanticAnalyzer(".")
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on error fixtures: %v", err)
	}

	errAnalyzer := NewErrorAnalyzer(".")
	for _, cfg := range model.CFGs {
		errAnalyzer.RegisterLocalFunction(cfg.FuncDecl)
	}

	expectedOutcomes := map[string]struct {
		isPositive    bool
		expectedState ErrorHandlingClassification
	}{
		"Pos_Err_UnhandledUnmarshal":         {true, ErrorUnhandled},
		"Pos_Err_UnhandledMarshalDiscard":    {true, ErrorExplicitlyIgnored},
		"Pos_Err_UnhandledWriteFile":         {true, ErrorUnhandled},
		"Pos_Err_UnhandledDBExec":            {true, ErrorUnhandled},
		"Pos_Err_UnhandledDial":              {true, ErrorUnhandled},
		"Pos_Err_ExplicitBlankCriticalOp":    {true, ErrorExplicitlyIgnored},
		"Pos_Err_MultipleReturnDiscardError": {true, ErrorExplicitlyIgnored},
		"Pos_Err_OverwrittenBeforeCheck":     {true, ErrorUnhandled},
		"Pos_Err_MismatchedErrorCheck":       {true, ErrorUnhandled},
		"Pos_Err_UnsafeBranchContinue":       {true, ErrorPartiallyHandled},

		"Neg_Err_ExplicitIfCheck":        {false, ErrorChecked},
		"Neg_Err_DirectPropagation":      {false, ErrorPropagated},
		"Neg_Err_WrappedPropagation":     {false, ErrorWrappedAndPropagated},
		"Neg_Err_NamedReturn":            {false, ErrorPropagated},
		"Neg_Err_MultipleReturnHandled":  {false, ErrorChecked},
		"Neg_Err_IgnoredCleanup":         {false, ErrorExplicitlyIgnored},
		"Neg_Err_TxRollbackWithCommit":   {false, ErrorPropagated}, // Commit is returned
		"Neg_Err_TestHelperFatal":        {false, ErrorChecked},
		"Neg_Err_VoidNoErrors":           {false, ""}, // Void without calls
		"Neg_Err_HelperCheck":            {false, ErrorChecked},
	}

	allEvs := make(map[string][]ErrorEvidence)
	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath, nil)
		if err != nil {
			t.Errorf("AnalyzeFunction failed on %s: %v", fnName, err)
		}
		allEvs[fnName] = evs
	}

	for fnName, expected := range expectedOutcomes {
		evs := allEvs[fnName]
		if !expected.isPositive && expected.expectedState == "" {
			// e.g. Neg_Err_VoidNoErrors has no error calls
			hasVuln := false
			for _, ev := range evs {
				if ev.Classification == ErrorUnhandled || (ev.Classification == ErrorExplicitlyIgnored && ev.Category != ErrorCatCleanup) {
					hasVuln = true
				}
			}
			if hasVuln {
				t.Errorf("fixture %s produced unexpected vulnerability finding", fnName)
			}
			continue
		}

		if len(evs) == 0 {
			t.Errorf("fixture %s produced 0 error evidence records", fnName)
			continue
		}

		matched := false
		for _, ev := range evs {
			if ev.Classification == expected.expectedState {
				matched = true
				break
			}
		}

		if !matched {
			var actualStates []string
			for _, ev := range evs {
				actualStates = append(actualStates, string(ev.Classification))
			}
			t.Errorf("fixture %s: expected %s, got %v", fnName, expected.expectedState, actualStates)
		}
	}
}

func TestErrorAnalyzerOnExternalHistoricalBenchmarks(t *testing.T) {
	errAnalyzer := NewErrorAnalyzer(".")

	// 1. External Repo: http_router (DEF-HTTP-02: context.go:61 json.NewEncoder(c.Writer).Encode(data))
	httpPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "http_router")
	saHttp := NewSemanticAnalyzer(httpPath)
	httpModel, _, err := saHttp.AnalyzeFile("context.go")
	if err != nil {
		t.Fatalf("AnalyzeFile failed on http_router context.go: %v", err)
	}

	foundDefHttp02 := false
	for fnName, cfg := range httpModel.CFGs {
		df := httpModel.Dataflow[fnName]
		evs, _ := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, httpModel.FileSet, cfg, df, "context.go", nil)
		for _, ev := range evs {
			if (fnName == "JSON" || fnName == "Context.JSON" || cfg.FuncDecl.Name.Name == "JSON") && ev.Classification == ErrorUnhandled && ev.Category == ErrorCatSerialization {
				foundDefHttp02 = true
			}
		}
	}
	if !foundDefHttp02 {
		t.Errorf("FAILED to detect DEF-HTTP-02 in http_router context.go (unhandled json.Encode)")
	}

	// 2. External Repo: sql_service (DEF-SQL-03: db.go:57 tx.Rollback() unhandled)
	sqlPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")
	saSql := NewSemanticAnalyzer(sqlPath)
	sqlModel, _, err := saSql.AnalyzeFile("db.go")
	if err != nil {
		t.Fatalf("AnalyzeFile failed on sql_service db.go: %v", err)
	}

	foundDefSql03 := false
	for fnName, cfg := range sqlModel.CFGs {
		df := sqlModel.Dataflow[fnName]
		evs, _ := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, sqlModel.FileSet, cfg, df, "db.go", nil)
		for _, ev := range evs {
			if fnName == "WithTransaction" && ev.Classification == ErrorUnhandled && ev.Category == ErrorCatCleanup {
				foundDefSql03 = true
			}
		}
	}
	if !foundDefSql03 {
		t.Errorf("FAILED to detect DEF-SQL-03 in sql_service db.go (unhandled tx.Rollback)")
	}

	// 3. External Repo: queue_worker (DEF-QUEUE-03: internal/consumer/consumer.go:60 json.Unmarshal unhandled)
	queuePath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "queue_worker")
	saQueue := NewSemanticAnalyzer(queuePath)
	queueModel, _, err := saQueue.AnalyzeFile(filepath.Join("internal", "consumer", "consumer.go"))
	if err != nil {
		t.Fatalf("AnalyzeFile failed on queue_worker consumer.go: %v", err)
	}

	foundDefQueue03 := false
	for fnName, cfg := range queueModel.CFGs {
		df := queueModel.Dataflow[fnName]
		evs, _ := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, queueModel.FileSet, cfg, df, "internal/consumer/consumer.go", nil)
		for _, ev := range evs {
			if fnName == "ParseRawPayload" && ev.Classification == ErrorUnhandled && ev.Category == ErrorCatSerialization {
				foundDefQueue03 = true
			}
		}
	}
	if !foundDefQueue03 {
		t.Errorf("FAILED to detect DEF-QUEUE-03 in queue_worker consumer.go (unhandled json.Unmarshal)")
	}
}

func TestExportErrorEvidenceReport(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "error_fixtures.go")
	sa := NewSemanticAnalyzer(".")
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	errAnalyzer := NewErrorAnalyzer(".")
	for _, cfg := range model.CFGs {
		errAnalyzer.RegisterLocalFunction(cfg.FuncDecl)
	}

	var allFixtureEvs []ErrorEvidence
	tp, fp, tn, fn := 0, 0, 0, 0

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, _ := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath, nil)
		allFixtureEvs = append(allFixtureEvs, evs...)

		isPositiveFixture := len(fnName) >= 3 && fnName[:3] == "Pos"
		hasVulnerableEv := false
		for _, ev := range evs {
			if ev.Classification == ErrorUnhandled ||
				(ev.Classification == ErrorExplicitlyIgnored && ev.Category != ErrorCatCleanup) ||
				ev.Classification == ErrorPartiallyHandled {
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
		"report_type": "Milestone 3.4 Phase 4 Error Handling Model and Fixture Results",
		"timestamp":   time.Now().Format(time.RFC3339),
		"analyzer":    "semantic.ErrorAnalyzer",
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
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_4_error_fixture_results.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("write report error: %v", err)
	}

	// Also export detailed cryptographic evidence report
	evidenceReport := map[string]interface{}{
		"report_type": "Milestone 3.4 Phase 4 Error Handling Cryptographic Evidence Report",
		"timestamp":   time.Now().Format(time.RFC3339),
		"analyzer":    "semantic.ErrorAnalyzer",
		"evidences":   allFixtureEvs,
	}
	evData, _ := json.MarshalIndent(evidenceReport, "", "  ")
	evPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_4_error_handling_evidence.json")
	if err := os.WriteFile(evPath, evData, 0644); err != nil {
		t.Fatalf("write evidence report error: %v", err)
	}
}
