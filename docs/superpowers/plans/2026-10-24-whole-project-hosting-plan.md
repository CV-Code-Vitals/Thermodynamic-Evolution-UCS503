# Whole-Project Hosting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Host the analysis API and Evolution workflow behind a stable HTTPS endpoint while keeping GitHub Pages and the working serverless archive unchanged, and isolating the Docker sandbox on a second private EC2 instance.

**Architecture:** Use a separate analysis stack with a REST API response-streaming Lambda proxy, a REST API Lambda authorizer, a non-VPC GitHub source stager, private S3 source/image buckets, an always-on analysis EC2 instance, and an on-demand private sandbox EC2 instance. The proxy stages GitHub sources, injects trusted server-side context, and streams Evolution SSE; the analysis service reads staged snapshots and uses an authenticated TLS broker for bounded ephemeral workspace evaluation.

**Tech Stack:** Go 1.25, Rust, Node.js Lambda response streaming, React 19/Vite 8, Docker Compose, AWS SAM/CloudFormation, API Gateway REST API, Lambda, EC2, SSM, S3, GitHub Actions OIDC.

**Spec:** `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`

## Global Constraints

- Keep the main website and Admin Portal UI on GitHub Pages.
- Leave the live `ucs503-admin-archive-v2` resources and deployment path unchanged.
- Keep both EC2 instances private; do not use NAT, public IPv4, a load balancer, VPC Link, Cloud Map, AWS Private CA, or paid interface endpoints by default.
- Use a REST API response-streaming Lambda proxy for SSE; keep Evolution runs below 15 minutes and send heartbeats more frequently than every five minutes.
- Use an S3 Gateway Endpoint for private S3 access and validate IPv6 reachability to the required AWS dual-stack control-plane endpoints before deployment.
- Use a bounded GitHub source stager and private S3 artifacts; private EC2 hosts must not clone from GitHub or pull from GHCR.
- Keep the analysis host free of Docker access; only the dedicated private sandbox host may access the Docker socket.
- Allow the sandbox broker only from the analysis service security group; use its TLS trust bundle and a random service token of at least 32 bytes.
- Keep the archive HMAC signing key readable only by the separate authorizer role; do not expose it to the proxy, either EC2 role, browser assets, URLs, or logs.
- The proxy may read only `/ucs503/analysis/api-token` to authenticate anonymous public-scan requests to the private Go service; it must never read the archive HMAC key. Evolution uses the token supplied through the validated authorizer context.
- Anonymous Graph Visualizer scans must enforce GitHub-only input, trusted API Gateway client identity, per-client rate limits, concurrency limits, and bounded repository, time, and engine resources.
- Do not deploy the legacy `ucs503-api` service or require RDS for initial analysis/Evolution hosting.
- Aim for $20/month; do not exceed $25/month without explicit additional approval. AWS credits do not count toward the cap.
- No AWS resource, secret, or GitHub setting may be created, updated, or deleted until the user separately approves the design, final itemized cost estimate, deployment plan, and then explicitly authorizes deployment.

## Review Focus

- **Untrusted repository URLs, redirects, DNS changes, and private/link-local targets:** source-stager and scan tests must reject non-allowlisted GitHub URLs and unsafe redirects before any fetch.
- **Archive bombs, traversal, symlinks, and excessive file counts:** stager, analysis extraction, and broker workspace tests must reject unsafe or over-limit archives and clean temporary data.
- **Forged client identity or spoofed internal headers:** proxy tests must overwrite client-supplied auth/IP headers; Go tests must accept only proxy-injected identity.
- **Invalid admin credentials, authorizer failures, or leaked secrets:** authorizer tests must reject missing, malformed, expired, tampered, and ambiguous bearer tokens and assert no signing key/token appears in response/log context.
- **SSE cancellation, stalled streams, and sandbox start/stop races:** proxy/backend/lifecycle tests must preserve progress and cancellation, release leases on every exit, and never stop an instance with active work.

---

### Task 1: Define and implement the REST API authorizer contract

**Files:**
- Modify: `analysis-authorizer/main.go`
- Modify: `analysis-authorizer/main_test.go`
- Modify: `analysis-authorizer/go.mod` only if the current AWS/Lambda dependencies are missing

