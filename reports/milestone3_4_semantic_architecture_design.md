# Milestone 3.4 — Phase 1: Semantic Intermediate Representation (IR) Architecture Design

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Active Branch:** `research/milestone-3.4-semantic-analysis`  
**Base Release Commit:** `52ff0f1822ff8c25bf64d2e860124b9dc9b7351b`  
**Target:** Formal Architecture Specification for Lightweight Semantic Analysis  
**Date:** 2026-11-04  

---

## 1. Design Overview & Goals

The Semantic Intermediate Representation (IR) provides a robust, conservative semantic foundation for the Go repository intelligence and diagnostic engine. It bridges the gap between raw AST/regex matching and heavy whole-program compiler backends.

### Key Design Goals:
1. **Lightweight & Fast:** Parse and analyze typical Go packages (hundreds of files) in milliseconds without requiring external compiler dependencies.
2. **Graceful Type Degradation:** Leverage standard `go/types` for exact type checking when standard/local packages are present; gracefully fall back to AST signature and selector inference when third-party module dependencies are missing offline.
3. **Explicit Multi-Exit Control Flow:** Construct basic blocks that explicitly represent branches (`if`, `switch`, `select`), loops (`for`, `range`), jump targets (`break`, `continue`, `return`), and deferred function executions (`defer`).
4. **General Dataflow Framework:** Provide forward and backward dataflow analysis with bitset/set representations, gen/kill sets, and fixed-point worklist algorithms.
5. **Specialized Diagnostic Projections:** Project semantic facts directly onto high-value defect models:
   - Resource Lifecycles (Acquisition $\to$ Release on all exit paths).
   - Error Propagation (Error generation $\to$ Checking/Return/Suppression).
   - Untrusted Dataflow / Taint Tracking (Source $\to$ Transform $\to$ Sink).
   - Concurrency & Synchronization (Mutex balance, goroutine lifecycle, context cancellation).

---

## 2. Architecture & Data Structures

```
                             +------------------------+
                             |   Source Go Files      |
                             +------------------------+
                                         |
                                         v
                         +------------------------------+
                         |     AST & Type Loader        |
                         |  (go/ast, go/token, go/types)|
                         +------------------------------+
                                         |
         +-------------------------------+-------------------------------+
         |                               |                               |
         v                               v                               v
+------------------+           +-------------------+           +-------------------+
|   Symbol Table   |           | Control Flow (CFG)|           | Dataflow Engine   |
| - Types          |           | - Basic Blocks    |           | - Reaching Defs   |
| - Signatures     |           | - Branch Edges    |           | - Def-Use Chains  |
| - Method Sets    |           | - Defer Stacks    |           | - Taint Lattice   |
+------------------+           +-------------------+           +-------------------+
         \                               |                               /
          \                              |                              /
           +-----------------------------+-----------------------------+
                                         |
                                         v
                        +----------------------------------+
                        | Semantic Diagnostic Projections  |
                        | - Expression-Aware SQL/Command   |
                        | - Path-Sensitive Resource State  |
                        | - Semantic Error Flow            |
                        | - Concurrency & Goroutine State  |
                        +----------------------------------+
```

### 2.1 Package & Symbol Model (`symbol_table.go`)

The symbol table indexes definitions across files in a package:

```go
type SymbolKind string

const (
    SymbolFunc      SymbolKind = "FUNC"
    SymbolMethod    SymbolKind = "METHOD"
    SymbolType      SymbolKind = "TYPE"
    SymbolVar       SymbolKind = "VAR"
    SymbolConst     SymbolKind = "CONST"
    SymbolParam     SymbolKind = "PARAM"
    SymbolInterface SymbolKind = "INTERFACE"
)

type Symbol struct {
    Name         string             `json:"name"`
    Kind         SymbolKind         `json:"kind"`
    Package      string             `json:"package"`
    File         string             `json:"file"`
    Line         int                `json:"line"`
    Exported     bool               `json:"exported"`
    ReceiverType string             `json:"receiver_type,omitempty"`
    ReceiverName string             `json:"receiver_name,omitempty"`
    TypeString   string             `json:"type_string"`
    Signature    *FunctionSignature `json:"signature,omitempty"`
    ASTNode      ast.Node           `json:"-"`
}

type FunctionSignature struct {
    Parameters []ParamInfo  `json:"parameters"`
    Results    []ReturnInfo `json:"results"`
    Variadic   bool         `json:"variadic"`
}

type ParamInfo struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type ReturnInfo struct {
    Name    string `json:"name,omitempty"`
    Type    string `json:"type"`
    IsError bool   `json:"is_error"`
}
```

