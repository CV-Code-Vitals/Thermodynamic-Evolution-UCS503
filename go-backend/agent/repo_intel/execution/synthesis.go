package execution

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// IntentType classifies the semantic repair goal.
type IntentType string

const (
	IntentResourceCleanup IntentType = "RESOURCE_CLEANUP"
	IntentHandleError     IntentType = "HANDLE_ERROR"
	IntentParameterizeSQL IntentType = "PARAMETERIZE_SQL"
	IntentAbstain         IntentType = "ABSTAIN"
)

// AbstainCategory specifies why autonomous synthesis was rejected.
type AbstainCategory string

const (
	AbstainCommandInjection    AbstainCategory = "COMMAND_INJECTION_ARCHITECTURAL_REDESIGN"
	AbstainLoopResource        AbstainCategory = "LOOP_DEFERRED_CLEANUP_ACCUMULATION"
	AbstainConcurrentEscape    AbstainCategory = "CONCURRENT_OWNERSHIP_ESCAPE"
	AbstainAmbiguousReturn     AbstainCategory = "AMBIGUOUS_RETURN_SIGNATURE"
	AbstainSyntaxError         AbstainCategory = "SOURCE_SYNTAX_ERROR"
	AbstainUnrecognizedPattern AbstainCategory = "UNRECOGNIZED_DEFECT_PATTERN"
)

// RepairIntent represents the parsed, goal-directed intent of a repair.
type RepairIntent struct {
	IntentType      IntentType      `json:"intent_type"`
	AbstainCategory AbstainCategory `json:"abstain_category,omitempty"`
	TargetFile      string          `json:"target_file"`
	TargetSymbol    string          `json:"target_symbol,omitempty"`
	LineNumber      int             `json:"line_number,omitempty"`
	Rationale       string          `json:"rationale"`
	Confidence      float64         `json:"confidence"`
}

// CandidatePatch holds a synthesized code modification.
type CandidatePatch struct {
	PatchID        string    `json:"patch_id"`
	TaskID         string    `json:"task_id"`
	TargetFile     string    `json:"target_file"`
	OriginalSource string    `json:"-"`
	ModifiedSource string    `json:"-"`
	UnifiedDiff    string    `json:"unified_diff"`
	DiffHash       string    `json:"diff_hash"`
	Mutations      []string  `json:"mutations"`
	CreatedAt      time.Time `json:"created_at"`
}

// SynthesisDecisionStatus is the disposition of a synthesis attempt.
type SynthesisDecisionStatus string

const (
	DecisionProposed  SynthesisDecisionStatus = "PROPOSED"
	DecisionAbstained SynthesisDecisionStatus = "ABSTAINED"
	DecisionFailed    SynthesisDecisionStatus = "FAILED"
)

// SynthesisDecision encapsulates the full result of autonomous patch synthesis.
type SynthesisDecision struct {
	Status    SynthesisDecisionStatus `json:"status"`
	Intent    *RepairIntent           `json:"intent"`
	Patch     *CandidatePatch         `json:"patch,omitempty"`
	Reason    string                  `json:"reason"`
	DurationMs int64                  `json:"duration_ms"`
}

// AutonomousPatchSynthesizer transforms semantic findings into verified AST-level patches.
type AutonomousPatchSynthesizer struct{}

// NewAutonomousPatchSynthesizer creates a new synthesizer instance.
func NewAutonomousPatchSynthesizer() *AutonomousPatchSynthesizer {
	return &AutonomousPatchSynthesizer{}
}

