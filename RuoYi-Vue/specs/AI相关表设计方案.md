# AI 平台数据表设计方案

> 按域拆分设计、逐域定稿。全部 DB 配置驱动，业务扩展优先加数据、不改代码。公共字段对齐项目规范：`create_by`/`update_by` 为 varchar(64) 登录账号、`del_flag`（0 存在 2 删除），无 `tenant_id`/`create_dept`（本项目无多租户）。

## 域划分总览

| 域 | 表 | 状态 |
|---|---|---|
| 第一部分：基础域 | `ai_provider`、`ai_api_key`、`ai_model` | ✅ 已定稿 |
| 第二部分：对话域 | `ai_agent`、`ai_conversation`、`ai_message` | ✅ 已定稿 |
| 第三部分：RAG 域 | `ai_vector_store`、`ai_knowledge`、`ai_knowledge_document`、`ai_knowledge_segment` | ✅ 已定稿 |
| 第四部分：工具/MCP 域 | `ai_tool`、`ai_mcp`、`ai_tool_call_record` | ✅ 已定稿 |
| 第五部分：用量域 | `ai_usage_log`、`ai_usage_daily`、`ai_usage_today` | ✅ 已定稿 |
| 第六部分：平台使用助手（问答增强） | `ai_problem`、`ai_problem_segment`、`ai_feedback` + 两表加列 | ✅ 已定稿 |

域间依赖方向：对话域 → 基础域（智能体/会话/消息都引用 `model_id`）；RAG 域 → 基础域（embedding/rerank 模型）+ 对话域（消息溯源 `segment_ids`）。

---

# 第一部分：基础域

> 目标：支撑「厂商维护 → 模型维护 → API Key 维护」三个管理场景，全部 DB 配置驱动，新增一家模型供应商不需要发版改代码。

## 一、基础域结构

```
ai_provider（厂商）  1 ──── N  ai_api_key（密钥）  1 ──── N  ai_model（模型）
        │                                                        │
        └──────────────────── provider_id ───────────────────────┘
```

- **ai_provider**：厂商元数据 + 协议适配键 + 默认接入地址。一条 = 一家模型平台（DeepSeek、智谱、Ollama、火山引擎……）。
- **ai_api_key**：厂商的接入凭据。一个厂商可以有多把 Key（官方 / 中转 / 不同配额池），一把 Key 可被多个模型共用。
- **ai_model**：可用模型实例 = 厂商 + 密钥 + 模型标识 + 类型 + 能力位 + 默认推理参数。对话、绘画、知识库等业务全部只认 `model_id`。

三张表的分工对应三个管理页面：

| 管理页面 | 表 | 用户操作 |
|---|---|---|
| 厂商管理 | ai_provider | 维护平台名称、编码、协议类型、默认地址、图标 |
| 密钥管理 | ai_api_key | 维护各厂商的 API Key（支持环境变量占位） |
| 模型管理 | ai_model | 维护模型标识、类型、能力开关、默认参数、排序 |

## 二、基础域表设计

### 2.1 `ai_provider` — AI 厂商表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `provider_id` | bigint PK 自增 | ✓ | 厂商 ID |
| `provider_name` | varchar(100) | ✓ | 厂商名称，如「深度求索」 |
| `provider_code` | varchar(50) | ✓ | **厂商编码，唯一**（业务标识，如 `deepseek`、`siliconflow`、`ollama`） |
| `protocol` | varchar(32) | ✓ | **协议类型 = 客户端适配键**：`openai`（OpenAI 兼容协议）/ `anthropic` / `ollama` / `gemini` / `dify` / `coze` / `custom`。后端按它选 SDK 分支 |
| `api_host` | varchar(500) | ✓ | 厂商默认 API 地址（baseUrl），**必填**——每家厂商的官方地址不同，协议层给不出来（协议官方地址只有 OpenAI 官方自己）。例外：protocol=openai 且确实直连 OpenAI 官方可填 `https://api.openai.com`；解析时 host 为空直接报配置错误 |
| `icon` | varchar(1000) | | 厂商图标（外链） |
| `website` | varchar(1000) | | 官网地址 |
| `key_apply_url` | varchar(1000) | | 密钥申请地址（页面引导用户去哪申请 Key） |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用），停用后该厂商下所有模型不可用 |
| `sort_order` | int | ✓ | 显示顺序（升序） |
| `remark` | varchar(500) | | 备注 / 厂商描述 |
| 公共字段 | — | | `create_by/time`、`update_by/time`、`del_flag`（对齐 BaseEntity：create_by/update_by 为登录账号 varchar(64)；本项目无多租户） |

**约束**：`UNIQUE KEY uk_provider_code (provider_code, del_flag)`

### 2.2 `ai_api_key` — API 密钥表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `key_id` | bigint PK 自增 | ✓ | 密钥 ID |
| `provider_id` | bigint | ✓ | 所属厂商（关联 `ai_provider`） |
| `key_name` | varchar(64) | ✓ | 密钥名称，如「官方主 Key」「中转 Key」 |
| `api_key` | varchar(512) | ✓ | **密钥内容。支持 `${ENV_VAR}` 占位符**——库里存环境变量名，读取时解析，避免明文落库 |
| `base_url` | varchar(500) | | 该 Key 专属网关地址（一级地址）。NULL = 用厂商默认 `api_host`。中转/代理 Key 配中转地址，官方 Key 不配 |
| `api_path` | varchar(500) | | **二级地址**：拼接在网关地址之后的请求路径。部分厂商/网关的端点路径非标准（如中转站带版本段、DashScope 兼容模式 `/compatible-mode/v1`），需要显式指定。NULL = 用协议默认路径 |
| `status` | char(1) | ✓ | 状态（0 启用 1 停用），停用后挂在该 Key 下的模型不可用 |
| `expire_time` | datetime | | 过期时间（可空，用于到期提醒 / 轮换） |
| 公共字段 | — | | 同上 |

**约束**：`KEY idx_provider (provider_id, status)`

### 2.3 `ai_model` — AI 模型表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `model_id` | bigint PK 自增 | ✓ | 模型 ID |
| `provider_id` | bigint | ✓ | 所属厂商（关联 `ai_provider`） |
| `key_id` | bigint | ✓ | 绑定密钥（关联 `ai_api_key`，**应用层校验须与厂商一致**） |
| `model_name` | varchar(100) | ✓ | 模型名称（前端展示，如「DeepSeek V3.2」） |
| `model_code` | varchar(100) | ✓ | **模型标识（传给供应商 API 的真实名字**，如 `deepseek-chat`、`glm-4.6`、`qwen3:4b`） |
| `model_type` | tinyint | ✓ | **模型类型（1 对话 2 图片 3 语音 4 视频 5 向量 embedding 6 重排序 rerank）**——别漏 5/6，知识库绑 embedding、检索链路绑 rerank 都靠它 |
| `support_thinking` | tinyint(1) | ✓ | 支持深度思考（推理模型，思考内容与正文分开返回） |
| `support_tool` | tinyint(1) | ✓ | 支持工具调用（Function Calling） |
| `support_vision` | tinyint(1) | ✓ | 支持图片输入（多模态） |
| `support_audio` | tinyint(1) | ✓ | 支持音频输入 |
| `support_video` | tinyint(1) | ✓ | 支持视频输入 |
| `model_dimension` | int | | **向量维度**（仅 type=5 有意义，如 embedding 模型 2048；建向量集合时用） |
| `temperature` | double | | 默认采样温度。NULL = 供应商默认 |
| `max_tokens` | int | | 默认最大输出 Token |
| `max_contexts` | int | | 默认携带的历史消息条数上限（多轮对话裁剪窗口，裁的是条不是轮） |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用），下架模型不删记录 |
| `sort_order` | int | ✓ | 显示顺序（升序，小的在前；也是前端默认模型的优先级） |
| `remark` | varchar(500) | | 备注（选型建议，如「复杂编程 / 长文本」） |
| 公共字段 | — | | 同上 |

**约束**：`UNIQUE KEY uk_provider_key_model (provider_id, key_id, model_code, del_flag)`、`KEY idx_type_status (model_type, status)`——同一模型可经多把 Key 建多条通道（官方主用 + 中转兜底同时在线），通道由 (provider, key, model_code) 三元组唯一定位；业务表引用的是 `model_id`，切换通道 = 改引用，历史消息里的 `model_code` 冗余不受影响

## 三、基础域运行时解析链路

业务侧拿到 `model_id` 后的取数顺序（每一级 NULL 回落下一级）：

```
1. ai_model      → model_code / model_type / 能力位 / 默认参数
2. ai_api_key    → api_key 原文（若匹配 ${VAR} 则读环境变量）
3. ai_provider   → protocol → 选择客户端实现（openai 兼容 / ollama / ...）
4. 完整请求地址   = (key.base_url ?? provider.api_host) + (key.api_path ?? 协议默认路径)
0. 可用性校验（第 0 步，任一不过即拒绝调用并报明确错误）：
   model.status=0 且 key.status=0 且 provider.status=0 且（expire_time 为空或未过期）
   ——三级状态逐级检查，避免「模型正常但厂商停用」这类半失效态静默打出 401
```

**路径拼接约定（须与实现严格一致；2026-10-11 按 Spring AI 2.0.1 实测修正）**：一级地址（host）填到「端点目录」为止、**不带尾斜杠**（如 `https://api.deepseek.com`、`https://dashscope.aliyuncs.com/compatible-mode`）；「协议默认路径」= **版本段**（openai 协议即 `/v1`，ollama 协议即 `/api`）——因为 openai-java SDK 的 baseUrl 语义是"含版本段、SDK 自己追加 `/chat/completions`"（其默认值 `https://api.openai.com/v1` 即此形态，`completions-path` 配置在 2.0.1 无消费点）。拼接产出 = 传给 SDK 的 baseUrl，如百炼 = `.../compatible-mode` + `/v1` = `.../compatible-mode/v1`。二级地址（key.api_path）用于 host 与版本段之间有路径段的场景（如中转站 `https://relay.example.com` + `/openai/v1`）。

> 二级地址示例：DashScope 的 OpenAI 兼容模式完整端点为 `https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`——按拼接约定，厂商 api_host 填 `https://dashscope.aliyuncs.com/compatible-mode`，api_path 留空走协议默认 `/v1`（SDK baseUrl = `.../compatible-mode/v1`，SDK 追加 `/chat/completions`）；中转站同理拆两段。

推理参数的三层覆盖（对话链路）：**模型默认 → 会话覆盖 → 单次请求临时参数**，上层 NULL 回落下层。

## 四、基础域关键设计决策

1. **密钥独立成表，不放在模型上**。同一厂商常见多把 Key（官方 / 中转 / 不同配额池轮换），密钥独立后：多模型共享一把 Key 不用重复粘贴、换 Key 只改 `key_id` 一处、`${ENV_VAR}` 解析逻辑统一在密钥层。本地厂商（如 Ollama 不需要 Key）建一条占位 Key 记录即可，保持 `key_id NOT NULL`、解析链路单一。
2. **`provider_code` 与 `protocol` 拆成两个字段**。硅基流动、火山引擎、阿里百炼协议上都是 `openai` 兼容，但它们是不同厂商、不同 Key、不同地址。拆开后「新增一家供应商 = 插一行数据」，客户端适配代码只按 `protocol` 分支，不会像平台大枚举那样每加一家就发版。
3. **地址拆两级、多级回落**：网关地址按 `key.base_url > provider.api_host > 协议官方默认` 回落；二级地址按 `key.api_path > 协议默认路径` 回落，两者拼接出完整请求地址。中转 Key 配中转地址，端点非标准的厂商（DashScope 兼容模式等）只补路径，互不干扰。
4. **类型用 tinyint、能力用布尔位**。类型 6 种数字枚举好做字典管理；能力位（thinking/tool/vision/audio/video）支撑前端按能力过滤可选模型、后端拦截不支持的多模态消息，比「类型大类」表达力强（比如「这个对话模型能不能看图」只有能力位能表达）。

## 五、基础域建表 SQL

```sql
-- ----------------------------
-- 1. AI 厂商表
-- ----------------------------
CREATE TABLE ai_provider (
  provider_id     bigint        NOT NULL AUTO_INCREMENT COMMENT '厂商ID',
  provider_name   varchar(100)  NOT NULL                COMMENT '厂商名称',
  provider_code   varchar(50)   NOT NULL                COMMENT '厂商编码（唯一）',
  protocol        varchar(32)   NOT NULL DEFAULT 'openai' COMMENT '协议类型（openai/anthropic/ollama/gemini/dify/coze/custom）',
  api_host        varchar(500)  NOT NULL                COMMENT '默认API地址（必填，各厂商官方地址不同）',
  icon            varchar(1000) DEFAULT NULL            COMMENT '厂商图标',
  website         varchar(1000) DEFAULT NULL            COMMENT '官网地址',
  key_apply_url   varchar(1000) DEFAULT NULL            COMMENT '密钥申请地址',
  status          char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order      int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  remark          varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by       varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time     datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by       varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time     datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag        char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (provider_id),
  UNIQUE KEY uk_provider_code (provider_code, del_flag),
  KEY idx_status (status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI厂商表';

-- ----------------------------
-- 2. AI 密钥表
-- ----------------------------
CREATE TABLE ai_api_key (
  key_id        bigint       NOT NULL AUTO_INCREMENT COMMENT '密钥ID',
  provider_id   bigint       NOT NULL                COMMENT '所属厂商ID',
  key_name      varchar(64)  NOT NULL                COMMENT '密钥名称',
  api_key       varchar(512) NOT NULL                COMMENT '密钥内容（支持${ENV_VAR}占位符）',
  base_url      varchar(500) DEFAULT NULL            COMMENT 'Key专属网关地址（NULL用厂商默认）',
  api_path      varchar(500) DEFAULT NULL            COMMENT '二级地址：拼接在网关地址后的请求路径（NULL用协议默认路径）',
  status        char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0启用 1停用）',
  expire_time   datetime     DEFAULT NULL            COMMENT '过期时间',
  create_by     varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time   datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by     varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time   datetime     DEFAULT NULL            COMMENT '更新时间',
  remark        varchar(500) DEFAULT NULL            COMMENT '备注',
  del_flag      char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (key_id),
  KEY idx_provider (provider_id, status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI密钥表';

-- ----------------------------
-- 3. AI 模型表
-- ----------------------------
CREATE TABLE ai_model (
  model_id          bigint       NOT NULL AUTO_INCREMENT COMMENT '模型ID',
  provider_id       bigint       NOT NULL                COMMENT '所属厂商ID',
  key_id            bigint       NOT NULL                COMMENT '绑定密钥ID（须与厂商一致；同一模型可经多把Key建多条通道）',
  model_name        varchar(100) NOT NULL                COMMENT '模型名称（展示用）',
  model_code        varchar(100) NOT NULL                COMMENT '模型标识（传给API）',
  model_type        tinyint      NOT NULL DEFAULT 1      COMMENT '模型类型（1对话 2图片 3语音 4视频 5向量 6重排序）',
  support_thinking  tinyint      NOT NULL DEFAULT 0      COMMENT '支持深度思考（0否 1是）',
  support_tool      tinyint      NOT NULL DEFAULT 0      COMMENT '支持工具调用',
  support_vision    tinyint      NOT NULL DEFAULT 0      COMMENT '支持图片输入',
  support_audio     tinyint      NOT NULL DEFAULT 0      COMMENT '支持音频输入',
  support_video     tinyint      NOT NULL DEFAULT 0      COMMENT '支持视频输入',
  model_dimension   int          DEFAULT NULL            COMMENT '向量维度（仅type=5时有意义）',
  temperature       double       DEFAULT NULL            COMMENT '默认温度（NULL=供应商默认）',
  max_tokens        int          DEFAULT NULL            COMMENT '默认最大输出Token',
  max_contexts      int          DEFAULT NULL            COMMENT '默认携带历史消息条数上限',
  status            char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order        int          NOT NULL DEFAULT 0      COMMENT '显示顺序（升序）',
  remark            varchar(500) DEFAULT NULL            COMMENT '备注（选型建议）',
  create_by         varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time       datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by         varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time       datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag          char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (model_id),
  UNIQUE KEY uk_provider_key_model (provider_id, key_id, model_code, del_flag),
  KEY idx_key (key_id),
  KEY idx_type_status (model_type, status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI模型表';
```

