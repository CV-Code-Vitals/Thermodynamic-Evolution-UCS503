package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ResourceAnalyzer performs CFG-aware, function-scoped resource lifecycle and leak analysis.
type ResourceAnalyzer struct {
	repoRoot      string
	registry      *ResourceRegistry
	summaryEngine *FunctionSummaryEngine
	hocAnalyzer   *HigherOrderCleanupAnalyzer
	fieldTracker  *FieldSensitiveTracker
	aliasGraph    *AliasGraph
}

// SetSummaryEngine attaches an interprocedural function summary engine.
func (a *ResourceAnalyzer) SetSummaryEngine(se *FunctionSummaryEngine) {
	a.summaryEngine = se
	if a.hocAnalyzer != nil {
		a.hocAnalyzer.summaryEngine = se
	}
}

// NewResourceAnalyzer creates a new ResourceAnalyzer instance with the canonical resource registry.
func NewResourceAnalyzer(repoRoot string) *ResourceAnalyzer {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	reg := DefaultResourceRegistry()
	return &ResourceAnalyzer{
		repoRoot:     repoRoot,
		registry:     reg,
		hocAnalyzer:  NewHigherOrderCleanupAnalyzer(reg, nil),
		fieldTracker: NewFieldSensitiveTracker(3),
	}
}

// AnalyzeFunction inspects a single function or method for resource lifecycle management.
func (a *ResourceAnalyzer) AnalyzeFunction(
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
	reachingDefs *ReachingDefinitionsAnalysis,
	filePath string,
) ([]ResourceEvidence, error) {
	var results []ResourceEvidence
	if fnDecl == nil || fnDecl.Body == nil || cfg == nil {
		return results, nil
	}

	fnName := fnDecl.Name.Name
	pkgName := ""

	// 1. Discover all resource acquisition calls in the function
	instances := a.discoverAcquisitions(fnDecl, fset, cfg)
	if len(instances) == 0 {
		return results, nil
	}

	// 2. Discover variable aliases for each resource
	a.discoverAliases(fnDecl, fset, instances)

	// 3. Register local functions for higher-order wrapper inspection
	if a.hocAnalyzer != nil {
		a.hocAnalyzer.RegisterLocalFunc(fnDecl)
	}

	// 4. Track field-sensitive assignments in function body
	a.fieldTracker = NewFieldSensitiveTracker(3)
	knownResources := make(map[string]ResourceKind)
	for _, inst := range instances {
		knownResources[inst.VarName] = inst.Kind
	}
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for i := 0; i < len(stmt.Lhs) && i < len(stmt.Rhs); i++ {
				a.fieldTracker.TrackAssignment(stmt.Lhs[i], stmt.Rhs[i], fset, knownResources)
				if lhsId, ok := stmt.Lhs[i].(*ast.Ident); ok {
					if rhsId, ok := stmt.Rhs[i].(*ast.Ident); ok {
						a.fieldTracker.TrackStructAlias(lhsId.Name, rhsId.Name)
					}
				}
			}
		case *ast.ReturnStmt:
			a.fieldTracker.TrackReturnTransfer(stmt)
		}
		return true
	})

	// 5. Check for deferred releases in the CFG
	a.evaluateDeferredReleases(cfg, instances, fset)

	// 6. CFG Reachability & Release Analysis per instance
	for i := range instances {
		inst := &instances[i]
		a.evaluateInstanceLifecycle(inst, fnDecl, fset, cfg)

		ev := a.constructEvidence(inst, filePath, pkgName, fnName)
		results = append(results, ev)
	}

	return results, nil
}

