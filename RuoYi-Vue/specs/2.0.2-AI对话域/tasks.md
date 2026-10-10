# Tasks · 2.0.2 AI 对话域（阶段 A：前端 + Mock）

> 勾选纪律：做完即勾（含验证跑通）；没做不勾并注明原因。
> 依赖：无后端依赖（纯前端 + Mock）；复用 2.0.0 静态路由 /ai/chat。
> 环境约束：本阶段不碰 Java 代码、不碰数据库；前端 dev server 热加载验证。

## Task 1: Mock API 层
- [x] `src/api/ai/chat.js`：接口签名（spec 第二节全量 9 个函数）+ Mock/真实开关（`USE_MOCK = true`）
- [x] `src/api/ai/mock/mockChat.js`：内存存储（字段名逐字对齐对话域表设计）+ 假流式（~45ms/片）+ stop/regenerate 语义——**小步重写完成：输出退化污染两次写坏后改用 python 行循环分段写 + .mjs 副本 node --check 逐段校验；语义自测 A-F 全过（流式完成/停止保留半截 msgStatus=3/重生成同 replyId 且旧回答 useContext=0/出错注入 msgStatus=2/删除级联）**
- [x] Mock 内置 1 个演示会话（含一条历史问答）方便打开即有内容看

## Task 2: 页面骨架与布局
- [x] index.vue 双栏布局（左 280px 会话栏 + 右对话区），替换 2.0.0 冒烟页；静态路由不动
- [x] ConversationSidebar：置顶/普通分组、lastMessageTime 倒序、hover 操作（重命名/置顶/删除）、新建会话按钮（重命名用 prompt、删除带 ElMessageBox 确认）
- [x] 空状态：无会话时引导新建；"演示数据不持久"提示条（el-alert 常驻顶部）

## Task 3: 对话区与消息渲染
- [x] MessageList：自动滚底 + 用户上滚暂停跟随 + "回到底部"浮钮
- [x] MessageItem：user/assistant 气泡样式；msg_status 四态（0 生成中光标动效 / 1 正文+耗时+token 元信息+重生成按钮 / 2 出错样式+error / 3 已取消+半截灰化）
- [x] ChatInput：Enter 发送 / Shift+Enter 换行；发送中按钮变"停止"（调 stopGenerate）；生成中输入框禁用

## Task 4: 全状态走查（Mock 驱动）
- [x] 状态链语义自测（node 动态 import，A-F 六路径全过）：流式完成/停止保留半截(msgStatus=3)/重生成(同 replyId+旧回答 useContext=0)/出错注入(msgStatus=2)/删除级联/演示种子会话
- [x] vite 编译验证：4 组件 + mock 层全部 transform 成功（HTTP 200 零错误），2.0.1 三页不受影响
- [x] 浏览器交互实操——**用户实操通过（2026-10-11）：「页面正常了，交互都没问题」**；过程抓到并修复 MessageList 白屏 bug（defineProps 未接变量名，watch 引用了不存在的 messages 符号）

## Task 5: 评审交接（阶段 A 收口）
- [x] 向用户提交 UI 评审清单（见下方评审请求，含 5 个必答问题）
- [x] 评审结论回写 spec 第四节（用户采纳 agent 全部建议，2026-10-11）；**阶段 B 三件套细化后二次确认，才动后端**

---

# 阶段 B Tasks（细化稿 2026-10-11，二次确认后动工）

> 表权威源：[../AI相关表设计方案.md](../AI相关表设计方案.md) 第二部分（已回写 suggested_questions 列）。接口清单见 spec 4.2 表。

## Task B1: 数据库
- [ ] `RuoYi-Vue/sql/ai_20261012.sql`：三表 DDL（搬运设计文档，含 msg_status/水位线/context_msg_count/suggested_questions）+ 2 个预置智能体 + 菜单（2020 号段「AI 对话」）+ delete 前置
- [ ] docker/mysql/init/01-ruoyi-schema.sql 同步三表 DDL；13306 执行两次验证幂等
- [ ] 表设计文档第二部分标"已实施"注记

## Task B2: domain + mapper + service（CRUD 层）
- [ ] 三实体 + 三 Mapper（会话列表分页排序、消息按会话升序）
- [ ] IAiConversationService：新建时 agent 配置快照复制（system_prompt/knowledge_ids/tool_ids/mcp_ids/enable_thinking + title=角色名 + last_message_time 初始化）——**快照字段一个不漏（设计文档决策 1"复制必须彻底"）**
- [ ] 重命名/置顶/删除（连带消息软删）/我的会话列表；单元测试（快照复制完整性）

## Task B3: AiPromptPlanner + SSE 流式端点
- [ ] AiPromptPlanner 纯函数：system_prompt + summary + 最近 max_contexts 条 useContext=1（同 replyId 取最新）+ 当前消息 → Prompt；**单测全绿**（裁剪窗口/replyId 去重/summary 拼装位）
- [ ] POST /ai/chat/stream（SSE：begin/token/done 信封事件）+ DELETE /ai/chat/stream/{messageId}（dispose 订阅 + cancel 落库）
- [ ] 三路落库：onComplete msgStatus=1 / onError=2 / cancel=3；占位行先插后更；启动孤儿清理（0 超 1 小时→3）
- [ ] AiModelResolver/ChatModelRouter 接入；SSE 鉴权方案（token 传参方式）实施时定并记坑
- [ ] curl 级端到端：SSE 事件序列断言（begin→token*→done）

## Task B4: 智能体接口
- [ ] GET /ai/agents（公开+本人私有，status=0）+ AiAgentController CRUD（权限点 ai:agent:*，菜单按钮 SQL）
- [ ] checkUnique：agent_name 唯一；visibility=0 时 user_id 必填（应用层校验）

## Task B5: 前端切真实 API
- [ ] chat.js USE_MOCK=false + real 实现（request + SSE 消费，fetchEventSource 或降级方案）；stop 语义换服务端 dispose；**组件零改动验收**
- [ ] npm i marked highlight.js dompurify；assistant 气泡 Markdown 渲染 + 代码高亮 + sanitize（user 气泡纯文本）
- [ ] 模型选择器：ChatInput 左侧下拉（2.0.1 模型接口 type=1）+ changeConversationModel 契约（Mock/real 同步补）+ PUT /ai/conversations/{id}/model
- [ ] 浏览器全链路实操：角色会话/开场白 chips/流式/停止/重生成/切模型

## Task B6: 回归收口
- [ ] 既有接口零影响 + 2.0.1 三管理页回归 + 测试数据清理
- [ ] checklist 勾完 + specs/README 登记 2.0.2 ✅ + CHANGELOG 条目（随提交）
