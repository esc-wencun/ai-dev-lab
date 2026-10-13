# 2.0.2 AI 对话域：前端先行（阶段 A：UI + Mock）

> **状态：📝 阶段 B 细化稿待二次确认（2026-10-11）；阶段 A 已完成（用户评审通过）**
>
> 对话域采用**两阶段交付**（2026-10-11 用户定的流程）：
> - **阶段 A（本 spec 现在的内容）**：先做前端界面，API 用 Mock（藏 在真实接口签名后的假实现）——用户看真实界面定交互、找优化点；
> - **阶段 B（文末仅骨架）**：根据 UI 评审结论**细化需求、调整方案，二次确认后**再开发服务端与业务逻辑（建表 SQL、后端接口、真实流式）。
>
> 表设计权威源：[AI相关表设计方案.md](../AI相关表设计方案.md) 第二部分对话域（已过逻辑审查：快照齐全、msg_status 状态机、水位线压缩、重生成去重）。阶段 A 的 Mock 数据结构**按该表设计建模**——UI 吃的就是未来真实接口的形状，阶段 B 换数据源时组件不改。

## 一、阶段 A 目标与范围

**做**：
- 正式对话页替换 2.0.0 的临时冒烟页（复用静态路由 `/ai/chat`，不动菜单表）；
- 完整交互面：左侧会话列表（新建/重命名/置顶/删除/按时间排序）+ 右侧对话区（多轮气泡、Markdown 暂纯文本、生成中动效、出错/取消样式）+ 底部输入区（Enter 发送 / Shift+Enter 换行 / 生成中可停止）；
- Mock API 层：`src/api/ai/chat.js` 定义真实接口签名，`src/api/ai/mock/` 放假实现（内存会话存储 + 定时切片假流式），组件只 import 签名；
- 消息状态机按设计文档 `msg_status` 建模：生成中/完成/出错/已取消 四态都有 UI 表现。

**不做**（阶段 B 再说）：
- 真实后端接口、建表 SQL、菜单 SQL、权限点；
- 智能体选择（本阶段会话一律"自由对话"形态，Mock 里的 agent 字段预留 null）；
- Markdown 渲染 / 代码高亮（列为 UI 评审候选项：需新增 `marked` + `highlight.js` 依赖，看了纯文本效果再定）；
- RAG 溯源展示（segment_ids 引用跳转）、深度思考折叠（仅对话时展示、不落库——2026-10-12 决策）、附件上传——消息卡预留位置，不做实现。

## 二、Mock API 契约（阶段 A 的"接口先行"，签名即阶段 B 的真实接口）

`src/api/ai/chat.js` 导出（签名对齐对话域表设计）：

```js
// 会话（对齐 ai_conversation）
listConversations()                      // → [{conversationId, title, pinned, pinnedTime, lastMessageTime, messageCount}]
createConversation({ title?, modelId? }) // → conversation（Mock 默认模型 modelId=100 glm-4.7）
renameConversation(id, title)
pinConversation(id, pinned)              // 置顶/取消（pinnedTime 由 Mock 维护）
deleteConversation(id)                   // Mock 连带删消息
listMessages(id)                         // → [{messageId, msgType, content, msgStatus, replyId, modelCode, createTime, ...}]

// 消息发送（流式三态回调——阶段 B 换 SSE，签名不变）
sendMessage({ conversationId, content, handlers: { onBegin(msg), onToken(chunk), onDone(msg), onError(err) } })
//   Mock 行为：立刻落 user 消息 → 插 assistant 占位行(onBegin, msgStatus=0) →
//   定时器按 30~80ms/片 推假 token(onToken) → 完成(onDone, msgStatus=1)；累计约 2~4 秒
stopGenerate(conversationId)             // 中止当前流：占位行置 msgStatus=3，保留已生成部分
regenerate(messageId, handlers)          // 重生成：旧回答 useContext 置 0，新 assistant 挂同 replyId（防重复进上下文的约定在 Mock 里就体现）
```

