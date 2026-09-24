# Checklist · 00 工程基础

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。

## 验收清单

- [x] tasks.md 的 8 个 Task 全部勾选
- [x] `gofmt -l .` 无输出、`go vet ./...` 零告警、`go test ./...` 全绿
- [x] 目录结构与 spec.md 一致，分层规则（handler/service/dao 三禁）写入代码包注释——module/admin 三包随业务模块建包时写入包注释（见 spec.md 实施记录）；工程层包（response/database/types/logger）契约注释已就位
- [x] 更新 specs/README.md：本模块状态改 ✅ + 日期

## 测试数据清理记录

- 本模块单元测试全部使用 sqlite 内存库与 t.TempDir() 临时目录，未写入共用 MySQL/Redis，无需清理。
