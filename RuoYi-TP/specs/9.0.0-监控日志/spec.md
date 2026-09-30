# 9.0.0-监控日志 · spec

> **状态：✅ 已完成（2026-10-01 联调验收通过）**
> 对位经典若依 SysUserOnlineController（5 端点）+ ServerController（1）+ CacheController（7）+ SysOperlogController（6）+ SysLogininforController（6）——**25 端点实核**；其中 DruidController（/monitor/data → druid 跳转）不复刻（deviations #6，菜单 /monitor/data 映射自研「数据库监控」页）。页面 7 个（monitor/online/cache/server/operlog×2/logininfor），逐方法核对实锤。
> 依赖：1.0.0（#[Perm] / #[Log] / PageQuery / TableDataInfo / ExcelExportService / /common/download）、2.0.0（SessionService 会话设施 / LoginService.recordLogininfor）；页面渲染于 1.5.0 主框架 iframe 内。
> 调研依据：reference 源码逐文件核对（五个 Controller / SysUserOnlineServiceImpl / SysUserOnlineMapper.xml（**sys_user_online 表在 TP 版不用**）/ OnlineSessionDAO.syncToDb·doDelete / CacheService + CacheUtils（ehcache JCache 体系）/ ehcache-shiro.xml 七个缓存名 / SysOperLogMapper / SysLogininforMapper / ServerController + server.html 四区块 / 页面 HTML×7 / ry-ui.js removeAll·clean·post / common.js beginOfTime·endOfTime）+ **ry-tp 库实测**（sys_oper_log 17 列 / sys_logininfor 9 列 / sys_menu monitor:* 权限串 24 行 / sys_oper_type、sys_common_status 字典 15 行）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、POST 分页参数（params[beginTime]/[endTime]）、TableDataInfo、权限两通道。**Redis 键前缀以 app/common/TpConstant.php 实际常量为准**（session:/pwd_retry:/repeat_submit:/rate_limit:/config:/dict:；验证码无独立键——存会话内 captcha 字段，TpConstant 注释实锤）。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/monitor/OnlineController.php | SysUserOnlineController | 3 方法 3 路由（online 页 / list / batchForceLogout 单批合一）；**数据源 = Redis db1 `session:*` SCAN + JSON 解析**，sys_user_online 表不读写 |
| app/controller/monitor/ServerController.php | ServerController | 1 方法；**oshi 降级**（deviations #11 定稿方案见特殊行为 6） |
| app/controller/monitor/CacheController.php | CacheController | 7 方法 7 路由；对位对象从 ehcache JCache 换成 Redis 前缀枚举（见特殊行为 4） |
| app/controller/monitor/OperlogController.php | SysOperlogController | 6 方法 6 路由（list/export/remove/detail/clean） |
| app/controller/monitor/LogininforController.php | SysLogininforController | 6 方法 6 路由（list/export/remove/clean/unlock） |
| app/service/OnlineService.php | ISysUserOnlineService（仅取其语义） | SCAN 会话解析 / 过滤 / 强退删除；无 DB mapper |
| app/service/MonitorCacheService.php | CacheService | 前缀枚举 → 键列表 → 键值 → 清理（全走 RedisCache 门面） |
| app/service/OperLogService.php + LogininforService.php | ISysOperLogService / ISysLogininforService | 列表查询（like+in+时间范围）/ 批量物理删 / truncate |
| app/service/ServerInfoService.php | Server（oshi） | PHP 可得指标采集（详见特殊行为 6），不可得项置 null 显示「—」 |
| app/view/monitor/online/online.html、cache/cache.html、server/server.html、operlog/{operlog,detail}.html、logininfor/logininfor.html | templates/monitor/ 同名 7 页 | 7 页，静态 JS 零改动对接（bootstrap-table / bootstrap-select / jsonview / ibox 布局） |

**范围外**：定时任务监控（10.0.0）；/monitor/data 数据库监控页（本模块范围**内**做成 `SHOW STATUS`/`SHOW PROCESSLIST` 简页——deviations #6 承诺项，见特殊行为 7）；demo/tool 菜单（有意 404）。

## 页面清单（7 + 数据库监控 1 = 8）

