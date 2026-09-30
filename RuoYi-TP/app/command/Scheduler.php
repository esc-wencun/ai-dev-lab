<?php
declare(strict_types=1);

namespace app\command;

use app\service\CronService;
use app\service\JobService;
use app\task\TaskExecutor;
use RedisCache;
use TpConstant;
use think\console\Command;
use think\console\Input;
use think\console\Output;
use think\facade\Db;

/**
 * 定时任务常驻调度器（deviations #10 落地：自研循环替代 Quartz；qrtz_* 表零接触）
 *
 * 用法：另开终端前台常驻 `php think scheduler`（Windows 关窗即停；CLI 中文乱码先 chcp 65001）。
 * 主循环无状态每轮重算：status='0' 扫库 → 坏 cron 单任务跳过 → 基准回拨 2s 算下次 →
 * now 触发 → SETNX jobfire:{jobId}:{YmdHis} 幂等键（TTL 120s 防同 tick/双进程双跑）→ TaskExecutor；sleep(1)。
 * 不补跑：停机期间错过的触发点一律跳过（misfire 策略无对应物——deviations #10 说明）。
 */
class Scheduler extends Command
{
    protected function configure(): void
    {
        $this->setName('scheduler')
            ->setDescription('定时任务调度器（常驻前台进程；每秒扫描启用任务并按 cron 触发）');
    }

    protected function execute(Input $input, Output $output): void
    {
        $output->writeln('<info>[scheduler] started at ' . date('Y-m-d H:i:s') . '（Ctrl+C 停止；中文乱码先执行 chcp 65001）</info>');
        // Ctrl+C 可停（Windows PHP cli 无 pcntl，依赖默认 SIGINT 终止；循环体不吞信号即可）
        while (true) {
            try {
                $this->tick($output);
            } catch (\Throwable $e) {
                // 单轮之死不拖垮进程（DB 抖动等），告警后继续
                $output->writeln('<error>[scheduler] tick error: ' . $e->getMessage() . '</error>');
            }
            sleep(1);
        }
    }

    /** 单轮扫描（public 供测试钩子；每轮从库读全量启用任务——规避 Python 版 resume 空操作坑） */
    public function tick(Output $output): void
    {
        $jobs = Db::table('sys_job')->where('status', '0')->select()->toArray();
        $now = new \DateTimeImmutable('now');
        foreach ($jobs as $job) {
            $cron = (string)$job['cron_expression'];
            if (!CronService::isValid($cron)) {
                // 坏 cron 单任务跳过（经典版 init 遇坏 cron 整个启动失败——降级，deviations #10 说明）
                $output->writeln('<comment>[scheduler] bad cron skipped: job ' . $job['job_id'] . ' ' . $cron . '</comment>');
                continue;
            }
            // 基准回拨 2s 覆盖 tick 间隔 + 处理耗时
            $base = $now->modify('-2 seconds');
            $next = CronService::getNextRunDate($cron, \DateTime::createFromInterface($base));
            if ($next === null) {
                continue;
            }
            if ($next > $now) {
                continue; // 尚未到触发点
            }
            // 同一触发点幂等键：SETNX TTL 120s 防同 tick 重复 + 防误开双调度进程双跑
            $fireKey = TpConstant::PREFIX_JOB_FIRE . $job['job_id'] . ':' . $next->format('YmdHis');
            if (!RedisCache::setNx($fireKey, 1, 120)) {
                continue;
            }
            $output->writeln('[scheduler] fire job ' . $job['job_id'] . ' ' . $job['job_name'] . ' at ' . $next->format('H:i:s'));
            TaskExecutor::run($job);
        }
    }
}
