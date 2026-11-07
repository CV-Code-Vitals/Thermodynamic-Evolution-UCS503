package execution

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

func TestHardenedOraclesRejectFlawedNAT03Patch(t *testing.T) {
	suite := NewIndependentOracleSuite(15)

	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "oracle-nat03-*")
	if err != nil {
		t.Fatalf("Failed to create temp fixture directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Minimal reproducing client.go with DialContext returning *http.Response
	clientCode := `package client

import (
	"net/http"
)

type Dialer struct{}

func (d *Dialer) DialContext(urlStr string) (*http.Response, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}
`
	targetFile := "client.go"
	targetPath := filepath.Join(tmpDir, targetFile)
	if err := os.WriteFile(targetPath, []byte(clientCode), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	task := &planning.RemediationTask{
		TaskID:      "TASK-NAT-03",
		Title:       "Remediate Resource Leak in client.go",
		Description: "HTTP response body unclosed along handshake failure redirect path",
		TargetFiles: []string{targetFile},
	}

	patch := &CandidatePatch{
		PatchID:    "PATCH-NAT-03-FLAWED",
		TaskID:     task.TaskID,
		TargetFile: targetFile,
		UnifiedDiff: `--- a/client.go
+++ b/client.go
@@ -14,2 +14,3 @@
 	if err != nil {
 		return nil, err
 	}
+	defer resp.Body.Close()
 	return resp, nil
`,
	}

	report := suite.VerifyAll(tmpDir, task, patch, "", "")

	if report.AllPassed {
		t.Fatalf("Hardened oracle suite unexpectedly PASSED flawed NAT-03 patch!")
	}

	// Verify Oracle 5: Negative Properties caught the escaping resource violation
	negRes := report.OracleResults[OracleNegativeProperties]
	if negRes == nil || negRes.Passed {
		t.Fatalf("OracleNegativeProperties did not fail on escaping resource cleanup!")
	}
	if !strings.Contains(negRes.FailureReason, "escaping resource") {
		t.Fatalf("Unexpected failure reason from OracleNegativeProperties: %s", negRes.FailureReason)
	}
	t.Logf("CONFIRMED: OracleNegativeProperties correctly rejected flawed NAT-03 patch: %s", negRes.FailureReason)

	// Verify Oracle 3: Defect Reproducer caught the escaping return contract violation
	reproRes := report.OracleResults[OracleDefectReproducer]
	if reproRes == nil || reproRes.Passed {
		t.Fatalf("OracleDefectReproducer did not fail on escaping resource cleanup!")
	}
	if !strings.Contains(reproRes.FailureReason, "returns resource to caller") {
		t.Fatalf("Unexpected failure reason from OracleDefectReproducer: %s", reproRes.FailureReason)
	}
	t.Logf("CONFIRMED: OracleDefectReproducer correctly rejected flawed NAT-03 patch: %s", reproRes.FailureReason)
}

func TestHardenedOraclesRejectPreErrorCheckCleanup(t *testing.T) {
	suite := NewIndependentOracleSuite(15)

	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "oracle-pre-check-*")
	if err != nil {
		t.Fatalf("Failed to create temp fixture directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flawedCode := `package storage

import "os"

func ReadFile(path string) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}
	return nil
}
`
	targetFile := "reader.go"
	targetPath := filepath.Join(tmpDir, targetFile)
	if err := os.WriteFile(targetPath, []byte(flawedCode), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	task := &planning.RemediationTask{
		TaskID:      "TASK-PRE-ERR",
		Title:       "Close Resource Leak",
		Description: "Deferred close placed before error check",
		TargetFiles: []string{targetFile},
	}

	patch := &CandidatePatch{
		PatchID:    "PATCH-PRE-ERR",
		TaskID:     task.TaskID,
		TargetFile: targetFile,
		UnifiedDiff: `--- a/reader.go
+++ b/reader.go
@@ -6,1 +6,2 @@
 	f, err := os.Open(path)
+	defer f.Close()
 	if err != nil {
`,
	}

	report := suite.VerifyAll(tmpDir, task, patch, "", "")

	if report.AllPassed {
		t.Fatalf("Hardened oracle suite unexpectedly PASSED pre-error-check cleanup placement!")
	}

	negRes := report.OracleResults[OracleNegativeProperties]
	if negRes == nil || negRes.Passed {
		t.Fatalf("OracleNegativeProperties did not fail on pre-error-check cleanup placement!")
	}
	if !strings.Contains(negRes.FailureReason, "placed prior to error check") {
		t.Fatalf("Unexpected failure reason from OracleNegativeProperties: %s", negRes.FailureReason)
	}
	t.Logf("CONFIRMED: OracleNegativeProperties correctly rejected pre-error-check placement: %s", negRes.FailureReason)
}

func TestHardenedOraclesRejectMultipleCloseCalls(t *testing.T) {
	suite := NewIndependentOracleSuite(15)

	tmpDir, err := os.MkdirTemp("E:\\SWE\\.tmp", "oracle-multi-close-*")
	if err != nil {
		t.Fatalf("Failed to create temp fixture directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	flawedCode := `package storage

import "os"

func ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	f.Close()
	return nil
}
`
	targetFile := "reader.go"
	targetPath := filepath.Join(tmpDir, targetFile)
	if err := os.WriteFile(targetPath, []byte(flawedCode), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	task := &planning.RemediationTask{
		TaskID:      "TASK-MULTI-CLOSE",
		Title:       "Close Resource Leak",
		Description: "Redundant double close calls",
		TargetFiles: []string{targetFile},
	}

	patch := &CandidatePatch{
		PatchID:    "PATCH-MULTI-CLOSE",
		TaskID:     task.TaskID,
		TargetFile: targetFile,
		UnifiedDiff: `--- a/reader.go
+++ b/reader.go
@@ -9,1 +9,2 @@
 	defer f.Close()
+	f.Close()
`,
	}

	report := suite.VerifyAll(tmpDir, task, patch, "", "")

	if report.AllPassed {
		t.Fatalf("Hardened oracle suite unexpectedly PASSED multiple close calls!")
	}

	negRes := report.OracleResults[OracleNegativeProperties]
	if negRes == nil || negRes.Passed {
		t.Fatalf("OracleNegativeProperties did not fail on multiple close calls!")
	}
	if !strings.Contains(negRes.FailureReason, "multiple Close() calls") {
		t.Fatalf("Unexpected failure reason from OracleNegativeProperties: %s", negRes.FailureReason)
	}
	t.Logf("CONFIRMED: OracleNegativeProperties correctly rejected redundant close: %s", negRes.FailureReason)
}

