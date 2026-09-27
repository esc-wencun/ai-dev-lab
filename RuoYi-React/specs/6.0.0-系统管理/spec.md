# Spec 6.0.0 系统管理

>
> **状态：🔶 进行中（2026-09-27 起）**。useCrud 范式 + DictTag 批次 C 核心件已落地；参数设置页完整 CRUD 已端到端验证（新增→列表刷新→删除确认框「是否确认删除编号为"101"的数据项？」逐字一致→测试数据清理无残留）。其余七页按 post → dict → notice → dept → menu → role → user 顺序推进。
> **背景**：八个页面，由简入繁：post → config → dict → notice → dept → menu → role → user（user 最后：依赖部门树/导入/分配角色）。先落 useCrud 范式再铺页面。
> **契约侦察**：[../reference/01-pages-and-api-baseline.md](../reference/01-pages-and-api-baseline.md) §1.2/§2；组件行为 reference/03 §1。
> **依赖**：03 批次 C（Pagination/RightToolbar/DictTag/Auth/TreePanel/IconSelect/RichEditor/ExcelImportDialog/HeaderNotice）。

## 范围

- `hooks/useCrud.ts`：查询/表格/分页/重置/多选 ids/删除确认/导出/刷新统一范式。
- 八页面 + 四个子路由页（authRole/authUser/dict-data/jobLog 在 7.0.0）+ 两个抽屉/弹窗详情。

## 关键行为契约（确认框文案逐字对齐基准）

1. **useCrud 范式**：查询参数（pageNum/pageSize + 业务字段 + addDateRange）；删除确认「是否确认删除编号为"xxx"的数据项？」；状态开关「是否确认改变"xxx"的状态?」；导出走 download()。
2. **post/config**：标准 CRUD + DictTag(sys_normal_disable / sys_yes_no) + config 刷新缓存「是否确认刷新参数缓存?」。
3. **dict**：类型列表 + 行内数据抽屉（700px Drawer 内嵌 CRUD）+ 独立路由 `/system/dict-data/index/:dictId` 两用 + 刷新缓存 + optionselect。
4. **notice**：RichEditor 富文本（HTML）+ 详情 dangerouslySetInnerHTML + ReadUsers 已读用户弹窗 + HeaderNotice 联动（发布 → 铃铛未读 → 已读）。
5. **dept/menu 树表**：antd Table treeData + 展开/折叠切换 + **批量排序**（勾选 + Sort → updateDeptSort/updateMenuSort，ids+orderNums）+ 新增时排除自身子树（listDeptExcludeChild）+ menu 的 IconSelect/三种类型字段。
6. **role**：菜单权限树（roleMenuTreeselect，提交含**半选** menuIds——getCheckedKeys+getHalfCheckedKeys 语义）+ 数据权限弹窗（deptTree + dataScope 五选）+ 分配用户子页（allocated/unallocated 两列表 + cancel/cancelAll/selectAll）+ 导出。
7. **user**：TreePanel 部门树过滤 + CRUD（部门 TreeSelect/岗位角色多选/sex DictTag）+ 重置密码 prompt 等价（「请输入"xxx"的新密码」）+ 导入导出（ExcelImportDialog action/templateAction）+ 详情 Drawer + 分配角色子页（reserve-selection 等价 rowSelection 保留勾选 + 前端分页）。
8. getUser(userId) 空参调 `/system/user/`（尾斜杠）拿角色岗位选项；changeStatus body 只含 {userId, status} 字段集。

## 设计决策

1. useCrud 返回查询/表格状态 + 操作方法，页面模板自由（不强求 CrudTable 统一组件——页面间表格样板差异大）。
2. 树表展开/折叠切换用受控 expandedRowKeys 实现（对齐基准 toggleExpandAll 行为）。
3. 半选提交是 role 页最容易丢的行为——单测一个「勾父带半选子」的提交 payload 组装函数。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
