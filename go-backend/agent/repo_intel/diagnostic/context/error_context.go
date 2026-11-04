package context

import (
	"fmt"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/semantic"
)

// ErrorContextAnalyzer evaluates Go source files for unhandled, suppressed, or partially handled error return values.
type ErrorContextAnalyzer struct {
	repoRoot string
}

func NewErrorContextAnalyzer(repoRoot string) *ErrorContextAnalyzer {
	return &ErrorContextAnalyzer{repoRoot: repoRoot}
}

// ErrorContextEvaluation holds findings and counter-evidences for error handling inspection.
type ErrorContextEvaluation struct {
	Findings         []ContextFinding
	CounterEvidences []CounterEvidence
}

// EvaluateFile performs Semantic IR and CFG context analysis on a Go file for error handling safety.
func (a *ErrorContextAnalyzer) EvaluateFile(relPath string) (*ErrorContextEvaluation, error) {
	eval := &ErrorContextEvaluation{
		Findings:         make([]ContextFinding, 0),
		CounterEvidences: make([]CounterEvidence, 0),
	}

	semAnalyzer := semantic.NewSemanticAnalyzer(a.repoRoot)
	fileModel, _, err := semAnalyzer.AnalyzeFile(relPath)
	if err != nil || fileModel == nil {
		return eval, nil
	}

	errAnalyzer := semantic.NewErrorAnalyzer(a.repoRoot)
	for _, cfg := range fileModel.CFGs {
		errAnalyzer.RegisterLocalFunction(cfg.FuncDecl)
	}

	findingIDsSeen := make(map[string]bool)

	for fnName, cfg := range fileModel.CFGs {
		df := fileModel.Dataflow[fnName]
		evs, err := errAnalyzer.AnalyzeFunction(cfg.FuncDecl, fileModel.FileSet, cfg, df, relPath, nil)
		if err != nil {
			continue
		}

		for _, ev := range evs {
			findingID := fmt.Sprintf("CTX-REL-ERROR-%s-L%d", sanitizeRel(relPath), ev.Line)

			isVulnerability := ev.Classification == semantic.ErrorUnhandled ||
				(ev.Classification == semantic.ErrorExplicitlyIgnored && ev.Category != semantic.ErrorCatCleanup) ||
				ev.Classification == semantic.ErrorPartiallyHandled

			if isVulnerability {
				if !findingIDsSeen[findingID] {
					findingIDsSeen[findingID] = true
					eval.Findings = append(eval.Findings, ContextFinding{
						FindingID:           findingID,
						Category:            repo_intel.CategoryReliability,
						Severity:            ev.Severity,
						Confidence:          ev.Confidence,
						ConfidenceRationale: ev.RootCause,
						State:               FindingState(ev.FindingState),
						Title:               ev.Title,
						Description:         ev.Description,
						RootCause:           ev.RootCause,
						AffectedFiles:       []string{relPath},
						EvidenceTraces: []EvidenceTrace{
							{
								Location: repo_intel.EvidenceLocation{
									File:        relPath,
									LineStart:   ev.Line,
									LineEnd:     ev.Line,
									Description: ev.CallSnippet,
								},
								ASTNodeType: "ast.CallExpr",
								ContextType: "ERROR_CALL_SITE",
								CodeSnippet: ev.CallSnippet,
								Rationale:   fmt.Sprintf("Error return value classification: %s (Category: %s)", ev.Classification, ev.Category),
							},
						},
						DetectionMethod:   "SEMANTIC_ERROR_ANALYSIS",
						RecommendedAction: ev.RecommendedAction,
					})
				}
			} else {
				eval.CounterEvidences = append(eval.CounterEvidences, CounterEvidence{
					RuleID:      "SAFE_ERROR_HANDLING",
					Description: fmt.Sprintf("Error from %s safely handled: %s", ev.CallSnippet, ev.Classification),
					Location:    fmt.Sprintf("%s:%d", relPath, ev.Line),
					VerifiedAt:  time.Now(),
				})
			}
		}
	}

	return eval, nil
}
