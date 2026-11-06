# Milestone 3.5 — Phase 1.1: Adversarial Counterexamples & Boundary Assessment

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Candidate Commit:** `5e1c375`  
**Reference Baseline:** `4a128f0`  
**Evaluation File:** `go-backend/agent/repo_intel/diagnostic/semantic/summary_adversarial_test.go`  
**Status:** COMPLETE / RIGOROUS EMPIRICAL AUDIT

---

## 1. Executive Summary

Milestone 3.5 Phase 1.1 subjected the new Interprocedural Function Summary Engine (`function_summary.go`) to ten adversarial counterexample scenarios designed to stress dataflow composition, resource lifecycle tracking, cache invalidation, and conservative boundary isolation.

The evaluation rigorously verified:
1. **Safety Invariant Preservation:** In no scenario did the engine invent false exact summaries or unsafely drop unhandled defects.
2. **Conservative Uncertainty:** Incomplete type information, missing dependencies, and ambiguous polymorphic dispatches correctly degrade to `EFFECT_UNKNOWN` / conservative handling.
3. **Explicit Boundary Mapping:** Concrete limitations (transitive multi-hop aliasing and struct field escapes) were empirically confirmed and cataloged as candidates for Phase 2.

---

## 2. Adversarial Scenarios Evaluation Matrix

| # | Adversarial Scenario | Evaluation Test Case | Engine Response & Mechanism | Classification | Handling Verdict |
| :---: | :--- | :--- | :--- | :---: | :---: |
| **1** | Sanitized vs. Unsanitized SQL Inputs | `TestAdversarial_SQL_SanitizedVsUnsanitized` | Unsanitized `fmt.Sprintf` flags `EffectTaintTransform`; parameterized static template yields 0 taint flows. | Semantic Flow | **SUPPORTED** |
| **2** | Parameterized SQL Controls | `TestAdversarial_SQL_ParameterizedControls` | Query string containing `?` with discrete argument certified as `COUNTER_EVIDENCE_PRESENT`. | Semantic Flow | **SUPPORTED** |
| **3** | Conditional Resource Cleanup | `TestAdversarial_Resource_ConditionalCleanup` | Helper `if shouldClose { f.Close() }` cannot guarantee execution on all caller exit paths; caller treats as uncertified. | Path Safety | **CONSERVATIVE** |
| **4A**| Direct Resource Aliasing (1-hop) | `TestAdversarial_Resource_MultipleAliases` (4A) | `alias1 := f; CloserHelper(alias1)` tracked via `discoverAliases`; caller recognizes cleanup without leak. | Alias Set | **SUPPORTED** |
| **4B**| Transitive Resource Aliasing (>1 hop) | `TestAdversarial_Resource_MultipleAliases` (4B) | `alias2 := alias1; CloserHelper(alias2)` exceeds 1-hop AST alias depth; conservatively flagged as leak. | Alias Limit | **UNSUPPORTED (Known Limit)** |
| **5** | Resource Return via Nested Wrappers | `TestAdversarial_Resource_NestedWrappers` | Depth $K=2$ chain (`WrapOpen -> OpenRaw -> os.Open`) propagates `EffectResourceAcquire` through return tuple. | Bottom-Up Composition | **SUPPORTED** |
| **6** | Ambiguous Interface Dispatch | `TestAdversarial_Interface_AmbiguousDispatch` | Monomorphic interface resolves exact; polymorphic interface falls back to conservative selector without guessing. | Dispatch Safety | **CONSERVATIVE FALLBACK** |
| **7** | Recursive Summary Dependencies | `TestAdversarial_Recursion_MutualCyclic` | Cyclic mutual dependency (`Ping <-> Pong`) terminates within fixed-point bound ($\le 3$ iterations). | SCC Fixed-Point | **SUPPORTED** |
| **8** | Stale Cache Invalidation | `TestAdversarial_Cache_StaleInvalidation` | Modifying function AST body alters SHA-256 `SourceHash`, discarding stale summary and recomputing fresh state. | Cache Integrity | **SUPPORTED** |
| **9** | Missing Dependencies | `TestAdversarial_MissingDependencies` | Unresolved external package or missing source returns `nil` summary; downstream analyzers assume `EFFECT_UNKNOWN`. | Fault Isolation | **CONSERVATIVE FALLBACK** |
| **10**| Unresolved Ownership Transfer | `TestAdversarial_Resource_UnresolvedOwnershipTransfer` | Resource stored in struct field `h.file = f` or sent to channel `ch <- f` is flagged as unclosed exit path. | Field Sensitivity | **UNSUPPORTED (Known Limit)** |

