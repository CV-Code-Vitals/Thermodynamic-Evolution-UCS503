# MILESTONE 3.4 — PHASE 5 CROSS-PACKAGE RESOLUTION REPORT
## Bounded Interprocedural Call Resolution, Explicit State Machines & Wrapper Auditing

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Baseline Commit:** `be917a4`  
**Evaluation Date:** November 2026  
**Classification:** Interprocedural Semantic Analysis Specification  

---

## 1. Executive Summary

Static analysis in large Go repositories frequently encounters boundaries across functions, files, packages, and external dependencies. A major finding from Milestone 3.3 was that intra-procedural detectors fail when resources or error checks are handled inside helper functions or wrappers across files.

Phase 5 introduces a conservative, bounded interprocedural call resolution engine ([`callgraph.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/callgraph.go)). The resolver models caller-callee relationships with explicit mathematical state transitions, ensuring that certainty is never assumed without evidence.

---

## 2. Explicit Resolution States

The engine maps every call site to one of five mutually exclusive resolution states:

| Resolution State | Constant Name | Operational Semantics | Example Call Site | Remediation Safety |
|---|---|---|---|---|
| **RESOLVED_EXACT** | `CallResolvedExact` | Target function or method is uniquely defined within the same package or repository AST | `loadConfig("app.yaml")` where `loadConfig` is defined in `config.go` | Eligible for autonomous reasoning |
| **RESOLVED_CONSERVATIVE** | `CallResolvedConservative` | Target is matched via known wrapper contract or standard library summary | `os.Open(path)` matched to stdlib `os.Open` contract | Eligible for autonomous reasoning |
| **MULTIPLE_TARGETS** | `CallMultipleTargets` | Call site resolves to two or more candidate targets (e.g. interface dispatch with multiple implementations) | `worker.Execute()` where `worker` may be `FastWorker` or `SafeWorker` | Human review required |
| **EXTERNAL_OPAQUE** | `CallExternalOpaque` | Call target belongs to an imported third-party package without available source AST | `gorm.Open(...)`, `aws.Client.Send(...)` | Fail-closed (assumed opaque) |
| **UNRESOLVED** | `CallUnresolved` | Receiver type is unknown, dynamic reflection, or AST lookup failed | `fn.(func())()` | Quarantined from autonomous action |

---

## 3. Supported Call Resolution Features

The call-site resolver (`CallSiteResolver`) implements conservative resolution for 10 structural patterns:

1. **Local Function Calls:** Direct function invocations within the same file (e.g., `check(err)`).
2. **Same-Package Cross-File Calls:** Functions declared in another file of the same Go package.
3. **Imported Package Calls:** Static invocations with qualified selectors (e.g., `os.Open`, `exec.Command`).
4. **Method Receivers:** Methods invoked on identifiable struct pointers or value types within the analyzed package.
5. **Interface Dispatch:** Concrete resolution when the underlying concrete assignment is statically provable within the local scope or reaching definitions.
6. **Package Aliases:** Renamed imports (e.g., `import ssql "database/sql"` correctly resolving `ssql.Open`).
7. **Type Aliases:** Aliased types mapped back to underlying canonical types.
8. **Named Return Values:** Propagation of named returns across inlined or summarized helper scopes.
9. **Resource Wrappers:** Identification of helper functions that take a resource and close it (e.g., `closeFile(f)`), or acquire and return it (e.g., `openDB() (*sql.DB, error)`).
10. **Error Propagation Helpers:** Identification of error audit functions (e.g., `checkErr(err)` terminating execution or propagating).

---

## 4. Handling Cycles, Recursion, and Edge Cases

- **Call Cycles & Recursion:** Call graph traversal employs cycle detection via visited sets. Recursive functions terminate expansion immediately without infinite loops.
- **Bounded Expansion Depth:** To prevent combinatorial explosion in large projects, call chain tracing is bounded to depth $K=3$. Deeper chains fall back to `CallUnresolved`.
- **Missing Dependencies & Incomplete Types:** If an import path cannot be located on disk, the analyzer classifies it as `CallExternalOpaque` rather than fabricating certainty.
- **Fail-Closed Principle:** Under no circumstances does an `EXTERNAL_OPAQUE` or `UNRESOLVED` call permit the assumption that a resource was closed or an error was checked. Such findings remain quarantined.

---

## 5. Verification Results

In [`unified_orchestrator_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/diagnostic/semantic/unified_orchestrator_test.go):

```
=== RUN   TestCrossPackageCallResolution
--- PASS: TestCrossPackageCallResolution (0.01s)
```

The test verified:
- Local intra-file calls resolve with `CallResolvedExact`.
- Standard library calls (`os.Open`, `http.Get`, `exec.Command`) resolve with `CallResolvedConservative`.
- External package calls (`github.com/foo/bar.Do`) resolve with `CallExternalOpaque`.
- All edges maintain line numbers, caller symbols, and package identity.

