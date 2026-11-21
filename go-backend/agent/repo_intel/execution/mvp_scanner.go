package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/diagnostic/semantic"
)

// ScannedHotspot represents an AST hotspot from thermodynamic_report.json
type ScannedHotspot struct {
	FunctionName         string  `json:"function_name"`
	ByteRange            [2]int  `json:"byte_range"`
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	EntropyScore         float64 `json:"entropy_score"`
	CognitiveDepth       int     `json:"cognitive_depth"`
}

// ScannedFileReport represents a file analysis in thermodynamic_report.json
type ScannedFileReport struct {
	FilePath string           `json:"file_path"`
	Language string           `json:"language"`
	Hotspots []ScannedHotspot `json:"hotspots"`
}

// ScannedThermodynamicReport represents the output produced by the AST engine
type ScannedThermodynamicReport struct {
	FilesAnalyzed int                 `json:"files_analyzed"`
	TotalHotspots int                 `json:"total_hotspots"`
	FileReports   []ScannedFileReport `json:"file_reports"`
}

// DiscoverWorkspaceTargets scans the target repository workspace for diagnostic findings.
func (s *MVPService) DiscoverWorkspaceTargets(workspaceID string) ([]MVPTarget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.discoverWorkspaceTargetsLocked(workspaceID)
}

