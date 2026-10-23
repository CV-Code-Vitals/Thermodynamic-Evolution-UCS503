package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrStaleSnapshot          = errors.New("mutation proposal rejected: base snapshot hash does not match current state")
	ErrNoApplicableStrategy   = errors.New("no mutation strategy produced an applicable candidate for target snippet")
	ErrConcurrentModification = errors.New("persistence aborted: target file was modified externally during evolution")
)

// MutationRequest provides full snapshot context to a mutation strategy.
type MutationRequest struct {
	CandidateID      string            `json:"candidate_id"`
	BaseSnapshotHash string            `json:"base_snapshot_hash"`
	FilePath         string            `json:"file_path"`
	FunctionName     string            `json:"function_name"`
	StartByte        int               `json:"start_byte"`
	EndByte          int               `json:"end_byte"`
	CurrentCode      string            `json:"current_code"`
	FullContent      []byte            `json:"-"`
	Iteration        int               `json:"iteration"`
	Temperature      float64           `json:"temperature"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// MutationProposal represents a generated candidate patch with provenance metadata.
type MutationProposal struct {
	CandidateID      string            `json:"candidate_id"`
	StrategyName     string            `json:"strategy_name"`
	BaseSnapshotHash string            `json:"base_snapshot_hash"`
	MutatedCode      string            `json:"mutated_code"`
	Description      string            `json:"description"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// MutationStrategy defines a specialized candidate mutation operator.
type MutationStrategy interface {
	Name() string
	CanMutate(code string) bool
	Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error)
}

// TargetedMutator extends CodeMutator with rich snapshot-aware candidate proposals.
type TargetedMutator interface {
	CodeMutator
	ProposeTargeted(ctx context.Context, req MutationRequest) (*MutationProposal, error)
	GetStrategyStats() map[string]StrategyStats
}

// StrategyStats tracks the empirical success of a mutation operator.
type StrategyStats struct {
	ProposalsGenerated int `json:"proposals_generated"`
	ValidCandidates    int `json:"valid_candidates"`
	Improvements       int `json:"improvements"`
}

// ── Strategy 1: Targeted Simplification (Nesting Depth Reduction) ──────────────

// TargetedSimplificationStrategy flattens deeply nested control flow blocks (reducing D_max).
type TargetedSimplificationStrategy struct{}

func (s *TargetedSimplificationStrategy) Name() string {
	return "targeted_simplification"
}

func (s *TargetedSimplificationStrategy) CanMutate(code string) bool {
	if _, ok := flattenNestedIf(code); ok {
		return true
	}
	if _, ok := removeElseAfterReturn(code); ok {
		return true
	}
	return false
}

func (s *TargetedSimplificationStrategy) Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	code := req.CurrentCode

	// Pattern A: Flatten nested `if A { if B { body } }` -> `if A && B { body }`
	if strings.Contains(code, "if ") {
		transformed, ok := flattenNestedIf(code)
		if ok && transformed != code {
			return &MutationProposal{
				CandidateID:      req.CandidateID,
				StrategyName:     s.Name(),
				BaseSnapshotHash: req.BaseSnapshotHash,
				MutatedCode:      transformed,
				Description:      "Flattened nested conditional branches into single composite guard",
			}, nil
		}
	}

	// Pattern B: Invert guard and remove else block: `if cond { return ... } else { body }`
	if strings.Contains(code, "} else {") {
		transformed, ok := removeElseAfterReturn(code)
		if ok && transformed != code {
			return &MutationProposal{
				CandidateID:      req.CandidateID,
				StrategyName:     s.Name(),
				BaseSnapshotHash: req.BaseSnapshotHash,
				MutatedCode:      transformed,
				Description:      "Inverted guard clause to eliminate else-block nesting",
			}, nil
		}
	}

	return nil, ErrNoApplicableStrategy
}

// ── Strategy 2: Control-Flow Simplification (Cyclomatic Complexity Reduction) ──

// ControlFlowSimplificationStrategy simplifies boolean returns and branching logic (reducing M).
type ControlFlowSimplificationStrategy struct{}

func (s *ControlFlowSimplificationStrategy) Name() string {
	return "control_flow_simplification"
}

func (s *ControlFlowSimplificationStrategy) CanMutate(code string) bool {
	_, ok := simplifyBooleanReturns(code)
	return ok
}

