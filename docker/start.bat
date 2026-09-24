@echo off
rem ===== RuoYi local dev environment - START (Windows) =====
rem Starts MySQL 8.0 + Redis 7.0 containers. First start auto-imports schema.
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

docker info >nul 2>nul
if errorlevel 1 (
    echo [WAIT] Docker engine not ready. Waiting 15s for Docker Desktop to start...
    timeout /t 15 /nobreak >nul
    docker info >nul 2>nul
    if errorlevel 1 (
        echo [ERROR] Docker engine still unavailable. Start Docker Desktop manually, then retry.
        pause
        exit /b 1
    )
)

echo [START] MySQL 8.0 + Redis 7.0 ...
docker compose up -d
if errorlevel 1 (
    echo [ERROR] Start failed. Check the messages above.
    pause
    exit /b 1
)

echo.
echo [DONE] Containers started. First-time initialization takes 1-2 minutes.
echo   MySQL: 127.0.0.1:3306  user: wencun / 111111  database: ry-vue
echo   Redis: 127.0.0.1:6379  password: 111111
echo Status: docker compose ps
pause
