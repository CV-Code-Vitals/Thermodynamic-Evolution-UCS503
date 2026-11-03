# MILESTONE 3.2 — PHASE 0: BASELINE FREEZE & REPRODUCTION REPORT
## Precision Engineering, False-Positive Suppression & Controlled Production Pilot

**Project:** Autonomous Self-Healing Repository Intelligence Pipeline  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `4dbd599a337d0ff426e84d8a248091713a7acd80`  
**Target Branch:** `research/milestone-3.2-precision-engineering`  
**Phase Status:** COMPLETED & FROZEN  
**Audit Timestamp:** 2026-11-03  

---

## 1. Executive Summary & Mission Scope

Milestone 3.1 established an independently audited and reproducible baseline for the autonomous repository intelligence and diagnostic pipeline across the 96-fixture expanded benchmark (`benchmarks/self_healing_suite_v2`). The independent audit verified high defect recall (96.88%, 62/64 defects detected) and solid overall precision (86.11%, F1 = 91.18%). 

However, the empirical evaluation revealed a critical bottleneck in specificity and false-positive handling:
- **True Positives (TP):** 62
- **False Positives (FP):** 10
- **True Negatives (TN):** 22
- **False Negatives (FN):** 2
- **Specificity:** 68.75% (95% CI: [51.43%, 82.05%])
- **False Positive Rate (FPR):** 31.25% (95% CI: [17.95%, 48.57%])

In a real-world software repository, an FPR of 31.25% on non-defective code generates unacceptable engineer alert fatigue and prevents autonomous remediation from running safely without continuous human supervision.

**Milestone 3.2 Mission:**
1. Perform forensic root-cause analysis on the 10 False Positives observed in Suite v2.
2. Architect and implement **Contextual Precision Engineering** across all AST analyzers and diagnostic engines to suppress heuristic false positives without sacrificing recall.
3. Conduct root-cause forensics on the 2 missed Security False Negatives and enhance detection accuracy.
4. Construct and validate an expanded, independent **Benchmark Suite v3** (with >=96 defects, >=48 clean controls, and real-world repository subsets) with zero structural leakage.
5. Deploy and execute controlled, read-only autonomous diagnostic pilots across three real-world production Go repositories.

---

## 2. Baseline Reproduction & Metric Verification

Phase 0 executed the automated verification harness `TestScientificVerificationStage2` under Go 1.25.6 (`windows/amd64`). The baseline metrics from Milestone 3.1 were reproduced identically with zero variance:

| Metric | Suite v2 Baseline Value | 95% Confidence Interval | Sample Count |
| :--- | :---: | :---: | :---: |
| **Total Evaluated Fixtures** | 96 | — | 96 / 96 |
| **Total Planted Defects** | 64 | — | 64 / 96 |
| **Total Clean Controls** | 32 | — | 32 / 96 |
| **True Positives (TP)** | 62 | — | 62 / 64 defects |
| **False Negatives (FN)** | 2 | — | 2 / 64 defects |
| **False Positives (FP)** | 10 | — | 10 / 32 controls |
| **True Negatives (TN)** | 22 | — | 22 / 32 controls |
| **Precision** | **86.11%** | [76.29%, 92.28%] | 62 / 72 predictions |
| **Recall (Sensitivity)** | **96.88%** | [89.30%, 99.14%] | 62 / 64 defects |
| **Specificity** | **68.75%** | [51.43%, 82.05%] | 22 / 32 clean controls |
| **False Positive Rate (FPR)** | **31.25%** | [17.95%, 48.57%] | 10 / 32 clean controls |
| **False Negative Rate (FNR)** | **3.13%** | [0.86%, 10.70%] | 2 / 64 defects |
| **F1 Score** | **0.9118** | — | Harmonic mean |

### Per-Category Baseline Performance

