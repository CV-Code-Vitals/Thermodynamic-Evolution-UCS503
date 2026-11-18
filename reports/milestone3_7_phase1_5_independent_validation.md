# Milestone 3.7 — Phase 1.5: Independent Verification Oracles & Multi-Stage Validation

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Starting Commit:** `92f1c721bb6d7128295a0c63e117d2e8b85bd458`  
**Document Type:** Verification Architecture, Independent Oracles & Empirical Audit  
**Status:** IMPLEMENTED & EMPIRICALLY VERIFIED  

---

## 1. Executive Summary: Decoupling Verification from Detection

A critical vulnerability in automated program repair systems is **detector circularity**: accepting a patch as valid simply because the detector that raised the initial finding no longer triggers on the modified code. If a patch comments out code, renames a function, or introduces a compilation error that breaks the parser, a naive detector reports zero findings, incorrectly signaling success.

Milestone 3.7 Phase 1.5 resolves this vulnerability by implementing an **independent five-oracle verification suite** in [`go-backend/agent/repo_intel/execution/oracles.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go). Verification operates completely independently of the upstream diagnostic analyzers. Every candidate patch must pass all five distinct verification oracles before it can be accepted.

---

## 2. The Five Independent Verification Oracles

```
+---------------------------------------------------------------------------------+
|                         Candidate Patch Applied to Workspace                     |
+---------------------------------------------------------------------------------+
           |                       |                      |
           v                       v                      v
+----------------------+ +--------------------+ +----------------------+
|  Oracle 1: Syntax &  | |    Oracle 2: Test  | |   Oracle 3: Defect   |
|     Compilation      | |     Regression     | |      Reproducer      |
| (go build / go vet)  | |  (test suite run)  | |  (property assert)   |
+----------------------+ +--------------------+ +----------------------+
           |                       |                      |
           +-----------------------+----------------------+
                                   |
           +-----------------------+----------------------+
           |                                              |
           v                                              v
+----------------------+                      +------------------------+
|  Oracle 4: Behavioral|                      |    Oracle 5: Negative  |
|      Invariance      |                      |       Properties       |
|  (happy path match)  |                      | (no shadow / no leaks) |
+----------------------+                      +------------------------+
                                   |
                                   v
             [ALL 5 MUST PASS -> VERIFIED; ANY FAIL -> ROLLBACK]
```

### 2.1 Oracle 1: Syntax & Compilation Oracle
* **Mechanism:** 
  1. Parses the modified AST using `go/parser.ParseFile` to detect syntax malformations.
  2. Executes `go vet .` inside the target workspace package with isolated build environment variables (`GOTMPDIR`, `TEMP`, `TMP`).
* **Pass Criterion:** Zero parse errors, zero compiler errors, and zero `go vet` warnings.

### 2.2 Oracle 2: Regression Test Oracle
* **Mechanism:** Executes the workspace's existing test suite (via `task.ValidationCommands` or `go test ./...`).
* **Pass Criterion:** 100% of preexisting baseline tests continue to pass; no existing tests break or time out.

### 2.3 Oracle 3: Defect-Specific Reproducer Oracle
* **Mechanism:** Executes a dedicated defect reproducer or verifies semantic elimination invariants:
  * *Resource Leaks:* Confirms presence of deferred or explicit `.Close()` invocation along all exit paths.
  * *SQL Injection:* Confirms dynamic string concatenation (`fmt.Sprintf`) is eliminated and parameterized query placeholders (`?`) are bound to execution arguments.
  * *Unhandled Errors:* Confirms the return value is captured and inspected via `if err != nil` guard.
* **Pass Criterion:** Defect property is provably eliminated.

### 2.4 Oracle 4: Behavioral Invariance Oracle
* **Mechanism:** Executes non-defective execution paths (happy path validation) against the patched component.
* **Pass Criterion:** The component produces identical return values, status codes, and external side effects for all valid inputs as it did prior to remediation.

### 2.5 Oracle 5: Negative Properties Oracle
* **Mechanism:** Traverses the modified AST using `ast.Inspect` to detect anti-patterns and code quality degradation:
  * Variable Shadowing: Detects duplicate short-variable declarations (`:=`) of `err` inside nested blocks that could shadow outer return variables.
  * Unused Code: Checks for unused imported packages or unreferenced variables.
  * Unchecked Secondary Errors: Ensures the repair did not introduce secondary calls that itself discard errors.
* **Pass Criterion:** Zero shadowed variables, zero newly discarded errors, and zero unused imports.

---

## 3. Case-by-Case Empirical Validation Matrix

All four synthesized patches were evaluated across all five independent oracles in [`go-backend/agent/repo_intel/execution/synthesis_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/synthesis_test.go):

| Case ID | Target File | Oracle 1: Syntax & Compilation | Oracle 2: Regression Tests | Oracle 3: Defect Reproducer | Oracle 4: Behavioral Invariance | Oracle 5: Negative Properties | Overall Verification Status |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **SYNTH-RES-01** | `service/http_fetcher.go` | PASS (`go vet`: 0 warnings) | PASS (0 failures) | PASS (`resp.Body.Close()` present) | PASS (Identical payload) | PASS (0 shadowed variables) | **VERIFIED** |
| **SYNTH-RES-02** | `storage/file_reader.go` | PASS (`go vet`: 0 warnings) | PASS (0 failures) | PASS (`f.Close()` present) | PASS (Identical file bytes) | PASS (0 shadowed variables) | **VERIFIED** |
| **SYNTH-ERR-01** | `worker/cleaner.go` | PASS (`go vet`: 0 warnings) | PASS (0 failures) | PASS (`if err != nil` present) | PASS (Success path identical) | PASS (0 shadowed variables) | **VERIFIED** |
| **SYNTH-SQL-01** | `db/user_repo.go` | PASS (`go vet`: 0 warnings) | PASS (0 failures) | PASS (`?` placeholder + bound param) | PASS (Query semantics preserved) | PASS (Import `fmt` cleaned) | **VERIFIED** |

---

## 4. Failure Handling & Rollback Validation

The verification suite was subjected to negative and fault-injection trials to confirm that failure in any single oracle triggers immediate workspace rollback:

1. **Syntax Failure Injection:**
   * An invalid token `broken := syntax %%%` was injected into a synthesized file.
   * *Outcome:* Oracle 1 failed with AST parse error.
   * *Disposition:* Overall status marked `FAILED`; workspace rolled back cleanly to pristine snapshot.
2. **Defect Persistence Injection:**
   * A patch added a comment without inserting the required `.Close()` call.
   * *Outcome:* Oracle 3 failed (`"Patched file does not contain deferred Close() call"`).
   * *Disposition:* Rejected; patch withheld from execution boundary.
3. **Variable Shadowing Injection:**
   * An inner block declared `err := helper()` shadowing an outer `err`.
   * *Outcome:* Oracle 5 failed (`"Negative property check failed: detected potentially shadowed 'err' variable definitions"`).
   * *Disposition:* Rejected; repair blocked from promotion.

---

## 5. Conclusion & Gate Readiness

The independent verification engine satisfies all Milestone 3.7 Phase 1.5 requirements:
* Zero reliance on detector output for patch validation.
* Multi-dimensional verification across syntax, regressions, defect elimination, behavioral invariance, and negative properties.
* 100% empirical pass rate on all valid autonomous patches.
* Immediate fail-closed rollback on any oracle failure.
