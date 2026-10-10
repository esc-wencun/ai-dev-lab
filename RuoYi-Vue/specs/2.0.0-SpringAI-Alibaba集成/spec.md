# 2.0.0 AI 功能基础设施：Spring AI Alibaba 集成（精简版，2026-10-10 瘦身）

> **状态：✅ 已完成（2026-10-08）；2026-10-10 应用户要求瘦身为单文件**——过程性记录（探针细节/Task 分解/验收清单）已删，保留长期生效的结论：版本选型依据、缺陷 D1-D4、遗留事项。代码即事实：依赖见根 [pom.xml](../../pom.xml)，模块见 [ruoyi-ai/](../../ruoyi-ai/pom.xml)，配置见 ruoyi-admin `application.yml` 的 `spring.ai` 段。
>
> 交付物：Spring AI 2.0.1 + Spring AI Alibaba 2.0.0-M1.1 依赖地基、`ruoyi-ai` 模块、冒烟端点 `POST /ai/chat`（`{prompt}` → `{code,msg,data:AI回答}`，登录态，无权限点）、前端临时对话页（`RuoYi-Vue3` 静态路由 `/ai/chat`，hidden 不进侧边栏，`src/api/ai.js`）。端到端全绿：编译/冷启动/登录冒烟/GLM 真实对话/401 与 500 边界/密钥零泄漏。

## 版本选型结论（2026-10-08 Maven Central 实查 + 独立探针工程实证）

- **Spring AI 2.0.1**：1.x 线只支持 Boot 3；Boot 4 线从 2.0.0 GA 起。选 2.0.1（首个补丁版）。
- **Spring AI Alibaba 2.0.0-M1.1**：唯一 Boot 4 线（1.1.x 线是 Boot 3）。里程碑版，其声明的 spring-ai 2.0.0-M1 / Boot 4.0.0 由我方 BOM 托管到 2.0.1 / 4.1.0，仲裁零冲突。其 BOM 只管 8 个构件，dashscope starter 须在根 pom 显式声明版本。
- **LangChain4j** 备而未用（对 Boot 宽松）；SAA 若再爆阻断缺陷按 D1 模式排除，仍不通则切换。
- GLM 网关走 Spring AI 官方 OpenAI starter（OpenAI 兼容协议），环境变量 `glm_api_key`/`glm_base_url`/`glm_model` 注入，密钥不入库不入 git。

## 缺陷与 API 变化（升级 SAA / 排查启动问题必读）

- **D1（阻断）**：`spring-ai-alibaba-autoconfigure-dashscope:2.0.0-M1.1` 打包缺陷——autoconfigure imports 声明了 `DashScopeMultimodalEmbeddingAutoConfiguration` 但 class 未入 jar，启动即抛 `Unable to read meta-data`。**解法**：application.yml `spring.autoconfigure.exclude` 排除该类。**SAA GA 后移除**。
- **D4（阻断）**：`DashScopeAgentAutoConfiguration` 急切创建 Bean 并强校验 API key（不用 DashScope 也绕不开）。**解法**：`spring.ai.dashscope.api-key: ${DASHSCOPE_API_KEY:sk-no-op}` 占位。**SAA GA 后移除**。
- **D2（API 变化）**：Spring AI 2.0 的 `ChatClient...options()` 改收 `ChatOptions.Builder`（不再收构建好的 ChatOptions）——1.x 示例代码会编译失败。
- **D3（须知）**：Boot 4 线 Spring AI 依赖中 Jackson 2（2.21.4）与 Jackson 3（`tools.jackson`）并存；与 RuoYi 既有 fastjson2 体系隔离，编译兜底已验。
- **Java 侧坑（控制器）**：`AjaxResult.success(answer)` 对 String 实参绑定到 `success(String message)` 重载，答案会进 `msg` 而非 `data`——须用双参 `success("操作成功", answer)`。

## 遗留事项

- SAA 出 GA 后：移除 D1 排除项与 D4 占位 key，升级版本号，回归启动 + AI 调用。
- `POST /ai/chat` 与前端对话页均为临时冒烟实现（单轮、非流式、不落库），**2.1.0 正式会话模块就位后重写或删除**。
- 前端对话页属探索转正（用户 2026-10-08 直接指令，静态路由不配菜单），2.1.0 正式化时迁入菜单体系并挂权限点。