// Synthesize executes the bounded synthesis pipeline:
// SemanticFinding → RepairIntent → CandidatePatch → Decision
func (s *AutonomousPatchSynthesizer) Synthesize(
	repoRoot string,
	task *planning.RemediationTask,
	finding *repo_intel.Finding,
) *SynthesisDecision {
	start := time.Now()

	if task == nil {
		return &SynthesisDecision{
			Status:     DecisionFailed,
			Reason:     "Task cannot be nil",
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	if len(task.TargetFiles) == 0 {
		return &SynthesisDecision{
			Status:     DecisionAbstained,
			Reason:     "No target files specified in remediation task",
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	targetRel := task.TargetFiles[0]
	targetAbs := filepath.Join(repoRoot, filepath.FromSlash(targetRel))

	sourceBytes, err := os.ReadFile(targetAbs)
	if err != nil {
		return &SynthesisDecision{
			Status:     DecisionFailed,
			Reason:     fmt.Sprintf("Cannot read target file %s: %v", targetRel, err),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}
	originalSrc := string(sourceBytes)

	// Step 1: Infer Repair Intent from Semantic Context
	intent := s.inferIntent(task, finding, targetRel, originalSrc)
	if intent.IntentType == IntentAbstain {
		return &SynthesisDecision{
			Status:     DecisionAbstained,
			Intent:     intent,
			Reason:     intent.Rationale,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Step 2: Parse AST
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, targetAbs, originalSrc, parser.ParseComments)
	if err != nil {
		return &SynthesisDecision{
			Status: DecisionAbstained,
			Intent: &RepairIntent{
				IntentType:      IntentAbstain,
				AbstainCategory: AbstainSyntaxError,
				TargetFile:      targetRel,
				Rationale:       fmt.Sprintf("Parse error in source file: %v", err),
			},
			Reason:     fmt.Sprintf("Source file syntax error: %v", err),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Step 3: Perform AST Synthesis according to Intent
	var modifiedSrc string
	var mutations []string
	var synthErr error

	switch intent.IntentType {
	case IntentResourceCleanup:
		modifiedSrc, mutations, synthErr = s.synthesizeResourceCleanup(fset, node, intent, originalSrc)
	case IntentHandleError:
		modifiedSrc, mutations, synthErr = s.synthesizeErrorHandling(fset, node, intent, originalSrc)
	case IntentParameterizeSQL:
		modifiedSrc, mutations, synthErr = s.synthesizeSQLParameterization(fset, node, intent, originalSrc)
	default:
		synthErr = errors.New("unsupported intent type")
	}

	if synthErr != nil {
		var abstainErr *abstainError
		if errors.As(synthErr, &abstainErr) {
			return &SynthesisDecision{
				Status: DecisionAbstained,
				Intent: &RepairIntent{
					IntentType:      IntentAbstain,
					AbstainCategory: abstainErr.category,
					TargetFile:      targetRel,
					Rationale:       abstainErr.Error(),
				},
				Reason:     abstainErr.Error(),
				DurationMs: time.Since(start).Milliseconds(),
			}
		}

		return &SynthesisDecision{
			Status:     DecisionFailed,
			Intent:     intent,
			Reason:     fmt.Sprintf("Synthesis failed: %v", synthErr),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	if modifiedSrc == originalSrc || len(mutations) == 0 {
		return &SynthesisDecision{
			Status:     DecisionAbstained,
			Intent:     intent,
			Reason:     "AST transformation produced no changes to target file",
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Step 4: Verify Synthesized Code Compiles / Parses
	testFset := token.NewFileSet()
	if _, err := parser.ParseFile(testFset, targetAbs, modifiedSrc, parser.ParseComments); err != nil {
		return &SynthesisDecision{
			Status:     DecisionFailed,
			Intent:     intent,
			Reason:     fmt.Sprintf("Synthesized code failed syntax verification: %v", err),
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Step 5: Generate Unified Diff
	diffStr := formatUnifiedDiff(targetRel, originalSrc, modifiedSrc)
	hashBytes := sha256.Sum256([]byte(diffStr))
	diffHash := hex.EncodeToString(hashBytes[:])

	patch := &CandidatePatch{
		PatchID:        fmt.Sprintf("CANDIDATE-%s", task.TaskID),
		TaskID:         task.TaskID,
		TargetFile:     targetRel,
		OriginalSource: originalSrc,
		ModifiedSource: modifiedSrc,
		UnifiedDiff:    diffStr,
		DiffHash:       diffHash,
		Mutations:      mutations,
		CreatedAt:      time.Now(),
	}

	return &SynthesisDecision{
		Status:     DecisionProposed,
		Intent:     intent,
		Patch:      patch,
		Reason:     "Genuine AST patch synthesized and syntax-verified",
		DurationMs: time.Since(start).Milliseconds(),
	}
}

type abstainError struct {
	category AbstainCategory
	msg      string
}

func (e *abstainError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.msg)
}

func (s *AutonomousPatchSynthesizer) inferIntent(
	task *planning.RemediationTask,
	finding *repo_intel.Finding,
	targetRel string,
	src string,
) *RepairIntent {
	taskText := strings.ToLower(task.TaskID + " " + task.Title + " " + task.Description)
	if finding != nil {
		taskText += " " + strings.ToLower(finding.FindingID+" "+finding.Title+" "+finding.Description)
	}

	// Counterfactual / Abstention 1: Shell Command Injection
	if strings.Contains(taskText, "command") || strings.Contains(taskText, "cmd") || strings.Contains(taskText, "sh -c") {
		return &RepairIntent{
			IntentType:      IntentAbstain,
			AbstainCategory: AbstainCommandInjection,
			TargetFile:      targetRel,
			Rationale:       "Shell command execution requires manual architectural redesign to separate executable from arguments safely",
			Confidence:      1.0,
		}
	}

	// Intent: SQL Injection Parameterization
	if strings.Contains(taskText, "sql") || strings.Contains(taskText, "query") {
		return &RepairIntent{
			IntentType:   IntentParameterizeSQL,
			TargetFile:   targetRel,
			Rationale:    "Convert dynamic SQL string concatenation into parameterized query placeholders",
			Confidence:   0.95,
		}
	}

	// Intent: Resource Leak Cleanup
	if strings.Contains(taskText, "resource") || strings.Contains(taskText, "leak") || strings.Contains(taskText, "close") {
		return &RepairIntent{
			IntentType:   IntentResourceCleanup,
			TargetFile:   targetRel,
			Rationale:    "Insert deterministic deferred resource cleanup after allocation error check",
			Confidence:   0.95,
		}
	}

	// Intent: Unhandled Error Propagation
	if strings.Contains(taskText, "error") || strings.Contains(taskText, "unhandled") || strings.Contains(taskText, "err") {
		return &RepairIntent{
			IntentType:   IntentHandleError,
			TargetFile:   targetRel,
			Rationale:    "Insert error check and propagate error value to caller",
			Confidence:   0.90,
		}
	}

	return &RepairIntent{
		IntentType:      IntentAbstain,
		AbstainCategory: AbstainUnrecognizedPattern,
		TargetFile:      targetRel,
		Rationale:       "Finding pattern does not match supported deterministic repair intents",
		Confidence:      0.0,
	}
}

// -----------------------------------------------------------------------------------------
// Resource Leak Synthesis (Pattern 1)
// -----------------------------------------------------------------------------------------

func (s *AutonomousPatchSynthesizer) synthesizeResourceCleanup(
	fset *token.FileSet,
	node *ast.File,
	intent *RepairIntent,
	originalSrc string,
) (string, []string, error) {
	var mutations []string
	var loopAbstain, escapeAbstain bool
	var modified bool

	// Inspect all function bodies
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Check for concurrent escape across the function
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if _, isGo := n.(*ast.GoStmt); isGo {
				escapeAbstain = true
				return false
			}
			return true
		})
		if escapeAbstain {
			return "", nil, &abstainError{
				category: AbstainConcurrentEscape,
				msg:      "Resource handle escapes to concurrent goroutine",
			}
		}

		// Find resource allocation calls in statement list
		newList := make([]ast.Stmt, 0, len(fn.Body.List)+2)
		for i, stmt := range fn.Body.List {
			newList = append(newList, stmt)

			// Check if statement is inside a loop (for or range)
			if _, isFor := stmt.(*ast.ForStmt); isFor {
				ast.Inspect(stmt, func(n ast.Node) bool {
					if call, isCall := n.(*ast.CallExpr); isCall {
						if isResourceAllocCall(call) {
							loopAbstain = true
						}
					}
					return true
				})
			}
			if _, isRange := stmt.(*ast.RangeStmt); isRange {
				ast.Inspect(stmt, func(n ast.Node) bool {
					if call, isCall := n.(*ast.CallExpr); isCall {
						if isResourceAllocCall(call) {
							loopAbstain = true
						}
					}
					return true
				})
			}
			if loopAbstain {
				return "", nil, &abstainError{
					category: AbstainLoopResource,
					msg:      "Resource allocated inside loop body; deferred cleanup would cause handle accumulation",
				}
			}

			// Look for: resp, err := http.Get(...) OR f, err := os.Open(...)
			assign, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) < 1 || len(assign.Rhs) < 1 {
				continue
			}

			// Check if statement reassigns resp.Body = NopCloser(...) without closing prior (e.g. NAT-03)
			if len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
				if sel, isSel := assign.Lhs[0].(*ast.SelectorExpr); isSel && sel.Sel.Name == "Body" {
					if call, isCall := assign.Rhs[0].(*ast.CallExpr); isCall && isNopCloserCall(call) {
						if !hasClosePrior(fn.Body.List, i, "Body") {
							closeStmt := &ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   sel,
										Sel: ast.NewIdent("Close"),
									},
								},
							}
							newList = append(newList, closeStmt, stmt)
							varName := "resp"
							if id, ok := sel.X.(*ast.Ident); ok {
								varName = id.Name
							}
							mutations = append(mutations, fmt.Sprintf("Inserted '%s.Body.Close()' before reassigning body to NopCloser", varName))
							fn.Body.List = append(newList, fn.Body.List[i+1:]...)
							modified = true
							break
						}
					}
				}
			}

			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok {
				continue
			}

			resKind, isAlloc := detectResourceKind(call)
			if !isAlloc {
				continue
			}

			resIdent, ok := assign.Lhs[0].(*ast.Ident)
			if !ok {
				continue
			}
			resVar := resIdent.Name

			// Verify if the next statement is: if err != nil { return ... }
			if i+1 >= len(fn.Body.List) {
				continue
			}
			nextStmt := fn.Body.List[i+1]
			ifStmt, ok := nextStmt.(*ast.IfStmt)
			if !ok || !isErrorCheckIfStmt(ifStmt) {
				continue
			}

			// Check if cleanup already exists downstream
			alreadyCleaned := false
			for j := i + 2; j < len(fn.Body.List); j++ {
				if isCleanupStmt(fn.Body.List[j], resVar) {
					alreadyCleaned = true
					break
				}
			}
			if alreadyCleaned {
				continue
			}

			// Add the nextStmt (if err != nil)
			newList = append(newList, nextStmt)

			// Synthesize defer statement immediately after error check
			var deferStmt *ast.DeferStmt
			if resKind == "http_response" {
				// defer resp.Body.Close()
				deferStmt = &ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent(resVar),
								Sel: ast.NewIdent("Body"),
							},
							Sel: ast.NewIdent("Close"),
						},
					},
				}
				mutations = append(mutations, fmt.Sprintf("Inserted 'defer %s.Body.Close()' after error check", resVar))
			} else {
				// defer f.Close()
				deferStmt = &ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent(resVar),
							Sel: ast.NewIdent("Close"),
						},
					},
				}
				mutations = append(mutations, fmt.Sprintf("Inserted 'defer %s.Close()' after error check", resVar))
			}

			newList = append(newList, deferStmt)
			// Skip the next statement in loop since we already appended it
			fn.Body.List = append(newList, fn.Body.List[i+2:]...)
			modified = true
			break
		}

		if modified {
			cleanUnusedImport(node, "fmt")
			break
		}
	}

	if !modified {
		return originalSrc, nil, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", nil, fmt.Errorf("format modified AST: %w", err)
	}

	return buf.String(), mutations, nil
}

