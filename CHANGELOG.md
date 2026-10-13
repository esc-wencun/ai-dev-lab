# 更新日志（CHANGELOG）

本文件只记录**里程碑级功能更新**：新功能、既有能力的重大改造（如框架替换、架构级重构）。文档调整、叙事修改、配置说明、规范补齐等非功能性操作不记录。逐 commit 细节以 [git log](https://github.com/esc-wencun/ai-dev-lab/commits/main) 为准；模块级的过程与验收记录见各子项目 specs 目录。维护约定：每次完成一个功能里程碑，在此追加一条。

格式参照 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，日期倒序。

## 2026-10-14

### feat(ai): AI 功能线管理面收口——智能体管理页、软删过滤与多会话隔离（specs 2.0.4 ✅ / 2.0.3 🚧）

- **2.0.4 智能体管理页完成**：智能体 CRUD（提示词/默认模型/挂载知识库/开场白/建议问题/可见性四档）、菜单 2030 放出、suggestedQuestions 契约修复（入口前端序列化、出口经 AiAgentPayload 后端转数组，三口实测）；单测 37/37 全绿
- **AI 域软删过滤补齐**：MP 全局逻辑删除开启后 AI 域八处 del_flag 手写过滤补齐，修复删后漏出、查重误报、删会话被引用阻塞三类缺陷；顺带沉淀"MP update(实体) 写不进逻辑删除字段"踩坑（根 AGENTS.md 坑 8）
- **深度思考全链路修复与根除**：定位 reasoning_content text 列容量不足导致流中断（错误路带同一超长字段再落库再炸、异步 ERROR 派发重走 Security 过滤器构成 AuthorizationDeniedException 误诊），拍板"思考仅对话时展示、不落库"，列与实体字段一并移除（迁移 SQL 备好）
- **多会话隔离收口**：后台流回调会话守卫与会话切换聚焦（specs 2.0.2 D14/D15）
- **基础设施**：Redis init SQL 导出脚本化（export-init-sql skill）

## 2026-10-12

### feat(ai): AI 对话域全链路——三表/SSE 流式/思考过程/Claude Code 风格对话页（specs 2.0.2 ✅）

- 对话域三表 + 会话/智能体接口 + AiPromptPlanner + SSE 流式三路落库（specs 2.0.2）
- 深度思考：enable_thinking 会话快照驱动、reasoning 独立 SSE 事件、思考耗时折叠行（specs 2.0.2 D7/D8）
- 前端对话页 Claude Code 终端风格定稿，真实 API 接入（fetch+ReadableStream SSE），组件与 Mock 契约零改动
- ruoyi-ai 模块 Lombok 化；基础域模型连通性测试接口（specs 2.0.1 收尾）
- **AI 开发线首批落地**（specs 2.0.0 ✅ / 2.0.1 ✅，2026-10-11）：Spring AI 2.0.1 + SAA 2.0.0-M1.1（Boot 4 线）、GLM OpenAI 兼容接入、厂商/密钥/模型三表三管理页、AiModelResolver + ChatModelRouter（21 家厂商目录）


## 2026-10-03

- **本地 Docker 环境升级 Redis Stack**：共享 Redis 由 `redis:7.0` 换为 `redis/redis-stack-server`（内置 RediSearch / RedisJSON），具备向量检索能力，满足 Spring AI `RedisVectorStore` 等场景；端口/密码/数据卷不变，三版后端与既有数据零影响。启动参数改经 `REDIS_ARGS` 传入（Stack 镜像用 `command:` 覆盖会绕过模块加载），新增 `maxmemory 512mb` + `noeviction`；`docker/.env` 同步移出 gitignore 刻意入库（内容全为公开本地默认值）。

## 2026-10-01

- **RuoYi-TP（ThinkPHP 8 版）全部 12 模块完成**：复刻经典若依 4.8.3（前后端不分离、服务端渲染），specs 0.0.0~11.0.0 收官——末批完成字典参数（06）、通知公告 summernote 富文本（07）、个人中心含锁屏/注册（08）、监控日志（09）、定时任务 `php think scheduler` 常驻调度（10）、代码生成器按拍板范围 A 完成（11，模板生成链有意排除）；PHPUnit 11 单测 100 项全绿，与经典版差异登记 deviations 33 条。工作区第三条服务端复刻线（学习 PHP / ThinkPHP 生态），自用库 `ry-tp` + Redis db1、端口 8888，与三版完全隔离。
- **登录页默认口令清空**：RuoYi-Vue3 登录表单不再预填 `admin/admin123`。

## 2026-09-29

- **RuoYi-React 主体完成**：布局余项（TagsView 页签/设置抽屉/navType 三模式/暗色主题/通知公告铃铛）与通用组件批次（Pagination/RightToolbar/TreePanel/ExcelImportDialog/Crontab 七域生成器）全部落地，00~07 模块完成、深度验收（CRUD/树表排序/导入导出/角色半选/job 调度全生命周期/锁屏/页签持久化）浏览器实操通过；实现期间修复多处集成 bug，含角色菜单/数据权限提交字段名（menuIdList→menuIds）与 updateSort 参数契约两处接口契约级问题。KeepAlive 页签缓存经演示后拍板不做（deviations #1 定稿）；纯逻辑单测扩至 127 项全绿。README 已声明「尚未进行详细测试」。

## 2026-09-28

- **RuoYi-React 工程落地**：前端工程整体入库，51 个基准页面对应功能全部实现或按拍板排除（build 砍除、gen 暂缓占位）；模块 00/01/05/07 完成、02/03/04/06/08 核心完成，深度验收与 KeepAlive 页签缓存等遗留项待做。
- **Go 版内嵌 tzdata**：打包产物在无 GOROOT 环境下时区（`loc=Asia/Shanghai` DSN）不再失效。

## 2026-09-27

- **RuoYi-React 前端立项**：React 19 + TypeScript + Ant Design 5 + Redux Toolkit，功能等价复刻 RuoYi-Vue3、服务三版后端；specs 0.0.0~8.0.0 任务书建立（进行中）。
- **四端一键打包**：package.bat 升级为 Java 服务端 + Vue3 客户端 + Python 源码包 + Go 三平台二进制的一键打包，按日期时分出版本号。
- **Java 版 MyBatis-Plus 收尾**：代码生成器模板适配 MP（Mapper/Service 增加 IPage 分页方法），DataScope 数据权限栈配对修复。

## 2026-09-26

- **Java 版引入 MyBatis-Plus**（[PR #3](https://github.com/esc-wencun/ai-dev-lab/pull/3)）：移除 PageHelper 切换 MP 分页，数据权限改为 MP 原生适配（specs 0.0.0/1.0.0）。
- **一键打包脚本**：新增 `/package` 命令与 package.bat（Java 服务端 + Vue3 客户端，按日期版本输出）。

## 2026-09-25

- **工作区开源发布**：初始提交——RuoYi 多语言服务端工作区（Java/Go/Python 三版复刻 + 共用前端）。
- **平台标识模块**（[PR #1](https://github.com/esc-wencun/ai-dev-lab/pull/1)、[PR #2](https://github.com/esc-wencun/ai-dev-lab/pull/2)）：新增 `GET /getPlatformInfo` 返回语言/版本/features 能力开关（含 `swaggerDocs`），前端数据监控、服务监控、系统接口页对非 Java 后端降级提示。
