# Milestone 3.4 — Phase 1: Diagnostic Subsystem Architecture Audit & Gap Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Active Branch:** `research/milestone-3.4-semantic-analysis`  
**Base Release Commit:** `52ff0f1822ff8c25bf64d2e860124b9dc9b7351b`  
**Target:** Subsystem Architecture Audit, Blind Spot Catalog, and Semantic Requirements  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Milestone 3.3 evaluated the self-healing platform on five unseen external Go repositories (`http_router`, `sql_service`, `cache_engine`, `cli_tooling`, `queue_worker`). The evaluation exposed a severe diagnostic generalization gap:
- **Synthetic Suite v3 Recall:** 100.00% (96/96 defects detected)
- **External Unseen Repositories Recall:** 14.29% (2/14 defects detected)
- **External Unseen Precision:** 100.00% (2/2 findings confirmed)
- **External Unseen F1 Score:** 25.00%

This architecture audit conducts an exhaustive inspection of the current diagnostic subsystem across `go-backend/agent/repo_intel/` and its subpackages. We trace finding lifecycles from raw discovery to remediation planning, dissect the precise failure modes across the codebase with file paths and line numbers, and establish the architectural specifications for the Semantic Intermediate Representation (IR).

---

## 2. Current Diagnostic Pipeline Architecture & Interaction Flow

The existing diagnostic subsystem operates as a multi-stage sequential filter:

```
[1. Discovery]
       |
       v
[2. Raw Heuristic Analyzers]
       |-- ReliabilitySecurityAnalyzer (regex-based, line-by-line)
       |-- QualityAnalyzer (file/test existence)
       |-- ManifestAnalyzer (go.mod, lockfiles)
       v
[3. ContextEngine.FilterAndEnrich]
       |-- SQLContextAnalyzer (AST BinaryExpr checks)
       |-- CommandContextAnalyzer (AST exec.Command checks)
       |-- ResourceContextAnalyzer (AST loop-body checks)
       |-- StructuralContextAnalyzer (structural clone hashing)
       |-- ArchitectureContextAnalyzer (package cycle & layer rules)
       |-- QualityContextAnalyzer (symbol-level test link & godoc)
       |-- Counter-Evidence Gatekeeper & Suppression Engine
       v
[4. DiagnosticEngine]
       |-- FindingValidator (evidence file/line existence & freshness)
       |-- FindingCorrelator (graph correlation & deduplication)
       |-- ConfidenceCalibrator (evidence & heuristic scoring)
       |-- SeverityAssessor (impact scope & risk classification)
       |-- ReadinessClassifier (autonomous vs approval gating)
       |-- DiagnosticReporter (JSON & Markdown diagnostic reports)
       v
[5. Remediation Planning & Execution Engine]
```

### 2.1 Interaction Coupling & Data Flow
1. **Raw Ingestion:** `repo_intel.ReliabilitySecurityAnalyzer` scans files line-by-line using regular expressions (`unhandledErrorRegex`, `sqlInjectionRegex`, `commandExecRegex`). Findings enter the pipeline with heuristic verification status (`StatusHeuristic` or `StatusProbable`).
2. **Contextual Enrichment & Suppression:** `ContextEngine` runs AST analyzers on Go files. It suppresses false positives by matching candidate findings against counter-evidence rules (e.g. `SAFE_SQL_PARAMETERIZATION`, `SAFE_STATIC_EXEC_ARGUMENTS`). It also generates newly discovered contextual findings (`ContextFinding`) with `StateConfirmed`.
3. **Validation & Integrity Verification:** `FindingValidator` verifies that every evidence location points to an existing file and valid line bounds on the physical filesystem.
4. **Readiness & Gatekeeping:** `ReadinessClassifier` assigns findings to `AUTONOMOUS`, `APPROVAL_REQUIRED`, or `MANUAL_INVESTIGATION`. Findings flagged with `INSUFFICIENT_EVIDENCE` or `HUMAN_INVESTIGATION_REQUIRED` are strictly blocked from automated remediation.

---

## 3. Dissection of the 5 Architectural Failure Modes

An in-depth code inspection revealed exactly how and why 12 of the 14 external ground truth defects were missed in Milestone 3.3.

