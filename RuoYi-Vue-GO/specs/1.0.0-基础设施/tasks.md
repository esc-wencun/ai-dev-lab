# Tasks · 01 基础设施

> 勾选纪律：做完即勾（含对应单元测试跑通）；没做的不许勾，行尾注明原因；勾选 = 验收通过。
> 依赖：0.0.0-工程基础（RedisCache 门面 / 响应信封 / 统一错误处理 / 常量）。

## Task 1: 分页（对应 PageHelper startPage + getDataTable）

- [x] `pkg/utils/page/`：`Paginate(ctx, db, query, isPage)` 通用封装——
  - [x] 读 query params `pageNum`（默认 1）、`pageSize`（默认 10）
  - [x] `orderByColumn` 驼峰转下划线 + `isAsc`（asc/desc 默认 asc），**白名单防注入**
  - [x] 返回 `{code:200, msg:查询成功, rows:[], total:n}`，与 Java TableDataInfo 一致
- [x] 单元测试：排序字段转换、非法字段拒绝、total 计算（不连库，构造条件断言）
  - 注：Paginate 签名落为 `Paginate(ctx, query, dest, Domain, isPage)`，Count/Find 各自 Session 克隆；sqlite 内存库验证 total/偏移/倒序/带条件分页；isAsc 未知值比 Java 更严格直接拒绝

## Task 2: 认证与权限（对应 JwtAuthenticationTokenFilter + PreAuthorize）

- [x] 认证中间件：从 `Authorization` 头取 token（剥 `Bearer `）→ golang-jwt 解析（HS512，secret 同 Java）→ 取 claim `login_user_key` → RedisCache 查 `login_tokens:{uuid}` 会话 → 会话存入 gin context（`c.Set`）
  - [x] 解析失败/会话不存在：不直接拦截放行链，转由权限检查返回 401 信封（对齐 Java 行为：匿名可过白名单，受保护端点 401）
  - [x] **剩余有效期 < 20 分钟自动续期**（对位 TokenService.verifyToken + MILLIS_MINUTE_TWENTY）
- [x] 白名单机制：`/login`、`/captchaImage`、`/register`、`/` 等免认证端点集中配置（对位 SecurityConfig permitAll）
  - 注：落地为 `middleware.AuthWhitelist` map（swagger 路径 Task 8 接入）
- [x] `RequirePerm(perm string)` gin 中间件工厂：admin（userId=1）放行；从会话 permissions 匹配（`*:*:*` 与层级通配语义对齐 Java PermissionService.hasPermi）；失败 403 信封
- [x] `RequireRole(role)` 同款（对位 @ss.hasRole）
- [x] 端到端：ry 用户（common 角色）访问 `system:user:list` 得 403，admin 放行，无 token 得 401
  - 注：6 个测试全绿（httptest+miniredis）；`LoginUser.User` 用 `json.RawMessage` 原样保留，续期回写不丢 SysUser 字段，且可解析 Java FastJson 带 `@type` 的会话（有测试）；secret 语义对齐 Python 已验收基准——原文 UTF-8 字节直接作 HMAC key，不做 base64 解码（会话不互通前提下 JWT 跨版互通无意义）；`RedisCache` 新增 `SetRaw`（测试模拟 Java 原始 JSON 用）；golang-jwt v5.2.1 已入 go.mod；顺手 gofmt 修复 0.0.0 遗留 5 文件格式

## Task 3: 操作日志（对应 @Log + LogAspect）

- [x] DO：`model/do/sys_oper_log.go`（对位 sys_oper_log 表字段，逐字对表 Java SysOperLog.java 的 17 个落库字段；businessTypes 是 Java 查询用非落库字段不复刻）
- [x] `OperLog(title, businessType)` handler 包装：异步落库（goroutine + recover，对位 AsyncManager）——操作人、URL、method、IP、参数（截断 2000）、结果（截断 2000）、耗时、status（0/1）、error_msg
- [x] 异常也记录（status=1）后继续上抛；敏感参数 `password` 过滤（对位 Java 日志过滤）
  - 注：落地含 panic 场景（recover 记录后 re-panic 交 Recovery 兜底）、@Log 属性全量（operatorType/isSaveRequestData/isSaveResponseData/excludeParamNames）、body 快照回放（handler 可正常读 body）、form-urlencoded 按 parameter map 记录
