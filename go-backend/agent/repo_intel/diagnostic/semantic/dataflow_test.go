package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestComputeReachingDefinitions(t *testing.T) {
	src := `package demo

func Calculate(x int, flag bool) int {
	y := 10
	if flag {
		y = x + 5
	} else {
		y = x * 2
	}
	return y
}
`

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "calc.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name.Name == "Calculate" {
			fn = f
			break
		}
	}

	cfg := BuildCFG(fn, fset, "calc.go")
	analysis := ComputeReachingDefinitions(cfg, fset)

	if analysis == nil {
		t.Fatal("analysis is nil")
	}

	// 1. Verify Definitions
	// Parameters: x, flag (2 defs)
	// Body: y := 10 (1 def), y = x + 5 (1 def), y = x * 2 (1 def)
	// Total at least 5 definitions
	if len(analysis.Definitions) < 5 {
		t.Fatalf("expected at least 5 definitions, got %d", len(analysis.Definitions))
	}

	paramXFound := false
	paramFlagFound := false
	for _, def := range analysis.Definitions {
		if def.IsParameter && def.VarName == "x" {
			paramXFound = true
		}
		if def.IsParameter && def.VarName == "flag" {
			paramFlagFound = true
		}
	}

	if !paramXFound || !paramFlagFound {
		t.Errorf("expected parameter definitions for x and flag")
	}

	// 2. Verify Reaching Definitions at Exit Block
	// At exit block (where return y is evaluated), both y = x + 5 and y = x * 2 should reach
	exitIn := analysis.In[cfg.Exit.ID]
	yDefsReachingExit := 0
	for defID := range exitIn {
		def := analysis.Definitions[defID]
		if def.VarName == "y" {
			yDefsReachingExit++
		}
	}

	if yDefsReachingExit < 2 {
		t.Errorf("expected at least 2 definitions of y to reach exit (from then & else branches), got %d", yDefsReachingExit)
	}

	// 3. Verify Def-Use chains
	usesOfX := 0
	for _, def := range analysis.Definitions {
		if def.VarName == "x" {
			usesOfX += len(analysis.DefUseChains[def.ID])
		}
	}
	if usesOfX == 0 {
		t.Errorf("expected uses of parameter x to be recorded in DefUseChains")
	}
}
