# Checklist · 01 基础设施

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。

## 验收清单

- [x] tasks.md 的 8 个 Task 全部勾选
- [x] 一个临时演示 handler（分页+权限+日志+导出全挂上）端到端联动验证
  - 注：`internal/router/integration_test.go`——sqlite 内存库 15 行演示表，/demo/list 走分页（TableDataInfo rows/total 顶层字段，默认页 10/15）+ 权限（admin 放行、普通用户 403 信封）+ 操作日志切面；/demo/export 走 Excel 导出（spreadsheetml Content-Type + download-filename 头）+ 日志（BusinessTypeExport 落库）；异步落库轮询断言 2 条日志字段正确
- [x] `go test ./...` 全绿
  - 注：14 个包全部 ok（含新增 aspect 19 个、excel 7 个、rate_limit 6 个、xss 5 个、router 3 个测试）；`gofmt -l .` 无输出、`go vet ./...` 零告警
- [x] 更新 specs/README.md：本模块状态改 ✅ + 日期

## 测试数据清理记录

- 全部验证使用 sqlite 内存库 / miniredis / httptest，未写入共用 MySQL/Redis，无需清理。
- 唯一一次真服务冷启动（swagger 端到端验证）只读访问了 /swagger-ui、/v3/api-docs，无业务数据写入；服务已停止（taskkill server.exe），8080 端口已释放。
