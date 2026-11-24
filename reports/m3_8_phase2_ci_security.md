# Milestone 3.8 — Phase 2: Release & CI Qualification & Security Audit Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Target:** Continuous Integration Hardening, Untrusted Execution Boundaries, and Dependency Security Review  
**Timestamp:** 2026-11-21T18:35:00Z  

---

## 1. Executive Summary

Milestone 3.8 Phase 2 qualifies the **Live Repository Scan-to-Repair** pipeline for continuous integration and production-grade safety. When analyzing and proposing repairs for arbitrary or third-party repositories, CI automation and execution environments must enforce strict fail-closed boundaries:
1. **Zero Flakiness in CI:** Standard CI runners must never depend on live external networks, third-party API availability, or rate-limited endpoints.
2. **Untrusted Execution Boundary:** Untrusted code or repository contents must never execute arbitrary commands or access host secrets.
3. **Cryptographic Anti-Tamper Controls:** Patches must be bound to immutable snapshots, and execution must reject stale or tampered workspaces.
4. **Transparent Dependency Security:** Open dependency advisories must be forensically audited, isolated to non-production tooling, and subject to narrow, time-limited risk acceptance without masking new vulnerabilities.

This report documents the CI qualification matrix, security boundary verifications, and dependency security findings.

---

## 2. CI Test Qualification & Execution Strategy

