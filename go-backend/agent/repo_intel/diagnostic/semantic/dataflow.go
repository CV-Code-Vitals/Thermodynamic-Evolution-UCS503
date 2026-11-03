package semantic

import (
	"go/ast"
	"go/token"
)

// ComputeReachingDefinitions executes an intraprocedural reaching-definitions dataflow analysis.
func ComputeReachingDefinitions(cfg *ControlFlowGraph, fset *token.FileSet) *ReachingDefinitionsAnalysis {
	if cfg == nil || cfg.Entry == nil {
		return nil
	}

	analysis := &ReachingDefinitionsAnalysis{
		CFG:          cfg,
		Definitions:  make([]*Definition, 0),
		Gen:          make(map[int]map[int]bool),
		Kill:         make(map[int]map[int]bool),
		In:           make(map[int]map[int]bool),
		Out:          make(map[int]map[int]bool),
		DefUseChains: make(map[int][]VariableUse),
	}

	// 1. Collect all variable definitions
	defCounter := 0
	varDefMap := make(map[string][]int) // VarName -> list of Def IDs

	// Entry block definitions: function parameters & receiver
	if cfg.FuncDecl != nil {
		if cfg.FuncDecl.Recv != nil {
			for _, field := range cfg.FuncDecl.Recv.List {
				for _, name := range field.Names {
					pos := fset.Position(name.Pos())
					def := &Definition{
						ID:          defCounter,
						VarName:     name.Name,
						TypeName:    formatTypeExpr(field.Type, fset),
						BlockID:     cfg.Entry.ID,
						Line:        pos.Line,
						IsParameter: true,
						Expr:        name,
					}
					analysis.Definitions = append(analysis.Definitions, def)
					varDefMap[name.Name] = append(varDefMap[name.Name], def.ID)
					defCounter++
				}
			}
		}

		if cfg.FuncDecl.Type.Params != nil {
			for _, field := range cfg.FuncDecl.Type.Params.List {
				typeStr := formatTypeExpr(field.Type, fset)
				for _, name := range field.Names {
					pos := fset.Position(name.Pos())
					def := &Definition{
						ID:          defCounter,
						VarName:     name.Name,
						TypeName:    typeStr,
						BlockID:     cfg.Entry.ID,
						Line:        pos.Line,
						IsParameter: true,
						Expr:        name,
					}
					analysis.Definitions = append(analysis.Definitions, def)
					varDefMap[name.Name] = append(varDefMap[name.Name], def.ID)
					defCounter++
				}
			}
		}
	}

	// Block-level definitions
	blockDefs := make(map[int][]*Definition)
	blockUses := make(map[int][]VariableUse)

	for _, block := range cfg.Blocks {
		for _, stmt := range block.Statements {
			defs, uses := extractStmtDefsAndUses(stmt, block.ID, fset, &defCounter)
			for _, d := range defs {
				analysis.Definitions = append(analysis.Definitions, d)
				varDefMap[d.VarName] = append(varDefMap[d.VarName], d.ID)
				blockDefs[block.ID] = append(blockDefs[block.ID], d)
			}
			blockUses[block.ID] = append(blockUses[block.ID], uses...)
		}
		// Also inspect terminator condition for uses
		if block.Terminator != nil {
			uses := extractExprUses(block.Terminator, block.ID, fset)
			blockUses[block.ID] = append(blockUses[block.ID], uses...)
		}
	}

	// 2. Compute Gen and Kill sets for each block
	for _, block := range cfg.Blocks {
		genSet := make(map[int]bool)
		killSet := make(map[int]bool)

		// For entry block, include parameter definitions in Gen
		if block.Kind == BlockEntry {
			for _, d := range analysis.Definitions {
				if d.IsParameter && d.BlockID == block.ID {
					genSet[d.ID] = true
				}
			}
		}

		// Process statement definitions within the block
		for _, def := range blockDefs[block.ID] {
			// This def kills all prior defs of the same variable
			for _, otherID := range varDefMap[def.VarName] {
				if otherID != def.ID {
					killSet[otherID] = true
				}
			}
			// This def is in Gen
			genSet[def.ID] = true
		}

		analysis.Gen[block.ID] = genSet
		analysis.Kill[block.ID] = killSet
		analysis.In[block.ID] = make(map[int]bool)
		analysis.Out[block.ID] = copySet(genSet)
	}

	// 3. Worklist iteration for reaching definitions
	worklist := make([]*Block, len(cfg.Blocks))
	copy(worklist, cfg.Blocks)

	for len(worklist) > 0 {
		curr := worklist[0]
		worklist = worklist[1:]

		// In[curr] = Union of Out[P] for all P in Pred(curr)
		newIn := make(map[int]bool)
		for _, pred := range curr.Predecessors {
			for defID := range analysis.Out[pred.ID] {
				newIn[defID] = true
			}
		}
		analysis.In[curr.ID] = newIn

		// Out[curr] = Gen[curr] Union (In[curr] - Kill[curr])
		newOut := make(map[int]bool)
		for defID := range newIn {
			if !analysis.Kill[curr.ID][defID] {
				newOut[defID] = true
			}
		}
		for defID := range analysis.Gen[curr.ID] {
			newOut[defID] = true
		}

		if !equalSets(newOut, analysis.Out[curr.ID]) {
			analysis.Out[curr.ID] = newOut
			for _, succ := range curr.Successors {
				if !containsBlock(worklist, succ) {
					worklist = append(worklist, succ)
				}
			}
		}
	}

	// 4. Build Definition-Use Chains
	for _, block := range cfg.Blocks {
		reachingAtBlockEntry := copySet(analysis.In[block.ID])

		// Track defs within block to associate intra-block uses
		localDefs := make(map[string]int) // VarName -> latest local Def ID

		for _, stmt := range block.Statements {
			stmtUses := extractExprUsesFromStmt(stmt, block.ID, fset)
			for _, use := range stmtUses {
				if localDefID, ok := localDefs[use.VarName]; ok {
					analysis.DefUseChains[localDefID] = append(analysis.DefUseChains[localDefID], use)
				} else {
					// Use comes from reaching definitions at block entry
					for defID := range reachingAtBlockEntry {
						def := analysis.Definitions[defID]
						if def.VarName == use.VarName {
							analysis.DefUseChains[defID] = append(analysis.DefUseChains[defID], use)
						}
					}
				}
			}

			// Update local definitions from this statement
			stmtDefs, _ := extractStmtDefsAndUses(stmt, block.ID, fset, nil)
			for _, d := range stmtDefs {
				localDefs[d.VarName] = d.ID
			}
		}
	}

	return analysis
}

