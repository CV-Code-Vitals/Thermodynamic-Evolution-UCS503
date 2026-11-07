#!/usr/bin/env bash
# =============================================================================
# scripts/run-demo.sh
#
# Deterministic Autonomous Code Repair MVP Demonstration (Linux/macOS)
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

export TEMP="${REPO_ROOT}/.tmp"
export TMP="${REPO_ROOT}/.tmp"
export GOTMPDIR="${REPO_ROOT}/.tmp"
mkdir -p "${TEMP}"

BACKEND_DIR="${REPO_ROOT}/go-backend"
BINARY_PATH="${BACKEND_DIR}/bin/remediation_mvp"

echo ">>> Building Autonomous Repair MVP Release Candidate..."
(cd "${BACKEND_DIR}" && go build -o bin/remediation_mvp ./cmd/remediation_mvp)

echo ">>> Executing Deterministic Demonstration..."
"${BINARY_PATH}" --demo --workspace "${REPO_ROOT}" --output "${REPO_ROOT}/reports/milestone3_8_demo_audit.json"
