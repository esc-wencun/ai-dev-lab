# 0.0.0-工程基础 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因。测试数据清理记录写在本文件末尾。

## 环境（2026-09-29 核对通过）

- [x] `php -v` 输出 8.3.x（8.3.35；用户 PATH 已加 C:\php\8.3，新终端生效）
- [x] `php -m` 含 pdo_mysql / gd / intl / zip / mbstring / openssl / fileinfo / curl（8/8）
- [x] `composer -V` 正常（2.10.3，官方 packagist 源——阿里云镜像与 think-trace dev 包冲突已弃用）
- [x] `php think run --port 8888` 可启动，浏览器出 TP 欢迎页（ThinkPHP 8.1.4）

## 库与连通（2026-09-29 核对通过）

- [x] `ry-tp` 库存在，经典版官方 SQL 导入完成（sys_user 2 行 / sys_menu 85 行带 url 列 / sys_config **11 条预置**，无 captcha 相关键）
- [x] quartz.sql 未导入（information_schema 查询 qrtz 表 = 0）
- [x] Redis db1 连通，测试键已清理（KEYS * 为空）
- [x] .env / runtime/ / vendor/ / reference/ 已 gitignore（.gitignore 四条齐全；git status 待首次提交时核对）

## 门面与信封（code 体系 = 经典版 0/301/500，2026-09-29 核对通过）

- [x] RedisCache 端到端验证通过（真库 set/get/ttl/has/incr/scan）；grep 检查：当前业务代码仅 RedisCache 自身引用 predis（业务模块尚未生长，1.0.0 起每模块动工时复查）
- [x] AjaxResult::success() 输出 `{"code":0,"msg":"操作成功"}`——**code 是 0 不是 200**（对齐经典版 AjaxResult.Type，CLI 实测）
- [x] TableDataInfo 输出 `{"code":0,"msg":"查询成功","rows":[...],"total":n}`（CLI 实测）
- [x] BusinessException → `{"code":500,"msg":"..."}`，HTTP 状态 200（ExceptionHandle 已接线；端到端确认在 1.0.0/2.0.0 有真实业务异常时复查）

## 密码方案（2026-09-29 核对通过）

- [x] PHPUnit：`PasswordService::encrypt('admin','admin123','111111')` === `29c67a30398638269fe600f73a054934`（另固化 ry/222222 官方预置）
- [x] `randomSalt()` 输出 6 位 hex（正则断言 `[0-9a-f]{6}` × 20 次）
- [x] 本模块对库的写入仅为导入官方 SQL，Redis 测试键全部清理

## 测试数据清理记录

- 2026-09-29：Redis db1 测试键（session:test-*、session:scan-*、pwd_retry:admin）全部删除，`KEYS *` 复查为空；MySQL 仅有官方 SQL 预置数据，无测试写入。
