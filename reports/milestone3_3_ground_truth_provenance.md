# MILESTONE 3.3 — PHASE 1: GROUND TRUTH PROVENANCE & ADJUDICATION ARCHITECTURE
## Lineage of Synthetic Benchmark Suites & Ground-Truth Protocol for External Repositories

**Project:** Autonomous Self-Healing Repository Intelligence Pipeline  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Audit Date:** 2026-11-04  
**Status:** AUDITED & SPECIFIED  

---

## 1. Lineage of Synthetic Ground-Truth Suites

The autonomous pipeline has evolved across three successive benchmark generations:

```
+-------------------------------------------------------------------------------------------------+
| SUITE V1 (Milestone 2.9)                                                                        |
| - Origin: Initial 30-fixture prototype (15 defects, 15 clean controls)                          |
| - Manifest: benchmarks/ground_truth_manifest.json (SHA-256: 5f8841cc...)                        |
| - Scope: Basic SQL injection, command execution, error ignores                                  |
+-------------------------------------------------------------------------------------------------+
                                                |
                                                v
+-------------------------------------------------------------------------------------------------+
| SUITE V2 (Milestone 3.0 & 3.1)                                                                  |
| - Origin: Programmatic generator go-backend/agent/repo_intel/benchmark/v2_builder.go             |
| - Manifest: benchmarks/ground_truth_manifest_v2.json (SHA-256: 79fcea50...)                     |
| - Dimensions: 96 fixtures across 8 categories (64 defects, 32 clean controls)                   |
| - Audit Finding: 62.5% template duplication identified during Milestone 3.1 verification         |
+-------------------------------------------------------------------------------------------------+
                                                |
                                                v
+-------------------------------------------------------------------------------------------------+
| SUITE V3 (Milestone 3.2)                                                                        |
| - Origin: Programmatic generator go-backend/agent/repo_intel/benchmark/v3_builder.go             |
| - Manifest: benchmarks/ground_truth_manifest_v3.json (SHA-256: 0e6ea1b4...)                     |
| - Dimensions: 144 fixtures across 8 categories (96 defects, 48 clean controls)                  |
| - Audit Finding: 0% file collisions with v2; high construct-family coupling with diagnostic rules|
+-------------------------------------------------------------------------------------------------+
```

---

## 2. Cryptographic Integrity of Ground Truth

In all synthetic suites, ground truth was recorded with cryptographic provenance:
- Each fixture embeds `file_sha256` matching the exact byte content of the injected source file.
- The manifest file embeds `manifest_sha256` calculated from the canonical JSON serialization.
- Start and end line coordinates specify exact source positions.

While this ensures that synthetic benchmark files cannot be silently altered or tampered with, **it does not make the benchmark external or independent**.

---

## 3. External Ground-Truth Adjudication Protocol

For external repositories evaluated in Milestone 3.3, ground truth cannot be synthesized programmatically because the repositories were created independently by third-party developers.

To evaluate predictions objectively, Milestone 3.3 institutes a **Six-Class Independent Adjudication Framework**:

| Adjudication Class | Definition | Diagnostic Metric Impact | Remediation Eligibility |
| :--- | :--- | :---: | :---: |
| **`CONFIRMED_DEFECT`** | Genuine defect violating Go language semantics, security, or reliability contracts. | **True Positive (TP)** | Eligible for remediation trial |
| **`VALID_CONCERN`** | Code smell or suboptimal style, but not a defect causing operational or security failure. | Non-Defect Finding | Manual review required |
| **`FALSE_POSITIVE`** | Engine flagged code that is idiomatic, safe, intentionally designed, or protected by guards. | **False Positive (FP)** | Prohibited from remediation |
| **`INSUFFICIENT_EVIDENCE`** | Finding lacks sufficient context or AST trace to determine defect status deterministically. | Suppressed Finding | Withheld from remediation |
| **`DUPLICATE_CORRELATED`** | Multiple findings pointing to the same underlying root cause in a single module. | Clustered Finding | Consolidated into single task |
| **`OUT_OF_SCOPE`** | Code located in third-party vendor directories, generated files, or mock test helpers. | Excluded | Excluded from evaluation |

### Core Adjudication Rule:
> **A finding is never classified as a True Positive simply because the diagnostic engine marked it `CONFIRMED`. Independent engineering evidence (e.g. proof of defect, compiler failure, runtime panic, security vulnerability) is mandatory.**

This framework governs all evaluations conducted in Phase 3 through Phase 6.

