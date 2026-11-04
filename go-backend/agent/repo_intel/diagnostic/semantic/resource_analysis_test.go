package semantic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResourceAnalyzerOnFixtures(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "resource_fixtures.go")
	sa := NewSemanticAnalyzer(".")

	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on resource fixtures: %v", err)
	}

	resAnalyzer := NewResourceAnalyzer(".")

	// Expected classifications for all 24 required test cases
	expectedResults := map[string]struct {
		ExpectedState      ResourceLifecycleState
		ShouldBeVulnerable bool
		MinConfidence      float64
	}{
		// 1. Linear os.File leak
		"Pos_Res_LinearFileLeak": {StateResourceLeakConfirmed, true, 0.95},
		// 2. Deferred os.File.Close
		"Neg_Res_DeferredFileClose": {StateResourceDeferredRelease, false, 0.98},
		// 3. Explicit os.File.Close
		"Neg_Res_ExplicitFileClose": {StateResourceReleased, false, 0.95},
		// 4. Linear net.Conn leak
		"Pos_Res_LinearNetConnLeak": {StateResourceLeakConfirmed, true, 0.95},
		// 5. Deferred connection close
		"Neg_Res_DeferredNetConnClose": {StateResourceDeferredRelease, false, 0.98},
		// 6. HTTP response body leak
		"Pos_Res_HTTPBodyLeak": {StateResourceLeakConfirmed, true, 0.95},
		// 7. HTTP response body correctly closed
		"Neg_Res_HTTPBodyClosed": {StateResourceDeferredRelease, false, 0.98},
		// 8. SQL rows leak
		"Pos_Res_SQLRowsLeak": {StateResourceLeakConfirmed, true, 0.95},
		// 9. SQL rows deferred close
		"Neg_Res_SQLRowsDeferredClose": {StateResourceDeferredRelease, false, 0.98},
		// 10. Transaction with commit
		"Neg_Res_TxWithCommit": {StateResourceReleased, false, 0.95},
		// 11. Transaction with rollback
		"Neg_Res_TxWithRollback": {StateResourceReleased, false, 0.95},
		// 12. Transaction with deferred rollback
		"Neg_Res_TxWithDeferredRollback": {StateResourceDeferredRelease, false, 0.95},
		// 13. Conditional resource release
		"Pos_Res_ConditionalRelease": {StateResourceLeakProbable, true, 0.85},
		// 14. Early-return cleanup
		"Neg_Res_EarlyReturnCleanup": {StateResourceDeferredRelease, false, 0.98},
		// 15. Error-path cleanup
		"Neg_Res_ErrorPathCleanup": {StateResourceDeferredRelease, false, 0.98},
		// 16. Multiple resource aliases
		"Neg_Res_MultipleAliases": {StateResourceDeferredRelease, false, 0.95},
		// 17. Resource returned to caller
		"Neg_Res_ReturnedToCaller": {StateResourceTransferred, false, 0.95},
		// 18. Resource transferred through a supported ownership contract
		"Neg_Res_TransferredOwnership": {StateResourceTransferred, false, 0.95},
		// 19. Resource stored in a struct with unknown lifecycle
		"Pos_Res_StoredInStructUnknownLifecycle": {StateResourceUnresolved, false, 0.50},
		// 20. Timer stopped
		"Neg_Res_TimerStopped": {StateResourceReleased, false, 0.95},
		// 21. Timer ownership transferred
		"Neg_Res_TimerTransferred": {StateResourceTransferred, false, 0.95},
		// 22. Acquisition failure without resource creation
		"Neg_Res_AcquisitionFailurePath": {StateResourceDeferredRelease, false, 0.95},
		// 23. Resource acquired in a loop without closure
		"Pos_Res_AcquiredInLoop": {StateResourceLeakConfirmed, true, 0.95},
		// 24. Deferred release inside a loop
		"Pos_Res_DeferredInLoop": {StateResourceLeakProbable, true, 0.85},
	}

	allEvidence := make(map[string][]ResourceEvidence)

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		if err != nil {
			t.Errorf("AnalyzeFunction error on %s: %v", fnName, err)
		}
		allEvidence[fnName] = evs
	}

	// Verify all 24 test cases
	for fnName, expected := range expectedResults {
		evs, ok := allEvidence[fnName]
		if !ok || len(evs) == 0 {
			t.Errorf("fixture %s produced no resource evidence records", fnName)
			continue
		}

		matched := false
		for _, ev := range evs {
			if ev.LifecycleState == expected.ExpectedState {
				matched = true
				if ev.Confidence < expected.MinConfidence {
					t.Errorf("fixture %s: confidence %.2f lower than expected %.2f", fnName, ev.Confidence, expected.MinConfidence)
				}
				if expected.ShouldBeVulnerable {
					if ev.FindingState != StateConfirmed && ev.FindingState != StateProbable {
						t.Errorf("fixture %s: expected confirmed/probable finding state, got %v", fnName, ev.FindingState)
					}
				}
			}
		}

		if !matched {
			var got []ResourceLifecycleState
			for _, e := range evs {
				got = append(got, e.LifecycleState)
			}
			t.Errorf("fixture %s: expected lifecycle %s, got %v", fnName, expected.ExpectedState, got)
		}
	}
}

