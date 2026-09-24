@echo off
echo.
echo [run web]
echo.

%~d0
cd %~dp0

echo [activate node version]
nvm use 22.14.0

cd ..
npm run dev

pause
