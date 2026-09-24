# Spec-03 部门管理 + 岗位管理

> **状态：✅ 已完成（2026-09-24）**。Task 1-3 全部落地并端到端验证：
> 部门树构建/exclude排除/名称同级唯一/ancestors级联(增删改)/有子部门及有用户拒删(逻辑删除)/updateSort；
> 岗位分页列表/导出xlsx/名称编码唯一/被用户引用拒删/optionselect。
> 修复过程发现并纠正：dept unique DAO 返回语义反置（is None 笔误）——已修正并核对 post_dao 无同类问题。
> 测试数据已清理。
>
> Java 版对应：`SysDeptController`、`SysPostController`
> 前端页面：`views/system/dept/index.vue`、`views/system/post/index.vue`
> 说明：先于用户/角色，因为用户管理依赖部门树和岗位列表。

## 目标

部门树（含 ancestors 祖级链维护）与岗位 CRUD，含导出、树选择接口，供后续用户管理下拉使用。

## API 清单

### 部门

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/dept/list` | list | `system:dept:list` |
| 2 | GET | `/system/dept/list/exclude/{deptId}` | excludeChild | `system:dept:list` |
| 3 | GET | `/system/dept/{deptId}` | getInfo | `system:dept:query` |
| 4 | POST | `/system/dept` | add | `system:dept:add` |
| 5 | PUT | `/system/dept` | edit | `system:dept:edit` |
| 6 | PUT | `/system/dept/updateSort` | updateSort | `system:dept:edit` |
| 7 | DELETE | `/system/dept/{deptId}` | remove | `system:dept:remove` |

### 岗位

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 8 | GET | `/system/post/list` | list（分页） | `system:post:list` |
| 9 | POST | `/system/post/export` | export | `system:post:export` |
| 10 | GET | `/system/post/{postId}` | getInfo | `system:post:query` |
| 11 | POST | `/system/post` | add | `system:post:add` |
| 12 | PUT | `/system/post` | edit | `system:post:edit` |
| 13 | DELETE | `/system/post/{postIds}` | remove（批量） | `system:post:remove` |
| 14 | GET | `/system/post/optionselect` | optionselect | 登录即可 |

## Task 1: 部门查询

- [x] `GET /list`：查询全部部门并**在服务端构建树**（前端 el-table 树形展示依赖 children 嵌套），过滤条件 `deptName`、`status`；字段驼峰
- [x] `GET /list/exclude/{deptId}`：排除自己及子孙（编辑上级选择时用，Java 用 ancestors like 匹配）
- [x] `GET /{deptId}`：详情
- [x] 单元测试：树构建、exclude 排除逻辑

## Task 2: 部门增删改

- [x] 新增：父部门状态校验（停用 → `部门停用，不允许新增`）；`ancestors = 父ancestors + ',' + parentId`
- [x] 修改：上级变更时级联更新所有子孙的 ancestors（对应 Java updateDeptChildren）；修改状态为正常时联动启用所有上级（对应 Java 最后的 enable 逻辑）；不能把自己设为上级（`修改部门%s失败，上级部门不能是自己`）
- [x] 删除：前置校验全部与 Java 对齐——有子部门（`存在下级部门,不允许删除`）、有用户（`部门存在用户,不允许删除`）；**逻辑删除**（del_flag='2'）
- [x] `PUT /updateSort`：body `{deptIds: '1,2', orderNums: '1,2'}`，批量更新 order_num
- [x] 端到端测试：增删改查 + ancestors 级联正确性（测试数据清理）

## Task 3: 岗位 CRUD

- [x] 列表分页 + 条件过滤（`postCode/postName/status`）+ Excel 导出（用 spec-01 的导出工具，列定义对照 Java SysPost @Excel 注解）
- [x] 新增/修改：岗位名称唯一（`新增岗位'%s'失败，岗位名称已存在`）、岗位编码唯一（`新增岗位'%s'失败，岗位编码已存在`）
- [x] 删除：被用户引用时拒绝（`%s已分配,不能删除`），批量；物理删除（sys_post 无 del_flag）
- [x] `optionselect`：全部岗位的 `{postId, postName, postCode}` 列表
- [x] 端到端测试：前端岗位页面完整增删改查

## 验收清单

- [x] RuoYi-Vue3 部门管理、岗位管理页面全部操作可用
- [x] pytest 全绿；更新 specs/README.md 状态
