# Tasks · 12 平台标识

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行《动工检查单》：本模块为新增端点（无 Java 既有 Controller 可对照），契约按三版对称设计写实（见 spec.md）（2026-09-25）
- [x] Task Java 基准：新建 `SysPlatformController`（GET /getPlatformInfo，登录即可，无 @PreAuthorize）（2026-09-25，mvn compile 通过）
- [x] Task GO：login_handler 新增 `GetPlatformInfo`，常量 + runtime.Version 组装，路由注册 + 白名单不加（需登录）（2026-09-25，go build/test 全绿）
- [ ] Task Python：login_controller 新增 `GET /getPlatformInfo`（对应任务书 spec-11）
- [ ] Task 前端：`src/api/platform.js` 新建；druid / server 两页按 features 降级提示（不关菜单、不改路由）
- [ ] Task 文档：deviations #17 更新（服务监控定版为「平台提示」方案）；根 readme 开发方式一节补跨端模块说明
