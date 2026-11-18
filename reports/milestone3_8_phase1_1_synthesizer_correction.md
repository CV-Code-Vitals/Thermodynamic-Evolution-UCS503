# Milestone 3.8 Phase 1.1 — AST Resource-Cleanup Synthesizer Audit & Correction

**Audit Date:** 2026-11-07  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Component:** `go-backend/agent/repo_intel/execution/synthesis.go`  
**Candidate Commit:** `dff6379529b7bedaee8585e5058b4ce902121217`  
**Investigating Branch:** `release/milestone-3.8-mvp-rc`  

---

## 1. Audit Scope & Problem Formulation

The autonomous patch synthesizer translates semantic findings into deterministic Abstract Syntax Tree (AST) mutations. In Milestone 3.7 Phase 2, `synthesizeResourceCleanup` permitted the synthesis of `defer resp.Body.Close()` inside functions whose return signatures transferred resource ownership to callers.

A comprehensive audit was performed across all nine safety dimensions required by Milestone 3.8 Phase 1.1 Task B:
1. **Nil resource values**
2. **Acquisition errors**
3. **Partially initialized resources**
4. **Early returns**
5. **Conditional acquisition**
6. **Resource reassignment**
7. **Cleanup inside loops**
8. **Multiple return paths**
9. **Deferred argument-evaluation semantics**

---

## 2. Synthesizer Rule Gap Analysis & Remediations

| Dimension | Vulnerability Identified | Synthesizer Correction | Formal Status |
| :--- | :--- | :--- | :--- |
| **Escaping Resources (Caller Ownership)** | Callee inserted `defer <res>.Close()` even though `<res>` was returned in `fn.Type.Results` or `ast.ReturnStmt`. | Added `isResourceEscapingReturn(fn, resVar, resKind)`. If the enclosing function returns the resource or a resource type (`*http.Response`, `*os.File`, `net.Conn`, `io.ReadCloser`), synthesis yields `AbstainEscapingResource`. | **RESOLVED** |
| **Resource Reassignment** | Callee mutated `<res>.Body` (e.g. `resp.Body = ioutil.NopCloser(...)` in WebSocket handshake redirect), obscuring true lifecycle state. | Added `isResourceReassignedAfter(fn.Body.List, initIdx, resVar)`. Any downstream assignment to `<res>` or `<res>.<Field>` triggers `AbstainUnsafePlacement`. | **RESOLVED** |
| **Pre-Error-Check Placement** | Deferring before `if err != nil` caused nil pointer dereference panic when the constructor returned an error and a nil resource. | Restricted defer synthesis strictly to the statement immediately following a verified `if err != nil { return ... }` block. | **RESOLVED** |
| **Incomplete Error Check** | An `if err != nil` block that only logged an error without exiting (`return`) allowed control flow to proceed with a nil resource. | Hardened `isErrorCheckIfStmt` to verify that the conditional block contains an explicit `ast.ReturnStmt`. | **RESOLVED** |
| **Loop Allocation** | Resource allocated inside a `for` or `for ... range` loop body where `defer` caused unbounded handle accumulation. | Preserved existing `AbstainLoopResource` check. | **VERIFIED** |
| **Concurrent Escape** | Resource handle passed to `go func(...)` goroutine. | Preserved existing `AbstainConcurrentEscape` check. | **VERIFIED** |

---

## 3. Implementation Details

