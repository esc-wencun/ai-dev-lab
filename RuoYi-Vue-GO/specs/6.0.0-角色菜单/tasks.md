# Tasks · 06 角色菜单

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：对照 SysRoleController（13）+ SysMenuController（9）+ ServiceImpl + Mapper XML 写实 API 清单
- [x] Task 角色 13 端点（list/export/getInfo/optionselect/add/edit/dataScope/changeStatus/remove/allocatedList/unallocatedList/cancel+cancelAll+selectAll）
  - 注：admin(role_id=1) 保护、已分配保护、名称/权限字符唯一（文案逐字）；menuCheckStrictly 语义的 checkedKeys 查询
- [x] Task 菜单 7 端点（list/getInfo/treeselect/roleMenuTreeselect/add/edit/remove）
  - 注：menu_id=1 保护、上级不能自己、同父同名、子菜单/已分配删除保护；updateSort 未实现（前端无调用，登记 deviations）
- [x] Task 在线用户权限刷新（refreshPermissionByRoleId 的 Java 定制能力）
  - 注：未实现 SCAN 刷新——采用重新登录生效策略，登记 deviations（与 #11 一致）；2.0.0 spec 预留的接口在此闭环
- [x] 端到端：角色增（含菜单关联）/key 重复/已分配删/admin 保护/物理删；菜单增/删/treeselect/roleMenuTreeselect checkedKeys
  - 注：2026-09-24 全部通过；测试角色 test06/菜单 2000 已清除、Redis 会话清空
