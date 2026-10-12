# 2.0.4 智能体管理页

> **状态：📝 spec 待确认（2026-10-12）**
>
> 目标：智能体的正式管理页（`views/ai/agent/index.vue`），放开菜单 2030 的隐藏占位。顺带修复调研中发现的线上契约缺口（suggestedQuestions 传 JSON 字符串而非数组，对话页 chips 在真实接口路径下会逐字符迭代）。
>
> 后端接口面已就位（2.0.2 B4 + 2.0.3 权限点）：`/ai/agents` CRUD（名称查重/私有必填归属人/被会话引用拒删）、`/options` 下拉（公开+本人私有）。**本模块预期后端零改动或仅小改**，重心在前端页 + 一个格式修复。

## 一、目标与范围

**做**：
1. **管理页**（标准 CRUD 模式，比照 provider 管理页）：
   - 列表：图标(avatar)/名称/分类/可见性/状态/排序/默认模型/备注 + 搜索（名称 like、category、visibility、status）；
   - 表单：agentName*、avatar（外链）、category、description、**systemPrompt*（textarea 大框，核心字段）**、默认模型下拉（type=1）、**welcomeMessage**（开场白）、**suggestedQuestions（动态行编辑，前端数组 ↔ 存库 JSON 字符串转换）**、visibility（0 私有 1 公开）、sortOrder、status、remark；
   - knowledge_ids/tool_ids/mcp_ids 三个挂载列本期不做编辑入口（RAG/工具域未建，留文本框只读展示或隐藏——实施时按简约取舍，登记决策）；
   - enableThinking 开关（表单单选）；
2. **菜单 2030 放出**：`visible` 改 `0`（ai_20261012.sql 同步 + 增量 SQL）；
3. **suggestedQuestions 格式修复**（契约缺口）：后端 `/options` 与 `/list`、`/{id}` 出口把 JSON 字符串解析为数组返回（实体字段不动，controller 出口转换）；add/edit 入口接收数组序列化入库（或前端转换——实施时按"出口后端统一转、入口前端转"取舍，登记决策）；
4. **回归**：对话页智能体下拉 + 开场白 chips 在修复后正常。

**不做**：
- knowledge/tool/mcp 挂载编辑器（等 RAG/工具域，列就位即可）；
- 智能体编排（SAA Graph 状态图/多智能体模式）——已登记规划表（specs/README"规划中的后续模块"），ai_agent 表届时扩展编排配置列；
- 智能体市场/分类管理页（category 自由文本够用）。

## 二、现状事实（动手前核对）

| # | 事实 | 位置 |
|---|---|---|
| 1 | 后端 CRUD 全就位：list（全量非分页，like 名称 + eq category/visibility/status）、getInfo、add（名称查重 throw）、edit（同名 throw）、remove（被会话引用拒删，报错带数量）、options（公开+本人私有，登录态） | AiAgentController + AiAgentServiceImpl |
| 2 | 权限点 `ai:agent:list/query/add/edit/remove` 注解已挂；菜单 2030-2034 在库（2030 visible='1' 隐藏占位） | 2.0.3 / ai_20261013.sql |
| 3 | AiAgent 实体 18 字段：基础 5（name/avatar/category/description/remark）+ systemPrompt + modelId + 挂载 3（knowledge/tool/mcp，String 存 JSON）+ enableThinking + welcomeMessage + **suggestedQuestions（String 存 JSON 数组字符串）** + visibility/userId/sortOrder/status | AiAgent.java |
| 4 | **契约缺口（实测确认）**：DB 存 `["怎么重置密码？",...]` JSON 字符串，`/options` 直传 → 前端拿到 str；对话页 `welcomeSuggestions` 用它渲染 chips，真实接口路径下 chips 会逐字符迭代（Mock 阶段是数组所以没暴露） | 实测 2026-10-12 |
| 5 | 管理页可照抄的模板：`views/ai/provider/index.vue`（列表+表单+搜索+分页模式——注意 agent list 是全量非分页，列表页不分页直接 el-table 全量） | RuoYi-Vue3 |
| 6 | 对话页消费形状：`agents` 数组元素需有 agentId/agentName/welcomeMessage/suggestedQuestions(数组)/modelId | chat/index.vue:94 |

## 三、技术方案

### 3.1 suggestedQuestions 格式修复（Task 1，先做——对话页 chips 依赖）

**取舍决策：出口转换放后端（controller 出口统一），入口转换放前端（管理页表单提交前序列化）。理由：出口集中一处（options/list/getInfo 三个口）后端转一次所有消费者受益（对话页零改动）；入口只有管理页一个消费者，前端转避免后端做"对象↔字符串"双态字段。**

- 后端：AiAgentController 的 list/getInfo/options 出口，`suggestedQuestions` 字段 JSON 字符串 → 数组（fastjson2 `JSON.parseArray`，解析失败原样返回字符串容错）；实现方式：controller 出口对 List/AiAgent 逐个转（实体字段类型不动，避免 MP 序列化层连锁改）；
- 前端管理页：表单里 suggestedQuestions 是动态行数组（可增删行），提交时 `JSON.stringify`；回显时 `JSON.parse`（容错空/坏值）；
- 单测：出口转换的容错（合法数组/空串/null/坏 JSON）。

### 3.2 管理页（Task 2-3）

- `src/api/ai/agent.js`：listAgentsManage / getAgent / addAgent / updateAgent / delAgent（对齐 controller 五端点）；
- `views/ai/agent/index.vue`：照 provider 页模式（搜索栏/工具栏/表格/弹窗表单）；列表不分页（后端全量，量级小）；表单含动态行编辑器（suggestedQuestions）+ 模型下拉（复用 real listChatModels）；
- `updateConversationParams` 式契约不动对话域任何文件。

### 3.3 菜单放出（Task 4）

- `ai_20261014.sql`：update 菜单 2030 set visible='0'；同步 `ai_20261012.sql` 对应行；
- 13306 执行 + 双跑幂等验证。

### 3.4 验证（Task 5，用户 Build 后）

- 管理页 CRUD 往返（含名称查重 500、被引用拒删、建议问题数组编辑→落库 JSON→对话页 chips 正常渲染）；
- 对话页回归：智能体下拉/开场白/建议问题 chips（真实接口路径）；
- 私有智能体可见性：非作者账号 options 不可见。

## 四、风险与约束

| # | 风险/约束 | 应对 |
|---|---|---|
| R1 | suggestedQuestions 出口转换改动三个读口，若漏一处则前端形状不一致 | 转换函数收口一个工具方法，三处调用；端到端逐口验证 |
| R2 | 列表全量非分页，智能体多了列表长 | 量级小（个位数）；前端加前端筛选即可，后端分页列 2.0.5 再说 |
| R3 | IDEA 运行中编译纪律 | javac + 用户 Build；端到端待 Build 后 |
| R4 | 隐藏菜单 2030 放出后非 admin 角色是否可见 | 菜单按角色分配（admin 通配）；其他角色要看得手动配角色菜单，语义正确 |
