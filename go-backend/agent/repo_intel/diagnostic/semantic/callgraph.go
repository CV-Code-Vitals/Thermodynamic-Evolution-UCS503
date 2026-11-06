package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

// CallResolutionState classifies the certainty of static call target resolution.
type CallResolutionState string

const (
	CallResolvedExact        CallResolutionState = "RESOLVED_EXACT"
	CallResolvedConservative CallResolutionState = "RESOLVED_CONSERVATIVE"
	CallMultipleTargets      CallResolutionState = "MULTIPLE_TARGETS"
	CallExternalOpaque       CallResolutionState = "EXTERNAL_OPAQUE"
	CallUnresolved           CallResolutionState = "UNRESOLVED"
)

// CallTarget describes the resolved or estimated target of a function or method invocation.
type CallTarget struct {
	PackageName        string              `json:"package_name"`
	PackagePath        string              `json:"package_path"`
	FunctionName       string              `json:"function_name"`
	ReceiverType       string              `json:"receiver_type,omitempty"`
	File               string              `json:"file,omitempty"`
	Line               int                 `json:"line,omitempty"`
	Signature          *FunctionSignature  `json:"signature,omitempty"`
	ResolutionState    CallResolutionState `json:"resolution_state"`
	IsResourceConsumer bool                `json:"is_resource_consumer"`
	IsErrorPropagator  bool                `json:"is_error_propagator"`
	Summary            string              `json:"summary,omitempty"`
}

// CallEdge records an invocation link between a caller context and a target symbol.
type CallEdge struct {
	CallerPackage  string        `json:"caller_package"`
	CallerFile     string        `json:"caller_file"`
	CallerFunction string        `json:"caller_function"`
	CallSiteLine   int           `json:"call_site_line"`
	CallSnippet    string        `json:"call_snippet"`
	Target         *CallTarget   `json:"target"`
	CallExpr       *ast.CallExpr `json:"-"`
}

// BoundedCallGraph captures intra- and inter-package call relationships with depth bounds.
type BoundedCallGraph struct {
	PackageName string                   `json:"package_name"`
	PackagePath string                   `json:"package_path"`
	Targets     map[string]*CallTarget   `json:"targets"`  // Symbol key -> target
	Edges       []*CallEdge              `json:"edges"`
	Outgoing    map[string][]*CallEdge   `json:"outgoing"` // Caller symbol -> edges
	Incoming    map[string][]*CallEdge   `json:"incoming"` // Target symbol -> edges
}

// NewBoundedCallGraph initializes an empty call graph for a package.
func NewBoundedCallGraph(pkgName, pkgPath string) *BoundedCallGraph {
	return &BoundedCallGraph{
		PackageName: pkgName,
		PackagePath: pkgPath,
		Targets:     make(map[string]*CallTarget),
		Edges:       make([]*CallEdge, 0),
		Outgoing:    make(map[string][]*CallEdge),
		Incoming:    make(map[string][]*CallEdge),
	}
}

// CallSiteResolver resolves call expressions across local files, same-package files, and imported modules.
type CallSiteResolver struct {
	repoRoot    string
	pkgModels   map[string]*PackageSemanticModel // Package path / dir -> model
	allSymbols  map[string]*Symbol               // Qualified symbol name -> symbol
}

// NewCallSiteResolver creates a resolver instance rooted at repoRoot.
func NewCallSiteResolver(repoRoot string) *CallSiteResolver {
	return &CallSiteResolver{
		repoRoot:   repoRoot,
		pkgModels:  make(map[string]*PackageSemanticModel),
		allSymbols: make(map[string]*Symbol),
	}
}

// RegisterPackage adds a PackageSemanticModel to the cross-package resolver index.
func (r *CallSiteResolver) RegisterPackage(pkgModel *PackageSemanticModel) {
	if pkgModel == nil {
		return
	}
	r.pkgModels[pkgModel.PackagePath] = pkgModel
	r.pkgModels[pkgModel.PackageName] = pkgModel

	if pkgModel.SymbolTable != nil {
		for _, sym := range pkgModel.SymbolTable.Symbols {
			key := fmt.Sprintf("%s.%s", pkgModel.PackageName, sym.Name)
			if sym.ReceiverType != "" {
				recvClean := strings.TrimPrefix(sym.ReceiverType, "*")
				key = fmt.Sprintf("%s.%s.%s", pkgModel.PackageName, recvClean, sym.Name)
			}
			r.allSymbols[key] = sym
		}
	}
}

