# =============================================================================
# Prepyo Local Runner
# =============================================================================
# Starts both processes the app needs:
#   - Go API  on http://localhost:8080  (owns accounts, content and grading)
#   - Next.js on http://localhost:3000  (proxies /api/v1/* to the API)
# =============================================================================

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "   PREPYO - LOCAL ENVIRONMENT LAUNCHER   " -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

# 1. Prerequisites
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] Node.js is not installed or not on PATH." -ForegroundColor Red
    Write-Host "Install Node 18+ from https://nodejs.org" -ForegroundColor Yellow
    exit 1
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] Go is not installed or not on PATH." -ForegroundColor Red
    Write-Host "Install Go 1.22+ from https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# 2. The API refuses to boot without a database, so check before starting.
if (-not (Test-Path "backend\.env")) {
    Write-Host "[ERROR] backend\.env is missing." -ForegroundColor Red
    Write-Host "Copy backend\.env.example to backend\.env and set DATABASE_URL." -ForegroundColor Yellow
    exit 1
}

# 3. Dependencies
if (-not (Test-Path "node_modules")) {
    Write-Host "[INFO] node_modules not found. Running npm install..." -ForegroundColor Yellow
    npm install
}

# 4. Go API in its own window so its logs stay readable.
Write-Host "`n[STARTING] Go API on http://localhost:8080..." -ForegroundColor Green
Start-Process -FilePath "cmd.exe" -ArgumentList "/k", "npm run dev:api" -WindowStyle Normal

Write-Host "[WAITING] Letting the API connect and run migrations..." -ForegroundColor DarkGray
Start-Sleep -Seconds 5

try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 5
    if ($health.status -eq "healthy") {
        Write-Host "[OK] API is healthy and reached the database." -ForegroundColor Green
    }
} catch {
    Write-Host "[WARN] API is not answering yet. Check the 'Prepyo API' window for errors." -ForegroundColor Yellow
    Write-Host "       The web app will show connection errors until it is up." -ForegroundColor Yellow
}

# 5. Web app in this window.
Write-Host "`n[STARTING] Next.js on http://localhost:3000..." -ForegroundColor Green
Write-Host "Press Ctrl+C to stop the web server. Close the API window to stop the API.`n" -ForegroundColor DarkGray

npm run dev
