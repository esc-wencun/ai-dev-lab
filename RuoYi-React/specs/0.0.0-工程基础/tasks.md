# Tasks · 0 工程基础

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 脚手架：Vite react-ts 模板经 scaffold 子目录上移合并；包名 `ruoyi-react`（2026-09-27）
- [x] `npm install` 基础依赖（react/react-dom/@types/vite/plugin-react/typescript）（2026-09-27）
- [x] 安装运行时依赖：react-router / @reduxjs/toolkit / react-redux / antd / @ant-design/icons / axios / js-cookie / jsencrypt / nprogress / file-saver / fuse.js / js-beautify / echarts / quill / react-cropper / @dnd-kit/core / @dnd-kit/sortable / dayjs / sass-embedded（2026-09-27）
- [x] 安装开发依赖：@types/js-cookie / @types/file-saver / @types/nprogress / vite-plugin-svg-icons（+ 其 peer fast-glob）/ vitest（2026-09-27）
- [ ] 如 React 19 + antd 出现兼容告警：引入 `@ant-design/v5-patch-for-react-19` 并在此注明版本——未出现告警，未安装（2026-09-27 实测 dev/build 无告警，如后续出现再补）
- [x] vite.config.ts：alias `@`→src；dev `port 8090 + host + open`；代理 `/dev-api`→8080（rewrite 剥离）+ `^/v3/api-docs` 透传；build 产物路径 `static/js|css/`（2026-09-27，build:prod 通过）
- [x] svg 精灵插件接入 + `src/assets/icons/svg/` 从 RuoYi-Vue3 全量复制 90 个（2026-09-27；浏览器可见性已验证）
- [x] tsconfig：strict + paths `@/*`（TS 6 弃用 baseUrl，paths 用 `./src/*` 相对形式）（2026-09-27，tsc -b 通过）
- [x] vite-env.d.ts 补 ImportMetaEnv 类型 + `virtual:svg-icons-register` 模块声明（2026-09-27）
- [x] 环境变量三份（development/production/staging），变量与值见 spec.md 关键实现要点 3（2026-09-27）
- [x] 目录骨架占位：api / assets / components / config / hooks / layout / plugins / router / store / utils / views（2026-09-27）
- [x] main.tsx：ConfigProvider(zhCN + theme token #409EFF) + AntdApp + `virtual:svg-icons-register`（RTK Provider/BrowserRouter 随 1.0.0/2.0.0 接入）（2026-09-27）
- [x] App.tsx 占位页 + svg 精灵测试 `<use href="#icon-user">`（2026-09-27）
- [x] Vitest 接入 + 冒烟测试（`npm test` 通过）（2026-09-27）
- [x] bin/run-web.bat / bin/buildProd.bat（nvm use 22.14.0，对齐 RuoYi-Vue3/bin 风格）（2026-09-27）
- [x] 工作区文档同步：根 AGENTS.md 工作区结构与常用命令补 RuoYi-React 条目；根 readme.md 摘要检查；RuoYi-Vue3/AGENTS.md 项目定位补基准声明；新建 RuoYi-React/AGENTS.md（2026-09-27）
- [x] RuoYi-React 自身 readme.md（启动方式/目录/与 Vue3 差异指针 → deviations.md）（2026-09-27）
