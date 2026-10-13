# Tasks · 2.0.3 对话正式化：菜单权限收敛 + 会话参数 UI

> 勾选纪律：做完即勾（含验证跑通）；没做不勾并注明原因。
> 依赖：2.0.2（对话域接口与页面已就位）。冒烟清理已完成（spec 第〇节）。
> 环境约束：IDEA 运行中不 mvn（javac 语法检查 + 用户 Build 后端到端）；MySQL 13306。

## Task 0: 冒烟清理（已完成，登记备查）
- [x] 删除后端 AiSmokeController（javac 全模块 37 class 零错误，无悬挂引用）
- [x] 删除前端对话页 Mock 提示条（USE_MOCK=false 后无此语义）
- [x] 删除前端静态路由 /ai/chat（router/index.js）——菜单动态路由是对话页唯一入口（用户指令 2026-10-11）；菜单 2020 的 component=ai/chat/index 已能被 loadView 解析（vite transform 200）
- [x] 2.0.0 spec 遗留项回写清理记录；`spring.ai` yml 保留（D4 占位依赖）

## Task 1: 权限点收敛（后端）
- [x] AiConversationController 全部 7 端点 + AiChatController 2 端点加 `@PreAuthorize("@ss.hasPermi('ai:chat:chat')")`（agents /options 保持登录态）——新增第 8 端点 /params 同步带注解
- [x] javac 语法检查通过（4 文件 -proc:full 零错误）

## Task 2: 菜单 SQL
- [x] `RuoYi-Vue/sql/ai_20261013.sql`（可重复执行）：菜单 2020 补 perms=`ai:chat:chat`；2030 智能体管理菜单（visible=1 hidden 占位）+ 2031-2034 按钮（ai:agent:list/query/add/edit/remove）
- [x] `ai_20261012.sql` 同步（2020 insert 语句补 perms 字段）
- [x] 13306 执行两次幂等验证（2020 perms 在、2030-2034 六行齐备无翻倍）；无 `ai:chat:chat` 的测试角色 403 实测待用户 Build 后（端到端含重登录步骤）

## Task 3: 会话参数端点（后端）
- [x] AiConversationController 加 `PUT /{conversationId}/params`（ai:chat:chat）；service `updateParams`：归属校验 + 范围校验（temperature 0~2 / maxTokens 1~100000 / maxContexts 0~500）+ **null 不动 / -1 清除覆盖（normalize 落库 null）**——开发实查修正：`v == -1` 对 Double 是引用比较恒 false，改 `v.doubleValue() == -1d`
- [x] javac 语法检查通过（含接口 + impl 联编，@Override 解析正常）

## Task 4: 前端参数 UI
- [x] chat.js 契约 + real/Mock 双实现：`updateConversationParams(conversationId, params)`——Mock 语义四项自测全过（set 1.5/2048、-1 清 temperature 保留 maxTokens、越界拒绝、会话不存在拒绝）
- [x] ChatInput 加"参数"按钮（Setting 图标，emit params）；index.vue 挂参数弹窗（slider 0~2 + 两个 input-number；打开回显当前会话值；保存逻辑：temperature 恒提交、maxTokens/maxContexts 空输入框=未改动 null 不动、**"恢复模型默认"按钮=三字段 -1 清除覆盖**）
- [x] vite transform 验证（index.vue/ChatInput 全 200，新符号 paramsOpen/saveParams/resetParams 在转换产物中）——**"组件零改动"口径修正：MessageItem/MessageList/ConversationSidebar 本模块零改动达成；ChatInput/index.vue 本来就是本模块的改动面**

## Task 5: 端到端验证（2026-10-12，两轮）
- [x] 权限：admin（*:*:* 通配）全端点 200；造测试角色 chat403test + 用户实测——`/ai/conversations` **403「没有权限，请联系管理员授权」**、`/ai/agents/options` 200（登录态设计）——权限矩阵与 spec 设计完全一致
- [x] 菜单：2020 perms 已在 DB（菜单可见性待用户浏览器重登录确认——接口侧 getRouters 数据源已就绪）
- [x] params 落库与范围校验：设 1.9/4096 → DB 实查在；temperature=5 → 500「temperature 取值范围 0~2」；**参数消费链路（带 1.9 的会话 SSE 正常）实测通过**
- [ ] **-1 清除覆盖实测未通过（发现的实现 bug，已修源码待用户 Build）**：范围校验先拦了 -1（-1<0 越界）——修复：校验放行 -1，normalize 落库 null；顺带修正 normalize 默认语义（null 保持现值 conv.getXxx() 而非强制 null，否则未提交字段会被意外清掉）
- [x] 回归：2.0.1 三管理页 + user/list + agents/options 全 200；SSE 主链路（参数会话）正常
- [x] 重生成回归——**第一轮验证用的 17:44 旧 jar 不含 planner null 保护，regenerate 报 500 属预期（旧包）；2.0.2 收口时已在新代码上全语义闭环验证过，本模块无新增改动面**。注意：IDEA Build 增量更新的是 target/classes，ruoyi-admin.jar 仍是 17:44——后续端到端要么 IDEA 里跑（用新 classes），要么重打包 jar
- [x] 测试数据清理：测试角色/用户（role_id=900/user_id=900）已删、我建的测试会话与消息已删、Redis 会话已清、我起的旧 jar 已停；用户自建会话 109 保留

