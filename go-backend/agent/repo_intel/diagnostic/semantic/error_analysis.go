package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ErrorAnalyzer performs AST and CFG-aware inspection of error-returning calls and their surrounding control flow.
type ErrorAnalyzer struct {
	repoRoot   string
	registry   *ErrorRegistry
	localFuncs map[string]*ast.FuncDecl
}

// NewErrorAnalyzer constructs an ErrorAnalyzer instance.
func NewErrorAnalyzer(repoRoot string) *ErrorAnalyzer {
	return &ErrorAnalyzer{
		repoRoot:   repoRoot,
		registry:   NewErrorRegistry(),
		localFuncs: make(map[string]*ast.FuncDecl),
	}
}

// RegisterLocalFunction records local function declarations for package-level signature resolution.
func (a *ErrorAnalyzer) RegisterLocalFunction(fnDecl *ast.FuncDecl) {
	if fnDecl != nil && fnDecl.Name != nil {
		a.localFuncs[fnDecl.Name.Name] = fnDecl
	}
}

// AnalyzeFunction analyzes a single function declaration for error handling completeness.
func (a *ErrorAnalyzer) AnalyzeFunction(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	df *ReachingDefinitionsAnalysis,
	filePath string,
	typesInfo *types.Info,
) ([]ErrorEvidence, error) {
	if fnDecl == nil || fnDecl.Body == nil {
		return nil, nil
	}

	pkgName := ""
	fnName := fnDecl.Name.Name

	// Discover named return error variable if any
	namedErrorVar := a.extractNamedErrorReturn(fnDecl)

	var evidences []ErrorEvidence
	instCounter := 1

	// Walk function body to discover call sites
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sig, cat := a.registry.ResolveCallSignature(call, typesInfo, a.localFuncs)
		if !sig.IsErrorReturning {
			return true
		}

		callPos := fset.Position(call.Pos())
		callSnippet := exprToString(call, fset)
		blockID := a.findBlockIDForPos(call.Pos(), cfg)

		inst := ErrorCallInstance{
			ID:               instCounter,
			CallSnippet:      callSnippet,
			Line:             callPos.Line,
			BlockID:          blockID,
			Package:          pkgName,
			FunctionOrMethod: callExprToString(call.Fun),
			Signature:        sig,
			Category:         cat,
			History:          make([]string, 0),
			BranchEvidence:   make([]string, 0),
		}
		instCounter++

		// Analyze the context and usage of this error-returning call
		a.classifyCallSite(&inst, call, fnDecl, fset, cfg, namedErrorVar)

		ev := a.constructEvidence(&inst, filePath, pkgName, fnName)
		evidences = append(evidences, ev)

		return true
	})

	return evidences, nil
}

// extractNamedErrorReturn checks if the function has a named return parameter of type error.
func (a *ErrorAnalyzer) extractNamedErrorReturn(fnDecl *ast.FuncDecl) string {
	if fnDecl.Type == nil || fnDecl.Type.Results == nil {
		return ""
	}
	for _, field := range fnDecl.Type.Results.List {
		typeStr := typeExprToString(field.Type)
		if typeStr == "error" && len(field.Names) > 0 {
			return field.Names[0].Name
		}
	}
	return ""
}

// classifyCallSite analyzes how the error from a call is received, checked, or discarded.
func (a *ErrorAnalyzer) classifyCallSite(
	inst *ErrorCallInstance,
	call *ast.CallExpr,
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	namedErrorVar string,
) {
	// 1. Check if call is directly inside a ReturnStmt
	if a.isCallDirectlyReturned(call, fnDecl) {
		inst.Classification = ErrorPropagated
		inst.History = append(inst.History, "Direct error return to caller")
		return
	}

	// 2. Check if call is an init statement in an IfStmt: if err := call(); err != nil
	if ifStmt := a.findContainingIfInit(call, fnDecl); ifStmt != nil {
		a.evaluateIfInitCall(inst, call, ifStmt, fset)
		return
	}

	// 3. Check if call is inside an AssignStmt
	if assignStmt := a.findContainingAssignStmt(call, fnDecl); assignStmt != nil {
		a.evaluateAssignStmt(inst, call, assignStmt, fnDecl, fset, cfg, namedErrorVar)
		return
	}

	// 4. Check if call is inside an ExprStmt (unassigned bare expression)
	if a.isContainingExprStmt(call, fnDecl) {
		a.evaluateExprStmt(inst, call, fnDecl, fset)
		return
	}

	// Fallback to unresolved
	inst.Classification = ErrorUnresolved
}

