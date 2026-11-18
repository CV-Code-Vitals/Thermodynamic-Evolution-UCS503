# Milestone 3.6 — Phase 0: Pipeline Integration Readiness Audit

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `8bea13bb5cbd67f20cf59d2607886fc5a7d1bffe`  
**Branch:** `integration/milestone-3.6-controlled-remediation`  
**Audit Date:** 2026-11-07  
**Authorization:** Integration Readiness Audit & Controlled Repair Readiness  
**Status:** COMPLETE — INTEGRATION GAPS IDENTIFIED  

---

## 1. Executive Summary

An exhaustive architectural audit of the end-to-end pipeline was conducted across all eight subsystem boundaries:
$$\text{Repo Intel} \longrightarrow \text{Semantic Diagnostics} \longrightarrow \text{Finding Reconciliation} \longrightarrow \text{Remediation Planner} \longrightarrow \text{Eligibility Policy} \longrightarrow \text{Isolated Executor} \longrightarrow \text{Independent Validator} \longrightarrow \text{Audit Reporter}$$

The audit established that while individual subsystems are robust in isolation—Milestone 3.5 provides high-precision interprocedural, alias-aware, field-sensitive diagnostics, and the Milestone 3.3 execution subsystem provides clean workspace sandboxing—**critical integration gaps exist at the interfaces between diagnostics, planning, execution, and validation**.

Specifically:
1. **Contract Disconnection:** The semantic diagnostic engine (`UnifiedOrchestrator`) outputs `UnifiedDiagnosticReport` containing `ReconciliationLedger` with `ReconciledFinding`, whereas the remediation planning subsystem (`ReportIngester`, `RemediationPlanner`) exclusively ingests `diagnostic.DiagnosticReport` (schema v1.0) with `ValidatedFinding`. No adapter currently bridges these structures.
2. **Metadata Stripping & False Certainty:** Rich semantic metadata introduced in Milestone 3.5 (multi-hop alias graphs, deferred wrapper invocation contracts, field access paths, and call graph traces) is discarded during reconciliation, reducing semantic findings to primitive file-and-line markers.
3. **Unsafe Concurrency Default:** Concurrency ownership states (`ConcurrencyPossibleEscape`, `ConcurrencyStateGoroutineCaptured`, `ConcurrencyStateChannelSent`) are present in `ResourceEvidence`, but `FindingReconciler` fails to inspect them, assigning `EligibleAutonomous` to confirmed leaks regardless of concurrent escape.
4. **Strategy & Validation Void for Semantic Defect Classes:** The execution engine (`StrategyRegistry`) only implements rule-based handlers for Dockerfile and `go.mod`, AST guard-clause injection for JS/JSX, and secret redaction. It possesses zero remediation strategies for semantic SQL injection, command injection, resource leaks, or unhandled errors. Similarly, `IndependentValidator` only verifies lockfiles, Dockerfiles, and secrets, completely lacking the ability to re-invoke semantic diagnostics to verify defect elimination.

---

## 2. Pipeline Interface Audit

### Interface 1: Repository Intelligence $\longrightarrow$ Semantic Diagnostics
- **Source:** `repo_intel.Orchestrator`, AST parsers, `PackageModel`.
- **Destination:** `semantic.UnifiedOrchestrator`, `SemanticAnalyzer`, `FunctionSummaryEngine`.
- **Audit Findings:**
  - Interface is healthy. `UnifiedOrchestrator` groups Go source files by directory and builds structured `PackageSemanticModel` instances.
  - CFGs, reaching-definition dataflow sets, and bounded call graphs are constructed deterministically with failure isolation (`recover()` guards around each analyzer).
  - Schema compatibility: Native Go AST and token sets are shared without translation loss.

### Interface 2: Semantic Diagnostics $\longrightarrow$ Finding Reconciliation
- **Source:** `SQLAnalyzer`, `CommandAnalyzer`, `ResourceAnalyzer`, `ErrorAnalyzer`.
- **Destination:** `FindingReconciler.ReconcileFindings` (`semantic/reconciliation.go`).
- **Audit Findings:**
  - **Schema Loss:** Analyzers produce typed evidence (`SQLEvidence`, `CommandEvidence`, `ResourceEvidence`, `ErrorEvidence`). The reconciler maps them into `ReconciledFinding`.
  - **Critical Concurrency Gap:** In `reconciliation.go` (line 195), resource leaks with `StateResourceLeakConfirmed` are automatically assigned `RemediationEligibility = EligibleAutonomous`. The `ev.ConcurrencyState` field is completely bypassed. A resource captured in a goroutine (`go func() { f.Close() }()`) or sent over a channel is inappropriately labeled as eligible for autonomous patching!
  - **Evidence Dropping:** Alias sets (`ev.Aliases`), alias hop depth (`ev.AliasHopDepth`), struct field access paths (`ev.AccessPath`), and wrapper invocation contracts (`INVOKE_CONDITIONAL`, `INVOKE_STORED`) are omitted from `ReconciledFinding`.

