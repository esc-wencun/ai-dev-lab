@echo off
rem ===== RuoYi local dev environment - STOP (Windows) =====
rem Stops and removes containers. Data volumes are KEPT (data survives restart).
rem To wipe all data: docker compose down -v
rem Note: keep this file ASCII-only (no Chinese) -- cmd.exe parses batch files
rem with the OEM codepage, and UTF-8 Chinese breaks parsing. Docs: see README.md

setlocal
cd /d "%~dp0"

where docker >nul 2>nul
if errorlevel 1 (
    echo [ERROR] docker command not found. Install and start Docker Desktop first.
    pause
    exit /b 1
)

echo [STOP] Stopping MySQL + Redis containers (data volumes kept)...
docker compose down
if errorlevel 1 (
    echo [ERROR] Stop failed. Check the messages above.
    pause
    exit /b 1
)

echo.
echo [DONE] Stopped. Data kept in volumes ruoyi-mysql-data / ruoyi-redis-data.
pause
