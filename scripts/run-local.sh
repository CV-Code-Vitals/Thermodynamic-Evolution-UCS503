#!/usr/bin/env bash
# Run API and frontend concurrently (Unix/macOS)
# For local development only. Uses ADMIN_PASSKEY=localdevtoken.
set -e
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Use a predictable local-only credential. Production must provide a secret.
if [ -z "$ADMIN_PASSKEY" ]; then
  export ADMIN_PASSKEY=admin
  echo "ADMIN_PASSKEY was not set; using local development passkey 'admin'."
fi

# Start the static dashboard server so links to /admin-portal/ and
# /graph-visualizer/ work from the root dashboard.
echo "Starting dashboard server on port 8000..."
(cd "$ROOT_DIR" && PORT=8000 BACKEND_PORT=8080 node serve.js) &
STATIC_PID=$!

# Start API in background
echo "Starting API on port 8080..."
ADMIN_PASSKEY="$ADMIN_PASSKEY" LOCAL_DEV=true PORT=8080 bash -c "cd \"$ROOT_DIR/api\" && go run main.go" &
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