// isCallDirectlyReturned checks if call is directly an operand of a ReturnStmt.
func (a *ErrorAnalyzer) isCallDirectlyReturned(call *ast.CallExpr, fnDecl *ast.FuncDecl) bool {
	isRet := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		retStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, res := range retStmt.Results {
			if res == call {
				isRet = true
				return false
			}
		}
		return true
	})
	return isRet
}

// findContainingIfInit finds an IfStmt whose Init statement contains call.
func (a *ErrorAnalyzer) findContainingIfInit(call *ast.CallExpr, fnDecl *ast.FuncDecl) *ast.IfStmt {
	var found *ast.IfStmt
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if ifStmt.Init != nil {
			ast.Inspect(ifStmt.Init, func(in ast.Node) bool {
				if in == call {
					found = ifStmt
					return false
				}
				return true
			})
		}
		return found == nil
	})
	return found
}

// evaluateIfInitCall classifies an error returned in an if-init clause (if err := do(); err != nil).
func (a *ErrorAnalyzer) evaluateIfInitCall(
	inst *ErrorCallInstance,
	call *ast.CallExpr,
	ifStmt *ast.IfStmt,
	fset *token.FileSet,
) {
	condStr := exprToString(ifStmt.Cond, fset)
	if strings.Contains(condStr, "!= nil") {
		// Check if body wraps or returns
		bodyStr := stmtToString(ifStmt.Body, fset)
		if strings.Contains(bodyStr, "%w") {
			inst.Classification = ErrorWrappedAndPropagated
			inst.History = append(inst.History, "Error checked in if-init condition and wrapped in body")
		} else {
			inst.Classification = ErrorChecked
			inst.History = append(inst.History, "Error checked in if-init condition")
		}
	} else {
		inst.Classification = ErrorUnhandled
		inst.History = append(inst.History, "Error assigned in if-init but not tested in condition")
	}
}

// findContainingAssignStmt locates the AssignStmt containing call on its Rhs.
func (a *ErrorAnalyzer) findContainingAssignStmt(call *ast.CallExpr, fnDecl *ast.FuncDecl) *ast.AssignStmt {
	var found *ast.AssignStmt
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if rhs == call {
				found = assign
				return false
			}
		}
		return found == nil
	})
	return found
}

// evaluateAssignStmt analyzes the LHS assignment target, checking for blank discard, reaching defs, and overwrites.
func (a *ErrorAnalyzer) evaluateAssignStmt(
	inst *ErrorCallInstance,
	call *ast.CallExpr,
	assignStmt *ast.AssignStmt,
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	namedErrorVar string,
) {
	errIdx := inst.Signature.ErrorIndex
	if errIdx >= len(assignStmt.Lhs) {
		errIdx = len(assignStmt.Lhs) - 1
	}

	lhsTarget := assignStmt.Lhs[errIdx]
	targetIdent, isIdent := lhsTarget.(*ast.Ident)

	if !isIdent {
		inst.Classification = ErrorUnresolved
		return
	}

	// 1. Blank identifier assignment: _ = call() or val, _ := call()
	if targetIdent.Name == "_" {
		inst.IsBlankDiscard = true
		inst.AssignedVar = "_"
		inst.Classification = ErrorExplicitlyIgnored
		inst.History = append(inst.History, "Error explicitly discarded via blank identifier '_'")
		return
	}

	varName := targetIdent.Name
	inst.AssignedVar = varName
	assignLine := fset.Position(assignStmt.Pos()).Line

	// 2. Check if varName was overwritten before being checked
	if a.isVariableOverwrittenBeforeCheck(fnDecl, varName, assignLine, fset) {
		inst.OverwrittenBefore = true
		inst.Classification = ErrorUnhandled
		inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' was overwritten by subsequent assignment before being checked", varName))
		return
	}

	// 3. Inspect subsequent if statements for checks of varName
	checked, wrapped, handlesGracefully := a.inspectErrorConditionCheck(fnDecl, varName, assignLine, fset)
	if checked {
		if wrapped {
			inst.Classification = ErrorWrappedAndPropagated
			inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' checked and wrapped with %%w", varName))
		} else if handlesGracefully {
			inst.Classification = ErrorChecked
			inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' checked in if condition with terminating or handling branch", varName))
		} else {
			inst.Classification = ErrorPartiallyHandled
			inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' checked in branch but execution continues unsafely", varName))
		}
		return
	}

	// 4. Check for test helper assertion or check helper: e.g. assert.NoError(t, err) or check(err)
	if a.isPassedToErrorHelper(fnDecl, varName, assignLine, fset) {
		inst.Classification = ErrorChecked
		inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' checked via recognized assertion or helper", varName))
		return
	}

	// 5. Named return variable check
	if namedErrorVar != "" && varName == namedErrorVar {
		if a.hasReturnAfterLine(fnDecl, assignLine, fset) {
			inst.Classification = ErrorPropagated
			inst.History = append(inst.History, fmt.Sprintf("Error assigned to named return variable '%s' and returned", varName))
			return
		}
	}

	// 6. Check for direct unconditional return of varName: return err or return nil, err
	if a.isVariableReturnedDirectly(fnDecl, varName, assignLine, fset) {
		inst.Classification = ErrorPropagated
		inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' returned directly", varName))
		return
	}

	// Not checked anywhere
	inst.Classification = ErrorUnhandled
	inst.History = append(inst.History, fmt.Sprintf("Error variable '%s' assigned at line %d but never checked or returned", varName, assignLine))
}

