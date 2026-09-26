# ai-dev-lab：用工程纪律驱动 AI，跨四个技术栈交付一套系统

这个仓库的主角不是"复刻了一个若依"——复刻只是载体。它展示的是一件事：**如何用 spec 任务书、checklist 验收纪律和数据实证，驱动 AI（Claude Code）在四个技术栈之间保持一致性，包括 AI 在哪跌倒、靠什么机制抓住的**。全部代码由 AI 编写，人工负责需求界定、纪律制定与最终验收。

载体本身：以 [RuoYi-Vue](https://github.com/yangzongzhuan/RuoYi-Vue)（若依官方 Java 版，Spring Boot）为接口契约基准，用 **Go（Gin + GORM）** 和 **Python（FastAPI + SQLAlchemy 2.0 async）** 从零复刻出接口完全兼容的服务端，共用同一个 [RuoYi-Vue3](https://github.com/yangzongzhuan/RuoYi-Vue3) 前端（Vue 3 + Element Plus），**切换后端前端零适配**。`RuoYi-Vue/`、`RuoYi-Vue3/` 为若依官方仓库的本地副本，充当契约基准与共用前端，2026-09-25 起也纳入增量演进（如平台标识模块对 Java 版的新增）。

## AI 开发工作流：spec 驱动，验收以数据为准

每个模块动工前走同一套流程，全部过程文档在仓库中可查（两个 specs 目录共 **57 份**）：

1. **契约先行**：动手前先通读 Java 基准版对应的 Controller / ServiceImpl / Mapper XML，把端点路径、方法、参数、返回 JSON 结构逐条核实，写成该模块的 spec 任务书——不凭记忆，AI 记忆里的东西必须对着源码核实（为什么，见下面的跌倒案例 1）。
2. **spec 三件套**：`spec.md`（API 契约 + 设计决策）、`tasks.md`（任务分解）、`checklist.md`（验收清单）。跨四端的增量功能用**一份契约主文档**统一管理，各语言版挂自己的实施任务书（范本：[12.0.0 平台标识](RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md)）。
3. **验收以数据为准，不凭 AI 自述**：AI 说"做完了"不算数——要扫数据库实际数据（如 sys_job 预置任务是否覆盖）、抓前端实际调用（api/*.js + 页面内调用）、跑单元测试、浏览器端到端操作。
4. **checklist 纪律**：做完即勾、没做不勾并在行尾注明原因、**宁可留白不可虚勾**。两版 specs 目录现有 **491 项已勾验收记录**，44 项如实留白。

## AI 在哪跌倒，我怎么抓住的

以下案例全部有第一手记录，链接可点开验证。每个案例的完整链条：AI 的错误行为 → 被哪个环节抓住 → 纪律如何演进。

**1. 凭记忆写错数据库列名** —— AI 按 Java 实体类的记忆写 gen_table 的列定义，两处列名是错的（如 `author`，实际是 `function_author`），端到端验证时才暴露。修正后该模块 spec 里留了一句记录："**凭 Java 实体记忆写的列名有两处错，端到端暴露后修正**"。"契约先行、表结构以数据库为准、不凭记忆"由此成为动工前检查单的固定条款。见 [11.0.0-代码生成器/spec.md](RuoYi-Vue-GO/specs/11.0.0-代码生成器/spec.md)。

**2. "状态头写了已完成，勾选框全空"** —— 开发中期人工核对时发现，AI 产出的 checklist 存在"状态头标 ✅ 但验收项全空""勾选与实际实现不符"的虚勾。处置不是返工了事，而是把它变成制度：AGENTS.md 新增 **spec checklist 纪律五条**（宁可留白不可虚勾 / 宣称完成前逐条自查 / 功能核对以数据为准）。按新纪律做扫库二次核查，随即抓出两处此前漏实现的校验（checkUserDataScope、菜单名称唯一），并顺藤摸出一个真 bug：暂停态定时任务启用后 `resume_job` 静默空操作——表现为"暂停的任务永远无法启用"，已修复。见根 [AGENTS.md](AGENTS.md) 纪律节、[spec-09-job.md](RuoYi-Vue-FastApi/specs/spec-09-job.md)。

**3. 解析失败就删缓存键——差点砸掉共享会话** —— 三版后端共用同一个 Redis。AI 初版方案是"Java 写的 FastJson 缓存（带 `@type` 类型头）解析失败就删除回源"——在共享 Redis 下这会直接**杀掉 Java 侧的登录会话**。复审发现后改为保守策略：无法解析就原样返回文本、永不删除键。这条后来写进了 AGENTS.md 的踩坑清单（"严禁解析失败就删除缓存键"）。见 [spec-08-monitor-log.md](RuoYi-Vue-FastApi/specs/spec-08-monitor-log.md)、[redis_cache.py](RuoYi-Vue-FastApi/config/redis_cache.py)。

**4. 异步日志静默丢失，接口却返回 200** —— 退出登录的操作日志落库报 `Data too long`，但 HTTP 响应仍是 200，页面无任何异常——curl 层的端到端完全看不出问题，**浏览器级验收**才抓到。修复之外留了一条排查方法论："同类'异步日志静默丢失'问题先查落库行，别只看接口响应"。见 [2.0.0-登录闭环/spec.md](RuoYi-Vue-GO/specs/2.0.0-登录闭环/spec.md)。

**5. 验证码答案带引号** —— Go 版向 Redis 写值用 `json.Marshal`，验证码答案成了带引号的 JSON 字符串，而 Java/Python 是裸文本——同一份数据三种语言三种形态。端到端脚本第一次登录失败暴露，修正脚本后通过，差异记入 spec。这类"序列化形态差异"正是多语言复刻里最容易漏的坑，见 [12.0.0-平台标识/tasks.md](RuoYi-Vue-GO/specs/12.0.0-平台标识/tasks.md)。

其余踩坑（uvicorn --reload 残留 worker、路由遮蔽 `/{param}` 吞固定路径、日志装饰器预读 body 杀死 multipart 上传等）集中登记在根 [AGENTS.md](AGENTS.md)"已踩过的坑"一节——每条都是复现成本高的坑，修一处、记一条、下一版预埋规避。

## 跨四端一致性怎么维持

- **兼容契约**：三版严守同一套契约——统一 `{code, msg, ...}` 响应信封（全部 HTTP 200，前端按 body code 判断）、返回 JSON 一律驼峰、日期统一 `yyyy-MM-dd HH:mm:ss`、同库同 Redis、Redis 键前缀与 Java `CacheConstants` 逐字一致、BCrypt 存量哈希三版互相可验、JWT HS512。权威定义见 [AGENTS.md](AGENTS.md)。
- **差异必须有意且可查**：无法逐字节对齐 Java 的行为，全部登记进 [deviations.md](RuoYi-Vue-GO/specs/deviations.md)（现有 **20 条**），规则是"**影响前端契约的差异一律不允许**——那说明实现错了，不是差异"。例如：会话 value 序列化格式不同（已知设计，切换后端需重新登录）、代码生成器模板端点有意排除（产出对 Go 项目无价值，前端 404 属有意行为）。
- **能力声明替代语言硬编码**：跨四端增量模块（[12.0.0 平台标识](RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md)）新增 `GET /getPlatformInfo` 返回 `features` 能力开关，前端数据监控 / 服务监控 / 系统接口页按能力降级提示"该功能仅 Java 版提供"，而不是 if-else 判断后端语言——新增语言或开关只需一行改动。三版端到端逐字段 curl 断言记录在 [checklist](RuoYi-Vue-GO/specs/12.0.0-平台标识/checklist.md)。
- **量化证据**：Python 版 142 个路由端点、Go 版 127 条路由注册，接口行为与 Java 版逐条对齐；单元测试 **178 个全绿**（Python 55 + Go 123）；checklist 已勾验收记录 491 项。
- **诚实留白也是纪律的一部分**：Go 版 10 个模块的浏览器级页面验收因开发环境限制尚未逐项确认（curl 层全部通过），checklist 里如实注明"待用户确认"而非默默勾掉；Python 版 ruff 接入标记"遗留未做"。验收记录里看到的每个 ✅ 都对应一次真实执行。

## 项目结构

```
ai-dev-lab（原 ruoyi 工作区）
├── RuoYi-Vue/          Java 版服务端（Spring Boot + MyBatis-Plus）—— 接口契约的唯一基准，2026-09-25 起可按学习需要增量演进
├── RuoYi-Vue3/         前端（Vue 3 + Element Plus + Vite）—— 三个后端共用，同样可增量演进
├── RuoYi-Vue-FastApi/  Python 版服务端（FastAPI + SQLAlchemy 2.0 async）
├── RuoYi-Vue-GO/       Go 版服务端（Gin + GORM）
└── docker/             本地开发环境（MySQL 8.0 + Redis 7.0，开箱即用）
```

> 「前端零改动」指三版后端切换时前端无需适配，不代表前端与 Java 版冻结：2026-09-25 起两者均可按学习需要增量演进，以不破坏三版后端通用性为前提（详见 [AGENTS.md](AGENTS.md)）。

## 三版实现进度

| 模块 | Java（基准） | Go | Python |
|------|:---:|:---:|:---:|
| 登录闭环（验证码/登录/getInfo/getRouters/logout） | ✅ | ✅ | ✅ |
| 个人中心 / 注册 / 通用上传下载 | ✅ | ✅ | ✅ |
| 部门管理 / 岗位管理 | ✅ | ✅ | ✅ |
| 用户管理（含导入导出 / authRole） | ✅ | ✅ | ✅ |
| 角色管理 / 菜单管理 | ✅ | ✅ | ✅ |
| 字典管理 / 参数管理 | ✅ | ✅ | ✅ |
| 通知公告（含本版定制已读功能） | ✅ | ✅ | ✅ |
| 在线用户 / 缓存监控 / 操作与登录日志 | ✅ | ✅ | ✅ |
| 定时任务（对位 Quartz） | ✅ | ✅ | ✅ |
| 代码生成器 | ✅ | 🔶 数据层端点 | ✅ |
| 接口文档（springdoc / swagger 兼容路径） | ✅ | ✅ | ✅ |
| 平台标识 `GET /getPlatformInfo`（跨四端增量） | ✅ | ✅ | ✅ |

Go / Python 版每模块的 API 契约、任务分解与验收记录见各自 specs 目录（[GO](RuoYi-Vue-GO/specs/README.md) / [Python](RuoYi-Vue-FastApi/specs/README.md)），与 Java 版的有意差异集中登记在 [deviations.md](RuoYi-Vue-GO/specs/deviations.md)。

## 技术栈对照

| 组件 | Java 版 | Go 版 | Python 版 |
|------|---------|-------|-----------|
| Web 框架 | Spring Boot | Gin | FastAPI + uvicorn |
| ORM | MyBatis-Plus + Druid | GORM | SQLAlchemy 2.0 (async) |
| 缓存/会话 | Spring Data Redis | go-redis v9 | redis-py (asyncio) |
| JWT | jjwt (HS512) | golang-jwt v5 | python-jose |
| 密码加密 | BCryptPasswordEncoder | x/crypto/bcrypt | passlib（三版哈希互相可验） |
| 验证码 | kaptcha | 标准库 image 自绘 | Pillow |
| 定时任务 | Quartz | robfig/cron v3（Quartz 秒级表达式） | APScheduler |
| Excel 导入导出 | Apache POI (EasyExcel 风格) | excelize | openpyxl |
| 接口文档 | springdoc | swaggo | FastAPI 自带 + 路径兼容 |

## 快速开始

### 0. 启动本地环境（MySQL + Redis）

推荐使用内置的 Docker 环境（首次启动自动建库 `ry-vue` 并灌入若依预置数据）：

```bash
cd docker
docker compose up -d     # 或 Windows 双击 start.bat
```

详情见 [docker/README.md](docker/README.md)。也可以使用自有的 MySQL 8.0 / Redis，需导入 [`RuoYi-Vue/sql/`](RuoYi-Vue/sql/) 下的初始化脚本。

### 1. 启动后端（三选一，同监听 8080，**同一时间只能运行一个**）

<details open>
<summary><b>Go 版</b></summary>

```bash
cd RuoYi-Vue-GO
cp configs/.env.example configs/.env.dev   # 按需修改数据库/Redis 连接
go run ./cmd/server --env=dev              # 监听 8080
```

环境要求：Go 1.22+。单元测试：`go test ./...`。
</details>

<details>
<summary><b>Python 版</b></summary>

```bash
cd RuoYi-Vue-FastApi
pip install -r requirements.txt
python3 app.py --env=dev                   # 需 Python 3.10+
```
</details>

<details>
<summary><b>Java 版</b></summary>

```bash
cd RuoYi-Vue
mvn clean package -Dmaven.test.skip=true
ruoyi-admin/target/ruoyi-admin.jar         # 需先按 docker/README.md 调整数据源配置，JDK 17
```
</details>

### 2. 启动前端

```bash
cd RuoYi-Vue3
npm install
npm run dev                                # 监听 80，/dev-api 代理到 localhost:8080
```

### 3. 访问

浏览器打开 `http://localhost`，默认账号 **`admin` / `admin123`**。

## 目录速查

| 路径 | 内容 |
|------|------|
| [AGENTS.md](AGENTS.md) | 工作区唯一 AI 编码规范源（契约 / checklist 纪律 / 踩坑清单） |
| [AGENTS.local.md](AGENTS.local.md) | 本机环境配置参考（解释器路径、数据库/Redis 当前指向） |
| [RuoYi-Vue-GO/specs/README.md](RuoYi-Vue-GO/specs/README.md) | Go 版模块总表、动工检查单与遗留待办 |
| [RuoYi-Vue-GO/specs/deviations.md](RuoYi-Vue-GO/specs/deviations.md) | 与 Java 版的全部有意差异（20 条） |
| [RuoYi-Vue-GO/AGENTS.md](RuoYi-Vue-GO/AGENTS.md) | Go 版 AI 编码规范（契约先行/勾选纪律等） |
| [docker/README.md](docker/README.md) | 本地 MySQL + Redis 环境说明 |

## 说明

- 分支管理、提交信息与 Pull Request 同样由 AI 操作完成（提交元数据里的 `Co-Authored-By` 联署可查），人工负责审查与合并
- Java 版 `application.yml` 是数据库/Redis/JWT 配置的基准，各语言版配置与其保持同步
- 欢迎按各目录内的规范文档参与贡献

## License

各子项目沿用其上游 License：[RuoYi-Vue](RuoYi-Vue/LICENSE)（MIT）、[RuoYi-Vue3](RuoYi-Vue3/LICENSE)（MIT）。Go / Python 版代码同样以 MIT 发布。
