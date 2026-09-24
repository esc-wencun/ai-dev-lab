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

## 当前进度

- [x] 验证码 `/captchaImage`
- [x] 登录 `/login`（JSON 提交，验证码校验、密码错误 5 次锁 10 分钟、IP 黑名单、登录日志）
- [x] 用户信息 `/getInfo`（user / roles / permissions）
- [x] 动态路由 `/getRouters`（完整复刻 Java 版 buildMenus：目录/菜单/外链/内链/ParentView）
- [x] 退出登录 `/logout`、解锁屏幕 `/unlockscreen`
- [ ] 用户/角色/菜单/部门/岗位管理
- [ ] 字典/参数/通知/日志管理
- [ ] 在线用户/服务监控/缓存监控
- [ ] 定时任务/代码生成

## 快速开始

### 1. 环境要求

- Python 3.10+
- MySQL 5.7+（与 Java 版共用 `ry-vue` 库）
- Redis

### 2. 配置

**方式 A（推荐，本地 Docker 环境）**：先按根目录 [docker/README.md](../docker/README.md) 启动 MySQL 8.0 + Redis 7.0，然后直接使用现成的 `.env.docker`（已指向 `127.0.0.1`，账号 `wencun` / `111111`）：

```bash
python app.py --env=docker
```

**方式 B（自备 MySQL/Redis）**：复制 `.env.docker` 为 `.env.dev` 并填入自己的数据库和 Redis 信息（`.env.dev` 已被 git 忽略，不会提交）：

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
"C:\Program Files\Python310\python.exe" -m pip install -r requirements.txt
```

### 4. 启动

```bash
"C:\Program Files\Python310\python.exe" app.py --env=dev
```

服务监听 `http://localhost:8080`（与 Java 版同端口，前端 `vite.config.js` 代理无需修改）。

接口文档：http://localhost:8080/docs

### 5. 启动前端

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
