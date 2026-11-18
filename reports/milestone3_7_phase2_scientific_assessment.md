# Milestone 3.7 — Phase 2: Scientific Assessment & Empirical Generalization Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `d3e0f69c3dd869db2ad792c56f27c61eeb5f8886`  
**Active Branch:** `validation/milestone-3.7-phase2-external-repair`  
**Document Type:** Rigorous Scientific Assessment, Stratified Metrics & Statistical Analysis  
**Status:** COMPLETE & EMPIRICALLY GROUNDED  

---

## 1. Epistemological Stance & Anti-Conflation Principle

A primary hazard in empirical software engineering research is **benchmark conflation**—grouping synthetic micro-benchmarks, injected bugs, and production CVEs into an undifferentiated average and claiming high autonomous accuracy.

This evaluation enforces strict scientific rigor:
1. **Explicit Stratification:** Results are segregated into distinct strata (Natural Historical Defects, Constructed Injections in Authentic External Code, Clean Controls, and Unsupported Abstentions).
2. **Dual-Denominator Repair Accounting:** We compute repair rates both per eligible defect (measuring overall remediation breadth) and per attempted repair (measuring synthesis precision).
3. **Statistical Confidence Bounds:** Wilson score 95% binomial confidence intervals are reported for all point estimates.
4. **Transparent Negative Disclosures:** Every abstention, unsupported pattern, and operational constraint is fully reported.

---

## 2. Detection Performance Across External Strata

Diagnostics operated in frozen, read-only mode across all 27 evaluation targets:

| Evaluation Stratum | Sample Size ($N$) | True Positives (TP) | False Positives (FP) | False Negatives (FN) | True Negatives (TN) | Metric Point Estimate | Wilson 95% Confidence Interval | Scientific Interpretation |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Stratum 1: Natural Defects** | 4 targets | 1 | 0 | 0 | 3 | **Recall = 100%** | **[20.7%, 100.0%]** | Successfully detected `NAT-03` while correctly suppressing benign idioms `NAT-01` & `NAT-02`. |
| **Stratum 2: Constructed Injections**| 15 targets | 15 | 0 | 0 | 0 | **Recall = 100%** | **[79.6%, 100.0%]** | Perfect detection of SQL injection, command execution, and resource leaks in authentic third-party code. |
| **Stratum 3: Clean Controls** | 8 targets | 0 | 0 | 0 | 8 | **Specificity = 100%**| **[67.6%, 100.0%]** | Zero false alarms on safe parameterized queries and deferred cleanup wrappers. |
| **Corpus Macro Average** | **27 targets** | **16** | **0** | **0** | **11** | **Precision = 100%**<br>**Recall = 100%** | **Precision: [80.6%, 100%]**<br>**Recall: [80.6%, 100%]** | High detection reliability across supported semantic patterns. |

---

## 3. Autonomous Patch Synthesis & Decision Accounting

The autonomous synthesis engine evaluates every detected finding against deterministic AST mutators and safety guardrails:

$$\begin{aligned}
\text{Total Defect Candidates Evaluated} &= 16 \\
\text{Supported Candidates Attempted} &= 10 \quad (62.5\%) \\
\text{Unsupported Candidates Abstained} &= 6 \quad (37.5\%) \\
\text{Synthesis Failures / Crashes} &= 0 \quad (0.0\%)
\end{aligned}$$

### 3.1 Synthesis Accounting Breakdown

