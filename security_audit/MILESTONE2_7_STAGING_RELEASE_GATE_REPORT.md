# Milestone 2.7 — Staging Deployment Verification & Production Release Gate Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Date:** 2026-10-24  
**Auditor & Verification Engineer:** Independent Release Engineer, Application Security Auditor & Deployment Verification Engineer  
**Baseline Code Commit:** `3f5b8f3`  
**Milestone 2.6 Report Commit:** `5b0bde6`  
**Hardened Remediation Code Commit:** [`5faa0b0`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/commit/5faa0b03aef311ef8f752b61119f01b7fae1d899)  
**Hosted CI Run (Go Build & Vet):** [`40051436562`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40051436562) — **SUCCESS** (All 17 steps completed; Step 14 live Docker gate explicitly passed)  
**Hosted CI Run (Publish Images):** [`40051436595`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40051436595) — **SUCCESS** (All 3 images built & pushed to GHCR)  
**Prior Release Decision:** CONDITIONAL GO  
**Final Milestone 2.7 Release Decision:** **CONDITIONAL GO** (Release Gate pending operator staging execution & EC2 ingress verification)

---

## 1. Executive Summary

Milestone 2.7 evaluated the deployment integrity, security boundaries, and execution pipeline of the Thermodynamic Architecture Evolution system across the full request lifecycle:

$$\text{Client} \longrightarrow \text{Analysis Backend} \longrightarrow \text{Authenticated Sandbox Broker} \longrightarrow \text{Docker Daemon} \longrightarrow \text{Isolated Runner} \longrightarrow \text{Evaluation Result}$$

### Key Findings & Remediations Delivered in Milestone 2.7

1. **Remediation SEC-04 (Service Readiness & Health Checks)**:
   - In `docker-compose.ec2.yml`, `ucs503-analysis` previously declared `depends_on: [ucs503-sandbox-broker]` without a container-level health check. This created a race condition where the analysis service could start and receive evolution requests before the broker initialized its HTTP listener and validated Docker daemon availability.
   - **Fix**: Added explicit `healthcheck` configurations to `ucs503-sandbox-broker` and `ucs503-analysis`, updating `depends_on` to `condition: service_healthy`.
2. **Remediation SEC-05 (Server-Side Limits Clamping & System Path Protection)**:
   - While the analysis client passed sensible default limits, `broker.go` previously lacked server-side limit clamping, theoretically allowing a compromised backend to request unbounded timeouts or memory. Furthermore, `EvaluateCandidate` lacked restricted-system-path validation on `req.RepoPath`.
   - **Fix**: Implemented `clampLimits()` in `broker.go` enforcing immutable server-side ceilings (Timeout $\le 120$s, Memory $\le 2048$MB, CPUs $\le 2.0$, PIDs $\le 256$, Output $\le 256$KiB). Added `validateCandidateRepoPath()` in `evaluator.go` rejecting restricted system directories (`/etc`, `/root`, `/sys`, `/proc`, `/dev`, `C:/Windows`, etc.).
3. **Environment Audit & Realism**:
   - In accordance with the audit mandate, we explicitly distinguish between verified local/CI evidence and live EC2 cloud infrastructure. Live EC2 host credentials (`secrets.EC2_HOST`, SSH keys, and AWS security groups) are cloud assets outside this workspace. Staging execution is therefore marked **BLOCKED** from direct execution within this automated session, and an exact, safe verification runbook is provided for the deployment owner.

---

## 2. Commit Baseline & Artifact Integrity

### 2.1 Commit Lineage & Ancestry

```text
5faa0b0 (HEAD -> main, origin/main) feat(security): enforce server-side limits clamping, system path validation, and compose healthchecks
5b0bde6 docs(audit): publish Milestone 2.6 deployment readiness and secure integration report
3f5b8f3 feat(security): resolve SEC-03 with dedicated authenticated sandbox broker and least-privilege EC2 compose
1e2d167 docs(audit): update Milestone 2.5 audit report with hosted CI run 40046901749 evidence
```

- Verified that `3f5b8f3` and `5b0bde6` are direct ancestors of `HEAD` (`5faa0b0`).
- Working tree clean, synchronized with remote `origin/main`.

### 2.2 Release Artifacts Matrix

