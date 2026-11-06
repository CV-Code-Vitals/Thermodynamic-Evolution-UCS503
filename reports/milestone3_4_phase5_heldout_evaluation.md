# MILESTONE 3.4 — PHASE 5 HELDOUT EVALUATION REPORT
## Independent Held-Out Multi-Package Benchmark Evaluation (Suite v1)

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Baseline Commit:** `be917a4`  
**Evaluation Date:** November 2026  
**Classification:** Independent Evaluation & Statistical Generalization Assessment  

---

## 1. Executive Summary & Experimental Methodology

To evaluate external generalization beyond the historical Milestone 3.3 repositories, a newly constructed held-out evaluation suite (`benchmarks/heldout_v1/`) was implemented and frozen.

### Key Experimental Safeguards
1. **Separation from Training/Development:** The held-out benchmark was not used during the design or tuning of Phases 1–4.
2. **Immutable Cryptographic Freeze:** Every repository directory has its complete file tree hashed via SHA-256 and locked in `benchmarks/heldout_v1/manifest.json`.
3. **Multi-Package Architectures:** Each repository features realistic Go module layouts with multiple interacting internal packages (`api`, `storage`, `config`, `ingest`, `exec`, `telemetry`, `token`, `session`, `audit`).
4. **Independent Ground Truth:** Injected defects and clean negative controls were authored and labeled independently of analyzer internals.
5. **Statistical Interval Estimation:** Wilson score 95% confidence intervals are reported to account for finite sample size limitations.

---

## 2. Benchmark Composition & Cryptographic Digests

| Repository ID | Repository Name | Module Identifier | Internal Packages | File Count | LOC | Tree SHA-256 Digest | Defects | Negative Controls |
|---|---|---|---|---|---|---|---|---|
| `heldout-repo-01` | `microservice_api` | `github.com/heldout-bench/microservice-api` | `api`, `storage`, `config` | 3 | 85 | `5cdf37babeef1e7e67e353d8cdaa612f118410e4428c5a0ee637b536194d2e2d` | 3 | 4 |
| `heldout-repo-02` | `stream_ingestor` | `github.com/heldout-bench/stream-ingestor` | `ingest`, `exec`, `telemetry` | 3 | 76 | `c7773b87a10f77f94eb49e66f9edfd9b453838313ef3adf2c4b1db8e8c74d63e` | 3 | 4 |
| `heldout-repo-03` | `auth_broker` | `github.com/heldout-bench/auth-broker` | `token`, `session`, `audit` | 3 | 79 | `fdee20493174109035ffbaf62cbefd5b4b50102aeaff1ee430245558af8d5544` | 3 | 4 |
| **TOTALS** | **3 Repositories** | — | **9 Packages** | **9 Files** | **240** | — | **9 Defects** | **12 Controls** |

---

## 3. Defect-Level Evaluation Ledger

| Defect ID | Target Repo | File & Function | Target Category | Ground Truth | Detected By | Line | Classification | Outcome |
|---|---|---|---|---|---|---|---|---|
| `DEF-HOLDOUT-01` | `microservice_api` | `config/config.go:LoadConfigUnchecked` | `UNHANDLED_ERROR` | Defect | Error Analyzer | 29 | `SEMANTIC_UNHANDLED_ERROR` | **True Positive (TP)** |
| `DEF-HOLDOUT-02` | `microservice_api` | `storage/repo.go:SearchUsersDynamic` | `SQL_INJECTION` | Defect | SQL Analyzer | 26 | `SEMANTIC_SQL_INJECTION` | **True Positive (TP)** |
| `DEF-HOLDOUT-03` | `microservice_api` | `api/handler.go:ExportReportLeak` | `RESOURCE_LEAK` | Defect | Resource Analyzer | 22 | `SEMANTIC_RESOURCE_LEAK` | **True Positive (TP)** |
| `DEF-HOLDOUT-04` | `stream_ingestor` | `exec/runner.go:DispatchShellTask` | `COMMAND_INJECTION` | Defect | Command Analyzer | 16 | `SEMANTIC_COMMAND_INJECTION` | **True Positive (TP)** |
| `DEF-HOLDOUT-05` | `stream_ingestor` | `ingest/fetcher.go:IngestEventStream` | `RESOURCE_LEAK` | Defect | Resource Analyzer | 20 | `SEMANTIC_RESOURCE_LEAK` | **True Positive (TP)** |
| `DEF-HOLDOUT-06` | `stream_ingestor` | `exec/runner.go:DumpTelemetryLog` | `UNHANDLED_ERROR` | Defect | Error Analyzer | 22 | `SEMANTIC_UNHANDLED_ERROR` | **True Positive (TP)** |
| `DEF-HOLDOUT-07` | `auth_broker` | `session/store.go:InvalidateSessionUnchecked` | `UNHANDLED_ERROR` | Defect | Error Analyzer | 28 | `SEMANTIC_UNHANDLED_ERROR` | **True Positive (TP)** |
| `DEF-HOLDOUT-08` | `auth_broker` | `session/store.go:FindSessionDynamic` | `SQL_INJECTION` | Defect | SQL Analyzer | 34 | `SEMANTIC_SQL_INJECTION` | **True Positive (TP)** |
| `DEF-HOLDOUT-09` | `auth_broker` | `audit/logger.go:RotateAuditLog` | `RESOURCE_LEAK` | Defect | Resource Analyzer | 21 | `SEMANTIC_RESOURCE_LEAK` | **True Positive (TP)** |

