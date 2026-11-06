# MILESTONE 3.4 — PHASE 5 KNOWN LIMITATIONS & RESEARCH HORIZONS
## Unrecovered Defect Categories, Static Analysis Boundaries & Future Roadmap

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Baseline Commit:** `be917a4`  
**Evaluation Date:** November 2026  
**Classification:** Scientific Limitations & Architectural Disclosure  

---

## 1. Executive Summary

Milestone 3.4 achieved its core objective: eliminating AST pattern rigidity for SQL injection, command injection, resource lifecycle leaks, and unhandled errors. Across targeted categories, recall reached 100% on synthetic fixtures and 100% on external/held-out targeted subsets with 0 false positives.

However, scientific integrity requires transparent disclosure of the remaining architectural boundaries. The engine is not an omniscient verifier and must not be deployed under that presumption.

---

## 2. Unrecovered Historical Defect Categories (External v1)

On Milestone 3.3's full 14-defect external denominator, full-suite recall is **64.29% (9/14)**. The 5 unrecovered defects belong to categories deliberately outside the scope of Phases 2–4:

| Defect ID | Repository | Target Category | Defect Description | Root Architectural Cause | Proposed Future Milestone |
|---|---|---|---|---|---|
| `DEF-CACHE-01` | `cache_engine` | `CONCURRENCY_RACE` | Concurrent read to map without RLock | Requires happens-before memory model and thread-safety analysis | Milestone 3.5 (Concurrency Analysis) |
| `DEF-CACHE-02` | `cache_engine` | `GOROUTINE_LEAK` | Unbounded goroutine spawn without termination | Requires channel communication analysis and lifetime join tracking | Milestone 3.5 (Concurrency Analysis) |
| `DEF-HTTP-03` | `http_router` | `CONTEXT_LEAK` | `context.WithTimeout` cancel function ignored | Cancel func is a closure rather than a struct method or error return | Milestone 3.5 (Context & Timer Lifecycles) |
| `DEF-CLI-02` | `cli_tooling` | `PATH_TRAVERSAL` | `filepath.Join` with untrusted directory traversal | Requires path sanitization and boundary check taint propagation | Milestone 3.6 (Taint & Path Traversal) |
| `DEF-QUEUE-02` | `queue_worker` | `TIMER_LEAK` | `time.After` in select loop accumulating channels | Requires loop-scoped timer channel lifecycle abstraction | Milestone 3.5 (Context & Timer Lifecycles) |

---

## 3. Interprocedural Call Resolution Boundaries

1. **Bounded Traversal Depth ($K=3$):** Call graphs are traversed to a maximum call chain depth of 3 hops. While sufficient for modular helpers and wrapper functions, deeply nested framework layers (e.g., ORM reflection internals) fall back to `CallUnresolved`.
2. **Third-Party Dependency Opacity:** Imported packages without accessible source ASTs (e.g., binary vendor packages or external CGo modules) are treated as `CallExternalOpaque`. The engine conservatively refuses to assume external methods release handles or check errors.
3. **Dynamic Reflection & Method Dispatch:** Invocations using `reflect.Value.Call`, `unsafe.Pointer`, or unresolvable dynamic function pointers cannot be statically traced.

---

## 4. Remediation Scope & Polyglot Boundaries

1. **Autonomous Remediation Language Restriction:** Autonomous code modification remains strictly restricted to Go. C++ and Rust codebases remain read-only research targets under the Milestone 3.1 capability qualification.
2. **Contextual Suppressions Are Not Universally Invariant:** While suppression of `defer tx.Rollback()` when followed by `Commit()` represents standard Go idiom, unconventional transaction managers where `Rollback` performs mandatory state reset could be masked if not explicitly verified.
3. **Sample Size Limitations:** While the held-out evaluation verified 100% precision and recall across 21 test units (9 defects, 12 controls), statistical generalization to heterogeneous, multi-million LOC codebases requires broader real-world empirical validation.

