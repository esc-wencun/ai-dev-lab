# AGENTS.md

本文件是本工作区的**唯一 AI 编码规范源**（Single Source of Truth），所有 AI 编码工具（Claude Code / Cursor / Codex / Gemini CLI / Windsurf 等）统一识别。`CLAUDE.md` 通过 `@AGENTS.md` 导入指向本文件，请勿在两处重复维护——修改规范只改这里。

## 工作区结构

本目录是 RuoYi 管理系统的四项目工作区：

- **RuoYi-Vue/** — Java 版服务端（Spring Boot / Java 17 / MyBatis），功能与接口契约的**唯一基准**。只读参考，原则上不修改。
- **RuoYi-Vue3/** — Vue 3 + Element Plus + Vite 前端，两个后端共用，**不修改任何前端代码**。
- **RuoYi-Vue-FastApi/** — Python (FastAPI) 版服务端，当前开发重点（学习 AI + Python 项目）。目标是与 Java 版接口完全兼容，前端零改动即可切换后端。
- **RuoYi-Vue-GO/** — Go 版服务端（**未动工，spec 已就绪**）。定位与 Python 版相同：复刻 Java 版接口、前端零改动。规范入口 `RuoYi-Vue-GO/AGENTS.md`，任务清单 `RuoYi-Vue-GO/specs/README.md`，技术选型 `RuoYi-Vue-GO/specs/tech-stack.md`。同监听 8080，与 Java/Python 版互斥。

## 常用命令

Python 后端（在 `RuoYi-Vue-FastApi/` 下；必须用 `C:\Program Files\Python310\python.exe`，PATH 默认的 Anaconda 是 3.8 不能用）：

```bash
"C:\Program Files\Python310\python.exe" -m pip install -r requirements.txt   # 安装依赖
"C:\Program Files\Python310\python.exe" app.py --env=dev                     # 启动，监听 8080
"C:\Program Files\Python310\python.exe" -m pytest tests/ -v                  # 全部单元测试
"C:\Program Files\Python310\python.exe" -m pytest tests/test_login_logic.py -v            # 单个测试文件
"C:\Program Files\Python310\python.exe" -m pytest tests/test_login_logic.py -k bcrypt -v  # 按名过滤
```

前端（在 `RuoYi-Vue3/` 下）：

```bash
npm install
npm run dev           # 开发服务器，端口 80，/dev-api 代理到 localhost:8080
npm run build:prod    # 生产构建
```

Java 版（仅对照时）：`RuoYi-Vue/` 下 `mvn clean package -Dmaven.test.skip=true`，然后运行 `ruoyi-admin/target/ruoyi-admin.jar`。

接口文档：后端启动后访问 `http://localhost:8080/docs`。默认账号 `admin` / `admin123`。

## 核心架构：Python 版对 Java 版的兼容契约

Python 版存在的意义就是复刻 Java 版接口，以下兼容点是所有开发的前提：

- **响应格式**：统一 `{code, msg, ...}` 信封（`utils/response_util.py` 的 `ResponseUtil`，对位 Java `AjaxResult`）。状态码约定：业务失败 HTTP 200 + body code 500/601；无权限 code 403；未登录 code 401（也是 HTTP 200）。前端按 body code 判断，不能改成 HTTP 状态码语义。
- **字段命名**：返回 JSON 一律驼峰（对位 Jackson 序列化）。
- **同库同 Redis**：MySQL（阿里云 RDS `ry-vue-26-09-24` 库）和 Redis（db11）与 Java 版共用。数据库/Redis 配置以 Java 版 `RuoYi-Vue/ruoyi-admin/src/main/resources/application.yml` 和 `application-druid.yml` 为准，Java 端改动后需手动同步 `.env.dev`。
- **关键兼容点**：BCrypt 密码互相可验；Redis 键前缀 `login_tokens:` / `captcha_codes:` / `pwd_err_cnt:` / `sys_config:`（常量在 `common/constant.py`）；JWT HS512。
- **会话不互通（已知设计，不是 bug）**：Redis 会话 value 格式不同（Java 是 FastJson 带 @type，Python 是纯 JSON），切换后端后所有用户需重新登录。

## Python 版分层规范（对位 Java ruoyi-common / ruoyi-system）

`RuoYi-Vue-FastApi/` 内按 controller → service → dao 分层（`module_admin/` 下），对位 Java 的 Controller / ServiceImpl / Mapper：

- **controller**：参数接收与校验（pydantic vo）、调 service、用 `ResponseUtil` 组响应；禁止直接操作 db/redis。
- **service**：业务逻辑、事务边界、缓存维护；禁止拼 SQL。
- **dao**：仅查询与写库，一个业务域一个 dao 模块；禁止业务判断和缓存操作。
- **entity/do** = SQLAlchemy 2.0 async ORM 表模型（对位 Java domain）；**entity/vo** = pydantic 请求/响应模型。两套模型刻意分离，不用 SQLModel（决策记录见 spec-00）。
- **Redis 访问**一律走 `config/redis_cache.py` 的 `RedisCache` 门面（key 拼装/JSON 序列化/SCAN 替代 KEYS），业务代码禁止直接 `request.app.state.redis`。
- 全局异常处理在 `exceptions/handle.py`（对位 `GlobalExceptionHandler`）；事务约定：多表写入必须包显式事务，`config/get_db.py` 的 `get_db` 异常时兜底回滚。
- 新模块的路由在 `server.py` 的 `controller_list` 注册。

## 开发流程（spec 驱动）

- `RuoYi-Vue-FastApi/specs/README.md` 是总任务清单：spec-00 工程基础 → spec-01 基础设施 → spec-02~10 各业务模块，状态以勾选为准。每个 spec 是一份独立任务书。
- **契约先行**：动手前先读 Java 版对应 Controller + ServiceImpl + Mapper XML，把端点路径、方法、参数、返回 JSON 结构写进 spec 的 API 清单，不凭记忆。
- **验收以前端为准**：接口完成的标准是 RuoYi-Vue3 对应页面能正常操作，且响应字段名、状态码约定与 Java 版一致。
- 测试分层：纯逻辑（树构建、格式转换等）写 pytest 单元测试；涉及数据库/Redis 的用真实环境端到端验证。

## spec checklist 纪律（2026-09 核对事故后新增，必须遵守）

checklist 是任务的验收记录，不是进度装饰。曾出现"状态头写了已完成、勾选框全空"以及"勾选与实际实现不符"的问题，以下规则防止复发：

1. **做完即勾**：每完成一个 checklist 子项（含对应的单元测试/端到端验证跑通），立即在该 spec 文件勾选 `- [x]`，不要攒到最后批量补。
2. **没做的不许勾**：测试没跑、验证没过、依赖外部条件没满足的子项保持 `- [ ]`，并在行尾注明原因（如"——遗留未做（依赖ruff接入）"）。宁可留白，不可虚勾。
3. **spec 状态头 = 全部勾选才许标 ✅**：状态头的"✅ 已完成"只能在该文件所有可完成项勾选、剩余项注明原因之后写。
4. **宣称"已完成"前必须逐条自查**：对照 checklist 逐项问"这个我实际做了吗？验证了吗？"，而不是凭印象。核对时发现的缺口（如 checkUserDataScope、菜单名称唯一校验曾漏实现）要当场补齐再勾。
5. **功能核对以数据为准**：判断"某功能做了没"要扫数据库实际数据（如 sys_job 预置任务）、扫前端实际调用（api/*.js + 页面内 proxy.download/组件内调用），不凭记忆断言全覆盖。

## 已踩过的坑（复现成本高，优先按此排查）

1. **测试脚本 401 假象**：Authorization 头必须带 `Bearer ` 前缀，裸 token 一律 401——排查登录问题先看这个。
2. **uvicorn --reload 残留 worker**：改代码后 reload 子进程可能仍持旧代码服务请求（时灵时不灵的"旧代码"假象）。改完代码验证前，先杀干净所有 python 进程再冷启动。
3. **函数内 `from x import Y` 后在列定义等模块级代码引用 Y**：Python 作用域会把 Y 变局部变量，报 "referenced before assignment"。共享对象（如模板列定义用的 ExcelColumn）放模块顶部导入。
4. **log_decorator 预读 body 会杀死 multipart 上传**（Stream consumed）：装饰器已对 `multipart/form-data` 跳过预读；新增装饰器时注意同类冲突。
5. **暂停态定时任务启用后不调度**：`resume_job` 对不在调度器里的 job 是静默空操作，必须从库补载（`scheduler_util.resume_job` 已修复，勿回退）。
6. **共享 Redis 里 Java 写的 FastJson 缓存**（带 `@type`/`1L`）：解析走 `RedisCache.get_cache_object` 的兼容逻辑；严禁解析失败就删除缓存键（会破坏 Java 侧会话）。
7. **路由遮蔽**：`/xxx/{param}` 会吞掉 `/xxx/clean` 这类固定路径——固定路径路由的处理函数内要对保留字转发（参考 job_controller 的 clean 处理）。

## 环境约束（重要）

1. **端口互斥**：Python 版与 Java 版同监听 8080，**同一时间只能运行一个**。切换后端前先停掉另一个。
2. **共用库纪律**：MySQL 和 Redis 与 Java 版共用，端到端测试产生的数据（测试用户/角色/公告等）**测试完必须清理**；admin 的密码与会话状态测完必须复原（admin/admin123）。
3. 技术选型对位表和已定决策（如不用 SQLModel、MyBatis-Plus 对位暂缓）记录在 `specs/spec-00-engineering-foundation.md`，不要重新发明。
