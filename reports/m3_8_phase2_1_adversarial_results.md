# Milestone 3.8 Phase 2.1 — Security Adversarial Test Results

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Test Suite:** `go-backend/agent/repo_intel/execution/adversarial_execution_test.go`  
**Status:** 9 / 9 Deterministic Vectors PASSED (100% Fail-Closed)

---

## 1. Executive Summary

Milestone 3.8 Phase 2.1 requires nine deterministic adversarial attack scenarios covering container runtime containment, boundary protection, credential defense, snapshot integrity, and cryptographic authorization replay. All nine attack scenarios were implemented and validated against the hardened execution pipeline. Every scenario confirmed complete fail-closed behavior with zero host leakage.

---

## 2. Adversarial Vectors & Validation Results

### Vector 1 — Host Sentinel File Read Prevention
- **Attack Scenario:** An adversarial candidate test or reproducer attempts to read a host-private sentinel key file (`sentinel_host_secret.key`) located outside the workspace.
- **Enforced Defense:** In the absence of container isolation, host toolchain execution is strictly forbidden; dynamic checks fail closed with `DYNAMIC_VERIFICATION_UNAVAILABLE`. Under container execution, `DockerContainerRunner` mounts strictly the sandbox directory with `--read-only` root; host roots and parent directories are never mounted.
- **Test:** `TestAdversarial_1_HostSentinelFileAccessPrevented`
- **Result:** **PASSED** (0.09s).

### Vector 2 — External Network Egress & SSRF Blockade
- **Attack Scenario:** Candidate code attempts outbound HTTP/TCP calls (`net.Dial`, `curl`) to cloud metadata endpoints (`169.254.169.254`) or malicious C2 servers.
- **Enforced Defense:** `DockerContainerRunner` enforces `--network none` in container launch arguments. Dynamic verification fails closed if attempted without container isolation.
- **Test:** `TestAdversarial_2_ExternalNetworkEgressBlocked`
- **Result:** **PASSED** (0.02s).

### Vector 3 — Unauthorized Subprocess & Privilege Escalation Blockade
- **Attack Scenario:** Candidate code executes arbitrary shell interpreters (`sh`, `bash`, `cmd.exe`) or attempts container privilege escalation via setuid binaries.
- **Enforced Defense:** Host `exec.CommandContext` toolchain execution on untrusted repositories is removed. Container runner sets `--security-opt no-new-privileges`, `--cap-drop ALL`, `--user 10001:10001`, and bounded quotas (`--pids-limit 128`, `--memory 1024m`, `--cpus 2.0`).
- **Test:** `TestAdversarial_3_UnauthorizedSubprocessExecutionBlocked`
- **Result:** **PASSED** (0.00s).

### Vector 4 — Host Credential Location Access Rejection
- **Attack Scenario:** Target paths attempt path traversal into sensitive host credential directories (`~/.aws/credentials`, `~/.ssh/id_rsa`, `/etc/shadow`, Windows UNC credential paths).
- **Enforced Defense:** `ValidatePathSafety` and `ValidateContainedPath` explicitly reject home directory expansions (`~`), directory traversal (`..`), absolute paths, and UNC shares. No host user directories are mounted into the container.
- **Test:** `TestAdversarial_4_HostCredentialsAccessRejected`
- **Result:** **PASSED** (0.00s).

### Vector 5 — Resource Quota Exhaustion Enforcement
- **Attack Scenario:** Malicious code attempts infinite compilation loops, memory exhaustion, fork bombs, or blocking sleep commands.
- **Enforced Defense:** `DockerContainerRunner` sets explicit memory limits (`1024MB`), CPU quotas (`2.0`), PID limits (`128`), and context deadline timeouts (`30s`). Context cancellation immediately terminates the container process.
- **Test:** `TestAdversarial_5_ResourceQuotaExhaustionEnforced`
- **Result:** **PASSED** (0.02s).

