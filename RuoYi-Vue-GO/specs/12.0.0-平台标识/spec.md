# Spec 12.0.0 平台标识：`GET /getPlatformInfo` + 前端监控页降级提示

>
> **状态：✅ 已完成（2026-09-25）**。三版端点 + 前端两页降级全部实现并验证：Java mvn compile 通过、Go build/test 全绿、三版端到端 curl 逐字段断言通过、前端 vite 编译与代理链路验证通过。仅浏览器页面级操作确认待用户（同 02~11 遗留项，checklist 已注明原因）。
> **增量（2026-09-25 二期）**：系统接口（tool/swagger）菜单纳入同一机制——契约侦察发现 Python 版无 `/swagger-ui/index.html` 路由（iframe 404，仅有 `/swagger-ui.html`→`/docs` 重定向）、Go 版 swaggo docs paths 为空（页面空内容），均不等价于 Java springdoc 页面。features 新增 `swaggerDocs`，前端 tool/swagger 页降级。
> **背景**：数据监控（Druid 控制台）为 Java 特有（Druid 是 Java 连接池，Go/Python 无对应物）；服务监控 Go 版未实现（deviations #17）、Python 版已实现（其 spec-08）。对这两个菜单采用「平台标识接口 + 前端降级提示」而非关闭菜单——`sys_menu` 三版共用，改 visible 会连带关闭 Java 版菜单。
> **跨端模块**：本 spec 是契约主文档，覆盖 Java（基准新增）/ Go / Python / 前端四端；Python 版对应任务书见 [`../../RuoYi-Vue-FastApi/specs/spec-11-platform-info.md`](../../../../RuoYi-Vue-FastApi/specs/spec-11-platform-info.md)。
> **依赖**：01（认证机制）

## 契约侦察结论（2026-09-25 对照源码）

- Java 匿名/认证机制：`SecurityConfig` 默认所有路径需认证（`permitAll` 仅登录闭环 + `@Anonymous` 注解收集）；新端点不加 `@PreAuthorize` 即「登录即可访问」，与 `getInfo` 同级。
- `AjaxResult.put(String, Object)` 链式返回 this（AjaxResult.java:211）；`RuoYiConfig.getName()/getVersion()` 取 application.yml 的 `spring.name`(RuoYi) / `spring.version`(3.9.2)，SysIndexController 已在用。
- Go 鉴权：`Auth` 中间件只解析不拦截，受保护端点 handler 内 `middleware.GetLoginUser(c) == nil` → 401 信封（login_handler.go GetInfo 模式）。
- Python 鉴权：`LoginService.get_current_user(request, query_db)`，未登录抛 AuthException → `ResponseUtil.unauthorized`（login_controller.py unlockscreen 模式）。

## API 清单

| # | 方法 | 路径 | 鉴权 | 权限串 |
|---|------|------|------|--------|
| 1 | GET | `/getPlatformInfo` | 登录即可 | 无（不设 @PreAuthorize） |

返回（AjaxResult 信封，驼峰）：

```json
{
  "code": 200,
  "msg": "操作成功",
  "framework": "RuoYi-Vue-GO",
  "version": "1.0.0",
  "language": "go",
  "languageVersion": "1.22.x",
  "features": { "druidMonitor": false, "serverMonitor": false, "swaggerDocs": false }
}
```

各版取值（如实声明，不做硬编码）：

| 版本 | framework | version 来源 | language | languageVersion 来源 | druidMonitor | serverMonitor | swaggerDocs |
|------|-----------|--------------|----------|----------------------|:---:|:---:|:---:|
| Java | `RuoYiConfig.getName()`（RuoYi） | `RuoYiConfig.getVersion()`（3.9.2） | java | `System.getProperty("java.version")` | true | true | true |
| GO | RuoYi-Vue-GO | 常量 1.0.0 | go | `runtime.Version()` 去前缀 | false | false | false |
| Python | `AppConfig.app_name`（RuoYi-Vue-FastApi） | `AppConfig.app_version`（1.0.0） | python | `platform.python_version()` | false | true | false |

`swaggerDocs` 侦察依据（2026-09-25 实测）：前端 tool/swagger 页 iframe 加载 `/swagger-ui/index.html`。Java springdoc 原生提供；Go swaggo 有路由但 docs.go `"paths":{}`（注释未生成端点，页面空）；Python 仅 `/swagger-ui.html`→`/docs` 重定向、无 `/swagger-ui/index.html` 路由（iframe 404）。FastAPI `/docs` 是另一套 UI，不等于本菜单功能。

## 设计决策

1. **features 能力开关 + language 双层**：前端按 `features` 降级而非硬编码语言判断，后端如实声明能力。Python 服务监控已实现故保持可用（打开即正常展示）；Go 两功能均提示；两版数据监控均提示。若需强制 Python 服务监控也提示「仅 Java」，将其 `features.serverMonitor` 改为 false 即可，一行改动。
2. **菜单不关闭**：`sys_menu` 共库，`visible` 字段无法按后端区分；前端降级是不动共享数据的唯一方案。
3. **鉴权取「登录即可」**：返回含版本号信息，不匿名暴露；三版鉴权深度一致。
4. **Java 版与前端纳入可修改范围**（工作区纪律 2026-09-25 起）：本模块 Java/前端为基准新增改动，后续按学习需要演进，Go/Python 评估契约同步。

## 前端改动（RuoYi-Vue3）

- `src/api/platform.js`（新建）：`getPlatformInfo()`
- `src/views/monitor/druid/index.vue`：`features.druidMonitor === false` → `el-result` 提示「该功能仅 Java 版提供」，不加载 druid iframe
- `src/views/monitor/server/index.vue`：`features.serverMonitor === false` → `el-result` 提示，不请求 `/monitor/server`；否则行为不变
- `src/views/tool/swagger/index.vue`（二期增量）：`features.swaggerDocs === false` → `el-result` 提示，不加载 swagger iframe

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
