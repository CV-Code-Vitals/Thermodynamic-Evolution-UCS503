package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const workspaceLifetime = 30 * time.Minute
const maxWorkspaces = 8

type evolutionTarget struct {
	FilePath     string  `json:"file_path"`
	FunctionName string  `json:"function_name"`
	StartByte    int     `json:"start_byte"`
	EndByte      int     `json:"end_byte"`
	BaseCode     string  `json:"base_code"`
	BaseEnergy   float64 `json:"base_energy"`
	SourceHash   string  `json:"source_hash"`
}

type evolutionWorkspace struct {
	path      string
	busy      bool
	expiresAt time.Time
	timer     *time.Timer
}

type workspaceStore struct {
	mu      sync.Mutex
	entries map[string]*evolutionWorkspace
}

func newWorkspaceStore() *workspaceStore {
	s := &workspaceStore{entries: make(map[string]*evolutionWorkspace)}
	s.mu.Lock()
	defer s.mu.Unlock()
	paths, _ := filepath.Glob(filepath.Join("uploads", "workspace_*"))
	for _, path := range paths {
		id := strings.TrimPrefix(filepath.Base(path), "workspace_")
		if _, err := uuid.Parse(id); err != nil {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		path, err = filepath.Abs(path)
		if err != nil {
			continue
		}
		expiresAt := info.ModTime().Add(workspaceLifetime)
		if !expiresAt.After(time.Now()) {
			_ = os.RemoveAll(path)
			continue
		}
		w := &evolutionWorkspace{path: path, expiresAt: expiresAt}
		s.entries[id] = w
		w.timer = time.AfterFunc(time.Until(expiresAt), func() { s.expire(id) })
	}
	return s
}

func (s *workspaceStore) create() (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= maxWorkspaces {
		return "", "", errors.New("workspace capacity reached; wait for an existing workspace to expire")
	}
	if err := os.MkdirAll("uploads", 0o750); err != nil {
		return "", "", err
	}
	id := uuid.NewString()
	path, err := filepath.Abs(filepath.Join("uploads", "workspace_"+id))
	if err != nil {
		return "", "", err
	}
	w := &evolutionWorkspace{path: path, busy: true, expiresAt: time.Now().Add(workspaceLifetime)}
	s.entries[id] = w
	w.timer = time.AfterFunc(workspaceLifetime, func() { s.expire(id) })
	return id, path, nil
}

func (s *workspaceStore) expire(id string) {
	s.mu.Lock()
	w := s.entries[id]
	if w == nil {
		s.mu.Unlock()
		return
	}
	if w.busy {
		w.timer.Reset(time.Minute)
		s.mu.Unlock()
		return
	}
	delete(s.entries, id)
	s.mu.Unlock()
	_ = os.RemoveAll(w.path)
}

func (s *workspaceStore) finishScan(id string, success bool) {
	s.mu.Lock()
	w := s.entries[id]
	if w != nil {
		w.busy = false
		if !success {
			w.timer.Stop()
			delete(s.entries, id)
		}
	}
	s.mu.Unlock()
	if w != nil && !success {
		_ = os.RemoveAll(w.path)
	}
}

func (s *workspaceStore) acquire(id string) (string, func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.entries[id]
	if w == nil || !w.expiresAt.After(time.Now()) {
		return "", nil, errors.New("workspace expired or unavailable; scan the repository again")
	}
	if w.busy {
		return "", nil, errors.New("workspace is already in use")
	}
	w.busy = true
	return w.path, func() {
		s.mu.Lock()
		w.busy = false
		s.mu.Unlock()
	}, nil
}

func (s *workspaceStore) get(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.entries[id]
	if w == nil || !w.expiresAt.After(time.Now()) {
		return "", errors.New("workspace expired or unavailable; scan the repository again")
	}
	return w.path, nil
}

func readWorkspaceTarget(root, relativePath, functionName string, startByte int) (*evolutionTarget, error) {
	findTarget, err := loadWorkspaceTargets(root, relativePath)
	if err != nil {
		return nil, err
	}
	return findTarget(functionName, startByte)
}

func loadWorkspaceTargets(root, relativePath string) (func(string, int) (*evolutionTarget, error), error) {
	if filepath.Ext(relativePath) != ".go" || strings.HasSuffix(relativePath, "_test.go") || filepath.IsAbs(relativePath) || strings.Contains(relativePath, "\\") {
		return nil, errors.New("select a Go function from the scan results")
	}
	path, err := filepath.EvalSymlinks(filepath.Join(root, relativePath))
	if err != nil {
		return nil, err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(resolvedRoot, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("function path escapes workspace")
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, relativePath, source, 0)
	if err != nil {
		return nil, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(source))
	functions := make(map[string][]*ast.FuncDecl)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			functions[function.Name.Name] = append(functions[function.Name.Name], function)
		}
	}
	return func(functionName string, startByte int) (*evolutionTarget, error) {
		matches := functions[functionName]
		if len(matches) != 1 {
			return nil, errors.New("energy analysis requires an unambiguous function name within the file")
		}
		function := matches[0]
		if function.Body != nil {
			start := fset.Position(function.Pos()).Offset
			end := fset.Position(function.End()).Offset
			if start == startByte {
				return &evolutionTarget{FilePath: filepath.ToSlash(rel), FunctionName: functionName,
					StartByte: start, EndByte: end, BaseCode: string(source[start:end]),
					SourceHash: hash}, nil
			}
		}
		return nil, fmt.Errorf("function changed or was not found; scan the repository again")
	}, nil
}

func workspaceModule(root, filePath string) (string, string, error) {
	fullPath := filepath.Join(root, filePath)
	for directory := filepath.Dir(fullPath); ; directory = filepath.Dir(directory) {
		info, err := os.Lstat(filepath.Join(directory, "go.mod"))
		if err == nil && info.Mode().IsRegular() {
			relative, err := filepath.Rel(directory, fullPath)
			return directory, relative, err
		}
		if directory == root || filepath.Dir(directory) == directory {
			break
		}
	}
	return "", "", errors.New("selected function has no Go module; choose a repository with a go.mod file")
}
