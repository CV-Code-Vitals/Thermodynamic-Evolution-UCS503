package execution

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

var (
	ErrTokenForged          = errors.New("FAIL_AUTH: approval token signature is forged or invalid")
	ErrTokenExpired         = errors.New("FAIL_EXPIRED: approval token has expired")
	ErrTokenAlreadyConsumed = errors.New("FAIL_REPLAY: approval token has already been consumed")
	ErrCorruptLedger        = errors.New("FAIL_INTEGRITY: persistent approval ledger corrupted or unreadable")
	ErrCommitMismatch       = errors.New("FAIL_SCOPE: target commit does not match approval")
	ErrFileHashMismatch     = errors.New("FAIL_SCOPE: target file modified after approval")
	ErrPatchHashMismatch    = errors.New("FAIL_SCOPE: patch diff payload modified after approval")
	ErrWorkspaceMismatch    = errors.New("FAIL_SCOPE: workspace boundary mismatch")
)

// ApprovalToken is an immutable, cryptographically authentic authorization credential.
type ApprovalToken struct {
	TokenID          string                 `json:"token_id"`
	ApproverID       string                 `json:"approver_id"`
	ApproverRole     string                 `json:"approver_role"`
	TaskID           string                 `json:"task_id"`
	RiskTier         string                 `json:"risk_tier"`
	ExecutionClass   planning.ExecutionClass `json:"execution_class"`
	BaselineCommit   string                 `json:"baseline_commit"`
	TargetFileHashes map[string]string      `json:"target_file_hashes"` // relPath -> SHA256
	PatchDiffHash    string                 `json:"patch_diff_hash"`    // SHA256 of unified diff
	AllowedOperation string                 `json:"allowed_operation"`  // "APPLY_PATCH"
	WorkspaceID      string                 `json:"workspace_id"`
	IssuedAt         time.Time              `json:"issued_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	HMACSignature    string                 `json:"hmac_signature"`
}

// CanonicalPayload returns the deterministic bytes used for HMAC signing and verification.
func (t *ApprovalToken) CanonicalPayload() ([]byte, error) {
	clone := *t
	clone.HMACSignature = "" // Strip signature
	return json.Marshal(clone)
}

// ApprovalConsumptionRegistry maintains single-use token consumption state with persistent backing.
type ApprovalConsumptionRegistry struct {
	mu          sync.Mutex
	consumed    map[string]time.Time
	persistPath string
}

// NewApprovalConsumptionRegistry initializes a registry with optional on-disk persistence.
func NewApprovalConsumptionRegistry(persistPath string) (*ApprovalConsumptionRegistry, error) {
	reg := &ApprovalConsumptionRegistry{
		consumed:    make(map[string]time.Time),
		persistPath: persistPath,
	}

	if persistPath != "" {
		if data, err := os.ReadFile(persistPath); err == nil && len(data) > 0 {
			var loaded map[string]time.Time
			if err := json.Unmarshal(data, &loaded); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrCorruptLedger, err)
			}
			reg.consumed = loaded
		}
	}

	return reg, nil
}

// RecoverCorruptLedger safely quarantines a corrupted ledger file and initializes a fresh ledger.
func RecoverCorruptLedger(persistPath string) (*ApprovalConsumptionRegistry, error) {
	if persistPath == "" {
		return NewApprovalConsumptionRegistry("")
	}
	backupPath := persistPath + ".corrupt." + time.Now().UTC().Format("20060102150405")
	if err := os.Rename(persistPath, backupPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to quarantine corrupt ledger: %w", err)
	}
	return NewApprovalConsumptionRegistry(persistPath)
}

// Consume atomically consumes a token ID, preventing concurrent and future replays.
func (r *ApprovalConsumptionRegistry) Consume(tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.consumed[tokenID]; exists {
		return ErrTokenAlreadyConsumed
	}

	r.consumed[tokenID] = time.Now()

	// Persist to disk if path configured
	if r.persistPath != "" {
		if data, err := json.Marshal(r.consumed); err == nil {
			_ = os.WriteFile(r.persistPath, data, 0600)
		}
	}

	return nil
}

// IsConsumed checks if a token ID has been consumed.
func (r *ApprovalConsumptionRegistry) IsConsumed(tokenID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.consumed[tokenID]
	return exists
}

// ApprovalAuthority issues and verifies tamper-proof approval tokens.
type ApprovalAuthority struct {
	secretKey []byte
	registry  *ApprovalConsumptionRegistry
}

// NewApprovalAuthority creates a new authority with the specified secret key and registry.
func NewApprovalAuthority(secretKey []byte, registry *ApprovalConsumptionRegistry) *ApprovalAuthority {
	if len(secretKey) == 0 {
		secretKey = make([]byte, 32)
		_, _ = rand.Read(secretKey)
	}
	return &ApprovalAuthority{
		secretKey: secretKey,
		registry:  registry,
	}
}

// IssueToken signs and issues a new ApprovalToken with bounded TTL.
func (a *ApprovalAuthority) IssueToken(
	approverID string,
	approverRole string,
	taskID string,
	riskTier string,
	execClass planning.ExecutionClass,
	baselineCommit string,
	fileHashes map[string]string,
	patchDiffHash string,
	allowedOp string,
	workspaceID string,
	ttl time.Duration,
) (*ApprovalToken, error) {
	// Generate cryptographically random TokenID
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate token id: %w", err)
	}
	tokenID := hex.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	token := &ApprovalToken{
		TokenID:          tokenID,
		ApproverID:       approverID,
		ApproverRole:     approverRole,
		TaskID:           taskID,
		RiskTier:         riskTier,
		ExecutionClass:   execClass,
		BaselineCommit:   baselineCommit,
		TargetFileHashes: fileHashes,
		PatchDiffHash:    patchDiffHash,
		AllowedOperation: allowedOp,
		WorkspaceID:      workspaceID,
		IssuedAt:         now,
		ExpiresAt:        now.Add(ttl),
	}

	// Sign payload with HMAC-SHA256
	payload, err := token.CanonicalPayload()
	if err != nil {
		return nil, fmt.Errorf("canonicalize token: %w", err)
	}

	mac := hmac.New(sha256.New, a.secretKey)
	mac.Write(payload)
	token.HMACSignature = hex.EncodeToString(mac.Sum(nil))

	return token, nil
}

// VerifyAndConsumeToken performs full cryptographic and scope verification, atomically consuming the token.
func (a *ApprovalAuthority) VerifyAndConsumeToken(
	token *ApprovalToken,
	currentCommit string,
	currentFileHashes map[string]string,
	currentPatchDiffHash string,
	currentWorkspace string,
) error {
	if token == nil {
		return errors.New("FAIL_AUTH: approval token is nil")
	}

	// 1. Authenticity: verify HMAC signature
	payload, err := token.CanonicalPayload()
	if err != nil {
		return fmt.Errorf("%w: canonicalization error: %v", ErrTokenForged, err)
	}

	mac := hmac.New(sha256.New, a.secretKey)
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(token.HMACSignature), []byte(expectedSig)) {
		return ErrTokenForged
	}

	// 2. Expiry: check temporal validity
	now := time.Now().UTC()
	if now.After(token.ExpiresAt) {
		return fmt.Errorf("%w: expired at %s (now: %s)", ErrTokenExpired, token.ExpiresAt.Format(time.RFC3339), now.Format(time.RFC3339))
	}

	// 3. Scope Binding: Target commit
	if token.BaselineCommit != "" && currentCommit != "" && token.BaselineCommit != currentCommit {
		return fmt.Errorf("%w: expected commit %s, got %s", ErrCommitMismatch, token.BaselineCommit, currentCommit)
	}

	// 4. Scope Binding: Target file hashes
	for relPath, expectedHash := range token.TargetFileHashes {
		currentHash, exists := currentFileHashes[relPath]
		if !exists || currentHash != expectedHash {
			return fmt.Errorf("%w: file %s modified on disk after approval (expected %.8s, got %.8s)",
				ErrFileHashMismatch, relPath, expectedHash, currentHash)
		}
	}

	// 5. Scope Binding: Patch diff hash (Tamper protection)
	if token.PatchDiffHash != "" && currentPatchDiffHash != "" && token.PatchDiffHash != currentPatchDiffHash {
		return fmt.Errorf("%w: patch diff modified after approval (expected %.8s, got %.8s)",
			ErrPatchHashMismatch, token.PatchDiffHash, currentPatchDiffHash)
	}

	// 6. Scope Binding: Workspace isolation
	if token.WorkspaceID != "" && currentWorkspace != "" && token.WorkspaceID != currentWorkspace {
		return fmt.Errorf("%w: expected workspace %s, got %s", ErrWorkspaceMismatch, token.WorkspaceID, currentWorkspace)
	}

	// 7. Atomic Replay Prevention
	if a.registry != nil {
		if err := a.registry.Consume(token.TokenID); err != nil {
			return err
		}
	}

	return nil
}
