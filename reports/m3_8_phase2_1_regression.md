# Milestone 3.8 Phase 2.1 — Comprehensive Regression & Benchmark Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Status:** COMPLETE & VERIFIED

---

## 1. Executive Summary

Milestone 3.8 Phase 2.1 introduces hardened container isolation, honest oracle outcomes, and cryptographic snapshot binding without degrading system performance or destabilizing existing suites.

All test suites across the Go backend, Rust AST engine, and React frontend were executed and verified clean. Historical ground-truth benchmarks remain strictly frozen at 9/15 injected repairs, 0 natural repairs, and NAT-03 permanently disqualified.

---

## 2. Test Execution Results by Subsystem

### A. Go Backend Execution & Remediations
- **Package:** `go-backend/agent/repo_intel/execution`
- **Tests Executed:**
  - `TestAdversarial_1_HostSentinelFileAccessPrevented`: PASSED
  - `TestAdversarial_2_ExternalNetworkEgressBlocked`: PASSED
  - `TestAdversarial_3_UnauthorizedSubprocessExecutionBlocked`: PASSED
  - `TestAdversarial_4_HostCredentialsAccessRejected`: PASSED
  - `TestAdversarial_5_ResourceQuotaExhaustionEnforced`: PASSED
  - `TestAdversarial_6_MalformedWorkspaceAndSnapshotIDsRejected`: PASSED
  - `TestAdversarial_7_ApprovalTokenReplayAcrossWorkspacesRejected`: PASSED
  - `TestAdversarial_8_PostApprovalSourceFileMutationRejected`: PASSED
  - `TestAdversarial_9_SymlinkEscapeDetectedAndRejected`: PASSED
  - `TestApprovalSecurityAdversarialVectors` (Vectors 1–10): PASSED
  - `TestControlledFixtureRepairTrials` (Cases 1–7): PASSED
  - `TestAdversarialSafetyScenarios` (Scenarios 1–5): PASSED
  - `TestHardenedOraclesRejectFlawedNAT03Patch`: PASSED
  - `TestHardenedOraclesRejectPreErrorCheckCleanup`: PASSED
  - `TestHardenedOraclesRejectMultipleCloseCalls`: PASSED
  - `TestMilestone31SafetyInvariants`: PASSED
  - `TestScanToRepairWorkflow_Success`: PASSED
  - `TestScanToRepairWorkflow_UnsupportedCommandInjection`: PASSED
  - `TestScanToRepairWorkflow_EscapingResourceAbstention`: PASSED
  - `TestScanToRepairWorkflow_VerificationFailure_FailClosed`: PASSED
  - `TestScanToRepairWorkflow_MissingWorkspace_FailClosed`: PASSED
  - `TestAutonomousPatchSynthesisAndIndependentOracles`: PASSED
  - `TestAutonomousEndToEndRemediationPipeline`: PASSED
- **Package Status:** **ok (100% tests passed)**

### B. Rust Thermodynamic AST Engine
- **Path:** `thermodynamic-ast-engine/Cargo.toml`
- **Command:** `cargo test --manifest-path thermodynamic-ast-engine/Cargo.toml`
- **Tests Executed:** 12 tests
  - `test_energy_formula_basic`: ok
  - `test_energy_zero_nodes`: ok
  - `test_go_cyclomatic_complexity`: ok
  - `test_go_function_parsing`: ok
  - `test_language_detection`: ok
  - `test_python_function_parsing`: ok
  - `test_scan_test_sample_python`: ok
  - `test_scan_test_sample_go`: ok
  - `test_is_binary_file_detection`: ok
  - `unknown_text_is_included_without_invented_function_metrics`: ok
  - `multilingual_function_analysis`: ok
  - `resolves_local_imports_and_unambiguous_calls_without_inventing_package_links`: ok
- **Subsystem Status:** **ok (12 passed; 0 failed)**

### C. Admin Portal (Frontend & Client SDK)
- **Path:** `admin-portal`
- **Command:** `npm test -- --run`
- **Tests Executed:** 49 tests (including 9 Autonomous Remediation Client API tests)
  - `fetchRemediationTargets`: ok
  - `fetchRemediationTargets (with workspace_id)`: ok
  - `analyzeRemediationTarget`: ok
  - `proposeRemediationRepair`: ok
  - `approveRemediationRepair`: ok
  - `executeRemediationRepair`: ok
  - `rollbackRemediationSandbox`: ok
  - `fetchRemediationAuditReport`: ok
  - `runRemediationDemo`: ok
  - All archive, evolution, and graph layout tests: ok
- **Subsystem Status:** **ok (49 passed; 0 failed)**

---

## 3. Ground Truth Benchmark Integrity Preservation

The benchmark ledger and reporting invariants remain strictly preserved:

1. **Injected Defect Repair Rate:**
   - 9 out of 15 candidate defects repaired and verified under controlled evaluation.
   - 6 defects safely abstained due to high cyclomatic complexity, escaping references, or architectural risk.
2. **Natural Defect Repair Rate:**
   - Exactly 0 confirmed natural defect repairs.
   - `NAT-03` (`gorilla/websocket`) is permanently marked `GROUND_TRUTH_DISQUALIFIED`. As proven in Phase 1.2 forensic audit, `DialContext` does not require application body closure, and Issue #789 was an unrelated error code inquiry. The synthesizer correctly abstains from introducing callee cleanup on escaping socket objects.
3. **Demo vs. Live Finding Separation:**
   - Preconfigured demonstration fixtures (`DefaultPreconfiguredTargets()`) are explicitly tagged with `IsDemoTarget: true` and prefixed with `[DEMO]`.
   - Real findings discovered during live repository scans are tagged with `Stratum: "SCANNED_REPOSITORY"` and `IsDemoTarget: false`, preventing confusion between pilot benchmarks and production repository scans.