// discoverAcquisitions scans statements in the function body for resource-producing calls.
func (a *ResourceAnalyzer) discoverAcquisitions(fnDecl *ast.FuncDecl, fset *token.FileSet, cfg *ControlFlowGraph) []ResourceInstance {
	var instances []ResourceInstance
	instCounter := 0

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, rhs := range assign.Rhs {
			call, isCall := rhs.(*ast.CallExpr)
			if !isCall {
				continue
			}

			var resKind ResourceKind
			var resIdx int

			contract, _, matched := a.registry.MatchAcquisition(call, fset)
			if matched && contract != nil {
				resKind = contract.ResourceKind
				resIdx = contract.ResultIndex
			} else if a.summaryEngine != nil {
				var calleeSummary *FunctionSummary
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok {
						calleeSummary = a.summaryEngine.GetSummary(id.Name, "", sel.Sel.Name)
					}
				} else if id, ok := call.Fun.(*ast.Ident); ok {
					calleeSummary = a.summaryEngine.GetSummary("", "", id.Name)
				}

				if calleeSummary != nil {
					for _, eff := range calleeSummary.ResourceEffects {
						if eff.Effect == EffectResourceAcquire {
							matched = true
							resKind = eff.ResourceKind
							resIdx = eff.ReturnIndex
							break
						}
					}
				}
			}

			if !matched {
				continue
			}

			if resIdx >= len(assign.Lhs) {
				continue
			}

			lhsIdent, ok := assign.Lhs[resIdx].(*ast.Ident)
			if !ok || lhsIdent.Name == "_" {
				continue
			}

			pos := fset.Position(call.Pos())
			blockID := a.findBlockIDForPos(call.Pos(), cfg)

			inst := ResourceInstance{
				ID:                 instCounter,
				Kind:               resKind,
				VarName:            lhsIdent.Name,
				Aliases:            make([]string, 0),
				AcquisitionCall:    exprToString(call, fset),
				AcquisitionLine:    pos.Line,
				AcquisitionBlockID: blockID,
				CurrentState:       StateResourceAcquired,
				History:            make([]ResourceTransition, 0),
			}
			inst.AddTransition(StateResourceAcquired, inst.AcquisitionCall, pos.Line, blockID, "Acquired via supported API contract")
			instCounter++
			instances = append(instances, inst)
		}

		return true
	})

	return instances
}

// discoverAliases finds multi-hop variable aliases and reassignments using AliasGraph.
func (a *ResourceAnalyzer) discoverAliases(fnDecl *ast.FuncDecl, fset *token.FileSet, instances []ResourceInstance) {
	a.aliasGraph = BuildAliasGraphFromAST(fnDecl, fset, 5, 32)
	for i := range instances {
		transitiveAliases := a.aliasGraph.GetAliases(instances[i].VarName)
		seen := make(map[string]bool)
		for _, al := range instances[i].Aliases {
			seen[al] = true
		}
		for _, al := range transitiveAliases {
			if !seen[al] && al != instances[i].VarName {
				instances[i].Aliases = append(instances[i].Aliases, al)
				seen[al] = true
			}
		}
	}
}

// evaluateDeferredReleases checks whether defer statements in the function release the resource.
func (a *ResourceAnalyzer) evaluateDeferredReleases(cfg *ControlFlowGraph, instances []ResourceInstance, fset *token.FileSet) {
	for _, deferSite := range cfg.AllDefers {
		if deferSite.CallExpr == nil {
			continue
		}

		for i := range instances {
			inst := &instances[i]

			// 1. Direct release call
			if a.registry.IsReleaseCall(deferSite.CallExpr, inst.Kind, inst.VarName, inst.Aliases) {
				if deferSite.InLoop {
					inst.DeferredInLoop = true
					inst.AddTransition(StateResourceLeakProbable, "defer "+deferSite.TargetFunc+"()", deferSite.Line, 0, "Deferred release inside loop delays closure until function exit")
				} else {
					inst.IsDeferred = true
					inst.AddTransition(StateResourceDeferredRelease, "defer "+deferSite.TargetFunc+"()", deferSite.Line, 0, "Scheduled deferred release on function return")
				}
				continue
			}

			// 2. Higher-Order Cleanup / Method Values
			if a.hocAnalyzer != nil {
				res := a.hocAnalyzer.AnalyzeDeferredCall(deferSite.CallExpr, inst, fset)
				if res.IsVerifiedRelease {
					if deferSite.InLoop {
						inst.DeferredInLoop = true
						inst.AddTransition(StateResourceLeakProbable, "defer "+res.WrapperName+"()", deferSite.Line, 0, "Deferred release inside loop delays closure until function exit")
					} else {
						inst.IsDeferred = true
						inst.AddTransition(StateResourceDeferredRelease, "defer "+res.WrapperName+"("+res.MethodName+")", deferSite.Line, 0, res.Evidence)
					}
					continue
				}
			}

			// 3. Field-sensitive struct release: defer h.file.Close()
			if a.fieldTracker != nil && a.fieldTracker.TrackReleaseCall(deferSite.CallExpr, fset) {
				for path, field := range a.fieldTracker.Fields {
					if field.OriginalVar == inst.VarName && field.State == StateFieldReleased {
						if deferSite.InLoop {
							inst.DeferredInLoop = true
							inst.AddTransition(StateResourceLeakProbable, "defer "+path+".Close()", deferSite.Line, 0, "Deferred release inside loop delays closure until function exit")
						} else {
							inst.IsDeferred = true
							inst.AddTransition(StateResourceDeferredRelease, "defer "+path+".Close()", deferSite.Line, 0, "Deferred struct field release")
						}
						break
					}
				}
			}
		}
	}
}

