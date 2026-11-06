package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// TestApprovalSecurityAdversarialVectors subjects the approval token engine to 7 adversarial attack vectors.
func TestApprovalSecurityAdversarialVectors(t *testing.T) {
	secretKey := []byte("test-super-secret-key-32bytes!!")
	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "approval-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	persistPath := filepath.Join(tmpDir, "consumed_approvals.json")
	registry, err := NewApprovalConsumptionRegistry(persistPath)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}
	authority := NewApprovalAuthority(secretKey, registry)

	baselineCommit := "92f1c721bb6d7128295a0c63e117d2e8b85bd458"
	workspaceID := "ws-adv-test-01"
	taskID := "TASK-ADV-001"
	targetFile := "service/handler.go"

	fileContent := []byte("package service\n\nfunc Handle() {}\n")
	fileHashBytes := sha256.Sum256(fileContent)
	fileHash := hex.EncodeToString(fileHashBytes[:])
	fileHashes := map[string]string{targetFile: fileHash}

	patchDiff := "--- a/service/handler.go\n+++ b/service/handler.go\n@@ -1,3 +1,4 @@\n package service\n+func Close() {}\n func Handle() {}\n"
	patchHashBytes := sha256.Sum256([]byte(patchDiff))
	patchHash := hex.EncodeToString(patchHashBytes[:])

	// Vector 1: Valid Token Execution (Baseline)
	t.Run("Vector1_ValidTokenAccepted", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
		if err != nil {
			t.Fatalf("Expected valid token to be consumed, got: %v", err)
		}
	})

	// Vector 2: Forged Approval Token (Invalid Signature)
	t.Run("Vector2_ForgedSignatureRejected", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		// Tamper with signature
		token.HMACSignature = "deadbeefcafebabe0123456789abcdefdeadbeefcafebabe0123456789abcdef"

		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
		if err == nil || err != ErrTokenForged {
			t.Fatalf("Expected ErrTokenForged, got: %v", err)
		}
	})

	// Vector 3: Expired Approval Token
	t.Run("Vector3_ExpiredTokenRejected", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			-5*time.Second, // Already expired
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
		if err == nil || (!errors.Is(err, ErrTokenExpired) && !strings.Contains(err.Error(), "FAIL_EXPIRED")) {
			t.Fatalf("Expected FAIL_EXPIRED, got: %v", err)
		}
	})

	// Vector 4: Tampered Patch Content
	t.Run("Vector4_TamperedPatchDiffRejected", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		// Present tampered patch diff hash
		tamperedHash := "1111222233334444555566667777888899990000aaaabbbbccccddddeeeeffff"
		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, tamperedHash, workspaceID)
		if err == nil || !errors.Is(err, ErrPatchHashMismatch) {
			t.Fatalf("Expected ErrPatchHashMismatch, got: %v", err)
		}
	})

	// Vector 5: Sequential Replay Attack
	t.Run("Vector5_SequentialReplayRejected", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		// First execution succeeds
		if err := authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID); err != nil {
			t.Fatalf("First execution should succeed, got: %v", err)
		}

		// Second execution with same token must fail
		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
		if err == nil || err != ErrTokenAlreadyConsumed {
			t.Fatalf("Expected ErrTokenAlreadyConsumed, got: %v", err)
		}
	})

	// Vector 6: Concurrent Replay Race (10 goroutines racing on 1 token)
	t.Run("Vector6_ConcurrentReplayRaceCondition", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		numGoroutines := 10
		var wg sync.WaitGroup
		var successCount int32
		var rejectedCount int32

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
				if err == nil {
					atomic.AddInt32(&successCount, 1)
				} else if err == ErrTokenAlreadyConsumed {
					atomic.AddInt32(&rejectedCount, 1)
				}
			}()
		}
		wg.Wait()

		if successCount != 1 {
			t.Fatalf("Expected exactly 1 execution to succeed, got %d (rejected: %d)", successCount, rejectedCount)
		}
		if rejectedCount != int32(numGoroutines-1) {
			t.Fatalf("Expected %d executions to be rejected, got %d", numGoroutines-1, rejectedCount)
		}
	})

	// Vector 7: Restart Replay (Survives Process Restart via Disk Persistence)
	t.Run("Vector7_RestartReplayPersistence", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		// Consume in current process
		if err := authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID); err != nil {
			t.Fatalf("Consume failed: %v", err)
		}

		// Simulate process restart: initialize fresh registry from same disk file
		restartedRegistry, err := NewApprovalConsumptionRegistry(persistPath)
		if err != nil {
			t.Fatalf("Failed to reload registry: %v", err)
		}
		restartedAuthority := NewApprovalAuthority(secretKey, restartedRegistry)

		// Attempt replay after restart
		err = restartedAuthority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, workspaceID)
		if err == nil || err != ErrTokenAlreadyConsumed {
			t.Fatalf("Expected ErrTokenAlreadyConsumed across restart, got: %v", err)
		}
	})

	// Vector 8: Scope Escalation (Wrong Workspace or Wrong File)
	t.Run("Vector8_ScopeEscalationRejected", func(t *testing.T) {
		token, err := authority.IssueToken(
			"security-lead@company.internal",
			"Principal Engineer",
			taskID,
			"TIER_2_CONTROLLED",
			planning.ClassHumanApprovalRequired,
			baselineCommit,
			fileHashes,
			patchHash,
			"APPLY_PATCH",
			workspaceID,
			10*time.Minute,
		)
		if err != nil {
			t.Fatalf("IssueToken failed: %v", err)
		}

		// Attempt execution in foreign workspace
		err = authority.VerifyAndConsumeToken(token, baselineCommit, fileHashes, patchHash, "attacker-workspace-99")
		if err == nil || !errors.Is(err, ErrWorkspaceMismatch) {
			t.Fatalf("Expected ErrWorkspaceMismatch, got: %v", err)
		}
	})
}
