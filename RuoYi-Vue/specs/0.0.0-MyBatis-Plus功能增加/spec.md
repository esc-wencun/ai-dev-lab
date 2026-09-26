# 0.0.0 MyBatis-Plus 功能增加：删 PageHelper 切 MP 分页 + BaseService 抽取

> **状态：✅ 已完成（2026-09-26）**
>
> Java 版第一个自研演进模块。三个决策（2026-09-25 用户定）：
> 1. **pagehelper-spring-boot-starter 删除**，分页全面切换到 MyBatis-Plus 分页插件——不是共存，是替换；
> 2. Service 层新增 **BaseService** 抽取公用方法（唯一性校验、分页组装等样板）；
> 3. 本模块**只出文档先行**（spec/tasks/checklist 三件套），按勾选纪律逐项开发。
>
> 版本结论（2026-09-25 经阿里云 Maven 镜像实查）：Spring Boot 4.1.0 必须用 `com.baomidou:mybatis-plus-spring-boot4-starter:3.5.17`（SB4 专用 starter 自 3.5.13 起提供）；分页插件自 3.5.9 起拆分出 `com.baomidou:mybatis-plus-jsqlparser`，需单独引入（同 3.5.17）。

## 一、目标与范围

**做**：
- 引入 MyBatis-Plus（下称 MP），改造 `MyBatisConfig` 使 MP 增强生效（BaseMapper / Wrapper / 分页 / 逻辑删除可用）；
- **彻底删除 PageHelper**：依赖、`PageUtils`、`BaseController.startPage()/startOrderBy()/clearPage()`、16 处 `startPage()` 调用点、代码生成模板，全部改造；
- 新建 `BaseService` 层基类，抽取 ServiceImpl 公用逻辑；首批 8 个简单 CRUD 的 ServiceImpl 继承（复杂的后续模块再议，见 3.6）；
- 所有分页接口对外行为不变：请求参数仍是 `pageNum/pageSize/orderByColumn/isAsc`，响应仍是 `TableDataInfo{total, rows, code, msg}`。

**不做**（后续模块再议）：
- 存量 XML 复杂 join 查询（`selectUserList` 带 dept/roles 关联的）不重写成 Wrapper——保留 XML，MP 与 XML 共存；
- 代码生成器模板只改分页相关两处，不做全套 MP 风格模板改造；
- 不引 MP 的 IService/ServiceImpl 体系（`ServiceImpl<M, T>`）：与本项目自建 BaseService 思路冲突，二选一，选自建（轻、可控、对齐三端分层叙事）。

## 二、现状盘点（动手前必读，2026-09-25 实查）

| # | 事实 | 位置 |
|---|------|------|
| 1 | Spring Boot 4.1.0 + Java 17；`mybatis-spring-boot-starter` 4.1.0；MyBatis 依赖实际是 **pagehelper-starter 传递引入**（根 pom 只在 dependencyManagement 声明过版本，无任何模块显式依赖） | [pom.xml:21](../../pom.xml)、[ruoyi-common/pom.xml:38-42](../../ruoyi-common/pom.xml) |
| 2 | **手工 `SqlSessionFactoryBean`**（为支持 `com.ruoyi.**.domain` 通配别名扫描）——与 MP 自动配置冲突，见风险 R1 | [MyBatisConfig.java:116-131](../../ruoyi-framework/src/main/java/com/ruoyi/framework/config/MyBatisConfig.java) |
| 3 | 分页 = PageHelper ThreadLocal 模式：controller `startPage()` → 拦截第一条 SQL；`getDataTable(list)` 里 `new PageInfo(list).getTotal()` 取总数 | [BaseController.java:53-91](../../ruoyi-common/src/main/java/com/ruoyi/common/core/controller/BaseController.java) |
| 4 | `startPage()` 调用点共 **16 处 / 12 个 controller + 1 个模板**（见附录 A 清单），另有 `startOrderBy` 1 处（BaseController，无调用者——顺带删） | 附录 A |
| 5 | 分页参数解析：`TableSupport.buildPageRequest()` 从请求参数取 `pageNum/pageSize/orderByColumn/isAsc/reasonable`，`PageDomain.getOrderBy()` 负责驼峰→下划线 + `ascending/descending` 兼容——**这套参数解析与 PageHelper 无关，全部保留复用** | [TableSupport.java](../../ruoyi-common/src/main/java/com/ruoyi/common/core/page/TableSupport.java)、[PageDomain.java](../../ruoyi-common/src/main/java/com/ruoyi/common/core/page/PageDomain.java) |
| 6 | 代码生成模板 `controller.java.vm` 生成 `startPage(); ... getDataTable(list)` | [controller.java.vm:45-51](../../ruoyi-generator/src/main/resources/vm/java/controller.java.vm) |
| 7 | `@MapperScan("com.ruoyi.**.mapper")` 已有 | [ApplicationConfig.java:16](../../ruoyi-framework/src/main/java/com/ruoyi/framework/config/ApplicationConfig.java) |
| 8 | `mapUnderscoreToCamelCase` 在 mybatis-config.xml 里被注释掉（全靠 XML resultMap 映射）；MP 默认开启驼峰映射——新增 BaseMapper 单表查询依赖此默认值，行为正确，无需改 XML | [mybatis-config.xml:17](../../ruoyi-admin/src/main/resources/mybatis/mybatis-config.xml) |
| 9 | Service 层无基类：全项目 19 个 @Service 类（system 13、generator 2、quartz 2、framework 2），重复样板集中在唯一性校验（`checkXxxUnique` 9 处同构：查库比 ID，排除自身）与 CRUD 透传 | [SysPostServiceImpl.java](../../ruoyi-system/src/main/java/com/ruoyi/system/service/impl/SysPostServiceImpl.java)（典型样本） |

