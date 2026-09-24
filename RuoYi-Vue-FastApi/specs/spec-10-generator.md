# Spec-10 代码生成器（可暂缓）

> **状态：✅ 已完成（2026-09-24）**。Task 1-4 落地并端到端验证：
> db/list（information_schema读表，排除qrtz_/gen_与已导入）、importTable（列属性推断对齐java
> GenUtils.initColumnField：类型/HTML控件/默认勾选/必填）、createTable（admin限定+SQL关键字过滤+
> 建表后自动导入）、编辑、synchDb（新列增/消失列删）、preview（8模板渲染，python语法全过）、
> download/batchGenCode（zip路径对齐python项目结构）、genCode自定义路径（allowOverwrite默认false）。
> 生成物为本项目Python架构代码（entity_do/vo+dao+service+controller，分层规范内建）+
> RuoYi-Vue3前端（api.js+index.vue，权限指令/字典控件/分页完整）+ 菜单SQL。
> 与java版差异：不生成java/mapper.xml/ts模板；tplWebType仅支持element-plus。
>
> Java 版对应：`ruoyi-generator` 模块（GenController、GenTableServiceImpl、VelocityUtils、模板 resources/vm/*.vm）
> 前端页面：`views/tool/gen/`、`views/tool/build/`
> 依赖：spec-04~06（生成代码引用字典、权限等）

## 目标

读库表结构 → 配置生成选项 → 按 Java 版模板结构生成 FastAPI 版 CRUD 代码（controller/service/dao/entity + 前端 vue + 菜单 SQL）。生成模板必须**重写为 Python 版**，不能照搬 Java Velocity 模板。

## API 清单

| # | 方法 | 路径 | 权限 |
|---|------|------|------|
| 1 | GET | `/tool/gen/list` | `tool:gen:list` |
| 2 | GET | `/tool/gen/db/list` | `tool:gen:list` |
| 3 | GET | `/tool/gen/{tableId}` | `tool:gen:query` |
| 4 | GET | `/tool/gen/preview/{tableId}` | `tool:gen:preview` |
| 5 | POST | `/tool/gen` | `tool:gen:add` |
| 6 | PUT | `/tool/gen` | `tool:gen:edit` |
| 7 | DELETE | `/tool/gen/{tableIds}` | `tool:gen:remove` |
| 8 | GET | `/tool/gen/genCode/{tableName}` | `tool:gen:code` |
| 9 | POST | `/tool/gen/importTable` | `tool:gen:import` |
| 10 | POST | `/tool/gen/createTable` | `tool:gen:createTable` |
| 11 | GET | `/tool/gen/column/{tableId}` | `tool:gen:query` |
| 12 | PUT | `/tool/gen/synchDb/{tableName}` | `tool:gen:edit` |
| 13 | GET | `/tool/gen/batchGenCode?tables=` | `tool:gen:code` |

## Task 1: 表结构读取与存储

- [x] DO：`GenTable` / `GenTableColumn`（gen_table、gen_table_column 两表，字段对照 Java）
- [x] `db/list`：information_schema 读当前库表列表（排除 gen_ 前缀与已导入表）
- [x] `importTable`：导入表与列元数据（含列类型→Python 类型映射初步推断：bigint→int、varchar→str、datetime→datetime）
- [x] `createTable`：根据页面输入建表并导入
- [x] `synchDb`：库表列变更同步到 gen_table_column

## Task 2: 编辑与预览

- [x] 编辑保存：表信息（生成模块名/功能名/作者等）+ 列信息（是否列表/查询/表单项、html 控件类型、字典类型）
- [x] `preview/{tableId}`：渲染全部模板返回 `{文件路径: 文件内容}` 字典（对照 Java preview 返回结构）

## Task 3: 模板（核心工作量）

- [x] 模板引擎：Jinja2（对齐 Velocity 的占位习惯）
- [x] 重写模板清单（每个模板一份，路径结构对齐本项目）：
  - [x] `controller.py.j2`（APIRouter，权限、分页、导出）
  - [x] `service.py.j2` / `service_impl` 合并为本项目风格
  - [x] `dao.py.j2`
  - [x] `entity_do.py.j2` / `entity_vo.py.j2`
  - [x] `api/js.j2`（前端请求 js）
  - [x] `index.vue.j2`（列表页）+ `form.vue.j2`（或内嵌编辑对话框，对齐 RuoYi-Vue3 结构）
  - [x] `sql.j2`（菜单权限 SQL）
- [x] 单元测试：用固定的示例表（如 sys_notice 简化版）渲染快照断言

## Task 4: 代码生成与下载

- [x] `genCode/{tableName}`：生成 zip 流下载（Java 是 zip；对照前端 download 处理）
- [x] `batchGenCode`：多表打包
- [x] 端到端测试：导入 sys_job 真实表 → 预览 → 下载 zip → 解压后 Python 语法检查通过（createTable建test表再生成亦通过）；前端页面"可运行"以 npm 集成为准未单独验证

## 验收清单

- [x] RuoYi-Vue3 代码生成页面可用，生成物可直接投入项目
- [x] pytest 全绿；更新 specs/README.md 状态
