# Milestone 3.3 — Phase 8: Comprehensive Known Limitations Document

**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Branch:** `validation/milestone-3.3-external-validation`  
**Baseline Commit:** `5d41b531f46d10bc972b2d1518a0fb4cf139c09a`  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Milestone 3.3 conducted the platform's first rigorous evaluation against unseen external repositories. While the platform demonstrated extraordinary execution safety (0% FPR, 100% precision, zero regressions), the trials surfaced fundamental technical limitations in the diagnostic engine that prevent an unconditional production release verdict.

This document formally records all known technical limitations to provide an evidence-based foundation for operational deployment boundaries.

---

## 2. Technical Limitations Taxonomy

### 2.1. Language Scope Boundaries
- **Go (Target Language):** Fully supported across all lifecycle stages (discovery, AST parsing, contextual analysis, counter-evidence, patch generation, isolated execution, `go test` validation).
- **Rust (Restricted Scope):** Restricted to read-only discovery, basic syntax checks, and dependency manifest analysis. Autonomous patch generation and remediation planning for Rust are UNSUPPORTED.
- **C/C++ (Restricted Scope):** Restricted to read-only file discovery. Autonomous remediation is strictly UNSUPPORTED.

### 2.2. AST Pattern Blindness (SQL & Command Injection)
- **Concatenation vs. Formatting:** The AST rules in `SQLContextAnalyzer` and `CommandContextAnalyzer` specifically target `*ast.BinaryExpr` with `token.ADD` (e.g., `query := "SELECT ... " + id`).
- **Limitation:** The engine does not recognize dynamic queries constructed via `fmt.Sprintf(...)`, string builders (`strings.Builder`), or template interpolation (`text/template`).
- **Empirical Impact:** Ground-truth defect `DEF-SQL-01` in `sql_service/repository.go` was undetected because the dynamic query was constructed via `fmt.Sprintf`.

### 2.3. Structural Scope Restrictions (Resource Leaks)
- **Intra-Loop Restriction:** `ResourceContextAnalyzer` is explicitly scoped to inspect resource allocations (`os.Open`, `net.Dial`, `http.Get`, `sql.Open`) inside `for` and `range` loops (`ResourceAllocInLoop`, `DeferInLoop`).
- **Limitation:** Resource handles opened in linear, top-level function bodies without loop structures are invisible to the analyzer.
- **Empirical Impact:** Unclosed HTTP responses (`DEF-HTTP-01`), unclosed database rows (`DEF-SQL-02`), and unclosed TCP sockets (`DEF-QUEUE-01`) outside loops went undetected.

### 2.4. Error Handling Idiom Rigidity
- **Blank Identifier Requirement:** `ReliabilitySecurityAnalyzer` relies on regex pattern `unhandledErrorRegex = regexp.MustCompile(`(?i)^\s*_[,\s]*=\s*[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+\(`)`.
- **Limitation:** The analyzer requires the programmer to explicitly write `_ = call()`. If an error-returning function is invoked as a bare expression statement (`ast.ExprStmt`, e.g., `json.Unmarshal(data, &res)` or `tx.Rollback()`), the rule is completely blind.
- **Empirical Impact:** Three unhandled error defects across the external repositories went undetected due to expression statement omission.

### 2.5. Absence of Inter-Procedural Taint Tracking
- **Intra-File Scope:** The contextual command analyzer traces variable assignments within the same file.
- **Limitation:** It cannot trace incoming function parameters across function boundaries or callers. If a function accepts `userInput string` and passes it to `exec.CommandContext`, the parameter is treated as an ordinary identifier.
- **Empirical Impact:** Command injection vulnerability in `cli_tooling/runner.go` was undetected.

### 2.6. Concurrency & Goroutine Lifecycle Blindness
- **Limitation:** The platform lacks a formal semantic model for Go channels, goroutine termination conditions, and struct-level mutex invariants.
- **Empirical Impact:** Concurrent map reads without read locks (`DEF-CACHE-01`) and infinite goroutine ticker leaks (`DEF-CACHE-02`) are undetected by static AST rules.

---

## 3. Deployment Guidance & Operational Safeguards

To mitigate these known limitations in real development workflows:
1. **Mandatory Human Approval for Security & Concurrency:** Automated remediation must never execute concurrency refactorings or public API signature changes autonomously.
2. **Defect-Class Gating:** Autonomous self-healing must be restricted to verified low-risk patterns (intra-loop resource closures, explicitly discarded error returns, and localized parameterization).
3. **Complementary Tooling:** The platform should be paired with standard linters (`golangci-lint`, `govet`, `errcheck`) to cover expression-statement and linear resource defects until the diagnostic engine incorporates full control-flow graph (CFG) analysis.
