// =============================================================================
// go-backend/agent/agent.go
//
// Agentic Thermodynamic Audit & Rectification Engine
//
// This agent consumes the Rust AST engine's JSON output and executes three
// coordinated phases:
//   Phase 1: Volatility & Entropy Profiling
//   Phase 2: Stress Failure Prediction
//   Phase 3: Thermodynamic Equilibrium & Rectification Plan
//
// The agent produces both a JSON artifact and a formatted Markdown report.
// =============================================================================

package agent

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// =============================================================================
// § 1  Data model — mirrors the Rust engine's v2 JSON output schema
// =============================================================================

// FunctionMetrics holds the per-function AST metrics from the Rust engine.
type FunctionMetrics struct {
	CyclomaticComplexity usize_t `json:"cyclomatic_complexity"`
	MaxNestingDepth      usize_t `json:"max_nesting_depth"`
	ASTNodeCount         usize_t `json:"ast_node_count"`
	ConcurrencyPrims     usize_t `json:"concurrency_primitives"`
	Energy               float64 `json:"energy"`
}

// Hotspot represents a single entropy signal within a source file.
type Hotspot struct {
	FunctionName      string          `json:"function_name"`
	LineNumber        int             `json:"line_number"`
	ByteRange         [2]int          `json:"byte_range"`
	SourceSnippet     string          `json:"source_snippet"`
	EntropyScore      float64         `json:"entropy_score"`
	VulnerabilityType string          `json:"vulnerability_type"`
	Description       string          `json:"description"`
	Metrics           FunctionMetrics `json:"metrics"`
}

// FileReport holds the per-file analysis summary.
type FileReport struct {
	FilePath           string    `json:"file_path"`
	Language           string    `json:"language"`
	LinesScanned       int       `json:"lines_scanned"`
	TotalEntropy       float64   `json:"total_entropy"`
	MeanHotspotEntropy float64   `json:"mean_hotspot_entropy"`
	Hotspots           []Hotspot `json:"hotspots"`
}

// ThermodynamicReport is the root document produced by the Rust engine.
type ThermodynamicReport struct {
	EngineVersion    string       `json:"engine_version"`
	GeneratedAt      string       `json:"generated_at"`
	ScannedDirectory string       `json:"scanned_directory"`
	FilesAnalyzed    int          `json:"files_analyzed"`
	TotalHotspots    int          `json:"total_hotspots"`
	GlobalEntropy    float64      `json:"global_entropy"`
	FileReports      []FileReport `json:"file_reports"`
}

// usize_t is an alias for int to match Rust's usize JSON serialization.
type usize_t = int

// =============================================================================
// § 2  Audit report data structures
// =============================================================================

// SystemTemperature classifies the global codebase state.
type SystemTemperature string

const (
	TempCool     SystemTemperature = "Cool"
	TempStable   SystemTemperature = "Stable"
	TempVolatile SystemTemperature = "Volatile"
	TempCritical SystemTemperature = "Critical"
)

// VolatilityEntry represents a ranked hotspot in the volatility table.
type VolatilityEntry struct {
	Rank                int     `json:"rank"`
	FunctionName        string  `json:"function_name"`
	FilePath            string  `json:"file_path"`
	LineNumber          int     `json:"line_number"`
	Energy              float64 `json:"energy"`
	PrimaryEntropyDriver string `json:"primary_entropy_driver"`
	CyclomaticComplexity int    `json:"cyclomatic_complexity"`
	MaxNestingDepth      int    `json:"max_nesting_depth"`
	ASTNodeCount         int    `json:"ast_node_count"`
	ConcurrencyPrims     int    `json:"concurrency_primitives"`
}

// StressFailure represents a predicted failure mode for a hotspot.
type StressFailure struct {
	Component        string `json:"component"`
	FilePath         string `json:"file_path"`
	FailureMode      string `json:"failure_mode"`
	TriggerCondition string `json:"trigger_condition"`
	Mechanism        string `json:"mechanism"`
	Severity         string `json:"severity"` // Critical / High / Medium
}

