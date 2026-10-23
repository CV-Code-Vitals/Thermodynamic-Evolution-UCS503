package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Typed errors for energy evaluation.
var (
	ErrEngineNotFound    = errors.New("thermodynamic AST engine executable not found")
	ErrEngineFailed      = errors.New("thermodynamic AST engine execution failed")
	ErrFunctionNotFound  = errors.New("target function not found in AST analysis report")
	ErrInvalidEnergy     = errors.New("invalid or non-finite structural energy score")
	ErrCorruptReportJSON = errors.New("failed to parse thermodynamic report JSON")
)

// EnergyMeasurement represents an immutable, typed structural energy measurement.
type EnergyMeasurement struct {
	Energy        float64         `json:"energy"`
	Metrics       FunctionMetrics `json:"metrics"`
	FunctionName  string          `json:"function_name"`
	FilePath      string          `json:"file_path"`
	EngineVersion string          `json:"engine_version"`
	Duration      time.Duration   `json:"duration"`
	SourceHash    string          `json:"source_hash"`
}

// EnergyEvaluator defines the interface for calculating structural thermodynamic energy.
type EnergyEvaluator interface {
	EvaluateEnergy(ctx context.Context, sourceContent []byte, relFilePath string, funcName string) (*EnergyMeasurement, error)
}

// RustEngineEvaluator implements EnergyEvaluator by invoking the compiled Rust thermodynamic-ast-engine CLI.
type RustEngineEvaluator struct {
	EnginePath    string
	Timeout       time.Duration
	Cache         *EnergyCache
	EngineVersion string
}

// EnergyCache provides bounded in-memory caching for deterministic AST energy measurements.
type EnergyCache struct {
	mu      sync.RWMutex
	entries map[string]*EnergyMeasurement
	maxSize int
}

// NewEnergyCache creates an EnergyCache with bounded entry capacity.
func NewEnergyCache(maxSize int) *EnergyCache {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &EnergyCache{
		entries: make(map[string]*EnergyMeasurement),
		maxSize: maxSize,
	}
}

// Get retrieves a cached measurement, returning a copy to prevent mutation.
func (c *EnergyCache) Get(key string) (*EnergyMeasurement, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	copyMeas := *m
	return &copyMeas, true
}

// Put stores a copy of the measurement in cache, evicting an entry if at max capacity.
func (c *EnergyCache) Put(key string, m *EnergyMeasurement) {
	if c == nil || m == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
	copyMeas := *m
	c.entries[key] = &copyMeas
}

// Size returns the current number of cached evaluations.
func (c *EnergyCache) Size() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// NewRustEngineEvaluator constructs an evaluator with automatic engine discovery.
func NewRustEngineEvaluator(customPath string) (*RustEngineEvaluator, error) {
	path := customPath
	if strings.TrimSpace(path) == "" {
		path = ResolveEnginePath()
	}

	// Verify engine executable exists
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return nil, fmt.Errorf("%w at '%s'", ErrEngineNotFound, path)
	}

	return &RustEngineEvaluator{
		EnginePath:    path,
		Timeout:       10 * time.Second,
		Cache:         NewEnergyCache(512),
		EngineVersion: "2.0.0",
	}, nil
}

// LazyRustEngineEvaluator dynamically resolves the engine path on demand, enabling runtime recovery.
type LazyRustEngineEvaluator struct {
	CustomPath string
	Timeout    time.Duration
}

// EvaluateEnergy evaluates energy by resolving the engine path at invocation time.
func (l *LazyRustEngineEvaluator) EvaluateEnergy(ctx context.Context, sourceContent []byte, relFilePath string, funcName string) (*EnergyMeasurement, error) {
	evaluator, err := NewRustEngineEvaluator(l.CustomPath)
	if err != nil {
		return nil, err
	}
	if l.Timeout > 0 {
		evaluator.Timeout = l.Timeout
	}
	return evaluator.EvaluateEnergy(ctx, sourceContent, relFilePath, funcName)
}

