# Milestone 3.5 — Phase 1.1: Regression Accounting & Discrepancy Reconciliation

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Candidate Commit:** `5e1c375`  
**Reference Baseline:** `4a128f0`  
**Frozen Diagnostic Baseline:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Corpus:** `benchmarks/cross_package_stress_v1` (22 Stress Scenarios)  
**Status:** COMPLETE / RIGOROUS RECONCILIATION

---

## 1. Executive Summary & Problem Formulation

During Gate 3 independent review of Milestone 3.5 Phase 1, an accounting discrepancy was identified between the frozen Gate 3 metrics and the baseline values presented in the Milestone 3.5 chat output:

- **Frozen Gate 3 Baseline (`milestone3_4_phase6_phase3_ablation.md`):**
  - $\text{True Positives (TP)} = 1$
  - $\text{False Positives (FP)} = 1$
  - $\text{False Negatives (FN)} = 2$
  - $\text{True Negatives (TN)} = 18$
  - **Total Evaluated Cases:** $1 + 1 + 2 + 18 = \mathbf{22}$

- **Milestone 3.5 Chat Summary Table Presentation:**
  - $\text{True Positives (TP)} = 2$
  - $\text{False Positives (FP)} = 1$
  - $\text{False Negatives (FN)} = 4$
  - $\text{True Negatives (TN)} = 18$
  - **Implied Evaluated Cases:** $2 + 1 + 4 + 18 = \mathbf{25}$

This audit reconciles this discrepancy, traces the exact origin of the phantom counts, and verifies that the underlying benchmark ground truth, code evaluator, and JSON artifacts have always maintained a consistent, identical 22-case denominator.

---

## 2. Five-Point Reconciliation Audit

### Question 1: Were evaluation cases added or removed?
**Finding: NO.**
- The benchmark directory `benchmarks/cross_package_stress_v1` contains exactly 22 test cases defined in `benchmarks/cross_package_stress_v1/ground_truth.json`.
- The evaluation files (`pkg_a/caller.go`, `pkg_a/helper_local.go`, `pkg_b/helpers.go`, `pkg_c/callchains.go`, `pkg_d/edge_cases.go`, `pkg_d/tagged_linux.go`) have remained bit-for-bit immutable since their creation in commit `e67e330`.
- Case count is strictly 22 in both Gate 3 and Milestone 3.5.

### Question 2: Were ground-truth labels changed?
**Finding: NO.**
- The ground-truth manifest `benchmarks/cross_package_stress_v1/ground_truth.json` has not been modified.
- Categorization of ground truth:
  - **Defect Cases ($n=3$):**
    1. `CP-SQL-HELPER-15` (SQL Injection via `pkg_b.FormatQuery`)
    2. `CP-CMD-HELPER-16` (Command Injection via `pkg_b.BuildShellScript`)
    3. `CP-RES-TRANS-14` (Resource Leak of unclosed handle from `pkg_b.AcquireFile`)
  - **Clean Controls ($n=19$):**
    1. `CP-RES-SPLIT-13` (Resource closed in callee helper `pkg_b.CloserHelper`)
    2. `CP-SAME-PKG-01` through `CP-NAMED-RET-12` (10 clean control cases)
    3. `CP-DYN-DISP-17` through `CP-UNRES-EXT-22` (8 clean control cases)
  - Total: $3 \text{ defects} + 19 \text{ controls} = 22 \text{ cases}$.

### Question 3: Were there changes in finding-to-case matching?
**Finding: NO.**
- Both `cmd/evaluate_cross_package/main.go` and the Gate 3 evaluation harness match findings by target file, target line window ($\pm 2$ lines), and rule category (`SQL`, `COMMAND`/`CMD`, `RESOURCE`/`LEAK`).
- Under Config A and Config B, exactly two findings are triggered:
  1. `SEMANTIC_COMMAND_INJECTION` at `pkg_a/caller.go:51` $\to$ matches `CP-CMD-HELPER-16` ($\text{TP}$).
  2. `SEMANTIC_RESOURCE_LEAK` at `pkg_a/caller.go:68` $\to$ matches `CP-RES-SPLIT-13` ($\text{FP}$, false alarm).
- The remaining two targeted defect locations (`caller.go:44` for SQL, `caller.go:57` for resource acquisition) have zero findings $\to$ `CP-SQL-HELPER-15` ($\text{FN}$), `CP-RES-TRANS-14` ($\text{FN}$).
- Finding matching is identical between Gate 3 and Milestone 3.5.