func isResourceAllocCall(call *ast.CallExpr) bool {
	_, alloc := detectResourceKind(call)
	return alloc
}

func detectResourceKind(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	pkgName := pkgIdent.Name
	funName := sel.Sel.Name

	if pkgName == "http" && (funName == "Get" || funName == "Post" || funName == "Do" || funName == "ReadResponse") {
		return "http_response", true
	}
	if pkgName == "os" && (funName == "Open" || funName == "Create" || funName == "OpenFile") {
		return "file", true
	}
	if pkgName == "net" && (funName == "Dial" || funName == "DialTimeout" || funName == "DialContext") {
		return "conn", true
	}
	if pkgName == "sql" && (funName == "Open") {
		return "db", true
	}
	if funName == "Query" || funName == "QueryContext" {
		return "rows", true
	}

	return "", false
}

func isErrorCheckIfStmt(ifStmt *ast.IfStmt) bool {
	binary, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if binary.Op != token.NEQ {
		return false
	}
	xIdent, okX := binary.X.(*ast.Ident)
	yIdent, okY := binary.Y.(*ast.Ident)
	if okX && okY {
		return (xIdent.Name == "err" && yIdent.Name == "nil") || (xIdent.Name == "nil" && yIdent.Name == "err")
	}
	return false
}

