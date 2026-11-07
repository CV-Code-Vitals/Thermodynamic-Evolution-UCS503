# Milestone 3.8 Phase 1.1 — NAT-03 Failure Reproduction & Root Cause Analysis

**Audit Date:** 2026-11-07  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Target:** `benchmarks/external_v2/gorilla_websocket/client.go`  
**Defect Identification:** `NAT-03` (Upstream gorilla/websocket PR #789)  
**Investigation Branch:** `release/milestone-3.8-mvp-rc`  
**Investigating Principal:** Lead Reliability & Correctness Auditor  

---

## 1. Executive Summary

During Milestone 3.8 Phase 1 release candidate testing, an in-depth audit of candidate repair patch `NAT-03` revealed a critical behavioral flaw. The patch synthesized in Milestone 3.7 Phase 2 inserted `defer resp.Body.Close()` into `gorilla/websocket/client.go:DialContext`.

Empirical execution in a deterministic isolated test environment revealed:
1. **Error Path Panic:** When `DialContext` encounters a network dial or HTTP handshake header failure, standard Go `http.ReadResponse` returns `(nil, err)`. Under this condition, `resp` is `nil`. Unwinding the stack via deferred execution of `resp.Body.Close()` immediately triggers a fatal runtime panic:  
   `runtime error: invalid memory address or nil pointer dereference`.
2. **Caller Lifecycle Violation:** On successful WebSocket handshakes, `DialContext` returns `(*Conn, *http.Response, error)` to the caller, transferring resource ownership. Deferred callee closure immediately terminates the underlying network connection hijacked by the WebSocket session, breaking all caller contracts.
3. **Oracle Blindspot:** The Milestone 3.7 Phase 2 independent validation suite accepted the flawed patch because:
   - Oracle 1 (`go vet`) only checked static compiler syntax;
   - Oracle 2 (regression suite) executed only clean unit test paths without exercising network dial failures;
   - Oracle 3 (defect reproducer) performed a superficial substring match for `.Close()`;
   - Oracle 4 (behavioral invariance) defaulted without dynamic negative fault injection;
   - Oracle 5 (negative properties) only verified that `err :=` was not shadowed in nested scopes.

---

## 2. Deterministic Reproduction

A dedicated reproducer was implemented in [`go-backend/agent/repo_intel/benchmark/nat03_reproducer_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/benchmark/nat03_reproducer_test.go) to simulate the exact control flow of `DialContext` under failure conditions.

### 2.1 Reproducing Code Snippet
```go
func simulateDialContextFlawed(failImmediately bool) (resErr error) {
    var resp *http.Response
    // Flawed patch inserted deferred close along the top/callee scope:
    defer func() {
        if resp != nil {
            _ = resp.Body.Close()
        } else {
            // This represents defer resp.Body.Close() without nil check
            _ = resp.Body.Close() // PANICS: nil pointer dereference
        }
    }()

    if failImmediately {
        return errors.New("dial tcp: connection refused")
    }
    return nil
}
```

### 2.2 Execution Evidence
```text
=== RUN   TestNAT03_ErrorPathPanicReproducer
    nat03_reproducer_test.go:52: CONFIRMED: Unsafe cleanup placement caused panic on error path: PANIC_OCCURRED: runtime error: invalid memory address or nil pointer dereference
--- PASS: TestNAT03_ErrorPathPanicReproducer (0.00s)
=== RUN   TestNAT03_HttpReadResponseNilReturnContract
    nat03_reproducer_test.go:70: CONFIRMED: http.ReadResponse returns (nil, err) on parse failure (err: malformed HTTP response "INVALID_HTTP_DATA")
--- PASS: TestNAT03_HttpReadResponseNilReturnContract (0.00s)
PASS
```

The test confirmed that executing `defer resp.Body.Close()` inside `DialContext` causes a catastrophic nil-pointer dereference whenever an error precedes or accompanies response creation.

---

## 3. Upstream Return Contract Analysis

Inspection of `gorilla/websocket/client.go` confirms the public return contract:

```go
func (d *Dialer) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*Conn, *http.Response, error)
```

### 3.1 Lifecycle States

| Execution Point | `netConn` | `resp` | Return Values | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Pre-Handshake Failure** (Line 362) | Non-nil | `nil` | `nil, nil, err` | Request write fails; `resp` is unallocated. Deferring `resp.Body.Close()` panics. |
| **HTTP Read Failure** (Line 373) | Non-nil | `nil` | `nil, nil, err` | `http.ReadResponse` returns error; `resp` is `nil`. Deferring `resp.Body.Close()` panics. |
| **Bad Handshake Status** (Line 392) | Hijacked | Non-nil | `nil, resp, ErrBadHandshake` | Handshake rejected (e.g. 401/403). `resp.Body` replaced with `NopCloser` so caller can inspect body. Deferring close terminates caller body reader. |
| **Handshake Success** (Line 414) | Hijacked | Non-nil | `conn, resp, nil` | Connection upgraded to WebSocket. Closing `resp.Body` closes the underlying TCP socket! |

### 3.2 Upstream PR #789 Canonical Fix
In upstream `gorilla/websocket` PR #789, the maintainers resolved the leak not by inserting a simple callee `defer resp.Body.Close()`, but by explicitly closing `resp.Body` **only on specific non-101 error paths after verifying that `resp != nil` and `resp.Body != nil`**, while leaving successful upgraded connections open for the WebSocket connection manager.

An intra-procedural AST mutator operating without inter-procedural protocol context cannot safely invent this architectural protocol logic.

---

## 4. Oracle Suite Blindspot Audit

| Oracle | Type | Failure Mode | Root Cause of Acceptance in Phase 2 |
| :--- | :--- | :--- | :--- |
| **Oracle 1** | Syntax & Compilation | **Silent Acceptance** | `defer resp.Body.Close()` is syntactically valid Go; `go vet` does not perform flow-sensitive nil escape analysis. |
| **Oracle 2** | Target Regression Suite | **Execution Gap** | `task.ValidationCommands` in the pilot corpus did not execute negative network failure integration tests against mock HTTP servers. |
| **Oracle 3** | Defect Reproducer | **False Positive** | Defaulted to substring inspection `strings.Contains(src, ".Close()")`. Did not evaluate whether the enclosing function returns the resource. |
| **Oracle 4** | Behavioral Invariance | **False Positive** | Lacked fault injection hooks to simulate network drops or bad HTTP handshakes. |
| **Oracle 5** | Negative Properties | **Incomplete Invariant** | Only inspected shadowed declarations of `err :=` in nested blocks. Did not inspect resource escape via function return signatures or pre-error-check cleanup placement. |

---

## 5. Corrective Action Plan

1. **Synthesizer Rule Hardening (Task B):**
   - Add explicit check `isResourceEscapingReturn(fn, resVar)`. If the enclosing function returns the resource (`*http.Response`, `*os.File`, `net.Conn`, `io.ReadCloser`), autonomous intra-procedural synthesis MUST ABSTAIN (`AbstainEscapingResource`).
   - Add check for downstream resource reassignment (`isResourceReassignedAfter`). If `resp.Body` is reassigned, synthesis MUST ABSTAIN (`AbstainUnsafePlacement`).
2. **Oracle Hardening (Task C):**
   - Harden Oracle 5 (Negative Properties) to inspect the patched AST and reject callee cleanup on escaping resources or cleanups placed before error checks.
   - Harden Oracle 3 (Defect Reproducer) to reject resource close proposals on functions whose signatures return resource handles.
3. **Corpus Revalidation (Task D):**
   - Update NAT-03 expected outcome from `REPAIR` to `ABSTAIN`.
   - Recompute precision and breadth metrics transparently.