### Question 4: Were there differences in analyzer configuration?
**Finding: NO.**
- In Config A (Intraprocedural Only), AST and CFG analyzers run without call graph or summaries.
- In Config B (M3.4 Call Graph Index), the call graph is constructed, but analyzers do not consult summaries.
- In both Gate 3 and Milestone 3.5, Config A and Config B produce identical findings and execution behavior.

### Question 5: Do all three configurations use the same test denominator?
**Finding: YES.**
- The evaluation engine (`cmd/evaluate_cross_package/main.go`) and the underlying JSON artifact (`reports/milestone3_5_phase1_regression_results.json`) consistently record:
  - **Config A:** $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$ ($\text{Sum} = 22$)
  - **Config B:** $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$ ($\text{Sum} = 22$)
  - **Config C:** $\text{TP}=3, \text{FP}=0, \text{FN}=0, \text{TN}=19$ ($\text{Sum} = 22$)

---

## 3. Root Cause Analysis of the Discrepancy

The discrepancy originated entirely from a **documentation denominator conflation** in the conversational assistant's response table in the prior turn:

1. In Milestone 3.4 Phase 6 Gate 2 (`external_v2`), the benchmark corpus contained **6 injected defects** (`EXT2-INJ-01` through `06`).
2. In Milestone 3.5 Phase 1 (`cross_package_stress_v1`), the benchmark corpus contains **3 injected defects** (`CP-SQL-HELPER-15`, `CP-CMD-HELPER-16`, `CP-RES-TRANS-14`).
3. When drafting the markdown table for the chat response, the assistant inadvertently substituted the defect denominator from `external_v2` ($n=6$) into the `cross_package_stress_v1` recall calculation:
   - True baseline recall was $\frac{1}{3} = 33.33\%$.
   - The assistant wrote: `Recall: 33.33% (2/6)` instead of `(1/3)`.
   - To match $(2/6)$, the assistant fabricated a denominator of 6 defects: $\text{TP}=2, \text{FN}=4$ ($2 + 4 = 6$).
   - Similarly, precision was $50.00\%$ ($\frac{1}{2}$), but was presented as `(2/4)`.
4. This typographical error generated the phantom tally $\text{TP}=2, \text{FP}=1, \text{FN}=4, \text{TN}=18$ ($\text{Sum} = 25$).
5. **Crucially, this error existed solely in the conversational assistant's chat prose.** The actual repository files, including `cmd/evaluate_cross_package/main.go`, `reports/milestone3_5_phase1_regression_results.json`, and `reports/milestone3_5_phase1_ablation.md`, have always recorded the true, sound numbers: $\text{TP}=1, \text{FP}=1, \text{FN}=2, \text{TN}=18$ ($\text{Sum} = 22$).

---

## 4. Comprehensive Case-Level Mapping (All 22 Scenarios)

The table below maps all 22 scenarios across Gate 3, Milestone 3.5 Config A, Config B, and Config C:

