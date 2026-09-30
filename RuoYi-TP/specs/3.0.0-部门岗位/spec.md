# 3.0.0-部门岗位 · spec

> **状态：✅ 已完成（2026-09-29，浏览器级端到端通过；权限隔离验证遗留至 5.0.0）**
> 对位经典若依 SysDeptController / SysPostController + SysDeptServiceImpl / SysPostServiceImpl + templates/system/dept|post（部门 11 端点 / 岗位 10 端点 / 页面 7 个，逐方法核对实锤）。
> 依赖：1.0.0（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / DataScope / LoginAuth）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下（按序号串行隐式满足）。
> 调研依据：reference 源码逐文件核对（Controller×2 / ServiceImpl×2 / Mapper XML×2 / 页面 HTML×7 / SysDept、SysPost 实体 / ry-ui.js $.operate·$.treeTable·$.tree·$.table.exportExcel）+ **ry-tp 库实测**（sys_dept 14 列 / sys_post 10 列 / sys_role_dept / sys_user_post；预置部门 10 行、岗位 4 行、用户岗位关联 2 行；sys_menu perms 11 行）。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 体系 0/301/500、POST 分页参数、信封结构、权限两通道（#[Perm] 注解 + check_perm 模板函数）。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/DeptController.php | SysDeptController | 11 个方法（selectDeptTree 双路由 pattern，共 12 条路由） |
| app/controller/PostController.php | SysPostController | 10 个方法 |
| app/service/DeptService.php | ISysDeptService / SysDeptServiceImpl | 树查询（DataScope）、ancestors 级联、停用级联、唯一校验、软删、排序 |
| app/service/PostService.php | ISysPostService / SysPostServiceImpl | 分页列表、双唯一校验、占用删除（物理删） |
| app/service/DictService.php（**最小只读**） | SysDictTypeServiceImpl#selectDictDataByType | 只做「按类型取字典数据」；字典管理界面与缓存维护 6.0.0 落地时收编 |
| app/service/ExcelExportService.php + composer 引入 **phpoffice/phpspreadsheet** | ExcelUtil（本模块只消费 SysPost 的 5 个 @Excel 列） | 列定义数组驱动的通用导出工具首落，7.0.0 公告等后续模块复用；php.ini 需 intl/gd/zip（已在 tech-stack 第六节启用清单内） |
| GET /common/download 端点 | CommonController.fileDownload | **通用下载端点本模块落地**（export 响应文件名 → 前端 GET download），7.0.0 起复用 |
| app/view/system/dept/*.html ×4、app/view/system/post/*.html ×3 | templates/system/dept/、templates/system/post/ | 7 页，静态 JS（ry-ui.js / bootstrap-tree-table / zTree / jquery.validate）零改动对接 |

**范围外**（后续模块，勿在本模块实现）：roleDeptTreeData / selectRoleDeptTree（5.0.0 角色数据权限树）、selectPostAll / selectPostsByUserId 与 sys_user_post 写维护（4.0.0 用户模块）、字典管理页面与缓存失效联动（6.0.0）。

## 页面清单（7）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/树组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 部门列表页 | view/system/dept/index.html | dept/dept.html | **bootstrap-tree-table 树表**（父子展开，非分页） | `prefix = ctx + "system/dept"`；options：url=prefix+/list（POST，**返回裸数组**）、createUrl=prefix+/add/{id}、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove/{id}；保存排序 `$.operate.post(prefix+"/updateSort", {deptIds:"…,…", orderNums:"…,…"})`（仅提交改动行，originalOrders 对比）；搜索字段 deptName / status；操作列 **parentId!=0 才渲染**（根部门行无操作按钮）；字典徽章 datas=sys_normal_disable |
| 2 | 部门新增弹窗 | view/system/dept/add.html | dept/add.html | 表单 | validate remote：POST prefix+/checkDeptNameUnique（data: parentId+deptName，失败文案「部门已经存在」）；orderNum digits / email / phone(isPhone)；提交 `$.operate.save(prefix+"/add", serialize)`；上级部门输入框点击 → `$.modal.openOptions` 打开 prefix+/selectDeptTree/{treeId}/0（excludeId=0 不排除） |
| 3 | 部门修改弹窗 | view/system/dept/edit.html | dept/edit.html | 表单 | validate remote 同上（多带 deptId）；提交 save(prefix+"/edit")；选上级部门 → prefix+/selectDeptTree/{treeId}/{excludeId=当前deptId}（**排除自身及子树**）；doSubmit 经 `$.tree.notAllowLastLevel` 校验——**选中叶子节点报「不能选择最后层级节点（xx）」拒绝回填（经典版原样 quirk，保留勿"修复"）**；parentId=0 时点击提示「父部门不能选择」 |
| 4 | 部门树选择弹窗 | view/system/dept/tree.html | dept/tree.html | **zTree** | 隐藏域 treeId/treeName 初始值 = 模板变量 dept；`url = prefix + "/treeData/" + excludeId`（GET，返回 Ztree 裸数组）；expandLevel=2；搜索/展开/折叠为纯前端 |
| 5 | 岗位列表页 | view/system/post/index.html | post/post.html | bootstrap-table（**POST + server 分页**） | `prefix = ctx + "system/post"`；options：url=prefix+/list、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove（**无 {id}**，ids 走 body）、exportUrl=prefix+/export；sortName=postSort（初始请求 orderByColumn=postSort&isAsc=asc）；sortable 列 postCode/postName/postSort/createTime；工具栏：新增/修改(single)/删除(multiple)/导出 |
| 6 | 岗位新增弹窗 | view/system/post/add.html | post/add.html | 表单 | validate remote：POST ctx+"system/post/checkPostNameUnique"（仅 postName）、checkPostCodeUnique（仅 postCode），文案「岗位名称已经存在」「岗位编码已经存在」；postSort digits；提交 save(prefix+"/add") |
| 7 | 岗位修改弹窗 | view/system/post/edit.html | post/edit.html | 表单 | 同上，remote data 多带 postId；提交 save(prefix+"/edit") |

> 术语说明：部门**列表页**用的是 bootstrap-tree-table 树表插件；zTree 插件仅用于「部门树选择弹窗」（页面 4）。TP 模板命名统一 index/add/edit/tree（控制器显式指定渲染文件），对位关系见上表。

## 端点级 API 清单（dept 11 + post 10 + 通用 download 1）

### 部门 /system/dept（对位 SysDeptController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /system/dept | system:dept:view | 无 | — | 渲染页面 1 |
| 2 | POST | /system/dept/list | system:dept:list | 无 | deptName（like）/ status；支持 deptId/parentId 精确（内部调用）；**无分页参数** | **裸 JSON 数组**（List&lt;SysDept&gt; 直出，非 TableDataInfo！）；固定 order by parent_id, order_num；行字段驼峰：deptId/parentId/ancestors/deptName/orderNum/leader/phone/email/status/delFlag/createBy/createTime/updateBy/updateTime（delFlag 恒 '0'，已过滤） |
| 3 | GET | /system/dept/add/{parentId} | system:dept:add | 无 | 路径 parentId（工具栏固定 100，行内按钮传行 deptId） | 渲染页面 2，模板变量 dept=父部门记录；**非 admin 强制 parentId = 会话用户 deptId**（isAdmin 判断） |
| 4 | POST | /system/dept/add | system:dept:add | 部门管理, 1新增 | parentId、deptName*、orderNum*、leader、phone、email、status；**表单不传 ancestors**（service 重算） | 唯一冲突 → error(500)「新增部门'{deptName}'失败，部门名称已存在」；父部门停用 → error(500)「部门停用，不允许新增」（ServiceException）；成功/失败 toAjax：{code:0,msg:操作成功} / {code:500,msg:操作失败}；createBy=登录名 |
| 5 | GET | /system/dept/edit/{deptId} | system:dept:edit | 无 | 路径 deptId；**先 checkDeptDataScope** | 渲染页面 3，变量 dept（含 parentName 子查询）；**deptId=100 时 parentName 置「无」**（经典版硬编码 100，原样复刻） |
| 6 | POST | /system/dept/edit | system:dept:edit | 部门管理, 2修改 | deptId、parentId、deptName*、orderNum*、leader、phone、email、status | 校验顺序（经典版实锤）：checkDeptDataScope → 唯一「修改部门'{deptName}'失败，部门名称已存在」(500) → parentId==deptId「修改部门'{deptName}'失败，上级部门不能是自己」(500) → 改停用且存在未停用子部门「该部门包含未停用的子部门！」(500)；ancestors 级联见特殊行为；updateBy=登录名；toAjax |
| 7 | POST | /system/dept/updateSort | system:dept:edit | 保存部门排序, 2修改 | deptIds、orderNums（前端 join(",") 的同序两串；经典版 Spring 拆 String[]，TP 版 explode 等价） | 逐行 update order_num（**事务**）；异常 → error(500)「保存排序异常，请联系管理员」；成功固定 success()「操作成功」（非 toAjax，0 行也算成功） |
| 8 | POST | /system/dept/remove/{deptId} | system:dept:remove | 部门管理, 3删除 | 路径 deptId（treeTable 分支 $.operate.post(url)，**无 body**） | 校验顺序：子部门 count>0 → **warn(301)**「存在下级部门,不允许删除」；部门下有未删用户 → **warn(301)**「部门存在用户,不允许删除」；checkDeptDataScope（无权 → error(500)「没有权限访问部门数据！」）；**软删** del_flag='2'；toAjax |
| 9 | POST | /system/dept/checkDeptNameUnique | **无（仅登录态）** | 无 | deptName、parentId、（编辑时多传 deptId） | **裸 boolean**：true=唯一 / false=重复（jquery validate remote 直接消费）；唯一性口径 = **同 parentId 下**同名且 del_flag='0' limit 1；deptId 相同视为自身放行 |
| 10 | GET | /system/dept/selectDeptTree/{deptId}[/{excludeId}] | system:dept:list | 无 | 路径 deptId + 可选 excludeId（页面 2 传 0，页面 3 传当前 deptId） | 渲染页面 4，变量 dept / excludeId（经典版双路由 pattern 合并为一条可选参数路由） |
| 11 | GET | /system/dept/treeData/{excludeId} | system:dept:list | 无 | 路径 excludeId（0=不排除） | **Ztree 裸数组**：[{id, pId, name, title, checked:false, open:false, nocheck:false}]（pId 键名精确）；**只含 status='0'** 的部门；excludeId>0 时排除自身及 ancestors 含 excludeId 的后代；**带数据权限过滤**（selectDeptTreeExcludeChild 有 @DataScope） |

### 岗位 /system/post（对位 SysPostController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 12 | GET | /system/post | system:post:view | 无 | — | 渲染页面 5 |
| 13 | POST | /system/post/list | system:post:list | 无 | pageNum/pageSize/orderByColumn/isAsc（PageQuery，白名单 post_code/post_name/post_sort/create_time）+ postCode（like）/ postName（like）/ status | **TableDataInfo** {code:0, msg:"查询成功", rows:[…], total:n}；排序由 orderBy 拼 SQL（对位 PageHelper OrderBy；mapper 本身无 order by） |
| 14 | POST | /system/post/export | system:post:export | 岗位管理, 5导出 | 同搜索参数 + orderByColumn/isAsc；**无分页（导全量）** | {code:0, msg:"&lt;uuid&gt;_岗位数据.xlsx"}——**文件名装在 msg 里**；前端随后 GET /common/download?fileName=…&delete=true；Excel 列定义见下节 |
| 15 | POST | /system/post/remove | system:post:remove | 岗位管理, 3删除 | ids（逗号串；单行删除同样走 ids body——bootstrapTable 分支与 dept 的路径参数不同） | 已分配（sys_user_post count>0）→ **error(500)**「{postName}已分配,不能删除」（ServiceException，非 warn）；**物理删除** delete from sys_post；toAjax |
| 16 | GET | /system/post/add | system:post:add | 无 | — | 渲染页面 6 |
| 17 | POST | /system/post/add | system:post:add | 岗位管理, 1新增 | postName*、postCode*、postSort*、status、remark | 名称唯一「新增岗位'{postName}'失败，岗位名称已存在」(500) → 编码唯一「新增岗位'{postName}'失败，岗位编码已存在」(500)；createBy；toAjax |
| 18 | GET | /system/post/edit/{postId} | system:post:edit | 无 | 路径 postId | 渲染页面 7，变量 post（无数据权限校验——经典版同样没有） |
| 19 | POST | /system/post/edit | system:post:edit | 岗位管理, 2修改 | postId、postName*、postCode*、postSort*、status、remark | 双唯一校验（文案前缀「修改岗位…」）；updateBy；toAjax |
| 20 | POST | /system/post/checkPostNameUnique | **无（仅登录态）** | 无 | postName（+编辑时 postId） | 裸 boolean；口径 = post_name **全局**唯一（无 parent 维度）limit 1 |
| 21 | POST | /system/post/checkPostCodeUnique | **无（仅登录态）** | 无 | postCode（+postId） | 裸 boolean；post_code 全局唯一 |
| 22 | GET | /common/download | **无（随静态资源放行同等待遇）** | 无 | fileName、delete | 通用下载：runtime/download 目录读文件流（application/octet-stream + attachment 文件名 URL 编码）；realFileName = timestamp + fileName 第一个 "_" 之后部分（&lt;uuid&gt;_岗位数据.xlsx → &lt;ts&gt;岗位数据.xlsx）；delete=true 下载后删除；防目录穿越（禁 ".."）+ 扩展名白名单（至少含 xlsx） |

统计：**控制器方法 21 个**（dept 11 + post 10）+ 通用下载 1；**URL pattern 共 23 条**（selectDeptTree 双 pattern 计 2，上表 22 行中该行覆盖 2 条）；**#[Perm] 共 18 处**（dept 10 + post 8；#9/#20/#21 三个 check 端点无权限注解，仅登录态）；**#[Log] 共 8 处**（部门 4：INSERT/UPDATE×2/DELETE，岗位 4：EXPORT/DELETE/INSERT/UPDATE）。

## 特殊行为清单（唯一性文案 / admin 保护 / 级联 / 关联 / 排序）

1. **唯一性校验文案全集**（前后端双保险：validate remote 提前提示 + 提交后端兜底，文案不同属经典版原样）：
   - 后端 add/edit：`新增部门'{deptName}'失败，部门名称已存在` / `修改部门'{deptName}'失败，部门名称已存在`；`新增岗位'{postName}'失败，岗位名称已存在` / `修改岗位'{postName}'失败，岗位名称已存在`；`新增岗位'{postName}'失败，岗位编码已存在` / `修改岗位'{postName}'失败，岗位编码已存在`
   - 前端 remote 提示：部门「部门已经存在」；岗位「岗位名称已经存在」「岗位编码已经存在」
   - 口径差异：部门唯一 = **同 parentId 下** dept_name 唯一（不同父部门可同名）；岗位名称/编码 = **全局**唯一。均只查 del_flag='0'（部门）/全表（岗位）。
2. **admin 保护（仅部门）**：① GET add 非 admin 强制 parentId = 本人 deptId；② checkDeptDataScope（edit GET / edit POST / remove POST）非 admin 访问数据权限外部门 → 「没有权限访问部门数据！」(500)。岗位模块**无任何 admin 保护**（经典版同样没有）。根部门无显式后端保护——其行内按钮前端不渲染（parentId!=0 条件）+ 默认数据有子部门删不掉，行为与经典版一致。
3. **部门 ancestors 级联**（updateDept，事务）：newAncestors = newParent.ancestors + "," + newParent.deptId；自身更新后，全部子孙（find_in_set(deptId, ancestors)）的 ancestors 按 replaceFirst(oldAncestors, newAncestors) 重算。TP 实现用**前缀锚定**替换（`preg_replace('/^' . preg_quote($old, '/') . '/', $new, …)`）——合法数据中子链必以 oldAncestors 开头，行为等价，且规避经典版 replaceFirst 无边界的数字误匹配隐患（如 "0,10" 误匹配 "0,101"），此为实现加固非行为差异。**启用级联**：改启用状态（status='0' 且 ancestors 非空且 ancestors≠"0"）时自动把 ancestors 链上全部部门 status 置 '0'。
4. **停用/新增的父子状态约束**：insertDept 父部门非正常状态 → 「部门停用，不允许新增」；editSave 改停用且 find_in_set 子部门中存在 status='0' → 「该部门包含未停用的子部门！」。
5. **删除语义**：部门软删（del_flag='2'，重复删已删部门 → toAjax 0 行 → 「操作失败」）；岗位物理删。岗位删除前查 sys_user_post 占用（本模块只读该表，写入归 4.0.0）。
6. **排序字段**：部门列表固定 order by parent_id, order_num（无前端排序）；岗位列表排序由 orderByColumn 驱动（默认 postSort asc）。部门另有 /updateSort 批量排序端点（仅提交改动行）。
7. **后端参数校验文案**（对位 @Validated + BindException → error 首条消息）：部门名称「部门名称不能为空」「部门名称长度不能超过30个字符」（DB varchar(30) 同口径）；显示顺序「显示顺序不能为空」；联系电话「联系电话长度不能超过11个字符」；邮箱「邮箱格式不正确」「邮箱长度不能超过50个字符」；岗位「岗位编码不能为空/长度不能超过64个字符」「岗位名称不能为空/长度不能超过50个字符」「显示顺序不能为空」。
8. **防重复提交**：经典版拦截器虽注册 /** 但仅对 @RepeatSubmit 注解方法生效，dept/post 控制器**零处注解**（grep 实锤）→ 本模块不挂 RepeatSubmit 中间件。
9. **前端 quirk（原样保留，勿修）**：① 部门编辑弹窗选父部门时叶子节点被拒（「不能选择最后层级节点（xx）」）；② 根部门（parentId=0）行无操作按钮、点上级部门输入框提示「父部门不能选择」；③ deptId=100 的 parentName 显示「无」为硬编码。

## 数据权限声明

- **部门三个查询方法带 @DataScope(deptAlias="d")**（selectDeptList / selectDeptTree / selectDeptTreeExcludeChild 实锤）：列表、树数据均按 sys_role.data_scope 过滤（admin 不过滤）；对位 `DataScope::apply($query, $user, deptAlias: 'd')`（1.0.0 已落封装），条件作用于 d.dept_id，别名为 sys_dept 查询的 `d`。
- **checkDeptDataScope**：edit/remove 入口的横向越权防护（非 admin 对数据范围外 deptId 抛异常）。
- **岗位模块无数据权限**（SysPostServiceImpl 无 @DataScope，实锤）。
- selectDeptCount / checkDeptExistUser / checkDeptNameUnique 不带数据权限（经典版一致）。

## Excel 导出列定义（SysPost @Excel 逐字段抄录，共 5 列）

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | postId | name="岗位序号", cellType=NUMERIC | 列头「岗位序号」，数字格式 |
| 2 | postCode | name="岗位编码" | 列头「岗位编码」，文本 |
| 3 | postName | name="岗位名称" | 列头「岗位名称」，文本 |
| 4 | postSort | name="岗位排序", cellType=NUMERIC | 列头「岗位排序」，数字格式 |
| 5 | status | name="状态", readConverterExp="0=正常,1=停用" | 列头「状态」，按表达式转换后输出文本 |

- BaseEntity 的 createBy/createTime/updateBy/updateTime/remark **无 @Excel 注解 → 不导出**；sheet 名「岗位数据」；文件名 `&lt;uuid&gt;_岗位数据.xlsx`。
- **部门无导出**：SysDept 无 @Excel 注解、SysDeptController 无 export 端点（动工检查单第 4 条核对结论）。

## 关键设计说明

1. **dept /list 返回裸数组**是本模块最大响应格式特例：bootstrap-tree-table 的 responseHandler 对 `res.code == undefined` 放行；TP 版返回 `json($rows)`，不得套 TableDataInfo 信封。行时间字段 `Y-m-d H:i:s`（对位 @JsonFormat）。
2. **check 三端点返回裸 boolean**（JSON true/false），jquery validate remote 以响应值判定校验成败；文案由前端 messages 提供。
3. **权限两通道接线**：按钮 `shiro:hasPermission` → `{if check_perm('system:dept:add')}`；页面 JS 变量 `var addFlag = '[[${@permission.hasPermi(...)}]]'`（有=""无="hidden"）→ TP 模板输出 `{:check_perm('system:dept:add') ? '' : 'hidden'}`；字典数组 `var datas = {:json(DictService::listByType('sys_normal_disable'))}`（行内须含 dictLabel/dictValue/listClass/cssClass 供 selectDictLabel 徽章渲染；正常=primary、停用=danger）。
4. **DictService 最小只读**：`listByType(type)` 读 sys_dict_data（status='0'，order by dict_sort），Redis 缓存键 `dict:<type>`（TP 自定键前缀，tech-stack 已授权）；6.0.0 字典管理落地时收编并接管缓存失效。经典版 sys_normal_disable 预置：正常/0（默认）/停用/1。
5. **/common/download**：经典版 export 链路 = 生成文件到 download 目录 → AjaxResult.success(文件名) → 前端 GET download 读流后删除。TP 版同机制，runtime/download 目录；checkAllowDownload 防 ".." 穿越 + 扩展名白名单。
6. **PageQuery 白名单**：post 列表 sortable 列 postCode/postName/postSort/createTime → post_code/post_name/post_sort/create_time；orderByColumn 传驼峰由 PageQuery::camelToSnake 转换（1.0.0 已实现）。
7. **事务边界**：updateDept（自身+子孙 ancestors 级联 + 启用上级链）、updateDeptSort（逐行排序）必须包显式事务（对位 @Transactional）。
8. **getSysUser/isAdmin**：会话内容（2.0.0 已含 userId/loginName/deptId/permissions/roles/isAdmin）直接读取，不新增会话字段。