### Interface 3: Finding Reconciliation $\longrightarrow$ Remediation Planner
- **Source:** `semantic.ReconciliationLedger` (`ReconciledFinding`).
- **Destination:** `planning.ReportIngester`, `planning.RemediationPlanner`.
- **Audit Findings:**
  - **Disconnected Interface:** `ReportIngester.IngestReportFile` expects a JSON file matching `diagnostic.DiagnosticReport` schema version `"1.0"`. `UnifiedOrchestrator` serializes `UnifiedDiagnosticReport`.
  - **Finding Identity Mismatch:** `RemediationPlanner` expects `diagnostic.ValidatedFinding` with fields `FindingID`, `ValidationState`, `CalibratedConfidence`, `RemediationReadiness`, etc. It does not ingest `ReconciledFinding`.
  - **Missing Opportunity Rules:** `OpportunityAnalyzer.EvaluateActionability` maps finding IDs based on substring heuristics (`UNPINNED-BASE`, `ROOT-USER`, `SECRET`, `CYCLE`, `HIGH-NESTING`, `NOTEST`). It has no classification rules for `SEM-SEC-SQL`, `SEM-SEC-CMD`, `SEM-REL-RES`, or `SEM-REL-ERR`. Unrecognized semantic findings default to `Informational` (non-actionable) or fallback behavior!

### Interface 4: Remediation Planner $\longrightarrow$ Eligibility Policy
- **Source:** `planning.RemediationTask` (generated by `TaskFactory`).
- **Destination:** `execution.CheckTaskEligibility` (`execution/eligibility.go`).
- **Audit Findings:**
  - Policy logic is sound: tasks with `HumanApprovalRequired: true` or `ExecutionClass: ClassHumanApprovalRequired` fail closed unless an explicit approval token is provided in `approvedIDs`.
  - **Upstream Propagation Vulnerability:** Because `OpportunityAnalyzer` and `StrategySelector` lack semantic-specific rules, high-risk semantic vulnerabilities (e.g. cross-function SQL query refactoring or concurrent resource fixes) might not be marked `HumanApprovalRequired: true` if upstream confidence is artificially high ($\ge 0.60$).
  - Fail-closed requirement: Unrecognized execution classes are rejected with `EXEC_REJECTED`.

### Interface 5: Eligibility Policy $\longrightarrow$ Isolated Executor
- **Source:** `execution.Executor.ExecuteTask`.
- **Destination:** `execution.WorkspaceManager`, `execution.StrategyRegistry`.
- **Audit Findings:**
  - Workspace sandboxing is robust: `WorkspaceManager` creates isolated ephemeral workspaces under temporary directories, copies source trees, computes pre-execution git SHA hashes, and verifies source immutability of the repository root.
  - **Strategy Deficit:** `StrategyRegistry` contains five strategies:
    1. `RuleBasedStrategy`: Only handles `go.sum` and Dockerfile `USER`.
    2. `ASTTransformationStrategy`: Only handles JS/JSX guard clauses and trivial Go `len(a.findings) == 0`.
    3. `HumanGuidedStrategy`: Redacts `AKIA...` hardcoded secrets.
    4. `LLMAssistedStrategy`: Placeholder returning unchanged target files.
    5. `InvestigationOnlyStrategy`: Read-only.
  - **Zero semantic repair implementations:** There are no strategies capable of parameterizing a dynamic SQL query, rewriting an `exec.CommandContext` slice, wrapping an error return in an `if err != nil` block, or inserting a deferred `Close()` call.

### Interface 6: Isolated Executor $\longrightarrow$ Independent Validator
- **Source:** Modified workspace files and generated `UnifiedPatch`.
- **Destination:** `execution.IndependentValidator.Validate` (`execution/validator.go`).
- **Audit Findings:**
  - `IndependentValidator` checks four stages: Defect Resolution, Build/Syntax (`go vet`), Tests (`go test`), and Regression checks.
  - **False Verification Risk:** In `verifyDefectResolution`, the validator only inspects lockfiles, secrets, and Dockerfiles. For any other finding category (including ALL semantic findings), `verifyDefectResolution` defaults to `res.DefectEliminated = true`!
  - If a patch merely touches a Go file and `go vet` compiles, the validator declares `ValPassed` even if the vulnerability was not repaired or was replaced by worse code!
  - **Mandate Violation:** An improved build check or lack of compilation failure is NOT evidence that the semantic defect was resolved. Semantic re-analysis is strictly required.

### Interface 7: Independent Validator $\longrightarrow$ Audit Reporter
- **Source:** `TaskExecutionResult`, `ValidationReport`.
- **Destination:** `execution.ExecutionReporter.WriteExecutionReports`.
- **Audit Findings:**
  - Generates machine-readable `execution_summary.json` and Markdown `execution_report.md`.
  - Accurately tracks rollback occurrences, duration, workspace paths, and review requirements.
  - Lacks structured provenance linking the patch SHA256 to the upstream diagnostic finding's semantic evidence hash.

