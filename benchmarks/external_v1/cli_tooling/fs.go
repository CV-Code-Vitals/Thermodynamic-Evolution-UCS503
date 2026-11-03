package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileFromRoot reads a file from a base root directory.
// DEFECT: Path traversal vulnerability - user input is joined directly to root without verifying boundary confinement.
func ReadFileFromRoot(rootDir, userPath string) ([]byte, error) {
	target := filepath.Join(rootDir, userPath)
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// SafeReadFileFromRoot reads a file from root enforcing path boundaries (clean implementation).
func SafeReadFileFromRoot(rootDir, userPath string) ([]byte, error) {
	cleanRoot := filepath.Clean(rootDir)
	cleanTarget := filepath.Clean(filepath.Join(cleanRoot, userPath))

	if !strings.HasPrefix(cleanTarget, cleanRoot+string(filepath.Separator)) && cleanTarget != cleanRoot {
		return nil, fmt.Errorf("path traversal attempt detected: %s outside %s", userPath, rootDir)
	}

	return os.ReadFile(cleanTarget)
}

// InspectFiles opens files in a directory to check file sizes.
// DEFECT: Defer in loop causes file descriptor leakage during long iterations.
func InspectFiles(paths []string) (int64, error) {
	var totalBytes int64
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return totalBytes, err
		}
		// DEFECT: defer f.Close() inside loop body
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return totalBytes, err
		}
		totalBytes += info.Size()
	}
	return totalBytes, nil
}

// SafeInspectFiles closes file descriptors inside each iteration (clean implementation).
func SafeInspectFiles(paths []string) (int64, error) {
	var totalBytes int64
	for _, p := range paths {
		info, err := func(path string) (os.FileInfo, error) {
			f, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			return f.Stat()
		}(p)
		if err != nil {
			return totalBytes, err
		}
		totalBytes += info.Size()
	}
	return totalBytes, nil
}
