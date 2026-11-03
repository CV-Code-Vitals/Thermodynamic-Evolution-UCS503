package context

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// StructuralContextAnalyzer detects AST-level code clones and duplicated function bodies.
type StructuralContextAnalyzer struct {
	repoRoot string
}

func NewStructuralContextAnalyzer(repoRoot string) *StructuralContextAnalyzer {
	return &StructuralContextAnalyzer{repoRoot: repoRoot}
}

type FuncSignature struct {
	Name       string
	File       string
	StartLine  int
	EndLine    int
	BodyHash   string
	StmtCount  int
}

// AnalyzeDuplicates inspects a set of Go source files and finds identical function bodies.
func (a *StructuralContextAnalyzer) AnalyzeDuplicates(files []string) ([]ContextFinding, error) {
	var findings []ContextFinding
	bodyHashMap := make(map[string][]FuncSignature)

	fset := token.NewFileSet()

	for _, relPath := range files {
		if !strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, "_test.go") {
			continue
		}

		fullPath := filepath.Join(a.repoRoot, filepath.FromSlash(relPath))
		src, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		node, err := parser.ParseFile(fset, fullPath, src, 0)
		if err != nil {
			continue
		}

		for _, decl := range node.Decls {
			fn, isFn := decl.(*ast.FuncDecl)
			if !isFn || fn.Body == nil || len(fn.Body.List) < 3 {
				// Require at least 3 statements to avoid trivial one-liners
				continue
			}

			startPos := fset.Position(fn.Pos())
			endPos := fset.Position(fn.End())

			// Normalize AST body by printing statement list without parameter/local name variance
			var buf strings.Builder
			for _, stmt := range fn.Body.List {
				printer.Fprint(&buf, fset, stmt)
				buf.WriteString("\n")
			}

			normalized := buf.String()
			h := sha256.Sum256([]byte(normalized))
			hashStr := hex.EncodeToString(h[:])

			sig := FuncSignature{
				Name:      fn.Name.Name,
				File:      relPath,
				StartLine: startPos.Line,
				EndLine:   endPos.Line,
				BodyHash:  hashStr,
				StmtCount: len(fn.Body.List),
			}

			bodyHashMap[hashStr] = append(bodyHashMap[hashStr], sig)
		}
	}

	for hashStr, sigs := range bodyHashMap {
		if len(sigs) >= 2 {
			// Found duplicate clone!
			primary := sigs[0]
			var affectedFiles []string
			var traces []EvidenceTrace
			var names []string

			for _, s := range sigs {
				affectedFiles = append(affectedFiles, s.File)
				names = append(names, s.Name)
				traces = append(traces, EvidenceTrace{
					Location: repo_intel.EvidenceLocation{
						File:        s.File,
						LineStart:   s.StartLine,
						LineEnd:     s.EndLine,
						Description: fmt.Sprintf("Function %s (%d statements)", s.Name, s.StmtCount),
					},
					ASTNodeType: "ast.FuncDecl",
					ContextType: "CLONE_INSTANCE",
					Rationale:   fmt.Sprintf("Identical normalized AST statement block (hash: %s)", hashStr[:8]),
				})
			}

			findingID := fmt.Sprintf("CTX-MNT-DUPLICATE-FN-%s", sanitizeRel(primary.File))
			findings = append(findings, ContextFinding{
				FindingID:           findingID,
				Category:            repo_intel.CategoryMaintainability,
				Severity:            repo_intel.SeverityLow,
				Confidence:          0.94,
				ConfidenceRationale: fmt.Sprintf("AST exact statement clone detected across %d functions (%s)", len(sigs), strings.Join(names, ", ")),
				Title:               fmt.Sprintf("Duplicated Function Implementation (%s)", strings.Join(names, ", ")),
				Description:         fmt.Sprintf("Functions %s share identical implementation logic of %d statements.", strings.Join(names, ", "), primary.StmtCount),
				RootCause:           "Copy-pasting helper functions instead of extracting shared reusable package utility.",
				AffectedFiles:       uniqueRelFiles(affectedFiles),
				EvidenceTraces:      traces,
				DetectionMethod:     "AST_BODY_HASHING",
				RecommendedAction:   "Extract the duplicate implementation into a single shared helper function and replace duplicates with calls to it.",
			})
		}
	}

	return findings, nil
}

func uniqueRelFiles(files []string) []string {
	seen := make(map[string]bool)
	var res []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			res = append(res, f)
		}
	}
	return res
}

