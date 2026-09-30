<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\DictService;
use app\service\ExcelExportService;
use app\service\JobLogService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use think\Request;
use think\Response;

/**
 * 调度日志控制器（对位经典版 SysJobLogController 6 路由）
 */
class JobLogController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '日志序号', 'field' => 'jobLogId', 'numeric' => true],
        ['name' => '任务名称', 'field' => 'jobName'],
        ['name' => '任务组名', 'field' => 'jobGroup'],
        ['name' => '调用目标字符串', 'field' => 'invokeTarget'],
        ['name' => '日志信息', 'field' => 'jobMessage'],
        ['name' => '执行状态', 'field' => 'status', 'convert' => '0=正常,1=失败'],
        ['name' => '异常信息', 'field' => 'exceptionInfo'],
    ];

    /** GET /monitor/jobLog：可选 jobId 回填任务名/分组 */
    #[Perm('monitor:job:view')]
    public function index(Request $request): Response
    {
        $assign = [
            'groups'   => DictService::listByType('sys_job_group'),
            'statuses' => DictService::listByType('sys_common_status'),
        ];
        $jobId = (int)$request->get('jobId', 0);
        if ($jobId > 0) {
            $assign['job'] = \app\service\JobService::selectJobById($jobId);
        }
        return Response::create('jobLog/index', 'view')->assign($assign);
    }

    /** POST /monitor/jobLog/list：固定 create_time desc（不吃排序参数） */
    #[Perm('monitor:job:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), []);
        $params = $request->post('params', []);
        $query = JobLogService::selectJobLogList([
            'jobName'      => trim((string)$request->post('jobName', '')),
            'jobGroup'     => (string)$request->post('jobGroup', ''),
            'status'       => (string)$request->post('status', ''),
            'invokeTarget' => trim((string)$request->post('invokeTarget', '')),
            'beginTime'    => (string)($params['beginTime'] ?? ''),
            'endTime'      => (string)($params['endTime'] ?? ''),
        ]);
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([JobLogService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /monitor/jobLog/export */
    #[Perm('monitor:job:export')]
    #[Log('调度日志', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $params = $request->post('params', []);
        $query = JobLogService::selectJobLogList([
            'jobName'      => trim((string)$request->post('jobName', '')),
            'jobGroup'     => (string)$request->post('jobGroup', ''),
            'status'       => (string)$request->post('status', ''),
            'invokeTarget' => trim((string)$request->post('invokeTarget', '')),
            'beginTime'    => (string)($params['beginTime'] ?? ''),
            'endTime'      => (string)($params['endTime'] ?? ''),
        ]);
        $rows = array_map([JobLogService::class, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '调度日志');
        return AjaxResult::success($fileName);
    }

    /** POST /monitor/jobLog/remove */
    #[Perm('monitor:job:remove')]
    #[Log('调度日志', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        return AjaxResult::success(JobLogService::deleteJobLogByIds($ids) > 0 ? '操作成功' : '操作失败');
    }

    /** GET /monitor/jobLog/detail/{jobLogId}：详情页 jobLog 形态 */
    #[Perm('monitor:job:detail')]
    public function detail(Request $request, int $jobLogId): Response
    {
        $jobLog = JobLogService::selectJobLogById($jobLogId);
        if ($jobLog === null) {
            throw new \BusinessException('调度日志不存在');
        }
        return Response::create('job/detail', 'view')->assign([
            'name'        => 'jobLog',
            'jobLog'      => $jobLog,
            'groupLabel'  => DictService::getLabel('sys_job_group', (string)$jobLog['job_group']),
            'statusLabel' => DictService::getLabel('sys_common_status', (string)$jobLog['status']),
        ]);
    }

    /** POST /monitor/jobLog/clean：truncate（#[Log] 调度日志, 9清空——CLEAN=9 勘误后） */
    #[Perm('monitor:job:remove')]
    #[Log('调度日志', Log::CLEAN)]
    public function clean(Request $request): Response
    {
        JobLogService::cleanJobLog();
        return AjaxResult::success();
    }
}
