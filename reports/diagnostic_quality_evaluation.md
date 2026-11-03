# Diagnostic Quality Evaluation Report

**Evaluation Date:** 2026-11-03T09:01:03Z  
**Evaluated Fixtures:** 12 controlled test cases with known ground truth  

## Evaluation Metrics Summary

| Metric | Value | Target Threshold | Status |
|---|---|---|---|
| **Precision** | 100.00% | >= 80.0% | PASS |
| **Recall** | 100.00% | >= 80.0% | PASS |
| **F1-Score** | 1.00 | >= 0.80 | PASS |
| **Evidence Accuracy** | 100.0% | >= 90.0% | PASS |
| **Deduplication Accuracy** | 100.0% | 100.0% | PASS |

## Fixture Test Results Matrix

| Fixture Name | Expected Validation | Actual Validation | Expected Readiness | Actual Readiness | Confidence | Severity | Result |
|---|---|---|---|---|---|---|---|
| 1. Confirmed Architectural Cycle | `VALIDATED` | `VALIDATED` | `HUMAN_APPROVAL_REQUIRED` | `HUMAN_APPROVAL_REQUIRED` | `0.98` | `HIGH` | **PASS** |
| 2. Confirmed High-Complexity Monolithic Function | `VALIDATED` | `VALIDATED` | `AUTOMATION_WITH_REVIEW` | `AUTOMATION_WITH_REVIEW` | `0.95` | `MEDIUM` | **PASS** |
| 3. Genuine Command Injection Risk | `REQUIRES_HUMAN_REVIEW` | `REQUIRES_HUMAN_REVIEW` | `INVESTIGATION_REQUIRED` | `INVESTIGATION_REQUIRED` | `0.45` | `HIGH` | **PASS** |
| 4. Confirmed Hardcoded Secret Credential | `REQUIRES_HUMAN_REVIEW` | `REQUIRES_HUMAN_REVIEW` | `HUMAN_APPROVAL_REQUIRED` | `HUMAN_APPROVAL_REQUIRED` | `0.55` | `CRITICAL` | **PASS** |
| 5. Confirmed Excessive Nesting Depth | `VALIDATED` | `VALIDATED` | `AUTOMATION_WITH_REVIEW` | `AUTOMATION_WITH_REVIEW` | `0.95` | `MEDIUM` | **PASS** |
| 6. Confirmed Missing Test Coverage | `REQUIRES_HUMAN_REVIEW` | `REQUIRES_HUMAN_REVIEW` | `INVESTIGATION_REQUIRED` | `INVESTIGATION_REQUIRED` | `0.60` | `LOW` | **PASS** |
| 7. Stale Finding with Out-of-Bounds Line Number | `STALE` | `STALE` | `INFORMATIONAL_ONLY` | `INFORMATIONAL_ONLY` | `0.15` | `LOW` | **PASS** |
| 8. Stale Finding on Non-Existent Deleted File | `STALE` | `STALE` | `INFORMATIONAL_ONLY` | `INFORMATIONAL_ONLY` | `0.15` | `LOW` | **PASS** |
| 9. Malformed / Invalid Finding Metadata | `INVALID` | `INVALID` | `INFORMATIONAL_ONLY` | `INFORMATIONAL_ONLY` | `0.05` | `LOW` | **PASS** |
| 10. Path Traversal Attempt Escaping Root | `INVALID` | `INVALID` | `INFORMATIONAL_ONLY` | `INFORMATIONAL_ONLY` | `0.05` | `LOW` | **PASS** |
| 11. Primary Finding for Deduplication Check | `REQUIRES_HUMAN_REVIEW` | `REQUIRES_HUMAN_REVIEW` | `INVESTIGATION_REQUIRED` | `INVESTIGATION_REQUIRED` | `0.45` | `HIGH` | **PASS** |
| 12. Duplicate Finding for Deduplication Check | `REQUIRES_HUMAN_REVIEW` | `REQUIRES_HUMAN_REVIEW` | `INVESTIGATION_REQUIRED` | `INVESTIGATION_REQUIRED` | `0.45` | `HIGH` | **PASS** |

## Deduplication & Correlation Telemetry

- **Total correlations generated:** 2
- **[DUPLICATE] CORR-DUP-0001**: Primary finding `FIXTURE-RI-STR-HIGH-COMPLEXITY-SAMPLE-L21` linked with FIXTURE-RI-STR-HIGH-NESTING-SAMPLE-L21 (*Multiple findings targeting identical file location and category*)
- **[DUPLICATE] CORR-DUP-0002**: Primary finding `FIXTURE-RI-SEC-CMD-INJECT-SAMPLE-L10` linked with FIXTURE-RI-SEC-CMD-INJECT-DUP-PRIMARY-L10, FIXTURE-RI-SEC-CMD-INJECT-DUP-SECONDARY-L10 (*Multiple findings targeting identical file location and category*)

