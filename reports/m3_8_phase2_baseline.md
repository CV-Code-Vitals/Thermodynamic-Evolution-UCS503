# Milestone 3.8 — Phase 2: Live Repository Scan-to-Repair Baseline & Evidence Freeze

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace Root:** `E:\SWE`  
**Target Branch:** `release/milestone-3.8-mvp-rc`  
**Starting Baseline Commit:** `421b91712b2d10276a410cddda19e21b55e6492f`  
**Clean Working Tree Verified:** Yes (`git status --porcelain` clean)  
**Timestamp:** 2026-11-21T18:00:00Z  

---

## 1. Executive Summary & Objective

Milestone 3.8 Phase 2 qualifies the **Live Repository Scan-to-Repair** pipeline, integrating full-repository diagnostic scanning with controlled, sandboxed AST-level remediation, decoupled independent oracles, cryptographic authorization binding, and 100% source tree immutability.

This baseline freeze establishes the frozen scientific evidence, source manifests, component inventory, and target distributions discovered from the live reference repository `https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503` on pinned release commit `421b91712b2d10276a410cddda19e21b55e6492f`.

---

## 2. Integrated Component Inventory

The Live Repository Scan-to-Repair feature unifies the frontend operator interface, backend service orchestration, semantic diagnostic engines, AST mutator synthesis, independent verification oracles, and ephemeral execution containment.

| Component | File Path | Functional Role in Scan-to-Repair Workflow |
| :--- | :--- | :--- |
| **MVP Scanner & Target Discovery** | [mvp_scanner.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/mvp_scanner.go) | Recursively scans workspace Go files, invokes `semantic.UnifiedOrchestrator`, parses `thermodynamic_report.json`, extracts cyclomatic/entropy hotspots, binds snapshot metadata (`SnapshotID`, `SourceFileHash`). |
| **Autonomous Repair MVP Service** | [mvp_service.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/mvp_service.go) | Central controller managing target listing, semantic analysis, AST patch synthesis, cryptographic HMAC-SHA256 authorization, sandbox execution, and audit logging. |
| **Independent Verification Oracles** | [oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go) | Decoupled 5-stage verification suite: Syntax/Compilation, Regression Tests, Defect Reproducer, Behavioral Invariance, and Negative Properties. |
| **Workspace Store & Retention** | [workspace.go](file:///E:/SWE/go-backend/go-backend/workspace.go) | Retains scanned workspace directories with lifetime timers, allowing seamless transition from Scan & Optimize to Remediation Portal. |
| **Remediation HTTP Endpoints** | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | REST API exposing `/api/remediation/targets`, `/analyze`, `/propose`, `/approve`, `/execute`, `/rollback`, `/audit`, and `/demo`. |
| **Remediation Portal UI** | [RemediationPortal.jsx](file:///E:/SWE/admin-portal/src/components/RemediationPortal.jsx) | Operator console featuring category filtering, action filtering, search, 25-item pagination, target preview, approval sign-off, diff visualization, and audit downloads. |
| **Remediation API Client** | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | Frontend client binding for all remediation endpoints with error handling and parameterization. |
| **End-to-End Test Suite** | [scan_to_repair_e2e_test.go](file:///E:/SWE/go-backend/scan_to_repair_e2e_test.go) | Automated end-to-end integration tests covering live repository workflow, offline fixtures, credential stripping, arbitrary command blocking, and stale workspace rejection. |

---

## 3. Pinned Live Repository Scan Reproduction

The live scan was executed against the official repository `https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503` under retained workspace mode.

### Execution Summary
- **Cloned Source Files:** 673 Go source files (excluding `.git`, `vendor`, `node_modules`, `.tmp`, `uploads`)
- **AST Functions Analyzed:** 22,763 AST hotspot records across the repository
- **Total Discovered Targets:** 1,293 distinct repair/abstain candidates
- **Preconfigured Demo Fixtures:** 5 pilot corpus targets preserved for deterministic evaluation

### Discovered Target Classification & Severity Breakdown

| Defect Category | Discovered Count | Action Disposition | Automatic Eligibility | Rationale & Safety Policy |
| :--- | :--- | :--- | :--- | :--- |
| **High Complexity Hotspots** | 1,250 | `ABSTAIN` | False (Requires Human Review) | Functions with Cyclomatic Complexity $\ge 15$ or Entropy Score $\ge 75$. Method extraction and decomposition require architectural specification. |
| **Command Injection (`sh -c`)** | 42 | `ABSTAIN` | False (Refused by Policy) | Unsanitized command passed to shell interpreter. Automated splitting into discrete arguments requires semantic knowledge; autonomous execution forbidden. |
| **SQL Injection** | 1 | `REPAIR` | True (Automatic AST Synthesis) | Dynamic SQL query formatted via `fmt.Sprintf`. Parameterization with `?` placeholder is deterministic and provably safe. |
| **Resource Leaks (Escaping)** | Varied | `ABSTAIN` | False (Refused by Contract) | Resource allocated intra-procedurally is returned to caller or transferred across functional boundary. Callee cleanup violates interface contract. |
| **Preconfigured Demo Targets** | 5 | 2 REPAIR / 3 ABSTAIN | Controlled | Pilot evaluation targets: `INJ-EXT2-RES-02`, `INJ-EXT2-SQL-01`, `INJ-EXT2-CMD-01`, `DEF-HOLD-01`, and `NAT-03`. |

---

## 4. Preservation of Historical MVP Benchmarks

In accordance with strict scientific integrity principles:
1. **Historical Repair Rate Denominators:** The 9/15 constructed/injected repair metric from Milestone 3.7 Phase 2 and Milestone 3.8 Phase 1.1 remains **unrevised and frozen**.
2. **Natural Defect Benchmark (`NAT-03`):** Confirmed at **0 natural repairs**. Upstream forensic investigation in Milestone 3.8 Phase 1.2 confirmed that `gorilla/websocket` issue `#789` is unrelated to resource leaks, and upstream `DialContext` manages socket cleanup via `defer` at line 330. `NAT-03` remains labeled `GROUND_TRUTH_DISQUALIFIED` and safely abstains.
3. **No Retrospective Revision:** Historical metrics in `reports/` are protected from accidental timestamp touches or retroactive score inflation.

---

## 5. Verification Invariants

The following invariants are formally established at baseline freeze:
- **Sandbox Boundary:** All repair synthesis and execution occur inside ephemeral directories (`.tmp/mvp-sandbox-*`).
- **Bitwise Source Immutability:** Pre-repair and post-repair cryptographic SHA-256 tree digests must match 100% bitwise on the original repository directory.
- **Fail-Closed Oracles:** If any verification oracle fails or times out, execution status is marked `FAILED` or `EXPLICITLY_ABSTAINED`, and tokens cannot be reused.
- **Honest Verdict Classification:** `INDEPENDENTLY_VERIFIED` is reserved exclusively for runs where an independent dynamic test reproducer confirmed defect elimination; runs relying on static AST checks are classified as `SYNTAX_ONLY_VERIFIED`.
