# Disposal and persistence contract

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

Supported operation: DISCARD_SANDBOX. It deletes only a registered execution's immediate `mvp-sandbox-*` child under the managed temporary root. Unknown IDs, missing paths, paths outside that root, unsafe managed roots, and active proposal executions are rejected. The API returns `operation: DISCARD_SANDBOX` and `revision_restored: false` through the compatibility `/rollback` route. The UI describes verified deletion and preserves failure details.

Deletion errors propagate, and a successful remove call is not enough: Lstat must confirm absence. The audit-linked execution status becomes PASSED only after that check; failures remain FAILED and retryable. Repeated disposal of an already absent owned directory is safe. Tests cover actual deletion, repeated deletion, unmanaged-path protection, active-execution exclusion, injected deletion denial, injected no-op deletion, and original source digest preservation. Injection tests test control flow; actual filesystem deletion tests are separate.

Baseline workspaces have their own synchronous deletion/absence checks before candidate execution can finish. All Docker command records must include VERIFIED_REMOVAL container evidence to count as proof.

No persisted-revision restoration, versioned checkpoint, or filesystem rollback is implemented or claimed. Source repositories are never modified. In-memory audit records and proposal/execution state do not survive service restart. Only consumed approval nonces have a persistence contract. Restart tests verify those nonces remain consumed, while missing outstanding authority/proposal state is rejected. The audit export now returns a detached snapshot and computes source immutability from recorded executions rather than a constant true value.

Prior v2 disposal work is preserved, with the active-execution and deletion-failure protections added here. Public traces include post-disposal audit status.
