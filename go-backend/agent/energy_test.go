package agent

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── Test 1: Numerical Validity & Error Handling ───────────────────────────────

func TestEnergyMeasurement_NumericalValidation(t *testing.T) {
	eval := &MockEnergyEvaluator{}

	tests := []struct {
		name   string
		energy float64
	}{
		{"NaN energy", math.NaN()},
		{"Positive infinity", math.Inf(1)},
		{"Negative infinity", math.Inf(-1)},
		{"Zero energy", 0.0},
		{"Negative energy", -5.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval.FixedEnergy = tt.energy
			_, err := eval.EvaluateEnergy(context.Background(), []byte("package test"), "test.go", "Func")
			if !errors.Is(err, ErrInvalidEnergy) {
				t.Fatalf("expected ErrInvalidEnergy for %s (%.2f), got: %v", tt.name, tt.energy, err)
			}
		})
	}
}

// ── Test 2: Real Rust Engine Integration on Go Source ─────────────────────────

func TestRustEngineEvaluator_RealBinaryOnGoSource(t *testing.T) {
	enginePath := ResolveEnginePath()
	if fi, err := os.Stat(enginePath); err != nil || fi.IsDir() {
		t.Skipf("Skipping real Rust engine test: binary not found at '%s'", enginePath)
	}

	eval, err := NewRustEngineEvaluator(enginePath)
	if err != nil {
		t.Fatalf("NewRustEngineEvaluator failed: %v", err)
	}

	goSource := `package sample

func ComputeTotal(items []int) int {
	total := 0
	for _, x := range items {
		if x > 0 {
			total += x
		}
	}
	return total
}
`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	meas, err := eval.EvaluateEnergy(ctx, []byte(goSource), "sample.go", "ComputeTotal")
	if err != nil {
		t.Fatalf("EvaluateEnergy failed: %v", err)
	}

	if meas.FunctionName != "ComputeTotal" {
		t.Errorf("expected function name 'ComputeTotal', got '%s'", meas.FunctionName)
	}
	if meas.Energy <= 0.0 || math.IsNaN(meas.Energy) {
		t.Errorf("expected positive finite energy, got %.4f", meas.Energy)
	}
	if meas.Metrics.CyclomaticComplexity < 2 {
		t.Errorf("expected cyclomatic complexity >= 2, got %d", meas.Metrics.CyclomaticComplexity)
	}
	if meas.SourceHash == "" {
		t.Error("expected non-empty SourceHash")
	}
	t.Logf("Evaluated ComputeTotal energy: %.2f (cyclomatic=%d, nesting=%d, nodes=%d, duration=%v)",
		meas.Energy, meas.Metrics.CyclomaticComplexity, meas.Metrics.MaxNestingDepth, meas.Metrics.ASTNodeCount, meas.Duration)
}

// ── Test 3: Real Rust Engine on Python Source ──────────────────────────────────

func TestRustEngineEvaluator_RealBinaryOnPythonSource(t *testing.T) {
	enginePath := ResolveEnginePath()
	if fi, err := os.Stat(enginePath); err != nil || fi.IsDir() {
		t.Skipf("Skipping real Rust engine test: binary not found at '%s'", enginePath)
	}

	eval, err := NewRustEngineEvaluator(enginePath)
	if err != nil {
		t.Fatalf("NewRustEngineEvaluator failed: %v", err)
	}

	pySource := `def process_data(items):
    out = []
    for x in items:
        if x is not None:
            out.append(x * 2)
    return out
`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	meas, err := eval.EvaluateEnergy(ctx, []byte(pySource), "pipeline.py", "process_data")
	if err != nil {
		t.Fatalf("EvaluateEnergy on Python failed: %v", err)
	}

	if meas.FunctionName != "process_data" {
		t.Errorf("expected function name 'process_data', got '%s'", meas.FunctionName)
	}
	if meas.Energy <= 0.0 {
		t.Errorf("expected positive energy, got %.4f", meas.Energy)
	}
	t.Logf("Evaluated Python process_data energy: %.2f", meas.Energy)
}

// ── Test 4: Missing Function Reports ErrFunctionNotFound ──────────────────────

