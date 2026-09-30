<?php
declare(strict_types=1);

namespace app\service;

use AppserverIo\Microcron\CronExpression;

/**
 * Cron 服务（对位经典版 CronUtils；底层 = appserver-io/microcron 2.0——秒级 Quartz 兼容，选型变更见 spec 实施记录）
 *
 * 方言转换层：Quartz 的 `?`（日/周互斥）→ `*`（microcron 无互斥语义，通配等价）；
 * 其余 Quartz 语法（6/7 位、秒级 n/s 步进、年字段）microcron 原生支持。
 * 周字段编号按 microcron（Linux 惯例 0/7=SUN、1=MON）——与 Quartz 的 1=SUN 错位，deviations #20 登记。
 */
class CronService
{
    /** 表达式有效性（对位 CronUtils.isValid；内部先做 ? → * 转换） */
    public static function isValid(string $cronExpression): bool
    {
        if (trim($cronExpression) === '') {
            return false;
        }
        try {
            self::factory($cronExpression);
            return true;
        } catch (\Throwable) {
            return false;
        }
    }

    /** 下一次执行时间（对位 getNextValidTimeAfter）；无效返回 null */
    public static function getNextRunDate(string $cronExpression, ?\DateTimeInterface $base = null): ?\DateTime
    {
        try {
            return self::factory($cronExpression)->getNextRunDate($base ?? 'now');
        } catch (\Throwable) {
            return null;
        }
    }

    /** 后 N 次执行时间（对位 getNextNCertificates...即 CronUtils.getNextNCronDates）；无效返回 [] */
    public static function getMultipleRunDates(string $cronExpression, int $count = 10): array
    {
        try {
            $dates = self::factory($cronExpression)->getMultipleRunDates($count, 'now');
            return array_map(fn(\DateTime $d) => $d->format('Y-m-d H:i:s'), is_array($dates) ? $dates : iterator_to_array($dates));
        } catch (\Throwable) {
            return [];
        }
    }

    /** 工厂（? → * 方言转换后 microcron factory） */
    private static function factory(string $cronExpression): CronExpression
    {
        return CronExpression::factory(str_replace('?', '*', trim($cronExpression)));
    }
}