Mock 实现纪律：内存存储的字段名**逐字对齐表设计**（camelCase 对齐 DB 列），刷新页面即重置（可接受，Mock 无持久化义务）。

## 三、页面结构（阶段 A 的 UI 蓝本）

```
views/ai/chat/
├── index.vue                 # 布局：左 280px 会话栏 + 右对话区；路由复用 /ai/chat（2.0.0 静态路由）
├── components/
│   ├── ConversationSidebar.vue   # 会话列表：置顶组/普通组按 lastMessageTime 倒序；hover 出重命名/置顶/删除
│   ├── MessageList.vue           # 消息滚动区：自动滚底（用户上滚时暂停跟随，出现"回到底部"）
│   ├── MessageItem.vue           # 气泡：user 右/assistant 左；msgStatus 四态渲染
│   │                             #   （0 生成中光标动效 / 1 正文 / 2 出错样式+error / 3 已取消+半截内容）
│   └── ChatInput.vue             # 输入框：Enter 发送、Shift+Enter 换行、生成中变"停止"按钮
└──（原 2.0.0 单轮冒烟页内容废弃删除）
```

交互状态清单（阶段 A 验收就是把这串状态全用 Mock 触发一遍）：空会话 → 发送中 → 流式增量 → 完成；停止 → 半截+已取消标；Mock 强制出错路径 → 错误气泡；刷新 → 会话清空（Mock 提示条说明"演示数据不持久"）。

## 四、阶段 B 技术方案（细化稿，2026-10-11；二次确认后动工）

### 4.1 数据库变更（Task B1）

新建 `RuoYi-Vue/sql/ai_20261012.sql`（可重复执行，风格同 2.0.1）：

1. **三张表**（对话域建表 SQL 原样搬运自设计文档第二部分第五节，含审查修正：tool_ids/mcp_ids/enable_thinking/summary_upto_message_id/context_msg_count 快照与水位线列、msg_status 状态机、suggested_questions）；
2. **内置数据**：
   - 2 个预置智能体（平台使用助手：限定答题范围 system_prompt + welcome_message + suggested_questions；写作助手），visibility=1 公开、user_id=admin 的 userId（取 sys_user.admin 的 id=1）；
   - **不预置会话/消息**（用户数据运行时产生）；
3. **菜单**（2020 号段）：「AI对话」C 类型菜单（component=`ai/chat/index`，perms 留空——对话页本人数据无需按按钮授权；菜单可见性走角色分配）+ 角色授权 admin；（2026-10-13 起升级为**顶级菜单** parent=0，与「AI管理」并列——对话是用户侧功能不入管理目录，见 2.0.4 实施记录）
4. `/ai/chat` 静态路由**保留**（浏览器直达）；菜单与静态路由并存不冲突（菜单渲染的是动态路由，同一组件两处入口）。

同步 `docker/mysql/init/01-ruoyi-schema.sql`（追加三表 DDL，风格同 2.0.1）。本机验证：docker exec 对 13306 执行两次验证幂等。

### 4.2 后端（Task B2-B4，ruoyi-ai 模块内）

```
com.ruoyi.ai/
├── domain/      AiAgent / AiConversation / AiMessage（MP 注解；JSON 数组列用 String 存原文，service 层序列化）
├── mapper/      三 Mapper（BaseMapper + default 分页查询：会话按 userId+pinned+lastMessageTime；消息按 conversationId 升序）
├── service/     IAiAgentService（CRUD + checkUnique 名称）
│                IAiConversationService（我的会话列表/新建[agent 配置快照复制]/重命名/置顶/删除[连带软删消息]）
│                IAiMessageService（按会话拉历史/落库/更新）
├── planner/     AiPromptPlanner（纯函数可单测）：system_prompt + context_summary + 最近 max_contexts 条
│                useContext=1 历史消息（同 replyId 只取最新）+ 单轮 user 消息 → Spring AI Prompt
└── controller/  AiAgentController（list 公开下拉+CRUD）
                 AiConversationController（/ai/conversations CRUD + /ai/conversations/{id}/messages）
                 AiChatController（POST /ai/chat/stream —— SSE 流式端点）
```

