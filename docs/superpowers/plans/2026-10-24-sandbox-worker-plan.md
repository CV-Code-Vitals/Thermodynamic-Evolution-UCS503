> **Superseded draft:** Do not execute this plan independently. The authoritative plan is [2026-10-24-whole-project-hosting-plan.md](2026-10-24-whole-project-hosting-plan.md), which integrates this worker's workspace contract with staged source input, the REST streaming API, and the approved sandbox lifecycle.

# Sandbox Worker Isolation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make candidate evaluation work safely when the analysis API and Docker broker run on separate EC2 instances, without shared host paths or EFS.

**Architecture:** The analysis process creates one bounded repository snapshot per Evolution session and uploads it to the private authenticated broker. The broker stores it in a temporary, random workspace, evaluates candidates using the existing isolated runner, and removes it on completion or TTL expiry.

**Tech Stack:** Go 1.25, Docker Engine, Docker Compose, ZIP archives, existing Go test suite and mandatory Docker integration test.

**Spec:** `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`

## Global Constraints

- Analysis EC2 never mounts the Docker socket.
- The broker and Docker socket exist only on the dedicated private sandbox EC2.
- Broker requests require the server-to-server sandbox token, use verified TLS with no plaintext fallback, and are reachable only from the analysis security group.
- Candidate containers remain non-root, no-network, read-only-root, capability-dropped, resource-limited, and ephemeral.
- The two hosts do not share named Docker volumes, host directories, or EFS.
- Repository snapshots are temporary, bounded, traversal-safe, and removed on completion, cancellation, failure, or TTL expiry.
- Do not deploy the old `ucs503-api` service or require RDS.
- Do not change AWS resources until separate approval.

## Review Focus

- Malformed, traversal, symlink, or oversized workspace archives must never escape the temporary workspace.
- Missing or invalid broker credentials must fail closed before workspace creation or evaluation.
- Duplicate, expired, or unknown workspace IDs must not access another evaluation's files.
- Failed evaluations and client disconnects must release concurrency slots and clean workspace data.
- Candidate limits must remain clamped server-side even if the analysis process sends larger values.

---

### Task 1: Add bounded workspace snapshot creation and extraction

**Files:**
- Create: `go-backend/agent/workspace.go`
- Create: `go-backend/agent/workspace_test.go`
- Modify: `go-backend/agent/evaluator.go`

**Interfaces:**
- Produces: `CreateWorkspaceArchive(repoPath string, maxArchiveBytes int64) ([]byte, error)` and `ExtractWorkspaceArchive(archive []byte, destination string, limits WorkspaceArchiveLimits) error`.
- Initial limits: 50 MiB compressed snapshot, matching the existing default upload limit; 256 MiB expanded content; at most 10,000 entries; at most 16 MiB per file. Exclude `.git`, reject symlinks and unsafe paths, and use the existing Go extraction limits as the upper safety boundary.

- [ ] **Step 1: Add failing snapshot/extraction tests**

Add `TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles`, `TestCreateWorkspaceArchive_RejectsOversizeFile`, `TestExtractWorkspaceArchive_RejectsTraversal`, `TestExtractWorkspaceArchive_RejectsSymlinkEntries`, and `TestExtractWorkspaceArchive_EnforcesExpandedLimits`. Assert extracted paths remain inside the destination and input bytes are not written elsewhere.

- [ ] **Step 2: Run the focused tests and confirm failure**

Run from `go-backend`: `go test ./agent -run 'Test(Create|Extract)WorkspaceArchive' -v`

Expected: FAIL because the workspace archive helpers are missing.

- [ ] **Step 3: Implement bounded archive helpers**

Create a deterministic ZIP snapshot of repository files while excluding `.git`; reject symlinks, path escapes, unsupported file types, and size/count overages. Extract only after validating each entry and enforce limits before writing bytes.

- [ ] **Step 4: Verify archive boundary tests**

Run from `go-backend`: `go test ./agent -run 'Test(Create|Extract)WorkspaceArchive' -v`

Expected: PASS, including the 50 MiB compressed, 256 MiB expanded, 10,000-entry, and 16 MiB per-file boundary cases.

- [ ] **Step 5: Commit**

```text
feat: add bounded sandbox workspace archives
```

### Task 2: Add remote workspace lifecycle to the broker protocol

**Files:**
- Modify: `go-backend/agent/evaluator.go`
- Modify: `go-backend/agent/broker_evaluator.go`
- Modify: `go-backend/agent/broker_test.go`
- Modify: `go-backend/agent/annealer.go`
- Modify: `go-backend/broker.go`
- Modify: `go-backend/main_test.go`

**Interfaces:**
- Produces: `RemoteWorkspaceEvaluator` with `PrepareWorkspace(ctx context.Context, repoPath string) (string, error)` and `CleanupWorkspace(ctx context.Context, workspaceID string) error`.
- Produces broker routes `POST /v1/workspaces` (raw `application/zip` body), `DELETE /v1/workspaces/:id`, and `POST /v1/evaluate` accepting a `workspace_id`. Workspace IDs are cryptographically random; every lifecycle request also requires the broker token. The client verifies the broker TLS certificate; do not add an insecure verification bypass.