## 六、基础域内置数据建议

| 厂商 | provider_code | protocol | api_host |
|---|---|---|---|
| DeepSeek | `deepseek` | `openai` | `https://api.deepseek.com` |
| 智谱开放平台 | `zhipu` | `openai` | `https://open.bigmodel.cn/api/paas` |
| 阿里百炼（通义千问） | `bailian` | `openai` | `https://dashscope.aliyuncs.com/compatible-mode` |
| 硅基流动 | `siliconflow` | `openai` | `https://api.siliconflow.cn` |
| Ollama（本地） | `ollama` | `ollama` | `http://localhost:11434` |
| 文心一言 | `yiyan` | `openai` | `https://qianfan.baidubce.com/v2` |
| 讯飞星火 | `xinghuo` | `openai` | `https://spark-api-open.xf-yun.com` |
| 豆包 | `doubao` | `openai` | `https://ark.cn-beijing.volces.com/api/v3` |
| 混元 | `hunyuan` | `openai` | `https://api.hunyuan.cloud.tencent.com` |
| MiniMax | `minimax` | `openai` | `https://api.minimax.chat` |
| 月之暗面 | `moonshot` | `openai` | `https://api.moonshot.cn` |
| 百川智能 | `baichuan` | `openai` | `https://api.baichuan-ai.com` |
| 阶跃星辰 | `stepfun` | `openai` | `https://api.stepfun.com` |
| OpenAI | `openai` | `openai` | `https://api.openai.com` |
| Azure OpenAI | `azureopenai` | `custom` | 按资源域名配置 |
| Anthropic | `anthropic` | `anthropic` | `https://api.anthropic.com` |
| Gemini | `gemini` | `gemini` | `https://generativelanguage.googleapis.com` |
| StableDiffusion | `stablediffusion` | `custom` | `https://api.stability.ai` |
| Midjourney | `midjourney` | `custom` | 反代地址按部署填 |
| Suno | `suno` | `custom` | 代理地址按部署填 |
| Grok | `grok` | `openai` | `https://api.x.ai` |

> 非标准端点的厂商（豆包/百炼/文心 v2/MiniMax 等）在密钥上配二级地址；anthropic/gemini/azure 协议适配器为后续扩展，本期入库仅作目录登记（可用性校验会按 protocol 拒绝无适配器的调用）。
> 厂商 api_host 填到「版本段」为止（如百炼的 `/compatible-mode`），chat/completions 之类的端点路径由二级地址/协议默认路径拼接（见解析链路第 4 步）。host 必填：NULL 回落只对「protocol=openai 且直连 OpenAI 官方」这一种情况成立，统一填显式地址可少一种分支。

Ollama 这类本地厂商建一条占位密钥（key_name=「本地无需密钥」，api_key=任意占位值）。

## 七、基础域注意事项与坑

1. **软删 + 唯一索引冲突**：`uk_provider_code` 与 `uk_provider_key_model` 都含 `del_flag`，逻辑删除后若再删第二条同编码记录会撞唯一键。规避：删除时把 `del_flag` 写成记录 id（或另加 delete_time 进唯一键）；记录少时可接受现状。两处约束共用同一规避方案。
2. **`${ENV_VAR}` 解析要集中一处**：读取密钥的唯一入口做占位符解析（匹配不上环境变量就报配置错误，别静默回退），避免密钥明文散落在日志/导出里。
3. **key_id 与 provider_id 一致性要双向校验**：保存模型时校验 key 的 provider_id 与模型的相同；**编辑密钥时同样校验**——若密钥已挂模型，禁止把 `provider_id` 改到另一厂商（否则已挂靠的模型静默坏掉：key 校验通过、请求打错厂商）。挪厂商一律走「删了重建」。
4. **`model_type` 的 5/6 别漏**：知识库向量化、检索重排都依赖这两个类型，建字典时一次配全。
5. **删除链路的级联约定**：软删厂商前校验其下无未删密钥（有则拒绝或级联软删）；软删密钥前校验其下无未删模型。不做级联校验的话，软删后的 key/model 依然 status=0，解析链会取到「挂在已删除厂商下的模型」这类悬空数据。本期从简：**有子数据一律拒绝删除，先删子再删父**。
6. **过期 Key 本期不自动拦截调用**：可用性校验会检查 `expire_time`（未过期才放行），但到期提醒/自动停用靠人工或后续的定时探活（见预留扩展）；别依赖「过期了系统会自动切 Key」。
7. **同 Key 录两行无数据库约束**：`api_key` 内容不设唯一索引（密文化后也无法唯一），同一把 Key 重复录入靠管理页防重校验（同厂商内 API Key 内容查重）。
8. **model_code 大小写**：utf8mb4_general_ci 下唯一键不区分大小写，`GPT-4` 与 `gpt-4` 会撞键——传给供应商的名字有大小写语义的（罕见）须留意，管理页录入时按原样保存即可，撞键视作重复录入。

## 八、基础域预留扩展（本期不建）

- **Key 健康检查**：`expire_time` + 定时探活，自动停用失效 Key。
- **模型级连接信息覆盖**：若个别模型需要脱离厂商密钥直连，再给 ai_model 加可选的 api_key/base_url 覆盖列（当前用「给这个模型单独建一把 Key」即可表达，不必加列）。
- **用量/计费**：已由第五部分用量域承接（`ai_usage_log` 流水的 `key_id` 维度即本域预留的计费落点），本域不再单独立表。

---

# 第二部分：对话域

> 目标：支撑「智能体（角色）维护 → 会话 → 消息」三个场景，承载流式对话落库、参数逐层覆盖、RAG/联网/附件三种上下文的溯源。设计主线：智能体作为能力挂载中枢、配置快照进会话、上下文压缩、token 逐消息计量。

## 一、对话域结构

```
ai_agent（智能体） 1 ──── N  ai_conversation（会话） 1 ──── N  ai_message（消息）
        │                            │                            │
        └─ model_id（基础域）         └─ model_id（基础域）          └─ model_id（基础域）
                                     └─ agent_id 冗余进消息        └─ reply_id 配对 user/assistant
```

- **ai_agent**：智能体/角色 = systemPrompt + 默认模型 + 能力挂载（知识库/工具/MCP）的一组预设，是 Agent 能力的**运行时挂载中枢**。新建会话选一个智能体，其配置注入后续所有请求。
- **ai_conversation**：会话上下文容器。**参数三件套 + system_prompt 冗余到会话级**——创建时从智能体复制，之后与会话走、独立于智能体后续修改（保证历史会话行为可复现）。
- **ai_message**：逐条消息落库，user/assistant 靠 `reply_id` 配对，是流式对话三路落库（完成/出错/取消）的终点，也是 RAG 溯源、联网来源、附件、思考内容、token 计量的锚点。

| 管理页面 | 表 | 用户操作 |
|---|---|---|
| 智能体管理 | ai_agent | 维护名称、提示词、默认模型、挂载知识库/工具/MCP |
| 会话列表 | ai_conversation | 用户侧"我的会话"（置顶/重命名/删除） |
| 消息记录 | ai_message | 会话详情、历史回放 |

## 二、对话域表设计

### 2.1 `ai_agent` — 智能体表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `agent_id` | bigint PK 自增 | ✓ | 智能体 ID |
| `agent_name` | varchar(100) | ✓ | 名称，如「写作助手」 |
| `avatar` | varchar(1000) | | 头像（外链），列表展示 |
| `category` | varchar(64) | | 分类（如「编程」「翻译」），智能体市场分组 |
| `description` | varchar(500) | | 一句话描述（列表/下拉展示） |
| `system_prompt` | text | ✓ | **系统提示词（核心字段）**：每次请求拼在 Prompt 首位 |
| `model_id` | bigint | | 默认模型（关联基础域 ai_model）。NULL = 创建会话时另选 |
| `knowledge_ids` | varchar(1024) | | 挂载的知识库 ID 数组（JSON 数组，RAG 检索源） |
| `tool_ids` | varchar(1024) | | 挂载的工具 ID 数组（JSON 数组，运行时解析成工具回调挂进请求） |
| `mcp_ids` | varchar(1024) | | 挂载的 MCP 服务 ID 数组（JSON，关联工具域 ai_mcp）。**存 ID 而非名字**——MCP 服务入库（ai_mcp 表）提供 DB 真源，比配置文件 client 名方式可管理（改配置/停用有据可查） |
| `enable_thinking` | tinyint(1) | ✓ | 默认深度思考开关（仅 support_thinking 的模型生效） |
| `welcome_message` | varchar(1000) | | 开场白（进入会话展示） |
| `suggested_questions` | varchar(1024) | | 开场白后的「猜你想问」建议问题数组（JSON），前端渲染为可点击 chip，点击即发送该问题 |
| `visibility` | tinyint | ✓ | 可见性（0 私有 1 公开）。私有=仅 `user_id` 本人可见；公开=进智能体市场 |
| `user_id` | bigint | | 归属用户。配合 visibility：公开智能体归属人仍是作者（用于「我创建的」筛选） |
| `sort_order` | int | ✓ | 显示顺序 |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用） |
| `remark` | varchar(500) | | 备注 |
| 公共字段 | — | | 同基础域（create/update 审计、del_flag）。应用层校验：visibility=0（私有）时 user_id 必填；「我的会话」列表排序用 `COALESCE(last_message_time, create_time)` 或依赖建会话初始化 |

**约束**：`KEY idx_user (user_id, status)`、`KEY idx_category (category)`

**设计要点——JSON 数组挂载 vs 中间表**：knowledge_ids/tool_ids/mcp_ids 三个 JSON 数组列实现"智能体级能力挂载"，避免智能体-知识库、智能体-工具两张中间表。依据：数组规模小（个位数）、只按 agent_id 整体读写、不参与按值查询（不查"哪些智能体挂了库 3"），JSON 列是合理取舍。若未来出现"按挂载反查"的需求再补中间表。

### 2.2 `ai_conversation` — 会话表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `conversation_id` | bigint PK 自增 | ✓ | 会话 ID |
| `user_id` | bigint | ✓ | 归属用户（「我的会话列表」高频查询，建索引） |
| `title` | varchar(256) | ✓ | 会话标题，通常取首条消息截断（首条消息后由模型自动起名） |
| `agent_id` | bigint | | 绑定的智能体。NULL = 无智能体的自由对话 |
| `model_id` | bigint | ✓ | 会话使用的模型（关联基础域 ai_model） |
| `model_code` | varchar(100) | | **冗余模型标识**（避免查历史消息时联表拿模型名） |
| `system_prompt` | text | | 会话级系统提示词：创建时从智能体复制，之后与会话走 |
| `knowledge_ids` | varchar(1024) | | 会话级知识库快照：创建时从智能体复制（理由同上） |
| `tool_ids` | varchar(1024) | | 会话级工具快照：创建时从智能体复制（快照语义与 knowledge_ids 一致，保证可复现） |
| `mcp_ids` | varchar(1024) | | 会话级 MCP 服务快照：创建时从智能体复制（快照语义同上） |
| `enable_thinking` | tinyint(1) | ✓ | 深度思考开关快照：创建时从智能体复制；用户在会话里切换后持久化在此 |
| `summary_upto_message_id` | bigint | | **压缩水位线**：context_summary 覆盖到的最后一条消息 ID。压缩任务只摘水位线之后的消息，摘完推进水位线；NULL = 尚未压缩 |
| `temperature` | double | | 会话级覆盖（NULL = 用模型默认） |
| `max_tokens` | int | | 会话级覆盖（NULL = 用模型默认） |
| `max_contexts` | int | | 会话级覆盖：构建 Prompt 时携带的历史消息条数上限（裁的是条不是轮） |
| `pinned` | tinyint(1) | ✓ | 是否置顶 |
| `pinned_time` | datetime | | 置顶时间（置顶组内排序用） |
| `last_message_time` | datetime | | 最后一条消息时间（会话列表按它排序，避免按消息表聚合）；**建会话时初始化为创建时间**（否则 NULL 行在倒序列表里沉底） |
| `message_count` | int | ✓ | 消息总条数（冗余计数，列表展示；**压缩触发不用它**——见 context_msg_count） |
| `context_msg_count` | int | ✓ | 上下文活跃消息数（use_context=1 的条数冗余）：**长对话压缩的触发依据**；消息落库 +1、压缩置 0 或重生成剔除时 -1 |
| `context_summary` | text | | **上下文压缩摘要**：超长会话把早期轮次摘要成一段存这里，构建 Prompt 时置于历史之前。长对话成本优化的关键设计 |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用/归档） |
| 公共字段 | — | | 同上 |

**约束**：`KEY idx_user_pinned (user_id, pinned, last_message_time)`、`KEY idx_agent (agent_id)`

**设计要点——为什么 system_prompt/knowledge_ids 复制到会话**：会话一旦开始，其行为应由**当时的配置**决定。智能体后来改了提示词，不应影响进行中的历史会话（可复现性）；反之用户在会话里临时改提示词也不应污染智能体全局默认。参数三件套则走"NULL 回落"：会话没改就用模型默认。

