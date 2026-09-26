# Tasks · 0.0.0 MyBatis-Plus 功能增加

> 勾选纪律：做完即勾（含对应验证跑通）；没做的不许勾，行尾注明原因；勾选 = 验收通过。
> 依赖：无。Task 排序即建议实施顺序：新能力先就位（Task 1~3）→ 全部调用点切换（Task 4~7）→ 最后删 PageHelper（Task 8，防中间态编译失败，见 spec 风险 R2）。
> 环境约束：启动验证前先停掉 8080 上的 Python/Go 版后端（端口互斥）；冷启动（杀干净残留 java 进程）；MySQL/Redis 为本地 Docker（127.0.0.1，ry-vue / db0）。

## 阶段一：MP 就位（不破坏现状）

## Task 1: 依赖引入
- [x] 根 `pom.xml`：properties 加 `<mybatis-plus.version>3.5.17</mybatis-plus.version>`；dependencyManagement 加 `mybatis-plus-spring-boot4-starter` 与 `mybatis-plus-jsqlparser`
- [x] `ruoyi-common/pom.xml`：加两个 MP 依赖（此 Task 只加不减，PageHelper 在 Task 8 才删）
- [x] `mvn clean package -Dmaven.test.skip=true` 编译通过（`mvn dependency:tree` 复核：MP 3.5.17 全家桶就位，mybatis:3.5.19 由 mybatis-spring-boot-starter 4.1.0 统一提供无版本漂移）
- [x] 启动 + 登录 + 用户列表分页冒烟——2026-09-26 与 Task 2 合并执行：启动成功（banner+8080 监听+MP 上下文注册），登录/用户/岗位/字典分页接口行为不变，boot 日志仅 4 条我方登录脚手架误操作的 CaptchaException（与改造无关）

## Task 2: MyBatisConfig 改造（spec 3.2，风险 R1）
- [x] `MyBatisConfig.sqlSessionFactory`：`SqlSessionFactoryBean` → `MybatisSqlSessionFactoryBean`；typeAliases/mapperLocations 解析逻辑不动；`setConfigLocation` 沿用 mybatis-config.xml（**坑：MP 3.5.17 该类在 `com.baomidou.mybatisplus.spring`，不在 extension.spring，见实施记录坑 1**）
- [x] 注册 `MybatisPlusInterceptor` + `PaginationInnerInterceptor(DbType.MYSQL)`
- [x] GlobalConfig 逻辑删除全局值：`logicDeleteValue="2"` / `logicNotDeleteValue="0"`
- [x] 验证：MP 生效判据——启动后 XML 查询正常（sys_job 查到 3 条）+ `MybatisPlusApplicationContextAware` 注册日志出现；BaseMapper 直调判据随 Task 4 试点落地（生产代码首条 BaseMapper 路径即判据）
- [x] 验证：存量 XML 查询（用户列表 join dept/roles）不受影响（启动时 quartz/字典等 XML 查询全部正常，用户列表接口冒烟通过）

## Task 3: 分页支撑类（spec 3.3）
- [x] `PageUtils` 重写为 MP 实现：`buildPage()` 从 `TableSupport.buildPageRequest()` 解析参数，组装 `Page<T>`（orderBy → `OrderItem`，`SqlUtil.escapeOrderBySql` 转义保留——R9 注入防护）。**旧 `startPage()/clearPage()` 暂留文件内**（16 处调用点未迁完前删了必编译失败，Task 8 删）。**开发实查修正：MP `Page` 无 `setReasonable`（立项预判有误，见 spec 3.3 reasonable 说明）**
- [x] `LambdaQueryWrapperX` 新建（放 `com.ruoyi.common.core.mybatis`）：`eq/ne/gt/ge/lt/le/like/in/between` IfPresent 族（between 半开区间退化）+ 链式返回类型重写防断链
- [x] `BaseController` 新增 `getDataTable(IPage<?>)` 重载（rows=records、total=getTotal；MP Page 不是 List，必须重载不能复用 List 签名——spec 3.3）；原 `getDataTable(List<?>)` 本 Task 不动
- [x] 单元测试（不连库）10 个全绿：`PageUtils.buildPage` 参数解析（默认值/驼峰转下划线/ascending-descending/注入拒绝）；`getDataTable(IPage)` total/rows 口径并入端到端验证（Task 4 冒烟）；`LambdaQueryWrapperX` IfPresent 生成与跳过、between 半开退化、链式类型、IN 空集合短路。**测试基建为新配**：surefire 2.12.4（2012 年默认版）不识别 JUnit 5，根 pom 升 3.5.3；ruoyi-common 加 junit-jupiter 5.12.2（test scope）；测试用 JDK 代理 stub 代替 MockHttpServletRequest（spring-test 只有 6.x，避免与 Spring 7 错位）
- [x] （此时 PageHelper 仍在位，旧 `startPage()` 链路不受影响——编译通过确认）

