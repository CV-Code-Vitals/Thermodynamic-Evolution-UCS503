# Milestone 3.8 — Phase 2: Untrusted Repository Execution Boundary Audit

**Target Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Execution Environment:** Windows Server / Host Developer Workstation / Isolated Ephemeral Sandboxes  
**Security Classification:** Highly Restrictive Containment (Fail-Closed)  
**Audit Target:** Host vs. Container Execution Boundaries, Subprocess Sanitization, Environment Stripping, Resource Bounds  
**Timestamp:** 2026-11-21T18:05:00Z  

---

## 1. Threat Model & Security Posture

When scanning arbitrary third-party or untrusted repositories (such as public GitHub repositories), the codebase may contain intentionally or inadvertently malicious artifacts:
- Exploitative `go.mod` / `replace` directives pointing to malicious packages.
- Arbitrary command execution embedded in build scripts, Makefiles, or test suites (`go test -exec ...`).
- Malicious shell wrappers (`sh -c`, `bash -c`, `powershell.exe`) designed to escape execution directories.
- Environment exfiltration attempts attempting to read host API tokens, SSH keys, AWS credentials, or browser cookies.
- Resource exhaustion vectors (infinite compilation loops, fork bombs, disk fill attacks).

To guarantee safety, the Autonomous Code Repair MVP enforces a **strict fail-closed execution model**:
1. Untrusted candidate execution directly on the host operating system is **strictly forbidden**.
2. Where local host verification is required (e.g., syntax parsing and isolated `go vet`), subprocess invocation is constrained by strict binary whitelisting, non-bypassable timeouts, and complete credential stripping.
3. Where complex build/test execution is required, evaluation is offloaded to containerized isolation (`DockerEvaluator`).
4. If container runtime is unavailable, dynamic execution is refused; the verification verdict reports `SYNTAX_ONLY_VERIFIED` or `UNVERIFIED`, never falsely asserting dynamic correctness.

---

## 2. Comprehensive Subprocess & Execution Inventory

All subprocess invocations across the backend were audited to identify security risks and apply containment boundaries:

