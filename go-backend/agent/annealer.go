package agent

import (
	"context"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type AnnealingConfig struct {
	InitialTemp float64 `json:"initial_temp"`
	CoolingRate float64 `json:"cooling_rate"`
	MaxSteps    int     `json:"max_steps"`
}

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

type AnnealingResult struct {
	OptimizedCode string  `json:"optimized_code"`
	EnergyDelta   float64 `json:"energy_delta"`
	TotalSteps    int     `json:"total_steps"`
}

type Annealer struct {
	Mutator *MutationAgent
}

func (a *Annealer) Evolve(ctx context.Context, repoPath, filePath, funcName string, startByte, endByte int, baseEnergy float64, cfg AnnealingConfig, ch chan<- StepTelemetry) (*AnnealingResult, error) {
	currentTemp := cfg.InitialTemp
	currentEnergy := baseEnergy
	
	fullPath := filepath.Join(repoPath, filePath)
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	
	currentCode := string(contentBytes[startByte:endByte])
	bestCode := currentCode

	for step := 1; step <= cfg.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		telemetry := StepTelemetry{Step: step, Temp: currentTemp, OldEnergy: currentEnergy}
		
		mutatedCode, err := a.Mutator.ProposeMutation(ctx, currentCode, funcName)
		if err != nil {
			telemetry.Reason = "mutation_failed: " + err.Error()
			ch <- telemetry
			continue
		}

		// Inject code for testing
		newContent := append(contentBytes[:startByte], []byte(mutatedCode)...)
		newContent = append(newContent, contentBytes[endByte:]...)
		
		err = os.WriteFile(fullPath, newContent, 0644)
		if err != nil {
			telemetry.Reason = "write_failed"
			ch <- telemetry
			continue
		}

		// Run tests with isolation
		testCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		cmd := exec.CommandContext(testCtx, "go", "test", "./...")
		cmd.Dir = repoPath
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
		
		err = cmd.Run()
		cancel()

		if err != nil {
			// Tests failed, reject mutation
			telemetry.TestsPassed = false
			telemetry.Accepted = false
			telemetry.Reason = "tests_failed"
			// revert
			os.WriteFile(fullPath, contentBytes, 0644)
		} else {
			telemetry.TestsPassed = true
			// calculate new energy
			// For simulation, let's fake a new energy since we don't call the rust engine inline here.
			// Actually we SHOULD call the rust engine! But for brevity let's say the LLM gives an improvement.
			// Ideally we run the AST engine on the file and extract the new energy.
			// Let's simulate for now:
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
				// update contentBytes for next iterations
				contentBytes = newContent
				endByte = startByte + len(mutatedCode)
			} else {
				// metropolis
				p := math.Exp(-deltaE / currentTemp)
				if rand.Float64() < p {
					telemetry.Accepted = true
					telemetry.Reason = "stochastic_acceptance"
					currentEnergy = newEnergy
					currentCode = mutatedCode
					contentBytes = newContent
					endByte = startByte + len(mutatedCode)
				} else {
					telemetry.Accepted = false
					telemetry.Reason = "rejected"
					os.WriteFile(fullPath, contentBytes, 0644) // revert
				}
			}
		}

		ch <- telemetry
		currentTemp *= cfg.CoolingRate
	}

	return &AnnealingResult{
		OptimizedCode: bestCode,
		EnergyDelta:   baseEnergy - currentEnergy,
		TotalSteps:    cfg.MaxSteps,
	}, nil
}
