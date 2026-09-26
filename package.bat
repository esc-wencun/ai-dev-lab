@echo off
chcp 65001 >nul
setlocal
rem ============================================================
rem package.bat - 一键打包服务端(Java) + 客户端(Vue3)
rem 用法: 任意位置双击, 或命令行执行; 产物输出到 package\<当天日期>\
rem   package\YYYY-MM-DD\ruoyi-admin.jar   服务端
rem   package\YYYY-MM-DD\ruoyi-vue3.zip    客户端 (zip 根目录即站点内容)
rem 任何一步失败立即终止并返回非 0 退出码
rem ============================================================

cd /d "%~dp0" || exit /b 1
set "ROOT=%CD%"

rem JAVA_HOME 兜底: Maven 需要 JDK 17 (本机 PATH 上的 java 是 1.6)
if not defined JAVA_HOME (
    if exist "C:\Program Files\Java\jdk-17.0.12" set "JAVA_HOME=C:\Program Files\Java\jdk-17.0.12"
    if exist "C:\Program Files\Java\jdk-17" set "JAVA_HOME=C:\Program Files\Java\jdk-17"
)

rem 版本号 = 当天日期
for /f %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-dd"') do set "VERSION=%%i"
set "OUT=%ROOT%\package\%VERSION%"
echo [1/5] 版本号: %VERSION%    输出: %OUT%
if not exist "%OUT%" mkdir "%OUT%" || goto :fail

echo [2/5] 服务端: mvn clean package (RuoYi-Vue) ...
cd /d "%ROOT%\RuoYi-Vue" || goto :fail
call mvn clean package -Dmaven.test.skip=true -q
if errorlevel 1 goto :fail
copy /y "ruoyi-admin\target\ruoyi-admin.jar" "%OUT%\ruoyi-admin.jar" >nul || goto :fail

echo [3/5] 客户端: npm run build:prod (RuoYi-Vue3) ...
cd /d "%ROOT%\RuoYi-Vue3" || goto :fail
if not exist node_modules (
    echo        node_modules 不存在, 先 npm install ...
    call npm install --no-audit --no-fund
    if errorlevel 1 goto :fail
)
call npm run build:prod
if errorlevel 1 goto :fail

echo [4/5] 压缩 dist -^> ruoyi-vue3.zip ...
powershell -NoProfile -Command "Compress-Archive -Path dist/* -DestinationPath '%OUT%\ruoyi-vue3.zip' -Force"
if errorlevel 1 goto :fail

echo [5/5] 打包完成, 产物:
dir "%OUT%" | findstr /i "ruoyi"
cd /d "%ROOT%"
exit /b 0

:fail
echo.
echo [错误] 打包失败, 请查看上方错误输出.
exit /b 1