**Interfaces:**
- Consumes: The existing 15-minute HMAC bearer token format in `api/auth.go`.
- Produces: A REST API Lambda `REQUEST` authorizer event and policy response with result caching disabled. A valid token returns an `Allow` policy for the exact `methodArn` and the scalar context value `analysisApiToken`; invalid credentials return `Deny` without reading the internal analysis token. The signing key remains `/ucs503/admin/token-hmac-key`, and the internal token remains `/ucs503/analysis/api-token`.

- [ ] **Step 1: Add failing REST-authorizer contract tests**

Add `TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer`, `TestAuthorizerHandler_DeniesInvalidBearer`, `TestAuthorizerHandler_RejectsAmbiguousAuthorizationHeaders`, and `TestAuthorizerHandler_PropagatesSecretStoreErrors`. Assert exact `methodArn` scope, no read of the internal token for invalid credentials, and no bearer/signing key in serialized output.

- [ ] **Step 2: Run authorizer tests to expose the current HTTP API contract**

Run from `analysis-authorizer`: `go test ./...`

Expected: FAIL for the REST API event/response contract; the current `isAuthorized` simple response is the HTTP API shape.

- [ ] **Step 3: Implement the REST `REQUEST` authorizer**

Change the event/response types and handler to return an IAM policy document (`execute-api:Invoke`, exact method ARN, `Allow` or `Deny`) plus only the approved internal token context after valid HMAC verification. Preserve SecureString validation and explicit SSM error propagation.

- [ ] **Step 4: Verify contract and static checks**

Run from `analysis-authorizer`: `go test ./...` and `go vet ./...`

Expected: PASS; invalid requests fail closed, valid policy scope is exact, and no secret material is serialized.

---

### Task 2: Add the bounded GitHub source stager

**Files:**
- Create: `source-stager/go.mod`
- Create: `source-stager/main.go`
- Create: `source-stager/main_test.go`
- Create: `source-stager/README.md`

**Interfaces:**
- Produces: `StageRepository(ctx context.Context, repositoryURL string) (snapshotID string, err error)`. IDs are fixed-length lowercase hexadecimal opaque identifiers; object keys are derived server-side as `snapshots/<id>/source.zip` and are never accepted from callers.
- Source limits: accept only public `https://github.com/{owner}/{repo}` URLs (optionally ending in `.git`); cap compressed bytes at 50 MiB, expanded bytes at 256 MiB, entries at 10,000, and individual files at 16 MiB. Reject symlinks, unsafe paths, redirects outside approved GitHub archive hosts, and private-address resolutions. Source objects expire after one day.

- [ ] **Step 1: Add failing URL, fetch, archive, and S3-store tests**

Add `TestValidateGitHubRepositoryURL_AcceptsPublicRepository`, `TestValidateGitHubRepositoryURL_RejectsUntrustedURL`, `TestStageRepository_RejectsUnsafeRedirect`, `TestStageRepository_RejectsPrivateAddress`, `TestStageRepository_EnforcesArchiveLimits`, and `TestStageRepository_PropagatesS3Failure`. Assert credentials, non-HTTPS, IP literals, lookalike hosts, extra path/query/fragment, non-GitHub redirects, private/link-local targets, oversized archives, ZIP traversal, symlinks, and excessive expanded size/count are rejected. Use an injectable HTTP client and fake object store.

- [ ] **Step 2: Run focused stager tests**

Run from `source-stager`: `go test ./...`

Expected: FAIL because the stager module is not implemented.

- [ ] **Step 3: Implement Lambda handler and bounded staging**

Fetch only validated public GitHub archives, enforce streaming byte limits before buffering, validate every archive entry before upload, write encrypted objects under the server-generated prefix, and return only the opaque ID plus expiry metadata. Keep the stager outside the VPC and do not expose it as a public endpoint.

- [ ] **Step 4: Verify stager tests and vet**

Run from `source-stager`: `go test ./...` and `go vet ./...`

Expected: PASS for both the accepted path and each invalid/oversize case.

---

### Task 3: Make analysis endpoints consume staged snapshots and enforce scan limits

