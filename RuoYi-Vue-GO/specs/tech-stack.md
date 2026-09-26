# RuoYi-Vue-GO 技术选型调研

> 日期：2026-09-24　|　性质：**纯调研，不含开发**
> 目标：与 RuoYi-Vue-FastApi 同一定位——用 Go 复刻 Java 版 RuoYi 接口，前端 RuoYi-Vue3 零改动切换。
> 选型倾向：**国内流行、社区认知度高**的框架优先（兼顾简历价值）；文中 star 数为 2026-09-24 通过 GitHub API 实查。

---

## 一、结论一览（推荐组合）

| 层次 | Java 版对位 | Python 版对位 | **Go 推荐** | 备选 |
|---|---|---|---|---|
| Web 框架 | Spring Boot / Spring MVC | FastAPI | **Gin**（89.3k★） | GoFrame（13.3k★） |
| ORM | MyBatis / MyBatis-Plus | SQLAlchemy 2.0 async | **GORM**（40k★） | Ent（17.2k★）、sqlx |

> 注：Java 基准 2026-09-26 已引入 MyBatis-Plus 3.5.17 并删除 PageHelper（见 Java 版 specs/0.0.0）；本表 ORM 对位行为历史调研记录，Go 版 GORM 选型不受影响。
| Redis | Spring Data Redis | redis-py + RedisCache 门面 | **go-redis v9**（22.2k★） | redigo |
| JWT | jjwt（HS512） | PyJWT | **golang-jwt/jwt v5**（9.2k★） | — |
| 密码 BCrypt | spring-security-crypto | passlib[bcrypt] | **golang.org/x/crypto/bcrypt** | — |
| 参数校验 | Hibernate Validator / @Validated | pydantic | **validator v10**（20.2k★，gin 内置 binding） | — |
| 配置 | application.yml + .env.dev | pydantic-settings + .env.dev | **viper**（30.5k★） | GoFrame 自带 |
| 日志 | Logback（logback.xml） | loguru | **zap**（24.7k★）+ lumberjack 轮转 | 标准库 slog |
| Excel 导入导出 | Apache POI（ExcelUtil） | pandas/openpyxl | **excelize**（20.9k★，阿里系维护） | — |
| 验证码 | Kaptcha（Google 风格 math） | 自绘 PIL/自定义 | **base64Captcha**（约 4k★） | — |
| API 文档 | springfox（/v3/api-docs） | FastAPI 自带 /docs | **swaggo/swag**（13k★）+ gin-swagger | — |
| 定时任务（sys_job） | Quartz + ScheduleUtils | APScheduler | **go-co-op/gocron**（约 6k★）或 robfig/cron（14.2k★） | asynq（约 12k★，分布式） |
| 限流（可选） | Sentinel（Spring Cloud Alibaba） | — | **sentinel-golang**（约 8.5k★，阿里开源） | x/time/rate |
| 数据库迁移 | 共用 Java 版库表，无迁移 | 同左 | **不需要**（同库同表，沿用 ry_xxx.sql） | goose（11.5k★）/ golang-migrate（18.9k★） |

一句话：**Gin + GORM + go-redis + golang-jwt + bcrypt + viper + zap** 的"国内事实标准组合"，其余按需加。

---

## 二、Web 框架对比（本次最关键决策）

| 框架 | star | 出身 | 维护 | 适配本项目 | 备注 |
|---|---|---|---|---|---|
| **Gin** | 89.3k | 社区 | 活跃（2026-09 有提交） | ★★★★★ | 国内事实标准，生态最全（jwt/casbin/cors/swag 均有现成中间件），go-admin、gin-vue-admin 等国内后台全是 Gin+GORM。与 FastApi 版"轻框架+自由分层"思路同构 |
| GoFrame (gf) | 13.3k | 国人主导 | 活跃（2026-09-24 有提交） | ★★★★ | 国产全家桶：自带 ORM/缓存/日志/校验/定时，中文文档最好；代价是与主流生态绑定较深，换框架成本高。已有 kshdb/RuoYi-Go-Plus 用它兼容 RuoYi-Vue-Plus |
| go-zero | 33.3k | 好未来 | 活跃 | ★★★ | 微服务全家桶（goctl 代码生成），单体复刻场景偏重；若未来拆微服务再引入 |
| kratos | 25.9k | B 站 | 活跃 | ★★★ | 微服务 + DDD 脚手架，gRPC 优先，HTTP 复刻 RuoYi 不顺手 |
| beego | 32.4k | 国内老牌 | 维护中 | ★★ | 巅峰已过，新项目选它的人少了 |
| Hertz | 约 5k+ | 字节 CloudWeGo | 活跃 | ★★ | 性能好，但国内后台管理场景生态弱于 Gin |

