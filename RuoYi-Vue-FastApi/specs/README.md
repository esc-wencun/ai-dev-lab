# RuoYi-Vue-FastApi 功能开发总任务清单

> 目标：将 Python (FastAPI) 版服务端功能对齐 Java 版 RuoYi-Vue，前端复用 RuoYi-Vue3，不修改前端任何代码。
> 基准：Java 版代码位于 `../RuoYi-Vue`，每个 spec 开发前必须先对照对应 Controller/Service/Mapper 确认接口契约。

## 已完成（Phase 0 登录闭环，2026-09-24）

- [x] 项目骨架（config / exceptions / middlewares / utils / module_admin 分层）
- [x] `GET /captchaImage` 算术验证码
- [x] `POST /login`（验证码校验、密码错误锁定、IP 黑名单、登录日志）
- [x] `GET /getInfo`、`GET /getRouters`、`POST /logout`、`POST /unlockscreen`
- [x] TokenService（JWT HS512 + Redis 会话）、密码 BCrypt 兼容 Java 版
- [x] 应用日志趋同 logback（sys-info/sys-error/sys-user 三文件按天滚动，见 spec-01 Task 9，已提前完成）

## 技术选型对位表（Java → Python）

| Java 生态 | Python 对位 | 说明 |
|-----------|-------------|------|
| Spring Boot | FastAPI + uvicorn | Web 框架 |
| MyBatis（XML 映射） | SQLAlchemy 2.0 async（已在用） | 不引入 SQL 映射框架，也**不采用 SQLModel**（alpha 状态、与 DO/VO 分层冲突，理由见 spec-00 决策记录）；MyBatis-Plus 对位为自建薄 DAO 工具，模块数≥4 后评估 sqlalchemy-crud-plus |
| Spring Data Redis / Lettuce | redis-py asyncio + 自建 RedisCache 门面（spec-00 Task 3） | 驱动已对位，补上层封装 |
| Java enum | 标准库 enum（spec-00 Task 2） | 带字段枚举一比一复刻 |
| SLF4J + Logback | loguru（已完成对齐） | 三文件滚动 + sys-user 命名 logger |
| Jackson 驼峰序列化 | pydantic alias_generator + CamelCaseUtil（spec-00 Task 4） | |
| @Transactional | SQLAlchemy 显式事务约定（spec-00 Task 5） | |
| AsyncManager | asyncio.create_task 封装（spec-00 Task 6） | |
| Quartz | APScheduler（spec-09） | |
| kaptcha | Pillow 自绘验证码 | 已完成 |
| ExcelUtil（POI） | openpyxl（spec-01 Task 5） | |
| Velocity 模板 | Jinja2（spec-10） | |

## 任务阶段划分

| 阶段 | Spec | 内容 | 状态 | 依赖 |
|------|------|------|------|------|
| 0 | [spec-00-engineering-foundation](spec-00-engineering-foundation.md) | 工程规范基础：分层/常量枚举/Redis门面/序列化/事务/后台任务/文案/工具链 | ✅ 完成 2026-09-24 | 无 |
| 1 | [spec-01-infrastructure](spec-01-infrastructure.md) | 基础设施：分页/权限/日志/数据权限/导出/防重限流/XSS/文档兼容 | ✅ 完成 2026-09-24 | spec-00 |
| 2 | [spec-02-profile](spec-02-profile.md) | 个人中心 + 用户注册 + 通用文件上传下载 | ✅ 完成 2026-09-24 | spec-01 |
| 3 | [spec-03-dept-post](spec-03-dept-post.md) | 部门管理 + 岗位管理 | ✅ 完成 2026-09-24 | spec-01 |
| 4 | [spec-04-user](spec-04-user.md) | 用户管理 | ✅ 完成 2026-09-24 | spec-01, spec-03 |
| 5 | [spec-05-role-menu](spec-05-role-menu.md) | 角色管理 + 菜单管理 | ✅ 完成 2026-09-24 | spec-01, spec-03, spec-04 |
| 6 | [spec-06-dict-config](spec-06-dict-config.md) | 字典管理 + 参数管理 | ✅ 完成 2026-09-24 | spec-01 |
| 7 | [spec-07-notice](spec-07-notice.md) | 通知公告 | ✅ 完成 2026-09-24 | spec-01 |
| 8 | [spec-08-monitor-log](spec-08-monitor-log.md) | 监控模块：在线用户/服务监控/缓存监控/日志查询 | ✅ 完成 2026-09-24 | spec-01 |
| 9 | [spec-09-job](spec-09-job.md) | 定时任务 | ✅ 完成 2026-09-24 | spec-01 |
| 10 | [spec-10-generator](spec-10-generator.md) | 代码生成器 | ✅ 完成 2026-09-24 | spec-04~06 |

