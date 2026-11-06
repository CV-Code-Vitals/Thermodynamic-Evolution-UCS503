package semantic

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// SQLClassification classifies the safety and construction style of a SQL expression.
type SQLClassification string

const (
	ClassificationStaticSQL                 SQLClassification = "STATIC_SQL"
	ClassificationSafeParameterizedSQL       SQLClassification = "SAFE_PARAMETERIZED_SQL"
	ClassificationDynamicSQLStructure       SQLClassification = "DYNAMIC_SQL_STRUCTURE"
	ClassificationUntrustedSQLStructure     SQLClassification = "UNTRUSTED_SQL_STRUCTURE"
	ClassificationUnresolvedSQLConstruction SQLClassification = "UNRESOLVED_SQL_CONSTRUCTION"
	ClassificationNonSQLExpression          SQLClassification = "NON_SQL_EXPRESSION"
)

var (
	sqlKeywordsRegex     = regexp.MustCompile(`(?i)\b(SELECT|INSERT\s+INTO|UPDATE|DELETE\s+FROM|DROP\s+TABLE|ALTER\s+TABLE|CREATE\s+TABLE)\b`)
	sqlPlaceholdersRegex = regexp.MustCompile(`(\$[0-9]+|\?|:[a-zA-Z0-9_]+|@[a-zA-Z0-9_]+)`)
)

// SQLEvidence represents the detailed semantic trace of a SQL query construction and execution.
type SQLEvidence struct {
	FilePath            string             `json:"file_path"`
	Package             string             `json:"package"`
	Function            string             `json:"function"`
	Line                int                `json:"line"`
	Classification      SQLClassification  `json:"classification"`
	SinkName            string             `json:"sink_name"`
	SinkCallSnippet     string             `json:"sink_call_snippet"`
	QueryExprSnippet    string             `json:"query_expr_snippet"`
	ConstructionExpr    string             `json:"construction_expr,omitempty"`
	ReachingDefs        []string           `json:"reaching_defs,omitempty"`
	ParameterArguments  []string           `json:"parameter_arguments,omitempty"`
	UntrustedInputs     []string           `json:"untrusted_inputs,omitempty"`
	AnalysisLimitations []string           `json:"analysis_limitations,omitempty"`
	CounterEvidence     []string           `json:"counter_evidence,omitempty"`
	Confidence          float64            `json:"confidence"`
	LifecycleState      FindingState       `json:"lifecycle_state"`
	Severity            repo_intel.Severity `json:"severity"`
	Title               string             `json:"title"`
	Description         string             `json:"description"`
	RootCause           string             `json:"root_cause,omitempty"`
	RecommendedAction   string             `json:"recommended_action,omitempty"`
	ASTNodeType         string             `json:"ast_node_type"`
}

// SQLAnalyzer performs semantic expression and dataflow analysis on SQL constructions.
type SQLAnalyzer struct {
	repoRoot      string
	summaryEngine *FunctionSummaryEngine
}

// SetSummaryEngine attaches an interprocedural function summary engine.
func (a *SQLAnalyzer) SetSummaryEngine(se *FunctionSummaryEngine) {
	a.summaryEngine = se
}

// NewSQLAnalyzer creates a new SQLAnalyzer instance.
func NewSQLAnalyzer(repoRoot string) *SQLAnalyzer {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return &SQLAnalyzer{repoRoot: repoRoot}
}