// evaluateInstanceLifecycle performs forward CFG path reachability to verify release along all exit paths.
func (a *ResourceAnalyzer) evaluateInstanceLifecycle(
	inst *ResourceInstance,
	fnDecl *ast.FuncDecl,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
) {
	// If deferred at function level outside of loops, it executes on every exit
	if inst.IsDeferred && !inst.DeferredInLoop {
		inst.ReachableExits = len(cfg.ReturnBlocks)
		inst.CleanExits = len(cfg.ReturnBlocks)
		inst.CurrentState = StateResourceDeferredRelease
		return
	}

	// If deferred inside a loop, it leaks descriptors across iterations
	if inst.DeferredInLoop {
		inst.CurrentState = StateResourceLeakProbable
		return
	}

	// Check if resource is transferred to an external helper or struct
	if a.isResourceTransferredOrStored(fnDecl, inst, fset) {
		return
	}

	acqLine := inst.AcquisitionLine
	unacquiredExits := 0
	cleanExits := 0
	leakedExits := 0

	// Inspect each return block in the CFG
	for _, retBlock := range cfg.ReturnBlocks {
		inst.ReachableExits++

		// Check if this return statement is on an acquisition-failure error path
		if a.isAcquisitionErrorExit(retBlock, acqLine, fnDecl, fset) {
			unacquiredExits++
			continue
		}

		// Check if the resource is returned to caller (ownership transfer)
		if a.isResourceReturnedDirectly(retBlock, inst.VarName, inst.Aliases) {
			cleanExits++
			inst.AddTransition(StateResourceTransferred, "return "+inst.VarName, retBlock.LineStart, retBlock.ID, "Ownership transferred to caller via return value")
			continue
		}

		// Check if explicit release call exists along the path reaching this return
		if a.hasExplicitReleaseOnPath(retBlock, inst, fset, cfg) {
			cleanExits++
			inst.AddTransition(StateResourceReleased, "explicit release", retBlock.LineStart, retBlock.ID, "Explicit release executed prior to return")
			continue
		}

		// Path reached exit without release or transfer: LEAK!
		leakedExits++
	}

	inst.CleanExits = cleanExits
	inst.LeakedExits = leakedExits

	// Calculate final lifecycle state
	if leakedExits > 0 {
		if cleanExits == 0 {
			inst.CurrentState = StateResourceLeakConfirmed
		} else {
			inst.CurrentState = StateResourceLeakProbable
		}
	} else if cleanExits > 0 {
		if inst.CurrentState == StateResourceAcquired {
			inst.CurrentState = StateResourceReleased
		}
	} else {
		// All exits were unacquired error exits
		if inst.CurrentState == StateResourceAcquired {
			inst.CurrentState = StateResourceReleased
		}
	}
}

