package semantic

import (
	"path/filepath"
	"testing"
	"time"
)

func TestResilienceToMalformedAndIncompleteCode(t *testing.T) {
	sa := NewSemanticAnalyzer(".")

	// 1. Syntax Error / Truncated Source
	malformedSrc := `package broken
func IncompleteFunc( {
	x := 10
`
	_, _, err := sa.AnalyzeSource("malformed.go", malformedSrc)
	if err == nil {
		t.Errorf("expected error on malformed source, got nil")
	}

	// 2. Missing Dependencies and Unknown Packages
	missingDepSrc := `package custom
import "nonexistent/thirdparty/unknownpkg"

func ExecuteCustom(u *unknownpkg.Client, input string) {
	u.DoSomething(input)
}
`
	model, st, err := sa.AnalyzeSource("missing_dep.go", missingDepSrc)
	if err != nil {
		t.Fatalf("AnalyzeSource should not crash on missing third-party packages, got err: %v", err)
	}
	if st == nil {
		t.Fatal("expected fallback SymbolTable on missing dependency")
	}
	if model == nil || len(model.CFGs) == 0 {
		t.Fatal("expected CFG construction to succeed even with missing external types")
	}

	// 3. Unsupported / Opaque Expressions
	opaqueSrc := `package opaque
import (
	"database/sql"
	"reflect"
)

func DynamicReflection(db *sql.DB, val interface{}) {
	v := reflect.ValueOf(val)
	// Dynamic unresolvable query extraction
	q := v.MethodByName("GetQuery").Call(nil)[0].String()
	db.Query(q)
}
`
	modelOpaque, _, err := sa.AnalyzeSource("opaque.go", opaqueSrc)
	if err != nil {
		t.Fatalf("AnalyzeSource failed on opaque code: %v", err)
	}

	sqlAnalyzer := NewSQLAnalyzer(".")
	cfg := modelOpaque.CFGs["DynamicReflection"]
	df := modelOpaque.Dataflow["DynamicReflection"]

	evs, err := sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, modelOpaque.FileSet, cfg, df, "opaque.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction crashed on reflection query: %v", err)
	}

	if len(evs) == 0 {
		t.Fatal("expected evidence record for reflection query at sink")
	}

	ev := evs[0]
	// Should fail-closed to UNRESOLVED_SQL_CONSTRUCTION with INSUFFICIENT_EVIDENCE
	if ev.Classification != ClassificationUnresolvedSQLConstruction {
		t.Errorf("expected UNRESOLVED_SQL_CONSTRUCTION, got %s", ev.Classification)
	}
	if ev.LifecycleState != StateInsufficientEvidence {
		t.Errorf("expected StateInsufficientEvidence for unresolvable reflection, got %v", ev.LifecycleState)
	}
	if len(ev.AnalysisLimitations) == 0 {
		t.Errorf("expected analysis limitations to be recorded for unresolvable reflection")
	}
}

func TestPerformanceMetrics(t *testing.T) {
	// 1. Measure per-fixture analysis runtime
	fixturePath := filepath.Join("fixtures", "sql_fixtures.go")
	sa := NewSemanticAnalyzer(".")

	start := time.Now()
	model, _, err := sa.AnalyzeFile(fixturePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}
	fileParseDuration := time.Since(start)

	sqlAnalyzer := NewSQLAnalyzer(".")
	startAnalysis := time.Now()
	totalEvs := 0
	for fnName, cfg := range model.CFGs {
		df := model.Dataflow[fnName]
		evs, _ := sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, model.FileSet, cfg, df, fixturePath)
		totalEvs += len(evs)
	}
	analysisDuration := time.Since(startAnalysis)

	t.Logf("SQL Fixture file parse & CFG build duration: %v", fileParseDuration)
	t.Logf("SQL Fixture function analysis duration: %v for %d functions (%d evidence records)",
		analysisDuration, len(model.CFGs), totalEvs)

	if fileParseDuration > 200*time.Millisecond {
		t.Errorf("file parse duration unexpectedly high: %v", fileParseDuration)
	}
	if analysisDuration > 200*time.Millisecond {
		t.Errorf("analysis duration unexpectedly high: %v", analysisDuration)
	}

	// 2. Measure complete package analysis on external benchmark (http_router)
	extRouterPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "http_router")
	saRouter := NewSemanticAnalyzer(extRouterPath)

	pkgStart := time.Now()
	pkgModel, err := saRouter.AnalyzePackage(".")
	if err != nil {
		t.Fatalf("AnalyzePackage failed on http_router: %v", err)
	}
	pkgDuration := time.Since(pkgStart)

	t.Logf("http_router package analysis duration: %v for %d files", pkgDuration, len(pkgModel.Files))
	if pkgDuration > 500*time.Millisecond {
		t.Errorf("package analysis duration unexpectedly high: %v", pkgDuration)
	}
}
