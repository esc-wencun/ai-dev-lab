# Spec 4.0.0 布局主题

>
> **状态：✅ 完成（2026-09-28）**。布局壳 + Sidebar + Navbar（三导航模式）+ TagsView 页签栏 + Settings 抽屉 + HeaderNotice + 暗色主题 + 响应式 + footer + 锁屏视觉全部落地；核心链路（页签操作族/设置抽屉保存/navType 3 切换/暗色切换/持久化恢复/HeaderNotice 联动）浏览器实操验证。navType 2 单独实操、锁屏解锁回跳、重置配置按钮等次级路径代码就位未逐一实操（任务表已注明）。实现期间修复三处集成 bug：store 循环导入（dynamicRoutes 抽纯数据模块）、useMatches 需 data router（自制 useRouteMatches + matchRoutes）、react-router v7 路由语法（splat `*` 替代 `:path(.*)`；五条 dynamicRoutes 因 hidden 过滤被吞 → buildRouteObjects 加 keepHidden）。
> **背景**：布局壳是全部页面的容器；功能开关与数据流等价，视觉糖可简化（逐项登记 deviations）。
> **契约侦察**：[../reference/02-infra-contract.md](../reference/02-infra-contract.md) §7.2（$tab）/§9（布局与设置）；组件行为 reference/03 §1。
> **依赖**：02（store/路由）、03 批次 A。

## 范围

- 布局骨架（Sidebar + Navbar + TagsView + AppMain + Settings 抽屉 + footer）。
- 三种导航模式（1 纯左 / 2 混合 / 3 纯顶部）及联动。
- $tab 等价 hook（页签操作族）。
- 暗色主题 + 主题色；锁屏页 UI。

## 关键行为契约

1. **三种模式**（严格对齐 reference/02 §9.3）：navType 2 点一级菜单 activeRoutes 写 sidebarRouters 联动左侧，`hideList=['/index','/user/profile']` 不参与；navType 3 TopBar 按窗口宽/3/85 计算可视数溢出进更多菜单；Settings 切 navType 的 sidebar/hide/defaultRoutes 复位副作用。
2. **TagsView**：affix 不可关；右键六操作 + 下拉同款 + 全屏（Esc 退出）；中键关闭；card/chrome 两样式 + 图标开关 + 持久化联动；刷新走 `$tab.refreshPage` 等价（delCachedView + /redirect 中转——**KeepAlive 暂缓下刷新语义仍成立**：重挂载即刷新）。
3. **tabActions hook**（对位 $tab）：refreshPage/closeOpenPage/closePage（关当前跳 visited 末位，无则 `/`）/closeAll/Left/Right/OtherPage/openPage/updatePage——语义逐条 reference/02 §7.2。
4. **Navbar 右侧条**：HeaderSearch / HeaderNotice / Screenfull / 明暗切换 / SizeSelect / 布局设置 / 锁屏（lockScreen(fullPath)+push /lock）/ 头像下拉（个人中心/布局设置/退出登录——退出确认「确定注销并退出系统吗？」）。
5. **Settings 抽屉**：11 项配置；保存 → localStorage `layout-setting`（loading 提示），tagsViewPersist 关时删 `tags-view-visited`；重置 → 删两键 + reload；预置色板八色。
6. **响应式**：<992px 切 mobile 收侧栏 + 遮罩；侧栏 200px/收起 54px。
7. **iframe 页**：IframeToggle 按 tagsView.iframeViews 渲染（KeepAlive 暂缓下退化为普通切换，src 拼接 query 语义保留）；InnerLink 全高。
8. **暗色**：antd theme.darkAlgorithm + html.dark class + 自定义 CSS 变量（侧栏/页签/导航背景）；默认跟随系统。主题色色阶由 antd 派生（deviations #5）。

## 设计决策

1. 页签刷新在无 KeepAlive 阶段的实现：`/redirect` 中转路由 + 组件 key 变化强制重挂载——与基准用户可感行为一致（刷新后状态重置）。
2. 视觉动画（圆形扩散/粒子背景/搜索表单折叠动画）省略——deviations #4。
3. TopNav/TopBar 溢出计算的 85px/3 系数照抄基准（窗口宽自适应行为等价）。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
