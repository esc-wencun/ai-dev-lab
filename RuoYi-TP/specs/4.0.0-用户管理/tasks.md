# 4.0.0-用户管理 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：模型/service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。

## Task 1 · 域模型与基础增补（2026-09-29 完成）

^- [x] `app/model/SysUser.php`（think-orm 模型，表 sys_user；autoWriteTimestamp 关闭，时间由 service 显式写；纯表映射薄模型，对位 3.0.0 SysDept 风格）
^- [x] sys_user_role / sys_user_post **不建模型**（复合主键无自增，service 内 Db::table 直操作，类注释注明）
^- [x] PostService 增 selectPostAll()（全量，对位 selectPostVo）与 selectPostsByUserId(userId)（全量 + flag=true 合并用户已有岗位；输出驼峰键含 flag）
^- [x] DictService 增最小 `getLabel(string $type, string $value): string`（view 页性别 label 反查；6.0.0 收编）
- [ ] PHPUnit：flag 合并逻辑、getLabel 命中/未命中

## Task 2 · RoleService 最小只读（2026-09-29 完成）

^- [x] `app/service/RoleService.php`：selectRoleAll($user)——del_flag='0'，join user_role/user/dept 链（对位 selectRoleContactVo distinct）+ 非 admin DataScope('d')；输出驼峰（roleId/roleName/roleKey/roleSort/dataScope/status/createTime）
^- [x] selectRolesByUserId(userId, $user)：selectRoleAll 全量 + flag=true 合并（sys_user_role 查用户已有角色）
^- [x] checkRoleDataScope(array $roleIds, $user)：非 admin 逐个 roleId 用带数据权限的查询判空 → \BusinessException('没有权限访问角色数据！')；空数组直接放行（经典可变参数空数组不进循环）
^- [x] 类注释注明 5.0.0 收编边界
^- [x] PHPUnit：flag 合并、空 roleIds 放行；数据权限过滤 DB 级在端到端验证（当前库仅 admin，负向留 5.0.0）

## Task 3 · UserService 主体（2026-09-29 完成）

^- [x] selectUserList(filter, user)：u.del_flag='0' + loginName like / status / phonenumber like / beginTime-endTime（create_time 按 %Y%m%d 日粒度）/ deptId 含子树（`u.dept_id=X OR dept_id IN (SELECT … FIND_IN_SET(X, ancestors))`）+ left join sys_dept d + **DataScope::apply($query, $user, 'd', 'u')**（双别名首用；确认 TP 封装 SELF 作用域产出 u.user_id 条件语义）——**返回 Query 构造器**（分页/全量由调用方决定，3.0.0 PostService 踩坑预防）
^- [x] selectUserById（selectUserVo 全列 + dept 嵌套；join user_role/role 组装 roles 集合供会话刷新用）；selectUserByLoginName（导入判存用）
^- [x] checkLoginNameUnique / checkPhoneUnique / checkEmailUnique（全局 del_flag='0' limit 1，userId 自身放行）
^- [x] checkUserAllowed(userId)：userId===1 → \BusinessException('不允许操作超级管理员用户')
^- [x] checkUserDataScope(userId, user)：非 admin 用带权限 selectUserList(userId) 判空 → '没有权限访问用户数据！'
^- [x] selectUserRoleGroup / selectUserPostGroup：角色名/岗位名逗号拼接，空返回 ''
^- [x] insertUser（事务）：insert sys_user（salt/md5 密码/pwd_update_date/create_by 由控制器传入）+ insertUserPost + insertUserRole（roleIds/postIds 空跳过）
^- [x] updateUser（事务）：user_role 删→插 → user_post 删→插 → update sys_user（非空列语义对位 mapper `<if>`；login_name 不在 set）
^- [x] insertUserAuth（事务）：user_role 删→插
^- [x] deleteUserByIds(ids)：逐个 checkUserAllowed + checkUserDataScope 后事务内 user_role/user_post 物理 in 删 + sys_user 软删
^- [x] changeStatus / resetUserPwd（password+salt+pwd_update_date+update_time）
^- [x] importUser(rows, updateSupport, operName)：spec 特殊行为 5 全链（初始密码 md5(loginName+initPassword) 不写 salt、deptId 新增生效/更新被覆盖、EscapeUtil→htmlspecialchars、<br/> 消息拼装、失败整批抛）
^- [x] PHPUnit：ids 解析、唯一口径、checkUserAllowed、导入消息拼装（成功/失败/混合文案）、md5(loginName+initPassword) 断言；数据权限与事务 DB 级端到端验证

