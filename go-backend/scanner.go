package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ScanRepoRequest struct {
	SnapshotID string `json:"snapshot_id"`
}

type scanEngineRunner func(context.Context, string, string, string) (string, error)

func scanRepoHandler(
	enginePath string,
	engineTimeout time.Duration,
	snapshots SnapshotWorkspaceProvider,
	controller *PublicScanController,
) gin.HandlerFunc {
	return scanRepoHandlerWithRunner(enginePath, engineTimeout, snapshots, controller, runEngine)
}

func scanRepoHandlerWithRunner(
	enginePath string,
	engineTimeout time.Duration,
	snapshots SnapshotWorkspaceProvider,
	controller *PublicScanController,
	run scanEngineRunner,
) gin.HandlerFunc {
	if engineTimeout <= 0 {
		engineTimeout = 5 * time.Minute
	}
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8<<10)
		var req ScanRepoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "invalid_json",
				Detail: "Expected a bounded JSON body with a snapshot_id.",
			})
			return
		}

		snapshotID := strings.TrimSpace(req.SnapshotID)
		if !snapshotIDPattern.MatchString(snapshotID) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "invalid_snapshot_id",
				Detail: "A valid source snapshot ID is required.",
			})
			return
		}
		if snapshots == nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error:  "snapshot_store_unavailable",
				Detail: "Source snapshot storage is not configured.",
			})
			return
		}
		if controller == nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error:  "scan_admission_unavailable",
				Detail: "Public scan admission is not configured.",
			})
			return
		}

		clientIP := strings.TrimSpace(c.GetHeader("X-Trusted-Client-IP"))
		if net.ParseIP(clientIP) == nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "trusted_client_ip_required",
				Detail: "A trusted API Gateway client IP is required.",
			})
			return
		}
		release, err := controller.Acquire(c.Request.Context(), clientIP)
		if err != nil {
			switch {
			case errors.Is(err, ErrPublicScanRateLimited):
				c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "scan_rate_limited"})
			case errors.Is(err, ErrPublicScanBusy):
				c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "scan_capacity_reached"})
			case errors.Is(err, ErrLimiterCancelled), errors.Is(err, context.Canceled):
				c.JSON(http.StatusRequestTimeout, ErrorResponse{Error: "scan_cancelled"})
			case errors.Is(err, ErrInvalidTrustedClientIP):
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "trusted_client_ip_required"})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "scan_admission_failed"})
			}
			return
		}
		defer release()

		workspace, cleanup, err := snapshots.ExtractSnapshot(c.Request.Context(), snapshotID)
		if err != nil {
			log.Printf("source snapshot extraction failed (snapshot_id=%s): %v", snapshotID, err)
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:  "snapshot_unavailable",
				Detail: "The staged source snapshot could not be read or extracted.",
			})
			return
		}
		defer cleanup()

		engineCtx, cancel := context.WithTimeout(c.Request.Context(), engineTimeout)
		defer cancel()
		reportPath := filepath.Join(workspace, "thermodynamic_report.json")
		if _, err := run(engineCtx, enginePath, workspace, reportPath); err != nil {
			status := http.StatusInternalServerError
			code := "engine_failed"
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
				code = "engine_timeout"
			} else if errors.Is(err, context.Canceled) {
				status = http.StatusRequestTimeout
				code = "scan_cancelled"
			}
			log.Printf("scan engine failed (snapshot_id=%s, error=%v)", snapshotID, err)
			c.JSON(status, ErrorResponse{Error: code, Detail: "The analysis engine did not complete the staged-source scan."})
			return
		}

		report, err := readReport(reportPath)
		if err != nil {
			log.Printf("scan report could not be read (snapshot_id=%s): %v", snapshotID, err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "report_parse_failed",
				Detail: "The analysis engine returned an invalid report.",
			})
			return
		}

		log.Printf("scan completed (snapshot_id=%s, files=%d, hotspots=%d)", snapshotID, report.FilesAnalyzed, report.TotalHotspots)
		c.JSON(http.StatusOK, report)
	}
}