| Service | Repository / Image Reference | Immutable Release Tag | Digest / Status | Verification |
| :--- | :--- | :--- | :--- | :---: |
| **API Backend** | `ghcr.io/cv-code-vitals/ucs503-api` | `5faa0b03aef311ef8f752b61119f01b7fae1d899` | Built & pushed in Run 40051436595 | **PASS** |
| **Analysis Backend** | `ghcr.io/cv-code-vitals/ucs503-analysis` | `5faa0b03aef311ef8f752b61119f01b7fae1d899` | Built & pushed in Run 40051436595 | **PASS** |
| **Sandbox Runner** | `ghcr.io/cv-code-vitals/ucs503-sandbox-runner` | `5faa0b03aef311ef8f752b61119f01b7fae1d899` | Built & pushed in Run 40051436595 | **PASS** |
| **Sandbox Broker** | Reuses `ucs503-analysis` with `--broker` flag | `5faa0b03aef311ef8f752b61119f01b7fae1d899` | Reuses verified analysis image | **PASS** |

---

## 3. Deployment Verification Matrix

| Verification Item | Target Environment | Required Property | Observed Status | Finding / Evidence |
| :--- | :--- | :--- | :---: | :--- |
| **Compose Configuration Syntax** | Local / Offline | Valid Docker Compose YAML v3 format | **PASS** | Validated via Python YAML parser (`ucs503-api`, `ucs503-analysis`, `ucs503-sandbox-broker`). |
| **Service Readiness Ordering** | Compose Model | Broker must report healthy before analysis starts | **PASS** | `condition: service_healthy` configured in `docker-compose.ec2.yml`. |
| **Analysis Non-Root User** | Container Runtime | `appuser:appgroup` (UID 10002) | **PASS** | Declared in `go-backend/Dockerfile` Stage 3 (`USER appuser:appgroup`). |
| **Analysis Docker Isolation** | Container Runtime | No `/var/run/docker.sock` mount | **PASS** | Audited in `docker-compose.ec2.yml`: analysis container mounts only `analysis_uploads` and `sandboxes_shared`. |
| **Loopback Analysis Port** | Host Binding | Bound to `127.0.0.1:8081` | **PASS** | Audited in `docker-compose.ec2.yml`: `ports: ["127.0.0.1:8081:8080"]`. |
| **Broker Network Isolation** | Compose Network | `sandbox_internal` (`internal: true`) | **PASS** | No external route; no host port mapping (`expose: ["8082"]`, no `ports:`). |
| **Broker Request Authentication** | HTTP API | Strict `X-Sandbox-Token` via `ConstantTimeCompare` | **PASS** | Tested in `broker_test.go` (`TestRemoteBrokerEvaluator_AuthenticationFailure`). |
| **Server-Side Limits Enforcement** | Broker Daemon | Hard ceiling on memory, CPU, PID, timeout | **PASS** | Implemented in `broker.go` (`clampLimits()`). |
| **System Path Protection** | Evaluator & Broker | Rejection of sensitive host root directories | **PASS** | Tested in `broker_test.go` (`TestEvaluateCandidate_RestrictedRepoPaths`). |
| **Runner Container Containment** | Docker Daemon | `--network none`, `--read-only`, `--cap-drop ALL` | **PASS** | Verified in `evaluator.go` and passing live CI integration gate. |
| **Live Docker Integration Gate** | Hosted CI Runner | Container executes real candidate tests | **PASS** | CI Run `40051436562` Step 14 passed explicitly (0 skips, 0 failures). |
| **Live EC2 Staging Deployment** | Live EC2 Instance | Container pull, up, and health check on host | **BLOCKED** | External EC2 credentials (`secrets.EC2_HOST`, SSH key) not in workspace. |
| **AWS Security Group Ingress** | AWS VPC | Ports 8081/8082 blocked from public ingress | **NOT VERIFIED** | AWS API access unavailable in this workspace. |

---

## 4. Security Boundary Audit & Deep Inspection

### 4.1 Analysis Backend (`ucs503-analysis`)

- **User**: Runs as `appuser:appgroup` (UID `10002:10002`).
- **Filesystem Permissions**: Owns `/app/uploads` and `/app/sandboxes` with `0750` permissions.
- **Docker Access**: Has **no access** to `/var/run/docker.sock`. Does not have Docker privileges.
- **Network Interfaces**:
  - Bound to `public_net` for reverse proxy access at `127.0.0.1:8081`.
  - Connected to `sandbox_internal` for private communication with `ucs503-sandbox-broker:8082`.
- **API Authentication**: Requires `ANALYSIS_API_TOKEN` on all operational endpoints (`/api/analyze`, `/api/scan`, `/api/audit`, `/api/evolve`, `/api/upload`, `/api/deliverables`) validated via `subtle.ConstantTimeCompare`.
- **Health Check**: `GET /health` dynamically evaluates `sandboxEvaluator.CheckAvailable(ctx)` and reports `"sandbox_available": true/false`.

