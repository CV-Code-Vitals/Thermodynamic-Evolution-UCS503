package agent

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxWorkspaceArchiveBytes   int64 = 50 << 20
	maxWorkspaceExpandedBytes  int64 = 256 << 20
	maxWorkspaceArchiveEntries       = 10_000
	maxWorkspaceFileBytes      int64 = 32 << 20
)

var errWorkspaceArchiveTooLarge = errors.New("workspace archive exceeds size limit")

type WorkspaceArchiveLimits struct {
	MaxArchiveBytes  int64
	MaxExpandedBytes int64
	MaxEntries       int
	MaxFileBytes     int64
}

func DefaultWorkspaceArchiveLimits() WorkspaceArchiveLimits {
	return WorkspaceArchiveLimits{
		MaxArchiveBytes:  maxWorkspaceArchiveBytes,
		MaxExpandedBytes: maxWorkspaceExpandedBytes,
		MaxEntries:       maxWorkspaceArchiveEntries,
		MaxFileBytes:     maxWorkspaceFileBytes,
	}
}

func CreateWorkspaceArchive(repoPath string, maxArchiveBytes int64) ([]byte, error) {
	if maxArchiveBytes <= 0 {
		return nil, errors.New("workspace archive size limit must be positive")
	}
	if maxArchiveBytes > maxWorkspaceArchiveBytes {
		maxArchiveBytes = maxWorkspaceArchiveBytes
	}

	repoRoot, err := filepath.EvalSymlinks(repoPath)
	if err != nil {
		return nil, fmt.Errorf("resolve repository path: %w", err)
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("make repository path absolute: %w", err)
	}
	rootInfo, err := os.Stat(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("stat repository path: %w", err)
	}
	if !rootInfo.IsDir() {
		return nil, errors.New("repository path is not a directory")
	}

	output := &limitedWorkspaceBuffer{maxBytes: maxArchiveBytes}
	archive := zip.NewWriter(output)
	var expandedBytes int64
	entryCount := 0
	walkErr := filepath.Walk(repoRoot, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath == repoRoot {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace contains symlink %q", filePath)
		}
		if info.IsDir() {
			if shouldSkipProjectDirectory(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("workspace contains unsupported file type at %q", filePath)
		}
		if !shouldCopyProjectFile(info) {
			return nil
		}
		if info.Size() > maxWorkspaceFileBytes {
			return fmt.Errorf("workspace file %q exceeds %d-byte limit", filePath, maxWorkspaceFileBytes)
		}
		if expandedBytes > maxWorkspaceExpandedBytes-info.Size() {
			return fmt.Errorf("workspace exceeds %d-byte expanded limit", maxWorkspaceExpandedBytes)
		}
		entryCount++
		if entryCount > maxWorkspaceArchiveEntries {
			return fmt.Errorf("workspace contains more than %d files", maxWorkspaceArchiveEntries)
		}

		relativePath, err := filepath.Rel(repoRoot, filePath)
		if err != nil {
			return fmt.Errorf("resolve workspace file path: %w", err)
		}
		if !filepath.IsLocal(relativePath) {
			return fmt.Errorf("workspace file path escapes repository: %q", relativePath)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create archive header for %q: %w", relativePath, err)
		}
		header.Name = filepath.ToSlash(relativePath)
		header.Method = zip.Deflate
		header.Modified = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)
		header.SetMode(0o640)
		entry, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create archive entry for %q: %w", relativePath, err)
		}

		source, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open workspace file %q: %w", relativePath, err)
		}
		written, copyErr := io.Copy(entry, io.LimitReader(source, maxWorkspaceFileBytes+1))
		closeErr := source.Close()
		if copyErr != nil {
			return errors.Join(fmt.Errorf("archive workspace file %q: %w", relativePath, copyErr), closeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close workspace file %q: %w", relativePath, closeErr)
		}
		if written > maxWorkspaceFileBytes {
			return fmt.Errorf("workspace file %q exceeds %d-byte limit", relativePath, maxWorkspaceFileBytes)
		}
		if expandedBytes > maxWorkspaceExpandedBytes-written {
			return fmt.Errorf("workspace exceeds %d-byte expanded limit", maxWorkspaceExpandedBytes)
		}
		expandedBytes += written
		return nil
	})
	closeErr := archive.Close()
	if walkErr != nil {
		return nil, fmt.Errorf("create workspace archive: %w", walkErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close workspace archive: %w", closeErr)
	}
	return append([]byte(nil), output.buffer.Bytes()...), nil
}