### 2.3 `ai_message` — 消息表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `message_id` | bigint PK 自增 | ✓ | 消息 ID |
| `conversation_id` | bigint | ✓ | 所属会话（「按会话拉历史」最高频查询，必须建索引——只建主键不建业务索引是消息表设计的经典事故） |
| `reply_id` | bigint | | **消息配对**：assistant 消息存它回复的 user 消息 ID；user 消息为 NULL。删 user 消息可级联判断删回复 |
| `msg_type` | varchar(16) | ✓ | 消息类型：`user` / `assistant` / `system`（前端按类型渲染气泡） |
| `user_id` | bigint | ✓ | 发消息的用户（assistant 消息也记触发用户，便于统计） |
| `agent_id` | bigint | | 冗余会话的智能体 ID（跨会话按智能体统计，免联会话表） |
| `model_id` | bigint | | 实际生成该回复的模型。**会话中途切模型时可能与会话当前 model_id 不同**，所以逐条记 |
| `model_code` | varchar(100) | | 冗余模型标识（同会话表理由） |
| `msg_status` | tinyint | ✓ | **消息状态**：0 生成中（流式占位行）/ 1 完成 / 2 出错 / 3 已取消。前端据此渲染"生成中"动效或失败样式；服务启动时把残留的 0 置 3（孤儿清理） |
| `content` | mediumtext | | 正文：user 存提问原文；assistant 存**流式聚合完成后的完整回答** |
| `reasoning_content` | text | | **深度思考模型的推理过程**（R1/QwQ 类），与正文分开存，前端折叠展示 |
| `use_context` | tinyint(1) | ✓ | 是否纳入后续请求的历史拼装。默认 1；超长回复/测试消息置 0 即从上下文窗口剔除但不删记录 |
| `segment_ids` | varchar(1024) | | **RAG 溯源**：回答引用的知识库分段 ID 数组（JSON），前端点引用跳知识库段落 |
| `web_search_pages` | text | | 联网来源：JSON 数组（标题/URL/摘要），前端展示来源列表 |
| `attachment_urls` | text | | 附件地址数组（JSON）：多模态消息的图片/文件 |
| `prompt_tokens` | int | ✓ | 本次请求消耗的输入 token（usage 级计量） |
| `completion_tokens` | int | ✓ | 本次回复消耗的输出 token |
| `duration_ms` | int | | 生成耗时（毫秒），速度体感统计 |
| `error_message` | varchar(1000) | | 失败原因（流式出错落库，前端可回放失败原因） |
| 公共字段 | — | | 同上 |

**约束**：`KEY idx_conversation (conversation_id, message_id)`、`KEY idx_user (user_id)`

**设计要点——流式落库三路终点**：流式对话的 onComplete / onError / onCancel 三路都终止于本表——正常完成写 content + tokens + duration；出错写 error_message（content 留已生成的部分）；取消同样落已生成部分并标记。因此 content 是"流式聚合后"的完整值，消息列表页永远不需要再调模型接口。

## 三、对话域运行时链路

**发起一次对话的取数顺序**：

```
1. ai_conversation → system_prompt / knowledge_ids / tool_ids / mcp_ids /
   enable_thinking / 参数三件套（全部会话级快照或覆盖）
   （创建会话时从 ai_agent 复制，之后每轮请求都不再读智能体表）
2. ai_model（基础域）→ model_code / 能力位 / 默认参数（会话级 NULL 的字段回落到这里）
3. 构建请求 = system_prompt + context_summary（压缩摘要）
   + 最近 max_contexts 条 use_context=1 的历史消息
     （同一 reply_id 只取最新一条 assistant——"重新生成"后旧回答不得重复进上下文）
   + RAG 检索（knowledge_ids 快照）命中分段拼 <Reference> 段
   + 工具/MCP（tool_ids/mcp_ids 会话快照，运行时解析）并入工具链
4. 流式返回 → 三路落库 ai_message（content / tokens / duration / error）
```

**参数覆盖层级**（每层 NULL 回落下层）：

```
ai_model（模型默认）→ ai_conversation（会话覆盖）→ 单次请求（前端临时参数）
```

**长对话压缩**：`context_msg_count` 超过阈值（如 40 条）时，从 `summary_upto_message_id` 水位线之后取较早轮次摘要写入 `context_summary`、推进水位线、对应消息 `use_context` 置 0 并同步 `context_msg_count`——记录保留、上下文瘦身，且不会因历史总量增长而反复误触发。

## 四、对话域关键设计决策

1. **智能体 = 挂载中枢，配置复制进会话**。智能体管"预设"，会话管"快照"：提示词、知识库、**工具、MCP、深度思考开关**都在创建会话时复制，参数走 NULL 回落——复制必须彻底，凡"每个会话应可独立演化"的配置一项都不能漏（漏一项，那一项就跟着智能体漂移，可复现性即破）。这样智能体改配置不影响历史会话。
2. **消息表补齐 token 计量与耗时**：`prompt_tokens`/`completion_tokens`/`duration_ms` 三列，是后续用量域统计的数据源——用量域可以从消息表聚合起步，不必先建独立调用日志表。
3. **reply_id 配对而非轮次号**：user/assistant 显式配对让"重新生成"（新 assistant 消息挂同一 reply_id）、"删除某轮"（级联判断）都是一条 SQL 的事。
4. **上下文压缩落点在会话表**：`context_summary` + `use_context` 开关配合，长对话把早期轮次摘要化，token 成本可控。
5. **智能体市场用 visibility 而非 user_id 判空**：用 `user_id IS NULL` 表达公开语义是隐式约定且无法记录作者；拆成 `visibility` + `user_id` 两列，"公开市场"和"我创建的"两个查询都直接。
6. **消息冗余 model_code/agent_id**：会话中途切模型、跨会话统计都是真实场景，逐消息记实际模型 + 冗余标识避免联表。
7. **本期不做**：草稿-发布版本链（parent_id/ancestors/version）——多版本管理是重运营场景，本期智能体直接生效；需要时再加版本四件套，表结构兼容。开场白变量表单——需要时给 ai_agent 加 JSON 列即可。

## 五、对话域建表 SQL

```sql
-- ----------------------------
-- 4. AI 智能体表
-- ----------------------------
CREATE TABLE ai_agent (
  agent_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '智能体ID',
  agent_name       varchar(100)  NOT NULL                COMMENT '智能体名称',
  avatar           varchar(1000) DEFAULT NULL            COMMENT '头像URL',
  category         varchar(64)   DEFAULT NULL            COMMENT '分类（市场分组）',
  description      varchar(500)  DEFAULT NULL            COMMENT '一句话描述',
  system_prompt    text          NOT NULL                COMMENT '系统提示词',
  model_id         bigint        DEFAULT NULL            COMMENT '默认模型ID（NULL=会话时另选）',
  knowledge_ids    varchar(1024) DEFAULT NULL            COMMENT '知识库ID数组（JSON）',
  tool_ids         varchar(1024) DEFAULT NULL            COMMENT '工具ID数组（JSON）',
  mcp_ids          varchar(1024) DEFAULT NULL            COMMENT 'MCP服务ID数组（JSON，关联ai_mcp）',
  enable_thinking  tinyint       NOT NULL DEFAULT 0      COMMENT '默认深度思考开关（0关 1开）',
  welcome_message  varchar(1000) DEFAULT NULL            COMMENT '开场白',
  suggested_questions varchar(1024) DEFAULT NULL        COMMENT '建议问题数组（JSON，开场白chip）',
  visibility       tinyint       NOT NULL DEFAULT 0      COMMENT '可见性（0私有 1公开）',
  user_id          bigint        DEFAULT NULL            COMMENT '归属用户（作者）',
  sort_order       int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  status           char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  remark           varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by        varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time      datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by        varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time      datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag         char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (agent_id),
  KEY idx_user (user_id, status),
  KEY idx_category (category)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI智能体表';

-- ----------------------------
-- 5. AI 会话表
-- ----------------------------
CREATE TABLE ai_conversation (
  conversation_id   bigint       NOT NULL AUTO_INCREMENT COMMENT '会话ID',
  user_id           bigint       NOT NULL                COMMENT '归属用户',
  title             varchar(256) NOT NULL                COMMENT '会话标题',
  agent_id          bigint       DEFAULT NULL            COMMENT '绑定的智能体（NULL=自由对话）',
  model_id          bigint       NOT NULL                COMMENT '会话使用模型ID',
  model_code        varchar(100) DEFAULT NULL            COMMENT '模型标识（冗余，免联表）',
  system_prompt     text                                 COMMENT '会话级系统提示词（创建时从智能体复制）',
  knowledge_ids     varchar(1024) DEFAULT NULL           COMMENT '会话级知识库ID数组（JSON快照）',
  tool_ids          varchar(1024) DEFAULT NULL           COMMENT '会话级工具ID数组（JSON快照，创建时从智能体复制）',
  mcp_ids           varchar(1024) DEFAULT NULL           COMMENT '会话级MCP服务ID数组（JSON快照，创建时从智能体复制）',
  enable_thinking   tinyint      NOT NULL DEFAULT 0      COMMENT '深度思考开关快照（0关 1开）',
  summary_upto_message_id bigint  DEFAULT NULL           COMMENT '压缩水位线（context_summary覆盖到的最后一条消息ID；NULL=未压缩）',
  temperature       double       DEFAULT NULL            COMMENT '温度覆盖（NULL=模型默认）',
  max_tokens        int          DEFAULT NULL            COMMENT '最大输出Token覆盖（NULL=模型默认）',
  max_contexts      int          DEFAULT NULL            COMMENT '历史消息条数上限覆盖（NULL=模型默认）',
  pinned            tinyint      NOT NULL DEFAULT 0      COMMENT '是否置顶（0否 1是）',
  pinned_time       datetime     DEFAULT NULL            COMMENT '置顶时间',
  last_message_time datetime     DEFAULT NULL            COMMENT '最后消息时间（列表排序；建会话时初始化为创建时间防NULL沉底）',
  message_count     int          NOT NULL DEFAULT 0      COMMENT '消息总条数（冗余计数）',
  context_msg_count int          NOT NULL DEFAULT 0      COMMENT '上下文活跃消息数（use_context=1 条数，压缩触发依据）',
  context_summary   text                                 COMMENT '上下文压缩摘要（长对话优化）',
  status            char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1归档）',
  create_by         varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time       datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by         varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time       datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag          char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (conversation_id),
  KEY idx_user_pinned (user_id, pinned, last_message_time),
  KEY idx_agent (agent_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI会话表';

-- ----------------------------
-- 6. AI 消息表
-- ----------------------------
CREATE TABLE ai_message (
  message_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  conversation_id    bigint        NOT NULL                COMMENT '所属会话ID',
  reply_id           bigint        DEFAULT NULL            COMMENT '回复的消息ID（user/assistant配对；user消息为NULL）',
  msg_type           varchar(16)   NOT NULL                COMMENT '消息类型（user/assistant/system）',
  user_id            bigint        NOT NULL                COMMENT '触发用户',
  agent_id           bigint        DEFAULT NULL            COMMENT '智能体ID（冗余，统计用）',
  model_id           bigint        DEFAULT NULL            COMMENT '实际生成模型ID（中途切模型时与会话当前值不同）',
  model_code         varchar(100)  DEFAULT NULL            COMMENT '模型标识（冗余，免联表）',
  msg_status         tinyint       NOT NULL DEFAULT 1      COMMENT '消息状态（0生成中 1完成 2出错 3已取消；启动时残留0置3）',
  content            mediumtext                            COMMENT '消息内容（流式聚合后完整值）',
  reasoning_content  text                                  COMMENT '深度思考推理过程（与正文分开存）',
  use_context        tinyint       NOT NULL DEFAULT 1      COMMENT '是否纳入上下文拼装（0剔除但不删）',
  segment_ids        varchar(1024) DEFAULT NULL            COMMENT '引用的知识库分段ID数组（JSON，RAG溯源）',
  web_search_pages   text                                  COMMENT '联网来源（JSON：标题/URL/摘要）',
  attachment_urls    text                                  COMMENT '附件地址数组（JSON，多模态）',
  prompt_tokens      int           NOT NULL DEFAULT 0      COMMENT '输入Token用量',
  completion_tokens  int           NOT NULL DEFAULT 0      COMMENT '输出Token用量',
  duration_ms        int           DEFAULT NULL            COMMENT '生成耗时（毫秒）',
  error_message      varchar(1000) DEFAULT NULL            COMMENT '失败原因（流式出错落库）',
  create_by          varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time        datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by          varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time        datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag           char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (message_id),
  KEY idx_conversation (conversation_id, message_id),
  KEY idx_user (user_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI消息表';
```

## 六、对话域注意事项与坑

1. **消息表索引是硬要求**：只建主键不建业务索引，session_id 高频查询就是全表扫——本设计 `idx_conversation` 必建，"按会话拉历史"是对话域最高频查询。
2. **assistant 消息先插占位再流式更新**：流式开始时插一条 content 为空的 assistant 消息（拿到 message_id 供前端 SSE 对账），流式过程中分片刷新 content，完成时补 tokens/duration 并置 msg_status=1。别等流结束才落库——中途断线会丢整条消息。
3. **msg_type 与 reply_id 联动校验靠应用层**：reply_id 只能指向同会话的 user 消息；system 消息不参与配对。
4. **"重新生成"的两步约定**：新 assistant 消息挂同一 reply_id；同时把旧回答 `use_context` 置 0 并 `context_msg_count` 减一（若旧回答还在窗口内）。上下文组装再按"同 reply_id 取最新"兜底，双保险防重复。
5. **孤儿占位行清理**：服务启动时批量把 `msg_status=0` 且超过一定时长（如 1 小时）的行置 3——进程崩溃留下的生成中行不能永远占着"生成中"动效。
6. **context_summary 的写入时机**：压缩是后台动作（消息落库后异步判断），别在请求链路里同步做——摘要本身也是一次模型调用，同步做会拖慢响应。压缩窗口从 `summary_upto_message_id` 水位线之后开始，摘完推进水位线。
7. **knowledge_ids/tool_ids/mcp_ids 两处（智能体/会话）语义不同**：智能体的是"预设"，会话的是"快照"，修改入口在各自的页面；不要做双向同步。
8. **消息表增长最快**：mediumtext 正文 + 多个 JSON 列，量级上来了再考虑按月归档或分区；本期不加分区，建表语句保持简单。

---

# 第三部分：RAG 域

> 目标：支撑「向量库实例管理 → 知识库维护 → 文档上传/解析 → 分段管理」的完整 RAG 链路，检索链支持向量 + 关键词混合召回 + rerank 重排，全部参数 DB 配置驱动。设计主线：知识库-文档-分段三级结构与段级治理（vector_id 回写/召回计数/段级开关）、检索参数列级化、向量库实例化与库级 rerank、ES 文档引擎、解析进度反馈、三级冗余计数、融合加权与降级重试。**文档引擎选定 Elasticsearch（本机已部署）**。

## 一、RAG 域结构

```
ai_vector_store（向量库实例）
        │ vector_store_id
        ▼
ai_knowledge（知识库） 1 ──── N  ai_knowledge_document（文档） 1 ──── N  ai_knowledge_segment（分段）
        │ embedding/rerank 模型 ← 基础域 ai_model                    │ vector_id ↔ 向量库一条向量
        └────────────────────────────────────────────────────────────┘
                                    ▲
              对话域 ai_message.segment_ids（溯源链终点）
```

