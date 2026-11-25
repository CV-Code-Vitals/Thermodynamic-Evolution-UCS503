# Phase 1 — Complete Functional Capability Inventory Matrix

**Date:** 2026-11-24  
**Auditor:** Principal Systems Architect & Independent Release Qualification Authority  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Commit:** `4b992caf951aef397ec9ef043bb9f1e3755b0cae`

---

## 1. Domain A — Repository Scanning & Analysis

| Feature ID | Capability | Entry Point | Backend Implementation | Frontend Implementation | Data Inputs / Outputs | Dependencies | Security Requirements | Test Coverage | Functional Status |
|---|---|---|---|---|---|---|---|---|---|
| **FEAT-A01** | Public GitHub URL Scanning | `POST /api/scan`, `POST /api/analysis/scan` | `go-backend/scanner.go:scanRepoHandler` | `admin-portal/src/components/AdminPortal.jsx` | In: `{repo_url: string, retain_workspace: bool}`<br>Out: `ScanResponse` (metrics, file_reports, workspace_id, targets) | `git clone --depth 1`, `thermodynamic-ast-engine` | SSRF/command injection validation (`safeRepoURL`, HTTPS only, no credentials/flags); timeout 120s | `scan_handler_test.go`, `analysis_proxy_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-A02** | Uploaded Project Archive Analysis | `POST /api/upload`, `POST /api/scan` | `api/router.go`, `go-backend/scanner.go` | `admin-portal/src/components/AdminPortal.jsx` (Dropzone) | In: Multipart `.zip` archive<br>Out: JSON report + workspace ID | `unzipSource`, `workspaceStore` | Zip Slip prevention (`filepath.Clean`, prefix checks); Max size 50MB; decompression bomb checks | `archive_test.go`, `main_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-A03** | Rust Tree-Sitter AST & Complexity Parsing | Subprocess CLI invocation | `thermodynamic-ast-engine/src/main.rs`, `languages.rs`, `relationships.rs` | Visualized via `GraphVisualizer.jsx` | In: Directory path<br>Out: `thermodynamic_report.json` (AST nodes, edges, cyclomatic complexity, energy) | `tree-sitter` parsers (Go, Python, JS, TS, Rust, C, Java, etc.) | Subprocess timeout; sanitized paths; execution in non-root environment | `thermodynamic-ast-engine/src/main.rs` (12 unit tests) | **HEALTHY_VERIFIED** |
| **FEAT-A04** | Thermodynamic Scoring & Energy Calculation | Internal engine calculation | `thermodynamic-ast-engine/src/main.rs:calculate_energy` | Metrics cards in `AdminPortal.jsx`, `EvolutionCharts.jsx` | In: AST nodes, Cyclomatic complexity $C$, Entropy $H$<br>Out: Energy $E = \alpha C + \beta H$ | Rust AST engine | Deterministic formula; zero nodes yields unavailable energy rather than false 0 | `test_energy_formula_basic`, `test_energy_zero_nodes` | **HEALTHY_VERIFIED** |
| **FEAT-A05** | Unified Semantic Diagnostic Orchestrator | Internal pipeline stage | `go-backend/agent/repo_intel/orchestrator.go` | Remediation finding list | In: Target files, AST representation<br>Out: Structural & semantic diagnostic findings | AST traversal, field sensitivity, callgraph analysis | Read-only analysis; no host execution of target code | `repo_intel_test.go`, `remediation_trial_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-A06** | Finding Deduplication & Hotspot Ranking | Orchestrator post-processing | `go-backend/agent/repo_intel/aggregator.go` | Ranked findings in Admin Portal | In: Raw diagnostics<br>Out: Deduplicated, severity-weighted finding list | In-memory aggregation | Prevents alert fatigue and duplicate patch targets | `repo_intel_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-A07** | 3D Graph Visualization Data Export | `GET /api/scan`, `GET /graph-visualizer/` | `go-backend/scanner.go`, `graph-visualizer/graph.mjs` | `graph-visualizer/index.html` iframe in `GraphVisualizer.jsx` | In: Scanned repository AST<br>Out: 3D Force-directed node-link graph JSON | Three.js, D3-Force-3D | CSP headers; frame-ancestors 'self'; sanitized node labels | `graph-visualizer-config.test.mjs` | **HEALTHY_VERIFIED** |
| **FEAT-A08** | Workspace Retention & Expiration Lifecycle | `workspaceStore` background reaper | `go-backend/workspace.go:workspaceStore` | Workspace ID persistence in frontend state | In: Workspace ID<br>Out: Directory lease or 404/410 on expiration | OS filesystem, sync.RWMutex | 30-minute TTL; path traversal rejection; reap stale workspaces | `workspace_test.go`, `mvp_stale_test.go` | **HEALTHY_VERIFIED** |

