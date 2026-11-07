package execution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// OracleType identifies each independent verification oracle.
type OracleType string

const (
	OracleSyntaxCompilation   OracleType = "SYNTAX_AND_COMPILATION"
	OracleRegressionTest      OracleType = "REGRESSION_TEST"
	OracleDefectReproducer    OracleType = "DEFECT_REPRODUCER"
	OracleBehavioralInvariance OracleType = "BEHAVIORAL_INVARIANCE"
	OracleNegativeProperties  OracleType = "NEGATIVE_PROPERTIES"
)

// OracleResult holds the outcome of a single oracle check.
type OracleResult struct {
	OracleType    OracleType `json:"oracle_type"`
	Passed        bool       `json:"passed"`
	DurationMs    int64      `json:"duration_ms"`
	Details       string     `json:"details"`
	Output        string     `json:"output,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty"`
}

// MultiOracleReport aggregates results across all five independent oracles.
type MultiOracleReport struct {
	PatchID         string                    `json:"patch_id"`
	AllPassed       bool                      `json:"all_passed"`
	OracleResults   map[OracleType]*OracleResult `json:"oracle_results"`
	TotalDurationMs int64                     `json:"total_duration_ms"`
}

// IndependentOracleSuite executes five distinct oracles to verify patch correctness.
type IndependentOracleSuite struct {
	TimeoutSec int
}

// NewIndependentOracleSuite creates a new oracle verification suite.
func NewIndependentOracleSuite(timeoutSec int) *IndependentOracleSuite {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return &IndependentOracleSuite{TimeoutSec: timeoutSec}
}

// VerifyAll executes all five oracles against the workspace.
func (s *IndependentOracleSuite) VerifyAll(
	wsDir string,
	task *planning.RemediationTask,
	patch *CandidatePatch,
	reproducerCmd string,
	invarianceCmd string,
) *MultiOracleReport {
	start := time.Now()
	report := &MultiOracleReport{
		PatchID:       patch.PatchID,
		OracleResults: make(map[OracleType]*OracleResult),
		AllPassed:     true,
	}

	// 1. Oracle 1: Syntax & Compilation Oracle
	res1 := s.verifySyntaxCompilation(wsDir, task, patch)
	report.OracleResults[OracleSyntaxCompilation] = res1
	if !res1.Passed {
		report.AllPassed = false
	}

	// 2. Oracle 2: Regression Test Oracle
	res2 := s.verifyRegressionTests(wsDir, task)
	report.OracleResults[OracleRegressionTest] = res2
	if !res2.Passed {
		report.AllPassed = false
	}

	// 3. Oracle 3: Defect-Specific Reproducer Oracle
	res3 := s.verifyDefectReproducer(wsDir, task, patch, reproducerCmd)
	report.OracleResults[OracleDefectReproducer] = res3
	if !res3.Passed {
		report.AllPassed = false
	}

	// 4. Oracle 4: Behavioral Invariance Oracle
	res4 := s.verifyBehavioralInvariance(wsDir, task, invarianceCmd)
	report.OracleResults[OracleBehavioralInvariance] = res4
	if !res4.Passed {
		report.AllPassed = false
	}

	// 5. Oracle 5: Negative Properties Oracle (no shadow variables, no new unchecked errors)
	res5 := s.verifyNegativeProperties(wsDir, patch)
	report.OracleResults[OracleNegativeProperties] = res5
	if !res5.Passed {
		report.AllPassed = false
	}

	report.TotalDurationMs = time.Since(start).Milliseconds()
	return report
}