func TestRustEngineEvaluator_MissingFunctionError(t *testing.T) {
	enginePath := ResolveEnginePath()
	if fi, err := os.Stat(enginePath); err != nil || fi.IsDir() {
		t.Skipf("Skipping: binary not found at '%s'", enginePath)
	}

	eval, err := NewRustEngineEvaluator(enginePath)
	if err != nil {
		t.Fatal(err)
	}

	goSource := `package sample
func RealFunc() {}
`
	_, err = eval.EvaluateEnergy(context.Background(), []byte(goSource), "sample.go", "NonExistentFunc")
	if !errors.Is(err, ErrFunctionNotFound) {
		t.Fatalf("expected ErrFunctionNotFound, got: %v", err)
	}
}

// ── Test 5: Dynamic Baseline Evaluation ───────────────────────────────────────

func TestAnnealer_DynamicBaselineEvaluation(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	energyEval := &MockEnergyEvaluator{
		FixedEnergy: 45.5,
	}

	annealer := &Annealer{
		Mutator:         &mockMutator{},
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
	}

	cfg := AnnealingConfig{InitialTemp: 10.0, CoolingRate: 0.5, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)

	// Pass baseEnergy = 0.0 to trigger dynamic baseline calculation
	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 0.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if res.BaselineEnergy != 45.5 {
		t.Errorf("expected dynamically measured baseline energy 45.5, got %.2f", res.BaselineEnergy)
	}

	telem := <-ch
	if telem.BaselineEnergy != 45.5 {
		t.Errorf("expected telemetry baseline energy 45.5, got %.2f", telem.BaselineEnergy)
	}
}

// ── Test 6: Downhill Improvement Acceptance & Best Tracking ────────────────────

func TestAnnealer_DownhillImprovement(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	// Step 1: Energy drops from baseline 50.0 to 35.0 (Improvement)
	// Step 2: Energy drops from 35.0 to 25.0 (Further improvement)
	energyEval := &MockEnergyEvaluator{
		Energies: []float64{50.0 /* baseline */, 35.0 /* step 1 */, 25.0 /* step 2 */},
	}

	mutator := &mockMutator{
		mutations: []string{
			"// patch 1\n\treturn a + b",
			"// patch 2\n\treturn a + b",
		},
	}

	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
			return &EvaluationResult{
				TestsPassed:      true,
				Outcome:          OutcomePassed,
				UpdatedFileBytes: append(append([]byte{}, req.OriginalBytes[:req.StartByte]...), []byte(req.MutatedCode)...),
			}, nil
		},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: energyEval,
	}

	cfg := AnnealingConfig{InitialTemp: 100.0, CoolingRate: 0.85, MaxSteps: 2}
	ch := make(chan StepTelemetry, 5)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 50.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if !res.BestImproved {
		t.Error("expected BestImproved=true")
	}
	if res.FinalEnergy != 25.0 {
		t.Errorf("expected final energy 25.0, got %.2f", res.FinalEnergy)
	}
	if res.EnergyDelta != 25.0 { // 50 - 25 = 25 reduction
		t.Errorf("expected energy delta 25.0, got %.2f", res.EnergyDelta)
	}
	if res.AcceptedMoves != 2 {
		t.Errorf("expected 2 accepted moves, got %d", res.AcceptedMoves)
	}

	// Verify telemetry events
	telem1 := <-ch
	if !telem1.Accepted || telem1.Reason != "improvement" {
		t.Errorf("step 1 expected accepted improvement, got accepted=%v reason=%s", telem1.Accepted, telem1.Reason)
	}
	if telem1.DeltaE != -15.0 { // 35 - 50 = -15
		t.Errorf("step 1 expected deltaE=-15, got %.2f", telem1.DeltaE)
	}
	if !telem1.IsBestSoFar {
		t.Error("step 1 expected IsBestSoFar=true")
	}

	telem2 := <-ch
	if !telem2.Accepted || !telem2.IsBestSoFar {
		t.Error("step 2 expected accepted best move")
	}
	if telem2.DeltaE != -10.0 { // 25 - 35 = -10
		t.Errorf("step 2 expected deltaE=-10, got %.2f", telem2.DeltaE)
	}
}

// ── Test 7: Uphill Move Metropolis Acceptance & Best State Isolation ───────────

func TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	// Baseline: 30.0
	// Step 1: Energy drops to 20.0 (Best improved!)
	// Step 2: Energy increases to 25.0 (Uphill move: deltaE = +5.0)
	// At high temp T=100.0, p = exp(-5/100) = 0.951. With rng returning 0.1, it is stochastically accepted.
	energyEval := &MockEnergyEvaluator{
		Energies: []float64{30.0 /* baseline */, 20.0 /* step 1 */, 25.0 /* step 2 */},
	}

	mutator := &mockMutator{
		mutations: []string{
			"// best patch\n\treturn a + b",
			"// worse patch but explored\n\treturn a + b",
		},
	}

	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
			return &EvaluationResult{
				TestsPassed:      true,
				Outcome:          OutcomePassed,
				UpdatedFileBytes: append(append([]byte{}, req.OriginalBytes[:req.StartByte]...), []byte(req.MutatedCode)...),
			}, nil
		},
	}

	// Deterministic RNG that returns 0.1 for Float64
	deterministicRand := rand.New(rand.NewSource(1)) // seed 1 gives low float on first draw

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: energyEval,
		RandSource:      deterministicRand,
	}

	cfg := AnnealingConfig{InitialTemp: 100.0, CoolingRate: 0.99, MaxSteps: 2}
	ch := make(chan StepTelemetry, 5)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 30.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	_ = <-ch // discard step 1

	telem2 := <-ch
	if telem2.DeltaE <= 0.0 {
		t.Errorf("step 2 expected uphill move deltaE > 0, got %.2f", telem2.DeltaE)
	}

	// CRITICAL VERIFICATION:
	// Even though step 2 was explored, the final persisted result MUST be the best candidate (Step 1: 20.0), NOT Step 2 (25.0)!
	if res.FinalEnergy != 20.0 {
		t.Errorf("expected BestEnergy 20.0 to be preserved, got %.2f", res.FinalEnergy)
	}
	if res.OptimizedCode != "// best patch\n\treturn a + b" {
		t.Errorf("expected best patch to be returned, got %q", res.OptimizedCode)
	}

	// Verify disk has the best patch, not the worse exploratory patch
	savedBytes, _ := os.ReadFile(fullPath)
	if !strings.Contains(string(savedBytes), "// best patch") {
		t.Errorf("expected disk file to contain best patch, got:\n%s", string(savedBytes))
	}
}

// ── Test 8: Original File Preserved If No Candidate Improves ───────────────────

func TestAnnealer_OriginalFilePreservedIfNoCandidateImproves(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	// Baseline: 20.0. All candidate mutations are worse: 35.0, 40.0
	energyEval := &MockEnergyEvaluator{
		Energies: []float64{20.0 /* baseline */, 35.0, 40.0},
	}

	// Deterministic RNG that rejects uphill moves
	annealer := &Annealer{
		Mutator:         &mockMutator{mutations: []string{"patch 1", "patch 2"}},
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
		RandSource:      rand.New(rand.NewSource(9999)), // High numbers
	}

	cfg := AnnealingConfig{InitialTemp: 1.0, CoolingRate: 0.5, MaxSteps: 2}
	ch := make(chan StepTelemetry, 5)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 20.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if res.BestImproved {
		t.Error("expected BestImproved=false")
	}
	if res.EnergyDelta != 0.0 {
		t.Errorf("expected energy delta 0.0, got %.2f", res.EnergyDelta)
	}

	// Verify original file on host was completely untouched
	postBytes, _ := os.ReadFile(fullPath)
	if string(postBytes) != string(originalBytes) {
		t.Errorf("host file was modified despite no candidate improvement!")
	}
}

// ── Test 9: Energy Evaluation Failure Rejects Candidate Safely ────────────────

func TestAnnealer_EnergyEvaluationFailureRejectsCandidate(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	// Energy evaluator fails on candidate step
	callCount := 0
	energyEval := &MockEnergyEvaluator{
		EvaluateFunc: func(_ context.Context, _ []byte, _ string, _ string) (*EnergyMeasurement, error) {
			callCount++
			if callCount == 1 {
				// Baseline succeeds
				return &EnergyMeasurement{Energy: 30.0}, nil
			}
			return nil, ErrEngineFailed
		},
	}

	annealer := &Annealer{
		Mutator:         &mockMutator{mutations: []string{"mutated"}},
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
	}

	cfg := AnnealingConfig{InitialTemp: 10.0, CoolingRate: 0.5, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 30.0, cfg, ch)
	if err != nil {
		t.Fatalf("expected graceful step rejection, got: %v", err)
	}

	telem := <-ch
	if telem.Accepted {
		t.Error("expected candidate to be rejected when energy evaluation fails")
	}
	if !strings.Contains(telem.Reason, "energy_evaluation_failed") {
		t.Errorf("expected reason to contain 'energy_evaluation_failed', got '%s'", telem.Reason)
	}
	if res.BestImproved {
		t.Error("expected BestImproved=false")
	}
}

