package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
)

func resolveWorkspaceRoot() string {
	return execution.ResolveWorkspaceRoot()
}

func setupRemediationRoutes(r *gin.Engine, tokenMiddleware gin.HandlerFunc, workspaces *workspaceStore, runners ...execution.ContainerRunner) {
	wsRoot := execution.ResolveWorkspaceRoot()
	tmpDir, _ := execution.EnsureTmpDir(wsRoot)
	ledgerPath := filepath.Join(tmpDir, "consumed_approvals_mvp.json")
	mvpService, err := execution.NewMVPService(wsRoot, ledgerPath, runners...)
	if err != nil {
		log.Printf("remediation routes unavailable: %v", err)
		return
	}

	if config := os.Getenv("HEALPROOF_TRUSTED_PLANS"); config != "" {
		data, readErr := os.ReadFile(config)
		var plans []execution.VerificationPlan
		if readErr != nil {
			log.Printf("trusted plan configuration unavailable: %v", readErr)
			return
		}
		if err := json.Unmarshal(data, &plans); err != nil {
			log.Printf("trusted plan configuration invalid: %v", err)
			return
		}
		if err := mvpService.ConfigureTrustedVerificationPlans(plans); err != nil {
			log.Printf("trusted plan configuration rejected: %v", err)
			return
		}
	}
	setupRemediationRoutesWithService(r, tokenMiddleware, workspaces, mvpService)
}
func setupRemediationRoutesWithService(r *gin.Engine, tokenMiddleware gin.HandlerFunc, workspaces *workspaceStore, mvpService *execution.MVPService) {

	if workspaces != nil {
		mvpService.SetWorkspaceResolver(workspaces.get)
	}

	targetsHandler := func(c *gin.Context) {
		workspaceID := c.Query("workspace_id")
		if workspaceID == "" && c.Request.Method == http.MethodPost {
			var bodyReq struct {
				WorkspaceID string `json:"workspace_id"`
			}
			if err := c.ShouldBindJSON(&bodyReq); err == nil && bodyReq.WorkspaceID != "" {
				workspaceID = bodyReq.WorkspaceID
			}
		}
		if workspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required for live targets; use /demo explicitly"})
			return
		}
		targets, err := mvpService.ListLiveTargets(workspaceID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		if targets == nil {
			targets = []execution.MVPTarget{}
		}
		c.JSON(http.StatusOK, gin.H{"source_kind": "SCANNED_REPOSITORY", "analysis_scope": "Go non-test source; semantic heuristics, not proof of defects", "targets": targets})
	}

	analyzeHandler := func(c *gin.Context) {
		var req struct {
			TargetID string `json:"target_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.TargetID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_id is required"})
			return
		}
		for _, fixture := range execution.DefaultPreconfiguredTargets() {
			if req.TargetID == fixture.TargetID && !strings.Contains(c.FullPath(), "/demo/") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "DEMO_FIXTURE requires explicit /demo mode"})
				return
			}
		}
		analysis, err := mvpService.AnalyzeTarget(req.TargetID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, analysis)
	}

	proposeHandler := func(c *gin.Context) {
		var req struct {
			AnalysisID string `json:"analysis_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.AnalysisID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "analysis_id is required"})
			return
		}
		proposal, err := mvpService.ProposeRepair(req.AnalysisID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, proposal)
	}

	approveHandler := func(c *gin.Context) {
		var req struct {
			ProposalID string `json:"proposal_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.ProposalID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "proposal_id is required"})
			return
		}
		approval, err := mvpService.ApproveVerifiedRepair(req.ProposalID, execution.MVPAnalysisOperator, "analysis-operator")
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "verification_verdict": execution.VerificationFailureVerdict(err)})
			return
		}
		c.JSON(http.StatusOK, approval)
	}

	executeHandler := func(c *gin.Context) {
		var req struct {
			TokenID string `json:"token_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.TokenID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token_id is required"})
			return
		}
		executionRes, err := mvpService.ExecuteRepair(req.TokenID)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "verification_verdict": executionRes.VerificationVerdict, "execution": executionRes})
			return
		}
		c.JSON(http.StatusOK, executionRes)
	}

	rollbackHandler := func(c *gin.Context) {
		var req struct {
			ExecutionID string `json:"execution_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.ExecutionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution_id is required"})
			return
		}
		ok, err := mvpService.RollbackSandbox(req.ExecutionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": ok, "operation": "DISCARD_SANDBOX", "revision_restored": false})
	}

	auditHandler := func(c *gin.Context) {
		report := mvpService.ExportAuditReport()
		c.JSON(http.StatusOK, report)
	}

	demoHandler := func(c *gin.Context) {
		result, err := mvpService.RunDeterministicDemo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}

	for _, prefix := range []string{"/api/remediation", "/remediation"} {
		routes := r.Group(prefix, tokenMiddleware, func(c *gin.Context) {
			c.Header("Cache-Control", "no-store")
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		})
		routes.GET("/demo/targets", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"source_kind": "DEMO_FIXTURE", "targets": execution.DefaultPreconfiguredTargets()})
		})
		routes.POST("/demo/analyze", analyzeHandler)
		routes.GET("/targets", targetsHandler)
		routes.POST("/targets", targetsHandler)
		routes.POST("/analyze", analyzeHandler)
		routes.POST("/propose", proposeHandler)
		routes.POST("/approve", approveHandler)
		routes.POST("/execute", executeHandler)
		routes.POST("/rollback", rollbackHandler)
		routes.GET("/audit", auditHandler)
		routes.POST("/demo", demoHandler)
	}
}