// isResourceTransferredOrStored checks if the resource was passed to a consumer or stored in a struct.
func (a *ResourceAnalyzer) isResourceTransferredOrStored(fnDecl *ast.FuncDecl, inst *ResourceInstance, fset *token.FileSet) bool {
	// 1. Check FieldSensitiveTracker for struct field ownership
	if a.fieldTracker != nil {
		for path, field := range a.fieldTracker.Fields {
			if field.OriginalVar == inst.VarName {
				inst.AccessPath = path
				if field.State == StateFieldReleased {
					inst.CurrentState = StateResourceReleased
					inst.AddTransition(StateResourceReleased, "struct_field_release", field.ReleasedLine, 0, field.Evidence)
					return true
				}
				if field.State == StateOwnedTransferred {
					inst.CurrentState = StateResourceTransferred
					inst.AddTransition(StateResourceTransferred, "struct_return_transfer", 0, 0, field.Evidence)
					return true
				}
				if field.State == StateOwnedLocal {
					// Owned locally in struct; do not abort analysis, allow CFG exit reachability to verify release
					return false
				}
			}
		}
	}

	isTransferred := false
	isStoredInStruct := false
	isConcurrentEscape := false

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			pos := fset.Position(node.Pos())
			if pos.Line <= inst.AcquisitionLine {
				return true
			}
			// Check if passed to consume* or transfer*
			fnStr := exprToString(node.Fun, fset)
			if strings.HasPrefix(fnStr, "consume") || strings.HasPrefix(fnStr, "transfer") {
				for _, arg := range node.Args {
					if id, ok := arg.(*ast.Ident); ok {
						if id.Name == inst.VarName || a.isAliasMatch(id.Name, inst.Aliases) {
							isTransferred = true
						}
					}
				}
			}

		case *ast.SendStmt:
			// Channel send: ch <- f (concurrency ownership boundary)
			pos := fset.Position(node.Pos())
			if pos.Line > inst.AcquisitionLine {
				if id, ok := node.Value.(*ast.Ident); ok {
					if id.Name == inst.VarName || a.isAliasMatch(id.Name, inst.Aliases) {
						isConcurrentEscape = true
					}
				}
			}

		case *ast.GoStmt:
			// Spawning goroutine: go worker(f)
			pos := fset.Position(node.Pos())
			if pos.Line > inst.AcquisitionLine {
				for _, arg := range node.Call.Args {
					if id, ok := arg.(*ast.Ident); ok {
						if id.Name == inst.VarName || a.isAliasMatch(id.Name, inst.Aliases) {
							isConcurrentEscape = true
						}
					}
				}
			}

		case *ast.CompositeLit:
			pos := fset.Position(node.Pos())
			if pos.Line <= inst.AcquisitionLine {
				return true
			}
			// Check if stored into struct
			for _, elt := range node.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if id, ok := kv.Value.(*ast.Ident); ok && (id.Name == inst.VarName || a.isAliasMatch(id.Name, inst.Aliases)) {
						isStoredInStruct = true
					}
				}
			}
		}
		return true
	})

	if isConcurrentEscape {
		inst.ConcurrencyState = ConcurrencyPossibleEscape
		inst.CurrentState = StateResourceUnresolved
		inst.EscapeReason = "Concurrent resource transfer across channel or goroutine boundary (Manual review required)"
		inst.AddTransition(StateResourceUnresolved, "concurrent_transfer", inst.AcquisitionLine, inst.AcquisitionBlockID, inst.EscapeReason)
		return true
	}

	if isStoredInStruct {
		inst.CurrentState = StateResourceUnresolved
		inst.EscapeReason = "Stored in struct with unverified lifecycle"
		inst.AddTransition(StateResourceUnresolved, "struct_storage", inst.AcquisitionLine, inst.AcquisitionBlockID, inst.EscapeReason)
		return true
	}

	if isTransferred {
		inst.CurrentState = StateResourceTransferred
		inst.AddTransition(StateResourceTransferred, "ownership_transfer", inst.AcquisitionLine, inst.AcquisitionBlockID, "Transferred via consuming function contract")
		return true
	}

	return false
}

// isAcquisitionErrorExit checks whether a return block belongs to an error-check block immediately following acquisition.
func (a *ResourceAnalyzer) isAcquisitionErrorExit(retBlock *Block, acqLine int, fnDecl *ast.FuncDecl, fset *token.FileSet) bool {
	if len(retBlock.Statements) == 0 {
		return false
	}

	retStmt, ok := retBlock.Statements[len(retBlock.Statements)-1].(*ast.ReturnStmt)
	if !ok {
		return false
	}

	isAcqErr := false
	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		if isAcqErr {
			return false
		}
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		ifLine := fset.Position(ifStmt.Pos()).Line
		// The error check must be immediately after acquisition (within 3 lines)
		if ifLine < acqLine || ifLine > acqLine+3 {
			return true
		}

		condStr := exprToString(ifStmt.Cond, fset)
		if strings.Contains(condStr, "!= nil") {
			if retStmt.Pos() >= ifStmt.Body.Pos() && retStmt.End() <= ifStmt.Body.End() {
				isAcqErr = true
				return false
			}
		}
		return true
	})

	return isAcqErr
}

