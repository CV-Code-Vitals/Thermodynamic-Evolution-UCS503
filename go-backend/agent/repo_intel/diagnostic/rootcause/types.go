package rootcause

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// EdgeType defines causal, structural, and symptom relationships between findings.
type EdgeType string

const (
	EdgeSymptomOf     EdgeType = "SYMPTOM_OF"
	EdgeSharedOrigin  EdgeType = "SHARED_ORIGIN"
	EdgeCausalPrecede EdgeType = "CAUSAL_PRECEDE"
	EdgeCoRemediate   EdgeType = "CO_REMEDIATE"
)

// RootCausePattern defines recognizable architectural or code-level root cause archetypes.
type RootCausePattern string

const (
	PatternStructuralCongestion   RootCausePattern = "STRUCTURAL_CONGESTION"
	PatternErrorSuppression       RootCausePattern = "ERROR_SUPPRESSION_PATTERN"
	PatternArchitecturalDrift     RootCausePattern = "ARCHITECTURAL_DRIFT"
	PatternReproducibilityDeficit RootCausePattern = "REPRODUCIBILITY_DEFICIT"
	PatternResourceNeglect        RootCausePattern = "RESOURCE_LIFECYCLE_NEGLECT"
	PatternInputTaintLeak         RootCausePattern = "INPUT_TAINT_EXPOSURE"
	PatternVerificationDeficit    RootCausePattern = "VERIFICATION_DEFICIT"
	PatternMaintainabilityDecay   RootCausePattern = "MAINTAINABILITY_DECAY"
)

// GraphNode represents a finding or synthetic root cause in the causal graph.
type GraphNode struct {
	ID        string              `json:"id"`
	IsRoot    bool                `json:"is_root"`
	Category  repo_intel.Category `json:"category"`
	Finding   *repo_intel.Finding `json:"finding,omitempty"`
	File      string              `json:"file"`
	Title     string              `json:"title"`
	Weight    float64             `json:"weight"`
}

// GraphEdge represents a directed relationship between two nodes.
type GraphEdge struct {
	SourceID string   `json:"source_id"`
	TargetID string   `json:"target_id"`
	Type     EdgeType `json:"type"`
	Weight   float64  `json:"weight"`
	Rationale string  `json:"rationale"`
}

// RootCauseEntity aggregates multiple symptom findings under a unified root cause.
type RootCauseEntity struct {
	RootCauseID           string              `json:"root_cause_id"`
	Pattern               RootCausePattern    `json:"pattern"`
	Title                 string              `json:"title"`
	Description           string              `json:"description"`
	PrimaryFile           string              `json:"primary_file"`
	AffectedFiles         []string            `json:"affected_files"`
	SymptomFindingIDs     []string            `json:"symptom_finding_ids"`
	Confidence            float64             `json:"confidence"`
	ConfidenceRationale   string              `json:"confidence_rationale"`
	RemediationPriority   int                 `json:"remediation_priority"` // 1 (highest) to 5
	RecommendedResolution string              `json:"recommended_resolution"`
}

// RootCauseEvaluationReport captures the evaluation metrics for the root cause analyzer.
type RootCauseEvaluationReport struct {
	Timestamp            time.Time         `json:"timestamp"`
	TotalFindings        int               `json:"total_findings"`
	RootCausesIdentified int               `json:"root_causes_identified"`
	CompressionRatio     float64           `json:"compression_ratio"` // findings / root causes
	ClusterDistribution  map[string]int    `json:"cluster_distribution"`
	RootCauses           []RootCauseEntity `json:"root_causes"`
	EvaluationVerdict    string            `json:"evaluation_verdict"`
}

