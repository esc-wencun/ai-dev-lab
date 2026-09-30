# 11.0.0-代码生成器 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 前置：**范围拍板已确认（2026-10-01，用户选 A）**——数据层 8 端点照常 + 模板生成链整段排除（deviations #31）。
> 顺序：模型/service（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。**动工后勘误 5 条见 spec「实施记录」**（gen_table_column 22 列 / longblob+content→summernote / list 含 options / list 权限 5 处 / deviations 编号顺延）。

## Task 1 · 模型与 GenService（数据层）

- [x] app/model/GenTable.php（表 gen_table，主键 table_id；autoWriteTimestamp 关闭）+ GenTableColumn.php（表 gen_table_column）
- [x] config/gen.php：author/packageName/autoRemovePre/tablePrefix/allowOverwrite 五项照存经典版 generator.yml 原值
- [x] `initTable` 等价（落 `GenUtils::initTable`，**纯函数 + 配置入参**）：className=表名驼峰（不去前缀）、businessName=最后 _ 后段、functionName=tableComment 正则去「表|若依」、packageName/moduleName/author 取配置、createBy=操作人
- [x] `initColumnField` 等价（spec 特殊行为 1 全规则）：javaField 驼峰 / 类型映射（str/text/time/number 四族 + varchar≥500→textarea + 标度>0→BigDecimal、宽度≤10→Integer）/ isInsert/isEdit/isList/isQuery 勾选规则 / queryType（EQ 默认 + name 结尾 LIKE）/ htmlType 特判（status→radio、type|sex→select、file→upload、content→summernote）
- [x] selectDbTableList（information_schema.tables：排除 qrtz_/gen_/已导入 + lower like + order by create_time desc）与 selectDbTableListByNames（按名取，不含已导入过滤）
- [x] selectDbTableColumnsByName（information_schema.columns：is_required/is_pk/sort/is_increment 推断 SQL 照抄）
- [x] **information_schema 列标签归一**（MySQL 8 返回大写 `TABLE_NAME`/`COLUMN_TYPE`，PHP 键大小写敏感；`GenService::normalizeRow` 统一小写化，对位 MyBatis resultMap 大小写不敏感匹配）——**动工实测发现的坑，不做则 db/list 与推断全为 null**
- [x] PHPUnit：initTable/initColumnField 纯函数全规则断言（`tests/unit/GenUtilsTest.php`，**16 tests / 196 assertions**；夹具 = sys_notice 实测形态 10 列 + 合成类型覆盖 int(1)/int(11)/decimal(10,2)/varchar(499/500)/text 族/blob 族）
- [x] `selectGenTableById` / `selectGenTableByName` / `selectGenTableAll` **不实现**（注明原因）：三者仅被**已排除端点**（预览/同步/编辑页渲染）消费，实现即死代码；spec 原文括注「方法可留接口不实现，注明」允许此处理

## Task 2 · 控制器与路由（curl 可验）

- [x] GenController 8 方法：index / list / dbList / columnList / importTable(GET 渲染) / importTableSave / editSave / remove；**#[Perm] 8 处**（view 1 / **list 5** / edit 1 / remove 1——勘误 4）+ **#[Log] 3 处**（6 导入/2 修改/3 删除，title=代码生成）
- [x] importTableSave：tables 逗号串 → 按名查 information_schema → **整包事务**（逐表主表+列写入）→ 异常 error「导入失败：{msg}」；editSave：`@NotBlank` 8 条 + 列级 javaField 文案 + validateEdit 四文案 + options=params JSON 序列化 + 主表列逐列 update 事务（**保留「主表影响行数>0 才更新列」经典语义**）；remove：**主表+列两表事务删、固定 success()**
- [x] columnList：TableDataInfo::of($rows, count($rows)) 手工组装（无分页）；order by sort
- [x] 路由注册 tool/gen 组 **8 条**（get '/'、post list、post db/list、post column/list、get+post importTable、post edit、post remove；**preview/download/genCode/batchGenCode/createTable/synchDb/edit GET 不注册**——404 即排除形态，7 条 URL 实测 404）
- [x] curl 级自测（UTF-8 脚本）：list 空表 `{code:0,rows:[],total:0}`、db/list 出 19 张可选表（无 qrtz_/gen_/重复）、importTable 导入 sys_notice（**10 列推断落库逐列断言全通过**：notice_id is_pk/is_increment='1'、notice_title→String input、notice_type char(1)→String + type 结尾→select、**notice_content longblob → String + summernote**（勘误 2）、status→radio、create_by/create_time/update_by/update_time/remark 勾选规则）、column/list 驼峰行无分页（total=10）、editSave 改 functionName + 勾选翻转落库 + 未提交列字段写 NULL、remove 两表行消失、remove 不存在 id 恒 code 0
- [x] **导入原子性实测**：造列注释 600 字符（> gen_table_column.column_comment varchar(500)）→ `{"code":500,"msg":"导入失败：SQLSTATE[22001]…"}`，同批另一张已插入的表**一并回滚**（gen 两表零残留）
- [x] 权限负向实测（临时非 admin 用户，仅 view+list）：list/column-list/importTable 放行；edit/remove 返回「您没有操作权限，请联系管理员添加权限【tool:gen:edit|remove】」——**反证 8 处权限串无笔误**

