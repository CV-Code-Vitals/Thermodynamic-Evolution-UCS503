package semantic

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// 1. Test chains of length 1, 2, 3, 5, and beyond configured budget
func TestAliasAnalysis_TransitiveChainsAndBudget(t *testing.T) {
	src := `package testalias
import "os"
func ChainFunction(path string) {
	f, _ := os.Open(path)
	a1 := f
	a2 := a1
	a3 := a2
	a4 := a3
	a5 := a4
	a6 := a5 // exceeds 5 hops!
	_ = a6
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "chain.go", src, 0)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	fnDecl := funcOrGeneric(file.Decls[1])
	graph := BuildAliasGraphFromAST(fnDecl, fset, 5, 32)

	closure := graph.ComputeTransitiveClosure("f")

	// Chain 1: a1 (1 hop)
	if edge, ok := closure["a1"]; !ok || edge.HopCount != 1 || edge.Relation != AliasMustAlias {
		t.Errorf("Expected a1 at 1 hop MUST_ALIAS, got %+v", edge)
	}
	// Chain 2: a2 (2 hops)
	if edge, ok := closure["a2"]; !ok || edge.HopCount != 2 || edge.Relation != AliasMustAlias {
		t.Errorf("Expected a2 at 2 hops MUST_ALIAS, got %+v", edge)
	}
	// Chain 3: a3 (3 hops)
	if edge, ok := closure["a3"]; !ok || edge.HopCount != 3 || edge.Relation != AliasMustAlias {
		t.Errorf("Expected a3 at 3 hops MUST_ALIAS, got %+v", edge)
	}
	// Chain 5: a5 (5 hops - exactly at budget)
	if edge, ok := closure["a5"]; !ok || edge.HopCount != 5 || edge.Relation != AliasMustAlias {
		t.Errorf("Expected a5 at 5 hops MUST_ALIAS, got %+v", edge)
	}
	// Beyond budget: a6 (6 hops - must be truncated / not in closure)
	if _, ok := closure["a6"]; ok {
		t.Errorf("a6 exceeds MaxHops (5) and should be truncated out of closure")
	}
	if !graph.Truncated["f"] {
		t.Errorf("Expected graph.Truncated['f'] to be true")
	}
}

// 2. Test reassignments and alias invalidation
func TestAliasAnalysis_ReassignmentAndInvalidation(t *testing.T) {
	src := `package testalias
import "os"
func ReassignFunction(path string) {
	f1, _ := os.Open(path)
	alias := f1
	alias.Close() // valid alias here

	// Reassign alias to another file
	f2, _ := os.Open(path)
	alias = f2
	// At this point, alias is no longer an alias of f1!
	_ = alias
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "reassign.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])
	graph := BuildAliasGraphFromAST(fnDecl, fset, 5, 32)

	// Line 7: alias.Close() -> valid alias of f1
	isValid7, rel7 := graph.IsAliasAtLine("f1", "alias", 7)
	if !isValid7 || rel7 != AliasMustAlias {
		t.Errorf("Expected alias to be MUST_ALIAS of f1 at line 7, got %v (%s)", isValid7, rel7)
	}

	// Line 13: after alias = f2 -> alias must be INVALIDATED for f1
	isValid13, _ := graph.IsAliasAtLine("f1", "alias", 13)
	if isValid7 && isValid13 {
		t.Errorf("Alias should be invalidated for f1 after reassignment at line 11")
	}
}

// 3. Test pointer aliases
func TestAliasAnalysis_PointerAliases(t *testing.T) {
	src := `package testalias
import "os"
func PointerFunction(path string) {
	f, _ := os.Open(path)
	p := &f
	f2 := *p
	_ = f2
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "ptr.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])
	graph := BuildAliasGraphFromAST(fnDecl, fset, 5, 32)

	closure := graph.ComputeTransitiveClosure("f")

	if edge, ok := closure["p"]; !ok || !edge.IsPointer {
		t.Errorf("Expected pointer alias 'p' for 'f', got %+v", edge)
	}
	if edge, ok := closure["f2"]; !ok {
		t.Errorf("Expected 'f2' to be recognized as alias of 'f' via pointer dereference, got %+v", edge)
	}
}

// 4. Test branch-dependent alias relationships (MAY_ALIAS vs MUST_ALIAS)
func TestAliasAnalysis_BranchDependentMayAlias(t *testing.T) {
	src := `package testalias
import "os"
func ConditionalAlias(path string, cond bool) {
	f, _ := os.Open(path)
	var target *os.File
	if cond {
		target = f
	}
	_ = target
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "cond.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])
	graph := BuildAliasGraphFromAST(fnDecl, fset, 5, 32)

	closure := graph.ComputeTransitiveClosure("f")

	if edge, ok := closure["target"]; !ok || edge.Relation != AliasMayAlias {
		t.Errorf("Expected 'target' to be MAY_ALIAS due to conditional branch, got %+v", edge)
	}
}

// 5. Test cyclic alias graphs
func TestAliasAnalysis_CyclicGraph(t *testing.T) {
	src := `package testalias
func CyclicFunction() {
	var a, b string
	a = b
	b = a
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "cycle.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[0])
	graph := BuildAliasGraphFromAST(fnDecl, fset, 5, 32)

	// Must terminate cleanly without infinite loop
	closure := graph.ComputeTransitiveClosure("a")
	if len(closure) == 0 {
		t.Errorf("Expected closure to contain 'b'")
	}
}

// 6. Test resource release through multi-hop alias in ResourceAnalyzer
func TestAliasAnalysis_ResourceReleaseThroughMultiHopAlias(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	sem := NewSemanticAnalyzer(repoRoot)
	engine := NewFunctionSummaryEngine(repoRoot)

	// pkg_b.CloserHelper releases its parameter
	fPathB := filepath.Join("benchmarks", "cross_package_stress_v1", "pkg_b", "helpers.go")
	fModelB, _, _ := sem.AnalyzeFile(fPathB)
	cfgCloser := fModelB.CFGs["CloserHelper"]
	sumCloser := engine.ExtractIntraproceduralSummary(cfgCloser.FuncDecl, fModelB.FileSet, cfgCloser, fModelB.Dataflow["CloserHelper"], "pkg_b", "pkg_b")
	engine.RegisterSummary(sumCloser)

	src := `package testalias
import (
	"os"
	"pkg_b"
)
func ReleaseThroughChain(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	h1 := f
	h2 := h1
	h3 := h2
	return pkg_b.CloserHelper(h3) // 3 hops away!
}`
	fset := token.NewFileSet()
	fParsed, _ := parser.ParseFile(fset, "chain_rel.go", src, 0)
	fnDecl := funcOrGeneric(fParsed.Decls[1])
	cfg := BuildCFG(fnDecl, fset, "chain_rel.go")

	resAnalyzer := NewResourceAnalyzer(repoRoot)
	resAnalyzer.SetSummaryEngine(engine)

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "chain_rel.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Multi-hop alias release failed to recognize CloserHelper(h3) as clean exit: %+v", ev)
		}
	}
}
