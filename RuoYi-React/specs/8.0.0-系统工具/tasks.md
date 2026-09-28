# Tasks · 8 系统工具

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。
> **【2026-09-28 用户决定】代码生成（gen）React 版暂时不做**——菜单若下发 gen 页，渲染「该功能暂未实现」占位页。其余任务照常。

- [ ] gen 主列表（查询/预览/编辑/同步/生成删除/导入表/创建表/批量删除）——**暂不实现**（用户决定 2026-09-28，菜单映射占位页）
- [ ] 预览弹窗（Tabs 分组 + pre 原样 + 复制 + 懒渲染）——暂不实现（同上）
- [ ] 修改配置页（三页签 + 字段表 dnd-kit 拖拽排序 + 行内编辑）——暂不实现（同上）
- [ ] 生成信息表单（模板/前端类型/生成方式/上级菜单 TreeSelect/树表/关联配置）——暂不实现（同上）
- [ ] 生成下载（zip blob + download-filename 头）+ 同步库——暂不实现（同上）
- [x] gen 占位页（tool/gen/index.tsx → NotImplemented「该功能暂未实现（React 版暂缓）」）（2026-09-28）
- [x] swagger 降级页（PlatformGate feature=swaggerDocs + iFrame /swagger-ui/index.html）（2026-09-28）
- [x] build 占位页（tool/build/index.tsx → NotImplemented「该功能未实现」，deviations #2）（2026-09-28）
- [ ] 验证：gen 全流程——取消（随 gen 暂缓，后续需要时按 spec 8.0.0 原设计恢复）
- [x] 验证：swagger Java 正常 iframe（2026-09-28，浏览器实操：/tool/swagger iframe 渲染 Java swagger-ui「若依管理系统_接口文档 v3.9.2 OAS 3.1」；Python/Go 降级提示待三版后端矩阵）