## Task 4 · Excel 导入解析与模板生成（2026-09-29 完成）

^- [x] ExcelExportService 增 `parse`(string $filePath, array $headerMap): array`——首行表头名→列索引映射，按名取值；readConverterExp 反向转换（男→0）；phonenumber 文本读取（FormattedValue/ValueBinder 兜底）；空行跳过；**phpspreadsheet 2.0 API（getCell 坐标数组）**
^- [x] ExcelExportService 增 `exportTemplate`(array $columns, string $sheetName): string`——只写列头空表（IMPORT 列 7 列），复用列头样式，输出 `<uuid>_用户数据.xlsx`
^- [x] SysUser 列定义两套常量（导出 11 列 / 模板 7 列，spec Excel 节逐列）；**全局类 use 纪律**（ExcelExportService 已在 namespaced 文件踩过坑）
^- [x] PHPUnit：真实 xlsx 生成模板 + 读回断言（7 列列头）；构造含反向转换值的 xlsx → parse 读回断言（性别「男」→'0'、状态「正常」→'0'、手机号文本无科学计数）；表头名乱序文件按名匹配

## Task 5 · 控制器与路由（2026-09-29 curl 级完成）

^- [x] `app/controller/system/UserController.php` 21 方法，layer 'system.UserController' → 渲染路径 'user/index' 等（layer 相对，3.0.0 实锤）
^- [x] #[Perm] 18 处 / #[Log] 8 处与 spec 表逐一对应（resetPwd title=**重置密码**、insertAuthRole businessType=4授权、importData=6导入；check 三端点无注解；**未挂 RepeatSubmit**）
^- [x] list 输出：TableDataInfo + 行驼峰 + dept 嵌套对象 + **password/salt 剔除**；PageQuery 白名单 ['login_name','create_time']
^- [x] addSave/editSave 校验顺序逐条对位（数据权限 → 三唯一 → 密码加密）；editSave 前置 checkUserAllowed；remove 前置「当前用户不能删除」；changeStatus checkUserAllowed（admin 开关后端拦）
^- [x] resetPwdSave：重置对象=当前登录用户 → SessionService 写回会话 user
^- [x] importData：multipart `$request->file('file')` + updateSupport（勾选 "on" → true）；updateSupport 入操作日志参数（经典一致）；空文件/非法扩展校验
^- [x] importTemplate：#[Perm('system:user:view')]（经典原样）+ 无 #[Log]
^- [x] 路由注册 route/app.php：user 组 21 条（`system.user/xxx` 斜杠风格；deptTreeData/selectDeptTree/:deptId 路径参数）
- [ ] curl 级全端点自测（UTF-8 脚本）：list 信封+嵌套+无 password 字段、add 全链（关联表落库 DB 核对）、edit 换角色换岗位（旧关联删新关联插）、remove（软删+关联物理删）、resetPwd（md5 验证+会话刷新仅本人场景）、changeStatus admin 拦截文案、check 三端点裸 bool、export→download、importTemplate→download、importData 成功/已存在不更新/更新分支三场景

## Task 6 · 页面模板（7 页）（2026-09-29 完成）

