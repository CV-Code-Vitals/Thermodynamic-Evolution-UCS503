# Phase 0 — Baseline Architecture & Commit Attribution Audit

**Date:** 2026-11-24  
**Auditor:** Principal Systems Architect & Independent Release Qualification Authority  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Current HEAD:** `4b992caf951aef397ec9ef043bb9f1e3755b0cae`  
**Prior Reviewed Baseline:** `d2e9c37717c84a9fe52404c2dee45bd5db99fce2`  
**Latest Remote CI:** PASS (GitHub Actions Run `45376522437` on `4b992ca`)

---

## 1. Repository State & Git Forensics

### 1.1 Working Tree & Remote Status
- **Current Branch:** `release/milestone-3.8-mvp-rc`
- **Tracking Branch:** `origin/release/milestone-3.8-mvp-rc` (synchronized, 0 commits ahead, 0 commits behind).
- **Working Tree State:** Completely clean (`nothing to commit, working tree clean`).
- **Remote Configuration:**
  - `origin`: `https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503.git` (fetch & push).
- **Working Tree Protection:** In accordance with operational constraints, zero uncommitted collaborator changes existed upon start. Local verification preserves all existing code, commits, and branches without resetting, rebasing, or force-pushing.

### 1.2 Commit Topology & Integration History
The baseline review was conducted against `d2e9c37` (`test: add adversarial execution tests and oracle suites for container isolation`). Following that baseline, collaborator `Abhay-coding` merged the parallel repair gateway branch (`6fd1fbb`) into `d2e9c37` via merge commit `6c52cb0`, followed by commit `4b992ca` adjusting the cold-cache Docker vet timeout in `synthesis_test.go`.

```
* 4b992ca (HEAD, origin/release/milestone-3.8-mvp-rc) allow cold-cache Docker vet in synthesis tests
*   6c52cb0 harden remediation security and enforce live sandbox checks in CI
|\
| * d2e9c37 test: add adversarial execution tests and oracle suites for container isolation
| * a0928fb test: add adversarial execution tests and oracle suites
| * 0843ca6 test: add benchmark and adversarial execution tests for repo intel remediation
| * ef65814 fix: align container command policy and offline oracle verification
| * ab07861 test: enforce containerized oracle execution and sandbox availability
* | 6fd1fbb (origin/deployment) fix: restore authenticated repair gateway routes
* | 55b18bf fix: synchronize graph report with viewer readiness
* | eaf34c4 fix: restore admin scan and repair integration
* | 2e4188f fix: connect local repair demo and report unavailable service
* | db3f58d fix: harden repair validation and analysis workflows
* | bb8422a fix: correct benchmark verification results and CI caching
* | c056a41 fix: bound repair verification and release service lock
* | 3ed2b5b fix: secure repair routes and limit public scan staging
|/
* 958d5f1 Merge pull request #29 from CV-Code-Vitals/release/milestone-3.8-mvp-rc
```

---

## 2. Commit-by-Commit Attribution Matrix (Intervening & Collaborator Commits)

| Commit SHA | Author | Date | Modified Subsystems | Intended Feature / Fix | Potential Regression Surfaces | Tests Added / Modified | Verification Status |
|---|---|---|---|---|---|---|---|
| `3ed2b5b` | Abhay-coding | 2026-11-19 | `admin-portal` | Secure repair routes and limit public scan staging | Frontend demo state handling, demo target selection | `remediation-api.test.js` | Verified |
| `c056a41` | Abhay-coding | 2026-11-19 | `go-backend`, `infra/analysis`, `serve.js`, `tests` | Bound repair verification and release service lock | Mutex lock contention during long repair oracles; serve.js proxy paths | `oracles_test.go`, `synthesis_test.go`, `remediation_routes_test.go`, `static-server.test.mjs` | Verified |
| `bb8422a` | Abhay-coding | 2026-11-19 | `.github/workflows`, `go-backend/benchmark` | Correct benchmark verification results and CI caching | Benchmark ground truth preservation; CI cache keys | `milestone3_7_phase2_pilot_test.go`, `milestone3_8_phase1_1_revalidation_test.go` | Verified |
| `db3f58d` | Abhay-coding | 2026-11-19 | `analysis-authorizer`, `go-backend`, `infra/analysis`, `thermodynamic-ast-engine` | Harden repair validation, stale check, and analysis workflows | Stale workspace detection; Rust AST CLI flag handling; stream-proxy CORS | `analysis-authorizer/main_test.go`, `mvp_stale_test.go`, `repo_intel_test.go`, `index.test.mjs` | Verified |
| `2e4188f` | Abhay-coding | 2026-11-19 | `admin-portal`, `api`, `go-backend` | Connect local repair demo and report unavailable service | Local API router fallback; demo button state; error toast | `remediation-api.test.js`, `analysis_proxy_test.go`, `remediation_routes_test.go` | Verified |
| `eaf34c4` | Abhay-coding | 2026-11-19 | `admin-portal`, `go-backend`, `infra/analysis`, `stream-proxy` | Restore admin scan, evolution workspace, and repair integration | Docker image builds; SSE stream proxy routing; evolution API endpoints | `evolution-api.test.js`, `index.test.mjs`, `analysis-template.test.mjs` | Verified |
| `55b18bf` | Abhay-coding | 2026-11-20 | `admin-portal`, `graph-visualizer` | Synchronize 3D graph report with viewer readiness | Race condition between report load and Three.js canvas mount | Manual UI integration | Verified |
| `6fd1fbb` | Abhay-coding | 2026-11-20 | `admin-portal`, `infra/analysis`, `tests` | Restore authenticated repair gateway routes | SAM API Gateway route definitions for `/remediation/*` | `remediation-api.test.js`, `analysis-template.test.mjs` | Verified |
| `6c52cb0` | Abhay-coding | 2026-11-24 | Root `.github`, `go-backend`, `reports` | Merge `6fd1fbb` and `d2e9c37`; enforce live sandbox checks in CI | CI runtime gating; Docker runner container execution | Full Go backend suite, `scan_to_repair_e2e_test.go` | Verified |
| `4b992ca` | Abhay-coding | 2026-11-24 | `go-backend/agent/repo_intel/execution` | Allow cold-cache Docker vet in synthesis tests (raised timeout to 90s) | Test execution duration in CI when Docker image has cold Go module cache | `synthesis_test.go` | Verified |

