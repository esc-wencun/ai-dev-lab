# Spec-09 定时任务

> **状态：✅ 已完成（2026-09-24）**。Task 1-3 全部落地并端到端验证：
> 注册表机制（坏cron/未注册目标拒绝）、任务CRUD+APScheduler动态调度（每5秒任务12秒产生3条日志、
> 暂停后无新日志、恢复+立即执行）、执行日志（成功/耗时记录、批量删除、clean）。
> 修复记录：scheduler_util缺datetime导入；/jobLog/clean被/{job_log_ids}路由遮蔽（clean转发处理）。
> 测试数据已清理。
>
> **补充对齐（2026-09-24 二次核查）**：扫描库内数据发现java预置3条任务（ryTask.ryNoParams /
> ryParams('ry') / ryMultipleParams('ry', true, 2000L, 316.50D, 100)）未覆盖，已补齐：
> - `module_task/ry_task.py`：RyTask三方法等价实现
> - `module_task/target_resolver.py`：java语法参数解析器（'str'/bool/2000L/1.5D/int推断）+
>   完整校验链（rmi/ldap/http黑名单、违规包名、白名单，文案对齐SysJobController）
> - 修复关键bug：暂停态任务启动时不加载进调度器，changeStatus启用时resume_job静默空操作
>   （表现为"暂停的任务永远无法启用"）——resume_job现在对不在调度器的job从库补载
> - 3条预置任务run执行成功（status=0），启用后按cron周期调度验证通过
>
> Java 版对应：`ruoyi-quartz` 模块（SysJobController、SysJobLogController、ScheduleUtils、JobInvokeUtil）
> 前端页面：`views/monitor/job/`、`views/monitor/jobLog/`
> 依赖：spec-01
> 技术选型：APScheduler（对齐参考项目），cron 表达式兼容 Quartz 六位格式

## 目标

定时任务的管理（库存储）+ APScheduler 动态调度 + 执行日志。Python 侧不支持 Java 反射调用目标字符串，需设计等价的任务目标机制。

## API 清单

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 1 | GET | `/monitor/job/list` | `monitor:job:list` |
| 2 | GET | `/monitor/job/{jobId}` | `monitor:job:query` |
| 3 | POST | `/monitor/job` | `monitor:job:add` |
| 4 | PUT | `/monitor/job` | `monitor:job:edit` |
| 5 | PUT | `/monitor/job/changeStatus` | `monitor:job:changeStatus` |
| 6 | PUT | `/monitor/job/run` | `monitor:job:changeStatus` |
| 7 | DELETE | `/monitor/job/{jobIds}` | `monitor:job:remove` |
| 8 | GET | `/monitor/jobLog/list` | `monitor:joblog:list` |
| 9 | DELETE | `/monitor/jobLog/{jobLogIds}` | `monitor:joblog:remove` |
| 10 | DELETE | `/monitor/jobLog/clean` | `monitor:joblog:remove` |

## Task 1: 任务目标机制（替代 Java 反射 invokeTarget）

- [x] 设计 `module_task/registry.py`：任务注册表（字符串目标 → 可调用函数），如 `task.sample` → `module_task.sample_task.run`
- [x] `invoke_target` 合法性校验：只允许注册表中的目标（安全上比 Java 白名单常量 JOB_WHITELIST_STR 更严格）
- [x] 写 2 个示例任务：无参任务、带参任务（对应 Java ryTask.ryParams('ry') 演示）
- [x] 单元测试：目标解析、非法目标拒绝

## Task 2: 任务 CRUD + 调度联动

- [x] DO：`SysJob`（`sys_job` 表）+ `SysJobLog`（`sys_job_log` 表）
- [x] 启动时：加载 status='0' 的任务到 APScheduler（对应 Java ScheduleUtils.init；cron 用参考项目 MyCronTrigger 的六位兼容方案）
- [x] 新增：cron 表达式合法性校验（`新增任务%s失败，cron表达式不正确`）；目标校验；加入调度器
- [x] 修改：暂停→移除→重建（对应 Java 同名流程）；`changeStatus`：暂停/恢复；`run`：立即触发一次（APScheduler add_job 直接执行）
- [x] 删除：物理删除 + 从调度器移除；批量
- [x] 并发/防丢策略记录：misfire_grace_time 与 coalesce 配置在代码注释中说明（Java quartz 的 @DisallowConcurrentExecution 对应 APScheduler max_instances=1）
- [x] 端到端测试：建一个每 10 秒任务 → job_log 持续产生记录 → 暂停后停止 → 删除后消失

## Task 3: 执行日志

- [x] 执行器包装：任务执行前后写 `sys_job_log`（任务名、组名、目标字符串、日志信息、状态 0 成功 1 失败、异常信息、开始时间、耗时）；异常捕获后记录不外抛
- [x] `jobLog/list`：分页 + 过滤（`jobName / jobGroup / status / 时间区间`）
- [x] `jobLog/{ids}` 批量删除、`clean`
- [x] 端到端测试：失败任务（抛异常的示例）日志记录 error_msg

## 验收清单

- [x] RuoYi-Vue3 定时任务 + 任务日志页面全部可用
- [x] 服务重启后任务自动恢复调度
- [x] pytest 全绿；更新 specs/README.md 状态
