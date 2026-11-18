# Milestone 3.6 — Phase 1: Semantic Evidence & Remediation Contract

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `8bea13bb5cbd67f20cf59d2607886fc5a7d1bffe`  
**Branch:** `integration/milestone-3.6-controlled-remediation`  
**Contract Version:** `2.0-semantic`  
**Document Type:** Versioned Interface Specification  

---

## 1. Objective and Invariants

This document establishes the versioned evidence contract bridging **Milestone 3.5 Interprocedural Semantic Intelligence** with the **Autonomous Remediation Planning & Execution Pipeline**.

### Invariants:
1. **Provenance Preservation:** Every finding must record its source commit SHA, target file relative path, and content SHA256 at the moment of analysis.
2. **Uncertainty Propagation & Fail-Closed Gating:** If analysis completeness is partial, or if uncertainty involves concurrent ownership escape (`POSSIBLE_ESCAPE`, `UNKNOWN_CONCURRENT_OWNERSHIP`), the finding must fail closed (`RemediationEligibility = HUMAN_APPROVAL_REQUIRED` or `QUARANTINED`). Autonomous remediation is strictly forbidden.
3. **Semantic Trace Integrity:** Structural alias sets, struct-field access paths, call graph path traces, and higher-order cleanup invocation contracts must be carried through the contract into planning and validation.
4. **Validation Grounding:** Every actionable finding must specify concrete independent validation requirements (compiler check, semantic re-scan, regression tests). A patch is never certified solely by a build passing or an energy score improving.

---

## 2. Schema Specification: `SemanticFindingContract`

```json
{
  "$schema": "https://cv-code-vitals.internal/schemas/semantic-finding-v2.json",
  "schema_version": "2.0-semantic",
  "finding_id": "SEM-SEC-SQL-cmd-runner-L42",
  "rule_id": "SEMANTIC_SQL_INJECTION",
  "source_commit": "8bea13bb5cbd67f20cf59d2607886fc5a7d1bffe",
  "file_sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "target_file": "cmd/runner/query.go",
  "line_start": 42,
  "line_end": 42,
  "category": "SECURITY",
  "severity": "HIGH",
  "confidence": 0.98,
  "analysis_completeness": "COMPLETE",
  "uncertainty_class": "DETERMINISTIC",
  "semantic_trace": {
    "call_path": ["HandleRequest", "ExecuteQuery", "db.Query"],
    "cfg_node_ids": [1, 3, 5],
    "dataflow_hops": 2,
    "access_path": "",
    "wrapper_contract": "",
    "aliases": ["rawQuery", "formattedSQL"],
    "alias_hop_depth": 2,
    "concurrency_state": "PROVABLE_LOCAL"
  },
  "remediation_eligibility": "AUTONOMOUS_ELIGIBLE",
  "human_approval_required": false,
  "approval_rationale": "",
  "validation_requirements": [
    "SEMANTIC_RESCAN",
    "COMPILER_GO_VET",
    "UNIT_TEST_REGRESSION"
  ],
  "counter_evidences": [],
  "title": "Unsanitized Dynamic SQL Construction",
  "description": "SQL query expression is dynamically constructed with unescaped string formatting.",
  "root_cause": "String formatting function fmt.Sprintf used to construct SQL query.",
  "recommended_action": "Use parameterized queries with placeholder syntax.",
  "generated_at": "2026-11-07T00:10:00Z"
}
```

---

## 3. Field Definitions & Behavioral Semantics

### 3.1 Identification & Source Provenance
- `schema_version` (string, required): Explicit version `"2.0-semantic"`.
- `finding_id` (string, required): Deterministic canonical ID formatted as `SEM-<CAT>-<TYPE>-<RelPath>-L<Line>`.
- `rule_id` (string, required): Canonical rule identifier (`SEMANTIC_SQL_INJECTION`, `SEMANTIC_COMMAND_INJECTION`, `SEMANTIC_RESOURCE_LEAK`, `SEMANTIC_UNHANDLED_ERROR`).
- `source_commit` (string, required): Repository HEAD commit SHA at scan time.
- `file_sha256` (string, required): Bitwise SHA256 of the target file content at scan time. Enables immediate detection of stale findings or modified local working copies.