---

## 3. Detailed Case Analyses

### Scenario 1: Sanitized vs. Unsanitized SQL Inputs
- **Implementation:** Analyzed `UnsanitizedQuery` (interpolating `user` parameter into format string) versus `ParameterizedQuery` (returning static query string `"SELECT * FROM users WHERE name = $1"`).
- **Behavior:** `analyzeReturnDataflows` detected parameter interpolation via `fmt.Sprintf` and assigned `EffectTaintTransform` with confidence 0.95. For `ParameterizedQuery`, the return expression is a constant `*ast.BasicLit`, resulting in 0 parameter return flows.
- **Verdict:** **SUPPORTED**. Prevents false positive taint propagation when functions return parameterized query templates.

### Scenario 2: Parameterized SQL Controls
- **Implementation:** Analyzed a database query call `db.QueryContext(ctx, query, id)` where `query` contains bind placeholder `?` and `id` is passed as an execution parameter.
- **Behavior:** `SQLAnalyzer.analyzeCallExpressionSQL` identified bind placeholder in query string and matched trailing argument `id`. It registered evidence with `FindingState: COUNTER_EVIDENCE_PRESENT`, preventing a false positive `SEMANTIC_SQL_INJECTION` alarm.
- **Verdict:** **SUPPORTED**.

### Scenario 3: Conditional Resource Cleanup
- **Implementation:** Function `ConditionalCloser(f *os.File, shouldClose bool)` executes `f.Close()` inside `if shouldClose { ... }`.
- **Behavior:** Intraprocedural extraction notes the existence of `f.Close()`, but because the exit path reachability cannot guarantee that `shouldClose` is true at caller call sites, caller analysis conservatively treats the resource as not unconditionally closed on all paths.
- **Verdict:** **CONSERVATIVE**. Avoids unsound leak suppression.

### Scenario 4: Resource Aliasing (Direct vs. Transitive)
- **Direct 1-Hop Aliasing (`alias1 := f`):**
  - Tracked by `ResourceAnalyzer.discoverAliases`. When caller invokes `pkg_b.CloserHelper(alias1)`, the alias set `{f, alias1}` matches the parameter passed to the callee summary, certifying release (**SUPPORTED**).
- **Transitive Multi-Hop Aliasing (`alias1 := f; alias2 := alias1`):**
  - In `resource_analysis.go:174-177`, `discoverAliases` compares `rhsId.Name == instances[i].VarName`. Because `rhsId` is `alias1` (not `f`), `alias2` is not added to the alias set.
  - When caller invokes `pkg_b.CloserHelper(alias2)`, the analyzer cannot resolve `alias2` back to `f`, reporting a conservative leak alarm.
  - **Verdict:** **UNSUPPORTED (KNOWN LIMITATION)**. Preserves conservative safety; scheduled for multi-hop transitive closure expansion in Phase 2.

### Scenario 5: Nested Resource Acquisition Wrappers ($K \ge 2$)
- **Implementation:** `OpenRaw` acquires `os.File` directly; `WrapOpen` calls `OpenRaw` and returns its handle.
- **Behavior:** Bottom-up composition pass analyzes `OpenRaw` first, recording `EffectResourceAcquire` on return index 0. When analyzing `WrapOpen`, the call to `OpenRaw` is resolved via the summary registry, propagating the acquisition to `WrapOpen`'s return tuple.
- **Verdict:** **SUPPORTED**. Enables multi-layered factory functions.

