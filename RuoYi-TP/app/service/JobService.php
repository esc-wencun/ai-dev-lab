<?php
declare(strict_types=1);

namespace app\service;

use think\db\Query;
use think\facade\Db;

/**
 * 定时任务服务（对位经典版 SysJobServiceImpl；无 Quartz 内存态——status 字段即调度器唯一状态源）
 */
class JobService
{
    /** 列表 Query（jobName like / jobGroup = / status = / invokeTarget like；排序白名单 job_name/create_time） */
    public static function selectJobList(array $filter): Query
    {
        $query = Db::table('sys_job');
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
        return $query;
    }

    public static function selectJobById(int $jobId): ?array
    {
        return Db::table('sys_job')->where('job_id', $jobId)->find();
    }

    /** 新增（**status 强制 '1' 暂停落库**——经典版 insertJob setStatus(PAUSE)；createBy 由表单 hidden 传入不覆盖） */
    public static function insertJob(array $input): int
    {
        return Db::table('sys_job')->insert(array_merge($input, [
            'status'      => '1',
            'create_time' => date('Y-m-d H:i:s'),
        ])) ? 1 : 0;
    }

    /** 修改（status 由表单 radio 直接携带；update_time 显式写——对位 sysdate()） */
    public static function updateJob(array $input): int
    {
        return Db::table('sys_job')->where('job_id', $input['job_id'])->update(array_merge($input, [
            'update_time' => date('Y-m-d H:i:s'),
        ]));
    }

    /** 批量物理删（循环逐个——经典版 deleteJobByIds 单个循环原样；返回计数） */
    public static function deleteJobByIds(array $ids): int
    {
        $n = 0;
        foreach ($ids as $id) {
            $n += Db::table('sys_job')->where('job_id', $id)->delete();
        }
        return $n;
    }

    /** 启停（selectJobById 前置 → 仅改 status；'0' 恢复 / '1' 暂停） */
    public static function changeStatus(int $jobId, string $status): int
    {
        $job = self::selectJobById($jobId);
        if ($job === null) {
            return 0;
        }
        return Db::table('sys_job')->where('job_id', $jobId)->update(['status' => $status]);
    }

    /** 行 → 驼峰（对位 SysJob Jackson 形态 11 键） */
    public static function toResponseRow(array $r): array
    {
        return [
            'jobId'          => (int)$r['job_id'],
            'jobName'        => $r['job_name'],
            'jobGroup'       => $r['job_group'],
            'invokeTarget'   => $r['invoke_target'],
            'cronExpression' => $r['cron_expression'],
            'misfirePolicy'  => $r['misfire_policy'],
            'concurrent'     => $r['concurrent'],
            'status'         => $r['status'],
            'createTime'     => $r['create_time'],
            'updateTime'     => $r['update_time'],
            'remark'         => $r['remark'],
        ];
    }
}