---

## 2. Domain B — Evolution & Optimization

| Feature ID | Capability | Entry Point | Backend Implementation | Frontend Implementation | Data Inputs / Outputs | Dependencies | Security Requirements | Test Coverage | Functional Status |
|---|---|---|---|---|---|---|---|---|---|
| **FEAT-B01** | Evolution Target Selection | `POST /api/evolve`, `GET /api/evolve` | `go-backend/evolve.go:evolveHandler` | `admin-portal/src/components/EvolutionWorkspace.jsx` | In: `{workspace_id, target_files, max_iterations}`<br>Out: SSE stream or JSON result | `workspaceStore`, `thermodynamic-ast-engine` | Workspace path containment; lease locking | `evolve_stream_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-B02** | Mutation Generation & Strategy Selection | Internal agent loop | `go-backend/agent/mutation.go`, `strategy.go` | Displayed in live evolution charts | In: AST target node<br>Out: Mutated candidate source | AST rewriting | AST-preserving mutations; no unverified syntax | `function_mutation_test.go`, `strategy_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-B03** | Simulated Annealing Decision Engine | Internal loop | `go-backend/agent/annealer.go:Annealer` | Live temperature and acceptance curve | In: Current energy $E$, Candidate energy $E'$, Temp $T$<br>Out: Accept/Reject boolean | Math probability $e^{-\Delta E / T}$ | Deterministic cooling; preserves best candidate seen | `annealer_benchmark_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-B04** | Sandbox Evaluation Broker | Internal evaluation | `go-backend/agent/broker_evaluator.go`, `evaluator.go` | Status pills in live workspace | In: Candidate source code<br>Out: Compilation & runtime validation result | Container runner or AST check | Non-root execution; memory and CPU limits | `evaluator_test.go`, `broker_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-B05** | Live Telemetry & SSE Streaming | `GET /api/evolve` (SSE) | `go-backend/evolve.go`, `stream-proxy/index.mjs` | `admin-portal/src/evolution-api.js:readEvolutionStream` | In: Active evolution job<br>Out: `event: progress`, `event: complete`, `event: error` | Node.js stream-proxy / Gin SSE | Connection heartbeats; abort signal handling; buffer flushing | `index.test.mjs`, `evolution-api.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-B06** | Execution Cancellation | Client abort / context cancel | `go-backend/evolve.go`, `agent/workspace.go` | Cancel button in `EvolutionLive.jsx` | In: Abort signal / disconnect<br>Out: Early termination & resource cleanup | Go context cancellation | Guaranteed worker termination; release workspace lease | `evolve_stream_test.go` | **HEALTHY_VERIFIED** |

---

## 3. Domain C — Autonomous Code Repair

