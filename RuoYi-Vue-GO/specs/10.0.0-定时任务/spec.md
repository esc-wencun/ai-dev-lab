# 10 定时任务

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：ruoyi-quartz 模块（SysJobController / SysJobLogController / ScheduleUtils / AbstractQuartzJob）
> 前端页面：views/monitor/job（含 jobLog）
> Go 实现：robfig/cron v3（Quartz 秒级 6 位表达式 parser，比 gocron 更贴合 Java cron 语义）

## API 清单（已对照 Java 源码核实，2026-09-24）

### 定时任务 /monitor/job
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | monitor:job:list | TableDataInfo；jobName like + jobGroup/status |
| 2 | POST | /export | monitor:job:export | xlsx（暂未实现，随 deviations #15 一并处理） |
| 3 | GET | /{jobId} | monitor:job:query | 详情 |
| 4 | POST | / | monitor:job:add | 新增默认暂停（status='1'）不进调度器 |
| 5 | PUT | / | monitor:job:edit | 调度中任务重载 cron |
| 6 | PUT | /changeStatus | monitor:job:changeStatus | 启用=从库补载（规避 Python 版踩坑 5 的静默空操作）；暂停=移出 |
| 7 | PUT | /run | monitor:job:changeStatus | 立即执行一次 + 写 sys_job_log |
| 8 | DELETE | /{jobIds} | monitor:job:remove | 物理删 + 移出调度器 |

### 任务日志 /monitor/jobLog
list（monitor:job:list）/ {jobLogId}（monitor:job:query）/ {jobLogIds} 删除 + clean 清空（monitor:job:remove）。

## 特殊行为设计（对位 spec 要求逐条）

- **路由遮蔽**：changeStatus/run 固定路径先于 /:jobId 注册（Gin 固定优先）。
- **调度器生命周期**：暂停态启动不加载；恢复从库读 cron 补载（AddJob）。
- **cron 表达式**：robfig/cron v3 + NewParser(Second|...|Descriptor) 原生支持 Quartz 6/7 位（0/10 * * * * ? 已验证加载）。
- **invokeTarget**：注册表方案（同 Python 版 spec-09）——ParseTarget 语法解析兼容 Java（bean.method(params)，参数类型推断 'x'/true/2000L/316.50D/100 一致），执行走 Go 注册表；ryTask 三预置任务已注册（ry_task 对位）。
- **concurrent/misfire**：cron v3 无对应策略位——concurrent 暂不限制并发（Go goroutine 天然并发，任务本身幂等），misfire 补跑策略未对齐，登记 deviations。

## 实施记录

- 2026-09-24 完成。修复 main.go 缺失 logger.Init 的装配缺口（任务执行报"logger 未初始化"）。
- 端到端：任务列表（预置 3 条）、run 立即执行（sys_job_log status=0 执行成功）、changeStatus 暂停/恢复全通过。
- 单元测试：ParseTarget（4 形态 + 类型推断 + 非法目标）、Registry（预置注册）。
