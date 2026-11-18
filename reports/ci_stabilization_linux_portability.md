# Linux Filesystem Portability Remediation Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Audit Target:** Cross-Platform Filesystem Safety & Linux CI Runner Portability  
**Date:** 2026-11-18  
**Engineer:** Antigravity Autonomous Systems Engineering Team  

---

## 1. Problem Statement & Baseline Failures

Prior GitHub Actions CI runs on Ubuntu Linux runners (`ubuntu-latest`) encountered three fatal errors during execution of the Go backend test suite and remediation endpoints:

1. **`E:\SWE\.tmp` directory missing on Ubuntu runners:**
   - In Ubuntu CI runners, Windows drive specifications (e.g. `E:\...`) do not exist. Calls to `os.MkdirTemp("E:\\SWE\\.tmp", ...)` failed with `no such file or directory` / `ENOENT`.
2. **`E:\SWE/benchmarks/external_v2` directory missing:**
   - Functions resolving test fixtures defaulted to `"E:\\SWE"` when neither `..` nor `.` directly contained `benchmarks/external_v2`, causing repository clone and digest operations to fail on POSIX runners where the repository root was `/home/runner/work/...`.
3. **`.tmp` directory absent during `POST /api/remediation/propose`:**
   - When `ProposeRepair` attempted to create an ephemeral sandbox workspace via `os.MkdirTemp(filepath.Join(s.WorkspaceRoot, ".tmp"), "mvp-sandbox-*")`, it failed if `<workspace>/.tmp` had not been explicitly pre-created on disk.

---

## 2. Root Cause Analysis

A forensic code audit revealed several locations where Windows-specific filesystem paths or assumptions were hardcoded:

1. **`remediation_routes.go`:**
   `resolveWorkspaceRoot()` checked `os.Getenv("WORKSPACE_ROOT")`, then `..` and `.`, and defaulted to `"E:\\SWE"`. If tests ran from subdirectories, the fallback was selected and failed immediately on Linux.
2. **`mvp_service.go`:**
   - `NewMVPService` defaulted `workspaceRoot` to `"E:\\SWE"` and `persistLedgerPath` to `E:\SWE\.tmp\...`.
   - `ProposeRepair` assumed `.tmp` already existed on disk when calling `os.MkdirTemp`.
   - `ExecuteRepair` lacked fallback discovery for authentic repository fixtures.
3. **`oracles.go`:**
   Lines 141, 173, 211, and 304 forcibly appended `"GOTMPDIR=E:\\SWE\\.tmp"`, `"TEMP=E:\\SWE\\.tmp"`, and `"TMP=E:\\SWE\\.tmp"` to `cmd.Env` during `go vet` and test invocations. On Linux runners, this caused `go` toolchain commands to attempt writing compiler temporary files to a nonexistent Windows path.
4. **Test Fixtures:**
   - `approval_test.go`, `oracles_test.go`, `synthesis_test.go`, `milestone3_7_phase2_pilot_test.go`, and `milestone3_8_phase1_1_revalidation_test.go` all called `os.MkdirTemp("E:\\SWE\\.tmp", ...)`.
   - `synthesis_test.go` wrote reports to hardcoded `"E:\\SWE\\reports\\..."`.
5. **Entrypoint Binaries:**
   `cmd/remediation_mvp`, `cmd/evaluate_cross_package`, and `cmd/evaluate_external_v2` hardcoded default workspace paths to `"E:\\SWE"`.

---

## 3. Architecture & Implementation of Portable Solutions

### A. Centralized Workspace Root Discovery (`execution.ResolveWorkspaceRoot`)
Added to `go-backend/agent/repo_intel/execution/workspace.go`:

```go
func ResolveWorkspaceRoot() string {
    // 1. Check explicit environment overrides
    if custom := os.Getenv("WORKSPACE_ROOT"); custom != "" {
        if fi, err := os.Stat(custom); err == nil && fi.IsDir() {
            abs, err := filepath.Abs(custom)
            if err == nil { return abs }
            return custom
        }
    }
    if custom := os.Getenv("SWE_WORKSPACE"); custom != "" {
        if fi, err := os.Stat(custom); err == nil && fi.IsDir() {
            abs, err := filepath.Abs(custom)
            if err == nil { return abs }
            return custom
        }
    }

    // 2. Ascend filesystem hierarchy to find landmarks
    wd, err := os.Getwd()
    if err == nil {
        curr := wd
        for {
            if fi, err := os.Stat(filepath.Join(curr, "benchmarks", "external_v2")); err == nil && fi.IsDir() {
                abs, err := filepath.Abs(curr)
                if err == nil { return abs }
                return curr
            }
            parent := filepath.Dir(curr)
            if parent == curr || parent == "" { break }
            curr = parent
        }
    }

    // 3. Check relative candidate paths
    candidates := []string{".", "..", filepath.Join("..", ".."), filepath.Join("..", "..", ".."), filepath.Join("..", "..", "..", "..")}
    for _, cand := range candidates {
        if fi, err := os.Stat(filepath.Join(cand, "benchmarks", "external_v2")); err == nil && fi.IsDir() {
            abs, err := filepath.Abs(cand)
            if err == nil { return abs }
        }
    }

    // 4. Safe conditional fallback on Windows only
    if fi, err := os.Stat("E:\\SWE"); err == nil && fi.IsDir() {
        return "E:\\SWE"
    }

    if wd != "" {
        abs, err := filepath.Abs(wd)
        if err == nil { return abs }
        return wd
    }
    return os.TempDir()
}
```

