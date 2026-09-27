# Spec 8.0.0 系统工具

>
> **状态：🔶 完成（2026-09-28，gen 暂缓）**。swagger 降级页 + gen/build 占位页落地，tsc 全绿。gen 前端暂缓（用户 2026-09-28 决定，deviations #2 更新）：菜单 gen → NotImplemented 占位；后续需要时按本 spec「关键行为契约」原设计恢复。
> **背景**：gen 全套 + swagger 降级页。**tool/build 不实现**（2026-09-27 用户拍板：AI 可直接生成表单代码，设计器无复刻价值）——deviations #2；菜单若下发 build 页渲染「该功能未实现」占位。
> **契约侦察**：[../reference/01-pages-and-api-baseline.md](../reference/01-pages-and-api-baseline.md) §1.4/§2。
> **依赖**：03、06（useCrud）。

## 范围

- gen：主列表 / 导入表弹窗 / 创建表弹窗 / 预览弹窗 / 修改生成配置（子路由 /tool/gen-edit/index/:tableId）/ 生成下载 / 同步库。
- swagger：PlatformGate 降级页（features.swaggerDocs）。
- build：占位页。

## 关键行为契约

1. **gen 主列表**：查询 + 预览/编辑/同步（synchDb）/生成删除 + 导入表（listDbTable 查询多选 importTable）+ 创建表（createTable）+ 批量删除。
2. **预览弹窗**：Tabs 按生成文件分组 + `<pre>` 原样展示（**无语法高亮**，对齐基准）+ 复制（v-copyText 语义：clipboard API + fallback）；Tabs 懒渲染防长内容卡顿。
3. **修改配置页**：Tabs 三页签（基本信息/字段信息/生成信息）：
   - 字段表：**行拖拽排序**（@dnd-kit，onEnd 重排 sort 序号——对齐基准 sortablejs onEnd 语义）+ 行内编辑（javaType/javaField/htmlType/queryType 等下拉）。
   - 生成信息：模板类型（单表/树表/主子表）/前端类型/生成方式（zip/自定义路径）/上级菜单 TreeSelect（menu treeselect）/树表配置/子表关联。
4. **生成下载**：`$download.zip` 等价（GET `/tool/gen/download/{tableName}` blob + `download-filename` 头）；genCode 先行。
5. **生成内容是后端 codegen 产物**（Java/Vue 代码），前端原样展示下载，不属前端等价改造范围。
6. **swagger**：PlatformGate（swaggerDocs，副标题「系统接口页基于 springdoc swagger-ui…」）→ 正常时 iFrame `/swagger-ui/index.html`。

## 设计决策

1. build 占位页：antd Result icon=info「该功能未实现」+ 原因说明（AI 生成替代）——与 features 降级页视觉同构。
2. 字段表拖拽排序 hook（useDragSort）独立成 hook，7.0.0/其他场景可复用。

## Task 分解 → [tasks.md](tasks.md) ｜ 验收 → [checklist.md](checklist.md)
