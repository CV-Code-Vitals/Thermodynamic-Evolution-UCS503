# Milestone 3.8 Phase 1.1 — Independent Oracle Suite Hardening

**Audit Date:** 2026-11-07  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Component:** `go-backend/agent/repo_intel/execution/oracles.go`  
**Candidate Commit:** `dff6379529b7bedaee8585e5058b4ce902121217`  
**Investigating Branch:** `release/milestone-3.8-mvp-rc`  

---

## 1. Overview & Objectives

In Milestone 3.7 Phase 2, the five-oracle verification suite allowed a flawed patch for NAT-03 to pass verification. Milestone 3.8 Phase 1.1 Task C mandated hardening the independent oracles to catch:
1. **Panics on error paths** (e.g. nil dereferences caused by premature cleanup).
2. **Multiple `Close()` calls** (redundant double-close anti-patterns).
3. **Inverted error checks** (`if err == nil { return ... }`).
4. **Premature body consumption / escaping resource lifecycle violations**.
5. **Incompatible interface implementations**.

Specifically, the oracles were required to independently reject candidate patches that introduce callee cleanup on escaping resources.

---

## 2. Hardening Interventions by Oracle

### 2.1 Oracle 5: Negative Properties Oracle
Oracle 5 was originally restricted to detecting shadowed `err :=` declarations. It was augmented with comprehensive AST and flow-sensitive invariant inspections:

1. **Callee Cleanup on Escaping Resources:**
   Inspects the patched AST. If the candidate patch introduces a `.Close()` or `defer .Close()` call, Oracle 5 verifies whether the enclosing function returns the resource variable or declares a return parameter of matching resource type (`*http.Response`, `*os.File`, `net.Conn`, `io.ReadCloser`). If an escaping resource is deferred or closed in the callee, Oracle 5 immediately rejects the patch:  
   `Negative property check failed: candidate patch introduces callee cleanup on escaping resource <res> returned to caller; violates ownership transfer and panics on error returns`.

2. **Pre-Error-Check Cleanup Placement:**
   Checks statement order in the enclosing function. If a `Close()` or `defer .Close()` statement is placed prior to the constructor's error check (`if err != nil`), Oracle 5 rejects the patch:  
   `Negative property check failed: resource cleanup for <res> placed prior to error check; triggers nil pointer dereference panic when allocation fails`.

3. **Multiple / Redundant Close Invocations:**
   Traverses the AST to detect multiple invocations of `.Close()` on the same resource variable without intermediate re-acquisition. Flags:  
   `Negative property check failed: multiple Close() calls detected on resource <res> without reassignment`.

4. **Inverted Error Check Conditions:**
   Analyzes candidate diff lines for `if err == nil` patterns that return errors or trigger unexpected early exits:  
   `Negative property check failed: candidate patch introduces inverted error check condition (if err == nil)`.

### 2.2 Oracle 3: Defect-Specific Reproducer Oracle
In addition to running dynamic reproducer commands when supplied, Oracle 3's semantic property analyzer was hardened:
- In resource leak tasks, Oracle 3 now parses the target AST and verifies that the enclosing function does not return the resource to callers.
- If a function returns the resource, callee `.Close()` insertion is flagged as an invalid reproducer resolution:  
  `Defect reproducer check failed: function returns resource to caller; callee cleanup violates interface contract`.

---

## 3. Empirical Verification of Hardened Oracles

Dedicated unit tests in [`go-backend/agent/repo_intel/execution/oracles_test.go`](file:///E:/SWE/go-backend/agent/repo_intel/execution/oracles_test.go) empirically verified the rejection of defective patches:

```text
=== RUN   TestHardenedOraclesRejectFlawedNAT03Patch
    oracles_test.go:81: CONFIRMED: OracleNegativeProperties correctly rejected flawed NAT-03 patch: Negative property check failed: candidate patch introduces callee cleanup on escaping resource "resp" returned to caller; violates ownership transfer and panics on error returns
    oracles_test.go:91: CONFIRMED: OracleDefectReproducer correctly rejected flawed NAT-03 patch: Defect reproducer check failed: function returns resource to caller; callee cleanup violates interface contract
--- PASS: TestHardenedOraclesRejectFlawedNAT03Patch (0.46s)
=== RUN   TestHardenedOraclesRejectPreErrorCheckCleanup
    oracles_test.go:155: CONFIRMED: OracleNegativeProperties correctly rejected pre-error-check placement: Negative property check failed: resource cleanup for "f" placed prior to error check; triggers nil pointer dereference panic when allocation fails
--- PASS: TestHardenedOraclesRejectPreErrorCheckCleanup (0.40s)
=== RUN   TestHardenedOraclesRejectMultipleCloseCalls
    oracles_test.go:219: CONFIRMED: OracleNegativeProperties correctly rejected redundant close: Negative property check failed: multiple Close() calls detected on resource "f" without reassignment
--- PASS: TestHardenedOraclesRejectMultipleCloseCalls (0.41s)
PASS
```

### Key Result
When the exact candidate patch generated in Milestone 3.7 Phase 2 for NAT-03 is presented to the hardened suite:
- **Oracle 5 (Negative Properties):** **FAILS** (Escaping resource violation)
- **Oracle 3 (Defect Reproducer):** **FAILS** (Callee cleanup on returned resource violates contract)
- **Overall Verdict:** **REJECTED**

The hardened independent oracles now provide a robust secondary safety boundary even in the hypothetical event that a synthesizer proposed an unsound repair.

