# Docker 本地环境（MySQL 8.0 + Redis Stack）

一套 **Windows / macOS / Linux 通用** 的本地 MySQL + Redis 环境，供开源用户开箱即用，避免把真实云数据库凭据提交到 GitHub。

## 前提

- 已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)（Windows / macOS）或 Docker Engine + Compose v2（Linux）
- 本机 13306、16379 端口空闲（默认端口+10000 约定，避免与其他项目栈冲突；被占用时可在 `docker/.env` 里改 `MYSQL_PORT` / `REDIS_PORT`）

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
2. 执行 `mysql/init/01-ruoyi-schema.sql` + `02-quartz.sql` 建表并灌入数据（含 `admin / admin123` 账号与 AI 模块预置数据）

> `01-ruoyi-schema.sql` 由 `export-init-sql.sh` 从运行库导出生成（日志表只留结构、软删行剔除），不要手工编辑；运行库表结构变更后执行 `bash docker/export-init-sql.sh` 重新生成即可。注意 init SQL 只在数据卷为空时执行，对已有数据卷不生效——要应用新导出的数据需 `docker compose down -v` 重置。

等待 MySQL 健康检查通过（首次初始化约 1~2 分钟）：

```bash
docker compose ps        # STATE 变为 healthy 即就绪
```

## 启动后端

**Python 版**（配置文件 `.env.docker` 已指向本地 Docker，密码与默认 compose 配置一致）：

```bash
cd ../RuoYi-Vue-FastApi
.venv\Scripts\python app.py --env=docker    # Windows（项目 venv，建立方式见该项目 README「安装依赖」）
python3 app.py --env=docker                 # macOS / Linux
```

**Java 版**：把 `RuoYi-Vue/ruoyi-admin/src/main/resources/application-druid.yml` 的数据库 URL 指到 `127.0.0.1:13306/ry-vue`（账号 `wencun` / `111111`）、`application.yml` 的 redis 指到 `127.0.0.1:16379`（密码 `111111`）即可。

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
├── export-init-sql.sh      # 从运行库重新导出 init SQL（表结构变更后执行）
├── docker-compose.yml      # MySQL 8.0 + Redis Stack 服务定义
├── .env.example            # 端口/密码/内存覆盖模板（.env 与其同内容，均已提交）
└── mysql/init/
    ├── 01-ruoyi-schema.sql # 业务表（26 张）+ 数据快照（export-init-sql.sh 生成；日志表只留结构、软删行剔除）
    └── 02-quartz.sql       # Quartz 表（11 张，供 Java 版使用）
```

## 说明

- Compose 项目名固定为 `ruoyi`（`docker-compose.yml` 顶部 `name:`），实际资源名：容器 `ruoyi-mysql` / `ruoyi-redis`，数据卷 `ruoyi_ruoyi-mysql-data` / `ruoyi_ruoyi-redis-data`，网络 `ruoyi_default`——不随目录名变化。
- MySQL 数据卷：`ruoyi-mysql-data`；Redis 开启 AOF 持久化，数据卷 `ruoyi-redis-data`。`down -v` 彻底重置。
- Redis 使用 **Redis Stack** 镜像（`redis/redis-stack-server`），在普通 Redis 之上内置 RediSearch（全文/向量检索）、RedisJSON 等模块——除 RuoYi 的会话/缓存外，也满足 Spring AI `RedisVectorStore` 等向量场景（`FT.CREATE` / `JSON.*` 可用）。启动参数经 `REDIS_ARGS` 环境变量传入，密码/端口/数据卷与原 `redis:7.0` 部署完全一致，客户端零改动；默认 `maxmemory 512mb` + `noeviction`（向量场景写满报错不丢数据，可在 `.env` 里调整）。
- 字符集统一 `utf8mb4`，时区 `+08:00`，表名小写（`lower_case_table_names=1`，与 Java 版线上行为一致）。
- 镜像 `mysql:8.0` / `redis/redis-stack-server` 均支持 amd64 与 arm64（Apple Silicon Mac 可直接跑）。
- 这套环境是**本地开发用**，请勿原样暴露到公网（默认密码是公开的）。
