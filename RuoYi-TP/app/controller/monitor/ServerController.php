<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Perm;

use app\service\ServerInfoService;
use think\Request;
use think\Response;

/**
 * 服务监控控制器（对位经典版 ServerController；oshi 降级——deviations #11 三档采集）
 */
class ServerController extends \app\BaseController
{
    /** GET /monitor/server：四区块纯服务端渲染 */
    #[Perm('monitor:server:view')]
    public function index(Request $request): Response
    {
        return Response::create('server/server', 'view')->assign(ServerInfoService::collect());
    }
}
