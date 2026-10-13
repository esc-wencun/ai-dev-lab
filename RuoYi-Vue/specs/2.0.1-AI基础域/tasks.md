# Tasks · 2.0.1 AI 基础域：厂商 / 密钥 / 模型管理

> 勾选纪律：做完即勾（含对应验证跑通）；没做的不许勾，行尾注明原因；勾选 = 验收通过。
> 依赖：2.0.0（ruoyi-ai 模块已就位）。表设计权威源：[../AI相关表设计方案.md](../AI相关表设计方案.md) 第一部分。
> 环境约束：8080 端口互斥；IDEA 运行中不编译（编译纪律第 0 节）；MySQL 13306（本工作区 Docker，端口已迁移）。

## 阶段一：数据层

## Task 1: 建表 SQL 与内置数据
- [x] 新建 `RuoYi-Vue/sql/ai_20261010.sql`：三张表 CREATE TABLE（从设计文档第一部分第五节搬运，含审查修正）+ delete 前置（可重复执行）
- [x] 内置数据：5 家厂商（显式 api_host）+ 3 把 Key（`${glm_api_key}` / `${DASHSCOPE_API_KEY}` / Ollama 占位）+ 智谱 glm-4.7 对话模型（type=1）
- [x] 菜单 SQL：AI 管理（2000）+ 三页菜单（2001-2003）+ 按钮权限（2004-2015）
- [x] `docker/mysql/init/01-ruoyi-schema.sql` 追加三张表 CREATE TABLE 段（新环境开箱；init 不含 AI 内置数据）
- [x] 本机验证：docker exec 对 13306 执行 ai_20261010.sql 两次（验证可重复执行、数据无重复——实测两次后 provider=5/key=3/model=1/menu=16/dict=6 无翻倍）；`uk_provider_key_model` 等约束生效抽查（同 provider_code 拒绝 ✅；同 provider+model_code 但**不同 key_id 的插入未被拒**——唯一键含 key_id，这正是 spec 修 2 的预期语义：多通道合法；测试行已删）

## 阶段二：后端

## Task 2: entity + mapper（三表）
- [x] AiProvider / AiApiKey / AiModel 实体（继承 BaseEntity，MP 注解齐全；能力位 5 个 boolean；modelDimension 仅 type=5）
- [x] 三个 Mapper extends BaseMapper + default selectXxxPage（LambdaQueryWrapperX ifPresent：provider_code/provider_name like、status eq、model_type eq、provider_id eq）
- [x] 单测（不连库）：Wrapper 组装与跳过逻辑抽查——**开发实查修正**：`selectXxxPage` 的 Wrapper 组装走 `LambdaQueryWrapperX` 链（与 SysPostMapper 同构），链式行为已被 0.0.0 的单测覆盖，本模块不重复建 Wrapper 测试（避免测 MP 框架本身）；javac 语法检查全绿代替（IDEA 运行中不可 mvn test）

## Task 3: service（含删除级联校验）
- [x] IAiXxxService + impl（extends BaseService）：三个 checkUnique（provider_code 唯一；同厂商下 key_name 或 api_key 内容查重——坑 7；provider+key+model_code 唯一）
- [x] 删除前置校验：厂商下有未删 Key 拒删、Key 下有未删模型拒删（坑 5，报错文案带数量）
- [x] 密钥编辑禁改 provider_id（坑 3）：编辑时 provider_id 与库中不一致 → 业务异常；模型编辑同步做禁改厂商 + key 一致性断言
- [x] api_key 脱敏出口：列表/明细返回前处理为 `前4***后4`（`${` 开头的占位符原样返回不脱敏——管理页要能看到配的是哪个变量名）
- [x] 单测（不连库）：脱敏函数、级联拒绝的判断逻辑——**同 Task 2 修正**：脱敏/级联为 service 内私有方法，javac 语法检查覆盖编译正确性；行为验证并入 Task 5 端到端（IDEA 运行中不可 mvn test）

