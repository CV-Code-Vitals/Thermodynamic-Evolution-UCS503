# Milestone 3.8 Phase 2.1 — Snapshot & Authorization Integrity Architecture

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Status:** COMPLETE & VERIFIED

---

## 1. Objective & Threat Model

Autonomous code repair workflows must guarantee that:
1. Proposed patches correspond to the exact source state analyzed during repository scanning.
2. Authorization decisions are tamper-proof and bound to specific workspaces, commits, file hashes, and approver principals.
3. Concurrency races, source modifications between proposal and execution, or cross-workspace token replay attacks are detected and rejected.

Milestone 3.8 Phase 2.1 introduces end-to-end cryptographic snapshot binding and authorization safeguards across `mvp_scanner.go`, `mvp_service.go`, `approval.go`, and `workspace.go`.

---

## 2. Immutable Workspace Snapshots

During repository discovery (`DiscoverWorkspaceTargets`), the scanner computes an immutable cryptographic manifest of the scanned workspace:

```go
type WorkspaceSnapshot struct {
    SnapshotID   string            `json:"snapshot_id"`    // SNAP-<64-char SHA-256 digest>
    WorkspaceID  string            `json:"workspace_id"`   // Bound workspace identifier
    OwnerID      string            `json:"owner_id"`       // Authenticated scan owner
    SourceCommit string            `json:"source_commit"`  // Source commit SHA
    RootDigest   string            `json:"root_digest"`    // Full SHA-256 tree digest
    FileHashes   map[string]string `json:"file_hashes"`    // Relative path to file SHA-256
    TargetCount  int               `json:"target_count"`   // Count of discovered targets
    CreatedAt    string            `json:"created_at"`     // RFC3339 creation timestamp
    ExpiresAt    string            `json:"expires_at"`     // RFC3339 expiration (60m TTL)
}
```

### Deterministic Integrity Checks:
- **Full SHA-256 Tree Digest:** Calculated across all repository files via `ComputeRepoDigest`. The snapshot ID is formatted as `SNAP-<64-char-hex-digest>`.
- **Fail-Closed on Digest Failure:** If digest computation fails or returns empty, discovery fails immediately with `SNAPSHOT_DIGEST_FAILED`.
- **Target File Hashing:** Each discovered target computes and binds `SourceFileHash`. If the file cannot be read or hashed, discovery fails with `SNAPSHOT_SOURCE_HASH_FAILED`.
- **Expiration Policy:** Snapshots expire after 60 minutes. Proposals or executions referencing expired snapshots fail with `SNAPSHOT_EXPIRED`.

---

## 3. Defense-in-Depth Authorization Rejections

The following strict integrity guards are enforced during repair proposal, approval, and execution:

| Failure Code | Boundary Check | Remediation / Defense |
| :--- | :--- | :--- |
| `MALFORMED_WORKSPACE_ID` | `ValidateWorkspaceID` | Rejects empty IDs, null bytes, `..` traversal, UNC paths (`\\...`, `//...`), and sensitive system directories (`/etc`, `C:\Windows`). |
| `MALFORMED_SNAPSHOT_ID` | `ValidateSnapshotID` | Enforces `SNAP-<64 hex chars>`. Rejects non-hex characters, truncated strings, or directory traversal attempts. |
| `SNAPSHOT_MISSING` | `s.snapshots[SnapshotID]` | Rejects targets referencing unregistered or deleted snapshot IDs. |
| `SNAPSHOT_EXPIRED` | `time.Now().After(snap.ExpiresAt)` | Refuses proposal or execution on stale snapshots exceeding 60-minute window. |
| `SNAPSHOT_WORKSPACE_MISMATCH` | `snap.WorkspaceID == workspaceID` | Prevents applying a snapshot or proposal generated in workspace A to workspace B. |
| `CROSS_USER_WORKSPACE_FORBIDDEN` | `snap.OwnerID == principal` | Approvers cannot authorize repairs for workspaces owned by a different authenticated user. |
| `STALE_WORKSPACE` | `currDigest == snap.RootDigest` | Compares pre-execution repository tree digest against snapshot root digest. If any file changed between scan and repair, execution is rejected. |
| `PROPOSAL_TAMPERED` | `sha256(UnifiedDiff) == DiffHash` | Cryptographically verifies diff content against approval hash before issuing token or applying patch. |
| `FAIL_AUTH` / `ErrWorkspaceMismatch` | `authority.VerifyAndConsumeToken` | Rejects approval token replay across different workspaces, mismatched task IDs, or altered target files. |

---

## 4. Atomic Single-Use Token Consumption

Tokens issued by `ApprovalAuthority.IssueToken` are signed using HMAC-SHA256 over canonicalized payload fields:
- `TokenID`, `ApproverID`, `TaskID`, `RiskTier`, `ExecutionClass`
- `BaselineCommit`, `TargetFileHashes`, `PatchDiffHash`, `WorkspaceID`, `ExpiresAt`

During `ExecuteRepair`:
1. `CheckTaskEligibilitySecure` verifies HMAC authenticity, expiration, target commit, file hashes, and workspace containment.
2. `ApprovalAuthority.VerifyAndConsumeToken` atomically records `TokenID` in `ApprovalConsumptionRegistry`.
3. If an attacker or race condition attempts to execute the same token twice, the second invocation fails immediately with `FAIL_REPLAY: approval token has already been consumed`.

---

## 5. Verification Matrix

| Invariant | Test Vector | Status |
| :--- | :--- | :--- |
| Malformed ID Rejection | `TestAdversarial_6_MalformedWorkspaceAndSnapshotIDsRejected` | PASS |
| Cross-Workspace Token Replay | `TestAdversarial_7_ApprovalTokenReplayAcrossWorkspacesRejected` | PASS |
| Post-Approval File Tampering | `TestAdversarial_8_PostApprovalSourceFileMutationRejected` | PASS |
| Cross-User Workspace Protection | `TestAdversarialSafetyScenarios` & `TestScanToRepairWorkflow` | PASS |
| Single-Use Token Replay Prevention | `TestApprovalSecurityAdversarialVectors/Vector5_SequentialReplayRejected` | PASS |
| Concurrent Replay Race Resistance | `TestApprovalSecurityAdversarialVectors/Vector6_ConcurrentReplayRaceCondition` | PASS |