**Files:**
- Modify: `go-backend/scanner.go`
- Modify: `go-backend/evolve.go`
- Modify: `go-backend/main.go`
- Modify: `go-backend/limiter.go`
- Modify: `go-backend/main_test.go`
- Modify: `go-backend/limiter_test.go`
- Create or modify: `go-backend/snapshot_store.go` and `go-backend/snapshot_store_test.go`

**Interfaces:**
- Consumes: The `snapshot_id` contract from Task 2 and trusted forwarding contract from Task 4.
- Produces: Internal `/api/scan` and `/api/evolve` requests that accept an opaque `snapshot_id`, never a server-local path or caller-supplied S3 key. The S3 reader may resolve only `snapshots/<validated-id>/source.zip`. Public scans use a server-injected client IP and an initial one-request-per-minute per-IP limit, a single concurrent scan, 50 MiB repository size, and the existing configured engine timeout.

- [ ] **Step 1: Add failing snapshot and admission tests**

Add `TestSnapshotStore_RejectsInvalidID`, `TestScanRepoHandler_UsesStagedSnapshot`, `TestPublicScanController_RateLimitsClient`, `TestPublicScanController_RejectsConcurrentScan`, `TestScanRepoHandler_IgnoresClientForwardedIP`, and `TestScanRepoHandler_CleansTemporaryWorkspace`. Assert invalid IDs/missing objects fail explicitly, valid snapshots extract temporarily, rate/concurrency limits return retryable `429`/`503`, client IP spoofing has no effect, and cleanup runs on success, failure, timeout, and cancellation.

- [ ] **Step 2: Run focused Go tests**

Run from `go-backend`: `go test ./...`

Expected: FAIL on the new staged-source contract and policy cases.

- [ ] **Step 3: Replace direct `git clone` and local-path request fields**

Read the bounded ZIP using the instance role's S3 permissions, safely extract to a unique temporary workspace, pass that workspace to the existing Rust engine/evolution logic, and remove the current `git clone` path from public scan handling. Keep the existing internal `ANALYSIS_API_TOKEN` middleware for proxy-to-EC2 authentication.

- [ ] **Step 4: Verify Go backend tests and race-sensitive admission behavior**

Run from `go-backend`: `go test ./...` and `go test -race ./...`

Expected: PASS; no public route can trigger arbitrary URL fetching or bypass the server-side limits.

---

### Task 4: Implement REST API proxying and response-streamed Evolution

**Files:**
- Create: `infra/analysis/stream-proxy/package.json`
- Create: `infra/analysis/stream-proxy/index.mjs`
- Create: `infra/analysis/stream-proxy/index.test.mjs`
- Modify: `go-backend/evolve.go`
- Modify: `go-backend/main_test.go`

**Interfaces:**
- Consumes: REST authorizer context from Task 1, stager invocation from Task 2, and backend `snapshot_id` request shape from Task 3.
- Produces: An API Gateway proxy for `POST /api/scan` and `POST /api/evolve`. For each request it validates bounded JSON, stages `repo_url`, replaces it with `snapshot_id`, strips/replaces client-supplied internal headers, injects `analysisApiToken` from validated authorizer context for Evolution, and reads the same narrowly scoped token from SSM for anonymous scans. It forwards trusted `requestContext.identity.sourceIp` for anonymous scans. Evolution is streamed; scans use a bounded non-streaming response.

- [ ] **Step 1: Add failing proxy and SSE tests**

Use a local fake stager/upstream and add `test('rejects oversized and malformed request bodies')`, `test('overwrites client-supplied internal headers')`, `test('uses authorizer token for Evolution and SSM token for anonymous scan')`, `test('preserves streamed SSE frames and status')`, `test('cancels upstream when client disconnects')`, and `test('does not log bodies or credentials')`. Assert trusted client IP comes only from API Gateway context and failed staging never reaches the analysis host.

- [ ] **Step 2: Run proxy tests**

Run from `infra/analysis/stream-proxy`: `npm test`

Expected: FAIL until the proxy handler and tests are added.

- [ ] **Step 3: Implement the Node.js streaming handler**