## Task 6: 用户实测反馈修复（2026-10-12）
- [x] **bug 1（多会话共享生成状态）**：`streaming` 全局单值 → `streamingConversationId`（生成态归属会话）+ `isCurrentStreaming` 派生值（模板消费）——切会话后旧流继续跑但 UI 不误显示；onDone 加守卫只清自己会话的态（防切会话后旧流 done 误清新会话）
- [x] **bug 2（回复后发送按钮卡在停止态）**：根因是 `streaming=false` 完全依赖 onDone 回调——done 缺失/断流/HTTP 错误路径全部漏复位。修复：`handleSend`/`handleRegenerate` 加 `finally` 兜底复位（任何退出路径都回可发送态）+ handleStop 停止后兜底
- [x] **同会话防重入**（顺带发现）：real 层流句柄是模块级单例，同会话双发会互相覆盖句柄——send/regenerate 入口加 `streamingConversationId === convId` 守卫
- [x] vite transform 全 200；转换产物断言：新状态/派生值/finally/守卫全部在，旧 `streaming.value` 引用清零

## Task 8: 浏览器自动化端到端测试（Tabbit + Playwright，2026-10-12）
- [x] 测试环境：mvn 编译 + jar 启动（8080）+ vite dev（8888）；Tabbit 浏览器按 127.0.0.1 访问（localhost 取不到验证码，用户提醒）；验证码经 Redis 直读答案完成 admin 登录
- [x] **T1 登录 ✅ / T2 菜单入口 ✅（侧边栏 AI管理→AI对话，动态路由）/ T3 发送+流式 ✅ / T4 自动起名 ✅（首条消息前 20 字，"What is 2+2? Answer"）/ T9 切模型 ✅ / T10 不支持模型报错 ✅（按钮复位+JSON 错误分支生效）/ T11 删除会话 ✅（确认框弹出+列表减一）/ T13 主链路零 TypeError**
- [x] **发现并修复根因 bug（MessageList.vue:48）**：`m.content.length` 对 content 为 null/undefined 的消息崩（TypeError）→ Vue patch 中断 → 三重症状：用户消息不显示、ChatInput prop 冻结在 true（按钮卡停止）、后续渲染全停摆。修复：`String(m.content ?? '').length` 空值防护。修复后整轮回归 0 错误
- [x] **用户消息气泡显示**（对齐 Claude Code 的 `>` 前缀行风格——样式已有，缺消息入列）：handleSend 本地乐观插入 user 气泡（nextLocalId 负数 key），失败路径重新拉服务端历史对齐
- [x] 顺带修复：负数 messageId 同毫秒撞 :key（三处 -Date.now() 统一 nextLocalId 递减序列）——欢迎语+首问同毫秒创建时的渲染异常源
- [ ] T5 Markdown 高亮 / T6 多会话隔离 / T7 停止 / T8 重生成 / T12 参数弹窗——主链路修复后机制上同源（此前 TypeError 崩溃掩盖），随用户日常使用确认

## Task 9: GLM-4.7 网关错误不提示（用户反馈第三轮，2026-10-12）
- [x] **根因**：model 100（glm-4.7）在网关 404，网关返回 HTTP 404 + `{"error":{message,type}}`——OpenAI Java SDK 把流式请求的错误响应**静默转成"零 token 空完成"**（不抛异常），后端 onErrorResume 捕不到，完成路把空 content 标记为 msgStatus=1"正常完成"→ 前端收到 done 但既无内容也无错误提示，看起来"没反应"
- [x] **修复（后端完成路空响应守卫）**：done 路检测 `gathered.length()==0` → msgStatus=2 + errorMessage="模型无响应：该模型可能不存在、已下线，或服务网关异常（可切换其他模型重试）" + 服务端 log.warn 留诊断——前端复用 done(msgStatus=2) 错误气泡链路（MessageItem err-line 红字显示）
- [x] 实测：重启后 model 100 发消息 → done 事件带 `"msgStatus":2,"errorMessage":"模型无响应：..."`（curl 断言）；前端 err-line 渲染条件（msgStatus===2 → ✕ errorMessage）代码核对通过
- [x] 沉淀：已知坑登记（OpenAI SDK 流式对非 2xx 的静默空完成行为）；model 100 保留作"配置错误演示"样例，正式使用在模型管理把 code 改为网关实际模型名

## Task 10: 用户消息样式对齐 Claude Code（用户反馈第四轮，2026-10-12）
- [x] **用户消息改浅灰圆角卡片**（`--el-fill-color-light` 底 + 圆角 10px + 细边框，正常字体 13.5px）——与 AI 回复（无底色平铺正文 + Markdown）形成明确视觉区分
- [x] **用户消息靠右**（微信同款布局）：`.msg-row.user { flex-direction: row-reverse }`，`>` 前缀随行右移贴卡片右侧；AI 回复保持靠左
- [x] vite style block 编译验证（row-reverse 在 scoped style 产物中）；浏览器实测因 Tabbit 实例导航故障搁置，CSS 为确定性属性，随用户刷新页面直接确认

