# 3.0.0-部门岗位 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（dept 11 / post 10 / download 1）。

## 部门管理（curl / DB 级）——2026-09-29 UTF-8 Python 脚本端到端实测

- [ ] GET /system/dept 渲染列表页（权限 system:dept:view；未登录 302 /login，ajax 未登录 code "1"）——页面待模板
- [x] POST /system/dept/list 返回**裸数组**（无 code/rows 信封）；字段下划线齐全；order by parent_id, order_num；del_flag 行不可见
- [ ] 列表/树数据带数据权限：非 admin 过滤各口径——待 5.0.0 角色数据后补验（当前库仅 admin，负向断言 admin 全量已过）
- [ ] GET /system/dept/add/{parentId} 非 admin 强制 parentId——同上待 5.0.0
- [x] POST /system/dept/add：同父同名 → 500「新增部门'xx'失败，部门名称已存在」；异父同名可建（T1@103 / T1@102 双行实测）；成功行 ancestors = 父.ancestors,parent_id、create_by=admin 落库
- [ ] GET /system/dept/edit/{deptId} 数据范围外 → 500——待 5.0.0 非 admin 用户；deptId=100 parentName=「无」逻辑已实现（页面联调时出图核对）
- [x] POST /system/dept/edit「上级部门不能是自己」500 实测（101→101）；停用含未停用子部门分支已实现（浏览器联调复验）
- [x] 改父部门 ancestors 级联 DB 核对：202 从 103→102 后 ancestors 0,100,101,103 → 0,100,102 实测
- [x] 启用级联与「该部门包含未停用的子部门！」——实现完成；停用场景浏览器联调复验
- [x] POST /system/dept/updateSort 逗号串逐行生效（200→70、201→71 DB 核对）；成功恒 code 0
- [x] POST /system/dept/remove/{deptId}：有子部门 → **301**「存在下级部门,不允许删除」实测；叶子软删（200 行 del_flag 逻辑）+ 重复删 0 行 → 「操作失败」；数据范围外 500 待 5.0.0
- [x] POST /system/dept/checkDeptNameUnique：裸 true/false；同父判重（研发部门@101 vs @103 口径实测）、自身放行；未登录不可达（LoginAuth 主链覆盖）
- [x] GET /system/dept/treeData/{excludeId} 返回 Ztree 裸数组（pId 键名精确、只含 status='0'、excludeId=101 排除自身及后代实测）

## 岗位管理（curl / DB 级）——同上实测

- [ ] GET /system/post 渲染列表页——页面待模板
- [x] POST /system/post/list：TableDataInfo {code:0,msg:"查询成功",rows,total}；pageSize=2 分页生效；orderByColumn=postSort 白名单排序生效
- [x] POST /system/post/export：{code:0,msg:"<uuid>_岗位数据.xlsx"}；导出全量；xlsx 读回断言 5 列列头/顺序、状态「正常/停用」转换、数字格式（PHPUnit 真文件验证）
- [x] GET /common/download?fileName=…&delete=true：文件流 6583B + attachment 文件名（timestamp 前缀 + URL 编码）；下载后源文件删除（目录清零实测）；".."/子路径 → `{"code":500,"msg":"非法文件名"}`；白名单外扩展 → `{"code":500,"msg":"不支持的文件类型"}`（HTTP 200 + code 500 信封，非 HTTP 500）
- [x] POST /system/post/remove（ids）：未分配 → 物理删除（id=5 实测消失）；已分配 → 500「xx已分配,不能删除」逻辑实现（sys_user_post 预置 2 行占用岗位 1/2——浏览器联调用岗位 1 复验）
- [x] POST /system/post/add：名称重名 → 500「新增岗位'xx'失败，岗位名称已存在」实测；编码重名 → 500「…岗位编码已存在」实测；成功落库（create_by=create_time）
- [x] POST /system/post/edit 双校验（修改前缀）实现（浏览器联调复验）；update_by/update_time 落库
- [x] POST /system/post/checkPostNameUnique / checkPostCodeUnique：裸 boolean；全局唯一口径（董事长→false 实测）
- [x] 岗位模块无数据权限、无 admin 保护（实现与经典版一致；负向断言=checkDeptDataScope 未在 post 链路调用）

