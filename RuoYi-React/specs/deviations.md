# Deviations · 与 RuoYi-Vue3 基准的有意差异清单

> 目的：「功能等价、界面不强求一致」的前提是**所有差异都有意且可查**。本项目所有已知差异集中在此，总验收时逐条核对。新发现差异随时登记，禁止只留在某次对话里。
> 格式：差异点 / RuoYi-Vue3 行为 / 本项目行为 / 原因。
> 登记规则：影响**接口契约**的差异（端点/参数/返回字段/状态码语义/存储键名）一律不允许——那说明实现错了，不是差异；影响**用户可感知行为**的差异必须在此登记并注明用户是否已拍板。

| # | 差异点 | RuoYi-Vue3 行为 | 本项目行为 | 原因 |
|---|--------|----------------|-----------|------|
| 1 | 页签缓存（keep-alive） | keep-alive 按 cachedViews（组件 name）缓存，切页签后查询条件/分页状态保留 | **暂不实现**（用户拍板 2026-09-27）：切页签组件重挂载、页面内状态丢失；tagsView.cachedViews 的 noCache 规则照常维护，为将来补自研留好数据面 | React 无 keep-alive，自研成本高风险大；用户要求先看无缓存版实际效果再决策（specs/README 遗留待办 #1） |
| 2 | tool/build 表单构建器 + tool/gen 代码生成 | build：完整拖拽设计器；gen：全套代码生成前端页 | **build 不实现**（用户拍板 2026-09-27）；**gen 暂缓**（用户拍板 2026-09-28）：两菜单渲染「该功能未实现/暂未实现」占位页 | build：AI 可直接生成表单代码，设计器无复刻价值；gen：React 版暂缓，后续需要时按 spec 8.0.0 原设计恢复 |
| 3 | UI 组件库与视觉 | Element Plus，若依经典视觉 | Ant Design 5，布局结构等价（侧边栏/导航/页签/设置抽屉/暗色）但像素级样式不同 | 用户明确「界面不需要完全一致，功能保持等价」 |
| 4 | 视觉动画 | 主题切换圆形扩散动画、锁屏粒子背景、搜索表单折叠动画 | 简化或省略（功能开关与数据流等价） | 纯视觉糖，不属于功能等价范围 |
| 5 | 主题色派生 | utils/theme.js 手写 light-1..9/dark-1..9 十八档色阶 | antd ConfigProvider colorPrimary 自动派生 | 组件库机制不同；视觉色阶精度差异可接受 |
| 6 | dev 端口 | 80 | **8090** | 与 RuoYi-Vue3 并行运行，同屏对照验收 |
| 7 | 语言 | JavaScript（无 lint/无测试） | TypeScript strict + Vitest + oxlint（脚手架自带） | 用户拍板 TS；新增测试能力属增强不属差异 |
| 8 | 权限控制形态 | v-hasPermi / v-hasRole 指令（DOM 移除） | `<Auth>` 包裹组件（无权限渲染 null，语义等价移除）+ useAuth() hook | React 无指令体系 |
| 9 | API 自动导入 | unplugin-auto-import（vue/pinia API 免 import） | 显式 import（React 惯例） | React 生态无等价自动导入惯例 |
| 10 | 首页内容 | 若依介绍 + 技术选型列表（Vue 栈）+ 捐赠二维码卡 | 本工程介绍 + React/TS/antd 真实技术栈 + 去掉捐赠卡 | 文案归属本工程；捐赠卡不属功能等价范围 |
| 11 | sortablejs 依赖 | 直接 import 但非直接依赖（靠 vuedraggable 传递依赖存活） | @dnd-kit 显式依赖 | 修复基准版的依赖隐患 |
| 12 | uploadAvatar Content-Type 声明 | api 函数声明 urlencoded（实际 FormData 按 multipart 发送的历史瑕疵） | 直接按 multipart 声明 | 行为一致（后端收到的就是 multipart），修正声明 |
| 13 | echarts macarons 主题 | `echarts.init(el, "macarons")` 但主题未注册，实际回退默认主题 | 直接用默认主题 | 行为等价（视觉同基准实际效果） |
| 14 | 富文本 XSS 面 | v-html 渲染公告 HTML（无 sanitize） | dangerouslySetInnerHTML 同等渲染（不额外引入 sanitize） | 保持等价；如统一加固需与基准前端同步决策 |
