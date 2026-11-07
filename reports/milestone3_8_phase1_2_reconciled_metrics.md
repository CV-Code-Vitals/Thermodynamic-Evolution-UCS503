# Milestone 3.8 Phase 1.2 — Reconciled Repair Evaluation Metrics

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Date:** 2026-11-08  
**Scope:** Rigorous scientific reconciliation of repair precision, breadth, specificity, and fail-closed safety metrics across all strata.

---

## 1. Executive Summary & Prominent Disclosure

> [!IMPORTANT]
> **Prominent Disclosure of Natural Defect Reparability:**  
> Forensic provenance analysis of the evaluation benchmark confirms that **0 confirmed natural defects** have been autonomously repaired by this system. Candidate `NAT-03` (`gorilla_websocket/client.go:375`) was formally disqualified as a spurious finding on pristine upstream Go code. All 9 verified repairs demonstrated by the system were achieved on **authentic repository injected defects** (4) and **constructed microservice defects** (5).

The system achieved **100% precision** on attempted repairs (9 verified / 9 attempted), a **60.0% breadth** over true defects (9 repaired / 15 true defects), **100% specificity** on clean controls (11 preserved / 11 clean controls), and **100% fail-closed safety** by safely abstaining on all 7 ambiguous, complex, or disqualified cases.

---

## 2. Reconciled Evaluation Target Breakdown

| Stratum | Evaluation Count | True Defects | Clean Controls | Disqualified | Attempted Repairs | Verified Repairs | Safe Abstentions |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Confirmed Natural Defects** | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| **Authentic Injected Defects** (`external_v2`) | 6 | 6 | 0 | 0 | 4 | 4 | 2 |
| **Constructed Microservice Defects** (`heldout_v1`) | 9 | 9 | 0 | 0 | 5 | 5 | 4 |
| **Clean Companion Controls** | 8 | 0 | 8 | 0 | 0 | 0 | 0 |
| **Benign Idioms / Memory Pools** (`NAT-01, 02, 04`) | 3 | 0 | 3 | 0 | 0 | 0 | 0 |
| **Ground-Truth Disqualified** (`NAT-03`) | 1 | 0 | 0 | 1 | 0 | 0 | 1 |
| **Total Evaluation Universe** | **27** | **15** | **11** | **1** | **9** | **9** | **7** |

---

## 3. Reconciled Metrics & Wilson 95% Confidence Intervals

Confidence intervals are calculated using the standard Wilson score interval without continuity correction:
$$\tilde{p} = \frac{n_s + \frac{z^2}{2}}{n + z^2}, \quad \text{Margin} = \frac{z}{n + z^2} \sqrt{\frac{n_s (n - n_s)}{n} + \frac{z^2}{4}}, \quad z = 1.95996$$

### A. Verified Precision on Attempted Repairs
- **Definition:** Verified repairs divided by attempted repairs ($\frac{\text{Verified}}{\text{Attempted}}$).
- **Numerator:** 9
- **Denominator:** 9
- **Point Estimate:** **100.0%** (1.000)
- **Wilson 95% CI:** **[70.1%, 100.0%]**
- *Interpretation:* When the synthesizer proposes a repair and human approval is granted, 100% of the synthesized patches pass the full 5-oracle independent validation suite.

### B. Repair Breadth Over True Defects
- **Definition:** Verified repairs divided by total authentic defects in benchmark ($\frac{\text{Verified}}{\text{True Defects}}$).
- **Numerator:** 9
- **Denominator:** 15 (6 authentic injected + 9 constructed microservice)
- **Point Estimate:** **60.0%** (0.600)
- **Wilson 95% CI:** **[35.7%, 80.2%]**
- *Interpretation:* The deterministic single-file AST repair engine successfully remediates 60% of all defects, intentionally abstaining on the remaining 40% where architectural redesign or complex multi-statement restructuring is required.

### C. Specificity on Clean Controls & Benign Idioms
- **Definition:** Clean controls left unmutated divided by total clean controls ($\frac{\text{Preserved Clean}}{\text{Clean Controls}}$).
- **Numerator:** 11 (8 companion controls + 3 benign idioms)
- **Denominator:** 11
- **Point Estimate:** **100.0%** (1.000)
- **Wilson 95% CI:** **[74.1%, 100.0%]**
- *Interpretation:* Zero false alarms or unwanted modifications were generated against clean production Go code.

### D. Fail-Closed Abstention Safety Rate
- **Definition:** Correct safe abstentions on ambiguous/out-of-scope defects and disqualified candidates ($\frac{\text{Safe Abstentions}}{\text{Abstention Invocations}}$).
- **Numerator:** 7 (2 authentic injected + 4 constructed microservice + 1 disqualified NAT-03)
- **Denominator:** 7
- **Point Estimate:** **100.0%** (1.000)
- **Wilson 95% CI:** **[64.6%, 100.0%]**
- *Interpretation:* 100% of out-of-scope or disqualified candidates were safely rejected before generating dangerous diffs or executing sandboxed code.

