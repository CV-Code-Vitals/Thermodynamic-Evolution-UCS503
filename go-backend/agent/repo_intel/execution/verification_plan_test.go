package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type contractRunner struct {
	mode     string
	commands []string
}

func (r *contractRunner) CheckAvailable(context.Context) error {
	if r.mode == "unavailable" {
		return ErrDynamicVerificationUnavailable
	}
	return nil
}
func (r *contractRunner) RunCommand(_ context.Context, dir, cmd string) (*ContainerRunResult, error) {
	r.commands = append(r.commands, cmd)
	code := 0
	output := ""
	for _, test := range []string{"TestIndependentDescriptorLeak", "TestApprovedRegression", "TestBehaviorUnchanged"} {
		if strings.Contains(cmd, test) {
			output = "--- PASS: " + test + " (0.00s)\nPASS"
		}
	}
	if strings.Contains(dir, "mvp-baseline-") && strings.Contains(cmd, "TestIndependentDescriptorLeak") {
		code = 1
		output = "--- FAIL: TestIndependentDescriptorLeak (0.00s)\nHEALPROOF_DESCRIPTOR_LEAK: 4 -> 68"
		if r.mode == "baseline_pass" {
			code = 0
			output = "--- PASS: TestIndependentDescriptorLeak (0.00s)"
		}
		if r.mode == "compile" {
			output = "[build failed]"
		}
	}
	if !strings.Contains(dir, "mvp-baseline-") && ((r.mode == "wrong_patch" && strings.Contains(cmd, "TestIndependentDescriptorLeak")) || (r.mode == "regression" && strings.Contains(cmd, "TestApprovedRegression")) || (r.mode == "behavior" && strings.Contains(cmd, "TestBehaviorUnchanged"))) {
		code = 1
		output = "FAIL"
	}
	if r.mode == "no_test" && !strings.Contains(dir, "mvp-baseline-") && strings.Contains(cmd, "TestIndependentDescriptorLeak") {
		output = "testing: warning: no tests to run\nPASS"
	}
	e := &ContainerExecutionEvidence{ContainerID: "TEST_DOUBLE", ImageDigest: "sha256:TEST_DOUBLE", ExitStatus: code, Command: cmd, WorkspaceDir: dir, CleanupResult: "VERIFIED_REMOVAL: TEST_DOUBLE", IsolationPolicy: "DOCKER_HARDENED_CONTAINER TEST_DOUBLE", NetworkPolicy: "DISABLED (--network none)", RootFSPolicy: "READ_ONLY (--read-only)", MemoryLimitMB: 1024, CPULimit: 2}
	return &ContainerRunResult{ExitCode: code, Output: output, Evidence: e}, nil
}