Use the Lambda response-streaming runtime interface, call the private source-stager function, forward only allowlisted headers, and connect to the fixed private analysis ENI address. Send correct API Gateway streaming response metadata, preserve upstream SSE chunks, and stop upstream work when the client disconnects. Do not log bodies, Authorization headers, or authorizer context.

- [ ] **Step 4: Add SSE heartbeat and bounded server runtime**

In `evolve.go`, emit periodic SSE comments while waiting for telemetry; ensure the run duration ceiling is less than 15 minutes and the heartbeat interval is less than five minutes. Keep context cancellation wired to the evaluator and admission lease.

- [ ] **Step 5: Verify proxy and backend stream behavior**

Run from `infra/analysis/stream-proxy`: `npm test`; from `go-backend`: `go test ./...`

Expected: PASS; frames reach the client incrementally, cancellation terminates upstream work, and server error events remain visible.

---

### Task 5: Complete the remote sandbox workspace and lifecycle contract

**Files:**
- Verify/modify: `go-backend/agent/workspace.go`
- Verify/modify: `go-backend/agent/workspace_test.go`
- Modify: `go-backend/agent/evaluator.go`
- Modify: `go-backend/agent/annealer.go`
- Modify: `go-backend/agent/broker_evaluator.go`
- Modify: `go-backend/agent/broker_test.go`
- Modify: `go-backend/broker.go`
- Modify: `go-backend/broker_workspace.go`
- Modify: `go-backend/main_test.go`

**Interfaces:**
- Consumes: The extracted staged repository path from Task 3 and private broker TLS/token configuration in `docker-compose.analysis.yml` and `docker-compose.sandbox.yml`.
- Produces: `RemoteWorkspaceEvaluator.PrepareWorkspace(ctx, repoPath) (workspaceID, error)` and `CleanupWorkspace(ctx, workspaceID) error`; authenticated broker routes `POST /v1/workspaces`, `DELETE /v1/workspaces/:id`, and `POST /v1/evaluate`. Workspace IDs are unguessable, archive limits are 50 MiB compressed / 256 MiB expanded / 10,000 entries / 16 MiB per file, and cleanup runs on all exits and TTL expiry.

- [ ] **Step 1: Add or finish boundary and lifecycle tests**

Add `TestExtractWorkspaceArchive_RejectsTraversal`, `TestExtractWorkspaceArchive_RejectsSymlinkEntries`, `TestBroker_RejectsMissingOrInvalidToken`, `TestRemoteBrokerEvaluator_RejectsUntrustedCA`, `TestBroker_RejectsUnknownWorkspace`, `TestBroker_CleansWorkspaceAfterFailure`, and `TestBroker_ClampsResourceLimits`. Assert malformed/oversize archives fail before evaluation, cleanup is retry-safe, and no shared host path or volume is required.

- [ ] **Step 2: Run focused broker tests**

Run from `go-backend`: `go test ./agent ./...`

Expected: FAIL for any incomplete workspace path or lifecycle behavior.

- [ ] **Step 3: Complete the authenticated cross-host protocol**

Create the archive from the staged local source, transfer it over HTTPS to the broker, evaluate only by opaque workspace ID, clean up on every result, and keep candidate limits enforced by the broker. Require the configured private CA bundle; never add insecure TLS fallback.

- [ ] **Step 4: Verify broker and candidate-container security**

Run from `go-backend`: `go test ./...`; validate both split Compose files with `docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config --quiet` and `docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config --quiet`.

Expected: PASS; only the sandbox service mounts Docker, while the analysis compose file has no Docker socket.

---

### Task 6: Add sandbox start, readiness, idle-stop, and crash recovery

**Files:**
- Modify: `go-backend/main.go`
- Create or modify: `go-backend/sandbox_lifecycle.go`
- Create or modify: `go-backend/sandbox_lifecycle_test.go`
- Modify: `go-backend/go.mod` and `go-backend/go.sum` only for the AWS SDK modules required
- Modify: `docker-compose.analysis.yml`

