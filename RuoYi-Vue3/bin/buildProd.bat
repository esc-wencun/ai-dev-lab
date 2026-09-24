@echo off
echo.
echo [npm run build:prod]
echo.

%~d0
cd %~dp0

echo [activate node version]
nvm use 22.14.0

cd ..
npm run build:prod

pause
