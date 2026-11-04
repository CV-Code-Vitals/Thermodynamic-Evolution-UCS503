package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestBuildCFGBranchingAndLoops(t *testing.T) {
	src := `package demo

import "os"

func ProcessData(path string, count int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < count; i++ {
		if i == 5 {
			continue
		}
		if i == 10 {
			break
		}
	}

	return nil
}
`

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "demo.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "ProcessData" {
			fn = f
			break
		}
	}

	if fn == nil {
		t.Fatal("ProcessData function not found")
	}

	cfg := BuildCFG(fn, fset, "demo.go")
	if cfg == nil {
		t.Fatal("CFG is nil")
	}

	// 1. Verify Entry and Exit blocks
	if cfg.Entry == nil || cfg.Entry.Kind != BlockEntry {
		t.Errorf("expected valid Entry block")
	}
	if cfg.Exit == nil || cfg.Exit.Kind != BlockExit {
		t.Errorf("expected valid Exit block")
	}

	// 2. Verify Return Blocks
	if len(cfg.ReturnBlocks) != 2 {
		t.Errorf("expected 2 return blocks (one in error check, one at end), got %d", len(cfg.ReturnBlocks))
	}

	// 3. Verify Defer Extraction
	if len(cfg.AllDefers) != 1 {
		t.Fatalf("expected 1 defer site, got %d", len(cfg.AllDefers))
	}
	def := cfg.AllDefers[0]
	if def.TargetFunc != "Close" {
		t.Errorf("expected TargetFunc 'Close', got '%s'", def.TargetFunc)
	}
	if def.TargetVar != "f" {
		t.Errorf("expected TargetVar 'f', got '%s'", def.TargetVar)
	}
	if def.InLoop {
		t.Errorf("expected InLoop to be false")
	}

	// 4. Verify Total Blocks and Edges
	if len(cfg.Blocks) < 6 {
		t.Errorf("expected at least 6 blocks, got %d", len(cfg.Blocks))
	}

	// Verify all blocks are reachable from Entry or part of CFG
	for _, b := range cfg.Blocks {
		if b != cfg.Entry && len(b.Predecessors) == 0 {
			t.Errorf("block %d (%s) has 0 predecessors and is unreachable", b.ID, b.Kind)
		}
	}
}

func TestBuildCFGDeferInLoop(t *testing.T) {
	src := `package demo

import "os"

func BadLoop(paths []string) {
	for _, p := range paths {
		f, _ := os.Open(p)
		defer f.Close()
	}
}
`

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "bad_loop.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "BadLoop" {
			fn = f
			break
		}
	}

	cfg := BuildCFG(fn, fset, "bad_loop.go")
	if cfg == nil {
		t.Fatal("CFG is nil")
	}

	if len(cfg.AllDefers) != 1 {
		t.Fatalf("expected 1 defer, got %d", len(cfg.AllDefers))
	}
	if !cfg.AllDefers[0].InLoop {
		t.Errorf("expected defer f.Close() to be flagged as InLoop=true")
	}
}
