# 00 工程规范基础：目录分层 / 常量枚举 / Redis 门面 / 序列化 / 事务 / context / 错误处理

> **状态：✅ 已完成（2026-09-24）**
>
> 所有模块（含 01）的地基，对齐 Java 版 ruoyi-common 的工程能力；分层与 Python 版同构（handler → service → dao），便于经验迁移。
> Java 版对应：`ruoyi-common`（常量/枚举/RedisCache/AjaxResult）、Jackson 驼峰序列化、`@Transactional`。AsyncManager 的对位（goroutine + recover 异步封装）不单设任务，落在 1.0.0-基础设施 Task 3 操作日志的异步落库中。
> 技术选型依据见 [../tech-stack.md](../tech-stack.md)。

## 目标

立目录分层、语言级约定（Go 特有，Python 版没有对应物，必须在此定死）、常量枚举、Redis 门面、响应信封与统一错误处理。完成后 1.0.0-基础设施 及业务模块都在此骨架上生长。

## 目录结构约定

> 不照搬 Python 版顶层布局，骨架按 Go 社区惯例（project-layout：cmd / internal / pkg），内部分层纪律与调研选型依据见 [../tech-stack.md](../tech-stack.md) 3.2~3.3 节。

```
RuoYi-Vue-GO/
├── cmd/
│   └── server/main.go      # 入口：装配 config/redis/db/router，监听 8080
├── configs/                # .env.dev / .env.prod（值与 Java 版手动同步）
├── internal/               # 私有代码（Go 编译期禁止外部工程 import）
│   ├── config/             # 配置读取（viper），全局 Config 结构
│   ├── common/
│   │   ├── constant/       # CacheConstants / UserConstants / Constants（对位 Java 同名类）
│   │   ├── enums/          # BusinessType / UserStatus / HttpMethod ...（对位 Java enums）
│   │   ├── message/        # 文案字典（对位 messages.properties + MessageUtils）
│   │   └── errors/         # BusinessError 统一业务错误类型
│   ├── cache/              # RedisCache 门面（对位 RedisCache + RedisTemplate；带 key 白名单业务语义，故留 internal）
│   ├── middleware/         # gin 中间件：recovery/响应封装/auth/日志/防重/限流/xss
│   ├── router/             # 路由注册中心（对位 server.py controller_list）
│   └── module/
│       └── admin/          # 业务域（对位 Python module_admin，两版代码可互相索引）
│           ├── handler/    # 参数绑定、调 service、组响应；禁止直接摸 db/redis
│           ├── service/    # 业务逻辑、事务边界、缓存维护；禁止拼 SQL
│           ├── dao/        # 仅 GORM 查询与写库，一个业务域一个文件；禁止业务判断和缓存操作
│           ├── aspect/     # 数据权限等切面（对位 module_admin/aspect）
│           └── model/
│               ├── do/     # GORM 表模型（对位 Java domain）
│               └── vo/     # 请求/响应 struct（json tag 显式驼峰）
├── pkg/                    # 通用工具（无业务依赖，可被外部安全引用）
│   ├── response/           # ResponseUtil 响应信封（AjaxResult 对位）
│   ├── logger/             # zap 封装，业务统一日志入口
│   └── utils/              # 纯函数工具：camel / datetime / ip / useragent / page ...
├── go.mod
└── Makefile                # run / test / lint
```

与 Python 版布局的有意差异（选型期决策，非偏差）：
- 顶层套 `internal/`（Go 特有编译期私有边界）、纯工具上移 `pkg/`
- 单元测试与源码同包（`xxx_test.go`），**不设顶层 tests/ 目录**——Go 惯例
- 不采用 gin-vue-admin 式 `global/` 全局变量：db/redis 句柄显式注入结构体字段，依赖可追溯
- 分层纪律不变：handler → service → dao 逐层约束与 Python 版完全同构

## 关键语言级约定（Go 特有，全局唯一模式）

- **错误处理**：`errors.BusinessError{Code, Msg}`；service 层业务失败统一 return，gin Recovery 兜底 panic，一个统一中间件把 error 映射为响应信封——禁止各 handler 自行拼错误信封（Task 4）
- **JSON 驼峰**：所有 vo struct json tag 显式写驼峰，禁止依赖默认字段名（Task 5）
- **时间序列化**：自定义 `DateTime` 类型输出 `yyyy-MM-dd HH:mm:ss`，对位 Jackson 全局日期格式，是前端日期不乱的前提（Task 5）
- **context 传播**：dao 层第一参数收 `ctx context.Context`（来自 `c.Request.Context()`），GORM/go-redis 一律 WithContext（Task 6）

任务分解见 [tasks.md](tasks.md)，验收清单见 [checklist.md](checklist.md)。

## 实施记录

- 2026-09-24 全部 8 个 Task 完成（gofmt/vet/test 全绿）。
- DateTime 类型落在 `pkg/types`（类型定义归 types，`pkg/utils` 留给纯函数工具），与本文档目录注释略有出入，属有意安排非偏差。
- 按天滚动日志用自研 `dailyRotator`（lumberjack 仅支持按大小滚动）；zap v1.27 异步缓冲写。
- sqlite 内存库选 `glebarez/sqlite`（纯 Go，Windows 免 CGO），仅测试用。
- `module/admin` 的 handler/service/dao"三禁"注释随业务模块建包时写入包注释；工程层各包（response/database/types/logger）契约注释已就位。
- 环境注记：go 工具链在 `C:\Program Files\Go\bin`（go1.27.1），不在系统 PATH，命令行需先 `$env:Path += ";C:\Program Files\Go\bin"`。