func (s *ControlFlowSimplificationStrategy) Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	code := req.CurrentCode

	// Pattern: `if cond { return true } return false` -> `return cond`
	transformed, ok := simplifyBooleanReturns(code)
	if ok && transformed != code {
		return &MutationProposal{
			CandidateID:      req.CandidateID,
			StrategyName:     s.Name(),
			BaseSnapshotHash: req.BaseSnapshotHash,
			MutatedCode:      transformed,
			Description:      "Replaced redundant boolean if-return branch with direct evaluation",
		}, nil
	}

	return nil, ErrNoApplicableStrategy
}

// ── Strategy 3: Duplication Reduction (AST Node Count Reduction) ──────────────

// DuplicationReductionStrategy extracts common repeated expressions (reducing S).
type DuplicationReductionStrategy struct{}

func (s *DuplicationReductionStrategy) Name() string {
	return "duplication_reduction"
}

func (s *DuplicationReductionStrategy) CanMutate(code string) bool {
	_, ok := extractCommonSubexpressions(code)
	return ok
}

func (s *DuplicationReductionStrategy) Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	code := req.CurrentCode

	transformed, ok := extractCommonSubexpressions(code)
	if ok && transformed != code {
		return &MutationProposal{
			CandidateID:      req.CandidateID,
			StrategyName:     s.Name(),
			BaseSnapshotHash: req.BaseSnapshotHash,
			MutatedCode:      transformed,
			Description:      "Extracted duplicate subexpressions into hoisted local variable",
		}, nil
	}

	return nil, ErrNoApplicableStrategy
}

// ── Strategy 4: Conservative Refactoring (Allocation & Concurrency Cleanup) ───

// ConservativeRefactoringStrategy performs localized AST cleanups (reducing C and S).
type ConservativeRefactoringStrategy struct{}

func (s *ConservativeRefactoringStrategy) Name() string {
	return "conservative_refactoring"
}

func (s *ConservativeRefactoringStrategy) CanMutate(code string) bool {
	_, ok := hoistLoopAllocations(code)
	return ok
}

func (s *ConservativeRefactoringStrategy) Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	code := req.CurrentCode

	transformed, ok := hoistLoopAllocations(code)
	if ok && transformed != code {
		return &MutationProposal{
			CandidateID:      req.CandidateID,
			StrategyName:     s.Name(),
			BaseSnapshotHash: req.BaseSnapshotHash,
			MutatedCode:      transformed,
			Description:      "Hoisted slice allocation out of hot loop body to reduce allocation pressure",
		}, nil
	}

	return nil, ErrNoApplicableStrategy
}

// ── Strategy 5: LLM Mutation Strategy Adapter ─────────────────────────────────

// LLMMutationStrategy wraps MutationAgent as a structured strategy.
type LLMMutationStrategy struct {
	Agent *MutationAgent
}

func (s *LLMMutationStrategy) Name() string {
	return "llm_mutation"
}

func (s *LLMMutationStrategy) CanMutate(code string) bool {
	return s.Agent != nil && s.Agent.APIBase != ""
}

func (s *LLMMutationStrategy) Mutate(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	if s.Agent == nil {
		return nil, errors.New("llm agent is not configured")
	}

	mutated, err := s.Agent.ProposeMutation(ctx, req.CurrentCode, req.FunctionName)
	if err != nil {
		return nil, fmt.Errorf("llm mutation failed: %w", err)
	}

	return &MutationProposal{
		CandidateID:      req.CandidateID,
		StrategyName:     s.Name(),
		BaseSnapshotHash: req.BaseSnapshotHash,
		MutatedCode:      mutated,
		Description:      "Refactored candidate generated by LLM agent",
	}, nil
}

// ── Adaptive / Multi-Strategy Selector ────────────────────────────────────────

// AdaptiveMutator combines multiple strategies with tracking and deterministic random selection.
type AdaptiveMutator struct {
	Strategies []MutationStrategy
	RandSource *rand.Rand
	mu         sync.Mutex
	stats      map[string]*StrategyStats
}

