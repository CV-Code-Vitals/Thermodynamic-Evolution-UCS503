package semantic

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// RemediationEligibility classifies whether a reconciled finding is safe for automated patching.
type RemediationEligibility string

const (
	EligibleAutonomous    RemediationEligibility = "AUTONOMOUS_ELIGIBLE"
	EligibleHumanApproval RemediationEligibility = "HUMAN_APPROVAL_REQUIRED"
	EligibleQuarantined   RemediationEligibility = "QUARANTINED"
	EligibleBenign        RemediationEligibility = "BENIGN"
)

// ReconciledFinding represents a canonically validated defect with full provenance and remediation gating.
type ReconciledFinding struct {
	FindingID              string                        `json:"finding_id"`
	RuleID                 string                        `json:"rule_id"`
	Category               repo_intel.Category           `json:"category"`
	Severity               repo_intel.Severity           `json:"severity"`
	Confidence             float64                       `json:"confidence"`
	FindingState           FindingState                  `json:"finding_state"`
	TargetFile             string                        `json:"target_file"`
	LineStart              int                           `json:"line_start"`
	LineEnd                int                           `json:"line_end"`
	Title                  string                        `json:"title"`
	Description            string                        `json:"description"`
	RootCause              string                        `json:"root_cause"`
	RecommendedAction      string                        `json:"recommended_action"`
	EvidenceLocations      []repo_intel.EvidenceLocation `json:"evidence_locations"`
	CounterEvidences       []string                      `json:"counter_evidences,omitempty"`
	RemediationEligibility RemediationEligibility        `json:"remediation_eligibility"`
	ApprovalRationale      string                        `json:"approval_rationale,omitempty"`
	FileSHA256             string                        `json:"file_sha256,omitempty"`
	UncertaintyClass       string                        `json:"uncertainty_class,omitempty"`
	ConcurrencyState       string                        `json:"concurrency_state,omitempty"`
	AccessPath             string                        `json:"access_path,omitempty"`
	Provenance             string                        `json:"provenance"`
	Timestamp              time.Time                     `json:"timestamp"`
}

// ReconciliationDecision captures an auditable action taken by the reconciliation layer.
type ReconciliationDecision struct {
	Action           string    `json:"action"` // "RETAIN", "MERGE", "SUPPRESS", "QUARANTINE"
	FindingID        string    `json:"finding_id"`
	RelatedFindingID string    `json:"related_finding_id,omitempty"`
	RuleID           string    `json:"rule_id"`
	Reason           string    `json:"reason"`
	Timestamp        time.Time `json:"timestamp"`
}

// ReconciliationLedger summarizes all retention, merging, and suppression choices.
type ReconciliationLedger struct {
	TotalProcessed   int                       `json:"total_processed"`
	RetainedCount    int                       `json:"retained_count"`
	MergedCount      int                       `json:"merged_count"`
	SuppressedCount  int                       `json:"suppressed_count"`
	QuarantinedCount int                       `json:"quarantined_count"`
	Decisions        []ReconciliationDecision  `json:"decisions"`
	Findings         []ReconciledFinding       `json:"findings"`
	Contracts        []SemanticFindingContract `json:"contracts,omitempty"`
}

// FindingReconciler resolves overlapping findings across regex heuristics, AST contexts, and semantic IR analyzers.
type FindingReconciler struct {
	repoRoot string
}

// NewFindingReconciler initializes a reconciler instance.
func NewFindingReconciler(repoRoot string) *FindingReconciler {
	return &FindingReconciler{repoRoot: repoRoot}
}

