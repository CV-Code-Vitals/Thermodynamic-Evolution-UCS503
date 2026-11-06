package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// CommandClassification classifies the safety and execution style of an external process invocation.
type CommandClassification string

const (
	ClassificationStaticCommand                 CommandClassification = "STATIC_COMMAND"
	ClassificationSafeArgumentSeparation        CommandClassification = "SAFE_ARGUMENT_SEPARATION"
	ClassificationDynamicArguments              CommandClassification = "DYNAMIC_ARGUMENTS"
	ClassificationShellInterpretedCommand       CommandClassification = "SHELL_INTERPRETED_COMMAND"
	ClassificationUntrustedShellCommand         CommandClassification = "UNTRUSTED_SHELL_COMMAND"
	ClassificationUnresolvedCommandConstruction CommandClassification = "UNRESOLVED_COMMAND_CONSTRUCTION"
)

// Recognized shell interpreters and their command execution flags
var shellInterpreters = map[string]map[string]bool{
	"sh":             {"-c": true},
	"/bin/sh":        {"-c": true},
	"bash":           {"-c": true},
	"/bin/bash":      {"-c": true},
	"zsh":            {"-c": true},
	"/bin/zsh":       {"-c": true},
	"cmd":            {"/c": true, "/C": true, "/k": true, "/K": true},
	"cmd.exe":        {"/c": true, "/C": true, "/k": true, "/K": true},
	"powershell":     {"-c": true, "-Command": true, "-command": true},
	"powershell.exe": {"-c": true, "-Command": true, "-command": true},
	"pwsh":           {"-c": true, "-Command": true, "-command": true},
	"pwsh.exe":       {"-c": true, "-Command": true, "-command": true},
}

