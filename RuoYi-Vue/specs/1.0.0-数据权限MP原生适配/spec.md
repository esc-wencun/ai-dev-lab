# 1.0.0 数据范围过滤 MP 原生适配（纯 MyBatis-Plus 查询的 DataScope 方案）

> **状态：✅ 已完成（2026-09-26）**
>
> 0.0.0 spec 3.5 节"长期红线"的兑现：走 BaseMapper/Wrapper 的查询没有 XML 里的 `${params.dataScope}` 注入点，数据范围过滤需要 MP 原生等价机制。本文件是调研结论、方案设计与实施记录。
> **2026-09-26 参考实现调研**：dcoa-new-admin（芋道二开版）的 `dcoa-spring-boot-starter-biz-data-permission` 是同类问题的成熟生产实现（MP 3.5.10.1 + `DataPermissionInterceptor` + `MultiDataPermissionHandler` 逐表回调 + 规则工厂）。本方案已按其结构校准，逐点对照见第七节。

## 一、问题与背景

现状机制（XML 路径，工作正常）：service 方法标 `@DataScope(deptAlias="d", userAlias="u")` → `DataScopeAspect` 按当前用户角色计算过滤 SQL（如 ` AND (d.dept_id = 100 OR d.dept_id IN (...)) `）→ 写入首参实体的 `params.dataScope` → Mapper XML 用 `${params.dataScope}` 拼进 WHERE。

**纯 MP 查询（BaseMapper / Wrapper）没有这个注入点**：`selectPage(wrapper)` 的 WHERE 由框架生成，没有 `${}` 占位可拼。当前无实际故障（走 Wrapper 的 4 张表 post/config/dict_type/dict_data 均无 dept_id/user_id 列，本就无数据范围语义），但**未来任何带数据范围语义的业务表一旦用 BaseMapper 查询，就会静默查全库**——与 0.0.0 R 系风险同级的越权不可见问题。红线原文：`届时必须先给 MP 侧补等价机制（DataPermissionInterceptor + 自定义数据权限处理器），且补齐后方可迁移`。

## 二、MP 原生能力调研（3.5.17 jar 实查）

| 组件 | 位置 | 作用 |
|------|------|------|
| `DataPermissionInterceptor` | `mybatis-plus-jsqlparser` jar（已是依赖，0.0.0 分页引入） | InnerInterceptor，`beforeQuery` 阶段用 jsqlparser 解析最终 SQL，**逐表回调** handler，把返回的条件 Expression AND 进 WHERE；对 count 语句同样生效（分页 total 口径一致） |
| `DataPermissionHandler` | 同上 | `Expression getSqlSegment(Expression where, String mappedStatementId)`——整条 SQL 一个条件（单表场景） |
| `MultiDataPermissionHandler` | 同上 | `Expression getSqlSegment(Table table, Expression where, String msId)`——**按表回调**，返回 null 表示该表不加条件；join 场景必须用这个 |
| 注册顺序 | MP 官方约定（dcoa 两处注释同款） | 数据权限 InnerInterceptor **插到拦截器链 index 0**，保证先于分页插件执行 |

## 三、dcoa 实现要点（2026-09-26 实查，方案校准依据）

dcoa 的完整链路：`@DataPermission` 注解 → Spring Advisor（`AnnotationMatchingPointcut` 类/方法双匹配 + `MethodClassKey` 缓存注解查找 + NULL 占位对象防缓存穿透）→ `DataPermissionContextHolder`（TTL 栈，装注解不装数据）→ MP `DataPermissionInterceptor`（链首）→ `MultiDataPermissionHandler` 逐表回调 → `DataPermissionRuleFactory`（按上下文注解过滤规则集，**无注解 = 默认全部规则生效**）→ `DeptDataPermissionRule.getExpression(tableName, alias)` 表名命中才生成 JSqlParser `Expression`。

关键结构与结论：

