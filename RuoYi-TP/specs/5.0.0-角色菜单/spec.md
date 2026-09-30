# 5.0.0-角色菜单 · spec

> **状态：✅ 已完成（2026-09-29，curl/DB + 浏览器级 + 权限隔离验证通过）**
> 对位经典若依 SysRoleController / SysMenuController + SysRoleServiceImpl / SysMenuServiceImpl（+ SysDeptServiceImpl.roleDeptTreeData / SysUserServiceImpl.selectAllocated·UnallocatedList）+ templates/system/role|menu（角色 23 端点 / 菜单 13 端点 / 页面 12 个，逐方法核对实锤）。
> 依赖：1.0.0（#[Perm] / #[Log] / PageQuery / TableDataInfo / DataScope / LoginAuth）、3.0.0（DataScope 风格 / DeptService / ExcelExportService / DictService / /common/download）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下（按序号串行隐式满足）。4.0.0 未动工不阻塞本模块（本模块只读 sys_user / sys_user_role，用户管理侧的 authRole 分配入口归 4.0.0）。
> 调研依据：reference 源码逐文件核对（SysRoleController 23 方法 / SysMenuController 13 方法 / SysRoleServiceImpl / SysMenuServiceImpl / SysRole·SysMenu Mapper XML / SysUserRole·SysRoleDept·SysRoleMenu Mapper XML / SysDeptServiceImpl.roleDeptTreeData / SysUserMapper allocated·unallocated SQL / 页面 HTML×12 / SysRole·SysMenu·Ztree 实体 / ry-ui.js $.tree·$.table）+ **ry-tp 库实测**（sys_role 12 列 / sys_menu 16 列（**无 del_flag**）/ sys_role_menu·sys_role_dept 复合主键；预置角色 2 行、菜单 85 行（M4/C19/F62）、sys_role_menu 85 行全挂 role 2、sys_role_dept 3 行（role 2→100/101/105）、sys_user_role 2 行（admin→1、ry→2）；role/menu 权限串 sys_menu.perms 共 11 行）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、POST 分页参数、信封结构、权限两通道（#[Perm] 注解 + check_perm 模板函数）。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/system/RoleController.php | SysRoleController | 23 个方法，其中 **selectMenuTree 为死端点不复刻**（见 API 清单 #13 与拟登记 deviations 候选 2），实落 22 |
| app/controller/system/MenuController.php | SysMenuController | 13 个方法（含 roleMenuTreeData/menuTreeData 树数据、icon 图标页、updateSort） |
| app/service/RoleService.php | ISysRoleService / SysRoleServiceImpl | 列表（DataScope）、双唯一校验、role_menu/role_dept 关联维护（事务）、checkRoleAllowed / checkRoleDataScope、authUser 系写入 |
| app/service/MenuService.php（**扩展**） | ISysMenuService / SysMenuServiceImpl | 2.0.0 已有 menusOf/buildTree；本模块补 selectMenuAll（含 F 型、含隐藏，admin 全量/非 admin 按角色关联）、roleMenuTreeData / menuTreeData（Ztree 组装）、checkMenuNameUnique、insert/update/updateSort/deleteMenu |
| app/service/DeptService.php（**扩展一个方法**） | SysDeptServiceImpl.roleDeptTreeData + initZtree | 角色数据权限部门树（Ztree，checked=role_dept 勾选；复用 selectDeptList 带 DataScope） |
| app/model/SysRole.php（think-orm 模型，薄） | SysRole 实体 | 表 sys_role；菜单/部门组、permissions 等非表字段由 service 数组承载，不进模型 |
| app/view/system/role/*.html ×7、app/view/system/menu/*.html ×5 | templates/system/role/ ×7、templates/system/menu/ ×5 | 12 页，静态 JS（bootstrap-table / bootstrap-tree-table / zTree / jquery.validate / layer）零改动对接 |
| Excel 导出 | ExcelUtil（本模块消费 SysRole 的 6 个 @Excel 列） | 复用 3.0.0 ExcelExportService + /common/download，无新通用能力 |

**范围外**（后续模块，勿在本模块实现）：用户管理页的「更多→分配角色」（GET/POST user/authRole，归 4.0.0）；selectRolesByUserId / selectRoleKeys（登录链路已在 LoginService.rolesOf 自实现，4.0.0 authRole 回显时若需再扩展）；字典管理页面与 sys_show_hide 缓存失效联动（6.0.0）；demo/tool/swagger 菜单对应页面（deviations #4/#5，菜单行存在但点击 404 属有意行为）。

## 页面清单（12）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/树组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 角色列表页 | view/system/role/index.html | role/role.html | bootstrap-table（POST + server 分页） | `prefix = ctx + "system/role"`；options：url=prefix+/list、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove（**无 {id}**，ids 走 body）、exportUrl=prefix+/export、**viewUrl=prefix+/view/{id}**（roleName 链接点开详情）；sortName=roleSort；搜索字段 roleName/roleKey/status/params[beginTime]/params[endTime]（状态下拉服务端渲染 sys_normal_disable）；列 roleId/roleName(sortable,链接)/roleKey(sortable)/dataScope(徽章五态硬编码 formatter)/roleSort(sortable)/状态(statusTools 开关，editFlag 控制列 visible)/createTime(sortable)/操作列（**roleId!=1 才渲染**：编辑/删除/更多 popover（数据权限=authDataScope/分配用户=authUser））；状态开关 disable/enable → confirm 后 `$.operate.post(prefix+"/changeStatus", {roleId, status:1|0})` |
| 2 | 角色新增弹窗 | view/system/role/add.html | role/add.html | 表单 + zTree | 菜单权限树 `url = ctx+"system/menu/roleMenuTreeData"`（**无 roleId → 全不勾**）、expandLevel=0、check enable；三个控制 checkbox：展开/折叠 / 全选全不选 / 父子联动（chkboxType Y/N:"ps" ↔ ""）；validate remote：checkRoleNameUnique（仅 roleName）、checkRoleKeyUnique（仅 roleKey），文案「角色名称已经存在」「角色权限已经存在」；roleSort digits；提交**手工字段 ajax** POST prefix+/add：{roleName, roleKey, roleSort, status(勾=0/不勾=1), remark, menuIds}——menuIds 为 `$.tree.getCheckedNodes()` 返回的**逗号 join 字符串** |
| 3 | 角色修改弹窗 | view/system/role/edit.html | role/edit.html | 表单 + zTree | 树 `url = roleMenuTreeData?roleId=xx`（预勾选）；remote data 多带 roleId；提交手工字段 POST prefix+/edit：{roleId, roleName, roleKey, roleSort, status, remark, menuIds}；status checkbox 回显 th:checked=role.status=='0' |
| 4 | 角色数据权限弹窗 | view/system/role/dataScope.html | role/dataScope.html | 表单 + zTree | 角色名称/权限字符 **readonly**；数据范围 select 五项（1全部/2自定义/3本部门/4本部门及以下/5仅本人）；dataScope=2 时显示部门树区（切换非 2 时 `$._tree.checkAllNodes(false)` 清勾选再隐藏）；树 `url = ctx+"system/role/deptTreeData?roleId=xx"`（预勾选 sys_role_dept）、expandLevel=2、nocheckInherit、父子联动；提交 POST prefix+/authDataScope：{roleId, roleName, roleKey, dataScope, deptIds}（roleName/roleKey 随表单传回，readonly 原值无害） |
| 5 | 分配用户页 | view/system/role/authUser.html | role/authUser.html | bootstrap-table | **$.modal.openTab 页签打开**（非弹窗）；hidden roleId；搜索 loginName/phonenumber；表格 url=prefix+/authUser/allocatedList、queryParams 附加 roleId、sortName=createTime sortOrder=desc；列 userId(hidden)/loginName(sortable)/userName/email/phonenumber/status(字典徽章 sys_normal_disable)/createTime(sortable)/操作（取消授权，removeFlag 显隐）；工具栏：添加用户（**shiro:hasPermission="system:role:add"**）→ `$.modal.open(prefix+"/authUser/selectUser/"+roleId)`、批量取消授权（**system:role:remove**）→ confirm「确认要删除选中的N条数据吗?」POST cancelAll {roleId, userIds:selectFirstColumns().join()}、关闭（closeItem）；行内取消授权 → confirm「确认要取消该用户角色吗？」POST cancel {roleId, userId} |
| 6 | 选择用户弹窗 | view/system/role/selectUser.html | role/selectUser.html | bootstrap-table | url=prefix+/authUser/unallocatedList + queryParams 附加 roleId；showSearch/showRefresh/showToggle/showColumns 全 false、clickToSelect、rememberSelected；首列 state checkbox；无工具栏；submitHandler → `$.operate.save(prefix+"/selectAll", {roleId, userIds})` |
| 7 | 角色详情页 | view/system/role/view.html | role/view.html | **服务端注入数据 + 前端渲染**（无表格插件） | 变量 role / menuTree（Ztree 数组，JS 变量 menuNodes）/ deptTree（**仅 dataScope='2' 时 assign，否则 JS 侧 '[]'**）/ userCount；页面纯展示：基本信息 + 数据范围徽章 + 自定义部门树（dataScope=2，checked 节点高亮、未勾选祖先淡色路径）+ 菜单权限（计数 = menuNodes 中 checked===true 数；展开渲染**只显示勾选节点及其祖先链**，name 含 perms 的 font 灰字）+ 关联用户（userCount 徽章 + 「查看已分配用户」展开后 **ajax 分页拉 allocatedList（pageSize=100）渲染用户卡片**，前端搜索/加载更多）——allocatedList 端点在 view 页复用（需 system:role:list，与本页 #[Perm] 一致） |
| 8 | 菜单列表页 | view/system/menu/index.html | menu/menu.html | **bootstrap-tree-table 树表**（非分页） | `prefix = ctx + "system/menu"`；code=menuId/parentCode=parentId/uniqueId=menuId/expandAll=false/expandFirst=false；options：url=prefix+/list（POST，**返回裸数组**）、createUrl=prefix+/add/{id}（工具栏固定 0，行内传 menuId）、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove/{id}（**路径参数**，与 role 的 body ids 不同）；列：radio selectItem/菜单名称(icon+name)/排序（隐藏 input name=menuIds + 文本 input name=orderNums，originalOrders 记原值）/请求地址(tooltip)/类型徽章(M目录 label-success/C菜单 label-primary/F按钮 label-warning)/可见（**F 型显示 '-'**，其余字典 sys_show_hide 徽章，JS 变量 datas）/权限标识(tooltip)/操作（编辑/新增/删除按 flag 显隐）；保存排序 saveSort：对比 originalOrders **仅提交改动行** `$.operate.post(prefix+"/updateSort", {menuIds:"…", orderNums:"…"})`，无改动 alertWarning「未检测到排序修改」；搜索字段 menuName/visible |
| 9 | 菜单新增弹窗 | view/system/menu/add.html | menu/add.html | 表单（**图标选择器内嵌**） | parentId 隐藏域 treeId 初始=menu.menuId、上级菜单输入框 treeName 初始=menu.menuName；菜单类型 radio M/C/F **无默认选中**（required 校验）；类型联动（ifChecked）：M 只显 icon 区 / C 显 url+perms+icon+target+isRefresh / F 只显 perms；字段 url/target(menuItem 页签|menuBlank 新窗口)/perms/orderNum(required digits)/visible(radio，sys_show_hide 字典 volist 渲染，is_default 勾选)/isRefresh(radio 1否 0是，默认否)；图标：input icon focus 弹 .icon-drop（`{include file="menu/icon"}` 片段），点击 i 取 class 回填；validate remote：checkMenuNameUnique data {parentId, menuName}，文案「菜单已经存在」；提交 **$('#form-menu-add').serialize() 整表单** POST prefix+/add；选上级菜单 → `$.modal.openOptions` prefix+/selectMenuTree/{treeId>0?treeId:1} |
| 10 | 菜单修改弹窗 | view/system/menu/edit.html | menu/edit.html | 表单 | 同 add + menuId 隐藏域；上级菜单显示 parentName（null→**「无」**）；remote data 多 menuId；页面加载即按当前 menuType 调 menuVisible() 联动显隐；选上级菜单 treeId>0 才开弹窗，否则 alertError「主菜单不能选择」；提交 serialize POST prefix+/edit |
| 11 | 菜单树选择弹窗 | view/system/menu/tree.html | menu/tree.html | zTree | hidden treeId/treeName 初始=模板变量 menu；`url = ctx + "system/menu/menuTreeData"`（GET 裸数组）；expandLevel=1；onClick 回填 treeId/treeName；搜索/展开/折叠纯前端 |
| 12 | 图标选择片段 | view/system/menu/icon.html | menu/icon.html | 纯 HTML 片段 | .ico-list 内 Font Awesome `<i class="fa fa-xxx">` 全量列表（27KB，被 add/edit 的 icon-drop include 消费；点击取 class 填 #icon）。经典版是**完整 HTML 文档**被 th:include 整页嵌入（含 head/jquery script，浏览器容错渲染）——TP 版做纯片段：font-awesome.css 由父页 header 提供、jq 依赖去掉，视觉一致（加固见「拟登记 deviations」候选 3） |

> 术语说明：角色列表页用 bootstrap-table 分页表；菜单列表页用 bootstrap-tree-table 树表；zTree 用于角色新增/修改的菜单权限树、数据权限部门树、菜单树选择弹窗三处。「分配用户」实况是 openTab 页签（页面 5）+ 「添加用户」弹窗（页面 6），**不是双 tab 同屏**——按实况复刻。

## 端点级 API 清单（role 23 + menu 13）

### 角色 /system/role（对位 SysRoleController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/role | system:role:view | 无 | — | 渲染页面 1 |
| 2 | POST | /system/role/list | system:role:list | 无 | pageNum/pageSize/orderByColumn/isAsc（白名单 role_name/role_key/role_sort/create_time）+ roleName(like)/roleKey(like)/status/dataScope/params[beginTime]/params[endTime] | **TableDataInfo** {code:0, msg:"查询成功", rows, total}；SQL 对位 selectRoleContactVo：**distinct + sys_role r left join sys_user_role ur left join sys_user u left join sys_dept d**（DataScope 作用于 d.dept_id）；where r.del_flag='0'；**行字段 9 列驼峰**：roleId/roleName/roleKey/roleSort/dataScope/status/delFlag/createTime/remark（selectRoleContactVo 无 create_by/update_by/update_time） |
| 3 | POST | /system/role/export | system:role:export | 角色管理, 5导出 | 同搜索参数；**无分页（导全量）** | {code:0, msg:"&lt;uuid&gt;_角色数据.xlsx"}；Excel 列定义见下节 |
| 4 | GET | /system/role/add | system:role:add | 无 | — | 渲染页面 2（无模板变量） |
| 5 | POST | /system/role/add | system:role:add | 角色管理, 1新增 | roleName*、roleKey*、roleSort*、status、remark、menuIds（逗号串，可空） | 校验顺序：名称唯一 →「新增角色'{roleName}'失败，角色名称已存在」(500)；权限字符唯一 →「新增角色'{roleName}'失败，角色权限已存在」(500)；insertRole（**事务**：insert sys_role + batch insert sys_role_menu）；createBy=登录名；toAjax；**menuIds 空时不写 role_menu 也算成功**（rows=1） |
| 6 | GET | /system/role/edit/{roleId} | system:role:edit | 无 | 路径 roleId；**先 checkRoleDataScope** | 渲染页面 3，变量 role = selectRoleById（12 列全字段，where del_flag='0'） |
| 7 | POST | /system/role/edit | system:role:edit | 角色管理, 2修改 | roleId、roleName*、roleKey*、roleSort*、status、remark、menuIds | 校验顺序：**checkRoleAllowed（roleId=1 → 「不允许操作超级管理员角色」500）→ checkRoleDataScope → 双唯一**（文案前缀「修改角色…」）；updateRole（**事务**：update sys_role + **delete role_menu by roleId + 重插**）；updateBy；toAjax |
| 8 | GET | /system/role/authDataScope/{roleId} | **无（仅登录态）** | 无 | 路径 roleId；checkRoleDataScope | 渲染页面 4，变量 role——经典版此 GET **无 @RequiresPermissions** 实锤（POST 才有 edit） |
| 9 | POST | /system/role/authDataScope | system:role:edit | 角色管理, 2修改 | roleId、roleName、roleKey、dataScope、deptIds（逗号串，可空） | checkRoleAllowed → checkRoleDataScope → authDataScope（**事务**：update sys_role（data_scope 生效）+ **delete sys_role_dept by roleId + 重插**）；updateBy；**成功（>0）后刷新当前登录会话的用户数据**（对位经典版 setSysUser(selectUserById(getUserId()))——TP 版重算 LoginService.rolesOf 写回会话，使本人数据权限立即生效）→ success()「操作成功」；0 行 → error()「操作失败」 |
| 10 | POST | /system/role/remove | system:role:remove | 角色管理, 3删除 | ids（逗号串） | deleteRoleByIds：**先逐个** checkRoleAllowed + checkRoleDataScope + countUserRoleByRoleId>0 → **error(500)**「{roleName}已分配,不能删除」（ServiceException，整批拒绝）；然后（事务）delete sys_role_menu in + delete sys_role_dept in + **软删** sys_role del_flag='2'；toAjax |
| 11 | POST | /system/role/checkRoleNameUnique | **无（仅登录态）** | 无 | roleName（+编辑时 roleId） | **裸 boolean**：true=唯一 / false=重复；口径 = role_name **全局**唯一（del_flag='0'）limit 1；roleId 相同视为自身放行 |
| 12 | POST | /system/role/checkRoleKeyUnique | **无（仅登录态）** | 无 | roleKey（+roleId） | 裸 boolean；role_key 全局唯一（del_flag='0'） |
| 13 | GET | /system/role/selectMenuTree | **无** | 无 | — | **经典版死端点**：渲染 system/role/tree，但该模板不存在、全站零引用（grep 实锤）。TP 版**不复刻**（路由不注册，404）——处置见拟登记 deviations 候选 2 |
| 14 | POST | /system/role/changeStatus | system:role:edit | 角色管理, 2修改 | roleId、status（0/1） | checkRoleAllowed → checkRoleDataScope → updateRole（动态 set 仅 status 列生效 + update_time=sysdate()）；toAjax。注意**经典版此处不清授权缓存**——停用角色后已登录用户权限仍有效，TP 版会话 permissions 不变，行为一致 |
| 15 | GET | /system/role/authUser/{roleId} | system:role:edit | 无 | 路径 roleId；checkRoleDataScope | 渲染页面 5，变量 role |
| 16 | POST | /system/role/authUser/allocatedList | system:role:list | 无 | 分页参数 + roleId*、loginName(like)、phonenumber(like)；sortName=createTime desc；白名单增 login_name | TableDataInfo；SQL 对位 selectAllocatedList：distinct u 7 表字段 + **join sys_user_role/sys_role where r.role_id={roleId}**；**@DataScope(deptAlias="d", userAlias="u")**；行驼峰：userId/deptId/loginName/userName/userType/email/avatar/phonenumber/status/createTime |
| 17 | POST | /system/role/authUser/cancel | system:role:edit | 角色管理, 4授权 | roleId、userId | delete sys_user_role where user_id=? and role_id=?（**物理删单行**）；toAjax |
| 18 | POST | /system/role/authUser/cancelAll | system:role:edit | 角色管理, 4授权 | roleId、userIds（逗号串） | delete sys_user_role where role_id=? and user_id in（物理删）；toAjax |
| 19 | GET | /system/role/authUser/selectUser/{roleId} | system:role:list | 无 | 路径 roleId | 渲染页面 6，变量 role。**路由注册顺序坑：此条须先于 #15 的 authUser/:roleId 注册**（TP 按序匹配，防 :roleId 吞掉固定段——1.0.0 job clean 同款坑） |
| 20 | POST | /system/role/authUser/unallocatedList | system:role:list | 无 | 同 #16 | TableDataInfo；SQL 对位 selectUnallocatedList：`where u.del_flag='0' and (r.role_id != {roleId} or r.role_id IS NULL) and u.user_id not in (select … ur.role_id={roleId})` + @DataScope(d,u)；行驼峰同 #16 |
| 21 | POST | /system/role/authUser/selectAll | system:role:edit | 角色管理, 4授权 | roleId、userIds | checkRoleDataScope → batch insert sys_user_role（**经典版无已存在校验，重复选择会产生重复行**——原样复刻，页面侧 unallocatedList 已排除已分配用户，正常流程不触发）；toAjax |
| 22 | GET | /system/role/deptTreeData | system:role:edit | 无 | roleId（query） | **Ztree 裸数组** [{id,pId,name,title,checked,open,nocheck}]；数据源 = selectDeptList（**带 DataScope**，status='0' 过滤）+ roleDeptList = concat(dept_id, dept_name)（sys_role_dept join sys_dept del_flag='0'）；**checked = in_array(dept_id . dept_name, roleDeptList)**（整串 equals 对位）；无 roleId 时全不勾（view 页等场景） |
| 23 | GET | /system/role/view/{roleId} | system:role:list | 无 | 路径 roleId；checkRoleDataScope | 渲染页面 7；变量：role、menuTree = roleMenuTreeData(role, userId)（见菜单 #22）、**deptTree 仅 dataScope='2' 时传**（同 #22 数据源）、userCount = count(sys_user_role where role_id) |

### 菜单 /system/menu（对位 SysMenuController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 24 | GET | /system/menu | system:menu:view | 无 | — | 渲染页面 8 |
| 25 | POST | /system/menu/list | system:menu:list | 无 | menuName(like)、visible；**无分页参数**；**非 admin 按 userId 过滤**（join user_role/role_menu/role where ur.user_id=会话 userId——selectMenuListByUserId 实锤，非 DataScope 是角色关联过滤） | **裸 JSON 数组**（List&lt;SysMenu&gt; 直出，bootstrap-tree-table 特例禁套信封）；固定 order by parent_id, order_num；行字段 **13 列驼峰**（对位 selectMenuVo）：menuId/menuName/parentId/orderNum/url/target/menuType/visible/isRefresh/perms(ifnull '' 字符串)/icon/createBy/createTime |
| 26 | POST | /system/menu/remove/{menuId} | system:menu:remove | 菜单管理, 3删除 | 路径 menuId | 校验顺序：子菜单 count>0 → **warn(301)**「存在子菜单,不允许删除」；sys_role_menu count>0 → **warn(301)**「菜单已分配,不允许删除」；**物理删除 `delete from sys_menu where menu_id=? or parent_id=?`**（经典版原样 quirk：连带删一级子菜单——正常流程被前置检查拦截，仅防脏数据兜底）；toAjax |
| 27 | GET | /system/menu/add/{parentId} | system:menu:add | 无 | 路径 parentId（工具栏 0 / 行内 menuId） | 渲染页面 9；parentId≠0 → menu=selectMenuById；**parentId=0 → menu={menuId:0, menuName:"主目录"}**（经典版构造对象原样） |
| 28 | POST | /system/menu/add | system:menu:add | 菜单管理, 1新增 | parentId、menuType*、menuName*、orderNum*、url、target、perms、visible、isRefresh、icon；**serialize 整表单** | checkMenuNameUnique →「新增菜单'{menuName}'失败，菜单名称已存在」(500)；insertMenu（createBy）；toAjax |
| 29 | GET | /system/menu/edit/{menuId} | system:menu:edit | 无 | 路径 menuId | 渲染页面 10，变量 menu = selectMenuById（**含 parent_name 子查询**） |
| 30 | POST | /system/menu/edit | system:menu:edit | 菜单管理, 2修改 | menuId + 同 #28 | checkMenuNameUnique →「修改菜单'{menuName}'失败，菜单名称已存在」(500)；updateMenu（动态 set：menuName/parentId/orderNum/url/target/menuType/visible/isRefresh/perms/icon/updateBy + update_time=sysdate()）；toAjax |
| 31 | POST | /system/menu/updateSort | system:menu:edit | 保存菜单排序, 2修改 | menuIds、orderNums（前端 join(",") 同序两串；TP explode 等价 Spring String[]） | 逐行 update order_num（**事务**）；异常 → error(500)「保存排序异常，请联系管理员」；成功固定 success()「操作成功」（非 toAjax，0 行也算成功）——与 dept/updateSort 同款 |
| 32 | GET | /system/menu/icon | **无（仅登录态）** | 无 | — | 渲染页面 12（被 add/edit include 消费，非独立访问） |
| 33 | POST | /system/menu/checkMenuNameUnique | **无（仅登录态）** | 无 | menuName、parentId、（编辑时 menuId） | 裸 boolean；口径 = **同 parentId 下** menu_name 唯一（sys_menu 无 del_flag，全量物理口径）limit 1；menuId 相同视为自身放行 |
| 34 | GET | /system/menu/roleMenuTreeData | **无（仅登录态）** | 无 | roleId（query，可空）；**非 admin 菜单范围按角色关联过滤**（selectMenuAll 内部分支） | Ztree 裸数组：[{id,pId,name,title,checked,open,nocheck}]；**name = menuName + '&lt;font color="#888"&gt;&amp;nbsp;&amp;nbsp;&amp;nbsp;' + perms + '&lt;/font&gt;'**（permsFlag 恒 true；ztree nameIsHTML 渲染灰字）；**checked = in_array(menu_id . perms, roleMenuList)**，roleMenuList = select concat(menu_id, ifnull(perms,'')) where rm.role_id（整串 equals 对位，非子串包含）；roleId 空 → 全不勾 |
| 35 | GET | /system/menu/menuTreeData | **无（仅登录态）** | 无 | — | Ztree 裸数组；同 #34 数据范围但 **permsFlag=false（name 不带 perms 后缀）、无 checked 逻辑**（全 false） |
| 36 | GET | /system/menu/selectMenuTree/{menuId} | **无（仅登录态）** | 无 | 路径 menuId（add 页传 treeId>0?treeId:1，edit 页传 treeId） | 渲染页面 11，变量 menu = selectMenuById |

统计：**控制器方法 36 个**（role 23 + menu 13），**URL pattern 36 条**（role/selectMenuTree 死端点不注册后实落 35 条）；**#[Perm] 共 27 处**（role 19 + menu 8；无注解仅登录态 9 处：role 的 authDataScope GET/checkRoleNameUnique/checkRoleKeyUnique/selectMenuTree + menu 的 icon/checkMenuNameUnique/roleMenuTreeData/menuTreeData/selectMenuTree——grep 实锤）；**#[Log] 共 13 处**（role 9：EXPORT/INSERT/UPDATE×3/DELETE/GRANT×3；menu 4：DELETE/INSERT/UPDATE×2）；**RepeatSubmit 零处**（经典版 SysRoleController/SysMenuController 无 @RepeatSubmit 注解 grep 实锤 → 本模块不挂）。

## 特殊行为清单（唯一性文案 / admin 保护 / 级联 / 关联 / 前端 quirk）

1. **唯一性校验文案全集**（前后端双保险，文案不同属经典版原样）：
   - 后端 add/edit：`新增角色'{roleName}'失败，角色名称已存在` / `修改角色…同款`；`新增角色'{roleName}'失败，角色权限已存在` / 修改同款；`新增菜单'{menuName}'失败，菜单名称已存在` / `修改菜单…同款`
   - 前端 remote 提示：角色「角色名称已经存在」「角色权限已经存在」；菜单「菜单已经存在」
   - 口径差异：角色名称/权限字符 = **全局**唯一（del_flag='0'）；菜单名称 = **同 parentId 下**唯一（sys_menu 无 del_flag）。均 limit 1、自身放行（roleId/menuId 相等）。
2. **admin 保护（仅角色）**：① checkRoleAllowed：roleId=1 →「不允许操作超级管理员角色」(500)，作用于 editSave/authDataScopeSave/changeStatus/remove（remove 逐个查）；② checkRoleDataScope：非 admin 对数据范围外 roleId →「没有权限访问角色数据！」(500)，作用于 edit GET/editSave/authDataScope GET+POST/remove/changeStatus/authUser GET/selectAll/view 共 9 处——实现 = 按 roleId 走带 DataScope 的 selectRoleList 查空即拒（**角色数据权限的实义 = 「角色的关联用户所在部门」在我的数据范围内**，因 selectRoleContactVo join 到 d）；③ **前端 roleId==1 行操作列空**。菜单模块**无任何 admin 保护**（经典版同样没有）。
3. **role_menu 关联维护**：新增/修改角色时事务内 delete by roleId + batch 重插（修改即全量替换）；menuIds 逗号串 explode，空数组跳过写入。**role_dept 关联维护**（authDataScope）：同款 delete + 重插；deptIds 空数组跳过。
4. **数据权限五种 data_scope 的落库**：dataScope 页面 select 1~5；**仅 =2（自定义）时** deptIds 才有勾选意义（页面切非 2 清空勾选）；sys_role_dept 落 deptIds（复合主键 role_id+dept_id）；sys_role.data_scope 恒更新。保存成功后**刷新当前会话**（对位 setSysUser）——本人立即生效；其他会话的数据权限过滤条件来自会话 roles[].data_scope，**重新登录生效**（见 deviations 候选 1）。
5. **分配用户语义**：allocatedList/unallocatedList 双列表（已/未分配，未分配 SQL 见 #20）；cancel/cancelAll 物理删 sys_user_role；selectAll 批量插入**无去重**（经典版原样，页面流程保证不重复）；sys_user_role 在本模块**只被写入/删除，不由本模块建**（预置 2 行：admin→1、ry→2）。
6. **菜单类型 M/C/F 联动**（页面侧）：新增页 radio 无默认（必选校验）；M=目录（无 url/perms/target/刷新，有 icon）/C=菜单（全字段）/F=按钮（仅 perms，无 url/icon/target/刷新，可见列显示 '-'）；修改页加载即按当前类型联动显隐。**DB 无联动约束**（service 不因类型裁字段，update 动态 set 原样写入）。
7. **菜单删除语义**：物理删；前置双 warn(301)（有子菜单 / 已分配角色）；`or parent_id=?` 连带删一级子菜单为经典版原样兜底（正常流程不触发，**照抄勿"修复"**）。
8. **排序**：role 列表由 orderByColumn 驱动（默认 roleSort asc）；menu 列表固定 order by parent_id, order_num（无前端排序列），另有 /updateSort 批量排序端点（仅提交改动行，与 dept 同款）。
9. **后端参数校验文案**（对位 @Validated + BindException → error 首条消息）：角色「角色名称不能为空」「角色名称长度不能超过30个字符」（DB varchar(30) 同口径）「权限字符不能为空」「权限字符长度不能超过100个字符」「显示顺序不能为空」；菜单「菜单名称不能为空」「菜单名称长度不能超过50个字符」「显示顺序不能为空」「请求地址不能超过200个字符」「菜单类型不能为空」「权限标识长度不能超过100个字符」。
10. **防重复提交**：两控制器 @RepeatSubmit 零处（grep 实锤）→ 不挂 RepeatSubmit 中间件。
11. **前端 quirk（原样保留，勿修）**：① role 列表 dataScope 徽章 formatter 硬编码五态（非字典渲染）；② 分配用户页按钮权限串用 add/remove 而非 edit（经典版原样，照抄）；③ 菜单 add 页类型 radio 无默认选中、页面初始所有联动区块按 M 态隐藏；④ menu edit 选上级「主菜单不能选择」；⑤ role/view 用户卡片前端过滤渲染、pageSize=100 分页加载更多。
12. **PHP 特有坑（前序踩坑防复发）**：① **列表 API 输出键名必须驼峰**（bootstrap-table/tree-table columns field 按驼峰取值）——role/list 9 字段、allocated/unallocated 10 字段、menu/list 13 字段、Ztree {id,pId,name,title,checked,open,nocheck} **pId 键名精确**；② **视图 layer**：`system.RoleController` / `system.MenuController` → 视图根自动变 `app/view/system/`，渲染路径用 layer 相对（'role/index'、'menu/index'），include 片段（header/footer/ztree/icon）用 layer 内副本 `app/view/system/include/`（3.0.0 已建，缺的补复制）；③ **全局类在 namespaced 控制器必须 use**（AjaxResult/TableDataInfo/TpConstant/RedisCache/各 Service）；④ **`empty('0') === true` 陷阱**：status/visible/dataScope 等字符 '0' 值判空一律用 `!== ''` / `!== null`，禁用 empty()（status='0' 启用态最易踩）；⑤ 路由遮蔽：`authUser/selectUser/:roleId` 必须注册在 `authUser/:roleId` 之前。

## 数据权限声明

- **selectRoleList 带 @DataScope(deptAlias="d")**（实锤）：角色列表 / export / checkRoleDataScope 内部查询三处消费；别名 d 指向 join 的 sys_dept（selectRoleContactVo 三 left join 后 DataScope::apply($query, $user, 'd')，条件落 d.dept_id）——**对位 3.0.0 DataScope 封装原样复用，无需扩展**。
- **selectAllocatedList / selectUnallocatedList 带 @DataScope(deptAlias="d", userAlias="u")**（SysUserServiceImpl 实锤）：分配用户双列表；DataScope::apply($query, $user, 'd', 'u')（封装已支持 userAlias 参数，1.0.0 落）。
- **roleDeptTreeData 内部走 selectDeptList**（@DataScope(d)）：deptTreeData 端点与 view 页 deptTree 均带数据权限。
- **菜单三处按 userId/角色关联过滤（非 DataScope 注解）**：menu/list 非 admin 走 selectMenuListByUserId；selectMenuAll（roleMenuTreeData/menuTreeData 数据源）非 admin 走 selectMenuAllByUserId——TP 版用 PermissionService::isAdmin 分支 + join sys_role_menu/sys_user_role/sys_role（参考 MenuService::menusOf 既有非 admin 分支写法，但**不过滤 menu_type/visible**——selectMenuAll 系含 F 型与隐藏菜单）。
- checkRoleNameUnique / checkRoleKeyUnique / checkMenuNameUnique / countUserRoleByRoleId / selectRoleById / selectMenuById **不带数据权限**（经典版一致）；越权入口由 checkRoleDataScope 把守。
- admin（user_id=1）全部不过滤。

## Excel 导出列定义（SysRole @Excel 逐字段抄录，共 6 列）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | roleId | name="角色序号", cellType=NUMERIC | 列头「角色序号」，数字格式 |
| 2 | roleName | name="角色名称" | 列头「角色名称」，文本 |
| 3 | roleKey | name="角色权限" | 列头「角色权限」，文本 |
| 4 | roleSort | name="角色排序", cellType=NUMERIC | 列头「角色排序」，数字格式（String 字段，输出转数值） |
| 5 | dataScope | name="数据范围", readConverterExp="1=所有数据权限,2=自定义数据权限,3=本部门数据权限,4=本部门及以下数据权限,5=仅本人数据权限" | 列头「数据范围」，按表达式转换后输出文本 |
| 6 | status | name="角色状态", readConverterExp="0=正常,1=停用" | 列头「角色状态」，转换输出 |

- BaseEntity 五字段**无 @Excel → 不导出**；delFlag/flag/menuIds/deptIds/permissions 无 @Excel；sheet 名「角色数据」；文件名 `&lt;uuid&gt;_角色数据.xlsx`（复用 3.0.0 ExcelExportService，deviations #19 样式从简已定）。
- **菜单无导出**：SysMenu 无 @Excel 注解、SysMenuController 无 export 端点（动工检查单第 4 条核对结论）。

## 关键设计说明

1. **两个裸数组特例**：role/deptTreeData、menu 全部树数据返回 `json($rows)` 裸数组；menu/list 同 dept/list 裸数组（tree-table responseHandler 对无 code 放行）；role/list、allocatedList、unallocatedList 三处才套 TableDataInfo。check 四端点（#11/#12/#33）返回裸 boolean。
2. **MenuService 扩展边界**：2.0.0 的 menusOf（导航树，M/C+visible=0）**不动**；新增 selectMenuAll($user)（admin=selectMenuVo 全量 13 列；非 admin=join 三表 distinct 同 13 列，**不筛 menu_type/visible**）+ roleMenuTreeData(roleId, $user) / menuTreeData($user) 组装 Ztree + checkMenuNameUnique / insertMenu / updateMenu / updateMenuSort / deleteMenuById / selectMenuById（含 parent_name 子查询） / selectCountMenuByParentId / selectCountRoleMenuByMenuId。树构建直接平铺 Ztree（父子关系由 pId 表达，无需 buildTree）。
3. **ztree checked 判定用整串相等**：`in_array($menuId . $perms, $roleMenuList, true)` / `in_array($deptId . $deptName, $roleDeptList, true)`——对位 Java List.contains(equals)，**不是子串包含**；roleMenuList/roleDeptList 用 select concat 查询一次拉全。
4. **会话刷新点**：authDataScopeSave 成功后重算 LoginService.rolesOf(userId) 并 SessionService::write 写回**当前会话**（对位 setSysUser）——需要会话 uuid，从中间件传递（实现时核对 LoginAuth 是否把 uuid 放进 request，无则从 cookie 读）。add/edit 角色菜单权限**不刷会话**（见 deviations 候选 1）。
5. **事务边界**：insertRole / updateRole / authDataScope / deleteRoleByIds（删关联+软删） / updateMenuSort 必须包显式事务（对位 @Transactional）。
6. **权限两通道接线**：按钮 `shiro:hasPermission` → `{if check_perm('system:role:add')}`；页面 JS 变量 `var editFlag = '[[${@permission.hasPermi(...)}]]'` → `{:check_perm('system:role:edit') ? '' : 'hidden'}`；字典数组 `var datas = {:json(DictService::listByType('sys_normal_disable'))}`（authUser/selectUser 徽章）与 `sys_show_hide`（menu 列表可见列 + add/edit radio volist，双键并存同 3.0.0）。
7. **view 页数据注入**：menuNodes = `{:json($menuTree)}`；deptTree 仅 dataScope=='2' 时 assign（模板侧 JS 默认 '[]'）；userCount 服务端渲染徽章；用户卡片由页面 JS 分页拉 allocatedList（同会话同权限，无需新端点）。
8. **getSysUser/isAdmin**：会话内容（2.0.0 已含 userId/loginName/deptId/permissions/roles/isAdmin，roles[] 含 role_id/role_key/role_name/data_scope）直接读取，不新增会话字段。

## 拟登记 deviations

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| 1 | 修改角色菜单权限/菜单增删改后 `AuthorizationUtils.clearAllCachedAuthorizationInfo()` 清 Shiro 授权缓存 → **全部已登录用户权限立即生效** | TP 版 permissions 登录时算好存 Redis 会话，无全局清缓存机制 → 已登录会话按钮显隐与 CheckPerm 判定**重新登录生效**（数据权限因 authDataScopeSave 刷本人会话不受影响） | 实施时定：A=接受「重新登录生效」登记本条（推荐，成本零收益实）；B=SCAN `tp_session:*` 重算各会话 permissions（对位全局清缓存，SCAN 成本与并发写回风险需评估）。**动工时拍板后登记 deviations.md 并回填 checklist** |
| 2 | GET /system/role/selectMenuTree 渲染 system/role/tree——**模板不存在、全站零引用的死端点**（运行时点开必报错） | 不复刻，路由不注册（404） | 死端点无行为可复刻；登记一条「role/selectMenuTree 死端点不复刻」或作为调研注记保留于本 spec 均可，**审核时拍板** |
| 3 | menu/icon.html 是完整 HTML 文档被 th:include 整页嵌入 add/edit（head/link/jquery script 全部进 div，浏览器容错渲染） | TP 版 icon.html 做纯 `.ico-list` 片段，font-awesome 由父页 header 提供，不重复引 jquery | 视觉一致、结构加固，倾向**不登记**（作为实现加固注明本 spec 页面 12）；若审核认为属行为差异再登记 |
| 4 | sys_user_role 无唯一约束，selectAll 重复授权产生重复行 | 原样复刻（无去重） | 经典版原样行为照抄，**不登记**（防误改说明，见特殊行为 5） |

## 任务分解 → tasks.md

## 验收清单 → checklist.md


## 实施记录

- 2026-09-29 全部 Task 完成（PHPUnit 40 tests；curl 22 项 + 浏览器级 + **3.0.0 遗留的权限两通道隔离验证全过**）。
- **权限隔离验证实测**（isouser 挂隔离角色：仅 dept view/list + data_scope=4）：①无权端点 CheckPerm 拦截（「您没有操作权限…【system:dept:add】」精确输出缺的权限串）；②dept/list 数据权限过滤（仅见本部门及以下 1 行 vs admin 10 行）；③admin 全量不受影响。
- **联调修复的坑**（对后续模块关键）：
  1. **CheckPerm/OperLog 挂载方式**：路由 group ->middleware() 在全局管线执行（dispatch 前 rule 未就位）；正确挂法 = config/route.php 的 'middleware' 项（route 管线，dispatch->init 之后）。
  2. **request->controller() 已含 layer 前缀**（'system.Dept'）——resolvePerm 拼类名不能再拼 layer（曾拼出 system\System\DeptController 不存在 → perm=null 放行）。
  3. **DataScope 会话键双兼容**：会话 deptId（驼峰）vs 直查行 dept_id（下划线）——apply 内 (deptId ?? dept_id) 双取。
  4. **DataScope 闭包条件**：where(...) 数组展开在闭包内生成 0=1；改显式 whereIn/where。
  5. **DataScope 闭包缺 use**（ 未传入内层闭包）——admin 场景永不触发，非 admin 首次即炸。
  6. **unallocatedList 路由遮蔽**：POST authUser/unallocatedList 被先注册的 GET authUser/:roleId 吞——全部固定段提到通配之前。
  7. **authUser/selectUser 页缺 $datas assign**（字典徽章）——控制器补 DictService。
- deviations 拍板：候选 1 = **A 方案登记**（「重新登录生效」，deviations #21 已含）；候选 2 = **role/selectMenuTree 死端点不复刻**（本表）；候选 3 = icon 纯片段（加固，不登记）。
- 预置数据复原核对：sys_role 2 / sys_menu 85 / sys_user_role 2 / sys_user 2。
