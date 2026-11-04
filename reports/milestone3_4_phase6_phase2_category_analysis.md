# Milestone 3.4 — Phase 6 / Phase 2: Category-Level Diagnostic Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Detector Frozen Baseline:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Ground Truth Baseline:** `benchmarks/external_v2/ground_truth_amendment_v2_1.json`  
**Execution Timestamp:** `2026-11-05T05:19:25+05:30`  
**Status:** COMPLETE / RIGOROUSLY STRATIFIED

---

## 1. Executive Category Breakdown

The blinded first-look evaluation evaluated the frozen semantic diagnostic engine across the four targeted defect categories expanded during Milestone 3.4:

| Category | Injected Defects ($n$) | Detected (TP) | Missed (FN) | Category Recall | Clean Controls ($n$) | Clean TN | False Alarms (FP) | Category Specificity |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **SQL Injection** | 2 | 2 | 0 | **100.0%** | 1 | 1 | 0 | **100.0%** |
| **Command Injection** | 1 | 1 | 0 | **100.0%** | 1 | 1 | 0 | **100.0%** |
| **Resource Lifecycle** | 2 | 2 | 0 | **100.0%** | 2 | 2 | 0 | **100.0%** |
| **Unhandled Error** | 1 | 1 | 0 | **100.0%** | 16 | 15 | 1 | **93.75%** |
| **Overall (Primary Strata B+C)** | **6** | **6** | **0** | **100.0%** | **20** | **19** | **1** | **95.00%** |

---

## 2. Deep Dive by Diagnostic Category

### 2.1 SQL Expression & Injection Analysis
- **Target Cases:**
  - `INJ-EXT2-SQL-01` (`stretchr_objx/sql_helper.go:25`, `SQLQueryExecutor.QueryUserByRole`): Sprintf concatenation of user parameter into `db.QueryRowContext`.
  - `INJ-EXT2-SQL-02` (`stretchr_objx/sql_helper.go:31`, `SQLQueryExecutor.SearchByTagRaw`): String addition (`+`) of untrusted tag directly into SQL clause for `db.ExecContext`.
- **Clean Controls:**
  - `CTL-EXT2-16` (`stretchr_objx/sql_helper.go:19`, `QueryUserSafe`): Parameterized placeholder query (`SELECT ... WHERE id = $1`).
- **Performance:**
  - Recall: 2 / 2 (100.0%).
  - Specificity: 1 / 1 (100.0%).
  - False Alarms in Pristine Open-Source Code: 0 across all 10 repositories.
- **Architectural Mechanism:**
  The `SQLAnalyzer` operates on AST call expressions to known SQL execution sinks (`Query`, `QueryRowContext`, `Exec`, `ExecContext`). It uses reaching-definitions dataflow to trace query arguments back to string concatenation (`token.ADD`) or formatting functions (`fmt.Sprintf`), while verifying that parameterized placeholders (`?`, `$1`) with bound parameter lists are recognized as safe and suppressed.

---

### 2.2 Command Execution Analysis
- **Target Cases:**
  - `INJ-EXT2-CMD-01` (`schollz_progressbar/cmd_runner.go:15`, `TriggerProgressHook`): Passing unsanitized user-supplied script string directly into shell wrapper execution sink `exec.CommandContext(ctx, "sh", "-c", hookScript)`.
- **Clean Controls:**
  - `CTL-EXT2-19` (`schollz_progressbar/cmd_runner.go:8`, `RunTerminalResetSafe`): Constant, fixed binary invocation with literal static arguments (`exec.Command("tput", "reset")`).
- **Performance:**
  - Recall: 1 / 1 (100.0%).
  - Specificity: 1 / 1 (100.0%).
  - False Alarms in Pristine Open-Source Code: 0 across all 10 repositories.
- **Architectural Mechanism:**
  The `CommandAnalyzer` analyzes calls to `os/exec` APIs (`Command`, `CommandContext`). It evaluates whether the binary is an interpreter or shell (`sh`, `bash`, `cmd.exe`) and inspects whether the command arguments are constructed dynamically from untrusted function parameters without argument array tokenization.

---

### 2.3 Resource Lifecycle Analysis
- **Target Cases:**
  - `INJ-EXT2-RES-02` (`fatih_color/file_printer.go:21`, `DumpColorProfile`): File descriptor acquisition via `os.Open(cfgPath)` without `Close()` or `defer` along exit path.
  - `INJ-EXT2-RES-03` (`gorilla_websocket/trace_conn.go:19`, `ProbeWebSocketGateway`): Network socket acquisition via `net.DialTimeout("tcp", addr, timeout)` where the socket handle is used but never closed.
