package agent

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdaptiveMutatorTransformsScannedFunctionDeclarations(t *testing.T) {
	source := `func Allow(a, b int) bool {
	if a > 0 {
		if b > 0 {
			return true
		}
	}
	return false
}`
	mutator := NewAdaptiveMutator(nil, 42)
	proposal, err := mutator.ProposeTargeted(context.Background(), MutationRequest{CurrentCode: source, FunctionName: "Allow"})
	if err != nil {
		t.Fatalf("could not mutate the full function returned by scanning: %v", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "candidate.go", "package sample\n"+proposal.MutatedCode, 0)
	if err != nil {
		t.Fatalf("mutation did not preserve a valid function declaration: %v", err)
	}
	function := file.Decls[0].(*ast.FuncDecl)
	condition := function.Body.List[0].(*ast.IfStmt).Cond.(*ast.BinaryExpr)
	if function.Name.Name != "Allow" || condition.Op != token.LAND {
		t.Fatalf("function signature or combined condition is incorrect: %s", proposal.MutatedCode)
	}
}

func TestAnnealerStopsWhenNoRewritesRemain(t *testing.T) {
	root, path := setupTestRepo(t)
	source, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), "return a + b")
	annealer := &Annealer{Mutator: NewAdaptiveMutator(nil, 42), Evaluator: &MockEvaluator{}, EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 40}}
	events := make(chan StepTelemetry, 100)
	result, err := annealer.Evolve(context.Background(), root, path, "Add", start, start+len("return a + b"), 0, AnnealingConfig{MaxSteps: 100}, events)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalSteps != 1 || result.ProposalsGenerated != 0 || result.TerminationReason != "no_applicable_mutations" || result.CorrectnessPassed != 0 {
		t.Fatalf("search did not stop with accurate counters: %+v", result)
	}
	event := <-events
	if event.EvaluationAttempted {
		t.Fatal("reported tests for a candidate that was never produced")
	}
}