- **ai_vector_store**：向量/文档引擎本身是可管理资源。一个实例 = 一个 Elasticsearch/Milvus/Qdrant/Redis 连接配置，不同知识库可绑不同实例。
- **ai_knowledge**：RAG 一级容器，绑定 embedding 模型 + 向量库实例 + 全套检索/重排参数（库级调参）。
- **ai_knowledge_document**：二级 = 一个上传的原始文档，存解析全文 + 切片参数 + 解析状态机 + 防重哈希。
- **ai_knowledge_segment**：三级 = 召回最小单元，一段文本 ↔ 向量库一条向量，靠 `vector_id` 双向关联；段级开关、召回计数支撑治理。

| 管理页面 | 表 | 用户操作 |
|---|---|---|
| 向量库管理 | ai_vector_store | 维护 Elasticsearch 等实例连接（部署级配置） |
| 知识库管理 | ai_knowledge | 建库、绑向量模型/重排模型/向量库、调检索参数 |
| 文档管理 | ai_knowledge_document | 上传/重解析/启停文档、看解析状态 |
| 分段管理 | ai_knowledge_segment | 查看/编辑/禁用分段、命中测试 |

**溯源链**（与对话域打通）：`ai_message.segment_ids → ai_knowledge_segment（召回+计数+内容）→ ai_knowledge_document（哪份文档）→ ai_knowledge（哪个库）`。

## 二、RAG 域表设计

### 2.1 `ai_vector_store` — 向量库实例表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `store_id` | bigint PK 自增 | ✓ | 实例 ID |
| `store_name` | varchar(100) | ✓ | 实例名称（如「本机 Elasticsearch」） |
| `store_type` | varchar(32) | ✓ | 向量库类型：`milvus` / `qdrant` / `redis` / `elasticsearch` / `weaviate` / `pgvector` 等，后端按它选 DocumentStore 实现分支 |
| `config_options` | text | ✓ | 连接配置（JSON）：host/port/用户名/密码/集合前缀等，运行时反序列化建连接 |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用），停用后挂在其上的知识库检索不可用 |
| `remark` | varchar(500) | | 备注 |
| 公共字段 | — | | 同前 |

**约束**：无特殊索引（配置表量级小）

> **选型结论（学习项目）**：向量库一选定终身，本方案选定 **Elasticsearch**（middle 已部署）：dense_vector KNN 向量检索 + BM25 全文检索（IK 中文分词）一个引擎全包，是主流 RAG 产品的默认文档引擎（常见架构：元数据在 MySQL、正文+向量+分词在 ES）。本方案与该常见架构的差别只在一点：MySQL 同时存分段正文（可读、可编辑、可治理），ES 存正文副本 + 向量 + 分词索引。`ai_vector_store` 表定位为**部署级配置**（换环境改连接配置用），不开放给每个知识库单独选引擎；`ai_knowledge.vector_store_id` 默认取唯一启用实例。
>
> **middle 实例参数**（`middle/elasticsearch`，compose 项目 `middle-elasticsearch`）：镜像 `elasticsearch:8.18.3`，`http://localhost:9200`，`xpack.security.enabled=false` 未开认证（仅本机），JVM 1G / 容器上限 2G，单节点。内置一条默认记录即可：`store_name=本机 Elasticsearch`、`store_type=elasticsearch`、`config_options={"hosts":["http://localhost:9200"]}`。

### 2.2 `ai_knowledge` — 知识库表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `knowledge_id` | bigint PK 自增 | ✓ | 知识库 ID |
| `knowledge_name` | varchar(100) | ✓ | 名称 |
| `description` | varchar(500) | | 描述 |
| `cover` | varchar(1000) | | 封面图（库列表展示） |
| `user_id` | bigint | ✓ | 归属用户（配合 visibility 做个人/公开） |
| `visibility` | tinyint | ✓ | 可见性（0 私有 1 公开）。与对话域 ai_agent 同款约定（显式可见性列，优于 user_id 判空的隐式语义） |
| `embedding_model_id` | bigint | ✓ | 向量模型（基础域 ai_model，type=5）。**同库所有分段必须同一模型向量化（维度一致），绑死在库级**，换模型=全库重嵌入 |
| `embedding_model_code` | varchar(100) | | 冗余模型标识（检索时免联表） |
| `model_dimension` | int | | 向量维度（建向量集合时用；从模型带出，落库防模型记录变更后失配） |
| `vector_store_id` | bigint | ✓ | 向量库实例（关联 ai_vector_store）。默认取唯一启用实例（部署级配置） |
| `collection_name` | varchar(100) | ✓ | 向量集合名（= 向量库索引名，默认 `kb_{knowledge_id}` 自动生成） |
| `rerank_model_id` | bigint | | 重排模型（基础域 ai_model，type=6）。NULL = 不重排 |
| `enable_hybrid` | tinyint(1) | ✓ | 混合检索开关（向量 + 关键词双路召回） |
| `hybrid_alpha` | double | ✓ | 混合权重：0.0=纯关键词 ~ 1.0=纯向量。**默认 0.3**（实证经验：0.3 向量 + 0.7 关键词——词法信号在企业语料上强于语义） |
| `top_k` | int | ✓ | 召回数量 TopK（默认 10。实际执行 topK×4 粗召回 → rerank → 取 topK） |
| `similarity_threshold` | double | ✓ | 相似度阈值。**默认 0.2**（实证默认值，比 0.5 宽松；rerank 归一化后再过滤） |
| `rerank_top_n` | int | | 重排后返回条数（NULL = 取 top_k） |
| `parser_id` | varchar(32) | ✓ | **库级默认解析器**（解析策略是配置不是代码）：`general` 通用 / `qa` 问答对 / `table` 表格 / `paper` 论文 / `book` 书籍 / `laws` 法律 / `manual` 手册 / `presentation` PPT / `one` 整文不切 / `email`。文档可覆盖 |
| `system_prompt` | text | | 知识库级提示词（知识库问答的默认人设） |
| `doc_num` | int | ✓ | 文档数（冗余计数：文档增删时维护，列表页免 COUNT） |
| `chunk_num` | int | ✓ | 分段总数（冗余计数） |
| `token_num` | bigint | ✓ | 全库 Token 总量（冗余计数：容量水位展示） |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用），整库停用 |
| `sort_order` | int | ✓ | 显示顺序 |
| 公共字段 | — | | 同前 |

**约束**：`KEY idx_user_vis (user_id, visibility)`、`UNIQUE KEY uk_collection (vector_store_id, collection_name)`——同一向量实例内集合名唯一

**设计要点——参数全部列级化**：切片/检索/rerank/混合参数都独立成列，而不是收进一个 options JSON。列化的好处：有默认值语义、可加 NOT NULL 约束、管理页表单直映射；JSON 的好处只有"免 DDL"。这些参数是产品核心调参项，值得列级。

### 2.3 `ai_knowledge_document` — 知识库文档表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `document_id` | bigint PK 自增 | ✓ | 文档 ID |
| `knowledge_id` | bigint | ✓ | 所属库（高频按库查文档列表，建索引） |
| `document_name` | varchar(255) | ✓ | 文档名称（原始文件名） |
| `file_url` | varchar(2048) | | 原文件存储地址（对象存储外链） |
| `file_hash` | varchar(64) | | **文件 SHA-256 摘要，库内防重**：唯一索引含 del_flag，同库同文件秒传/防重传 |
| `file_type` | varchar(50) | | 文件类型（pdf/docx/md/txt…） |
| `content` | mediumtext | | **解析后的全文**（Tika 抽取），切片的输入 |
| `content_length` | int | | 全文字符数（列表展示、切片预估） |
| `tokens` | int | | 全文 Token 数（精确计算非字符数/4，切片预算用） |
| `segment_max_tokens` | int | ✓ | 切片参数：每段 Token 上限（默认 800） |
| `segment_overlap` | int | ✓ | 切片参数：相邻段重叠字符数（默认 100，防语义被切断） |
| `split_type` | varchar(20) | ✓ | 切片策略：`auto`（自动检测）/ `token` / `paragraph` / `markdown_qa` / `semantic`。**NULL = 继承库级 `parser_id`**（文档解析策略默认跟库走，可单文档覆盖） |
| `parse_progress` | tinyint | ✓ | **解析进度 0~100**（前端进度条直接绑定，体验远好于只看状态） |
| `parse_status` | tinyint | ✓ | **解析状态机**：0 待解析 / 1 解析中 / 2 已解析 / 3 解析失败（异步解析流水线，配合进度列轮询展示） |
| `error_message` | varchar(1000) | | 解析失败原因（逐阶段失败信息，重试时能看到卡在哪一步） |
| `chunk_num` | int | ✓ | 切出的分段数（冗余计数，文档列表直接展示） |
| `process_begin_at` | datetime | | 开始解析时间 |
| `process_duration_ms` | int | | 解析耗时（毫秒；识别慢文档/慢解析器） |
| `file_size` | bigint | | 原文件大小（字节，列表展示 + 容量统计） |
| `retrieval_count` | int | ✓ | 该文档下分段被召回累计次数（运营分析：哪份文档在被使用） |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用）。停用文档 = 其所有分段不参与召回（切片数据保留，免重切） |
| 公共字段 | — | | 同前 |

**约束**：`KEY idx_knowledge (knowledge_id)`、`UNIQUE KEY uk_kb_hash (knowledge_id, file_hash, del_flag)`

### 2.4 `ai_knowledge_segment` — 知识库分段表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `segment_id` | bigint PK 自增 | ✓ | 分段 ID |
| `knowledge_id` | bigint | ✓ | 所属库（冗余父级：删库批量清理 + 库内跨文档检索直接用它） |
| `document_id` | bigint | ✓ | 所属文档 |
| `content` | text | ✓ | 分段内容（召回后拼进 Prompt `<Reference>` 的就是它）。**编辑段落后须重新向量化并更新向量库**（管理页编辑入口要触发） |
| `content_length` | int | ✓ | 段内字符数 |
| `tokens` | int | ✓ | 段 Token 数 |
| `seg_index` | int | ✓ | 段序号（文档内顺序，前端按序展示、重建向量时按序遍历） |
| `parent_id` | bigint | | **父分段 ID（父子块双层索引）**：小段负责命中（向量精准），命中后取父段大段内容进 Prompt（上下文完整）。NULL = 自身是父段 |
| `vector_id` | varchar(64) | | **该段在向量库中的文档 id**：切片入库后回写；删除/重索引时按它精准删向量，防向量库垃圾残留。NULL = 尚未向量化 |
| `retrieval_count` | int | ✓ | 被召回次数（检索链路自增，热点段落分析 + 命中测试台统计基础） |
| `status` | tinyint | ✓ | 段级开关（0 正常 1 禁用）：某段质量差/涉密单独禁用不参与召回，不动整篇文档 |
| 公共字段 | — | | 同前 |

**约束**：`KEY idx_document (document_id)`、`KEY idx_knowledge (knowledge_id, status)`、`UNIQUE KEY uk_vector_id (vector_id)`——一段一向量

**设计要点——关键词路升级为 ES BM25（换 ES 后的关键收益）**：原 MySQL ngram 方案升级为：分段向 ES 写入正文副本 + IK 分词索引（`*_tks` 式分词字段走 BM25）。IK 对中文分词远好于 ngram 二元切分，BM25 相关性评分是工业标准。MySQL 的 content 列保留（管理页展示/编辑/审计的唯一真源），ES 侧是可随时全量重建的**派生索引**——重建脚本按 `document_id` 从 MySQL 重灌即可。`enable_hybrid` 开关背后的关键词路 = ES `multi_match`（BM25），不再需要 MySQL 全文索引。

## 三、RAG 域运行时链路

**上传解析流水线**：

```
上传文件 → 传对象存储 → 插 ai_knowledge_document（parse_status=0 待解析）
→ 异步任务：Tika 拉文件抽全文 → 算 tokens → 按 segment_max_tokens/overlap/split_type 切片
→ 批量插 ai_knowledge_segment → 逐段向量化 → 写向量库 → 回写 vector_id（parse_status=2）
→ 任一步失败：parse_status=3 + error_message，可重试
```

**检索链路**（对话域 RAG 步骤的展开，含融合权重与降级重试链）：

```
用户问题 → 向量化（库绑 embedding 模型）
→ 向量路：向量库召回 topK×4 条（按 collection_name，KNN cosine）
→ 关键词路（enable_hybrid=1）：ES BM25 召回（IK 分词，按 collection 索引 multi_match）
→ 双路融合：min-max 归一化分数后按 hybrid_alpha 加权（归一化防跨引擎量纲打架；加权可解释，弃黑盒 RRF）
→ rerank（配置了 rerank_model_id）：三个实践坑直接规避——候选池定长（64）、送自然文本不送分词 token、分数 min-max 归一化防负 logit；rerank 后二次过滤
→ 过滤：similarity_threshold 阈值筛 + segment.status=0 + document.status=0
→ 降级重试链（"检索永不空手"，保底可用性）：
    0 命中 → 放宽阈值重试 → 仍 0 → 降 min_match 重试 → 仍 0 → 纯向量兜底
→ 父子块扩展（parent_id 非空：命中子段，取父段内容进 Prompt）
→ 命中分段 retrieval_count 自增（段+文档+库三级计数）
→ 拼 <Reference> 段进 Prompt；segment_ids 写回 ai_message（溯源）
```

> 0.7 词法 + 0.3 向量默认权重与 threshold=0.2：企业语料术语密集、词法信号强，且 rerank 分数归一化后再过滤。学习项目初始参数直接照抄，用命中测试台实测后再调。

## 四、RAG 域关键设计决策

1. **向量库实例化**：向量库本身是 `ai_vector_store` 可管理资源，知识库绑实例。单机部署建一条默认实例即可，架构上不为多引擎付复杂度，却保住了「不同库用不同向量引擎」的自由度。
2. **embedding 模型绑死库级 + 维度落库**：同库所有分段同一模型向量化（维度必须一致），换模型 = 全库重嵌入，所以不让会话/文档级覆盖。维度从模型带出落库，防模型记录后续变更导致建集合参数失配。
3. **文件级防重与解析状态机**：SHA-256 库内唯一索引做秒传/防重；解析状态机 0/1/2/3 让异步流水线可观测可重试——只设一个启停位、失败重试靠人工删了重传的做法可观测性差。
4. **段级治理三件套**：`vector_id` 回写（精准删向量防垃圾残留）、`retrieval_count` 召回计数（热点分析）、段级 `status`（单段禁用不动文档）——三者齐备，分段数据才可治理。
5. **存储架构：MySQL 为真源、ES 为派生索引**：「分段正文/向量/分词全在 ES、MySQL 只存元数据」是重运维路线，学习项目不跟。值得吸收的是四点设计思想：①解析进度列（progress 0~100 + 逐阶段失败信息，文档管理体验的关键）；②解析器配置化（parser_id 十种模板，解析策略是配置不是代码）；③融合默认权重 0.3 向量 + 0.7 词法 + threshold 0.2（实证值，命中测试台调参的起点）；④三级降级重试（检索永不空手）。
6. **父子块双层索引**：`parent_id` 一列实现——切片时按两层大小各切一遍，小段向量化负责精准命中，命中后回溯父段拼进 Prompt 保证上下文完整。本期切片器先只出单层（parent_id 全 NULL），字段先落位，进阶时切片器升级即可、表不动。
7. **混合检索单 alpha**：`hybrid_alpha` 一个系数表达向量/关键词配比（向量权重 w、词法权重 1-w），管理页一个滑杆。
8. **参数列级化**：top_k/阈值/rerank 参数全列级，不收进 options JSON——这些是核心调参项不是杂项配置。
9. **本期不做**：知识图谱 GraphRAG / RAPTOR / 思维导图索引（索引增强，v2 方向）；文档版本历史；知识库对外发布 slug/站点；引用机械补引（实现层设计，不是表设计）。

