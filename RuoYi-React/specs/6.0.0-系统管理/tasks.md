# Tasks · 6 系统管理

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。
> 页面顺序：post → config → dict → notice → dept → menu → role → user。

- [x] hooks/useCrud.ts（查询/分页/多选/删除确认/导出/刷新范式）（2026-09-27）
- [x] post 页（CRUD + DictTag + 状态开关——基准状态列即 DictTag 无开关，已对齐）（2026-09-27）
- [x] config 页（CRUD + 内建参数 + 刷新缓存）——完整版端到端验证：新增 test.claude.verify 落库→列表刷新→删除确认框文案逐字一致→清理无残留；Auth 按钮权限/刷新缓存确认框就位（2026-09-27）
- [x] dict 类型页 + 数据抽屉 + 独立路由页两用（DictDataDrawer 组件复用）+ 刷新缓存——浏览器验证：类型列表 11 行真实数据、sys_user_sex 抽屉 3 条数据渲染（2026-09-27）
- [x] notice 页（RichEditor/详情 HTML/ReadUsers/HeaderNotice 联动验证）——页面端到端验证通过：富文本新增（Quill 工具栏 14 键渲染）→ 列表显示 → 详情 `<strong>` 加粗渲染（deviations #14 dangerouslySetInnerHTML）→ 测试公告 API 层清理无残留（2026-09-27）。HeaderNotice 顶部铃铛组件随 4.0.0 余项实现后补联动验证
- [x] dept 页（树表 + 展开折叠 + 批量排序 + excludeChild）——浏览器验证真实部门树渲染（2026-09-27）
- [x] menu 页（树表 + IconSelect + 三类型 + 批量排序）——浏览器验证菜单树渲染（2026-09-27）
- [x] role 页（菜单树半选提交 + 数据权限 + 分配用户子页）——浏览器验证角色列表 3 行渲染；半选 payload 单测遗留未做（随 3.0.0 批次收口）
- [x] user 页（部门树过滤/CRUD/重置密码/导入导出/详情 Drawer/分配角色子页）——浏览器验证部门树+用户列表 4 行渲染（2026-09-27）
- [ ] 验证：admin 逐页 CRUD 全通（新增→查询→修改→删除落库正确）
- [ ] 验证：树表批量排序保存后顺序持久化
- [ ] 验证：角色半选提交 → 后端菜单树回显一致；分配用户/角色全流程
- [ ] 验证：用户导入（含 updateSupport 更新分支）/导出/重置密码后新密码可登录
- [ ] 验证：新增字典类型+数据 → 其他页 DictTag 立即正确显示
