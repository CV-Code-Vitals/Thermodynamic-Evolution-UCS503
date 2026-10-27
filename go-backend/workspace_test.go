package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceRetainsSourceAndSerializesEvolutionUntilExpiry(t *testing.T) {
	t.Chdir(t.TempDir())
	store := newWorkspaceStore()
	id, path, err := store.create()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.finishScan(id, false) })
	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.acquire(id); err == nil {
		t.Fatal("acquired workspace while scan is running")
	}
	store.finishScan(id, true)
	got, release, err := store.acquire(id)
	if err != nil || got != path {
		t.Fatalf("acquire: %s %v", got, err)
	}
	if _, _, err := store.acquire(id); err == nil {
		t.Fatal("allowed simultaneous changes to a workspace")
	}
	store.mu.Lock()
	store.entries[id].expiresAt = time.Now().Add(-time.Second)
	store.mu.Unlock()
	store.expire(id)
	if _, err := os.Stat(path); err != nil {
		t.Fatal("deleted a workspace while evolution was active")
	}
	release()
	if _, _, err := store.acquire(id); err == nil {
		t.Fatal("acquired expired workspace")
	}
	store.expire(id)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expired workspace was not deleted")
	}
}

func TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	source := "package sample\n// π\ntype A struct{}\ntype B struct{}\nfunc (A) Other() int { return 1 }\nfunc (B) Run() int { return 2 }\n"
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	start := strings.Index(source, "func (B)")
	target, err := readWorkspaceTarget(root, "sample.go", "Run", start)
	if err != nil {
		t.Fatal(err)
	}
	if target.BaseCode != "func (B) Run() int { return 2 }" || target.EndByte != start+len(target.BaseCode) {
		t.Fatalf("incorrect target boundary: %+v", target)
	}
	if target.SourceHash != fmt.Sprintf("%x", sha256.Sum256([]byte(source))) {
		t.Fatal("target did not include the scanned file fingerprint")
	}
	for _, path := range []string{"../outside.go", "/etc/passwd", `..\outside.go`, "sample.py", "sample_test.go"} {
		if _, err := readWorkspaceTarget(root, path, "Run", start); err == nil {
			t.Errorf("accepted %q", path)
		}
	}
	if _, err := readWorkspaceTarget(root, "sample.go", "Run", start+1); err == nil {
		t.Fatal("accepted stale byte offset")
	}
	if err := os.WriteFile(filepath.Join(root, "duplicate.go"), []byte(strings.Replace(source, "Other", "Run", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkspaceTarget(root, "duplicate.go", "Run", start); err == nil {
		t.Fatal("accepted an ambiguous energy target")
	}
	outside := filepath.Join(t.TempDir(), "outside.go")
	if err := os.WriteFile(outside, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.go")); err == nil {
		if _, err := readWorkspaceTarget(root, "link.go", "Run", start); err == nil {
			t.Fatal("followed symlink outside workspace")
		}
	}
}

func TestWorkspaceRestartCleansExpiredCopies(t *testing.T) {
	t.Chdir(t.TempDir())
	store := newWorkspaceStore()
	id, path, err := store.create()
	if err != nil {
		t.Fatal(err)
	}
	store.entries[id].timer.Stop()
	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	restarted := newWorkspaceStore()
	if len(restarted.entries) != 0 {
		t.Fatal("restored expired workspace")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expired copy remained after restart")
	}
}

func TestWorkspaceModuleSelectsNearestGoModule(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "examples", "demo")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, _, err := workspaceModule(root, "examples/demo/main.go"); err == nil {
		t.Fatal("accepted a missing module")
	}
	for _, directory := range []string{root, nested} {
		if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module demo\n\ngo 1.25\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	module, relative, err := workspaceModule(root, "examples/demo/main.go")
	if err != nil || module != nested || relative != "main.go" {
		t.Fatalf("module=%q file=%q err=%v", module, relative, err)
	}
}
