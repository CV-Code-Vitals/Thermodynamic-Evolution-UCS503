# Milestone 3.7 — Phase 0: Baseline Freeze & Cryptographic Manifest

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `bfcca4f358a89c7554a66709687584ed843391f1`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Freeze Date:** 2026-11-07  
**Authorization:** Read-Only Independent External Evaluation  
**Detector Tuning Against Evaluation Corpus:** **STRICTLY PROHIBITED**  
**Autonomous External Patch Execution:** **STRICTLY PROHIBITED**  

---

## 1. Executive Summary

Milestone 3.7 executes an independent external evaluation and repair-readiness audit of the complete Milestone 3.6 semantic diagnostic and remediation planning system. 

Before introducing or evaluating any third-party Go code, this document formally establishes a **cryptographic baseline freeze** of the diagnostic engine, planning engine, eligibility policy, workspace executor, and independent validator.

Any modification to detector logic, heuristics, thresholds, or eligibility policies during Milestone 3.7 is strictly prohibited to guarantee scientific validity, reproducibility, and prevent benchmark contamination.

---

## 2. Frozen Subsystem Component Manifest

| Component | Package Path | Source File | SHA-256 Digest |
| :--- | :--- | :--- | :--- |
| **Evidence Contract** | `agent/repo_intel/diagnostic/semantic` | `contract.go` | `6ce928e2c609b0b6fc19770f2660348490a07eec5f50ef759d57a9ef4e0a7a0b` |
| **Finding Reconciler** | `agent/repo_intel/diagnostic/semantic` | `reconciliation.go` | `3db18498257e33e136ba9269c84156822a7f516df7b293699b66a50616b359f4` |
| **Resource Engine** | `agent/repo_intel/diagnostic/semantic` | `resource_analysis.go` | `83be1d99980c625e02a1cf3c40c484ee7dbf964402636f406ecb6d6168fe2811` |
| **Resource Model** | `agent/repo_intel/diagnostic/semantic` | `resource_model.go` | `a7f01e6fc6b8caf1b18fcc7baf1aab52378941725ceea241b77bfb4f9104ce83` |
| **Unified Orchestrator** | `agent/repo_intel/diagnostic/semantic` | `unified_orchestrator.go` | `81f567f3ff359ae4e7df82089673b00c33a92576ca6c4b2671ae23388a101fd1` |
| **Opportunity Policy** | `agent/repo_intel/planning` | `opportunity.go` | `3b1e76375707b6cc7cea08a50fea59e6a8dae76c1ad5fe2292f3473917805179` |
| **Strategy Selector** | `agent/repo_intel/planning` | `strategy.go` | `7e7fed1748335daedbc0669b4e4f7995371ad52771b695aeb482a0d922fa6530` |
| **Criteria Generator** | `agent/repo_intel/planning` | `criteria.go` | `2f4e321a1acc62ed07ca0e74ebc58b4c8430fe9ffcb27a20c567a9cf96b77227` |
| **Remediation Planner** | `agent/repo_intel/planning` | `planner.go` | `9deea5457f6f3350ba48ce098989977fbc266f571b05ff9ffdf98df5b3e2300b` |
| **Eligibility Policy** | `agent/repo_intel/execution` | `eligibility.go` | `aa01d2fcee5f7c1c35dfe437b8660245906efd01ee3440e53a23330689b0d1e3` |
| **Independent Validator** | `agent/repo_intel/execution` | `validator.go` | `71b3ea22a34f0f574edcf1bed641a2f2eb696426429f947ad5ae64b1f47ea5e7` |
| **Controlled Test Suite** | `agent/repo_intel/execution` | `controlled_remediation_test.go` | `32a36c0aeff8878bd4b28d1237c1bbe45f7783ee8964d50e828131339393ebca` |

---

## 3. Verification of Milestone 3.6 Prior Artifacts

1. **Seven Controlled Fixture Repair Records (`milestone3_6_phase1_fixture_repair_ledger.json`):**
   - Verified that all 7 fixture repair trials (`FIX-CASE-01` through `07`) passed independent multi-stage validation in ephemeral workspaces.
   - Initial and final bitwise directory digests of fixture repositories verified identical ($\mathcal{H}_{\text{initial}} \equiv \mathcal{H}_{\text{final}}$).
2. **Five Adversarial Safety Scenarios (`controlled_remediation_test.go`):**
   - Stale finding hash mismatch (`ADV-01`): Verified fail-closed.
   - Malicious path traversal (`ADV-02`): Verified rejected by report ingester.
   - Ambiguous concurrency escape (`ADV-03`): Verified fail-closed.
   - Compiler syntax error (`ADV-04`): Verified automatic rollback.
   - Conflicting patches (`ADV-05`): Verified hard conflict detection.

---

## 4. Evaluation Constraints & Firebreak Rules

1. **Bitwise Source Immutability:** External test repositories must remain read-only. Evaluation harness will execute diagnostics and dry-run planning without writing patches to the external repositories.
2. **First-Look Integrity:** All raw diagnostic predictions, confidence scores, and planning determinations must be collected and frozen before examining match status against ground truth.
3. **No Retuning:** If false positives or false negatives are observed on the external corpus, the detector code must NOT be modified to fit the data. Observed failures must be reported as scientific findings.