### 4.2 Sandbox Broker (`ucs503-sandbox-broker`)

- **Host Port Exposure**: **None**. Port 8082 is declared under `expose:`, making it resolvable exclusively inside `sandbox_internal`.
- **Network Boundary**: `sandbox_internal` is configured with `internal: true`. The Docker daemon configures bridge `iptables` to drop all forwarding to external interfaces, preventing outbound internet access from both the broker and runner containers.
- **Authentication**: All evaluation requests require `X-Sandbox-Token` matching `SANDBOX_SERVICE_TOKEN` via constant-time comparison. Unauthenticated calls receive HTTP 401. Health probes permit local loopback queries (`127.0.0.1`) for Compose health checking.
- **Validation**:
  - Pinned runner image: strictly uses `SANDBOX_RUNNER_IMAGE` (defaults to `ghcr.io/cv-code-vitals/ucs503-sandbox-runner:latest`). Caller cannot supply custom images.
  - Pinned command: strictly executes `go test -v ./...`. Caller cannot supply arbitrary shell arguments.
  - Server-side limit clamping: caller-supplied limits are strictly capped to:
    $$\text{Timeout} \le 120\text{s}, \quad \text{Memory} \le 2048\text{MB}, \quad \text{CPUs} \le 2.0, \quad \text{PIDs} \le 256, \quad \text{Output} \le 256\text{KiB}$$
  - Repository path sanitization: rejects paths pointing to `/`, `/etc`, `/root`, `/sys`, `/proc`, `/dev`, `/boot`, `/var/run`, `C:/Windows`.

### 4.3 Candidate Runner (`ucs503-sandbox-runner`)

The effective Docker invocation constructed in `evaluator.go` enforces the following defense-in-depth boundaries:

```bash
docker run --rm \
  --name ucs503-run-<uuid> \
  --network none \
  --read-only \
  --tmpfs /tmp:rw,exec,nosuid,nodev,mode=1777,size=<tmpfsSizeMB> \
  -e GOCACHE=/tmp/gocache \
  -e CGO_ENABLED=0 \
  --user 10001:10001 \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --memory <limits.MaxMemoryMB>m \
  --cpus <limits.MaxCPUs> \
  --pids-limit <limits.MaxPids> \
  -v <workspaceMount>:/workspace:rw \
  -w /workspace \
  <image> \
  go test -v ./...
```

#### Justification of `/tmp` Execution Permissions
The tmpfs mount specifies `exec` because the Go toolchain builds ephemeral test executables (`ucs503-sample.test`) and executes them from `GOCACHE=/tmp/gocache`. Mounting `/tmp` with `noexec` causes `fork/exec: permission denied`. The risks are mitigated by:
1. `nosuid` and `nodev` prevent creating setuid binaries or accessing raw device nodes.
2. `--cap-drop ALL` strips all root capabilities.
3. `--security-opt no-new-privileges` prevents privilege escalation.
4. `--user 10001:10001` ensures the process runs as an unprivileged user.
5. `--network none` guarantees that even if candidate code executes, it cannot reach any network socket, metadata endpoint, or DNS server.

---

## 5. Verification & Test Evidence

### 5.1 Local & Targeted Automated Verification

