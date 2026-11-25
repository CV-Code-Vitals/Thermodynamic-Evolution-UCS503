# Public executor contract

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

GATE 2 CODE COMPLETE at the source/control-flow level. Real container proof is not claimed.

The live route is scan -> retained workspace -> semantic targets -> analysis -> automatic synthesis -> operator-provisioned verification plan -> authenticated approval -> single-use execution -> baseline and candidate oracles -> verdict -> audit -> DISCARD_SANDBOX.

`go-backend/snapshot_scanner.go` now retains a managed source copy when production scan requests retention. `main.go` passes the same workspace store to scan and remediation. Previously, staged scan returned an ID after deleting its extraction; remediation could not resolve it. `mvp_scanner.go` uses the actual local checkout metadata only when that checkout owns `.git`; extracted snapshots use a content SHA identity and revision, never an enclosing repository's commit.

`MVPService.ConfigureTrustedVerificationPlans` is operator/bootstrap configuration. `HEALPROOF_TRUSTED_PLANS` can point to a host-managed JSON array of reviewed plan templates. No endpoint accepts commands or plans from request bodies or repository reports. Templates match repository identity, immutable revision, entire source tree digest, and target file hash. Propose fills only workspace/snapshot identity and the autonomously generated patch/candidate digests, returning the full plan and digest for review. No matching template means no trustworthy reproducer.

The live `/approve` route calls `ApproveVerifiedRepair`: missing or invalid plans are rejected with HTTP 422, an explicit verdict, and an audit entry. The older in-process `ApproveRepair` remains for the explicitly labeled legacy demo; a planless demo token cannot apply a patch or pass execution. Demo acknowledgment is not live execution authorization.

`/execute` rejects source/patch/plan/snapshot/command/workspace tampering, replay and unavailable runtime. Error responses retain the execution ID and verdict so the UI can display the failure and offer sandbox disposal. Every execution error is audited. Public tests use the actual authentication middleware (`X-API-Token`); identity is not accepted from JSON.

All new execution paths require every mandatory oracle. Historical SYNTAX_ONLY_VERIFIED remains a displayable reduced classification; it is never elevated into independent proof and is not emitted as success by this mandatory-plan path. Missing/not-run checks fail closed.

Changed files (relative to initial session head):

- `admin-portal/src/components/RemediationPortal.jsx`
- `admin-portal/src/remediation-api.js`
- `admin-portal/tests/remediation-api.test.js`
- `go-backend/agent/repo_intel/execution/mvp_scanner.go`
- `go-backend/agent/repo_intel/execution/mvp_service.go`
- `go-backend/agent/repo_intel/execution/scan_repair_test.go`
- `go-backend/agent/repo_intel/execution/verification_plan.go`
- `go-backend/agent/repo_intel/execution/verification_plan_test.go`
- `go-backend/healproof_contract_test.go`
- `go-backend/main.go`
- `go-backend/remediation_routes.go`
- `go-backend/remediation_routes_test.go`
- `go-backend/scan_to_repair_e2e_test.go`
- `go-backend/snapshot_scanner.go`
- `go-backend/testdata/healproof/go.mod`
- `go-backend/testdata/healproof/independent_test.go`
- `go-backend/testdata/healproof/reader.go`

See `source-changes.patch` for the complete source delta, including the session checkpoint already preserved in d0e560a. Existing tests without plans now assert public approval rejection; the new public harness covers their downstream stages with trusted plans.