func extractStmtDefsAndUses(stmt ast.Stmt, blockID int, fset *token.FileSet, counter *int) ([]*Definition, []VariableUse) {
	var defs []*Definition
	var uses []VariableUse

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for i, lhs := range s.Lhs {
			pos := fset.Position(lhs.Pos())
			if id, ok := lhs.(*ast.Ident); ok {
				if id.Name != "_" {
					var rhsExpr ast.Expr
					if i < len(s.Rhs) {
						rhsExpr = s.Rhs[i]
					}
					defID := 0
					if counter != nil {
						defID = *counter
						*counter++
					}
					defs = append(defs, &Definition{
						ID:       defID,
						VarName:  id.Name,
						BlockID:  blockID,
						Line:     pos.Line,
						Expr:     rhsExpr,
						Stmt:     s,
					})
				}
			}
		}
		// RHS expressions are uses
		for _, rhs := range s.Rhs {
			uses = append(uses, extractExprUses(rhs, blockID, fset)...)
		}

	case *ast.DeclStmt:
		if gen, ok := s.Decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				if val, ok := spec.(*ast.ValueSpec); ok {
					typeStr := formatTypeExpr(val.Type, fset)
					for i, name := range val.Names {
						pos := fset.Position(name.Pos())
						var rhsExpr ast.Expr
						if i < len(val.Values) {
							rhsExpr = val.Values[i]
						}
						defID := 0
						if counter != nil {
							defID = *counter
							*counter++
						}
						defs = append(defs, &Definition{
							ID:       defID,
							VarName:  name.Name,
							TypeName: typeStr,
							BlockID:  blockID,
							Line:     pos.Line,
							Expr:     rhsExpr,
							Stmt:     s,
						})
					}
					for _, v := range val.Values {
						uses = append(uses, extractExprUses(v, blockID, fset)...)
					}
				}
			}
		}

	case *ast.ExprStmt:
		uses = append(uses, extractExprUses(s.X, blockID, fset)...)

	case *ast.ReturnStmt:
		for _, res := range s.Results {
			uses = append(uses, extractExprUses(res, blockID, fset)...)
		}
	}

	return defs, uses
}

func extractExprUsesFromStmt(stmt ast.Stmt, blockID int, fset *token.FileSet) []VariableUse {
	var uses []VariableUse
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for _, rhs := range s.Rhs {
			uses = append(uses, extractExprUses(rhs, blockID, fset)...)
		}
	case *ast.ExprStmt:
		uses = append(uses, extractExprUses(s.X, blockID, fset)...)
	case *ast.ReturnStmt:
		for _, res := range s.Results {
			uses = append(uses, extractExprUses(res, blockID, fset)...)
		}
	}
	return uses
}

func extractExprUses(expr ast.Expr, blockID int, fset *token.FileSet) []VariableUse {
	var uses []VariableUse
	if expr == nil {
		return uses
	}

	ast.Inspect(expr, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name == "_" || id.Name == "nil" || id.Name == "true" || id.Name == "false" {
			return true
		}

		pos := fset.Position(id.Pos())
		uses = append(uses, VariableUse{
			VarName: id.Name,
			BlockID: blockID,
			Line:    pos.Line,
			Expr:    id,
		})
		return true
	})

	return uses
}

func copySet(src map[int]bool) map[int]bool {
	dst := make(map[int]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func equalSets(a, b map[int]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func containsBlock(list []*Block, target *Block) bool {
	for _, b := range list {
		if b == target {
			return true
		}
	}
	return false
}
