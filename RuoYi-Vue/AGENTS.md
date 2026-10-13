# AGENTS.md（RuoYi-Vue）

本文件是 Java 版子目录的 AI 编码规范入口。工作区级规范（兼容契约、spec checklist 纪律、编码行为纪律、共用库纪律）以根目录 [../AGENTS.md](../AGENTS.md) 为唯一规范源，此处不重复；Java 项目特有的编码约定与 MP 使用规则维护在本文件。

## 项目定位

用 Spring Boot 4 / Java 17 / MyBatis-Plus 实现的服务端，是三版（Java/Python/Go）接口契约的**唯一基准**。演进任务台账在 [specs/README.md](specs/README.md)（三件套：spec.md / tasks.md / checklist.md，编号从 0.0.0 起独立递增）。

## Java 编码约定

1. **接口方法不加 `public` 修饰符**：接口方法隐式 public，写出来是冗余样板。全项目 Mapper 接口（`ruoyi-*/src/main/java/**/mapper/*.java`）已按此规则执行；新增接口（含 Service 接口）一律遵循，改动既有接口时顺带按此规则收敛。
2. **MyBatis-Plus 使用约定**（0.0.0 spec 沉淀，细则与实施坑见 [specs/0.0.0-MyBatis-Plus功能增加/spec.md](specs/0.0.0-MyBatis-Plus功能增加/spec.md) 实施记录）：
   - 单表 CRUD 优先 BaseMapper + `LambdaQueryWrapperX`（IfPresent 条件族），动态条件组装写在 Mapper 接口的 default 方法里，列名一律方法引用；
   - join / `${params.dataScope}` 查询不迁 Wrapper：走 **IPage 首参 + @Param 命名引用 + XML** 分页模式，XML 参数引用与 `<if test>` 比较式同步加前缀；
   - **无审计列的表**（sys_oper_log/sys_logininfor/sys_job_log 等，实体继承 BaseEntity 但表缺 create_by 等列）**不得迁 BaseMapper**——MP 实体解析含父类字段，会拼出不存在的列；
   - IPage 首参模式的 service Page 方法**必须带与 List 版相同的 `@DataScope`**，漏注解即静默全库越权；
   - XML insert/update 的 `sysdate()` 兜底删除后，service 显式赋值 createTime/updateTime。
3. **数据范围过滤双路径**（1.0.0 spec）：XML 查询走 `@DataScope` 切面 → `params.dataScope`（既有机制）；**纯 BaseMapper/Wrapper 查询走** MP `DataPermissionInterceptor`——service 标 `@DataScope` → 切面入栈范围 DTO → `DeptDataPermissionRule` 白名单命中才拼条件。带范围语义的表接 BaseMapper 前**必须先在 `DataPermissionConfiguration` 白名单注册**（注册即生效，改动走 checklist 受限角色实测）；豁免用 `@DataScope(enable=false)` 或 `DataPermissionUtils.executeIgnore`。
3. **分页请求解析**：一律 `PageUtils.buildPage()`（内部 TableSupport 解析 + SqlUtil 注入转义），禁止在新代码里直连 MP `Page` 构造器绕过转义。

## 构建与验证

```bash
mvn clean package -Dmaven.test.skip=true        # 全模块编译（JDK 17，本机 java 全路径见根 AGENTS.local.md）
mvn test -pl ruoyi-common                        # 单元测试（surefire 3.5.3 + JUnit 5）
```

启动验证遵循根 AGENTS.md 环境约束：8080 端口互斥（三版同时只能跑一个）、共用库测试数据测完清理、编译规则见下节《编译纪律》（唯一判据：服务是否在运行）。

## 编译纪律（必须遵守）

**唯一判据：服务端是否在运行。与任务性质无关**——spec 任务、临时改动、探索验证一视同仁（2026-10-08 重写；此前的"按任务性质分流 A/B"方案在 IDEA 调试场景实测出坑，废除）。

### 检测：服务是否在运行（动 mvn 前必查）

任一命中即视为"服务运行中"：

1. **8080 端口被监听**：`netstat -ano | findstr :8080` 后 `tasklist /fi "PID eq <pid>"` 确认是 java.exe（不区分谁启动的——IDEA 调试的、agent 起的、jar 残留的，一律算运行中）；
2. **IDEA 正打开本项目**：`idea64.exe` 进程存在且 `.idea/` 属于本项目（此时即使服务没起，也可能随时被用户点起来）。

> 判定拿不准时（如 8080 空闲但 IDEA 进程在）：按"运行中"处理——宁可少编译，不可干扰开发者。

### 服务运行中：不编译、不启停

- **不执行任何 mvn 编译类命令**（compile / package / clean / install / repackage 全部禁止）——无论什么任务；
- 原因：IDEA 跑的服务 classpath 就是本项目 `target/classes`，agent 一编译，devtools 会感知并热重启服务，等于暗中改写开发者正在调试的运行时；`clean/package` 还会删除/改写 `ruoyi-admin.jar`（2026-10-08 实测踩坑：一次 `mvn compile` 热重启了用户的 IDEA debug 会话）；
- 不启停服务：不 `taskkill`、不启动新实例（8080 也起不来）；
- 此时代码改动的做法：javac 语法检查——
  ```bash
  # classpath 清单首次生成（只读依赖树不产编译产物；只能用 ruoyi-common 生成——
  # 其他模块依赖 workspace 内模块、本地仓库无 install 过的 jar，会解析失败）：
  mvn dependency:build-classpath -Dmdep.outputFile=%TEMP%\ruoyi-cp.txt -q -pl ruoyi-common
  # 对改动的 .java 做语法检查（-d 指向临时目录；模块间依赖用 IDEA 编译好的
  # target/classes——服务在跑时它们就是最新产物）：
  javac -proc:none -encoding UTF-8 -d "%TEMP%\ruoyi-syntax-check" \
    -cp "$(cat %TEMP%/ruoyi-cp.txt);ruoyi-common/target/classes;ruoyi-system/target/classes;ruoyi-framework/target/classes;ruoyi-quartz/target/classes;ruoyi-generator/target/classes" \
    <改动的 .java 文件...>
  ```
  javac 报语法/符号错误当场修掉；需要运行时才能验证的行为，在回复里写清"改了什么、哪些要等用户重新 Build / 重启后验证"——**编译与重启由用户自己完成**。前端/Python/Go 的改动不受此限（不碰 Java 产物）。

### 服务未运行：可自主编译、启动、验证

- 编译：`mvn clean package -Dmaven.test.skip=true`（或按需 `compile`）；
- 启动验证：`"C:\Program Files\Java\jdk-17.0.12\bin\java.exe" -jar ruoyi-admin/target/ruoyi-admin.jar` 冷启动（启动前确认 8080 仍空闲）；
- 端到端验证完**必须停掉服务**（按 PID 精确 `taskkill //F //PID <pid>`，禁止 `//IM java.exe` 批杀），不留后台进程占用 8080 干扰开发者下一次 IDEA 启动；
- 已知坑：服务在跑时执行 `mvn clean`（删 jar 失败）与 `repackage`（rename 失败）都会报错——这正是"运行中禁止编译"的原因之一。
