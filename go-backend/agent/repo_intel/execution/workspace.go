package execution

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WorkspaceManager manages isolated temporary workspaces.
type WorkspaceManager struct {
	BaseDir        string
	RepoRoot       string
	KeepWorkspaces bool
}

// NewWorkspaceManager initializes a WorkspaceManager.
func NewWorkspaceManager(repoRoot string, baseDir string, keepWorkspaces bool) (*WorkspaceManager, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repoRoot cannot be empty")
	}
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute repoRoot: %w", err)
	}

	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "swe_remediation_workspaces")
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace base dir %s: %w", baseDir, err)
	}

	return &WorkspaceManager{
		BaseDir:        baseDir,
		RepoRoot:       absRepoRoot,
		KeepWorkspaces: keepWorkspaces,
	}, nil
}

// CreateWorkspace prepares an isolated copy of repository files for a given task.
func (wm *WorkspaceManager) CreateWorkspace(taskID string, targetFiles []string) (string, *WorkspaceAuditRecord, error) {
	sanitizedID := strings.ReplaceAll(taskID, "/", "_")
	sanitizedID = strings.ReplaceAll(sanitizedID, "\\", "_")
	wsName := fmt.Sprintf("ws_%s_%d", sanitizedID, time.Now().UnixNano())
	wsDir := filepath.Join(wm.BaseDir, wsName)

	if err := os.MkdirAll(wsDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create workspace dir %s: %w", wsDir, err)
	}

	audit := &WorkspaceAuditRecord{
		TaskID:       taskID,
		WorkspaceDir: wsDir,
		CreatedAt:    time.Now(),
		Retained:     wm.KeepWorkspaces,
		CleanedUp:    false,
	}

	// Copy repository contents into the workspace, omitting .git to prevent locking
	err := filepath.Walk(wm.RepoRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(wm.RepoRoot, path)
		if err != nil {
			return err
		}

		// Skip .git directory and previous temporary workspaces
		if relPath == ".git" || strings.HasPrefix(relPath, ".git"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if relPath == ".workspaces" || strings.HasPrefix(relPath, ".workspaces"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(wsDir, relPath)

		// Security check: ensure destPath is inside wsDir
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(wsDir)) {
			return fmt.Errorf("path traversal attempt detected: %s", destPath)
		}

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Copy file
		return copyFile(path, destPath)
	})

	if err != nil {
		os.RemoveAll(wsDir)
		return "", nil, fmt.Errorf("failed to populate workspace %s: %w", wsDir, err)
	}

	// Measure size
	var totalSize int64
	_ = filepath.Walk(wsDir, func(_ string, fi os.FileInfo, _ error) error {
		if fi != nil && !fi.IsDir() {
			totalSize += fi.Size()
		}
		return nil
	})
	audit.SizeBytes = totalSize

	return wsDir, audit, nil
}

// DestroyWorkspace removes an isolated workspace unless retention is enabled.
func (wm *WorkspaceManager) DestroyWorkspace(wsDir string, audit *WorkspaceAuditRecord) error {
	if audit != nil {
		audit.DestroyedAt = time.Now()
	}

	if wm.KeepWorkspaces {
		if audit != nil {
			audit.Retained = true
		}
		return nil
	}

	// Guard against deleting unintended directories
	if !strings.HasPrefix(filepath.Clean(wsDir), filepath.Clean(wm.BaseDir)) {
		return fmt.Errorf("refusing to destroy directory outside base workspace: %s", wsDir)
	}

	err := os.RemoveAll(wsDir)
	if err == nil && audit != nil {
		audit.CleanedUp = true
	}
	return err
}

// ValidateContainedPath ensures that a target file path resolves strictly within the workspace.
func (wm *WorkspaceManager) ValidateContainedPath(wsDir string, targetRelPath string) (string, error) {
	cleanRel := filepath.Clean(filepath.FromSlash(targetRelPath))
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return "", fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
	}

	fullPath := filepath.Join(wsDir, cleanRel)
	cleanFull := filepath.Clean(fullPath)
	cleanWs := filepath.Clean(wsDir)

	if !strings.HasPrefix(cleanFull, cleanWs) {
		return "", fmt.Errorf("security violation: path escapes workspace boundary: %s", targetRelPath)
	}

	return fullPath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

