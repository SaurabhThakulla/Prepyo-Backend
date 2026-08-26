@echo off
title Prepyo Local Launcher
echo ====================================================
echo             PREPYO LOCAL LAUNCHER
echo ====================================================
echo.
echo Prepyo needs two processes:
echo   - Go API  on http://localhost:8080  (owns all data)
echo   - Next.js on http://localhost:3000  (the interface)
echo.

where go >nul 2>nul
if errorlevel 1 (
  echo [ERROR] Go is not installed or not on PATH.
  echo Install Go 1.22+ from https://go.dev/dl/
  exit /b 1
)

if not exist "backend\.env" (
  echo [ERROR] backend\.env is missing. Copy backend\.env.example and set DATABASE_URL.
  exit /b 1
)

if not exist "node_modules" (
  echo [INFO] Installing dependencies...
  call npm install
)

echo [STARTING] Go API in a separate window...
start "Prepyo API" cmd /k "npm run dev:api"

echo [WAITING] Giving the API a few seconds to run migrations...
timeout /t 5 /nobreak >nul

echo [STARTING] Next.js on http://localhost:3000
echo Press Ctrl+C to stop the web server. Close the "Prepyo API" window to stop the API.
echo.
npm run dev
