---
name: export-init-sql
description: 把 ry-vue 运行库（Docker MySQL）导出为 docker/mysql/init/01-ruoyi-schema.sql 开箱初始化脚本。凡是用户提到"导出 init SQL""同步表结构到 docker""更新 docker 初始化脚本""导出建库 SQL""重新生成 01-ruoyi-schema.sql"，或表结构/预置数据变更后要把变更固化进 docker 初始化环境时，使用本技能。
---

# 导出 init SQL 到 docker 目录

## 唯一入口：导出脚本

**不要手写或手工编辑 `docker/mysql/init/01-ruoyi-schema.sql`**——它是运行库的衍生副本，永远通过脚本从运行中的 Docker MySQL 导出重新生成：

```bash
cd 仓库根目录 && bash docker/export-init-sql.sh
```

脚本自动完成全部分组逻辑，输出文件固定为 `docker/mysql/init/01-ruoyi-schema.sql`。

## 导出规则（脚本内固化，勿在对话中临时更改）

1. **Quartz 表（`qrtz_*`）不导出**——由 `docker/mysql/init/02-quartz.sql` 单独负责，不要混进 01 文件。
2. **日志表 `sys_oper_log` / `sys_logininfor` / `sys_job_log` 只导结构不导数据**（2026-10-14 用户明确要求）。
3. **软删数据不导出**：凡有 `del_flag` 列的表，导出时带 `--where="del_flag<>'2'"` 剔除软删行。
4. 其余业务表全量导出（结构 + 数据）。数据里允许包含对话内容、会话等（用户 2026-10-14 确认无隐私顾虑）。
5. 文件头保留 `SET NAMES utf8mb4;`（镜像初始化导入时保证中文不乱码）。

## 环境前提与覆盖

- 默认连 `ruoyi-mysql` 容器、库 `ry-vue`、root 密码 `111111`（与 docker/ 目录 compose 默认一致）。
- 容器没起先 `cd docker && docker compose up -d` 等健康检查通过再导出。
- 非默认环境时用环境变量覆盖，不改脚本：

```bash
CONTAINER=<容器名> DB=<库名> MYSQL_ROOT_PASSWORD=<密码> bash docker/export-init-sql.sh
```

## 导出后必须验证

init SQL 只在数据卷为空时执行一次（`/docker-entrypoint-initdb.d` 机制），改动对已有卷不生效。验证方式是起临时容器空库初始化（用完即删，勿占用 13306）：

```bash
docker rm -f ruoyi-init-verify 2>/dev/null; docker run -d --name ruoyi-init-verify \
  -e MYSQL_ROOT_PASSWORD=111111 -e MYSQL_DATABASE=ry-vue \
  -e MYSQL_USER=wencun -e MYSQL_PASSWORD=111111 \
  -v "$(pwd)/docker/mysql/init:/docker-entrypoint-initdb.d:ro" mysql:8.0 \
  mysqld --character-set-server=utf8mb4 --collation-server=utf8mb4_general_ci \
  --character-set-client-handshake=FALSE --default-time-zone=+08:00 --lower_case_table_names=1
```

等 1~2 分钟初始化完成后核对（核对项按本次变更内容调整）：

```bash
V="docker exec ruoyi-init-verify mysql -uroot -p111111 --default-character-set=utf8mb4 ry-vue -N -B -e"
$V "select count(*) from information_schema.tables where table_schema='ry-vue'"   # 表数（26 业务+11 Quartz=37）
$V "select count(*) from sys_oper_log"                                            # 日志表应为 0
$V "select menu_name from sys_menu where menu_id=2020"                            # 中文应正常显示
docker rm -f ruoyi-init-verify                                                    # 用完即删
```

脚本自带静态自检（日志表 INSERT 必须为 0，否则退出码非 0），但动态验证（临时容器）不能省——尤其涉及表结构变更时。

## 相关约束（见 docker/AGENTS.md）

- 表结构权威源在 `RuoYi-Vue/sql/`（增量 SQL），本 init 是运行库的衍生副本；正常流程是先在运行库执行增量 SQL，再跑本导出。
- 生成的 init SQL 对已有数据卷不生效；用户要"重置数据"需 `cd docker && docker compose down -v` 后重启。
- 改动提交 git 时，`docker/mysql/init/01-ruoyi-schema.sql` 与脚本 `docker/export-init-sql.sh` 一并提交。