// NewAdaptiveMutator creates an adaptive mutator with default AST strategies and optional LLM agent.
func NewAdaptiveMutator(llmAgent *MutationAgent, rngSeed int64) *AdaptiveMutator {
	var rng *rand.Rand
	if rngSeed != 0 {
		rng = rand.New(rand.NewSource(rngSeed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	strats := []MutationStrategy{
		&TargetedSimplificationStrategy{},
		&ControlFlowSimplificationStrategy{},
		&DuplicationReductionStrategy{},
		&ConservativeRefactoringStrategy{},
	}

	if llmAgent != nil && llmAgent.APIBase != "" {
		strats = append(strats, &LLMMutationStrategy{Agent: llmAgent})
	}

	stats := make(map[string]*StrategyStats)
	for _, s := range strats {
		stats[s.Name()] = &StrategyStats{}
	}

	return &AdaptiveMutator{
		Strategies: strats,
		RandSource: rng,
		stats:      stats,
	}
}

// ProposeMutation satisfies the legacy CodeMutator interface.
func (a *AdaptiveMutator) ProposeMutation(ctx context.Context, currentCode, funcName string) (string, error) {
	h := sha256.Sum256([]byte(currentCode))
	req := MutationRequest{
		CandidateID:      uuid.New().String()[:8],
		BaseSnapshotHash: hex.EncodeToString(h[:]),
		FunctionName:     funcName,
		CurrentCode:      currentCode,
	}

	proposal, err := a.ProposeTargeted(ctx, req)
	if err != nil {
		return "", err
	}
	return proposal.MutatedCode, nil
}

// ProposeTargeted selects and executes an applicable strategy with provenance tracking.
func (a *AdaptiveMutator) ProposeTargeted(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.Strategies) == 0 {
		return nil, ErrNoApplicableStrategy
	}

	// Filter to strategies that can handle this snippet
	applicable := make([]MutationStrategy, 0, len(a.Strategies))
	for _, s := range a.Strategies {
		if s.CanMutate(req.CurrentCode) {
			applicable = append(applicable, s)
		}
	}

	if len(applicable) == 0 {
		// Fallback to all strategies
		applicable = a.Strategies
	}

	// Permute applicable strategies using deterministic RNG
	indices := a.RandSource.Perm(len(applicable))

	for _, idx := range indices {
		strat := applicable[idx]
		proposal, err := strat.Mutate(ctx, req)
		if err == nil && proposal != nil && proposal.MutatedCode != "" && proposal.MutatedCode != req.CurrentCode {
			// Record statistics
			st, ok := a.stats[strat.Name()]
			if !ok {
				st = &StrategyStats{}
				a.stats[strat.Name()] = st
			}
			st.ProposalsGenerated++

			return proposal, nil
		}
	}

	return nil, ErrNoApplicableStrategy
}

// RecordOutcome logs validation and improvement outcomes for adaptive learning.
func (a *AdaptiveMutator) RecordOutcome(strategyName string, passedCorrectness bool, improvedEnergy bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	st, ok := a.stats[strategyName]
	if !ok {
		return
	}
	if passedCorrectness {
		st.ValidCandidates++
	}
	if improvedEnergy {
		st.Improvements++
	}
}

// GetStrategyStats returns a snapshot copy of the strategy performance table.
func (a *AdaptiveMutator) GetStrategyStats() map[string]StrategyStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make(map[string]StrategyStats, len(a.stats))
	for k, v := range a.stats {
		out[k] = *v
	}
	return out
}

// ── AST Transformation Helpers ────────────────────────────────────────────────

// flattenNestedIf flattens `if A {\n\tif B {\n\tbody\n\t}\n}` to `if (A) && (B) {\n\tbody\n}`.
func flattenNestedIf(src string) (string, bool) {
	wrapped := fmt.Sprintf("package p\nfunc _() {\n%s\n}", src)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return src, false
	}

	modified := false
	ast.Inspect(file, func(n ast.Node) bool {
		if ifStmt, ok := n.(*ast.IfStmt); ok {
			// Check if ifStmt contains exactly one statement which is another IfStmt without Else
			if len(ifStmt.Body.List) == 1 && ifStmt.Else == nil {
				if innerIf, ok := ifStmt.Body.List[0].(*ast.IfStmt); ok && innerIf.Else == nil && innerIf.Init == nil && ifStmt.Init == nil {
					// Guard operator precedence: wrap operands with precedence lower than LAND (e.g. LOR)
					leftCond := ifStmt.Cond
					if bin, ok := leftCond.(*ast.BinaryExpr); ok && bin.Op.Precedence() < token.LAND.Precedence() {
						leftCond = &ast.ParenExpr{X: leftCond}
					}
					rightCond := innerIf.Cond
					if bin, ok := rightCond.(*ast.BinaryExpr); ok && bin.Op.Precedence() < token.LAND.Precedence() {
						rightCond = &ast.ParenExpr{X: rightCond}
					}

					combinedCond := &ast.BinaryExpr{
						X:  leftCond,
						Op: token.LAND,
						Y:  rightCond,
					}
					ifStmt.Cond = combinedCond
					ifStmt.Body = innerIf.Body
					modified = true
					return false
				}
			}
		}
		return true
	})

	if !modified {
		return src, false
	}

	return extractFunctionBody(fset, file)
}