## 五、RAG 域建表 SQL

```sql
-- ----------------------------
-- 7. AI 向量库实例表
-- ----------------------------
CREATE TABLE ai_vector_store (
  store_id       bigint       NOT NULL AUTO_INCREMENT COMMENT '实例ID',
  store_name     varchar(100) NOT NULL                COMMENT '实例名称',
  store_type     varchar(32)  NOT NULL                COMMENT '向量库类型（milvus/qdrant/redis/elasticsearch/weaviate/pgvector）',
  config_options text         NOT NULL                COMMENT '连接配置（JSON：host/port/密码/集合前缀等）',
  status         char(1)      NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  remark         varchar(500) DEFAULT NULL            COMMENT '备注',
  create_by      varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time    datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by      varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time    datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag       char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (store_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI向量库实例表';

-- ----------------------------
-- 8. AI 知识库表
-- ----------------------------
CREATE TABLE ai_knowledge (
  knowledge_id          bigint        NOT NULL AUTO_INCREMENT COMMENT '知识库ID',
  knowledge_name        varchar(100)  NOT NULL                COMMENT '知识库名称',
  description           varchar(500)  DEFAULT NULL            COMMENT '描述',
  cover                 varchar(1000) DEFAULT NULL            COMMENT '封面图URL',
  user_id               bigint        NOT NULL                COMMENT '归属用户',
  visibility            tinyint       NOT NULL DEFAULT 0      COMMENT '可见性（0私有 1公开）',
  embedding_model_id    bigint        NOT NULL                COMMENT '向量模型ID（ai_model type=5，库级绑死）',
  embedding_model_code  varchar(100)  DEFAULT NULL            COMMENT '向量模型标识（冗余，免联表）',
  model_dimension       int           DEFAULT NULL            COMMENT '向量维度（建集合用）',
  vector_store_id       bigint        NOT NULL                COMMENT '向量库实例ID（部署级配置，默认唯一启用实例）',
  collection_name       varchar(100)  NOT NULL                COMMENT '向量集合名（默认kb_{knowledge_id}）',
  rerank_model_id       bigint        DEFAULT NULL            COMMENT '重排模型ID（ai_model type=6，NULL=不重排）',
  enable_hybrid         tinyint       NOT NULL DEFAULT 1      COMMENT '混合检索开关（0关 1开）',
  hybrid_alpha          double        NOT NULL DEFAULT 0.3    COMMENT '向量权重（0纯关键词~1纯向量，实证默认0.3）',
  top_k                 int           NOT NULL DEFAULT 10     COMMENT '召回数量TopK',
  similarity_threshold  double        NOT NULL DEFAULT 0.2    COMMENT '相似度阈值（实证默认0.2）',
  rerank_top_n          int           DEFAULT NULL            COMMENT '重排后返回条数（NULL=取top_k）',
  parser_id             varchar(32)   NOT NULL DEFAULT 'general' COMMENT '库级默认解析器（general/qa/table/paper/book/laws/manual/presentation/one/email）',
  system_prompt         text                                  COMMENT '知识库级系统提示词',
  doc_num               int           NOT NULL DEFAULT 0      COMMENT '文档数（冗余计数）',
  chunk_num             int           NOT NULL DEFAULT 0      COMMENT '分段总数（冗余计数）',
  token_num             bigint        NOT NULL DEFAULT 0      COMMENT '全库Token总量（冗余计数）',
  status                char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order            int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  create_by             varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time           datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by             varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time           datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag              char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (knowledge_id),
  KEY idx_user_vis (user_id, visibility),
  KEY idx_vector_store (vector_store_id),
  UNIQUE KEY uk_collection (vector_store_id, collection_name)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI知识库表';

-- ----------------------------
-- 9. AI 知识库文档表
-- ----------------------------
CREATE TABLE ai_knowledge_document (
  document_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '文档ID',
  knowledge_id        bigint        NOT NULL                COMMENT '所属知识库ID',
  document_name       varchar(255)  NOT NULL                COMMENT '文档名称',
  file_url            varchar(2048) DEFAULT NULL            COMMENT '原文件地址（对象存储）',
  file_hash           varchar(64)   DEFAULT NULL            COMMENT '文件SHA-256（库内防重/秒传）',
  file_type           varchar(50)   DEFAULT NULL            COMMENT '文件类型（pdf/docx/md/txt）',
  content             mediumtext                            COMMENT '解析后全文（Tika抽取，切片输入）',
  content_length      int           DEFAULT NULL            COMMENT '全文字符数',
  tokens              int           DEFAULT NULL            COMMENT '全文Token数（精确计算）',
  segment_max_tokens  int           NOT NULL DEFAULT 800    COMMENT '切片参数：每段Token上限',
  segment_overlap     int           NOT NULL DEFAULT 100    COMMENT '切片参数：相邻段重叠字符数',
  split_type          varchar(20)   DEFAULT 'auto'          COMMENT '切片策略（NULL=继承库级parser_id；显式指定则覆盖）',
  parse_progress      tinyint       NOT NULL DEFAULT 0      COMMENT '解析进度0~100（进度条直接绑定）',
  parse_status        tinyint       NOT NULL DEFAULT 0      COMMENT '解析状态（0待解析 1解析中 2已解析 3失败）',
  error_message       varchar(1000) DEFAULT NULL            COMMENT '解析失败信息（逐阶段，重试时看卡在哪）',
  chunk_num           int           NOT NULL DEFAULT 0      COMMENT '切出的分段数（冗余计数）',
  process_begin_at    datetime      DEFAULT NULL            COMMENT '开始解析时间',
  process_duration_ms int           DEFAULT NULL            COMMENT '解析耗时（毫秒）',
  file_size           bigint        DEFAULT NULL            COMMENT '原文件大小（字节）',
  retrieval_count     int           NOT NULL DEFAULT 0      COMMENT '召回次数（运营统计）',
  status              char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用；停用=分段不参与召回）',
  create_by           varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time         datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by           varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time         datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag            char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (document_id),
  KEY idx_knowledge (knowledge_id),
  UNIQUE KEY uk_kb_hash (knowledge_id, file_hash, del_flag)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI知识库文档表';

-- ----------------------------
-- 10. AI 知识库分段表
-- ----------------------------
CREATE TABLE ai_knowledge_segment (
  segment_id      bigint       NOT NULL AUTO_INCREMENT COMMENT '分段ID',
  knowledge_id    bigint       NOT NULL                COMMENT '所属知识库ID（冗余父级）',
  document_id     bigint       NOT NULL                COMMENT '所属文档ID',
  content         text         NOT NULL                COMMENT '分段内容（召回拼Prompt）',
  content_length  int          NOT NULL DEFAULT 0      COMMENT '段内字符数',
  tokens          int          DEFAULT NULL            COMMENT '段Token数',
  seg_index       int          NOT NULL DEFAULT 0      COMMENT '段序号（文档内顺序）',
  parent_id       bigint       DEFAULT NULL            COMMENT '父分段ID（父子块双层索引；NULL=自身是父段）',
  vector_id       varchar(64)  DEFAULT NULL            COMMENT '向量库中的向量ID（入库回写；NULL=未向量化）',
  retrieval_count int          NOT NULL DEFAULT 0      COMMENT '被召回次数（检索自增）',
  status          tinyint      NOT NULL DEFAULT 0      COMMENT '段级开关（0正常 1禁用）',
  create_by       varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time     datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by       varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time     datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag        char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (segment_id),
  KEY idx_document (document_id),
  KEY idx_knowledge (knowledge_id, status),
  KEY idx_parent (parent_id),
  UNIQUE KEY uk_vector_id (vector_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI知识库分段表';

-- 注意：MySQL 不建全文索引。关键词召走路 ES（IK 分词 + BM25），
-- ES 索引按知识库 collection 建（对应每个库一个索引），字段参考：
--   content        text    （ik_max_word 分词，BM25）—— MySQL content 的副本
--   segment_id     keyword （回查 MySQL 主键，命中后取 MySQL 最新正文）
--   knowledge_id / document_id  keyword（过滤条件）
--   embedding      dense_vector(dims=模型维度, index=true, similarity=cosine)
-- 向量与分词字段可以同段双字段并存（一段同时有向量字段和分词字段，双字段并存）
```

## 六、RAG 域注意事项与坑

1. **换 embedding 模型 = 全库重嵌入**：管理页要做二次确认 + 提示影响范围。实施时换模型入口直接校验「库下有分段则禁止改绑」更简单。
2. **删除链路要清向量**：删库/删文档/删段必须按 `vector_id` 精准删向量库侧数据，防止垃圾向量残留；按 collection 整删只在删库时用。向量库不可用时删库要做「标记删除 + 后台补偿」。
3. **uk_vector_id 的 NULL 语义**：MySQL 唯一索引对 NULL 不去重（多行 NULL 合法），恰好符合"未向量化"语义，不用额外处理。
4. **ES 索引是派生物、MySQL 是真源**：分段正文以 MySQL 为唯一真源，ES 只存副本 + 向量 + 分词。编辑段落 = 先更 MySQL 再同步 ES；ES 数据丢失/版本升级后按 `document_id` 从 MySQL 全量重灌即可，不用备份 ES。**双写顺序**：先写 MySQL（事务内）再写 ES，ES 失败标记该段「未向量化」等重试，别让 ES 失败回滚 MySQL。
5. **IK 分词器还没装（middle 实例当前状态）**：`middle-elasticsearch` 8.18.3 目前是裸的，standard 分词对中文按单字切。开检索功能前先装插件并重启容器：
   ```bash
   docker compose exec elasticsearch elasticsearch-plugin install https://get.infini.cloud/elasticsearch/analysis-ik/8.18.3
   ```
   插件版本必须与 ES 严格一致（8.18.3）。建索引 mapping 里显式写 `"analyzer": "ik_max_word"`，别依赖默认值。装完用 `_cat/plugins` 确认。
6. **解析任务要幂等**：parse_status=1 解析中的文档被服务重启打断会卡死在中间态——启动时把卡在 1 的重置回 0（或任务表带心跳）。切片+向量化分段批量做，别一段一事务。**progress 只增不减**：先推进度再执行该阶段，前端进度条才不会回跳。
7. **retrieval_count 高频自增**：检索链路每命中一段就 UPDATE 一次，量级大后改异步批量刷（内存累加 + 定时落库）；本期直接更新即可，量级远没到。**三级计数（库/文档/段）在同一事务或同一段代码里维护**，别只刷一层。
8. **相似度阈值语义随引擎不同**：ES dense_vector 的 cosine similarity 返回 0~1（可直比 `similarity_threshold`），但换 Milvus/Qdrant 就要换算（内积可能为负、L2 越小越近）——阈值过滤统一放应用层换算，别下推给引擎。
9. **hybrid_alpha 语义要与实现严格一致**：列注释约定=向量权重（0.3 向量 + 0.7 词法）。实现时融合公式写 `score = alpha×向量分 + (1-alpha)×BM25分`，两边只有一个定义者；注释写反是这类字段的经典事故。**融合前两路分数都要 min-max 归一化**（BM25 分无上限、向量分 0~1，直接加权会让 BM25 压死向量路）。
10. **父子块本期只落字段**：切片器先单层（parent_id 全 NULL），进阶时切片器出两层；检索代码里"命中子段取父段"逻辑可以先写好空判断（parent_id 为 NULL 就返回自身），表结构和检索链路都不用再动。

---

# 第四部分：工具/MCP 域

> 目标：工具、MCP 服务入库统一管理，智能体按 ID 挂载；每次工具调用落审计记录（Agent 可回放）。设计主线：代码内注册工具与 DB 配置工具并存（JSON 数组挂载）、MCP 服务入库管理、HTTP 工具调试闭环、调用全程审计。

## 一、工具/MCP 域结构

```
ai_tool（工具注册） ←── ai_agent.tool_ids / mcp_ids（对话域挂载）──→ ai_mcp（MCP 服务）
        │
        └── ai_tool_call_record（调用审计：谁、哪个会话、调了什么工具、结果如何）
```

- **ai_tool**：一个"工具" = 模型 function calling 里的一项能力。三种来源并存：**builtin**（代码内注册，如时间/计算器，DB 只存元数据）、**http**（DB 配置的 HTTP API 工具）、**mcp**（来自 MCP 服务的远程工具）。智能体 `tool_ids` 挂载后运行时解析成工具回调。
- **ai_mcp**：MCP 服务（server）接入配置。标准 MCP client 配置 JSON（stdio 的 command/args/env，或 HTTP/SSE 的 url/headers）入库，运行时解析建连、枚举工具清单。
- **ai_tool_call_record**：Agent 每次工具调用落一行——回放"智能体是怎么一步步干活的"，是 Agent 可观测性的基础。

| 管理页面 | 表 | 用户操作 |
|---|---|---|
| 工具管理 | ai_tool | 查看 builtin 工具、新增/调试 HTTP 工具、启停 |
| MCP 服务管理 | ai_mcp | 添加 MCP server（JSON 配置）、查看其工具清单、启停 |
| 调用记录 | ai_tool_call_record | 按会话/智能体回放工具调用链 |

## 二、工具/MCP 域表设计

### 2.1 `ai_tool` — 工具表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `tool_id` | bigint PK 自增 | ✓ | 工具 ID |
| `tool_name` | varchar(100) | ✓ | **工具名（function calling 的 name 参数，唯一）**：内置工具=bean 名；HTTP/MCP 工具=注册时生成或导入时指定。给 LLM 的 schema 与回查都靠它 |
| `display_name` | varchar(100) | ✓ | 展示名（管理页/前端列表） |
| `description` | varchar(1000) | ✓ | **给 LLM 看的工具说明**（决定模型何时调用它，写清楚很关键） |
| `tool_type` | varchar(20) | ✓ | `builtin` 代码内 / `http` HTTP API 工具 / `mcp` MCP 远程工具（RuoYi 三分类） |
| `bean_name` | varchar(100) | | tool_type=builtin 时的 Spring bean 名（运行时按名取容器内工具回调） |
| `http_method` | varchar(10) | | tool_type=http 时：GET/POST/PUT/DELETE |
| `base_url` | varchar(500) | | tool_type=http 时：工具服务地址（可含路径模板 `/weather/{city}`） |
| `request_config` | text | | tool_type=http 时：请求配置 JSON（headers / 参数映射 / 超时 / 鉴权头），运行时构造请求 |
| `mcp_id` | bigint | | tool_type=mcp 时：来源 MCP 服务（关联 ai_mcp） |
| `mcp_remote_name` | varchar(200) | | tool_type=mcp 时：MCP server 侧的原始工具名（调用时回传给它；本地名与远程名解耦，防不同 server 重名） |
| `input_schema` | text | | **参数 JSON Schema**（LLM 生成入参的依据）：builtin 从代码注解生成落库展示；http/mcp 导入/枚举时写入 |
| `debug_status` | tinyint | | 调试状态：0 未调试 / 1 成功 / 2 失败。**HTTP/MCP 工具先试跑成功才能上架**（调试闭环，builtin 恒 1） |
| `category` | varchar(64) | | 分类（前端分组展示） |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用） |
| `sort_order` | int | ✓ | 显示顺序 |
| 公共字段 | — | | 同前 |

