<?php
declare(strict_types=1);

namespace app\service;

use think\db\Query;
use think\facade\Db;

/**
 * 调度日志服务（对位经典版 SysJobLogServiceImpl）
 */
class JobLogService
{
    /** 列表 Query（jobName like / jobGroup = / status = / invokeTarget like / 时间范围）；**固定 create_time desc 不吃排序参数**（mapper 实锤） */
    public static function selectJobLogList(array $filter): Query
    {
        $query = Db::table('sys_job_log');
        if (($filter['jobName'] ?? '') !== '') {
            $query->whereLike('job_name', '%' . $filter['jobName'] . '%');
        }
        if (($filter['jobGroup'] ?? '') !== '') {
            $query->where('job_group', $filter['jobGroup']);
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('status', $filter['status']);
        }
        if (($filter['invokeTarget'] ?? '') !== '') {
            $query->whereLike('invoke_target', '%' . $filter['invokeTarget'] . '%');
        }
        if (($filter['beginTime'] ?? '') !== '') {
            $query->whereTime('create_time', '>=', $filter['beginTime']);
        }
        if (($filter['endTime'] ?? '') !== '') {
            $query->whereTime('create_time', '<=', $filter['endTime']);
        }
        $query->order('create_time', 'desc');
        return $query;
    }

    public static function selectJobLogById(int $jobLogId): ?array
    {
        return Db::table('sys_job_log')->where('job_log_id', $jobLogId)->find();
    }

    /** 执行日志写入（TaskExecutor 与 run 端点共用；jobMessage 格式对位 after：{jobName} 总共耗时：{ms}毫秒） */
    public static function insertJobLog(array $row): bool
    {
        return (bool)Db::table('sys_job_log')->insert(array_merge($row, [
            'create_time' => date('Y-m-d H:i:s'),
        ]));
    }

    public static function deleteJobLogByIds(array $ids): int
    {
        if (!$ids) {
            return 0;
        }
        return Db::table('sys_job_log')->where('job_log_id', 'in', $ids)->delete();
    }

    public static function cleanJobLog(): void
    {
        Db::execute('TRUNCATE TABLE sys_job_log');
    }

    /** 行 → 驼峰（10 键） */
    public static function toResponseRow(array $r): array
    {
        return [
            'jobLogId'      => (int)$r['job_log_id'],
            'jobName'       => $r['job_name'],
            'jobGroup'      => $r['job_group'],
            'invokeTarget'  => $r['invoke_target'],
            'jobMessage'    => $r['job_message'],
            'status'        => $r['status'],
            'exceptionInfo' => $r['exception_info'] ?? '',
            'startTime'     => $r['start_time'],
            'endTime'       => $r['end_time'],
            'createTime'    => $r['create_time'],
        ];
    }
}
