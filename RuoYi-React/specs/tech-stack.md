# RuoYi-React 技术选型调研与结论

> 日期：2026-09-27　|　性质：选型决策记录（唯一选型依据，不要重新发明）
> 目标：用 React 复刻 RuoYi-Vue3 前端功能（界面不强求一致），服务三版后端；求职导向对齐 2026-09 JD 样本（`W:\sinosoft\简历\求职分析`——北京全栈 JD 明确点名 React 优先 + Redux/Zustand，TS/antd 高频）。

## 一、结论一览

| 职责 | 选型 | RuoYi-Vue3 对位 | 说明 |
|---|---|---|---|
| 语言 | TypeScript 5（strict） | JavaScript | JD 高频；antd 自带类型 |
| 框架 | React 19 | Vue 3.5 | 如遇 antd 兼容告警按官方引入 `@ant-design/v5-patch-for-react-19` |
| UI 组件库 | Ant Design 5.x | Element Plus 2.13 | 视觉不要求一致，交互功能等价 |
| 状态管理 | **Redux Toolkit + react-redux** | Pinia 3 | JD 点名 Redux/Zustand，用户拍板取 RTK；**不用 RTK Query**（保证请求层契约只有一份 axios 实现） |
| 路由 | react-router 7（library 模式） | vue-router 4 | 无 addRoute，动态路由用 Gate 全量重渲染方案 |
| HTTP | axios 1.x | axios 1.x | request.ts 逐条复刻契约 |
| 构建 | Vite | Vite 6 | react-ts 模板；dev 端口 8090 |
| 图表 | echarts 5 + 自研 useEChart hook | echarts 5 | 仅缓存监控两图 |
| 富文本 | Quill 2 直封自研 RichEditor | @vueup/vue-quill | 存储 HTML；图片上传/粘贴走 /common/upload |
| 头像裁剪 | react-cropper（cropperjs） | vue-cropper | 200×200 固定框；getCroppedCanvas().toBlob 等价 getCropBlob |
| 拖拽 | @dnd-kit/core + @dnd-kit/sortable | vuedraggable + sortablejs（隐式传递依赖） | 表单设计器已砍（deviations #2），仅 gen 字段排序、上传列表排序用 |
| 菜单图标 | vite-plugin-svg-icons（精灵与 Vue3 版同源复制）+ @ant-design/icons 回退 | 同左 | **DB meta.icon 存精灵名，资源必须同源复制**，否则菜单图标全丢 |
| 测试 | Vitest | 无（基准无测试框架） | 新增能力：纯逻辑必须带单测 |
| 样式 | Sass + antd ConfigProvider theme token | Sass + element css-vars | 暗色 = theme.darkAlgorithm + 自定义 CSS 变量 |
| 工具库 | js-cookie / jsencrypt / nprogress / file-saver / fuse.js / js-beautify / dayjs | 同左 | jsencrypt 沿用硬编码密钥对（仅"记住我"cookie 加解密，与后端无关） |

## 二、关键决策记录

1. **RTK 而非 RTK Query**：401/500/601 分支、防重复提交、POST form-urlencoded 下载等契约集中在 `utils/request.ts` 一份实现；RTK Query 会把请求散到各 store，破坏契约单点。
2. **RTK 而非 Zustand**：用户规则——JD 提及即取 RTK（北京 JD「熟悉 Redux/Zustand 或 Vuex/Pinia」两者皆提，用户默认 Redux Toolkit）。
3. **react-router library 模式**（`createBrowserRouter`/`useRoutes`），不引 framework 模式（无 SSR 需求，纯 SPA 对位）。
4. **动态路由 Gate 方案**（React 无 `router.addRoute`）：权限数据（getInfo + getRouters）就绪前渲染全屏 Loading，就绪后 `useRoutes` 全量重渲染；404 catch-all 放最后防误伤深层链接；外链（http）不注册 Route，渲染为侧边栏 `<a>` 跳转。
5. **KeepAlive 自研方案已设计、暂不实现**（用户 2026-09-27）：按路由 name 缓存组件实例 + 非激活隐藏；现阶段页签切换即重挂载（deviations #1）。
6. **dev 端口 8090**：与 RuoYi-Vue3（80）并行，逐页对照验收。
7. **主题色派生交给 antd token**：不手写十八档色阶（Vue3 版 utils/theme.js 的 hex 混白混黑），ConfigProvider colorPrimary 自动派生——界面不等价前提下允许的简化。
8. **不引入的状态管理/请求备选**：Zustand（JD 也点名，但用户默认 RTK）、TanStack Query（服务端状态层与现有 axios 拦截器体系重叠）、Next.js（JD 加分项，但本项目是纯 SPA 复刻，不适用）。