**推荐 Gin**。核心理由：① 面试/简历认知度最高（"你 Go 用什么？"答 Gin+GORM 最稳）；② 中间件生态直接覆盖本项目全部需求；③ 与 FastApi 版同为"薄框架 + 自主分层"，架构经验可直接迁移。

## 三、分层映射与目录结构（含主流 Go 项目调研）

### 3.1 分层映射（对齐 FastApi 版分层纪律）

CLAUDE.md 的 controller → service → dao 分层纪律在 Go 版延续，只换顶层布局、不换分层语义：

| FastApi 版 | Go 版（Gin） | 说明 |
|---|---|---|
| controller + pydantic vo | `internal/module/admin/handler/` + `model/vo`（struct + json tag） | handler 只做参数绑定、调 service、组响应 |
| service | `internal/module/admin/service/` | 业务逻辑、事务边界、缓存维护 |
| dao（SQLAlchemy） | `internal/module/admin/dao/`（GORM） | 仅查库写库 |
| entity/do | `model/do`（GORM Model struct） | 对位 Java domain |
| entity/vo | `model/vo`（DTO struct） | **json tag 必须显式驼峰** |
| ResponseUtil（AjaxResult 对位） | `pkg/response` 包 | `{code, msg, ...}` 信封、code 200/500/601/403/401 全部 HTTP 200 |
| RedisCache 门面 | `internal/cache` 封装 go-redis | key 前缀、SCAN 替代 KEYS、JSON 序列化统一收口 |
| exceptions/handle.py 全局异常 | `internal/middleware` Recovery + 统一错误中间件 | panic 与业务错误统一转 `{code:500}` 信封 |
| server.py controller_list | `internal/router` 注册中心 | 路由集中注册 |
| module_admin/aspect | `internal/module/admin/aspect/` | 数据权限等切面 |

### 3.2 目录结构调研（GitHub/Gitee 高 star Go Web 项目，2026-09-24）

**风格 A · 按层分包（gin-vue-admin 25k★，国内最流行后台脚手架）**

```
server/
├── api/v1/        # handler（按版本分）
├── config/        # 配置结构定义
├── core/          # 启动逻辑
├── global/        # 全局变量（GVA_DB / GVA_REDIS / GVA_CONFIG）
├── initialize/    # 初始化装配（router/db/redis）
├── middleware/
├── model/         # request / response 分开
├── router/
├── service/
└── utils/
```

- 优点：扁平直观、入门零门槛；handler/service 分层心智与 RuoYi 完全一致；生态验证最充分（go-admin 同风格）
- 缺点：顶层包多、无编译期私有边界；global/ 全局变量风格在 Go 社区有争议（隐式依赖）；业务域内聚弱（service 下 N 个业务混一个包）

**风格 B · 业务域 + internal 隔离（kratos 25.9k★，B 站 DDD 脚手架）**

```
cmd/server/main.go
internal/
├── biz/       # 业务逻辑 + repo 接口定义（用例层）
├── data/      # repo 接口实现（数据访问层）
├── service/   # 对外服务编排（DTO 组装）
├── server/    # http/gRPC server 装配
└── conf/      # 配置结构
```

- 优点：internal/ 编译期禁止外部工程引用；biz 定义接口、data 实现（依赖倒置）面向领域而非表；演进微服务最顺
- 缺点：DDD 味重；每个实体都要 biz 接口 + data 实现 + wire 依赖注入，单人复刻项目是显著额外负担；分层名（biz/data/service）与 Java/Python 版（service/dao）心智不对应

**风格 C · 骨架惯例（golang-standards/project-layout，社区事实指南，非官方标准）**

- 不是完整结构，而是两条原则：`cmd/` 放 main 入口；`internal/` 放私有代码（编译器强制外部不可 import），可对外公开的通用库才放 `pkg/`
- 该原则与 A、B 不冲突——gin-vue-admin/kratos 均在此骨架内组织

