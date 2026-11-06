# Milestone 3.4 — Phase 6 / Phase 3: Scientific Assessment

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Frozen Detector Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Status:** COMPLETE / EMPIRICAL HYPOTHESIS TESTING

---

## 1. Executive Summary

Milestone 3.4 Phase 6 Phase 3 investigated whether the frozen cross-package semantic integration engine provides genuine interprocedural reasoning capabilities beyond intraprocedural pattern detection.

Four formal scientific claims were evaluated against empirical results from `benchmarks/cross_package_stress_v1`, the controlled ablation study, and the adversarial stress matrix:

| Claim | Proposition | Scientific Determination | Key Supporting Evidence & Counterexamples |
| :--- | :--- | :---: | :--- |
| **Claim A** | The frozen engine resolves supported cross-package calls correctly. | **PARTIALLY_SUPPORTED** | Exact resolution verified on direct imports and same-package files (0.00% false-exact rate); however, cross-package receiver dispatch and interface polymorphism fall back to conservative approximations. |
| **Claim B** | The frozen engine performs useful cross-package semantic propagation. | **NOT_SUPPORTED** | SQL taint, command arguments, and resource ownership do **not** propagate across function call boundaries. Taint generated in helpers is invisible to caller sinks; cleanup in helpers triggers false positive leak alarms. |
| **Claim C** | Cross-package integration measurably improves detection. | **NOT_SUPPORTED** | Controlled ablation proved $\Delta \text{Defects} = 0$. Both configurations produced identical findings. Integration added +400% runtime overhead with zero diagnostic recall improvement. |
| **Claim D** | The engine safely handles ambiguous and unsupported cases. | **SUPPORTED** | 100% pass on 13 adversarial stress scenarios. Zero stack overflows on recursion, zero fabricated exact edges, zero crashes on malformed code, and robust remediation gating. |

---

## 2. In-Depth Claim Evaluations

### 2.1 Claim A: Cross-Package Call Resolution
> *"The frozen engine resolves supported cross-package calls correctly."*

- **Verdict:** **PARTIALLY_SUPPORTED**
- **Supporting Evidence:**
  - `CallSiteResolver` correctly resolved 100% of direct same-package cross-file calls (`pkg_a/caller.go` $\to$ `pkg_a/helper_local.go:LocalFormat`) as `RESOLVED_EXACT`.
  - Package import paths and rename aliases (`b_alias`) were accurately tracked via AST import specs and resolved to `pkg_b` (`RESOLVED_EXACT`).
  - Third-party/standard library imports were safely mapped to `EXTERNAL_OPAQUE`.
  - **False-Exact Resolution Rate was 0.00%** (zero non-exact targets were falsely labeled exact).
- **Counterexamples & Limitations:**
  - Value and pointer receiver methods across package boundaries (`pkg_b.Service.Process`) could not be resolved exactly because receiver type estimation (`estimateReceiverType`) is restricted to the local file/package scope, falling back to `RESOLVED_CONSERVATIVE`.
  - Dynamic interface dispatch and multi-target polymorphism are not modeled.

---

### 2.2 Claim B: Cross-Package Semantic Propagation
> *"The frozen engine performs useful cross-package semantic propagation."*

- **Verdict:** **NOT_SUPPORTED**
- **Supporting Evidence:**
  - Signature-level error return checking is shared across package functions via `RegisterLocalFunction`.
- **Counterexamples:**
  - **SQL Analysis:** `CallerSQLCrossPackage` invokes `pkg_b.FormatQuery(..., user)`. The SQL analyzer does not inline or summarize `FormatQuery`; it treats `query` as an opaque return value, producing a **False Negative**.
  - **Resource Lifecycle (Caller-to-Callee):** `CallerResourceSplit` acquires `os.Open` and passes the handle to `pkg_b.CloserHelper(f)`. The analyzer ignores the callee cleanup and flags a **False Positive** leak on line 68.
  - **Resource Lifecycle (Callee-to-Caller):** `CallerResourceTransfer` calls `pkg_b.AcquireFile(path)` and discards the handle without closing it. The analyzer fails to detect the leak (**False Negative**) because acquisition is not recognized across the boundary.

---

### 2.3 Claim C: Diagnostic Improvement over Intraprocedural Analysis
> *"Cross-package integration measurably improves detection over intraprocedural-only analysis."*

- **Verdict:** **NOT_SUPPORTED**
- **Supporting Evidence:**
  - None.
- **Counterexamples:**
  - The controlled ablation study compared Configuration A (Intraprocedural Only) with Configuration B (Full Integration).
  - Both configurations produced **identical findings** (2 findings in `cross_package_stress_v1`):
    - `SEMANTIC_COMMAND_INJECTION` at `caller.go:51`
    - `SEMANTIC_RESOURCE_LEAK` at `caller.go:68`
  - Delta in defect detection: **$\Delta = 0$**.
  - Delta in false positive reduction: **$\Delta = 0$**.
  - Delta in diagnostic runtime: **$+400\%$** (from 2 ms to 10 ms).
  - Delta in memory allocation: **$+60.4\%$** (from 1,095 KB to 1,757 KB).

---

### 2.4 Claim D: Adversarial and Ambiguity Safety
> *"The engine safely handles ambiguous and unsupported cases."*

- **Verdict:** **SUPPORTED**
- **Supporting Evidence:**
  - **Recursion & Cycle Safety:** Direct self-recursion (`Factorial`) and mutual recursion (`Ping`/`Pong`) executed safely without infinite loops or stack overflows because AST call mapping is non-recursive.
  - **Uncertainty Preservation:** When facing ambiguous interface implementations or missing modules, the engine strictly defaulted to `RESOLVED_CONSERVATIVE` or `EXTERNAL_OPAQUE`, maintaining a 0.00% false-exact rate.
  - **Failure Isolation:** Parser failures, missing modules, and syntax errors were safely isolated; the panic recovery wrapper in `unified_orchestrator.go` ensured sibling analyzers continued execution.
  - **Remediation Safety:** Zero findings in uncertain or stress code were marked `AUTONOMOUS_ELIGIBLE`.

---

## 3. Scientific Synthesis and Architectural Takeaway

The empirical evidence resolves the core research questions of Milestone 3.4 Phase 6:

1. **The Nature of Milestone 3.4 Progress:** Milestone 3.4 succeeded in upgrading the diagnostic engine from fragile regexes to **rigorous, conservative intraprocedural CFG and AST dataflow analysis**.
2. **The Generalization Illusion:** Claims that Milestone 3.4 achieved "cross-package semantic generalization" are refuted by the ablation study. The engine can *scan* multi-package repositories and *map* call edges, but its *vulnerability reasoning* remains strictly bounded within individual function scopes.
3. **Future Milestone Roadmap (Milestone 3.5):** True interprocedural analysis will require whole-program call-graph summaries, interprocedural taint propagation, and interprocedural resource escape/ownership tracking.