**约束**：`UNIQUE KEY uk_tool_name (tool_name, del_flag)`、`KEY idx_type_status (tool_type, status)`

**设计要点——三种来源一张表**：不拆 builtin/http/mcp 三张表。function calling 运行时只关心三样东西：名字、描述、参数 schema——这些是公共列；差异（bean 名 / HTTP 配置 / MCP 来源）是类型化的私有列，NULL 互斥即可。工具挂载、调用审计、前端列表全部只对 `tool_id`，一张表贯穿。

### 2.2 `ai_mcp` — MCP 服务表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `mcp_id` | bigint PK 自增 | ✓ | MCP 服务 ID |
| `mcp_name` | varchar(100) | ✓ | 服务名称（唯一，运行时 client 实例标识） |
| `display_name` | varchar(100) | ✓ | 展示名 |
| `description` | varchar(1000) | | 描述（服务用途，辅助判断是否挂载） |
| `transport` | varchar(20) | ✓ | 传输类型：`stdio`（本地子进程）/ `sse` / `streamable_http`（远程） |
| `config_json` | text | ✓ | **标准 MCP client 配置 JSON**：stdio 给 `command/args/env`，远程给 `url/headers`。原样入库，运行时解析建连 |
| `tool_names` | text | | **枚举到的工具清单快照**（JSON 数组：name/description/schema）：连上后拉一次存下来，管理页展示 + 智能体挂载时选择，免每次连 server |
| `conn_status` | tinyint | | 连接状态：0 未连 / 1 已连接 / 2 连接失败（最近一次探活结果） |
| `last_connected_at` | datetime | | 最近连接成功时间 |
| `status` | char(1) | ✓ | 状态（0 正常 1 停用），停用后挂其工具的智能体运行时跳过该 MCP |
| `remark` | varchar(500) | | 备注 |
| 公共字段 | — | | 同前 |

**约束**：`UNIQUE KEY uk_mcp_name (mcp_name, del_flag)`

**设计要点——服务入库 + 工具快照**：配置入库形成 DB 真源（管理页可改），优于只放配置文件的方式（改配置要发版）。工具清单落快照列而不是实时的：MCP server 不常连着，管理页和挂载选择器读快照即可；连接探活按需触发（管理页"测试连接"按钮 / 智能体运行时惰性建连）。

### 2.3 `ai_tool_call_record` — 工具调用记录表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `record_id` | bigint PK 自增 | ✓ | 记录 ID |
| `message_id` | bigint | ✓ | 触发本轮的对话消息（关联对话域 ai_message，回放时按它串起调用链） |
| `conversation_id` | bigint | ✓ | 所属会话（冗余，按会话回放高频） |
| `user_id` | bigint | ✓ | 触发用户 |
| `agent_id` | bigint | | 冗余智能体 ID（跨会话按智能体统计工具使用率） |
| `tool_id` | bigint | ✓ | 被调用的工具 |
| `tool_name` | varchar(100) | ✓ | 冗余工具名（工具删掉后记录仍可读） |
| `tool_type` | varchar(20) | ✓ | 冗余工具类型 |
| `request_args` | text | | 模型生成的入参（JSON 原文） |
| `response_result` | mediumtext | | 工具返回结果（JSON/文本原文；超长截断存储，如只留前 64KB） |
| `success` | tinyint(1) | ✓ | 是否成功（0 失败 1 成功） |
| `error_message` | varchar(1000) | | 失败原因（超时/连接失败/参数校验失败） |
| `duration_ms` | int | | 调用耗时（毫秒） |
| 公共字段 | — | | 同前 |

**约束**：`KEY idx_message (message_id)`、`KEY idx_conversation (conversation_id)`、`KEY idx_agent (agent_id)`

## 三、工具/MCP 域运行时链路

**工具解析**（对话请求构建时的工具装配）：

```
读取 ai_agent.tool_ids + mcp_ids
→ 按 tool_ids 查 ai_tool（一次查出）
   ├─ builtin：按 bean_name 取 Spring 容器里的 ToolCallback
   ├─ http：按 request_config 构造动态 HTTP 工具
   └─ mcp：按 mcp_id 找 ai_mcp → 建连（有连接池用池）→ 按 mcp_remote_name 取远端工具
→ 合并成工具列表 → input_schema 汇总进请求的 tools 参数
```

**调用审计**（模型每次发起工具调用）：

```
模型返回 tool_call（名字+入参）→ 插 ai_tool_call_record（request_args）
→ 执行工具 → 回写 response_result / success / error_message / duration_ms
→ 结果回传模型 → 记录与 ai_message.message_id 关联（一次回复可能多次工具调用）
```

## 四、工具/MCP 域关键设计决策

1. **三来源一张工具表**（builtin/http/mcp 类型化私有列，NULL 互斥）：function calling 只认 name/description/schema 三件公共套，拆三张表反而让挂载、审计、列表处处 UNION。
2. **MCP 挂载按服务、调用按工具**：智能体 `mcp_ids` 存服务 ID 数组（挂载粒度=服务），运行时把该服务的**全部工具**并入工具链；工具级精细过滤（只挑某几个）是重运营需求，需要时给工具表加开关列即可，不引入按工具粒度挂载的复杂度。
3. **工具清单快照进 ai_mcp**（`tool_names` 列）：管理页展示、智能体挂载选择都读快照，连接探活按需触发——MCP server 不可用时功能不瘫。
4. **调试闭环**：HTTP/MCP 工具 `debug_status=1` 才可上架供智能体挂载——先试跑成功再开放，避免模型调到一个坏工具后反复重试浪费 token。
5. **审计记录冗余 tool_name/tool_type**：工具可能被删除，记录里的冗余保证历史可读（对齐消息表冗余 model_code 的同款逻辑）。
6. **本期不做**：插件市场两张表（工具市场 + 分类映射——自用工具量小，管理页分组够用）；MCP 市场两表（浏览-加载-实例化三段式——装 MCP 服务靠管理页手填 JSON 已够）；技能域（技能文件包治理，v2 再说）；工作流编排表（工具域之外的编排大件，单独域设计）。

## 五、工具/MCP 域建表 SQL

```sql
-- ----------------------------
-- 11. AI 工具表
-- ----------------------------
CREATE TABLE ai_tool (
  tool_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '工具ID',
  tool_name       varchar(100)  NOT NULL                COMMENT '工具名（function calling name，唯一）',
  display_name    varchar(100)  NOT NULL                COMMENT '展示名',
  description     varchar(1000) NOT NULL                COMMENT '给LLM看的工具说明',
  tool_type       varchar(20)   NOT NULL DEFAULT 'builtin' COMMENT '工具类型（builtin/http/mcp）',
  bean_name       varchar(100)  DEFAULT NULL            COMMENT 'Spring bean名（builtin用）',
  http_method     varchar(10)   DEFAULT NULL            COMMENT 'HTTP方法（http用）',
  base_url        varchar(500)  DEFAULT NULL            COMMENT '工具服务地址（http用，可含路径模板）',
  request_config  text                                  COMMENT '请求配置JSON：headers/参数映射/超时/鉴权（http用）',
  mcp_id          bigint        DEFAULT NULL            COMMENT '来源MCP服务ID（mcp用）',
  mcp_remote_name varchar(200)  DEFAULT NULL            COMMENT 'MCP server侧原始工具名（mcp用）',
  input_schema    text                                  COMMENT '参数JSON Schema（LLM生成入参依据）',
  debug_status    tinyint       NOT NULL DEFAULT 0      COMMENT '调试状态（0未调试 1成功 2失败；成功才能上架）',
  category        varchar(64)   DEFAULT NULL            COMMENT '分类',
  status          char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  sort_order      int           NOT NULL DEFAULT 0      COMMENT '显示顺序',
  create_by       varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time     datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by       varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time     datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag        char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (tool_id),
  UNIQUE KEY uk_tool_name (tool_name, del_flag),
  KEY idx_type_status (tool_type, status)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI工具表';

-- ----------------------------
-- 12. AI MCP服务表
-- ----------------------------
CREATE TABLE ai_mcp (
  mcp_id            bigint        NOT NULL AUTO_INCREMENT COMMENT 'MCP服务ID',
  mcp_name          varchar(100)  NOT NULL                COMMENT '服务名称（唯一，client实例标识）',
  display_name      varchar(100)  NOT NULL                COMMENT '展示名',
  description       varchar(1000) DEFAULT NULL            COMMENT '描述',
  transport         varchar(20)   NOT NULL DEFAULT 'stdio' COMMENT '传输类型（stdio/sse/streamable_http）',
  config_json       text          NOT NULL                COMMENT '标准MCP client配置JSON（stdio:command/args/env；远程:url/headers）',
  tool_names        text                                  COMMENT '工具清单快照（JSON：name/description/schema）',
  conn_status       tinyint       NOT NULL DEFAULT 0      COMMENT '连接状态（0未连 1已连接 2失败）',
  last_connected_at datetime      DEFAULT NULL            COMMENT '最近连接成功时间',
  status            char(1)       NOT NULL DEFAULT '0'    COMMENT '状态（0正常 1停用）',
  remark            varchar(500)  DEFAULT NULL            COMMENT '备注',
  create_by         varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time       datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by         varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time       datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag          char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (mcp_id),
  UNIQUE KEY uk_mcp_name (mcp_name, del_flag)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI MCP服务表';

-- ----------------------------
-- 13. AI 工具调用记录表
-- ----------------------------
CREATE TABLE ai_tool_call_record (
  record_id       bigint        NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  message_id      bigint        NOT NULL                COMMENT '触发的对话消息ID（ai_message）',
  conversation_id bigint        NOT NULL                COMMENT '所属会话ID（冗余，按会话回放）',
  user_id         bigint        NOT NULL                COMMENT '触发用户',
  agent_id        bigint        DEFAULT NULL            COMMENT '智能体ID（冗余，统计用）',
  tool_id         bigint        NOT NULL                COMMENT '被调用工具ID',
  tool_name       varchar(100)  NOT NULL                COMMENT '工具名（冗余，工具删除后记录仍可读）',
  tool_type       varchar(20)   DEFAULT NULL            COMMENT '工具类型（冗余）',
  request_args    text                                  COMMENT '模型生成的入参（JSON原文）',
  response_result mediumtext                            COMMENT '工具返回结果（超长截断存储）',
  success         tinyint       NOT NULL DEFAULT 1      COMMENT '是否成功（0失败 1成功）',
  error_message   varchar(1000) DEFAULT NULL            COMMENT '失败原因',
  duration_ms     int           DEFAULT NULL            COMMENT '调用耗时（毫秒）',
  create_by       varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time     datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by       varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time     datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag        char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (record_id),
  KEY idx_message (message_id),
  KEY idx_conversation (conversation_id),
  KEY idx_agent (agent_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI工具调用记录表';
```

## 六、工具/MCP 域注意事项与坑

1. **tool_name 是给模型的名字，别用中文**：function calling 的 name 有字符约束（字母数字下划线连字符），管理页要校验；展示名另设 display_name，两者别混用。
2. **HTTP 工具的注入风险**：request_config 里的 URL/headers 是管理员配置的，但**入参来自模型生成的 JSON**——拼 URL 路径模板/查询串前必须对参数做白名单校验（schema 里没声明的参数丢弃），防 SSRF/注入；响应也别原样全量进 Prompt（截断 + 提取关键字段；本期用固定截断即可）。
3. **MCP stdio 进程生命周期**：stdio 型 MCP server 是子进程，连接池要管好进程启停（退出清理、超时回收）；服务频繁启停会留僵尸进程，探活失败要带进程清理。远程型（sse/http）无此问题。
4. **审计表的 response_result 截断阈值**：工具响应可能巨大（如网页抓取），落库前截断（建议 64KB）——审计要的是"调了什么、成了没、返回大概什么"，不是全量副本。
5. **一次回复多次工具调用**：模型一轮可能连续调多个工具（并行或串行），全部挂在同一个 message_id 下，回放按 create_time 排序。
6. **MCP 工具与本表工具的同步**：MCP server 端工具清单会变（升级加工具/下线工具）——快照列会过期。策略：连接成功时刷新快照；运行时按 mcp_remote_name 取不到工具时重新枚举一次再试，仍取不到才报错。

---

# 第五部分：用量域

> 目标：所有模型调用（对话/向量化/重排/工具间接消耗）统一记流水，按天聚合出「谁、哪个智能体、哪个模型、花了多少 token」的统计视图，支撑用量看板与配额管控。常见实现这一层普遍残缺：要么完全没有计量，要么只有消息级总量，要么外挂独立可观测平台——本域以消息级明细计量为基础（消息表照记 + 汇总逐笔维护），补上统一调用流水与预聚合，三层并立。

## 一、用量域结构

```
明细层（已在前四部落位）                    汇总层（本域新建）
ai_message.prompt/completion_tokens ─┐
ai_tool_call_record（工具间接消耗）   ├→ ai_usage_log（调用流水：每次模型API调用一行）
RAG 向量化/重排调用（无明细表）      ─┘        │ 定时聚合（每日）
                                              ▼
                                     ai_usage_daily（天粒度汇总：看板/配额用）
```

**设计原则——流水只记模型 API，看板读预聚合、配额读实时累计**：凡是通过 `ai_model`（基础域）发起的供应商调用都产生一行流水（不管业务来源是对话、RAG 还是工具）；看板读 `ai_usage_daily`（定时聚合）；配额/今日用量读 `ai_usage_today`（调用时逐笔维护的当日累计——天汇总凌晨才聚昨天，白天没有今天的数据，配额必须另有实时来源）。

| 管理页面 | 表 | 用户操作 |
|---|---|---|
| 调用记录 | ai_usage_log | 管理员查流水：按用户/模型/时间筛选 |
| 用量看板 | ai_usage_daily | 天粒度图表：token 趋势、模型分布、用户排行、费用估算 |

**为什么明细要冗余记 token**：消息表 token 列供会话详情页直读；跨来源统计走用量流水，不在明细上做聚合。

## 二、用量域表设计

