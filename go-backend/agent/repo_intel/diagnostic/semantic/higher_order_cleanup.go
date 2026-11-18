package semantic

import (
	"go/ast"
	"go/token"
)

// CallbackInvocationType classifies how a higher-order wrapper invokes its function argument.
type CallbackInvocationType string

const (
	InvocationMustOnce    CallbackInvocationType = "INVOKE_MUST_ONCE"
	InvocationConditional CallbackInvocationType = "INVOKE_CONDITIONAL"
	InvocationNever       CallbackInvocationType = "INVOKE_NEVER"
	InvocationAsync       CallbackInvocationType = "INVOKE_ASYNC"
	InvocationStored      CallbackInvocationType = "INVOKE_STORED"
	InvocationUnknown     CallbackInvocationType = "INVOKE_UNKNOWN"
)

// HigherOrderCleanupResult describes the outcome of analyzing a deferred call or method-value wrapper.
type HigherOrderCleanupResult struct {
	IsVerifiedRelease  bool                   `json:"is_verified_release"`
	InvocationType     CallbackInvocationType `json:"invocation_type"`
	WrapperName        string                 `json:"wrapper_name"`
	TargetResourceVar  string                 `json:"target_resource_var"`
	MethodName         string                 `json:"method_name"`
	Evidence           string                 `json:"evidence"`
	IsAnonymousClosure bool                   `json:"is_anonymous_closure"`
}

// HigherOrderCleanupAnalyzer inspects deferred calls, method values, and closure wrappers.
type HigherOrderCleanupAnalyzer struct {
	registry      *ResourceRegistry
	summaryEngine *FunctionSummaryEngine
	localDecls    map[string]*ast.FuncDecl // local functions in package
}

// NewHigherOrderCleanupAnalyzer initializes a higher-order cleanup analyzer.
func NewHigherOrderCleanupAnalyzer(registry *ResourceRegistry, summaryEngine *FunctionSummaryEngine) *HigherOrderCleanupAnalyzer {
	if registry == nil {
		registry = DefaultResourceRegistry()
	}
	return &HigherOrderCleanupAnalyzer{
		registry:      registry,
		summaryEngine: summaryEngine,
		localDecls:    make(map[string]*ast.FuncDecl),
	}
}

// RegisterLocalFunc registers a local function AST declaration for wrapper inspection.
func (h *HigherOrderCleanupAnalyzer) RegisterLocalFunc(fnDecl *ast.FuncDecl) {
	if fnDecl != nil && fnDecl.Name != nil {
		h.localDecls[fnDecl.Name.Name] = fnDecl
	}
}

