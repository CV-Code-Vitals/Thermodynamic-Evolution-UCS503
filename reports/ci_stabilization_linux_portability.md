# Linux Filesystem Portability Remediation Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Audit Target:** Cross-Platform Filesystem Safety & Linux CI Runner Portability  
**Date:** 2026-11-23  
**Engineer:** Antigravity Autonomous Systems Engineering Team  

---

## 1. Problem Statement & Baseline Failures

Prior GitHub Actions CI runs on Ubuntu Linux runners (`ubuntu-latest`) encountered three fatal errors during execution of the Go backend test suite and remediation endpoints:

1. **`E:\SWE\.tmp` directory missing on Ubuntu runners:**
   - In Ubuntu CI runners, Windows drive specifications (e.g. `E:\...`) do not exist. Calls attempting to write or read from `E:\SWE\.tmp` failed with `no such file or directory` / `ENOENT`.
2. **`E:\SWE/benchmarks/external_v2` directory missing:**
   - Functions resolving test fixtures defaulted to `"E:\\SWE"` when neither `..` nor `.` directly contained `benchmarks/external_v2`, causing repository clone and digest operations to fail on POSIX runners where the repository root was `/home/runner/work/...`.
3. **`.tmp` directory absent during `POST /api/remediation/propose`:**
   - When `ProposeRepair` attempted to create an ephemeral sandbox workspace via `os.MkdirTemp(filepath.Join(s.WorkspaceRoot, ".tmp"), "mvp-sandbox-*")`, it failed if `<workspace>/.tmp` had not been explicitly pre-created on disk.

---

## 2. Root Cause Analysis

A forensic code audit revealed several locations where Windows-specific filesystem paths or assumptions were hardcoded:

1. **`remediation_routes.go`:**
   `resolveWorkspaceRoot()` previously checked `os.Getenv("WORKSPACE_ROOT")`, then `..` and `.`, and defaulted to `"E:\\SWE"`. If tests ran from subdirectories, the fallback was selected and failed immediately on Linux.
2. **`mvp_service.go`:**
   - `NewMVPService` defaulted `workspaceRoot` to `"E:\\SWE"` and `persistLedgerPath` to `E:\SWE\.tmp\...`.
   - `ProposeRepair` assumed `.tmp` already existed on disk when calling `os.MkdirTemp`.
   - `ExecuteRepair` lacked fallback discovery for authentic repository fixtures.
3. **`oracles.go`:**
   `GetSanitizedExecutionEnv()` previously set `tmpBase := "E:\\SWE\\.tmp"` and appended `GOTMPDIR`, `TEMP`, and `TMP` pointing to that Windows path. On Linux runners, this caused `go` toolchain commands to attempt writing compiler temporary files to a nonexistent Windows path.
4. **`workspace.go`:**
   `ResolveWorkspaceRoot()` contained an explicit fallback check `if fi, err := os.Stat("E:\\SWE"); err == nil && fi.IsDir() { return "E:\\SWE" }`, which caused environment-specific resolution leaks.
5. **`mvp_scanner.go`:**
   `bindTargetSnapshot` previously joined `target.TargetFile` without verifying whether diagnostic engines emitted an absolute or relative path, risking path concatenation failures on Linux runners.
6. **Test Fixtures:**
   - `approval_test.go`, `oracles_test.go`, `synthesis_test.go`, `milestone3_7_phase2_pilot_test.go`, and `milestone3_8_phase1_1_revalidation_test.go` historically called `os.MkdirTemp("E:\\SWE\\.tmp", ...)`.
   - `synthesis_test.go` wrote reports to hardcoded `"E:\\SWE\\reports\\..."`.

---

## 3. Architecture & Implementation of Portable Solutions

### A. Centralized Workspace Root Discovery (`execution.ResolveWorkspaceRoot`)
In [go-backend/agent/repo_intel/execution/workspace.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/workspace.go), `ResolveWorkspaceRoot()` was refactored to rely exclusively on upward landmark traversal and environment configuration:

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
            if fi, err := os.Stat(filepath.Join(curr, "go-backend", "go.mod")); err == nil {
                abs, err := filepath.Abs(curr)
                if err == nil { return abs }
                return curr
            }
            parent := filepath.Dir(curr)
            if parent == curr || parent == "" { break }
            curr = parent
        }
    }

    // 3. Fallback to current working directory or system temp
    if wd != "" {
        abs, err := filepath.Abs(wd)
        if err == nil { return abs }
        return wd
    }
    return os.TempDir()
}
```
All hardcoded Windows fallbacks (`"E:\\SWE"`) were completely excised.

### B. Fail-Safe Temporary Directory Management (`execution.EnsureTmpDir`)
Added to [go-backend/agent/repo_intel/execution/workspace.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/workspace.go):

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

### C. Dynamic Subprocess Toolchain Environment (`oracles.go`)
In [go-backend/agent/repo_intel/execution/oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go), `GetSanitizedExecutionEnv()` now dynamically resolves the temporary directory using `EnsureTmpDir(ResolveWorkspaceRoot())`:

```go
    // Ensure isolated temp directories via dynamic workspace resolution
    tmpBase, err := EnsureTmpDir(ResolveWorkspaceRoot())
    if err == nil && tmpBase != "" {
        cleanEnv = append(cleanEnv, `GOTMPDIR=`+tmpBase, `TEMP=`+tmpBase, `TMP=`+tmpBase)
    } else {
        tmpBase = os.TempDir()
    }

    // Ensure GOCACHE is always set to prevent go vet / build cache failures
    if !hasGocache {
        gocacheDir := filepath.Join(tmpBase, "gocache")
        _ = os.MkdirAll(gocacheDir, 0755)
        cleanEnv = append(cleanEnv, "GOCACHE="+gocacheDir)
    }
```
On Ubuntu Linux runners, this points safely to the runner's native temporary hierarchy without crashing.

### D. Target Path Sanitization (`mvp_scanner.go`)
In [go-backend/agent/repo_intel/execution/mvp_scanner.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/mvp_scanner.go), `bindTargetSnapshot` now explicitly converts absolute target paths returned by diagnostic engines into workspace-relative paths:

```go
    targetRel := target.TargetFile
    if filepath.IsAbs(targetRel) {
        if rel, err := filepath.Rel(wsDir, targetRel); err == nil && !strings.HasPrefix(rel, "..") {
            targetRel = rel
        }
    }
    sourcePath := filepath.Join(wsDir, targetRel)
```
This prevents malformed cross-platform path concatenations like `/home/runner/work/repo//home/runner/work/repo/...`.

### E. Elimination of Test Sandbox Leaks
All test cases across `approval_test.go`, `oracles_test.go`, `synthesis_test.go`, `milestone3_7_phase2_pilot_test.go`, and `milestone3_8_phase1_1_revalidation_test.go` use Go's standard `t.TempDir()`. This guarantees:
- Native pathing on both Windows and POSIX.
- Automatic creation and cleanup.
- Zero leftover disk clutter or permission collisions between test runs.
- Zero tests skipped on Linux.

---

## 4. Verification Verdict

All Go backend tests compile, pass `go vet`, and execute successfully across packages without relying on hardcoded paths, skipped tests, or platform-specific directory conventions.
