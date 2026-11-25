package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
)

func TestRemediationHttpWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, nil, execution.NewUnavailableContainerRunner())

	// 1. GET /api/remediation/demo/targets
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/remediation/demo/targets", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/demo/targets returned %d: %s", w1.Code, w1.Body.String())
	}
	var targetsResp struct {
		Targets []struct {
			TargetID string `json:"target_id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &targetsResp); err != nil || len(targetsResp.Targets) == 0 {
		t.Fatalf("Failed to parse targets: %v (len: %d)", err, len(targetsResp.Targets))
	}

	// 2. POST /api/remediation/demo/analyze for primary verified target (INJ-EXT2-RES-02)
	targetID := targetsResp.Targets[0].TargetID
	analyzeBody, _ := json.Marshal(map[string]string{"target_id": targetID})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/remediation/demo/analyze", bytes.NewReader(analyzeBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/demo/analyze returned %d: %s", w2.Code, w2.Body.String())
	}
	var analysisResp struct {
		AnalysisID string `json:"analysis_id"`
		FindingID  string `json:"finding_id"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &analysisResp); err != nil || analysisResp.AnalysisID == "" {
		t.Fatalf("Failed to parse analysis response: %v", err)
	}

	// 3. POST /api/remediation/propose
	propBody, _ := json.Marshal(map[string]string{"analysis_id": analysisResp.AnalysisID})
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/api/remediation/propose", bytes.NewReader(propBody))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/propose returned %d: %s", w3.Code, w3.Body.String())
	}
	var propResp struct {
		ProposalID  string `json:"proposal_id"`
		Status      string `json:"status"`
		UnifiedDiff string `json:"unified_diff"`
	}
	if err := json.Unmarshal(w3.Body.Bytes(), &propResp); err != nil || propResp.Status != "PROPOSED" {
		t.Fatalf("Proposal failed: %v, status: %s", err, propResp.Status)
	}

	// 4. POST /api/remediation/approve
	apprBody, _ := json.Marshal(map[string]string{
		"proposal_id": propResp.ProposalID,
		"principal":   "test-lead@code-vitals.internal",
		"role":        "Reliability Lead",
	})
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/api/remediation/approve", bytes.NewReader(apprBody))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/approve returned %d: %s", w4.Code, w4.Body.String())
	}
	var apprResp struct {
		TokenID   string `json:"token_id"`
		Valid     bool   `json:"valid"`
		Principal string `json:"principal"`
		Role      string `json:"role"`
	}
	if err := json.Unmarshal(w4.Body.Bytes(), &apprResp); err != nil || !apprResp.Valid {
		t.Fatalf("Approval failed: %v", err)
	}
	if apprResp.Principal != "authenticated-analysis-client" || apprResp.Role != "analysis-operator" {
		t.Fatalf("caller supplied approval identity: %+v", apprResp)
	}

	// 5. POST /api/remediation/execute
	execBody, _ := json.Marshal(map[string]string{"token_id": apprResp.TokenID})
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/api/remediation/execute", bytes.NewReader(execBody))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/execute returned %d: %s", w5.Code, w5.Body.String())
	}
	var execResp struct {
		ExecutionID         string `json:"execution_id"`
		AllOraclesPassed    bool   `json:"all_oracles_passed"`
		VerificationVerdict string `json:"verification_verdict"`
	}
	if err := json.Unmarshal(w5.Body.Bytes(), &execResp); err != nil || execResp.AllOraclesPassed || execResp.VerificationVerdict == "INDEPENDENTLY_VERIFIED" {
		t.Fatalf("Missing checks must prevent verification: %v (verdict: %s)", err, execResp.VerificationVerdict)
	}

	// 6. GET /api/remediation/audit
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodGet, "/api/remediation/audit", nil)
	r.ServeHTTP(w6, req6)
	if w6.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/audit returned %d: %s", w6.Code, w6.Body.String())
	}

	// 7. Test Abstention via Analyze & Propose for INJ-EXT2-CMD-01
	absAnalyzeBody, _ := json.Marshal(map[string]string{"target_id": "INJ-EXT2-CMD-01"})
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest(http.MethodPost, "/api/remediation/demo/analyze", bytes.NewReader(absAnalyzeBody))
	req7.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w7, req7)
	if w7.Code != http.StatusOK {
		t.Fatalf("Analyze abstention returned %d", w7.Code)
	}
	var absAnalysis struct {
		AnalysisID string `json:"analysis_id"`
	}
	_ = json.Unmarshal(w7.Body.Bytes(), &absAnalysis)

	absPropBody, _ := json.Marshal(map[string]string{"analysis_id": absAnalysis.AnalysisID})
	w8 := httptest.NewRecorder()
	req8, _ := http.NewRequest(http.MethodPost, "/api/remediation/propose", bytes.NewReader(absPropBody))
	req8.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w8, req8)
	if w8.Code != http.StatusOK {
		t.Fatalf("Propose abstention returned %d", w8.Code)
	}
	var absProp struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(w8.Body.Bytes(), &absProp)
	if absProp.Status != "ABSTAINED" {
		t.Fatalf("Expected ABSTAINED, got %s", absProp.Status)
	}
}

func TestRemediationRoutesRequireToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupRemediationRoutes(r, tokenMiddleware("test-secret"), nil, execution.NewUnavailableContainerRunner())
	for _, prefix := range []string{"/api/remediation", "/remediation"} {
		for _, route := range []struct{ method, path string }{
			{"GET", "/targets"}, {"POST", "/targets"}, {"GET", "/audit"}, {"POST", "/analyze"},
			{"POST", "/propose"}, {"POST", "/approve"}, {"POST", "/execute"},
			{"POST", "/rollback"}, {"POST", "/demo"},
		} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(route.method, prefix+route.path, nil))
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s %s: got %d", route.method, prefix+route.path, w.Code)
			}
		}
	}
}

func TestRemediationDemoRunsRepairCandidateAndAbstention(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, nil, execution.NewUnavailableContainerRunner())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/remediation/demo", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("demo returned %d: %s", w.Code, w.Body.String())
	}
	var result struct {
		Status string `json:"status"`
		Repair struct {
			Target       string `json:"target"`
			Verification string `json:"verification_verdict"`
		} `json:"repair_candidate"`
		Abstention struct {
			Status string `json:"abstain_status"`
		} `json:"conservative_abstention"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "DEMO_COMPLETED" || result.Repair.Target == "" || result.Repair.Verification == "INDEPENDENTLY_VERIFIED" || result.Abstention.Status != "ABSTAINED" {
		t.Fatalf("unexpected demo result: %+v", result)
	}
}

func TestRemediationHttpWorkflow_WithScannedWorkspace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newWorkspaceStore()
	wsID, wsDir, err := store.create()
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	store.finishScan(wsID, true)

	// Write Go source file to workspace
	srcFile := filepath.Join(wsDir, "service", "reader.go")
	_ = os.MkdirAll(filepath.Dir(srcFile), 0755)
	_ = os.WriteFile(filepath.Join(wsDir, "go.mod"), []byte("module testservice\n\ngo 1.21\n"), 0644)
	err = os.WriteFile(srcFile, []byte(`package service

import (
	"io"
	"os"
)

func LoadConfig(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}
`), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, store, execution.NewUnavailableContainerRunner())

	// 1. Query targets with query parameter ?workspace_id=...
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/remediation/targets?workspace_id="+wsID, nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/targets returned %d", w1.Code)
	}
	var targetsResp struct {
		Targets []struct {
			TargetID       string `json:"target_id"`
			Stratum        string `json:"stratum"`
			TargetFile     string `json:"target_file"`
			DefectCategory string `json:"defect_category"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &targetsResp); err != nil {
		t.Fatalf("Failed to parse targets: %v", err)
	}

	var scannedTargetID string
	for _, trg := range targetsResp.Targets {
		if trg.Stratum == "SCANNED_REPOSITORY" && strings.Contains(trg.TargetFile, "reader.go") {
			scannedTargetID = trg.TargetID
			break
		}
	}
	if scannedTargetID == "" {
		t.Fatalf("Expected scanned repository target for reader.go, got %d targets", len(targetsResp.Targets))
	}

	// 2. Query targets via POST with JSON body
	bodyJSON, _ := json.Marshal(map[string]string{"workspace_id": wsID})
	w1b := httptest.NewRecorder()
	req1b, _ := http.NewRequest(http.MethodPost, "/api/remediation/targets", bytes.NewReader(bodyJSON))
	req1b.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1b, req1b)
	if w1b.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/targets returned %d", w1b.Code)
	}

	// 3. Analyze scanned target
	anBody, _ := json.Marshal(map[string]string{"target_id": scannedTargetID})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(anBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/analyze returned %d: %s", w2.Code, w2.Body.String())
	}
	var anResp struct {
		AnalysisID           string  `json:"analysis_id"`
		AutomaticEligibility bool    `json:"automatic_eligibility"`
		UncertaintyScore     float64 `json:"uncertainty_score"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &anResp); err != nil || anResp.AnalysisID == "" {
		t.Fatalf("Failed to parse analysis: %v", err)
	}
	if !anResp.AutomaticEligibility {
		t.Errorf("Expected automatic eligibility for intra-procedural leak")
	}

	// 4. Propose repair
	prBody, _ := json.Marshal(map[string]string{"analysis_id": anResp.AnalysisID})
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/api/remediation/propose", bytes.NewReader(prBody))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/propose returned %d: %s", w3.Code, w3.Body.String())
	}
	var prResp struct {
		ProposalID  string `json:"proposal_id"`
		Status      string `json:"status"`
		UnifiedDiff string `json:"unified_diff"`
	}
	if err := json.Unmarshal(w3.Body.Bytes(), &prResp); err != nil || prResp.Status != "PROPOSED" {
		t.Fatalf("Failed to parse proposal: %v, status: %s", err, prResp.Status)
	}
	if !strings.Contains(prResp.UnifiedDiff, "defer f.Close()") {
		t.Errorf("Expected diff to contain defer f.Close()")
	}

	// 5. Approve repair
	apBody, _ := json.Marshal(map[string]string{
		"proposal_id": prResp.ProposalID,
		"principal":   "qa-sre@code-vitals.internal",
		"role":        "Staff Reliability Engineer",
	})
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/api/remediation/approve", bytes.NewReader(apBody))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/approve returned %d: %s", w4.Code, w4.Body.String())
	}
	var apResp struct {
		TokenID string `json:"token_id"`
		Valid   bool   `json:"valid"`
	}
	if err := json.Unmarshal(w4.Body.Bytes(), &apResp); err != nil || !apResp.Valid {
		t.Fatalf("Failed to parse approval: %v", err)
	}

	// 6. Execute repair
	exBody, _ := json.Marshal(map[string]string{"token_id": apResp.TokenID})
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/api/remediation/execute", bytes.NewReader(exBody))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/execute returned %d: %s", w5.Code, w5.Body.String())
	}
	var exResp struct {
		ExecutionID           string `json:"execution_id"`
		AppliedInSandbox      bool   `json:"applied_in_sandbox"`
		OriginalRepoImmutable bool   `json:"original_repo_immutable"`
		AllOraclesPassed      bool   `json:"all_oracles_passed"`
	}
	if err := json.Unmarshal(w5.Body.Bytes(), &exResp); err != nil {
		t.Fatalf("Failed to parse execution: %v", err)
	}
	if !exResp.AppliedInSandbox || !exResp.OriginalRepoImmutable || exResp.AllOraclesPassed {
		t.Errorf("Unexpected execution outcome: %+v", exResp)
	}
}
