# 01 基础设施：分页 / 权限 / 操作日志 / 数据权限 / Excel / 防重限流 / XSS / 文档兼容

> **状态：✅ 已完成（2026-09-24）**
>
> 所有业务模块的公共依赖，必须最先完成（对位 Python 版 spec-01，能力清单一致）。
> Java 版对应：`BaseController` + PageHelper（分页）、`@PreAuthorize @ss.hasPermi`（权限）、`@Log` + LogAspect（操作日志）、DataScopeAspect（数据权限）、ExcelUtil（导出）、`@RepeatSubmit` + `@RateLimiter`（防重/限流）、XssFilter、springdoc。
> 实现手法：全部做成 **gin 中间件或 handler 包装函数**（对位 Python 版装饰器/依赖式）。

## 目标

提供业务模块开发所需的横切能力，行为对齐 Java 版，接口签名对 Go 侧保持简洁（中间件 / 闭包包装）。

任务分解见 [tasks.md](tasks.md)，验收清单见 [checklist.md](checklist.md)。

## 实施记录

- Task 3（操作日志）：测试原共享 `file::memory:?cache=shared` 内存库导致跨测试数据污染，改为每测试独立命名内存库；method 名断言 `aspect_test` 前缀修正为 `aspect.`。
- Task 4（数据权限）：UserAlias 默认空串对齐 Java 注解默认值（非任务原文的 u），调用点显式传；条件值参数绑定防注入；角色读会话内嵌 user.roles[]，无需查库。
- Task 5（Excel）：excelize SetColWidth 参数是列名非单元格名；导出响应头三元组逐字对位 FileUtils.setAttachmentResponseHeader；包不依赖 gin。
- Task 6（防重限流）：防重返回 code 500 非 601（对位 AjaxResult.error，任务原文 601 有误，已按 Java 纠正）；Lua 脚本逐字移植 miniredis 验证；RedisCache 新增 EvalInt。
- Task 7（XSS）：清洗语义=正则剥标签（对齐 Python 版），标签间文本保留，非 Java HTMLFilter 逐字复刻。
- Task 8（文档）：ginSwagger.URL 必须相对路径 "doc.json"（iframe /dev-api 前缀代理场景绝对路径 404）；补齐 configs/.env.dev + .env.example + gitignore。
- 依赖新增：excelize v2.11.0、gin-swagger v1.6.1 + swag v1.16.6 + swagger files、miniredis v2.33.0（测试）。
