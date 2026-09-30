<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\DictService;
use app\service\ExcelExportService;
use app\service\LogininforService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use think\Request;
use think\Response;

/**
 * 登录日志控制器（对位经典版 SysLogininforController 6 路由）
 */
class LogininforController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '序号', 'field' => 'infoId', 'numeric' => true],
        ['name' => '用户账号', 'field' => 'loginName'],
        ['name' => '登录状态', 'field' => 'status', 'convert' => '0=成功,1=失败'],
        ['name' => '登录地址', 'field' => 'ipaddr'],
        ['name' => '登录地点', 'field' => 'loginLocation'],
        ['name' => '浏览器', 'field' => 'browser'],
        ['name' => '操作系统', 'field' => 'os'],
        ['name' => '提示消息', 'field' => 'msg'],
        ['name' => '访问时间', 'field' => 'loginTime'],
    ];

    /** GET /monitor/logininfor */
    #[Perm('monitor:logininfor:view')]
    public function index(Request $request): Response
    {
        return Response::create('logininfor/logininfor', 'view')->assign([
            'statuses' => DictService::listByType('sys_common_status'),
        ]);
    }

    /** POST /monitor/logininfor/list：TableDataInfo（驼峰 9 列；固定 login_time desc） */
    #[Perm('monitor:logininfor:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['login_name', 'login_time']);
        $query = LogininforService::selectLogininforList($this->filter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([LogininforService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /monitor/logininfor/export */
    #[Perm('monitor:logininfor:export')]
    #[Log('登录日志', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['login_name', 'login_time']);
        $query = LogininforService::selectLogininforList($this->filter($request));
        if ($pq->orderBy !== null) {
            $query->order($pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([LogininforService::class, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '登录日志');
        return AjaxResult::success($fileName);
    }

    /** POST /monitor/logininfor/remove */
    #[Perm('monitor:logininfor:remove')]
    #[Log('登录日志', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $ids = array_values(array_filter(array_map('intval', explode(',', (string)$request->post('ids', ''))), fn($v) => $v > 0));
        return AjaxResult::success(LogininforService::deleteLogininforByIds($ids) > 0 ? '操作成功' : '操作失败');
    }

    /** POST /monitor/logininfor/clean：truncate */
    #[Perm('monitor:logininfor:remove')]
    #[Log('登录日志', Log::CLEAN)]
    public function clean(Request $request): Response
    {
        LogininforService::cleanLogininfor();
        return AjaxResult::success();
    }

    /**
     * POST /monitor/logininfor/unlock?loginName=xxx：解锁（删 pwd_retry 键；恒 success）
     * loginName 走 query string（页面 JS 实锤）；多选逗号串逐个处理；#[Log] 账户解锁,0其它
     */
    #[Perm('monitor:logininfor:unlock')]
    #[Log('账户解锁', Log::OTHER)]
    public function unlock(Request $request): Response
    {
        LogininforService::unlock((string)$request->param('loginName', ''));
        return AjaxResult::success();
    }

    private function filter(Request $request): array
    {
        $params = $request->post('params', []);
        return [
            'ipaddr'    => trim((string)$request->post('ipaddr', '')),
            'loginName' => trim((string)$request->post('loginName', '')),
            'status'    => (string)$request->post('status', ''),
            'beginTime' => (string)($params['beginTime'] ?? ''),
            'endTime'   => (string)($params['endTime'] ?? ''),
        ];
    }
}