// ReconcileFindings performs canonical deduplication, counter-evidence suppression, and eligibility assignment.
func (r *FindingReconciler) ReconcileFindings(
	rawCandidates []repo_intel.Finding,
	sqlEvs []SQLEvidence,
	cmdEvs []CommandEvidence,
	resEvs []ResourceEvidence,
	errEvs []ErrorEvidence,
	repoCommit string,
) *ReconciliationLedger {
	ledger := &ReconciliationLedger{
		Decisions: make([]ReconciliationDecision, 0),
		Findings:  make([]ReconciledFinding, 0),
		Contracts: make([]SemanticFindingContract, 0),
	}

	seenAnchors := make(map[string]int) // "file:line:category:rule" -> index in ledger.Findings

	// Helper to add or retain a finding
	addReconciled := func(rf ReconciledFinding, ruleID, reason string) {
		ledger.TotalProcessed++
		anchorKey := fmt.Sprintf("%s:%d:%s:%s", rf.TargetFile, rf.LineStart, rf.Category, ruleID)

		if idx, exists := seenAnchors[anchorKey]; exists {
			// Exact duplicate on same anchor and rule -> MERGE
			existing := &ledger.Findings[idx]
			if rf.Confidence > existing.Confidence {
				existing.Confidence = rf.Confidence
			}
			ledger.MergedCount++
			ledger.Decisions = append(ledger.Decisions, ReconciliationDecision{
				Action:           "MERGE",
				FindingID:        rf.FindingID,
				RelatedFindingID: existing.FindingID,
				RuleID:           ruleID,
				Reason:           fmt.Sprintf("Merged duplicate evidence on anchor %s", anchorKey),
				Timestamp:        time.Now(),
			})
			return
		}

		// Retain distinct finding
		seenAnchors[anchorKey] = len(ledger.Findings)
		ledger.RetainedCount++
		ledger.Findings = append(ledger.Findings, rf)
		ledger.Decisions = append(ledger.Decisions, ReconciliationDecision{
			Action:    "RETAIN",
			FindingID: rf.FindingID,
			RuleID:    ruleID,
			Reason:    reason,
			Timestamp: time.Now(),
		})

		// Build and append corresponding versioned contract
		var fileHash string
		if r.repoRoot != "" && rf.TargetFile != "" {
			fullPath := filepath.Join(r.repoRoot, filepath.FromSlash(rf.TargetFile))
			if h, err := ComputeFileSHA256(fullPath); err == nil {
				fileHash = h
				rf.FileSHA256 = h
			}
		}

		contract := SemanticFindingContract{
			SchemaVersion:          "2.0-semantic",
			FindingID:              rf.FindingID,
			RuleID:                 ruleID,
			SourceCommit:           repoCommit,
			FileSHA256:             fileHash,
			TargetFile:             rf.TargetFile,
			LineStart:              rf.LineStart,
			LineEnd:                rf.LineEnd,
			Category:               rf.Category,
			Severity:               rf.Severity,
			Confidence:             rf.Confidence,
			AnalysisCompleteness:   CompletenessFull,
			UncertaintyClass:       UncertaintyDeterministic,
			RemediationEligibility: rf.RemediationEligibility,
			HumanApprovalRequired:  rf.RemediationEligibility != EligibleAutonomous,
			ApprovalRationale:      rf.ApprovalRationale,
			ValidationRequirements: []string{"SEMANTIC_RESCAN", "COMPILER_GO_VET", "UNIT_TEST_REGRESSION"},
			CounterEvidences:       rf.CounterEvidences,
			Title:                  rf.Title,
			Description:            rf.Description,
			RootCause:              rf.RootCause,
			RecommendedAction:      rf.RecommendedAction,
			GeneratedAt:            time.Now(),
		}
		if rf.UncertaintyClass != "" {
			contract.UncertaintyClass = UncertaintyClassification(rf.UncertaintyClass)
		}
		contract.SemanticTrace.ConcurrencyState = rf.ConcurrencyState
		contract.SemanticTrace.AccessPath = rf.AccessPath
		ledger.Contracts = append(ledger.Contracts, contract)
	}

	// 1. Process Semantic SQL Evidence
	for _, ev := range sqlEvs {
		if ev.Classification == ClassificationUntrustedSQLStructure || ev.Classification == ClassificationDynamicSQLStructure {
			rf := ReconciledFinding{
				FindingID:         fmt.Sprintf("SEM-SEC-SQL-%s-L%d", sanitizeRel(ev.FilePath), ev.Line),
				RuleID:            "SEMANTIC_SQL_INJECTION",
				Category:          repo_intel.CategorySecurity,
				Severity:          ev.Severity,
				Confidence:        ev.Confidence,
				FindingState:      ev.LifecycleState,
				TargetFile:        ev.FilePath,
				LineStart:         ev.Line,
				LineEnd:           ev.Line,
				Title:             ev.Title,
				Description:       ev.Description,
				RootCause:         ev.RootCause,
				RecommendedAction: ev.RecommendedAction,
				EvidenceLocations: []repo_intel.EvidenceLocation{
					{
						File:        ev.FilePath,
						LineStart:   ev.Line,
						LineEnd:     ev.Line,
						Description: ev.QueryExprSnippet,
					},
				},
				RemediationEligibility: EligibleAutonomous,
				Provenance:             repoCommit,
				Timestamp:              time.Now(),
			}
			addReconciled(rf, "SEMANTIC_SQL_INJECTION", "Verified semantic SQL injection sink with unescaped formatting")
		}
	}

	// 2. Process Semantic Command Evidence
	for _, ev := range cmdEvs {
		if ev.Classification == ClassificationUntrustedShellCommand || ev.Classification == ClassificationShellInterpretedCommand {
			rf := ReconciledFinding{
				FindingID:         fmt.Sprintf("SEM-SEC-CMD-%s-L%d", sanitizeRel(ev.FilePath), ev.Line),
				RuleID:            "SEMANTIC_COMMAND_INJECTION",
				Category:          repo_intel.CategorySecurity,
				Severity:          ev.Severity,
				Confidence:        ev.Confidence,
				FindingState:      ev.LifecycleState,
				TargetFile:        ev.FilePath,
				LineStart:         ev.Line,
				LineEnd:           ev.Line,
				Title:             ev.Title,
				Description:       ev.Description,
				RootCause:         ev.RootCause,
				RecommendedAction: ev.RecommendedAction,
				EvidenceLocations: []repo_intel.EvidenceLocation{
					{
						File:        ev.FilePath,
						LineStart:   ev.Line,
						LineEnd:     ev.Line,
						Description: ev.SinkCallSnippet,
					},
				},
				RemediationEligibility: EligibleAutonomous,
				Provenance:             repoCommit,
				Timestamp:              time.Now(),
			}
			addReconciled(rf, "SEMANTIC_COMMAND_INJECTION", "Verified dynamic command passed to shell interpreter")
		}
	}

	// 3. Process Semantic Resource Evidence
	for _, ev := range resEvs {
		if ev.LifecycleState == StateResourceLeakConfirmed || ev.LifecycleState == StateResourceLeakProbable {
			eligibility := EligibleAutonomous
			approvalRationale := ""
			uncertainty := UncertaintyDeterministic

			if ev.LifecycleState == StateResourceLeakProbable {
				eligibility = EligibleHumanApproval
				approvalRationale = "Probable resource leak requires human review"
				uncertainty = UncertaintyBounded
			}

			// Fail closed on concurrent ownership escape or channel transfer
			if ev.ConcurrencyState == ConcurrencyPossibleEscape || ev.ConcurrencyState == ConcurrencyUnknownOwnership {
				eligibility = EligibleHumanApproval
				approvalRationale = "Resource handle has uncertain concurrent ownership or goroutine escape; autonomous remediation prohibited"
				uncertainty = UncertaintyPossibleEscape
			}

			rf := ReconciledFinding{
				FindingID:              fmt.Sprintf("SEM-REL-RES-%s-L%d", sanitizeRel(ev.FilePath), ev.Line),
				RuleID:                 "SEMANTIC_RESOURCE_LEAK",
				Category:               repo_intel.CategoryReliability,
				Severity:               ev.Severity,
				Confidence:             ev.Confidence,
				FindingState:           ev.FindingState,
				TargetFile:             ev.FilePath,
				LineStart:              ev.Line,
				LineEnd:                ev.Line,
				Title:                  ev.Title,
				Description:            ev.Description,
				RootCause:              ev.RootCause,
				RecommendedAction:      ev.RecommendedAction,
				EvidenceLocations: []repo_intel.EvidenceLocation{
					{
						File:        ev.FilePath,
						LineStart:   ev.Line,
						LineEnd:     ev.Line,
						Description: ev.AcquisitionSnippet,
					},
				},
				RemediationEligibility: eligibility,
				ApprovalRationale:      approvalRationale,
				UncertaintyClass:       string(uncertainty),
				ConcurrencyState:       string(ev.ConcurrencyState),
				AccessPath:             ev.AccessPath,
				Provenance:             repoCommit,
				Timestamp:              time.Now(),
			}
			addReconciled(rf, "SEMANTIC_RESOURCE_LEAK", fmt.Sprintf("Unclosed resource handle of type %s along exit paths", ev.ResourceKind))
		} else if ev.LifecycleState == StateResourceUnresolved {
			ledger.QuarantinedCount++
			ledger.Decisions = append(ledger.Decisions, ReconciliationDecision{
				Action:    "QUARANTINE",
				FindingID: fmt.Sprintf("SEM-REL-RES-%s-L%d", sanitizeRel(ev.FilePath), ev.Line),
				RuleID:    "RESOURCE_UNRESOLVED",
				Reason:    "Unverified struct storage or ownership transfer quarantined from autonomous remediation",
				Timestamp: time.Now(),
			})
		}
	}

	// 4. Process Semantic Error Evidence
	for _, ev := range errEvs {
		isVuln := ev.Classification == ErrorUnhandled ||
			(ev.Classification == ErrorExplicitlyIgnored && ev.Category != ErrorCatCleanup) ||
			ev.Classification == ErrorPartiallyHandled

		if isVuln {
			eligibility := EligibleAutonomous
			if ev.Classification == ErrorPartiallyHandled || ev.Category == ErrorCatCleanup {
				eligibility = EligibleHumanApproval
			}

			rf := ReconciledFinding{
				FindingID:         fmt.Sprintf("SEM-REL-ERR-%s-L%d", sanitizeRel(ev.FilePath), ev.Line),
				RuleID:            "SEMANTIC_UNHANDLED_ERROR",
				Category:          repo_intel.CategoryReliability,
				Severity:          ev.Severity,
				Confidence:        ev.Confidence,
				FindingState:      ev.FindingState,
				TargetFile:        ev.FilePath,
				LineStart:         ev.Line,
				LineEnd:           ev.Line,
				Title:             ev.Title,
				Description:       ev.Description,
				RootCause:         ev.RootCause,
				RecommendedAction: ev.RecommendedAction,
				EvidenceLocations: []repo_intel.EvidenceLocation{
					{
						File:        ev.FilePath,
						LineStart:   ev.Line,
						LineEnd:     ev.Line,
						Description: ev.CallSnippet,
					},
				},
				RemediationEligibility: eligibility,
				Provenance:             repoCommit,
				Timestamp:              time.Now(),
			}
			addReconciled(rf, "SEMANTIC_UNHANDLED_ERROR", fmt.Sprintf("Error return value from %s unhandled or discarded", ev.CallSnippet))
		}
	}

	// 5. Reconcile Raw Candidates (Heuristics)
	for _, cand := range rawCandidates {
		ledger.TotalProcessed++

		// Check if a verified semantic finding already exists for this exact defect
		isSuppressedOrMerged := false
		targetFile := ""
		lineStart := 0
		lineEnd := 0
		if len(cand.AffectedFiles) > 0 {
			targetFile = cand.AffectedFiles[0]
		}
		if len(cand.Evidence) > 0 {
			lineStart = cand.Evidence[0].LineStart
			lineEnd = cand.Evidence[0].LineEnd
		}

		for _, rf := range ledger.Findings {
			if targetFile != "" && targetFile == rf.TargetFile {
				// SQL injection heuristic covered by semantic SQL finding
				if cand.Category == repo_intel.CategorySecurity && rf.RuleID == "SEMANTIC_SQL_INJECTION" && lineStart == rf.LineStart {
					isSuppressedOrMerged = true
					ledger.MergedCount++
					ledger.Decisions = append(ledger.Decisions, ReconciliationDecision{
						Action:           "MERGE",
						FindingID:        cand.FindingID,
						RelatedFindingID: rf.FindingID,
						RuleID:           "HEURISTIC_UPGRADED_TO_SEMANTIC",
						Reason:           "Raw SQL regex heuristic subsumed by verified semantic SQL expression finding",
						Timestamp:        time.Now(),
					})
					break
				}
				// Command injection heuristic covered by semantic command finding
				if cand.Category == repo_intel.CategorySecurity && rf.RuleID == "SEMANTIC_COMMAND_INJECTION" && lineStart == rf.LineStart {
					isSuppressedOrMerged = true
					ledger.MergedCount++
					ledger.Decisions = append(ledger.Decisions, ReconciliationDecision{
						Action:           "MERGE",
						FindingID:        cand.FindingID,
						RelatedFindingID: rf.FindingID,
						RuleID:           "HEURISTIC_UPGRADED_TO_SEMANTIC",
						Reason:           "Raw command regex heuristic subsumed by verified semantic command finding",
						Timestamp:        time.Now(),
					})
					break
				}
			}
		}

		if !isSuppressedOrMerged {
			// Distinct heuristic finding (e.g. secret scanning, hardcoded credentials)
			rf := ReconciledFinding{
				FindingID:              cand.FindingID,
				RuleID:                 cand.DetectionMethod,
				Category:               cand.Category,
				Severity:               cand.Severity,
				Confidence:             cand.Confidence,
				FindingState:           StateProbable,
				TargetFile:             targetFile,
				LineStart:              lineStart,
				LineEnd:                lineEnd,
				Title:                  cand.Title,
				Description:            cand.Description,
				RootCause:              cand.RootCause,
				RecommendedAction:      cand.RecommendedAction,
				EvidenceLocations:      cand.Evidence,
				RemediationEligibility: EligibleHumanApproval,
				Provenance:             repoCommit,
				Timestamp:              time.Now(),
			}
			addReconciled(rf, cand.DetectionMethod, "Distinct heuristic finding retained")
		}
	}

	return ledger
}

func sanitizeRel(p string) string {
	s := strings.ReplaceAll(p, "\\", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "_")
	return s
}
