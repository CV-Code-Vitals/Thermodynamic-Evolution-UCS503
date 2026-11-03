# Milestone 2.8 — Phase 3: Remediation Planning Engine & Dependency-Ordered Task Generation
## Final Implementation, Security Audit & Verification Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Host Environment:** Windows (NTFS)  
**Baseline Commit:** `3aca118` (descendant of `e117de3` and `d6fbb80`)  
**Role:** Independent Release Engineer, Compiler/AST Specialist & Security Auditor  
**Final Milestone Decision:** **GO / READY FOR PHASE 4 AUTONOMOUS REMEDIATION**

---

## 1. Executive Summary

Milestone 2.8 Phase 3 delivers the **Remediation Planning Engine**, converting the validated diagnostic findings of Phase 2 into a deterministic, dependency-aware, conflict-checked, and machine-executable remediation plan.

The planning engine was engineered as an independent modular subsystem under `go-backend/agent/repo_intel/planning/` and validated against both the live physical repository (`E:\SWE`) and a comprehensive 15-scenario controlled fixture test harness. The engine ingested 486 diagnostic findings, analyzed actionability, clustered related issues via root-cause correlations into 237 structured remediation tasks, computed topological execution schedules, classified 466 task conflicts, and generated objective acceptance criteria for every task.

### Key Deliverables Produced:
1. [`reports/remediation_plan.json`](file:///E:/SWE/reports/remediation_plan.json) (1.1 MB machine-readable execution plan for Phase 4)
2. [`reports/remediation_plan.md`](file:///E:/SWE/reports/remediation_plan.md) (352 KB reviewable technical execution schedule)
3. [`reports/remediation_dependency_graph.json`](file:///E:/SWE/reports/remediation_dependency_graph.json) (4.0 KB directed acyclic task dependency graph)
4. [`reports/remediation_conflict_report.json`](file:///E:/SWE/reports/remediation_conflict_report.json) (208 KB conflict classification report)
5. [`reports/remediation_planning_evaluation.md`](file:///E:/SWE/reports/remediation_planning_evaluation.md) (15-fixture controlled evaluation benchmark)
6. [`security_audit/MILESTONE2_8_PHASE3_INPUT_AUDIT.md`](file:///E:/SWE/security_audit/MILESTONE2_8_PHASE3_INPUT_AUDIT.md) (Baseline input audit & discrepancy proof)
7. [`security_audit/MILESTONE2_8_PHASE3_IMPLEMENTATION_REPORT.md`](file:///E:/SWE/security_audit/MILESTONE2_8_PHASE3_IMPLEMENTATION_REPORT.md) (This report)

---

## 2. Architecture & Modular Decomposition

The planning subsystem is implemented in [`go-backend/agent/repo_intel/planning/`](file:///E:/SWE/go-backend/agent/repo_intel/planning/):

```
go-backend/agent/repo_intel/planning/
├── types.go           # Master remediation plan contracts (RemediationPlan, RemediationTask, Conflicts)
├── ingest.go          # Module A: Report ingestion, SHA256 hashing, path safety, commit provenance
├── opportunity.go     # Module B: Actionability analysis & qualification filtering
├── grouping.go        # Module D: Root-cause clustering & parent-child finding aggregation
├── strategy.go        # Module H: Conservative execution strategy selector & approval gates
├── criteria.go        # Module I: Measurable acceptance criteria & rollback generator
├── priority.go        # Module G: Deterministic priority ranking & numeric scoring formula
├── task.go            # Module C: Structured remediation task factory
├── dependency.go      # Module E: Directed dependency graph builder & Kahn topological sorter
├── conflict.go        # Module F: Hard, soft, and coordination conflict detector
├── planner.go         # Master pipeline orchestrator coordinating Modules A through I
├── reporter.go        # Module J & K: Machine JSON contract & reviewable Markdown generator
├── evaluator.go       # 15-scenario controlled evaluation benchmark harness
└── planning_test.go   # Comprehensive unit, negative, and fixture verification tests
```

### Module Breakdown & Core Guarantees:

| Module | File | Core Capabilities & Safety Invariants |
|---|---|---|
| **Module A: Ingestion & Provenance** | [`ingest.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/ingest.go) | Ingests Phase 2 diagnostic report; verifies schema version (`1.0`); computes SHA256 digest; verifies baseline commit match; rejects paths escaping repository root (`../../`); halts on malformed metadata. |
| **Module B: Opportunity Analysis** | [`opportunity.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/opportunity.go) | Evaluates finding validity, confidence, severity, and readiness. Filters out stale or invalid findings; defers low-confidence heuristics to investigation; gates high-risk items to human approval. |
| **Module C & D: Task Factory & Grouping** | [`task.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/task.go), [`grouping.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/grouping.go) | Consolidates duplicate and parent-child structural findings on the same symbol into unified tasks; preserves complete finding ID provenance. |
| **Module E: Dependency Resolution** | [`dependency.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/dependency.go) | Builds directed task dependency graph (`A -> B`); enforces prerequisite ordering (dependency/lockfile tasks precede application code refactoring; cycle breaking precedes function edits); detects task cycles and blocks affected groups without arbitrary ordering. |
| **Module F: Conflict Detection** | [`conflict.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/conflict.go) | Detects `HARD_CONFLICT` (concurrent edits to the same function symbol), `COORDINATION_REQUIRED` (shared critical configs `Dockerfile`, `go.mod`), and `SOFT_CONFLICT` (same file, distinct functions). |
| **Module G: Priority Engine** | [`priority.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/priority.go) | Computes deterministic priority scores based on severity (CRITICAL: 60, HIGH: 40, MEDIUM: 25, LOW: 10) + confidence points (up to 20); sorts tasks reproducibly with tie-breaker `TaskID` ascending. |
| **Module H: Strategy Selection** | [`strategy.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/strategy.go) | Assigns conservative execution strategies (`RULE_BASED`, `AST_TRANSFORMATION`, `LLM_ASSISTED`, `HUMAN_GUIDED`, `INVESTIGATION_ONLY`). Mandates human approval for secrets, credentials, and circular dependencies. |
| **Module I: Acceptance Criteria** | [`criteria.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/criteria.go) | Generates objective, testable completion conditions, concrete test commands (`go test`, `docker build`), and automated rollback triggers. |
| **Module J & K: Plan Reporting** | [`reporter.go`](file:///E:/SWE/go-backend/agent/repo_intel/planning/reporter.go) | Emits machine JSON contracts, dependency DAGs, conflict reports, and reviewable Markdown execution schedules. |

---

## 3. Test Verification & Evidence

All unit, negative, and fixture tests passed with 100% success.

### A. Subsystem Unit & Negative Suite (`planning`)
Executed via `go test -v ./agent/repo_intel/planning/...`:

```text
=== RUN   TestIngester_PathTraversalRejection
--- PASS: TestIngester_PathTraversalRejection (0.00s)
=== RUN   TestOpportunity_ActionabilityFilter
--- PASS: TestOpportunity_ActionabilityFilter (0.00s)
=== RUN   TestTaskGrouper_CorrelationConsolidation
--- PASS: TestTaskGrouper_CorrelationConsolidation (0.00s)
=== RUN   TestConflictDetector_HardAndSoftConflicts
--- PASS: TestConflictDetector_HardAndSoftConflicts (0.00s)
=== RUN   TestStrategySelector_ConservativeAssignment
--- PASS: TestStrategySelector_ConservativeAssignment (0.00s)
=== RUN   TestRemediationPlanner_EndToEnd
--- PASS: TestRemediationPlanner_EndToEnd (0.01s)
=== RUN   TestPlanningEvaluationHarness
--- PASS: TestPlanningEvaluationHarness (0.00s)
PASS
ok  	github.com/saksham/self-healing-pipeline/agent/repo_intel/planning	0.968s
```
**Result:** 7 / 7 tests PASS (100% success rate).

### B. Full Go Backend Suite Regression Check
Executed via `go test ./...` across all backend packages:
```text
ok  	github.com/saksham/self-healing-pipeline                              1.741s
ok  	github.com/saksham/self-healing-pipeline/agent                        (cached)
ok  	github.com/saksham/self-healing-pipeline/agent/repo_intel             (cached)
ok  	github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic  (cached)
ok  	github.com/saksham/self-healing-pipeline/agent/repo_intel/planning    0.578s
```
**Result:** 100% PASS across all existing and new packages without regression.

### C. Native Rust AST Engine Test Suite
Executed via `cargo test` in `thermodynamic-ast-engine`:
```text
test result: ok. 12 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.02s
```
**Result:** 12 / 12 tests PASS.

---

## 4. Controlled 15-Fixture Planning Quality Evaluation Benchmark

The planning engine was benchmarked against the 15 required planning evaluation scenarios in [`reports/remediation_planning_evaluation.md`](file:///E:/SWE/reports/remediation_planning_evaluation.md):

| Metric | Target | Measured Score | Status |
|---|---|---|---|
| **Planning Success Rate** | $100.0\%$ | **100.00% (15 / 15)** | **PASS** |
| **Task Generation Fidelity** | $15 / 15$ | **15 / 15 (100.0%)** | **PASS** |
| **Conflict Detection Accuracy** | $15 / 15$ | **15 / 15 (100.0%)** | **PASS** |
| **Human Approval Enforcement** | $15 / 15$ | **15 / 15 (100.0%)** | **PASS** |

### Scenario Outcome Matrix:
1. **Single Automation-Ready Finding:** Generated 1 `RULE_BASED` task without approval gate — **PASS**
2. **Multiple Findings Sharing One Root Cause:** Parent-child complexity findings clustered into 1 unified task — **PASS**
3. **Confirmed Architectural Dependency Cycle:** Generated `HUMAN_GUIDED` task requiring human approval — **PASS**
4. **Independent Tasks Across Distinct Files:** Generated 2 independent tasks with 0 conflicts — **PASS**
5. **Hard-Conflicting Tasks (Overlapping Symbol):** Flagged `HARD_CONFLICT` on same function in file — **PASS**
6. **Soft-Conflicting Tasks (Same File, Distinct Symbols):** Flagged `SOFT_CONFLICT` requiring sequential execution — **PASS**
7. **Stale Report Lineage Check:** Detected commit divergence and noted revalidation rationale — **PASS**
8. **Invalid Repository Provenance (Path Escaping Root):** Rejection enforced with `ErrPathEscapeDetected` — **PASS**
9. **Invalid Finding Filtered Out:** Findings marked `INVALID` rejected with 0 tasks generated — **PASS**
10. **Human-Approval Required (Hardcoded Secret):** Enforced `HUMAN_APPROVAL_REQUIRED` status — **PASS**
11. **Investigation-Only Heuristic Finding:** Heuristic findings deferred; 0 code mutation tasks generated — **PASS**
12. **Task Dependency Cycle:** Kahn algorithm identified blocked tasks and halted circular order — **PASS**
13. **Duplicate Report ID Rejection:** Handled gracefully with deterministic output — **PASS**
14. **Finding Whose Source File Was Deleted:** Rejected as stale by opportunity analyzer — **PASS**
15. **Task Without Specific Criteria:** Provided concrete build and regression test criteria fallback — **PASS**

---

## 5. Live Repository Remediation Plan Assessment (`E:\SWE`)

The Remediation Planning Engine executed against the complete diagnostic report [`reports/diagnostic_report.json`](file:///E:/SWE/reports/diagnostic_report.json):

- **Diagnostic Findings Ingested:** 486
- **Actionable Findings Qualified:** 367 (75.5%)
- **Deferred / Informational Observations:** 119 (24.5%) — heuristic taint patterns deferred to investigation
- **Remediation Tasks Generated:** **237 structured tasks**
- **Tasks by Execution Class:**
  - `AUTOMATION_READY`: **1 task** (container Dockerfile hygiene)
  - `AUTOMATION_WITH_REVIEW`: **235 tasks** (AST cyclomatic complexity decomposition and guard clause refactorings)
  - `HUMAN_APPROVAL_REQUIRED`: **1 task** (architectural cycle resolution)
  - `INVESTIGATION_REQUIRED`: **0 tasks** (heuristic warnings withheld from execution schedule)
  - `BLOCKED`: **0 tasks** (topological dependency ordering resolved without unresolvable deadlocks)
- **Conflicts Classified:** **466 conflicts** (0 hard symbol collisions, 466 soft file overlaps and shared config coordinations)

---

## 6. Safety, Isolation & Non-Goals Verification

1. **Read-Only Invariant Maintained:**
   The entire planning engine operates exclusively in read-only mode against inspected repository source files. Zero lines of application source code were altered.
2. **Untrusted Input Protection:**
   The planner treats diagnostic reports as untrusted external inputs. JSON schemas are strictly validated; paths escaping the repository boundary via relative traversals are immediately rejected.
3. **Strings Never Executed:**
   Report field strings are never passed to shell interpreters or execution environments during planning.
4. **Approval Enforcement:**
   All high-impact security findings and architectural cycle tasks strictly enforce `HumanApprovalRequired: true`, preventing autonomous agents from modifying auth or architectural boundaries without operator consent.

---

## 7. Phase 4 Machine-Readable Integration Contract

Phase 4 (Autonomous Remediation Agent) consumes [`reports/remediation_plan.json`](file:///E:/SWE/reports/remediation_plan.json).

```typescript
interface RemediationPlan {
  plan_schema_version: "1.0";
  plan_id: string;
  repository_identity: RepositoryMetadata;
  baseline_commit: string;
  source_diagnostic_report_id: string;
  source_report_hash: string;
  generated_at: string;
  summary: {
    total_findings_considered: number;
    actionable_findings: number;
    total_tasks_generated: number;
    automation_ready_tasks: number;
    automation_with_review_tasks: number;
    human_approval_tasks: number;
    investigation_tasks: number;
    blocked_tasks: number;
    total_conflicts_detected: number;
  };
  tasks: RemediationTask[];
  execution_order: string[]; // Ordered TaskIDs
  dependency_graph: Record<string, string[]>;
  conflicts: TaskConflict[];
  blocked_task_groups: string[];
  planning_limitations: string[];
}

interface RemediationTask {
  task_id: string;
  title: string;
  source_finding_ids: string[];
  category: string;
  description: string;
  target_files: string[];
  target_symbols: string[];
  execution_strategy: "RULE_BASED" | "AST_TRANSFORMATION" | "LLM_ASSISTED" | "HUMAN_GUIDED" | "INVESTIGATION_ONLY";
  execution_class: "AUTOMATION_READY" | "AUTOMATION_WITH_REVIEW" | "HUMAN_APPROVAL_REQUIRED" | "INVESTIGATION_REQUIRED" | "INFORMATIONAL_ONLY";
  priority: "URGENT" | "HIGH" | "MEDIUM" | "LOW";
  priority_score: number;
  priority_rationale: string;
  expected_benefit: string;
  regression_risk: string;
  dependencies: string[]; // Prerequisite TaskIDs
  conflicts: TaskConflict[];
  acceptance_criteria: string[];
  validation_commands: string[];
  rollback_conditions: string[];
  human_approval_required: boolean;
  approval_rationale?: string;
  status: "PLANNED" | "BLOCKED" | "REQUIRES_APPROVAL";
}
```

---

## 8. Capability Audit Classification

In accordance with acceptance criteria, every capability has been evaluated:

| Capability | Status | Verification Evidence |
|---|---|---|
| Ingestion & Schema Validation of Diagnostic Reports | **IMPLEMENTED AND VERIFIED** | Validated report `DR-1793696463` with SHA256 integrity hash |
| Repository & Report Provenance Verification | **IMPLEMENTED AND VERIFIED** | Verified commit lineage and accounted for 433 vs 486 finding delta |
| Stale & Invalid Finding Filtering | **IMPLEMENTED AND VERIFIED** | Verified on negative fixture scenarios 7, 9, and 14 |
| Structured Remediation Task Factory | **IMPLEMENTED AND VERIFIED** | Generated 237 complete tasks conforming to Phase 4 schema |
| Root-Cause Finding Grouping | **IMPLEMENTED AND VERIFIED** | Clustered compound parent-child findings into single cohesive tasks |
| Directed Task Dependency Graph (DAG) | **IMPLEMENTED AND VERIFIED** | Emitted `reports/remediation_dependency_graph.json` |
| Deterministic Topological Execution Ordering | **IMPLEMENTED AND VERIFIED** | Kahn topological sorting scheduled tasks by priority and prerequisites |
| Resource & Semantic Conflict Detection | **IMPLEMENTED AND VERIFIED** | Classified 466 conflicts in `reports/remediation_conflict_report.json` |
| Reproducible Task Prioritization | **IMPLEMENTED AND VERIFIED** | Severity + confidence formula produced stable, deterministic scores |
| Conservative Strategy Selection | **IMPLEMENTED AND VERIFIED** | Strict assignment across rule-based, AST, and human-guided strategies |
| Human Approval Enforcement | **IMPLEMENTED AND VERIFIED** | Verified approval gates for secrets and circular dependency tasks |
| Measurable Acceptance Criteria & Rollback Conditions | **IMPLEMENTED AND VERIFIED** | Populated concrete test commands and rollback triggers on all tasks |
| Versioned Plan Serialization (`1.0`) | **IMPLEMENTED AND VERIFIED** | Emitted `reports/remediation_plan.json` conforming to Phase 4 contract |
| 15-Scenario Controlled Evaluation Benchmark | **IMPLEMENTED AND VERIFIED** | 15 / 15 scenarios passed (100.0% fidelity, accuracy, and approval) |
| Read-Only Planning Execution Invariant | **IMPLEMENTED AND VERIFIED** | Verified zero source code modifications across repository |

---

## 9. Final Milestone Decision

# **DECISION: GO**

The Remediation Planning Engine is complete, mathematically deterministic, dependency-aware, and verified. The outputs provide a machine-executable remediation plan ready for consumption by Phase 4 (Autonomous Remediation Agent).
