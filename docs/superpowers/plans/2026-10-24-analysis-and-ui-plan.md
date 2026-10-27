> **Superseded draft:** Do not execute this plan independently. The authoritative plan is [2026-10-24-whole-project-hosting-plan.md](2026-10-24-whole-project-hosting-plan.md), which replaces the HTTP API authorizer assumptions with the approved REST API response-streaming architecture and integrates source staging, sandbox lifecycle, and the cost/authorization gates.

# Analysis API and UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the public Graph Visualizer's live scans and the Admin Portal's authenticated Evolution flow against a separate hosted analysis API.

**Architecture:** A separate API Gateway Lambda authorizer validates Admin Portal bearer tokens for admin-only analysis routes. Public Graph scans have a distinct anonymous route with strict request, rate, concurrency, and resource controls. The Graph page remains independent; Evolution accepts a GitHub repository and target file/function from within the Admin Portal.

**Tech Stack:** Go 1.25, AWS Lambda Go authorizer, API Gateway HTTP API, SSM Parameter Store, React 19, Vite 8, Node's built-in test runner.

**Spec:** `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`

## Global Constraints

- Keep the main site and Admin Portal UI on GitHub Pages.
- Keep the live `ucs503-admin-archive-v2` API Gateway, Lambda, DynamoDB, private S3, secrets, and CI/CD unchanged.
- Keep `VITE_API_BASE` for the archive API; use a separate `VITE_ANALYSIS_API_BASE` for analysis calls.
- Public Graph Visualizer scans are anonymous, but strictly rate-limited, input-validated, and resource-bounded.
- Evolution and other admin analysis routes require the existing short-lived Admin Portal bearer token.
- Never place a passkey, HMAC key, sandbox token, `ANALYSIS_API_TOKEN`, or AWS credential in browser code, URLs, logs, or source control.
- Keep the archive token-signing key off both EC2 instances.
- Do not add RDS or deploy the duplicate `ucs503-api` service.
- No AWS resources are changed without separate approval of the final cost estimate and deployment plan.

## Review Focus

- Malformed, expired, or tampered Admin Portal tokens must fail closed at the authorizer.
- Public scan requests must not accept arbitrary hosts, private addresses, or forged client-IP headers.
- Scan bursts and concurrent jobs must be rejected with an explicit retryable response rather than overloading the instance.
- Oversized, invalid, or timed-out GitHub repositories must be rejected and cleaned up.
- Evolution SSE must surface HTTP/API errors and preserve cancellation without leaving an active job or workspace.

---

### Task 1: Add the isolated API Gateway authorizer

**Files:**
- Create: `analysis-authorizer/go.mod`
- Create: `analysis-authorizer/main.go`
- Create: `analysis-authorizer/main_test.go`

**Interfaces:**
- Consumes: The existing token format in `api/auth.go`: base64url payload, expiration timestamp, random nonce, and HMAC-SHA256 signature.
- Produces: `VerifyArchiveBearer(token, signingKey string, now time.Time) bool` and an API Gateway HTTP API Lambda authorizer handler that returns `isAuthorized` plus an internal `analysisApiToken` context value only after successful verification.

- [ ] **Step 1: Write failing token-verification tests**

Add `TestVerifyArchiveBearer_AcceptsValidToken`, `TestVerifyArchiveBearer_RejectsMalformedToken`, `TestVerifyArchiveBearer_RejectsTamperedToken`, and `TestVerifyArchiveBearer_RejectsExpiredToken`. Use a test-only signing key and fixed clock; assert that only a valid, unexpired token authorizes.

- [ ] **Step 2: Run the tests and confirm the verifier is missing**

Run from `analysis-authorizer`: `go test ./...`

Expected: FAIL because `VerifyArchiveBearer` is not implemented.

- [ ] **Step 3: Implement the verifier and handler**

