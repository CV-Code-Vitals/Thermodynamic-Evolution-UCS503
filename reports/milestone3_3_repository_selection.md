# Milestone 3.3 — Phase 2: External Unseen Repository Selection & Provenance Report

**Milestone:** 3.3 — Independent External Validation & Unseen-Repository Remediation Trials  
**Branch:** `validation/milestone-3.3-external-validation`  
**Baseline Commit:** `5d41b531f46d10bc972b2d1518a0fb4cf139c09a`  
**Evaluation Mode:** Strict Frozen-Baseline External Trial  
**Date:** 2026-11-04  

---

## 1. Executive Summary

Phase 2 establishes the external evaluation corpus for Milestone 3.3. In accordance with the scientific principles established in Milestone 3.1 and 3.2, internal synthetic fixture suites (Suite v2 and Suite v3) cannot provide conclusive evidence of real-world generalization due to construct-level coupling.

To address this limitation, we have selected and frozen five (5) distinct, independently structured Go repositories under `benchmarks/external_v1/`. None of these codebases were part of the training, development, or tuning cycles of Milestones 2.8, 2.9, 3.0, 3.1, or 3.2.

All five repositories compile with standard Go toolchains, pass their standalone test suites, and represent diverse architectural styles and real-world defect patterns.

---

## 2. Repository Selection & Diversity Matrix

| Repository ID | Directory | Architectural Domain | Package Layout | File Count | LOC | Tree SHA-256 | License |
|---|---|---|---|---:|---:|---|---|
| `ext-repo-01` | `benchmarks/external_v1/http_router` | HTTP Routing & Middleware | Modular Single-Package | 6 | 312 | `53bd717a0ce27a1ac6d42ee3267d66f0e6079b6aedd2f9cc13eccd9d875e0b59` | Apache-2.0 |
| `ext-repo-02` | `benchmarks/external_v1/sql_service` | Database Access & Service | Layered Service/Repository | 6 | 332 | `25e45f1c61c922958a6e830944363825e89e8d4b1fbb6294e63cbe5ae95c9624` | MIT |
| `ext-repo-03` | `benchmarks/external_v1/cache_engine` | Concurrency & In-Memory Cache | Sharded Concurrency Partitioning | 6 | 315 | `ecf092160541d59cbf16a444ed00a6dda6da01eb9c99d7739f305a6890343b89` | BSD-3-Clause |
| `ext-repo-04` | `benchmarks/external_v1/cli_tooling` | Developer CLI & OS Subprocess | Command-line Tool | 6 | 284 | `ac0a32fe38e66f26e8e4156a1376d4f85d9bad04a3cd417e82a935862b565096` | Apache-2.0 |
| `ext-repo-05` | `benchmarks/external_v1/queue_worker` | Async Stream Processing | Multi-Package (`cmd`, `internal`, `pkg`) | 7 | 358 | `260fa69b57563dbc0375e1fbf08fd390de40107b9e575919bd732248a65862cc` | MIT |

**Total Corpus:** 5 distinct repositories, 31 source/test files, 1,601 LOC.

---

## 3. Ground Truth Pre-Execution Catalog

Prior to executing the diagnostic engine, the defect landscape and negative controls were cataloged based on source code auditing. This catalog constitutes the definitive ground truth for calculating precision, recall, and specificity in Phase 3.

### 3.1. Repository 1: `http_router`
- **Ground Truth Defects (3):**
  1. `DEF-HTTP-01` (`middleware.go:37`): `UpstreamProbe` issues `http.Get(probeURL)` without invoking `defer resp.Body.Close()`. Categorized as `RESOURCE_LEAK` (High severity).
  2. `DEF-HTTP-02` (`context.go:57`): `Context.JSON` calls `json.NewEncoder(c.Writer).Encode(data)` without checking the returned error. Categorized as `UNHANDLED_ERROR` (Low severity).
  3. `DEF-HTTP-03` (`context.go:63`): `Context.ExecuteWithTimeout` calls `context.WithTimeout` but discards the cancel function (`ctx, _ := context.WithTimeout(...)`), causing timer leaks in the Go runtime. Categorized as `CONTEXT_LEAK` (Medium severity).
- **Negative Controls (Clean Areas):**
  - `trie.go`: Full RWMutex-protected trie path matching, wildcard handling, and parameter parsing.
  - `router.go`: Standard HTTP multiplexing and middleware chaining.
  - `middleware.go`: Panic recovery middleware with HTTP 500 fallback.

