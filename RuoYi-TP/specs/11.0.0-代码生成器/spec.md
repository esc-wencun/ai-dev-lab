# 11.0.0-代码生成器 · spec

> **状态：✅ 已完成 2026-10-01（范围拍板 A：数据层 8 路由照常 + 模板生成链整段排除，用户 2026-10-01 确认）**
> 对位经典若依 ruoyi-generator 模块：GenController（16 个方法/路由）/ GenTableServiceImpl / GenTableColumnServiceImpl / GenUtils / VelocityUtils / templates/tool/gen/ 4 页，逐方法核对实锤。
> 依赖：1.0.0（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / LoginAuth）；6.0.0（字典——列编辑页字典类型选择需 sys_dict_type 只读，若 11.0.0 先于 6.0.0 动工则该弹窗按降级方案处理，见排除范围）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下。
> **模板生成与经典版有本质差异**：经典版 Velocity 产出 Java/Vue/thymeleaf 代码；TP 版若做模板只能产出「TP 控制器/模型/模板」——产出自用。**已拍板：照 GO 版 deviations #19 先例，数据层端点照常实现、模板生成整段排除**（理由与影响面见专节），正式登记 deviations #31。
> 调研依据：reference 源码逐文件核对（GenController / GenTableServiceImpl 全量 / GenTableColumnServiceImpl / GenUtils / VelocityUtils 模板清单 / GenConfig+generator.yml / Mapper XML×2 / gen.html / importTable.html / edit.html / createTable.html）+ **ry-tp 库实测**（gen_table 21 列 / gen_table_column 22 列——**两表已随 ry_20260319.sql 导入且为空表**；sys_menu tool:gen 权限 6 行）。
>
> **动工后勘误（5 条，以经典版源码为准，见「实施记录」节）**：① gen_table_column 实为 **22 列**；② `notice_content`(longblob) 的 htmlType 实为 **summernote**（content 结尾特判覆盖四族默认）；③ list 行**含 options 键**；④ `tool:gen:list` 权限注解实为 **5 处**；⑤ deviations 编号 **#21 已被 4.0.0/5.0.0 占用**，本模块顺延 #31~#33。

## 范围拍板建议（待你确认，先给结论与理由）

**建议照 GO 版 deviations #19 先例：数据层端点照常实现（GenController 全部 16 条路由中的 8 条），模板生成类 8 条有意排除（登记 deviations）。**

理由：

1. **产出无去向**：经典版生成器的价值 = 产出「本项目技术栈的 CRUD 代码」直接投入使用。TP 版若完整复刻，模板要产出的是 **TP 控制器 + think-orm 模型 + ThinkTemplate 页面**——而这些正是本项目手写的主体（3.0.0 已产 7 页，后续模块逐个手写）。生成器产出的代码质量（固定套路模板）不会优于按 spec 手写，且本项目模块数有限（10 个业务模块写完即收官），「反复生成新模块」的场景在本项目生命周期内几乎不出现——学习价值也集中在「读表/导入/同步」的数据层，而非 Velocity→ThinkTemplate 的模板翻译。
2. **模板翻译成本高且无对齐锚点**：Velocity 模板 10+ 个（domain/controller/serviceImpl/mapper.xml/list.html/add.html/edit.html/sql…），翻译成 TP 等价物后**没有「逐字节对齐」的验收基准**（经典版模板产出 Java，TP 产出 PHP，无法对照验证正确性）——违背本项目「页面保真、行为可对齐」的验收主线。
3. **前端按钮保真的处理**：gen 列表页照常渲染（数据层端点支撑列表/导入/删除），「预览/生成代码/同步」按钮保留在页面上，点击 404/错误提示属有意行为（GO 版同款处理，deviations #19 原话「前端 gen 列表页可用，预览/生成按钮点击 404 属有意行为」）。编辑入口整页排除（其字段全部服务于模板生成参数）。
4. **与 workspace 其他版的差异说明**：Python 版 spec-10 做了全量模板生成（产 Python/Vue 代码），因为 Python 版有「接口契约」验收锚点且模块多；GO 版拍板降级。TP 版与 GO 版处境相同（无契约锚点、项目收官在望），**但 TP 版是不分离模式，连「前端 api.js+vue 生成物」这条 Python 版的产出价值都没有**——降级理由比 GO 版更充分。

