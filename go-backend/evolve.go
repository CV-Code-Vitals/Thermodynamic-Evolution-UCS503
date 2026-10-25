package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

const (
	maxEvolutionDuration       = 14 * time.Minute
	evolutionHeartbeatInterval = 30 * time.Second
)

type evolutionOutcome struct {
	result *agent.AnnealingResult
	err    error
}

type EvolveRequest struct {
	SnapshotID   string                `json:"snapshot_id"`
	FilePath     string                `json:"file_path"`
	FunctionName string                `json:"function_name"`
	StartByte    int                   `json:"start_byte"`
	EndByte      int                   `json:"end_byte"`
	BaseEnergy   float64               `json:"base_energy"`
	Config       agent.AnnealingConfig `json:"config"`
}

func evolveHandler(snapshotStores ...SnapshotWorkspaceProvider) gin.HandlerFunc {
	var snapshots SnapshotWorkspaceProvider
	if len(snapshotStores) > 0 {
		snapshots = snapshotStores[0]
	}
	var llmAgent *agent.MutationAgent
	if os.Getenv("MUTATION_API_KEY") != "" {
		llmAgent = &agent.MutationAgent{
			Client:  &http.Client{Timeout: 30 * time.Second},
			APIBase: os.Getenv("MUTATION_API_BASE"),
			APIKey:  os.Getenv("MUTATION_API_KEY"),
			Model:   os.Getenv("MUTATION_MODEL"),
		}
		if llmAgent.APIBase == "" {
			llmAgent.APIBase = "https://api.openai.com"
		}
	}

	mutator := agent.NewAdaptiveMutator(llmAgent, 0)
	evaluator := agent.NewEvaluator(agent.DockerEvaluatorConfig{
		Image: os.Getenv("SANDBOX_RUNNER_IMAGE"),
	})
	energyEvaluator := &agent.LazyRustEngineEvaluator{
		CustomPath: os.Getenv("ENGINE_PATH"),
	}
	annealer := &agent.Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: energyEvaluator,
	}
	limiter := NewEvolutionAdmissionControllerFromEnv()

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		var req EvolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		req.SnapshotID = strings.TrimSpace(req.SnapshotID)
		if !snapshotIDPattern.MatchString(req.SnapshotID) || strings.TrimSpace(req.FilePath) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "snapshot_id and file_path are required"})
			return
		}
		if snapshots == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "source snapshot storage is not configured"})
			return
		}

		cleanInput := filepath.ToSlash(req.FilePath)
		cleanRel := filepath.Clean(filepath.FromSlash(req.FilePath))
		isWinDrive := len(req.FilePath) >= 2 && req.FilePath[1] == ':' &&
			((req.FilePath[0] >= 'a' && req.FilePath[0] <= 'z') || (req.FilePath[0] >= 'A' && req.FilePath[0] <= 'Z'))
		if filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanInput, "/") || strings.HasPrefix(cleanInput, "\\") ||
			strings.HasPrefix(cleanRel, "..") || filepath.VolumeName(cleanRel) != "" || isWinDrive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path traversal in file_path is forbidden"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), maxEvolutionDuration)
		defer cancel()
		release, err := limiter.Acquire(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, ErrLimiterCancelled) {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled while waiting for evolution capacity"})
				return
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Evolution engine at maximum capacity. Please retry shortly.",
				"code":  "server_busy",
			})
			return
		}
		defer release()

		workspace, cleanup, err := snapshots.ExtractSnapshot(ctx, req.SnapshotID)
		if err != nil {
			log.Printf("evolution source snapshot extraction failed (snapshot_id=%s)", req.SnapshotID)
			c.JSON(http.StatusBadGateway, gin.H{"error": "The staged source snapshot is unavailable"})
			return
		}
		defer cleanup()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		telemetry := make(chan agent.StepTelemetry, 16)
		outcomes := make(chan evolutionOutcome, 1)
		go func() {
			defer close(telemetry)
			result, evolveErr := annealer.Evolve(
				ctx,
				workspace,
				req.FilePath,
				req.FunctionName,
				req.StartByte,
				req.EndByte,
				req.BaseEnergy,
				req.Config,
				telemetry,
			)
			outcomes <- evolutionOutcome{result: result, err: evolveErr}
		}()

		streamEvolutionEvents(ctx, c.Writer, telemetry, outcomes, evolutionHeartbeatInterval)
	}
}

func streamEvolutionEvents(
	ctx context.Context,
	writer http.ResponseWriter,
	telemetry <-chan agent.StepTelemetry,
	outcomes <-chan evolutionOutcome,
	heartbeatInterval time.Duration,
) {
	if heartbeatInterval <= 0 {
		heartbeatInterval = evolutionHeartbeatInterval
	}
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	flush := func() {
		if flusher, ok := writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	writeEvent := func(name string, value any) bool {
		data, err := json.Marshal(value)
		if err != nil {
			log.Printf("failed to encode Evolution %s event", name)
			return false
		}
		if _, err := fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", name, data); err != nil {
			return false
		}
		flush()
		return true
	}

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				writeEvent("error", map[string]string{"error": "Evolution exceeded its runtime limit"})
			}
			return
		case event, ok := <-telemetry:
			if !ok {
				telemetry = nil
				continue
			}
			if !writeEvent("step", event) {
				return
			}
		case outcome := <-outcomes:
			if outcome.err != nil {
				log.Printf("Evolution worker failed")
				writeEvent("error", map[string]string{"error": "Evolution failed"})
				return
			}
			writeEvent("complete", outcome.result)
			return
		case <-ticker.C:
			if _, err := fmt.Fprint(writer, ": heartbeat\n\n"); err != nil {
				return
			}
			flush()
		}
	}
}