// AnalyzeFunction inspects a function or method declaration for SQL query execution and construction.
func (a *SQLAnalyzer) AnalyzeFunction(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	reachingDefs *ReachingDefinitionsAnalysis,
	filePath string,
) ([]SQLEvidence, error) {
	var results []SQLEvidence
	if fnDecl == nil || fnDecl.Body == nil || cfg == nil {
		return results, nil
	}

	fnName := fnDecl.Name.Name
	pkgName := ""
	if cfg.FuncDecl != nil && cfg.FuncDecl.Name != nil {
		fnName = cfg.FuncDecl.Name.Name
	}

	// Helper to extract function parameter names
	paramMap := make(map[string]string) // name -> type
	if fnDecl.Type.Params != nil {
		for _, p := range fnDecl.Type.Params.List {
			typeStr := formatTypeExpr(p.Type, fset)
			for _, name := range p.Names {
				paramMap[name.Name] = typeStr
			}
		}
	}

	// Traverse function AST looking for call expressions at SQL sinks
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sinkName, queryExpr, paramArgs, isSQLSink := a.identifySQLSink(callExpr)
		if !isSQLSink {
			return true
		}

		pos := fset.Position(callExpr.Pos())
		sinkCallSnip := exprToString(callExpr, fset)
		querySnip := exprToString(queryExpr, fset)

		paramArgSnips := make([]string, len(paramArgs))
		for i, pa := range paramArgs {
			paramArgSnips[i] = exprToString(pa, fset)
		}

		// Find containing block
		var currentBlock *Block
		for _, b := range cfg.Blocks {
			for _, st := range b.Statements {
				if st.Pos() <= callExpr.Pos() && callExpr.End() <= st.End() {
					currentBlock = b
					break
				}
			}
			if currentBlock != nil {
				break
			}
		}
		if currentBlock == nil && len(cfg.Blocks) > 0 {
			currentBlock = cfg.Entry
		}

		evidence := a.evaluateQueryExpression(
			queryExpr,
			sinkName,
			sinkCallSnip,
			querySnip,
			paramArgSnips,
			paramMap,
			currentBlock,
			reachingDefs,
			fset,
			filePath,
			pkgName,
			fnName,
			pos.Line,
		)

		results = append(results, evidence)
		return true
	})

	return results, nil
}

// identifySQLSink checks whether a call expression invokes a supported database/sql execution API.
func (a *SQLAnalyzer) identifySQLSink(call *ast.CallExpr) (string, ast.Expr, []ast.Expr, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", nil, nil, false
	}

	methodName := sel.Sel.Name
	args := call.Args

	switch methodName {
	case "Query", "QueryRow", "Exec", "Prepare":
		if len(args) > 0 {
			return methodName, args[0], args[1:], true
		}
	case "QueryContext", "QueryRowContext", "ExecContext", "PrepareContext":
		if len(args) > 1 {
			return methodName, args[1], args[2:], true
		}
	}

	return "", nil, nil, false
}

// evaluateQueryExpression traces the query construction and determines its classification and evidence.
func (a *SQLAnalyzer) evaluateQueryExpression(
	queryExpr ast.Expr,
	sinkName, sinkCallSnip, querySnip string,
	paramArgSnips []string,
	paramMap map[string]string,
	block *Block,
	reachingDefs *ReachingDefinitionsAnalysis,
	fset *token.FileSet,
	filePath, pkgName, fnName string,
	line int,
) SQLEvidence {
	ev := SQLEvidence{
		FilePath:           filePath,
		Package:            pkgName,
		Function:           fnName,
		Line:               line,
		SinkName:           sinkName,
		SinkCallSnippet:    sinkCallSnip,
		QueryExprSnippet:   querySnip,
		ParameterArguments: paramArgSnips,
		ASTNodeType:        fmt.Sprintf("%T", queryExpr),
	}

	// Resolve root expression via reaching definitions if queryExpr is an identifier or variable
	rootExpr := queryExpr
	if reachingDefs != nil && block != nil {
		resolved, defTrace := reachingDefs.ResolveRootExpression(queryExpr, block.ID, queryExpr.Pos())
		rootExpr = resolved
		for _, d := range defTrace {
			ev.ReachingDefs = append(ev.ReachingDefs, fmt.Sprintf("def:%s (line %d)", d.VarName, d.Line))
		}
	}

	// 1. Direct String Literal
	if lit, ok := rootExpr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		val := cleanStringLit(lit.Value)
		if !sqlKeywordsRegex.MatchString(val) {
			ev.Classification = ClassificationNonSQLExpression
			ev.Confidence = 0.90
			ev.LifecycleState = StateCounterEvidencePresent
			ev.CounterEvidence = []string{"Expression passed to method does not contain recognized SQL statement keywords"}
			return ev
		}

		hasPlaceholders := sqlPlaceholdersRegex.MatchString(val)
		if hasPlaceholders || len(paramArgSnips) > 0 {
			ev.Classification = ClassificationSafeParameterizedSQL
			ev.Confidence = 0.98
			ev.LifecycleState = StateCounterEvidencePresent
			ev.CounterEvidence = []string{
				fmt.Sprintf("Query string contains bind placeholder (%s) and receives %d discrete parameter arguments at sink",
					sqlPlaceholdersRegex.FindString(val), len(paramArgSnips)),
			}
			ev.Title = "Safe Parameterized SQL Query"
			ev.Description = "Query uses parameterized placeholders with separately supplied execution arguments."
			return ev
		}

		ev.Classification = ClassificationStaticSQL
		ev.Confidence = 0.99
		ev.LifecycleState = StateCounterEvidencePresent
		ev.CounterEvidence = []string{"Query is an unconcatenated static string literal with no dynamic arguments"}
		ev.Title = "Static Constant SQL Query"
		ev.Description = "Static SQL query executed without runtime string formatting."
		return ev
	}

	// 2. Binary Addition (String Concatenation)
	if binExpr, ok := rootExpr.(*ast.BinaryExpr); ok && binExpr.Op == token.ADD {
		ev.ConstructionExpr = exprToString(binExpr, fset)
		return a.evaluateBinaryConcat(binExpr, ev, paramMap, reachingDefs, block, fset)
	}

	// 3. Call to Formatting Function (fmt.Sprintf / strings.Join)
	if callExpr, ok := rootExpr.(*ast.CallExpr); ok {
		ev.ConstructionExpr = exprToString(callExpr, fset)
		return a.evaluateCallConstruction(callExpr, ev, paramMap, reachingDefs, block, fset)
	}

	// 4. Identifier that remains a function parameter directly
	if id, ok := rootExpr.(*ast.Ident); ok {
		varName := id.Name
		if paramType, isParam := paramMap[varName]; isParam {
			ev.Classification = ClassificationUntrustedSQLStructure
			ev.Confidence = 0.96
			ev.LifecycleState = StateConfirmed
			ev.Severity = repo_intel.SeverityCritical
			ev.UntrustedInputs = []string{fmt.Sprintf("param:%s (%s)", varName, paramType)}
			ev.Title = "Untrusted SQL Query Passed as Function Parameter"
			ev.Description = fmt.Sprintf("SQL query at %s is supplied directly from function parameter %s without bind parameterization.", sinkName, varName)
			ev.RootCause = "Exposing SQL query structure directly to caller parameter."
			ev.RecommendedAction = "Pass query parameters separately using bind placeholders ($1, ?) rather than dynamic query strings."
			return ev
		}
	}

	// 5. Default Fallback
	ev.Classification = ClassificationUnresolvedSQLConstruction
	ev.Confidence = 0.50
	ev.LifecycleState = StateInsufficientEvidence
	ev.AnalysisLimitations = []string{"Complex or unsupported AST expression type for SQL query argument"}
	ev.Title = "Unresolved SQL Query Construction"
	ev.Description = "Query construction could not be statically verified."
	return ev
}