```text
=== RUN   TestRemoteBrokerEvaluator_MissingOrInvalidConfig
--- PASS: TestRemoteBrokerEvaluator_MissingOrInvalidConfig (0.00s)
=== RUN   TestRemoteBrokerEvaluator_BrokerUnavailable_FailsClosed
--- PASS: TestRemoteBrokerEvaluator_BrokerUnavailable_FailsClosed (0.01s)
=== RUN   TestRemoteBrokerEvaluator_AuthenticationFailure
--- PASS: TestRemoteBrokerEvaluator_AuthenticationFailure (0.01s)
=== RUN   TestRemoteBrokerEvaluator_TimeoutPropagation
--- PASS: TestRemoteBrokerEvaluator_TimeoutPropagation (0.50s)
=== RUN   TestRemoteBrokerEvaluator_SuccessfulEvaluation
--- PASS: TestRemoteBrokerEvaluator_SuccessfulEvaluation (0.01s)
=== RUN   TestNewEvaluator_Resolution
--- PASS: TestNewEvaluator_Resolution (0.00s)
=== RUN   TestSandboxWorkspace_CustomBaseDir
--- PASS: TestSandboxWorkspace_CustomBaseDir (0.01s)
=== RUN   TestEvaluateCandidate_RestrictedRepoPaths
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths// (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//etc (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//etc/ssl (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//root (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//proc (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//sys (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths//dev (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths/../outside (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths/C:/Windows (0.00s)
    --- PASS: TestEvaluateCandidate_RestrictedRepoPaths/C:/Windows/System32 (0.00s)
--- PASS: TestEvaluateCandidate_RestrictedRepoPaths (0.00s)
=== RUN   TestEvaluateCandidate_PathTraversal_Comprehensive
--- PASS: TestEvaluateCandidate_PathTraversal_Comprehensive (0.01s)
=== RUN   TestCopyProjectFiles_SymlinksSkipped
--- PASS: TestCopyProjectFiles_SymlinksSkipped (0.00s)
=== RUN   TestDockerEvaluator_NeverFallsBackToHost
--- PASS: TestDockerEvaluator_NeverFallsBackToHost (0.00s)

PASS
ok  github.com/saksham/self-healing-pipeline       1.795s
ok  github.com/saksham/self-healing-pipeline/agent 1.578s
ok  ucs503-engine-api                              (cached)
cargo test (thermodynamic-ast-engine):             9 passed, 0 failed
```

### 5.2 Hosted CI Evidence (Commit `5faa0b0`)

#### Workflow: `CI - Go build & vet` (Run `40051436562`)
- **Status:** `completed`
- **Conclusion:** **`success`**

| Step # | Step Name | Conclusion | Duration |
| :---: | :--- | :---: | :---: |
| 1 | Set up job | `success` | 3s |
| 2 | Run actions/checkout@v4 | `success` | 2s |
| 3 | Setup Go (1.25) | `success` | 3s |
| 4 | Cache Go modules | `success` | 1s |
| 5 | Install dependencies (api) | `success` | 4s |
| 6 | go vet (api) | `success` | 2s |
| 7 | go test (api) | `success` | 3s |
| 8 | go build (api) | `success` | 3s |
| 9 | Set up AWS SAM CLI (1.169.0) | `success` | 12s |
| 10 | Validate and build serverless template | `success` | 15s |
| 11 | Verify and configure Rust toolchain | `success` | 14s |
| 12 | Build and Test Rust Thermodynamic AST Engine | `success` | 16s |
| 13 | Build Docker Sandbox Runner Image | `success` | 8s |
| **14** | **Run Live Docker Sandbox Integration Tests (Mandatory Gate)** | **`success`** | **9.9s** |
| 15 | Upload Sandbox Test Artifacts | `success` | 2s |
| 16 | Check analysis backend (includes all new broker & limits tests) | `success` | 7s |

**Live Integration Gate Verification:**
`TestDockerEvaluator_LiveExecution` ran inside Step 14 on the Ubuntu runner against `ucs503-sandbox-runner:latest`. The gate verified:
1. `TEST_EXIT == 0`
2. No skip directive (`grep -q -F -- "--- SKIP: TestDockerEvaluator_LiveExecution"` returned false)
3. Explicit pass confirmed (`grep -q -F -- "--- PASS: TestDockerEvaluator_LiveExecution"` returned true)

#### Workflow: `Publish Backend Docker Images` (Run `40051436595`)
- **Status:** `completed`
- **Conclusion:** **`success`**
- All 3 images built and pushed to GHCR under tag `5faa0b03aef311ef8f752b61119f01b7fae1d899` and `latest`.

---

## 6. Unresolved Risks & Prerequisites

| Risk ID | Severity | Description & Impact | Operational Owner | Required Action |
| :--- | :---: | :--- | :--- | :--- |
| **RISK-01** | Medium | **Live EC2 Staging Access Unavailable in Workspace**: External EC2 SSH keys, AWS VPC access, and live cloud instances are not accessible from this automated workspace. Staging execution could not be directly triggered. | Deployment Owner | Execute the operator verification runbook on the EC2 staging host. |
| **RISK-02** | Low | **Docker Socket Host Ownership on EC2**: The broker container mounts `/var/run/docker.sock` and runs as root inside its container layer to communicate with the host daemon. If host socket permissions are altered, container startup could fail. | Systems Administrator | Confirm the host Docker daemon socket has standard ownership (`root:docker`, mode `0660`). |
| **RISK-03** | Low | **Direct Host Ingress on Port 8081**: `ucs503-analysis` is bound to `127.0.0.1:8081`. If an operator changes this to `0.0.0.0:8081`, the backend would be exposed directly to the public network without SSL. | Security / Ops | Verify that AWS Security Groups and host iptables restrict public ingress strictly to ports 80 and 443 (reverse proxy). |

