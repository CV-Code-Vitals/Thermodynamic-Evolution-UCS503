# Milestone 3.3 — Phase 3: External Diagnostic Intelligence Report

**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Branch:** `validation/milestone-3.3-external-validation`  
**Baseline Commit:** `5d41b531f46d10bc972b2d1518a0fb4cf139c09a`  
**Evaluation Mode:** Frozen-Baseline External Trial  
**Platform:** Windows NT 10.0.26200.0 / Go 1.25.6  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Phase 3 executed the platform's complete repository discovery, structural analysis, and contextual diagnostic engine (`ContextEngine.FilterAndEnrich`) against five external, unseen Go repositories under `benchmarks/external_v1/`.

Zero diagnostic rules, threshold calibrations, or suppression filters were modified prior to or during evaluation.

### Key Empirical Metrics

| Metric | Measured Baseline (External) | Synthetic Suite v3 Baseline | Delta |
|---|---:|---:|---:|
| Total Repositories Evaluated | 5 | N/A (1 monolithic suite) | +5 |
| Total Source LOC | 1,601 | ~4,800 | - |
| Ground Truth Known Defects | 14 | 96 | - |
| Negative Control Invariants | 15 | 48 | - |
| **True Positives (TP)** | **2** | **96** | -94 |
| **False Positives (FP)** | **0** | **0** | 0 |
| **False Negatives (FN)** | **12** | **0** | +12 |
| **True Negatives (TN)** | **15** | **48** | - |
| **Precision** | **100.00%** | **100.00%** | **0.00%** |
| **Recall** | **14.29%** | **100.00%** | **-85.71%** |
| **Specificity** | **100.00%** | **100.00%** | **0.00%** |
| **F1 Score** | **25.00%** | **100.00%** | **-75.00%** |
| **False Positive Rate (FPR)** | **0.00%** | **0.00%** | **0.00%** |
| **False Negative Rate (FNR)** | **85.71%** | **0.00%** | **+85.71%** |
| **Suppression Rate** | **28.57%** | **33.33%** | **-4.76%** |
| **Human-Investigation Rate** | **0.00%** | **0.00%** | **0.00%** |

---

## 2. Per-Repository Confusion Matrix & Findings

```
                              DIAGNOSTIC ADJUDICATION BREAKDOWN
                                
   Repository          TP    FP    FN    TN    Precision    Recall    F1 Score
   -------------------------------------------------------------------------
   http_router          0     0     3     3        0.0%       0.0%       0.0%
   sql_service          0     0     3     3        0.0%       0.0%       0.0%
   cache_engine         0     0     2     3        0.0%       0.0%       0.0%
   cli_tooling          1     0     2     3      100.0%      33.3%      50.0%
   queue_worker         1     0     2     3      100.0%      33.3%      50.0%
   -------------------------------------------------------------------------
   AGGREGATE TOTAL      2     0    12    15      100.0%      14.3%      25.0%
```

### 2.1. `ext-repo-01: http_router`
- **Discovered Files:** 6
- **LOC:** 312
- **Raw Alerts:** 4
- **Active Diagnostic Findings:** 8 (all categorized as `VALID_CONCERN`)
  - CI Workflow Missing (`RI-QUAL-NOCI-WORKFLOWS`)
  - Documentation / License Missing (`RI-HYG-MISSING-README`, `RI-HYG-MISSING-LICENSE`)
  - Dependency Lockfile Missing (`RI-DEP-GOSUM-MISSING-ROOT`)
  - Untested Exported Symbols (`CTX-TST-UNTESTED-SYM-...` on 4 files)
- **Defects Detected:** 0 (3 Ground Truth defects undetected due to linear function scope and expression statements)
- **Adjudication:** TP=0, FP=0, FN=3, TN=3.

### 2.2. `ext-repo-02: sql_service`
- **Discovered Files:** 6
- **LOC:** 332
- **Active Diagnostic Findings:** 6 (`VALID_CONCERN`)
- **Defects Detected:** 0 (3 Ground Truth defects undetected: `fmt.Sprintf` dynamic SQL formatting not matched by `ast.BinaryExpr`, `rows.Close()` unclosed outside loop, unhandled `tx.Rollback()`)
- **Adjudication:** TP=0, FP=0, FN=3, TN=3.

### 2.3. `ext-repo-03: cache_engine`
- **Discovered Files:** 6
- **LOC:** 315
- **Active Diagnostic Findings:** 7 (`VALID_CONCERN`)
- **Defects Detected:** 0 (2 Ground Truth defects undetected: struct-level lock omission and background goroutine leak not modeled by static AST rules)
- **Adjudication:** TP=0, FP=0, FN=2, TN=3.

### 2.4. `ext-repo-04: cli_tooling`
- **Discovered Files:** 6
- **LOC:** 284
- **Active Diagnostic Findings:** 7 (1 `CONFIRMED_DEFECT`, 6 `VALID_CONCERN`)
- **Confirmed Defect Detected:**
  - `CTX-REL-DEFER-IN-LOOP-fs_go-L38` at `fs.go:38` (Severity: Medium, Calibrated Confidence: 0.95). Correctly identified file descriptor resource accumulation inside loop body.
- **Undetected Defects:** `ExecuteShellScript` command injection via function argument parameter; `ReadFileFromRoot` path traversal.
- **Adjudication:** TP=1, FP=0, FN=2, TN=3. Precision: 100.0%, Recall: 33.3%.

### 2.5. `ext-repo-05: queue_worker`
- **Discovered Files:** 7
- **LOC:** 358
- **Active Diagnostic Findings:** 7 (1 `CONFIRMED_DEFECT`, 6 `VALID_CONCERN`)
- **Confirmed Defect Detected:**
  - `RI-REL-IGNORED-ERR-INTERNAL_CONSUMER_CONSUMER_GO-L47` at `internal/consumer/consumer.go:47` (Severity: Low, Calibrated Confidence: 0.85). Correctly identified `_ = c.handler(ctx, msg)` discarding error return value.
- **Undetected Defects:** `SendDirect` unclosed TCP socket; `RetryWithTimerLeak` timer allocation without `Stop()`.
- **Adjudication:** TP=1, FP=0, FN=2, TN=3. Precision: 100.0%, Recall: 33.3%.

---

## 3. Suppression Efficiency Audit

The contextual engine suppressed 14 raw alerts across the evaluation:
- In `http_router`: suppressed 4 invalid heuristic alerts on safe trie node and context structures.
- In `sql_service`: suppressed 2 alerts on safe parameterized queries in `CreateUser` and `GetUserByID`.
- In `cache_engine`: suppressed 3 alerts on safe mutex-protected map operations in `SafeKeys`.
- In `cli_tooling`: suppressed 3 alerts on safe direct process execution in `ExecuteDirect`.
- In `queue_worker`: suppressed 2 alerts on safe backoff retry operations.

Every suppression was corroborated by positive counter-evidence in the contextual engine's suppression audit trail.

---

## 4. Evaluation Conclusion

The external evaluation establishes that the platform's false-positive suppression is exceptionally robust: in no case did the system report a defect on correct code. However, the system's defect detection rules are brittle and tightly coupled to the specific AST structures present in synthetic suites. Production qualification requires expanding the AST representations and introducing inter-procedural dataflow.