**接口清单**（前端 chat.js 契约的落地，字段与 Mock 逐字一致）：

| 方法/路径 | 说明 | 对应 chat.js |
|---|---|---|
| GET /ai/agents | 可用智能体下拉（status=0，公开 + 本人私有） | listAgents |
| GET /ai/conversations | 我的会话列表（置顶组+时间倒序） | listConversations |
| POST /ai/conversations | 新建（agentId 可空；服务端复制 agent 快照：system_prompt/knowledge_ids/tool_ids/mcp_ids/enable_thinking + title=角色名；last_message_time=now） | createConversation |
| PUT /ai/conversations/{id}/title | 重命名 | renameConversation |
| PUT /ai/conversations/{id}/pin | 置顶/取消（body: pinned） | pinConversation |
| DELETE /ai/conversations/{id} | 删除（连带消息软删；进行中流拒绝或中止） | deleteConversation |
| GET /ai/conversations/{id}/messages | 历史消息升序 | listMessages |
| POST /ai/chat/stream | **SSE 流式**：body {conversationId, content}。事件序列：`begin`（assistant 占位消息 JSON，msgStatus=0）→ 多个 `token`（增量文本）→ `done`（最终消息 JSON：msgStatus=1/2/3 + tokens + duration）| sendMessage/stopGenerate |

**SSE 端点实现要点**（对位生产级流式落库范式）：
- `produces = TEXT_EVENT_STREAM`，返回 Flux<ServerSentEvent<String>>；事件名做 domain 前缀（`data: {"event":"token","payload":...}` 信封，阶段 A 前端 handlers 三回调直接对位）；
- **取消**：前端停止按钮调 `DELETE /ai/chat/stream/{messageId}`（新增一个非流式端点），服务端 dispose 该会话 Flux 订阅 + doOnCancel 分支置 msgStatus=3 落库——三路落库（onComplete=1 / onError=2 / cancel=3）；
- 鉴权：SSE 端点走登录态（token 经 query 参数传递——EventSource 不支持自定义 header，前端 fetchEventSource 或 axios-download 模式二选一，实施时定并记坑）；
- AiModelResolver/ChatModelRouter（2.0.1 产物）取模型；msg_status 启动孤儿清理（0 且超 1 小时 → 3）放定时任务或启动钩子（本期启动钩子即可）。

### 4.3 前端切换（Task B5-B6）

1. `chat.js`：USE_MOCK=false，补 real 实现（request 实例 + SSE 消费）；**组件零改动是验收项**（含 stop 语义：Mock 是本地 clearInterval，真实版是调服务端 dispose 端点——chat.js 内部差异，签名不变）；
2. **Markdown 渲染**（阶段 A 定案）：`npm i marked highlight.js dompurify`；MessageItem 的 assistant 气泡 content 走 marked+DOMPurify 渲染（user 气泡保持纯文本）；
3. **模型选择器**（阶段 A 定案）：ChatInput 左侧下拉（读 `2.0.1` 的模型列表接口，type=1 过滤），切换调 `PUT /ai/conversations/{id}/model`（新增契约函数 changeConversationModel——Mock 同步补）；
4. welcome chips / 角色头条 / 懒创建逻辑不变（已是定稿形态）。

### 4.4 参数覆盖与不做项（重申）

- 会话级 temperature/max_tokens/max_contexts 列落库、会话创建时从模型默认带入，但**不暴露 UI**（阶段 A 定案）；
- 上下文压缩（水位线摘要是异步任务）本期只做**数据落点**（context_summary/summary_upto_message_id 列就位 + planner 预留拼装位），**压缩任务本身不做**——contextMsgCount 远不到阈值前无实际需求，登记为后续模块；
- RAG 溯源（segment_ids）/联网/附件：列就位，功能不做（后续 RAG 模块）。

### 4.5 实施记录与实施期缺陷（2026-10-11 端到端验证期）

