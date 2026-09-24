@echo off
echo.
echo [npm install]
echo.

%~d0
cd %~dp0

echo [activate node version]
nvm use 22.14.0

cd ..
npm install --registry=https://registry.npmmirror.com

pause
