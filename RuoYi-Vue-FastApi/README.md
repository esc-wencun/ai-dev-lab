# RuoYi-Vue-FastApi

RuoYi-Vue 的 Python (FastAPI) 版服务端，用于学习 AI + Python 开发。功能对齐 Java 版 [RuoYi-Vue](../RuoYi-Vue) 服务端，前端直接复用 [RuoYi-Vue3](../RuoYi-Vue3)。

> 目录结构和技术选型参考开源项目 [RuoYi-Vue3-FastAPI](https://gitee.com/if-according-to-the-framework_1/ruo-yi-vue3-fast-api)。

## 技术栈

| 组件 | 说明 | 对应 Java 版 |
|------|------|--------------|
| FastAPI + uvicorn | Web 框架 | Spring Boot |
| SQLAlchemy 2.0 (async) + asyncmy | ORM / MySQL 驱动 | MyBatis + Druid |
| redis-py (asyncio) | 缓存 / 会话 | Spring Data Redis |
| python-jose (HS512) | JWT | jjwt |
| passlib (bcrypt) | 密码加密 | BCryptPasswordEncoder（互相兼容） |
| Pillow | 算术验证码 | kaptcha |
| loguru | 日志 | slf4j + logback |
| pydantic-settings + dotenv | 配置 | application.yml |

## 项目结构

```
RuoYi-Vue-FastApi/
├── app.py                  # 启动入口（uvicorn，--env 指定环境配置）
├── server.py               # FastAPI 应用：生命周期事件、controller_list 路由注册、全局组件挂载
├── requirements.txt        # 依赖清单
├── .env.docker / .env.dev  # 环境配置（pydantic-settings 读取，随仓库提交；.env.prod 不入库）
├── common/                 # 通用层（对位 ruoyi-common）：常量、枚举、提示文案、异步后台任务
├── config/                 # 配置层：env 环境加载、database、get_db（事务兜底回滚）、redis_cache 门面
├── exceptions/             # 自定义异常 + 全局异常处理（对位 GlobalExceptionHandler）
├── middlewares/            # 中间件：CORS 等
├── module_admin/           # 系统管理业务模块（对位 ruoyi-system + admin 层）
│   ├── annotation/         # log_decorator 操作日志、防重提交/限流
│   ├── aspect/             # require_perm 接口鉴权、data_scope 数据权限
│   ├── controller/         # 控制器：路由与参数接收，ResponseUtil 组响应
│   ├── service/            # 业务逻辑（登录/验证码/用户/代码生成等）+ gen_templates 生成模板
│   ├── dao/                # 数据访问层：仅查询与写库
│   └── entity/
│       ├── do/             # SQLAlchemy 2.0 表模型（对位 Java domain）
│       └── vo/             # pydantic 请求/响应模型（与 do 刻意分离，见 spec-00）
├── module_task/            # 定时任务（对位 ruoyi-quartz）：APScheduler 调度、任务注册表、RyTask
├── sub_applications/       # 子应用挂载：/profile 静态文件
├── utils/                  # 工具：分页/Excel 导出/驼峰序列化/响应封装/密码/上传/XSS 过滤等
├── tests/                  # pytest 单元测试（纯逻辑）
├── specs/                  # spec 任务书：总清单 README + spec-00~11 + final-acceptance 总验收
├── assets/font/            # 验证码字体（未入库时自动回退系统字体）
├── logs/                   # 运行时日志：sys-info/sys-error/sys-user 按天滚动（git 忽略）
├── upload_path/            # 上传文件存储（git 忽略）
└── download_path/          # 导出/下载临时目录（git 忽略）
```

分层规范（controller → service → dao 的职责边界、Redis 访问纪律等）见根目录 [AGENTS.md](../AGENTS.md)，各模块开发契约见 [specs/README.md](specs/README.md)。

## 当前进度

- [x] 验证码 `/captchaImage`
- [x] 登录 `/login`（JSON 提交，验证码校验、密码错误 5 次锁 10 分钟、IP 黑名单、登录日志）
- [x] 用户信息 `/getInfo`（user / roles / permissions）
- [x] 动态路由 `/getRouters`（完整复刻 Java 版 buildMenus：目录/菜单/外链/内链/ParentView）
- [x] 退出登录 `/logout`、解锁屏幕 `/unlockscreen`
- [x] 用户/角色/菜单/部门/岗位管理（spec-03~05，2026-09-24）
- [x] 字典/参数/通知/日志管理（spec-06~08，2026-09-24）
- [x] 在线用户/服务监控/缓存监控（spec-08，2026-09-24）
- [x] 定时任务/代码生成（spec-09~10，2026-09-24）
- [x] 平台标识 `/getPlatformInfo`（spec-11，跨端模块，2026-09-25）

> 各模块完成状态以 [specs/README.md](specs/README.md) 总清单为准；整体回归验收（[final-acceptance.md](specs/final-acceptance.md)）尚未执行。

## 快速开始

### 1. 环境要求

- Python 3.10+（本机用 uv 建项目专用 3.12 venv，见「3. 安装依赖」）
- MySQL 5.7+（与 Java 版共用 `ry-vue` 库）
- Redis

### 2. 配置

**方式 A（推荐，本地 Docker 环境）**：先按根目录 [docker/README.md](../docker/README.md) 启动 MySQL 8.0 + Redis 7.0，然后直接使用现成的 `.env.docker`（已指向 `127.0.0.1`，账号 `wencun` / `111111`）：

```bash
python app.py --env=docker
```

**方式 B（自备 MySQL/Redis）**：复制 `.env.docker` 为 `.env.dev` 并填入自己的数据库和 Redis 信息（注意：`.env.dev` 随本仓库提交，请勿写入真实凭据；含真实凭据的配置请放 `.env.prod`——已被 git 忽略，不会提交）：

```properties
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USERNAME=wencun
DB_PASSWORD=******
DB_DATABASE=ry-vue
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=******
REDIS_DATABASE=0
```

### 3. 安装依赖

```bash
cd RuoYi-Vue-FastApi
uv venv --python 3.12                        # 首次：创建项目专用 venv（.venv，已 gitignore）
$env:UV_INDEX_URL = "https://mirrors.aliyun.com/pypi/simple"   # uv 走镜像（PowerShell）
uv pip install -r requirements.txt
```

### 4. 启动

```bash
.venv\Scripts\python app.py --env=dev
```

服务监听 `http://localhost:8080`（与 Java 版同端口，前端 `vite.config.js` 代理无需修改）。

接口文档：http://localhost:8080/docs

### 5. 提交前自查

```bash
.venv\Scripts\activate        # 激活项目 venv
ruff check .                  # 代码静态检查（配置见 ruff.toml），零 error 才提交
python -m pytest tests/ -v    # 单元测试
```

一条命令串行执行：`ruff check . && python -m pytest tests/`

### 6. 启动前端

```bash
cd ../RuoYi-Vue3
npm install
npm run dev
```

默认账号：`admin` / `admin123`

## 与 Java 版的关键对应关系

| Java | Python |
|------|--------|
| `SysLoginService` | `module_admin/service/login_service.py` `LoginService` |
| `TokenService`（JWT 存 uuid + Redis 会话） | `login_service.py` `TokenService`（Redis key `login_tokens:` 一致） |
| `SysPermissionService` | `LoginService._get_menu_permission` / `_get_role_permission` |
| `SysPasswordService`（5 次锁 10 分钟，`pwd_err_cnt:`） | `LoginService._validate_password` |
| `CaptchaController`（`captcha_codes:` 2 分钟） | `module_admin/controller/login_controller.py` |
| `AjaxResult {code, msg, data}` | `utils/response_util.py` `ResponseUtil` |
| `CacheConstants` | `config/env.py` `RedisInitKeyConfig` |
| `GlobalExceptionHandler` | `exceptions/handle.py` |

会话存 Redis 的 value 采用 base64url(JSON) 文本（Java 版为 FastJson 带 @type 的格式），因此**登录会话不与 Java 版互通**（同时只有一个服务端在线使用）；但密码哈希（BCrypt）、数据库、Redis 键名约定完全兼容，切换服务端时无需改库。