- [x] 单元测试：参数脱敏、截断逻辑
  - 注：8 个测试全绿（含端到端成功/异常/panic/query/form/保存开关）。修复记录：① 测试原用共享 `file::memory:?cache=shared` 内存库导致跨测试数据污染，改为每测试独立命名内存库；② method 名断言误写 `aspect_test`（外部测试包命名），实际同包测试为 `aspect.` 前缀；③ gofmt 补齐 2 文件格式

## Task 4: 数据权限（对应 DataScopeAspect）

- [x] `internal/module/admin/aspect/data_scope.go`：按角色 data_scope（1 全部/2 自定/3 本部门/4 本部门及以下/5 仅本人）生成 GORM where 条件，别名占位对齐 Java（`dept_alias` 默认 d、`user_alias` 默认 u），自定权限走 sys_role_dept
  - 注：契约差异——UserAlias 默认空串（对齐 Java 注解默认值，非 u）；需要用户过滤的调用点显式传 `UserAlias: "u"`。条件值一律参数绑定（防注入），SQL 文本逐条对位 Java format 模板（含多自定权限合并 IN、scope=5 无别名时 dept=0、无生效角色时 dept=0、ALL 短路清空）。角色来源=会话内嵌 user.roles[]（Java 登录时 getMenuPermission 已按角色填 permissions，会话里就有，无需查库）
- [x] 注入方式：GORM scope 函数（`db.Scopes(DataScope(...))`），业务 dao 调用时声明
- [x] 单元测试：五种 data_scope 生成的 SQL 条件断言
  - 注：11 个测试全绿（五 scope 条件文本+参数断言、去重/停用/权限过滤、ALL 短路、别名默认与覆盖、无会话防御、sqlite 内存库 count 集成验证）

## Task 5: Excel 导入导出（对应 ExcelUtil）

- [x] `pkg/utils/excel/`：基于 excelize——输入列定义（标题/字段名/字典转换/类型）+ 行数据，输出 xlsx 流；`POST /xxx/export` 直接流式下载（前端 blob 接收）
  - 注：Column 对位 @Excel 必需属性（Title/Field/Converter/Width 默认 16）；表头样式灰底白字加粗居中+细边框对位 annotationHeaderStyles；响应头三元组（Content-Disposition/download-filename/Expose-Headers，percentEncode 空格转 %20）逐字对位 FileUtils.setAttachmentResponseHeader；包不依赖 gin（net/http 接口，gin.Context 直接可用）
- [x] 导入：解析 xlsx → []map，供业务校验（5.0.0-用户管理导入是首个使用者）
  - 注：按表头标题定位列（容忍标题前置行，对齐 Python 版）；Converter 反查（显示值→编码，查不到保留原值）；空行跳过；全部 string 承载由业务侧转类型
- [x] 单元测试：列定义转表头、数据格式化、读回验证
  - 注：7 个测试全绿（响应头三元组断言、读回表头/数据/Converter 正查、导入反查/编码原值、无表头报错、模板下载、错误信封）。修复记录：excelize SetColWidth 参数是列名（"A"）不是单元格名（"A1"）；读回空尾列被截断需补齐断言。excelize v2.11.0 已入 go.mod

## Task 6: 防重复提交 + 限流（对应 @RepeatSubmit + @RateLimiter）