// CommandEvidence represents the detailed semantic trace of a command invocation.
type CommandEvidence struct {
	FilePath            string                `json:"file_path"`
	Package             string                `json:"package"`
	Function            string                `json:"function"`
	Line                int                   `json:"line"`
	Classification      CommandClassification `json:"classification"`
	SinkName            string                `json:"sink_name"`
	SinkCallSnippet     string                `json:"sink_call_snippet"`
	BinaryExpr          string                `json:"binary_expr"`
	IsShellInterpreter  bool                  `json:"is_shell_interpreter"`
	ShellCommandSnippet string                `json:"shell_command_snippet,omitempty"`
	ArgumentSnippets    []string              `json:"argument_snippets"`
	UntrustedInputs     []string              `json:"untrusted_inputs,omitempty"`
	ReachingDefs        []string              `json:"reaching_defs,omitempty"`
	AnalysisLimitations []string              `json:"analysis_limitations,omitempty"`
	CounterEvidence     []string              `json:"counter_evidence,omitempty"`
	Confidence          float64               `json:"confidence"`
	LifecycleState      FindingState          `json:"lifecycle_state"`
	Severity            repo_intel.Severity   `json:"severity"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	RootCause           string                `json:"root_cause,omitempty"`
	RecommendedAction   string                `json:"recommended_action,omitempty"`
	ASTNodeType         string                `json:"ast_node_type"`
}

// CommandAnalyzer performs semantic expression, shell-interpretation, and dataflow analysis on command executions.
type CommandAnalyzer struct {
	repoRoot      string
	summaryEngine *FunctionSummaryEngine
}

// SetSummaryEngine attaches an interprocedural function summary engine.
func (a *CommandAnalyzer) SetSummaryEngine(se *FunctionSummaryEngine) {
	a.summaryEngine = se
}

// NewCommandAnalyzer creates a new CommandAnalyzer instance.
func NewCommandAnalyzer(repoRoot string) *CommandAnalyzer {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return &CommandAnalyzer{repoRoot: repoRoot}
}

// AnalyzeFunction inspects a function or method declaration for process execution invocations.
func (a *CommandAnalyzer) AnalyzeFunction(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	reachingDefs *ReachingDefinitionsAnalysis,
	filePath string,
) ([]CommandEvidence, error) {
	var results []CommandEvidence
	if fnDecl == nil || fnDecl.Body == nil || cfg == nil {
		return results, nil
	}

	fnName := fnDecl.Name.Name
	pkgName := ""

	// Extract function parameters
	paramMap := make(map[string]string)
	if fnDecl.Type.Params != nil {
		for _, p := range fnDecl.Type.Params.List {
			typeStr := formatTypeExpr(p.Type, fset)
			for _, name := range p.Names {
				paramMap[name.Name] = typeStr
			}
		}
	}

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sinkName, binArg, cmdArgs, isCmdSink := a.identifyCommandSink(callExpr)
		if !isCmdSink {
			return true
		}

		pos := fset.Position(callExpr.Pos())
		sinkCallSnip := exprToString(callExpr, fset)
		binSnip := exprToString(binArg, fset)

		argSnips := make([]string, len(cmdArgs))
		for i, arg := range cmdArgs {
			argSnips[i] = exprToString(arg, fset)
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

		ev := a.evaluateCommandInvocation(
			binArg,
			cmdArgs,
			sinkName,
			sinkCallSnip,
			binSnip,
			argSnips,
			paramMap,
			currentBlock,
			reachingDefs,
			fset,
			filePath,
			pkgName,
			fnName,
			pos.Line,
		)

		results = append(results, ev)
		return true
	})

	return results, nil
}

// identifyCommandSink checks whether a call expression invokes exec.Command or exec.CommandContext.
func (a *CommandAnalyzer) identifyCommandSink(call *ast.CallExpr) (string, ast.Expr, []ast.Expr, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", nil, nil, false
	}

	pkgIdent, hasPkg := sel.X.(*ast.Ident)
	if !hasPkg || pkgIdent.Name != "exec" {
		return "", nil, nil, false
	}

	methodName := sel.Sel.Name
	args := call.Args

	switch methodName {
	case "Command":
		if len(args) > 0 {
			return "exec.Command", args[0], args[1:], true
		}
	case "CommandContext":
		if len(args) > 1 {
			return "exec.CommandContext", args[1], args[2:], true
		}
	}

	return "", nil, nil, false
}

// evaluateCommandInvocation analyzes whether an invocation uses safe argument separation or untrusted shell interpretation.
func (a *CommandAnalyzer) evaluateCommandInvocation(
	binArg ast.Expr,
	cmdArgs []ast.Expr,
	sinkName, sinkCallSnip, binSnip string,
	argSnips []string,
	paramMap map[string]string,
	block *Block,
	reachingDefs *ReachingDefinitionsAnalysis,
	fset *token.FileSet,
	filePath, pkgName, fnName string,
	line int,
) CommandEvidence {
	ev := CommandEvidence{
		FilePath:         filePath,
		Package:          pkgName,
		Function:         fnName,
		Line:             line,
		SinkName:         sinkName,
		SinkCallSnippet:  sinkCallSnip,
		BinaryExpr:       binSnip,
		ArgumentSnippets: argSnips,
		ASTNodeType:      fmt.Sprintf("%T", binArg),
	}

	binLit, isBinLit := binArg.(*ast.BasicLit)
	binName := ""
	if isBinLit && binLit.Kind == token.STRING {
		binName = cleanStringLit(binLit.Value)
	}

	// 1. Check for Shell Interpreter Invocation
	cleanBinBase := strings.ToLower(filepath.Base(binName))
	if flagMap, isShell := shellInterpreters[cleanBinBase]; isShell {
		ev.IsShellInterpreter = true
		return a.evaluateShellInterpretation(binName, cmdArgs, flagMap, ev, paramMap, reachingDefs, block, fset)
	}

	// 2. Direct Process Execution (Non-Shell Binary)
	// Check if the binary itself is a dynamic variable or parameter
	if !isBinLit {
		if id, isId := binArg.(*ast.Ident); isId {
			if paramType, isParam := paramMap[id.Name]; isParam {
				ev.Classification = ClassificationDynamicArguments
				ev.Confidence = 0.90
				ev.LifecycleState = StateProbable
				ev.Severity = repo_intel.SeverityMedium
				ev.UntrustedInputs = []string{fmt.Sprintf("binary_param:%s (%s)", id.Name, paramType)}
				ev.Title = "Dynamic Executable Path from Parameter"
				ev.Description = fmt.Sprintf("Executable path in %s is derived from function parameter %s.", sinkName, id.Name)
				return ev
			}
		}
	}

	// Direct argument separation check
	allArgsStatic := isBinLit
	var dynamicArgNames []string

	for _, arg := range cmdArgs {
		if lit, isLit := arg.(*ast.BasicLit); isLit && lit.Kind == token.STRING {
			continue
		}
		allArgsStatic = false
		argStr := exprToString(arg, fset)
		if id, isId := arg.(*ast.Ident); isId {
			dynamicArgNames = append(dynamicArgNames, id.Name)
		} else {
			dynamicArgNames = append(dynamicArgNames, argStr)
		}
	}

	if allArgsStatic {
		ev.Classification = ClassificationStaticCommand
		ev.Confidence = 0.99
		ev.LifecycleState = StateCounterEvidencePresent
		ev.CounterEvidence = []string{"Fixed binary invoked with static string literal arguments without shell interpretation"}
		ev.Title = "Static Command Execution"
		ev.Description = "Subprocess invoked with static command name and constant arguments."
		return ev
	}

	// Dynamic arguments passed directly to non-shell binary (e.g. exec.Command("git", "checkout", branch))
	ev.Classification = ClassificationSafeArgumentSeparation
	ev.Confidence = 0.95
	ev.LifecycleState = StateCounterEvidencePresent
	ev.CounterEvidence = []string{
		fmt.Sprintf("Arguments (%s) are passed as discrete array elements directly to executable %s without shell interpretation",
			strings.Join(dynamicArgNames, ", "), binName),
	}
	ev.Title = "Safe Discrete Argument Separation"
	ev.Description = "Command uses OS-level argv separation; arguments are not subject to shell parsing or word splitting."
	return ev
}

// evaluateShellInterpretation examines command strings passed to shells (sh -c, cmd.exe /c).
func (a *CommandAnalyzer) evaluateShellInterpretation(
	binName string,
	cmdArgs []ast.Expr,
	flagMap map[string]bool,
	ev CommandEvidence,
	paramMap map[string]string,
	reachingDefs *ReachingDefinitionsAnalysis,
	block *Block,
	fset *token.FileSet,
) CommandEvidence {
	if len(cmdArgs) < 2 {
		ev.Classification = ClassificationShellInterpretedCommand
		ev.Confidence = 0.70
		ev.LifecycleState = StateInsufficientEvidence
		ev.AnalysisLimitations = []string{"Shell invoked with insufficient arguments for command string analysis"}
		return ev
	}

	// Check if first argument is a command flag like "-c" or "/c"
	flagArg := cmdArgs[0]
	flagLit, isFlagLit := flagArg.(*ast.BasicLit)
	if !isFlagLit || flagLit.Kind != token.STRING {
		ev.Classification = ClassificationUnresolvedCommandConstruction
		ev.Confidence = 0.60
		ev.LifecycleState = StateInsufficientEvidence
		return ev
	}

	flagVal := cleanStringLit(flagLit.Value)
	if !flagMap[flagVal] {
		// Shell called without -c flag
		ev.Classification = ClassificationSafeArgumentSeparation
		ev.Confidence = 0.85
		ev.LifecycleState = StateCounterEvidencePresent
		ev.CounterEvidence = []string{"Shell binary invoked as direct process without execution flags"}
		return ev
	}

	// The second argument is the Shell Command String!
	commandExpr := cmdArgs[1]
	ev.ShellCommandSnippet = exprToString(commandExpr, fset)

	// Resolve root expression via reaching definitions if commandExpr is an identifier or variable
	rootCmdExpr := commandExpr
	if reachingDefs != nil && block != nil {
		resolved, defTrace := reachingDefs.ResolveRootExpression(commandExpr, block.ID, commandExpr.Pos())
		rootCmdExpr = resolved
		for _, d := range defTrace {
			ev.ReachingDefs = append(ev.ReachingDefs, fmt.Sprintf("def:%s (line %d)", d.VarName, d.Line))
		}
	}

	// Case A: Shell command is a string literal
	if lit, ok := rootCmdExpr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		ev.Classification = ClassificationShellInterpretedCommand
		ev.Confidence = 0.95
		ev.LifecycleState = StateCounterEvidencePresent
		ev.CounterEvidence = []string{"Shell command string is a static unconcatenated string literal"}
		ev.Title = "Static Shell Command Execution"
		ev.Description = fmt.Sprintf("Shell %s executes a constant command string.", binName)
		return ev
	}

	// Case B: Shell command is a direct function parameter
	if id, ok := rootCmdExpr.(*ast.Ident); ok {
		if paramType, isParam := paramMap[id.Name]; isParam {
			ev.Classification = ClassificationUntrustedShellCommand
			ev.Confidence = 0.98
			ev.LifecycleState = StateConfirmed
			ev.Severity = repo_intel.SeverityCritical
			ev.UntrustedInputs = []string{fmt.Sprintf("param:%s (%s)", id.Name, paramType)}
			ev.Title = "Command Injection Risk via Shell Interpreter"
			ev.Description = fmt.Sprintf("Command string in %s at line %d is passed directly from function parameter '%s' to shell interpreter '%s -c'.",
				ev.Function, ev.Line, id.Name, binName)
			ev.RootCause = "Executing arbitrary untrusted shell strings with shell interpretation (-c)."
			ev.RecommendedAction = "Invoke target executable directly with discrete arguments instead of passing shell commands through 'sh -c'."
			return ev
		}
	}

	// Case C: Shell command is an inline fmt.Sprintf call or resolved from one
	if call, isCall := rootCmdExpr.(*ast.CallExpr); isCall {
		return a.evaluateShellCallExpr(call, binName, ev, paramMap, fset)
	}

	// Case D: Shell command is an inline binary concatenation or resolved from one
	if bin, isBin := rootCmdExpr.(*ast.BinaryExpr); isBin && bin.Op == token.ADD {
		return a.evaluateShellBinaryConcat(bin, binName, ev, paramMap, fset)
	}

	ev.Classification = ClassificationUnresolvedCommandConstruction
	ev.Confidence = 0.60
	ev.LifecycleState = StateInsufficientEvidence
	ev.AnalysisLimitations = []string{"Unresolved dynamic shell command expression"}
	return ev
}

func (a *CommandAnalyzer) evaluateShellCallExpr(
	call *ast.CallExpr,
	binName string,
	ev CommandEvidence,
	paramMap map[string]string,
	fset *token.FileSet,
) CommandEvidence {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		ev.Classification = ClassificationUnresolvedCommandConstruction
		ev.Confidence = 0.50
		ev.LifecycleState = StateInsufficientEvidence
		ev.AnalysisLimitations = []string{"Complex non-selector call in shell command expression"}
		return ev
	}

	pkgIdent, hasPkg := sel.X.(*ast.Ident)
	if hasPkg && pkgIdent.Name == "fmt" && (sel.Sel.Name == "Sprintf" || sel.Sel.Name == "Sprint") {
		var untrusted []string
		for _, arg := range call.Args[1:] {
			if id, isId := arg.(*ast.Ident); isId {
				if paramType, isParam := paramMap[id.Name]; isParam {
					untrusted = append(untrusted, fmt.Sprintf("param:%s (%s)", id.Name, paramType))
				} else {
					untrusted = append(untrusted, fmt.Sprintf("var:%s", id.Name))
				}
			} else {
				untrusted = append(untrusted, fmt.Sprintf("expr:%s", exprToString(arg, fset)))
			}
		}

		if len(untrusted) > 0 {
			ev.Classification = ClassificationUntrustedShellCommand
			ev.Confidence = 0.98
			ev.LifecycleState = StateConfirmed
			ev.Severity = repo_intel.SeverityCritical
			ev.UntrustedInputs = untrusted
			ev.Title = "Command Injection Risk via Formatted Shell String"
			ev.Description = fmt.Sprintf("Shell command executed via '%s -c' dynamically formats shell string with untrusted input(s): %s.",
				binName, strings.Join(untrusted, ", "))
			ev.RootCause = "Formatting user variables into shell interpreter string (-c)."
			ev.RecommendedAction = "Execute target binary directly without shell interpretation, passing arguments as discrete array."
			return ev
		}
	}

	// Interprocedural Helper Summary Lookup
	if a.summaryEngine != nil {
		var calleeSummary *FunctionSummary
		if hasPkg {
			calleeSummary = a.summaryEngine.GetSummary(pkgIdent.Name, "", sel.Sel.Name)
		} else {
			calleeSummary = a.summaryEngine.GetSummary("", "", sel.Sel.Name)
		}

		if calleeSummary != nil {
			var untrusted []string
			for _, flow := range calleeSummary.ParamReturnFlows {
				if flow.Effect == EffectTaintTransform || flow.Effect == EffectTaintPassthrough {
					if flow.ParamIndex < len(call.Args) {
						arg := call.Args[flow.ParamIndex]
						if id, isId := arg.(*ast.Ident); isId {
							if paramType, isParam := paramMap[id.Name]; isParam {
								untrusted = append(untrusted, fmt.Sprintf("param:%s (%s)", id.Name, paramType))
							} else {
								untrusted = append(untrusted, fmt.Sprintf("var:%s", id.Name))
							}
						} else {
							untrusted = append(untrusted, fmt.Sprintf("expr:%s", exprToString(arg, fset)))
						}
					}
				}
			}

			if len(untrusted) > 0 {
				ev.Classification = ClassificationUntrustedShellCommand
				ev.Confidence = 0.98
				ev.LifecycleState = StateConfirmed
				ev.Severity = repo_intel.SeverityCritical
				ev.UntrustedInputs = untrusted
				ev.Title = "Command Injection Risk via Interprocedural Shell Helper"
				ev.Description = fmt.Sprintf("Shell command executed via '%s -c' receives shell string from helper '%s' tainted with untrusted input(s): %s.",
					binName, calleeSummary.FunctionName, strings.Join(untrusted, ", "))
				ev.RootCause = "Executing dynamic shell command string assembled from untrusted input in helper function."
				ev.RecommendedAction = "Execute target binary directly without shell interpreter (-c)."
				return ev
			}
		}
	}

	ev.Classification = ClassificationShellInterpretedCommand
	ev.Confidence = 0.85
	ev.LifecycleState = StateProbable
	return ev
}

func (a *CommandAnalyzer) evaluateShellBinaryConcat(
	bin *ast.BinaryExpr,
	binName string,
	ev CommandEvidence,
	paramMap map[string]string,
	fset *token.FileSet,
) CommandEvidence {
	leaves := collectAdditionLeaves(bin)
	var untrusted []string

	for _, leaf := range leaves {
		if _, isLit := leaf.(*ast.BasicLit); isLit {
			continue
		}
		if id, isId := leaf.(*ast.Ident); isId {
			if paramType, isParam := paramMap[id.Name]; isParam {
				untrusted = append(untrusted, fmt.Sprintf("param:%s (%s)", id.Name, paramType))
			} else {
				untrusted = append(untrusted, fmt.Sprintf("var:%s", id.Name))
			}
		} else {
			untrusted = append(untrusted, fmt.Sprintf("expr:%s", exprToString(leaf, fset)))
		}
	}

	if len(untrusted) > 0 {
		ev.Classification = ClassificationUntrustedShellCommand
		ev.Confidence = 0.98
		ev.LifecycleState = StateConfirmed
		ev.Severity = repo_intel.SeverityCritical
		ev.UntrustedInputs = untrusted
		ev.Title = "Command Injection Risk via Shell String Concatenation"
		ev.Description = fmt.Sprintf("Shell command executed via '%s -c' concatenates untrusted input(s): %s.",
			binName, strings.Join(untrusted, ", "))
		ev.RootCause = "Direct binary concatenation (+) of dynamic input into shell command string."
		ev.RecommendedAction = "Execute target binary directly without shell interpreter (-c)."
		return ev
	}

	ev.Classification = ClassificationShellInterpretedCommand
	ev.Confidence = 0.85
	ev.LifecycleState = StateProbable
	return ev
}
