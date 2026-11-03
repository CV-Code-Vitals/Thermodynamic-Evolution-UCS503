package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateUnifiedPatch creates a standard unified diff between original and modified files.
func GenerateUnifiedPatch(taskID string, originalRepoRoot string, wsDir string, targetFiles []string) (*UnifiedPatch, error) {
	var diffBuilder strings.Builder
	var modifiedFiles []string

	for _, relPath := range targetFiles {
		origPath := filepath.Join(originalRepoRoot, filepath.FromSlash(relPath))
		modPath := filepath.Join(wsDir, filepath.FromSlash(relPath))

		var origContent string
		if origData, err := os.ReadFile(origPath); err == nil {
			origContent = string(origData)
		}

		var modContent string
		if modData, err := os.ReadFile(modPath); err == nil {
			modContent = string(modData)
		} else if os.IsNotExist(err) && origContent != "" {
			// File was deleted
			modContent = ""
		} else {
			continue // No change or file doesn't exist in either
		}

		if origContent == modContent {
			continue // No changes
		}

		fileDiff := formatUnifiedDiff(relPath, origContent, modContent)
		if fileDiff != "" {
			diffBuilder.WriteString(fileDiff)
			diffBuilder.WriteString("\n")
			modifiedFiles = append(modifiedFiles, relPath)
		}
	}

	diffStr := diffBuilder.String()
	if diffStr == "" {
		return nil, nil // No modifications generated
	}

	hashBytes := sha256.Sum256([]byte(diffStr))
	diffHash := hex.EncodeToString(hashBytes[:])
	lineCount := len(strings.Split(diffStr, "\n"))

	patch := &UnifiedPatch{
		PatchID:     fmt.Sprintf("PATCH-%s", taskID),
		TaskID:      taskID,
		TargetFiles: modifiedFiles,
		DiffContent: diffStr,
		SHA256:      diffHash,
		LineCount:   lineCount,
		BytesCount:  len([]byte(diffStr)),
		CreatedAt:   time.Now(),
	}

	return patch, nil
}

// formatUnifiedDiff generates a standard unified diff header and hunk.
func formatUnifiedDiff(relPath string, origText, modText string) string {
	cleanRel := filepath.ToSlash(relPath)
	origLines := splitLines(origText)
	modLines := splitLines(modText)

	// Simple hunk generation
	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- a/%s\n", cleanRel))
	b.WriteString(fmt.Sprintf("+++ b/%s\n", cleanRel))

	if len(origLines) == 0 && len(modLines) > 0 {
		// New file
		b.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(modLines)))
		for _, l := range modLines {
			b.WriteString("+" + l + "\n")
		}
		return b.String()
	}

	if len(origLines) > 0 && len(modLines) == 0 {
		// Deleted file
		b.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(origLines)))
		for _, l := range origLines {
			b.WriteString("-" + l + "\n")
		}
		return b.String()
	}

	// Line-by-line diff
	var hunks []string
	origIdx, modIdx := 0, 0

	for origIdx < len(origLines) || modIdx < len(modLines) {
		if origIdx < len(origLines) && modIdx < len(modLines) && origLines[origIdx] == modLines[modIdx] {
			origIdx++
			modIdx++
			continue
		}

		// Find difference window
		startOrig := origIdx
		startMod := modIdx

		// Advance until matching lines or end
		for origIdx < len(origLines) && (modIdx >= len(modLines) || origLines[origIdx] != modLines[modIdx]) {
			origIdx++
		}
		for modIdx < len(modLines) && (origIdx >= len(origLines) || modLines[modIdx] != origLines[origIdx]) {
			modIdx++
		}

		// Emit hunk
		var hunk strings.Builder
		hunk.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", startOrig+1, origIdx-startOrig, startMod+1, modIdx-startMod))
		for i := startOrig; i < origIdx; i++ {
			hunk.WriteString("-" + origLines[i] + "\n")
		}
		for i := startMod; i < modIdx; i++ {
			hunk.WriteString("+" + modLines[i] + "\n")
		}
		hunks = append(hunks, hunk.String())
	}

	if len(hunks) == 0 {
		return ""
	}

	for _, h := range hunks {
		b.WriteString(h)
	}

	return b.String()
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	// If ends with newline, strip last empty element
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ValidatePatchScope verifies that the patch only modifies allowed target files.
func ValidatePatchScope(patch *UnifiedPatch, allowedTargets []string) error {
	if patch == nil {
		return fmt.Errorf("patch is nil")
	}

	allowedMap := make(map[string]bool)
	for _, target := range allowedTargets {
		allowedMap[filepath.ToSlash(target)] = true
	}

	for _, modified := range patch.TargetFiles {
		if !allowedMap[filepath.ToSlash(modified)] {
			return fmt.Errorf("scope violation: patch touches unpermitted file %s", modified)
		}
	}

	return nil
}

