package semantic

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ErrorHandlingClassification categorizes the control-flow handling state of an error return value.
type ErrorHandlingClassification string

const (
	ErrorChecked              ErrorHandlingClassification = "ERROR_CHECKED"
	ErrorPropagated           ErrorHandlingClassification = "ERROR_PROPAGATED"
	ErrorWrappedAndPropagated ErrorHandlingClassification = "ERROR_WRAPPED_AND_PROPAGATED"
	ErrorExplicitlyIgnored    ErrorHandlingClassification = "ERROR_EXPLICITLY_IGNORED"
	ErrorUnhandled            ErrorHandlingClassification = "ERROR_UNHANDLED"
	ErrorPartiallyHandled     ErrorHandlingClassification = "ERROR_PARTIALLY_HANDLED"
	ErrorUnresolved           ErrorHandlingClassification = "ERROR_UNRESOLVED"
)

// ErrorOperationCategory classifies the domain and risk profile of an error-returning operation.
type ErrorOperationCategory string

const (
	ErrorCatSerialization ErrorOperationCategory = "SERIALIZATION"
	ErrorCatFileIO        ErrorOperationCategory = "FILE_IO"
	ErrorCatNetwork       ErrorOperationCategory = "NETWORK"
	ErrorCatDatabaseTx    ErrorOperationCategory = "DATABASE_TX"
	ErrorCatDatabaseQuery ErrorOperationCategory = "DATABASE_QUERY"
	ErrorCatCleanup       ErrorOperationCategory = "CLEANUP"
	ErrorCatTestHelper    ErrorOperationCategory = "TEST_HELPER"
	ErrorCatGeneral       ErrorOperationCategory = "GENERAL"
)

// SignatureResolutionType tracks whether function return types were derived via go/types or AST inference.
type SignatureResolutionType string

const (
	SignatureTypeResolved SignatureResolutionType = "TYPE_RESOLVED"
	SignatureASTInferred  SignatureResolutionType = "AST_INFERRED"
	SignatureUnresolved   SignatureResolutionType = "UNRESOLVED"
)

// ErrorReturnSignature describes the error-producing contract of a callable target.
type ErrorReturnSignature struct {
	IsErrorReturning bool                    `json:"is_error_returning"`
	ErrorIndex       int                     `json:"error_index"` // 0-based index of error in return values
	NumReturns       int                     `json:"num_returns"`
	Resolution       SignatureResolutionType `json:"resolution"`
	HasNamedReturns  bool                    `json:"has_named_returns"`
	NamedErrorVar    string                  `json:"named_error_var,omitempty"`
}

// ErrorCallInstance captures a discovered call site producing an error value.
type ErrorCallInstance struct {
	ID                int                         `json:"id"`
	CallSnippet       string                      `json:"call_snippet"`
	Line              int                         `json:"line"`
	BlockID           int                         `json:"block_id"`
	Package           string                      `json:"package"`
	FunctionOrMethod  string                      `json:"function_or_method"`
	Signature         ErrorReturnSignature        `json:"signature"`
	AssignedVar       string                      `json:"assigned_var"`
	IsBlankDiscard    bool                        `json:"is_blank_discard"`
	IsUnassignedExpr  bool                        `json:"is_unassigned_expr"`
	Classification    ErrorHandlingClassification `json:"classification"`
	Category          ErrorOperationCategory      `json:"category"`
	ReachingDef       string                      `json:"reaching_def"`
	History           []string                    `json:"history"`
	BranchEvidence    []string                    `json:"branch_evidence"`
	OverwrittenBefore bool                        `json:"overwritten_before"`
}

// ErrorEvidence represents a verified diagnostic evidence record for error-handling analysis.
type ErrorEvidence struct {
	FilePath            string                      `json:"file_path"`
	Package             string                      `json:"package"`
	Function            string                      `json:"function"`
	Line                int                         `json:"line"`
	CallSnippet         string                      `json:"call_snippet"`
	Signature           ErrorReturnSignature        `json:"signature"`
	Classification      ErrorHandlingClassification `json:"classification"`
	Category            ErrorOperationCategory      `json:"category"`
	Severity            repo_intel.Severity         `json:"severity"`
	Confidence          float64                     `json:"confidence"`
	FindingState        FindingState                `json:"finding_state"`
	Title               string                      `json:"title"`
	Description         string                      `json:"description"`
	RootCause           string                      `json:"root_cause"`
	RecommendedAction   string                      `json:"recommended_action"`
	BranchEvidence      []string                    `json:"branch_evidence,omitempty"`
	CounterEvidence     []string                    `json:"counter_evidence,omitempty"`
	AnalysisLimitations []string                    `json:"analysis_limitations,omitempty"`
	Timestamp           time.Time                   `json:"timestamp"`
}

