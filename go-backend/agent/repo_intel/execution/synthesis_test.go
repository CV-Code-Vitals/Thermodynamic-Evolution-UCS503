package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// SynthesisTrialRecord records empirical evidence for a single synthesis target.
type SynthesisTrialRecord struct {
	CaseID                 string                 `json:"case_id"`
	Category               string                 `json:"category"`
	TargetFile             string                 `json:"target_file"`
	ExpectedIntent         string                 `json:"expected_intent"`
	ActualStatus           string                 `json:"actual_status"` // PROPOSED, ABSTAINED, FAILED
	AbstainReason          string                 `json:"abstain_reason,omitempty"`
	PatchGenerated         bool                   `json:"patch_generated"`
	UnifiedDiff            string                 `json:"unified_diff,omitempty"`
	DiffHash               string                 `json:"diff_hash,omitempty"`
	IndependentValidations map[string]bool        `json:"independent_validations,omitempty"`
	AllOraclesPassed       bool                   `json:"all_oracles_passed"`
	DurationMs             int64                  `json:"duration_ms"`
}

// SynthesisBenchmarkReport aggregates Phase 1.5 synthesis trial results.
type SynthesisBenchmarkReport struct {
	Timestamp          time.Time              `json:"timestamp"`
	TotalCases         int                    `json:"total_cases"`
	SynthesizedCount   int                    `json:"synthesized_count"`
	AbstainedCount     int                    `json:"abstained_count"`
	FailedCount        int                    `json:"failed_count"`
	VerifiedCount      int                    `json:"verified_count"`
	SynthesisRate      float64                `json:"synthesis_rate"`
	AbstentionRate     float64                `json:"abstention_rate"`
	OracleSuccessRate  float64                `json:"oracle_success_rate"`
	Records            []SynthesisTrialRecord `json:"records"`
}

