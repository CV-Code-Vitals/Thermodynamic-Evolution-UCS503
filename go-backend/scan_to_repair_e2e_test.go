package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestEndToEndScanToRepairWorkflow_LiveRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping live end-to-end repository scan in short mode")
	}

	gin.SetMode(gin.TestMode)
	enginePath := resolveEnginePath()
	engineTimeout := 120 * time.Second
	workspaces := newWorkspaceStore()

	r := gin.New()
	scanH := scanRepoHandler(enginePath, engineTimeout, workspaces)
	r.POST("/api/scan", scanH)
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, workspaces)

	// Step 1: Scan repository with retain_workspace: true
	scanReq := map[string]interface{}{
		"repo_url":         "https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503",
		"retain_workspace": true,
	}
	body, _ := json.Marshal(scanReq)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("POST /api/scan failed with %d: %s", w1.Code, w1.Body.String())
	}

	var scanResp struct {
		WorkspaceID string `json:"workspace_id"`
		Report      struct {
			FilesAnalyzed int `json:"files_analyzed"`
			TotalHotspots int `json:"total_hotspots"`
		} `json:"report"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &scanResp); err != nil {
		t.Fatalf("Failed to parse scan response: %v", err)
	}

	if scanResp.WorkspaceID == "" {
		t.Fatalf("Expected retained workspace_id, got empty")
	}
	t.Logf("Scan completed: WorkspaceID=%s, Files=%d, Hotspots=%d",
		scanResp.WorkspaceID, scanResp.Report.FilesAnalyzed, scanResp.Report.TotalHotspots)

	// Step 2: Fetch remediation targets for the scanned repository
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/remediation/targets?workspace_id="+scanResp.WorkspaceID, nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/targets failed with %d: %s", w2.Code, w2.Body.String())
	}

	var targetsResp struct {
		Targets []struct {
			TargetID       string `json:"target_id"`
			Repository     string `json:"repository"`
			TargetFile     string `json:"target_file"`
			DefectCategory string `json:"defect_category"`
			Stratum        string `json:"stratum"`
			ExpectedAction string `json:"expected_action"`
			Description    string `json:"description"`
			IsDemoTarget   bool   `json:"is_demo_target"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &targetsResp); err != nil {
		t.Fatalf("Failed to parse targets: %v", err)
	}

	var scannedTargets []string
	var demoTargets []string
	for _, trg := range targetsResp.Targets {
		if trg.Stratum == "SCANNED_REPOSITORY" {
			scannedTargets = append(scannedTargets, trg.TargetID)
		} else {
			demoTargets = append(demoTargets, trg.TargetID)
		}
	}

	t.Logf("Discovered %d scanned targets and %d demo targets", len(scannedTargets), len(demoTargets))
	if len(scannedTargets) == 0 {
		t.Fatalf("Expected scanned repository targets for CV-Code-Vitals/Thermodynamic-Evolution-UCS503, found 0")
	}
	if len(demoTargets) == 0 {
		t.Fatalf("Expected demo targets to be preserved, found 0")
	}

	// Step 3: Test a repairable candidate target (e.g. SQL Injection parameterization)
	repairableTargetID := "SCAN-SEMANTIC_SQL_INJECTION-2"
	anBody, _ := json.Marshal(map[string]string{"target_id": repairableTargetID})
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(anBody))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/analyze failed for %s: %s", repairableTargetID, w3.Body.String())
	}
	var repAnalysis struct {
		AnalysisID           string  `json:"analysis_id"`
		AutomaticEligibility bool    `json:"automatic_eligibility"`
		UncertaintyScore     float64 `json:"uncertainty_score"`
		RequiresHumanReview  bool    `json:"requires_human_review"`
	}
	_ = json.Unmarshal(w3.Body.Bytes(), &repAnalysis)
	if !repAnalysis.AutomaticEligibility {
		t.Errorf("Expected automatic eligibility for %s", repairableTargetID)
	}
	if !repAnalysis.RequiresHumanReview {
		t.Errorf("Safety gate requires human review for all targets")
	}

	// Step 4: Propose repair for the repairable candidate
	prBody, _ := json.Marshal(map[string]string{"analysis_id": repAnalysis.AnalysisID})
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/api/remediation/propose", bytes.NewReader(prBody))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/propose failed: %s", w4.Body.String())
	}
	var repProposal struct {
		ProposalID  string `json:"proposal_id"`
		Status      string `json:"status"`
		UnifiedDiff string `json:"unified_diff"`
	}
	_ = json.Unmarshal(w4.Body.Bytes(), &repProposal)
	if repProposal.Status != "PROPOSED" {
		t.Fatalf("Expected PROPOSED status, got %s", repProposal.Status)
	}
	t.Logf("Successfully synthesized patch diff for %s (length %d chars)", repairableTargetID, len(repProposal.UnifiedDiff))

	// Step 5: Approve repair with cryptographic HMAC token
	apBody, _ := json.Marshal(map[string]string{
		"proposal_id": repProposal.ProposalID,
		"principal":   "security-officer@code-vitals.internal",
		"role":        "Autonomous Systems Compliance Lead",
	})
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodPost, "/api/remediation/approve", bytes.NewReader(apBody))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)

	if w5.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/approve failed: %s", w5.Body.String())
	}
	var approvalResp struct {
		TokenID string `json:"token_id"`
		Valid   bool   `json:"valid"`
	}
	_ = json.Unmarshal(w5.Body.Bytes(), &approvalResp)
	if !approvalResp.Valid || approvalResp.TokenID == "" {
		t.Fatalf("Approval failed: %+v", approvalResp)
	}

	// Step 6: Execute repair in sandbox and verify oracles + immutability
	exBody, _ := json.Marshal(map[string]string{"token_id": approvalResp.TokenID})
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodPost, "/api/remediation/execute", bytes.NewReader(exBody))
	req6.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w6, req6)

	if w6.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/execute failed: %s", w6.Body.String())
	}
	var execResp struct {
		ExecutionID           string `json:"execution_id"`
		AppliedInSandbox      bool   `json:"applied_in_sandbox"`
		OriginalRepoImmutable bool   `json:"original_repo_immutable"`
		TokenConsumed         bool   `json:"token_consumed"`
		VerificationVerdict   string `json:"verification_verdict"`
	}
	_ = json.Unmarshal(w6.Body.Bytes(), &execResp)
	if !execResp.AppliedInSandbox || !execResp.OriginalRepoImmutable || !execResp.TokenConsumed {
		t.Fatalf("Execution safety invariant violated: %+v", execResp)
	}
	t.Logf("Execution completed: Verdict=%s, SandboxApplied=%v, SourceImmutable=%v",
		execResp.VerificationVerdict, execResp.AppliedInSandbox, execResp.OriginalRepoImmutable)

	// Step 7: Test an unsupported / abstention candidate (e.g. Command Injection)
	abstainTargetID := "SCAN-SEMANTIC_COMMAND_INJECTION-41"
	absAnBody, _ := json.Marshal(map[string]string{"target_id": abstainTargetID})
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(absAnBody))
	req7.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w7, req7)

	if w7.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/analyze failed for %s: %s", abstainTargetID, w7.Body.String())
	}
	var absAnalysis struct {
		AnalysisID           string  `json:"analysis_id"`
		AutomaticEligibility bool    `json:"automatic_eligibility"`
		UncertaintyScore     float64 `json:"uncertainty_score"`
		EligibilityRationale string  `json:"eligibility_rationale"`
	}
	_ = json.Unmarshal(w7.Body.Bytes(), &absAnalysis)
	if absAnalysis.AutomaticEligibility {
		t.Errorf("Command injection must not be automatically eligible")
	}
	if absAnalysis.UncertaintyScore < 0.90 {
		t.Errorf("Expected high uncertainty for command injection, got %.2f", absAnalysis.UncertaintyScore)
	}

	absPrBody, _ := json.Marshal(map[string]string{"analysis_id": absAnalysis.AnalysisID})
	w8 := httptest.NewRecorder()
	req8, _ := http.NewRequest(http.MethodPost, "/api/remediation/propose", bytes.NewReader(absPrBody))
	req8.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w8, req8)

	if w8.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/propose failed for %s: %s", abstainTargetID, w8.Body.String())
	}
	var absProposal struct {
		Status        string `json:"status"`
		AbstainReason string `json:"abstain_reason"`
	}
	_ = json.Unmarshal(w8.Body.Bytes(), &absProposal)
	if absProposal.Status != "ABSTAINED" {
		t.Errorf("Expected ABSTAINED status for %s, got %s", abstainTargetID, absProposal.Status)
	}
	t.Logf("Abstention verified for %s: %s", abstainTargetID, absProposal.AbstainReason)

	// Step 8: Fetch audit report
	w9 := httptest.NewRecorder()
	req9, _ := http.NewRequest(http.MethodGet, "/api/remediation/audit", nil)
	r.ServeHTTP(w9, req9)
	if w9.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/audit failed: %s", w9.Body.String())
	}
	var auditResp struct {
		TotalSessionsRecorded int  `json:"total_sessions_recorded"`
		SourceImmutability100 bool `json:"source_immutability_100_percent"`
	}
	_ = json.Unmarshal(w9.Body.Bytes(), &auditResp)
	if !auditResp.SourceImmutability100 {
		t.Errorf("Audit report must verify 100%% source immutability")
	}
	t.Logf("Audit report verified: TotalSessions=%d, 100%% Immutability=%v",
		auditResp.TotalSessionsRecorded, auditResp.SourceImmutability100)
}