| # | 页面 | TP 模板路径 | 对位经典版 | 组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 在线用户 | view/monitor/online/online.html | online/online.html | bootstrap-table | uniqueId=sessionId；url=prefix+/list（POST 分页）；sortName=lastAccessTime desc；showExport=true；escape=true；列：checkbox/序号/sessionId(tooltip)/loginName(sortable)/deptName/ipaddr/loginLocation/browser/os/status（on_line→「在线」primary、off_line→「离线」danger 徽章）/startTimestamp(sortable 登录时间)/lastAccessTime(sortable 最后访问)/操作（强退，forceFlag=monitor:online:forceLogout 显隐）；单条强退 confirm「确定要强制选中用户下线吗？」→ `$.operate.post(prefix+"/batchForceLogout", {ids: sessionId})`；批量 → ids=selectColumns("sessionId").join()（confirm「确认要强退选中的N条数据吗?」）；搜索字段 ipaddr/loginName |
| 2 | 缓存监控 | view/monitor/cache/cache.html | cache/cache.html | 三栏 ibox（缓存列表/键名列表/缓存内容） | getCacheNames() → POST prefix+/getNames **返回 HTML 片段**塞 #cacheNames；getCacheKeys(cacheName) → POST /getKeys 返回 HTML 片段塞 #cacheKeys；getCacheValue → POST /getValue 返回 HTML 片段塞 #cacheValue；clearCacheName/clearCacheKey → `$.post` JSON {code} → 成功 msg + 刷新键名列表；clearAll → **GET** /clearAll；「清理全部」在内容栏刷新钮 |
| 3 | 服务监控 | view/monitor/server/server.html | server/server.html | 四区块 ibox（CPU/内存/服务器信息/JVM→**运行环境**/磁盘） | 纯服务端渲染无 ajax；CPU 四行（核心数/用户/系统/空闲%）；内存表（系统内存+**PHP memory** 双列对位 内存+JVM）；服务器信息（名称/IP/OS/架构）；磁盘 tr volist（盘符/文件系统/类型/总/可用/已用/用量%，>80 红字） |
| 4 | 操作日志 | view/monitor/operlog/operlog.html | operlog/operlog.html | bootstrap-table + **bootstrap-select**（多选下拉） | cleanUrl=prefix+/clean、detailUrl=prefix+"/detail/{id}"、removeUrl/exportUrl；sortName=operTime desc；showPageGo/rememberSelected；queryParams 覆写：search.params={beginTime: beginOfTime(startTime), endTime: endOfTime(endTime)} + businessTypes=selectpicker('val') join；列：operId/title(tooltip)/businessType（sys_oper_type 字典徽章 selectDictLabel）/operName(sortable)/deptName/operIp/operLocation/status（0成功/1失败徽章）/operTime(sortable)/costTime（「%s毫秒」sprintf）/操作（详细，detailFlag）；工具栏：删除 removeAll / 清空 $.operate.clean（confirm「确定清空所有操作日志吗？」）/ 导出；搜索：operIp/title/operName/businessTypes(多选)/status(sys_common_status 下拉)/时间范围 |
| 5 | 操作日志详情 | view/monitor/operlog/detail.html | operlog/detail.html | 卡片式详情 + **JSONView** | 服务端渲染基本信息（title/业务类型 dict.getLabel/operTime/状态 tag）+操作人员+请求信息（method 彩签+operUrl/method 全名/costTime）+异常卡片（status!=0 才渲染，errorMsg）；JS：operParam/jsonResult 长度<2000 走 JSONView 否则 text（「（无参数）」「（无返回数据）」）；复制按钮 clipboard |
| 6 | 登录日志 | view/monitor/logininfor/logininfor.html | logininfor/logininfor.html | bootstrap-table | cleanUrl/removeUrl/exportUrl；sortName=loginTime desc；showPageGo/rememberSelected；queryParams 同上（无 businessTypes）；列：infoId/loginName(sortable)/ipaddr(tooltip)/loginLocation/browser/os/status（sys_common_status 字典徽章）/msg/loginTime(sortable)；工具栏：删除/清空/**解锁**（unlock() → `$.operate.post(prefix+"/unlock?loginName=" + selectColumns("loginName"))`——**loginName 走 query string**，选中列去重）/导出；搜索：ipaddr/loginName/status/时间范围 |
| 7 | 数据库监控 | view/monitor/data/data.html | DruidController → druid 外部页 | **自研简页**（deviations #6）：MySQL `SHOW STATUS`（连接数/线程/QPS 相关项）+ `SHOW PROCESSLIST` 表格 + 版本/字符集；只读无操作；菜单 /monitor/data 映射此页（monitor:data:view） |

> 页面 1/4/6 为 bootstrap-table 分页表（POST + server 端分页）；页面 2/3/5/7 无表格插件（ibox 布局 / 详情卡片 / 纯服务端渲染）。

## 端点级 API 清单（online 3 + server 1 + cache 7 + operlog 6 + logininfor 6 + data 1 = 24 条路由）

### 在线用户 /monitor/online（对位 SysUserOnlineController；batchForceLogout 单批合一）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /monitor/online | monitor:online:view | 无 | — | 渲染页面 1 |
| 2 | POST | /monitor/online/list | monitor:online:list | 无 | pageNum/pageSize/orderByColumn(isAsc)（**白名单 login_name/start_timestamp/last_access_time**）+ ipaddr(like)/loginName(like) | **TableDataInfo**；数据源：`RedisCache::keysScan('session:*')` → 逐键 get 解析 JSON → 过滤**含 loginName 键的会话**（无 loginName = 匿名会话/外来格式，见特殊行为 1）→ ipaddr/loginName like 过滤 → 排序 → 手工分页；行字段**驼峰**：sessionId（uuid）/loginName/deptName/ipaddr/loginLocation/browser/os/status='on_line' 恒定/startTimestamp/lastAccessTime（Redis 会话无这些原生字段——**落会话时补写元数据**，见特殊行为 2）；expireTime 不出（前端列未用） |
| 3 | POST | /monitor/online/batchForceLogout | monitor:online:batchForceLogout **或** monitor:online:forceLogout（OR 语义，Perm 注解 IS_REPEATABLE 双值） | 在线用户, 7强退 | ids（逗号串 sessionId） | 逐个：键不存在 → error(500)「用户已下线」（经典版两处 null 检查合并语义）；**sessionId == 当前会话 uuid → error(500)「当前登录用户无法强退」**（自踢防护，经典版实锤）；**解析结果为「未知」（外来会话，见特殊行为 1）→ error(500)「该会话无法识别，禁止强退」并中止整批**；通过 → `RedisCache::delete('session:'.$uuid, raw)`；整批成功 → success()「操作成功」；无 sys-userCache 队列维护（TP 无 kickout 设施，removeUserCache 无对应物） |

### 服务监控 /monitor/server

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 4 | GET | /monitor/server | monitor:server:view | 无 | — | 渲染页面 3；ServerInfoService 采集（降级方案见特殊行为 6） |

### 缓存监控 /monitor/cache（对位 CacheController；#5~#7 返回 **HTML 片段**非 JSON）

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 5 | GET | /monitor/cache | monitor:cache:view | 无 | — | 渲染页面 2，变量 cacheNames=前缀枚举 |
| 6 | POST | /monitor/cache/getNames | monitor:cache:view | 无 | — | **HTML 片段**（tr 列表，对位 fragment-cache-names）：前缀枚举逐行（序号/名称/清空钮） |
| 7 | POST | /monitor/cache/getKeys | monitor:cache:view | 无 | cacheName（前缀名） | **HTML 片段**（对位 fragment-cache-kyes——经典版拼写原样）：`keysScan(prefix.'*')` 键列表；cacheName 空 → 扫全库**已知前缀合并**（对位经典版 getCacheKeys('') 刷新钮） |
| 8 | POST | /monitor/cache/getValue | monitor:cache:view | 无 | cacheName、cacheKey（**完整键**） | **HTML 片段**（三输入框+textarea）：`RedisCache::get($cacheKey, raw:true)` JSON 解码 → pretty print 字符串；键不存在 → 「（不存在或已过期）」 |
| 9 | POST | /monitor/cache/clearCacheName | monitor:cache:view | 无 | cacheName | `keysScan(prefix.'*')` 逐键 delete；返回 AjaxResult success（JSON） |
| 10 | POST | /monitor/cache/clearCacheKey | monitor:cache:view | 无 | cacheName、cacheKey | delete 完整键；JSON success |
| 11 | GET | /monitor/cache/clearAll | monitor:cache:view | 无 | — | 前缀枚举全部逐键清（**session: 前缀默认排除**——清了会踢掉所有登录用户含自己，经典版 CacheService 有 SYS_AUTH_CACHE 排除先例；页面按钮语义=清业务缓存）；JSON success |

### 操作日志 /monitor/operlog（对位 SysOperlogController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 12 | GET | /monitor/operlog | monitor:operlog:view | 无 | — | 渲染页面 4（sys_oper_type 多选下拉 + sys_common_status 状态下拉服务端渲染字典） |
| 13 | POST | /monitor/operlog/list | monitor:operlog:list | 无 | 分页参数（白名单 **oper_name/oper_time/cost_time**）+ operIp(like)/title(like)/operName(like)/businessType/businessTypes(逗号串多选 in)/status/params[beginTime]/params[endTime] | TableDataInfo；行**驼峰** 17 列对位 SysOperLog（operId/title/businessType/method/requestMethod/operatorType/operName/deptName/operUrl/operIp/operLocation/operParam/jsonResult/status/errorMsg/operTime/costTime）；固定 order by oper_time desc（sortName 默认，白名单可改） |
| 14 | POST | /monitor/operlog/export | monitor:operlog:export | 操作日志, 5导出 | 同搜索参数；无分页导全量 | {code:0, msg:"&lt;uuid&gt;_操作日志.xlsx"}；Excel 列定义见下节 |
| 15 | POST | /monitor/operlog/remove | monitor:operlog:remove | 操作日志, 3删除 | ids（逗号串 operId） | 物理删除 in；toAjax |
| 16 | GET | /monitor/operlog/detail/{operId} | monitor:operlog:detail | 无 | 路径 operId | 渲染页面 5，变量 operLog 全行 |
| 17 | POST | /monitor/operlog/clean | monitor:operlog:remove | 操作日志, 8清空 | — | `truncate table sys_oper_log`；success() |

### 登录日志 /monitor/logininfor（对位 SysLogininforController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 18 | GET | /monitor/logininfor | monitor:logininfor:view | 无 | — | 渲染页面 6（sys_common_status 下拉） |
| 19 | POST | /monitor/logininfor/list | monitor:logininfor:list | 无 | 分页（白名单 **login_name/login_time**）+ ipaddr(like)/loginName(like)/status/params[beginTime]/params[endTime] | TableDataInfo；行驼峰 9 列（infoId/loginName/ipaddr/loginLocation/browser/os/status/msg/loginTime）；固定 order by login_time desc |
| 20 | POST | /monitor/logininfor/export | monitor:logininfor:export | 登录日志, 5导出 | 同搜索参数；导全量 | {code:0, msg:"&lt;uuid&gt;_登录日志.xlsx"}；列定义见下节 |
| 21 | POST | /monitor/logininfor/remove | monitor:logininfor:remove | 登录日志, 3删除 | ids | 物理删除 in；toAjax |
| 22 | POST | /monitor/logininfor/clean | monitor:logininfor:remove | 登录日志, 8清空 | — | truncate sys_logininfor；success() |
| 23 | POST | /monitor/logininfor/unlock | monitor:logininfor:unlock | 账户解锁, 0其它 | loginName（**query string 传**，页面 JS 实锤 `unlock?loginName=` + 选中列；多选逗号串——TP 版逐个 explode 处理） | 对位 passwordService.clearLoginRecordCache：`RedisCache::delete(TpConstant::PREFIX_PWD_RETRY . $loginName)` 逐个；恒 success()「操作成功」（键不存在也成功，经典版 remove 语义原样） |

### 数据库监控 /monitor/data（自研，deviations #6 兑现）

| # | 方法 | 路径 | #[Perm] | #[Log] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 24 | GET | /monitor/data | monitor:data:view | 无 | — | 渲染页面 7：SHOW STATUS 关键项（Threads_connected/Threads_running/Questions/Uptime/Innodb_buffer_pool 相关）+ SHOW PROCESSLIST（Id/User/Host/db/Command/Time/State/Info）+ version()/字符集；只读 |

统计：**控制器方法 24 个 / 24 条路由**；**#[Perm] 24 处**（逐路由 1 处；batchForceLogout 为一处注解双权限串 IS_REPEATABLE 双值）；**#[Log] 9 处**（强退 7×1、导出 5×2、删除 3×2、清空 8×2、解锁 0×1——Log attribute 已含 CLEAN=8 与 FORCE=7 常量实锤）。

## 特殊行为清单（在线用户 / 解锁 / 缓存枚举 / 服务降级 / quirk）

1. **在线用户解析与「外来会话」防护（本模块第一安全约束）**：TP 版会话 = `session:<uuid>` JSON `{user_id, loginName, userName, deptId, avatar, isAdmin, permissions, roles, pwdUpdateDate, rememberMe, …}`。列表只展示**含非空 loginName 键**的会话（匿名会话跳过）；缺 loginName 或 JSON 解析失败的键视为**外来会话**（本机 Redis db1 理论上只有 TP 版写入，但防护必须显式——tech-stack 定位「自用 db」不排除误配）：登录名称等字段显示「未知」、不参与强退。**强退外来会话严禁**：解析失败/无 loginName 的 sessionId 一律 error(500)「该会话无法识别，禁止强退」并**中止整批**（先到先拦，已完成部分不回滚——经典版循环 return 语义同为部分执行后中止，原样对位）。deptName：TP 会话只有 deptId → 显示查 sys_dept（或按会话 userName 兜底）；browser/os：经典版存 OnlineSession 字段，TP 版**落会话时补写**（见 2）。status 恒 'on_line'（键存在即在线；off_line 徽章分支保留给前端零改动）。
2. **会话元数据补写**：LoginService::login 成功组会话时新增 `meta` 子结构（或平铺键）：`ip` / `browser` / `os`（复用 LoginService::parseBrowser/parseOs）/ `loginTime`（Y-m-d H:i:s）/ `lastAccessTime`（LoginAuth touch 时顺带刷新，写回成本可接受——**替代方案**：lastAccessTime 不存，用 Redis TTL 换算剩余秒反推，动工时按实现难度择一，行为差异登记）。startTimestamp=loginTime；lastAccessTime 排序列消费它。**改 LoginService/LoginAuth 属 2.0.0 文件的定点扩展，不改既有键结构**（旧会话无 meta → 显示「未知」并按外来逻辑禁强退，向后兼容）。
3. **解锁 = 删 pwd_retry 键**：页面选中行取 loginName（selectColumns 去重）→ query string 传后端 → explode 逐个 `RedisCache::delete(TpConstant::PREFIX_PWD_RETRY . $name)`；不校验用户存在（经典版 clearLoginRecordCache 原样）；恒 success。
4. **缓存监控键前缀枚举（以 TpConstant.php 实际常量为准，动工时再 grep 复核）**：`session:`（会话）、`pwd_retry:`（密码重试）、`repeat_submit:`（防重复提交）、`rate_limit:`（限流——当前无使用方，枚举保留）、`config:`（sys_config 缓存）、`dict:`（sys_dict 缓存）。**验证码不设独立键**（存会话 captcha 字段，TpConstant 注释实锤）——枚举里没有 captcha 项，页面展示以此为准。cacheName 在页面即前缀名本身；getKeys=SCAN `prefix*`；getValue=RAW get + JSON 解码美化；clearAll 排除 session: 前缀（防全站踢线，见端点 11）。对位经典版七缓存名（sys-cache/sys-config/sys-dict/sys-authCache/sys-userCache/loginRecordCache/shiro-activeSessionCache）**不做名称映射**——TP 展示自家前缀，属数据源换代的自然差异（deviations 登记，见文末）。
5. **日志两表的删除/清空**：remove 物理删除（对位 delete … in）；clean=truncate（对位 mapper cleanOperLog/cleanLogininfor 实锤）；工具栏「清空」confirm 文案「确定清空所有操作日志/登录日志吗？」由 $.operate.clean 用 modalName 拼（ry-ui.js 实锤）。
6. **服务监控降级方案定稿建议（deviations #11 落地）**：分三档采集，**可得即显、不可得显示「—」**，布局照经典版四区块不重画：① **可得（PHP 原生）**——服务器信息（php_uname 名称/OS/架构）、PHP 环境（PHP 版本对位 JVM 区块改名「运行环境」、memory_get_usage/ini memory_limit 对位 JVM 内存、php SAPI、启动时间=进程起时估算——built-in server 单进程 worker 起时取不上则以「—」处理）、磁盘（**Windows 用 `disk_total_space('C:')` 等逐盘符**，COM wmi 不可依赖；Linux 挂载点遍历）；② **条件可得（尽力）**——CPU 核心数（Windows：环境变量 NUMBER_OF_PROCESSORS；Linux：/proc/cpuinfo）、CPU 使用率（Windows 不可靠 → 显示「—」；Linux 可 /proc/stat 双采样）、负载（仅 Linux load file）；③ **不可得显示「—」**——用户/系统使用率分项（Windows）、JVM 全项（无对应物，区块改「PHP 运行环境」）。页面四区块标题：CPU / 内存 / 服务器信息 / PHP 运行环境 / 磁盘状态。**不接受全页「平台不支持」占位**（页面对比度太差），也不引第三方系统监控扩展（composer 包 Windows 兼容性差、违背零依赖原则）。
7. **数据库监控（deviations #6 兑现）**：菜单 /monitor/data 在经典版是 DruidController 302 到 druid 外部页；TP 版渲染自研 data.html（SHOW STATUS 关键项 + PROCESSLIST + 版本信息）。wencun 账号权限若不足见 AGENTS.local.md root 方案。**PROCESSLIST 可能含其他会话敏感 SQL 文本——admin-only 权限串本就限制暴露面，不做额外脱敏**（自用环境）。
8. **quirk（原样保留，勿"修复"）**：① online 列表 exportUrl 指向 prefix+/export 但后端无 export 端点（经典版 SysUserOnlineController 无 export 实锤——showExport 的前端按钮点击后 404，经典版原样，**不复刻该 404，按 deviations 处理**：模板去掉 showExport 或后端补 export，动工时拍板，倾向**去 showExport**——登记 deviations）；② cache.html 经典版 fragment 拼写 `fragment-cache-kyes`（少 r）——TP 模板无 fragment 机制，仅备注；③ unlock 多选时 query string 逗号串含重复也无害（delete 幂等）；④ operlog 列表 escape:true（XSS 转义由 bootstrap-table 承担，jsonParam 展示在 detail 页 JSONView）；⑤ detail 页 operParam>2000 时降级 text（JSONView 不解析超长文本，经典版原样）。
9. **防重复提交**：经典版五个控制器零处 @RepeatSubmit（grep 实锤）→ 本模块不挂 RepeatSubmit。
10. **status 字典联动**：operlog/logininfor 的状态下拉与徽章全部走 DictService::listByType（sys_common_status/sys_oper_type），键 dict:* 已在 3.0.0 落地。

## 数据权限声明

- **本模块全部查询无 @DataScope**（五个 ServiceImpl/Mapper grep 实锤）——监控日志是全局视图，权限只到「能不能进页面/按钮」层级（#[Perm] 22 处）。
- unlock/强退可作用于任意用户（含 admin 会话强退——经典版仅拦「强退自己」，拦 admin 会话无特殊保护，原样）。

## Excel 导出列定义（@Excel 逐字段抄录）

**SysOperLog 17 列全带 @Excel（grep 实锤，逐列）**：

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | operId | name="操作序号", cellType=NUMERIC | 数字格式 |
| 2 | title | name="操作模块" | 文本 |
| 3 | businessType | name="业务类型", readConverterExp="0=其它,1=新增,2=修改,3=删除,4=授权,5=导出,6=导入,7=强退,8=生成代码,9=清空数据" | **注意 8=生成代码 9=清空数据**（与 Log attribute 的 8=CLEAN 编号体系不同——Excel 转换按此表达式，前端徽章字典按 sys_oper_type 8=生成代码/9=清空数据，两处一致） |
| 4 | method | name="请求方法" | 文本（类::方法全名） |
| 5 | requestMethod | name="请求方式" | 文本 |
| 6 | operatorType | name="操作类别", readConverterExp="0=其它,1=后台用户,2=手机端用户" | 转换输出 |
| 7 | operName | name="操作人员" | 文本 |
| 8 | deptName | name="部门名称" | 文本 |
| 9 | operUrl | name="请求地址" | 文本 |
| 10 | operIp | name="操作地址" | 文本 |
| 11 | operLocation | name="操作地点" | 文本 |
| 12 | operParam | name="请求参数" | 文本（截断后 2000 内原样） |
| 13 | jsonResult | name="返回参数" | 文本 |
| 14 | status | name="状态", readConverterExp="0=正常,1=异常" | 转换输出 |
| 15 | errorMsg | name="错误消息" | 文本 |
| 16 | operTime | name="操作时间", width=30, dateFormat="yyyy-MM-dd HH:mm:ss" | 日期格式 |
| 17 | costTime | name="消耗时间", suffix="毫秒" | 数字+「毫秒」后缀 |

**SysLogininfor 9 列**：infoId「序号」NUMERIC / loginName「用户账号」/ status「登录状态」0=成功,1=失败 / ipaddr「登录地址」/ loginLocation「登录地点」/ browser「浏览器」/ os「操作系统」/ msg「提示消息」/ loginTime「访问时间」width=30 日期格式。

- sheet 名「操作日志」「登录日志」；文件名 `&lt;uuid&gt;_操作日志.xlsx` / `&lt;uuid&gt;_登录日志.xlsx`；复用 3.0.0 ExcelExportService（列头样式从简同 deviations #19）。

## 关键设计说明

0. **CheckPerm 扩展点（前置依赖）**：经典版 batchForceLogout 为 `@RequiresPermissions(value={batchForceLogout, forceLogout}, logical=OR)`；现 `Perm` attribute 已声明 IS_REPEATABLE 但 `CheckPerm::resolvePerm` 只读 `attrs[0]`——本模块需小幅扩展：resolvePerm 返回**全部**注解值数组，CheckPerm 改为「任一命中即放行」（单值场景行为不变，1.0.0 已验链路回归）。此为 1.0.0 设施的**兼容扩展**，不改既有调用方。

1. **OnlineService 扫描管道**：`keysScan('session:*')` → 每键 `get(raw:true)` + json_decode → assoc 数组 → `uuid = substr(key, 8)`；LIKE 过滤在 PHP 侧做（数据在 Redis 非 DB，无法 SQL where）；排序在 PHP 侧 usort（loginName 字典序 / startTimestamp、lastAccessTime 字符串比较即可）；分页 array_slice + total=count。会话量级（自用环境 <100）性能无虞；keysScan 本就是 SCAN 非 KEYS。
2. **强退自保护判据**：`sessionId === $request->middleware('session_uuid')`（对位经典版 `sessionId.equals(ShiroUtils.getSessionId())`）。
3. **CacheController HTML 片段**：三处 get* 返回 `Response::create('monitor/cache/fragment_xxx', 'view')` 部分模板（无布局），$.ajax success 直接 `.html(data)` 塞容器——对位经典版 th:fragment 机制；片段模板放在 cache 页同目录（layer 相对路径机制核对）。
4. **MonitorCacheService 前缀注册表**：`['session' => TpConstant::PREFIX_SESSION, 'pwd_retry' => …, …]`（键=展示名/值=前缀常量）；新增缓存前缀时只改这一处 + TpConstant。clearCacheName/clearAll 逐键 delete 用 raw 模式（键已是全名）。
5. **OperLogService/LogininforService 查询构造**：like 三处（operIp/title/operName；ipaddr/loginName）+ businessType= / businessTypes in（逗号串 explode 后 whereIn）+ status= / beginTime/endTime（>= beginTime、<= endTime——前端 beginOfTime/endOfTime 已拼 00:00:00/23:59:59）。
6. **ServerInfoService 零依赖**：只内建函数（php_uname/disk_total_space/memory_get_usage/getrusage(条件)/get_cfg_var/NUMBER_OF_PROCESSORS），不引 exec 类调用（Windows exec 权限与杀毒干扰，违背零折腾原则）；每项采集 try-catch 兜底 null。
7. **模板 layer 归属**：monitor 控制器在 `app\controller\monitor` layer → 视图根 app/view/monitor/；include 片段（bootstrap-select/js、jsonview、bootstrap-table-export）需复制到 app/view/monitor/include/（3.0.0 踩坑先例）。
8. **operlog 详情页业务类型**：服务端 `dict.getLabel('sys_oper_type', businessType)`——DictService 扩一个 getLabel(type, value) 静态方法（listByType 内过滤），或模板内 volist 查找；取前者（一处小扩展，7.0.0 公告也可能用）。

## 拟登记 deviations

| # | 差异点 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|---|
| 1 | 缓存监控对象 | ehcache 七缓存名（sys-cache/sys-config/sys-dict/sys-authCache/sys-userCache/loginRecordCache/shiro-activeSessionCache） | Redis 六前缀（session:/pwd_retry:/repeat_submit:/rate_limit:/config:/dict:；验证码在会话内无独立键）；clearAll 排除 session: | 数据源换代的自然差异（ehcache→RedisCache），端点契约（7 条路由 + 片段响应）保持；clearAll 排除会话为防全站踢线的加固 |
| 2 | 在线用户存储 | OnlineSession 同步 sys_user_online 表（dbSyncPeriod 1 分钟）+ 表查询 | Redis `session:*` 实时 SCAN，sys_user_online 表不读写（表随官方 SQL 存在但闲置） | Redis 会话即真源（deviations #1 的延伸），无同步延迟；表闲置留痕 |
| 3 | 在线用户 export | 前端 showExport=true 但后端无 export 端点（点击 404，经典版原样缺陷） | 倾向模板去 showExport（或补 export 端点，动工时拍板） | 不复刻 404 缺陷；拍板后回填本行 |
| 4 | 服务监控指标 | oshi 全量（CPU 使用率分项/JVM/磁盘多分区） | 三档降级（可得/条件可得/「—」）；JVM 区块改「PHP 运行环境」 | deviations #11 定稿方案（见特殊行为 6）；页面布局与区块结构保持 |
| 5 | 数据库监控 | DruidController 302 跳 druid 外部页 | 自研 SHOW STATUS/PROCESSLIST 简页 | deviations #6 已登记，本模块兑现；菜单入口行为从「跳外部」变「内置页」 |
| 6 | 会话元数据来源 | OnlineSession 字段（host/browser/os/lastAccessTime 由 Shiro 维护） | 登录时写入会话 meta（ip/browser/os/loginTime）+ touch 刷新 lastAccessTime（或 TTL 换算，动工时择一）；旧会话/外来会话显示「未知」 | 会话结构自管的连带实现；行为等价 |

另：解锁=删 pwd_retry 键、强退自保护、日志 truncate、批量删除物理删等均为经典版原样照抄，无差异需登记。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-10-01）

- 五控制器 24 路由（online 3/server 1/cache 7/operlog 6/logininfor 6/data 1）+ 六服务（OnlineService SCAN 管道/MonitorCacheService 前缀枚举/OperLog/Logininfor/ServerInfo 三档采集/Data 自研页）。
- CheckPerm 兼容扩展：resolvePerm 返回注解值数组、任一命中放行（batchForceLogout 双权限串 OR）。
- 会话元数据补写：LoginService 平铺 ip/browser/os/loginTime/lastAccessTime；LoginAuth touch 刷新（择写回方案）；旧会话「未知」兜底。
- 10 模板 + 3 片段 + 7 include 副本（主线程串行移植——用户指令改串行后由主线完成）；联调期修复 2 处：detail.html json_encode 加 |raw（HTML 转义破坏 JS 字面量致 JSONView 不渲染）、DataController 去掉 get_object_vars（think-orm 返回数组）。
- 浏览器全链路：在线用户（自踢防护/强退 ry）、缓存三栏联动（config→8 键→skin-blue）、服务五区块（16 核/5 盘）、操作日志（JSONView/删除/徽章）、登录日志解锁（制造锁定→unlock→键消失）、数据监控三区块。
- PHPUnit 64 tests 155 assertions 全绿（MonitorOnlineTest 6 用例新增）；双导出 xlsx 读回 17/9 列核对。
- 测试数据全清理；clearCacheName 对 session 前缀拒绝（clearAll 排除之外的加固）。
