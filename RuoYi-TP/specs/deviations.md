# Deviations · 有意差异登记（RuoYi-TP）

> 目的：本项目复刻**经典若依 4.8.3**（单一基准，源码在 [../reference/RuoYi-classic/](../reference/RuoYi-classic/)，库与 Redis 自用不共享）。凡无法对齐基准的行为必须登记在此，禁止只留在对话里。
> 与 Go/Python 版 deviations 的区别：本项目**无前端兼容契约**（前端就是 PHP 模板，静态资源同源移植），允许行为差异，但逐条登记保总验收可查。
> 格式：差异点 / 经典若依行为 / TP 版行为 / 原因。

| # | 差异点 | 经典若依行为 | TP 版行为 | 原因 |
|---|--------|-----------|-----------|------|
| 1 | 会话存储 | Shiro ehcache（本地内存）+ 同步到内存缓存，JSESSIONID cookie | cookie(uuid) + Redis 会话（30 分钟**空闲**超时语义对齐） | PHP built-in server 无跨请求内存缓存设施；Redis 即可用且顺带支持「在线用户」页面数据源 |
| 2 | 密码重试计数 | ehcache loginRecordCache（tti 10 分钟） | Redis 键 + 10 分钟过期 | 同上，行为语义一致（5 次锁 10 分钟） |
| 3 | 验证码存储 | Session attribute（`KAPTCHA_SESSION_KEY`） | 存自建会话（Redis）同名键语义 | 会话设施换成 Redis 的连带差异，登录校验逻辑一致 |
| 4 | demo 演示模块 | DemoTable/Form/Dialog/Icon/Operate/Report 六组演示控制器与页面 | **有意不复刻** | 纯演示性质（bootstrap-table 各玩法展示），无业务价值；demo 菜单随官方 SQL 导入，点击 404 属有意行为（或 1.5.0 移植时过滤 demo 菜单 SQL，实施时定并回填） |
| 5 | tool 模块 | BuildController（表单构建）/ SwaggerController（跳转外部文档）/ TestController（swagger 演示 CRUD） | **有意不复刻**（swagger 文档可后期评估 think-swagger） | 依赖 swagger/velocity 生态；前端表单构建页为独立 JS 工具，与后台业务无关 |
| 6 | Druid 监控 | DruidController 跳转 druid 监控页（数据库连接池监控） | **有意不复刻** | PHP 无 Druid 连接池；数据库监控用 `SHOW STATUS`/`SHOW PROCESSLIST` 做等价页面（9.0.0），菜单入口映射到该页 |
| 7 | 同账号互踢（KickoutSessionFilter） | maxSession 配置支持互踢（默认 -1 不限制，即默认不启用） | 首期不实现 | 默认配置即不启用；login.js 的 ?kickout=1 提示逻辑保留，无后端触发时自然不生效；后期按需补 |
| 8 | CSRF 校验 | CsrfValidateFilter 挂主链但 yml `csrf.enabled: false`，实际关闭 | 同样不启用 | 对齐经典版实际行为（机制存在但默认关闭）；后期开启需自产 token 注入模板 |
| 9 | 密码加密算法 | md5(loginName + password + salt)（**已实测复现**） | 原样复刻 `md5()` | 学经典版本来的方案（自用库无互验需求）；salt = 6 位 hex `bin2hex(random_bytes(3))` |
| 10 | 调度器 | Quartz（qrtz_* 表随库启动，支持并发策略/misfire 补跑） | `php think scheduler` 自研循环（cron-expression），qrtz_* 表不导入 | Windows/PHP 无 Quartz；sys_job 表的增删改查与页面照常复刻；并发/补跑策略无对应物 |
| 11 | 服务监控 | oshi 取 CPU/内存/JVM | **降级：部分指标或平台提示**（P4 定方案） | PHP/Windows 系统指标获取不可靠；JVM 无对应物 |
| 12 | 定时任务目标字符串 | Java 反射 `ryTask.ryParams('ry')` | 注册表映射 + Java 语法参数解析器（对位 Python/Go 版方案） | 无 Java 反射语义 |
| 13 | 操作日志异步 | AsyncManager（队列线程） | 同步写（built-in server 无 async 设施）| 日志量小可接受；换 php-fpm 部署时可改 fastcgi_finish_request |
| 14 | 文件上传路径 | `RuoYiConfig.profile`（yml 配置）→ 磁盘目录 + /profile/** 静态映射 | config/profile.php 同语义配置项；**8.0.0 落地为 `public/profile`（web server 静态直服，零读取端点），头像子目录 avatar/{Y/m/d}/{uuid}.{ext}** | 本机路径本就不同 |
| 15 | i18n 菜单 | jquery.i18n.properties + static/i18n（未完全启用，页面基本硬编码中文） | 不引入 i18n 设施，文案直接中文 | 经典版实际未启用；引入徒增复杂度 |
| 16 | 数据库/Redis 隔离 | 经典版设计为独立部署、独立库 | **自用库 `ry-tp` + Redis db1**（用户拍板 2026-09-28） | 与工作区三版完全隔离；无共享 schema 约束，表结构随经典版官方 SQL |
| 17 | rememberMe 机制 | Shiro CustomCookieRememberMeManager（rememberMe cookie 加密存用户凭证，30 天；yml 默认开启） | rememberMe 勾选时仅延长 cookie 有效期（30 天），登录态仍以 Redis 会话为准（30 分钟空闲超时**不清除**——即经典版 rememberMe 可免输密码自动登录，TP 版做不到免登录） | Shiro rememberMe 是独立的凭证记忆机制；TP 版完整复刻需加密 cookie 存凭证，首期从简：cookie 30 天内免重新输验证码外的完整登录。**动工时再评估是否完整复刻，结论回填 2.0.0 实施记录** |
| 18 | logout 跳转 | LogoutFilter 支持 redirectUrl/backUrl 逻辑（登出后回跳） | 清会话后固定跳 /login | 经典版 backUrl 依赖 request 参数场景少；前端 logout 调用点核对后再定是否补 |
| 19 | Excel 导出列头样式（3.0.0） | POI ExcelUtil 列头灰底居中 + 特定字体字号 | phpspreadsheet 从简复刻：加粗 + 灰底（EEEEEE），不追字号/边框/对齐细节；数据区列宽按列头长度估算 | 样式细节对功能零影响；phpspreadsheet 全量复刻 POI 样式成本高收益低；文件内容（5 列列头/数字格式/状态转换）与经典版一致 |
| 20 | 导入用户初始密码可登录（4.0.0，实测确认） | importUser 存 md5(loginName+初始密码) 不写 salt；登录时 Java 对 null 做 "null" 串接 → 哈希不匹配，**导入用户实际无法登录**（经典 bug） | PHP null 拼接为空串 → md5(loginName+密码+'') 恰好匹配，导入用户**可用初始密码登录**（2026-09-29 impuser1/123456 实测登录成功） | 顺带修正经典 bug 且符合 sys.user.initPassword 设计意图；TP 版无 salt 列语义包袱 |
| 21 | 角色/授权变更后在线用户权限生效时机（4.0.0/5.0.0） | clearAllCachedAuthorizationInfo 清 Shiro 缓存，被改用户下次鉴权即生效 | 会话 permissions 登录时算好存 Redis；改用户角色/insertAuthRole 后该用户需**重新登录**生效 | 无统一权限缓存设施可清；重登录语义对前端无感；5.0.0 复核是否做目标会话权限重算 |
| 22 | 改密后登录态（8.0.0） | resetPwd 成功后 setSysUser 刷新会话用户数据，**登录态保持、不踢出** | **销毁当前 Redis 会话 + 清 cookie 强制重新登录**，前端提示「密码修改成功，请重新登录」跳 /logout（profile 双 tab 与 resetPwd 弹窗两处回调微调） | 任务指明的安全加固要求（2026-09-29 写实任务明示）；改密后旧会话权限/资料快照失效问题随会话重建不存在 |
| 23 | profile 页用户数据源（8.0.0） | 会话快照（getSysUser，Shiro Principal 存全量） | 查 DB（selectUserById——TP 会话刻意不含 phone/email/sex） | 实现简化；行为差异为「他人代改资料时显示实时值」，无害 |
| 24 | 缓存监控对象（9.0.0） | ehcache 七缓存名（sys-cache/sys-config/sys-dict/sys-authCache/sys-userCache/loginRecordCache/shiro-activeSessionCache） | Redis 六前缀枚举（session:/pwd_retry:/repeat_submit:/rate_limit:/config:/dict:；验证码在会话内无独立键）；clearAll 排除 session:（clearCacheName 对 session 前缀也拒绝——防全站踢线加固） | 数据源换代的自然差异（ehcache→RedisCache）；7 条路由契约保持 |
| 25 | 在线用户存储（9.0.0） | OnlineSession 同步 sys_user_online 表 + 表查询 | Redis `session:*` 实时 SCAN，sys_user_online 表不读写（表随官方 SQL 存在但闲置） | Redis 会话即真源（#1 延伸），无同步延迟 |
| 26 | 在线用户导出（9.0.0） | 前端 showExport=true 但后端无 export 端点（点击 404，经典版原样缺陷） | **模板去掉 showExport**（9.0.0 动工拍板，不复刻 404 缺陷） | 功能缺陷不复刻 |
| 27 | 服务监控指标（9.0.0） | oshi 全量（CPU 使用率分项/JVM/磁盘多分区） | 三档降级（可得/条件可得/「—」）；JVM 区块改「PHP 运行环境」；零 exec 零扩展 | Windows exec 不可靠 + composer 监控包兼容性差；布局与区块结构保持 |
| 28 | 数据库监控（9.0.0） | DruidController 302 跳 druid 外部页 | 自研 SHOW STATUS/PROCESSLIST 简页（/monitor/data 内置渲染） | #6 兑现；菜单入口行为从「跳外部」变「内置页」 |
| 29 | 会话元数据来源（9.0.0） | OnlineSession 字段由 Shiro 维护（host/browser/os/lastAccessTime） | 登录时写入会话平铺键（ip/browser/os/loginTime/lastAccessTime）+ LoginAuth touch 刷新；旧会话/外来会话显示「未知」并禁强退 | 会话结构自管的连带实现；行为等价 |
| 30 | cron 库选型（10.0.0） | Quartz CronExpression（6/7 位秒级、年字段、周 1=SUN、misfire 四策略） | **appserver-io/microcron 2.0**（spec 原选 dragonmantank v3 实测不支持秒级——5 位分钟制，动工实测后用户拍板换库）+ CronService `?`→`*` 转换层 | microcron 支持秒级/年字段/`n/s` 步进；周编号 1=MON…7=SUN Linux 惯例与 Quartz 错位一天；misfire/concurrent 策略字段照存无调度语义（不补跑/天然串行）——#10 落地说明 |
| 31 | 代码生成器模板链（11.0.0） | GenController 16 路由全量：preview / download / genCode / batchGenCode / createTable / synchDb **+ GET edit 编辑页渲染**；Velocity 渲染 10+ 模板产出 Java/Vue/thymeleaf 代码（zip 下载 / 自定义路径写盘 / Druid 解析 SQL 建表 / information_schema 三方对账） | **有意排除，路由不注册即 404**；数据层 8 路由照常（GET /tool/gen、POST list、POST db/list、POST column/list、GET+POST importTable、POST edit 保存、POST remove）。列表页与导入弹窗照常可用；页面上「预览/编辑/同步/生成代码/批量生成/创建」按钮点击 **404 或 loading 遮罩不消失**属有意行为（2026-10-01 浏览器实测确认预览 404） | ① 经典版生成器产出 Java/Vue 代码，对 TP 项目零产出价值；② TP 版若翻译模板只能产 TP 代码，**无「逐字节对齐」验收锚点**；③ 本版**前后端不分离**，连 Python 版的「api.js + vue 生成物」价值都没有，降级理由比 GO 版 #19 更充分；④ 模块收官在望，「反复生成新模块」场景几乎不出现，学习价值集中在读表/导入/同步的数据层。**范围经用户 2026-10-01 拍板确认（选项 A）** |
| 32 | gen 库表列表行键（11.0.0） | selectDbTableList 经 GenTable resultMap 映射 → Jackson 序列化**整个实体**：除 tableName/tableComment/createTime/updateTime 外，tableId/className/genPath… 等字段以 `null` 一并输出（`params:{}`、`columns:[]`） | db/list 行**只输出** tableName / tableComment / createTime / updateTime 四键（驼峰） | 导入弹窗只消费这四列（importTable.html 的 columns 定义实锤）；补齐 20+ 个恒为 null 的键对页面零影响，反而把「哪些字段真有值」淹没。列表侧（gen_table 21 列）则**完整输出**、不做裁剪 |
| 33 | remove 空 ids 边界（11.0.0） | `Convert.toLongArray("")` → 空数组 → Mapper foreach 拼出 `delete from gen_table where table_id in ()` → **SQL 语法错误**（HTTP 200 + code 500 报错） | 空 ids 视为**无操作**，仍返回 `{"code":0,"msg":"操作成功"}`（不发起 SQL） | 经典版该分支是「前端保证不会发空 ids」前提下的意外缺陷；TP 版按防御式处理避免无意义 500。**不存在 id 的语义照抄经典版**：固定 success、不查影响行数（2026-10-01 curl 实测 ids=999999 返回 code 0） |

## 登记规则

1. 实现时发现无法对齐经典版的行为 → 当场登记（原因 + 影响面）。
2. 本项目允许行为差异（无外部前端契约），但**页面保真**是目标：登录页、主框架、各管理页的观感与交互应与经典版一致。
3. 每条差异须在对应模块 checklist 中体现核对项；总验收按本表逐条过。