### Scenario 6: Ambiguous Interface Dispatch
- **Implementation:** Lookup on polymorphic interface method `Processor.Process` with multiple possible concrete implementations.
- **Behavior:** The summary engine does not fabricate an exact match across multiple candidate types. `GetSummary` returns a conservative fallback with reduced confidence (<0.8), ensuring downstream analyzers treat polymorphic effects as `EFFECT_UNKNOWN`.
- **Verdict:** **CONSERVATIVE FALLBACK**. Prevents unsound assumption of callee behavior across dynamic boundaries.

### Scenario 7: Cyclic Mutual Recursion
- **Implementation:** Mutually recursive functions `Ping(n, s) <-> Pong(n-1, s)`.
- **Behavior:** The SCC fixed-point iteration in `BuildSummariesBottomUp` limits convergence cycles to at most 3 rounds (`iteration_count <= 3`). The analyzer completed in <1 ms without stack overflow or infinite looping.
- **Verdict:** **SUPPORTED**. Guarantees bounded termination on cyclic call graphs.

### Scenario 8: Stale Cache Invalidation via Source Hash
- **Implementation:** Evaluated cache replacement when function AST body is modified from a tainted return (`fmt.Sprintf(...)`) to a static literal (`"v2: constant"`).
- **Behavior:** `ExtractIntraproceduralSummary` hashes the printed AST body (`SourceHash = hex(sha256(body))`). The hash changed from `v1` to `v2`, successfully evicting the old summary and updating the cache with the benign summary.
- **Verdict:** **SUPPORTED**. Deterministic cache soundness under incremental edits.

### Scenario 9: Missing External Dependencies
- **Implementation:** Caller invokes functions from an external package without registered source models (`missing_third_party_pkg.ExternalCgoFunc`).
- **Behavior:** `GetSummary` returns `nil`. Downstream SQL, Command, and Resource analyzers handle `nil` summaries by defaulting to `EFFECT_UNKNOWN`. No crash, no panic, no assumed sanitization or release.
- **Verdict:** **CONSERVATIVE FALLBACK**.

### Scenario 10: Unresolved Ownership Transfer (Field & Channel Escapes)
- **Implementation:** Resource handle `f` is assigned to a struct field `h.file = f` or sent to a channel `ch <- f`.
- **Behavior:** The current intraprocedural CFG exit-path solver does not model struct field lifecycles or concurrent channel consumption. Consequently, `f` is treated as unreleased on function exit and flagged as a leak finding.
- **Verdict:** **UNSUPPORTED (KNOWN LIMITATION)**. Struct field sensitivity and channel ownership transfer are explicitly identified as the core subject of Milestone 3.5 Phase 2.

---

## 4. Summary of Supported vs. Unsupported Capabilities

```
================================================================================
MILESTONE 3.5 PHASE 1 SEMANTIC CAPABILITY BOUNDARIES
================================================================================
[+] FULLY SUPPORTED:
    - Intraprocedural parameter-to-return taint flows (TaintTransform, TaintPassthrough)
    - SQL helper formatting and sink taint propagation
    - Command argument formatting and shell sink verification
    - Helper-based resource acquisition (factory functions)
    - Helper-based explicit resource cleanup (closer functions)
    - Direct 1-hop variable aliasing (alias := f; Closer(alias))
    - Chained wrapper acquisitions (Depth K >= 2)
    - Bounded recursive fixed-point convergence (<= 3 rounds)
    - AST-level SHA-256 source hash cache invalidation

[~] CONSERVATIVE FALLBACK (SAFE ABSTENTION):
    - Conditional resource cleanup (if cond { f.Close() })
    - Ambiguous polymorphic interface dispatch
    - Missing third-party / CGO package dependencies (EFFECT_UNKNOWN)

[-] UNSUPPORTED / KNOWN LIMITATIONS (PHASE 2+ ROADMAP):
    - Multi-hop transitive aliasing (f -> alias1 -> alias2 -> Closer)
    - Field-sensitive struct summaries (h.file = f)
    - Concurrency & channel ownership escape (ch <- f)
    - Higher-order closure wrappers (defer check(f.Close))
================================================================================
```