Implement the archive token verification format without an issuance function. Read `/ucs503/admin/token-hmac-key` and `/ucs503/analysis/api-token` as SecureStrings through the authorizer's least-privilege SSM role. Return the internal API token only in authorizer context for API Gateway integration mapping; never log token values or return them in a client response.

- [ ] **Step 4: Verify token and handler behavior**

Run from `analysis-authorizer`: `go test ./...` and `go vet ./...`

Expected: PASS; invalid tokens return `isAuthorized=false`, and valid tokens return no signing key or bearer token in the response context.

- [ ] **Step 5: Commit**

```text
feat: add analysis API token authorizer
```

### Task 2: Bound and validate anonymous repository scans

**Files:**
- Modify: `go-backend/scanner.go`
- Modify: `go-backend/main.go`
- Modify: `go-backend/limiter.go`
- Modify: `go-backend/limiter_test.go`
- Modify: `go-backend/main_test.go`
- Modify: `go-backend/go.mod` only if an already-approved implementation needs a dependency

**Interfaces:**
- Consumes: `ScanRepoRequest` and `scanRepoHandler` in `scanner.go`; `AdmissionController` in `limiter.go`.
- Produces: A shared GitHub URL validator and a public scan admission controller that receives the trusted API Gateway client IP, enforces one concurrent scan and an initial per-IP limit of one scan per minute, and returns an explicit `429`/`503` when full. The initial public scan limit is 50 MiB GitHub-reported repository size, 120 seconds to clone, and the configured engine timeout.

- [ ] **Step 1: Add failing scan-boundary tests**

Add `TestValidatePublicGitHubURL_RejectsNonHTTPSAndNonGitHub`, `TestValidatePublicGitHubURL_RejectsUserInfoAndIPLiteral`, `TestPublicScanController_RateLimitsClient`, `TestPublicScanController_RejectsConcurrentScan`, and `TestScanRepoHandler_CleansTemporaryClone`. Assert that requests for non-public or non-GitHub targets never invoke `git`.

- [ ] **Step 2: Run the focused Go tests and confirm failure**

Run from `go-backend`: `go test ./...`

Expected: FAIL for the new missing validators/controller tests.

- [ ] **Step 3: Implement bounded scan processing**

Require HTTPS `github.com` repository URLs with no user info, custom port, or IP-literal host. Check GitHub repository metadata and reject repositories above 50 MiB before cloning; enforce a 120-second clone timeout and configured engine timeout. Disable clone redirects to unapproved hosts, reject DNS results resolving to private/reserved targets, and fail closed on validation errors. Use only the API Gateway-overwritten client-IP header for per-IP limits; do not trust client-supplied `X-Forwarded-For`. Register `/api/scan` without the static-token middleware only because the analysis instance is reachable through the private API Gateway integration; all admin analysis routes remain protected.

- [ ] **Step 4: Verify scan limits and cleanup**

Run from `go-backend`: `go test ./...` and `go vet ./...`

Expected: PASS; rejected requests do not clone, concurrent/rate-limited requests receive explicit status codes, and temporary repositories are removed on success, failure, and timeout.

- [ ] **Step 5: Commit**

```text
feat: add bounded public repository scans
```

### Task 3: Add an authenticated GitHub-based Evolution workflow

**Files:**
- Modify: `go-backend/evolve.go`
- Modify: `go-backend/main.go`
- Modify: `go-backend/main_test.go`
- Modify: `go-backend/scanner.go`
- Modify: `go-backend/agent/annealer.go`

**Interfaces:**
- Consumes: The validated repository clone helper from Task 2 and the workspace-session contract from `2026-10-24-sandbox-worker-plan.md`.
- Produces: `POST /api/evolution/targets` accepting `{ "repo_url": "..." }` and returning available Go file/function targets; `POST /api/evolve` accepting `{ "repo_url": "...", "commit_sha": "...", "file_path": "...", "function_name": "...", "config": {...} }`. Neither endpoint accepts a server-local `repo_path` or browser-supplied byte offsets. Go is the initial supported Evolution language because the existing runner executes `go test ./...`.

