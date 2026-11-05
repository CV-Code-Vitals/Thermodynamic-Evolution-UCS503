# Milestone 3.4 — Phase 6 / Phase 3: Cross-Package Capability Inventory

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Frozen Detector Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Status:** COMPLETE / RIGOROUS SCIENTIFIC CODE AUDIT

---

## 1. Executive Summary

A comprehensive architectural audit of the frozen semantic diagnostic engine was conducted across `callgraph.go`, `semantic.go`, `cfg.go`, `unified_orchestrator.go`, and the four domain analyzers (`error_analysis.go`, `resource_analysis.go`, `sql_analysis.go`, and `command_analysis.go`).

The audit establishes a critical scientific distinction:
1. **Call Graph Modeling:** The engine successfully models package structures, extracts symbol tables, builds bounded call graphs, and resolves call edges across files and packages with conservative uncertainty classifications.
2. **Signature Resolution:** The error analyzer shares a cross-function registry (`RegisterLocalFunction`) to resolve whether a local function returns an `error`.
3. **Dataflow Propagation Limitation:** **The SQL, Command, and Resource analyzers operate strictly intraprocedurally**. While `CallSiteResolver` records call edges, callee summaries and interprocedural taint/lifecycle propagation are **not** piped into the intraprocedural reaching-definitions engine.

---

## 2. Four-Tier Capability Classification Model

To avoid conflating AST/call-edge resolution with true interprocedural dataflow, each construct is classified across four rigorous tiers:

- **`CALL_EDGE_RESOLVED`:** The engine matches a call site AST node to a target function name, file, line, and package.
- **`CALLEE_SUMMARY_AVAILABLE`:** The engine computes a structured behavioral summary of the callee (e.g. side-effects, resource consumption, error generation, taint transformation).
- **`SEMANTIC_PROPAGATION_VERIFIED`:** Facts or constraints discovered in the callee or caller flow across the call boundary and update the caller's or callee's dataflow state.
- **`FINDING_ATTRIBUTION_VERIFIED`:** A diagnostic finding is generated whose root cause, evidence chain, or classification depends directly on cross-boundary propagated state.

---

## 3. Detailed Construct-by-Construct Audit Matrix

| Language Construct / Scenario | Symbol Resolution | Call Edge Resolved | Callee Summary Available | Semantic Propagation | Finding Attribution | Implementation Status & Exact Code Path |
| :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **Same-Package Local Call** (`foo()`) | YES | **`CALL_EDGE_RESOLVED`** | PARTIAL (`IsErrorPropagator`, `IsResourceConsumer`) | **NO** (except Error signature) | **NO** | `callgraph.go:183-200` |
| **Same-Package Cross-File Call** | YES | **`CALL_EDGE_RESOLVED`** | PARTIAL | **NO** (except Error signature) | **NO** | `callgraph.go:202-218` |
| **Imported Project Package Call** (`pkg.Func()`) | YES | **`CALL_EDGE_RESOLVED`** | PARTIAL | **NO** | **NO** | `callgraph.go:234-252` |
| **Imported External Stdlib Call** (`os.Open()`) | YES | **`CALL_EDGE_RESOLVED`** (`EXTERNAL_OPAQUE`) | HEURISTIC (regex/name checks) | **NO** | **NO** | `callgraph.go:254-263` |
| **Package Aliases** (`import p "path/to/pkg"`) | YES | **`CALL_EDGE_RESOLVED`** | PARTIAL | **NO** | **NO** | `callgraph.go:120-127` |
| **Method on Local Variable** (`obj.Method()`) | YES | **`CALL_EDGE_RESOLVED`** | PARTIAL | **NO** | **NO** | `callgraph.go:265-296`, `estimateReceiverType` |
| **Chained Method Call** (`a.B().C()`) | NO | `CallResolvedConservative` | NONE | **NO** | **NO** | `callgraph.go:298-304` |
| **Interface Dispatch (Single Target)** | NO | `CallResolvedConservative` | NONE | **NO** | **NO** | Not modeled in dynamic dispatch; treated as conservative selector |
| **Interface Dispatch (Multiple Targets)**| NO | `CallResolvedConservative` | NONE | **NO** | **NO** | Enum `CallMultipleTargets` exists in types but is never emitted |
| **Recursive Call Cycles** | YES | `CALL_EDGE_RESOLVED` | NONE | **NO** | **NO** | Traversal is non-recursive (single-pass AST walk per function) |
| **Call Depth Bounding (K=2,3,4)** | N/A | N/A | N/A | N/A | N/A | No interprocedural inlining or multi-hop chaining is executed |
| **Error Wrapper Helper** | YES | `CALL_EDGE_RESOLVED` | Signature only (`IsErrorReturning`) | **NO** | **NO** | Intraprocedural error checker only inspects local call statement |
| **Resource Acquisition Helper** | YES | `CALL_EDGE_RESOLVED` | Flag `IsResourceConsumer` only | **NO** | **NO** | `ResourceAnalyzer` only flags acquisitions in local function body |
| **Resource Release Helper** (`close(f)`) | YES | `CALL_EDGE_RESOLVED` | Flag `IsResourceConsumer` only | **NO** | **NO** | Caller still flags resource as leaked if closed in callee helper |
| **SQL Construction Helper** | YES | `CALL_EDGE_RESOLVED` | NONE | **NO** | **NO** | `SQLAnalyzer` only traces `df.ResolveRootExpression` locally |
| **Command Construction Helper** | YES | `CALL_EDGE_RESOLVED` | NONE | **NO** | **NO** | `CommandAnalyzer` only traces within local reaching definitions |
| **Higher-Order Function / Method Value** | NO | `CallUnresolved` | NONE | **NO** | **NO** | Pass by value (e.g. `defer check(resp.Body.Close)`) unhandled |

