# 2.0.3 AI 对话正式化：菜单权限收敛 + 会话参数 UI

> **状态：📝 spec 待确认（2026-10-11）**
>
> 三个收敛目标：①对话域接口从"登录态裸奔"收敛为**正式权限点**（对话功能本人可用、智能体管理按按钮授权）；②「AI 对话」进入**侧边栏菜单**（去掉浏览器敲 URL 的临时形态）；③**会话参数 UI**（temperature/max_tokens/max_contexts 在对话页可调，落库生效）。同时清理冒烟残留（已完成，见下）。

## 〇、冒烟清理（2026-10-11 已完成，本 spec 开工项）

- ✅ 后端 `AiSmokeController` 删除（`POST /ai/chat` 单轮端点，2.0.0 产物；javac 全模块 37 class 零错误确认无悬挂引用）
- ✅ 前端对话页 Mock 提示条移除（USE_MOCK=false 后无"演示数据"语义）
- ✅ `application.yml` 的 `spring.ai` 配置段**保留**（DashScope 自动配置占位 key 防启动失败，D4 依赖；openai 段为无 modelId 场景兜底，暂留）
- ✅ 前端静态路由 `/ai/chat` 删除（用户指令）——菜单动态路由（菜单 2020，component=`ai/chat/index`）是对话页唯一入口；无 `ai:chat:chat` 权限者菜单不可见/接口 403，语义收敛完整

## 一、目标与范围

**做**：
1. **权限点收敛**（后端）：
   - 对话功能（本人数据，无需按按钮授权）：`ai:chat:chat` 一个权限点覆盖会话 CRUD/消息/SSE/停止/智能体下拉/模型下拉——挂菜单 2020 的 perms（进入对话页即授权）；后端全部对话端点加 `@PreAuthorize("@ss.hasPermi('ai:chat:chat')")`；
   - 智能体管理（管理页 CRUD，本期先补权限点与接口面，管理页 UI 列 2.0.4）：`ai:agent:list/query/add/edit/remove` 五点已就位（2.0.2 B4），补 SQL 菜单按钮段（2030 号段）；
   - 归属校验兜底不变（getOwnedConversation——权限点是功能门禁，数据隔离靠归属校验，双层）。
2. **菜单**（SQL）：菜单 2020 补 perms=`ai:chat:chat`；admin 角色默认通配无需授权 SQL；2030-2035 智能体管理菜单+按钮（挂 2000 目录下，管理页 UI 本期不建，菜单 hidden=true 占位——避免出现点不开的菜单）。
3. **会话参数 UI**（前端 + 后端一个小端点）：
   - 对话页头部加"参数设置"弹窗（el-dialog + 三控件：temperature 滑杆 0~2 步 0.1 / maxTokens 数字 / maxContexts 数字，均可空=用模型默认）；
   - 新端点 `PUT /ai/conversations/{id}/params`（body: temperature/maxTokens/maxContexts，校验范围后落库，归属校验同 changeModel）；
   - 生效链路已存在（planner 读会话参数 → 覆盖模型默认），UI 只是数据入口；
   - ChatInput 的"参数"按钮 + 参数气泡回显（当前值提示）。

**不做**：
- 智能体管理页 UI（权限点与接口面本期就位，页面列 2.0.4——避免本模块摊大饼）；
- 会话参数的"系统级上限配额"（超配额拦截，属用量域配额管控，后续模块）；
- 深度思考开关 UI（enable_thinking 列已在快照里，UI 随智能体管理页做）。

## 二、现状事实（动手前核对）

| # | 事实 | 位置 |
|---|---|---|
| 1 | 对话域全部端点（会话 CRUD/SSE/停止/agents options）目前**无 @PreAuthorize**，仅登录态 | AiConversationController/AiChatController |
| 2 | `ai:agent:*` 五个权限点注解已在 AiAgentController（list/query/add/edit/remove），但菜单表无对应按钮 SQL | AiAgentController + sys_menu |
| 3 | 菜单 2020「AI对话」已存在（component=ai/chat/index），**perms 为空** | ai_20261012.sql:135 |
| 4 | AiConversation 实体 temperature/maxTokens/maxContexts 列已就位；planner 已消费（会话覆盖→模型默认回落） | AiConversation.java:64-70 + AiPromptPlanner |
| 5 | 会话更新端点现有 title/pin/model 三个，无 params 端点；service 层 changeModel 模式可平移 | AiConversationController |
| 6 | ChatInput 已有模型下拉（models prop + model emit），参数按钮挂同排 | ChatInput.vue:21-28 |
| 7 | 冒烟残留已清（AiSmokeController 删除 javac 零错、Mock 提示条移除） | 本 spec 第〇节 |

