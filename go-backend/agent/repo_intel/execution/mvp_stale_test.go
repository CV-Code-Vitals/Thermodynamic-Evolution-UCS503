package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteRepairRejectsFileChangedAfterApproval(t *testing.T) {
	root := t.TempDir()
	target := DefaultPreconfiguredTargets()[0]
	if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(target.RelativeSource)), 0755); err != nil {
		t.Fatal(err)
	}
	sandbox := t.TempDir()
	path := filepath.Join(sandbox, target.TargetFile)
	initial := []byte("original")
	if err := os.WriteFile(path, initial, 0600); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(initial)
	service, err := NewMVPService(root, filepath.Join(t.TempDir(), "ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	service.proposals["proposal"] = &MVPProposalResult{ProposalID: "proposal", TargetID: target.TargetID, Status: "PROPOSED", FileHash: hex.EncodeToString(hash[:]), TargetAbsPath: path, SandboxDir: sandbox, ModifiedSource: "patched", WorkspaceID: "workspace", ExecutionClass: "HUMAN_APPROVAL_REQUIRED"}
	approval, err := service.ApproveRepair("proposal", "operator", "operator")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = service.ExecuteRepair(approval.TokenID)
	if err == nil || !strings.Contains(err.Error(), "modified on disk") {
		t.Fatalf("expected stale-file rejection, got %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "changed" {
		t.Fatalf("changed file was overwritten: %q %v", data, err)
	}
	if service.Registry.IsConsumed(approval.TokenID) {
		t.Fatal("rejected approval was consumed")
	}
}