- [ ] **Step 1: Add failing lifecycle tests**

Add `TestRemoteBrokerEvaluator_PrepareWorkspace`, `TestRemoteBrokerEvaluator_CleanupWorkspace`, `TestBroker_RejectsWorkspaceRequestsWithoutToken`, `TestBroker_RejectsUnknownWorkspace`, and `TestBroker_CleansWorkspaceAfterExpiry`. Assert the workspace is created once per Evolution session, reused for all candidate evaluations, and deleted on completion.

- [ ] **Step 2: Run the focused tests and confirm failure**

Run from `go-backend`: `go test ./agent -run 'TestRemoteBrokerEvaluator_(PrepareWorkspace|CleanupWorkspace)' -v` and `go test . -run 'TestBroker_' -v`

Expected: FAIL because workspace lifecycle routes and client methods are missing.

- [ ] **Step 3: Implement authenticated, ephemeral workspace lifecycle**

Require `X-Sandbox-Token` for create/evaluate/delete over verified HTTPS. Stream a raw ZIP body capped at 50 MiB, extract into a broker-local temporary directory, return an opaque ID, and enforce a 30-minute TTL cleanup sweep. Evaluation must use the broker-local extracted path, never the analysis host's `RepoPath`. Add bounded cleanup with a short background deadline when a client cancels.

- [ ] **Step 4: Verify lifecycle and fail-closed behavior**

Run from `go-backend`: `go test ./...` and `go vet ./...`

Expected: PASS; remote evaluation works without shared mounts, the 50 MiB/256 MiB/entry/file limits are enforced, and every error/cancellation path releases admission capacity and cleans ephemeral workspaces.

- [ ] **Step 5: Commit**

```text
feat: transfer ephemeral workspaces to sandbox broker
```

### Task 3: Split analysis and sandbox deployment configuration

**Files:**
- Create: `docker-compose.analysis.yml`
- Create: `docker-compose.sandbox.yml`
- Create: `.env.analysis.example`
- Create: `.env.sandbox.example`
- Create: `scripts/compose-config.test.mjs`
- Modify: `docker-compose.ec2.yml`

**Interfaces:**
- Analysis Compose runs only the Go API/Rust engine and has no Docker socket mount or broker service.
- Sandbox Compose runs only the broker, mounts the local Docker socket, stores only temporary workspaces, and exposes no host-published port.
- Both Compose files deploy immutable commit-SHA image tags; no `latest` tag is used for deployment.

- [ ] **Step 1: Add Compose validation fixtures**

Add `scripts/compose-config.test.mjs` using Node's built-in test runner. Assert the analysis configuration has no Docker socket, public port binding, or `ucs503-api`; assert the sandbox configuration has no public port binding and only the required broker socket mount.

- [ ] **Step 2: Verify the current configuration fails the new checks**

Run from the repository root: `node --test scripts/compose-config.test.mjs`

Expected: FAIL because the split Compose files do not yet exist; the test must also ensure the legacy Compose is not the new deployment target.

- [ ] **Step 3: Split the Compose files**

Move the analysis service and broker into separate configurations, remove the duplicate `ucs503-api`, remove cross-host named-volume assumptions, use SSM-delivered runtime secrets, and keep host-published ports disabled. Leave `docker-compose.ec2.yml` as a documented legacy/manual path until the replacement workflow is approved.

- [ ] **Step 4: Validate both service configurations**

Run from the repository root:

```text
docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config
docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config
node --test scripts/compose-config.test.mjs
```

Expected: Both configurations are valid; no sample file contains real secrets or public broker bindings.

- [ ] **Step 5: Commit**

```text
feat: split analysis and sandbox compose deployments
```

### Task 4: Preserve the mandatory runner security gate

**Files:**
- Modify: `.github/workflows/ci-go.yml`
- Modify: `.github/workflows/publish-api-image.yml`
- Modify: `go-backend/agent/broker_test.go`

- [ ] **Step 1: Add image/protocol regression checks**

Assert that the candidate runner remains non-root, the evaluator retains `--network none`, `--read-only`, `--cap-drop ALL`, `no-new-privileges`, and all CPU/memory/PID/output limits, and deployed image references resolve to a commit SHA.

- [ ] **Step 2: Run Go and Docker validation**

Run from `go-backend`: `go test ./...` and `go vet ./...`.
Run from the repository root: `docker build -t ucs503-sandbox-runner:test -f docker/sandbox-runner/Dockerfile .`

Expected: PASS; CI's existing mandatory `TestDockerEvaluator_LiveExecution` remains required and may not silently skip.

- [ ] **Step 3: Commit**

```text
test: enforce sandbox isolation and image pinning
```