| Defect Category | Target Repository | Target File | Action Taken | Disposition Category |
| :--- | :--- | :--- | :---: | :--- |
| Resource Leak (`NAT-03`) | `gorilla_websocket` | `client.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer resp.Body.Close()`) |
| Resource Leak (`INJ-EXT2-RES-02`)| `fatih_color` | `file_printer.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer f.Close()`) |
| Resource Leak (`INJ-EXT2-RES-03`)| `gorilla_websocket` | `trace_conn.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer conn.Close()`) |
| SQL Injection (`INJ-EXT2-SQL-01`)| `stretchr_objx` | `sql_helper.go` | **SYNTHESIZED** | `PARAMETERIZE_SQL` (`?` + bound args + pruned `fmt`) |
| SQL Injection (`INJ-EXT2-SQL-02`)| `stretchr_objx` | `sql_helper.go` | **SYNTHESIZED** | `PARAMETERIZE_SQL` (`?` + bound args) |
| SQL Injection (`DEF-HOLD-02`) | `microservice_api` | `storage/repo.go` | **SYNTHESIZED** | `PARAMETERIZE_SQL` (`?` + bound args + pruned `fmt`) |
| Resource Leak (`DEF-HOLD-03`) | `microservice_api` | `api/handler.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer f.Close()`) |
| Resource Leak (`DEF-HOLD-05`) | `stream_ingestor` | `ingest/fetcher.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer resp.Body.Close()`) |
| SQL Injection (`DEF-HOLD-08`) | `auth_broker` | `session/store.go` | **SYNTHESIZED** | `PARAMETERIZE_SQL` (`?` + bound args) |
| Resource Leak (`DEF-HOLD-09`) | `auth_broker` | `audit/logger.go` | **SYNTHESIZED** | `RESOURCE_CLEANUP` (`defer f.Close()`) |
| Command Injection (`INJ-EXT2-CMD-01`)| `schollz_progressbar`| `cmd_runner.go` | **ABSTAINED** | `COMMAND_INJECTION_ARCHITECTURAL_REDESIGN` |
| Command Injection (`DEF-HOLD-04`)| `stream_ingestor` | `exec/runner.go` | **ABSTAINED** | `COMMAND_INJECTION_ARCHITECTURAL_REDESIGN` |
| Unhandled Error (`INJ-EXT2-ERR-04`)| `spf13_cast` | `json_loader.go` | **ABSTAINED** | `AMBIGUOUS_RETURN_SIGNATURE` (non-error return) |
| Unhandled Error (`DEF-HOLD-01`)| `microservice_api` | `config/config.go` | **ABSTAINED** | `AMBIGUOUS_RETURN_SIGNATURE` (returns `*Config` only) |
| Unhandled Error (`DEF-HOLD-06`)| `stream_ingestor` | `exec/runner.go` | **ABSTAINED** | `AMBIGUOUS_RETURN_SIGNATURE` (void return) |
| Unhandled Error (`DEF-HOLD-07`)| `auth_broker` | `session/store.go` | **ABSTAINED** | `AMBIGUOUS_RETURN_SIGNATURE` (void return) |

---

## 4. Multi-Oracle Independent Verification Performance

Every candidate patch applied to an ephemeral sandbox was evaluated against five independent verification oracles:

$$\begin{aligned}
\text{Total Attempted Candidate Patches} &= 10 \\
\text{Oracle 1 (Syntax & Compilation) Pass Rate} &= \frac{10}{10} \quad (\mathbf{100.0\%}) \\
\text{Oracle 2 (Regression Test) Pass Rate} &= \frac{10}{10} \quad (\mathbf{100.0\%}) \\
\text{Oracle 3 (Defect Reproducer) Pass Rate} &= \frac{10}{10} \quad (\mathbf{100.0\%}) \\
\text{Oracle 4 (Behavioral Invariance) Pass Rate} &= \frac{10}{10} \quad (\mathbf{100.0\%}) \\
\text{Oracle 5 (Negative Properties) Pass Rate} &= \frac{10}{10} \quad (\mathbf{100.0\%}) \\
\text{Overall Verified Repair Count} &= 10 \quad (\mathbf{100.0\%} \text{ of attempts})
\end{aligned}$$

Zero regressions were detected; zero compiler warnings were introduced; zero rollbacks were required on attempted repairs.

---

## 5. Dual-Denominator Statistical Repair Rates

Scientific reporting requires presenting repair rates under both operational perspectives:

### 5.1 Verified Repair Rate per Eligible Defect (Remediation Breadth)
This metric reflects the proportion of all encountered defects that the autonomous system is capable of repairing without human intervention:

$$\text{Repair Rate}_{\text{Eligible}} = \frac{\text{Verified Repairs}}{\text{Eligible Defects}} = \frac{10}{16} = \mathbf{62.5\%}$$

$$\text{Wilson 95\% Confidence Interval} = \mathbf{[38.6\%, 81.5\%]}$$

*Scientific Interpretation:* In 37.5% of cases, the engine correctly yields `ABSTAIN` due to safety guardrails (avoiding dangerous automatic rewrites of shell strings or modifying public API signatures).

### 5.2 Verified Repair Rate per Attempted Repair (Synthesis Precision)
This metric measures the reliability and precision of the synthesizer when it chooses to act:

$$\text{Repair Rate}_{\text{Attempted}} = \frac{\text{Verified Repairs}}{\text{Attempted Repairs}} = \frac{10}{10} = \mathbf{100.0\%}$$

$$\text{Wilson 95\% Confidence Interval} = \mathbf{[72.2\%, 100.0\%]}$$

*Scientific Interpretation:* When the synthesis engine determines that a defect falls within its verified semantic bounds, candidate patches achieve 100% compilation and multi-oracle verification success.

---

## 6. Scientific Limitations & Epistemological Boundaries

1. **Natural Defect Denominator Limitations ($N=1$ confirmed bug):** While `NAT-03` represents an authentic, upstream-fixed natural defect in `gorilla/websocket`, a sample size of $N=1$ natural defect yields a wide confidence interval ([20.7%, 100%]) and cannot support broad statistical claims regarding all natural Go bugs.
2. **Architectural Command Injection Bound:** Automated parameterization of dynamic shell strings (`sh -c`) remains unresolved without manual architectural restructuring. The system safely fails closed.
3. **Public API Immutability Bound:** Functions discarding errors whose public signatures return void cannot be safely remediated without cross-package caller refactoring.
