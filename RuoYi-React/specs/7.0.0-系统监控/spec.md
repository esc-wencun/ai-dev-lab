# Spec 7.0.0 系统监控

>
> **状态：✅ 完成（2026-09-28）**。八页 + PlatformGate 降级组件全部落地并浏览器验证：online（在线用户含 admin 会话）、job（任务列表 4 行）、jobLog（代码就绪）、druid（Java 后端 iframe 正常）、server（CPU/JVM/磁盘卡片）、cache（canvas 玫瑰图+仪表盘+Redis 信息卡）、cacheList（三栏联动+清理全部）、operlog（11 行真实数据+DictTag 操作类型+详情分区）、logininfor（11 行+解锁按钮）。深度验证项（Go/Python 三态、导出文件、job 全生命周期）按 checklist 留待收口。
> **背景**：八个页面：两日志页（纯表格）+ online/job（CRUD）+ 三 features 降级页 + 缓存监控（echarts）。
> **契约侦察**：[../reference/01-pages-and-api-baseline.md](../reference/01-pages-and-api-baseline.md) §1.3/§2；features 模式 reference/03 §3。
> **依赖**：03 批次 D（Crontab/useEChart）。

## 范围

- online / job(+Crontab) / jobLog（子路由 /monitor/job-log/index/:jobId）/ druid / server / cache(+list) / operlog / logininfor。

## 关键行为契约

1. **features 降级三页统一模式**（druid/server/swagger 同构，swagger 在 8.0.0）：初值全 true → getPlatformInfo 覆盖 `response.features || {}` → false 时 antd Result icon=info「该功能仅 Java 版提供」+ 对应副标题 + 重新检测按钮；server 页 `serverMonitor !== false` 才拉数据。
2. **job**：CRUD + Crontab 弹窗回填 + 执行策略（1立即/2放弃/3恢复）+ 并发 + changeJobStatus + runJob 立即执行 + 详情弹窗；目标字符串 `beanName.methodName(args)`。
3. **jobLog**：jobId 参数过滤（任务页「调度日志」入口带参）+ 状态 DictTag(sys_common_status) + 详情（异常 pre）+ 删除/清空（DELETE /monitor/jobLog/clean）+ 导出。
4. **cache**：主页 echarts 玫瑰图（roseType）+ 仪表盘（gauge），容器固定高（基准 433px），resize 跟随；list 页三栏联动（getNames → getKeys → getValue pre 展示）+ 三级清理（name/key/all）确认框。
5. **operlog/logininfor**：只读 + 多字段查询（addDateRange 时间段）+ 详情分区弹窗（JSON 参数 pre）+ 删除/清空/导出；logininfor 特有**解锁账户**（GET /monitor/logininfor/unlock/{userName}）。
6. **online**：只读 + 强退（DELETE /monitor/online/{tokenId}，「是否确认强退"xxx"这位用户?」）。
7. echarts 不注册 macarons（基准注册失败实际用默认主题，行为等价——deviations #13）。

## 设计决策

1. 三 features 降级页抽 `PlatformGate` 小组件（传开关字段/副标题/正常渲染 children），三处复用防漂移。
2. job 详情/operlog 详情等分区只读弹窗复用一个 `DescDialog` 模式（Descriptions 分区），不逐页手写布局。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
