@echo off
rem RuoYi-React 前端生产构建脚本
call nvm use 22.14.0
cd /d %~dp0..
npm run build:prod
