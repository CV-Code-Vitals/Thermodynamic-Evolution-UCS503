package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	mutator := &agent.MutationAgent{
		Client:  &http.Client{},
		APIBase: os.Getenv("MUTATION_API_BASE"),
		APIKey:  os.Getenv("MUTATION_API_KEY"),
		Model:   os.Getenv("MUTATION_MODEL"),
	}
	if mutator.APIBase == "" {
		mutator.APIBase = "https://api.openai.com" // default
	}

	annealer := &agent.Annealer{
		Mutator: mutator,
	}

	return func(c *gin.Context) {
		var req EvolveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		ctx := c.Request.Context()
		ch := make(chan agent.StepTelemetry)

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
