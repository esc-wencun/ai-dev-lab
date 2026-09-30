<?php
declare(strict_types=1);

namespace app\controller\monitor;

use app\attribute\Perm;

use think\facade\Db;
use think\Request;
use think\Response;

/**
 * 数据库监控控制器（自研简页——deviations #6 兑现：经典版此处 302 跳 druid 外部页）
 */
class DataController extends \app\BaseController
{
    /** GET /monitor/data：SHOW STATUS 关键项 + PROCESSLIST + 版本/字符集；只读 */
    #[Perm('monitor:data:view')]
    public function index(Request $request): Response
    {
        $version = (string)Db::query('SELECT version() AS v')[0]['v'];
        $charset = (string)Db::query("SELECT @@character_set_database AS c")[0]['c'];

        // SHOW STATUS 关键项
        $wanted = ['Threads_connected', 'Threads_running', 'Questions', 'Uptime',
                   'Innodb_buffer_pool_read_requests', 'Innodb_buffer_pool_reads', 'Max_used_connections'];
        $statusRows = Db::query('SHOW GLOBAL STATUS');
        $status = [];
        foreach ($statusRows as $row) {
            if (in_array($row['Variable_name'], $wanted, true)) {
                $status[$row['Variable_name']] = $row['Value'];
            }
        }
        $qps = null;
        if (isset($status['Questions'], $status['Uptime']) && (int)$status['Uptime'] > 0) {
            $qps = number_format((int)$status['Questions'] / (int)$status['Uptime'], 1);
        }

        // 进程列表（自用环境不做脱敏——spec 特殊行为 7）
        try {
            $processlist = Db::query('SHOW PROCESSLIST');
        } catch (\Throwable) {
            $processlist = [];
        }

        return Response::create('data/data', 'view')->assign([
            'dbInfo' => [
                'version'     => $version,
                'charset'     => $charset,
                'status'      => $status,
                'qps'         => $qps,
                'processlist' => $processlist, // think-orm Db::query 返回 assoc 数组行，无需转换
            ],
        ]);
    }
}