**国内 RuoYi-Go 参考品**（第五节详表）：lostvip-com/ruoyi-go 等结构简单，参考价值有限，不构成第四种风格。

### 3.3 选型结论：project-layout 骨架 + 按层分包 + 业务域收拢（混合式）

```
RuoYi-Vue-GO/
├── cmd/server/main.go        # 入口
├── configs/                  # .env.dev / .env.prod
├── internal/                 # 私有代码（编译期禁止外部引用）
│   ├── config/               # viper 读取
│   ├── common/               # constant / enums / message / errors
│   ├── cache/                # RedisCache 门面
│   ├── middleware/
│   ├── router/
│   └── module/admin/         # 业务域：handler / service / dao / aspect / model{do,vo}
└── pkg/                      # 通用工具（无业务依赖）：response / logger / utils
```

决策依据（对照本项目三个诉求）：
1. **前端零改动复刻优先** → 分层纪律直接延续 Python 版（handler→service→dao），不引入 Kratos 的 biz/data 接口倒置——复刻的敌人是复杂度，DDD 抽象对单人项目是负资产。
2. **Go 工程素养叙事** → 骨架采用 cmd/ + internal/ + pkg/（project-layout 惯例），internal/ 私有边界、pkg/ 通用工具，面试可讲"按 Go 社区惯例组织而非照搬 Python 布局"；比 gin-vue-admin 的扁平顶层 + global/ 全局变量更显工程规范。
3. **经验迁移** → module/admin 业务域包对位 Python module_admin，代码结构可互相索引；config/cache/middleware/router 与 Python 版一一对应，遇到问题先查 Python 版同位文件。

与 Python 版的差异（有意为之，登记 deviations 不需要——属选型期决策）：
- 顶层套 internal/（Python 无此机制，Go 特有编译期边界）
- 纯工具上移 pkg/（response/logger/utils 无业务依赖）；带业务语义的 cache（key 白名单）留 internal/
- 单元测试与源码同包（xxx_test.go），不设顶层 tests/ 目录——Go 惯例
- global/ 全局变量不采用：db/redis 句柄显式注入（结构体字段），依赖可追溯

---

## 四、兼容性风险点（沿用 FastApi 版契约，Go 特有的坑）

与 Java 版同库同 Redis（MySQL ry-vue-26-09-24 + Redis db11），以下契约必须逐条对齐：

1. **BCrypt 互验**：`x/crypto/bcrypt` 默认生成 `$2a$` 前缀、cost 10，与 Java `BCryptPasswordEncoder` 互验 ✅（低成本风险）。
2. **JWT HS512**：HMAC 算法通用，golang-jwt 与 jjwt 用同一 secret 可互相解析；claims 里的 `login_user_key` 等 key 按原样读写即可。风险低，动手时先做互验冒烟。
3. **Redis 会话 value 第三种格式**：Java 是 FastJson 带 @type，Python 是纯 JSON，Go 又是一种——沿用"已知设计，切换后需重新登录"，不是 bug。
4. **键前缀**：`login_tokens:` / `captcha_codes:` / `pwd_err_cnt:` / `sys_config:`，常量集中在 `internal/common/constant`（对位 Python 版 `common/constant.py`）。
5. **JSON 驼峰**：Go struct 零值导出字段默认原样输出，**所有 vo 的 json tag 必须手写驼峰**，建议 lint 卡点。
6. **时间格式（隐藏大坑）**：Java Jackson 全局序列化 `yyyy-MM-dd HH:mm:ss`；Go `time.Time` 默认输出 RFC3339（带 T 和时区）。需为 do/vo 统一自定义时间类型（或统一 `Format` 后存 string），否则前端日期显示全乱。FastApi 版踩过同一坑，方案可复用。
7. **bigint 主键**：Java 直接序列化数字；Go int64 同样直接输出数字即可，不要学 Web 习惯转 string，保持与 Java 一致。
8. **端口互斥**：Go 版同监听 8080，三版本（Java/Python/Go）同时只能跑一个。
9. **测试数据清理**：共用库纪律不变，端到端测试后清理数据、复原 admin 会话。

---

