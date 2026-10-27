package agent

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles(t *testing.T) {
	repo := t.TempDir()
	files := map[string]string{
		"README.txt":                    "sample project",
		"src/main.go":                   "package src\n",
		"nested/.git/config":            "private metadata",
		".env":                          "SECRET=value",
		"src/.hidden.go":                "package src",
		"src/private.pem":               "private key",
		"src/assets/image.png":          "binary",
		"node_modules/example/index.go": "package example",
	}
	for name, contents := range files {
		path := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	first, err := CreateWorkspaceArchive(repo, 1<<20)
	if err != nil {
		t.Fatalf("CreateWorkspaceArchive() error = %v", err)
	}
	second, err := CreateWorkspaceArchive(repo, 1<<20)
	if err != nil {
		t.Fatalf("second CreateWorkspaceArchive() error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("workspace archive is not deterministic")
	}

	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil {
		t.Fatalf("open workspace archive: %v", err)
	}
	archiveFiles := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		if strings.Contains(file.Name, ".git/") || strings.HasSuffix(file.Name, "/.git") {
			t.Fatalf("archive contains Git metadata: %q", file.Name)
		}
		stream, err := file.Open()
		if err != nil {
			t.Fatalf("open %q: %v", file.Name, err)
		}
		contents, readErr := io.ReadAll(stream)
		closeErr := stream.Close()
		if readErr != nil {
			t.Fatalf("read %q: %v", file.Name, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close %q: %v", file.Name, closeErr)
		}
		archiveFiles[file.Name] = string(contents)
	}
	if got := archiveFiles["README.txt"]; got != files["README.txt"] {
		t.Fatalf("README.txt = %q, want %q", got, files["README.txt"])
	}
	if got := archiveFiles["src/main.go"]; got != files["src/main.go"] {
		t.Fatalf("src/main.go = %q, want %q", got, files["src/main.go"])
	}
	if len(archiveFiles) != 2 {
		t.Fatalf("archive contains %d files, want only the two evaluator-supported files", len(archiveFiles))
	}
}

func TestCreateWorkspaceArchive_RejectsOversizeFile(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "large.go")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, (32<<20)+1); err != nil {
		t.Fatal(err)
	}

	if _, err := CreateWorkspaceArchive(repo, 50<<20); err == nil {
		t.Fatal("CreateWorkspaceArchive() accepted a file larger than 32 MiB")
	}
}

func TestCreateWorkspaceArchive_RejectsCompressedLimit(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "small.txt"), []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := CreateWorkspaceArchive(repo, 1); err == nil {
		t.Fatal("CreateWorkspaceArchive() accepted an archive larger than its limit")
	}
}

