# 10.0.0-定时任务 · 任务分解

> 做完即勾；没做不勾并注明原因（勾选纪律见根 AGENTS.md）。
> 顺序：依赖补装+纯逻辑（PHPUnit 可先行）→ 控制器+路由（curl 可验）→ 调度进程（CLI 可验）→ 页面模板（浏览器可验）→ 端到端 → 收尾。
> 每个 Task 可独立验收。2026-10-01 动工并完成。

## Task 1 · 依赖补装与 CronService（纯逻辑先行）

- [x] composer 依赖：**选型变更（用户拍板）——dragonmantank v3.6 实测不支持秒级 cron（5 位分钟制）→ appserver-io/microcron ^2.0**（fork 自 mtdowling v1.2.3 加 SecondsField；6/7 位、n/s 步进、年字段全支持）
- [x] app/service/CronService.php：isValid / getNextRunDate / getMultipleRunDates（+ `?`→`*` 方言转换层）
- [x] PHPUnit：CronServiceTest 7 用例 38 断言（预置 3 条有效/坏表达式/带年 7 位【实测支持——比预判兼容】/秒步进倍数断言/`?` 通配/**周字段实测 1=MON…7=SUN Linux 惯例**）
- [x] spec 差异表实测结果回填（Task 1 重大变更块）；#20 事实定稿（周字段错位确认；年字段实际支持）

## Task 2 · 注册表与目标解析（纯逻辑先行）

- [x] app/task/TaskRegistry.php：bean → 类名映射（ryTask）；resolve 未注册抛「不在白名单内」
- [x] app/task/RyTask.php：三方法 echo 对位 System.out.println
- [x] app/task/TargetParser.php：bean/method/params 拆解 + 引号内逗号不切 + 类型推断五型
- [x] app/task/TargetValidator.php：黑名单四步 + JOB_ERROR_STR 违规串（字面量逐字照抄 Constants.java）
- [x] PHPUnit：JobTargetTest 13 用例 33 断言（解析四形态/类型矩阵/黑名单/注册表）

## Task 3 · 模型与 JobService / JobLogService

- [x] JobService：selectJobList（过滤+白名单排序）/ selectJobById / insertJob（**status 强制 '1'**）/ updateJob / deleteJobByIds（循环物理删）/ changeStatus（selectJobById 前置）
- [x] JobLogService：selectJobLogList（**固定 create_time desc**）/ selectJobLogById / insertJobLog / deleteJobLogByIds / cleanJobLog（truncate）
- [x] 模型层省略（直接 Db 门面操作——项目既有惯例，3.0.0~9.0.0 同）
- [x] PHPUnit：Registry×Parser 集成在 JobTargetTest 覆盖

## Task 4 · TaskExecutor + #[Log] 常量勘误

- [x] app/task/TaskExecutor.php：SETNX joblock（TTL 300s）→ 实例化目标 → 调用 → runMs → finally DEL → jobMessage「{jobName} 总共耗时：{runMs}毫秒」逐字；异常截 2000 写 exception_info；**echo 输出 ob 缓冲捕获（web 端 run 不污染 HTTP 响应——联调期发现并修复）**
- [x] **#[Log] 勘误落地**：GENCODE=8 新增、CLEAN 8→9（app/attribute/Log.php；存量引用 4 处 grep 全核对，语义同步更新）
- [x] TpConstant 新增 PREFIX_JOB_FIRE/PREFIX_JOB_LOCK；RedisCache 补 setNx
- [x] Executor 消息拼装在 curl 级实测（jobMessage 格式逐字核对）

## Task 5 · JobController + 路由（curl 可验）

- [x] JobController 14 方法 + #[Perm] 11 + #[Log] 6
- [x] 校验链顺序与文案逐字（五条 500 实测）
- [x] queryCronExpression 直接构造 {code:0, msg, data:[10 项]}（不走数组 merge）；checkCron 裸 bool
- [x] run 存在性校验 + 同步执行（「任务不存在或已过期！」负向实测）
- [x] detail 页 nextValidTime 实时计算
- [x] 路由 monitor/job 组 14 条注册
- [x] curl 级 15 项全过（m10_test.py）

## Task 6 · JobLogController + 路由（curl 可验）

- [x] JobLogController 6 方法 + #[Perm] 6 + #[Log] 3（clean=9 勘误后）
- [x] GET /monitor/jobLog：可选 jobId 回填（浏览器实测）
- [x] 路由 monitor/jobLog 组 6 条（clean 固定路径显式注册）
- [x] curl 级：list/export/remove/detail/clean 全实测

## Task 7 · 调度进程 `php think scheduler`（CLI 可验）

- [x] app/command/Scheduler.php + config/console.php 注册（think 识别实测）
- [x] 主循环按 spec 伪码：扫库 → 坏 cron 跳过告警 → 回拨 2s → SETNX jobfire 幂等 → TaskExecutor；sleep(1)
- [x] CLI 中文 chcp 65001 提示在注释与启动横幅
- [x] **CLI 级实测**：5 秒周期任务 12 秒 3 行日志；暂停停止（6→6）；**重启进程后启用任务恢复调度（6→9）**
- [x] misfire 无策略语义记录（不补跑设计即此——deviations #10 回填）
- [x] 双开防双跑由 jobfire SETNX 保证（机制在位）

## Task 8 · 页面模板（6 页）

- [x] job index/add/edit/detail 四页（主线程串行移植）
- [x] jobLog/index.html（**两次 Write 故障后第三轮干净重写**——输出流撕裂事故，写后 grep 自查通过）
- [x] job/cron.html（1173 行静态页 volist 矩阵化 ~380 行；输出等价）
- [x] 静态自查：th: 残留 0、杂质串 0、volist/if 配平、check_perm 输出形式
- [x] 联调期修复 2 处：cron.html 缺 `var ctx = '/'`（独立页不套 header）；模板缓存（runtime/temp 清理）

## Task 9 · 端到端验收（浏览器级）

- [x] 任务列表：3 预置行 + 分组徽章 + popover（执行一次/调度日志）实测
- [x] 新增弹窗：createBy hidden/misfire 默认/concurrent 默认/三行帮助（实测）
- [x] 执行一次：确认框 → sys_job_log 成功行（jobMessage 实测）
- [x] 调度日志 openTab + 任务名/分组回填（实测）
- [x] 详情页：双形态（job 形态徽章 + nextValidTime 实测）
- [x] 编辑：全字段回显实测（status radio/updateBy）
- [x] cron 生成器：三按钮全实测（10 次运行时间逐行/验证两文案）；矩阵渲染 60/60
- [x] ry 表零 schema 变更；预置 3 任务复原（status/misfire/concurrent 原值核对）

## Task 10 · 收尾

- [x] PHPUnit 全绿（84 tests 227 assertions，CronServiceTest+JobTargetTest 20 用例新增）；spec.md 实施记录回填；checklist 勾选；README 总表更新（10.0.0 ✅）
- [x] deviations 登记完成：#20 cron 方言 + 选型变更说明；#10/#12 落地一致性自查
- [x] 测试数据清理：测试任务/日志/操作日志/登录日志/Redis 键/导出残留全清；调度器进程已停；admin 会话正常
