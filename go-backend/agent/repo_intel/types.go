package repo_intel

import (
	"encoding/json"
	"time"
)

// Category classifications for findings.
type Category string

const (
	CategoryStructural   Category = "STRUCTURAL"
	CategoryArchitecture Category = "ARCHITECTURE"
	CategoryReliability  Category = "RELIABILITY"
	CategorySecurity     Category = "SECURITY"
	CategoryTestQuality     Category = "TEST_QUALITY"
	CategoryDependency      Category = "DEPENDENCY"
	CategoryMaintainability Category = "MAINTAINABILITY"
	CategoryDocumentation   Category = "DOCUMENTATION"
)

// Severity classifications for findings.
type Severity string

const (
	SeverityCritical      Severity = "CRITICAL"
	SeverityHigh          Severity = "HIGH"
	SeverityMedium        Severity = "MEDIUM"
	SeverityLow           Severity = "LOW"
	SeverityInformational Severity = "INFORMATIONAL"
)

// VerificationStatus indicates confidence/verification state.
type VerificationStatus string

const (
	StatusConfirmed             VerificationStatus = "CONFIRMED"
	StatusProbable              VerificationStatus = "PROBABLE"
	StatusHeuristic             VerificationStatus = "HEURISTIC"
	StatusRequiresInvestigation VerificationStatus = "REQUIRES_INVESTIGATION"
)

// EvidenceLocation pinpoints where an issue was found.
type EvidenceLocation struct {
	File        string `json:"file"`
	LineStart   int    `json:"line_start,omitempty"`
	LineEnd     int    `json:"line_end,omitempty"`
	ByteStart   int    `json:"byte_start,omitempty"`
	ByteEnd     int    `json:"byte_end,omitempty"`
	Description string `json:"description"`
	Snippet     string `json:"snippet,omitempty"`
}

// Finding represents a single structured, evidence-backed diagnostic finding.
type Finding struct {
	FindingID              string             `json:"finding_id"`
	Category               Category           `json:"category"`
	Severity               Severity           `json:"severity"`
	Confidence             float64            `json:"confidence"` // 0.0 to 1.0
	Title                  string             `json:"title"`
	Description            string             `json:"description"`
	RootCause              string             `json:"root_cause"`
	AffectedFiles          []string           `json:"affected_files"`
	AffectedSymbols        []string           `json:"affected_symbols,omitempty"`
	Evidence               []EvidenceLocation `json:"evidence"`
	RecommendedAction      string             `json:"recommended_action"`
	ExpectedBenefit        string             `json:"expected_benefit"`
	RegressionRisk         Severity           `json:"regression_risk"`
	ValidationRequirements []string           `json:"validation_requirements"`
	Dependencies           []string           `json:"dependencies,omitempty"`
	DetectionMethod        string             `json:"detection_method"`
	VerificationStatus     VerificationStatus `json:"verification_status"`
}

// LanguageStats records stats for a detected programming language.
type LanguageStats struct {
	Language         string `json:"language"`
	FileCount        int    `json:"file_count"`
	TotalBytes       int64  `json:"total_bytes"`
	TotalLines       int    `json:"total_lines"`
	ParsedCount      int    `json:"parsed_count"`
	ParseErrorsCount int    `json:"parse_errors_count"`
}

// ManifestDependency records a direct or dev dependency parsed from a manifest.
type ManifestDependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Type       string `json:"type"` // "direct", "dev", "indirect", "build"
	Constraint string `json:"constraint,omitempty"`
}

// ManifestInfo records a manifest file and its parsed dependencies.
type ManifestInfo struct {
	FilePath     string               `json:"file_path"`
	ManifestType string               `json:"manifest_type"` // "go.mod", "Cargo.toml", "package.json", "Dockerfile", etc.
	Dependencies []ManifestDependency `json:"dependencies"`
	HasLockfile  bool                 `json:"has_lockfile"`
	LockfilePath string               `json:"lockfile_path,omitempty"`
}

