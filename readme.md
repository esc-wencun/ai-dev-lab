# RuoYi 多语言服务端：一套前端，三语言后端

以 [RuoYi-Vue](https://github.com/yangzongzhuan/RuoYi-Vue)（Java + Spring Boot）为**接口契约基准**，用 **Go** 和 **Python** 各复刻一套服务端。三个后端共用同一个前端 [RuoYi-Vue3](https://github.com/yangzongzhuan/RuoYi-Vue3)，**前端零改动**即可切换后端。

> 本项目的主要价值在于学习与实践：同一套系统契约，如何在三种语言生态中各自落地（框架选型、ORM 会话、认证鉴权、缓存联动、定时任务等），以及"多后端实现同一契约"时必须严守的兼容纪律。

## 项目来历

[RuoYi-Vue](https://github.com/yangzongzhuan/RuoYi-Vue) 与 [RuoYi-Vue3](https://github.com/yangzongzhuan/RuoYi-Vue3) 来自国内知名开源快速开发平台 **若依（RuoYi）**（官方仓库 [gitee.com/y_project/RuoYi](https://gitee.com/y_project/RuoYi)）：前者是 Spring Boot + Vue 前后端分离版，后者是其 Vue 3 版前端。本仓库中的 `RuoYi-Vue/`、`RuoYi-Vue3/` 为这两个官方仓库的本地副本，仅做运行所需的最低限度调整，分别充当接口契约基准与共用前端；Go / Python 两个服务端则是在此契约之上全新实现的服务端复刻。

## 开发方式：完全 AI 开发，spec 驱动落地

**本项目全部代码由 AI（Claude Code）开发**，人工仅负责需求界定与最终验收。开发流程采用 **spec 驱动**：

- **契约先行**：动手前先通读 Java 版对应 Controller / ServiceImpl / Mapper XML，把端点路径、方法、参数与返回 JSON 结构逐条核实并写成 spec 任务书，不凭记忆实现；
- **按 spec 实现**：每个模块对应一份独立任务书，工程基础 → 基础设施 → 各业务模块顺序推进（spec-00 → spec-10）；
- **checklist 验收**：每个 spec 附验收清单，做完即勾、没做不勾，功能覆盖以数据库实际数据与前端实际调用核对为准。

规范入口见根目录 [AGENTS.md](AGENTS.md)，任务清单见 [`RuoYi-Vue-GO/specs/README.md`](RuoYi-Vue-GO/specs/README.md) 与 [`RuoYi-Vue-FastApi/specs/README.md`](RuoYi-Vue-FastApi/specs/README.md)。

## 项目结构

```
ruoyi/
├── RuoYi-Vue/          Java 版服务端（Spring Boot + MyBatis）—— 接口契约的唯一基准，只读参考
├── RuoYi-Vue3/         前端（Vue 3 + Element Plus + Vite）—— 三个后端共用，不做修改
├── RuoYi-Vue-FastApi/  Python 版服务端（FastAPI + SQLAlchemy 2.0 async）
├── RuoYi-Vue-GO/       Go 版服务端（Gin + GORM）—— 当前开发重点，功能已全部完成
└── docker/             本地开发环境（MySQL 8.0 + Redis 7.0，开箱即用）
```

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

**Go 版已实现全部 11 个模块**，每模块的 API 契约（逐端点对照 Java 源码核实）、任务分解与验收记录见 [`RuoYi-Vue-GO/specs/`](RuoYi-Vue-GO/specs/README.md)，与 Java 版的有意差异集中登记在 [`RuoYi-Vue-GO/specs/deviations.md`](RuoYi-Vue-GO/specs/deviations.md)。

**Python 版亦已实现全部 11 个模块**（142 个路由端点，含算术验证码、密码错误锁定、代码生成器、定时任务 ryTask 预置任务等），任务分解与逐项验收记录见 [`RuoYi-Vue-FastApi/specs/README.md`](RuoYi-Vue-FastApi/specs/README.md)（spec-00~10 全部 ✅，55 个单元测试全绿）。

## 技术栈对照

| 组件 | Java 版 | Go 版 | Python 版 |
|------|---------|-------|-----------|
| Web 框架 | Spring Boot | Gin | FastAPI + uvicorn |
| ORM | MyBatis + Druid | GORM | SQLAlchemy 2.0 (async) |
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
<summary><b>Go 版（推荐，功能最全）</b></summary>

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
python3 app.py --env=dev                   # Windows 用本地 Python 3.10+ 全路径
```
</details>

<details>
<summary><b>Java 版</b></summary>

```bash
cd RuoYi-Vue
mvn clean package -Dmaven.test.skip=true
ruoyi-admin/target/ruoyi-admin.jar         # 需先按 docker/README.md 调整数据源配置
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

## 核心兼容契约（三版共同遵守）

多后端切换"前端零改动"的前提，是所有实现严守以下契约：

- **响应信封**：全部 HTTP 200；成功 `code: 200`，业务失败 `code: 500`，警告 `code: 601`，无权限 `403`，未登录 `401`（前端按 body code 判断）
- **字段命名**：返回 JSON 一律驼峰；日期统一 `yyyy-MM-dd HH:mm:ss`
- **同库同 Redis**：三版共用同一个 MySQL（`ry-vue` 库）与 Redis（键前缀 `login_tokens:` / `captcha_codes:` / `sys_config:` / `sys_dict:` 等逐字一致）
- **密码互相可验**：BCrypt 存量哈希三版通用（Go/Python 可直接登录 Java 创建的账号）
- **会话不互通（已知设计）**：三版 Redis 会话 value 序列化格式不同，切换后端后需重新登录

## 目录速查

| 路径 | 内容 |
|------|------|
| [RuoYi-Vue-GO/specs/README.md](RuoYi-Vue-GO/specs/README.md) | Go 版模块总表、动工检查单与遗留待办 |
| [RuoYi-Vue-GO/specs/deviations.md](RuoYi-Vue-GO/specs/deviations.md) | 与 Java 版的全部有意差异（19 条） |
| [RuoYi-Vue-GO/AGENTS.md](RuoYi-Vue-GO/AGENTS.md) | Go 版 AI 编码规范（契约先行/勾选纪律等） |
| [docker/README.md](docker/README.md) | 本地 MySQL + Redis 环境说明 |

## 说明

- `RuoYi-Vue` / `RuoYi-Vue3` 为若依官方代码的本地副本，仅做运行所需的最低限度调整，新增功能集中在 Go / Python 两个服务端
- Java 版 `application.yml` 是数据库/Redis/JWT 配置的基准，各语言版配置与其保持同步
- 欢迎按各目录内的规范文档参与贡献

## License

各子项目沿用其上游 License：[RuoYi-Vue](RuoYi-Vue/LICENSE)（MIT）、[RuoYi-Vue3](RuoYi-Vue3/LICENSE)（MIT）。Go / Python 版代码同样以 MIT 发布。
