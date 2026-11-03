# Milestone 3.0: Diagnostic Intelligence Error Analysis Report
## Structured False-Positive & False-Negative Investigation

**Analysis Timestamp:** `2026-11-03T19:03:05+05:30`  
**Total Errors Investigated:** `11`  
**False Positives Analyzed:** `4`  
**False Negatives Analyzed:** `7`  

## 1. False Positive Error Catalog (Over-Detection)

| Error ID | Target Fixture | Target File & Symbol | Failure Category | Root Cause & Corrective Action |
|---|---|---|---|---|
| `ERR-FP-01` | `CLEAN-STR-01` | `structural/clean_complex.go` (`MapHTTPStatusToCategory`) | Missing tests incorrectly treated as code defect | **Cause:** Directory-level NOTEST hygiene heuristic (RI-QUAL-NOTEST-STRUCTURAL) tagged all files in the package, including benign declarative switch tables.<br>**Fix:** Scope NOTEST findings strictly as repository/package hygiene indicators rather than actionable code defects; exempt clean declarative control tables. |
| `ERR-FP-02` | `CLEAN-REL-01` | `reliability/clean_error.go` (`ReadConfigWithFallback`) | Missing tests incorrectly treated as code defect | **Cause:** Benign error handling logic contaminated by directory-level RI-QUAL-NOTEST-RELIABILITY.<br>**Fix:** Partition diagnostic findings into functional defects vs test coverage recommendations; do not generate code remediation tasks for missing tests on benign files. |
| `ERR-FP-03` | `CLEAN-SEC-01` | `security/clean_sql.go` (`FetchUserSafe`) | Coarse package-level hygiene heuristic leaking into clean security control | **Cause:** Parameterized SQL query correctly ignored by injection regex but contaminated by package-level RI-QUAL-NOTEST-SECURITY.<br>**Fix:** Introduce contextual validation that produces explicit negative evidence (clean certificate) when parameterization is confirmed. |
| `ERR-FP-04` | `CLEAN-SEC-02` | `security/clean_exec.go` (`CheckGitVersion`) | Coarse package-level hygiene heuristic leaking into clean command execution control | **Cause:** Static fixed-arg subprocess invocation correctly passed injection regex but contaminated by RI-QUAL-NOTEST-SECURITY.<br>**Fix:** Context-aware command execution analyzer certifying discrete constant arguments as safe subprocesses. |

## 2. False Negative Error Catalog (Under-Detection / Missed Defects)

| Error ID | Target Fixture | Target File & Symbol | Failure Category | Root Cause & Corrective Action |
|---|---|---|---|---|
| `ERR-FN-01` | `FIX-REL-02` | `reliability/resource_lifecycle.go` (`ReadFilesUnsafely`) | Resource lifecycle defects not recognized | **Cause:** AST engine and heuristic regexes lack loop-boundary resource escape analysis (os.Open called in a loop without close or defer).<br>**Fix:** Implement AST loop resource tracking detector that flags file/connection handles opened in loops without deterministic closure. |
| `ERR-FN-02` | `FIX-DEP-03` | `go.mod` (`go.sum`) | Path normalization and target scope mismatch | **Cause:** The engine detected RI-DEP-GOSUM-MISSING-ROOT but reported target path as 'ROOT' rather than relative file 'go.mod'.<br>**Fix:** Standardize lockfile finding affected components to explicit relative file paths ('go.mod'). |
| `ERR-FN-03` | `FIX-ARC-01` | `architecture/boundary/api_to_db.go` (`HandleUserRequest`) | Incomplete architectural boundary information | **Cause:** Engine lacked layered architecture boundary rules (e.g., presentation/API package importing database/sql directly).<br>**Fix:** Add semantic layer boundary policy engine enforcing that API/handler packages cannot directly import persistence/driver packages. |
| `ERR-FN-04` | `FIX-ARC-02` | `architecture/circular/cycle.go` (`ServiceA`) | Circular dependency not detected at intra-package struct level | **Cause:** Cycle detection operated strictly on Go package import statements (package A importing B), not struct-level type mutual references.<br>**Fix:** Extend architectural cycle analyzer to construct type-level reference graphs and flag tightly coupled circular structs. |
| `ERR-FN-05` | `FIX-TST-01` | `testing/untested_billing.go` (`CalculateInvoiceTax`) | Missing symbol-level test coverage | **Cause:** Quality analyzer evaluated test presence at the directory level. Because testing/ had tested_auth_test.go, the entire package was marked 'tested'.<br>**Fix:** Implement symbol-level test cross-referencing linking exported public functions to test function names and AST calls. |
| `ERR-FN-06` | `FIX-MNT-01` | `maintainability/duplicate_helpers.go` (`NormalizeEmailB`) | Duplicated helper implementation not detected | **Cause:** Structural analyzer had no clone-detection or AST body hashing mechanism across functions.<br>**Fix:** Implement AST body normalization and token hash comparison to detect duplicate function bodies exceeding 10 statements. |
| `ERR-FN-07` | `FIX-DOC-01` | `documentation/undocumented_service/service.go` (`PaymentGatewayClient`) | Missing required public API documentation | **Cause:** Engine lacked doc comment inspection on exported Go types and methods.<br>**Fix:** Add AST doc comment checker flagging exported structs and methods that lack descriptive doc comments. |

## 3. Strategic Action Items for Milestone 3.0 Architecture

- 1. Contextual Evidence Aggregation: Combine AST structure, symbol definitions, loop boundaries, and type call-graphs.
- 2. Symbol-Level Test Cross-Referencing: Replace coarse directory-level NOTEST heuristics with function-level test call verification.
- 3. Layer Boundary & Architectural Policy: Detect presentation layers directly querying databases and struct-level mutual circularities.
- 4. Intra-Loop Resource Escape Analysis: Detect unclosed file and connection handles opened inside loop bodies.
- 5. Code Duplication & Exported Doc Analysis: Add AST token body hashing for clone detection and godoc comment verification.
- 6. Root Cause Consolidation: Group symptomatic findings (nesting + complexity; unhandled error + ignored return) under singular root cause entities.