func TestAutonomousPatchSynthesisAndIndependentOracles(t *testing.T) {
	synthesizer := NewAutonomousPatchSynthesizer()
	oracles := NewIndependentOracleSuite(15)

	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "synth-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp fixture directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		caseID         string
		category       string
		relPath        string
		code           string
		task           *planning.RemediationTask
		finding        *repo_intel.Finding
		expectedStatus SynthesisDecisionStatus
		expectedIntent IntentType
		expectAbstain  AbstainCategory
	}{
		// Case 1: Resource Leak - http.Get response body
		{
			caseID:   "SYNTH-RES-01",
			category: "RESOURCE_LEAK",
			relPath:  "service/http_fetcher.go",
			code: `package service

import (
	"io"
	"net/http"
)

func FetchData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	return body, err
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-RES-01",
				Title:       "Close HTTP Response Body",
				Description: "Fix resource leak: deferred Close() missing for http.Get response body",
				TargetFiles: []string{"service/http_fetcher.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-RES-01",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Resource leak on http response body",
			},
			expectedStatus: DecisionProposed,
			expectedIntent: IntentResourceCleanup,
		},

		// Case 2: Resource Leak - os.Open file handle
		{
			caseID:   "SYNTH-RES-02",
			category: "RESOURCE_LEAK",
			relPath:  "storage/file_reader.go",
			code: `package storage

import (
	"io"
	"os"
)

func ReadConfig(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-RES-02",
				Title:       "Close File Handle",
				Description: "Fix resource leak: file opened with os.Open is never closed",
				TargetFiles: []string{"storage/file_reader.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-RES-02",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Unclosed file descriptor leak",
			},
			expectedStatus: DecisionProposed,
			expectedIntent: IntentResourceCleanup,
		},

		// Case 3: Unhandled Error Propagation
		{
			caseID:   "SYNTH-ERR-01",
			category: "UNHANDLED_ERROR",
			relPath:  "worker/cleaner.go",
			code: `package worker

import (
	"os"
)

func CleanWorkspace(path string) error {
	_ = os.Remove(path)
	return nil
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-ERR-01",
				Title:       "Handle Discarded Error",
				Description: "Fix unhandled error: return value of os.Remove discarded via blank identifier",
				TargetFiles: []string{"worker/cleaner.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-ERR-01",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Discarded error return value",
			},
			expectedStatus: DecisionProposed,
			expectedIntent: IntentHandleError,
		},

		// Case 4: SQL Parameterization
		{
			caseID:   "SYNTH-SQL-01",
			category: "SQL_INJECTION",
			relPath:  "db/user_repo.go",
			code: `package db

import (
	"context"
	"database/sql"
	"fmt"
)

func QueryUser(ctx context.Context, db *sql.DB, username string) error {
	query := fmt.Sprintf("SELECT id, username FROM users WHERE username = '%s'", username)
	_, err := db.ExecContext(ctx, query)
	return err
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-SQL-01",
				Title:       "Parameterize SQL Query",
				Description: "Fix SQL injection: dynamic fmt.Sprintf query concatenation",
				TargetFiles: []string{"db/user_repo.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-SQL-01",
				Category:  repo_intel.CategorySecurity,
				Title:     "Dynamic SQL query concatenation",
			},
			expectedStatus: DecisionProposed,
			expectedIntent: IntentParameterizeSQL,
		},

		// Case 5: Abstain on Command Injection (Requires Architectural Redesign)
		{
			caseID:   "SYNTH-CMD-ABSTAIN",
			category: "COMMAND_INJECTION",
			relPath:  "cmd/exec_runner.go",
			code: `package cmd

import (
	"os/exec"
)

func RunScript(userInput string) error {
	cmd := exec.Command("sh", "-c", "echo " + userInput)
	return cmd.Run()
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-CMD-01",
				Title:       "Remediate Shell Command Injection",
				Description: "Command injection: shell invocation with unsanitized userInput in sh -c",
				TargetFiles: []string{"cmd/exec_runner.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-CMD-01",
				Category:  repo_intel.CategorySecurity,
				Title:     "Command injection in sh -c",
			},
			expectedStatus: DecisionAbstained,
			expectedIntent: IntentAbstain,
			expectAbstain:  AbstainCommandInjection,
		},

		// Case 6: Abstain on Resource Allocation inside Loop
		{
			caseID:   "SYNTH-LOOP-ABSTAIN",
			category: "RESOURCE_LEAK_LOOP",
			relPath:  "batch/downloader.go",
			code: `package batch

import (
	"net/http"
)

func DownloadAll(urls []string) error {
	for _, u := range urls {
		resp, err := http.Get(u)
		if err != nil {
			return err
		}
		_ = resp.StatusCode
	}
	return nil
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-LOOP-01",
				Title:       "Close Resources in Loop",
				Description: "Resource leak inside for loop",
				TargetFiles: []string{"batch/downloader.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-LOOP-01",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Resource leak inside loop",
			},
			expectedStatus: DecisionAbstained,
			expectedIntent: IntentResourceCleanup,
			expectAbstain:  AbstainLoopResource,
		},

		// Case 7: Abstain on Concurrent Resource Escape
		{
			caseID:   "SYNTH-ESCAPE-ABSTAIN",
			category: "CONCURRENT_ESCAPE",
			relPath:  "async/worker.go",
			code: `package async

import (
	"net/http"
)

func ProcessAsync(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	go func() {
		_ = resp.Body
	}()
	return nil
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-ESCAPE-01",
				Title:       "Close Resource Escaping to Goroutine",
				Description: "Resource leak with concurrent ownership escape",
				TargetFiles: []string{"async/worker.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-ESCAPE-01",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Resource escapes to concurrent goroutine",
			},
			expectedStatus: DecisionAbstained,
			expectedIntent: IntentResourceCleanup,
			expectAbstain:  AbstainConcurrentEscape,
		},

		// Case 8: Abstain on Ambiguous Return Signature
		{
			caseID:   "SYNTH-RETURN-ABSTAIN",
			category: "AMBIGUOUS_RETURN",
			relPath:  "log/audit.go",
			code: `package log

import (
	"os"
)

func RecordEvent(path string) {
	_ = os.Remove(path)
}
`,
			task: &planning.RemediationTask{
				TaskID:      "TASK-RET-01",
				Title:       "Handle Error in Void Function",
				Description: "Unhandled error in function with no error return signature",
				TargetFiles: []string{"log/audit.go"},
			},
			finding: &repo_intel.Finding{
				FindingID: "FIND-RET-01",
				Category:  repo_intel.CategoryMaintainability,
				Title:     "Unhandled error in void return function",
			},
			expectedStatus: DecisionAbstained,
			expectedIntent: IntentHandleError,
			expectAbstain:  AbstainAmbiguousReturn,
		},
	}

	var records []SynthesisTrialRecord

	for _, tc := range testCases {
		t.Run(tc.caseID, func(t *testing.T) {
			caseDir := filepath.Join(tmpDir, tc.caseID)
			targetAbs := filepath.Join(caseDir, filepath.FromSlash(tc.relPath))
			if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
				t.Fatalf("MkdirAll failed: %v", err)
			}
			if err := os.WriteFile(targetAbs, []byte(tc.code), 0644); err != nil {
				t.Fatalf("WriteFile failed: %v", err)
			}

			// Execute synthesis
			decision := synthesizer.Synthesize(caseDir, tc.task, tc.finding)

			record := SynthesisTrialRecord{
				CaseID:         tc.caseID,
				Category:       tc.category,
				TargetFile:     tc.relPath,
				ExpectedIntent: string(tc.expectedIntent),
				ActualStatus:   string(decision.Status),
				DurationMs:     decision.DurationMs,
			}

			if decision.Status != tc.expectedStatus {
				t.Fatalf("[%s] Expected status %s, got %s (Reason: %s)", tc.caseID, tc.expectedStatus, decision.Status, decision.Reason)
			}

			if tc.expectedStatus == DecisionAbstained {
				record.AbstainReason = decision.Reason
				if tc.expectAbstain != "" && decision.Intent != nil && decision.Intent.AbstainCategory != tc.expectAbstain {
					t.Fatalf("[%s] Expected abstain category %s, got %s", tc.caseID, tc.expectAbstain, decision.Intent.AbstainCategory)
				}
				records = append(records, record)
				return
			}

			// Decision is PROPOSED
			if decision.Patch == nil {
				t.Fatalf("[%s] Decision proposed but patch is nil", tc.caseID)
			}
			record.PatchGenerated = true
			record.UnifiedDiff = decision.Patch.UnifiedDiff
			record.DiffHash = decision.Patch.DiffHash

			// Verify that patch actually modified the source and is a real diff
			if decision.Patch.UnifiedDiff == "" || decision.Patch.DiffHash == "" {
				t.Fatalf("[%s] Candidate patch diff or hash is empty", tc.caseID)
			}

			// Apply patch to workspace
			if err := os.WriteFile(targetAbs, []byte(decision.Patch.ModifiedSource), 0644); err != nil {
				t.Fatalf("Failed to apply patch: %v", err)
			}

			// Run Multi-Oracle Verification
			oracleReport := oracles.VerifyAll(caseDir, tc.task, decision.Patch, "", "")
			record.IndependentValidations = make(map[string]bool)
			for oType, oRes := range oracleReport.OracleResults {
				record.IndependentValidations[string(oType)] = oRes.Passed
			}
			record.AllOraclesPassed = oracleReport.AllPassed

			if !oracleReport.AllPassed {
				var failed []string
				for oType, oRes := range oracleReport.OracleResults {
					if !oRes.Passed {
						failed = append(failed, fmt.Sprintf("%s: %s", oType, oRes.FailureReason))
					}
				}
				t.Fatalf("[%s] Oracle verification failed: %s", tc.caseID, strings.Join(failed, "; "))
			}

			records = append(records, record)
		})
	}

	// Compute summary metrics
	report := SynthesisBenchmarkReport{
		Timestamp:  time.Now().UTC(),
		TotalCases: len(records),
		Records:    records,
	}

	for _, r := range records {
		switch r.ActualStatus {
		case string(DecisionProposed):
			report.SynthesizedCount++
			if r.AllOraclesPassed {
				report.VerifiedCount++
			}
		case string(DecisionAbstained):
			report.AbstainedCount++
		case string(DecisionFailed):
			report.FailedCount++
		}
	}

	if report.TotalCases > 0 {
		report.SynthesisRate = float64(report.SynthesizedCount) / float64(report.TotalCases)
		report.AbstentionRate = float64(report.AbstainedCount) / float64(report.TotalCases)
	}
	if report.SynthesizedCount > 0 {
		report.OracleSuccessRate = float64(report.VerifiedCount) / float64(report.SynthesizedCount)
	}

	// Persist to Deliverable 4: `reports/milestone3_7_phase1_5_synthesis_results.json`
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	reportPath := "E:\\SWE\\reports\\milestone3_7_phase1_5_synthesis_results.json"
	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		t.Fatalf("Failed to write synthesis results json: %v", err)
	}
}

// TestAutonomousEndToEndRemediationPipeline tests the closed-loop remediation lifecycle:
// Finding → Synthesis → Approval Token → Execution Boundary Check → Sandboxed Apply → 5 Oracles → Replay Protection.
func TestAutonomousEndToEndRemediationPipeline(t *testing.T) {
	synthesizer := NewAutonomousPatchSynthesizer()
	oracles := NewIndependentOracleSuite(15)

	secretKey := []byte("closed-loop-pipeline-secret-key-32b")
	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "pipeline-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	persistPath := filepath.Join(tmpDir, "consumed_tokens.json")
	registry, err := NewApprovalConsumptionRegistry(persistPath)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}
	authority := NewApprovalAuthority(secretKey, registry)

	targetRel := "client/fetcher.go"
	initialCode := `package client

import (
	"io"
	"net/http"
)

func DoFetch(endpoint string) ([]byte, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
`
	targetAbs := filepath.Join(tmpDir, filepath.FromSlash(targetRel))
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(targetAbs, []byte(initialCode), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	task := &planning.RemediationTask{
		TaskID:                "TASK-E2E-001",
		Title:                 "Close Response Body in DoFetch",
		Description:           "Remediate resource leak: missing defer resp.Body.Close()",
		TargetFiles:           []string{targetRel},
		ExecutionClass:        planning.ClassHumanApprovalRequired,
		HumanApprovalRequired: true,
	}

	finding := &repo_intel.Finding{
		FindingID: "FIND-E2E-001",
		Category:  repo_intel.CategoryMaintainability,
		Title:     "Resource leak on http response",
	}

	// 1. Autonomous Synthesis
	decision := synthesizer.Synthesize(tmpDir, task, finding)
	if decision.Status != DecisionProposed || decision.Patch == nil {
		t.Fatalf("Expected synthesis PROPOSED, got: %s (Reason: %s)", decision.Status, decision.Reason)
	}

	// 2. Prepare Immutable State Hashes
	commitSHA := "92f1c721bb6d7128295a0c63e117d2e8b85bd458"
	workspaceID := "ws-e2e-isolated"
	fileHashes := map[string]string{
		targetRel: computeFileHash(targetAbs),
	}
	patchHash := decision.Patch.DiffHash

	// 3. Issue Cryptographic Approval Token
	token, err := authority.IssueToken(
		"sec-auditor@enterprise.local",
		"Principal Security Architect",
		task.TaskID,
		"TIER_2_CONTROLLED",
		task.ExecutionClass,
		commitSHA,
		fileHashes,
		patchHash,
		"APPLY_PATCH",
		workspaceID,
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// 4. Boundary Eligibility Check (Cryptographic Verification + Single-Use Consumption)
	eligible, _, rejectReason := CheckTaskEligibilitySecure(
		task,
		token,
		authority,
		commitSHA,
		fileHashes,
		patchHash,
		workspaceID,
	)
	if !eligible {
		t.Fatalf("Execution boundary check failed: %s", rejectReason)
	}

	// 5. Sandboxed Apply
	if err := os.WriteFile(targetAbs, []byte(decision.Patch.ModifiedSource), 0644); err != nil {
		t.Fatalf("Failed to apply patch: %v", err)
	}

	// 6. Independent Multi-Oracle Verification
	oracleReport := oracles.VerifyAll(tmpDir, task, decision.Patch, "", "")
	if !oracleReport.AllPassed {
		t.Fatalf("Oracles failed after patch application: %+v", oracleReport)
	}

	// 7. Verify Replay Attack Prevention
	reEligible, _, reReason := CheckTaskEligibilitySecure(
		task,
		token,
		authority,
		commitSHA,
		fileHashes,
		patchHash,
		workspaceID,
	)
	if reEligible {
		t.Fatalf("Security failure: consumed token was accepted for replay!")
	}
	if !strings.Contains(reReason, "FAIL_REPLAY") {
		t.Fatalf("Expected FAIL_REPLAY, got: %s", reReason)
	}
}

func computeFileHash(path string) string {
	data, _ := os.ReadFile(path)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
