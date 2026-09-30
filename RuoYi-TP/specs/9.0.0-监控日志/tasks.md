# 9.0.0-监控日志 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。2026-10-01 动工并完成（模板移植改为主线程串行完成）。

## Task 1 · 会话元数据补写（2.0.0 文件定点扩展）

- [x] LoginService::login 会话数组补平铺键：ip/browser/os/loginTime/lastAccessTime（parseBrowser/parseOs 复用）
- [x] LoginAuth touch 时刷新 lastAccessTime（择「写回」方案：仅含 loginName 的会话且日期变化才写——写放大最小；TTL 换算备选弃用）
- [x] 旧格式会话（无 meta）解析不报错（OnlineService 全部 `?? '未知'` 兜底）
- [x] 登录链路回归：admin 登录后会话含全部 meta 键（Redis 核对）；PHPUnit 64 tests 全绿

## Task 2 · OnlineService + OnlineController

- [x] CheckPerm 兼容扩展：resolvePerm 返回全部注解值数组、任一命中放行（单值行为不变；无其他调用方；全量回归绿）
- [x] OnlineService：keysScan → 解析 → 匿名/外来过滤 → like 过滤 → 排序（白名单驼峰转换）→ 手工分页
- [x] OnlineController 3 路由：双权限串 IS_REPEATABLE；自会话拒；外来会话拒且中止整批；#[Log] 7强退
- [x] PHPUnit：MonitorOnlineTest 6 用例（过滤/排序稳定性/白名单拒绝/分页）
- [x] curl 级：list 驼峰行/过滤；强退 ry 键消失；强退自己 500；外来键强退被拒

## Task 3 · MonitorCacheService + CacheController

- [x] 前缀注册表（TpConstant 六常量；验证码无键不入枚举）
- [x] 7 路由：三片段 HTML（fragment_names/keys/value）；clearCacheName/clearCacheKey/clearAll JSON；**clearCacheName 对 session 前缀拒绝 + clearAll 排除 session**（spec 只定 clearAll 排除——clearCacheName 单前缀清 session 等同全站踢线，一并防护，加固点）
- [x] getKeys 空前缀合并；getValue 键不存在「（不存在或已过期）」
- [x] curl + 浏览器：片段消费/清理复零/session 存活全实测

## Task 4 · OperLogService + LogininforService + 两 Controller

- [x] OperLogService：like 三处 + businessType(s) in + status + 时间范围 + 白名单 + deleteByIds/truncate/selectById + 驼峰 17 列
- [x] OperlogController 6 路由（#[Perm] 6 + #[Log] 3）
- [x] LogininforService：like 两处 + status + 时间范围 + 白名单 + deleteByIds/truncate/unlock
- [x] LogininforController 6 路由；unlock query string → explode 逐个删 pwd_retry；#[Log] 账户解锁,0其它；恒 success
- [x] DictService::getLabel 已存在（2.0.0 预留），零扩展
- [x] PHPUnit：businessTypes 解析在 curl 实测（42 行过滤正确）；MonitorOnlineTest 覆盖排序分页
- [x] curl 级：list/过滤/remove ids 实测；unlock 键消失实测

## Task 5 · Excel 导出（两表 26 列）+ ServerInfoService

- [x] 17 列/9 列列定义（convert 含 8=生成代码 9=清空数据口径；costTime 列头带「(毫秒)」替代 suffix——ExcelExportService 无 suffix 机制，列头语义等价）
- [x] export 两端点：xlsx 真实生成 + 下载读回断言（17/9 列列头核对）
- [x] ServerInfoService 三档采集：php_uname/disk_total_space 逐盘符/memory/NUMBER_OF_PROCESSORS；逐项 try-catch 兜底 null；零 exec
- [x] ServerController 1 路由

## Task 6 · 数据库监控页

- [x] DataController（monitor:data:view）+ data.html：SHOW GLOBAL STATUS 7 关键项 + PROCESSLIST + version/字符集 + QPS
- [x] wencun 权限足够（SHOW PROCESSLIST 正常返回）；修复 get_object_vars 误用（think-orm 返回数组）

## Task 7 · 页面模板（7 页 + 3 片段 + include 副本）——主线程串行完成

- [x] online.html（去 showExport——deviations #3 拍板；强退链路 check_perm 输出形式）
- [x] cache.html 三栏 + 三个 fragment 片段；server.html 五区块；data.html 三区块
- [x] operlog.html（bootstrap-select + queryParams 覆写 + 字典徽章）+ detail.html（JSONView + 复制 + 异常卡片）
- [x] logininfor.html（解锁按钮 query string 原样）
- [x] include 副本 7 个（header/footer/bootstrap-select×2/bootstrap-table-export/jsonview×2）
- [x] 静态自查：th: 残留 0、check_perm 形式合规、volist/if 配平
- [x] 联调期修复 2 处：detail.html json_encode 输出加 |raw（HTML 转义破坏 JS 字面量）；DataController 去掉 get_object_vars

## Task 8 · 端到端验收（浏览器级）

- [x] 在线用户：admin 行 + 强退 confirm + 自踢防护提示；强退 ry 键消失（curl 侧）
- [x] 缓存监控：三栏联动全链路（config→8 键→skin-blue）
- [x] 服务监控：五区块（16 核/5 磁盘分区）；数据库监控：三区块（7 指标/进程列表）
- [x] 操作日志：列表/删除/详情 JSONView（修复后 pretty 渲染）/「强退」徽章
- [x] 登录日志：解锁链路（制造锁定→unlock→键消失）
- [x] 菜单六入口全部可开
- [x] ry 表零 schema 变更；sys_user_online 零读写

## Task 9 · 收尾

- [x] PHPUnit 全绿（64 tests 155 assertions）；spec.md 实施记录回填；checklist 勾选；README 总表更新（9.0.0 ✅）
- [x] deviations 处置回填：六条候选全部定稿（showExport 去除/lastAccessTime 写回方案）
- [x] 测试数据清理：外来键/锁定键/测试日志行/导出残留全清；admin 会话正常
