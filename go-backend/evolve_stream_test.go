package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent"
)

func TestEvolutionSSE_EmitsHeartbeatWhileWaiting(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	telemetry := make(chan agent.StepTelemetry)
	outcomes := make(chan evolutionOutcome)
	response := httptest.NewRecorder()

	streamEvolutionEvents(ctx, response, telemetry, outcomes, time.Millisecond)

	if !strings.Contains(response.Body.String(), ": heartbeat\n\n") {
		t.Fatalf("idle Evolution stream did not emit an SSE heartbeat: %q", response.Body.String())
	}
}

func TestEvolutionSSE_EmitsCompletionAndErrorEvents(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		outcome evolutionOutcome
		event   string
	}{
		{name: "completion", outcome: evolutionOutcome{result: &agent.AnnealingResult{}}, event: "event: complete"},
		{name: "error", outcome: evolutionOutcome{err: errors.New("failed")}, event: "event: error"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			telemetry := make(chan agent.StepTelemetry)
			outcomes := make(chan evolutionOutcome, 1)
			outcomes <- testCase.outcome
			response := httptest.NewRecorder()

			streamEvolutionEvents(ctx, response, telemetry, outcomes, time.Minute)

			if !strings.Contains(response.Body.String(), testCase.event) {
				t.Fatalf("stream %q does not contain %q", response.Body.String(), testCase.event)
			}
		})
	}
}

func TestEvolutionStreamRuntimeFitsGatewayLimits(t *testing.T) {
	if maxEvolutionDuration >= 15*time.Minute {
		t.Fatalf("Evolution runtime limit = %s, must be below 15 minutes", maxEvolutionDuration)
	}
	if evolutionHeartbeatInterval >= 5*time.Minute {
		t.Fatalf("Evolution heartbeat interval = %s, must be below 5 minutes", evolutionHeartbeatInterval)
	}
}
