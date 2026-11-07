# =============================================================================
# scripts/run-demo.ps1
#
# Deterministic Autonomous Code Repair MVP Demonstration
# Reproduces:
#   1. Natural Defect (NAT-03) Detection, Genuine AST Synthesis & Approval
#   2. Isolated Sandbox Execution & Replay Protection
#   3. Independent 5-Oracle Verification
#   4. Conservative Abstention (INJ-EXT2-CMD-01)
#   5. Bitwise Original Source Immutability
#   6. Full Machine-Readable Audit Report
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
