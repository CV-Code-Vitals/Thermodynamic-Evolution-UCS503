# Security Findings Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remediate the confirmed security and operational gaps from the supplied review without configuring AWS or deploying.

**Architecture:** Keep the existing archive and analysis interfaces, adding fail-fast engine validation, KMS alias conditions, rootless Docker socket use, bounded bearer renewal, and checked deployment rollout. Protect each change with focused regression tests and keep AWS and GitHub-side setup as documented prerequisites only.

**Tech Stack:** Go, React/Vite, Node test runner, Docker Compose, AWS SAM/CloudFormation, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-10-25-security-findings-remediation-design.md`

## Global Constraints

- No AWS resource, secret, or GitHub setting may be created, updated, or deleted.
- Do not run `sam deploy`, dispatch a deployment workflow, send SSM commands, commit, or push.
- Do not edit unrelated user changes or regenerate `api/ucs503-engine-api.exe`.
- Keep the archive API exact-origin CORS behavior and in-memory bearer storage.
- Rootless Docker AMI/bootstrap preparation is an operator prerequisite; do not build or provision an AMI.
- Tokens may be renewed only while valid and no later than eight hours after issue.
- Preserve verification failures explicitly; do not convert rollout failures to success.

## Review Focus

- **Invalid/missing/non-executable engine paths:** startup must exit before listening; Task 1 tests regular-file and executable checks.
- **Legacy and renewed token boundaries:** legacy tokens verify but cannot renew; new tokens renew only before expiry and until the eight-hour maximum; Task 4 tests each boundary.
- **Rootless socket/host path mismatch:** Compose and deployment must use the same dedicated UID's socket and workspace-visible paths; Task 3 tests configuration and documents host prerequisites.
- **Stopped sandbox rollout or failed post-check:** the workflow must start a stopped host only for deployment, restore its initial stopped state even on failure, and reject image/health mismatches; Task 5 tests the safety gates.
- **KMS alias and parameter-context mismatch:** all decrypt permissions remain limited to alias/aws/ssm and their existing SSM parameter contexts; Task 2 tests each role.

---

### Task 1: Fail fast on invalid analysis engine binaries

**Files:**
- Modify: `go-backend/main.go`
- Test: `go-backend/main_test.go`

**Interfaces:**
- Produces: `validateEngineBinary(path string) error`, validating that the path exists, is a regular file, and is executable on non-Windows systems.

- [ ] **Step 1: Write failing validation tests**

Add `TestValidateEngineBinary_RejectsMissingPath`, `TestValidateEngineBinary_RejectsDirectory`, and `TestValidateEngineBinary_RejectsNonExecutableFile` (skip the permission-specific case on Windows). Assert an executable regular file succeeds.

- [ ] **Step 2: Run focused tests and verify expected failures**

Run from `go-backend`: `go test . -run '^TestValidateEngineBinary_'`

Expected: FAIL because the validation helper is not implemented.

- [ ] **Step 3: Implement validation and fatal startup behavior**

Call `validateEngineBinary` before initializing stores or serving requests. On failure, terminate with an actionable message naming `ENGINE_PATH`; do not log a warning and continue.

- [ ] **Step 4: Verify**

Run from `go-backend`: `go test ./...` and `go vet ./...`.

Expected: PASS.

### Task 2: Scope SSM decryption to the AWS-managed SSM key alias

**Files:**
- Modify: `template.yaml`
- Modify: `infra/analysis/template.yaml`
- Test: `tests/analysis-template.test.mjs`
- Test: existing root SAM template tests, if present

**Interfaces:**
- Preserves: existing `kms:ViaService` and parameter ARN encryption-context conditions.
- Produces: each `kms:Decrypt` statement used for SSM SecureStrings additionally requires `kms:ResourceAliases` to include exactly `alias/aws/ssm`.

- [ ] **Step 1: Add failing policy assertions**

Add tests that inspect every decrypt statement in both templates and assert the SSM alias restriction and existing parameter-context restriction are present.

- [ ] **Step 2: Run tests and verify they fail**

Run: `node --test tests/analysis-template.test.mjs`

Expected: FAIL because decrypt statements do not yet constrain the key alias.

- [ ] **Step 3: Add alias conditions without changing parameter names**

Add `ForAnyValue:StringEquals: { kms:ResourceAliases: alias/aws/ssm }` to each SSM decrypt statement. Preserve every existing `kms:ViaService` and `kms:EncryptionContext:PARAMETER_ARN` condition.

- [ ] **Step 4: Verify policy tests and SAM lint**

Run: `node --test tests/analysis-template.test.mjs` and `sam validate --lint --template-file infra/analysis/template.yaml`.

Expected: PASS. If the policy condition is invalid for the SSM managed key, stop and report the blocker; do not broaden the policy or add a customer-managed key.

### Task 3: Move sandbox broker to the rootless Docker socket

**Files:**
- Modify: `docker-compose.sandbox.yml`
- Modify: `.env.sandbox.example`
- Modify: `.github/workflows/deploy-analysis.yml`
- Modify: `infra/analysis/README.md`
- Modify: `README_DEPLOY.md`
- Test: `tests/analysis-template.test.mjs` or a focused Compose-policy test

**Interfaces:**
- Consumes: host rootless Docker daemon for UID 10002 at `/run/user/10002/docker.sock`.
- Produces: broker container `DOCKER_HOST=unix:///run/docker.sock`, with the required host socket path mounted at `/run/docker.sock`; no rootful `/var/run/docker.sock` or Docker group ID contract.

