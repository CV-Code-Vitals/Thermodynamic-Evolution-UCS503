package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"sync"
	"time"
)

// SemanticEffectType classifies interprocedural dataflow and lifecycle actions.
type SemanticEffectType string

const (
	EffectTaintPassthrough     SemanticEffectType = "TAINT_PASSTHROUGH"
	EffectTaintTransform       SemanticEffectType = "TAINT_TRANSFORM"
	EffectTaintSanitizeVerified SemanticEffectType = "TAINT_SANITIZE_VERIFIED"
	EffectResourceAcquire      SemanticEffectType = "RESOURCE_ACQUIRE"
	EffectResourceRelease      SemanticEffectType = "RESOURCE_RELEASE"
	EffectResourceTransfer     SemanticEffectType = "RESOURCE_TRANSFER"
	EffectErrorReturn          SemanticEffectType = "ERROR_RETURN"
	EffectErrorPropagate       SemanticEffectType = "ERROR_PROPAGATE"
	EffectUnknown              SemanticEffectType = "EFFECT_UNKNOWN"
)

// ParameterReturnFlow records dataflow taint relationships between input parameters and return tuple elements.
type ParameterReturnFlow struct {
	ParamIndex  int                `json:"param_index"`
	ReturnIndex int                `json:"return_index"`
	Effect      SemanticEffectType `json:"effect"`
	TaintDomain string             `json:"taint_domain"` // "SQL", "SHELL", "RAW"
	Confidence  float64            `json:"confidence"`
	Evidence    string             `json:"evidence"`
}

// ResourceSummaryEffect captures resource lifecycle actions performed by a function.
type ResourceSummaryEffect struct {
	Effect       SemanticEffectType `json:"effect"`
	ResourceKind ResourceKind       `json:"resource_kind"`
	ParamIndex   int                `json:"param_index"`  // Parameter released or transferred (-1 if none)
	ReturnIndex  int                `json:"return_index"` // Return tuple index where new resource is produced (-1 if none)
	Condition    string             `json:"condition"`
	Evidence     string             `json:"evidence"`
}

