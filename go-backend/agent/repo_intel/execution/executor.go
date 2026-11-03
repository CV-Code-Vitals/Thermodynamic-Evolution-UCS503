package execution

import (
	"fmt"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// RemediationExecutor orchestrates the end-to-end execution lifecycle.
type RemediationExecutor struct {
	Config       ExecutionConfig
	WorkspaceMgr *WorkspaceManager
	StrategyReg  *StrategyRegistry
	Validator    *IndependentValidator
	RollbackEng  *RollbackEngine
	Reporter     *Reporter
}

// NewRemediationExecutor creates an initialized executor instance.
func NewRemediationExecutor(cfg ExecutionConfig) (*RemediationExecutor, error) {
	wm, err := NewWorkspaceManager(cfg.RepoRoot, cfg.WorkspaceBaseDir, cfg.KeepWorkspaces)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize workspace manager: %w", err)
	}

	return &RemediationExecutor{
		Config:       cfg,
		WorkspaceMgr: wm,
		StrategyReg:  NewStrategyRegistry(),
		Validator:    NewIndependentValidator(cfg.TimeoutSec),
		RollbackEng:  NewRollbackEngine(cfg.RepoRoot),
		Reporter:     NewReporter("reports"),
	}, nil
}

// ExecutePlan executes a loaded remediation plan under configured isolation and constraints.
func (e *RemediationExecutor) ExecutePlan(plan *planning.RemediationPlan, planHash string) (*ExecutionReport, error) {
	start := time.Now()

	schedule, err := BuildExecutionSchedule(plan, e.Config.TargetTaskIDs, e.Config.BatchLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to build execution schedule: %w", err)
	}

	approvedSet := make(map[string]bool)
	for _, id := range e.Config.ApprovedTaskIDs {
		approvedSet[id] = true
	}

	report := &ExecutionReport{
		ReportID:       fmt.Sprintf("EXEC-%d", time.Now().Unix()),
		PlanID:         plan.PlanID,
		PlanHash:       planHash,
		GeneratedAt:    time.Now(),
		BaselineCommit: plan.BaselineCommit,
		Summary: ExecutionSummary{
			TotalScheduled: len(schedule.OrderedTasks),
		},
		TaskResults:    make([]TaskExecutionResult, 0),
		PatchManifest:  make([]UnifiedPatch, 0),
		WorkspaceAudit: make([]WorkspaceAuditRecord, 0),
	}

	for _, task := range schedule.OrderedTasks {
		res, audit, err := e.ExecuteTask(task, approvedSet)
		if err != nil {
			// Record execution error
			res = &TaskExecutionResult{
				TaskID:         task.TaskID,
				Title:          task.Title,
				Status:         StatusFailed,
				ExecutionClass: string(task.ExecutionClass),
				ErrorMessage:   err.Error(),
			}
		}

		if audit != nil {
			report.WorkspaceAudit = append(report.WorkspaceAudit, *audit)
		}

		report.TaskResults = append(report.TaskResults, *res)

		// Aggregate summaries
		switch res.Status {
		case StatusCompleted:
			report.Summary.Completed++
			if res.Patch != nil {
				report.PatchManifest = append(report.PatchManifest, *res.Patch)
			}
			if res.ReviewRequired {
				report.Summary.ReviewRequiredCount++
			}
		case StatusRejected:
			report.Summary.Rejected++
		case StatusRolledBack:
			report.Summary.RolledBack++
		case StatusFailed:
			report.Summary.Failed++
		}
	}

	report.Summary.TotalDurationMs = time.Since(start).Milliseconds()
	return report, nil
}

