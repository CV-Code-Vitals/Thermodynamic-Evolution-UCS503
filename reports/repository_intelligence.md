# Repository Intelligence Diagnostic Report

**Version:** `1.0.0`  
**Generated At:** 2026-11-03T08:42:19Z  
**Repository Root:** `E:\SWE`  
**Git Branch:** `main` (Commit: `a95d242018273d8efc28c04395223b10c05635b1`, Clean: `true`)  
**Total Files Discovered:** 238  
**Total Duration:** 2903 ms  

## Executive Summary

The autonomous Repository Intelligence Agent completed whole-repository discovery and multi-dimensional analysis. A total of **433 findings** were identified across 300 files.

| Severity | Count |
|---|---|
| Critical | 0 |
| High | 73 |
| Medium | 189 |
| Low | 171 |
| Informational | 0 |

### Findings by Category

| Category | Count |
|---|---|
| RELIABILITY | 64 |
| SECURITY | 8 |
| STRUCTURAL | 348 |
| DEPENDENCY | 1 |
| TEST_QUALITY | 12 |

### Detected Languages & Code Footprint

| Language | Files | Code Lines | Total Size | Parsed / Errors |
|---|---|---|---|---|
| Dockerfile | 3 | 92 | 0.0 KB | 3 / 0 |
| Go | 78 | 24051 | 930.7 KB | 78 / 0 |
| JSON | 13 | 12423 | 1384.0 KB | 50 / 0 |
| JavaScript | 38 | 0 | 1186.3 KB | 0 / 0 |
| JavaScript / JSX | 7 | 6431 | 55.2 KB | 41 / 1 |
| Makefile | 1 | 2 | 0.0 KB | 1 / 0 |
| Markdown | 31 | 0 | 756.2 KB | 0 / 0 |
| Python | 2 | 269 | 4.5 KB | 43 / 0 |
| Rust | 4 | 1946 | 85.8 KB | 4 / 1 |
| SQL | 1 | 13 | 0.3 KB | 1 / 0 |
| Shell | 5 | 94 | 5.0 KB | 5 / 0 |
| TOML | 3 | 81 | 4.1 KB | 3 / 0 |
| YAML | 13 | 2176 | 137.5 KB | 7 / 0 |

### Dependency Manifests & Ecosystems

| Manifest File | Ecosystem | Dependencies | Lockfile Present |
|---|---|---|---|
| `admin-portal/package-lock.json` | package-lock.json | 0 | No |
| `admin-portal/package.json` | package.json | 10 | Yes |
| `analysis-authorizer/go.mod` | go.mod | 14 | Yes |
| `analysis-authorizer/go.sum` | go.sum | 0 | No |
| `api/Dockerfile` | Dockerfile | 2 | No |
| `api/go.mod` | go.mod | 31 | Yes |
| `api/go.sum` | go.sum | 0 | No |
| `docker/sandbox-runner/Dockerfile` | Dockerfile | 1 | No |
| `docker-compose.analysis.yml` | docker-compose.analysis.yml | 0 | No |
| `docker-compose.ec2.yml` | docker-compose.ec2.yml | 0 | No |
| `docker-compose.sandbox.yml` | docker-compose.sandbox.yml | 0 | No |
| `docker-compose.verify.yml` | docker-compose.verify.yml | 0 | No |
| `docker-compose.yml` | docker-compose.yml | 0 | No |
| `go-backend/Dockerfile` | Dockerfile | 3 | No |
| `go-backend/go.mod` | go.mod | 54 | Yes |
| `go-backend/go.sum` | go.sum | 0 | No |
| `infra/analysis/stream-proxy/package-lock.json` | package-lock.json | 0 | No |
| `infra/analysis/stream-proxy/package.json` | package.json | 2 | Yes |
| `infra/analysis/template.yaml` | template.yaml | 0 | No |
| `package-lock.json` | package-lock.json | 0 | No |
| `parser/Cargo.lock` | Cargo.lock | 0 | No |
| `parser/Cargo.toml` | Cargo.toml | 5 | Yes |
| `samples/benchmark/go.mod` | go.mod | 0 | No |
| `source-stager/go.mod` | go.mod | 19 | Yes |
| `source-stager/go.sum` | go.sum | 0 | No |
| `template.yaml` | template.yaml | 0 | No |
| `thermodynamic-ast-engine/Cargo.lock` | Cargo.lock | 0 | No |
| `thermodynamic-ast-engine/Cargo.toml` | Cargo.toml | 21 | Yes |

### Architecture & Dependency Graph

- **Total Nodes (Files):** 519
- **Total Directed Edges (Dependencies):** 1102
- **Discovered Service Boundaries:** 12

## Detailed Findings

### 1. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-graph_visualizer_graph_mjs-L92`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in graph-visualizer/graph.mjs at line 92 appears to concatenate variables directly.

**Evidence Locations:**

- `graph-visualizer/graph.mjs` (Lines 92-92) — *if (node.type === 'folder') { ctx.fillText('Select to open folder', node.x + 13, node.y + 57); }*

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 2. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-reports_repository_intelligence_json-L9725`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in reports/repository_intelligence.json at line 9725 appears to concatenate variables directly.

**Evidence Locations:**

- `reports/repository_intelligence.json` (Lines 9725-9725) — *"description": "if (node.type === 'folder') { ctx.fillText('Select to open folder', node.x + 13, node.y + 57); }"*

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 3. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-reports_repository_intelligence_json-L9753`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in reports/repository_intelligence.json at line 9753 appears to concatenate variables directly.

**Evidence Locations:**

- `reports/repository_intelligence.json` (Lines 9753-9753) — *"description": "assert.match(deploymentReadme, /Delete the analysis stack\\s+first|stack must already be deleted/i);"*

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 4. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-reports_repository_intelligence_md-L102`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in reports/repository_intelligence.md at line 102 appears to concatenate variables directly.

**Evidence Locations:**

- `reports/repository_intelligence.md` (Lines 102-102) — *- `graph-visualizer/graph.mjs` (Lines 92-92) — *if (node.type === 'folder') { ctx.fillText('Select to open folder', node.x + 13, node.y + 57); }**

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 5. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-reports_repository_intelligence_md-L122`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in reports/repository_intelligence.md at line 122 appears to concatenate variables directly.

**Evidence Locations:**

- `reports/repository_intelligence.md` (Lines 122-122) — *- `tests/analysis-workflow.test.mjs` (Lines 577-577) — *assert.match(deploymentReadme, /Delete the analysis stack\s+first|stack must already be deleted/i);**

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 6. [HIGH] Potential SQL injection via query string concatenation (`RI-SEC-SQL-CONCAT-tests_analysis_workflow_test_mjs-L577`)