func contractFixture(t *testing.T, r ContainerRunner) (*MVPService, *MVPProposalResult, string) {
	t.Helper()
	root := t.TempDir()
	ws := filepath.Join(root, "source")
	if e := CopyDirectory(filepath.Join("..", "..", "..", "testdata", "healproof"), ws); e != nil {
		t.Fatal(e)
	}
	s, e := NewMVPService(root, filepath.Join(root, "ledger.json"), r)
	if e != nil {
		t.Fatal(e)
	}
	s.SetWorkspaceResolver(func(string) (string, error) { return ws, nil })
	targets, e := s.ListLiveTargets("ws-contract")
	if e != nil {
		t.Fatal(e)
	}
	var target MVPTarget
	for _, candidate := range targets {
		if candidate.DefectCategory == "RESOURCE_LEAK" {
			target = candidate
			break
		}
	}
	if target.TargetID == "" {
		t.Fatal("scanner found no leak")
	}
	snap := s.snapshots[target.SnapshotID]
	testData, e := os.ReadFile(filepath.Join(ws, "independent_test.go"))
	if e != nil {
		t.Fatal(e)
	}
	plan := VerificationPlan{Repository: snap.Repository, Revision: snap.SourceCommit, RootDigest: snap.RootDigest, SourceHash: target.SourceFileHash, Reproducer: VerificationCheck{".", "TestIndependentDescriptorLeak"}, Regression: []VerificationCheck{{".", "TestApprovedRegression"}}, Behavioral: VerificationCheck{".", "TestBehaviorUnchanged"}, TestFiles: map[string]string{"independent_test.go": sourceSHA(testData)}, FailureAssertion: "HEALPROOF_DESCRIPTOR_LEAK", BeforeExit: 1, AfterExit: 0, NegativeProperties: "AST_SAFETY_V1", Issuer: "test-operator", Provenance: "independent testdata/healproof contract; TESTED_WITH_MOCK -- NOT_DYNAMICALLY_VERIFIED", Limits: DefaultContainerLimits()}
	if e = s.ConfigureTrustedVerificationPlans([]VerificationPlan{plan}); e != nil {
		t.Fatal(e)
	}
	a, e := s.AnalyzeTarget(target.TargetID)
	if e != nil {
		t.Fatal(e)
	}
	p, e := s.ProposeRepair(a.AnalysisID)
	if e != nil {
		t.Fatal(e)
	}
	if p.Status != "PROPOSED" || p.VerificationPlan == nil {
		t.Fatalf("no planned proposal: %+v", p)
	}
	return s, p, ws
}
func TestVerificationContractControlFlow(t *testing.T) {
	t.Log("TESTED_WITH_MOCK -- NOT_DYNAMICALLY_VERIFIED")
	for _, tc := range []struct{ mode, want string }{{"ok", "EXECUTION_FAILED"}, {"wrong_patch", "EXECUTION_FAILED"}, {"regression", "REGRESSION_FAILED"}, {"behavior", "BEHAVIORAL_INVARIANCE_FAILED"}, {"baseline_pass", "EXECUTION_FAILED"}, {"compile", "EXECUTION_FAILED"}, {"unavailable", "DYNAMIC_VERIFICATION_UNAVAILABLE"}, {"no_test", "EXECUTION_FAILED"}} {
		t.Run(tc.mode, func(t *testing.T) {
			r := &contractRunner{mode: tc.mode}
			s, p, ws := contractFixture(t, r)
			pre, _ := ComputeRepoDigest(ws)
			a, e := s.ApproveRepair(p.ProposalID, MVPAnalysisOperator, "operator")
			if e != nil {
				t.Fatal(e)
			}
			if a.VerificationPlanDigest != p.VerificationPlan.Digest() {
				t.Fatal("approval does not bind plan")
			}
			result, err := s.ExecuteRepair(a.TokenID)
			if result == nil || result.VerificationVerdict != tc.want || result.AllOraclesPassed {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			if tc.mode == "ok" {
				for _, test := range []string{"TestIndependentDescriptorLeak", "TestApprovedRegression", "TestBehaviorUnchanged"} {
					if !strings.Contains(strings.Join(r.commands, "\n"), test) {
						t.Fatalf("missing command %s", test)
					}
				}
				if result.BeforeReproduction == nil || result.BeforeReproduction.CleanupStatus != "VERIFIED_REMOVAL" {
					t.Fatal("baseline disposal not verified")
				}
			}
			if _, err = s.ExecuteRepair(a.TokenID); err == nil {
				t.Fatal("replay accepted")
			}
			post, _ := ComputeRepoDigest(ws)
			if pre != post {
				t.Fatal("original source changed")
			}
			if s.ExportAuditReport().VerifiedRepairsCount != 0 {
				t.Fatal("mock certified")
			}
			if _, err = s.RollbackSandbox(result.ExecutionID); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestVerificationContractSecurity(t *testing.T) {
	for _, kind := range []string{"plan", "command", "test", "patch", "candidate", "source", "workspace", "expiry", "missing", "restart", "redigest"} {
		t.Run(kind, func(t *testing.T) {
			s, p, ws := contractFixture(t, &contractRunner{})
			a, e := s.ApproveRepair(p.ProposalID, MVPAnalysisOperator, "operator")
			if e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "plan":
				p.VerificationPlan.Issuer = "attacker"
			case "command":
				p.VerificationPlan.Reproducer.Test = "TestOther"
			case "test":
				e = os.WriteFile(filepath.Join(p.SandboxDir, "independent_test.go"), []byte("package healproof"), 0600)
			case "patch":
				p.UnifiedDiff += "tamper"
			case "candidate":
				p.ModifiedSource += "// tamper"
			case "source":
				e = os.WriteFile(filepath.Join(ws, "reader.go"), []byte("package healproof"), 0600)
			case "workspace":
				p.WorkspaceID = "ws-other"
			case "expiry":
				s.tokenObjs[a.TokenID].ExpiresAt = time.Now().Add(-time.Hour)
			case "missing":
				p.VerificationPlan = nil
			case "restart":
				s, e = NewMVPService(s.WorkspaceRoot, filepath.Join(s.WorkspaceRoot,"ledger.json"), &contractRunner{})
			case "redigest":
				p.VerificationPlan.Reproducer.Test = "TestOther"
				p.VerificationPlanDigest = p.VerificationPlan.Digest()
			}
			if e != nil {
				t.Fatal(e)
			}
			result, err := s.ExecuteRepair(a.TokenID)
			if err == nil || result == nil || result.VerificationVerdict == "INDEPENDENTLY_VERIFIED" || result.AppliedInSandbox {
				t.Fatalf("tamper accepted: %+v %v", result, err)
			}
		})
	}
}
func TestVerificationCommandPolicy(t *testing.T) {
	for _, c := range []VerificationCheck{{".; echo bad", "TestGood"}, {"../../outside", "TestGood"}, {".", ""}, {".", "TestGood;sh"}, {"-exec=sh", "TestGood"}} {
		if _, e := c.Command(); e == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
}
func TestVerificationPlanFrozen(t *testing.T) {
	s, p, _ := contractFixture(t, &contractRunner{})
	a, e := s.ApproveRepair(p.ProposalID, MVPAnalysisOperator, "operator")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.AttachTrustedVerificationPlan(p.ProposalID, *p.VerificationPlan); e == nil {
		t.Fatal("approved plan replaced")
	}
	if a.Signature == "" {
		t.Fatal("unsigned")
	}
	if fmt.Sprint(s.tokenObjs[a.TokenID].TargetFileHashes) == "" {
		t.Fatal("missing scope")
	}
}