1. **规则抽象 = `getTableNames()` 白名单 + `getExpression(tableName, tableAlias)` 逐表表达式工厂**。两个 Map（`deptColumns`/`userColumns`：表名→列名）就是"表注册表"，与方案 C 草案结构一致；dcoa 用函数式 Customizer Bean（`rule.addDeptColumn(实体类)` / `addDeptColumn("表名","列名")`）填充，支持实体类自动解析表名。
2. **范围语义用结构化 DTO 而非 SQL 字符串**：`DeptDataPermissionRespDTO{all: bool, self: bool, deptIds: Set<Long>}`——5 种 RuoYi 范围类型全部可映射（1全部→all=true；2自定义→deptIds=查 sys_role_dept；3本部门→{本人deptId}；4本部门及以下→部门缓存预展开 ID 集后 IN 而非 find_in_set；5仅本人→self=true）。**范围解析与 SQL 生成解耦**，规则层只拼 `dept_id IN (...) / user_id = ?`，无角色用户默认 self=true。
3. **per-request 缓存必须做**：拦截器对同一查询的每张表都回调 getExpression，范围解析结果（DTO）缓存在 LoginUser.context（key=规则类名），一次请求只算一次；配套 userDeptId 惰性求值、角色/部门长缓存。
4. **兜底语义三件套**：无权限（既无 deptIds 又非 self）→ `new EqualsTo(null, null)`（WHERE null=null 返回空集，不报错不漏数据）；范围 DTO 取不到 → 快速失败 NPE；dept+user 条件并存 → 带括号 `OR` 组合。
5. **注册顺序**：`MyBatisUtils.addInterceptor(interceptor, inner, 0)` 插链首（分页之前），MP 官方规定；dcoa 的租户拦截器同款操作。
6. **JSqlParser 解析缓存**（Caffeine 1024 条 / 5s）：拦截器每条 SQL 都要 parse，这是主要性能开销，两行配置必须配。
7. **编程式豁免** `DataPermissionUtils.executeIgnore(Runnable)`：反射取常量注解单例入栈出栈；典型用途是唯一性校验前关数据权限（否则查不到数据误判"不重复"）。
8. **不照搬**：dcoa 的"子表别名 t1..t6 豁免"约定（依赖 mybatis-plus-join 的别名规则，RuoYi 无 MPJ）；`UserTypeEnum==ADMIN` 口径（与 admin 旁路无关）；`selectById/updateById/deleteById` 内置方法短路（安全宽口子，RuoYi 不做）；TTL 依赖（RuoYi 无重度线程池场景，普通 ThreadLocal 栈够用）；PermissionApi/SecurityFramework（换 RuoYi 自有 ServiceUtils/LoginUser）。

## 四、方案对比

| 方案 | 思路 | 结论 |
|------|------|------|
| A. 全量迁移拦截器 | 删掉 `params.dataScope` 机制，所有查询（含现有 5 处 XML）统一走拦截器 | ❌ 本期不做。动 XML 与切面核心路径，回归面大；`${params.dataScope}` 需要切面永远先 `put("")` 占位否则绑定报错，迁移牵连 5 处 XML 逐个清理 |
| B. Wrapper 手工拼接 | service 里显式调 helper 把范围条件 `wrapper.apply(sql)` | ❌ 不做。依赖每个开发者记得调用，漏调即静默越权——把纪律问题留给人的方案不可取 |
| **C. 拦截器 + 桥接上下文 + Rule（推荐，dcoa 同构）** | 保留现有 `@DataScope` 注解与切面计算逻辑；范围解析产出**结构化 DTO**；注册 MP `DataPermissionInterceptor`，自定义 Handler 在"上下文激活 且 表在规则白名单"时按 DTO 生成 JSqlParser Expression AND 进 WHERE | ✅ 推荐。与 dcoa 生产形态同构；对现有代码零行为变化（白名单空即拦截器不生效）；未来业务表一处注册 + service 标注解即得过滤 |

## 五、推荐方案（C）详细设计（按 dcoa 校准后）

### 5.1 数据流

```
service 方法 @DataScope(deptField="dept_id")
   │ 现有 DataScopeAspect @Before（XML 路径不动：算片段、写 params.dataScope，5 处 XML 行为不变）
   │ └─ 新增：范围解析产出 RuoYiDeptDataPermissionDTO{all, self, deptIds}
   │          写入 DataScopeContextHolder（栈式 ThreadLocal，装"本次请求要过滤"标记 + DTO）
   ▼
mapper.selectPage(wrapper)  ← 纯 MP 查询
   ▼
DataPermissionInterceptor（插链首，先于分页）
   │ RuoYiDataPermissionHandler implements MultiDataPermissionHandler
   │   逐表 getSqlSegment(table, where, msId)：
   │     上下文未激活 → return null（全项目其余查询零影响）
   │     表不在 DeptDataPermissionRule.getTableNames() → return null
   │     命中 → 按 DTO 生成 Expression（dept_id IN / user_id = / null=null 兜底）
   ▼
ContextHolder @After 清理（栈 removeLast + 判空 remove，防泄漏）
```

### 5.2 新增/改造组件

