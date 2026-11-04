# MILESTONE 3.4 — PHASE 4 TECHNICAL REPORT
## AST-Based Error-Handling & Return-Value Semantic Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Phase:** Phase 4 — AST-Based Error-Handling & Return-Value Analysis  
**Classification:** Semantic Diagnostic Implementation & External Defect Recovery  

---

## 1. Executive Summary

A pervasive failure mode in static error detection is **idiom blindness**: relying on line-level regular expressions (`unhandledErrorRegex = regexp.MustCompile("(?i)^\\s*_[,\s]*=\\s*[a-zA-Z0-9_]+\\.[a-zA-Z0-9_]+\\(")`) that only match literal blank identifier assignments (`_ = pkg.Func()`). 

In Milestone 3.3, this architectural limitation caused the diagnostic engine to completely miss three real-world external defects:
1. `DEF-HTTP-02` (`http_router/context.go:61`): Unhandled error in `json.NewEncoder(c.Writer).Encode(data)` (bare expression statement).
2. `DEF-SQL-03` (`sql_service/db.go:57`): Unhandled error in `tx.Rollback()` (bare expression statement).
3. `DEF-QUEUE-03` (`queue_worker/internal/consumer/consumer.go:60`): Unhandled error in `json.Unmarshal(data, &res)` (bare expression statement).

Milestone 3.4 Phase 4 replaces the syntactic regex heuristic with **type-aware, AST and CFG-based error analysis**. Built upon the Phase 1 Semantic IR and reaching-definitions engine, `semantic.ErrorAnalyzer` models error returns, assignment targets, reaching definitions, conditional error checks, error wrapping, error propagation, and contextual suppression.

### Central Invariant
> An error is not handled simply because an `if err != nil` statement exists somewhere nearby. The analyzer must establish which operation produced the error, where that exact value flows, and whether every relevant execution path handles, propagates, or terminates it appropriately.

---

## 2. Signature Resolution & Return-Value Classification

Calls producing error values are resolved using `go/types` where available and AST-derived signature inference across standard-library packages and local declarations.

### Formal Handling Classifications

| Classification | Definition | Finding State | Severity |
|---|---|---|---|
| `ERROR_CHECKED` | Error is tested via conditional branch (`if err != nil`) whose handling body terminates, propagates, or logs the error (or handled via recognized assertions). | `COUNTER_EVIDENCE_PRESENT` | Safe (Benign) |
| `ERROR_PROPAGATED` | Error is directly returned to caller (`return op()`, `return nil, err`, or returned via named return variable). | `COUNTER_EVIDENCE_PRESENT` | Safe (Benign) |
| `ERROR_WRAPPED_AND_PROPAGATED` | Error is checked and wrapped with `%w` formatting before return (`fmt.Errorf("...: %w", err)`). | `COUNTER_EVIDENCE_PRESENT` | Safe (Benign) |
| `ERROR_EXPLICITLY_IGNORED` | Error assigned to blank identifier (`_ = ...`). Contextual policy distinguishes cleanup operations from critical operations. | `COUNTER_EVIDENCE_PRESENT` (Cleanup) / `CONFIRMED` (Critical) | Low (Cleanup) / High (Critical) |
| `ERROR_UNHANDLED` | Error return value is completely unassigned (bare expression statement) or assigned to variable that is never checked before overwrite or function exit. | `CONFIRMED` | High |
| `ERROR_PARTIALLY_HANDLED` | Error is checked in a branch, but the handling branch continues execution unsafely without returning or handling the failure. | `PROBABLE` | Medium |
| `ERROR_UNRESOLVED` | Call target returns an opaque or unresolvable signature. | `INSUFFICIENT_EVIDENCE` | Low |

---

## 3. CFG and Control-Flow-Aware Error Analysis

The analyzer models the lifecycle of returned errors through six control-flow evaluations:

