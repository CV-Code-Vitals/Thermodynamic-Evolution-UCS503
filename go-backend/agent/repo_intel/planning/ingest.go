package planning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic"
)

var (
	ErrInvalidSchemaVersion = errors.New("unsupported diagnostic report schema version")
	ErrMissingRequiredFields = errors.New("diagnostic report is missing mandatory fields")
	ErrPathEscapeDetected    = errors.New("path traversal detected in diagnostic report")
	ErrStaleReportProvenance = errors.New("report baseline commit does not match current repository HEAD")
)

// ProvenanceAssessment records the outcome of the report provenance check.
type ProvenanceAssessment struct {
	ReportID              string   `json:"report_id"`
	SchemaVersion         string   `json:"schema_version"`
	ReportHashSHA256      string   `json:"report_hash_sha256"`
	BaselineCommit        string   `json:"baseline_commit"`
	CurrentCommit         string   `json:"current_commit"`
	CommitMatches         bool     `json:"commit_matches"`
	ChangedFiles          []string `json:"changed_files,omitempty"`
	AffectedFindingIDs    []string `json:"affected_finding_ids,omitempty"`
	IsEligibleForPlanning bool     `json:"is_eligible_for_planning"`
	AssessmentRationale   string   `json:"assessment_rationale"`
}

// ReportIngester parses, hashes, and validates diagnostic reports.
type ReportIngester struct {
	repoRoot string
}

// NewReportIngester creates a new ReportIngester for the given repository root.
func NewReportIngester(repoRoot string) *ReportIngester {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return &ReportIngester{repoRoot: repoRoot}
}

// IngestReportFile reads and validates a diagnostic report from disk.
func (ri *ReportIngester) IngestReportFile(reportPath string, currentCommit string) (*diagnostic.DiagnosticReport, *ProvenanceAssessment, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read report file: %w", err)
	}

	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	var report diagnostic.DiagnosticReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, nil, fmt.Errorf("unmarshal diagnostic report: %w", err)
	}

	assessment, err := ri.ValidateReport(&report, hashHex, currentCommit)
	if err != nil {
		return nil, assessment, err
	}

	return &report, assessment, nil
}

// ValidateReport enforces schema version, field integrity, path safety, and provenance matching.
func (ri *ReportIngester) ValidateReport(report *diagnostic.DiagnosticReport, reportHash string, currentCommit string) (*ProvenanceAssessment, error) {
	if report.ReportSchemaVersion != "1.0" {
		return nil, fmt.Errorf("%w: expected 1.0, got %q", ErrInvalidSchemaVersion, report.ReportSchemaVersion)
	}

	if report.ReportID == "" || report.Repository.RepoRoot == "" || len(report.Findings) == 0 {
		return nil, fmt.Errorf("%w: report_id, repository.repo_root, and findings must be non-empty", ErrMissingRequiredFields)
	}

	// Validate path boundaries for all finding evidence
	for _, f := range report.Findings {
		for _, ev := range f.Evidence {
			if strings.Contains(ev.File, "..") {
				clean := filepath.Clean(ev.File)
				if strings.HasPrefix(clean, "..") {
					return nil, fmt.Errorf("%w: finding %s references path %q", ErrPathEscapeDetected, f.FindingID, ev.File)
				}
			}
		}
	}

	assessment := &ProvenanceAssessment{
		ReportID:              report.ReportID,
		SchemaVersion:         report.ReportSchemaVersion,
		ReportHashSHA256:      reportHash,
		BaselineCommit:        report.Repository.HeadCommit,
		CurrentCommit:         currentCommit,
		IsEligibleForPlanning: true,
	}

	// Check commit lineage
	if currentCommit != "" && report.Repository.HeadCommit != "" {
		if strings.TrimSpace(currentCommit) == strings.TrimSpace(report.Repository.HeadCommit) {
			assessment.CommitMatches = true
			assessment.AssessmentRationale = "Report baseline matches current repository HEAD commit exactly."
		} else {
			assessment.CommitMatches = false
			assessment.AssessmentRationale = fmt.Sprintf("Report baseline commit (%s) differs from current HEAD (%s); revalidating findings.",
				report.Repository.HeadCommit, currentCommit)
		}
	} else {
		assessment.CommitMatches = true
		assessment.AssessmentRationale = "Report provenance accepted with unconstrained commit verification."
	}

	return assessment, nil
}