- [ ] **Step 1: Add failing rootless Compose policy assertions**

Assert the broker references the required socket source and `DOCKER_HOST`, removes `DOCKER_GID` and `/var/run/docker.sock`, and documents the matching UID and rootless daemon prerequisites.

- [ ] **Step 2: Run the focused policy test and verify failure**

Run: `node --test tests/analysis-template.test.mjs`

Expected: FAIL because Compose currently mounts the rootful socket.

- [ ] **Step 3: Update Compose and the example environment**

Replace the rootful socket mount and group addition with `DOCKER_SOCKET_SOURCE` default/example `/run/user/10002/docker.sock`; set `DOCKER_HOST` to the broker container's mounted socket.

- [ ] **Step 4: Align deployment commands**

For sandbox-host SSM commands, set the host Docker CLI endpoint to `unix:///run/user/10002/docker.sock`; leave analysis-host Docker commands unchanged.

- [ ] **Step 5: Document host prerequisites and verify**

Document rootless daemon installation, the dedicated UID, systemd linger, socket permissions, Docker Compose availability, and workspace path visibility. Run the policy test and:

```powershell
docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config --quiet
```

Expected: PASS; the operator must prepare the AMI separately before future deployment.

### Task 4: Add bounded admin bearer renewal

**Files:**
- Modify: `api/auth.go`
- Test: `api/auth_test.go`
- Modify: `api/router.go`
- Test: `api/router_test.go` or existing auth route tests
- Modify: `admin-portal/src/archive-api.js`
- Test: `admin-portal/tests/archive-api.test.js`
- Modify: `admin-portal/src/components/AdminPortal.jsx`

**Interfaces:**
- Produces: a `TokenRenewer` interface implemented by `HMACTokenAuthenticator` with `Renew(token string, now time.Time) (string, time.Time, error)`.
- Produces: `POST /api/refresh` accepting an existing bearer token and returning `{status:"authenticated", token, expiresAt}` only when renewal is allowed.
- Preserves: verification of valid legacy 40-byte payloads; legacy tokens cannot renew and must reauthenticate.

- [ ] **Step 1: Write failing token renewal tests**

Test new-token renewal, invalid/expired token rejection, legacy token verification without renewal, and refusal after eight hours from issuance.

- [ ] **Step 2: Run focused Go tests and verify failure**

Run from `api`: `go test . -run 'Test(HMACTokenAuthenticator|RefreshHandler)'`

Expected: FAIL because renewal and the route do not exist.

- [ ] **Step 3: Implement a versioned token payload**

