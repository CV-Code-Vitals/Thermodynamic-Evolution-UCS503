# Milestone 3.4 — Phase 2: Test Results & Confusion Matrix Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Active Branch:** `research/milestone-3.4-semantic-analysis`  
**Base Commit:** `41c3602` (Phase 1 Commit)  
**Target:** Positive & Negative Fixture Verification, Real-Repository Validation, and Confusion Matrices  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Phase 2 constructed isolated positive and negative fixtures for SQL query construction and command execution analysis. Every positive fixture is paired with an equivalent negative control that differs only in the critical security property (e.g. parameterization vs formatting, discrete argv vs shell interpretation).

The evaluation demonstrates:
- **Controlled SQL Fixture Benchmark:** Precision: 100.00%, Recall: 100.00%, F1: 100.00%, Specificity: 100.00%
- **Controlled Command Fixture Benchmark:** Precision: 100.00%, Recall: 100.00%, F1: 100.00%, Specificity: 100.00%
- **External Real-World Repositories:**
  - `sql_service/repository.go`: `DEF-SQL-01` detected (`fmt.Sprintf` dynamic SQL injection, confidence 0.97).
  - `cli_tooling/runner.go`: `DEF-CLI-01` detected (`cmd.exe /C` & `sh -c` shell script injection, confidence 0.98).
- **Regression Safety:** 100% test pass rate across existing Go repository intelligence, execution engine, and Rust AST test suites.

---

## 2. SQL Analysis Fixture Results