// removeElseAfterReturn converts `if cond { return ... } else { body }` -> `if cond { return ... }\nbody`.
func removeElseAfterReturn(src string) (string, bool) {
	wrapped := fmt.Sprintf("package p\nfunc _() {\n%s\n}", src)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return src, false
	}

	modified := false
	ast.Inspect(file, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			var newStmts []ast.Stmt
			for _, stmt := range block.List {
				if ifStmt, ok := stmt.(*ast.IfStmt); ok && ifStmt.Else != nil {
					// Check if ifStmt.Body ends with a return statement
					if len(ifStmt.Body.List) > 0 {
						lastStmt := ifStmt.Body.List[len(ifStmt.Body.List)-1]
						if _, isReturn := lastStmt.(*ast.ReturnStmt); isReturn {
							if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
								ifStmt.Else = nil
								newStmts = append(newStmts, ifStmt)
								newStmts = append(newStmts, elseBlock.List...)
								modified = true
								continue
							}
						}
					}
				}
				newStmts = append(newStmts, stmt)
			}
			if modified {
				block.List = newStmts
			}
		}
		return true
	})

	if !modified {
		return src, false
	}

	return extractFunctionBody(fset, file)
}

// simplifyBooleanReturns simplifies `if cond { return true }; return false` -> `return cond`.
func simplifyBooleanReturns(src string) (string, bool) {
	wrapped := fmt.Sprintf("package p\nfunc _() bool {\n%s\n}", src)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return src, false
	}

	modified := false
	ast.Inspect(file, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			for i := 0; i < len(block.List)-1; i++ {
				ifStmt, isIf := block.List[i].(*ast.IfStmt)
				retStmt, isRet := block.List[i+1].(*ast.ReturnStmt)

				if isIf && isRet && ifStmt.Else == nil && ifStmt.Init == nil && len(ifStmt.Body.List) == 1 {
					innerRet, innerIsRet := ifStmt.Body.List[0].(*ast.ReturnStmt)
					if innerIsRet && len(innerRet.Results) == 1 && len(retStmt.Results) == 1 {
						innerIdent, innerIsIdent := innerRet.Results[0].(*ast.Ident)
						outerIdent, outerIsIdent := retStmt.Results[0].(*ast.Ident)

						if innerIsIdent && outerIsIdent {
							// Case 1: if cond { return true }; return false -> return cond
							if innerIdent.Name == "true" && outerIdent.Name == "false" {
								newRet := &ast.ReturnStmt{
									Return:  ifStmt.Pos(),
									Results: []ast.Expr{ifStmt.Cond},
								}
								block.List = append(block.List[:i], append([]ast.Stmt{newRet}, block.List[i+2:]...)...)
								modified = true
								return false
							}
							// Case 2: if cond { return false }; return true -> return !(cond)
							if innerIdent.Name == "false" && outerIdent.Name == "true" {
								negCond := ifStmt.Cond
								// Parenthesize complex / binary conditions to preserve precedence
								if _, isIdent := negCond.(*ast.Ident); !isIdent {
									negCond = &ast.ParenExpr{X: negCond}
								}
								notCond := &ast.UnaryExpr{
									Op: token.NOT,
									X:  negCond,
								}
								newRet := &ast.ReturnStmt{
									Return:  ifStmt.Pos(),
									Results: []ast.Expr{notCond},
								}
								block.List = append(block.List[:i], append([]ast.Stmt{newRet}, block.List[i+2:]...)...)
								modified = true
								return false
							}
						}
					}
				}
			}
		}
		return true
	})

	if !modified {
		return src, false
	}

	return extractFunctionBody(fset, file)
}

// isPureLeafExpr checks if an expression is a pure leaf expression (ident or basic literal)
func isPureLeafExpr(e ast.Expr) bool {
	switch e.(type) {
	case *ast.Ident, *ast.BasicLit:
		return true
	default:
		return false
	}
}

