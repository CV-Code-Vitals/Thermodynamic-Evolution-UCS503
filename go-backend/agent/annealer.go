package agent

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// CodeMutator defines the interface for generating candidate code patches.
type CodeMutator interface {
	ProposeMutation(ctx context.Context, currentCode, funcName string) (string, error)
}

// AnnealingConfig defines the hyperparameters for simulated annealing.
type AnnealingConfig struct {
	InitialTemp float64 `json:"initial_temp"`
	CoolingRate float64 `json:"cooling_rate"`
	MaxSteps    int     `json:"max_steps"`
}

// StepTelemetry captures per-iteration telemetry for the live evolution stream.
type StepTelemetry struct {
	Step        int     `json:"step"`
	Temp        float64 `json:"temp"`
	OldEnergy   float64 `json:"old_energy"`
	NewEnergy   float64 `json:"new_energy"`
	DeltaE      float64 `json:"delta_e"`
	TestsPassed bool    `json:"tests_passed"`
	Accepted    bool    `json:"accepted"`
	Reason      string  `json:"reason"`
	CodePatch   string  `json:"code_patch,omitempty"`
}

// AnnealingResult contains the final accepted code and metrics.
type AnnealingResult struct {
	OptimizedCode string  `json:"optimized_code"`
	EnergyDelta   float64 `json:"energy_delta"`
	TotalSteps    int     `json:"total_steps"`
}

// Annealer orchestrates code optimization through simulated annealing.
// All candidate evaluation is strictly mediated by Evaluator — no direct host execution occurs.
type Annealer struct {
	Mutator   CodeMutator
	Evaluator CandidateEvaluator
}

// Evolve runs the simulated annealing loop against a target function.
// It generates candidates, delegates evaluation to the isolated sandbox boundary,
// and only updates disk if an optimization is accepted.
func (a *Annealer) Evolve(ctx context.Context, repoPath, filePath, funcName string, startByte, endByte int, baseEnergy float64, cfg AnnealingConfig, ch chan<- StepTelemetry) (*AnnealingResult, error) {
	if a.Evaluator == nil {
		return nil, fmt.Errorf("%w: annealer requires a configured CandidateEvaluator", ErrHostExecutionBlocked)
	}
	if a.Mutator == nil {
		return nil, errors.New("annealer requires a configured CodeMutator")
	}

	// Validate path boundaries
	cleanedRel := filepath.Clean(filepath.FromSlash(filePath))
	if filepath.IsAbs(cleanedRel) || strings.HasPrefix(cleanedRel, "..") {
		return nil, fmt.Errorf("%w: invalid target file path '%s'", ErrSecurityViolation, filePath)
	}

	fullPath := filepath.Join(repoPath, cleanedRel)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read target file: %w", err)
	}

	if startByte < 0 || endByte < startByte || startByte > len(contentBytes) || endByte > len(contentBytes) {
		return nil, fmt.Errorf("%w: byte range [%d:%d] out of range for file size %d", ErrInvalidRequest, startByte, endByte, len(contentBytes))
	}

	currentCode := string(contentBytes[startByte:endByte])
	bestCode := currentCode
	initialCode := currentCode

	currentTemp := cfg.InitialTemp
	currentEnergy := baseEnergy
	workingFileBytes := append([]byte{}, contentBytes...)

	for step := 1; step <= cfg.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		telemetry := StepTelemetry{Step: step, Temp: currentTemp, OldEnergy: currentEnergy}

		// 1. Propose mutation via LLM / agent mutator
		mutatedCode, err := a.Mutator.ProposeMutation(ctx, currentCode, funcName)
		if err != nil {
			telemetry.Reason = "mutation_failed: " + err.Error()
			ch <- telemetry
			continue
		}

		// 2. Evaluate candidate code through the isolated execution boundary
		// Notice: host files are NEVER modified during this evaluation!
		evalReq := EvaluationRequest{
			RepoPath:      repoPath,
			FilePath:      cleanedRel,
			OriginalBytes: workingFileBytes,
			MutatedCode:   mutatedCode,
			StartByte:     startByte,
			EndByte:       endByte,
			Limits:        DefaultResourceLimits(),
		}

		evalRes, err := a.Evaluator.EvaluateCandidate(ctx, evalReq)
		if err != nil {
			telemetry.TestsPassed = false
			telemetry.Accepted = false
			telemetry.Reason = "evaluation_error: " + err.Error()
			ch <- telemetry

			// If the sandbox is unavailable, fail closed immediately — do not silently loop or fallback!
			if errors.Is(err, ErrSandboxUnavailable) || errors.Is(err, ErrHostExecutionBlocked) {
				return nil, fmt.Errorf("secure execution failed: %w", err)
			}
			continue
		}

		// 3. Process test results
		if !evalRes.TestsPassed {
			telemetry.TestsPassed = false
			telemetry.Accepted = false
			telemetry.Reason = string(evalRes.Outcome)
		} else {
			telemetry.TestsPassed = true

			// Simulated energy reduction (until inline AST evaluation is integrated in Milestone 1)
			newEnergy := currentEnergy - rand.Float64()*5.0 + 1.0
			deltaE := newEnergy - currentEnergy
			telemetry.NewEnergy = newEnergy
			telemetry.DeltaE = deltaE
			telemetry.CodePatch = mutatedCode

			if deltaE < 0 {
				telemetry.Accepted = true
				telemetry.Reason = "improvement"
				currentEnergy = newEnergy
				currentCode = mutatedCode
				bestCode = currentCode
				workingFileBytes = evalRes.UpdatedFileBytes
				endByte = startByte + len(mutatedCode)
			} else {
				// Metropolis acceptance criterion for stochastic exploration
				p := math.Exp(-deltaE / currentTemp)
				if rand.Float64() < p {
					telemetry.Accepted = true
					telemetry.Reason = "stochastic_acceptance"
					currentEnergy = newEnergy
					currentCode = mutatedCode
					workingFileBytes = evalRes.UpdatedFileBytes
					endByte = startByte + len(mutatedCode)
				} else {
					telemetry.Accepted = false
					telemetry.Reason = "rejected"
				}
			}
		}

		ch <- telemetry
		currentTemp *= cfg.CoolingRate
	}

	// 4. If an improvement was accepted, safely persist the verified code to disk
	if bestCode != initialCode {
		if err := os.WriteFile(fullPath, workingFileBytes, 0o644); err != nil {
			return nil, fmt.Errorf("failed to persist accepted mutation: %w", err)
		}
	}

	return &AnnealingResult{
		OptimizedCode: bestCode,
		EnergyDelta:   baseEnergy - currentEnergy,
		TotalSteps:    cfg.MaxSteps,
	}, nil
}
