# UCS503 Sandbox Runner Image

This directory contains the Docker configuration for the unprivileged sandbox runner used by the agentic annealing pipeline.

## Security Controls Enforced

1. **Non-Root Execution:** Runs as unprivileged user `runner` (`UID 10001:10001`).
2. **Network Isolation:** Executed with `--network none` in `DockerEvaluator` to block outbound connections, C2 callbacks, and AWS IMDS metadata exfiltration (`169.254.169.254`).
3. **Read-Only Root Filesystem:** Executed with `--read-only` preventing system package modification.
4. **Dropped Capabilities:** Executed with `--cap-drop ALL` and `--security-opt no-new-privileges`.
5. **Strict Resource Quotas:** Enforces memory caps (`--memory 512m`), CPU limits (`--cpus 1.0`), process limits (`--pids-limit 128`), and bounded execution timeouts (15s default).
6. **Ephemeral Workspaces:** Binds only a temporary, sanitized workspace snapshot (`os.MkdirTemp`). The host repository root and credentials (`.git`, `.env*`) are filtered and never mounted.
7. **CGO Disabled:** `CGO_ENABLED=0` prevents execution of arbitrary C code or unsafe shared libraries.

## How to Build the Image

From the repository root:

```bash
docker build -t ucs503-sandbox-runner:latest -f docker/sandbox-runner/Dockerfile .
```

## How to Test

Verify the runner executes in an offline container:

```bash
docker run --rm --network none --read-only --tmpfs /tmp:rw,noexec,nosuid,size=64m ucs503-sandbox-runner:latest go version
```

## Maintenance & Base Image Updates

The base image is pinned to `golang:1.25.0-alpine3.21`. When updating the Go minor version or Alpine patch version, rebuild the container and verify that `CGO_ENABLED=0` and non-root UID 10001 permissions are preserved.
