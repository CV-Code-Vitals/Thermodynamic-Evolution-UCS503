package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/semantic"
)

type GroundTruthAmendment struct {
	AmendmentVersion string `json:"amendment_version"`
	CertifiedStrata  struct {
		InjectedDefects []struct {
			DefectID   string `json:"defect_id"`
			Repository string `json:"repository"`
			File       string `json:"file"`
			Line       int    `json:"line"`
			Symbol     string `json:"symbol"`
			Category   string `json:"category"`
			Severity   string `json:"severity"`
		} `json:"external_injected_defects"`
		CleanControls []struct {
			ControlID  string `json:"control_id"`
			Repository string `json:"repository"`
			File       string `json:"file"`
			Line       int    `json:"line"`
			Symbol     string `json:"symbol"`
		} `json:"clean_controls"`
	} `json:"certified_evaluation_strata"`
}

type FirstLookPredictionReport struct {
	EvaluationRunID      string                          `json:"evaluation_run_id"`
	Timestamp            string                          `json:"timestamp"`
	DetectorFrozenCommit string                          `json:"detector_frozen_commit"`
	RepositoryEvaluations []RepoEvaluationResult          `json:"repository_evaluations"`
	TotalDurationMs      int64                           `json:"total_duration_ms"`
	AllRawFindings       []semantic.ReconciledFinding    `json:"all_raw_findings"`
	ArtifactSHA256       string                          `json:"artifact_sha256,omitempty"`
}

type RepoEvaluationResult struct {
	Repository   string                       `json:"repository"`
	RelativePath string                       `json:"relative_path"`
	FilesScanned int                          `json:"files_scanned"`
	PackagesScanned int                       `json:"packages_scanned"`
	DurationMs   int64                        `json:"duration_ms"`
	ExitStatus   string                       `json:"exit_status"`
	FindingsCount int                         `json:"findings_count"`
	Findings     []semantic.ReconciledFinding `json:"findings"`
}

type CaseAuditItem struct {
	CaseID           string   `json:"case_id"`
	CaseType         string   `json:"case_type"` // "INJECTED_DEFECT" or "CLEAN_CONTROL"
	Repository       string   `json:"repository"`
	File             string   `json:"file"`
	Line             int      `json:"line"`
	Symbol           string   `json:"symbol"`
	GroundTruthClass string   `json:"ground_truth_class"` // Expected Category or "CLEAN"
	EvaluationOutcome string  `json:"evaluation_outcome"` // "TP", "FN", "TN", "FP"
	MatchedFindingIDs []string `json:"matched_finding_ids"`
	Notes            string   `json:"notes"`
}

