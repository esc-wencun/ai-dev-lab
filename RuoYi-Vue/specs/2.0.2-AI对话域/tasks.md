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
- [x] `RuoYi-Vue/sql/ai_20261012.sql`：三表 DDL（搬运设计文档，含 msg_status/水位线/context_msg_count/suggested_questions）+ 2 个预置智能体（100 平台使用助手/101 写作助手，带开场白与建议问题 JSON）+ 菜单（2020「AI对话」，2026-10-13 起为顶级菜单，seed SQL 已同步）
- [x] docker/mysql/init/01-ruoyi-schema.sql 同步三表 DDL（+95 行）；13306 执行两次幂等验证通过（agents=2/convs=0/msgs=0/menu 2020 在；suggested_questions JSON_EXTRACT 正常）
- [ ] 表设计文档第二部分标"已实施"注记——随模块收口一并做

## Task B2: domain + mapper + service（CRUD 层）
- [x] 三实体（AiAgent/AiConversation/AiMessage，含 delFlag + 全部审查修正列）（实施坑：BaseEntity.remark 真实列映射且子类遮蔽无效，两表补建 remark 列——见 spec 4.5）+ 三 Mapper（会话置顶+时间倒序、消息按会话升序、最近活跃窗口、replyId 活跃查询、孤儿查询）
- [x] IAiConversationService：新建时 agent 配置快照复制（system_prompt/knowledge_ids/tool_ids/mcp_ids/enable_thinking + title=角色名 + last_message_time 初始化）——快照五字段完整；私有智能体可见性校验；模型解析链 显式>agent默认>最新type=1
- [x] 重命名/置顶/删除（连带消息软删）/我的会话列表/切模型（type=1 校验）/归属校验（getOwnedConversation）——**javac 语法检查全过（JDK17 全路径，17 文件 0 错误）；单测与端到端待服务停止窗口/用户 IDEA Build 后执行（编译纪律第 0 节）**
- [x] IAiAgentService（名称查重/私有必填归属人/删除前置校验被会话引用拒删）+ AiAgentController（/options 下拉登录态 + 管理 CRUD 权限点）

## Task B3: AiPromptPlanner + SSE 流式端点
- [x] AiPromptPlanner 纯函数：system_prompt + summary 拼装位 + 最近 max_contexts 条 useContext=1（同 replyId 取最新）+ 当前消息 → Prompt；测试类 6 用例就绪（裁剪窗口/重生成去重/useContext 剔除/summary 顺序/空装配/默认窗口）
- [x] POST /ai/chat/stream（SSE：begin/token/done 信封）+ DELETE /ai/chat/stream/{messageId}（Subscription.cancel 触发 doOnCancel 落库）
- [x] 三路落库代码就位：concatWith 完成路=1 / onErrorResume=2 / doOnCancel=3（保留半截）；占位行先插后更；AiStartupRunner 启动孤儿清理（0 超 1 小时→3）
- [x] AiModelResolver/ChatModelRouter 接入；SSE 鉴权方案定案：**前端 fetch+ReadableStream 消费（带标准 Authorization 头），不用原生 EventSource**——无 query 传参降级（spec 4.2 注记兑现）
- [ ] 单测执行（mvn test）+ curl 级 SSE 事件序列断言——**待服务停止窗口/用户 IDEA Build 后（编译纪律第 0 节）**

## Task B4: 智能体接口
- [x] GET /ai/agents/options（公开+本人私有，status=0，登录态）+ 管理页 CRUD（ai:agent:* 权限点；按钮菜单 SQL 待管理页模块时配——对话页只需 /options）
- [x] agent_name 查重；visibility=0 时 user_id 必填（应用层兜底作者）

## Task B5: 前端切真实 API
- [x] chat.js USE_MOCK=false + real 实现（`src/api/ai/real/`：conversation.js REST 走统一 request 实例 + stream.js fetch+ReadableStream 手工 SSE 解析，标准 Authorization 头——spec 4.5 定案；stop 语义 = AbortController 断流 + DELETE /ai/chat/stream/{messageId}；**组件零改动**：vite transform 全 200，chat.js 转换后 USE_MOCK=false 且 real import 生效）
- [x] npm i marked highlight.js dompurify（阶段 A 期已装）+ **marked-highlight（B5 补装）**：MessageItem 原 `marked.setOptions({highlight})` 在 marked v18 静默失效（v5+ 移除该选项，实测 `setOptions` 不报错但渲染无 hljs span），改 `marked.use(markedHighlight(...))` 扩展——修复后 hljs-keyword 等 class 正常产出（node 直跑验证）
- [x] 模型选择器契约：changeConversationModel（Mock 已有）+ real conversation.js 对应 PUT /ai/conversations/{id}/model（后端端点 B2 已就位）
- [x] **regenerate 后端支持（B5 发现的契约缺口）**：组件的 regenerate(messageId, handlers) 在后端无对应端点——DTO 加 regenerateMessageId、service 接口加 regenerate、impl 抽 doStream(conversationId, content, regenReplyId) 共用主链路（旧回答 useContext 置 0 / 新 assistant 挂原 replyId / 原问题不重复落库不重复进 prompt——planner 加 null 保护跳过尾部追加）；javac -proc:full 5 文件 0 错误（MediaContent 报错是我 classpath 漏 spring-ai-commons，非代码问题）
- [x] 浏览器全链路实操——**接口层全绿（2026-10-11 两轮）**：agents/options、会话 CRUD、SSE 完整序列（begin→391×token→done，msgStatus=1 content/durationMs 落库）；cancel 路（流中 DELETE → msgStatus=3 落库；"该消息不在生成中"防重复停止）；**regenerate 全语义闭环（用户 Build 后实测）**：问 2+2 → 答 4（msg 165）→ regenerate(165) → begin/done 正常、新消息 166 挂同 replyId=164、旧回答 useContext 1→0；再问"answer plus 10"模型答 14——证明上下文用的是新回答而非被剔除的旧回答；页面级实操由用户日常使用确认

## Task B6: 回归收口
- [ ] 既有接口零影响 + 2.0.1 三管理页回归 + 测试数据清理
- [ ] checklist 勾完 + specs/README 登记 2.0.2 ✅ + CHANGELOG 条目（随提交）
