# MILESTONE 3.4 — PHASE 6 / PHASE 2 PRE-EXECUTION INTEGRITY REPORT
## Immutable Candidate Freeze, Pre-Run Verification & First-Look Audit Certification

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Validation Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Current Commit:** `27d73cd5116673749625576144151d71e6a7cdab`  
**Frozen Detector:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Audit Timestamp:** November 5, 2026  
**Execution Mode:** Blinded Diagnostic First-Look Only (Remediation Blocked)  

---

## 1. Candidate Baseline & Working Tree Verification

- **Validation HEAD Commit SHA:** `27d73cd5116673749625576144151d71e6a7cdab`
- **Frozen Detector Commit SHA:** `a64e7632476c1341343481661a3e246c69e5a644`
- **Detector Code Diff:** Verified byte-for-byte identical (`git diff a64e763 HEAD -- go-backend/agent/repo_intel/` produces 0 output).
- **Working Tree State:** Completely clean (0 modified or untracked files).
- **Toolchain Specification:**
  - Go Toolchain: `go version go1.25.6 windows/amd64`
  - Rust Toolchain: `rustc 1.97.1 (8bab26f4f 2026-07-14)` / `cargo 1.97.1`
  - Host OS: Windows 11 Enterprise (amd64 / x86_64)

---

## 2. Cryptographic Source Integrity & Manifest Verification

All 10 external repository evaluation snapshots were independently hashed and matched against [`benchmarks/external_v2/source_integrity.json`](file:///E:/SWE/benchmarks/external_v2/source_integrity.json):

| Repository | Upstream Module | Origin Ref & Commit | Files | SHA-256 Evaluation Tree Hash | Integrity Audit Result |
|---|---|---|---|---|---|
| `fatih_color` | `github.com/fatih/color` | `v1.15.0` (`12126ed`) | 11 | `5795bbbbb4a5a08a08a527ce36c99268c4420161a712efaad81886e49a175067` | **MATCH (VERIFIED)** |
| `google_uuid` | `github.com/google/uuid` | `v1.3.0` (`44b5fee`) | 25 | `7a7b422ab54fd7ec747f5a7f713d60a2a4a6e84495dd04b0e19b064451efde1d` | **MATCH (VERIFIED)** |
| `gorilla_websocket` | `github.com/gorilla/websocket` | `v1.5.0` (`9111bb8`) | 52 | `8b212e7233c1baeb3975bd4d0f5a17a4a0a424e57ef8f01ea0d678f0c9af6b2c` | **MATCH (VERIFIED)** |
| `hashicorp_go_version` | `github.com/hashicorp/go-version` | `v1.6.0` (`78d058c`) | 11 | `28b6f4731229af57b745da2dde95ffac6d4ffbb2f261a117e886270a6fe392e6` | **MATCH (VERIFIED)** |
| `jinzhu_inflection` | `github.com/jinzhu/inflection` | `master` (`041403f`) | 6 | `b5cb5e50cc52d6abce4f9a58395d0cbb91a8b380e1e493f50214370b4e1e108d` | **MATCH (VERIFIED)** |
| `pkg_errors` | `github.com/pkg/errors` | `v0.9.1` (`614d223`) | 16 | `4169b542218532c240a643e92055017722cde2df456d7e490d81e2c346ce7fe4` | **MATCH (VERIFIED)** |
| `schollz_progressbar` | `github.com/schollz/progressbar` | `v3.13.0` (`17ae25c`) | 21 | `a813c1efbbd45ba8269d6a1022a4fc8b00cbb4cfbc26d9de9a47371cd6dabb99` | **MATCH (VERIFIED)** |
| `spf13_cast` | `github.com/spf13/cast` | `v1.5.1` (`bbed555`) | 14 | `b0152a97d320988ed09e62af86c3c4b461cfac7a2b9e033ca0351508b8205c35` | **MATCH (VERIFIED)** |
| `stretchr_objx` | `github.com/stretchr/objx` | `v0.5.0` (`50a2c06`) | 88 | `d8a69fc5a9f1109450c655907dab736b154472e14266990d599c6a3261500adb` | **MATCH (VERIFIED)** |
| `tidwall_gjson` | `github.com/tidwall/gjson` | `v1.14.4` (`e14b8d3`) | 9 | `e262e736ca5b0212c0e044736b57d60534936481467e8435b85e1ac39341b93a` | **MATCH (VERIFIED)** |

---

## 3. Evaluation Denominator & Stratification Binding

Under [`benchmarks/external_v2/ground_truth_amendment_v2_1.json`](file:///E:/SWE/benchmarks/external_v2/ground_truth_amendment_v2_1.json):
1. **Natural Defect Stratum ($n=0$):** All 4 unverified/invalid natural candidates are disqualified and formally excluded. Natural-defect recall is explicitly documented as **NOT ESTIMABLE**.
2. **Injected Defect Stratum ($n=6$):** Primary positive defect denominator is fixed at exactly **6 verified injected defects** (`INJ-EXT2-SQL-01`, `INJ-EXT2-SQL-02`, `INJ-EXT2-CMD-01`, `INJ-EXT2-RES-02`, `INJ-EXT2-RES-03`, `INJ-EXT2-ERR-04`).
3. **Clean Control Stratum ($n=20$):** Primary negative control denominator is fixed at exactly **20 audited clean controls** (`CTL-EXT2-01` through `CTL-EXT2-20`).
4. **Uncertain / Excluded Code:** Test files and vendor directories are excluded from confusion-matrix calculations. Findings outside labeled ranges are routed to unadjudicated findings.

---

## 4. Prior Execution Audit & Blinding Confirmation

- Prior executions of the diagnostic engine on `benchmarks/external_v2`: **0 runs**.
- No prediction files, cache artifacts, or pre-computed results exist for `external_v2`.
- The first-look evaluation will be executed strictly on the frozen detector without post-hoc tuning.

**Pre-Execution Integrity Gate: PASS**. Cleared for frozen first-look execution.