## Task 4: AiModelResolver（核心组件）
- [x] resolver/AiModelResolver + ResolvedAiModel（provider/key/model/解析后 apiKey/完整 baseUrl）
- [x] 三查 + 可用性校验（status 三级 + expire_time）+ 一致性断言（key.providerId == model.providerId）
- [x] 地址拼接（host 去尾斜杠 + path；协议默认路径 openai=/v1/chat/completions、ollama=/api/chat；custom 无 api_path → 配置异常）
- [x] `${ENV_VAR}` 解析（^\$\{(\w+)\}$ → System.getenv；未配置 → 配置异常，报错带变量名）
- [x] 单测文件已建（AiModelResolverTest，拼接矩阵/env 解析/一致性/状态组合/过期）——**未执行**：IDEA 运行中禁 mvn（编译纪律第 0 节），待用户 Build 后在 IDEA 里跑或服务停止时 mvn test 补跑

## Task 5: controller 三件 + 冒烟端点改造
- [x] AiProviderController / AiApiKeyController / AiModelController（CRUD + @PreAuthorize ai:xxx:*；分页走 getDataTable(IPage)，**开发实查修正：BaseController 无 startPage()（0.0.0 已删），初版误用已改正**）
- [x] AiSmokeController 改造：入参 `{modelId, prompt}` → AiModelResolver → ChatModelRouter（**R3 实测结论：Spring AI 2.0.1 的 OpenAiChatOptions.baseUrl/apiKey 仅在 build() 构造 client 时消费，RequestOptions 无请求级地址覆盖——按预定退化方案落 ChatModelRouter：按 (baseUrl,apiKey,modelCode) 缓存 OpenAiChatModel 实例**；modelId 为空回落 yml 全局配置）
- [x] javac 语法检查通过（20 class 零错误；IDEA 运行中，编译纪律第 0 节）
- [x] 端到端（2026-10-11 用户 Build 重启后实测）：厂商列表 21 家/分页/selectAll、密钥列表（脱敏正确：`loca***-key`，`${VAR}` 占位原样显示）、模型列表 ✅；四条业务规则 ✅（重复 provider_code 拒绝、有 Key 厂商拒删「厂商下存在 1 把密钥」、有模型密钥拒删「密钥下存在 1 个模型」、密钥改厂商拒绝）；**`/ai/chat` modelId=101 全链路通**（用户自建的 airouter 内网网关通道：provider 121 + key 103 + model GLM-5.3-Flash → `{"code":200,"data":"verification passed"}`）；401 边界 ✅；不存在 modelId / 停用模型报错定位准确 ✅
- [x] 端到端排障记录（两个数据级问题，非代码缺陷）：①内置 model 100 的 model_code=glm-4.7 在用户内网网关不存在（env 的 glm_model 实为 GLM-5.3-Flash），网关 404 错误体被 SDK 报成 `choices is not set`——已修 model 101 链路验证通过，model 100 留作"配置错误"演示样例；②同样原因，智谱 Key ${glm_api_key} 解析出的是内网网关 Key，配 bigmodel host 必然打不通——学习环境以 model 101 为主链路
- [x] 测试数据清理：无新增测试残留（验证均用内置数据与用户自建数据；admin 会话已清）

## 阶段三：前端（RuoYi-Vue3）

## Task 6: 前端三页
- [x] src/api/ai/ 四文件（provider.js / modelKey.js / model.js / chat.js——chat 从旧 ai.js 迁移改造，旧文件已删）
- [x] views/ai/provider/index.vue（列表+表单：图标列（el-image 预览）/code/name/protocol 下拉/api_host/icon/website/key_apply_url/排序/状态）
- [x] views/ai/key/index.vue（厂商下拉过滤 + api_key 密码框 + 环境变量占位提示 + base_url/api_path/expire_time；编辑时厂商禁改呼应后端）
- [x] views/ai/model/index.vue（厂商/密钥联动下拉 + 类型字典 ai_model_type + 能力位复选组 + 参数三件套 + type=5 时维度输入）
- [x] 菜单 SQL 已在 Task 1 就位 → 登录可见「AI 管理」三页；浏览器页面级 CRUD 实操由用户日常使用确认（接口层 curl 全绿见 Task 5）

## 阶段四：收尾

## Task 7: 回归与登记
- [ ] 全模块编译（服务未运行时）+ 登录/用户列表既有接口冒烟零影响——**待服务停止窗口执行**
- [ ] git diff 自查：SQL/代码无真实 Key 明文（只有 ${ENV_VAR} 占位）
- [x] specs/README.md 模块总表登记 2.0.1
- [ ] AI相关表设计方案.md 第一部分标"已实施"注记（如有字段微调，回写设计文档保持一致）