// exprString returns the formatted string representation of an AST expression
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	_ = format.Node(&buf, fset, expr)
	return buf.String()
}

// identUses checks if an AST node references an identifier by name.
func identUses(n ast.Node, name string) bool {
	if n == nil {
		return false
	}
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		if found {
			return false
		}
		if id, ok := node.(*ast.Ident); ok && id.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

// stmtMutatesIdent checks if a statement assigns, increments, or takes the address of an identifier.
func stmtMutatesIdent(stmt ast.Stmt, name string) bool {
	mutates := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if mutates {
			return false
		}
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == name {
					mutates = true
					return false
				}
			}
		case *ast.IncDecStmt:
			if id, ok := node.X.(*ast.Ident); ok && id.Name == name {
				mutates = true
				return false
			}
		case *ast.UnaryExpr:
			if node.Op == token.AND {
				if id, ok := node.X.(*ast.Ident); ok && id.Name == name {
					mutates = true
					return false
				}
			}
		}
		return true
	})
	return mutates
}

// nodeContainsRepr checks if a node contains an expression matching targetRepr.
func nodeContainsRepr(n ast.Node, targetRepr string, fset *token.FileSet) bool {
	contains := false
	ast.Inspect(n, func(node ast.Node) bool {
		if contains {
			return false
		}
		if expr, ok := node.(ast.Expr); ok {
			if exprString(fset, expr) == targetRepr {
				contains = true
				return false
			}
		}
		return true
	})
	return contains
}

// countReprOccurrences counts occurrences of targetRepr in a node.
func countReprOccurrences(n ast.Node, targetRepr string, fset *token.FileSet) int {
	count := 0
	ast.Inspect(n, func(node ast.Node) bool {
		if expr, ok := node.(ast.Expr); ok {
			if exprString(fset, expr) == targetRepr {
				count++
			}
		}
		return true
	})
	return count
}

// replaceSubexprInStmt replaces candidate expressions matching targetRepr with replacementIdent.
func replaceSubexprInStmt(stmt ast.Stmt, targetRepr, replacementIdent string, fset *token.FileSet) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		switch parent := n.(type) {
		case *ast.AssignStmt:
			for idx, rhs := range parent.Rhs {
				if exprString(fset, rhs) == targetRepr {
					parent.Rhs[idx] = ast.NewIdent(replacementIdent)
				}
			}
		case *ast.BinaryExpr:
			if exprString(fset, parent.X) == targetRepr {
				parent.X = ast.NewIdent(replacementIdent)
			}
			if exprString(fset, parent.Y) == targetRepr {
				parent.Y = ast.NewIdent(replacementIdent)
			}
		case *ast.ExprStmt:
			if exprString(fset, parent.X) == targetRepr {
				parent.X = ast.NewIdent(replacementIdent)
			}
		}
		return true
	})
}

