# Spec-05 角色管理 + 菜单管理

> **状态：✅ 已完成（2026-09-24）**。Task 1-4 全部落地并端到端验证：
> 角色CRUD(名称/权限字符唯一)、roleMenuTree勾选裁剪(menuCheckStrictly)、分配用户三接口、
> 数据权限dataScope+deptTree回显、admin角色保护、逻辑删除、菜单CRUD、权限变更后会话刷新。
> 测试数据已清理。
>
> Java 版对应：`SysRoleController`、`SysMenuController`
> 前端页面：`views/system/role/`、`views/system/menu/`
> 依赖：spec-01、spec-03、spec-04（角色分配用户依赖用户列表查询）
> 说明：这是权限体系的核心，菜单路由构建（buildMenus）已在 Phase 0 实现并测试，本 spec 复用。

## 目标

角色 CRUD、数据权限配置、分配用户/取消分配；菜单 CRUD、树选择、角色菜单授权；改权限后**刷新在线用户会话权限**（对应 Java refreshPermissionByRoleId）。

## API 清单

### 角色

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 1 | GET | `/system/role/list` | list（分页） | `system:role:list` |
| 2 | POST | `/system/role/export` | export | `system:role:export` |
| 3 | GET | `/system/role/{roleId}` | getInfo | `system:role:query` |
| 4 | POST | `/system/role` | add | `system:role:add` |
| 5 | PUT | `/system/role` | edit | `system:role:edit` |
| 6 | PUT | `/system/role/dataScope` | dataScope（授权数据范围） | `system:role:edit` |
| 7 | PUT | `/system/role/changeStatus` | changeStatus | `system:role:edit` |
| 8 | DELETE | `/system/role/{roleIds}` | remove（批量） | `system:role:remove` |
| 9 | GET | `/system/role/optionselect` | optionselect | `system:role:query` |
| 10 | GET | `/system/role/authUser/allocatedList` | 已分配用户列表（分页） | `system:role:list` |
| 11 | GET | `/system/role/authUser/unallocatedList` | 未分配用户列表（分页） | `system:role:list` |
| 12 | PUT | `/system/role/authUser/cancel` | 取消单个授权 | `system:role:edit` |
| 13 | PUT | `/system/role/authUser/cancelAll` | 批量取消 | `system:role:edit` |
| 14 | PUT | `/system/role/authUser/selectAll` | 批量授权 | `system:role:edit` |
| 15 | GET | `/system/role/deptTree/{roleId}` | 角色部门树（数据权限弹窗） | `system:role:query` |

### 菜单

| # | 方法 | 路径 | Java 对应 | 权限 |
|---|------|------|-----------|------|
| 16 | GET | `/system/menu/list` | list（平铺列表，服务端组树由前端展示？——以 Java 为准：返回平铺，前端构建） | `system:menu:list` |
| 17 | GET | `/system/menu/{menuId}` | getInfo | `system:menu:query` |
| 18 | GET | `/system/menu/treeselect` | 树选择（`{id,label,children}`） | 登录即可 |
| 19 | GET | `/system/menu/roleMenuTreeselect/{roleId}` | 树选择+选中态（checkedKeys） | 登录即可 |
| 20 | POST | `/system/menu` | add | `system:menu:add` |
| 21 | PUT | `/system/menu` | edit | `system:menu:edit` |
| 22 | PUT | `/system/menu/updateSort` | updateSort | `system:menu:edit` |
| 23 | DELETE | `/system/menu/{menuId}` | remove | `system:menu:remove` |

## Task 1: 角色 CRUD

- [x] 列表分页（`roleName / roleKey / status / 时间区间`）+ 导出（对照 Java SysRole @Excel 列）
- [x] 新增/修改校验：角色名唯一（`新增角色'%s'失败，角色名称已存在`）、角色权限字符唯一（`...角色权限字符已存在`）；不允许修改 admin 角色（`不允许操作超级管理员角色`）
- [x] 新增/修改同时写 `sys_role_menu`；修改后调用会话权限刷新（Task 4）
- [x] `changeStatus`：停用角色；删除：校验非 admin、逻辑删除 del_flag='2'、清 `sys_role_menu`
- [x] `dataScope`：更新角色的 data_scope 字段 + 重建 `sys_role_dept`（自定权限的部门勾选）
- [x] 单元测试：菜单勾选的父子联动展开（menuCheckStrictly 模式下 checkedKeys 的裁剪逻辑，对应 Java selectMenuListByRoleId 的 not-in-parent SQL）

## Task 2: 角色分配用户

- [x] `allocatedList` / `unallocatedList`：分页 + 过滤（userName / phonenumber），语义分别是在/不在该角色中的用户
- [x] `cancel` / `cancelAll` / `selectAll`：维护 `sys_user_role`
- [x] 端到端测试：分配 → allocatedList 出现 → cancelAll 后消失；被授权用户的 getInfo roles/permissions 变化

## Task 3: 菜单管理

- [x] `list`：平铺列表 + 过滤（menuName / status），**不排序嵌套**（Java list 直接返回按 parent_id, order_num 排序的列表）
- [x] 新增校验（对照 Java SysMenuServiceImpl checkXxxUnique 与校验顺序）：菜单类型规则（目录必须有 path；菜单必须有 path，component 可空=ParentView/InnerLink 逻辑）；路由地址/权限字符唯一性（checkRouteConfigUnique：同级 path 冲突 / 根级 path 冲突 / routeName 全局唯一，三种告警消息照抄）
- [x] 修改：不允许把 parent_id 设为自己（`上级菜单不能选择自己`）；核对结果：Java 当前版本仅防"选自己"不防子孙，已照抄；另补齐 checkMenuNameUnique 同级菜单名唯一（add/edit均生效）
- [x] 删除：有子菜单拒绝（`存在子菜单,不允许删除`）、菜单已分配给角色拒绝（`菜单已分配,不允许删除`）
- [x] `updateSort`：批量更新 order_num
- [x] 修改/删除后触发路由权限刷新（Task 4）
- [x] 端到端测试：建目录+菜单 → getRouters 出现新路由 → 删除后消失

## Task 4: 权限变更的会话刷新（对应 Java TokenService.refreshPermissionByRoleId）

- [x] 角色权限/菜单变更后：扫描 `login_tokens:*`，对拥有该角色的在线用户重算 permissions 并回写会话（admin 跳过）
- [x] 端到端测试：用户 A 在线 → 给其角色加菜单 → A 的 getInfo/getRouters 立即反映新权限，无需重新登录

## 验收清单

- [x] RuoYi-Vue3 角色、菜单页面全部可用；角色授权弹窗、数据权限弹窗、菜单树勾选正常
- [x] pytest 全绿；更新 specs/README.md 状态
