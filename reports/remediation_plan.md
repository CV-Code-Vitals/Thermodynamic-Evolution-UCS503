# Remediation Plan & Dependency-Ordered Execution Schedule

**Plan Schema Version:** `1.0`  
**Plan Identifier:** `PLAN-1793697073`  
**Generated At:** 2026-11-03T09:11:13Z  
**Baseline Commit:** `d6fbb8096a08f8928588a1714dec37fcc898e052`  
**Source Diagnostic Report ID:** `DR-1793696463` (Hash: `1988a71ec18d...`)  
**Repository Root:** `E:\SWE`  

## Executive Summary

The Remediation Planning Engine evaluated **486 diagnostic findings**, identified **367 actionable findings**, and generated a dependency-ordered plan comprising **237 distinct tasks**.

### Plan Composition by Execution Class

| Execution Class | Count | Description & Policy |
|---|---|---|
| `AUTOMATION_READY` | 1 | Deterministic modifications safe for automated execution |
| `AUTOMATION_WITH_REVIEW` | 235 | Local refactoring tasks validated via automated test gates |
| `HUMAN_APPROVAL_REQUIRED` | 1 | High-impact security or architectural changes requiring human sign-off |
| `INVESTIGATION_REQUIRED` | 0 | Heuristic or ambiguous findings deferred pending manual review |
| `BLOCKED` | 0 | Tasks blocked by circular dependencies or conflicting locks |

### Conflict Analysis

Identified **466 total conflicts** across planned tasks.