// isResourceReturnedDirectly checks whether a return statement includes the resource or an alias as a direct result.
func (a *ResourceAnalyzer) isResourceReturnedDirectly(retBlock *Block, varName string, aliases []string) bool {
	if len(retBlock.Statements) == 0 {
		return false
	}

	retStmt, ok := retBlock.Statements[len(retBlock.Statements)-1].(*ast.ReturnStmt)
	if !ok {
		return false
	}

	for _, res := range retStmt.Results {
		if id, ok := res.(*ast.Ident); ok {
			if id.Name == varName {
				return true
			}
			for _, al := range aliases {
				if id.Name == al {
					return true
				}
			}
		}
	}

	return false
}

// hasExplicitReleaseOnPath checks whether a release call appears in retBlock or in any predecessor block reachable from acquisition.
func (a *ResourceAnalyzer) hasExplicitReleaseOnPath(
	retBlock *Block,
	inst *ResourceInstance,
	fset *token.FileSet,
	cfg *ControlFlowGraph,
) bool {
	visited := make(map[int]bool)
	var queue []*Block
	queue = append(queue, retBlock)
	visited[retBlock.ID] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Check if curr contains a release call
		for _, stmt := range curr.Statements {
			pos := fset.Position(stmt.Pos())
			if pos.Line <= inst.AcquisitionLine {
				continue
			}

			isRel := false
			ast.Inspect(stmt, func(n ast.Node) bool {
				if _, isGo := n.(*ast.GoStmt); isGo {
					// Asynchronous goroutines do not guarantee synchronous cleanup prior to function return
					return false
				}
				if call, ok := n.(*ast.CallExpr); ok {
					if a.registry.IsReleaseCall(call, inst.Kind, inst.VarName, inst.Aliases) {
						isRel = true
						return false
					}

					// Higher-Order Cleanup / Method Values
					if a.hocAnalyzer != nil {
						res := a.hocAnalyzer.AnalyzeDeferredCall(call, inst, fset)
						if res.IsVerifiedRelease {
							isRel = true
							return false
						}
					}

					// Field-Sensitive Struct Release: h.file.Close()
					if a.fieldTracker != nil && a.fieldTracker.TrackReleaseCall(call, fset) {
						for _, field := range a.fieldTracker.Fields {
							if field.OriginalVar == inst.VarName && field.State == StateFieldReleased {
								isRel = true
								return false
							}
						}
					}

					// Interprocedural Callee Release Check
					if a.summaryEngine != nil {
						var calleeSummary *FunctionSummary
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if id, ok := sel.X.(*ast.Ident); ok {
								calleeSummary = a.summaryEngine.GetSummary(id.Name, "", sel.Sel.Name)
							}
						} else if id, ok := call.Fun.(*ast.Ident); ok {
							calleeSummary = a.summaryEngine.GetSummary("", "", id.Name)
						}

						if calleeSummary != nil {
							for _, eff := range calleeSummary.ResourceEffects {
								if eff.Effect == EffectResourceRelease && eff.ParamIndex < len(call.Args) {
									arg := call.Args[eff.ParamIndex]
									if argId, ok := arg.(*ast.Ident); ok {
										if argId.Name == inst.VarName {
											isRel = true
											return false
										}
										for _, al := range inst.Aliases {
											if argId.Name == al {
												isRel = true
												return false
											}
										}
									}
								}
							}
						}
					}
				}
				return true
			})
			if isRel {
				return true
			}
		}

		if curr.ID == inst.AcquisitionBlockID {
			continue
		}

		for _, pred := range curr.Predecessors {
			if !visited[pred.ID] {
				visited[pred.ID] = true
				queue = append(queue, pred)
			}
		}
	}

	return false
}

// findBlockIDForPos maps an AST position to its containing CFG basic block.
func (a *ResourceAnalyzer) findBlockIDForPos(pos token.Pos, cfg *ControlFlowGraph) int {
	for _, b := range cfg.Blocks {
		for _, stmt := range b.Statements {
			if stmt.Pos() <= pos && pos <= stmt.End() {
				return b.ID
			}
		}
	}
	if cfg.Entry != nil {
		return cfg.Entry.ID
	}
	return 0
}

