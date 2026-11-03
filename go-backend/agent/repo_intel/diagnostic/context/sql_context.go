package context

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// SQLContextAnalyzer evaluates SQL queries for safety, parameterized bindings, and dynamic concatenation.
type SQLContextAnalyzer struct {
	repoRoot string
}

func NewSQLContextAnalyzer(repoRoot string) *SQLContextAnalyzer {
	return &SQLContextAnalyzer{repoRoot: repoRoot}
}

// SQLContextEvaluation holds the analysis outcome for a file's SQL usage.
type SQLContextEvaluation struct {
	HasVulnerabilities bool
	Findings           []ContextFinding
	CounterEvidences   []CounterEvidence
}

var (
	parameterizedMarker = regexp.MustCompile(`(\$[0-9]+|\?|:[a-zA-Z0-9_]+)`)
	rawSQLSelect        = regexp.MustCompile(`(?i)(SELECT\s+.+\s+FROM\s+[a-zA-Z0-9_]+)`)
)

// EvaluateFile performs AST and lexical context analysis on a Go file for SQL operations.
func (a *SQLContextAnalyzer) EvaluateFile(relPath string) (*SQLContextEvaluation, error) {
	eval := &SQLContextEvaluation{}
	fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))

	src, err := os.ReadFile(fullPath)
	if err != nil {
		return eval, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
	if err != nil {
		return eval, nil // Not valid Go or partial file; skip AST inspection
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for binary expressions adding strings with SQL keywords
		binExpr, ok := n.(*ast.BinaryExpr)
		if !ok || binExpr.Op != token.ADD {
			return true
		}

		// Check if left or right contains SQL query strings
		leftLit, leftIsLit := binExpr.X.(*ast.BasicLit)
		rightLit, rightIsLit := binExpr.Y.(*ast.BasicLit)

		leftHasSQL := leftIsLit && leftLit.Kind == token.STRING && rawSQLSelect.MatchString(leftLit.Value)
		rightHasSQL := rightIsLit && rightLit.Kind == token.STRING && rawSQLSelect.MatchString(rightLit.Value)

		if leftHasSQL || rightHasSQL {
			// Dynamic concatenation detected!
			pos := fset.Position(binExpr.Pos())
			snippet := string(src[binExpr.Pos()-1 : binExpr.End()-1])

			eval.HasVulnerabilities = true
			eval.Findings = append(eval.Findings, ContextFinding{
				FindingID:           "CTX-SEC-SQL-INJECT-" + sanitizeRel(relPath),
				Category:            repo_intel.CategorySecurity,
				Severity:            repo_intel.SeverityHigh,
				Confidence:          0.96,
				ConfidenceRationale: "AST BinaryExpr concatenation with raw SQL string literal and dynamic variable operand",
				Title:               "Unparameterized SQL Query Concatenation",
				Description:         "Dynamic SQL query constructed via binary string concatenation instead of parameterized binding.",
				RootCause:           "Direct concatenation of dynamic values into SQL query string.",
				AffectedFiles:       []string{relPath},
				EvidenceTraces: []EvidenceTrace{
					{
						Location: repo_intel.EvidenceLocation{
							File:        relPath,
							LineStart:   pos.Line,
							LineEnd:     pos.Line,
							Description: "Dynamic SQL concatenation",
						},
						ASTNodeType: "ast.BinaryExpr",
						ContextType: "SQL_CONCATENATION",
						CodeSnippet: snippet,
						Rationale:   "Direct string addition (+) operator combines SQL syntax with dynamic variable.",
					},
				},
				DetectionMethod:   "AST_TAINT_CONTEXT",
				RecommendedAction: "Replace dynamic concatenation with parameterized placeholders ($1, ?) and pass parameters to QueryRow/Exec.",
			})
		}

		return true
	})

	// Check for safe parameterized queries that might have been falsely flagged
	lines := strings.Split(string(src), "\n")
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if rawSQLSelect.MatchString(trimmed) {
			if parameterizedMarker.MatchString(trimmed) && !strings.Contains(trimmed, "+") {
				// Query is safely parameterized and contains no concatenation!
				eval.CounterEvidences = append(eval.CounterEvidences, CounterEvidence{
					RuleID:      "SAFE_SQL_PARAMETERIZATION",
					Description: "SQL query uses parameterized binding placeholders without string concatenation",
					Location:    filepath.ToSlash(relPath) + fmtLine(lineNum),
					VerifiedAt:  time.Now(),
				})
			}
		}
	}

	return eval, nil
}

func sanitizeRel(rel string) string {
	r := strings.ReplaceAll(rel, "/", "-")
	r = strings.ReplaceAll(r, "\\", "-")
	r = strings.ReplaceAll(r, ".", "_")
	return r
}

func fmtLine(line int) string {
	return ":" + strings.TrimSpace(strings.ReplaceAll(string(rune('0'+line%10)), "", "")) + ""
}

