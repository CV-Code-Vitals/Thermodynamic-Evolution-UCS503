# MILESTONE 3.4 — PHASE 3 KNOWN LIMITATIONS & RESEARCH BOUNDARIES
## Scientific Boundaries of Function-Level Resource Lifecycle Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.4-semantic-analysis`  
**Phase:** Phase 3 — Function-Level Resource Lifecycle Analysis  

---

## 1. Explicit Research Scope & Non-Claims

The Phase 3 implementation establishes function-level control-flow reachability analysis for system resources in Go. To maintain scientific integrity, the platform explicitly defines its operational boundaries and does not claim:

1. **Whole-Program Whole-Heap Escape Analysis:** The analyzer does not model the global heap or solve arbitrary pointer aliasing across third-party dynamic libraries.
2. **Arbitrary Interprocedural Call Chains:** Interprocedural analysis is limited to recognized ownership transfer contracts (`return f`, `consume*(f)`, `transfer*(f)`). Arbitrary deep call stacks are not recursively traversed without interprocedural summaries.
3. **Arbitrary Struct Encapsulation Verification:** Resources stored into custom struct fields whose lifecycle is managed by unverified external constructors are marked as `RESOURCE_UNRESOLVED` rather than claiming definitive safety or leak.

---

## 2. Detailed Technical Limitations

### 2.1 Interprocedural Ownership Boundaries
- **Current Behavior:** When a resource handle is passed as an argument to a user-defined function `processHandle(h)`:
  - If the function name conveys consumption/transfer semantics (`consume...`, `transfer...`), ownership is treated as transferred.
  - If the function is an unknown utility (e.g. `logMetadata(h)`), the local function retains its release obligation.
- **Limitation:** In the absence of whole-program summary caches, if `logMetadata` unexpectedly assumes ownership or closes the resource, local analysis cannot deduce this without interprocedural summaries.
- **Mitigation & Safety Invariant:** When ambiguous, the finding is assigned `FindingState: INSUFFICIENT_EVIDENCE` and quarantined from autonomous remediation.

### 2.2 Goroutine Lifecycles & Asynchronous Closures
- **Scenario:**
  ```go
  f, err := os.Open(path)
  go func() {
      defer f.Close()
      process(f)
  }()
  return nil
  ```
- **Current Behavior:** The analyzer detects that `f` escapes the synchronous scope of the enclosing function into a concurrent closure.
- **Limitation:** The analyzer cannot statically verify synchronization primitives (e.g. `sync.WaitGroup`, context cancellation) that ensure the asynchronous goroutine completes before program termination.
- **Classification:** Categorized as `RESOURCE_ESCAPED` / `INSUFFICIENT_EVIDENCE`.

### 2.3 Struct Field Lifetime Tracking
- **Scenario:**
  ```go
  type Session struct {
      Conn net.Conn
  }
  func OpenSession() *Session {
      c, _ := net.Dial("tcp", "localhost:8080")
      return &Session{Conn: c}
  }
  ```
- **Current Behavior:** Tested in fixture `Pos_Res_StoredInStructUnknownLifecycle`. The analyzer flags `isStoredInStruct = true`, transitioning to `RESOURCE_UNRESOLVED`.
- **Safety Invariant:** `FindingState` is set to `INSUFFICIENT_EVIDENCE` (`Confidence: 0.60`). The autonomous execution engine will **never** generate an automated patch inserting `defer c.Close()` into `OpenSession`, as doing so would close the socket before the returned `Session` could use it.

### 2.4 Channel Ownership Transfers
- **Scenario:**
  ```go
  func DispatchFile(ch chan *os.File, path string) {
      f, _ := os.Open(path)
      ch <- f
  }
  ```
- **Current Behavior:** Sending a resource handle over a channel transfers custody across thread boundaries.
- **Limitation:** The analyzer recognizes that `f` has escaped local scope, but does not track whether receiver goroutines properly drain the channel and close received descriptors.

---

## 3. Autonomous Execution & Remediation Safety Gating

To prevent unsafe code modifications, the remediation engine enforces strict qualification gates:

| Lifecycle State | Diagnostic Finding State | Autonomous Remediation Policy | Justification |
|---|---|---|---|
| `RESOURCE_LEAK_CONFIRMED` | `CONFIRMED` (`Confidence >= 0.95`) | **ELIGIBLE** | 100% of reachable normal exit paths leak; local `defer Close()` is safe. |
| `RESOURCE_LEAK_PROBABLE` | `PROBABLE` (`Confidence >= 0.85`) | **REQUIRES HUMAN APPROVAL** | Partial path leaks or in-loop defers require context-aware refactoring. |
| `RESOURCE_UNRESOLVED` | `INSUFFICIENT_EVIDENCE` | **STRICTLY QUARANTINED** | Never eligible for autonomous execution; prevents closing active struct handles. |
| `RESOURCE_DEFERRED_RELEASE` | `COUNTER_EVIDENCE_PRESENT` | **BENIGN** | Validated clean lifecycle; no remediation required. |
| `RESOURCE_RELEASED` | `COUNTER_EVIDENCE_PRESENT` | **BENIGN** | Validated clean lifecycle; no remediation required. |
| `RESOURCE_TRANSFERRED` | `COUNTER_EVIDENCE_PRESENT` | **BENIGN** | Ownership transferred to caller or consumer; no remediation required. |

---

## 4. Planned Extensions for Subsequent Phases

1. **Phase 4 (AST-Based Error-Handling Analysis):** Complement resource tracking with structured analysis of error return patterns (`if err != nil`), error wrapping idioms, and error masking.
2. **Phase 5 (Cross-Package Semantic Integration):** Construct interprocedural summary tables for repository-internal packages, allowing precise tracking across function call boundaries within the same module.
