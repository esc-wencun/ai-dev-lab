# RuoYi-React 功能开发总任务清单

> 任务台账：记录模块划分、依赖关系与进度状态。开发规范（契约先行 / 勾选纪律 / 环境约束等）统一维护在 [../../AGENTS.md](../../AGENTS.md)（工作区级），本项目规范入口在 0.0.0 落地时创建的 [../AGENTS.md](../AGENTS.md)，此处不重复。
> 基准：`../../RuoYi-Vue3`（v3.9.2）是**行为契约基准前端**——本项目用 React + Ant Design 复刻其功能，界面不强求一致；行为疑问一律以 RuoYi-Vue3 源码为准。
> 契约侦察基线（2026-09-27 AI 对 RuoYi-Vue3 全量调查产出，实现前必读对应章节）：[reference/01-pages-and-api-baseline.md](reference/01-pages-and-api-baseline.md)（51 页面 + 20 接口文件全端点）、[reference/02-infra-contract.md](reference/02-infra-contract.md)(请求层/守卫/store/布局逐条行为契约 + 易错清单)、[reference/03-components-deps.md](reference/03-components-deps.md)（组件行为 + 依赖映射 + features 开关）。
> 技术选型结论见 [tech-stack.md](tech-stack.md)；与基准前端的有意差异集中登记在 [deviations.md](deviations.md)，总验收对着它过。

## 目录结构

```
specs/
├── README.md            # 本文件：任务台账（模块总表 + 动工检查单）
├── tech-stack.md        # 技术选型调研与结论（唯一选型依据）
├── deviations.md        # 与 RuoYi-Vue3 基准的有意差异登记
├── final-acceptance.md  # 全部完成后的总验收清单
├── reference/           # 契约侦察基线（对 RuoYi-Vue3 v3.9.2 的全量调查，只读参照）
└── <序号>.<功能名>/      # 每个功能模块一个文件夹，内含三件套：
    ├── spec.md          # 契约：范围/对位关系/行为契约要点/设计决策/实施记录
    ├── tasks.md         # 任务分解（原子 Task，做完即勾）
    └── checklist.md     # 验收清单 + 测试数据清理记录
```

## 模块总表

| 序号 | 模块 | 内容 | 状态 | 依赖 |
|------|------|------|------|------|
| 00 | [0.0.0-工程基础](0.0.0-工程基础/spec.md) | Vite+React+TS 脚手架 / 代理 / svg 精灵 / 环境变量 / 目录骨架 / Vitest / 工作区文档同步 | ✅ 2026-09-27 | 无 |
| 01 | [1.0.0-基础设施](1.0.0-基础设施/spec.md) | 请求层契约（401/500/601 分支/防重/下载）/ token / 工具函数 / useDict / API 层 20 文件 | ✅ 2026-09-27 | 00 |
| 02 | [2.0.0-状态与路由](2.0.0-状态与路由/spec.md) | RTK 七 store / 动态路由挂载（Gate 方案）/ 全局守卫 / 登录闭环打通（KeepAlive 不实现见 deviations #1 定稿） | ✅ 2026-09-29 | 01 |
| 03 | [3.0.0-通用组件](3.0.0-通用组件/spec.md) | Pagination/RightToolbar/DictTag/Auth 权限组件/TreePanel/ExcelImportDialog/Crontab/上传四件套等 | ✅ 2026-09-28（FileUpload/ImageUpload/ImagePreview 按需延后，见 tasks） | 01, 02 |
| 04 | [4.0.0-布局主题](4.0.0-布局主题/spec.md) | 布局壳 / 三种导航模式 / TagsView / 设置抽屉 / 暗色主题 / 锁屏页 | ✅ 2026-09-28 | 02, 03 |
| 05 | [5.0.0-登录与个人中心](5.0.0-登录与个人中心/spec.md) | 登录（验证码/记住我）/ 注册 / 错误页 / 首页 / 个人中心（头像裁剪） | ✅ 2026-09-29（改资料/改密表单提交实操遗留，见 tasks） | 02, 03 |
| 06 | [6.0.0-系统管理](6.0.0-系统管理/spec.md) | useCrud 范式 + user/role/menu/dept/post/dict/config/notice 八页 | ✅ 2026-09-29（分配用户子页实操遗留） | 03 |
| 07 | [7.0.0-系统监控](7.0.0-系统监控/spec.md) | online/job(+表达式)/jobLog/druid/server/cache(+echarts)/operlog/logininfor 八页 | ✅ 2026-09-29（Go/Python 降级矩阵随总验收） | 03 |
| 08 | [8.0.0-系统工具](8.0.0-系统工具/spec.md) | 代码生成 gen（**暂缓**，占位页）+ swagger 降级页 + build 占位（deviations #2） | 🔶 2026-09-28（gen 暂缓；swagger Java 已验；降级矩阵随总验收） | 03, 06 |

