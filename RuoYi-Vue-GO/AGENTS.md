# AGENTS.md（RuoYi-Vue-GO）

本文件是 GO 版子目录的 AI 编码规范入口与**规范全集**。工作区级规范（兼容契约、勾选纪律细则、共用库纪律）以根目录 [../AGENTS.md](../AGENTS.md) 为唯一规范源；GO 版任务台账（模块总表/进度/动工检查单）在 [specs/README.md](specs/README.md)，规范与台账分离，不重复维护。

## 项目定位

- 用 Go 复刻 Java 版 RuoYi-Vue 服务端接口，前端 RuoYi-Vue3 零改动切换（定位与 RuoYi-Vue-FastApi 相同，同一套系统三语言实现）。
- **当前状态：未动工，spec 驱动开发**。任务总清单见 [specs/README.md](specs/README.md)（12 个模块 0.0.0~11.0.0），技术选型结论见 [specs/tech-stack.md](specs/tech-stack.md)，与 Java 版的有意差异登记在 [specs/deviations.md](specs/deviations.md)。

## 技术栈（已定，依据 specs/tech-stack.md）

Gin + GORM + go-redis v9 + golang-jwt v5（HS512）+ viper + zap + excelize + base64Captcha + swaggo；BCrypt 用 `golang.org/x/crypto/bcrypt`。目录骨架采用 Go 社区惯例 `cmd/ + internal/ + pkg/`，分层 handler → service → dao 与 Python 版同构——目录结构唯一约定见 [specs/0.0.0-工程基础/spec.md](specs/0.0.0-工程基础/spec.md)。

## 开发约定（每个模块共同遵守）

1. **动工前**：对目标模块执行 [specs/README.md](specs/README.md)《动工检查单》7 条，把该模块三件套（spec/tasks/checklist）写实后再开发。
2. **契约先行**：开发前先读 Java 版对应 Controller + ServiceImpl + Mapper XML，把端点的路径、方法、参数、返回 JSON 结构写进该模块 spec.md 的 API 清单（表格四列起：方法/路径/参数要点/返回要点），不凭记忆（范本：2.0.0-登录闭环）。
3. **验收以前端为准**：接口完成的标准是 RuoYi-Vue3 对应页面能正常操作，响应字段名（驼峰）、状态码约定与 Java 版一致。
4. **状态码约定**（对位 Java AjaxResult + GlobalExceptionHandler）：全部 HTTP 200；成功 body code 200；业务失败 code 500；警告 code 601；无权限 code 403；未登录 code 401。前端按 body code 判断，不得改成 HTTP 状态码语义。
5. **权限标识**：每个端点标注 Java `@PreAuthorize` 对应的权限字符串（如 `system:user:list`），由 1.0.0-基础设施 权限机制实现。
6. **横切能力统一供给**：数据权限过滤（data_scope）、Excel 导出（excelize）、防重/限流/XSS 均由 1.0.0-基础设施 统一提供，业务模块只声明需求与列定义，不自行实现。
7. **测试分层**：纯逻辑（树构建、格式转换、cron 解析等）写 `go test` 单元测试，不连库；涉及数据库/Redis 的用真实环境端到端验证（对齐 Python 版方式）。
8. **差异登记**：无法逐字节对齐 Java 的行为必须登记进 [specs/deviations.md](specs/deviations.md)（差异点/Java 行为/GO 行为/原因）；影响前端契约的差异（字段名/结构/状态码）不允许存在，总验收对着它过。
9. **勾选纪律**：做完即勾（含对应单元测试跑通）、没做不勾并在行尾注明原因、三件套全部勾选才许在总表标 ✅ + 日期；宣称完成前逐条自查。发现的坑写进该模块 spec.md"实施记录"（细则见根 AGENTS.md"spec checklist 纪律"）。

## 常用命令（动工后生效，骨架由 0.0.0-工程基础 Task 8 固化为 Makefile）

```bash
go run ./cmd/server            # 启动，监听 8080（需 .env.dev，值与 Java 版配置同步）
go test ./...                  # 全部单元测试（纯逻辑不连库）
gofmt -l . && go vet ./...     # 提交前自查
```

## 环境约束（重要）

1. **端口互斥**：Go 版同监听 8080，Java / Python / Go 三版**同一时间只能运行一个**，切换前端代理指向前先停掉另外两个。
2. **共用库纪律**：MySQL（阿里云 RDS `ry-vue-26-09-24`）和 Redis（db11）与 Java、Python 版共用。配置以 Java 版 `application.yml` / `application-druid.yml` 为准，Java 端改动后手动同步 Go 版 `.env.dev`。端到端测试产生的数据**测试完必须清理**，admin 密码与会话测完必须复原（admin/admin123）。
3. **会话不互通（已知设计，不是 bug）**：三个版本的 Redis 会话 value 格式各不相同（Java 是 FastJson 带 @type，Python/Go 是纯 JSON 但结构有差异），切换后端后所有用户需重新登录，详见 [specs/deviations.md](specs/deviations.md)。
4. **Go 特有兼容坑**：time.Time RFC3339 需自定义驼峰时间输出、json tag 显式驼峰、int64 主键不转 string 等，见 [specs/tech-stack.md](specs/tech-stack.md) 第四节。
