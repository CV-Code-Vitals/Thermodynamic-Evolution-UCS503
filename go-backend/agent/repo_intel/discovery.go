package repo_intel

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrPathOutsideRoot = errors.New("path traversal detected: target directory is outside allowed boundary")
	ErrScanCancelled   = errors.New("repository discovery scan cancelled or timed out")
)

// DiscoveryOptions configures the file discovery process.
type DiscoveryOptions struct {
	MaxFiles        int
	MaxFileSizeBytes int64
	Timeout         time.Duration
	CustomExcludes  []string
}

// DefaultDiscoveryOptions returns safe defaults.
func DefaultDiscoveryOptions() DiscoveryOptions {
	return DiscoveryOptions{
		MaxFiles:        50000,
		MaxFileSizeBytes: 10 << 20, // 10 MiB
		Timeout:         60 * time.Second,
		CustomExcludes:  nil,
	}
}

// GitIgnoreMatcher holds compiled gitignore patterns.
type GitIgnoreMatcher struct {
	patterns []string
}

// NewGitIgnoreMatcher reads patterns from a .gitignore file if present.
func NewGitIgnoreMatcher(gitignorePath string) *GitIgnoreMatcher {
	m := &GitIgnoreMatcher{
		patterns: []string{
			".git",
			"node_modules",
			"target",
			"vendor",
			".dist",
			"dist",
			".vite",
			".aws-sam",
			"*.exe",
			"*.dll",
			"*.so",
			"*.dylib",
		},
	}
	f, err := os.Open(gitignorePath)
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			m.patterns = append(m.patterns, line)
		}
	}
	return m
}

// Matches returns true if the relative path matches any ignore pattern.
func (m *GitIgnoreMatcher) Matches(relPath string, isDir bool) bool {
	normalized := filepath.ToSlash(relPath)
	base := filepath.Base(relPath)

	for _, p := range m.patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Negation not supported for simplicity
		if strings.HasPrefix(p, "!") {
			continue
		}
		cleanP := filepath.ToSlash(p)
		cleanP = strings.TrimPrefix(cleanP, "/")
		cleanP = strings.TrimSuffix(cleanP, "/")

		if cleanP == base || cleanP == normalized {
			return true
		}
		if isDir && (strings.HasPrefix(normalized, cleanP+"/") || normalized == cleanP) {
			return true
		}
		// Match wildcards
		if matched, _ := filepath.Match(cleanP, base); matched {
			return true
		}
		if matched, _ := filepath.Match(cleanP, normalized); matched {
			return true
		}
	}
	return false
}

// DiscoveredFile records information about a found file.
type DiscoveredFile struct {
	RelativePath string
	Size         int64
	Extension    string
	IsSource     bool
	IsConfig     bool
	IsManifest   bool
	IsBinary     bool
}

// DiscoveryResult holds the output of the discovery phase.
type DiscoveryResult struct {
	RepoRoot        string
	Files           []DiscoveredFile
	TotalScanned    int
	IgnoredCount    int
	LanguageCounts  map[string]int
	ManifestFiles   []string
	TestDirectories []string
	CIWorkflows     []string
	ScanDuration    time.Duration
	Warnings        []string
}

// Known extension mappings to languages.
var extensionLanguageMap = map[string]string{
	".go":    "Go",
	".rs":    "Rust",
	".py":    "Python",
	".js":    "JavaScript",
	".mjs":   "JavaScript",
	".cjs":   "JavaScript",
	".jsx":   "JavaScript / JSX",
	".ts":    "TypeScript",
	".tsx":   "TSX",
	".java":  "Java",
	".c":     "C",
	".h":     "C",
	".cpp":   "C++",
	".hpp":   "C++",
	".cc":    "C++",
	".cs":    "C#",
	".rb":    "Ruby",
	".php":   "PHP",
	".sh":    "Shell",
	".bash":  "Shell",
	".swift": "Swift",
	".json":  "JSON",
	".yaml":  "YAML",
	".yml":   "YAML",
	".toml":  "TOML",
	".md":    "Markdown",
	".sql":   "SQL",
}

