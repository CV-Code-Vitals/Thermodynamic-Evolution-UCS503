# Milestone 3.5 — Phase 2: Bounded Field-Sensitive Struct Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Implementation File:** `go-backend/agent/repo_intel/diagnostic/semantic/field_sensitive.go`  
**Test Suite:** `go-backend/agent/repo_intel/diagnostic/semantic/field_sensitive_test.go`  
**Status:** COMPLETE / VERIFIED

---

## 1. Executive Summary

Milestone 3.5 Phase 2 introduces bounded access-path tracking for local struct objects. The implementation allows the semantic engine to track resources embedded within struct fields (e.g. `h.file`, `h.conn`, `obj.inner.file`) while keeping unrelated fields (e.g. `cfg.SQL` vs `cfg.Timeout`) strictly isolated.

### Key Capabilities Delivered:
1. **Access-Path Decomposition:** Deconstructs nested selectors up to configured bound ($\text{MaxDepth} = 3$).
2. **Explicit Field State Machine:** Models lifecycle transitions across `OWNED_LOCAL`, `OWNED_TRANSFERRED`, `RELEASED`, `MAYBE_RELEASED`, and `ESCAPED_UNKNOWN`.
3. **Dual Initialization Vectors:** Supports both direct field assignment (`h.file = f`) and composite literal initialization (`h := &Holder{file: f}`).
4. **Struct Pointer & Copy Aliasing:** Tracks struct aliases (`p := h; p.file.Close()`) and resolves field operations back to the canonical access path.
5. **Ownership Escape via Return:** Recognizes when a struct holding a resource handle is returned to the caller (`return h`), classifying the resource as `OWNED_TRANSFERRED` rather than an unclosed leak.
6. **Independence of Unrelated Fields:** Prevents false alarms when non-resource fields (strings, ints, flags) are assigned or modified.

---

## 2. Access-Path State Machine & Soundness Rules

```
       [ Local Resource f ]
                 │
                 ▼  (h.file = f or &Holder{file: f})
         [ OWNED_LOCAL ]
           │         │
(return h) │         │ (h.file.Close() or defer h.file.Close())
           ▼         ▼
[ OWNED_TRANSFERRED ] [ RELEASED ]
```

### Invariants:
1. **Never Assume Automatic Cleanup:** Storing a handle into a struct field (`h.file = f`) is NOT proof of release. If `h.file` is not closed or returned, it remains `OWNED_LOCAL` and is flagged as an unclosed resource leak.
2. **Never Conflate Unrelated Fields:** An access path `cfg.Timeout = 30` or `cfg.SQL = query` does not trigger resource tracking and does not interfere with tracked file or connection handles.
3. **Depth Budgeting:** Access paths exceeding $\text{MaxDepth} = 3$ (e.g. `a.b.c.d.file`) are marked as `IsTruncated` and safely degrade to `ESCAPED_UNKNOWN` without failing analysis.

---

## 3. Empirical Test Suite Results

The implementation was verified against seven automated test cases in `field_sensitive_test.go`:

| Test Name | Evaluated Pattern | Expected State | Observed State | Status |
| :--- | :--- | :--- | :--- | :---: |
| `TestFieldSensitive_AssignmentAndRelease` | `h.file = f; defer h.file.Close()` | `RELEASED` | `StateResourceDeferredRelease`; 0 leaks | **PASS** |
| `TestFieldSensitive_CompositeLiteralInit` | `h := &Holder{file: f}; defer h.file.Close()` | `RELEASED` | `StateResourceDeferredRelease`; 0 leaks | **PASS** |
| `TestFieldSensitive_OwnershipEscapeThroughReturn` | `h := &Holder{file: f}; return h` | `OWNED_TRANSFERRED` | `StateResourceTransferred`; clean exit | **PASS** |
| `TestFieldSensitive_StructPointerAlias` | `p := h; defer p.file.Close()` | Alias resolved to `h.file` | `StateResourceDeferredRelease`; 0 leaks | **PASS** |
| `TestFieldSensitive_NestedAccessPath` | `obj.inner.file = f; obj.inner.file.Close()` | Depth 2 `RELEASED` | Path tracked at depth 2; `RELEASED` | **PASS** |
| `TestFieldSensitive_UnrelatedFieldsIndependent` | `cfg.SQL = "..."`; `cfg.Timeout = 30` | 0 resource fields | `len(tracker.Fields) == 0` | **PASS** |
| `TestFieldSensitive_StoredAndNeverReleased_Counterfactual`| `h := &Holder{file: f}; _ = h` | `OWNED_LOCAL` unclosed | Confirmed `RESOURCE_LEAK_CONFIRMED` | **PASS** |

---

## 4. Controlled Counterfactual Verification

- **Safe Case (`TestFieldSensitive_AssignmentAndRelease`):** When `h.file = f` is accompanied by `defer h.file.Close()`, the analyzer maps the release back to $f$ and certifies a clean lifecycle.
- **Counterfactual Case (`TestFieldSensitive_StoredAndNeverReleased_Counterfactual`):** When `h.file = f` is stored but neither `h.file.Close()` nor `return h` is executed, the analyzer detects that the handle is held in a locally owned struct without release, correctly issuing a high-confidence diagnostic finding.
