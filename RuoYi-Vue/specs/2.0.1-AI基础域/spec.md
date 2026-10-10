# 2.0.1 AI 基础域：厂商 / 密钥 / 模型管理

> **状态：📝 spec 待确认（2026-10-10）**
>
> AI 演进序列第二个模块（依赖 2.0.0 的 ruoyi-ai 模块与依赖地基）。目标：把 2.0.0 里写死在 application.yml 的单模型配置，升级为 **DB 配置驱动的三表管理体系**（`ai_provider` / `ai_api_key` / `ai_model`），从此新增厂商/换 Key/换模型 = 插数据，不改代码不发版。
>
> 表设计唯一权威源：[AI相关表设计方案.md](../AI相关表设计方案.md) 第一部分（已过逻辑漏洞审查，含可用性校验链与路径拼接约定）。本 spec 只写实现任务，不重复表设计。

## 一、目标与范围

**做**：
- 基础域三张表落库（`ai_provider` / `ai_api_key` / `ai_model`，含内置厂商数据 + 字典）；
- 厂商/密钥/模型三个管理页面的后端全套（entity/mapper/service/controller，MP 分页 + BaseService 唯一性校验）；
- **模型可用性解析器**（`AiModelResolver`）：按 model_id 取数 + 第 0 步可用性校验 + `${ENV_VAR}` 解析 + 两级地址拼接——2.1.0 对话模块从这里拿"可调用的模型+地址+Key"；
- AI 管理菜单（sys_menu 三页 + 按钮权限）+ 两个字典（模型类型/状态复用 sys_normal_disable）；
- 前端三个管理页面（RuoYi-Vue3，`views/ai/` 下）；
- `POST /ai/chat` 冒烟端点改造为按 model_id 调用（验证解析器端到端可用）。

**不做**（后续模块再议）：
- 对话域三表（ai_agent/ai_conversation/ai_message，2.1.0）；
- 密钥加密存储（本期 `${ENV_VAR}` 占位 + 管理页脱敏展示，真加密等有需求）；
- Key 定时探活/到期自动停用（预留扩展）；
- SSE 流式（2.1.0）。

## 二、现状盘点（动手前核对过的事实）

| # | 事实 | 位置 |
|---|---|---|
| 1 | 2.0.0 已建 ruoyi-ai 模块（依赖 ruoyi-common + Spring AI 2.0.1 + SAA 2.0.0-M1.1），含临时 AiSmokeController（`success("操作成功", answer)` 双参信封） | [ruoyi-ai/pom.xml](../../ruoyi-ai/pom.xml)、AiSmokeController.java |
| 2 | MP 全家桶已就位：BaseMapper + LambdaQueryWrapperX（eqIfPresent 族）+ PageUtils.buildPage() + `getDataTable(IPage)` 重载 | 0.0.0 spec 沉淀，ruoyi-common |
| 3 | BaseService.checkUnique(keyField, keyValue, field, value) 公用唯一性校验已存在 | ruoyi-common/core/service/BaseService.java:24 |
| 4 | sys_menu 最大 ID **1060**，AI 菜单从 **2000** 起（百位预留段，避开官方增量）；insert 列序 `(menu_id, menu_name, parent_id, order_num, path, component, query, route_name, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)` | ry_20260417.sql |
| 5 | 字典插入列序 `(dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, create_by, create_time, update_by, update_time, remark)`；dict_code 自增不指定 | ry_20260417.sql |
| 6 | Docker init SQL 只在空卷执行：**本机数据卷已有数据，AI 建表 SQL 不会自动执行**——须手动对 13306 实例执行一次，并同步进 `docker/mysql/init/01-ruoyi-schema.sql`（新环境开箱即用） | docker/AGENTS.md 约束 1 |
| 7 | GLM 环境变量（glm_api_key/glm_base_url/glm_model）本机可用——内置数据将把当前 yml 配置迁移为一条厂商+Key+模型记录 | 根 AGENTS.local.md |
| 8 | 官方 SQL 无 `database` 前缀、utf8mb4；`sys_normal_disable` 字典已有（0 正常 1 停用），AI 表状态列直接复用 | ry_20260417.sql |

## 三、技术方案

### 3.1 数据库变更（Task 1）

新建 `RuoYi-Vue/sql/ai_20261010.sql`（独立文件，不混入官方 ry_*.sql），内容：