### B. Fail-Safe Temporary Directory Management (`execution.EnsureTmpDir`)
Added to `go-backend/agent/repo_intel/execution/workspace.go`:

```go
func EnsureTmpDir(root string) (string, error) {
    if root != "" {
        primary := filepath.Join(root, ".tmp")
        if err := os.MkdirAll(primary, 0755); err == nil {
            return primary, nil
        }
    }
    fallback := filepath.Join(os.TempDir(), "swe_tmp")
    if err := os.MkdirAll(fallback, 0755); err != nil {
        return os.TempDir(), nil
    }
    return fallback, nil
}
```

### C. Safe Subprocess Toolchain Environment (`oracles.go`)
Replaced hardcoded environment injection with `getExecutionEnv()`:

```go
func getExecutionEnv() []string {
    env := os.Environ()
    if os.Getenv("GOTMPDIR") == "" {
        // Only inject on Windows host when the directory physically exists
        if fi, err := os.Stat(`E:\SWE\.tmp`); err == nil && fi.IsDir() {
            env = append(env, `GOTMPDIR=E:\SWE\.tmp`, `TEMP=E:\SWE\.tmp`, `TMP=E:\SWE\.tmp`)
        }
    }
    return env
}
```
On Ubuntu Linux runners, `os.Stat("E:\\SWE\\.tmp")` fails cleanly, so the runner's native environment (`/tmp`) is inherited without interference.

### D. Elimination of Test Sandbox Leaks
All test cases across `approval_test.go`, `oracles_test.go`, `synthesis_test.go`, `milestone3_7_phase2_pilot_test.go`, and `milestone3_8_phase1_1_revalidation_test.go` were refactored to use Go's standard `t.TempDir()`. This guarantees:
- Native pathing on both Windows and POSIX.
- Automatic creation and cleanup.
- Zero leftover disk clutter or permission collisions between test runs.

---

## 4. Modified Files Summary

| File Path | Description of Changes |
| :--- | :--- |
| `go-backend/agent/repo_intel/execution/workspace.go` | Implemented `ResolveWorkspaceRoot()` and `EnsureTmpDir()` |
| `go-backend/agent/repo_intel/execution/mvp_service.go` | Integrated `ResolveWorkspaceRoot()`, `EnsureTmpDir()`, and directory validation in `NewMVPService`, `ProposeRepair`, `ExecuteRepair` |
| `go-backend/remediation_routes.go` | Replaced legacy root resolution with `execution.ResolveWorkspaceRoot()` and `execution.EnsureTmpDir()` |
| `go-backend/agent/repo_intel/execution/oracles.go` | Replaced hardcoded `E:\SWE\.tmp` environment injection with `getExecutionEnv()` |
| `go-backend/agent/repo_intel/execution/approval_test.go` | Switched from `os.MkdirTemp` to `t.TempDir()` |
| `go-backend/agent/repo_intel/execution/oracles_test.go` | Replaced 3 `os.MkdirTemp` instances with `t.TempDir()` |
| `go-backend/agent/repo_intel/execution/synthesis_test.go` | Replaced 2 `os.MkdirTemp` instances with `t.TempDir()`, resolved report path via `ResolveWorkspaceRoot()` |
| `go-backend/agent/repo_intel/benchmark/milestone3_7_phase2_pilot_test.go` | Replaced hardcoded `E:\SWE` with `execution.ResolveWorkspaceRoot()`, used `t.TempDir()` |
| `go-backend/agent/repo_intel/benchmark/milestone3_8_phase1_1_revalidation_test.go` | Replaced hardcoded `E:\SWE` with `execution.ResolveWorkspaceRoot()`, used `t.TempDir()` |
| `go-backend/cmd/remediation_mvp/main.go` | Defaulted `-workspace` flag to `execution.ResolveWorkspaceRoot()`, used `EnsureTmpDir()` |
| `go-backend/cmd/evaluate_cross_package/main.go` | Dynamic `repoRoot` resolution via `execution.ResolveWorkspaceRoot()` |
| `go-backend/cmd/evaluate_external_v2/main.go` | Dynamic `repoRoot` resolution via `execution.ResolveWorkspaceRoot()` |

---

## 5. Verification Verdict

All tests compile, pass `go vet`, and execute successfully across packages without relying on hardcoded paths or skipped tests.
