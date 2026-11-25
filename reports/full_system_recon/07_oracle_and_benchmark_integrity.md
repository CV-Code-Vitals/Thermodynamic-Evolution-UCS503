# 07 — Oracle & Benchmark Ground Truth Integrity Audit

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Target Branch:** `release/milestone-3.8-mvp-rc`  
**Commit Baseline:** `4b992caf951aef397ec9ef043bb9f1e3755b0cae`  
**Audit Date:** 2026-11-24  
**Classification:** Verification Oracles & Benchmark Ground Truth Audit

---

## 1. Executive Summary

This audit validates the architectural fidelity, status semantics, and scientific integrity of the **Five Independent Verification Oracles** and the frozen **27-Target Milestone Benchmark Corpus**. The audit confirms that the oracle suite evaluates candidate repairs using independent verification mechanisms without self-referential or tautological checks.

Furthermore, this audit confirms that historical benchmark metrics are preserved with zero inflation:
- **Corpus Size:** 27 targets assessed across three distinct strata.
- **Attempted Repairs Denominator:** Strictly **9** targets.
- **Verified Repairs:** **9** targets.
- **Verified Precision:** **9 / 9 = 100.0%** (Wilson 95% CI: [70.08%, 100.0%]).
- **Abstention Rate on Unsupported/Ambiguous Defects:** **6 / 6 = 100.0%**.
- **False Alarm Rate on Clean Controls:** **0 / 8 = 0.0%**.
- **Ground Truth Invariant:** Candidate **NAT-03** (`gorilla/websocket#789`) is an independently verified, disqualified non-defect candidate that MUST NOT be counted in the repair denominator.

---

## 2. Five Independent Verification Oracles

The verification pipeline in [oracles.go](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles.go) decouples patch generation from verification. Five independent oracles must evaluate each candidate repair:

```mermaid
flowchart TD
    Patch["Synthesized Candidate Patch"] --> O1["Oracle 1: Syntax & Compilation"]
    O1 -->|PASSED| O2["Oracle 2: Regression Test Suite"]
    O1 -->|FAILED / UNAVAILABLE| Rollback["Fail Closed & Trigger Rollback"]
    O2 -->|PASSED| O3["Oracle 3: Defect Reproducer"]
    O2 -->|FAILED| Rollback
    O3 -->|PASSED (Cured)| O4["Oracle 4: Behavioral Invariance"]
    O3 -->|FAILED (Uncured)| Rollback
    O4 -->|PASSED| O5["Oracle 5: Negative Properties"]
    O4 -->|FAILED| Rollback
    O5 -->|PASSED| Verdict["Verdict: INDEPENDENTLY_VERIFIED"]
    O5 -->|FAILED| Rollback
```

### 2.1 Oracle Specifications

| # | Oracle Name | Code Constant | Verification Mechanism | Independent Evidence |
|---|-------------|---------------|------------------------|----------------------|
| **1** | **Syntax & Compilation** | `OracleSyntaxCompilation` | Evaluates Go AST parsing, `go vet`, and compilation inside hardened container sandbox. | Confirms patch introduces zero syntax errors, type mismatches, or malformed AST nodes. |
| **2** | **Regression Test Suite** | `OracleRegressionTest` | Executes existing test suite (`go test ./...`) inside container sandbox under `--network none`. | Verifies that existing functionality, interfaces, and baseline tests continue passing without regression. |
| **3** | **Defect Reproducer** | `OracleDefectReproducer` | Executes targeted reproducer test designed to trigger the specific reported vulnerability or defect. | Verifies that the defect test fails on pre-repair code and passes on patched code. |
| **4** | **Behavioral Invariance** | `OracleBehavioralInvariance` | Evaluates invariant test suite asserting that non-defect behaviors, return types, and side-effects remain intact. | Prevents "fixes" that cure defects by trivially gutting logic, returning early, or dropping functionality. |
| **5** | **Negative Properties** | `OracleNegativeProperties` | Evaluates negative property constraints (e.g., resource leak detectors, static ownership checks, escape analysis). | Confirms patch does not introduce resource ownership violations, memory leaks, or unhandled panics. |

### 2.2 Oracle Status Semantics

