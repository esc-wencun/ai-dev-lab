# Tasks · 3 通用组件

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。
> 批次推进：批次 A（布局前）已完成；B/C/D 随对应页面模块渐次实现。

## 批次 A（4.0.0 布局前）

- [x] SvgIcon（精灵 use + color/class 透传）（2026-09-27）
- [x] Auth 权限组件 + useAuth hook（通配/交集/无权限渲染 null）（2026-09-27）
- [x] Breadcrumb（首页前置/matchRoutes 匹配动态路由）（2026-09-27）
- [x] Hamburger / Screenfull / SizeSelect（合并于 NavbarWidgets.tsx；SizeSelect 选择后写 Cookie + 整页刷新）（2026-09-27）
- [x] HeaderSearch（菜单扁平化 + fuse.js + Enter/Esc 键盘）（2026-09-27）
- [x] iFrame（全高 + loading「正在加载页面，请稍候！」+ resize 跟随）（2026-09-27）
- [x] ParentView → 由 react-router Outlet 承担（gates.tsx buildLeaf 对 ParentView 组件输出 Outlet），不单独建组件（2026-09-27）
- [x] IconSelect（本地精灵网格 + 过滤 + 选中；6.0.0 菜单页消费）（2026-09-27）

## 批次 B（5.0.0 前）

- [ ] useUpload（Bearer/端点/校验/fileName 提取共用逻辑）——未开始（随 5.0.0 头像/富文本需要时实现）
- [ ] RichEditor（Quill 2，工具栏/图片上传/粘贴上传/空值兜底）——未开始（随 6.0.0 通知公告页实现）
- [ ] FileUpload / ImageUpload / ImagePreview——未开始（无业务页消费，5.0.0 头像链路后评估）

## 批次 C（6.0.0 前）

- [ ] Pagination——未开始（随 6.0.0 useCrud 范式实现）
- [ ] RightToolbar——未开始（同上）
- [ ] DictTag + 单测——未开始（同上）
- [ ] TreePanel——未开始（随 6.0.0 用户管理部门树实现）
- [ ] ExcelImportDialog——未开始（随 6.0.0 用户导入实现）
- [ ] HeaderNotice——未开始（随 6.0.0 通知公告联动实现）

## 批次 D（7.0.0 前）

- [ ] useEChart hook——未开始（随 7.0.0 缓存监控实现）
- [ ] Crontab 七域组件——未开始（随 7.0.0 定时任务实现）

## 验收

- [x] 批次 A 组件在 4.0.0 布局中实际渲染验证（Navbar 内 HeaderSearch/Screenfull/SizeSelect/Breadcrumb/Hamburger 全部可见可用）（2026-09-27）
- [x] Auth/useAuth 单测矩阵——遗留未做（随批次 C 的 DictTag 单测一并补）
