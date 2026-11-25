# M3.8 Phase 2.1.2 — CI Failure Reconciliation Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Review Baseline Commit:** `a0928fba56d81f90241460788322e89cdedc7334`  
**Prior Observed Failing Commit:** `0843ca613c602764ef3f534fc21bed7843486ba0` (Failed workflow `45149021367`)  
**Audit Scope:** Trust-boundary hardening, repository-directed dependency download prohibition, minimal isolated dependency staging, and pre-evidence container cleanup lifecycle.

---

## 1. Baseline & Vulnerability Diagnosis

### 1.1 Section A — Untrusted Repository-Directed Host Dependency Downloads
In previous iterations, `StageVerifiedDependencies()` executed `go mod download` directly on the host using the scanned repository's `go.mod` (`cmd.Dir = wsDir`). Furthermore, errors were discarded (`_ = cmd.Run()`), and `resolveStagedModuleCache()` located the host developer's global module cache (`~/go/pkg/mod` or `go env GOMODCACHE`) and mounted it entirely into the container (`-v <modCache>:/go/pkg/mod:ro`).

**Vulnerabilities Identified:**
1. **Host Toolchain Compromise Risk:** Running `go mod download` on the host from untrusted repository metadata exposes the host to Go toolchain vulnerabilities, arbitrary network egress, and poisoned global module caches.
2. **Global Module Cache Exposure:** Exposing the entire developer global module cache into an untrusted container grants read access to private, proprietary dependencies, cached VCS tokens, and internal modules unrelated to the target repository.
3. **Dynamic Toolchain Network Fetch:** Without `-e GOTOOLCHAIN=local`, Go 1.21+ toolchains attempt to fetch newer Go toolchain releases over the network if declared in `go.mod`.
4. **Silent Error Suppression:** Discarding staging errors prevented oracles from failing closed when dependencies were missing or invalid.

### 1.2 Section B — Premature Container Cleanup Telemetry
In `DockerContainerRunner.RunCommand()`, cleanup was previously scheduled in a `defer` block that executed *after* `evidence` was constructed and returned. Consequently:
- `evidence.CleanupResult` was recorded as `"NOT_NEEDED"` prior to actual removal.
- Cleanup failure, timeout, or daemon failure was never captured in the execution evidence.
- No post-removal existence check (e.g. `docker inspect`) was performed to guarantee the container was eradicated.

---

## 2. Engineering Corrections Applied

### 2.1 Direct Host Downloads Prohibited & Minimal Isolated Cache Staged
- **Prohibition of Direct Host Downloads:** Direct execution of `go mod download` on untrusted repository metadata is strictly prohibited. `StageVerifiedDependencies` and `RunCommand` no longer invoke host downloads from scanned workspaces.
- **Independent Checksum & Provenance Verification:**
  - `parseGoModRequirements()` parses `go.mod` and validates `replace` directives. Any directive pointing to host paths (`../`, absolute paths, UNC paths) is rejected with a security violation.
  - `parseGoSumChecksums()` parses `go.sum` and indexes SHA-256 (`h1:`) hashes. Workspaces declaring external dependencies without pinned checksums fail closed immediately.
- **Approved Immutable Cache Source:** Dependencies are sourced exclusively from approved immutable cache artifacts (`APPROVED_DEPENDENCY_CACHE`, `STAGED_MODULE_CACHE`, `CI_DEPENDENCY_CACHE`). If a required dependency is absent from the approved cache, host download is forbidden and execution fails closed with `DYNAMIC_VERIFICATION_UNAVAILABLE`.
- **Minimal Disposable Cache Directory:** `prepareMinimalDependencyCache()` creates an ephemeral, disposable directory (`os.MkdirTemp("", "ucs503-minmod-*")`) and copies **only** the specific verified module versions required by the target workspace.
- **Strict Read-Only Container Mount:** Only the minimal disposable cache is mounted into the container: `-v <disposableDir>:/go/pkg/mod:ro` with `-e GOPATH=/go`. The host global module cache is never exposed. Upon completion, deferred cleanup (`os.RemoveAll`) destroys the disposable cache.
- **Offline Toolchain Environment:** Containers enforce `-e GOPROXY=off`, `-e GOSUMDB=off`, `-e GONOSUMDB=*`, and `-e GOTOOLCHAIN=local` alongside `--network none`.
- **Fail-Closed Propagation:** Staging errors are propagated up to oracles (`verifySyntaxCompilation`, `verifyRegressionTests`, `verifyDefectReproducer`, `verifyBehavioralInvariance`), which classify them as `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`).

### 2.2 Synchronous Container Cleanup & Existence Verification
- **Pre-Evidence Synchronous Execution:** Container removal executes synchronously *before* `ContainerExecutionEvidence` is assembled.
- **Granular Cleanup Status:** `performVerifiedContainerCleanup()` distinguishes:
  - `VERIFIED_REMOVAL: container <id> successfully removed and verified absent` (confirmed via `docker inspect` failure / no such container).
  - `CLEANUP_FAILURE: container <id> still exists after removal (state: <status>)`.
  - `CLEANUP_TIMEOUT: container <id> removal exceeded 5s deadline`.
  - `DAEMON_UNAVAILABLE: failed to remove container <id>: <error>`.
- **Pluggable Command Execution Hook:** Added `ExecCommand CommandExecutor` to `DockerContainerRunner` to enable deterministic unit testing of normal completion, cancellation, timeout, and daemon failure.

---

## 3. Unit Test Verification Evidence

The new unit tests in `go-backend/agent/repo_intel/execution/adversarial_execution_test.go` confirm:

1. `TestContainerCleanup_NormalCompletionVerifiedRemoval`: Synchronous removal verified absent before evidence assembly (`VERIFIED_REMOVAL`).
2. `TestContainerCleanup_CancellationAndTimeout`:
   - Cancellation: exit status `-1`, `TimeoutCause: execution canceled`, verified removal recorded.
   - Timeout: exit status `-1`, `TimeoutCause: context deadline exceeded`, verified removal recorded.
   - Cleanup timeout: `CLEANUP_TIMEOUT` recorded when removal deadline expires.
3. `TestContainerCleanup_DockerDaemonFailure`:
   - Daemon offline: `DAEMON_UNAVAILABLE` recorded.
   - Container lingering after rm: `CLEANUP_FAILURE` recorded.
4. `TestDependencyStaging_ProhibitsArbitraryHostDownloadsAndEnforcesMinimalCache`:
   - Missing `go.sum` fails closed with `DYNAMIC_VERIFICATION_UNAVAILABLE`.
   - Unsafe `replace` directive targeting host path is rejected with security violation.
   - Missing module in approved cache fails closed without host network download.
   - Standard library module requires zero cache mounts.
   - Approved cache stages **only** the requested module into an isolated disposable directory, never exposing unrequested or secret modules, and completely deletes the directory on cleanup.

All 17 execution test suites passed:
```
ok  github.com/saksham/self-healing-pipeline/agent/repo_intel/execution  35.240s
```

---

## 4. Reconciliation Verdict

Milestone 3.8 Phase 2.1.2 release-blocking security corrections are complete and verified. The trust boundary is fully enforced: zero host downloads from untrusted metadata, zero exposure of the global module cache, strict local toolchain pinning, fail-closed offline classification, and synchronous verified container cleanup.
