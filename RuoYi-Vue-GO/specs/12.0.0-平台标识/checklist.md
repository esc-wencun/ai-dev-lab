# Checklist · 12 平台标识

> 勾选 = 验收通过；没验证的不许勾。

## 验收清单

- [x] Java：`mvn compile` 通过（2026-09-25）
- [x] GO：`go build ./...` 通过；`go test ./...` 全绿（2026-09-25）
- [x] Python：端到端 curl 验证通过（登录 → Bearer 调 /getPlatformInfo → 200 信封字段对齐；无 token → 401 信封）（2026-09-25，字段逐一断言通过）
- [x] GO：端到端 curl 验证通过（同上，languageVersion=1.27.1）（2026-09-25）
- [x] Java：端到端 curl 验证通过（同上，framework=RuoYi / version=3.9.2 / languageVersion=17.0.12）（2026-09-25）
- [x] 前端：vite 编译通过（druid/server 两页与 platform.js 转换输出含新逻辑）；/dev-api 代理链路三版可达（2026-09-25）
- [ ] 前端：浏览器页面级验收——Java 后端下数据监控/服务监控/系统接口三页正常加载；Go/Python 后端下三页按 features 降级提示（开发环境浏览器预览不可用，待用户确认，同 02~11 遗留项）
- [x] 端到端测试数据清理完毕（登记见下）

## 测试数据清理记录

- 2026-09-25：三版端到端均为只读端点 + 登录会话，login_tokens:/captcha_codes: 键测完即删；一次 Go 版登录失败（验证码解码问题，脚本修正后通过）与各次登录成功会写入 sys_logininfor 日志，属操作日志非业务数据，沿用既有 spec 惯例不清理。
- 2026-09-25（二期 swaggerDocs）：三版 8080/8081 各一轮端到端，会话/验证码键测完即删；Java 旧包复验共两次冷启动。无业务数据写入。
