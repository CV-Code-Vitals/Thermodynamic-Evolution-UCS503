# MILESTONE 3.4 — PHASE 6 GATE 1 FINAL DECISION REPORT
## Reconciled Corpus Accounting, Stratum Certification & Release Determination

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Validation Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Candidate Commit:** `ad18d32`  
**Frozen Detector:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Audit Gate:** Gate 1 Addendum — Final Decision  
**Gate 1 Determination:** **CONDITIONAL PASS** (Corpus Usable with Explicit Exclusions under Amendment v2.1.0)  

---

## 1. Executive Summary

This report delivers the final decision for the Milestone 3.4 Phase 6 Gate 1 Verification Addendum.
In accordance with instructions:
- The diagnostic engine was **not executed**.
- No detector code was modified.
- All 10 external repositories, 20 clean controls, and 6 injected defects were audited for source integrity.
- Ambiguous or unverified natural defect claims were rejected rather than falsely promoted.

---

## 2. Reconciled Benchmark Accounting Ledger

| Stratum / Category | Initial Claimed | Audited Status | Validated Denominator for Phase 2 | Scientific Handling & Restriction |
|---|---|---|---|---|
| **Stratum A: Natural Defects** | 4 candidates | 0 Confirmed, 2 Rejected, 2 Unverified | **0 (EXCLUDED)** | Disqualified from primary defect denominator. Cannot be pooled into headline recall scores. |
| **Stratum B: Injected Defects** | 6 candidates | 6 Verified Present | **6 (CONFIRMED)** | Primary evaluation denominator for Phase 2 ($n=6$). |
| **Stratum C: Clean Controls** | 20 candidates | 20 Verified Benign | **20 (CONFIRMED)** | Primary specificity denominator for Phase 2 ($n=20$). |
| **Stratum D: Uncertain / Excluded** | Tests & Vendor | Excluded | **EXCLUDED** | Unlabeled lines are strictly excluded from true negative counts. |

---

## 3. Stratum-Level & Per-Repository Breakdown

### Confirmed Injected Defect Stratum ($n=6$)
1. `INJ-EXT2-SQL-01` (`stretchr_objx`): SQL Injection via `fmt.Sprintf` query formatting.
2. `INJ-EXT2-SQL-02` (`stretchr_objx`): Dynamic SQL injection via raw concatenation into `db.Query`.
3. `INJ-EXT2-CMD-01` (`schollz_progressbar`): Command Injection via `sh -c` shell execution with untrusted argument.
4. `INJ-EXT2-RES-02` (`fatih_color`): Resource Leak via unclosed `os.Open` descriptor.
5. `INJ-EXT2-RES-03` (`gorilla_websocket`): Resource Leak via unclosed `net.Dial` TCP socket.
6. `INJ-EXT2-ERR-04` (`spf13_cast`): Unhandled Error via discarded `json.Unmarshal` return value.

### Confirmed Clean Control Stratum ($n=20$)
- 15 Controls in pristine upstream files:
  - `google_uuid`: `New` (`CTL-EXT2-01`), `NewHash` (`CTL-EXT2-02`)
  - `pkg_errors`: `Wrap` (`CTL-EXT2-03`), `Frame.Format` (`CTL-EXT2-04`)
  - `jinzhu_inflection`: `Plural` (`CTL-EXT2-05`), `Singular` (`CTL-EXT2-06`)
  - `hashicorp_go_version`: `NewVersion` (`CTL-EXT2-07`), `NewConstraint` (`CTL-EXT2-08`)
  - `spf13_cast`: `ToStringE` (`CTL-EXT2-09`)
  - `fatih_color`: `New` (`CTL-EXT2-11`)
  - `tidwall_gjson`: `Get` (`CTL-EXT2-13`), `Parse` (`CTL-EXT2-14`)
  - `stretchr_objx`: `Map.Get` (`CTL-EXT2-15`)
  - `gorilla_websocket`: `tokenListContainsValue` (`CTL-EXT2-18`)
  - `schollz_progressbar`: `NewOptions64` (`CTL-EXT2-20`)
- 5 Controls in companion injection files:
  - `spf13_cast`: `LoadJSONConfigSafe` (`CTL-EXT2-10`)
  - `fatih_color`: `PrintFileSafe` (`CTL-EXT2-12`)
  - `stretchr_objx`: `QueryUserSafe` (`CTL-EXT2-16`)
  - `gorilla_websocket`: `DialProbeSafe` (`CTL-EXT2-17`)
  - `schollz_progressbar`: `RunTerminalResetSafe` (`CTL-EXT2-19`)

---

## 4. Upstream Snapshot Integrity Verification

All 10 repositories have their cryptographic tree digests locked in [`benchmarks/external_v2/source_integrity.json`](file:///E:/SWE/benchmarks/external_v2/source_integrity.json):
- 5 repositories remain **100% bit-for-bit pristine upstream code**.
- 5 repositories contain exactly **1 newly added standalone `.go` file** housing the isolated injected defects and companion clean controls.
- **Zero existing upstream files were modified or deleted**.

---

## 5. Formal Gate 1 Determination: CONDITIONAL PASS

### Decision Justification:
1. **Source Integrity:** Verified. No unintended or untracked modifications occurred in any of the 10 acquired repositories.
2. **Independence:** Verified. The repositories are authentic open-source libraries. The detector has never been executed against `external_v2`.
3. **Explicit Exclusions Established:** The 4 unverified natural defect candidates have been formally disqualified and excluded via [`benchmarks/external_v2/ground_truth_amendment_v2_1.json`](file:///E:/SWE/benchmarks/external_v2/ground_truth_amendment_v2_1.json).
4. **Controlled Evaluation Scope:** Phase 2 diagnostic evaluation is formally bounded to the **6 certified injected defects** and **20 certified clean controls**.

### Stopping Criterion Enforced:
Work halts immediately at Gate 1. Phase 2 (Blinded Diagnostic Generalization Evaluation) will not begin without explicit user authorization.
