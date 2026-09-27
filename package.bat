@echo off
chcp 65001 >nul
setlocal EnableExtensions
rem ============================================================
rem package.bat - 一键打包四端产物
rem 用法: 任意位置双击, 或命令行执行; 产物输出到 package\<日期-时分>\
rem   package\YYYY-MM-DD-HHmm\ruoyi-admin.jar         Java 服务端
rem   package\YYYY-MM-DD-HHmm\ruoyi-vue3.zip          Vue3 客户端 (zip 根目录即站点内容)
rem   package\YYYY-MM-DD-HHmm\ruoyi-fastapi-src.zip   Python 服务端源码 (部署时建 venv 装 requirements)
rem   package\YYYY-MM-DD-HHmm\go\server-*             Go 服务端三平台二进制, configs/ 需与 exe 同级
rem 打包源码取自当前工作区完整快照(含未提交改动, 排除 gitignore 忽略项):
rem 临时 index + git archive 展开到 %TEMP%, 构建全程不碰工作区, 结束自动清理.
rem 版本号带时分, 单日多次打包互不覆盖.
rem 任何一步失败立即终止并返回非 0 退出码
rem ============================================================

cd /d "%~dp0" || exit /b 1
set "ROOT=%CD%"

rem JAVA_HOME 兜底: Maven 需要 JDK 17 (本机 PATH 上的 java 是 1.6)
if not defined JAVA_HOME (
    if exist "C:\Program Files\Java\jdk-17.0.12" set "JAVA_HOME=C:\Program Files\Java\jdk-17.0.12"
    if exist "C:\Program Files\Java\jdk-17" set "JAVA_HOME=C:\Program Files\Java\jdk-17"
)

rem 版本号 = 日期-时分
for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd-HHmm"') do set "VERSION=%%i"
set "OUT=%ROOT%\package\%VERSION%"
set "SNAP=%TEMP%\ruoyi-package-%VERSION%"
set "PIDX=%TEMP%\ruoyi-package-index"
echo [1/7] 版本号: %VERSION%    输出: %OUT%
if not exist "%OUT%" mkdir "%OUT%" || goto :fail

echo [2/7] 生成工作区快照: 含未提交改动, 排除 gitignore 项 ...
if exist "%SNAP%" rmdir /s /q "%SNAP%"
set "GIT_INDEX_FILE=%PIDX%"
del "%PIDX%" 2>nul
git read-tree HEAD || goto :fail
git add --all || goto :fail
for /f %%i in ('git write-tree') do set "TREE=%%i"
if not defined TREE goto :fail
set "GIT_INDEX_FILE="
del "%PIDX%" 2>nul
mkdir "%SNAP%" || goto :fail
rem 用 zip 展开: System32 bsdtar 在 GBK 代码页下解不开 UTF-8 中文路径, PowerShell(.NET) 认 UTF-8 标志位
set "SNAPZIP=%TEMP%\ruoyi-package-snapshot.zip"
del "%SNAPZIP%" 2>nul
git archive --format=zip --output="%SNAPZIP%" "%TREE%" || goto :fail
powershell -NoProfile -Command "Expand-Archive -Path '%SNAPZIP%' -DestinationPath '%SNAP%' -Force"
if errorlevel 1 goto :fail
del "%SNAPZIP%" 2>nul

echo [3/7] 服务端: mvn clean package (RuoYi-Vue) ...
cd /d "%SNAP%\RuoYi-Vue" || goto :fail
call mvn clean package -Dmaven.test.skip=true -q
if errorlevel 1 goto :fail
copy /y "ruoyi-admin\target\ruoyi-admin.jar" "%OUT%\ruoyi-admin.jar" >nul || goto :fail

echo [4/7] 客户端: npm run build:prod (RuoYi-Vue3) ...
cd /d "%SNAP%\RuoYi-Vue3" || goto :fail
if exist "%ROOT%\RuoYi-Vue3\node_modules" (
    echo        复用工作区 node_modules, 复制到快照 ...
    robocopy "%ROOT%\RuoYi-Vue3\node_modules" "node_modules" /E /NFL /NDL /NJH /NJS /NP /MT:16
    if errorlevel 8 goto :fail
)
if not exist node_modules (
    echo        node_modules 不存在, 先 npm install ...
    call npm install --no-audit --no-fund
    if errorlevel 1 goto :fail
)
call npm run build:prod
if errorlevel 1 goto :fail

echo [5/7] 压缩 dist -^> ruoyi-vue3.zip ...
powershell -NoProfile -Command "Compress-Archive -Path dist/* -DestinationPath '%OUT%\ruoyi-vue3.zip' -Force"
if errorlevel 1 goto :fail

echo [6/7] Go 服务端: 三平台交叉编译 (RuoYi-Vue-GO) ...
cd /d "%SNAP%\RuoYi-Vue-GO" || goto :fail
set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=amd64"
go build -trimpath -ldflags="-s -w" -o "%OUT%\go\server-windows-amd64.exe" ./cmd/server
if errorlevel 1 goto :fail
set "GOOS=darwin"
set "GOARCH=arm64"
go build -trimpath -ldflags="-s -w" -o "%OUT%\go\server-darwin-arm64" ./cmd/server
if errorlevel 1 goto :fail
set "GOOS=linux"
set "GOARCH=amd64"
go build -trimpath -ldflags="-s -w" -o "%OUT%\go\server-linux-amd64" ./cmd/server
if errorlevel 1 goto :fail
if not exist "%OUT%\go\configs" mkdir "%OUT%\go\configs"
copy /y "configs\.env.dev" "%OUT%\go\configs\.env.dev" >nul || goto :fail

echo [7/7] Python 服务端: 源码压缩 (RuoYi-Vue-FastApi) ...
powershell -NoProfile -Command "Compress-Archive -Path '%SNAP%\RuoYi-Vue-FastApi/*' -DestinationPath '%OUT%\ruoyi-fastapi-src.zip' -Force"
if errorlevel 1 goto :fail

cd /d "%ROOT%"
rmdir /s /q "%SNAP%"
echo.
echo 打包完成, 产物:
dir "%OUT%"
echo go 子目录:
dir "%OUT%\go"
exit /b 0

:fail
set "GIT_INDEX_FILE="
cd /d "%ROOT%" 2>nul
if exist "%SNAP%" rmdir /s /q "%SNAP%"
echo.
echo [错误] 打包失败, 请查看上方错误输出.
exit /b 1
