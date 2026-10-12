# Tasks · 2.0.4 智能体管理页

> 勾选纪律：做完即勾（含验证跑通）；没做不勾并注明原因。
> 依赖：2.0.3（权限点/菜单占位已就位）。后端接口面已有（AiAgentController CRUD）。

## Task 1: suggestedQuestions 格式修复（后端出口 + 前端入口）
- [ ] AiAgentController list/getInfo/options 三出口：suggestedQuestions JSON 字符串→数组（fastjson2 parseArray，容错坏值原样返回）；转换收口一个静态工具方法
- [ ] 单测：转换容错四态（合法数组/空串/null/坏 JSON）——javac + JUnit 直跑（IDEA 运行中绕 mvn 方式）
- [ ] curl 实测 options 出口返回数组

## Task 2: 前端 API 层 + 管理页
- [ ] `src/api/ai/agent.js`：listAgentsManage/getAgent/addAgent/updateAgent/delAgent
- [ ] `views/ai/agent/index.vue`：搜索（名称/分类/可见性/状态）+ 表格（avatar 图标/名称/分类/默认模型/可见性/状态/排序/操作）+ 弹窗表单（systemPrompt 大 textarea、模型下拉、welcomeMessage、suggestedQuestions 动态行编辑、enableThinking 开关、visibility 单选）；提交前 JSON.stringify 建议问题、回显 JSON.parse
- [ ] knowledge/tool/mcp 挂载三列本期不做编辑入口（登记决策：等 RAG/工具域）
- [ ] vite transform 验证

## Task 3: 菜单 2030 放出
- [ ] `RuoYi-Vue/sql/ai_20261014.sql`：update 菜单 2030 visible='0'；`ai_20261012.sql` 同步行
- [ ] 13306 执行 + 双跑幂等验证

## Task 4: 端到端验证（用户 Build 后）
- [ ] 管理页 CRUD 往返：新增（含建议问题数组）→ 列表可见 → 编辑回显 → 名称查重 500 → 被引用拒删（agent 100 被会话引用场景）
- [ ] 对话页回归：智能体下拉正常、开场白显示、建议问题 chips 正常渲染（不再逐字符）
- [ ] 私有智能体可见性：非作者账号 options 不可见（造测试数据或用现有私有数据）
- [ ] 测试数据清理 + checklist/README 收口