- **D5**：Reactor 异步线程没有 SecurityContext——落库回调里调 SecurityUtils 抛「获取用户账户异常」（生产级流式范式的「异步线程丢上下文」同款问题）。解法：请求线程先取 userId/username，闭包捕获传入各回调。
- **D6**：停止句柄若在 doOnSubscribe（订阅传播时）才注册，模型思考期（首 token 前数十个空 delta）内到达的 stop 请求会扑空。解法：请求线程组装 Flux 时同步注册句柄（AtomicReference<Subscription> 占位，doOnSubscribe 时 set）。
- **remark 列坑**：BaseEntity.remark 是真实列映射（无 exist=false），表未建该列则实体任何 SELECT 报错；子类声明同名字段 + @TableField(exist=false) **不能**遮蔽父类映射。解法：ai_conversation/ai_message 补建 remark 列（增量 SQL 与 init SQL 已同步）。
- SSE 鉴权落地形态：前端 fetch + ReadableStream 消费（标准 Authorization 头），无 query 传 token（spec 4.2 预留问题的定案）。
- **对话页视觉风格定稿（2026-10-11 用户确认）**：Claude Code 终端风格——无气泡无头像、全左对齐平铺、角色等宽前缀（`>` 用户 / `●` 助手）、用户输入灰底窄条+左竖线、助手 Markdown 平铺正文、建议问题终端菜单式（`›` 前缀竖排）、元信息等宽弱化（`2.3s`/`48 tok`/`↻`）、输入框带 `>` 提示符行 + Esc 停止快捷键、会话栏左竖线 active 指示。配色仍用 RuoYi 主题变量（浅色），字体 JetBrains Mono/Consolas 等宽链。后续模块（SSE 真实接入等）样式不再动。
- 验证结果：Planner 单测 6/6、Resolver 单测 12/12（顺带清了 2.0.1 遗留账）、SSE 三路落库全断言通过、快照复制/CRUD/删除级联/智能体删除校验实测通过。
- **思考过程功能（2026-10-12 实施，参考客户端做法）**：四项全部落地——
  ① **模型思考开关接通**：`doStream` 读会话 `enable_thinking` 快照，OpenAI 兼容协议经 `OpenAiChatOptions.extraBody` 透传（qwen3→`enable_thinking:true`、GLM→`thinking:{type:enabled}`、DeepSeek R1 等默认思考模型不传参）。**坑 D7：请求级 options 必须带全 baseUrl/apiKey/model**——只传 extraBody 会整体覆盖、回退到 spring-ai 默认（gpt-5-mini/官方地址），404。
  ② **思考耗时显示**：前端 onReasoning 首帧起表、首个正文 token 止表，折叠行显示 `✻ 思考了 Ns`（生成中为 `✻ 思考中…`）；_thinkingMs 仅本地态不落库（回放无此数据可接受）。
  ③ 思考过程 UI：`✻` 前缀折叠行（黄星）、思考体斜体灰字、生成中默认展开。
  ④ **坑 D8（端到端抓到的真 bug）**：GLM 网关帧结构——`reasoning_content` 走 metadata 累积、**同一帧的 text 同时携带正文增量**，原 map() 的"reasoning/content 互斥"假设导致正文增量被 reasoning 分支吞掉（表现为"只思考不回答"）。修复：map→flatMap，一帧两路各自非空各发一个事件。修复后实测：reasoning 事件 ×90、token ×1（content="2"）、reasoning 落库 17072 字符、13s 完成。