func ExtractWorkspaceArchive(archiveBytes []byte, destination string, limits WorkspaceArchiveLimits) error {
	limits, err := clampWorkspaceArchiveLimits(limits)
	if err != nil {
		return err
	}
	if int64(len(archiveBytes)) > limits.MaxArchiveBytes {
		return fmt.Errorf("workspace archive exceeds %d-byte compressed limit", limits.MaxArchiveBytes)
	}
	if strings.TrimSpace(destination) == "" {
		return errors.New("workspace destination is required")
	}
	if err := os.MkdirAll(destination, 0o750); err != nil {
		return fmt.Errorf("create workspace destination: %w", err)
	}
	destination, err = filepath.EvalSymlinks(destination)
	if err != nil {
		return fmt.Errorf("resolve workspace destination: %w", err)
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("make workspace destination absolute: %w", err)
	}
	existing, err := os.ReadDir(destination)
	if err != nil {
		return fmt.Errorf("read workspace destination: %w", err)
	}
	if len(existing) != 0 {
		return errors.New("workspace destination must be empty")
	}

	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return fmt.Errorf("open workspace archive: %w", err)
	}
	if len(reader.File) > limits.MaxEntries {
		return fmt.Errorf("workspace archive contains more than %d entries", limits.MaxEntries)
	}

	seen := make(map[string]struct{}, len(reader.File))
	var expandedBytes int64
	for _, file := range reader.File {
		name, isDir, err := validateWorkspaceArchiveEntry(file)
		if err != nil {
			return err
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("workspace archive contains duplicate path %q", name)
		}
		seen[key] = struct{}{}

		targetPath := filepath.Join(destination, filepath.FromSlash(name))
		relativePath, err := filepath.Rel(destination, targetPath)
		if err != nil || !filepath.IsLocal(relativePath) {
			return fmt.Errorf("workspace archive path escapes destination: %q", name)
		}
		if isDir {
			if file.UncompressedSize64 != 0 {
				return fmt.Errorf("workspace directory entry %q contains data", name)
			}
			if err := os.MkdirAll(targetPath, 0o750); err != nil {
				return fmt.Errorf("create workspace directory %q: %w", name, err)
			}
			continue
		}
		if file.UncompressedSize64 > uint64(limits.MaxFileBytes) {
			return fmt.Errorf("workspace file %q exceeds %d-byte limit", name, limits.MaxFileBytes)
		}
		if file.UncompressedSize64 > uint64(limits.MaxExpandedBytes-expandedBytes) {
			return fmt.Errorf("workspace archive exceeds %d-byte expanded limit", limits.MaxExpandedBytes)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o750); err != nil {
			return fmt.Errorf("create workspace parent for %q: %w", name, err)
		}
		written, err := extractWorkspaceFile(file, targetPath, min(limits.MaxFileBytes, limits.MaxExpandedBytes-expandedBytes))
		if err != nil {
			return fmt.Errorf("extract workspace file %q: %w", name, err)
		}
		if written > limits.MaxExpandedBytes-expandedBytes {
			return fmt.Errorf("workspace archive exceeds %d-byte expanded limit", limits.MaxExpandedBytes)
		}
		expandedBytes += written
	}
	return nil
}