### Vector 6 — Malformed Workspace & Snapshot Identifier Rejection
- **Attack Scenario:** Attacker submits malformed workspace or snapshot identifiers containing path traversal (`../../etc`), UNC paths (`\\?\UNC\...`), null bytes (`ws\x00`), or non-hex snapshot digests (`SNAP-ZZZZ`).
- **Enforced Defense:** `ValidateWorkspaceID` and `ValidateSnapshotID` enforce strict lexical and character validation. All malformed inputs return immediate `MALFORMED_WORKSPACE_ID` or `MALFORMED_SNAPSHOT_ID` errors.
- **Test:** `TestAdversarial_6_MalformedWorkspaceAndSnapshotIDsRejected`
- **Result:** **PASSED** (0.00s).

### Vector 7 — Approval Token Cross-Workspace Replay Rejection
- **Attack Scenario:** An attacker obtains a cryptographically valid `ApprovalToken` for workspace `ws-1` and attempts to use it to authorize or execute a repair in workspace `ws-2`.
- **Enforced Defense:** `ApprovalAuthority` binds tokens to `WorkspaceID`, `TargetFileHashes`, and `PatchDiffHash`. Replay across workspaces is rejected with `FAIL_AUTH` / `ErrWorkspaceMismatch`.
- **Test:** `TestAdversarial_7_ApprovalTokenReplayAcrossWorkspacesRejected`
- **Result:** **PASSED** (0.02s).

### Vector 8 — Post-Approval Source File Mutation Rejection
- **Attack Scenario:** A source file is legitimately approved, but altered on disk before `ExecuteRepair` runs (TOCTOU / race condition).
- **Enforced Defense:** `ExecuteRepair` re-computes `ComputeRepoDigest` and source file hashes against immutable scan snapshot manifests. When any modification is detected, execution fails closed with `STALE_WORKSPACE`.
- **Test:** `TestAdversarial_8_PostApprovalSourceFileMutationRejected`
- **Result:** **PASSED** (0.01s).

### Vector 9 — Symlink Escape Detection & Rejection
- **Attack Scenario:** A file inside the workspace is replaced with a symbolic link pointing to a file outside the workspace boundary.
- **Enforced Defense:** `ValidateContainedPath` evaluates symlinks via `filepath.EvalSymlinks` against the canonical workspace directory, detecting and rejecting any symlink that escapes the boundary.
- **Test:** `TestAdversarial_9_SymlinkEscapeDetectedAndRejected`
- **Result:** **PASSED** (0.01s).

---

## 3. Comprehensive Summary Matrix

| ID | Adversarial Vector | Security Boundary | Enforcement Code | Verdict |
| :---: | :--- | :--- | :--- | :---: |
| ADV-01 | Host Sentinel Read | Container Isolation / No Host Fallback | `oracles.go:470` | **PASSED** |
| ADV-02 | Network Egress / SSRF | Container Network Policy (`--network none`) | `oracles.go:186` | **PASSED** |
| ADV-03 | Subprocess Hijack | Container Capabilities (`--cap-drop ALL`) | `oracles.go:193` | **PASSED** |
| ADV-04 | Credential Path Access | Component-Aware Path Containment | `workspace.go:70` | **PASSED** |
| ADV-05 | Quota Exhaustion | Container Limits & Bounded Context | `oracles.go:167` | **PASSED** |
| ADV-06 | Malformed IDs | ID Lexical Validators | `mvp_service.go:163` | **PASSED** |
| ADV-07 | Cross-Workspace Replay | Cryptographic Scope Binding | `approval.go:244` | **PASSED** |
| ADV-08 | Post-Approval Tamper | Pre-Execution Tree Digest Comparison | `mvp_service.go:962` | **PASSED** |
| ADV-09 | Symlink Traversal | `filepath.EvalSymlinks` Boundary Check | `workspace.go:324` | **PASSED** |
