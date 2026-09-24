# 04 部门管理 + 岗位管理

> **状态：✅ 已完成（2026-09-24；curl 端到端通过）**
>
> Java 版对应：`SysDeptController` / `SysPostController` + `SysDeptServiceImpl` + `SysPostServiceImpl` + Mapper XML
> 前端页面：`views/system/dept`、`views/system/post`
> 依赖：1.0.0-基础设施（权限/日志/分页由 01 统一供给）

## API 清单（已对照 Java 源码核实，2026-09-24）

### 部门 /system/dept

| # | 方法 | 路径 | 权限串 | @Log | 返回要点 |
|---|------|------|--------|------|---------|
| 1 | GET | /list | system:dept:list | 无 | data=扁平列表（前端树化），条件 deptName like/status |
| 2 | GET | /list/exclude/{deptId} | system:dept:list | 无 | 排除自身+子孙（ancestors find_in_set 语义） |
| 3 | GET | /{deptId} | system:dept:query | 无 | data=部门详情 |
| 4 | POST | / | system:dept:add | INSERT | 同父同名拒绝 `新增部门'x'失败，部门名称已存在`；父停用拒绝 `部门停用，不允许新增`；ancestors=父ancestors+parentId |
| 5 | PUT | / | system:dept:edit | UPDATE | 同名/上级是自己 `…上级部门不能是自己`/停用含正常子 `该部门包含未停用的子部门！`；ancestors 级联更新子孙；启用时启用全部祖先 |
| 6 | DELETE | /{deptId} | system:dept:remove | DELETE | 有子部门 601 `存在下级部门,不允许删除`；有用户 601 `部门存在用户,不允许删除`；软删 del_flag=2 |

（updateSort 端点为本 Java 版定制：PUT /updateSort，body {deptIds,orderNums} 逗号分隔——**未实现**，前端 dept 页面无调用，已登记 deviations）

### 岗位 /system/post

| # | 方法 | 路径 | 权限串 | @Log | 返回要点 |
|---|------|------|--------|------|---------|
| 1 | GET | /list | system:post:list | 无 | TableDataInfo 分页（rows/total 顶层） |
| 2 | POST | /export | system:post:export | EXPORT | xlsx 流（前端 proxy.download；**暂未实现**，随 5.0.0 用户导出一起补） |
| 3 | GET | /optionselect | 无（登录即可） | 无 | data=全部岗位（选择框） |
| 4 | GET | /{postId} | system:post:query | 无 | data=详情 |
| 5 | POST | / | system:post:add | INSERT | 岗位名称/编码唯一（文案 `新增岗位'x'失败，岗位名称已存在`） |
| 6 | PUT | / | system:post:edit | UPDATE | 同上（修改前缀） |
| 7 | DELETE | /{postIds} | system:post:remove | DELETE | 逗号分隔批量；已分配 `x已分配,不能删除`；物理 delete |

## Python 版同名 spec 踩坑复用

- 部门树构建、ancestors 级联更新（父部门变更时所有子孙的 ancestors 重算）——Go 版 service.Edit 已实现（旧前缀替换法）

## 实施记录

- 2026-09-24 完成。SQL 逐字对位 SysDeptMapper.xml/SysPostMapper.xml；校验文案逐字对位 Controller/ServiceImpl。
- 差异登记：① updateSort 未实现（本 Java 版定制，前端无调用）；② post/export 未实现（导出列定义与 5.0.0 用户导出共用 Excel 设施，届时一并补）；③ 部门 checkDeptDataScope（数据权限校验）在 1.0.0 DataScope 就绪后接入（当前 admin 放行，行为等价——普通用户访问本组端点由权限串控制到按钮级）。
- 端到端：部门增（ancestors 拼接验证）/同名拒绝/上级是自己拒绝/有子 601/有用户 601/软删 del_flag=2；岗位增/已分配拒绝/物理删。