### Negative Controls (Benign Code Verification)

All 12 negative controls were verified benign with 0 false positives:
1. `microservice_api/config/config.go`: `LoadConfigSafe` (Checked error) — Clean ($\text{TN}$)
2. `microservice_api/storage/repo.go`: `FindUserSafe` (Parameterized SQL) — Clean ($\text{TN}$)
3. `microservice_api/api/handler.go`: `ExportSafe` (Deferred `f.Close()`) — Clean ($\text{TN}$)
4. `microservice_api/api/handler.go`: `FetchExternalMetadata` (Deferred `resp.Body.Close()`) — Clean ($\text{TN}$)
5. `stream_ingestor/exec/runner.go`: `RunBackupSafe` (Safe argument separation) — Clean ($\text{TN}$)
6. `stream_ingestor/ingest/fetcher.go`: `FetchStreamSafe` (Deferred `resp.Body.Close()`) — Clean ($\text{TN}$)
7. `stream_ingestor/telemetry/metrics.go`: `IncrementRecord` (Safe mutex handling) — Clean ($\text{TN}$)
8. `stream_ingestor/telemetry/metrics.go`: `FormatStats` (Safe string formatting) — Clean ($\text{TN}$)
9. `auth_broker/session/store.go`: `ValidateSessionSafe` (Parameterized SQL) — Clean ($\text{TN}$)
10. `auth_broker/audit/logger.go`: `LogEventSafe` (Deferred `f.Close()`) — Clean ($\text{TN}$)
11. `auth_broker/token/jwt.go`: `VerifySignature` (Cryptographic HMAC check) — Clean ($\text{TN}$)
12. `auth_broker/token/jwt.go`: `ParseTokenSafe` (Properly propagated errors) — Clean ($\text{TN}$)

---

## 4. Aggregate Benchmark Metrics & Statistical Analysis

### Confusion Matrix
- **True Positives (TP):** 9
- **False Positives (FP):** 0
- **True Negatives (TN):** 12
- **False Negatives (FN):** 0
- **Total Evaluated Cases:** 21

$$\text{Precision} = \frac{\text{TP}}{\text{TP} + \text{FP}} = \frac{9}{9 + 0} = 100.00\%$$

$$\text{Recall} = \frac{\text{TP}}{\text{TP} + \text{FN}} = \frac{9}{9 + 0} = 100.00\%$$

$$\text{Specificity} = \frac{\text{TN}}{\text{TN} + \text{FP}} = \frac{12}{12 + 0} = 100.00\%$$

$$F_1 = 2 \times \frac{\text{Precision} \times \text{Recall}}{\text{Precision} + \text{Recall}} = 100.00\%$$

### Wilson Score 95% Confidence Intervals
Given sample sizes ($n=9$ for defect recall, $n=12$ for negative control specificity), point estimates of 100% must be interpreted with proper statistical bounds. Using the Wilson score interval with continuity correction:

$$\tilde{p} = \frac{x + \frac{z^2}{2}}{n + z^2} \pm \frac{z}{n + z^2}\sqrt{\frac{x(n - x)}{n} + \frac{z^2}{4}} \quad (z = 1.96)$$

- **Recall (95% CI):** $[70.09\%, 100.00\%]$
- **Precision (95% CI):** $[70.09\%, 100.00\%]$
- **Specificity (95% CI):** $[75.75\%, 100.00\%]$

### Sample Size & Scope Disclosures
- **Scope Limitation:** The held-out benchmark consists of 3 microservices and 9 target defects across SQL, command, resource, and error domains.
- **Statistical Representation:** While these results demonstrate that the semantic analyzers generalize beyond the training fixtures to previously unseen multi-package structures without false positive regressions, a sample of 21 observations is not sufficient to claim universal reliability across the entirety of open-source Go ecosystems. Broader multi-repository studies remain necessary before unrestricted autonomous remediation can be authorized.

