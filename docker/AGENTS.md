# AGENTS.md — docker/

docker/ 子目录的专用约定（所有 AI 编码工具通用）。工作区级规范见根目录 [../AGENTS.md](../AGENTS.md)，其中「共用库纪律」（测试数据清理、admin 密码/会话复原）同样适用于本目录管理的 MySQL 与 Redis。

## 定位

本目录提供**开箱即用的本地 MySQL 8.0 + Redis 7.0**（Docker Compose），供三个后端版本共用，避免把真实云数据库凭据提交到 GitHub。面向人类的完整使用说明在 [README.md](./README.md)，本文件只写 AI 修改代码时需要知道的约束。

## 结构与职责

- `docker-compose.yml` — 唯一的服务定义（mysql:8.0 + redis:7.0）。端口/密码/库名全部走 `${VAR:-default}` 环境变量，默认值即 README 承诺的开箱值（3306/6379，密码 111111，库 `ry-vue`，账号 `wencun`）。
- `.env.example` — 覆盖模板；真实 `.env` 含密码，已被 gitignore，禁止提交。
- `start.sh` / `start.bat` / `stop.sh` / `stop.bat` — 平台启动/停止包装脚本，逻辑必须保持四份脚本语义一致（等价能力跨平台成对）。
- `mysql/init/01-ruoyi-schema.sql` — 若依业务表（20 张）+ 预置数据（来自 RuoYi-Vue/sql）。
- `mysql/init/02-quartz.sql` — Quartz 11 张表（仅 Java 版调度用）。

## 关键约束（改坏会直接影响三版后端）

1. **init SQL 只在数据卷为空时执行一次**（`/docker-entrypoint-initdb.d` 机制）。修改 init SQL 对已有数据卷**不生效**，验证改动必须 `docker compose down -v` 删卷重来。
2. **表结构变更的权威源在 Java 版** `RuoYi-Vue/sql/`（及后续 MP 演进），本目录的 init SQL 是它的衍生副本。同步来源改动时整文件对应搬运，不要在本目录里单独发明表结构。
3. **默认凭据是公开的本地开发值**，改动默认值会破坏「开箱即用 + 与各版 `.env.dev` 一致」的承诺；确需改密码/端口，只改 `.env.example` 说明与 README，并同步提示用户改各版 `.env.dev`（见 AGENTS.local.md「MySQL / Redis 当前指向」的更新约定）。
4. **字符集三件套不能拆**：`--character-set-server=utf8mb4` + `--collation-server=utf8mb4_general_ci` + `--character-set-client-handshake=FALSE`。缺最后一件，镜像初始化导入 SQL 时中文会双重编码成乱码（README/compose 注释已记录）。
5. **stop 脚本必须保留数据卷**（`docker compose down`，不带 `-v`）；`down -v` 只作为 README 里的手动重置选项。
6. **批处理脚本保持 ASCII-only**（`start.bat`/`stop.bat`）——cmd.exe 用 OEM 代码页解析批处理，UTF-8 中文会导致解析错误。中文输出只写在 `.sh` 版本里。
7. 连接信息（host/port/账号/库名）与根 [../AGENTS.local.md](../AGENTS.local.md) 「MySQL / Redis 当前指向」一节是**同一份事实的两处表述**：改了 compose 默认值必须同步那一节与各版 `.env.dev`。