## 阶段二：调用点切换（按模块，每个 Task 内 mapper/service/controller 一起改）

> 每个 Task 完成标准：该模块前端页面列表查询正常（翻页/跳页/筛选/排序），`startPage()` 已从该 controller 删除。
> service 层签名约定（全模块统一）：`List<Xxx> selectXxxList(Xxx xxx)` 保留给非分页/导出场景，新增 `Page<Xxx> selectXxxPage(Xxx xxx)` 供分页列表——**对外接口契约不变的前提下，service 接口（ISysXxxService）加方法不加不改**。

## Task 4: 岗位模块试点（post，最简单单表，模式样板）
- [x] `SysPost` 实体 MP 注解：`@TableName`/`@TableId(type=AUTO)`/`flag` 加 `@TableField(exist=false)`；`BaseEntity` 的 `searchValue`/`params` 加 `@TableField(exist = false)`（共用基类一次改完全局受益）
- [x] `SysPostMapper extends BaseMapper<SysPost>` + `selectPostPage`（Mapper default 方法，LambdaQueryWrapperX 组装）；**XML 单表 CRUD 已删**（selectPostById/insertPost/updatePost/deletePostById/deletePostByIds/checkXxxUnique ×2），service 改调 BaseMapper 等价方法（selectById/insert/updateById/deleteById/deleteBatchIds）；保留 XML：selectPostList（导出）/selectPostAll/selectPostListByUserId/selectPostsByUserName
- [x] `BaseService` 新建（com.ruoyi.common.core.service）+ `checkUnique(keyField, keyValue, field, value)`；`SysPostServiceImpl extends BaseService<SysPostMapper, SysPost>`，两个 checkUnique 切换实现（sys_post 无 del_flag，R5 不适用可安全迁）
- [x] `SysPostController.list` 切新分页：`return getDataTable(postService.selectPostPage(post));`
- [x] 端到端 8 步全绿（2026-09-26 冷启动实测）：新增（主键回填+createTime 由 service 显式赋值兜底）/读回/编辑（判空更新）/重名拦截（checkUnique→500 提示）/删除/getInfo(selectById 驼峰映射)/列表分页(total=4 真实 count)/排序筛选——**SQL 日志确认全部走 MP 注入模板**（INSERT INTO/UPDATE...WHERE/DELETE...IN/SELECT COUNT），测试数据已清理、会话已清
- [x] SQL 日志抽查：BaseMapper 生成语句与旧 XML 语义一致（R4/R6）；**XML updatePost 原对 `update_by != null and != ''` 才更新、updateById 判空策略是仅 null 跳过——空串会进 SET 子句写空值，实际行为差异登记为已知（controller 恒 set updateBy=username，前端全量提交场景不受影响）**
- [x] 实施发现（补记 spec 实施记录）：① mapper 方法 ID 与 XML 语句同存时 SQL 走 XML（Task 4 第一次验证 insert 未走 MP 即此因，删 XML 后才真正切 BaseMapper）；② XML 的 `sysdate()` 兜底删除后由 service 显式赋值 createTime/updateTime 等价替代（RuoYi controller 惯例只 set createBy/updateBy 不 set 时间）；③ selectById 走 MP 驼峰映射，SysPostResult resultMap 为纯列映射无 association，等价成立

