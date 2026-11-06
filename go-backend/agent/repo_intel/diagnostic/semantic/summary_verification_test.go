package semantic

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Task B Verification: 8 Core Properties with Controlled Counterfactuals

// 1. Verify summaries encode genuine parameter-to-return dependencies (and counterfactual on constants)
func TestVerifyParamReturnFlows(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)

	srcPositive := `package testpkg
import "fmt"
func FormatUserQuery(table, col, val string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = '%s'", table, col, val)
}`

	srcCounterfactual := `package testpkg
func SafeConstantQuery(table, col, val string) string {
	return "SELECT * FROM static_table WHERE static_col = '1'"
}`

	fset := token.NewFileSet()
	filePos, err := parser.ParseFile(fset, "pos.go", srcPositive, 0)
	if err != nil {
		t.Fatalf("ParseFile pos failed: %v", err)
	}
	fileNeg, err := parser.ParseFile(fset, "neg.go", srcCounterfactual, 0)
	if err != nil {
		t.Fatalf("ParseFile neg failed: %v", err)
	}

	fnDeclPos := funcOrGeneric(filePos.Decls[1])
	fnDeclNeg := funcOrGeneric(fileNeg.Decls[0])

	sumPos := engine.ExtractIntraproceduralSummary(fnDeclPos, fset, nil, nil, "testpkg", "testpkg")
	sumNeg := engine.ExtractIntraproceduralSummary(fnDeclNeg, fset, nil, nil, "testpkg", "testpkg")

	// Positive assertion: all 3 parameters flow to return 0 with TAINT_TRANSFORM
	if len(sumPos.ParamReturnFlows) != 3 {
		t.Fatalf("Expected 3 ParamReturnFlows in FormatUserQuery, got %d", len(sumPos.ParamReturnFlows))
	}
	for i, flow := range sumPos.ParamReturnFlows {
		if flow.ParamIndex != i || flow.ReturnIndex != 0 || flow.Effect != EffectTaintTransform {
			t.Errorf("Flow %d mismatch: got %+v", i, flow)
		}
	}

	// Counterfactual assertion: zero parameter return flows because return is a constant literal
	if len(sumNeg.ParamReturnFlows) != 0 {
		t.Fatalf("Counterfactual failed: expected 0 ParamReturnFlows in SafeConstantQuery, got %d", len(sumNeg.ParamReturnFlows))
	}
}

// 2. Verify SQL taint propagation uses composed semantic effects (and counterfactual on safe query)
func TestVerifySQLTaintComposedPropagation(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	orch := NewUnifiedOrchestrator(repoRoot)

	// Evaluate pkg_a/caller.go against pkg_b/helpers.go
	files := []string{
		filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_a", "caller.go"),
		filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go"),
	}

	report, err := orch.RunCompleteDiagnostic(files, "a64e7632476c1341343481661a3e246c69e5a644")
	if err != nil {
		t.Fatalf("RunCompleteDiagnostic failed: %v", err)
	}

	// Verify that CallerSQLCrossPackage at line 44 is detected as SQL injection via helper
	foundSQL := false
	for _, f := range report.ReconciliationLedger.Findings {
		if f.RuleID == "SEMANTIC_SQL_INJECTION" && strings.Contains(f.TargetFile, "caller.go") && f.LineStart == 44 {
			foundSQL = true
			if !strings.Contains(f.Description, "FormatQuery") {
				t.Errorf("Expected finding description to reference helper 'FormatQuery', got: %s", f.Description)
			}
			break
		}
	}
	if !foundSQL {
		t.Errorf("Expected SQL injection finding at caller.go:44 via helper composition")
	}

	// Verify that CleanControls (e.g. CallerRunCrossFile or LocalFormat) do NOT trigger SQL injection
	for _, f := range report.ReconciliationLedger.Findings {
		if f.RuleID == "SEMANTIC_SQL_INJECTION" && strings.Contains(f.TargetFile, "helper_local.go") {
			t.Errorf("Unexpected SQL injection in helper_local.go: %+v", f)
		}
	}
}

