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

- [ ] ai_20261012.sql 三表落地（字段与设计文档逐字一致）+ 2 预置智能体 + 菜单；幂等执行验证
- [ ] 快照复制完整性：新建角色会话后，会话行 system_prompt/knowledge_ids/tool_ids/mcp_ids/enable_thinking 与智能体一致（SQL 抽查）
- [ ] AiPromptPlanner 单测全绿（裁剪/replyId 去重/summary 拼装）
- [ ] SSE 端到端：curl 断言 begin→token*→done 事件序列 + 消息落库（content/tokens/duration/msg_status=1）
- [ ] 三路落库：正常完成=1 / 构造出错=2 / 中途停止=3（保留半截）；孤儿清理生效
- [ ] 前端切真实 API 后**组件零改动**（git diff 仅 chat.js 与新增依赖）
- [ ] Markdown 渲染 + sanitize 生效（assistant 气泡；XSS 注入样例被转义）
- [ ] 模型选择器：切换后新消息走新模型（消息表逐条 model_id 可查）
- [ ] 浏览器全链路实操（角色/开场白/chips/流式/停止/重生成/切模型）通过
- [ ] 既有接口零影响 + 2.0.1 三页回归 + 测试数据清理 + specs/README 登记 2.0.2

## 遗留与决策记录

- Markdown/代码高亮依赖：UI 评审后定（spec 一节"不做"）。
- 阶段 B 细化时的必答问题清单见 spec 第四节。