- [x] view/system/user/index.html：ui-layout 左 zTree（deptTreeData onClick 写隐藏域）+ 右 bootstrap-table（options 九个 URL、sortName=createTime desc、modalName=用户）；五工具栏按钮 check_perm（add/edit/remove/import/export）；状态开关列 editFlag 控制 + statusTools/changeStatus；操作列 userId!=1 条件 + 更多弹层（resetPwd modal 800x300 / authRole addTab）；dept.deptName 嵌套列；loginName 链接 view 右滑出；resetPre 清隐藏域；#importTpl 导入模板；窄屏 layout 折叠
- [x] view/system/user/add.html：字段齐全（treeId/treeName 隐藏+输入、userName/loginName/password 预填 initPassword/sex 字典/status 开关/roles checkbox 去 admin/posts select2 multiple/remark）；validate remote 三处（loginName/email/phonenumber，文案「用户已经存在」等）；checkpwd(chrtype)（sys.account.chrtype=0 不拦）；saveTab 提交 +roleIds/postIds 逗号串；selectDeptTree() treeId 空→"100"、openOptions 三按钮（确认/清除/关闭）
- [x] view/system/user/edit.html：loginName 只读无 remote；email/phonenumber remote 带 userId；roles/posts flag 回显（checked/selected、status=1 disabled）；saveTab(prefix+"/edit")
- [x] view/system/user/view.html：只读 plaintext 15 字段（含 postGroup/roleGroup 空兜底「无岗位/无角色」、sex 字典 label、loginIp/loginDate）
- [x] view/system/user/resetPwd.html：userId 隐藏 + loginName 只读 + password 预填；validate required/5-20/specialSign；`$.operate.save`（非 saveTab）
- [x] view/system/user/authRole.html：bootstrap-table client 分页 data=roles；checkbox formatter checked=flag disabled=status=='1'；maintainSelected；submitHandler 遍历 getData 收勾选 roleId join(",")
- [x] view/system/user/deptTree.html：zTree url=prefix+"/deptTreeData"；treeId/treeName 隐藏域初始值=dept；expandLevel=2
- [x] 静态自查：无 th:*/@{ 残留、check_perm 用 {:} 输出、volist/if 配平、模板变量键名与控制器 assign 一致、include 用 system/include/ 局部副本（select2/ztree/layout）

## Task 7 · 端到端验收（浏览器级）（2026-09-29 完成）

- [x] 用户列表页：主框架「用户管理」标签 + iframe；左树 10 部门 zTree 渲染 + 点击过滤右侧表格（研发部门下 2 人）；重置还原；展开/折叠/刷新树
- [x] 列表行为：2 行预置数据（admin/ry）渲染；loginName 链接右滑详情；状态开关（ry 行）切换生效且 DB 核对；admin 行操作列空、开关点击被后端拦（alertError 文案）
- [x] 新增用户：标签页打开；部门树选择回填；角色/岗位勾选；提交 → sys_user（salt 6 位 hex + md5 验证可登录）+ sys_user_role + sys_user_post 三表 DB 核对；重复登录账号 remote 即时提示「用户已经存在」+ 后端兜底 500 文案
- [x] 修改用户：回显（含 flag 勾选态）；换部门/换角色/换岗位 → 关联表删旧插新 DB 核对；loginName 只读未变
- [x] 删除：勾选多删 + 行内单删 → sys_user del_flag='2' + 两关联表行消失；ids 含自己 → 「当前用户不能删除」
- [x] 重置密码：弹窗预填 123456 → 提交 → DB md5(loginName+新密码+新salt) 核对；重置自己密码后会话仍有效
- [ ] 分配角色：authRole 标签页表格勾选 → sys_user_role 落库
- [x] 导出：确认框 → 浏览器下载 xlsx（11 列列头/顺序、状态性别转换、手机号文本格式、部门名称/负责人嵌套取值）+ 服务端源文件删除
- [x] 导入：下载模板（7 列）→ 填数 → 导入成功提示逐行文案 + DB 落库（初始密码可登录验证）；已存在不勾更新 → 整批 500「账号 xx 已存在」；勾选更新 → 更新生效（部门编号不改的 quirk 核对）；性别「男」反查 '0'
- [ ] 权限两通道隔离验证（测试角色 + 非 admin 用户）：**遗留至 5.0.0**（沿用 3.0.0 决策；数据权限 SELF 作用域的 u.user_id 语义届时一并验）
- [x] ry 表数据零 schema 变更；预置数据复原（见 checklist 清理记录）

## Task 8 · 收尾（2026-09-29 完成）

- [x] PHPUnit 全绿；spec.md 实施记录回填；checklist 勾选；README 总表更新（状态 ✅ + 日期）
- [x] deviations 处置回填：候选 1（导入用户可登录）与候选 2（权限生效时机）结论落 deviations.md 或注明不登记理由
- [x] 测试数据清理（对 checklist「测试数据清理记录」逐项）：测试用户及关联行、导入测试行、sys_oper_log/sys_logininfor 增量、runtime/download 残留、Redis 测试会话键、admin 会话复原（admin/admin123 可正常登录）
