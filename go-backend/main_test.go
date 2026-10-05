// go-backend · main_test.go
//
// Unit tests for the utility functions — no HTTP server or engine required.
// Run with: go test ./... -v

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ── extractZip tests ──────────────────────────────────────────────────────────

// makeTestZip creates a small in-memory zip and writes it to `path`.
func makeTestZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range entries {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatalf("create entry %q: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write entry %q: %v", name, err)
		}
	}
}

// TestExtractZip_HappyPath verifies that a well-formed zip is extracted correctly.
func TestExtractZip_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "test.zip")

	makeTestZip(t, zipPath, map[string]string{
		"hello.py":      "print('hello')\n",
		"sub/world.go":  "package main\n",
	})

	destDir := filepath.Join(tmp, "out")
	if err := extractZip(zipPath, destDir); err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}

	// Verify files exist with expected content
	checkFile := func(rel, want string) {
		t.Helper()
		got, err := os.ReadFile(filepath.Join(destDir, rel))
		if err != nil {
			t.Errorf("read %q: %v", rel, err)
			return
		}
		if string(got) != want {
			t.Errorf("%q: got %q, want %q", rel, got, want)
		}
	}

	checkFile("hello.py", "print('hello')\n")
	checkFile("sub/world.go", "package main\n")
}

// TestExtractZip_ZipSlip ensures a path-traversal zip is rejected.
func TestExtractZip_ZipSlip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "evil.zip")

	// Manually craft a zip with a path traversal entry
	f, _ := os.Create(zipPath)
	w := zip.NewWriter(f)
	// "../escape.txt" would land outside the dest dir
	entry, _ := w.Create("../../escape.txt")
	_, _ = entry.Write([]byte("pwned"))
	w.Close()
	f.Close()

	destDir := filepath.Join(tmp, "out")
	err := extractZip(zipPath, destDir)
	if err == nil {
		t.Fatal("expected error for zip slip, got nil")
	}
	t.Logf("correctly rejected: %v", err)
}

// TestExtractZip_CorruptZip verifies that a corrupted zip returns an error.
func TestExtractZip_CorruptZip(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "corrupt.zip")
	// Write garbage bytes
	if err := os.WriteFile(zipPath, []byte("this is not a zip file"), 0o640); err != nil {
		t.Fatal(err)
	}

	err := extractZip(zipPath, filepath.Join(tmp, "out"))
	if err == nil {
		t.Fatal("expected error for corrupt zip, got nil")
	}
	t.Logf("correctly rejected: %v", err)
}

// ── readReport tests ──────────────────────────────────────────────────────────

// TestReadReport_Valid verifies that a well-formed JSON report is parsed.
func TestReadReport_Valid(t *testing.T) {
	tmp := t.TempDir()
	reportPath := filepath.Join(tmp, "report.json")

	json := `{
		"engine_version": "0.1.0",
		"generated_at": "2026-08-13T00:00:00Z",
		"scanned_directory": "/tmp/code",
		"files_analyzed": 2,
		"total_hotspots": 5,
		"global_entropy": 123.45,
		"file_reports": []
	}`

	if err := os.WriteFile(reportPath, []byte(json), 0o640); err != nil {
		t.Fatal(err)
	}

	report, err := readReport(reportPath)
	if err != nil {
		t.Fatalf("readReport: %v", err)
	}

	if report.FilesAnalyzed != 2 {
		t.Errorf("FilesAnalyzed: got %d, want 2", report.FilesAnalyzed)
	}
	if report.GlobalEntropy != 123.45 {
		t.Errorf("GlobalEntropy: got %f, want 123.45", report.GlobalEntropy)
	}
	if report.EngineVersion != "0.1.0" {
		t.Errorf("EngineVersion: got %q, want 0.1.0", report.EngineVersion)
	}
}

