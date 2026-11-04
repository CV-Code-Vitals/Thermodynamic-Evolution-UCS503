# Milestone 3.4 — Phase 2: Expression-Aware SQL & Command Execution Analysis Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Active Branch:** `research/milestone-3.4-semantic-analysis`  
**Base Commit:** `41c3602` (Phase 1 Commit)  
**Evaluation Target:** Semantic SQL and Command Execution Diagnostic Expansion  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Milestone 3.4 Phase 2 addresses the critical syntactic rigidity and dataflow blind spots identified in Milestone 3.3. Specifically, the previous engine missed SQL injections formatted via `fmt.Sprintf` (`DEF-SQL-01` in `sql_service/repository.go`) and command injections executed via shell interpreters (`DEF-CLI-01` in `cli_tooling/runner.go`) because it relied on line-level regexes and AST binary operator addition checks (`binExpr.Op == token.ADD`).

Phase 2 replaces these rigid heuristics with dedicated expression-aware semantic analyzers implemented in:
- [`go-backend/agent/repo_intel/diagnostic/semantic/sql_analysis.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/sql_analysis.go)
- [`go-backend/agent/repo_intel/diagnostic/semantic/command_analysis.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/command_analysis.go)

Both analyzers integrate directly with the Phase 1 Semantic IR, symbol tables, Control Flow Graphs (CFGs), and statement-level reaching definitions dataflow framework.

### Key Phase 2 Accomplishments:
1. **`fmt.Sprintf` Blind Spot Eliminated:** Dynamic SQL query formatting via `fmt.Sprintf` is now tracked from definition to sink, resolving untrusted arguments with reaching definitions.
2. **Safe Parameterization Certified:** Queries with bind placeholders (`?`, `$1`, `:param`) and separate execution arguments are certified as `SAFE_PARAMETERIZED_SQL` with explicit counter-evidence, preventing false-positive alerts.
3. **Shell Interpretation Distinguished from Direct Execution:** Direct process argument separation (`exec.Command("git", "checkout", branch)`) is certified as `SAFE_ARGUMENT_SEPARATION`, whereas shell invocations (`sh -c`, `cmd.exe /C`) inspect the shell command string's provenance and flag untrusted inputs as `UNTRUSTED_SHELL_COMMAND`.
4. **100% Precision & Recall on Controlled Fixtures:** Evaluated on 11 SQL fixtures (5 positive, 6 negative) and 11 Command fixtures (5 positive, 6 negative), achieving 0 False Positives and 0 False Negatives.
5. **Real-World Defect Detection Confirmed:** Both `DEF-SQL-01` (`sql_service/repository.go:54`) and `DEF-CLI-01` (`cli_tooling/runner.go:29`) were detected and verified with 0.97–0.98 calibrated confidence and complete untrusted input provenance traces.

---

## 2. Phase 2A: SQL Expression Analyzer

### 2.1 Supported Sinks
The SQL analyzer inspects standard `database/sql` execution APIs:
- `Query`, `QueryContext`
- `QueryRow`, `QueryRowContext`
- `Exec`, `ExecContext`
- `Prepare`, `PrepareContext`
- Prepared statement execution methods (`(*sql.Stmt).Query`, `(*sql.Stmt).Exec`, `(*sql.Stmt).QueryRow`)

### 2.2 Construction Patterns Analyzed
1. **Direct String Literals:** Checked for SQL statement keywords (`SELECT`, `INSERT INTO`, `UPDATE`, `DELETE FROM`, etc.) and bind placeholders (`?`, `$1`, `:name`, `@p1`).
2. **Binary String Concatenation (`ast.BinaryExpr` with `+`):** Traverses the concatenation operand tree, decomposes leaves into literals, variables, and function calls, and traces whether untrusted inputs reach dynamic operands.
3. **`fmt.Sprintf` Formatting:** Inspects format string for SQL keywords, verifies verb positions (`%s`, `%v`), and traces format arguments through reaching definitions.
4. **Variable Propagation via Reaching Definitions:** When a variable identifier is passed to a SQL sink, the analyzer resolves the root definition statement within the CFG block or preceding blocks.
5. **Function Parameter Ingestion:** Direct passing of a function parameter as a SQL statement string without parameterization is identified and classified.

### 2.3 Six-State Classification Schema

