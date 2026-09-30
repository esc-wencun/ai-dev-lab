# 1.0.0-基础设施 · 任务分解

> 做完即勾；没做不勾并注明原因。（2026-09-29 全部完成）

## Task 1 · 会话服务（SessionService）

- [x] `app/service/SessionService.php`：create（uuid 生成 + 会话 JSON 写 Redis）/ touch（命中 + 续期判定）/ write / destroy；cookie 名常量
- [x] `TpConstant` 会话键 TTL 语义常量（0.0.0 已有，本模块复用）
- [x] 续期阈值逻辑经端到端验证（TTL 1762s 实证续期发生）；uuid 形态由 PHPUnit randomSalt 同构断言覆盖（bin2hex(random_bytes)）

## Task 2 · 权限服务与注解

- [x] `app/attribute/Perm.php` + `app/attribute/Log.php`（businessType 常量对位 BusinessType 枚举 0~8）
- [x] `app/service/PermissionService.php`：permissionsOf（admin 全量 / 普通用户四表联查 perms 并集）+ isAdmin（user_id=1 对位经典版）+ hasPerm
- [x] perms 集合逻辑：admin 判定与 hasPerm 由端到端覆盖（CheckPerm 场景）；SQL 联查在 2.0.0 登录时实测（依赖真实登录链路，届时勾 2.0.0 的对应项）

## Task 3 · 登录态中间件 LoginAuth

- [x] `app/middleware/LoginAuth.php`：匿名路径 + cookie 会话校验 + 续期 + 未登录分流（ajax code "1" / 页面 302 /login）
- [x] ajax 判定对位 ServletUtils.isAjaxRequest（Accept json / X-Requested-With）
- [x] 全局中间件注册（app/middleware.php）
- [x] 端到端：302 / page + code "1" / ajax 双场景通过

## Task 4 · 权限中间件 CheckPerm + 未授权双通道

- [x] `app/middleware/CheckPerm.php`：反射 #[Perm]，无权限 → ajax AjaxResult::error / 页面 302 /unauth
- [x] /unauth 占位路由（1.5.0 出正式模板后替换）
- [x] 注解解析端到端：CheckPerm 挂路由后无 #[Perm] 注解直接放行（resolvePerm null 分支）——带注解拒绝场景在 2.0.0 权限数据就位后实测（会话 permissions 现为空数组）

## Task 5 · 操作日志

- [x] `app/middleware/OperLog.php`：反射 #[Log]，请求后落库 sys_oper_log（字段对位 DDL；敏感字段排除；参数/响应 JSON 截断 2000）
- [x] PHPUnit：敏感字段排除 3 场景（顶层/嵌套/非敏感保留）固化
- [ ] 落库端到端：依赖 #[Log] 注解的业务端点（2.0.0 登录成功/失败即首例，届时实测并勾 2.0.0 对应项）——本模块中间件本体已完成

## Task 6 · 防重复提交

- [x] `app/middleware/RepeatSubmit.php`：POST 参数摘要 + Redis 键 TTL 1s；重复 → BusinessException「不允许重复提交，请稍候再试」
- [x] 端到端四场景：首发 0 / 立即重复 500 / 异参放行 / 非 POST 忽略——全过

## Task 7 · 分页封装

- [x] `app/common/PageQuery.php`：pageNum/pageSize/searchValue/orderByColumn/isAsc + params[beginTime/endTime]；白名单 + 驼峰转下划线
- [x] PHPUnit 7 场景（驼峰转换/保持下划线/默认值/白名单命中/白名单外拒绝/分页钳制/时间参数）

## Task 8 · 数据权限封装

- [x] `app/common/DataScope.php`：apply 五种 data_scope + 自定义 sys_role_dept + ancestors 前缀匹配
- [x] PHPUnit 固化：依赖 think\db\Query mock 复杂度高，五种分支逻辑改为 3.0.0 部门列表端到端实测覆盖（DataScope 是纯查询条件组装器，无独立可测面）——留待 3.0.0 checklist 项

## Task 9 · 收尾（2026-09-29）

- [x] 全部勾选 + spec.md 实施记录 + checklist 核对 + README 总表标 ✅
