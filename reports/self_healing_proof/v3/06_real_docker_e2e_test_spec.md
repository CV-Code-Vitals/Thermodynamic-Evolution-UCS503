# Implemented opt-in Docker proof

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

Implementation: `go-backend/healproof_contract_test.go`, `TestHealproofDockerPublicLifecycle`. Status: IMPLEMENTED, COMPILED, NOT RUN WITH DOCKER.

Prerequisites on an explicitly authorized Docker-capable machine:

1. Existing clean checkout of https://github.com/Zylus08/mlsc-timetable at `cbd082eafd24b290391145cb0399e81c6b41303c`. The test verifies HEAD, exact remote, clean status and original digest. It does not clone or fetch.
2. Prebuilt trusted Rust AST scanner supplied via ENGINE_PATH.
3. Actual Docker daemon and the project's hardened sandbox image, default `ucs503-sandbox-runner:latest` (SANDBOX_RUNNER_IMAGE may select the operator-built image). The existing Dockerfile pins the Go base image digest. Network access stays disabled during repository checks; the fixture is standard-library-only in an approved nested module.
4. RUN_HEALPROOF_DOCKER=1, HEALPROOF_PINNED_REPO pointing to the pinned checkout, and optionally HEALPROOF_EVIDENCE_DIR pointing to a new proof output directory.

Run from go-backend:

```text
go test . -run '^TestHealproofDockerPublicLifecycle$' -count=1 -v -timeout 10m
```

The test creates an immutable archive of a disposable copy plus the controlled fixture. It calls real authenticated `/scan`, `/targets`, `/analyze`, `/propose`, `/approve`, `/execute`, `/audit`, `/rollback`, then `/audit` again. No target ID, defect location, or repair is submitted to the positive synthesizer. The target is selected from discovered source findings. Out-of-band trusted plan provisioning is the same operator API used by production; authorization is still issued and consumed through REST.

Positive case requires INDEPENDENTLY_VERIFIED, the before-fail/after-pass lifecycle, regression/behavior/negative checks, unchanged original source and verified disposal. Negative case introduces a deliberate test-only synthesis collaborator: it delegates real synthesis, then substitutes a syntactically valid comment-only patch before manifest binding and approval. The defect remains; the same independent postcondition must fail, and INDEPENDENTLY_VERIFIED is forbidden. This is fault injection, not a claimed autonomous repair. Both cases use the real runtime and public request sequence.

Public trace artifacts include plan, generated diff, modified source, signed approval, baseline output/evidence, every oracle invocation, final verdict, source/candidate digests, and post-disposal audit state. When opted in, unavailable prerequisites cause failure rather than a silent skip. The absent opt-in flag causes an explicit NOT RUN skip.

This dedicated test is ready for a controlled existing Docker-capable CI job after separate operator authorization. No workflow file was changed, no remote push was made, and no CI workflow was triggered.