// EvaluateEnergy evaluates the thermodynamic energy of a target function in a candidate snapshot.
func (r *RustEngineEvaluator) EvaluateEnergy(ctx context.Context, sourceContent []byte, relFilePath string, funcName string) (*EnergyMeasurement, error) {
	start := time.Now()

	if len(sourceContent) == 0 {
		return nil, fmt.Errorf("%w: source content is empty", ErrInvalidEnergy)
	}
	if strings.TrimSpace(relFilePath) == "" {
		return nil, errors.New("relFilePath is required")
	}

	// Calculate deterministic source hash
	h := sha256.Sum256(sourceContent)
	sourceHash := hex.EncodeToString(h[:])

	cleanRel := filepath.Clean(filepath.FromSlash(relFilePath))

	version := r.EngineVersion
	if version == "" {
		version = "unknown"
	}

	// Fast cache lookup
	cacheKey := fmt.Sprintf("%s:%s:%s:v=%s", sourceHash, cleanRel, funcName, version)
	if r.Cache != nil {
		if cached, ok := r.Cache.Get(cacheKey); ok {
			cached.Duration = time.Since(start)
			return cached, nil
		}
	}

	// 1. Prepare sanitized temporary directory for engine analysis
	tempDir, err := os.MkdirTemp("", "ucs503-ast-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary analysis dir: %w", err)
	}
	defer os.RemoveAll(tempDir) // Guarantee cleanup

	targetDest := filepath.Join(tempDir, cleanRel)
	if err := os.MkdirAll(filepath.Dir(targetDest), 0o750); err != nil {
		return nil, fmt.Errorf("create target dir: %w", err)
	}
	if err := os.WriteFile(targetDest, sourceContent, 0o600); err != nil {
		return nil, fmt.Errorf("write candidate source: %w", err)
	}

	// 2. Invoke Rust thermodynamic-ast-engine binary
	reportPath := filepath.Join(tempDir, fmt.Sprintf("report_%s.json", uuid.New().String()[:8]))

	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	evalCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(evalCtx,
		r.EnginePath,
		tempDir,
		"--output", reportPath,
		"--min-score", "0.0", // Include all functions regardless of score
	)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		if errors.Is(evalCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: AST engine timed out after %v", ErrEngineFailed, timeout)
		}
		return nil, fmt.Errorf("%w: %v (stderr: %s, stdout: %s)", ErrEngineFailed, cmdErr, errBuf.String(), string(output))
	}

	// 3. Read and parse thermodynamic report
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("%w: report file missing: %v", ErrEngineFailed, err)
	}

	var report ThermodynamicReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptReportJSON, err)
	}

	// 4. Locate target file and function metrics
	measurement, err := extractHotspotMetric(&report, cleanRel, funcName)
	if err != nil {
		return nil, err
	}

	// 5. Numerical validity checks
	if math.IsNaN(measurement.Energy) || math.IsInf(measurement.Energy, 0) || measurement.Energy <= 0.0 {
		return nil, fmt.Errorf("%w: energy score %.4f is non-finite or non-positive", ErrInvalidEnergy, measurement.Energy)
	}

	measurement.Duration = time.Since(start)
	measurement.SourceHash = sourceHash
	measurement.EngineVersion = report.EngineVersion
	measurement.FilePath = cleanRel

	if report.EngineVersion != "" && report.EngineVersion != version {
		cacheKey = fmt.Sprintf("%s:%s:%s:v=%s", sourceHash, cleanRel, funcName, report.EngineVersion)
	}

	// Populate cache
	if r.Cache != nil {
		r.Cache.Put(cacheKey, measurement)
	}

	return measurement, nil
}

// pathMatches checks if a reported file path matches the target relative path using strict path-separator boundaries.
func pathMatches(reportedPath, targetRelPath string) bool {
	cleanReported := filepath.Clean(filepath.FromSlash(reportedPath))
	cleanTarget := filepath.Clean(filepath.FromSlash(targetRelPath))

	if cleanReported == cleanTarget {
		return true
	}

	sep := string(filepath.Separator)

	// cleanReported is absolute or prefixed (e.g. C:\tmp\target\file.go) and ends with \target\file.go
	if strings.HasSuffix(cleanReported, sep+cleanTarget) {
		return true
	}

	// cleanTarget may have relative or leading separator
	trimmedTarget := strings.TrimPrefix(cleanTarget, "."+sep)
	trimmedTarget = strings.TrimPrefix(trimmedTarget, sep)
	if strings.HasSuffix(cleanReported, sep+trimmedTarget) {
		return true
	}

	// If cleanReported is a relative subpath matching target exactly at boundary
	if strings.HasSuffix(cleanTarget, sep+cleanReported) {
		return true
	}

	return false
}

