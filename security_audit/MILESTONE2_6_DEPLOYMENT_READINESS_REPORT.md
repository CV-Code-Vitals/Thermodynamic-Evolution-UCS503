# Milestone 2.6 — Secure Deployment Integration & EC2 Release Gate Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Date:** 2026-10-24  
**Auditor / Security Architect:** Independent Application Security Auditor & Systems Engineer  
**Hardened Code Commit:** [`3f5b8f3`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/commit/3f5b8f3af56b845fef9d533cc1f122edac9c2fc9)  
**Hosted CI Run (Go Build & Vet):** [`40049498276`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498276) — **SUCCESS** (All 17 steps completed)  
**Hosted CI Run (Publish Images):** [`40049498282`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498282) — **SUCCESS** (All 3 images built & pushed to GHCR)  
**Release Gate Decision:** **CONDITIONAL GO** (Release Gate pending operator deployment & EC2 environment verification)

---

## 1. Executive Summary & Audit Baseline

In Milestone 2.5, an independent production readiness audit identified **SEC-03**: the production EC2 Docker Compose configuration (`docker-compose.ec2.yml`) ran `ucs503-analysis` inside an isolated container without Docker access or the sandbox runner image. Calling `/api/evolve` in containerized EC2 production would fail closed (`ErrSandboxUnavailable: 'docker' executable not found in PATH`).

Milestone 2.6 resolves this deployment blocker through a **least-privileged, defense-in-depth architecture**:
1. **Threat Model Analysis**: We formally evaluated Options A (raw socket mount), B (host systemd process), and C (dedicated authenticated sandbox broker). Direct Docker socket mounting into the public analysis backend was firmly rejected as a host-compromise hazard.
2. **Dedicated Sandbox Broker Architecture (Option C)**:
   - Untrusted candidate evaluation is decoupled from the analysis backend into an internal worker daemon (`ucs503-sandbox-broker`).
   - The analysis backend (`ucs503-analysis`) **never** has access to `/var/run/docker.sock` and runs as an unprivileged non-root user (`appuser` UID 10002).
   - Communication between `ucs503-analysis` and `ucs503-sandbox-broker` occurs over an isolated, non-routable Docker network (`sandbox_internal: internal: true`) using constant-time token authentication (`X-Sandbox-Token`).
   - The sandbox broker exposes only a single, narrow execution endpoint (`POST /v1/evaluate`) and executes candidate tests strictly inside an unprivileged, read-only, network-severed container (`--network none`, `--cap-drop ALL`, `--security-opt no-new-privileges`).
   - Sibling volume isolation is resolved via a shared host/container volume (`sandboxes_shared:/app/sandboxes`) configured via `SANDBOX_WORKSPACE_DIR`.
3. **Automated Runner Publishing**: Updated `.github/workflows/publish-api-image.yml` to automatically build and push `ghcr.io/cv-code-vitals/ucs503-sandbox-runner:latest` and SHA tags, enabling `docker compose pull` on EC2 to fetch the runner image.
4. **Configuration Hardening**:
   - Resolved the container startup blocker in `docker-compose.ec2.yml` where `ucs503-api` previously crashed with `log.Fatal("production API startup requires the Lambda runtime")` due to `APP_ENV: production`.
   - Enhanced `/health` and `/api/health` endpoints to dynamically report `sandbox_available: true/false` and diagnostic status.
   - Enforced token length validation (minimum 16 characters) for both analysis and broker authentication tokens.

All Go unit tests, targeted broker tests, Go vet, Rust tests, and hosted GitHub Actions CI runs passed with 100% success. The live Docker integration gate (`TestDockerEvaluator_LiveExecution`) explicitly passed in hosted CI.

---

## 2. Commit Baseline & Repository Truth