- **Category:** `SECURITY`
- **Verification Status:** `HEURISTIC` (Confidence: `0.80`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.

SQL query string in tests/analysis-workflow.test.mjs at line 577 appears to concatenate variables directly.

**Evidence Locations:**

- `tests/analysis-workflow.test.mjs` (Lines 577-577) — *assert.match(deploymentReadme, /Delete the analysis stack\s+first|stack must already be deleted/i);*

> [!TIP]
> **Recommended Action:** Use parameterized queries with bind variables ($1, ?, :name).
>
> **Expected Benefit:** Prevents SQL injection attacks entirely.

---

### 7. [HIGH] Potential command injection risk in exec.Command (`RI-SEC-CMD-INJECT-go_backend_agent_repo_intel_repo_intel_test_go-L66`)

- **Category:** `SECURITY`
- **Verification Status:** `PROBABLE` (Confidence: `0.88`)
- **Detection Method:** `STATIC_TAINT_HEURISTIC`
- **Root Cause:** String concatenation into process execution arguments.

Command execution in go-backend/agent/repo_intel/repo_intel_test.go at line 66 dynamically concatenates arguments or command string.

**Evidence Locations:**

- `go-backend/agent/repo_intel/repo_intel_test.go` (Lines 66-66) — *cmd := exec.Command("sh", fmt.Sprintf("-c %s", input))*

> [!TIP]
> **Recommended Action:** Pass arguments as separate string elements rather than formatting shell strings.
>
> **Expected Benefit:** Eliminates arbitrary shell argument injection vulnerabilities.

---

### 8. [HIGH] Excessive cyclomatic complexity (M=103) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-build-L260`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "build" has cyclomatic complexity of 103, requiring at least 103 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 260-260):
```
pub fn build(files: &[FileReport], root: &Path) -> Vec<Relationship> {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 9. [HIGH] Excessive cyclomatic complexity (M=20) (`RI-STR-COMPLEX-E:_SWE_api_archive_go-isPowerPointZip-L465`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isPowerPointZip" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 465-465):
```
func isPowerPointZip(content []byte) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 10. [HIGH] Excessive cyclomatic complexity (M=20) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-ProposeTargeted-L317`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ProposeTargeted" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 317-317):
```
func (a *AdaptiveMutator) ProposeTargeted(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 11. [HIGH] Excessive cyclomatic complexity (M=20) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-Reconcile-L307`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "Reconcile" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 307-307):
```
func (l *SandboxLifecycle) Reconcile(ctx context.Context, now time.Time) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 12. [HIGH] Excessive cyclomatic complexity (M=20) (`RI-STR-COMPLEX-E:_SWE_source_stager_main_go-validateArchive-L382`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateArchive" has cyclomatic complexity of 20, requiring at least 20 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 382-382):
```
func (s *RepositoryStager) validateArchive(archive []byte) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 13. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_components_AdminPortal_jsx-<anonymous@699>-L699`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@699>" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/AdminPortal.jsx` (Lines 699-699):
```
(d, index) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 14. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-ServeHTTP-L96`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ServeHTTP" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 96-96):
```
func (a *githubAccountService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 15. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_api_main_go-serveDownload-L337`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "serveDownload" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/main.go` (Lines 337-337):
```
func serveDownload(w http.ResponseWriter, r *http.Request) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 16. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_energy_go-EvaluateEnergy-L155`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "EvaluateEnergy" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy.go` (Lines 155-155):
```
func (r *RustEngineEvaluator) EvaluateEnergy(ctx context.Context, sourceContent []byte, relFilePath string, funcName str
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 17. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-hoistLoopAllocations-L843`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "hoistLoopAllocations" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 843-843):
```
func hoistLoopAllocations(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 18. [HIGH] Excessive cyclomatic complexity (M=21) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-parseArgs-L14`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "parseArgs" has cyclomatic complexity of 21, requiring at least 21 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 14-14):
```
function parseArgs(args) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 19. [HIGH] Excessive cyclomatic complexity (M=22) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_orchestrator_go-Run-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "Run" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/orchestrator.go` (Lines 42-42):
```
func (o *Orchestrator) Run(ctx context.Context) (*IntelligenceReport, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 20. [HIGH] Excessive cyclomatic complexity (M=22) (`RI-STR-COMPLEX-E:_SWE_go_backend_evolve_go-streamEvolutionEvents-L206`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "streamEvolutionEvents" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/evolve.go` (Lines 206-206):
```
func streamEvolutionEvents(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 21. [HIGH] Excessive cyclomatic complexity (M=22) (`RI-STR-COMPLEX-E:_SWE_go_backend_workspace_go-readWorkspaceTarget-L145`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "readWorkspaceTarget" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/workspace.go` (Lines 145-145):
```
func readWorkspaceTarget(root, relativePath, functionName string, startByte int) (*evolutionTarget, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 22. [HIGH] Excessive cyclomatic complexity (M=22) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-assertOwnedResource-L337`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "assertOwnedResource" has cyclomatic complexity of 22, requiring at least 22 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 337-337):
```
function assertOwnedResource(resource, prefix, region, stackId, stackName) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 23. [HIGH] Excessive cyclomatic complexity (M=23) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_archive_api_js-isValidUpload-L49`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isValidUpload" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/archive-api.js` (Lines 49-49):
```
(upload) =>
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 24. [HIGH] Excessive cyclomatic complexity (M=23) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-simplifyBooleanReturns-L511`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "simplifyBooleanReturns" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 511-511):
```
func simplifyBooleanReturns(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 25. [HIGH] Excessive cyclomatic complexity (M=23) (`RI-STR-COMPLEX-E:_SWE_go_backend_snapshot_portal_test_go-TestSnapshotPortalScanEvolutionAndPatch-L19`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestSnapshotPortalScanEvolutionAndPatch" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_portal_test.go` (Lines 19-19):
```
func TestSnapshotPortalScanEvolutionAndPatch(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 26. [HIGH] Excessive cyclomatic complexity (M=23) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_OrbitControls_js-handleKeyDown-L546`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "handleKeyDown" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 546-546):
```
function handleKeyDown( event ) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 27. [HIGH] Excessive cyclomatic complexity (M=23) (`RI-STR-COMPLEX-E:_SWE_source_stager_main_go-isPublicAddress-L320`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isPublicAddress" has cyclomatic complexity of 23, requiring at least 23 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 320-320):
```
func isPublicAddress(ip net.IP) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 28. [HIGH] Excessive cyclomatic complexity (M=24) (`RI-STR-COMPLEX-E:_SWE_api_lambda_test_go-TestAPIGatewayV2HandlerMapsRequestAndResponse-L241`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestAPIGatewayV2HandlerMapsRequestAndResponse" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/lambda_test.go` (Lines 241-241):
```
func TestAPIGatewayV2HandlerMapsRequestAndResponse(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 29. [HIGH] Excessive cyclomatic complexity (M=24) (`RI-STR-COMPLEX-E:_SWE_api_local_store_test_go-TestLocalArchiveUploadFinalizeListDownloadDelete-L18`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestLocalArchiveUploadFinalizeListDownloadDelete" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/local_store_test.go` (Lines 18-18):
```
func TestLocalArchiveUploadFinalizeListDownloadDelete(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 30. [HIGH] Excessive cyclomatic complexity (M=24) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_workspace_go-ExtractWorkspaceArchive-L154`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ExtractWorkspaceArchive" has cyclomatic complexity of 24, requiring at least 24 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 154-154):
```
func ExtractWorkspaceArchive(archiveBytes []byte, destination string, limits WorkspaceArchiveLimits) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 31. [HIGH] Excessive cyclomatic complexity (M=26) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-accountServiceFromEnvironment-L36`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "accountServiceFromEnvironment" has cyclomatic complexity of 26, requiring at least 26 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 36-36):
```
func accountServiceFromEnvironment(store accountStore, secret string) (*githubAccountService, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 32. [HIGH] Excessive cyclomatic complexity (M=27) (`RI-STR-COMPLEX-E:_SWE_api_main_go-validateCORSAllowedOrigins-L145`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateCORSAllowedOrigins" has cyclomatic complexity of 27, requiring at least 27 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/main.go` (Lines 145-145):
```
func validateCORSAllowedOrigins(value string) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 33. [HIGH] Excessive cyclomatic complexity (M=27) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_OrbitControls_js-onMouseDown-L810`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "onMouseDown" has cyclomatic complexity of 27, requiring at least 27 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 810-810):
```
function onMouseDown( event ) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 34. [HIGH] Excessive cyclomatic complexity (M=28) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-parseSavedRepository-L366`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "parseSavedRepository" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 366-366):
```
func parseSavedRepository(raw string) (SavedRepository, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 35. [HIGH] Excessive cyclomatic complexity (M=28) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_workspace_go-CreateWorkspaceArchive-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "CreateWorkspaceArchive" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 41-41):
```
func CreateWorkspaceArchive(repoPath string, maxArchiveBytes int64) ([]byte, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 36. [HIGH] Excessive cyclomatic complexity (M=28) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_OrbitControls_js-update-L166`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "update" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 166-166):
```
function update() {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 37. [HIGH] Excessive cyclomatic complexity (M=28) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-detect_language-L657`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "detect_language" has cyclomatic complexity of 28, requiring at least 28 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 657-657):
```
pub fn detect_language(path: &Path) -> Option<Language> {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 38. [HIGH] Excessive cyclomatic complexity (M=29) (`RI-STR-COMPLEX-E:_SWE_api_router_go-registerArchiveServiceRoutes-L80`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "registerArchiveServiceRoutes" has cyclomatic complexity of 29, requiring at least 29 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/router.go` (Lines 80-80):
```
func registerArchiveServiceRoutes(mux *http.ServeMux, deps Dependencies) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 39. [HIGH] Excessive cyclomatic complexity (M=29) (`RI-STR-COMPLEX-E:_SWE_go_backend_scanner_go-scanRepoHandler-L25`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "scanRepoHandler" has cyclomatic complexity of 29, requiring at least 29 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/scanner.go` (Lines 25-25):
```
func scanRepoHandler(enginePath string, engineTimeout time.Duration, workspaces *workspaceStore) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 40. [HIGH] Excessive cyclomatic complexity (M=30) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-scan_file-L700`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "scan_file" has cyclomatic complexity of 30, requiring at least 30 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 700-700):
```
pub fn scan_file(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 41. [HIGH] Excessive cyclomatic complexity (M=31) (`RI-STR-COMPLEX-E:_SWE_admin_portal_vite_config_js-<anonymous@36>-L36`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@36>" has cyclomatic complexity of 31, requiring at least 31 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/vite.config.js` (Lines 36-36):
```
({ mode }) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 42. [HIGH] Excessive cyclomatic complexity (M=31) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-inventory-L380`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "inventory" has cyclomatic complexity of 31, requiring at least 31 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 380-380):
```
function inventory(stackName, prefix, region, expectedStackId) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 43. [HIGH] Excessive cyclomatic complexity (M=32) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-extractCommonSubexpressions-L705`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "extractCommonSubexpressions" has cyclomatic complexity of 32, requiring at least 32 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 705-705):
```
func extractCommonSubexpressions(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 44. [HIGH] Excessive cyclomatic complexity (M=33) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_graph_mjs-details-L131`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "details" has cyclomatic complexity of 33, requiring at least 33 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 131-131):
```
function details(node) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 45. [HIGH] Excessive cyclomatic complexity (M=33) (`RI-STR-COMPLEX-E:_SWE_source_stager_main_go-validGitHubSegment-L149`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validGitHubSegment" has cyclomatic complexity of 33, requiring at least 33 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 149-149):
```
func validGitHubSegment(value string) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 46. [HIGH] Excessive cyclomatic complexity (M=34) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_workspace_go-shouldCopyProjectFile-L330`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "shouldCopyProjectFile" has cyclomatic complexity of 34, requiring at least 34 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 330-330):
```
func shouldCopyProjectFile(info os.FileInfo) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 47. [HIGH] Excessive cyclomatic complexity (M=37) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-normalizeResource-L157`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "normalizeResource" has cyclomatic complexity of 37, requiring at least 37 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 157-157):
```
function normalizeResource(resource) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 48. [HIGH] Excessive cyclomatic complexity (M=37) (`RI-STR-COMPLEX-E:_SWE_source_stager_main_go-ValidateGitHubRepositoryURL-L126`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ValidateGitHubRepositoryURL" has cyclomatic complexity of 37, requiring at least 37 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 126-126):
```
func ValidateGitHubRepositoryURL(rawURL string) (string, string, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 49. [HIGH] Excessive cyclomatic complexity (M=38) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_evaluator_go-EvaluateCandidate-L209`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "EvaluateCandidate" has cyclomatic complexity of 38, requiring at least 38 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_evaluator.go` (Lines 209-209):
```
func (r *RemoteBrokerEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error)
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 50. [HIGH] Excessive cyclomatic complexity (M=40) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-accountRequest-L294`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "accountRequest" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 294-294):
```
func (a *githubAccountService) accountRequest(w http.ResponseWriter, r *http.Request) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 51. [HIGH] Excessive cyclomatic complexity (M=40) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_quality_go-AnalyzeQuality-L21`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "AnalyzeQuality" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/quality.go` (Lines 21-21):
```
func (qa *QualityAnalyzer) AnalyzeQuality(files []DiscoveredFile, ciWorkflows []string) ([]Finding, string) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 52. [HIGH] Excessive cyclomatic complexity (M=40) (`RI-STR-COMPLEX-E:_SWE_go_backend_broker_workspace_go-newSandboxBrokerHandler-L276`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "newSandboxBrokerHandler" has cyclomatic complexity of 40, requiring at least 40 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/broker_workspace.go` (Lines 276-276):
```
func newSandboxBrokerHandler(token string, evaluator agent.CandidateEvaluator, workspaces *brokerWorkspaceManager, limit
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 53. [HIGH] Excessive cyclomatic complexity (M=41) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-checkBufferEscapes-L910`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "checkBufferEscapes" has cyclomatic complexity of 41, requiring at least 41 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 910-910):
```
func checkBufferEscapes(forStmt *ast.ForStmt, allocStmt ast.Stmt, bufName string) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 54. [HIGH] Excessive cyclomatic complexity (M=42) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-<anonymous@219>-L219`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@219>" has cyclomatic complexity of 42, requiring at least 42 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 219-219):
```
async (event, rawResponseStream) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 55. [HIGH] Excessive cyclomatic complexity (M=43) (`RI-STR-COMPLEX-E:_SWE_api_archive_go-isPowerPointCompoundFile-L528`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isPowerPointCompoundFile" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 528-528):
```
func isPowerPointCompoundFile(content []byte) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 56. [HIGH] Excessive cyclomatic complexity (M=43) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_discovery_go-DiscoverRepository-L169`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "DiscoverRepository" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/discovery.go` (Lines 169-169):
```
func DiscoverRepository(ctx context.Context, repoRoot string, opts DiscoveryOptions) (*DiscoveryResult, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 57. [HIGH] Excessive cyclomatic complexity (M=43) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-is_allocation-L434`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "is_allocation" has cyclomatic complexity of 43, requiring at least 43 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 434-434):
```
fn is_allocation(text: &str, lang: Language) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 58. [HIGH] Excessive cyclomatic complexity (M=45) (`RI-STR-COMPLEX-E:_SWE_api_analysis_proxy_go-newAnalysisProxy-L15`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "newAnalysisProxy" has cyclomatic complexity of 45, requiring at least 45 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy.go` (Lines 15-15):
```
func newAnalysisProxy(address, token string) (http.Handler, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 59. [HIGH] Excessive cyclomatic complexity (M=45) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-validateRepositoryURL-L54`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateRepositoryURL" has cyclomatic complexity of 45, requiring at least 45 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 54-54):
```
function validateRepositoryURL(value) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 60. [HIGH] Excessive cyclomatic complexity (M=46) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_languages_rs-name-L43`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "name" has cyclomatic complexity of 46, requiring at least 46 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/languages.rs` (Lines 43-43):
```
pub fn name(self, path: &Path) -> String {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 61. [HIGH] Excessive cyclomatic complexity (M=49) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-visit-L490`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "visit" has cyclomatic complexity of 49, requiring at least 49 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 490-490):
```
fn visit(&mut self, node: tree_sitter::Node) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 62. [HIGH] Excessive cyclomatic complexity (M=51) (`RI-STR-COMPLEX-E:_SWE_go_backend_evolve_go-evolveHandler-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "evolveHandler" has cyclomatic complexity of 51, requiring at least 51 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/evolve.go` (Lines 42-42):
```
func evolveHandler(workspaces *workspaceStore, snapshots ...SnapshotWorkspaceProvider) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 63. [HIGH] Excessive cyclomatic complexity (M=52) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_evaluator_go-EvaluateCandidate-L188`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "EvaluateCandidate" has cyclomatic complexity of 52, requiring at least 52 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/evaluator.go` (Lines 188-188):
```
func (d *DockerEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 64. [HIGH] Excessive cyclomatic complexity (M=53) (`RI-STR-COMPLEX-E:_SWE_api_analysis_proxy_go-registerAnalysisRoutes-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "registerAnalysisRoutes" has cyclomatic complexity of 53, requiring at least 53 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy.go` (Lines 42-42):
```
func registerAnalysisRoutes(mux *http.ServeMux, deps Dependencies) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 65. [HIGH] Excessive cyclomatic complexity (M=53) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_model_mjs-buildGraphModel-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "buildGraphModel" has cyclomatic complexity of 53, requiring at least 53 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/model.mjs` (Lines 23-23):
```
function buildGraphModel(files, { query = '', language = '', focusPath = '', showFunctions = false, sort = 'path', relat
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 66. [HIGH] Excessive cyclomatic complexity (M=60) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_graph_mjs-draw-L44`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "draw" has cyclomatic complexity of 60, requiring at least 60 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 44-44):
```
function draw() {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 67. [HIGH] Excessive cyclomatic complexity (M=64) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-is_concurrency_primitive-L341`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "is_concurrency_primitive" has cyclomatic complexity of 64, requiring at least 64 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 341-341):
```
fn is_concurrency_primitive(kind: &str, text: &str, lang: Language) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 68. [HIGH] Excessive cyclomatic complexity (M=69) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_annealer_go-Evolve-L99`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "Evolve" has cyclomatic complexity of 69, requiring at least 69 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/annealer.go` (Lines 99-99):
```
func (a *Annealer) Evolve(ctx context.Context, repoPath, filePath, funcName string, startByte, endByte int, baseEnergy f
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 69. [HIGH] Excessive cyclomatic complexity (M=73) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-walk-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "walk" has cyclomatic complexity of 73, requiring at least 73 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 41-41):
```
fn walk(node: Node, source: &[u8], lang: Language, found: &mut BTreeSet<(String, String)>) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 70. [HIGH] Excessive cyclomatic complexity (M=80) (`RI-STR-COMPLEX-E:_SWE_tests_analysis_workflow_test_mjs-isValidBucket-L217`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isValidBucket" has cyclomatic complexity of 80, requiring at least 80 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/tests/analysis-workflow.test.mjs` (Lines 217-217):
```
(value) =>
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 71. [HIGH] Excessive cyclomatic complexity (M=88) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_components_AdminPortal_jsx-AdminPortal-L30`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "AdminPortal" has cyclomatic complexity of 88, requiring at least 88 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/AdminPortal.jsx` (Lines 30-30):
```
() => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 72. [HIGH] Excessive cyclomatic complexity (M=89) (`RI-STR-COMPLEX-E:_SWE_api_router_go-newRouter-L224`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "newRouter" has cyclomatic complexity of 89, requiring at least 89 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/router.go` (Lines 224-224):
```
func newRouter(deps Dependencies) http.Handler {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 73. [HIGH] Excessive cyclomatic complexity (M=97) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-is_blocking_io-L400`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "is_blocking_io" has cyclomatic complexity of 97, requiring at least 97 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 400-400):
```
fn is_blocking_io(text: &str, lang: Language) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 74. [MEDIUM] Missing go.sum lockfile (`RI-DEP-GOSUM-samples_benchmark`)

- **Category:** `DEPENDENCY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `MANIFEST_ANALYSIS`
- **Root Cause:** go.sum file not committed or generated

Go module at samples/benchmark/go.mod lacks a go.sum lockfile, preventing cryptographic build verification.

**Evidence Locations:**

- `samples/benchmark/go.mod` — *go.mod without adjacent go.sum*

> [!TIP]
> **Recommended Action:** Run 'go mod download' or 'go mod tidy' and commit go.sum.
>
> **Expected Benefit:** Ensures reproducible and tamper-proof Go dependency builds.

---

### 75. [MEDIUM] AST parse error encountered (`RI-STR-PARSE-E:_SWE_admin_portal_src_components_AdminPortal_jsx`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `1.00`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Syntax error or unsupported grammar construct.

Source file E:/SWE/admin-portal/src/components/AdminPortal.jsx contains syntax or grammar parse errors preventing complete AST analysis.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/AdminPortal.jsx` — *Tree-sitter failed to parse file completely*

> [!TIP]
> **Recommended Action:** Inspect file syntax and ensure compatibility with standard language grammars.
>
> **Expected Benefit:** Enables full structural AST analysis and prevents compiler failures.

---

### 76. [MEDIUM] AST parse error encountered (`RI-STR-PARSE-E:_SWE_thermodynamic_ast_engine_src_relationships_rs`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `1.00`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Syntax error or unsupported grammar construct.

Source file E:/SWE/thermodynamic-ast-engine/src/relationships.rs contains syntax or grammar parse errors preventing complete AST analysis.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` — *Tree-sitter failed to parse file completely*

> [!TIP]
> **Recommended Action:** Inspect file syntax and ensure compatibility with standard language grammars.
>
> **Expected Benefit:** Enables full structural AST analysis and prevents compiler failures.

---

### 77. [MEDIUM] Excessive control flow nesting depth (10) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-hoistLoopAllocations-L843`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "hoistLoopAllocations" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 10 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 843-843):
```
func hoistLoopAllocations(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 78. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_admin_portal_src_evolution_api_js-processLine-L20`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "processLine" in E:/SWE/admin-portal/src/evolution-api.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-api.js` (Lines 20-20):
```
(line) => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 79. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_admin_portal_src_evolution_api_js-readEvolutionStream-L5`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "readEvolutionStream" in E:/SWE/admin-portal/src/evolution-api.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-api.js` (Lines 5-5):
```
async (response, onEvent) => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 80. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_admin_portal_src_evolution_measurements_js-evolutionMeasurements-L1`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "evolutionMeasurements" in E:/SWE/admin-portal/src/evolution-measurements.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-measurements.js` (Lines 1-1):
```
function evolutionMeasurements(initialEnergy, telemetry, result) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 81. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_admin_portal_tests_graph_model_test_js-assertClearConnections-L6`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "assertClearConnections" in E:/SWE/admin-portal/tests/graph-model.test.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/admin-portal/tests/graph-model.test.js` (Lines 6-6):
```
function assertClearConnections({nodes, links}) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 82. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_admin_portal_vite_config_js-<anonymous@36>-L36`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@36>" in E:/SWE/admin-portal/vite.config.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/admin-portal/vite.config.js` (Lines 36-36):
```
({ mode }) => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 83. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_analysis_authorizer_main_test_go-TestSSMSecretStore_RequiresDecryptedSecureString-L227`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestSSMSecretStore_RequiresDecryptedSecureString" in E:/SWE/analysis-authorizer/main_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/main_test.go` (Lines 227-227):
```
func TestSSMSecretStore_RequiresDecryptedSecureString(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 84. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_analysis_authorizer_rest_authorizer_test_go-TestAuthorizerHandler_DeniesInvalidBearer-L57`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestAuthorizerHandler_DeniesInvalidBearer" in E:/SWE/analysis-authorizer/rest_authorizer_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/rest_authorizer_test.go` (Lines 57-57):
```
func TestAuthorizerHandler_DeniesInvalidBearer(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 85. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_accounts_test_go-exerciseAccountStore-L305`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "exerciseAccountStore" in E:/SWE/api/accounts_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 305-305):
```
func exerciseAccountStore(t *testing.T, store accountStore) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 86. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_lambda_test_go-TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey-L339`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey" in E:/SWE/api/lambda_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/lambda_test.go` (Lines 339-339):
```
func TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 87. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_local_store_go-PromoteUpload-L189`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "PromoteUpload" in E:/SWE/api/local_store.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 189-189):
```
func (s *LocalObjectStore) PromoteUpload(_ context.Context, pendingKey, archiveKey string) error {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 88. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_main_test_go-TestLegacyUploadRejectsInvalidFilesBeforeStorage-L443`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestLegacyUploadRejectsInvalidFilesBeforeStorage" in E:/SWE/api/main_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/main_test.go` (Lines 443-443):
```
func TestLegacyUploadRejectsInvalidFilesBeforeStorage(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 89. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_refresh_test_go-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials" in E:/SWE/api/refresh_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/refresh_test.go` (Lines 82-82):
```
func TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 90. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_api_router_go-newRouter-L224`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "newRouter" in E:/SWE/api/router.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/router.go` (Lines 224-224):
```
func newRouter(deps Dependencies) http.Handler {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 91. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_energy_test_go-TestEnergyCache_ConcurrentAccess-L671`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestEnergyCache_ConcurrentAccess" in E:/SWE/go-backend/agent/energy_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy_test.go` (Lines 671-671):
```
func TestEnergyCache_ConcurrentAccess(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 92. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_discovery_go-DiscoverRepository-L169`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "DiscoverRepository" in E:/SWE/go-backend/agent/repo_intel/discovery.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/discovery.go` (Lines 169-169):
```
func DiscoverRepository(ctx context.Context, repoRoot string, opts DiscoveryOptions) (*DiscoveryResult, error) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 93. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_manifests_go-parseGoMod-L154`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "parseGoMod" in E:/SWE/go-backend/agent/repo_intel/manifests.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 154-154):
```
func parseGoMod(fullPath, relPath string) ([]ManifestDependency, []Finding) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 94. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_reliability_security_go-AnalyzeReliabilityAndSecurity-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "AnalyzeReliabilityAndSecurity" in E:/SWE/go-backend/agent/repo_intel/reliability_security.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/reliability_security.go` (Lines 42-42):
```
func (rsa *ReliabilitySecurityAnalyzer) AnalyzeReliabilityAndSecurity(files []DiscoveredFile) []Finding {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 95. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_repo_intel_test_go-TestReliabilitySecurityAnalyzer_RedactionAndDetection-L184`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestReliabilitySecurityAnalyzer_RedactionAndDetection" in E:/SWE/go-backend/agent/repo_intel/repo_intel_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/repo_intel_test.go` (Lines 184-184):
```
func TestReliabilitySecurityAnalyzer_RedactionAndDetection(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 96. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_structural_go-convertReportToFindings-L67`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "convertReportToFindings" in E:/SWE/go-backend/agent/repo_intel/structural.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/structural.go` (Lines 67-67):
```
func (sa *StructuralAnalyzer) convertReportToFindings(report *agent.ThermodynamicReport) []Finding {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 97. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-ProposeTargeted-L317`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "ProposeTargeted" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 317-317):
```
func (a *AdaptiveMutator) ProposeTargeted(ctx context.Context, req MutationRequest) (*MutationProposal, error) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 98. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-checkUsesLoopVar-L1017`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "checkUsesLoopVar" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 1017-1017):
```
func checkUsesLoopVar(forStmt *ast.ForStmt, sizeExpr ast.Expr) bool {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 99. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-extractFunctionBody-L1068`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "extractFunctionBody" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 1068-1068):
```
func extractFunctionBody(fset *token.FileSet, file *ast.File) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 100. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_test_go-TestStrategy_DifferentialEquivalence-L610`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "TestStrategy_DifferentialEquivalence" in E:/SWE/go-backend/agent/strategy_test.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy_test.go` (Lines 610-610):
```
func TestStrategy_DifferentialEquivalence(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 101. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_broker_workspace_go-newSandboxBrokerHandler-L276`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "newSandboxBrokerHandler" in E:/SWE/go-backend/broker_workspace.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/broker_workspace.go` (Lines 276-276):
```
func newSandboxBrokerHandler(token string, evaluator agent.CandidateEvaluator, workspaces *brokerWorkspaceManager, limit
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 102. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_deliverables_go-NewJSONStore-L105`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "NewJSONStore" in E:/SWE/go-backend/deliverables.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/deliverables.go` (Lines 105-105):
```
func NewJSONStore(filePath string) (*JSONStore, error) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 103. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_deliverables_go-initDeliverableStore-L155`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "initDeliverableStore" in E:/SWE/go-backend/deliverables.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/deliverables.go` (Lines 155-155):
```
func initDeliverableStore() DeliverableStore {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 104. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_evolve_go-evolveHandler-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "evolveHandler" in E:/SWE/go-backend/evolve.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/evolve.go` (Lines 42-42):
```
func evolveHandler(workspaces *workspaceStore, snapshots ...SnapshotWorkspaceProvider) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 105. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_main_go-analyzeHandler-L341`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "analyzeHandler" in E:/SWE/go-backend/main.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 341-341):
```
func analyzeHandler(enginePath string, engineTimeout time.Duration, maxUploadBytes int64) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 106. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_sandbox_lifecycle_go-Reconcile-L307`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "Reconcile" in E:/SWE/go-backend/sandbox_lifecycle.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 307-307):
```
func (l *SandboxLifecycle) Reconcile(ctx context.Context, now time.Time) error {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 107. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_go_backend_snapshot_scanner_go-scanRepoHandlerWithRunner-L32`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "scanRepoHandlerWithRunner" in E:/SWE/go-backend/snapshot_scanner.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_scanner.go` (Lines 32-32):
```
func scanRepoHandlerWithRunner(
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 108. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_OrbitControls_js-onMouseDown-L810`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "onMouseDown" in E:/SWE/graph-visualizer/OrbitControls.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 810-810):
```
function onMouseDown( event ) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 109. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_OrbitControls_js-onTouchStart-L919`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "onTouchStart" in E:/SWE/graph-visualizer/OrbitControls.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 919-919):
```
function onTouchStart( event ) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 110. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_OrbitControls_js-update-L166`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "update" in E:/SWE/graph-visualizer/OrbitControls.js reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 166-166):
```
function update() {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 111. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_graph_mjs-<anonymous@312>-L312`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@312>" in E:/SWE/graph-visualizer/graph.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 312-312):
```
event => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 112. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_graph_mjs-draw-L44`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "draw" in E:/SWE/graph-visualizer/graph.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 44-44):
```
function draw() {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 113. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_routes_mjs-<anonymous@86>-L86`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@86>" in E:/SWE/graph-visualizer/routes.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 86-86):
```
link => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 114. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_graph_visualizer_routes_mjs-simplify-L5`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "simplify" in E:/SWE/graph-visualizer/routes.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 5-5):
```
function simplify(points) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 115. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_infra_analysis_stream_proxy_index_mjs-<anonymous@219>-L219`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@219>" in E:/SWE/infra/analysis/stream-proxy/index.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 219-219):
```
async (event, rawResponseStream) => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 116. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_scripts_analysis_cleanup_mjs-awsProcessEnvironment-L69`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "awsProcessEnvironment" in E:/SWE/scripts/analysis-cleanup.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 69-69):
```
function awsProcessEnvironment(testMode) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 117. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_scripts_analysis_cleanup_mjs-inventory-L380`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "inventory" in E:/SWE/scripts/analysis-cleanup.mjs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 380-380):
```
function inventory(stackName, prefix, region, expectedStackId) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 118. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-check-L870`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "check" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 870-870):
```
fn check(node: tree_sitter::Node, source: &[u8], lang: Language, name: &str) -> bool {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 119. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-check-L949`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "check" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 949-949):
```
fn check(node: tree_sitter::Node, source: &[u8], lang: Language, in_loop: bool) -> bool {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 120. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-collect_functions-L762`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "collect_functions" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 762-762):
```
fn collect_functions(
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 121. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-count-L912`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "count" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 912-912):
```
fn count(node: tree_sitter::Node, source: &[u8], lang: Language) -> usize {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 122. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-find_func-L1546`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "find_func" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 1546-1546):
```
fn find_func(node: tree_sitter::Node) -> Option<tree_sitter::Node> {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 123. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-is_concurrency_primitive-L341`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "is_concurrency_primitive" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 341-341):
```
fn is_concurrency_primitive(kind: &str, text: &str, lang: Language) -> bool {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 124. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-visit-L490`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "visit" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 490-490):
```
fn visit(&mut self, node: tree_sitter::Node) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 125. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-visitor_allocation_in_loop-L944`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "visitor_allocation_in_loop" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 944-944):
```
fn visitor_allocation_in_loop(
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 126. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_main_rs-visitor_has_recursive_call-L864`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "visitor_has_recursive_call" in E:/SWE/thermodynamic-ast-engine/src/main.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 864-864):
```
fn visitor_has_recursive_call(
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 127. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-normalize-L242`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "normalize" in E:/SWE/thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 242-242):
```
fn normalize(path: &str) -> String {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 128. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-walk-L194`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "walk" in E:/SWE/thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 194-194):
```
fn walk(node: Node, source: &[u8], lang: Language, names: &mut BTreeSet<String>) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 129. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_crawler_go-BuildIndex-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "BuildIndex" in E:/SWE/thermodynamic-ast-engine/test_samples/crawler.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/crawler.go` (Lines 79-79):
```
func BuildIndex(results []CrawlResult) map[string][]string {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 130. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_crawler_go-RecursiveMerge-L113`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "RecursiveMerge" in E:/SWE/thermodynamic-ast-engine/test_samples/crawler.go reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/crawler.go` (Lines 113-113):
```
func RecursiveMerge(dst, src map[string]interface{}) map[string]interface{} {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 131. [MEDIUM] Excessive control flow nesting depth (4) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_data_pipeline_py-recursive_flatten-L58`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "recursive_flatten" in E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 4 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 58-58):
```
def recursive_flatten(data, depth=0):
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 132. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_agent_annealer_go-Evolve-L99`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "Evolve" in E:/SWE/go-backend/agent/annealer.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/annealer.go` (Lines 99-99):
```
func (a *Annealer) Evolve(ctx context.Context, repoPath, filePath, funcName string, startByte, endByte int, baseEnergy f
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 133. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_agent_energy_go-extractHotspotMetric-L299`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "extractHotspotMetric" in E:/SWE/go-backend/agent/energy.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy.go` (Lines 299-299):
```
func extractHotspotMetric(report *ThermodynamicReport, targetRelPath string, funcName string) (*EnergyMeasurement, error
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 134. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_architecture_go-detectCycles-L103`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "detectCycles" in E:/SWE/go-backend/agent/repo_intel/architecture.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/architecture.go` (Lines 103-103):
```
func (aa *ArchitecturalAnalyzer) detectCycles(graph DependencyGraph) []Finding {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 135. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_manifests_go-parseCargoToml-L225`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "parseCargoToml" in E:/SWE/go-backend/agent/repo_intel/manifests.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 225-225):
```
func parseCargoToml(fullPath, relPath string) ([]ManifestDependency, []Finding) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 136. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-flattenNestedIf-L419`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "flattenNestedIf" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 419-419):
```
func flattenNestedIf(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 137. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_go_backend_evolve_go-streamEvolutionEvents-L206`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "streamEvolutionEvents" in E:/SWE/go-backend/evolve.go reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/evolve.go` (Lines 206-206):
```
func streamEvolutionEvents(
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 138. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_graph_visualizer_graph_mjs-<anonymous@330>-L330`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@330>" in E:/SWE/graph-visualizer/graph.mjs reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 330-330):
```
event => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 139. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_graph_visualizer_model_mjs-buildGraphModel-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "buildGraphModel" in E:/SWE/graph-visualizer/model.mjs reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/model.mjs` (Lines 23-23):
```
function buildGraphModel(files, { query = '', language = '', focusPath = '', showFunctions = false, sort = 'path', relat
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 140. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_graph_visualizer_routes_mjs-obstacleIndex-L31`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "obstacleIndex" in E:/SWE/graph-visualizer/routes.mjs reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 31-31):
```
function obstacleIndex(nodes) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 141. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_data_pipeline_py-build_feature_matrix-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "build_feature_matrix" in E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 41-41):
```
def build_feature_matrix(records: list) -> np.ndarray:
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 142. [MEDIUM] Excessive control flow nesting depth (5) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_data_pipeline_py-fetch_user_data-L18`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "fetch_user_data" in E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 5 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 18-18):
```
def fetch_user_data(user_ids: list) -> dict:
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 143. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_api_analysis_proxy_go-registerAnalysisRoutes-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "registerAnalysisRoutes" in E:/SWE/api/analysis_proxy.go reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy.go` (Lines 42-42):
```
func registerAnalysisRoutes(mux *http.ServeMux, deps Dependencies) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 144. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_aggregator_go-GenerateMarkdown-L86`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "GenerateMarkdown" in E:/SWE/go-backend/agent/repo_intel/aggregator.go reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/aggregator.go` (Lines 86-86):
```
func (a *Aggregator) GenerateMarkdown(report *IntelligenceReport) string {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 145. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-checkBufferEscapes-L910`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "checkBufferEscapes" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 910-910):
```
func checkBufferEscapes(forStmt *ast.ForStmt, allocStmt ast.Stmt, bufName string) bool {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 146. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-extractCommonSubexpressions-L705`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "extractCommonSubexpressions" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 705-705):
```
func extractCommonSubexpressions(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 147. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_graph_visualizer_routes_mjs-<anonymous@42>-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "<anonymous@42>" in E:/SWE/graph-visualizer/routes.mjs reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 42-42):
```
(points, source, target) => {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 148. [MEDIUM] Excessive control flow nesting depth (6) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_test_samples_data_pipeline_py-process_batch-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "process_batch" in E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py reaches nesting depth 6 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 79-79):
```
def process_batch(self, batch_ids: list) -> list:
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 149. [MEDIUM] Excessive control flow nesting depth (7) (`RI-STR-NEST-E:_SWE_go_backend_agent_repo_intel_orchestrator_go-Run-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "Run" in E:/SWE/go-backend/agent/repo_intel/orchestrator.go reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/orchestrator.go` (Lines 42-42):
```
func (o *Orchestrator) Run(ctx context.Context) (*IntelligenceReport, error) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 150. [MEDIUM] Excessive control flow nesting depth (7) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-removeElseAfterReturn-L466`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "removeElseAfterReturn" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 466-466):
```
func removeElseAfterReturn(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 151. [MEDIUM] Excessive control flow nesting depth (7) (`RI-STR-NEST-E:_SWE_samples_benchmark_benchmark_target_go-EvaluateRiskEngine-L21`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "EvaluateRiskEngine" in E:/SWE/samples/benchmark/benchmark_target.go reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/samples/benchmark/benchmark_target.go` (Lines 21-21):
```
func EvaluateRiskEngine(records []Transaction, threshold float64) EngineReport {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 152. [MEDIUM] Excessive control flow nesting depth (7) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-build-L260`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "build" in E:/SWE/thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 260-260):
```
pub fn build(files: &[FileReport], root: &Path) -> Vec<Relationship> {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 153. [MEDIUM] Excessive control flow nesting depth (7) (`RI-STR-NEST-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-walk-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "walk" in E:/SWE/thermodynamic-ast-engine/src/relationships.rs reaches nesting depth 7 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 41-41):
```
fn walk(node: Node, source: &[u8], lang: Language, found: &mut BTreeSet<(String, String)>) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 154. [MEDIUM] Excessive control flow nesting depth (8) (`RI-STR-NEST-E:_SWE_go_backend_agent_strategy_go-simplifyBooleanReturns-L511`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Deeply nested if/for/switch blocks without early returns or helper decomposition.

Function "simplifyBooleanReturns" in E:/SWE/go-backend/agent/strategy.go reaches nesting depth 8 (threshold is 3). Deep nesting impairs cognitive readability and increases cyclomatic volatility.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 511-511):
```
func simplifyBooleanReturns(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Refactor with guard clauses, early returns, or decompose into helper functions.
>
> **Expected Benefit:** Lowers thermodynamic structural energy and simplifies test branch coverage.

---

### 155. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_components_AdminPortal_jsx-handlePublish-L225`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "handlePublish" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/AdminPortal.jsx` (Lines 225-225):
```
async () => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 156. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_analysis_authorizer_main_test_go-TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer-L99`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/main_test.go` (Lines 99-99):
```
func TestAuthorizerHandler_ReturnsInternalTokenOnlyForValidBearer(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 157. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-callback-L202`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "callback" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 202-202):
```
func (a *githubAccountService) callback(w http.ResponseWriter, r *http.Request) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 158. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_accounts_test_go-testAccountService-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "testAccountService" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 23-23):
```
func testAccountService(t *testing.T) (*githubAccountService, http.Handler) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 159. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_archive_go-Delete-L272`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "Delete" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 272-272):
```
func (s *archiveService) Delete(ctx context.Context, id string) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 160. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_auth_go-decodeAndVerify-L129`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "decodeAndVerify" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/auth.go` (Lines 129-129):
```
func (a *HMACTokenAuthenticator) decodeAndVerify(token string) ([]byte, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 161. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_aws_store_test_go-TestArchiveServiceBeginUploadCreatesPrivatePendingRecord-L26`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestArchiveServiceBeginUploadCreatesPrivatePendingRecord" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 26-26):
```
func TestArchiveServiceBeginUploadCreatesPrivatePendingRecord(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 162. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_main_go-runMigrations-L441`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "runMigrations" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/main.go` (Lines 441-441):
```
func runMigrations(pool *pgxpool.Pool) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 163. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_api_refresh_test_go-TestLoginHandlerReportsServerRelativeSessionLifetime-L52`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestLoginHandlerReportsServerRelativeSessionLifetime" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/refresh_test.go` (Lines 52-52):
```
func TestLoginHandlerReportsServerRelativeSessionLifetime(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 164. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_evaluator_go-PrepareWorkspace-L124`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "PrepareWorkspace" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_evaluator.go` (Lines 124-124):
```
func (r *RemoteBrokerEvaluator) PrepareWorkspace(ctx context.Context, repoPath string) (string, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 165. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_test_go-TestCIWorkflow_ThreeOutcomesSimulation-L1026`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestCIWorkflow_ThreeOutcomesSimulation" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy_test.go` (Lines 1026-1026):
```
func TestCIWorkflow_ThreeOutcomesSimulation(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 166. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_workspace_go-validateWorkspaceArchiveEntry-L260`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateWorkspaceArchiveEntry" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 260-260):
```
func validateWorkspaceArchiveEntry(file *zip.File) (string, bool, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 167. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-ReconcileStaleLeases-L674`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ReconcileStaleLeases" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 674-674):
```
func (s *DynamoSandboxLifecycleStore) ReconcileStaleLeases(ctx context.Context, now int64) (SandboxLifecycleRecord, erro
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 168. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-stopClaimedInstance-L364`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "stopClaimedInstance" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 364-364):
```
func (l *SandboxLifecycle) stopClaimedInstance(ctx context.Context) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 169. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_account_mjs-<anonymous@94>-L94`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@94>" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/account.mjs` (Lines 94-94):
```
async () => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 170. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_graph_mjs-<anonymous@330>-L330`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@330>" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 330-330):
```
event => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 171. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_model_mjs-<anonymous@26>-L26`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@26>" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/model.mjs` (Lines 26-26):
```
file => (!folderPath || directory(file) === folderPath) && (!language || file.language === language)
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 172. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-applyMode-L546`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "applyMode" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 546-546):
```
async function applyMode(options) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 173. [MEDIUM] Excessive cyclomatic complexity (M=12) (`RI-STR-COMPLEX-E:_SWE_source_stager_main_go-StageRepository-L206`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "StageRepository" has cyclomatic complexity of 12, requiring at least 12 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 206-206):
```
func (s *RepositoryStager) StageRepository(ctx context.Context, repositoryURL string) (string, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 174. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_archive_download_js-fetchArchiveDownload-L1`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "fetchArchiveDownload" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/archive-download.js` (Lines 1-1):
```
async (apiBase, downloadURL, token, fetchImpl = fetch) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 175. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_evolution_api_js-processLine-L20`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "processLine" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-api.js` (Lines 20-20):
```
(line) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 176. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_evolution_measurements_js-evolutionMeasurements-L1`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "evolutionMeasurements" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-measurements.js` (Lines 1-1):
```
function evolutionMeasurements(initialEnergy, telemetry, result) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 177. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_api_accounts_test_go-beginAccountSignIn-L80`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "beginAccountSignIn" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 80-80):
```
func beginAccountSignIn(t *testing.T, handler http.Handler) (string, *http.Cookie, string) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 178. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_api_archive_go-CompleteUpload-L175`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "CompleteUpload" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 175-175):
```
func (s *archiveService) CompleteUpload(ctx context.Context, id string) (ArchiveEntry, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 179. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_api_local_store_go-PromoteUpload-L189`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "PromoteUpload" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 189-189):
```
func (s *LocalObjectStore) PromoteUpload(_ context.Context, pendingKey, archiveKey string) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 180. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_test_go-TestRemoteBrokerEvaluator_PrepareWorkspace-L525`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestRemoteBrokerEvaluator_PrepareWorkspace" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_test.go` (Lines 525-525):
```
func TestRemoteBrokerEvaluator_PrepareWorkspace(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 181. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_test_go-TestRemoteBrokerEvaluator_SuccessfulEvaluation-L322`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestRemoteBrokerEvaluator_SuccessfulEvaluation" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_test.go` (Lines 322-322):
```
func TestRemoteBrokerEvaluator_SuccessfulEvaluation(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 182. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_energy_test_go-TestAnnealer_DownhillImprovement-L197`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestAnnealer_DownhillImprovement" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy_test.go` (Lines 197-197):
```
func TestAnnealer_DownhillImprovement(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 183. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-stmtMutatesIdent-L614`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "stmtMutatesIdent" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 614-614):
```
func stmtMutatesIdent(stmt ast.Stmt, name string) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 184. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_broker_workspace_go-AcquireWorkspace-L137`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "AcquireWorkspace" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/broker_workspace.go` (Lines 137-137):
```
func (m *brokerWorkspaceManager) AcquireWorkspace(workspaceID string) (string, func() error, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 185. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_main_go-auditHandler-L503`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "auditHandler" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 503-503):
```
func auditHandler(enginePath string, engineTimeout time.Duration, maxUploadBytes int64) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 186. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-Acquire-L155`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "Acquire" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 155-155):
```
func (l *SandboxLifecycle) Acquire(ctx context.Context) (func() error, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 187. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-sandboxLifecycleFromEnv-L895`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "sandboxLifecycleFromEnv" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 895-895):
```
func sandboxLifecycleFromEnv(ctx context.Context) (*SandboxLifecycle, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 188. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-collect_functions-L762`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "collect_functions" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 762-762):
```
fn collect_functions(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 189. [MEDIUM] Excessive cyclomatic complexity (M=13) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_test_samples_data_pipeline_py-process_batch-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "process_batch" has cyclomatic complexity of 13, requiring at least 13 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 79-79):
```
def process_batch(self, batch_ids: list) -> list:
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 190. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_evolution_api_js-readEvolutionStream-L5`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "readEvolutionStream" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/evolution-api.js` (Lines 5-5):
```
async (response, onEvent) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 191. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_analysis_authorizer_main_test_go-TestSSMSecretStore_RequiresDecryptedSecureString-L227`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestSSMSecretStore_RequiresDecryptedSecureString" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/main_test.go` (Lines 227-227):
```
func TestSSMSecretStore_RequiresDecryptedSecureString(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 192. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_api_analysis_proxy_test_go-TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide-L16`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy_test.go` (Lines 16-16):
```
func TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 193. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_api_aws_store_go-validGeneratedObjectKey-L346`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validGeneratedObjectKey" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/aws_store.go` (Lines 346-346):
```
func validGeneratedObjectKey(key, prefix string) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 194. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_api_local_store_go-read-L385`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "read" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 385-385):
```
func (r *localArchiveRepository) read() (map[string]ArchiveEntry, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 195. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_architecture_go-detectCycles-L103`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "detectCycles" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/architecture.go` (Lines 103-103):
```
func (aa *ArchitecturalAnalyzer) detectCycles(graph DependencyGraph) []Finding {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 196. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_manifests_go-parseGoMod-L154`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "parseGoMod" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 154-154):
```
func parseGoMod(fullPath, relPath string) ([]ManifestDependency, []Finding) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 197. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_repo_intel_test_go-TestDiscovery_GitIgnoreAndFilter-L89`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestDiscovery_GitIgnoreAndFilter" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/repo_intel_test.go` (Lines 89-89):
```
func TestDiscovery_GitIgnoreAndFilter(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 198. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-checkUsesLoopVar-L1017`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "checkUsesLoopVar" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 1017-1017):
```
func checkUsesLoopVar(forStmt *ast.ForStmt, sizeExpr ast.Expr) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 199. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_broker_go-clampLimits-L72`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "clampLimits" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/broker.go` (Lines 72-72):
```
func clampLimits(limits agent.ResourceLimits) agent.ResourceLimits {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 200. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_go_backend_workspace_test_go-TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes-L52`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/workspace_test.go` (Lines 52-52):
```
func TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 201. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-assertStackDeleted-L261`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "assertStackDeleted" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 261-261):
```
function assertStackDeleted(stackName, region, expectedStackId) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 202. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-normalizePlan-L221`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "normalizePlan" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 221-221):
```
function normalizePlan(plan) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 203. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-count-L912`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "count" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 912-912):
```
fn count(node: tree_sitter::Node, source: &[u8], lang: Language) -> usize {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 204. [MEDIUM] Excessive cyclomatic complexity (M=14) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-is_boolean_operator-L265`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "is_boolean_operator" has cyclomatic complexity of 14, requiring at least 14 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 265-265):
```
fn is_boolean_operator(kind: &str, text: &str, lang: Language) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 205. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_admin_portal_tests_graph_model_test_js-onBorder-L7`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "onBorder" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/tests/graph-model.test.js` (Lines 7-7):
```
(node, [x, y]) => x >= node.x && x <= node.x + node.width && y >= node.y && y <= node.y + node.height
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 206. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_analysis_authorizer_main_go-VerifyArchiveBearer-L66`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "VerifyArchiveBearer" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/main.go` (Lines 66-66):
```
func VerifyArchiveBearer(token, signingKey string, now time.Time) bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 207. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_api_archive_go-validateUploadMetadata-L437`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateUploadMetadata" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 437-437):
```
func validateUploadMetadata(name, contentType string, size int64) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 208. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_api_main_test_go-TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation-L345`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/main_test.go` (Lines 345-345):
```
func TestInjectedAuthenticatorProtectsEveryArchiveServiceOperation(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 209. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_api_refresh_test_go-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/refresh_test.go` (Lines 82-82):
```
func TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 210. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_evaluator_go-NewRemoteBrokerEvaluator-L45`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "NewRemoteBrokerEvaluator" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_evaluator.go` (Lines 45-45):
```
func NewRemoteBrokerEvaluator(brokerURL, token string) *RemoteBrokerEvaluator {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 211. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_evaluator_go-copyProjectFiles-L447`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "copyProjectFiles" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/evaluator.go` (Lines 447-447):
```
func copyProjectFiles(srcDir, destDir string) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 212. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_manifests_go-AnalyzeManifests-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "AnalyzeManifests" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 23-23):
```
func (ma *ManifestAnalyzer) AnalyzeManifests(manifestPaths []string) ([]ManifestInfo, []ServiceBoundary, []Finding) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 213. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_go-flattenNestedIf-L419`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "flattenNestedIf" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 419-419):
```
func flattenNestedIf(src string) (string, bool) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 214. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-decodeSandboxLifecycleRecord-L465`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "decodeSandboxLifecycleRecord" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 465-465):
```
func decodeSandboxLifecycleRecord(item map[string]types.AttributeValue) (SandboxLifecycleRecord, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 215. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_go_backend_snapshot_store_go-ExtractSnapshot-L77`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ExtractSnapshot" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_store.go` (Lines 77-77):
```
func (s *SnapshotStore) ExtractSnapshot(ctx context.Context, snapshotID string) (string, func(), error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 216. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_OrbitControls_js-onTouchMove-L975`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "onTouchMove" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 975-975):
```
function onTouchMove( event ) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 217. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_routes_mjs-<anonymous@86>-L86`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@86>" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 86-86):
```
link => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 218. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_routes_mjs-simplify-L5`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "simplify" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 5-5):
```
function simplify(points) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 219. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-buildUpstreamBody-L116`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "buildUpstreamBody" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 116-116):
```
function buildUpstreamBody(route, payload, snapshotID) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 220. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-parseRequestBody-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "parseRequestBody" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 79-79):
```
function parseRequestBody(event, maxBytes) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 221. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_tests_sandbox_host_validation_test_mjs-run-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "run" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/tests/sandbox-host-validation.test.mjs` (Lines 23-23):
```
(command, args) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 222. [MEDIUM] Excessive cyclomatic complexity (M=15) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-walk-L194`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "walk" has cyclomatic complexity of 15, requiring at least 15 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 194-194):
```
fn walk(node: Node, source: &[u8], lang: Language, names: &mut BTreeSet<String>) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 223. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_admin_portal_tests_graph_model_test_js-<anonymous@58>-L58`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@58>" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/tests/graph-model.test.js` (Lines 58-58):
```
() => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 224. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_api_accounts_go-exchangeGitHubCode-L243`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "exchangeGitHubCode" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 243-243):
```
func (a *githubAccountService) exchangeGitHubCode(ctx context.Context, code, verifier string) (GitHubUser, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 225. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_api_aws_store_go-VerifyUpload-L242`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "VerifyUpload" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/aws_store.go` (Lines 242-242):
```
func (s *S3ObjectStore) VerifyUpload(ctx context.Context, pendingKey, originalName, contentType string, expectedSize int
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 226. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_api_aws_store_test_go-TestS3ObjectStorePresignsBoundedPendingUploadAndDownload-L689`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestS3ObjectStorePresignsBoundedPendingUploadAndDownload" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 689-689):
```
func TestS3ObjectStorePresignsBoundedPendingUploadAndDownload(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 227. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_api_lambda_go-loadSSMSecureString-L77`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "loadSSMSecureString" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/lambda.go` (Lines 77-77):
```
func loadSSMSecureString(ctx context.Context, client ssmParameterAPI, name string) (string, error) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 228. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_broker_test_go-TestSandboxRunnerIsolationAndCIPinsImage-L30`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestSandboxRunnerIsolationAndCIPinsImage" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_test.go` (Lines 30-30):
```
func TestSandboxRunnerIsolationAndCIPinsImage(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 229. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_aggregator_go-GenerateMarkdown-L86`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "GenerateMarkdown" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/aggregator.go` (Lines 86-86):
```
func (a *Aggregator) GenerateMarkdown(report *IntelligenceReport) string {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 230. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_strategy_test_go-TestStrategy_DifferentialEquivalence-L610`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestStrategy_DifferentialEquivalence" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy_test.go` (Lines 610-610):
```
func TestStrategy_DifferentialEquivalence(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 231. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_go_backend_main_go-analyzeHandler-L341`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "analyzeHandler" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 341-341):
```
func analyzeHandler(enginePath string, engineTimeout time.Duration, maxUploadBytes int64) gin.HandlerFunc {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 232. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_go_backend_sandbox_lifecycle_go-ensureReady-L217`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "ensureReady" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 217-217):
```
func (l *SandboxLifecycle) ensureReady(ctx context.Context, startOwner bool) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 233. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-aws-L93`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "aws" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 93-93):
```
function aws(region, service, operation, args = []) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 234. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-emptyBucket-L488`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "emptyBucket" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 488-488):
```
async function emptyBucket(resource, stackName, stackId, prefix, region) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 235. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_serve_js-<anonymous@76>-L76`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@76>" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/serve.js` (Lines 76-76):
```
(req, res) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 236. [MEDIUM] Excessive cyclomatic complexity (M=16) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_languages_rs-grammar-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "grammar" has cyclomatic complexity of 16, requiring at least 16 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/languages.rs` (Lines 23-23):
```
pub fn grammar(self) -> Option<tree_sitter::Language> {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 237. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_admin_portal_src_archive_api_js-isAuthenticatedSession-L66`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isAuthenticatedSession" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/archive-api.js` (Lines 66-66):
```
(payload) =>
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 238. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_admin_portal_tests_graph_model_test_js-assertClearConnections-L6`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "assertClearConnections" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/admin-portal/tests/graph-model.test.js` (Lines 6-6):
```
function assertClearConnections({nodes, links}) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 239. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_api_aws_store_test_go-TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls-L726`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 726-726):
```
func TestS3SDKGeneratesPOSTPolicyWithoutAWSCalls(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 240. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_api_local_store_go-StageUpload-L85`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "StageUpload" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 85-85):
```
func (s *LocalObjectStore) StageUpload(ctx context.Context, id, contentType string, body io.Reader) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 241. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_workspace_test_go-TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles-L13`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace_test.go` (Lines 13-13):
```
func TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 242. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_go_backend_main_go-extractZip-L165`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "extractZip" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 165-165):
```
func extractZip(zipPath, destDir string) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 243. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-isPrivateIPv4-L16`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "isPrivateIPv4" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 16-16):
```
function isPrivateIPv4(address) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 244. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_samples_benchmark_benchmark_target_go-EvaluateRiskEngine-L21`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "EvaluateRiskEngine" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/samples/benchmark/benchmark_target.go` (Lines 21-21):
```
func EvaluateRiskEngine(records []Transaction, threshold float64) EngineReport {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 245. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-check-L870`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "check" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 870-870):
```
fn check(node: tree_sitter::Node, source: &[u8], lang: Language, name: &str) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 246. [MEDIUM] Excessive cyclomatic complexity (M=17) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-determine_primary_driver-L986`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "determine_primary_driver" has cyclomatic complexity of 17, requiring at least 17 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 986-986):
```
fn determine_primary_driver(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 247. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_analysis_authorizer_rest_authorizer_test_go-TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer-L13`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/analysis-authorizer/rest_authorizer_test.go` (Lines 13-13):
```
func TestAuthorizerHandler_ReturnsScopedRESTPolicyForValidBearer(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 248. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_api_accounts_test_go-exerciseAccountStore-L305`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "exerciseAccountStore" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 305-305):
```
func exerciseAccountStore(t *testing.T, store accountStore) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 249. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_api_refresh_test_go-TestRefreshHandlerRenewsActiveBearerToken-L14`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestRefreshHandlerRenewsActiveBearerToken" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/refresh_test.go` (Lines 14-14):
```
func TestRefreshHandlerRenewsActiveBearerToken(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 250. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_go_backend_agent_repo_intel_reliability_security_go-AnalyzeReliabilityAndSecurity-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "AnalyzeReliabilityAndSecurity" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/reliability_security.go` (Lines 42-42):
```
func (rsa *ReliabilitySecurityAnalyzer) AnalyzeReliabilityAndSecurity(files []DiscoveredFile) []Finding {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 251. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_routes_mjs-<anonymous@42>-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@42>" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/routes.mjs` (Lines 42-42):
```
(points, source, target) => {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 252. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_infra_analysis_stream_proxy_index_mjs-validateAnalysisBaseURL-L32`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateAnalysisBaseURL" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 32-32):
```
function validateAnalysisBaseURL(value) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 253. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_scripts_validate_sandbox_host_mjs-validateSandboxHost-L14`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "validateSandboxHost" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/validate-sandbox-host.mjs` (Lines 14-14):
```
({
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 254. [MEDIUM] Excessive cyclomatic complexity (M=18) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-check-L949`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "check" has cyclomatic complexity of 18, requiring at least 18 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 949-949):
```
fn check(node: tree_sitter::Node, source: &[u8], lang: Language, in_loop: bool) -> bool {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 255. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_api_account_store_go-access-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "access" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/account_store.go` (Lines 79-79):
```
func (s *localAccountStore) access(ctx context.Context, write bool, action func(*localAccountData) error) error {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 256. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_api_accounts_test_go-TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart-L149`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 149-149):
```
func TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart(t *testing.T) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 257. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_api_local_store_go-registerLocalFileRoutes-L505`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "registerLocalFileRoutes" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 505-505):
```
func registerLocalFileRoutes(mux *http.ServeMux, deps Dependencies) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 258. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_go_backend_snapshot_scanner_go-scanRepoHandlerWithRunner-L32`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "scanRepoHandlerWithRunner" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_scanner.go` (Lines 32-32):
```
func scanRepoHandlerWithRunner(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 259. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_graph_visualizer_OrbitControls_js-onTouchStart-L919`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "onTouchStart" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 919-919):
```
function onTouchStart( event ) {
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 260. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_scripts_analysis_cleanup_mjs-<anonymous@197>-L197`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "<anonymous@197>" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/scripts/analysis-cleanup.mjs` (Lines 197-197):
```
(attachment) =>
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 261. [MEDIUM] Excessive cyclomatic complexity (M=19) (`RI-STR-COMPLEX-E:_SWE_thermodynamic_ast_engine_src_main_rs-scan_directory-L1079`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `CONFIRMED` (Confidence: `0.95`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Accumulation of branching logic and conditional checks in a single function.

Function "scan_directory" has cyclomatic complexity of 19, requiring at least 19 branch test cases to achieve full coverage.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 1079-1079):
```
pub fn scan_directory(
```

> [!TIP]
> **Recommended Action:** Extract distinct condition branches into cohesive helper functions or use polymorphism/table dispatch.
>
> **Expected Benefit:** Improves maintainability and reduces likelihood of latent edge-case bugs.

---

### 262. [MEDIUM] Rust codebase detected without dedicated Rust CI workflow (`RI-QUAL-CI-MISSING-RUST`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `PROBABLE` (Confidence: `0.80`)
- **Detection Method:** `CI_COVERAGE_GAP`
- **Root Cause:** CI coverage gap for Rust toolchain.

Repository contains Rust source files but no CI workflow explicitly dedicated to cargo test/check.

**Evidence Locations:**

- `.github/workflows/check-frontend.yml` — *Existing CI workflows omit cargo test/check steps*

> [!TIP]
> **Recommended Action:** Ensure Rust compilation and cargo test are included in CI workflow steps.
>
> **Expected Benefit:** Validates that native Rust engine passes all unit tests.

---

### 263. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_account_store_go-L128`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/account_store.go at line 128 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/account_store.go` (Lines 128-128) — *_ = temporary.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 264. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_account_store_go-L132`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/account_store.go at line 132 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/account_store.go` (Lines 132-132) — *_ = temporary.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 265. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_account_store_go-L89`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/account_store.go at line 89 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/account_store.go` (Lines 89-89) — *_ = file.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 266. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_accounts_go-L77`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/accounts.go at line 77 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/accounts.go` (Lines 77-77) — *_ = json.NewEncoder(w).Encode(value)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 267. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L125`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 125 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 125-125) — *_ = os.Remove(stagedPath)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 268. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L130`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 130 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 130-130) — *_ = staged.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 269. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L134`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 134 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 134-134) — *_ = staged.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 270. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L138`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 138 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 138-138) — *_ = staged.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 271. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L219`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 219 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 219-219) — *_ = os.Remove(target)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 272. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L223`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 223 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 223-223) — *_ = targetFile.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 273. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L227`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 227 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 227-227) — *_ = targetFile.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 274. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L456`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 456 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 456-456) — *_ = temporary.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 275. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L460`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 460 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 460-460) — *_ = temporary.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 276. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_local_store_go-L486`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/local_store.go at line 486 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/local_store.go` (Lines 486-486) — *_ = os.Remove(backupPath)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 277. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L124`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 124 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 124-124) — *_ = json.NewEncoder(w).Encode(upload)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 278. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L152`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 152 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 152-152) — *_ = json.NewEncoder(w).Encode(entry)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 279. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L174`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 174 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 174-174) — *_ = json.NewEncoder(w).Encode(entries)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 280. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L198`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 198 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 198-198) — *_ = json.NewEncoder(w).Encode(download)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 281. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L278`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 278 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 278-278) — *_ = json.NewEncoder(w).Encode(map[string]any{*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 282. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L320`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 320 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 320-320) — *_ = json.NewEncoder(w).Encode(map[string]string{"status": "authenticated"})*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 283. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L35`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 35 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 35-35) — *_ = json.NewEncoder(w).Encode(map[string]string{"error": code})*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 284. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L359`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 359 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 359-359) — *_ = json.NewEncoder(w).Encode(map[string]any{*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 285. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L542`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 542 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 542-542) — *_ = json.NewEncoder(w).Encode(entry)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 286. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-api_router_go-L584`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File api/router.go at line 584 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `api/router.go` (Lines 584-584) — *_ = json.NewEncoder(w).Encode(entries)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 287. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_annealer_go-L510`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/annealer.go at line 510 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/annealer.go` (Lines 510-510) — *_ = os.Chmod(tmpName, perm)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 288. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_broker_test_go-L352`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/broker_test.go at line 352 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/broker_test.go` (Lines 352-352) — *_ = json.NewEncoder(w).Encode(map[string]any{*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 289. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_broker_test_go-L399`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/broker_test.go at line 399 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/broker_test.go` (Lines 399-399) — *_ = json.NewEncoder(w).Encode(map[string]any{*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 290. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_energy_test_go-L693`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/energy_test.go at line 693 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/energy_test.go` (Lines 693-693) — *_ = cache.Size()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 291. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_go-L280`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator.go at line 280 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator.go` (Lines 280-280) — *_ = os.MkdirAll(baseWS, 0o777)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 292. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_go-L292`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator.go at line 292 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator.go` (Lines 292-292) — *_ = os.Chmod(tempWorkspace, 0o777)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 293. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_go-L380`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator.go at line 380 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator.go` (Lines 380-380) — *_ = exec.CommandContext(killCtx, d.dockerBinary, "rm", "-f", containerName).Run()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 294. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L407`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 407 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 407-407) — *_ = os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 295. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L408`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 408 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 408-408) — *_ = os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module test"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 296. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L411`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 411 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 411-411) — *_ = os.WriteFile(filepath.Join(srcDir, ".env"), []byte("SECRET=123"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 297. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L412`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 412 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 412-412) — *_ = os.WriteFile(filepath.Join(srcDir, ".env.production"), []byte("AWS_SECRET=xyz"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 298. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L413`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 413 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 413-413) — *_ = os.WriteFile(filepath.Join(srcDir, "app.exe"), []byte("MZbinary"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 299. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L414`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 414 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 414-414) — *_ = os.WriteFile(filepath.Join(srcDir, "test.zip"), []byte("PKzip"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 300. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L417`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 417 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 417-417) — *_ = os.MkdirAll(gitDir, 0o750)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 301. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L418`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 418 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 418-418) — *_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 302. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L611`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 611 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 611-611) — *_ = os.WriteFile(secretFile, []byte("root:supersecret"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 303. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_evaluator_test_go-L614`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/evaluator_test.go at line 614 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/evaluator_test.go` (Lines 614-614) — *_ = os.WriteFile(filepath.Join(srcDir, "valid.go"), []byte("package main"), 0o600)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 304. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_repo_intel_repo_intel_test_go-L30`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/repo_intel/repo_intel_test.go at line 30 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/repo_intel/repo_intel_test.go` (Lines 30-30) — *_ = os.WriteFile(filepath.Join(dir, "target", "ignored.o"), []byte("binary data"), 0644)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 305. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_repo_intel_repo_intel_test_go-L31`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/repo_intel/repo_intel_test.go at line 31 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/repo_intel/repo_intel_test.go` (Lines 31-31) — *_ = os.WriteFile(filepath.Join(dir, "test.secret.txt"), []byte("super_secret"), 0644)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 306. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_repo_intel_repo_intel_test_go-L67`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/repo_intel/repo_intel_test.go at line 67 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/repo_intel/repo_intel_test.go` (Lines 67-67) — *_ = cmd.Run()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 307. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_strategy_go-L590`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/strategy.go at line 590 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/strategy.go` (Lines 590-590) — *_ = format.Node(&buf, fset, expr)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 308. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_agent_strategy_test_go-L409`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/agent/strategy_test.go at line 409 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/agent/strategy_test.go` (Lines 409-409) — *_ = os.WriteFile(targetFile, externalContent, 0o644)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 309. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_deliverables_go-L178`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/deliverables.go at line 178 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/deliverables.go` (Lines 178-178) — *_ = db.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 310. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_deliverables_go-L224`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/deliverables.go at line 224 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/deliverables.go` (Lines 224-224) — *_ = os.MkdirAll(uploadsDir, 0o750)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 311. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_go-L404`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main.go at line 404 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main.go` (Lines 404-404) — *_ = os.Remove(zipDest) // best-effort cleanup before early return*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 312. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_go-L538`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main.go at line 538 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main.go` (Lines 538-538) — *_ = os.Remove(zipDest)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 313. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_go-L546`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main.go at line 546 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main.go` (Lines 546-546) — *_ = os.RemoveAll(extractDir)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 314. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_go-L547`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main.go at line 547 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main.go` (Lines 547-547) — *_ = os.Remove(zipDest)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 315. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_go-L718`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main.go at line 718 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main.go` (Lines 718-718) — *_ = os.MkdirAll(uploadsDir, 0o750)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 316. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_test_go-L399`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main_test.go at line 399 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main_test.go` (Lines 399-399) — *_ = writer.WriteField("title", "Project Milestone 1")*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 317. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_test_go-L400`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main_test.go at line 400 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main_test.go` (Lines 400-400) — *_ = writer.WriteField("version", "1.0.0")*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 318. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_test_go-L403`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main_test.go at line 403 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main_test.go` (Lines 403-403) — *_ = writer.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 319. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_main_test_go-L462`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/main_test.go at line 462 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/main_test.go` (Lines 462-462) — *_ = writer.Close()*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 320. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_sandbox_lifecycle_test_go-L451`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/sandbox_lifecycle_test.go at line 451 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/sandbox_lifecycle_test.go` (Lines 451-451) — *_ = lifecycle.Reconcile(context.Background(), now.Add(24*time.Hour))*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 321. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_scanner_go-L69`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/scanner.go at line 69 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/scanner.go` (Lines 69-69) — *_ = os.MkdirAll(uploadsDir, 0o750)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 322. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_scanner_go-L85`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/scanner.go at line 85 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/scanner.go` (Lines 85-85) — *_ = os.RemoveAll(targetDir)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 323. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_workspace_go-L108`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/workspace.go at line 108 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/workspace.go` (Lines 108-108) — *_ = os.RemoveAll(w.path)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 324. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_workspace_go-L123`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/workspace.go at line 123 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/workspace.go` (Lines 123-123) — *_ = os.RemoveAll(w.path)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 325. [LOW] Ignored function error return value (`RI-REL-IGNORED-ERR-go_backend_workspace_go-L64`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `STATIC_ERROR_PATTERN`
- **Root Cause:** Error return value suppressed without handling or logging.

File go-backend/workspace.go at line 64 explicitly discards error return value via blank identifier '_'.

**Evidence Locations:**

- `go-backend/workspace.go` (Lines 64-64) — *_ = os.RemoveAll(path)*

> [!TIP]
> **Recommended Action:** Inspect returned error and propagate, log, or handle gracefully.
>
> **Expected Benefit:** Prevents silent failures and simplifies operational diagnosis.

---

### 326. [LOW] Missing repository LICENSE file (`RI-HYG-MISSING-LICENSE`)

- **Category:** `RELIABILITY`
- **Verification Status:** `CONFIRMED` (Confidence: `1.00`)
- **Detection Method:** `LICENSE_AUDIT`
- **Root Cause:** License file omitted.

No software license file (LICENSE or COPYING) was detected in repository root.

> [!TIP]
> **Recommended Action:** Add an explicit LICENSE file (e.g. Apache-2.0 or MIT) to govern code redistribution and usage.
>
> **Expected Benefit:** Clarifies legal rights and license boundaries.

---

### 327. [LOW] Dockerfile uses unpinned or :latest base image (`RI-SEC-DOCKER-LATEST-go_backend_Dockerfile-L24`)

- **Category:** `SECURITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.90`)
- **Detection Method:** `MANIFEST_ANALYSIS`
- **Root Cause:** Use of mutable :latest tag in container image specification.

Dockerfile at go-backend/Dockerfile references base image "alpine:latest" without an explicit version pin.

**Evidence Locations:**

- `go-backend/Dockerfile` (Lines 24-24) — *FROM alpine:latest*

> [!TIP]
> **Recommended Action:** Pin base image to an immutable tag or SHA256 digest.
>
> **Expected Benefit:** Guarantees build reproducibility and prevents surprise upstream base image vulnerabilities.

---

### 328. [LOW] Large function AST node volume (1082 nodes) (`RI-STR-VOLUME-E:_SWE_api_router_go-registerArchiveServiceRoutes-L80`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "registerArchiveServiceRoutes" contains 1082 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/router.go` (Lines 80-80) — *AST node count: 1082*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 329. [LOW] Large function AST node volume (1177 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_broker_evaluator_go-EvaluateCandidate-L209`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "EvaluateCandidate" contains 1177 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_evaluator.go` (Lines 209-209) — *AST node count: 1177*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 330. [LOW] Large function AST node volume (1185 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_aggregator_go-GenerateMarkdown-L86`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "GenerateMarkdown" contains 1185 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/aggregator.go` (Lines 86-86) — *AST node count: 1185*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 331. [LOW] Large function AST node volume (1211 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_quality_go-AnalyzeQuality-L21`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "AnalyzeQuality" contains 1211 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/quality.go` (Lines 21-21) — *AST node count: 1211*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 332. [LOW] Large function AST node volume (1235 nodes) (`RI-STR-VOLUME-E:_SWE_graph_visualizer_graph_mjs-draw-L44`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "draw" contains 1235 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 44-44) — *AST node count: 1235*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 333. [LOW] Large function AST node volume (1245 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_orchestrator_go-Run-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "Run" contains 1245 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/orchestrator.go` (Lines 42-42) — *AST node count: 1245*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 334. [LOW] Large function AST node volume (1337 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_evolve_go-evolveHandler-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "evolveHandler" contains 1337 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/evolve.go` (Lines 42-42) — *AST node count: 1337*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 335. [LOW] Large function AST node volume (1507 nodes) (`RI-STR-VOLUME-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-build-L260`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "build" contains 1507 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 260-260) — *AST node count: 1507*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 336. [LOW] Large function AST node volume (1545 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_broker_workspace_go-newSandboxBrokerHandler-L276`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "newSandboxBrokerHandler" contains 1545 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/broker_workspace.go` (Lines 276-276) — *AST node count: 1545*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 337. [LOW] Large function AST node volume (1551 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_evaluator_go-EvaluateCandidate-L188`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "EvaluateCandidate" contains 1551 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/evaluator.go` (Lines 188-188) — *AST node count: 1551*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 338. [LOW] Large function AST node volume (2063 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_annealer_go-Evolve-L99`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "Evolve" contains 2063 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/annealer.go` (Lines 99-99) — *AST node count: 2063*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 339. [LOW] Large function AST node volume (2698 nodes) (`RI-STR-VOLUME-E:_SWE_admin_portal_src_components_AdminPortal_jsx-AdminPortal-L30`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "AdminPortal" contains 2698 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/AdminPortal.jsx` (Lines 30-30) — *AST node count: 2698*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 340. [LOW] Large function AST node volume (3200 nodes) (`RI-STR-VOLUME-E:_SWE_api_router_go-newRouter-L224`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "newRouter" contains 3200 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/router.go` (Lines 224-224) — *AST node count: 3200*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 341. [LOW] Large function AST node volume (401 nodes) (`RI-STR-VOLUME-E:_SWE_thermodynamic_ast_engine_src_main_rs-visit-L490`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "visit" contains 401 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 490-490) — *AST node count: 401*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 342. [LOW] Large function AST node volume (402 nodes) (`RI-STR-VOLUME-E:_SWE_api_archive_go-BeginUpload-L125`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "BeginUpload" contains 402 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 125-125) — *AST node count: 402*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 343. [LOW] Large function AST node volume (405 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_evaluator_test_go-TestCopyProjectFiles_SanitizesSecretsAndBins-L402`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestCopyProjectFiles_SanitizesSecretsAndBins" contains 405 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/evaluator_test.go` (Lines 402-402) — *AST node count: 405*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 344. [LOW] Large function AST node volume (407 nodes) (`RI-STR-VOLUME-E:_SWE_api_aws_store_test_go-TestS3ObjectStorePresignsBoundedPendingUploadAndDownload-L689`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestS3ObjectStorePresignsBoundedPendingUploadAndDownload" contains 407 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 689-689) — *AST node count: 407*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 345. [LOW] Large function AST node volume (409 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_sandbox_lifecycle_go-AcquireLease-L516`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "AcquireLease" contains 409 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 516-516) — *AST node count: 409*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 346. [LOW] Large function AST node volume (409 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_workspace_go-readWorkspaceTarget-L145`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "readWorkspaceTarget" contains 409 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/workspace.go` (Lines 145-145) — *AST node count: 409*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 347. [LOW] Large function AST node volume (410 nodes) (`RI-STR-VOLUME-E:_SWE_api_archive_go-CompleteUpload-L175`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "CompleteUpload" contains 410 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 175-175) — *AST node count: 410*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 348. [LOW] Large function AST node volume (411 nodes) (`RI-STR-VOLUME-E:_SWE_api_lambda_test_go-TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey-L339`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestLoadSSMSecureStringRequiresStandardTierAndAWSManagedKey" contains 411 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/lambda_test.go` (Lines 339-339) — *AST node count: 411*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 349. [LOW] Large function AST node volume (412 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_main_go-extractZip-L165`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "extractZip" contains 412 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 165-165) — *AST node count: 412*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 350. [LOW] Large function AST node volume (412 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_workspace_test_go-TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes-L52`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestWorkspaceTargetUsesRealFunctionBoundaryAndRejectsEscapes" contains 412 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/workspace_test.go` (Lines 52-52) — *AST node count: 412*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 351. [LOW] Large function AST node volume (417 nodes) (`RI-STR-VOLUME-E:_SWE_api_refresh_test_go-TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials-L82`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestRefreshHandlerAllowsConcurrentRenewalsWithoutLeakingCredentials" contains 417 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/refresh_test.go` (Lines 82-82) — *AST node count: 417*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 352. [LOW] Large function AST node volume (420 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_broker_test_go-TestSandboxRunnerIsolationAndCIPinsImage-L30`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestSandboxRunnerIsolationAndCIPinsImage" contains 420 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_test.go` (Lines 30-30) — *AST node count: 420*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 353. [LOW] Large function AST node volume (422 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_sandbox_lifecycle_go-ReleaseLease-L623`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ReleaseLease" contains 422 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 623-623) — *AST node count: 422*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 354. [LOW] Large function AST node volume (422 nodes) (`RI-STR-VOLUME-E:_SWE_thermodynamic_ast_engine_src_languages_rs-name-L43`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "name" contains 422 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/languages.rs` (Lines 43-43) — *AST node count: 422*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 355. [LOW] Large function AST node volume (426 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_energy_test_go-TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation-L278`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestAnnealer_UphillMove_MetropolisAcceptanceAndBestIsolation" contains 426 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy_test.go` (Lines 278-278) — *AST node count: 426*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 356. [LOW] Large function AST node volume (427 nodes) (`RI-STR-VOLUME-E:_SWE_api_main_test_go-TestLegacyUploadRejectsInvalidFilesBeforeStorage-L443`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestLegacyUploadRejectsInvalidFilesBeforeStorage" contains 427 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/main_test.go` (Lines 443-443) — *AST node count: 427*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 357. [LOW] Large function AST node volume (429 nodes) (`RI-STR-VOLUME-E:_SWE_api_account_store_go-access-L79`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "access" contains 429 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/account_store.go` (Lines 79-79) — *AST node count: 429*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 358. [LOW] Large function AST node volume (434 nodes) (`RI-STR-VOLUME-E:_SWE_api_aws_store_test_go-TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst-L644`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestDynamoDBArchiveRepositoryReturnsNotFoundAndPagesNewestFirst" contains 434 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 644-644) — *AST node count: 434*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 359. [LOW] Large function AST node volume (440 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_agent_test_go-makeTestReport-L15`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "makeTestReport" contains 440 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/agent_test.go` (Lines 15-15) — *AST node count: 440*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 360. [LOW] Large function AST node volume (441 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_test_go-testAccountService-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "testAccountService" contains 441 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 23-23) — *AST node count: 441*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 361. [LOW] Large function AST node volume (446 nodes) (`RI-STR-VOLUME-E:_SWE_api_aws_store_test_go-TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects-L800`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestS3ObjectStoreRejectsWrongMissingOversizedAndMalformedStagedObjects" contains 446 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 800-800) — *AST node count: 446*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 362. [LOW] Large function AST node volume (448 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_strategy_go-hoistLoopAllocations-L843`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "hoistLoopAllocations" contains 448 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 843-843) — *AST node count: 448*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 363. [LOW] Large function AST node volume (454 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_broker_test_go-TestRemoteBrokerEvaluator_SuccessfulEvaluation-L322`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestRemoteBrokerEvaluator_SuccessfulEvaluation" contains 454 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/broker_test.go` (Lines 322-322) — *AST node count: 454*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 364. [LOW] Large function AST node volume (461 nodes) (`RI-STR-VOLUME-E:_SWE_source_stager_main_go-validateArchive-L382`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "validateArchive" contains 461 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/source-stager/main.go` (Lines 382-382) — *AST node count: 461*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 365. [LOW] Large function AST node volume (470 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_main_test_go-TestBroker_EvaluationResponseCarriesUpdatedBytesAndError-L850`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestBroker_EvaluationResponseCarriesUpdatedBytesAndError" contains 470 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/main_test.go` (Lines 850-850) — *AST node count: 470*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 366. [LOW] Large function AST node volume (471 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_energy_test_go-TestAnnealer_DownhillImprovement-L197`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestAnnealer_DownhillImprovement" contains 471 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy_test.go` (Lines 197-197) — *AST node count: 471*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 367. [LOW] Large function AST node volume (474 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_snapshot_store_go-ExtractSnapshot-L77`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ExtractSnapshot" contains 474 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_store.go` (Lines 77-77) — *AST node count: 474*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 368. [LOW] Large function AST node volume (485 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_test_go-exerciseAccountStore-L305`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "exerciseAccountStore" contains 485 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 305-305) — *AST node count: 485*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 369. [LOW] Large function AST node volume (489 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_manifests_go-parseDockerfile-L313`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "parseDockerfile" contains 489 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 313-313) — *AST node count: 489*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 370. [LOW] Large function AST node volume (489 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_strategy_go-checkBufferEscapes-L910`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "checkBufferEscapes" contains 489 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 910-910) — *AST node count: 489*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 371. [LOW] Large function AST node volume (492 nodes) (`RI-STR-VOLUME-E:_SWE_api_local_store_go-StageUpload-L85`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "StageUpload" contains 492 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 85-85) — *AST node count: 492*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 372. [LOW] Large function AST node volume (497 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_scan_handler_test_go-TestScanRepoHandler_CleansTemporaryWorkspace-L136`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestScanRepoHandler_CleansTemporaryWorkspace" contains 497 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/scan_handler_test.go` (Lines 136-136) — *AST node count: 497*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 373. [LOW] Large function AST node volume (498 nodes) (`RI-STR-VOLUME-E:_SWE_api_main_go-serveDownload-L337`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "serveDownload" contains 498 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/main.go` (Lines 337-337) — *AST node count: 498*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 374. [LOW] Large function AST node volume (499 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_go-ServeHTTP-L96`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ServeHTTP" contains 499 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 96-96) — *AST node count: 499*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 375. [LOW] Large function AST node volume (499 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_workspace_test_go-TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles-L13`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestCreateWorkspaceArchive_ExcludesGitAndPreservesFiles" contains 499 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace_test.go` (Lines 13-13) — *AST node count: 499*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 376. [LOW] Large function AST node volume (500 nodes) (`RI-STR-VOLUME-E:_SWE_api_aws_store_test_go-TestS3ObjectStoreSurfacesEveryAWSOperationFailure-L835`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestS3ObjectStoreSurfacesEveryAWSOperationFailure" contains 500 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 835-835) — *AST node count: 500*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 377. [LOW] Large function AST node volume (504 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_deliverables_go-uploadDeliverableHandler-L249`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "uploadDeliverableHandler" contains 504 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/deliverables.go` (Lines 249-249) — *AST node count: 504*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 378. [LOW] Large function AST node volume (509 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_go-exchangeGitHubCode-L243`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "exchangeGitHubCode" contains 509 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 243-243) — *AST node count: 509*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 379. [LOW] Large function AST node volume (509 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_strategy_go-simplifyBooleanReturns-L511`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "simplifyBooleanReturns" contains 509 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 511-511) — *AST node count: 509*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 380. [LOW] Large function AST node volume (514 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_architecture_go-detectCycles-L103`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "detectCycles" contains 514 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/architecture.go` (Lines 103-103) — *AST node count: 514*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 381. [LOW] Large function AST node volume (514 nodes) (`RI-STR-VOLUME-E:_SWE_thermodynamic_ast_engine_src_main_rs-scan_file-L700`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "scan_file" contains 514 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/main.rs` (Lines 700-700) — *AST node count: 514*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 382. [LOW] Large function AST node volume (515 nodes) (`RI-STR-VOLUME-E:_SWE_graph_visualizer_OrbitControls_js-update-L166`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "update" contains 515 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 166-166) — *AST node count: 515*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 383. [LOW] Large function AST node volume (518 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_repo_intel_test_go-setupMockRepository-L16`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "setupMockRepository" contains 518 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/repo_intel_test.go` (Lines 16-16) — *AST node count: 518*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 384. [LOW] Large function AST node volume (536 nodes) (`RI-STR-VOLUME-E:_SWE_api_main_test_go-TestDownloadHandlerStreamsAuthenticatedPrivateS3Object-L593`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestDownloadHandlerStreamsAuthenticatedPrivateS3Object" contains 536 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/main_test.go` (Lines 593-593) — *AST node count: 536*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 385. [LOW] Large function AST node volume (540 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_test_go-TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions-L202`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestGitHubAccountsBindStatePKCEAndDenyExpiredOrAdminSessions" contains 540 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 202-202) — *AST node count: 540*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 386. [LOW] Large function AST node volume (555 nodes) (`RI-STR-VOLUME-E:_SWE_api_aws_store_test_go-TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures-L533`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestArchiveServiceSurfacesRepositoryAndObjectStoreFailures" contains 555 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/aws_store_test.go` (Lines 533-533) — *AST node count: 555*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 387. [LOW] Large function AST node volume (571 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_strategy_go-ProposeTargeted-L317`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ProposeTargeted" contains 571 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 317-317) — *AST node count: 571*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 388. [LOW] Large function AST node volume (573 nodes) (`RI-STR-VOLUME-E:_SWE_api_analysis_proxy_test_go-TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide-L16`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestAnalysisProxyUsesPortalAuthAndKeepsServiceTokenServerSide" contains 573 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy_test.go` (Lines 16-16) — *AST node count: 573*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 389. [LOW] Large function AST node volume (574 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_go-accountRequest-L294`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "accountRequest" contains 574 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts.go` (Lines 294-294) — *AST node count: 574*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 390. [LOW] Large function AST node volume (582 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_agent_go-WriteAuditMarkdown-L649`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "WriteAuditMarkdown" contains 582 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/agent.go` (Lines 649-649) — *AST node count: 582*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 391. [LOW] Large function AST node volume (583 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_sandbox_lifecycle_go-ReconcileStaleLeases-L674`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ReconcileStaleLeases" contains 583 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/sandbox_lifecycle.go` (Lines 674-674) — *AST node count: 583*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 392. [LOW] Large function AST node volume (585 nodes) (`RI-STR-VOLUME-E:_SWE_api_accounts_test_go-TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart-L149`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestGitHubAccountsKeepRepositoriesPrivateAndPersistAcrossRestart" contains 585 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/accounts_test.go` (Lines 149-149) — *AST node count: 585*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 393. [LOW] Large function AST node volume (589 nodes) (`RI-STR-VOLUME-E:_SWE_api_lambda_test_go-TestAPIGatewayV2HandlerMapsRequestAndResponse-L241`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestAPIGatewayV2HandlerMapsRequestAndResponse" contains 589 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/lambda_test.go` (Lines 241-241) — *AST node count: 589*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 394. [LOW] Large function AST node volume (594 nodes) (`RI-STR-VOLUME-E:_SWE_graph_visualizer_OrbitControls_js-constructor-L21`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "constructor" contains 594 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/OrbitControls.js` (Lines 21-21) — *AST node count: 594*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 395. [LOW] Large function AST node volume (595 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_workspace_go-ExtractWorkspaceArchive-L154`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "ExtractWorkspaceArchive" contains 595 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 154-154) — *AST node count: 595*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 396. [LOW] Large function AST node volume (604 nodes) (`RI-STR-VOLUME-E:_SWE_graph_visualizer_graph_mjs-details-L131`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "details" contains 604 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/graph.mjs` (Lines 131-131) — *AST node count: 604*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 397. [LOW] Large function AST node volume (628 nodes) (`RI-STR-VOLUME-E:_SWE_api_analysis_proxy_go-registerAnalysisRoutes-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "registerAnalysisRoutes" contains 628 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/analysis_proxy.go` (Lines 42-42) — *AST node count: 628*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 398. [LOW] Large function AST node volume (629 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_annealer_benchmark_test_go-TestBenchmark_OptimizationEvidence-L56`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestBenchmark_OptimizationEvidence" contains 629 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/annealer_benchmark_test.go` (Lines 56-56) — *AST node count: 629*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 399. [LOW] Large function AST node volume (632 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_agent_go-predictFailures-L285`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "predictFailures" contains 632 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/agent.go` (Lines 285-285) — *AST node count: 632*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 400. [LOW] Large function AST node volume (637 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_annealer_benchmark_test_go-TestBenchmark_Milestone2_ComparativeStrategyHarness-L171`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestBenchmark_Milestone2_ComparativeStrategyHarness" contains 637 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/annealer_benchmark_test.go` (Lines 171-171) — *AST node count: 637*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 401. [LOW] Large function AST node volume (646 nodes) (`RI-STR-VOLUME-E:_SWE_infra_analysis_stream_proxy_index_mjs-<anonymous@219>-L219`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "<anonymous@219>" contains 646 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/infra/analysis/stream-proxy/index.mjs` (Lines 219-219) — *AST node count: 646*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 402. [LOW] Large function AST node volume (650 nodes) (`RI-STR-VOLUME-E:_SWE_api_local_store_go-registerLocalFileRoutes-L505`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "registerLocalFileRoutes" contains 650 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/local_store.go` (Lines 505-505) — *AST node count: 650*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 403. [LOW] Large function AST node volume (673 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_energy_go-EvaluateEnergy-L155`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "EvaluateEnergy" contains 673 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/energy.go` (Lines 155-155) — *AST node count: 673*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 404. [LOW] Large function AST node volume (682 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_strategy_go-extractCommonSubexpressions-L705`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "extractCommonSubexpressions" contains 682 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/strategy.go` (Lines 705-705) — *AST node count: 682*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 405. [LOW] Large function AST node volume (695 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_workspace_go-CreateWorkspaceArchive-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "CreateWorkspaceArchive" contains 695 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/workspace.go` (Lines 41-41) — *AST node count: 695*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 406. [LOW] Large function AST node volume (700 nodes) (`RI-STR-VOLUME-E:_SWE_graph_visualizer_model_mjs-buildGraphModel-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "buildGraphModel" contains 700 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/graph-visualizer/model.mjs` (Lines 23-23) — *AST node count: 700*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 407. [LOW] Large function AST node volume (717 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_main_go-setupEngineWithSnapshots-L624`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "setupEngineWithSnapshots" contains 717 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 624-624) — *AST node count: 717*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 408. [LOW] Large function AST node volume (739 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_agent_go-generateDirective-L436`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "generateDirective" contains 739 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/agent.go` (Lines 436-436) — *AST node count: 739*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 409. [LOW] Large function AST node volume (763 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_snapshot_scanner_go-scanRepoHandlerWithRunner-L32`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "scanRepoHandlerWithRunner" contains 763 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_scanner.go` (Lines 32-32) — *AST node count: 763*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 410. [LOW] Large function AST node volume (784 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_main_go-auditHandler-L503`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "auditHandler" contains 784 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 503-503) — *AST node count: 784*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 411. [LOW] Large function AST node volume (810 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_scanner_go-scanRepoHandler-L25`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "scanRepoHandler" contains 810 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/scanner.go` (Lines 25-25) — *AST node count: 810*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 412. [LOW] Large function AST node volume (828 nodes) (`RI-STR-VOLUME-E:_SWE_api_local_store_test_go-TestLocalArchiveUploadFinalizeListDownloadDelete-L18`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestLocalArchiveUploadFinalizeListDownloadDelete" contains 828 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/local_store_test.go` (Lines 18-18) — *AST node count: 828*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 413. [LOW] Large function AST node volume (839 nodes) (`RI-STR-VOLUME-E:_SWE_api_archive_go-isPowerPointCompoundFile-L528`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "isPowerPointCompoundFile" contains 839 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/api/archive.go` (Lines 528-528) — *AST node count: 839*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 414. [LOW] Large function AST node volume (858 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_manifests_go-AnalyzeManifests-L23`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "AnalyzeManifests" contains 858 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/manifests.go` (Lines 23-23) — *AST node count: 858*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 415. [LOW] Large function AST node volume (880 nodes) (`RI-STR-VOLUME-E:_SWE_thermodynamic_ast_engine_src_relationships_rs-walk-L41`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "walk" contains 880 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/thermodynamic-ast-engine/src/relationships.rs` (Lines 41-41) — *AST node count: 880*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 416. [LOW] Large function AST node volume (882 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_snapshot_portal_test_go-TestSnapshotPortalScanEvolutionAndPatch-L19`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestSnapshotPortalScanEvolutionAndPatch" contains 882 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/snapshot_portal_test.go` (Lines 19-19) — *AST node count: 882*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 417. [LOW] Large function AST node volume (916 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_reliability_security_go-AnalyzeReliabilityAndSecurity-L42`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "AnalyzeReliabilityAndSecurity" contains 916 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/reliability_security.go` (Lines 42-42) — *AST node count: 916*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 418. [LOW] Large function AST node volume (924 nodes) (`RI-STR-VOLUME-E:_SWE_admin_portal_src_components_EvolutionLive_jsx-EvolutionLive-L9`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "EvolutionLive" contains 924 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/admin-portal/src/components/EvolutionLive.jsx` (Lines 9-9) — *AST node count: 924*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 419. [LOW] Large function AST node volume (926 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_structural_go-convertReportToFindings-L67`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "convertReportToFindings" contains 926 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/structural.go` (Lines 67-67) — *AST node count: 926*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 420. [LOW] Large function AST node volume (936 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_agent_repo_intel_discovery_go-DiscoverRepository-L169`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "DiscoverRepository" contains 936 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/agent/repo_intel/discovery.go` (Lines 169-169) — *AST node count: 936*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 421. [LOW] Large function AST node volume (955 nodes) (`RI-STR-VOLUME-E:_SWE_go_backend_main_go-analyzeHandler-L341`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "analyzeHandler" contains 955 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/go-backend/main.go` (Lines 341-341) — *AST node count: 955*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 422. [LOW] Large function AST node volume (977 nodes) (`RI-STR-VOLUME-E:_SWE_samples_benchmark_benchmark_target_test_go-TestEvaluateRiskEngine-L8`)

- **Category:** `STRUCTURAL`
- **Verification Status:** `PROBABLE` (Confidence: `0.90`)
- **Detection Method:** `TREE_SITTER_AST`
- **Root Cause:** Function fulfills multiple responsibilities violating Single Responsibility Principle.

Function "TestEvaluateRiskEngine" contains 977 AST nodes, indicating monolithic implementation.

**Evidence Locations:**

- `E:/SWE/samples/benchmark/benchmark_target_test.go` (Lines 8-8) — *AST node count: 977*

> [!TIP]
> **Recommended Action:** Split function into smaller single-purpose procedures.
>
> **Expected Benefit:** Reduces mental overhead and facilitates focused unit testing.

---

### 423. [LOW] Directory .github/workflows lacks unit tests for 6 source files (`RI-QUAL-NOTEST-_github_workflows`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 6 production source files (check-frontend.yml, ci-go.yml, cleanup-analysis.yml, deploy-analysis.yml, deploy-frontend.yml, publish-api-image.yml) but no co-located unit or integration test files.

**Evidence Locations:**

- `.github/workflows/check-frontend.yml` (Lines 1-1) — *Source file without co-located unit tests*
- `.github/workflows/ci-go.yml` (Lines 1-1) — *Source file without co-located unit tests*
- `.github/workflows/cleanup-analysis.yml` (Lines 1-1) — *Source file without co-located unit tests*
- `.github/workflows/deploy-analysis.yml` (Lines 1-1) — *Source file without co-located unit tests*
- `.github/workflows/deploy-frontend.yml` (Lines 1-1) — *Source file without co-located unit tests*
- `.github/workflows/publish-api-image.yml` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in .github/workflows to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 424. [LOW] Directory admin-portal lacks unit tests for 5 source files (`RI-QUAL-NOTEST-admin_portal`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 5 production source files (.oxlintrc.json, README.md, package-lock.json, package.json, vite.config.js) but no co-located unit or integration test files.

**Evidence Locations:**

- `admin-portal/.oxlintrc.json` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/README.md` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/package-lock.json` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/package.json` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/vite.config.js` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in admin-portal to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 425. [LOW] Directory admin-portal/src lacks unit tests for 6 source files (`RI-QUAL-NOTEST-admin_portal_src`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 6 production source files (App.jsx, archive-api.js, archive-download.js, evolution-api.js, evolution-measurements.js, main.jsx) but no co-located unit or integration test files.

**Evidence Locations:**

- `admin-portal/src/App.jsx` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/archive-api.js` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/archive-download.js` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/evolution-api.js` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/evolution-measurements.js` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/main.jsx` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in admin-portal/src to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 426. [LOW] Directory admin-portal/src/components lacks unit tests for 5 source files (`RI-QUAL-NOTEST-admin_portal_src_components`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 5 production source files (AdminPortal.jsx, EvolutionCharts.jsx, EvolutionLive.jsx, EvolutionWorkspace.jsx, GraphVisualizer.jsx) but no co-located unit or integration test files.

**Evidence Locations:**

- `admin-portal/src/components/AdminPortal.jsx` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/components/EvolutionCharts.jsx` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/components/EvolutionLive.jsx` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/components/EvolutionWorkspace.jsx` (Lines 1-1) — *Source file without co-located unit tests*
- `admin-portal/src/components/GraphVisualizer.jsx` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in admin-portal/src/components to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 427. [LOW] Directory graph-visualizer lacks unit tests for 12 source files (`RI-QUAL-NOTEST-graph_visualizer`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 12 production source files (OrbitControls.js, account.mjs, config.mjs, d3-force-3d.min.js, d3-octree.min.js, d3.min.js, graph.mjs, measurements.mjs, model.mjs, routes.mjs, scan.mjs, three.min.js) but no co-located unit or integration test files.

**Evidence Locations:**

- `graph-visualizer/OrbitControls.js` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/account.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/config.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/d3-force-3d.min.js` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/d3-octree.min.js` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/d3.min.js` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/graph.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/measurements.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/model.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/routes.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/scan.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `graph-visualizer/three.min.js` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in graph-visualizer to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 428. [LOW] Directory infra/analysis lacks unit tests for 4 source files (`RI-QUAL-NOTEST-infra_analysis`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 4 production source files (README.md, parameters.example.json, samconfig.toml, template.yaml) but no co-located unit or integration test files.

**Evidence Locations:**

- `infra/analysis/README.md` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/parameters.example.json` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/samconfig.toml` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/template.yaml` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in infra/analysis to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 429. [LOW] Directory infra/analysis/stream-proxy lacks unit tests for 4 source files (`RI-QUAL-NOTEST-infra_analysis_stream_proxy`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 4 production source files (index.mjs, index.test.mjs, package-lock.json, package.json) but no co-located unit or integration test files.

**Evidence Locations:**

- `infra/analysis/stream-proxy/index.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/stream-proxy/index.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/stream-proxy/package-lock.json` (Lines 1-1) — *Source file without co-located unit tests*
- `infra/analysis/stream-proxy/package.json` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in infra/analysis/stream-proxy to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 430. [LOW] Directory tests lacks unit tests for 5 source files (`RI-QUAL-NOTEST-tests`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 5 production source files (analysis-cleanup.test.mjs, analysis-template.test.mjs, analysis-workflow.test.mjs, graph-visualizer-config.test.mjs, sandbox-host-validation.test.mjs) but no co-located unit or integration test files.

**Evidence Locations:**

- `tests/analysis-cleanup.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `tests/analysis-template.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `tests/analysis-workflow.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `tests/graph-visualizer-config.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*
- `tests/sandbox-host-validation.test.mjs` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in tests to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 431. [LOW] Directory thermodynamic-ast-engine lacks unit tests for 3 source files (`RI-QUAL-NOTEST-thermodynamic_ast_engine`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 3 production source files (Cargo.toml, README.md, thermodynamic_report.json) but no co-located unit or integration test files.

**Evidence Locations:**

- `thermodynamic-ast-engine/Cargo.toml` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/README.md` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/thermodynamic_report.json` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in thermodynamic-ast-engine to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 432. [LOW] Directory thermodynamic-ast-engine/src lacks unit tests for 3 source files (`RI-QUAL-NOTEST-thermodynamic_ast_engine_src`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 3 production source files (languages.rs, main.rs, relationships.rs) but no co-located unit or integration test files.

**Evidence Locations:**

- `thermodynamic-ast-engine/src/languages.rs` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/src/main.rs` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/src/relationships.rs` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in thermodynamic-ast-engine/src to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

### 433. [LOW] Directory thermodynamic-ast-engine/test_samples lacks unit tests for 3 source files (`RI-QUAL-NOTEST-thermodynamic_ast_engine_test_samples`)

- **Category:** `TEST_QUALITY`
- **Verification Status:** `CONFIRMED` (Confidence: `0.85`)
- **Detection Method:** `FILE_CORRELATION_HEURISTIC`
- **Root Cause:** Modules authored without corresponding unit test suites.

Directory contains 3 production source files (crawler.go, data_pipeline.py, utils.py) but no co-located unit or integration test files.

**Evidence Locations:**

- `thermodynamic-ast-engine/test_samples/crawler.go` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/test_samples/data_pipeline.py` (Lines 1-1) — *Source file without co-located unit tests*
- `thermodynamic-ast-engine/test_samples/utils.py` (Lines 1-1) — *Source file without co-located unit tests*

> [!TIP]
> **Recommended Action:** Add unit tests for modules in thermodynamic-ast-engine/test_samples to prevent regressions during automated evolution.
>
> **Expected Benefit:** Provides safety gates and regression catchers during autonomous refactoring.

---

## Performance Telemetry

- **Discovery Duration:** 44 ms
- **Manifest Analysis Duration:** 15 ms
- **Structural AST Duration:** 632 ms
- **Architecture Graph Duration:** 5 ms
- **Security Duration:** 2204 ms
- **Quality Duration:** 0 ms
- **Total Pipeline Runtime:** 2903 ms
