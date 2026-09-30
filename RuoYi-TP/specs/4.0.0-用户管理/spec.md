# 4.0.0-用户管理 · spec

> **状态：✅ 已完成（2026-09-29，curl/DB + 浏览器级端到端通过）**
> 对位经典若依 SysUserController + SysUserServiceImpl / SysUserMapper.xml + templates/system/user（部门树联动、导入导出、重置密码、分配角色；**profile 归 8.0.0，本模块不做**；用户 21 端点 / 页面 7 个，逐方法核对实锤）。
> 依赖：01（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / DataScope / LoginAuth / 防重）+ 03（DeptService 的 selectDeptTreeData / checkDeptDataScope / selectDeptById 复用）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下。
> 调研依据：reference 源码逐文件核对（SysUserController / SysUserServiceImpl / SysUserMapper.xml / SysUser 实体 / 页面 HTML×7 / SysRoleServiceImpl·SysPostServiceImpl 的 flag 合并逻辑 / ExcelUtil 的 EXPORT·IMPORT 列过滤与表头名匹配导入 / DataScopeAspect 的 userAlias 分支 / SysPasswordService.encryptPassword）+ **ry-tp 库实测**（sys_user 21 列 / sys_user_role·sys_user_post 复合主键双列；预置用户 2 行 admin(1)/ry(2)、用户角色 2 行 1→1/2→2、用户岗位 2 行 1→1/2→2、岗位 4 行、部门 10 行；sys_menu perms 8 行 view/list/add/edit/remove/export/import/resetPwd；sys_config `sys.user.initPassword=123456`、`sys.account.chrtype=0`）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 0/301/500、POST 分页、信封结构、权限两通道、md5(loginName+password+salt) 密码方案。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/system/UserController.php | SysUserController | 21 个方法（无双 pattern，21 条路由） |
| app/service/UserService.php | ISysUserService / SysUserServiceImpl | 列表（DataScope 双别名）、三唯一校验、admin/数据权限双防护、事务关联维护、导入校验链、角色/岗位组拼接 |
| app/service/RoleService.php（**最小只读**） | SysRoleServiceImpl 的 selectRoleAll / selectRolesByUserId / checkRoleDataScope | 仅用户模块消费的三方法（含 flag 合并）；完整角色管理 5.0.0 落地时收编 |
| PostService 增补 | SysPostServiceImpl 的 selectPostAll / selectPostsByUserId | 全量岗位 + flag 合并（现有 selectPostList 之外新增两方法） |
| ExcelExportService 增补 | ExcelUtil 的 importExcel / importTemplateExcel | 导入解析（表头名匹配 + readConverterExp 反向转换 + 文本列）与 IMPORT 列模板生成；导出复用 3.0.0 |
| app/model/SysUser.php | SysUser 实体 | 表映射；sys_user_role / sys_user_post 复合主键无自增，**不建模型**，service 内 Db 直操作 |
| app/view/system/user/*.html ×7 | templates/system/user/ 7 页 | user/add/edit/view/resetPwd/authRole/deptTree；静态 JS（ry-ui.js addTab·saveTab·popupRight / select2 / zTree / validate）零改动对接 |

**范围外**（后续模块，勿在本模块实现）：profile / checkPassword / updateAvatar（8.0.0 个人中心）；selectAllocatedList / selectUnallocatedList（5.0.0 角色已分配/未分配用户）；角色管理界面与 sys_role 写维护（5.0.0）；注册 /register（8.0.0 拍板）。

## 页面清单（7）

| # | 页面 | TP 模板路径 | 对位经典版 | 组件/形态 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 用户列表页 | view/system/user/index.html | user/user.html | **ui-layout 左树右表**：左 zTree 组织机构树 + 右 bootstrap-table（POST server 分页）；新增/修改走**标签页**非弹窗 | `prefix = ctx + "system/user"`；options：url=prefix+/list、**viewUrl=prefix+/view/{id}**（loginName 列链接 → $.operate.view 右滑出）、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove（无 {id}）、exportUrl、**importUrl=prefix+/importData、importTemplateUrl=prefix+/importTemplate**；sortName=createTime sortOrder=desc；modalName=用户；工具栏五按钮（新增 addTab/修改 editTab single/删除 removeAll multiple/导入 importExcel/导出 exportExcel，权限串 add/edit/remove/import/export）；列：checkbox、userId、loginName（sortable+链接）、userName、**dept.deptName（嵌套取值）**、email（visible:false）、phonenumber、状态开关列（editFlag 控制可见，statusTools 开关图标 → confirm → `$.operate.post(prefix+"/changeStatus", {userId, status})`）、createTime（sortable）、操作列（**userId!=1 才渲染**：编辑/删除/更多弹层（重置密码 `$.modal.open(prefix+'/resetPwd/'+id, 800x300)` + 分配角色 addTab authRole））；zTree `url = ctx+"system/user/deptTreeData"`，onClick 写 #deptId+#parentId 隐藏域 → $.table.search()；重置 resetPre 清两隐藏域+去 curSelectedNode；展开/折叠/刷新树按钮；dept() 快捷 openTab 部门管理；窄屏 (<769) 自动折叠左栏；页尾 `#importTpl` 导入模板脚本（file + updateSupport checkbox + 下载模板 + 仅 xls/xlsx 提示） |
| 2 | 新增用户页（标签页） | view/system/user/add.html | user/add.html | 表单 + select2 岗位多选 + zTree 部门弹窗 | validate：loginName minlength2/maxlength20 + **remote POST prefix+/checkLoginNameUnique**（文案「用户已经存在」）、password 5-20+specialSign、email+remote checkEmailUnique（「Email已经存在」）、phonenumber isPhone+remote checkPhoneUnique（「手机号码已经存在」）；密码框预填 `sys.user.initPassword`（123456）+ 按下显示；sex 字典 sys_user_sex；status 开关默认正常；角色 checkbox（模板变量 roles=**去 admin 角色**，status=1 disabled）；岗位 select2 multiple（posts 全量，status=1 disabled）；归属部门输入框点击 selectDeptTree()（treeId 空则用 **"100"**）→ `$.modal.openOptions` 打开 prefix+/selectDeptTree/{deptId}（确认/清除/关闭三按钮，清除清空 treeId/treeName）；提交 `checkpwd(chrtype)`（chrtype=sys.account.chrtype 默认 0 不校验）→ saveTab(prefix+"/add", serializeArray + status + roleIds(selectCheckeds "role" 逗号串) + postIds(selectSelects "post" 逗号串))；关闭 closeItem() |
| 3 | 修改用户页（标签页） | view/system/user/edit.html | user/edit.html | 同新增 | **loginName 只读**（rules 无 loginName remote）；email/phonenumber remote **多带 userId**；岗位 selected=flag、角色 checked=flag（模板 flag 标记）；提交 saveTab(prefix+"/edit")；选部门同新增 |
| 4 | 用户详情（右侧滑出） | view/system/user/view.html | user/view.html | 只读 form-control-plaintext | 字段：userName / dept.deptName / phonenumber / email / loginName / status(正常停用) / 岗位（postGroup，空→「无岗位」）/ sex（字典 sys_user_sex 取 label）/ 角色（roleGroup，空→「无角色」）/ createBy / createTime / updateBy / updateTime / loginIp / loginDate / remark；无 JS 交互 |
| 5 | 重置密码弹窗 | view/system/user/resetPwd.html | user/resetPwd.html | 表单（modal 弹窗 800x300） | userId 隐藏 + loginName 只读回显（**仍随表单提交**，后端加密要用）+ password 预填 initPassword；validate password required/5-20/specialSign；提交 `$.operate.save(ctx+"system/user/resetPwd", serialize)`（save 非 saveTab，成功回调刷父表） |
| 6 | 分配角色页（标签页） | view/system/user/authRole.html | user/authRole.html | bootstrap-table（**client 分页**，data=模板变量 roles） | `prefix = ctx + "system/user/authRole"`；showSearch/Refresh/Toggle/Columns 全关、clickToSelect、maintainSelected；列：checkbox（formatter checked=flag，disabled=status=='1'）、roleId、roleSort（隐藏 sortable）、roleName、roleKey（sortable）、createTime；提交遍历 `bootstrapTable('getData')` 收集勾选行 roleId → saveTab(prefix+"/insertAuthRole", {userId, roleIds: join(",")}) |
| 7 | 部门树选择弹窗 | view/system/user/deptTree.html | user/deptTree.html | zTree | 隐藏域 treeId/treeName 初始值 = 模板变量 dept；`url = prefix + "/deptTreeData"`（**user 模块自己的端点，非 dept 模块 treeData；无 excludeId 语义**——用户归属部门可选任意部门含自身子树）；expandLevel=2；onClick 回填 treeId/treeName；搜索/展开/折叠纯前端 |

> TP 模板命名 index/add/edit/view/resetPwd/authRole/deptTree；控制器渲染路径用 layer 相对路径（`system.UserController` → 视图根 `app/view/system/`，渲染 'user/index' 等——3.0.0 联调实锤）。include 片段用 app/view/system/include/ 局部副本（select2-css/js、ztree-css/js 副本已就位）。

## 端点级 API 清单（user 21）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/user | system:user:view | 无 | — | 渲染页面 1（左树右表 layout） |
| 2 | POST | /system/user/list | system:user:list | 无 | pageNum/pageSize/orderByColumn/isAsc（白名单 login_name/create_time；首请求 orderByColumn=createTime&isAsc=desc）+ loginName（like）/ status / phonenumber（like）/ params[beginTime]、params[endTime]（create_time 按 %Y%m%d 日粒度比较）/ deptId（树点击）/ parentId（树点击附带，**后端忽略**） | **TableDataInfo** {code:0, msg:"查询成功", rows, total}；固定 u.del_flag='0'；deptId 命中含子部门（`u.dept_id=X OR u.dept_id IN (SELECT dept_id FROM sys_dept WHERE FIND_IN_SET(X, ancestors))`）；**行结构对位 Jackson**：驼峰键 + **dept 嵌套对象**（deptId/parentId/deptName/leader/orderNum/status，columns field 'dept.deptName' 按嵌套取值）；**password/salt 不输出**（@JsonIgnore 对位——列表 SQL 虽查了这两列，组装行时剔除）；时间 Y-m-d H:i:s，NULL 透传 |
| 3 | POST | /system/user/export | system:user:export | 用户管理, 5导出 | 同搜索参数 + 排序；**无分页（导全量）**，走带数据权限的 selectUserList | {code:0, msg:"&lt;uuid&gt;_用户数据.xlsx"}——文件名装 msg；前端 GET /common/download；Excel 11 列见下节 |
| 4 | POST | /system/user/importData | system:user:import | 用户管理, 6导入 | **multipart**：file（xls/xlsx）+ updateSupport（checkbox 勾选传 "on"，未勾不传） | 全部成功 → {code:0, msg:"恭喜您，数据已全部导入成功！共 N 条，数据如下：&lt;br/&gt;1、账号 xx 导入成功…"}；任一失败 → **整批抛 500** {code:500, msg:"很抱歉，导入失败！共 N 条数据格式不正确，错误如下：&lt;br/&gt;…"}（含「账号 xx 已存在」「账号 xx 导入失败：&lt;原因&gt;」行）；校验链见特殊行为 5 |
| 5 | GET | /system/user/importTemplate | system:user:view | 无 | — | {code:0, msg:"&lt;uuid&gt;_用户数据.xlsx"}——**模板只有 IMPORT/ALL 型 7 列**（见 Excel 节）；权限是 view 不是 import（经典原样实锤） |
| 6 | GET | /system/user/add | system:user:add | 无 | — | 渲染页面 2；变量 roles（selectRoleAll 去 roleId=1）、posts（selectPostAll 全量）、initPassword（模板预填密码框） |
| 7 | POST | /system/user/add | system:user:add | 用户管理, 1新增 | deptId（隐藏 treeId）、userName*、loginName*、password*、phonenumber、email、sex、status、roleIds（逗号串可空）、postIds（逗号串可空）、remark | 校验顺序（经典实锤）：checkDeptDataScope(deptId，空跳过) → checkRoleDataScope(roleIds) → 登录账号唯一「新增用户'{loginName}'失败，登录账号已存在」(500) → 手机唯一（非空才查）「…手机号码已存在」→ 邮箱唯一（非空才查）「…邮箱账号已存在」；salt=randomSalt + **md5(loginName+password+salt)** + pwd_update_date=now + create_by；**事务**：sys_user insert + sys_user_post batch + sys_user_role batch；toAjax |
| 8 | GET | /system/user/edit/{userId} | system:user:edit | 无 | 路径 userId；**先 checkUserDataScope** | 渲染页面 3；变量 user（含 dept 嵌套）、roles（selectRolesByUserId 全量+flag 合并；**目标用户非 admin 时过滤 admin 角色**）、posts（selectPostsByUserId 全量+flag） |
| 9 | GET | /system/user/view/{userId} | system:user:list | 无 | 路径 userId；先 checkUserDataScope | 渲染页面 4（右滑出）；变量 user、roleGroup（角色名逗号拼接，空返回空串）、postGroup（岗位名逗号拼接） |
| 10 | POST | /system/user/edit | system:user:edit | 用户管理, 2修改 | userId、deptId、userName*、phonenumber、email、sex、status、roleIds、postIds、remark；loginName 只读框仍提交（供唯一性文案）但 **update 不改 login_name**（mapper set 无此列，经典实锤） | 校验顺序：checkUserAllowed（拦 userId=1「不允许操作超级管理员用户」500）→ checkUserDataScope → checkDeptDataScope → checkRoleDataScope → 三唯一（前缀「修改用户…」）；update_by=登录名；**事务**：user_role 删+插 → user_post 删+插 → update user（set 含 dept_id/user_name/user_type/email/phonenumber/sex/avatar/password/salt/status/login_ip/login_date/pwd_update_date/update_by/remark + update_time，按非空列更新）；toAjax |
| 11 | GET | /system/user/resetPwd/{userId} | system:user:resetPwd | 无 | 路径 userId；先 checkUserDataScope | 渲染页面 5；变量 user（userId/loginName 回显）、initPassword 预填 |
| 12 | POST | /system/user/resetPwd | system:user:resetPwd | **重置密码**, 2修改 | userId、loginName（只读框提交）、password | checkUserAllowed → checkUserDataScope；salt=randomSalt + md5(loginName+password+salt)；resetUserPwd 更新 password/salt/pwd_update_date/update_time；**重置对象=当前登录用户时刷新会话 user**（setSysUser 对位，SessionService 写回）；>0 → success()，否则 error()（非 toAjax） |
| 13 | GET | /system/user/authRole/{userId} | system:user:edit | 无 | 路径 userId；先 checkUserDataScope | 渲染页面 6；变量 user、roles（同 edit 的 flag 合并与 admin 过滤规则） |
| 14 | POST | /system/user/authRole/insertAuthRole | system:user:edit | 用户管理, **4授权** | userId、roleIds（逗号串，可空串=清空授权） | checkUserDataScope → checkRoleDataScope；**事务**：user_role 按 userId 全删 + 批量插；固定 success()（非 toAjax，0 行也算成功） |
| 15 | POST | /system/user/remove | system:user:remove | 用户管理, 3删除 | ids（逗号串；单行删除同样走 ids body） | ids 含当前登录用户 → error(500)「当前用户不能删除」；逐个 checkUserAllowed + checkUserDataScope；**事务**：sys_user_role / sys_user_post 按 user_ids **物理删** + sys_user **软删** del_flag='2'；toAjax |
| 16 | POST | /system/user/checkLoginNameUnique | **无（仅登录态）** | 无 | loginName（+编辑时 userId） | **裸 boolean**：login_name 全局唯一（del_flag='0' limit 1）；userId 相同视为自身放行 |
| 17 | POST | /system/user/checkPhoneUnique | **无（仅登录态）** | 无 | phonenumber（+userId） | 裸 boolean；phonenumber 全局唯一 del_flag='0' |
| 18 | POST | /system/user/checkEmailUnique | **无（仅登录态）** | 无 | email（+userId） | 裸 boolean；email 全局唯一 del_flag='0' |
| 19 | POST | /system/user/changeStatus | system:user:edit | 用户管理, 2修改 | userId、status | checkUserAllowed（拦 admin——**列表页 admin 行开关可见可点，点击后端拦截报「不允许操作超级管理员用户」，经典原样 quirk**）→ checkUserDataScope；update status + update_time；toAjax |
| 20 | GET | /system/user/deptTreeData | system:user:list | 无 | — | **Ztree 裸数组** {id, pId, name, title, checked:false, open:false, nocheck:false}；只含 status='0'；带数据权限——等价 `DeptService::selectDeptTreeData(0, $user)` 复用（3.0.0 已实现） |
| 21 | GET | /system/user/selectDeptTree/{deptId} | system:user:list | 无 | 路径 deptId（新增页 treeId 空传 100） | 渲染页面 7，变量 dept（selectDeptById 复用 3.0.0） |

统计：**控制器方法 21 个；URL pattern 21 条**（无双 pattern）；**#[Perm] 共 18 处**（view 2 / list 4（list、view、deptTreeData、selectDeptTree）/ add 2 / edit 5（edit GET、editSave、authRole GET、insertAuthRole、changeStatus）/ resetPwd 2 / remove 1 / export 1 / import 1；#16/17/18 三个 check 端点无注解仅登录态）；**#[Log] 共 8 处**（用户管理 7：EXPORT/IMPORT/INSERT/UPDATE×2/GRANT/DELETE + 重置密码 1：UPDATE）。

## 特殊行为清单（唯一性文案 / admin 保护 / 关联表 / 导入校验 / 密码）

1. **唯一性校验文案全集**（前后端双保险，文案不同属经典原样）：
   - 后端 add/edit：`新增用户'{loginName}'失败，登录账号已存在` / `修改用户'{loginName}'失败，登录账号已存在`；`…失败，手机号码已存在`；`…失败，邮箱账号已存在`（手机/邮箱仅非空才校验）
   - 前端 remote 提示：登录账号「**用户已经存在**」（不是"登录账号已经存在"——原样）；「Email已经存在」「手机号码已经存在」
   - 口径：三者均**全局唯一**（无部门维度），查 del_flag='0' limit 1，userId 相同视为自身放行；edit 页 loginName 只读故 checkLoginNameUnique remote 只挂 add 页
2. **admin 保护三件**：① checkUserAllowed（editSave/resetPwdSave/changeStatus + remove 逐个 + 导入更新分支）userId=1 → 500「不允许操作超级管理员用户」（TP 对位 `userId === 1` 判断）；② remove ids 含当前登录用户 → 500「当前用户不能删除」（在 checkUserAllowed 之前）；③ 列表操作列 userId!=1 才渲染按钮；**状态开关列对 admin 行仍渲染可点**（后端拦截兜底，quirk 原样保留）。resetPwd **无** admin 禁止（admin 密码可被重置，经典原样——仅数据权限拦非 admin）。
3. **关联表写维护**（四处事务，roleIds/postIds 空串或未传 → 跳过对应关联写，对位 `isNotNull + size>0`）：① insertUser：user insert + user_post batch + user_role batch；② updateUser：**先角色后岗位**（user_role 删→插、user_post 删→插）+ user update；③ insertUserAuth：user_role 删→插；④ deleteUserByIds：user_role/user_post **物理删** + user **软删**。
4. **密码方案**：addSave/resetPwdSave `PasswordService::randomSalt()`（6 位 hex）+ `md5(loginName . password . salt)` + pwd_update_date=now（PasswordService 0.0.0 已有）。**导入例外**：`md5(loginName . sys.user.initPassword)`，**salt 不写（保持 NULL）**——经典原样 quirk，衍生行为差异见拟登记 deviations 1。
5. **导入校验链**（importUser 逐行）：新账号 → JSR303 校验（文案同上第 12 条全集：登录账号 NotBlank/≤30/防脚本、用户昵称 ≤30/防脚本、邮箱格式/≤50、手机 ≤11）→ checkDeptDataScope(导入部门编号) → 初始密码 + create_by + insert；已存在 + updateSupport → 同校验 + checkUserAllowed + checkUserDataScope + checkDeptDataScope + **deptId 用库中现值覆盖（导入的「部门编号」列在更新分支不生效——经典原样 quirk）** + update（update_by）；已存在 + 不更新 → failure「账号 xx 已存在」；任何异常 → failure「账号 xx 导入失败：」+ 消息（**校验类异常时 loginName 经 HTML 转义**，EscapeUtil.clean 对位 htmlspecialchars）；**不写角色/岗位关联**（Excel 无此列）；空列表 → 「导入用户数据不能为空！」。
6. **列表输出**：password/salt 剔除（@JsonIgnore 对位，**严禁泄出**）；dept 嵌套对象供 `dept.deptName` 取值；deptId/parentId 双隐藏域随表单提交，后端只消费 deptId。
7. **会话刷新**：resetPwd 重置对象为当前登录用户 → SessionService 重新 selectUserById 写回会话 user（不新增会话字段）。
8. **部门树联动**：列表页 zTree onClick → deptId+parentId 隐藏域 → 搜索；deptTreeData 带 status='0' 过滤 + 数据权限（复用 3.0.0）；新增/修改页选部门弹窗走 **user 自己的 deptTreeData**（无 excludeId，可选任意部门）。
9. **后端参数校验文案**（对位 @Validated + BindException → error 首条）：「登录账号不能为空」「登录账号长度不能超过30个字符」「登录账号不能包含脚本字符」（Xss：正则 `<(\\S*?)[^>]*>.*?|<.*? />` 含 HTML 判违规）「用户昵称长度不能超过30个字符」「用户昵称不能包含脚本字符」「邮箱格式不正确」「邮箱长度不能超过50个字符」「手机号码长度不能超过11个字符」。
10. **防重复提交**：user 控制器零处 @RepeatSubmit（grep 实锤）→ 不挂 RepeatSubmit 中间件。
11. **前端 quirk（原样保留勿修）**：admin 行状态开关可点（后端拦）；remote 文案「用户已经存在」；check 三端点无 #[Perm]；importTemplate 权限是 system:user:view；新增页 treeId 空时选部门 URL 硬编码 "100"。
12. **编辑页更新列**：update set 按「非空才更新」语义对位 mapper `<if>` 列表（dept_id 需非 0 才更新）；login_name/status 外的 avatar/login_ip/login_date/pwd_update_date 编辑表单不涉及，天然不动。

## 数据权限声明

- **UserService::selectUserList 带 DataScope 双别名 deptAlias="d" + userAlias="u"**——本工作区首个双别名用例（`DataScope::apply($query, $user, 'd', 'u')`，1.0.0 封装已支持 userAlias 参数）；**仅本人（DATA_SCOPE_SELF）作用域靠 u.user_id 条件生效**，无 userAlias 时该 scope 退化为不查任何数据（DataScopeAspect 114-124 行实锤，TP 封装需含同语义——实测时用仅本人角色验证）。
- **checkUserDataScope**：非 admin 用带数据权限的 selectUserList(userId) 判空 → 「没有权限访问用户数据！」(500)；edit GET / editSave / resetPwd GET/POST / authRole GET / insertAuthRole / remove（逐个） / changeStatus / 导入更新分支均挂。
- **checkDeptDataScope**：addSave/editSave/导入两分支复用 3.0.0 DeptService（「没有权限访问部门数据！」）。
- **checkRoleDataScope**：经典 selectRoleList 带 @DataScope(deptAlias="d")，SQL 经 sys_user_role/sys_user/sys_dept 链 join——RoleService 最小只读需复刻该 join + DataScope('d')（「没有权限访问角色数据！」）；非 admin 逐个 roleId 校验。
- deptTreeData：DeptService::selectDeptTreeData 已带 DataScope('d')（3.0.0）。
- **export 与 list 同源**（selectUserList）→ 导出结果同样受数据权限过滤（经典一致）。

## Excel 导入导出列定义（SysUser @Excel 逐字段抄录）

### 导出（EXPORT + ALL 型，stable sort = 声明序，共 11 列）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | userId | name="用户序号", type=EXPORT, cellType=NUMERIC, prompt="用户编号" | 列头「用户序号」，数字格式 |
| 2 | loginName | name="登录名称" | 文本 |
| 3 | userName | name="用户名称" | 文本 |
| 4 | email | name="用户邮箱" | 文本 |
| 5 | phonenumber | name="手机号码", cellType=TEXT | **文本格式**（防科学计数/丢前导零） |
| 6 | sex | name="用户性别", readConverterExp="0=男,1=女,2=未知" | 转换输出文本 |
| 7 | status | name="账号状态", readConverterExp="0=正常,1=停用" | 转换输出文本 |
| 8 | loginIp | name="最后登录IP", type=EXPORT | 文本 |
| 9 | loginDate | name="最后登录时间", width=30, dateFormat="yyyy-MM-dd HH:mm:ss", type=EXPORT | 日期格式化输出 |
| 10 | dept（@Excels） | name="部门名称", targetAttr="deptName", type=EXPORT | 嵌套取 dept.dept_name |
| 11 | dept（@Excels） | name="部门负责人", targetAttr="leader", type=EXPORT | 嵌套取 dept.leader |

- deptId 的 @Excel 为 type=**IMPORT** → 导出不含「部门编号」列；BaseEntity 字段无 @Excel → 不导出；sheet 名「用户数据」；文件名 `<uuid>_用户数据.xlsx`（与模板同名，uuid 区分）。
- @Excel sort() 默认 MAX_VALUE，Java stream sorted 稳定排序 → 列序 = 字段声明序（上表即实锤列序）。

### 导入模板（IMPORT + ALL 型，共 7 列，importTemplate 生成空表）

| 序 | Java 字段 | @Excel 注解 | 导入解析要点 |
|---|---|---|---|
| 1 | deptId | name="部门编号", type=IMPORT | 数字；仅新增分支生效（更新分支被库中现值覆盖，见特殊行为 5） |
| 2 | loginName | name="登录名称" | 文本 |
| 3 | userName | name="用户名称" | 文本 |
| 4 | email | name="用户邮箱" | 文本 |
| 5 | phonenumber | name="手机号码", cellType=TEXT | 按文本读 |
| 6 | sex | name="用户性别", readConverterExp="0=男,1=女,2=未知" | **反向转换**：男→0/女→1/未知→2 |
| 7 | status | name="账号状态", readConverterExp="0=正常,1=停用" | 反向转换：正常→0/停用→1 |

- 导入按**表头名匹配**列（cellMap 对位），列序无关；空行跳过；解析异常 → UtilException「导入Excel异常」消息透出（对位消息透传到导入失败行）。

## 关键设计说明

1. **list 行结构是本模块响应格式关键点**：TableDataInfo 信封 + 行驼峰 + **dept 嵌套对象** + password/salt 剔除；`dept.deptName` 嵌套取值要求行内必须有 dept 对象（无部门用户给空对象，bootstrap-table 取 undefined 显示空）。
2. **PageQuery 白名单**：['login_name', 'create_time']（页面可排序列仅 loginName/createTime）；orderByColumn 驼峰转下划线复用 1.0.0。
3. **复用 3.0.0 资产**（不重复造）：ExcelExportService（导出；本模块增导入解析与 IMPORT 模板生成）、GET /common/download、DictService（sys_normal_disable / sys_user_sex，**view 页需 label 反查——DictService 增最小 `getLabel(type, value)`**，6.0.0 收编）、DeptService（selectDeptTreeData / checkDeptDataScope / selectDeptById）、PostService（增 selectPostAll / selectPostsByUserId）、PasswordService。
4. **RoleService 最小只读**（5.0.0 收编接管）：selectRoleAll（del_flag='0'，join user_role/user/dept 链 + DataScope('d') 对位 selectRoleContactVo）、selectRolesByUserId（全量角色 + flag=true 合并用户已有角色）、checkRoleDataScope。角色行输出键驼峰：roleId/roleName/roleKey/roleSort/dataScope/status/createTime/flag（authRole 页 client 分页消费 + add/edit 页 checkbox volist 消费）。
5. **导入解析**（phpspreadsheet 2.0，注意 getCell 坐标数组 API）：读首行表头建 name→col 映射；readConverterExp 反向转换；phonenumber 列按文本读（`ValueBinder` 或格式化值兜底）；返回以表头名为键的行数组列表。
6. **multipart**：importData 用 `$request->file('file')`；OperLog 参数记录天然不含文件内容（$request->param() 不含 $_FILES），updateSupport 会入日志（经典一致）。
7. **事务边界四处**（insertUser / updateUser / insertUserAuth / deleteUserByIds）`Db::startTrans/commit/rollback`（对位 @Transactional）。
8. **ids 解析**：remove ids、roleIds、postIds 逗号串 → array_map('intval', explode(',')) + 空串过滤（selectCheckeds 空选返回 ""）。
9. **getSysUser/isAdmin**：会话直接读取（2.0.0 已含 userId/loginName/deptId/permissions/roles/isAdmin）；isAdmin = user_id==1（SysUser.isAdmin 判据）。

## 拟登记 deviations

经逐点核对（响应格式 / 权限串 / 校验文案 / 关联维护语义 / 导入校验链 / 导出列），核心行为均可与经典版一致。候补两条，实施时定：

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| 1 导入用户 salt 为空时的登录行为 | importUser 存 `md5(loginName+初始密码)` 不写 salt；登录时 Java 对 null 做 `"null"` 串接 → 哈希不匹配，**导入用户实际无法登录**（经典 bug） | PHP null 拼接为空串 → `md5(loginName+密码+'')` 恰好匹配 → 导入用户**可用初始密码登录** | 顺带修正经典 bug 且符合 sys.user.initPassword 设计意图；倾向保留可登录并登记 deviations（实施时定） |
| 2 角色/授权变更后在线用户权限生效时机 | AuthorizationUtils.clearAllCachedAuthorizationInfo() 清 Shiro 缓存，被改用户**下次鉴权即生效** | TP 会话 permissions 登录时算好存 Redis；编辑用户角色 / insertAuthRole 后该用户**需重新登录**生效 | 无统一权限缓存设施可清；登记差异，5.0.0 复核是否做目标会话权限重算 |

另注（非 deviation，防误改说明）：check 三端点无权限注解、importTemplate 用 view 权限、remote 文案「用户已经存在」、导入更新分支不改部门、admin 行状态开关可点后端拦、编辑不改 login_name——均为经典版原样行为，**照抄不修正**。

## 实施记录

- 2026-09-29 全部 8 个 Task 完成（PHPUnit 40 tests 118 assertions；curl/DB 级 + 浏览器级端到端通过）。
- **浏览器级实测**：左树（zTree 10 部门）右表（2 行预置）；新增标签页（密码预填 123456/角色 checkbox/岗位 select2 4 项）→ 提交 → sys_user + user_role + user_post 三表落库；编辑回显/换部门/关联换血（DB ur=1/up=1 核对）；状态开关（确认框 → changeStatus → DB，admin 行后端拦）；loginName 链接右滑详情（岗位「项目经理」/角色「普通角色」/性别「女」全渲染）；重置密码弹窗（loginName 回显 + 密码预填）→ 提交 → **md5('ry'+'RyNew@2026'+salt) DB 精确匹配**；导出/导入模板 → 下载；导入三分支（不更新「已存在」整批 500 / 更新生效且 deptId 不改 quirk / 成功消息拼装）+ **导入用户实测可登录**（deviations #20）。
- **新语法坑（模板子任务实测发现，后续模块适用）**：`{:...}` 输出标签经 parseTag 的 stripslashes，命名空间调用 `{:\app\...\X::m()}` 会被剥反斜杠——命名空间调用须走 `{assign name="x" value=":..." /}` 再 `{$x}`。
- **命名空间 use 坑第 4~5 次**：UserController `use app\service\PasswordService`（实际是全局类）、UserService 缺 `use PasswordService`——**规范定死：app/common/ 下五类（TpConstant/RedisCache/AjaxResult/TableDataInfo/PasswordService/PageQuery/DataScope/BusinessException）一律 `use 类名;` 短引入，禁止写 app\service\ 前缀**。
- deviations #20（导入用户可登录）、#21（授权重登录生效）已登记。
