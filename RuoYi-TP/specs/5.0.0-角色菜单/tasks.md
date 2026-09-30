# 5.0.0-角色菜单 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：模型/service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端（含 3.0.0 遗留的权限两通道隔离验证）→ 收尾。
> 每个 Task 可独立验收。

## Task 1 · 模型与角色服务 RoleService

- [x] 模型：SysRole 不建独立模型（RoleService 内 Db::table 直操作，同 user_role/user_post 先例；RoleService 类注释注明）（think-orm 模型，表 sys_role；时间显式写，同 3.0.0 风格）
- [x] selectRoleList(filter, user)：**对位 selectRoleContactVo**——distinct r 9 列 + left join sys_user_role/sys_user/sys_dept + del_flag='0' + roleName/roleKey like、status、dataScope 精确、beginTime/endTime（date_format 口径）+ DataScope::apply(…, 'd')（admin 外）；**返回 Query 构造器**（分页/全量由调用方决定，同 PostService 踩坑教训）
- [x] selectRoleById（全列 selectRoleVo，del_flag='0'）
- [x] checkRoleNameUnique / checkRoleKeyUnique（全局唯一 del_flag='0' limit 1，roleId 自身放行）→ bool
- [x] countUserRoleByRoleId（sys_user_role count）
- [x] insertRole（事务：insert sys_role + batch insert sys_role_menu；menuIds 空跳过 rows=1）；updateRole（事务：update + delete role_menu + 重插）
- [x] authDataScope（事务：update sys_role + delete sys_role_dept + 重插 deptIds）
- [x] deleteRoleByIds（防护在控制器逐个；服务层事务删关联+软删）——原「先逐个 checkRoleAllowed」上移控制器 checkRoleAllowed + checkRoleDataScope + 已分配拒删（整批）→ 事务内 delete role_menu in + delete role_dept in + 软删 del_flag='2'
- [x] changeStatus（update status+update_time 仅 status 列 + update_time）
- [x] checkRoleAllowed（roleId=1 →「不允许操作超级管理员角色」）；checkRoleDataScope（非 admin 按 roleId 走带 DataScope 的 selectRoleList，查空 →「没有权限访问角色数据！」）
- [x] insertAuthUsers（selectAll 批插无去重）/ cancel / cancelAll（控制器直删，逻辑简单不进 service） sys_user_role 无去重原样）/ deleteAuthUser / deleteAuthUsers（物理删）
- [x] PHPUnit：ids 解析与唯一口径并入 curl 级实测（双唯一/自身放行实测过）；checkRoleAllowed 边界端到端实测（admin 拦截文案）（自身放行/他行判重）；menuIds/deptIds 逗号串解析与空数组语义；checkRoleAllowed 边界

## Task 2 · MenuService 扩展 + DeptService.roleDeptTreeData

- [x] selectMenuAll($user)：admin = selectMenuVo 13 列全量；非 admin = join role_menu/user_role/role（**r.status 条款核对**：经典版 selectMenuAllByUserId **不过滤角色状态**——以 XML 为准原样）+ distinct；不筛 menu_type/visible
- [x] selectMenuList（admin/非 admin 双分支合一，非 admin join 三表）（menu/list 非 admin 分支，带 menuName/visible 过滤）；selectMenuList（admin 分支）
- [x] roleMenuTreeData(roleId|null, $user)：Ztree 数组；name 拼 perms 灰字 `&lt;font color="#888"&gt;…&lt;/font&gt;`；checked = in_array(menu_id.perms, roleMenuList, true)（整串相等，concat 查询）；roleId 空 → 全不勾
- [x] menuTreeData($user)：同数据源、name 不带 perms、checked 恒 false
- [x] selectMenuById（含 parent_name 子查询）；checkMenuNameUnique（同 parent_id 下同名，menuId 放行）；selectCountMenuByParentId / selectCountRoleMenuByMenuId
- [x] insertMenu / updateMenu（动态 set + update_time）；updateMenuSort（事务，异常 →「保存排序异常，请联系管理员」）；deleteMenuById（物理删 `menu_id=? or parent_id=?` 原样）
- [x] RoleService::deptTreeData（收在 RoleService 而非 DeptService——角色域内聚；复用 DeptService::selectDeptList）(roleId|null, $user)：复用 selectDeptList（status='0' 在 service 内过滤——核对 initZtree 的 DEPT_NORMAL 条款）+ checked = in_array(dept_id.dept_name, roleDeptList, true)
- [x] 既有 menusOf / buildTree **不动**（导航树不受影响，回归验证主框架菜单渲染）
- [x] Ztree 组装口径端到端实测（checked=[103,104]/[1,2,100,101] 精确命中；perms 空串 concat 边界=纯 id）——PHPUnit 不重复设断言（checked 整串相等口径、perms 空串 concat 无歧义——**注意 concat(menu_id, '') 与 menu_id+'1' 粘连边界：perms 为空时串=纯 id**）；树数据不筛 M/C/F/visible 的字段面

## Task 3 · 控制器与路由（role 22 + menu 13 路由）