## 三、技术方案

### 3.1 依赖变更（Task 1）

根 `pom.xml` properties + dependencyManagement：

```xml
<mybatis-plus.version>3.5.17</mybatis-plus.version>

<dependency>
    <groupId>com.baomidou</groupId>
    <artifactId>mybatis-plus-spring-boot4-starter</artifactId>
    <version>${mybatis-plus.version}</version>
</dependency>
<dependency>
    <groupId>com.baomidou</groupId>
    <artifactId>mybatis-plus-jsqlparser</artifactId>
    <version>${mybatis-plus.version}</version>
</dependency>
```

`ruoyi-common/pom.xml`：删 `pagehelper-spring-boot-starter`，加以上两个依赖。
**注意**：`mybatis-spring-boot-starter`（org.mybatis.spring.boot）维持 dependencyManagement 声明不动——MP starter 传递依赖它提供 mybatis-spring；PageHelper 删除后它是 MP 的传递依赖，显式声明与否在 Task 1 以 `mvn dependency:tree` 实际结果为准，能省则省。

### 3.2 MyBatisConfig 改造（Task 2，本模块技术核心）

**冲突原理**：MP 自动配置类带 `@ConditionalOnMissingBean(SqlSessionFactory)`。本项目手工定义了 `sqlSessionFactory` Bean → MP 自动配置整体退让 → `MybatisSqlSessionFactoryBean`（MP 的工厂子类，负责向 BaseMapper 注入 CRUD SQL）不会被创建。症状：**启动正常、XML 查询正常，一调 BaseMapper 方法就报 `Invalid bound statement (not found)`**——启动期不报错，必须按本方案显式改造。

**改造**：保留 `MyBatisConfig` 类与 `setTypeAliasesPackage`（`**` 通配扫描能力不能丢），仅替换工厂实现：

```java
@Bean
public SqlSessionFactory sqlSessionFactory(DataSource dataSource) throws Exception
{
    // typeAliasesPackage / mapperLocations 解析逻辑原样保留 ...

    MybatisSqlSessionFactoryBean factory = new MybatisSqlSessionFactoryBean();  // MP 工厂，替代 SqlSessionFactoryBean
    factory.setDataSource(dataSource);
    factory.setTypeAliasesPackage(typeAliasesPackage);
    factory.setMapperLocations(resolveMapperLocations(StringUtils.split(mapperLocations, ",")));
    factory.setConfigLocation(new DefaultResourceLoader().getResource(configLocation));  // 与 setConfiguration 二选一，沿用 XML

    // 分页插件（3.5.9+ 需 mybatis-plus-jsqlparser 单独依赖）
    MybatisPlusInterceptor interceptor = new MybatisPlusInterceptor();
    interceptor.addInnerInterceptor(new PaginationInnerInterceptor(DbType.MYSQL));
    factory.setPlugins(interceptor);

    // 逻辑删除全局值：对齐 RuoYi del_flag 语义（0 正常 / 2 删除）
    GlobalConfig globalConfig = new GlobalConfig();
    GlobalConfig.DbConfig dbConfig = new GlobalConfig.DbConfig();
    dbConfig.setLogicDeleteValue("2");
    dbConfig.setLogicNotDeleteValue("0");
    globalConfig.setDbConfig(dbConfig);
    factory.setGlobalConfig(globalConfig);
    return factory.getObject();
}
```

