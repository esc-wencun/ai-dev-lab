# 06 角色管理 + 菜单管理

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：`SysRoleController`（13 端点）+ `SysMenuController`（9 端点）+ ServiceImpl + Mapper XML
> 前端页面：`views/system/role`、`views/system/menu`
> 依赖：1.0.0-基础设施、4.0.0-部门岗位、5.0.0-用户管理（分配用户列表）

## API 清单（已对照 Java 源码核实，2026-09-24）

### 角色 /system/role

| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:role:list | TableDataInfo；roleName/roleKey like + status |
| 2 | POST | /export | system:role:export | xlsx（序号/名称/权限/顺序/状态/创建时间） |
| 3 | GET | /optionselect | system:role:query | 正常状态角色 |
| 4 | GET | /{roleId} | system:role:query | 详情 + menuIds（menuCheckStrictly 排除父）+ deptIds |
| 5 | POST | / | system:role:add | 名称/权限字符唯一（文案逐字）；事务主表+菜单关联 |
| 6 | PUT | / | system:role:edit | admin(role_id=1)保护 `不允许操作超级管理员角色`；关联重置 |
| 7 | PUT | /dataScope | system:role:edit | 数据权限；scope=2 时重置 sys_role_dept |
| 8 | PUT | /changeStatus | system:role:edit | admin 保护 |
| 9 | DELETE | /{roleIds} | system:role:remove | `x已分配,不能删除`；事务清菜单+部门关联，物理删 |
| 10 | GET | /authUser/allocatedList | system:role:list | 已分配用户 |
| 11 | GET | /authUser/unallocatedList | system:role:list | 未分配用户 |
| 12 | PUT | /authUser/cancel | system:role:edit | body {userId,roleId} |
| 13 | PUT | /authUser/cancelAll + /authUser/selectAll | system:role:edit | query roleId/userIds 逗号分隔 |

### 菜单 /system/menu

| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /list | system:menu:list | 扁平不分页；menuName like/status |
| 2 | GET | /{menuId} | system:menu:query | 详情 |
| 3 | GET | /treeselect | 无 | TreeSelect {id,label,children} |
| 4 | GET | /roleMenuTreeselect/{roleId} | system:menu:query | checkedKeys + menus |
| 5 | POST | / | system:menu:add | 同父同名 `新增菜单'x'失败，菜单名称已存在` |
| 6 | PUT | / | system:menu:edit | menu_id=1 保护 `不允许操作超级管理员角色`语义对应菜单 `不允许操作超级管理员角色`（Java checkMenuAllowed 为 menu_id=1）；上级不能自己 |
| 7 | DELETE | /{menuId} | system:menu:remove | `存在子菜单,不允许删除`/`菜单已分配,不允许删除`；物理删 |

（menu updateSort 为本 Java 版定制、前端无调用，未实现——同 04 updateSort 登记 deviations）

## 实施记录

- 2026-09-24 完成。角色 edit 成功后对在线用户的权限刷新：Go 版当前采用"用户重新登录生效"策略（deviations #11 一致），6.0.0 范围内不实现 refreshPermissionByRoleId 的 SCAN 刷新（Java 定制能力，登记 deviations）。
- 端到端：角色增（菜单关联写入）/key 重复拒/已分配拒/admin 保护/物理删；菜单增/删/treeselect/roleMenuTreeselect（checkedKeys 正确）。