| Classification | Meaning | Confidence | Lifecycle State | Action / Evidence |
|---|---|:---:|:---:|---|
| `STATIC_SQL` | Constant string literal with no dynamic interpolation | 0.99 | `COUNTER_EVIDENCE_PRESENT` | Emits counter-evidence `SAFE_SQL_PARAMETERIZATION` |
| `SAFE_PARAMETERIZED_SQL` | Query uses bind placeholders (`?`, `$1`) with separate execution args | 0.98 | `COUNTER_EVIDENCE_PRESENT` | Emits counter-evidence `SAFE_SQL_PARAMETERIZATION` |
| `DYNAMIC_SQL_STRUCTURE` | Query structure assembled dynamically from internal safe strings | 0.85–0.88 | `PROBABLE` | Requires human review |
| `UNTRUSTED_SQL_STRUCTURE` | Query structure incorporates untrusted parameters or external inputs | 0.96–0.97 | `CONFIRMED` | Flags `SeverityCritical` injection finding |
| `UNRESOLVED_SQL_CONSTRUCTION` | Query is constructed via unresolvable reflection or opaque helpers | 0.50–0.60 | `INSUFFICIENT_EVIDENCE` | Fail-closed; records analysis limitations |
| `NON_SQL_EXPRESSION` | Evaluated expression contains no SQL keywords | 0.85–0.90 | `COUNTER_EVIDENCE_PRESENT` | Suppressed |

---

## 3. Phase 2B: Command Execution Analyzer

### 3.1 Supported Sinks
- `os/exec.Command(binary, args...)`
- `os/exec.CommandContext(ctx, binary, args...)`
- Recognized wrapper functions

### 3.2 Distinguishing Direct Argument Separation from Shell Interpretation

The analyzer implements the critical architectural distinction between direct OS-level process execution and shell interpretation:

```
                          exec.Command(bin, args...)
                                      |
                 +--------------------+--------------------+
                 |                                         |
                 v                                         v
         Known Shell Binary?                       Non-Shell Binary?
    (sh, bash, cmd.exe, pwsh)                  (git, ls, docker, go...)
                 |                                         |
                 v                                         v
       Has Command Flag? (-c, /C)                 Arguments passed as
                 |                                separate argv elements?
        +--------+--------+                                |
        |                 |                                v
       YES                NO                     SAFE_ARGUMENT_SEPARATION
        |                 |                     (No shell word-splitting
        v                 v                     or injection possible)
 Inspect Shell Command   SAFE_ARGUMENT_SEP
 String via Reaching Defs
        |
        v
 Is string dynamic & untrusted?
        |
    +---+---+
    |       |
   YES      NO
    |       |
    v       v
UNTRUSTED  STATIC_COMMAND /
SHELL_CMD  SHELL_INTERPRETED
```

### 3.3 Six-State Classification Schema

| Classification | Meaning | Confidence | Lifecycle State |
|---|---|:---:|:---:|
| `STATIC_COMMAND` | Constant binary and static string arguments | 0.99 | `COUNTER_EVIDENCE_PRESENT` |
| `SAFE_ARGUMENT_SEPARATION` | Arguments passed as discrete array elements to non-shell binary | 0.95 | `COUNTER_EVIDENCE_PRESENT` |
| `DYNAMIC_ARGUMENTS` | Non-shell executable path or arguments derived from parameters | 0.85–0.90 | `PROBABLE` |
| `SHELL_INTERPRETED_COMMAND` | Shell interpreter invoked with static command string | 0.90–0.95 | `COUNTER_EVIDENCE_PRESENT` |
| `UNTRUSTED_SHELL_COMMAND` | Shell interpreter receives untrusted dynamic command string | 0.98 | `CONFIRMED` |
| `UNRESOLVED_COMMAND_CONSTRUCTION` | Invocation uses unsupported indirection or reflection | 0.50–0.60 | `INSUFFICIENT_EVIDENCE` |

---

## 4. Phase 2C: Dataflow & Reaching Definitions Integration

The analyzers leverage the Phase 1 intraprocedural reaching-definitions framework:
1. **Intra-Block & Cross-Block Def-Use Resolution:**
   - Definitions created within the same basic block prior to the call site are tracked via statement position ordering (`pos < callPos`).
   - Definitions reaching across basic blocks are resolved via the worklist fixed-point solution ($In[B]$).
