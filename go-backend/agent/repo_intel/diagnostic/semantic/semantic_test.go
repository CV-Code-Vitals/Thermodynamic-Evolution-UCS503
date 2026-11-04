package semantic

import (
	"path/filepath"
	"testing"
)

func TestSemanticAnalyzerOnBenchmarkRepositories(t *testing.T) {
	// Test on an external benchmark repository to verify real-world robustness
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "http_router")
	sa := NewSemanticAnalyzer(repoPath)

	pkgModel, err := sa.AnalyzePackage(".")
	if err != nil {
		t.Fatalf("AnalyzePackage failed on http_router: %v", err)
	}

	if pkgModel.PackageName != "httprouter" {
		t.Errorf("expected package name 'httprouter', got '%s'", pkgModel.PackageName)
	}

	if len(pkgModel.Files) == 0 {
		t.Fatalf("expected discovered files, got 0")
	}

	// Verify symbol table has symbols extracted
	if len(pkgModel.SymbolTable.Symbols) == 0 {
		t.Errorf("expected non-empty symbol table")
	}

	// Verify CFG and dataflow models exist for each file
	totalCFGs := 0
	for _, fModel := range pkgModel.Files {
		totalCFGs += len(fModel.CFGs)
		for fnName, cfg := range fModel.CFGs {
			df := fModel.Dataflow[fnName]
			if df == nil {
				t.Errorf("missing dataflow analysis for function %s in %s", fnName, fModel.FilePath)
			}
			if cfg.Entry == nil || cfg.Exit == nil {
				t.Errorf("invalid CFG entry/exit for %s", fnName)
			}
		}
	}

	if totalCFGs == 0 {
		t.Errorf("expected CFGs to be built across package files")
	}
}

func TestSemanticAnalyzerOnSQLService(t *testing.T) {
	repoPath := filepath.Join("..", "..", "..", "..", "..", "benchmarks", "external_v1", "sql_service")
	sa := NewSemanticAnalyzer(repoPath)

	pkgModel, err := sa.AnalyzePackage(".")
	if err != nil {
		t.Fatalf("AnalyzePackage failed on sql_service: %v", err)
	}

	// Check if repository.go was analyzed
	foundRepo := false
	for path, fModel := range pkgModel.Files {
		if filepath.Base(path) == "repository.go" {
			foundRepo = true
			if _, ok := fModel.CFGs["UserRepository.FindUserByNameRaw"]; !ok && fModel.CFGs["FindUserByNameRaw"] == nil {
				t.Errorf("expected FindUserByNameRaw CFG in repository.go, got keys: %+v", getMapKeys(fModel.CFGs))
			}
		}
	}

	if !foundRepo {
		t.Errorf("repository.go not found in sql_service package files")
	}
}

func getMapKeys(m map[string]*ControlFlowGraph) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