func TestResourceAnalyzerOnExternalHistoricalBenchmarks(t *testing.T) {
	resAnalyzer := NewResourceAnalyzer(".")

	// 1. External Repo: http_router (DEF-HTTP-01: UpstreamProbe HTTP body leak)
	httpRouterPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "http_router")
	saRouter := NewSemanticAnalyzer(httpRouterPath)
	routerModel, _, err := saRouter.AnalyzeFile("middleware.go")
	if err != nil {
		t.Fatalf("AnalyzeFile failed on http_router/middleware.go: %v", err)
	}

	foundDefHTTP01 := false
	for fnName, cfg := range routerModel.CFGs {
		df := routerModel.Dataflow[fnName]
		evs, _ := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, routerModel.FileSet, cfg, df, "middleware.go")
		for _, ev := range evs {
			if fnName == "UpstreamProbe" && ev.LifecycleState == StateResourceLeakConfirmed && ev.ResourceKind == ResourceKindHTTPBody {
				foundDefHTTP01 = true
			}
		}
	}
	if !foundDefHTTP01 {
		t.Errorf("FAILED to detect DEF-HTTP-01 in http_router/middleware.go (UpstreamProbe HTTP body leak)")
	}

	// 2. External Repo: sql_service (DEF-SQL-02: ListRecentUsers SQL rows leak)
	sqlPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")
	saSQL := NewSemanticAnalyzer(sqlPath)
	sqlModel, _, err := saSQL.AnalyzeFile("repository.go")
	if err != nil {
		t.Fatalf("AnalyzeFile failed on sql_service/repository.go: %v", err)
	}

	foundDefSQL02 := false
	for fnName, cfg := range sqlModel.CFGs {
		df := sqlModel.Dataflow[fnName]
		evs, _ := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, sqlModel.FileSet, cfg, df, "repository.go")
		for _, ev := range evs {
			if (fnName == "ListRecentUsers" || fnName == "UserRepository.ListRecentUsers") &&
				ev.LifecycleState == StateResourceLeakConfirmed && ev.ResourceKind == ResourceKindSQLRows {
				foundDefSQL02 = true
			}
		}
	}
	if !foundDefSQL02 {
		t.Errorf("FAILED to detect DEF-SQL-02 in sql_service/repository.go (ListRecentUsers SQL rows leak)")
	}

	// 3. External Repo: queue_worker (DEF-QUEUE-01: SendDirect net.Conn leak)
	queuePath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "queue_worker")
	saQueue := NewSemanticAnalyzer(queuePath)
	queueModel, _, err := saQueue.AnalyzeFile(filepath.Join("internal", "producer", "producer.go"))
	if err != nil {
		t.Fatalf("AnalyzeFile failed on queue_worker producer.go: %v", err)
	}

	foundDefQueue01 := false
	for fnName, cfg := range queueModel.CFGs {
		df := queueModel.Dataflow[fnName]
		evs, _ := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, queueModel.FileSet, cfg, df, "internal/producer/producer.go")
		for _, ev := range evs {
			if fnName == "SendDirect" && ev.LifecycleState == StateResourceLeakConfirmed && ev.ResourceKind == ResourceKindNetConn {
				foundDefQueue01 = true
			}
		}
	}
	if !foundDefQueue01 {
		t.Errorf("FAILED to detect DEF-QUEUE-01 in queue_worker producer.go (SendDirect net.Conn leak)")
	}
}

func TestExportResourceEvidenceReport(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "resource_fixtures.go")
	sa := NewSemanticAnalyzer(".")
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	resAnalyzer := NewResourceAnalyzer(".")
	var allFixtureEvs []ResourceEvidence
	tp, fp, tn, fn := 0, 0, 0, 0

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, _ := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		allFixtureEvs = append(allFixtureEvs, evs...)

		isPositiveFixture := len(fnName) >= 3 && fnName[:3] == "Pos"
		hasVulnerableEv := false
		for _, ev := range evs {
			if ev.LifecycleState == StateResourceLeakConfirmed || ev.LifecycleState == StateResourceLeakProbable {
				hasVulnerableEv = true
			}
		}

		if isPositiveFixture {
			// Pos_Res_StoredInStructUnknownLifecycle is an uncertain ownership test case (not a confirmed leak)
			if fnName == "Pos_Res_StoredInStructUnknownLifecycle" {
				continue
			}
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
		"report_type": "Milestone 3.4 Phase 3 Resource Lifecycle Model and Fixture Results",
		"timestamp":   time.Now().Format(time.RFC3339),
		"analyzer":    "semantic.ResourceAnalyzer",
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

	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_4_resource_fixture_results.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("write report error: %v", err)
	}
}
