# AGENTS.md（RuoYi-TP）

本文件是 TP 版子目录的 AI 编码规范入口。工作区级规范以根目录 [../AGENTS.md](../AGENTS.md) 为唯一规范源；任务台账在 [specs/README.md](specs/README.md)，选型依据 [specs/tech-stack.md](specs/tech-stack.md)，差异登记 [specs/deviations.md](specs/deviations.md)。

## 项目定位

- **前后端不分离**（服务端渲染），复刻**经典若依 4.8.3**（只读源码参照 `reference/RuoYi-classic/`，gitignore），学习 PHP / ThinkPHP。
- **自用库自用缓存**：MySQL `ry-tp`（经典版官方 SQL 原样导入）+ Redis db1，与工作区三版（8080 互斥的 Java/Python/Go）完全隔离。
- 页面保真是验收目标：登录页、主框架、各管理页的观感与交互对齐经典版。

## 技术栈（已定，依据 specs/tech-stack.md）

ThinkPHP 8.1（think-orm 3 + think-view）+ predis v2 + PHPUnit 11；密码 `md5(loginName + password + salt)` 原样复刻经典版（已实测对上官方 SQL 预置 hash）；会话 cookie(uuid) + Redis（30 分钟空闲超时）；PHP 8.3（C:\php\8.3，已入用户 PATH）。

## 分层纪律

```
app/
├── controller/   # 参数接收与校验、调 service、返回 AjaxResult/TableDataInfo 或模板渲染
│                 #   禁止直接操作 db/redis
├── service/      # 业务逻辑、事务边界、缓存维护；禁止拼查询
├── model/        # think-orm 表模型，仅数据访问（对位 dao）；禁止业务判断和缓存操作
├── validate/     # TP 验证器
├── middleware/   # 登录态/权限/操作日志/防重（1.0.0 起）
├── common/       # TpConstant（键前缀与业务常量）/ AjaxResult（code 0/301/500）/
│                 #   TableDataInfo（{code:0,msg,rows,total}）/ PasswordService / BusinessException
├── view/         # ThinkTemplate 模板（1.5.0 起，经典版页面逐页改写）
└── facade/ 逻辑   # RedisCache（app/common/RedisCache.php）：Redis 访问唯一入口
```

- **响应信封**：业务接口一律 `AjaxResult::success/warn/error`（code **0/301/500**，对位经典版 `AjaxResult.Type`，**不是 RuoYi-Vue 的 200/500/601**）；分页用 `TableDataInfo::of($rows, $total)`（`{code:0, msg, rows, total}`）。
- **Redis 访问**一律走 `RedisCache` 门面（键拼装/JSON 编解码/SCAN 收口），业务代码禁止直接用 predis 客户端；键前缀常量集中在 `TpConstant`。
- **业务失败**：service 抛 `BusinessException`，由 `app/ExceptionHandle.php` 统一渲染 `AjaxResult::error`（HTTP 200 + code 500）。
- **时间**：think-orm `datetime_format` 全局 `Y-m-d H:i:s`。
- **契约对照**：动工前读 `reference/RuoYi-classic` 对应 Controller + ServiceImpl + 模板 HTML + 页面 JS（每模块动工检查单见 specs/README.md）。

## 常用命令

本机 `C:\php\8.3` 已入 PATH（`php -v` 应为 8.3.35；若命令找不到或版本不符，一律改用全路径 `"C:\php\8.3\php.exe"`）：

```bash
php think run --port 8888           # 开发服务器（8888 与 8080 三版并行）
php vendor/bin/phpunit              # 单元测试（纯逻辑，不连库）
php think                           # 查看可用命令（含 10.0.0 的 scheduler）
```

Composer 用 `php C:\php\8.3\composer.phar <cmd>`（全局 PATH 的 composer.bat 在新终端生效）。

## 环境约束

1. **端口**：本版 8888；Java/Python/Go 三版在 8080 互斥运行，互不影响。
2. **数据隔离**：MySQL `ry-tp` + Redis db1 自用；测试数据测完清理（自己的库自己管）。建库/授权曾用 docker root（wencun 已获 ry-tp 全权）。
3. **参考源码只读**：`reference/` 已 gitignore，禁止改其中的文件。
4. **本机 PHP 坑**：System32 旧版 VCRUNTIME140.dll 会导致 PHP 静默不输出（已通过把新版 DLL 复制进 `C:\php\8.3\` 解决，勿删该 DLL）；CLI 中文输出乱码时 `chcp 65001`。