// isContainingExprStmt checks if call is the root expression of an ExprStmt.
func (a *ErrorAnalyzer) isContainingExprStmt(call *ast.CallExpr, fnDecl *ast.FuncDecl) bool {
	isExpr := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		exprStmt, ok := n.(*ast.ExprStmt)
		if ok && exprStmt.X == call {
			isExpr = true
			return false
		}
		return true
	})
	return isExpr
}

// evaluateExprStmt analyzes unassigned call statements (json.Unmarshal(data, &res)).
func (a *ErrorAnalyzer) evaluateExprStmt(
	inst *ErrorCallInstance,
	call *ast.CallExpr,
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
) {
	inst.IsUnassignedExpr = true

	// Check if enclosed in defer
	isDeferred := a.isCallEnclosedInDefer(call, fnDecl)

	if isDeferred {
		if inst.Category == ErrorCatCleanup {
			inst.Classification = ErrorExplicitlyIgnored
			inst.History = append(inst.History, "Cleanup call deferred without capturing error (standard idiom)")
		} else {
			inst.Classification = ErrorUnhandled
			inst.History = append(inst.History, "Non-cleanup error-producing call deferred without error handling")
		}
		return
	}

	if inst.Category == ErrorCatCleanup {
		inst.Classification = ErrorUnhandled
		inst.History = append(inst.History, "Cleanup error return value unhandled in bare expression statement")
	} else {
		inst.Classification = ErrorUnhandled
		inst.History = append(inst.History, "Critical operation error return unassigned and discarded")
	}
}

// isCallEnclosedInDefer checks if call is directly or indirectly executed within a DeferStmt.
func (a *ErrorAnalyzer) isCallEnclosedInDefer(call *ast.CallExpr, fnDecl *ast.FuncDecl) bool {
	isDef := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		defStmt, ok := n.(*ast.DeferStmt)
		if !ok {
			return true
		}
		ast.Inspect(defStmt, func(dn ast.Node) bool {
			if dn == call {
				isDef = true
				return false
			}
			return true
		})
		return !isDef
	})
	return isDef
}

// isVariableOverwrittenBeforeCheck verifies if varName is reassigned after startLine before any if check.
func (a *ErrorAnalyzer) isVariableOverwrittenBeforeCheck(
	fnDecl *ast.FuncDecl,
	varName string,
	startLine int,
	fset *token.FileSet,
) bool {
	firstCheckLine := 999999
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		ifPos := fset.Position(ifStmt.Pos())
		if ifPos.Line > startLine {
			condStr := exprToString(ifStmt.Cond, fset)
			if strings.Contains(condStr, varName) && strings.Contains(condStr, "!= nil") {
				if ifPos.Line < firstCheckLine {
					firstCheckLine = ifPos.Line
				}
			}
		}
		return true
	})

	isOverwritten := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		pos := fset.Position(assign.Pos())
		if pos.Line > startLine && pos.Line < firstCheckLine {
			for _, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == varName {
					isOverwritten = true
					return false
				}
			}
		}
		return true
	})

	return isOverwritten
}