**若你要求全量做**（数据层 + 模板生成），追加工作量主要在：TP 版模板组设计（建议产出 controller/model/view 三件 + 菜单 SQL，模板自写不翻译 Velocity）、preview 用 highlight 分 tab 展示、zip 打包下载（php ZipArchive）、synchDb 三方对账逻辑。届时本 spec「排除范围」节改写为「模板生成范围」节、tasks.md 追加 3~4 个 Task。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/tool/GenController.php | GenController | 8 条路由照常（index / list / db/list / column/list / importTable GET+POST / edit 保存 / remove）＋排除 8 条（见排除范围） |
| app/service/GenService.php | IGenTableService / GenTableServiceImpl（数据层部分） | 列表、查 DB 表、导入装配（GenUtils.initTable/initColumnField 等价推断）、删除（主表+列事务）、编辑保存（含 validateEdit） |
| app/model/GenTable.php + GenTableColumn.php | GenTable / GenTableColumn | think-orm 模型；GenTable 带列关联读取（对位 selectGenTableById join） |
| app/view/tool/gen/gen.html + importTable.html | templates/tool/gen/gen.html / importTable.html | 2 页照常（列表页 + 导入弹窗页） |
| （排除）preview/genCode/download/batchGenCode/createTable/synchDb/edit 编辑保存 | GenController 对应端点 | **有意排除**，登记 deviations（GO 版 #19 同款）；见排除范围节 |

¹ `GET /tool/gen/edit/{tableId}`（编辑页渲染）与 `POST /tool/gen/edit`（编辑保存）**拆开拍板**：编辑保存是纯数据层（gen_table/gen_table_column 两表更新 + validateEdit 校验），**建议照做**（导入后想改功能名/列勾选是真实使用路径）；编辑页 edit.html 本身因深度绑定生成参数（模板选择/包路径/树表字段/主子表关联）**建议排除**——即「保存端点在、编辑页不在」，前端 `$.operate.editTab` 按钮点击 404 属有意行为。**此条与 GO 版不同（GO 版 edit 保存也排除了），若你希望与 GO 版先例完全一致，把保存端点也划入排除即可——默认建议：保存端点做，理由是它不依赖任何模板设施。**

## 页面清单（2）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 代码生成列表页 | view/tool/gen/index.html | gen/gen.html | bootstrap-table（POST + server 分页，**showExport: true 前端导出**——纯前端 CSV，无后端端点，零改动保留） | `prefix = ctx + "tool/gen"`；options：url=prefix+/list、updateUrl=prefix+/edit/{id}（editTab 弹 tab 页——排除后点击 404）、removeUrl=prefix+/remove；sortName=createTime desc；**rememberSelected: true + uniqueId: "tableId"**（跨页勾选记忆）；列：checkbox / tableId（隐藏）/ 序号 / tableName（tooltip, sortable）/ tableComment（tooltip, sortable）/ className（sortable）/ createTime（sortable）/ updateTime（sortable）/ 操作列（预览 previewFlag / 编辑 editTab / 删除 / 同步 synchDb / 生成代码 genCode——按 flag 显隐）；工具栏：生成 batchGenCode（multiple, tool:gen:code）/ 创建 createTable（**shiro:hasRole="admin"**——按角色非权限）/ 导入 importTable / 修改 editTab / 删除；页内函数：preview（排除→404）、genCode（排除→404）、synchDb（排除→404）、batchGenCode（排除→404）、importTable（$.modal.open prefix+/importTable）、createTable（$.modal.open prefix+/createTable——排除→404）；highlight.min.js 引入保留（排除后无预览消费，静态资产不影响） |
| 2 | 导入表结构弹窗页 | view/tool/gen/importTable.html | gen/importTable.html | bootstrap-table（POST + server 分页，弹窗内精简配置 showSearch/Refresh/Toggle/Columns 全 false + clickToSelect + rememberSelected + uniqueId: "tableName"） | options：url=prefix+"/db/list"；列：checkbox / 序号 / tableName（tooltip）/ tableComment（tooltip）/ createTime / updateTime；submitHandler：`$.table.selectColumns("tableName")` 取勾选 → rows.join() → `$.operate.save(prefix + "/importTable", {tables: joined})`；搜索 tableName / tableComment |

> createTable.html（创建表结构页）随 createTable 端点一并排除——页面是纯 textarea 提交 `$.operate.save(prefix+"/createTable", {sql: rows})`，若日后要做模板生成再一并补。

