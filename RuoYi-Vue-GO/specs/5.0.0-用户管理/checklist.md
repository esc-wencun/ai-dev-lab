# Checklist · 05 用户管理

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 契约先行完成：13 端点 API 清单对照 Java 源码写实（见 spec.md）
- [ ] RuoYi-Vue3 用户管理页面（views/system/user）增删改查/授权/导入导出正常
  - 注：curl 层全部通过；浏览器页面级操作与 02/03/04 一并待用户确认
- [x] `go test ./...` 全绿（16 包 ok）
- [x] specs/README.md 状态更新为 ✅ + 日期
- [x] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

- 2026-09-24：sys_user 测试用户 testuser05 物理删除；Redis login_tokens:/captcha_codes: 清空；导出文件为临时目录产物不落库。