### 3.1 Failure Mode 1: AST Pattern Rigidity (Missed: `DEF-SQL-01`)
- **Target Defect:** `sql_service/repository.go:50` (`FindUserByNameRaw`) constructs dynamic SQL using `fmt.Sprintf("SELECT id, username, email FROM users WHERE username = '%s'", name)`.
- **Code Audit:** In `go-backend/agent/repo_intel/diagnostic/context/sql_context.go`:
  - **Lines 54-58:**
    ```go
    binExpr, ok := n.(*ast.BinaryExpr)
    if !ok || binExpr.Op != token.ADD {
        return true
    }
    ```
    The analyzer exclusively checks `*ast.BinaryExpr` nodes with `binExpr.Op == token.ADD`. It completely ignores `*ast.CallExpr` nodes invoking formatting functions like `fmt.Sprintf`, `fmt.Sprint`, `strings.Join`, or query builders.
  - **Line 34:**
    ```go
    rawSQLSelect = regexp.MustCompile(`(?i)(SELECT\s+.+\s+FROM\s+[a-zA-Z0-9_]+)`)
    ```
    The query regex only recognizes `SELECT ... FROM ...`. Dynamic queries involving `INSERT INTO ...`, `UPDATE ... SET ...`, `DELETE FROM ...`, or schema migrations are invisible to the analyzer.
  - **In `reliability_security.go:37`:**
    ```go
    sqlInjectionRegex = regexp.MustCompile(`(?i)(?:SELECT|INSERT|UPDATE|DELETE)\s+.*\+\s*[a-zA-Z0-9_]+`)
    ```
    The fallback regex also requires a literal `+` operator on the same line, missing formatted strings.

### 3.2 Failure Mode 2: Function-Level Resource Lifecycle Limitations (Missed: `DEF-HTTP-01`, `DEF-SQL-02`, `DEF-QUEUE-01`)
- **Target Defects:**
  - `http_router/middleware.go:37` (`UpstreamProbe`): `http.Get(url)` response body not closed via `defer resp.Body.Close()`.
  - `sql_service/repository.go:78` (`ListRecentUsers`): `sql.Rows` opened via `r.db.QueryContext` without `defer rows.Close()`.
  - `queue_worker/internal/producer/producer.go:47` (`SendDirect`): `net.Dial` connection opened without `defer conn.Close()`.
- **Code Audit:** In `go-backend/agent/repo_intel/diagnostic/context/resource_context.go`:
  - **Lines 46-47 & 118-121:**
    ```go
    checkLoopBody := func(body *ast.BlockStmt, loopPos token.Pos) { ... }
    ast.Inspect(node, func(n ast.Node) bool {
        switch stmt := n.(type) {
        case *ast.ForStmt:
            checkLoopBody(stmt.Body, stmt.Pos())
        case *ast.RangeStmt:
            checkLoopBody(stmt.Body, stmt.Pos())
        }
        return true
    })
    ```
    Resource leak detection is **only executed on loop bodies** (`*ast.ForStmt` and `*ast.RangeStmt`). Normal function declarations (`*ast.FuncDecl`) are never analyzed for unclosed resources!
  - **Lines 60-70:**
    ```go
    sel, isSel := call.Fun.(*ast.SelectorExpr)
    pkgIdent, isPkgIdent := sel.X.(*ast.Ident)
    if methods, ok := resourceAllocators[pkgIdent.Name]; ok && methods[sel.Sel.Name] { ... }
    ```
    `sel.X` must be an `*ast.Ident` whose name matches a package identifier (`"os"`, `"net"`, `"http"`, `"sql"`). When a resource is allocated on an instance or client (e.g. `client.Get(...)`, `db.Query(...)`, `db.Begin(...)`), `sel.X` is a local variable, not a package identifier. The allocator map does not recognize instance methods, and without type resolution, all method-based allocations are silently ignored.

### 3.3 Failure Mode 3: Error-Handling Idiom Blindness (Missed: `DEF-HTTP-02`, `DEF-SQL-03`, `DEF-QUEUE-03`)
- **Target Defects:**
  - `http_router/context.go:JSON`: `json.NewEncoder(c.Writer).Encode(data)` executed as an unassigned expression statement.
  - `sql_service/db.go:WithTransaction`: `tx.Rollback()` return value discarded.
  - `queue_worker/internal/consumer/consumer.go:ParseRawPayload`: `json.Unmarshal(data, &v)` called without checking error.