### 2.2 Control Flow Graph (CFG) Model (`cfg.go`)

Every function or method body is compiled into a directed graph of basic blocks:

```go
type BlockKind string

const (
    BlockEntry      BlockKind = "ENTRY"
    BlockExit       BlockKind = "EXIT"
    BlockNormal     BlockKind = "NORMAL"
    BlockCondition  BlockKind = "CONDITION"
    BlockLoopHeader BlockKind = "LOOP_HEADER"
    BlockLoopBody   BlockKind = "LOOP_BODY"
)

type Block struct {
    ID           int          `json:"id"`
    Kind         BlockKind    `json:"kind"`
    Statements   []ast.Stmt   `json:"-"`
    Terminator   ast.Expr     `json:"-"` // e.g. branch condition
    Predecessors []*Block     `json:"-"`
    Successors   []*Block     `json:"-"`
    TrueBranch   *Block       `json:"-"`
    FalseBranch  *Block       `json:"-"`
    Defers       []*DeferSite `json:"-"`
    LineStart    int          `json:"line_start"`
    LineEnd      int          `json:"line_end"`
}

type DeferSite struct {
    Call       *ast.CallExpr `json:"-"`
    TargetFunc string        `json:"target_func"`
    TargetVar  string        `json:"target_var"`
    MethodName string        `json:"method_name"`
    Line       int           `json:"line"`
}

type ControlFlowGraph struct {
    FunctionName string       `json:"function_name"`
    ReceiverType string       `json:"receiver_type,omitempty"`
    File         string       `json:"file"`
    Entry        *Block       `json:"-"`
    Exit         *Block       `json:"-"`
    Blocks       []*Block     `json:"blocks"`
    ReturnBlocks []*Block     `json:"-"`
    AllDefers    []*DeferSite `json:"defers"`
}
```

### 2.3 Intraprocedural Dataflow Framework (`dataflow.go`)

The dataflow framework solves reaching definitions and definition-use associations:

$$\text{In}[B] = \bigcup_{P \in \text{Pred}(B)} \text{Out}[P]$$
$$\text{Out}[B] = \text{Gen}[B] \cup (\text{In}[B] \setminus \text{Kill}[B])$$

```go
type Definition struct {
    ID           int       `json:"id"`
    VarName      string    `json:"var_name"`
    TypeName     string    `json:"type_name"`
    BlockID      int       `json:"block_id"`
    Line         int       `json:"line"`
    IsParameter  bool      `json:"is_parameter"`
    Expr         ast.Expr  `json:"-"`
    Stmt         ast.Stmt  `json:"-"`
}

type VariableUse struct {
    VarName string    `json:"var_name"`
    BlockID int       `json:"block_id"`
    Line    int       `json:"line"`
    Expr    ast.Expr  `json:"-"`
}

type ReachingDefinitionsAnalysis struct {
    CFG         *ControlFlowGraph
    Definitions []*Definition
    Gen         map[int]map[int]bool // BlockID -> Set of Def IDs
    Kill        map[int]map[int]bool // BlockID -> Set of Def IDs
    In          map[int]map[int]bool // BlockID -> Set of Def IDs
    Out         map[int]map[int]bool // BlockID -> Set of Def IDs
    DefUseChains map[int][]VariableUse // Def ID -> Uses
}
```

