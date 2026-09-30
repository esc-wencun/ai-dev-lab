<?php
declare(strict_types=1);

namespace app\service;

use think\db\Query;
use think\facade\Db;

/**
 * 操作日志服务（对位 ISysOperLogService：列表查询/导出全量/批量物理删/truncate/单条）
 */
class OperLogService
{
    /** 列表 Query：like 三处 + businessType/businessTypes in + status + 时间范围；固定 oper_time desc */
    public static function selectOperLogList(array $filter): Query
    {
        $query = Db::table('sys_oper_log');
        if (($filter['operIp'] ?? '') !== '') {
            $query->whereLike('oper_ip', '%' . $filter['operIp'] . '%');
        }
        if (($filter['title'] ?? '') !== '') {
            $query->whereLike('title', '%' . $filter['title'] . '%');
        }
        if (($filter['operName'] ?? '') !== '') {
            $query->whereLike('oper_name', '%' . $filter['operName'] . '%');
        }
        if (($filter['businessType'] ?? '') !== '') {
            $query->where('business_type', (int)$filter['businessType']);
        }
        if (($filter['businessTypes'] ?? '') !== '') {
            $types = array_values(array_filter(array_map('intval', explode(',', (string)$filter['businessTypes'])), fn($v) => $v >= 0));
            if ($types) {
                $query->whereIn('business_type', $types);
            }
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', (int)$filter['status']);
        }
        if (($filter['beginTime'] ?? '') !== '') {
            $query->whereTime('oper_time', '>=', $filter['beginTime']);
        }
        if (($filter['endTime'] ?? '') !== '') {
            $query->whereTime('oper_time', '<=', $filter['endTime']);
        }
        $query->order('oper_time', 'desc');
        return $query;
    }

    /** 批量物理删 */
    public static function deleteOperLogByIds(array $ids): int
    {
        if (!$ids) {
            return 0;
        }
        return Db::table('sys_oper_log')->where('oper_id', 'in', $ids)->delete();
    }

    /** 清空（truncate） */
    public static function cleanOperLog(): void
    {
        Db::execute('TRUNCATE TABLE sys_oper_log');
    }

    /** 单条（详情页） */
    public static function selectOperLogById(int $operId): ?array
    {
        return Db::table('sys_oper_log')->where('oper_id', $operId)->find();
    }

    /** 行 → TableDataInfo 驼峰 17 列（对位 SysOperLog selectVo） */
    public static function toResponseRow(array $r): array
    {
        return [
            'operId'        => (int)$r['oper_id'],
            'title'         => $r['title'],
            'businessType'  => (int)$r['business_type'],
            'method'        => $r['method'],
            'requestMethod' => $r['request_method'],
            'operatorType'  => (int)$r['operator_type'],
            'operName'      => $r['oper_name'],
            'deptName'      => $r['dept_name'],
            'operUrl'       => $r['oper_url'],
            'operIp'        => $r['oper_ip'],
            'operLocation'  => $r['oper_location'],
            'operParam'     => $r['oper_param'],
            'jsonResult'    => $r['json_result'],
            'status'        => (int)$r['status'],
            'errorMsg'      => $r['error_msg'],
            'operTime'      => $r['oper_time'],
            'costTime'      => (int)$r['cost_time'],
        ];
    }
}