---

## 4. Subsystem Audits

### 4.1 `CallSiteResolver` (`callgraph.go`)
- **Implemented Capabilities:**
  - Indexing registered `PackageSemanticModel`s by `PackagePath` and `PackageName`.
  - Mapping AST imports, including explicit rename aliases (`import mypkg "github.com/foo/bar"`).
  - Resolving identifiers to local CFG functions (`CallResolvedExact`).
  - Resolving selector expressions against imported packages (`CallResolvedExact` for registered project packages, `CallExternalOpaque` for third-party).
  - Estimating receiver types for variable selectors from function parameter lists and local composite literals (`&MyType{}`).
- **Limitations:**
  - Does not construct an interprocedural whole-program dataflow graph.
  - The resolution state `CallMultipleTargets` is defined in `CallResolutionState` enum but never instantiated by `ResolveCall`.
  - Execution budgets (e.g. timeout, maximum hop depth) are structural constants rather than dynamically metered graph traversal steps.

### 4.2 `ErrorAnalyzer` (`error_analysis.go`)
- **Implemented Capabilities:**
  - Registers local function declarations across the package via `RegisterLocalFunction(fnDecl)`.
  - Uses `a.localFuncs` in `ResolveCallSignature` to determine whether a locally defined helper returns an `error`.
- **Limitations:**
  - Does not propagate error values through wrappers across packages.
  - Does not track named return error aliases across multiple functions.

### 4.3 `ResourceAnalyzer` (`resource_analysis.go`)
- **Implemented Capabilities:**
  - CFG block construction, intraprocedural reaching-definitions, loop-scoped leak detection, fallthrough and explicit exit path reachability.
  - Local escape analysis (detecting if the resource variable is returned in `ast.ReturnStmt`).
- **Limitations:**
  - **No interprocedural ownership transfer:** If function `A` acquires `f := os.Open(...)` and passes `f` to function `B(f)` which closes it, function `A` flags `f` as leaked unless `f` is returned by `A`.
  - If a cleanup function is deferred via a helper (e.g. `defer check(resp.Body.Close)`), the intraprocedural analyzer flags it as unclosed.

### 4.4 `SQLAnalyzer` & `CommandAnalyzer` (`sql_analysis.go`, `command_analysis.go`)
- **Implemented Capabilities:**
  - Intraprocedural reaching-definitions traversal (`df.ResolveRootExpression`) tracing variables back to `fmt.Sprintf`, `+`, `strings.Join`, or parameter inputs.
- **Limitations:**
  - **Zero interprocedural taint propagation:** If function `Sanitize(x)` or `BuildQuery(x)` is called, the analyzer cannot trace taint into or out of the helper function.

---

## 5. Architectural Verification Conclusion

The capability inventory confirms:
1. `callgraph.go` provides an **architectural foundation for call graph representation and resolution state classification**.
2. However, **the diagnostic findings in Milestone 3.4 are generated by intraprocedural CFG and reaching-definitions engines**.
3. Therefore, claims of "cross-package defect detection" in Milestone 3.4 reflect **the ability to analyze repositories containing multiple packages**, not whole-program interprocedural dataflow propagation.