> 全部完成后执行 [final-acceptance.md](final-acceptance.md) 总验收。
> **进度快照（2026-09-29 三轮收口）**：00~07 全部 ✅，08 🔶（gen 暂缓为长期状态，其余完成）。三轮补齐并实操验证：dept/menu 树表行内排序（修复交互偏差与 updateSort 参数契约）、用户导入导出（updateSupport 双分支）、重置密码后新密码登录、字典联动（refreshCache 进 Redis）、**初始密码提醒弹窗（发现实现缺口并补齐于 getInfo thunk，与基准同位）**、记住我 RSA 密文 Cookie、锁屏硬劫持与解锁回跳（**修复守卫解锁竞态**：!isLock 分支移出守卫，改锁屏页自检 + 渲染期 Navigate）、job 全生命周期（调度器真实运行 102 条日志）。**遗留**：① 5.0.0 改资料/改密表单提交实操（避免动 admin 密码）；② 6.0.0 分配用户子页实操；③ 三版后端矩阵（Python/Go 降级）随 final-acceptance 总验收；④ 8.0.0 gen 暂缓（用户拍板，长期状态）。
> **已拍板决策（2026-09-27 用户）**：① KeepAlive 页签缓存暂不实现，先看无缓存版实际效果再定（spec 02 遗留项）；② 动态路由采用「数据就绪后全量重渲染」Gate 方案；③ tool/build 表单构建器不实现（AI 可直接生成代码）。**追加（2026-09-28）**：④ tool/gen 代码生成 React 版暂缓（占位页）。**追加（2026-09-29）**：⑤ KeepAlive 演示后确认不补（deviations #1 定稿）。

## 动工检查单（每个模块动工前逐条执行）

1. [ ] 读 [reference/](reference/) 对应章节 + RuoYi-Vue3 该模块源码（页面 .vue / store / 组件），列出该模块的页面清单、消费端点清单与行为契约要点，写进该文件夹 `spec.md`
2. [ ] 标注每个页面的按钮级权限串（v-hasPermi 对应值）与字典依赖（sys_xxx）
3. [ ] 识别特殊行为：确认框文案、防重关闭场景（headers.repeatSubmit=false）、下载导出、features 降级、树表/批量排序、富文本 XSS 面
4. [ ] 该模块依赖的通用组件是否已在 3.0.0 实现？没有则先把所需组件补进 3.0.0 的 tasks 并实现
5. [ ] 对照 RuoYi-Vue3 同名文件的实施细节（reference/03 组件行为基线），逐组件登记 props/事件对位关系
6. [ ] 拆 Task 写实该文件夹 `tasks.md`（每 Task 可独立验收）、按写实结果替换细化 `checklist.md`
7. [ ] 与基准前端无法对齐的点登记进 [deviations.md](deviations.md)（格式见该文件登记规则）

## 开发规范与环境约束（指针）

- **契约先行 / 勾选纪律 / 外科手术式修改 / 不主动 git 提交**：统一见 [../../AGENTS.md](../../AGENTS.md)，每个模块动工前通读。
- **三版后端通用性**：本前端与 RuoYi-Vue3 一样服务 Java / Python / Go 三个后端；不判断后端语言，能力差异只消费 `GET /getPlatformInfo` 的 features 三字段（`druidMonitor` / `serverMonitor` / `swaggerDocs`，见 reference/03 §3）。
- **环境约束**：后端 8080 三版互斥；本前端 dev 端口 **8090**（与 RuoYi-Vue3 的 80 并行，便于同屏对照验收）；共用库测试数据测完清理（admin/admin123 复原）。
- 有意差异：[deviations.md](deviations.md)，总验收对着它过。

## 遗留待办

1. ~~KeepAlive 演示后决策~~ → **已闭环（2026-09-28 用户拍板：不补）**，演示记录与定稿见 deviations #1。
2. **tool/build 占位页**：已落地（8.0.0，NotImplemented「该功能未实现」）。
