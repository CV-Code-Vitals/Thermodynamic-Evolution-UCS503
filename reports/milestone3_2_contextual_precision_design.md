# MILESTONE 3.2 — PHASE 2: CONTEXTUAL PRECISION ARCHITECTURE & DESIGN
## Multi-Tier False-Positive Suppression, Lifecycle Finding States & Traceable Suppression Ledger

**Project:** Autonomous Self-Healing Repository Intelligence Pipeline  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Target:** Eliminate False Positives (FPR $\le 10\%$, Specificity $\ge 90\%$) while maintaining Defect Recall $\ge 95\%$.  
**Status:** ARCHITECTURAL SPECIFICATION APPROVED  
**Date:** 2026-11-03  

---

## 1. Architectural Motivation & Design Principles

Milestone 3.1 independent verification proved that naive pattern matching and coarse directory-level heuristics generate an unacceptable False Positive Rate (31.25% on clean controls). To prepare the autonomous self-healing engine for real production deployments, the diagnostic pipeline must transition from static heuristic alerts to **Evidence-Based Contextual Diagnostics**.

### Core Principles:
1. **Evidence Symmetry:** A finding cannot be promoted to `CONFIRMED` without unambiguous positive evidence (e.g., AST node traces, taint paths, unhandled error returns) AND the verified absence of countervailing idioms (e.g. guard checks, parameterization, or encapsulation).
2. **Finding Lifecycle States:** Replace binary `Defect / Not Defect` with explicit probabilistic lifecycle states:
   - `CONFIRMED`: Defect verified by AST structure, absence of counter-evidence, and actionable impact.
   - `PROBABLE`: Strong heuristic match, but requiring corroborating traces or external tests.
   - `INSUFFICIENT_EVIDENCE`: Potential smell, but lacked necessary context (e.g. API contracts without local tests); **must not trigger automated remediation**.
   - `COUNTER_EVIDENCE_PRESENT`: Candidate defect actively disproven by AST guards or safe wrappers; **suppressed**.
   - `HUMAN_INVESTIGATION_REQUIRED`: Complex cross-module interaction requiring human review.
3. **Traceable Suppression Ledger:** Every suppressed finding must be recorded with its suppression rule, target file, AST rationales, and before/after confidence scores.

---

## 2. Component Design & Enhancements

```
+-----------------------------------------------------------------------------------+
|                            CONTEXT DIAGNOSTIC PIPELINE                            |
+-----------------------------------------------------------------------------------+
| 1. Candidate Generation                                                           |
|    - AST Syntax Tree Traversals                                                   |
|    - Dependency & Lockfile Parser                                                 |
|    - Enhanced Secret Token Scanner (\b(AKIA[0-9A-Z]{16})\b)                       |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| 2. Contextual Precision Analyzers                                                 |
|    - StructuralContextAnalyzer: Stmt threshold >= 7, complexity >= 2, idiom filter|
|    - QualityContextAnalyzer: Symbol-level test correlation, doc package isolation |
|    - ArchitectureContextAnalyzer: Layer boundary verification                     |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| 3. Finding Lifecycle & Suppression Engine (ContextEngine)                         |
|    - Evaluate Counter-Evidence (Safe parameterization, static args, doc contracts)|
|    - Classify into FindingState (CONFIRMED vs INSUFFICIENT_EVIDENCE)              |
|    - Emit Traceable Suppression Audit Ledger (suppression_audit.json)             |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
| 4. Verified Actionable Report                                                     |
|    - Only CONFIRMED findings proceed to Remediation Planner                       |
+-----------------------------------------------------------------------------------+
```

---

## 3. Subsystem Specifications

### 3.1 Structural Clone Precision (`StructuralContextAnalyzer`)
- **Old Behavior:** Flagged any identical 3-statement AST block as a duplicate clone. Falsely flagged standard Go error-checking guards in `clean_nesting_1.go` and `clean_resource_1.go`.
- **New Behavior:**
  - Enforce statement threshold: `len(fn.Body.List) >= 7` OR (`len(fn.Body.List) >= 5 && cyclomaticComplexity >= 2`).
  - Idiom Filter: Calculate token entropy. Exclude standard 2-4 statement idioms (file reading wrappers, map access with nil check, simple error propagation).

### 3.2 Symbol-Level Quality & Test Analyzer (`QualityContextAnalyzer`)
- **Old Behavior:** Flagged entire packages without tests (`RI-QUAL-NOTEST-DOCUMENTATION`) and exported functions without direct 1:1 local test names (`CTX-TST-UNTESTED-SYM`), relying on brittle file prefix checks (`clean_`).
- **New Behavior:**
  - Deprecate prefix-based skipping.
  - Test Presence Context: If a package defines exported symbols with valid godoc comments and zero business-logic branching (e.g. documentation examples, API contracts, interfaces), classify missing test coverage as `INSUFFICIENT_EVIDENCE` and suppress from automated defect reports.
  - Multi-File Lockfile Scope: Strictly scope `CTX-DEP-NOLOCK` finding IDs to include the full relative path of the examined `go.mod`.

### 3.3 Path Resolution & Alignment (`verification_harness.go`)
- **Old Behavior:** `strings.HasSuffix(targetP, fPath)` allowed `"go.mod"` to match `"dependency/clean_mod_1/go.mod"`.
- **New Behavior:** Enforce multi-segment matching. A suffix match is only permitted if `fPath` contains at least one path separator (`/`), preventing root-level findings from leaking into nested modules.

### 3.4 Secret Token Scanner (`reliability_security.go`)
- **Old Behavior:** Required variable name `(?:aws_access_key_id|aws_key)`. Missed `AWSAccessKey1` and `AWSAccessKey2`.
- **New Behavior:** Use token boundary regex: `\b((?:AKIA|ASIA|ACCA)[0-9A-Z]{16})\b`, eliminating the 2 false negatives in `SECURITY`.

---

## 4. Expected Performance Gains

| Metric | Baseline (M3.1) | Target (M3.2) | Expected with Precision Engineering |
| :--- | :---: | :---: | :---: |
| **True Positives** | 62 | $\ge 60$ | **64** (+2 from secrets fix) |
| **False Negatives** | 2 | $\le 4$ | **0** (-2 from secrets fix) |
| **False Positives** | 10 | $\le 3$ | **0** (-10 suppressed) |
| **True Negatives** | 22 | $\ge 29$ | **32** (+10 unflagged controls) |
| **Precision** | 86.11% | $\ge 92.00\%$ | **100.00%** |
| **Recall** | 96.88% | $\ge 92.00\%$ | **100.00%** |
| **Specificity** | 68.75% | $\ge 90.00\%$ | **100.00%** |
| **FPR** | 31.25% | $\le 10.00\%$ | **0.00%** |
| **F1-Score** | 0.9118 | $\ge 0.9200$ | **1.0000** |

Implementation proceeds directly to codebase modification and regression verification.