```
+-----------------+----------+---------+---------+---------+---------+---------+-------------+
| Category        | Fixtures | Defects | Clean   | TP / FN | FP / TN | Prec %  | Specificity |
+-----------------+----------+---------+---------+---------+---------+---------+-------------+
| ARCHITECTURE    | 12       | 8       | 4       | 8 / 0   | 0 / 4   | 100.0%  | 100.0%      |
| DEPENDENCY      | 12       | 8       | 4       | 8 / 0   | 2 / 2   | 80.00%  |  50.0%      |
| DOCUMENTATION   | 12       | 8       | 4       | 8 / 0   | 4 / 0   | 66.67%  |   0.0%      |
| MAINTAINABILITY | 12       | 8       | 4       | 8 / 0   | 0 / 4   | 100.0%  | 100.0%      |
| RELIABILITY     | 12       | 8       | 4       | 8 / 0   | 2 / 2   | 80.00%  |  50.0%      |
| SECURITY        | 12       | 8       | 4       | 6 / 2   | 0 / 4   | 100.0%  | 100.0%      |
| STRUCTURAL      | 12       | 8       | 4       | 8 / 0   | 2 / 2   | 80.00%  |  50.0%      |
| TEST_QUALITY    | 12       | 8       | 4       | 8 / 0   | 0 / 4   | 100.0%  | 100.0%      |
+-----------------+----------+---------+---------+---------+---------+---------+-------------+
| TOTAL OVERALL   | 96       | 64      | 32      | 62 / 2  | 10 / 22 | 86.11%  |  68.75%     |
+-----------------+----------+---------+---------+---------+---------+---------+-------------+
```

---

## 3. Forensic Identification of Defect Deficiencies

The baseline reveals two distinct technical anomalies that require architectural intervention:

### 3.1 The 10 False Positives
1. **Documentation Category (4 False Positives):**
   - Fixtures: `documented_api_1.go`, `documented_api_2.go`, `documented_api_3.go`, `documented_api_4.go`
   - Mechanism: The heuristic scanner flags `RI-QUAL-NOTEST-DOCUMENTATION` and `CTX-TST-UNTESTED-SYM` because the documentation package directory contained zero `*_test.go` files. The detector improperly coupled package test presence to symbol documentation validity.
2. **Structural & Reliability Categories (4 False Positives):**
   - Fixtures: `clean_nesting_1.go`, `clean_nesting_2.go`, `clean_resource_1.go`, `clean_resource_2.go`
   - Mechanism: The AST structural clone detector (`StructuralContextAnalyzer`) flags identical 3-statement boilerplate function structures across distinct files as duplicated code (`RI-STRUCT-DUPLICATE-BLOCK`).
3. **Dependency Category (2 False Positives):**
   - Fixtures: `clean_mod_1/go.mod`, `clean_mod_2/go.mod`
   - Mechanism: File-path suffix matching mistakenly attributed a workspace-level missing checksum finding (`RI-DEP-GOSUM-MISSING-ROOT`) to every isolated fixture containing a `go.mod` file.

### 3.2 The 2 Security False Negatives
1. Fixture: `sec_def_hardcoded_jwt_1.go` (Hardcoded Secret Token).
2. Fixture: `sec_def_cmd_exec_taint_1.go` (Command Injection via tainted parameter).
- Mechanism: The AST security analyzer failed to trace the synthetic assignment pattern across intermediate variables, resulting in missed true defects.

---

## 4. Milestone 3.2 Engineering Objectives & Targets

To declare Milestone 3.2 successful and achieve production qualification, the pipeline must satisfy the following strict quantitative criteria:

| Metric | Suite v2 Baseline | Suite v3 Target | Production Requirement |
| :--- | :---: | :---: | :---: |
| **Precision** | 86.11% | **>= 92.00%** | >= 90.00% |
| **Recall** | 96.88% | **>= 92.00%** | >= 90.00% |
| **Specificity** | 68.75% | **>= 90.00%** | >= 90.00% |
| **False Positive Rate (FPR)** | 31.25% | **<= 10.00%** | <= 10.00% |
| **F1-Score** | 91.18% | **>= 92.00%** | >= 90.00% |
| **Execution Safety Violations** | 0 | **0** | **0 (Strict INV-1 to INV-5)** |
| **Remediation Regression Rate** | 0.00% | **0.00%** | **0.00%** |

---

## 5. Artifact Provenance & Integrity Hashes

The cryptographic fingerprints of all baseline inputs, source engines, and audit documentation are recorded in `reports/milestone3_2_artifact_hashes.json`. Any divergence during subsequent phases will be immediately detected.

All Phase 0 baseline reproduction criteria have been met. Milestone 3.2 proceeds directly to **Phase 1: False-Positive Forensic Analysis**.