Issue new tokens with issue time and expiry in the signed payload. Keep `Verify` compatible with valid legacy payloads. Renew only new-format, currently valid tokens and cap expiry at eight hours after issue.

- [ ] **Step 4: Implement the authenticated refresh route**

Require POST and the configured trusted Origin. Accept only an unambiguous valid bearer; return generic unauthorized errors for invalid, expired, legacy, or over-age renewal attempts. Do not accept a passkey on the renewal route.

- [ ] **Step 5: Add frontend refresh API and session scheduling**

Add `archiveApi.refresh(token)` response validation. Track `expiresAt` in component state and renew shortly before expiry; on renewal failure clear the bearer and show a re-login message. Keep tokens memory-only.

- [ ] **Step 6: Verify Go and frontend tests**

Run from `api`: `go test ./...` and `go vet ./...`; from `admin-portal`: `npm test`, `npm run lint`, and `npm run build`.

Expected: PASS; no token is written to localStorage or sessionStorage.

### Task 5: Verify deployment image identity and safely handle stopped sandbox

**Files:**
- Modify: `.github/workflows/deploy-analysis.yml`
- Modify: `tests/analysis-workflow.test.mjs`
- Modify: `infra/analysis/README.md`
- Modify: `README_DEPLOY.md`

**Interfaces:**
- Consumes: exact `docker image inspect --format '{{.Id}}'` IDs for the selected analysis and runner SHA tags on the GitHub runner.
- Produces: successful deployment requires each host's running broker/backend image ID to match the selected analysis image ID; the sandbox host must have the selected runner image ID loaded; analysis must pass `/health`.

- [ ] **Step 1: Add failing workflow policy tests**

Assert image IDs are captured and checked remotely, analysis health is checked, a stopped sandbox is started only for the deployment, and an initially stopped sandbox is stopped again through failure cleanup. Assert the workflow requires the documented `ec2:DescribeInstances`, `ec2:StartInstances`, and `ec2:StopInstances` permissions for only the sandbox instance where supported.

- [ ] **Step 2: Run focused workflow tests and verify failure**

Run: `node --test tests/analysis-workflow.test.mjs`

Expected: FAIL because the workflow currently checks only SSM command success and prints Compose status.

- [ ] **Step 3: Capture expected image IDs and add remote checks**

Capture IDs after pulling the immutable SHA tags. After Compose starts, compare the service container image ID to the expected ID. On the analysis host, poll its configured healthcheck with a bounded timeout and fail on unhealthy/missing status.

- [ ] **Step 4: Preserve sandbox initial power state**

Before sandbox SSM deployment, read its EC2 state. If stopped, start it and wait for SSM `Online`. Install an EXIT cleanup that stops it if and only if this workflow started it, including deploy/check failure paths. Leave an initially running instance running.

- [ ] **Step 5: Update role prerequisites and verify**

Document the future protected OIDC role's narrowly scoped sandbox start/stop and SSM status/command permissions. Run `node --test tests/analysis-workflow.test.mjs`; do not configure the role or invoke the workflow.

Expected: PASS with every mismatch or failed health probe producing a nonzero workflow result.

### Task 6: Run complete local verification

**Files:**
- Verify all files changed by Tasks 1–5; no new implementation scope.

- [ ] **Step 1: Run repository policy tests**

Run: `node --test tests/analysis-template.test.mjs tests/analysis-workflow.test.mjs tests/graph-visualizer-config.test.mjs`.

- [ ] **Step 2: Run SAM and Compose validation**

Run: `sam validate --lint --template-file infra/analysis/template.yaml`, and both analysis and sandbox Compose `config --quiet` commands using their example env files.

- [ ] **Step 3: Run impacted language suites**

Run Go tests and vet in `api` and `go-backend`, plus `npm test`, `npm run lint`, and `npm run build` in `admin-portal`.

- [ ] **Step 4: Inspect final diff**

Run `git diff --check` and inspect status/diff, preserving unrelated modifications and confirming no AWS calls, workflow dispatch, commit, or push occurred.

Expected: all targeted local checks pass; any blocked test or external prerequisite is listed explicitly.
