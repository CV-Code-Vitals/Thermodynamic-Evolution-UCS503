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

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute baseDir: %w", err)
	}

	return &WorkspaceManager{
		BaseDir:        filepath.Clean(absBaseDir),
		RepoRoot:       filepath.Clean(absRepoRoot),
		KeepWorkspaces: keepWorkspaces,
	}, nil
}

// ValidatePathSafety verifies that a candidate relative path does not attempt any form
// of path traversal, absolute escape, drive escape, UNC escape, device namespace escape,
// or null byte injection across all operating systems.
func ValidatePathSafety(targetRelPath string) error {
	trimmed := strings.TrimSpace(targetRelPath)
	if trimmed == "" {
		return fmt.Errorf("security violation: path traverses outside workspace (empty path)")
	}
	if strings.ContainsRune(targetRelPath, 0) {
		return fmt.Errorf("security violation: path traverses outside workspace (null byte detected)")
	}

	// Reject native absolute paths on current host
	if filepath.IsAbs(targetRelPath) {
		return fmt.Errorf("security violation: path traverses outside workspace (native absolute path): %s", targetRelPath)
	}

	// Reject Unix/POSIX root paths or Windows root-relative paths on any OS: starts with '/' or '\'
	if strings.HasPrefix(targetRelPath, "/") || strings.HasPrefix(targetRelPath, "\\") {
		return fmt.Errorf("security violation: path traverses outside workspace (root path): %s", targetRelPath)
	}

	// Reject home directory expansion paths ("~", "~/", "~\")
	if strings.HasPrefix(targetRelPath, "~") {
		return fmt.Errorf("security violation: path traverses outside workspace (home directory expansion '~'): %s", targetRelPath)
	}

	// Reject Windows drive-absolute ("C:\...", "c:/...") and drive-relative ("C:foo", "c:relative") on any OS
	if len(targetRelPath) >= 2 && targetRelPath[1] == ':' {
		c := targetRelPath[0]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return fmt.Errorf("security violation: path traverses outside workspace (windows drive specification): %s", targetRelPath)
		}
	}

	// Reject UNC paths ("\\...", "//...", "\/...", "/\...") and device paths ("\\.\", "\\?\", "//./", "//?/")
	if strings.HasPrefix(targetRelPath, `\\`) || strings.HasPrefix(targetRelPath, `//`) ||
		strings.HasPrefix(targetRelPath, `\/`) || strings.HasPrefix(targetRelPath, `/\`) {
		return fmt.Errorf("security violation: path traverses outside workspace (UNC or device namespace): %s", targetRelPath)
	}

	// Reject directory traversal components ("..") across mixed separators
	norm := strings.ReplaceAll(targetRelPath, "\\", "/")
	segments := strings.Split(norm, "/")
	for _, seg := range segments {
		if seg == ".." {
			return fmt.Errorf("security violation: path traverses outside workspace (traversal segment '..'): %s", targetRelPath)
		}
	}

	// Clean canonical relative path check
	cleanRel := filepath.Clean(filepath.FromSlash(norm))
	if cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || strings.HasPrefix(cleanRel, "../") || strings.HasPrefix(cleanRel, `..\`) {
		return fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
	}
	if cleanRel == "." {
		return fmt.Errorf("security violation: path traverses outside workspace (targets workspace root itself): %s", targetRelPath)
	}

	return nil
}

// CreateWorkspace prepares an isolated copy of repository files for a given task.
func (wm *WorkspaceManager) CreateWorkspace(taskID string, targetFiles []string) (string, *WorkspaceAuditRecord, error) {
	sanitizedID := filepath.Base(taskID)
	sanitizedID = strings.ReplaceAll(sanitizedID, "/", "_")
	sanitizedID = strings.ReplaceAll(sanitizedID, "\\", "_")
	sanitizedID = strings.ReplaceAll(sanitizedID, ":", "_")
	sanitizedID = strings.ReplaceAll(sanitizedID, "..", "_")
	if strings.TrimSpace(sanitizedID) == "" || sanitizedID == "." || sanitizedID == "_" {
		sanitizedID = "task"
	}

	wsName := fmt.Sprintf("ws_%s_%d", sanitizedID, time.Now().UnixNano())
	wsDir := filepath.Join(wm.BaseDir, wsName)

	cleanBase, err := filepath.Abs(wm.BaseDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}
	cleanBase = filepath.Clean(cleanBase)

	cleanWs, err := filepath.Abs(wsDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve workspace directory: %w", err)
	}
	cleanWs = filepath.Clean(cleanWs)

	// Ensure wsDir is strictly contained in cleanBase (no sibling-prefix or root confusion)
	relWs, err := filepath.Rel(cleanBase, cleanWs)
	if err != nil || relWs == "." || relWs == ".." || strings.HasPrefix(relWs, ".."+string(filepath.Separator)) || filepath.IsAbs(relWs) {
		return "", nil, fmt.Errorf("security violation: workspace directory escapes base directory: %s", wsDir)
	}

	if err := os.MkdirAll(cleanWs, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create workspace dir %s: %w", cleanWs, err)
	}

	audit := &WorkspaceAuditRecord{
		TaskID:       taskID,
		WorkspaceDir: cleanWs,
		CreatedAt:    time.Now(),
		Retained:     wm.KeepWorkspaces,
		CleanedUp:    false,
	}

	// Copy repository contents into the workspace, omitting .git to prevent locking
	err = filepath.Walk(wm.RepoRoot, func(path string, info os.FileInfo, walkErr error) error {
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

		destPath := filepath.Join(cleanWs, relPath)
		cleanDest := filepath.Clean(destPath)

		// Security check: ensure destPath is strictly inside cleanWs using component-aware validation
		relDest, err := filepath.Rel(cleanWs, cleanDest)
		if err != nil || relDest == ".." || strings.HasPrefix(relDest, ".."+string(filepath.Separator)) || filepath.IsAbs(relDest) {
			return fmt.Errorf("path traversal attempt detected: %s", destPath)
		}

		if info.IsDir() {
			return os.MkdirAll(cleanDest, 0755)
		}

		// Copy file
		return copyFile(path, cleanDest)
	})

	if err != nil {
		os.RemoveAll(cleanWs)
		return "", nil, fmt.Errorf("failed to populate workspace %s: %w", cleanWs, err)
	}

	// Validate targetFiles if specified
	for _, tf := range targetFiles {
		if _, err := wm.ValidateContainedPath(cleanWs, tf); err != nil {
			os.RemoveAll(cleanWs)
			return "", nil, fmt.Errorf("security violation in target file %q: %w", tf, err)
		}
	}

	// Measure size
	var totalSize int64
	_ = filepath.Walk(cleanWs, func(_ string, fi os.FileInfo, _ error) error {
		if fi != nil && !fi.IsDir() {
			totalSize += fi.Size()
		}
		return nil
	})
	audit.SizeBytes = totalSize

	return cleanWs, audit, nil
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

	if strings.TrimSpace(wsDir) == "" {
		return fmt.Errorf("refusing to destroy empty workspace directory")
	}

	cleanBase, err := filepath.Abs(wm.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve base directory: %w", err)
	}
	cleanBase = filepath.Clean(cleanBase)

	cleanWs, err := filepath.Abs(wsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace directory: %w", err)
	}
	cleanWs = filepath.Clean(cleanWs)

	// Refuse to destroy base directory itself!
	if cleanWs == cleanBase {
		return fmt.Errorf("refusing to destroy base directory itself: %s", wsDir)
	}

	// Canonical relative-path semantics verification (prevents sibling-prefix confusion)
	rel, err := filepath.Rel(cleanBase, cleanWs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to destroy directory outside base workspace: %s", wsDir)
	}

	// Symlink escape check: ensure cleanWs is not a symlink pointing outside cleanBase
	realBase := cleanBase
	if evalBase, err := filepath.EvalSymlinks(cleanBase); err == nil {
		realBase = filepath.Clean(evalBase)
	}
	if realWs, err := filepath.EvalSymlinks(cleanWs); err == nil {
		realWs = filepath.Clean(realWs)
		if realWs == realBase {
			return fmt.Errorf("refusing to destroy base directory itself via symlink: %s", wsDir)
		}
		relReal, errRel := filepath.Rel(realBase, realWs)
		if errRel != nil || relReal == "." || relReal == ".." || strings.HasPrefix(relReal, ".."+string(filepath.Separator)) || filepath.IsAbs(relReal) {
			return fmt.Errorf("refusing to destroy directory outside base workspace via symlink: %s", wsDir)
		}
	}

	err = os.RemoveAll(cleanWs)
	if err == nil && audit != nil {
		audit.CleanedUp = true
	}
	return err
}

// ValidateContainedPath ensures that a target file path resolves strictly within the workspace.
func (wm *WorkspaceManager) ValidateContainedPath(wsDir string, targetRelPath string) (string, error) {
	if err := ValidatePathSafety(targetRelPath); err != nil {
		return "", err
	}

	cleanWs, err := filepath.Abs(wsDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace directory: %w", err)
	}
	cleanWs = filepath.Clean(cleanWs)

	// Normalize separators and clean
	normRel := filepath.FromSlash(strings.ReplaceAll(targetRelPath, "\\", "/"))
	cleanRel := filepath.Clean(normRel)

	fullPath := filepath.Join(cleanWs, cleanRel)
	cleanFull := filepath.Clean(fullPath)

	// Canonical relative-path semantics verification (prevents sibling-prefix confusion)
	rel, err := filepath.Rel(cleanWs, cleanFull)
	if err != nil {
		return "", fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
	}
	if rel == "." {
		return "", fmt.Errorf("security violation: path traverses outside workspace (targets workspace root itself): %s", targetRelPath)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("security violation: path traverses outside workspace: %s", targetRelPath)
	}

	// Symlink escape and TOCTOU protection
	realWs := cleanWs
	if evalWs, err := filepath.EvalSymlinks(cleanWs); err == nil {
		realWs = filepath.Clean(evalWs)
	}

	if fi, err := os.Lstat(cleanFull); err == nil {
		// Target path exists on disk: evaluate symlinks
		realFull, err := filepath.EvalSymlinks(cleanFull)
		if err != nil {
			return "", fmt.Errorf("security violation: failed to resolve symlink target: %w", err)
		}
		realFull = filepath.Clean(realFull)
		relReal, errRel := filepath.Rel(realWs, realFull)
		if errRel != nil || relReal == "." || relReal == ".." || strings.HasPrefix(relReal, ".."+string(filepath.Separator)) || filepath.IsAbs(relReal) {
			return "", fmt.Errorf("security violation: symlink resolves outside workspace boundary: %s", targetRelPath)
		}
		_ = fi
	} else {
		// Target path does not exist yet: find existing ancestor and verify its symlink containment
		curr := filepath.Dir(cleanFull)
		for {
			if _, err := os.Lstat(curr); err == nil {
				realAncestor, err := filepath.EvalSymlinks(curr)
				if err == nil {
					realAncestor = filepath.Clean(realAncestor)
					relAnc, errAnc := filepath.Rel(realWs, realAncestor)
					if errAnc != nil || (relAnc != "." && (relAnc == ".." || strings.HasPrefix(relAnc, ".."+string(filepath.Separator)) || filepath.IsAbs(relAnc))) {
						return "", fmt.Errorf("security violation: parent directory symlink resolves outside workspace boundary: %s", targetRelPath)
					}
				}
				break
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	return cleanFull, nil
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

// ResolveWorkspaceRoot dynamically detects the workspace or repository root
// across all platforms (Linux, macOS, Windows) without relying on hardcoded paths.
func ResolveWorkspaceRoot() string {
	if custom := os.Getenv("WORKSPACE_ROOT"); custom != "" {
		if fi, err := os.Stat(custom); err == nil && fi.IsDir() {
			abs, err := filepath.Abs(custom)
			if err == nil {
				return abs
			}
			return custom
		}
	}
	if custom := os.Getenv("SWE_WORKSPACE"); custom != "" {
		if fi, err := os.Stat(custom); err == nil && fi.IsDir() {
			abs, err := filepath.Abs(custom)
			if err == nil {
				return abs
			}
			return custom
		}
	}

	// 1. Walk upwards from current working directory
	wd, err := os.Getwd()
	if err == nil {
		curr := wd
		for {
			if fi, err := os.Stat(filepath.Join(curr, "benchmarks", "external_v2")); err == nil && fi.IsDir() {
				abs, err := filepath.Abs(curr)
				if err == nil {
					return abs
				}
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr || parent == "" {
				break
			}
			curr = parent
		}
	}

	// 2. Check candidate relative directories
	candidates := []string{
		".",
		"..",
		filepath.Join("..", ".."),
		filepath.Join("..", "..", ".."),
		filepath.Join("..", "..", "..", ".."),
	}
	for _, cand := range candidates {
		if fi, err := os.Stat(filepath.Join(cand, "benchmarks", "external_v2")); err == nil && fi.IsDir() {
			abs, err := filepath.Abs(cand)
			if err == nil {
				return abs
			}
		}
	}

	// 3. Fallback to current working directory or system temp
	if wd != "" {
		abs, err := filepath.Abs(wd)
		if err == nil {
			return abs
		}
		return wd
	}
	return os.TempDir()
}

// EnsureTmpDir verifies or creates the temporary directory for sandboxes and ledgers.
// It tries <root>/.tmp first; if creating or accessing it fails, it safely falls back
// to os.TempDir()/swe_tmp.
func EnsureTmpDir(root string) (string, error) {
	if root != "" {
		primary := filepath.Join(root, ".tmp")
		if err := os.MkdirAll(primary, 0755); err == nil {
			return primary, nil
		}
	}
	fallback := filepath.Join(os.TempDir(), "swe_tmp")
	if err := os.MkdirAll(fallback, 0755); err != nil {
		return os.TempDir(), nil
	}
	return fallback, nil
}


