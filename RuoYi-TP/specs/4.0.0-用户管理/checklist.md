# 4.0.0-用户管理 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（user 21；#[Perm] 18 / #[Log] 8）。

## 用户管理（curl / DB 级）

- [ ] GET /system/user 渲染列表页（权限 system:user:view；未登录 302 /login，ajax 未登录 code "1"）
- [ ] POST /system/user/list：TableDataInfo {code:0,msg:"查询成功",rows,total}；pageSize 分页生效；默认排序 createTime desc；行驼峰 + **dept 嵌套对象**（dept.deptName 可取值）+ **password/salt 键不存在**（grep 响应体断言）；del_flag='2' 行不可见
- [ ] list 搜索：loginName like / status / phonenumber like / params[beginTime]/[endTime]（日粒度）/ deptId 含子树（点深圳总公司 100 → 含 101~106 下用户）逐项生效；parentId 参数被忽略不报错
- [ ] POST /system/user/export：{code:0,msg:"<uuid>_用户数据.xlsx"}；导全量且与 list 同数据权限源；xlsx 读回断言 11 列列头与顺序、性别「男/女/未知」、状态「正常/停用」、手机号文本格式、部门名称/负责人两列嵌套取值、最后登录时间格式（PHPUnit 真文件验证）
- [ ] GET /system/user/importTemplate：{code:0,msg:"<uuid>_用户数据.xlsx"}；xlsx 读回断言 **7 列**（部门编号/登录名称/用户名称/用户邮箱/手机号码/用户性别/账号状态）+ 空数据行；→ GET /common/download 链路通
- [ ] POST /system/user/importData（multipart）：新账号导入成功 → msg「恭喜您，数据已全部导入成功！共 N 条，数据如下：<br/>1、账号 xx 导入成功…」；DB 核对 password=md5(loginName+123456)、**salt 为 NULL**、create_by=admin、**不写角色/岗位关联**；该用户用初始密码登录成功（ deviations 候选 1 行为核对）
- [ ] importData 已存在不勾更新 → 整批 500「很抱歉，导入失败！共 N 条数据格式不正确，错误如下：<br/>…账号 xx 已存在」且**成功行不落库**（整批异常回滚或逐行先判——按经典语义逐行 insert 不回滚已成功行，实测核对经典版口径后固定断言）
- [ ] importData 勾选更新 → 更新生效（update_by=admin）；**导入的部门编号不覆盖库中现部门**（经典 quirk 实测）；目标为 admin → 「不允许操作超级管理员用户」计入失败行
- [ ] importData 校验失败行：「账号 xx 导入失败：登录账号不能为空」等（JSR303 文案全集抽查：登录账号空/超30、邮箱格式、手机超11、脚本字符）；空文件 → 「导入用户数据不能为空！」
- [ ] GET /system/user/add：roles 不含 admin 角色、posts 全量 4 岗、密码框 initPassword
- [ ] POST /system/user/add：校验顺序（部门数据权限 → 角色数据权限 → 登录账号唯一 500「新增用户'xx'失败，登录账号已存在」→ 手机唯一（非空才查）→ 邮箱唯一（非空才查））；成功 → sys_user 落库（salt 6 位 hex、md5(loginName+password+salt) 可验证、pwd_update_date、create_by）+ sys_user_post/sys_user_role 关联落库；roleIds/postIds 空串 → 关联表零行
- [ ] GET /system/user/edit/{userId}：回显 user/roles(flag)/posts(flag)；数据范围外 → 500「没有权限访问用户数据！」（admin 全量已验，非 admin 负向留 5.0.0）
- [ ] POST /system/user/edit：checkUserAllowed（userId=1 → 500）→ 三唯一（「修改用户…」前缀）→ 事务内 user_role/user_post 删旧插新 + user update（update_by/update_time）；**login_name 未被更新**；非空列语义（remark 清空场景核对经典 `<if>` 口径后固定断言）
- [ ] GET /system/user/view/{userId}：roleGroup/postGroup 逗号拼接正确（多角色/无角色两态）；无角色 → 「无角色」、无岗位 → 「无岗位」由页面兜底
- [ ] GET+POST /system/user/resetPwd：新 salt + md5 核对 + pwd_update_date 更新；admin 可被重置（经典原样）；**重置自己 → 会话 user 刷新仍登录有效**；checkUserDataScope 前置
- [ ] GET+POST /system/user/authRole/insertAuthRole：勾选集 → user_role 全删重插 DB 核对；空 roleIds → 全清（清空授权场景）；固定 code 0（非 toAjax，0 行也成功）；#[Log] business_type=**4**
- [ ] POST /system/user/remove：ids 含自己 → 500「当前用户不能删除」；含 admin → 500「不允许操作超级管理员用户」；正常删除 → sys_user del_flag='2' 软删 + sys_user_role/sys_user_post **物理删** DB 核对；重复删已删用户 → toAjax 0 行 → 「操作失败」
- [ ] POST /system/user/checkLoginNameUnique / checkPhoneUnique / checkEmailUnique：裸 boolean true/false；全局唯一口径（ry 手机号 → false）；userId 自身放行；未登录不可达
- [ ] POST /system/user/changeStatus：ry 停用/启用往返 DB 核对；userId=1 → 500「不允许操作超级管理员用户」（quirk：列表页开关可点、后端拦）
- [ ] GET /system/user/deptTreeData：Ztree 裸数组（pId 键名、只含 status='0'、10 部门、数据权限 admin 全量）
- [ ] GET /system/user/selectDeptTree/100：渲染部门树弹窗页；dept 变量回填 treeId/treeName 隐藏域

