<?php
declare(strict_types=1);

namespace app\task;

/**
 * 任务目标注册表（deviations #12 落地：bean 名 → 类名白名单映射，替代经典版包名前缀白名单）
 *
 * 新增可调度任务类：在此注册 bean 名（_invoke_target 里 `.` 前的第一段）。
 */
class TaskRegistry
{
    /** bean 名 → 类名（1.0.0 起集中管理；未注册 = 不在白名单） */
    private const BEANS = [
        'ryTask' => \app\task\RyTask::class,
    ];

    /** 解析 bean 名（invokeTarget 第一个 `(` 前、最后一个 `.` 前的串） */
    public static function beanOf(string $invokeTarget): string
    {
        $head = $invokeTarget;
        $paren = strpos($head, '(');
        if ($paren !== false) {
            $head = substr($head, 0, $paren);
        }
        $dot = strrpos($head, '.');
        return $dot === false ? $head : substr($head, 0, $dot);
    }

    /** 注册表查名（校验链 ⑥ 白名单落地）；未注册/类不存在抛 BusinessException */
    public static function resolve(string $bean): string
    {
        $class = self::BEANS[$bean] ?? null;
        if ($class === null || !class_exists($class)) {
            throw new \BusinessException('目标字符串不在白名单内');
        }
        return $class;
    }

    /** 全部注册名（校验链测试/调试用） */
    public static function names(): array
    {
        return array_keys(self::BEANS);
    }
}