- [x] RoleController 22 个存活方法（selectMenuTree 死端点不复刻）+ #[Perm] 19 处（spec 表逐一对应：#8 authDataScope GET 无注解、#11/#12 check 无注解）+ #[Log] 9 处（5导出/1新增/2修改×3/3删除/4授权×3）
- [x] authDataScopeSave 成功后刷本人会话（refreshOwnSession：middleware session_uuid + cookie 兜底）（重算 rolesOf + SessionService::write；实现时核对会话 uuid 从中间件/cookie 取法）
- [x] MenuController 13 方法 + #[Perm] 8 处（icon/check/roleMenuTreeData/menuTreeData/selectMenuTree 5 处无注解）+ #[Log] 4 处
- [x] list 三处输出驼峰键（role 9 列 / allocated·unallocated 10 列 / menu 13 列）；树数据 4 处裸数组；check 三处裸 boolean
- [x] 路由注册（authUser 全固定段先于 :roleId——unallocatedList 曾被 :roleId 吞掉，实测修复） 先于 `authUser/:roleId`（遮蔽坑）；menu add/edit/remove 路径参数对位
- [x] 后端校验文案（roleInput/menuInput 手工校验器）：角色名 30/权限字符 100/菜单名 50/url 200/perms 100/显示顺序必填——手工校验器统一抛文案
- [x] **未挂 RepeatSubmit**；全局类 use 齐全（AjaxResult/TableDataInfo/TpConstant/各 Service）
- [x] curl 级 22 项自测全过（Python 脚本；unallocatedList whereNotIn 子查询修复后 code=0 total 语义正确）：list TableDataInfo / 树数据裸数组 / check 裸 bool / remove 已分配 500 文案 / changeStatus 停启用 / authDataScope 五种 data_scope 落库 / cancel·cancelAll·selectAll 写 sys_user_role / menu remove 双 warn 301

## Task 4 · 页面模板（12 页）

- [ ] role 7 页（index/add/edit/dataScope/authUser/selectUser/view）：JS 请求约定逐条对齐 spec 页面清单（openTab 分配用户、viewUrl 详情、dataScope=2 树显隐、statusTools 开关、roleId!=1 操作列空、btn 权限串 add/remove 原样）
- [ ] menu 5 页（index/add/edit/tree/icon）：树表 code/parentCode/uniqueId、类型 M/C/F 联动、图标片段 include、saveSort 仅提交改动行、「主菜单不能选择」「菜单已经存在」
- [ ] check_perm 接线两通道：shiro:hasPermission → {if check_perm}；JS 变量 editFlag 等 → {:check_perm(...) ? '' : 'hidden'}；sys_normal_disable / sys_show_hide 字典双键输出（JS 消费驼峰 / volist 消费下划线，同 3.0.0）
- [ ] include 片段核对：ztree-css/ztree-js 在 app/view/system/include/ 已有；icon 片段被 add/edit include；layer 相对路径渲染（'role/index'、'menu/index'）
- [ ] 视图层静态自查：无 th:*/@{ 残留、check_perm 全 {:} 输出、volist/if 配平、字符串 '0' 判空未用 empty()

## Task 5 · 端到端验收（浏览器级）——含 3.0.0 遗留权限隔离验证

- [ ] 角色列表：分页/搜索（名称/权限字符/状态/时间区间）/导出→下载 xlsx（6 列读回断言）/状态开关 changeStatus
- [ ] 角色新增/修改：菜单树勾选（父子联动 checkbox）、remote 双校验即时提示、修改后 sys_role_menu delete+重插 DB 核对
- [ ] 数据权限弹窗：五种 data_scope 逐一保存落库核对；=2 时部门树勾选写 sys_role_dept；切换清勾选行为
- [ ] 分配用户：allocated/unallocated 双列表（用预置 ry 用户）+ selectAll 授权 + cancel/cancelAll 取消；view 页用户卡片分页加载
- [ ] view 详情页：菜单权限计数与勾选树渲染、自定义部门树、关联用户卡片
- [ ] 菜单管理：树表渲染（预置 85 行三级+按钮层级）、新增 M/C/F 三类型联动、同父同名校验、updateSort、删除双 warn（有子/已分配）
- [ ] **权限两通道隔离验证（3.0.0 遗留，本模块主验收项）**：建测试角色（仅勾选部分菜单如 部门管理+岗位管理 部分 F 按钮）→ 建测试用户挂该角色 → 浏览器登录该用户：① 主框架仅出现被授权菜单；② 无权按钮不渲染（check_perm 通道）；③ 直接 POST 无权端点被 CheckPerm 拦截（500）；④ 数据权限过滤：该角色 data_scope=2/3/4 各验一，部门列表/角色列表行数实测过滤；⑤ 越权入口：直接访问 /system/role/edit/1 →「没有权限访问角色数据！」
- [ ] 「重新登录生效」语义实测（deviations 候选 1 拍板依据）：改测试角色菜单后，已登录会话权限不变、重登后生效——记录实测结果供 deviations 登记

## Task 6 · 收尾

- [ ] PHPUnit 全绿；spec.md 实施记录回填；checklist 逐项勾选；README 总表状态更新
- [ ] deviations 拍板与登记（候选 1 授权缓存/候选 2 死端点/候选 3 icon 片段）
- [ ] 测试数据清理并登记 checklist（见清理记录节）；预置数据复原核对：sys_role 2 行、sys_menu 85 行、sys_role_menu 85 行（role 2）、sys_role_dept 3 行（role 2→100/101/105）、sys_user_role 2 行、admin/admin123 可正常登录
