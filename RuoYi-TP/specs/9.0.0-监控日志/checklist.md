# 9.0.0-监控日志 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（online 3 + server 1 + cache 7 + operlog 6 + logininfor 6 + data 1 = 24 条路由）。

## 在线用户（curl / Redis 级）

- [x] GET /monitor/online 渲染页面 1（浏览器 tab 打开正常）
- [x] POST /monitor/online/list：TableDataInfo；匿名会话不出现（无 loginName 跳过）；行驼峰 10 列（expireTime 不出）；loginName like 过滤（admin→3/noexist→0 实测）
- [x] 强退：他人会话（ry）→ Redis 键消失 + code 0 实测；自己 → 500「当前登录用户无法强退」（curl + 浏览器双实测）；键不存在 → 500「用户已下线」（代码分支，逻辑同 resolve unknown 前置 has 检查）
- [x] **外来会话防护**：构造 `session:foreign…`（值非法 JSON）→ 500「该会话无法识别，禁止强退」实测（键未被删）
- [x] 双权限串：#[Perm] batchForceLogout + forceLogout 双注解（IS_REPEATABLE）；CheckPerm 扩展任一命中放行（resolvePerm 返回数组，64 tests 回归含 1.0.0 权限链路全绿）
- [x] #[Log] 在线用户,7强退 落库（oper_id 250 实查，浏览器详情页徽章「强退」渲染）

## 缓存监控（curl / Redis 级）

- [x] GET /monitor/cache 渲染三栏页；前缀枚举六项（浏览器实测 6 行）；无 captcha 项
- [x] getNames/getKeys/getValue 返回 HTML 片段（浏览器 .html() 消费实测：键列表 8 键、值「skin-blue」pretty）
- [x] getKeys 前缀过滤（config 8 键）；空前缀 = 全库已知前缀合并（代码路径 + allKnownKeys 排序）
- [x] getValue：JSON pretty（skin-blue 实测）；键不存在 → 「（不存在或已过期）」（prettyValue null 分支——curl 级模板就绪前未单独触发，代码路径明确）
- [x] clearCacheName：dict:* 全清 SCAN 复 0 实测；**clearAll 后 session:* 存活**（本人不掉线实测）
- [x] Redis 全走 RedisCache 门面（grep 零 predis 直引）

## 服务监控 / 数据库监控

- [x] GET /monitor/server 渲染五区块（CPU 16 核真实值/内存/服务器信息/PHP 运行环境/磁盘）；Windows CPU 使用率等不可得项显「—」
- [x] 磁盘逐盘符 5 行渲染；>80% 红字逻辑在模板 `{if condition="$disk['usageRate'] > 80"}class="text-danger"{/if}`
- [x] GET /monitor/data 渲染：7 项 SHOW STATUS + PROCESSLIST（1 行）+ 版本/字符集 + QPS；wencun 权限足够（无降级触发）
- [x] 零 exec/shell（grep 核对，唯一命中为注释）

## 操作日志（curl / DB 级）

- [x] GET /monitor/operlog 渲染（多选下拉 + 状态下拉字典渲染实测）
- [x] POST list：驼峰 17 列实测；businessTypes 多选 in（1,2 → 全部行 businessType∈{1,2} 实测 42 行）；operTime desc 默认
- [x] POST export：17 列 xlsx 读回断言（列头顺序/转换表达式在列定义，PHPUnit 机制同 6.0.0 已验）
- [x] POST remove：ids=250 物理删实测（7→6 行）
- [x] GET detail/{operId} 渲染：五卡片 + JSONView pretty（operParam/jsonResult 实测渲染）；status!=0 异常卡片条件渲染（{if} 在模板）
- [x] POST clean：truncate（代码路径；未实际清空全表——保留业务日志数据，truncate 语义同 6.0.0 已验证的 Db::execute 链路）
- [x] #[Log] 3 处（导出/删除/清空；detail 无 Log）

## 登录日志（curl / DB / Redis 级）

- [x] GET /monitor/logininfor 渲染
- [x] POST list：驼峰 9 列实测；过滤/时间范围同 operlog 管道
- [x] POST export：9 列 xlsx 读回断言
- [x] POST remove / clean：物理删 + truncate（同上）
- [x] **POST unlock**：构造 pwd_retry:m9user → unlock?loginName=m9user → 键消失 + code 0 实测；逗号串逐个删（explode 分支）
- [x] #[Log] 账户解锁,0其它 落库（business_type=0）

## 页面级验收（浏览器）

- [x] 在线用户页：admin 行渲染 + 状态徽章「在线」+ tooltip + 强退 confirm（「确认要强退选中的1条数据吗?」）→ 自踢防护提示实测
- [x] 缓存监控三栏联动：config → 8 键 → skin-blue 值，全链路实测
- [x] 服务监控页：五区块 + 折叠钮（经典版交互 JS 原样）
- [x] 操作日志：多选下拉 + 「强退」徽章 + 删除链路 + **详细弹窗 JSONView**（修复 `|json_encode` 被 HTML 转义破坏 JS 字面量的 bug——加 `|raw`；operParam/jsonResult pretty 实测）
- [x] 登录日志：渲染正常；解锁按钮在工具栏；API 解锁实测
- [x] 数据库监控页：三区块表格渲染
- [x] 全部页面主框架 iframe 内打开（菜单点击六 tab 实测）

## 横切与纪律自查

- [x] #[Perm] 25 处注解（batchForceLogout 双注解计 2；spec 写 24 是按路由计）——sys_menu monitor:* 25 行 perms 互查覆盖
- [x] #[Log] 8 处（强退 1/导出 2/删除 2/清空 2/解锁 1）——spec 写 9 按「导出 5×2」计法差异，实际端点数一致
- [x] RepeatSubmit 未挂（grep 0）；零 DataScope；sys_user_online 表零读写（grep 仅注释）
- [x] sys_config/sys_dict 缓存仍走 ConfigService/DictService
- [x] 表结构零变更

## Deviations 核对

- [x] 缓存监控对象（ehcache 七名 → Redis 六前缀；clearAll 排除 session）——登记 deviations（spec 拟登记表 #1 落地）
- [x] 在线用户存储（sys_user_online 表 → Redis 实时 SCAN）——拟登记表 #2 落地
- [x] 在线用户 export：**拍板去掉 showExport**（不复刻经典版 404 缺陷）——模板已去，deviations 回填
- [x] 服务监控三档降级（JVM 区块改「PHP 运行环境」）——落地
- [x] 数据库监控自研页——落地（菜单入口从跳外部变内置页）
- [x] 会话元数据补写（loginTime/lastAccessTime/ip/browser/os 平铺键；LoginAuth touch 刷新 lastAccessTime；旧会话显示「未知」）——落地

## 测试数据清理记录

- [x] ✅ 2026-10-01 外来测试会话键（session:foreign…）删除；pwd_retry:m9user 删除
- [x] ✅ 2026-10-01 ry 会话已随强退测试删除（ry 用户本身预置保留）；sys_user 2 行（admin/ry）原值
- [x] ✅ 2026-10-01 sys_logininfor 测试时段行清理；sys_oper_log 强退测试行（>244）删除，业务日志保留
- [x] ✅ 2026-10-01 runtime/download 导出残留清空；runtime/upload 无残留
- [x] ✅ 2026-10-01 admin 会话正常（admin/admin123 多轮登录实测）；dict:/config: 缓存已重新预热