// AnalyzeDeferredCall evaluates a deferred call expression against a tracked resource instance.
func (h *HigherOrderCleanupAnalyzer) AnalyzeDeferredCall(
	call *ast.CallExpr,
	inst *ResourceInstance,
	fset *token.FileSet,
) HigherOrderCleanupResult {
	if call == nil {
		return HigherOrderCleanupResult{InvocationType: InvocationUnknown}
	}

	// 1. Direct release call: defer f.Close() or defer resp.Body.Close()
	if h.registry.IsReleaseCall(call, inst.Kind, inst.VarName, inst.Aliases) {
		return HigherOrderCleanupResult{
			IsVerifiedRelease: true,
			InvocationType:    InvocationMustOnce,
			TargetResourceVar: inst.VarName,
			MethodName:        "Close",
			Evidence:          "Direct deferred release call",
		}
	}

	// 2. Anonymous function literal: defer func() { f.Close() }()
	// The call.Fun could be an *ast.FuncLit or a call expression
	if funcLit, isLit := call.Fun.(*ast.FuncLit); isLit {
		return h.analyzeAnonymousFuncLit(funcLit, inst, fset)
	}

	// 3. Higher-Order Wrapper with Method Value: defer check(resp.Body.Close) or defer closer(f.Close)
	sel, isSel := call.Fun.(*ast.SelectorExpr)
	var wrapperName string
	if isSel {
		wrapperName = sel.Sel.Name
	} else if id, isId := call.Fun.(*ast.Ident); isId {
		wrapperName = id.Name
	}

	for _, arg := range call.Args {
		// A. Check if argument is a Method Value: resp.Body.Close or f.Close
		if argSel, isArgSel := arg.(*ast.SelectorExpr); isArgSel {
			methodName := argSel.Sel.Name
			if h.isReleaseMethodName(inst.Kind, methodName) {
				if h.matchesTargetResource(argSel.X, inst) {
					// Method value for this resource is passed to wrapper!
					invType := h.inspectWrapperInvocation(wrapperName, fset)
					isRelease := (invType == InvocationMustOnce)
					return HigherOrderCleanupResult{
						IsVerifiedRelease: isRelease,
						InvocationType:    invType,
						WrapperName:       wrapperName,
						TargetResourceVar: inst.VarName,
						MethodName:        methodName,
						Evidence:          "Deferred method-value wrapper: " + wrapperName + "(" + methodName + ")",
					}
				}
			}
		}

		// B. Check if argument is the resource handle directly: defer cleanup(f)
		if argId, isArgId := arg.(*ast.Ident); isArgId {
			if argId.Name == inst.VarName || h.isAliasMatch(argId.Name, inst.Aliases) {
				// Check if wrapper is a known summary or local func that closes its parameter
				if h.wrapperReleasesParameter(wrapperName, 0) {
					return HigherOrderCleanupResult{
						IsVerifiedRelease: true,
						InvocationType:    InvocationMustOnce,
						WrapperName:       wrapperName,
						TargetResourceVar: inst.VarName,
						Evidence:          "Deferred helper function releases resource parameter",
					}
				}
			}
		}
	}

	return HigherOrderCleanupResult{InvocationType: InvocationUnknown}
}

// analyzeAnonymousFuncLit unrolls the statements inside an anonymous deferred closure.
func (h *HigherOrderCleanupAnalyzer) analyzeAnonymousFuncLit(
	funcLit *ast.FuncLit,
	inst *ResourceInstance,
	fset *token.FileSet,
) HigherOrderCleanupResult {
	if funcLit == nil || funcLit.Body == nil {
		return HigherOrderCleanupResult{InvocationType: InvocationUnknown}
	}

	foundRelease := false
	var methodName string

	ast.Inspect(funcLit.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if h.registry.IsReleaseCall(call, inst.Kind, inst.VarName, inst.Aliases) {
			foundRelease = true
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				methodName = sel.Sel.Name
			}
			return false
		}
		return true
	})

	if foundRelease {
		return HigherOrderCleanupResult{
			IsVerifiedRelease:  true,
			InvocationType:     InvocationMustOnce,
			IsAnonymousClosure: true,
			TargetResourceVar:  inst.VarName,
			MethodName:         methodName,
			Evidence:           "Deferred anonymous function invokes " + methodName,
		}
	}

	return HigherOrderCleanupResult{
		IsVerifiedRelease:  false,
		InvocationType:     InvocationNever,
		IsAnonymousClosure: true,
		TargetResourceVar:  inst.VarName,
		Evidence:           "Deferred anonymous closure does not invoke release on target resource",
	}
}

// inspectWrapperInvocation analyzes whether a named wrapper function invokes its callback parameter.
func (h *HigherOrderCleanupAnalyzer) inspectWrapperInvocation(wrapperName string, fset *token.FileSet) CallbackInvocationType {
	// 1. Look up in registered local function declarations
	if fnDecl, ok := h.localDecls[wrapperName]; ok {
		return h.analyzeWrapperFuncDecl(fnDecl)
	}

	// 2. Standard library / common patterns:
	// A standard helper named "check" or "must" or "ignoreError" that is not local:
	// If unverified, default conservatively to InvocationUnknown
	return InvocationUnknown
}