Fixtures defined in [`go-backend/agent/repo_intel/diagnostic/semantic/fixtures/sql_fixtures.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/fixtures/sql_fixtures.go):

| Fixture Function | Pattern Tested | Expected Classification | Observed Classification | Confidence | Lifecycle State | Result |
|---|---|:---:|:---:|:---:|:---:|:---:|
| `Pos_SQL_Sprintf` | `fmt.Sprintf` dynamic query | `UNTRUSTED_SQL_STRUCTURE` | `UNTRUSTED_SQL_STRUCTURE` | 0.97 | `CONFIRMED` | **PASS** |
| `Pos_SQL_Concat` | Binary string addition `+` | `UNTRUSTED_SQL_STRUCTURE` | `UNTRUSTED_SQL_STRUCTURE` | 0.96 | `CONFIRMED` | **PASS** |
| `Pos_SQL_PropagatedLocal` | Untrusted input via local var | `UNTRUSTED_SQL_STRUCTURE` | `UNTRUSTED_SQL_STRUCTURE` | 0.97 | `CONFIRMED` | **PASS** |
| `Pos_SQL_ParamAsQuery` | Raw query parameter to sink | `UNTRUSTED_SQL_STRUCTURE` | `UNTRUSTED_SQL_STRUCTURE` | 0.96 | `CONFIRMED` | **PASS** |
| `Pos_SQL_ContextSink` | `fmt.Sprintf` to `QueryContext` | `UNTRUSTED_SQL_STRUCTURE` | `UNTRUSTED_SQL_STRUCTURE` | 0.97 | `CONFIRMED` | **PASS** |
| `Neg_SQL_StaticQuery` | Constant string literal | `STATIC_SQL` | `STATIC_SQL` | 0.99 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_SQL_ParameterizedQuery` | `?` placeholder + separate args | `SAFE_PARAMETERIZED_SQL` | `SAFE_PARAMETERIZED_SQL` | 0.98 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_SQL_ParameterizedContext` | `$1` placeholder + context arg | `SAFE_PARAMETERIZED_SQL` | `SAFE_PARAMETERIZED_SQL` | 0.98 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_SQL_PreparedStatement` | Prepared stmt with placeholder | `SAFE_PARAMETERIZED_SQL` | `SAFE_PARAMETERIZED_SQL` | 0.98 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_SQL_SprintfNonSQL` | `fmt.Sprintf` used for logging | Non-Sink / Suppressed | `NON_SQL_EXPRESSION` | 0.90 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_SQL_DynamicUnused` | Query formatted but unexecuted | Non-Sink / Unused | Zero Sink Calls | — | Clean | **PASS** |

### 2.1 SQL Fixture Confusion Matrix

$$\begin{array}{c|cc}
& \textbf{Vulnerable (Predicted)} & \textbf{Safe / Non-Vulnerable (Predicted)} \\
\hline
\textbf{Vulnerable (Actual)} & \text{TP} = 5 & \text{FN} = 0 \\
\textbf{Safe (Actual)} & \text{FP} = 0 & \text{TN} = 6 \\
\end{array}$$

- **True Positives (TP):** 5
- **False Positives (FP):** 0
- **True Negatives (TN):** 6
- **False Negatives (FN):** 0
- **Precision:** $\frac{5}{5 + 0} = \mathbf{100.00\%}$
- **Recall:** $\frac{5}{5 + 0} = \mathbf{100.00\%}$
- **Specificity:** $\frac{6}{6 + 0} = \mathbf{100.00\%}$
- **F1 Score:** $\mathbf{100.00\%}$

---

## 3. Command Execution Analysis Fixture Results

Fixtures defined in [`go-backend/agent/repo_intel/diagnostic/semantic/fixtures/command_fixtures.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/fixtures/command_fixtures.go):

| Fixture Function | Pattern Tested | Expected Classification | Observed Classification | Confidence | Lifecycle State | Result |
|---|---|:---:|:---:|:---:|:---:|:---:|
| `Pos_Cmd_ShellSprintf` | `sh -c fmt.Sprintf` | `UNTRUSTED_SHELL_COMMAND` | `UNTRUSTED_SHELL_COMMAND` | 0.98 | `CONFIRMED` | **PASS** |
| `Pos_Cmd_ShellConcat` | `bash -c "echo " + input` | `UNTRUSTED_SHELL_COMMAND` | `UNTRUSTED_SHELL_COMMAND` | 0.98 | `CONFIRMED` | **PASS** |
| `Pos_Cmd_ShellPropagated` | `cmd.exe /c` via local variable | `UNTRUSTED_SHELL_COMMAND` | `UNTRUSTED_SHELL_COMMAND` | 0.98 | `CONFIRMED` | **PASS** |
| `Pos_Cmd_ShellParamDirect` | Parameter passed to `sh -c` | `UNTRUSTED_SHELL_COMMAND` | `UNTRUSTED_SHELL_COMMAND` | 0.98 | `CONFIRMED` | **PASS** |
| `Pos_Cmd_ContextShell` | `sh -c` with `CommandContext` | `UNTRUSTED_SHELL_COMMAND` | `UNTRUSTED_SHELL_COMMAND` | 0.98 | `CONFIRMED` | **PASS** |
| `Neg_Cmd_StaticExec` | Fixed binary + static args | `STATIC_COMMAND` | `STATIC_COMMAND` | 0.99 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_Cmd_SafeArgSeparation` | Non-shell dynamic argv | `SAFE_ARGUMENT_SEPARATION` | `SAFE_ARGUMENT_SEPARATION` | 0.95 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_Cmd_DirectArgBranch` | Git checkout branch arg | `SAFE_ARGUMENT_SEPARATION` | `SAFE_ARGUMENT_SEPARATION` | 0.95 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_Cmd_StaticShell` | `sh -c` with constant string | `SHELL_INTERPRETED_COMMAND` | `SHELL_INTERPRETED_COMMAND` | 0.95 | `COUNTER_EVIDENCE_PRESENT` | **PASS** |
| `Neg_Cmd_SprintfNonExec` | `fmt.Sprintf` for logging | Non-Sink / Suppressed | Zero Sink Calls | — | Clean | **PASS** |
| `Neg_Cmd_UnusedCommand` | Dynamic command string unused | Non-Sink / Unused | Zero Sink Calls | — | Clean | **PASS** |

### 3.1 Command Fixture Confusion Matrix

$$\begin{array}{c|cc}
& \textbf{Vulnerable (Predicted)} & \textbf{Safe / Non-Vulnerable (Predicted)} \\
\hline
\textbf{Vulnerable (Actual)} & \text{TP} = 5 & \text{FN} = 0 \\
\textbf{Safe (Actual)} & \text{FP} = 0 & \text{TN} = 6 \\
\end{array}$$

- **True Positives (TP):** 5
- **False Positives (FP):** 0
- **True Negatives (TN):** 6
- **False Negatives (FN):** 0
- **Precision:** $\frac{5}{5 + 0} = \mathbf{100.00\%}$
- **Recall:** $\frac{5}{5 + 0} = \mathbf{100.00\%}$
- **Specificity:** $\frac{6}{6 + 0} = \mathbf{100.00\%}$
- **F1 Score:** $\mathbf{100.00\%}$

---

## 4. Real-World External Benchmark Validation

The new analyzers were evaluated against the unseen external repositories from Milestone 3.3:

### 4.1 SQL Service (`benchmarks/external_v1/sql_service/repository.go`)
- **`DEF-SQL-01` (`FindUserByNameRaw:54`):**
  - **Construct:** `query := fmt.Sprintf("SELECT ... WHERE username = '%s'", name)`
  - **Sink:** `r.db.QueryRowContext(ctx, query).Scan(...)`
  - **Detection:** `UNTRUSTED_SQL_STRUCTURE`, `Confidence 0.97`, `LifecycleState: CONFIRMED`, `Severity: CRITICAL`.
  - **Evidence Trace:** Untrusted parameter `param:name (string)` traced from function signature through `reaching_defs` to `fmt.Sprintf` argument.
  - **Verdict:** **DEFECT DETECTED (Blind spot eliminated).**
- **Negative Controls in `repository.go`:**
  - `GetUserByID`: Certified as `SAFE_PARAMETERIZED_SQL` (`?` bind placeholder).
  - `CreateUser`: Certified as `SAFE_PARAMETERIZED_SQL` (`VALUES (?, ?, ?, ?)`).
  - `ListRecentUsers`: Certified as `SAFE_PARAMETERIZED_SQL` (`ORDER BY id DESC LIMIT ?`).
  - **Verdict:** **ZERO FALSE POSITIVES.**

### 4.2 CLI Tooling (`benchmarks/external_v1/cli_tooling/runner.go`)
- **`DEF-CLI-01` (`ExecuteShellScript:29, 31`):**
  - **Construct:** `cmd := exec.CommandContext(ctx, "cmd.exe", "/C", userInput)` & `exec.CommandContext(ctx, "sh", "-c", userInput)`
  - **Detection:** `UNTRUSTED_SHELL_COMMAND`, `Confidence 0.98`, `LifecycleState: CONFIRMED`, `Severity: CRITICAL`.
  - **Evidence Trace:** Untrusted parameter `param:userInput (string)` passed to recognized shell interpreter (`cmd.exe /C`, `sh -c`).
  - **Verdict:** **DEFECT DETECTED (Blind spot eliminated).**
- **Negative Control in `runner.go`:**
  - `ExecuteDirect`: Classified as `DYNAMIC_ARGUMENTS` (`is_shell_interpreter: false`, non-shell direct process execution).
  - **Verdict:** **ZERO FALSE POSITIVES.**

---

## 5. Subsystem Regression Suite Execution Summary

| Subsystem | Package Path | Tests Run | Result | Duration |
|---|---|:---:|:---:|:---:|
| Semantic Diagnostics | `agent/repo_intel/diagnostic/semantic/...` | 12 | **PASS** | 1.02s |
| Diagnostic Core | `agent/repo_intel/diagnostic` | 8 | **PASS** | 0.61s |
| Diagnostic Context | `agent/repo_intel/diagnostic/context` | 1 | **PASS** | 1.08s |
| Diagnostic Analysis | `agent/repo_intel/diagnostic/analysis` | 1 | **PASS** | 0.00s |
| Root Cause Clustering | `agent/repo_intel/diagnostic/rootcause` | 1 | **PASS** | 0.00s |
| Remediation Execution | `agent/repo_intel/execution` | 7 | **PASS** | 1.11s |
| Rust AST Engine | `thermodynamic-ast-engine` | 12 | **PASS** | 0.08s |
| **Total** | | **42** | **PASS (100%)** | **3.90s** |

All tests passed with zero regressions.
