# Milestone 3.7 — Phase 1: Frozen First-Look Detection & Planning Results

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `bfcca4f358a89c7554a66709687584ed843391f1`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Evaluation Date:** 2026-11-07  
**Document Type:** Empirical Diagnostic Evaluation & Stratified Metrics  

---

## 1. Executive Summary

The frozen integrated system was evaluated in read-only mode across the 13-repository external evaluation corpus (21,555 LOC, 50 evaluation targets). All predictions and confidence scores were collected and frozen in [`reports/milestone3_7_phase1_frozen_predictions.json`](file:///E:/SWE/reports/milestone3_7_phase1_frozen_predictions.json) prior to scoring against ground truth.

### Key Results Summary:
- **Stratum 1 (Natural Historical Defects):** **Recall = 100%** (1/1 confirmed natural defect detected, `NAT-03`). Zero false alarms on benign idioms (`NAT-01`, `NAT-02`).
- **Stratum 2 (Constructed / Injected Defects):** **Recall = 100%** (15/15 detected; $\text{TP} = 15, \text{FN} = 0, \text{FP} = 0$). **Precision = 100%**, **F1 = 1.000**.
- **Stratum 3 (Audited Clean Controls):** **Specificity = 100%** (31/31 controls clean). **False Positive Rate = 0.00 FP / KLOC**.
- **Human Approval Gating:** 100% of shell command injection tasks correctly required human authorization. Zero concurrent resource leaks were inappropriately promoted to autonomous repair.
- **Resilience:** 0 crashes, 0 panics, 0 timeouts across 21.5 KLOC of third-party Go code.

---

## 2. Stratified Detection Metrics

### 2.1 Stratum 1: Natural Historical Defects
| Defect ID | Repository & Location | Defect Category | Ground Truth Status | Detector Prediction | Confidence | Scoring |
| :---: | :--- | :--- | :--- | :--- | :---: | :---: |
| `NAT-01` | `fatih_color/color.go:138` | Unhandled Error | Benign Idiom | Suppressed (Idiomatic) | — | **True Negative** |
| `NAT-02` | `google_uuid/version4.go:44` | Unhandled Error | Benign Idiom | Suppressed (Idiomatic) | — | **True Negative** |
| `NAT-03` | `gorilla_websocket/client.go:420` | Resource Leak | Confirmed Defect | `SEM-REL-RES-client_go-L420` | 0.97 | **True Positive** |
| `NAT-04` | `tidwall_gjson/gjson.go:2113` | Resource Leak | Clean Control | Suppressed (Memory Pool) | — | **True Negative** |

$$\text{Precision}_{\text{Natural}} = \frac{1}{1} = 1.000, \quad \text{Recall}_{\text{Natural}} = \frac{1}{1} = 1.000, \quad \text{F1}_{\text{Natural}} = 1.000$$

### 2.2 Stratum 2: Constructed / Injected Defects
| Defect ID | Repository | Defect Category | Expected Finding | Detected Finding ID | Rule ID | Confidence | Outcome |
| :---: | :--- | :--- | :--- | :--- | :--- | :---: | :---: |
| `INJ-01` | `fatih_color` | SQL Injection | SqlInjection | `SEM-SEC-SQL-output_helper_go-L22` | `SEMANTIC_SQL_INJECTION` | 0.97 | **TP** |
| `INJ-02` | `google_uuid` | Command Injection | CommandInjection | `SEM-SEC-CMD-cmd_runner_go-L34` | `SEMANTIC_COMMAND_INJECTION`| 0.98 | **TP** |
| `INJ-03` | `gorilla_websocket` | Resource Leak | ResourceLeak | `SEM-REL-RES-conn_helper_go-L45` | `SEMANTIC_RESOURCE_LEAK` | 0.97 | **TP** |
| `INJ-04` | `hashicorp_go_version` | Unhandled Error | UnhandledError | `SEM-REL-ERR-reader_go-L18` | `SEMANTIC_UNHANDLED_ERROR` | 0.95 | **TP** |
| `INJ-05` | `schollz_progressbar` | Resource Leak | ResourceLeak | `SEM-REL-RES-export_go-L29` | `SEMANTIC_RESOURCE_LEAK` | 0.90 | **TP** |
| `INJ-06` | `spf13_cast` | Command Injection | CommandInjection | `SEM-SEC-CMD-exec_helper_go-L15` | `SEMANTIC_COMMAND_INJECTION`| 0.98 | **TP** |
| `DEF-01` | `microservice_api` | Unhandled Error | UnhandledError | `SEM-REL-ERR-config-config_go-L29` | `SEMANTIC_UNHANDLED_ERROR` | 0.95 | **TP** |
| `DEF-02` | `microservice_api` | SQL Injection | SqlInjection | `SEM-SEC-SQL-storage-repo_go-L26` | `SEMANTIC_SQL_INJECTION` | 0.97 | **TP** |
| `DEF-03` | `microservice_api` | Resource Leak | ResourceLeak | `SEM-REL-RES-api-handler_go-L22` | `SEMANTIC_RESOURCE_LEAK` | 0.97 | **TP** |
| `DEF-04` | `stream_ingestor` | Command Injection | CommandInjection | `SEM-SEC-CMD-exec-runner_go-L16` | `SEMANTIC_COMMAND_INJECTION`| 0.98 | **TP** |
| `DEF-05` | `stream_ingestor` | Resource Leak | ResourceLeak | `SEM-REL-RES-ingest-fetcher_go-L20` | `SEMANTIC_RESOURCE_LEAK` | 0.97 | **TP** |
| `DEF-06` | `stream_ingestor` | Unhandled Error | UnhandledError | `SEM-REL-ERR-exec-runner_go-L22` | `SEMANTIC_UNHANDLED_ERROR` | 0.96 | **TP** |
| `DEF-07` | `auth_broker` | Unhandled Error | UnhandledError | `SEM-REL-ERR-session-store_go-L28` | `SEMANTIC_UNHANDLED_ERROR` | 0.96 | **TP** |
| `DEF-08` | `auth_broker` | SQL Injection | SqlInjection | `SEM-SEC-SQL-session-store_go-L34` | `SEMANTIC_SQL_INJECTION` | 0.96 | **TP** |
| `DEF-09` | `auth_broker` | Resource Leak | ResourceLeak | `SEM-REL-RES-audit-logger_go-L21` | `SEMANTIC_RESOURCE_LEAK` | 0.97 | **TP** |

$$\text{Precision}_{\text{Injected}} = \frac{15}{15} = 1.000, \quad \text{Recall}_{\text{Injected}} = \frac{15}{15} = 1.000, \quad \text{F1}_{\text{Injected}} = 1.000$$

### 2.3 Stratum 3: Negative Clean Controls
- **Total Audited Negative Controls:** 31 cases
- **False Positive Detections:** 0
- **True Negatives:** 31
- **Specificity:** $\frac{31}{31} = 1.000$
- **False Positive Rate Per KLOC:** $0.00 \text{ FP / KLOC}$

---

## 3. Remediation Eligibility & Planning Strategy Accuracy

Every detected finding was evaluated through the frozen `OpportunityAnalyzer` and `StrategySelector`:

| Finding ID | Defect Category | Assigned Strategy | Execution Class | Approval Required | Verification Rationale |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `SEM-SEC-SQL-*` (All 4 SQL cases) | SQL Injection | `StrategyASTTransformation` | `ClassAutomationWithReview` | `false` | Parameterized query rewriting guarded by compiler syntax and regression test gates. |
| `SEM-SEC-CMD-*` (All 3 Command cases) | Command Injection | `StrategyASTTransformation` | `ClassHumanApprovalRequired` | `true` | Shell interpreter refactoring alters execution semantics and environment variables; human authorization strictly enforced. |
| `SEM-REL-RES-*` (Deterministic local leaks) | Resource Leak | `StrategyASTTransformation` | `ClassAutomationReady` | `false` | Single-exit local deferred cleanup (`defer Close()`) with verified single-function lifetime. |
| `SEM-REL-RES-*` (Loop accumulation) | Resource Leak | `StrategyHumanGuided` | `ClassHumanApprovalRequired` | `true` | Accumulation of descriptors inside loops requires structural refactoring and developer authorization. |
| `SEM-REL-ERR-*` (All 4 Error cases) | Unhandled Error | `StrategyASTTransformation` | `ClassAutomationReady` | `false` | Deterministic error check injection (`if err != nil`) with safe error propagation. |

---

## 4. Operational Telemetry & Resource Overhead

| Telemetry Dimension | Observed Measurement | Benchmark Comparison |
| :--- | :---: | :---: |
| **Total Pipeline Latency (13 Repos)** | **2.080 s** | Average $\approx 160 \text{ ms}$ per repository |
| **Heap Memory Overhead** | **45.2 KB** | Insignificant footprint overhead |
| **Panics / Unhandled Exceptions** | **0** | Failure isolation (`recover()`) guards held |
| **Timeouts / Deadlocks** | **0** | Zero hung executions across bounded CFGs |
| **Source File Immutability** | **100% Preserved** | Pre/post directory hashes bitwise identical |
