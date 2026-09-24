# Checklist · 12 平台标识

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [ ] Java：`mvn compile` 通过（附日期）
- [ ] GO：`go build ./...` 通过；`go test ./...` 全绿（附日期）
- [ ] Python：端到端 curl 验证（登录 → Bearer 调 /getPlatformInfo → 200 信封字段对齐；无 token → 401 信封）（附日期）
- [ ] GO：端到端 curl 验证（同上）（附日期）
- [ ] Java：端到端 curl 验证（同上）（附日期）
- [ ] 前端：`npm run dev` 编译通过；Java 后端下两页正常加载；Go/Python 后端下数据监控/服务监控按 features 降级提示（附日期）
- [ ] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

（暂无——端到端验证开始后在此登记）