1. 三张表 CREATE TABLE（从 AI相关表设计方案.md 第一部分第五节原样搬运，含 `uk_provider_key_model`、api_host NOT NULL 等审查修正）；
2. 内置数据：
   - 5 家厂商（DeepSeek/智谱/百炼/硅基流动/Ollama，api_host 按设计文档内置数据建议表）；
   - 智谱厂商下一把 Key：`api_key = '${glm_api_key}'`（环境变量占位，本机环境变量已配置）+ 百炼厂商下一把占位 Key（`${DASHSCOPE_API_KEY}`）+ Ollama 占位 Key（`local-no-key`）；
   - 智谱下一条对话模型记录：`model_code = glm-4.7`（对齐本机环境变量 glm_model 默认值）、type=1、sort_order=1——**迁移 application.yml 的 glm 配置为 DB 数据**；
3. 菜单 SQL：`AI 管理` 一级目录（menu_id=2000）+ 厂商/密钥/模型三个页面菜单（2001-2003，perms 各配 `ai:provider:list` 等）+ 每页 query/edit/remove 四按钮（2004-2015，perms `ai:xxx:add` 等）——**给 admin 角色（role_id=1）无需授权（admin 通配），不插 sys_role_menu**；
4. `delete` 语句前置（可重复执行）。

同步：把三张表的 CREATE TABLE 段**追加**进 `docker/mysql/init/01-ruoyi-schema.sql`（新环境开箱即用；菜单/字典/内置数据不同步——init SQL 保持与官方 schema 对齐，AI 内置数据属于业务初始化，由 ai_20261010.sql 承担）。**本机验证流程：docker exec 手动执行 ai_20261010.sql → 不动数据卷。**

### 3.2 后端结构（Task 2-4，`ruoyi-ai` 模块内）

```
ruoyi-ai/src/main/java/com/ruoyi/ai/
├── domain/          AiProvider / AiApiKey / AiModel（MP 注解：@TableName/@TableId(AUTO)/
│                    @TableField(exist=false)；继承 BaseEntity；AiModel.modelDimension 仅 type=5 校验）
├── mapper/          AiProviderMapper / AiApiKeyMapper / AiModelMapper（extends BaseMapper +
│                    default selectXxxPage(PageUtils.buildPage() + LambdaQueryWrapperX ifPresent 族)）
├── service/         IAiProviderService / IAiApiKeyService / IAiModelService + impl/
│                    （BaseService 子类：checkUnique 校验 provider_code/api_key+provider/model_code+provider+key；
│                     删除前置校验：厂商下有 Key 拒删、Key 下有模型拒删——对应设计文档坑 5；
│                     密钥编辑禁改 provider_id——对应坑 3）
├── resolver/        AiModelResolver（本模块核心新组件，spec 3.4）
└── controller/      AiProviderController / AiApiKeyController / AiModelController
                     （@PreAuthorize ai:provider:xxx 族；AiApiKeyController 返回列表时 api_key
                      脱敏为前 4+后 4 位，编辑时后端原样保存、明细接口不回显完整 Key）
```

**service 层返回约定**：列表/明细走 MP 查询；`api_key` 字段在 controller 出口脱敏（`gl***key`），新增/编辑入参原样入库（支持 `${ENV_VAR}` 占位原样保存）。

### 3.3 权限点

| 菜单/按钮 | perms |
|---|---|
| 厂商管理 | `ai:provider:list / query / add / edit / remove` |
| 密钥管理 | `ai:key:list / query / add / edit / remove` |
| 模型管理 | `ai:model:list / query / add / edit / remove` |

冒烟端点 `/ai/chat` 维持登录态即可（不加权限点）。

### 3.4 模型可用性解析器 `AiModelResolver`（核心组件）

```
resolve(modelId) → ResolvedAiModel { provider, apiKey(解析后), model, baseUrl(完整), }
1. 查 ai_model（不存在/软删 → 业务异常）
2. 查 ai_api_key（同上）；查 ai_provider（同上）
3. 可用性校验（设计文档解析链第 0 步）：
   model.status=0 && key.status=0 && provider.status=0 && (key.expire_time 为空 || 未过期)
   任一不过 → 抛业务异常，消息带明确定位（如「模型已停用」「密钥已过期」）
4. 一致性断言：key.providerId == model.providerId（数据被人手工改库的兜底校验）
5. 地址拼接（设计文档解析链第 4 步 + 路径拼接约定）：
   host = key.base_url ?? provider.api_host（两者都空 → 配置异常）
   path = key.api_path ?? 协议默认路径（openai=/v1/chat/completions，ollama=/api/chat；custom 必须显式配置 api_path）
   baseUrl = host 去尾斜杠 + path
6. ${ENV_VAR} 解析：api_key 匹配 ^\$\{(\w+)\}$ → System.getenv；匹配不上环境变量 → 配置异常
   （集中一处，拒绝静默回退——设计文档坑 2）
```

