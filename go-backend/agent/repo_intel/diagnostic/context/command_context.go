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

// CommandContextAnalyzer evaluates exec.Command calls for injection vulnerabilities and static safety.
type CommandContextAnalyzer struct {
	repoRoot string
}

func NewCommandContextAnalyzer(repoRoot string) *CommandContextAnalyzer {
	return &CommandContextAnalyzer{repoRoot: repoRoot}
}

// CommandContextEvaluation holds the analysis outcome for command execution.
type CommandContextEvaluation struct {
	HasVulnerabilities bool
	Findings           []ContextFinding
	CounterEvidences   []CounterEvidence
}

// EvaluateFile performs AST and semantic IR inspection of exec.Command invocations.
func (a *CommandContextAnalyzer) EvaluateFile(relPath string) (*CommandContextEvaluation, error) {
	eval := &CommandContextEvaluation{}
	fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))

	src, err := os.ReadFile(fullPath)
	if err != nil {
		return eval, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
	if err != nil {
		return eval, nil
	}

	findingIDsSeen := make(map[string]bool)

	// 1. Semantic IR Analysis Pass (Expression-aware shell interpretation, parameter tracing, discrete argv)
	semAnalyzer := semantic.NewSemanticAnalyzer(a.repoRoot)
	fileModel, _, err := semAnalyzer.AnalyzeFile(relPath)
	if err == nil && fileModel != nil {
		cmdSem := semantic.NewCommandAnalyzer(a.repoRoot)
		for fnName, cfg := range fileModel.CFGs {
			df := fileModel.Dataflow[fnName]
			evs, err := cmdSem.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, relPath)
			if err != nil {
				continue
			}
			for _, ev := range evs {
				if ev.Classification == semantic.ClassificationUntrustedShellCommand {
					eval.HasVulnerabilities = true
					fid := fmt.Sprintf("CTX-SEC-CMD-INJECT-%s-L%d", sanitizeRel(relPath), ev.Line)
					if !findingIDsSeen[fid] {
						findingIDsSeen[fid] = true
						eval.Findings = append(eval.Findings, ContextFinding{
							FindingID:           fid,
							Category:            repo_intel.CategorySecurity,
							Severity:            repo_intel.SeverityHigh,
							Confidence:          ev.Confidence,
							ConfidenceRationale: fmt.Sprintf("%s: %s", ev.Title, ev.Description),
							State:               StateConfirmed,
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
										Description: ev.Title,
									},
									ASTNodeType: ev.ASTNodeType,
									ContextType: "SHELL_COMMAND_SEMANTIC",
									CodeSnippet: ev.ShellCommandSnippet,
									Rationale:   ev.Description,
								},
							},
							DetectionMethod:   "SEMANTIC_DATAFLOW_ANALYSIS",
							RecommendedAction: ev.RecommendedAction,
						})
					}
				} else if ev.Classification == semantic.ClassificationSafeArgumentSeparation || ev.Classification == semantic.ClassificationStaticCommand {
					eval.CounterEvidences = append(eval.CounterEvidences, CounterEvidence{
						RuleID:      "SAFE_STATIC_EXEC_ARGUMENTS",
						Description: fmt.Sprintf("%s: %s", ev.Title, ev.Description),
						Location:    fmt.Sprintf("%s:%d", filepath.ToSlash(relPath), ev.Line),
						VerifiedAt:  time.Now(),
					})
				}
			}
		}
	}

	// 2. Legacy AST Inspection Pass (for backward compatibility)
	dynamicVars := make(map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for idx, rhs := range assign.Rhs {
			isDyn := false
			var snip string
			if bin, isBin := rhs.(*ast.BinaryExpr); isBin && bin.Op == token.ADD {
				isDyn = true
				snip = string(src[bin.Pos()-1 : bin.End()-1])
			} else if call, isCall := rhs.(*ast.CallExpr); isCall {
				if sel, isSel := call.Fun.(*ast.SelectorExpr); isSel {
					if id, isId := sel.X.(*ast.Ident); isId && id.Name == "fmt" && sel.Sel.Name == "Sprintf" {
						isDyn = true
						snip = string(src[call.Pos()-1 : call.End()-1])
					}
				}
			}
			if isDyn && idx < len(assign.Lhs) {
				if lhsId, isLhsId := assign.Lhs[idx].(*ast.Ident); isLhsId {
					dynamicVars[lhsId.Name] = snip
				}
			}
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "exec" || (sel.Sel.Name != "Command" && sel.Sel.Name != "CommandContext") {
			return true
		}

		pos := fset.Position(callExpr.Pos())
		args := callExpr.Args
		if sel.Sel.Name == "CommandContext" && len(args) > 0 {
			args = args[1:]
		}

		isDynamic := false
		var dynamicSnippet string

		for _, arg := range args {
			if bin, isBin := arg.(*ast.BinaryExpr); isBin && bin.Op == token.ADD {
				isDynamic = true
				dynamicSnippet = string(src[bin.Pos()-1 : bin.End()-1])
				break
			}
			if subCall, isSubCall := arg.(*ast.CallExpr); isSubCall {
				if subSel, isSubSel := subCall.Fun.(*ast.SelectorExpr); isSubSel {
					if subIdent, isSubIdent := subSel.X.(*ast.Ident); isSubIdent {
						if subIdent.Name == "fmt" && subSel.Sel.Name == "Sprintf" {
							isDynamic = true
							dynamicSnippet = string(src[subCall.Pos()-1 : subCall.End()-1])
							break
						}
					}
				}
			}
			if id, isId := arg.(*ast.Ident); isId {
				if snip, ok := dynamicVars[id.Name]; ok {
					isDynamic = true
					dynamicSnippet = snip
					break
				}
			}
		}

		if isDynamic {
			eval.HasVulnerabilities = true
			fid := "CTX-SEC-CMD-INJECT-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line)
			if !findingIDsSeen[fid] {
				findingIDsSeen[fid] = true
				eval.Findings = append(eval.Findings, ContextFinding{
					FindingID:           fid,
					Category:            repo_intel.CategorySecurity,
					Severity:            repo_intel.SeverityHigh,
					Confidence:          0.95,
					ConfidenceRationale: "Dynamic string interpolation or binary concatenation detected in exec.Command arguments",
					State:               StateConfirmed,
					Title:               "Command Argument Injection Risk",
					Description:         "Process execution command dynamically formats or concatenates arguments, risking shell/argument injection.",
					RootCause:           "Dynamic argument composition rather than discrete argument slicing.",
					AffectedFiles:       []string{relPath},
					EvidenceTraces: []EvidenceTrace{
						{
							Location: repo_intel.EvidenceLocation{
								File:        relPath,
								LineStart:   pos.Line,
								LineEnd:     pos.Line,
								Description: "Dynamic argument in exec.Command",
							},
							ASTNodeType: "ast.CallExpr",
							ContextType: "EXEC_ARGUMENT",
							CodeSnippet: dynamicSnippet,
							Rationale:   "Argument expression dynamically formats input instead of passing fixed discrete tokens.",
						},
					},
					DetectionMethod:   "AST_CALL_ANALYSIS",
					RecommendedAction: "Pass command arguments as separate discrete string elements without dynamic string concatenation.",
				})
			}
		} else if len(args) > 0 {
			eval.CounterEvidences = append(eval.CounterEvidences, CounterEvidence{
				RuleID:      "SAFE_STATIC_EXEC_ARGUMENTS",
				Description: "exec.Command uses discrete, unconcatenated arguments without shell interpolation",
				Location:    fmt.Sprintf("%s:%d", filepath.ToSlash(relPath), pos.Line),
				VerifiedAt:  time.Now(),
			})
		}

		return true
	})

	return eval, nil
}
