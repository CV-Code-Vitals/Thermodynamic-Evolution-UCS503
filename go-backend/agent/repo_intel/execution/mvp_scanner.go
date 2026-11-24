package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
func (s *MVPService) DiscoverWorkspaceTargets(workspaceID string, ownerID ...string) ([]MVPTarget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.discoverWorkspaceTargetsLocked(workspaceID, ownerID...)
}

func (s *MVPService) discoverWorkspaceTargetsLocked(workspaceID string, ownerID ...string) ([]MVPTarget, error) {
	// Reject malformed workspace identifiers, absolute paths, or path traversal
	if err := ValidateWorkspaceID(workspaceID); err != nil {
		return nil, fmt.Errorf("SECURITY_VIOLATION: %w", err)
	}

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
			return walkErr
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

	// Parse every selected source file before analysis; parser failure is not a clean scan.
	sourceAST := make(map[string]*ast.File)
	sourceSets := make(map[string]*token.FileSet)
	for _, rel := range goFiles {
		content, err := readWorkspaceFile(wsDir, rel)
		if err != nil {
			return nil, err
		}
		fs := token.NewFileSet()
		node, err := parser.ParseFile(fs, rel, content, parser.AllErrors)
		if err != nil {
			return nil, fmt.Errorf("SOURCE_PARSE_FAILED: %s: %w", rel, err)
		}
		sourceAST[rel], sourceSets[rel] = node, fs
	}
	discovered := []MVPTarget{}
	repoName := filepath.Base(wsDir)
	gitValue := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", wsDir}, args...)...)
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	commit, repoURL := gitValue("rev-parse", "HEAD"), gitValue("config", "--get", "remote.origin.url")

	rootDigest, err := ComputeRepoDigest(wsDir)
	if err != nil || rootDigest == "" {
		return nil, fmt.Errorf("SNAPSHOT_DIGEST_FAILED: failed to compute repository digest for workspace %s: %w", workspaceID, err)
	}
	snapshotKey := sha256.Sum256([]byte(workspaceID + "\x00" + rootDigest))
	snapshotID := fmt.Sprintf("SNAP-%x", snapshotKey)
	fileHashes := make(map[string]string)

	// Helper to bind snapshot details to discovered target
	bindTargetSnapshot := func(target *MVPTarget) error {
		target.SnapshotID = snapshotID
		targetRel := filepath.FromSlash(target.TargetFile)
		if filepath.IsAbs(targetRel) {
			var err error
			targetRel, err = filepath.Rel(wsDir, targetRel)
			if err != nil {
				return err
			}
		}
		target.TargetFile = filepath.ToSlash(targetRel)
		fb, fbErr := readWorkspaceFile(wsDir, target.TargetFile)
		if fbErr != nil {
			return fmt.Errorf("SNAPSHOT_SOURCE_HASH_FAILED: cannot read source file %s for target %s: %w", target.TargetFile, target.TargetID, fbErr)
		}
		h := sha256.Sum256(fb)
		target.SourceFileHash = hex.EncodeToString(h[:])
		if target.SourceFileHash == "" {
			return fmt.Errorf("SNAPSHOT_SOURCE_HASH_FAILED: missing source-file hash for target %s", target.TargetID)
		}
		fileHashes[target.TargetFile] = target.SourceFileHash
		return nil
	}

	// 2. Run Unified Semantic Diagnostic Orchestrator if Go files are present
	if len(goFiles) > 0 {
		orch := semantic.NewUnifiedOrchestrator(wsDir)
		report, diagErr := orch.RunCompleteDiagnostic(goFiles, commit)
		if diagErr != nil {
			return nil, fmt.Errorf("SCANNER_FAILED: %w", diagErr)
		}
		if report == nil || report.ReconciliationLedger == nil {
			return nil, fmt.Errorf("SCANNER_FAILED: missing diagnostic report")
		}
		if report != nil && report.ReconciliationLedger != nil {
			sort.SliceStable(report.ReconciliationLedger.Findings, func(i, j int) bool {
				a, b := report.ReconciliationLedger.Findings[i], report.ReconciliationLedger.Findings[j]
				if a.TargetFile != b.TargetFile {
					return a.TargetFile < b.TargetFile
				}
				if a.LineStart != b.LineStart {
					return a.LineStart < b.LineStart
				}
				if a.RuleID != b.RuleID {
					return a.RuleID < b.RuleID
				}
				return a.Description < b.Description
			})
			for i, f := range report.ReconciliationLedger.Findings {
				targetID := fmt.Sprintf("SCAN-%s-%d", f.RuleID, i+1)
				expectedAction := "REPAIR"
				if f.RemediationEligibility == semantic.EligibleQuarantined ||
					f.Category == repo_intel.CategoryArchitecture ||
					strings.Contains(f.RuleID, "CMD") ||
					strings.Contains(f.RuleID, "COMMAND") {
					expectedAction = "ABSTAIN"
				}

				categoryStr := "UNSUPPORTED"
				if strings.Contains(f.RuleID, "SQL") {
					categoryStr = "SQL_INJECTION"
				} else if strings.Contains(f.RuleID, "CMD") || strings.Contains(f.RuleID, "COMMAND") {
					categoryStr = "COMMAND_INJECTION"
				} else if strings.Contains(f.RuleID, "RES") {
					categoryStr = "RESOURCE_LEAK"
				} else if strings.Contains(f.RuleID, "ERR") {
					categoryStr = "UNHANDLED_ERROR"
				}

				if strings.Contains(f.Description, "time.Timer") || strings.Contains(f.Title, "time.Timer") {
					categoryStr = "UNSUPPORTED"
					f.ApprovalRationale = "Timer lifecycle requires timer-channel and Go-version semantics; not a supported Close repair."
				}
				if categoryStr == "UNHANDLED_ERROR" || categoryStr == "UNSUPPORTED" || f.RemediationEligibility == semantic.EligibleBenign {
					expectedAction = "ABSTAIN"
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
					UpstreamCommit: commit, UpstreamURL: repoURL,
					RuleID: f.RuleID, LineEnd: f.LineEnd, Confidence: f.Confidence,
					UncertaintyClass: f.UncertaintyClass, DetectionEvidence: f.Description,
					EligibilityReason: f.ApprovalRationale,
				}
				if err := bindTargetSnapshot(&target); err != nil {
					return nil, err
				}
				node, fs := sourceAST[target.TargetFile], sourceSets[target.TargetFile]
				if node == nil || fs == nil {
					return nil, fmt.Errorf("INVALID_SOURCE_EVIDENCE: missing parsed file")
				}
				for _, decl := range node.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && fs.Position(fn.Pos()).Line <= target.LineNumber && fs.Position(fn.End()).Line >= target.LineNumber {
						target.Symbol = fn.Name.Name
						break
					}
				}
				content, err := readWorkspaceFile(wsDir, target.TargetFile)
				if err != nil {
					return nil, err
				}
				lines := strings.Split(string(content), "\n")
				if target.LineNumber < 1 || target.LineNumber > len(lines) || target.Symbol == "" {
					return nil, fmt.Errorf("INVALID_SOURCE_EVIDENCE: %s:%d", target.TargetFile, target.LineNumber)
				}
				target.SourceEvidence = lines[target.LineNumber-1]
				discovered = append(discovered, target)
			}
		}
	}

	// Repository-supplied reports and broad AST patterns are not diagnostic evidence.
	// Only findings from the semantic pass above are returned in live mode.

	owner := ""
	if len(ownerID) > 0 && ownerID[0] != "" {
		owner = ownerID[0]
	}
	now := time.Now().UTC()

	if s.snapshots == nil {
		s.snapshots = make(map[string]*WorkspaceSnapshot)
	}
	s.snapshots[snapshotID] = &WorkspaceSnapshot{
		SnapshotID:   snapshotID,
		WorkspaceID:  workspaceID,
		OwnerID:      owner,
		SourceCommit: commit,
		RootDigest:   rootDigest,
		FileHashes:   fileHashes,
		TargetCount:  len(discovered),
		CreatedAt:    now.Format(time.RFC3339),
		ExpiresAt:    now.Add(60 * time.Minute).Format(time.RFC3339),
	}

	for i := range discovered {
		// Keep earlier scans and approval references bound to their own immutable snapshot.
		discovered[i].TargetID = fmt.Sprintf("%s-%x-%d", discovered[i].TargetID, snapshotKey[:12], i)
		s.dynamicTargets[discovered[i].TargetID] = &discovered[i]
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