| 组件 | 位置 | 职责 | dcoa 对应物 |
|------|------|------|------------|
| `RuoYiDeptDataPermissionDTO` | `com.ruoyi.common.core.domain.dto`（新建包） | `{all: boolean, self: boolean, deptIds: Set<Long>}` 结构化范围 + `all()`/`none()` 静态工厂 | DeptDataPermissionRespDTO |
| `DataScopeContextHolder` | `com.ruoyi.common.core.mybatis` | ThreadLocal 栈（装 `{enable, DTO}`）；`add/get/remove`，removeLast 后判空 remove 防泄漏；**DTO 入栈时算好，栈即请求级缓存**（多表回调读同一份） | DataPermissionContextHolder（TTL→普通 ThreadLocal） |
| `DeptDataScopeResolver` | framework | **范围语义来源**：输入 LoginUser（取 `user.getRoles()` 登录快照，与切面同源）+ permission 串 → 输出 DTO；判断结构照 DataScopeAspect（admin→all、停用跳过、并集累加）；**切面本体不动**，语义差异点（见 5.3.2）登记实施记录 | （dcoa 的 PermissionApi 等价物） |
| `DataPermissionRule`（接口） | **framework**（返回 jsqlparser Expression，放 common 会污染依赖方向） | `Set<String> getTableNames()` + `Expression getExpression(String tableName, Alias alias, RuoYiDeptDataPermissionDTO dto)` | 同名接口 |
| `RuoYiDataPermissionHandler` | framework | implements `MultiDataPermissionHandler`：规则集空→null；表不命中→null；否则逐规则 AndExpression | DataPermissionRuleHandler |
| 注册 | 现有 `MyBatisConfig` | `DataPermissionInterceptor` 插链首（分页之前）+ JSqlParser Caffeine 解析缓存 | DcoaDataPermissionAutoConfiguration |
| 表注册入口 | system 模块配置类 | `rule.addDeptColumn(实体类)` / `addDeptColumn("表名","列名")` / `addUserColumn(...)`（本期白名单空表起步，注册随首个业务需求加） | SystemDeptDataPermissionRuleCustomizer |
| 豁免 | `DataPermissionUtils.executeIgnore(Runnable)` | 编程式忽略（唯一性校验等场景） | 同名工具（照搬） |

### 5.3 关键设计点

1. **双路径并存、互不重复过滤**：XML 语句走 `params.dataScope`（sys_user/sys_role/sys_dept 不进规则白名单）；纯 MP 查询走拦截器。规则白名单按表名登记，天然互斥。
2. **角色数据源单一（登录快照）**：resolver 从 `LoginUser.user.getRoles()` 取角色——**与 DataScopeAspect 完全同源**，双路径任何时刻行为一致；不另查 sys_user_role 实时库（否则"切面用快照、拦截器用实时库"会分裂）。角色变更经重新登录生效，与现状一致。**切面本体不动**，XML 片段生成与 DTO 生成并存；resolver 判断结构照抄切面（admin→all、停用跳过、permission 串过滤），差异点在实施记录登记。
3. **显式注解才过滤（RuoYi 语义）**：与 dcoa"默认开启"相反——service 标 `@DataScope` 才入栈生效，不标注解的查询（含全部现有 BaseMapper 查询）零影响，保住"零行为变化"红线。
4. **"本部门及以下"**：resolver 复用现有 `selectChildrenDeptById`（一条 SQL 查全部子孙）预展开 ID 集 + `IN`（索引友好），弃用 `find_in_set(ancestors)`；与 XML 路径结果语义一致、实现不同——端到端专测覆盖。
5. **count 口径**：拦截器插链首 → 分页 count 自动带条件，total 与明细一致。
6. **兜底三件套**：无权限（deptIds 空且 self=false）→ `EqualsTo(null,null)` 空集；DTO 取不到 → 快速失败；dept+user 列并存 → 括号 OR。
7. **豁免双通道**：`@DataScope(enable=false)`（注解扩展）或编程式 `DataPermissionUtils.executeIgnore`。
8. **非 Web 上下文**（定时任务）：LoginUser 取不到 → handler 返回 null 全部放行（dcoa 同款语义）。
9. **未来业务表接入成本** = 白名单注册一行 + service 标 `@DataScope`。admin（userId=1）在 resolver 产出 all=true，规则层无特殊分支。
10. **JSqlParser Caffeine 解析缓存的 API**（`JsqlParserGlobal.setJsqlParseCache`）在 3.5.17 是否同名开发期核实，无则降级跳过并登记实施记录。

### 5.4 风险

