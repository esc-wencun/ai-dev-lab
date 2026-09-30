# 1.5.0-静态资源与布局 · spec

> **状态：✅ 已完成（2026-09-29，与 2.0.0 联调收口）**
> 经典版静态资源移植 + 主框架布局模板。是「UI 移植」决策（tech-stack.md 第五节）的落地模块。
> 依赖：0.0.0。与 1.0.0 并行推进，2.0.0 前完成。

## 范围

1. **静态资源移植**：从 `reference/RuoYi-classic/ruoyi-admin/src/main/resources/static/`（8.3MB / 90 JS）复制到 `public/`——**static/ 的内容直接映射站点根**（结果形态 `public/js/`、`public/ajax/`、`public/css/`、`public/ruoyi/`、`public/img/` 等）：经典版模板以根路径引用资源（`@{/js/jquery.min.js}`、`@{/ajax/libs/...}`、`@{/ruoyi/js/ry-ui.js?v=4.8.3}`，include.html 实锤），TP 的 public 目录即 Web 根，**不能套一层 static/**，否则全部资源 404。保留版权头，不做美化；favicon.ico、file/（上传演示目录）一并带上。
2. **布局模板**：`app/view/layout/default.html`（`{block}` 布局）+ 头部/侧栏/标签页片段；对位经典版 `templates/index.html` + `include.html`（`include :: header('xx')` / `include :: footer` 片段机制 → TP `{include}`）。
3. **菜单渲染**：对位经典版 index.html 的四级菜单嵌套（`th:href="@{${cmenu.url}}"`，href 取 sys_menu.url 原值，target/is_refresh 原样输出）——**改为 service 递归组装 + 模板 volist 渲染**（经典版 getChildPerms 已是递归树，语义相同）；admin 全量菜单 `menu_type in ('M','C') and visible='0'`。
4. **样板页**：选「部门管理」（经典版 dept 页最简，无 bootstrap-table 分页，是树表）或「岗位管理」（标准 bootstrap-table 列表）做第一个业务页范本，在 2.0.0 后 3.0.0 落地——本模块先出「页面骨架约定」文档小节（工具栏 + 表格容器 + 弹窗 div + 页面专属 JS 的目录约定）。
5. **demo 菜单处理定案**：官方 SQL 带 demo/若依官网等菜单（deviations #4），本模块决定：导入 SQL 保留全部菜单，TP 端对 demo/tool/druid 路由出「建设中」占位页（比 404 保真——经典版这些页也是打开就能看）。结论回填实施记录。

## 关键设计说明

1. **经典版 JS 对后端的约定**（1.5.0 内出文档小节，作为 3.0.0+ 各模块 spec 的依据）：
   - bootstrap-table 默认 POST + server 分页；`queryParams` 拼 pageNum/pageSize + 表单字段；
   - `common.js` ajaxSetup：超时/错误统一提示；未登录跳转由后端 302 驱动（ajax 场景 code "1" 由 ry-ui.js 处理——实施时核对具体处理点并记录）；
   - `ry-ui.js` 的 `$.operate`（add/edit/remove/view）、`$.table`（search/init/exportExcel/importExcel）、`$.modal`（弹窗）函数签名与 options.url 系列约定（createUrl/updateUrl/removeUrl/viewUrl/exportUrl）。
2. **不做的事**：不重画 UI、不换 UI 库、不引入前端构建工具、不启用 i18n（deviations #15）。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录

（未动工）
