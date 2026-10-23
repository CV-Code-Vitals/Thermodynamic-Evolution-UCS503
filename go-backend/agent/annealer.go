package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CodeMutator defines the interface for generating candidate code patches.
type CodeMutator interface {
	ProposeMutation(ctx context.Context, currentCode, funcName string) (string, error)
}

// AnnealingConfig defines the hyperparameters for simulated annealing.
type AnnealingConfig struct {
	InitialTemp              float64 `json:"initial_temp"`
	CoolingRate              float64 `json:"cooling_rate"`
	MaxSteps                 int     `json:"max_steps"`
	TargetEnergy             float64 `json:"target_energy,omitempty"`
	MinTemp                  float64 `json:"min_temp,omitempty"`
	EnableDuplicateDetection bool    `json:"enable_duplicate_detection,omitempty"`
}

// StepTelemetry captures per-iteration telemetry for the live evolution stream.
type StepTelemetry struct {
	Step            int              `json:"step"`
	Temp            float64          `json:"temp"`
	OldEnergy       float64          `json:"old_energy"`
	NewEnergy       float64          `json:"new_energy"`
	DeltaE          float64          `json:"delta_e"`
	BestEnergy      float64          `json:"best_energy"`
	BaselineEnergy  float64          `json:"baseline_energy"`
	CandidateEnergy float64          `json:"candidate_energy"`
	AcceptanceProb  float64          `json:"acceptance_prob,omitempty"`
	IsBestSoFar     bool             `json:"is_best_so_far"`
	TestsPassed     bool             `json:"tests_passed"`
	Accepted        bool             `json:"accepted"`
	Reason          string           `json:"reason"`
	CodePatch       string           `json:"code_patch,omitempty"`
	SourceHash      string           `json:"source_hash,omitempty"`
	Metrics         *FunctionMetrics `json:"metrics,omitempty"`
	CandidateID     string           `json:"candidate_id,omitempty"`
	Strategy        string           `json:"strategy,omitempty"`
}

// AnnealingResult contains the final accepted code and metrics.
type AnnealingResult struct {
	OptimizedCode      string                   `json:"optimized_code"`
	BaselineEnergy     float64                  `json:"baseline_energy"`
	FinalEnergy        float64                  `json:"final_energy"`
	EnergyDelta        float64                  `json:"energy_delta"`
	TotalSteps         int                      `json:"total_steps"`
	AcceptedMoves      int                      `json:"accepted_moves"`
	BestImproved       bool                     `json:"best_improved"`
	TerminationReason  string                   `json:"termination_reason"`
	ProposalsGenerated int                      `json:"proposals_generated"`
	CorrectnessPassed  int                      `json:"correctness_passed"`
	CorrectnessFailed  int                      `json:"correctness_failed"`
	DuplicatesSkipped  int                      `json:"duplicates_skipped"`
	StrategyStats      map[string]StrategyStats `json:"strategy_stats,omitempty"`
	InitialSourceHash  string                   `json:"initial_source_hash"`
	FinalSourceHash    string                   `json:"final_source_hash"`
}

// Annealer orchestrates code optimization through simulated annealing.
// All candidate evaluation is strictly mediated by Evaluator, and all energy
// metrics are measured by the real AST EnergyEvaluator.
type Annealer struct {
	Mutator         CodeMutator
	Evaluator       CandidateEvaluator
	EnergyEvaluator EnergyEvaluator
	RandSource      *rand.Rand
}