## 五、国内现成 RuoYi-Go 参考项目（GitHub 实查）

无成熟移植品（最高仅 267★），**自研复刻仍是主线**，这些只作参考：

| 项目 | star | 技术栈 | 参考价值 |
|---|---|---|---|
| lostvip-com/ruoyi-go | 267 | Gin + GORM，模板引擎版 | SQL 与代码分离（模仿 MyBatis XML）的写法 |
| Kun-GitHub/RuoYi-Go | 118 | Iris + GORM，DDD 六边形 | 分层组织方式 |
| fivepmcoder/ruoyi-go | 47 | Gin + GORM | 脚手架结构 |
| kshdb/RuoYi-Go-Plus | 6 | GoFrame | 兼容 RuoYi-Vue-Plus 生态的思路 |
| atlas-u/micro-go | 9 | kratos | 微服务化演进参考 |

生态参照系：gin-vue-admin（25k★）、go-admin（12.8k★）两个国内主流 Go 后台框架均验证了 Gin+GORM+Casbin+JWT 组合在"RuoYi 类"产品上的可行性。

---

## 六、简历联动（选型兼顾求职叙事）

- 技能行 "Python / Go（AI 辅助快速交付）"：本项目落地后，Go 从"会"升级为"有完整后台系统实战"，且是**同一套系统三语言实现**（Java/Python/Go），是很好的面试故事。
- 选 Gin+GORM 的面试话术：国内生态/认知度、复刻兼容优先于框架花活——与"接口完全兼容、前端零改动"的工程判断一致。
- 简历呼应点（第二阶段可选扩展，本期不做）：
  - **sentinel-golang**（阿里）↔ 简历的 Spring Cloud Alibaba Sentinel 限流熔断；
  - **gnet**（约 13k★，国产作者）+ **paho.mqtt.golang** ↔ 简历的 Netty TCP / Modbus / MQTT 物联网主线——若 GO 版后续接 IoT 需求，"Netty → gnet"是现成的叙事线；
  - **nacos-sdk-go** ↔ 简历的 Nacos 注册配置中心（微服务拆分时）；
  - **emmansun/gmsm 国密 SM2/3/4** ↔ 简历的 SM2 国密、政务信创背景。
- 信创注意：本期共用 MySQL 不涉及国产库；若未来要适配 KingbaseES，Go 有官方驱动（database/sql 接口），选型时 ORM 用 GORM 可走 postgres 方言，保留可行性即可。

---

## 七、分阶段路线建议（本期不开发，仅供后续 spec 排期）

1. **阶段 0 · 工程基础**：go.mod + 目录分层（对齐 3.3 节选型结构）+ viper + zap + ResponseUtil + 统一异常中间件 + 配置读取（.env.dev 与 Java 版同步纪律）。
2. **阶段 1 · 登录链路（对齐 FastApi 版本期范围）**：captcha（base64Captcha）→ login（bcrypt + golang-jwt + Redis 会话）→ getInfo/getRouters → 与 Java 版互验 BCrypt/JWT/Redis 键。
3. **阶段 2 · 系统管理 CRUD**：dept/user/role/menu/config/dict/post/notice/log，GORM 通用分页封装（对位 TableDataInfo）、Excel 导入导出（excelize）。
4. **阶段 3 · 监控与任务**：sys_job 调度器（gocron + 数据库驱动）、在线用户、缓存监控、Druid 监控页按 Java 版实际开放范围决定做不做。
5. **阶段 4+ · 演进项**（可选）：casbin、sentinel-golang、微服务化（go-zero/nacos-sdk-go）、IoT 接入（gnet/MQTT）。

---

## 附： star 数据实查记录（GitHub API，2026-09-24）

gin 89,250 · gorm 39,961 · go-zero 33,349 · beego 32,426 · viper 30,468 · kratos 25,942 · gin-vue-admin 25,045 · zap 24,662 · go-redis 22,246 · excelize 20,936 · validator 20,178 · ent 17,202 · robfig/cron 14,190 · GoFrame 13,279 · swag 13,035 · go-admin 12,786 · goose 11,494 · golang-jwt 9,227（未查到精确值：casbin 约 18k、sentinel-golang 约 8.5k、gocron 约 6k、gnet 约 13k、base64Captcha 约 4k、asynq 约 12k）
