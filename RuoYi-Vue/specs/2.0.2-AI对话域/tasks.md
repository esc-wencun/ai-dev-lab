# Tasks · 2.0.2 AI 对话域（阶段 A：前端 + Mock）

> 勾选纪律：做完即勾（含验证跑通）；没做不勾并注明原因。
> 依赖：无后端依赖（纯前端 + Mock）；复用 2.0.0 静态路由 /ai/chat。
> 环境约束：本阶段不碰 Java 代码、不碰数据库；前端 dev server 热加载验证。

## Task 1: Mock API 层
- [ ] `src/api/ai/chat.js`：接口签名（spec 第二节全量 9 个函数）+ Mock/真实开关（`USE_MOCK = true`）
- [ ] `src/api/ai/mock/mockChat.js`：内存存储（字段名逐字对齐对话域表设计）+ 假流式（30~80ms/片，总 2~4s）+ stop/regenerate 语义
- [ ] Mock 内置 1 个演示会话（含一条历史问答）方便打开即有内容看

## Task 2: 页面骨架与布局
- [ ] index.vue 双栏布局（左 280px 会话栏 + 右对话区），替换 2.0.0 冒烟页；静态路由不动
- [ ] ConversationSidebar：置顶/普通分组、lastMessageTime 倒序、hover 操作（重命名/置顶/删除）、新建会话按钮
- [ ] 空状态：无会话时引导新建；"演示数据不持久"提示条

## Task 3: 对话区与消息渲染
- [ ] MessageList：自动滚底 + 用户上滚暂停跟随 + "回到底部"浮钮
- [ ] MessageItem：user/assistant 气泡样式；msg_status 四态（0 生成中光标动效 / 1 正文 / 2 出错样式 / 3 已取消+半截灰化）
- [ ] ChatInput：Enter 发送 / Shift+Enter 换行；发送中按钮变"停止"（调 stopGenerate）

## Task 4: 全状态走查（Mock 驱动）
- [ ] 状态链实测：新建会话 → 发送（生成中动效）→ 流式增量渲染 → 完成
- [ ] 停止路径实测：生成中点停止 → 半截内容 + 已取消样式；重生成路径实测（旧回答不再进 Mock 上下文）
- [ ] 出错路径实测（Mock 提供强制出错开关）；刷新页面 Mock 重置提示
- [ ] 浏览器实操全过（dev server 热加载；截图或要点记录给用户评审）

## Task 5: 评审交接（阶段 A 收口）
- [ ] 向用户提交 UI 评审清单：交互是否合意 / Markdown 渲染加不加 / 模型选择器位置 / 会话内参数编辑要不要 / 智能体入口形态
- [ ] 评审结论回写 spec 第四节；**阶段 B 三件套细化后二次确认，才动后端**
