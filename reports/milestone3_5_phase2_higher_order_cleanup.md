# Milestone 3.5 — Phase 2: Higher-Order Cleanup & Deferred Method Values

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Implementation File:** `go-backend/agent/repo_intel/diagnostic/semantic/higher_order_cleanup.go`  
**Test Suite:** `go-backend/agent/repo_intel/diagnostic/semantic/higher_order_cleanup_test.go`  
**Status:** COMPLETE / VERIFIED

---

## 1. Executive Summary

Milestone 3.5 Phase 2 implements conservative higher-order analysis for deferred cleanup wrappers and method values. The engine addresses the mandatory pattern:

```go
defer check(resp.Body.Close)
```

The analyzer strictly distinguishes **callback capture** from **callback invocation**. Simply passing a cleanup function or method value into another function is not treated as proof of release. The engine inspects the wrapper function's AST or interprocedural summary to verify whether the callback is guaranteed to be invoked on relevant execution paths.

---

## 2. Mandatory Pattern Analysis: `defer check(resp.Body.Close)`

### 2.1 Go Semantics of Deferred Method Values
In Go, `defer check(resp.Body.Close)` has two distinct execution phases:
1. **Defer Registration (Synchronous):** The receiver expression `resp.Body` and method value `Close` are evaluated and captured when the `defer` statement is reached.
2. **Function Return (Deferred):** The wrapper function `check` is called with the captured method value as its argument.

### 2.2 Static Analysis Algorithm
1. `HigherOrderCleanupAnalyzer.AnalyzeDeferredCall` detects that an argument to `check(...)` is an `*ast.SelectorExpr` whose method name matches `Close` for `ResourceKindHTTPBody`.
2. The receiver `resp.Body` is matched against the active resource instance or its parent variable (`resp`).
3. The wrapper function `check` is looked up in the local package function declarations (`RegisterLocalFunc`) or summary engine.
4. `analyzeWrapperFuncDecl` inspects `check`'s body:
   - If `fn()` is called directly and unconditionally: classified as `INVOKE_MUST_ONCE` $\to$ **Verified Release**.
   - If `fn()` is enclosed in `if cond { fn() }`: classified as `INVOKE_CONDITIONAL` $\to$ **Conservative (Not Certified)**.
   - If `fn()` is not invoked: classified as `INVOKE_NEVER` $\to$ **Confirmed Leak**.
   - If `fn()` is called via `go fn()`: classified as `INVOKE_ASYNC` $\to$ **Conservative (Not Synchronous)**.
   - If `fn` is assigned to a variable or slice: classified as `INVOKE_STORED` $\to$ **Conservative (Not Immediate)**.

---

## 3. Callback Invocation Classification Taxonomy

| Invocation Classification | Code Pattern Inside Wrapper | Static Certainty | Certified as Release? |
| :--- | :--- | :---: | :---: |
| `INVOKE_MUST_ONCE` | `func check(fn func() error) { _ = fn() }` | 100% Guaranteed | **YES (Verified)** |
| `INVOKE_CONDITIONAL` | `func check(fn func() error) { if flag { fn() } }` | Path-Dependent | **NO (Conservative)** |
| `INVOKE_NEVER` | `func check(fn func() error) { /* no call */ }` | 0% Invocation | **NO (Confirmed Leak)** |
| `INVOKE_ASYNC` | `func check(fn func() error) { go fn() }` | Concurrent/Race | **NO (Non-Synchronous)** |
| `INVOKE_STORED` | `func check(fn func() error) { global = fn }` | Stored for later | **NO (Unresolved)** |
| `INVOKE_UNKNOWN` | External unindexed wrapper function | Opaque | **NO (Conservative Fallback)** |

---

## 4. Empirical Test Suite Results

The implementation was verified against six automated test cases in `higher_order_cleanup_test.go`:

| Test Name | Evaluated Pattern | Expected Classification | Observed Result | Status |
| :--- | :--- | :--- | :--- | :---: |
| `TestHigherOrder_DeferredMethodValue_MandatoryExample` | `defer check(resp.Body.Close)` | `INVOKE_MUST_ONCE` | Certified `RESOURCE_DEFERRED_RELEASE`; 0 leaks | **PASS** |
| `TestHigherOrder_DeferredAnonymousFunc` | `defer func() { _ = f.Close() }()` | Direct unrolling | Certified `RESOURCE_DEFERRED_RELEASE`; 0 leaks | **PASS** |
| `TestHigherOrder_WrapperNeverInvokes_Counterfactual` | `func noopWrapper(fn func() error) {}` | `INVOKE_NEVER` | Confirmed `RESOURCE_LEAK_CONFIRMED`; finding generated | **PASS** |
| `TestHigherOrder_WrapperStoresCallback` | `savedCallback = fn` | `INVOKE_STORED` | Confirmed leak; release not assumed | **PASS** |
| `TestHigherOrder_WrapperAsyncGoroutine` | `go fn()` | `INVOKE_ASYNC` | Confirmed leak; async execution not certified | **PASS** |
| `TestHigherOrder_WrapperConditional` | `if shouldCloseGlobal { fn() }` | `INVOKE_CONDITIONAL` | Confirmed leak; conditional execution not certified | **PASS** |

---

## 5. Controlled Counterfactual Verification

A critical scientific requirement of Milestone 3.5 Phase 2 is the counterfactual test:
- **Baseline Case (`TestHigherOrder_DeferredMethodValue_MandatoryExample`):** Wrapper `check` unconditionally calls `fn()`. The engine recognizes the method value and discharges the resource lifecycle obligation.
- **Counterfactual Case (`TestHigherOrder_WrapperNeverInvokes_Counterfactual`):** The code is identical except the wrapper body does not call `fn()`. The engine immediately detects that invocation is absent (`INVOKE_NEVER`) and issues a confirmed `RESOURCE_LEAK_CONFIRMED` diagnostic finding.

This conclusively proves that the analyzer reasons about **actual wrapper semantics and invocation contracts**, rather than syntactically accepting any call containing `.Close`.

