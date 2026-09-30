# 10.0.0-定时任务 · spec

> **状态：✅ 已完成（2026-10-01 联调验收通过）**
> 对位经典若依 ruoyi-quartz 模块：SysJobController（14 条路由）/ SysJobLogController（6 条路由）/ SysJobServiceImpl / SysJobLogServiceImpl / ScheduleUtils / JobInvokeUtil / CronUtils / AbstractQuartzJob / RyTask / templates/monitor/job/ 6 页，逐方法核对实锤。
> 依赖：1.0.0（代码层：#[Perm] / #[Log] / PageQuery / TableDataInfo / LoginAuth / ExcelExportService / DictService 只读）；页面渲染于 1.5.0 主框架 iframe 内、运行在 2.0.0 登录态下。
> 与 deviations **#10（自研调度器替代 Quartz，qrtz_* 表不导入）/#12（invoke_target 注册表映射）** 强相关——本 spec 的调度器与目标字符串设计与两条登记严格一致，并把细节定案。
> 调研依据：reference 源码逐文件核对（Controller×2 / ServiceImpl×2 / Mapper XML×2 / 页面 HTML×6 / SysJob、SysJobLog 实体 / RyTask / ScheduleUtils / JobInvokeUtil / CronUtils / AbstractQuartzJob / Constants JOB_* 常量 / ScheduleConstants）+ **ry-tp 库实测**（sys_job 13 列 / sys_job_log 10 列；预置 3 任务 ryTask 系；sys_menu monitor:job 权限 8 行；sys_job_group/sys_job_status/sys_common_status 字典齐全）+ composer 实况核对。
> 已定契约直接引用 tech-stack.md 第四节，不重复展开：code 0/301/500、POST 分页参数、权限两通道（#[Perm] + check_perm）。

## 范围

