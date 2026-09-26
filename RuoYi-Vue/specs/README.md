# RuoYi-Vue（Java 版）任务台账

> Java 版自 2026-09-25 起纳入可修改范围（按学习需要增量演进）。本目录是 Java 侧演进任务的台账，规程（契约先行 / 勾选纪律 / spec checklist 纪律）统一见工作区根 [../AGENTS.md](../AGENTS.md)，此处不重复。
> 目录结构与三件套格式沿用 [RuoYi-Vue-GO/specs](../../RuoYi-Vue-GO/specs/README.md) 约定：每个模块一个文件夹，内含 spec.md（契约与方案）/ tasks.md（原子任务）/ checklist.md（验收清单）。

## 序号约定

按根 AGENTS.md《开发流程（spec 驱动）》规则，**各子项目 spec 编号相互独立**：Java 侧序列自成一体，从 **0.0.0** 开始递增（与 GO 版同一惯例）；跨四端模块引用契约主文档时注明所在项目编号（如「GO 版 12.0.0」）。

## 模块总表

| 序号 | 模块 | 内容 | 状态 | 依赖 |
|------|------|------|------|------|
| 00 | [0.0.0-MyBatis-Plus功能增加](0.0.0-MyBatis-Plus功能增加/spec.md) | 引入 MyBatis-Plus：删 PageHelper 全面切 MP 分页 + BaseService 抽取公用方法 | ✅ 2026-09-26 | 无 |
| 01 | [1.0.0-数据权限MP原生适配](1.0.0-数据权限MP原生适配/spec.md) | 纯 MP 查询（无 XML）的 DataScope：DataPermissionInterceptor + 规则白名单 + 范围 DTO（dcoa 校准） | ✅ 2026-09-26 | 0.0.0 |

## 动工检查单（每个模块动工前逐条执行）

1. [ ] 通读本模块 spec.md 的技术方案与风险清单，确认方案仍然成立（依赖版本、Spring Boot 版本没有变化）
2. [ ] 按根 AGENTS.md《已踩过的坑》与《环境约束》自查：端口互斥（8080 只跑一个后端）、共用库纪律、jdk17 全路径
3. [ ] 把 tasks.md 的 Task 与当天实际改动范围对齐后再动手
4. [ ] 与本工作区其他三端（前端/Go/Python）的影响面：改接口契约才需要同步，只动分层内部实现不需要——本模块预期契约零变化
