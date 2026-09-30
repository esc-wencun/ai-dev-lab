# 11.0.0-代码生成器 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（照常实注册 8 条；排除 8 条不注册即 404）。
> 前置：**范围拍板已确认（2026-10-01，用户选 A：数据层照常 + 模板生成排除）**。
> 验证方式：curl 级（管理员会话 + 非 admin 负向会话）+ Tabbit 真机浏览器（列表页 / 导入弹窗 / 主框架菜单 / 提交删除闭环）。

## 数据层（curl / DB 级）

- [x] GET /tool/gen 渲染列表页（tool:gen:view；未登录 302 /login）
- [x] POST /tool/gen/list：TableDataInfo {code:0,rows,total}；分页；tableName/tableComment lower like；params[beginTime]/[endTime] 过滤；行键驼峰
  - [x] **行含 `options` 键**（勘误 3：selectGenTableVo 确实 select options，原写「无 options 键」不成立，按源码输出）
- [x] POST /tool/gen/db/list：information_schema 查询——**qrtz_/gen_ 前缀不可见、gen_table 已导入表不可见**；order by create_time desc；分页与 like 过滤生效
  - [x] 行键裁剪为四列（deviations #32；页面只消费这四列）
- [x] POST /tool/gen/column/list：按 tableId 出列（驼峰行、order by sort、rows/total 手工组装无分页）
- [x] GET+POST /tool/gen/importTable：同权限 tool:gen:list；导入 sys_notice 全列推断核对——java_field 驼峰、类型映射（varchar→String、datetime→Date+datetime、int→Long…）、is_pk/is_increment/is_required SQL 推断、勾选规则（create_by/create_time/update_by/update_time/remark 不在 is_list/is_query；id 不在 is_edit）、queryType（name 结尾 LIKE 其余 EQ）、htmlType 特判（status→radio / type→select / **content→summernote（勘误 2）**）；**functionName 去「表/若依」**
- [x] 导入原子性：人为造异常（列注释 600 字符 > varchar(500)）→ 整批回滚 → error「导入失败：{msg}」（同批已插入的另一张表一并回滚，DB 零残留）
- [x] POST /tool/gen/edit：validateEdit 四文案（tplCategory=tree 缺树字段、=sub 缺子表字段——curl 直发构造）+ `@NotBlank` 8 条 + 列级「Java属性不能为空」（共 7 条实测通过）；options JSON 落库；columns[i] 逐列 update（未提交字段写 NULL）；成功恒「操作成功」
- [x] POST /tool/gen/remove：主表+列两表事务删（DB 双查）；**不存在 id 也返回 code 0「操作成功」**（经典版固定 success 语义）
  - [x] 空 ids 行为按 deviations #33 处理（无操作 + code 0，不复制经典版 `in ()` 语法错误）
  - [x] 隔离性实测：两张表同存时只删目标行（不误删、不连带）
- [x] gen 模块无数据权限、无 admin 保护（createTable 排除；负向核对：list/import/remove/edit 无 isAdmin 判断）
- [x] 排除端点负向：preview/{id} / download/{name} / genCode/{name} / batchGenCode / createTable / synchDb/{name} / GET edit/{id} 全部 404（路由未注册）

## 页面级验收（浏览器）

- [x] 列表页：空态/有数据态渲染
- [x] 排序点击（实测点「表名称」表头发出 `orderByColumn=tableName&isAsc=asc`；tableName/tableComment/className/createTime/updateTime 均在排序白名单内）
- [x] showExport **实际导出 CSV 文件**——2026-10-01 用户侧浏览器实测：勾选数据行 → 导出下拉（CSV/TXT/Word/Excel 四格式）→ 点 CSV 触发下载，**用户确认文件已保存**（前端导出为 bootstrap-table 纯本地能力，零后端端点）
- [x] rememberSelected：第 1 页勾选 → 翻页 → 回第 1 页勾选保持（**实测 stillChecked=true**，13 行数据分 2 页场景）
- [x] 导入弹窗：精简配置（无搜索/刷新/翻页器切换栏）；勾选 → 提交 → 父页表格刷新出新行（实测导入 sys_config）；再次打开弹窗该表消失
- [x] 操作列五按钮按权限串显隐（admin 全显；非 admin 会话四 flag 全 `hidden`）；删除确认框链路（**确认后真删** + 取消不删均实测）；**预览/编辑/同步/生成代码点击 404 属有意**（浏览器实测预览 404）
- [x] 工具栏「创建」按钮 isAdmin 显隐逻辑存在（admin 会话可见；非 admin 实测 `class="btn btn-success hidden"`）；批量生成按钮 multiple 禁用态（`class="btn btn-success multiple disabled"`）
- [x] 页面在主框架 iframe 内打开、标签页标题正确（侧栏菜单进入，活动标签「代码生成」，iframe src=/tool/gen）

