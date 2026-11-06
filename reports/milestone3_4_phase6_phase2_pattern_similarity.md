# Milestone 3.4 — Phase 6 / Phase 2: Pattern Similarity Audit

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.4-phase6-independent-pilot`  
**Detector Frozen Commit:** `a64e7632476c1341343481661a3e246c69e5a644`  
**Evaluation Target:** External Injected Defect Corpus (`benchmarks/external_v2`)  
**Status:** COMPLETE / SCIENTIFICALLY CONSERVATIVE

---

## 1. Objective and Evaluation Methodology

The objective of this pattern similarity audit is to assess whether the 6 verified injected defects in `external_v2` represent genuinely novel external defect representations or whether they structurally mirror the synthetic benchmark fixtures used during Milestone 3.4 development.

Scientific validity requires acknowledging whether test cases present authentic semantic diversity or whether high detection recall is driven by structural archetype familiarity.

---

## 2. Per-Defect Pattern Similarity Audit

### Case 1: `INJ-EXT2-SQL-01`
- **Location:** `benchmarks/external_v2/stretchr_objx/sql_helper.go:25`
- **Function:** `SQLQueryExecutor.QueryUserByRole`
- **Defect Pattern:**
  ```go
  query := fmt.Sprintf("SELECT id, name, role FROM users WHERE role = '%s'", role)
  row := e.DB.QueryRowContext(ctx, query)
  ```
- **Comparison to Development Fixtures:**
  Structurally identical to `agent/repo_intel/diagnostic/semantic/fixtures/sql_fixtures.go:TestQueryContextSprintf`. Both use `fmt.Sprintf` with a string format specifier `%s` directly fed into a `database/sql` query method.
- **Novelty Factors:**
  The host repository (`stretchr_objx`) uses a struct receiver `*SQLQueryExecutor` embedding `*sql.DB`, with methods accepting standard Go `context.Context`.
- **Similarity Classification:** **HIGH STRUCTURAL SIMILARITY**

---

### Case 2: `INJ-EXT2-SQL-02`
- **Location:** `benchmarks/external_v2/stretchr_objx/sql_helper.go:31`
- **Function:** `SQLQueryExecutor.SearchByTagRaw`
- **Defect Pattern:**
  ```go
  q := "UPDATE items SET tags = '" + tag + "' WHERE id = " + id
  _, err := e.DB.ExecContext(ctx, q)
  ```
- **Comparison to Development Fixtures:**
  Directly mirrors `agent/repo_intel/diagnostic/semantic/fixtures/sql_fixtures.go:TestConcatSQL`. Both construct a query string via binary expression concatenation (`token.ADD`) of untrusted variables.
- **Novelty Factors:**
  Evaluated within an external package context containing unrelated type declarations and map helper utilities.
- **Similarity Classification:** **HIGH STRUCTURAL SIMILARITY**

---

### Case 3: `INJ-EXT2-CMD-01`
- **Location:** `benchmarks/external_v2/schollz_progressbar/cmd_runner.go:15`
- **Function:** `TriggerProgressHook`
- **Defect Pattern:**
  ```go
  cmd := exec.CommandContext(ctx, "sh", "-c", hookScript)
  return cmd.Run()
  ```
- **Comparison to Development Fixtures:**
  Matches the shell wrapper pattern in `agent/repo_intel/diagnostic/semantic/fixtures/command_fixtures.go:TestShellWrapperCommand`.
- **Novelty Factors:**
  Framed as a progress hook runner interacting with context cancellation.
- **Similarity Classification:** **HIGH STRUCTURAL SIMILARITY**

---

### Case 4: `INJ-EXT2-RES-02`
- **Location:** `benchmarks/external_v2/fatih_color/file_printer.go:21`
- **Function:** `DumpColorProfile`
- **Defect Pattern:**
  ```go
  f, err := os.Open(cfgPath)
  if err != nil {
      return err
  }
  _, err = io.Copy(os.Stdout, f)
  return err // Missing f.Close() along return
  ```
- **Comparison to Development Fixtures:**
  Identical to standard missing `os.Open` close in `agent/repo_intel/diagnostic/semantic/fixtures/resource_fixtures.go`.
- **Novelty Factors:**
  Operates within `fatih_color` alongside terminal color formatting methods.
- **Similarity Classification:** **HIGH STRUCTURAL SIMILARITY**

---

### Case 5: `INJ-EXT2-RES-03`
- **Location:** `benchmarks/external_v2/gorilla_websocket/trace_conn.go:19`
- **Function:** `ProbeWebSocketGateway`
- **Defect Pattern:**
  ```go
  conn, err := net.DialTimeout("tcp", addr, timeout)
  if err != nil {
      return nil, err
  }
  buf := make([]byte, 256)
  n, _ := conn.Read(buf)
  return buf[:n], nil // Missing conn.Close() along return
  ```
- **Comparison to Development Fixtures:**
  Follows network acquisition lifecycle archetype. In development fixtures, `net.Dial` tests were scoped to mock helpers.
- **Novelty Factors:**
  Network socket lifecycle combined with partial slice return; tests escape analysis and resource retention across function returns.
- **Similarity Classification:** **MODERATE-TO-HIGH STRUCTURAL SIMILARITY**

---

### Case 6: `INJ-EXT2-ERR-04`
- **Location:** `benchmarks/external_v2/spf13_cast/json_loader.go:23`
- **Function:** `ParsePayloadUnchecked`
- **Defect Pattern:**
  ```go
  var res map[string]interface{}
  _ = json.Unmarshal(raw, &res)
  return res
  ```
- **Comparison to Development Fixtures:**
  Directly matches `agent/repo_intel/diagnostic/semantic/fixtures/error_fixtures.go:TestIgnoredUnmarshal`.
- **Novelty Factors:**
  Placed in `spf13_cast` alongside casting functions that frequently handle dynamic interface types.
- **Similarity Classification:** **HIGH STRUCTURAL SIMILARITY**

---

## 3. Generalization Implications and Epistemic Limits

The pattern similarity audit reveals a crucial epistemological boundary:

1. **What Has Been Proven:**
   - **Host Code Invariance:** The semantic diagnostic engine executes stably and deterministically on large, complex, real-world open-source Go repositories without AST crashing, CFG stack overflows, symbol table resolution failures, or unreasonable execution latencies (total run duration across 10 repos was 1.38 seconds).
   - **Noise Resistance:** The engine demonstrated high specificity (95.0% on clean controls; 0 security false alarms across thousands of lines of open-source Go code).
   - **Accurate Injected Retrieval:** The engine flawlessly identified all 6 injected defect archetypes embedded within authentic third-party repositories.

2. **What Has NOT Been Proven:**
   - **True Natural Generalization:** Because all 4 proposed naturally occurring defects were disqualified during Gate 1 forensic audit, the engine's performance on *naturally occurring, subtle, real-world bug idioms* remains **untested and not estimable**.
   - **Syntactic Generalization Gap:** The injected defects represent canonical syntactic instantiations of the target vulnerabilities. They do not test obscure or highly convoluted idioms (e.g., reflections, multi-hop indirect channel passing, interface-wrapped SQL builders).

---

## 4. Conclusion and Recommendation

The 100% recall on Stratum B must be reported with the strict qualification that **it reflects injected archetype recall in external host code**, not verified natural defect recall. The primary headline metric must present the stratified performance faithfully:

$$\text{Stratum B (Injected Defect Recall)} = 100.0\% \quad (95\%\text{ CI: } [60.97\%, 100.0\%])$$
$$\text{Stratum C (Clean Control Specificity)} = 95.0\% \quad (95\%\text{ CI: } [76.39\%, 99.11\%])$$
$$\text{Stratum A (Natural Defect Recall)} = \text{NOT ESTIMABLE } (n=0)$$

