package semantic

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// FileSemanticModel encapsulates AST, Symbol, CFG, and Dataflow views of a single source file.
type FileSemanticModel struct {
	FilePath string                                 `json:"file_path"`
	Package  string                                 `json:"package"`
	FileSet  *token.FileSet                         `json:"-"`
	ASTFile  *ast.File                              `json:"-"`
	CFGs     map[string]*ControlFlowGraph           `json:"cfgs"`
	Dataflow map[string]*ReachingDefinitionsAnalysis `json:"dataflow"`
}

// PackageSemanticModel encapsulates the aggregate semantic intelligence of a package directory.
type PackageSemanticModel struct {
	PackagePath string                        `json:"package_path"`
	PackageName string                        `json:"package_name"`
	SymbolTable *SymbolTable                  `json:"symbol_table"`
	Files       map[string]*FileSemanticModel `json:"files"`
}

// SemanticAnalyzer coordinates parsing, semantic representation generation, and dataflow modeling.
type SemanticAnalyzer struct {
	repoRoot string
}

// NewSemanticAnalyzer creates an instance of SemanticAnalyzer rooted at the specified repository path.
func NewSemanticAnalyzer(repoRoot string) *SemanticAnalyzer {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return &SemanticAnalyzer{repoRoot: repoRoot}
}

// AnalyzeSource parses an in-memory Go source string and returns its semantic model and symbol table.
func (sa *SemanticAnalyzer) AnalyzeSource(fileName, src string) (*FileSemanticModel, *SymbolTable, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, fileName, src, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse source: %w", err)
	}

	st, err := ExtractPackageSymbols(fset, []*ast.File{astFile}, filepath.Dir(fileName))
	if err != nil {
		return nil, nil, fmt.Errorf("extract symbols: %w", err)
	}

	model := &FileSemanticModel{
		FilePath: fileName,
		Package:  astFile.Name.Name,
		FileSet:  fset,
		ASTFile:  astFile,
		CFGs:     make(map[string]*ControlFlowGraph),
		Dataflow: make(map[string]*ReachingDefinitionsAnalysis),
	}

	for _, decl := range astFile.Decls {
		fnDecl, ok := decl.(*ast.FuncDecl)
		if !ok || fnDecl.Body == nil {
			continue
		}

		key := fnDecl.Name.Name
		if fnDecl.Recv != nil && len(fnDecl.Recv.List) > 0 {
			recvType := formatTypeExpr(fnDecl.Recv.List[0].Type, fset)
			key = fmt.Sprintf("%s.%s", strings.TrimPrefix(recvType, "*"), fnDecl.Name.Name)
		}

		cfg := BuildCFG(fnDecl, fset, fileName)
		if cfg != nil {
			model.CFGs[key] = cfg
			model.Dataflow[key] = ComputeReachingDefinitions(cfg, fset)
		}
	}

	return model, st, nil
}

// AnalyzeFile loads and analyzes a single Go file from disk.
func (sa *SemanticAnalyzer) AnalyzeFile(relPath string) (*FileSemanticModel, *SymbolTable, error) {
	fullPath := filepath.Join(sa.repoRoot, filepath.FromSlash(relPath))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read file %s: %w", relPath, err)
	}
	return sa.AnalyzeSource(relPath, string(data))
}

// AnalyzePackage discovers all Go files in a directory and builds a PackageSemanticModel.
func (sa *SemanticAnalyzer) AnalyzePackage(dirRelPath string) (*PackageSemanticModel, error) {
	fullDir := filepath.Join(sa.repoRoot, filepath.FromSlash(dirRelPath))
	entries, err := os.ReadDir(fullDir)
	if err != nil {
		return nil, fmt.Errorf("read package dir %s: %w", dirRelPath, err)
	}

	fset := token.NewFileSet()
	var astFiles []*ast.File
	var filePaths []string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		filePath := filepath.Join(fullDir, e.Name())
		relFile := filepath.Join(dirRelPath, e.Name())
		node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}
		astFiles = append(astFiles, node)
		filePaths = append(filePaths, relFile)
	}

	if len(astFiles) == 0 {
		return nil, fmt.Errorf("no go source files found in %s", dirRelPath)
	}

	st, err := ExtractPackageSymbols(fset, astFiles, dirRelPath)
	if err != nil {
		return nil, fmt.Errorf("extract package symbols: %w", err)
	}

	pkgModel := &PackageSemanticModel{
		PackagePath: dirRelPath,
		PackageName: st.PackageName,
		SymbolTable: st,
		Files:       make(map[string]*FileSemanticModel),
	}

	for i, astFile := range astFiles {
		relFile := filePaths[i]
		fModel := &FileSemanticModel{
			FilePath: relFile,
			Package:  astFile.Name.Name,
			FileSet:  fset,
			ASTFile:  astFile,
			CFGs:     make(map[string]*ControlFlowGraph),
			Dataflow: make(map[string]*ReachingDefinitionsAnalysis),
		}

		for _, decl := range astFile.Decls {
			fnDecl, ok := decl.(*ast.FuncDecl)
			if !ok || fnDecl.Body == nil {
				continue
			}

			key := fnDecl.Name.Name
			if fnDecl.Recv != nil && len(fnDecl.Recv.List) > 0 {
				recvType := formatTypeExpr(fnDecl.Recv.List[0].Type, fset)
				key = fmt.Sprintf("%s.%s", strings.TrimPrefix(recvType, "*"), fnDecl.Name.Name)
			}

			cfg := BuildCFG(fnDecl, fset, relFile)
			if cfg != nil {
				fModel.CFGs[key] = cfg
				fModel.Dataflow[key] = ComputeReachingDefinitions(cfg, fset)
			}
		}

		pkgModel.Files[relFile] = fModel
	}

	return pkgModel, nil
}