2. **Recursive Variable Unwrapping:**
   - The method `rda.ResolveRootExpression(expr, blockID, pos)` recursively unwraps local variable aliases (e.g. `raw := fmt.Sprintf(...); statement := raw; return db.Exec(statement)`).
3. **Control-Flow Sensitivity:**
   - Mutually exclusive definitions from separate `if/else` branches maintain distinct dataflow states, preventing invalid cross-branch contamination.
4. **Boundary Qualification:**
   - Analysis is strictly intraprocedural in Phase 2. Function parameters are classified as `param:<name> (<type>)` untrusted inputs. Interprocedural call-graph propagation is reserved for Phase 5.

---

## 5. Phase 2E: Integration with Diagnostic Pipeline & Safety Rules

Both analyzers are integrated into `go-backend/agent/repo_intel/diagnostic/context/`:
- `SQLContextAnalyzer.EvaluateFile` invokes `semantic.SQLAnalyzer`, emitting `ContextFinding` with `StateConfirmed` for untrusted queries and `CounterEvidence` (`SAFE_SQL_PARAMETERIZATION`) for safe queries.
- `CommandContextAnalyzer.EvaluateFile` invokes `semantic.CommandAnalyzer`, emitting `ContextFinding` with `StateConfirmed` for untrusted shell commands and `CounterEvidence` (`SAFE_STATIC_EXEC_ARGUMENTS`) for safe separated arguments.

### Safety Invariants Enforced:
1. **Fail-Closed Gatekeeping:** Any finding classified with `StateInsufficientEvidence` or `StateHumanInvestigationReq` is strictly prevented from entering automated remediation planning.
2. **High-Risk Approval Policies:** Confirmed security-sensitive findings (`SeverityCritical`, `SeverityHigh`) remain classified as `APPROVAL_REQUIRED` or blocked from autonomous modification.
3. **Zero Autonomous Remediation Introduced:** Phase 2 introduces zero code modifications or patches to analyzed repositories.

---

## 6. Performance & Failure Handling

Benchmark measurements from `resilience_perf_test.go`:
- **SQL Fixture File Parse & CFG Build:** $517.4\ \mu\text{s}$ ($< 0.6\ \text{ms}$)
- **SQL Function Semantic Analysis (11 functions):** $531.1\ \mu\text{s}$ ($< 0.6\ \text{ms}$)
- **External Package Analysis (`http_router`, 5 files):** $3.65\ \text{ms}$
- **Resilience Testing:**
  - Syntax errors / truncated files: gracefully caught and reported without panic.
  - Missing third-party packages: graceful type-checking degradation; full AST and CFG analysis succeeds.
  - Opaque reflection queries: fail-closed to `UNRESOLVED_SQL_CONSTRUCTION` with `INSUFFICIENT_EVIDENCE` and explicit analysis limitations recorded.

---

## 7. Phase 2 Gate Review Status

| Acceptance Criteria | Target | Observed Result | Status |
|---|---|:---:|:---:|
| 1. `fmt.Sprintf` SQL blind spot addressed | Detect `DEF-SQL-01` | **DETECTED** (`Confidence 0.97`) | **PASSED** |
| 2. Safe parameterized SQL not flagged | Zero false positives | **0 FP** on parameterized fixtures & controls | **PASSED** |
| 3. Dynamic SQL structure distinguished | Differentiate structure vs value | **PROBABLE** vs **CONFIRMED** | **PASSED** |
| 4. Shell interpretation distinguished | Safe argv separation certified | Certified as `SAFE_ARGUMENT_SEPARATION` | **PASSED** |
| 5. Untrusted input traceable | Source-to-sink evidence recorded | Recorded in `untrusted_inputs` & `reaching_defs` | **PASSED** |
| 6. Unresolved dataflow fail-closed | `INSUFFICIENT_EVIDENCE` emitted | No unsupported high-confidence findings | **PASSED** |
| 7. Positive and negative fixtures pass | 100% pass on 22 test cases | **22 / 22 PASS** | **PASSED** |
| 8. Regression tests pass | Go & Rust suites green | **ALL PASS** (0 regressions) | **PASSED** |
| 9. Historical benchmarks preserved | Frozen SHA-256 intact | **100% UNMODIFIED** | **PASSED** |
| 10. Remediation safety intact | Zero autonomous execution on security defects | **PRESERVED** | **PASSED** |

Phase 2 implementation is complete and verified. Ready for Phase 3 (Function-Level Resource Lifecycle Analysis).
