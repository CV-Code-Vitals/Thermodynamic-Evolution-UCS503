package execution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRealityControlledTrials(t *testing.T) {
	base := os.Getenv("REALITY_TRIALS")
	if base == "" {
		t.Skip("opt-in disposable external-repository mutation trials")
	}
	out := []map[string]interface{}{}
	for _, category := range []string{"RESOURCE_LEAK", "SQL_INJECTION", "UNHANDLED_ERROR", "COMMAND_INJECTION"} {
		root := t.TempDir()
		ws := filepath.Join(base, category)
		svc, err := NewMVPService(root, filepath.Join(root, "ledger.json"))
		if err != nil {
			t.Fatal(err)
		}
		svc.SetWorkspaceResolver(func(id string) (string, error) {
			if id != category {
				return "", os.ErrNotExist
			}
			return ws, nil
		})
		before, err := ComputeRepoDigest(ws)
		if err != nil {
			t.Fatal(err)
		}
		targets, err := svc.ListLiveTargets(category)
		if err != nil {
			t.Fatal(err)
		}
		row := map[string]interface{}{"category": category, "track": "CONTROLLED_INJECTION", "before_digest": before, "targets": targets, "dynamic_verification": "NOT_RUN", "human_approval": "NOT_REQUESTED_DOCKER_UNAVAILABLE"}
		found := false
		for _, trg := range targets {
			if trg.TargetFile != "realityfixture/defect.go" || trg.DefectCategory != category {
				continue
			}
			found = true
			analysis, e := svc.AnalyzeTarget(trg.TargetID)
			if e != nil {
				t.Fatal(e)
			}
			row["analysis"] = analysis
			proposal, e := svc.ProposeRepair(analysis.AnalysisID)
			if e != nil {
				row["proposal_error"] = e.Error()
			} else {
				row["proposal"] = proposal
			}
			if e == nil && proposal.Status == "PROPOSED" {
				if err := os.WriteFile(filepath.Join(ws, "generated-repair.diff"), []byte(proposal.UnifiedDiff), 0600); err != nil {
					t.Fatal(err)
				}
				// This evidence belongs outside the source snapshot.
				if err := os.Rename(filepath.Join(ws, "generated-repair.diff"), filepath.Join(base, category+"-repair.diff")); err != nil {
					t.Fatal(err)
				}
			}
			// Scan/source sensitivity: source removal invalidates the old target.
			source := filepath.Join(ws, trg.TargetFile)
			saved, e := os.ReadFile(source)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.Remove(source); e != nil {
				t.Fatal(e)
			}
			_, e = svc.AnalyzeTarget(trg.TargetID)
			row["deleted_source_rejected"] = e != nil
			removed, e := svc.ListLiveTargets(category)
			if e != nil {
				t.Fatal(e)
			}
			for _, r := range removed {
				if r.TargetFile == trg.TargetFile {
					t.Fatal("removed source still generates findings")
				}
			}
			if e = os.WriteFile(source, saved, 0600); e != nil {
				t.Fatal(e)
			}
			restored, e := svc.ListLiveTargets(category)
			if e != nil {
				t.Fatal(e)
			}
			restoredID := false
			for _, r := range restored {
				if r.TargetID == trg.TargetID {
					restoredID = true
				}
			}
			row["restored_target_id"] = restoredID
			if !restoredID {
				t.Fatal("restoring identical source changed target")
			}
			svc.snapshots[trg.SnapshotID].ExpiresAt = time.Now().Add(-time.Minute).Format(time.RFC3339)
			_, e = svc.AnalyzeTarget(trg.TargetID)
			row["expired_snapshot_rejected"] = e != nil
			if e == nil || !strings.Contains(e.Error(), "EXPIRED") {
				t.Fatal("expired analysis accepted")
			}
			break
		}
		row["discovered_without_target_id"] = found
		after, err := ComputeRepoDigest(ws)
		if err != nil {
			t.Fatal(err)
		}
		row["after_digest"] = after
		if before != after {
			t.Fatal("trial source changed")
		}
		dockerErr := svc.Runner.CheckAvailable(context.Background())
		if dockerErr != nil {
			row["docker_error"] = dockerErr.Error()
		}
		out = append(out, row)
		t.Logf("%s discovered=%v", category, found)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("REALITY_TRIAL_OUTPUT"), b, 0600); err != nil {
		t.Fatal(err)
	}
}