// analyzeWrapperFuncDecl inspects the AST of a wrapper function to classify callback invocation.
func (h *HigherOrderCleanupAnalyzer) analyzeWrapperFuncDecl(fnDecl *ast.FuncDecl) CallbackInvocationType {
	if fnDecl == nil || fnDecl.Body == nil {
		return InvocationUnknown
	}

	// Identify callback parameter name (e.g. fn func() error)
	var callbackParamName string
	if fnDecl.Type != nil && fnDecl.Type.Params != nil {
		for _, field := range fnDecl.Type.Params.List {
			if _, isFuncType := field.Type.(*ast.FuncType); isFuncType {
				if len(field.Names) > 0 {
					callbackParamName = field.Names[0].Name
					break
				}
			}
		}
	}

	if callbackParamName == "" {
		return InvocationUnknown
	}

	calledUnconditionally := false
	calledInGoRoutine := false
	calledConditionally := false
	storedInVariable := false

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GoStmt:
			// Asynchronous invocation: go fn()
			if call, ok := node.Call.Fun.(*ast.Ident); ok && call.Name == callbackParamName {
				calledInGoRoutine = true
				return false
			}
		case *ast.IfStmt:
			// Conditional check surrounding invocation
			ast.Inspect(node.Body, func(inner ast.Node) bool {
				if call, ok := inner.(*ast.CallExpr); ok {
					if id, isId := call.Fun.(*ast.Ident); isId && id.Name == callbackParamName {
						calledConditionally = true
					}
				}
				return true
			})
		case *ast.CallExpr:
			if id, ok := node.Fun.(*ast.Ident); ok && id.Name == callbackParamName {
				if !calledConditionally && !calledInGoRoutine {
					calledUnconditionally = true
				}
			}
		case *ast.AssignStmt:
			// Checking if callback is stored for later: global = fn
			for _, rhs := range node.Rhs {
				if id, ok := rhs.(*ast.Ident); ok && id.Name == callbackParamName {
					storedInVariable = true
				}
			}
		}
		return true
	})

	if storedInVariable {
		return InvocationStored
	}
	if calledInGoRoutine {
		return InvocationAsync
	}
	if calledConditionally {
		return InvocationConditional
	}
	if calledUnconditionally {
		return InvocationMustOnce
	}

	return InvocationNever
}

// matchesTargetResource checks if an AST expression matches the target resource or its subfield.
func (h *HigherOrderCleanupAnalyzer) matchesTargetResource(expr ast.Expr, inst *ResourceInstance) bool {
	// Direct variable: f
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name == inst.VarName || h.isAliasMatch(id.Name, inst.Aliases)
	}

	// Subfield: resp.Body
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if id, ok := sel.X.(*ast.Ident); ok {
			if (id.Name == inst.VarName || id.Name == inst.ParentVarName || h.isAliasMatch(id.Name, inst.Aliases)) && sel.Sel.Name == "Body" {
				return true
			}
		}
	}

	return false
}

func (h *HigherOrderCleanupAnalyzer) isReleaseMethodName(kind ResourceKind, name string) bool {
	contract, ok := h.registry.Releases[kind]
	if !ok {
		return false
	}
	for _, m := range contract.ReleaseMethods {
		if m == name {
			return true
		}
	}
	return false
}

func (h *HigherOrderCleanupAnalyzer) isAliasMatch(name string, aliases []string) bool {
	for _, a := range aliases {
		if name == a {
			return true
		}
	}
	return false
}

func (h *HigherOrderCleanupAnalyzer) wrapperReleasesParameter(wrapperName string, paramIdx int) bool {
	if h.summaryEngine == nil {
		return false
	}
	sum := h.summaryEngine.GetSummary("", "", wrapperName)
	if sum == nil {
		return false
	}
	for _, eff := range sum.ResourceEffects {
		if eff.Effect == EffectResourceRelease && eff.ParamIndex == paramIdx {
			return true
		}
	}
	return false
}