func isCleanupStmt(stmt ast.Stmt, varName string) bool {
	isClean := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Close" {
				isClean = true
				return false
			}
		}
		return true
	})
	return isClean
}

func isNopCloserCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "NopCloser" {
		return false
	}
	if ident, ok := sel.X.(*ast.Ident); ok {
		return ident.Name == "ioutil" || ident.Name == "io"
	}
	return false
}

func hasClosePrior(stmts []ast.Stmt, currentIndex int, fieldName string) bool {
	for j := 0; j < currentIndex; j++ {
		isClosed := false
		ast.Inspect(stmts[j], func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
					if innerSel, ok := sel.X.(*ast.SelectorExpr); ok && innerSel.Sel.Name == fieldName {
						isClosed = true
						return false
					}
				}
			}
			return true
		})
		if isClosed {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------------------
// Unhandled Error Synthesis (Pattern 2)
// -----------------------------------------------------------------------------------------

func (s *AutonomousPatchSynthesizer) synthesizeErrorHandling(
	fset *token.FileSet,
	node *ast.File,
	intent *RepairIntent,
	originalSrc string,
) (string, []string, error) {
	var mutations []string
	var modified bool

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Inspect function signature to ensure error is returned
		if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
			return "", nil, &abstainError{
				category: AbstainAmbiguousReturn,
				msg:      "Function does not return an error; cannot synthesize error propagation without altering public signature",
			}
		}

		// Check if last return type is 'error'
		lastResult := fn.Type.Results.List[len(fn.Type.Results.List)-1]
		lastIdent, ok := lastResult.Type.(*ast.Ident)
		if !ok || lastIdent.Name != "error" {
			return "", nil, &abstainError{
				category: AbstainAmbiguousReturn,
				msg:      "Enclosing function signature does not end in 'error' return type",
			}
		}

		// Build zero returns for preceding values
		zeroExprs := make([]ast.Expr, 0, len(fn.Type.Results.List))
		for idx := 0; idx < len(fn.Type.Results.List)-1; idx++ {
			retType := fn.Type.Results.List[idx].Type
			zeroExprs = append(zeroExprs, getZeroExpr(retType))
		}
		zeroExprs = append(zeroExprs, ast.NewIdent("err"))

		// Find statement discarding error: e.g. `_ = doSomething()` or `_, _ = doSomething()`
		newList := make([]ast.Stmt, 0, len(fn.Body.List)+2)
		for _, stmt := range fn.Body.List {
			// Check if statement discards error in assignment
			assign, isAssign := stmt.(*ast.AssignStmt)
			if isAssign && len(assign.Lhs) > 0 {
				lastLhs, isIdent := assign.Lhs[len(assign.Lhs)-1].(*ast.Ident)
				if isIdent && lastLhs.Name == "_" {
					// Replace blank identifier with 'err' and ensure it is declared
					lastLhs.Name = "err"
					assign.Tok = token.DEFINE

					// Insert `if err != nil { return ..., err }`
					ifReturn := &ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: zeroExprs,
								},
							},
						},
					}

					newList = append(newList, assign, ifReturn)
					mutations = append(mutations, "Replaced discarded error with 'err' and added error propagation check")
					modified = true
					continue
				}
			}

			// Check standalone expression call: e.g. `doSomething()`
			exprStmt, isExpr := stmt.(*ast.ExprStmt)
			if isExpr {
				if call, isCall := exprStmt.X.(*ast.CallExpr); isCall {
					// Convert `doSomething()` into `if err := doSomething(); err != nil { return ..., err }`
					ifStmt := &ast.IfStmt{
						Init: &ast.AssignStmt{
							Lhs: []ast.Expr{ast.NewIdent("err")},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{call},
						},
						Cond: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: zeroExprs,
								},
							},
						},
					}

					newList = append(newList, ifStmt)
					mutations = append(mutations, "Wrapped naked call with 'if err := ...; err != nil' propagation")
					modified = true
					continue
				}
			}

			newList = append(newList, stmt)
		}

		if modified {
			fn.Body.List = newList
			break
		}
	}

	if !modified {
		return originalSrc, nil, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", nil, fmt.Errorf("format modified AST: %w", err)
	}

	return buf.String(), mutations, nil
}