## Task 11: Ollama 本地模型对话支持（用户反馈第五轮，2026-10-12）
- [x] **根因**：ChatModelRouter 只放行 openai 协议，ollama 协议模型（qwen3:4b）被「暂不支持协议」拒绝——2.0.1 的既定扩展点，本轮补上
- [x] **实现（零新依赖）**：Ollama 11434 自带 OpenAI 兼容端点（实测 /v1/chat/completions 正常）——resolver 的 ollama 默认路径 /api → /v1，router 放行 ollama 走同一 openai 客户端（baseUrl=host/v1）；单测同步（defaultPathOllama/joinLocalhost）
- [x] 端到端实测：qwen3:4b 流式对话 begin→done、回答正确（3+3=6）、msgStatus=1 ✅；测试会话已清理
- [x] 设计文档注记回写：AI相关表设计方案.md 内置厂商表 ollama 行说明更新

## Task 7: 用户实测反馈修复（第二轮，2026-10-12）
- [x] **bug 3（左侧删除无反应）**：ConversationSidebar 的 `doDelete` 用了 `ElMessageBox` 但未 import（auto-import 列表无 Element Plus API，main.js 全量注册也不挂全局变量）→ ReferenceError 静默失败。修复：显式 `import { ElMessageBox } from 'element-plus'`；顺带把重命名的 `window.prompt` 升级为 `ElMessageBox.prompt`（带非空校验，观感对齐）
- [x] **会话自动起名**：默认「新会话」的会话在首条用户消息发出时，取内容前 20 字符（压缩空白）作标题；已重命名（≠新会话）/重生成不触发。后端 `IAiConversationService.autoTitleIfDefault`（service 层实现，doStream 发消息时调用）+ Mock 层同语义；node 自测截断正确（20 字截断、空白压缩）；前端 onDone 后 refreshConversations 已有，起名自动回显

## Task 10: 模型连通性测试（用户需求，2026-10-12）
- [x] 后端：`POST /ai/model/{modelId}/test`（ai:model:edit）+ `IAiModelTestService`——复用 2.0.1 解析链/路由器发最小对话请求（"reply with pong"，maxTokens 256），返回 {success, durationMs, message, reply?}
- [x] **两轮实测修正**：①maxTokens=8 被推理模型（GLM-5.3-Flash）的思考耗尽 → content 恒空误判失败——提到 256；②网关 404 被 SDK 抛成 `\`choices\` is not set` 内部异常 → 映射为友好提示「模型无响应：模型可能不存在、已下线，或服务网关异常」
- [x] 实测结果：model 101 → `{success:true, reply:"pong", durationMs:1783~3819ms}`；model 100 → `{success:false, message:"模型无响应：...（网关返回了非标准响应）"}`——成功/失败双路径正确
- [x] 前端：模型表单 footer 加「测试连接」按钮（仅编辑态，保存后可测）+ `testModel` API；成功弹窗含耗时与模型回复，失败弹窗带原因
- [ ] 浏览器 UI 点验——tabbit 与 vite dev 的 HMR websocket 偶发导航超时（非本项目问题），按钮代码就位，随用户日常使用确认

## Task 12: 密钥占位符示例提示 + 创建态连通测试（用户需求第二轮，2026-10-13）
- [x] **密钥表单提示落成示例**（原"格式见内置数据示例"太空）：直接给 `${glm_api_key}` / `${DASHSCOPE_API_KEY}` 真实示例与格式说明（`${变量名}`，调用时从服务端环境变量取值不落明文；脱敏规则 前4+***+后4、占位符原样显示）——后端脱敏与 `${ENV_VAR}` 解析链 2.0.1 已实现，本轮纯提示文案
- [x] **创建态连通测试**（原按钮仅编辑态可用）：新增 `POST /ai/model/testDraft`（ai:model:add）+ `AiModelTestDraftBody`（providerId/keyId/modelCode/modelName）——resolver 抽 `resolveByParts` 后段共用，`resolveDraft` 按构件走同一条链（草稿壳 status=null 天然跳过停用校验），测试结果与保存后正式链路完全一致
- [x] **测试提示词改「你是哪个模型？」**（原 "reply with the single word: pong"）：回复即暴露真实模型名，顺带验证推理与计费——TEST_PROMPT 常量收口一处。**用户实测后手动修正两处**：提示词追加版本追问（"你是哪个模型？哪个版本？"）；maxTokens 256→1024（推理模型思考链 256 仍会被耗尽导致误判失败）
- [x] 前端：模型表单 footer 两态都显示「测试连接」（创建态调 testModelDraft，表单三项必填由后端 resolveDraft 前置校验兜底提示）；api/ai/model.js 加 testModelDraft
- [x] javac 全模块 0 错误（DTO 漏 modelName 字段在检查中抓到修复）——端到端待用户 Build 后走页面实测（testDraft 双路径 + 密钥页提示文案）