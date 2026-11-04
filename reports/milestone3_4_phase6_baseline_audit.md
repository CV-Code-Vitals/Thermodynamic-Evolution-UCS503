# MILESTONE 3.4 — PHASE 6 BASELINE AUDIT REPORT
## Immutable Candidate Freeze, Benchmark Manifest Verification & Held-Out Provenance Deconstruction

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Validation Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Frozen Candidate Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Audit Timestamp:** November 2026  
**Execution Classification:** Independent Validation / Restricted Remediation Pilot  
**Audit Gate:** Gate 0 — Baseline Freeze & Provenance Audit  

---

## 1. Commit, Branch, and Environment Verification

- **Candidate Baseline Commit SHA:** `a64e7632476c1341343481661a3e246c69e5a644`
- **Validation Branch:** `validation/milestone-3.4-phase6-independent-pilot` (branched cleanly from `research/milestone-3.4-semantic-analysis` at `a64e763`)
- **Working Tree State:** Completely clean (`git status` reports 0 untracked or modified files)
- **Host Platform:** Windows 11 Enterprise (amd64 / x86_64)
- **Toolchain Specifications:**
  - Go: `go1.25.6 windows/amd64`
  - Rust: `rustc 1.97.1 (8bab26f4f 2026-07-14)` / `cargo 1.97.1`

---

## 2. Cryptographic Manifest Integrity Audit

All historical and existing evaluation manifests were cryptographically verified using SHA-256 digests:

| Benchmark Asset | Path | SHA-256 Digest | Status | Composition |
|---|---|---|---|---|
| **Suite v2 Manifest** | `benchmarks/ground_truth_manifest_v2.json` | `79fcea5048600c9a5f80583fb5d8f66aa42cf79f45cb5b2f243359e40feaec04` | **LOCKED & VERIFIED** | 64 defects, 32 controls (96 fixtures) |
| **Suite v3 Manifest** | `benchmarks/ground_truth_manifest_v3.json` | `c1ccbd657eb090d6daf948fccedc47428f31eb2ffaa8b275054823f684b56689` | **LOCKED & VERIFIED** | 96 defects, 48 controls (144 fixtures) |
| **External v1 Manifest** | `benchmarks/external_v1/repository_manifest.json` | `6d1568b7fada7a1929ba49f82a3ce813f94b51dc40b24977ff0a28cf6189c592` | **LOCKED & VERIFIED** | 14 defects, 15 controls (5 repos) |
| **Held-Out v1 Manifest** | `benchmarks/heldout_v1/manifest.json` | `5b9677d97f75a666a87fbd243db32811a93c0641f65a352a14b3536d721e5f53` | **LOCKED & VERIFIED** | 9 defects, 12 controls (3 repos) |

---

## 3. Regression Suite Re-Execution Ledger

The complete Phase 5 diagnostic and execution regression suites were re-executed against the frozen baseline:

| Test Target / Package | Command | Exit Code | Duration | Outcome |
|---|---|---|---|---|
| **Semantic Diagnostic Suite** | `cd go-backend && go test -count=1 ./agent/repo_intel/diagnostic/semantic/...` | 0 | 0.569s | **PASS (All tests pass)** |
| **Thermodynamic AST Engine** | `cd thermodynamic-ast-engine && cargo test --offline` | 0 | 0.390s | **PASS (12 passed; 0 failed)** |
| **Diagnostic & Context Engines** | `cd go-backend && go test -count=1 ./agent/repo_intel/diagnostic/... ./agent/repo_intel/execution/...` | 0 | 2.590s | **PASS (All packages pass)** |

Zero code changes or regressions occurred between Phase 5 candidate freeze and Phase 6 validation baseline.

---

## 4. Historical Confusion Matrix Reconstruction

Historical confusion matrices reconstructed from raw prediction outputs and frozen manifests:

| Evaluation Suite | Ground Truth Defects | Negative Controls | TP | FP | FN | TN | Precision | Recall | Specificity | F1 Score |
|---|---|---|---|---|---|---|---|---|---|---|
| **Suite v2** | 64 | 32 | 64 | 0 | 0 | 32 | 100.0% | 100.0% | 100.0% | 100.0% |
| **Suite v3** | 96 | 48 | 96 | 0 | 0 | 48 | 100.0% | 100.0% | 100.0% | 100.0% |
| **Phase 2 (SQL/Cmd)** | 10 | 10 | 10 | 0 | 0 | 10 | 100.0% | 100.0% | 100.0% | 100.0% |
| **Phase 3 (Resource)** | 12 | 12 | 12 | 0 | 0 | 12 | 100.0% | 100.0% | 100.0% | 100.0% |
| **Phase 4 (Error)** | 10 | 13 | 10 | 0 | 0 | 13 | 100.0% | 100.0% | 100.0% | 100.0% |
| **External v1 (Targeted)** | 9 | 15 | 9 | 0 | 0 | 15 | 100.0% | 100.0% | 100.0% | 100.0% |
| **External v1 (Full Denominator)** | 14 | 15 | 9 | 0 | 5 | 15 | 100.0% | **64.29%** | 100.0% | **78.26%** |
| **Held-Out v1 (Internal)** | 9 | 12 | 9 | 0 | 0 | 12 | 100.0% | **100.0%** | 100.0% | **100.0%** |

