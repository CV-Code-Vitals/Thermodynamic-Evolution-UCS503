# Source-level regression results

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

All final selected tests passed. No untrusted repository tests or dependency downloads ran on the host.

- `final-go-tests.log`: broad public REST, scan admission/extraction, verification contract, approval HMAC/ledger, containment, oracle-negative, Docker cleanup/staging control-flow and disposal regressions. Counts are in `test-results.json` (top-level and subtests counted separately).
- `final-contract-tests.log`: focused final rerun after tightening complete-proof evidence and recording post-disposal audit state.
- `frontend-all-tests.log`: all 57 frontend tests passed; zero failures.
- `frontend-tests.log`: 15 API-client tests passed, including preservation of HTTP 422 execution/verdict/disposal references.
- `frontend-build-final.log`: build passed with `VITE_API_BASE=https://healproof.invalid`, a local build placeholder. Nothing was deployed.
- `git diff --check`: passed.
- `TestHealproofDockerPublicLifecycle`: skipped because RUN_HEALPROOF_DOCKER is not enabled. This is NOT a dynamic pass.

The broad Go test command selects the root and execution packages and excludes historical benchmark/report writers and opt-in external execution suites. Tests use TMP/TEMP under `E:/SWE/.tmp/reality-testtemp` and GOCACHE under `E:/SWE/.tmp/gocache`. The final contract command is `go test . ./agent/repo_intel/execution -run 'TestHealproof|TestVerification|TestDiscard' -count=1 -v -timeout 120s`.

Public API coverage includes real middleware, retained `/scan` transport, live semantic discovery, automatic synthesis, reviewed manifest provisioning, signed approval, command propagation, mock failure/unavailability, audit, replay rejection, and verified disposal. The Rust report-producing process is a double in the non-Docker harness; live Go semantic target discovery is real. The opt-in Docker harness uses the actual prebuilt Rust engine.

TESTED_WITH_MOCK — NOT_DYNAMICALLY_VERIFIED applies to all injected runtime/engine outcomes. Files named `focused-tests*.log`, `contract-tests.log`, and `frontend-build.log` preserve intermediate development failures or earlier runs; the final logs above supersede them. Initial failures included harness authentication/archive setup, a test map initialization, and missing production build URL configuration; they were corrected and rerun.

Previously running repository revalidation is reconciled: `../v1/revalidation.log` ends PASS (12.103s). Its reproducible source findings (0/52/0/0) are historical source scan outcomes, not independently confirmed defects or new dynamic repairs. No historical benchmark artifacts were rewritten.
