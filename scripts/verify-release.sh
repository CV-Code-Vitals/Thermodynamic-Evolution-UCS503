#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
: "${VITE_API_BASE:?Set the production HTTPS API URL including /api}"
docker info >/dev/null
(cd "$ROOT_DIR/thermodynamic-ast-engine" && cargo test --locked && cargo build --release --locked)
docker build -t ucs503-sandbox-runner:latest -f "$ROOT_DIR/docker/sandbox-runner/Dockerfile" "$ROOT_DIR"
(cd "$ROOT_DIR/analysis-authorizer" && go test ./... && go vet ./... && go build -o /dev/null .)
(cd "$ROOT_DIR/source-stager" && go test ./... && go vet ./... && go build -o /dev/null .)
(cd "$ROOT_DIR/infra/analysis/stream-proxy" && npm ci && npm test)
(cd "$ROOT_DIR" && node --test scripts/compose-config.test.mjs tests/*.test.mjs)
(cd "$ROOT_DIR/api" && go test ./... && go vet ./... && go build -o /dev/null .)
export ENGINE_PATH="$ROOT_DIR/thermodynamic-ast-engine/target/release/thermodynamic-ast-engine"
export SANDBOX_RUNNER_IMAGE=ucs503-sandbox-runner:latest
export SANDBOX_WORKSPACE_DIR="${SANDBOX_WORKSPACE_DIR:-$ROOT_DIR/go-backend/uploads/sandboxes}"
(cd "$ROOT_DIR/go-backend" && go test ./... && go vet ./... && go build -o /dev/null .)
(cd "$ROOT_DIR/admin-portal" && npm ci && npm test && npm run lint && npm run build && npm audit --audit-level=high)
git -C "$ROOT_DIR" diff --check
printf 'Local release checks passed. Run scripts/smoke-portal.mjs against the deployed API before release.\n'
