# Milestone 3.3 — Phase 3: External Diagnostic Error Analysis & Failure Mode Taxonomy

**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Branch:** `validation/milestone-3.3-external-validation`  
**Baseline Commit:** `5d41b531f46d10bc972b2d1518a0fb4cf139c09a`  
**Evaluation Mode:** Frozen-Baseline External Trial  
**Date:** 2026-11-04  

---

## 1. Executive Summary

In Phase 3, the frozen Milestone 3.2 baseline was evaluated against five unseen external Go repositories comprising 1,601 LOC and 14 ground-truth defects.

The evaluation yielded the following empirical divergence:
- **Synthetic Benchmark (Suite v3, 144 fixtures):** 100% Precision, 100% Recall, 100% Specificity.
- **External Real-World Repositories (5 repositories, 14 defects):** 100% Precision, 14.29% Recall, 100% Specificity.

This finding demonstrates that the system does not generate spurious false alarms on unseen code (0% FPR, 100% Precision). However, defect recall drops precipitously when encountering real-world idioms that deviate from the AST patterns synthesized during benchmark construction.

---

## 2. Exhaustive Failure Mode Taxonomy

Every false negative has been traced directly to its underlying architectural limitation in the diagnostic codebase.

```
                           ROOT CAUSES OF DEFECT UNDER-DETECTION
                                            |
         +------------------+---------------+----------------+------------------+
         |                  |                                |                  |
   AST Limitation    Scope Restriction               Idiom Blindness     Taint Absence
  (BinaryExpr only)  (Loop bodies only)             (Ignored expressions) (Function args)
```

### Failure Mode 1: AST Pattern Rigidity (`AST_LIMITATION`)
- **Impacted Defects:** `DEF-SQL-01` (`sql_service/repository.go:50`)
- **Code Construct:**
  ```go
  query := fmt.Sprintf("SELECT id, username, email, role, created_at FROM users WHERE username = '%s'", name)
  ```
- **Engine Behavior:**
  In `go-backend/agent/repo_intel/diagnostic/context/sql_context.go`, line 55, the AST analyzer specifically targets `*ast.BinaryExpr` where `binExpr.Op == token.ADD` (e.g., `query := "SELECT ... WHERE id = " + id`).
  In `go-backend/agent/repo_intel/reliability_security.go`, line 37, the regex requires `(?:SELECT|...)\s+.*\+\s*[a-zA-Z0-9_]+`.
- **Root Cause:** Both the regex heuristic and the contextual AST engine assume dynamic SQL is formed via string concatenation (`+`). Neither component recognizes `fmt.Sprintf` or string formatting wrappers into SQL statements.

### Failure Mode 2: Structural Scope Restriction (`SCOPE_RESTRICTION`)
- **Impacted Defects:** `DEF-HTTP-01` (`http_router/middleware.go:37`), `DEF-SQL-02` (`sql_service/repository.go:66`), `DEF-QUEUE-01` (`queue_worker/internal/producer/producer.go:47`)
- **Code Constructs:**
  - `resp, err := http.Get(probeURL)` without `defer resp.Body.Close()`
  - `rows, err := r.db.QueryContext(...)` without `defer rows.Close()`
  - `conn, err := net.DialTimeout(...)` without `defer conn.Close()`
- **Engine Behavior:**
  In `go-backend/agent/repo_intel/diagnostic/context/resource_context.go`, line 46:
  ```go
  checkLoopBody := func(body *ast.BlockStmt, loopPos token.Pos)
  ```
- **Root Cause:** `ResourceContextAnalyzer` was explicitly built to detect resource allocation inside `for` and `range` loops (`ResourceAllocInLoop`, `DeferInLoop`). It does not traverse top-level function bodies to verify if resources opened outside loops are closed. Consequently, linear resource leaks are outside the analyzer's search horizon.

### Failure Mode 3: Idiomatic Blindness (`IDIOM_BLINDNESS`)
- **Impacted Defects:** `DEF-HTTP-02` (`http_router/context.go:57`), `DEF-SQL-03` (`sql_service/db.go:42,48`), `DEF-QUEUE-03` (`queue_worker/internal/consumer/consumer.go:56`)
- **Code Constructs:**
  - `json.NewEncoder(c.Writer).Encode(data)` (return value unchecked)
  - `tx.Rollback()` (return value unchecked)
  - `json.Unmarshal(data, &res)` (return value unchecked)