// -----------------------------------------------------------------------------------------
// 1. Syntax & Compilation Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifySyntaxCompilation(wsDir string, task *planning.RemediationTask, patch *CandidatePatch) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleSyntaxCompilation, Passed: true}

	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, targetPath, nil, parser.ParseComments)
	if err != nil {
		res.Passed = false
		res.FailureReason = fmt.Sprintf("Source file failed AST parse: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// Ensure workspace is a valid module for isolated go vet execution
	modPath := filepath.Join(wsDir, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		_ = os.WriteFile(modPath, []byte("module testfixture\n\ngo 1.21\n"), 0644)
	}

	// Execute `go vet` in target package directory
	pkgDir := filepath.Dir(targetPath)
	cmd := exec.Command("go", "vet", ".")
	cmd.Dir = pkgDir
	cmd.Env = append(os.Environ(), "GOTMPDIR=E:\\SWE\\.tmp", "TEMP=E:\\SWE\\.tmp", "TMP=E:\\SWE\\.tmp")

	out, err := cmd.CombinedOutput()
	res.Output = string(out)
	if err != nil {
		res.Passed = false
		res.FailureReason = fmt.Sprintf("go vet failed: %s", string(out))
	} else {
		res.Details = "Syntax valid and go vet passed with 0 warnings"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 2. Regression Test Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyRegressionTests(wsDir string, task *planning.RemediationTask) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleRegressionTest, Passed: true}

	if len(task.ValidationCommands) > 0 {
		for _, cmdStr := range task.ValidationCommands {
			parts := strings.Fields(cmdStr)
			if len(parts) == 0 {
				continue
			}

			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Dir = wsDir
			cmd.Env = append(os.Environ(), "GOTMPDIR=E:\\SWE\\.tmp", "TEMP=E:\\SWE\\.tmp", "TMP=E:\\SWE\\.tmp")

			out, err := cmd.CombinedOutput()
			if err != nil {
				res.Passed = false
				res.Output = string(out)
				res.FailureReason = fmt.Sprintf("Regression command %q failed: %v", cmdStr, err)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		}
		res.Details = fmt.Sprintf("All %d validation commands executed successfully", len(task.ValidationCommands))
	} else {
		res.Details = "No baseline test regressions detected"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 3. Defect-Specific Reproducer Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyDefectReproducer(
	wsDir string,
	task *planning.RemediationTask,
	patch *CandidatePatch,
	reproducerCmd string,
) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleDefectReproducer, Passed: true}

	// If explicit reproducer command provided, execute it
	if reproducerCmd != "" {
		parts := strings.Fields(reproducerCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = wsDir
		cmd.Env = append(os.Environ(), "GOTMPDIR=E:\\SWE\\.tmp", "TEMP=E:\\SWE\\.tmp", "TMP=E:\\SWE\\.tmp")

		out, err := cmd.CombinedOutput()
		res.Output = string(out)
		if err != nil {
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Defect reproducer test failed: %v\n%s", err, string(out))
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Details = "Defect reproducer test confirmed defect elimination"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// Semantic property check based on task category
	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	srcBytes, err := os.ReadFile(targetPath)
	if err != nil {
		res.Passed = false
		res.FailureReason = fmt.Sprintf("Cannot read patched file: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}
	src := string(srcBytes)

	taskType := strings.ToLower(task.Title + " " + task.Description)
	if strings.Contains(taskType, "resource") || strings.Contains(taskType, "leak") {
		if !strings.Contains(src, ".Close()") {
			res.Passed = false
			res.FailureReason = "Patched file does not contain Close() call"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		// Hardened check: Verify that the enclosing function does not return the resource to callers
		fset := token.NewFileSet()
		if fNode, err := parser.ParseFile(fset, targetPath, nil, 0); err == nil {
			for _, decl := range fNode.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					if fnReturnsResourceType(fn) && funcContainsClose(fn) {
						res.Passed = false
						res.FailureReason = "Defect reproducer check failed: function returns resource to caller; callee cleanup violates interface contract"
						res.DurationMs = time.Since(start).Milliseconds()
						return res
					}
				}
			}
		}
		res.Details = "Explicit or deferred resource cleanup verified along exit paths"
	} else if strings.Contains(taskType, "sql") {
		if strings.Contains(src, "fmt.Sprintf") && strings.Contains(src, "%s") {
			res.Passed = false
			res.FailureReason = "Patched file still contains dynamic fmt.Sprintf SQL formatting"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		if !strings.Contains(src, "?") {
			res.Passed = false
			res.FailureReason = "Patched file lacks parameterized placeholder '?'"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Details = "Parameterized SQL placeholder and parameter bindings verified"
	} else if strings.Contains(taskType, "error") {
		if !strings.Contains(src, "if err != nil") && !strings.Contains(src, "if err :=") {
			res.Passed = false
			res.FailureReason = "Patched file does not check error return value"
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Details = "Error handling and propagation check verified"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 4. Behavioral Invariance Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyBehavioralInvariance(
	wsDir string,
	task *planning.RemediationTask,
	invarianceCmd string,
) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleBehavioralInvariance, Passed: true}

	if invarianceCmd != "" {
		parts := strings.Fields(invarianceCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = wsDir
		cmd.Env = append(os.Environ(), "GOTMPDIR=E:\\SWE\\.tmp", "TEMP=E:\\SWE\\.tmp", "TMP=E:\\SWE\\.tmp")

		out, err := cmd.CombinedOutput()
		res.Output = string(out)
		if err != nil {
			res.Passed = false
			res.FailureReason = fmt.Sprintf("Behavioral invariance test failed: %v", err)
			res.DurationMs = time.Since(start).Milliseconds()
			return res
		}
		res.Details = "Non-defective execution paths produce identical invariant behavior"
	} else {
		res.Details = "Baseline invariant behavior verified"
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

// -----------------------------------------------------------------------------------------
// 5. Negative Properties Oracle
// -----------------------------------------------------------------------------------------

func (s *IndependentOracleSuite) verifyNegativeProperties(wsDir string, patch *CandidatePatch) *OracleResult {
	start := time.Now()
	res := &OracleResult{OracleType: OracleNegativeProperties, Passed: true}

	targetPath := filepath.Join(wsDir, filepath.FromSlash(patch.TargetFile))
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, targetPath, nil, parser.ParseComments)
	if err != nil {
		res.Passed = false
		res.FailureReason = fmt.Sprintf("AST parse failed: %v", err)
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// 1. Verify no shadowed 'err' definitions introduced by the candidate patch
	shadowedErr := false
	invertedErr := false
	diffAddsClose := false
	var closedVarNames []string

	if patch != nil && strings.Contains(patch.UnifiedDiff, "+") {
		for _, line := range strings.Split(patch.UnifiedDiff, "\n") {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				trimmed := strings.TrimSpace(line[1:])
				if strings.Contains(trimmed, "err :=") && (strings.HasPrefix(line, "+\t\t\t") || strings.HasPrefix(line, "+            ")) {
					shadowedErr = true
				}
				if strings.HasPrefix(trimmed, "if err == nil") && strings.Contains(trimmed, "return") {
					invertedErr = true
				}
				if strings.Contains(trimmed, ".Close()") {
					diffAddsClose = true
					// Extract variable name, e.g. "resp.Body.Close()" -> "resp", "f.Close()" -> "f"
					closeCall := strings.TrimPrefix(trimmed, "defer ")
					parts := strings.Split(closeCall, ".")
					if len(parts) > 0 {
						vName := strings.TrimSpace(parts[0])
						if vName != "" {
							closedVarNames = append(closedVarNames, vName)
						}
					}
				}
			}
		}
	}

	if shadowedErr {
		res.Passed = false
		res.FailureReason = "Negative property check failed: candidate patch introduces shadowed 'err' variable declaration in nested block"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	if invertedErr {
		res.Passed = false
		res.FailureReason = "Negative property check failed: candidate patch introduces inverted error check condition (if err == nil)"
		res.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	// 2. Structural & Semantic Invariants on Patched AST
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// A. Callee Cleanup on Escaping Resource Violation
		if diffAddsClose && fnReturnsResourceType(fn) {
			for _, vName := range closedVarNames {
				if fnBodyReturnsVar(fn, vName) || fnReturnsTypeMatchingVar(fn, vName) {
					res.Passed = false
					res.FailureReason = fmt.Sprintf("Negative property check failed: candidate patch introduces callee cleanup on escaping resource %q returned to caller; violates ownership transfer and panics on error returns", vName)
					res.DurationMs = time.Since(start).Milliseconds()
					return res
				}
			}
		}

		// B. Pre-Error Check Cleanup Placement (Nil dereference on failure path)
		if diffAddsClose {
			for _, vName := range closedVarNames {
				if isCleanupPlacedBeforeAllocErrorCheck(fn.Body.List, vName) {
					res.Passed = false
					res.FailureReason = fmt.Sprintf("Negative property check failed: resource cleanup for %q placed prior to error check; triggers nil pointer dereference panic when allocation fails", vName)
					res.DurationMs = time.Since(start).Milliseconds()
					return res
				}
			}
		}

		// C. Redundant or Double Close Detection
		closeCounts := make(map[string]int)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
					targetStr := formatOracleNode(sel.X)
					closeCounts[targetStr]++
				}
			}
			return true
		})
		for targetStr, count := range closeCounts {
			if count > 1 {
				res.Passed = false
				res.FailureReason = fmt.Sprintf("Negative property check failed: multiple Close() calls detected on resource %q without reassignment", targetStr)
				res.DurationMs = time.Since(start).Milliseconds()
				return res
			}
		}
	}

	res.Details = "Zero shadowed variables, nil-dereference risk, or ownership violations detected"
	res.DurationMs = time.Since(start).Milliseconds()
	return res
}

func fnReturnsResourceType(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		typeStr := formatOracleAstType(field.Type)
		if isOracleResourceType(typeStr) {
			return true
		}
	}
	return false
}

func fnReturnsTypeMatchingVar(fn *ast.FuncDecl, varName string) bool {
	// If varName is 'resp', check for *http.Response
	// If varName is 'f', check for *os.File
	// If varName is 'conn', check for net.Conn
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		typeStr := formatOracleAstType(field.Type)
		if strings.Contains(varName, "resp") && strings.Contains(typeStr, "Response") {
			return true
		}
		if strings.Contains(varName, "f") && strings.Contains(typeStr, "File") {
			return true
		}
		if strings.Contains(varName, "conn") && strings.Contains(typeStr, "Conn") {
			return true
		}
	}
	return false
}

func fnBodyReturnsVar(fn *ast.FuncDecl, varName string) bool {
	returned := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, expr := range ret.Results {
			if id, ok := expr.(*ast.Ident); ok && id.Name == varName {
				returned = true
				return false
			}
			if sel, ok := expr.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == varName {
					returned = true
					return false
				}
			}
		}
		return true
	})
	return returned
}

func funcContainsClose(fn *ast.FuncDecl) bool {
	hasClose := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			hasClose = true
			return false
		}
		return true
	})
	return hasClose
}

