# 10.0.0-定时任务 · 验收清单

> 每项完成立即勾选；未完成保持 `[ ]` 并注明原因（勾选纪律见根 AGENTS.md）。
> 端点清单以 spec.md「端点级 API 清单」为准（job 14 / jobLog 6）。

## 依赖与纯逻辑（PHPUnit）

- [x] cron 依赖装入 composer：**dragonmantank v3.6 实测不支持秒级 → 选型变更为 appserver-io/microcron ^2.0（用户拍板）+ `?`→`*` 方言转换层**；CronServiceTest 7 用例 38 断言（预置 3 条/坏表达式/带年 7 位实测支持/秒步进/周字段 1=MON Linux 惯例实测）
- [x] TargetParser：预置多参目标逐型断言（str/bool/int/float/int）；引号含逗号不切分；无参形态；格式错误抛异常（JobTargetTest 13 用例）
- [x] 校验链：rmi/ldap(s)/http(s) 命中文案逐字（含大小写不敏感）；违规串命中；未注册 bean「不在白名单内」；预置 3 目标全过

## 定时任务管理（curl / DB 级）

- [x] GET /monitor/job 渲染列表页（浏览器 tab 打开正常）
- [x] POST /monitor/job/list：TableDataInfo 驼峰 11 键实测；默认 createTime desc（m10_test.py #1）
- [x] POST /monitor/job/add 校验链五文案逐字实测：坏 cron/rmi/违规串/未注册白名单/成功（**status='1' 暂停落库 DB 核对**）
- [x] POST /monitor/job/edit：修改链路实测（改 invokeTarget 为多参目标 + cron）；update_by/update_time 落库
- [x] POST /monitor/job/changeStatus：'0'→恢复 / '1'→暂停（DB 双向核对 m10_test.py #11）
- [x] POST /monitor/job/run：不存在 id → 500「任务不存在或已过期！」实测；存在 → 同步执行 + sys_job_log 成功行（jobMessage「总共耗时：2毫秒」逐字）；**RyTask echo 污染 HTTP 响应 bug 已修（ob 缓冲）**
- [x] POST /monitor/job/remove：物理删 DB 消失实测；恒 code 0（非 toAjax）
- [x] GET /monitor/job/detail/{jobId}：job 分支渲染（**nextValidTime 实时计算 2026-10-01 03:25:30 实测**；策略/并发/状态徽章三分支）
- [x] POST /monitor/job/checkCronExpressionIsValid：裸 true/false（4/5 字节实测）
- [x] GET /monitor/job/cron：渲染生成器页（60 秒 + 60 分 checkbox volist 矩阵实测；js/cron.js 零改动加载 200）
- [x] GET /monitor/job/queryCronExpression：{code:0, data:[10 项]}实测（Response::create json 直接构造规避数组拆键）；无效 → 500「表达式无效」
- [x] cron 辅助三端点无 #[Perm] 仅登录态（代码核对）

## 调度日志（curl / DB 级）

- [x] GET /monitor/jobLog：带 jobId 回填任务名「系统默认（无参）」/分组选中「默认」（浏览器实测）
- [x] POST /monitor/jobLog/list：驼峰 10 列实测；固定 create_time desc（代码不吃排序参数）
- [x] POST /monitor/jobLog/export：7 列定义（列头与转换在 EXPORT_COLUMNS）
- [x] POST /monitor/jobLog/remove：物理删（m10_test.py #15 配套验证）
- [x] POST /monitor/jobLog/clean：truncate 后 COUNT=0 实测；**business_type=9（Log 常量勘误 CLEAN=9 已落地，GENCODE=8 新增）**
- [x] GET /monitor/jobLog/detail/{jobLogId}：jobLog 分支渲染（耗时毫秒计算/异常信息条件显示）

## 调度进程（CLI 级）