### Interface 8: Cross-Cutting Integrity & Safety Invariants
- **Stale-Source Detection:** `ReportIngester` compares `report.Repository.HeadCommit` to `currentCommit`. However, in uncommitted working trees or multi-stage pipelines where files are edited, commit hashes alone do not detect stale file contents. File content SHA256 hashing is required.
- **Stale-Summary Detection:** If a callee function is patched, caller function summaries in `FunctionSummaryEngine` must be invalidated; otherwise, subsequent validation runs will evaluate against stale summary contracts.
- **Rollback & Quarantine:** `RollbackManager` cleanly disposes of failed workspaces and reverts changes. Quarantined findings in `reconciliation.go` (`StateResourceUnresolved`) are correctly omitted from findings, but should be explicitly reported in the audit trail.

---

## 3. Integration Gap Matrix

| Gap ID | Severity | Root Cause | Affected Components | Proposed Remediation |
| :---: | :---: | :--- | :--- | :--- |
| **GAP-01** | **CRITICAL** | **Schema Disconnection:** `UnifiedDiagnosticReport` / `ReconciledFinding` cannot be ingested by `ReportIngester` / `RemediationPlanner`. | `semantic/reconciliation.go`, `planning/ingest.go`, `planning/planner.go` | Implement a bidirectional, versioned adapter (`SemanticReportAdapter`) converting `ReconciliationLedger` into versioned `DiagnosticReport` (schema v2.0) with rich semantic finding types. |
| **GAP-02** | **CRITICAL** | **Concurrency Fail-Open:** `FindingReconciler` marks confirmed resource leaks as `EligibleAutonomous` without inspecting `ev.ConcurrencyState`. | `semantic/reconciliation.go`, `semantic/resource_analysis.go` | Enforce invariant: if `ev.ConcurrencyState` is `POSSIBLE_ESCAPE`, `GOROUTINE_CAPTURED`, or `CHANNEL_SENT`, force `RemediationEligibility = EligibleHumanApproval` or `QUARANTINED`, strictly prohibiting autonomous remediation. |
| **GAP-03** | **HIGH** | **Semantic Metadata Stripping:** Multi-hop alias sets, access paths, wrapper invocation contracts, and call graph traces are dropped during reconciliation. | `semantic/reconciliation.go`, `diagnostic/types.go`, `planning/types.go` | Expand finding schema (`SemanticFindingEvidence`) to preserve `AliasChain`, `AccessPath`, `WrapperContract`, and `CallTrace` in `ValidatedFinding` and `RemediationTask`. |
| **GAP-04** | **HIGH** | **Planner Opportunity & Strategy Blindness:** `OpportunityAnalyzer` and `StrategySelector` only recognize legacy heuristic IDs. | `planning/opportunity.go`, `planning/strategy.go`, `planning/task.go` | Add explicit semantic rules for `SEM-SEC-SQL`, `SEM-SEC-CMD`, `SEM-REL-RES`, and `SEM-REL-ERR`, assigning conservative strategies, classes, and approval flags. |
| **GAP-05** | **CRITICAL** | **Missing Semantic Repair Strategies:** `StrategyRegistry` lacks patch generation strategies for semantic vulnerabilities. | `execution/strategies.go`, `execution/executor.go` | Implement deterministic AST / template-based semantic repair strategies for SQL parameterization, command arg splitting, error handling injection, and deferred resource close. |
| **GAP-06** | **CRITICAL** | **Independent Validator Blindness:** `IndependentValidator` does not verify semantic defect resolution; defaults to `DefectEliminated = true`. | `execution/validator.go` | Integrate `UnifiedOrchestrator` / semantic analyzer re-scan inside `verifyDefectResolution` to confirm the targeted finding is absent post-patch. |
| **GAP-07** | **MEDIUM** | **Stale-Source & File Hash Invalidation:** Report provenance only checks repository commit SHA, not individual target file hashes. | `planning/ingest.go`, `execution/workspace.go` | Include per-file SHA256 hashes in `EvidenceLocation` and verify target file content before patch generation. Fail closed if file has diverged. |

---

## 4. Remediation Architecture & Readiness Assessment

To achieve **Controlled Repair Readiness (Milestone 3.6 Phase 1)**, the implementation must:
1. **Contract Standard:** Establish `SemanticFindingContract` (Task A) bridging semantic diagnostics and planning with 100% provenance and uncertainty preservation.
2. **Conservative Planner Integration:** Update `OpportunityAnalyzer`, `StrategySelector`, and `TaskFactory` (Task B) so that semantic findings are routed safely with human approval gates on complex/concurrent categories.
3. **Fixture-Level Repair Trials:** Author isolated, disposable test fixtures and semantic remediation strategies (Task C) covering SQL injection, command injection, unhandled error, resource leak, cross-function ownership, deferred wrappers, and struct fields.
4. **Independent Semantic Verification:** Upgrade `IndependentValidator` (Task E) to re-evaluate patched workspaces with semantic analyzers, ensuring a repair is marked `VERIFIED` only if the defect is mathematically eliminated, compiler checks pass, and tests pass.
5. **Adversarial Safety Harness:** Execute adversarial stress tests (Task D) proving fail-closed behavior on stale findings, ambiguous ownership, validator failures, and interrupted execution.

**Gate Status:** **PROCEED TO PHASE 1 (CONTROLLED INTEGRATION)**