func (a *Annealer) rng() *rand.Rand {
	if a.RandSource != nil {
		return a.RandSource
	}
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Evolve runs the simulated annealing loop against a target function.
// It measures a real AST baseline energy, iteratively proposes mutations,
// gates evaluation through the container sandbox boundary, calculates genuine
// energy differentials using the Rust engine, and updates state according to
// mathematically sound Metropolis simulated annealing rules.
func (a *Annealer) Evolve(ctx context.Context, repoPath, filePath, funcName string, startByte, endByte int, baseEnergy float64, cfg AnnealingConfig, ch chan<- StepTelemetry) (*AnnealingResult, error) {
	if a.Evaluator == nil {
		return nil, fmt.Errorf("%w: annealer requires a configured CandidateEvaluator", ErrHostExecutionBlocked)
	}
	if a.Mutator == nil {
		return nil, errors.New("annealer requires a configured CodeMutator")
	}
	if a.EnergyEvaluator == nil {
		return nil, errors.New("annealer requires a configured EnergyEvaluator")
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

	// 1. Establish baseline energy using real AST engine
	hInit := sha256.Sum256(contentBytes)
	initialSourceHash := hex.EncodeToString(hInit[:])

	baselineMeas, err := a.EnergyEvaluator.EvaluateEnergy(ctx, contentBytes, cleanedRel, funcName)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate baseline structural energy: %w", err)
	}
	baselineEnergy := baselineMeas.Energy
	if baselineEnergy <= 0.0 {
		return nil, fmt.Errorf("%w: baseline energy %.4f is non-positive or invalid", ErrInvalidEnergy, baselineEnergy)
	}

	currentCode := string(contentBytes[startByte:endByte])
	initialCode := currentCode

	// Separate representations:
	// - current state: actively explored state in annealing (may accept uphill moves)
	// - best state: lowest valid energy discovered so far
	currentEnergy := baselineEnergy
	bestEnergy := baselineEnergy
	bestCode := currentCode

	workingFileBytes := append([]byte{}, contentBytes...)
	bestFileBytes := append([]byte{}, contentBytes...)

	currentTemp := cfg.InitialTemp
	if currentTemp <= 0.0 || math.IsNaN(currentTemp) || math.IsInf(currentTemp, 0) {
		currentTemp = 100.0
	}
	coolingRate := cfg.CoolingRate
	if coolingRate <= 0.0 || coolingRate >= 1.0 || math.IsNaN(coolingRate) || math.IsInf(coolingRate, 0) {
		coolingRate = 0.85
	}

	// Search metrics tracking
	acceptedMoves := 0
	proposalsGenerated := 0
	correctnessPassed := 0
	correctnessFailed := 0
	duplicatesSkipped := 0
	terminationReason := "max_steps_reached"

	// Duplicate detection set seeded with baseline
	seenCandidateHashes := map[string]bool{initialSourceHash: true}

	targetedMutator, isTargeted := a.Mutator.(TargetedMutator)

	for step := 1; step <= cfg.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check search budget termination conditions
		if cfg.MinTemp > 0 && currentTemp < cfg.MinTemp {
			terminationReason = "temperature_exhausted"
			break
		}
		if cfg.TargetEnergy > 0 && bestEnergy <= cfg.TargetEnergy {
			terminationReason = "target_energy_reached"
			break
		}

		telemetry := StepTelemetry{
			Step:           step,
			Temp:           currentTemp,
			OldEnergy:      currentEnergy,
			BestEnergy:     bestEnergy,
			BaselineEnergy: baselineEnergy,
		}

		// Current state snapshot hash
		hCur := sha256.Sum256(workingFileBytes)
		currentSourceHash := hex.EncodeToString(hCur[:])

		var mutatedCode string
		var strategyName string
		var candidateID string

		proposalsGenerated++

		if isTargeted {
			candidateID = fmt.Sprintf("cand-%d-%s", step, uuid.New().String()[:8])
			req := MutationRequest{
				CandidateID:      candidateID,
				BaseSnapshotHash: currentSourceHash,
				FilePath:         cleanedRel,
				FunctionName:     funcName,
				StartByte:        startByte,
				EndByte:          endByte,
				CurrentCode:      currentCode,
				FullContent:      workingFileBytes,
				Iteration:        step,
				Temperature:      currentTemp,
			}
			proposal, propErr := targetedMutator.ProposeTargeted(ctx, req)
			if propErr != nil {
				telemetry.Reason = "mutation_failed: " + propErr.Error()
				if !sendTelemetry(ctx, ch, telemetry) {
					return nil, ctx.Err()
				}
				currentTemp *= coolingRate
				continue
			}

			// Snapshot integrity: verify proposal was derived from current state
			if proposal.BaseSnapshotHash != currentSourceHash {
				telemetry.Reason = fmt.Sprintf("stale_proposal_rejected: base hash mismatch (expected %s, got %s)",
					currentSourceHash[:8], proposal.BaseSnapshotHash[:8])
				if !sendTelemetry(ctx, ch, telemetry) {
					return nil, ctx.Err()
				}
				currentTemp *= coolingRate
				continue
			}

			mutatedCode = proposal.MutatedCode
			strategyName = proposal.StrategyName
			candidateID = proposal.CandidateID
		} else {
			candidateID = fmt.Sprintf("cand-%d", step)
			strategyName = "legacy_mutator"
			mutCode, mutErr := a.Mutator.ProposeMutation(ctx, currentCode, funcName)
			if mutErr != nil {
				telemetry.Reason = "mutation_failed: " + mutErr.Error()
				if !sendTelemetry(ctx, ch, telemetry) {
					return nil, ctx.Err()
				}
				currentTemp *= coolingRate
				continue
			}
			mutatedCode = mutCode
		}

		telemetry.CandidateID = candidateID
		telemetry.Strategy = strategyName

		// Check duplicate candidate before container execution
		prospectiveBytes := make([]byte, 0, startByte+len(mutatedCode)+(len(workingFileBytes)-endByte))
		prospectiveBytes = append(prospectiveBytes, workingFileBytes[:startByte]...)
		prospectiveBytes = append(prospectiveBytes, []byte(mutatedCode)...)
		prospectiveBytes = append(prospectiveBytes, workingFileBytes[endByte:]...)

		hCand := sha256.Sum256(prospectiveBytes)
		candHash := hex.EncodeToString(hCand[:])

		if cfg.EnableDuplicateDetection && seenCandidateHashes[candHash] {
			duplicatesSkipped++
			telemetry.Accepted = false
			telemetry.Reason = "duplicate_candidate_skipped"
			telemetry.SourceHash = candHash
			if !sendTelemetry(ctx, ch, telemetry) {
				return nil, ctx.Err()
			}
			currentTemp *= coolingRate
			continue
		}
		seenCandidateHashes[candHash] = true

		// 3. Evaluate candidate correctness in secure sandbox boundary
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
			correctnessFailed++
			telemetry.TestsPassed = false
			telemetry.Accepted = false
			telemetry.Reason = "evaluation_error: " + err.Error()
			if !sendTelemetry(ctx, ch, telemetry) {
				return nil, ctx.Err()
			}

			if errors.Is(err, ErrSandboxUnavailable) || errors.Is(err, ErrHostExecutionBlocked) {
				return nil, fmt.Errorf("secure execution failed: %w", err)
			}
			currentTemp *= coolingRate
			continue
		}

		if !evalRes.TestsPassed {
			correctnessFailed++
			telemetry.TestsPassed = false
			telemetry.Accepted = false
			telemetry.Reason = string(evalRes.Outcome)
			if !sendTelemetry(ctx, ch, telemetry) {
				return nil, ctx.Err()
			}
			currentTemp *= coolingRate
			continue
		}

		correctnessPassed++
		telemetry.TestsPassed = true

		// 4. Evaluate actual structural energy using the Rust AST engine
		energyMeas, err := a.EnergyEvaluator.EvaluateEnergy(ctx, evalRes.UpdatedFileBytes, cleanedRel, funcName)
		if err != nil {
			telemetry.Accepted = false
			telemetry.Reason = "energy_evaluation_failed: " + err.Error()
			if !sendTelemetry(ctx, ch, telemetry) {
				return nil, ctx.Err()
			}
			currentTemp *= coolingRate
			continue
		}

		candidateEnergy := energyMeas.Energy
		telemetry.CandidateEnergy = candidateEnergy
		telemetry.SourceHash = energyMeas.SourceHash
		telemetry.Metrics = &energyMeas.Metrics
		telemetry.CodePatch = mutatedCode

		// 5. Calculate genuine energy differential: lower is better!
		deltaE := candidateEnergy - currentEnergy
		telemetry.DeltaE = deltaE
		telemetry.NewEnergy = candidateEnergy

		// 6. Simulated annealing acceptance rule
		if deltaE <= 0.0 {
			// Downhill transition: structural improvement
			telemetry.Accepted = true
			telemetry.Reason = "improvement"
			currentEnergy = candidateEnergy
			currentCode = mutatedCode
			workingFileBytes = evalRes.UpdatedFileBytes
			endByte = startByte + len(mutatedCode)
			acceptedMoves++

			// Track best-so-far valid candidate
			if candidateEnergy < bestEnergy {
				bestEnergy = candidateEnergy
				bestCode = mutatedCode
				bestFileBytes = evalRes.UpdatedFileBytes
				telemetry.IsBestSoFar = true

				if isTargeted {
					if adaptive, ok := a.Mutator.(*AdaptiveMutator); ok {
						adaptive.RecordOutcome(strategyName, true, true)
					}
				}
			}
		} else {
			// Uphill transition: complexity increased
			// Metropolis acceptance probability: p = exp(-deltaE / T)
			p := 0.0
			if !math.IsNaN(currentTemp) && !math.IsInf(currentTemp, 0) && currentTemp > 1e-6 {
				exponent := -deltaE / currentTemp
				if exponent < -700.0 {
					p = 0.0 // Underflow guard: exp(-700) ~ 1e-304
				} else {
					p = math.Exp(exponent)
				}
			}
			telemetry.AcceptanceProb = p

			if a.rng().Float64() < p {
				// Stochastically accepted into current state to escape local minima
				telemetry.Accepted = true
				telemetry.Reason = "stochastic_acceptance"
				currentEnergy = candidateEnergy
				currentCode = mutatedCode
				workingFileBytes = evalRes.UpdatedFileBytes
				endByte = startByte + len(mutatedCode)
				acceptedMoves++
				// Best state is NOT overwritten by an uphill transition
			} else {
				telemetry.Accepted = false
				telemetry.Reason = "rejected"
			}
		}

		if !sendTelemetry(ctx, ch, telemetry) {
			return nil, ctx.Err()
		}
		currentTemp *= coolingRate
	}

	// 7. Persist ONLY if the best discovered candidate improved over the initial baseline
	bestImproved := bestEnergy < baselineEnergy && bestCode != initialCode
	if bestImproved {
		if err := AtomicWriteFileWithConflictDetection(fullPath, bestFileBytes, initialSourceHash, 0o644); err != nil {
			return nil, fmt.Errorf("failed to persist accepted mutation: %w", err)
		}
	}

	hFinal := sha256.Sum256(bestFileBytes)
	finalSourceHash := hex.EncodeToString(hFinal[:])

	var stratStats map[string]StrategyStats
	if isTargeted {
		stratStats = targetedMutator.GetStrategyStats()
	}

	return &AnnealingResult{
		OptimizedCode:      bestCode,
		BaselineEnergy:     baselineEnergy,
		FinalEnergy:        bestEnergy,
		EnergyDelta:        baselineEnergy - bestEnergy, // Positive value represents energy reduced (improvement)
		TotalSteps:         cfg.MaxSteps,
		AcceptedMoves:      acceptedMoves,
		BestImproved:       bestImproved,
		TerminationReason:  terminationReason,
		ProposalsGenerated: proposalsGenerated,
		CorrectnessPassed:  correctnessPassed,
		CorrectnessFailed:  correctnessFailed,
		DuplicatesSkipped:  duplicatesSkipped,
		StrategyStats:      stratStats,
		InitialSourceHash:  initialSourceHash,
		FinalSourceHash:    finalSourceHash,
	}, nil
}