func getZeroExpr(t ast.Expr) ast.Expr {
	switch expr := t.(type) {
	case *ast.StarExpr:
		return ast.NewIdent("nil")
	case *ast.ArrayType:
		return ast.NewIdent("nil")
	case *ast.MapType:
		return ast.NewIdent("nil")
	case *ast.InterfaceType:
		return ast.NewIdent("nil")
	case *ast.ChanType:
		return ast.NewIdent("nil")
	case *ast.Ident:
		switch expr.Name {
		case "string":
			return &ast.BasicLit{Kind: token.STRING, Value: `""`}
		case "int", "int64", "int32", "uint", "uint64", "byte", "rune":
			return &ast.BasicLit{Kind: token.INT, Value: "0"}
		case "bool":
			return ast.NewIdent("false")
		case "error":
			return ast.NewIdent("nil")
		default:
			return ast.NewIdent("nil")
		}
	default:
		return ast.NewIdent("nil")
	}
}

// -----------------------------------------------------------------------------------------
// SQL Injection Parameterization Synthesis (Pattern 3)
// -----------------------------------------------------------------------------------------

func (s *AutonomousPatchSynthesizer) synthesizeSQLParameterization(
	fset *token.FileSet,
	node *ast.File,
	intent *RepairIntent,
	originalSrc string,
) (string, []string, error) {
	var mutations []string
	var modified bool

	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Look for `query := fmt.Sprintf("... '%s' ...", param)`
		var sprintfArgs []ast.Expr
		var queryVar string

		for _, stmt := range fn.Body.List {
			assign, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) == 0 || len(assign.Rhs) == 0 {
				continue
			}

			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok {
				// Check binary concatenation e.g. "SELECT ... = '" + tag + "'"
				if bin, isBin := assign.Rhs[0].(*ast.BinaryExpr); isBin && bin.Op == token.ADD {
					parts := flattenBinaryAdd(bin)
					if len(parts) >= 2 {
						if firstLit, ok := parts[0].(*ast.BasicLit); ok && firstLit.Kind == token.STRING {
							val := firstLit.Value
							upperVal := strings.ToUpper(val)
							if strings.Contains(upperVal, "SELECT") || strings.Contains(upperVal, "INSERT") || strings.Contains(upperVal, "UPDATE") || strings.Contains(upperVal, "DELETE") {
								cleanVal := strings.TrimSuffix(val, "'\"")
								cleanVal = strings.TrimSuffix(cleanVal, "'")
								cleanVal = strings.TrimSuffix(cleanVal, " = '")
								cleanVal = strings.TrimSuffix(cleanVal, "=\"")
								cleanVal = strings.TrimSuffix(cleanVal, "\"")
								newQueryStr := cleanVal + " = ?\""
								if !strings.HasPrefix(newQueryStr, "\"") {
									newQueryStr = "\"" + newQueryStr
								}

								assign.Rhs[0] = &ast.BasicLit{
									Kind:  token.STRING,
									Value: newQueryStr,
								}

								identLhs, ok := assign.Lhs[0].(*ast.Ident)
								if ok {
									queryVar = identLhs.Name
								}

								sprintfArgs = []ast.Expr{parts[1]}
								mutations = append(mutations, fmt.Sprintf("Replaced dynamic concatenated SQL query with parameterized literal %s", newQueryStr))
								modified = true
								break
							}
						}
					}
				}
				continue
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Sprintf" {
				continue
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok || pkgIdent.Name != "fmt" {
				continue
			}

			if len(call.Args) < 2 {
				continue
			}

			// First argument must be basic lit format string
			fmtLit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || fmtLit.Kind != token.STRING {
				continue
			}

			queryStr := fmtLit.Value // Includes quotes
			if !strings.Contains(queryStr, "%s") && !strings.Contains(queryStr, "%d") {
				continue
			}

			// Parameterize format string: replace '%s' with ? and %s with ? and %d with ?
			newQueryStr := queryStr
			newQueryStr = strings.ReplaceAll(newQueryStr, "'%s'", "?")
			newQueryStr = strings.ReplaceAll(newQueryStr, "\"%s\"", "?")
			newQueryStr = strings.ReplaceAll(newQueryStr, "%s", "?")
			newQueryStr = strings.ReplaceAll(newQueryStr, "%d", "?")

			// Mutate assignment into static string literal
			fmtLit.Value = newQueryStr
			assign.Rhs[0] = fmtLit

			identLhs, ok := assign.Lhs[0].(*ast.Ident)
			if ok {
				queryVar = identLhs.Name
			}
			sprintfArgs = call.Args[1:]

			mutations = append(mutations, fmt.Sprintf("Replaced dynamic fmt.Sprintf query with parameterized literal %s", newQueryStr))
			modified = true
			break
		}

		if modified && queryVar != "" && len(sprintfArgs) > 0 {
			// Find downstream DB query call (e.g. `db.ExecContext(ctx, query)`) and pass args
			for _, stmt := range fn.Body.List {
				ast.Inspect(stmt, func(n ast.Node) bool {
					call, isCall := n.(*ast.CallExpr)
					if !isCall {
						return true
					}

					sel, isSel := call.Fun.(*ast.SelectorExpr)
					if !isSel {
						return true
					}

					fnName := sel.Sel.Name
					if fnName == "Exec" || fnName == "ExecContext" || fnName == "Query" || fnName == "QueryContext" || fnName == "QueryRow" || fnName == "QueryRowContext" {
						// Check if queryVar is passed
						hasQueryVar := false
						for _, arg := range call.Args {
							if ident, ok := arg.(*ast.Ident); ok && ident.Name == queryVar {
								hasQueryVar = true
								break
							}
						}

						if hasQueryVar {
							// Append the parameters to DB call
							call.Args = append(call.Args, sprintfArgs...)
							mutations = append(mutations, fmt.Sprintf("Appended %d bound parameters to %s call", len(sprintfArgs), fnName))
						}
					}
					return true
				})
			}
		}

		if modified {
			cleanUnusedImport(node, "fmt")
			break
		}
	}

	if !modified {
		return originalSrc, nil, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", nil, fmt.Errorf("format modified AST: %w", err)
	}

	return buf.String(), mutations, nil
}