// constructEvidence builds the final ResourceEvidence record from the analyzed instance.
func (a *ResourceAnalyzer) constructEvidence(
	inst *ResourceInstance,
	filePath, pkgName, fnName string,
) ResourceEvidence {
	ev := ResourceEvidence{
		FilePath:           filePath,
		Package:            pkgName,
		Function:           fnName,
		Line:               inst.AcquisitionLine,
		ResourceKind:       inst.Kind,
		VarName:            inst.VarName,
		LifecycleState:     inst.CurrentState,
		AcquisitionSnippet: inst.AcquisitionCall,
		History:            inst.History,
		Timestamp:          time.Now(),
	}

	switch inst.CurrentState {
	case StateResourceLeakConfirmed:
		ev.FindingState = StateConfirmed
		ev.Severity = repo_intel.SeverityHigh
		ev.Confidence = 0.97
		ev.Title = fmt.Sprintf("Unclosed System Resource Leak (%s)", inst.Kind)
		ev.Description = fmt.Sprintf("Resource handle '%s' of type %s acquired in function '%s' at line %d is never closed along normal execution paths.",
			inst.VarName, inst.Kind, fnName, inst.AcquisitionLine)
		ev.RootCause = "Resource handle acquired without explicit or deferred release call along reachable exit paths."
		ev.RecommendedAction = fmt.Sprintf("Ensure 'defer %s.Close()' is called immediately following successful acquisition error checking.", inst.VarName)

	case StateResourceLeakProbable:
		ev.FindingState = StateProbable
		ev.Severity = repo_intel.SeverityMedium
		ev.Confidence = 0.90
		if inst.DeferredInLoop {
			ev.Title = fmt.Sprintf("Deferred Resource Release Inside Loop (%s)", inst.Kind)
			ev.Description = fmt.Sprintf("Resource '%s' acquired in loop with 'defer Close()'; cleanup is delayed until function exit, accumulating descriptors.",
				inst.VarName)
			ev.RootCause = "Registering deferred cleanup inside loop body."
			ev.RecommendedAction = "Close resource explicitly at the end of each loop iteration or encapsulate loop body in a helper function."
		} else {
			ev.Title = fmt.Sprintf("Conditional Resource Leak on Exit Path (%s)", inst.Kind)
			ev.Description = fmt.Sprintf("Resource '%s' is released on some execution paths but leaks along %d of %d reachable exits.",
				inst.VarName, inst.LeakedExits, inst.ReachableExits)
			ev.RootCause = "Conditional release branch does not cover all function exit paths."
			ev.RecommendedAction = "Unify cleanup under a top-level deferred release."
		}

	case StateResourceDeferredRelease:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.99
		ev.CounterEvidence = []string{
			fmt.Sprintf("Resource '%s' of type %s has verified deferred release covering all reachable exit paths",
				inst.VarName, inst.Kind),
		}
		ev.Title = "Safe Deferred Resource Lifecycle"
		ev.Description = fmt.Sprintf("Resource '%s' is safely released via deferred call.", inst.VarName)

	case StateResourceReleased:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.98
		ev.CounterEvidence = []string{
			fmt.Sprintf("Resource '%s' of type %s is explicitly closed along all exit paths", inst.VarName, inst.Kind),
		}
		ev.Title = "Safe Explicit Resource Lifecycle"
		ev.Description = fmt.Sprintf("Resource '%s' is explicitly closed prior to return.", inst.VarName)

	case StateResourceTransferred:
		ev.FindingState = StateCounterEvidencePresent
		ev.Confidence = 0.98
		ev.CounterEvidence = []string{
			fmt.Sprintf("Resource '%s' of type %s ownership transferred to caller via return value or consumer", inst.VarName, inst.Kind),
		}
		ev.Title = "Safe Resource Ownership Transfer"
		ev.Description = fmt.Sprintf("Resource '%s' ownership was transferred, discharging local release obligation.", inst.VarName)

	default:
		ev.FindingState = StateInsufficientEvidence
		ev.Confidence = 0.60
		ev.AnalysisLimitations = []string{"Complex ownership transfer or unverified struct lifecycle"}
		ev.Title = "Unresolved Resource Lifecycle"
		ev.Description = fmt.Sprintf("Lifecycle for resource '%s' could not be statically verified.", inst.VarName)
	}

	return ev
}

func (a *ResourceAnalyzer) isAliasMatch(name string, aliases []string) bool {
	for _, al := range aliases {
		if name == al {
			return true
		}
	}
	return false
}
