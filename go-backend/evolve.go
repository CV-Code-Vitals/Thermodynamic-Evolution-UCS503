package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

type EvolveRequest struct {
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

func evolveHandler(workspaces *workspaceStore) gin.HandlerFunc {
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
		var req EvolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if req.WorkspaceID != "" {
			path, releaseWorkspace, err := workspaces.acquire(req.WorkspaceID)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			defer releaseWorkspace()
			target, err := readWorkspaceTarget(path, req.FilePath, req.FunctionName, req.StartByte)
			if err != nil || req.SourceHash == "" || target.SourceHash != req.SourceHash {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Target changed or is invalid; scan the repository again."})
				return
			}
			req.RepoPath, req.FilePath, err = workspaceModule(path, target.FilePath)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
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

		ctx, cancel := context.WithTimeout(c.Request.Context(), 9*time.Minute)
		defer cancel()
		ch := make(chan agent.StepTelemetry, 16)
		type completion struct {
			result *agent.AnnealingResult
			err    error
		}
		done := make(chan completion, 1)
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()
		annealer := &agent.Annealer{
			Mutator:   agent.NewAdaptiveMutator(llmAgent, 0),
			Evaluator: evaluator, EnergyEvaluator: energyEvaluator,
		}

		go func() {
			defer close(ch)
			result, err := annealer.Evolve(ctx, req.RepoPath, req.FilePath, req.FunctionName, req.StartByte, req.EndByte, req.BaseEnergy, req.Config, ch)
			done <- completion{result, err}
		}()

		for {
			select {
			case <-heartbeat.C:
				fmt.Fprint(c.Writer, ": keepalive\n\n")
				c.Writer.Flush()
			case <-ctx.Done():
				// Wait for cleanup before releasing the workspace for another run.
				<-done
				return
			case telemetry, ok := <-ch:
				if !ok {
					finished := <-done
					if finished.err != nil {
						data, _ := json.Marshal(gin.H{"error": finished.err.Error()})
						fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", data)
					} else {
						if req.WorkspaceID != "" {
							command := exec.CommandContext(ctx, "git", "diff", "--no-ext-diff", "--no-textconv", "--binary", "--", req.FilePath)
							command.Dir = req.RepoPath
							patch, patchErr := command.Output()
							if patchErr != nil {
								fmt.Fprint(c.Writer, "event: error\ndata: {\"error\":\"Could not export the optimized patch. Scan again before retrying.\"}\n\n")
								c.Writer.Flush()
								return
							}
							finished.result.Patch = string(patch)
						}
						data, _ := json.Marshal(finished.result)
						fmt.Fprintf(c.Writer, "event: complete\ndata: %s\n\n", data)
					}
					c.Writer.Flush()
					return
				}
				data, err := json.Marshal(telemetry)
				if err != nil {
					log.Printf("Failed to marshal telemetry: %v", err)
					continue
				}
				fmt.Fprintf(c.Writer, "event: step\ndata: %s\n\n", string(data))
				c.Writer.Flush()
			}
		}
	}
}