### 2.1 `ai_usage_log` — 调用流水表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `log_id` | bigint PK 自增 | ✓ | 流水 ID |
| `biz_type` | varchar(20) | ✓ | **业务来源**：`chat` 对话 / `embedding` 向量化 / `rerank` 重排 / `tool` 工具间接（工具内部再调模型）/ `workflow` 工作流 / `other` |
| `biz_id` | bigint | | 业务明细 ID：chat→ai_message_id、embedding/rerank→document_id（解析期）或无、tool→tool_call_record_id。溯源入口 |
| `user_id` | bigint | ✓ | 归属用户（后台批量任务给 0，如文档解析向量化） |
| `agent_id` | bigint | | 冗余智能体（chat/tool 来源才有） |
| `conversation_id` | bigint | | 冗余会话（chat 来源才有） |
| `model_id` | bigint | ✓ | 使用的模型（关联基础域 ai_model） |
| `model_code` | varchar(100) | | 冗余模型标识（模型记录删除后流水仍可读） |
| `provider_id` | bigint | | 冗余厂商（按厂商聚合费用用） |
| `key_id` | bigint | | 冗余密钥（基础域预留的计费维度：中转 Key 与官方 Key 单价不同） |
| `prompt_tokens` | int | ✓ | 输入 token |
| `completion_tokens` | int | ✓ | 输出 token（embedding/rerank 调用为 0） |
| `total_tokens` | int | ✓ | 总量（冗余和，聚合提速） |
| `duration_ms` | int | | 调用耗时（毫秒） |
| `success` | tinyint(1) | ✓ | 是否成功（失败的调用也可能计费部分 token） |
| `error_type` | varchar(32) | | 失败归类：`timeout` / `rate_limit` / `auth` / `server` / `param`（失败率分析） |
| `request_date` | date | ✓ | **调用日期（生成列或写入时落）**：流水表按天分区/归档的依据，免函数索引 |
| 公共字段 | — | | 仅 create_time（流水只追加不修改，不设 update/del 列） |

**约束**：`KEY idx_user_date (user_id, request_date)`、`KEY idx_model_date (model_id, request_date)`、`KEY idx_biz (biz_type, biz_id)`

### 2.2 `ai_usage_daily` — 天粒度用量汇总表

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `stat_id` | bigint PK 自增 | ✓ | 记录 ID |
| `stat_date` | date | ✓ | 统计日期 |
| `user_id` | bigint | ✓ | 用户（0=系统任务，配合配额判断「个人当天用量」） |
| `agent_id` | bigint | | 智能体（0=非智能体来源；按智能体统计用） |
| `model_id` | bigint | ✓ | 模型 |
| `provider_id` | bigint | | 冗余厂商 |
| `call_count` | int | ✓ | 调用次数 |
| `success_count` | int | ✓ | 成功次数（失败率 = 1 - success/call） |
| `prompt_tokens` | bigint | ✓ | 输入 token 合计 |
| `completion_tokens` | bigint | ✓ | 输出 token 合计 |
| `total_tokens` | bigint | ✓ | token 总计 |
| `duration_ms` | bigint | ✓ | 耗时合计（均值 = duration/call_count） |
| `error_count` | int | ✓ | 失败次数（error_type 分布从流水按需查，汇总只留总数） |
| `est_cost` | decimal(12,4) | | **估算费用**（按模型单价表折算；单价变化后重算历史可接受，故存快照值） |
| 公共字段 | — | | 同前 |

**约束**：`UNIQUE KEY uk_dim (stat_date, user_id, agent_id, model_id)`——维度组合唯一，聚合任务可重跑幂等（先删后插或 UPSERT）

**设计要点——为什么有明细还要汇总**：明细表（含 ai_message）只增不减、量级百万起，看板要的「近 30 天按模型分布」在明细上聚合是全表扫描；天粒度汇总把看板查询变成主键级小表查询。维度四列（date/user/agent/model）是看板的全部分析角度，多一列少一列都会让唯一键失效或聚合浪费。

### 2.3 `ai_usage_today` — 当日实时累计表（配额实时数据源）

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `today_id` | bigint PK 自增 | ✓ | 记录 ID |
| `stat_date` | date | ✓ | 统计日期（=当天） |
| `user_id` | bigint | ✓ | 用户（0=系统任务） |
| `model_id` | bigint | ✓ | 模型 |
| `total_tokens` | bigint | ✓ | 当日 token 累计（调用时增量 UPDATE） |
| `call_count` | int | ✓ | 当日调用次数 |
| `est_cost` | decimal(12,4) | | 当日估算费用累计 |
| `updated_time` | datetime | | 最后更新时间 |

**约束**：`UNIQUE KEY uk_day_user_model (stat_date, user_id, model_id)`

**设计要点——为什么配额不能读天汇总**：`ai_usage_daily` 是凌晨定时聚「昨天」的，今天白天它没有今天的数据——配额判断（"这个用户今天用了多少、超没超"）读它会全漏。解决方式：**调用时逐笔增量维护当日累计**。本表：写入流水的同时 UPSERT 本表（一行 update，量级=用户数×当日活跃模型数，极小），跨天由聚合任务顺手清零（或按 stat_date 天然隔离只查当天行）。配额判断 = 查本表当天 user 行合计 vs 阈值，一条索引查询。

## 三、用量域运行时链路

**流水写入**（一次模型调用的计量路径）：

```
业务调用模型 API（经统一的 ModelInvoker 出口）
→ 拿到响应 usage（prompt_tokens/completion_tokens，供应商响应自带）
→ 插 ai_usage_log 一行（biz_type/biz_id 溯源 + 冗余维度）
→ 同步 UPSERT ai_usage_today 当日累计（一行增量 UPDATE，配额实时数据源）
→ 同步维护相关明细表：chat 场景 ai_message 的 token 列照写（明细自读自用）
```

**计量纪律**：所有模型调用必须走统一出口（ModelInvoker），计量插表在这个出口做——散在各业务里写必然漏。embedding/rerank 调用（RAG 解析期）user_id 传 0、biz_id 存 document_id。

**每日聚合**（定时任务，凌晨跑前一天）：

```
扫 ai_usage_log WHERE request_date = 昨天
→ GROUP BY user_id, agent_id, model_id
→ UPSERT ai_usage_daily（uk_dim 幂等，重跑覆盖）
→ 顺手清理 ai_usage_today 前天及更早的行（当日累计表只保留最近两天，防膨胀）
→ 可选：超配额的用户在 ai_usage_today 命中阈值 → 触发限流标记
```

## 四、用量域关键设计决策

1. **流水+汇总两层，明细表不兼任统计源**。本可只靠 ai_message 聚合，但消息表只有对话、RAG 向量化/重排没有明细表、工具内部再调模型更是两层——统一流水把三个来源拉平到一张表，溯源靠 biz_type+biz_id 回跳明细。
2. **配额读当日累计表、看板读天汇总，两者数据源不同**：天汇总凌晨聚昨天，白天配额读它永远是 0——所以配额判断读 `ai_usage_today`，看板读 `ai_usage_daily`。
3. **明细冗余计量 + 流水 + 预聚合三层并存，各司其职**：消息表 token 列给会话详情页直接读；流水给溯源和跨来源统计；天汇总给看板。三层都在同一次 ModelInvoker 出口里维护，一次写齐。
4. **费用列存估算快照（est_cost）**：模型单价天然变动（降价是常态），历史费用按调用时点单价折算快照落库——统计口径稳定，重算历史永远是可选项而非依赖。输入/输出单价拆开折算再合计（decimal(12,4)），比单一总量列口径更准。
5. **失败也记流水**：失败调用可能消耗了部分 token（供应商按实际用量计费），且失败率/错误类型分布本身是模型健康度指标（error_type 归类）。
6. **冗余维度一次落齐**（model_code/provider_id/key_id）：流水是只追加表，供应商侧模型改名、Key 换绑都不影响历史可读性；key_id 维度正是基础域预留计费的落点。
7. **本期不做**：外接可观测平台（自建看板是学习目的本身）；月粒度汇总表（天表够用，需要时天表再聚）；计费扣费/余额体系（est_cost 只是估算展示，真实计费要账户体系，超出学习项目范围）；按小时粒度（当前量级无意义）；投票/标注反馈统计（v2 做反馈闭环时一起规划）。

## 五、用量域建表 SQL

```sql
-- ----------------------------
-- 14. AI 调用流水表
-- ----------------------------
CREATE TABLE ai_usage_log (
  log_id            bigint       NOT NULL AUTO_INCREMENT COMMENT '流水ID',
  biz_type          varchar(20)  NOT NULL                COMMENT '业务来源（chat/embedding/rerank/tool/workflow/other）',
  biz_id            bigint       DEFAULT NULL            COMMENT '业务明细ID（chat→message，embedding/rerank→document，tool→call_record）',
  user_id           bigint       NOT NULL DEFAULT 0      COMMENT '归属用户（0=系统任务）',
  agent_id          bigint       DEFAULT NULL            COMMENT '智能体ID（冗余）',
  conversation_id   bigint       DEFAULT NULL            COMMENT '会话ID（冗余）',
  model_id          bigint       NOT NULL                COMMENT '模型ID（ai_model）',
  model_code        varchar(100) DEFAULT NULL            COMMENT '模型标识（冗余）',
  provider_id       bigint       DEFAULT NULL            COMMENT '厂商ID（冗余）',
  key_id            bigint       DEFAULT NULL            COMMENT '密钥ID（冗余，计费维度）',
  prompt_tokens     int          NOT NULL DEFAULT 0      COMMENT '输入Token',
  completion_tokens int          NOT NULL DEFAULT 0      COMMENT '输出Token',
  total_tokens      int          NOT NULL DEFAULT 0      COMMENT 'Token总量',
  duration_ms       int          DEFAULT NULL            COMMENT '调用耗时（毫秒）',
  success           tinyint      NOT NULL DEFAULT 1      COMMENT '是否成功（0失败 1成功）',
  error_type        varchar(32)  DEFAULT NULL            COMMENT '失败归类（timeout/rate_limit/auth/server/param）',
  request_date      date         NOT NULL                COMMENT '调用日期（分区归档依据）',
  create_time       datetime     DEFAULT NULL            COMMENT '创建时间',
  PRIMARY KEY (log_id),
  KEY idx_user_date (user_id, request_date),
  KEY idx_model_date (model_id, request_date),
  KEY idx_biz (biz_type, biz_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI调用流水表';

-- ----------------------------
-- 15. AI 用量天汇总表
-- ----------------------------
CREATE TABLE ai_usage_daily (
  stat_id           bigint        NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  stat_date         date          NOT NULL                COMMENT '统计日期',
  user_id           bigint        NOT NULL DEFAULT 0      COMMENT '用户（0=系统）',
  agent_id          bigint        NOT NULL DEFAULT 0      COMMENT '智能体（0=非智能体来源）',
  model_id          bigint        NOT NULL                COMMENT '模型ID',
  provider_id       bigint        DEFAULT NULL            COMMENT '厂商ID（冗余）',
  call_count        int           NOT NULL DEFAULT 0      COMMENT '调用次数',
  success_count     int           NOT NULL DEFAULT 0      COMMENT '成功次数',
  error_count       int           NOT NULL DEFAULT 0      COMMENT '失败次数',
  prompt_tokens     bigint        NOT NULL DEFAULT 0      COMMENT '输入Token合计',
  completion_tokens bigint        NOT NULL DEFAULT 0      COMMENT '输出Token合计',
  total_tokens      bigint        NOT NULL DEFAULT 0      COMMENT 'Token总计',
  duration_ms       bigint        NOT NULL DEFAULT 0      COMMENT '耗时合计',
  est_cost          decimal(12,4) NOT NULL DEFAULT 0     COMMENT '估算费用（调用时点单价快照）',
  create_time       datetime      DEFAULT NULL            COMMENT '创建时间',
  update_time       datetime      DEFAULT NULL            COMMENT '更新时间',
  PRIMARY KEY (stat_id),
  UNIQUE KEY uk_dim (stat_date, user_id, agent_id, model_id),
  KEY idx_date (stat_date)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI用量天汇总表';

-- ----------------------------
-- 16. AI 当日实时累计表（配额判断数据源）
-- ----------------------------
CREATE TABLE ai_usage_today (
  today_id    bigint        NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  stat_date   date          NOT NULL                COMMENT '统计日期（=当天）',
  user_id     bigint        NOT NULL DEFAULT 0      COMMENT '用户（0=系统）',
  model_id    bigint        NOT NULL                COMMENT '模型ID',
  total_tokens bigint       NOT NULL DEFAULT 0      COMMENT '当日Token累计',
  call_count  int           NOT NULL DEFAULT 0      COMMENT '当日调用次数',
  est_cost    decimal(12,4) NOT NULL DEFAULT 0     COMMENT '当日估算费用累计',
  updated_time datetime     DEFAULT NULL            COMMENT '最后更新时间',
  PRIMARY KEY (today_id),
  UNIQUE KEY uk_day_user_model (stat_date, user_id, model_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI当日实时累计表';
```

> 三张统计表都是**只追加/只覆盖**语义（流水只插、天汇总聚合任务 UPSERT、当日累计增量 UPDATE），不设 create_by/update_by/del_flag 审计列：无业务修改路径。「统计表最简」优先于与业务表公共字段风格一致，这是本域刻意的一次偏离。

## 六、用量域注意事项与坑

1. **计量出口必须唯一**：ModelInvoker 之外任何绕过出口直接调供应商 SDK 的代码都会让流水漏计——统计数字对不上时先查这个。Spring 里用 AOP 或模板方法收口，别靠 code review 自觉。
2. **流式调用的 usage 位置**：OpenAI 兼容协议要显式传 `stream_options.include_usage=true` 才在最后一个 chunk 带 usage；有的中转网关会吞掉 usage——收不到时按 tokenizer 本地估算并打标（本地估算与供应商计数有偏差，流水里可加 est 标记列或先容忍误差）。
3. **聚合任务要幂等且可补跑**：uk_dim 唯一键 + UPSERT 保证重跑不重不漏；某天任务失败第二天补跑即可，别设计成「只能跑一次」。聚合窗口参数化（昨天/指定日期），支持手动触发重算。
4. **est_cost 的单价来源**：维护在代码常量或字典表（model_code → 输入/输出单价），本期不建价格表；单价变动后历史快照不改，仅新调用按新价。
5. **流水表增长与归档**：量级百万后按 request_date 归档（归档到冷表或直接按天分区）；汇总表是天粒度的最终真相，归档删流水不影响历史统计。
6. **配额管控的最小实现**：判断只读 ai_usage_today（当天该 user 的行合计 vs 阈值），超了在对话入口拒绝并提示——不需要引入计数器/Redis 原子操作，天粒度下的秒级误差无意义。

---

# 第六部分：平台使用助手（问答增强）

> 定位：平台内嵌的「使用助手」——用户问「用户管理怎么用」「部门管理如何配」「怎么重置密码」，助手基于平台操作文档作答。知识源封闭可控（平台自身文档）、用户是已登录平台用户（复用 user_id）、入口是内嵌对话（复用对话域）。**不新增渠道/终端用户域**（那是对外开放客服的事，见本部分末尾边界说明）。
>
> 设计要点：企业知识库问答验证过的三个机制——FAQ 直接返回、问题池多向量召回、反馈标注闭环；外加标题链注入。