---

## 3. Specialized Diagnostic Projections

The Semantic IR directly powers the 5 major diagnostic expansions planned for Phases 2–6:

### 3.1 Expression-Aware SQL & Command Analysis
- Recognizes calls to formatting functions (`fmt.Sprintf`, `fmt.Sprint`, `strings.Join`, string builders) as dynamic expression trees.
- Recursively inspects argument expressions: if any argument traces back to a parameter or external input, the query/command is classified as untrusted dynamic injection.
- Emits explicit negative certification (`SAFE_PARAMETERIZED_SQL`, `SAFE_STATIC_EXEC_ARGUMENTS`) when arguments are proven static constants or parameterized placeholders.

### 3.2 Path-Sensitive Resource Lifecycle Analysis
- Tracks resource acquisitions (`os.Open`, `net.Dial`, `http.Get`, `sql.Open`, `(*sql.DB).Query`, `(*sql.DB).QueryContext`, `(*sql.DB).Begin`, `time.NewTimer`).
- For every acquisition, traverses all forward paths in the CFG from the acquisition block to all `ReturnBlocks` / `Exit`:
  - If a path exists that reaches `Exit` without encountering a matching `.Close()`, `.Rollback()`, or `.Stop()`, AND no deferred release was registered, a resource leak is confirmed.
  - Detects deferred calls registered inside loops (`BlockLoopBody`), flagging delayed descriptor exhaustion.

### 3.3 Semantic Error Flow Analysis
- Inspects every `ast.ExprStmt` (bare expression statements) and unassigned call in `ast.AssignStmt`.
- Resolves the function signature:
  - If the return tuple ends in `error` (e.g. `(int, error)` or bare `error`), and the return is dropped or assigned to `_`, records an error discard event.
  - Distinguishes standard ignored errors (e.g. `defer resp.Body.Close()`) from critical operational errors (`json.Unmarshal`, `tx.Rollback`, `net.Conn.Write`).

### 3.4 Interprocedural Taint & Parameter Dataflow
- Initial sources: Function parameters of public/handler endpoints, HTTP request inputs (`r.URL.Query()`, `r.Body`), CLI arguments (`os.Args`).
- Propagates taint across variable definitions and assignments via Reaching Definitions.
- Sink verification: `exec.Command` arguments, SQL query arguments, filesystem path arguments (`os.Open`, `filepath.Join`).
- Checks for sanitizer calls (e.g. `filepath.Clean` + `strings.HasPrefix` boundary check).

### 3.5 Concurrency & Goroutine Synchronization
- Associates struct mutex fields (`sync.Mutex`, `sync.RWMutex`) with other struct fields.
- Checks that read/write field accesses occur while the corresponding lock is acquired.
- Flags infinite goroutine loops (`*ast.GoStmt`) that lack `select` blocks receiving on a shutdown channel or `ctx.Done()`.
- Verifies that `context.WithTimeout` / `context.WithCancel` cancel functions are invoked via `defer cancel()`.

---

## 4. Verification & Testing Strategy

The Semantic IR will be validated across three testing tiers:
1. **Unit Tests (`semantic/*_test.go`):**
   - Symbol extraction across functions, methods, interfaces, and struct types.
   - CFG generation across branching, nested loops, break/continue, and defers.
   - Reaching definitions and def-use chains on multi-block straight-line and looping code.
2. **Synthetic Validation:**
   - Regression against existing Suite v3 fixtures to ensure 100% preservation of historical precision and recall.
3. **External Repository Verification:**
   - Analysis of external benchmark repositories (`http_router`, `sql_service`, `cache_engine`, `cli_tooling`, `queue_worker`) to confirm rich symbol tables and complete CFG construction.

---

## 5. Summary

This design provides a clean, modular foundation for all subsequent Phase 2 through Phase 6 diagnostic analyzers. With formal symbols, CFGs, and dataflow engines in place, the diagnostic pipeline transitions from rigid AST pattern matching to principled semantic analysis.
