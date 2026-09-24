# Checklist · 06 角色菜单

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 契约先行完成：22 端点 API 清单对照 Java 源码写实（见 spec.md）
- [ ] RuoYi-Vue3 角色管理/菜单管理页面增删改查正常（浏览器）
  - 注：curl 层全部通过；页面级操作与 02~05 一并待用户确认
- [x] `go test ./...` 全绿（16 包 ok）
- [x] specs/README.md 状态更新为 ✅ + 日期
- [x] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

- 2026-09-24：sys_role 测试角色（role_key=test06，role_id=100）物理删除（含 sys_role_menu 关联）；sys_menu 测试菜单（menu_id=2000）物理删除；Redis login_tokens:/captcha_codes: 清空。sys_role/sys_menu 预置数据未动。
