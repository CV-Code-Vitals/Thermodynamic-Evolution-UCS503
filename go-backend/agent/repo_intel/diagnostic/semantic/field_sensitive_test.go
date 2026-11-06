package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// 1. Direct struct field assignment and release: h.file = f; h.file.Close()
func TestFieldSensitive_AssignmentAndRelease(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package teststruct
import "os"

type Holder struct {
	file *os.File
}

func FieldClean(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	h := &Holder{}
	h.file = f
	defer h.file.Close() // Explicit field release!
	return nil
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "clean.go", src, 0)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	fnDecl := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnDecl, fset, "clean.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "clean.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Field release h.file.Close() was flagged as leak: %+v", ev)
		}
	}
	if len(evs) == 0 || evs[0].LifecycleState != StateResourceDeferredRelease {
		t.Fatalf("Expected StateResourceDeferredRelease, got: %+v", evs)
	}
}

// 2. Composite literal assignment: h := &Holder{file: f}
func TestFieldSensitive_CompositeLiteralInit(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package teststruct
import "os"

type Holder struct {
	file *os.File
}

func CompositeClean(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	h := &Holder{file: f}
	defer h.file.Close()
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "comp.go", src, 0)

	fnDecl := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnDecl, fset, "comp.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "comp.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Composite literal field release was flagged as leak: %+v", ev)
		}
	}
}

// 3. Ownership escape through return: return h (OWNED_TRANSFERRED)
func TestFieldSensitive_OwnershipEscapeThroughReturn(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package teststruct
import "os"

type Holder struct {
	file *os.File
}

func FactoryHolder(path string) (*Holder, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	h := &Holder{file: f}
	return h, nil // Transferred to caller!
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "factory.go", src, 0)

	fnDecl := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnDecl, fset, "factory.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "factory.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Returning struct containing resource handle was flagged as leak: %+v", ev)
		}
	}
	if len(evs) == 0 || evs[0].LifecycleState != StateResourceTransferred {
		t.Fatalf("Expected StateResourceTransferred, got: %+v", evs)
	}
}

// 4. Local pointer aliases: p := h; p.file.Close()
func TestFieldSensitive_StructPointerAlias(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package teststruct
import "os"

type Holder struct {
	file *os.File
}

func PointerAliasClean(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	h := &Holder{file: f}
	p := h // Struct pointer alias!
	defer p.file.Close()
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "ptr_alias.go", src, 0)

	fnDecl := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnDecl, fset, "ptr_alias.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "ptr_alias.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Struct pointer alias p.file.Close() was flagged as leak: %+v", ev)
		}
	}
}

// 5. Nested fields up to depth 3: obj.inner.file
func TestFieldSensitive_NestedAccessPath(t *testing.T) {
	tracker := NewFieldSensitiveTracker(3)

	src := `package teststruct
import "os"

type Inner struct {
	file *os.File
}
type Outer struct {
	inner Inner
}

func NestedTest(path string) {
	f, _ := os.Open(path)
	var obj Outer
	obj.inner.file = f
	obj.inner.file.Close()
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "nested.go", src, 0)

	fnDecl := funcOrGeneric(file.Decls[3])
	assign := fnDecl.Body.List[2].(*ast.AssignStmt)
	call := fnDecl.Body.List[3].(*ast.ExprStmt).X.(*ast.CallExpr)

	known := map[string]ResourceKind{"f": ResourceKindFile}
	tracker.TrackAssignment(assign.Lhs[0], assign.Rhs[0], fset, known)

	field, ok := tracker.Fields["obj.inner.file"]
	if !ok || field.Depth != 2 {
		t.Fatalf("Expected obj.inner.file tracked at depth 2, got: %+v", field)
	}

	isRel := tracker.TrackReleaseCall(call, fset)
	if !isRel || field.State != StateFieldReleased {
		t.Fatalf("Expected nested release to mark obj.inner.file as RELEASED")
	}
}

// 6. Independent unrelated fields: cfg.SQL vs cfg.Timeout
func TestFieldSensitive_UnrelatedFieldsIndependent(t *testing.T) {
	tracker := NewFieldSensitiveTracker(3)
	fset := token.NewFileSet()

	src := `package teststruct
type Config struct {
	SQL     string
	Timeout int
	file    string
}
func ConfigTest() {
	var cfg Config
	cfg.SQL = "SELECT * FROM users"
	cfg.Timeout = 30
}`
	file, _ := parser.ParseFile(fset, "cfg.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])

	assignSQL := fnDecl.Body.List[1].(*ast.AssignStmt)
	assignTimeout := fnDecl.Body.List[2].(*ast.AssignStmt)

	known := make(map[string]ResourceKind)
	tracker.TrackAssignment(assignSQL.Lhs[0], assignSQL.Rhs[0], fset, known)
	tracker.TrackAssignment(assignTimeout.Lhs[0], assignTimeout.Rhs[0], fset, known)

	// Neither cfg.SQL nor cfg.Timeout should be tracked as a resource leak!
	if len(tracker.Fields) != 0 {
		t.Errorf("Expected 0 resource fields for non-resource struct Config, got %d: %+v", len(tracker.Fields), tracker.Fields)
	}
}

// 7. Counterfactual: Stored in struct and never released (unclosed leak)
func TestFieldSensitive_StoredAndNeverReleased_Counterfactual(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package teststruct
import "os"

type Holder struct {
	file *os.File
}

func LeakyHolder(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	h := &Holder{file: f}
	_ = h // Stored, but never closed and never returned!
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "leaky.go", src, 0)

	fnDecl := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnDecl, fset, "leaky.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "leaky.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	hasLeak := false
	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			hasLeak = true
			break
		}
	}
	if !hasLeak {
		t.Fatalf("Counterfactual failed: storing in struct without closing or returning must be flagged as leak: %+v", evs)
	}
}
