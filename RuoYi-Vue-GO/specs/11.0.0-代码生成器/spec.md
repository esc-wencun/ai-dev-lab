# 11 代码生成器（降级范围实现）

> **状态：✅ 已完成（2026-09-24；按 spec 建议降级为数据层端点，curl 端到端通过）**
>
> Java 版对应：GenController + ruoyi-generator 模块
> 前端页面：views/tool/gen（views/tool/build 不在范围）
> 依赖：1.0.0-基础设施

## 降级范围（spec 建议方案落地）

**实现**（数据层端点，前端列表页可用）：
| # | 方法 | 路径 | 权限串 | 要点 |
|---|------|------|--------|------|
| 1 | GET | /tool/gen/list | tool:gen:list | TableDataInfo；tableName/tableComment like |
| 2 | GET | /tool/gen/db/list | tool:gen:list | information_schema 未导入表 |
| 3 | GET | /tool/gen/{tableId} | tool:gen:query | 详情含列 |
| 4 | POST | /tool/gen/importTable | tool:gen:import | tables 逗号分隔；information_schema 读列装配 gen_table_column（is_pk/is_increment/is_required 推断） |
| 5 | DELETE | /tool/gen/{tableIds} | tool:gen:remove | 主表+列事务删 |

**有意排除**（模板生成类，登记 deviations #19）：
- preview/{tableId}（Velocity 模板渲染预览）、genCode/{tableName}（生成下载）、download、batchGenCode、createTable（SQL 建表导入）、synchDb、editTable 编辑保存——这些端点的产出是 Java/Vue 模板代码，对 Go 项目无产出价值（生成器本身生成 Java 代码）；Go 侧如需代码生成应另行设计 Go 模板。

## 实施记录

- 2026-09-24 完成。gen_table 实际结构以数据库为准（function_author 而非 author、无 engine 列——凭 Java 实体记忆写的列名有两处错，端到端暴露后修正）。
- 端到端：db list、导入 sys_dept（14 列自动装配）、list、getInfo 含列、删除清理关联，全部通过。
