package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
)

func resolveWorkspaceRoot() string {
	return execution.ResolveWorkspaceRoot()
}

func setupRemediationRoutes(r *gin.Engine, tokenMiddleware gin.HandlerFunc, workspaces ...*workspaceStore) {
	wsRoot := execution.ResolveWorkspaceRoot()
	tmpDir, _ := execution.EnsureTmpDir(wsRoot)
	ledgerPath := filepath.Join(tmpDir, "consumed_approvals_mvp.json")
	mvpService, err := execution.NewMVPService(wsRoot, ledgerPath)
	if err != nil {
		// Log error but do not crash server
		return
	}

	if len(workspaces) > 0 && workspaces[0] != nil {
		mvpService.SetWorkspaceResolver(workspaces[0].get)
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
		targets := mvpService.ListTargets(workspaceID)
		c.JSON(http.StatusOK, gin.H{"targets": targets})
	}

	analyzeHandler := func(c *gin.Context) {
		var req struct {
			TargetID string `json:"target_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.TargetID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_id is required"})
			return
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
			Principal  string `json:"principal"`
			Role       string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.ProposalID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "proposal_id is required"})
			return
		}
		approval, err := mvpService.ApproveRepair(req.ProposalID, req.Principal, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusOK, gin.H{"success": ok})
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

	// Register with and without /api prefix
	r.GET("/api/remediation/targets", targetsHandler)
	r.GET("/remediation/targets", targetsHandler)
	r.POST("/api/remediation/targets", targetsHandler)
	r.POST("/remediation/targets", targetsHandler)

	r.POST("/api/remediation/analyze", analyzeHandler)
	r.POST("/remediation/analyze", analyzeHandler)

	r.POST("/api/remediation/propose", proposeHandler)
	r.POST("/remediation/propose", proposeHandler)

	r.POST("/api/remediation/approve", approveHandler)
	r.POST("/remediation/approve", approveHandler)

	r.POST("/api/remediation/execute", executeHandler)
	r.POST("/remediation/execute", executeHandler)

	r.POST("/api/remediation/rollback", rollbackHandler)
	r.POST("/remediation/rollback", rollbackHandler)

	r.GET("/api/remediation/audit", auditHandler)
	r.GET("/remediation/audit", auditHandler)

	r.POST("/api/remediation/demo", demoHandler)
	r.POST("/remediation/demo", demoHandler)
}