// evaluateBinaryConcat inspects a string addition tree for SQL keywords and untrusted dynamic operands.
func (a *SQLAnalyzer) evaluateBinaryConcat(
	binExpr *ast.BinaryExpr,
	ev SQLEvidence,
	paramMap map[string]string,
	reachingDefs *ReachingDefinitionsAnalysis,
	block *Block,
	fset *token.FileSet,
) SQLEvidence {
	leaves := collectAdditionLeaves(binExpr)
	hasSQLKeyword := false
	var dynamicLeaves []string
	var untrustedLeaves []string

	for _, leaf := range leaves {
		switch l := leaf.(type) {
		case *ast.BasicLit:
			if l.Kind == token.STRING {
				val := cleanStringLit(l.Value)
				if sqlKeywordsRegex.MatchString(val) {
					hasSQLKeyword = true
				}
			}
		case *ast.Ident:
			dynamicLeaves = append(dynamicLeaves, l.Name)
			if paramType, isParam := paramMap[l.Name]; isParam {
				untrustedLeaves = append(untrustedLeaves, fmt.Sprintf("param:%s (%s)", l.Name, paramType))
			} else {
				untrustedLeaves = append(untrustedLeaves, fmt.Sprintf("var:%s", l.Name))
			}
		case *ast.CallExpr:
			callStr := exprToString(l, fset)
			dynamicLeaves = append(dynamicLeaves, callStr)
			untrustedLeaves = append(untrustedLeaves, fmt.Sprintf("call:%s", callStr))
		default:
			dynamicLeaves = append(dynamicLeaves, exprToString(leaf, fset))
		}
	}

	if !hasSQLKeyword {
		ev.Classification = ClassificationNonSQLExpression
		ev.Confidence = 0.85
		ev.LifecycleState = StateCounterEvidencePresent
		ev.CounterEvidence = []string{"Concatenation expression does not contain recognized SQL keywords"}
		return ev
	}

	if len(untrustedLeaves) > 0 {
		ev.Classification = ClassificationUntrustedSQLStructure
		ev.Confidence = 0.96
		ev.LifecycleState = StateConfirmed
		ev.Severity = repo_intel.SeverityCritical
		ev.UntrustedInputs = untrustedLeaves
		ev.Title = "SQL Injection Risk via String Concatenation"
		ev.Description = fmt.Sprintf("SQL query in %s combines raw SQL literals with dynamic variable(s): %s.",
			ev.Function, strings.Join(dynamicLeaves, ", "))
		ev.RootCause = "Direct binary concatenation (+) of dynamic values into SQL statement string."
		ev.RecommendedAction = "Replace dynamic concatenation with parameterized bind placeholders ($1, ?) and pass values to Exec/Query."
		return ev
	}

	ev.Classification = ClassificationDynamicSQLStructure
	ev.Confidence = 0.88
	ev.LifecycleState = StateProbable
	ev.Severity = repo_intel.SeverityMedium
	ev.Title = "Dynamic SQL Construction via String Concatenation"
	ev.Description = "SQL statement assembled from multiple string fragments at runtime."
	return ev
}