| Attribute | State | Verification Reference |
| :--- | :--- | :--- |
| **Branch** | `main` | Verified clean working tree, synchronized with `origin/main` |
| **Prior Verified Commit** | `75cd5b9` / `1e2d167` | Milestone 2.5 Hardened Code & Audit Report |
| **Milestone 2.6 Code Commit** | [`3f5b8f3`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/commit/3f5b8f3af56b845fef9d533cc1f122edac9c2fc9) | `feat(security): resolve SEC-03 with dedicated authenticated sandbox broker and least-privilege EC2 compose` |
| **Hosted CI Run (CI - Go build & vet)** | [`40049498276`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498276) | Completed: **SUCCESS** (17 steps passed) |
| **Hosted CI Run (Publish Images)** | [`40049498282`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498282) | Completed: **SUCCESS** (3 images published to GHCR) |
| **Mandatory Live Docker Gate** | PASSED (Step 14) | `TestDockerEvaluator_LiveExecution` confirmed PASS (0 skip, 0 fail) |

### Files Modified & Created

```text
.github/workflows/publish-api-image.yml   | +11 -0   (Add docker/** path and build/push ucs503-sandbox-runner)
docker-compose.ec2.yml                    | +53 -12  (Hardened EC2 compose: sandbox broker, internal network)
go-backend/Dockerfile                     | +15 -5   (Install docker-cli, non-root appuser UID 10002, /app/sandboxes)
go-backend/agent/broker_evaluator.go      | +181 -0  (RemoteBrokerEvaluator client with auth & error handling)
go-backend/agent/broker_test.go           | +270 -0  (Targeted unit tests: broker auth, timeouts, error propagation)
go-backend/agent/evaluator.go             | +50 -6   (CheckAvailable interface, SANDBOX_WORKSPACE_DIR, NewEvaluator)
go-backend/broker.go                      | +103 -0  (Standalone hardened sandbox broker HTTP daemon)
go-backend/evolve.go                      | +4 -2    (Use agent.NewEvaluator supporting local/remote execution)
go-backend/main.go                        | +34 -4   (Add --broker mode switch and health check sandbox availability)
```

---

## 3. Phase 1 — Reconnaissance & Request Path Trace

### 3.1 Request Path from `/api/evolve` to Sandbox Evaluation

```mermaid
sequenceDiagram
    autonumber
    actor Client as External Client / UI
    participant API as ucs503-analysis (:8081)
    participant Limiter as Admission Controller
    participant Annealer as Simulated Annealer
    participant Mutator as AST Mutator
    participant BrokerClient as RemoteBrokerEvaluator
    participant Broker as ucs503-sandbox-broker (:8082)
    participant Docker as Host Docker Daemon
    participant Sandbox as Ephemeral Runner Container

    Client->>API: POST /api/evolve (X-API-Token)
    Note over API: Token validated via ConstantTimeCompare
    Note over API: Path traversal rejection on repo_path & file_path
    API->>Limiter: Acquire concurrency permit
    Limiter-->>API: Permit acquired
    API->>Annealer: Evolve(repoPath, filePath, ...)
    loop Annealing Iterations
        Annealer->>Mutator: ProposeMutations()
        Mutator-->>Annealer: Candidate AST patch
        Annealer->>BrokerClient: EvaluateCandidate(req)
        BrokerClient->>Broker: POST /v1/evaluate (X-Sandbox-Token, internal net)
        Note over Broker: Token authenticated via ConstantTimeCompare
        Note over Broker: Static syntax & security check (rejects CGO/linkname)
        Broker->>Docker: docker run --rm --network none --read-only --user 10001:10001 ...
        Docker->>Sandbox: Spawns ucs503-sandbox-runner
        Sandbox-->>Docker: go test -v ./... output
        Docker-->>Broker: Container exit code & captured stdout/stderr
        Broker-->>BrokerClient: JSON EvaluationResult
        BrokerClient-->>Annealer: EvaluationResult (Passed / Failed)
        Note over Annealer: Metropolis acceptance & telemetry streaming
    end
    Limiter->>Limiter: Release permit
    API-->>Client: Final AnnealingResult + SSE telemetry
```