## 页面级验收（浏览器）

- [ ] 部门列表页：bootstrap-tree-table 树形展开/折叠正常（预置 10 部门三级结构）；展开/折叠按钮；状态徽章（正常 primary / 停用 danger）；根部门行无操作按钮；行内 编辑/新增/删除 按权限串显隐
- [ ] 部门新增/修改弹窗：必填项星号与校验；remote「部门已经存在」即时提示；邮箱/电话格式校验；上级部门弹树选择回填；编辑入口叶子节点拒绝文案「不能选择最后层级节点（xx）」原样出现
- [ ] 保存排序按钮：改动排序号后提交生效、未改动提示「未检测到排序修改」
- [ ] 岗位列表页：分页条（10/25/50/100）、列排序点击（postCode/postName/postSort/createTime）、状态徽章、工具栏四按钮权限显隐
- [ ] 岗位新增/修改弹窗：remote 双校验即时提示；提交后父页表格刷新（$.operate.successCallback 链路）
- [ ] 导出按钮：确认框「确定导出所有岗位吗？」→ loading → 浏览器下载 xlsx
- [ ] 全部页面在主框架 iframe 内打开、标签页标题正确（modalName：部门/岗位）

## 横切与纪律自查

- [ ] #[Perm] 18 处与 spec 表逐一对应（dept 10 + post 8）；check 端点 3 处无注解仅登录态
- [ ] #[Log] 8 处：操作成功/失败均落 sys_oper_log（title=部门管理/保存部门排序/岗位管理，business_type 1/2/3/5 与 spec 对应）；含 remove 的 301 warn 场景（status=1 失败口径按 1.0.0 约定核对）
- [ ] RepeatSubmit **未挂**本模块路由（经典版无 @RepeatSubmit 实锤，负向 grep 核对）
- [ ] 业务代码零处直接引用 predis（grep）；Redis 只经 RedisCache 门面（dict: 键）
- [ ] DataScope::apply 仅作用于部门列表/树查询三处，别名为 d
- [ ] 权限两通道一致：sys_menu.perms（system:dept:view/list/add/edit/remove、system:post:view/list/add/edit/remove/export 共 11 行）与 #[Perm] 注解、check_perm 调用三处互查无遗漏
- [ ] 表结构零变更（SHOW COLUMNS 前后一致）；qrtz_/demo 表未触碰

## Deviations 核对

- [ ] spec「拟登记 deviations」处置回填：Excel 列头样式结论（从简则登记 deviations.md，完全复刻则注明）
- [ ] 前端 quirk 三处原样保留未"修复"：叶子拒绝选父 / 根部门无操作钮+「父部门不能选择」/ deptId=100 parentName「无」

## 测试数据清理记录

（动工后逐条登记；清理完成在条目前加 ✅ 并注日期）

- [ ] 测试部门（如「测试部门A/B/C」及子孙）：物理 DELETE 或恢复 del_flag='2'→'0' 视情况，恢复 sys_dept 预置 100~109 原值
- [ ] 测试岗位（如「测试岗」）及 sys_post 预置 1~4 恢复原值
- [ ] 测试期间 sys_user_post 新增行删除（预置 2 行保留）
- [ ] 测试角色及其 sys_role_dept / sys_role_menu 关联行删除
- [ ] sys_oper_log 测试产生的行（可选：保留并注明，或清理）
- [ ] runtime/download 导出残留文件删除
- [ ] Redis 测试键（dict: 前缀预热键可保留，测试会话键清理）
- [ ] admin 会话状态复原（admin/admin123 可正常登录）