// inspectErrorConditionCheck verifies whether varName is tested with != nil and whether the branch terminates/wraps.
func (a *ErrorAnalyzer) inspectErrorConditionCheck(
	fnDecl *ast.FuncDecl,
	varName string,
	startLine int,
	fset *token.FileSet,
) (checked bool, wrapped bool, handlesGracefully bool) {
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		pos := fset.Position(ifStmt.Pos())
		if pos.Line <= startLine {
			return true
		}

		condStr := exprToString(ifStmt.Cond, fset)
		if strings.Contains(condStr, varName) && (strings.Contains(condStr, "!= nil") || strings.Contains(condStr, "!=nil")) {
			checked = true

			hasTerminator := false
			hasWrap := false

			ast.Inspect(ifStmt.Body, func(bn ast.Node) bool {
				switch bnode := bn.(type) {
				case *ast.ReturnStmt:
					hasTerminator = true
				case *ast.CallExpr:
					callStr := exprToString(bnode.Fun, fset)
					if strings.Contains(callStr, "Fatal") || strings.Contains(callStr, "panic") {
						hasTerminator = true
					}
				case *ast.BasicLit:
					if strings.Contains(bnode.Value, "%w") {
						hasWrap = true
					}
				}
				return true
			})

			if hasWrap {
				wrapped = true
				handlesGracefully = true
			} else if hasTerminator {
				handlesGracefully = true
			} else {
				handlesGracefully = false
			}
			return false
		}
		return true
	})

	return checked, wrapped, handlesGracefully
}

// isVariableReturnedDirectly checks if varName is an unconditional operand of a ReturnStmt.
func (a *ErrorAnalyzer) isVariableReturnedDirectly(
	fnDecl *ast.FuncDecl,
	varName string,
	startLine int,
	fset *token.FileSet,
) bool {
	returned := false
	for _, stmt := range fnDecl.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		pos := fset.Position(retStmt.Pos())
		if pos.Line < startLine {
			continue
		}
		for _, res := range retStmt.Results {
			if id, ok := res.(*ast.Ident); ok && id.Name == varName {
				returned = true
				break
			}
		}
	}
	return returned
}

// hasReturnAfterLine checks if any ReturnStmt exists after startLine.
func (a *ErrorAnalyzer) hasReturnAfterLine(fnDecl *ast.FuncDecl, startLine int, fset *token.FileSet) bool {
	hasRet := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		retStmt, ok := n.(*ast.ReturnStmt)
		if ok && fset.Position(retStmt.Pos()).Line >= startLine {
			hasRet = true
			return false
		}
		return true
	})
	return hasRet
}

// isPassedToErrorHelper checks if varName is passed to an assertion or recognized error handler.
func (a *ErrorAnalyzer) isPassedToErrorHelper(
	fnDecl *ast.FuncDecl,
	varName string,
	startLine int,
	fset *token.FileSet,
) bool {
	isHelper := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		pos := fset.Position(call.Pos())
		if pos.Line < startLine {
			return true
		}
		fnStr := exprToString(call.Fun, fset)
		if strings.Contains(fnStr, "NoError") || strings.Contains(fnStr, "check") || strings.Contains(fnStr, "assert") || strings.Contains(fnStr, "require") {
			for _, arg := range call.Args {
				if id, ok := arg.(*ast.Ident); ok && id.Name == varName {
					isHelper = true
					return false
				}
			}
		}
		return true
	})
	return isHelper
}

// findBlockIDForPos locates the basic block ID containing the pos token.
func (a *ErrorAnalyzer) findBlockIDForPos(pos token.Pos, cfg *ControlFlowGraph) int {
	if cfg == nil {
		return 0
	}
	for _, b := range cfg.Blocks {
		for _, stmt := range b.Statements {
			if stmt.Pos() <= pos && pos <= stmt.End() {
				return b.ID
			}
		}
	}
	if cfg.Entry != nil {
		return cfg.Entry.ID
	}
	return 0
}

