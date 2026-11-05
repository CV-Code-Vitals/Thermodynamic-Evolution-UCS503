# Milestone 3.4 — Phase 6 / Phase 3: Semantic Propagation Validation

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Frozen Detector:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Corpus:** `benchmarks/cross_package_stress_v1`  
**Status:** COMPLETE / EMPIRICALLY GROUNDED

---

## 1. Executive Summary

Empirical execution of the frozen Milestone 3.4 semantic diagnostic engine against `benchmarks/cross_package_stress_v1` confirms the structural boundary of the engine:

1. **Resolution vs. Propagation Distinction:** While `CallSiteResolver` correctly identifies call edges without false exact resolutions (0.00% false-exact rate), **semantic state does not propagate interprocedurally across function boundaries**.
2. **SQL Analysis:** Assembling an untrusted SQL query string in a helper function (`pkg_b.FormatQuery`) and executing it in the caller (`pkg_a.CallerSQLCrossPackage`) results in a **False Negative**; the caller treats the helper's return value as opaque dataflow.
3. **Command Analysis:** In contrast, shell execution wrapper sinks (`exec.CommandContext(ctx, "sh", "-c", cmd)`) flag non-constant dynamic command variables even when generated via helper functions, resulting in **Correct Detection**.
4. **Resource Analysis:**
   - **Caller-to-Callee Cleanup (`CallerResourceSplit`):** When a resource is acquired in the caller and passed to a callee helper (`pkg_b.CloserHelper(f)`) that closes it, the engine reports a **False Positive** leak because the intraprocedural CFG does not see the `Close()` call locally.
   - **Callee-to-Caller Acquisition (`CallerResourceTransfer`):** When a callee helper acquires a resource (`pkg_b.AcquireFile(path)`) and returns it to the caller who fails to close it, the engine reports a **False Negative** because the acquisition site is outside the caller's function body.
5. **Error Analysis:** Cross-package function signatures (`error` return types) are recognized when type definitions are present, but error value wrapping and propagation are not traced across call chains.

---

## 2. Domain-by-Domain Semantic Propagation Matrix

| Domain | Test Scenario | Code Location | Observed Diagnostic Result | Ground-Truth Classification | Semantic Propagation Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **SQL** | Untrusted query assembled in callee helper, executed in caller | `pkg_a/caller.go:42` | **Not Detected** | Defect (SQL Injection) | **FALSE NEGATIVE** (`UNSUPPORTED_INTERPROCEDURAL_TAINT`) |
| **SQL** | Parameterized query in helper, executed in caller | `pkg_a/helper_local.go:16` | **Not Detected** | Clean Control | **CORRECT SAFE CLASSIFICATION** |
| **Command** | Shell script string assembled in callee helper, executed in caller | `pkg_a/caller.go:49` | **Detected** (`SEMANTIC_COMMAND_INJECTION`) | Defect (Command Injection) | **CORRECT DETECTION** (due to shell sink heuristic) |
| **Command** | Static command array executed safely | `pkg_a/helper_local.go:22` | **Not Detected** | Clean Control | **CORRECT SAFE CLASSIFICATION** |
| **Resource** | Caller acquires resource, callee helper closes it | `pkg_a/caller.go:68` | **Detected** (`SEMANTIC_RESOURCE_LEAK`) | Clean Control (Closed in helper) | **FALSE POSITIVE** (`UNSUPPORTED_CALLEE_CLEANUP`) |
| **Resource** | Callee helper acquires resource, caller leaks it | `pkg_a/caller.go:58` | **Not Detected** | Defect (Unclosed Resource) | **FALSE NEGATIVE** (`UNSUPPORTED_OWNERSHIP_TRANSFER`) |
| **Resource** | Caller acquires resource, closes with `defer` | `pkg_a/helper_local.go:28` | **Not Detected** | Clean Control | **CORRECT SAFE CLASSIFICATION** |
| **Error** | Callee helper returns wrapped error, caller ignores it | `pkg_a/caller.go:76` | **Not Detected** (no error assign) | Code Smell / Ignored Error | **UNSUPPORTED CAPABILITY** |
| **Error** | Callee helper returns named error, caller handles it | `pkg_a/caller.go:82` | **Not Detected** | Clean Control | **CORRECT SAFE CLASSIFICATION** |

---

## 3. Detailed Failure Mode Analysis

### 3.1 Interprocedural Taint Blindness in SQL Analysis
- **Code:**
  ```go
  func CallerSQLCrossPackage(ctx context.Context, db *sql.DB, user string) error {
      query := pkg_b.FormatQuery("users", "name", user)
      _, err := db.QueryContext(ctx, query)
      return err
  }
  ```
- **Engine Trace:**
  `df.ResolveRootExpression` in `sql_analysis.go` inspects `query`. It identifies the defining assignment `query := pkg_b.FormatQuery(...)`. The RHS is an `*ast.CallExpr`. Because `pkg_b.FormatQuery` is not one of the hardcoded formatting sinks (`fmt.Sprintf`, `strings.Join`, string concatenation `+`), the expression resolver terminates taint traversal and assumes `query` is an opaque return value.
- **Consequence:** Defects constructed through query builder helpers, sanitization wrappers, or repository abstractions are invisible to the intraprocedural SQL engine.

### 3.2 Interprocedural Ownership Blindness in Resource Analysis
- **Code (False Positive):**
  ```go
  func CallerResourceSplit(path string) error {
      f, err := os.Open(path)
      if err != nil {
          return err
      }
      return pkg_b.CloserHelper(f) // f is closed here!
  }
  ```
- **Engine Trace:**
  `ResourceAnalyzer` traces `f` acquired at `os.Open`. It inspects all statements in `CallerResourceSplit`'s CFG. It finds no local `f.Close()` call and no `defer`. Although `pkg_b.CloserHelper(f)` is invoked, the analyzer does not have an interprocedural callee summary indicating that `CloserHelper` consumes and closes its argument. Consequently, the engine reports a high-confidence resource leak on line 68.
- **Code (False Negative):**
  ```go
  func CallerResourceTransfer(path string) error {
      f, err := pkg_b.AcquireFile(path) // Acquisition in callee!
      if err != nil {
          return err
      }
      _ = f.Name()
      return nil // Leaked!
  }
  ```
- **Engine Trace:**
  The analyzer scans `CallerResourceTransfer` for resource acquisition calls matching the registry (`os.Open`, `net.Dial`, etc.). Because `pkg_b.AcquireFile` is not in the hardcoded acquisition table, `f` is never registered as a tracked resource handle.

---

## 4. Scientific Conclusion

Semantic propagation across call boundaries is **UNSUPPORTED** in Milestone 3.4 for dataflow and lifecycle tracking.

The engine operates as a suite of **high-precision intraprocedural AST and CFG analyzers**, augmented with an AST-level call-graph representation. Claims of interprocedural taint propagation or cross-function resource lifecycle analysis cannot be scientifically supported.