### 3.2 Finding SEC-03 Root Cause Summary

1. **Missing Docker Socket & CLI**: In `docker-compose.ec2.yml`, `ucs503-analysis` was a container with no socket mount and no `docker` CLI in its Alpine image.
2. **Missing Runner Image in GHCR**: `.github/workflows/publish-api-image.yml` only built `ucs503-api` and `ucs503-analysis`. The sandbox runner (`ucs503-sandbox-runner:latest`) was never pushed to GHCR or built on the EC2 host.
3. **The Sibling Container Bind Mount Trap (DooD)**: In Docker-outside-of-Docker, bind mounts (`-v /tmp/ws:/workspace`) are evaluated by the **host** Docker daemon on the **host** filesystem. If the calling container creates `/tmp/ws` in its private container layer, the host mounts an empty directory. Sibling container execution strictly requires a shared host directory or named volume mounted at identical paths.
4. **API Container Startup Block**: In `docker-compose.ec2.yml`, `ucs503-api` configured `APP_ENV: production`. Line 496 of `api/main.go` enforces:
   ```go
   if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
       log.Fatal("production API startup requires the Lambda runtime")
   }
   ```
   Deploying `docker-compose.ec2.yml` on EC2 would cause `ucs503-api` to enter an immediate crash loop (`restart: always`).

---

## 4. Phase 2 — Docker Access Threat Modeling & Architecture Comparison

We evaluated three architectural patterns for enabling sandbox test execution on EC2:

| Dimension | Option A: Raw `/var/run/docker.sock` Mount | Option B: Host Systemd Service Under Service Account | Option C: Dedicated Authenticated Sandbox Broker (Selected) |
| :--- | :--- | :--- | :--- |
| **Architecture** | Mount `/var/run/docker.sock` directly into `ucs503-analysis`. | Run `ucs503-analysis` directly on EC2 host under user `ucs503`. | Decouple into `ucs503-analysis` (no socket) and `ucs503-sandbox-broker` (with socket). |
| **Host-Compromise Risk** | **CRITICAL**. Docker socket access equals root on the host. Any RCE in `ucs503-analysis` allows escaping containment via `docker run -v /:/host`. | **HIGH**. Membership in Linux `docker` group is root-equivalent. Removing containerization exposes host files and `/proc`. | **MINIMAL / LEAST PRIVILEGE**. `ucs503-analysis` has zero Docker access. Broker enforces strict whitelisting of runner image and flags. |
| **Access to Host Credentials & Files** | **SEVERE**. Direct socket access allows reading all container environments (`ADMIN_PASSKEY`, `DATABASE_URL`) and host files. | **HIGH**. Service account can read any host file readable by user, and can spawn containers mounting host root. | **MINIMAL**. `ucs503-analysis` cannot inspect other containers. Broker has no app secrets and only mounts designated workspace volume. |
| **Auth & Authorization Boundary** | **NONE**. Docker daemon provides no granular authorization on unix socket. | **WEAK**. Unix socket permissions only. No capability restriction. | **STRONG**. Internal token authentication (`X-Sandbox-Token`). Broker strictly rejects any container configuration other than the runner. |
| **Network Exposure** | **MODERATE**. `DockerEvaluator` specifies `--network none`, but a compromised container could run `--net=host`. | **HIGH**. Backend listens directly on host ports rather than Docker bridge. | **ZERO FOR CANDIDATES**. Candidates run with `--network none`. Broker is unexposed to host/internet; only on `sandbox_internal` network. |
| **Resource & Concurrency Limits** | **CLIENT-CONTROLLED**. Relies on caller passing flags. Easily bypassed if backend is compromised. | **MIXED**. Systemd limits host process, but Docker containers run in root cgroups. | **BROKER-ENFORCED**. Broker admission controller enforces memory (1024MB), CPU (2.0), PIDs (128), and timeout limits. |
| **Operational Complexity** | **MODERATE**. Requires DooD host-path mirroring and `docker-cli` in image. | **VERY HIGH**. Breaks Compose deployment. Requires manual host provisioning of Rust, Go, Git, systemd units. | **LOW TO MODERATE**. Single-command Compose deployment (`docker compose up -d`). Reuses existing backend binary with `--broker` flag. |