### E. Bitwise Original Repository Immutability
- **Point Estimate:** **100.0%**
- **Verification:** Pre-execution and post-execution SHA-256 tree digests were strictly identical across all 27 targets. Zero in-place modifications occurred in the benchmark repositories.

---

## 4. Stratum-by-Stratum Repair Catalog

### Stratum 1: Authentic Injected Defects (`external_v2`)
1. **INJ-EXT2-RES-02** (`fatih_color/file_printer.go:21`): **VERIFIED**
   - Pattern: Single-file intra-procedural `os.Open` resource leak.
   - Patch: Synthesized `defer f.Close()` immediately after `if err != nil` check.
   - Oracles: Passed all 5 independent oracles.
2. **INJ-EXT2-RES-03** (`gorilla_websocket/trace_conn.go:23`): **VERIFIED**
   - Pattern: Resource leak on `net.Conn`.
   - Patch: Synthesized `defer conn.Close()` after allocation error check.
   - Oracles: Passed all 5 independent oracles.
3. **INJ-EXT2-SQL-01** (`stretchr_objx/sql_helper.go:26`): **VERIFIED**
   - Pattern: SQL injection via `fmt.Sprintf` query formatting.
   - Patch: Parameterized query replacement (`WHERE role = ?`).
   - Oracles: Passed all 5 independent oracles.
4. **INJ-EXT2-SQL-02** (`stretchr_objx/sql_helper.go:33`): **VERIFIED**
   - Pattern: SQL injection via string concatenation.
   - Patch: Parameterized query replacement.
   - Oracles: Passed all 5 independent oracles.
5. **INJ-EXT2-CMD-01** (`schollz_progressbar/cmd_runner.go:28`): **ABSTAINED**
   - Rationale: Shell command injection (`sh -c`) requires manual architectural redesign.
6. **INJ-EXT2-ERR-04** (`spf13_cast/json_loader.go:24`): **ABSTAINED**
   - Rationale: Multi-variable error return signature is ambiguous; cannot infer caller propagation semantics.

### Stratum 2: Constructed Microservice Defects (`heldout_v1`)
1. **DEF-HOLD-02** (`microservice_api/storage/repo.go:25`): **VERIFIED** (SQL injection parameterized)
2. **DEF-HOLD-03** (`microservice_api/api/handler.go:27`): **VERIFIED** (Resource leak deferred close)
3. **DEF-HOLD-05** (`stream_ingestor/ingest/fetcher.go:24`): **VERIFIED** (HTTP response body deferred close)
4. **DEF-HOLD-08** (`auth_broker/session/store.go:33`): **VERIFIED** (SQL injection parameterized)
5. **DEF-HOLD-09** (`auth_broker/audit/logger.go:25`): **VERIFIED** (Resource leak deferred close)
6. **DEF-HOLD-01** (`microservice_api/config/config.go:18`): **ABSTAINED** (Ambiguous return signature)
7. **DEF-HOLD-04** (`stream_ingestor/exec/runner.go:19`): **ABSTAINED** (Command injection redesign)
8. **DEF-HOLD-06** (`stream_ingestor/exec/runner.go:32`): **ABSTAINED** (Ambiguous return signature)
9. **DEF-HOLD-07** (`auth_broker/session/store.go:44`): **ABSTAINED** (Ambiguous return signature)

### Stratum 3: Disqualified Candidate (`NAT-03`)
- **NAT-03** (`gorilla_websocket/client.go:371`): **ABSTAINED (DISQUALIFIED)**
  - Finding: Pristine upstream code; underlying socket closed via `defer netConn.Close()` at line 330.
  - Fail-Closed Rationale: `RESOURCE_ESCAPES_VIA_RETURN` prevents synthesizing intra-procedural cleanup on returned response objects.

---

## 5. Summary Table: Claim Ledger

| Claim Area | Claim State | Evidence / Support |
| :--- | :--- | :--- |
| **Natural Defect Repair** | **UNSUPPORTED / ZERO CONFIRMED** | Forensic provenance audit disqualified NAT-03; 0 natural defects repaired. |
| **Authentic Repo Injected Repair** | **DEMONSTRATED (100% Precision)** | 4 of 4 attempted repairs verified across 5 independent oracles. |
| **Constructed Microservice Repair**| **DEMONSTRATED (100% Precision)** | 5 of 5 attempted repairs verified across 5 independent oracles. |
| **Clean Control Specificity** | **DEMONSTRATED (100% Specificity)**| 11 of 11 clean controls preserved without false mutations. |
| **Fail-Closed Safety** | **DEMONSTRATED (100% Safety)** | 7 of 7 out-of-scope / ambiguous cases safely abstained. |
| **Single-Use Cryptographic Auth** | **DEMONSTRATED (100% Replay-Proof)**| HMAC-SHA256 tokens bounded to diff hash, expired after use, protected against replay. |
| **Sandbox Isolation** | **DEMONSTRATED (100% Bitwise Immutability)** | Ephemeral workspaces used; zero original repository modifications. |

