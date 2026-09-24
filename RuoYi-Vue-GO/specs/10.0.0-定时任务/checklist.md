# Checklist · 10 定时任务

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 契约先行完成：11 端点 API 清单对照 Java 源码写实（见 spec.md）
- [ ] RuoYi-Vue3 定时任务页面（列表/暂停恢复/执行一次/日志）正常（浏览器）
  - 注：curl 层全部通过；页面级操作与 02~09 一并待用户确认
- [x] `go test ./...` 全绿（17 包 ok，新增 scheduler 2 测试）
- [x] specs/README.md 状态更新为 ✅ + 日期
- [x] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

- 2026-09-24：sys_job_log 测试日志（job_log_id>=15，含 logger 缺陷期 2 条失败记录）删除；sys_job 预置 3 条未动（status 恢复原值）；Redis 会话/验证码键清空。