## 端点级 API 清单（照常 8 条实注册 + 排除 8 条 = GenController 全量 16 条路由）

### 照常实现（数据层）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /tool/gen | tool:gen:view | 无 | — | 渲染页面 1 |
| 2 | POST | /tool/gen/list | tool:gen:list | 无 | pageNum/pageSize/orderByColumn/isAsc（白名单 table_name/table_comment/class_name/create_time/update_time，默认 createTime desc）+ tableName（**lower like**）/ tableComment（lower like）/ params[beginTime] / params[endTime] | TableDataInfo {code:0, rows, total}；行驼峰：tableId/tableName/tableComment/subTableName/subTableFkName/className/tplCategory/packageName/moduleName/businessName/functionName/functionAuthor/formColNum/genType/genPath/remark/createTime/updateTime（options 不出列表——经典版列表 SQL 不查 options 列） |
| 3 | POST | /tool/gen/db/list | tool:gen:list | 无 | 分页 + tableName（lower like）/ tableComment（lower like） | TableDataInfo；**information_schema.tables 查询**：`table_schema=(select database()) AND table_name NOT LIKE 'qrtz\_%' AND table_name NOT LIKE 'gen\_%' AND table_name NOT IN (select table_name from gen_table)` + 过滤 + **order by create_time desc**；行只含 tableName/tableComment/createTime/updateTime |
| 4 | POST | /tool/gen/column/list | tool:gen:list | 无 | tableId | TableDataInfo（rows/total 手工组装，无分页——经典版 columnList 直接 setRows）；gen_table_column by table_id **order by sort**；行驼峰全列（含 queryType/htmlType/dictType/isInsert/isEdit/isList/isQuery/isRequired/isPk/isIncrement/columnComment/columnType/javaType/javaField/sort/columnId/tableId） |
| 5 | GET | /tool/gen/importTable | tool:gen:list | 无 | — | 渲染页面 2 |
| 6 | POST | /tool/gen/importTable | tool:gen:list | 代码生成, 6导入 | tables（逗号串表名） | ① selectDbTableListByNames（information_schema，排除规则同 #3，**无需 NOT IN 已导入**——经典版按名取）；② 逐表：GenUtils::initTable 等价推断（className=驼峰表名/tablePrefix 处理、businessName=最后一个 _ 后段、functionName=tableComment 去「表/若依」、packageName/moduleName/functionAuthor 取配置）→ insert gen_table；③ selectDbTableColumnsByName（information_schema.columns：is_required=is_nullable='no' && column_key!='PRI'、is_pk=column_key='PRI'、sort=ordinal_position、is_increment=extra='auto_increment'、column_comment/column_type）逐列 initColumnField 等价推断 → insert gen_table_column；任一异常 → **error「导入失败：{msg}」**；成功 success()「操作成功」 |
| 7 | GET | /tool/gen/edit/{tableId} | tool:gen:edit | 无 | 路径 tableId | **默认建议：排除渲染**（404，见范围拍板）；若你选做：渲染 edit.html + 变量 table（含 columns）+ data（cxselect 其他表树 JSON——JSON.toJSON 结构） |
| 8 | POST | /tool/gen/edit | tool:gen:edit | 代码生成, 2修改 | tableId、tableName*、tableComment*、className*、functionAuthor*、functionName*、moduleName*、businessName*、packageName*、tplCategory、formColNum、genType、genPath、remark、columns[i].columnId/columnComment/javaType/javaField/sort/isInsert/isEdit/isList/isQuery/queryType/isRequired/htmlType/dictType、params[treeCode]/params[treeParentCode]/params[treeName]/params[parentMenuId]/params[parentMenuName]/params[genView] | **默认建议：照做**（纯数据层）；validateEdit：tplCategory=tree → 树三字段必填「树编码字段/树父编码字段/树名称字段不能为空」；tplCategory=sub → 「关联子表的表名/子表关联的外键名不能为空」；options=params 的 JSON 序列化落 gen_table.options；主表 update + 逐列 update（**事务**）；success()「操作成功」 |
| 9 | POST | /tool/gen/remove | tool:gen:remove | 代码生成, 3删除 | ids（逗号串 tableId） | **主表+列两表事务删**（deleteGenTableByIds + deleteGenTableColumnByIds）；success()「操作成功」（经典版 return AjaxResult.success() 非 toAjax） |