## Task 5: 字典与参数模块（dict_type / dict_data / config）
- [x] 实体注解：SysDictType、SysDictData、SysConfig（@TableName/@TableId；三实体均为纯表字段无关联对象）
- [x] 三个 Mapper extends BaseMapper；**XML 清空**（语句全部改由 BaseMapper + Mapper default 方法提供：selectXxxPage 分页 / selectConfig / selectDictTypeByType / selectDictDataByType / selectDictLabel / countDictDataByType / updateDictDataType(update+wrapper) / selectConfigList）；checkXxxUnique 切 BaseService.checkUnique（三表均无 del_flag，R5 不适用）
- [x] 三个 ServiceImpl extends BaseService；insert/update 补 createTime/updateTime 显式赋值（坑 5 模式）；selectXxxById → selectById（resultMap 纯列映射，MP 驼峰等价）
- [x] 三个 controller 切新分页（selectXxxPage + getDataTable(IPage)）
- [x] 端到端 17 步全绿（2026-09-26 冷启动实测）：dict_type 分页/名称筛选/时间区间筛选（betweenIfPresent 带 params[beginTime/endTime]）、dict_data 分页（dict_sort 升序+状态筛选）、config 分页+时间区间、字典类型增改（含改名级联 updateDictDataType 同步 dict_data）、重名拦截 500、字典数据增查删、参数增改删+缓存联动（configKey 回源正常，值在 msg 字段系原版行为）、唯一性校验走 selectCount——SQL 日志确认全 MP 模板，测试数据清理（DB 实查 0 残留）、会话已清、日志无异常

## Task 6: 日志模块（oper_log / logininfor）
- [x] 实体注解：**不迁**（坑 6：sys_oper_log/sys_logininfor 无审计列，BaseEntity 继承字段无法按表排除 exist，实体保持原样）；`businessTypes` 数组查询辅助字段在 XML 路径下无需注解
- [x] **两 Mapper 不继承 BaseMapper，改走 spec 3.4 IPage 首参 + XML 分页模式**：`selectOperLogPage(IPage, @Param("operLog"))` / `selectLogininforPage(IPage, @Param("logininfor"))` 新增 XML 语句（参数引用前缀化，SQL 逻辑同原版）；insert/clean/deleteByIds/详情 保留 XML 原句（sysdate() 兜底原样保留）
- [x] 两 ServiceImpl 恢复直调 Mapper（extends BaseService **回退**——BaseService 泛型注入要求 BaseMapper，本模块不适用），selectXxxPage 走 `PageUtils.buildPage()` + IPage 首参
- [x] 两 controller 切新分页（getDataTable(IPage)）
- [x] 端到端 7 步全绿（2026-09-26 冷启动实测）：operlog 分页 oper_id 倒序 / title+时间区间筛选 / businessType 筛选；logininfor 分页 info_id 倒序 / userName+status 筛选 / 时间区间筛选 / 删除——测试脚本对"详情接口"的断言系误设（该路径只挂 DELETE，前端详情弹窗用本地行数据不调后端），非缺陷；时间区间 `endTime=当日` 不含当日数据系原版 XML `<=` 拼日期语义，行为与迁移前一致