// TestReadReport_MissingFile verifies error on absent report.
func TestReadReport_MissingFile(t *testing.T) {
	_, err := readReport("/nonexistent/path/report.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	t.Logf("correctly rejected: %v", err)
}

// TestReadReport_InvalidJSON verifies error on malformed JSON.
func TestReadReport_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.json")
	if err := os.WriteFile(path, []byte("{broken json"), 0o640); err != nil {
		t.Fatal(err)
	}

	_, err := readReport(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	t.Logf("correctly rejected: %v", err)
}

// ── envOr / envOrInt tests ────────────────────────────────────────────────────

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_KEY", "hello")
	if got := envOr("TEST_KEY", "fallback"); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
	if got := envOr("UNSET_KEY_XYZ", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

func TestEnvOrInt(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if got := envOrInt("TEST_INT", 0); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
	if got := envOrInt("UNSET_INT_XYZ", 99); got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestResolveEnginePath(t *testing.T) {
	// Custom env override
	t.Setenv("ENGINE_PATH", "/custom/path/engine")
	if got := resolveEnginePath(); got != "/custom/path/engine" {
		t.Errorf("got %q, want %q", got, "/custom/path/engine")
	}

	// Without env override, should return a sensible path without crashing
	t.Setenv("ENGINE_PATH", "")
	got := resolveEnginePath()
	if got == "" {
		t.Error("expected non-empty engine path")
	}
}

func TestJSONStore_SaveAndGetAll(t *testing.T) {
	tmp := t.TempDir()
	jsonPath := filepath.Join(tmp, "deliverables.json")

	store, err := NewJSONStore(jsonPath)
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}

	items, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll empty: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}

	d := Deliverable{
		Title:    "Test Deck",
		Version:  "1.0",
		Filename: "deck.pptx",
		FileURL:  "/uploads/deck.pptx",
	}

	saved, err := store.Save(d)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID != 1 {
		t.Errorf("expected ID 1, got %d", saved.ID)
	}

	// Reload from disk to verify persistence
	reloadedStore, err := NewJSONStore(jsonPath)
	if err != nil {
		t.Fatalf("reloadedStore: %v", err)
	}

	items, err = reloadedStore.GetAll()
	if err != nil {
		t.Fatalf("reloadedStore.GetAll: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Test Deck" {
		t.Errorf("expected title 'Test Deck', got %q", items[0].Title)
	}
}

func TestAllowedFileExtensions(t *testing.T) {
	valid := []string{".ppt", ".pptx", ".pdf", ".zip", ".txt", ".md"}
	for _, ext := range valid {
		if !allowedExtensions[ext] {
			t.Errorf("extension %q should be allowed", ext)
		}
	}

	invalid := []string{".exe", ".sh", ".bat", ".py", ".js", ".html"}
	for _, ext := range invalid {
		if allowedExtensions[ext] {
			t.Errorf("extension %q must NOT be allowed", ext)
		}
	}
}

// ── End-to-End HTTP Integration Tests ────────────────────────────────────────

func TestIntegration_HealthAndStatus(t *testing.T) {
	engine := setupEngine("./thermodynamic-ast-engine", 5*time.Second, 10<<20, nil)

	for _, path := range []string{"/health", "/api/health", "/status", "/api/status"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("%s: expected 200 OK, got %d", path, rec.Code)
		}

		var body map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: invalid json: %v", path, err)
		}
		if body["status"] != "ok" {
			t.Errorf("%s: expected status 'ok', got %v", path, body["status"])
		}
	}
}

func TestIntegration_DeliverablesFlow(t *testing.T) {
	t.Setenv("ANALYSIS_API_TOKEN", "test-token")
	tmp := t.TempDir()
	store, _ := NewJSONStore(filepath.Join(tmp, "deliv.json"))
	engine := setupEngine("./thermodynamic-ast-engine", 5*time.Second, 10<<20, store)

	// 1. Upload a deliverable
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "Project Milestone 1")
	_ = writer.WriteField("version", "1.0.0")
	part, _ := writer.CreateFormFile("file", "presentation.pdf")
	_, _ = part.Write([]byte("%PDF-1.4 mock content"))
	_ = writer.Close()

	uploadReq, _ := http.NewRequest(http.MethodPost, "/api/upload", &body)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("X-API-Token", "test-token")
	uploadRec := httptest.NewRecorder()
	engine.ServeHTTP(uploadRec, uploadReq)

	if uploadRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", uploadRec.Code, uploadRec.Body.String())
	}

	// 2. Fetch deliverables
	getReq, _ := http.NewRequest(http.MethodGet, "/api/deliverables", nil)
	getReq.Header.Set("X-API-Token", "test-token")
	getRec := httptest.NewRecorder()
	engine.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getRec.Code)
	}

	var items []Deliverable
	if err := json.Unmarshal(getRec.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal deliverables: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 deliverable, got %d", len(items))
	}
	if items[0].Title != "Project Milestone 1" {
		t.Errorf("expected title 'Project Milestone 1', got %q", items[0].Title)
	}
}

func TestIntegration_AnalyzeZipFlow(t *testing.T) {
	t.Setenv("ANALYSIS_API_TOKEN", "test-token")
	enginePath := resolveEnginePath()
	if _, err := os.Stat(enginePath); os.IsNotExist(err) {
		t.Skip("skipping analyze test: thermodynamic-ast-engine binary not found")
	}

	engine := setupEngine(enginePath, 30*time.Second, 50<<20, nil)

	// Create sample zip with nested loops in Python
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "sample.zip")
	makeTestZip(t, zipPath, map[string]string{
		"pipeline.py": "def run_batches(batches):\n    for b in batches:\n        for item in b:\n            print(item)\n",
	})

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("read sample zip: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("zipfile", "code.zip")
	_, _ = part.Write(zipData)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/analyze", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Token", "test-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp AnalyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal analyze response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Report == nil {
		t.Fatal("expected report in response, got nil")
	}
	if resp.Report.FilesAnalyzed == 0 {
		t.Errorf("expected files_analyzed > 0, got %d", resp.Report.FilesAnalyzed)
	}
	if resp.Report.TotalHotspots == 0 {
		t.Errorf("expected hotspots detected for nested loop, got 0")
	}
}

func TestIntegration_ScanRepoValidation(t *testing.T) {
	t.Setenv("ANALYSIS_API_TOKEN", "test-token")
	engine := setupEngine("./thermodynamic-ast-engine", 5*time.Second, 10<<20, nil)

	// Test invalid / malicious repo URLs
	badURLs := []string{
		"-v",
		"file:///etc/passwd",
		"ftp://evil.com/repo",
		"https://github.com/repo; rm -rf /",
	}

	for _, bad := range badURLs {
		jsonBody := []byte(fmt.Sprintf(`{"repo_url": %q}`, bad))
		req, _ := http.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Token", "test-token")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for malicious url %q, got %d", bad, rec.Code)
		}
	}
}
