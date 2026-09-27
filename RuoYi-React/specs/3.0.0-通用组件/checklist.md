# Checklist · 3 通用组件

> 勾选 = 验收通过；没验证的不许勾。
> 按批次随依赖页面模块验收；全部批次完成才算本模块完成。

## 验收清单

- [ ] 批次 A：布局内嵌验收（Breadcrumb/HeaderSearch/全屏/尺寸在 4.0.0 场景验证）；Auth 用 admin 与受限账号各验证通配与拦截
- [ ] 批次 B：RichEditor 图片上传（选文件 + 粘贴截图两路径）真实后端通过；上传三件套 demo 验证回显/剥补前缀
- [ ] 批次 C：DictTag 真实字典（sys_normal_disable/sys_job_group）五类 tag 渲染正确；Pagination 分页/回顶正常；TreePanel 拖宽持久化；ExcelImportDialog 真实导入一条
- [ ] 批次 D：Crontab 生成的表达式在后端解析成功；useEChart 窗口缩放 resize 正常
- [ ] `npm test` 全绿（Auth 矩阵/elTagType 映射）

## 测试数据清理记录

- （ExcelImportDialog 验证导入的用户数据随 6.0.0 联动清理）
