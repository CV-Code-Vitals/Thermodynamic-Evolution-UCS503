package agent

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// BenchmarkRustEngineEvaluator measures raw evaluation latency and allocation
// overhead of the Rust AST engine on a real source file.
func BenchmarkRustEngineEvaluator(b *testing.B) {
	enginePath := ResolveEnginePath()
	if fi, err := os.Stat(enginePath); err != nil || fi.IsDir() {
		b.Skipf("Skipping benchmark: engine binary not found at %s", enginePath)
	}

	evaluator, err := NewRustEngineEvaluator(enginePath)
	if err != nil {
		b.Fatalf("NewRustEngineEvaluator failed: %v", err)
	}

	sampleGo := `package sample

func DeepCalculation(items []int, threshold int) int {
	total := 0
	for _, x := range items {
		if x > threshold {
			for j := 0; j < 3; j++ {
				if (x+j)%2 == 0 {
					total += (x * j)
				}
			}
		}
	}
	return total
}
`
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := evaluator.EvaluateEnergy(ctx, []byte(sampleGo), "sample.go", "DeepCalculation")
		if err != nil {
			b.Fatalf("EvaluateEnergy failed at iteration %d: %v", i, err)
		}
	}
}

// TestBenchmark_OptimizationEvidence performs a reproducible multi-step simulated
// annealing run using real Rust engine evaluations, documenting baseline, candidate,
// deltas, and duration telemetry as required by Phase 7.
func TestBenchmark_OptimizationEvidence(t *testing.T) {
	enginePath := ResolveEnginePath()
	if fi, err := os.Stat(enginePath); err != nil || fi.IsDir() {
		t.Skipf("Skipping benchmark: engine binary not found at %s", enginePath)
	}

	rustEval, err := NewRustEngineEvaluator(enginePath)
	if err != nil {
		t.Fatalf("Failed to initialize RustEngineEvaluator: %v", err)
	}

	tempDir := t.TempDir()
	relPath := "calc.go"
	fullPath := filepath.Join(tempDir, relPath)

	// Complex baseline code with nesting and branches
	originalSource := `package sample

func TargetLogic(a, b int) int {
	sum := 0
	for i := 0; i < a; i++ {
		if i > 0 {
			if (i % 2) == 0 {
				sum += i * b
			}
		}
	}
	return sum
}
`
	if err := os.WriteFile(fullPath, []byte(originalSource), 0o644); err != nil {
		t.Fatal(err)
	}

	startByte := strings.Index(originalSource, "sum := 0")
	endByte := strings.Index(originalSource, "return sum") + len("return sum")

	// Pre-planned progressive mutations:
	// Candidate 1: slightly flatter
	// Candidate 2: simplified direct formula (substantially lower AST complexity)
	mutations := []string{
		"sum := 0\n\tfor i := 2; i < a; i += 2 {\n\t\tsum += i * b\n\t}\n\treturn sum",
		"n := (a - 1) / 2\n\treturn b * n * (n + 1)",
	}

	mutator := &mockMutator{mutations: mutations}
	mockEval := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
			newBytes := append(append([]byte{}, req.OriginalBytes[:req.StartByte]...), []byte(req.MutatedCode)...)
			newBytes = append(newBytes, req.OriginalBytes[req.EndByte:]...)
			return &EvaluationResult{
				TestsPassed:      true,
				Outcome:          OutcomePassed,
				UpdatedFileBytes: newBytes,
			}, nil
		},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       mockEval,
		EnergyEvaluator: rustEval,
		RandSource:      rand.New(rand.NewSource(42)),
	}

	cfg := AnnealingConfig{
		InitialTemp: 50.0,
		CoolingRate: 0.85,
		MaxSteps:    2,
	}

	ch := make(chan StepTelemetry, 10)
	startTime := time.Now()

	result, err := annealer.Evolve(context.Background(), tempDir, relPath, "TargetLogic", startByte, endByte, 0.0, cfg, ch)
	totalDuration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Annealing run failed: %v", err)
	}

	t.Logf("═══════════════════════════════════════════════════════════════")
	t.Logf(" UCS503 Milestone 1: Optimization & Telemetry Evidence")
	t.Logf("═══════════════════════════════════════════════════════════════")
	t.Logf(" Baseline Energy       : %.4f", result.BaselineEnergy)
	t.Logf(" Final Best Energy     : %.4f", result.FinalEnergy)
	t.Logf(" Energy Delta (ΔE)     : %.4f (Lower is better)", result.EnergyDelta)
	t.Logf(" Best Improved         : %v", result.BestImproved)
	t.Logf(" Total Annealing Steps : %d", result.TotalSteps)
	t.Logf(" Accepted Moves        : %d", result.AcceptedMoves)
	t.Logf(" Total Execution Time  : %v", totalDuration)
	t.Logf("───────────────────────────────────────────────────────────────")

	stepNum := 1
	for len(ch) > 0 {
		telem := <-ch
		t.Logf(" Step %d | Temp=%.2f | CandidateE=%.2f | ΔE=%.2f | Accepted=%v (%s) | BestSoFar=%v",
			stepNum, telem.Temp, telem.CandidateEnergy, telem.DeltaE, telem.Accepted, telem.Reason, telem.IsBestSoFar)
		stepNum++
	}
	t.Logf("═══════════════════════════════════════════════════════════════")

	if !result.BestImproved {
		t.Error("expected candidate to improve baseline energy")
	}
	if result.FinalEnergy >= result.BaselineEnergy {
		t.Errorf("expected final energy (%.2f) < baseline energy (%.2f)", result.FinalEnergy, result.BaselineEnergy)
	}
}