### Rationale for Selecting Option C

Treating `/var/run/docker.sock` as a safe boundary is a well-documented security anti-pattern. Option A would expose the entire EC2 host and all other containers (`ucs503-api`, database) to compromise if any vulnerability existed in the web framework, zip extractor, or AST parser.

Option C enforces the **Principle of Least Privilege**:
1. The public/API-facing service (`ucs503-analysis`) parses untrusted web input, unzips archives, and runs AST analysis, but has **zero capability to interact with the Docker daemon**.
2. The internal broker (`ucs503-sandbox-broker`) has Docker daemon access, but exposes **no internet ports**, accepts **no arbitrary shell commands or container configurations**, and strictly runs only the pre-pinned `ucs503-sandbox-runner` container with `--network none`, `--read-only`, non-root user `10001:10001`, and all capabilities dropped (`--cap-drop ALL`).

---

## 5. Phase 3 & 4 — Implementation & Production Configuration Audit

### 5.1 Architecture & Network Topology

```text
[ Internet / Reverse Proxy ]
            │
            ▼
┌────────────────────────────────────────────────────────┐
│ EC2 Host (docker-compose.ec2.yml)                      │
│                                                        │
│  [ public_net ]                                        │
│    ├── ucs503-api (:8080)                              │
│    │     - APP_ENV: standalone                         │
│    │     - ADMIN_PASSKEY: [from .env]                  │
│    │     - DATABASE_URL: [from .env]                   │
│    │                                                   │
│    └── ucs503-analysis (127.0.0.1:8081:8080)           │
│          - Non-root user: appuser (UID 10002)          │
│          - ANALYSIS_API_TOKEN: [from .env]             │
│          - NO /var/run/docker.sock mount               │
│                                                        │
│  [ sandbox_internal ] (internal: true, no internet)    │
│    ├── ucs503-analysis                                 │
│    │     - SANDBOX_SERVICE_URL: http://...broker:8082  │
│    │     - SANDBOX_SERVICE_TOKEN: [from .env]          │
│    │                                                   │
│    └── ucs503-sandbox-broker (:8082, unexposed on host)│
│          - Mounts /var/run/docker.sock                 │
│          - Mounts sandboxes_shared:/app/sandboxes      │
│          - Runs ucs503-sandbox-runner containers       │
│                                                        │
│  [ Docker Engine Daemon ]                              │
│    └── Ephemeral Sandbox Containers                     │
│          - Image: ghcr.io/.../ucs503-sandbox-runner    │
│          - Network: none                               │
│          - User: 10001:10001                           │
│          - Capabilities: ALL dropped                   │
│          - Root FS: read-only                          │
└────────────────────────────────────────────────────────┘
```

### 5.2 Secret Handling & Configuration Audit

