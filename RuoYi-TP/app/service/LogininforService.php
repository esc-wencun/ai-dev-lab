<?php
declare(strict_types=1);

namespace app\service;

use RedisCache;
use TpConstant;
use think\db\Query;
use think\facade\Db;

/**
 * 登录日志服务（对位 ISysLogininforService：列表查询/导出/批量删/truncate/unlock=删 pwd_retry 键）
 */
class LogininforService
{
    /** 列表 Query：like 两处 + status + 时间范围；固定 login_time desc */
    public static function selectLogininforList(array $filter): Query
    {
        $query = Db::table('sys_logininfor');
        if (($filter['ipaddr'] ?? '') !== '') {
            $query->whereLike('ipaddr', '%' . $filter['ipaddr'] . '%');
        }
        if (($filter['loginName'] ?? '') !== '') {
            $query->whereLike('login_name', '%' . $filter['loginName'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', (string)$filter['status']);
        }
        if (($filter['beginTime'] ?? '') !== '') {
            $query->whereTime('login_time', '>=', $filter['beginTime']);
        }
        if (($filter['endTime'] ?? '') !== '') {
            $query->whereTime('login_time', '<=', $filter['endTime']);
        }
        $query->order('login_time', 'desc');
        return $query;
    }

    public static function deleteLogininforByIds(array $ids): int
    {
        if (!$ids) {
            return 0;
        }
        return Db::table('sys_logininfor')->where('info_id', 'in', $ids)->delete();
    }

    public static function cleanLogininfor(): void
    {
        Db::execute('TRUNCATE TABLE sys_logininfor');
    }

    /**
     * 账户解锁（对位 passwordService.clearLoginRecordCache）：删 pwd_retry 键；
     * 不校验用户存在（键不存在也成功——经典版 remove 语义原样）；$loginNames 支持逗号串多选。
     */
    public static function unlock(string $loginNames): void
    {
        foreach (array_filter(array_map('trim', explode(',', $loginNames))) as $name) {
            RedisCache::delete(TpConstant::PREFIX_PWD_RETRY . $name);
        }
    }

    /** 行 → TableDataInfo 驼峰 9 列 */
    public static function toResponseRow(array $r): array
    {
        return [
            'infoId'        => (int)$r['info_id'],
            'loginName'     => $r['login_name'],
            'ipaddr'        => $r['ipaddr'],
            'loginLocation' => $r['login_location'],
            'browser'       => $r['browser'],
            'os'            => $r['os'],
            'status'        => $r['status'],
            'msg'           => $r['msg'],
            'loginTime'     => $r['login_time'],
        ];
    }
}