### 3.2 Analysis Completeness & Uncertainty Classification
- `analysis_completeness` (enum):
  - `COMPLETE`: Full CFG, call graph, and dataflow traversal succeeded without truncation.
  - `PARTIAL_CALL_GRAPH_TRUNCATED`: Call graph depth exceeded bounded limits ($k > 4$) or callee body was unavailable in source tree.
  - `INCOMPLETE_DEPENDENCIES`: Unresolved types or external package interfaces prevented full reaching-definition computation.
- `uncertainty_class` (enum):
  - `DETERMINISTIC`: Defect state proven by complete intraprocedural or interprocedural summary propagation.
  - `BOUNDED_UNCERTAINTY`: Defect state inferred under conservative bounded approximations (e.g. `MAY_ALIAS` branches).
  - `POSSIBLE_ESCAPE`: Resource handle passed to asynchronous goroutine or channel; ownership lifetime cannot be statically proven.
  - `UNKNOWN_CONCURRENT_OWNERSHIP`: Cross-thread or synchronization primitive ownership unknown.

**Gating Rule:** Any finding with `uncertainty_class` equal to `POSSIBLE_ESCAPE` or `UNKNOWN_CONCURRENT_OWNERSHIP` must set `remediation_eligibility = HUMAN_APPROVAL_REQUIRED` and `human_approval_required = true`.

### 3.3 Semantic Evidence Trace
- `call_path` ([]string): Interprocedural call sequence from root caller to vulnerability sink.
- `cfg_node_ids` ([]int): CFG basic block sequence demonstrating control flow reaching the sink.
- `dataflow_hops` (int): Number of assignment/dereference steps between taint source / resource acquisition and sink.
- `access_path` (string): Field path (e.g. `h.file`, `req.ctx.conn`) up to depth 3.
- `wrapper_contract` (string): Certified higher-order cleanup invocation contract (`INVOKE_MUST_ONCE`, `INVOKE_CONDITIONAL`, `INVOKE_NEVER`, `INVOKE_ASYNC`, `INVOKE_STORED`).
- `aliases` ([]string): Transitive alias set computed by fixed-point alias analysis.
- `alias_hop_depth` (int): Maximum depth traversed in the alias graph.
- `concurrency_state` (string): `PROVABLE_LOCAL`, `PROVABLE_TRANSFER`, `POSSIBLE_ESCAPE`, or `UNKNOWN_CONCURRENT_OWNERSHIP`.

### 3.4 Remediation Eligibility & Approval Governance
- `remediation_eligibility` (enum):
  - `AUTONOMOUS_ELIGIBLE`: High-confidence, deterministic, local or verified interprocedural defect with deterministic patch strategy.
  - `HUMAN_APPROVAL_REQUIRED`: High-impact security modification, architectural refactoring, or uncertain/concurrent ownership.
  - `QUARANTINED`: Ambiguous finding, incomplete call graph, or unresolved counter-evidence.
  - `BENIGN`: Certified clean lifecycle (safe counter-evidence present).
- `human_approval_required` (bool): If `true`, the isolated executor will fail closed (`EXEC_REJECTED_UNAPPROVED`) unless an explicit human token is supplied.
- `approval_rationale` (string): Mandatory explanation of why human authorization is required.

### 3.5 Independent Validation Requirements
- `validation_requirements` ([]string): List of mandatory post-patch verification gates:
  - `SEMANTIC_RESCAN`: Must re-run the exact semantic analyzer that generated the finding. The finding must not be reproduced.
  - `COMPILER_GO_VET`: `go vet ./...` must exit 0 with zero errors or warnings.
  - `UNIT_TEST_REGRESSION`: Existing package unit tests (`go test -count=1 ./...`) must pass without regression.

---

## 4. Fail-Closed Invariants & Serialization Rules

1. If `file_sha256` does not match the workspace file on disk at patch execution time:
   $$\text{Decision} = \text{REJECTED\_AS\_STALE}$$
2. If `analysis_completeness != "COMPLETE"`:
   $$\text{RemediationEligibility} \le \text{HUMAN\_APPROVAL\_REQUIRED}$$
3. If `concurrency_state \in \{\text{POSSIBLE\_ESCAPE}, \text{UNKNOWN\_CONCURRENT\_OWNERSHIP}\}`:
   $$\text{RemediationEligibility} = \text{HUMAN\_APPROVAL\_REQUIRED}, \quad \text{HumanApprovalRequired} = \text{true}$$
4. An adapter translates `SemanticFindingContract` into `diagnostic.ValidatedFinding` for compatibility with `planning.RemediationPlanner`, preserving all semantic tags in `EvidenceLocation.Description` and `ReadinessRationale`.
