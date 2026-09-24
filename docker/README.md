# Docker 本地环境（MySQL 8.0 + Redis 7.0）

一套 **Windows / macOS / Linux 通用** 的本地 MySQL + Redis 环境，供开源用户开箱即用，避免把真实云数据库凭据提交到 GitHub。

## 前提

- 已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)（Windows / macOS）或 Docker Engine + Compose v2（Linux）
- 本机 3306、6379 端口空闲（被占用时可在 `docker/.env` 里改 `MYSQL_PORT` / `REDIS_PORT`）

## 快速开始

双击运行 `start.bat`（Windows），或在终端执行（macOS / Linux）：

```bash
./start.sh
```

也可以直接用 compose 命令：

```bash
cd docker
docker compose up -d
```

首次启动会自动：

1. 创建数据库 `ry-vue`，普通账号 `wencun` / `111111`（root 密码也是 `111111`）
2. 执行 `mysql/init/01-ruoyi-schema.sql` + `02-quartz.sql` 建表并灌入若依预置数据（含 `admin / admin123` 账号）

等待 MySQL 健康检查通过（首次初始化约 1~2 分钟）：

```bash
docker compose ps        # STATE 变为 healthy 即就绪
```

## 启动后端

**Python 版**（配置文件 `.env.docker` 已指向本地 Docker，密码与默认 compose 配置一致）：

```bash
cd ../RuoYi-Vue-FastApi
"C:\Program Files\Python310\python.exe" app.py --env=docker    # Windows
python3 app.py --env=docker                                    # macOS / Linux
```

**Java 版**：把 `RuoYi-Vue/ruoyi-admin/src/main/resources/application-druid.yml` 的数据库 URL 指到 `127.0.0.1:3306/ry-vue`（账号 `wencun` / `111111`）、`application.yml` 的 redis 指到 `127.0.0.1:6379`（密码 `111111`）即可。

默认账号：`admin` / `admin123`

## 常用命令

| 操作 | Windows | macOS / Linux |
|------|---------|---------------|
| 启动 | 双击 `start.bat` | `./start.sh` |
| 停止（保留数据） | 双击 `stop.bat` | `./stop.sh` |
| 重置数据（删库重来） | `docker compose down -v` 后重新启动 | 同左 |
| 查看状态 | `docker compose ps` | 同左 |
| 查看日志 | `docker compose logs -f mysql` | 同左 |

所有命令均在 `docker` 目录下执行。

## 自定义端口 / 密码

```bash
cp .env.example .env    # 修改端口或密码
docker compose up -d
```

改了 MySQL/Redis 密码后，需同步修改 `RuoYi-Vue-FastApi/.env.docker`（或自建 `.env.dev`）。
注意：修改密码不影响已初始化的数据卷；若要换数据库名，需 `docker compose down -v` 重置。

## 目录结构

```
docker/
├── start.bat / start.sh    # 启动脚本（Windows / macOS / Linux）
├── stop.bat / stop.sh      # 停止脚本（数据保留）
├── docker-compose.yml      # MySQL 8.0 + Redis 7.0 服务定义
├── .env.example            # 端口/密码覆盖模板（.env 已被 git 忽略）
└── mysql/init/
    ├── 01-ruoyi-schema.sql # 若依业务表（20 张，来自 RuoYi-Vue/sql）
    └── 02-quartz.sql       # Quartz 表（11 张，供 Java 版使用）
```

## 说明

- MySQL 数据卷：`ruoyi-mysql-data`；Redis 开启 AOF 持久化，数据卷 `ruoyi-redis-data`。`down -v` 彻底重置。
- 字符集统一 `utf8mb4`，时区 `+08:00`，表名小写（`lower_case_table_names=1`，与 Java 版线上行为一致）。
- 镜像 `mysql:8.0` / `redis:7.0` 均支持 amd64 与 arm64（Apple Silicon Mac 可直接跑）。
- 这套环境是**本地开发用**，请勿原样暴露到公网（默认密码是公开的）。