约束：`setConfigLocation` 与 `setConfiguration` 互斥（MyBatis 规定），沿用现有 `mybatis-config.xml` 不改。

### 3.3 分页切换：新分页支撑类（Task 3）

保留 `TableSupport` / `PageDomain`（纯参数解析，与 PageHelper 无关），删除 `PageUtils`（`extends PageHelper` 无可救药），新增 `PageUtils` 的 MP 实现（路径不变，`com.ruoyi.common.utils.PageUtils`，调用点改动最小）：

```java
public class PageUtils
{
    /** 构建 MP 分页对象（pageNum/pageSize/orderBy 从请求参数解析，orderBy 已做注入转义与驼峰转下划线） */
    public static <T> Page<T> buildPage()
    {
        PageDomain pageDomain = TableSupport.buildPageRequest();
        Page<T> page = new Page<>(pageDomain.getPageNum(), pageDomain.getPageSize());
        page.setReasonable(pageDomain.getReasonable());
        String orderBy = pageDomain.getOrderBy();          // "column asc" 形式
        if (StringUtils.isNotEmpty(orderBy))
        {
            page.addOrder(OrderItem.asc(...));             // 按空格拆成 OrderItem；沿用 SqlUtil.escapeOrderBySql 转义
        }
        return page;
    }
}
```

设计要点：
- **防御性分页上限**：`pageSize` 无上限是现状行为，迁移保持不变（不悄悄加上限改变契约），但在 spec 实施记录里登记该已知行为，是否加上限留待用户后续决策；
- **reasonable（pageNum 超界回正）无 MP 等价物（2026-09-26 开发实查修正：`Page` 无 `setReasonable`，立项时的预判有误）**：PageHelper 的 reasonable=true 超界回**末页**，MP 对应能力是 `PaginationInnerInterceptor.setOverflow(true)` 超界回**首页**且属全局插件配置，语义不同。本期不启用 overflow（超界返回空列表）——RuoYi-Vue3 前端 el-pagination 由 total 驱动，用户正常操作触不到超界；该边角差异登记进 checklist 遗留，如需保护后续全局开 overflow。`reasonable` 请求参数本身继续被 `PageDomain` 解析（保留兼容），只是不再传给 MP；
- 排序仍走 `PageDomain.getOrderBy()` 的驼峰→下划线转换，`SqlUtil.escapeOrderBySql` 注入转义不丢。

`BaseController` 改造：
- **Task 3 只新增重载** `protected TableDataInfo getDataTable(IPage<?> page)`（rows=page.getRecords()，total=page.getTotal()，code/msg 同现状）。注意 MP 的 `Page` 实现的是 `IPage` 接口、**不是 `java.util.List`**，不能塞进现有 `getDataTable(List<?>)` 签名——必须重载，不要在 List 签名里 instanceof 判断；
- `startPage()` / `startOrderBy()` / `clearPage()` 与 pagehelper import **本 Task 不删**（16 处调用点还在调 `startPage()`，提前删必然编译失败）——Task 8 随 PageHelper 一起删（R2 的"依赖最后删"）；
- 原 `getDataTable(List<?>)` 保留不动（存量非分页调用与模板 `AjaxResult` 分支还依赖它），Task 8 时其内部 `new PageInfo(list).getTotal()` 同步退化为 `list.size()`。

**controller 调用点新模式**（16 处统一改法，对照 [SysPostController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysPostController.java) :44 起）：

```java
// before
startPage();
List<SysPost> list = postService.selectPostList(post);
return getDataTable(list);

// after
Page<SysPost> page = postService.selectPostPage(post);
return getDataTable(page);
```

### 3.4 Mapper / Service 改造模式（Task 4 试点、Task 5~7 推广）