| Conflict ID | Type | Task A | Task B | Resource | Resolution Strategy |
|---|---|---|---|---|---|
| `CONF-SOFT-0001` | `SOFT_CONFLICT` | `RP-0128` | `RP-0203` | `analysis-authorizer/rest_authorizer_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0002` | `SOFT_CONFLICT` | `RP-0017` | `RP-0023` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0003` | `SOFT_CONFLICT` | `RP-0017` | `RP-0028` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0004` | `SOFT_CONFLICT` | `RP-0017` | `RP-0056` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0005` | `SOFT_CONFLICT` | `RP-0017` | `RP-0060` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0006` | `SOFT_CONFLICT` | `RP-0017` | `RP-0085` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0007` | `SOFT_CONFLICT` | `RP-0017` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0008` | `SOFT_CONFLICT` | `RP-0017` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0009` | `SOFT_CONFLICT` | `RP-0017` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0010` | `SOFT_CONFLICT` | `RP-0017` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0011` | `SOFT_CONFLICT` | `RP-0017` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0012` | `SOFT_CONFLICT` | `RP-0017` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0013` | `SOFT_CONFLICT` | `RP-0017` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0014` | `SOFT_CONFLICT` | `RP-0017` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0015` | `SOFT_CONFLICT` | `RP-0017` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0016` | `SOFT_CONFLICT` | `RP-0017` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0017` | `SOFT_CONFLICT` | `RP-0023` | `RP-0028` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0018` | `SOFT_CONFLICT` | `RP-0023` | `RP-0056` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0019` | `SOFT_CONFLICT` | `RP-0023` | `RP-0060` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0020` | `SOFT_CONFLICT` | `RP-0023` | `RP-0085` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0021` | `SOFT_CONFLICT` | `RP-0023` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0022` | `SOFT_CONFLICT` | `RP-0023` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0023` | `SOFT_CONFLICT` | `RP-0023` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0024` | `SOFT_CONFLICT` | `RP-0023` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0025` | `SOFT_CONFLICT` | `RP-0023` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0026` | `SOFT_CONFLICT` | `RP-0023` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0027` | `SOFT_CONFLICT` | `RP-0023` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0028` | `SOFT_CONFLICT` | `RP-0023` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0029` | `SOFT_CONFLICT` | `RP-0023` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0030` | `SOFT_CONFLICT` | `RP-0023` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0031` | `SOFT_CONFLICT` | `RP-0028` | `RP-0056` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0032` | `SOFT_CONFLICT` | `RP-0028` | `RP-0060` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0033` | `SOFT_CONFLICT` | `RP-0028` | `RP-0085` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0034` | `SOFT_CONFLICT` | `RP-0028` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0035` | `SOFT_CONFLICT` | `RP-0028` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0036` | `SOFT_CONFLICT` | `RP-0028` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0037` | `SOFT_CONFLICT` | `RP-0028` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0038` | `SOFT_CONFLICT` | `RP-0028` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0039` | `SOFT_CONFLICT` | `RP-0028` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0040` | `SOFT_CONFLICT` | `RP-0028` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0041` | `SOFT_CONFLICT` | `RP-0028` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0042` | `SOFT_CONFLICT` | `RP-0028` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0043` | `SOFT_CONFLICT` | `RP-0028` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0044` | `SOFT_CONFLICT` | `RP-0056` | `RP-0060` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0045` | `SOFT_CONFLICT` | `RP-0056` | `RP-0085` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0046` | `SOFT_CONFLICT` | `RP-0056` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0047` | `SOFT_CONFLICT` | `RP-0056` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0048` | `SOFT_CONFLICT` | `RP-0056` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0049` | `SOFT_CONFLICT` | `RP-0056` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0050` | `SOFT_CONFLICT` | `RP-0056` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0051` | `SOFT_CONFLICT` | `RP-0056` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0052` | `SOFT_CONFLICT` | `RP-0056` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0053` | `SOFT_CONFLICT` | `RP-0056` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0054` | `SOFT_CONFLICT` | `RP-0056` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0055` | `SOFT_CONFLICT` | `RP-0056` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0056` | `SOFT_CONFLICT` | `RP-0060` | `RP-0085` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0057` | `SOFT_CONFLICT` | `RP-0060` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0058` | `SOFT_CONFLICT` | `RP-0060` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0059` | `SOFT_CONFLICT` | `RP-0060` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0060` | `SOFT_CONFLICT` | `RP-0060` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0061` | `SOFT_CONFLICT` | `RP-0060` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0062` | `SOFT_CONFLICT` | `RP-0060` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0063` | `SOFT_CONFLICT` | `RP-0060` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0064` | `SOFT_CONFLICT` | `RP-0060` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0065` | `SOFT_CONFLICT` | `RP-0060` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0066` | `SOFT_CONFLICT` | `RP-0060` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0067` | `SOFT_CONFLICT` | `RP-0085` | `RP-0089` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0068` | `SOFT_CONFLICT` | `RP-0085` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0069` | `SOFT_CONFLICT` | `RP-0085` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0070` | `SOFT_CONFLICT` | `RP-0085` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0071` | `SOFT_CONFLICT` | `RP-0085` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0072` | `SOFT_CONFLICT` | `RP-0085` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0073` | `SOFT_CONFLICT` | `RP-0085` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0074` | `SOFT_CONFLICT` | `RP-0085` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0075` | `SOFT_CONFLICT` | `RP-0085` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0076` | `SOFT_CONFLICT` | `RP-0085` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0077` | `SOFT_CONFLICT` | `RP-0089` | `RP-0115` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0078` | `SOFT_CONFLICT` | `RP-0089` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0079` | `SOFT_CONFLICT` | `RP-0089` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0080` | `SOFT_CONFLICT` | `RP-0089` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0081` | `SOFT_CONFLICT` | `RP-0089` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0082` | `SOFT_CONFLICT` | `RP-0089` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0083` | `SOFT_CONFLICT` | `RP-0089` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0084` | `SOFT_CONFLICT` | `RP-0089` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0085` | `SOFT_CONFLICT` | `RP-0089` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0086` | `SOFT_CONFLICT` | `RP-0115` | `RP-0120` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0087` | `SOFT_CONFLICT` | `RP-0115` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0088` | `SOFT_CONFLICT` | `RP-0115` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0089` | `SOFT_CONFLICT` | `RP-0115` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0090` | `SOFT_CONFLICT` | `RP-0115` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0091` | `SOFT_CONFLICT` | `RP-0115` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0092` | `SOFT_CONFLICT` | `RP-0115` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0093` | `SOFT_CONFLICT` | `RP-0115` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0094` | `SOFT_CONFLICT` | `RP-0120` | `RP-0124` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0095` | `SOFT_CONFLICT` | `RP-0120` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0096` | `SOFT_CONFLICT` | `RP-0120` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0097` | `SOFT_CONFLICT` | `RP-0120` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0098` | `SOFT_CONFLICT` | `RP-0120` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0099` | `SOFT_CONFLICT` | `RP-0120` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0100` | `SOFT_CONFLICT` | `RP-0120` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0101` | `SOFT_CONFLICT` | `RP-0124` | `RP-0136` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0102` | `SOFT_CONFLICT` | `RP-0124` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0103` | `SOFT_CONFLICT` | `RP-0124` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0104` | `SOFT_CONFLICT` | `RP-0124` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0105` | `SOFT_CONFLICT` | `RP-0124` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0106` | `SOFT_CONFLICT` | `RP-0124` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0107` | `SOFT_CONFLICT` | `RP-0136` | `RP-0137` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0108` | `SOFT_CONFLICT` | `RP-0136` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0109` | `SOFT_CONFLICT` | `RP-0136` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0110` | `SOFT_CONFLICT` | `RP-0136` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0111` | `SOFT_CONFLICT` | `RP-0136` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0112` | `SOFT_CONFLICT` | `RP-0137` | `RP-0138` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0113` | `SOFT_CONFLICT` | `RP-0137` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0114` | `SOFT_CONFLICT` | `RP-0137` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0115` | `SOFT_CONFLICT` | `RP-0137` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0116` | `SOFT_CONFLICT` | `RP-0138` | `RP-0179` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0117` | `SOFT_CONFLICT` | `RP-0138` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0118` | `SOFT_CONFLICT` | `RP-0138` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0119` | `SOFT_CONFLICT` | `RP-0179` | `RP-0202` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0120` | `SOFT_CONFLICT` | `RP-0179` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0121` | `SOFT_CONFLICT` | `RP-0202` | `RP-0209` | `thermodynamic-ast-engine/src/main.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0122` | `SOFT_CONFLICT` | `RP-0018` | `RP-0180` | `admin-portal/tests/graph-model.test.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0123` | `SOFT_CONFLICT` | `RP-0018` | `RP-0191` | `admin-portal/tests/graph-model.test.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0124` | `SOFT_CONFLICT` | `RP-0180` | `RP-0191` | `admin-portal/tests/graph-model.test.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0125` | `SOFT_CONFLICT` | `RP-0008` | `RP-0088` | `admin-portal/src/evolution-api.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0126` | `SOFT_CONFLICT` | `RP-0027` | `RP-0049` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0127` | `SOFT_CONFLICT` | `RP-0027` | `RP-0077` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0128` | `SOFT_CONFLICT` | `RP-0027` | `RP-0145` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0129` | `SOFT_CONFLICT` | `RP-0049` | `RP-0077` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0130` | `SOFT_CONFLICT` | `RP-0049` | `RP-0145` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0131` | `SOFT_CONFLICT` | `RP-0077` | `RP-0145` | `graph-visualizer/routes.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0132` | `SOFT_CONFLICT` | `RP-0030` | `RP-0038` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0133` | `SOFT_CONFLICT` | `RP-0030` | `RP-0070` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0134` | `SOFT_CONFLICT` | `RP-0030` | `RP-0134` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0135` | `SOFT_CONFLICT` | `RP-0038` | `RP-0070` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0136` | `SOFT_CONFLICT` | `RP-0038` | `RP-0134` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0137` | `SOFT_CONFLICT` | `RP-0070` | `RP-0134` | `graph-visualizer/graph.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0138` | `SOFT_CONFLICT` | `RP-0082` | `RP-0098` | `api/router.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0139` | `SOFT_CONFLICT` | `RP-0102` | `RP-0183` | `api/main_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0140` | `SOFT_CONFLICT` | `RP-0102` | `RP-0226` | `api/main_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0141` | `SOFT_CONFLICT` | `RP-0183` | `RP-0226` | `api/main_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0142` | `SOFT_CONFLICT` | `RP-0109` | `RP-0199` | `admin-portal/src/archive-api.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0143` | `SOFT_CONFLICT` | `RP-0004` | `RP-0057` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0144` | `SOFT_CONFLICT` | `RP-0004` | `RP-0074` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0145` | `SOFT_CONFLICT` | `RP-0004` | `RP-0110` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0146` | `SOFT_CONFLICT` | `RP-0004` | `RP-0187` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0147` | `SOFT_CONFLICT` | `RP-0004` | `RP-0230` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0148` | `SOFT_CONFLICT` | `RP-0057` | `RP-0074` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0149` | `SOFT_CONFLICT` | `RP-0057` | `RP-0110` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0150` | `SOFT_CONFLICT` | `RP-0057` | `RP-0187` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0151` | `SOFT_CONFLICT` | `RP-0057` | `RP-0230` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0152` | `SOFT_CONFLICT` | `RP-0074` | `RP-0110` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0153` | `SOFT_CONFLICT` | `RP-0074` | `RP-0187` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0154` | `SOFT_CONFLICT` | `RP-0074` | `RP-0230` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0155` | `SOFT_CONFLICT` | `RP-0110` | `RP-0187` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0156` | `SOFT_CONFLICT` | `RP-0110` | `RP-0230` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0157` | `SOFT_CONFLICT` | `RP-0187` | `RP-0230` | `graph-visualizer/OrbitControls.js` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0158` | `SOFT_CONFLICT` | `RP-0033` | `RP-0157` | `go-backend/agent/broker_evaluator.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0159` | `SOFT_CONFLICT` | `RP-0033` | `RP-0184` | `go-backend/agent/broker_evaluator.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0160` | `SOFT_CONFLICT` | `RP-0157` | `RP-0184` | `go-backend/agent/broker_evaluator.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0161` | `SOFT_CONFLICT` | `RP-0053` | `RP-0122` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0162` | `SOFT_CONFLICT` | `RP-0053` | `RP-0188` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0163` | `SOFT_CONFLICT` | `RP-0053` | `RP-0189` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0164` | `SOFT_CONFLICT` | `RP-0053` | `RP-0201` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0165` | `SOFT_CONFLICT` | `RP-0053` | `RP-0206` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0166` | `SOFT_CONFLICT` | `RP-0122` | `RP-0188` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0167` | `SOFT_CONFLICT` | `RP-0122` | `RP-0189` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0168` | `SOFT_CONFLICT` | `RP-0122` | `RP-0201` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0169` | `SOFT_CONFLICT` | `RP-0122` | `RP-0206` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0170` | `SOFT_CONFLICT` | `RP-0188` | `RP-0189` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0171` | `SOFT_CONFLICT` | `RP-0188` | `RP-0201` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0172` | `SOFT_CONFLICT` | `RP-0188` | `RP-0206` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0173` | `SOFT_CONFLICT` | `RP-0189` | `RP-0201` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0174` | `SOFT_CONFLICT` | `RP-0189` | `RP-0206` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0175` | `SOFT_CONFLICT` | `RP-0201` | `RP-0206` | `infra/analysis/stream-proxy/index.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0176` | `SOFT_CONFLICT` | `RP-0229` | `RP-0232` | `go-backend/agent/agent.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0177` | `SOFT_CONFLICT` | `RP-0229` | `RP-0235` | `go-backend/agent/agent.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0178` | `SOFT_CONFLICT` | `RP-0232` | `RP-0235` | `go-backend/agent/agent.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0179` | `SOFT_CONFLICT` | `RP-0015` | `RP-0025` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0180` | `SOFT_CONFLICT` | `RP-0015` | `RP-0037` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0181` | `SOFT_CONFLICT` | `RP-0015` | `RP-0043` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0182` | `SOFT_CONFLICT` | `RP-0015` | `RP-0069` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0183` | `SOFT_CONFLICT` | `RP-0015` | `RP-0091` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0184` | `SOFT_CONFLICT` | `RP-0015` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0185` | `SOFT_CONFLICT` | `RP-0015` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0186` | `SOFT_CONFLICT` | `RP-0015` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0187` | `SOFT_CONFLICT` | `RP-0015` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0188` | `SOFT_CONFLICT` | `RP-0025` | `RP-0037` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0189` | `SOFT_CONFLICT` | `RP-0025` | `RP-0043` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0190` | `SOFT_CONFLICT` | `RP-0025` | `RP-0069` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0191` | `SOFT_CONFLICT` | `RP-0025` | `RP-0091` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0192` | `SOFT_CONFLICT` | `RP-0025` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0193` | `SOFT_CONFLICT` | `RP-0025` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0194` | `SOFT_CONFLICT` | `RP-0025` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0195` | `SOFT_CONFLICT` | `RP-0025` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0196` | `SOFT_CONFLICT` | `RP-0037` | `RP-0043` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0197` | `SOFT_CONFLICT` | `RP-0037` | `RP-0069` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0198` | `SOFT_CONFLICT` | `RP-0037` | `RP-0091` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0199` | `SOFT_CONFLICT` | `RP-0037` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0200` | `SOFT_CONFLICT` | `RP-0037` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0201` | `SOFT_CONFLICT` | `RP-0037` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0202` | `SOFT_CONFLICT` | `RP-0037` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0203` | `SOFT_CONFLICT` | `RP-0043` | `RP-0069` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0204` | `SOFT_CONFLICT` | `RP-0043` | `RP-0091` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0205` | `SOFT_CONFLICT` | `RP-0043` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0206` | `SOFT_CONFLICT` | `RP-0043` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0207` | `SOFT_CONFLICT` | `RP-0043` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0208` | `SOFT_CONFLICT` | `RP-0043` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0209` | `SOFT_CONFLICT` | `RP-0069` | `RP-0091` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0210` | `SOFT_CONFLICT` | `RP-0069` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0211` | `SOFT_CONFLICT` | `RP-0069` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0212` | `SOFT_CONFLICT` | `RP-0069` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0213` | `SOFT_CONFLICT` | `RP-0069` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0214` | `SOFT_CONFLICT` | `RP-0091` | `RP-0100` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0215` | `SOFT_CONFLICT` | `RP-0091` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0216` | `SOFT_CONFLICT` | `RP-0091` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0217` | `SOFT_CONFLICT` | `RP-0091` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0218` | `SOFT_CONFLICT` | `RP-0100` | `RP-0131` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0219` | `SOFT_CONFLICT` | `RP-0100` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0220` | `SOFT_CONFLICT` | `RP-0100` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0221` | `SOFT_CONFLICT` | `RP-0131` | `RP-0148` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0222` | `SOFT_CONFLICT` | `RP-0131` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0223` | `SOFT_CONFLICT` | `RP-0148` | `RP-0168` | `go-backend/agent/strategy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0224` | `SOFT_CONFLICT` | `RP-0016` | `RP-0106` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0225` | `SOFT_CONFLICT` | `RP-0016` | `RP-0126` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0226` | `SOFT_CONFLICT` | `RP-0016` | `RP-0149` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0227` | `SOFT_CONFLICT` | `RP-0106` | `RP-0126` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0228` | `SOFT_CONFLICT` | `RP-0106` | `RP-0149` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0229` | `SOFT_CONFLICT` | `RP-0126` | `RP-0149` | `admin-portal/src/components/AdminPortal.jsx` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0230` | `SOFT_CONFLICT` | `RP-0020` | `RP-0107` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0231` | `SOFT_CONFLICT` | `RP-0020` | `RP-0108` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0232` | `SOFT_CONFLICT` | `RP-0020` | `RP-0118` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0233` | `SOFT_CONFLICT` | `RP-0020` | `RP-0135` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0234` | `SOFT_CONFLICT` | `RP-0020` | `RP-0163` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0235` | `SOFT_CONFLICT` | `RP-0020` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0236` | `SOFT_CONFLICT` | `RP-0020` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0237` | `SOFT_CONFLICT` | `RP-0020` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0238` | `SOFT_CONFLICT` | `RP-0020` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0239` | `SOFT_CONFLICT` | `RP-0020` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0240` | `SOFT_CONFLICT` | `RP-0107` | `RP-0108` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0241` | `SOFT_CONFLICT` | `RP-0107` | `RP-0118` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0242` | `SOFT_CONFLICT` | `RP-0107` | `RP-0135` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0243` | `SOFT_CONFLICT` | `RP-0107` | `RP-0163` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0244` | `SOFT_CONFLICT` | `RP-0107` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0245` | `SOFT_CONFLICT` | `RP-0107` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0246` | `SOFT_CONFLICT` | `RP-0107` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0247` | `SOFT_CONFLICT` | `RP-0107` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0248` | `SOFT_CONFLICT` | `RP-0107` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0249` | `SOFT_CONFLICT` | `RP-0108` | `RP-0118` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0250` | `SOFT_CONFLICT` | `RP-0108` | `RP-0135` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0251` | `SOFT_CONFLICT` | `RP-0108` | `RP-0163` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0252` | `SOFT_CONFLICT` | `RP-0108` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0253` | `SOFT_CONFLICT` | `RP-0108` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0254` | `SOFT_CONFLICT` | `RP-0108` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0255` | `SOFT_CONFLICT` | `RP-0108` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0256` | `SOFT_CONFLICT` | `RP-0108` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0257` | `SOFT_CONFLICT` | `RP-0118` | `RP-0135` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0258` | `SOFT_CONFLICT` | `RP-0118` | `RP-0163` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0259` | `SOFT_CONFLICT` | `RP-0118` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0260` | `SOFT_CONFLICT` | `RP-0118` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0261` | `SOFT_CONFLICT` | `RP-0118` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0262` | `SOFT_CONFLICT` | `RP-0118` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0263` | `SOFT_CONFLICT` | `RP-0118` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0264` | `SOFT_CONFLICT` | `RP-0135` | `RP-0163` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0265` | `SOFT_CONFLICT` | `RP-0135` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0266` | `SOFT_CONFLICT` | `RP-0135` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0267` | `SOFT_CONFLICT` | `RP-0135` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0268` | `SOFT_CONFLICT` | `RP-0135` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0269` | `SOFT_CONFLICT` | `RP-0135` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0270` | `SOFT_CONFLICT` | `RP-0163` | `RP-0177` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0271` | `SOFT_CONFLICT` | `RP-0163` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0272` | `SOFT_CONFLICT` | `RP-0163` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0273` | `SOFT_CONFLICT` | `RP-0163` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0274` | `SOFT_CONFLICT` | `RP-0163` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0275` | `SOFT_CONFLICT` | `RP-0177` | `RP-0178` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0276` | `SOFT_CONFLICT` | `RP-0177` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0277` | `SOFT_CONFLICT` | `RP-0177` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0278` | `SOFT_CONFLICT` | `RP-0177` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0279` | `SOFT_CONFLICT` | `RP-0178` | `RP-0195` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0280` | `SOFT_CONFLICT` | `RP-0178` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0281` | `SOFT_CONFLICT` | `RP-0178` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0282` | `SOFT_CONFLICT` | `RP-0195` | `RP-0196` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0283` | `SOFT_CONFLICT` | `RP-0195` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0284` | `SOFT_CONFLICT` | `RP-0196` | `RP-0208` | `scripts/analysis-cleanup.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0285` | `SOFT_CONFLICT` | `RP-0073` | `RP-0113` | `api/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0286` | `SOFT_CONFLICT` | `RP-0073` | `RP-0155` | `api/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0287` | `SOFT_CONFLICT` | `RP-0113` | `RP-0155` | `api/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0288` | `SOFT_CONFLICT` | `RP-0013` | `RP-0086` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0289` | `SOFT_CONFLICT` | `RP-0013` | `RP-0099` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0290` | `SOFT_CONFLICT` | `RP-0013` | `RP-0166` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0291` | `SOFT_CONFLICT` | `RP-0013` | `RP-0227` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0292` | `SOFT_CONFLICT` | `RP-0086` | `RP-0099` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0293` | `SOFT_CONFLICT` | `RP-0086` | `RP-0166` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0294` | `SOFT_CONFLICT` | `RP-0086` | `RP-0227` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0295` | `SOFT_CONFLICT` | `RP-0099` | `RP-0166` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0296` | `SOFT_CONFLICT` | `RP-0099` | `RP-0227` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0297` | `SOFT_CONFLICT` | `RP-0166` | `RP-0227` | `api/accounts_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0298` | `SOFT_CONFLICT` | `RP-0042` | `RP-0169` | `go-backend/broker_workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0299` | `SOFT_CONFLICT` | `RP-0045` | `RP-0078` | `go-backend/agent/broker_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0300` | `SOFT_CONFLICT` | `RP-0045` | `RP-0167` | `go-backend/agent/broker_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0301` | `SOFT_CONFLICT` | `RP-0078` | `RP-0167` | `go-backend/agent/broker_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0302` | `SOFT_CONFLICT` | `RP-0048` | `RP-0154` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0303` | `SOFT_CONFLICT` | `RP-0048` | `RP-0200` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0304` | `SOFT_CONFLICT` | `RP-0048` | `RP-0216` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0305` | `SOFT_CONFLICT` | `RP-0048` | `RP-0218` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0306` | `SOFT_CONFLICT` | `RP-0048` | `RP-0223` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0307` | `SOFT_CONFLICT` | `RP-0048` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0308` | `SOFT_CONFLICT` | `RP-0154` | `RP-0200` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0309` | `SOFT_CONFLICT` | `RP-0154` | `RP-0216` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0310` | `SOFT_CONFLICT` | `RP-0154` | `RP-0218` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0311` | `SOFT_CONFLICT` | `RP-0154` | `RP-0223` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0312` | `SOFT_CONFLICT` | `RP-0154` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0313` | `SOFT_CONFLICT` | `RP-0200` | `RP-0216` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0314` | `SOFT_CONFLICT` | `RP-0200` | `RP-0218` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0315` | `SOFT_CONFLICT` | `RP-0200` | `RP-0223` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0316` | `SOFT_CONFLICT` | `RP-0200` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0317` | `SOFT_CONFLICT` | `RP-0216` | `RP-0218` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0318` | `SOFT_CONFLICT` | `RP-0216` | `RP-0223` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0319` | `SOFT_CONFLICT` | `RP-0216` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0320` | `SOFT_CONFLICT` | `RP-0218` | `RP-0223` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0321` | `SOFT_CONFLICT` | `RP-0218` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0322` | `SOFT_CONFLICT` | `RP-0223` | `RP-0228` | `api/aws_store_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0323` | `SOFT_CONFLICT` | `RP-0012` | `RP-0121` | `api/analysis_proxy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0324` | `SOFT_CONFLICT` | `RP-0010` | `RP-0185` | `go-backend/agent/evaluator.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0325` | `SOFT_CONFLICT` | `RP-0034` | `RP-0198` | `thermodynamic-ast-engine/src/languages.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0326` | `SOFT_CONFLICT` | `RP-0006` | `RP-0035` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0327` | `SOFT_CONFLICT` | `RP-0006` | `RP-0075` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0328` | `SOFT_CONFLICT` | `RP-0006` | `RP-0127` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0329` | `SOFT_CONFLICT` | `RP-0006` | `RP-0139` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0330` | `SOFT_CONFLICT` | `RP-0035` | `RP-0075` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0331` | `SOFT_CONFLICT` | `RP-0035` | `RP-0127` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0332` | `SOFT_CONFLICT` | `RP-0035` | `RP-0139` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0333` | `SOFT_CONFLICT` | `RP-0075` | `RP-0127` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0334` | `SOFT_CONFLICT` | `RP-0075` | `RP-0139` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0335` | `SOFT_CONFLICT` | `RP-0127` | `RP-0139` | `thermodynamic-ast-engine/src/relationships.rs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0336` | `SOFT_CONFLICT` | `RP-0026` | `RP-0129` | `go-backend/agent/energy_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0337` | `SOFT_CONFLICT` | `RP-0026` | `RP-0215` | `go-backend/agent/energy_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0338` | `SOFT_CONFLICT` | `RP-0129` | `RP-0215` | `go-backend/agent/energy_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0339` | `SOFT_CONFLICT` | `RP-0140` | `RP-0141` | `thermodynamic-ast-engine/test_samples/crawler.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0340` | `SOFT_CONFLICT` | `RP-0040` | `RP-0044` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0341` | `SOFT_CONFLICT` | `RP-0040` | `RP-0050` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0342` | `SOFT_CONFLICT` | `RP-0040` | `RP-0234` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0343` | `SOFT_CONFLICT` | `RP-0044` | `RP-0050` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0344` | `SOFT_CONFLICT` | `RP-0044` | `RP-0234` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0345` | `SOFT_CONFLICT` | `RP-0050` | `RP-0234` | `go-backend/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0346` | `SOFT_CONFLICT` | `RP-0064` | `RP-0080` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0347` | `SOFT_CONFLICT` | `RP-0064` | `RP-0104` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0348` | `SOFT_CONFLICT` | `RP-0064` | `RP-0152` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0349` | `SOFT_CONFLICT` | `RP-0064` | `RP-0182` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0350` | `SOFT_CONFLICT` | `RP-0064` | `RP-0210` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0351` | `SOFT_CONFLICT` | `RP-0080` | `RP-0104` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0352` | `SOFT_CONFLICT` | `RP-0080` | `RP-0152` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0353` | `SOFT_CONFLICT` | `RP-0080` | `RP-0182` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0354` | `SOFT_CONFLICT` | `RP-0080` | `RP-0210` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0355` | `SOFT_CONFLICT` | `RP-0104` | `RP-0152` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0356` | `SOFT_CONFLICT` | `RP-0104` | `RP-0182` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0357` | `SOFT_CONFLICT` | `RP-0104` | `RP-0210` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0358` | `SOFT_CONFLICT` | `RP-0152` | `RP-0182` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0359` | `SOFT_CONFLICT` | `RP-0152` | `RP-0210` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0360` | `SOFT_CONFLICT` | `RP-0182` | `RP-0210` | `api/archive.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0361` | `SOFT_CONFLICT` | `RP-0024` | `RP-0051` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0362` | `SOFT_CONFLICT` | `RP-0024` | `RP-0068` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0363` | `SOFT_CONFLICT` | `RP-0024` | `RP-0173` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0364` | `SOFT_CONFLICT` | `RP-0051` | `RP-0068` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0365` | `SOFT_CONFLICT` | `RP-0051` | `RP-0173` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0366` | `SOFT_CONFLICT` | `RP-0068` | `RP-0173` | `api/local_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0367` | `SOFT_CONFLICT` | `RP-0172` | `RP-0192` | `api/aws_store.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0368` | `SOFT_CONFLICT` | `RP-0065` | `RP-0162` | `graph-visualizer/model.mjs` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0369` | `SOFT_CONFLICT` | `RP-0090` | `RP-0096` | `go-backend/evolve.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0370` | `SOFT_CONFLICT` | `RP-0032` | `RP-0063` | `api/lambda_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0371` | `SOFT_CONFLICT` | `RP-0029` | `RP-0142` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0372` | `SOFT_CONFLICT` | `RP-0029` | `RP-0146` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0373` | `SOFT_CONFLICT` | `RP-0029` | `RP-0147` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0374` | `SOFT_CONFLICT` | `RP-0142` | `RP-0146` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0375` | `SOFT_CONFLICT` | `RP-0142` | `RP-0147` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0376` | `SOFT_CONFLICT` | `RP-0146` | `RP-0147` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0377` | `SOFT_CONFLICT` | `RP-0087` | `RP-0158` | `go-backend/agent/strategy_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0378` | `SOFT_CONFLICT` | `RP-0132` | `RP-0133` | `go-backend/deliverables.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0379` | `SOFT_CONFLICT` | `RP-0132` | `RP-0224` | `go-backend/deliverables.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0380` | `SOFT_CONFLICT` | `RP-0133` | `RP-0224` | `go-backend/deliverables.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0381` | `SOFT_CONFLICT` | `RP-0007` | `RP-0019` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0382` | `SOFT_CONFLICT` | `RP-0007` | `RP-0160` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0383` | `SOFT_CONFLICT` | `RP-0007` | `RP-0170` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0384` | `SOFT_CONFLICT` | `RP-0007` | `RP-0171` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0385` | `SOFT_CONFLICT` | `RP-0007` | `RP-0186` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0386` | `SOFT_CONFLICT` | `RP-0007` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0387` | `SOFT_CONFLICT` | `RP-0007` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0388` | `SOFT_CONFLICT` | `RP-0007` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0389` | `SOFT_CONFLICT` | `RP-0019` | `RP-0160` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0390` | `SOFT_CONFLICT` | `RP-0019` | `RP-0170` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0391` | `SOFT_CONFLICT` | `RP-0019` | `RP-0171` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0392` | `SOFT_CONFLICT` | `RP-0019` | `RP-0186` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0393` | `SOFT_CONFLICT` | `RP-0019` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0394` | `SOFT_CONFLICT` | `RP-0019` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0395` | `SOFT_CONFLICT` | `RP-0019` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0396` | `SOFT_CONFLICT` | `RP-0160` | `RP-0170` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0397` | `SOFT_CONFLICT` | `RP-0160` | `RP-0171` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0398` | `SOFT_CONFLICT` | `RP-0160` | `RP-0186` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0399` | `SOFT_CONFLICT` | `RP-0160` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0400` | `SOFT_CONFLICT` | `RP-0160` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0401` | `SOFT_CONFLICT` | `RP-0160` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0402` | `SOFT_CONFLICT` | `RP-0170` | `RP-0171` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0403` | `SOFT_CONFLICT` | `RP-0170` | `RP-0186` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0404` | `SOFT_CONFLICT` | `RP-0170` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0405` | `SOFT_CONFLICT` | `RP-0170` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0406` | `SOFT_CONFLICT` | `RP-0170` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0407` | `SOFT_CONFLICT` | `RP-0171` | `RP-0186` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0408` | `SOFT_CONFLICT` | `RP-0171` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0409` | `SOFT_CONFLICT` | `RP-0171` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0410` | `SOFT_CONFLICT` | `RP-0171` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0411` | `SOFT_CONFLICT` | `RP-0186` | `RP-0194` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0412` | `SOFT_CONFLICT` | `RP-0186` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0413` | `SOFT_CONFLICT` | `RP-0186` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0414` | `SOFT_CONFLICT` | `RP-0194` | `RP-0212` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0415` | `SOFT_CONFLICT` | `RP-0194` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0416` | `SOFT_CONFLICT` | `RP-0212` | `RP-0214` | `go-backend/sandbox_lifecycle.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0417` | `SOFT_CONFLICT` | `RP-0031` | `RP-0143` | `go-backend/agent/energy.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0418` | `SOFT_CONFLICT` | `RP-0062` | `RP-0174` | `go-backend/agent/repo_intel/diagnostic/correlation.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0419` | `SOFT_CONFLICT` | `RP-0021` | `RP-0084` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0420` | `SOFT_CONFLICT` | `RP-0021` | `RP-0144` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0421` | `SOFT_CONFLICT` | `RP-0021` | `RP-0221` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0422` | `SOFT_CONFLICT` | `RP-0084` | `RP-0144` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0423` | `SOFT_CONFLICT` | `RP-0084` | `RP-0221` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0424` | `SOFT_CONFLICT` | `RP-0144` | `RP-0221` | `go-backend/agent/repo_intel/manifests.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0425` | `SOFT_CONFLICT` | `RP-0059` | `RP-0156` | `api/refresh_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0426` | `SOFT_CONFLICT` | `RP-0059` | `RP-0204` | `api/refresh_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0427` | `SOFT_CONFLICT` | `RP-0156` | `RP-0204` | `api/refresh_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0428` | `SOFT_CONFLICT` | `RP-0231` | `RP-0233` | `go-backend/agent/annealer_benchmark_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0429` | `SOFT_CONFLICT` | `RP-0003` | `RP-0011` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0430` | `SOFT_CONFLICT` | `RP-0003` | `RP-0076` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0431` | `SOFT_CONFLICT` | `RP-0003` | `RP-0112` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0432` | `SOFT_CONFLICT` | `RP-0003` | `RP-0114` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0433` | `SOFT_CONFLICT` | `RP-0003` | `RP-0151` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0434` | `SOFT_CONFLICT` | `RP-0011` | `RP-0076` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0435` | `SOFT_CONFLICT` | `RP-0011` | `RP-0112` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0436` | `SOFT_CONFLICT` | `RP-0011` | `RP-0114` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0437` | `SOFT_CONFLICT` | `RP-0011` | `RP-0151` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0438` | `SOFT_CONFLICT` | `RP-0076` | `RP-0112` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0439` | `SOFT_CONFLICT` | `RP-0076` | `RP-0114` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0440` | `SOFT_CONFLICT` | `RP-0076` | `RP-0151` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0441` | `SOFT_CONFLICT` | `RP-0112` | `RP-0114` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0442` | `SOFT_CONFLICT` | `RP-0112` | `RP-0151` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0443` | `SOFT_CONFLICT` | `RP-0114` | `RP-0151` | `api/accounts.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0444` | `SOFT_CONFLICT` | `RP-0054` | `RP-0111` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0445` | `SOFT_CONFLICT` | `RP-0054` | `RP-0116` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0446` | `SOFT_CONFLICT` | `RP-0054` | `RP-0119` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0447` | `SOFT_CONFLICT` | `RP-0054` | `RP-0164` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0448` | `SOFT_CONFLICT` | `RP-0111` | `RP-0116` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0449` | `SOFT_CONFLICT` | `RP-0111` | `RP-0119` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0450` | `SOFT_CONFLICT` | `RP-0111` | `RP-0164` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0451` | `SOFT_CONFLICT` | `RP-0116` | `RP-0119` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0452` | `SOFT_CONFLICT` | `RP-0116` | `RP-0164` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0453` | `SOFT_CONFLICT` | `RP-0119` | `RP-0164` | `source-stager/main.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0454` | `SOFT_CONFLICT` | `RP-0022` | `RP-0052` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0455` | `SOFT_CONFLICT` | `RP-0022` | `RP-0117` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0456` | `SOFT_CONFLICT` | `RP-0022` | `RP-0159` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0457` | `SOFT_CONFLICT` | `RP-0052` | `RP-0117` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0458` | `SOFT_CONFLICT` | `RP-0052` | `RP-0159` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0459` | `SOFT_CONFLICT` | `RP-0117` | `RP-0159` | `go-backend/agent/workspace.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0460` | `SOFT_CONFLICT` | `RP-0066` | `RP-0150` | `analysis-authorizer/main_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0461` | `SOFT_CONFLICT` | `RP-0103` | `RP-0130` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0462` | `SOFT_CONFLICT` | `RP-0103` | `RP-0175` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0463` | `SOFT_CONFLICT` | `RP-0103` | `RP-0225` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0464` | `SOFT_CONFLICT` | `RP-0130` | `RP-0175` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0465` | `SOFT_CONFLICT` | `RP-0130` | `RP-0225` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |
| `CONF-SOFT-0466` | `SOFT_CONFLICT` | `RP-0175` | `RP-0225` | `go-backend/agent/repo_intel/repo_intel_test.go` | Safe to execute sequentially; verify line offset stability |

## Dependency-Ordered Execution Schedule

The following sequence represents the deterministic topological execution order:

| Step | Task ID | Priority | Category | Strategy | Target Files | Approval Required |
|---|---|---|---|---|---|---|
| 1 | `RP-0103` | `URGENT` | `SECURITY` | `HUMAN_GUIDED` | `go-backend/agent/repo_intel/repo_intel_test.go` | **YES** |
| 2 | `RP-0126` | `HIGH` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/components/AdminPortal.jsx` | No |
| 3 | `RP-0127` | `HIGH` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/relationships.rs` | No |
| 4 | `RP-0001` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/evolution-measurements.js` | No |
| 5 | `RP-0002` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/evaluator.go` | No |
| 6 | `RP-0003` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 7 | `RP-0004` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 8 | `RP-0005` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/aggregator.go` | No |
| 9 | `RP-0006` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/relationships.rs` | No |
| 10 | `RP-0007` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 11 | `RP-0008` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/evolution-api.js` | No |
| 12 | `RP-0009` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/analysis_proxy_test.go` | No |
| 13 | `RP-0010` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/evaluator.go` | No |
| 14 | `RP-0011` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 15 | `RP-0012` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/analysis_proxy.go` | No |
| 16 | `RP-0013` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts_test.go` | No |
| 17 | `RP-0014` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/reporter.go` | No |
| 18 | `RP-0015` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 19 | `RP-0016` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/components/AdminPortal.jsx` | No |
| 20 | `RP-0017` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 21 | `RP-0018` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/tests/graph-model.test.js` | No |
| 22 | `RP-0019` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 23 | `RP-0020` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 24 | `RP-0021` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/manifests.go` | No |
| 25 | `RP-0022` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/workspace.go` | No |
| 26 | `RP-0023` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 27 | `RP-0024` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/local_store.go` | No |
| 28 | `RP-0025` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 29 | `RP-0026` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/energy_test.go` | No |
| 30 | `RP-0027` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/routes.mjs` | No |
| 31 | `RP-0028` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 32 | `RP-0029` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | No |
| 33 | `RP-0030` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/graph.mjs` | No |
| 34 | `RP-0031` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/energy.go` | No |
| 35 | `RP-0032` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/lambda_test.go` | No |
| 36 | `RP-0033` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_evaluator.go` | No |
| 37 | `RP-0034` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/languages.rs` | No |
| 38 | `RP-0035` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/relationships.rs` | No |
| 39 | `RP-0036` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/vite.config.js` | No |
| 40 | `RP-0037` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 41 | `RP-0038` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/graph.mjs` | No |
| 42 | `RP-0039` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/orchestrator.go` | No |
| 43 | `RP-0040` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/main.go` | No |
| 44 | `RP-0041` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/account_store.go` | No |
| 45 | `RP-0042` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/broker_workspace.go` | No |
| 46 | `RP-0043` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 47 | `RP-0044` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/main.go` | No |
| 48 | `RP-0045` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_test.go` | No |
| 49 | `RP-0046` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/quality.go` | No |
| 50 | `RP-0047` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/snapshot_scanner.go` | No |
| 51 | `RP-0048` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 52 | `RP-0049` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/routes.mjs` | No |
| 53 | `RP-0050` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/main.go` | No |
| 54 | `RP-0051` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/local_store.go` | No |
| 55 | `RP-0052` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/workspace.go` | No |
| 56 | `RP-0053` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 57 | `RP-0054` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `source-stager/main.go` | No |
| 58 | `RP-0055` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/local_store_test.go` | No |
| 59 | `RP-0056` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 60 | `RP-0057` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 61 | `RP-0058` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/scanner.go` | No |
| 62 | `RP-0059` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/refresh_test.go` | No |
| 63 | `RP-0060` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 64 | `RP-0061` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/discovery.go` | No |
| 65 | `RP-0062` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/correlation.go` | No |
| 66 | `RP-0063` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/lambda_test.go` | No |
| 67 | `RP-0064` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 68 | `RP-0065` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/model.mjs` | No |
| 69 | `RP-0066` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `analysis-authorizer/main_test.go` | No |
| 70 | `RP-0067` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `samples/benchmark/benchmark_target.go` | No |
| 71 | `RP-0068` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/local_store.go` | No |
| 72 | `RP-0069` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 73 | `RP-0070` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/graph.mjs` | No |
| 74 | `RP-0071` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/severity.go` | No |
| 75 | `RP-0072` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/workspace_test.go` | No |
| 76 | `RP-0073` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main.go` | No |
| 77 | `RP-0074` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 78 | `RP-0075` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/relationships.rs` | No |
| 79 | `RP-0076` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 80 | `RP-0077` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/routes.mjs` | No |
| 81 | `RP-0078` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_test.go` | No |
| 82 | `RP-0079` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/structural.go` | No |
| 83 | `RP-0080` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 84 | `RP-0081` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/annealer.go` | No |
| 85 | `RP-0082` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/router.go` | No |
| 86 | `RP-0083` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/architecture.go` | No |
| 87 | `RP-0084` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/manifests.go` | No |
| 88 | `RP-0085` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 89 | `RP-0086` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts_test.go` | No |
| 90 | `RP-0087` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy_test.go` | No |
| 91 | `RP-0088` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/evolution-api.js` | No |
| 92 | `RP-0089` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 93 | `RP-0090` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/evolve.go` | No |
| 94 | `RP-0091` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 95 | `RP-0092` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/snapshot_portal_test.go` | No |
| 96 | `RP-0093` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/workspace_test.go` | No |
| 97 | `RP-0094` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/snapshot_store.go` | No |
| 98 | `RP-0095` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/reliability_security.go` | No |
| 99 | `RP-0096` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/evolve.go` | No |
| 100 | `RP-0097` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/validator.go` | No |
| 101 | `RP-0098` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/router.go` | No |
| 102 | `RP-0099` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts_test.go` | No |
| 103 | `RP-0100` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 104 | `RP-0101` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/workspace.go` | No |
| 105 | `RP-0102` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main_test.go` | No |
| 106 | `RP-0104` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 107 | `RP-0105` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/readiness.go` | No |
| 108 | `RP-0106` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/components/AdminPortal.jsx` | No |
| 109 | `RP-0107` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 110 | `RP-0108` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 111 | `RP-0109` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/archive-api.js` | No |
| 112 | `RP-0110` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 113 | `RP-0111` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `source-stager/main.go` | No |
| 114 | `RP-0112` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 115 | `RP-0113` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main.go` | No |
| 116 | `RP-0114` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 117 | `RP-0115` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 118 | `RP-0116` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `source-stager/main.go` | No |
| 119 | `RP-0117` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/workspace.go` | No |
| 120 | `RP-0118` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 121 | `RP-0119` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `source-stager/main.go` | No |
| 122 | `RP-0120` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 123 | `RP-0121` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/analysis_proxy.go` | No |
| 124 | `RP-0122` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 125 | `RP-0123` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `tests/analysis-workflow.test.mjs` | No |
| 126 | `RP-0124` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 127 | `RP-0128` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `analysis-authorizer/rest_authorizer_test.go` | No |
| 128 | `RP-0129` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/energy_test.go` | No |
| 129 | `RP-0130` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/repo_intel_test.go` | No |
| 130 | `RP-0131` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 131 | `RP-0132` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/deliverables.go` | No |
| 132 | `RP-0133` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/deliverables.go` | No |
| 133 | `RP-0134` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/graph.mjs` | No |
| 134 | `RP-0135` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 135 | `RP-0136` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 136 | `RP-0137` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 137 | `RP-0138` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 138 | `RP-0139` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/relationships.rs` | No |
| 139 | `RP-0140` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/crawler.go` | No |
| 140 | `RP-0141` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/crawler.go` | No |
| 141 | `RP-0142` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | No |
| 142 | `RP-0143` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/energy.go` | No |
| 143 | `RP-0144` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/manifests.go` | No |
| 144 | `RP-0145` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/routes.mjs` | No |
| 145 | `RP-0146` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | No |
| 146 | `RP-0147` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/test_samples/data_pipeline.py` | No |
| 147 | `RP-0148` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 148 | `RP-0149` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/components/AdminPortal.jsx` | No |
| 149 | `RP-0150` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `analysis-authorizer/main_test.go` | No |
| 150 | `RP-0151` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts.go` | No |
| 151 | `RP-0152` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 152 | `RP-0153` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/auth.go` | No |
| 153 | `RP-0154` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 154 | `RP-0155` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main.go` | No |
| 155 | `RP-0156` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/refresh_test.go` | No |
| 156 | `RP-0157` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_evaluator.go` | No |
| 157 | `RP-0158` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy_test.go` | No |
| 158 | `RP-0159` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/workspace.go` | No |
| 159 | `RP-0160` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 160 | `RP-0161` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/account.mjs` | No |
| 161 | `RP-0162` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/model.mjs` | No |
| 162 | `RP-0163` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 163 | `RP-0164` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `source-stager/main.go` | No |
| 164 | `RP-0165` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/archive-download.js` | No |
| 165 | `RP-0166` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts_test.go` | No |
| 166 | `RP-0167` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_test.go` | No |
| 167 | `RP-0168` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/strategy.go` | No |
| 168 | `RP-0169` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/broker_workspace.go` | No |
| 169 | `RP-0170` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 170 | `RP-0171` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 171 | `RP-0172` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store.go` | No |
| 172 | `RP-0173` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/local_store.go` | No |
| 173 | `RP-0174` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/correlation.go` | No |
| 174 | `RP-0175` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/repo_intel_test.go` | No |
| 175 | `RP-0176` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/broker.go` | No |
| 176 | `RP-0177` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 177 | `RP-0178` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 178 | `RP-0179` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 179 | `RP-0180` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/tests/graph-model.test.js` | No |
| 180 | `RP-0181` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `analysis-authorizer/main.go` | No |
| 181 | `RP-0182` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 182 | `RP-0183` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main_test.go` | No |
| 183 | `RP-0184` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/broker_evaluator.go` | No |
| 184 | `RP-0185` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/evaluator.go` | No |
| 185 | `RP-0186` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 186 | `RP-0187` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 187 | `RP-0188` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 188 | `RP-0189` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 189 | `RP-0190` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `tests/sandbox-host-validation.test.mjs` | No |
| 190 | `RP-0191` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/tests/graph-model.test.js` | No |
| 191 | `RP-0192` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store.go` | No |
| 192 | `RP-0193` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/lambda.go` | No |
| 193 | `RP-0194` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 194 | `RP-0195` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 195 | `RP-0196` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 196 | `RP-0197` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `serve.js` | No |
| 197 | `RP-0198` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/languages.rs` | No |
| 198 | `RP-0199` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/archive-api.js` | No |
| 199 | `RP-0200` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 200 | `RP-0201` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 201 | `RP-0202` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 202 | `RP-0203` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `analysis-authorizer/rest_authorizer_test.go` | No |
| 203 | `RP-0204` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/refresh_test.go` | No |
| 204 | `RP-0205` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/calibration.go` | No |
| 205 | `RP-0206` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `infra/analysis/stream-proxy/index.mjs` | No |
| 206 | `RP-0207` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/validate-sandbox-host.mjs` | No |
| 207 | `RP-0208` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `scripts/analysis-cleanup.mjs` | No |
| 208 | `RP-0209` | `MEDIUM` | `STRUCTURAL` | `AST_TRANSFORMATION` | `thermodynamic-ast-engine/src/main.rs` | No |
| 209 | `RP-0125` | `MEDIUM` | `DEPENDENCY` | `RULE_BASED` | `samples/benchmark/go.mod` | No |
| 210 | `RP-0210` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/archive.go` | No |
| 211 | `RP-0211` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/evaluator_test.go` | No |
| 212 | `RP-0212` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 213 | `RP-0213` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/diagnostic/diagnostic_test.go` | No |
| 214 | `RP-0214` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/sandbox_lifecycle.go` | No |
| 215 | `RP-0215` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/energy_test.go` | No |
| 216 | `RP-0216` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 217 | `RP-0217` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/agent_test.go` | No |
| 218 | `RP-0218` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 219 | `RP-0219` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/cmd/diagnostic/main.go` | No |
| 220 | `RP-0220` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/main_test.go` | No |
| 221 | `RP-0221` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/manifests.go` | No |
| 222 | `RP-0222` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/scan_handler_test.go` | No |
| 223 | `RP-0223` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 224 | `RP-0224` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/deliverables.go` | No |
| 225 | `RP-0225` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/repo_intel/repo_intel_test.go` | No |
| 226 | `RP-0226` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/main_test.go` | No |
| 227 | `RP-0227` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/accounts_test.go` | No |
| 228 | `RP-0228` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `api/aws_store_test.go` | No |
| 229 | `RP-0229` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/agent.go` | No |
| 230 | `RP-0230` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `graph-visualizer/OrbitControls.js` | No |
| 231 | `RP-0231` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/annealer_benchmark_test.go` | No |
| 232 | `RP-0232` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/agent.go` | No |
| 233 | `RP-0233` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/annealer_benchmark_test.go` | No |
| 234 | `RP-0234` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/main.go` | No |
| 235 | `RP-0235` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `go-backend/agent/agent.go` | No |
| 236 | `RP-0236` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `admin-portal/src/components/EvolutionLive.jsx` | No |
| 237 | `RP-0237` | `LOW` | `STRUCTURAL` | `AST_TRANSFORMATION` | `samples/benchmark/benchmark_target_test.go` | No |

## Detailed Remediation Tasks

### 1. [URGENT] Potential hardcoded secret (AWS Access Key ID) (`RP-0103`)

- **Category:** `SECURITY` | **Execution Strategy:** `HUMAN_GUIDED`
- **Execution Class:** `HUMAN_APPROVAL_REQUIRED` | **Priority Score:** `81.0`
- **Target Files:** `go-backend/agent/repo_intel/repo_intel_test.go`
- **Source Finding IDs:** `RI-SEC-SECRET-GO_BACKEND_AGENT_REPO_INTEL_REPO_INTEL_TEST_GO-L65`
- **Regression Risk:** `LOW`
- > [!CAUTION]
> **Human Approval Required:** Credential revocation, cloud secret rotation, and IAM policy provisioning require human authorization

File go-backend/agent/repo_intel/repo_intel_test.go appears to contain a hardcoded credential or token: AWS Access Key ID

**Acceptance Criteria:**
- [ ] Hardcoded credential is completely removed from source code and git tracking
- [ ] Credential is verified revoked and rotated with cloud provider
- [ ] Application successfully loads replacement credential via environment variable or secret manager

**Validation Commands:**
```bash
git grep -n -i 'AKIA' || true
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 2. [HIGH] AST parse error encountered (`RP-0126`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `59.0`
- **Target Files:** `admin-portal/src/components/AdminPortal.jsx`
- **Source Finding IDs:** `RI-STR-PARSE-ERROR-ADMIN_PORTAL_SRC_COMPONENTS_ADMINPORTAL_JSX`
- **Regression Risk:** `LOW`

Source file admin-portal/src/components/AdminPortal.jsx contains syntax or grammar parse errors preventing complete AST analysis.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src/components
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 3. [HIGH] AST parse error encountered (`RP-0127`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `59.0`
- **Target Files:** `thermodynamic-ast-engine/src/relationships.rs`
- **Source Finding IDs:** `RI-STR-PARSE-ERROR-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS`
- **Regression Risk:** `LOW`

Source file thermodynamic-ast-engine/src/relationships.rs contains syntax or grammar parse errors preventing complete AST analysis.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 4. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0001`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/evolution-measurements.js`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ADMIN_PORTAL_SRC_EVOLUTION_MEASUREMENTS_JS-evolutionMeasurements-L1`, `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_EVOLUTION_MEASUREMENTS_JS-evolutionMeasurements-L1`
- **Regression Risk:** `LOW`

Function "evolutionMeasurements" in admin-portal/src/evolution-measurements.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 5. [MEDIUM] Excessive cyclomatic complexity (M=32) (+ 1 related findings) (`RP-0002`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/evaluator.go`
- **Target Symbols:** RunControlledQualityEvaluation
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_EVALUATOR_GO-RunControlledQualityEvaluation-L45`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_EVALUATOR_GO-RunControlledQualityEvaluation-L45`
- **Regression Risk:** `MEDIUM`

Function "RunControlledQualityEvaluation" has cyclomatic complexity of 32, requiring at least 32 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 6. [MEDIUM] Excessive cyclomatic complexity (M=40) (+ 1 related findings) (`RP-0003`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Target Symbols:** accountRequest
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-accountRequest-L294`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_GO-accountRequest-L294`
- **Regression Risk:** `MEDIUM`

Function "accountRequest" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 7. [MEDIUM] Excessive cyclomatic complexity (M=27) (+ 1 related findings) (`RP-0004`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ORBITCONTROLS_JS-onMouseDown-L810`, `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ORBITCONTROLS_JS-onMouseDown-L810`
- **Regression Risk:** `MEDIUM`

Function "onMouseDown" has cyclomatic complexity of 27, requiring at least 27 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 8. [MEDIUM] Excessive control flow nesting depth (6) (+ 2 related findings) (`RP-0005`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/aggregator.go`
- **Target Symbols:** GenerateMarkdown
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_AGGREGATOR_GO-GenerateMarkdown-L86`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_AGGREGATOR_GO-GenerateMarkdown-L86`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_AGGREGATOR_GO-GenerateMarkdown-L86`
- **Regression Risk:** `LOW`

Function "GenerateMarkdown" in go-backend/agent/repo_intel/aggregator.go reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 9. [MEDIUM] Excessive cyclomatic complexity (M=73) (+ 2 related findings) (`RP-0006`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/relationships.rs`
- **Target Symbols:** walk
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-walk-L41`, `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-walk-L41`, `RI-STR-LARGE-AST-VOLUME-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-walk-L41`
- **Regression Risk:** `MEDIUM`

Function "walk" has cyclomatic complexity of 73, requiring at least 73 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 10. [MEDIUM] Excessive cyclomatic complexity (M=20) (+ 1 related findings) (`RP-0007`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-Reconcile-L307`, `RI-STR-HIGH-NESTING-GO_BACKEND_SANDBOX_LIFECYCLE_GO-Reconcile-L307`
- **Regression Risk:** `MEDIUM`

Function "Reconcile" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 11. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0008`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/evolution-api.js`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ADMIN_PORTAL_SRC_EVOLUTION_API_JS-processLine-L20`, `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_EVOLUTION_API_JS-processLine-L20`
- **Regression Risk:** `LOW`

Function "processLine" in admin-portal/src/evolution-api.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 12. [MEDIUM] Excessive cyclomatic complexity (M=14) (+ 1 related findings) (`RP-0009`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/analysis_proxy_test.go`
- **Target Symbols:** TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ANALYSIS_PROXY_TEST_GO-TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide-L16`, `RI-STR-LARGE-AST-VOLUME-API_ANALYSIS_PROXY_TEST_GO-TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide-L16`
- **Regression Risk:** `MEDIUM`

Function "TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 13. [MEDIUM] Excessive cyclomatic complexity (M=52) (+ 1 related findings) (`RP-0010`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/evaluator.go`
- **Target Symbols:** EvaluateCandidate
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_EVALUATOR_GO-EvaluateCandidate-L188`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_EVALUATOR_GO-EvaluateCandidate-L188`
- **Regression Risk:** `MEDIUM`

Function "EvaluateCandidate" has cyclomatic complexity of 52, requiring at least 52 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 14. [MEDIUM] Excessive cyclomatic complexity (M=21) (+ 1 related findings) (`RP-0011`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Target Symbols:** ServeHTTP
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-ServeHTTP-L96`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_GO-ServeHTTP-L96`
- **Regression Risk:** `MEDIUM`

Function "ServeHTTP" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 15. [MEDIUM] Excessive cyclomatic complexity (M=53) (+ 2 related findings) (`RP-0012`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/analysis_proxy.go`
- **Target Symbols:** registerAnalysisRoutes
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ANALYSIS_PROXY_GO-registerAnalysisRoutes-L42`, `RI-STR-HIGH-NESTING-API_ANALYSIS_PROXY_GO-registerAnalysisRoutes-L42`, `RI-STR-LARGE-AST-VOLUME-API_ANALYSIS_PROXY_GO-registerAnalysisRoutes-L42`
- **Regression Risk:** `MEDIUM`

Function "registerAnalysisRoutes" has cyclomatic complexity of 53, requiring at least 53 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 16. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0013`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts_test.go`
- **Target Symbols:** exerciseAccountStore
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-API_ACCOUNTS_TEST_GO-exerciseAccountStore-L305`, `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_TEST_GO-exerciseAccountStore-L305`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_TEST_GO-exerciseAccountStore-L305`
- **Regression Risk:** `LOW`

Function "exerciseAccountStore" in api/accounts_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 17. [MEDIUM] Excessive control flow nesting depth (5) (+ 2 related findings) (`RP-0014`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/reporter.go`
- **Target Symbols:** GenerateMarkdown
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_REPORTER_GO-GenerateMarkdown-L149`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_REPORTER_GO-GenerateMarkdown-L149`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_REPORTER_GO-GenerateMarkdown-L149`
- **Regression Risk:** `LOW`

Function "GenerateMarkdown" in go-backend/agent/repo_intel/diagnostic/reporter.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 18. [MEDIUM] Excessive cyclomatic complexity (M=41) (+ 2 related findings) (`RP-0015`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Target Symbols:** checkBufferEscapes
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-checkBufferEscapes-L910`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-checkBufferEscapes-L910`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_STRATEGY_GO-checkBufferEscapes-L910`
- **Regression Risk:** `MEDIUM`

Function "checkBufferEscapes" has cyclomatic complexity of 41, requiring at least 41 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 19. [MEDIUM] Excessive cyclomatic complexity (M=88) (+ 1 related findings) (`RP-0016`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/components/AdminPortal.jsx`
- **Target Symbols:** AdminPortal
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_COMPONENTS_ADMINPORTAL_JSX-AdminPortal-L30`, `RI-STR-LARGE-AST-VOLUME-ADMIN_PORTAL_SRC_COMPONENTS_ADMINPORTAL_JSX-AdminPortal-L30`
- **Regression Risk:** `MEDIUM`

Function "AdminPortal" has cyclomatic complexity of 88, requiring at least 88 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src/components
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 20. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0017`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-check-L949`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-check-L949`
- **Regression Risk:** `LOW`

Function "check" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 21. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0018`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/tests/graph-model.test.js`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ADMIN_PORTAL_TESTS_GRAPH_MODEL_TEST_JS-assertClearConnections-L6`, `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_TESTS_GRAPH_MODEL_TEST_JS-assertClearConnections-L6`
- **Regression Risk:** `LOW`

Function "assertClearConnections" in admin-portal/tests/graph-model.test.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/tests
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 22. [MEDIUM] Excessive cyclomatic complexity (M=12) (+ 1 related findings) (`RP-0019`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Target Symbols:** ReconcileStaleLeases
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-ReconcileStaleLeases-L674`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SANDBOX_LIFECYCLE_GO-ReconcileStaleLeases-L674`
- **Regression Risk:** `MEDIUM`

Function "ReconcileStaleLeases" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 23. [MEDIUM] Excessive cyclomatic complexity (M=31) (+ 1 related findings) (`RP-0020`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-inventory-L380`, `RI-STR-HIGH-NESTING-SCRIPTS_ANALYSIS_CLEANUP_MJS-inventory-L380`
- **Regression Risk:** `MEDIUM`

Function "inventory" has cyclomatic complexity of 31, requiring at least 31 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 24. [MEDIUM] Excessive cyclomatic complexity (M=17) (+ 1 related findings) (`RP-0021`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/manifests.go`
- **Target Symbols:** AnalyzeManifests
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-AnalyzeManifests-L23`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-AnalyzeManifests-L23`
- **Regression Risk:** `MEDIUM`

Function "AnalyzeManifests" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 25. [MEDIUM] Excessive cyclomatic complexity (M=24) (+ 1 related findings) (`RP-0022`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/workspace.go`
- **Target Symbols:** ExtractWorkspaceArchive
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_WORKSPACE_GO-ExtractWorkspaceArchive-L154`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_WORKSPACE_GO-ExtractWorkspaceArchive-L154`
- **Regression Risk:** `MEDIUM`

Function "ExtractWorkspaceArchive" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 26. [MEDIUM] Excessive cyclomatic complexity (M=30) (+ 1 related findings) (`RP-0023`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Target Symbols:** scan_file
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-scan_file-L700`, `RI-STR-LARGE-AST-VOLUME-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-scan_file-L700`
- **Regression Risk:** `MEDIUM`

Function "scan_file" has cyclomatic complexity of 30, requiring at least 30 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 27. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0024`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/local_store.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-API_LOCAL_STORE_GO-PromoteUpload-L189`, `RI-STR-HIGH-COMPLEXITY-API_LOCAL_STORE_GO-PromoteUpload-L189`
- **Regression Risk:** `LOW`

Function "PromoteUpload" in api/local_store.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 28. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0025`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-checkUsesLoopVar-L1017`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-checkUsesLoopVar-L1017`
- **Regression Risk:** `LOW`

Function "checkUsesLoopVar" in go-backend/agent/strategy.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 29. [MEDIUM] Excessive cyclomatic complexity (M=13) (+ 1 related findings) (`RP-0026`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/energy_test.go`
- **Target Symbols:** TestAnnealer_DownhillImprovement
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_ENERGY_TEST_GO-TestAnnealer_DownhillImprovement-L197`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ENERGY_TEST_GO-TestAnnealer_DownhillImprovement-L197`
- **Regression Risk:** `MEDIUM`

Function "TestAnnealer_DownhillImprovement" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 30. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0027`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/routes.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ROUTES_MJS-<anonymous@86>-L86`, `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ROUTES_MJS-<anonymous@86>-L86`
- **Regression Risk:** `LOW`

Function "<anonymous@86>" in graph-visualizer/routes.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 31. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0028`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-collect_functions-L762`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-collect_functions-L762`
- **Regression Risk:** `LOW`

Function "collect_functions" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 32. [MEDIUM] Excessive control flow nesting depth (6) (+ 1 related findings) (`RP-0029`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/data_pipeline.py`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_DATA_PIPELINE_PY-process_batch-L79`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_DATA_PIPELINE_PY-process_batch-L79`
- **Regression Risk:** `LOW`

Function "process_batch" in thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 33. [MEDIUM] Excessive cyclomatic complexity (M=60) (+ 2 related findings) (`RP-0030`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/graph.mjs`
- **Target Symbols:** draw
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_GRAPH_MJS-draw-L44`, `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_GRAPH_MJS-draw-L44`, `RI-STR-LARGE-AST-VOLUME-GRAPH_VISUALIZER_GRAPH_MJS-draw-L44`
- **Regression Risk:** `MEDIUM`

Function "draw" has cyclomatic complexity of 60, requiring at least 60 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 34. [MEDIUM] Excessive cyclomatic complexity (M=21) (+ 1 related findings) (`RP-0031`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/energy.go`
- **Target Symbols:** EvaluateEnergy
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_ENERGY_GO-EvaluateEnergy-L155`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ENERGY_GO-EvaluateEnergy-L155`
- **Regression Risk:** `MEDIUM`

Function "EvaluateEnergy" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 35. [MEDIUM] Excessive cyclomatic complexity (M=24) (+ 1 related findings) (`RP-0032`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/lambda_test.go`
- **Target Symbols:** TestAPIGatewayV2HandlerMapsRequestAndResponse
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LAMBDA_TEST_GO-TestAPIGatewayV2HandlerMapsRequestAndResponse-L241`, `RI-STR-LARGE-AST-VOLUME-API_LAMBDA_TEST_GO-TestAPIGatewayV2HandlerMapsRequestAndResponse-L241`
- **Regression Risk:** `MEDIUM`

Function "TestAPIGatewayV2HandlerMapsRequestAndResponse" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 36. [MEDIUM] Excessive cyclomatic complexity (M=38) (+ 1 related findings) (`RP-0033`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_evaluator.go`
- **Target Symbols:** EvaluateCandidate
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_EVALUATOR_GO-EvaluateCandidate-L209`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_BROKER_EVALUATOR_GO-EvaluateCandidate-L209`
- **Regression Risk:** `MEDIUM`

Function "EvaluateCandidate" has cyclomatic complexity of 38, requiring at least 38 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 37. [MEDIUM] Excessive cyclomatic complexity (M=46) (+ 1 related findings) (`RP-0034`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/languages.rs`
- **Target Symbols:** name
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_LANGUAGES_RS-name-L43`, `RI-STR-LARGE-AST-VOLUME-THERMODYNAMIC_AST_ENGINE_SRC_LANGUAGES_RS-name-L43`
- **Regression Risk:** `MEDIUM`

Function "name" has cyclomatic complexity of 46, requiring at least 46 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 38. [MEDIUM] Excessive cyclomatic complexity (M=103) (+ 2 related findings) (`RP-0035`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/relationships.rs`
- **Target Symbols:** build
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-build-L260`, `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-build-L260`, `RI-STR-LARGE-AST-VOLUME-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-build-L260`
- **Regression Risk:** `MEDIUM`

Function "build" has cyclomatic complexity of 103, requiring at least 103 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 39. [MEDIUM] Excessive cyclomatic complexity (M=31) (+ 1 related findings) (`RP-0036`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/vite.config.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_VITE_CONFIG_JS-<anonymous@36>-L36`, `RI-STR-HIGH-NESTING-ADMIN_PORTAL_VITE_CONFIG_JS-<anonymous@36>-L36`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@36>" has cyclomatic complexity of 31, requiring at least 31 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 40. [MEDIUM] Excessive cyclomatic complexity (M=32) (+ 2 related findings) (`RP-0037`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Target Symbols:** extractCommonSubexpressions
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-extractCommonSubexpressions-L705`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-extractCommonSubexpressions-L705`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_STRATEGY_GO-extractCommonSubexpressions-L705`
- **Regression Risk:** `MEDIUM`

Function "extractCommonSubexpressions" has cyclomatic complexity of 32, requiring at least 32 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 41. [MEDIUM] Excessive control flow nesting depth (5) (+ 1 related findings) (`RP-0038`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/graph.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_GRAPH_MJS-<anonymous@330>-L330`, `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_GRAPH_MJS-<anonymous@330>-L330`
- **Regression Risk:** `LOW`

Function "<anonymous@330>" in graph-visualizer/graph.mjs reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 42. [MEDIUM] Excessive cyclomatic complexity (M=22) (+ 2 related findings) (`RP-0039`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/orchestrator.go`
- **Target Symbols:** Run
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_ORCHESTRATOR_GO-Run-L42`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_ORCHESTRATOR_GO-Run-L42`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_ORCHESTRATOR_GO-Run-L42`
- **Regression Risk:** `MEDIUM`

Function "Run" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 43. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0040`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/main.go`
- **Target Symbols:** analyzeHandler
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_MAIN_GO-analyzeHandler-L341`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_MAIN_GO-analyzeHandler-L341`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_MAIN_GO-analyzeHandler-L341`
- **Regression Risk:** `LOW`

Function "analyzeHandler" in go-backend/main.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 44. [MEDIUM] Excessive cyclomatic complexity (M=19) (+ 1 related findings) (`RP-0041`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/account_store.go`
- **Target Symbols:** access
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNT_STORE_GO-access-L79`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNT_STORE_GO-access-L79`
- **Regression Risk:** `MEDIUM`

Function "access" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 45. [MEDIUM] Excessive cyclomatic complexity (M=40) (+ 2 related findings) (`RP-0042`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/broker_workspace.go`
- **Target Symbols:** newSandboxBrokerHandler
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_BROKER_WORKSPACE_GO-newSandboxBrokerHandler-L276`, `RI-STR-HIGH-NESTING-GO_BACKEND_BROKER_WORKSPACE_GO-newSandboxBrokerHandler-L276`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_BROKER_WORKSPACE_GO-newSandboxBrokerHandler-L276`
- **Regression Risk:** `MEDIUM`

Function "newSandboxBrokerHandler" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 46. [MEDIUM] Excessive control flow nesting depth (5) (+ 1 related findings) (`RP-0043`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-flattenNestedIf-L419`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-flattenNestedIf-L419`
- **Regression Risk:** `LOW`

Function "flattenNestedIf" in go-backend/agent/strategy.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 47. [MEDIUM] Excessive cyclomatic complexity (M=17) (+ 1 related findings) (`RP-0044`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/main.go`
- **Target Symbols:** extractZip
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_MAIN_GO-extractZip-L165`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_MAIN_GO-extractZip-L165`
- **Regression Risk:** `MEDIUM`

Function "extractZip" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 48. [MEDIUM] Excessive cyclomatic complexity (M=13) (+ 1 related findings) (`RP-0045`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_test.go`
- **Target Symbols:** TestRemoteBrokerEvaluator_SuccessfulEvaluation
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_TEST_GO-TestRemoteBrokerEvaluator_SuccessfulEvaluation-L322`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_BROKER_TEST_GO-TestRemoteBrokerEvaluator_SuccessfulEvaluation-L322`
- **Regression Risk:** `MEDIUM`

Function "TestRemoteBrokerEvaluator_SuccessfulEvaluation" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 49. [MEDIUM] Excessive cyclomatic complexity (M=40) (+ 1 related findings) (`RP-0046`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/quality.go`
- **Target Symbols:** AnalyzeQuality
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_QUALITY_GO-AnalyzeQuality-L21`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_QUALITY_GO-AnalyzeQuality-L21`
- **Regression Risk:** `MEDIUM`

Function "AnalyzeQuality" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 50. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0047`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/snapshot_scanner.go`
- **Target Symbols:** scanRepoHandlerWithRunner
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_SNAPSHOT_SCANNER_GO-scanRepoHandlerWithRunner-L32`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SNAPSHOT_SCANNER_GO-scanRepoHandlerWithRunner-L32`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SNAPSHOT_SCANNER_GO-scanRepoHandlerWithRunner-L32`
- **Regression Risk:** `LOW`

Function "scanRepoHandlerWithRunner" in go-backend/snapshot_scanner.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 51. [MEDIUM] Excessive cyclomatic complexity (M=16) (+ 1 related findings) (`RP-0048`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/aws_store_test.go`
- **Target Symbols:** TestS3ObjectStorePresignsBoundedPendingUploadAndDownload
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AWS_STORE_TEST_GO-TestS3ObjectStorePresignsBoundedPendingUploadAndDownload-L689`, `RI-STR-LARGE-AST-VOLUME-API_AWS_STORE_TEST_GO-TestS3ObjectStorePresignsBoundedPendingUploadAndDownload-L689`
- **Regression Risk:** `MEDIUM`

Function "TestS3ObjectStorePresignsBoundedPendingUploadAndDownload" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 52. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0049`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/routes.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ROUTES_MJS-simplify-L5`, `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ROUTES_MJS-simplify-L5`
- **Regression Risk:** `LOW`

Function "simplify" in graph-visualizer/routes.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 53. [MEDIUM] Excessive cyclomatic complexity (M=13) (+ 1 related findings) (`RP-0050`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/main.go`
- **Target Symbols:** auditHandler
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_MAIN_GO-auditHandler-L503`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_MAIN_GO-auditHandler-L503`
- **Regression Risk:** `MEDIUM`

Function "auditHandler" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 54. [MEDIUM] Excessive cyclomatic complexity (M=17) (+ 1 related findings) (`RP-0051`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/local_store.go`
- **Target Symbols:** StageUpload
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LOCAL_STORE_GO-StageUpload-L85`, `RI-STR-LARGE-AST-VOLUME-API_LOCAL_STORE_GO-StageUpload-L85`
- **Regression Risk:** `MEDIUM`

Function "StageUpload" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 55. [MEDIUM] Excessive cyclomatic complexity (M=28) (+ 1 related findings) (`RP-0052`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/workspace.go`
- **Target Symbols:** CreateWorkspaceArchive
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_WORKSPACE_GO-CreateWorkspaceArchive-L41`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_WORKSPACE_GO-CreateWorkspaceArchive-L41`
- **Regression Risk:** `MEDIUM`

Function "CreateWorkspaceArchive" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 56. [MEDIUM] Excessive cyclomatic complexity (M=42) (+ 2 related findings) (`RP-0053`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Target Symbols:** <anonymous@219>
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-<anonymous@219>-L219`, `RI-STR-HIGH-NESTING-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-<anonymous@219>-L219`, `RI-STR-LARGE-AST-VOLUME-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-<anonymous@219>-L219`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@219>" has cyclomatic complexity of 42, requiring at least 42 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 57. [MEDIUM] Excessive cyclomatic complexity (M=20) (+ 1 related findings) (`RP-0054`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `source-stager/main.go`
- **Target Symbols:** validateArchive
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SOURCE_STAGER_MAIN_GO-validateArchive-L382`, `RI-STR-LARGE-AST-VOLUME-SOURCE_STAGER_MAIN_GO-validateArchive-L382`
- **Regression Risk:** `MEDIUM`

Function "validateArchive" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./source-stager
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 58. [MEDIUM] Excessive cyclomatic complexity (M=24) (+ 1 related findings) (`RP-0055`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/local_store_test.go`
- **Target Symbols:** TestLocalArchiveUploadFinalizeListDownloadDelete
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LOCAL_STORE_TEST_GO-TestLocalArchiveUploadFinalizeListDownloadDelete-L18`, `RI-STR-LARGE-AST-VOLUME-API_LOCAL_STORE_TEST_GO-TestLocalArchiveUploadFinalizeListDownloadDelete-L18`
- **Regression Risk:** `MEDIUM`

Function "TestLocalArchiveUploadFinalizeListDownloadDelete" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 59. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0056`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-check-L870`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-check-L870`
- **Regression Risk:** `LOW`

Function "check" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 60. [MEDIUM] Excessive cyclomatic complexity (M=28) (+ 2 related findings) (`RP-0057`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Target Symbols:** update
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ORBITCONTROLS_JS-update-L166`, `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ORBITCONTROLS_JS-update-L166`, `RI-STR-LARGE-AST-VOLUME-GRAPH_VISUALIZER_ORBITCONTROLS_JS-update-L166`
- **Regression Risk:** `MEDIUM`

Function "update" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 61. [MEDIUM] Excessive cyclomatic complexity (M=29) (+ 1 related findings) (`RP-0058`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/scanner.go`
- **Target Symbols:** scanRepoHandler
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SCANNER_GO-scanRepoHandler-L25`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SCANNER_GO-scanRepoHandler-L25`
- **Regression Risk:** `MEDIUM`

Function "scanRepoHandler" has cyclomatic complexity of 29, requiring at least 29 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 62. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0059`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/refresh_test.go`
- **Target Symbols:** TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-API_REFRESH_TEST_GO-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`, `RI-STR-HIGH-COMPLEXITY-API_REFRESH_TEST_GO-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`, `RI-STR-LARGE-AST-VOLUME-API_REFRESH_TEST_GO-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`
- **Regression Risk:** `LOW`

Function "TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials" in api/refresh_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 63. [MEDIUM] Excessive cyclomatic complexity (M=49) (+ 2 related findings) (`RP-0060`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Target Symbols:** visit
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-visit-L490`, `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-visit-L490`, `RI-STR-LARGE-AST-VOLUME-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-visit-L490`
- **Regression Risk:** `MEDIUM`

Function "visit" has cyclomatic complexity of 49, requiring at least 49 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 64. [MEDIUM] Excessive cyclomatic complexity (M=43) (+ 2 related findings) (`RP-0061`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/discovery.go`
- **Target Symbols:** DiscoverRepository
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DISCOVERY_GO-DiscoverRepository-L169`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_DISCOVERY_GO-DiscoverRepository-L169`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DISCOVERY_GO-DiscoverRepository-L169`
- **Regression Risk:** `MEDIUM`

Function "DiscoverRepository" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 65. [MEDIUM] Excessive cyclomatic complexity (M=16) (+ 1 related findings) (`RP-0062`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/correlation.go`
- **Target Symbols:** CorrelateFindings
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_CORRELATION_GO-CorrelateFindings-L19`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_CORRELATION_GO-CorrelateFindings-L19`
- **Regression Risk:** `MEDIUM`

Function "CorrelateFindings" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 66. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0063`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/lambda_test.go`
- **Target Symbols:** TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-API_LAMBDA_TEST_GO-TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey-L339`, `RI-STR-LARGE-AST-VOLUME-API_LAMBDA_TEST_GO-TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey-L339`
- **Regression Risk:** `LOW`

Function "TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey" in api/lambda_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 67. [MEDIUM] Excessive cyclomatic complexity (M=43) (+ 1 related findings) (`RP-0064`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/archive.go`
- **Target Symbols:** isPowerPointCompoundFile
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ARCHIVE_GO-isPowerPointCompoundFile-L528`, `RI-STR-LARGE-AST-VOLUME-API_ARCHIVE_GO-isPowerPointCompoundFile-L528`
- **Regression Risk:** `MEDIUM`

Function "isPowerPointCompoundFile" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 68. [MEDIUM] Excessive cyclomatic complexity (M=53) (+ 2 related findings) (`RP-0065`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/model.mjs`
- **Target Symbols:** buildGraphModel
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_MODEL_MJS-buildGraphModel-L23`, `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_MODEL_MJS-buildGraphModel-L23`, `RI-STR-LARGE-AST-VOLUME-GRAPH_VISUALIZER_MODEL_MJS-buildGraphModel-L23`
- **Regression Risk:** `MEDIUM`

Function "buildGraphModel" has cyclomatic complexity of 53, requiring at least 53 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 69. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0066`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `analysis-authorizer/main_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ANALYSIS_AUTHORIZER_MAIN_TEST_GO-TestSSMSecretStore_RequiresDecryptedSecureString-L227`, `RI-STR-HIGH-COMPLEXITY-ANALYSIS_AUTHORIZER_MAIN_TEST_GO-TestSSMSecretStore_RequiresDecryptedSecureString-L227`
- **Regression Risk:** `LOW`

Function "TestSSMSecretStore_RequiresDecryptedSecureString" in analysis-authorizer/main_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./analysis-authorizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 70. [MEDIUM] Excessive control flow nesting depth (7) (+ 1 related findings) (`RP-0067`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `samples/benchmark/benchmark_target.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-SAMPLES_BENCHMARK_BENCHMARK_TARGET_GO-EvaluateRiskEngine-L21`, `RI-STR-HIGH-COMPLEXITY-SAMPLES_BENCHMARK_BENCHMARK_TARGET_GO-EvaluateRiskEngine-L21`
- **Regression Risk:** `LOW`

Function "EvaluateRiskEngine" in samples/benchmark/benchmark_target.go reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./samples/benchmark
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 71. [MEDIUM] Excessive cyclomatic complexity (M=19) (+ 1 related findings) (`RP-0068`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/local_store.go`
- **Target Symbols:** registerLocalFileRoutes
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LOCAL_STORE_GO-registerLocalFileRoutes-L505`, `RI-STR-LARGE-AST-VOLUME-API_LOCAL_STORE_GO-registerLocalFileRoutes-L505`
- **Regression Risk:** `MEDIUM`

Function "registerLocalFileRoutes" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 72. [MEDIUM] Excessive cyclomatic complexity (M=21) (+ 2 related findings) (`RP-0069`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Target Symbols:** hoistLoopAllocations
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-hoistLoopAllocations-L843`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-hoistLoopAllocations-L843`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_STRATEGY_GO-hoistLoopAllocations-L843`
- **Regression Risk:** `MEDIUM`

Function "hoistLoopAllocations" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 73. [MEDIUM] Excessive cyclomatic complexity (M=33) (+ 1 related findings) (`RP-0070`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/graph.mjs`
- **Target Symbols:** details
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_GRAPH_MJS-details-L131`, `RI-STR-LARGE-AST-VOLUME-GRAPH_VISUALIZER_GRAPH_MJS-details-L131`
- **Regression Risk:** `MEDIUM`

Function "details" has cyclomatic complexity of 33, requiring at least 33 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 74. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0071`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/severity.go`
- **Target Symbols:** AssessSeverityAndImpact
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_SEVERITY_GO-AssessSeverityAndImpact-L20`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_SEVERITY_GO-AssessSeverityAndImpact-L20`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_SEVERITY_GO-AssessSeverityAndImpact-L20`
- **Regression Risk:** `LOW`

Function "AssessSeverityAndImpact" in go-backend/agent/repo_intel/diagnostic/severity.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 75. [MEDIUM] Excessive cyclomatic complexity (M=17) (+ 1 related findings) (`RP-0072`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/workspace_test.go`
- **Target Symbols:** TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_WORKSPACE_TEST_GO-TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles-L13`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_WORKSPACE_TEST_GO-TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles-L13`
- **Regression Risk:** `MEDIUM`

Function "TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 76. [MEDIUM] Excessive cyclomatic complexity (M=21) (+ 1 related findings) (`RP-0073`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/main.go`
- **Target Symbols:** serveDownload
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_MAIN_GO-serveDownload-L337`, `RI-STR-LARGE-AST-VOLUME-API_MAIN_GO-serveDownload-L337`
- **Regression Risk:** `MEDIUM`

Function "serveDownload" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 77. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0074`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ORBITCONTROLS_JS-onTouchStart-L919`, `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ORBITCONTROLS_JS-onTouchStart-L919`
- **Regression Risk:** `LOW`

Function "onTouchStart" in graph-visualizer/OrbitControls.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 78. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0075`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/relationships.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-walk-L194`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-walk-L194`
- **Regression Risk:** `LOW`

Function "walk" in thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 79. [MEDIUM] Excessive cyclomatic complexity (M=16) (+ 1 related findings) (`RP-0076`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Target Symbols:** exchangeGitHubCode
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-exchangeGitHubCode-L243`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_GO-exchangeGitHubCode-L243`
- **Regression Risk:** `MEDIUM`

Function "exchangeGitHubCode" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 80. [MEDIUM] Excessive control flow nesting depth (6) (+ 1 related findings) (`RP-0077`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/routes.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ROUTES_MJS-<anonymous@42>-L42`, `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ROUTES_MJS-<anonymous@42>-L42`
- **Regression Risk:** `LOW`

Function "<anonymous@42>" in graph-visualizer/routes.mjs reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 81. [MEDIUM] Excessive cyclomatic complexity (M=16) (+ 1 related findings) (`RP-0078`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_test.go`
- **Target Symbols:** TestSandboxRunnerIsolationAndCIPinsImage
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_TEST_GO-TestSandboxRunnerIsolationAndCIPinsImage-L30`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_BROKER_TEST_GO-TestSandboxRunnerIsolationAndCIPinsImage-L30`
- **Regression Risk:** `MEDIUM`

Function "TestSandboxRunnerIsolationAndCIPinsImage" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 82. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0079`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/structural.go`
- **Target Symbols:** convertReportToFindings
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_STRUCTURAL_GO-convertReportToFindings-L67`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_STRUCTURAL_GO-convertReportToFindings-L67`
- **Regression Risk:** `LOW`

Function "convertReportToFindings" in go-backend/agent/repo_intel/structural.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 83. [MEDIUM] Excessive cyclomatic complexity (M=13) (+ 1 related findings) (`RP-0080`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/archive.go`
- **Target Symbols:** CompleteUpload
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ARCHIVE_GO-CompleteUpload-L175`, `RI-STR-LARGE-AST-VOLUME-API_ARCHIVE_GO-CompleteUpload-L175`
- **Regression Risk:** `MEDIUM`

Function "CompleteUpload" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 84. [MEDIUM] Excessive cyclomatic complexity (M=69) (+ 2 related findings) (`RP-0081`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/annealer.go`
- **Target Symbols:** Evolve
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_ANNEALER_GO-Evolve-L99`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_ANNEALER_GO-Evolve-L99`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ANNEALER_GO-Evolve-L99`
- **Regression Risk:** `MEDIUM`

Function "Evolve" has cyclomatic complexity of 69, requiring at least 69 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 85. [MEDIUM] Excessive cyclomatic complexity (M=89) (+ 2 related findings) (`RP-0082`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/router.go`
- **Target Symbols:** newRouter
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ROUTER_GO-newRouter-L224`, `RI-STR-HIGH-NESTING-API_ROUTER_GO-newRouter-L224`, `RI-STR-LARGE-AST-VOLUME-API_ROUTER_GO-newRouter-L224`
- **Regression Risk:** `MEDIUM`

Function "newRouter" has cyclomatic complexity of 89, requiring at least 89 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 86. [MEDIUM] Excessive control flow nesting depth (5) (+ 2 related findings) (`RP-0083`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/architecture.go`
- **Target Symbols:** detectCycles
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_ARCHITECTURE_GO-detectCycles-L103`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_ARCHITECTURE_GO-detectCycles-L103`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_ARCHITECTURE_GO-detectCycles-L103`
- **Regression Risk:** `LOW`

Function "detectCycles" in go-backend/agent/repo_intel/architecture.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 87. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0084`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/manifests.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-parseGoMod-L162`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-parseGoMod-L162`
- **Regression Risk:** `LOW`

Function "parseGoMod" in go-backend/agent/repo_intel/manifests.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 88. [MEDIUM] Excessive cyclomatic complexity (M=64) (+ 1 related findings) (`RP-0085`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-is_concurrency_primitive-L341`, `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-is_concurrency_primitive-L341`
- **Regression Risk:** `MEDIUM`

Function "is_concurrency_primitive" has cyclomatic complexity of 64, requiring at least 64 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 89. [MEDIUM] Excessive cyclomatic complexity (M=19) (+ 1 related findings) (`RP-0086`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts_test.go`
- **Target Symbols:** TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_TEST_GO-TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart-L149`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_TEST_GO-TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart-L149`
- **Regression Risk:** `MEDIUM`

Function "TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 90. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0087`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_TEST_GO-TestStrategy_DifferentialEquivalence-L610`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_TEST_GO-TestStrategy_DifferentialEquivalence-L610`
- **Regression Risk:** `LOW`

Function "TestStrategy_DifferentialEquivalence" in go-backend/agent/strategy_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 91. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0088`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/evolution-api.js`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ADMIN_PORTAL_SRC_EVOLUTION_API_JS-readEvolutionStream-L5`, `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_EVOLUTION_API_JS-readEvolutionStream-L5`
- **Regression Risk:** `LOW`

Function "readEvolutionStream" in admin-portal/src/evolution-api.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 92. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0089`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-count-L912`, `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-count-L912`
- **Regression Risk:** `LOW`

Function "count" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 93. [MEDIUM] Excessive cyclomatic complexity (M=51) (+ 2 related findings) (`RP-0090`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/evolve.go`
- **Target Symbols:** evolveHandler
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_EVOLVE_GO-evolveHandler-L42`, `RI-STR-HIGH-NESTING-GO_BACKEND_EVOLVE_GO-evolveHandler-L42`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_EVOLVE_GO-evolveHandler-L42`
- **Regression Risk:** `MEDIUM`

Function "evolveHandler" has cyclomatic complexity of 51, requiring at least 51 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 94. [MEDIUM] Excessive cyclomatic complexity (M=23) (+ 2 related findings) (`RP-0091`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Target Symbols:** simplifyBooleanReturns
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-simplifyBooleanReturns-L511`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-simplifyBooleanReturns-L511`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_STRATEGY_GO-simplifyBooleanReturns-L511`
- **Regression Risk:** `MEDIUM`

Function "simplifyBooleanReturns" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 95. [MEDIUM] Excessive cyclomatic complexity (M=23) (+ 1 related findings) (`RP-0092`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/snapshot_portal_test.go`
- **Target Symbols:** TestSnapshotPortalScanEvolutionAndPatch
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SNAPSHOT_PORTAL_TEST_GO-TestSnapshotPortalScanEvolutionAndPatch-L19`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SNAPSHOT_PORTAL_TEST_GO-TestSnapshotPortalScanEvolutionAndPatch-L19`
- **Regression Risk:** `MEDIUM`

Function "TestSnapshotPortalScanEvolutionAndPatch" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 96. [MEDIUM] Excessive cyclomatic complexity (M=14) (+ 1 related findings) (`RP-0093`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/workspace_test.go`
- **Target Symbols:** TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_WORKSPACE_TEST_GO-TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes-L52`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_WORKSPACE_TEST_GO-TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes-L52`
- **Regression Risk:** `MEDIUM`

Function "TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 97. [MEDIUM] Excessive cyclomatic complexity (M=15) (+ 1 related findings) (`RP-0094`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/snapshot_store.go`
- **Target Symbols:** ExtractSnapshot
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SNAPSHOT_STORE_GO-ExtractSnapshot-L77`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SNAPSHOT_STORE_GO-ExtractSnapshot-L77`
- **Regression Risk:** `MEDIUM`

Function "ExtractSnapshot" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 98. [MEDIUM] Excessive control flow nesting depth (4) (+ 2 related findings) (`RP-0095`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/reliability_security.go`
- **Target Symbols:** AnalyzeReliabilityAndSecurity
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_RELIABILITY_SECURITY_GO-AnalyzeReliabilityAndSecurity-L42`, `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_RELIABILITY_SECURITY_GO-AnalyzeReliabilityAndSecurity-L42`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_RELIABILITY_SECURITY_GO-AnalyzeReliabilityAndSecurity-L42`
- **Regression Risk:** `LOW`

Function "AnalyzeReliabilityAndSecurity" in go-backend/agent/repo_intel/reliability_security.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 99. [MEDIUM] Excessive cyclomatic complexity (M=22) (+ 1 related findings) (`RP-0096`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/evolve.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_EVOLVE_GO-streamEvolutionEvents-L206`, `RI-STR-HIGH-NESTING-GO_BACKEND_EVOLVE_GO-streamEvolutionEvents-L206`
- **Regression Risk:** `MEDIUM`

Function "streamEvolutionEvents" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 100. [MEDIUM] Excessive cyclomatic complexity (M=25) (+ 1 related findings) (`RP-0097`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/validator.go`
- **Target Symbols:** validateSingleFinding
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_VALIDATOR_GO-validateSingleFinding-L53`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_VALIDATOR_GO-validateSingleFinding-L53`
- **Regression Risk:** `MEDIUM`

Function "validateSingleFinding" has cyclomatic complexity of 25, requiring at least 25 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 101. [MEDIUM] Excessive cyclomatic complexity (M=29) (+ 1 related findings) (`RP-0098`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/router.go`
- **Target Symbols:** registerArchiveServiceRoutes
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ROUTER_GO-registerArchiveServiceRoutes-L80`, `RI-STR-LARGE-AST-VOLUME-API_ROUTER_GO-registerArchiveServiceRoutes-L80`
- **Regression Risk:** `MEDIUM`

Function "registerArchiveServiceRoutes" has cyclomatic complexity of 29, requiring at least 29 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 102. [MEDIUM] Excessive cyclomatic complexity (M=12) (+ 1 related findings) (`RP-0099`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts_test.go`
- **Target Symbols:** testAccountService
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_TEST_GO-testAccountService-L23`, `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_TEST_GO-testAccountService-L23`
- **Regression Risk:** `MEDIUM`

Function "testAccountService" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 103. [MEDIUM] Excessive cyclomatic complexity (M=20) (+ 2 related findings) (`RP-0100`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Target Symbols:** ProposeTargeted
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-ProposeTargeted-L317`, `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-ProposeTargeted-L317`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_STRATEGY_GO-ProposeTargeted-L317`
- **Regression Risk:** `MEDIUM`

Function "ProposeTargeted" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 104. [MEDIUM] Excessive cyclomatic complexity (M=22) (+ 1 related findings) (`RP-0101`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/workspace.go`
- **Target Symbols:** readWorkspaceTarget
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_WORKSPACE_GO-readWorkspaceTarget-L145`, `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_WORKSPACE_GO-readWorkspaceTarget-L145`
- **Regression Risk:** `MEDIUM`

Function "readWorkspaceTarget" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 105. [MEDIUM] Excessive control flow nesting depth (4) (+ 1 related findings) (`RP-0102`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/main_test.go`
- **Target Symbols:** TestLegacyUploadRejectsInvalidFilesBeforeStorage
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-API_MAIN_TEST_GO-TestLegacyUploadRejectsInvalidFilesBeforeStorage-L443`, `RI-STR-LARGE-AST-VOLUME-API_MAIN_TEST_GO-TestLegacyUploadRejectsInvalidFilesBeforeStorage-L443`
- **Regression Risk:** `LOW`

Function "TestLegacyUploadRejectsInvalidFilesBeforeStorage" in api/main_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 106. [MEDIUM] Excessive cyclomatic complexity (M=20) (`RP-0104`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/archive.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ARCHIVE_GO-isPowerPointZip-L465`
- **Regression Risk:** `MEDIUM`

Function "isPowerPointZip" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 107. [MEDIUM] Excessive cyclomatic complexity (M=20) (`RP-0105`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/readiness.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_READINESS_GO-ClassifyRemediationReadiness-L18`
- **Regression Risk:** `MEDIUM`

Function "ClassifyRemediationReadiness" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 108. [MEDIUM] Excessive cyclomatic complexity (M=21) (`RP-0106`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/components/AdminPortal.jsx`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_COMPONENTS_ADMINPORTAL_JSX-<anonymous@699>-L699`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@699>" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src/components
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 109. [MEDIUM] Excessive cyclomatic complexity (M=21) (`RP-0107`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-parseArgs-L14`
- **Regression Risk:** `MEDIUM`

Function "parseArgs" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 110. [MEDIUM] Excessive cyclomatic complexity (M=22) (`RP-0108`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-assertOwnedResource-L337`
- **Regression Risk:** `MEDIUM`

Function "assertOwnedResource" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 111. [MEDIUM] Excessive cyclomatic complexity (M=23) (`RP-0109`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/archive-api.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_ARCHIVE_API_JS-isValidUpload-L49`
- **Regression Risk:** `MEDIUM`

Function "isValidUpload" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 112. [MEDIUM] Excessive cyclomatic complexity (M=23) (`RP-0110`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ORBITCONTROLS_JS-handleKeyDown-L546`
- **Regression Risk:** `MEDIUM`

Function "handleKeyDown" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 113. [MEDIUM] Excessive cyclomatic complexity (M=23) (`RP-0111`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `source-stager/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SOURCE_STAGER_MAIN_GO-isPublicAddress-L320`
- **Regression Risk:** `MEDIUM`

Function "isPublicAddress" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./source-stager
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 114. [MEDIUM] Excessive cyclomatic complexity (M=26) (`RP-0112`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-accountServiceFromEnvironment-L36`
- **Regression Risk:** `MEDIUM`

Function "accountServiceFromEnvironment" has cyclomatic complexity of 26, requiring at least 26 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 115. [MEDIUM] Excessive cyclomatic complexity (M=27) (`RP-0113`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_MAIN_GO-validateCORSAllowedOrigins-L145`
- **Regression Risk:** `MEDIUM`

Function "validateCORSAllowedOrigins" has cyclomatic complexity of 27, requiring at least 27 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 116. [MEDIUM] Excessive cyclomatic complexity (M=28) (`RP-0114`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-parseSavedRepository-L366`
- **Regression Risk:** `MEDIUM`

Function "parseSavedRepository" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 117. [MEDIUM] Excessive cyclomatic complexity (M=28) (`RP-0115`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-detect_language-L657`
- **Regression Risk:** `MEDIUM`

Function "detect_language" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 118. [MEDIUM] Excessive cyclomatic complexity (M=33) (`RP-0116`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `source-stager/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SOURCE_STAGER_MAIN_GO-validGitHubSegment-L149`
- **Regression Risk:** `MEDIUM`

Function "validGitHubSegment" has cyclomatic complexity of 33, requiring at least 33 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./source-stager
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 119. [MEDIUM] Excessive cyclomatic complexity (M=34) (`RP-0117`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/workspace.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_WORKSPACE_GO-shouldCopyProjectFile-L330`
- **Regression Risk:** `MEDIUM`

Function "shouldCopyProjectFile" has cyclomatic complexity of 34, requiring at least 34 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 120. [MEDIUM] Excessive cyclomatic complexity (M=37) (`RP-0118`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-normalizeResource-L157`
- **Regression Risk:** `MEDIUM`

Function "normalizeResource" has cyclomatic complexity of 37, requiring at least 37 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 121. [MEDIUM] Excessive cyclomatic complexity (M=37) (`RP-0119`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `source-stager/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SOURCE_STAGER_MAIN_GO-ValidateGitHubRepositoryURL-L126`
- **Regression Risk:** `MEDIUM`

Function "ValidateGitHubRepositoryURL" has cyclomatic complexity of 37, requiring at least 37 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./source-stager
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 122. [MEDIUM] Excessive cyclomatic complexity (M=43) (`RP-0120`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-is_allocation-L434`
- **Regression Risk:** `MEDIUM`

Function "is_allocation" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 123. [MEDIUM] Excessive cyclomatic complexity (M=45) (`RP-0121`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/analysis_proxy.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ANALYSIS_PROXY_GO-newAnalysisProxy-L15`
- **Regression Risk:** `MEDIUM`

Function "newAnalysisProxy" has cyclomatic complexity of 45, requiring at least 45 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 124. [MEDIUM] Excessive cyclomatic complexity (M=45) (`RP-0122`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-validateRepositoryURL-L54`
- **Regression Risk:** `MEDIUM`

Function "validateRepositoryURL" has cyclomatic complexity of 45, requiring at least 45 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 125. [MEDIUM] Excessive cyclomatic complexity (M=80) (`RP-0123`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `tests/analysis-workflow.test.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-TESTS_ANALYSIS_WORKFLOW_TEST_MJS-isValidBucket-L217`
- **Regression Risk:** `MEDIUM`

Function "isValidBucket" has cyclomatic complexity of 80, requiring at least 80 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./tests
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 126. [MEDIUM] Excessive cyclomatic complexity (M=97) (`RP-0124`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-is_blocking_io-L400`
- **Regression Risk:** `MEDIUM`

Function "is_blocking_io" has cyclomatic complexity of 97, requiring at least 97 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 127. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0128`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `analysis-authorizer/rest_authorizer_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-ANALYSIS_AUTHORIZER_REST_AUTHORIZER_TEST_GO-TestAuthorizerHandler_DeniesInvalidBearer-L57`
- **Regression Risk:** `LOW`

Function "TestAuthorizerHandler_DeniesInvalidBearer" in analysis-authorizer/rest_authorizer_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./analysis-authorizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 128. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0129`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/energy_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_ENERGY_TEST_GO-TestEnergyCache_ConcurrentAccess-L671`
- **Regression Risk:** `LOW`

Function "TestEnergyCache_ConcurrentAccess" in go-backend/agent/energy_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 129. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0130`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/repo_intel_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_REPO_INTEL_TEST_GO-TestReliabilitySecurityAnalyzer_RedactionAndDetection-L184`
- **Regression Risk:** `LOW`

Function "TestReliabilitySecurityAnalyzer_RedactionAndDetection" in go-backend/agent/repo_intel/repo_intel_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 130. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0131`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-extractFunctionBody-L1068`
- **Regression Risk:** `LOW`

Function "extractFunctionBody" in go-backend/agent/strategy.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 131. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0132`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/deliverables.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_DELIVERABLES_GO-NewJSONStore-L105`
- **Regression Risk:** `LOW`

Function "NewJSONStore" in go-backend/deliverables.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 132. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0133`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/deliverables.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_DELIVERABLES_GO-initDeliverableStore-L155`
- **Regression Risk:** `LOW`

Function "initDeliverableStore" in go-backend/deliverables.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 133. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0134`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/graph.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_GRAPH_MJS-<anonymous@312>-L312`
- **Regression Risk:** `LOW`

Function "<anonymous@312>" in graph-visualizer/graph.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 134. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0135`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-SCRIPTS_ANALYSIS_CLEANUP_MJS-awsProcessEnvironment-L69`
- **Regression Risk:** `LOW`

Function "awsProcessEnvironment" in scripts/analysis-cleanup.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 135. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0136`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-find_func-L1546`
- **Regression Risk:** `LOW`

Function "find_func" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 136. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0137`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-visitor_allocation_in_loop-L944`
- **Regression Risk:** `LOW`

Function "visitor_allocation_in_loop" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 137. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0138`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-visitor_has_recursive_call-L864`
- **Regression Risk:** `LOW`

Function "visitor_has_recursive_call" in thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 138. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0139`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/relationships.rs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_SRC_RELATIONSHIPS_RS-normalize-L242`
- **Regression Risk:** `LOW`

Function "normalize" in thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 139. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0140`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/crawler.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_CRAWLER_GO-BuildIndex-L79`
- **Regression Risk:** `LOW`

Function "BuildIndex" in thermodynamic-ast-engine/test_samples/crawler.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 140. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0141`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/crawler.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_CRAWLER_GO-RecursiveMerge-L113`
- **Regression Risk:** `LOW`

Function "RecursiveMerge" in thermodynamic-ast-engine/test_samples/crawler.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 141. [MEDIUM] Excessive control flow nesting depth (4) (`RP-0142`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/data_pipeline.py`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_DATA_PIPELINE_PY-recursive_flatten-L58`
- **Regression Risk:** `LOW`

Function "recursive_flatten" in thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 142. [MEDIUM] Excessive control flow nesting depth (5) (`RP-0143`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/energy.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_ENERGY_GO-extractHotspotMetric-L299`
- **Regression Risk:** `LOW`

Function "extractHotspotMetric" in go-backend/agent/energy.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 143. [MEDIUM] Excessive control flow nesting depth (5) (`RP-0144`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/manifests.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-parseCargoToml-L233`
- **Regression Risk:** `LOW`

Function "parseCargoToml" in go-backend/agent/repo_intel/manifests.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 144. [MEDIUM] Excessive control flow nesting depth (5) (`RP-0145`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/routes.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GRAPH_VISUALIZER_ROUTES_MJS-obstacleIndex-L31`
- **Regression Risk:** `LOW`

Function "obstacleIndex" in graph-visualizer/routes.mjs reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 145. [MEDIUM] Excessive control flow nesting depth (5) (`RP-0146`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/data_pipeline.py`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_DATA_PIPELINE_PY-build_feature_matrix-L41`
- **Regression Risk:** `LOW`

Function "build_feature_matrix" in thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 146. [MEDIUM] Excessive control flow nesting depth (5) (`RP-0147`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/test_samples/data_pipeline.py`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-THERMODYNAMIC_AST_ENGINE_TEST_SAMPLES_DATA_PIPELINE_PY-fetch_user_data-L18`
- **Regression Risk:** `LOW`

Function "fetch_user_data" in thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/test_samples
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 147. [MEDIUM] Excessive control flow nesting depth (7) (`RP-0148`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Source Finding IDs:** `RI-STR-HIGH-NESTING-GO_BACKEND_AGENT_STRATEGY_GO-removeElseAfterReturn-L466`
- **Regression Risk:** `LOW`

Function "removeElseAfterReturn" in go-backend/agent/strategy.go reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 148. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0149`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/components/AdminPortal.jsx`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_COMPONENTS_ADMINPORTAL_JSX-handlePublish-L225`
- **Regression Risk:** `MEDIUM`

Function "handlePublish" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src/components
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 149. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0150`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `analysis-authorizer/main_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ANALYSIS_AUTHORIZER_MAIN_TEST_GO-TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer-L99`
- **Regression Risk:** `MEDIUM`

Function "TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./analysis-authorizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 150. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0151`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_GO-callback-L202`
- **Regression Risk:** `MEDIUM`

Function "callback" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 151. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0152`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/archive.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ARCHIVE_GO-Delete-L272`
- **Regression Risk:** `MEDIUM`

Function "Delete" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 152. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0153`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/auth.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AUTH_GO-decodeAndVerify-L129`
- **Regression Risk:** `MEDIUM`

Function "decodeAndVerify" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 153. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0154`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/aws_store_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AWS_STORE_TEST_GO-TestArchiveServiceBeginUploadCreatesPrivatePendingRecord-L26`
- **Regression Risk:** `MEDIUM`

Function "TestArchiveServiceBeginUploadCreatesPrivatePendingRecord" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 154. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0155`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_MAIN_GO-runMigrations-L441`
- **Regression Risk:** `MEDIUM`

Function "runMigrations" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 155. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0156`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/refresh_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_REFRESH_TEST_GO-TestLoginHandlerReportsServerRelativeSessionLifetime-L52`
- **Regression Risk:** `MEDIUM`

Function "TestLoginHandlerReportsServerRelativeSessionLifetime" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 156. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0157`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_evaluator.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_EVALUATOR_GO-PrepareWorkspace-L124`
- **Regression Risk:** `MEDIUM`

Function "PrepareWorkspace" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 157. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0158`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_TEST_GO-TestCIWorkflow_ThreeOutcomesSimulation-L1026`
- **Regression Risk:** `MEDIUM`

Function "TestCIWorkflow_ThreeOutcomesSimulation" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 158. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0159`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/workspace.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_WORKSPACE_GO-validateWorkspaceArchiveEntry-L260`
- **Regression Risk:** `MEDIUM`

Function "validateWorkspaceArchiveEntry" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 159. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0160`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-stopClaimedInstance-L364`
- **Regression Risk:** `MEDIUM`

Function "stopClaimedInstance" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 160. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0161`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/account.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ACCOUNT_MJS-<anonymous@94>-L94`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@94>" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 161. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0162`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/model.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_MODEL_MJS-<anonymous@26>-L26`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@26>" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 162. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0163`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-applyMode-L546`
- **Regression Risk:** `MEDIUM`

Function "applyMode" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 163. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RP-0164`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `source-stager/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SOURCE_STAGER_MAIN_GO-StageRepository-L206`
- **Regression Risk:** `MEDIUM`

Function "StageRepository" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./source-stager
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 164. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0165`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/archive-download.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_ARCHIVE_DOWNLOAD_JS-fetchArchiveDownload-L1`
- **Regression Risk:** `MEDIUM`

Function "fetchArchiveDownload" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 165. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0166`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/accounts_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ACCOUNTS_TEST_GO-beginAccountSignIn-L80`
- **Regression Risk:** `MEDIUM`

Function "beginAccountSignIn" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 166. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0167`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_TEST_GO-TestRemoteBrokerEvaluator_PrepareWorkspace-L525`
- **Regression Risk:** `MEDIUM`

Function "TestRemoteBrokerEvaluator_PrepareWorkspace" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 167. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0168`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/strategy.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_STRATEGY_GO-stmtMutatesIdent-L614`
- **Regression Risk:** `MEDIUM`

Function "stmtMutatesIdent" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 168. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0169`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/broker_workspace.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_BROKER_WORKSPACE_GO-AcquireWorkspace-L137`
- **Regression Risk:** `MEDIUM`

Function "AcquireWorkspace" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 169. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0170`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-Acquire-L155`
- **Regression Risk:** `MEDIUM`

Function "Acquire" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 170. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RP-0171`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-sandboxLifecycleFromEnv-L895`
- **Regression Risk:** `MEDIUM`

Function "sandboxLifecycleFromEnv" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 171. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0172`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/aws_store.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AWS_STORE_GO-validGeneratedObjectKey-L346`
- **Regression Risk:** `MEDIUM`

Function "validGeneratedObjectKey" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 172. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0173`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/local_store.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LOCAL_STORE_GO-read-L385`
- **Regression Risk:** `MEDIUM`

Function "read" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 173. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0174`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/correlation.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_CORRELATION_GO-extractSymbolsFromID-L147`
- **Regression Risk:** `MEDIUM`

Function "extractSymbolsFromID" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 174. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0175`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/repo_intel_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_REPO_INTEL_TEST_GO-TestDiscovery_GitIgnoreAndFilter-L89`
- **Regression Risk:** `MEDIUM`

Function "TestDiscovery_GitIgnoreAndFilter" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 175. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0176`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/broker.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_BROKER_GO-clampLimits-L72`
- **Regression Risk:** `MEDIUM`

Function "clampLimits" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 176. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0177`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-assertStackDeleted-L261`
- **Regression Risk:** `MEDIUM`

Function "assertStackDeleted" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 177. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0178`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-normalizePlan-L221`
- **Regression Risk:** `MEDIUM`

Function "normalizePlan" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 178. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RP-0179`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-is_boolean_operator-L265`
- **Regression Risk:** `MEDIUM`

Function "is_boolean_operator" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 179. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0180`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/tests/graph-model.test.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_TESTS_GRAPH_MODEL_TEST_JS-onBorder-L7`
- **Regression Risk:** `MEDIUM`

Function "onBorder" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/tests
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 180. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0181`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `analysis-authorizer/main.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ANALYSIS_AUTHORIZER_MAIN_GO-VerifyArchiveBearer-L66`
- **Regression Risk:** `MEDIUM`

Function "VerifyArchiveBearer" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./analysis-authorizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 181. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0182`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/archive.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_ARCHIVE_GO-validateUploadMetadata-L437`
- **Regression Risk:** `MEDIUM`

Function "validateUploadMetadata" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 182. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0183`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/main_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_MAIN_TEST_GO-TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation-L345`
- **Regression Risk:** `MEDIUM`

Function "TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 183. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0184`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/broker_evaluator.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_BROKER_EVALUATOR_GO-NewRemoteBrokerEvaluator-L45`
- **Regression Risk:** `MEDIUM`

Function "NewRemoteBrokerEvaluator" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 184. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0185`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/evaluator.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_EVALUATOR_GO-copyProjectFiles-L447`
- **Regression Risk:** `MEDIUM`

Function "copyProjectFiles" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 185. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0186`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-decodeSandboxLifecycleRecord-L465`
- **Regression Risk:** `MEDIUM`

Function "decodeSandboxLifecycleRecord" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 186. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0187`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GRAPH_VISUALIZER_ORBITCONTROLS_JS-onTouchMove-L975`
- **Regression Risk:** `MEDIUM`

Function "onTouchMove" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 187. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0188`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-buildUpstreamBody-L116`
- **Regression Risk:** `MEDIUM`

Function "buildUpstreamBody" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 188. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0189`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-parseRequestBody-L79`
- **Regression Risk:** `MEDIUM`

Function "parseRequestBody" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 189. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RP-0190`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `tests/sandbox-host-validation.test.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-TESTS_SANDBOX_HOST_VALIDATION_TEST_MJS-run-L23`
- **Regression Risk:** `MEDIUM`

Function "run" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./tests
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 190. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0191`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/tests/graph-model.test.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_TESTS_GRAPH_MODEL_TEST_JS-<anonymous@58>-L58`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@58>" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/tests
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 191. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0192`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/aws_store.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AWS_STORE_GO-VerifyUpload-L242`
- **Regression Risk:** `MEDIUM`

Function "VerifyUpload" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 192. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0193`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/lambda.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_LAMBDA_GO-loadSSMSecureString-L77`
- **Regression Risk:** `MEDIUM`

Function "loadSSMSecureString" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 193. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0194`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_SANDBOX_LIFECYCLE_GO-ensureReady-L217`
- **Regression Risk:** `MEDIUM`

Function "ensureReady" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 194. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0195`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-aws-L93`
- **Regression Risk:** `MEDIUM`

Function "aws" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 195. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0196`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-emptyBucket-L488`
- **Regression Risk:** `MEDIUM`

Function "emptyBucket" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 196. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0197`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `serve.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SERVE_JS-<anonymous@76>-L76`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@76>" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./.
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 197. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RP-0198`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/languages.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_LANGUAGES_RS-grammar-L23`
- **Regression Risk:** `MEDIUM`

Function "grammar" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 198. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RP-0199`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `admin-portal/src/archive-api.js`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ADMIN_PORTAL_SRC_ARCHIVE_API_JS-isAuthenticatedSession-L66`
- **Regression Risk:** `MEDIUM`

Function "isAuthenticatedSession" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 199. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RP-0200`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/aws_store_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_AWS_STORE_TEST_GO-TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls-L726`
- **Regression Risk:** `MEDIUM`

Function "TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 200. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RP-0201`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-isPrivateIPv4-L16`
- **Regression Risk:** `MEDIUM`

Function "isPrivateIPv4" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 201. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RP-0202`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-determine_primary_driver-L986`
- **Regression Risk:** `MEDIUM`

Function "determine_primary_driver" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 202. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RP-0203`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `analysis-authorizer/rest_authorizer_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-ANALYSIS_AUTHORIZER_REST_AUTHORIZER_TEST_GO-TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer-L13`
- **Regression Risk:** `MEDIUM`

Function "TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./analysis-authorizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 203. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RP-0204`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `api/refresh_test.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-API_REFRESH_TEST_GO-TestRefreshHandlerRenewsActiveBearerToken-L14`
- **Regression Risk:** `MEDIUM`

Function "TestRefreshHandlerRenewsActiveBearerToken" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 204. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RP-0205`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/calibration.go`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_CALIBRATION_GO-CalibrateConfidence-L20`
- **Regression Risk:** `MEDIUM`

Function "CalibrateConfidence" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 205. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RP-0206`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `infra/analysis/stream-proxy/index.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-INFRA_ANALYSIS_STREAM_PROXY_INDEX_MJS-validateAnalysisBaseURL-L32`
- **Regression Risk:** `MEDIUM`

Function "validateAnalysisBaseURL" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./infra/analysis/stream-proxy
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 206. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RP-0207`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/validate-sandbox-host.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_VALIDATE_SANDBOX_HOST_MJS-validateSandboxHost-L14`
- **Regression Risk:** `MEDIUM`

Function "validateSandboxHost" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 207. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RP-0208`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `scripts/analysis-cleanup.mjs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-SCRIPTS_ANALYSIS_CLEANUP_MJS-<anonymous@197>-L197`
- **Regression Risk:** `MEDIUM`

Function "<anonymous@197>" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./scripts
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 208. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RP-0209`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `44.0`
- **Target Files:** `thermodynamic-ast-engine/src/main.rs`
- **Source Finding IDs:** `RI-STR-HIGH-COMPLEXITY-THERMODYNAMIC_AST_ENGINE_SRC_MAIN_RS-scan_directory-L1079`
- **Regression Risk:** `MEDIUM`

Function "scan_directory" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./thermodynamic-ast-engine/src
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 209. [MEDIUM] Missing go.sum lockfile (`RP-0125`)

- **Category:** `DEPENDENCY` | **Execution Strategy:** `RULE_BASED`
- **Execution Class:** `AUTOMATION_READY` | **Priority Score:** `43.4`
- **Target Files:** `samples/benchmark/go.mod`
- **Source Finding IDs:** `RI-DEP-GOSUM-MISSING-SAMPLES_BENCHMARK`
- **Regression Risk:** `LOW`

Go module at samples/benchmark/go.mod lacks a go.sum lockfile, preventing cryptographic build verification.

**Acceptance Criteria:**
- [ ] All dependencies are pinned to immutable cryptographic digests or explicit versions
- [ ] Lockfile (go.sum, Cargo.lock, or package-lock.json) exists and is verified valid

**Validation Commands:**
```bash
go mod verify
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 210. [LOW] Large function AST node volume (402 nodes) (`RP-0210`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/archive.go`
- **Target Symbols:** BeginUpload
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_ARCHIVE_GO-BeginUpload-L125`
- **Regression Risk:** `LOW`

Function "BeginUpload" contains 402 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 211. [LOW] Large function AST node volume (405 nodes) (`RP-0211`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/evaluator_test.go`
- **Target Symbols:** TestCopyProjectFiles_SanitizesSecretsAndBins
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_EVALUATOR_TEST_GO-TestCopyProjectFiles_SanitizesSecretsAndBins-L402`
- **Regression Risk:** `LOW`

Function "TestCopyProjectFiles_SanitizesSecretsAndBins" contains 405 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 212. [LOW] Large function AST node volume (409 nodes) (`RP-0212`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Target Symbols:** AcquireLease
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SANDBOX_LIFECYCLE_GO-AcquireLease-L516`
- **Regression Risk:** `LOW`

Function "AcquireLease" contains 409 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 213. [LOW] Large function AST node volume (411 nodes) (`RP-0213`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/repo_intel/diagnostic/diagnostic_test.go`
- **Target Symbols:** TestDiagnosticEngine_EndToEnd
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_DIAGNOSTIC_DIAGNOSTIC_TEST_GO-TestDiagnosticEngine_EndToEnd-L195`
- **Regression Risk:** `LOW`

Function "TestDiagnosticEngine_EndToEnd" contains 411 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 214. [LOW] Large function AST node volume (422 nodes) (`RP-0214`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/sandbox_lifecycle.go`
- **Target Symbols:** ReleaseLease
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SANDBOX_LIFECYCLE_GO-ReleaseLease-L623`
- **Regression Risk:** `LOW`

Function "ReleaseLease" contains 422 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 215. [LOW] Large function AST node volume (426 nodes) (`RP-0215`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/energy_test.go`
- **Target Symbols:** TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ENERGY_TEST_GO-TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation-L278`
- **Regression Risk:** `LOW`

Function "TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation" contains 426 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 216. [LOW] Large function AST node volume (434 nodes) (`RP-0216`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/aws_store_test.go`
- **Target Symbols:** TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_AWS_STORE_TEST_GO-TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst-L644`
- **Regression Risk:** `LOW`

Function "TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst" contains 434 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 217. [LOW] Large function AST node volume (440 nodes) (`RP-0217`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/agent_test.go`
- **Target Symbols:** makeTestReport
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_AGENT_TEST_GO-makeTestReport-L15`
- **Regression Risk:** `LOW`

Function "makeTestReport" contains 440 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 218. [LOW] Large function AST node volume (446 nodes) (`RP-0218`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/aws_store_test.go`
- **Target Symbols:** TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_AWS_STORE_TEST_GO-TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects-L800`
- **Regression Risk:** `LOW`

Function "TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects" contains 446 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 219. [LOW] Large function AST node volume (456 nodes) (`RP-0219`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/cmd/diagnostic/main.go`
- **Target Symbols:** main
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_CMD_DIAGNOSTIC_MAIN_GO-main-L13`
- **Regression Risk:** `LOW`

Function "main" contains 456 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/cmd/diagnostic
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 220. [LOW] Large function AST node volume (470 nodes) (`RP-0220`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/main_test.go`
- **Target Symbols:** TestBroker_EvaluationResponseCarriesUpdatedBytesAndError
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_MAIN_TEST_GO-TestBroker_EvaluationResponseCarriesUpdatedBytesAndError-L850`
- **Regression Risk:** `LOW`

Function "TestBroker_EvaluationResponseCarriesUpdatedBytesAndError" contains 470 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 221. [LOW] Large function AST node volume (489 nodes) (`RP-0221`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/repo_intel/manifests.go`
- **Target Symbols:** parseDockerfile
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_MANIFESTS_GO-parseDockerfile-L321`
- **Regression Risk:** `LOW`

Function "parseDockerfile" contains 489 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 222. [LOW] Large function AST node volume (497 nodes) (`RP-0222`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/scan_handler_test.go`
- **Target Symbols:** TestScanRepoHandler_CleansTemporaryWorkspace
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_SCAN_HANDLER_TEST_GO-TestScanRepoHandler_CleansTemporaryWorkspace-L136`
- **Regression Risk:** `LOW`

Function "TestScanRepoHandler_CleansTemporaryWorkspace" contains 497 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 223. [LOW] Large function AST node volume (500 nodes) (`RP-0223`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/aws_store_test.go`
- **Target Symbols:** TestS3ObjectStoreSurfacesEveryAWSOperationFailure
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_AWS_STORE_TEST_GO-TestS3ObjectStoreSurfacesEveryAWSOperationFailure-L835`
- **Regression Risk:** `LOW`

Function "TestS3ObjectStoreSurfacesEveryAWSOperationFailure" contains 500 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 224. [LOW] Large function AST node volume (504 nodes) (`RP-0224`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/deliverables.go`
- **Target Symbols:** uploadDeliverableHandler
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_DELIVERABLES_GO-uploadDeliverableHandler-L249`
- **Regression Risk:** `LOW`

Function "uploadDeliverableHandler" contains 504 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 225. [LOW] Large function AST node volume (518 nodes) (`RP-0225`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/repo_intel/repo_intel_test.go`
- **Target Symbols:** setupMockRepository
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_REPO_INTEL_REPO_INTEL_TEST_GO-setupMockRepository-L16`
- **Regression Risk:** `LOW`

Function "setupMockRepository" contains 518 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent/repo_intel
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 226. [LOW] Large function AST node volume (536 nodes) (`RP-0226`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/main_test.go`
- **Target Symbols:** TestDownloadHandlerStreamsAuthenticatedPrivateS3Object
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_MAIN_TEST_GO-TestDownloadHandlerStreamsAuthenticatedPrivateS3Object-L593`
- **Regression Risk:** `LOW`

Function "TestDownloadHandlerStreamsAuthenticatedPrivateS3Object" contains 536 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 227. [LOW] Large function AST node volume (540 nodes) (`RP-0227`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/accounts_test.go`
- **Target Symbols:** TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_ACCOUNTS_TEST_GO-TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions-L202`
- **Regression Risk:** `LOW`

Function "TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions" contains 540 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 228. [LOW] Large function AST node volume (555 nodes) (`RP-0228`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `api/aws_store_test.go`
- **Target Symbols:** TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-API_AWS_STORE_TEST_GO-TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures-L533`
- **Regression Risk:** `LOW`

Function "TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures" contains 555 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./api
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 229. [LOW] Large function AST node volume (582 nodes) (`RP-0229`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/agent.go`
- **Target Symbols:** WriteAuditMarkdown
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_AGENT_GO-WriteAuditMarkdown-L649`
- **Regression Risk:** `LOW`

Function "WriteAuditMarkdown" contains 582 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 230. [LOW] Large function AST node volume (594 nodes) (`RP-0230`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `graph-visualizer/OrbitControls.js`
- **Target Symbols:** constructor
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GRAPH_VISUALIZER_ORBITCONTROLS_JS-constructor-L21`
- **Regression Risk:** `LOW`

Function "constructor" contains 594 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./graph-visualizer
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 231. [LOW] Large function AST node volume (629 nodes) (`RP-0231`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/annealer_benchmark_test.go`
- **Target Symbols:** TestBenchmark_OptimizationEvidence
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ANNEALER_BENCHMARK_TEST_GO-TestBenchmark_OptimizationEvidence-L56`
- **Regression Risk:** `LOW`

Function "TestBenchmark_OptimizationEvidence" contains 629 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 232. [LOW] Large function AST node volume (632 nodes) (`RP-0232`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/agent.go`
- **Target Symbols:** predictFailures
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_AGENT_GO-predictFailures-L285`
- **Regression Risk:** `LOW`

Function "predictFailures" contains 632 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 233. [LOW] Large function AST node volume (637 nodes) (`RP-0233`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/annealer_benchmark_test.go`
- **Target Symbols:** TestBenchmark_Milestone2_ComparativeStrategyHarness
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_ANNEALER_BENCHMARK_TEST_GO-TestBenchmark_Milestone2_ComparativeStrategyHarness-L171`
- **Regression Risk:** `LOW`

Function "TestBenchmark_Milestone2_ComparativeStrategyHarness" contains 637 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 234. [LOW] Large function AST node volume (717 nodes) (`RP-0234`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/main.go`
- **Target Symbols:** setupEngineWithSnapshots
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_MAIN_GO-setupEngineWithSnapshots-L624`
- **Regression Risk:** `LOW`

Function "setupEngineWithSnapshots" contains 717 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 235. [LOW] Large function AST node volume (739 nodes) (`RP-0235`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `go-backend/agent/agent.go`
- **Target Symbols:** generateDirective
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-GO_BACKEND_AGENT_AGENT_GO-generateDirective-L436`
- **Regression Risk:** `LOW`

Function "generateDirective" contains 739 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./go-backend/agent
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 236. [LOW] Large function AST node volume (924 nodes) (`RP-0236`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `admin-portal/src/components/EvolutionLive.jsx`
- **Target Symbols:** EvolutionLive
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-ADMIN_PORTAL_SRC_COMPONENTS_EVOLUTIONLIVE_JSX-EvolutionLive-L9`
- **Regression Risk:** `LOW`

Function "EvolutionLive" contains 924 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./admin-portal/src/components
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

### 237. [LOW] Large function AST node volume (977 nodes) (`RP-0237`)

- **Category:** `STRUCTURAL` | **Execution Strategy:** `AST_TRANSFORMATION`
- **Execution Class:** `AUTOMATION_WITH_REVIEW` | **Priority Score:** `29.0`
- **Target Files:** `samples/benchmark/benchmark_target_test.go`
- **Target Symbols:** TestEvaluateRiskEngine
- **Source Finding IDs:** `RI-STR-LARGE-AST-VOLUME-SAMPLES_BENCHMARK_BENCHMARK_TARGET_TEST_GO-TestEvaluateRiskEngine-L8`
- **Regression Risk:** `LOW`

Function "TestEvaluateRiskEngine" contains 977 AST nodes, indicating monolithic implementation.

**Acceptance Criteria:**
- [ ] Target function cyclomatic complexity is reduced below threshold (M < 12)
- [ ] Control flow nesting depth is reduced to <= 3
- [ ] Existing package unit tests pass with zero regressions

**Validation Commands:**
```bash
go test -v ./samples/benchmark
```

**Rollback Conditions:**
- Rollback working tree immediately if build fails, compilation errors occur, or existing tests regress

---

## Planning Limitations & Safety Notes

- Remediation tasks define declarative specifications and validation gates; execution requires Phase 4 approval
- Soft conflicts indicate shared file modifications that must execute sequentially rather than concurrently
- Human approval is mandatory for all secret revocation, cryptographic token rotation, and circular dependency refactorings
- Planning engine does not modify repository files or apply automated code transformations in Phase 3