// FunctionSummary is the canonical interprocedural behavioral summary of a function or method.
type FunctionSummary struct {
	PackageName        string                  `json:"package_name"`
	PackagePath        string                  `json:"package_path"`
	FunctionName       string                  `json:"function_name"`
	ReceiverType       string                  `json:"receiver_type,omitempty"`
	ParamCount         int                     `json:"param_count"`
	ReturnCount        int                     `json:"return_count"`
	ParamNames         []string                `json:"param_names"`
	ParamReturnFlows   []ParameterReturnFlow   `json:"param_return_flows"`
	ResourceEffects    []ResourceSummaryEffect `json:"resource_effects"`
	ProducesError      bool                    `json:"produces_error"`
	ErrorReturnIndex   int                     `json:"error_return_index"`
	IsComplete         bool                    `json:"is_complete"`
	Confidence         float64                 `json:"confidence"`
	SourceHash         string                  `json:"source_hash"`
	AnalysisDurationUs int64                   `json:"analysis_duration_us"`
	IterationCount     int                     `json:"iteration_count"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// FunctionSummaryEngine manages the extraction, composition, and caching of function summaries.
type FunctionSummaryEngine struct {
	mu          sync.RWMutex
	repoRoot    string
	summaries   map[string]*FunctionSummary // "pkg.func" or "pkg.recv.func" -> Summary
	resRegistry *ResourceRegistry
}

// NewFunctionSummaryEngine initializes an interprocedural summary engine.
func NewFunctionSummaryEngine(repoRoot string) *FunctionSummaryEngine {
	return &FunctionSummaryEngine{
		repoRoot:    repoRoot,
		summaries:   make(map[string]*FunctionSummary),
		resRegistry: DefaultResourceRegistry(),
	}
}

// SummaryKey generates a canonical map key for a function or method.
func SummaryKey(pkgName, recvType, fnName string) string {
	if recvType != "" {
		recvClean := strings.TrimPrefix(recvType, "*")
		return fmt.Sprintf("%s.%s.%s", pkgName, recvClean, fnName)
	}
	return fmt.Sprintf("%s.%s", pkgName, fnName)
}

// GetSummary retrieves a function summary if available.
func (e *FunctionSummaryEngine) GetSummary(pkgName, recvType, fnName string) *FunctionSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Direct lookup
	key := SummaryKey(pkgName, recvType, fnName)
	if s, ok := e.summaries[key]; ok {
		return s
	}

	// Try without receiver
	if recvType != "" {
		fallbackKey := fmt.Sprintf("%s.%s", pkgName, fnName)
		if s, ok := e.summaries[fallbackKey]; ok {
			return s
		}
	}

	// Fallback lookup: search by function name if package name matches suffix
	for k, s := range e.summaries {
		if strings.HasSuffix(k, "."+fnName) {
			if pkgName == "" || strings.Contains(k, pkgName) {
				return s
			}
		}
	}

	return nil
}

// RegisterSummary stores a computed summary in the registry.
func (e *FunctionSummaryEngine) RegisterSummary(s *FunctionSummary) {
	if s == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	key := SummaryKey(s.PackageName, s.ReceiverType, s.FunctionName)
	e.summaries[key] = s
}

// ExtractIntraproceduralSummary analyzes a single function AST and CFG to build its initial summary.
func (e *FunctionSummaryEngine) ExtractIntraproceduralSummary(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	df *ReachingDefinitionsAnalysis,
	pkgName, pkgPath string,
) *FunctionSummary {
	t0 := time.Now()
	if fnDecl == nil || fnDecl.Name == nil {
		return nil
	}

	fnName := fnDecl.Name.Name
	recvType := ""
	if fnDecl.Recv != nil && len(fnDecl.Recv.List) > 0 {
		recvType = formatTypeExpr(fnDecl.Recv.List[0].Type, fset)
	}

	summary := &FunctionSummary{
		PackageName:      pkgName,
		PackagePath:      pkgPath,
		FunctionName:     fnName,
		ReceiverType:     recvType,
		ParamReturnFlows: make([]ParameterReturnFlow, 0),
		ResourceEffects:  make([]ResourceSummaryEffect, 0),
		ErrorReturnIndex: -1,
		IsComplete:       true,
		Confidence:       0.95,
		UpdatedAt:        time.Now(),
		IterationCount:   1,
	}

	// Compute source hash
	var buf strings.Builder
	if fnDecl.Body != nil {
		_ = printer.Fprint(&buf, fset, fnDecl.Body)
	}
	h := sha256.Sum256([]byte(buf.String()))
	summary.SourceHash = hex.EncodeToString(h[:])

	// Inspect parameters
	paramMap := make(map[string]int) // paramName -> index
	if fnDecl.Type.Params != nil {
		pIdx := 0
		for _, field := range fnDecl.Type.Params.List {
			for _, name := range field.Names {
				paramMap[name.Name] = pIdx
				summary.ParamNames = append(summary.ParamNames, name.Name)
				pIdx++
			}
		}
		summary.ParamCount = pIdx
	}

	// Inspect returns
	if fnDecl.Type.Results != nil {
		rIdx := 0
		for _, field := range fnDecl.Type.Results.List {
			typeStr := formatTypeExpr(field.Type, fset)
			if typeStr == "error" {
				summary.ProducesError = true
				summary.ErrorReturnIndex = rIdx
			}
			rIdx++
		}
		summary.ReturnCount = rIdx
	}

	if fnDecl.Body == nil {
		summary.AnalysisDurationUs = time.Since(t0).Microseconds()
		return summary
	}

	// 1. Analyze Dataflow: Parameters flowing to Return Statements
	e.analyzeReturnDataflows(fnDecl, fset, paramMap, summary)

	// 2. Analyze Resource Effects: Acquisitions & Releases
	e.analyzeResourceEffects(fnDecl, fset, paramMap, summary)

	// 3. Analyze Error Propagation
	e.analyzeErrorPropagation(fnDecl, fset, paramMap, summary)

	summary.AnalysisDurationUs = time.Since(t0).Microseconds()
	return summary
}

// analyzeReturnDataflows tracks parameter flow into return expressions (concatenations, fmt.Sprintf, direct returns).
func (e *FunctionSummaryEngine) analyzeReturnDataflows(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	paramMap map[string]int,
	summary *FunctionSummary,
) {
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for rIdx, retExpr := range ret.Results {
			// A. Direct parameter passthrough: return x
			if id, ok := retExpr.(*ast.Ident); ok {
				if pIdx, isParam := paramMap[id.Name]; isParam {
					summary.ParamReturnFlows = append(summary.ParamReturnFlows, ParameterReturnFlow{
						ParamIndex:  pIdx,
						ReturnIndex: rIdx,
						Effect:      EffectTaintPassthrough,
						TaintDomain: "RAW",
						Confidence:  0.98,
						Evidence:    fmt.Sprintf("Direct return of parameter '%s'", id.Name),
					})
				}
			}

			// B. Binary expression concatenation: return "..." + param + "..."
			if bin, ok := retExpr.(*ast.BinaryExpr); ok {
				operands := collectAdditionLeaves(bin)
				for _, op := range operands {
					if id, ok := op.(*ast.Ident); ok {
						if pIdx, isParam := paramMap[id.Name]; isParam {
							summary.ParamReturnFlows = append(summary.ParamReturnFlows, ParameterReturnFlow{
								ParamIndex:  pIdx,
								ReturnIndex: rIdx,
								Effect:      EffectTaintTransform,
								TaintDomain: "SQL", // Conservative potential SQL/command string concatenation
								Confidence:  0.95,
								Evidence:    fmt.Sprintf("Parameter '%s' concatenated into return string", id.Name),
							})
						}
					}
				}
			}

			// C. Formatting function: return fmt.Sprintf("...", param)
			if call, ok := retExpr.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "fmt" && (sel.Sel.Name == "Sprintf" || sel.Sel.Name == "Sprint") {
						for _, arg := range call.Args {
							if id, ok := arg.(*ast.Ident); ok {
								if pIdx, isParam := paramMap[id.Name]; isParam {
									summary.ParamReturnFlows = append(summary.ParamReturnFlows, ParameterReturnFlow{
										ParamIndex:  pIdx,
										ReturnIndex: rIdx,
										Effect:      EffectTaintTransform,
										TaintDomain: "SQL",
										Confidence:  0.96,
										Evidence:    fmt.Sprintf("Parameter '%s' formatted via fmt.Sprintf into return string", id.Name),
									})
								}
							}
						}
					}
				}
			}
		}

		return true
	})
}

// analyzeResourceEffects detects resource acquisitions and releases inside the function.
func (e *FunctionSummaryEngine) analyzeResourceEffects(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	paramMap map[string]int,
	summary *FunctionSummary,
) {
	// 1. Detect Resource Acquisition Returned to Caller
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for rIdx, retExpr := range ret.Results {
			// Direct return of acquisition call: return os.Open(...)
			if call, ok := retExpr.(*ast.CallExpr); ok {
				contract, _, matched := e.resRegistry.MatchAcquisition(call, fset)
				if matched {
					summary.ResourceEffects = append(summary.ResourceEffects, ResourceSummaryEffect{
						Effect:       EffectResourceAcquire,
						ResourceKind: contract.ResourceKind,
						ParamIndex:   -1,
						ReturnIndex:  rIdx,
						Evidence:     fmt.Sprintf("Acquires and returns %s via %s", contract.ResourceKind, contract.MethodName),
					})
				}
			}

			// Return of local variable: f, _ := os.Open(...); return f
			if id, ok := retExpr.(*ast.Ident); ok {
				// Trace where id was assigned
				ast.Inspect(fnDecl.Body, func(inner ast.Node) bool {
					assign, ok := inner.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assign.Lhs {
						if lhsId, ok := lhs.(*ast.Ident); ok && lhsId.Name == id.Name {
							if i < len(assign.Rhs) {
								if call, isCall := assign.Rhs[i].(*ast.CallExpr); isCall {
									contract, _, matched := e.resRegistry.MatchAcquisition(call, fset)
									if matched {
										summary.ResourceEffects = append(summary.ResourceEffects, ResourceSummaryEffect{
											Effect:       EffectResourceAcquire,
											ResourceKind: contract.ResourceKind,
											ParamIndex:   -1,
											ReturnIndex:  rIdx,
											Evidence:     fmt.Sprintf("Acquires %s via %s and returns at index %d", contract.ResourceKind, contract.MethodName, rIdx),
										})
									}
								}
							}
						}
					}
					return true
				})
			}
		}
		return true
	})

	// 2. Detect Resource Release of Function Parameter: e.g. f.Close() or timer.Stop()
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			methodName := sel.Sel.Name
			if id, ok := sel.X.(*ast.Ident); ok {
				if pIdx, isParam := paramMap[id.Name]; isParam {
					// Check if method is a release method
					if methodName == "Close" || methodName == "Stop" {
						summary.ResourceEffects = append(summary.ResourceEffects, ResourceSummaryEffect{
							Effect:       EffectResourceRelease,
							ResourceKind: ResourceKindFile, // Generic system resource
							ParamIndex:   pIdx,
							ReturnIndex:  -1,
							Evidence:     fmt.Sprintf("Invokes %s() on parameter '%s'", methodName, id.Name),
						})
					}
				}
			}
		}
		return true
	})
}

// analyzeErrorPropagation detects error wrapping and propagation.
func (e *FunctionSummaryEngine) analyzeErrorPropagation(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	paramMap map[string]int,
	summary *FunctionSummary,
) {
	if !summary.ProducesError {
		return
	}

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for rIdx, retExpr := range ret.Results {
			// Direct error return: return err
			if id, ok := retExpr.(*ast.Ident); ok {
				if pIdx, isParam := paramMap[id.Name]; isParam {
					summary.ParamReturnFlows = append(summary.ParamReturnFlows, ParameterReturnFlow{
						ParamIndex:  pIdx,
						ReturnIndex: rIdx,
						Effect:      EffectErrorPropagate,
						TaintDomain: "ERROR",
						Confidence:  0.98,
						Evidence:    fmt.Sprintf("Propagates incoming error parameter '%s'", id.Name),
					})
				}
			}

			// Wrapped error: return fmt.Errorf("...", err)
			if call, ok := retExpr.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "fmt" && sel.Sel.Name == "Errorf" {
						for _, arg := range call.Args {
							if id, ok := arg.(*ast.Ident); ok {
								if pIdx, isParam := paramMap[id.Name]; isParam {
									summary.ParamReturnFlows = append(summary.ParamReturnFlows, ParameterReturnFlow{
										ParamIndex:  pIdx,
										ReturnIndex: rIdx,
										Effect:      EffectErrorPropagate,
										TaintDomain: "ERROR",
										Confidence:  0.96,
										Evidence:    fmt.Sprintf("Wraps incoming error parameter '%s' via fmt.Errorf", id.Name),
									})
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}

// BuildSummariesBottomUp computes summaries for all functions in the repository in bottom-up call order.
func (e *FunctionSummaryEngine) BuildSummariesBottomUp(
	pkgModels map[string]*PackageSemanticModel,
	callGraphs map[string]*BoundedCallGraph,
) {
	// Step 1: Extract initial intraprocedural summaries for all functions
	for _, pm := range pkgModels {
		for _, fModel := range pm.Files {
			for fnName, cfg := range fModel.CFGs {
				if cfg.FuncDecl == nil {
					continue
				}
				s := e.ExtractIntraproceduralSummary(cfg.FuncDecl, fModel.FileSet, cfg, fModel.Dataflow[fnName], pm.PackageName, pm.PackagePath)
				if s != nil {
					e.RegisterSummary(s)
				}
			}
		}
	}

	// Step 2: Bounded Composition & Propagation (up to 3 iterations for SCCs / call chains)
	maxIterations := 3
	for iter := 1; iter <= maxIterations; iter++ {
		changed := false

		for _, pm := range pkgModels {
			for _, fModel := range pm.Files {
				for fnName, cfg := range fModel.CFGs {
					if cfg.FuncDecl == nil || cfg.FuncDecl.Body == nil {
						continue
					}

					key := SummaryKey(pm.PackageName, "", fnName)
					callerSummary := e.GetSummary(pm.PackageName, "", fnName)
					if callerSummary == nil {
						continue
					}

					// Compose callee summaries into caller
					ast.Inspect(cfg.FuncDecl.Body, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}

						// Look up callee summary
						var calleeSummary *FunctionSummary
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if id, ok := sel.X.(*ast.Ident); ok {
								calleeSummary = e.GetSummary(id.Name, "", sel.Sel.Name)
							}
						} else if id, ok := call.Fun.(*ast.Ident); ok {
							calleeSummary = e.GetSummary(pm.PackageName, "", id.Name)
						}

						if calleeSummary != nil {
							// Propagate resource release
							for _, rEff := range calleeSummary.ResourceEffects {
								if rEff.Effect == EffectResourceRelease && rEff.ParamIndex < len(call.Args) {
									// Caller passes argument that callee releases!
									argExpr := call.Args[rEff.ParamIndex]
									if argId, ok := argExpr.(*ast.Ident); ok {
										for pIdx, pName := range callerSummary.ParamNames {
											if pName == argId.Name {
												// Caller also releases parameter
												newEff := ResourceSummaryEffect{
													Effect:       EffectResourceRelease,
													ResourceKind: rEff.ResourceKind,
													ParamIndex:   pIdx,
													ReturnIndex:  -1,
													Evidence:     fmt.Sprintf("Transitively released via callee '%s'", calleeSummary.FunctionName),
												}
												if !containsResourceEffect(callerSummary.ResourceEffects, newEff) {
													callerSummary.ResourceEffects = append(callerSummary.ResourceEffects, newEff)
													changed = true
												}
											}
										}
									}
								}
							}
						}

						return true
					})

					callerSummary.IterationCount = iter
					_ = key
				}
			}
		}

		if !changed {
			break
		}
	}
}

func containsResourceEffect(effects []ResourceSummaryEffect, target ResourceSummaryEffect) bool {
	for _, eff := range effects {
		if eff.Effect == target.Effect && eff.ParamIndex == target.ParamIndex && eff.ReturnIndex == target.ReturnIndex {
			return true
		}
	}
	return false
}
