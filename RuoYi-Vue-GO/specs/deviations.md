# Deviations · 与 Java 版的有意差异清单

> 目的：前端零改动的前提是**所有差异都有意且可查**。本项目（含 Python 版历史结论）所有已知差异集中在此，总验收时逐条核对。新发现差异随时登记，禁止只留在某次对话里。
> 格式：差异点 / Java 行为 / 本项目行为 / 原因。状态：Python 版 = RuoYi-Vue-FastApi 已存在；GO 版 = 本项目。

| # | 差异点 | Java 行为 | 本项目行为 | 原因 |
|---|--------|-----------|-----------|------|
| 1 | 登录会话 Redis value 格式 | FastJson 序列化，带 `@type` 类型头 | 纯 JSON（三版互不解析） | 语言序列化体系不同，无法逐字节兼容；后果=切换后端后所有人重新登录。**已知设计，不是 bug** |
| 2 | 会话续期精度 | 每请求 verifyToken，剩余 <20 分钟续期 | 同语义实现（GO/Python 一致） | 行为对齐，无差异；登记以便核对 |
| 3 | 验证码图片 | kaptcha 渲染（干扰线/噪点风格） | Python：Pillow 自绘；GO：base64Captcha 渲染 | 视觉不同但同样人眼可读、后端校验逻辑一致，前端只展示 base64 |
| 4 | 定时任务目标字符串 | Java 反射调用 `ryTask.ryParams('ry')` | Python：注册表 + Java 语法参数解析器；GO：待 10.0.0-定时任务 设计（建议同 Python 方案） | Go/Python 无 Java 反射语义 |
| 5 | 接口文档后端 | springdoc 原生 | Python：FastAPI 自带 + 路径兼容；GO：swaggo + 路径兼容 | 前端只依赖 `/v3/api-docs` 与 `/swagger-ui.html` 可达 |
| 6 | JWT 库实现 | jjwt | Python：PyJWT；GO：golang-jwt v5 | HS512 + 同 secret + 同 claims 键名，互验通过即等效 |
| 7 | 操作日志异步机制 | AsyncManager（队列线程） | Python：asyncio.create_task；GO：goroutine + recover | 语言异步模型不同，行为语义（不阻塞请求、失败进日志）一致 |
| 8 | 静态资源/文件上传路径 | `RuoYiConfig.profile` 配置目录 | 同配置项同名行为（各版自行实现存储） | 行为对齐；差异细节在 GO 3.0.0-个人中心注册（Python 版为 spec-02-profile）落地时核对 |
| 9 | tool 演示接口 | `tool/TestController`（swagger 演示 CRUD）与 springdoc 自带端点 | **有意不复刻**（Python 版同样未做）；前端接口文档页由 1.0.0 Task 8 的 `/v3/api-docs` + `/swagger-ui.html` 路径兼容覆盖；`views/tool/build` 表单构建排除（11.0.0 已注明） | 演示性质、前端零调用；总验收做接口完整性对照时按有意排除处理 |
| 10 | IP 黑名单匹配（GO 2.0.0） | `isMatchedIp`：`;` 分隔，支持精确 IP / 通配 `*` / 网段三形态 | `,` 分隔 + 通配前缀匹配（同 Python 版语义） | 配置值兼容常见形态（单 IP/通配段均正确）；`;` 分隔的存量配置需改写为 `,`；不影响前端契约 |
| 11 | getInfo 权限变更回写（GO 2.0.0） | permissions 变化时即时 refreshToken 回写会话 | 重算后仅返回给前端，不回写会话；6.0.0 角色-菜单接线时统一补 | 权限变更后 Go 版需重新登录生效（Java 即时生效）；前端展示不受影响 |
| 12 | 首页欢迎语（GO 2.0.0） | `RuoYiConfig.name/version` 动态拼接 | 文案写死 v1.0 | 纯文本提示语，前端无解析逻辑 |
| 13 | 会话 uuid 生成（GO 2.0.0） | `IdUtils.fastUUID`（36 位带连字符）/ `simpleUUID`（32 位） | 32 位 hex 随机串（crypto/rand） | 仅作 Redis 键后缀与 JWT claim，前端原样回传，格式差异契约无感 |
| 14 | 部门排序端点（GO 4.0.0） | `PUT /system/dept/updateSort`（本 Java 版定制，保存拖拽排序） | 未实现 | 前端 dept 页面无该调用（grep 核实）；若前端实际调用再补 |
| 15 | 岗位导出（GO 4.0.0） | `POST /system/post/export`（@Excel 注解列定义） | 暂未实现，与 5.0.0 用户导出共用 Excel 设施一并补 | 前端导出按钮在权限 `system:post:export` 之后，不影响主流程 |
| 16 | 角色修改的在线权限刷新（GO 6.0.0） | `refreshPermissionByRoleId`：SCAN 全部 login_tokens，角色变更即时生效于在线用户 | 未实现——用户重新登录后生效 | Java 本版定制能力；重新登录语义对前端无感（会话键不变），高频场景（改权限）低频 |
| 17 | 服务监控（GO 9.0.0） | `/monitor/server` 返回 OSHI CPU/内存/JVM 信息 | 未实现。**2026-09-25 定版**：经 12.0.0 平台标识模块，前端服务监控页对 Go/Python 后端显示「该功能仅 Java 版提供」降级提示（Python 版已实现不受影响）；如需补齐走 gopsutil 对位 OSHI | JVM 数据无 Go 对应物；平台降级提示方案见 12.0.0-平台标识/spec.md |
| 18 | 任务并发/misfire 策略（GO 10.0.0） | Quartz concurrent 禁止并发 + misfire 补跑策略 | cron v3 无策略位：并发不限制、misfire 忽略 | 预置任务均为轻量日志型；策略语义对前端展示无影响（字段照常存储） |
| 19 | 代码生成器模板端点（GO 11.0.0） | preview/genCode/download/batchGenCode/createTable/synchDb/edit 全套 Velocity 模板生成 | 有意排除，仅实现数据层 5 端点（spec 建议降级方案） | 生成器产出 Java/Vue 代码对 Go 项目无产出价值；前端 gen 列表页可用，预览/生成按钮点击 404 属有意行为 |
| 20 | 平台标识端点（GO 12.0.0，Java 基准新增） | 无（新增 `GET /getPlatformInfo`，登录即可，返回 framework/version/language/languageVersion/features） | GO/Python 按同一契约实现，features 如实声明（GO 全 false；Python serverMonitor=true 因服务监控已实现、swaggerDocs=false 因无 `/swagger-ui/index.html` 路由） | 多语言复刻的降级提示依据；Java/前端自此纳入可修改范围（工作区纪律 2026-09-25 起），Go/Python 评估契约同步 |

## 登记规则

1. 实现 spec 时发现无法逐字节对齐 Java 的行为 → 必须登记（原因写清楚为什么对不齐 + 对前端的影响面）。
2. **影响前端契约的差异（字段名/结构/状态码）一律不允许**——那种情况说明实现错了，不是差异。
3. 每条差异须在对应 spec 的验收清单中体现核对项；总验收（final-acceptance）按本表逐条过。
