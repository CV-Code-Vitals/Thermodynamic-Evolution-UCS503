# =============================================================================
# scripts/run-demo.ps1
#
# Deterministic Autonomous Code Repair MVP Demonstration
# Reproduces:
#   1. Autonomous AST Repair (INJ-EXT2-RES-02) in Authentic External Repository
#   2. Isolated Sandbox Execution, Cryptographic Authorization & Replay Protection
#   3. Multi-Stage Independent 5-Oracle Verification
#   4. Disqualified / Escaping Return Safe Abstention (NAT-03)
#   5. Conservative Command Injection Architectural Abstention (INJ-EXT2-CMD-01)
#   6. Bitwise Original Source Immutability
#   7. Full Machine-Readable Audit Report
# =============================================================================

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path -Path (Join-Path $scriptDir "..")

$env:TEMP = Join-Path $repoRoot ".tmp"
$env:TMP = Join-Path $repoRoot ".tmp"
$env:GOTMPDIR = Join-Path $repoRoot ".tmp"

if (-not (Test-Path $env:TEMP)) {
    New-Item -ItemType Directory -Path $env:TEMP -Force | Out-Null
}

$backendDir = Join-Path $repoRoot "go-backend"
$binaryPath = Join-Path $backendDir "bin\remediation_mvp.exe"

Write-Host ">>> Building Autonomous Repair MVP Release Candidate..." -ForegroundColor Cyan
Push-Location $backendDir
try {
    go build -o bin\remediation_mvp.exe .\cmd\remediation_mvp
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to compile remediation_mvp binary"
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

Write-Host ">>> Executing Deterministic Demonstration..." -ForegroundColor Green
& $binaryPath --demo --workspace "$repoRoot" --output (Join-Path $repoRoot "reports\milestone3_8_demo_audit.json")
