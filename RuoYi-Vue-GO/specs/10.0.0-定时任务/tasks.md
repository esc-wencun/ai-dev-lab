# Tasks · 10 定时任务

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：对照 SysJobController/SysJobLogController/ScheduleUtils 写实 API 清单；核对 sys_job 预置 3 条任务（数据库实际数据为准）
- [x] Task 调度器（robfig/cron v3 + Quartz 秒级 parser + 注册表执行；规避踩坑 5 的从库补载设计）
- [x] Task 任务 7 端点（list/getInfo/add/edit/changeStatus/run/remove；export 随 deviations #15）
- [x] Task 任务日志 4 端点（list/getInfo/remove/clean）
- [x] Task invokeTarget 解析器（语法兼容 Java、参数类型推断一致、注册表执行）
- [x] 单元测试：ParseTarget 4 形态 + Registry（2 个测试）
- [x] 端到端：任务列表/run 立即执行写日志/changeStatus 暂停恢复（2026-09-24 通过；测试日志清理）
