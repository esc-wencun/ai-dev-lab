# RuoYi-TP（ThinkPHP 版若依）技术选型调研与结论

> 日期：2026-09-28（同日经源码调研修订）　|　性质：**纯调研与选型结论，未开发**
> 定位：工作区**第五子项目**，**前后端不分离**（服务端渲染），复刻**经典若依 4.8.3**（gitee `y_project/RuoYi` master，Spring Boot + Shiro + Thymeleaf + Bootstrap）的功能与界面，学习 PHP。
> 已拍板决策（2026-09-28 用户确认）：① UI **移植经典若依 Bootstrap/jQuery 静态资源**；② PHP **手动 zip + Composer** 安装；③ 端口 **8888**；④ **PHP 版用自己的数据库和 Redis**（2026-09-28 修订时确认，见第一节）。

---

## 一、定位声明（2026-09-28 源码调研后修订）

- **复刻基准 = 经典若依 4.8.3 单一来源**。源码已下载到 [../reference/RuoYi-classic/](../reference/RuoYi-classic/)（gitignore，仅作只读参照），本 spec 所有契约以它为准，不再凭记忆。
- **PHP 版用自己的库和 Redis，不与工作区三版共享**（用户拍板）：
  - MySQL：同机 Docker 实例上新建库 **`ry-tp`**，直接导入经典版官方 SQL（`sql/ry_20260319.sql`）——表结构与经典版逐字段一致（`sys_menu.url/target/is_refresh`、`sys_user.login_name/salt` 全都在），**零 schema 适配**；
  - Redis：同机 Docker 实例上用 **db 1**（工作区三版用 db 0），键名按 TP 版自定常量，不要求与 Java `CacheConstants` 一致；
  - 三版数据/会话与 TP 版完全隔离，互不干扰；「共用库清理纪律」对 TP 版简化为「自己的库自己管」。
  - > 待确认假设：MySQL/Redis 复用本机 Docker 实例（端口 3306/6379 不变），只是库名与 db 索引分开。若你要彻底独立（另一套容器），改 config 即可，不影响架构。**`wencun` 账号可能无建库权限，P0 时若失败改用 docker 内 root 或提前授权（0.0.0 Task 4 注明）。**
- **不参与「前端零改动切换」契约**——PHP 自己渲染页面，RuoYi-Vue3 / RuoYi-React 无法驱动它。
- 学习价值定位：看「同一套经典 RBAC 后台，PHP/ThinkPHP 生态怎么实现」；与 FastApi/Go 版（接口复刻 RuoYi-Vue）是**两条不同的复刻线**，互不依赖。

## 二、结论一览（推荐组合）

