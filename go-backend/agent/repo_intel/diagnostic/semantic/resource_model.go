package semantic

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ResourceKind represents the concrete system resource abstraction being tracked.
type ResourceKind string

const (
	ResourceKindFile     ResourceKind = "os.File"
	ResourceKindNetConn  ResourceKind = "net.Conn"
	ResourceKindHTTPBody ResourceKind = "http.Response.Body"
	ResourceKindSQLRows  ResourceKind = "sql.Rows"
	ResourceKindSQLTx    ResourceKind = "sql.Tx"
	ResourceKindTimer    ResourceKind = "time.Timer"
)

// ResourceLifecycleState classifies the lifecycle status of a tracked resource handle.
type ResourceLifecycleState string

const (
	StateResourceAcquired        ResourceLifecycleState = "RESOURCE_ACQUIRED"
	StateResourceReleased        ResourceLifecycleState = "RESOURCE_RELEASED"
	StateResourceDeferredRelease ResourceLifecycleState = "RESOURCE_DEFERRED_RELEASE"
	StateResourceTransferred     ResourceLifecycleState = "RESOURCE_TRANSFERRED"
	StateResourceEscaped         ResourceLifecycleState = "RESOURCE_ESCAPED"
	StateResourceUnresolved      ResourceLifecycleState = "RESOURCE_UNRESOLVED"
	StateResourceLeakConfirmed   ResourceLifecycleState = "RESOURCE_LEAK_CONFIRMED"
	StateResourceLeakProbable    ResourceLifecycleState = "RESOURCE_LEAK_PROBABLE"
)

// ResourceTransition captures a single lifecycle event transition in the resource's history.
type ResourceTransition struct {
	FromState ResourceLifecycleState `json:"from_state"`
	ToState   ResourceLifecycleState `json:"to_state"`
	Action    string                 `json:"action"` // e.g. "os.Open", "defer f.Close()", "rows.Close()"
	Line      int                    `json:"line"`
	BlockID   int                    `json:"block_id"`
	Detail    string                 `json:"detail,omitempty"`
}

// ResourceInstance models a concrete resource allocation across its lifetime in a function.
type ResourceInstance struct {
	ID                 int                    `json:"id"`
	Kind               ResourceKind           `json:"kind"`
	VarName            string                 `json:"var_name"`
	ParentVarName      string                 `json:"parent_var_name,omitempty"` // e.g. "resp" for resp.Body
	Aliases            []string               `json:"aliases"`
	AcquisitionCall    string                 `json:"acquisition_call"`
	AcquisitionLine    int                    `json:"acquisition_line"`
	AcquisitionBlockID int                    `json:"acquisition_block_id"`
	IsDeferred         bool                   `json:"is_deferred"`
	DeferredInLoop     bool                   `json:"deferred_in_loop"`
	CurrentState       ResourceLifecycleState `json:"current_state"`
	History            []ResourceTransition   `json:"history"`
	ReachableExits     int                    `json:"reachable_exits"`
	CleanExits         int                    `json:"clean_exits"`
	LeakedExits        int                    `json:"leaked_exits"`
	EscapeReason       string                 `json:"escape_reason,omitempty"`
}

// AddTransition appends an event to the resource instance history and updates its current state.
func (r *ResourceInstance) AddTransition(toState ResourceLifecycleState, action string, line, blockID int, detail string) {
	trans := ResourceTransition{
		FromState: r.CurrentState,
		ToState:   toState,
		Action:    action,
		Line:      line,
		BlockID:   blockID,
		Detail:    detail,
	}
	r.History = append(r.History, trans)
	r.CurrentState = toState
}

// ResourceEvidence represents an auditable diagnostic finding or certification for a resource.
type ResourceEvidence struct {
	FilePath            string                 `json:"file_path"`
	Package             string                 `json:"package"`
	Function            string                 `json:"function"`
	Line                int                    `json:"line"`
	ResourceKind        ResourceKind           `json:"resource_kind"`
	VarName             string                 `json:"var_name"`
	LifecycleState      ResourceLifecycleState `json:"lifecycle_state"`
	FindingState        FindingState           `json:"finding_state"`
	Severity            repo_intel.Severity    `json:"severity"`
	Confidence          float64                `json:"confidence"`
	AcquisitionSnippet  string                 `json:"acquisition_snippet"`
	ReleaseSnippet      string                 `json:"release_snippet,omitempty"`
	History             []ResourceTransition   `json:"history"`
	AnalysisLimitations []string               `json:"analysis_limitations,omitempty"`
	CounterEvidence     []string               `json:"counter_evidence,omitempty"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	RootCause           string                 `json:"root_cause,omitempty"`
	RecommendedAction   string                 `json:"recommended_action,omitempty"`
	Timestamp           time.Time              `json:"timestamp"`
}