| # | 风险 | 对策 |
|---|------|------|
| R1 | ThreadLocal 泄漏（异常路径未清理）→ 线程复用跨请求越权 | 栈式 removeLast + 判空 remove；单测覆盖异常分支 |
| R2 | jsqlparser 解析片段失败 | Expression 由 JSqlParser AST 构建（非字符串拼接），语法由类型保证；DTO→Expression 单测逐 scope 断言 |
| R3 | 白名单误登记无范围语义的表 | 登记即生效是设计意图；注册改动必须走 checklist 受限角色实测（0.0.0 数据权限专条模式） |
| R4 | 拦截器对 IPage count 路径与预估不符 | 实测：受限角色翻页 total 必须等于过滤后行数 |
| R5 | 拦截器对全项目每条 SQL 做 jsqlparser parse 的性能开销 | Caffeine 解析缓存（1024/5s）必配；单测带性能冒烟 |

## 六、验收标准（开发期写实到 checklist）

- 单测：DTO→Expression 逐 scope 类型断言（IN/=/(A OR B)/null=null）；ContextHolder 栈语义与异常路径清理；规则白名单命中/未命中
- 端到端（对照 0.0.0 checklist 模式）：造业务测试表 + 白名单注册 + 受限角色（每种 scope 各一）实测查询与 total；admin 全量；非 Web 上下文放行
- 回归：现有 5 处 XML 数据范围行为不变（受限角色实测用户/角色/部门列表）
- 测试数据清理 + admin 会话复原

## 七、dcoa 对照校准记录（2026-09-26）

| 设计点 | 草案（校准前） | 校准后（采纳 dcoa） |
|--------|---------------|---------------------|
| 范围中间表示 | ThreadLocal 装"无别名 SQL 片段字符串" | **结构化 DTO{all, self, deptIds}**——范围语义与 SQL 方言/别名解耦，规则层用 JSqlParser AST 拼 Expression |
| 列配置归属 | 独立静态注册表 | **规则实例内部双 Map**（deptColumns/userColumns），Customizer Bean 填充——支持未来多规则（按商户/按项目）并存 |
| 逐表粒度 | MultiDataPermissionHandler（已对） | 维持，并明确 join 子表不靠别名豁免（dcoa 的 t1..t6 约定不搬），由白名单显式声明 |
| 桥接上下文载体 | ThreadLocal 装片段 | ThreadLocal **栈**装 DTO + enable（支持嵌套）；removeLast 判空 remove |
| 豁免机制 | `@InterceptorIgnore` | 追加编程式 `executeIgnore`（唯一性校验场景刚需）+ 注解 enable=false |
| 缓存 | 未设计 | **per-request 范围 DTO 缓存**（多表回调必须）+ JSqlParser Caffeine 解析缓存（全局必配） |
| 兜底语义 | 未细化 | null=null 空集 / DTO 缺失快速失败 / 括号 OR 三件套 |
| "本部门及以下" | 沿用 find_in_set | 预展开 ID 集 IN（索引友好） |

## 八、实施记录

- 2026-09-26 全部 6 个 Task 完成。最终组件落位：DTO（common/core/domain/dto）、ContextHolder+Utils（common/core/mybatis）、Resolver+Rule+Handler（framework/aspectj 与 framework/mybatis）、拦截器注册在 MyBatisConfig（链首，先于分页）、白名单 Bean 在 ruoyi-admin/web/config（**放 system 会违反依赖方向，已纠**）。
- 坑 1：`ExpressionList` 单元素时 jsqlparser 5.2 生成 `IN 105`（缺括号）非法 SQL——必须用 `ParenthesedExpressionList`（实测踩中后修复，单测同步锁定）。
- 坑 2：切面假定标注解方法首参为 BaseEntity（`getArgs()[0]` 越界）——纯 MP 无参方法场景加防御跳过 XML 路径仅走拦截器桥接。
- 坑 3：`JsqlParserGlobal.setJsqlParseCache` 在 3.5.17 存在（jsqlparser 分册），Caffeine 缓存（1024/5s）已配。
- 决策落定：显式 @DataScope 才过滤（区别 dcoa 默认开启）；角色数据源用登录快照（LoginUser.user.roles，与切面同源）；内置单条方法不短路；非 Web 上下文放行。
- 端到端实测全绿：admin 全量 4 条；受限角色（本部门及以下，dept105）只见 ry-order-1/2，total=2 与明细一致（count 带条件）；executeIgnore 豁免 4 条；无注解放行 4 条；无登录态被 Security 拦截。
- 测试基建已清理：biz_order 表 DROP、探针 controller/mapper/实体删除、测试角色/用户删除、白名单恢复空注册（DataPermissionConfiguration 保留空 DeptDataPermissionRule 供未来登记）、会话清空。
