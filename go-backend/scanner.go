package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScanRepoRequest struct {
	RepoURL string `json:"repo_url"`
}

// scanRepoHandler handles POST /scan and POST /api/scan for live repository analysis.
func scanRepoHandler(enginePath string, engineTimeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScanRepoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "invalid_json",
				Detail: "Expected JSON body with 'repo_url' field.",
			})
			return
		}

		repoURL := strings.TrimSpace(req.RepoURL)
		if repoURL == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "missing_repo_url",
				Detail: "Repository URL cannot be empty.",
			})
			return
		}

		// Security: Prevent argument injection & SSRF
		if strings.HasPrefix(repoURL, "-") || strings.ContainsAny(repoURL, " \t\r\n;|&`$") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "invalid_url",
				Detail: "The provided repository URL contains disallowed characters.",
			})
			return
		}

		parsed, err := url.Parse(repoURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "invalid_url_scheme",
				Detail: "Only valid http(s) repository URLs are allowed (e.g. https://github.com/org/repo).",
			})
			return
		}

		// Create temp directory for cloning
		uploadsDir := "./uploads"
		_ = os.MkdirAll(uploadsDir, 0o750)
		targetDir := filepath.Join(uploadsDir, fmt.Sprintf("clone_%d_%s", time.Now().UnixMilli(), uuid.New().String()[:8]))

		// Deferred cleanup of the cloned repository
		defer func() {
			go func() {
				time.Sleep(1 * time.Second)
				_ = os.RemoveAll(targetDir)
				log.Printf("[INFO] Cleaned up cloned repo at %s", targetDir)
			}()
		}()

		// Clone with shallow depth to save bandwidth and time
		cloneCtx, cloneCancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
		defer cloneCancel()

		log.Printf("[SCAN] Cloning %s into %s ...", repoURL, targetDir)
		cloneCmd := exec.CommandContext(cloneCtx, "git", "clone", "--depth", "1", repoURL, targetDir)
		out, err := cloneCmd.CombinedOutput()
		if err != nil {
			log.Printf("[ERROR] Git clone failed: %v, output: %s", err, string(out))
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:  "clone_failed",
				Detail: fmt.Sprintf("Failed to clone repository: %v", err),
			})
			return
		}

		// Run engine against the cloned codebase
		reportPath := filepath.Join(targetDir, "thermodynamic_report.json")
		engineCtx, engineCancel := context.WithTimeout(c.Request.Context(), engineTimeout)
		defer engineCancel()

		log.Printf("[SCAN] Running engine on %s ...", targetDir)
		if _, err := runEngine(engineCtx, enginePath, targetDir, reportPath); err != nil {
			log.Printf("[ERROR] Scan engine failed: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "engine_failed",
				Detail: fmt.Sprintf("Thermodynamic engine execution failed: %v", err),
			})
			return
		}

		// Parse the output report
		report, err := readReport(reportPath)
		if err != nil {
			log.Printf("[ERROR] Scan readReport failed: %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "report_parse_failed",
				Detail: fmt.Sprintf("Could not parse engine report: %v", err),
			})
			return
		}

		log.Printf("[SCAN] Completed for %s: %d files, %d hotspots, entropy %.2f",
			repoURL, report.FilesAnalyzed, report.TotalHotspots, report.GlobalEntropy)

		c.JSON(http.StatusOK, report)
	}
}
