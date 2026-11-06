# Milestone 3.5 — Phase 1: Semantic Effects Taxonomy & Propagation Rules

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Classification:** Formal Semantic Modeling & Dataflow Propagation Specifications  
**Status:** COMPLETE / SPECIFICATION LOCKED

---

## 1. Formal Taxonomy of Semantic Effects

| Effect Enum | Category | Formal Meaning | Preconditions | Postconditions in Caller Context |
| :--- | :---: | :--- | :--- | :--- |
| `TAINT_PASSTHROUGH` | Dataflow | Argument is returned unmodified or passed directly to return tuple index $R$. | Parameter $P$ flows to `return` without mutation. | Return value at index $R$ inherits identical taint label and origin as argument at index $P$. |
| `TAINT_TRANSFORM` | Dataflow | Argument is formatted (`fmt.Sprintf`), concatenated (`+`), or converted into return string. | Parameter $P$ is an operand in string formatting or binary expression reaching `return`. | Return value at index $R$ is marked tainted if argument at index $P$ is tainted. |
| `TAINT_SANITIZE_VERIFIED` | Dataflow | Parameter passes through a context-verified sanitization contract. | Function applies verified validation (e.g. integer conversion, strict whitelist regex). | Return value is marked safe for the specified domain; taint flag is cleared. |
| `RESOURCE_ACQUIRE` | Lifecycle | Function instantiates and returns a tracked system resource handle. | Callee contains acquisition API call (`os.Open`, `net.Dial`) whose handle is returned. | Caller registers newly received variable as an active resource requiring explicit release. |
| `RESOURCE_RELEASE` | Lifecycle | Function closes or disposes of a resource handle passed via parameter. | Callee invokes `.Close()` or `.Stop()` along all normal reachable return paths. | Resource instance in caller is transitioned to `StateResourceReleased` at the call site. |
| `RESOURCE_TRANSFER` | Lifecycle | Function transfers resource ownership to another owner (e.g. struct storage, external queue). | Resource is stored into a long-lived receiver or external consumer. | Resource instance in caller is marked `StateResourceTransferred`; no local leak is flagged. |
| `ERROR_RETURN` | Error | Function produces an `error` interface return value on failure paths. | Callee return tuple includes `error` type. | Caller call site is registered as requiring error handling verification. |
| `ERROR_PROPAGATE` | Error | Function receives an incoming error and propagates it (wrapped or unwrapped) to return tuple. | Incoming error parameter is returned or wrapped via `fmt.Errorf("%w")`. | Caller error lineage is maintained through the wrapper function. |
| `EFFECT_UNKNOWN` | Fallback | Behavior cannot be proven due to incomplete source, external module, or analysis budget timeout. | Target is external opaque, dynamic interface with no known implementation, or analysis timed out. | Conservative fallback: does NOT clear taint, does NOT assume resource is closed. |

---

## 2. Invariant Rules for Conservative Propagation

### 2.1 The Principle of Pessimistic Incomplete Knowledge
> **Invariant 1:** *An unknown function call (`EFFECT_UNKNOWN`) must never be assumed to sanitize taint or release a resource.*
- If a tainted variable is passed to an unknown function `foo(x)` and the result `y` is used in a SQL sink, `y` must remain classified as potentially tainted unless an explicit sanitization contract is verified.
- If an active resource handle `f` is passed to an unknown function `bar(f)`, `bar` must **not** be assumed to close `f`. The caller retains ownership obligation unless `bar` has an explicit `RESOURCE_RELEASE` summary.

### 2.2 String Transformation vs. Sanitization
> **Invariant 2:** *String manipulation (concatenation, trimming, upper/lower casing, replacements) constitutes `TAINT_TRANSFORM`, NEVER `TAINT_SANITIZE_VERIFIED`.*
- Replacing quotes or escaping characters without parameterized binding is prone to injection bypasses and must not suppress SQL or Command injection findings.
- Only parameterization placeholders (`$1`, `?`) combined with bound parameter lists or infallible type coercions (e.g. `strconv.Atoi`) qualify as safe.

### 2.3 Resource Ownership Transition Invariants
> **Invariant 3:** *Resource handles have exactly one responsible owner at any point along a control-flow path.*
- **Callee Acquisition (`RESOURCE_ACQUIRE`):** When `f, err := pkg_b.AcquireFile(path)` is called, the return value transfers ownership to the caller. The caller's CFG must ensure `f` is closed or returned along all exits.
- **Callee Release (`RESOURCE_RELEASE`):** When `pkg_b.CloserHelper(f)` is called, caller relinquishes cleanup responsibility at that call site block. All downstream exits reachable from that call site treat `f` as closed.

### 2.4 Ambiguous Dynamic Dispatch Merging
> **Invariant 4:** *When a call site targets an interface with multiple possible implementations, summaries are merged via conservative join ($\sqcap$).*
- **Taint:** If *any* implementation exhibits `TAINT_PASSTHROUGH` or `TAINT_TRANSFORM`, the merged effect is tainted.
- **Resource Cleanup:** A resource is considered released only if *all* possible implementations exhibit `RESOURCE_RELEASE`. If any implementation leaves the resource open, the merged summary reports `EFFECT_UNKNOWN` / unreleased.