### 有意排除（模板生成类，登记 deviations；点击 404 属有意行为）

| 经典版端点 | 排除内容 | 理由（同范围拍板节） |
|---|---|---|
| GET /tool/gen/edit/{tableId} | 编辑页渲染（模板选择/包路径/树表字段/主子表关联深度绑定生成参数） | 保存端点照做（纯数据层），编辑页不渲染——$.operate.editTab 点击 404；若拍板与 GO 版完全一致，保存端点也排除（见范围拍板节括注） |
| GET /tool/gen/preview/{tableId} | Velocity 上下文组装 + 10+ 模板渲染 → {模板路径: 内容} Map | 产出 Java/Vue 代码对 TP 无用；TP 版模板无验收锚点 |
| GET /tool/gen/download/{tableName} | zip 流（Content-Disposition ruoyi.zip） | 同上 |
| GET /tool/gen/genCode/{tableName} | 自定义路径写盘（allowOverwrite 开关） | 同上 |
| GET /tool/gen/batchGenCode | 多表 zip | 同上 |
| POST /tool/gen/createTable | admin 限定 + SqlUtil.filterKeyword + Druid SQL 解析多建表语句执行 + 建后自动导入 | 依赖 Druid SQL parser（PHP 无对位）；生产建表走 SQL 导入本就是常规路径；GET 渲染页与 POST 端点一并排除 |
| GET /tool/gen/synchDb/{tableName} | information_schema 对账（新列增/消失列删/保留勾选） | 服务于「生成前列结构同步」，模板链排除后无消费方 |

统计：**照常实注册 8 条路由**（list / db list / column list / importTable GET+POST / edit 保存 / remove / index 页）；**#[Perm] 8 处**（view 1 / list 4 / edit 保存 1 / remove 1 / index 1，importTable GET+POST 同 tool:gen:list 权限）；**#[Log] 3 处**（6导入 / 2修改 / 3删除，title=代码生成）；排除侧 8 条（edit 渲染 / preview / download / genCode / batchGenCode / createTable×2〔GET 渲染+POST 保存〕/ synchDb）路由不注册即 404。**排除侧的注解形态**：createTable POST 端点带 @RequiresRoles("admin")（按角色）——随排除无需处理。

## 特殊行为清单

1. **导入的列属性推断（GenUtils.initColumnField 等价，全部照抄规则）**：
   - javaField = 列名驼峰（sys_user → sysUser——经典版 `StringUtils.toCamelCase` **不去前缀**，实锤）；
   - 类型映射：char/varchar/nvarchar/varchar2 → String；tinytext/text/mediumtext/longtext → String + **htmlType=textarea（varchar 长度 ≥500 同）**；datetime/time/date/timestamp → Date + htmlType=datetime；数值类 → 列宽 (精度,标度)：标度>0 → BigDecimal、整型宽度≤10 → Integer、否则 Long，htmlType=input；
   - isInsert 全 '1'；isEdit：非 pk 且列名不在 {id, create_by, create_time, del_flag} → '1'；isList：非 pk 且不在 {id, create_by, create_time, del_flag, update_by, update_time, remark} → '1'；isQuery：非 pk 且不在 {id, create_by, create_time, del_flag, update_by, update_time, remark}（与 NOT_LIST 同集）→ '1'；
   - queryType：默认 EQ；列名 name 结尾（忽略大小写）→ LIKE；
   - htmlType 特判：status 结尾 → radio；type/sex 结尾 → select；file 结尾 → upload；content 结尾 → summernote；
   - **blob 族无分支（经典版 quirk 照抄）**：GenUtils 类型四族不含 tinyblob/blob/mediumblob/longblob/decimal 等——longblob（如 sys_notice.notice_content，ry-tp 库实测）落入默认 String + input；decimal 落 COLUMNTYPE_NUMBER（列表含 decimal，实锤）；**其余未列类型同样默认 String + input，不做"改进"**；
   - **输出形态说明**：java_type/java_field 两列**照存 Java 语义值**（String/Long/驼峰）——DB 列即契约不改造；TP 版若日后做模板生成可再映射 PHP 类型（届时映射关系写入模板节）。