func (s *MVPService) discoverWorkspaceTargetsLocked(workspaceID string) ([]MVPTarget, error) {
	wsDir, err := s.resolveWorkspaceDirLocked(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace %s: %w", workspaceID, err)
	}

	fi, err := os.Stat(wsDir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("workspace directory %s does not exist", wsDir)
	}

	// 1. Collect all candidate Go files in workspace (ignoring .git, vendor, .tmp, etc.)
	var goFiles []string
	err = filepath.Walk(wsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" || base == ".tmp" || base == "node_modules" || base == "uploads" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			rel, relErr := filepath.Rel(wsDir, path)
			if relErr == nil {
				goFiles = append(goFiles, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan workspace files: %w", err)
	}

	var discovered []MVPTarget
	repoName := filepath.Base(wsDir)

	rootDigest, _ := ComputeRepoDigest(wsDir)
	snapshotID := "SNAP-EMPTY"
	if len(rootDigest) >= 16 {
		snapshotID = fmt.Sprintf("SNAP-%s", rootDigest[:16])
	}
	fileHashes := make(map[string]string)

	// Helper to bind snapshot details to discovered target
	bindTargetSnapshot := func(target *MVPTarget) {
		target.SnapshotID = snapshotID
		targetAbs := filepath.Join(wsDir, filepath.FromSlash(target.TargetFile))
		if fb, fbErr := os.ReadFile(targetAbs); fbErr == nil {
			h := sha256.Sum256(fb)
			target.SourceFileHash = hex.EncodeToString(h[:])
			fileHashes[target.TargetFile] = target.SourceFileHash
		}
	}

	// 2. Run Unified Semantic Diagnostic Orchestrator if Go files are present
	if len(goFiles) > 0 {
		orch := semantic.NewUnifiedOrchestrator(wsDir)
		report, diagErr := orch.RunCompleteDiagnostic(goFiles, "HEAD")
		if diagErr == nil && report != nil && report.ReconciliationLedger != nil {
			for i, f := range report.ReconciliationLedger.Findings {
				targetID := fmt.Sprintf("SCAN-%s-%d", f.RuleID, i+1)
				expectedAction := "REPAIR"
				if f.RemediationEligibility == semantic.EligibleQuarantined ||
					f.Category == repo_intel.CategoryArchitecture ||
					strings.Contains(f.RuleID, "CMD") ||
					strings.Contains(f.RuleID, "COMMAND") {
					expectedAction = "ABSTAIN"
				}

				categoryStr := "RESOURCE_LEAK"
				if strings.Contains(f.RuleID, "SQL") {
					categoryStr = "SQL_INJECTION"
				} else if strings.Contains(f.RuleID, "CMD") || strings.Contains(f.RuleID, "COMMAND") {
					categoryStr = "COMMAND_INJECTION"
				} else if strings.Contains(f.RuleID, "ERR") {
					categoryStr = "UNHANDLED_ERROR"
				}

				target := MVPTarget{
					TargetID:       targetID,
					Repository:     repoName,
					RelativeSource: f.TargetFile,
					TargetFile:     f.TargetFile,
					DefectCategory: categoryStr,
					Stratum:        "SCANNED_REPOSITORY",
					ExpectedAction: expectedAction,
					Description:    fmt.Sprintf("%s in %s:%d", f.Title, f.TargetFile, f.LineStart),
					WorkspaceID:    workspaceID,
					LineNumber:     f.LineStart,
					IsDemoTarget:   false,
					VerifiedBy:     "Unified Semantic Diagnostic Engine",
				}
				bindTargetSnapshot(&target)
				discovered = append(discovered, target)
				s.dynamicTargets[targetID] = &target
			}
		}
	}

	// 3. Inspect thermodynamic_report.json for high-complexity hotspots
	reportPath := filepath.Join(wsDir, "thermodynamic_report.json")
	if reportBytes, err := os.ReadFile(reportPath); err == nil {
		var thermoReport ScannedThermodynamicReport
		if err := json.Unmarshal(reportBytes, &thermoReport); err == nil {
			for _, file := range thermoReport.FileReports {
				if file.Language != "Go" {
					continue
				}
				for _, hotspot := range file.Hotspots {
					if hotspot.CyclomaticComplexity >= 15 || hotspot.EntropyScore >= 75 {
						targetID := fmt.Sprintf("SCAN-COMPLEXITY-%s", sanitizeTargetID(hotspot.FunctionName))
						target := MVPTarget{
							TargetID:       targetID,
							Repository:     repoName,
							RelativeSource: file.FilePath,
							TargetFile:     file.FilePath,
							DefectCategory: "HIGH_COMPLEXITY",
							Stratum:        "SCANNED_REPOSITORY",
							ExpectedAction: "ABSTAIN",
							Description:    fmt.Sprintf("High Complexity Hotspot: %s() in %s (Complexity: %d, Entropy: %.1f)", hotspot.FunctionName, file.FilePath, hotspot.CyclomaticComplexity, hotspot.EntropyScore),
							WorkspaceID:    workspaceID,
							LineNumber:     1,
							IsDemoTarget:   false,
							VerifiedBy:     "Thermodynamic AST Hotspot Engine",
						}
						bindTargetSnapshot(&target)
						discovered = append(discovered, target)
						s.dynamicTargets[targetID] = &target
					}
				}
			}
		}
	}

	// 4. Fallback AST heuristic scan if no semantic findings were identified
	if len(discovered) == 0 && len(goFiles) > 0 {
		fallbackTargets := s.astFallbackScan(wsDir, goFiles, workspaceID, repoName)
		for i := range fallbackTargets {
			bindTargetSnapshot(&fallbackTargets[i])
		}
		discovered = append(discovered, fallbackTargets...)
	}

	if s.snapshots == nil {
		s.snapshots = make(map[string]*WorkspaceSnapshot)
	}
	s.snapshots[snapshotID] = &WorkspaceSnapshot{
		SnapshotID:   snapshotID,
		WorkspaceID:  workspaceID,
		SourceCommit: "HEAD",
		RootDigest:   rootDigest,
		FileHashes:   fileHashes,
		TargetCount:  len(discovered),
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	return discovered, nil
}

// astFallbackScan performs basic AST inspection for common patterns in Go source files
func (s *MVPService) astFallbackScan(wsDir string, goFiles []string, workspaceID, repoName string) []MVPTarget {
	var results []MVPTarget
	fset := token.NewFileSet()

	for _, relPath := range goFiles {
		absPath := filepath.Join(wsDir, filepath.FromSlash(relPath))
		node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		ast.Inspect(node, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}

			// Check for unclosed os.Open / os.Create without defer Close
			for _, stmt := range fn.Body.List {
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok || len(assign.Rhs) == 0 {
					continue
				}
				call, ok := assign.Rhs[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					continue
				}

				if ident.Name == "os" && (sel.Sel.Name == "Open" || sel.Sel.Name == "Create") {
					line := fset.Position(assign.Pos()).Line
					targetID := fmt.Sprintf("SCAN-AST-RES-%s-%d", sanitizeTargetID(fn.Name.Name), line)
					target := MVPTarget{
						TargetID:       targetID,
						Repository:     repoName,
						RelativeSource: relPath,
						TargetFile:     relPath,
						DefectCategory: "RESOURCE_LEAK",
						Stratum:        "SCANNED_REPOSITORY",
						ExpectedAction: "REPAIR",
						Description:    fmt.Sprintf("Resource handle os.%s() in %s():%d requires lifecycle management", sel.Sel.Name, fn.Name.Name, line),
						WorkspaceID:    workspaceID,
						LineNumber:     line,
						IsDemoTarget:   false,
						VerifiedBy:     "AST Structural Pattern Engine",
					}
					results = append(results, target)
					s.dynamicTargets[targetID] = &target
				}

				// Check for exec.Command with "sh", "-c"
				if ident.Name == "exec" && sel.Sel.Name == "Command" && len(call.Args) >= 2 {
					if arg0, ok := call.Args[0].(*ast.BasicLit); ok && arg0.Value == `"sh"` {
						line := fset.Position(assign.Pos()).Line
						targetID := fmt.Sprintf("SCAN-AST-CMD-%s-%d", sanitizeTargetID(fn.Name.Name), line)
						target := MVPTarget{
							TargetID:       targetID,
							Repository:     repoName,
							RelativeSource: relPath,
							TargetFile:     relPath,
							DefectCategory: "COMMAND_INJECTION",
							Stratum:        "SCANNED_REPOSITORY",
							ExpectedAction: "ABSTAIN",
							Description:    fmt.Sprintf("Arbitrary shell execution via sh -c in %s():%d requires architectural redesign", fn.Name.Name, line),
							WorkspaceID:    workspaceID,
							LineNumber:     line,
							IsDemoTarget:   false,
							VerifiedBy:     "AST Security Boundary Engine",
						}
						results = append(results, target)
						s.dynamicTargets[targetID] = &target
					}
				}
			}
			return true
		})
	}
	return results
}

// checkResourceEscapes inspects the function enclosing lineNo in targetFile to determine
// if an allocated resource handle is returned to the caller or escapes function scope.
func checkResourceEscapes(wsDir, targetFile string, lineNo int) bool {
	absPath := filepath.Join(wsDir, filepath.FromSlash(targetFile))
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return true // Fail-closed: assume escape if unparseable
	}

	escapes := false
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}

		startLine := fset.Position(fn.Pos()).Line
		endLine := fset.Position(fn.End()).Line
		if lineNo < startLine || lineNo > endLine {
			return true
		}

		// Check return types for io.Closer, io.ReadCloser, *os.File, net.Conn, *http.Response
		if fn.Type.Results != nil {
			for _, field := range fn.Type.Results.List {
				typeStr := ""
				switch t := field.Type.(type) {
				case *ast.StarExpr:
					if sel, ok := t.X.(*ast.SelectorExpr); ok {
						typeStr = sel.Sel.Name
					} else if ident, ok := t.X.(*ast.Ident); ok {
						typeStr = ident.Name
					}
				case *ast.SelectorExpr:
					typeStr = t.Sel.Name
				case *ast.Ident:
					typeStr = t.Name
				}

				if typeStr == "File" || typeStr == "ReadCloser" || typeStr == "Conn" || typeStr == "Response" || typeStr == "Closer" {
					escapes = true
					return false
				}
			}
		}

		// Check for return statements in body that return an identifier
		ast.Inspect(fn.Body, func(bn ast.Node) bool {
			ret, ok := bn.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			for _, res := range ret.Results {
				if ident, ok := res.(*ast.Ident); ok {
					if ident.Name != "nil" && ident.Name != "err" {
						escapes = true
						return false
					}
				}
			}
			return true
		})

		return false
	})

	return escapes
}

// extractSourceSnippet extracts lines surrounding lineNo with 1-based line markers.
func extractSourceSnippet(content string, lineNo int) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return "// Empty file"
	}

	start := lineNo - 4
	if start < 0 {
		start = 0
	}
	end := lineNo + 3
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	for i := start; i < end; i++ {
		currLine := i + 1
		marker := "  "
		if currLine == lineNo {
			marker = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%4d | %s\n", marker, currLine, lines[i]))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func sanitizeTargetID(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, " ", "_")
	if len(s) > 20 {
		return s[:20]
	}
	return s
}
