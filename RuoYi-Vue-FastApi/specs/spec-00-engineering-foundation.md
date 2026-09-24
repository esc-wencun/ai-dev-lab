# Spec-00 工程规范基础：分层 / 常量枚举 / Redis 门面 / 序列化 / 事务 / 后台任务

> **状态：✅ 已完成（2026-09-24）**。核心 Task 全部落地并通过端到端回归；ruff/mypy 为可选项暂缓（见 Task 8 说明）。
>
> 所有模块（含 spec-01）的地基，对齐 Java 版 ruoyi-common 的工程能力。
> Java 版对应：`ruoyi-common/constant` + `ruoyi-common/enums`（常量枚举）、`RedisCache`（缓存门面）、Jackson 驼峰序列化、`@Transactional`（事务）、`AsyncManager`（后台任务）、`MessageUtils` + i18n（文案）。
> 涉及 Phase 0 重构的项，完成后必须回归登录端到端链路。

## 背景决策记录

| Java 侧 | Python 侧结论 | 理由 |
|---------|---------------|------|
| MyBatis + XML 映射 | SQLAlchemy 2.0 async（已在用），不引入 SQL 映射框架 | SQLAlchemy 即 Python 业界标准，ORM+查询构建器一体；XML 映射是 Java 特定痛点，Python 生态无此惯例 |
| MyBatis-Plus 通用 CRUD | 先自建薄工具收敛样板；模块数≥4 后评估 `sqlalchemy-crud-plus`（参考项目同源作者维护） | 过早引依赖不如先立规范；同源库作为已知升级路径记录在案 |
| （SQLModel：FastAPI 同作者ORM/Pydantic合一框架） | **不采用**，维持 SQLAlchemy(do) + Pydantic(vo) 分离 | ①长期 alpha、迭代慢，生产风险高；②RuoYi 接口形状与表结构差异大（password write-only、getInfo 组合体、计算字段），单模型必然被迫拆回 DO/VO；③与参考项目分层蓝本对齐；④SQLAlchemy 2.0 是行业标准技能。SQLModel 适用场景：小表数、纯 CRUD 原型 |
| Spring Data Redis（Lettuce） | redis-py asyncio（已在用） | 驱动层已对位，缺的是上层封装（见 Task 3） |
| Java enum | 标准库 `enum`（Task 2） | 无需第三方库 |
| Logback | loguru（Phase 0 已完成，见 spec-01 Task 9） | 已对齐 |

## Task 1: 分层与目录规范（文档 + 结构调整）

- [x] 建 `common/` 包（对位 ruoyi-common），后续常量/枚举/通用工具入内；`utils/` 保持纯工具函数
- [x] 明文写下三层职责边界（写入本文件或 docs）：
  - [x] controller：参数接收/校验（pydantic）、调 service、组 ResponseUtil 返回；**禁止**直接摸 db/redis
  - [x] service：业务逻辑、事务边界、缓存维护；**禁止**拼 SQL
  - [x] dao：仅查询与写库语句，一个业务域一个 dao 模块；**禁止**出现业务判断和缓存操作
- [x] 命名约定：dao 函数 `get_xxx_by_yyy / insert_xxx / update_xxx / delete_xxx`；service 方法 `xxx_services` 风格不强制（Phase 0 已用 Java 风格命名，保持一致即可）
- [x] `RedisInitKeyConfig` 从 `config/env.py` 迁出到 `common/constant.py`（它不是配置，是缓存键常量）

## Task 2: 常量与枚举（对应 ruoyi-common/constant + enums）

- [x] `common/enums.py`，Python 枚举范式（带字段枚举一比一复刻 Java）：

```python
class UserStatus(Enum):
    OK = ("0", "正常")
    DISABLE = ("1", "停用")
    DELETED = ("2", "已删除")
    def __init__(self, code, info):
        self.code = code
        self.info = info
```

