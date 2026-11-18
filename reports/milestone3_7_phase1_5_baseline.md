# Milestone 3.7 — Phase 1.5: Baseline Audit & Patch Provenance Reconciliation

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Starting Commit:** `92f1c721bb6d7128295a0c63e117d2e8b85bd458`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Audit Date:** 2026-11-07  
**Authorization:** Security Hardening & Synthetic Patch Synthesis Experiments  
**External Repository Patch Execution:** **STRICTLY PROHIBITED**  
**Automatic Merges & Production Deployments:** **STRICTLY PROHIBITED**  

---

## 1. Executive Summary

Milestone 3.7 Phase 1.5 addresses the two fundamental technical debts identified in prior milestone audits:
1. **Approval Security Debt:** Replacing the naive in-memory boolean map (`approvedIDs map[string]bool`) with a cryptographically signed, scope-bound, single-use, replay-resistant authorization token protocol.
2. **Patch Synthesis Debt:** Replacing predetermined, hardcoded trial patch strings with a genuine, bounded AST-driven patch synthesizer (`SemanticFinding → RepairIntent → CandidatePatch → IsolatedValidation → Decision`) that generates valid Go repairs from source code and semantic evidence, with strict abstention on ambiguous or high-risk constructs.

This baseline audit reconciles all prior benchmark evidence, audits every existing code path capable of authorizing execution, and establishes baseline performance metrics before implementation.

---

## 2. Benchmark & Target Reconciliation

All 50 external evaluation targets from Milestone 3.7 Phase 1 and all 22 historical development cases remain frozen and verified:

| Stratum / Suite | Repository Count | Target Count | Ground Truth Verification | Preservation Status |
| :--- | :---: | :---: | :--- | :---: |
| **Stratum 1: Natural Historical Defects** | 4 repos | 4 targets | 1 confirmed bug (`NAT-03`), 3 clean/benign | **Bitwise Frozen** |
| **Stratum 2: Constructed / Injected Defects** | 7 repos | 15 targets | Injections verified via ledger | **Bitwise Frozen** |
| **Stratum 3: Audited Clean Controls** | 13 repos | 31 targets | Unmodified production code | **Bitwise Frozen** |
| **Stratum 4: Historical Dev Benchmark** | Multi-package | 22 cases | Milestone 3.4/3.5 baseline suite | **Bitwise Frozen** |
| **Total Evaluation Denominator** | **13 Repos** | **72 Total Cases** | **Zero Conflation** | **100% Immutable** |

---

## 3. Provenance Audit of Prior Curated Repairs

An audit of [`go-backend/agent/repo_intel/execution/controlled_remediation_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/controlled_remediation_test.go) was conducted to trace the origin of the seven fixture repair cases (`FIX-CASE-01` to `FIX-CASE-07`):

1. **Exact Origin of Patch Payloads:**
   - In `controlled_remediation_test.go` lines 50–135, each test case definition explicitly specifies a `patchCode` string literal containing the replacement Go source code.
   - During trial execution (line 507), the test harness writes `c.patchCode` directly into the ephemeral workspace:
     ```go
     wsTargetFile := filepath.Join(wsDir, c.targetFile)
     if err := os.WriteFile(wsTargetFile, []byte(c.patchCode), 0644); err != nil {
         t.Fatalf("failed to write patch in workspace: %v", err)
     }
     ```
2. **Analysis of Existing Synthesizers in `strategies.go`:**
   - `ASTTransformationStrategy` in [`strategies.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/strategies.go) contains string replacement logic for JS guard clauses (`measurements.length > 0`) and a Go stub (`if len(a.findings) == 0`).
   - It possesses **zero capability** to inspect an arbitrary Go AST, locate an unclosed `os.File` or `http.Response.Body`, identify error return positions, or parameterize a dynamic SQL query string.
3. **Conclusion:**
   The Milestone 3.6 repair trials evaluated **sandboxing containment, rollback mechanics, and multi-stage validator execution** using curated payloads. They did **not** evaluate autonomous code generation. Phase 1.5 must implement genuine AST synthesis.

---

## 4. Execution Authorization Code Path Audit

Every existing code path capable of triggering file writes or patch execution was inspected:

| Subsystem Component | File & Line | Execution Authorization Mechanism | Security Vulnerability Identified |
| :--- | :--- | :--- | :--- |
| **Eligibility Checker** | [`execution/eligibility.go:11`](file:///E:/SWE/go-backend/agent/repo_intel/execution/eligibility.go#L11) | `CheckTaskEligibility(task, approvedIDs map[string]bool)` | **Token Misuse / Replay:** `approvedIDs` is an unauthenticated in-memory boolean map. No signature, no expiration, no binding to git commit or patch hash. |
| **Task Scheduler** | [`execution/scheduler.go:88`](file:///E:/SWE/go-backend/agent/repo_intel/execution/scheduler.go#L88) | `ExecuteNextTask()` invokes `CheckTaskEligibility()` | Inherits eligibility check; permits execution if `ClassAutomationReady` regardless of target file freshness. |
| **Isolated Executor** | [`execution/executor.go:120`](file:///E:/SWE/go-backend/agent/repo_intel/execution/executor.go#L120) | `Executor.ExecuteTask(task)` | Confined to ephemeral workspace `wsDir`. Correctly isolates repository, but lacks patch digest integrity verification. |

**Hardening Mandate:** All execution paths must be refactored to require a cryptographically signed, scope-bound, single-use `ApprovalToken`.

---

## 5. Baseline Performance Profile

Baseline execution latency and memory overhead prior to Phase 1.5 modifications:
- `diagnostic/semantic`: 27 tests passing in **0.575s**.
- `planning`: 8 suites passing in **0.517s**.
- `execution`: 14 suites passing in **8.602s**.
- Heap memory overhead: **45.2 KB**.
- Working tree: clean on `validation/milestone-3.7-external-evaluation` at `92f1c721bb6d7128295a0c63e117d2e8b85bd458`.