---

## 3. Repository Architecture & Component Discovery

### 3.1 Component & Service Inventory

```
+---------------------------------------------------------------------------------------------------------+
|                                        THERMODYNAMIC EVOLUTION                                          |
+---------------------------------------------------------------------------------------------------------+
|                                                                                                         |
|  [ FRONTEND TIER ]                                                                                      |
|  - admin-portal (React 18 + Vite 6) - Port 5173 / Static dist                                           |
|    * Views: Scan & Optimize, Graph Visualizer, Evolution Workspace, Autonomous Repair, Archive        |
|  - graph-visualizer (Vanilla JS + Three.js + D3 Force 3D) - Standalone canvas & iframe viewer              |
|                                                                                                         |
|  [ API GATEWAY & PROXY TIER ]                                                                           |
|  - serve.js (Node.js static server & dev reverse proxy) - Port 3000 -> 8080 / 8081 / 5173              |
|  - infra/analysis/stream-proxy (Node.js Express / Lambda SSE proxy) - Port 3001                         |
|  - nginx (Reverse proxy config) - Port 80 / 443                                                         |
|  - analysis-authorizer (Go Lambda / binary) - REST Token Authorizer for AWS API Gateway                 |
|                                                                                                         |
|  [ BUSINESS & EXECUTION TIER ]                                                                          |
|  - go-backend (Go 1.21+ / Gin HTTP Service) - Port 8080                                                 |
|    * /api/scan (Repository cloning, workspace staging, scan orchestration)                              |
|    * /api/evolve (Simulated annealing, thermodynamic mutation, AST evolution)                          |
|    * /api/remediation/* (Autonomous repair: discover, analyze, propose, approve, execute, rollback)    |
|    * Evaluator / Broker (Sandbox execution broker for container containment)                            |
|  - api (Go 1.21+ / Gin / Lambda Service) - Port 8081                                                    |
|    * /api/v1/auth, /api/v1/accounts (Personal account management, session tokens)                      |
|    * /api/v1/archive (Upload, deliverables download, S3 signed forms, retention)                        |
|    * /api/v1/analysis (Proxy to go-backend when running in integrated mode)                             |
|  - source-stager (Go Lambda / binary) - Pre-scan repository download & S3 archive unpacking             |
|                                                                                                         |
|  [ CORE ANALYSIS ENGINE ]                                                                               |
|  - thermodynamic-ast-engine (Rust release binary)                                                        |
|    * Language detection, cyclomatic complexity, AST graph nodes/edges, thermodynamic energy E = αC + βH |
|                                                                                                         |
|  [ ISOLATION SANDBOX RUNNER ]                                                                           |
|  - docker/sandbox-runner/Dockerfile -> ucs503-sandbox-runner:latest                                     |
|    * Ephemeral Docker container execution with --network none, --read-only, --user 10001, --cap-drop ALL |
|    * Offline Go toolchain execution under GOPROXY=off, GOSUMDB=off, GOTOOLCHAIN=local                  |
+---------------------------------------------------------------------------------------------------------+
```

### 3.2 Storage & Trust Boundaries
1. **Host Filesystem vs Workspace Sandbox:**
   - Scanned repositories are cloned or unpacked into ephemeral directories inside `go-backend/uploads/` or OS temp paths.
   - Workspaces enforce path containment via `ValidateContainedPath` and `ValidateWorkspaceID`.
2. **Container Sandbox Boundary:**
   - Untrusted repository code (including test files and dependencies) is strictly forbidden from executing on the host.
   - Dynamic checks (`go vet`, `go test`, reproducers) execute inside `ucs503-sandbox-runner` with `--network none`.
   - Host module cache is protected: only an isolated minimal cache is mounted read-only (`:ro`).
3. **Approval Security Boundary:**
   - Autonomous code repair proposals require HMAC-SHA256 signed approval tokens (`ApprovalToken`).
   - Tokens are single-use, bound to exact target file content hashes, and expire after a strict TTL.
4. **Cloud / Local Abstraction Boundary:**
   - In cloud mode: S3 for archives/deliverables, DynamoDB for account stores, API Gateway with `analysis-authorizer`.
   - In local mode: `local_store.go` for JSON/in-memory accounts, local directory storage for archives, `serve.js` or direct port connections.

---

## 4. Phase 0 Verification & Confirmation

The repository is fully synchronized at `4b992caf951aef397ec9ef043bb9f1e3755b0cae`. All 10 intervening commits by `Abhay-coding` have been audited. No merge conflicts or uncommitted collaborator modifications exist. Architecture mapping confirms clear separation between the Admin Portal/Archive API (`api/`), the Analysis Backend (`go-backend/`), the AST Engine (`thermodynamic-ast-engine/`), and the Sandbox Runner (`docker/sandbox-runner/`).