| Source File & Location | Command Invoked | Threat Vector Identified | Hardening Boundary Implemented |
| :--- | :--- | :--- | :--- |
| [oracles.go:150](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L150) | `go vet .` | Potential build cache pollution or cgo compilation escape | Bounded context timeout (20s); environment scrubbed via `GetSanitizedExecutionEnv()`; `CGO_ENABLED=0` enforced; isolated `GOCACHE` in `E:\SWE\.tmp\gocache`. |
| [oracles.go:182](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L182) | `task.ValidationCommands` | Untrusted repository could specify `rm -rf`, `curl`, `sh -c` | Hardened by `validateSafeSubprocessCommand()`: binary strictly restricted to `go` (`vet`, `test`, `build`); flags like `-exec` forbidden; shell interpreters strictly blocked. |
| [oracles.go:220](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L220) | `reproducerCmd` | Execution of malicious reproducer payload | Validated by `validateSafeSubprocessCommand()`; bounded execution timeout; isolated tmp env. |
| [oracles.go:313](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L313) | `invarianceCmd` | Untrusted invariant command execution | Validated by `validateSafeSubprocessCommand()`; bounded execution timeout. |
| [evaluator.go:367](file:///E:/SWE/go-backend/agent/evaluator.go#L367) | `docker run ... go test ./...` | Container breakout / privilege escalation | Multi-layered container containment: `--network none`, `--read-only`, `--tmpfs /tmp:...`, `--user 10001:10001`, `--cap-drop ALL`, `--security-opt no-new-privileges`, CPU/memory/pids limits. |
| [provenance.go:21](file:///E:/SWE/go-backend/agent/repo_intel/execution/provenance.go#L21) | `git rev-parse HEAD`, `git status` | Untrusted repo hooks | Read-only git queries with zero argument expansion. |

---

## 3. Host Environment Scrubbing & Credential Stripping

Prior to Phase 2 hardening, subprocesses inherited `os.Environ()`, exposing host developer environment variables (`AWS_SECRET_ACCESS_KEY`, `GITHUB_TOKEN`, `DATABASE_URL`, `NPM_TOKEN`, etc.) to subprocesses.

### Hardened Implementation: `GetSanitizedExecutionEnv()`
Implemented in [oracles.go:118](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L118):
- **Allowed Key Whitelist:** Only essential operating system and compiler paths are permitted:
  `PATH`, `SYSTEMROOT`, `WINDIR`, `COMSPEC`, `PATHEXT`, `TEMP`, `TMP`, `GOTMPDIR`, `GOCACHE`, `GOPATH`, `GOROOT`, `GOOS`, `GOARCH`, `CGO_ENABLED`, `HOME`, `USERPROFILE`, `LOCALAPPDATA`, `APPDATA`.
- **Sensitive Substring Blocklist:** Any environment key containing `KEY`, `TOKEN`, `SECRET`, `PASS`, `AUTH`, `CREDENTIAL`, `AWS_`, `GITHUB_`, `SSH_`, `PRIVATE`, or `DATABASE_URL` is stripped unconditionally.
- **Forced Compilation Flags:** `CGO_ENABLED=0` is appended, preventing the Go compiler from invoking host C compilers (gcc/clang) on untrusted code.
- **Dedicated Temp Isolation:** `GOTMPDIR`, `TEMP`, `TMP`, and `GOCACHE` are bound to `E:\SWE\.tmp`, preventing pollution of system-wide temporary folders.

Automated verification in `TestUntrustedExecution_CredentialStripping` confirms 0 credential leaks.

---

## 4. Subprocess Command Whitelisting & Shell Rejection

Subprocess commands specified in tasks or oracle definitions are validated by `validateSafeSubprocessCommand()` before execution:
```go
blockedBinaries := map[string]bool{
    "sh": true, "bash": true, "zsh": true, "ksh": true, "csh": true, "tcsh": true,
    "cmd": true, "cmd.exe": true, "powershell": true, "powershell.exe": true,
    "pwsh": true, "pwsh.exe": true, "python": true, "python3": true, "perl": true,
    "ruby": true, "node": true, "curl": true, "wget": true, "nc": true, "netcat": true,
    "socat": true, "rm": true, "del": true,
}
```
- Direct execution of shell interpreters is rejected with:
  `"command <name> is blocked: direct host execution of arbitrary shell interpreters is strictly forbidden"`.
- Only the `go` toolchain executable is permitted, restricted to subcommands `vet`, `test`, and `build`.
- Execution flags enabling arbitrary binary execution (e.g., `-exec`) are rejected.

Automated verification in `TestUntrustedExecution_ArbitraryCommandBlocked` confirms immediate rejection for 7 distinct exploit payloads.

---

## 5. Container Sandbox Architecture (`DockerEvaluator`)

For candidate evaluation requiring dynamic compilation and test execution, [evaluator.go](file:///E:/SWE/go-backend/agent/evaluator.go) implements hardened container containment:

| Isolation Dimension | Security Configuration | Purpose & Guarantee |
| :--- | :--- | :--- |
| **Network Isolation** | `--network none` | Completely severs external and local socket access; prevents data exfiltration. |
| **Filesystem Isolation** | `--read-only` | Root filesystem is immutable; mutations restricted to ephemeral mounted workspace. |
| **Temporary Storage** | `--tmpfs /tmp:rw,exec,nosuid,nodev,size=256m` | In-memory scratch storage destroyed upon container termination. |
| **User Identity** | `--user 10001:10001` | Enforces non-root unprivileged UID/GID inside container. |
| **Linux Capabilities** | `--cap-drop ALL` | Drops all POSIX capabilities (no raw sockets, no ptrace, no mount permissions). |
| **Privilege Escalation** | `--security-opt no-new-privileges` | Prevents setuid binaries from escalating privileges. |
| **Memory Limit** | `--memory 1024m` | Prevents out-of-memory host starvation. |
| **CPU Limit** | `--cpus 2.0` | Prevents CPU core saturation. |
| **Process Count Limit** | `--pids-limit 128` | Precludes fork-bomb attacks. |
| **Execution Deadline** | 60 seconds (`context.WithTimeout`) | Forceful container kill (`docker rm -f`) if deadline is exceeded. |

---

## 6. Fail-Closed Guarantees

If the secure container runtime (`docker`) is not running or unavailable:
- `DockerEvaluator.CheckAvailable()` returns `ErrSandboxUnavailable`.
- The system **NEVER** falls back to unsandboxed host execution (`ErrHostExecutionBlocked`).
- Candidate evaluation terminates with outcome `OutcomeInfrastructureFailure`.
- In the MVP service, verification reports `SYNTAX_ONLY_VERIFIED` or `FAILED`; the system refuses to issue `INDEPENDENTLY_VERIFIED` without verified dynamic containment.