## 横切与纪律自查

- [x] #[Perm] 8 处与 spec 表逐一对应（view 1 / **list 5** / edit 保存 1 / remove 1——勘误 4 更正计数，总数与端点一致）
- [x] #[Log] 3 处：title=代码生成，business_type 6/2/3 与 spec 对应；sys_oper_log 落库核对（实测 6=导入 / 2=修改 / 3=删除 三类均落库）
- [x] RepeatSubmit 未挂本模块路由（经典版无 @RepeatSubmit 实锤，负向 grep：route/app.php 的 tool/gen 组无 RepeatSubmit）
- [x] 业务代码零处直接引用 predis；无 Redis 消费（gen 模块无缓存需求——负向 grep 仅命中 GenUtils 文档注释里的「不碰 Db / Redis」字样）
- [x] 权限两通道一致：sys_menu.perms（tool:gen:view/list/edit/remove/preview/code 共 6 行——preview/code 两串仅被 check_perm 消费，#[Perm] 无对应端点，注明）与 #[Perm] 注解互查无遗漏；**非 admin 负向实测反证权限串无笔误**
- [x] information_schema 查询 SQL 与经典版 Mapper XML 逐句对照（排除规则/`(select database())`/lower like/三个 case 推断）；**另补列标签大小写归一**（MySQL 8 大写，见实施记录）
- [x] 表结构零变更（gen_table 21 列 / gen_table_column 22 列，动工前后一致）；sys_job/sys_job_log 未触碰（10.0.0 的事，sys_job 3 行保持）
- [x] config/gen.php 五项与经典版 generator.yml 原值一致（packageName 字面照存）

## Deviations 核对

- [x] deviations **#31** 正式登记（模板链排除：preview/download/genCode/batchGenCode/createTable/synchDb + GET edit 渲染），文字与范围拍板结论一致；GO 版 #19 先例已在 spec 引用
- [x] deviations **#32**（db list 行键裁剪）与 **#33**（remove 空 ids）登记
- [x] 页面 quirk 照抄核对：showExport 前端导出保留 / 排除按钮点击 404 有意 / 创建按钮按角色不按权限

## 测试数据清理记录

清理完成日期：**2026-10-01**

- [x] gen_table / gen_table_column 清空复原（官方 SQL 无预置数据，终态 **COUNT=0** 实测；动工中曾达 13 行 / 146 列）
- [x] 测试导入涉及的实体表（sys_notice/sys_config/sys_user… 共 13 张）零写入（生成器只读 information_schema + 写 gen 两表——行数未变）
- [x] 临时表 **tmp_gen_ok / tmp_gen_atomicity 已 DROP**（原子性测试造物，SHOW TABLES LIKE 'tmp_gen%' 为空）
- [x] 临时权限验证角色/用户清理：sys_role(gen_test) / sys_user(gen_tester) / sys_user_role / sys_role_menu 关联行**全删**（实测各表计数 0）
- [x] Redis 测试会话键清理：删除 admin（curl 会话）、gen_tester、浏览器注入会话共 3 个 `session:*`；匿名会话由 30 分钟空闲 TTL 自清
- [x] Redis 保留 1 个演示会话——**有意保留**：为便于用户在浏览器窗口中继续观察 11.0.0 页面（admin 会话，30 分钟空闲自动过期；不影响数据）
- [x] runtime/download 残留（本模块无导出链路，核对为空：文件数 0）
- [x] admin 会话状态复原（admin/admin123 冷登录实测「登录成功」）
- [x] sys_oper_log / sys_logininfor 测试行**保留并注明**：sys_oper_log 中 title='代码生成' 为真实操作留痕（含被拦截/被校验拒绝的尝试），sys_logininfor 含测试登录；二者为审计日志，按 9.0.0 先例保留不删