// ServiceBoundary records an identified service or module boundary.
type ServiceBoundary struct {
	Name        string   `json:"name"`
	RootPath    string   `json:"root_path"`
	EntryPoints []string `json:"entry_points"`
	Manifests   []string `json:"manifests"`
	Dockerfiles []string `json:"dockerfiles,omitempty"`
	Type        string   `json:"type"` // "backend", "frontend", "cli", "library", "service"
}

// RelationshipEdge represents a directed relationship between two entities.
type RelationshipEdge struct {
	Source     string  `json:"source"`
	Target     string  `json:"target"`
	Type       string  `json:"type"` // "import", "call", "manifest_dep", "service_dep"
	Evidence   string  `json:"evidence,omitempty"`
	Confidence float64 `json:"confidence"`
}

// DependencyGraph represents the cross-file and cross-module relationship graph.
type DependencyGraph struct {
	Nodes []string           `json:"nodes"`
	Edges []RelationshipEdge `json:"edges"`
}

// RepositoryMetadataRegistry stores comprehensive metadata about the scanned repository.
type RepositoryMetadataRegistry struct {
	RepoRoot          string                   `json:"repo_root"`
	GitBranch         string                   `json:"git_branch,omitempty"`
	HeadCommit        string                   `json:"head_commit,omitempty"`
	WorkingTreeClean  bool                     `json:"working_tree_clean"`
	Languages         map[string]LanguageStats `json:"languages"`
	TotalFiles        int                      `json:"total_files"`
	SourceFiles       []string                 `json:"source_files"`
	IgnoredFilesCount int                      `json:"ignored_files_count"`
	Manifests         []ManifestInfo           `json:"manifests"`
	ServiceBoundaries []ServiceBoundary        `json:"service_boundaries"`
	TestDirectories   []string                 `json:"test_directories"`
	CIWorkflows       []string                 `json:"ci_workflows"`
}

// IntelligenceReport is the master deliverable of the Repository Intelligence subsystem.
type IntelligenceReport struct {
	Version            string                     `json:"version"`
	GeneratedAt        time.Time                  `json:"generated_at"`
	RepositoryMetadata RepositoryMetadataRegistry `json:"repository_metadata"`
	DependencyGraph    DependencyGraph            `json:"dependency_graph"`
	Findings           []Finding                  `json:"findings"`
	AnalysisCoverage   AnalysisCoverage           `json:"analysis_coverage"`
	PerformanceMetrics PerformanceMetrics         `json:"performance_metrics"`
	Warnings           []string                   `json:"warnings,omitempty"`
}

// AnalysisCoverage records which modules ran and their success status.
type AnalysisCoverage struct {
	DiscoveryComplete    bool     `json:"discovery_complete"`
	ManifestsAnalyzed    int      `json:"manifests_analyzed"`
	StructuralFilesCount int      `json:"structural_files_count"`
	ArchitecturalModules int      `json:"architectural_modules"`
	ReliabilityCoverage  string   `json:"reliability_coverage"`
	SecurityCoverage     string   `json:"security_coverage"`
	QualityCoverage      string   `json:"quality_coverage"`
	ModuleErrors         []string `json:"module_errors,omitempty"`
}

// PerformanceMetrics tracks timing and resource usage.
type PerformanceMetrics struct {
	DiscoveryDurationMs     int64 `json:"discovery_duration_ms"`
	ManifestDurationMs      int64 `json:"manifest_duration_ms"`
	StructuralDurationMs    int64 `json:"structural_duration_ms"`
	ArchitecturalDurationMs int64 `json:"architectural_duration_ms"`
	SecurityDurationMs      int64 `json:"security_duration_ms"`
	QualityDurationMs       int64 `json:"quality_duration_ms"`
	TotalDurationMs         int64 `json:"total_duration_ms"`
	MemoryAllocatedBytes    int64 `json:"memory_allocated_bytes,omitempty"`
}

// Helper to serialize an IntelligenceReport to JSON.
func (r *IntelligenceReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