| 层次 | 经典若依对位 | **TP 版选型** | 说明 |
|---|---|---|---|
| 语言/运行时 | Java 17 + Spring Boot 3（jakarta 包名，4.8.3 实况） | **PHP 8.3 NTS x64**（Windows zip 手动装） | 8.3 为当前广泛部署稳定版 |
| Web 框架 | Spring Boot + Shiro | **ThinkPHP 8**（topthink/framework 8.x） | 国内 PHP 管理系统事实标准（调研见第三节） |
| 模板引擎 | Thymeleaf | **ThinkTemplate**（topthink/think-view） | layout 继承 / include，Thymeleaf 片段近机械对译 |
| ORM | MyBatis | **think-orm 3** | 模型 + 查询构造器；无 Mapper XML 层 |
| Redis 客户端 | Shiro ehcache（本地缓存） | **predis v2**（纯 PHP）+ 自建 RedisCache 门面 | 会话、验证码、防重全部走它；免装 dll，Windows 零折腾 |
| **密码方案** | `Md5Utils.hash(loginName + password + salt)` | **`md5($loginName . $password . $salt)`**（原样复刻） | **2026-09-28 已实测验证**：`md5('admin'.'admin123'.'111111')` = `29c67a30...` 与官方 SQL 完全一致；salt = 6 位 hex（`bin2hex(random_bytes(3))` 对位 `SecureRandomNumberGenerator.nextBytes(3).toHex()`） |
| 会话 | Shiro Session（ehcache + 30 分钟空闲超时） | **cookie(uuid) + Redis 会话**（30 分钟空闲超时，语义对齐） | 不分离模式正统做法；不用 JWT |
| 验证码 | kaptcha（math/char 双型，存 Session） | **PHP GD 自绘**（math 型，存自建会话） | 开关/类型来自**配置文件**（对位 yml `shiro.user.captchaEnabled/captchaType`；**经典版 sys_config 无 captcha 键**，见第四节第 3 条） |
| Excel 导入导出 | Apache POI（ExcelUtil） | **phpoffice/phpspreadsheet** | 需 intl / gd / zip 扩展 |
| cron 表达式解析 | Quartz cron | **dragonmantank/cron-expression v3** | Laravel 同款，PHP 事实标准 |
| 定时任务执行 | Quartz 调度器（ruoyi-quartz） | **`php think scheduler` 常驻命令**（自研循环 + cron-expression） | Windows 下 workerman 受限；qrtz_* 表不导入不触碰 |
| 分页 | PageHelper + TableDataInfo | 自研 TableDataInfo 组装 | 输出 `{code:0, msg, rows, total}`，**code 是 0 不是 200**（调研实锤，见第四节） |
| 依赖管理 | Maven | **Composer**（阿里云镜像） | |
| 单元测试 | — | **PHPUnit 11**（树构建 / cron 解析 / md5 密码方案 / 权限串匹配） | 涉库走真实环境端到端 |

一句话：**ThinkPHP 8 + think-orm + predis + GD + phpspreadsheet + cron-expression + PHPUnit**；密码 md5(login_name+password+salt) 原样复刻，会话 cookie + Redis，端口 8888。

## 三、Gitee 调研结论（2026-09-28 API 实查）

