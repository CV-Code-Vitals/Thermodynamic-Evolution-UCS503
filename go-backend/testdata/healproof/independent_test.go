package healproof

import (
 "os"
 "path/filepath"
 "runtime"
 "runtime/debug"
 "testing"
)

// Authored independently of the synthesizer: measure live descriptors, not AST
// shape or the generated diff. Disabling GC prevents finalizers hiding the leak.
func TestIndependentDescriptorLeak(t *testing.T) {
 if runtime.GOOS != "linux" { t.Fatal("Linux container required") }
 previous := debug.SetGCPercent(-1)
 defer debug.SetGCPercent(previous)
 path := filepath.Join(t.TempDir(), "input")
 if err := os.WriteFile(path, []byte("expected"), 0600); err != nil { t.Fatal(err) }
 before, err := os.ReadDir("/proc/self/fd"); if err != nil { t.Fatal(err) }
 for i := 0; i < 64; i++ {
  got, err := Read(path)
  if err != nil || string(got) != "expected" { t.Fatalf("input behavior: %q %v", got, err) }
 }
 after, err := os.ReadDir("/proc/self/fd"); if err != nil { t.Fatal(err) }
 if len(after) > len(before)+2 { t.Fatalf("HEALPROOF_DESCRIPTOR_LEAK: %d -> %d", len(before), len(after)) }
}

func TestApprovedRegression(t *testing.T) {
 if _, err := Read(filepath.Join(t.TempDir(), "missing")); !os.IsNotExist(err) { t.Fatalf("missing-file error changed: %v", err) }
}

// This exact expectation is run against both baseline and candidate.
func TestBehaviorUnchanged(t *testing.T) {
 path := filepath.Join(t.TempDir(), "input")
 for _, expected := range []string{"", "hello\n", "\x00\xff\x01"} {
  if err := os.WriteFile(path, []byte(expected), 0600); err != nil { t.Fatal(err) }
  got, err := Read(path)
  if err != nil || string(got) != expected { t.Fatalf("read behavior changed: %q %v", got, err) }
 }
}
