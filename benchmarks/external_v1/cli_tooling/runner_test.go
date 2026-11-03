package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlags(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg, err := ParseFlags([]string{"-cmd", "scan", "-path", "/tmp", "-v"}, buf)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if cfg.Command != "scan" {
		t.Errorf("expected cmd=scan, got %s", cfg.Command)
	}
	if cfg.Path != "/tmp" {
		t.Errorf("expected path=/tmp, got %s", cfg.Path)
	}
	if !cfg.Verbose {
		t.Errorf("expected verbose=true")
	}
}

func TestSafeReadFileFromRoot(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.txt")
	os.WriteFile(filePath, []byte("hello safe fs"), 0644)

	// Valid read
	data, err := SafeReadFileFromRoot(tmpDir, "sample.txt")
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "hello safe fs" {
		t.Errorf("expected content match")
	}

	// Traversal attempt
	_, err = SafeReadFileFromRoot(tmpDir, "../escape.txt")
	if err == nil {
		t.Fatalf("expected path traversal detection error, got nil")
	}
}

func TestExecuteDirect(t *testing.T) {
	ctx := context.Background()
	out, err := ExecuteDirect(ctx, "go", "version")
	if err != nil {
		t.Fatalf("failed to execute go version: %v", err)
	}
	if len(out) == 0 {
		t.Errorf("expected output from go version")
	}
}

