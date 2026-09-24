# Tasks · 09 监控日志

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：对照在线用户/缓存/操作日志/登录日志四 Controller 写实 API 清单
- [x] Task 在线用户 2 端点（list 聚合 + forceLogout 强退）
- [x] Task 缓存监控 6 端点（info/getNames/getKeys/getValue/clearCacheKey/clearCacheAll）
- [x] Task 操作日志 3 端点（list/remove/clean；数据源=01 操作日志切面）
- [x] Task 登录日志 4 端点（list/remove/clean/unlock 清锁定计数）
- [x] Task 服务监控 /monitor/server
  - 注：未实现（Java 依赖 OSHI+JVM；Go 版 gopsutil 方案待设计，登记 deviations #17）
- [x] 端到端：operlog/logininfor/online/cache 端点列表返回正确（2026-09-24 通过；未造数据，在线用户来自登录态本身）
