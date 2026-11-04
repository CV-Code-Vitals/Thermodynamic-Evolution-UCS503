# Milestone 3.3 — Phase 4: Remediation Candidate Adjudication & Selection Report

**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Branch:** `validation/milestone-3.3-external-validation`  
**Baseline Commit:** `5d41b531f46d10bc972b2d1518a0fb4cf139c09a`  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Phase 4 defines the boundary between diagnosis and remediation. Following the adjudication of diagnostic findings from Phase 3, twelve (12) distinct remediation candidates were cataloged across all five external repositories.

To uphold the core safety invariants of the platform (INV-1 through INV-7), each candidate is classified according to risk tier, blast radius, acceptance criteria, and autonomous execution eligibility.

---

## 2. Risk Classification & Eligibility Matrix

| Candidate ID | Repository | Defect Class | Risk Tier | Blast Radius | Approval Policy | Remediation Trial Eligibility |
|---|---|---|---|---|---|---|
| `REM-CAND-01` | `cli_tooling` | Resource Leak (`defer` in loop) | **LOW** | Function scope | Autonomous | **Eligible for Autonomous Execution** |
| `REM-CAND-02` | `queue_worker` | Ignored Error Return | **LOW** | Goroutine loop | Autonomous | **Eligible for Autonomous Execution** |
| `REM-CAND-03` | `http_router` | Resource Leak (HTTP response body) | **LOW** | Function scope | Autonomous | **Eligible for Autonomous Execution** |
| `REM-CAND-04` | `sql_service` | Resource Leak (SQL Rows) | **LOW** | Function scope | Autonomous | **Eligible for Autonomous Execution** |
| `REM-CAND-05` | `http_router` | Context Timer Leak | **LOW** | Function scope | Autonomous | **Eligible for Autonomous Execution** |
| `REM-CAND-06` | `sql_service` | SQL Injection (`fmt.Sprintf`) | **MEDIUM** | Query contract | Simulated Token | **Simulated Approval Required** |
| `REM-CAND-07` | `cli_tooling` | Command Injection (`exec.Command`) | **MEDIUM** | Subprocess shell | Simulated Token | **Simulated Approval Required** |
| `REM-CAND-08` | `queue_worker` | Socket Leak (`net.Dial`) | **MEDIUM** | Socket lifecycle | Simulated Token | **Simulated Approval Required** |
| `REM-CAND-09` | `cache_engine` | Concurrency Data Race | **HIGH** | Concurrency locks | Human Architect | **BLOCKED: Architect Review Required** |
| `REM-CAND-10` | `cache_engine` | Goroutine Leak (Ticker loop) | **HIGH** | Public API contract | Human Architect | **BLOCKED: Architect Review Required** |
| `REM-CAND-11` | `cli_tooling` | Path Traversal | **HIGH** | File security boundary | Human Architect | **BLOCKED: Architect Review Required** |
| `REM-CAND-12` | `queue_worker` | Timer Leak (`time.NewTimer`) | **LOW** | Loop scope | Autonomous | **Eligible for Autonomous Execution** |

---

## 3. Autonomous Execution Protocol for Phase 5

In strict compliance with **INV-3 (Approval Enforcement)** and **INV-4 (Restricted Execution)**:
1. **Low-Risk Candidates (6 candidates: CAND-01, 02, 03, 04, 05, 12):** Are granted automated execution authorization. Each will be executed in a dedicated, disposable isolated temporary workspace created by `WorkspaceManager`.
2. **Medium-Risk Candidates (3 candidates: CAND-06, 07, 08):** Require an explicit cryptographic approval token (`SIM-APP-*`) to proceed to isolated patch trial.
3. **High-Risk Candidates (3 candidates: CAND-09, 10, 11):** Are quarantined and prevented from autonomous execution. Their plans and diffs are documented for human architectural review.