func TestCreateWorkspaceArchive_RejectsSymlink(t *testing.T) {
	repo := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte("package outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(repo, "linked.go")); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}

	if _, err := CreateWorkspaceArchive(repo, 1<<20); err == nil {
		t.Fatal("CreateWorkspaceArchive() accepted a symlink")
	}
}

func TestDefaultWorkspaceArchiveLimits(t *testing.T) {
	limits := DefaultWorkspaceArchiveLimits()
	if limits.MaxArchiveBytes != 50<<20 {
		t.Errorf("MaxArchiveBytes = %d, want %d", limits.MaxArchiveBytes, 50<<20)
	}
	if limits.MaxExpandedBytes != 256<<20 {
		t.Errorf("MaxExpandedBytes = %d, want %d", limits.MaxExpandedBytes, 256<<20)
	}
	if limits.MaxEntries != 10_000 {
		t.Errorf("MaxEntries = %d, want %d", limits.MaxEntries, 10_000)
	}
	if limits.MaxFileBytes != 32<<20 {
		t.Errorf("MaxFileBytes = %d, want %d", limits.MaxFileBytes, 32<<20)
	}
}

func TestExtractWorkspaceArchive_RejectsTraversal(t *testing.T) {
	archive := makeWorkspaceTestArchive(t, []workspaceTestEntry{{name: "../outside.txt", contents: "escaped"}})
	root := t.TempDir()
	destination := filepath.Join(root, "workspace")

	err := ExtractWorkspaceArchive(archive, destination, DefaultWorkspaceArchiveLimits())
	if err == nil {
		t.Fatal("ExtractWorkspaceArchive() accepted a traversal path")
	}
	if _, statErr := os.Stat(filepath.Join(root, "outside.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("traversal target exists outside workspace, stat error = %v", statErr)
	}
}

func TestExtractWorkspaceArchive_RejectsSymlinkEntries(t *testing.T) {
	archive := makeWorkspaceTestArchive(t, []workspaceTestEntry{{
		name:     "link",
		contents: "target",
		mode:     os.ModeSymlink | 0o777,
	}})
	destination := filepath.Join(t.TempDir(), "workspace")

	if err := ExtractWorkspaceArchive(archive, destination, DefaultWorkspaceArchiveLimits()); err == nil {
		t.Fatal("ExtractWorkspaceArchive() accepted a symlink entry")
	}
}

func TestExtractWorkspaceArchive_AllowsFileLargerThanLegacyLimit(t *testing.T) {
	archive := makeWorkspaceTestArchive(t, []workspaceTestEntry{{
		name:     "data.npz",
		contents: strings.Repeat("x", 17<<20),
	}})
	destination := filepath.Join(t.TempDir(), "workspace")

	if err := ExtractWorkspaceArchive(archive, destination, DefaultWorkspaceArchiveLimits()); err != nil {
		t.Fatalf("ExtractWorkspaceArchive() rejected a file within the 32 MiB per-file limit: %v", err)
	}
	info, err := os.Stat(filepath.Join(destination, "data.npz"))
	if err != nil {
		t.Fatalf("Stat(extracted file): %v", err)
	}
	if info.Size() != 17<<20 {
		t.Fatalf("extracted file size = %d, want %d", info.Size(), 17<<20)
	}
}

func TestExtractWorkspaceArchive_EnforcesExpandedLimits(t *testing.T) {
	tests := []struct {
		name    string
		entries []workspaceTestEntry
		limits  WorkspaceArchiveLimits
	}{
		{
			name:    "total expanded bytes",
			entries: []workspaceTestEntry{{name: "file.txt", contents: "12345"}},
			limits:  WorkspaceArchiveLimits{MaxArchiveBytes: 1 << 20, MaxExpandedBytes: 4, MaxEntries: 10, MaxFileBytes: 10},
		},
		{
			name:    "entry count",
			entries: []workspaceTestEntry{{name: "one.txt", contents: "1"}, {name: "two.txt", contents: "2"}},
			limits:  WorkspaceArchiveLimits{MaxArchiveBytes: 1 << 20, MaxExpandedBytes: 10, MaxEntries: 1, MaxFileBytes: 10},
		},
		{
			name:    "per-file bytes",
			entries: []workspaceTestEntry{{name: "file.txt", contents: "12345"}},
			limits:  WorkspaceArchiveLimits{MaxArchiveBytes: 1 << 20, MaxExpandedBytes: 10, MaxEntries: 10, MaxFileBytes: 4},
		},
		{
			name:    "compressed archive bytes",
			entries: []workspaceTestEntry{{name: "file.txt", contents: "12345"}},
			limits:  WorkspaceArchiveLimits{MaxArchiveBytes: 1, MaxExpandedBytes: 10, MaxEntries: 10, MaxFileBytes: 10},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			archive := makeWorkspaceTestArchive(t, test.entries)
			destination := filepath.Join(t.TempDir(), "workspace")
			if err := ExtractWorkspaceArchive(archive, destination, test.limits); err == nil {
				t.Fatal("ExtractWorkspaceArchive() accepted an archive exceeding configured limits")
			}
		})
	}
}

type workspaceTestEntry struct {
	name     string
	contents string
	mode     os.FileMode
}

func makeWorkspaceTestArchive(t *testing.T, entries []workspaceTestEntry) []byte {
	t.Helper()
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("create archive entry %q: %v", entry.name, err)
		}
		if _, err := file.Write([]byte(entry.contents)); err != nil {
			t.Fatalf("write archive entry %q: %v", entry.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close archive writer: %v", err)
	}
	return archive.Bytes()
}
