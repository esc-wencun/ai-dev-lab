# Tasks · 4 布局主题

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。
> 本轮完成布局核心（Sidebar/Navbar/首页内容区），TagsView/Settings/navType 2·3 随下轮任务。

- [x] layout/index.tsx 骨架（Sidebar + Navbar + 内容区 + collapsed 联动）（2026-09-27）
- [x] Sidebar（Logo + 递归菜单 + 唯一可见子级 + 外链 `<a>` + 图标精灵 + collapsed 54px）（2026-09-27，浏览器验证菜单渲染）
- [x] Navbar navType 1（Hamburger + Breadcrumb + 右侧组件条全量）（2026-09-27）
- [x] 退出登录（Modal 确认「确定注销并退出系统吗？」→ logOut → /login）（2026-09-27）
- [x] 锁屏入口（lockScreen(当前路径) → /lock）（2026-09-27，入口按钮就位）
- [x] TagsView 页签栏（affix/右键六操作/全屏/中键/card+chrome/持久化联动）（2026-09-28，浏览器验证：三页签累积/右键菜单条件渲染/关闭左侧/持久化恢复）
- [x] TopNav（navType 2 一级联动 + hideList）（2026-09-28，代码就位；navType 3 路径浏览器验证，navType 2 联动逻辑同构复用）
- [x] TopBar（navType 3 溢出更多菜单）（2026-09-28，浏览器验证：顶部菜单 + 更多菜单溢出 + sidebar hide + 汉堡隐藏 + 复位联动）
- [x] Settings 抽屉（11 项 + 保存/重置 + navType 切换副作用）（2026-09-28，浏览器验证：11 项齐全/保存写 layout-setting/navType 副作用/复位 defaultRoutes）
- [x] 暗色主题（darkAlgorithm + html.dark + CSS 变量完整覆盖）（2026-09-28，浏览器验证：明暗切换整页生效 + 刷新恢复 + vueuse-color-scheme 持久化）
- [x] 锁屏页视觉完整版 + 响应式 <992 + footer（2026-09-28，响应式 mobile 切换浏览器验证；footer/锁屏视觉代码就位随 walkthrough 检查）
- [x] HeaderNotice（铃铛 + 未读 + 全部已读 + 详情联动）（2026-09-28，浏览器验证：Badge 3/Popover 三条公告/全部已读 Badge 归零/sys_notice_read 落库）
- [x] useTabActions（$tab 九方法对位）+ useRouteMatches（非 data-router 的 useMatches 等价物，修复 "must be used within a data router" 崩溃）（2026-09-28）
- [x] 验证：admin 登录 → 菜单渲染 → 点击跳转 /system/user 与 /system/config → config 页真实数据渲染（2026-09-27）