## 三、技术方案

### 3.1 权限点（Task 1-2）

```java
// AiConversationController / AiChatController 全部端点统一加：
@PreAuthorize("@ss.hasPermi('ai:chat:chat')")
// agents options 例外：保留登录态（对话页下拉用，权限由 ai:chat:chat 门禁；
// 若无 ai:chat:chat 则连 options 也拿不到——下拉在对话页内，语义自洽）
```

SQL（`ai_20261013.sql`，可重复执行）：
- update 菜单 2020 补 perms=`ai:chat:chat`；
- insert 2030「智能体管理」菜单（hidden=true, component=ai/agent/index 占位）+ 2031-2035 五个按钮（`ai:agent:list/query/add/edit/remove`）；
- 同步 `ai_20261012.sql`（2020 的 insert 语句补 perms）与 init SQL 无涉（菜单不进 init——AI 菜单归业务 SQL，2.0.1 既定）。

### 3.2 会话参数端点（Task 3）

```java
@PutMapping("/{conversationId}/params")
@PreAuthorize("@ss.hasPermi('ai:chat:chat')")
public AjaxResult updateParams(@PathVariable Long conversationId, @RequestBody AiConversation body)
// service: updateParams(conversationId, temperature, maxTokens, maxContexts)
//   归属校验（getOwnedConversation）→ 范围校验（temperature 0~2、maxTokens 1~100000、maxContexts 0~500，
//   超界抛 ServiceException）→ 三字段选择性更新（null 不动）
```

### 3.3 前端参数 UI（Task 4-5）

- `chat.js` + real 层补 `updateConversationParams(conversationId, params)`（Mock 同步补，语义对齐）；
- 对话页头部（模型下拉旁）加"参数"图标按钮 → el-dialog 三控件（el-slider + 两个 el-input-number，均带"清空=用模型默认"）→ 保存调契约函数 → 成功后刷新会话对象；
- 会话对象上回显当前值（对话框打开时填充表单）；
- 组件改动面：index.vue（状态 + 弹窗挂载）+ ChatInput.vue（emit params 按钮）——**MessageItem/MessageList/ConversationSidebar 零改动**。

## 四、风险与约束

| # | 风险/约束 | 应对 |
|---|---|---|
| R1 | 权限点上线后，已登录用户旧 token 的 permissions 列表不含 ai:chat:chat → 对话页 403 | 权限来自 Redis 会话（login_tokens），重新登录即刷新；验证流程含重登录 |
| R2 | 菜单 2020 补 perms 后静态路由 /ai/chat 与菜单动态路由并存——静态路由进页面但接口 403（无权限者） | 语义正确：静态路由是直达通道，无权限本就该 403；不属于要修的问题 |
| R3 | SSE 端点加 @PreAuthorize 的鉴权时机在 Flux 组装前（方法级），不影响流内异步 | 与 D5（闭包捕获身份）无冲突；端到端含 SSE 验证 |
| R4 | IDEA 运行中编译纪律 | 全程 javac 语法检查 + 移交用户 Build；端到端待 Build 后 |
| R5 | 参数 UI 校验前后端重复 | 前端控件限界（slider min/max）+ 后端范围校验兜底，两层都要（防绕过 UI 直调接口） |

## 五、验收标准（对应 checklist.md）

1. 无 `ai:chat:chat` 权限的用户（造测试角色验证）访问对话接口返回 403，有权限正常；
2. 菜单 2020 出现在侧边栏（AI管理 → AI对话），perms 生效；
3. `PUT /{id}/params` 落库生效：设 temperature=1.5 后发消息，planner 构建的请求参数变化（日志或行为验证），空值回落模型默认；
4. 参数 UI 三控件限界正确、保存回显正确；
5. 既有接口零影响（2.0.1 三管理页 + 2.0.2 对话链路回归）；契约新增不破坏（新端点纯增量）；
6. 测试数据清理；git diff 无密钥。
