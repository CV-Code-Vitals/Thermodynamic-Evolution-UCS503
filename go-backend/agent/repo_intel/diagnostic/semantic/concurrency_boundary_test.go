package semantic

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// 1. Channel send: ch <- f (Classified as POSSIBLE_ESCAPE)
func TestConcurrency_ChannelSendEscape(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testconcurrency
import "os"

func SendChannel(ch chan *os.File, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	ch <- f // Concurrency transfer!
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "chan.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])
	cfg := BuildCFG(fnDecl, fset, "chan.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "chan.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	if len(evs) == 0 {
		t.Fatalf("Expected evidence for channel send")
	}

	ev := evs[0]
	// Must be classified as UNRESOLVED / POSSIBLE_ESCAPE, NOT verified local release
	if ev.LifecycleState == StateResourceReleased {
		t.Errorf("Channel send must not be treated as clean local release")
	}
	// Ineligibility check: findings with concurrent uncertainty must NOT be eligible for autonomous remediation
	reconciler := NewFindingReconciler(repoRoot)
	ledger := reconciler.ReconcileFindings(nil, nil, nil, evs, nil, "test-commit")
	for _, f := range ledger.Findings {
		if f.RemediationEligibility == "AUTONOMOUS_ELIGIBLE" {
			t.Errorf("Concurrent channel transfer must NOT be AUTONOMOUS_ELIGIBLE, got: %s", f.RemediationEligibility)
		}
	}
}

// 2. Goroutine spawn: go worker(f) (Classified as POSSIBLE_ESCAPE / UNKNOWN_CONCURRENT_OWNERSHIP)
func TestConcurrency_GoroutineSpawn(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testconcurrency
import "os"

func worker(f *os.File) {
	defer f.Close()
}

func SpawnGoroutine(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	go worker(f) // Asynchronous handoff!
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "go.go", src, 0)
	fnSpawn := funcOrGeneric(file.Decls[2])
	cfg := BuildCFG(fnSpawn, fset, "go.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnSpawn, fset, cfg, nil, "go.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	if len(evs) == 0 {
		t.Fatalf("Expected evidence for goroutine worker spawn")
	}

	ev := evs[0]
	if ev.LifecycleState == StateResourceReleased {
		t.Errorf("Goroutine handoff must not be treated as clean local release")
	}
}

// 3. Provable Local vs. Concurrent Escape classification check
func TestConcurrency_ClassificationAudit(t *testing.T) {
	scenarios := []struct {
		Name           string
		Code           string
		ExpectedClass  ConcurrencyOwnershipClass
	}{
		{
			Name: "Provable Local (Strict defer f.Close)",
			Code: `package c
import "os"
func LocalOnly(p string) {
	f, _ := os.Open(p)
	defer f.Close()
}`,
			ExpectedClass: ConcurrencyProvableLocal,
		},
		{
			Name: "Possible Escape (Channel Send)",
			Code: `package c
import "os"
func ChanEscape(p string, ch chan *os.File) {
	f, _ := os.Open(p)
	ch <- f
}`,
			ExpectedClass: ConcurrencyPossibleEscape,
		},
		{
			Name: "Possible Escape (Goroutine spawn)",
			Code: `package c
import "os"
func GoEscape(p string) {
	f, _ := os.Open(p)
	go func() { f.Close() }()
}`,
			ExpectedClass: ConcurrencyPossibleEscape,
		},
	}

	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	for _, sc := range scenarios {
		t.Run(sc.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, _ := parser.ParseFile(fset, "test.go", sc.Code, 0)
			fnDecl := funcOrGeneric(file.Decls[1])
			cfg := BuildCFG(fnDecl, fset, "test.go")

			evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "test.go")
			if err != nil {
				t.Fatalf("AnalyzeFunction failed: %v", err)
			}
			if len(evs) == 0 {
				t.Fatalf("Expected evidence for scenario %s", sc.Name)
			}
			// In provable local, lifecycle is released; in escape, lifecycle is unresolved/leak
			if sc.ExpectedClass == ConcurrencyProvableLocal {
				if evs[0].LifecycleState != StateResourceDeferredRelease && evs[0].LifecycleState != StateResourceReleased {
					t.Errorf("Expected provable local release, got: %s", evs[0].LifecycleState)
				}
			} else {
				if evs[0].LifecycleState == StateResourceReleased {
					t.Errorf("Concurrent scenario %s should not be certified as clean release", sc.Name)
				}
			}
		})
	}
}
