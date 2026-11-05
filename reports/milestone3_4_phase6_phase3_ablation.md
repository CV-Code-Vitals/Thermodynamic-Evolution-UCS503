# Milestone 3.4 — Phase 6 / Phase 3: Controlled Ablation Study

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Detector Frozen Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Evaluation Corpus:** `benchmarks/cross_package_stress_v1` (22 Stress Cases)  
**Status:** COMPLETE / RIGOROUS COMPARATIVE EVALUATION

---

## 1. Executive Summary

A controlled ablation study was executed to evaluate the incremental diagnostic contribution of the Milestone 3.4 cross-package integration layer (`BoundedCallGraph`, `CallSiteResolver`, package models) compared to isolated intraprocedural analysis.

The comparative evaluation was conducted without altering frozen detector logic by comparing:
- **Configuration A (Intraprocedural Only):** Analyzers executed on isolated function bodies without cross-package model indexing, call graph construction, or cross-file function registration.
- **Configuration B (Frozen Cross-Package Integration):** Full `UnifiedOrchestrator` execution with package discovery, symbol table caching, call-edge resolution, and finding reconciliation.

---

## 2. Quantitative Ablation Matrix

| Metric | Configuration A (Intraprocedural Only) | Configuration B (Integrated Call Graph) | Incremental Delta ($\Delta = B - A$) | Interpretation |
| :--- | :---: | :---: | :---: | :--- |
| **True Positives (TP)** | 1 (`CP-CMD-HELPER-16`) | 1 (`CP-CMD-HELPER-16`) | **0** | No change in defect detection |
| **False Positives (FP)** | 1 (`CP-RES-SPLIT-13`) | 1 (`CP-RES-SPLIT-13`) | **0** | False leak persists identically |
| **False Negatives (FN)** | 2 (`CP-SQL-HELPER-15`, `CP-RES-TRANS-14`) | 2 (`CP-SQL-HELPER-15`, `CP-RES-TRANS-14`) | **0** | Interprocedural defects missed identically |
| **True Negatives (TN)** | 18 | 18 | **0** | Clean controls pass identically |
| **Abstentions** | 0 | 0 | **0** | No abstention mechanism |
| **Unsupported Cases** | 20 | 20 | **0** | Interprocedural taint/lifecycle unsupported |
| **Call Edges Resolved** | 0 (Call graph disabled) | 41 | **+41** | Call graph construction functional |
| **False Exact Resolutions**| 0 | 0 (0.00%) | **0** | High resolution safety |
| **Diagnostic Runtime** | **2 ms** | **10 ms** | **+8 ms (+400%)** | Overhead of package indexing & AST call graph |
| **Allocated Memory** | **1,095 KB** | **1,757 KB** | **+662 KB (+60.4%)** | Package models and edge structures in heap |
| **Evidence Completeness** | 100% intraprocedural AST trace | 100% intraprocedural AST trace | **0.0%** | Defect evidence retains intraprocedural scope |

---

## 3. Findings Comparison

Both configurations produced **identically identical defect sets**:

### Finding 1: `SEMANTIC_COMMAND_INJECTION`
- **Location:** `benchmarks/cross_package_stress_v1/pkg_a/caller.go:51`
- **Config A:** Triggered by intraprocedural sink check (`exec.CommandContext(..., "sh", "-c", shellCmd)`).
- **Config B:** Triggered identically.
- **Delta:** 0.

### Finding 2: `SEMANTIC_RESOURCE_LEAK`
- **Location:** `benchmarks/cross_package_stress_v1/pkg_a/caller.go:68`
- **Config A:** Triggered by intraprocedural CFG exit reachability (missing local `Close()`).
- **Config B:** Triggered identically (caller does not consult callee `CloserHelper` summary).
- **Delta:** 0.

---

## 4. Incremental Contribution Analysis

1. **Defect Detection Impact:** **ZERO INCREMENTAL DEFECT RETRIEVAL ($\Delta = 0$)**.  
   The call graph and package resolution layer did not uncover any additional defects beyond the intraprocedural baseline.
2. **Precision & False Alarm Impact:** **ZERO REDUCTION IN FALSE POSITIVES ($\Delta = 0$)**.  
   The call graph did not suppress the false positive in `CallerResourceSplit`, because callee summaries are not connected to the resource exit reachability solver.
3. **Architectural Overhead:**  
   The cross-package infrastructure incurs an additional 8 ms (400% latency increase) and 662 KB (60.4% memory increase) to construct symbol tables and call edges that are not consumed by the vulnerability detection rules.

---

## 5. Scientific Conclusion

The ablation study decisively proves:
> **In Milestone 3.4, the diagnostic reasoning capabilities of the engine are 100% intraprocedural.**

The cross-package subsystem functions as an **architectural skeleton and call-graph generator**, but does not participate in semantic interprocedural defect detection or false-alarm suppression.