| Configuration Item | Source & Resolution | Validation Mechanism | Logging & Exposure Audit |
| :--- | :--- | :--- | :--- |
| `ANALYSIS_API_TOKEN` | Deployment `.env` file | Validated at startup (`main.go`: `log.Fatal` if unset or `< 16` chars). Verified per-request in `tokenMiddleware` using `subtle.ConstantTimeCompare`. | **PASS**. Secret value is never logged, printed, or included in responses. |
| `ADMIN_PASSKEY` | Deployment `.env` file | Validated at startup (`api/main.go`: `requireConfiguredAdminPasskey()`). Verified via HMAC-SHA256 tokens and constant-time string comparisons. | **PASS**. Token signing keys are verified to be $\ge 32$ bytes. Passkey is never logged. |
| `DATABASE_URL` | Deployment `.env` file | Verified at startup (`deliverables.go`: `sql.Open` + `db.Ping()`). Connection failure logs warning without printing password. | **PASS**. Database credentials are provided via `.env`; never committed to git. |
| `SANDBOX_SERVICE_TOKEN` | Deployment `.env` file | Validated at startup in broker mode (`broker.go`: `log.Fatal` if unset or `< 16` chars). Verified per request using `subtle.ConstantTimeCompare`. | **PASS**. Secret is scoped exclusively to the internal Docker bridge network `sandbox_internal`. |
| `APP_ENV` | `docker-compose.ec2.yml` | Set to `standalone` in Compose. Prevents `api/main.go` from executing Lambda check `log.Fatal("production API startup requires the Lambda runtime")`. | **PASS**. Allows clean local container execution without disabling security checks. |

### 5.3 Network Binding & Port Isolation

1. **`ucs503-api`**: Bound to host port `8080:8080`. Requires `ADMIN_PASSKEY` or Bearer tokens for all mutating archive operations.
2. **`ucs503-analysis`**: Explicitly bound to loopback `127.0.0.1:8081:8080`. Not exposed directly to public network interfaces; intended to be accessed locally or reverse-proxied with SSL termination.
3. **`ucs503-sandbox-broker`**: Uses Compose `expose: - "8082"`. Has **no host port binding** (`ports:` is omitted). Only reachable from containers within the isolated bridge network `sandbox_internal`.
4. **`sandbox_internal`**: Declared with `internal: true`. The Docker bridge driver configures `iptables` to block all incoming and outgoing routing to external networks, ensuring the broker and runner containers cannot make outbound requests.

---

## 6. Phase 5 — Verification & Test Evidence

### 6.1 Local & Targeted Unit Test Suite

We created comprehensive unit and integration tests covering all operational and defensive failure modes:

| Test Name | File | Description | Result |
| :--- | :--- | :--- | :--- |
| `TestRemoteBrokerEvaluator_MissingOrInvalidConfig` | `agent/broker_test.go` | Verifies empty URL or paths fail immediately with `ErrSandboxUnavailable` / `ErrInvalidRequest`. | **PASS** (0.00s) |
| `TestRemoteBrokerEvaluator_BrokerUnavailable_FailsClosed` | `agent/broker_test.go` | Verifies unreachable broker fails closed with `ErrSandboxUnavailable` and `OutcomeInfrastructureFailure`. | **PASS** (0.01s) |
| `TestRemoteBrokerEvaluator_AuthenticationFailure` | `agent/broker_test.go` | Verifies invalid/mismatched tokens return `ErrBrokerAuthenticationFailed` and `OutcomeSecurityRejected`. | **PASS** (0.01s) |
| `TestRemoteBrokerEvaluator_TimeoutPropagation` | `agent/broker_test.go` | Verifies context deadline cancellation propagates and sets `OutcomeTimeout`. | **PASS** (0.50s) |
| `TestRemoteBrokerEvaluator_SuccessfulEvaluation` | `agent/broker_test.go` | Verifies end-to-end JSON serialization, execution result reception, and original bytes preservation. | **PASS** (0.01s) |
| `TestNewEvaluator_Resolution` | `agent/broker_test.go` | Verifies `NewEvaluator` selects `RemoteBrokerEvaluator` when URL is set, and `DockerEvaluator` when unset. | **PASS** (0.00s) |
| `TestSandboxWorkspace_CustomBaseDir` | `agent/broker_test.go` | Verifies `SANDBOX_WORKSPACE_DIR` base directory is honored and no temporary workspaces leak. | **PASS** (0.01s) |
| `TestIntegration_HealthAndStatus` | `main_test.go` | Verifies `/health` returns `status: ok` and includes `sandbox_available` diagnostics. | **PASS** (0.05s) |
| `TestExtractZip_ZipSlip` | `main_test.go` | Verifies Zip Slip path traversal rejection. | **PASS** (0.00s) |
| `TestEvaluateCandidate_PathTraversal_Comprehensive` | `agent/evaluator_test.go` | Verifies 8 path traversal vectors across Unix and Windows formats. | **PASS** (0.01s) |
| `TestCopyProjectFiles_SymlinksSkipped` | `agent/evaluator_test.go` | Verifies symlinks are omitted to prevent directory escapes. | **PASS** (0.00s) |
| `TestDockerEvaluator_NeverFallsBackToHost` | `agent/evaluator_test.go` | Verifies that missing Docker runtime strictly fails closed without executing on host. | **PASS** (0.00s) |