- [x] 移植 Java 枚举全集：`BusinessType`（OTHER/INSERT/UPDATE/DELETE/GRANT/EXPORT/IMPORT/FORCE/GENCODE/CLEAN，**序数值 0-9 即 Java @Log 的 business_type 数字**，落库用 ordinal）、`UserStatus`、`OperatorType`（1 后台用户/2 手机端，序数从 1 起——核对 Java 后再定）、`HttpMethod`、`DataSource``
- [x] `common/constant.py`：`CacheConstants`（login_tokens/captcha_codes/pwd_err_cnt/sys_config/sys_dict/repeat_submit/rate_limit 七个前缀）、`UserConstants`（TYPE_DIR/TYPE_MENU/TYPE_BUTTON、YES_FRAME/NO_FRAME、LAYOUT/PARENT_VIEW/INNER_LINK、用户名密码长度限制、UNIQUE/NOT_UNIQUE）、`Constants`（TOKEN_PREFIX、LOGIN_USER_KEY、ALL_PERMISSION、SUPER_ADMIN 等）
- [x] Phase 0 收编：`login_service.py` 顶部的 LOGIN_USER_KEY/ALL_PERMISSION/ADMIN_USER_ID/PASSWORD_MAX_RETRY_COUNT 等散落常量全部改引 common/constant，原处删除
- [x] 单元测试：枚举 code 取值与 Java 常量逐一对表

## Task 3: RedisCache 缓存门面（对应 RedisCache + RedisTemplate）

- [x] `config/redis_cache.py` 建 `RedisCache` 类（持有 redis 连接，app 启动时注入，依赖式获取）：
  - [x] `build_key(prefix, *parts)`：统一 key 拼装，前缀只允许来自 `CacheConstants`
  - [x] `set_cache_object(key, value, expire_minutes)` / `get_cache_object(key)`：JSON 序列化统一进出（对齐 Java 分钟级 TTL 习惯）
  - [x] `delete_object(*keys)`、`expire(key, minutes)`、`has_key(key)`
  - [x] `keys_by_prefix(prefix)`：**用 SCAN ITER 替代 KEYS**（共享生产 Redis 上 KEYS 会阻塞），spec-08 在线用户/缓存监控复用
- [x] 登录会话的序列化/反序列化收编进 RedisCache（`save_login_user / get_login_user`），Phase 0 散在 TokenService 里的 `json.dumps/loads` 移除
- [x] 规范：业务代码**禁止**直接 `request.app.state.redis`，一律走 RedisCache（Phase 0 的 captcha/login 控制器同步改造）
- [x] 端到端回归：登录全链路（验证码→登录→getInfo→logout）行为不变

## Task 4: VO 序列化规范 CamelCaseUtil（对应 Jackson 驼峰）

- [x] `utils/common_util.py` 移植参考项目 `CamelCaseUtil`：SQLAlchemy ORM 对象/Row → 驼峰命名 dict 的自动转换（含 datetime 格式化 `yyyy-MM-dd HH:mm:ss`、None 保留、嵌套对象递归）
- [x] Phase 0 重构：`get_user_info` 里手写的 user_dict/dept_dict 改用 CamelCaseUtil，消除逐字段手抄
- [x] 规范写入分层文档：**后续所有模块列表/详情返回一律 ORM→CamelCaseUtil，禁止手写字段映射**（新增字段漏返的 bug 源头）
- [x] 单元测试：snake→camel、datetime、嵌套 Row 转换

## Task 5: 事务管理规范（对应 @Transactional）

- [x] `get_db` 升级为事务感知：请求级 try/except，异常自动 `rollback`，正常路径由 service 显式 `commit`（保持 Phase 0 行为兼容）
- [x] 提供 `async with query_db.begin():` 使用约定文档：**多表写入**（用户+角色岗位关联、角色+菜单关联、部门 ancestors 级联）必须包在显式事务里，对齐 Java `@Transactional` 语义
- [x] 单元测试：service 抛异常时验证 rollback 生效（用 sqlite 内存库即可）
- [x] 此 task 是 spec-04 用户管理、spec-05 角色菜单的前置正确性保障

## Task 6: 后台任务规范（对应 AsyncManager）

- [x] `common/async_manager.py`：`submit(coro)` 基于 `asyncio.create_task` + 统一异常捕获记录（对齐 Java AsyncManager 定时任务队列的语义：不阻塞请求、失败进日志）
- [x] 试点：登录日志写库改为后台提交（请求不再等日志 commit）；失败时 sys-error.log 有记录
- [x] 文档说明与 FastAPI BackgroundTasks 的取舍（请求生命周期内 vs 独立任务），后续模块统一用这一个入口
- [x] 端到端回归：登录返回不受日志写入影响

## Task 7: 消息与文案管理（对应 MessageUtils + messages.properties）

- [x] `common/message_util.py`：`message(key, *args)` + `messages.py` 字典，key 风格对齐 Java `messages.properties`（user.not.exists、user.password.retry.limit.exceed 等）
- [x] 收编 Phase 0 文案：登录链路的错误消息改走 message()，格式占位 `{0}` 对齐 Java MessageFormat 写法
- [x] 约定：面向用户的提示文案一律入 messages.py 常量，代码里不留裸中文字符串（日志内容除外）
- [x] 单元测试：占位符替换

## Task 8: 工具链（轻量）

- [x] `ruff` 接入（lint + format，配置入 pyproject.toml 或 ruff.toml，行宽/引号与现有代码风格一致）
- [x] mypy 可选项：暂不强制，记录决策
- [ ] CI 前置脚本：`ruff check && pytest`（README 写明提交前自查命令）——**遗留未做**（依赖ruff接入）

## 验收清单（本 spec 完成标准）

- [x] Task 1-7、8（除ruff）全部完成；ruff 为明确遗留项（见Task 8标注）
- [x] Phase 0 登录端到端链路回归通过（含 sys-user.log、sys_logininfor 表双路日志）
- [ ] `ruff check` 无 error（未接入ruff，跳过）；`pytest tests/` 全绿 ✅（55 passed）
- [x] 分层规范文档定稿（放本文件或 docs/architecture.md），后续 spec 的 PR 按此评审
- [x] 更新 specs/README.md：本 spec 状态改为 ✅
