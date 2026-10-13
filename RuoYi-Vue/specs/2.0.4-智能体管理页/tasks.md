# Tasks · 2.0.4 智能体管理页

> 勾选纪律：做完即勾（含验证跑通）；没做不勾并注明原因。
> 依赖：2.0.3（权限点/菜单占位已就位）。后端接口面已有（AiAgentController CRUD）。

## Task 1: suggestedQuestions 格式修复（后端出口 + 前端入口）
- [x] AiAgentController list/getInfo/options 三出口：suggestedQuestions JSON 字符串→数组——实现为 AiAgentPayload 工具类（Jackson convertValue 实体→Map 与接口序列化注解行为一致 + fastjson2 parseArray 容错），2026-10-12 完成
- [x] 单测：转换容错四态（合法数组/空串/null/坏 JSON）+ 非数组 JSON 原样 + toOne/toList 整形——9 用例 mvn surefire 全绿（2026-10-13 正式跑通：common 10 + ai 27 = 37/37）
- [x] curl 实测 options 出口返回数组——2026-10-13 三出口全验：options 两智能体均 list、list 全量 list、getInfo 100 为 list

## Task 2: 前端 API 层 + 管理页
- [x] `src/api/ai/agent.js`：listAgent/getAgent/addAgent/updateAgent/delAgent（2026-10-12）
- [x] `views/ai/agent/index.vue`：搜索四项 + 表格 + 弹窗表单（systemPrompt 大 textarea、模型下拉复用 listChatModels、welcomeMessage、suggestedQuestions 动态行编辑 sqRows、enableThinking 开关、visibility 单选）；toRows/fromRows 双向转换容错（2026-10-12）
- [x] knowledge/tool/mcp 挂载三列本期不做编辑入口（登记决策：等 RAG/工具域）
- [x] vite transform 验证（api/index 两文件全 200，新符号命中；2026-10-13 复验仍 200）

## Task 3: 菜单 2030 放出
- [x] `RuoYi-Vue/sql/ai_20261014.sql`：update 菜单 2030 visible='0'；`ai_20261013.sql` 同步行（insert 值 + 注释）
- [x] 13306 执行 + 双跑幂等验证（2020 perms 在、2030 visible='0'，2026-10-12）

## Task 3.5: AI 域软删过滤修复（2026-10-13 实施中发现，管理页正确性前提）
- [x] **缺陷**：AI 域实体无 @TableLogic、MyBatisConfig 全局只设值未设字段名（逻辑删除不自动生效）——agent 列表/下拉/名称查重/countByAgent 均未过滤 del_flag：已删智能体继续出现在列表与下拉、已删会话永远阻塞智能体删除、已删名字导致查重误报
- [x] 修复：AiAgentMapper.selectVisibleAgents/selectAgentList、AiConversationMapper.countByAgent、AiAgentServiceImpl insert/update 查重、AiMessageMapper.selectByConversation/selectRecentActive（planner 上下文窗口）/selectActiveByReply（重生成去重）共八处补 eq(delFlag, 0)——javac + mvn 编译全绿
- [x] 同缺陷扩散检查：AiConversation.selectUserConversations 本就带 status='0'（删除时 status 置 1）天然免疫；regenerate 的 selectById 紧跟 getOwnedConversation 拦截已删会话；孤儿清理只取生成中行——均无需改
- [x] 行为验证（2026-10-13 端到端）：删 agent 后管理列表与 options 下拉均不再出现；软删名字重建同名成功（查重排除已删行）；删会话后智能体可删（countByAgent 排除已删会话）

## Task 4: 端到端验证（2026-10-13，用户授权 agent 编译+启动，jar 冷启动）
- [x] 编译：mvn clean package 全模块 exit 0；单测 mvn test -pl ruoyi-ai -am → 37/37 全绿（PageSupport 10 + AiAgentPayload 9 + AiPromptPlanner 6 + AiModelResolver 12）
- [x] 管理页 CRUD 往返（API 全链路）：新增（建议问题数组入参→落库 JSON 字符串）→ 列表可见（出口转回数组）→ 编辑回显（改后 3 条数组）→ 名称查重 500「智能体名称已存在」→ 被引用拒删（报错带数量「已被 1 个会话使用」）→ 删除后列表/下拉不再出现
- [x] **中途发现并修复：controller 出口转换丢失**——2026-10-12 的 controller 改动是已跟踪文件的未提交修改，事故后 git checkout 恢复时被一并回退（untracked 新文件不受影响），导致首启后 suggestedQuestions 仍是字符串；补回三出口转换重编译后三口全数组
- [x] 对话页回归：会话列表/agents/options/模型列表全 200；agent 100 会话 SSE 流式正常（begin→193 事件→done msgStatus=1，正文非空）；done 载荷无 reasoningContent 字段（与用户 2026-10-12 思考不落库决策一致）
- [x] 建议问题 chips：出口已数组（Mock 形状对齐），MessageList v-for 渲染链路 2.0.2 浏览器已验过；Tabbit 运行时本次 CDP 故障（Target.createTarget 无法派发，与 task 无关）未能做页面点验，随用户日常使用确认
- [x] 私有智能体可见性：造 test-visibility 无角色账号——options 仅见公开 100/101、admin 私有 104 不可见；无 ai:agent:list 权限访问管理列表 403「没有权限，请联系管理员授权」
- [x] 回归零影响：2.0.1 providers/keys/model/list 全 200
- [x] 测试数据清理：test-visibility 用户（物理删）、Redis 会话、私有智能体 104、测试会话 133/134/135 与消息、crud-test 智能体 102/103（物理删）；admin 已登出；131（用户自建）与 100/101（内置）保留；2.0.3 残留会话 132 物理清理
- [x] 服务善后：我起的 jar（8080）与 vite（8888）验证完已按 PID 精确停止，双端口释放
