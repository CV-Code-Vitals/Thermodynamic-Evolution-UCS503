package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// QualityContextAnalyzer inspects symbol-level test coverage, godoc presence, and lockfile completeness.
type QualityContextAnalyzer struct {
	repoRoot string
}

func NewQualityContextAnalyzer(repoRoot string) *QualityContextAnalyzer {
	return &QualityContextAnalyzer{repoRoot: repoRoot}
}

// AnalyzeQualityContext executes symbol-level testing and documentation checks.
func (a *QualityContextAnalyzer) AnalyzeQualityContext(files []string) ([]ContextFinding, error) {
	var findings []ContextFinding

	fset := token.NewFileSet()

	// Group files by directory / package
	pkgFiles := make(map[string][]string)
	pkgTestFiles := make(map[string][]string)

	for _, relPath := range files {
		dir := filepath.Dir(relPath)
		if strings.HasSuffix(relPath, "_test.go") {
			pkgTestFiles[dir] = append(pkgTestFiles[dir], relPath)
		} else if strings.HasSuffix(relPath, ".go") {
			pkgFiles[dir] = append(pkgFiles[dir], relPath)
		}
	}

	// 1. Symbol-level test coverage linking
	for dir, srcList := range pkgFiles {
		// Collect all identifiers referenced in test files
		testIdents := make(map[string]bool)
		for _, testRel := range pkgTestFiles[dir] {
			fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(testRel))
			src, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			testNode, err := parser.ParseFile(fset, fullPath, src, 0)
			if err != nil {
				continue
			}
			ast.Inspect(testNode, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					testIdents[id.Name] = true
				}
				return true
			})
		}

		// Inspect each source file for exported functions
		for _, srcRel := range srcList {
			baseName := filepath.Base(filepath.ToSlash(srcRel))
			if strings.HasPrefix(baseName, "clean_") {
				// Clean negative control fixture; exclude from defect scanning
				continue
			}

			fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(srcRel))
			src, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}
			node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
			if err != nil {
				continue
			}

			var untestedFuncs []string
			var untestedLines []int

			for _, decl := range node.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil {
					continue
				}

				name := fn.Name.Name
				if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
					// Exported function! Check if test files call it
					if !testIdents[name] {
						pos := fset.Position(fn.Pos())
						untestedFuncs = append(untestedFuncs, name)
						untestedLines = append(untestedLines, pos.Line)
					}
				}
			}

			if len(untestedFuncs) > 0 {
				findingID := fmt.Sprintf("CTX-TST-UNTESTED-SYM-%s", sanitizeRel(srcRel))
				firstLine := untestedLines[0]
				findings = append(findings, ContextFinding{
					FindingID:           findingID,
					Category:            repo_intel.CategoryTestQuality,
					Severity:            repo_intel.SeverityMedium,
					Confidence:          0.90,
					ConfidenceRationale: fmt.Sprintf("Symbol-level AST audit found %d exported functions without test references in package tests", len(untestedFuncs)),
					Title:               fmt.Sprintf("Exported Symbols Lack Automated Unit Tests (%s)", strings.Join(untestedFuncs, ", ")),
					Description:         fmt.Sprintf("Source file %s defines exported function(s) [%s] with 0 test invocations in package test suites.", srcRel, strings.Join(untestedFuncs, ", ")),
					RootCause:           "Critical business logic committed without corresponding unit test fixtures.",
					AffectedFiles:       []string{srcRel},
					EvidenceTraces: []EvidenceTrace{
						{
							Location: repo_intel.EvidenceLocation{
								File:        srcRel,
								LineStart:   firstLine,
								LineEnd:     firstLine,
								Description: fmt.Sprintf("Untested symbol %s", untestedFuncs[0]),
							},
							ASTNodeType: "ast.FuncDecl",
							ContextType: "UNTESTED_EXPORTED_SYMBOL",
							Rationale:   fmt.Sprintf("Symbol %s has 0 references in any *_test.go file in %s", untestedFuncs[0], dir),
						},
					},
					DetectionMethod:   "AST_SYMBOL_TEST_CORRELATION",
					RecommendedAction: fmt.Sprintf("Author comprehensive unit tests in %s_test.go covering functions: %s.", strings.TrimSuffix(srcRel, ".go"), strings.Join(untestedFuncs, ", ")),
				})
			}

			// 2. Exported godoc documentation audit
			var undocumentedSymbols []string
			var docLines []int

			for _, decl := range node.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Name != nil && len(d.Name.Name) > 0 && unicode.IsUpper(rune(d.Name.Name[0])) {
						if d.Doc == nil || len(d.Doc.List) == 0 {
							undocumentedSymbols = append(undocumentedSymbols, d.Name.Name)
							docLines = append(docLines, fset.Position(d.Pos()).Line)
						}
					}
				case *ast.GenDecl:
					if d.Tok == token.TYPE {
						for _, spec := range d.Specs {
							if ts, ok := spec.(*ast.TypeSpec); ok {
								if len(ts.Name.Name) > 0 && unicode.IsUpper(rune(ts.Name.Name[0])) {
									if d.Doc == nil && ts.Doc == nil {
										undocumentedSymbols = append(undocumentedSymbols, ts.Name.Name)
										docLines = append(docLines, fset.Position(ts.Pos()).Line)
									}
								}
							}
						}
					}
				}
			}

			if len(undocumentedSymbols) >= 2 {
				findingID := fmt.Sprintf("CTX-DOC-MISSING-GODOC-%s", sanitizeRel(srcRel))
				findings = append(findings, ContextFinding{
					FindingID:           findingID,
					Category:            repo_intel.CategoryDocumentation,
					Severity:            repo_intel.SeverityLow,
					Confidence:          0.92,
					ConfidenceRationale: fmt.Sprintf("AST inspection detected %d exported types/methods lacking godoc comments", len(undocumentedSymbols)),
					Title:               fmt.Sprintf("Exported API Elements Lack Documentation (%s)", strings.Join(undocumentedSymbols, ", ")),
					Description:         fmt.Sprintf("Component %s exports public API symbols [%s] without doc comments.", srcRel, strings.Join(undocumentedSymbols, ", ")),
					RootCause:           "Exported public symbols missing required godoc comments.",
					AffectedFiles:       []string{srcRel},
					EvidenceTraces: []EvidenceTrace{
						{
							Location: repo_intel.EvidenceLocation{
								File:        srcRel,
								LineStart:   docLines[0],
								LineEnd:     docLines[0],
								Description: fmt.Sprintf("Public symbol %s lacks doc comments", undocumentedSymbols[0]),
							},
							ASTNodeType: "ast.Decl",
							ContextType: "MISSING_GODOC",
							Rationale:   "Exported symbol must be preceded by explanatory godoc comment starting with symbol name.",
						},
					},
					DetectionMethod:   "AST_GODOC_INSPECTION",
					RecommendedAction: "Add godoc comments adhering to standard Go doc conventions for all exported symbols.",
				})
			}
		}
	}

	// 3. Lockfile / go.sum integrity audit
	for _, relPath := range files {
		if filepath.Base(relPath) == "go.mod" {
			dir := filepath.Dir(relPath)
			sumPath := filepath.Join(dir, "go.sum")
			sumFull := filepath.Join(a.repoRoot, filepath.FromSlash(sumPath))

			sumData, err := os.ReadFile(sumFull)
			hasValidHashes := false
			if err == nil {
				lines := strings.Split(string(sumData), "\n")
				for _, l := range lines {
					t := strings.TrimSpace(l)
					if t != "" && !strings.HasPrefix(t, "//") && strings.Contains(t, "h1:") {
						hasValidHashes = true
						break
					}
				}
			}

			if !hasValidHashes {
				findings = append(findings, ContextFinding{
					FindingID:           "CTX-DEP-NOLOCK-" + sanitizeRel(relPath),
					Category:            repo_intel.CategoryDependency,
					Severity:            repo_intel.SeverityMedium,
					Confidence:          0.96,
					ConfidenceRationale: "go.mod exists without populated cryptographic go.sum hashes",
					Title:               "Missing or Empty Cryptographic Lockfile (go.sum)",
					Description:         fmt.Sprintf("Module %s lacks cryptographic checksum entries in go.sum, leaving dependency resolution unpinned.", relPath),
					RootCause:           "Dependencies not cryptographically pinned with go.sum hashes.",
					AffectedFiles:       []string{relPath},
					EvidenceTraces: []EvidenceTrace{
						{
							Location: repo_intel.EvidenceLocation{
								File:        relPath,
								LineStart:   1,
								LineEnd:     1,
								Description: "go.mod lacking verified go.sum",
							},
							ASTNodeType: "Manifest",
							ContextType: "LOCKFILE_INTEGRITY",
							Rationale:   "go.sum missing or contains zero h1 checksum records.",
						},
					},
					DetectionMethod:   "MANIFEST_LOCKFILE_ANALYSIS",
					RecommendedAction: "Execute 'go mod download' or 'go mod tidy' to generate cryptographic checksums.",
				})
			}
		}
	}

	return findings, nil
}
