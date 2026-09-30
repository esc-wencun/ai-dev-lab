<?php
declare(strict_types=1);

namespace app\task;

/**
 * 示例任务（对位经典版 com.ruoyi.quartz.task.RyTask——预置 3 任务的调用目标）
 *
 * echo 对位 System.out.println（调度进程 CLI 输出；web 端 run 触发时输出进 web 进程 stdout）。
 */
class RyTask
{
    public function ryNoParams(): void
    {
        echo 'ryTask.ryNoParams 执行成功，无参数' . PHP_EOL;
    }

    public function ryParams(string $params): void
    {
        echo 'ryTask.ryParams 执行成功，字符串参数：' . $params . PHP_EOL;
    }

    public function ryMultipleParams(string $s, bool $b, int $l, float $d, int $i): void
    {
        echo 'ryTask.ryMultipleParams 执行成功，参数：' . $s . ', ' . var_export($b, true) . ', ' . $l . ', ' . $d . ', ' . $i . PHP_EOL;
    }
}
