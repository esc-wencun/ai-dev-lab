<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\DictService;
use app\service\ExcelExportService;
use app\service\OperLogService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use think\Request;
use think\Response;

/**
 * 操作日志控制器（对位经典版 SysOperlogController 6 路由）
 */
class OperlogController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '操作序号', 'field' => 'operId', 'numeric' => true],
        ['name' => '操作模块', 'field' => 'title'],
        ['name' => '业务类型', 'field' => 'businessType', 'numeric' => true, 'convert' => '0=其它,1=新增,2=修改,3=删除,4=授权,5=导出,6=导入,7=强退,8=生成代码,9=清空数据'],
        ['name' => '请求方法', 'field' => 'method'],
        ['name' => '请求方式', 'field' => 'requestMethod'],
        ['name' => '操作类别', 'field' => 'operatorType', 'numeric' => true, 'convert' => '0=其它,1=后台用户,2=手机端用户'],
        ['name' => '操作人员', 'field' => 'operName'],
        ['name' => '部门名称', 'field' => 'deptName'],
        ['name' => '请求地址', 'field' => 'operUrl'],
        ['name' => '操作地址', 'field' => 'operIp'],
        ['name' => '操作地点', 'field' => 'operLocation'],
        ['name' => '请求参数', 'field' => 'operParam'],
        ['name' => '返回参数', 'field' => 'jsonResult'],
        ['name' => '状态', 'field' => 'status', 'numeric' => true, 'convert' => '0=正常,1=异常'],
        ['name' => '错误消息', 'field' => 'errorMsg'],
        ['name' => '操作时间', 'field' => 'operTime'],
        ['name' => '消耗时间(毫秒)', 'field' => 'costTime', 'numeric' => true],
    ];

    /** GET /monitor/operlog */
    #[Perm('monitor:operlog:view')]
    public function index(Request $request): Response
    {
        return Response::create('operlog/operlog', 'view')->assign([
            'types'    => DictService::listByType('sys_oper_type'),
            'statuses' => DictService::listByType('sys_common_status'),
        ]);
    }

    /** POST /monitor/operlog/list：TableDataInfo（驼峰 17 列；固定 oper_time desc） */
    #[Perm('monitor:operlog:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['oper_name', 'oper_time', 'cost_time']);
        $query = OperLogService::selectOperLogList($this->filter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([OperLogService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /monitor/operlog/export */
    #[Perm('monitor:operlog:export')]
    #[Log('操作日志', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['oper_name', 'oper_time', 'cost_time']);
        $query = OperLogService::selectOperLogList($this->filter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([OperLogService::class, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '操作日志');
        return AjaxResult::success($fileName);
    }

    /** POST /monitor/operlog/remove */
    #[Perm('monitor:operlog:remove')]
    #[Log('操作日志', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        return AjaxResult::success(OperLogService::deleteOperLogByIds($ids) > 0 ? '操作成功' : '操作失败');
    }

    /** GET /monitor/operlog/detail/{operId} */
    #[Perm('monitor:operlog:detail')]
    public function detail(Request $request, int $operId): Response
    {
        $operLog = OperLogService::selectOperLogById($operId);
        if ($operLog === null) {
            throw new \BusinessException('操作日志不存在');
        }
        return Response::create('operlog/detail', 'view')->assign([
            'operLog'   => $operLog,
            'typeLabel' => DictService::getLabel('sys_oper_type', (string)$operLog['business_type']),
        ]);
    }

    /** POST /monitor/operlog/clean：truncate */
    #[Perm('monitor:operlog:remove')]
    #[Log('操作日志', Log::CLEAN)]
    public function clean(Request $request): Response
    {
        OperLogService::cleanOperLog();
        return AjaxResult::success();
    }

    private function filter(Request $request): array
    {
        $params = $request->post('params', []);
        return [
            'operIp'        => trim((string)$request->post('operIp', '')),
            'title'         => trim((string)$request->post('title', '')),
            'operName'      => trim((string)$request->post('operName', '')),
            'businessType'  => (string)$request->post('businessType', ''),
            'businessTypes' => (string)$request->post('businessTypes', ''),
            'status'        => (string)$request->post('status', ''),
            'beginTime'     => (string)($params['beginTime'] ?? ''),
            'endTime'       => (string)($params['endTime'] ?? ''),
        ];
    }
}