// BuildCallGraph constructs a bounded call graph for the given package model.
func (r *CallSiteResolver) BuildCallGraph(pkgModel *PackageSemanticModel) *BoundedCallGraph {
	if pkgModel == nil {
		return nil
	}

	cg := NewBoundedCallGraph(pkgModel.PackageName, pkgModel.PackagePath)

	for relFile, fModel := range pkgModel.Files {
		fset := fModel.FileSet

		// Discover package imports and aliases in this file
		importAliases := make(map[string]string) // alias -> import path
		if fModel.ASTFile != nil {
			for _, imp := range fModel.ASTFile.Imports {
				pathVal := strings.Trim(imp.Path.Value, `"`)
				alias := filepath.Base(pathVal)
				if imp.Name != nil {
					alias = imp.Name.Name
				}
				importAliases[alias] = pathVal
			}
		}

		for fnKey, cfg := range fModel.CFGs {
			if cfg.FuncDecl == nil || cfg.FuncDecl.Body == nil {
				continue
			}

			ast.Inspect(cfg.FuncDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				pos := fset.Position(call.Pos())
				snippet := exprToString(call, fset)

				target := r.ResolveCall(call, relFile, fnKey, fModel, pkgModel, importAliases)

				edge := &CallEdge{
					CallerPackage:  pkgModel.PackageName,
					CallerFile:     relFile,
					CallerFunction: fnKey,
					CallSiteLine:   pos.Line,
					CallSnippet:    snippet,
					Target:         target,
					CallExpr:       call,
				}

				cg.Edges = append(cg.Edges, edge)
				cg.Outgoing[fnKey] = append(cg.Outgoing[fnKey], edge)

				targetKey := fmt.Sprintf("%s.%s", target.PackageName, target.FunctionName)
				cg.Incoming[targetKey] = append(cg.Incoming[targetKey], edge)
				cg.Targets[targetKey] = target

				return true
			})
		}
	}

	return cg
}

// ResolveCall resolves an ast.CallExpr to a structured CallTarget.
func (r *CallSiteResolver) ResolveCall(
	call *ast.CallExpr,
	callerFile, callerFn string,
	fModel *FileSemanticModel,
	pkgModel *PackageSemanticModel,
	importAliases map[string]string,
) *CallTarget {
	fset := fModel.FileSet

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// 1. Local or same-package function call: foo(...)
		fnName := fun.Name

		// Check local file declarations
		if fModel != nil {
			if cfg, ok := fModel.CFGs[fnName]; ok && cfg.FuncDecl != nil {
				return &CallTarget{
					PackageName:        pkgModel.PackageName,
					PackagePath:        pkgModel.PackagePath,
					FunctionName:       fnName,
					File:               callerFile,
					Line:               fset.Position(cfg.FuncDecl.Pos()).Line,
					ResolutionState:    CallResolvedExact,
					IsResourceConsumer: aCheckResourceConsumer(cfg.FuncDecl, fset),
					IsErrorPropagator:  aCheckErrorPropagator(cfg.FuncDecl),
				}
			}
		}

		// Check other files in same package
		if pkgModel != nil {
			for otherFile, otherFModel := range pkgModel.Files {
				if cfg, ok := otherFModel.CFGs[fnName]; ok && cfg.FuncDecl != nil {
					return &CallTarget{
						PackageName:        pkgModel.PackageName,
						PackagePath:        pkgModel.PackagePath,
						FunctionName:       fnName,
						File:               otherFile,
						Line:               otherFModel.FileSet.Position(cfg.FuncDecl.Pos()).Line,
						ResolutionState:    CallResolvedExact,
						IsResourceConsumer: aCheckResourceConsumer(cfg.FuncDecl, otherFModel.FileSet),
						IsErrorPropagator:  aCheckErrorPropagator(cfg.FuncDecl),
					}
				}
			}
		}

		// Built-in or unresolvable identifier
		return &CallTarget{
			PackageName:     "builtin",
			FunctionName:    fnName,
			ResolutionState: CallResolvedConservative,
		}

	case *ast.SelectorExpr:
		methodName := fun.Sel.Name

		// Case A: Package-qualified call: pkg.Func(...)
		if id, ok := fun.X.(*ast.Ident); ok {
			pkgIdent := id.Name

			// Check if pkgIdent is an imported package in this file
			if impPath, ok := importAliases[pkgIdent]; ok {
				// Check if imported package exists in our registered project packages
				if impPkg, ok := r.pkgModels[impPath]; ok {
					for impFile, impFModel := range impPkg.Files {
						if cfg, ok := impFModel.CFGs[methodName]; ok && cfg.FuncDecl != nil {
							return &CallTarget{
								PackageName:        impPkg.PackageName,
								PackagePath:        impPkg.PackagePath,
								FunctionName:       methodName,
								File:               impFile,
								Line:               impFModel.FileSet.Position(cfg.FuncDecl.Pos()).Line,
								ResolutionState:    CallResolvedExact,
								IsResourceConsumer: aCheckResourceConsumer(cfg.FuncDecl, impFModel.FileSet),
								IsErrorPropagator:  aCheckErrorPropagator(cfg.FuncDecl),
							}
						}
					}
				}

				// Imported external standard library or third-party dependency
				return &CallTarget{
					PackageName:        pkgIdent,
					PackagePath:        impPath,
					FunctionName:       methodName,
					ResolutionState:    CallExternalOpaque,
					IsResourceConsumer: strings.Contains(methodName, "Close") || strings.Contains(methodName, "consume"),
					IsErrorPropagator:  strings.Contains(methodName, "wrap") || strings.Contains(methodName, "check"),
				}
			}

			// Case B: Method call on local variable receiver: obj.Method(...)
			varName := pkgIdent
			recvType := r.estimateReceiverType(varName, callerFn, fModel)
			if recvType != "" {
				targetKey := fmt.Sprintf("%s.%s", strings.TrimPrefix(recvType, "*"), methodName)
				// Look up in same package
				if pkgModel != nil {
					for otherFile, otherFModel := range pkgModel.Files {
						if cfg, ok := otherFModel.CFGs[targetKey]; ok && cfg.FuncDecl != nil {
							return &CallTarget{
								PackageName:        pkgModel.PackageName,
								PackagePath:        pkgModel.PackagePath,
								FunctionName:       methodName,
								ReceiverType:       recvType,
								File:               otherFile,
								Line:               otherFModel.FileSet.Position(cfg.FuncDecl.Pos()).Line,
								ResolutionState:    CallResolvedExact,
								IsResourceConsumer: aCheckResourceConsumer(cfg.FuncDecl, otherFModel.FileSet),
								IsErrorPropagator:  aCheckErrorPropagator(cfg.FuncDecl),
							}
						}
					}
				}

				return &CallTarget{
					PackageName:     pkgModel.PackageName,
					FunctionName:    methodName,
					ReceiverType:    recvType,
					ResolutionState: CallResolvedConservative,
				}
			}
		}

		// Chained method call or complex expression: obj.Field.Method(...)
		return &CallTarget{
			FunctionName:    methodName,
			ResolutionState: CallResolvedConservative,
			Summary:         exprToString(fun.X, fset),
		}
	}

	return &CallTarget{
		FunctionName:    exprToString(call.Fun, fset),
		ResolutionState: CallUnresolved,
	}
}

