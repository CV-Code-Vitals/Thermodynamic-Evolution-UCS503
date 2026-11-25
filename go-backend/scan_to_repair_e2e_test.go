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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

func TestEndToEndScanToRepairWorkflow_LiveRepo(t *testing.T) {
	if testing.Short() || os.Getenv("RUN_LIVE_REPO_TESTS") != "1" {
		t.Skip("set RUN_LIVE_REPO_TESTS=1 to scan the live GitHub repository")
	}

	gin.SetMode(gin.TestMode)
	enginePath := resolveEnginePath()
	engineTimeout := 120 * time.Second
	workspaces := newWorkspaceStore()

	r := gin.New()
	scanH := scanRepoHandler(enginePath, engineTimeout, workspaces)
	r.POST("/api/scan", scanH)
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, workspaces, execution.NewUnavailableContainerRunner())

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
	if len(demoTargets) != 0 {
		t.Fatalf("Live discovery contaminated with demo targets")
	}

	// Step 3: Test a repairable candidate target (e.g. SQL Injection parameterization)
	repairableTargetID := ""
	for _, target := range targetsResp.Targets {
		if target.DefectCategory == "SQL_INJECTION" && target.ExpectedAction == "REPAIR" {
			repairableTargetID = target.TargetID
			break
		}
	}
	if repairableTargetID == "" {
		t.Skip("real scan completed; no eligible SQL candidate; repair lifecycle NOT VERIFIED")
	}
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

	if w5.Code != http.StatusUnprocessableEntity || !strings.Contains(w5.Body.String(), "REPRODUCER_NOT_AVAILABLE") {
		t.Fatalf("unplanned live approval must reject: %s", w5.Body.String())
	}
}