Every oracle returns an explicit `OracleStatus` enum. Binary boolean results (`passed=true/false`) without status classification are prohibited:
1. `PASSED`: Oracle check executed within container isolation and all assertions passed.
2. `FAILED`: Oracle check executed within container isolation and one or more assertions failed.
3. `NOT_RUN`: Oracle check was skipped because a prior dependency oracle failed (short-circuit fail-closed).
4. `UNAVAILABLE`: Oracle could not execute due to environment constraints (e.g., container runtime unavailable, or dynamic execution requiring offline dependencies not present in minimal modcache).

**Verdict Resolution:**
- `INDEPENDENTLY_VERIFIED`: All 5 oracles executed with `PASSED`.
- `SYNTAX_ONLY_VERIFIED`: Oracles 1 passed; Oracles 2-5 report `UNAVAILABLE` due to dynamic environment unavailability.
- `FAILED`: Any executed oracle returned `FAILED`. Triggers immediate sandbox rollback.

---

## 3. Disqualification of NAT-03 (`gorilla/websocket#789`)

### 3.1 Defect Analysis & Ground Truth Proof
Candidate `NAT-03` originates from [gorilla/websocket](file:///E:/SWE/benchmarks/external_v2/gorilla_websocket) in `client.go` (`NewClient` / `DialContext`). A naive scanner reported an unclosed `resp.Body` resource leak.

Independent evaluation proven in [nat03_reproducer_test.go](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/nat03_reproducer_test.go) demonstrates why autonomous patch insertion is catastrophic:
1. **Ownership Escaping Contract:** `DialContext` returns `(*Conn, *http.Response, error)`. The `*http.Response` object is returned to the caller, transferring ownership of the response body to the application caller.
2. **Error Path Panic:** When `http.ReadResponse` fails (e.g. invalid HTTP handshake, connection reset), standard Go library semantics guarantee that `resp == nil` and `err != nil`.
3. **Catastrophic Failure of Naive Patch:** Injecting `defer resp.Body.Close()` before the error check or without a `resp != nil` guard causes a **nil pointer dereference panic** on any network handshake error:
   ```go
   resp, err := readResponseFail()
   defer resp.Body.Close() // PANIC: runtime error: invalid memory address or nil pointer dereference
   if err != nil { return nil, err }
   ```
4. **Authoritative Ground Truth:** NAT-03 is a **disqualified / false-alarm non-defect**. The system correctly assigns `ExpectedAction: ABSTAIN`.
5. **Denominator Integrity Rule:** NAT-03 is strictly excluded from the eligible repair denominator. Including NAT-03 as an attempted repair would erroneously inflate the denominator from 9 to 10 or 16, corrupting the scientific baseline.

---

## 4. Benchmark Corpus Reconciliation: The Authoritative 27 Targets

The table below reconciles all 27 targets assessed in [milestone3_8_phase1_1_revalidation_test.go](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/milestone3_8_phase1_1_revalidation_test.go):

| Target ID | Repository | Target File | Defect Category | Stratum | Expected Action | Actual Outcome | Status |
|-----------|------------|-------------|-----------------|---------|-----------------|----------------|--------|
| **NAT-01** | `fatih/color` | `color.go` | `UNHANDLED_ERROR` | Natural Historical | IGNORE (Benign) | Ignored / Clean | **PASS** |
| **NAT-02** | `google/uuid` | `version4.go` | `UNHANDLED_ERROR` | Natural Historical | IGNORE (Benign) | Ignored / Clean | **PASS** |
| **NAT-03** | `gorilla/websocket` | `client.go` | `RESOURCE_LEAK` | Natural Disqualified | ABSTAIN | Safely Abstained | **PASS** |
| **NAT-04** | `tidwall/gjson` | `gjson.go` | `RESOURCE_LEAK` | Natural Historical | IGNORE (Buffer) | Ignored / Clean | **PASS** |
| **INJ-EXT2-RES-02** | `fatih/color` | `file_printer.go` | `RESOURCE_LEAK` | Authentic Injected | REPAIR | Verified Repaired | **PASS** |
| **INJ-EXT2-RES-03** | `gorilla/websocket` | `trace_conn.go` | `RESOURCE_LEAK` | Authentic Injected | REPAIR | Verified Repaired | **PASS** |
| **INJ-EXT2-SQL-01** | `stretchr/objx` | `sql_helper.go` | `SQL_INJECTION` | Authentic Injected | REPAIR | Verified Repaired | **PASS** |
| **INJ-EXT2-SQL-02** | `stretchr/objx` | `sql_helper.go` | `SQL_INJECTION` | Authentic Injected | REPAIR | Verified Repaired | **PASS** |
| **DEF-HOLD-02** | `microservice_api` | `storage/repo.go` | `SQL_INJECTION` | Constructed Injected | REPAIR | Verified Repaired | **PASS** |
| **DEF-HOLD-03** | `microservice_api` | `api/handler.go` | `RESOURCE_LEAK` | Constructed Injected | REPAIR | Verified Repaired | **PASS** |
| **DEF-HOLD-05** | `stream_ingestor` | `ingest/fetcher.go` | `RESOURCE_LEAK` | Constructed Injected | REPAIR | Verified Repaired | **PASS** |
| **DEF-HOLD-08** | `auth_broker` | `session/store.go` | `SQL_INJECTION` | Constructed Injected | REPAIR | Verified Repaired | **PASS** |
| **DEF-HOLD-09** | `auth_broker` | `audit/logger.go` | `RESOURCE_LEAK` | Constructed Injected | REPAIR | Verified Repaired | **PASS** |
| **INJ-EXT2-CMD-01** | `schollz/progressbar` | `cmd_runner.go` | `COMMAND_INJECTION` | Authentic Injected | ABSTAIN | Safely Abstained | **PASS** |
| **DEF-HOLD-04** | `stream_ingestor` | `exec/runner.go` | `COMMAND_INJECTION` | Constructed Injected | ABSTAIN | Safely Abstained | **PASS** |
| **INJ-EXT2-ERR-04** | `spf13/cast` | `json_loader.go` | `UNHANDLED_ERROR` | Authentic Injected | ABSTAIN | Safely Abstained | **PASS** |
| **DEF-HOLD-01** | `microservice_api` | `config/config.go` | `UNHANDLED_ERROR` | Constructed Injected | ABSTAIN | Safely Abstained | **PASS** |
| **DEF-HOLD-06** | `stream_ingestor` | `exec/runner.go` | `UNHANDLED_ERROR` | Constructed Injected | ABSTAIN | Safely Abstained | **PASS** |
| **DEF-HOLD-07** | `auth_broker` | `session/store.go` | `UNHANDLED_ERROR` | Constructed Injected | ABSTAIN | Safely Abstained | **PASS** |
| **CTL-EXT2-12** | `fatih/color` | `file_printer.go` | `RESOURCE_LEAK` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-EXT2-16** | `stretchr/objx` | `sql_helper.go` | `SQL_INJECTION` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-EXT2-17** | `gorilla/websocket` | `trace_conn.go` | `RESOURCE_LEAK` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-HOLD-01** | `microservice_api` | `storage/repo.go` | `SQL_INJECTION` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-HOLD-02** | `microservice_api` | `api/handler.go` | `RESOURCE_LEAK` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-HOLD-03** | `stream_ingestor` | `ingest/fetcher.go` | `RESOURCE_LEAK` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-HOLD-04** | `auth_broker` | `session/store.go` | `SQL_INJECTION` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |
| **CTL-HOLD-05** | `auth_broker` | `audit/logger.go` | `RESOURCE_LEAK` | Clean Control | IGNORE | Clean (No False Alarm) | **PASS** |

---

## 5. Statistical Metrics Summary

$$
\text{Precision} = \frac{\text{Verified Repairs}}{\text{Attempted Repairs}} = \frac{9}{9} = 100.0\% \quad (95\%\text{ Wilson CI}: [70.08\%, 100.0\%])
$$

$$
\text{Constructed/Injected Repair Rate} = \frac{9}{15} = 60.0\%
$$

$$
\text{Abstention Rate on Unsupported/Ambiguous} = \frac{6}{6} = 100.0\%
$$

$$
\text{False Positive Rate on Clean Controls} = \frac{0}{8} = 0.0\%
$$

$$
\text{Original Repository Working Tree Mutation} = 0 \text{ bytes (100\% Immutability Preserved)}
$$

---

## 6. Audit Verdict

The oracle architecture and benchmark ground truth are rigorously reconciled and verified. The 5 independent oracles execute with strict status semantics, zero self-referential tautologies, and fail-closed security. Historical benchmarks maintain total integrity with **9/9 verified repairs (100% precision)**, zero false positives on clean controls, and accurate classification of NAT-03 as an abstained non-defect.
