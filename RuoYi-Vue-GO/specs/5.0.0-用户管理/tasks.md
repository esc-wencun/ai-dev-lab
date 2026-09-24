# Tasks · 05 用户管理

> 勾选纪律：做完即勾；没做的不许勾，行尾注明原因。

- [x] 执行 specs/README.md《动工检查单》，写实本文件夹三件套
  - 注：对照 SysUserController 13 端点 + ServiceImpl + Mapper XML 源码（2026-09-24）
- [x] 模块特有注意点：导入导出列对齐 @Excel 注解；停用/改密/删除后清 Redis 会话（踢下线）
  - 注：kickUsers=SCAN login_tokens + 解析 userId 删除；changeStatus(停用)/resetPwd/remove/insertAuthRole 四处触发
- [x] Task 列表/部门树（list 分页 + deptTree TreeSelect）
- [x] Task 详情/新增/修改/删除（getInfo 多数据组装、事务三表、admin 保护、当前用户保护）
- [x] Task resetPwd/changeStatus/authRole（授权页+提交）
- [x] Task 导入导出（export 列对位 @Excel、importTemplate、importData 校验链文案逐字）
- [x] 端到端：curl 全链（列表/部门树/新增含角色岗位/详情 roleIds postIds/导出 xlsx 6693b/软删 del_flag=2+关联清理）
  - 注：2026-09-24 通过；测试用户 testuser05 已物理清除、Redis 会话清空