### Verification of the 14-Defect External v1 Denominator
The full external denominator is confirmed at exactly **14 defects**:
- **Recovered (9):**
  - `DEF-HTTP-01` (Resource leak: `resp.Body.Close()`)
  - `DEF-HTTP-02` (Unhandled error: `json.NewEncoder().Encode()`)
  - `DEF-SQL-01` (SQL injection: `fmt.Sprintf` query)
  - `DEF-SQL-02` (Resource leak: `sql.Rows` unclosed)
  - `DEF-SQL-03` (Unhandled error: `tx.Rollback()` error ignored)
  - `DEF-CLI-01` (Command injection: `exec.Command("sh", "-c", ...)`)
  - `DEF-CLI-03` (Unhandled error: `cmd.Run()` unchecked)
  - `DEF-QUEUE-01` (Resource leak: `net.Dial` connection unclosed)
  - `DEF-QUEUE-03` (Unhandled error: `os.WriteFile` unchecked)
- **Unrecovered Historical Defects (5):**
  1. `DEF-CACHE-01` (`CONCURRENCY_RACE`: Concurrent map access without RLock)
  2. `DEF-CACHE-02` (`GOROUTINE_LEAK`: Unbounded goroutine spawn)
  3. `DEF-HTTP-03` (`CONTEXT_LEAK`: `context.WithTimeout` cancel ignored)
  4. `DEF-CLI-02` (`PATH_TRAVERSAL`: `filepath.Join` untrusted traversal)
  5. `DEF-QUEUE-02` (`TIMER_LEAK`: `time.After` in select loop)

These 5 defects remain explicitly classified as **UNRESOLVED / OUT-OF-SCOPE** for Phase 5 and are not falsely claimed as recovered.

---

## 5. Special Audit: Provenance of Phase 5 Held-Out Benchmark (`benchmarks/heldout_v1`)

A forensic provenance audit of `benchmarks/heldout_v1/manifest.json` was conducted to determine its scientific independence.

### Forensic Findings
1. **Author Provenance:** The 3 repositories (`microservice_api`, `stream_ingestor`, `auth_broker`) were authored internally by the implementing agent during Phase 5 execution.
2. **Ground Truth Timing:** Ground truth labels and code fixtures were authored concurrently within the same development cycle.
3. **Analyzer Adjustments During Testing:**
   - During the first execution of `TestEvaluateHeldOutSuiteV1`, the CFG engine failed to detect the resource leak in `ExportReportLeak` because the void function fell through to the exit block without an explicit `return`. The analyzer source code ([`cfg.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/cfg.go#L78-L82)) was subsequently modified (`cfg.ReturnBlocks = append(cfg.ReturnBlocks, lastBlock)`) to resolve this issue.
   - Fixture source files (`auth_broker/token/jwt.go` and `microservice_api/api/handler.go`) were modified during test execution to adjust error handling on secondary operations (`mac.Write` and `io.ReadAll`).
4. **Pattern Similarity:** The defect scenarios were designed to mirror the known semantic rules developed in Phases 2–4.

### Scientific Provenance Classification
Based on the above findings, `benchmarks/heldout_v1` is formally reclassified from an "Independent External Benchmark" to:

$$\mathbf{DEVELOPMENT\_VALIDATION\ /\ INTERNALLY\_CONSTRUCTED\_HELDOUT}$$

> **Scientific Directive Compliance:** The term **independent** is hereby revoked for `benchmarks/heldout_v1`. Its 100% detection metrics are strictly documented as **internal development validation**. A genuine, non-synthetic, third-party external evaluation corpus must be established in Phase 1 before claims of external generalization can be validated.

---

## 6. Deviations from Phase 5 Completion Report

1. **Benchmark Classification Correction:** The Phase 5 Completion Report referred to `heldout_v1` as an "independent held-out benchmark." This audit supersedes that classification: `heldout_v1` was internally constructed and subject to post-observation analyzer adjustment.
2. **Full-Suite Recall vs Headline Claims:** Full external suite recall on the 14-defect denominator is 64.29% (9/14), not 100%. The 100% figure applies strictly to the 4 targeted categories (SQL, Command, Resource, Error).

---

## 7. Gate 0 Determination: PASS

- Baseline commit `a64e763` is frozen and reproducible.
- Validation branch `validation/milestone-3.4-phase6-independent-pilot` is active.
- Manifest hashes verified.
- Regression suite verified (0 failures).
- Held-out benchmark provenance transparently audited and reclassified.

Phase 0 is complete. The system is cleared to proceed to **Phase 1 — Independent External Repository Acquisition**.