// estimateReceiverType scans local variable definitions or parameters to find the receiver type string.
func (r *CallSiteResolver) estimateReceiverType(varName, fnKey string, fModel *FileSemanticModel) string {
	if fModel == nil {
		return ""
	}
	cfg, ok := fModel.CFGs[fnKey]
	if !ok || cfg.FuncDecl == nil {
		return ""
	}

	// Check parameters
	if cfg.FuncDecl.Type != nil && cfg.FuncDecl.Type.Params != nil {
		for _, param := range cfg.FuncDecl.Type.Params.List {
			for _, name := range param.Names {
				if name.Name == varName {
					return formatTypeExpr(param.Type, fModel.FileSet)
				}
			}
		}
	}

	// Check local short declarations in function body
	var foundType string
	ast.Inspect(cfg.FuncDecl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name == varName {
				if i < len(assign.Rhs) {
					// Check composite literal: v := &MyType{...}
					if u, ok := assign.Rhs[i].(*ast.UnaryExpr); ok && u.Op == token.AND {
						if comp, ok := u.X.(*ast.CompositeLit); ok {
							foundType = "*" + formatTypeExpr(comp.Type, fModel.FileSet)
							return false
						}
					}
				}
			}
		}
		return foundType == ""
	})

	return foundType
}

// aCheckResourceConsumer inspects whether a function signature or body takes a resource and releases it.
func aCheckResourceConsumer(fnDecl *ast.FuncDecl, fset *token.FileSet) bool {
	if fnDecl == nil || fnDecl.Body == nil {
		return false
	}
	fnName := strings.ToLower(fnDecl.Name.Name)
	if strings.HasPrefix(fnName, "consume") || strings.HasPrefix(fnName, "close") || strings.HasPrefix(fnName, "transfer") {
		return true
	}

	hasClose := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Close" {
				hasClose = true
				return false
			}
		}
		return true
	})
	return hasClose
}

// aCheckErrorPropagator inspects whether a function returns error or handles errors.
func aCheckErrorPropagator(fnDecl *ast.FuncDecl) bool {
	if fnDecl == nil || fnDecl.Type == nil || fnDecl.Type.Results == nil {
		return false
	}
	for _, res := range fnDecl.Type.Results.List {
		if typeExprToString(res.Type) == "error" {
			return true
		}
	}
	return false
}

