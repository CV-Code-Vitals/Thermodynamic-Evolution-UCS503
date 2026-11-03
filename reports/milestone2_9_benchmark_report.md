# Milestone 2.9 — Autonomous Self-Healing Benchmark Report
## End-to-End Verification, Scientific Metrics & Controlled Integration

**Benchmark Run ID:** `BENCHMARK-1793709119`  
**Execution Date:** `2026-11-03T18:01:59+05:30`  
**Target Suite:** `E:\SWE\benchmarks\self_healing_suite`  
**Overall Verdict:** **`GO`**  
**Execution Time:** `9280 ms`  

## 1. Executive Summary & Scientific Metrics

### Detection Metrics (Ground Truth Labeled)
| Metric | Value | Description |
|---|---|---|
| **True Positives (TP)** | 8 | Ground-truth defects accurately detected |
| **False Positives (FP)** | 8 | Non-defective clean controls falsely flagged |
| **False Negatives (FN)** | 7 | Ground-truth defects missed by analyzers |
| **True Negatives (TN)** | 2 | Clean controls correctly ignored |
| **Precision** | **50.00%** | TP / (TP + FP) |
| **Recall** | **53.33%** | TP / (TP + FN) |
| **F1 Score** | **51.61%** | Harmonic mean of Precision and Recall |

### Remediation Effectiveness Metrics
| Metric | Value | Formula / Basis |
|---|---|---|
| **Eligible Ground-Truth Defects** | 15 | Known defects selected for remediation |
| **Verified Remediations** | 3 | Ground-truth defects independently resolved |
| **Verified Remediation Rate** | **20.00%** | Verified / Eligible |
| **Introduced Regressions** | 0 | Broken tests or syntax failures introduced |
| **Regression Rate** | **0.00%** | Regressions / IntegratedPatches |
| **False Remediations** | 0 | Patches rolled back or failed validation |
| **False Remediation Rate** | **0.00%** | FalseRemediations / TotalScheduled |
| **Integration Success Rate** | **0.00%** | IntegratedPatches / ApprovedPatches |

## 2. Five-Stage Pipeline Execution Telemetry

| Stage | Subsystem | Status | Duration | Artifacts Produced |
|---|---|---|---|---|
| `STAGE_1_REPOSITORY_INTELLIGENCE` | repo_intel.Orchestrator | **`COMPLETED`** | 80 ms | E:\SWE\reports\repository_intelligence.json, E:\SWE\reports\repository_intelligence.md, E:\SWE\reports\repository_dependency_graph.json |
| `STAGE_2_DIAGNOSTIC_INTELLIGENCE` | diagnostic.DiagnosticEngine | **`COMPLETED`** | 14 ms | E:\SWE\reports\diagnostic_report.json, E:\SWE\reports\finding_validation.json |
| `STAGE_3_REMEDIATION_PLANNING` | planning.RemediationPlanner | **`COMPLETED`** | 185 ms | E:\SWE\reports\remediation_plan.json |
| `STAGE_4_REMEDIATION_EXECUTION` | execution.RemediationExecutor | **`COMPLETED`** | 7445 ms | E:\SWE\reports\execution_report.json, E:\SWE\reports\execution_patch_manifest.json |
| `STAGE_5_CONTROLLED_BRANCH_INTEGRATION` | benchmark.StageRunner.RunStage5Integration | **`COMPLETED`** | 1455 ms |  |