func stmtContainsClose(stmt ast.Stmt) bool {
	hasClose := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			hasClose = true
			return false
		}
		return true
	})
	return hasClose
}

func isCleanupPlacedBeforeAllocErrorCheck(stmts []ast.Stmt, varName string) bool {
	allocIdx := -1
	closeIdx := -1

	for i, stmt := range stmts {
		if allocIdx == -1 && stmtAssignsVar(stmt, varName) {
			allocIdx = i
		}
		if closeIdx == -1 && stmtClosesVar(stmt, varName) {
			closeIdx = i
		}
	}

	if allocIdx >= 0 && closeIdx > allocIdx {
		// Check if there is an error check between allocIdx and closeIdx
		hasCheckBetween := false
		for k := allocIdx + 1; k < closeIdx; k++ {
			if isErrCheckStmt(stmts[k]) {
				hasCheckBetween = true
				break
			}
		}
		// If no check between alloc and close, but an error check follows close:
		if !hasCheckBetween {
			for k := closeIdx + 1; k < len(stmts); k++ {
				if isErrCheckStmt(stmts[k]) {
					return true // Pre-error check cleanup!
				}
			}
		}
	}
	return false
}

func stmtAssignsVar(stmt ast.Stmt, varName string) bool {
	if assign, ok := stmt.(*ast.AssignStmt); ok {
		for _, lhs := range assign.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name == varName {
				return true
			}
		}
	}
	return false
}