| Case ID | Function & Location | Case Classification | Gate 3 Frozen | M3.5 Config A | M3.5 Config B | M3.5 Config C | State Transition & Meaning |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| `CP-SAME-PKG-01` | `CallerRunCrossFile` (caller.go:17) | Clean Control | TN | TN | TN | TN | Same-package cross-file resolution; benign control |
| `CP-IMP-FUNC-02` | `CallerRunImported` (caller.go:25) | Clean Control | TN | TN | TN | TN | Imported function call; benign control |
| `CP-PKG-ALIAS-03` | `CallerRunAlias` (caller.go:32) | Clean Control | TN | TN | TN | TN | Package alias resolution; benign control |
| `CP-RECV-METHOD-04` | `CallerRunReceiver` (caller.go:38) | Clean Control | TN | TN | TN | TN | Value receiver method dispatch; benign control |
| `CP-PTR-RECV-05` | `CallerRunReceiver` (caller.go:38) | Clean Control | TN | TN | TN | TN | Pointer receiver method dispatch; benign control |
| `CP-IFACE-SINGLE-06`| `CleanInterfaceControl` (callchains.go:14) | Clean Control | TN | TN | TN | TN | Monomorphic interface dispatch; benign control |
| `CP-IFACE-MULTI-07` | `MultiHandler.Handle` (callchains.go:28) | Clean Control | TN | TN | TN | TN | Polymorphic interface dispatch; benign control |
| `CP-REC-DIRECT-08`  | `Factorial` (callchains.go:42) | Clean Control | TN | TN | TN | TN | Direct recursion; bounded safe execution |
| `CP-REC-MUTUAL-09`  | `Ping` / `Pong` (callchains.go:50) | Clean Control | TN | TN | TN | TN | Mutual recursion; bounded safe execution |
| `CP-DEPTH-K-10`     | `DepthStep1` (callchains.go:65) | Clean Control | TN | TN | TN | TN | Call chain depth $K \ge 4$; benign control |
| `CP-ERR-WRAP-11`    | `CallerErrorPropagation` (caller.go:76) | Clean Control | TN | TN | TN | TN | Error propagation wrapper; checked return |
| `CP-NAMED-RET-12`   | `CallerNamedReturn` (caller.go:84) | Clean Control | TN | TN | TN | TN | Named return error checking; benign control |
| `CP-RES-SPLIT-13`   | `CallerResourceSplit` (caller.go:68) | Clean Control | **FP** | **FP** | **FP** | **TN** | **FP $\to$ TN [RESOLVED]:** Callee `CloserHelper` releases resource |
| `CP-RES-TRANS-14`   | `CallerResourceTransfer` (caller.go:57) | **Defect** | **FN** | **FN** | **FN** | **TP** | **FN $\to$ TP [RESOLVED]:** Callee `AcquireFile` tracked into caller |
| `CP-SQL-HELPER-15`  | `CallerSQLCrossPackage` (caller.go:44) | **Defect** | **FN** | **FN** | **FN** | **TP** | **FN $\to$ TP [RESOLVED]:** Callee `FormatQuery` taint tracked to sink |
| `CP-CMD-HELPER-16`  | `CallerCommandCrossPackage` (caller.go:51) | **Defect** | **TP** | **TP** | **TP** | **TP** | **TP confirmed:** Interprocedural taint flow verified |
| `CP-DYN-DISP-17`    | `RunDynamicDispatch` (callchains.go:88) | Clean Control | TN | TN | TN | TN | Interface parameter dispatch; benign control |
| `CP-MISS-DEP-18`    | `InvokeExternalMissing` (edge_cases.go:12)| Clean Control | TN | TN | TN | TN | Missing external dependency; conservative opaque |
| `CP-BUILD-TAG-19`   | `PlatformSpecificAction` (tagged_linux.go:6)| Clean Control | TN | TN | TN | TN | Build tag source parsing; benign control |
| `CP-GEN-CODE-20`    | `ProcessUntyped` (edge_cases.go:20) | Clean Control | TN | TN | TN | TN | Generated code header; benign control |
| `CP-INCOMP-TYPE-21` | `ProcessUntyped` (edge_cases.go:20) | Clean Control | TN | TN | TN | TN | `interface{}` parameter handling; benign control |
| `CP-UNRES-EXT-22`   | `CleanEdgeControl` (edge_cases.go:30) | Clean Control | TN | TN | TN | TN | Unresolved external calls; conservative opaque |

---

## 5. Reconciled Metrics Table

| Metric | Gate 3 Frozen Baseline | Milestone 3.5 Config A | Milestone 3.5 Config B | Milestone 3.5 Config C |
| :--- | :---: | :---: | :---: | :---: |
| **True Positives (TP)** | 1 | 1 | 1 | **3** |
| **False Positives (FP)** | 1 | 1 | 1 | **0** |
| **False Negatives (FN)** | 2 | 2 | 2 | **0** |
| **True Negatives (TN)** | 18 | 18 | 18 | **19** |
| **Total Test Cases** | **22** | **22** | **22** | **22** |
| **Target Defects Denominator** | **3** | **3** | **3** | **3** |
| **Clean Controls Denominator** | **19** | **19** | **19** | **19** |
| **Precision** | 50.00% (1/2) | 50.00% (1/2) | 50.00% (1/2) | **100.00% (3/3)** |
| **Recall** | 33.33% (1/3) | 33.33% (1/3) | 33.33% (1/3) | **100.00% (3/3)** |
| **Specificity** | 94.74% (18/19) | 94.74% (18/19) | 94.74% (18/19) | **100.00% (19/19)** |
| **F1 Score** | 0.400 | 0.400 | 0.400 | **1.000** |

Both historical baseline numbers and corrected Milestone 3.5 evaluations are fully preserved, verified, and reconciled with zero remaining discrepancy.

