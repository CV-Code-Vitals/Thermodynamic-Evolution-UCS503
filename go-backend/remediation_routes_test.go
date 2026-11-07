package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRemediationHttpWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() })

	// 1. GET /api/remediation/targets
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/remediation/targets", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/targets returned %d: %s", w1.Code, w1.Body.String())
	}
	var targetsResp struct {
		Targets []struct {
			TargetID string `json:"target_id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &targetsResp); err != nil || len(targetsResp.Targets) == 0 {
		t.Fatalf("Failed to parse targets: %v (len: %d)", err, len(targetsResp.Targets))
	}

	// 2. POST /api/remediation/analyze for NAT-03
	analyzeBody, _ := json.Marshal(map[string]string{"target_id": "NAT-03"})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(analyzeBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/analyze returned %d: %s", w2.Code, w2.Body.String())
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
		TokenID string `json:"token_id"`
		Valid   bool   `json:"valid"`
	}
	if err := json.Unmarshal(w4.Body.Bytes(), &apprResp); err != nil || !apprResp.Valid {
		t.Fatalf("Approval failed: %v", err)
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
	if err := json.Unmarshal(w5.Body.Bytes(), &execResp); err != nil || !execResp.AllOraclesPassed {
		t.Fatalf("Execution oracles failed: %v (verdict: %s)", err, execResp.VerificationVerdict)
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
	req7, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(absAnalyzeBody))
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
