# AGENTS.md（RuoYi-React 前端）

本文件是 RuoYi-React 子目录的 AI 编码规范入口与**规范全集**。工作区级规范（三版兼容契约、勾选纪律、编码行为纪律、已踩过的坑）以根目录 [../AGENTS.md](../AGENTS.md) 为唯一规范源，本文件不重复；本机环境见 [../AGENTS.local.md](../AGENTS.local.md)。结构对齐 [../RuoYi-Vue3/AGENTS.md](../RuoYi-Vue3/AGENTS.md) 与 [../RuoYi-Vue-GO/AGENTS.md](../RuoYi-Vue-GO/AGENTS.md)。

## 项目定位

- **React 19 + TypeScript + Ant Design 5 + Redux Toolkit + Vite** 复刻 RuoYi-Vue3 前端功能（界面不强求一致，功能等价），服务三版后端（Java/Python/Go，8080 互斥）。
- `RuoYi-Vue3` 是**基准前端与行为参照**——所有行为疑问以其源码为准；与基准的有意差异集中登记在 [specs/deviations.md](specs/deviations.md)，接口契约零差异。
- 任务台账与模块三件套在 [specs/README.md](specs/README.md)，技术选型唯一依据 [specs/tech-stack.md](specs/tech-stack.md)，契约侦察基线在 [specs/reference/](specs/reference/)（对 RuoYi-Vue3 v3.9.2 的全量调查，实现前必读对应章节）。
- **已拍板决策（2026-09-27 用户）**：KeepAlive 页签缓存暂不实现（先看效果）；动态路由用 Gate 全量重渲染方案；tool/build 表单构建器不实现（AI 直接生成）。

## 常用命令

```bash
npm install
npm run dev          # 开发服务器，端口 8090（与 RuoYi-Vue3 的 80 并行），/dev-api 代理到 localhost:8080
npm run build:prod   # 生产构建（tsc -b && vite build）
npm test             # Vitest 纯逻辑单测
```

- Windows 包装脚本：`bin/run-web.bat`、`bin/buildProd.bat`（nvm use 22.14.0）。
- 验证方式：浏览器实操（admin/admin123）+ `npm test`；接口字段名用 Network 面板对照 Java 版返回。
- 环境变量三份（.env.development / .env.staging / .env.production），新增变量三处同步。

## 编码约定

1. **契约先行**：每个模块动工前读 [specs/reference/](specs/reference/) 对应章节 + RuoYi-Vue3 该模块源码，把行为要点写进该模块 spec.md 后再写码。
2. **三版通用性**：不判断后端语言、不读环境变量做后端分支；能力差异只消费 `GET /getPlatformInfo` 的 features 三字段（druidMonitor/serverMonitor/swaggerDocs）。
3. **存储键名是契约**：`Admin-Token` / `sessionObj` / `pwrChrtype` / `layout-setting` / `tags-view-visited` / `screen-lock` / `screen-lock-path` / `sidebarStatus` / `size`——不得改名。
4. **请求层单点**：所有请求走 `src/utils/request.ts`，文案/分支/防重逻辑禁止在页面层绕过或"顺手优化"。
5. **权限组件化**：按钮级权限一律 `<Auth>` / `useAuth()`（对位 v-hasPermi/hasRole），不发明新形态。
6. 有意差异必须先登记 [specs/deviations.md](specs/deviations.md) 再实现；发现基准可疑行为按原样复刻并注释"与 Vue3 版对齐"，不自作主张修。
7. svg 精灵资源与 RuoYi-Vue3 **同源**（src/assets/icons/svg 全量复制），symbolId `icon-[dir]-[name]`，不改名。
8. 勾选纪律同根 AGENTS.md：做完即勾、没做不勾注明原因、三件套全勾才在 specs/README.md 总表标 ✅ + 日期。

## 已知本机坑（继承工作区）

- Node 走 nvm 22.14.0（bin 脚本已固化）。
- 不主动 git commit/push（根 AGENTS.md 纪律 5）。