**Interfaces:**
- Produces: A lifecycle controller that starts only the configured sandbox instance when an authenticated Evolution request is admitted, waits for EC2 `running` and the authenticated broker health check, tracks active job/workspace leases, and stops only after 10 idle minutes with no active work.
- AWS permissions: Scope `StartInstances` and `StopInstances` to the sandbox instance ARN. Where EC2 requires `Resource: "*"`, constrain `DescribeInstances` with supported resource-tag conditions and test that condition. Give the watchdog a separate narrowly scoped role; the sandbox role cannot read the archive signing key. A watchdog reconciles stale leases after analysis process failure.

- [ ] **Step 1: Add failing lifecycle-controller tests**

With fake EC2 and broker clients, add `TestSandboxLifecycle_StartsBeforeEvolution`, `TestSandboxLifecycle_FailsWhenReadinessTimesOut`, `TestSandboxLifecycle_DoesNotStopWithActiveLease`, `TestSandboxLifecycle_StopsAfterTenIdleMinutes`, `TestSandboxLifecycle_NewLeaseCancelsIdleStop`, and `TestSandboxWatchdog_StopsAfterStaleLease`. Assert AWS errors remain explicit and concurrent requests cannot race a stop against an active job.

- [ ] **Step 2: Run lifecycle tests**

Run from `go-backend`: `go test ./...`

Expected: FAIL for missing controller behavior.

- [ ] **Step 3: Implement scoped lifecycle calls and lease recovery**

Start the sandbox only after admin authorization, wait with a bounded context, fail with a retryable response if readiness fails, and use a watchdog/lease timestamp to stop after inactivity even if the analysis process restarts. Keep the sandbox private and do not stop while active jobs or workspaces exist.

- [ ] **Step 4: Verify lifecycle and race tests**

Run from `go-backend`: `go test -race ./...`

Expected: PASS; concurrent requests cannot create start/stop races or stop an active worker.

---

### Task 7: Restore the Admin Evolution view and route Graph scans to the API

**Files:**
- Modify: `admin-portal/src/components/AdminPortal.jsx`
- Modify: `admin-portal/src/components/EvolutionLive.jsx`
- Modify: `admin-portal/src/components/EvolutionLive.css`
- Create or modify: `admin-portal/src/evolution-api.js`
- Create or modify: `admin-portal/tests/evolution-api.test.js`
- Modify: `graph-visualizer/index.html`
- Create: `tests/graph-visualizer-config.test.mjs`
- Modify: `.github/workflows/deploy-frontend.yml`

**Interfaces:**
- Produces: `VITE_ANALYSIS_API_BASE`, separate from the unchanged archive `VITE_API_BASE`. Evolution accepts a public GitHub repository URL and target file/function, sends `Authorization: Bearer <in-memory admin token>`, and consumes SSE with visible progress, completion, error, and abort states. The Graph Visualizer makes anonymous calls only to the configured `/api/scan` API.
- Graph Visualizer config: the Pages workflow writes `_site/graph-visualizer/config.js` with the public endpoint in `window.UCS503_ANALYSIS_API_BASE`; `graph-visualizer/index.html` loads it before its scan code. The value is a public URL, never a credential.

- [ ] **Step 1: Add failing API adapter tests**

Add `test('keeps archive and analysis API bases separate')`, `test('sends the in-memory bearer only to Evolution')`, `test('surfaces HTTP and SSE errors')`, `test('aborts the Evolution stream')`, and static `TestGraphVisualizer_UsesConfiguredAnalysisAPIWithoutFallback`. Assert no internal service token exists in request headers or built assets and a missing Graph API configuration produces an explicit UI error.

- [ ] **Step 2: Run frontend tests to confirm missing behavior**

Run from `admin-portal`: `npm test`

Expected: FAIL for missing analysis API adapter behavior.

- [ ] **Step 3: Re-enable Evolution inside the authenticated portal**

Replace the unavailable placeholder with the Evolution component; pass the existing in-memory token as a prop; provide repository, file, and function inputs; use the separate analysis API base; and retain SSE progress and user cancellation. Never persist the token in localStorage or put any service token into Vite variables.

- [ ] **Step 4: Remove the scan fallback and make endpoint configuration explicit**