func clampWorkspaceArchiveLimits(limits WorkspaceArchiveLimits) (WorkspaceArchiveLimits, error) {
	if limits.MaxArchiveBytes <= 0 || limits.MaxExpandedBytes <= 0 || limits.MaxEntries <= 0 || limits.MaxFileBytes <= 0 {
		return WorkspaceArchiveLimits{}, errors.New("all workspace archive limits must be positive")
	}
	defaults := DefaultWorkspaceArchiveLimits()
	if limits.MaxArchiveBytes > defaults.MaxArchiveBytes {
		limits.MaxArchiveBytes = defaults.MaxArchiveBytes
	}
	if limits.MaxExpandedBytes > defaults.MaxExpandedBytes {
		limits.MaxExpandedBytes = defaults.MaxExpandedBytes
	}
	if limits.MaxEntries > defaults.MaxEntries {
		limits.MaxEntries = defaults.MaxEntries
	}
	if limits.MaxFileBytes > defaults.MaxFileBytes {
		limits.MaxFileBytes = defaults.MaxFileBytes
	}
	return limits, nil
}

func validateWorkspaceArchiveEntry(file *zip.File) (string, bool, error) {
	if file.Mode()&os.ModeSymlink != 0 {
		return "", false, fmt.Errorf("workspace archive contains symlink %q", file.Name)
	}
	isDir := file.FileInfo().IsDir()
	name := strings.TrimSuffix(file.Name, "/")
	if name == "" || strings.ContainsRune(name, 0) || strings.Contains(name, `\`) ||
		!fs.ValidPath(name) || !filepath.IsLocal(filepath.FromSlash(name)) {
		return "", false, fmt.Errorf("workspace archive contains unsafe path %q", file.Name)
	}
	if !isDir && !file.Mode().IsRegular() {
		return "", false, fmt.Errorf("workspace archive contains unsupported entry type %q", file.Name)
	}
	return name, isDir, nil
}

func extractWorkspaceFile(file *zip.File, targetPath string, maxBytes int64) (int64, error) {
	source, err := file.Open()
	if err != nil {
		return 0, fmt.Errorf("open archive entry: %w", err)
	}

	output, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		closeErr := source.Close()
		return 0, errors.Join(fmt.Errorf("create destination file: %w", err), closeErr)
	}
	written, copyErr := io.Copy(output, io.LimitReader(source, maxBytes+1))
	sourceCloseErr := source.Close()
	closeErr := output.Close()
	if copyErr != nil {
		return written, errors.Join(copyErr, sourceCloseErr, closeErr, removePartialWorkspaceFile(targetPath))
	}
	if sourceCloseErr != nil || closeErr != nil {
		return written, errors.Join(sourceCloseErr, closeErr, removePartialWorkspaceFile(targetPath))
	}
	if written > maxBytes {
		return written, errors.Join(fmt.Errorf("entry exceeds %d-byte limit", maxBytes), removePartialWorkspaceFile(targetPath))
	}
	return written, nil
}

func removePartialWorkspaceFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove partial workspace file: %w", err)
	}
	return nil
}

type limitedWorkspaceBuffer struct {
	buffer   bytes.Buffer
	maxBytes int64
}

func (w *limitedWorkspaceBuffer) Write(data []byte) (int, error) {
	remaining := w.maxBytes - int64(w.buffer.Len())
	if remaining <= 0 {
		return 0, errWorkspaceArchiveTooLarge
	}
	if int64(len(data)) > remaining {
		written, _ := w.buffer.Write(data[:int(remaining)])
		return written, errWorkspaceArchiveTooLarge
	}
	return w.buffer.Write(data)
}

func shouldSkipProjectDirectory(name string) bool {
	return strings.HasPrefix(name, ".") || name == "uploads" || name == "node_modules" || name == "target"
}

func shouldCopyProjectFile(info os.FileInfo) bool {
	base := info.Name()
	if strings.HasPrefix(base, ".") || strings.HasPrefix(base, ".env") ||
		strings.HasSuffix(base, ".exe") || strings.HasSuffix(base, ".bin") ||
		strings.HasSuffix(base, ".dll") || strings.HasSuffix(base, ".so") ||
		strings.HasSuffix(base, ".zip") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(base))
	return ext == ".go" || ext == ".mod" || ext == ".sum" || ext == ".json" || ext == ".txt" || ext == ".yaml" || ext == ".yml"
}
