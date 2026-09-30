# 1.0.0-基础设施 · spec

> **状态：✅ 已完成（2026-09-29）**
> 对位经典若依 ruoyi-framework（shiro 过滤器链/权限/日志/防重）+ ruoyi-common 通用封装。为 2.0.0 起所有业务模块供给横切能力。
> 依赖：0.0.0 已完成（2026-09-29）。

## 范围

| 能力 | 对位经典版（源码已核） | 实现载体 |
|---|---|---|
| 登录态中间件 | Shiro 主链 `user`：未认证跳 loginUrl；ajax 未登录返回 `{"code":"1","msg":"未登录或登录超时。请重新登录"}`（LoginController 特供同款文案） | `app/middleware/LoginAuth.php`：cookie uuid → Redis 会话命中 → 刷新空闲超时（30 分钟滑动，剩余 <20min 续期）；未登录分流：ajax（Accept 含 application/json 或 X-Requested-With: XMLHttpRequest，对位 ServletUtils.isAjaxRequest）出 code "1" JSON，页面 302 /login |
| 权限校验 | `@RequiresPermissions`；授权失败 GlobalExceptionHandler：ajax → `AjaxResult.error(msg)`（code 500），页面 → 渲染 error/unauth | 注解 `#[Perm('system:user:list')]` + 中间件 `CheckPerm`；同双通道语义 |
| 按钮级权限（模板） | `shiro:hasPermission` 标签 + 页面 JS 变量 `@permission.hasPermi(...)` | 模板函数 `check_perm($perm)`：读会话 permissions 集合 |
| 数据权限 | @DataScope AOP | service 层 `DataScope::apply($query, $user)`（3.0.0+ 按需声明，本模块先落封装） |
| 操作日志 | `@Log(title, businessType)` + LogAspect（异步）；sys_oper_log 17 列（DDL 已核）；BusinessType 枚举 0其它/1新增/2修改/3删除/4授权/5导出/6导入/7强退/8清空；敏感字段排除 password/oldPassword/newPassword/confirmPassword | `#[Log(title, businessType)]` 注解 + 中间件 `OperLog`，同步落库（deviations #13）；参数 JSON 截断 2000；失败记 error_msg/status=1 |
| 防重复提交 | SameUrlDataInterceptor：session 按 URL 存 {参数JSON, 时间}，参数相同且间隔 < interval（默认 1000ms）判重复，抛异常「不允许重复提交，请稍候再试」 | 中间件 `RepeatSubmit`：Redis `repeat_submit:<uuid>:<md5(url+参数)>` + TTL interval 秒——语义等价（参数摘要相同 + 间隔内即拦） |
| 登录日志 | AsyncFactory.recordLogininfor → sys_logininfor（info/login_name/ipaddr/browser/os/status/msg/login_time） | 2.0.0 登录链路封装 `LoginLogService` 直接落库 |
| 分页 | PageHelper startPage + BaseController.getDataTable（code=0） | 1.0.0 落 `PageQuery` 解析（pageNum/pageSize/orderByColumn/isAsc 白名单）+ `TableDataInfo::of`（0.0.0 已有）；排序白名单防注入 |

## API（端点级清单）

本模块无独立页面与端点，产出物为中间件/注解/模板函数/服务类。权限字符串以 ry-tp 库 sys_menu.perms 为准（经典版：C 型菜单 `system:user:view`、F 型按钮 `system:user:list`——与 RuoYi-Vue 相反）。

## 关键设计说明

1. **会话续期语义**：30 分钟空闲超时，每次请求 touch（对位 Shiro）；TP 版 Redis TTL 天然实现，剩余 <20 分钟续满 30 分钟（RuoYi-Vue 语义，行为无害对齐）。
2. **`#[Perm]` / `#[Log]` Attribute**：PHP 8 注解声明在控制器方法，中间件反射读取（对位注解 AOP）。
3. **权限集合**：登录时从 sys_menu 算好 permissions 存会话；本模块提供 `PermissionService::hasPerm($session, $perm)`（admin 全通过）。
4. **中间件注册**：全局中间件 LoginAuth（对位主链 user），路由级 CheckPerm/OperLog/RepeatSubmit 由各控制器按需挂（对位注解逐方法生效）。
5. **数据权限**：经典版 @DataScope 别名占位 `${params.dataScope}` 拼 where；TP 版 query->where 注入等价条件，封装 `DataScope::apply()`。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录

- 2026-09-29 全部 9 个 Task 完成（PHPUnit 20 tests 69 assertions 全绿；端到端 curl 实测通过）。
- **端到端实测结果**：未登录页面 302 /login；未登录 ajax `{"code":"1","msg":"未登录或登录超时。请重新登录"}`（code 字符串特供格式正确）；登录态访问会话数据传递正常且 TTL 续期生效（请求后 TTL≈1762s 证明 <20min 续满 30min 逻辑工作）；防重四场景（首发 0/立即重复 500/异参放行/非 POST 忽略）全对。
- **踩坑（命名空间）**：全局类（RedisCache/TpConstant/PermissionService 等）在 namespaced 文件（app/service、app/middleware）里必须 use 引入或 `\` 前缀，否则报 Class not found——3 处已修（SessionService/RepeatSubmit/CheckPerm）。后续模块注意。
- 实现细节与 spec 目录树的差异（非偏差）：SessionService 落 `app/service/`、DataScope 落 `app/common/`（纯函数无依赖容器）、注解落 `app/attribute/`（Perm/Log 两个 Attribute 类）。
- LoginAuth 匿名路径含 `/index` 与 `/system/main`（对位经典版登录前可访问主框架跳转语义——实际 2.0.0 会挂受保护逻辑，届时按经典版 filter chain 复核）。