- **Code Audit:** In `go-backend/agent/repo_intel/reliability_security.go`:
  - **Line 36:**
    ```go
    unhandledErrorRegex = regexp.MustCompile(`(?i)^\s*_[,\s]*=\s*[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+\(`)
    ```
    The error detection heuristic requires an explicit blank assignment (`_ = call()`).
  - In idiomatic Go, developers frequently drop errors simply by calling the function as a bare expression statement (`ast.ExprStmt`):
    ```go
    json.NewEncoder(w).Encode(obj) // No _ = assignment!
    ```
    Neither `ReliabilitySecurityAnalyzer` nor `ContextEngine` inspects `ast.ExprStmt` to determine whether the invoked function or method has an error return type.

### 3.4 Failure Mode 4: Incomplete Interprocedural Dataflow (Missed: `DEF-CLI-01`, `DEF-CLI-02`)
- **Target Defects:**
  - `cli_tooling/runner.go:27` (`ExecuteShellScript`): `func ExecuteShellScript(script string)` passes `script` directly into `exec.Command("sh", "-c", script)`.
  - `cli_tooling/fs.go:16` (`ReadFileFromRoot`): `func ReadFileFromRoot(root, userPath string)` combines paths via `filepath.Join(root, userPath)` without verifying that the resolved path stays within `root`.
- **Code Audit:** In `go-backend/agent/repo_intel/diagnostic/context/command_context.go`:
  - **Lines 48-73:**
    ```go
    ast.Inspect(node, func(n ast.Node) bool {
        assign, ok := n.(*ast.AssignStmt)
        ...
        dynamicVars[lhsId.Name] = snip
    })
    ```
    Taint tracking only records variables assigned inside the file via `ast.AssignStmt`. Function parameters (`script string`, `userPath string`) are never added to the taint map.
  - Furthermore, arguments passed to shell binaries (`"sh"`, `"-c"`, `script` or `"cmd.exe"`, `"/c"`, `script`) are treated as separate discrete arguments, failing to recognize that the second argument is interpreted as arbitrary shell script by the shell binary.

### 3.5 Failure Mode 5: Concurrency and Goroutine Lifecycle Blindness (Missed: `DEF-CACHE-01`, `DEF-CACHE-02`, `DEF-HTTP-03`)
- **Target Defects:**
  - `cache_engine/cache.go:88` (`UnsafeKeys`): Reads shared `c.items` map without acquiring `c.mu.RLock()`.
  - `cache_engine/eviction.go:30` (`StartBackgroundEvictor`): Spawns `go func() { for range ticker.C { ... } }()` without context cancellation, shutdown channel, or termination signal.
  - `http_router/context.go:63` (`ExecuteWithTimeout`): `ctx, _ := context.WithTimeout(...)` discards the `cancel` function.
- **Code Audit:**
  - The entire diagnostic subsystem contains **zero** analyzers for concurrency or goroutine lifecycles.
  - There is no model linking struct fields to synchronization primitives (`sync.Mutex`, `sync.RWMutex`).
  - Goroutine invocations (`*ast.GoStmt`) are never inspected for loop termination conditions or shutdown channel receivers.
  - Calls returning context cancellation functions (`context.WithCancel`, `context.WithTimeout`, `context.WithDeadline`) are not checked for cancel function execution.

---

## 4. Benchmark Overfitting & Synthetic Bias Audit

Beyond missing detectors, the audit discovered explicit overfitting to synthetic benchmark naming conventions:

### 4.1 Synthetic Prefix Hardcoding in `context_engine.go`
In `go-backend/agent/repo_intel/diagnostic/context/context_engine.go` lines 186-188:
```go
base := filepath.Base(filepath.ToSlash(aff))
if strings.HasPrefix(base, "clean_") || strings.HasPrefix(base, "tested_") || strings.HasPrefix(base, "documented_") {
    // Excluded from NOTEST rules
}
```
And in `quality_context.go` lines 76-78:
```go
baseName := filepath.Base(filepath.ToSlash(srcRel))
if strings.HasPrefix(baseName, "clean_") || strings.HasPrefix(baseName, "tested_") || strings.HasPrefix(baseName, "documented_") {
    continue
}
```

### 4.2 Impact on Generalization
- In Suite v2 and Suite v3, all negative controls were artificially prefixed with `clean_`, `tested_`, or `documented_`.
- The analyzers relied on these filename substrings rather than semantically verifying whether functions had unit tests or documentation.
- When applied to real-world code (e.g. `http_router/middleware.go`), this heuristic fails completely.

---

## 5. Architectural Requirements for Semantic IR

To achieve robust external generalization without sacrificing precision or execution safety, the diagnostic engine must be re-architected around a formal **Semantic Intermediate Representation (IR)**.

### Core Architecture Principles:
1. **Semantic over Syntactic:** Diagnose code based on symbol definitions, types, control flow, and dataflow, rather than lexical tokens or AST shape matching.
2. **Graceful Type Degradation:** Leverage `go/types` where full type checking succeeds; gracefully fall back to AST-based signature and identifier inference when third-party packages are missing or incomplete.
3. **Explicit Multi-Exit CFG:** Represent function control flow as a directed graph of basic blocks with explicit entry, exit, branch conditions, deferred execution stacks, and error return paths.
4. **Principled Dataflow Framework:** Support forward and backward dataflow analysis with formal gen/kill sets and fixed-point iteration for reaching definitions and def-use chains.
5. **Unified Lifecycle Modeling:** Explicitly track state transitions for resources (Acquired $\to$ Deferred/Released $\to$ Escaped) and errors (Generated $\to$ Checked/Returned/Discarded).
6. **Conservative Gating:** Preserve uncertainty states (`INSUFFICIENT_EVIDENCE`) and never promote an ambiguous finding into automated remediation.

---

## 6. Audit Conclusion

The diagnostic engine's external generalization gap is not a consequence of minor regex bugs; it is an architectural limitation of relying on local AST node matching without control flow or dataflow context. Implementing a dedicated Semantic IR is the necessary and sufficient foundation for external generalization recovery.

Phase 1 Task 1.1 is complete. Ready to proceed with Semantic IR Design (Task 1.2).
