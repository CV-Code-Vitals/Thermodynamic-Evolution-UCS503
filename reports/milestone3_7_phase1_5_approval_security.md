# Milestone 3.7 — Phase 1.5: Cryptographic Approval-Token Security Specification

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `92f1c721bb6d7128295a0c63e117d2e8b85bd458`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Document Type:** Cryptographic Protocol & Security Verification Specification  

---

## 1. Executive Summary

Milestone 3.7 Phase 0 audited the legacy approval mechanism and identified critical vulnerabilities: approvals were represented by an in-memory boolean map (`approvedIDs map[string]bool`) keyed solely by string `task.TaskID`. This mechanism lacked principal identity, cryptographic authenticity, expiration timestamps, commit/file hash bindings, patch diff integrity, and atomic replay resistance.

Phase 1.5 replaces this vulnerability with a **tamper-proof, cryptographically signed, scope-bound, single-use Approval-Token Protocol** implemented in [`execution/approval.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/approval.go).

### Key Invariants Enforced:
1. **Cryptographic Authenticity:** Tokens are signed using **HMAC-SHA256** over a canonical JSON payload. Any tampering with token attributes immediately invalidates the signature (`ErrTokenForged`).
2. **Temporal Validity:** Tokens carry strict UTC `IssuedAt` and `ExpiresAt` timestamps (typically 15-minute TTL). Expired tokens are rejected at the execution boundary (`ErrTokenExpired`).
3. **Multi-Dimensional Scope Binding:** Authorization is cryptographically bound to:
   - Target git commit SHA (`token.BaselineCommit`).
   - Bitwise SHA-256 digests of all targeted source files (`token.TargetFileHashes`).
   - Bitwise SHA-256 digest of the exact unified patch diff payload (`token.PatchDiffHash`).
   - Ephemeral workspace directory identifier (`token.WorkspaceID`).
4. **Atomic Replay Prevention:** `ApprovalConsumptionRegistry` enforces thread-safe single-use consumption. If a token ID is submitted concurrently or re-submitted after execution, it is rejected (`ErrTokenAlreadyConsumed`).
5. **Replay Persistence Across Restarts:** Consumed token states can be persisted to disk, preventing replay attacks across server restarts.

---

## 2. Schema Specification: `ApprovalToken`

```json
{
  "$schema": "https://cv-code-vitals.internal/schemas/approval-token-v1.json",
  "token_id": "8f3b21c4e91024ab7d983c12f45091a2",
  "approver_id": "secops-lead@cv-vitals.internal",
  "approver_role": "SECURITY_ADMIN",
  "task_id": "RP-0001",
  "risk_tier": "HIGH",
  "execution_class": "HUMAN_APPROVAL_REQUIRED",
  "baseline_commit": "92f1c721bb6d7128295a0c63e117d2e8b85bd458",
  "target_file_hashes": {
    "cmd/runner.go": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  },
  "patch_diff_hash": "a1b2c3d4e5f60718293a4b5c6d7e8f901234567890abcdef1234567890abcdef",
  "allowed_operation": "APPLY_PATCH",
  "workspace_id": "ws_RP-0001_1762490000",
  "issued_at": "2026-11-07T02:00:00Z",
  "expires_at": "2026-11-07T02:15:00Z",
  "hmac_signature": "7d98b1a4c032e54f98124b67c31245089fae12093845dc7189234bfa01928374"
}
```

---

## 3. Defense Against Attack Vectors

| Attack Vector | Attacker Action | Defensive Mechanism | System Response |
| :--- | :--- | :--- | :--- |
| **Token Forgery** | Attacker generates an arbitrary token JSON or modifies `approver_role` to `"ADMIN"`. | `VerifyAndConsumeToken` recalculates HMAC-SHA256 with authority secret key. | Signature mismatch; rejects with `ErrTokenForged`. |
| **Token Expiry** | Attacker captures a valid token and replays it after the 15-minute validity window. | Evaluates `time.Now().After(token.ExpiresAt)`. | Rejects with `ErrTokenExpired`. |
| **Patch Tampering / Substitution** | Operator approves patch $P_1$, but attacker replaces it with malicious patch $P_2$. | Compares `SHA256(diff(P_2))` against `token.PatchDiffHash`. | Hash mismatch; rejects with `ErrPatchHashMismatch`. |
| **Source Race / Stale State** | Target source file is modified on disk after approval was granted. | Compares `SHA256(currentFile)` against `token.TargetFileHashes[f]`. | File digest mismatch; rejects with `ErrFileHashMismatch`. |
| **Token Replay (Sequential)** | Attacker intercepts a valid token and attempts to execute it a second time. | Token ID is looked up in `ApprovalConsumptionRegistry`. | Detected as consumed; rejects with `ErrTokenAlreadyConsumed`. |
| **Token Replay (Concurrent)** | Attacker submits 10 simultaneous HTTP requests with the identical Token ID. | `ApprovalConsumptionRegistry.Consume` is synchronized via mutex locks. | Exactly 1 request succeeds; 9 fail with `ErrTokenAlreadyConsumed`. |
| **Cross-Workspace Reuse** | Token approved for test workspace `ws_A` executed against `ws_B`. | Evaluates `token.WorkspaceID == currentWorkspace`. | Workspace mismatch; rejects with `ErrWorkspaceMismatch`. |

---

## 4. Execution Boundary Verification

In [`execution/eligibility.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/eligibility.go), `CheckTaskEligibilitySecure` enforces these controls at the kernel boundary:
```go
func CheckTaskEligibilitySecure(
    task *planning.RemediationTask,
    token *ApprovalToken,
    authority *ApprovalAuthority,
    currentCommit string,
    currentFileHashes map[string]string,
    currentPatchDiffHash string,
    currentWorkspace string,
) (bool, bool, string) {
    ...
    if isHumanApprovalRequired {
        if token == nil || authority == nil {
            return false, false, "EXEC_REJECTED_UNAPPROVED: Task requires cryptographically signed ApprovalToken"
        }
        if err := authority.VerifyAndConsumeToken(token, currentCommit, currentFileHashes, currentPatchDiffHash, currentWorkspace); err != nil {
            return false, false, fmt.Sprintf("EXEC_REJECTED_UNAPPROVED: %v", err)
        }
        return true, false, ""
    }
    ...
}
```
Client-side flags such as `approved=true` or unauthenticated database booleans are completely disregarded.

