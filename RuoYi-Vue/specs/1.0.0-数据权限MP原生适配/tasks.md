# Tasks · 1.0.0 数据权限 MP 原生适配

> 勾选纪律：做完即勾（含对应验证跑通）；没做的不许勾，行尾注明原因；勾选 = 验收通过。
> 依赖：0.0.0（MP 3.5.17 就位、分页插件已注册）。方案依据：spec.md 第三/五/七节（dcoa 校准版）。
> 环境约束：编译验证遵循项目 AGENTS.md《编译纪律》A 类（spec 任务可自主编译测试，动 mvn 前先停 IDEA 里的服务并在回复中明说）；数据权限验证必须用受限角色实测（admin 全权限测不出失效）。

## Task 1: 范围 DTO 与范围解析器（语义层，不碰 SQL）

- [x] `RuoYiDeptDataPermissionDTO` 新建（`com.ruoyi.common.core.domain.dto` 新建包）：`{all: boolean, self: boolean, deptIds: Set<Long>}` + 静态工厂 `all()` / `none()`
- [x] `DeptDataScopeResolver` 新建（framework）：输入 LoginUser + permission 串 → DTO；角色从 `LoginUser.user.getRoles()` 登录快照取（与 DataScopeAspect 同源，不查实时库）；判断结构照抄切面——admin→all、停用跳过、permission 串过滤、多角色并集累加、任一 all 即 all；2自定义→查 sys_role_dept（`selectDeptListByRoleId(roleId, false)` 取全量 ID 集）；4本部门及以下→复用 `selectChildrenDeptById` 预展开 + 本部门；3本部门→{本人deptId}；5仅本人→self=true；无角色→self=true
- [x] `@DataScope` 注解扩展：新增 `boolean enable() default true`（豁免通道，向后兼容——既有 5 处用法零改动）
- [x] 单元测试：7 种输入（5 种 scope + 多角色混合 + admin + 无角色）→ DTO 断言（mock mapper，不连库）

## Task 2: 桥接上下文（ContextHolder）

- [x] `DataScopeContextHolder` 新建（`com.ruoyi.common.core.mybatis`）：ThreadLocal 栈（`Deque`）装 `{enable, DTO}`；`add/get/peek/remove`；removeLast 后栈空即 `ThreadLocal.remove()` 防泄漏
- [x] `DataPermissionUtils.executeIgnore(Runnable)` / `<T> executeIgnore(Callable<T>)`：编程式豁免
- [x] `DataScopeAspect` 改造：@Before 在现有 params.dataScope 写入之后**新增**——resolver 算 DTO 入栈（enable 取注解 enable 属性）；**新增 @After（finally 语义）出栈清理**；XML 路径 params.dataScope 行为零变化
- [x] 单元测试：栈嵌套（两层 add/peek/remove）；异常路径清理（切面方法抛异常后栈仍清空）；executeIgnore 期间 get 返回 enable=false

## Task 3: 规则与 Handler（SQL 生成层）

- [x] `DataPermissionRule` 接口新建（**framework**——返回 jsqlparser Expression，不进 common）：`Set<String> getTableNames()` + `Expression getExpression(String tableName, Alias tableAlias, RuoYiDeptDataPermissionDTO dto)`
- [x] `DeptDataPermissionRule` 新建（framework）：内部 `deptColumns`/`userColumns` 两个 `Map<String,String>`；注册方法 `addDeptColumn(Class<?>)`（TableInfoHelper 解析表名，默认列 dept_id）/ `addDeptColumn(String table, String column)` / `addUserColumn(Class<?>, String column)` / `addUserColumn(String table, String column)`；`getTableNames()` 返回双 Map key 合集
- [x] `getExpression` 实现（按 DTO 拼 JSqlParser AST，参照 dcoa DeptDataPermissionRule）：all→null；无 deptIds 且非 self→`EqualsTo(null,null)` 空集兜底；dept 维度→`dept_id IN (deptIds)`；user 维度→`user_id = userId`；并存→括号 OR；该表无对应列配置→该维度返回 null
- [x] `RuoYiDataPermissionHandler` 新建（framework）implements `MultiDataPermissionHandler`：ContextHolder 未激活或 enable=false→null；逐规则表名命中判断（剥反引号），未命中→null；命中→getExpression
- [x] 表白名单注册入口：system 模块配置类提供 `@Bean` 组装（本期**白名单为空表起步**——现无带范围语义的纯 MP 查询）；注释写明"注册即生效，改动必须走 checklist 数据权限实测"
- [x] 单元测试：DTO→Expression 逐 scope 断言（生成 SQL 字符串比对：IN/=/(A OR B)/null=null）；白名单命中与未命中；缺列配置时维度降级

## Task 4: 拦截器注册与装配（装配层）

- [x] `MyBatisConfig`：`DataPermissionInterceptor`（构造传 handler）插拦截器链 **index 0**（先于分页插件）；JSqlParser 解析缓存——先核实 3.5.17 是否暴露 `JsqlParserGlobal.setJsqlParseCache`，有则配 Caffeine（1024/5s），无则降级跳过并登记实施记录
- [x] DTO 请求级缓存确认：ContextHolder 入栈时算好 DTO（Task 2 已含），handler 多表回调只读不重算——本 Task 仅验证
- [x] 非 Web 上下文行为：`SecurityUtils.getLoginUser()` 抛错/取不到 → resolver 返回 all=true 之外的特殊值或 handler 直接 null 放行（实现时定为"切面入栈前拦截：取不到 LoginUser 则不入栈"，handler 自然放行；单测断言）
- [x] 编译 + 全模块 `mvn clean package`（按编译纪律 A 类：8080 当前空闲可直接编译；编译完不留后台进程）

## Task 5: 端到端验证（数据权限专测，对照 0.0.0 checklist 模式）

- [x] **回归基线**：现有 5 处 XML 数据范围行为不变——造受限角色（每种 scope 各一）+ 测试用户实测用户/角色/部门列表过滤与 PageHelper→MP 分页 total 口径
- [x] **拦截器路径实测**：白名单临时注册一张测试表（带 dept_id 列）+ BaseMapper 查询——受限角色各 scope 下查询结果与 total 正确过滤；admin 全量；`executeIgnore` 豁免生效；非 Web 上下文（无登录态调用 mapper）放行
- [x] **性能冒烟**：核心列表接口（用户/岗位）带缓存前后响应时间对比，登记数据
- [x] 测试表/测试角色/测试用户全部清理；admin 会话复原；**白名单恢复为空表**（临时注册仅测试期）
- [x] 实施记录回写本 spec 第八节（含实测发现的坑，格式对照 0.0.0 坑 1~8）

## Task 6: 文档同步

- [x] 项目 `AGENTS.md` MP 使用约定补充：数据范围过滤双路径说明（XML 走 params.dataScope / 纯 MP 走规则白名单 + @DataScope）、"带范围语义的表接 BaseMapper 前必须先白名单注册"
- [x] 0.0.0 spec 3.5"长期红线"更新为指向本模块（机制已就位，红线解除条件 = 白名单注册 + 受限角色实测）
- [x] `RuoYi-Vue/specs/README.md` 模块总表本模块标 ✅ + 日期（三件套全部勾选后才许标）