func TestEndToEndScanToRepairWorkflow_OfflineFixture(t *testing.T) {
	gin.SetMode(gin.TestMode)
	workspaces := newWorkspaceStore()
	wsID, wsDir, err := workspaces.create()
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}
	defer os.RemoveAll(wsDir)
	workspaces.finishScan(wsID, true)

	// Create synthetic offline repository structure
	if err := os.MkdirAll(filepath.Join(wsDir, "db"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(wsDir, "ops"), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	_ = os.WriteFile(filepath.Join(wsDir, "go.mod"), []byte("module testfixture\n\ngo 1.21\n"), 0644)

	// 1. Repairable SQL injection target file
	sqlSrc := `package db

import (
	"context"
	"database/sql"
	"fmt"
)

func QueryUserByRole(ctx context.Context, db *sql.DB, role string) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT id, name FROM users WHERE role = '%s'", role)
	return db.QueryContext(ctx, query)
}
`
	if err := os.WriteFile(filepath.Join(wsDir, "db", "query.go"), []byte(sqlSrc), 0644); err != nil {
		t.Fatalf("Write query.go failed: %v", err)
	}

	// 2. Command injection target file
	cmdSrc := `package ops

import (
	"os/exec"
)

func RunScript(userInput string) error {
	cmd := exec.Command("sh", "-c", "echo " + userInput)
	return cmd.Run()
}
`
	if err := os.WriteFile(filepath.Join(wsDir, "ops", "runner.go"), []byte(cmdSrc), 0644); err != nil {
		t.Fatalf("Write runner.go failed: %v", err)
	}

	// 3. Thermodynamic report with complexity hotspot
	thermoReport := `{
  "files_analyzed": 2,
  "total_hotspots": 1,
  "file_reports": [
    {
      "file_path": "db/query.go",
      "language": "Go",
      "hotspots": [
        {
          "function_name": "QueryUserByRole",
          "byte_range": [0, 200],
          "cyclomatic_complexity": 18,
          "entropy_score": 82.5,
          "cognitive_depth": 4
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(wsDir, "thermodynamic_report.json"), []byte(thermoReport), 0644); err != nil {
		t.Fatalf("Write thermodynamic_report.json failed: %v", err)
	}

	r := gin.New()
	setupRemediationRoutes(r, func(c *gin.Context) { c.Next() }, workspaces, execution.NewUnavailableContainerRunner())

	// Fetch targets
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/remediation/targets?workspace_id="+wsID, nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("GET /api/remediation/targets returned %d: %s", w1.Code, w1.Body.String())
	}

	var targetsResp struct {
		Targets []struct {
			TargetID       string `json:"target_id"`
			DefectCategory string `json:"defect_category"`
			ExpectedAction string `json:"expected_action"`
			SnapshotID     string `json:"snapshot_id"`
			SourceFileHash string `json:"source_file_hash"`
		} `json:"targets"`
	}
	_ = json.Unmarshal(w1.Body.Bytes(), &targetsResp)

	var repairableTarget string
	for _, trg := range targetsResp.Targets {
		if trg.DefectCategory == "SQL_INJECTION" {
			repairableTarget = trg.TargetID
		}
	}

	if repairableTarget == "" {
		t.Fatalf("Expected SQL_INJECTION target in offline fixture")
	}

	// Analyze repairable target
	anBody, _ := json.Marshal(map[string]string{"target_id": repairableTarget})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/remediation/analyze", bytes.NewReader(anBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST /api/remediation/analyze returned %d: %s", w2.Code, w2.Body.String())
	}

	var anResp struct {
		AnalysisID           string `json:"analysis_id"`
		AutomaticEligibility bool   `json:"automatic_eligibility"`
		SnapshotID           string `json:"snapshot_id"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &anResp)
	if !anResp.AutomaticEligibility {
		t.Errorf("Expected automatic eligibility for SQL injection")
	}

	// Propose repair
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
	_ = json.Unmarshal(w3.Body.Bytes(), &prResp)
	if prResp.Status != "PROPOSED" {
		t.Fatalf("Expected PROPOSED status, got %s", prResp.Status)
	}

	// Approve repair
	apBody, _ := json.Marshal(map[string]string{
		"proposal_id": prResp.ProposalID,
		"principal":   "ci-security-gate@code-vitals.internal",
		"role":        "Autonomous Quality Engineer",
	})
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodPost, "/api/remediation/approve", bytes.NewReader(apBody))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnprocessableEntity || !strings.Contains(w4.Body.String(), "REPRODUCER_NOT_AVAILABLE") {
		t.Fatalf("unplanned fixture approval must reject: %s", w4.Body.String())
	}
}

func TestUntrustedExecution_CredentialStripping(t *testing.T) {
	// Set artificial host credentials
	t.Setenv("AWS_SECRET_ACCESS_KEY", "AKIAFAKEACCESSKEY123456")
	t.Setenv("GITHUB_TOKEN", "ghp_FakeGitHubSecretToken987654321")
	t.Setenv("DATABASE_URL", "postgres://admin:secretpass@db.internal:5432/prod")
	t.Setenv("PRIVATE_SIGNING_KEY", "MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC")

	cleanEnv := execution.GetSanitizedExecutionEnv()

	for _, entry := range cleanEnv {
		upper := strings.ToUpper(entry)
		if strings.Contains(upper, "AWS_SECRET_ACCESS_KEY") {
			t.Errorf("Security violation: AWS_SECRET_ACCESS_KEY leaked to sanitized environment: %s", entry)
		}
		if strings.Contains(upper, "GITHUB_TOKEN") {
			t.Errorf("Security violation: GITHUB_TOKEN leaked to sanitized environment: %s", entry)
		}
		if strings.Contains(upper, "DATABASE_URL") {
			t.Errorf("Security violation: DATABASE_URL leaked to sanitized environment: %s", entry)
		}
		if strings.Contains(upper, "PRIVATE_SIGNING_KEY") {
			t.Errorf("Security violation: PRIVATE_SIGNING_KEY leaked to sanitized environment: %s", entry)
		}
	}

	// Verify CGO_ENABLED=0 is explicitly enforced
	foundCgo := false
	for _, entry := range cleanEnv {
		if entry == "CGO_ENABLED=0" {
			foundCgo = true
			break
		}
	}
	if !foundCgo {
		t.Errorf("Security violation: CGO_ENABLED=0 not enforced in sanitized environment")
	}
}