// evaluateCallConstruction inspects formatting calls like fmt.Sprintf or query builders.
func (a *SQLAnalyzer) evaluateCallConstruction(
	callExpr *ast.CallExpr,
	ev SQLEvidence,
	paramMap map[string]string,
	reachingDefs *ReachingDefinitionsAnalysis,
	block *Block,
	fset *token.FileSet,
) SQLEvidence {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		ev.Classification = ClassificationUnresolvedSQLConstruction
		ev.Confidence = 0.50
		ev.LifecycleState = StateInsufficientEvidence
		ev.AnalysisLimitations = []string{"Complex non-selector call expression for SQL query"}
		return ev
	}

	pkgIdent, hasPkg := sel.X.(*ast.Ident)
	callTarget := sel.Sel.Name

	// 1. fmt.Sprintf / fmt.Sprint
	if hasPkg && pkgIdent.Name == "fmt" && (callTarget == "Sprintf" || callTarget == "Sprint") {
		if len(callExpr.Args) == 0 {
			ev.Classification = ClassificationNonSQLExpression
			return ev
		}

		formatArg := callExpr.Args[0]
		formatLit, isLit := formatArg.(*ast.BasicLit)
		if !isLit || formatLit.Kind != token.STRING {
			ev.Classification = ClassificationUnresolvedSQLConstruction
			ev.Confidence = 0.60
			ev.LifecycleState = StateInsufficientEvidence
			ev.AnalysisLimitations = []string{"Non-literal format string passed to fmt.Sprintf"}
			return ev
		}

		formatVal := cleanStringLit(formatLit.Value)
		if !sqlKeywordsRegex.MatchString(formatVal) {
			ev.Classification = ClassificationNonSQLExpression
			ev.Confidence = 0.90
			ev.LifecycleState = StateCounterEvidencePresent
			ev.CounterEvidence = []string{"fmt.Sprintf format string does not contain recognized SQL keywords"}
			return ev
		}

		// Inspect format arguments (args[1:])
		var dynamicArgs []string
		var untrustedArgs []string

		for _, arg := range callExpr.Args[1:] {
			argStr := exprToString(arg, fset)
			dynamicArgs = append(dynamicArgs, argStr)

			if id, isId := arg.(*ast.Ident); isId {
				if paramType, isParam := paramMap[id.Name]; isParam {
					untrustedArgs = append(untrustedArgs, fmt.Sprintf("param:%s (%s)", id.Name, paramType))
				} else {
					untrustedArgs = append(untrustedArgs, fmt.Sprintf("var:%s", id.Name))
				}
			} else {
				untrustedArgs = append(untrustedArgs, fmt.Sprintf("expr:%s", argStr))
			}
		}

		if len(untrustedArgs) > 0 {
			ev.Classification = ClassificationUntrustedSQLStructure
			ev.Confidence = 0.97
			ev.LifecycleState = StateConfirmed
			ev.Severity = repo_intel.SeverityCritical
			ev.UntrustedInputs = untrustedArgs
			ev.Title = "SQL Injection Risk via Dynamic Formatting (fmt.Sprintf)"
			ev.Description = fmt.Sprintf("SQL query constructed via fmt.Sprintf interpolates untrusted argument(s) [%s] directly into SQL syntax.",
				strings.Join(dynamicArgs, ", "))
			ev.RootCause = "Direct interpolation of dynamic variables into SQL format string instead of bind placeholders."
			ev.RecommendedAction = "Use parameterized SQL placeholders ($1, ?) in query string and pass variables as execution arguments."
			return ev
		}

		ev.Classification = ClassificationDynamicSQLStructure
		ev.Confidence = 0.85
		ev.LifecycleState = StateProbable
		ev.Severity = repo_intel.SeverityMedium
		return ev
	}

	// 2. strings.Join
	if hasPkg && pkgIdent.Name == "strings" && callTarget == "Join" {
		ev.Classification = ClassificationDynamicSQLStructure
		ev.Confidence = 0.80
		ev.LifecycleState = StateProbable
		ev.Severity = repo_intel.SeverityLow
		ev.Title = "SQL Clause Constructed via strings.Join"
		ev.Description = "Dynamic SQL query clause assembled using strings.Join."
		return ev
	}

	// 3. Interprocedural Helper Function Summary Lookup
	if a.summaryEngine != nil {
		var calleeSummary *FunctionSummary
		if hasPkg {
			calleeSummary = a.summaryEngine.GetSummary(pkgIdent.Name, "", callTarget)
		} else {
			calleeSummary = a.summaryEngine.GetSummary("", "", callTarget)
		}

		if calleeSummary != nil {
			var untrustedArgs []string
			var dynamicArgs []string

			for _, flow := range calleeSummary.ParamReturnFlows {
				if flow.Effect == EffectTaintTransform || flow.Effect == EffectTaintPassthrough {
					if flow.ParamIndex < len(callExpr.Args) {
						arg := callExpr.Args[flow.ParamIndex]
						argStr := exprToString(arg, fset)
						dynamicArgs = append(dynamicArgs, argStr)

						if reachingDefs != nil && block != nil {
							rootArg, _ := reachingDefs.ResolveRootExpression(arg, block.ID, arg.Pos())
							if id, ok := rootArg.(*ast.Ident); ok {
								if paramType, isParam := paramMap[id.Name]; isParam {
									untrustedArgs = append(untrustedArgs, fmt.Sprintf("param:%s (%s)", id.Name, paramType))
								} else {
									untrustedArgs = append(untrustedArgs, fmt.Sprintf("var:%s", id.Name))
								}
							} else {
								untrustedArgs = append(untrustedArgs, fmt.Sprintf("expr:%s", argStr))
							}
						}
					}
				}
			}

			if len(untrustedArgs) > 0 {
				ev.Classification = ClassificationUntrustedSQLStructure
				ev.Confidence = 0.96
				ev.LifecycleState = StateConfirmed
				ev.Severity = repo_intel.SeverityCritical
				ev.UntrustedInputs = untrustedArgs
				ev.Title = "SQL Injection Risk via Interprocedural Helper Formatting"
				ev.Description = fmt.Sprintf("SQL query string constructed via helper '%s' interpolates untrusted argument(s) [%s] directly into SQL syntax.",
					calleeSummary.FunctionName, strings.Join(dynamicArgs, ", "))
				ev.RootCause = "Direct interpolation of dynamic variables into SQL format string inside helper function without parameter placeholders."
				ev.RecommendedAction = "Use parameterized SQL placeholders ($1, ?) in query string and pass variables as execution arguments."
				return ev
			}
		}
	}

	ev.Classification = ClassificationUnresolvedSQLConstruction
	ev.Confidence = 0.50
	ev.LifecycleState = StateInsufficientEvidence
	ev.AnalysisLimitations = []string{fmt.Sprintf("Call expression '%s' is an unsupported or unresolvable query constructor", exprToString(callExpr, fset))}
	return ev
}

// collectAdditionLeaves decomposes a binary addition tree into its constituent operands.
func collectAdditionLeaves(expr ast.Expr) []ast.Expr {
	var leaves []ast.Expr
	var walk func(e ast.Expr)
	walk = func(e ast.Expr) {
		if bin, ok := e.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
			walk(bin.X)
			walk(bin.Y)
		} else {
			leaves = append(leaves, e)
		}
	}
	walk(expr)
	return leaves
}

func cleanStringLit(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") && len(raw) >= 2 {
		return raw[1 : len(raw)-1]
	}
	if strings.HasPrefix(raw, "`") && strings.HasSuffix(raw, "`") && len(raw) >= 2 {
		return raw[1 : len(raw)-1]
	}
	return raw
}

func exprToString(expr ast.Expr, fset *token.FileSet) string {
	if expr == nil {
		return ""
	}
	var buf bytes.Buffer
	if fset != nil {
		_ = printer.Fprint(&buf, fset, expr)
		return buf.String()
	}
	return fmt.Sprintf("%v", expr)
}