2. **表名推断（initTable 等价）**：className = 表名驼峰（autoRemovePre=false、tablePrefix=sys_ 仅在 autoRemovePre=true 时去除——经典版默认不去）；businessName = 最后一个 `_` 后段（sys_user → user）；functionName = tableComment 正则去「表|若依」；packageName/moduleName/author 来自配置（经典版 generator.yml：ruoyi / com.ruoyi.system / false / sys_ / false——**TP 版等价配置建议 config/gen.php 或沿用 env：author=ruoyi、packageName 概念对 TP 无意义但字段照存配置原值**）。
3. **admin 保护**：仅 createTable 端点 @RequiresRoles("admin")（随排除）；list/import/remove/edit 均无 admin 限制（经典版一致）；**gen 模块无数据权限**（无 @DataScope，实锤）。
4. **导入原子性**：importGenTable 单表「主表行 + 全部列行」应包事务；多表循环经典版整体一个 @Transactional，异常统一「导入失败：{msg}」——TP 照抄整包一个事务。
5. **remove 的 toAjax 语义**：经典版 `return AjaxResult.success()`——**固定成功不查行数**（删不存在 id 也「操作成功」），照抄。
6. **页面 quirk（照抄勿修）**：① gen.html `showExport: true` 是 bootstrap-table **前端导出**（浏览器本地 CSV），无后端端点——零改动保留，不算导出功能缺失；② 操作列五按钮在排除后仅「删除/导入相关」可用，预览/编辑/同步/生成代码点击 404 属有意（deviations 登记）；③ rememberSelected 跨页勾选是 bootstrap-table 前端能力，无需后端配合；④「创建」按钮 hasRole="admin"（admin 角色判断，check_perm 管不了——TP 版对位 `session.isAdmin` 模板判断）。
7. **防重复提交**：经典版 GenController 无 @RepeatSubmit（grep 实锤）→ 不挂。
8. **edit 保存的 params 序列化**：经典版 `JSON.toJSONString(genTable.getParams())`——params 是动态键 Map（treeCode/treeParentCode/treeName/parentMenuId/parentMenuName/genView）；TP 版收 `params[xxx]` 数组 → json_encode 落 options 列；selectGenTableById 回读时 json_decode 还原（setTableFromOptions 等价）。

## 「gen_table 列的驼峰输出对页面」专节（任务书点名要求）

经典版两条链路的字段形态**刻意不同**，TP 版照抄：

| 链路 | 形态 | 实锤依据 |
|---|---|---|
| **list / db/list（bootstrap-table 消费）** | 行键**驼峰**：tableId/tableName/tableComment/className/createTime/updateTime——columns field 按驼峰取值（对位 GenTable 实体 Jackson 序列化） | gen.html columns 定义 |
| **column/list（编辑页字段表消费）** | 行键**驼峰**：columnId/columnName/columnComment/columnType/javaType/javaField/isPk/isIncrement/isRequired/isInsert/isEdit/isList/isQuery/queryType/htmlType/dictType/sort——edit.html 的 `columns[i].xxx` 动态表单按驼峰取值 | edit.html bootstrap-table columns |
| **编辑页模板变量（服务端渲染）** | 下划线原值（对位 Thymeleaf th:field 直取实体属性——TP 版 volist 消费 think-orm 查询原行） | edit.html th:each="column : ${table.columns}" |
| **column/list 特例——rows/total 手工组装** | 经典版 columnList 不走 startPage，直接 new TableDataInfo setRows/setTotal(list.size)——TP 版 TableDataInfo::of($rows, count($rows))，**无分页参数消费** | GenController#columnList |

控制器统一驼峰组装输出（3.0.0 联调纪律：bootstrap-table 的 field 永远驼峰）；3.0.0「DictService 双键并存」经验在此**不需要**（gen 页面无字典消费）。

## Excel 导出列定义

**无后端导出**：GenController 无 export 端点、GenTable/GenTableColumn 无 @Excel 注解（动工检查单第 4 条核对结论）；gen.html 的 showExport 为前端本地导出。本模块零 ExcelExportService 消费。

## 关键设计说明

