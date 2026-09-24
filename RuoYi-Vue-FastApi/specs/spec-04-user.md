# Spec-04 用户管理

> **状态：✅ 已完成（2026-09-24）**。Task 1-4 全部落地并端到端验证（导入导出于同日补齐：导出列对齐java @Excel注解、导入校验链对齐importUser、模板含示例行）：
> 列表分页+数据权限+部门联表、deptTree、详情(roleIds/postIds/posts)、新增(初始密码/唯一性/关联表)、
> 编辑、authRole回显+分配、停用启用+踢下线(_kick_user_sessions)、resetPwd、admin保护、逻辑删除。
>
> Java 版对应：`SysUserController` + `SysUserServiceImpl`
> 前端页面：`views/system/user/index.vue`
> 依赖：spec-01（分页/权限/日志/数据权限）、spec-03（部门树、岗位）

## 目标

用户 CRUD、分配角色/岗位、重置密码、状态切换、Excel 导入导出、个人数据权限过滤。这是管理模块里端点最多、校验最细的之一。

## API 清单

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/user/list` | list（分页+数据权限） | `system:user:list` |
| 2 | POST | `/system/user/export` | export | `system:user:export` |
| 3 | POST | `/system/user/importData` | importData | `system:user:import` |
| 4 | POST | `/system/user/importTemplate` | importTemplate（下载模板） | 登录即可 |
| 5 | GET | `/system/user/` 或 `/{userId}` | getInfo（详情+角色岗位下拉） | `system:user:query` |
| 6 | POST | `/system/user` | add | `system:user:add` |
| 7 | PUT | `/system/user` | edit | `system:user:edit` |
| 8 | PUT | `/system/user/resetPwd` | resetPwd | `system:user:resetPwd` |
| 9 | PUT | `/system/user/changeStatus` | changeStatus | `system:user:edit` |
| 10 | DELETE | `/system/user/{userIds}` | remove（批量） | `system:user:remove` |
| 11 | GET | `/system/user/deptTree` | deptTree（部门树下拉） | `system:user:list` |
| 12 | PUT | `/system/user/authRole` | insertAuthRole（分配角色） | `system:user:edit` |
| 13 | GET | `/system/user/authRole/{userId}` | authRole（分配角色页回显：用户信息+全部角色+选中 roleIds） | `system:user:edit` |

## Task 1: 用户查询

- [x] 列表分页 + 过滤（`userName / phonenumber / status / deptId / 时间区间 beginTime endTime`）+ 数据权限（spec-01 Task 4，Java 版 selectUserList 带 dataScope）
- [x] 返回驼峰字段，含 `dept.deptName`；password 不返回
- [x] `deptTree`：构建部门下拉树（`{id, label, children}`，与 Java selectDeptTree 一致）
- [x] 详情 getInfo：`data`（用户+roles）、`postIds`、`roleIds`、`roles`（全部角色，admin 过滤逻辑与 Java 一致）、`posts`
- [x] 数据范围校验 `checkUserDataScope`：非 admin 只能查权限内的用户（已实现 check_user_data_scope，挂在详情/编辑/删除三处，越权返回 `没有权限访问用户数据!`）

## Task 2: 用户新增/修改

- [x] 新增校验（顺序对齐 Java）：用户名唯一（`新增用户'%s'失败，登录账号已存在`）、手机号唯一、邮箱唯一；密码取 `sys.user.initPassword` 配置（123456）并 bcrypt；写入 `sys_user_post`、`sys_user_role` 关联
- [x] 修改：不允许操作 admin（`不允许操作超级管理员用户`）；唯一性校验（排除自身）；关联表重建；状态停用时**清除该用户的 Redis 会话**（对应 Java 同步踢下线逻辑）
- [x] `GET /authRole/{userId}`：返回用户信息 + 全部可选角色列表 + 该用户已选中的 roleIds（对照 Java 同名方法返回结构，供分配角色对话框回显）
- [x] `PUT /authRole`：重建 `sys_user_role`；成功后刷新该用户 Redis 会话的权限（立即生效）
- [x] 修改/删除后清 Redis 会话，保证权限即时生效
- [x] 端到端测试：建用户 → 分配角色岗位 → 登录（用初始密码）→ 被停用后会话失效

## Task 3: 状态/密码/删除

- [x] `changeStatus`：校验非 admin；更新 status；停用踢下线
- [x] `resetPwd`：非 admin 校验；bcrypt 新密码；更新 `pwd_update_date`；清会话
- [x] 删除：逻辑删除（del_flag='2'）；批量前逐个 admin 校验 + `check_user_data_scope`；删除前清 `sys_user_role` / `sys_user_post`
- [x] 端到端测试：批量删除后 del_flag 生效、关联表清理

## Task 4: Excel 导入导出

- [x] 导出：列对照 Java SysUser @Excel 注解（部门用 Type.EXPORT 联动 `dept.deptName`，性别/状态用 dictType 转换 `sys_user_sex` / `sys_normal_disable`——依赖 spec-06 字典，若先做则用内置映射，spec-06 完成后切换）
- [x] 导入：解析 xlsx → 校验（用户名唯一性等，行号提示 `%s%s 已存在`）；`updateSupport` 为 true 时更新已存在用户，否则报 `很遗憾，共 %d 条数据导入失败`
- [x] `importTemplate`：返回仅含表头的 xlsx
- [x] 端到端测试：导出→导入回灌→失败行提示正确

## 验收清单

- [x] RuoYi-Vue3 用户管理页面全部功能可用（含部门树筛选、导入导出对话框）
- [x] pytest 全绿；更新 specs/README.md 状态
