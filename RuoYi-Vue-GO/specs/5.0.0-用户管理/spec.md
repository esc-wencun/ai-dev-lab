# 05 用户管理

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：`SysUserController`（13 端点）+ `SysUserServiceImpl` + Mapper XML
> 前端页面：`views/system/user`
> 依赖：1.0.0-基础设施、4.0.0-部门岗位（部门树/TreeSelect）

## API 清单（已对照 Java 源码核实，2026-09-24）

| # | 方法 | 路径 | 权限串 | @Log | 返回/行为要点 |
|---|------|------|--------|------|--------------|
| 1 | GET | /list | system:user:list | 无 | TableDataInfo；条件 userName/phonenumber like、status、deptId（IN 子查询含 find_in_set 子孙） |
| 2 | POST | /export | system:user:export | EXPORT | xlsx 流；列对位 @Excel 注解（序号/部门编号/登录名称/用户名称/邮箱/手机号/性别 readConverterExp/状态/最后登录IP/时间/部门名称） |
| 3 | POST | /importData | system:user:import | IMPORT | multipart file + updateSupport；初始密码 sys.user.initPassword；成功/失败文案逐字对位（含 <br/> 拼接） |
| 4 | POST | /importTemplate | system:user:import | 无 | 导入模板（表头+示例行） |
| 5 | GET | / 与 /{userId} | system:user:query | 无 | data=user + roleIds/postIds + roles（admin 过滤 role_id=1）+ posts |
| 6 | POST | / | system:user:add | INSERT | 校验链（账号/手机/邮箱唯一，文案逐字）；bcrypt；事务主表+角色+岗位 |
| 7 | PUT | / | system:user:edit | UPDATE | admin 保护；同 add 校验；事务更新+重置关联 |
| 8 | DELETE | /{userIds} | system:user:remove | DELETE | `当前用户不能删除`；admin 保护；事务清关联+软删 del_flag=2；踢下线 |
| 9 | PUT | /resetPwd | system:user:resetPwd | UPDATE | admin 保护；pwd_update_date 重置；踢下线 |
| 10 | PUT | /changeStatus | system:user:edit | UPDATE | admin 保护；停用踢下线 |
| 11 | GET | /authRole/{userId} | system:user:query | 无 | user + roles（含 flag 标记已授权） |
| 12 | PUT | /authRole | system:user:edit | GRANT | query 参数 userId/roleIds；清空重插 sys_user_role；踢下线 |
| 13 | GET | /deptTree | system:user:list | 无 | TreeSelect {id,label,children} |

## 实施记录

- 2026-09-24 完成。踢下线（停用/删除/重置密码/改角色后清 login_tokens 会话）按 Python 版结论实现（SCAN + 解析 userId）。
- 差异登记：数据权限 checkUserDataScope/checkRoleDataScope 未接入（1.0.0 DataScope 就绪后统一补；当前权限串已控制到按钮级，admin 放行行为等价）。
- 导入 Excel 列与 @Excel 注解对齐（IMPORT 字段：部门编号/登录名称/用户名称/邮箱/手机号码/性别）；导出追加 EXPORT 字段（账号状态/最后登录IP/时间/部门名称）。
