# AGENTS.md（RuoYi-Vue3 前端）

本文件是前端子目录的 AI 编码规范入口与**前端规范全集**。工作区级规范（三版兼容契约、共用库纪律、spec 勾选纪律、编码行为纪律、已踩过的坑）以根目录 [../AGENTS.md](../AGENTS.md) 为**唯一规范源**，本文件不重复；本机环境（Node 版本、端口、数据库/Redis 指向）见 [../AGENTS.local.md](../AGENTS.local.md)。结构对齐 [RuoYi-Vue-GO/AGENTS.md](../RuoYi-Vue-GO/AGENTS.md)。

## 项目定位

- Vue 3 + Element Plus + Vite 前端，**三个后端（Java / Python / Go）共用同一份前端代码**。
- 2026-09-25 起纳入可修改范围（按学习需要演进），前提是**不破坏三版后端通用性**：后端能力差异用接口下发的能力开关表达，不用语言/环境硬编码（见〈功能开关〉）。
- 基线为上游 RuoYi-Vue3 v3.9.2；相对上游的本地改动目前只有「平台标识降级提示」一组（`src/api/platform.js` + 三个监控/工具页）。
- **RuoYi-React**（`../RuoYi-React/`）是本前端的 React + antd 学习性等价复刻工程（2026-09-27 起）；本前端仍是**基准前端与行为参照**，RuoYi-React 的行为契约以其源码为准（差异登记在 `RuoYi-React/specs/deviations.md`）。

## 常用命令

```bash
npm install
npm run dev           # 开发服务器，端口 80，自动开浏览器；/dev-api 代理到 http://localhost:8080
npm run build:prod    # 生产构建，产物 dist/
npm run build:stage   # 预发布构建（读 .env.staging）
npm run preview       # 预览构建产物
```

- Windows 包装脚本：`bin/run-web.bat`（nvm use 22.14.0 后跑 dev）、`bin/buildProd.bat`、`bin/package.bat`。
- **本项目没有测试框架、没有 ESLint / Prettier**（`package.json` 只有 dev/build/preview，根目录无任何 lint 配置）——不要臆造 `npm test` / `npm run lint` / `npm run format`。
- 验证方式是浏览器实操：先起一个后端（8080，三版互斥）→ `npm run dev` → admin / admin123 登录走页面；接口字段名/结构用浏览器 Network 面板对照 Java 版返回确认（根 AGENTS.md「验收以前端为准」）。
- 环境变量三份（`.env.development` / `.env.staging` / `.env.production`），新增变量三处都要同步。

## 架构要点

### 启动与路由链

`index.html` → `src/main.js`（注册 Element Plus、store、router、全局插件与全局组件）→ `src/permission.js`（全局路由守卫：无 token 跳登录；有 token 但 user store 里 roles 为空时先 `getInfo()` 再 `generateRoutes()` 挂动态路由）。

路由有**两个来源**，改菜单类需求时最容易踩错：

1. 静态：`src/router/index.js` 的 `constantRoutes`（login / register / 404 / 401 / redirect 等）与 `dynamicRoutes`（按 `permissions` / `roles` 过滤的页面，权限判断走 `filterDynamicRoutes`）。
2. 动态（业务页面主来源）：后端 `GET /getRouters` 返回菜单树 → `src/store/modules/permission.js` 的 `filterAsyncRouter()` 把菜单里的 `component` 字符串映射到 `src/views/**/*.vue`（`import.meta.glob` + `loadView`）。

所以**新增一个页面要同时满足两处**：`src/views/` 下有对应 `.vue`，且菜单表里 `component` 填相对 `views/` 的无扩展名路径（如 `system/user/index`）。只在 `router/index.js` 里加路由不会出现在侧边栏；只建 `.vue` 不配菜单则 `loadView` 找不到组件。`component` 为 `Layout` / `ParentView` / `InnerLink` 时是布局占位，不是文件路径。

### 权限

`GET /getInfo` 返回的 `roles` / `permissions` 存在 `src/store/modules/user.js`（`roles` 为空数组时会被置为 `['ROLE_DEFAULT']`）。判断入口三处：模板指令 `v-hasPermi` / `v-hasRole`（`src/directive/permission/`）、脚本里 `$auth.hasPermi*()` / `hasRole*()`（`src/plugins/auth.js`）、动态路由过滤 `filterDynamicRoutes()`。权限 `*:*:*` 与角色 `admin` 视作通配。

### 请求层（`src/utils/request.js`，全站唯一 axios 实例）