## 拟登记 deviations

经逐点核对（响应格式 / 权限串 / 校验文案 / 级联语义 / 删除语义 / 导出链路），**无行为级差异需登记**——复刻点均可与经典版行为一致。候补一条，实施时定：

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| Excel 导出内部样式 | POI ExcelUtil 列头灰底居中、字体等固定样式 | phpspreadsheet 尽量复刻同款列头样式；若对齐成本过高则从简 | 实施时定，若从简登记 deviations（「导出 Excel 列头样式简化」） |

另注（非 deviation，仅为防误改说明）：部门编辑弹窗「叶子节点不能选为父部门」、根部门 parentName 硬编码「无」、check 端点无权限注解，均为经典版原样行为，**照抄不修正**。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录

- 2026-09-29 全部 9 个 Task 完成（PHPUnit 32 tests 97 assertions；curl/DB 级 + 浏览器级端到端通过）。
- **浏览器级实测**：部门树表 10 行层级正确渲染；新增/编辑/删除全链路（弹窗回显、ancestors 级联、软删 del_flag='2'）；岗位列表分页 + 新增 + 导出→下载（浏览器实际下载 1790668379岗位数据.xlsx + 服务端源文件删除）。
- **联调返工记录**（对后续业务模块有直接参考价值）：
  1. **视图 layer 机制**：`system.DeptController` → 视图根自动变 `app/view/system/`——控制器渲染路径用 layer 相对路径（'dept/index'），include 片段需在 layer 内有副本（app/view/system/include/ 从全局复制）。
  2. **API 输出键名必须驼峰**：bootstrap-table/tree-table 的 columns field（deptName/postId 等）与经典版实体 Jackson 序列化一致用驼峰；控制器组响应用驼峰键。DictService 双键并存（JS 消费驼峰 / radio volist 消费下划线）。
  3. **命名空间 use 第三次踩坑**（DictService 缺 TpConstant/RedisCache、PostService 类型声明、CommonController 缺 ExcelExportService）——全局类在 namespaced 文件必须 use。
  4. `paginate()->items()` 返回 array（非 Collection）；`Db::getLastInsID()` 需 PDO 参数（改 max(pk)）。
- **遗留**：权限两通道的隔离验证（测试角色 + 非 admin 用户）→ 5.0.0 角色模块落地后补验（CheckPerm/check_perm 机制本身已在 1.0.0 验证）。
- deviations #19（Excel 列头样式从简）已登记。
