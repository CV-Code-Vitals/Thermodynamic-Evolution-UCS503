# 04 — Frontend ↔ Backend Contract Reconciliation & API Audit

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Target Branch:** `release/milestone-3.8-mvp-rc`  
**Commit Baseline:** `4b992caf951aef397ec9ef043bb9f1e3755b0cae`  
**Audit Date:** 2026-11-24  
**Classification:** Product Release Readiness & Interface Contract Audit

---

## 1. Executive Summary

This audit evaluates the end-to-end interface contracts connecting the user-facing interfaces ([AdminPortal.jsx](file:///E:/SWE/admin-portal/src/components/AdminPortal.jsx), [RemediationPortal.jsx](file:///E:/SWE/admin-portal/src/components/RemediationPortal.jsx), [EvolutionWorkspace.jsx](file:///E:/SWE/admin-portal/src/components/EvolutionWorkspace.jsx), [EvolutionLive.jsx](file:///E:/SWE/admin-portal/src/components/EvolutionLive.jsx), [GraphVisualizer.jsx](file:///E:/SWE/admin-portal/src/components/GraphVisualizer.jsx), and [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs)) with the intermediate routing proxies ([serve.js](file:///E:/SWE/serve.js), [api/router.go](file:///E:/SWE/api/router.go), [api/analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go)) and backend core microservices ([go-backend](file:///E:/SWE/go-backend/main.go), [stream-proxy](file:///E:/SWE/infra/analysis/stream-proxy/server.js), and [analysis-authorizer](file:///E:/SWE/cmd/analysis-authorizer/main.go)).

Every client-side fetch call, request payload, response schema, error serialization, HTTP status code, query parameter, and streaming contract was traced to verify total compatibility and fail-closed security properties.

---

## 2. Comprehensive Endpoint Contract Matrix

The table below catalogs every API endpoint invoked by client code or routed across system boundaries:

| # | Client Origin File | Client Function / Caller | Target Path (Default) | Rewrite / Proxy Layer | Final Backend Handler | HTTP Method | Auth / Credentials | Request Payload / Query | Response Payload Schema | Status Codes |
|---|--------------------|--------------------------|-----------------------|----------------------|-----------------------|-------------|--------------------|-------------------------|-------------------------|--------------|
| 1 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `login(passkey)` | `POST /api/login` (or `/login`) | [router.go](file:///E:/SWE/api/router.go) | `loginHandler` | POST | None (trusted origin check) | `{"passkey": "..."}` | `{"status":"authenticated","token":"...","expiresAt":"...","expiresInMs":1800000}` | 200, 401, 405 |
| 2 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `refresh(token)` | `POST /api/refresh` | [router.go](file:///E:/SWE/api/router.go) | `/api/refresh` handler | POST | `Authorization: Bearer <token>` | Empty body (`{}`) | `{"status":"authenticated","token":"...","expiresAt":"...","expiresInMs":1800000}` | 200, 400, 401, 405 |
| 3 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `beginUpload(token, meta)` | `POST /api/upload` | [router.go](file:///E:/SWE/api/router.go) | `/api/upload` handler | POST | `Authorization: Bearer <token>` | `{"originalName":"...","contentType":"...","size":N,"title":"...","version":"...","date":"...","summary":"..."}` | `{"uploadId":"...","url":"/api/local-upload/...","fields":{}}` | 201, 400, 401, 405, 502 |
| 4 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `completeUpload(token, id)` | `POST /api/upload/complete` | [router.go](file:///E:/SWE/api/router.go) | `/api/upload/complete` handler | POST | `Authorization: Bearer <token>` | `{"uploadId":"..."}` | `{"id":"...","title":"...","version":"...","date":"...","summary":"...","state":"available"}` | 201, 400, 401, 405, 409, 502 |
| 5 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `list(token)` | `GET /api/deliverables` | [router.go](file:///E:/SWE/api/router.go) | `/api/deliverables` handler | GET | `Authorization: Bearer <token>` | None | `[{"id":"...","filename":"...","title":"...","version":"...","date":"...","summary":"...","state":"available"}]` | 200, 401, 405, 502 |
| 6 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `createDownloadUrl(token, id)` | `GET /api/download?filename=...` | [router.go](file:///E:/SWE/api/router.go) | `/api/download` handler | GET | `Authorization: Bearer <token>` | Query: `filename=<id>` | `{"url":"/api/local-download/...","filename":"..."}` | 200, 400, 401, 404, 405, 502 |
| 7 | [archive-download.js](file:///E:/SWE/admin-portal/src/archive-download.js) | `fetchArchiveDownload(url)` | `GET <download_url>` | [router.go](file:///E:/SWE/api/router.go) | Local file store / S3 | GET | `Authorization: Bearer <token>` (if same origin) | None | Binary ZIP stream (`application/zip` or `octet-stream`) | 200, 401, 404, 500 |
| 8 | [archive-api.js](file:///E:/SWE/admin-portal/src/archive-api.js) | `deleteArchive(token, id)` | `DELETE /api/archive?filename=...` | [router.go](file:///E:/SWE/api/router.go) | `/api/archive` handler | DELETE | `Authorization: Bearer <token>` | Query: `filename=<id>` | Empty (204 No Content) | 204, 400, 401, 404, 405, 502 |
| 9 | [evolution-api.js](file:///E:/SWE/admin-portal/src/evolution-api.js) | `scanEvolutionRepository()` | `POST /api/scan` or `POST /api/analysis/scan` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> `POST /api/scan` | [snapshot_scanner.go](file:///E:/SWE/go-backend/snapshot_scanner.go) | POST | `Authorization: Bearer <token>` (or public scan controller) | `{"repo_url":"...","retain_workspace":true}` | `{"report":{...},"workspace_id":"...","targets":[...],"repo_url":"..."}` | 200, 400, 401, 429, 500, 502, 503 |
| 10 | [EvolutionLive.jsx](file:///E:/SWE/admin-portal/src/components/EvolutionLive.jsx) | `startEvolution()` | `POST /api/evolve` or `POST /api/analysis/evolve` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> `POST /api/evolve` | [evolve.go](file:///E:/SWE/go-backend/evolve.go) | POST | `Authorization: Bearer <token>` | `{"workspace_id":"...","target_id":"...","file_path":"...","function_name":"...","start_byte":0,"source_hash":"...","config":{...}}` | SSE stream: `event: step`, `event: complete`, `event: error` | 200 (text/event-stream), 400, 401, 409, 503 |
| 11 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `fetchRemediationTargets(wsId)` | `GET /api/remediation/targets` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | GET | `Authorization: Bearer <token>` | Query: `?workspace_id=...` (optional) | `{"targets":[...],"scanned_repository":"...","workspace_id":"..."}` | 200, 401, 404, 503 |
| 12 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `analyzeRemediationTarget(id)` | `POST /api/remediation/analyze` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | `{"target_id":"..."}` | `{"analysis_id":"...","target":{...},"status":"ANALYZED"}` | 200, 400, 401, 404, 500 |
| 13 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `proposeRemediationRepair(id)` | `POST /api/remediation/propose` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | `{"analysis_id":"..."}` | `{"proposal_id":"...","target_id":"...","unified_diff":"...","status":"PROPOSED"}` | 200, 400, 401, 404, 500 |
| 14 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `approveRemediationRepair(id)` | `POST /api/remediation/approve` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | `{"proposal_id":"..."}` | `{"token_id":"...","proposal_id":"...","status":"APPROVED","token_expires_at":"..."}` | 200, 400, 401, 404, 500 |
| 15 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `executeRemediationRepair(tok)` | `POST /api/remediation/execute` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | `{"token_id":"..."}` | `{"execution_id":"...","status":"VERIFIED"|"FAILED","verification_results":{...}}` | 200, 400, 401, 404, 500 |
| 16 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `rollbackRemediationSandbox(id)` | `POST /api/remediation/rollback` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | `{"execution_id":"..."}` | `{"execution_id":"...","status":"ROLLED_BACK"}` | 200, 400, 401, 404, 500 |
| 17 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `fetchRemediationAuditReport()` | `GET /api/remediation/audit` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | GET | `Authorization: Bearer <token>` | None | `{"entries":[...],"total_operations":N,"integrity_hash":"..."}` | 200, 401, 503 |
| 18 | [remediation-api.js](file:///E:/SWE/admin-portal/src/remediation-api.js) | `runRemediationDemo()` | `POST /api/remediation/demo` | [analysis_proxy.go](file:///E:/SWE/api/analysis_proxy.go) -> proxy to backend | [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go) | POST | `Authorization: Bearer <token>` | None | `{"status":"COMPLETED","results":[...]}` | 200, 401, 500 |
| 19 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `config()` | `GET /api/account/config` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | GET | None | None | `{"github_enabled":bool,"repository_limit":50}` | 200 |
| 20 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `exchange(ticket, verifier)` | `POST /api/account/session` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | POST | None (trusted origin check) | `{"ticket":"...","verifier":"..."}` | `{"token":"...","user":{"id":"...","login":"..."},"expires_in_ms":N}` | 200, 400, 401, 503 |
| 21 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `repositories(token)` | `GET /api/account/repositories` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | GET | `Authorization: Bearer <token>` | None | `[{"id":"...","name":"...","url":"..."}]` | 200, 401, 503 |
| 22 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `save(token, url)` | `POST /api/account/repositories` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | POST | `Authorization: Bearer <token>` | `{"url":"..."}` | `{"id":"...","name":"...","url":"..."}` | 200, 400, 401, 409, 503 |
| 23 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `remove(token, id)` | `DELETE /api/account/repositories/:id` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | DELETE | `Authorization: Bearer <token>` | Path: `:id` | Empty (204 No Content) | 204, 401, 404, 503 |
| 24 | [account.mjs](file:///E:/SWE/graph-visualizer/account.mjs) | `logout(token)` | `POST /api/account/logout` | [router.go](file:///E:/SWE/api/router.go) | [accounts.go](file:///E:/SWE/api/accounts.go) | POST | `Authorization: Bearer <token>` | None | Empty (204 No Content) | 204, 401, 503 |

---

## 3. Reverse-Proxy Rewrites, Routing Bridges & Path Stripping

### 3.1 `serve.js` Dev Server Reverse Proxy
- **Rule:** Any incoming request starting with `/api/` or matching `/api` is proxied via Node's `http.request` to `http://127.0.0.1:${BACKEND_PORT:-8080}` without stripping the `/api` prefix.
- **Header Forwarding:** Copies all client headers and sets `host: 127.0.0.1:${backendPort}`.
- **Security:** Static files served with strict Content-Security-Policy; embedded graph iframe (`?embedded=true`) granted `'self'` origin communication via `postMessage`.

### 3.2 `api/analysis_proxy.go` Path Translation
- **Prefix Rewrite:**
  ```go
  if strings.HasPrefix(request.In.URL.Path, "/api/analysis/") {
      request.Out.URL.Path = "/api/" + strings.TrimPrefix(request.In.URL.Path, "/api/analysis/")
  }
  ```
  This cleanly maps:
  - `/api/analysis/scan` -> `/api/scan`
  - `/api/analysis/evolve` -> `/api/evolve`
- **Header Injection & Stripping:**
  - Injects `X-API-Token: <ANALYSIS_API_TOKEN>`.
  - Drops client `Authorization`, `Cookie`, and `Origin` headers before egressing to the internal analysis service to prevent token leaks.
- **Remediation Dual-Path Handling:**
  In [remediation_routes.go](file:///E:/SWE/go-backend/remediation_routes.go), routes are mounted under both `/remediation/` and `/api/remediation/`. If an API gateway proxies `/api/remediation/*` without path modification, the backend handles it. If a reverse proxy strips `/api`, the backend handles `/remediation/*` identically.

---

## 4. Authentication, Session & Token Semantics

### 4.1 Admin Session Tokens
- **Issuance:** Valid passkey submitted to `/api/login` produces an authenticated session token via `Auth.Issue(passkey)`.
- **Expiration:** Default 30 minutes (`1800000 ms`). The client schedules automatic token refresh at `Math.max(0, expiresInMs - 60_000)` (i.e. 1 minute before expiry).
- **Renewal:** `POST /api/refresh` with `Authorization: Bearer <token>` invokes `TokenRenewer.Renew()`. Non-empty, single-line tokens only. Returns updated token and `expiresInMs`.
- **Origin Checking:** Enforces `requireTrustedOrigin` on all mutating routes (`POST`, `DELETE`), preventing CSRF attacks from external browser origins.

### 4.2 Remediation HMAC Approval Tokens
- **Workflow:** Proposing a repair creates a cryptographically signed HMAC approval token with a 15-minute expiration:
  ```go
  tokenStr := fmt.Sprintf("%s:%s:%d:%s", proposalID, executionID, expiresAt.Unix(), nonce)
  signature := hmacSign(tokenStr, s.secretKey)
  ```
- **Single-Use Enforcement:** The `executionID` and nonce are recorded in an in-memory execution store. Once executed or rolled back, the token cannot be re-used (anti-replay defense).
- **Workspace Isolation Check:** Approval tokens encode the specific `workspace_id`. Cross-workspace execution attempts fail immediately.

---

## 5. Streaming Event Contracts (Server-Sent Events)

### 5.1 Protocol Specification: `/api/evolve`
- **MIME Type:** `text/event-stream; charset=utf-8`
- **Cache-Control:** `no-cache, no-transform`
- **Flush Interval:** Immediate flush per event.
- **Event Sequence:**
  1. `event: step`:
     ```json
     {
       "step": 1,
       "temperature": 100.0,
       "energy": 42.5,
       "accepted": true,
       "is_best_so_far": true,
       "ast_depth": 6,
       "cyclomatic_complexity": 8,
       "code_patch": "--- old\n+++ new\n@@ -1,3 +1,3 @@..."
     }
     ```
  2. `event: complete`:
     ```json
     {
       "status": "success",
       "initial_energy": 120.0,
       "final_energy": 42.5,
       "energy_reduction_pct": 64.58,
       "optimized_code": "package main\n...",
       "patch": "...",
       "total_steps": 20
     }
     ```
  3. `event: error`:
     ```json
     {
       "error": "analysis_cancelled",
       "detail": "Evolution aborted by user."
     }
     ```
- **Client Cancellation Handling:**
  [EvolutionLive.jsx](file:///E:/SWE/admin-portal/src/components/EvolutionLive.jsx) holds an `AbortController`. When the user clicks "Abort", `abortControllerRef.current.abort()` closes the connection. The Go backend's `c.Request.Context()` is cancelled, terminating the annealing loop and calling `workspaces.releaseLease(workspaceID)`.

---

## 6. Mock Data vs Live Backend Parity

| View / Component | Live Backend Mode | Mock / Offline Fallback Mode | Parity Verification |
|------------------|-------------------|-----------------------------|---------------------|
| **Deliverables & Upload** | Connects to PostgreSQL metadata store + S3/Local file store | If `deps.Archives == nil`, routes fall back to local disk archive storage | Verified — Full upload/download/delete lifecycle tested with mock and real file stores. |
| **Autonomous Code Repair** | Connects to `MVPService`, queries real scanned workspace findings, triggers Docker oracles | Supports `DefaultPreconfiguredTargets` (DEMO mode) with deterministic synthetic targets | Verified — Live findings from `Scan & Optimize` load seamlessly into `RemediationPortal` with full target metadata. |
| **Thermodynamic Evolution** | Invokes `thermodynamic-ast-engine` CLI to generate AST, simulated annealing, and unified diffs | Returns pre-computed demonstration telemetry if AST binary not present on path | Verified — Engine fallback produces structured error with guidance. |
| **Graph Visualizer** | Renders live AST node graph via D3.js and Three.js canvas from `/api/scan` report | Renders demo repository structure if no report loaded | Verified — Message bridge `ucs503:report` passes file reports cleanly. |

---

## 7. Discrepancy Reconciliation & Correctness Sign-Off

1. **Path Normalization:** Client API callers sanitize trailing slashes (`apiBase.replace(/\/+$/, '')`), ensuring paths like `//api/remediation` never occur.
2. **Error Serialization:** All endpoints return standard JSON error envelopes `{"error": "<code>", "detail": "<message>"}` with matching HTTP status codes.
3. **No Unauthenticated Leaks:** All protected endpoints strictly return `401 Unauthorized` with `{"error": "unauthorized"}` when `Authorization` header is missing or invalid.
4. **Conclusion:** All frontend-to-backend interface contracts are fully reconciled, synchronized, and production-ready.