> 全部完成后执行 [final-acceptance.md](final-acceptance.md) 总验收清单。

## 开发约定（每个 spec 共同遵守）

1. **契约先行**：开发前先读 Java 版对应 Controller + ServiceImpl + Mapper XML，把每个端点的路径、方法、参数、返回 JSON 结构写进 spec 的 API 清单，不凭记忆。
2. **验收以前端为准**：每个接口的验收标准是 RuoYi-Vue3 对应页面能正常增删改查，响应字段名（驼峰）、状态码约定（业务失败 HTTP 200 + body code 500/601；无权限 code 403；未登录 code 401）与 Java 版一致。
3. **权限标识**：每个端点标注 `@PreAuthorize` 对应的权限字符串（如 `system:user:list`），由 spec-01 的权限校验依赖实现。
4. **测试分层**：纯逻辑（树构建、格式转换、字段脱敏）写 pytest 单元测试；涉及数据库/Redis 的用真实环境做端到端验证（同 Phase 0 的方式）。
5. **数据权限**：数据范围过滤（本部门/本部门及以下/仅本人）统一由 spec-01 的 data_scope 机制提供，业务模块声明自己需要的 data_scope 即可。
6. **导出功能**：Java 版大量列表有 Excel 导出（`POST /xxx/export`），统一在 spec-01 提供基于 openpyxl 的导出工具，各模块 spec 只需声明列定义。

## 部署与环境约束（重要）

1. **端口互斥**：Python 版与 Java 版同监听 8080，**同一时间只能运行一个**。切换后端前先停掉另一个，否则 vite 代理会打到错误服务。
2. **共用库纪律**：MySQL（阿里云 RDS `ry-vue-26-09-24`）和 Redis（db11）与 Java 版共用。端到端测试会产生脏数据（测试用户/角色/公告），每个 spec 的端到端测试**必须在最后清理自己造的数据**；admin 的密码与会话状态测完必须复原（admin/admin123）。
3. **登录会话不互通**：Python 版 Redis 会话 value 是纯 JSON，Java 版是 FastJson 带 @type——切换后端后所有用户需重新登录（JWT 无法互相解析会话）。属已知设计，不视为 bug。
4. **配置同步**：数据库/Redis 改动以 Java 版 `application.yml` / `application-druid.yml` 为准，Java 端更新后需手动同步 Python 版 `.env.dev`。
5. **建议**：业务模块多起来后，为 Python 版建独立的测试库 schema（如 `ry-vue-test`），`.env.test` 指向它，pytest 端到端测试统一跑在测试库上，避免污染共用库。

## 功能对照总表（Java → Python）

| Java Controller | 端点数 | Python 状态 | 所属 Spec |
|-----------------|--------|-------------|-----------|
| SysLoginController / CaptchaController / SysIndexController | 8 | ✅ 已完成 | Phase 0 |
| SysProfileController + CommonController | 6 | ✅ 已完成 | spec-02 |
| SysDeptController | 7 | ✅ 已完成 | spec-03 |
| SysPostController | 6 | ✅ 已完成 | spec-03 |
| SysUserController | 13 | ✅ 已完成 | spec-04 |
| SysRoleController | 15 | ✅ 已完成 | spec-05 |
| SysMenuController | 8 | ✅ 已完成 | spec-05 |
| SysDictTypeController + SysDictDataController | 11 | ✅ 已完成 | spec-06 |
| SysConfigController | 8 | ✅ 已完成 | spec-06 |
| SysNoticeController | 9 | ✅ 已完成 | spec-07 |
| SysOperlogController + SysLogininforController | 8 | ✅ 已完成 | spec-08 |
| SysUserOnlineController + ServerController + CacheController | 12 | ✅ 已完成 | spec-08 |
| SysJobController + SysJobLogController（quartz） | 12 | ✅ 已完成 | spec-09 |
| GenController（代码生成） | 10 | ✅ 已完成 | spec-10 |