- **Lombok 化（2026-10-11，用户指令）**：ruoyi-ai 全模块 35 类转 Lombok——实体用 `@Getter @Setter`（**不用 @Data**：避免 MP 实体上生成全字段 equals/hashCode 的陷阱）、服务/控制器注入字段转 `private final` + `@RequiredArgsConstructor`、手写 Logger 转 `@Slf4j`；内部纯数据类 ResolvedAiModel 同步转换。依赖只在 ruoyi-ai pom（provided 1.18.48），**其余模块不动**（保持手写风格，不扩大影响面）。验证：javac -proc:full 全量 0 错误 + mvn BUILD SUCCESS + 冷启动 SSE 全链路（model 101 流式正常落库）。**前置条件：IDEA 需开启 annotation processing**（Settings → Build → Compiler → Annotation Processors），否则本地编译报"找不到 getXxx"。
- **D9（2026-10-12 端到端）：reasoning_content 列容量不足**——建表时该列用 `text`（64KB 字节上限，utf8mb4 中文约 2.1 万字），正文 content 却是 `mediumtext`（16MB）；GLM 深度思考推理链超限后完成路 `updateById` 抛 `Data too long for column 'reasoning_content'`，**onErrorResume 错误路带着同一超长字段再落库再炸**，done 事件未下发、流异常终止。两个连锁误诊点：①后端日志随之打出成串 `AuthorizationDeniedException`——那是 SSE 响应已提交后 Tomcat 异步 ERROR 派发重走 Security 过滤器所致（异步线程无 SecurityContext），**不是权限问题**；②流失败留下 msg_status=0 的孤儿占位行、会话 message_count/context_msg_count 未刷新（refreshMessageStats 在写库处崩掉没执行）。**最终处置（2026-10-12 决策，10-13 落地完毕）**：思考内容改为**仅对话时展示、不落库**，`reasoning_content` 列连同实体字段一并删除（D9 随之根除；线上库已 DROP，服务以新代码运行实测正常）。顺带修复既有前端 bug：onToken/onDone 用服务端载荷整体替换消息对象，而 fastjson2 不序列化 null 字段、载荷无 reasoningContent key，正文 token 一到思考缓冲即被清掉——改为本地态字段保留合并。**10-13 语义确认（用户）**：思考由会话 `enable_thinking` 快照门控——开启 = 思考仅流式返回前端展示、不落库；关闭 = 不向模型请求思考（曾短暂改为"恒开"，已回退）。当前局限：自由对话（无智能体）没有开关 UI，仅选开思考的智能体（如"平台使用助手"）建会话才有思考；开关 UI 随 2.0.4 智能体管理页落地。
- **D9 附带修复（2026-10-13，UI 层隐匿 bug）**：SSE `token`/`reasoning` 事件载荷只带 `{delta}` 不带 messageId，前端 stream.js 把裸载荷传给组件，处理器按 `a.messageId` 定位消息行永远 miss（findIndex 恒 -1）——**流式期间 UI 从不更新**，正文一直靠 done 事件整段蹦出、思考完全不显示（curl 断言只数事件数未暴露；Mock 载荷带完整消息所以阶段 A 正常）。修复：stream.js 在分发 token/reasoning 时合并 begin 登记的占位 messageId 再回调。
- **D10（2026-10-14，用户报"思考文字拼接错乱"）：reasoning 累积/增量语义错配**——GLM 网关 metadata 里的 reasoning_content 是**累积全量**（每帧含全部前文：`'The'`→`'The user'`→`'The user is'`，尾帧 732 字符），后端按增量逐帧下发、前端再 `prev+chunk` 累加，187 帧拼出 8.4 万字符重复膨胀文本（D8 当时的"落库 17072 字符"实为同源膨胀值，恰被 UI 不显示掩盖）。D8 注释"reasoning_content 走 metadata 累积"写了结论但代码按增量处理了——**两处教训：注释里的语义结论必须落进代码行为；网关流式字段语义（增量 vs 累积）要抓原始帧验证，不能靠单轮结果反推**。修复：后端下发前归一化——新帧以前文为前缀且更长判为累积帧取后缀（StringBuilder 同步置换为全量），否则增量帧透传；qwen 等真增量网关不受影响，前端拼接逻辑不变恒正确。归一化已抽静态纯函数 normalizeReasoningDelta 并补单测 5 用例（累积帧取后缀/增量透传/全量重复帧空增量/异常帧兜底/空 buffer）。
- **D11（2026-10-14，用户报"思考过程过长时 Stream failed"）：请求超时 60s 掐断长思考**——库实查 5 条失败消息 duration_ms 全部 ≈60000 整（error_message="Stream failed"），qwen3:4b 与 GLM 均中招；排查链：openai-java SDK Timeout 默认（connect 1min/read=request 10min）→ OkHttp 装配逐项排除 → 钉到 **Spring AI 2.0.1 `AbstractOpenAiOptions.DEFAULT_TIMEOUT = Duration.ofSeconds(60)`**（OpenAiChatOptions.timeout 属性，构建 options 时未显式设置即吃默认）。深度思考"仅思考不出正文"期超过 60s 请求即被掐断。修复：ChatModelRouter 实例级与 buildThinkingOptions 请求级 options 均显式 `timeout=30min`。教训：**第三方客户端库的"默认值"要用字节码/源码实证，报错文案（Stream failed）与真实原因（超时）可能完全对不上**；60s 整数倍数时长是最快的超时指纹。
- 附带：doStream 重构（2026-10-14，用户指令）——编排层从 210 行收到 ~90 行：三路落库合一 finishAssistant、prepareUserMessage/insertAssistantPlaceholder、buildEffectivePrompt（思考门控收口）、normalizeReasoningDelta、SSE 事件构造三小件；行为零变化（javac 41 文件 0 错误 + 归一化单测 5/5）。
- **D12（2026-10-14，用户报"切换会话后前端不再响应"）：流句柄模块级单例无会话隔离**——stream.js 的 currentController/currentMessageId/activeResolve/activeReject 全局一份，思考模式把流拉长到分钟级后跨会话并发流互相踩：后发流覆盖单例 → 旧流 done 误 resolve 新流的 Promise 并清空停止句柄（停止失灵、状态错乱）。修复：句柄表改 Map（conversationId → {controller,messageId,resolve,reject}），finishStream 加条目一致性检查（只结算自己的条目），stopStream(conversationId) 按会话精确停止；顺带封 stop 后 Promise 悬挂缺口（主动断流本地立即 settle，组件 finally 兜底照常执行）。**用户浏览器实测通过**。
- **D13（2026-10-14，用户报"切换会话后思考过程没了"）：思考缓冲只存当前可见数组**——思考不落库（D9 处置）后思考只在前端内存，此前仅存当前可见 messages 数组：切会话被服务端数据覆盖丢失、后台流增量 findIndex 不到行静默丢弃。修复：按 会话:消息 键控的内存 overlay（reasoningOverlay Map），增量恒写 overlay、openConversation/失败重拉时贴回、onDone 终态并入、删会话清理条目；仍是纯前端态（刷新即失，"不落库"边界不变），只是跨会话切换不丢。**用户浏览器实测通过**。
- 小改进（2026-10-14，用户指令）：①输入框 autosize（minRows 1 / maxRows 10，Claude Code 同款随内容长高、超限内滚）；②新建会话后光标自动落输入框（ChatInput defineExpose focusInput + handleCreate nextTick 调用）。均已实测通过。
## 五、风险与约束

| # | 风险/约束 | 应对 |
|---|---|---|
| R1 | 假流式与真 SSE 行为差异（背压/断连/乱序）在 Mock 里暴露不出来 | 阶段 B 联调时预留"流式异常注入"验证项；组件层把 onToken/onError 的调用假设写在注释里 |
| R2 | Mock 与真实接口形状漂移 | Mock 字段名逐字对齐表设计（spec 第二节纪律）；阶段 B 切换时组件零改动是验收项 |
| R3 | 三版通用性：真实接口 Java 先行 | 阶段 A 纯前端+Mock，天然三版无差；阶段 B 沿用 2.0.1 R5 的声明式偏离 |
| R4 | 编译纪律：IDEA 运行中不编译 | 本阶段纯前端改动，不触发；阶段 B 时按 RuoYi-Vue/AGENTS.md 编译纪律执行 |
| R5 | 替换 2.0.0 冒烟页后旧入口行为变化 | 临时页本来就是过渡产物；2.0.0 checklist 已登记"2.1.0 正式化时重写"，本模块即该事项的落地 |