## Task 7: 剩余分页模块（notice + quartz + generator + user/role 分页切换）
- [x] 实体注解：SysNotice（isRead exist=false）、SysJob、GenTable（columns/options exist=false）、GenTableColumn；**SysJobLog 不注解不迁 BaseMapper**（表只有 create_time 无其余审计列，坑 6 同型）；SysUser/SysRole 不做实体注解（IPage 首参模式无需）
- [x] **范围决策（对 spec 3.4 的执行口径收窄，已记实施记录）**：notice/job/gen 四个模块 CRUD 留 XML（notice_content 的 cast() 投影、gen 的 information_schema 查询等 Wrapper 化收益低风险高），**仅分页切换**——XML 加 selectXxxPage 语句（原语句参数前缀化 @Param 命名引用 + 复制为 Page 版），mapper/service/controller 三层加 Page 方法。SysJobLog 同模式
- [x] **user/role 4 处 join+DataScope 查询切分页**：`selectUserList/selectAllocatedList/selectUnallocatedList/selectRoleList` 前缀化 + Page 版语句；**service Page 方法必须带与 List 版相同的 @DataScope 注解**（否则切面不注入 params.dataScope → Page 版查全库越权——实施中发现并规避，checklist 数据权限专条验证通过）
- [x] `SysDeptMapper.selectDeptList` 第 5 处 dataScope 注入无分页（树形接口），本期不动（符合计划）
- [x] 端到端 11 项 + 数据权限专测全绿（2026-09-26 冷启动实测）：notice 分页/类型筛选、job 分页、jobLog 分页、gen 业务表分页 + db 表分页（information_schema 查询正常）、user/role 分页、user 筛选+排序、allocated/unallocated 分页；**数据权限实测**：造 deptId=100 + 角色2（dataScope=2）测试用户登录，user list 仅见本部门 2 人（admin 的研发部 103 不可见），证明 `${user.params.dataScope}` 在 Page 版语句正确生效——测试用户已删、DB 复核无残留、日志零异常
- [x] 实施坑（已记坑 7）：XML 参数前缀化必须连 `<if test>` 里**比较用的裸属性名**一起改（`test="user.userName != null and user.userName != ''"`），漏改后半截运行期才报 Parameter not found；且前缀化脚本对同语句重复执行会叠加前缀（genTable.genTable...），XML 修改前先 git 还原再手工改

## 阶段三：收尾

## Task 8: 删除 PageHelper（最后做，防中间态编译失败）
- [x] 全局 grep `com.github.pagehelper` 确认为零（.java/.xml/.vm/pom 全查，0 命中）
- [x] 附录 A 逐行核对：16 处调用点均已切走（**实施中揪出 5 处漏网**——user.list/role.list/gen 两处此前脚本锚点静默未中、notice readUsersList 需补 IPage 三层链路，均已补齐后 grep 复核为 0）
- [x] `ruoyi-common/pom.xml` 删 pagehelper-spring-boot-starter；根 pom 删 dependencyManagement 条目与 `pagehelper.boot.version` 属性
- [x] `BaseController` 删 `startPage()/startOrderBy()/clearPage()` 与 pagehelper import、无用 import；`getDataTable(List<?>)` 内部 `new PageInfo(list).getTotal()` 改为 `list.size()`（语义：PageHelper 删除后 List 版仅剩非分页场景）；`PageUtils` 重写为纯 MP 版（旧 startPage/clearPage 已删）
- [x] 连锁修复（预期内）：`SpringBootVFS` 原由 PageHelper 传递的 mybatis-spring-boot-autoconfigure 提供，删除后换用 MP 自带 `com.baomidou.mybatisplus.autoconfigure.SpringBootVFS`
- [x] `mvn clean package` + `mvn test`（common 10 单测）+ 全量前端核心页面回归 19 项（登录/post/user/role/menu/dept/dict×2/config/notice+readUsers/operlog/logininfor/online/job/joblog/gen×2/allocated/unallocated）+ post CRUD 往复——全绿，日志零异常，测试数据清理
- [x] **mybatis-spring-boot-starter 依赖形态复核**：MP starter 传递提供 `mybatis-spring:4.0.0`，根 pom dependencyManagement 保留显式声明（版本对齐 mybatis:3.5.19 无漂移）

## Task 9: 文档同步
- [x] 本 spec.md「实施记录」补齐（坑 1~8 + 各 Task 结论）；坑 1/3/6/7 已具普适性，待下次提交时评估是否上移根 AGENTS.md《已踩过的坑》（Java 版专属坑先留本 spec）
- [x] Python 版 `specs/README.md` 技术对位表 MyBatis 行补"Java 侧 2026-09-26 已切 MyBatis-Plus"备注；GO 版 `specs/tech-stack.md` ORM 行同步
- [x] 工作区 `readme.md` ORM 行（MyBatis → MyBatis-Plus）+ 目录注释同步；根 `AGENTS.md` 工作区结构行同步
- [x] `RuoYi-Vue/specs/README.md` 模块总表本模块标 ✅ + 日期（2026-09-26，checklist 总验收通过后执行）
