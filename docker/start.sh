#!/usr/bin/env bash
# ===== RuoYi 本地环境启动脚本（macOS / Linux）=====
# 启动 MySQL 8.0 + Redis 7.0 容器，首次启动自动建库建表

set -e
cd "$(dirname "$0")"

if ! command -v docker >/dev/null 2>&1; then
    echo "[错误] 未检测到 docker 命令，请先安装 Docker Desktop (macOS) 或 Docker Engine (Linux)。"
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "[提示] Docker 引擎未就绪。"
    echo "  macOS: 请先启动 Docker Desktop 应用（Spotlight 搜索 Docker）"
    echo "  Linux: 执行 sudo systemctl start docker"
    exit 1
fi

echo "[启动] MySQL 8.0 + Redis 7.0 ..."
docker compose up -d

echo ""
echo "[完成] 容器已启动（首次初始化约 1~2 分钟）。"
echo "  MySQL:  127.0.0.1:3306  账号 wencun / 111111  库名 ry-vue"
echo "  Redis:  127.0.0.1:6379  密码 111111"
echo "查看状态: docker compose ps"