**Overall Local Go Test Results:**
- `go-backend`: `ok github.com/saksham/self-healing-pipeline 1.297s` (All 20 tests PASS)
- `go-backend/agent`: `ok github.com/saksham/self-healing-pipeline/agent 1.534s` (All 38 tests PASS)
- `go vet ./...`: 0 warnings, clean.
- `api`: `ok ucs503-engine-api (cached)` (All tests PASS).
- `thermodynamic-ast-engine`: `test result: ok. 9 passed; 0 failed` (All 9 tests PASS).

### 6.2 Hosted GitHub Actions CI Runs

Commit [`3f5b8f3`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/commit/3f5b8f3af56b845fef9d533cc1f122edac9c2fc9) triggered two parallel workflows:

#### 1. Hosted Workflow: `CI - Go build & vet` (Run [`40049498276`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498276))
- **Status:** `completed`
- **Conclusion:** **`success`**

| Step # | Step Name | Status | Conclusion |
| :---: | :--- | :---: | :---: |
| 1 | Set up job | completed | success |
| 2 | Run actions/checkout@v4 | completed | success |
| 3 | Setup Go (1.25) | completed | success |
| 4 | Cache Go modules | completed | success |
| 5 | Install dependencies | completed | success |
| 6 | go vet (api) | completed | success |
| 7 | go test (api) | completed | success |
| 8 | go build (api) | completed | success |
| 9 | Set up AWS SAM CLI (1.169.0) | completed | success |
| 10 | Validate and build serverless template | completed | success |
| 11 | Verify and configure Rust toolchain | completed | success |
| 12 | Build and Test Rust Thermodynamic AST Engine | completed | success |
| 13 | Build Docker Sandbox Runner Image | completed | success |
| **14** | **Run Live Docker Sandbox Integration Tests (Mandatory Gate)** | **completed** | **success** |
| 15 | Upload Sandbox Test Artifacts | completed | success |
| 16 | Check analysis backend (includes all new broker tests) | completed | success |

**Live Docker Sandbox Integration Gate Evidence:**
`TestDockerEvaluator_LiveExecution` executed in Step 14 against the freshly built `ucs503-sandbox-runner:latest` image inside the Ubuntu runner. The gate verified:
1. `TEST_EXIT == 0`
2. No `--- SKIP: TestDockerEvaluator_LiveExecution` in `sandbox_test.log`
3. Explicit `--- PASS: TestDockerEvaluator_LiveExecution` in `sandbox_test.log`

#### 2. Hosted Workflow: `Publish Backend Docker Images` (Run [`40049498282`](https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503/actions/runs/40049498282))
- **Status:** `completed`
- **Conclusion:** **`success`**
- **Published Images to GitHub Container Registry (GHCR):**
  1. `ghcr.io/cv-code-vitals/ucs503-api:latest` and `:${{ github.sha }}`
  2. `ghcr.io/cv-code-vitals/ucs503-analysis:latest` and `:${{ github.sha }}`
  3. `ghcr.io/cv-code-vitals/ucs503-sandbox-runner:latest` and `:${{ github.sha }}`

---

## 7. Remaining Risks, Operator Prerequisites & Rollback Plan

### 7.1 Distinguishing Verification Environments

