# 5.0.0-角色菜单 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（role 22 存活 + menu 13；死端点 selectMenuTree 不复刻）。

## 角色管理（curl / DB 级）

- [ ] GET /system/role 渲染列表页（system:role:view）；未登录 302 /login、ajax code "1"
- [ ] POST /system/role/list：TableDataInfo {code:0, rows, total}；行 9 列驼峰；分页/排序（roleSort 默认 asc，白名单 role_name/role_key/role_sort/create_time）；roleName/roleKey like、status/dataScope 精确、beginTime/endTime date_format 口径
- [ ] 角色列表带数据权限：非 admin 按 join 的 d.dept_id 过滤（=「其关联用户所在部门在我的范围」）——**权限隔离验证 Task 5 一并实测**
- [ ] POST /system/role/export：{code:0, msg:"&lt;uuid&gt;_角色数据.xlsx"}；导全量（带搜索参数）；xlsx 读回 6 列断言（序号/名称/权限/排序数字格式/数据范围五态转换/状态转换）；GET /common/download 链路复用
- [ ] POST /system/role/add：名称重名 → 500「新增角色'xx'失败，角色名称已存在」；权限字符重名 → 500「…角色权限已存在」；成功事务落库（sys_role + sys_role_menu batch；create_by/admin）；menuIds 空 = 成功不写关联
- [ ] GET /system/role/edit/{roleId}：checkRoleDataScope 前置；变量 role 12 列
- [ ] POST /system/role/edit：roleId=1 → 500「不允许操作超级管理员角色」实测；双唯一（修改前缀）；sys_role_menu delete+重插 DB 核对；update_by 落库
- [ ] GET /system/role/authDataScope/{roleId} 无 #[Perm]（仅登录态可渲染，与 POST 端不对称——经典版原样）
- [ ] POST /system/role/authDataScope：五种 data_scope 逐一落库核对；=2 时 deptIds 写 sys_role_dept（delete+重插）；=1/3/4/5 时旧 role_dept 行仍被清空后不重插；成功刷本人会话（本人数据权限立即生效实测）；0 行 → 「操作失败」
- [ ] POST /system/role/remove：已分配（ry→role2）→ 500「普通角色已分配,不能删除」整批拒绝；未分配测试角色 → 软删 del_flag='2' + role_menu/role_dept 关联行消失；ids 含 roleId=1 → 500 拒绝
- [ ] POST /system/role/checkRoleNameUnique / checkRoleKeyUnique：裸 boolean；全局唯一口径；自身放行
- [ ] POST /system/role/changeStatus：停用/启用 status 落库 + update_time；roleId=1 拒绝
- [ ] GET /system/role/authUser/{roleId} 渲染分配用户页（checkRoleDataScope 前置）
- [ ] POST /system/role/authUser/allocatedList：roleId 过滤 + loginName/phonenumber like + @DataScope(d,u) + 分页排序（createTime desc）
- [ ] POST /system/role/authUser/unallocatedList：SQL 口径核对（`r.role_id != ? or IS NULL` + not in 子查询）
- [ ] POST /system/role/authUser/cancel / cancelAll：sys_user_role 物理删单行/批量
- [ ] GET /system/role/authUser/selectUser/{roleId} 路由不被 authUser/:roleId 吞（先注册）
- [ ] POST /system/role/authUser/selectAll：batch 插入（重复 id 产生重复行=经典版原样，仅记录不视为 bug）
- [ ] GET /system/role/deptTreeData：Ztree 裸数组；checked = dept_id.dept_name 整串相等（role 2 预置 3 行实测勾选）；非 admin 带 DataScope；无 roleId 全不勾
- [ ] GET /system/role/view/{roleId}：menuTree（勾选态 Ztree）/ deptTree（仅 =2）/ userCount 注入；数据范围外 → 500「没有权限访问角色数据！」

## 菜单管理（curl / DB 级）

- [ ] GET /system/menu 渲染列表页（system:menu:view）
- [ ] POST /system/menu/list：**裸数组** 13 列驼峰；order by parent_id, order_num；admin 全量；非 admin 按角色关联过滤（selectMenuListByUserId 原样）
- [ ] POST /system/menu/remove/{menuId}：有子菜单 → **301**「存在子菜单,不允许删除」；已分配角色 → **301**「菜单已分配,不允许删除」；叶子未分配 → 物理删成功
- [ ] GET /system/menu/add/{parentId}：parentId=0 → menu={menuId:0, menuName:"主目录"}
- [ ] POST /system/menu/add：同父同名 → 500「新增菜单'xx'失败，菜单名称已存在」；异父同名可建；落库 create_by
- [ ] GET/POST /system/menu/edit：parent_name 子查询回显；同款唯一校验（修改前缀）；动态 set 更新
- [ ] POST /system/menu/updateSort：逗号串逐行生效；异常 →「保存排序异常，请联系管理员」；成功恒 code 0
- [ ] POST /system/menu/checkMenuNameUnique：裸 boolean；同 parent_id 口径
- [ ] GET /system/menu/roleMenuTreeData：name 带 perms 灰字 font 标签；checked 整串相等（role 2 预置全量 85 行实测全勾）；roleId 空/新增页全不勾
- [ ] GET /system/menu/menuTreeData：name 无 perms 后缀、checked 恒 false
- [ ] GET /system/menu/selectMenuTree/{menuId} 渲染树选择页
- [ ] 菜单模块无 admin 保护、无数据权限（负向断言：CheckPerm 不拦 admin 外的 menu 端点合法 perm；无 DataScope 调用）

