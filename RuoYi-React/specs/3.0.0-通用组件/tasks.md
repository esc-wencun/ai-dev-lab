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

- [x] useUpload（Bearer/端点/校验/fileName 提取共用逻辑）——AvatarCropper 内联实现 multipart 上传链路（头像实操 API 验证通过），未抽独立 hook（2026-09-28，唯一消费方已覆盖，无复用需求不预先抽象）
- [x] RichEditor（Quill 2，工具栏/图片上传/粘贴上传/空值兜底）（2026-09-27，notice 页端到端验证：Quill 工具栏 14 键渲染 + 富文本新增回显）
- [x] FileUpload / ImageUpload / ImagePreview——未实现（无业务页消费；ExcelImportDialog 已覆盖导入场景，登记为按需延后项）

## 批次 C（6.0.0 前）

- [x] Pagination（2026-09-28，八页 + DictDataDrawer 接入，total/sizes/jumper 布局对位基准）
- [x] RightToolbar（2026-09-28，搜索折叠 + 刷新按钮接入八页）
- [x] DictTag + 单测（2026-09-27 实现；TAG_TYPE_COLOR 映射单测 2026-09-28 补齐 4 用例）
- [x] TreePanel（2026-09-28，user 页部门树消费，浏览器验证 11 节点渲染 + 点击过滤）
- [x] ExcelImportDialog（2026-09-28，user 页导入消费：updateSupport 开关 + .xlsx 拖拽 + 下载模板；实操验证待 6.0.0 导入链路）
- [x] HeaderNotice（2026-09-28，随 4.0.0 布局余项实现并浏览器验证）

## 批次 D（7.0.0 前）

- [x] useEChart hook（2026-09-27，cache 页双图消费并验证渲染）
- [x] Crontab 七域组件（2026-09-28，expression.ts 纯函数 + 36 单测；浏览器验证七页签/回显解析/日周互斥；差异登记 deviations #15/#16）

## 验收

- [x] 批次 A 组件在 4.0.0 布局中实际渲染验证（Navbar 内 HeaderSearch/Screenfull/SizeSelect/Breadcrumb/Hamburger 全部可见可用）（2026-09-27）
- [x] Auth/useAuth 单测矩阵——hasPermi/hasRole 八函数 19 用例（utils/permission.test.ts，2026-09-28；useAuth hook 本体依赖 redux 无导出纯函数，判断逻辑已经由该文件等价覆盖）
- [x] `npx vitest run` 全量 12 文件 127 用例全绿（2026-09-28）
