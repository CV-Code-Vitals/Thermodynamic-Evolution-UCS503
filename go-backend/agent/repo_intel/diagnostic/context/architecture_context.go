package context

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// ArchitectureContextAnalyzer inspects layer boundaries, domain leaks, and mutual type circularities.
type ArchitectureContextAnalyzer struct {
	repoRoot string
}

func NewArchitectureContextAnalyzer(repoRoot string) *ArchitectureContextAnalyzer {
	return &ArchitectureContextAnalyzer{repoRoot: repoRoot}
}

// AnalyzeArchitecturalRules evaluates boundary violations and tight structural cycles.
func (a *ArchitectureContextAnalyzer) AnalyzeArchitecturalRules(files []string) ([]ContextFinding, error) {
	var findings []ContextFinding

	fset := token.NewFileSet()

	// 1. Layer boundary violations (API / Controller importing raw database/sql)
	for _, relPath := range files {
		if !strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, "_test.go") {
			continue
		}

		fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))
		src, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		node, err := parser.ParseFile(fset, fullPath, src, parser.ParseComments)
		if err != nil {
			continue
		}

		isPresentationLayer := strings.Contains(relPath, "api") ||
			strings.Contains(relPath, "handler") ||
			strings.Contains(relPath, "controller") ||
			strings.Contains(relPath, "boundary") ||
			node.Name.Name == "api" || node.Name.Name == "boundary" || node.Name.Name == "handler"

		if isPresentationLayer {
			for _, imp := range node.Imports {
				pathVal := strings.Trim(imp.Path.Value, `"`)
				if pathVal == "database/sql" || strings.Contains(pathVal, "gorm.io") || strings.Contains(pathVal, "sqlx") {
					pos := fset.Position(imp.Pos())
					findings = append(findings, ContextFinding{
						FindingID:           "CTX-ARC-LAYER-BYPASS-" + sanitizeRel(relPath),
						Category:            repo_intel.CategoryArchitecture,
						Severity:            repo_intel.SeverityMedium,
						Confidence:          0.93,
						ConfidenceRationale: fmt.Sprintf("Presentation/API layer file imports %s directly, violating layered architecture", pathVal),
						Title:               "Architectural Layer Boundary Violation (Direct Database Bypass)",
						Description:         fmt.Sprintf("Component %s directly imports database driver %q rather than communicating through repository/domain abstraction.", relPath, pathVal),
						RootCause:           "Presentation/handler layer directly coupled to persistence layer.",
						AffectedFiles:       []string{relPath},
						EvidenceTraces: []EvidenceTrace{
							{
								Location: repo_intel.EvidenceLocation{
									File:        relPath,
									LineStart:   pos.Line,
									LineEnd:     pos.Line,
									Description: fmt.Sprintf("Direct import of %s", pathVal),
								},
								ASTNodeType: "ast.ImportSpec",
								ContextType: "IMPORT_BOUNDARY_LEAK",
								Rationale:   "Higher-level module depends on concrete lower-level data store infrastructure.",
							},
						},
						DetectionMethod:   "AST_IMPORT_BOUNDARY_ANALYSIS",
						RecommendedAction: "Introduce a repository or service interface boundary to decouple presentation logic from direct SQL persistence.",
					})
				}
			}
		}

		// 2. Struct-level mutual recursion / circular coupling
		type StructRef struct {
			Name   string
			Fields map[string]string // fieldName -> typeName
			Line   int
		}

		structs := make(map[string]StructRef)

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, isType := spec.(*ast.TypeSpec)
				if !isType {
					continue
				}

				structType, isStruct := typeSpec.Type.(*ast.StructType)
				if !isStruct {
					continue
				}

				ref := StructRef{
					Name:   typeSpec.Name.Name,
					Fields: make(map[string]string),
					Line:   fset.Position(typeSpec.Pos()).Line,
				}

				for _, field := range structType.Fields.List {
					typeName := extractTypeName(field.Type)
					for _, name := range field.Names {
						ref.Fields[name.Name] = typeName
					}
				}

				structs[typeSpec.Name.Name] = ref
			}
		}

		// Check for mutual references: StructA references StructB and StructB references StructA
		checkedPairs := make(map[string]bool)
		for nameA, structA := range structs {
			for _, targetB := range structA.Fields {
				if structB, ok := structs[targetB]; ok && targetB != nameA {
					// Check if structB references structA
					for _, targetA := range structB.Fields {
						if targetA == nameA {
							pairKey := minString(nameA, targetB) + ":" + maxString(nameA, targetB)
							if !checkedPairs[pairKey] {
								checkedPairs[pairKey] = true
								findings = append(findings, ContextFinding{
									FindingID:           "CTX-ARC-MUTUAL-CYCLE-" + sanitizeRel(relPath) + "-" + nameA + "-" + targetB,
									Category:            repo_intel.CategoryArchitecture,
									Severity:            repo_intel.SeverityMedium,
									Confidence:          0.91,
									ConfidenceRationale: fmt.Sprintf("Mutual circular struct dependency between %s and %s in %s", nameA, targetB, relPath),
									Title:               fmt.Sprintf("Circular Type Coupling (%s <-> %s)", nameA, targetB),
									Description:         fmt.Sprintf("Types %s and %s mutually reference each other directly, forming a tight structural cycle.", nameA, targetB),
									RootCause:           "Unmediated bidirectional dependency between domain service types.",
									AffectedFiles:       []string{relPath},
									EvidenceTraces: []EvidenceTrace{
										{
											Location: repo_intel.EvidenceLocation{
												File:        relPath,
												LineStart:   structA.Line,
												LineEnd:     structA.Line,
												Description: fmt.Sprintf("Struct %s holds reference to %s", nameA, targetB),
											},
											ASTNodeType: "ast.StructType",
											ContextType: "CIRCULAR_FIELD_REF",
											Rationale:   fmt.Sprintf("%s has field pointing to %s", nameA, targetB),
										},
										{
											Location: repo_intel.EvidenceLocation{
												File:        relPath,
												LineStart:   structB.Line,
												LineEnd:     structB.Line,
												Description: fmt.Sprintf("Struct %s holds reference to %s", targetB, nameA),
											},
											ASTNodeType: "ast.StructType",
											ContextType: "CIRCULAR_FIELD_REF",
											Rationale:   fmt.Sprintf("%s has field pointing to %s", targetB, nameA),
										},
									},
									DetectionMethod:   "AST_TYPE_GRAPH_ANALYSIS",
									RecommendedAction: "Break circular reference using interface inversion or mediator pattern.",
								})
							}
						}
					}
				}
			}
		}
	}

	return findings, nil
}

func extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return extractTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

func minString(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func maxString(a, b string) string {
	if a > b {
		return a
	}
	return b
}