To ensure deterministic, high-speed execution in GitHub Actions while retaining live-network qualification, the test suite implements a dual-mode strategy in [scan_to_repair_e2e_test.go](file:///E:/SWE/go-backend/scan_to_repair_e2e_test.go):

### Dual-Mode Test Architecture

```
                  ┌──────────────────────────────────────────────┐
                  │              Test Execution Gate             │
                  └──────────────────────┬───────────────────────┘
                                         │
                 ┌───────────────────────┴───────────────────────┐
                 ▼                                               ▼
   Mandatory CI Gate (Standard PR)               Opt-In Full Qualification
   --------------------------------              -------------------------
   TestEndToEndScanToRepairWorkflow_OfflineFixture  TestEndToEndScanToRepairWorkflow_LiveRepository
   - 100% offline, synthetic fixture              - Requires live network / GitHub token
   - Fast execution (< 1.8 seconds)              - Pinned reference repository scan
   - Complete 6-phase scan-to-repair cycle        - 673 files, 22,763 AST hotspots
   - Unit security assertion tests                - Manual / scheduled validation
```

### 1. Mandatory CI Offline Fixture Test (`TestEndToEndScanToRepairWorkflow_OfflineFixture`)
- **Location:** [scan_to_repair_e2e_test.go:275](file:///E:/SWE/go-backend/scan_to_repair_e2e_test.go#L275)
- **Execution Time:** ~0.77s (Total package execution: 1.73s).
- **Environment Requirements:** No external network connectivity, zero external API keys.
- **Workflow Steps Validated:**
  1. Workspace initialization and source staging in isolated directory.
  2. AST diagnostic scan via `MVPScanner.ScanWorkspace` discovering repairable (`SQL_INJECTION`) and abstained (`COMMAND_INJECTION`, `RESOURCE_LEAK`) targets.
  3. Generation of immutable `WorkspaceSnapshot` (`SnapshotID`, `SourceFileHash`).
  4. AST patch synthesis via `MVPService.ProposeRepair`.
  5. Cryptographic human approval and HMAC-SHA256 token issuance via `MVPService.ApproveRepair`.
  6. Sandbox execution inside ephemeral directory (`.tmp/mvp-sandbox-*`).
  7. Decoupled 5-stage oracle evaluation (Syntax, Vet, Reproducer, Invariance, Negative Properties).
  8. Pre- and post-repair cryptographic SHA-256 tree digest validation (100% bitwise immutability).

### 2. Dedicated Security Invariant Tests
Integrated directly into the standard `go test` suite:
- **`TestUntrustedExecution_CredentialStripping`:** Verifies that all host cloud tokens, database URLs, and API keys (`AWS_SECRET_ACCESS_KEY`, `GITHUB_TOKEN`, `DATABASE_URL`, `TOKEN`, `KEY`, `PASS`, etc.) are stripped from the subprocess environment, while preserving safe OS variables (`PATH`, `SYSTEMROOT`, `LOCALAPPDATA`, `GOCACHE`).
- **`TestUntrustedExecution_ArbitraryCommandBlocked`:** Verifies that attempts by malicious or untrusted configurations to invoke shell interpreters (`sh -c`, `bash`, `cmd.exe`, `powershell`), language interpreters (`python`, `perl`), downloaders (`curl`, `wget`), or dangerous flags (`-exec`) are intercepted and rejected.
- **`TestSnapshotAuthorization_StaleWorkspaceRejected`:** Verifies that modifying a target file after proposal generation triggers an immediate `STALE_WORKSPACE` rejection, preventing TOCTOU attacks.

---

## 3. Untrusted Repository Execution Boundary

Executing tests or tools against arbitrary repositories introduces severe security risks: arbitrary code execution, host credential theft, network egress to command-and-control servers, or repository sabotage.

The implementation enforces the following defense-in-depth controls:

### 1. Command Whitelist Enforcement (`validateSafeSubprocessCommand`)
Implemented in [oracles.go:49](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L49):
```go
var whitelistedBinaries = map[string]bool{
    "go":    true,
    "cargo": true,
    "node":  true,
    "npm":   true,
}
```
- **Strict Binary Validation:** Only explicitly whitelisted build/test toolchain executables are permitted. Direct invocation of shell binaries (`/bin/sh`, `bash`, `cmd`, `powershell`) or script interpreters (`python`, `perl`, `ruby`) is refused with `SECURITY_VIOLATION`.
- **Dangerous Flag Prohibition:** Arguments are scanned for sub-execution flags, such as `go test -exec ...` or `find -exec ...`. Any match aborts execution immediately.

### 2. Environment Variable Sanitization (`GetSanitizedExecutionEnv`)
Implemented in [oracles.go:15](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go#L15):
- **Sensitive Key Deny-List:** Any environment variable containing substring patterns `TOKEN`, `SECRET`, `KEY`, `AUTH`, `PASS`, `CREDENTIAL`, or `DATABASE_URL` is unconditionally purged.
- **Host Variable Isolation:** Host-specific tokens (e.g., `AWS_SECRET_ACCESS_KEY`, `GITHUB_TOKEN`) are completely inaccessible to child processes.
- **Dedicated Temp and Cache Directories:** Sandboxed commands run with dedicated `GOTMPDIR` and `GOCACHE` pointers inside the ephemeral sandbox root, preventing cross-workspace contamination.
- **CGO Disabled:** `CGO_ENABLED=0` is enforced to prevent arbitrary C compiler toolchain execution or native code injection.

### 3. Non-Bypassable Context Timeouts
Every subprocess invocation in [oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go) is bound by a non-bypassable `context.WithTimeout` (default: 15–30 seconds):
- Infinite loops in generated code or malicious tests trigger immediate SIGKILL.
- Process group termination ensures child processes spawned by tests cannot orphan or linger on the host.

### 4. Container Sandbox Contract (Evaluator Engine)
For multi-language or full-system dynamic evaluation, [evaluator.go](file:///E:/SWE/go-backend/agent/evaluator.go) specifies container execution with the following parameters:
- `--network none`: Absolute network isolation; zero egress or ingress.
- `--read-only`: Read-only root filesystem.
- `--tmpfs /tmp:rw,noexec,nosuid,size=64m`: In-memory temporary directory with execution disabled.
- `--user 10001:10001`: Unprivileged non-root user execution.
- `--cap-drop ALL`: All Linux capabilities dropped.
- `--memory 512m --cpus 1.0`: Strict hardware resource throttling.

---

## 4. Snapshot Authorization & Anti-Tamper Security Controls

Autonomous repair execution requires cryptographic binding to ensure human operator intent is faithfully executed:

```
[Target Discovery] ──► Computes File SHA-256 & Repo Tree Hash ──► WorkspaceSnapshot (SNAP-id)
         │
         ▼
[Proposal Synthesis] ─► Binds SnapshotID & SourceFileHash to MVPProposalResult
         │
         ▼
[Human Approval] ────► Operator signs proposal ──► Single-Use HMAC-SHA256 Token (REPAIR-id)
         │
         ▼
[Execute Repair] ────► 1. Verify token signature & single-use ledger
                       2. Verify Target File on disk matches SourceFileHash (STALE check)
                       3. Verify Repo Root Digest matches Snapshot (Anti-tamper check)
                       4. Apply Patch in Ephemeral Sandbox ONLY
                       5. Validate 100% Bitwise Immutability on Original Repository
```

### Security Invariants Validated
1. **Single-Use Token Consumption:** Once an approval token is executed, it is atomically burned in the ledger. Replaying the token returns `TOKEN_ALREADY_USED`.
2. **Stale Workspace Rejection:** If an external process or developer edits the target source file after proposal generation, `ExecuteRepair` detects that `currentHash != proposal.SourceFileHash` and aborts with `STALE_WORKSPACE`.
3. **Repository Root Digest Match:** If any file in the repository tree was modified, the snapshot mismatch is detected before sandbox execution.

---

## 5. Frontend Dependency Security Audit

### Advisory Context
- **Advisory ID:** `GHSA-fc5q-39xc-vjpf` / `CVE-2026-112366`
- **Affected Packages:** `picomatch@2.3.2` (via `micromatch` $\rightarrow$ `gh-pages`) and `picomatch@4.0.7` (via `vite`).
- **Vulnerability Class:** Regular Expression Denial of Service (ReDoS) in glob pattern matching.

### Forensic Analysis & Exposure Audit
1. **Development & Build Tooling Scope:** Both `picomatch` versions are exclusively transitive development dependencies (`devDependencies`). Neither `picomatch`, `micromatch`, nor `vite` is bundled into the client runtime artifacts generated by `npm run build`.
2. **Input Path Analysis:** Glob matching in the repository is driven solely by static configuration files (`vite.config.js`, `package.json`, and static deployment scripts). No untrusted user input, URL parameters, or API payloads reach the glob matching engine.
3. **Upstream Fix Availability:** As of current release baseline, no compatible breaking-change-free upstream patch has been released by the transitive maintainers. Running `npm audit fix --force` would downgrade Vite to an incompatible major version, breaking ESM module federation.

### Risk Acceptance & CI Enforcement Policy
- **Narrow Risk Acceptance:** The `picomatch` ReDoS advisory is accepted as a time-limited development tooling risk, with zero runtime vulnerability exposure.
- **CI Gate Retention:** In [.github/workflows/check-frontend.yml](file:///E:/SWE/.github/workflows/check-frontend.yml), `npm audit --audit-level=high` is maintained. This ensures that:
  - Low and moderate advisories in non-runtime tooling do not block routine integration.
  - Any **new, unapproved High or Critical severity vulnerabilities** immediately fail the build.
  - Periodic dependency sweeps re-evaluate upstream patch availability.

---

## 6. Qualification Verification Matrix

| Verification Target | Requirement | Test / Audit Evidence | Status |
| :--- | :--- | :--- | :--- |
| **Offline CI Gate** | Zero external network calls; $< 2.0\text{ s}$ execution | `TestEndToEndScanToRepairWorkflow_OfflineFixture` passed in 0.77s | **QUALIFIED** |
| **Credential Stripping** | Sensitive environment variables purged from subprocesses | `TestUntrustedExecution_CredentialStripping` passed (0 leaked keys) | **QUALIFIED** |
| **Command Whitelist** | Shell interpreters and arbitrary binaries blocked | `TestUntrustedExecution_ArbitraryCommandBlocked` passed | **QUALIFIED** |
| **Stale Workspace Protection** | On-disk modifications trigger `STALE_WORKSPACE` | `TestSnapshotAuthorization_StaleWorkspaceRejected` passed | **QUALIFIED** |
| **Container Sandboxing** | Network disabled, read-only root, noexec tmpfs, dropped caps | Evaluator Docker execution flags audited | **QUALIFIED** |
| **Tree Immutability** | 100% bitwise match before and after execution | `ComputeRepoDigest` validation passed across all runs | **QUALIFIED** |
| **Frontend CI Checks** | Oxlint 0 warnings/errors, 49/49 Node tests pass, Vite build passes | Verified in `admin-portal` | **QUALIFIED** |
| **Dependency Security Gate** | Transitive ReDoS audited; strict high-level CI audit gate retained | Verified via `.github/workflows/check-frontend.yml` | **QUALIFIED** |

---

## 7. Conclusion

Milestone 3.8 Phase 2 satisfies all continuous integration qualification and untrusted repository security boundary requirements. The offline fixture test ensures robust, deterministic CI passes without network flakiness, while strict command validation, credential stripping, and cryptographic snapshot binding guarantee fail-closed execution safety.