// ── Test 10: Path Matching Strictly Respects Separator Boundaries ──────────────

func TestExtractHotspotMetric_PathSeparatorBoundaries(t *testing.T) {
	report := &ThermodynamicReport{
		EngineVersion: "2.0.0",
		FileReports: []FileReport{
			{
				FilePath: "pkg/subcalc.go",
				Hotspots: []Hotspot{
					{
						FunctionName: "Compute",
						EntropyScore: 99.0,
					},
				},
			},
			{
				FilePath: "pkg/calc.go",
				Hotspots: []Hotspot{
					{
						FunctionName: "Compute",
						EntropyScore: 12.5,
					},
				},
			},
		},
	}

	// Seeking calc.go must NOT match subcalc.go even though subcalc ends with calc.go!
	meas, err := extractHotspotMetric(report, "calc.go", "Compute")
	if err != nil {
		t.Fatalf("extractHotspotMetric failed: %v", err)
	}

	if meas.Energy != 12.5 {
		t.Errorf("expected energy 12.5 from pkg/calc.go, got %.2f (incorrectly matched subcalc.go!)", meas.Energy)
	}
}

// ── Test 11: Non-Finite & Invalid Initial Temperature Handled Safely ──────────

func TestAnnealer_NonFiniteTemperatureSafety(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	energyEval := &MockEnergyEvaluator{
		Energies: []float64{30.0 /* baseline */, 25.0 /* step 1 */},
	}

	annealer := &Annealer{
		Mutator:         &mockMutator{mutations: []string{"patch"}},
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
	}

	// Passing NaN and +Inf temperatures and cooling rates
	cfg := AnnealingConfig{
		InitialTemp: math.NaN(),
		CoolingRate: math.Inf(1),
		MaxSteps:    1,
	}

	ch := make(chan StepTelemetry, 2)
	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 0.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve should handle non-finite temperature safely, got err: %v", err)
	}

	telem := <-ch
	if math.IsNaN(telem.Temp) || math.IsInf(telem.Temp, 0) {
		t.Errorf("telemetry temp must be finite and normalized, got %v", telem.Temp)
	}
	if !res.BestImproved {
		t.Error("expected candidate to be accepted and improve")
	}
}

// ── Test 12: Context Cancellation Unblocks Send Without Hanging ────────────────

func TestAnnealer_SendTelemetryDoesNotBlockOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	unbufferedCh := make(chan StepTelemetry) // Unbuffered, no reader

	cancel() // Cancel context before/during send

	telem := StepTelemetry{Step: 1, OldEnergy: 10.0}
	ok := sendTelemetry(ctx, unbufferedCh, telem)
	if ok {
		t.Error("expected sendTelemetry to return false when context is cancelled")
	}
}

// ── Test 13: Neutral Moves (DeltaE == 0) Explored But Not Claimed As Improvement

func TestAnnealer_EqualEnergyMoveExploration(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	originalBytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(originalBytes), "return a + b")
	endByte := startByte + len("return a + b")

	// Baseline is 20.0, candidate is identical 20.0
	energyEval := &MockEnergyEvaluator{
		Energies: []float64{20.0 /* baseline */, 20.0 /* step 1 */},
	}

	annealer := &Annealer{
		Mutator:         &mockMutator{mutations: []string{"// neutral\n\treturn a + b"}},
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
	}

	cfg := AnnealingConfig{InitialTemp: 50.0, CoolingRate: 0.85, MaxSteps: 1}
	ch := make(chan StepTelemetry, 2)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 0.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	telem := <-ch
	if !telem.Accepted {
		t.Error("neutral move (deltaE==0) should be accepted for exploratory plateau traversal")
	}
	if telem.IsBestSoFar {
		t.Error("neutral move should NOT be marked IsBestSoFar")
	}
	if res.BestImproved {
		t.Error("res.BestImproved must be false when no strict improvement occurred")
	}

	// File on disk must remain unchanged!
	diskBytes, _ := os.ReadFile(fullPath)
	if string(diskBytes) != string(originalBytes) {
		t.Errorf("original file was corrupted by equal-energy candidate")
	}
}