// constructEvidence synthesizes the final ErrorEvidence record based on classification and category.
func (a *ErrorAnalyzer) constructEvidence(
	inst *ErrorCallInstance,
	filePath, pkgName, fnName string,
) ErrorEvidence {
	ev := ErrorEvidence{
		FilePath:       filePath,
		Package:        pkgName,
		Function:       fnName,
		Line:           inst.Line,
		CallSnippet:    inst.CallSnippet,
		Signature:      inst.Signature,
		Classification: inst.Classification,
		Category:       inst.Category,
		BranchEvidence: inst.BranchEvidence,
		Timestamp:      time.Now(),
	}

	switch inst.Classification {
	case ErrorUnhandled:
		ev.FindingState = StateConfirmed
		ev.Confidence = 0.96
		if inst.Category == ErrorCatCleanup {
			ev.Severity = repo_intel.SeverityLow
			ev.Title = fmt.Sprintf("Unchecked Cleanup Error (%s)", inst.FunctionOrMethod)
			ev.Description = fmt.Sprintf("Function '%s' at line %d invokes cleanup call '%s' without checking returned error.",
				fnName, inst.Line, inst.CallSnippet)
			ev.RootCause = "Cleanup operation return error ignored."
			ev.RecommendedAction = "Inspect returned error and log if operation fails."
		} else {
			ev.Severity = repo_intel.SeverityHigh
			ev.Title = fmt.Sprintf("Unhandled Error Return (%s)", inst.FunctionOrMethod)
			ev.Description = fmt.Sprintf("Function '%s' at line %d invokes error-producing call '%s' but the returned error is discarded.",
				fnName, inst.Line, inst.CallSnippet)
			ev.RootCause = "Error return value ignored without handling, logging, or propagation."
			ev.RecommendedAction = "Assign returned error and handle with 'if err != nil { return ... }'."
		}

	case ErrorExplicitlyIgnored:
		if inst.Category == ErrorCatCleanup {
			ev.FindingState = StateCounterEvidencePresent
			ev.Severity = repo_intel.SeverityLow
			ev.Confidence = 0.95
			ev.Title = fmt.Sprintf("Safe Explicit Cleanup Error Suppression (%s)", inst.FunctionOrMethod)
			ev.Description = fmt.Sprintf("Cleanup error explicitly suppressed via blank identifier in '%s'.", inst.CallSnippet)
			ev.CounterEvidence = []string{"Explicit suppression of cleanup operation error adheres to standard idiom"}
		} else {
			ev.FindingState = StateConfirmed
			ev.Severity = repo_intel.SeverityHigh
			ev.Confidence = 0.95
			ev.Title = fmt.Sprintf("Explicitly Discarded Critical Error (%s)", inst.FunctionOrMethod)
			ev.Description = fmt.Sprintf("Function '%s' at line %d explicitly discards error from critical operation '%s' via blank identifier '_'.",
				fnName, inst.Line, inst.CallSnippet)
			ev.RootCause = "Critical operation error suppressed with blank identifier."
			ev.RecommendedAction = "Handle or propagate the error rather than suppressing with '_'."
		}

	case ErrorPartiallyHandled:
		ev.FindingState = StateProbable
		ev.Severity = repo_intel.SeverityMedium
		ev.Confidence = 0.88
		ev.Title = fmt.Sprintf("Partially Handled Error Return (%s)", inst.FunctionOrMethod)
		ev.Description = fmt.Sprintf("Error from '%s' at line %d is tested in branch but execution continues unsafely.",
			inst.CallSnippet, inst.Line)
		ev.RootCause = "Error check branch does not terminate or safely handle the error condition."
		ev.RecommendedAction = "Ensure the error branch returns, propagates, or terminates control flow."

	case ErrorChecked:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.98
		ev.Title = "Verified Error Check"
		ev.Description = fmt.Sprintf("Error from '%s' is checked and handled appropriately.", inst.CallSnippet)
		ev.CounterEvidence = []string{"Error checked with conditional branch and handled/returned"}

	case ErrorPropagated:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.98
		ev.Title = "Verified Error Propagation"
		ev.Description = fmt.Sprintf("Error from '%s' is directly propagated to caller.", inst.CallSnippet)
		ev.CounterEvidence = []string{"Error returned to caller"}

	case ErrorWrappedAndPropagated:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.99
		ev.Title = "Verified Wrapped Error Propagation"
		ev.Description = fmt.Sprintf("Error from '%s' is checked and wrapped with %%w before return.", inst.CallSnippet)
		ev.CounterEvidence = []string{"Error wrapped with %w and returned"}

	default:
		ev.FindingState = StateInsufficientEvidence
		ev.Confidence = 0.60
		ev.Title = "Unresolved Error Handling"
		ev.Description = fmt.Sprintf("Error return handling for '%s' could not be statically determined.", inst.CallSnippet)
		ev.AnalysisLimitations = []string{"Complex control flow or unverified error sink"}
	}

	return ev
}

func stmtToString(stmt ast.Stmt, fset *token.FileSet) string {
	if stmt == nil {
		return ""
	}
	var sb strings.Builder
	ast.Inspect(stmt, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			sb.WriteString(id.Name)
			sb.WriteString(" ")
		}
		if lit, ok := n.(*ast.BasicLit); ok {
			sb.WriteString(lit.Value)
			sb.WriteString(" ")
		}
		return true
	})
	return sb.String()
}
