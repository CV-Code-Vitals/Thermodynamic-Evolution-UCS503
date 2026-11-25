package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// An immutable, local archive transport; source extraction and scanning still use
// production implementations. No network, remote workflow, or host test execution.
type proofArchive struct{ data []byte }

func (a *proofArchive) GetObject(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(a.data)), nil
}

type publicProofRunner struct {
	unavailable bool
	wrong       bool
	calls       []string
}

func (r *publicProofRunner) CheckAvailable(context.Context) error {
	if r.unavailable {
		return execution.ErrDynamicVerificationUnavailable
	}
	return nil
}
func (r *publicProofRunner) RunCommand(_ context.Context, dir, cmd string) (*execution.ContainerRunResult, error) {
	r.calls = append(r.calls, cmd)
	code := 0
	output := ""
	for _, name := range []string{"TestIndependentDescriptorLeak", "TestApprovedRegression", "TestBehaviorUnchanged"} {
		if strings.Contains(cmd, name) {
			output = "--- PASS: " + name + " (0.00s)\nPASS"
		}
	}
	if strings.Contains(cmd, "TestIndependentDescriptorLeak") && (strings.Contains(dir, "mvp-baseline-") || r.wrong) {
		code = 1
		output = "--- FAIL: TestIndependentDescriptorLeak (0.00s)\nHEALPROOF_DESCRIPTOR_LEAK"
	}
	return &execution.ContainerRunResult{ExitCode: code, Output: output, Evidence: &execution.ContainerExecutionEvidence{ContainerID: "TEST_DOUBLE", ImageDigest: "sha256:TEST_DOUBLE", ExitStatus: code, Command: cmd, WorkspaceDir: dir, IsolationPolicy: "DOCKER_HARDENED_CONTAINER TEST_DOUBLE", CleanupResult: "VERIFIED_REMOVAL: TEST_DOUBLE", NetworkPolicy: "DISABLED (--network none)", RootFSPolicy: "READ_ONLY (--read-only)", MemoryLimitMB: 1024, CPULimit: 2}}, nil
}
func proofSHA(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func proofRequest(t *testing.T, r *gin.Engine, method, path string, body any, want int, out any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trusted-Client-IP", "8.8.8.8")
	req.Header.Set("Authorization", "Bearer proof-operator-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != want {
		t.Fatalf("%s %s: %d, expected %d: %s", method, path, w.Code, want, w.Body.String())
	}
	if out != nil {
		if err = json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatal(err)
		}
	}
}

