# Remaining blockers

Implementation commit: `5d1f97c4637e846a39b9cd776b9dcc2fd3574d0b`. Branch: `codex/healproof-preservation`. Date: 2026-11-25 (Asia/Calcutta).

No known source-level blocker remains for Gate 2's implemented, tested control flow.

Dynamic proof remains blocked by the already established local runtime environment: Windows Docker Desktop executable is absent at `C:/Program Files/Docker/Docker/Docker Desktop.exe`; Ubuntu WSL2 `/usr/bin/docker` points to the missing `/mnt/wsl/docker-desktop/cli-tools/usr/bin/docker`; the Docker service is inactive. No project image or container execution was verified. See `../v2/docker_runtime_proof.md`; v3 did not repeat a full environment/system audit.

An actual supported daemon, working WSL integration or an authorized Docker-capable runner, the hardened image, prebuilt AST scanner and clean pinned source checkout are required for the opt-in experiment. This is a dynamic-evidence blocker, not an excuse to omit source implementation. No remote authorization was requested because source changes, test harness and local evidence are complete and reviewable.

General repositories also require independently reviewed test plans. No plan is inferred from a patch or trusted just because repository metadata names a command. Lack of a reviewed matching plan produces REPRODUCER_NOT_AVAILABLE at live approval.

Limitations: audit/proposal state is process-local, rollback is deletion only, the proof fixture is controlled and additive, and local runtime doubles demonstrate wiring only. Docker-specific mode/readability and cold-cache runtime behavior remain subject to the opt-in test; no result is assumed. Historical benchmarks remain separate.
