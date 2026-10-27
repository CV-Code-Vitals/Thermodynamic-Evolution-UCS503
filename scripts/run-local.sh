#!/usr/bin/env bash
# Run API and frontend concurrently (Unix/macOS)
# For local development only.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Require an explicit local-only passkey; do not fall back to a hardcoded value.
if [ -z "${ADMIN_PASSKEY:-}" ]; then
  echo "ADMIN_PASSKEY must be set before starting the API. Example: export ADMIN_PASSKEY='your-local-only-passkey'" >&2
  exit 1
fi

export ANALYSIS_API_TOKEN="${ANALYSIS_API_TOKEN:-$(node -e "process.stdout.write(require('crypto').randomBytes(32).toString('hex'))")}"
export ANALYSIS_SERVICE_URL="${ANALYSIS_SERVICE_URL:-http://127.0.0.1:8081}"
export SANDBOX_WORKSPACE_DIR="${SANDBOX_WORKSPACE_DIR:-$ROOT_DIR/go-backend/uploads/sandboxes}"
(cd "$ROOT_DIR/thermodynamic-ast-engine" && cargo build --release --locked)
(cd "$ROOT_DIR/admin-portal" && npm ci && VITE_API_BASE=/api VITE_ALLOW_RELATIVE_API_BASE=true npm run build)

LOCAL_BIN_DIR="$(mktemp -d)"
(cd "$ROOT_DIR/api" && go build -o "$LOCAL_BIN_DIR/archive-api" .)
(cd "$ROOT_DIR/go-backend" && go build -o "$LOCAL_BIN_DIR/analysis-api" .)
SERVICE_PIDS=()
cleanup() {
  for service_pid in "${SERVICE_PIDS[@]}"; do
    kill "$service_pid" 2>/dev/null || true
  done
  rm -rf "$LOCAL_BIN_DIR"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

echo "Starting analysis service on port 8081..."
(cd "$ROOT_DIR/go-backend" && PORT=8081 exec "$LOCAL_BIN_DIR/analysis-api") &
SERVICE_PIDS+=("$!")

# Start the static dashboard server so links to /admin-portal/ and
# /graph-visualizer/ work from the root dashboard.
echo "Starting dashboard server on port 8000..."
(cd "$ROOT_DIR" && PORT=8000 BACKEND_PORT=8080 node serve.js) &
STATIC_PID=$!
SERVICE_PIDS+=("$STATIC_PID")

# Start API in background
echo "Starting API on port 8080..."
(cd "$ROOT_DIR/api" && APP_ENV=development CORS_ALLOWED_ORIGINS="http://localhost:5173,http://127.0.0.1:5173,http://localhost:8000,http://127.0.0.1:8000" PORT=8080 exec "$LOCAL_BIN_DIR/archive-api") &
API_PID=$!
SERVICE_PIDS+=("$API_PID")

echo "Starting frontend (Vite) on port 5173..."
cd "$ROOT_DIR/admin-portal"
if [ ! -d node_modules ]; then
  npm install
fi
npm run dev -- --host 0.0.0.0 --port 5173 &
FE_PID=$!
SERVICE_PIDS+=("$FE_PID")

echo "Dashboard PID: $STATIC_PID, API PID: $API_PID, Frontend PID: $FE_PID"
wait