## 权限两通道隔离验证（3.0.0 遗留，本模块补验）

- [ ] 建测试角色（仅勾选 部门+岗位 部分菜单/F 按钮）+ 测试用户挂接；登录态下主框架菜单仅显被授权项（MenuService 通道）
- [ ] 无权按钮不渲染（check_perm 通道：模板 {if check_perm} 与 JS flag 两处）
- [ ] 直接 POST 无权端点 → CheckPerm 中间件拦截 500（含未授权 check 端点外的全部 #[Perm] 端点）
- [ ] 数据权限过滤实测：测试角色 data_scope=2/3/4 各验一（部门列表行数、角色列表行数、allocated 列表）
- [ ] 越权入口实测：非 admin 访问 /system/role/edit/1、/system/role/view/1 → 500「没有权限访问角色数据！」；操作 role_id=1 的 edit/changeStatus → 500「不允许操作超级管理员角色」
- [ ] 授权缓存语义实测并拍板：改测试角色菜单后已登录会话权限不变、重新登录生效 → deviations 候选 1 处置回填

## 页面级验收（浏览器）

- [ ] 角色列表页：分页/列排序点击、dataScope 徽章五态、状态开关图标切换、roleName 链接开详情、roleId=1 行操作列空、更多 popover（数据权限/分配用户）
- [ ] 角色新增/修改弹窗：菜单 zTree 勾选与父子联动 checkbox 三态、remote 即时提示、提交后父页刷新
- [ ] 数据权限弹窗：readonly 两字段、select 切换树显隐（切非 2 清勾选）、=2 默认展开勾选回显
- [ ] 分配用户页签（openTab）：双列表切换、添加用户弹窗选择授权、批量/行内取消授权、关闭页签
- [ ] view 详情页：菜单权限计数、勾选树展开渲染、自定义部门树（=2）、关联用户卡片懒加载/搜索/加载更多
- [ ] 菜单列表页：树表层级渲染（85 行预置）、图标+名称列、可见列 F 型 '-'、类型徽章、保存排序（改动行才提交、无改动提示「未检测到排序修改」）、展开/折叠
- [ ] 菜单新增/修改弹窗：类型 radio 联动显隐（M/C/F）、图标选择器点击回填、上级菜单树弹窗回填、edit 页「主菜单不能选择」、parentName「无」
- [ ] 全部页面在主框架 iframe 内打开、标签页标题正确（modalName：角色/菜单）；主框架导航树回归正常（menusOf 未受扩展影响）

## 横切与纪律自查

- [ ] #[Perm] 27 处与 spec 表逐一对应（role 19 + menu 8）；9 处无注解端点仅登录态（负向核对无漏挂）
- [ ] #[Log] 13 处：操作成功/失败均落 sys_oper_log（title=角色管理/保存菜单排序/菜单管理；business_type 1/2/3/4/5 对应；授权 4 落 cancel/cancelAll/selectAll 三处）
- [ ] RepeatSubmit 未挂本模块路由（经典版无 @RepeatSubmit 实锤）
- [ ] 业务代码零处直接 predis（grep）；Redis 只经 RedisCache 门面
- [ ] DataScope::apply 仅四处：selectRoleList(d)、allocatedList(d,u)、unallocatedList(d,u)、roleDeptTreeData 内 selectDeptList(d)
- [ ] 权限串三处互查：sys_menu.perms 11 行（system:role:* 6 + system:menu:* 5）与 #[Perm] 注解、check_perm 调用一致
- [ ] 表结构零变更（SHOW COLUMNS 前后一致）；sys_role_menu/sys_role_dept 复合主键未触碰约束
- [ ] '0' 值判空抽查：status/data_scope/visible 相关代码无 empty()（grep 核对）

## Deviations 核对

- [ ] 候选 1（授权缓存「重新登录生效」）：Task 5 实测后拍板 A/B 并登记 deviations.md（含编号）
- [ ] 候选 2（role/selectMenuTree 死端点不复刻）：审核拍板后处置（登记或注记保留）
- [ ] 候选 3（icon 纯片段加固）：审核拍板是否登记
- [ ] 前端 quirk 原样保留未"修复"：dataScope 徽章硬编码 / 分配用户按钮权限串 add/remove / 菜单删除 or parent_id 连带删 / selectAll 无去重

## 测试数据清理记录

（动工后逐条登记；清理完成在条目前加 ✅ 并注日期）

- [ ] 测试角色（含其 sys_role_menu / sys_role_dept 关联行）删除
- [ ] 测试用户（含 sys_user_role 关联行）删除；sys_user 预置 2 行（admin/ry）复原
- [ ] sys_user_role 测试期间新增行删除（预置 2 行 admin→1、ry→2 保留）
- [ ] sys_role 预置 2 行（1 超级管理员 / 2 普通角色）字段原值复原（status/data_scope/remark 等）
- [ ] sys_role_menu 复原 85 行（role 2 全量）；sys_role_dept 复原 3 行（role 2→100/101/105）
- [ ] sys_menu 测试菜单行删除，复原 85 行原值（测试按钮/菜单物理删除）
- [ ] sys_oper_log / sys_logininfor 测试行清理
- [ ] runtime/download 导出残留删除
- [ ] Redis 测试键（测试会话清理；dict: 预热键可保留）
- [ ] admin/admin123 与 ry 用户登录状态复原
