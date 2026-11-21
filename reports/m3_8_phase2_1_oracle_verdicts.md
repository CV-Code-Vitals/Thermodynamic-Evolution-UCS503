# Milestone 3.8 Phase 2.1 — Honest Oracle Outcomes & Verdict Contract

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Status:** COMPLETE & ALIGNED

---

## 1. Principle of Honest Reporting

In prior implementations, oracle checks that did not execute (e.g. regression suites omitted from candidate task definitions, or behavioral invariance scripts not configured) defaulted to `Passed: true`. In environments where Docker was unavailable, dynamic checks were either silently bypassed or simulated on the host, conflating unverified candidates with independently verified candidates.

Milestone 3.8 Phase 2.1 enforces the **Honest Oracle Contract**:
1. Every oracle check reports an explicit, discrete execution status.
2. `NOT_RUN` is never equated with `PASSED`.
3. An unexecuted check represents an absence of dynamic evidence, not proof of correctness.

---

## 2. Oracle Status Taxonomy

Each of the five independent verification oracles reports one of four explicit states:

```go
type OracleStatus string

const (
    OracleStatusPassed      OracleStatus = "PASSED"
    OracleStatusFailed      OracleStatus = "FAILED"
    OracleStatusNotRun      OracleStatus = "NOT_RUN"
    OracleStatusUnavailable OracleStatus = "UNAVAILABLE"
)
```

| Oracle Status | Definition | Criteria |
| :--- | :--- | :--- |
| `PASSED` | The verification check executed and confirmed patch correctness. | AST parse valid, zero negative properties, or container dynamic test exited with code 0. |
| `FAILED` | An executed check detected a defect, syntax error, regression, or negative property violation. | AST parse error, container test exit code != 0, forbidden pattern (e.g. shadowed error, nil dereference) detected. |
| `NOT_RUN` | The verification check was not executed because no validation command was defined for the candidate. | Task definition lacks `ValidationCommands`, `reproducerCmd`, or `invarianceCmd`. |
| `UNAVAILABLE` | Dynamic verification was requested but could not execute due to lack of container isolation. | Docker daemon offline, runner not configured, or environment lacks container boundary. **Host fallback strictly prohibited.** |

---

## 3. Execution Verdict Hierarchy

The overall repair verdict computed by `MVPService.ExecuteRepair` reflects the true depth of verification achieved:

```
+-----------------------------------------------------------------------------------+
|                            OVERALL VERIFICATION VERDICT                          |
|                                                                                   |
|  [Any check FAILED?] -------------------------> FAILED                            |
|                                                                                   |
|  [Synthesis Abstained / Refused?] ------------> ABSTAINED                         |
|                                                                                   |
|  [Dynamic Reproducer PASSED in Container?] ---> INDEPENDENTLY_VERIFIED            |
|                                                                                   |
|  [Static checks PASSED, dynamic UNAVAILABLE?] -> SYNTAX_ONLY_VERIFIED             |
+-----------------------------------------------------------------------------------+
```

### Contract Specifications:
1. **`INDEPENDENTLY_VERIFIED`**:
   - Requires dynamic defect reproducer execution inside a hardened container runner (`DockerContainerRunner`).
   - Relevant regression test suite passes in container.
   - Behavioral invariance checks pass in container.
   - Negative-property AST inspection confirms zero forbidden patterns introduced.
   - Confirmed by `reproOutcome.Status == "PASSED" && reproOutcome.ContainerInfo != nil`.
2. **`SYNTAX_ONLY_VERIFIED`**:
   - Static AST parsing and negative property checks succeeded on the trusted host process.
   - Dynamic container verification was unavailable, skipped, or not configured.
   - Transparently indicates to the operator that runtime behavior has not been dynamically validated.
3. **`FAILED`**:
   - At least one executed static or dynamic check failed.
   - The patch is rolled back immediately inside the sandbox; the original repository remains pristine.
4. **`ABSTAINED`**:
   - Synthesis identified high ambiguity, escaping resources, loops, or complex multi-file dependencies.
   - Safely refuses patch generation with an explicit abstention rationale.

---

## 4. UI Representation in Remediation Portal

The administrative frontend (`admin-portal/src/components/RemediationPortal.jsx` & `RemediationPortal.css`) visually separates the four states:

- **`✓ PASSED`**: Styled in high-contrast emerald green (`.oracle-status.passed`).
- **`✗ FAILED`**: Styled in high-contrast ruby red (`.oracle-status.failed`).
- **`○ NOT RUN`**: Styled in neutral slate gray (`.oracle-status.not-run`) with a clear label explaining that no test was configured.
- **`⚠ UNAVAILABLE`**: Styled in amber warning tones (`.oracle-status.unavailable`), explicitly noting that dynamic container isolation was unavailable.
- **Container Telemetry Badge**: When `container_info` is present, the card renders an interactive telemetry chip detailing container name, memory limits, CPU quotas, and isolation policies.

---

## 5. Verification Matrix

| Verification Aspect | Test Suite Evidence | Result |
| :--- | :--- | :--- |
| Explicit `UNAVAILABLE` status without container | `TestAdversarial_1_HostSentinelFileAccessPrevented` | PASSED |
| Explicit `NOT_RUN` status on missing commands | `TestAutonomousPatchSynthesisAndIndependentOracles` | PASSED |
| `SYNTAX_ONLY_VERIFIED` verdict on container absence | `TestScanToRepairWorkflow_Success` | PASSED |
| `FAILED` verdict on broken compilation / AST error | `TestAdversarialSafetyScenarios/Scenario4_CompilerSyntaxErrorTriggersRollback` | PASSED |
| `FAILED` verdict on flawed NAT-03 patch | `TestHardenedOraclesRejectFlawedNAT03Patch` | PASSED |
| `FAILED` verdict on pre-error-check cleanup placement | `TestHardenedOraclesRejectPreErrorCheckCleanup` | PASSED |