// DiscoverRepository walks the given repository root safely and collects inventory.
func DiscoverRepository(ctx context.Context, repoRoot string, opts DiscoveryOptions) (*DiscoveryResult, error) {
	start := time.Now()

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve repo root: %w", err)
	}

	// Verify path exists and is a directory
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat repo root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repo root %q is not a directory", absRoot)
	}

	gitignorePath := filepath.Join(absRoot, ".gitignore")
	matcher := NewGitIgnoreMatcher(gitignorePath)
	for _, p := range opts.CustomExcludes {
		matcher.patterns = append(matcher.patterns, p)
	}

	result := &DiscoveryResult{
		RepoRoot:       absRoot,
		LanguageCounts: make(map[string]int),
	}

	testDirMap := make(map[string]bool)

	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		select {
		case <-ctx.Done():
			return ErrScanCancelled
		default:
		}

		if walkErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("cannot access path %s: %v", path, walkErr))
			return nil // Skip unreadable paths without failing entire scan
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return ErrPathOutsideRoot
		}

		if relPath == "." {
			return nil
		}

		isDir := d.IsDir()

		// Symlink safety check: evaluate symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("cannot resolve symlink %s: %v", path, err))
				return nil
			}
			relResolved, err := filepath.Rel(absRoot, resolved)
			if err != nil || strings.HasPrefix(relResolved, "..") {
				result.Warnings = append(result.Warnings, fmt.Sprintf("symlink %s points outside repository root to %s: skipping", path, resolved))
				return nil
			}
			// If symlink target is a directory and points inside repo, stat it
			targetInfo, err := os.Stat(resolved)
			if err != nil {
				return nil
			}
			isDir = targetInfo.IsDir()
		}

		// Check if path is ignored
		if matcher.Matches(relPath, isDir) {
			result.IgnoredCount++
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}

		if isDir {
			// Record test directories
			baseLower := strings.ToLower(d.Name())
			if baseLower == "test" || baseLower == "tests" || strings.HasPrefix(baseLower, "test_") {
				testDirMap[filepath.ToSlash(relPath)] = true
			}
			return nil
		}

		result.TotalScanned++
		if opts.MaxFiles > 0 && len(result.Files) >= opts.MaxFiles {
			result.Warnings = append(result.Warnings, fmt.Sprintf("reached maximum file limit of %d", opts.MaxFiles))
			return fs.SkipAll
		}

		fileInfo, err := d.Info()
		if err != nil {
			return nil
		}

		size := fileInfo.Size()
		if opts.MaxFileSizeBytes > 0 && size > opts.MaxFileSizeBytes {
			result.Warnings = append(result.Warnings, fmt.Sprintf("skipping oversized file %s (%d bytes)", relPath, size))
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		baseName := d.Name()

		df := DiscoveredFile{
			RelativePath: filepath.ToSlash(relPath),
			Size:         size,
			Extension:    ext,
		}

		// Check for CI workflows
		slashRel := filepath.ToSlash(relPath)
		if strings.HasPrefix(slashRel, ".github/workflows/") && (ext == ".yml" || ext == ".yaml") {
			result.CIWorkflows = append(result.CIWorkflows, slashRel)
		}

		// Check for manifests
		if isManifestFile(baseName) {
			df.IsManifest = true
			result.ManifestFiles = append(result.ManifestFiles, slashRel)
		}

		// Language mapping
		if lang, ok := extensionLanguageMap[ext]; ok {
			result.LanguageCounts[lang]++
			df.IsSource = true
		} else if baseName == "Dockerfile" || strings.HasPrefix(baseName, "Dockerfile.") {
			result.LanguageCounts["Dockerfile"]++
			df.IsConfig = true
		} else if baseName == "Makefile" {
			result.LanguageCounts["Makefile"]++
			df.IsConfig = true
		}

		// Fast binary heuristic
		if isBinaryExtension(ext) {
			df.IsBinary = true
			df.IsSource = false
		} else if size > 0 && size < 4096 {
			// Probe first few bytes
			if probeFileIsBinary(path) {
				df.IsBinary = true
				df.IsSource = false
			}
		}

		result.Files = append(result.Files, df)
		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
		return nil, walkErr
	}

	for td := range testDirMap {
		result.TestDirectories = append(result.TestDirectories, td)
	}

	result.ScanDuration = time.Since(start)
	return result, nil
}

func isManifestFile(name string) bool {
	switch name {
	case "go.mod", "go.sum",
		"Cargo.toml", "Cargo.lock",
		"package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock",
		"requirements.txt", "pyproject.toml", "poetry.lock", "Pipfile", "Pipfile.lock",
		"Dockerfile", "docker-compose.yml", "template.yaml":
		return true
	default:
		if strings.HasPrefix(name, "Dockerfile") || strings.HasPrefix(name, "docker-compose") {
			return true
		}
		return false
	}
}

func isBinaryExtension(ext string) bool {
	switch ext {
	case ".exe", ".bin", ".o", ".a", ".so", ".dll", ".dylib",
		".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz",
		".wasm", ".pyc", ".class":
		return true
	default:
		return false
	}
}

func probeFileIsBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