1. **information_schema 查询是本模块数据层核心**：db/list / importTable / synchDb（排除）三处读 information_schema.tables / .columns，SQL 逐句照抄经典版 Mapper XML（含 `NOT LIKE 'qrtz\_%' AND NOT LIKE 'gen\_%'`、`(select database())`——TP think-orm query 原生执行）；**「已导入过滤」只作用于 db/list**（导入后从可选列表消失），importTable 按名取不限。
2. **GenTable 模型带列聚合**：selectGenTableById（join gen_table_column order by sort）/ selectGenTableByName（edit 校验/子表场景用）两个聚合读取，think-orm relation 或手工两次查询组装均可（模型薄纪律：组装逻辑在 service）。
3. **权限两通道接线**：工具栏/操作列按钮 `{if check_perm('tool:gen:xxx')}`；**「创建」按钮按角色**：`{:($session.isAdmin ?? false) ? '' : 'hidden'}`（对位 shiro:hasRole="admin"；排除后按钮点击 404，显隐逻辑仍照做保真）。
4. **edit 保存若做**：columns[i] 动态表单数组的解析（TP request param('columns') 取数组）+ sort 拖拽重排（edit.html onReorderRow 写 sort 隐藏域——排除渲染则此逻辑不涉及）；validateEdit 文案四条照抄。
5. **layer 视图机制 / 驼峰行组装**：同 3.0.0 纪律（layer 相对路径、include 局部副本、控制器组驼峰行）。
6. **表前缀配置**：经典版 generator.yml 五项（author/packageName/autoRemovePre/tablePrefix/allowOverwrite）——TP 版落 `config/gen.php`（author='ruoyi'、autoRemovePre=false、tablePrefix='sys_'、allowOverwrite=false 照存；packageName='com.ruoyi.system' **字面照存**——gen_table.package_name 列即契约，值只是存进 DB 的字符串）；allowOverwrite 仅 genCode 用（排除），照存不用。

## deviations 登记（已落地，非「拟登记」）

