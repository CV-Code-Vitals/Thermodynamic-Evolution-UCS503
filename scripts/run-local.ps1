# Run both API and frontend for local development on Windows (PowerShell)
# Launches two new PowerShell windows (one for API, one for frontend).
# This script requires ADMIN_PASSKEY to be set in the environment before running.

# Determine repository root robustly (script lives in scripts/)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path -Path (Join-Path $scriptDir "..")

# Use a predictable local-only credential. Production still requires the
# ADMIN_PASSKEY environment variable or deployment secret.
if (-not $env:ADMIN_PASSKEY -or $env:ADMIN_PASSKEY -eq '') {
    $env:ADMIN_PASSKEY = 'admin'
    Write-Host "ADMIN_PASSKEY was not set; using local development passkey 'admin'." -ForegroundColor Yellow
}

# Paths
$apiPath = Join-Path $repoRoot 'api'
$fePath = Join-Path $repoRoot 'admin-portal'

function Test-PortFree($port) {
    return -not (Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue)
}

$apiPort = 8080
while (-not (Test-PortFree $apiPort)) {
    $apiPort++
}

if (-not (Test-Path $apiPath)) {
    Write-Host "API path not found: $apiPath" -ForegroundColor Red
    exit 1
}
if (-not (Test-Path $fePath)) {
    Write-Host "Frontend path not found: $fePath" -ForegroundColor Red
    exit 1
}

# Build the static portal for the local reverse proxy used by serve.js.
Push-Location -Path $fePath
$env:VITE_API_BASE = '/api'
if (-not (Test-Path (Join-Path $fePath 'node_modules'))) {
    npm ci
    if ($LASTEXITCODE -ne 0) {
        Pop-Location
        Write-Host "Unable to install frontend dependencies." -ForegroundColor Red
        exit 1
    }
}
npm run build
$buildExitCode = $LASTEXITCODE
Pop-Location
if ($buildExitCode -ne 0) {
    Write-Host "Unable to build the local Admin Portal." -ForegroundColor Red
    exit 1
}

# Start the static dashboard server so links to /admin-portal/ and
# /graph-visualizer/ work from the root dashboard.
$existingStatic = Get-NetTCPConnection -LocalPort 8000 -State Listen -ErrorAction SilentlyContinue
if ($existingStatic) {
    $staticProcess = Get-CimInstance Win32_Process -Filter "ProcessId=$($existingStatic.OwningProcess)" -ErrorAction SilentlyContinue
    if ($staticProcess.CommandLine -match 'serve\.js') {
        Stop-Process -Id $existingStatic.OwningProcess -Force
        $existingStatic = $null
    }
}
if (-not $existingStatic) {
    $staticCmd = "Set-Location -Path '$repoRoot'; `$env:PORT='8000'; `$env:BACKEND_PORT='$apiPort'; node serve.js"
    Start-Process powershell -ArgumentList "-NoExit","-Command",$staticCmd
} else {
    Write-Host 'Port 8000 is occupied by another application; stop it before using the dashboard.' -ForegroundColor Red
}

# Start API in new window (child process will inherit ADMIN_PASSKEY from environment)
if (Test-PortFree $apiPort) {
    $apiCmd = "Set-Location -Path '$apiPath'; `$env:PORT='$apiPort'; `$env:LOCAL_DEV='true'; go run main.go"
    Start-Process powershell -ArgumentList "-NoExit","-Command",$apiCmd
} else {
    Write-Host "Unable to start API on selected port $apiPort." -ForegroundColor Red
}

# Start frontend in new window
$feTarget = "http://localhost:$apiPort"
if (Test-PortFree 5173) {
    $feCmd = "Set-Location -Path '$fePath'; `$env:VITE_DEV_API_TARGET='$feTarget'; if (-not (Test-Path 'node_modules')) { npm install }; npm run dev"
    Start-Process powershell -ArgumentList "-NoExit","-Command",$feCmd
} else {
    Write-Host 'Frontend already running on port 5173; reusing it.' -ForegroundColor Yellow
}

Write-Host 'Started or reused local project services.' -ForegroundColor Green
Write-Host "API: http://localhost:$apiPort/status" -ForegroundColor Cyan
Write-Host "Frontend: http://localhost:5173" -ForegroundColor Cyan
Write-Host "Dashboard: http://localhost:8000" -ForegroundColor Cyan