func cleanUnusedImport(node *ast.File, pkgName string) {
	isUsed := false
	ast.Inspect(node, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == pkgName {
				isUsed = true
				return false
			}
		}
		return true
	})

	if isUsed {
		return
	}

	quoted := fmt.Sprintf("%q", pkgName)
	newDecls := make([]ast.Decl, 0, len(node.Decls))
	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			newDecls = append(newDecls, decl)
			continue
		}

		newSpecs := make([]ast.Spec, 0, len(gen.Specs))
		for _, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if !ok {
				newSpecs = append(newSpecs, spec)
				continue
			}
			if imp.Path.Value == quoted {
				continue // Drop unused import
			}
			newSpecs = append(newSpecs, spec)
		}
		if len(newSpecs) > 0 {
			gen.Specs = newSpecs
			newDecls = append(newDecls, gen)
		}
	}
	node.Decls = newDecls

	newImports := make([]*ast.ImportSpec, 0, len(node.Imports))
	for _, imp := range node.Imports {
		if imp.Path.Value != quoted {
			newImports = append(newImports, imp)
		}
	}
	node.Imports = newImports
}

func flattenBinaryAdd(bin *ast.BinaryExpr) []ast.Expr {
	var list []ast.Expr
	if leftBin, ok := bin.X.(*ast.BinaryExpr); ok && leftBin.Op == token.ADD {
		list = append(list, flattenBinaryAdd(leftBin)...)
	} else {
		list = append(list, bin.X)
	}
	list = append(list, bin.Y)
	return list
}