// extractCommonSubexpressions extracts repeated pure arithmetic operations within the same basic block into a hoisted variable.
func extractCommonSubexpressions(src string) (string, bool) {
	wrapped := fmt.Sprintf("package p\nfunc _() {\n%s\n}", src)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return src, false
	}

	modified := false

	ast.Inspect(file, func(n ast.Node) bool {
		if modified {
			return false
		}
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}

		for i := 0; i < len(block.List); i++ {
			// Skip compound control flow statements so subexpressions are handled in their own nested blocks
			switch block.List[i].(type) {
			case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt:
				continue
			}

			var candidateExpr *ast.BinaryExpr
			var candidateRepr string
			var operandNames []string

			ast.Inspect(block.List[i], func(subNode ast.Node) bool {
				if candidateExpr != nil {
					return false
				}
				bin, isBin := subNode.(*ast.BinaryExpr)
				if !isBin {
					return true
				}
				// 1. Operator check: strictly non-panicking binary arithmetic/bitwise operators
				switch bin.Op {
				case token.ADD, token.SUB, token.MUL, token.AND, token.OR, token.XOR:
				default:
					return true // Rejects division (/), remainder (%), shifts, etc. to prevent panics!
				}
				// 2. Operands must be pure leaves (literals or identifiers)
				if !isPureLeafExpr(bin.X) || !isPureLeafExpr(bin.Y) {
					return true
				}
				// At least one operand must be an identifier
				var idents []string
				if id, ok := bin.X.(*ast.Ident); ok {
					idents = append(idents, id.Name)
				}
				if id, ok := bin.Y.(*ast.Ident); ok {
					idents = append(idents, id.Name)
				}
				if len(idents) == 0 {
					return true
				}

				candidateExpr = bin
				candidateRepr = exprString(fset, bin)
				operandNames = idents
				return false
			})

			if candidateExpr == nil {
				continue
			}

			// Check for subsequent occurrences in the SAME block
			secondIdx := -1
			interveningControlFlowOrMutation := false

			for j := i; j < len(block.List); j++ {
				if j > i {
					// Check for intervening control flow or branch statements
					switch block.List[j].(type) {
					case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.SelectStmt, *ast.BranchStmt, *ast.ReturnStmt:
						interveningControlFlowOrMutation = true
						break
					}
					if interveningControlFlowOrMutation {
						break
					}
					// Check if any operand is mutated
					for _, name := range operandNames {
						if stmtMutatesIdent(block.List[j], name) {
							interveningControlFlowOrMutation = true
							break
						}
					}
					if interveningControlFlowOrMutation {
						break
					}
				}

				if j > i && nodeContainsRepr(block.List[j], candidateRepr, fset) {
					secondIdx = j
					break
				} else if j == i && countReprOccurrences(block.List[i], candidateRepr, fset) >= 2 {
					secondIdx = i
					break
				}
			}

			if secondIdx != -1 && !interveningControlFlowOrMutation {
				// Hoist candidate expression immediately before statement i in the SAME block
				identName := "subExpr"
				hoistStmt := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(identName)},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{candidateExpr},
				}

				block.List = append(block.List[:i], append([]ast.Stmt{hoistStmt}, block.List[i:]...)...)

				// Replace occurrences in the statements at and after i+1
				for k := i + 1; k <= secondIdx+1; k++ {
					replaceSubexprInStmt(block.List[k], candidateRepr, identName, fset)
				}

				modified = true
				return false
			}
		}

		return true
	})

	if !modified {
		return src, false
	}

	return extractFunctionBody(fset, file)
}