## 一、业务形态与复用关系

```
平台文档（操作手册/FAQ/功能说明）
   ↓ 上传到 ai_knowledge（建一个「平台使用文档」知识库，parser_id=manual/qa）
   ↓
ai_agent「平台使用助手」（system_prompt=助手人设 + knowledge_ids=[平台文档库]）
   ↓ 用户在平台内嵌聊天入口提问（复用对话域，无新表）
ai_conversation / ai_message ——回答带 segment_ids 溯源（点引用跳文档段落）
   ↓
ai_feedback 反馈 → 管理员纠错 → 问题池/文档更新 → 越用越准
```

**第一期能力边界**（不加表即可跑通）：上传文档 + 建预置智能体 + 会话挂知识库。三小件（本期加）解决的是检索命中率与运营闭环：直接返回（FAQ 零成本秒回）、标题链（操作手册层级结构切片不丢上下文）、问题池（同义问法召回）+ 反馈（纠错闭环）。

## 二、表设计（两新表 + 两加列）

### 2.1 加列一：`ai_knowledge_document` 增 FAQ 直接返回

| 新增字段 | 类型 | 说明 |
|---|---|---|
| `hit_handling_method` | varchar(20) | 命中处理方式：`optimization` 模型优化（默认，检索→LLM 生成）/ `directly_return` 直接返回（命中即返回段落原文，**不调 LLM**——零幻觉、秒回、不烧 token）。重置密码类 FAQ 配这条 |
| `directly_return_similarity` | double | 直接返回的相似度阈值（默认 0.9，文档级可调） |

检索链路变化（在 RAG 域检索链路的阈值过滤之后插入）：`directly_return` 文档的分段若最高分 ≥ directly_return_similarity → 直接取该段原文作答（回答的 msg_type 正常、溯源照写），跳过 LLM 调用。

### 2.2 加列二：`ai_knowledge_segment` 增标题链

| 新增字段 | 类型 | 说明 |
|---|---|---|
| `title` | varchar(500) | **标题链**：切片时把文档的层级标题注入（如「用户管理 > 密码重置 > 修改流程」），以 ` > ` 连接。切片器在切「用户管理」章节下的步骤段时，段内容可能只剩步骤，title 保证检索命中后上下文不丢；同时该列参与向量化/关键词索引（标题词是强信号） |
| `menu_paths` | varchar(1024) | **功能直达链接（JSON 数组）**：本段涉及的平台功能路由路径，如 `["/system/user","/system/user/profile"]`。FAQ 编辑时从平台菜单树（sys_menu）选择而非手填 URL——菜单表就是「功能名→路由」的真源，菜单改名/搬家只改一处。回答层把命中分段的 menu_paths 汇总去重后随答案返回，前端渲染「前往 用户管理 →」跳转按钮——**提问「如何增加用户」→ 得到步骤 + 一键跳到新增用户页** |

切片器规则：`markdown_qa`/`paragraph` 策略下按标题层级切，title = 祖先标题链 + 本段标题；两小时工作量，命中率提升明显（小技巧大提升）。

### 2.3 `ai_problem` + `ai_problem_segment` — 问题池

**解决同义问法**：「改密码」「重置密码」「找回密码」「password 怎么改」是四个不同的问法指向同一段落——只向量化段落原文，改写型问法必漏。问题池让**一个问题一条向量**，问题也进检索。

| 字段（ai_problem） | 类型 | 必填 | 说明 |
|---|---|---|---|
| `problem_id` | bigint PK 自增 | ✓ | 问题 ID |
| `knowledge_id` | bigint | ✓ | 所属知识库 |
| `content` | varchar(256) | ✓ | 问题内容（索引；同库内容唯一） |
| `source_type` | tinyint | ✓ | 0 用户添加 / 1 LLM 生成（解析时为每段生成候选问题）/ 2 用户反馈沉淀（ai_feedback 纠错时顺手入库） |
| `hit_num` | int | ✓ | 命中次数（问题被检索命中自增——运营看哪些问法高频） |
| 公共字段 | — | | 同前 |

**约束**：`UNIQUE KEY uk_kb_content (knowledge_id, content, del_flag)`

| 字段（ai_problem_segment） | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | bigint PK 自增 | ✓ | 关联 ID |
| `problem_id` / `segment_id` | bigint | ✓ | 问题×段落多对多 |
| `knowledge_id` / `document_id` | bigint | ✓ | 冗余父级（检索过滤/清理用） |

**约束**：`UNIQUE KEY uk_problem_segment (problem_id, segment_id)`

**向量化策略**：problem 每条独立向量化（embedding 表的 source_type 加 `problem` 类型，或复用 segment 的 vector 写入通道以 `p_{problem_id}` 作向量 id）——检索时用户问题先经向量匹配 problem，命中即按 mapping 跳到段落，同义问法召回的关键。

### 2.4 `ai_feedback` — 反馈表（赞踩 + 归因 + 纠错闭环）

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `feedback_id` | bigint PK 自增 | ✓ | 反馈 ID |
| `message_id` | bigint | ✓ | 针对的 assistant 消息（unique——一条消息一条反馈，重复提交覆盖） |
| `conversation_id` / `user_id` | bigint | ✓ | 冗余会话/用户（统计用） |
| `feedback_type` | tinyint | ✓ | 0 赞 / 1 踩 |
| `reason` | varchar(32) | | 踩的归因（枚举）：`inaccurate` 不准确 / `incomplete` 不完整 / `irrelevant` 不相关 / `other` 其他 |
| `comment` | varchar(1000) | | 补充说明（用户自由文本） |
| `corrected_content` | text | | **管理员纠错答案**（运营处理时填写） |
| `improve_segment_ids` | varchar(1024) | | 关联改进后的知识库分段（JSON 数组）——纠错内容入库向量化后指向它 |
| `handle_status` | tinyint | ✓ | 处理状态：0 未处理 / 1 已处理 / 2 已忽略 |
| `handled_by` / `handled_time` | bigint / datetime | | 处理人/时间 |
| 公共字段 | — | | 同前 |

**约束**：`UNIQUE KEY uk_message (message_id)`、`KEY idx_handle (handle_status, feedback_type)`

**闭环机制（标注即知识）**：用户踩+归因 → 运营在后台看到反馈列表 → 填纠错答案 → **答案以 problem(source_type=2)+segment 入库并异步向量化** → improve_segment_ids 指向新分段 → 下次同类问题直接命中纠错答案。知识库不需要重新上传文档，"越用越准"。

## 三、平台使用助手运行时链路

```
用户提问「部门管理怎么配」
→ 问题改写（多轮时结合历史改写完整问题，实现层）
→ 检索：
   ① 问题向量 匹配 ai_problem 向量 → 命中 → 按 mapping 取段落（同义问法直达）
   ② 段落向量 + BM25 混合召回（RAG 域标准链路）
   ③ 两路合并 → 若最高分段属 directly_return 文档且 ≥ 阈值 → 直接返回原文（不调 LLM）
   ④ 否则拼 <Reference> → LLM 生成（agent 的 system_prompt 限定"只答平台使用问题"）
→ 回答落 ai_message（segment_ids 溯源，前端点引用跳文档；
   命中分段的 menu_paths 汇总去重 → 随答案返回 → 前端渲染「前往 用户管理 →」功能直达按钮）
→ usage 流水照记（用量域）
→ 用户反馈 → ai_feedback → 纠错入库 → 下一轮检索变准
```

## 四、关键设计决策与边界

1. **复用优先，最小增量**：不加渠道表/终端用户表/会话参数——平台内用户的身份（user_id）、入口（内嵌对话）、权限（平台登录态）全部现成。区别于对外开放客服（那才需要 ai_channel + ai_cs_user，见第 5 条边界）。
2. **预置智能体 = 一行数据**：`ai_agent` 插一条「平台使用助手」（system_prompt 限定答题范围 + 拒答话术 + knowledge_ids=[平台文档库] + enable_thinking=0），前端聊天入口默认进它。换知识库/调人设是管理页改数据，不发版。
3. **问题池是命中率的第一杠杆**：企业文档问答的失败大多不是模型不行，是「用户的问法 ≠ 文档的写法」。问题池 + 问题改写两件配合，同义问法命中率提升最明显；且 problem 粒度小，向量化成本低。
4. **直接返回是 FAQ 场景的成本与体验双赢**：重置密码、查日志路径这类固定答案，命中率×阈值达标直接返回——秒回、零幻觉、不烧 token。注意它只在**文档级**开关（运营决定哪些文档的段落可信到可以直接返回），不是全局开关。
5. **边界——什么情况升级成「开放客服域」**：如果要服务**平台外**用户（官网访客/微信渠道），才需要两块新表设计：`ai_channel`（渠道+访问令牌+白名单+频控+兜底话术）与 `ai_cs_user`（终端用户与平台用户分离+组授权），届时会话表加 `channel_id`/`cs_user_id` 两列即可，本部分表结构不变。
6. **v2 预留**：转人工（conversation 加 transfer_status）、敏感词审核（实现层钩子）、命中测试台（管理页功能，读检索链路不改表）。

## 五、建表 SQL

```sql
-- ----------------------------
-- 17. AI 问题池表
-- ----------------------------
CREATE TABLE ai_problem (
  problem_id   bigint       NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  knowledge_id bigint       NOT NULL                COMMENT '所属知识库',
  content      varchar(256) NOT NULL                COMMENT '问题内容',
  source_type  tinyint      NOT NULL DEFAULT 0      COMMENT '来源（0用户添加 1LLM生成 2反馈沉淀）',
  hit_num      int          NOT NULL DEFAULT 0      COMMENT '命中次数',
  create_by    varchar(64)       DEFAULT ''         COMMENT '创建者',
  create_time  datetime     DEFAULT NULL            COMMENT '创建时间',
  update_by    varchar(64)       DEFAULT ''         COMMENT '更新者',
  update_time  datetime     DEFAULT NULL            COMMENT '更新时间',
  del_flag     char(1)      NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (problem_id),
  UNIQUE KEY uk_kb_content (knowledge_id, content, del_flag),
  KEY idx_knowledge (knowledge_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI问题池表';

-- ----------------------------
-- 18. AI 问题段落关联表
-- ----------------------------
CREATE TABLE ai_problem_segment (
  id           bigint   NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  problem_id   bigint   NOT NULL                COMMENT '问题ID',
  segment_id   bigint   NOT NULL                COMMENT '分段ID',
  knowledge_id bigint   NOT NULL                COMMENT '知识库ID（冗余）',
  document_id  bigint   NOT NULL                COMMENT '文档ID（冗余）',
  PRIMARY KEY (id),
  UNIQUE KEY uk_problem_segment (problem_id, segment_id),
  KEY idx_segment (segment_id)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI问题段落关联表';

-- ----------------------------
-- 19. AI 反馈表
-- ----------------------------
CREATE TABLE ai_feedback (
  feedback_id         bigint        NOT NULL AUTO_INCREMENT COMMENT '反馈ID',
  message_id          bigint        NOT NULL                COMMENT '针对的assistant消息ID',
  conversation_id     bigint        NOT NULL                COMMENT '会话ID（冗余）',
  user_id             bigint        NOT NULL                COMMENT '反馈用户',
  feedback_type       tinyint       NOT NULL                COMMENT '反馈类型（0赞 1踩）',
  reason              varchar(32)   DEFAULT NULL            COMMENT '踩的归因（inaccurate/incomplete/irrelevant/other）',
  comment             varchar(1000) DEFAULT NULL            COMMENT '补充说明',
  corrected_content   text                                  COMMENT '管理员纠错答案',
  improve_segment_ids varchar(1024) DEFAULT NULL            COMMENT '关联改进后的分段（JSON数组）',
  handle_status       tinyint       NOT NULL DEFAULT 0      COMMENT '处理状态（0未处理 1已处理 2已忽略）',
  handled_by          bigint        DEFAULT NULL            COMMENT '处理人',
  handled_time        datetime      DEFAULT NULL            COMMENT '处理时间',
  create_by           varchar(64)        DEFAULT ''         COMMENT '创建者',
  create_time         datetime      DEFAULT NULL            COMMENT '创建时间',
  update_by           varchar(64)        DEFAULT ''         COMMENT '更新者',
  update_time         datetime      DEFAULT NULL            COMMENT '更新时间',
  del_flag            char(1)       NOT NULL DEFAULT '0'    COMMENT '删除标志（0代表存在 2代表删除）',
  PRIMARY KEY (feedback_id),
  UNIQUE KEY uk_message (message_id),
  KEY idx_handle (handle_status, feedback_type)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='AI反馈表';

-- 加列（ALTER）
ALTER TABLE ai_knowledge_document
  ADD COLUMN hit_handling_method      varchar(20) NOT NULL DEFAULT 'optimization' COMMENT '命中处理方式（optimization模型优化/directly_return直接返回）',
  ADD COLUMN directly_return_similarity double    NOT NULL DEFAULT 0.9         COMMENT '直接返回相似度阈值';

ALTER TABLE ai_knowledge_segment
  ADD COLUMN title varchar(500) DEFAULT NULL COMMENT '标题链（祖先标题>本段标题，切片注入）';
```

## 六、注意事项与坑

1. **直接返回的阈值要按文档调**：0.9 是起步值，FAQ 问法高度一致可降到 0.85，自由文本类文档保持 0.9+ 宁可不直返；直返答错的代价是"理直气壮地错"，比 LLM 编造更伤信任，运营要能一键把文档切回 optimization。
2. **问题池的生成质量控制**：LLM 为段落生成候选问题（source_type=1）要有审核位（生成≠可用），或至少支持批量停用；垃圾问题会污染检索（向量命中错问题→跳错段落）。
3. **纠错入库走标准流水线**：corrected_content 入库不是插一行完事——要走"建 problem + 建 segment + 异步向量化 + 回写向量 id"完整链路（复用文档解析流水线的后半段），否则纠错答案根本检索不到。
4. **反馈的唯一键防刷**：uk_message 保证一条消息一条反馈（重复提交=覆盖），赞踩比统计不失真；运营后台默认按 handle_status=0 + feedback_type=1（未处理的踩）排序。
5. **助手人设要硬约束答题范围**：system_prompt 明确「只回答平台使用问题，无关问题礼貌拒答」，否则用户拿它当通用聊天机器人，token 烧得快且答非所问伤信任。拒答话术进 prompt，兜底行为（检索无命中时）在代码层判断。
6. **menu_paths 的真源是菜单表不是文档**：FAQ 编辑从平台菜单树（sys_menu）选功能，落路由路径数组——菜单改名/搬家只改菜单表一处，文档不用跟着改。运行时校验：返回前过滤掉当前用户无权限的路由（用户 A 没配「部门管理」权限，回答里就不该出现跳转按钮）；LLM 只负责组织语言，**链接由结构化数据给出，不让模型生成 URL**（幻觉链接是灾难）。
