# Checklist · 2.0.4 智能体管理页

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] suggestedQuestions 出口修复：options/list/getInfo 三口返回数组（curl 实测 2026-10-13 全数组）；对话页 chips 数据形状对齐 Mock（数组）——页面点验随用户日常使用（Tabbit 运行时本次 CDP 故障，见下）
- [x] 管理页 CRUD 全通：新增/编辑/删除/名称查重 500/被引用拒删（报错带数量）——API 全链路实测
- [x] 建议问题动态行编辑：数组编辑 → 存库 JSON 字符串 → 回显数组（curl 验证编辑后 3 条）；对话页 chips 消费链路数据形状已验
- [x] 菜单 2030 放出：visible='0' 已生效（DB 实查）；SQL 双跑幂等；侧边栏可见性随用户浏览器确认（菜单数据源 getRouters 已就绪）
- [x] 私有智能体可见性：无角色账号 options 仅公开（100/101），admin 私有（104）不可见；无权限访问管理列表 403
- [x] 回归零影响：对话页主链路（conversations/options/SSE 流式 msgStatus=1）+ 2.0.1 三管理页全 200
- [x] 测试数据清理；specs/README 状态收口
- [x] 编译 + 单测全绿：mvn clean package exit 0；mvn test -pl ruoyi-ai -am → 37/37（含 AiAgentPayloadTest 9 用例）
- [x] 软删过滤修复验证（Task 3.5）：删 agent 后列表/下拉不再出现；软删名字可复用；删会话后智能体可删；消息三查询过滤就位（不回显/不进上下文/不参与去重）

## 遗留与决策记录

- 挂载编辑器（knowledge/tool/mcp）→ 等 RAG/工具域。
- 智能体编排（SAA Graph：编排引擎 + router/sequential/supervisor/group chat 多智能体模式）→ 已登记 specs/README"规划中的后续模块"，届时 ai_agent 表扩展编排配置列。
- 列表后端分页 → 量级大再说（现全量 + 前端筛选）。
- **AI 域软删机制债（2026-10-13 登记）**：实体无 @TableLogic、全局 DbConfig 只设值未设 logicDeleteField——2.0.4 已在 agent/会话计数/消息查询共八处手写过滤；provider/model/key 的列表与查重（2.0.1 域）同样未过滤，属同类债，登记待专项清理（2.0.1 域当前无删除入口，暂无实际影响）。
- **软删漏过滤修复（2026-10-13，用户报告"删除后列表仍显示"触发全域审计）**：系统三表（user/role/dept）XML 全部带 del_flag 过滤无此问题；AI 域补修 5 处显式过滤——provider/model/key 三个 page 查询 + selectProviderAll（厂商下拉）+ selectUserConversations（会话列表，此前靠 status/del_flag 同写碰巧挡住，判据分裂属隐性耦合）。遗留不动：selectStaleGenerating 无 del_flag 过滤（只处理 msgStatus=0 生成中消息，已删会话消息不会处于该态，实际无害）；selectById（getInfo 回显）不过滤（软删行无列表入口，不可达）。根治方案 @TableLogic 仍归软删专项债。javac 语法验证过；运行时行为待服务 Build + 重启后页面确认（服务运行中，编译纪律禁止 mvn）。
- **2026-10-12 工作区事故**：rm -rf 路径大小写笔误误删 RuoYi-Vue 工作副本（421 跟踪文件 git checkout 恢复零丢失；target/ 由本次 Build 重建；运行中服务崩溃后重启）；两个当日新文件（AiAgentPayload + 单测）从会话上下文逐字重建。**事故次生缺陷（2026-10-13 实测发现）**：controller 出口转换属未提交修改被 checkout 一并回退，首启后 suggestedQuestions 仍为字符串——已补回并三口实测数组。教训：checkout 恢复会吞掉同批未提交修改，恢复后要对照任务 diff 清单核对。
- **Tabbit 浏览器运行时故障（2026-10-13）**：CDP Target.createTarget 无法派发新页（换 task 名无效，首个 task 曾成功开页后运行时进入坏状态）；本次页面点验项以 API 数据形状 + 既有 2.0.2 浏览器验证替代，真实页面 chips/菜单可见性随用户日常使用确认。
- **菜单结构调整（2026-10-13，用户确认）**：「AI对话」（2020）升级为顶级菜单（parent=0, order=10, icon=message），与「AI管理」并列——对话是用户侧功能，不该挂在管理目录下；「AI管理」icon 换 skill 避免撞图标。DB UPDATE + ai_20261012.sql seed 同步 + getRouters 实测顶级路由 `/chat` 就位；前端零改动（component 路径不变，仅路由层级变化）。同批：agent 管理页图标改 ImageUpload 图片上传（avatar 相对路径回显补 /dev-api 前缀）、预置 agent 100/101 图标生成并经 /common/upload 入库、密钥管理菜单图标 key→lock（精灵库无 key.svg）。