**Mapper**：接口继承 `BaseMapper<T>`，单表 CRUD 方法（`selectPostById` / `insertPost` / `updatePost` / `deletePostById(s)`）逐个核对后删除，由 BaseMapper 等价方法替代；XML 中对应语句删除。**join / 模糊条件查询（`selectXxxList` 带 `<if>` 动态条件的）保留 XML**——XML 与 BaseMapper 共存无冲突。

**分页查询组装模式**：各模块 `selectXxxPage` 的动态条件用 `LambdaQueryWrapperX` 组装，**写在 Mapper 接口的 default 方法里**（wrapper 即文档，Service 一行透传），列名一律走方法引用（编译期检查、重构安全）：

```java
// Mapper 接口内（对照参考实现的标准链路）
default Page<SysPost> selectPostPage(SysPost post)
{
    return selectPage(PageUtils.buildPage(), new LambdaQueryWrapperX<SysPost>()
            .likeIfPresent(SysPost::getPostName, post.getPostName())    // 值为空自动跳过，等价 XML <if test="...">
            .eqIfPresent(SysPost::getStatus, post.getStatus())
            .orderByAsc(SysPost::getPostSort));
}
```

`LambdaQueryWrapperX` 的方法族：`eq/ne/gt/ge/lt/le/like/in/between` 各配 `IfPresent` 变体（条件值为空/null 自动跳过）；`betweenIfPresent` 半开区间自动退化（只传下界 → ge，只传上界 → le）；重写父类链式方法返回 `LambdaQueryWrapperX` 类型防止断链。此外单条/计数/列表快捷方法（`selectOne(field, value)` / `selectList(field, values)` 等）可选补充，其中**列表版必须做空集合短路**（入参集合为空直接返回空列表，避免生成非法 `IN ()` SQL——原生 BaseMapper 的坑）。

替代等价性核对基准（每个方法迁移前对照）：

| 旧 XML 方法 | BaseMapper 替代 | 核对要点 |
|---|---|---|
| `selectXxxById` | `selectById(id)` | resultMap 特殊映射（如 dept 关联）的不能删，保留 XML 方法 |
| `insertXxx` | `insert(entity)` | 回填主键行为：XML 有 `useGeneratedKeys`，BaseMapper 默认同样回填（`@TableId(type = IdType.AUTO)` 需标注） |
| `updateXxx` | `updateById(entity)` | **XML 是显式 `<if test="xxx != null">` 判空更新，`updateById` 默认同策略（null 不更新）**——语义一致，逐字段核对一遍 |
| `deleteXxxById(s)` | `deleteById / deleteBatchIds` | 注意 sys_user 等带 `del_flag` 的表：XML 里删除即 `update del_flag='2'`——迁移后用 `@TableLogic` 让 BaseMapper 自动生成等价 update，或保留 XML，二选一，在各模块 Task（4~7）逐表决定 |
| `checkXxxUnique` | `selectOne(wrapper)` | 收进 BaseService，见 3.6 |

**join / DataScope 列表的分页切换模式（user/role 共 4 处专用）**：`selectUserList` / `selectAllocatedList` / `selectUnallocatedList` / `selectRoleList` 是带 join + `${params.dataScope}` 的 XML 查询，不迁 Wrapper，SQL 本体（SELECT/WHERE 逻辑）不动，只切分页方式——**Mapper 方法第一个参数加 `IPage<T> page`、实体参数标 `@Param`（MP 约定：方法含 IPage 参数时分页插件自动拦截拼 limit + count，返回 IPage）**。注意 MyBatis 多参数方法必须命名引用：XML 内 `#{userName}` 参数引用与 `<if>` test 表达式需同步加前缀（`#{user.userName}` / `test="user.userName != null"` / `${user.params.dataScope}`）——SQL 逻辑零改动，仅参数引用前缀化，逐行机械替换、diff 可追溯：

```java
// SysUserMapper.java：方法签名加 IPage 首参 + @Param；XML 的 selectUserList 仅参数引用加前缀
IPage<SysUser> selectUserList(IPage<SysUser> page, @Param("user") SysUser user);
```

service 对应加 `Page<SysUser> selectUserPage(SysUser user)`（内部 `selectUserList(PageUtils.buildPage(), user)` 转 Page），controller 拿 Page 走 `getDataTable(IPage)` 重载。**这是 join+DataScope 查询切 MP 分页的唯一安全做法**：`${params.dataScope}` 注入点与分页 count 语句生成互不干扰（count 会带上 dataScope 条件，total 口径与 PageHelper 时代一致）。