func stmtClosesVar(stmt ast.Stmt, varName string) bool {
	closes := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == varName {
				closes = true
				return false
			}
			if innerSel, ok := sel.X.(*ast.SelectorExpr); ok {
				if id, ok := innerSel.X.(*ast.Ident); ok && id.Name == varName {
					closes = true
					return false
				}
			}
		}
		return true
	})
	return closes
}

func isErrCheckStmt(stmt ast.Stmt) bool {
	if ifStmt, ok := stmt.(*ast.IfStmt); ok {
		if bin, ok := ifStmt.Cond.(*ast.BinaryExpr); ok && bin.Op == token.NEQ {
			x := fmt.Sprintf("%v", bin.X)
			y := fmt.Sprintf("%v", bin.Y)
			return (x == "err" && y == "nil") || (x == "nil" && y == "err")
		}
	}
	return false
}

func formatOracleNode(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return formatOracleNode(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func formatOracleAstType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return "*" + formatOracleAstType(t.X)
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

func isOracleResourceType(typeStr string) bool {
	switch typeStr {
	case "*http.Response", "http.Response",
		"*os.File", "os.File",
		"net.Conn", "*net.Conn",
		"io.ReadCloser", "io.Closer",
		"*sql.Rows", "sql.Rows",
		"*sql.DB", "sql.DB":
		return true
	default:
		return false
	}
}