- token 存在 Cookie `Admin-Token`（`src/utils/auth.js`），请求头固定 `Authorization: Bearer <token>`。
- **响应按 body `code` 分支，不按 HTTP 状态码**：401 弹重新登录确认框、500 报错、601 警告、其余非 200 走 Notification——后端必须保持「HTTP 200 + body code」语义（根 AGENTS.md 兼容契约）。
- 防重复提交在拦截器内做：POST/PUT 比对 sessionStorage 中的 URL + body + 时间（默认 1000ms 内视为重复），可传 `headers.repeatSubmit = false` 关闭或 `headers.interval` 调整；body ≥ 5M 时自动跳过该检查。
- 响应 `responseType` 为 blob / arraybuffer 时直接返回数据；通用下载用 `download()`（POST + 表单编码 + blob，失败时解析响应体里的错误码提示）。

### 全局挂载（写新页面前先看清能少写多少）

- 自动导入（`vite/plugins/auto-import.js`，`dts: false`）：`vue` / `vue-router` / `pinia` 的全部 API，外加 `useDict`、`selectDictLabel` ——**这些符号一律不要手写 import**（现有代码里 `ref`、`defineStore`、`computed` 就是不 import 的）。
- `main.js` 全局方法：`useDict`、`parseTime`、`resetForm`、`handleTree`、`addDateRange`、`getConfigKey`、`selectDictLabel(s)`、`download`；全局组件：`Pagination`、`RightToolbar`、`DictTag`、`Editor`、`FileUpload`、`ImageUpload`、`ImagePreview`。
- `src/plugins/index.js` 挂载：`$tab`（页签操作）、`$auth`、`$cache`（session/local 封装）、`$modal`（确认框快捷方法）、`$download`。
- 图标两套：`src/assets/icons/svg/`（约 90 个）经 `vite-plugin-svg-icons` 注册为 `icon-[dir]-[name]`，模板里用 `<svg-icon icon-class="...">`；Element Plus 图标是另一套，由 `src/components/SvgIcon/svgicon` 提供。
- 组件要被 tagsView 缓存，必须写 `<script setup name="X">`（由 `unplugin-vue-setup-extend-plus` 支持，`keep-alive` 按组件名匹配 `tagsView.cachedViews`）。

### 状态与布局

Pinia，`src/store/modules/`：`user`（登录态/权限）、`permission`（路由）、`dict`（字典缓存）、`app`、`settings`、`tagsView`、`lock`（锁屏）。布局与主题在 `src/layout/`，默认布局项在 `src/settings.js`（标签页、导航模式、深色主题等；深色需引入 Element Plus dark css-vars，`main.js` 已引）。

### 接口层

按业务域分模块放 `src/api/`（`system/`、`monitor/`、`tool/` 三个子目录，加 `login.js` / `menu.js` / `platform.js`），页面 import 后调用；命名沿用 `listXxx` / `getXxx` / `addXxx` / `updateXxx` / `delXxx`。新增域按此结构落位，不要在页面里裸写 axios。

## 三版通用性：功能开关（新增后端特有页面时必须遵守）

三版后端能力不同（服务监控、Druid 缓存监控、Swagger 文档等仅 Java 版提供），前端靠 `GET /getPlatformInfo` 返回的 `features` 字段降级，**不判断后端语言、不读环境变量做分支**：

- 三个现成范例：`src/views/monitor/server/index.vue`、`src/views/monitor/druid/index.vue`、`src/views/tool/swagger/index.vue`——`features` 初值取全 `true`（保持对 Java 版的原行为，也避免接口慢时误显降级页），拿到响应后按开关渲染 `el-result`「该功能仅 Java 版提供」，并提供「重新检测」按钮。
- 开关字段名由三版后端共同约定（契约主文档为准），前端只消费，不自行发明字段名。

## 前端改动纪律（根 AGENTS.md 编码行为纪律的补充）

1. 改接口调用前先确认三版后端都实现了该端点（对不上时先补后端或登记差异），不要在页面里对某一版后端做特判。
2. 判断某功能"前端有没有做"要扫实际调用（`src/api/*.js` + 页面内 `proxy.download` / 全局方法调用），不凭印象断言全覆盖（根 AGENTS.md「功能核对以数据为准」的前端侧做法）。
3. 一次任务只改与任务相关的行，不顺手改格式/组件结构——同一份前端服务三版后端，任何"顺手改进"都会扩大三版共同的回归面。