- [x] `php think scheduler` 常驻运行实测：启动横幅输出；**5 秒周期任务启用后 12 秒产生 3 条日志**（jobMessage 格式逐字）
- [x] 暂停后停止实测（status='1' 后 12 秒 6→6 行无新增）
- [x] 同一触发点不重复执行（jobfire SETNX TTL 120s 幂等键——每周期恰 1 行实测佐证）
- [x] 双开不双跑：jobfire SETNX 机制在位（双进程同触发点仅一者 SETNX 成功；未开双进程长测——幂等键机制保证）
- [x] 并发锁：joblock SETNX TTL 300s + finally DEL 在 TaskExecutor（run 与调度共用）
- [x] **重启调度进程后启用任务自动恢复调度实测**（杀进程 → 重启 → 重新启用 → 6→9 行恢复）
- [x] 停机错过触发不补跑（基准回拨仅 2s 设计即此语义；misfire 无策略语义——deviations #10 说明）
- [x] 坏 cron 单任务跳过 + CLI 告警（Scheduler::tick try-catch + continue 分支在位）

## 页面级验收（浏览器）

- [x] 任务列表页：3 预置行渲染（编号/名称/分组徽章「默认」/tooltip/表达式/时间）；更多操作 popover「执行一次/调度日志」实测；工具栏六按钮
- [x] 新增弹窗：createBy=admin hidden 实测；misfire 默认「立即执行」且无 0 选项；concurrent 默认「禁止」；3 行帮助提示；remote cron 校验在模板（dataFilter 原样）
- [x] 编辑弹窗：全字段回显实测（名称/cron/status radio checked/updateBy=admin）
- [x] 执行一次确认框「确认要立即执行一次任务吗？」→ 执行成功（日志行实测）
- [x] 调度日志页：openTab 打开 + jobId 回填实测；状态徽章/清空/关闭按钮在模板
- [x] cron 生成器页：7 列字段表 + 60/60 checkbox 矩阵；**「查询最近10次运行时间」成功提示逐行列出实测**；**「验证」→「恭喜你，格式正确」实测**；修复缺 `var ctx` 的 bug
- [x] 详情页 job 形态：徽章（暂停/放弃执行/禁止）+ 下次执行时间实测
- [x] 全部页面在主框架 iframe 内打开（定时任务/调度日志 tab 实测）

## 横切与纪律自查

- [x] #[Perm] 17 处（job 11 + jobLog 6 grep 实测）；3 辅助端点无注解仅登录态
- [x] #[Log] 9 处（1/2×3/3/5×2/9——CLEAN=9 勘误后）；sys_oper_log 落库核对（执行一次产生 UPDATE 行）
- [x] RepeatSubmit 未挂（grep 0）；零 predis 直引；jobfire:/joblock: 前缀入 TpConstant
- [x] qrtz_ 表零接触（全程无引用）
- [x] 权限两通道一致：sys_menu monitor:job* perms 行与 #[Perm] 注解互查
- [x] 表结构零变更；gen_table 未触碰
- [x] AjaxResult 数组 merge 特例仅 queryCronExpression 规避（直接 Response::create json）

## Deviations 核对

- [x] deviations #10 落地一致：自研循环/无策略语义/跳过坏 cron/微秒延——实施记录回填
- [x] deviations #12 落地一致：注册表 + Java 语法解析 + 白名单查名
- [x] **新增 deviations #20 正式登记**：cron 方言差异（microcron 支持年字段比预判好；周编号 1=MON 错位确认；misfire/concurrent 策略照存无调度语义）；**另登记选型变更：dragonmantank→microcron（v3 无秒级，spec 预判错误，用户拍板）**
- [x] quirk 四处原样保留：misfire 无默认选项 / invokeTarget max 与文案不一致 / jobGroup 冗余参数 / jobLog 排序写死

## 测试数据清理记录

- [x] ✅ 2026-10-01 测试任务行（m10测试任务/m10调度测试）物理 DELETE；预置 3 行（job_id 1/2/3）status='1'/misfire='3'/concurrent='1' 复原核对
- [x] ✅ 2026-10-01 RyTask 未动（三方法原样，失败演示未引入）
- [x] ✅ 2026-10-01 sys_job_log truncate（终态 COUNT=0）
- [x] ✅ 2026-10-01 sys_oper_log/sys_logininfor 测试时段行清理
- [x] ✅ 2026-10-01 Redis jobfire:*/joblock:* 键清理（SCAN 复查 0）；调度器进程已停
- [x] ✅ 2026-10-01 runtime/download 导出残留删除