| 项目 | star（实测） | 技术栈 | 对本项目的参考价值 |
|---|---|---|---|
| [FastAdmin](https://gitee.com/karson/fastadmin) | 7020 | ThinkPHP 5/8 + Bootstrap，**不分离** | **同模式最直接参照**：TP + Bootstrap 后台、一键 CRUD 思路 |
| [Dcat Admin](https://gitee.com/jqhph/dcat-admin) | ~9k | Laravel，不分离 | Laravel 系后台代表，仅了解 |
| [CatchAdmin](https://gitee.com/jaguarjack/catchadmin) | 4496 | ThinkPHP + Vue，分离 | 验证 TP 做 RBAC 后台的完整路径 |
| [likeadmin-php](https://gitee.com/likeadmin/likeadmin_php) | ~2k | ThinkPHP + Vue3，分离 | 同上 |
| [webman](https://gitee.com/walkor/webman) | 966 | workerman 常驻内存 | 高性能方向，本项目不用（服务端渲染场景收益小） |

结论：**国内 PHP 管理系统 ThinkPHP 是绝对主力**；经典不分离模式（FastAdmin）仍活跃维护。本项目选 ThinkPHP 与主流一致。

## 四、源码调研实锤的关键契约（2026-09-28，逐条对照 reference 源码）

以下是**改动过错误认知或必须写死的事实**，后续模块 spec 的前置依据：

1. **code 体系是 `0/301/500`，不是 RuoYi-Vue 的 200/500/601**。经典 `AjaxResult.Type`：SUCCESS(0) / WARN(301) / ERROR(500)；`TableDataInfo.code = 0`；前端 `web_status = {SUCCESS: 0, FAIL: 500, WARNING: 301}`（ry-ui.js:1810）。另有一处特例：GET /login 被 ajax 调用时返回 `{"code":"1","msg":"未登录或登录超时。请重新登录"}`（code 是**字符串 "1"**，login 页特供）。
2. **登录 POST /login 参数**：`username` / `password` / `validateCode` / `rememberMe`——**没有 uuid**，验证码答案存在 Session（`Constants.KAPTCHA_SESSION_KEY`）。登录成功 JS 跳 `location.href = ctx + 'index'`。
3. **验证码端点**：`GET /captcha/captchaImage?type=math&s=<随机数>`（匿名可访问）；**开关与类型来自 yml**（`shiro.user.captchaEnabled: true`、`shiro.user.captchaType: math`），不是 sys_config——登录 GET 时 `CaptchaValidateFilter` 把这两个值塞进 request attribute（`captchaEnabled`/`captchaType`），模板据此渲染；POST /login 挂 `anon,captchaValidate` 过滤器做验证码校验（答案存 Session）。**注意：共享库 ry-vue 的 `sys.account.captchaEnabled` 是 RuoYi-Vue 的机制，经典版 sys_config 预置 11 条里没有这个键**。
4. **列表分页**：bootstrap-table 默认 **POST**（ry-ui.js initTable defaults `method: 'post'`, `sidePagination: "server"`）；参数 `pageSize` / `pageNum` / `searchValue` / `orderByColumn` / `isAsc` + 搜索表单字段（含 `params[beginTime]` / `params[endTime]`）。响应经 `responseHandler` 校验 `res.code == 0` 后取 `res.rows/res.total`。
5. **菜单 href 直接来自 DB 的 `url` 列**（index.html `th:href="@{${cmenu.url}}"`），经典库自带 `/system/user` 等值；树构建是 `getChildPerms(list, 0)` 递归；admin 全量（`menu_type in ('M','C') and visible='0'`）。**TP 版用经典库 SQL，url/target/is_refresh 原样可用，无推导逻辑**。
6. **权限的两条通道**：模板标签 `shiro:hasPermission="system:user:add"`（按钮渲染）；页面 JS 变量 `var editFlag = [[${@permission.hasPermi('system:user:edit')}]]`（表格操作列）。TP 对位：模板函数 `check_perm()` + `{assign}`。
7. **密码重试**：`user.password.maxRetryCount: 5`，锁定 10 分钟（ehcache loginRecordCache tti=10min）——TP 对位 Redis 键 + 10 分钟过期。
8. **会话**：`shiro.session.expireTime: 30` 分钟（空闲超时，Shiro touch 语义）；maxSession 默认 -1；**KickoutSessionFilter（同账号互踢）存在但按 maxSession 配置**——TP 版首期不实现互踢（deviations #7），login.js 的 kickout 参数提示逻辑随静态资源保留、无后端支持时自然不触发。
9. **静态资源规模**：`static/` 共 8.3MB / 90 个 JS（ajax/libs 下 bootstrap-table、layer、zTree、summernote、validate、blockUI 等）；模板 137 个 HTML。主框架交互由 `static/ruoyi/index.js`（iframe 标签页）+ `js/ry-ui.js`（弹窗/表格封装）+ `js/common.js`（ajax 全局配置）承担。
10. **CSRF**：经典版有 `CsrfValidateFilter` 且挂在主链上（`user,kickout,onlineSession,syncOnlineSession,csrfValidateFilter`），include.html 也输出 csrf-token meta，但 yml `csrf.enabled: false`——**实际关闭状态**。TP 版同样默认不启用（deviations #8），若后期开启需自产 token 注入模板。
11. **经典版 controller 里 Demo/Test/Build/Druid/Swagger 等非业务控制器**（演示、表单构建、druid 监控、swagger 跳转）：TP 版不复刻或降级（deviations #4/#5/#6）。
12. **主框架的配置驱动项**（SysIndexController 实锤）：`sys.index.skinName`（皮肤）/ `sys.index.sideTheme`（侧栏主题）/ `sys.index.footer`（页脚）/ `sys.index.tagsView`（页签）/ `sys.index.menuStyle`（default 左侧导航 / topnav 顶部导航，**topnav 时渲染 index-topnav 模板**；移动端 UA 强制走 index）；另有 `GET /system/switchSkin`（皮肤切换页）与 `GET /system/menuStyle/{style}`（写 nav-style cookie）两个端点。**登录密码策略提醒**：`sys.account.initPasswordModify`/`sys.account.passwordValidateDays`/`sys.account.chrtype` 在 main 页逻辑用（initPasswordIsModify/passwordIsExpiration），登录时弹「修改初始密码」提醒。
13. **锁屏**：`GET /lockscreen` + `POST /unlockscreen`（密码验证解锁）——经典版功能项，P3 排期（8.0.0 一并做或单独小项，实施时定）。
14. **注册**：`GET/POST /register`（开关 `sys.account.registerUser`，默认 false 关闭，页面入口隐藏）——低优先级，P3 评估是否复刻（Python/Go 版有对应能力，复刻成本小）。

## 五、UI 方案（移植经典若依静态资源，已拍板）

- **资源来源**：reference 源码包 `ruoyi-admin/src/main/resources/static/`（8.3MB）原样移植——**注意是把 static/ 的内容映射到站点根**（`public/js/`、`public/ajax/`、`public/css/`、`public/ruoyi/` 等，而不是 `public/static/`）：经典版模板全部以根路径引用（`@{/js/jquery.min.js}`、`@{/ajax/libs/...}`、`@{/ruoyi/js/ry-ui.js}`），放 public/static/ 会全部 404。保留各库版权头，不做美化改造；
- **模板改写**：`templates/` 137 个 Thymeleaf 页面 → ThinkTemplate（`th:each`→`{volist}`、`th:if`→`{if}`、`include :: header('xx')`→`{include file="include/header" title="xx"}`）；菜单树因经典模板嵌套四级，**改为 service 递归渲染 HTML 片段**更干净（1.5.0 定）；
- **主框架**（顶部导航 + 左侧菜单 + iframe 内容区 + 多标签页）由经典版现成 JS 承担，PHP 只按权限渲染菜单 HTML（href 用 DB url 原值）；
- **不重画界面**、不引入任何前端构建工具（无 package.json）。

## 六、Windows 环境要点（本机）

- **PHP zip 已下载**：`C:\php\php-8.3.zip`（windows.php.net 官方 latest 8.3 NTS x64，2026-09-28 实下 33.8MB）。解压后复制 `php.ini-development` → `php.ini`，启用扩展：`pdo_mysql` / `gd` / `openssl` / `fileinfo` / `mbstring` / `curl` / `zip` / `intl`（phpspreadsheet 需要 intl+gd+zip）。
- **Composer**：`composer.phar` 直装 + 自建 `composer.bat`；镜像 `composer config -g repos.packagist composer https://mirrors.aliyun.com/composer/`。
- **开发服务器**：`php think run --port 8888`。8888 与 8080 三版并行互不干扰。
- **控制台 GBK 坑**：PHP CLI 输出 UTF-8 中文到 GBK 终端会乱码（同 Python 版踩过的坑），必要时 `chcp 65001`。
- **调度器进程**：`php think scheduler` 需另开一个终端前台常驻，Windows 关窗即停——开发期可接受，登记 deviations。

## 七、分期路线

P0（0.0.0 工程基础：环境 + 骨架 + **建 ry-tp 库导入经典 SQL** + RedisCache 门面）→ P1（1.0.0 基础设施 + 1.5.0 静态资源与布局 + 2.0.0 登录与主框架，**里程碑 = admin/admin123 登进经典若依同款主界面**）→ P2（3.0.0~5.0.0 RBAC 核心）→ P3（6.0.0~8.0.0 字典/参数/公告/个人中心）→ P4（9.0.0 监控日志 + 10.0.0 定时任务）→ P5（11.0.0 代码生成器，可选）。模块表见 [README.md](README.md)。

## 附：版本核对提醒

ThinkPHP 具体小版本（8.1.x 系）与 think-orm / think-view 的兼容矩阵以 composer 安装时实际解析为准，0.0.0 装完后**回填第二节选型表**；PHP 8.3 与 phpspreadsheet 等包的兼容性如有出入，以 composer 报错为准并回填。
