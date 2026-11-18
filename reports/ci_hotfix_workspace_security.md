# CI Hotfix — Workspace Path Containment & Isolation Security Architecture

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Date:** 2026-11-18  
**Component:** `go-backend/agent/repo_intel/execution/workspace.go`  

---

## 1. Executive Summary

Prior to this hotfix, workspace path boundary containment in [`workspace.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/workspace.go) relied on string-prefix matching (`strings.HasPrefix(destPath, wsDir)`). This pattern suffered from two critical vulnerabilities:
1. **Sibling-Prefix Confusion:** A directory sharing a common prefix (e.g. `/tmp/ws_test-evil` vs `/tmp/ws_test`) was accepted as contained because `strings.HasPrefix("/tmp/ws_test-evil", "/tmp/ws_test")` evaluates to `true`.
2. **Cross-Platform Drive & UNC Bypass on POSIX:** On Linux runners, Windows drive-letter (`C:\...`) and UNC (`\\...`) paths are not recognized as absolute by `filepath.IsAbs`, nor are backslashes converted by `filepath.FromSlash`. Consequently, paths targeting host system files were concatenated as relative paths and silently accepted.
3. **Base Directory Deletion:** `DestroyWorkspace` permitted deletion of `wm.BaseDir` itself if `wsDir == wm.BaseDir`.
4. **Symlink Boundary Traversal:** Symlinks pointing outside the workspace were not evaluated before path acceptance.

The security architecture of `WorkspaceManager` was refactored with **component-aware relative path validation**, **cross-platform path format rejection**, **symlink resolution**, and **strict containment invariants**.

---

## 2. Comprehensive Attack Vector Defense Matrix

| Attack Vector | Example Input | Pre-Fix Behavior | Post-Fix Defense | Status |
| :--- | :--- | :--- | :--- | :---: |
| **Windows Drive Absolute (Backslash)** | `C:\Windows\System32` | Accepted on Linux as relative path | Rejected cross-platform by `ValidatePathSafety` (drive spec check) | **BLOCKED** |
| **Windows Drive Absolute (Slash)** | `C:/Windows/System32` | Accepted on Linux as relative path | Rejected cross-platform by `ValidatePathSafety` (drive spec check) | **BLOCKED** |
| **Windows Drive Relative** | `C:relative`, `c:foo/bar` | Accepted as relative path | Rejected cross-platform by `ValidatePathSafety` (drive spec check) | **BLOCKED** |
| **Windows UNC Share (Backslash)** | `\\server\share\file` | Accepted on Linux as relative path | Rejected cross-platform by `ValidatePathSafety` (UNC check) | **BLOCKED** |
| **Windows UNC Share (Slash)** | `//server/share/file` | Accepted on Linux as relative path | Rejected cross-platform by `ValidatePathSafety` (UNC check) | **BLOCKED** |
| **Windows Device Namespace** | `\\?\C:\Windows`, `\\.\COM1` | Accepted on Linux as relative path | Rejected cross-platform by `ValidatePathSafety` (Device path check) | **BLOCKED** |
| **POSIX Root Absolute** | `/etc/passwd` | Rejected on Linux, accepted on Windows | Rejected on all platforms (starts with `/` or `\` or `IsAbs`) | **BLOCKED** |
| **Root-Relative Backslash** | `\etc\passwd`, `\Windows` | Accepted on Linux | Rejected on all platforms (starts with `/` or `\`) | **BLOCKED** |
| **Relative Directory Traversal** | `../../outside`, `..\..\outside` | Partially caught | Rejected by segment inspection (`..`) and `filepath.Clean` | **BLOCKED** |
| **Mixed Path Separators** | `sub/..\..\outside`, `sub\../..\outside` | Inconsistent cleaning | Normalized to `/`, segments parsed, rejected | **BLOCKED** |
| **Workspace-Root Targeting** | `.` | Target resolved to workspace root | Rejected explicitly: target must be file inside workspace | **BLOCKED** |
| **Empty Path / Null Bytes** | `""`, `"a\x00b"` | Edge case crash | Rejected by length and null rune validation | **BLOCKED** |
| **Sibling-Prefix Confusion** | `wsDir-evil`, `../ws_test-evil` | `strings.HasPrefix` allowed | `filepath.Rel` component validation rejects `../ws_test-evil` | **BLOCKED** |
| **Base Directory Destruction** | `DestroyWorkspace(wm.BaseDir)` | `strings.HasPrefix` allowed | Explicit check `cleanWs == cleanBase` refuses deletion | **BLOCKED** |
| **Base Sibling Destruction** | `DestroyWorkspace(baseDir + "-evil")` | `strings.HasPrefix` allowed | `filepath.Rel` component check refuses deletion | **BLOCKED** |
| **Symlink Escape** | `symlink_dir/secret.txt` | Symlinks followed blindly | `filepath.EvalSymlinks` detects target outside `realWs` | **BLOCKED** |

---

## 3. Implementation Details

### A. Pre-Validation Sanitizer (`ValidatePathSafety`)
```go
func ValidatePathSafety(targetRelPath string) error {
    trimmed := strings.TrimSpace(targetRelPath)
    if trimmed == "" {
        return fmt.Errorf("security violation: path traverses outside workspace (empty path)")
    }
    if strings.ContainsRune(targetRelPath, 0) {
        return fmt.Errorf("security violation: path traverses outside workspace (null byte detected)")
    }
    if filepath.IsAbs(targetRelPath) {
        return fmt.Errorf("security violation: path traverses outside workspace (native absolute path): %s", targetRelPath)
    }
    if strings.HasPrefix(targetRelPath, "/") || strings.HasPrefix(targetRelPath, "\\") {
        return fmt.Errorf("security violation: path traverses outside workspace (root path): %s", targetRelPath)
    }
    if len(targetRelPath) >= 2 && targetRelPath[1] == ':' {
        c := targetRelPath[0]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
            return fmt.Errorf("security violation: path traverses outside workspace (windows drive specification): %s", targetRelPath)
        }
    }
    if strings.HasPrefix(targetRelPath, `\\`) || strings.HasPrefix(targetRelPath, `//`) ||
        strings.HasPrefix(targetRelPath, `\/`) || strings.HasPrefix(targetRelPath, `/\`) {
        return fmt.Errorf("security violation: path traverses outside workspace (UNC or device namespace): %s", targetRelPath)
    }
    norm := strings.ReplaceAll(targetRelPath, "\\", "/")
    segments := strings.Split(norm, "/")
    for _, seg := range segments {
        if seg == ".." {
            return fmt.Errorf("security violation: path traverses outside workspace (traversal segment '..'): %s", targetRelPath)
        }
    }
    cleanRel := filepath.Clean(filepath.FromSlash(norm))
    if cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanRel, "../") || strings.HasPrefix(cleanRel, `..\`) {
        return fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
    }
    if cleanRel == "." {
        return fmt.Errorf("security violation: path traverses outside workspace (targets workspace root itself): %s", targetRelPath)
    }
    return nil
}
```

### B. Canonical Relative-Path Semantics (`ValidateContainedPath`)
Rather than substring checks, the containment boundary is verified mathematically using `filepath.Rel`:
```go
rel, err := filepath.Rel(cleanWs, cleanFull)
if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
    return "", fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
}
```
If `cleanWs` is `/tmp/ws` and `cleanFull` is `/tmp/ws-evil`, `filepath.Rel` returns `../ws-evil`. Because it starts with `..`, it is immediately rejected, completely neutralizing sibling-prefix confusion.

### C. Symlink Escape & TOCTOU Protection
For existing paths, `filepath.EvalSymlinks(cleanFull)` is evaluated against `filepath.EvalSymlinks(cleanWs)`. If the target does not exist yet (e.g. creating a new patch file), the ancestor directory chain is walked upward to find the deepest existing ancestor, and its symlink resolution is confirmed inside the workspace.

### D. Safe Workspace Cleanup (`DestroyWorkspace`)
```go
if cleanWs == cleanBase {
    return fmt.Errorf("refusing to destroy base directory itself: %s", wsDir)
}
rel, err := filepath.Rel(cleanBase, cleanWs)
if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
    return fmt.Errorf("refusing to destroy directory outside base workspace: %s", wsDir)
}
```
The workspace manager guarantees that only dynamically allocated subdirectories within `wm.BaseDir` can ever be deleted.

---

## 4. Regression Test Suite

The test suite in [`go-backend/agent/repo_intel/execution/execution_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/execution_test.go#L145) was expanded to execute 21 distinct adversarial subtests covering all vectors described above. All 21 subtests passed with 0 errors.
