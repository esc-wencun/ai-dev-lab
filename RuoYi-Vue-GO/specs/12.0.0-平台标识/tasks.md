# Tasks · 12 平台标识

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：本模块为新增端点（无 Java 既有 Controller 可对照），契约按三版对称设计写实（见 spec.md）（2026-09-25）
- [x] Task Java 基准：新建 `SysPlatformController`（GET /getPlatformInfo，登录即可，无 @PreAuthorize）（2026-09-25，mvn compile 通过）
- [x] Task GO：login_handler 新增 `GetPlatformInfo`，常量 + runtime.Version 组装，路由注册 + 白名单不加（需登录）（2026-09-25，go build/test 全绿）
- [x] Task Python：login_controller 新增 `GET /getPlatformInfo`（对应任务书 spec-11）（2026-09-25）
- [x] Task 前端：`src/api/platform.js` 新建；druid / server 两页按 features 降级提示（不关菜单、不改路由）（2026-09-25；浏览器页面级验收待用户确认，见 checklist）
- [x] Task 文档：deviations #17 更新（服务监控定版为「平台提示」方案）；根 readme 开发方式一节补跨端模块说明（2026-09-25）
- [x] 实施记录（2026-09-25）：
  - Java 路径上 java.exe 默认是 JDK 1.6，运行 jar 需显式用 `C:\Program Files\Java\jdk-17.0.12\bin\java.exe`
  - Go 版验证码答案在 Redis 是 JSON 字符串（SetObject json.Marshal 带引号），脚本读答案需 json.loads 解码；Python/Java 均为裸文本
  - 前端验证：开发环境浏览器预览不可用，用 vite 模块转换输出 + /dev-api 代理 curl 链路验证替代；页面级确认留给用户
  - 二期：Java 一期验证后仅 compile 未 package，8081 端到端首跑撞上旧 jar（缺 swaggerDocs 字段）——改 Java 代码后必须重新 package 再端到端
  - 二期：用户正用 VS Code debugpy 占 8080 跑 Python 版，不动用户进程；Go/Java 改用 APP_PORT=8081 / --server.port=8081 临时实例验证，验完即停
- [x] 二期增量（2026-09-25）：系统接口（tool/swagger）纳入平台降级机制
  - [x] 契约侦察：前端 iframe 加载 `/swagger-ui/index.html`；Go swaggo paths 为空、Python 无该路由（404）——写入 spec.md（2026-09-25）
  - [x] Java：SysPlatformController features 加 `swaggerDocs: true`（2026-09-25，重打包后端到端验证通过）
  - [x] GO：GetPlatformInfo features 加 `swaggerDocs: false`（2026-09-25，APP_PORT=8081 临时实例端到端验证，避开用户 8080 调试进程）
  - [x] Python：/getPlatformInfo features 加 `swaggerDocs: False`（2026-09-25，对用户 debugpy 运行实例直接验证通过）
  - [x] 前端：tool/swagger/index.vue 按 swaggerDocs 降级提示（2026-09-25，vite 编译验证通过）
  - [x] 三版端到端重验（新字段逐一断言）+ 文档同步（deviations/spec-11/readme）（2026-09-25）