// AtomicWriteFileWithConflictDetection atomically replaces targetPath with data,
// detecting concurrent modifications by verifying expectedBaselineHash against disk contents.
func AtomicWriteFileWithConflictDetection(targetPath string, data []byte, expectedBaselineHash string, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent directory for atomic write: %w", err)
	}

	// 1. Write to an ephemeral temporary file in the same directory first
	tmpFile, err := os.CreateTemp(dir, ".ucs503-persist-*")
	if err != nil {
		return fmt.Errorf("create temporary file for atomic write: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName) // Clean up if rename fails or conflict occurs

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write candidate data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return fmt.Errorf("sync candidate file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}

	_ = os.Chmod(tmpName, perm)

	// 2. Conflict detection: read disk content as close to atomic rename as practical
	currentDiskBytes, err := os.ReadFile(targetPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read target file for conflict check: %w", err)
		}
	} else {
		h := sha256.Sum256(currentDiskBytes)
		currentHash := hex.EncodeToString(h[:])
		if expectedBaselineHash != "" && currentHash != expectedBaselineHash {
			return fmt.Errorf("%w: file on disk was modified externally (expected hash %s, found %s)",
				ErrConcurrentModification, expectedBaselineHash, currentHash)
		}
	}

	// 3. Atomically replace targetPath
	if err := os.Rename(tmpName, targetPath); err != nil {
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}

// sendTelemetry sends a step event on ch with context cancellation protection to prevent goroutine leaks.
func sendTelemetry(ctx context.Context, ch chan<- StepTelemetry, telem StepTelemetry) bool {
	if ch == nil {
		return true
	}
	select {
	case ch <- telem:
		return true
	case <-ctx.Done():
		return false
	}
}