- [x] `PreventRepeatSubmit(interval, message)` 中间件：Redis key `repeat_submit:`，值=url+token+参数摘要；interval 毫秒内重复返回 601
  - 注：实现语义对位 SameUrlDataInterceptor：key=repeat_submit:{url}{token}，值 {url:{params,time}} TTL=interval；参数=body（快照回放，handler 可正常 bind）或空 body 时取 query；判重=参数相同且间隔<interval；返回注解 message 默认"不允许重复提交，请稍候再试"。**信封偏差更正**：Java AjaxResult.error(message) 是 code 500 非 601，实现按 500 对齐（601 仅 @Log BusinessType 警告场景），tasks 原文 601 有误
- [x] `RateLimiter(count, time, limitType)` 中间件：**Lua 脚本限流**（移植 Java RedisConfig.limitScriptText 语义），key 前缀 `rate_limit:`，支持全局/IP 维度；超限返回 `访问过于频繁，请稍候再试`
  - 注：Lua 脚本逐字移植（miniredis 真实执行验证）；key=rate_limit:{key}{ip-}{handlerIdent}（handlerIdent 由路由注册时显式传入，对位 类名-方法名）；Redis 异常时返回"服务器限流异常"信封（差异：Java 抛 RuntimeException 500，Go 同样回 500 信封阻断——未采用放行策略，保持对齐）
- [x] 单元测试：Lua 计数逻辑（miniredis）
  - 注：6 个测试全绿（同参拦截/默认文案、异参/异 token/窗口后放行、GET query 判重、Lua 计数 3 次窗口、IP 维度独立、窗口重置）。RedisCache 新增 EvalInt 门面方法

## Task 7: XSS 输入过滤（对应 XssFilter）

- [x] `pkg/utils/xss/`：JSON body 字符串字段 HTML 标签清洗，中间件按路由组启用；排除名单与 Java 一致（`/system/notice` 富文本保留）
  - 注：清洗语义=正则剥标签（对齐 Python 版 xss_util 的 `<[^>]*>` 实现，标签间文本保留——非 Java HTMLFilter 逐字复刻，差异点已核）；规则对位 Java：仅 POST/PUT、排除名单前缀匹配（默认 /system/notice）、仅 application/json body、递归清洗 dict/list（深度上限 16）；body 快照回放 handler 可正常 bind；包不依赖业务层
- [x] 单元测试：`<script>` 被清洗、排除名单不清洗
  - 注：5 个测试全绿（标签剥离、嵌套+数组清洗、排除路径富文本保留、GET/非 JSON 放行、multipart 原样）

## Task 8: API 文档兼容（前端 系统工具→接口文档）

- [x] 前端内嵌 swagger-ui 依赖 springdoc 路径：`/v3/api-docs` 返回 OpenAPI schema（swaggo 生成），`/swagger-ui.html` redirect
  - 注：swaggo/gin-swagger + swag init（main.go 注解 → docs/ 生成物）；路由 /swagger-ui/*any（UI 静态资源）、/swagger-ui.html 301 redirect、/v3/api-docs 返回 docs.SwaggerInfo.ReadDoc()。关键坑：ginSwagger.URL 必须用**相对路径 "doc.json"**——前端 iframe src=/dev-api/swagger-ui/index.html，绝对路径会绕过 vite 代理 404，相对路径随代理前缀自适应
- [x] 端到端：前端接口文档页正常加载
  - 注：已验证（curl 全链路）——直连 8080 三端点 200；经前端 vite 代理（:80 /dev-api → :8080）initializer/doc.json/静态资源全部 200，schema JSON 合法含 swagger 2.0 版本头。AuthWhitelist 新增 swagger 白名单（精确路径 + /swagger-ui/ 前缀），对位 Java SecurityConfig permitAll。补齐 configs/.env.dev（值同步 Java application.yml）+ .env.example 模板 + gitignore configs/.env.*

# Task Dependencies

- Task 2（认证权限）依赖 Task 1 之外全部可并行；Task 8 依赖 0.0.0-工程基础 路由骨架
- Task 1~7 相互独立，可并行开发；端到端联动验证放在全部 Task 完成后