// extractHotspotMetric locates the specific function or file in the report.
func extractHotspotMetric(report *ThermodynamicReport, targetRelPath string, funcName string) (*EnergyMeasurement, error) {
	for _, fileReport := range report.FileReports {
		if pathMatches(fileReport.FilePath, targetRelPath) {
			// If a specific function is requested, find it in the hotspots
			if funcName != "" {
				for _, hotspot := range fileReport.Hotspots {
					if hotspot.FunctionName == funcName {
						return &EnergyMeasurement{
							Energy:       hotspot.EntropyScore,
							Metrics:      hotspot.Metrics,
							FunctionName: hotspot.FunctionName,
						}, nil
					}
				}
				return nil, fmt.Errorf("%w: '%s' in file '%s'", ErrFunctionNotFound, funcName, targetRelPath)
			}

			// If no function specified, return file aggregate entropy
			return &EnergyMeasurement{
				Energy:       fileReport.TotalEntropy,
				FunctionName: "<file_aggregate>",
			}, nil
		}
	}

	return nil, fmt.Errorf("%w: file '%s' not found in AST analysis report", ErrFunctionNotFound, targetRelPath)
}

// ResolveEnginePath locates the compiled thermodynamic-ast-engine binary across OS platforms and common paths.
func ResolveEnginePath() string {
	if custom := os.Getenv("ENGINE_PATH"); custom != "" {
		return custom
	}

	binaryName := "thermodynamic-ast-engine"
	if runtime.GOOS == "windows" {
		binaryName = "thermodynamic-ast-engine.exe"
	}

	candidates := []string{
		filepath.Join(".", binaryName),
		filepath.Join("..", "thermodynamic-ast-engine", "target", "release", binaryName),
		filepath.Join("..", "thermodynamic-ast-engine", "target", "debug", binaryName),
		filepath.Join("..", "..", "thermodynamic-ast-engine", "target", "release", binaryName),
		filepath.Join("..", "..", "thermodynamic-ast-engine", "target", "debug", binaryName),
		filepath.Join(".", "thermodynamic-ast-engine", "target", "release", binaryName),
		filepath.Join(".", "thermodynamic-ast-engine", "target", "debug", binaryName),
		filepath.Join("thermodynamic-ast-engine", "target", "release", binaryName),
		filepath.Join("thermodynamic-ast-engine", "target", "debug", binaryName),
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}

	return "./" + binaryName
}

// MockEnergyEvaluator provides a deterministic in-memory evaluator for algorithmic unit tests.
type MockEnergyEvaluator struct {
	EvaluateFunc func(ctx context.Context, sourceContent []byte, relFilePath string, funcName string) (*EnergyMeasurement, error)
	FixedEnergy  float64
	Energies     []float64
	callIndex    int
}

// EvaluateEnergy evaluates energy using the configured mock function or sequence.
func (m *MockEnergyEvaluator) EvaluateEnergy(ctx context.Context, sourceContent []byte, relFilePath string, funcName string) (*EnergyMeasurement, error) {
	if m.EvaluateFunc != nil {
		return m.EvaluateFunc(ctx, sourceContent, relFilePath, funcName)
	}

	energy := m.FixedEnergy
	if len(m.Energies) > 0 {
		if m.callIndex < len(m.Energies) {
			energy = m.Energies[m.callIndex]
			m.callIndex++
		} else {
			energy = m.Energies[len(m.Energies)-1]
		}
	}

	if math.IsNaN(energy) || math.IsInf(energy, 0) || energy <= 0.0 {
		return nil, ErrInvalidEnergy
	}

	h := sha256.Sum256(sourceContent)

	return &EnergyMeasurement{
		Energy: energy,
		Metrics: FunctionMetrics{
			CyclomaticComplexity: int(energy),
			MaxNestingDepth:      2,
			ASTNodeCount:         50,
			ConcurrencyPrims:     0,
			Energy:               energy,
		},
		FunctionName:  funcName,
		FilePath:      relFilePath,
		EngineVersion: "2.0.0-mock",
		SourceHash:    hex.EncodeToString(h[:]),
	}, nil
}