| Feature ID | Capability | Entry Point | Backend Implementation | Frontend Implementation | Data Inputs / Outputs | Dependencies | Security Requirements | Test Coverage | Functional Status |
|---|---|---|---|---|---|---|---|---|---|
| **FEAT-C01** | Scanned & Demo Target Discovery | `GET /api/remediation/targets` | `go-backend/remediation_routes.go:targetsHandler` | `admin-portal/src/components/RemediationPortal.jsx` | In: `workspace_id` (optional)<br>Out: `targets: []RemediationTarget` | `MVPService.ListTargets`, `mvp_scanner.go` | Clear stratum separation (`SCANNED_REPOSITORY` vs `[DEMO]`) | `scan_repair_test.go`, `remediation-api.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-C02** | Target Diagnostic Analysis | `POST /api/remediation/analyze` | `go-backend/remediation_routes.go:analyzeHandler` | Target details view in `RemediationPortal.jsx` | In: `{target_id: string}`<br>Out: `RemediationAnalysis` | `MVPService.AnalyzeTarget` | Verifies target existence; path containment; ownership analysis | `scan_repair_test.go`, `remediation_routes_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-C03** | Autonomous Patch Synthesis | `POST /api/remediation/propose` | `go-backend/remediation_routes.go:proposeHandler` | Unified diff preview in `RemediationPortal.jsx` | In: `{analysis_id: string}`<br>Out: `RepairProposal` (unified_diff, patch_id) | `AutonomousPatchSynthesizer` | AST synthesis only; no unverified string splicing; enforces safe replace | `synthesis_test.go`, `scan_repair_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-C04** | HMAC Approval Authorization | `POST /api/remediation/approve` | `go-backend/remediation_routes.go:approveHandler` | Approval modal in `RemediationPortal.jsx` | In: `{proposal_id: string}`<br>Out: `ApprovalToken` (HMAC signature, token_id) | HMAC-SHA256, `crypto/hmac` | 15-minute expiry; bound to file hash; nonces prevent replay | `remediation_trial_test.go`, `adversarial_execution_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-C05** | Sandbox Execution & Oracle Verification | `POST /api/remediation/execute` | `go-backend/remediation_routes.go:executeHandler` | Oracle verification indicators | In: `{token_id: string}`<br>Out: `ExecutionResult` (5 oracle reports, verdict) | `DockerContainerRunner`, `IndependentOracleSuite` | Untrusted code executes in container only (`--network none`, read-only root, minimal modcache) | `oracles_test.go`, `adversarial_execution_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-C06** | Rollback & Sandbox Cleanup | `POST /api/remediation/rollback` | `go-backend/remediation_routes.go:rollbackHandler` | Reset button in `RemediationPortal.jsx` | In: `{execution_id: string}`<br>Out: `{success: bool}` | `WorkspaceManager.RollbackSnapshot` | Guarantees original workspace restoration | `scan_repair_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-C07** | Machine-Readable Audit Ledger | `GET /api/remediation/audit` | `go-backend/remediation_routes.go:auditHandler` | Audit tab in `RemediationPortal.jsx` | In: None<br>Out: Complete JSON audit trail | In-memory ledger + JSON file backing | Cryptographic token consumption records; tamper detection | `remediation_trial_test.go`, `remediation-api.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-C08** | Deterministic Demo Execution | `POST /api/remediation/demo` | `go-backend/remediation_routes.go:demoHandler` | "Run Demo" button in `RemediationPortal.jsx` | In: None<br>Out: Demo execution outcome | Pre-configured target execution | Labeled clearly as demo; isolated from live workspace | `remediation-api.test.js` | **HEALTHY_VERIFIED** |

---

## 4. Domain D — Admin Portal & User Experience

| Feature ID | Capability | Entry Point | Backend Implementation | Frontend Implementation | Data Inputs / Outputs | Dependencies | Security Requirements | Test Coverage | Functional Status |
|---|---|---|---|---|---|---|---|---|---|
| **FEAT-D01** | Administrator Authentication & Login | `/admin-portal/` login form | `api/auth.go`, `api/accounts.go` | `admin-portal/src/components/AdminPortal.jsx` | In: Passkey / credentials<br>Out: Session bearer token | `Authenticator` | Bearer tokens; no cookies; session rotation on refresh | `auth_test.go`, `accounts_test.go`, `App.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-D02** | Session Persistence & Refresh Lifecycle | Periodic background refresh | `POST /api/v1/auth/refresh` | `admin-portal/src/archive-api.js` | In: Active bearer<br>Out: Rotated bearer token | In-memory / DynamoDB store | Clock skew tolerance; server-directed expiry | `refresh_test.go`, `archive-api.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-D03** | Multi-View Tab Navigation | Top navigation bar | Static React routing | `admin-portal/src/App.jsx` | In: View tab selection<br>Out: Active view component | React state | Unauthenticated access redirects to login | `App.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-D04** | Error Feedback & Unavailable Service Alerts | Global alert banners | Backend error responses (`502`, `503`, `401`) | Toast/Banner components in `RemediationPortal.jsx`, `AdminPortal.jsx` | In: Error JSON<br>Out: User-friendly error message | React state | Clear distinction between network failure, missing container, and defect | `remediation-api.test.js`, `evolution-api.test.js` | **HEALTHY_VERIFIED** |
| **FEAT-D05** | Deliverables & Archive Listing | `/admin-portal/` deliverables view | `GET /api/archives` | `admin-portal/src/components/AdminPortal.jsx` | In: Bearer token<br>Out: Archive items array | `api/archive.go` | Authorized access only; authenticated download URL generation | `archive_test.go`, `archive-api.test.js` | **HEALTHY_VERIFIED** |