// 3. Verify Resource cleanup effects propagate through helper calls (and counterfactual on missing release)
func TestVerifyResourceCleanupPropagation(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	// Step 1: Extract summary for CloserHelper (positive)
	fPathB := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go")
	fModelB, _, err := sem.AnalyzeFile(fPathB)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_b failed: %v", err)
	}

	cfgCloser := fModelB.CFGs["CloserHelper"]
	sumCloser := engine.ExtractIntraproceduralSummary(cfgCloser.FuncDecl, fModelB.FileSet, cfgCloser, fModelB.Dataflow["CloserHelper"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	engine.RegisterSummary(sumCloser)

	// Step 2: Analyze CallerResourceSplit with summary active -> must NOT leak
	resAnalyzer := NewResourceAnalyzer(repoRoot)
	resAnalyzer.SetSummaryEngine(engine)

	fPathA := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_a", "caller.go")
	fModelA, _, err := sem.AnalyzeFile(fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_a failed: %v", err)
	}

	cfgSplit := fModelA.CFGs["CallerResourceSplit"]
	evsSplit, err := resAnalyzer.AnalyzeFunction(cfgSplit.FuncDecl, fModelA.FileSet, cfgSplit, fModelA.Dataflow["CallerResourceSplit"], fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFunction CallerResourceSplit failed: %v", err)
	}

	// Must certify clean release, NOT a leak finding
	hasLeak := false
	for _, ev := range evsSplit {
		if ev.FindingState == StateConfirmed {
			hasLeak = true
			break
		}
	}
	if hasLeak {
		t.Fatalf("Expected 0 leak findings for CallerResourceSplit when CloserHelper summary is registered, got confirmed leak: %+v", evsSplit)
	}
	if len(evsSplit) == 0 || evsSplit[0].LifecycleState != StateResourceReleased {
		t.Fatalf("Expected safe release certification (RESOURCE_RELEASED), got %+v", evsSplit)
	}

	// Counterfactual: Analyze with a dummy summary that does NOT have RESOURCE_RELEASE
	dummyEngine := NewFunctionSummaryEngine(repoRoot)
	dummySummary := &FunctionSummary{
		PackageName:     "pkg_b",
		FunctionName:    "CloserHelper",
		ResourceEffects: []ResourceSummaryEffect{}, // empty effects!
		IsComplete:      true,
	}
	dummyEngine.RegisterSummary(dummySummary)

	resAnalyzerDummy := NewResourceAnalyzer(repoRoot)
	resAnalyzerDummy.SetSummaryEngine(dummyEngine)

	evsDummy, err := resAnalyzerDummy.AnalyzeFunction(cfgSplit.FuncDecl, fModelA.FileSet, cfgSplit, fModelA.Dataflow["CallerResourceSplit"], fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFunction with dummy engine failed: %v", err)
	}

	if len(evsDummy) == 0 {
		t.Fatalf("Counterfactual failed: expected resource leak when CloserHelper does NOT have release effect, got 0 findings")
	}
	if evsDummy[0].LifecycleState != StateResourceLeakConfirmed && evsDummy[0].LifecycleState != StateResourceLeakProbable {
		t.Errorf("Unexpected counterfactual evidence lifecycle state: %+v", evsDummy[0].LifecycleState)
	}
}

// 4. Verify Resource acquisition effects propagate through returned handles (and counterfactual on non-resource factory)
func TestVerifyResourceAcquisitionPropagation(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	// Extract summary for AcquireFile
	fPathB := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go")
	fModelB, _, err := sem.AnalyzeFile(fPathB)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_b failed: %v", err)
	}

	cfgAcquire := fModelB.CFGs["AcquireFile"]
	sumAcquire := engine.ExtractIntraproceduralSummary(cfgAcquire.FuncDecl, fModelB.FileSet, cfgAcquire, fModelB.Dataflow["AcquireFile"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	engine.RegisterSummary(sumAcquire)

	// Analyze CallerResourceTransfer -> must be flagged as resource leak (f is unclosed)
	resAnalyzer := NewResourceAnalyzer(repoRoot)
	resAnalyzer.SetSummaryEngine(engine)

	fPathA := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_a", "caller.go")
	fModelA, _, err := sem.AnalyzeFile(fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_a failed: %v", err)
	}

	cfgTransfer := fModelA.CFGs["CallerResourceTransfer"]
	evsTransfer, err := resAnalyzer.AnalyzeFunction(cfgTransfer.FuncDecl, fModelA.FileSet, cfgTransfer, fModelA.Dataflow["CallerResourceTransfer"], fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFunction CallerResourceTransfer failed: %v", err)
	}

	if len(evsTransfer) == 0 {
		t.Fatalf("Expected resource leak in CallerResourceTransfer due to AcquireFile acquisition summary, got 0")
	}
	if evsTransfer[0].Line != 57 {
		t.Errorf("Expected leak at line 57, got %d", evsTransfer[0].Line)
	}

	// Counterfactual: Analyze with un-summarized engine (Config A behavior)
	resAnalyzerNoSum := NewResourceAnalyzer(repoRoot) // no summary engine
	evsNoSum, err := resAnalyzerNoSum.AnalyzeFunction(cfgTransfer.FuncDecl, fModelA.FileSet, cfgTransfer, fModelA.Dataflow["CallerResourceTransfer"], fPathA)
	if err != nil {
		t.Fatalf("AnalyzeFunction CallerResourceTransfer without summary failed: %v", err)
	}

	if len(evsNoSum) != 0 {
		t.Fatalf("Counterfactual failed: intraprocedural analyzer should NOT detect AcquireFile as acquisition, got %d", len(evsNoSum))
	}
}

// 5. Verify Command-flow results depend on semantic propagation where claimed
func TestVerifyCommandFlowSemanticPropagation(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	// Extract summary for BuildShellScript in pkg_b
	fPathB := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go")
	fModelB, _, err := sem.AnalyzeFile(fPathB)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_b failed: %v", err)
	}

	cfgBuild := fModelB.CFGs["BuildShellScript"]
	sumBuild := engine.ExtractIntraproceduralSummary(cfgBuild.FuncDecl, fModelB.FileSet, cfgBuild, fModelB.Dataflow["BuildShellScript"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	engine.RegisterSummary(sumBuild)

	if len(sumBuild.ParamReturnFlows) == 0 || sumBuild.ParamReturnFlows[0].Effect != EffectTaintTransform {
		t.Fatalf("BuildShellScript must have EffectTaintTransform flow, got: %+v", sumBuild.ParamReturnFlows)
	}

	// Counterfactual: a static shell script builder does NOT produce taint flow
	srcStatic := `package pkg_b
func BuildStaticScript() string {
	return "echo 'hello static world'"
}`
	fset := token.NewFileSet()
	fStatic, _ := parser.ParseFile(fset, "static.go", srcStatic, 0)
	fnStatic := funcOrGeneric(fStatic.Decls[0])
	sumStatic := engine.ExtractIntraproceduralSummary(fnStatic, fset, nil, nil, "pkg_b", "pkg_b")
	if len(sumStatic.ParamReturnFlows) != 0 {
		t.Fatalf("Expected 0 ParamReturnFlows in BuildStaticScript, got %d", len(sumStatic.ParamReturnFlows))
	}
}

// 6. Verify Unknown callees remain conservative (EFFECT_UNKNOWN)
func TestVerifyUnknownCalleesConservative(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)

	// Lookup a non-existent function
	sum := engine.GetSummary("nonexistent_pkg", "", "UnknownOperation")
	if sum != nil {
		t.Fatalf("Expected nil summary for unknown function, got %+v", sum)
	}

	// Verify that an unknown callee inside code is treated conservatively
	srcUnknown := `package testpkg
import "os"
func CallUnknown(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	// Call unknown external helper
	externalpkg.UnknownCleanup(f)
	return nil
}`
	fset := token.NewFileSet()
	fParsed, err := parser.ParseFile(fset, "unknown.go", srcUnknown, 0)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	fnDecl := funcOrGeneric(fParsed.Decls[1])
	cfg := BuildCFG(fnDecl, fset, "unknown.go")
	if cfg == nil {
		t.Fatalf("BuildCFG returned nil")
	}

	resAnalyzer := NewResourceAnalyzer(repoRoot)
	resAnalyzer.SetSummaryEngine(engine)

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "unknown.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	// Because externalpkg.UnknownCleanup is unknown, conservative analysis must NOT assume it closes f!
	if len(evs) == 0 {
		t.Fatalf("Expected leak finding for unverified unknown callee externalpkg.UnknownCleanup(f)")
	}
	if evs[0].LifecycleState != StateResourceLeakConfirmed && evs[0].LifecycleState != StateResourceLeakProbable {
		t.Errorf("Expected leak confirmed/probable, got: %s", evs[0].LifecycleState)
	}
}

// 7. Verify Recursion terminates within bounded iterations
func TestVerifyRecursionBoundedTermination(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	fPathC := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_c", "callchains.go")
	fModelC, _, err := sem.AnalyzeFile(fPathC)
	if err != nil {
		t.Fatalf("AnalyzeFile pkg_c failed: %v", err)
	}

	pkgModel := &PackageSemanticModel{
		PackageName: "pkg_c",
		PackagePath: filepath.Dir(fPathC),
		Files:       map[string]*FileSemanticModel{fPathC: fModelC},
	}
	pkgModels := map[string]*PackageSemanticModel{pkgModel.PackagePath: pkgModel}

	// Build summaries bottom-up (contains Factorial, Ping, Pong)
	engine.BuildSummariesBottomUp(pkgModels, nil)

	sumPing := engine.GetSummary("pkg_c", "", "Ping")
	if sumPing == nil {
		t.Fatalf("Summary for Ping not registered")
	}
	if sumPing.IterationCount > 3 {
		t.Errorf("Recursion iteration count exceeded bound (3): got %d", sumPing.IterationCount)
	}

	sumFact := engine.GetSummary("pkg_c", "", "Factorial")
	if sumFact == nil {
		t.Fatalf("Summary for Factorial not registered")
	}
	if sumFact.IterationCount > 3 {
		t.Errorf("Factorial iteration count exceeded bound (3): got %d", sumFact.IterationCount)
	}
}

// 8. Verify Summary caches invalidate correctly after source changes
func TestVerifySummaryCacheInvalidation(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)

	srcV1 := `package cachetest
import "fmt"
func Compute(val string) string {
	return fmt.Sprintf("v1: %s", val)
}`

	srcV2 := `package cachetest
func Compute(val string) string {
	return "v2: constant"
}`

	fset := token.NewFileSet()
	f1, _ := parser.ParseFile(fset, "v1.go", srcV1, 0)
	f2, _ := parser.ParseFile(fset, "v2.go", srcV2, 0)

	fn1 := funcOrGeneric(f1.Decls[1])
	fn2 := funcOrGeneric(f2.Decls[0])

	sum1 := engine.ExtractIntraproceduralSummary(fn1, fset, nil, nil, "cachetest", "cachetest")
	sum2 := engine.ExtractIntraproceduralSummary(fn2, fset, nil, nil, "cachetest", "cachetest")

	// Hashes must be strictly distinct
	if sum1.SourceHash == sum2.SourceHash {
		t.Fatalf("Source hashes must differ between v1 and v2, but both got %s", sum1.SourceHash)
	}

	// V1 has taint flow; V2 does not
	if len(sum1.ParamReturnFlows) == 0 {
		t.Errorf("Expected ParamReturnFlows in v1")
	}
	if len(sum2.ParamReturnFlows) != 0 {
		t.Errorf("Expected 0 ParamReturnFlows in v2, got %d", len(sum2.ParamReturnFlows))
	}

	// Cache invalidation registration check
	engine.RegisterSummary(sum1)
	retrieved1 := engine.GetSummary("cachetest", "", "Compute")
	if retrieved1.SourceHash != sum1.SourceHash {
		t.Errorf("Retrieved summary hash mismatch for v1")
	}

	// Register v2 (simulating source re-analysis after code change)
	engine.RegisterSummary(sum2)
	retrieved2 := engine.GetSummary("cachetest", "", "Compute")
	if retrieved2.SourceHash != sum2.SourceHash {
		t.Errorf("Retrieved summary hash mismatch for v2")
	}
	if len(retrieved2.ParamReturnFlows) != 0 {
		t.Errorf("Summary cache was not updated to v2 behavior: got %d flows", len(retrieved2.ParamReturnFlows))
	}
}

// helper for AST casting in test
func funcOrGeneric(decl interface{}) *ast.FuncDecl {
	if fd, ok := decl.(*ast.FuncDecl); ok {
		return fd
	}
	return nil
}

// Suppress unused imports
var (
	_ = sql.ErrNoRows
	_ = os.DevNull
	_ = fmt.Sprintf
)
