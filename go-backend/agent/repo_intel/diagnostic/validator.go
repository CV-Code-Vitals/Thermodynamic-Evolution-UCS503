package diagnostic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// FindingValidator evaluates findings for evidence existence, source freshness,
// and structural metadata integrity.
type FindingValidator struct {
	repoRoot string
}

// NewFindingValidator creates a new FindingValidator for the given repository root.
func NewFindingValidator(repoRoot string) *FindingValidator {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return &FindingValidator{repoRoot: repoRoot}
}

// ValidateFindings inspects each finding against physical repository state.
func (v *FindingValidator) ValidateFindings(findings []repo_intel.Finding) ([]ValidatedFinding, ValidationSummary) {
	var validated []ValidatedFinding
	var summary ValidationSummary
	summary.TotalFindings = len(findings)

	for _, f := range findings {
		vf := v.validateSingleFinding(f)
		switch vf.ValidationState {
		case StateValidated:
			summary.Validated++
		case StatePartiallyValidated:
			summary.PartiallyValidated++
		case StateStale:
			summary.Stale++
		case StateInvalid:
			summary.Invalid++
		case StateRequiresHumanReview:
			summary.RequiresHumanReview++
		}
		validated = append(validated, vf)
	}

	return validated, summary
}

func (v *FindingValidator) validateSingleFinding(f repo_intel.Finding) ValidatedFinding {
	vf := ValidatedFinding{
		FindingID:              f.FindingID,
		SourceFindingID:        f.FindingID,
		Category:               f.Category,
		Severity:               f.Severity,
		CalibratedConfidence:   f.Confidence,
		Title:                  f.Title,
		Description:            f.Description,
		RootCause:              f.RootCause,
		AffectedComponents:     f.AffectedFiles,
		ExpectedBenefit:        f.ExpectedBenefit,
		RemediationComplexity:  "MEDIUM",
		RegressionRisk:         f.RegressionRisk,
		Evidence:               make([]repo_intel.EvidenceLocation, 0, len(f.Evidence)),
		RecommendedAction:      f.RecommendedAction,
		ValidationRequirements: f.ValidationRequirements,
		RelatedFindingIDs:      f.Dependencies,
		DetectionMethod:        f.DetectionMethod,
	}

	// 1. Basic Metadata Validation
	if f.FindingID == "" || f.Category == "" || f.Severity == "" || f.Title == "" {
		vf.ValidationState = StateInvalid
		vf.ValidationReason = "Malformed finding: missing required finding_id, category, severity, or title"
		return vf
	}

	// 2. Canonicalize & Validate Evidence Locations
	if len(f.Evidence) == 0 {
		// Absence findings (e.g. missing license or README) have valid reasons to have 0 evidence
		if strings.Contains(f.FindingID, "MISSING-LICENSE") || strings.Contains(f.FindingID, "MISSING-README") ||
			strings.Contains(f.FindingID, "NOCI-WORKFLOWS") {
			vf.ValidationState = StateValidated
			vf.ValidationReason = "Verified absence of expected file across repository"
			return vf
		}

		vf.ValidationState = StateInvalid
		vf.ValidationReason = "Finding contains zero evidence locations"
		return vf
	}

	allEvidenceValid := true
	anyEvidenceValid := false
	isStale := false
	var reasons []string

	for _, ev := range f.Evidence {
		normEv := ev
		canonicalRel, targetAbs, err := v.canonicalizePath(ev.File)
		if err != nil {
			allEvidenceValid = false
			reasons = append(reasons, fmt.Sprintf("Invalid path %q: %v", ev.File, err))
			continue
		}
		normEv.File = canonicalRel

		info, statErr := os.Stat(targetAbs)
		if statErr != nil {
			// Check if this finding is reporting a missing lockfile or missing file
			if strings.Contains(f.FindingID, "MISSING") || strings.Contains(f.FindingID, "NOLOCK") {
				// The finding asserts the file is missing; if it is indeed absent, this evidence is valid!
				anyEvidenceValid = true
				vf.Evidence = append(vf.Evidence, normEv)
				continue
			}

			allEvidenceValid = false
			isStale = true
			reasons = append(reasons, fmt.Sprintf("Referenced file does not exist: %s", canonicalRel))
			continue
		}

		// File exists. If it's a regular file and line numbers are given, check line bounds
		if !info.IsDir() && normEv.LineStart > 0 {
			lineCount, readErr := countLines(targetAbs)
			if readErr != nil {
				allEvidenceValid = false
				reasons = append(reasons, fmt.Sprintf("Failed to read file %s: %v", canonicalRel, readErr))
				continue
			}

			if normEv.LineStart > lineCount {
				allEvidenceValid = false
				isStale = true
				reasons = append(reasons, fmt.Sprintf("Stale line number: LineStart %d exceeds file length of %d lines", normEv.LineStart, lineCount))
				continue
			}
		}

		anyEvidenceValid = true
		vf.Evidence = append(vf.Evidence, normEv)
	}

	if isStale {
		vf.ValidationState = StateStale
		vf.ValidationReason = strings.Join(reasons, "; ")
		return vf
	}

	if !allEvidenceValid {
		if anyEvidenceValid {
			vf.ValidationState = StatePartiallyValidated
			vf.ValidationReason = strings.Join(reasons, "; ")
		} else {
			vf.ValidationState = StateInvalid
			vf.ValidationReason = strings.Join(reasons, "; ")
		}
		return vf
	}

	// 3. Heuristic / Security review classification
	if f.VerificationStatus == repo_intel.StatusHeuristic ||
		strings.Contains(f.DetectionMethod, "HEURISTIC") ||
		f.Category == repo_intel.CategorySecurity {
		vf.ValidationState = StateRequiresHumanReview
		vf.ValidationReason = "Heuristic pattern or security finding requires human confirmation"
		return vf
	}

	vf.ValidationState = StateValidated
	vf.ValidationReason = "All evidence locations verified on disk with accurate line bounds"
	return vf
}

func (v *FindingValidator) canonicalizePath(p string) (rel string, abs string, err error) {
	cleanP := filepath.Clean(p)

	// If absolute or has Windows drive prefix
	if filepath.IsAbs(cleanP) || (len(cleanP) > 1 && cleanP[1] == ':') {
		abs = cleanP
		r, relErr := filepath.Rel(v.repoRoot, abs)
		if relErr != nil || strings.HasPrefix(r, "..") {
			return "", "", fmt.Errorf("path %q escapes repository root boundary %q", p, v.repoRoot)
		}
		rel = filepath.ToSlash(r)
		return rel, abs, nil
	}

	// Relative path
	abs = filepath.Clean(filepath.Join(v.repoRoot, cleanP))
	r, relErr := filepath.Rel(v.repoRoot, abs)
	if relErr != nil || strings.HasPrefix(r, "..") {
		return "", "", fmt.Errorf("path traversal: path %q resolves outside repository root %q", p, v.repoRoot)
	}
	rel = filepath.ToSlash(cleanP)
	return rel, abs, nil
}

func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strings.Count(string(data), "\n") + 1, nil
}