| TP 版产物 | 对位经典版（源码已核） | 说明 |
|---|---|---|
| app/controller/monitor/JobController.php | SysJobController | 14 条路由（11 个 #[Perm] 方法 + 3 个仅登录态方法） |
| app/controller/monitor/JobLogController.php | SysJobLogController | 6 条路由（clean 固定路径，TP 显式注册无遮蔽问题） |
| app/service/JobService.php | ISysJobService / SysJobServiceImpl | 列表、CRUD、changeStatus/run 状态机、cron 校验；**无 Quartz 内存态**（差异见调度器设计） |
| app/service/JobLogService.php | ISysJobLogService / SysJobLogServiceImpl | 日志查询/写入/批量删/清空（truncate） |
| app/service/CronService.php | CronUtils | cron-expression 封装：isValid / getNextRunDate / getMultipleRunDates |
| app/task/TaskRegistry.php + app/task/RyTask.php | JobInvokeUtil 反射语义 + RyTask | **注册表映射**（deviations #12）：bean 名 → 类名白名单；RyTask 三方法对位预置 3 任务 |
| app/task/TargetParser.php | JobInvokeUtil.getMethodParams | **Java 语法参数解析器**：`bean.method('s', true, 2000L, 316.50D, 100)` 切参 + 类型推断（Python/Go 版同方案先例） |
| app/task/TargetValidator.php | SysJobController 校验链 + ScheduleUtils.whiteList | rmi/ldap/http 黑名单 + 违规串 + 白名单（落地为注册表查名），**文案逐字照抄** |
| app/task/TaskExecutor.php | AbstractQuartzJob | 统一执行器：调目标 + 写 sys_job_log（耗时/异常），run 端点与调度进程共用 |
| app/command/Scheduler.php + config/console.php 注册 | SysJobServiceImpl#@PostConstruct init + Quartz Scheduler | **`php think scheduler` 常驻命令**（另开终端前台，README 环境约束第 4 条） |
| composer 引入 **dragonmantank/cron-expression ^3.x** | org.quartz.CronExpression | **核对结论：尚未安装**（composer.json/lock/vendor 均无，tech-stack 选型表已列但 0.0.0 未装）——Task 1 补装 |
| app/view/monitor/job/*.html ×5 + jobLog/index.html ×1 | templates/monitor/job/（job/add/edit/detail/cron/jobLog） | 6 页，静态 JS（ry-ui.js / bootstrap-table / jquery.validate / js/cron.js）零改动对接；highlight 库已在 static（gen 预览用，本模块 cron 页不用） |

**范围外**（后续模块，勿在本模块实现）：字典管理页面与缓存失效联动（6.0.0，本模块只经 DictService::listByType 只读 sys_job_group/sys_job_status/sys_common_status）；操作日志/登录日志页面（9.0.0，#[Log] 由 1.0.0 OperLog 中间件承接）。

## 页面清单（6）

| # | 页面 | TP 模板路径 | 对位经典版 | 表格/组件 | 核对实锤（页面 JS 请求约定） |
|---|---|---|---|---|---|
| 1 | 任务列表页 | view/monitor/job/index.html | job/job.html | bootstrap-table（POST + server 分页） | `prefix = ctx + "monitor/job"`；options：url=prefix+/list、detailUrl=prefix+/detail/{id}、createUrl=prefix+/add、updateUrl=prefix+/edit/{id}、removeUrl=prefix+/remove、exportUrl=prefix+/export；**sortName=createTime desc**；搜索 jobName / jobGroup（字典 sys_job_group）/ status（字典 sys_job_status）；列：jobId / jobName（链接 $.operate.detail）/ jobGroup（selectDictLabel）/ invokeTarget（$.table.tooltip）/ cronExpression / **任务状态列**（visible 由 statusFlag 控；statusTools 开关图标：status==1 → toggle-off 点击 start()，否则 toggle-on 点击 stop()，确认框后 POST prefix+/changeStatus {jobId, jobGroup, status}）/ createTime（sortable）/ 操作列（编辑/删除 + **更多操作**popover：执行一次 run(jobId)→POST prefix+/run、调度日志 jobLog(jobId)→openTab ctx+"monitor/jobLog?jobId="）；工具栏：新增/修改(single)/删除(multiple)/导出/**生成表达式**（top.layer.open prefix+/cron）/日志（权限 monitor:job:detail） |
| 2 | 任务新增弹窗 | view/monitor/job/add.html | job/add.html | 表单 | 字段：jobName*（maxlength 语义 64）/ jobGroup（sys_job_group 下拉）/ invokeTarget*（help-block 三行提示照抄：Bean 调用示例 ryTask.ryParams('ry')、Class 类调用示例 com.ruoyi…、参数说明支持字符串/布尔/长整型/浮点/整型）/ cronExpression*（validate remote POST prefix+/checkCronExpressionIsValid，文案「表达式不正确」）/ misfirePolicy radio **value=1 立即执行(默认 checked)/2 执行一次/3 放弃执行（无 0 选项）** / concurrent radio 0 允许/1 禁止(默认 checked) / remark；**createBy hidden 域 = 登录名**（对位 @permission.getPrincipalProperty('loginName')）；提交 save(prefix+"/add") |
| 3 | 任务修改弹窗 | view/monitor/job/edit.html | job/edit.html | 表单 | jobId hidden + **updateBy hidden（登录名）**；字段回显（th:field 对位 value 赋值）；cronExpression 同 remote 校验；misfirePolicy radio 回显；**status radio（sys_job_status 字典 volist）——编辑页可直接改状态**；提交 save(prefix+"/edit") |
| 4 | 详情页（双形态共用模板） | view/monitor/job/detail.html | job/detail.html | 静态表单展示 | 控制器传 name=job\|jobLog + 对象变量，模板 `{if $name=='job'}` 分支；job 形态：任务序号/名称/分组（字典 label）/执行状态徽章/cron code 样式/执行策略徽章（0 默认策略 1 立即执行 2 执行一次 3 放弃执行）/并发徽章/**下次执行（nextValidTime，null 显示「未计算」）**/执行方法 invokeTarget/元信息（createBy/createTime/updateBy/updateTime/remark）；jobLog 形态：日志序号/任务名称/分组/执行状态徽章/开始时间/结束时间/记录时间 createTime/**执行耗时（endTime-startTime 毫秒，仅成功态显示）**/调用目标/异常信息（pre，仅失败态）/日志信息 |
| 5 | 调度日志列表页 | view/monitor/jobLog/index.html | job/jobLog.html | bootstrap-table（POST + server 分页） | `prefix = ctx + "monitor/jobLog"`；options：url=prefix+/list、cleanUrl=prefix+/clean、detailUrl=prefix+/detail/{id}、removeUrl=prefix+/remove、exportUrl=prefix+/export；sortName=createTime desc；搜索 jobName（**GET jobId 进页时回填任务名**）/ jobGroup（字典，jobId 进页时选中）/ status（字典 **sys_common_status**）/ params[beginTime] / params[endTime]；列：jobLogId/jobName/jobGroup（字典）/invokeTarget（tooltip）/jobMessage/status（sys_common_status 徽章）/createTime（sortable）/操作详细；工具栏：删除(multiple)/**清空**（$.operate.clean→cleanUrl，确认框）/导出/关闭（closeItem） |
| 6 | Cron 生成器页 | view/monitor/job/cron.html | job/cron.html | 纯静态 + 3 按钮 | js/cron.js 已在 static（零改动）；页内三个交互：①runBtn「查询最近10次运行时间」→ GET prefix+/queryCronExpression?cronExpression=…（result.code==0 取 result.data 逐行拼 HTML）；②unrunBtn「Cron表达式转成字段」纯前端；③checkCron「Cron表达式验证」→ POST prefix+/checkCronExpressionIsValid（裸 boolean：true→「恭喜你，格式正确」/false→「很遗憾，格式错误」）；7 列字段表（秒/分/时/日/月/周/年，**年列仅展示**——库不支持年字段，见 cron 差异节） |

## 端点级 API 清单（job 14 + jobLog 6 = 20 条路由）

### 定时任务 /monitor/job（对位 SysJobController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 1 | GET | /monitor/job | monitor:job:view | 无 | — | 渲染页面 1 |
| 2 | POST | /monitor/job/list | monitor:job:list | 无 | pageNum/pageSize/orderByColumn/isAsc（PageQuery 白名单 job_name/job_group/invoke_target/status/create_time，默认 createTime desc）+ jobName（like）/ jobGroup（=）/ status（=）/ invokeTarget（like） | **TableDataInfo** {code:0, rows, total}；行驼峰：jobId/jobName/jobGroup/invokeTarget/cronExpression/misfirePolicy/concurrent/status/createTime/updateTime/remark（对位 SysJob Jackson 形态） |
| 3 | POST | /monitor/job/export | monitor:job:export | 定时任务, 5导出 | 同搜索参数 + 排序；**无分页（导全量）** | {code:0, msg:"&lt;uuid&gt;_定时任务.xlsx"}——文件名装 msg（经典版 success(String) 重载决议，前端取 result.msg）；Excel 8 列见导出节 |
| 4 | POST | /monitor/job/remove | monitor:job:remove | 定时任务, 3删除 | ids（逗号串） | 循环逐个**物理删**（经典版 deleteJobByIds 单个循环，无子表关联）；成功固定 success()「操作成功」（经典版 return success() 非 toAjax）；调度进程下轮扫不到自然停 |
| 5 | GET | /monitor/job/detail/{jobId} | monitor:job:detail | 无 | 路径 jobId | 渲染页面 4（name='job'），变量 job；模板另需 **nextValidTime**（cron 下一次执行时间，cron 无效/无下次显示「未计算」）与字典 label |
| 6 | POST | /monitor/job/changeStatus | monitor:job:changeStatus | 定时任务, 2修改 | jobId、jobGroup（前端传但后端不消费）、status（'0'/'1'） | **先 selectJobById 再只改 status**（经典版 newJob 语义）；'0'→恢复、'1'→暂停；toAjax；调度进程 ≤1 个 tick 内生效 |
| 7 | POST | /monitor/job/run | monitor:job:changeStatus | 定时任务, 2修改 | jobId | job 不存在 **或 cron 算不出下次执行** → error(500)「任务不存在或已过期！」；否则 web 进程内**同步执行一次**（TaskExecutor，与调度进程经 Redis 执行锁互斥）+ 写 sys_job_log；成功 success()「操作成功」 |
| 8 | GET | /monitor/job/add | monitor:job:add | 无 | — | 渲染页面 2 |
| 9 | POST | /monitor/job/add | monitor:job:add | 定时任务, 1新增 | jobName*、jobGroup、invokeTarget*、cronExpression*、misfirePolicy、concurrent、remark、createBy（表单 hidden 登录名） | **校验链顺序照抄**（见特殊行为 1）；全过 → **status 强制 '1'（暂停）落库**（经典版 insertJob setStatus(PAUSE)）；toAjax |
| 10 | GET | /monitor/job/edit/{jobId} | monitor:job:edit | 无 | 路径 jobId | 渲染页面 3，变量 job |
| 11 | POST | /monitor/job/edit | monitor:job:edit | 定时任务, 2修改 | jobId、jobName*、jobGroup、invokeTarget*、cronExpression*、misfirePolicy、concurrent、**status（radio 可直接改）**、remark、updateBy（表单 hidden 登录名） | 校验链同上（文案前缀「修改任务…」）；updateJob（update_time=sysdate()）；toAjax；调度进程下轮读新 cron 天然生效 |
| 12 | POST | /monitor/job/checkCronExpressionIsValid | **无（仅登录态）** | 无 | cronExpression | **裸 boolean**（jquery validate remote 直接消费）= CronService::isValid |
| 13 | GET | /monitor/job/cron | **无（仅登录态）** | 无 | — | 渲染页面 6（cron 生成器，静态资源 js/cron.js 零改动） |
| 14 | GET | /monitor/job/queryCronExpression | **无（仅登录态）** | 无 | cronExpression（**可空**——空串 checkCron=false → error） | 有效 → **{code:0, msg:"操作成功", data:[10 个 "Y-m-d H:i:s"]}**（从当前时刻起后 10 次触发点，对位 CronTriggerImpl.computeFireTimes(cronTrigger, null, 10)）；无效 → error(500)「表达式无效」 |

### 调度日志 /monitor/jobLog（对位 SysJobLogController）

| # | 方法 | 路径 | #[Perm] | #[Log(title, businessType)] | 参数要点 | 返回要点 |
|---|---|---|---|---|---|---|
| 16 | GET | /monitor/jobLog | monitor:job:view | 无 | 可选 jobId | 渲染页面 5；**带 jobId 时查 job 塞模板变量**（表单回填任务名/分组选中），不带则全量日志页 |
| 17 | POST | /monitor/jobLog/list | monitor:job:list | 无 | pageNum/pageSize + jobName（like）/ jobGroup（=）/ status（=）/ invokeTarget（like）/ params[beginTime] / params[endTime] | TableDataInfo；**固定 order by create_time desc**（mapper 实锤，非前端排序驱动）；行驼峰：jobLogId/jobName/jobGroup/invokeTarget/jobMessage/status/exceptionInfo/startTime/endTime/createTime |
| 18 | POST | /monitor/jobLog/export | monitor:job:export | 调度日志, 5导出 | 同搜索参数；无分页 | {code:0, msg:"&lt;uuid&gt;_调度日志.xlsx"}；Excel 7 列见导出节 |
| 19 | POST | /monitor/jobLog/remove | monitor:job:remove | 调度日志, 3删除 | ids（逗号串） | 物理 delete in；toAjax |
| 20 | GET | /monitor/jobLog/detail/{jobLogId} | monitor:job:detail | 无 | 路径 jobLogId | 渲染页面 4（name='jobLog'），变量 jobLog |
| 21 | POST | /monitor/jobLog/clean | monitor:job:remove | 调度日志, **9清空** | — | **truncate table sys_job_log**；success()「操作成功」 |

统计：**控制器方法 20 个 / 路由 20 条**（job 14 + jobLog 6）；**#[Perm] 17 处**（job 11——add/edit 的页面渲染与保存各自带注解；checkCronExpressionIsValid / cron 页 / queryCronExpression 3 端点无权限注解仅登录态；jobLog 6）；**#[Log] 9 处**（job 6：5导出 / 3删除 / 2修改×3〔changeStatus、run、editSave〕/ 1新增；jobLog 3：5导出 / 3删除 / 9清空）。

## 特殊行为清单（校验文案链 / 状态机 / 参数解析）

1. **新增/修改的 invoke_target 校验链（顺序与文案逐字照抄 SysJobController）**，每步失败即返回 error(500)：
   ① cron 不合法 → `新增任务'{jobName}'失败，Cron表达式不正确`（修改前缀「修改任务…」）
   ② invokeTarget 含 `rmi:` → `…失败，目标字符串不允许'rmi'调用`
   ③ 含 `ldap:` 或 `ldaps:` → `…失败，目标字符串不允许'ldap(s)'调用`
   ④ 含 `http://` 或 `https://` → `…失败，目标字符串不允许'http(s)'调用`
   ⑤ 含违规字符串 → `…失败，目标字符串存在违规`
   ⑥ 不在白名单 → `…失败，目标字符串不在白名单内`
   **TP 版落地语义（deviations #12 细化）**：②③④黑名单照抄（rmi:/ldap:/ldaps:/http(s):// 子串判断）；⑤违规串照抄经典版 JOB_ERROR_STR 字面量（java.net.URL / javax.naming.InitialContext / org.yaml.snakeyaml / org.springframework / org.apache / com.ruoyi.common.utils.file / com.ruoyi.common.config / com.ruoyi.generator——保持文案链完整可触发）；⑥白名单落地为 **TaskRegistry 查名**：解析 invoke_target 得 bean 名（第一个 `(` 前、最后一个 `.` 前的串），必须已注册且方法可调用——比经典版包名前缀白名单更严格（Python 版同先例）。
2. **状态机（无 Quartz 内存态的等价设计）**：
   - **新增** → status 强制 '1'（暂停）落库，不自动跑（经典版同：页面新增后需手动启用；Quartz 侧 createScheduleJob 后立即 pauseJob——TP 无内存 trigger，按 status='1' 跳过即天然等价）。
   - **启用**（changeStatus status='0'）→ 仅 update 库字段；调度进程下轮扫库感知（延迟 ≤1 tick≈1s；经典版 resumeJob 立即生效，微小时延并入 deviations #10 实现说明）。**规避 Python 版踩坑 5**：本方案无调度器内存注册表，不存在「resume 对不在调度器的 job 静默空操作」问题——每轮从库读全量启用任务。
   - **暂停**（changeStatus status='1'）→ 仅 update 库字段；正在执行中的那一次跑完为止，之后不再触发。
   - **编辑** → 库字段更新即生效（调度进程每轮重读 cron/invoke_target，无「改库不同步」问题——经典版 @PostConstruct init 注释里的痛点在 TP 版不存在）。
   - **删除** → 物理删行；调度进程扫不到即停。
   - **立即执行一次**（run）→ 不改 status；web 进程内同步执行 + 写日志；与调度进程的触发经 **Redis 执行锁**互斥（见调度器设计 4）。
3. **run 的存在性校验**：经典版 `scheduler.checkExists(jobKey)`（job 在 Quartz 里存在才 triggerJob，否则「任务不存在或已过期！」——过期指 cron 已无下次触发点）。TP 对位：job 行存在 **且** CronService 能算出下次执行时间，二者其一不满足 → error「任务不存在或已过期！」。
4. **执行日志口径（对位 AbstractQuartzJob.after，run 与调度进程共用 TaskExecutor）**：jobMessage = `{jobName} 总共耗时：{runMs}毫秒`（逐字）；成功 status='0'；失败 status='1' + exception_info（截 2000 字符）；start_time/end_time/create_time（执行时刻）。
5. **Java 语法参数解析器（对位 JobInvokeUtil.getMethodParams）**：
   - 切参正则对位 `,(?=(?:[^"']*["'][^"']*["'])*[^"']*$)`（引号内逗号不切分）；
   - 类型推断照抄：`'x'`/`"x"` 开头 → string（去引号）；true/false（忽略大小写）→ bool；`L` 结尾 → int（2000L→2000）；`D` 结尾 → float（316.50D→316.5）；其余纯数字 → int。PHP 无方法重载，Long/Integer 归一 int、Double 归一 float（PHP 语义，无行为差异感知面）；
   - 预置 3 任务目标实锤（ry-tp 库）：`ryTask.ryNoParams` / `ryTask.ryParams('ry')` / `ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)`——TaskRegistry 注册 `ryTask => app\task\RyTask`，三方法 echo 对位 System.out.println（CLI 输出，chcp 65001 防乱码）。
6. **后端参数校验文案（对位 SysJob @Validated）**：任务名称「任务名称不能为空」「任务名称不能超过64个字符」；调用目标「调用目标字符串不能为空」（长度文案照抄「调用目标字符串长度不能超过500个字符」——**经典版注解 max=1000 与文案 500 不一致、DB varchar(500)，TP 按 DB 500 校验，文案照抄 500**，quirk 勿"修复"）；cron「Cron执行表达式不能为空」。
7. **无唯一性校验**：job_name 可重名（经典版无 checkJobNameUnique；sys_job 主键是 job_id+job_name+job_group 联合主键，job_id 自增为主——think-orm 模型主键设 job_id）。
8. **防重复提交**：经典版 job 控制器无 @RepeatSubmit（grep 实锤）→ 不挂 RepeatSubmit。
9. **页面 quirk（照抄勿修）**：① add 页 misfirePolicy radio 无「默认(0)」选项、concurrent 默认「禁止」；② edit 页可直接改状态而列表页开关也改状态（双入口经典版原样）；③ changeStatus 前端多传 jobGroup 后端不消费；④ invokeTarget @Size max=1000 与文案 500 不一致（见第 6 条）；⑤ detail 页 job 形态有「下次执行」而列表页无此列。

## `php think scheduler` 常驻调度器设计（deviations #10 落地定案）

1. **形态**：TP8 自定义命令 `app/command/Scheduler.php`（config/console.php `'commands' => ['scheduler' => ...]`），另开终端前台常驻 `php think scheduler`，与 web 进程（8888）完全分离；Windows 关窗即停（tech-stack 第六节已记）。**qrtz_* 表零接触**（未导入，自研循环替代 Quartz 全部职责）。
2. **主循环（无状态每轮重算，规避一切内存态/库态同步问题）**：
   ```
   while (true):
     jobs = SELECT * FROM sys_job WHERE status='0'          -- 只扫启用任务（表行数个位数，成本可忽略）
     for job in jobs:
       if !CronService::isValid(cron): log 跳过; continue    -- 坏 cron 单任务跳过+CLI 告警（经典版 init 遇坏 cron 整个启动失败——TP 降级为跳过，登记见拟 deviations）
       next = CronService::getNextRunDate(cron, base=now-2s) -- 基准回拨 2s 覆盖 tick 间隔+处理耗时
       if next <= now:
         fireKey = "jobfire:{jobId}:{next->format('YmdHis')}" -- 同一触发点幂等键（SETNX TTL 120s）
         if RedisCache SETNX(fireKey):                        -- 防同 tick 重复触发 + 防误开双调度进程双跑
           TaskExecutor::run(job)                             -- 同步执行（见 3/4）
     sleep(1)
   ```
   - **不补跑**：基准只回拨 2s，进程停机期间错过的触发点一律跳过、从下个未来触发点恢复（misfire_policy 无对应物，字段照存——见拟 deviations）。
   - 无 lastFire 持久化：重启即「放弃执行」语义，无状态重启安全。
3. **misfire_policy / concurrent 字段处置**：照存照显（页面/导出/详情全保真），**调度策略无对应物**——misfire 四策略（0 默认/1 立即触发/2 触发一次/3 放弃）在本模型下统一退化为「错过不补跑」；concurrent='1'（禁止并发）由执行锁实现（见 4），='0'（允许并发）在**单线程串行**模型下无意义（同一时刻任务天然串行），字段照存不生效。均并入 deviations #10 说明，不另立新条目。
4. **执行锁（concurrent='1' 语义的跨进程实现）**：TaskExecutor 执行前 `SETNX joblock:{jobId}` TTL 300s、finally DEL——run 端点（web 进程）与调度触发（CLI 进程）共用，跨进程互斥；获取失败时调度侧静默跳过本轮、run 侧正常执行（经典版 DisallowConcurrentExecution 语义=同 job 排他，近似对齐）。RedisCache 门面需补 `setNx`/`delete` 原子方法（1.0.0 基建增强，非新依赖）；TpConstant 新增 `PREFIX_JOB_FIRE = 'jobfire:'`、`PREFIX_JOB_LOCK = 'joblock:'`（键前缀集中管理纪律）。
5. **执行失败面**：TaskExecutor 捕获一切异常（含目标类不存在/参数错），写失败日志不中断循环——单个任务之死不拖垮调度进程（对位 AbstractQuartzJob 的 try/catch after）。
6. **与经典版 init 的对应**：`SysJobServiceImpl#@PostConstruct init`（启动时 scheduler.clear() + 全量建 trigger）在本模型中不存在对应物——调度进程启动即从库扫，无需预载；「重启后任务自动恢复调度」的验收等价达成（Python 版验收项同款）。

## cron-expression 与 Quartz cron 语法差异对照（本模块选型关键）

| 维度 | Quartz（经典版） | dragonmantank/cron-expression v3 | 处置 |
|---|---|---|---|
| 字段数 | 6 位（秒 分 时 日 月 周）或 7 位（+年） | 5 位（分 起）或 **6 位（秒在最前）**；**不支持年字段** | 预置 3 任务全是 6 位无年 → 兼容；7 位带年表达式将校验失败（文案「Cron表达式不正确」），登记拟 deviations #20 |
| 日/周位 `?` | 必须互斥使用 | 支持（按通配处理） | 直接兼容（`0/10 * * * * ?` 可解析） |
| **周字段编号** | **1=SUN … 7=SAT** | Linux 惯例 **0/7=SUN，1=MON** | **数值周字段错位一天**——不转义、按 Linux 语义执行，spec 明示 + 登记 #20（预置任务不含周字段，无实际影响面） |
| L / W / # | 支持（Quartz 另有 L-n、LW） | v2.1+ 支持 L/W/#（边界形态以实测为准） | Task 1 PHPUnit 固化常用形态；冷门形态不承诺 |
| 秒级步进 | `0/10` 等 | 支持（秒字段在最前） | 预置任务三件全兼容，PHPUnit 固化 |
| 时区 | trigger 可带 TZ | 用传入 DateTime 的时区 | 全项目统一本地时区，无差异 |
| 下次/N 次计算 | CronExpression.getNextValidTimeAfter / TriggerUtils.computeFireTimes | getNextRunDate / getMultipleRunDates(10) | CronService 封装对位 CronUtils 四方法 |

> 版本核对提醒：以上以 composer 实际解析的 v3.x 为准，Task 1 装完先用 PHPUnit 对预置 3 条表达式 + 上述差异行逐条实测固化，结果回填本表。

> **⚠ Task 1 实测重大变更（2026-10-01，用户拍板）**：dragonmantank/cron-expression **v3.6 实测为 5 位分钟制——不支持秒级字段、不支持 `n/s` 起点步进**（`0/10` 报 Invalid CRON field value；6 位报 "6 is not a valid position"），spec 原差异表关于「v3 6 位秒在最前」的预判错误，预置 3 任务表达式全部无效。**选型变更为 appserver-io/microcron 2.0**（fork 自 mtdowling v1.2.3 加 SecondsField：支持 6/7 位、秒级 `n/s` 步进、年字段）+ CronService 方言转换层（Quartz `?` → `*`，microcron 无日/周互斥语义通配等价；5 位日位 `?` 原样拒绝）。周字段实测：**1=MON … 7=SUN、0=SUN（Linux 惯例）**，与 Quartz 1=SUN 错位确认（#20 登记依据）。mtdowling v1 为 abandoned 元数据提示（指向 dragonmantank），功能完好。

## Excel 导出列定义（@Excel 逐字段抄录）

**SysJob 8 列**（sheet「定时任务」，文件名 `<uuid>_定时任务.xlsx`）：

| 序 | Java 字段 | @Excel 注解 | TP 列定义 |
|---|---|---|---|
| 1 | jobId | name="任务序号", cellType=NUMERIC | 数字格式 |
| 2 | jobName | name="任务名称" | 文本 |
| 3 | jobGroup | name="任务组名" | 文本 |
| 4 | invokeTarget | name="调用目标字符串" | 文本 |
| 5 | cronExpression | name="执行表达式 "（**注解值带尾随空格，照抄**） | 文本 |
| 6 | misfirePolicy | name="计划策略 "（尾随空格照抄）, readConverterExp="0=默认,1=立即触发执行,2=触发一次执行,3=不触发立即执行" | 表达式转换后输出 |
| 7 | concurrent | name="并发执行", readConverterExp="0=允许,1=禁止" | 转换后输出 |
| 8 | status | name="任务状态", readConverterExp="0=正常,1=暂停" | 转换后输出 |

**SysJobLog 7 列**（sheet「调度日志」）：jobLogId「日志序号」/ jobName「任务名称」/ jobGroup「任务组名」/ invokeTarget「调用目标字符串」/ jobMessage「日志信息」/ status「执行状态」readConverterExp="0=正常,1=失败" / exceptionInfo「异常信息」。start_time/end_time/create_time 无 @Excel 不导出。

- 复用 3.0.0 ExcelExportService（列定义数组驱动，含 readConverterExp→convert），零基建增量。

## 关键设计说明

1. **响应结构特例——queryCronExpression**：经典版 `AjaxResult.success(List)` 走 `success(Object data)` 重载 → **data 键装数组**；前端 `result.data` 逐项取。TP 的 `AjaxResult::of` 对 is_array 数据是 merge 到顶层（对位 put(K,V) 多键语义），**数组列表会被拆键破坏结构**——本端点须直接构造 `{code:0, msg:"操作成功", data:[...]}`（think\Response json 或 AjaxResult 增补 dataList 工厂），不得走数组 merge 路径。3.0.0 的 export 走 `success(String)` 重载（文件名进 msg）已验证不受影响。
2. **裸 boolean**：checkCronExpressionIsValid 返回 JSON true/false（jquery validate remote 直接消费），与 3.0.0 check 系端点同模式。
3. **权限两通道接线**：按钮 `shiro:hasPermission` → `{if check_perm(...)}`；页面 JS 变量 `var editFlag/statusFlag/detailFlag = …`（有=''无='hidden'）→ `{:check_perm(...) ? '' : 'hidden'}`；任务状态列的 visible 由 statusFlag 控制（无权限整列隐藏）；字典 `var datas = {:json(DictService::listByType('sys_job_group'))}`（js 消费驼峰键，3.0.0 联调经验：DictService 双键并存）。
4. **createBy/updateBy 来源**：表单 hidden 域携带（对位 `@permission.getPrincipalProperty('loginName')` / edit.html 的 updateBy hidden），后端不覆盖（经典版 addSave 只 setCreateBy——表单已带；editSave 不 setUpdateBy——全靠表单 hidden；照抄）。
5. **detail 页「下次执行」**：控制器渲染时用 CronService 实时计算（对位 SysJob.getNextValidTime() getter），cron 无效时 null → 模板显示「未计算」。
6. **layer 视图机制**（3.0.0 联调经验）：控制器 `app\controller\monitor\JobController` → 视图根 `app/view/monitor/`，渲染路径用 layer 相对（'job/index'）；include 片段需 `app/view/monitor/include/` 局部副本。
7. **API 输出键名驼峰**（3.0.0 联调经验）：bootstrap-table columns field（jobId/invokeTarget 等）按驼峰取值，控制器组装驼峰行。
8. **TableDataInfo 排序**：list 端点排序由 orderByColumn 驱动（sortable 列仅 jobName/createTime，白名单 job_name/create_time）；**jobLog/list 固定 create_time desc 不吃排序参数**（mapper 实锤）。
9. **测试纪律（预置任务是秒级高频）**：预置 job 1/2/3 cron 为 10/15/20 秒步进——端到端验证启用后必须**停回 status='1' 并清 sys_job_log 测试行**，否则日志表被刷爆（写进 checklist 清理记录）。
10. **#[Log] 基建勘误（对 1.0.0 的必要修正，明示非默改）**：app/attribute/Log.php 现定义 CLEAN=8，而经典版 BusinessType 枚举 **GENCODE=8（生成代码）、CLEAN=9（清空）**，sys_dict_data 的 sys_oper_type 字典实锤（8=生成代码/9=清空数据）。本模块 clean 端点要落 9、11.0.0 若做生成端点要落 8——Task 4 修正 Log 常量（新增 GENCODE=8、CLEAN 改 9）。已核对存量影响面：3.0.0 仅用 1/2/3/5，0.0.0/2.0.0 不涉及 8/9，修正无涟漪。

## 拟登记 deviations

#10/#12 已在册，本 spec 为其落地定案；另有一条新增候选，动工核对后正式写入 deviations.md：

| 候选 | 经典若依行为 | TP 版行为 | 处置 |
|---|---|---|---|
| #20 cron 方言差异 | Quartz CronExpression：6/7 位（支持年字段）、周编号 1=SUN…7=SAT、misfire 四策略由 trigger 承接 | cron-expression v3：最多 6 位**不支持年字段**（带年表达式校验失败「Cron表达式不正确」）；周编号按 Linux 惯例 0/7=SUN、1=MON（**数值周字段语义错位一天**）；misfire/concurrent 策略字段照存无调度语义（见调度器设计 3） | 动工 Task 1 实测后正式登记（预置任务全部 6 位无年无周字段，实际影响面=手填 7 位/周级表达式场景） |

另注（非 deviation，仅为防误改说明）：① misfire radio 无「默认(0)」选项、② invokeTarget max 与文案不一致、③ changeStatus 的 jobGroup 冗余参数、④ jobLog/list 排序写死，均为经典版原样，**照抄不修正**。

## 任务分解 → tasks.md

## 验收清单 → checklist.md

## 实施记录（2026-10-01）

- **选型变更（用户拍板）**：dragonmantank/cron-expression v3.6 实测为 5 位分钟制（不支持秒级与 `n/s` 步进——spec 原预判错误），**换 appserver-io/microcron ^2.0**（支持 6/7 位、秒级步进、年字段）+ CronService `?`→`*` 方言转换层。周字段实测 1=MON…7=SUN（Linux 惯例，与 Quartz 错位）→ deviations #20。
- 全链落地：CronService / TaskRegistry（ryTask）/ RyTask / TargetParser（Java 语法参数）/ TargetValidator（黑名单字面量逐字）/ JobService / JobLogService / TaskExecutor（joblock 执行锁 + jobMessage 逐字）+ JobController 14 路由 + JobLogController 6 路由 + `php think scheduler` 常驻命令（jobfire 幂等键防双跑/双开）。
- #[Log] 勘误落地：GENCODE=8 新增、CLEAN 8→9（sys_oper_type 字典实锤口径）。
- CLI 实测：5 秒周期任务 12 秒 3 行日志、暂停停止、重启进程恢复调度；浏览器全链路（列表/popover/执行一次/jobLog 回填/详情双形态/编辑回显/cron 生成器三按钮）。
- 联调期修复 3 处：RyTask echo 污染 HTTP 响应（ob 缓冲）；cron.html 缺 var ctx；模板缓存需清 runtime/temp。
- PHPUnit 84 tests 227 assertions 全绿（新增 CronServiceTest 7 用例 + JobTargetTest 13 用例）；预置 3 任务复原、测试数据全清、调度器进程已停。
