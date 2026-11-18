# CI Hotfix — Benchmark Recall Regression & Cross-Platform Engine Resolution

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Date:** 2026-11-18  
**Component:** `go-backend/agent/repo_intel/benchmark/verification_harness.go`  

---

## 1. Frozen Benchmark Baseline & Cryptographic Hashes

To ensure strict scientific integrity, the benchmark manifest and dataset were frozen prior to audit:

- **Manifest Path:** [`benchmarks/ground_truth_manifest_v3.json`](file:///E:/SWE/benchmarks/ground_truth_manifest_v3.json)
- **Manifest SHA-256:** `c1ccbd657eb090d6daf948fccedc47428f31eb2ffaa8b275054823f684b56689`
- **Total Benchmark Fixtures:** 144
- **True Defect Fixtures:** 96 (8 categories × 12 defects each)
- **Clean Negative Control Fixtures:** 48 (8 categories × 6 controls each)
- **Acceptance Threshold:** Precision $\ge 0.90$, Recall $\ge 0.90$, Specificity $\ge 0.90$, FPR $\le 0.10$, F1 $\ge 0.90$

---

## 2. Regression Accounting & Identification of False Negatives

During the failing CI execution, detection metrics dropped to:
- **True Positives (TP):** 84
- **False Positives (FP):** 0
- **False Negatives (FN):** 12
- **True Negatives (TN):** 48
- **Recall:** $\frac{84}{84 + 12} = \mathbf{0.8750}$ ($< 0.90$ threshold $\rightarrow$ **FAIL**)

### Forensic Audit of the 12 False Negative Cases:
Analysis of [`reports/milestone3_2_detection_metrics.json`](file:///E:/SWE/reports/milestone3_2_detection_metrics.json) confirmed that 7 of the 8 categories achieved 100% recall (12/12 TP, 0 FN). The exact 12 false negatives occurred exclusively within the **`STRUCTURAL`** category:

| Fixture ID | Category | Target File | Defect Description | CI Outcome |
| :--- | :--- | :--- | :--- | :---: |
| `FIX-V3-STR-NEST-01` | STRUCTURAL | `structural/deep_nest_tree_1.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-NEST-02` | STRUCTURAL | `structural/deep_nest_tree_2.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-NEST-03` | STRUCTURAL | `structural/deep_nest_tree_3.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-NEST-04` | STRUCTURAL | `structural/deep_nest_tree_4.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-NEST-05` | STRUCTURAL | `structural/deep_nest_tree_5.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-NEST-06` | STRUCTURAL | `structural/deep_nest_tree_6.go` | Excessive control flow nesting depth ($\ge 5$ levels) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-07` | STRUCTURAL | `structural/high_complexity_calc_1.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-08` | STRUCTURAL | `structural/high_complexity_calc_2.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-09` | STRUCTURAL | `structural/high_complexity_calc_3.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-10` | STRUCTURAL | `structural/high_complexity_calc_4.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-11` | STRUCTURAL | `structural/high_complexity_calc_5.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |
| `FIX-V3-STR-COMPLEX-12` | STRUCTURAL | `structural/high_complexity_calc_6.go` | Cyclomatic complexity exceeding cognitive threshold ($\ge 10$) | **FALSE NEGATIVE** |

---

## 3. Root Cause: Environment & Execution Discrepancy

1. **Non-Detector Origin:** The regression was **not** caused by changes to diagnostic logic, context reconciliation, or suppression filters.
2. **Missing Binary on Linux:** The `STRUCTURAL` detector delegates AST metrics to `thermodynamic-ast-engine`. In `verification_harness.go:491`, `rustCandidates` only contained `.exe` file paths. On Linux CI runners (`ubuntu-latest`), the compiled binary is `thermodynamic-ast-engine` (without `.exe`).
3. **Environment Variable Ignored:** The workflow set `ENGINE_PATH`, but the benchmark harness never read it.
4. **Silent Skip:** The structural analysis step failed silently, generating 0 structural findings and causing all 12 structural defect fixtures to be recorded as false negatives.

---

## 4. Minimal Justified Correction

The minimal, non-invasive correction was implemented in [`go-backend/agent/repo_intel/benchmark/verification_harness.go`](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/verification_harness.go) and [`go-backend/agent/repo_intel/benchmark/stage_runner.go`](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/stage_runner.go):

```go
func ResolveRustEngineBinary() string {
    if custom := os.Getenv("ENGINE_PATH"); custom != "" {
        if fi, err := os.Stat(custom); err == nil && !fi.IsDir() {
            if abs, err := filepath.Abs(custom); err == nil {
                return abs
			}
            return custom
		}
	}

    binNames := []string{"thermodynamic-ast-engine"}
    if runtime.GOOS == "windows" {
        binNames = []string{"thermodynamic-ast-engine.exe", "thermodynamic-ast-engine"}
    } else {
        binNames = []string{"thermodynamic-ast-engine", "thermodynamic-ast-engine.exe"}
    }

    var candidates []string
    relPrefixes := []string{
        "", ".", "..",
        filepath.Join("..", ".."),
        filepath.Join("..", "..", ".."),
        filepath.Join("..", "..", "..", ".."),
        filepath.Join("..", "..", "..", "..", ".."),
    }

    for _, pfx := range relPrefixes {
        for _, bin := range binNames {
            candidates = append(candidates,
                filepath.Join(pfx, "thermodynamic-ast-engine", "target", "release", bin),
            )
        }
    }
    for _, pfx := range relPrefixes {
        for _, bin := range binNames {
            candidates = append(candidates,
                filepath.Join(pfx, "thermodynamic-ast-engine", "target", "debug", bin),
            )
        }
    }

    for _, cand := range candidates {
        if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
            if abs, err := filepath.Abs(cand); err == nil {
                return abs
            }
            return cand
        }
    }
    return ""
}
```

### Key Properties:
- Prioritizes `os.Getenv("ENGINE_PATH")` directly.
- Inspects both native and extension-bearing binary names cross-platform.
- Searches only valid `thermodynamic-ast-engine/target/{release,debug}` compilation outputs, preventing contamination from stale legacy binaries in other folders.

---

## 5. Re-Evaluation Results (Frozen Threshold Comparison)

| Metric | Target Threshold | Failing CI Score | Corrected Score | Status |
| :--- | :---: | :---: | :---: | :---: |
| **True Positives** | - | 84 | **96** | **RESTORED (+12)** |
| **False Negatives** | - | 12 | **0** | **ELIMINATED (-12)** |
| **False Positives** | - | 0 | **0** | **ZERO FALSE POSITIVES** |
| **True Negatives** | - | 48 | **48** | **ALL CONTROLS PRESERVED** |
| **Overall Precision** | $\ge 0.90$ | 1.0000 | **1.0000** [0.9615, 1.0000] | **PASS** |
| **Overall Recall** | $\ge 0.90$ | 0.8750 | **1.0000** [0.9615, 1.0000] | **PASS** |
| **Overall Specificity** | $\ge 0.90$ | 1.0000 | **1.0000** [0.9259, 1.0000] | **PASS** |
| **Overall FPR** | $\le 0.10$ | 0.0000 | **0.0000** [0.0000, 0.0741] | **PASS** |
| **Overall F1 Score** | $\ge 0.90$ | 0.9333 | **1.0000** | **PASS** |

**Zero Threshold Weakening:** No criteria were diluted, no ground-truth annotations were altered, and no detections were manufactured.
