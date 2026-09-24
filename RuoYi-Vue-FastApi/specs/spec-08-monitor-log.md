# Spec-08 监控模块：在线用户 / 服务监控 / 缓存监控 / 操作与登录日志

>
> **状态：✅ 已完成（2026-09-24）**。在线用户(SCAN共享redis会话)、服务监控(psutil对齐前端字段)、缓存监控(info/getNames/getKeys/getValue/三种clear)、操作日志查询删除清空、登录日志查询删除清空+解锁。端到端验证通过。重要修正：get_cache_object不再删除无法解析的java缓存（共享redis保护），改为原样返回文本。
> Java 版对应：`SysUserOnlineController`、`ServerController`、`CacheController`、`SysOperlogController`、`SysLogininforController`
> 前端页面：`views/monitor/online|server|cache|jobs|logininfor|operlog/`
> 依赖：spec-01（操作日志在 spec-01 已建，本 spec 补查询接口）

## 目标

四个监控页面 + 两类日志查询的后端支持。登录日志的写入在 Phase 0 已实现，此处只做查询管理。

## API 清单

### 在线用户

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 1 | GET | `/monitor/online/list` | `monitor:online:list` |
| 2 | DELETE | `/monitor/online/{tokenId}` | `monitor:online:forceLogout` |

### 服务监控

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 3 | GET | `/monitor/server` | `monitor:server:list` |

### 缓存监控

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 4 | GET | `/monitor/cache` | `monitor:cache:list` |
| 5 | GET | `/monitor/cache/getNames` | `monitor:cache:list` |
| 6 | GET | `/monitor/cache/getKeys/{cacheName}` | `monitor:cache:list` |
| 7 | GET | `/monitor/cache/getValue/{cacheName}/{cacheKey}` | `monitor:cache:list` |
| 8 | DELETE | `/monitor/cache/clearCacheName/{cacheName}` | `monitor:cache:list` |
| 9 | DELETE | `/monitor/cache/clearCacheKey/{cacheKey}` | `monitor:cache:list` |
| 10 | DELETE | `/monitor/cache/clearCacheAll` | `monitor:cache:list` |

### 操作日志

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 11 | GET | `/monitor/operlog/list` | `monitor:operlog:list` |
| 12 | DELETE | `/monitor/operlog/{operIds}` | `monitor:operlog:remove` |
| 13 | DELETE | `/monitor/operlog/clean` | `monitor:operlog:remove` |

### 登录日志

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 14 | GET | `/monitor/logininfor/list` | `monitor:logininfor:list` |
| 15 | DELETE | `/monitor/logininfor/{infoIds}` | `monitor:logininfor:remove` |
| 16 | DELETE | `/monitor/logininfor/clean` | `monitor:logininfor:remove` |
| 17 | GET | `/monitor/logininfor/unlock/{userName}` | `monitor:logininfor:unlock` |

## Task 1: 在线用户

- [x] `list`：扫描 `login_tokens:*`（SCAN，非 KEYS，避免阻塞共享 Redis），解析会话返回 `{tokenId(=uuid), userId, userName, deptName, ipaddr, loginLocation, browser, os, loginTime}`；过滤 `ipaddr / userName`
- [x] `forceLogout`：删除对应会话键（对应 Java delLoginUser），前端提示强退成功
- [x] 端到端测试：登录两个会话 → 列表可见 → 强退后该会话 401（另一会话不受影响）

## Task 2: 服务监控

- [x] `GET /monitor/server`：用 psutil 组装与 Java ServerController 相同结构的返回（CPU / 内存 / 服务器信息 / Python 环境信息（对应 jvm 信息：名称、版本、启动时间、运行时长、安装路径）/ 磁盘列表）
- [x] 字段命名对照前端 `views/monitor/server/index.vue` 全部对齐
- [x] 端到端测试：页面渲染正常无 undefined

## Task 3: 缓存监控

- [x] `getInfo`：Redis info 组装 `{info: {db_size, uptime, connected_clients, ...}, dbSize, commandStats}`（命令统计用 info commandstats，前端饼图所需 key/format 对照 Java）
- [x] `getNames`：固定缓存组列表（对照 Java CacheController 的常量：登录用户/用户信息/部门/岗位/角色/菜单/字典/参数/通知等 cacheName+remark）
- [x] `getKeys/{cacheName}`：该前缀下所有键（`{cacheName}:*`）
- [x] `getValue`：返回 `{cacheKey, cacheValue, remark}`（cacheValue 为缓存内容字符串）
- [x] 三个 clear 接口：按前缀删除 / 全部删除指定组（clearCacheAll 语义对照 Java：清配置的缓存组，非 flushall）
- [x] 端到端测试：登录后 getKeys(login_tokens) 可见会话键，getValue 内容正确

## Task 4: 操作日志查询

- [x] 分页列表 + 过滤（`title / operName / businessType / status / 时间区间`，排序默认 oper_time desc 对照 Java）
- [x] 批量删除、`clean`（truncate sys_oper_log——清空前确认 Java 行为是 delete all 还是 truncate）
- [x] 端到端测试：spec-01 写入的日志可查可删

## Task 5: 登录日志查询

- [x] 分页列表 + 过滤（`ipaddr / userName / status / 时间区间`）+ 批量删除 + clean
- [x] `unlock/{userName}`：删除 `pwd_err_cnt:{userName}`（解锁账户锁定）；成功提示 `%s账户解锁成功`
- [x] 端到端测试：故意错 5 次密码后正确密码也被拒（锁定生效）→ unlock 成功 → 登录恢复

## 验收清单

- [x] RuoYi-Vue3 六个监控页面全部可用
- [x] pytest 全绿；更新 specs/README.md 状态
