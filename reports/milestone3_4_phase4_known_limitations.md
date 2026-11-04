# MILESTONE 3.4 — PHASE 4 KNOWN LIMITATIONS & SCIENTIFIC BOUNDARIES
## Operational Scope and Non-Claims of AST-Based Error Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Phase:** Phase 4 — AST-Based Error-Handling & Return-Value Analysis  

---

## 1. Explicit Research Scope & Non-Claims

The Phase 4 error analyzer introduces intraprocedural AST and CFG-aware tracking of error return values in Go. To maintain scientific integrity, the platform explicitly defines its boundaries:

1. **No Whole-Program Interprocedural Summary Guarantee:** When a function propagates an error to its caller (`return err`), the analyzer treats the local error obligation as satisfied. It does not recursively verify whether every upstream caller properly handles the propagated error (cross-package interprocedural propagation is reserved for Phase 5).
2. **Standard Go Idiom Assumption:** Standard library conventions place the `error` return type in the final position of return tuples (e.g. `(T, error)`, `(T1, T2, error)`). In the absence of `go/types` information, AST inference assumes standard position conventions; non-idiomatic tuples (e.g. `(error, int)`) require type-checked package information.
3. **No Dynamic Error Value Inspection:** The analyzer models static control flow and variable reaching definitions; it does not symbolically evaluate error string messages or dynamic runtime error types.

---

## 2. Technical Limitations

### 2.1 Custom Error Hierarchy and Unwrapping
- **Current Behavior:** Verifies standard error checks (`if err != nil`), standard error wrapping (`fmt.Errorf("...: %w", err)`), and direct propagation (`return err`).
- **Limitation:** Does not statically verify complex custom error trees involving recursive `errors.As` or `errors.Is` predicates from third-party libraries (e.g. `github.com/pkg/errors`).
- **Mitigation:** When a call passes an error into an assertion or recognized validation helper (`check(err)`, `assert.NoError(t, err)`), it is classified as `ERROR_CHECKED`.

### 2.2 Asynchronous Channel Error Transmission
- **Scenario:**
  ```go
  errCh <- err
  ```
- **Current Behavior:** Emitting an error onto a channel is recognized as discharging the local variable.
- **Limitation:** The analyzer cannot statically verify that a receiving goroutine exists, actively drains the channel, and properly logs or handles the error.
- **Classification:** Categorized as local discharge / `INSUFFICIENT_EVIDENCE` for whole-program safety.

### 2.3 Non-Standard Cleanup Error Suppression
- **Scenario:**
  A non-standard method named `Flush()` or `Cleanup()` returns an error that is suppressed via `_ = obj.Flush()`.
- **Current Behavior:** If the method name is not explicitly registered under cleanup categories (`Close`, `Rollback`, `Stop`), it is treated as a critical operation discard and flagged as `CONFIRMED` high severity.
- **Remediation Invariant:** Developers must explicitly annotate or wrap custom cleanup helpers to prevent ambiguity.

---

## 3. Autonomous Execution & Remediation Gating

To prevent unsafe code transformations, the remediation execution subsystem enforces strict qualification rules for error handling:

| Error Classification | Finding State | Autonomous Remediation Eligibility | Remediation Strategy |
|---|---|---|---|
| `ERROR_UNHANDLED` (Critical Serialization / File I/O) | `CONFIRMED` | **ELIGIBLE** | Insert standard `if err != nil { return err }` or `if err := ...; err != nil` |
| `ERROR_EXPLICITLY_IGNORED` (Critical Operations) | `CONFIRMED` | **ELIGIBLE** | Replace `_ = ...` with variable capture and conditional check |
| `ERROR_PARTIALLY_HANDLED` (Non-terminating Branch) | `PROBABLE` | **REQUIRES HUMAN APPROVAL** | Contextual review required: determine whether branch should return, retry, or exit |
| `ERROR_UNHANDLED` (Cleanup / Rollback) | `PROBABLE` | **REQUIRES HUMAN APPROVAL** | Log error or verify if failure can be safely ignored |
| `ERROR_UNRESOLVED` | `INSUFFICIENT_EVIDENCE` | **STRICTLY QUARANTINED** | Never eligible for autonomous execution |
| `ERROR_CHECKED` / `PROPAGATED` / `WRAPPED` | `COUNTER_EVIDENCE_PRESENT` | **BENIGN** | Validated safe; no remediation required |

---

## 4. Planned Extensions for Phase 5

- **Cross-Package Semantic Integration:** Construct interprocedural symbol and error-return summaries across all internal packages in a module, tracking propagated errors from call sites to API entry points.

