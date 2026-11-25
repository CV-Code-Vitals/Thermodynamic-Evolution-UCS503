# Before/after defect proof

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

After approval scope verification and atomic nonce consumption, execution copies the unmodified proposal snapshot into a separate managed `mvp-baseline-*` directory. Its complete content digest must equal the approved source root. Independent test bytes are hash checked. Only disposable snapshot modes are normalized for UID/GID 10001; the original source is never chmod'ed or written.

The original reproducer runs in the hardened container. Required evidence: expected exit 1, the exact named Go test FAIL record, the independently defined defect assertion, no compilation/dependency/timeout failure, valid isolation evidence and verified container removal. A static AST finding cannot satisfy this step. The approved behavioral expectations must also pass against the baseline. Baseline content is rechecked and the baseline directory is deleted with errors propagated and absence verified.

Only then is the generated candidate written into the separate proposal sandbox. The same reproducer command and test bytes must pass with exit 0. Every approved regression test and behavioral check must actually run the named test; `PASS` with no matching test, a skipped test, missing result, or missing container/cleanup evidence fails. Syntax/vet and AST negative-property checks must pass. The candidate digest must remain stable during checks and the original repository pre/post content digests must match.

`INDEPENDENTLY_VERIFIED` additionally requires the actual DockerContainerRunner with its testing hooks disabled and the approved hardening limits. An injected runner cannot satisfy the gate even if it supplies apparently successful records. All command outputs, container evidence, baseline digest, candidate digest, plan digest and token consumption are recorded.

The independently authored fixture under `go-backend/testdata/healproof/` measures live Linux `/proc/self/fd` counts with GC disabled to prevent finalizers hiding the controlled leak. Separate tests check missing-file error semantics and unchanged byte reads, including empty and binary content. These tests were written without consulting the generated diff. They are a controlled fixture added to a real pinned repository, not a claim of a natural defect in that repository.

Current before/after outputs are control-flow doubles only: TESTED_WITH_MOCK — NOT_DYNAMICALLY_VERIFIED. See `source-digests-and-approval-index.json`, the three explicitly named `public-trace-real-false-*.json` records, and their generated patches. There are zero new real container oracle runs.
