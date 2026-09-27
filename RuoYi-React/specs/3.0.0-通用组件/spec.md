# Spec 3.0.0 通用组件库

>
> **状态：🔶 批次 A 完成（2026-09-27）**。布局所需组件全部落地并在 4.0.0 布局中实际渲染验证（SvgIcon/Auth+useAuth/Breadcrumb/Hamburger/Screenfull/SizeSelect/HeaderSearch/iFrame/IconSelect；ParentView 由 react-router Outlet 承担）。批次 B/C/D 随对应页面模块渐次实现（任务清单已注明依赖关系）。
> **背景**：React 无指令体系，权限组件化；其余组件逐个对位基准 props/事件/行为。组件**按需渐次实现**——跟随页面模块开工前完成其依赖项，不必一次做完。
> **契约侦察**：[../reference/03-components-deps.md](../reference/03-components-deps.md) §1（每个组件 props/事件/行为基线）。
> **依赖**：01, 02。

## 范围与实现顺序

**批次 A（4.0.0 布局前）**：SvgIcon / Breadcrumb / Hamburger / Screenfull / SizeSelect / HeaderSearch / iFrame / ParentView(Outlet 占位) / Auth 权限组件 + useAuth。
**批次 B（5.0.0 前）**：RichEditor(Quill) / FileUpload / ImageUpload / ImagePreview / useUpload 复用逻辑。
**批次 C（6.0.0 前）**：Pagination / RightToolbar / DictTag / IconSelect / TreePanel / ExcelImportDialog / HeaderNotice。
**批次 D（7.0.0 前）**：Crontab（工作量最大单件）/ useEChart。

## 关键行为契约

1. **Auth 组件**（对位 v-hasPermi/v-hasRole，deviations #8）：`<Auth permissions={['system:user:add']}>` / `<Auth roles={['admin']}>`，无权限渲染 null（等价 DOM 移除）；`*:*:*` 与角色 admin 通配；`useAuth()` 提供 hasPermi(Or/And)/hasRole(Or/And)。
2. **Pagination**：total/page/limit/pageSizes[10,20,30,50]/pagerCount(<992→5)；改 size 页码越界回 1；autoScroll 平滑回顶。
3. **DictTag**：值按 separator 拆分；elTagType → antd Tag color 映射（default/success/info/warning/danger）；未匹配追加原值（showValue）。
4. **RichEditor**：Quill 2 直封，HTML 存储；图片上传/粘贴上传走 `/common/upload`（Bearer、file 字段、插入 `VITE_APP_BASE_API + fileName`）；jpeg/jpg/png/svg 校验、默认 5MB；空值兜底 `<p></p>`。
5. **Crontab**：七域（秒分时日月周年）规则组 + 实时拼串（默认 `* * * * * ?`）+ 下次执行说明；**表达式生成语义必须与基准一致**（有后端解析兜底，底线是不生成非法表达式）；界面可简化。
6. **TreePanel**：折叠/拖宽（rAF 节流）/宽度持久化/搜索过滤/展开收起刷新头部；expose 方法集。
7. **HeaderSearch**：defaultRoutes 扁平化 + fuse.js（threshold 0.2，title 0.7/path 0.3）+ 键盘导航 + 高亮（dangerouslySetInnerHTML）。
8. **ExcelImportDialog**：action 必传、点确定才上传、`?updateSupport=0|1`、成功 Modal 弹 msg、模板下载走 download()。

## 设计决策

1. 组件 props/事件命名 React 惯例优先（onChange/onXxx），文件头注释标注对位的 Vue3 组件与差异——方便逐页照抄页面逻辑时对照。
2. 上传三件套（FileUpload/ImageUpload/RichEditor 图片）共用 `useUpload`（Bearer 头/端点拼接/校验文案/fileName 提取），避免三处重复实现漂移。
3. FileUpload/ImageUpload/ImagePreview 基准版无业务页消费，但仍属全局基线——实现，批次 B 收尾。
4. DictTag 的 elTagType→antd color 映射表放常量文件，单测覆盖五类映射。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