- [ ] **Step 1: Add failing target-resolution and request tests**

Add `TestResolveEvolutionTargets_ListsGoFunctions`, `TestResolveEvolutionTargets_RejectsMissingFunction`, `TestEvolutionTargets_RejectsInvalidRepository`, and `TestEvolveRequest_RejectsServerLocalPaths`. Assert that function byte offsets are resolved server-side from the selected file and the requested commit is used.

- [ ] **Step 2: Run the focused Go tests and confirm failure**

Run from `go-backend`: `go test ./...`

Expected: FAIL because the target resolver and new route contracts are missing.

- [ ] **Step 3: Implement target discovery and bounded Evolution**

Clone only the validated public GitHub repository into a temporary directory, pin the selected commit, resolve the named Go function using `go/parser`/`go/token`, and invoke the existing annealer with server-derived byte offsets. Use the remote workspace session from the sandbox plan; always clean up the clone and remote workspace, including cancellation and error paths.

- [ ] **Step 4: Verify authenticated Evolution behavior**

Run from `go-backend`: `go test ./...` and `go vet ./...`

Expected: PASS; invalid repository/target inputs fail visibly, and no request can select an arbitrary local filesystem path.

- [ ] **Step 5: Commit**

```text
feat: accept validated GitHub targets for evolution
```

### Task 4: Wire the frontend to distinct public and admin API flows

**Files:**
- Create: `admin-portal/src/evolution-api.js`
- Create: `admin-portal/tests/evolution-api.test.js`
- Modify: `admin-portal/src/components/AdminPortal.jsx`
- Modify: `admin-portal/src/components/EvolutionLive.jsx`
- Modify: `admin-portal/vite.config.js`
- Modify: `admin-portal/tests/vite-config.test.js`
- Create: `graph-visualizer/scan-client.js`
- Create: `admin-portal/tests/graph-scan-client.test.js`
- Modify: `graph-visualizer/index.html`
- Modify: `index.html`
- Modify: `.github/workflows/deploy-frontend.yml`

**Interfaces:**
- Consumes: The authorizer-protected `/api/evolution/targets` and `/api/evolve` routes and the anonymous, rate-limited `/api/scan` route.
- Produces: `createEvolutionApi(apiBase, fetchImpl)` with `listTargets(token, repoUrl)` and `start(token, request, signal)`; `createGraphScanClient(apiBase, fetchImpl)` with `scan(repoUrl, signal)` and no browser secret/token parameter.

- [ ] **Step 1: Add failing API-client and build-configuration tests**

Add tests asserting that Evolution sends `Authorization: Bearer <short-lived token>`, Graph scans send no authorization secret, API errors remain visible, and production configuration requires both `VITE_API_BASE` and `VITE_ANALYSIS_API_BASE` to be absolute HTTPS URLs.

- [ ] **Step 2: Run frontend tests and confirm failure**

Run from `admin-portal`: `npm test`

Expected: FAIL for the missing analysis clients/configuration.

- [ ] **Step 3: Implement the separate frontend flows**

Wire the Evolution tab to `EvolutionLive`, pass the in-memory Admin Portal token, add repository/target selection, preserve abort behavior, and clear Evolution view state on logout. Keep Graph Visualizer independent and unauthenticated; populate its non-secret analysis URL at Pages build time. Keep archive calls on `VITE_API_BASE` and `window.API_BASE`.

- [ ] **Step 4: Verify frontend behavior and production build**

Run from `admin-portal`: `npm test`, `npm run lint`, and `npm run build`.

Expected: all tests pass, no lint errors, and the production bundle contains only public API URLs—not tokens or credentials.

- [ ] **Step 5: Commit**

```text
feat: connect graph scans and admin evolution to analysis API
```
