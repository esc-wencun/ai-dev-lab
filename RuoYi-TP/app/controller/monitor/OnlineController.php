<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Log;
use app\attribute\Perm;

use app\service\OnlineService;
use AjaxResult;
use PageQuery;
use RedisCache;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 在线用户控制器（对位经典版 SysUserOnlineController 3 路由；数据源 = Redis session:*，sys_user_online 表不读写）
 */
class OnlineController extends \app\BaseController
{
    /** GET /monitor/online */
    #[Perm('monitor:online:view')]
    public function index(Request $request): Response
    {
        return Response::create('online/online', 'view');
    }

    /** POST /monitor/online/list：TableDataInfo（驼峰行；SCAN 管道 + PHP 侧过滤排序分页） */
    #[Perm('monitor:online:list')]
    public function list(Request $request): Response
    {
        $pq = PageQuery::from($request->post(), ['login_name', 'start_timestamp', 'last_access_time']);
        $rows = OnlineService::scan();
        $orderBy = $pq->orderBy !== null ? self::snakeToCamel($pq->orderBy) : '';
        [$pageRows, $total] = OnlineService::filterPage(
            $rows,
            ['ipaddr' => (string)$request->post('ipaddr', ''), 'loginName' => (string)$request->post('loginName', '')],
            $orderBy,
            $pq->isAsc,
            $pq->pageNum,
            $pq->pageSize
        );
        return TableDataInfo::of($pageRows, $total);
    }

    /**
     * POST /monitor/online/batchForceLogout：强退（batchForceLogout|forceLogout 双权限串 OR）
     * 自会话拒；外来会话「该会话无法识别，禁止强退」中止整批；#[Log] 在线用户,7强退
     */
    #[Perm('monitor:online:batchForceLogout')]
    #[Perm('monitor:online:forceLogout')]
    #[Log('在线用户', Log::FORCE)]
    public function batchForceLogout(Request $request): Response
    {
        $selfUuid = (string)($request->middleware('session_uuid') ?? '');
        $ids = array_values(array_filter(array_map('trim', explode(',', (string)$request->post('ids', '')))));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        foreach ($ids as $sessionId) {
            if ($sessionId === $selfUuid) {
                return AjaxResult::error('当前登录用户无法强退');
            }
            if (!RedisCache::has(TpConstant::PREFIX_SESSION . $sessionId, raw: true)) {
                return AjaxResult::error('用户已下线');
            }
            $row = OnlineService::resolve($sessionId);
            if (isset($row['unknown'])) {
                return AjaxResult::error('该会话无法识别，禁止强退');
            }
            OnlineService::forceLogout($sessionId);
        }
        return AjaxResult::success();
    }

    private static function snakeToCamel(string $s): string
    {
        return lcfirst(str_replace(' ', '', ucwords(str_replace('_', ' ', $s))));
    }
}
