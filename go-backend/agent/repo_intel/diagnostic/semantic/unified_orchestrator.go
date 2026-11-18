package semantic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

// UnifiedDiagnosticReport encapsulates the complete end-to-end diagnostic outcome.
type UnifiedDiagnosticReport struct {
	Timestamp            time.Time              `json:"timestamp"`
	RepoRoot             string                 `json:"repo_root"`
	FilesScanned         int                    `json:"files_scanned"`
	PackagesScanned      int                    `json:"packages_scanned"`
	CallGraphSummary     map[string]interface{} `json:"call_graph_summary"`
	ReconciliationLedger *ReconciliationLedger  `json:"reconciliation_ledger"`
	DurationMs           int64                  `json:"duration_ms"`
}

// ToDiagnosticReport converts the UnifiedDiagnosticReport into a standard diagnostic.DiagnosticReport
// that can be directly ingested and processed by the RemediationPlanner.
func (u *UnifiedDiagnosticReport) ToDiagnosticReport() *diagnostic.DiagnosticReport {
	if u == nil || u.ReconciliationLedger == nil {
		return &diagnostic.DiagnosticReport{
			ReportSchemaVersion: "1.0",
			ReportID:            fmt.Sprintf("DIAG-SEM-EMPTY-%d", time.Now().UnixNano()),
		}
	}
	return BuildDiagnosticReportFromContracts(u.RepoRoot, "commit-heldout", u.ReconciliationLedger.Contracts)
}

// UnifiedOrchestrator coordinates all semantic diagnostic engines with shared AST/CFG caching and isolated execution.
type UnifiedOrchestrator struct {
	repoRoot      string
	semAnalyzer   *SemanticAnalyzer
	resolver      *CallSiteResolver
	summaryEngine *FunctionSummaryEngine
	sqlAnalyzer   *SQLAnalyzer
	cmdAnalyzer   *CommandAnalyzer
	resAnalyzer   *ResourceAnalyzer
	errAnalyzer   *ErrorAnalyzer
	reconciler    *FindingReconciler
}

// NewUnifiedOrchestrator creates an integrated orchestration instance for a repository.
func NewUnifiedOrchestrator(repoRoot string) *UnifiedOrchestrator {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}

	sumEngine := NewFunctionSummaryEngine(repoRoot)
	sqlAn := NewSQLAnalyzer(repoRoot)
	sqlAn.SetSummaryEngine(sumEngine)

	cmdAn := NewCommandAnalyzer(repoRoot)
	cmdAn.SetSummaryEngine(sumEngine)

	resAn := NewResourceAnalyzer(repoRoot)
	resAn.SetSummaryEngine(sumEngine)

	return &UnifiedOrchestrator{
		repoRoot:      repoRoot,
		semAnalyzer:   NewSemanticAnalyzer(repoRoot),
		resolver:      NewCallSiteResolver(repoRoot),
		summaryEngine: sumEngine,
		sqlAnalyzer:   sqlAn,
		cmdAnalyzer:   cmdAn,
		resAnalyzer:   resAn,
		errAnalyzer:   NewErrorAnalyzer(repoRoot),
		reconciler:    NewFindingReconciler(repoRoot),
	}
}

// RunCompleteDiagnostic executes the unified multi-stage semantic analysis across all target files.
func (u *UnifiedOrchestrator) RunCompleteDiagnostic(files []string, repoCommit string) (*UnifiedDiagnosticReport, error) {
	t0 := time.Now()

	// 1. Group files by directory to construct package models
	dirMap := make(map[string][]string)
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			dir := filepath.Dir(f)
			dirMap[dir] = append(dirMap[dir], f)
		}
	}

	pkgModels := make(map[string]*PackageSemanticModel)
	for dir := range dirMap {
		pkgModel, err := u.semAnalyzer.AnalyzePackage(dir)
		if err == nil && pkgModel != nil {
			pkgModels[dir] = pkgModel
			u.resolver.RegisterPackage(pkgModel)
		}
	}

	// 2. Build Bounded Call Graphs for packages
	callGraphs := make(map[string]*BoundedCallGraph)
	totalCallEdges := 0
	for dir, pkgModel := range pkgModels {
		cg := u.resolver.BuildCallGraph(pkgModel)
		if cg != nil {
			callGraphs[dir] = cg
			totalCallEdges += len(cg.Edges)
		}
	}

	// 2.5 Build and propagate interprocedural function summaries bottom-up
	u.summaryEngine.BuildSummariesBottomUp(pkgModels, callGraphs)

	// 3. Deterministic Analyzer Execution with Failure Isolation
	var allSQLEvs []SQLEvidence
	var allCmdEvs []CommandEvidence
	var allResEvs []ResourceEvidence
	var allErrEvs []ErrorEvidence

	for _, f := range files {
		if !strings.HasSuffix(f, ".go") {
			continue
		}

		fileModel, _, err := u.semAnalyzer.AnalyzeFile(f)
		if err != nil || fileModel == nil {
			continue
		}

		// Register local functions for error signature resolution
		for _, cfg := range fileModel.CFGs {
			u.errAnalyzer.RegisterLocalFunction(cfg.FuncDecl)
		}

		for fnName, cfg := range fileModel.CFGs {
			if cfg.FuncDecl == nil || cfg.FuncDecl.Body == nil {
				continue
			}
			df := fileModel.Dataflow[fnName]

			// A. SQL Analysis (Isolated)
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Failure isolation: log and continue
					}
				}()
				evs, err := u.sqlAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, f)
				if err == nil {
					allSQLEvs = append(allSQLEvs, evs...)
				}
			}()

			// B. Command Execution Analysis (Isolated)
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Failure isolation
					}
				}()
				evs, err := u.cmdAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, f)
				if err == nil {
					allCmdEvs = append(allCmdEvs, evs...)
				}
			}()

			// C. Resource Lifecycle Analysis (Isolated)
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Failure isolation
					}
				}()
				evs, err := u.resAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, f)
				if err == nil {
					allResEvs = append(allResEvs, evs...)
				}
			}()

			// D. Error Handling Analysis (Isolated)
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Failure isolation
					}
				}()
				evs, err := u.errAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, f, nil)
				if err == nil {
					allErrEvs = append(allErrEvs, evs...)
				}
			}()
		}
	}

	// 4. Canonical Finding Reconciliation
	ledger := u.reconciler.ReconcileFindings(
		nil, // No raw candidates in isolated semantic run
		allSQLEvs,
		allCmdEvs,
		allResEvs,
		allErrEvs,
		repoCommit,
	)

	duration := time.Since(t0).Milliseconds()
	report := &UnifiedDiagnosticReport{
		Timestamp:       time.Now(),
		RepoRoot:        u.repoRoot,
		FilesScanned:    len(files),
		PackagesScanned: len(pkgModels),
		CallGraphSummary: map[string]interface{}{
			"packages_graphed": len(callGraphs),
			"total_call_edges": totalCallEdges,
		},
		ReconciliationLedger: ledger,
		DurationMs:           duration,
	}

	return report, nil
}

// AnalyzeRepository discovers all Go source files under a directory and runs complete diagnostics.
func (u *UnifiedOrchestrator) AnalyzeRepository(targetDir string) (*UnifiedDiagnosticReport, error) {
	var files []string
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			absPath, err := filepath.Abs(path)
			if err == nil {
				rel, err := filepath.Rel(u.repoRoot, absPath)
				if err == nil && !strings.HasPrefix(rel, "..") {
					files = append(files, rel)
					return nil
				}
			}
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.RunCompleteDiagnostic(files, "commit-heldout")
}