type UnadjudicatedFindingItem struct {
	FindingID         string  `json:"finding_id"`
	Repository        string  `json:"repository"`
	File              string  `json:"file"`
	LineStart         int     `json:"line_start"`
	LineEnd           int     `json:"line_end"`
	RuleID            string  `json:"rule_id"`
	Category          string  `json:"category"`
	Severity          string  `json:"severity"`
	Confidence        float64 `json:"confidence"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	RemediationStatus string  `json:"remediation_status"`
	Stratum           string  `json:"stratum"` // "STRATUM_D_UNCERTAIN_UNLABELED"
}

type WilsonCI struct {
	PointEstimate float64 `json:"point_estimate"`
	Lower95       float64 `json:"lower_95"`
	Upper95       float64 `json:"upper_95"`
}

type StratumMetrics struct {
	DenominatorDescription string    `json:"denominator_description"`
	SampleCount            int       `json:"sample_count"`
	TP                     int       `json:"tp"`
	FP                     int       `json:"fp"`
	FN                     int       `json:"fn"`
	TN                     int       `json:"tn"`
	Recall                 *WilsonCI `json:"recall"`
	Precision              *WilsonCI `json:"precision"`
	Specificity            *WilsonCI `json:"specificity"`
	F1Score                float64   `json:"f1_score"`
	Status                 string    `json:"status"`
}

type ConfusionMatrixReport struct {
	EvaluationRunID      string                    `json:"evaluation_run_id"`
	DetectorFrozenCommit string                    `json:"detector_frozen_commit"`
	EvaluatedAt          string                    `json:"evaluated_at"`
	GroundTruthAmendment string                    `json:"ground_truth_amendment"`
	Strata               map[string]StratumMetrics `json:"strata"`
	SummaryTable         struct {
		TotalGroundTruthDefects  int     `json:"total_ground_truth_defects"`
		TotalCleanControls       int     `json:"total_clean_controls"`
		TotalUnadjudicatedCode   int     `json:"total_unadjudicated_code"`
		InjectedDefectRecall     float64 `json:"injected_defect_recall"`
		CleanControlSpecificity  float64 `json:"clean_control_specificity"`
		PrimaryPrecision         float64 `json:"primary_precision"`
		PrimaryF1Score           float64 `json:"primary_f1_score"`
	} `json:"summary_table"`
}

func computeWilsonCI(k, n int) *WilsonCI {
	if n <= 0 {
		return nil
	}
	p := float64(k) / float64(n)
	z := 1.959964 // 95% confidence
	z2 := z * z
	denom := 1.0 + z2/float64(n)
	center := (p + z2/(2.0*float64(n))) / denom
	spread := z * math.Sqrt((p*(1.0-p) + z2/(4.0*float64(n)))/float64(n)) / denom

	lower := math.Max(0.0, center-spread)
	upper := math.Min(1.0, center+spread)

	return &WilsonCI{
		PointEstimate: p,
		Lower95:       lower,
		Upper95:       upper,
	}
}

func main() {
	repoRoot := "E:\\SWE"
	frozenCommit := "a64e7632476c1341343481661a3e246c69e5a644"

	gtPath := filepath.Join(repoRoot, "benchmarks", "external_v2", "ground_truth_amendment_v2_1.json")
	gtData, err := os.ReadFile(gtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read ground truth amendment: %v\n", err)
		os.Exit(1)
	}

	var gt GroundTruthAmendment
	if err := json.Unmarshal(gtData, &gt); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse ground truth amendment: %v\n", err)
		os.Exit(1)
	}

	repos := []string{
		"fatih_color",
		"google_uuid",
		"gorilla_websocket",
		"hashicorp_go_version",
		"jinzhu_inflection",
		"pkg_errors",
		"schollz_progressbar",
		"spf13_cast",
		"stretchr_objx",
		"tidwall_gjson",
	}

	fmt.Println("================================================================================")
	fmt.Println("Milestone 3.4 Phase 6 / Phase 2: First-Look Blinded Evaluation")
	fmt.Printf("Detector Commit: %s\n", frozenCommit)
	fmt.Printf("Ground Truth:    %s\n", gtPath)
	fmt.Println("================================================================================")

	orch := semantic.NewUnifiedOrchestrator(repoRoot)

	predReport := FirstLookPredictionReport{
		EvaluationRunID:      fmt.Sprintf("EVAL-EXT2-%d", time.Now().Unix()),
		Timestamp:            time.Now().Format(time.RFC3339),
		DetectorFrozenCommit: frozenCommit,
	}

	t0Total := time.Now()
	var allFindings []semantic.ReconciledFinding

	for _, r := range repos {
		targetDir := filepath.Join(repoRoot, "benchmarks", "external_v2", r)
		fmt.Printf("[*] Evaluating repository: %-25s ... ", r)
		tStart := time.Now()

		report, err := orch.AnalyzeRepository(targetDir)
		dur := time.Since(tStart).Milliseconds()

		res := RepoEvaluationResult{
			Repository:   r,
			RelativePath: filepath.ToSlash(filepath.Join("benchmarks", "external_v2", r)),
			DurationMs:   dur,
		}

		if err != nil {
			res.ExitStatus = fmt.Sprintf("ERROR: %v", err)
			fmt.Printf("FAILED (%d ms): %v\n", dur, err)
		} else {
			res.ExitStatus = "SUCCESS"
			res.FilesScanned = report.FilesScanned
			res.PackagesScanned = report.PackagesScanned
			if report.ReconciliationLedger != nil {
				res.Findings = report.ReconciliationLedger.Findings
				res.FindingsCount = len(report.ReconciliationLedger.Findings)
				allFindings = append(allFindings, report.ReconciliationLedger.Findings...)
			}
			fmt.Printf("SUCCESS (%d ms, %d files, %d findings)\n", dur, res.FilesScanned, res.FindingsCount)
		}

		predReport.RepositoryEvaluations = append(predReport.RepositoryEvaluations, res)
	}

	predReport.TotalDurationMs = time.Since(t0Total).Milliseconds()
	predReport.AllRawFindings = allFindings

	// Export First-Look Predictions
	predJSON, _ := json.MarshalIndent(predReport, "", "  ")
	predHash := sha256.Sum256(predJSON)
	predHashHex := hex.EncodeToString(predHash[:])
	predReport.ArtifactSHA256 = predHashHex

	predOutPath := filepath.Join(repoRoot, "reports", "milestone3_4_phase6_phase2_firstlook_predictions.json")
	if err := os.WriteFile(predOutPath, predJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write firstlook predictions: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n[+] Predictions exported: %s (SHA-256: %s)\n", predOutPath, predHashHex)

	// Now Adjudicate against Ground Truth Amendment
	// Map findings by file (normalized)
	findingsByRepoAndFile := make(map[string][]semantic.ReconciledFinding)
	for _, f := range allFindings {
		normFile := filepath.ToSlash(f.TargetFile)
		findingsByRepoAndFile[normFile] = append(findingsByRepoAndFile[normFile], f)
	}

	var caseLedger []CaseAuditItem
	matchedFindingIDs := make(map[string]bool)

	// Evaluate Certified Injected Defects (n=6)
	tpCount := 0
	fnCount := 0
	for _, d := range gt.CertifiedStrata.InjectedDefects {
		normDefectFile := filepath.ToSlash(filepath.Join("benchmarks", "external_v2", d.Repository, d.File))
		candidates := findingsByRepoAndFile[normDefectFile]

		item := CaseAuditItem{
			CaseID:           d.DefectID,
			CaseType:         "INJECTED_DEFECT",
			Repository:       d.Repository,
			File:             d.File,
			Line:             d.Line,
			Symbol:           d.Symbol,
			GroundTruthClass: d.Category,
		}

		var matched []string
		for _, f := range candidates {
			// Check line overlap or proximity (line inside LineStart..LineEnd or within +/- 3 lines)
			lineMatch := (d.Line >= f.LineStart && d.Line <= f.LineEnd) ||
				(math.Abs(float64(d.Line-f.LineStart)) <= 3) ||
				(math.Abs(float64(d.Line-f.LineEnd)) <= 3)

			catMatch := false
			switch d.Category {
			case "SQL_INJECTION":
				catMatch = strings.Contains(strings.ToUpper(f.RuleID), "SQL")
			case "COMMAND_INJECTION":
				catMatch = strings.Contains(strings.ToUpper(f.RuleID), "COMMAND") || strings.Contains(strings.ToUpper(f.RuleID), "CMD")
			case "RESOURCE_LEAK":
				catMatch = strings.Contains(strings.ToUpper(f.RuleID), "RESOURCE") || strings.Contains(strings.ToUpper(f.RuleID), "LEAK")
			case "UNHANDLED_ERROR":
				catMatch = strings.Contains(strings.ToUpper(f.RuleID), "ERROR")
			}

			if lineMatch && catMatch {
				matched = append(matched, f.FindingID)
				matchedFindingIDs[f.FindingID] = true
			}
		}

		item.MatchedFindingIDs = matched
		if len(matched) > 0 {
			item.EvaluationOutcome = "TP"
			item.Notes = fmt.Sprintf("Detected accurately: %s", strings.Join(matched, ", "))
			tpCount++
		} else {
			item.EvaluationOutcome = "FN"
			item.Notes = "Missed by semantic analyzer"
			fnCount++
		}
		caseLedger = append(caseLedger, item)
	}

	// Evaluate Clean Controls (n=20)
	tnCount := 0
	fpCount := 0
	for _, c := range gt.CertifiedStrata.CleanControls {
		normCtrlFile := filepath.ToSlash(filepath.Join("benchmarks", "external_v2", c.Repository, c.File))
		candidates := findingsByRepoAndFile[normCtrlFile]

		item := CaseAuditItem{
			CaseID:           c.ControlID,
			CaseType:         "CLEAN_CONTROL",
			Repository:       c.Repository,
			File:             c.File,
			Line:             c.Line,
			Symbol:           c.Symbol,
			GroundTruthClass: "CLEAN",
		}

		var matched []string
		for _, f := range candidates {
			// Check if finding touches this control region (+/- 5 lines or containing line)
			lineMatch := (c.Line >= f.LineStart && c.Line <= f.LineEnd) ||
				(math.Abs(float64(c.Line-f.LineStart)) <= 5) ||
				(math.Abs(float64(c.Line-f.LineEnd)) <= 5)

			if lineMatch {
				matched = append(matched, f.FindingID)
				matchedFindingIDs[f.FindingID] = true
			}
		}

		item.MatchedFindingIDs = matched
		if len(matched) > 0 {
			item.EvaluationOutcome = "FP"
			item.Notes = fmt.Sprintf("False alarm triggered on clean control: %s", strings.Join(matched, ", "))
			fpCount++
		} else {
			item.EvaluationOutcome = "TN"
			item.Notes = "Correctly ignored, no false alarms"
			tnCount++
		}
		caseLedger = append(caseLedger, item)
	}

	// Write Case Ledger
	ledgerJSON, _ := json.MarshalIndent(caseLedger, "", "  ")
	ledgerPath := filepath.Join(repoRoot, "reports", "milestone3_4_phase6_phase2_case_ledger.json")
	if err := os.WriteFile(ledgerPath, ledgerJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write case ledger: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Case ledger exported: %s\n", ledgerPath)

	// Collect Stratum D: Unadjudicated Findings (Outside Labeled Regions)
	var unadjudicated []UnadjudicatedFindingItem
	for _, f := range allFindings {
		if !matchedFindingIDs[f.FindingID] {
			// Determine repository from target file
			normFile := filepath.ToSlash(f.TargetFile)
			repoName := "unknown"
			parts := strings.Split(normFile, "/")
			for i, p := range parts {
				if p == "external_v2" && i+1 < len(parts) {
					repoName = parts[i+1]
					break
				}
			}

			unadjudicated = append(unadjudicated, UnadjudicatedFindingItem{
				FindingID:         f.FindingID,
				Repository:        repoName,
				File:              f.TargetFile,
				LineStart:         f.LineStart,
				LineEnd:           f.LineEnd,
				RuleID:            f.RuleID,
				Category:          string(f.Category),
				Severity:          string(f.Severity),
				Confidence:        f.Confidence,
				Title:             f.Title,
				Description:       f.Description,
				RemediationStatus: string(f.RemediationEligibility),
				Stratum:           "STRATUM_D_UNCERTAIN_UNLABELED",
			})
		}
	}

	unadjJSON, _ := json.MarshalIndent(unadjudicated, "", "  ")
	unadjPath := filepath.Join(repoRoot, "reports", "milestone3_4_phase6_phase2_unadjudicated_findings.json")
	if err := os.WriteFile(unadjPath, unadjJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write unadjudicated findings: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Unadjudicated findings exported: %s (%d findings)\n", unadjPath, len(unadjudicated))

	// Construct Confusion Matrix Report
	cmReport := ConfusionMatrixReport{
		EvaluationRunID:      predReport.EvaluationRunID,
		DetectorFrozenCommit: frozenCommit,
		EvaluatedAt:          time.Now().Format(time.RFC3339),
		GroundTruthAmendment: "benchmarks/external_v2/ground_truth_amendment_v2_1.json",
		Strata:               make(map[string]StratumMetrics),
	}

	// Stratum A: Natural Defects (0 confirmed)
	cmReport.Strata["stratum_a_natural_defects"] = StratumMetrics{
		DenominatorDescription: "Naturally occurring external defects independently confirmed",
		SampleCount:            0,
		TP:                     0,
		FP:                     0,
		FN:                     0,
		TN:                     0,
		Recall:                 nil,
		Precision:              nil,
		Specificity:            nil,
		F1Score:                0.0,
		Status:                 "NOT_ESTIMABLE_ZERO_CONFIRMED",
	}

	// Stratum B: Injected Defects (n=6)
	recB := computeWilsonCI(tpCount, tpCount+fnCount)
	cmReport.Strata["stratum_b_injected_defects"] = StratumMetrics{
		DenominatorDescription: "Verified injected defects in external repositories",
		SampleCount:            tpCount + fnCount,
		TP:                     tpCount,
		FP:                     0,
		FN:                     fnCount,
		TN:                     0,
		Recall:                 recB,
		Status:                 "PRIMARY_SENSITIVITY_EVALUATION",
	}

	// Stratum C: Clean Controls (n=20)
	specC := computeWilsonCI(tnCount, tnCount+fpCount)
	cmReport.Strata["stratum_c_clean_controls"] = StratumMetrics{
		DenominatorDescription: "Audited negative control functions in external repositories",
		SampleCount:            tnCount + fpCount,
		TP:                     0,
		FP:                     fpCount,
		FN:                     0,
		TN:                     tnCount,
		Specificity:            specC,
		Status:                 "PRIMARY_SPECIFICITY_EVALUATION",
	}

	// Primary Injected+Clean Pooled Binary Metrics
	precPrimary := 0.0
	var precCI *WilsonCI
	if tpCount+fpCount > 0 {
		precPrimary = float64(tpCount) / float64(tpCount+fpCount)
		precCI = computeWilsonCI(tpCount, tpCount+fpCount)
	}

	recPrimary := 0.0
	if tpCount+fnCount > 0 {
		recPrimary = float64(tpCount) / float64(tpCount+fnCount)
	}

	f1Primary := 0.0
	if precPrimary+recPrimary > 0 {
		f1Primary = 2.0 * (precPrimary * recPrimary) / (precPrimary + recPrimary)
	}

	cmReport.Strata["stratum_bc_labeled_primary"] = StratumMetrics{
		DenominatorDescription: "Combined labeled evaluation ground truth (6 injected + 20 clean)",
		SampleCount:            tpCount + fnCount + tnCount + fpCount,
		TP:                     tpCount,
		FP:                     fpCount,
		FN:                     fnCount,
		TN:                     tnCount,
		Recall:                 recB,
		Precision:              precCI,
		Specificity:            specC,
		F1Score:                f1Primary,
		Status:                 "PRIMARY_BENCHMARK_SCORE",
	}

	cmReport.SummaryTable.TotalGroundTruthDefects = len(gt.CertifiedStrata.InjectedDefects)
	cmReport.SummaryTable.TotalCleanControls = len(gt.CertifiedStrata.CleanControls)
	cmReport.SummaryTable.TotalUnadjudicatedCode = len(unadjudicated)
	cmReport.SummaryTable.InjectedDefectRecall = recPrimary
	cmReport.SummaryTable.CleanControlSpecificity = float64(tnCount) / float64(tnCount+fpCount)
	cmReport.SummaryTable.PrimaryPrecision = precPrimary
	cmReport.SummaryTable.PrimaryF1Score = f1Primary

	cmJSON, _ := json.MarshalIndent(cmReport, "", "  ")
	cmPath := filepath.Join(repoRoot, "reports", "milestone3_4_phase6_phase2_confusion_matrix.json")
	if err := os.WriteFile(cmPath, cmJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write confusion matrix: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Confusion matrix exported: %s\n", cmPath)

	fmt.Println("\n================================================================================")
	fmt.Println("EVALUATION SUMMARY RESULTS")
	fmt.Println("================================================================================")
	fmt.Printf("Stratum A (Natural Defects):  0 confirmed -> RECALL NOT ESTIMABLE\n")
	fmt.Printf("Stratum B (Injected Defects): TP=%d, FN=%d -> Recall: %.2f%%\n", tpCount, fnCount, recPrimary*100)
	if recB != nil {
		fmt.Printf("                              Wilson 95%% CI: [%.2f%%, %.2f%%]\n", recB.Lower95*100, recB.Upper95*100)
	}
	fmt.Printf("Stratum C (Clean Controls):   TN=%d, FP=%d -> Specificity: %.2f%%\n", tnCount, fpCount, cmReport.SummaryTable.CleanControlSpecificity*100)
	if specC != nil {
		fmt.Printf("                              Wilson 95%% CI: [%.2f%%, %.2f%%]\n", specC.Lower95*100, specC.Upper95*100)
	}
	fmt.Printf("Precision (TP/(TP+FP)):       %.2f%%\n", precPrimary*100)
	fmt.Printf("F1-Score:                     %.2f%%\n", f1Primary*100)
	fmt.Printf("Stratum D (Unadjudicated):    %d findings in unlabeled external code (quarantined)\n", len(unadjudicated))
	fmt.Println("================================================================================")
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

