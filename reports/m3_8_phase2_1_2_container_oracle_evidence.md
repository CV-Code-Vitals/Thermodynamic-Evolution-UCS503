# M3.8 Phase 2.1.2 — Container Oracle & Policy Evidence

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Review Baseline Commit:** `a0928fba56d81f90241460788322e89cdedc7334`  
**Audit Scope:** Container execution telemetry, command policy enforcement, pre-evidence cleanup lifecycle, and inspection verification.

---

## 1. Zero-Execution Command Policy Verification

`DockerContainerRunner.RunCommand()` validates all commands at the application policy boundary **prior to** checking container availability, incrementing execution counters, or spawning containers.

### 1.1 Policy Rules
- **Whitelisted Binary:** Only `go` / `go.exe` is permitted.
- **Whitelisted Subcommands:** Only `vet`, `test`, and `build` are permitted.
- **Blocked Binaries:** All shells (`sh`, `bash`, `zsh`, `cmd`, `powershell`, `pwsh`), interpreters (`python`, `perl`, `ruby`, `node`, `php`), network utilities (`curl`, `wget`, `nc`, `socat`, `ssh`), package managers (`apt`, `apk`, `yum`, `docker`, `sudo`), and destructive binaries (`rm`, `dd`, `del`).
- **Forbidden Flags:** Flag inspection explicitly blocks `-exec`, `--exec`, `-toolexec`, `--toolexec`, `-plugin`, `--plugin`.

### 1.2 Zero Execution Attempt Evidence
Unit test `TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands` evaluated 14 prohibited vectors (e.g. `sh -c`, `bash -c`, `curl`, `wget`, `rm -rf`, `go test -exec`, `go run`).  
**Outcome:** All 14 vectors were rejected with `security violation: command policy rejection`.  
**Execution Counter:** `atomic.LoadInt64(&runner.ExecutionAttempts) == 0`.

---

## 2. Pre-Evidence Synchronous Container Cleanup Lifecycle

Previously, container cleanup ran in a `defer` block after `ContainerExecutionEvidence` was constructed, causing `evidence.CleanupResult` to record `"NOT_NEEDED"` before removal occurred.

### 2.1 Corrected Lifecycle Architecture
The container lifecycle is strictly ordered as follows:
```
1. Policy Validation (validateSafeSubprocessCommand)
2. Workspace Containment Check (reject sensitive host mounts)
3. Disposable Module Generation (EnsureDisposableModule)
4. Minimal Dependency Cache Staging (prepareMinimalDependencyCache)
5. Runtime Availability Verification (CheckAvailable)
6. Counter Increment (atomic.AddInt64)
7. Subprocess Execution (d.exec)
8. SYNCHRONOUS CONTAINER CLEANUP & INSPECT (performVerifiedContainerCleanup)
9. Evidence Construction (ContainerExecutionEvidence populated with actual CleanupResult)
10. Return ContainerRunResult
```

### 2.2 Granular Cleanup Status Matrix
`performVerifiedContainerCleanup()` reports granular, auditable statuses:

| Status Prefix | Description | Detection Mechanism |
|---|---|---|
| `VERIFIED_REMOVAL` | Container successfully stopped, removed, and verified absent | `docker rm -f` followed by `docker inspect` returning error ("no such container") |
| `CLEANUP_FAILURE` | Container removal failed or container still present | `docker inspect` indicates container remains in state (e.g. `running`, `created`) |
| `CLEANUP_TIMEOUT` | Cleanup operation exceeded 5s deadline | Context timeout on cleanup context |
| `DAEMON_UNAVAILABLE` | Docker daemon unreachable during cleanup | Daemon connection error from docker CLI |

---

## 3. Container Execution Evidence Schema

Every container run generates an immutable audit record:

```json
{
  "container_id": "ucs503-oracle-4a8b1c9d2e3f",
  "isolation_policy": "DOCKER_HARDENED_CONTAINER (network=none, read-only-root, tmpfs=/tmp, cap-drop=ALL, no-new-privs, non-root=10001:10001)",
  "command": "go vet ./...",
  "exit_status": 0,
  "memory_limit_mb": 1024,
  "cpu_limit": 2.0,
  "network_policy": "DISABLED (--network none)",
  "root_fs_policy": "READ_ONLY (--read-only)",
  "workspace_dir": "E:/SWE/go-backend/uploads/workspace_test",
  "image": "ucs503-sandbox-runner:latest",
  "image_digest": "sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92",
  "timeout_cause": "",
  "required_dependencies": [
    "golang.org/x/sys v0.6.0"
  ],
  "cleanup_result": "VERIFIED_REMOVAL: container ucs503-oracle-4a8b1c9d2e3f successfully removed and verified absent"
}
```

### 3.1 Pinned Image Digest Resolution
`resolveImageDigest()` inspects the runtime image (`docker image inspect --format {{index .RepoDigests 0}}` falling back to `{{.Id}}`) to record the exact cryptographic image digest in execution evidence.

---

## 4. Verification Test Evidence

The following deterministic unit tests in `adversarial_execution_test.go` verify container lifecycle and cleanup behavior:

1. `TestContainerCleanup_NormalCompletionVerifiedRemoval`: Confirms `VERIFIED_REMOVAL` and command ordering (`run` -> `rm` -> `inspect`).
2. `TestContainerCleanup_CancellationAndTimeout`:
   - `ContextCancellation`: exit `-1`, `TimeoutCause: canceled`, `VERIFIED_REMOVAL`.
   - `CommandExecutionTimeout`: exit `-1`, `TimeoutCause: deadline exceeded`, `VERIFIED_REMOVAL`.
   - `CleanupOperationTimeout`: `CLEANUP_TIMEOUT` recorded when removal context expires.
3. `TestContainerCleanup_DockerDaemonFailure`:
   - `DaemonUnavailableDuringCleanup`: `DAEMON_UNAVAILABLE` recorded.
   - `ContainerStillExistsAfterRemoval`: `CLEANUP_FAILURE` recorded.
