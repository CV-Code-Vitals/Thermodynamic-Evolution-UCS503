# Milestone 3.4 — Phase 6 / Phase 3: Stress Corpus Provenance Report

**Corpus:** `benchmarks/cross_package_stress_v1`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Frozen Detector:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Status:** COMPLETE / PROVENANCE CERTIFIED

---

## 1. Corpus Objective & Stratification

The `cross_package_stress_v1` benchmark was constructed to provide an independent, stress-testing environment specifically targeting cross-file, cross-package, and interprocedural semantics under frozen detector constraints.

The corpus enforces a strict two-strata separation:
- **Stratum A (External Historical Cases):** Independent real-world historical defects.
- **Stratum B (Adversarial Constructed Cases):** Controlled semantic fixtures exercising 22 precise language and architectural constructs.

---

## 2. Stratum A — External Historical Cases Audit

### 2.1 Research & Forensics
To evaluate whether authentic external historical defects could serve as cross-package benchmarks, all historical defect candidates across Milestone 3.3 and Milestone 3.4 were re-audited:
- `external_v1` (Milestone 3.3): 14 defects evaluated. All 14 defects were strictly intraprocedural (scoped to single function bodies).
- `external_v2` (Milestone 3.4 Gate 1): 4 proposed natural defect candidates were audited. All 4 were disqualified (2 were false defect claims; 2 were accepted Go standard library print idioms).
- Stratum D exploratory candidates (`gorilla_websocket/conn.go:447`, `schollz_progressbar/examples/download/main.go:14`): Verified in Gate 3.0 as conservative static warnings or closure-wrapping false alarms, not upstream maintainer-confirmed bugs.

### 2.2 Documented Limitation
**Result:** **ZERO independent human-confirmed natural external cross-package defects are available in the frozen corpus.**  
In accordance with Gate 1 scientific integrity rules:
> *"Do not promote agent-generated source interpretation into independent human verification."*
The absence of qualified historical cross-package defects is explicitly documented as a foundational empirical limitation.

---

## 3. Stratum B — Adversarial Constructed Cases (n=22)

To evaluate the mechanical reasoning boundaries of the engine, 22 controlled adversarial scenarios were authored in `benchmarks/cross_package_stress_v1/`:

| Case ID | Construct Tested | Source File | Callee Target | Expected Resolution State | Expected Propagation |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `CP-SAME-PKG-01` | Same-package cross-file | `pkg_a/caller.go` | `pkg_a/helper_local.go:LocalFormat` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-IMP-FUNC-02` | Imported package function | `pkg_a/caller.go` | `pkg_b/helpers.go:FormatQuery` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-PKG-ALIAS-03` | Package import alias | `pkg_a/caller.go` | `b_alias.BuildShellScript` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-RECV-METHOD-04` | Value receiver method | `pkg_a/caller.go` | `pkg_b.Service.Process` | `RESOLVED_CONSERVATIVE` | `UNSUPPORTED` |
| `CP-PTR-RECV-05` | Pointer receiver method | `pkg_a/caller.go` | `pkg_b.(*Service).Mutate` | `RESOLVED_CONSERVATIVE` | `UNSUPPORTED` |
| `CP-IFACE-SINGLE-06` | Single-target interface | `pkg_c/callchains.go` | `FastProcessor.ProcessData` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-IFACE-MULTI-07` | Multi-target interface | `pkg_c/callchains.go` | `MultiHandler.Handle` | `RESOLVED_CONSERVATIVE` | `UNSUPPORTED` |
| `CP-REC-DIRECT-08` | Direct self-recursion | `pkg_c/callchains.go` | `Factorial` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-REC-MUTUAL-09` | Mutual recursion | `pkg_c/callchains.go` | `Ping / Pong` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-DEPTH-K-10` | Call depth K=2, 3, 4 | `pkg_c/callchains.go` | `DepthStep1 -> 2 -> 3 -> 4` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-ERR-WRAP-11` | Error wrapper | `pkg_a/caller.go` | `pkg_b.WrapError` | `RESOLVED_EXACT` | `SUPPORTED_SIGNATURE_ONLY` |
| `CP-NAMED-RET-12` | Named return across packages | `pkg_a/caller.go` | `pkg_b.NamedReturnCalc` | `RESOLVED_EXACT` | `SUPPORTED_SIGNATURE_ONLY` |
| `CP-RES-SPLIT-13` | Acquired locally, closed in callee | `pkg_a/caller.go` | `pkg_b.CloserHelper` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-RES-TRANS-14` | Acquired in callee, transferred | `pkg_a/caller.go` | `pkg_b.AcquireFile` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-SQL-HELPER-15` | SQL helper construction | `pkg_a/caller.go` | `pkg_b.FormatQuery` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-CMD-HELPER-16` | Shell command helper | `pkg_a/caller.go` | `pkg_b.BuildShellScript` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-DYN-DISP-17` | Dynamic interface dispatch | `pkg_c/callchains.go` | `p.ProcessData` | `RESOLVED_CONSERVATIVE` | `UNSUPPORTED` |
| `CP-MISS-DEP-18` | Missing external module | `pkg_d/edge_cases.go` | `fmt.Println` | `EXTERNAL_OPAQUE` | `UNSUPPORTED` |
| `CP-BUILD-TAG-19` | Build-tag-dependent file | `pkg_d/tagged_linux.go`| `PlatformSpecificAction` | `RESOLVED_EXACT` | `UNSUPPORTED` |
| `CP-GEN-CODE-20` | Generated code file marker | `pkg_d/edge_cases.go` | `fmt.Sprintf` | `EXTERNAL_OPAQUE` | `UNSUPPORTED` |
| `CP-INCOMP-TYPE-21` | Incomplete type (interface{}) | `pkg_d/edge_cases.go` | `fmt.Sprintf` | `EXTERNAL_OPAQUE` | `UNSUPPORTED` |
| `CP-UNRES-EXT-22` | Unresolved local invocation | `pkg_d/edge_cases.go` | `ProcessUntyped` | `RESOLVED_EXACT` | `UNSUPPORTED` |

---

## 4. Epistemic Limitation & Non-Contamination

1. **Constructed Nature:** All 22 cases were constructed specifically for stress testing during Milestone 3.4 Phase 6 Phase 3.
2. **Real-World Generalization Boundary:** These cases evaluate whether the frozen detector's algorithms satisfy mechanical properties (termination on recursion, alias mapping, receiver estimation). **They do not establish real-world external generalization.**
3. **No Contamination:** None of these cases or package directories were used during Milestone 3.4 Phase 1–5 development or Gate 2 first-look evaluation.
