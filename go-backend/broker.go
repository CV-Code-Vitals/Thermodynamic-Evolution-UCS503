package main

import (
	"context"
	"crypto/tls"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

const minSandboxServiceTokenBytes = 32

// runSandboxBroker starts the dedicated, hardened sandbox execution daemon.
// This daemon is the ONLY process on the host with Docker access.
// It exposes a strictly authenticated, narrow API with no arbitrary container execution capabilities.
func runSandboxBroker() {
	token := strings.TrimSpace(os.Getenv("SANDBOX_SERVICE_TOKEN"))
	if token == "" {
		log.Fatal("[FATAL] SANDBOX_SERVICE_TOKEN must be configured to run sandbox broker")
	}
	if len(token) < minSandboxServiceTokenBytes {
		log.Fatalf("[FATAL] SANDBOX_SERVICE_TOKEN must be at least %d bytes for security", minSandboxServiceTokenBytes)
	}

	port := envOr("PORT", "8082")
	certFile := strings.TrimSpace(os.Getenv("SANDBOX_TLS_CERT_FILE"))
	keyFile := strings.TrimSpace(os.Getenv("SANDBOX_TLS_KEY_FILE"))
	if certFile == "" || keyFile == "" {
		log.Fatal("[FATAL] SANDBOX_TLS_CERT_FILE and SANDBOX_TLS_KEY_FILE must be configured")
	}

	evaluator := agent.NewDockerEvaluator(agent.DockerEvaluatorConfig{Image: os.Getenv("SANDBOX_RUNNER_IMAGE")})
	workspaceRoot := envOr("SANDBOX_BROKER_WORKSPACE_DIR", filepath.Join(os.TempDir(), "ucs503-sandbox-workspaces"))
	workspaces, err := newBrokerWorkspaceManager(workspaceRoot, 30*time.Minute)
	if err != nil {
		log.Fatalf("[FATAL] initialize sandbox workspaces: %v", err)
	}
	limiter := NewEvolutionAdmissionControllerFromEnv()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	handler := newSandboxBrokerHandler(token, evaluator, workspaces, limiter)
	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	go workspaces.RunCleanup(cleanupCtx, time.Minute)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("[INFO] UCS503 Sandbox Broker listening on %s with TLS", server.Addr)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
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
	if math.IsNaN(limits.MaxCPUs) || math.IsInf(limits.MaxCPUs, 0) || limits.MaxCPUs <= 0 || limits.MaxCPUs > maxAllowedCPUs {
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
