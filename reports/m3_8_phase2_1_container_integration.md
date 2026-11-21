# Milestone 3.8 Phase 2.1 — Hardened Container Integration Architecture

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Status:** COMPLETE & VERIFIED

---

## 1. Architectural Overview

To eliminate the risks of arbitrary code execution, network exfiltration (SSRF), privilege escalation, and host filesystem tampering during untrusted repository verification, all dynamic execution operations are encapsulated within an immutable container isolation boundary.

The container integration defines a strict interface contract (`ContainerRunner`) implemented by `DockerContainerRunner` in `go-backend/agent/repo_intel/execution/oracles.go`.

```go
type ContainerRunner interface {
    CheckAvailable(ctx context.Context) error
    RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error)
}
```

When `MVPService` runs, it injects the configured `ContainerRunner` into `IndependentOracleSuite`. When the container runtime or Docker daemon is absent, the system fails closed immediately with `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`). No fallback to host execution is permitted under any circumstances.

---

## 2. Container Isolation Specification & Arguments

The `DockerContainerRunner.RunCommand` implementation generates container execution instances with the following defense-in-depth parameters:

```bash
docker run --rm \
  --name ucs503-oracle-<uuid> \
  --network none \
  --read-only \
  --tmpfs /tmp:rw,exec,nosuid,nodev,mode=1777,size=256m \
  -e GOCACHE=/tmp/gocache \
  -e CGO_ENABLED=0 \
  -e GOTMPDIR=/tmp \
  --user 10001:10001 \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --memory 1024m \
  --cpus 2.00 \
  --pids-limit 128 \
  -v <sandboxWorkspace>:/workspace:rw \
  -w /workspace \
  golang:1.21-alpine \
  <command>
```

### Security Controls Breakdown:

| Directive | Configuration | Enforcement Rationale |
| :--- | :--- | :--- |
| **Network Isolation** | `--network none` | Completely prevents outbound internet connections, exfiltration of source code or credentials, and internal network / metadata API scanning (`169.254.169.254`). |
| **Root Filesystem** | `--read-only` | Prevents persistence of malicious payloads or modification of container system binaries. |
| **Temporary Storage** | `--tmpfs /tmp:...` | Ephemeral, in-memory scratch space configured with `nosuid,nodev` flags for Go compilation caches (`/tmp/gocache`). Automatically discarded on container termination. |
| **Unprivileged Identity** | `--user 10001:10001` | Runs as an unprivileged UID/GID inside the container, preventing root exploits. |
| **Capability Dropping** | `--cap-drop ALL` | Strips all Linux capabilities (`CAP_NET_RAW`, `CAP_SYS_ADMIN`, `CAP_DAC_OVERRIDE`, etc.). |
| **Privilege Escalation** | `--security-opt no-new-privileges` | Disallows `setuid` / `setgid` binaries from escalating process privileges. |
| **Resource Quotas** | `--memory 1024m`, `--cpus 2.0`, `--pids-limit 128` | Mitigates resource exhaustion attacks, fork bombs, and compiler CPU spinning. |
| **Context Timeout** | Bounded Context (`30s`) | Cancels command execution and kills the container process if execution exceeds deadline. |
| **Mount Safety** | Strict Workspace Isolation | Only the isolated sandbox copy (`<sandboxWorkspace>`) is mounted. The original mutable repository root, host root (`/`, `C:\`), and the Docker daemon socket (`/var/run/docker.sock`) are explicitly validated and rejected if targeted. |

---

## 3. Container Execution Telemetry

Every container execution captures complete execution evidence returned in `ContainerExecutionEvidence`:

```json
{
  "container_id": "ucs503-oracle-4e7a2b910f1c",
  "isolation_policy": "DOCKER_HARDENED_CONTAINER (network=none, read-only-root, tmpfs=/tmp, cap-drop=ALL, no-new-privs, non-root=10001:10001)",
  "command": "go test -v ./...",
  "exit_status": 0,
  "memory_limit_mb": 1024,
  "cpu_limit": 2.0,
  "network_policy": "DISABLED (--network none)",
  "root_fs_policy": "READ_ONLY (--read-only)"
}
```

This telemetry is embedded within `MVPOracleOutcome.ContainerInfo` and transmitted to the frontend audit trail, enabling verification auditors to confirm that all test runs adhered to container isolation boundaries.

---

## 4. Fail-Closed Verification Semantics

When dynamic execution cannot be verified within a container (e.g. Docker daemon stopped, container image missing, or runner not configured):
1. The static AST syntax checks (`OracleSyntaxCompilation`) and negative property checks (`OracleNegativeProperties`) pass on the trusted host process.
2. The dynamic oracles (`OracleDefectReproducer`, `OracleRegressionTest`, `OracleBehavioralInvariance`) record `Status = "UNAVAILABLE"` and `Passed = false`.
3. The overall repair verdict is downgraded to `SYNTAX_ONLY_VERIFIED`.
4. The system never falls back to uncontained host toolchain execution.
