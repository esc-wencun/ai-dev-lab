# 2.0.2 AI 对话域：前端先行（阶段 A：UI + Mock）

> **状态：📝 阶段 A spec 待确认（2026-10-11）**
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
- RAG 溯源展示（segment_ids 引用跳转）、深度思考折叠（reasoning_content）、附件上传——消息卡预留位置，不做实现。

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

## 四、阶段 B 骨架（待阶段 A 评审后细化，二次确认才动工）

1. 建表 SQL：ai_agent / ai_conversation / ai_message 三表 + 菜单（2020 AI对话）+ 权限点（按 UI 定稿的实际功能定）；
2. 后端：三表 CRUD + 会话/消息 service + SSE 流式端点（`TEXT_EVENT_STREAM`，三路落库）+ 接 AiModelResolver（2.0.1 产物）；
3. 前端切真实 API：chat.js 关 Mock 开关补 request 实现；
4. 细化时必须回答的问题（UI 评审的可能产出）：模型选择器放哪、会话内切模型/改参数要不要、智能体入口形态、Markdown 依赖加不加。

## 五、风险与约束

| # | 风险/约束 | 应对 |
|---|---|---|
| R1 | 假流式与真 SSE 行为差异（背压/断连/乱序）在 Mock 里暴露不出来 | 阶段 B 联调时预留"流式异常注入"验证项；组件层把 onToken/onError 的调用假设写在注释里 |
| R2 | Mock 与真实接口形状漂移 | Mock 字段名逐字对齐表设计（spec 第二节纪律）；阶段 B 切换时组件零改动是验收项 |
| R3 | 三版通用性：真实接口 Java 先行 | 阶段 A 纯前端+Mock，天然三版无差；阶段 B 沿用 2.0.1 R5 的声明式偏离 |
| R4 | 编译纪律：IDEA 运行中不编译 | 本阶段纯前端改动，不触发；阶段 B 时按 RuoYi-Vue/AGENTS.md 编译纪律执行 |
| R5 | 替换 2.0.0 冒烟页后旧入口行为变化 | 临时页本来就是过渡产物；2.0.0 checklist 已登记"2.1.0 正式化时重写"，本模块即该事项的落地 |