**Domain 实体**：需要被 BaseMapper 操作的实体才做 MP 注解改造（`@TableName` / `@TableId` / `@TableField(exist=false)`）。首批清单：SysPost、SysNotice、SysConfig、SysDictType、SysDictData、SysOperLog、SysLogininfor、SysJob、SysJobLog、GenTable、GenTableColumn（分页相关 11 张单表）。**注意 SysUser/SysRole/SysMenu/SysDept 本期不做实体注解改造**（join 复杂、级联多，XML 保留即可；它们的分页切换走上面的 IPage 首参模式，不依赖实体注解）。`BaseEntity` 的 `searchValue`/`params` 加 `@TableField(exist = false)`——**这是共用基类，加了不影响未迁移实体**。

### 3.5 权限拦截影响评估（2026-09-25 实查，本期零改动 + 一条长期红线）

功能权限与横切 AOP（`@PreAuthorize` / `@Log` / `@RepeatSubmit` / 限流 / Spring Security 过滤链）全部工作在 Web/Service 调用层，与 ORM 选型无关——**零影响，不改**。

数据权限（`@DataScope`）的实现链路与 ORM 强耦合，是本模块唯一的权限侧关注点：

- 切面 [DataScopeAspect.dataScopeFilter](../../ruoyi-framework/src/main/java/com/ruoyi/framework/aspectj/DataScopeAspect.java) 在 **Service 层**把过滤 SQL 拼进实体的 `params.dataScope`；
- Mapper XML 用 `${params.dataScope}` 注入（实查共 **5 处**：SysUserMapper.xml ×3（selectUserList/allocated/unallocated）、SysRoleMapper.xml（selectRoleList）、SysDeptMapper.xml（selectDeptList））；
- 对应带 `@DataScope` 注解的 service 方法恰好 5 个（User×3 / Role×1 / Dept×1）。

**本期不改的依据**：这 5 个查询全是带 join 的复杂列表查询，命中本 spec"join 保留 XML 不迁"的范围决策，`${params.dataScope}` 注入点原样保留，机制不受影响。

**长期红线（后续模块动这 5 个查询前必读）**：~~若未来把任一 `${params.dataScope}` 查询迁到 BaseMapper/Wrapper，注入点即消失，数据权限将静默失效~~ **红线已于 1.0.0 解除**：MP 侧等价机制已就位（`DataPermissionInterceptor` + 规则白名单 + 范围 DTO，见 Java 版 1.0.0 spec）——迁 Wrapper 前先把该表注册进 `DataPermissionConfiguration` 白名单并用受限角色实测；XML 侧 5 处 `${params.dataScope}` 保持现状不受影响。

### 3.6 BaseService 设计（Task 4 创建、Task 5~7 推广）

位置：`com.ruoyi.common.core.service.BaseService`（ruoyi-common，与 BaseController 对称）。`checkUnique` 采用新增/编辑合一签名：keyValue 传 null = 新增校验，传当前 id = 编辑时排除自身。

```java
public class BaseService<M extends BaseMapper<T>, T>   // 泛型：注入 mapper，供公用方法使用
{
    @Autowired
    protected M baseMapper;

    /** 唯一性校验：true=唯一。keyValue 传 null = 新增校验；传当前 id = 编辑时排除自身。
        对位 9 处 checkXxxUnique 的同构逻辑（SysPostServiceImpl.checkPostNameUnique 样本） */
    public boolean checkUnique(SFunction<T, ?> keyField, Object keyValue, SFunction<T, ?> field, Object value)
    {
        LambdaQueryWrapper<T> wrapper = new LambdaQueryWrapper<T>().eq(field, value);
        if (keyValue != null)
        {
            wrapper.ne(keyField, keyValue);   // 编辑场景排除自身
        }
        return baseMapper.selectCount(wrapper) == 0;
    }
}
```

注意：RuoYi 的唯一性校验普遍带"软删表仍算唯一"语义（del_flag 不在条件里），`checkUnique` 不加逻辑删除过滤的行为需与现状一致——BaseMapper `selectCount` 在标了 `@TableLogic` 的实体会自动加 `del_flag='0'`，**与现状不一致**（现状是全表查）。迁移时逐表核对：本期 11 张单表里 sys_config/sys_dict_type/post 等无 del_flag 列的不受影响；有 del_flag 的表保持 XML 版 `checkXxxUnique` 不迁。

