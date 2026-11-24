package execution

import (
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
	"os"
	"path/filepath"
	"testing"
)

func TestRealityNoRepository(t *testing.T) {
	root := t.TempDir()
	svc, e := NewMVPService(root, filepath.Join(root, "ledger.json"), NewUnavailableContainerRunner())
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"", "missing-workspace"} {
		if got := svc.ListTargets(id); len(got) != 0 {
			t.Errorf("no source but %d targets", len(got))
		}
	}
}
func TestRealityApprovalPersistenceFailure(t *testing.T) {
	root := t.TempDir()
	r, e := NewApprovalConsumptionRegistry(filepath.Join(root, "missing", "ledger.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = r.Consume("token"); e == nil {
		t.Fatal("unpersisted consumption succeeded")
	}
	if e = os.WriteFile(filepath.Join(root, "empty.json"), nil, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = NewApprovalConsumptionRegistry(filepath.Join(root, "empty.json")); e == nil {
		t.Fatal("truncated ledger accepted")
	}
}
func TestRealityUnconfiguredReproducerNeverPasses(t *testing.T) {
	ws := t.TempDir()
	if e := os.WriteFile(filepath.Join(ws, "source.go"), []byte("package p\nfunc F() {}"), 0600); e != nil {
		t.Fatal(e)
	}
	suite := NewIndependentOracleSuite(1, NewUnavailableContainerRunner())
	task := &planning.RemediationTask{}
	patch := &CandidatePatch{TargetFile: "source.go"}
	r := suite.verifyDefectReproducer(ws, task, patch, "")
	if r.Passed || r.Status != OracleStatusNotRun {
		t.Fatalf("missing reproducer passed: %+v", r)
	}
	r = suite.verifySyntaxCompilation(ws, task, patch)
	if r.Passed || r.Status != OracleStatusUnavailable {
		t.Fatalf("missing compiler passed: %+v", r)
	}
}
func TestRealityRepositoryReportPoisoning(t *testing.T) {
	root, ws := t.TempDir(), t.TempDir()
	for name, content := range map[string]string{"source.go": "package p\nfunc Actual() {}", "thermodynamic_report.json": `{"files_analyzed":1,"file_reports":[{"file_path":"source.go","language":"Go","hotspots":[{"function_name":"Invented","cyclomatic_complexity":99,"entropy_score":99}]}]}`} {
		if e := os.WriteFile(filepath.Join(ws, name), []byte(content), 0600); e != nil {
			t.Fatal(e)
		}
	}
	svc, e := NewMVPService(root, filepath.Join(root, "ledger.json"))
	if e != nil {
		t.Fatal(e)
	}
	svc.SetWorkspaceResolver(func(string) (string, error) { return ws, nil })
	targets, e := svc.ListLiveTargets("poisoned")
	if e != nil {
		t.Fatal(e)
	}
	if len(targets) != 0 {
		t.Fatalf("poisoned findings: %+v", targets)
	}
}
func TestRealitySourceFailures(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		wantError    bool
	}{
		{"empty", "", false},
		{"broken-parser", "package broken\nfunc (", true},
		{"non-sql-query", "package p\nimport \"github.com/gin-gonic/gin\"\nfunc F(c *gin.Context) string {return c.Query(\"x\")}", false},
		{"non-error-encode", "package p\nimport \"encoding/hex\"\nfunc F(a,b []byte) {hex.Encode(a,b)}", false},
		{"hash-contract", "package p\nimport \"hash\"\nfunc F(h hash.Hash,b []byte) {h.Write(b)}", false},
		{"clean-closed-resource", "package p\nimport \"os\"\nfunc F(p string) error {f,e:=os.Open(p);if e!=nil{return e};defer f.Close();return nil}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, ws := t.TempDir(), t.TempDir()
			if tc.source != "" {
				if e := os.WriteFile(filepath.Join(ws, "source.go"), []byte(tc.source), 0600); e != nil {
					t.Fatal(e)
				}
			}
			svc, e := NewMVPService(root, filepath.Join(root, "ledger.json"))
			if e != nil {
				t.Fatal(e)
			}
			svc.SetWorkspaceResolver(func(string) (string, error) { return ws, nil })
			targets, e := svc.ListLiveTargets("reality")
			if tc.wantError {
				if e == nil {
					t.Fatal("invalid source accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			if len(targets) != 0 {
				t.Fatalf("false findings: %+v", targets)
			}
		})
	}
}
