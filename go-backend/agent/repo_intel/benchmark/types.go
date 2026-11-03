package benchmark

import (
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/execution"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// GroundTruthLocation specifies exact code coordinates for a defect fixture.
type GroundTruthLocation struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// DefectFixture defines an injected ground-truth defect.
type DefectFixture struct {
	FixtureID                          string              `json:"fixture_id"`
	Category                           repo_intel.Category `json:"category"`
	DefectDescription                  string              `json:"defect_description"`
	TargetFile                         string              `json:"target_file"`
	TargetSymbol                       string              `json:"target_symbol"`
	GroundTruthLocation                GroundTruthLocation `json:"ground_truth_location"`
	ExpectedDiagnosticClassification   string              `json:"expected_diagnostic_classification"`
	ExpectedSeverity                   repo_intel.Severity `json:"expected_severity"`
	ExpectedConfidenceRange            [2]float64          `json:"expected_confidence_range"`
	ExpectedRemediationEligibility     string              `json:"expected_remediation_eligibility"`
	ExpectedExecutionStrategy          string              `json:"expected_execution_strategy"`
	ExpectedApprovalRequirements       bool                `json:"expected_approval_requirements"`
	ExpectedResolutionConditions       []string            `json:"expected_resolution_conditions"`
	RegressionTests                    []string            `json:"regression_tests"`
	FileSHA256                         string              `json:"file_sha256"`
}

// CleanFixture defines a benign control fixture designed to measure false positives.
type CleanFixture struct {
	FixtureID    string              `json:"fixture_id"`
	Category     repo_intel.Category `json:"category"`
	Description  string              `json:"description"`
	TargetFile   string              `json:"target_file"`
	TargetSymbol string              `json:"target_symbol"`
	FileSHA256   string              `json:"file_sha256"`
}

// GroundTruthManifest contains the ground-truth benchmark metadata.
type GroundTruthManifest struct {
	BenchmarkName       string          `json:"benchmark_name"`
	Version             string          `json:"version"`
	CreatedAt           time.Time       `json:"created_at"`
	TargetRepoPath      string          `json:"target_repository_path"`
	TotalFixtures       int             `json:"total_fixtures"`
	DefectFixturesCount int             `json:"defect_fixtures_count"`
	CleanFixturesCount  int             `json:"clean_fixtures_count"`
	DefectFixtures      []DefectFixture `json:"defect_fixtures"`
	CleanFixtures       []CleanFixture  `json:"clean_fixtures"`
	ManifestSHA256      string          `json:"manifest_sha256,omitempty"`
}

// BenchmarkRunConfig configures the end-to-end benchmark execution.
type BenchmarkRunConfig struct {
	BenchmarkRepoPath        string   `json:"benchmark_repo_path"`
	GroundTruthManifestPath  string   `json:"ground_truth_manifest_path"`
	OutputDirectory          string   `json:"output_directory"`
	ApprovedTaskIDs          []string `json:"approved_task_ids"`
	TargetTaskIDs            []string `json:"target_task_ids"`
	SimulateFailure          bool     `json:"simulate_failure"`
	CreateIntegrationBranch  bool     `json:"create_integration_branch"`
	IntegrationBranchName    string   `json:"integration_branch_name"`
	Repetitions              int      `json:"repetitions"`
	TimeoutSec               int      `json:"timeout_sec"`
}

// PipelineStageResult records the execution of one pipeline stage.
type PipelineStageResult struct {
	StageName          string        `json:"stage_name"`
	Status             string        `json:"status"` // "COMPLETED", "FAILED", "SKIPPED"
	DurationMs         int64         `json:"duration_ms"`
	ArtifactsProduced  []string      `json:"artifacts_produced"`
	ErrorMessage       string        `json:"error_message,omitempty"`
	Telemetry          interface{}   `json:"telemetry,omitempty"`
}

// DetectionMetrics evaluates finding detection against ground truth.
type DetectionMetrics struct {
	TruePositives   int               `json:"true_positives"`
	FalsePositives  int               `json:"false_positives"`
	FalseNegatives  int               `json:"false_negatives"`
	TrueNegatives   int               `json:"true_negatives"`
	Precision       float64           `json:"precision"`
	Recall          float64           `json:"recall"`
	F1Score         float64           `json:"f1_score"`
	MatchedFixtures map[string]string `json:"matched_fixtures"` // FixtureID -> FindingID
	UnmatchedClean  []string          `json:"unmatched_clean"`  // Clean fixtures correctly ignored
	FalseDetections []string          `json:"false_detections"` // FindingIDs on clean fixtures
}

// RemediationMetrics evaluates patch generation, validation, and integration effectiveness.
type RemediationMetrics struct {
	EligibleGroundTruthDefects int     `json:"eligible_ground_truth_defects"`
	VerifiedRemediations       int     `json:"verified_remediations"`
	VerifiedRemediationRate    float64 `json:"verified_remediation_rate"`
	IntroducedRegressions      int     `json:"introduced_regressions"`
	RegressionRate             float64 `json:"regression_rate"`
	FalseRemediations          int     `json:"false_remediations"`
	FalseRemediationRate       float64 `json:"false_remediation_rate"`
	AcceptedPatches            int     `json:"accepted_patches"`
	IntegratedPatches          int     `json:"integrated_patches"`
	IntegrationSuccessRate     float64 `json:"integration_success_rate"`
}

// BeforeAfterComparison measures diagnostic metrics before and after remediation.
type BeforeAfterComparison struct {
	BeforeFindingsCount     int            `json:"before_findings_count"`
	AfterFindingsCount      int            `json:"after_findings_count"`
	BeforeBySeverity        map[string]int `json:"before_by_severity"`
	AfterBySeverity         map[string]int `json:"after_by_severity"`
	ResolvedFindingIDs      []string       `json:"resolved_finding_ids"`
	UnresolvedFindingIDs    []string       `json:"unresolved_finding_ids"`
	NewlyIntroducedFindings []string       `json:"newly_introduced_findings"`
	NetFindingsDelta        int            `json:"net_findings_delta"`
}

// BenchmarkRunReport compiles the complete end-to-end benchmark evaluation.
type BenchmarkRunReport struct {
	RunID                   string                     `json:"run_id"`
	Timestamp               time.Time                  `json:"timestamp"`
	Config                  BenchmarkRunConfig         `json:"config"`
	ManifestSummary         GroundTruthManifest        `json:"manifest_summary"`
	StageResults            []PipelineStageResult      `json:"stage_results"`
	DetectionMetrics        DetectionMetrics           `json:"detection_metrics"`
	RemediationMetrics      RemediationMetrics         `json:"remediation_metrics"`
	BeforeAfterComparison   BeforeAfterComparison      `json:"before_after_comparison"`
	IntegrationBranch       string                     `json:"integration_branch,omitempty"`
	IntegrationCommit       string                     `json:"integration_commit,omitempty"`
	ReproducibilityVerified bool                       `json:"reproducibility_verified"`
	TotalDurationMs         int64                      `json:"total_duration_ms"`
	OverallVerdict          string                     `json:"overall_verdict"` // "GO", "CONDITIONAL_GO", "NO_GO"

	// Live References for Pipeline Stages
	IntelReport     *repo_intel.IntelligenceReport    `json:"-"`
	DiagReport      *diagnostic.DiagnosticReport       `json:"-"`
	Plan            *planning.RemediationPlan          `json:"-"`
	ExecutionReport *execution.ExecutionReport         `json:"-"`
}
