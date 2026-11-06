# Milestone 3.7 — Phase 1.5: Autonomous Patch Synthesis Engine Architecture & Design

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Starting Commit:** `92f1c721bb6d7128295a0c63e117d2e8b85bd458`  
**Document Type:** Technical Architecture & Semantic Patch Synthesis Specification  
**Status:** IMPLEMENTED & EMPIRICALLY VERIFIED  

---

## 1. Executive Summary & Paradigm Shift

In prior milestones (such as Milestone 3.6), fixture repairs were validated using curated replacement strings in `controlled_remediation_test.go` to prove workspace sandboxing, rollback mechanics, and multi-stage verification pipelines. However, as audited in Phase 0 of this milestone, predetermined fixture strings do not constitute autonomous remediation.

Milestone 3.7 Phase 1.5 establishes a **genuine, bounded AST-level patch synthesis engine** located in [`go-backend/agent/repo_intel/execution/synthesis.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/synthesis.go). The synthesizer inspects the Abstract Syntax Tree of the target source file, infers the exact semantic repair intent, applies deterministic AST mutations, formats the modified tree, prunes unused imports, computes unified diffs and SHA-256 digests, and verifies syntax before submitting candidate patches to the approval and execution pipeline.

Crucially, the synthesizer operates with **fail-closed conservative abstention**: when semantic ambiguities, concurrent escapes, or architectural redesigns are required, it explicitly yields `ABSTAIN` rather than generating unsafe or malformed code.

---

## 2. Bounded Synthesis Pipeline Architecture

The engine implements a five-stage deterministic state machine:

```
+-------------------+     +------------------+     +-------------------+
|  SemanticFinding  | --> |   RepairIntent   | --> |     AST Parse     |
| (Category, Target)|     | (Intent / Abstain|     | (go/parser, fset) |
+-------------------+     +------------------+     +-------------------+
                                                             |
                                                             v
+-------------------+     +------------------+     +-------------------+
| SynthesisDecision | <-- | Multi-Oracle Val | <-- |   AST Mutation    |
| (Proposed/Abstain)|     | (go vet, diff)   |     | (go/ast rewriting)|
+-------------------+     +------------------+     +-------------------+
```

### 2.1 Interface Definitions

```go
type IntentType string
const (
    IntentResourceCleanup IntentType = "RESOURCE_CLEANUP"
    IntentHandleError     IntentType = "HANDLE_ERROR"
    IntentParameterizeSQL IntentType = "PARAMETERIZE_SQL"
    IntentAbstain         IntentType = "ABSTAIN"
)

type SynthesisDecision struct {
    Status     SynthesisDecisionStatus `json:"status"` // PROPOSED, ABSTAINED, FAILED
    Intent     *RepairIntent           `json:"intent"`
    Patch      *CandidatePatch         `json:"patch,omitempty"`
    Reason     string                  `json:"reason"`
    DurationMs int64                   `json:"duration_ms"`
}
```

---

## 3. Supported Deterministic Mutation Patterns

### 3.1 Pattern 1: Bounded Resource Leak Remediation (`SynthesizeResourceCleanup`)

* **Defect Class:** Leaked operating system descriptors, network sockets, database connections, and HTTP response bodies.
* **Semantic Analysis & Trigger:**
  1. The synthesizer traverses the AST statements of each function declaration (`*ast.FuncDecl`).
  2. Detects variable assignment (`*ast.AssignStmt`) calling resource-allocating constructors (`http.Get`, `http.Post`, `http.Do`, `os.Open`, `os.Create`, `sql.Open`, `db.Query`).
  3. Inspects the next statement in the block. It must be an error guard (`if err != nil { return ... }`).
  4. Scans the remainder of the function block to verify no existing `.Close()` call exists.
* **AST Mutation:**
  * For HTTP calls (`http.Response`): synthesizes a deferred selector call:
    `defer resp.Body.Close()`
  * For OS files, DB handles, and rows: synthesizes:
    `defer <varName>.Close()`
  * Inserts the `*ast.DeferStmt` immediately following the error check block.
* **Formatting:** Uses `go/format` to generate idiomatic Go source, ensuring exact indentation and comment preservation.

### 3.2 Pattern 2: Unhandled Error Propagation (`SynthesizeErrorHandling`)

* **Defect Class:** Error return values ignored via blank identifier (`_ = fn()`) or discarded expression statements (`fn()`).
* **Semantic Analysis & Trigger:**
  1. Inspects the enclosing function's return signature (`fn.Type.Results`).
  2. Verifies that the enclosing function returns an `error` as its final return value. If not, the engine aborts and abstains.
  3. Inspects preceding return types to construct typed zero-value expressions:
     * Pointers, slices, maps, channels, interfaces: `nil`
     * Strings: `""`
     * Integers / Floats / Bytes: `0`
     * Booleans: `false`
* **AST Mutation:**
  * For assignments discarding error (`_ = os.Remove(path)`):
    * Mutates assignment LHS identifier from `_` to `err`.
    * Elevates assignment token from `=` (`token.ASSIGN`) to `:=` (`token.DEFINE`) to ensure variable declaration.
    * Inserts guard `if err != nil { return <zeros...>, err }`.
  * For naked call expressions (`os.Remove(path)`):
    * Rewrites statement into:
      `if err := os.Remove(path); err != nil { return <zeros...>, err }`

### 3.3 Pattern 3: SQL Parameterization (`SynthesizeSQLParameterization`)

* **Defect Class:** Dynamic query string construction using `fmt.Sprintf` with format specifiers (`%s`, `%d`, `%v`).
* **Semantic Analysis & Trigger:**
  1. Locates `fmt.Sprintf` call assigned to a query variable (`query := fmt.Sprintf(...)`).
  2. Inspects format string literal (`*ast.BasicLit`): verifies presence of `%s`, `'%s'`, `"%s"`, or `%d`.
  3. Extracts format arguments (`call.Args[1:]`).
* **AST Mutation:**
  * Rewrites format string into static SQL with placeholder arguments:
    `"SELECT id, username FROM users WHERE username = '%s'"` $\rightarrow$ `"SELECT id, username FROM users WHERE username = ?"`
  * Scans downstream AST statements for database execution methods (`db.Exec`, `db.ExecContext`, `db.Query`, `db.QueryContext`, `db.QueryRow`, `db.QueryRowContext`).
  * Appends the extracted format arguments directly to the database execution call as bound parameters.
  * **Automated Import Clean-up:** Traverses `node.Decls` and `node.Imports`. If `fmt` is no longer referenced anywhere else in the file, it removes `import "fmt"` to prevent compiler unused-import errors.

---

## 4. Conservative Abstention Guardrails

A primary safety requirement of Phase 1.5 is that the synthesizer must **fail closed** when safe, deterministic synthesis cannot be mathematically guaranteed.

| Abstention Category | Semantic Trigger Condition | Architectural Rationale |
| :--- | :--- | :--- |
| `COMMAND_INJECTION_ARCHITECTURAL_REDESIGN` | Detection of `sh -c`, `cmd.exe /C`, or `exec.Command` with concatenated user input | Dynamic shell interpolation cannot be safely rewritten into positional arguments without re-architecting how the program invokes executables. Automated rewriting risks command string truncation or parsing bugs. |
| `LOOP_DEFERRED_CLEANUP_ACCUMULATION` | Resource allocation occurs inside `*ast.ForStmt` or `*ast.RangeStmt` | In Go, `defer` statements execute only when the enclosing function returns. Inserting `defer Close()` inside a loop body delays release until the loop completes, causing file descriptor or socket pool exhaustion. |
| `CONCURRENT_OWNERSHIP_ESCAPE` | Resource handle is passed to `*ast.GoStmt` or sent across a channel | Concurrently shared resources have non-lexical lifetimes. Deferring closure in the spawning thread creates race conditions and use-after-close panics. |
| `AMBIGUOUS_RETURN_SIGNATURE` | Target function signature returns void or does not end in `error` | Synthesizing error propagation requires modifying the public function signature and updating all call sites across the codebase, which exceeds bounded file-local scope. |
| `SOURCE_SYNTAX_ERROR` | File fails AST parsing (`parser.ParseFile`) | Engine refuses to operate on malformed input to prevent AST corruption. |

---

## 5. Verification & Diff Derivation

Following AST manipulation:

1. **Compilation Pre-Check:** The synthesized string is re-parsed through `parser.ParseFile`. If any syntax error exists, the decision immediately yields `FAILED`.
2. **Unified Diff Generation:** The synthesizer invokes `formatUnifiedDiff(relPath, originalSrc, modifiedSrc)`, computing unified diff hunks (`--- a/...`, `+++ b/...`, `@@ ... @@`).
3. **Cryptographic Diff Binding:** The SHA-256 digest of the diff content is calculated:
   $$\text{DiffHash} = \text{SHA256}(\text{UnifiedDiff})$$
   This hash is subsequently bound to the `ApprovalToken` to prevent any post-authorization tampering.

---

## 6. Implementation Verification Matrix

The synthesis engine was validated in [`go-backend/agent/repo_intel/execution/synthesis_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/synthesis_test.go):

| Test Case | Defect Pattern | Expected Decision | Actual Decision | Verified Diff Produced |
| :--- | :--- | :---: | :---: | :---: |
| `SYNTH-RES-01` | HTTP Response Body Leak | `PROPOSED` | `PROPOSED` | `defer resp.Body.Close()` |
| `SYNTH-RES-02` | OS File Descriptor Leak | `PROPOSED` | `PROPOSED` | `defer f.Close()` |
| `SYNTH-ERR-01` | Discarded Error Return | `PROPOSED` | `PROPOSED` | `err := os.Remove(...)` + `if err != nil` |
| `SYNTH-SQL-01` | Dynamic SQL Formatting | `PROPOSED` | `PROPOSED` | `?` placeholder + bound param + removed `fmt` |
| `SYNTH-CMD-ABSTAIN` | Shell Command Injection | `ABSTAINED` | `ABSTAINED` | None (Explicit Abstain) |
| `SYNTH-LOOP-ABSTAIN` | Resource in Range Loop | `ABSTAINED` | `ABSTAINED` | None (Explicit Abstain) |
| `SYNTH-ESCAPE-ABSTAIN` | Goroutine Resource Escape | `ABSTAINED` | `ABSTAINED` | None (Explicit Abstain) |
| `SYNTH-RETURN-ABSTAIN` | Void Return Function | `ABSTAINED` | `ABSTAINED` | None (Explicit Abstain) |

**Result:** 100% concordance with architectural specifications, zero regressions, and zero hardcoded patch payloads.
