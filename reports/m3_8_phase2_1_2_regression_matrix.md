# M3.8 Phase 2.1.2 — Full Regression Matrix Report

**Date:** 2026-11-23  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Commit:** `0843ca613c602764ef3f534fc21bed7843486ba0` (with local Phase 2.1.2 stabilization fixes)

---

## 1. Test Execution Summary

| Test Domain | Target Suite | Status | Execution Time | Details |
|---|---|---|---|---|
| **Go Backend Adversarial Security** | `agent/repo_intel/execution` (`TestAdversarial*`, `TestCommandPolicy*`) | **PASS** | 1.41s | 13/13 tests passed; zero forbidden execution attempts verified |
| **Go Backend Remediation & Oracles** | `agent/repo_intel/execution` (`TestControlled*`, `TestScanToRepair*`, `TestHardenedOracles*`) | **PASS** | 6.08s | All controlled trials and oracle invariants passed |
| **Go Backend Root E2E** | `scan_to_repair_e2e_test.go` (`TestEndToEndScanToRepairWorkflow_LiveRepo`) | **PASS** | 15.74s | Live repo scan, 1,293 targets discovered, SQL injection patch verified, immutability preserved |
| **Go Backend Authorization** | `scan_to_repair_e2e_test.go` (`TestSnapshotAuthorization*`) | **PASS** | 1.55s | Stale workspace rejection verified with opaque ID resolver |
| **Go Backend External Trials** | `agent/repo_intel/benchmark` (`TestExecuteExternalRemediationTrials`) | **PASS** | 40.61s | 9/9 candidate trials verified resolved |
| **Go Backend 27-Target Revalidation** | `agent/repo_intel/benchmark` (`TestMilestone3_8_Phase1_1_AutonomousRepairRevalidation`) | **PASS** | 0.81s | 27/27 targets evaluated: 9 verified, 6 abstained, 1 disqualified/abstained, 11 clean/ineligible |
| **Go Backend Pilot Benchmark** | `agent/repo_intel/benchmark` (`TestMilestone3_7_Phase2_ControlledExternalAutonomousRepairPilot`) | **PASS** | 0.75s | 27/27 pilot targets evaluated identically |
| **Go Toolchain Hygiene** | `go vet` across all non-upload packages | **PASS** | 9.80s | 0 warnings, 0 errors |
| **Rust AST Engine** | `thermodynamic-ast-engine` (`cargo test`) | **PASS** | 0.01s | 12/12 unit tests passed |
| **React Frontend Unit Tests** | `admin-portal` (`node --test tests/*.test.js`) | **PASS** | 4.60s | 49/49 unit tests passed |
| **React Frontend Linter** | `admin-portal` (`npm run lint` / `oxlint`) | **PASS** | 0.07s | 0 errors, 0 warnings on 23 files |
| **React Frontend Production Build** | `admin-portal` (`npm run build` / `vite build`) | **PASS** | 0.58s | Bundle built cleanly with `VITE_API_BASE` |

---

## 2. Invariant Verification

- **Zero Host Toolchain Execution on Untrusted Code:** Confirmed. AST parsing alone runs on host; all subprocess toolchain execution delegates to the container runner.
- **Fail-Closed Container Unavailability:** Confirmed. When container isolation is unavailable, dynamic checks fail closed with `OracleStatusUnavailable`.
- **Command Policy Boundary:** Confirmed. Rejects non-whitelisted binaries and dangerous flags before container creation.
- **Offline Isolation:** Confirmed. `--network none`, `GOPROXY=off`, read-only module cache mount, no network egress.
- **Original Source Immutability:** Confirmed. Bitwise pre- and post-execution tree digests match on all test runs.
- **Single-Use Replay Protection:** Confirmed. Consumed tokens cannot be replayed.
