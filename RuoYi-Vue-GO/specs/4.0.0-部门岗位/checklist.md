# Checklist · 04 部门岗位

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 契约先行完成：API 清单已对照 Java 源码逐端点写实（见 spec.md）
- [ ] RuoYi-Vue3 对应页面（views/system/dept、views/system/post）增删改查正常，字段与状态码约定与 Java 一致
  - 注：curl 层字段/信封/文案已逐项对照 Java 并通过；浏览器页面级操作与 02/03 一并待用户确认
- [x] `go test ./...` 全绿（16 包 ok）
- [x] specs/README.md 状态更新为 ✅ + 日期
- [x] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

- 2026-09-24：sys_dept 测试部门（dept_id=200）删除（先软删验证 del_flag=2 再物理清除）；sys_post 测试岗位（post_code=test04）物理删除；Redis login_tokens:/captcha_codes: 清空。sys_dept/sys_post 预置数据（100~103/1~4）未动。