// hoistLoopAllocations detects make([]T, ...) inside a for loop and hoists it IF provably non-escaping and zero-initializable.
func hoistLoopAllocations(src string) (string, bool) {
	wrapped := fmt.Sprintf("package p\nfunc _() {\n%s\n}", src)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return src, false
	}

	modified := false
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Body != nil {
			for i, stmt := range fn.Body.List {
				if forStmt, ok := stmt.(*ast.ForStmt); ok && forStmt.Body != nil {
					for j, loopStmt := range forStmt.Body.List {
						if assign, ok := loopStmt.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
							if id, ok := assign.Lhs[0].(*ast.Ident); ok {
								bufName := id.Name
								if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
									if makeIdent, ok := call.Fun.(*ast.Ident); ok && makeIdent.Name == "make" {
										// 1. ESCAPE ANALYSIS: Check if buffer escapes outside the iteration
										if checkBufferEscapes(forStmt, assign, bufName) {
											return false // Unsafe to hoist!
										}

										// 2. DYNAMIC SIZE CHECK: Check if allocation size depends on loop variables
										if len(call.Args) >= 2 && checkUsesLoopVar(forStmt, call.Args[1]) {
											return false // Unsafe to hoist!
										}

										// 3. Safe to hoist: buffer is purely local/temporary within iteration.
										// Hoist `buf := make(...)` before the for loop:
										hoistedAssign := &ast.AssignStmt{
											Lhs: assign.Lhs,
											Tok: token.DEFINE,
											Rhs: assign.Rhs,
										}
										// In the loop body, replace `buf := make(...)` with `clear(buf)`
										// to preserve Go's zero-initialization semantics for every iteration:
										clearStmt := &ast.ExprStmt{
											X: &ast.CallExpr{
												Fun:  ast.NewIdent("clear"),
												Args: []ast.Expr{ast.NewIdent(bufName)},
											},
										}
										forStmt.Body.List[j] = clearStmt
										fn.Body.List = append(fn.Body.List[:i], append([]ast.Stmt{hoistedAssign}, fn.Body.List[i:]...)...)
										modified = true
										return false
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	if !modified {
		return src, false
	}

	return extractFunctionBody(fset, file)
}

// checkBufferEscapes checks if bufName is assigned, escaped, captured, or passed to external code
func checkBufferEscapes(forStmt *ast.ForStmt, allocStmt ast.Stmt, bufName string) bool {
	escapes := false
	ast.Inspect(forStmt.Body, func(n ast.Node) bool {
		if escapes {
			return false
		}
		if n == allocStmt {
			return true // Ignore the original allocation statement
		}
		switch node := n.(type) {
		case *ast.ReturnStmt:
			for _, res := range node.Results {
				if identUses(res, bufName) {
					escapes = true
					return false
				}
			}
		case *ast.GoStmt:
			if identUses(node.Call, bufName) {
				escapes = true
				return false
			}
		case *ast.DeferStmt:
			if identUses(node.Call, bufName) {
				escapes = true
				return false
			}
		case *ast.SendStmt:
			if identUses(node.Value, bufName) || identUses(node.Chan, bufName) {
				escapes = true
				return false
			}
		case *ast.FuncLit:
			// Closure capturing bufName
			if identUses(node.Body, bufName) {
				escapes = true
				return false
			}
		case *ast.ForStmt:
			// Nested loop referencing bufName
			if node != forStmt && identUses(node, bufName) {
				escapes = true
				return false
			}
		case *ast.RangeStmt:
			// Nested range loop referencing bufName
			if identUses(node, bufName) {
				escapes = true
				return false
			}
		case *ast.AssignStmt:
			// Reassignment to buf: buf = ... or buf := ...
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == bufName {
					escapes = true
					return false
				}
			}
			// Assignment of buf to outer variables, struct fields, map keys, or interfaces
			for _, rhs := range node.Rhs {
				if identUses(rhs, bufName) {
					for _, lhs := range node.Lhs {
						// Discard is okay: `_ = buf`
						if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
							continue
						}
						// Index assignment into buf itself: `buf[i] = val` is okay
						if idxExpr, ok := lhs.(*ast.IndexExpr); ok {
							if id, ok := idxExpr.X.(*ast.Ident); ok && id.Name == bufName {
								continue
							}
						}
						// Otherwise it's assigned to an outer variable, struct, map, etc.
						escapes = true
						return false
					}
				}
			}
		case *ast.CallExpr:
			if identUses(node, bufName) {
				// Don't flag the initial make call itself
				if fnIdent, ok := node.Fun.(*ast.Ident); ok && fnIdent.Name == "make" {
					return true
				}
				// Safe builtins that do not retain references
				if fnIdent, ok := node.Fun.(*ast.Ident); ok {
					switch fnIdent.Name {
					case "len", "cap", "copy", "clear":
						return true
					}
				}
				// Any other function or method call escapes bufName
				escapes = true
				return false
			}
		case *ast.UnaryExpr:
			if node.Op == token.AND && identUses(node.X, bufName) {
				escapes = true
				return false
			}
		}
		return true
	})
	return escapes
}

// checkUsesLoopVar checks if sizeExpr refers to loop induction variables or variables mutated in the loop
func checkUsesLoopVar(forStmt *ast.ForStmt, sizeExpr ast.Expr) bool {
	loopVars := make(map[string]bool)
	// Induction vars from Init
	if forStmt.Init != nil {
		if assign, ok := forStmt.Init.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					loopVars[id.Name] = true
				}
			}
		}
	}
	// Any variables mutated in the loop body
	ast.Inspect(forStmt.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					loopVars[id.Name] = true
				}
			}
		case *ast.IncDecStmt:
			if id, ok := stmt.X.(*ast.Ident); ok {
				loopVars[id.Name] = true
			}
		}
		return true
	})

	uses := false
	ast.Inspect(sizeExpr, func(n ast.Node) bool {
		if uses {
			return false
		}
		switch e := n.(type) {
		case *ast.Ident:
			if loopVars[e.Name] {
				uses = true
				return false
			}
		case *ast.CallExpr:
			// Non-pure size expression!
			uses = true
			return false
		}
		return true
	})
	return uses
}

// extractFunctionBody serializes the AST body of the synthetic function wrapper back to a snippet.
func extractFunctionBody(fset *token.FileSet, file *ast.File) (string, bool) {
	var bodyBuf bytes.Buffer
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			for _, stmt := range fn.Body.List {
				if err := format.Node(&bodyBuf, fset, stmt); err != nil {
					return "", false
				}
				bodyBuf.WriteString("\n")
			}
		}
	}

	out := strings.TrimRight(bodyBuf.String(), "\n")
	return out, out != ""
}
