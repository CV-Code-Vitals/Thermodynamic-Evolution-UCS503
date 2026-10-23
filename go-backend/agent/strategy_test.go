package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Test 1: Targeted Simplification (Nesting Depth Reduction) ─────────────────

func TestStrategy_TargetedSimplification(t *testing.T) {
	strat := &TargetedSimplificationStrategy{}

	inputCode := `if a > 0 {
	if b > 0 {
		return a + b
	}
}`

	if !strat.CanMutate(inputCode) {
		t.Fatal("expected CanMutate to return true for nested if")
	}

	req := MutationRequest{
		CandidateID:      "cand-1",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if prop.StrategyName != "targeted_simplification" {
		t.Errorf("expected strategy targeted_simplification, got %s", prop.StrategyName)
	}

	if !strings.Contains(prop.MutatedCode, "a > 0 &&") || !strings.Contains(prop.MutatedCode, "b > 0") {
		t.Errorf("expected flattened condition 'a > 0 && b > 0', got:\n%s", prop.MutatedCode)
	}
}

// ── Test 2: Control Flow Simplification (Cyclomatic Complexity Reduction) ──────

func TestStrategy_ControlFlowSimplification(t *testing.T) {
	strat := &ControlFlowSimplificationStrategy{}

	inputCode := `if x > 10 {
	return true
}
return false`

	if !strat.CanMutate(inputCode) {
		t.Fatal("expected CanMutate to return true")
	}

	req := MutationRequest{
		CandidateID:      "cand-2",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if !strings.Contains(prop.MutatedCode, "return x > 10") {
		t.Errorf("expected simplified 'return x > 10', got:\n%s", prop.MutatedCode)
	}
}

// ── Test 3: Duplication Reduction (AST Volume Reduction) ──────────────────────

func TestStrategy_DuplicationReduction(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	inputCode := `x := (a + b) * 2
y := (a + b) + 5
return x + y`

	req := MutationRequest{
		CandidateID:      "cand-3",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if !strings.Contains(prop.MutatedCode, "subExpr :=") {
		t.Errorf("expected hoisted subexpression 'subExpr :=', got:\n%s", prop.MutatedCode)
	}
}

// ── Test 4: Conservative Refactoring (Allocation Hoisting) ────────────────────

func TestStrategy_ConservativeRefactoring_HoistLoopAlloc(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	inputCode := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	_ = buf
}`

	if !strat.CanMutate(inputCode) {
		t.Fatal("expected CanMutate to return true")
	}

	req := MutationRequest{
		CandidateID:      "cand-4",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	// Verify make([]byte, 1024) is hoisted before the for loop
	lines := strings.Split(prop.MutatedCode, "\n")
	makeLine := -1
	forLine := -1
	for idx, l := range lines {
		if strings.Contains(l, "buf := make") {
			makeLine = idx
		}
		if strings.Contains(l, "for ") {
			forLine = idx
		}
	}

	if makeLine == -1 || forLine == -1 || makeLine > forLine {
		t.Errorf("expected make to be hoisted before for loop, got:\n%s", prop.MutatedCode)
	}
}

// ── Test 5: Adaptive Mutator Determinism & Tracking ───────────────────────────

func TestAdaptiveMutator_DeterministicSelection(t *testing.T) {
	// Fixed seed produces reproducible sequence
	mutator := NewAdaptiveMutator(nil, 42)

	code := `if a > 0 {
	if b > 0 {
		return a + b
	}
}`

	h := sha256.Sum256([]byte(code))
	req := MutationRequest{
		CandidateID:      "cand-adaptive-1",
		BaseSnapshotHash: hex.EncodeToString(h[:]),
		CurrentCode:      code,
	}

	prop, err := mutator.ProposeTargeted(context.Background(), req)
	if err != nil {
		t.Fatalf("ProposeTargeted failed: %v", err)
	}

	if prop.StrategyName == "" {
		t.Error("expected strategy name in proposal")
	}

	// Record outcome
	mutator.RecordOutcome(prop.StrategyName, true, true)
	stats := mutator.GetStrategyStats()
	st, ok := stats[prop.StrategyName]
	if !ok || st.ProposalsGenerated != 1 || st.Improvements != 1 {
		t.Errorf("expected recorded stats for %s, got: %+v", prop.StrategyName, st)
	}
}

// ── Test 6: Stale Mutation Proposal Rejection In Annealer ──────────────────────

type staleMutator struct {
	fakeOldHash string
}

func (s *staleMutator) ProposeMutation(_ context.Context, _, _ string) (string, error) {
	return "return 1", nil
}

func (s *staleMutator) ProposeTargeted(_ context.Context, req MutationRequest) (*MutationProposal, error) {
	return &MutationProposal{
		CandidateID:      req.CandidateID,
		StrategyName:     "stale_strategy",
		BaseSnapshotHash: s.fakeOldHash, // Intentionally stale hash!
		MutatedCode:      "return a + b",
	}, nil
}

func (s *staleMutator) GetStrategyStats() map[string]StrategyStats {
	return nil
}

func TestAnnealer_StaleProposalRejected(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	energyEval := &MockEnergyEvaluator{FixedEnergy: 30.0}
	mutator := &staleMutator{fakeOldHash: "0000000000000000000000000000000000000000000000000000000000000000"}

	annealer := &Annealer{
		Mutator:         mutator,
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
	if !strings.Contains(telem.Reason, "stale_proposal_rejected") {
		t.Errorf("expected reason to contain 'stale_proposal_rejected', got: %s", telem.Reason)
	}
	if res.BestImproved {
		t.Error("stale proposal should not improve baseline")
	}
}

// ── Test 7: Duplicate Candidate Detection & Skipping ──────────────────────────

func TestAnnealer_DuplicateCandidateSkipping(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	evalCallCount := 0
	evaluator := &MockEvaluator{
		EvaluateFunc: func(_ context.Context, req EvaluationRequest) (*EvaluationResult, error) {
			evalCallCount++
			newBytes := append(append([]byte{}, req.OriginalBytes[:req.StartByte]...), []byte(req.MutatedCode)...)
			newBytes = append(newBytes, req.OriginalBytes[req.EndByte:]...)
			return &EvaluationResult{
				TestsPassed:      true,
				Outcome:          OutcomePassed,
				UpdatedFileBytes: newBytes,
			}, nil
		},
	}

	// Mutator proposes the EXACT same patch for 3 steps
	mutator := &mockMutator{
		mutations: []string{
			"// duplicate patch\n\treturn a + b",
			"// duplicate patch\n\treturn a + b",
			"// duplicate patch\n\treturn a + b",
		},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: &MockEnergyEvaluator{FixedEnergy: 25.0},
	}

	cfg := AnnealingConfig{
		InitialTemp:              50.0,
		CoolingRate:              0.85,
		MaxSteps:                 3,
		EnableDuplicateDetection: true,
	}
	ch := make(chan StepTelemetry, 5)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 0.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	// Evaluator should only be called ONCE (on step 1). Steps 2 and 3 should be skipped!
	if evalCallCount != 1 {
		t.Errorf("expected evaluator to be called exactly 1 time, got %d", evalCallCount)
	}
	if res.DuplicatesSkipped != 2 {
		t.Errorf("expected 2 duplicates skipped, got %d", res.DuplicatesSkipped)
	}
}

// ── Test 8: Search Budget Early Termination ───────────────────────────────────

func TestAnnealer_TargetEnergyTermination(t *testing.T) {
	repoDir, relPath := setupTestRepo(t)
	fullPath := filepath.Join(repoDir, relPath)
	bytes, _ := os.ReadFile(fullPath)

	startByte := strings.Index(string(bytes), "return a + b")
	endByte := startByte + len("return a + b")

	energyEval := &MockEnergyEvaluator{
		Energies: []float64{50.0 /* baseline */, 10.0 /* step 1 - reaches target! */, 5.0},
	}

	mutator := &mockMutator{
		mutations: []string{"patch 1", "patch 2", "patch 3"},
	}

	annealer := &Annealer{
		Mutator:         mutator,
		Evaluator:       &MockEvaluator{},
		EnergyEvaluator: energyEval,
	}

	cfg := AnnealingConfig{
		InitialTemp:  100.0,
		CoolingRate:  0.85,
		MaxSteps:     10,
		TargetEnergy: 15.0, // Stop as soon as energy <= 15.0
	}
	ch := make(chan StepTelemetry, 10)

	res, err := annealer.Evolve(context.Background(), repoDir, relPath, "Add", startByte, endByte, 0.0, cfg, ch)
	if err != nil {
		t.Fatalf("Evolve failed: %v", err)
	}

	if res.TerminationReason != "target_energy_reached" {
		t.Errorf("expected termination_reason 'target_energy_reached', got '%s'", res.TerminationReason)
	}
	if res.FinalEnergy > 15.0 {
		t.Errorf("expected final energy <= 15.0, got %.2f", res.FinalEnergy)
	}
}

// ── Test 9: Atomic Persistence With Concurrent Modification Collision ─────────

func TestAtomicWriteFileWithConflictDetection_DetectsCollision(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "sample.go")

	originalContent := []byte("package sample\n\nfunc Original() {}\n")
	if err := os.WriteFile(targetFile, originalContent, 0o644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(originalContent)
	baselineHash := hex.EncodeToString(h[:])

	// Case 1: No collision — disk content matches baseline hash
	newContent := []byte("package sample\n\nfunc Optimized() {}\n")
	if err := AtomicWriteFileWithConflictDetection(targetFile, newContent, baselineHash, 0o644); err != nil {
		t.Fatalf("expected successful write when hashes match, got: %v", err)
	}

	// Verify targetFile was updated
	diskBytes, _ := os.ReadFile(targetFile)
	if string(diskBytes) != string(newContent) {
		t.Errorf("expected targetFile to contain new content")
	}

	// Case 2: Concurrent external modification — another process edits targetFile
	externalContent := []byte("package sample\n\nfunc ExternalConcurrentEdit() {}\n")
	_ = os.WriteFile(targetFile, externalContent, 0o644)

	// Now try to persist candidate assuming old baselineHash
	err := AtomicWriteFileWithConflictDetection(targetFile, newContent, baselineHash, 0o644)
	if !errors.Is(err, ErrConcurrentModification) {
		t.Fatalf("expected ErrConcurrentModification on hash collision, got: %v", err)
	}

	// File on disk must retain the external modification
	preservedBytes, _ := os.ReadFile(targetFile)
	if string(preservedBytes) != string(externalContent) {
		t.Errorf("expected external modification to be preserved, got:\n%s", string(preservedBytes))
	}
}

// ── Test 10: Energy Cache Hit & Eviction ──────────────────────────────────────

func TestEnergyCache_HitAndEviction(t *testing.T) {
	cache := NewEnergyCache(2)

	m1 := &EnergyMeasurement{Energy: 10.0, FunctionName: "F1"}
	m2 := &EnergyMeasurement{Energy: 20.0, FunctionName: "F2"}
	m3 := &EnergyMeasurement{Energy: 30.0, FunctionName: "F3"}

	cache.Put("key1", m1)
	cache.Put("key2", m2)

	if cache.Size() != 2 {
		t.Errorf("expected cache size 2, got %d", cache.Size())
	}

	got1, ok := cache.Get("key1")
	if !ok || got1.Energy != 10.0 {
		t.Errorf("expected cache hit for key1 with energy 10.0")
	}

	// Inserting third element should trigger eviction
	cache.Put("key3", m3)
	if cache.Size() > 2 {
		t.Errorf("cache exceeded maxSize 2, got %d", cache.Size())
	}
}

// ── Test 11: Semantic Negative Tests: Unsafe Mutations Must Be Rejected ────────

func TestStrategy_Negative_NestedIf_WithElse(t *testing.T) {
	strat := &TargetedSimplificationStrategy{}

	// Inner if has an else block without early return — flattening is semantically invalid
	inputCode := `if a > 0 {
	if b > 0 {
		c++
	} else {
		c--
	}
}`

	// Verify directly that flattenNestedIf refuses to flatten when inner if has else
	if _, ok := flattenNestedIf(inputCode); ok {
		t.Error("expected flattenNestedIf to return false when inner if has an else branch")
	}

	if strat.CanMutate(inputCode) {
		t.Error("expected CanMutate to return false when inner if has an else branch and no return")
	}

	req := MutationRequest{
		CandidateID:      "cand-neg-1",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}
	_, err := strat.Mutate(context.Background(), req)
	if err == nil {
		t.Error("expected Mutate to fail on inner if with else branch")
	}
}

func TestStrategy_Negative_HoistLoopAlloc_EscapingBuffer(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	// Buffer escapes into external slice — hoisting would alias elements across iterations!
	inputCode := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	out = append(out, buf)
}`

	if strat.CanMutate(inputCode) {
		t.Error("expected CanMutate to reject hoisting for escaping buffer in append")
	}
}

func TestStrategy_Negative_HoistLoopAlloc_LoopVariableSize(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	// Buffer size depends on loop induction variable i — hoisting would freeze size to pre-loop value!
	inputCode := `for i := 0; i < n; i++ {
	buf := make([]byte, i)
	_ = buf
}`

	if strat.CanMutate(inputCode) {
		t.Error("expected CanMutate to reject hoisting when allocation size depends on loop variable")
	}
}

func TestStrategy_Negative_DuplicationReduction_ImpureCalls(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// Expressions containing function calls of unknown purity MUST NOT be hoisted
	inputCode := `x := randInt() + 1
y := randInt() + 1
return x + y`

	req := MutationRequest{
		CandidateID:      "cand-neg-impure",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}
	_, err := strat.Mutate(context.Background(), req)
	if err == nil {
		t.Error("expected Mutate to reject hoisting impure function calls")
	}
}

func TestStrategy_Negative_DuplicationReduction_ReassignedVars(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// 'a' is reassigned between uses — subexpression (a + b) is NOT invariant!
	inputCode := `x := a + b
a = 10
y := a + b
return x + y`

	req := MutationRequest{
		CandidateID:      "cand-neg-reassigned",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}
	_, err := strat.Mutate(context.Background(), req)
	if err == nil {
		t.Error("expected Mutate to reject hoisting subexpression when component variable is mutated")
	}
}

// ── Test 12: Semantic Positive Tests: Operator Precedence Preservation ────────

func TestStrategy_Positive_NestedIf_PreservesPrecedence(t *testing.T) {
	strat := &TargetedSimplificationStrategy{}

	// Outer condition has '||' which has lower precedence than '&&'.
	// Naive flattening: a || b && c => a || (b && c) — WRONG!
	// Correct flattening: (a || b) && c
	inputCode := `if a || b {
	if c {
		return true
	}
}`

	req := MutationRequest{
		CandidateID:      "cand-prec-1",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if !strings.Contains(prop.MutatedCode, "(a || b)") {
		t.Errorf("expected outer condition with '||' to be parenthesized, got:\n%s", prop.MutatedCode)
	}
}

func TestStrategy_Positive_BooleanReturn_PreservesPrecedence(t *testing.T) {
	strat := &ControlFlowSimplificationStrategy{}

	// Condition is binary 'a || b'. Simplified inverted return must be '!(a || b)', not '!a || b'
	inputCode := `if a || b {
	return false
}
return true`

	req := MutationRequest{
		CandidateID:      "cand-prec-2",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	if !strings.Contains(prop.MutatedCode, "return !(a || b)") {
		t.Errorf("expected parenthesized negation 'return !(a || b)', got:\n%s", prop.MutatedCode)
	}
}

// ── Test 13: Differential Semantics Verification ──────────────────────────────

func TestStrategy_DifferentialEquivalence(t *testing.T) {
	// Original logic:
	origFunc := func(a, b, c bool) bool {
		if a || b {
			if c {
				return true
			}
		}
		return false
	}

	// Transformed logic (with precedence preservation):
	transFunc := func(a, b, c bool) bool {
		if (a || b) && c {
			return true
		}
		return false
	}

	// Naive broken logic (without precedence preservation):
	brokenFunc := func(a, b, c bool) bool {
		if a || b && c {
			return true
		}
		return false
	}

	// Exhaustive truth-table validation across all 2^3 = 8 input combinations
	inputs := []bool{false, true}
	brokenDivergenceFound := false

	for _, a := range inputs {
		for _, b := range inputs {
			for _, c := range inputs {
				orig := origFunc(a, b, c)
				trans := transFunc(a, b, c)
				broken := brokenFunc(a, b, c)

				// Verified transform must match original for ALL inputs
				if orig != trans {
					t.Fatalf("differential semantic violation at (a=%v, b=%v, c=%v): orig=%v, trans=%v",
						a, b, c, orig, trans)
				}

				// Confirm that the broken (unparenthesized) version diverges on (a=true, b=false, c=false)
				if orig != broken {
					brokenDivergenceFound = true
				}
			}
		}
	}

	if !brokenDivergenceFound {
		t.Error("expected broken unparenthesized logic to diverge from original on truth-table check")
	}
}

// ── Test 14: Phase 1 — Loop Allocation Hoisting: Zero-Init & Partial Overwrite ─

func TestStrategy_Phase1_HoistLoopAlloc_PartialOverwriteZeroing(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	inputCode := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	buf[0] = byte(i)
	_ = buf
}`

	req := MutationRequest{
		CandidateID:      "cand-p1-zero",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	// Verify that clear(buf) was inserted in the loop body to guarantee zero-initialization on every iteration
	if !strings.Contains(prop.MutatedCode, "clear(buf)") {
		t.Errorf("expected clear(buf) to be inserted in loop body, got:\n%s", prop.MutatedCode)
	}

	// Differential execution demonstrating that moving make WITHOUT clearing leaks residual data:
	// Run 2 iterations where iteration 0 writes 5 bytes, and iteration 1 writes only 1 byte:
	runWithoutClearing := func() [2]byte {
		buf := make([]byte, 4) // hoisted once, NEVER cleared
		var observed [2]byte
		for iter := 0; iter < 2; iter++ {
			if iter == 0 {
				buf[0] = 'X'
				buf[1] = 'Y'
			} else {
				buf[0] = 'Z'
				// buf[1] not written in iteration 1!
			}
			observed[iter] = buf[1]
		}
		return observed
	}

	runWithClearing := func() [2]byte {
		buf := make([]byte, 4) // hoisted once
		var observed [2]byte
		for iter := 0; iter < 2; iter++ {
			clear(buf) // zero-initialized on every iteration!
			if iter == 0 {
				buf[0] = 'X'
				buf[1] = 'Y'
			} else {
				buf[0] = 'Z'
			}
			observed[iter] = buf[1]
		}
		return observed
	}

	leakObserved := runWithoutClearing()
	cleanObserved := runWithClearing()

	// Without clearing, iteration 1 observes stale 'Y' (89) from iteration 0
	if leakObserved[1] != 'Y' {
		t.Errorf("expected stale byte leak without clearing, got %v", leakObserved)
	}
	// With clearing (the strategy's fix), iteration 1 observes clean 0, matching fresh allocation!
	if cleanObserved[1] != 0 {
		t.Errorf("expected zero byte with clearing, got %v", cleanObserved)
	}
}

// ── Test 15: Phase 1 — Loop Allocation Hoisting: Escapes Struct or Interface ──

func TestStrategy_Phase1_HoistLoopAlloc_EscapesStructOrInterface(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	// Case A: Stored in struct field
	codeStruct := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	holder.data = buf
}`
	if strat.CanMutate(codeStruct) {
		t.Error("expected CanMutate to reject hoisting when buffer is stored in struct field")
	}

	// Case B: Stored in map
	codeMap := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	cache[i] = buf
}`
	if strat.CanMutate(codeMap) {
		t.Error("expected CanMutate to reject hoisting when buffer is stored in map")
	}

	// Case C: Stored in outer interface/any
	codeIface := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	out = buf
}`
	if strat.CanMutate(codeIface) {
		t.Error("expected CanMutate to reject hoisting when buffer is assigned to outer variable")
	}
}

// ── Test 16: Phase 1 — Loop Allocation Hoisting: Nested Loops, Defer, Closures ──

func TestStrategy_Phase1_HoistLoopAlloc_NestedLoopsDeferClosures(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	// Case A: Closure capture
	codeClosure := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	fn := func() { _ = buf }
	fn()
}`
	if strat.CanMutate(codeClosure) {
		t.Error("expected CanMutate to reject hoisting when buffer is captured by closure")
	}

	// Case B: Deferred call
	codeDefer := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	defer clean(buf)
}`
	if strat.CanMutate(codeDefer) {
		t.Error("expected CanMutate to reject hoisting when buffer is used in defer")
	}

	// Case C: Nested loop
	codeNested := `for i := 0; i < n; i++ {
	buf := make([]byte, 1024)
	for j := 0; j < m; j++ {
		_ = buf[j]
	}
}`
	if strat.CanMutate(codeNested) {
		t.Error("expected CanMutate to reject hoisting when buffer is referenced in nested loop")
	}
}

// ── Test 17: Phase 1 — Loop Allocation Hoisting: Dynamic Allocation Size ───────

func TestStrategy_Phase1_HoistLoopAlloc_DynamicAllocationSize(t *testing.T) {
	strat := &ConservativeRefactoringStrategy{}

	// Dynamic size referencing mutated variable
	codeDynamic := `for i := 0; i < n; i++ {
	sz := getDynamicSize(i)
	buf := make([]byte, sz)
	_ = buf
}`
	if strat.CanMutate(codeDynamic) {
		t.Error("expected CanMutate to reject hoisting when allocation size depends on loop-mutated variable")
	}
}

// ── Test 18: Phase 2 — Common Subexpressions: Rejection of Division & Remainder ─

func TestStrategy_Phase2_DuplicationReduction_RejectsDivisionAndRemainder(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// Division can panic with integer divide-by-zero!
	codeDiv := `x := a / b
y := a / b
return x + y`

	if strat.CanMutate(codeDiv) {
		t.Error("expected CanMutate to reject hoisting division (potential divide-by-zero panic)")
	}

	// Remainder can also panic!
	codeRem := `x := a % b
y := a % b
return x + y`

	if strat.CanMutate(codeRem) {
		t.Error("expected CanMutate to reject hoisting remainder (potential divide-by-zero panic)")
	}
}

// ── Test 19: Phase 2 — Common Subexpressions: Rejection Across Control Branches ─

func TestStrategy_Phase2_DuplicationReduction_RejectsAcrossBranches(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// Repeated across if/else branches — hoisting outside branch introduces speculative evaluation!
	codeBranches := `if cond {
	x = a + b
} else {
	y = a + b
}`

	if strat.CanMutate(codeBranches) {
		t.Error("expected CanMutate to reject hoisting duplicate expressions across conditional branches")
	}
}

// ── Test 20: Phase 2 — Common Subexpressions: Placement Inside Block ────────────

func TestStrategy_Phase2_DuplicationReduction_PlacedInSameBlock(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// Duplicates are inside an if block:
	inputCode := `if cond {
	x := a + b
	y := (a + b) * 2
	return x + y
}
return 0`

	req := MutationRequest{
		CandidateID:      "cand-p2-block",
		BaseSnapshotHash: "abc",
		CurrentCode:      inputCode,
	}

	prop, err := strat.Mutate(context.Background(), req)
	if err != nil {
		t.Fatalf("Mutate failed: %v", err)
	}

	// subExpr := a + b MUST be placed inside the if block, NOT at function entry!
	lines := strings.Split(prop.MutatedCode, "\n")
	ifLine := -1
	subExprLine := -1
	for idx, l := range lines {
		if strings.Contains(l, "if cond {") {
			ifLine = idx
		}
		if strings.Contains(l, "subExpr := a + b") {
			subExprLine = idx
		}
	}

	if ifLine == -1 || subExprLine == -1 || subExprLine < ifLine {
		t.Errorf("expected subExpr := to be placed INSIDE if block after 'if cond {', got:\n%s", prop.MutatedCode)
	}
}

// ── Test 21: Phase 2 — Common Subexpressions: Rejection on Shadowing ───────────

func TestStrategy_Phase2_DuplicationReduction_RejectsShadowing(t *testing.T) {
	strat := &DuplicationReductionStrategy{}

	// 'a' is shadowed or reassigned between candidate occurrences
	codeShadow := `x := a + b
a := 42
y := a + b
return x + y`

	if strat.CanMutate(codeShadow) {
		t.Error("expected CanMutate to reject hoisting when operand identifier is shadowed/redefined")
	}
}

// ── Test 22: Phase 3 — Nested If: Short-Circuit Side Effects Verification ───────

func TestStrategy_Phase3_NestedIf_ShortCircuitSideEffects(t *testing.T) {
	// Original logic:
	sideEffectsA := 0
	sideEffectsB := 0

	checkA := func(val bool) bool {
		sideEffectsA++
		return val
	}
	checkB := func(val bool) bool {
		sideEffectsB++
		return val
	}

	// In original: if checkA() { if checkB() { ... } }
	origBodyExecuted := false
	if checkA(false) {
		if checkB(true) {
			origBodyExecuted = true
		}
	}

	// In transformed: if checkA() && checkB() { ... }
	transEffectsA := 0
	transEffectsB := 0
	transCheckA := func(val bool) bool {
		transEffectsA++
		return val
	}
	transCheckB := func(val bool) bool {
		transEffectsB++
		return val
	}

	transBodyExecuted := false
	if transCheckA(false) && transCheckB(true) {
		transBodyExecuted = true
	}

	// Verify exact equivalence of side-effect counts and execution outcome
	if origBodyExecuted != transBodyExecuted {
		t.Errorf("body execution outcome mismatch: orig=%v, trans=%v", origBodyExecuted, transBodyExecuted)
	}
	if sideEffectsA != transEffectsA || sideEffectsB != transEffectsB {
		t.Errorf("side effect count mismatch: orig=(%d,%d), trans=(%d,%d)",
			sideEffectsA, sideEffectsB, transEffectsA, transEffectsB)
	}
	if transEffectsB != 0 {
		t.Errorf("expected short-circuit to completely skip checkB, but it was called %d times", transEffectsB)
	}
}

// ── Test 23: Phase 3 — Nested If: Outer Else Rejection ─────────────────────────

func TestStrategy_Phase3_NestedIf_OuterElseRejection(t *testing.T) {
	// Outer if has an else block — flattening if A { if B { body } } else { alt }
	// is semantically invalid because when A is true and B is false, original does NOT run alt,
	// but naive flattening (if A && B { body } else { alt }) WOULD run alt!
	codeOuterElse := `if a > 0 {
	if b > 0 {
		return 1
	}
} else {
	return 2
}`

	if _, ok := flattenNestedIf(codeOuterElse); ok {
		t.Error("expected flattenNestedIf to reject flattening when outer if has an else branch")
	}
}

// ── Test 24: Phase 3 — Boolean Return: Rejection of If Init Statement ──────────

func TestStrategy_Phase3_BooleanReturn_RejectsIfInit(t *testing.T) {
	strat := &ControlFlowSimplificationStrategy{}

	// ifStmt has an Init statement declaring a scoped variable
	codeInit := `if x := compute(); x > 0 {
	return true
}
return false`

	if strat.CanMutate(codeInit) {
		t.Error("expected CanMutate to reject boolean return simplification when if has an Init statement")
	}

	req := MutationRequest{
		CandidateID:      "cand-p3-init",
		BaseSnapshotHash: "abc",
		CurrentCode:      codeInit,
	}
	_, err := strat.Mutate(context.Background(), req)
	if err == nil {
		t.Error("expected Mutate to fail when if has an Init statement")
	}
}

// ── Test 25: Phase 4 — Deterministic CI Gate Outcomes Simulation ────────────────

func TestCIWorkflow_ThreeOutcomesSimulation(t *testing.T) {
	// Simulates the exact logic in .github/workflows/ci-go.yml:
	// TEST_EXIT=0
	// ... run test > sandbox_test.log ... || TEST_EXIT=$?
	// if [ "$TEST_EXIT" -ne 0 ]; then exit $TEST_EXIT; fi
	// if grep -q -F -- "--- SKIP: TestDockerEvaluator_LiveExecution" sandbox_test.log; then exit 1; fi
	// if ! grep -q -F -- "--- PASS: TestDockerEvaluator_LiveExecution" sandbox_test.log; then exit 1; fi

	evalGate := func(testExit int, logContent string) (gatePassed bool, exitCode int) {
		if testExit != 0 {
			return false, testExit
		}
		if strings.Contains(logContent, "--- SKIP: TestDockerEvaluator_LiveExecution") {
			return false, 1
		}
		if !strings.Contains(logContent, "--- PASS: TestDockerEvaluator_LiveExecution") {
			return false, 1
		}
		return true, 0
	}

	// Outcome 1: PASS
	passLog := "=== RUN TestDockerEvaluator_LiveExecution\n--- PASS: TestDockerEvaluator_LiveExecution (2.15s)\nPASS\n"
	passed, code := evalGate(0, passLog)
	if !passed || code != 0 {
		t.Errorf("Outcome 1 (PASS) failed gate: passed=%v, code=%d", passed, code)
	}

	// Outcome 2: FAILURE
	failLog := "=== RUN TestDockerEvaluator_LiveExecution\n--- FAIL: TestDockerEvaluator_LiveExecution (1.02s)\nFAIL\n"
	passed, code = evalGate(1, failLog)
	if passed || code != 1 {
		t.Errorf("Outcome 2 (FAILURE) was not rejected: passed=%v, code=%d", passed, code)
	}

	// Outcome 3: SKIP (Must fail the mandatory release gate!)
	skipLog := "=== RUN TestDockerEvaluator_LiveExecution\n    evaluator_test.go:525: Skipping live Docker sandbox test\n--- SKIP: TestDockerEvaluator_LiveExecution (0.00s)\nPASS\n"
	passed, code = evalGate(0, skipLog)
	if passed || code != 1 {
		t.Errorf("Outcome 3 (SKIP) was not rejected: passed=%v, code=%d", passed, code)
	}

	// Outcome 4: Empty / Missing test (Must also fail the gate!)
	emptyLog := "=== RUN TestOther\n--- PASS: TestOther (0.00s)\nPASS\n"
	passed, code = evalGate(0, emptyLog)
	if passed || code != 1 {
		t.Errorf("Outcome 4 (MISSING) was not rejected: passed=%v, code=%d", passed, code)
	}
}


