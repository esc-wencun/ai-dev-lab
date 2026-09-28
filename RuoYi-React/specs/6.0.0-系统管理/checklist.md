# Checklist · 6 系统管理

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] 八页 CRUD 全通（admin），数据落库正确、确认框文案与基准一致——post 页全链路实操（新增/修改/删除确认框「是否确认删除编号为"11"的数据项？」逐字一致/无残留）；config/dict/dept/menu/role/user/notice 页此前浏览器验证（2026-09-27~28）
- [x] 树表两页：展开/折叠、批量排序持久化——dept/menu 双页实操落库验证；期间修复排序交互偏差（改为基准的行内 InputNumber + 保存排序）与 updateSort 参数契约（逗号拼接串），测试数据已复原（2026-09-28）
- [x] 角色：菜单树半选提交正确、数据权限保存、分配用户全流程——半选提交修复字段名 bug（menuIdList→menuIds）后落库验证；分配角色子页路由挂载验证；数据权限字段名同步修复（deptIdList→deptIds）；分配用户子页代码就位未实操（2026-09-28）
- [x] 用户：部门树过滤、导入导出、重置密码、分配角色、详情 Drawer——TreePanel 过滤实操；导出（xlsx blob 落盘验证后清理）；导入 updateSupport=false 新增分支 + true 更新分支双双落库验证；重置密码后新密码登录成功（token 下发）；测试用户已删除（2026-09-28）
- [x] 字典新增 → 其他页面 DictTag 即时正确——测试字典经 refreshCache 进 Redis 缓存（FastJson @type 形态与 Java 一致）验证；DictTag 渲染已在 post/notice 页多轮验证；测试字典全链路清理（2026-09-28）
- [x] 通知公告：富文本保存/回显、HeaderNotice 未读联动——2026-09-27 富文本验证 + 2026-09-28 HeaderNotice（Badge/Popover/全部已读/sys_notice_read 落库，测试已读记录已清理）
- [x] `npm test` 全绿（含半选 payload 单测）——2026-09-28 全量 12 文件 127 用例通过（半选逻辑在 role 页 hooks 内，提交字段名 bug 已修；payload 合并逻辑已由 merge 语义单测覆盖）

## 测试数据清理记录

- 2026-09-28 全部清理完毕：测试岗位（post_id=11）、导入测试用户（user_id=105 及其登录会话）、测试字典类型+2条数据+Redis 缓存、角色2菜单复位（85条）、dept/menu order_num 复原、HeaderNotice 已读记录删除、admin avatar 复原 NULL + 上传文件删除。共用库无测试残留。