存量 ServiceImpl 接入：首批 8 个简单 CRUD 的（Post/Config/DictType/DictData/OperLog/Logininfor 6 个 + quartz 的 Job/JobLog 2 个）extends BaseService；`SysUserServiceImpl` 等复杂的本期不强制（多 mapper 注入场景泛型注入不适用，不强扭）。

### 3.7 代码生成模板（Task 8）

[controller.java.vm](../../ruoyi-generator/src/main/resources/vm/java/controller.java.vm) 的 `startPage()` 分支改为新分页模式（`Page<${ClassName}> page = ...getDataTable(page)`）；[mapper.java.vm](../../ruoyi-generator/src/main/resources/vm/java/mapper.java.vm) 与 [mapper.xml.vm](../../ruoyi-generator/src/main/resources/vm/xml/mapper.xml.vm) 是否改为 BaseMapper 风格，本期只评估记录、不实施（避免扩大面）。

## 四、风险清单（开发时逐条对答）


| # | 风险 | 对策 |
|---|------|------|
| R1 | 手工 SqlSessionFactory 吞掉 MP 自动配置 → BaseMapper 全部失效（启动不报错，运行期才炸） | 3.2 显式换 `MybatisSqlSessionFactoryBean`；Task 2 验收标准 = 启动后用一个 BaseMapper 方法跑通 |
| R2 | PageHelper 删除后，仍存活的 `startPage()` 调用点编译失败——**删依赖必须与 16 处调用点改造同一批次完成**，不许中间态提交 | Task 排序按"依赖最后删"：Task 3 新分页就位 → Task 4~7 全部调用点改完 → Task 8 最后删依赖。附录 A 是改造完成的自查清单 |
| R3 | 分页 total 口径：PageHelper 拦截第一条 SQL 取 total；MP 分页 count 由 PaginationInnerInterceptor 生成——两者对 `selectUserList` 这类带 left join + 动态条件的 SQL，count 语句形态不同，理论上 total 一致，但**每个迁移页面都要前端实际翻页验证**，不能只看第一页 | checklist 逐页面勾选"翻页/跳页/排序/筛选后翻页" |
| R4 | `updateById` 判空更新语义与 XML `<if>` 不完全等价的边角：XML 可按列定制（如某些列永不为 null 但参与更新），BaseMapper 做不到 | 迁移核对表逐列过；不放心的一律保留 XML，不硬迁 |
| R5 | `@TableLogic` 自动过滤 del_flag 与既有 `checkXxxUnique` 全表查询语义冲突（3.6 末段） | 有 del_flag 的表唯一性校验保留 XML；或该表不标 @TableLogic |
| R6 | MP 驼峰映射默认开启 vs 现有 resultMap：共存无冲突，但 BaseMapper 生成的 SQL 依赖实体注解齐全——`@TableField(exist=false)` 漏标一个关联字段（dept/roles/params/searchValue）运行即报"Unknown column" | Task 4~6 实体改造逐字段过；验收时打开 mybatis SQL 日志抽查生成语句 |
| R7 | quartz 模块（ruoyi-quartz）依赖 ruoyi-common，PageHelper 删除影响同样波及 `SysJobLogController`/`SysJobController`——跨模块改动不要漏 | 附录 A 含 quartz 两处 |
| R7a | **IPage 首参模式的 count 语句**：MP 对带 join 的 XML 查询自动生成的 count 语句可能把 select 列表包成子查询（`SELECT COUNT(*) FROM (原SQL)`），join 复杂时 count 性能不如 PageHelper 的智能改写。user/role 4 处迁移后若 count 明显变慢，可评估关闭自动 count（`Page.setSearchCount(false)`）配手写 count 语句，或接受现状（管理后台量级小） | Task 7 端到端时对比翻页响应时间；仅在明显劣化时登记优化项 |
| R8 | Python/Go 版对位：本项目 Python 侧 spec-00 记录"模块数≥4 后评估 sqlalchemy-crud-plus"——Java 基准切 MP 后，该评估口径需更新 | Task 9 文档同步（含 Go 版 tech-stack ORM 行备注） |
| R9 | MP 排序走 `OrderItem` 拼接 SQL（列名字符串进 ORDER BY），客户端可控的排序列若不做防护属注入敞口 | `PageUtils.buildPage` 组装 OrderItem 前沿用 `SqlUtil.escapeOrderBySql` 转义（现状已有）；列名来自 `PageDomain.getOrderBy()` 的驼峰转下划线输出，与 PageHelper 时代同源，行为不变 |

