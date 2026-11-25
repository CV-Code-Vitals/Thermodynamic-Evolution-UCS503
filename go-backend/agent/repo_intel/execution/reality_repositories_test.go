package execution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Opt-in source-only integration run. No repository executable or build script runs.
func TestRealityExternalRepositories(t *testing.T) {
	base := os.Getenv("REALITY_REPOSITORIES")
	if base == "" {
		t.Skip("set REALITY_REPOSITORIES to disposable pinned clones")
	}
	root := t.TempDir()
	svc, err := NewMVPService(root, filepath.Join(root, "ledger.json"), NewUnavailableContainerRunner())
	if err != nil {
		t.Fatal(err)
	}
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if err := ValidateWorkspaceID(id); err != nil {
			return "", err
		}
		return filepath.Join(base, id), nil
	})
	type result struct {
		Repository   string      `json:"repository"`
		Before       string      `json:"before_digest"`
		After        string      `json:"after_digest"`
		Targets      []MVPTarget `json:"targets"`
		Error        string      `json:"error,omitempty"`
		DurationMS   int64       `json:"duration_ms"`
		Reproducible bool        `json:"reproducible"`
	}
	results := []result{}
	for _, name := range []string{"mlsc-timetable", "chi", "uuid", "Finger-Counter", "mlsc-timetable"} {
		start := time.Now()
		r := result{Repository: name}
		r.Before, err = ComputeRepoDigest(filepath.Join(base, name))
		if err != nil {
			t.Fatal(err)
		}
		r.Targets, err = svc.ListLiveTargets(name)
		if err != nil {
			r.Error = err.Error()
		}
		for _, target := range r.Targets {
			content, e := os.ReadFile(filepath.Join(base, name, target.TargetFile))
			if e != nil {
				t.Fatal(e)
			}
			lines := strings.Split(string(content), "\n")
			if target.IsDemoTarget || target.WorkspaceID != name || target.LineNumber < 1 || target.LineNumber > len(lines) || lines[target.LineNumber-1] != target.SourceEvidence || target.UpstreamCommit == "" || target.Symbol == "" {
				t.Errorf("ungrounded target: %+v", target)
			}
		}
		again, e := svc.ListLiveTargets(name)
		r.Reproducible = e == nil && reflect.DeepEqual(r.Targets, again)
		if r.Error == "" && !r.Reproducible {
			t.Errorf("nonreproducible scan: %s", name)
		}
		r.After, err = ComputeRepoDigest(filepath.Join(base, name))
		if err != nil {
			t.Fatal(err)
		}
		if r.Before != r.After {
			t.Errorf("source modified: %s", name)
		}
		r.DurationMS = time.Since(start).Milliseconds()
		results = append(results, r)
		t.Logf("%s: %d findings, reproducible=%v error=%s", name, len(r.Targets), r.Reproducible, r.Error)
	}
	if output := os.Getenv("REALITY_OUTPUT"); output != "" {
		b, e := json.MarshalIndent(results, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(output, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
