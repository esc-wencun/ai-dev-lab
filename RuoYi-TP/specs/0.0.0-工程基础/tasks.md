# 0.0.0-工程基础 · 任务分解

> 做完即勾（含对应验证跑通）；没做不勾并在行尾注明原因——勾选纪律见根 AGENTS.md。

## Task 1 · PHP 8.3 环境安装（2026-09-29 完成）

- [x] 解压 `C:\php\php-8.3.zip` 到 `C:\php\8.3\`
- [x] 配置 `php.ini`：启用 pdo_mysql / gd / openssl / fileinfo / mbstring / curl / zip / intl，extension_dir，时区 Asia/Shanghai
- [x] `C:\php\8.3` 加入用户 PATH，新开终端 `php -v` 显示 8.3.x——**8.3.35**；踩坑：本机 System32 的 VCRUNTIME140.dll 版本过旧（14.0 vs 需要 14.29），PHP 直接不输出——从 Python314 目录复制新版 vcruntime140.dll 到 `C:\php\8.3\` 解决（零系统侵入）
- [x] `php -m` 确认扩展齐全（8/8）

## Task 2 · Composer 安装与镜像（2026-09-29 完成）

- [x] composer.phar 放 `C:\php\8.3\`，写 `composer.bat`
- [x] `composer -V` 正常（2.10.3）；已配置阿里云镜像
- [ ] 临时目录验证 `composer create-project topthink/think` 可用（验证完删除）——合并进 Task 3 直接建正式骨架验证

## Task 3 · TP8 工程骨架（2026-09-29 完成）

- [x] 在 `RuoYi-TP/` 生成 TP8 骨架——临时目录 create-project 后并入（目标目录非空）；**踩坑：阿里云镜像的 think-trace 2.0.x-dev 与 minimum-stability 冲突装不上，clear 全局镜像用官方 packagist 解决**
- [x] 单应用模式定案（本系统无多终端场景）；结论回填 spec.md 实施记录
- [x] `php think run --port 8888` 启动，浏览器出 TP 欢迎页（ThinkPHP 8.1.4 实测）
- [x] gitignore：runtime/、vendor/、.env、**reference/**（已预置于 .gitignore）
- [x] 写 `RuoYi-TP/AGENTS.md`（分层纪律、常用命令、环境约束）

## Task 4 · 建库 ry-tp + 导入经典 SQL + Redis db1（2026-09-29 完成）

- [x] 建库 `ry-tp`（utf8mb4）——`wencun` 无建库权限（spec 预判命中），用 docker 内 root（-p111111）建库并 `GRANT ALL ON ry-tp.* TO wencun@%`
- [x] 导入 `reference/RuoYi-classic/sql/ry_20260319.sql`（未导入 quartz.sql）
- [x] 验证：sys_user 2 行（admin/ry，hash 与 spec 断言一致）；sys_menu 85 行带 url；sys_config **11 条**；qrtz 表 0 张
- [x] Redis db1 连通验证（set/get 后删除测试键）
- [x] config/database.php + config/cache.php + .env 按上述值落定（database.php 本就 env 驱动零改；cache.php 加 redis 连接段；predis v2 已装）

## Task 5 · 常量与 RedisCache 门面（2026-09-29 完成）

- [x] `app/common/TpConstant.php`：键前缀常量（session:/pwd_retry:/repeat_submit:/rate_limit:/config:/dict:）+ code/ttl 常量
- [x] `app/common/RedisCache.php`：get/set/delete/keysScan/expire/has/ttl/incrWithExpire；JSON 编解码收口——**踩坑：predis scan 返回 [cursor, keys[]] 二元组且 cursor 是 string '0' 判断终止，按文档直觉写会踩**
- [x] PHPUnit 单元测试：常量值与经典版对位断言（code 0/301/500、5 次 600s、30min 空闲）+ 前缀冒号约定；RedisCache 走真库端到端（set/get/ttl/has/incr/scan 全过，测试键已清理）

## Task 6 · AjaxResult / TableDataInfo / PasswordService（2026-09-29 完成）

- [x] `AjaxResult`：success（code **0**）/ warn（**301**）/ error（**500**）——工厂模式返回 `think\Response::create($body,'json')`（**踩坑：think Response/Json 构造器需容器注入 Cookie，不能直接 new，改工厂**）
- [x] `TableDataInfo`：`{code:0, msg:"查询成功", rows, total}`
- [x] `PasswordService`：`encrypt($loginName, $password, $salt)` = `md5(loginName.password.salt)`；`randomSalt()` = `bin2hex(random_bytes(3))`
- [x] **密码实测**：PHPUnit 固化断言 admin/ry 两个官方预置 hash（10 tests 39 assertions 全绿）
- [x] 全局异常渲染：`app/ExceptionHandle.php` render 中 BusinessException → `AjaxResult::error`（HTTP 200 + code 500）

## Task 7 · 收尾（2026-09-29 完成）

- [x] 更新本文件全部勾选、spec.md 实施记录（含 Task 3 决策点结论与版本回填）
- [x] checklist.md 全部核对通过后，README.md 模块总表标 ✅ + 日期
