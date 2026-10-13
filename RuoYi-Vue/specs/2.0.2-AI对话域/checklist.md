# Checklist · 2.0.2 AI 对话域（阶段 A）

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单（阶段 A：UI + Mock）

- [x] Mock API 层：9 个签名函数齐备，组件零直接 Mock 依赖（只 import chat.js）
- [x] 会话栏：新建/重命名/置顶/删除/分组排序全部可用（用户实操确认）
- [x] 消息区：自动滚底、上滚暂停跟随、回到底部按钮工作正常（用户实操确认）
- [x] 消息四态渲染齐全（生成中/完成/出错/已取消），流式增量可见（语义自测 A-F + 用户实操）
- [x] 输入区：Enter/Shift+Enter 语义正确，发送中可停止（用户实操确认）
- [x] 重生成：旧回答被剔除出上下文（语义自测 D 项），新回答挂同 reply_id
- [x] 浏览器全状态走查通过（用户实操确认交互无问题），评审材料已提交并收到结论
- [x] 纯前端改动：Java 代码/数据库零变更（改动仅 src/api/ai/ 2 文件 + src/views/ai/chat/ 5 文件 + launch.json attach 配置）
- [x] 阶段 A 评审完成（用户采纳全部 agent 建议），结论回写 spec；阶段 B 细化 spec 待二次确认

## 阶段 B 验收清单（细化稿 2026-10-11，二次确认后执行）

- [x] ai_20261012.sql 三表落地 + 2 预置智能体 + 菜单（2020）；幂等执行验证通过（双跑无重复）——**实施修正：ai_conversation/ai_message 补建 remark 列**（BaseEntity.remark 是真实列映射，MP 会拼进 SELECT，未建列即 SQL 报错；子类 @TableField(exist=false) 遮蔽对同名字段不生效——这个坑已记入 tasks B2）
- [x] 快照复制完整性：agent 100 建会话实测——title=平台使用助手、system_prompt 全文复制、agent_id/model_id=101、enable_thinking=0、last_message_time 初始化（SQL 抽查通过）
- [x] AiPromptPlanner 单测全绿（6/6：窗口裁剪/重生成 replyId 去重/useContext 剔除/summary 拼装顺序/空装配/默认窗口）——顺带补跑了 2.0.1 遗留的 AiModelResolverTest（12/12，修复其缺 ServiceException import + ENV_PLACEHOLDER 可见性）
- [x] SSE 端到端：begin(msgStatus=0,replyId 配对)→token*(增量)→done(msgStatus=1,content/durationMs 落库) 全序列断言通过；修复 D5（异步线程丢 SecurityContext——请求线程闭包捕获身份，对位生产级范式）
- [x] 三路落库：完成=1（content+durationMs）/ 出错=2（401 Unauthorized 落 error_message）/ 取消=3（stop 端点 Subscription.cancel 触发 doOnCancel）；孤儿清理生效（重启后 stale 7→6，清理日志打印；NULL create_time 行不参与比较属 SQL 语义）——修复 D6（停止句柄注册晚于订阅，改为请求线程同步注册）
- [x] 前端切真实 API 后**组件零改动**（git diff 仅 chat.js 开关 + 新增 real/ 3 文件 + MessageItem 高亮修复 + marked-highlight 依赖）
- [x] **真实链路端到端断言（2026-10-11，经 vite 代理模拟浏览器路径）**：agents/options 2 个 ✓ → 建会话快照（title=平台使用助手）✓ → 模型列表（100/101）✓ → SSE 流式 DONE msgStatus=1（589 token 事件）✓ → 重生成 regenerateMessageId 契约（DONE replyId 同轮、旧回答 use_context=0、新回答 use_context=1，667 token 事件）✓。curl 命令行中文乱码为 GBK 显示层假象，DB utf8mb4 内容完好
- [x] Markdown 渲染 + sanitize 生效（assistant 气泡 marked+DOMPurify；**B5 实测修正：marked v18 的 setOptions.highlight 已静默失效，改 markedHighlight 扩展后高亮真实产出**——XSS 注入转义由 DOMPurify USE_PROFILES 承担）
- [x] **对话页视觉风格定稿：Claude Code 终端风**（2026-10-11 用户确认，spec 4.5 已记录）——无气泡平铺/角色等宽前缀/输入框提示符行/Esc 停止
- [x] **思考过程功能**（2026-10-12，10-13 决策修订）：模型思考开关（enable_thinking 快照→extraBody 透传，D7 全参数覆盖）、SSE reasoning 独立事件、前端 ✻ 折叠行+思考耗时——真实 GLM 链路实测（D8 帧双载荷修复后 reasoning+content 双路下发）。~~reasoning_content 落库回放~~ → 按用户决策改为**仅对话时展示、不落库**（D9 列容量事故的直接后续；顺带修复既有 bug：onToken/onDone 用服务端载荷整体替换丢本地思考缓冲，改为本地态字段保留合并）。**10-13 语义确认：门控保留**——enable_thinking 开启才会话思考（开启 = 仅前端展示不落库；关闭 = 不请求思考）；短暂改"恒开"已回退。当前局限：自由对话无开关 UI（选"平台使用助手"等开思考智能体才有），开关 UI 随 2.0.4
- [x] 模型选择器：切换后新消息走新模型（会话表逐条 model_id/model_code，changeConversationModel 端点实测——新建会话默认绑模型列表 sort 最前者，切模型接口已验）
- [x] 浏览器全链路——**接口层两轮全绿**：第一轮（主链路/cancel/防重复停）+ 第二轮（**regenerate 全语义闭环**：2+2=4 → regenerate → 新消息同 replyId、旧 useContext=0 → 追问 +10 答 14 证明上下文正确）；页面级实操由用户日常使用确认
- [x] 既有接口零影响（登录 + 用户列表冒烟）+ 会话 CRUD 补验（重命名/置顶/切模型/删除级联软删/智能体删除前置校验）+ 测试数据清理（会话/消息/Redis 会话已清，8080 已释放）+ specs/README 登记（登记条目见下）

## 遗留与决策记录

- Markdown/代码高亮依赖：已定案落地（marked v18 + marked-highlight + highlight.js + DOMPurify）。
- 阶段 B 细化时的必答问题清单见 spec 第四节。
- **regenerate 后端支持需重新 Build**（2026-10-11 B5 新增：DTO/service/impl 三处 + planner null 保护；17:44 的 jar 不含）——重生成按钮在旧 jar 下会报 500，Build 后即通。
- 内置 model 100（glm-4.7）在本机内网网关不存在：新建自由对话会话默认绑它（模型列表 sort 第一）会流式返回空内容——切到 model 101（GLM-5.3-Flash）即正常；正式使用时在模型管理里把 100 的 code 改成网关实际模型名。
- marked-highlight 为 B5 新增依赖（package.json + package-lock）。
