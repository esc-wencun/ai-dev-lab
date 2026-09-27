@echo off
rem RuoYi-React 前端开发启动脚本（对齐 RuoYi-Vue3/bin 风格）
call nvm use 22.14.0
cd /d %~dp0..
npm run dev
