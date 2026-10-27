package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// Snapshots have no Git metadata, so compare the exact scanned and tested bytes.
func exportEvolutionPatch(original, updated []byte, filePath string) (string, error) {
	if bytes.Equal(original, updated) {
		return "", nil
	}
	if !filepath.IsLocal(filepath.FromSlash(filePath)) {
		return "", errors.New("invalid patch file path")
	}
	dir, err := os.MkdirTemp("", "ucs503-patch-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	for name, content := range map[string][]byte{"a": original, "b": updated} {
		path := filepath.Join(dir, name, filepath.FromSlash(filePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, content, 0o600); err != nil {
			return "", err
		}
	}
	command := exec.Command("git", "diff", "--no-index", "--no-prefix", "--no-ext-diff", "--no-textconv", "--", "a/"+filePath, "b/"+filePath)
	command.Dir = dir
	patch, err := command.Output()
	var exitError *exec.ExitError
	if err != nil && (!errors.As(err, &exitError) || exitError.ExitCode() != 1) {
		return "", err
	}
	return string(patch), nil
}
