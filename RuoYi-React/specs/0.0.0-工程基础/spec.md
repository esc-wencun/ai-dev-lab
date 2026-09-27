# Spec 0.0.0 工程基础：Vite + React 19 + TS 脚手架

>
> **状态：✅ 已完成（2026-09-27）**。工程骨架全部落地并验证：dev 8090 起服务 + 占位页渲染、`/dev-api` 代理链路 200、`build:prod` 通过、Vitest 冒烟通过、svg 精灵（从 RuoYi-Vue3 同源复制 90 个）浏览器可见、工作区四处文档同步完成。React 19 + antd 无兼容告警，未引入 v5-patch。
> **背景**：本项目首个模块，建立可运行骨架。选型已定稿（见 [../tech-stack.md](../tech-stack.md)），本 spec 只做工程落地不做调研。
> **契约侦察**：Vite 配置与环境变量基线见 [../reference/02-infra-contract.md](../reference/02-infra-contract.md) §10~§11；svg 精灵配置与 Vue3 版 `vite/plugins/svg-icon.js` 对齐。
> **依赖**：无。

## 范围

- Vite `react-ts` 脚手架 + 全量依赖安装（选型表所列）。
- Vite 配置：dev 端口 **8090**、代理契约、别名、svg 精灵。
- TypeScript strict 配置 + 环境变量类型。
- 目录骨架占位 + `bin/` Windows 脚本。
- Vitest 接入（冒烟）。
- 工作区文档同步（根 AGENTS/readme + RuoYi-Vue3 AGENTS + 本项目 AGENTS.md）。

## 关键实现要点（与基准对齐项）

1. **代理契约必须一致**：`/dev-api` → `http://localhost:8080`（changeOrigin + rewrite 剥离前缀）；`^/v3/api-docs/(.*)` 同目标**不剥离**（springdoc）。这是与三版后端互通的生命线（reference/02 §12 易错清单 #10）。
2. **svg 精灵同源**：`vite-plugin-svg-icons`，`iconDirs: [src/assets/icons/svg]`，`symbolId: 'icon-[dir]-[name]'`；约 90 个 svg **从 RuoYi-Vue3/src/assets/icons/svg 全量复制**——DB 菜单 `meta.icon` 存的是精灵名，资源不同源则菜单图标全丢。
3. **环境变量三份**：`.env.development`（`VITE_APP_BASE_API='/dev-api'`）/ `.env.production`（`/prod-api`）/ `.env.staging`（`/stage-api`）；`VITE_APP_TITLE=若依管理系统` 三处同步。
4. React 19 + antd 兼容告警时按官方引入 `@ant-design/v5-patch-for-react-19`（记录版本）。
5. dev 端口 8090 与 RuoYi-Vue3（80）并行——有意差异，见 [../deviations.md](../deviations.md) #6。

## 设计决策

1. 脚手架产物为 `scaffold` 子目录再上移合并（specs/ 目录已存在，`create-vite .` 会取消）。
2. `tsconfig` 保持模板 strict 基础上加 `strict: true` 与 `@` 别名；`vite-env.d.ts` 补 `ImportMetaEnv` 四字段类型。
3. 本项目不设 `.oxlintrc.json` 定制（脚手架自带 oxlint 保留默认即可），不引入 ESLint/Prettier——基准前端亦无 lint 体系，差异登记 [../deviations.md](../deviations.md) #7。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)

## 实施记录

- 2026-09-27：脚手架已创建（react-ts 模板，React 19.2.8 / Vite 8.3 / TS 6.0），包名改 `ruoyi-react`，scripts 调整（dev 加 `--port 8090`、新增 `test: vitest run`）；基础 npm install 完成。后续 Task 见 tasks.md。
