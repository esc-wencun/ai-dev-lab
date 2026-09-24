#!/usr/bin/env bash
# ===== RuoYi 本地环境停止脚本（macOS / Linux）=====
# 停止并移除容器，数据卷保留（数据不丢）。如需彻底清空数据，运行: docker compose down -v

set -e
cd "$(dirname "$0")"

if ! command -v docker >/dev/null 2>&1; then
    echo "[错误] 未检测到 docker 命令。"
    exit 1
fi

echo "[停止] MySQL + Redis 容器（数据卷保留，重启后数据仍在）..."
docker compose down

echo ""
echo "[完成] 已停止。数据保留在卷 ruoyi-mysql-data / ruoyi-redis-data 中。"
