#!/usr/bin/env bash
# Run API and frontend concurrently (Unix/macOS)
# For local development only. Uses ADMIN_PASSKEY=localdevtoken.
set -e
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Require an explicit local-only passkey; do not fall back to a hardcoded value.
if [ -z "$ADMIN_PASSKEY" ]; then
  echo "ADMIN_PASSKEY must be set before starting the API. Example: export ADMIN_PASSKEY='your-local-only-passkey'" >&2
  exit 1
fi

# Start the static dashboard server so links to /admin-portal/ and
# /graph-visualizer/ work from the root dashboard.
echo "Starting dashboard server on port 8000..."
(cd "$ROOT_DIR" && PORT=8000 BACKEND_PORT=8080 node serve.js) &
STATIC_PID=$!

# Start API in background
echo "Starting API on port 8080..."
APP_ENV=development ADMIN_PASSKEY="$ADMIN_PASSKEY" CORS_ALLOWED_ORIGINS="http://localhost:5173,http://127.0.0.1:5173,http://localhost:8000,http://127.0.0.1:8000" PORT=8080 bash -c "cd \"$ROOT_DIR/api\" && go run ." &
API_PID=$!

echo "Starting frontend (Vite) on port 5173..."
cd "$ROOT_DIR/admin-portal"
if [ ! -d node_modules ]; then
  npm install
fi
npm run dev -- --host 0.0.0.0 --port 5173 &
FE_PID=$!

echo "Dashboard PID: $STATIC_PID, API PID: $API_PID, Frontend PID: $FE_PID"
wait $STATIC_PID $API_PID $FE_PID
