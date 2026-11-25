# M3.8 Phase 2.1.2 — Full Regression Matrix Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Review Baseline Commit:** `a0928fba56d81f90241460788322e89cdedc7334`  
**Working Tree State:** Local security fixes applied, working tree clean of untracked files.

---

## 1. Test Execution Summary

| Test Domain | Target Suite | Status | Execution Time | Details |
|---|---|---|---|---|
| **Go Backend Adversarial Security & Cleanup** | `agent/repo_intel/execution` (`TestAdversarial*`, `TestCommandPolicy*`, `TestContainerCleanup*`, `TestDependencyStaging*`) | **PASS** | 1.14s | 17/17 tests passed; zero forbidden execution attempts verified; cleanup and staging tested |
| **Go Backend Remediation & Oracles** | `agent/repo_intel/execution` (`TestControlled*`, `TestScanToRepair*`, `TestHardenedOracles*`, `TestAutonomous*`) | **PASS** | 35.24s | All controlled trials and oracle invariants passed |
| **Go Backend Authorization** | `.` (`TestSnapshotAuthorization*`) | **PASS** | 2.46s | Stale workspace rejection verified with opaque ID resolver |
| **Go Backend 27-Target Revalidation** | `agent/repo_intel/benchmark` (`TestMilestone3_8_Phase1_1_AutonomousRepairRevalidation`) | **PASS** | 1.06s | 27/27 targets evaluated: 9 verified, 6 abstained, 1 disqualified/abstained (NAT-03), 11 clean/ineligible |
| **Go Backend Pilot Benchmark** | `agent/repo_intel/benchmark` (`TestMilestone3_7_Phase2_ControlledExternalAutonomousRepairPilot`) | **PASS** | 1.36s | 27/27 pilot targets evaluated identically |
| **Go Toolchain Hygiene** | `go vet` across all non-upload packages | **PASS** | 8.42s | 0 warnings, 0 errors |
| **Rust AST Engine** | `thermodynamic-ast-engine` (`cargo test`) | **PASS** | 0.41s | 12/12 unit tests passed |
| **React Frontend Unit Tests** | `admin-portal` (`npm test -- --run`) | **PASS** | 6.49s | 49/49 unit tests passed |
| **React Frontend Linter** | `admin-portal` (`npm run lint` / `oxlint`) | **PASS** | 0.06s | 0 errors, 0 warnings on 23 files |
| **React Frontend Production Build** | `admin-portal` (`npm run build` / `vite build`) | **PASS** | 0.49s | Bundle built cleanly with `VITE_API_BASE` |

---

## 2. Invariant Verification

- **Zero Host Toolchain Execution on Untrusted Code:** Confirmed. AST parsing alone runs on host; all subprocess toolchain execution delegates to the container runner.
- **Fail-Closed Container Unavailability:** Confirmed. When container isolation is unavailable, dynamic checks fail closed with `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`).
- **Command Policy Boundary:** Confirmed. Rejects non-whitelisted binaries and dangerous flags before container creation (0 execution attempts).
- **Offline Isolation:** Confirmed. `--network none`, `-e GOPROXY=off`, `-e GOSUMDB=off`, `-e GONOSUMDB=*`, `-e GOTOOLCHAIN=local`, read-only module cache mount, no network egress.
- **Minimal Cache Mounting:** Confirmed. Only the minimal verified dependency set declared by `go.mod` is copied into disposable staging; developer global module cache is never exposed.
- **Synchronous Cleanup Lifecycle:** Confirmed. Container removal and existence verification via `docker inspect` complete before execution evidence is constructed.
- **Original Source Immutability:** Confirmed. Bitwise pre- and post-execution tree digests match on all test runs.
- **Single-Use Replay Protection:** Confirmed. Consumed tokens cannot be replayed across tasks or workspaces.
