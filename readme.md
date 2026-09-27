# ai-dev-lab：学习与实践 AI coding——用工程纪律驱动 AI 跨技术栈交付一套系统

这个仓库是**本人学习与实践 AI coding 的项目**：通过真实工程练习 spec 驱动开发、checklist 验收纪律与人机协作方法，并按由浅入深的顺序学习以下技术栈：

1. **Python（FastAPI + SQLAlchemy 2.0 async）**——服务端入门：异步 ORM 与分层工程化（spec-00~11 已完成）；
2. **Go（Gin + GORM）**——第二服务端语言：同一契约的跨语言复刻（已全部完成）；
3. **React（React 19 + TypeScript + Ant Design 5 + Redux Toolkit）**——前端栈：功能等价复刻共用前端（2026-09-27 起新增，进行中）；
4. **AI 应用技能**（陆续补充）：**Function Calling / Tool Use** → **RAG（检索增强生成）** → **Agent / 智能体（企业智能体工程化）** → **LangChain4j / Spring AI**（Java 生态 AI 集成）。

项目源于开源项目 [RuoYi](https://gitee.com/y_project/RuoYi)（若依）：`RuoYi-Vue/`（Java 版）与 `RuoYi-Vue3/`（前端）为其官方仓库的本地副本。

## AI 开发工作流：spec 驱动，验收以数据为准

每个模块动工前走同一套流程，全部过程文档在仓库中可查（两个 specs 目录共 **57 份**）：

1. **契约先行**：动手前先通读 Java 基准版对应的 Controller / ServiceImpl / Mapper XML，把端点路径、方法、参数、返回 JSON 结构逐条核实，写成该模块的 spec 任务书——不凭记忆，AI 记忆里的东西必须对着源码核实（为什么，见下面的跌倒案例 1）。
2. **spec 三件套**：`spec.md`（API 契约 + 设计决策）、`tasks.md`（任务分解）、`checklist.md`（验收清单）。跨四端的增量功能用**一份契约主文档**统一管理，各语言版挂自己的实施任务书（范本：[12.0.0 平台标识](RuoYi-Vue-GO/specs/12.0.0-平台标识/spec.md)）。
3. **验收以数据为准，不凭 AI 自述**：AI 说"做完了"不算数——要扫数据库实际数据（如 sys_job 预置任务是否覆盖）、抓前端实际调用（api/*.js + 页面内调用）、跑单元测试、浏览器端到端操作。
4. **checklist 纪律**：做完即勾、没做不勾并在行尾注明原因、**宁可留白不可虚勾**。两版 specs 目录现有 **491 项已勾验收记录**，44 项如实留白。

## 项目结构

```
ai-dev-lab（原 ruoyi 工作区）
├── RuoYi-Vue/          Java 版服务端（Spring Boot + MyBatis-Plus）—— 接口契约的唯一基准，2026-09-25 起可按学习需要增量演进
├── RuoYi-Vue3/         前端（Vue 3 + Element Plus + Vite）—— 三个后端共用，同样可增量演进，仍是 React 版的行为基准
├── RuoYi-Vue-FastApi/  Python 版服务端（FastAPI + SQLAlchemy 2.0 async）
├── RuoYi-Vue-GO/       Go 版服务端（Gin + GORM）
├── RuoYi-React/        React 前端（React 19 + TS + Ant Design 5 + Redux Toolkit）—— 功能等价复刻 RuoYi-Vue3，2026-09-27 起新增
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
| [CHANGELOG.md](CHANGELOG.md) | 里程碑级更新日志（按日期倒序；逐 commit 细节以 git log 为准） |
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
