<?php
declare(strict_types=1);

use app\service\CronService;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 10.0.0 纯逻辑单测：cron 方言（microcron + ? → * 转换层）——Task 1 实测固化
 * 选型变更实锤：dragonmantank v3 不支持秒级（5 位分钟制）→ 换 microcron（用户拍板）
 */
final class CronServiceTest extends TestCase
{
    /* ---------- 预置 3 任务表达式（Quartz 6 位秒级 + n/s 步进 + ? 通配） ---------- */

    public function testPresetThreeJobsValid(): void
    {
        $this->assertTrue(CronService::isValid('0/10 * * * * ?'));
        $this->assertTrue(CronService::isValid('0/15 * * * * ?'));
        $this->assertTrue(CronService::isValid('0/20 * * * * ?'));
    }

    public function testInvalidExpressions(): void
    {
        $this->assertFalse(CronService::isValid('* * *'));
        $this->assertFalse(CronService::isValid(''));
        $this->assertFalse(CronService::isValid('not a cron'));
    }

    public function testSevenFieldWithYearSupported(): void
    {
        // microcron 支持年字段（比 spec 预判的 v3 更兼容——spec 差异表按实测回填）
        $this->assertTrue(CronService::isValid('0/10 * * * * ? 2026'));
        $this->assertTrue(CronService::isValid('0/10 * * * * ? 2099'));
    }

    public function testNextRunDate(): void
    {
        $next = CronService::getNextRunDate('0/15 * * * * ?');
        $this->assertNotNull($next);
        $this->assertSame(0, (int)$next->format('s') % 15, '秒步进 15 的倍数');
        // 坏表达式 null
        $this->assertNull(CronService::getNextRunDate('* * *'));
    }

    public function testMultipleRunDates(): void
    {
        $dates = CronService::getMultipleRunDates('0/20 * * * * ?', 10);
        $this->assertCount(10, $dates);
        foreach ($dates as $d) {
            $this->assertMatchesRegularExpression('/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/', $d);
            $this->assertSame(0, (int)(new DateTime($d))->format('s') % 20);
        }
        $this->assertSame([], CronService::getMultipleRunDates('* * *', 3));
    }

    public function testWildcardQuestionMark(): void
    {
        // ? 出现在 6 位的日位与周位（Quartz 互斥符）均转 * 处理；5 位日位 ? 非法（非 Quartz 形态，原样拒绝）
        $this->assertTrue(CronService::isValid('? * * * * ?'));
        $this->assertTrue(CronService::isValid('0/10 * ? * * ?'));
        $this->assertFalse(CronService::isValid('0/10 * ? * *'));
    }

    /* ---------- 周字段编号（microcron = Linux 0/7=SUN、1=MON——与 Quartz 1=SUN 错位，deviations #20） ---------- */

    public function testWeekdayNumberingIsLinuxConvention(): void
    {
        // 从当前时刻找 week=1 的下一次触发：必须是周一（Mon）而非 Quartz 语义的周日
        $next = CronService::getNextRunDate('* * * * * 1');
        $this->assertNotNull($next);
        $this->assertSame('Mon', $next->format('D'), 'week=1 → 周一（Linux 语义）');
        $next2 = CronService::getNextRunDate('* * * * * 7');
        $this->assertSame('Sun', $next2->format('D'), 'week=7 → 周日');
    }
}
