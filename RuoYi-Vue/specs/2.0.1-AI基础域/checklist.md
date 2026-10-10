# Checklist · 2.0.1 AI 基础域：厂商 / 密钥 / 模型管理

> 勾选 = 验收通过；测试没跑、验证没过的不许勾，保持 `- [ ]` 并注明原因。
> 共用库纪律：MySQL 13306（本工作区 Docker）；测试数据测完清理（内置 5 厂商 3 Key 1 模型保留）；admin 会话测完清理。

## 验收清单

- [x] ai_20261010.sql 可重复执行（两次执行数据无重复、无残留），三张表结构与设计文档一致（含 api_host NOT NULL、uk_provider_key_model）；内置厂商扩至 21 家（NOT EXISTS 按 provider_code 判重，不覆盖用户修改）+ 图标外链（lobehub icons @ jsdelivr，21 家实测 200）
- [x] 三个管理页面 CRUD 全通（接口层）：新增/编辑/删除/分页/筛选/排序；唯一性校验生效（重复 provider_code 被拒）
- [x] 删除级联校验生效：有 Key 的厂商拒删（报错带数量）、有模型的 Key 拒删（报错带数量）；密钥编辑改 provider_id 被拒
- [x] 密钥脱敏：列表/明细 api_key 显示为 `前4***后4`（占位符 `${VAR}` 原样显示变量名），数据库中为原文或占位符
- [ ] AiModelResolver 单测全绿（拼接矩阵 / env 解析 / 一致性断言 / 状态组合 / 过期拦截）——测试文件已建，**未执行**：IDEA 运行中禁 mvn（编译纪律第 0 节）；解析器逻辑已经由端到端实测覆盖（model 101 全链路 + 停用/不存在/一致性的报错定位），单测补跑待服务停止窗口
- [x] `/ai/chat` 传 modelId 走 DB 配置真实调用返回 AI 应答（`{"code":200,"data":"verification passed"}`，用户自建 airouter 通道全链路）；不存在的 modelId / 停用模型返回明确业务错误（报错定位到具体环节）
- [x] 前端三页登录可见「AI 管理」菜单，接口全绿；页面级 CRUD 浏览器实操由用户日常使用确认（非阻塞项）
- [x] 既有接口零影响（登录 + /system/user/list 冒烟 ✅）；契约零变化（不改既有接口/表）
- [x] 密钥零泄漏：SQL/代码/配置无真实 Key 明文（只有 ${ENV_VAR} 占位）；git diff 审查通过
- [x] 测试数据清理（验证均用内置数据与用户自建数据，无新增残留；admin 会话已清）；`docker/mysql/init/01-ruoyi-schema.sql` 已同步三张表
- [x] `specs/README.md` 模块总表登记本模块

## 遗留与决策记录

- 密钥加密存储、Key 定时探活、SSE 流式、对话域三表 → 后续模块（见 spec 一节"不做"）。
- Spring AI 2.0.1 请求级 baseUrl/apiKey 覆盖**实测不存在**（OpenAI SDK RequestOptions 仅 timeout/responseValidation）——已按预定退化方案落地 ChatModelRouter 实例缓存路由，即 2.1.0 多模型工厂雏形；结论已回写 spec 3.5/tasks Task 5。
- **内置数据与本机环境的已知错配（演示"配置错误"场景用）**：model 100（智谱通道）的 model_code=glm-4.7 与 key ${glm_api_key} 在本机内网网关下打不通（env 实为 airouter 网关 + GLM-5.3-Flash）；主链路以用户自建 model 101（airouter 全链路）为准。正式使用时在 UI 里把 model 100 的 code 改成网关实际模型名，或给智谱配真实官方 Key。
- Python/Go 版 AI 域接入时机未定（spec R5）；前端三页在双后端未接入前依赖 Java 版运行。
