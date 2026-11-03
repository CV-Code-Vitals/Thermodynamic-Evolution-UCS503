package analysis

import (
	"time"
)

// DiagnosticError represents an in-depth investigation of a false-positive or false-negative finding.
type DiagnosticError struct {
	ErrorID                  string `json:"error_id"`
	ErrorType                string `json:"error_type"` // "FALSE_POSITIVE", "FALSE_NEGATIVE"
	TargetFixtureID          string `json:"target_fixture_id"`
	TargetFile               string `json:"target_file"`
	TargetSymbol             string `json:"target_symbol,omitempty"`
	FailureCategory          string `json:"failure_category"`
	RootCause                string `json:"root_cause"`
	DetectionLimitation      string `json:"detection_limitation"`
	ProposedCorrectiveAction string `json:"proposed_corrective_action"`
	RegressionFixture        string `json:"regression_fixture"`
}

// ErrorAnalysisReport aggregates the comprehensive error analysis for Milestone 3.0.
type ErrorAnalysisReport struct {
	Timestamp            time.Time         `json:"timestamp"`
	TotalErrorsAnalyzed  int               `json:"total_errors_analyzed"`
	FalsePositiveCount   int               `json:"false_positive_count"`
	FalseNegativeCount   int               `json:"false_negative_count"`
	Errors               []DiagnosticError `json:"errors"`
	StrategicActionItems []string          `json:"strategic_action_items"`
}

