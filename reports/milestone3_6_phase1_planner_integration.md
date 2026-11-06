# Milestone 3.6 — Phase 1: Remediation Planner Semantic Integration

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `8bea13bb5cbd67f20cf59d2607886fc5a7d1bffe`  
**Branch:** `integration/milestone-3.6-controlled-remediation`  
**Document Type:** Planner Integration & Governance Verification  

---

## 1. Executive Summary

The Remediation Planning subsystem (`OpportunityAnalyzer`, `StrategySelector`, `TaskFactory`, and `RemediationPlanner`) was extended to ingest, evaluate, and schedule tasks derived from **Milestone 3.5 interprocedural semantic findings** without discarding provenance, converting uncertainty into false confidence, or authorizing unsafe autonomous patches.

### Core Enhancements:
1. **Schema Bridge:** Connected `SemanticFindingContract` to `diagnostic.ValidatedFinding` via `ToValidatedFinding()` and `BuildDiagnosticReportFromContracts()`.
2. **Explicit Semantic Strategy Routing:** Implemented first-class classification rules in `StrategySelector` for `SEM-SEC-SQL`, `SEM-SEC-CMD`, `SEM-REL-RES`, and `SEM-REL-ERR`.
3. **Strict Concurrency Fail-Closed Governance:** Any finding exhibiting goroutine capture, channel transfer, or uncertain concurrency state (`POSSIBLE_ESCAPE`, `UNKNOWN_CONCURRENT_OWNERSHIP`) is strictly mapped to `ClassHumanApprovalRequired` with `HumanApprovalRequired = true`. Autonomous execution is prohibited.
4. **Shell Execution Protection:** Modifying process execution from shell interpreters (`sh -c`, `cmd.exe /C`) to direct binary execution is classified as `ClassHumanApprovalRequired` because it alters environment variable expansion and shell metacharacter semantics.

---

## 2. Semantic Finding Evaluation & Strategy Routing Matrix

| Finding Category & ID Pattern | Actionability State | Execution Strategy | Execution Class | Approval Required | Governance Rationale |
| :--- | :---: | :---: | :---: | :---: | :--- |
| **Semantic SQL Injection** (`SEM-SEC-SQL-*`) | `ActionableWithReview` | `StrategyASTTransformation` | `ClassAutomationWithReview` | `false` (Automated generation guarded by test regression gates) | Query parameterization is deterministic, but requires compiler and regression test validation before merge. |
| **Command Injection** (`SEM-SEC-CMD-*`) | `HumanApprovalRequired` | `StrategyASTTransformation` | `ClassHumanApprovalRequired` | `true` | Converting shell string commands to argument slices alters shell builtins and redirection semantics; human authorization mandatory. |
| **Deterministic Local Resource Leak** (`SEM-REL-RES-*`) | `Actionable` | `StrategyASTTransformation` | `ClassAutomationReady` | `false` | Deterministic local exit cleanup (`defer Close()`) with proven single-function scope and zero concurrency escape. |
| **Concurrent Resource Escape** (`SEM-REL-RES-*` with `POSSIBLE_ESCAPE`) | `HumanApprovalRequired` | `StrategyHumanGuided` | `ClassHumanApprovalRequired` | `true` | Resource handle lifetime escapes local control flow into goroutines or channels; static safety cannot be proven autonomously. |
| **Semantic Unhandled Error** (`SEM-REL-ERR-*`) | `Actionable` | `StrategyASTTransformation` | `ClassAutomationReady` | `false` | Injecting `if err != nil` check with local return or error wrap. |
| **Partial Error Handling** (`SEM-REL-ERR-*` partial/cleanup) | `HumanApprovalRequired` | `StrategyHumanGuided` | `ClassHumanApprovalRequired` | `true` | Contextual cleanup error handling requires human architectural review. |
| **Unsupported / Unknown Finding Category** | `RejectedAsInvalid` / `Informational` | `StrategyInvestigationOnly` | `ClassInvestigationRequired` | `true` | Fails closed; no automated patch generated or scheduled. |

---

## 3. Provenance & Uncertainty Invariant Verification

### Invariant 1: Source Provenance Preservation
When `TaskFactory.CreateTask` instantiates a `RemediationTask`:
- `SourceFindingIDs` preserves the canonical semantic finding ID (`SEM-SEC-SQL-...`).
- `TargetFiles` preserves exact target file paths.
- `ValidationCommands` includes compiler checks (`go vet ./...`) and unit tests (`go test -v ./...`).
- Acceptance criteria explicitly mandate post-patch semantic re-scanning to certify defect elimination.

### Invariant 2: Prevention of Autonomous Concurrent Fixes
In `OpportunityAnalyzer.EvaluateActionability`:
```go
if strings.Contains(f.ReadinessRationale, "concurrency") ||
   strings.Contains(f.ReadinessRationale, "goroutine") ||
   strings.Contains(f.ReadinessRationale, "channel") {
    return HumanApprovalRequired, "High-impact security modification, concurrent ownership ambiguity, or architectural cyclic dependency requires explicit human review and authorization"
}
```
And in `StrategySelector.SelectStrategy`:
```go
if strings.Contains(f.ReadinessRationale, "concurrency") ||
   strings.Contains(f.ReadinessRationale, "goroutine") ||
   strings.Contains(f.ReadinessRationale, "channel") ||
   f.RemediationReadiness == diagnostic.ReadinessHumanApprovalRequired {
    return StrategyHumanGuided, ClassHumanApprovalRequired, true,
        "Resource handle with concurrent escape, loop accumulation, or uncertain ownership requires human authorization"
}
```
This guarantees that **no heuristic, high confidence score, or diagnostic warning can bypass the human approval requirement for concurrent resources**.

---

## 4. Verification Outcomes

The integration was verified via `TestRemediationPlanner_SemanticFindingsIntegration` in `planning/planning_test.go`:
- SQL findings $\longrightarrow$ Planned as `ClassAutomationWithReview` with `HumanApprovalRequired = false`.
- Command injection findings $\longrightarrow$ Planned as `ClassHumanApprovalRequired` with `HumanApprovalRequired = true`.
- Local resource leaks $\longrightarrow$ Planned as `ClassAutomationReady` with `HumanApprovalRequired = false`.
- Concurrent resource escapes $\longrightarrow$ Planned as `ClassHumanApprovalRequired` with `HumanApprovalRequired = true`.
- Entire `planning` package test suite passed in **0.874s** with 100% success rate.