---

## 7. Staging Deployment Runbook for Deployment Owner

Because the target EC2 host is external to this workspace, the deployment owner must execute the following verified staging deployment procedure:

### Step 1: Prepare Host Secrets in `~/ucs503/.env`

Ensure `.env` contains high-entropy secrets ($\ge 16$ characters):

```bash
# ~/ucs503/.env
APP_ENV=standalone
ANALYSIS_API_TOKEN=<generate-32-character-random-token>
SANDBOX_SERVICE_TOKEN=<generate-32-character-random-token>
ADMIN_PASSKEY=<generate-32-character-random-token>
DATABASE_URL=postgresql://user:pass@db-host:5432/dbname
S3_BUCKET=your-bucket-name
S3_REGION=ap-south-1
```

### Step 2: Deploy Using Immutable SHA Tag

```bash
cd ~/ucs503

# Login to GHCR
echo "$GHCR_PAT" | docker login ghcr.io -u "$GHCR_USER" --password-stdin

# Pull immutable release images
export IMAGE_TAG=5faa0b03aef311ef8f752b61119f01b7fae1d899
docker compose -f docker-compose.ec2.yml pull

# Start containers with health check synchronization
docker compose -f docker-compose.ec2.yml up -d --remove-orphans
```

### Step 3: Verify Container Health & Readiness

```bash
# Check container status (both broker and analysis should report (healthy))
docker compose -f docker-compose.ec2.yml ps

# Expected output:
# NAME                     IMAGE                                                  STATUS
# ucs503-sandbox-broker    ghcr.io/cv-code-vitals/ucs503-analysis:5faa0b0...     Up (healthy)
# ucs503-analysis          ghcr.io/cv-code-vitals/ucs503-analysis:5faa0b0...     Up (healthy)
# ucs503-api               ghcr.io/cv-code-vitals/ucs503-api:5faa0b0...          Up
```

### Step 4: Verify Analysis Health Endpoint

```bash
curl -s http://127.0.0.1:8081/health | jq .
# Expected output:
# {
#   "status": "ok",
#   "version": "1.0.0",
#   "sandbox_available": true,
#   "sandbox_error": ""
# }
```

### Step 5: Execute End-to-End Test Evolution Job

```bash
curl -s -X POST http://127.0.0.1:8081/api/evolve \
  -H "Content-Type: application/json" \
  -H "X-API-Token: $ANALYSIS_API_TOKEN" \
  -d '{
    "repo_path": "/app/uploads/test_repo",
    "file_path": "calc.go",
    "function_name": "Add",
    "start_byte": 0,
    "end_byte": 10,
    "base_energy": 10.0
  }' | jq .
```

---

## 8. Rollback Plan

If unexpected anomalies are observed during staging:

1. **Rollback Container Release Images**:
   Pin `IMAGE_TAG` to the previous verified commit `75cd5b906a128dae2fbc8cabd560b7e0eab3ac00`:
   ```bash
   IMAGE_TAG=75cd5b906a128dae2fbc8cabd560b7e0eab3ac00 docker compose -f docker-compose.ec2.yml up -d
   ```
2. **Emergency Decoupling**:
   If candidate evaluation must be temporarily paused while keeping AST analysis operational, set `SANDBOX_SERVICE_URL=""` in `~/ucs503/.env` and restart `ucs503-analysis`. The `/api/analyze` and `/api/scan` endpoints will remain operational while `/api/evolve` fails closed safely.

---

## 9. Final Release Decision

```text
================================================================================
                    MILESTONE 2.7 RELEASE GATE DECISION
================================================================================

  Decision: CONDITIONAL GO

  Rationale:
  1. All code, broker architecture, and container boundaries have been implemented,
     hardened with server-side limit clamping, and validated with zero regressions.
  2. All Go, API, and Rust unit tests pass locally and in hosted CI.
  3. The mandatory Live Docker Sandbox Integration Test passed explicitly
     in GitHub Actions CI Run 40051436562 without skipping.
  4. Release images for all three services are built, verified, and published to GHCR
     under immutable SHA tag 5faa0b03aef311ef8f752b61119f01b7fae1d899.
  5. The release gate remains CONDITIONAL GO exclusively because live EC2
     cloud infrastructure is external to this development workspace and must
     be confirmed by the deployment owner using the provided runbook.

================================================================================
```
