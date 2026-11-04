package semantic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandAnalyzerOnFixtures(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "command_fixtures.go")
	sa := NewSemanticAnalyzer(".")

	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on fixtures: %v", err)
	}

	cmdAnalyzer := NewCommandAnalyzer(".")

	expectedResults := map[string]struct {
		ExpectedClassification CommandClassification
		ShouldBeVulnerable     bool
		MinConfidence          float64
	}{
		"Pos_Cmd_ShellSprintf":       {ClassificationUntrustedShellCommand, true, 0.95},
		"Pos_Cmd_ShellConcat":        {ClassificationUntrustedShellCommand, true, 0.95},
		"Pos_Cmd_ShellPropagated":    {ClassificationUntrustedShellCommand, true, 0.95},
		"Pos_Cmd_ShellParamDirect":   {ClassificationUntrustedShellCommand, true, 0.95},
		"Pos_Cmd_ContextShell":       {ClassificationUntrustedShellCommand, true, 0.95},
		"Neg_Cmd_StaticExec":         {ClassificationStaticCommand, false, 0.98},
		"Neg_Cmd_SafeArgSeparation":  {ClassificationSafeArgumentSeparation, false, 0.95},
		"Neg_Cmd_DirectArgBranch":    {ClassificationSafeArgumentSeparation, false, 0.95},
		"Neg_Cmd_StaticShell":        {ClassificationShellInterpretedCommand, false, 0.90},
	}

	allEvidence := make(map[string][]CommandEvidence)

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := cmdAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		if err != nil {
			t.Errorf("AnalyzeFunction error on %s: %v", fnName, err)
		}
		allEvidence[fnName] = evs
	}

	// Verify non-sink negative fixtures produce 0 command findings
	nonSinkFixtures := []string{"Neg_Cmd_SprintfNonExec", "Neg_Cmd_UnusedCommand"}
	for _, nsf := range nonSinkFixtures {
		evs := allEvidence[nsf]
		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedShellCommand {
				t.Errorf("fixture %s produced unexpected command injection finding: %+v", nsf, ev)
			}
		}
	}

	// Verify expected classifications on tested fixtures
	for fnName, expected := range expectedResults {
		evs, ok := allEvidence[fnName]
		if !ok || len(evs) == 0 {
			t.Errorf("fixture %s produced no command evidence records", fnName)
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
			var got []CommandClassification
			for _, e := range evs {
				got = append(got, e.Classification)
			}
			t.Errorf("fixture %s: expected classification %s, got %v", fnName, expected.ExpectedClassification, got)
		}
	}
}

func TestCommandAnalyzerOnExternalCLITooling(t *testing.T) {
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "cli_tooling")
	targetFile := "runner.go"

	sa := NewSemanticAnalyzer(repoPath)
	model, _, err := sa.AnalyzeFile(targetFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed on cli_tooling/runner.go: %v", err)
	}

	cmdAnalyzer := NewCommandAnalyzer(repoPath)
	foundDefCLI01 := false
	foundSafeDirect := false

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, err := cmdAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, targetFile)
		if err != nil {
			t.Fatalf("AnalyzeFunction failed on %s: %v", fnName, err)
		}

		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedShellCommand {
				if fnName == "ExecuteShellScript" {
					foundDefCLI01 = true
					if len(ev.UntrustedInputs) == 0 {
						t.Errorf("expected untrusted input trace in ExecuteShellScript")
					}
					if ev.Confidence < 0.95 {
						t.Errorf("expected confidence >= 0.95, got %.2f", ev.Confidence)
					}
				}
			}

			if ev.Classification == ClassificationSafeArgumentSeparation || ev.Classification == ClassificationDynamicArguments || ev.Classification == ClassificationStaticCommand {
				if fnName == "ExecuteDirect" {
					foundSafeDirect = true
				}
			}
		}
	}

	if !foundDefCLI01 {
		t.Errorf("FAILED to detect DEF-CLI-01 in cli_tooling/runner.go (ExecuteShellScript sh -c script injection)")
	}

	if !foundSafeDirect {
		t.Errorf("FAILED to certify safe argument separation in cli_tooling/runner.go (ExecuteDirect)")
	}
}

func TestExportCommandEvidenceReport(t *testing.T) {
	fixturePath := filepath.Join("fixtures", "command_fixtures.go")
	sa := NewSemanticAnalyzer(".")
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	cmdAnalyzer := NewCommandAnalyzer(".")
	var allFixtureEvs []CommandEvidence
	tp, fp, tn, fn := 0, 0, 0, 0

	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, _ := cmdAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		allFixtureEvs = append(allFixtureEvs, evs...)

		isPositiveFixture := len(fnName) >= 3 && fnName[:3] == "Pos"
		hasVulnerableEv := false
		for _, ev := range evs {
			if ev.Classification == ClassificationUntrustedShellCommand {
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

	// External evaluation on cli_tooling
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "cli_tooling")
	saExt := NewSemanticAnalyzer(repoPath)
	extModel, _, err := saExt.AnalyzeFile("runner.go")
	var extEvs []CommandEvidence
	if err == nil && extModel != nil {
		cmdExt := NewCommandAnalyzer(repoPath)
		for fnName, cfg := range extModel.CFGs {
			df := extModel.Dataflow[fnName]
			evs, _ := cmdExt.AnalyzeFunction(cfg.FuncDecl, extModel.FileSet, cfg, df, "runner.go")
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
		"report_type": "Milestone 3.4 Phase 2 Command Expression Evidence",
		"timestamp":   time.Now().Format(time.RFC3339),
		"analyzer":    "semantic.CommandAnalyzer",
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

	outPath := filepath.Join("..", "..", "..", "..", "..", "reports", "milestone3_4_command_expression_evidence.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("write report error: %v", err)
	}
}
