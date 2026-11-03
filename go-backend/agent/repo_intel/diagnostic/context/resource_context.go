package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ResourceContextAnalyzer detects resource handles opened inside loop bodies without proper per-iteration closure.
type ResourceContextAnalyzer struct {
	repoRoot string
}

func NewResourceContextAnalyzer(repoRoot string) *ResourceContextAnalyzer {
	return &ResourceContextAnalyzer{repoRoot: repoRoot}
}

// EvaluateFile inspects Go code for intra-loop resource leaks.
func (a *ResourceContextAnalyzer) EvaluateFile(relPath string) ([]ContextFinding, error) {
	var findings []ContextFinding
	fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))

	src, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
	if err != nil {
		return nil, nil
	}

	resourceAllocators := map[string]map[string]bool{
		"os":   {"Open": true, "OpenFile": true, "Create": true},
		"net":  {"Dial": true, "DialTimeout": true},
		"http": {"Get": true, "Post": true},
	}

	checkLoopBody := func(body *ast.BlockStmt, loopPos token.Pos) {
		for _, stmt := range body.List {
			// Check for assignments opening resources
			assignStmt, isAssign := stmt.(*ast.AssignStmt)
			if !isAssign {
				continue
			}

			for _, rhs := range assignStmt.Rhs {
				call, isCall := rhs.(*ast.CallExpr)
				if !isCall {
					continue
				}

				sel, isSel := call.Fun.(*ast.SelectorExpr)
				if !isSel {
					continue
				}

				pkgIdent, isPkgIdent := sel.X.(*ast.Ident)
				if !isPkgIdent {
					continue
				}

				if methods, ok := resourceAllocators[pkgIdent.Name]; ok && methods[sel.Sel.Name] {
					// Resource opened inside loop!
					// Now check if there is an explicit .Close() call inside the loop
					resourceVarName := ""
					if len(assignStmt.Lhs) > 0 {
						if lhsIdent, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
							resourceVarName = lhsIdent.Name
						}
					}

					hasCloseInLoop := false
					hasDeferInLoop := false

					for _, innerStmt := range body.List {
						// Check for defer inside loop
						if defStmt, isDef := innerStmt.(*ast.DeferStmt); isDef {
							if defCall, ok := defStmt.Call.Fun.(*ast.SelectorExpr); ok {
								if ident, ok := defCall.X.(*ast.Ident); ok && ident.Name == resourceVarName && defCall.Sel.Name == "Close" {
									hasDeferInLoop = true
								}
							}
						}
						// Check for direct close inside loop
						if exprStmt, isExpr := innerStmt.(*ast.ExprStmt); isExpr {
							if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
								if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
									if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == resourceVarName && selExpr.Sel.Name == "Close" {
										hasCloseInLoop = true
									}
								}
							}
						}
					}

					pos := fset.Position(assignStmt.Pos())
					loopPosition := fset.Position(loopPos)
					snippet := string(src[assignStmt.Pos()-1 : assignStmt.End()-1])

					if hasDeferInLoop {
						findings = append(findings, ContextFinding{
							FindingID:           "CTX-REL-DEFER-IN-LOOP-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line),
							Category:            repo_intel.CategoryReliability,
							Severity:            repo_intel.SeverityMedium,
							Confidence:          0.92,
							ConfidenceRationale: "Resource defer statement executed inside loop body delays handle cleanup until enclosing function returns",
							Title:               "Resource Defer Statement Inside Loop Body",
							Description:         fmt.Sprintf("Resource %s.%s opened at line %d uses defer inside loop, risking descriptor exhaustion.", pkgIdent.Name, sel.Sel.Name, pos.Line),
							RootCause:           "Defer statement within loop boundary postpones resource cleanup until outer function termination.",
							AffectedFiles:       []string{relPath},
							EvidenceTraces: []EvidenceTrace{
								{
									Location: repo_intel.EvidenceLocation{
										File:        relPath,
										LineStart:   pos.Line,
										LineEnd:     pos.Line,
										Description: fmt.Sprintf("defer %s.Close() in loop", resourceVarName),
									},
									ASTNodeType: "ast.DeferStmt",
									ContextType: "LOOP_BODY_DEFER",
									CodeSnippet: snippet,
									Rationale:   "Defers inside loops accumulate unclosed descriptors across iterations.",
								},
							},
							DetectionMethod:   "AST_LOOP_ANALYSIS",
							RecommendedAction: "Wrap iteration in an anonymous function closure or invoke Close() explicitly at loop iteration end.",
						})
					} else if !hasCloseInLoop {
						findings = append(findings, ContextFinding{
							FindingID:           "CTX-REL-LEAK-LOOP-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line),
							Category:            repo_intel.CategoryReliability,
							Severity:            repo_intel.SeverityHigh,
							Confidence:          0.95,
							ConfidenceRationale: "Resource handle acquired in loop iteration without corresponding Close() call within iteration scope",
							Title:               "Unclosed Resource Handle In Loop Iteration",
							Description:         fmt.Sprintf("File %s opens resource handle via %s.%s at line %d without closing it, causing descriptor leaks.", relPath, pkgIdent.Name, sel.Sel.Name, pos.Line),
							RootCause:           "Omission of resource cleanup within iterative loop lifecycle.",
							AffectedFiles:       []string{relPath},
							EvidenceTraces: []EvidenceTrace{
								{
									Location: repo_intel.EvidenceLocation{
										File:        relPath,
										LineStart:   pos.Line,
										LineEnd:     pos.Line,
										Description: fmt.Sprintf("%s.%s opened without Close()", pkgIdent.Name, sel.Sel.Name),
									},
									ASTNodeType: "ast.AssignStmt",
									ContextType: "LOOP_BODY_UNCLOSED",
									CodeSnippet: snippet,
									Rationale:   fmt.Sprintf("Handle %s created in loop at line %d (loop begins at line %d) lacks Close().", resourceVarName, pos.Line, loopPosition.Line),
								},
							},
							DetectionMethod:   "AST_LOOP_ANALYSIS",
							RecommendedAction: fmt.Sprintf("Add explicit %s.Close() or invoke closure function with defer before next iteration.", resourceVarName),
						})
					}
				}
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch loop := n.(type) {
		case *ast.ForStmt:
			checkLoopBody(loop.Body, loop.Pos())
		case *ast.RangeStmt:
			checkLoopBody(loop.Body, loop.Pos())
		}
		return true
	})

	return findings, nil
}
