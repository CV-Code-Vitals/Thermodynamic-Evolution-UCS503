package semantic

import (
	"path/filepath"
	"testing"
)

func TestFunctionSummaryIntraproceduralExtraction(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	// Analyze pkg_b/helpers.go
	fPath := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go")
	fModel, _, err := sem.AnalyzeFile(fPath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// 1. FormatQuery should have TAINT_TRANSFORM flow from params to return 0
	cfgFormat := fModel.CFGs["FormatQuery"]
	if cfgFormat == nil {
		t.Fatalf("CFG for FormatQuery not found")
	}

	sumFormat := engine.ExtractIntraproceduralSummary(cfgFormat.FuncDecl, fModel.FileSet, cfgFormat, fModel.Dataflow["FormatQuery"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	if sumFormat == nil {
		t.Fatalf("Summary for FormatQuery is nil")
	}

	if len(sumFormat.ParamReturnFlows) == 0 {
		t.Errorf("Expected ParamReturnFlows in FormatQuery, got 0")
	}
	hasTaint := false
	for _, flow := range sumFormat.ParamReturnFlows {
		if flow.Effect == EffectTaintTransform {
			hasTaint = true
			break
		}
	}
	if !hasTaint {
		t.Errorf("Expected EffectTaintTransform in FormatQuery")
	}

	// 2. CloserHelper should have RESOURCE_RELEASE on param 0
	cfgCloser := fModel.CFGs["CloserHelper"]
	if cfgCloser == nil {
		t.Fatalf("CFG for CloserHelper not found")
	}
	sumCloser := engine.ExtractIntraproceduralSummary(cfgCloser.FuncDecl, fModel.FileSet, cfgCloser, fModel.Dataflow["CloserHelper"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	hasRelease := false
	for _, eff := range sumCloser.ResourceEffects {
		if eff.Effect == EffectResourceRelease && eff.ParamIndex == 0 {
			hasRelease = true
			break
		}
	}
	if !hasRelease {
		t.Errorf("Expected EffectResourceRelease on param 0 in CloserHelper")
	}

	// 3. AcquireFile should have RESOURCE_ACQUIRE on return 0
	cfgAcquire := fModel.CFGs["AcquireFile"]
	if cfgAcquire == nil {
		t.Fatalf("CFG for AcquireFile not found")
	}
	sumAcquire := engine.ExtractIntraproceduralSummary(cfgAcquire.FuncDecl, fModel.FileSet, cfgAcquire, fModel.Dataflow["AcquireFile"], "pkg_b", "benchmarks/cross_package_stress_v1/pkg_b")
	hasAcquire := false
	for _, eff := range sumAcquire.ResourceEffects {
		if eff.Effect == EffectResourceAcquire && eff.ReturnIndex == 0 {
			hasAcquire = true
			break
		}
	}
	if !hasAcquire {
		t.Errorf("Expected EffectResourceAcquire on return 0 in AcquireFile")
	}
}

func TestFunctionSummaryRecursionAndTermination(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	engine := NewFunctionSummaryEngine(repoRoot)
	sem := NewSemanticAnalyzer(repoRoot)

	// Analyze pkg_c/callchains.go (contains Factorial, Ping, Pong)
	fPath := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_c", "callchains.go")
	fModel, _, err := sem.AnalyzeFile(fPath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	pkgModel := &PackageSemanticModel{
		PackageName: "pkg_c",
		PackagePath: filepath.Dir(fPath),
		Files:       map[string]*FileSemanticModel{fPath: fModel},
	}

	pkgModels := map[string]*PackageSemanticModel{pkgModel.PackagePath: pkgModel}
	engine.BuildSummariesBottomUp(pkgModels, nil)

	sumPing := engine.GetSummary("pkg_c", "", "Ping")
	if sumPing == nil {
		t.Fatalf("Summary for Ping not found")
	}

	if sumPing.IterationCount > 3 {
		t.Errorf("Expected iteration count <= 3 for recursive Ping, got %d", sumPing.IterationCount)
	}
}
