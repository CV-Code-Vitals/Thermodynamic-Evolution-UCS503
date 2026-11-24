# Milestone 3.8 Phase 2.1 — Host Execution Audit & Elimination Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Starting Commit:** `5bca84940923e3c965f6391826c1d90f9dbe4786`  
**Status:** COMPLETE (Zero Untrusted Code Execution on Host Enforced)

---

## 1. Executive Summary

Milestone 3.8 Phase 2.1 mandates the complete removal of unsandboxed host toolchain execution for scanned and untrusted repositories. In previous iterations, `IndependentOracleSuite` in `go-backend/agent/repo_intel/execution/oracles.go` utilized host-level `exec.CommandContext` to execute `go vet .`, `go test`, defect-specific reproducers, and behavioral invariance commands. While guarded by a command and argument whitelist (`isWhitelistedHostCommand`), a whitelist does not constitute an execution sandbox: an untrusted repository containing malicious Go code (e.g., `init()` functions, Go build directives, cgo directives, or malicious test cases) could execute arbitrary code with the privileges of the host process during compilation or test execution.

This audit confirms that all host toolchain execution of untrusted candidate code has been eliminated. Dynamic verification now requires hardened container isolation (`DockerContainerRunner`). When container containment cannot be established, dynamic verification fails closed with `DYNAMIC_VERIFICATION_UNAVAILABLE`. Trusted AST parsing remains bounded within host memory without executing untrusted target binaries.

---

## 2. Forensic Inventory of Previous Host Invocations

An exhaustive audit of `oracles.go`, `mvp_service.go`, and related execution files identified four primary vectors where host-level subprocess execution previously occurred:

| File | Function / Site | Previous Mechanism | Threat Vector | Remediation Applied |
| :--- | :--- | :--- | :--- | :--- |
| `oracles.go` | `verifySyntaxCompilation` | Host `exec.CommandContext("go", "vet", ".")` | Malicious `init()` in package or macro execution during vetting. | Executed strictly via `s.Runner.RunCommand` inside container; fails closed if container unavailable. Host performs AST parse only. |
| `oracles.go` | `verifyRegressionTests` | Host `exec.CommandContext(binary, args...)` for `task.ValidationCommands` | Untrusted test suites executing host shell commands, accessing host filesystem, or network. | Executed strictly in container with `--network none`, `--read-only`, and non-root user. Zero host fallback. |
| `oracles.go` | `verifyDefectReproducer` | Host `exec.CommandContext(binary, args...)` for `reproducerCmd` | Execution of arbitrary reproducer scripts or test binaries on host OS. | Routed exclusively to container runner. Non-container execution limited to trusted host AST static checks. |
| `oracles.go` | `verifyBehavioralInvariance` | Host `exec.CommandContext(binary, args...)` for `invarianceCmd` | Invariance verification running adversarial commands against host. | Routed exclusively to container runner; fails closed if container unavailable. |

---

## 3. Trusted Host AST Boundary vs. Untrusted Dynamic Execution

To balance safety with performance, the boundary between trusted host analysis and untrusted dynamic execution is strictly partitioned:

```
+-----------------------------------------------------------------------------------+
|                                  TRUSTED HOST PROCESS                             |
|                                                                                   |
|  [Source Code] ---> parser.ParseFile(fset, path, nil, parser.ParseComments)       |
|                            |                                                      |
|                            v                                                      |
|                   [In-Memory Go AST]                                              |
|                            |                                                      |
|         +------------------+------------------+                                   |
|         |                                     |                                   |
|         v                                     v                                   |
|  Negative Property Oracles         Structural Defect Oracles                     |
|  - Shadowed error checks           - Callee cleanup on escaping resources         |
|  - Inverted conditionals           - Pre-error check cleanup placement            |
|  - Nil-pointer dereference paths   - Redundant Close() calls                      |
|                                                                                   |
|   (Zero Code Execution, Zero Child Processes, Memory-Safe Go Standard Library)    |
+-----------------------------------------------------------------------------------+
                                     |
                       Dynamic Verification Required?
                                     |
                                     v
+-----------------------------------------------------------------------------------+
|                        HARDENED CONTAINER ISOLATION BOUNDARY                      |
|                                                                                   |
|  DockerContainerRunner:                                                           |
|  - Network:           --network none                                              |
|  - Filesystem:        --read-only root + tmpfs /tmp (nosuid, nodev, mode=1777)     |
|  - Privileges:        --user 10001:10001, --cap-drop ALL, no-new-privileges       |
|  - Resource Quotas:   --memory 1024m, --cpus 2.0, --pids-limit 128                |
|  - Workspace Mount:   -v <sandboxDir>:/workspace:rw (Host root/credentials barred)|
|                                                                                   |
|  * If container runtime is unavailable:                                           |
|    FAIL CLOSED -> DYNAMIC_VERIFICATION_UNAVAILABLE (No Host Fallback!)            |
+-----------------------------------------------------------------------------------+
```

### Safety Guarantees:
1. **AST Parsing Without Execution:** `go/parser.ParseFile` constructs the syntax tree by tokenizing text in memory. It does not invoke the compiler, linker, cgo, or `init()` functions.
2. **Whitelist Rejection:** The whitelist logic (`isWhitelistedHostCommand`) is preserved only as an ancillary filter for container arguments; it is explicitly prohibited from authorizing host-level execution.
3. **Fail-Closed Container Requirement:** When `DockerContainerRunner` or Docker daemon is unavailable, dynamic oracles immediately return `OracleStatusUnavailable` with error string `ErrDynamicVerificationUnavailable`. No dynamic test is ever executed on the host.

---

## 4. Verification Evidence

The elimination of host toolchain execution is verified by unit and adversarial test suites:

1. **`TestAdversarial_1_HostSentinelFileAccessPrevented`**:
   - Confirms that running verification without a container runner reports `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`) across dynamic checks and never attempts to read a host sentinel file.
2. **`TestAdversarial_3_UnauthorizedSubprocessExecutionBlocked`**:
   - Confirms that untrusted candidate code cannot spawn host processes or escalate privileges.
3. **`TestControlledFixtureRepairTrials`**:
   - Confirms that synthesized patches undergo AST static checks on the host, while dynamic verification respects container availability.

---

## 5. Conclusion

Host execution of untrusted toolchain commands is completely eliminated. The execution boundary now enforces strict container containment, protecting the host machine, developer workstation, and CI environment from malicious repository payloads.
