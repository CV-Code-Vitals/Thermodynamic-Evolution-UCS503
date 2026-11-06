# Milestone 3.5 — Phase 1: Known Architectural & Semantic Limitations

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Classification:** Scientific Boundary Definition & Static Analysis Roadmap  
**Status:** COMPLETE / SPECIFICATION LOCKED

---

## 1. Executive Summary

Milestone 3.5 Phase 1 established a working interprocedural function summary engine that successfully resolves parameter taint propagation, callee resource acquisition, and callee resource release across package boundaries.

To preserve scientific rigor, this report explicitly documents the known boundaries of Phase 1 and outlines the architectural requirements for future phases.

---

## 2. Summary of Implemented vs. Unaddressed Capabilities

| Semantic Construct | Milestone 3.4 Baseline | Milestone 3.5 Phase 1 | Status & Limitation |
| :--- | :---: | :---: | :--- |
| **Cross-Package SQL Helper Taint** | Unsupported (FN) | **Supported (TP)** | Tracks string formatting & concatenation to return tuple. |
| **Callee Resource Release** | Unsupported (FP) | **Supported (TN)** | Caller recognizes helper `.Close()` / `.Stop()`. |
| **Callee Resource Acquisition** | Unsupported (FN) | **Supported (TP)** | Helper factory calls registered in caller CFG. |
| **Package Import Aliasing** | AST Ident Only | **Fully Resolved** | Maps import renames to callee summaries. |
| **Recursive Call Cycles (SCCs)** | Acyclic Walk | **Bounded Fixed-Point** | Terminates in $\le 3$ iterations without divergence. |
| **Higher-Order Closures / Method Values** | Unsupported | **Unsupported** | Passing `resp.Body.Close` to `check(f func())` unmodeled. |
| **Field-Sensitive Struct Taint** | Unsupported | **Conservative** | Taint on struct fields is joined at struct level. |
| **Channel / Goroutine Dataflow** | Unsupported | **Unsupported** | Taint flow across `ch <- val` is not tracked. |
| **Path-Sensitive Condition Narrowing** | Unsupported | **Conservative** | Path conditions (`if err != nil`) merged conservatively. |

---

## 3. Deep Dive into Unaddressed Limitations

### 3.1 Higher-Order Functions & Method-Value Closures
- **Problem Statement:** In Go idioms such as:
  ```go
  defer check(resp.Body.Close)
  ```
  `resp.Body.Close` is evaluated as a method value of type `func() error` passed to `check`.
- **Phase 1 Behavior:** The function summary engine operates on direct function and method invocations (`*ast.CallExpr` with selector or identifier). It does not perform lambda lifting, defunctionalization, or escape analysis for first-class function pointers.
- **Consequence:** Scenarios like `schollz_progressbar/examples/download/main.go:14` still require manual review or future higher-order function summary extensions.

### 3.2 Field-Sensitive Struct Taint Tracking
- **Problem Statement:** When a helper populates a field in a complex struct:
  ```go
  cfg := HelperBuild(tainted) // cfg.SQL = tainted, cfg.Timeout = 10
  ```
  If caller reads `cfg.Timeout`, field-insensitive models can treat all fields as tainted.
- **Phase 1 Behavior:** Taint flows are currently indexed by parameter index and return tuple index. Field-level paths (`ReturnIndex 0.Field "SQL"`) are not yet decomposed into fine-grained access paths.
- **Mitigation:** The current engine focuses on direct string returns and resource pointers, avoiding false alarms on unrelated fields.

### 3.3 Dynamic Interface Polymorphism with Disjoint Behaviors
- **Problem Statement:** If interface `I` has two concrete implementations:
  - `ImplA.Execute(x)` sanitizes `x`.
  - `ImplB.Execute(x)` concatenates `x` into raw SQL.
- **Phase 1 Behavior:** When the receiver type is purely an interface without dynamic type flow narrowing, the engine applies conservative union ($\sqcap$), assuming the expression remains tainted.
- **Soundness Guarantee:** This guarantees zero false negatives, but may produce conservative warnings where runtime type analysis could have proven safety.

### 3.4 Concurrency & Channel Synchronization
- **Problem Statement:** Dataflow transmitted across goroutine channels (`go worker(ch)` followed by `ch <- untrusted`) is not tracked across the channel boundary.
- **Phase 1 Behavior:** Channels are treated as opaque synchronization primitives. Concurrency data race detection remains out of scope for Milestone 3.5.

---

## 4. Architectural Roadmap for Milestone 3.5 Phase 2+

1. **Phase 2:** Field-sensitive struct summaries (`StructPath` access paths).
2. **Phase 3:** Method-value closure analysis for deferred cleanup wrappers (`defer wrapper(f.Close)`).
3. **Phase 4:** Interface point-to set refinement via local variable type assertions (`v, ok := i.(*Concrete)`).
