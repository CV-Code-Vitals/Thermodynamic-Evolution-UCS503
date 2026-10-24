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

type EvolveRequest struct {
	RepoPath     string                `json:"repo_path"`
	FilePath     string                `json:"file_path"`
	FunctionName string                `json:"function_name"`
	StartByte    int                   `json:"start_byte"`
	EndByte      int                   `json:"end_byte"`
	BaseEnergy   float64               `json:"base_energy"`
	Config       agent.AnnealingConfig `json:"config"`
}

func evolveHandler() gin.HandlerFunc {
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

	// Configure the hardened sandbox evaluator (supports direct Docker or authenticated remote broker)
	evaluator := agent.NewEvaluator(agent.DockerEvaluatorConfig{
		Image: os.Getenv("SANDBOX_RUNNER_IMAGE"),
	})

	// Configure the AST thermodynamic energy evaluator (resolves Rust engine path dynamically)
	energyEvaluator := &agent.LazyRustEngineEvaluator{
		CustomPath: os.Getenv("ENGINE_PATH"),
	}

	annealer := &agent.Annealer{
		Mutator:         mutator,
		Evaluator:       evaluator,
		EnergyEvaluator: energyEvaluator,
	}

	// Global concurrency admission controller for expensive evolution workflows
	limiter := NewEvolutionAdmissionControllerFromEnv()

	return func(c *gin.Context) {
		var req EvolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
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

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		ctx := c.Request.Context()
		ch := make(chan agent.StepTelemetry, 16)

		go func() {
			defer close(ch)
			result, err := annealer.Evolve(ctx, req.RepoPath, req.FilePath, req.FunctionName, req.StartByte, req.EndByte, req.BaseEnergy, req.Config, ch)

			if err != nil {
				data, _ := json.Marshal(map[string]string{"error": err.Error()})
				fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", string(data))
				c.Writer.Flush()
				return
			}

			data, _ := json.Marshal(result)
			fmt.Fprintf(c.Writer, "event: complete\ndata: %s\n\n", string(data))
			c.Writer.Flush()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case telemetry, ok := <-ch:
				if !ok {
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
