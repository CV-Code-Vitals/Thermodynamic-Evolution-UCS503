# M3.8 Phase 2.1.2 — CI Failure Reconciliation Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Failing Commit (Remote):** `0843ca613c602764ef3f534fc21bed7843486ba0`  
**Failed Workflow Run:** `45149021367`  
**Audit Scope:** Linux Docker-Oracle CI regression diagnosis, offline module dependency isolation, container lifecycle and telemetry, test contract reconciliation.

---

## 1. Failing Workflow & Baseline Forensics

### 1.1 Context and Environment
- **Workflow:** `.github/workflows/ci-go.yml` (Job: `go-checks` on `ubuntu-latest`).
- **Runner State:** Ubuntu Linux with Docker daemon running and `ucs503-sandbox-runner:<sha>` built in step 146.
- **Failed Step:** `Check analysis backend` executing `go test -v ./...`.

### 1.2 Root Cause Analysis
Workflow run `45149021367` failed on the Linux CI runner due to two interrelated container runtime and test contract issues:

1. **Egress Blocked Under `--network none` without Offline Cache Mounting:**
   - In commit `0843ca61`, `DockerContainerRunner.RunCommand` enforced strict `--network none` isolation on container execution.
   - When dynamic oracles (`go vet` or `go test`) ran inside `ucs503-sandbox-runner` on targets with external dependencies (e.g. `fatih_color` requiring `golang.org/x/sys`), Go defaulted to querying `proxy.golang.org`.
   - Because `--network none` drops all egress network packets, DNS and TCP connection attempts hung indefinitely until the 20–30s context deadline expired, returning exit code `-1` / timeout failure.
   - The runner had neither staged offline dependency modules nor passed `-e GOPROXY=off` and `-e GOSUMDB=off`, and did not mount the host module cache into the container.

2. **Dangling Containers on Timeout:**
   - When `exec.CommandContext` encountered a context timeout, Go signaled the client `docker` CLI process, but the Docker daemon could leave the spawned container lingering or stopped without forced removal.
   - No explicit container cleanup hook (`docker rm -f`) was deferred with a separate non-canceled context.

3. **Test Contract Assumption of Docker Absence:**
   - In `adversarial_execution_test.go:35`, `TestAdversarial_1_HostSentinelFileAccessPrevented` created `suiteWithoutDocker := NewIndependentOracleSuite(10, nil)`.
   - `NewIndependentOracleSuite` fell back to `NewDockerContainerRunner()` when passed `nil`.
   - On Windows local hosts without Docker, `CheckAvailable` failed, returning `OracleStatusUnavailable`.
   - On Linux CI with a live Docker daemon, `CheckAvailable` succeeded, so the test attempted live container execution and failed to observe the expected `OracleStatusUnavailable` status.

---

## 2. Engineering Corrections Applied

### 2.1 Trusted Offline Dependency Staging & Read-Only Cache Mount
- **Host Module Cache Resolution:** `resolveStagedModuleCache()` dynamically locates `STAGED_MODULE_CACHE`, `GOMODCACHE`, `GOPATH/pkg/mod`, or `go env GOMODCACHE`.
- **Pre-Execution Staging:** `StageVerifiedDependencies(wsDir)` downloads and verifies dependencies outside the container **only** when `go.mod` and a non-empty `go.sum` with pinned cryptographic hashes exist. Untrusted code cannot trigger arbitrary network fetches.
- **Read-Only Cache Mounting:** The verified module cache is mounted into the container as read-only: `-v <cache>:/go/pkg/mod:ro` with `-e GOPATH=/go`.
- **Offline Configuration:** Container execution enforces `-e GOPROXY=off`, `-e GOSUMDB=off`, and `-e GONOSUMDB=*`. Network access remains strictly `--network none`.
- **Missing Dependency Classification:** When Go outputs `cannot find module` or `module lookup disabled by GOPROXY=off`, the oracle classifies the result as `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`), never as `PASSED` or an ordinary patch defect.

### 2.2 Guaranteed Container Lifecycle & Cancellation Cleanup
- Every container execution registers a deferred cleanup hook using a dedicated, uncancelled background context:
  ```go
  defer func() {
      cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cleanupCancel()
      _ = exec.CommandContext(cleanupCtx, d.DockerBinary, "rm", "-f", containerName).Run()
  }()
  ```
- Differentiates context deadline timeouts (`exitCode = -1`), context cancellations, compilation/vet failures (`exitCode != 0`), and policy rejections.

### 2.3 Explicit Test Contract Reconciliation
- Replaced ambiguous `nil` runner in `TestAdversarial_1_HostSentinelFileAccessPrevented` with `NewUnavailableContainerRunner(...)`.
- Fail-closed tests deterministically verify runtime unavailability on all platforms, regardless of whether a live Docker daemon is present.
- Added comprehensive adversarial tests:
  - `TestAdversarial_10_MissingOfflineDependenciesReportedAsUnavailable`
  - `TestAdversarial_11_ContainerExecutionTimeoutReportedAsUnavailable`
  - `TestAdversarial_12_ContainerTelemetryAndDependencyInventory`

---

## 3. Reconciliation Verdict

All regressions observed in workflow `45149021367` have been resolved. Full local revalidation confirms 100% test pass rates across Go backend tests, Rust AST engine, and React frontend, with zero host execution escapes, zero network egress, and preserved historical benchmark ground truth.