- **Clean Controls:**
  - `CTL-EXT2-12` (`fatih_color/file_printer.go:10`, `PrintFileSafe`): Properly guarded file read with immediate `defer f.Close()`.
  - `CTL-EXT2-17` (`gorilla_websocket/trace_conn.go:9`, `DialProbeSafe`): Network socket closed via `defer conn.Close()`.
- **Performance:**
  - Recall: 2 / 2 (100.0%).
  - Specificity: 2 / 2 (100.0%).
  - False Alarms on Clean Controls: 0 (100.0%).
- **Stratum D Findings:**
  - Found 2 unadjudicated resource lifecycle findings in upstream code:
    1. `gorilla_websocket/conn.go:447`: `time.Timer` conditional release (released on 3 paths, leaked along 3 of 6 reachable exits).
    2. `schollz_progressbar/examples/download/main.go:14`: Unclosed `http.Response.Body` in example code.
- **Architectural Mechanism:**
  The `ResourceAnalyzer` constructs a control flow graph (CFG) for each function, identifies resource acquisition points from the extensible API registry (`os.Open`, `net.Dial`, `http.Get`), performs intraprocedural reaching-definitions and alias tracking, monitors exit reachability (including fallthrough blocks and explicit `return` statements), and checks for release operations (`Close()`, `defer`) or resource escape (returning the resource to caller).

---

### 2.4 Error-Handling & Return-Value Analysis
- **Target Cases:**
  - `INJ-EXT2-ERR-04` (`spf13_cast/json_loader.go:23`, `ParsePayloadUnchecked`): Call to `json.Unmarshal(raw, &target)` with error return silently ignored.
- **Clean Controls:**
  - 16 control functions across the 10 repositories (`CTL-EXT2-01` through `CTL-EXT2-11`, `CTL-EXT2-13` through `CTL-EXT2-15`, `CTL-EXT2-18`, `CTL-EXT2-20`).
- **Performance:**
  - Recall: 1 / 1 (100.0%).
  - Specificity: 15 / 16 (93.75%).
  - False Positive: 1 (`CTL-EXT2-02` in `google_uuid/hash.go:27-30`).
- **Root Cause of the False Positive:**
  In `google_uuid/hash.go:27`, the function `NewHash(h hash.Hash, ...)` executes:
  ```go
  h.Reset()
  h.Write(space[:]) //nolint:errcheck
  h.Write(data)     //nolint:errcheck
  ```
  The Go standard library `hash.Hash` interface embeds `io.Writer`, whose signature is `Write(p []byte) (n int, err error)`. Consequently, `h.Write` technically returns an `error`. However, in standard hash implementations (MD5, SHA1, SHA256), the in-memory write operation cannot fail and always returns `nil` error. The upstream library author explicitly documented this with `//nolint:errcheck`. The semantic analyzer correctly noticed that the return error was not checked, but lacked an exception rule for in-memory hash writers.
- **Stratum D Findings:**
  - 49 unadjudicated error findings were identified across upstream code (e.g., ignored buffer writes, print returns, internal helper errors).

---

## 3. Unaddressed Defect Categories

To prevent misleading claims of universal defect coverage, the following five defect categories are explicitly recognized as unaddressed by Milestone 3.4 semantic analysis:

1. **Concurrency Race Conditions:**  
   Unsynchronized access to shared state across goroutines without mutex protection or atomic operations (e.g. `DEF-CACHE-01` in Milestone 3.3). Requires happens-before analysis and goroutine interleaving models.
2. **Goroutine Leaks:**  
   Goroutines blocked indefinitely on unbuffered channel sends/receives or missing stop triggers (e.g. `DEF-CACHE-02` in Milestone 3.3).
3. **Context Cancel Leaks:**  
   Contexts created via `context.WithCancel` / `WithTimeout` whose cancel functions are never invoked along exit branches (e.g. `DEF-HTTP-03` in Milestone 3.3).
4. **Path Traversal / Directory Escape:**  
   Constructing filesystem paths with untrusted user input without validation via `filepath.Clean` and prefix boundary enforcement (e.g. `DEF-CLI-02` in Milestone 3.3).
5. **Timer Reference Leaks in Loops:**  
   Use of `time.After(...)` within loops or select statements causing uncollected timer channel allocations (e.g. `DEF-QUEUE-02` in Milestone 3.3).

---

## 4. Summary Verdict

The category analysis confirms:
- 100% recall across SQL, Command, Resource, and Error handling on verified injected defects in external repositories.
- Zero false positives on SQL, Command, and Resource clean controls.
- Exactly 1 false positive on Error clean controls due to the known Go `hash.Hash.Write` interface convention.
- Zero security false positives on pristine external open-source code.