## 五、对外契约影响评估（按根 AGENTS.md 纪律）

- 接口路径 / 请求参数 / 响应 JSON 结构 / 状态码约定：**零变化**。`pageNum/pageSize/orderByColumn/isAsc` 解析逻辑原样保留，`TableDataInfo` 结构不变。
- 前端 RuoYi-Vue3：零改动；Python 版 / Go 版：零改动（分层内部实现变更不影响契约）。
- 文档同步（Task 9）：Python 版 `specs/README.md` 技术对位表 MyBatis 行、GO 版 `specs/tech-stack.md` ORM 行补"Java 侧已切 MyBatis-Plus"备注；工作区 readme.md 的 ORM 行同步；根 AGENTS.md 无需动（不含 ORM 细节）。

## 六、实施记录

- 2026-09-26 Task 1 完成：MP 3.5.17 引入，`dependency:tree` 复核无版本漂移（mybatis:3.5.19 由 mybatis-spring-boot-starter 4.1.0 统一提供）。
- 2026-09-26 坑 1：**MP 3.5.9+ 的 `MybatisSqlSessionFactoryBean` 不在 `com.baomidou.mybatisplus.extension.spring`（旧文档/常见教程路径），而在 `com.baomidou.mybatisplus.spring`**（`mybatis-plus-spring` 分册，随 SB4 starter 传递引入）。import 照抄旧路径编译报"找不到符号"。3.2 示例代码已同步修正。
- 2026-09-26 坑 2：**MP `Page` 没有 `setReasonable`**（立项时预判"语义一致"有误）——reasonable 是 PageHelper 特有概念，MP 对应物是 `PaginationInnerInterceptor.setOverflow`（超界回首页，语义不同且是全局配置）。本期不启用 overflow，详见 3.3 修正说明。
- 2026-09-26 坑 3：**项目原默认 surefire 2.12.4（Maven 2 时代）不识别 JUnit 5**，测试静默跑 0 个不报错——根 pom 已显式升 surefire 3.5.3。给老项目配测试基建先查 surefire 版本。
- 2026-09-26 坑 4：**mapper 接口方法与 XML 语句同 ID 共存时，SQL 走 XML**——Task 4 首轮验证发现 service 调 `baseMapper.insertPost()` 仍执行 XML 语句（日志小写 `insert into`），XML CRUD 删掉后才真正切到 BaseMapper 模板（大写 `INSERT INTO`）。验收"走没走 MP"以 SQL 日志大小写/语句形态为准，不能只看接口编译。
- 2026-09-26 坑 5：**RuoYi 的 XML insert/update 用 `sysdate()` 兜底时间字段**，controller 惯例只 set createBy/updateBy 不 set 时间——BaseMapper 判空插入下 create_time/update_time 会变 NULL。对策：service insertPost/updatePost 里显式赋值 `new Date()` 兜底（Task 4 已落）。后续模块沿用此模式（Task 5~7 勿漏）。
- 2026-09-26 坑 6：**无审计列表（sys_oper_log/sys_logininfor 等）不能继承 BaseEntity 走 BaseMapper**——MP 实体解析含父类字段，`@TableField(exist=false)` 是字段级注解无法按表定制，BaseMapper 的 SELECT/INSERT 会带上 `create_by` 等不存在列直接 SQL 报错（实测确认）。对策：此类表的 Mapper 不继承 BaseMapper，分页走 spec 3.4 的 IPage 首参 + XML 模式（Task 6 已落，service 不继承 BaseService）。后续模块遇到同型表（如日志/流水类）沿用此模式。
- 2026-09-26 坑 7：**XML 参数前缀化的完整性**——`<if test>` 里除 `xxx != null` 外，**比较用的裸属性名**（`xxx != ''`）也必须加前缀，漏改后半截编译不报错、运行期才炸 `Parameter not found`；且脚本对同一条语句重复前缀化会叠加（`genTable.genTable.genTable...`，实测踩中）——XML 前缀化改动先 `git checkout` 还原再一次性手工/脚本完成。
- 2026-09-26 坑 8：**IPage 首参模式的 service Page 方法必须带与 List 版相同的 `@DataScope`**——数据权限切面按注解在 service 层拦截，Page 方法漏注解则 `${user.params.dataScope}` 为空、查询退化为全库（静默越权）。Task 7 实施中发现并规避，且已用受限角色实测确认生效。
- 2026-09-26 Task 3 完成：`PageUtils.buildPage()` / `LambdaQueryWrapperX` / `getDataTable(IPage)` 就位，10 个不连库单测全绿（测试基建同 Task 新配）。
- 2026-09-26 Task 4 完成：岗位模块全链路切 BaseMapper（分页/CRUD/唯一性校验），冷启动端到端 8 步全绿，SQL 日志确认走 MP 注入模板，测试数据与会话已清理。
- 2026-09-26 Task 5 完成：字典/参数三模块全链路切 BaseMapper + Wrapper（含 betweenIfPresent 时间区间、字典改名级联 updateDictDataType(update+wrapper)），端到端 17 步全绿，DB 实查 0 残留。
- 2026-09-26 Task 6 完成：日志两模块按坑 6 模式落 IPage 首参 + XML 分页（SQL 逻辑零改动仅参数前缀化），端到端 7 步全绿。
- 2026-09-26 Task 7 完成：notice/job/gen 按范围决策"仅切分页、CRUD 留 XML"；SysJobLog 坑 6 模式；user/role 4 处 IPage+DataScope。端到端 11 项 + 数据权限受限角色专测全绿，测试数据全部清理。