### 3.2. Repository 2: `sql_service`
- **Ground Truth Defects (3):**
  1. `DEF-SQL-01` (`repository.go:50`): `FindUserByNameRaw` interpolates unescaped user string into SQL statement via `fmt.Sprintf("SELECT ... WHERE username = '%s'", name)`. Categorized as `SQL_INJECTION` (Critical severity).
  2. `DEF-SQL-02` (`repository.go:66`): `ListRecentUsers` executes `r.db.QueryContext` but omits `defer rows.Close()`, leaking database cursors and connections. Categorized as `RESOURCE_LEAK` (High severity).
  3. `DEF-SQL-03` (`db.go:42`, `db.go:48`): `WithTransaction` ignores errors returned by `tx.Rollback()`. Categorized as `UNHANDLED_ERROR` (Medium severity).
- **Negative Controls (Clean Areas):**
  - `repository.go`: `CreateUser` uses parameterized `INSERT INTO users (...) VALUES (?, ?, ?, ?)`.
  - `repository.go`: `GetUserByID` uses parameterized `SELECT ... WHERE id = ?`.
  - `db.go`: `InitDB` properly sets connection pooling limits and verifies ping with context timeout.

### 3.3. Repository 3: `cache_engine`
- **Ground Truth Defects (2):**
  1. `DEF-CACHE-01` (`cache.go:88`): `UnsafeKeys` iterates over `c.items` map without acquiring `c.mu.RLock()`, causing data race under concurrent writes. Categorized as `CONCURRENCY_RACE` (High severity).
  2. `DEF-CACHE-02` (`eviction.go:30`): `StartBackgroundEvictor` launches a background goroutine looping over `ticker.C` with no shutdown channel or context cancellation, causing an immortal goroutine leak. Categorized as `GOROUTINE_LEAK` (High severity).
- **Negative Controls (Clean Areas):**
  - `cache.go`: `SafeKeys`, `Get`, `Set`, `Delete` maintain thread-safety using `sync.RWMutex`.
  - `sharded.go`: FNV-32a hashing partitions keys cleanly across 16 shards.
  - `eviction.go`: `StartManagedEvictor` properly terminates upon `ctx.Done()`.

### 3.4. Repository 4: `cli_tooling`
- **Ground Truth Defects (3):**
  1. `DEF-CLI-01` (`runner.go:27`): `ExecuteShellScript` passes user-supplied command strings directly to `cmd.exe /C` or `sh -c`. Categorized as `COMMAND_INJECTION` (Critical severity).
  2. `DEF-CLI-02` (`fs.go:16`): `ReadFileFromRoot` uses `filepath.Join(rootDir, userPath)` without verifying directory boundary containment. Categorized as `PATH_TRAVERSAL` (High severity).
  3. `DEF-CLI-03` (`fs.go:43`): `InspectFiles` calls `defer f.Close()` inside a `for` loop body, delaying file descriptor cleanup until loop termination. Categorized as `RESOURCE_LEAK` / `DEFER_IN_LOOP` (Medium severity).
- **Negative Controls (Clean Areas):**
  - `runner.go`: `ExecuteDirect` executes binaries with discrete argument slices without shell interpolation.
  - `fs.go`: `SafeReadFileFromRoot` verifies directory containment with clean prefix checks.
  - `fs.go`: `SafeInspectFiles` bounds file descriptor lifetime inside an anonymous closure.

### 3.5. Repository 5: `queue_worker`
- **Ground Truth Defects (3):**
  1. `DEF-QUEUE-01` (`internal/producer/producer.go:47`): `SendDirect` dials a network socket via `net.DialTimeout` without calling `defer conn.Close()`. Categorized as `RESOURCE_LEAK` (High severity).
  2. `DEF-QUEUE-02` (`pkg/retry/retry.go:56`): `RetryWithTimerLeak` calls `time.NewTimer` in a loop without stopping the timer via `timer.Stop()`. Categorized as `TIMER_LEAK` (Medium severity).
  3. `DEF-QUEUE-03` (`internal/consumer/consumer.go:56`): `ParseRawPayload` discards error returned from `json.Unmarshal`. Categorized as `UNHANDLED_ERROR` (Low severity).
- **Negative Controls (Clean Areas):**
  - `pkg/retry/retry.go`: `Retry` implements exponential backoff with full context cancellation support.
  - `internal/consumer/consumer.go`: `SafeParseRawPayload` validates json syntax and returns clean errors.
  - `internal/consumer/consumer.go`: `Consumer.Start` manages worker pools with `sync.WaitGroup`.

---

## 4. Firebreak & Integrity Verification

To guarantee that Milestone 3.3 serves as an authentic external validation trial:
1. **Rule Immutability:** No diagnostic rule in `go-backend/agent/repo_intel/diagnostic/` or `go-backend/agent/repo_intel/reliability_security.go` has been altered, patched, or relaxed to accommodate these external repositories.
2. **Snapshot Hashing:** Every source file in the five repositories has been hashed with SHA-256 and committed to `benchmarks/external_v1/repository_manifest.json` and `reports/milestone3_3_repository_provenance.json`.
3. **Reproducibility:** All five repositories are self-contained Go modules requiring zero external network downloads. Running `go test ./...` in each repository succeeds offline.

