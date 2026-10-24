package main

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

// runSandboxBroker starts the dedicated, hardened sandbox execution daemon.
// This daemon is the ONLY process on the host with Docker access.
// It exposes a strictly authenticated, narrow API with no arbitrary container execution capabilities.
func runSandboxBroker() {
	token := strings.TrimSpace(os.Getenv("SANDBOX_SERVICE_TOKEN"))
	if token == "" {
		log.Fatal("[FATAL] SANDBOX_SERVICE_TOKEN must be configured to run sandbox broker")
	}
	if len(token) < 16 {
		log.Fatal("[FATAL] SANDBOX_SERVICE_TOKEN must be at least 16 characters for security")
	}

	port := envOr("PORT", "8082")
	runnerImage := os.Getenv("SANDBOX_RUNNER_IMAGE")

	evaluator := agent.NewDockerEvaluator(agent.DockerEvaluatorConfig{
		Image: runnerImage,
	})

	limiter := NewEvolutionAdmissionControllerFromEnv()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Broker token authentication middleware (allows loopback probes or valid X-Sandbox-Token)
	authMiddleware := func(c *gin.Context) {
		provided := c.GetHeader("X-Sandbox-Token")
		if len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: invalid or missing X-Sandbox-Token header",
			})
			return
		}
		c.Next()
	}

	// Health check endpoint (allows token or internal loopback health probes)
	r.GET("/health", func(c *gin.Context) {
		clientIP := c.ClientIP()
		isLoopback := clientIP == "127.0.0.1" || clientIP == "::1"
		provided := c.GetHeader("X-Sandbox-Token")
		hasValidToken := len(provided) == len(token) && subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1

		if !isLoopback && !hasValidToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: invalid or missing X-Sandbox-Token header",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := evaluator.CheckAvailable(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"runner": runnerImage,
		})
	})

	// Execution endpoint (evaluates untrusted candidate inside hardened sandbox)
	r.POST("/v1/evaluate", authMiddleware, func(c *gin.Context) {
		var req agent.EvaluationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evaluation request body"})
			return
		}

		// Enforce server-side resource bounds (defense-in-depth against client override)
		req.Limits = clampLimits(req.Limits)

		// Enforce admission control
		release, err := limiter.Acquire(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Sandbox execution engine at maximum capacity",
				"code":  "server_busy",
			})
			return
		}
		defer release()

		res, _ := evaluator.EvaluateCandidate(c.Request.Context(), req)
		c.JSON(http.StatusOK, res)
	})

	addr := ":" + port
	log.Printf("[INFO] UCS503 Sandbox Broker listening on %s (isolated Docker execution daemon)", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[FATAL] Sandbox Broker failed: %v", err)
	}
}

// clampLimits enforces server-side ceilings on candidate execution limits.
func clampLimits(limits agent.ResourceLimits) agent.ResourceLimits {
	def := agent.DefaultResourceLimits()
	const (
		maxAllowedTimeout     = 120 * time.Second
		maxAllowedMemoryMB    = 2048
		maxAllowedCPUs        = 2.0
		maxAllowedPids        = 256
		maxAllowedOutputBytes = 256 * 1024
	)

	if limits.Timeout <= 0 || limits.Timeout > maxAllowedTimeout {
		limits.Timeout = def.Timeout
	}
	if limits.MaxMemoryMB <= 0 || limits.MaxMemoryMB > maxAllowedMemoryMB {
		limits.MaxMemoryMB = def.MaxMemoryMB
	}
	if limits.MaxCPUs <= 0 || limits.MaxCPUs > maxAllowedCPUs {
		limits.MaxCPUs = def.MaxCPUs
	}
	if limits.MaxPids <= 0 || limits.MaxPids > maxAllowedPids {
		limits.MaxPids = def.MaxPids
	}
	if limits.MaxOutputBytes <= 0 || limits.MaxOutputBytes > maxAllowedOutputBytes {
		limits.MaxOutputBytes = def.MaxOutputBytes
	}
	return limits
}