Change Graph Visualizer to use `window.UCS503_ANALYSIS_API_BASE` for scans, remove the localStorage `CUSTOM_SCAN_API` and client-side GitHub scan fallback for deployment, and show a visible error when the backend is unavailable. Keep local development supported through explicit local configuration. Add a static test that catches a missing config script or a reintroduced fallback.

- [ ] **Step 5: Verify frontend tests, lint, and production build**

Run from `admin-portal`: `npm test`, `npm run lint`, and `npm run build`.

Expected: PASS; production assets contain the public API base only and no service credential.

---

### Task 8: Create the isolated analysis infrastructure stack

**Files:**
- Create: `infra/analysis/template.yaml`
- Create: `infra/analysis/samconfig.toml`
- Create: `infra/analysis/parameters.example.json`
- Create: `infra/analysis/README.md`
- Create: `tests/analysis-template.test.mjs`

**Interfaces:**
- Consumes: Authorizer, proxy, source-stager, lifecycle, and frontend contracts from Tasks 1–7.
- Produces: A separate `ucs503-analysis-v1` stack containing private networking, S3 Gateway Endpoint, private analysis/sandbox instances and separate roles, API Gateway REST API, response-streaming proxy, authorizer, source-stager, lifecycle watchdog, private encrypted source/artifact buckets with lifecycle rules, SSM managed access, least-privilege policies, and retained operational logs. Source objects expire after one day; deployment image bundles expire after 30 days. Persist analysis outputs on a separate encrypted gp3 data volume with daily snapshots retained for seven days; document restore steps and include volume/snapshot costs. It must not import, modify, or deploy the archive template.

- [ ] **Step 1: Add template policy tests**

Add named Node tests `test('creates two private instances with split roles')`, `test('blocks public and SSH ingress')`, `test('limits token permissions by role')`, `test('keeps analysis output snapshots for seven days')`, `test('expires sources after one day and bundles after 30 days')`, `test('configures anonymous scan and protected streaming Evolution routes')`, and `test('omits credentials from logs and archive resources')`. Assert the complete infrastructure policy described above, including the proxy's internal-token-only SSM access and zero authorizer cache TTL.

- [ ] **Step 2: Run template validation**

Run from the repository root: `sam validate --lint --template-file infra/analysis/template.yaml`

Expected: FAIL until the template is created.

- [ ] **Step 3: Implement the independent stack**

Use the chosen REST API plus Lambda response-streaming proxy. Keep instances private; use a fixed private analysis ENI, dedicated broker-only security group ingress, S3 Gateway Endpoint, and validated dual-stack AWS control-plane egress through an egress-only Internet Gateway. Add a non-VPC stager and a separate private artifact bucket. Store the broker certificate/key outside source control in the approved runtime secret-delivery path and distribute only the dedicated trust bundle to the analysis role. Do not add NAT, VPC Link, ALB/NLB, Cloud Map, public IPv4, RDS, or AWS Private CA.

- [ ] **Step 4: Verify template policy and syntax**

Run from the repository root: `node --test tests/analysis-template.test.mjs` and `sam validate --lint --template-file infra/analysis/template.yaml`.

Expected: PASS; no deployment command is run.

---

### Task 9: Add immutable image artifact delivery and deployment safeguards

**Files:**
- Modify: `.github/workflows/publish-api-image.yml`
- Create: `.github/workflows/deploy-analysis.yml`
- Create: `tests/analysis-workflow.test.mjs`
- Modify: `README_DEPLOY.md`

**Interfaces:**
- Produces: The existing push workflow publishes immutable SHA-addressed analysis and runner images to GHCR only. A separate `workflow_dispatch` workflow accepts a SHA, downloads those exact images on the GitHub runner, uploads an image bundle to the private S3 artifact bucket via short-lived GitHub OIDC credentials, and deploys through SSM only when both `enable_deployment` is true and `vars.ANALYSIS_DEPLOY_ENABLED == 'true'`, with a protected environment approval. GHCR is not used by private hosts at runtime. The existing archive deployment workflow remains unchanged.

- [ ] **Step 1: Add workflow policy tests**