## 附录 A：PageHelper 使用点全清单（改造自查表，16 处调用 / 12 controller 文件 + 1 模板 + 3 支撑类）

> 调用点行号为 2026-09-25 逐一实查。user/role 4 处（#1、#2）走 spec 3.4 的 IPage 首参模式，其余走 selectXxxPage + Wrapper 模式。

| # | 文件 | 位置 | 改造动作 |
|---|------|------|---------|
| 1 | [SysUserController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysUserController.java) | :63 list | 调用点改造 |
| 2 | [SysRoleController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysRoleController.java) | :60 list、:191 allocatedList、:203 unallocatedList | 3 处调用点改造 |
| 3 | [SysConfigController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysConfigController.java) | :44 list | 调用点改造 |
| 4 | [SysDictDataController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysDictDataController.java) | :47 list | 调用点改造 |
| 5 | [SysDictTypeController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysDictTypeController.java) | :41 list | 调用点改造 |
| 6 | [SysPostController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysPostController.java) | :44 list | 调用点改造 |
| 7 | [SysNoticeController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/system/SysNoticeController.java) | :48 list、:134 readUsersList | 2 处调用点改造 |
| 8 | [SysOperlogController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/monitor/SysOperlogController.java) | :38 list | 调用点改造 |
| 9 | [SysLogininforController.java](../../ruoyi-admin/src/main/java/com/ruoyi/web/controller/monitor/SysLogininforController.java) | :42 list | 调用点改造 |
| 10 | [SysJobController.java](../../ruoyi-quartz/src/main/java/com/ruoyi/quartz/controller/SysJobController.java) | :49 list | 调用点改造 |
| 11 | [SysJobLogController.java](../../ruoyi-quartz/src/main/java/com/ruoyi/quartz/controller/SysJobLogController.java) | :41 list | 调用点改造 |
| 12 | [GenController.java](../../ruoyi-generator/src/main/java/com/ruoyi/generator/controller/GenController.java) | :62 genlist、:91 dataList | 2 处调用点改造 |
| 13 | [controller.java.vm](../../ruoyi-generator/src/main/resources/vm/java/controller.java.vm) | :48 | 模板分页模式更新 |
| 14 | [BaseController.java](../../ruoyi-common/src/main/java/com/ruoyi/common/core/controller/BaseController.java) | :53/:61/:74/:89 | 删三方法、改 getDataTable |
| 15 | [PageUtils.java](../../ruoyi-common/src/main/java/com/ruoyi/common/utils/PageUtils.java) | 全文件 | 重写为 MP 实现 |
| 16 | [ruoyi-common/pom.xml](../../ruoyi-common/pom.xml) + [根 pom.xml](../../pom.xml) | :40 / :65 | 删 pagehelper 依赖 |
