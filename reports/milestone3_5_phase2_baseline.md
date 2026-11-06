# Milestone 3.5 — Phase 2: Baseline Freeze Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Verified Candidate Baseline Commit:** `8c38ef926210bd007ddb3dc97f2e77aa4cc4c9c8`  
**Frozen Diagnostic Baseline Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Authorization:** GO — BOUNDED SEMANTIC ENGINEERING  
**Autonomous Remediation:** PROHIBITED  
**Production Deployment:** PROHIBITED  
**Status:** COMPLETE / BASELINE FROZEN

---

## 1. Executive Summary

Prior to initiating code changes for Milestone 3.5 Phase 2 (Alias Resolution, Higher-Order Closures & Field-Sensitive Resource Semantics), this baseline freeze audit locks the system state, records benchmark metrics, preserves Phase 1.1 counterfactual evidence, and empirically reproduces all four targeted limitation categories.

### Pre-Implementation State:
- **Git Commit:** `8c38ef926210bd007ddb3dc97f2e77aa4cc4c9c8` (Head: `5e7f0b6460d0576f07a72911ee38afe41ce1112f`)
- **Working Tree:** 100% clean, all pre-existing tests passing.
- **Remediation Engine:** Inactive (Autonomous remediation strictly prohibited).

---

## 2. Frozen Benchmark Results

The table below records the verified diagnostic baseline across all frozen evaluation suites prior to Phase 2:

| Evaluation Suite | Evaluated Scope | Precision | Recall | Specificity | F1 Score | Status |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: |
| **Historical Suite v2** | 96 fixtures (64 defects, 32 controls) | 100.00% | 100.00% | 100.00% | 1.000 | Immutable Verified |
| **Historical Suite v3** | 144 fixtures (96 defects, 48 controls) | 100.00% | 100.00% | 100.00% | 1.000 | Immutable Verified |
| **External v1 (Held-out)** | 21 cases (9 defects, 12 controls) | 100.00% | 100.00% | 100.00% | 1.000 | Immutable Verified |
| **External v2 (Reused)** | 26 cases (6 injected, 20 controls) | 85.71% | 100.00% | 95.00% | 0.923 | Reused Evaluation Data |
| **Cross-Package Stress v1** | 22 cases (3 defects, 19 controls) | 100.00% | 100.00% | 100.00% | 1.000 | Gate 3 Resolved Baseline |

---

## 3. Reproduction of the Four Targeted Limitations

Empirical test runs in `summary_adversarial_test.go` and diagnostic analysis confirm the exact failure boundaries of the Phase 1.1 baseline:

### 1. Transitive Multi-Hop Aliasing Limitation
- **Symptom:** In `resource_analysis.go:174-177`, `discoverAliases` matches `assign.Rhs[0].Name == instances[i].VarName` (1-hop check).
- **Reproduction:** Given `f, _ := os.Open(path); alias1 := f; alias2 := alias1; CloserHelper(alias2)`, `alias1` is recognized as an alias of `f`, but `alias2` is NOT added to the alias set.
- **Diagnostic Consequence:** `CloserHelper(alias2)` is not recognized as releasing `f`, producing a false positive resource leak warning (`RESOURCE_LEAK_CONFIRMED`).

### 2. Higher-Order Deferred Closures & Method Values Limitation
- **Symptom:** When a resource cleanup method value is passed into a deferred wrapper function (e.g. `defer check(resp.Body.Close)` or `defer cleanup(f.Close)`), the syntactic AST visitor in `evaluateDeferredReleases` only checks for direct calls of the shape `defer <obj>.Close()`.
- **Reproduction:** A wrapper `func check(fn func() error) { _ = fn() }` called via `defer check(resp.Body.Close)` is not unrolled as a release contract.
- **Diagnostic Consequence:** `resp.Body` is falsely flagged as leaked at function exit.

### 3. Field-Sensitive Struct Resource Tracking Limitation
- **Symptom:** Resource handles stored in local struct fields (e.g. `h := &Holder{file: f}` or `cfg.SQL = query`) are not tracked with access-path sensitivity.
- **Reproduction:** Given `StoreInStruct(h *Holder, path string)` assigning `h.file = f`, the analyzer treats `f` as escaping local scope without knowing if `h` owns or releases it.
- **Diagnostic Consequence:** Flagged as an unclosed leak finding or unresolvable escape.

### 4. Concurrent Ownership Boundary Limitation
- **Symptom:** Handing off resource handles via channels (`ch <- f`) or spawning background goroutines (`go worker(f)`) escapes function scope without lifetime tracking.
- **Reproduction:** Handing off an open socket or file across a channel to a background worker.
- **Diagnostic Consequence:** Currently flagged as unclosed local leak or unhandled escape.

---

## 4. Baseline Performance and Memory Footprint

Measured on `benchmarks/cross_package_stress_v1` multi-package diagnostic execution:
- **Intraprocedural Extraction Latency:** 18.4 µs / function.
- **Bottom-Up Composition Pass:** 0.82 ms.
- **Total Pipeline Runtime:** 6.0 ms (Config C).
- **Heap Allocation:** 1,934 KB (160 KB incremental heap above intraprocedural baseline).
- **Summary Struct Density:** ~480 bytes per cached `FunctionSummary`.

---

## 5. Active Safety Invariants

- **Remediation Executor:** INACTIVE. (Autonomous remediation is strictly prohibited).
- **Benchmark Trees:** Immutable bitwise freeze enforced.
- **Gate Status:** Ready for Phase 1 (Bounded Multi-Hop Alias Analysis).

