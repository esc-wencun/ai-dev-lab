# Tasks · 00 工程基础

> 勾选纪律：做完即勾（含对应单元测试跑通）；没做的不许勾，行尾注明原因；勾选 = 验收通过。
> 依赖：无（本项目第一个模块）。

## Task 1: 工程骨架与配置加载

- [x] `go mod init` + 引入依赖（gin / gorm / go-redis v9 / golang-jwt v5 / viper / zap / validator v10，版本锁定 go.mod）
- [x] `internal/config/`：viper 读取 `.env.dev`（继承 Python 版字段命名），提供全局 `Config` 结构；**配置纪律**：数据库/Redis 值以 Java 版 `application.yml` / `application-druid.yml` 为准手动同步
- [x] `cmd/server/main.go`：装配 config → mysql(GORM) → redis → router → gin engine，监听 8080
- [x] 优雅关闭（监听 SIGINT/SIGTERM，关闭 db/redis 连接）
- [x] 单元测试：配置加载（含缺失必填项报错）

## Task 2: 常量与枚举（对应 ruoyi-common/constant + enums）

- [x] `internal/common/constant/`：CacheConstants 七个前缀（`login_tokens:`/`captcha_codes:`/`pwd_err_cnt:`/`sys_config:`/`sys_dict:`/`repeat_submit:`/`rate_limit:`，已对照 Java CacheConstants.java 逐一核实）、UserConstants（密码/用户名长度限制、菜单类型等）、Constants（TOKEN_PREFIX=`Bearer `、TOKEN=`token`、LOGIN_USER_KEY=`login_user_key`、JWT_USERNAME=subject、CAPTCHA_EXPIRATION=2 分钟、LOGIN_SUCCESS/LOGIN_FAIL/LOGOUT）
- [x] `internal/common/enums/`：BusinessType（序数 0-9 对应 @Log business_type）、UserStatus、OperatorType 等按 Java 逐一移植；用 typed const + String() 方法，不用 iota 制造隐式序数（落库值必须显式等于 Java 值）
- [x] 单元测试：常量值与 Java 源码逐一对表

## Task 3: RedisCache 门面（对应 RedisCache）

- [x] `internal/cache/redis_cache.go`：持有 go-redis client，统一能力——
  - [x] `BuildKey(prefix, parts...)`：前缀只允许来自 CacheConstants
  - [x] `SetObject/GetObject`：JSON 序列化统一进出；分钟级 TTL 对齐 Java 习惯
  - [x] `Delete(keys...)` / `Expire` / `HasKey`
  - [x] `KeysByPrefix`：**SCAN 替代 KEYS**（共享生产 Redis，KEYS 会阻塞）
- [x] 约定：业务代码禁止直接持有 redis client，一律走 RedisCache；2.0.0-登录闭环验证登录链路复用
- [x] 单元测试：key 拼装白名单、JSON 序列化边界（空串/嵌套/时间字段）

## Task 4: 响应信封与统一错误处理（对应 AjaxResult + GlobalExceptionHandler）

- [x] `pkg/response/`：`ResponseUtil`——`Ok(msg)` / `OkData(data)` / `Error(msg)` / `Warn(msg)`(601) / `Forbidden()`(403) / `Unauthorized()`(401)，全部 HTTP 200；自由字段（token/uuid/img/captchaEnabled 等）用链式 `Put(k, v)`（对位 AjaxResult.put）
- [x] **错误处理约定（Go 特有，全局唯一模式）**：定义 `errors.BusinessError{Code, Msg}`；service 层遇业务失败 `return BusinessError`；gin Recovery 中间件兜底 panic 转 `{code:500}`；一个统一中间件把 handler 返回的 error 映射为信封——**禁止**各 handler 自行 c.JSON 拼错误信封
- [x] 面向用户文案进 `internal/common/message/` 字典（key 对齐 Java messages.properties），代码不留裸中文（日志除外）
- [x] 单元测试：各信封 code/msg 断言、BusinessError 映射

## Task 5: JSON 驼峰与时间序列化（对应 Jackson 全局配置）

- [x] 规范：所有 vo struct **json tag 显式写驼峰**，禁止依赖默认字段名输出；加 lint/评审卡点
- [x] 时间统一：自定义 `DateTime` 类型（`time.Time` 别名 + MarshalJSON 输出 `yyyy-MM-dd HH:mm:ss`），do/vo 统一用它——对位 Jackson 全局日期格式；**这是前端日期显示不乱的前提**
- [x] bigint 主键：int64 直接输出数字，保持与 Java 一致（不转 string）
- [x] 单元测试：时间序列化格式、零值行为

## Task 6: 事务与 context 约定（对应 @Transactional）

- [x] GORM 事务封装：提供 `db.Transaction` 包装，约定**多表写入**（用户+关联表、角色+菜单等）必须显式事务；单表写由 service 决定
- [x] **context 传播约定**：dao 层所有方法第一个参数收 `ctx context.Context`，来自 `c.Request.Context()`；GORM/go-redis 调用一律 WithContext——超时与取消全链路生效
- [x] `get_db` 对位：请求结束检查 GORM Session 泄漏（评审项）——纪律已写入 `pkg/database` 包文档（禁止跨请求复用带 ctx 实例），评审时执行
- [x] 单元测试：事务回滚（service 报错时多表写入全部回滚，sqlite 内存库）

## Task 7: 日志框架（对应 logback.xml）

- [x] zap + lumberjack：三文件滚动对齐 Python 版成果——`sys-info.log`（INFO）、`sys-error.log`（ERROR）、`sys-user.log`（登录日志命名 logger），按天滚动、保留 60 天——实现注记：按天滚动用自研 `dailyRotator`（`pkg/logger/rotate.go`，lumberjack 仅支持按大小滚动无法按天），保留 60 天与跨天改名留档行为一致
- [x] 业务统一入口 `pkg/logger`，禁止各包自建 logger；异步写（带缓冲 channel）
- [x] 单元测试：三路日志落文件断言（临时目录）

## Task 8: 工具链（轻量）

- [x] `gofmt`/`go vet` 零告警；`golangci-lint` 可选项记录决策——决策：初期以 gofmt+vet 为门禁暂不引入（已记录 README），模块增多后再评估
- [x] README 写明提交前自查：`gofmt -l . && go vet ./... && go test ./...`
- [x] Makefile 或 task 脚本：run / test / lint 三命令

# Task Dependencies

- Task 2~8 均依赖 Task 1（工程骨架就位）
- Task 3（RedisCache 门面）与 Task 4（响应信封/统一错误处理）是 1.0.0-基础设施 全部 Task 的前置
