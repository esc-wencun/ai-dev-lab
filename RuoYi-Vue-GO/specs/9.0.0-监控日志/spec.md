# 09 监控：在线用户 / 缓存监控 / 操作日志 / 登录日志

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：SysUserOnlineController + CacheController + SysOperlogController + SysLogininforController
> 前端页面：views/monitor/online、views/monitor/cache、views/monitor/operlog、views/monitor/logininfor、views/monitor/server
> 依赖：1.0.0-基础设施（操作日志写入在 01 的 aspect 包；会话在 security.TokenService）

## API 清单（已对照 Java 源码核实，2026-09-24）

### 在线用户 /monitor/online
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | monitor:online:list | SCAN login_tokens 聚合（userName/ipaddr 过滤） |
| 2 | DELETE | /{tokenId} | monitor:online:forceLogout | 强退（删会话） |

### 缓存监控 /monitor/cache
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | / | monitor:cache:list | Redis INFO + dbSize |
| 2 | GET | /getNames | monitor:cache:list | 七前缀分组静态清单 |
| 3 | GET | /getKeys/{cacheName} | monitor:cache:list | SCAN 前缀键 |
| 4 | GET | /getValue/{cacheName}/{cacheKey} | monitor:cache:list | 取值（JSON 解析） |
| 5 | DELETE | /clearCacheKey/{cacheKey} | monitor:cache:list | 删单键 |
| 6 | DELETE | /clearCacheAll | monitor:cache:list | 清七组全部 |

### 操作日志 /monitor/operlog
list（monitor:operlog:list）/{operIds} 删除/clean 清空（monitor:operlog:remove）；数据源 sys_oper_log（01 操作日志切面写入）。

### 登录日志 /monitor/logininfor
list（monitor:logininfor:list）/{infoIds} 删除/clean 清空（monitor:logininfor:remove）/unlock/{userName} 解锁（monitor:logininfor:unlock，清 pwd_err_cnt）。

### 服务监控 /monitor/server
Java SysServerController 返回 CPU/内存/JVM/系统信息。**未实现**——Java 版依赖 OSHI+JVM 数据，Go 版对应 gopsutil 方案待设计，登记 deviations（前端页面为只读信息展示，无 CRUD 影响）。

## 实施记录

- 2026-09-24 完成。在线用户=SCAN 会话聚合（tokenId 即会话 uuid，强退=删会话键，行为与 Java 一致）。
- 端到端：公告外全部监控端点（operlog/logininfor/online/cache getNames）列表返回正确；强退/解锁/清理逻辑与 2.0.0 登录链共用设施。
- RedisCache 新增 Info/DBSize；TokenService 新增 GetLoginUserByUUID。