**单测**（不连库）：地址拼接矩阵（host 带/不带尾斜杠 × path 有/无 × 双 /v1 防御）、`${VAR}` 解析成功/变量不存在、一致性断言触发、状态位四种组合。连库逻辑（三查 + 校验）端到端覆盖。

### 3.5 冒烟端点改造（Task 5）

`AiSmokeController.chat(body)`：入参从 `{prompt}` 扩为 `{modelId, prompt}`（modelId 必填）——调 `AiModelResolver.resolve(modelId)` 拿 baseUrl/apiKey/model_code，用 Spring AI 的 `OpenAiChatOptions.builder().baseUrl()/apiKey()/model()/completionsPath()` **请求级覆盖**（不动全局 OpenAiAutoConfiguration，application.yml 的 spring.ai 配置保留为无 modelId 时的兜底）。GLM 迁移到 DB 数据后，yml 里的 `${glm_*}` 段保留（兜底 + 2.0.0 兼容）。

**验证路径**：DB 建智谱模型记录 → `/ai/chat` 传 modelId → GLM 真实应答（与 2.0.0 直连 yml 等效，证明解析器全链路通）。

### 3.6 前端三页（Task 6，RuoYi-Vue3）

`src/api/ai/`（provider.js / modelKey.js / model.js）+ `src/views/ai/provider|key|model/index.vue`——标准 CRUD 页模式（列表+搜索+对话框表单，对齐 system/post 页结构）：厂商页（含图标外链展示）、密钥页（api_key 密码框 + 脱敏展示 + `${ENV_VAR}` 占位提示）、模型页（类型字典渲染 + 能力位开关组 + 只读维度列）。

## 四、风险与约束

| # | 风险/约束 | 应对 |
|---|---|---|
| R1 | 本机数据卷已初始化，init SQL 不会自动执行新表 | ai_20261010.sql 手动执行（docker exec），同步更新 init SQL 供新环境 |
| R2 | 菜单 ID 冲突 | 从 2000 起（官方数据最大 1060，增量升级通常 <2000；冲突时执行报错可见，改号重跑） |
| R3 | Spring AI 请求级覆盖 baseUrl/apiKey 的 API 在 2.0 的确切形态未验证 | Task 5 第一步先写探针式单测/临时端点验证；若 OpenAiChatOptions 不支持请求级 baseUrl，退化为按 (baseUrl,key) 缓存 ChatModel 实例的 Map 路由（对位后续多模型工厂的雏形） |
| R4 | IDEA 运行中不可编译（编译纪律） | 实施期检测 8080/idea64.exe，命中则语法检查 + 移交用户编译验证 |
| R5 | 三版通用性：Python/Go 版尚未有 AI 域 | 本期只做 Java 版 + 前端；前端页面依赖 `/dev-api` 有对应端点，Python/Go 接入前菜单可通过 sys_menu 的 status 停用，或容忍 404（学习项目单端演进常态） |
| R6 | api_key 明文落库 | 本期接受（学习项目 + `${ENV_VAR}` 通道已备）；管理页脱敏展示 + 明细不回显兜底；真加密登记为预留扩展 |

## 五、验收标准（对应 checklist.md）

1. ai_20261010.sql 在 13306 实例可重复执行（先 delete 后 insert）；
2. 三个管理页面增删改查全通（含唯一性校验、删除级联拒绝、密钥脱敏）；
3. `/ai/chat` 传 modelId 走 DB 配置真实调用 GLM 应答；
4. AiModelResolver 单测全绿（拼接矩阵 + 环境变量解析 + 一致性断言）;
5. 既有接口零影响（登录 + 用户列表冒烟）；契约零变化（不改既有接口/表）；
6. 密钥零泄漏：SQL 内只有 `${ENV_VAR}` 占位无真实 Key；git diff 审查。