## 页面级验收（浏览器）

- [ ] 用户列表页：ui-layout 左树右表布局；zTree 10 部门三级结构 + 展开默认 2 级；点击部门 → 表格过滤；重置按钮清隐藏域与选中色；展开/折叠/刷新树按钮
- [ ] 列表渲染：2 行预置（userId/loginName/userName/部门/手机/创建时间）；loginName 链接 → 右滑详情；email 列默认隐藏（列控制可见）；状态开关图标（正常 on / 停用 off）；操作列 admin 行空、ry 行「编辑/删除/更多（重置密码、分配角色）」
- [ ] 工具栏五按钮权限显隐（admin 全可见）；导入按钮 → 弹窗（文件选择 + 是否更新已存在 checkbox + 下载模板 + 仅 xls/xlsx 提示）；导出确认框「确定导出所有用户吗？」→ 浏览器下载
- [ ] 新增页：字段齐全 + 星号必填标记；remote「用户已经存在」「Email已经存在」「手机号码已经存在」即时提示；密码框按下显示/抬起隐藏；部门树弹窗三按钮（确认回填/清除置空/关闭）；角色 checkbox、岗位 select2 多选；提交后回列表页刷新
- [ ] 修改页：loginName 只读灰显；flag 回显（角色勾选/岗位选中）；保存后列表刷新且 DB 关联核对
- [ ] 详情右滑：15 字段只读渲染、岗位/角色组拼接、性别 label
- [ ] 重置密码弹窗（800x300）：预填 123456；保存后父页提示并刷新
- [ ] 分配角色页：client 表格（角色编号/名称/权限字符/创建时间列）；已选角色勾选态；停用角色 disabled 不可勾；保存生效
- [ ] 全部页面在主框架 iframe 内打开、标签页标题正确（modalName：用户；新增/修改/分配角色为标签页形态、重置密码为弹窗形态、详情为右滑形态——四种容器形态逐一核对）

## 横切与纪律自查

- [ ] #[Perm] 18 处与 spec 表逐一对应（view 2/list 4/add 2/edit 5/resetPwd 2/remove 1/export 1/import 1）；check 三端点无注解仅登录态；importTemplate 挂 view
- [ ] #[Log] 8 处落 sys_oper_log（title=用户管理×7 + 重置密码×1；business_type 1/2/3/4/5/6 与 spec 对应）；importData 的 updateSupport 入 oper_param、文件内容不入；password 字段被 OperLog 敏感排除（password/oldPassword 在 1.0.0 排除清单内，实测 resetPwd 的 oper_param 无明文密码）
- [ ] RepeatSubmit **未挂**本模块路由（经典版零 @RepeatSubmit 实锤，负向 grep 核对）
- [ ] 业务代码零处直接 predis（grep）；Redis 只经 RedisCache/SessionService
- [ ] DataScope：selectUserList 双别名（'d','u'）；RoleService selectRoleAll('d')；deptTreeData 复用 3.0.0('d')；SELF 作用域 u.user_id 语义实现就位（非 admin 验证留 5.0.0 注明）
- [ ] 权限两通道一致：sys_menu.perms 8 行（view/list/add/edit/remove/export/import/resetPwd）与 #[Perm] 注解、check_perm 调用互查无遗漏
- [ ] password/salt 全链路不泄出（list 行、日志、异常消息 grep）
- [ ] 表结构零变更（SHOW COLUMNS 前后一致）；qrtz_/demo 表未触碰
- [ ] 全局类 use 纪律自查（PageQuery/TableDataInfo/AjaxResult/PasswordService/ExcelExportService 等 namespaced 文件全 use——3.0.0 三次踩坑预防）

## Deviations 核对

- [ ] spec「拟登记 deviations」处置回填：候选 1（导入用户 salt 为空可登录）与候选 2（授权变更在线用户生效时机）结论落 deviations.md 或注明不登记理由
- [ ] 前端/后端 quirk 原样保留未"修复"：remote 文案「用户已经存在」/ admin 行开关可点后端拦 / 导入更新分支不改部门 / importTemplate 权限 view / 新增页 treeId 空硬编码 "100" / 编辑不改 login_name

## 测试数据清理记录

（动工后逐条登记；清理完成在条目前加 ✅ 并注日期）

- [ ] 测试用户（新增/导入行）及 sys_user_role / sys_user_post 关联行物理 DELETE；预置 admin(1)/ry(2) 及其关联 2+2 行恢复原值（密码/salt/login_date 等——resetPwd 测试后必须复原 admin 密码 md5('admin'+'admin123'+'111111')）
- [ ] sys_oper_log / sys_logininfor 测试增量（可选：保留并注明，或清理）
- [ ] runtime/download 导出/模板残留文件删除
- [ ] Redis：测试会话键清理；dict: 预热键可保留
- [ ] admin 会话状态复原（admin/admin123 可正常登录，会话内 user 数据与预置一致）
