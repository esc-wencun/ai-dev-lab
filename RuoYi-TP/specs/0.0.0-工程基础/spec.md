# 00 工程规范基础：环境 / 骨架 / 自用库连通 / RedisCache 门面 / AjaxResult（code 0/301/500）

> **状态：✅ 已完成（2026-09-29）**
> 所有模块（含 01/1.5）的地基。目标：本机能跑起 TP8 + 建好自用库 `ry-tp`（经典版官方 SQL）+ 连通 Redis db1 + 密码方案落地验证。
> 基准：经典若依 4.8.3 源码（[../../reference/RuoYi-classic/](../../reference/RuoYi-classic/)）；选型依据 [../tech-stack.md](../tech-stack.md)。

## 目标

1. PHP 8.3 + Composer 可用（手动 zip 方式，已拍板）；
2. `RuoYi-TP/` TP8 骨架可跑（8888 端口）；
3. 建库 `ry-tp` 并导入 `reference/RuoYi-classic/sql/ry_20260319.sql`（**不导入 quartz.sql**）；
4. Redis db1 连通；
5. RedisCache 门面 + 键常量（TP 版自定，命名自由但集中管理）；
6. AjaxResult（**code 0/301/500**，对齐经典版 `AjaxResult.Type`）+ TableDataInfo（`{code:0, msg, rows, total}`）；
7. 密码方案类 `PasswordService::encrypt($loginName, $password, $salt)` = `md5(loginName.password.salt)` + `randomSalt()` = `bin2hex(random_bytes(3))`，**用 admin/admin123/111111 实测比对官方 SQL hash**。

## 环境安装步骤（Task 1~2，动手前先过一遍）

1. 解压 `C:\php\php-8.3.zip` → `C:\php\8.3\`（zip 已于 2026-09-28 下载）；
2. `php.ini-development` → `php.ini`，启用扩展：`pdo_mysql` / `gd` / `openssl` / `fileinfo` / `mbstring` / `curl` / `zip` / `intl`，`date.timezone = Asia/Shanghai`；
3. `C:\php\8.3` 加入用户 PATH；
4. Composer：composer.phar 放 `C:\php\8.3\` + `composer.bat`；阿里云镜像；
5. 验证：`php -v` ≥ 8.3、`php -m` 扩展齐全、`composer -V` 正常。

## 目录结构约定（TP8 标准骨架 + 工作区分层纪律）

```
RuoYi-TP/
├── app/
│   ├── controller/
│   │   ├── IndexController.php       # /index 主框架 + /system/main 首页
│   │   ├── LoginController.php       # /login /logout /captcha/captchaImage
│   │   └── system/                   # 系统管理（3.0.0+ 逐模块生长）
│   ├── service/                      # 业务逻辑、事务、缓存维护；禁止拼查询
│   ├── model/                        # think-orm 表模型；仅数据访问（对位 dao）
│   ├── validate/                     # TP 验证器
│   ├── middleware/                   # 登录态/权限/操作日志/防重（1.0.0 实现）
│   ├── common/
│   │   ├── constant.php              # Redis 键前缀 + 业务常量（TP 版自定命名，集中管理）
│   │   ├── AjaxResult.php            # {code, msg, ...}，code: 0 成功 / 301 警告 / 500 错误
│   │   ├── TableDataInfo.php         # {code:0, msg, rows, total}
│   │   └── PasswordService.php       # md5(loginName.password.salt) + randomSalt()
│   ├── facade/RedisCache.php         # 门面：键拼装/JSON 序列化/SCAN 替代 KEYS
│   ├── common/exception/             # 全局异常渲染 → AjaxResult(500)
│   ├── view/                         # ThinkTemplate 模板（1.5.0 起）
│   └── BaseController.php
├── config/
│   ├── database.php                  # MySQL ry-tp
│   ├── cache.php                     # predis（db1）
│   └── profile.php                   # 上传路径（deviations #14）
├── route/app.php
├── public/
│   ├── index.php
│   └── (js/ ajax/ css/ ruoyi/ img/ ...)   # 经典版 static/ 内容映射到站点根（1.5.0 移植，不能套 static/ 目录——模板以根路径引用）
├── runtime/                          # gitignore
├── reference/                        # 经典版源码（gitignore，只读）
├── composer.json
└── specs/                            # 本 spec 包
```

分层纪律：controller → service → model，约束同工作区（service 禁拼查询、model 仅数据访问）。

## 关键约定（PHP/TP 特有，全局唯一模式）

- **所有响应统一走 AjaxResult/TableDataInfo**，code 值体系 = 经典版（0/301/500），**不是 RuoYi-Vue 的 200/500/601**（调研实锤，tech-stack 第四节第 1 条）；
- **Redis 访问一律走 RedisCache 门面**，禁止业务代码直接摸 predis 客户端；
- **时间**：think-orm `datetime_format` 设为 `Y-m-d H:i:s`；
- **配置与代码分离**：.env 存连接信息。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录

- 2026-09-29 全部 6 个 Task 完成（PHPUnit 10 tests 39 assertions 全绿；RedisCache 真库端到端通过）。
- **Task 3 决策点（单应用 vs 多应用）**：定**单应用**（controller 子目录分模块：`controller/system/` 等），本系统无多终端场景，简单优先；TP 8.1.4 实测 `php think run` 正常。
- **版本回填**（tech-stack.md 第二节提醒项）：ThinkPHP **8.1.4**（framework ^8.0 实装）；think-orm ^3.0|^4.0 由 composer 解析；predis v2；PHPUnit 11.5。
- 环境踩坑（已记 AGENTS.md）：① System32 旧版 VCRUNTIME140.dll 致 PHP 静默失败，从 Python314 复制新版 DLL 到 C:\php\8.3 解决；② 阿里云 composer 镜像的 think-trace 2.0.x-dev 与 minimum-stability 冲突，改官方 packagist；③ think Response/Json 构造器需容器 Cookie 注入，信封类改工厂模式 `Response::create($body,'json')`；④ predis scan 返回 [cursor(string), keys[]] 二元组，cursor '0' 终止；⑤ wencun 建库权限不足（spec 预判命中），docker root 建库授权。
- 常量类命名 `TpConstant`（不是 tasks.md 初稿的 constant.php）——常量以类组织更规范，属实现细节非偏差。
- RedisCache 落位 `app/common/RedisCache.php`（不是 spec 目录树的 app/facade/）——避免 TP facade 机制混淆，普通静态门面即可，属实现细节非偏差。
