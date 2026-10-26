package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const testSnapshotID = "0123456789abcdef0123456789abcdef"

func TestSnapshotStore_RejectsInvalidID(t *testing.T) {
	reader := &fakeSnapshotArchiveReader{}
	store := NewSnapshotStore(reader, t.TempDir())

	if _, _, err := store.ExtractSnapshot(context.Background(), "../not-an-id"); err == nil {
		t.Fatal("expected invalid snapshot ID to be rejected")
	}
	if reader.key != "" {
		t.Fatalf("invalid snapshot ID triggered an object read for %q", reader.key)
	}
}

func TestSnapshotStore_RejectsMissingObject(t *testing.T) {
	reader := &fakeSnapshotArchiveReader{err: errors.New("object not found")}
	store := NewSnapshotStore(reader, t.TempDir())

	if _, _, err := store.ExtractSnapshot(context.Background(), testSnapshotID); err == nil {
		t.Fatal("expected missing snapshot object to fail explicitly")
	}
	if reader.key != "snapshots/"+testSnapshotID+"/source.zip" {
		t.Fatalf("resolved object key = %q, want derived snapshot key", reader.key)
	}
}

func TestSnapshotStore_ExtractsWorkspaceAndCleansIt(t *testing.T) {
	reader := &fakeSnapshotArchiveReader{body: makeSnapshotArchive(t, map[string]string{
		"main.go": "package main",
	})}
	store := NewSnapshotStore(reader, t.TempDir())

	workspace, cleanup, err := store.ExtractSnapshot(context.Background(), testSnapshotID)
	if err != nil {
		t.Fatalf("ExtractSnapshot returned error: %v", err)
	}
	if reader.key != "snapshots/"+testSnapshotID+"/source.zip" {
		t.Fatalf("resolved object key = %q, want derived snapshot key", reader.key)
	}
	content, err := os.ReadFile(filepath.Join(workspace, "main.go"))
	if err != nil {
		t.Fatalf("read extracted source: %v", err)
	}
	if string(content) != "package main" {
		t.Fatalf("extracted source = %q, want %q", content, "package main")
	}

	cleanup()
	if _, err := os.Stat(workspace); !os.IsNotExist(err) {
		t.Fatalf("workspace remains after cleanup; stat error = %v", err)
	}
}

func makeSnapshotArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create ZIP entry %q: %v", name, err)
		}
		if _, err := io.WriteString(file, content); err != nil {
			t.Fatalf("write ZIP entry %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP archive: %v", err)
	}
	return buffer.Bytes()
}

type fakeSnapshotArchiveReader struct {
	body []byte
	key  string
	err  error
}

func (r *fakeSnapshotArchiveReader) GetObject(_ context.Context, key string) (io.ReadCloser, error) {
	r.key = key
	if r.err != nil {
		return nil, r.err
	}
	return io.NopCloser(bytes.NewReader(r.body)), nil
}

type fakeSnapshotWorkspaceProvider struct {
	path      string
	snapshot  string
	calls     int
	cleanedUp bool
	err       error
}

func (p *fakeSnapshotWorkspaceProvider) ExtractSnapshot(_ context.Context, snapshotID string) (string, func(), error) {
	p.calls++
	p.snapshot = snapshotID
	if p.err != nil {
		return "", nil, p.err
	}
	return p.path, func() { p.cleanedUp = true }, nil
}