func proofPublicLifecycle(t *testing.T, runner execution.ContainerRunner, real bool, entries map[string]string, missingPlan bool, incorrectPatch bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	for _, name := range []string{"go.mod", "reader.go", "independent_test.go"} {
		data, err := os.ReadFile(filepath.Join("testdata", "healproof", name))
		if err != nil {
			t.Fatal(err)
		}
		entries["repo/healproof/"+name] = string(data)
	}
	provider := NewSnapshotStore(&proofArchive{makeSnapshotArchive(t, entries)}, filepath.Join(root, "archives"))
	store := &workspaceStore{entries: make(map[string]*evolutionWorkspace)}
	defer func() {
		for _, w := range store.entries {
			if w.timer != nil {
				w.timer.Stop()
			}
			_ = os.RemoveAll(w.path)
		}
	}()
	service, err := execution.NewMVPService(root, filepath.Join(root, "ledger.json"), runner)
	if err != nil {
		t.Fatal(err)
	}
	if incorrectPatch {
		service.Synthesizer = faultyProofSynthesizer{}
	}
	r := gin.New()
	auth := tokenMiddleware("proof-operator-token")
	if real {
		engine := os.Getenv("ENGINE_PATH")
		if engine == "" {
			t.Fatal("ENGINE_PATH must name the prebuilt trusted AST scanner")
		}
		r.POST("/scan", auth, snapshotScanRepoHandler(engine, time.Minute, provider, NewPublicScanController(time.Minute, 1), store))
	} else {
		r.POST("/scan", auth, scanRepoHandlerWithRunner("mock-static-engine", time.Minute, provider, NewPublicScanController(time.Minute, 1), func(_ context.Context, _, _, report string) (string, error) {
			return "", os.WriteFile(report, []byte(`{"files_analyzed":1,"file_reports":[]}`), 0600)
		}, store))
	}
	setupRemediationRoutesWithService(r, auth, store, service)
	var scan struct {
		WorkspaceID string `json:"workspace_id"`
	}
	proofRequest(t, r, "POST", "/scan", map[string]any{"snapshot_id": testSnapshotID, "retain_workspace": true}, 200, &scan)
	ws, err := store.get(scan.WorkspaceID)
	if err != nil {
		t.Fatal(err)
	}
	pre, err := execution.ComputeRepoDigest(ws)
	if err != nil {
		t.Fatal(err)
	}
	var targets struct {
		Targets []execution.MVPTarget `json:"targets"`
	}
	proofRequest(t, r, "GET", "/api/remediation/targets?workspace_id="+scan.WorkspaceID, nil, 200, &targets)
	var chosen execution.MVPTarget
	for _, v := range targets.Targets {
		if v.DefectCategory == "RESOURCE_LEAK" && v.ExpectedAction == "REPAIR" {
			chosen = v
			break
		}
	}
	if chosen.TargetID == "" {
		t.Fatal("public scanner did not discover controlled leak")
	}
	testBytes := []byte(entries["repo/healproof/independent_test.go"])
	if !missingPlan {
		plan := execution.VerificationPlan{ModuleDir: "healproof", Repository: "content-sha256:" + pre, Revision: "sha256:" + pre, RootDigest: pre, SourceHash: chosen.SourceFileHash, Reproducer: execution.VerificationCheck{Package: ".", Test: "TestIndependentDescriptorLeak"}, Regression: []execution.VerificationCheck{{Package: ".", Test: "TestApprovedRegression"}}, Behavioral: execution.VerificationCheck{Package: ".", Test: "TestBehaviorUnchanged"}, TestFiles: map[string]string{"healproof/independent_test.go": proofSHA(testBytes)}, FailureAssertion: "HEALPROOF_DESCRIPTOR_LEAK", BeforeExit: 1, AfterExit: 0, NegativeProperties: "AST_SAFETY_V1", Issuer: "healproof-reviewed-fixture", Provenance: "testdata/healproof independent descriptor-count and behavior assertions; controlled fixture, not natural defect", Limits: execution.DefaultContainerLimits()}
		if err = service.ConfigureTrustedVerificationPlans([]execution.VerificationPlan{plan}); err != nil {
			t.Fatal(err)
		}
	}
	var analysis execution.MVPAnalysisResult
	proofRequest(t, r, "POST", "/api/remediation/analyze", map[string]string{"target_id": chosen.TargetID}, 200, &analysis)
	var proposal execution.MVPProposalResult
	proofRequest(t, r, "POST", "/api/remediation/propose", map[string]string{"analysis_id": analysis.AnalysisID}, 200, &proposal)
	if proposal.Status != "PROPOSED" {
		t.Fatalf("synthesizer abstained: %+v", proposal)
	}
	if !missingPlan && (proposal.VerificationPlan == nil || proposal.VerificationPlanDigest != proposal.VerificationPlan.Digest()) {
		t.Fatal("public proposal missing valid manifest")
	}
	var approval execution.MVPApprovalResult
	proofRequest(t, r, "POST", "/api/remediation/approve", map[string]string{"proposal_id": proposal.ProposalID}, 200, &approval)
	if approval.Signature == "" || approval.VerificationPlanDigest != proposal.VerificationPlanDigest {
		t.Fatal("public approval not bound")
	}
	var result execution.MVPExecutionResult
	unavailable := false
	if mock, ok := runner.(*publicProofRunner); ok {
		unavailable = mock.unavailable
	}
	if missingPlan || unavailable {
		var rejection struct {
			Execution execution.MVPExecutionResult `json:"execution"`
		}
		proofRequest(t, r, "POST", "/api/remediation/execute", map[string]string{"token_id": approval.TokenID}, 422, &rejection)
		result = rejection.Execution
		want := "DYNAMIC_VERIFICATION_UNAVAILABLE"
		if missingPlan {
			want = "REPRODUCER_NOT_AVAILABLE"
		}
		if result.VerificationVerdict != want || result.AppliedInSandbox {
			t.Fatalf("unexpected rejection %+v", result)
		}
	} else {
		proofRequest(t, r, "POST", "/api/remediation/execute", map[string]string{"token_id": approval.TokenID}, 200, &result)
		if real && !incorrectPatch && result.VerificationVerdict != "INDEPENDENTLY_VERIFIED" {
			t.Fatalf("real container proof failed: %+v", result)
		}
		if (!real || incorrectPatch) && result.VerificationVerdict == "INDEPENDENTLY_VERIFIED" {
			t.Fatal("mock certified")
		}
	}
	proofRequest(t, r, "POST", "/api/remediation/execute", map[string]string{"token_id": approval.TokenID}, 422, nil)
	var audit execution.MVPAuditReport
	proofRequest(t, r, "GET", "/api/remediation/audit", nil, 200, &audit)
	if len(audit.AuditRecords) < 2 || audit.AuditRecords[0].FinalStatus != result.VerificationVerdict {
		t.Fatal("terminal verdict missing from audit")
	}
	var disposal struct {
		Operation string `json:"operation"`
		Restored  bool   `json:"revision_restored"`
	}
	proofRequest(t, r, "POST", "/api/remediation/rollback", map[string]string{"execution_id": result.ExecutionID}, 200, &disposal)
	if disposal.Operation != "DISCARD_SANDBOX" || disposal.Restored {
		t.Fatal("dishonest rollback contract")
	}
	if _, err = os.Lstat(result.SandboxDir); !os.IsNotExist(err) {
		t.Fatal("sandbox not deleted")
	}
	post, err := execution.ComputeRepoDigest(ws)
	if err != nil || pre != post {
		t.Fatal("source mutated")
	}
	if output := os.Getenv("HEALPROOF_EVIDENCE_DIR"); real && output != "" {
		if err = os.MkdirAll(output, 0755); err != nil {
			t.Fatal(err)
		}
		b, _ := json.MarshalIndent(map[string]any{"proposal": proposal, "approval": approval, "execution": result, "audit": audit, "source_pre": pre, "source_post": post, "disposal": disposal}, "", "  ")
		if err = os.WriteFile(filepath.Join(output, fmt.Sprintf("public-docker-trace-incorrect-%t.json", incorrectPatch)), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
func TestHealproofPublicContract(t *testing.T) {
	t.Log("TESTED_WITH_MOCK -- NOT_DYNAMICALLY_VERIFIED")
	for _, mode := range []string{"wired", "wrong_patch", "unavailable", "missing_plan"} {
		t.Run(mode, func(t *testing.T) {
			proofPublicLifecycle(t, &publicProofRunner{unavailable: mode == "unavailable", wrong: mode == "wrong_patch"}, false, map[string]string{}, mode == "missing_plan", mode == "wrong_patch")
		})
	}
}

// Requires an already-fetched pinned checkout and a trusted prebuilt AST scanner.
// It never clones, downloads dependencies, executes repository code on the host,
// or silently skips when explicitly opted in but prerequisites are missing.
func TestHealproofDockerPublicLifecycle(t *testing.T) {
	if os.Getenv("RUN_HEALPROOF_DOCKER") != "1" {
		t.Skip("DYNAMIC GATE BLOCKED / NOT RUN: opt in on Docker host with RUN_HEALPROOF_DOCKER=1")
	}
	checkout := os.Getenv("HEALPROOF_PINNED_REPO")
	if checkout == "" {
		t.Fatal("HEALPROOF_PINNED_REPO required")
	}
	const revision = "cbd082eafd24b290391145cb0399e81c6b41303c"
	git := func(args ...string) string {
		out, err := exec.Command("git", append([]string{"-C", checkout}, args...)...).Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	if got := git("rev-parse", "HEAD"); got != revision {
		t.Fatalf("pin mismatch: %s", got)
	}
	if remote := git("config", "--get", "remote.origin.url"); strings.TrimSuffix(strings.TrimSuffix(remote, "/"), ".git") != "https://github.com/Zylus08/mlsc-timetable" {
		t.Fatalf("repository mismatch: %s", remote)
	}
	if dirty := git("status", "--porcelain"); dirty != "" {
		t.Fatal("pinned source must be clean")
	}
	original, err := execution.ComputeRepoDigest(checkout)
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]string{}
	err = filepath.Walk(checkout, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular source entry")
		}
		rel, err := filepath.Rel(checkout, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries["repo/"+filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	runner := execution.NewDockerContainerRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = runner.CheckAvailable(ctx); err != nil {
		t.Fatalf("required Docker runtime unavailable: %v", err)
	}
	t.Run("positive", func(t *testing.T) { proofPublicLifecycle(t, runner, true, entries, false, false) })
	t.Run("incorrect_patch", func(t *testing.T) { proofPublicLifecycle(t, runner, true, entries, false, true) })
	after, err := execution.ComputeRepoDigest(checkout)
	if err != nil || original != after {
		t.Fatal("pinned original changed")
	}
}

// Deliberate pre-approval synthesis fault for the negative experiment only.
// The normal synthesizer still discovers and generates a candidate. This test
// collaborator substitutes a comment-only patch; normal manifest and human
// authorization gates must then reject its failed dynamic postcondition.
type faultyProofSynthesizer struct{}

func (faultyProofSynthesizer) Synthesize(dir string, task *planning.RemediationTask, finding *repo_intel.Finding) *execution.SynthesisDecision {
	decision := execution.NewAutonomousPatchSynthesizer().Synthesize(dir, task, finding)
	if decision.Patch != nil {
		patch := decision.Patch
		patch.ModifiedSource = patch.OriginalSource + "\n// HEALPROOF incorrect candidate: defect deliberately retained.\n"
		lines := strings.Count(strings.TrimSuffix(patch.OriginalSource, "\n"), "\n") + 1
		patch.UnifiedDiff = fmt.Sprintf("--- a/%s\n+++ b/%s\n@@ -%d,0 +%d,2 @@\n+\n+// HEALPROOF incorrect candidate: defect deliberately retained.\n", patch.TargetFile, patch.TargetFile, lines, lines+1)
		patch.DiffHash = proofSHA([]byte(patch.UnifiedDiff))
	}
	return decision
}