// RectificationDirective describes a specific refactoring action.
type RectificationDirective struct {
	Strategy       string  `json:"strategy"`
	TargetFunction string  `json:"target_function"`
	FilePath       string  `json:"file_path"`
	OriginalEnergy float64 `json:"original_energy"`
	ProjectedEnergy float64 `json:"projected_energy"`
	DeltaEnergy    float64 `json:"delta_energy"`
	CodePatch      string  `json:"code_patch"`
}

// AuditReport is the complete thermodynamic audit output.
type AuditReport struct {
	// Header metadata
	GeneratedAt       string            `json:"generated_at"`
	EngineVersion     string            `json:"engine_version"`
	ScannedDirectory  string            `json:"scanned_directory"`

	// Phase 1: Entropy Profiling
	SystemTemperature SystemTemperature `json:"system_temperature"`
	AverageFuncEnergy float64           `json:"average_function_energy"`
	TotalHotspots     int               `json:"total_hotspots"`
	VolatilityRanking []VolatilityEntry `json:"volatility_ranking"`

	// Phase 2: Stress Failure Prediction
	StressFailures    []StressFailure   `json:"stress_failures"`

	// Phase 3: Rectification Plan
	Directives        []RectificationDirective `json:"rectification_directives"`

	// System Guardrails
	Invariants        []string `json:"stability_invariants"`
}

// =============================================================================
// § 3  Agent core — the three-phase analysis engine
// =============================================================================

// RunAudit executes the full three-phase thermodynamic audit against a parsed
// engine report. Returns the structured audit result.
func RunAudit(report *ThermodynamicReport, energyThreshold float64) *AuditReport {
	audit := &AuditReport{
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		EngineVersion:    report.EngineVersion,
		ScannedDirectory: report.ScannedDirectory,
	}

	// ── Phase 1: Volatility & Entropy Profiling ──────────────────────────
	phase1(report, audit, energyThreshold)

	// ── Phase 2: Stress Failure Prediction ───────────────────────────────
	phase2(report, audit)

	// ── Phase 3: Rectification Plan ──────────────────────────────────────
	phase3(report, audit)

	// ── System Invariants ────────────────────────────────────────────────
	audit.Invariants = computeInvariants(audit)

	return audit
}

// =============================================================================
// § 3a  Phase 1: Volatility & Entropy Profiling
// =============================================================================

func phase1(report *ThermodynamicReport, audit *AuditReport, threshold float64) {
	// Collect all function hotspots across files
	type funcEntry struct {
		hotspot  Hotspot
		filePath string
	}

	var allFuncs []funcEntry
	for _, fr := range report.FileReports {
		for _, h := range fr.Hotspots {
			allFuncs = append(allFuncs, funcEntry{hotspot: h, filePath: fr.FilePath})
		}
	}

	// Compute average function energy
	if len(allFuncs) > 0 {
		totalE := 0.0
		for _, f := range allFuncs {
			totalE += f.hotspot.EntropyScore
		}
		audit.AverageFuncEnergy = math.Round(totalE/float64(len(allFuncs))*100) / 100
	}

	// Sort by energy descending
	sort.Slice(allFuncs, func(i, j int) bool {
		return allFuncs[i].hotspot.EntropyScore > allFuncs[j].hotspot.EntropyScore
	})

	// Identify top 10% most volatile functions ("Hot Entropy Nodes")
	hotCount := int(math.Ceil(float64(len(allFuncs)) * 0.10))
	if hotCount < 1 && len(allFuncs) > 0 {
		hotCount = 1
	}
	// Also include any function exceeding the threshold
	for i, f := range allFuncs {
		if i >= hotCount && f.hotspot.EntropyScore < threshold {
			break
		}
		audit.VolatilityRanking = append(audit.VolatilityRanking, VolatilityEntry{
			Rank:                 i + 1,
			FunctionName:         f.hotspot.FunctionName,
			FilePath:             f.filePath,
			LineNumber:           f.hotspot.LineNumber,
			Energy:               f.hotspot.EntropyScore,
			PrimaryEntropyDriver: f.hotspot.VulnerabilityType,
			CyclomaticComplexity: f.hotspot.Metrics.CyclomaticComplexity,
			MaxNestingDepth:      f.hotspot.Metrics.MaxNestingDepth,
			ASTNodeCount:         f.hotspot.Metrics.ASTNodeCount,
			ConcurrencyPrims:     f.hotspot.Metrics.ConcurrencyPrims,
		})
	}

	audit.TotalHotspots = len(audit.VolatilityRanking)

	// Classify system temperature
	audit.SystemTemperature = classifyTemperature(audit.AverageFuncEnergy, audit.TotalHotspots, len(allFuncs))
}

