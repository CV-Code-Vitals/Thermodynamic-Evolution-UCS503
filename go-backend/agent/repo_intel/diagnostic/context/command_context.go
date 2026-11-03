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

// EvaluateFile performs AST inspection of exec.Command invocations.
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

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for exec.Command or exec.CommandContext
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "exec" || (sel.Sel.Name != "Command" && sel.Sel.Name != "CommandContext") {
			return true
		}

		pos := fset.Position(callExpr.Pos())

		// Analyze arguments
		args := callExpr.Args
		if sel.Sel.Name == "CommandContext" && len(args) > 0 {
			args = args[1:] // skip context argument
		}

		isDynamic := false
		var dynamicSnippet string

		for _, arg := range args {
			// Check for string concatenation (+)
			if bin, isBin := arg.(*ast.BinaryExpr); isBin && bin.Op == token.ADD {
				isDynamic = true
				dynamicSnippet = string(src[bin.Pos()-1 : bin.End()-1])
				break
			}
			// Check for fmt.Sprintf calls within arguments
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
		}

		if isDynamic {
			eval.HasVulnerabilities = true
			eval.Findings = append(eval.Findings, ContextFinding{
				FindingID:           "CTX-SEC-CMD-INJECT-" + sanitizeRel(relPath) + fmt.Sprintf("-L%d", pos.Line),
				Category:            repo_intel.CategorySecurity,
				Severity:            repo_intel.SeverityHigh,
				Confidence:          0.95,
				ConfidenceRationale: "Dynamic string interpolation or binary concatenation detected in exec.Command arguments",
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
		} else if len(args) > 0 {
			// All arguments are static string literals or individual identifier references without concatenation
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
