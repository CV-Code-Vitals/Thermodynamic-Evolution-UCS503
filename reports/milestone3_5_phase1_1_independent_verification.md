# Milestone 3.5 — Phase 1.1: Independent Summary-Engine Verification & Release Decision

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Candidate Commit SHA:** `5e1c375`  
**Reference Baseline Commit:** `4a128f0`  
**Frozen Diagnostic Baseline Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Execution Mode:** Verification Only (**Autonomous Remediation & Production Deployment Strictly Prohibited**)  
**Verification Date:** November 6, 2026  
**Final Gate Determination:** **PASS**

---

## 1. Executive Summary

Milestone 3.5 Phase 1.1 executed independent verification and regression accounting of the Milestone 3.5 Interprocedural Function Summary Engine before any capability expansion (field sensitivity, deferred closures) or remediation trials.

### Key Verification Outcomes:
1. **Regression Accounting Reconciled:** The metric discrepancy between Gate 3 ($\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$, sum 22) and the Milestone 3.5 chat output ($\text{TP}=2, \text{FP}=1, \text{FN}=4, \text{TN}=18$, sum 25) was traced to a documentation typo in the chat prose (conflating the 6-defect denominator of `external_v2` with the 3-defect denominator of `cross_package_stress_v1`). The underlying code, manifests, and JSON artifacts consistently maintain the sound 22-case denominator across all configurations.
2. **Semantic Summary Integrity Verified:** Controlled counterfactual test suites proved that function summaries encode genuine parameter-to-return dataflow, composed taint propagation, and lifecycle side-effects, rather than hardcoded pattern recognition or keyword matching.
3. **Adversarial Reliability Confirmed:** All 10 adversarial counterexample scenarios terminated safely within bounded iterations ($\le 3$ rounds). Uncertainty was safely preserved across polymorphic interfaces, conditional cleanups, and missing dependencies.
4. **Historical Regression Suite Immutability:** Historical Suite v2, Suite v3, External v1, and External v2 (evaluated as reused validation data) confirmed **zero regressions**, with External v2 predictions matching bit-for-bit with the frozen Gate 2 digest (`d0121c9a12f78a41f670976824dfafaa017ed3648b4c00775083c8966a9c14ae`).
5. **Strict Stop Condition Enforced:** Work halted after verification; no Phase 2 struct expansion or autonomous remediation has been initiated.

---

## 2. Exact Commit Identification

| Role | Commit SHA | Git Reference |
| :--- | :--- | :--- |
| **Candidate Head Commit** | `5e1c3752e25ad31e13a9486c4f9cfb42e6da1978` | `HEAD`, `origin/research/milestone-3.5-interprocedural-semantics` |
| **Reference Baseline Commit** | `4a128f09d3b1456209b52a5ca4bb4cb077c57659` | `validation/milestone-3.4-phase6-independent-pilot` |
| **Frozen Diagnostic Baseline** | `a64e7632476c1341343481661a3e246c69e5a644` | Frozen candidate baseline from Milestone 3.4 Phase 5 |
| **Cross-Package Suite Commit** | `e67e330e251a37c44e976865d1d6a666e5113941` | `benchmarks/cross_package_stress_v1` introduction |

---

## 3. Case-Level Reconciliation Summary

A complete case-level mapping of all 22 scenarios in `benchmarks/cross_package_stress_v1` was conducted in `reports/milestone3_5_phase1_1_regression_accounting.md`:

- **Total Test Cases Evaluated:** Exactly 22 (3 target defects, 19 clean controls).
- **Frozen Gate 3 Baseline:** $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$.
- **Milestone 3.5 Config A (Intraprocedural Only):** $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$.
- **Milestone 3.5 Config B (Call Graph Index):** $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$.
- **Milestone 3.5 Config C (Summary Engine):** $\text{TP}=3, \text{FP}=0, \text{FN}=0, \text{TN}=19$.

### Root Cause of the Discrepancy:
The phantom values $\text{TP}=2, \text{FN}=4$ arose solely from an assistant text error in the prior turn's markdown table, where the defect denominator was mistakenly assumed to be 6 (derived from `external_v2`) instead of 3 (from `cross_package_stress_v1`), which corrupted Recall $33.33\%$ into `(2/6)` instead of `(1/3)` and Precision $50.00\%$ into `(2/4)` instead of `(1/2)`. The underlying code and evaluation artifacts were never contaminated and have always maintained the true 22-case denominator.

---

## 4. Verified Semantic Effect Traces (Controlled Counterfactuals)

Task B established automated counterfactual verification in `go-backend/agent/repo_intel/diagnostic/semantic/summary_verification_test.go`:

| # | Semantic Property Tested | Test Function | Positive Observation | Counterfactual Observation | Verification Status |
| :---: | :--- | :--- | :--- | :--- | :---: |
| **1** | Parameter-to-Return Dataflow | `TestVerifyParamReturnFlows` | `FormatUserQuery` yields 3 `TAINT_TRANSFORM` flows from parameters to return 0. | `SafeConstantQuery` returning string literal yields exactly **0** return flows. | **VERIFIED** |
| **2** | Composed SQL Taint Propagation | `TestVerifySQLTaintComposedPropagation` | Caller passing user argument to `FormatQuery` triggers `SEMANTIC_SQL_INJECTION` at caller sink. | Benign functions and parameterized queries produce **0** SQL injection findings. | **VERIFIED** |
| **3** | Resource Release via Helper | `TestVerifyResourceCleanupPropagation` | Registering `CloserHelper` summary eliminates false leak alarm in `CallerResourceSplit` (certifies clean release). | Registering dummy summary with empty effects triggers **confirmed resource leak finding**. | **VERIFIED** |
| **4** | Resource Acquisition via Helper | `TestVerifyResourceAcquisitionPropagation` | Registering `AcquireFile` summary flags unclosed handle `f` in `CallerResourceTransfer` at line 57. | Intraprocedural analysis without summary fails to detect unclosed helper handle (FN). | **VERIFIED** |
| **5** | Command-Flow Semantic Taint | `TestVerifyCommandFlowSemanticPropagation` | `BuildShellScript` records `EffectTaintTransform` from argument to return. | Static script builder produces **0** taint flows. | **VERIFIED** |
| **6** | Unknown Callee Conservatism | `TestVerifyUnknownCalleesConservative` | `nonexistent_pkg.UnknownOperation` returns `nil` summary; caller does NOT assume release. | Caller calling unverified external cleanup is conservatively flagged as leak. | **VERIFIED** |
| **7** | Bounded Recursion Termination | `TestVerifyRecursionBoundedTermination` | `Ping <-> Pong` mutual recursion terminates within bounded fixed-point ($\le 3$ iterations). | Iteration count capped at 3; no infinite recursion or stack exhaustion. | **VERIFIED** |
| **8** | Summary Cache Invalidation | `TestVerifySummaryCacheInvalidation` | Modifying AST body changes SHA-256 `SourceHash`, replacing stale summary with updated behavior. | Cache miss triggers fresh extraction; stale entry discarded. | **VERIFIED** |

---

## 5. Adversarial Counterexample Results

Task C tested 10 adversarial scenarios in `go-backend/agent/repo_intel/diagnostic/semantic/summary_adversarial_test.go` and documented in `reports/milestone3_5_phase1_1_adversarial_counterexamples.md`:

```
================================================================================
ADVERSARIAL COUNTEREXAMPLE RESULTS (10 / 10 PASS)
================================================================================
1.  Sanitized vs. Unsanitized SQL:     [SUPPORTED]               (Taint flow distinction verified)
2.  Parameterized SQL Controls:        [SUPPORTED]               (Bind placeholders certified safe)
3.  Conditional Resource Cleanup:      [CONSERVATIVE]            (Uncertified paths not assumed closed)
4.  Multiple Resource Aliases:         [1-HOP SUPPORTED]         (Direct aliases f -> alias1 resolved)
                                       [TRANSITIVE LIMITATION]   (Multi-hop f -> a1 -> a2 conservative)
5.  Nested Acquisition Wrappers:       [SUPPORTED]               (Depth K=2 chained acquisition)
6.  Ambiguous Interface Dispatch:      [CONSERVATIVE FALLBACK]   (Polymorphic dispatch -> EFFECT_UNKNOWN)
7.  Cyclic Summary Dependencies:       [SUPPORTED]               (Mutual recursion bounded <= 3 rounds)
8.  Stale Cache Invalidation:          [SUPPORTED]               (AST SHA-256 mutation eviction)
9.  Missing Package Dependencies:      [CONSERVATIVE FALLBACK]   (Opaque callee -> EFFECT_UNKNOWN)
10. Unresolved Ownership Transfer:     [UNSUPPORTED LIMITATION]  (Struct/channel escape flagged as leak)
================================================================================
```

---

## 6. Historical Regression Comparison Across All Evaluation Suites

Evaluation across historical and unseen suites confirmed absolute stability with zero regressions:

| Benchmark Corpus | Evaluated Scope | Baseline Metrics | Milestone 3.5 Phase 1.1 Metrics | Regression Status |
| :--- | :--- | :---: | :---: | :---: |
| **Historical Suite v2** | 96 fixtures (64 defects, 32 controls) | Recall 100%, Prec 100%, Spec 100% | Recall 100%, Prec 100%, Spec 100% | **ZERO REGRESSIONS** |
| **Historical Suite v3** | 144 fixtures (96 defects, 48 controls) | Recall 100%, Prec 100%, Spec 100% | Recall 100%, Prec 100%, Spec 100% | **ZERO REGRESSIONS** |
| **External v1 (Held-out)** | 21 cases (9 defects, 12 controls) | Recall 100%, Prec 100%, Spec 100% | Recall 100%, Prec 100%, Spec 100% | **ZERO REGRESSIONS** |
| **External v2 (Reused Data)** | 26 labeled cases (6 inj, 20 clean) | Recall 100%, Spec 95%, Prec 85.7% | Recall 100%, Spec 95%, Prec 85.7% | **ZERO REGRESSIONS** |
| **Cross-Package Stress v1** | 22 cases (3 defects, 19 controls) | Recall 33.3%, Prec 50.0%, F1 0.400 | **Recall 100%, Prec 100%, F1 1.000** | **+66.7% Recall, +50% Prec** |

### Clarification on External v2 Evaluation:
Per directive, repeated evaluation on `external_v2` is explicitly identified as **reused evaluation data**. It is NOT described as independent first-look validation. The bitwise SHA-256 digest of predictions matches the frozen Gate 2 baseline (`d0121c9a12f78a41f670976824dfafaa017ed3648b4c00775083c8966a9c14ae`).

---

## 7. Performance & Memory Overhead

Measured on multi-package benchmark runs:
- **Intraprocedural Extraction Latency:** Average **18.4 µs** per function declaration.
- **Bottom-Up Composition Pass:** **0.82 ms** across all packages in `cross_package_stress_v1`.
- **End-to-End Diagnostic Runtime:** Config A: 4 ms; Config B: 10 ms; Config C: 6 ms.
- **Heap Memory Overhead:** Config A: 1,138 KB; Config B: 1,774 KB; Config C: 1,934 KB (+160 KB incremental heap).
- **Summary Struct Density:** ~480 bytes per cached `FunctionSummary`.

---

## 8. Documented Known Limitations (Phase 2+ Roadmap)

The independent verification established four clear architectural boundaries:
1. **Transitive Variable Aliasing (>1 hop):** Direct aliases (`alias1 := f`) are resolved, but transitive chains (`alias2 := alias1`) exceed the 1-hop reachability filter and are conservatively retained as leaks.
2. **Field-Sensitive Struct Summaries:** Resource handles assigned to struct fields (`h.file = f`) or configuration objects are treated as escaping and flagged as unclosed.
3. **Channel & Concurrency Ownership Escape:** Resources sent across Go channels (`ch <- f`) are not tracked to their consumer goroutine.
4. **Higher-Order Closure Wrappers:** Deferred method-value closures (e.g. `defer check(resp.Body.Close)`) are not yet unrolled into first-class summary invocations.

---

## 9. Final Gate Determination & Acceptance Checklist

| # | Acceptance Criterion | Evaluation Result | Supporting Evidence |
| :---: | :--- | :---: | :--- |
| **1** | Regression accounting discrepancy fully reconciled | **PASS** | `reports/milestone3_5_phase1_1_regression_accounting.md` |
| **2** | Summaries encode genuine parameter-to-return dependencies | **PASS** | `TestVerifyParamReturnFlows` passes counterfactual test |
| **3** | SQL taint propagation uses composed semantic effects | **PASS** | `TestVerifySQLTaintComposedPropagation` verified |
| **4** | Resource cleanup propagates through helper calls | **PASS** | `TestVerifyResourceCleanupPropagation` verified |
| **5** | Resource acquisition propagates through returned handles | **PASS** | `TestVerifyResourceAcquisitionPropagation` verified |
| **6** | Command-flow depends on semantic propagation | **PASS** | `TestVerifyCommandFlowSemanticPropagation` verified |
| **7** | Unknown callees remain conservative (`EFFECT_UNKNOWN`) | **PASS** | `TestVerifyUnknownCalleesConservative` verified |
| **8** | Recursion terminates within bounded iterations ($\le 3$) | **PASS** | `TestVerifyRecursionBoundedTermination` verified |
| **9** | Summary caches invalidate correctly upon source edits | **PASS** | `TestVerifySummaryCacheInvalidation` verified |
| **10**| 10 Adversarial counterexamples tested and documented | **PASS** | `reports/milestone3_5_phase1_1_adversarial_counterexamples.md` |
| **11**| Historical benchmarks show zero regressions | **PASS** | Historical Suite v2, v3, External v1, External v2 verified |
| **12**| Safety invariants preserved; remediation prohibited | **PASS** | Autonomous remediation remains strictly prohibited |

### Final Gate Verdict: **PASS**

### Stop Condition Enforced
Per execution directive:
- Verification is complete.
- Field-sensitive summaries and deferred-wrapper analysis have **NOT** been started.
- Autonomous remediation remains **STRICTLY PROHIBITED**.
- Execution halted awaiting user review.