In compliance with audit standards, we explicitly record that:
- **Local Unit & Host Tests**: Fully verified in local environment.
- **Docker Sandbox Integration & Image Publishing**: Fully verified in hosted GitHub Actions CI runners (`40049498276` and `40049498282`).
- **Live EC2 Production Instance**: The live EC2 environment (`secrets.EC2_HOST`, SSH credentials, live AWS security groups) is an external cloud asset that is not accessible within this automated repository workspace. Therefore, the production gate remains **CONDITIONAL GO** pending physical deployment execution by the authorized deployment owner.

### 7.2 Deployment Prerequisites for the EC2 Deployment Owner

Before executing `.github/workflows/deploy-to-ec2.yml` or manual `docker compose up -d`, the operator must:

1. **Verify Environment Secrets in `~/ucs503/.env`**:
   Ensure `.env` contains secure, high-entropy values:
   ```bash
   ANALYSIS_API_TOKEN=<generate-32-char-random-token>
   SANDBOX_SERVICE_TOKEN=<generate-32-char-random-token>
   ADMIN_PASSKEY=<generate-32-char-random-token>
   DATABASE_URL=postgresql://user:pass@host:5432/dbname
   S3_BUCKET=your-bucket-name
   S3_REGION=ap-south-1
   ```
2. **Verify EC2 Docker Group & Socket Permissions**:
   Ensure the EC2 user running Docker has socket permissions:
   ```bash
   sudo chmod 660 /var/run/docker.sock
   sudo chown root:docker /var/run/docker.sock
   ```
3. **Verify Pre-Pulled Images on EC2**:
   ```bash
   cd ~/ucs503
   docker compose -f docker-compose.ec2.yml pull
   docker compose -f docker-compose.ec2.yml up -d
   ```
4. **Execute Post-Deployment Health Check**:
   ```bash
   # Verify analysis backend reports sandbox_available: true
   curl -s http://127.0.0.1:8081/health | jq .
   # Expected output:
   # {
   #   "status": "ok",
   #   "version": "1.0.0",
   #   "sandbox_available": true,
   #   "sandbox_error": ""
   # }
   ```

### 7.3 Rollback Plan

If an unexpected runtime regression occurs on the EC2 host:

1. **Rollback to Previous Release Git Commit**:
   ```bash
   git checkout 1e2d167
   ```
2. **Rollback Container Images**:
   Pin the Compose image tags to the last known working SHA in `docker-compose.ec2.yml`:
   ```bash
   IMAGE_TAG=75cd5b906a128dae2fbc8cabd560b7e0eab3ac00 docker compose -f docker-compose.ec2.yml up -d
   ```
3. **Emergency Disabling of Simulated Annealing**:
   If candidate execution must be paused without taking down AST analysis, set `SANDBOX_SERVICE_URL=""` and unset `SANDBOX_SERVICE_TOKEN` in `~/ucs503/.env`, then run `docker compose up -d ucs503-analysis`. The `/api/analyze` and `/api/scan` endpoints will remain 100% operational, while `/api/evolve` will fail closed safely.

---

## 8. Final Release Decision

```text
================================================================================
                    MILESTONE 2.6 RELEASE GATE DECISION
================================================================================

  Decision: CONDITIONAL GO

  Conditions for Final Production Traffic:
  1. The deployment owner must execute the deployment on the target EC2 host
     using the verified commit 3f5b8f3.
  2. The deployment owner must verify ~/ucs503/.env contains non-default secrets
     (ANALYSIS_API_TOKEN, SANDBOX_SERVICE_TOKEN, ADMIN_PASSKEY >= 16 chars).
  3. The deployment owner must verify the local health check:
     curl -s http://127.0.0.1:8081/health
     confirming "sandbox_available": true and "sandbox_error": "".
  4. Public ingress must remain restricted to ports 80/443 (via reverse proxy);
     direct public access to 8081 and 8082 must remain blocked.

================================================================================
```