// ExecuteTask executes a single remediation task in an isolated workspace.
func (e *RemediationExecutor) ExecuteTask(task *planning.RemediationTask, approvedSet map[string]bool) (*TaskExecutionResult, *WorkspaceAuditRecord, error) {
	start := time.Now()

	result := &TaskExecutionResult{
		TaskID:         task.TaskID,
		Title:          task.Title,
		Status:         StatusPending,
		ExecutionClass: string(task.ExecutionClass),
		StrategyUsed:   string(task.ExecutionStrategy),
		StartTime:      start,
		TargetFiles:    task.TargetFiles,
		TaskRef:        task,
	}

	// 1. Eligibility and Human Approval Verification
	eligible, reviewReq, rejectReason := CheckTaskEligibility(task, approvedSet)
	result.ReviewRequired = reviewReq
	if approvedSet[task.TaskID] {
		result.HumanApprovalVerified = true
	}

	if !eligible {
		result.Status = StatusRejected
		result.ErrorMessage = rejectReason
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil, nil
	}

	// 2. Dry-Run simulation bypass
	if e.Config.DryRun {
		result.Status = StatusCompleted
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil, nil
	}

	// 3. Isolated Workspace Creation
	result.Status = StatusPreparingWorkspace
	wsDir, audit, err := e.WorkspaceMgr.CreateWorkspace(task.TaskID, task.TargetFiles)
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed to create isolated workspace: %v", err)
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, nil, err
	}
	result.WorkspacePath = wsDir

	// 4. Strategy Execution
	result.Status = StatusExecuting
	strat, err := e.StrategyReg.Get(task.ExecutionStrategy)
	if err != nil {
		_ = e.WorkspaceMgr.DestroyWorkspace(wsDir, audit)
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("unsupported strategy: %v", err)
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, audit, err
	}

	modifiedFiles, err := strat.Execute(wsDir, task)
	if err != nil {
		_ = e.WorkspaceMgr.DestroyWorkspace(wsDir, audit)
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("strategy execution error: %v", err)
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, audit, err
	}

	// 5. Generate Reversible Unified Patch
	allTargets := make([]string, 0, len(task.TargetFiles)+len(modifiedFiles))
	targetMap := make(map[string]bool)
	for _, tf := range task.TargetFiles {
		if !targetMap[tf] {
			targetMap[tf] = true
			allTargets = append(allTargets, tf)
		}
	}
	for _, mf := range modifiedFiles {
		if !targetMap[mf] {
			targetMap[mf] = true
			allTargets = append(allTargets, mf)
		}
	}

	patch, err := GenerateUnifiedPatch(task.TaskID, e.Config.RepoRoot, wsDir, allTargets)
	if err != nil {
		_ = e.WorkspaceMgr.DestroyWorkspace(wsDir, audit)
		result.Status = StatusFailed
		result.ErrorMessage = fmt.Sprintf("failed generating unified patch: %v", err)
		result.EndTime = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		return result, audit, err
	}
	result.Patch = patch

	// Validate Patch Scope
	if patch != nil {
		if err := ValidatePatchScope(patch, allTargets); err != nil {
			_ = e.RollbackEng.RollbackWorkspace(wsDir, modifiedFiles, "scope violation")
			_ = e.WorkspaceMgr.DestroyWorkspace(wsDir, audit)
			result.Status = StatusRolledBack
			result.RollbackOccurred = true
			result.RollbackReason = err.Error()
			result.ErrorMessage = err.Error()
			result.EndTime = time.Now()
			result.DurationMs = time.Since(start).Milliseconds()
			return result, audit, nil
		}
	}

	// 6. Independent Multi-Stage Validation
	result.Status = StatusValidating
	valReport, err := e.Validator.Validate(wsDir, task, patch)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("validation engine error: %v", err)
	}
	if valReport != nil {
		result.Validation = *valReport
	}

	// Check if validation failed
	if valReport != nil && valReport.OverallStatus == ValFailed {
		result.Status = StatusRolledBack
		result.RollbackOccurred = true
		result.RollbackReason = "Post-patch validation checks failed (build, test, or defect unresolved)"
		_ = e.RollbackEng.RollbackWorkspace(wsDir, modifiedFiles, result.RollbackReason)
	} else {
		result.Status = StatusCompleted
	}

	// 7. Workspace Cleanup
	_ = e.WorkspaceMgr.DestroyWorkspace(wsDir, audit)

	result.EndTime = time.Now()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, audit, nil
}

