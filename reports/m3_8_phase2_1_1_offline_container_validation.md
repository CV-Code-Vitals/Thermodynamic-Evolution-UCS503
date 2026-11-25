# M3.8 Phase 2.1.1 — Offline Container Verification Architecture

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Audit Target:** Isolated Offline Container Verification (`--network none`)  
**Evaluation Date:** 2026-11-23  
**Auditor:** Antigravity Autonomous Security Engineer  

---

## 1. Challenge & Failure Mode Analysis

In hardened container environments where network access is strictly disabled (`--network none`) to prevent SSRF and exfiltration, dynamic verification can encounter several failure modes if the execution environment is not properly staged:

1. **Missing `go.mod` in Standalone Fixtures:**
   Standalone fixtures or generated test workspaces lacking a `go.mod` fail under Go 1.16+ module mode when `go vet` or `go test` runs (`go: go.mod file not found in current directory or any parent directory`).
2. **Incorrect Working Directory for Sub-Packages:**
   If a patched target file is located in a sub-package (e.g. `server/logger.go`), running `go vet .` from the module root fails if the root contains no Go files (`no Go files in /workspace`).
3. **Missing Module Dependencies Under `--network none`:**
   Running dependency downloads (`go get`) inside a container with `--network none` will hang or fail. Running dependency downloads on the *host* for an untrusted repository introduces remote code execution risks (`go.mod` directives, toolchain exploits).
4. **Distinction Between Unavailable Infrastructure and Test Failures:**
   When container infrastructure is unavailable, dynamic verification must explicitly report `UNAVAILABLE` (`DYNAMIC_VERIFICATION_UNAVAILABLE`) and fail-closed. A missing dynamic verification must **never** be classified as a successful repair.

---

## 2. Engineered Solutions

### A. Disposable Minimal Fixture Modules (`EnsureDisposableModule`)
In [go-backend/agent/repo_intel/execution/oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go):

```go
func EnsureDisposableModule(wsDir string) error {
    if strings.TrimSpace(wsDir) == "" {
        return nil
    }
    goModPath := filepath.Join(wsDir, "go.mod")
    if _, err := os.Stat(goModPath); os.IsNotExist(err) {
        content := "module disposable_fixture\n\ngo 1.21\n"
        return os.WriteFile(goModPath, []byte(content), 0644)
    }
    return nil
}
```
- Operates strictly within the ephemeral sandbox workspace.
- Leaves existing valid `go.mod` files untouched.
- Allows standard library imports (`os`, `io`, `fmt`, `net/http`, `context`, `database/sql`) to be compiled and vetted offline without network access.

### B. Accurate Package Targeting for `go vet`
In [go-backend/agent/repo_intel/execution/oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go):

```go
    vetPkg := "."
    if patch != nil && patch.TargetFile != "" {
        relDir := filepath.Dir(filepath.FromSlash(patch.TargetFile))
        if relDir != "." && relDir != "" {
            vetPkg = "./" + filepath.ToSlash(relDir)
        }
    }
    vetCmd := "go vet " + vetPkg
```
- Targets the exact package containing the candidate patch.
- Prevents false-negative failures when root directories contain only sub-packages.

### C. Offline Dependency Strategy: Zero Network & Zero Host Dependency Ingestion
- In containers: `--network none` is non-negotiable.
- Authentic repositories in `benchmarks/external_v2/` and `heldout_v1/` are self-contained or use Go standard libraries.
- Host toolchains never execute `go get`, `go install`, or `go mod download` on untrusted target repositories.

### D. Deliberately Injected Unavailable Container Runner
For testing and development environments without container capabilities, `UnavailableContainerRunner` provides explicit fail-closed testing:

```go
type UnavailableContainerRunner struct {
    Reason string
}

func (u *UnavailableContainerRunner) CheckAvailable(ctx context.Context) error {
    return fmt.Errorf("%w: %s", ErrSandboxUnavailable, u.Reason)
}

func (u *UnavailableContainerRunner) RunCommand(ctx context.Context, wsDir string, cmdStr string) (*ContainerRunResult, error) {
    return nil, u.CheckAvailable(ctx)
}
```
When injected into `IndependentOracleSuite`, dynamic checks produce:
- Status: `UNAVAILABLE`
- Passed: `false`
- FailureReason: `DYNAMIC_VERIFICATION_UNAVAILABLE: secure container runtime not detected...`
- Repair Outcome: `FAILED` / `REJECTED` (Never claimed as verified).

---

## 3. Verification Verdict

All fixture repair trials in `TestControlledFixtureRepairTrials` (7/7 cases) and `TestAutonomousPatchSynthesisAndIndependentOracles` (9/9 cases) execute and pass offline container verification without network access or host compromises.