func TestUntrustedExecution_ArbitraryCommandBlocked(t *testing.T) {
	suite := execution.NewIndependentOracleSuite(10)
	tmpDir := t.TempDir()

	maliciousCommands := []string{
		"sh -c 'echo pwned'",
		"bash -c 'whoami'",
		"powershell.exe -Command 'dir'",
		"cmd.exe /c 'dir'",
		"python -c 'import os; print(os.environ)'",
		"curl -s https://evil.example.com",
		"go test -exec 'rm -rf /'",
	}

	for _, cmdStr := range maliciousCommands {
		task := &planning.RemediationTask{
			TaskID:             "TASK-SECURITY-TEST",
			ValidationCommands: []string{cmdStr},
		}
		res := suite.VerifyAll(tmpDir, task, &execution.CandidatePatch{TargetFile: "main.go"}, "", "")
		regRes, exists := res.OracleResults[execution.OracleRegressionTest]
		if !exists || regRes.Passed {
			t.Errorf("Security violation: malicious command %q was not blocked: %+v", cmdStr, regRes)
		} else if !strings.Contains(regRes.FailureReason, "Security rejection") {
			t.Errorf("Expected Security rejection for %q, got: %s", cmdStr, regRes.FailureReason)
		}
	}
}

func TestSnapshotAuthorization_StaleWorkspaceRejected(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "stale_ws")
	_ = os.MkdirAll(wsDir, 0755)

	targetFile := "helper.go"
	initialContent := `package main

import (
	"context"
	"database/sql"
	"fmt"
)

func Query(ctx context.Context, db *sql.DB, val string) (*sql.Rows, error) {
	q := fmt.Sprintf("SELECT * FROM items WHERE k = '%s'", val)
	return db.QueryContext(ctx, q)
}
`
	fileAbs := filepath.Join(wsDir, targetFile)
	_ = os.WriteFile(fileAbs, []byte(initialContent), 0644)

	svc, err := execution.NewMVPService(tmpDir, filepath.Join(tmpDir, "ledger.json"))
	if err != nil {
		t.Fatalf("NewMVPService failed: %v", err)
	}

	wsID := "ws-snapshot-stale"
	svc.SetWorkspaceResolver(func(id string) (string, error) {
		if id == wsID {
			return wsDir, nil
		}
		return "", os.ErrNotExist
	})

	targets, err := svc.DiscoverWorkspaceTargets(wsID)
	if err != nil || len(targets) == 0 {
		t.Fatalf("DiscoverWorkspaceTargets failed: %v (targets: %d)", err, len(targets))
	}

	targetID := targets[0].TargetID
	analysis, err := svc.AnalyzeTarget(targetID)
	if err != nil {
		t.Fatalf("AnalyzeTarget failed: %v", err)
	}

	prop, err := svc.ProposeRepair(analysis.AnalysisID)
	if err != nil {
		t.Fatalf("ProposeRepair failed: %v", err)
	}

	appr, err := svc.ApproveRepair(prop.ProposalID, "auditor@code-vitals.internal", "Security Auditor")
	if err != nil {
		t.Fatalf("ApproveRepair failed: %v", err)
	}

	// Tamper with the original file on disk to simulate a stale workspace / race condition
	_ = os.WriteFile(fileAbs, []byte(initialContent+"\n// Stale mutation on disk\n"), 0644)

	// Attempt execution: must fail-closed with STALE_WORKSPACE
	_, execErr := svc.ExecuteRepair(appr.TokenID)
	if execErr == nil {
		t.Fatalf("Expected STALE_WORKSPACE error, but ExecuteRepair succeeded")
	}
	if !strings.Contains(execErr.Error(), "STALE_WORKSPACE") {
		t.Errorf("Expected STALE_WORKSPACE error message, got: %v", execErr)
	}
}
