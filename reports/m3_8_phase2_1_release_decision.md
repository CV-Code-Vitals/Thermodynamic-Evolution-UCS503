# Milestone 3.8 Phase 2.1 — Release Candidate Qualification Decision

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Release Target:** Milestone 3.8 Phase 2.1 (Live Repository Execution Safety Correction)  
**Final Release Decision:** **QUALIFIED FOR LOCAL RELEASE CANDIDATE (RC)**

---

## 1. Executive Summary & Authorization Scope

Under Milestone 3.8 Phase 2.1, authorization was granted strictly for local security hardening, execution safety correction, oracle alignment, and adversarial regression testing. Production deployment, upstream pushing, and automatic merges remain **STRICTLY PROHIBITED**.

Following the implementation of container isolation, complete removal of unsandboxed host toolchain execution, cryptographic snapshot binding, honest oracle outcome reporting, and verification across 9 deterministic adversarial vectors, the repository is declared **QUALIFIED** for the Milestone 3.8 Phase 2.1 Release Candidate.

---

## 2. Security & Correctness Audit Scorecard

| Area | Invariant / Requirement | Baseline State | Phase 2.1 Hardened State | Status |
| :--- | :--- | :--- | :--- | :---: |
| **Host Toolchain Execution** | Untrusted code toolchain execution (`go vet`, `go test`, reproducer scripts) | Executed via host `exec.CommandContext` using command whitelist | Completely eliminated on host. Dynamic checks routed strictly through container runtime; host execution forbidden. | **RESOLVED** |
| **Container Isolation** | Hardened container runner specifications | No container runner configured in `IndependentOracleSuite` | `DockerContainerRunner` enforces `--network none`, `--read-only`, tmpfs, non-root `10001:10001`, `--cap-drop ALL`, no-new-privs, memory/CPU quotas. | **ENFORCED** |
| **Fail-Closed Container Boundary** | Behavior when container runtime or daemon is unavailable | Fell back to host execution | Fails closed with `OracleStatusUnavailable` (`DYNAMIC_VERIFICATION_UNAVAILABLE`). No host fallback. | **ENFORCED** |
| **Snapshot Integrity** | Cryptographic binding to workspace and authenticated owner | Basic snapshot identifier without full SHA-256 tree digest | Immutable `WorkspaceSnapshot` bound to `WorkspaceID`, `OwnerID`, full 64-char SHA-256 tree digest, and 60m TTL. | **ENFORCED** |
| **Tamper & Race Protection** | Post-approval file changes or diff alterations | Undetected if modified between approval and execute | Re-computes repository tree digest and verifies diff hashes before execution (`STALE_WORKSPACE`, `PROPOSAL_TAMPERED`). | **ENFORCED** |
| **Cross-User Protection** | Approvers authorizing other users' workspaces | Unchecked | Refuses cross-user authorization (`CROSS_USER_WORKSPACE_FORBIDDEN`). | **ENFORCED** |
| **Honest Oracle Statuses** | Granular reporting of check outcomes | Unrun checks defaulted to `Passed: true` | Explicit statuses: `PASSED`, `FAILED`, `NOT_RUN`, `UNAVAILABLE`. Never equates `NOT_RUN` with `PASSED`. | **ENFORCED** |
| **Honest Verdict Contract** | Distinction between static-only and independently verified | Ambiguous verdict | `INDEPENDENTLY_VERIFIED` requires container dynamic test pass; static-only yields `SYNTAX_ONLY_VERIFIED`. | **ENFORCED** |
| **Adversarial Hardening** | 9 Deterministic adversarial attack vectors | Not fully covered | 9/9 vectors tested and verified fail-closed in `adversarial_execution_test.go`. | **PASSED** |
| **Benchmark Integrity** | Historical ground-truth benchmark metrics | 9/15 injected, 0 natural | Preserved strictly without modification (NAT-03 disqualified). | **PRESERVED** |

---

## 3. Deliverables Inventory

All seven required Phase 2.1 reports have been authored and placed under `reports/`:

1. [Host Execution Audit](file:///E:/SWE/reports/m3_8_phase2_1_host_execution_audit.md): Forensic audit and evidence of unsandboxed host execution removal.
2. [Container Integration](file:///E:/SWE/reports/m3_8_phase2_1_container_integration.md): Architecture of `DockerContainerRunner`, container arguments, dropped capabilities, and runtime limits.
3. [Snapshot Integrity](file:///E:/SWE/reports/m3_8_phase2_1_snapshot_integrity.md): Immutable snapshot binding, SHA-256 tree digests, anti-tamper guards, and atomic token consumption.
4. [Oracle Verdicts](file:///E:/SWE/reports/m3_8_phase2_1_oracle_verdicts.md): Alignment of `PASSED`, `FAILED`, `NOT_RUN`, and `UNAVAILABLE` statuses, verdict contracts, and UI representation.
5. [Adversarial Results](file:///E:/SWE/reports/m3_8_phase2_1_adversarial_results.md): Detailed results for all 9 deterministic adversarial vectors.
6. [Regression Report](file:///E:/SWE/reports/m3_8_phase2_1_regression.md): Full test results across Go backend, Rust AST engine, and React frontend, confirming benchmark preservation.
7. [Release Decision](file:///E:/SWE/reports/m3_8_phase2_1_release_decision.md): Formal qualification scorecard and release candidate determination.

---

## 4. Reaffirmation of Operational Constraints

The qualification of this Release Candidate is subject to the following absolute constraints:
1. **No Production Deployment:** This release candidate is restricted to evaluation and safety qualification environments.
2. **No Upstream Remotes:** No commits or tags shall be pushed to upstream GitHub repositories.
3. **No Automatic Merges:** The branch `release/milestone-3.8-mvp-rc` shall not be merged into `main` without formal sign-off.
4. **Mandatory Human Approval:** No autonomous repair may execute on mutable code without cryptographically signed human authorization.