func classifyTemperature(avgEnergy float64, hotspotCount, totalFuncs int) SystemTemperature {
	hotRatio := 0.0
	if totalFuncs > 0 {
		hotRatio = float64(hotspotCount) / float64(totalFuncs)
	}

	if avgEnergy > 12.0 || hotRatio > 0.4 {
		return TempCritical
	}
	if avgEnergy > 8.0 || hotRatio > 0.25 {
		return TempVolatile
	}
	if avgEnergy > 4.0 || hotRatio > 0.10 {
		return TempStable
	}
	return TempCool
}

// =============================================================================
// § 3b  Phase 2: Stress Failure Prediction
// =============================================================================

func phase2(report *ThermodynamicReport, audit *AuditReport) {
	for _, entry := range audit.VolatilityRanking {
		failures := predictFailures(entry, report)
		audit.StressFailures = append(audit.StressFailures, failures...)
	}
}

func predictFailures(entry VolatilityEntry, report *ThermodynamicReport) []StressFailure {
	var failures []StressFailure

	driver := entry.PrimaryEntropyDriver
	m := entry.CyclomaticComplexity
	d := entry.MaxNestingDepth
	s := entry.ASTNodeCount
	c := entry.ConcurrencyPrims

	// ── Concurrency-driven failures ──────────────────────────────────────
	if c > 0 && d >= 2 {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Deadlock / Thread Starvation",
			TriggerCondition: fmt.Sprintf(
				"Concurrent load > %d RPS with %d synchronization primitives at nesting depth %d",
				200*c, c, d,
			),
			Mechanism: fmt.Sprintf(
				"Function '%s' acquires %d concurrency primitives within %d-deep control flow. "+
					"Under sustained load, lock acquisition ordering is non-deterministic, "+
					"creating priority inversion and eventual thread starvation. "+
					"AST shows %d nodes in scope, indicating long critical sections.",
				entry.FunctionName, c, d, s,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	// ── Nesting-driven failures ──────────────────────────────────────────
	if d >= 3 {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Resource Exhaustion / Cascading Timeout",
			TriggerCondition: fmt.Sprintf(
				"Input collection size > %d items with nesting depth %d (O(n^%d) path)",
				int(math.Pow(10, float64(5-d))), d, d,
			),
			Mechanism: fmt.Sprintf(
				"Function '%s' has %d-level deep nesting with cyclomatic complexity %d. "+
					"Nested iteration over dynamic collections creates O(n^%d) runtime behavior. "+
					"Under large inputs, CPU time grows polynomially, triggering upstream timeout cascades.",
				entry.FunctionName, d, m, d,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	// ── Blocking I/O driven failures ─────────────────────────────────────
	if driver == "BlockingIO" {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Latency Spike / Connection Pool Exhaustion",
			TriggerCondition: "Sustained load with upstream service degradation (P99 > 2s)",
			Mechanism: fmt.Sprintf(
				"Function '%s' performs blocking I/O within a loop at nesting depth %d. "+
					"When downstream services slow down, each iteration blocks the calling "+
					"goroutine/thread. Without timeout guards, connection pools saturate and "+
					"back-pressure propagates to all callers. AST volume: %d nodes.",
				entry.FunctionName, d, s,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	// ── Recursive call failures ──────────────────────────────────────────
	if driver == "RecursiveCall" {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Stack Overflow / Memory Exhaustion",
			TriggerCondition: "Deeply nested input structure exceeding ~10,000 recursion levels",
			Mechanism: fmt.Sprintf(
				"Function '%s' makes recursive calls without visible memoization or depth guards. "+
					"Each recursive frame allocates stack space. With deeply nested input, "+
					"the call stack grows unboundedly until runtime panic (Go) or segfault (Python).",
				entry.FunctionName,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	// ── Allocation-driven failures ───────────────────────────────────────
	if driver == "HotAllocation" {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Memory Bloat / GC Pressure",
			TriggerCondition: fmt.Sprintf("Input batch size > %d items with %d AST nodes in scope", 1000*m, s),
			Mechanism: fmt.Sprintf(
				"Function '%s' allocates heap objects inside a loop body. "+
					"With large input batches, each iteration creates new allocations "+
					"that overwhelm the garbage collector, causing GC pauses and "+
					"latency spikes. Consider pre-allocating or pooling buffers.",
				entry.FunctionName,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	// ── High cyclomatic complexity failures ──────────────────────────────
	if m >= 10 && driver == "CognitiveBranch" {
		failures = append(failures, StressFailure{
			Component:   fmt.Sprintf("%s (L%d)", entry.FunctionName, entry.LineNumber),
			FilePath:    entry.FilePath,
			FailureMode: "Logic Error / Untested Branch Defect",
			TriggerCondition: fmt.Sprintf("Edge-case input exercising >%d distinct code paths", m),
			Mechanism: fmt.Sprintf(
				"Function '%s' has cyclomatic complexity %d, requiring at minimum %d test cases "+
					"for full branch coverage. Under production input diversity, untested branches "+
					"are likely to surface latent defects. Risk amplified by %d total AST nodes.",
				entry.FunctionName, m, m, s,
			),
			Severity: classifySeverity(entry.Energy),
		})
	}

	return failures
}

func classifySeverity(energy float64) string {
	if energy > 12.0 {
		return "Critical"
	}
	if energy > 8.0 {
		return "High"
	}
	return "Medium"
}

// =============================================================================
// § 3c  Phase 3: Thermodynamic Rectification Plan
// =============================================================================

func phase3(report *ThermodynamicReport, audit *AuditReport) {
	// Generate rectification directives for the top 3 hotspots
	limit := 3
	if len(audit.VolatilityRanking) < limit {
		limit = len(audit.VolatilityRanking)
	}

	for i := 0; i < limit; i++ {
		entry := audit.VolatilityRanking[i]
		directive := generateDirective(entry, report)
		audit.Directives = append(audit.Directives, directive)
	}
}

func generateDirective(entry VolatilityEntry, report *ThermodynamicReport) RectificationDirective {
	driver := entry.PrimaryEntropyDriver
	d := RectificationDirective{
		TargetFunction: entry.FunctionName,
		FilePath:       entry.FilePath,
		OriginalEnergy: entry.Energy,
	}

	switch driver {
	case "DeepNesting":
		d.Strategy = "Flatten nested control flow with early returns and guard clauses"
		// Project: reducing nesting by 2 levels reduces D_max contribution
		newD := entry.MaxNestingDepth - 2
		if newD < 0 {
			newD = 0
		}
		d.ProjectedEnergy = computeProjectedEnergy(
			entry.CyclomaticComplexity, newD,
			entry.ASTNodeCount, entry.ConcurrencyPrims,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: %d-level nested blocks in %s\n"+
				"// AFTER: Extract inner logic into helper functions,\n"+
				"//        convert nested if-chains to early-return guard clauses.\n"+
				"//\n"+
				"// func %s(...) {\n"+
				"//     if !precondition { return errInvalid }\n"+
				"//     if !secondCheck  { return errSkipped }\n"+
				"//     // core logic at depth 1 instead of depth %d\n"+
				"// }",
			entry.MaxNestingDepth, entry.FunctionName,
			entry.FunctionName, entry.MaxNestingDepth,
		)

	case "RecursiveCall":
		d.Strategy = "Replace unbounded recursion with iterative stack + memoization"
		d.ProjectedEnergy = computeProjectedEnergy(
			entry.CyclomaticComplexity+1, 1,
			entry.ASTNodeCount+10, entry.ConcurrencyPrims,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: Recursive %s without depth guard\n"+
				"// AFTER: Iterative approach with explicit stack\n"+
				"//\n"+
				"// func %s(input) result {\n"+
				"//     stack := []frame{{input, 0}}\n"+
				"//     memo  := map[key]result{}\n"+
				"//     for len(stack) > 0 {\n"+
				"//         frame := stack[len(stack)-1]\n"+
				"//         stack = stack[:len(stack)-1]\n"+
				"//         if cached, ok := memo[frame.key]; ok { continue }\n"+
				"//         // process frame, push sub-problems\n"+
				"//     }\n"+
				"// }",
			entry.FunctionName, entry.FunctionName,
		)

	case "SyncContention":
		d.Strategy = "Replace mutex with read-write lock or atomic operations; minimize critical section scope"
		newC := entry.ConcurrencyPrims / 2
		d.ProjectedEnergy = computeProjectedEnergy(
			entry.CyclomaticComplexity, entry.MaxNestingDepth,
			entry.ASTNodeCount, newC,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: sync.Mutex held across loop body in %s\n"+
				"// AFTER: Minimize critical section, use sync.RWMutex for read-heavy paths\n"+
				"//\n"+
				"// var mu sync.RWMutex  // was: sync.Mutex\n"+
				"//\n"+
				"// func %s(...) {\n"+
				"//     mu.RLock()           // read path: no write contention\n"+
				"//     cached := lookupCache(key)\n"+
				"//     mu.RUnlock()\n"+
				"//     if cached != nil { return cached }\n"+
				"//\n"+
				"//     result := computeExpensive(...)  // OUTSIDE lock\n"+
				"//\n"+
				"//     mu.Lock()            // write path: minimal scope\n"+
				"//     cache[key] = result\n"+
				"//     mu.Unlock()\n"+
				"// }",
			entry.FunctionName, entry.FunctionName,
		)

	case "BlockingIO":
		d.Strategy = "Inject context deadlines and bounded worker pools for I/O operations"
		newD := entry.MaxNestingDepth
		if newD > 1 {
			newD -= 1
		}
		d.ProjectedEnergy = computeProjectedEnergy(
			entry.CyclomaticComplexity+1, newD,
			entry.ASTNodeCount+5, entry.ConcurrencyPrims,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: Blocking I/O in loop body of %s\n"+
				"// AFTER: Context-guarded I/O with bounded concurrency\n"+
				"//\n"+
				"// func %s(ctx context.Context, ...) error {\n"+
				"//     sem := make(chan struct{}, maxWorkers)  // bounded pool\n"+
				"//     g, ctx := errgroup.WithContext(ctx)\n"+
				"//     for _, item := range items {\n"+
				"//         sem <- struct{}{}  // acquire slot\n"+
				"//         g.Go(func() error {\n"+
				"//             defer func() { <-sem }()  // release slot\n"+
				"//             reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)\n"+
				"//             defer cancel()\n"+
				"//             return doIO(reqCtx, item)\n"+
				"//         })\n"+
				"//     }\n"+
				"//     return g.Wait()\n"+
				"// }",
			entry.FunctionName, entry.FunctionName,
		)

	case "HotAllocation":
		d.Strategy = "Pre-allocate buffers outside loops; use sync.Pool for reusable objects"
		d.ProjectedEnergy = computeProjectedEnergy(
			entry.CyclomaticComplexity, entry.MaxNestingDepth,
			entry.ASTNodeCount-5, entry.ConcurrencyPrims,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: Allocation inside loop in %s\n"+
				"// AFTER: Pre-allocate with known capacity, reuse buffers\n"+
				"//\n"+
				"// var bufPool = sync.Pool{New: func() any { return make([]byte, 0, 4096) }}\n"+
				"//\n"+
				"// func %s(...) {\n"+
				"//     buf := bufPool.Get().([]byte)[:0]  // reuse from pool\n"+
				"//     defer bufPool.Put(buf)\n"+
				"//     results := make([]Result, 0, len(items))  // pre-allocate\n"+
				"//     for _, item := range items {\n"+
				"//         // use buf, no per-iteration allocation\n"+
				"//     }\n"+
				"// }",
			entry.FunctionName, entry.FunctionName,
		)

	default:
		d.Strategy = "Decompose high-complexity function into smaller, testable units"
		halfM := entry.CyclomaticComplexity / 2
		if halfM < 1 {
			halfM = 1
		}
		d.ProjectedEnergy = computeProjectedEnergy(
			halfM, entry.MaxNestingDepth,
			entry.ASTNodeCount/2, entry.ConcurrencyPrims,
		)
		d.CodePatch = fmt.Sprintf(
			"// BEFORE: Monolithic %s with complexity %d\n"+
				"// AFTER: Split into focused sub-functions\n"+
				"//\n"+
				"// func %s(...) { validate(...); process(...); finalize(...) }\n"+
				"// func validate(...) error { /* guard clauses */ }\n"+
				"// func process(...) Result { /* core logic */ }\n"+
				"// func finalize(...) error { /* cleanup */ }",
			entry.FunctionName, entry.CyclomaticComplexity,
			entry.FunctionName,
		)
	}

	d.DeltaEnergy = math.Round((d.ProjectedEnergy-d.OriginalEnergy)*100) / 100
	return d
}

// computeProjectedEnergy calculates the energy after a proposed refactoring.
func computeProjectedEnergy(m, d, s, c int) float64 {
	if s < 1 {
		s = 1
	}
	e := 0.45*float64(m) + 0.25*float64(d) + 0.15*math.Log2(float64(s)) + 0.15*float64(c)
	return math.Round(e*100) / 100
}

// =============================================================================
// § 4  System Invariants
// =============================================================================

func computeInvariants(audit *AuditReport) []string {
	invariants := []string{
		"CI Gate: Block merges where any function has E > 15.0 without an explicit waiver.",
		"Maximum nesting depth D_max ≤ 4 for all new/modified functions.",
		"Cyclomatic complexity M ≤ 15 per function; decompose above threshold.",
		"All blocking I/O operations must be wrapped with context.WithTimeout or equivalent.",
		"Concurrency primitives (mutexes, channels) must not be acquired inside loops without bounded worker pools.",
		"Recursive functions must include explicit depth limits and memoization where applicable.",
		"Pre-allocate collections with known capacity instead of growing inside hot loops.",
	}

	if audit.SystemTemperature == TempCritical {
		invariants = append(invariants,
			"CRITICAL: System temperature exceeds safe threshold. Prioritize top-3 rectification directives before next release.",
		)
	}

	return invariants
}

// =============================================================================
// § 5  Report serialization — JSON + Markdown
// =============================================================================

// WriteAuditJSON writes the audit report as a JSON file.
func WriteAuditJSON(audit *AuditReport, outputPath string) error {
	data, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit JSON: %w", err)
	}
	return os.WriteFile(outputPath, data, 0o640)
}

// WriteAuditMarkdown writes the audit report as a formatted Markdown file.
func WriteAuditMarkdown(audit *AuditReport, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("# Repository Thermodynamic Equilibrium & Stress Audit Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", audit.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Engine Version:** %s  \n", audit.EngineVersion))
	sb.WriteString(fmt.Sprintf("**Scanned Directory:** `%s`  \n\n", audit.ScannedDirectory))
	sb.WriteString("---\n\n")

	// ── Section 1: Executive Summary ─────────────────────────────────────
	sb.WriteString("## 1. Executive Entropy Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Global System Temperature:** %s\n", string(audit.SystemTemperature)))
	sb.WriteString(fmt.Sprintf("- **Average Function Energy:** %.2f\n", audit.AverageFuncEnergy))
	sb.WriteString(fmt.Sprintf("- **Total Hotspots Identified:** %d\n\n", audit.TotalHotspots))

	// ── Section 2: Volatility Ranking ────────────────────────────────────
	sb.WriteString("## 2. Volatility Ranking (High-Entropy Hotspots)\n\n")
	sb.WriteString("| Rank | Function / Subsystem | File & Lines | Energy (E) | M | D | S | C | Primary Entropy Driver |\n")
	sb.WriteString("|------|----------------------|--------------|------------|---|---|---|---|------------------------|\n")
	for _, v := range audit.VolatilityRanking {
		shortPath := filepath.Base(v.FilePath)
		sb.WriteString(fmt.Sprintf(
			"| %d | `%s` | `%s:L%d` | %.2f | %d | %d | %d | %d | %s |\n",
			v.Rank, v.FunctionName, shortPath, v.LineNumber,
			v.Energy, v.CyclomaticComplexity, v.MaxNestingDepth,
			v.ASTNodeCount, v.ConcurrencyPrims, v.PrimaryEntropyDriver,
		))
	}
	sb.WriteString("\n")

	// ── Section 3: Stress Failure Projections ────────────────────────────
	sb.WriteString("## 3. Stress Testing Failure Projections\n\n")
	for i, sf := range audit.StressFailures {
		sb.WriteString(fmt.Sprintf("### 3.%d — %s\n\n", i+1, sf.Component))
		sb.WriteString(fmt.Sprintf("- **File:** `%s`\n", sf.FilePath))
		sb.WriteString(fmt.Sprintf("- **Failure Mode:** %s\n", sf.FailureMode))
		sb.WriteString(fmt.Sprintf("- **Severity:** %s\n", sf.Severity))
		sb.WriteString(fmt.Sprintf("- **Trigger Condition:** %s\n", sf.TriggerCondition))
		sb.WriteString(fmt.Sprintf("- **Mechanism:** %s\n\n", sf.Mechanism))
	}

	// ── Section 4: Rectification Plan ────────────────────────────────────
	sb.WriteString("## 4. Thermodynamic Rectification & Neutralization Plan\n\n")
	for i, rd := range audit.Directives {
		sb.WriteString(fmt.Sprintf("### 4.%d — `%s`\n\n", i+1, rd.TargetFunction))
		sb.WriteString(fmt.Sprintf("- **File:** `%s`\n", rd.FilePath))
		sb.WriteString(fmt.Sprintf("- **Refactoring Strategy:** %s\n", rd.Strategy))
		sb.WriteString(fmt.Sprintf("- **Original Energy:** %.2f → **Projected Energy:** %.2f (ΔE = %.2f)\n\n",
			rd.OriginalEnergy, rd.ProjectedEnergy, rd.DeltaEnergy))
		sb.WriteString("```diff\n")
		sb.WriteString(rd.CodePatch)
		sb.WriteString("\n```\n\n")
	}

	// ── Section 5: Invariants ────────────────────────────────────────────
	sb.WriteString("## 5. System Invariants & Long-Term Stability Guardrails\n\n")
	for _, inv := range audit.Invariants {
		sb.WriteString(fmt.Sprintf("- %s\n", inv))
	}
	sb.WriteString("\n---\n\n")
	sb.WriteString("*Report generated by the Thermodynamic Audit & Rectification Engine.*\n")

	return os.WriteFile(outputPath, []byte(sb.String()), 0o640)
}

// =============================================================================
// § 6  Convenience: Load report from disk and run full audit
// =============================================================================

// LoadAndAudit reads an engine report JSON file, runs the full audit, and
// writes both JSON and Markdown outputs to the specified directory.
func LoadAndAudit(reportPath, outputDir string, energyThreshold float64) (*AuditReport, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("read engine report: %w", err)
	}

	var report ThermodynamicReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse engine report: %w", err)
	}

	audit := RunAudit(&report, energyThreshold)

	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	jsonPath := filepath.Join(outputDir, "thermodynamic_audit.json")
	if err := WriteAuditJSON(audit, jsonPath); err != nil {
		return nil, fmt.Errorf("write audit JSON: %w", err)
	}

	mdPath := filepath.Join(outputDir, "thermodynamic_audit.md")
	if err := WriteAuditMarkdown(audit, mdPath); err != nil {
		return nil, fmt.Errorf("write audit markdown: %w", err)
	}

	return audit, nil
}
