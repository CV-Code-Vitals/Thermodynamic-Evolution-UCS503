package benchmark

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// GenerateBenchmarkSuiteV2 generates the 96-fixture expanded benchmark suite v2 and its ground truth manifest.
func GenerateBenchmarkSuiteV2(baseDir string, manifestPath string) (*GroundTruthManifest, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create baseDir: %w", err)
	}

	manifest := &GroundTruthManifest{
		BenchmarkName:  "Self-Healing Autonomous Benchmark Suite v2 (Expanded Multi-Category)",
		Version:        "2.0.0",
		CreatedAt:      time.Now(),
		TargetRepoPath: filepath.ToSlash(baseDir),
		DefectFixtures: make([]DefectFixture, 0),
		CleanFixtures:  make([]CleanFixture, 0),
	}

	writeFile := func(relPath, content string) (string, error) {
		full := filepath.Join(baseDir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			return "", err
		}
		h := sha256.Sum256([]byte(content))
		return hex.EncodeToString(h[:]), nil
	}

	// Root go.mod
	_, err := writeFile("go.mod", `module github.com/saksham/self-healing-benchmark-v2

go 1.25.6
`)
	if err != nil {
		return nil, err
	}

	// Root README.md
	_, _ = writeFile("README.md", "# Self-Healing Suite v2\nExpanded 96-fixture ground truth benchmark.\n")

	// 1. Structural (8 defects, 4 clean)
	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("structural/nesting_%d.go", i)
		fnName := fmt.Sprintf("ProcessDeepNesting%d", i)
		code := fmt.Sprintf(`package structural

func %s(data map[string]interface{}) int {
	res := 0
	if data != nil {
		if val, ok := data["key"]; ok {
			if list, isList := val.([]int); isList {
				for _, item := range list {
					if item > 10 {
						if item%%2 == 0 {
							res += item
						}
					}
				}
			}
		}
	}
	return res
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-STR-NEST-%02d", i),
			Category:                           repo_intel.CategoryStructural,
			DefectDescription:                  "Excessive control flow nesting depth exceeds threshold",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 3, EndLine: 20},
			ExpectedDiagnosticClassification:   "STRUCTURAL",
			ExpectedSeverity:                   repo_intel.SeverityMedium,
			ExpectedConfidenceRange:            [2]float64{0.85, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:          "AST_TRANSFORMATION",
			ExpectedApprovalRequirements:       false,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("structural/complexity_%d.go", i)
		fnName := fmt.Sprintf("ComputeComplexDecision%d", i)
		code := fmt.Sprintf(`package structural

func %s(code int, mode string, active bool) int {
	score := 0
	if code == 1 { score += 10 } else if code == 2 { score += 20 } else if code == 3 { score += 30 }
	if mode == "A" { score += 5 } else if mode == "B" { score += 15 } else if mode == "C" { score += 25 }
	if active { score *= 2 } else { score /= 2 }
	if score > 50 { score = 50 } else if score < 0 { score = 0 }
	if code > 10 { score++ } else { score-- }
	if mode == "Z" { score += 100 }
	return score
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-STR-COMPLEX-%02d", i),
			Category:                           repo_intel.CategoryStructural,
			DefectDescription:                  "High cyclomatic complexity exceeding threshold",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 3, EndLine: 15},
			ExpectedDiagnosticClassification:   "STRUCTURAL",
			ExpectedSeverity:                   repo_intel.SeverityMedium,
			ExpectedConfidenceRange:            [2]float64{0.85, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:          "AST_TRANSFORMATION",
			ExpectedApprovalRequirements:       false,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("structural/clean_nesting_%d.go", i)
		code := fmt.Sprintf(`package structural

func SafeCleanNesting%d(data map[string]int) int {
	if data == nil { return 0 }
	val, ok := data["score"]
	if !ok || val <= 0 { return 0 }
	return val * 2
}
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-STR-NEST-%02d", i),
			Category:     repo_intel.CategoryStructural,
			Description:  "Clean control: Guard clauses prevent nesting",
			TargetFile:   rel,
			TargetSymbol: fmt.Sprintf("SafeCleanNesting%d", i),
			FileSHA256:   hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("structural/clean_complexity_%d.go", i)
		code := fmt.Sprintf(`package structural

func SafeCleanComplexity%d(x, y int) int {
	if x > y { return x - y }
	return y - x
}
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-STR-COMPLEX-%02d", i),
			Category:     repo_intel.CategoryStructural,
			Description:  "Clean control: Low cyclomatic complexity function",
			TargetFile:   rel,
			TargetSymbol: fmt.Sprintf("SafeCleanComplexity%d", i),
			FileSHA256:   hash,
		})
	}

	// 2. Reliability (8 defects, 4 clean)
	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("reliability/error_handling_%d.go", i)
		fnName := fmt.Sprintf("PerformOperationUnchecked%d", i)
		code := fmt.Sprintf(`package reliability

import "os"

func %s(path string) {
	// Ignored error return
	_ = os.Remove(path)
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-REL-ERR-%02d", i),
			Category:                           repo_intel.CategoryReliability,
			DefectDescription:                  "Ignored error return value allows silent failure",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 6, EndLine: 9},
			ExpectedDiagnosticClassification:   "RELIABILITY",
			ExpectedSeverity:                   repo_intel.SeverityLow,
			ExpectedConfidenceRange:            [2]float64{0.80, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:          "AST_TRANSFORMATION",
			ExpectedApprovalRequirements:       false,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("reliability/resource_leak_%d.go", i)
		fnName := fmt.Sprintf("ReadFilesLeaking%d", i)
		code := fmt.Sprintf(`package reliability

import (
	"io"
	"os"
)

func %s(files []string) ([]int64, error) {
	sizes := make([]int64, len(files))
	for i, f := range files {
		handle, err := os.Open(f)
		if err != nil { return nil, err }
		sz, _ := handle.Seek(0, io.SeekEnd)
		sizes[i] = sz
	}
	return sizes, nil
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-REL-LEAK-%02d", i),
			Category:                           repo_intel.CategoryReliability,
			DefectDescription:                  "Unclosed resource handle opened inside loop body",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 9, EndLine: 18},
			ExpectedDiagnosticClassification:   "RELIABILITY",
			ExpectedSeverity:                   repo_intel.SeverityHigh,
			ExpectedConfidenceRange:            [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy:          "RULE_BASED",
			ExpectedApprovalRequirements:       false,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("reliability/clean_error_%d.go", i)
		code := fmt.Sprintf(`package reliability

import "os"

func SafeErrorCheck%d(p string) error {
	if err := os.Remove(p); err != nil {
		return err
	}
	return nil
}
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-REL-ERR-%02d", i),
			Category:     repo_intel.CategoryReliability,
			Description:  "Clean control: Proper error propagation",
			TargetFile:   rel,
			TargetSymbol: fmt.Sprintf("SafeErrorCheck%d", i),
			FileSHA256:   hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("reliability/clean_resource_%d.go", i)
		code := fmt.Sprintf(`package reliability

import "os"

func SafeResourceCleanup%d(p string) ([]byte, error) {
	data, err := os.ReadFile(p)
	if err != nil { return nil, err }
	return data, nil
}
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID:    fmt.Sprintf("CLEAN-REL-RES-%02d", i),
			Category:     repo_intel.CategoryReliability,
			Description:  "Clean control: Safe scoped resource read",
			TargetFile:   rel,
			TargetSymbol: fmt.Sprintf("SafeResourceCleanup%d", i),
			FileSHA256:   hash,
		})
	}

	// 3. Security (8 defects, 4 clean)
	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("security/secrets_%d.go", i)
		code := fmt.Sprintf(`package security

// Hardcoded credential defect
const AWSAccessKey%d = "AKIA1111222233334444"
`, i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-SEC-SECRET-%02d", i),
			Category:                           repo_intel.CategorySecurity,
			DefectDescription:                  "Hardcoded AWS access key in source code",
			TargetFile:                         rel,
			TargetSymbol:                       fmt.Sprintf("AWSAccessKey%d", i),
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 4, EndLine: 4},
			ExpectedDiagnosticClassification:   "SECURITY",
			ExpectedSeverity:                   repo_intel.SeverityCritical,
			ExpectedConfidenceRange:            [2]float64{0.90, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy:          "HUMAN_GUIDED",
			ExpectedApprovalRequirements:       true,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 3; i++ {
		rel := fmt.Sprintf("security/command_exec_%d.go", i)
		fnName := fmt.Sprintf("RunUserScript%d", i)
		code := fmt.Sprintf(`package security

import (
	"fmt"
	"os/exec"
)

func %s(arg string) error {
	cmd := exec.Command("sh", "-c", "echo " + arg)
	return cmd.Run()
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-SEC-CMD-%02d", i),
			Category:                           repo_intel.CategorySecurity,
			DefectDescription:                  "Command injection risk via unescaped string concatenation",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 9, EndLine: 11},
			ExpectedDiagnosticClassification:   "SECURITY",
			ExpectedSeverity:                   repo_intel.SeverityHigh,
			ExpectedConfidenceRange:            [2]float64{0.88, 1.0},
			ExpectedRemediationEligibility:     "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy:          "AST_TRANSFORMATION",
			ExpectedApprovalRequirements:       false,
			FileSHA256:                         hash,
		})
	}

	for i := 1; i <= 3; i++ {
		rel := fmt.Sprintf("security/sql_injection_%d.go", i)
		fnName := fmt.Sprintf("QueryUnsafeSQL%d", i)
		code := fmt.Sprintf(`package security

import "database/sql"

func %s(db *sql.DB, input string) (*sql.Rows, error) {
	q := "SELECT id, name FROM items WHERE name = '" + input + "'"
	return db.Query(q)
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID:                          fmt.Sprintf("FIX-SEC-SQL-%02d", i),
			Category:                           repo_intel.CategorySecurity,
			DefectDescription:                  "SQL injection risk via raw query string concatenation",
			TargetFile:                         rel,
			TargetSymbol:                       fnName,
			GroundTruthLocation:                GroundTruthLocation{File: rel, StartLine: 6, EndLine: 8},
			ExpectedDiagnosticClassification:   "SECURITY",
			ExpectedSeverity:                   repo_intel.SeverityHigh,
			ExpectedConfidenceRange:            [2]float64{0.85, 1.0},
			ExpectedRemediationEligibility:     "HUMAN_APPROVAL_REQUIRED",
			ExpectedExecutionStrategy:          "HUMAN_GUIDED",
			ExpectedApprovalRequirements:       true,
			FileSHA256:                         hash,
		})
	}

	// Clean security controls
	{
		rel := "security/clean_secret_1.go"
		code := `package security
import "os"
func GetSecretSafe() string { return os.Getenv("AWS_SECRET_KEY") }
`
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: "CLEAN-SEC-SECRET-01", Category: repo_intel.CategorySecurity,
			Description: "Clean control: Credential loaded from environment", TargetFile: rel, FileSHA256: hash,
		})
	}
	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("security/clean_exec_%d.go", i)
		code := fmt.Sprintf(`package security
import "os/exec"
func RunSafeCmd%d() error { return exec.Command("git", "status", "--porcelain").Run() }
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-SEC-EXEC-%02d", i), Category: repo_intel.CategorySecurity,
			Description: "Clean control: Static discrete arguments in exec.Command", TargetFile: rel, FileSHA256: hash,
		})
	}
	{
		rel := "security/clean_sql_1.go"
		code := `package security
import "database/sql"
func QuerySafeSQL(db *sql.DB, id int) *sql.Row {
	return db.QueryRow("SELECT id, name FROM items WHERE id = $1", id)
}
`
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: "CLEAN-SEC-SQL-01", Category: repo_intel.CategorySecurity,
			Description: "Clean control: Parameterized SQL query placeholder", TargetFile: rel, FileSHA256: hash,
		})
	}

	// 4. Dependency (8 defects, 4 clean)
	for i := 1; i <= 3; i++ {
		rel := fmt.Sprintf("dependency/Dockerfile_%d", i)
		code := fmt.Sprintf("FROM alpine:latest\nWORKDIR /app\nUSER appuser\nCMD [\"./run%d\"]\n", i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-DEP-UNPINNED-%02d", i), Category: repo_intel.CategoryDependency,
			DefectDescription: "Mutable base image tag (alpine:latest)", TargetFile: rel, TargetSymbol: "alpine:latest",
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 1, EndLine: 1},
			ExpectedDiagnosticClassification: "DEPENDENCY", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy: "RULE_BASED", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("dependency/Dockerfile_root_%d", i)
		code := fmt.Sprintf("FROM alpine:3.19.1\nWORKDIR /app\nCMD [\"./service%d\"]\n", i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-DEP-ROOT-%02d", i), Category: repo_intel.CategoryDependency,
			DefectDescription: "Missing container non-root USER directive", TargetFile: rel, TargetSymbol: "USER",
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 1, EndLine: 3},
			ExpectedDiagnosticClassification: "SECURITY", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy: "RULE_BASED", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 3; i++ {
		rel := fmt.Sprintf("dependency/nolock_%d/go.mod", i)
		code := fmt.Sprintf("module github.com/saksham/submodule%d\ngo 1.25.6\n", i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-DEP-NOLOCK-%02d", i), Category: repo_intel.CategoryDependency,
			DefectDescription: "Module lacks cryptographic lockfile go.sum", TargetFile: rel, TargetSymbol: "go.sum",
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 1, EndLine: 2},
			ExpectedDiagnosticClassification: "DEPENDENCY", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy: "RULE_BASED", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("dependency/clean_dockerfile_%d", i)
		code := fmt.Sprintf("FROM alpine:3.19.1\nUSER nonroot\nCMD [\"./app%d\"]\n", i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-DEP-DOCKER-%02d", i), Category: repo_intel.CategoryDependency,
			Description: "Clean control: Pinned tag and explicit USER directive", TargetFile: rel, FileSHA256: hash,
		})
	}

	for i := 1; i <= 2; i++ {
		modRel := fmt.Sprintf("dependency/clean_mod_%d/go.mod", i)
		sumRel := fmt.Sprintf("dependency/clean_mod_%d/go.sum", i)
		mHash, _ := writeFile(modRel, fmt.Sprintf("module cleanmod%d\ngo 1.25.6\n", i))
		_, _ = writeFile(sumRel, "github.com/stretchr/testify v1.8.4 h1:abc123456789=\n")
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-DEP-LOCK-%02d", i), Category: repo_intel.CategoryDependency,
			Description: "Clean control: Valid go.mod with cryptographic go.sum", TargetFile: modRel, FileSHA256: mHash,
		})
	}

	// 5. Architecture (8 defects, 4 clean)
	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("architecture/api_handler_%d.go", i)
		code := fmt.Sprintf(`package architecture

import "database/sql"

type APIHandler%d struct {
	dbConn *sql.DB
}

func (h *APIHandler%d) ServeHTTPDirect(q string) error {
	_, err := h.dbConn.Exec(q)
	return err
}
`, i, i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-ARC-LAYER-%02d", i), Category: repo_intel.CategoryArchitecture,
			DefectDescription: "Direct database import in API presentation layer", TargetFile: rel, TargetSymbol: "database/sql",
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 3, EndLine: 13},
			ExpectedDiagnosticClassification: "ARCHITECTURE", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "INVESTIGATION_REQUIRED",
			ExpectedExecutionStrategy: "INVESTIGATION_ONLY", ExpectedApprovalRequirements: true, FileSHA256: hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("architecture/cycle_%d.go", i)
		code := fmt.Sprintf(`package architecture

type ServiceAlpha%d struct {
	Peer *ServiceBeta%d
}

type ServiceBeta%d struct {
	Peer *ServiceAlpha%d
}
`, i, i, i, i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-ARC-CYCLE-%02d", i), Category: repo_intel.CategoryArchitecture,
			DefectDescription: "Mutual circular type dependency", TargetFile: rel, TargetSymbol: fmt.Sprintf("ServiceAlpha%d", i),
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 3, EndLine: 10},
			ExpectedDiagnosticClassification: "ARCHITECTURE", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "INVESTIGATION_REQUIRED",
			ExpectedExecutionStrategy: "INVESTIGATION_ONLY", ExpectedApprovalRequirements: true, FileSHA256: hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("architecture/clean_layer_%d.go", i)
		code := fmt.Sprintf(`package architecture

type UserRepository%d interface {
	FindUser(id string) (string, error)
}

type CleanHandler%d struct {
	repo UserRepository%d
}
`, i, i, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-ARC-LAYER-%02d", i), Category: repo_intel.CategoryArchitecture,
			Description: "Clean control: Decoupled service interface abstraction", TargetFile: rel, FileSHA256: hash,
		})
	}

	for i := 1; i <= 2; i++ {
		rel := fmt.Sprintf("architecture/clean_arch_%d.go", i)
		code := fmt.Sprintf(`package architecture

type NodeA%d struct { ID int }
type NodeB%d struct { Parent NodeA%d }
`, i, i, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-ARC-CYCLE-%02d", i), Category: repo_intel.CategoryArchitecture,
			Description: "Clean control: Acyclic unidirectional composition", TargetFile: rel, FileSHA256: hash,
		})
	}

	// 6. Testing (8 defects, 4 clean)
	for i := 1; i <= 8; i++ {
		rel := fmt.Sprintf("testing/untested_service_%d.go", i)
		fnName := fmt.Sprintf("CalculateFinancialTaxRate%d", i)
		code := fmt.Sprintf(`package testing

func %s(amount float64) float64 {
	if amount <= 0 { return 0 }
	return amount * 0.15
}
`, fnName)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-TST-UNTESTED-%02d", i), Category: repo_intel.CategoryTestQuality,
			DefectDescription: "Exported function lacks automated test suite", TargetFile: rel, TargetSymbol: fnName,
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 3, EndLine: 6},
			ExpectedDiagnosticClassification: "TEST_QUALITY", ExpectedSeverity: repo_intel.SeverityMedium,
			ExpectedConfidenceRange: [2]float64{0.85, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy: "LLM_ASSISTED", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("testing/tested_service_%d.go", i)
		testRel := fmt.Sprintf("testing/tested_service_%d_test.go", i)
		fnName := fmt.Sprintf("CoveredCalculateSum%d", i)
		code := fmt.Sprintf(`package testing
func %s(a, b int) int { return a + b }
`, fnName)
		testCode := fmt.Sprintf(`package testing
import "testing"
func Test%s(t *testing.T) {
	if %s(2, 3) != 5 { t.Fail() }
}
`, fnName, fnName)
		hash, _ := writeFile(rel, code)
		_, _ = writeFile(testRel, testCode)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-TST-COVERED-%02d", i), Category: repo_intel.CategoryTestQuality,
			Description: "Clean control: Exported function with verified unit test", TargetFile: rel, TargetSymbol: fnName, FileSHA256: hash,
		})
	}

	// 7. Maintainability (8 defects, 4 clean)
	for i := 1; i <= 8; i++ {
		rel := fmt.Sprintf("maintainability/duplicate_fn_%d.go", i)
		code := fmt.Sprintf(`package maintainability

import "strings"

func NormalizePathSegmentA%d(p string) string {
	t := strings.TrimSpace(p)
	l := strings.ToLower(t)
	parts := strings.Split(l, "/")
	if len(parts) == 0 { return "" }
	return parts[0]
}

func NormalizePathSegmentB%d(p string) string {
	t := strings.TrimSpace(p)
	l := strings.ToLower(t)
	parts := strings.Split(l, "/")
	if len(parts) == 0 { return "" }
	return parts[0]
}
`, i, i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-MNT-DUPLICATE-%02d", i), Category: repo_intel.CategoryMaintainability,
			DefectDescription: "Exact duplicate function AST block", TargetFile: rel, TargetSymbol: fmt.Sprintf("NormalizePathSegmentA%d", i),
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 5, EndLine: 19},
			ExpectedDiagnosticClassification: "MAINTAINABILITY", ExpectedSeverity: repo_intel.SeverityLow,
			ExpectedConfidenceRange: [2]float64{0.90, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_WITH_REVIEW",
			ExpectedExecutionStrategy: "AST_TRANSFORMATION", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("maintainability/clean_helper_%d.go", i)
		code := fmt.Sprintf(`package maintainability

import "strings"

func CleanSharedHelper%d(s string) string {
	return strings.TrimSpace(s)
}
`, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-MNT-SHARED-%02d", i), Category: repo_intel.CategoryMaintainability,
			Description: "Clean control: Non-duplicated modular helper", TargetFile: rel, FileSHA256: hash,
		})
	}

	// 8. Documentation (8 defects, 4 clean)
	for i := 1; i <= 8; i++ {
		rel := fmt.Sprintf("documentation/undocumented_api_%d.go", i)
		code := fmt.Sprintf(`package documentation

type PublicClient%d struct {
	Host string
}

func NewPublicClient%d(host string) *PublicClient%d {
	return &PublicClient%d{Host: host}
}
`, i, i, i, i)
		hash, _ := writeFile(rel, code)
		manifest.DefectFixtures = append(manifest.DefectFixtures, DefectFixture{
			FixtureID: fmt.Sprintf("FIX-DOC-UNDOC-%02d", i), Category: repo_intel.CategoryDocumentation,
			DefectDescription: "Exported public API symbol lacks godoc documentation", TargetFile: rel, TargetSymbol: fmt.Sprintf("PublicClient%d", i),
			GroundTruthLocation: GroundTruthLocation{File: rel, StartLine: 3, EndLine: 9},
			ExpectedDiagnosticClassification: "DOCUMENTATION", ExpectedSeverity: repo_intel.SeverityLow,
			ExpectedConfidenceRange: [2]float64{0.88, 1.0}, ExpectedRemediationEligibility: "SUPPORTED_AUTOMATION",
			ExpectedExecutionStrategy: "RULE_BASED", ExpectedApprovalRequirements: false, FileSHA256: hash,
		})
	}

	for i := 1; i <= 4; i++ {
		rel := fmt.Sprintf("documentation/documented_api_%d.go", i)
		code := fmt.Sprintf(`package documentation

// DocumentedClient%d provides documented access to remote service endpoints.
type DocumentedClient%d struct {
	Endpoint string
}

// Connect%d establishes a verified TLS connection to the remote endpoint.
func (c *DocumentedClient%d) Connect%d() bool {
	return true
}
`, i, i, i, i, i)
		hash, _ := writeFile(rel, code)
		manifest.CleanFixtures = append(manifest.CleanFixtures, CleanFixture{
			FixtureID: fmt.Sprintf("CLEAN-DOC-DOC-%02d", i), Category: repo_intel.CategoryDocumentation,
			Description: "Clean control: Exported types and methods with comprehensive godoc", TargetFile: rel, FileSHA256: hash,
		})
	}

	manifest.TotalFixtures = len(manifest.DefectFixtures) + len(manifest.CleanFixtures)
	manifest.DefectFixturesCount = len(manifest.DefectFixtures)
	manifest.CleanFixturesCount = len(manifest.CleanFixtures)

	// Serialize manifest
	if manifestPath != "" {
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
			return nil, err
		}
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(manifestPath, data, 0644); err != nil {
			return nil, err
		}
	}

	return manifest, nil
}