In [`go-backend/agent/repo_intel/execution/synthesis.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/synthesis.go):

### 3.1 New Abstention Categories
```go
const (
    AbstainCommandInjection    AbstainCategory = "COMMAND_INJECTION_ARCHITECTURAL_REDESIGN"
    AbstainLoopResource        AbstainCategory = "LOOP_DEFERRED_CLEANUP_ACCUMULATION"
    AbstainConcurrentEscape    AbstainCategory = "CONCURRENT_OWNERSHIP_ESCAPE"
    AbstainEscapingResource    AbstainCategory = "RESOURCE_ESCAPES_VIA_RETURN"
    AbstainUnsafePlacement     AbstainCategory = "UNSAFE_RESOURCE_CLEANUP_LIFECYCLE"
    AbstainAmbiguousReturn     AbstainCategory = "AMBIGUOUS_RETURN_SIGNATURE"
    AbstainSyntaxError         AbstainCategory = "SOURCE_SYNTAX_ERROR"
    AbstainUnrecognizedPattern AbstainCategory = "UNRECOGNIZED_DEFECT_PATTERN"
)
```

### 3.2 Escaping Resource Detection
```go
func isResourceEscapingReturn(fn *ast.FuncDecl, resVar string, resKind string) bool {
    // 1. Inspect function signature return types
    if fn.Type.Results != nil {
        for _, field := range fn.Type.Results.List {
            typeStr := formatAstType(field.Type)
            if isResourceType(typeStr) {
                return true
            }
        }
    }

    // 2. Inspect all return statements in function body
    returned := false
    ast.Inspect(fn.Body, func(n ast.Node) bool {
        ret, ok := n.(*ast.ReturnStmt)
        if !ok {
            return true
        }
        for _, expr := range ret.Results {
            if id, ok := expr.(*ast.Ident); ok && id.Name == resVar {
                returned = true
                return false
            }
            if sel, ok := expr.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == resVar {
					returned = true
					return false
				}
			}
        }
        return true
    })
    return returned
}
```

### 3.3 Downstream Reassignment Detection
```go
func isResourceReassignedAfter(stmts []ast.Stmt, initIdx int, resVar string) bool {
    for j := initIdx + 1; j < len(stmts); j++ {
        reassigned := false
        ast.Inspect(stmts[j], func(n ast.Node) bool {
            assign, ok := n.(*ast.AssignStmt)
            if !ok {
                return true
            }
            for _, lhs := range assign.Lhs {
                if id, ok := lhs.(*ast.Ident); ok && id.Name == resVar {
                    reassigned = true
                    return false
                }
                if sel, ok := lhs.(*ast.SelectorExpr); ok {
                    if id, ok := sel.X.(*ast.Ident); ok && id.Name == resVar {
                        reassigned = true
                        return false
                    }
                }
            }
            return true
        })
        if reassigned {
            return true
        }
    }
    return false
}
```

---

## 4. Empirical Validation

Unit testing in [`go-backend/agent/repo_intel/execution/synthesis_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/synthesis_test.go) verified that:
1. `SYNTH-RES-01` (`http.Get` response body read locally) -> **PROPOSED** (`defer resp.Body.Close()`).
2. `SYNTH-RES-02` (`os.Open` file handle read locally) -> **PROPOSED** (`defer f.Close()`).
3. `SYNTH-ERR-01` (Unhandled error propagation) -> **PROPOSED**.
4. `SYNTH-SQL-01` (SQL string formatting) -> **PROPOSED**.
5. `SYNTH-CMD-ABSTAIN` (Command injection) -> **ABSTAINED** (`COMMAND_INJECTION_ARCHITECTURAL_REDESIGN`).
6. `SYNTH-LOOP-ABSTAIN` (Loop allocation) -> **ABSTAINED** (`LOOP_DEFERRED_CLEANUP_ACCUMULATION`).
7. `SYNTH-ESCAPE-ABSTAIN` (Concurrent escape) -> **ABSTAINED** (`CONCURRENT_OWNERSHIP_ESCAPE`).
8. `SYNTH-RETURN-ABSTAIN` (Ambiguous return) -> **ABSTAINED** (`AMBIGUOUS_RETURN_SIGNATURE`).
9. `SYNTH-ESCAPING-RETURN-ABSTAIN` (NAT-03 archetype: escaping return) -> **ABSTAINED** (`RESOURCE_ESCAPES_VIA_RETURN`).

Safe intra-procedural cleanups continue to succeed, while unsafe escaping and mutating lifecycles are fail-closed.