Add `test('publishes immutable SHA image tags')`, `test('gates AWS upload and deploy behind manual opt-in')`, and `test('uses OIDC and SSM without SSH secrets')`. Assert no `latest`, ordinary pushes cannot mutate AWS, both deployment gates default false, protected-environment approval is required, and EC2 never pulls from GHCR.

- [ ] **Step 2: Implement build/artifact workflow and gated deployment**

Keep the existing `publish-api-image.yml` GHCR build push-triggered and SHA-tagged, widening its path filters for relevant backend/runner changes. The new manually dispatched workflow must download only the selected SHA, save the image bundle, and upload/deploy only after both gates and protected-environment approval. Keep AWS role IDs and endpoint names in repository variables/configuration, not source files or image layers. Do not configure GitHub variables or environment protection in this task; those settings require the separate user authorization.

- [ ] **Step 3: Document operations and rollback**

Document image promotion and rollback, SSM access, token/CA rotation, health checks, log retention, alarms, analysis output retention/backup/restore, recovery within approximately one hour, sandbox lease recovery, cost monitoring, and the explicit approval gate. State that `docker-compose.ec2.yml` and `ucs503-api` are legacy and not used.

- [ ] **Step 4: Validate workflows and deployment configuration**

Run from the repository root: `node --test tests/analysis-workflow.test.mjs`, `docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config --quiet`, and `docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config --quiet`.

Expected: PASS; workflows can build and stage artifacts but cannot deploy without explicit protected approval.

---

### Task 10: Complete read-only regional validation and cost gate

**Files:**
- Read: `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`
- Update: `infra/analysis/README.md` with validated assumptions and operating commands

**Interfaces:**
- Produces: A user-reviewed `ap-south-1` itemized monthly estimate, representative measured instance sizes, validated response-streaming and IPv6 endpoint reachability, validated source/image S3 flows, and a deployment checklist. No AWS resource or GitHub setting is changed in this task.

- [ ] **Step 1: Run local builds, tests, and static validation**

Run these commands from the indicated working directories:

- `go-backend`: `go test ./...`; `go vet ./...`; `go test -race ./...`
- `analysis-authorizer`: `go test ./...`; `go vet ./...`
- `source-stager`: `go test ./...`; `go vet ./...`
- `thermodynamic-ast-engine`: `cargo test --locked`
- `admin-portal`: `npm ci`; `npm test`; `npm run lint`; `npm run build`
- Repository root: `node --test tests/analysis-template.test.mjs tests/analysis-workflow.test.mjs tests/graph-visualizer-config.test.mjs`
- Repository root: `sam validate --lint --template-file infra/analysis/template.yaml`
- Repository root: `docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config --quiet`
- Repository root: `docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config --quiet`

Expected: PASS for all commands before any deployment approval is requested.

- [ ] **Step 2: Validate costs and runtime sizing without mutating AWS**

Use current `ap-south-1` pricing and resource measurements to include EC2 uptime, EBS, API Gateway, Lambda, S3, CloudWatch, backups, image/source transfer, and all endpoints. Confirm REST API/Lambda response streaming and private-subnet IPv6 access to the required SSM, EC2, and Lambda dual-stack endpoints. Stop and present costed alternatives if validation fails; do not silently add NAT or paid endpoints.

- [ ] **Step 3: Present a separate final approval request**

Present the architecture, measured capacity, exact itemized cost (target $20/month, hard cap $25/month without approval), security/availability trade-offs, and deployment/rollback commands. Request explicit authorization to deploy separately from cost approval. Do not run `sam deploy`, `aws cloudformation deploy`, `aws ssm send-command`, alter GitHub settings, or mutate AWS resources in this plan.

---

## Superseded Draft Plans

The following drafts are historical only; this plan is authoritative because the approved design now uses a REST API response-streaming Lambda proxy, REST authorizer policy responses, S3 source/image staging, and the explicit pre-deployment cost/authorization gate:

- `2026-10-24-analysis-and-ui-plan.md`
- `2026-10-24-aws-hosting-and-deployment-plan.md`
- `2026-10-24-sandbox-worker-plan.md`

Do not execute the superseded task instructions independently. Their implementation details must be reconciled with this plan before work begins.
