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
	maxEvolutionDuration       = 9 * time.Minute
	evolutionHeartbeatInterval = 15 * time.Second
)

type evolutionOutcome struct {
	result *agent.AnnealingResult
	err    error
}

type EvolveRequest struct {
	SnapshotID   string                `json:"snapshot_id"`
	WorkspaceID  string                `json:"workspace_id"`
	RepoPath     string                `json:"repo_path"`
	FilePath     string                `json:"file_path"`
	FunctionName string                `json:"function_name"`
	StartByte    int                   `json:"start_byte"`
	EndByte      int                   `json:"end_byte"`
	BaseEnergy   float64               `json:"base_energy"`
	SourceHash   string                `json:"source_hash"`
	Config       agent.AnnealingConfig `json:"config"`
}

func evolveHandler(workspaces *workspaceStore, snapshots ...SnapshotWorkspaceProvider) gin.HandlerFunc {
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

	// Configure the hardened sandbox evaluator (supports direct Docker or authenticated remote broker)
	evaluator := agent.NewEvaluator(agent.DockerEvaluatorConfig{
		Image: os.Getenv("SANDBOX_RUNNER_IMAGE"),
	})

	// Configure the AST thermodynamic energy evaluator (resolves Rust engine path dynamically)
	energyEvaluator := &agent.LazyRustEngineEvaluator{
		CustomPath: os.Getenv("ENGINE_PATH"),
	}

	// Global concurrency admission controller for expensive evolution workflows
	limiter := NewEvolutionAdmissionControllerFromEnv()

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		var req EvolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		root := ""
		if req.SnapshotID != "" {
			if !snapshotIDPattern.MatchString(req.SnapshotID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "A valid source snapshot ID is required"})
				return
			}
			if len(snapshots) == 0 || snapshots[0] == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "source snapshot storage is not configured"})
				return
			}
			path, cleanup, err := snapshots[0].ExtractSnapshot(c.Request.Context(), req.SnapshotID)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "The staged source snapshot is unavailable"})
				return
			}
			defer cleanup()
			root = path
		} else if req.WorkspaceID != "" {
			path, releaseWorkspace, err := workspaces.acquire(req.WorkspaceID)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			defer releaseWorkspace()
			root = path
		}
		exportPath := req.FilePath
		if root != "" {
			target, err := readWorkspaceTarget(root, req.FilePath, req.FunctionName, req.StartByte)
			if err != nil || req.SourceHash == "" || target.SourceHash != req.SourceHash {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Target changed or is invalid; scan the repository again."})
				return
			}
			req.RepoPath, req.FilePath, err = workspaceModule(root, target.FilePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			exportPath = target.FilePath
			req.StartByte, req.EndByte = target.StartByte, target.EndByte
		}
		if req.Config.MaxSteps == 0 {
			req.Config.MaxSteps = 20
		}
		if req.Config.MaxSteps < 1 || req.Config.MaxSteps > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "max_steps must be between 1 and 100"})
			return
		}

		// Security pre-validation of target paths
		if strings.TrimSpace(req.RepoPath) == "" || strings.TrimSpace(req.FilePath) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo_path and file_path are required"})
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

		// Context-aware concurrency admission control
		release, err := limiter.Acquire(c.Request.Context())
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
		defer release() // Guarantees capacity is released on every exit path
		availabilityCtx, cancelAvailability := context.WithTimeout(c.Request.Context(), 2*time.Second)
		err = evaluator.CheckAvailable(availabilityCtx)
		cancelAvailability()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Sandbox unavailable. Start Docker or configure the sandbox broker before evolving."})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		fmt.Fprint(c.Writer, ": evolution started\n\n")
		c.Writer.Flush()

		ctx, cancel := context.WithTimeout(c.Request.Context(), maxEvolutionDuration)
		defer cancel()
		original, err := os.ReadFile(filepath.Join(req.RepoPath, req.FilePath))
		if err != nil {
			fmt.Fprint(c.Writer, "event: error\ndata: {\"error\":\"Could not read the scanned source\"}\n\n")
			c.Writer.Flush()
			return
		}
		telemetry := make(chan agent.StepTelemetry, 16)
		outcomes := make(chan evolutionOutcome, 1)
		workerDone := make(chan struct{})
		annealer := &agent.Annealer{
			Mutator:   agent.NewAdaptiveMutator(llmAgent, 0),
			Evaluator: evaluator, EnergyEvaluator: energyEvaluator,
		}
		go func() {
			defer close(workerDone)
			defer close(telemetry)
			result, err := annealer.Evolve(ctx, req.RepoPath, req.FilePath, req.FunctionName, req.StartByte, req.EndByte, req.BaseEnergy, req.Config, telemetry)
			if err == nil && root != "" {
				var updated []byte
				updated, err = os.ReadFile(filepath.Join(req.RepoPath, req.FilePath))
				if err == nil {
					result.Patch, err = exportEvolutionPatch(original, updated, exportPath)
				}
			}
			outcomes <- evolutionOutcome{result: result, err: err}
		}()
		streamEvolutionEvents(ctx, c.Writer, telemetry, outcomes, evolutionHeartbeatInterval)
		cancel()
		// Keep the source and sandbox lease until the worker finishes cleanup.
		<-workerDone
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
			for {
				select {
				case event, ok := <-telemetry:
					if !ok {
						telemetry = nil
						continue
					}
					if !writeEvent("step", event) {
						return
					}
				default:
					goto drained
				}
			}
		drained:
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