## Task 3 · 页面模板（2 页）（浏览器可验）

- [x] view/tool/gen/index.html（列表页：showExport 前端导出、rememberSelected 跨页勾选、操作列五按钮按 flag 显隐、工具栏「创建」按钮 isAdmin 判断）
- [x] view/tool/gen/importTable.html（弹窗列表：精简配置、submitHandler selectColumns → tables=join 提交）
- [x] `app/view/tool/include/` 三件局部副本（header / footer / bootstrap-table-export-js）——**本项目 include 片段按「层目录」解析**（对位 monitor/system 层惯例）
- [x] 静态自查：无 `th:*`/`@{` 残留（`\{:`/`\{if`/`\{\$` 计数 0）；check_perm `{:}` 输出正常；volist/if 配平；键名驼峰
- [x] 3.0.0 联调经验带入：渲染路径 **layer 相对（`'gen/index'`，写 `'tool/gen/index'` 会解析成 `app/view/tool/tool/gen/index.html`）**；list/db list 输出驼峰行
- [x] 模板限制实测：`{:` 后紧跟 `(` 不被 ThinkTemplate 识别 → 「创建」按钮改用 `{if !($isAdmin ?? false)} hidden{/if}`
- [x] 浏览器自查：列表页空态/有数据态渲染；导入弹窗开合；**勾选跨页保持（rememberSelected 实测 true）**；提交后父页刷新出新行

## Task 4 · 端到端验收（浏览器级）

- [x] 主框架「代码生成」标签打开：侧栏 系统工具 → 代码生成 → iframe `/tool/gen` 打开、活动标签「代码生成」、空列表 + 工具栏按钮（admin 全显：生成/创建/导入/修改/删除）
- [x] 导入链路：导入按钮 → 弹窗表列表（**gen_/qrtz_ 不可见、已导入 sys_notice 不可见**）→ 搜索过滤 → 勾选 sys_config → 确定 → 列表新行（表名/描述/实体类名/时间）
- [x] 多表导入与重复导入行为：一次导入 12 张表成功（13 行/146 列）；**经典版无查重——重复导入会插重复行**（原子性批次实测插入重复 tmp_gen_ok 行）；db/list 已导入表消失（19→18→…）
- [x] 操作列「删除」→ 确认框（文案「确定删除该条生成配置信息吗？」）→ 确认 → 两表行消失（浏览器实测回到 0 行 + DB 双查归零）；**取消**分支实测不删数据
- [x] 排除项负向确认：预览/编辑/同步/生成代码/批量生成/创建 点击 **404**（curl 7 条 URL 全 404 + 浏览器实测预览 404 且 loading 遮罩不消失——经典 JS 无 error 分支，属有意行为）；「创建」按钮非 admin 会话实测 `class="btn btn-success hidden"`
- [x] gen 表数据零 schema 变更（gen_table 21 列 / gen_table_column 22 列前后一致）；测后清空 gen_table/gen_table_column（终态 COUNT=0）

## Task 5 · 收尾

- [x] PHPUnit 全绿（全量套件 **100 tests / 423 assertions**）；spec.md 实施记录回填（含范围拍板结论与 5 条勘误）
- [x] deviations **#31/#32/#33** 正式登记（模板链排除 / db list 行键裁剪 / remove 空 ids 边界）——注意 spec 原写的「#21」已被 4.0.0 占用，已顺延
- [x] 测试数据清理：gen_table/gen_table_column 清空；临时表 tmp_gen_ok / tmp_gen_atomicity 已 DROP；临时角色 gen_test + 用户 gen_tester + 关联表行全删；Redis 测试会话删除；admin 会话复原（admin/admin123 冷登录实测通过）；runtime/download 无残留；sys_job/sys_job_log 未触碰
- [x] specs/README.md 总表 11.0.0 状态与「内容」列更新为拍板后实际范围
