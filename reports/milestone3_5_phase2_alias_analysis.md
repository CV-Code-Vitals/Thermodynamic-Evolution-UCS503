# Milestone 3.5 — Phase 2: Bounded Multi-Hop Alias Analysis

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `research/milestone-3.5-interprocedural-semantics`  
**Implementation File:** `go-backend/agent/repo_intel/diagnostic/semantic/alias_analysis.go`  
**Test Suite:** `go-backend/agent/repo_intel/diagnostic/semantic/alias_analysis_test.go`  
**Status:** COMPLETE / VERIFIED

---

## 1. Executive Summary

Milestone 3.5 Phase 2 replaces the legacy one-hop reachability filter with a formal, bounded multi-hop alias analysis engine. The engine models direct assignments, transitive alias chains, pointer dereferences, branch-dependent uncertainties, reassignments, and cyclic alias relationships.

### Key Capabilities Delivered:
1. **Transitive Chain Traversal:** Bounded worklist algorithm traverses chains of arbitrary depth up to the configured budget ($\text{MaxHops} = 5$).
2. **Reassignment and Invalidation:** When a variable is overwritten or reassigned (`alias = other`), the engine invalidates the alias relationship for all program points following the reassignment line.
3. **Certainty Classification:** Distinguishes unconditional aliases (`MUST_ALIAS`) from control-flow-dependent aliases (`MAY_ALIAS`).
4. **Pointer Aliasing:** Unpacks pointer creation (`p := &f`) and pointer dereference (`f2 := *p`).
5. **Cycle Protection:** Graph cycle detection terminates safely on circular assignments (`a = b; b = a`) without infinite loops or stack overflow.
6. **Budget Enforcement:** Tracks when an alias chain exceeds `MaxHops` and marks the root variable as `Truncated`, conservatively suppressing unverified assumptions.

---

## 2. Mathematical Model & Algorithm Design

### 2.1 Formal Data Structures

```go
type AliasRelationType string

const (
    AliasMustAlias AliasRelationType = "MUST_ALIAS"
    AliasMayAlias  AliasRelationType = "MAY_ALIAS"
    AliasNoAlias   AliasRelationType = "NO_ALIAS"
)

type AliasEdge struct {
    TargetVar    string
    Relation     AliasRelationType
    HopCount     int
    AssignedLine int
    IsPointer    bool
}
```

### 2.2 Worklist Fixed-Point Algorithm

Let $V$ be the set of local variables and $E \subseteq V \times V \times \text{Relation} \times \mathbb{N}$ be directed alias edges.

For a root variable $r \in V$:
1. Initialize worklist $W = \{(r, 0, \text{MUST\_ALIAS})\}$, visited set $S = \{r\}$, result closure $C = \emptyset$.
2. While $W \neq \emptyset$:
   - Pop $(v, h, rel)$ from $W$.
   - If $h \ge \text{MaxHops}$ or $|C| \ge \text{MaxSetSize}$:
     - Mark $r$ as `Truncated` and continue.
   - For each outgoing edge $e = (v \to v', rel', \text{line}) \in E$:
     - If $v' \in S$: continue (cycle protection).
     - Add $v'$ to $S$.
     - Compute composed relation:
       $$\text{effRel} = \begin{cases} \text{MAY\_ALIAS} & \text{if } rel = \text{MAY\_ALIAS} \lor rel' = \text{MAY\_ALIAS} \\ \text{MUST\_ALIAS} & \text{otherwise} \end{cases}$$
     - Insert $v' \mapsto (v', \text{effRel}, h+1, \text{line})$ into $C$.
     - Push $(v', h+1, \text{effRel})$ into $W$.

### 2.3 Reassignment Invalidation

For any query $\text{IsAliasAtLine}(r, c, L)$:
1. If $c \notin C$: return $(\text{false}, \text{NO\_ALIAS})$.
2. If edge assignment line $L_{\text{assign}} > L$: return $(\text{false}, \text{NO\_ALIAS})$ (alias not yet established).
3. For each reassignment line $L_{\text{inv}} \in \text{Invalidated}[c]$:
   - If $L_{\text{assign}} < L_{\text{inv}} \le L$: return $(\text{false}, \text{NO\_ALIAS})$ (alias was killed).
4. Return $(\text{true}, \text{effRel})$.

---

## 3. Empirical Test Suite Results

The implementation was verified against six automated test cases in `alias_analysis_test.go`:

| Test Name | Tested Scenario | Expected Outcome | Observed Outcome | Status |
| :--- | :--- | :--- | :--- | :---: |
| `TestAliasAnalysis_TransitiveChainsAndBudget` | Chains of length 1, 2, 3, 5, and 6 | Hops 1–5 in closure; hop 6 truncated | Hops 1–5 MUST_ALIAS; a6 truncated; Truncated=true | **PASS** |
| `TestAliasAnalysis_ReassignmentAndInvalidation` | `alias := f1; alias = f2` | Line 7 valid; Line 13 invalidated | Line 7 MUST_ALIAS; Line 13 NO_ALIAS | **PASS** |
| `TestAliasAnalysis_PointerAliases` | `p := &f; f2 := *p` | `p` is pointer alias; `f2` is deref alias | Both in closure; IsPointer=true for p | **PASS** |
| `TestAliasAnalysis_BranchDependentMayAlias` | `if cond { target = f }` | `target` is MAY_ALIAS | Relation=MAY_ALIAS | **PASS** |
| `TestAliasAnalysis_CyclicGraph` | `a = b; b = a` | Safe termination; b in closure | Terminated in 0.00s; 0 cycles | **PASS** |
| `TestAliasAnalysis_ResourceReleaseThroughMultiHopAlias` | `f -> h1 -> h2 -> h3 -> CloserHelper(h3)` | Clean exit certified without false leak | FP eliminated; StateConfirmed = 0 | **PASS** |

---

## 4. Resolution of the Transitive Aliasing Limitation

In Milestone 3.5 Phase 1.1, `TestAdversarial_Resource_MultipleAliases` confirmed that a 3-hop alias chain (`alias1 := f; alias2 := alias1; CloserHelper(alias2)`) produced a false positive leak alarm due to the 1-hop reachability filter.

With the new `AliasGraph`:
- `instances[i].Aliases` receives the complete transitive closure $\{h_1, h_2, h_3\}$.
- When `CloserHelper(h3)` is evaluated, `h3` matches the alias closure of $f$.
- The false positive leak alarm is completely eliminated, converting the test scenario from a documented limitation into a **fully supported capability**.