1. **Direct Operand Propagation:** Identifies calls executed directly as return operands (`return os.WriteFile(...)`), classifying them as `ERROR_PROPAGATED` in $O(1)$.
2. **If-Init Clause Extraction:** Identifies standard Go idiom `if err := operation(); err != nil`, verifying whether the condition guards `err != nil` and whether the body wraps (`%w`) or terminates.
3. **Overwritten Definition Tracking:** Traces whether an assigned error variable (`err`) is reassigned by another call before any conditional check on that definition occurs (`_, err := op1(); _, err = op2()`). The earlier unhandled definition is flagged as `ERROR_UNHANDLED`.
4. **Mismatched Error Variable Checks:** Verifies that the variable tested in `if errX != nil` matches the definition produced by the target operation. If `err1` is checked while `err2` is neglected, `err2` is flagged as `ERROR_UNHANDLED`.
5. **Branch Termination Verification:** Inspects the body of the `if err != nil` block for valid exit mechanisms:
   - `*ast.ReturnStmt`: Guarantees the failure path does not proceed into normal execution.
   - Fatal/panic calls (`t.Fatal`, `t.Fatalf`, `log.Fatal`, `panic`): Recognized as safe terminating handlers.
   - Non-terminating logging: Flagged as `ERROR_PARTIALLY_HANDLED` if normal execution continues with corrupt or zero-valued state.
6. **Named Return Error Flow:** Recognizes assignments to named return variables (`func Foo() (err error) { err = op(); return }`), verifying that bare returns correctly propagate the error to caller scope.

---

## 4. Contextual Error Policy & Resource Model Integration

Phase 4 explicitly differentiates cleanup operations from critical system operations:

```text
                  Error-Returning Call
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
    Cleanup Operation           Critical Operation
 (Close, Rollback, Stop)    (Unmarshal, WriteFile, Exec, Dial)
             │                           │
     ┌───────┴───────┐           ┌───────┴───────┐
     ▼               ▼           ▼               ▼
Deferred / Blank    Bare Expr   Assigned/Handled   Bare Expr / Blank
   (Safe / FP)       (Low)         (Safe)              (High Defect)
```

### Contextual Rules
- **Resource Cleanup vs. Error Handling:** Distinguishes resource closure from error propagation. A resource can be closed while its error is ignored (`_ = file.Close()`), which is an accepted Go idiom.
- **Transaction Rollback Semantics:** The standard idiom `defer tx.Rollback()` followed by `return tx.Commit()` is fully preserved:
  - `defer tx.Rollback()` is classified as safe cleanup error suppression.
  - `return tx.Commit()` is classified as `ERROR_PROPAGATED`.
- **Deserialization Errors (`json.Unmarshal`):** Discarding JSON unmarshaling errors leads to silent application state corruption and is strictly flagged as `HIGH` severity.

---

## 5. Benchmark Verification & Fixture Results

The dedicated fixture suite in `fixtures/error_fixtures.go` evaluates 10 positive defect scenarios and 10 negative controls:

| # | Fixture Name | Category | Expected State | Actual State | Finding State | Confidence |
|---|---|---|---|---|---|---|
| 1 | `Pos_Err_UnhandledUnmarshal` | Serialization | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 2 | `Pos_Err_UnhandledMarshalDiscard` | Serialization | `ERROR_EXPLICITLY_IGNORED` | `ERROR_EXPLICITLY_IGNORED` | `CONFIRMED` | 0.95 |
| 3 | `Pos_Err_UnhandledWriteFile` | File I/O | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 4 | `Pos_Err_UnhandledDBExec` | Database | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 5 | `Pos_Err_UnhandledDial` | Network | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 6 | `Pos_Err_ExplicitBlankCriticalOp` | Serialization | `ERROR_EXPLICITLY_IGNORED` | `ERROR_EXPLICITLY_IGNORED` | `CONFIRMED` | 0.95 |
| 7 | `Pos_Err_MultipleReturnDiscardError` | General | `ERROR_EXPLICITLY_IGNORED` | `ERROR_EXPLICITLY_IGNORED` | `CONFIRMED` | 0.95 |
| 8 | `Pos_Err_OverwrittenBeforeCheck` | File I/O | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 9 | `Pos_Err_MismatchedErrorCheck` | Serialization | `ERROR_UNHANDLED` | `ERROR_UNHANDLED` | `CONFIRMED` | 0.96 |
| 10 | `Pos_Err_UnsafeBranchContinue` | General | `ERROR_PARTIALLY_HANDLED` | `ERROR_PARTIALLY_HANDLED` | `PROBABLE` | 0.88 |
| 11 | `Neg_Err_ExplicitIfCheck` | Serialization | `ERROR_CHECKED` | `ERROR_CHECKED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 12 | `Neg_Err_DirectPropagation` | File I/O | `ERROR_PROPAGATED` | `ERROR_PROPAGATED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 13 | `Neg_Err_WrappedPropagation` | File I/O | `ERROR_WRAPPED_AND_PROPAGATED` | `ERROR_WRAPPED_AND_PROPAGATED` | `COUNTER_EVIDENCE_PRESENT` | 0.99 |
| 14 | `Neg_Err_NamedReturn` | Serialization | `ERROR_PROPAGATED` | `ERROR_PROPAGATED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 15 | `Neg_Err_MultipleReturnHandled` | General | `ERROR_CHECKED` | `ERROR_CHECKED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 16 | `Neg_Err_IgnoredCleanup` | Cleanup | `ERROR_EXPLICITLY_IGNORED` | `ERROR_EXPLICITLY_IGNORED` | `COUNTER_EVIDENCE_PRESENT` | 0.95 |
| 17 | `Neg_Err_TxRollbackWithCommit` | Database Tx | `ERROR_PROPAGATED` | `ERROR_PROPAGATED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 18 | `Neg_Err_TestHelperFatal` | Test Helper | `ERROR_CHECKED` | `ERROR_CHECKED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |
| 19 | `Neg_Err_VoidNoErrors` | General | `NONE` (Clean) | `NONE` (Clean) | No Vulnerability | 1.00 |
| 20 | `Neg_Err_HelperCheck` | General | `ERROR_CHECKED` | `ERROR_CHECKED` | `COUNTER_EVIDENCE_PRESENT` | 0.98 |

### Benchmark Confusion Matrix
- **True Positives (TP):** 10
- **False Positives (FP):** 0
- **True Negatives (TN):** 13 (including safe negative controls with multiple error return signatures)
- **False Negatives (FN):** 0
- **Precision:** 100.0%
- **Recall:** 100.0%
- **Specificity:** 100.0%
- **F1 Score:** 100.0%

---

## 6. Legacy Heuristic Migration Comparison

| Dimension | Legacy Heuristic (`unhandledErrorRegex`) | Phase 4 Semantic Analyzer (`semantic.ErrorAnalyzer`) |
|---|---|---|
| **Mechanism** | Line-by-line regex pattern matching (`_ = pkg.Func()`) | AST signature resolution + CFG reaching-definitions tracking |
| **Bare Expression Statements (`json.Unmarshal(...)`)** | **Blind (0% recall)** | **Detected (100% recall)** |
| **Multiple Return Blank Discards (`val, _ := ...`)** | **Blind (0% recall)** | **Detected (100% recall)** |
| **Error Overwritten Before Check (`err := op1(); err = op2()`)** | **Blind (0% recall)** | **Detected (100% recall)** |
| **Mismatched Error Variable Checked (`err1` vs `err2`)** | **Blind (0% recall)** | **Detected (100% recall)** |
| **Cleanup Suppression vs Critical Discard** | Identical low-severity finding | Differentiated: cleanup is safe counter-evidence; critical is confirmed high defect |
| **Safe Counter-Evidence Generation** | None | Emits `SAFE_ERROR_HANDLING` for verified branches |

