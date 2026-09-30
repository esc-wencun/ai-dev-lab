<?php
declare(strict_types=1);

namespace app\task;

use app\service\JobLogService;
use RedisCache;
use TpConstant;

/**
 * 统一任务执行器（对位经典版 AbstractQuartzJob；run 端点与调度进程共用）
 *
 * 执行锁：SETNX joblock:{jobId} TTL 300s（跨进程互斥 concurrent='1' 语义），finally DEL；
 * 日志口径：jobMessage = `{jobName} 总共耗时：{runMs}毫秒`（逐字对位 after）；异常捕获截 2000 写 exception_info。
 */
class TaskExecutor
{
    /** 执行任务并写日志；返回是否成功。任何异常不外抛（单个任务之死不拖垮调度进程）。 */
    public static function run(array $job): bool
    {
        $lockKey = TpConstant::PREFIX_JOB_LOCK . $job['job_id'];
        if (!RedisCache::setNx($lockKey, 1, 300)) {
            return false; // 已有同任务在跑（DisallowConcurrentExecution 近似）
        }
        $start = microtime(true);
        $startTime = date('Y-m-d H:i:s');
        $status = '0';
        $exceptionInfo = '';
        try {
            [$bean, $method, $params] = TargetParser::parse((string)$job['invoke_target']);
            $class = TaskRegistry::resolve($bean);
            $instance = new $class();
            if (!method_exists($instance, $method)) {
                throw new \BusinessException('目标方法不存在：' . $method);
            }
            // web 端 run 会把目标类的 echo 混进 HTTP 响应——输出缓冲捕获后丢弃（对位 System.out 进 catalina 日志不进响应）
            ob_start();
            try {
                $instance->$method(...$params);
            } finally {
                ob_end_clean();
            }
        } catch (\Throwable $e) {
            $status = '1';
            $exceptionInfo = mb_substr($e->getMessage(), 0, 2000);
        } finally {
            RedisCache::delete($lockKey, raw: true);
        }
        $runMs = (int)round((microtime(true) - $start) * 1000);
        JobLogService::insertJobLog([
            'job_name'       => $job['job_name'],
            'job_group'      => $job['job_group'],
            'invoke_target'  => $job['invoke_target'],
            'job_message'    => sprintf('%s 总共耗时：%d毫秒', $job['job_name'], $runMs),
            'status'         => $status,
            'exception_info' => $exceptionInfo,
            'start_time'     => $startTime,
            'end_time'       => date('Y-m-d H:i:s'),
        ]);
        return $status === '0';
    }
}
