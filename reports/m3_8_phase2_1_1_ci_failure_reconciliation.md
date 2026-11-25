# M3.8 Phase 2.1.1 — CI Failure Baseline & Reconciliation Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Baseline Remote & Local HEAD:** `ef65814596a8f6a00e6205e4fd663a70808a523f`  
**Prior Integration Base:** `ab07861e9bf7764a5f371ccac212706ba219a0b1`  
**Audit Date:** 2026-11-23  
**Auditor:** Antigravity Autonomous Security Engineer  

---

## 1. Baseline Commit & Remote Synchronization Reconciliation

### A. Git Head & Branch Alignment
- **Local HEAD:** `ef65814596a8f6a00e6205e4fd663a70808a523f`
- **Remote `origin/release/milestone-3.8-mvp-rc` HEAD:** `ef65814596a8f6a00e6205e4fd663a70808a523f`
- **Synchronization Status:** Local working branch is fully synchronized with remote HEAD.

### B. Upstream Push Audit & Report Discrepancy Reconciliation
In the preceding CI Stabilization Hotfix session, local engineering changes were verified and 5 reports were written under local constraints stating zero remote pushes were performed by the agent. Subsequent to that turn, commit `ef65814` (*"feat: implement multi-oracle verification engine and containerized sandbox workspace for safe patch execution"*) was authored and published to GitHub by repository maintainer `Saksham Mishra <mishrasaksham080705@gmail.com>` on `Mon Nov 23 12:08:23 2026 +0530`.

This audit reconciles that discrepancy: the published remote commit `ef65814` incorporates the previous hotfix file changes, and this Phase 2.1.1 reconciliation establishes its starting point strictly from `ef65814`.

---

## 2. Forensic Inspection of Changes in `oracles.go` After `ab07861e`

Inspection of `git diff ab07861e..ef658145 go-backend/agent/repo_intel/execution/oracles.go` confirms:
```diff
-	// Ensure isolated temp directories
-	tmpBase := `E:\SWE\.tmp`
-	if fi, err := os.Stat(tmpBase); err == nil && fi.IsDir() {
+	// Ensure isolated temp directories via dynamic workspace resolution
+	tmpBase, err := EnsureTmpDir(ResolveWorkspaceRoot())
+	if err == nil && tmpBase != "" {
 		cleanEnv = append(cleanEnv, `GOTMPDIR=`+tmpBase, `TEMP=`+tmpBase, `TMP=`+tmpBase)
+	} else {
+		tmpBase = os.TempDir()
 	}
```
While this diff eliminated the hardcoded host temp directory `E:\SWE\.tmp` on Linux runners, it left five critical blockers unaddressed:

1. **Unenforced Command Policy in Container Runner:**
   `DockerContainerRunner.RunCommand()` directly tokenized `strings.Fields(cmdStr)` and appended it to `docker run` arguments without calling `validateSafeSubprocessCommand()`. Prohibited interpreters (`sh`, `bash`, `python`) and downloaders (`curl`, `wget`) were not blocked before container creation.
2. **Evaluation Order Defect in Oracles:**
   In `verifyRegressionTests`, `verifyDefectReproducer`, and `verifyBehavioralInvariance`, `s.Runner.CheckAvailable(ctx)` was evaluated *before* checking command validity. Malicious commands like `sh -c 'echo pwned'` failed closed with `DYNAMIC_VERIFICATION_UNAVAILABLE` instead of returning an explicit `Security rejection`.
3. **Workspace Registration Violation in Snapshot Tests:**
   `TestSnapshotAuthorization_StaleWorkspaceRejected` passed an absolute filesystem directory (`C:\Users\...\stale_ws`) directly into `DiscoverWorkspaceTargets()`. Because `ValidateWorkspaceID()` properly forbids absolute paths to prevent traversal and ambiguity, the test failed closed.
4. **Target Path Concatenation Bug in Scanners:**
   In [mvp_scanner.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/mvp_scanner.go), `bindTargetSnapshot` joined `wsDir` and `target.TargetFile` without checking if `TargetFile` was already absolute, producing malformed paths on disk (`wsDir/wsDir/...`) and causing `TestEndToEndScanToRepairWorkflow_LiveRepo` to discover 0 scanned targets.
5. **CRLF Line-Ending Mismatch in Historical Fixture:**
   In [remediation_trial_test.go](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/remediation_trial_test.go), string comparison in trial `REM-CAND-01` against `fs.go` failed on Windows due to `\r\n` line endings.

---

## 3. Preserved Failure Logs vs. Reconciled Passing Status

### Failing Log Baseline (Commit `ef658145`)
```text
--- FAIL: TestUntrustedExecution_ArbitraryCommandBlocked (0.18s)
    scan_to_repair_e2e_test.go:547: Expected Security rejection for "sh -c 'echo pwned'", got: DYNAMIC_VERIFICATION_UNAVAILABLE: secure container runtime not detected or daemon not running; host execution forbidden
--- FAIL: TestSnapshotAuthorization_StaleWorkspaceRejected (0.01s)
    scan_to_repair_e2e_test.go:581: DiscoverWorkspaceTargets failed: SECURITY_VIOLATION: MALFORMED_WORKSPACE_ID: absolute workspace ID forbidden: C:\Users\mishr\AppData\Local\Temp\...\stale_ws
--- FAIL: TestEndToEndScanToRepairWorkflow_LiveRepo (26.66s)
    scan_to_repair_e2e_test.go:102: Expected scanned repository targets for CV-Code-Vitals/Thermodynamic-Evolution-UCS503, found 0
--- FAIL: TestExecuteExternalRemediationTrials (3.80s)
    remediation_trial_test.go:202: target text not found in fs.go
```

### Reconciled Passing Execution (Phase 2.1.1 Corrected)
```text
=== RUN   TestUntrustedExecution_CredentialStripping
--- PASS: TestUntrustedExecution_CredentialStripping (0.00s)
=== RUN   TestUntrustedExecution_ArbitraryCommandBlocked
--- PASS: TestUntrustedExecution_ArbitraryCommandBlocked (0.01s)
=== RUN   TestSnapshotAuthorization_StaleWorkspaceRejected
--- PASS: TestSnapshotAuthorization_StaleWorkspaceRejected (0.01s)
=== RUN   TestEndToEndScanToRepairWorkflow_LiveRepo
    scan_to_repair_e2e_test.go:63: Scan completed: WorkspaceID=edc9758b-73b7-4cc5-9a2d-f33e4f95ab6c, Files=673, Hotspots=22763
    scan_to_repair_e2e_test.go:100: Discovered 1293 scanned targets and 5 demo targets
    scan_to_repair_e2e_test.go:152: Successfully synthesized patch diff for SCAN-SEMANTIC_SQL_INJECTION-2 (length 2582 chars)
    scan_to_repair_e2e_test.go:201: Execution completed: Verdict=SYNTAX_ONLY_VERIFIED, SandboxApplied=true, SourceImmutable=true
--- PASS: TestEndToEndScanToRepairWorkflow_LiveRepo (33.33s)
=== RUN   TestExecuteExternalRemediationTrials
--- PASS: TestExecuteExternalRemediationTrials (66.48s)
=== RUN   TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands
--- PASS: TestCommandPolicy_ZeroExecutionAttemptsOnForbiddenCommands (0.02s)
```

All baseline failures have been systematically reproduced, debugged, and resolved.
