# Run both API and frontend for local development on Windows (PowerShell)
# Launches two new PowerShell windows (one for API, one for frontend).
# This script requires ADMIN_PASSKEY to be set in the environment before running.

# Determine repository root robustly (script lives in scripts/)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$repoRoot = Resolve-Path -Path (Join-Path $scriptDir "..")

# Require an explicit local-only passkey; do not fall back to a hardcoded value.
if (-not $env:ADMIN_PASSKEY -or $env:ADMIN_PASSKEY -eq '') {
    Write-Host "ADMIN_PASSKEY must be set before starting the API. Example: `$env:ADMIN_PASSKEY = 'your-local-only-passkey'" -ForegroundColor Red
    exit 1
}

# Paths
$apiPath = Join-Path $repoRoot 'api'
$fePath = Join-Path $repoRoot 'admin-portal'
$analysisPath = Join-Path $repoRoot 'go-backend'
$analysisPort = 8081
if (-not $env:ANALYSIS_API_TOKEN) {
    $env:ANALYSIS_API_TOKEN = node -e "process.stdout.write(require('crypto').randomBytes(32).toString('hex'))"
}
$env:ANALYSIS_SERVICE_URL = "http://127.0.0.1:$analysisPort"
if (-not $env:SANDBOX_WORKSPACE_DIR) {
    $env:SANDBOX_WORKSPACE_DIR = Join-Path $analysisPath 'uploads/sandboxes'
}
Push-Location (Join-Path $repoRoot 'thermodynamic-ast-engine')
cargo build --release --locked
$engineBuildExitCode = $LASTEXITCODE
Pop-Location
if ($engineBuildExitCode -ne 0) { exit $engineBuildExitCode }

function Test-PortFree($port) {
    return -not (Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue)
}

$apiPort = 8080
while (-not (Test-PortFree $apiPort)) {
    $apiPort++
}
if ($apiPort -eq $analysisPort) {
    $apiPort++
    while (-not (Test-PortFree $apiPort)) { $apiPort++ }
}
if (-not (Test-PortFree $analysisPort)) {
    Write-Host "Analysis port $analysisPort is occupied; stop that service before starting." -ForegroundColor Red
    exit 1
}
$analysisCmd = "Set-Location -Path '$analysisPath'; `$env:PORT='$analysisPort'; go run ."
Start-Process powershell -ArgumentList "-NoExit","-Command",$analysisCmd

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
$originalViteApiBase = $env:VITE_API_BASE
$originalAllowRelativeApiBase = $env:VITE_ALLOW_RELATIVE_API_BASE
$env:VITE_API_BASE = '/api'
$env:VITE_ALLOW_RELATIVE_API_BASE = 'true'
if (-not (Test-Path (Join-Path $fePath 'node_modules'))) {
    npm ci
    if ($LASTEXITCODE -ne 0) {
        if ($null -eq $originalViteApiBase) { Remove-Item Env:VITE_API_BASE -ErrorAction SilentlyContinue } else { $env:VITE_API_BASE = $originalViteApiBase }
        if ($null -eq $originalAllowRelativeApiBase) { Remove-Item Env:VITE_ALLOW_RELATIVE_API_BASE -ErrorAction SilentlyContinue } else { $env:VITE_ALLOW_RELATIVE_API_BASE = $originalAllowRelativeApiBase }
        Pop-Location
        Write-Host "Unable to install frontend dependencies." -ForegroundColor Red
        exit 1
    }
}
npm run build
$buildExitCode = $LASTEXITCODE
if ($null -eq $originalViteApiBase) { Remove-Item Env:VITE_API_BASE -ErrorAction SilentlyContinue } else { $env:VITE_API_BASE = $originalViteApiBase }
if ($null -eq $originalAllowRelativeApiBase) { Remove-Item Env:VITE_ALLOW_RELATIVE_API_BASE -ErrorAction SilentlyContinue } else { $env:VITE_ALLOW_RELATIVE_API_BASE = $originalAllowRelativeApiBase }
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

# Start API in new window (child process will inherit local-only credentials)
if (Test-PortFree $apiPort) {
    $apiCmd = "Set-Location -Path '$apiPath'; `$env:APP_ENV='development'; `$env:CORS_ALLOWED_ORIGINS='http://localhost:5173,http://127.0.0.1:5173,http://localhost:8000,http://127.0.0.1:8000'; `$env:PORT='$apiPort'; go run ."
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
