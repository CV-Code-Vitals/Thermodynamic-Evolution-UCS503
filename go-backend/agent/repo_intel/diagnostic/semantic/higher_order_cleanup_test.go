package semantic

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// 1. Mandatory Example: defer check(resp.Body.Close)
func TestHigherOrder_DeferredMethodValue_MandatoryExample(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "net/http"

func check(fn func() error) {
	_ = fn() // Unconditional invocation
}

func FetchData(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer check(resp.Body.Close) // Mandatory pattern!
	return nil
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "mandatory.go", src, 0)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	fnCheck := funcOrGeneric(file.Decls[1])
	fnFetch := funcOrGeneric(file.Decls[2])

	resAnalyzer.hocAnalyzer.RegisterLocalFunc(fnCheck)
	cfg := BuildCFG(fnFetch, fset, "mandatory.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnFetch, fset, cfg, nil, "mandatory.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Mandatory pattern defer check(resp.Body.Close) was flagged as leak: %+v", ev)
		}
	}
	if len(evs) == 0 || evs[0].LifecycleState != StateResourceDeferredRelease {
		t.Fatalf("Expected StateResourceDeferredRelease, got: %+v", evs)
	}
}

// 2. Deferred Anonymous Function: defer func() { f.Close() }()
func TestHigherOrder_DeferredAnonymousFunc(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "os"

func RunAnon(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "anon.go", src, 0)
	fnDecl := funcOrGeneric(file.Decls[1])
	cfg := BuildCFG(fnDecl, fset, "anon.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnDecl, fset, cfg, nil, "anon.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			t.Fatalf("Anonymous closure cleanup flagged as leak: %+v", ev)
		}
	}
	if len(evs) == 0 || evs[0].LifecycleState != StateResourceDeferredRelease {
		t.Fatalf("Expected StateResourceDeferredRelease, got: %+v", evs)
	}
}

// 3. Counterfactual: Wrapper that never invokes the callback (No-op wrapper)
func TestHigherOrder_WrapperNeverInvokes_Counterfactual(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "net/http"

func noopWrapper(fn func() error) {
	// Intentionally does NOT invoke fn!
}

func FetchLeaky(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer noopWrapper(resp.Body.Close) // Never executes release!
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "noop.go", src, 0)

	fnNoop := funcOrGeneric(file.Decls[1])
	fnFetch := funcOrGeneric(file.Decls[2])

	resAnalyzer.hocAnalyzer.RegisterLocalFunc(fnNoop)
	cfg := BuildCFG(fnFetch, fset, "noop.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnFetch, fset, cfg, nil, "noop.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	hasConfirmedLeak := false
	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			hasConfirmedLeak = true
			break
		}
	}
	if !hasConfirmedLeak {
		t.Fatalf("Counterfactual failed: noopWrapper should trigger confirmed leak finding, got: %+v", evs)
	}
}

// 4. Wrapper that stores callback for later
func TestHigherOrder_WrapperStoresCallback(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "os"

var savedCallback func() error

func storeWrapper(fn func() error) {
	savedCallback = fn // Stored, not invoked!
}

func StoreRun(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer storeWrapper(f.Close)
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "store.go", src, 0)

	fnStore := funcOrGeneric(file.Decls[2])
	fnRun := funcOrGeneric(file.Decls[3])

	resAnalyzer.hocAnalyzer.RegisterLocalFunc(fnStore)
	cfg := BuildCFG(fnRun, fset, "store.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnRun, fset, cfg, nil, "store.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	hasLeak := false
	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			hasLeak = true
			break
		}
	}
	if !hasLeak {
		t.Fatalf("Storing callback must not be treated as release proof: %+v", evs)
	}
}

// 5. Wrapper that invokes callback asynchronously (in a goroutine)
func TestHigherOrder_WrapperAsyncGoroutine(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "os"

func asyncWrapper(fn func() error) {
	go fn() // Asynchronous, not guaranteed synchronous cleanup on exit
}

func AsyncRun(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer asyncWrapper(f.Close)
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "async.go", src, 0)

	fnAsync := funcOrGeneric(file.Decls[1])
	fnRun := funcOrGeneric(file.Decls[2])

	resAnalyzer.hocAnalyzer.RegisterLocalFunc(fnAsync)
	cfg := BuildCFG(fnRun, fset, "async.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnRun, fset, cfg, nil, "async.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	hasLeak := false
	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			hasLeak = true
			break
		}
	}
	if !hasLeak {
		t.Fatalf("Asynchronous invocation must not be treated as provable local cleanup: %+v", evs)
	}
}

// 6. Conditional invocation inside wrapper
func TestHigherOrder_WrapperConditional(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")
	resAnalyzer := NewResourceAnalyzer(repoRoot)

	src := `package testwrapper
import "os"

var shouldCloseGlobal = false

func conditionalWrapper(fn func() error) {
	if shouldCloseGlobal {
		_ = fn()
	}
}

func CondRun(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer conditionalWrapper(f.Close)
	return nil
}`
	fset := token.NewFileSet()
	file, _ := parser.ParseFile(fset, "cond.go", src, 0)

	fnCond := funcOrGeneric(file.Decls[2])
	fnRun := funcOrGeneric(file.Decls[3])

	resAnalyzer.hocAnalyzer.RegisterLocalFunc(fnCond)
	cfg := BuildCFG(fnRun, fset, "cond.go")

	evs, err := resAnalyzer.AnalyzeFunction(fnRun, fset, cfg, nil, "cond.go")
	if err != nil {
		t.Fatalf("AnalyzeFunction failed: %v", err)
	}

	hasLeak := false
	for _, ev := range evs {
		if ev.FindingState == StateConfirmed {
			hasLeak = true
			break
		}
	}
	if !hasLeak {
		t.Fatalf("Conditional invocation in wrapper must be conservatively flagged: %+v", evs)
	}
}
