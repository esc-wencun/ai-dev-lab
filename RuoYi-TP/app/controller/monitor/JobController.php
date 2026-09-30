<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\CronService;
use app\service\DictService;
use app\service\ExcelExportService;
use app\service\JobService;
use app\task\TargetValidator;
use app\task\TaskRegistry;
use app\task\TargetParser;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 定时任务控制器（对位经典版 SysJobController 14 路由；校验链顺序与文案逐字照抄）
 */
class JobController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '任务序号', 'field' => 'jobId', 'numeric' => true],
        ['name' => '任务名称', 'field' => 'jobName'],
        ['name' => '任务组名', 'field' => 'jobGroup'],
        ['name' => '调用目标字符串', 'field' => 'invokeTarget'],
        ['name' => '执行表达式 ', 'field' => 'cronExpression'],
        ['name' => '计划策略 ', 'field' => 'misfirePolicy', 'convert' => '0=默认,1=立即触发执行,2=触发一次执行,3=不触发立即执行'],
        ['name' => '并发执行', 'field' => 'concurrent', 'convert' => '0=允许,1=禁止'],
        ['name' => '任务状态', 'field' => 'status', 'convert' => '0=正常,1=暂停'],
    ];

    /** GET /monitor/job */
    #[Perm('monitor:job:view')]
    public function index(Request $request): Response
    {
        return Response::create('job/index', 'view')->assign([
            'groups' => DictService::listByType('sys_job_group'),
            'statuses' => DictService::listByType('sys_job_status'),
        ]);
    }

    /** POST /monitor/job/list：TableDataInfo（驼峰 11 键；默认 createTime desc） */
    #[Perm('monitor:job:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['job_name', 'job_group', 'invoke_target', 'status', 'create_time']);
        $query = JobService::selectJobList($this->filter($request));
        // 默认 create_time desc（sortName 实锤）；显式排序参数覆盖
        $orderBy = $pq->orderBy ?? 'create_time';
        $query->order($orderBy, $pq->orderBy !== null ? $pq->isAsc : 'desc');
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([JobService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /monitor/job/export */
    #[Perm('monitor:job:export')]
    #[Log('定时任务', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['job_name', 'job_group', 'invoke_target', 'status', 'create_time']);
        $query = JobService::selectJobList($this->filter($request));
        $orderBy = $pq->orderBy ?? 'create_time';
        $query->order($orderBy, $pq->orderBy !== null ? $pq->isAsc : 'desc');
        $rows = array_map([JobService::class, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '定时任务');
        return AjaxResult::success($fileName);
    }

    /** POST /monitor/job/remove：循环物理删；固定 success（经典版 return success() 非 toAjax） */
    #[Perm('monitor:job:remove')]
    #[Log('定时任务', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        JobService::deleteJobByIds($ids);
        return AjaxResult::success();
    }

    /** GET /monitor/job/detail/{jobId}：双形态详情页（name=job；含实时下次执行） */
    #[Perm('monitor:job:detail')]
    public function detail(Request $request, int $jobId): Response
    {
        $job = JobService::selectJobById($jobId);
        if ($job === null) {
            throw new \BusinessException('任务不存在');
        }
        $nextValidTime = CronService::getNextRunDate((string)$job['cron_expression']);
        return Response::create('job/detail', 'view')->assign([
            'name'          => 'job',
            'job'           => $job,
            'nextValidTime' => $nextValidTime ? $nextValidTime->format('Y-m-d H:i:s') : null,
            'groupLabel'    => DictService::getLabel('sys_job_group', (string)$job['job_group']),
            'statusLabel'   => DictService::getLabel('sys_job_status', (string)$job['status']),
        ]);
    }

    /** POST /monitor/job/changeStatus：先 selectJobById 再只改 status */
    #[Perm('monitor:job:changeStatus')]
    #[Log('定时任务', Log::UPDATE)]
    public function changeStatus(Request $request): Response
    {
        $jobId = (int)$request->post('jobId', 0);
        $status = (string)$request->post('status', '1');
        return AjaxResult::success(JobService::changeStatus($jobId, in_array($status, ['0', '1'], true) ? $status : '1') > 0 ? '操作成功' : '操作失败');
    }

    /** POST /monitor/job/run：存在性校验（行存在 + cron 有下次）→ 同步执行一次 */
    #[Perm('monitor:job:changeStatus')]
    #[Log('定时任务', Log::UPDATE)]
    public function run(Request $request): Response
    {
        $job = JobService::selectJobById((int)$request->post('jobId', 0));
        if ($job === null || CronService::getNextRunDate((string)$job['cron_expression']) === null) {
            return AjaxResult::error('任务不存在或已过期！');
        }
        \app\task\TaskExecutor::run($job);
        return AjaxResult::success();
    }

    /** GET /monitor/job/add */
    #[Perm('monitor:job:add')]
    public function add(Request $request): Response
    {
        return $this->formView('job/add', [
            'groups'  => DictService::listByType('sys_job_group'),
            'sessionLoginName' => (string)(($request->middleware('session') ?? [])['loginName'] ?? ''),
        ]);
    }

    /** POST /monitor/job/add：校验链 → status 强制 '1' 落库 */
    #[Perm('monitor:job:add')]
    #[Log('定时任务', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $input = $this->jobInput($request);
        $msg = $this->validateChain($input, '新增任务');
        if ($msg !== null) {
            return AjaxResult::error($msg);
        }
        return AjaxResult::success(JobService::insertJob($input) > 0 ? '操作成功' : '操作失败');
    }

    /** GET /monitor/job/edit/{jobId} */
    #[Perm('monitor:job:edit')]
    public function edit(Request $request, int $jobId): Response
    {
        $job = JobService::selectJobById($jobId);
        if ($job === null) {
            throw new \BusinessException('任务不存在');
        }
        return $this->formView('job/edit', [
            'job'     => $job,
            'groups'  => DictService::listByType('sys_job_group'),
            'statuses' => DictService::listByType('sys_job_status'),
            'sessionLoginName' => (string)(($request->middleware('session') ?? [])['loginName'] ?? ''),
        ]);
    }

    /** POST /monitor/job/edit：校验链（修改前缀）→ update */
    #[Perm('monitor:job:edit')]
    #[Log('定时任务', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $input = $this->jobInput($request);
        $input['job_id'] = (int)$request->post('jobId', 0);
        $msg = $this->validateChain($input, '修改任务');
        if ($msg !== null) {
            return AjaxResult::error($msg);
        }
        return AjaxResult::success(JobService::updateJob($input) > 0 ? '操作成功' : '操作失败');
    }

    /** POST /monitor/job/checkCronExpressionIsValid：裸 boolean（无 #[Perm]） */
    public function checkCronExpressionIsValid(Request $request): Response
    {
        return \think\Response::create(CronService::isValid((string)$request->post('cronExpression', '')) ? 'true' : 'false');
    }

    /** GET /monitor/job/cron：cron 生成器页（js/cron.js 零改动；checkbox 矩阵用 volist 生成——输出与经典版静态 HTML 等价） */
    public function cron(Request $request): Response
    {
        return Response::create('job/cron', 'view')->assign([
            'seconds' => range(0, 59),
            'minutes' => range(0, 59),
            'hours'   => range(0, 23),
            'days'    => range(1, 31),
            'months'  => range(1, 12),
            'weeks'   => range(1, 7),
        ]);
    }

    /** GET /monitor/job/queryCronExpression：{code:0, msg, data:[10 项]}（**直接构造不走数组 merge**——spec 关键设计 1） */
    public function queryCronExpression(Request $request): Response
    {
        $cron = (string)$request->get('cronExpression', '');
        $dates = CronService::getMultipleRunDates($cron, 10);
        if (!$dates) {
            return AjaxResult::error('表达式无效');
        }
        return Response::create(['code' => TpConstant::CODE_SUCCESS, 'msg' => '操作成功', 'data' => $dates], 'json');
    }

    /* ---------- 内部 ---------- */

    private function formView(string $template, array $vars): Response
    {
        return Response::create($template, 'view')->assign($vars);
    }

    /** 校验链（顺序与文案逐字照抄 SysJobController）：cron → rmi → ldap(s) → http(s) → 违规串 → 白名单 */
    private function validateChain(array $input, string $prefix): ?string
    {
        $jobName = $input['job_name'];
        if (!CronService::isValid($input['cron_expression'])) {
            return "{$prefix}'{$jobName}'失败，Cron表达式不正确";
        }
        $target = $input['invoke_target'];
        $err = TargetValidator::check($target);
        if ($err !== null) {
            return "{$prefix}'{$jobName}'失败，{$err}";
        }
        try {
            $bean = TaskRegistry::beanOf($target);
            TaskRegistry::resolve($bean);
            TargetParser::parse($target);
        } catch (\Throwable) {
            return "{$prefix}'{$jobName}'失败，目标字符串不在白名单内";
        }
        return null;
    }

    /** 表单输入收敛（@Validated 文案照抄；invokeTarget 按 DB varchar(500) 校验——经典版注解 max=1000 与文案 500 不一致 quirk） */
    private function jobInput(Request $request): array
    {
        $jobName = trim((string)$request->post('jobName', ''));
        if ($jobName === '') {
            throw new \BusinessException('任务名称不能为空');
        }
        if (mb_strlen($jobName) > 64) {
            throw new \BusinessException('任务名称不能超过64个字符');
        }
        $invokeTarget = trim((string)$request->post('invokeTarget', ''));
        if ($invokeTarget === '') {
            throw new \BusinessException('调用目标字符串不能为空');
        }
        if (mb_strlen($invokeTarget) > 500) {
            throw new \BusinessException('调用目标字符串长度不能超过500个字符');
        }
        $cron = trim((string)$request->post('cronExpression', ''));
        if ($cron === '') {
            throw new \BusinessException('Cron执行表达式不能为空');
        }
        return [
            'job_name'        => $jobName,
            'job_group'       => (string)$request->post('jobGroup', 'DEFAULT'),
            'invoke_target'   => $invokeTarget,
            'cron_expression' => $cron,
            'misfire_policy'  => (string)$request->post('misfirePolicy', '1'),
            'concurrent'      => (string)$request->post('concurrent', '1'),
            'status'          => (string)$request->post('status', '1'),
            'remark'          => (string)$request->post('remark', ''),
            // createBy/updateBy 表单 hidden 域携带（经典版后端不覆盖——spec 关键设计 4）
            'create_by'       => (string)$request->post('createBy', ''),
            'update_by'       => (string)$request->post('updateBy', ''),
        ];
    }

    private function filter(Request $request): array
    {
        return [
            'jobName'      => trim((string)$request->post('jobName', '')),
            'jobGroup'     => (string)$request->post('jobGroup', ''),
            'status'       => (string)$request->post('status', ''),
            'invokeTarget' => trim((string)$request->post('invokeTarget', '')),
        ];
    }
}
