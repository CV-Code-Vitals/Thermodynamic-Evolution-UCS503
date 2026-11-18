# CI Hotfix — Forensic Root Cause Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Date:** 2026-11-18  
**Investigator:** Antigravity Autonomous Systems Audit Agent  

---

## 1. Exact CI Environment Baseline

The failing CI workflow run was identified and traced to the continuous integration pipeline configured in [`.github/workflows/ci-go.yml`](file:///E:/SWE/.github/workflows/ci-go.yml):

- **Workflow Name:** `CI - Go build & vet`
- **Job:** `go-checks`
- **Runner Environment:** `ubuntu-latest` (Linux x86_64, Ubuntu 24.04 LTS / glibc 2.39)
- **Go Toolchain:** Go `1.25` (configured via `actions/setup-go@v6` with cache dependencies)
- **Rust Toolchain:** Rust `stable` (installed via `rustup` default stable)
- **Target Commit SHA:** [`bb49185`](file:///E:/SWE) (`chore: add milestone 3.8 phase 1.2 reconciliation reports and ground truth data`) and preceding commit [`4039d49`](file:///E:/SWE) (`feat(milestone-3.8): complete Phase 1.2 upstream provenance, benchmark integrity & final MVP claim audit`)
- **Failing Tests:**
  1. `TestWorkspacePathTraversalProtection` (Package: `agent/repo_intel/execution`)
  2. `TestMilestone32PrecisionEvaluation` (Package: `agent/repo_intel/benchmark`)

---

## 2. Failure 1: `TestWorkspacePathTraversalProtection`

### Failure Symptom:
```
=== RUN   TestWorkspacePathTraversalProtection
    execution_test.go:164: Expected path traversal error for absolute path, got nil
--- FAIL: TestWorkspacePathTraversalProtection (0.00s)
```

### Forensic Root Cause:
In [`go-backend/agent/repo_intel/execution/execution_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/execution_test.go#L161-L165):
```go
// Absolute path escape
_, err = wm.ValidateContainedPath(wsDir, "C:\\Windows\\System32")
if err == nil {
    t.Errorf("Expected path traversal error for absolute path, got nil")
}
```

In the pre-fix implementation of [`go-backend/agent/repo_intel/execution/workspace.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/workspace.go#L147-L158):
```go
cleanRel := filepath.Clean(filepath.FromSlash(targetRelPath))
if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
    return "", fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
}

fullPath := filepath.Join(wsDir, cleanRel)
cleanFull := filepath.Clean(fullPath)
cleanWs := filepath.Clean(wsDir)

if !strings.HasPrefix(cleanFull, cleanWs) {
    return "", fmt.Errorf("security violation: path escapes workspace boundary: %s", targetRelPath)
}
```

### Cross-Platform Semantic Vulnerability:
1. **POSIX vs Windows Path Separators:** On a Linux runner (`ubuntu-latest`), `filepath.Separator` is `'/'`. The standard library function `filepath.FromSlash` only converts `'/'` to `filepath.Separator`—it does **not** convert Windows backslashes (`'\'`). Therefore, `"C:\\Windows\\System32"` remains `"C:\\Windows\\System32"` with literal backslashes.
2. **`filepath.IsAbs` Platform Asymmetry:** On POSIX systems, `filepath.IsAbs` strictly checks `strings.HasPrefix(path, "/")`. A Windows drive-letter path such as `"C:\\Windows\\System32"` is considered a relative path on Linux.
3. **`strings.HasPrefix` Sibling and Relative Containment Flaw:** Because `cleanRel` was treated as a relative filename, `filepath.Join("/tmp/ws_123", "C:\\Windows\\System32")` yielded `"/tmp/ws_123/C:\Windows\System32"`. The string prefix check `strings.HasPrefix("/tmp/ws_123/C:\Windows\System32", "/tmp/ws_123")` evaluated to `true`, and the function returned `nil` error.
4. **Result:** On Linux, any Windows drive path (`C:\...`), drive-relative path (`C:foo`), or UNC path (`\\server\share`) was silently accepted as a contained path instead of being rejected at the security boundary.

---

## 3. Failure 2: `TestMilestone32PrecisionEvaluation`

### Failure Symptom:
```
=== RUN   TestMilestone32PrecisionEvaluation
    verification_test.go:90: Expected recall >= 0.90, got 0.875000
--- FAIL: TestMilestone32PrecisionEvaluation (1.00s)
```

### Forensic Root Cause:
In [`go-backend/agent/repo_intel/benchmark/verification_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/verification_test.go#L89-L91):
```go
if rep.OverallRecall.Value < 0.90 {
    t.Errorf("Expected recall >= 0.90, got %f", rep.OverallRecall.Value)
}
```

The benchmark suite defines 96 total defects across 8 categories (12 defects per category):
$$\text{Recall} = \frac{\text{True Positives}}{\text{True Positives} + \text{False Negatives}} = \frac{84}{84 + 12} = \frac{84}{96} = 0.8750$$

Inspection of `reports/milestone3_2_detection_metrics.json` revealed the exact category with false negatives:
- `ARCHITECTURE`: TP=12, FN=0
- `DEPENDENCY`: TP=12, FN=0
- `DOCUMENTATION`: TP=12, FN=0
- `MAINTAINABILITY`: TP=12, FN=0
- `RELIABILITY`: TP=12, FN=0
- `SECURITY`: TP=12, FN=0
- **`STRUCTURAL`**: **TP=0, FN=12** (100% false negative rate in this category)
- `TEST_QUALITY`: TP=12, FN=0

### Why `STRUCTURAL` Dropped to TP=0 in CI:
In [`go-backend/agent/repo_intel/benchmark/verification_harness.go`](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/verification_harness.go#L491-L508):
```go
rustCandidates := []string{
    "thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "../thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "../../thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "../../../thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "../../../../thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "E:/SWE/thermodynamic-ast-engine/target/release/thermodynamic-ast-engine.exe",
    "E:/SWE/thermodynamic-ast-engine/target/debug/thermodynamic-ast-engine.exe",
}
for _, cand := range rustCandidates {
    if _, err := os.Stat(cand); err == nil {
        absP, _ := filepath.Abs(cand)
        structuralAnalyzer := repo_intel.NewStructuralAnalyzer(absP, suitePath)
        _, strFindings, _ := structuralAnalyzer.AnalyzeStructural(context.Background(), 30*time.Second)
        allFindings = append(allFindings, strFindings...)
        break
    }
}
```

1. **Hardcoded Windows Binary Names:** Every candidate in `rustCandidates` ended in `.exe`. On Linux, Cargo builds the binary as `thermodynamic-ast-engine` (without `.exe`).
2. **Unread CI Environment Variable:** In `.github/workflows/ci-go.yml` line 182, the CI workflow explicitly exported:
   `export ENGINE_PATH="$GITHUB_WORKSPACE/thermodynamic-ast-engine/target/release/thermodynamic-ast-engine"`
   However, `RunMilestone32Evaluation` (and `RunBenchmarkEvaluation` and `stage_runner.go`) **never inspected `os.Getenv("ENGINE_PATH")`**.
3. **Silent Execution Bypass:** When no `.exe` file was found, `structuralAnalyzer.AnalyzeStructural` was skipped completely without generating a fatal error.
4. **Missing Structural Findings:** The 12 AST-level structural fixtures (`FIX-V3-STR-NEST-01` through `06` and `FIX-V3-STR-COMPLEX-07` through `12`) depend on Tree-sitter cognitive nesting and cyclomatic complexity metrics emitted by the Rust engine. Without them, all 12 fixtures produced zero findings, resulting in exactly 12 False Negatives and driving recall from 1.000 down to 0.875.

---

## 4. Summary of Required Remediations

1. **Security Fix (`workspace.go`):** Replace naive `strings.HasPrefix` with component-aware canonical containment checks (`filepath.Rel`), reject Windows drive/UNC/device formats cross-platform, enforce symlink evaluation, and guard against sibling and root directory deletion.
2. **Benchmark Precision Fix (`verification_harness.go`, `stage_runner.go`):** Implement `ResolveRustEngineBinary()` to honor `ENGINE_PATH` first, search both `thermodynamic-ast-engine` and `thermodynamic-ast-engine.exe` in `target/release` and `target/debug`, and ensure the structural analyzer executes consistently across both Linux and Windows runners.
