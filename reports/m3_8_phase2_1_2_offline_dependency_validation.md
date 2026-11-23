# M3.8 Phase 2.1.2 — Offline Dependency Validation & Lifecycle Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Commit:** `0843ca613c602764ef3f534fc21bed7843486ba0` (with local Phase 2.1.2 stabilization fixes)

---

## 1. Threat Model & Offline Verification Strategy

Dynamic verification of untrusted patches requires strict network denial (`--network none`) to prevent remote code execution, token exfiltration, and SSRF attacks. However, Go module verification typically contacts `proxy.golang.org`.

### 1.1 Invariants Enforced
1. **Container Network Egress Forbidden:** Containers must execute with `--network none` under all circumstances. Network access is never enabled, even for dependency downloads.
2. **Untrusted Host Execution Forbidden:** Untrusted candidate source code is never executed directly on the host toolchain. AST parsing on trusted host is restricted to static syntax tree inspection without code execution.
3. **Verified Dependency Staging Only:** External dependencies are staged on the host **only** when accompanied by a pinned `go.sum` containing cryptographic SHA-256 hashes. If `go.sum` is absent or empty, no network download is attempted.
4. **Read-Only Cache Mount:** The module cache is mounted as `:ro` (`-v <cache>:/go/pkg/mod:ro`). Untrusted container processes cannot corrupt or inject modified packages into the module cache.
5. **Disposable Module Fixtures:** Standalone fixtures without `go.mod` (e.g. synthetic test cases) are initialized with an ephemeral `disposable_fixture` via `EnsureDisposableModule(wsDir)`, permitting standard library compilation without network access.

---

## 2. Oracle Classification Matrix

Under `GOPROXY=off`, Go toolchain behavior is partitioned into distinct oracle statuses:

| Outcome Observed | Underlying Event | Oracle Status | Verification Verdict Impact |
|---|---|---|---|
| Zero errors / warnings | Valid syntax & compilation | `PASSED` | Eligible for verification |
| Syntax / parse error | AST or compiler failure | `FAILED` | Repair marked `FAILED` |
| `module lookup disabled by GOPROXY=off` | Missing offline dependency | `UNAVAILABLE` | Fails closed; verdict `SYNTAX_ONLY_VERIFIED` or abstained |
| `cannot find module providing package` | Missing offline dependency | `UNAVAILABLE` | Fails closed; verdict `SYNTAX_ONLY_VERIFIED` or abstained |
| Context deadline exceeded (timeout) | Execution deadline expired | `UNAVAILABLE` | Fails closed; explicit timeout cause recorded |
| Prohibited command / flag | Security policy violation | `FAILED` (Security Rejection) | Fails closed immediately (zero execution attempts) |

### 2.1 Fail-Closed Integrity Guarantee
A missing dynamic test is **never** converted into an independently verified repair. When dynamic container execution reports `UNAVAILABLE` or `NOT_RUN`:
- The repair status is capped at `SYNTAX_ONLY_VERIFIED` (if static AST inspection succeeded) or `FAILED`.
- The verdict `INDEPENDENTLY_VERIFIED` is strictly gated on all required dynamic oracles executing and passing in the hardened container (`dynamicReproduced && regPassedOrNotRun && !hasUnavailable`).

---

## 3. Container Lifecycle & Cleanup Verification

- Ephemeral container names use cryptographic UUIDs: `ucs503-oracle-<12-hex-chars>`.
- Deferred cleanup executes with a dedicated 5-second background context (`context.Background()`), ensuring containers are forcefully terminated (`docker rm -f`) even when the primary command context times out or is canceled.
- Unit test `TestAdversarial_11_ContainerExecutionTimeoutReportedAsUnavailable` verified timeout classification and cleanup confirmation telemetry.