---

## 5. Domain E — APIs & Supporting Services

| Feature ID | Capability | Entry Point | Backend Implementation | Frontend Implementation | Data Inputs / Outputs | Dependencies | Security Requirements | Test Coverage | Functional Status |
|---|---|---|---|---|---|---|---|---|---|
| **FEAT-E01** | Reverse Proxy & Development Gateway | `serve.js` (:8000) | `serve.js` | Connects browser to backend (:8080) | In: HTTP requests<br>Out: Proxied responses | Node.js `http` | Strips hop-by-hop headers; CSP headers; prevents traversal | `static-server.test.mjs` | **HEALTHY_VERIFIED** |
| **FEAT-E02** | Streaming SSE Proxy | `infra/analysis/stream-proxy` (:3001) | `stream-proxy/index.mjs` | `admin-portal/src/evolution-api.js` | In: SSE stream request<br>Out: Buffered SSE events | Node.js `http`/`https` | Buffer size limit; CORS origin check; heartbeat timeout | `stream-proxy/index.test.mjs` | **HEALTHY_VERIFIED** |
| **FEAT-E03** | REST Token Authorizer | AWS Lambda API Gateway | `analysis-authorizer/main.go` | Implicit in API Gateway requests | In: `Authorization: Bearer <token>`<br>Out: IAM Policy (Allow/Deny) | AWS Lambda Go runtime | Constant-time token verification; rejects empty/malformed tokens | `analysis-authorizer/main_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-E04** | Source Staging Service | AWS Lambda | `source-stager/main.go` | Invoked during cloud scan | In: S3 Bucket & Key<br>Out: Unpacked directory in `/tmp` | AWS SDK v2 S3 | Zip Slip prevention; size limits | `source-stager/main_test.go` | **HEALTHY_VERIFIED** |
| **FEAT-E05** | Multi-Container Compose Configuration | `docker-compose*.yml` | Docker Engine / Compose | N/A | In: Compose YAML<br>Out: Networked container services | Docker CLI / Compose v2 | Network isolation; split configurations (`analysis` vs `sandbox`) | `scripts/compose-config.test.mjs` | **HEALTHY_VERIFIED** |
| **FEAT-E06** | CloudFormation / SAM Infrastructure | `template.yaml`, `infra/analysis/template.yaml` | AWS SAM CLI | N/A | In: SAM template<br>Out: CloudFormation stack | AWS SAM CLI | Strict IAM roles; least privilege; public/private subnet separation | `tests/analysis-template.test.mjs` | **HEALTHY_VERIFIED** |
