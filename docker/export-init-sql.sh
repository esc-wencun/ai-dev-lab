#!/usr/bin/env bash
# 导出 ry-vue 运行库 → docker/mysql/init/01-ruoyi-schema.sql（可重复执行）
#
# 导出规则（2026-10-14 约定）：
#   1. Quartz 表（qrtz_*）不导出——02-quartz.sql 负责
#   2. 日志表 sys_oper_log / sys_logininfor / sys_job_log 只导结构（--no-data）
#   3. 带 del_flag 的表剔除软删行（--where="del_flag<>'2'"）
#   4. 其余业务表全量结构 + 数据
#
# 用法：仓库根目录执行  bash docker/export-init-sql.sh
# 覆盖：CONTAINER=<容器名> DB=<库名> MYSQL_ROOT_PASSWORD=<密码> bash docker/export-init-sql.sh
set -euo pipefail

CONTAINER="${CONTAINER:-ruoyi-mysql}"
DB="${DB:-ry-vue}"
ROOT_PWD="${MYSQL_ROOT_PASSWORD:-111111}"
OUT="$(cd "$(dirname "$0")" && pwd)/mysql/init/01-ruoyi-schema.sql"
LOG_TABLES="sys_oper_log sys_logininfor sys_job_log"

docker exec "$CONTAINER" mysql -uroot -p"$ROOT_PWD" -e "select 1" >/dev/null 2>&1 \
  || { echo "错误：MySQL 容器 $CONTAINER 不可用" >&2; exit 1; }

mysql_exec() {
  docker exec "$CONTAINER" mysql -uroot -p"$ROOT_PWD" -N -B "$@" 2>/dev/null
}

all_tables=$(mysql_exec -e "select table_name from information_schema.tables where table_schema='$DB' and table_name not like 'qrtz%' order by table_name" | tr '\n' ' ')
delflag_tables=$(mysql_exec -e "select table_name from information_schema.columns where table_schema='$DB' and column_name='del_flag'" | tr '\n' ' ')

dump_args=(--compact --add-drop-table --skip-extended-insert --hex-blob --no-tablespaces --default-character-set=utf8mb4)

dump() {
  docker exec "$CONTAINER" mysqldump -uroot -p"$ROOT_PWD" "${dump_args[@]}" "$@" 2>/dev/null
}

where_group=(); plain_group=(); struct_group=()
for t in $all_tables; do
  if [[ " $LOG_TABLES " == *" $t "* ]]; then
    struct_group+=("$t")
  elif [[ " $delflag_tables " == *" $t "* ]]; then
    where_group+=("$t")
  else
    plain_group+=("$t")
  fi
done

{
  printf 'SET NAMES utf8mb4;\n\n'
  if [ ${#where_group[@]} -gt 0 ]; then dump --where="del_flag<>'2'" "$DB" "${where_group[@]}"; fi
  if [ ${#plain_group[@]} -gt 0 ]; then dump "$DB" "${plain_group[@]}"; fi
  if [ ${#struct_group[@]} -gt 0 ]; then dump --no-data "$DB" "${struct_group[@]}"; fi
} > "$OUT"

# 静态自检：日志表必须零数据
n_tables=$(grep -c '^CREATE TABLE' "$OUT" || true)
n_inserts=$(grep -c '^INSERT INTO' "$OUT" || true)
n_log_inserts=$(grep -cE "^INSERT INTO \`(sys_oper_log|sys_logininfor|sys_job_log)\`" "$OUT" || true)
echo "已导出 $OUT"
echo "  表：$n_tables 张（Quartz 除外）"
echo "  数据：$n_inserts 条 INSERT（日志表 $n_log_inserts 条，应为 0）"
[ "$n_log_inserts" -eq 0 ]