| # | 差异点 | 落点 |
|---|--------|------|
| [#31](../deviations.md) | 代码生成器模板链整段排除（preview/download/genCode/batchGenCode/createTable/synchDb + GET edit 渲染），路由不注册即 404 | 本 spec 范围拍板节 |
| [#32](../deviations.md) | db/list 行键裁剪为四列（经典版经 Jackson 输出整个实体含 20+ 个 null 键） | 本 spec「gen_table 列的驼峰输出对页面」节 |
| [#33](../deviations.md) | remove 空 ids 视为无操作（经典版拼出 `in ()` 触发 SQL 语法错误） | 本 spec 特殊行为 5 的边界补充 |

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-10-01）

### 范围拍板结论（用户确认：选项 A）

数据层 **8 条路由**照常实注册：`GET /tool/gen`、`POST /tool/gen/list`、`POST /tool/gen/db/list`、`POST /tool/gen/column/list`、`GET+POST /tool/gen/importTable`、`POST /tool/gen/edit`（保存）、`POST /tool/gen/remove`。
排除侧 **8 条**不注册即 404：`GET edit/{tableId}`、`preview/{tableId}`、`download/{tableName}`、`genCode/{tableName}`、`batchGenCode`、`createTable`(GET+POST)、`synchDb/{tableName}`。登记 deviations #31。

### 动工后勘误（对照经典版源码逐条核实，**以源码为准**）

1. **`gen_table_column` 实为 22 列**（本 spec 原写 23 列）：`SHOW COLUMNS` 实测 22 列，无遗漏列。
2. **`notice_content`（longblob）的 htmlType 实为 `summernote`**：本 spec「Task 2」写「默认 String + input」，与同一 spec「特殊行为 1」的「content 结尾 → summernote」自相矛盾。源码 `GenUtils.initColumnField` 的 htmlType 特判是**独立于类型四族的 else-if 链**（status > type/sex > file > content），`longblob` 落四族默认后仍被 `content` 后缀覆盖 → **summernote**；java_type 才是 `String`。已按源码实现，端到端落库实测 `notice_content | String | summernote`。
3. **list 行含 `options` 键**：`GenTableMapper.xml` 的 `selectGenTableVo` 确实 `select ... gen_path, options, create_by ...`，故经典版列表 JSON 含 options。本 spec checklist 原写「无 options 键」不成立，已按源码输出并在 checklist 更正该项。
4. **`tool:gen:list` 权限注解实为 5 处**（list / db list / column list / importTable GET / importTable POST），本 spec 原写「list 4」；`#[Perm]` 总数仍为 8（view 1 + list 5 + edit 1 + remove 1），与端点清单一致。
5. **deviations 编号顺延**：本 spec 写的「#21（候选）」已被 4.0.0/5.0.0 的「角色变更后权限生效时机」占用，本模块实际登记 **#31~#33**。

### 实现要点与设计决策

- **文件落点**：`app/common/GenConstants.php`（纯常量，classmap）、`app/service/GenUtils.php`（**纯函数**推断层）、`app/service/GenService.php`（DB/配置编排）、`app/model/GenTable.php` + `GenTableColumn.php`（薄模型）、`app/controller/tool/GenController.php`、`config/gen.php`、`app/view/tool/gen/{index,importTable}.html`、`app/view/tool/include/{header,footer,bootstrap-table-export-js}.html`、`tests/unit/GenUtilsTest.php`。
- **推断层与编排层拆分**（对位经典版 `GenUtils` / `GenTableServiceImpl`，但更彻底）：`GenUtils` 不碰 Db/config，配置值由调用方传入 → PHPUnit 可 `require_once` 直接做全规则断言，不启动 TP 容器（沿用 `MenuServiceTest` 范式）。
- **插入/更新照抄 Mapper 的 `<if>` 条件**：null/空串的字段不写库，由 DB 默认值兜底（如 `table_comment=''` 落 NULL、`is_edit` 为 null 落 NULL）；`gen_table_column` 的**更新**语句无任何 `<if>`，故表单未提交的列字段一律写 null —— 端到端实测 `column 1` 的 is_edit/is_list/is_query/query_type/dict_type 被写为 NULL，与经典版一致。
- **保留经典版「主表影响行数 > 0 才更新列」语义**（`updateGenTable` 中 `if ($row > 0)`）：MySQL 无实际变化时 affected=0，列不更新。
- **不实现三个聚合读**（`selectGenTableById` / `selectGenTableByName` / `selectGenTableAll`）：仅被排除端点（预览/同步/编辑页）消费，写了即死代码；spec Task 1 括注允许「留接口不实现，注明」。
- **`information_schema` 列标签大小写**：MySQL 8 返回**大写**标签（`TABLE_NAME` / `COLUMN_TYPE`…），而 PHP 数组键大小写敏感。经典版 MyBatis `resultMap` 的 column 匹配大小写不敏感，故 TP 侧在 information_schema 三处读取统一 `array_change_key_case(CASE_LOWER)` 归一（`GenService::normalizeRow`）。**这是本次最隐蔽的坑**：不归一则 db/list 与导入推断全部拿到 null。
- **视图渲染路径**：控制器在 `tool` 层 → 模板名写 `'gen/index'`（TP 自动补层目录；写 `'tool/gen/index'` 会解析成 `app/view/tool/tool/gen/index.html` 报模板不存在），并按本项目惯例为 `tool` 层建 `include/` 局部副本（header/footer/导出片段）。
- **`{:` 标签限制**：「创建」按钮的角色显隐不能用 `{:($isAdmin ?? false) ? '' : 'hidden'}`（ThinkTemplate 不识别 `{:` 后紧跟 `(`），改用标准 `{if !($isAdmin ?? false)} hidden{/if}`。
- **`composer dump-autoload`**：`app/common/` 走 composer classmap，新增 `GenConstants.php` 后必须 `php C:\php\8.3\composer.phar dump-autoload --no-scripts`，否则运行时报 `Class "GenConstants" not found`（已实测：未 dump 时导入报「导入失败：Class "GenConstants" not found」）。
- **必填校验对位 `@Validated`**：`GenController::validateRequired` 复刻 `GenTable` 的 8 条 `@NotBlank` 文案（表名称/表描述/实体类名称/生成包路径/生成模块名/生成业务名/生成功能名/作者）+ `GenTableColumn.javaField` 的「Java属性不能为空」；`formColNum` 按 Java `int` 基本类型语义（未提交即 0）落库。
- **`config/gen.php`** 五项与经典版 `generator.yml` 原值一致（author=ruoyi / packageName=com.ruoyi.system 字面照存 / autoRemovePre=false / tablePrefix=sys_ / allowOverwrite=false 照存不用）。

### 端到端实测证据（2026-10-01，curl + 真机浏览器）

| 项 | 结果 |
|---|---|
| 空列表 / 排序参数 | `{"code":0,"rows":[],"total":0}`；点表头实测发出 `orderByColumn=tableName&isAsc=asc` |
| db/list 排除规则 | 19 张可选表；`gen_table`/`gen_table_column`/`qrtz_*` 不可见；导入 sys_notice 后降至 18 且 sys_notice 消失（已导入过滤）；`tableName=notice` lower-like 命中 `sys_notice_read` |
| 导入推断（sys_notice 全 10 列） | 落库逐列与源码推导**完全一致**：notice_id `Long/input` is_pk=1 is_increment=1；notice_title `String/input`；**notice_type `String/select`**；**notice_content `String/summernote`**；status `String/radio`；create_time `Date/datetime`；create_by/create_time 的 is_edit/is_list/is_query 全 NULL；update_by/update_time is_edit=1、is_list/is_query NULL；remark is_query NULL；query_type 全 EQ |
| 导入原子性 | 造「列注释 600 字符 > varchar(500)」表：`{"code":500,"msg":"导入失败：SQLSTATE[22001]…Data too long for column 'column_comment'"}`，**整批回滚**（含同批已插入的另一张表行），gen 两表零残留 |
| edit 保存 | `function_name` 改「公告」、form_col_num=2、remark、`options={"genView":"false"}` 落库；列勾选翻转与 dict_type 生效；未提交列字段写 NULL |
| 校验文案 7 条 | 树三字段 3 条 / 主子表 2 条 / `@NotBlank` 1 条 / 列级 javaField 1 条，文案与经典版逐字一致 |
| remove | 两表事务删（主表+列 DB 双查归零）；不存在 id 恒 `code 0`；**空 ids 无操作**（#33）；隔离性实测：两张表同存时只删目标行 |
| 排除端点 | `GET edit/{id}` / `preview/{id}` / `download/{name}` / `genCode/{name}` / `batchGenCode` / `createTable` / `synchDb/{name}` 全部 **404** |
| 权限负向（非 admin：仅 view+list） | list/column-list/importTable 放行；edit/remove 返回 `{"code":500,"msg":"您没有操作权限，请联系管理员添加权限【tool:gen:edit|remove】"}`（**反证 8 处 `#[Perm]` 串无笔误**）；页面四 flag=`hidden`、「创建」按钮 `class="btn btn-success hidden"`、导入按钮可见 |
| 浏览器级 | 列表页真机渲染（标题「代码生成列表」、1 行数据、操作列五按钮、分页文案「显示第 1 到第 1 条记录，总共 1 条记录」）；导入弹窗 10 行（`sys_job` 在列、`gen_*`/`qrtz_*`/已导入 `sys_notice` 均被排除）；分页 13 行进 2 页；**跨页勾选保持（rememberSelected 实测 true）**；导出扩展已加载（`window.tableExport` 存在 + 导出按钮 2 个，实际文件下载本连接不可捕获）；删除确认框文案「确定删除该条生成配置信息吗？」+ 取消不删数据；预览点击实测 **404** |
| 单测 | `GenUtilsTest` 16 tests / 196 assertions；全量套件 100 tests / 423 assertions 全绿 |
| 静默项 | 零 Redis 消费、零 predis 直连、本模块未挂 RepeatSubmit（负向 grep）；`#[Perm]` 8 处 / `#[Log]` 3 处（6 导入 / 2 修改 / 3 删除，title=代码生成，sys_oper_log 实测落库） |

### 遗留 / 未做（如实登记）

1. **模板生成链**（preview / download / genCode / batchGenCode / createTable / synchDb / 编辑页渲染）按范围拍板 A 未做 —— 即 deviations #31。
2. **三件套勾选范围**：checklist 中「showExport 实际下载 CSV 文件」「操作列预览/编辑/同步按钮的完整点击链路」只验到扩展加载与 404 证据层级（浏览器下载事件在本 DSH×Tabbit 连接下不可捕获）；相应子项已注明验证层级，未按「完整验证」勾选。
3. **观察记录（与 11.0.0 代码无关）**：测试期间 gen_table 中一行（sys_notice，table_id=2）在 03:31:27~03:32:11 之间被清除，而 `sys_oper_log` 无对应 DELETE 记录、被 CheckPerm 拦截的请求也不写日志；remove 的隔离性已用「两张表同存只删目标行」「空 ids 无操作」「不存在 id 忽略」三组实测反证（本模块不会误删）。当时工作区有并行会话在跑（`php think run` 属 10.0.0），**推测为并行会话或直接 SQL 清理所致，未取得确证**；已记录备查，终态 gen 两表归零符合验收要求。
4. 本模块**零 Excel 导出**（GenController 无 export 端点、GenTable/GenTableColumn 无 @Excel），`runtime/download` 核对为空。
