package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/semantic"
)

// ResourceContextAnalyzer evaluates resource lifecycles using Semantic IR reachability analysis and loop-level AST checks.
type ResourceContextAnalyzer struct {
	repoRoot string
}

func NewResourceContextAnalyzer(repoRoot string) *ResourceContextAnalyzer {
	return &ResourceContextAnalyzer{repoRoot: repoRoot}
}

// ResourceContextEvaluation holds the analysis outcome for resource lifecycle inspection.
type ResourceContextEvaluation struct {
	Findings         []ContextFinding
	CounterEvidences []CounterEvidence
}

// EvaluateFile performs Semantic IR and AST context analysis on a Go file for resource lifecycle safety.
func (a *ResourceContextAnalyzer) EvaluateFile(relPath string) (*ResourceContextEvaluation, error) {
	eval := &ResourceContextEvaluation{
		Findings:         make([]ContextFinding, 0),
		CounterEvidences: make([]CounterEvidence, 0),
	}

	fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))
	src, err := os.ReadFile(fullPath)
	if err != nil {
		return eval, err
	}

	findingIDsSeen := make(map[string]bool)

	// 1. Semantic IR Resource Lifecycle Analysis Pass
	semAnalyzer := semantic.NewSemanticAnalyzer(a.repoRoot)
	fileModel, _, err := semAnalyzer.AnalyzeFile(relPath)
	if err == nil && fileModel != nil {
		resAnalyzer := semantic.NewResourceAnalyzer(a.repoRoot)
		for fnName, cfg := range fileModel.CFGs {
			df := fileModel.Dataflow[fnName]
			evs, err := resAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, relPath)
			if err != nil {
				continue
			}

			for _, ev := range evs {
				findingID := fmt.Sprintf("CTX-REL-RESOURCE-%s-%s-L%d", ev.ResourceKind, sanitizeRel(relPath), ev.Line)

				if ev.LifecycleState == semantic.StateResourceLeakConfirmed || ev.LifecycleState == semantic.StateResourceLeakProbable {
					if !findingIDsSeen[findingID] {
						findingIDsSeen[findingID] = true
						eval.Findings = append(eval.Findings, ContextFinding{
							FindingID:           findingID,
							Category:            repo_intel.CategoryReliability,
							Severity:            ev.Severity,
							Confidence:          ev.Confidence,
							ConfidenceRationale: ev.RootCause,
							State:               FindingState(ev.FindingState),
							Title:               ev.Title,
							Description:         ev.Description,
							RootCause:           ev.RootCause,
							AffectedFiles:       []string{relPath},
							EvidenceTraces: []EvidenceTrace{
								{
									Location: repo_intel.EvidenceLocation{
										File:        relPath,
										LineStart:   ev.Line,
										LineEnd:     ev.Line,
										Description: ev.AcquisitionSnippet,
									},
									ASTNodeType: "ast.CallExpr",
									ContextType: "RESOURCE_ACQUISITION",
									CodeSnippet: ev.AcquisitionSnippet,
									Rationale:   fmt.Sprintf("Resource handle '%s' (%s) lifecycle state: %s", ev.VarName, ev.ResourceKind, ev.LifecycleState),
								},
							},
							DetectionMethod:   "SEMANTIC_RESOURCE_ANALYSIS",
							RecommendedAction: ev.RecommendedAction,
						})
					}
				} else if ev.LifecycleState == semantic.StateResourceDeferredRelease || ev.LifecycleState == semantic.StateResourceReleased || ev.LifecycleState == semantic.StateResourceTransferred {
					eval.CounterEvidences = append(eval.CounterEvidences, CounterEvidence{
						RuleID:      "SAFE_RESOURCE_LIFECYCLE",
						Description: fmt.Sprintf("Resource '%s' (%s) verified safe lifecycle: %s", ev.VarName, ev.ResourceKind, ev.LifecycleState),
						Location:    fmt.Sprintf("%s:%d", relPath, ev.Line),
						VerifiedAt:  time.Now(),
					})
				}
			}
		}
	}

	// 2. Intra-Loop AST Fallback Analysis (Preserves baseline loop heuristic compatibility)
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
	if err != nil {
		return eval, nil
	}

	resourceAllocators := map[string]map[string]bool{
		"os":   {"Open": true, "OpenFile": true, "Create": true},
		"net":  {"Dial": true, "DialTimeout": true},
		"http": {"Get": true, "Post": true},
		"sql":  {"Open": true, "OpenDB": true},
	}

	checkLoopBody := func(body *ast.BlockStmt, loopPos token.Pos) {
		for _, stmt := range body.List {
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
					resourceVarName := ""
					if len(assignStmt.Lhs) > 0 {
						if lhsIdent, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
							resourceVarName = lhsIdent.Name
						}
					}

					hasCloseInLoop := false
					hasDeferInLoop := false

					for _, innerStmt := range body.List {
						if defStmt, isDef := innerStmt.(*ast.DeferStmt); isDef {
							if defCall, ok := defStmt.Call.Fun.(*ast.SelectorExpr); ok {
								if ident, ok := defCall.X.(*ast.Ident); ok && ident.Name == resourceVarName && defCall.Sel.Name == "Close" {
									hasDeferInLoop = true
								}
							}
						}
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
						findingID := "CTX-REL-DEFER-IN-LOOP-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line)
						if !findingIDsSeen[findingID] {
							findingIDsSeen[findingID] = true
							eval.Findings = append(eval.Findings, ContextFinding{
								FindingID:           findingID,
								Category:            repo_intel.CategoryReliability,
								Severity:            repo_intel.SeverityMedium,
								Confidence:          0.92,
								ConfidenceRationale: "Resource defer statement executed inside loop body delays handle cleanup until enclosing function returns",
								State:               StateConfirmed,
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
						}
					} else if !hasCloseInLoop {
						findingID := "CTX-REL-LEAK-LOOP-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line)
						if !findingIDsSeen[findingID] {
							findingIDsSeen[findingID] = true
							eval.Findings = append(eval.Findings, ContextFinding{
								FindingID:           findingID,
								Category:            repo_intel.CategoryReliability,
								Severity:            repo_intel.SeverityHigh,
								Confidence:          0.95,
								ConfidenceRationale: "Resource handle acquired in loop iteration without corresponding Close() call within iteration scope",
								State:               StateConfirmed,
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

	return eval, nil
}