- **Engine Behavior:**
  In `go-backend/agent/repo_intel/reliability_security.go`, line 36:
  ```go
  unhandledErrorRegex = regexp.MustCompile(`(?i)^\s*_[,\s]*=\s*[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+\(`)
  ```
- **Root Cause:** The engine only flags unhandled errors when the programmer explicitly writes `_ = func()`. When an error-returning function is called as an unassigned expression statement (`ast.ExprStmt`), the regex matches nothing, and the AST engine does not inspect expression statement return types.

### Failure Mode 4: Lack of Inter-Procedural Taint Tracking (`TAINT_ABSENCE`)
- **Impacted Defects:** `DEF-CLI-01` (`cli_tooling/runner.go:27`)
- **Code Construct:**
  ```go
  func ExecuteShellScript(userInput string) (string, error) {
      ...
      cmd = exec.CommandContext(ctx, "cmd.exe", "/C", userInput)
  }
  ```
- **Engine Behavior:**
  In `go-backend/agent/repo_intel/diagnostic/context/command_context.go`, lines 48-75, `dynamicVars` is populated only from intra-file assignments (`*ast.AssignStmt`) involving `token.ADD` or `fmt.Sprintf`.
- **Root Cause:** `userInput` is an incoming function parameter. The engine possesses no call-graph or inter-procedural taint propagation. Because `userInput` was not assigned from a string addition statement inside the function body, the analyzer treated it as a benign identifier.

### Failure Mode 5: Concurrency & Lifecycle Blindness (`CONCURRENCY_BLINDNESS`)
- **Impacted Defects:** `DEF-CACHE-01` (`cache_engine/cache.go:88`), `DEF-CACHE-02` (`cache_engine/eviction.go:30`), `DEF-HTTP-03` (`http_router/context.go:63`)
- **Code Constructs:**
  - `UnsafeKeys` reading `c.items` without `c.mu.RLock()`
  - `StartBackgroundEvictor` infinite `ticker.C` loop in goroutine without context cancellation
  - `context.WithTimeout` cancel function discarded
- **Root Cause:** Detecting concurrent map races requires whole-struct mutex synchronization modeling. Detecting goroutine and timer leaks requires goroutine escape analysis and lifetime verification. The baseline engine has no semantic models for Go channels, goroutine termination, or struct-level mutex invariants.

### Failure Mode 6: Resource Allocator Table Incompleteness (`CATALOG_OMISSION`)
- **Impacted Defects:** `DEF-QUEUE-02` (`queue_worker/pkg/retry/retry.go:56`)
- **Code Construct:**
  `timer := time.NewTimer(20 * time.Millisecond)` in loop without `timer.Stop()`.
- **Engine Behavior:**
  In `go-backend/agent/repo_intel/diagnostic/context/resource_context.go`, lines 39-44, `resourceAllocators` maps:
  `os` (Open, Create), `net` (Dial), `http` (Get, Post), `sql` (Open).
- **Root Cause:** `time.NewTimer` and `time.NewTicker` are not registered in the resource allocators map.

---

## 3. False Positive Analysis

- **Total False Positives:** 0
- **False Positive Rate:** 0.00%
- **Negative Controls Evaluated:** 15 across 5 repositories.
- **Analysis:**
  All clean negative controls (parameterized SQL in `repository.go`, direct binary execution in `runner.go`, thread-safe map access in `cache.go`, exponential backoff in `retry.go`) were correctly evaluated as clean. The contextual engine's suppression rules successfully eliminated 14 heuristic alerts, preventing false alarm pollution.

---

## 4. Engineering Recommendations for Next Major Release

1. **AST Generalization:** Expand SQL injection and command injection rules to match `fmt.Sprintf`, `strings.Join`, and template evaluations rather than binary addition alone.
2. **Expression Statement Error Tracking:** Implement Go type-checker integration (`go/types`) to flag any `ast.ExprStmt` whose tuple return includes an `error` interface.
3. **Linear Resource Lifecycles:** Extend `ResourceContextAnalyzer` to inspect all function declarations (`*ast.FuncDecl`) for unclosed handles, using control flow graphs (CFG) to verify `defer Close()` reachability.
4. **Inter-procedural Taint Propagation:** Introduce summary-based inter-procedural dataflow analysis to track user inputs from entry points into sink APIs.

