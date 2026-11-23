# M3.8 Phase 2.1.2 — Offline Dependency Validation & Threat Model Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Review Baseline Commit:** `a0928fba56d81f90241460788322e89cdedc7334`  
**Audit Scope:** Repository dependency staging threat model, minimal cache isolation, provenance verification, and toolchain offline enforcement.

---

## 1. Threat Model & Security Boundaries

Dynamic verification of untrusted repository code must maintain strict network isolation (`--network none`) while verifying Go code that may declare third-party modules.

### 1.1 Identified Threat Vectors & Mitigations

| Threat Vector | Mechanism | Applied Mitigation |
|---|---|---|
| **Host Toolchain Compromise** | Untrusted `go.mod` directs host `go mod download` to exploit Go compiler/vcs vulnerabilities or fetch malicious code onto the host | Direct host `go mod download` from untrusted repository metadata is **strictly prohibited**. |
| **Developer Cache Poisoning / Tampering** | Scanned repository modifies or corrupts shared global module cache files | Scanned repository runs inside isolated container; module cache is mounted strictly **read-only** (`:ro`). |
| **Host Module Cache Exposure** | Mounting full global module cache (`~/go/pkg/mod`) exposes proprietary packages, internal tokens, and unrelated module code | Only a **minimal disposable directory** containing the exact verified dependencies required by the workspace is staged and mounted. |
| **Toolchain Network Egress** | Go 1.21+ toolchain detects newer Go version in `go.mod` and attempts automatic network toolchain download | Container execution enforces `-e GOTOOLCHAIN=local` alongside `-e GOPROXY=off`, `-e GOSUMDB=off`, and `--network none`. |
| **Path Traversal via Replace Directives** | `replace` directive points to host files (e.g. `replace evil => ../../sensitive`) | `parseGoModRequirements()` parses and validates all `replace` targets, rejecting any targeting host or parent paths. |
| **Unpinned Dependency Tampering** | Workspace requests external modules without cryptographic checksums | `parseGoSumChecksums()` enforces non-empty pinned `go.sum` with SHA-256 (`h1:`) hashes for all requirements. |

---

## 2. Minimal Dependency Staging Architecture

### 2.1 Staging Flow (`prepareMinimalDependencyCache`)
1. **Requirements Parsing:** `parseGoModRequirements(wsDir)` extracts all `require` directives and verifies no unsafe `replace` directives exist. If zero external dependencies are required, staging completes immediately with no cache directory needed.
2. **Checksum Verification:** `parseGoSumChecksums(wsDir)` indexes `go.sum`. If any required module lacks a pinned checksum, execution aborts with `DYNAMIC_VERIFICATION_UNAVAILABLE`.
3. **Approved Cache Resolution:** Identifies pre-approved immutable cache sources (`APPROVED_DEPENDENCY_CACHE`, `STAGED_MODULE_CACHE`, `CI_DEPENDENCY_CACHE`).
4. **Disposable Cache Creation:** Allocates an ephemeral temporary directory (`os.MkdirTemp("", "ucs503-minmod-*")`).
5. **Selective Module Copy:** Copies **only** the required module version directory trees (`<modPath>@<version>`) into the disposable cache directory. Unrequested packages remain completely unexposed.
6. **Isolated Mount & Cleanup:** Passes `-v <disposableDir>:/go/pkg/mod:ro` to container execution. Deferred cleanup deletes the disposable directory upon completion.

---

## 3. Unit Test Verification Evidence

`TestDependencyStaging_ProhibitsArbitraryHostDownloadsAndEnforcesMinimalCache` evaluated all five threat vectors:

1. `MissingGoSumFailsClosed`: Workspace with unpinned module dependency failed closed with `DYNAMIC_VERIFICATION_UNAVAILABLE: repository declares external dependencies but lacks pinned go.sum checksums`.
2. `UnsafeReplaceDirectiveRejected`: Workspace with `replace ... => ../../sensitive` failed closed with `security violation: unsafe replace directive ... targeting host or escaped path`.
3. `MissingModuleInApprovedCacheFailsClosedWithoutDownload`: Uncached module failed closed with `DYNAMIC_VERIFICATION_UNAVAILABLE: required dependency ... is missing from approved offline cache; host network download is prohibited`. Zero host downloads attempted.
4. `StandardLibraryWorkspaceRequiresZeroCacheMounts`: Standard library workspaces execute with empty cache mount.
5. `MinimalCacheStagesOnlyRequiredModuleAndCleansUp`: Verified that:
   - Required module was staged into the disposable directory.
   - Unrequested secret/internal modules in the approved cache were **never** copied into the disposable directory.
   - Disposable directory was completely deleted after `cleanup()`.
