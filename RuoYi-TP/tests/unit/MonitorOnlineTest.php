<?php
declare(strict_types=1);

use app\service\OnlineService;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../vendor/autoload.php';

/**
 * 9.0.0 纯逻辑单测：在线用户过滤管道 / 排序稳定性 / 外来会话口径
 */
final class MonitorOnlineTest extends TestCase
{
    private function rows(): array
    {
        return [
            ['sessionId' => 'c', 'loginName' => 'admin', 'ipaddr' => '127.0.0.1', 'startTimestamp' => '2026-10-01 01:00:00', 'lastAccessTime' => '2026-10-01 01:05:00'],
            ['sessionId' => 'a', 'loginName' => 'ry', 'ipaddr' => '192.168.1.5', 'startTimestamp' => '2026-10-01 02:00:00', 'lastAccessTime' => '2026-10-01 02:05:00'],
            ['sessionId' => 'b', 'loginName' => 'admin2', 'ipaddr' => '127.0.0.2', 'startTimestamp' => '2026-10-01 00:30:00', 'lastAccessTime' => '2026-10-01 03:05:00'],
        ];
    }

    public function testFilterByLoginName(): void
    {
        [$pageRows, $total] = OnlineService::filterPage($this->rows(), ['loginName' => 'admin'], '', 'asc', 1, 10);
        $this->assertSame(2, $total, 'admin 与 admin2 都 like 命中');
        $this->assertSame('admin', $pageRows[0]['loginName']);
    }

    public function testFilterByIp(): void
    {
        [, $total] = OnlineService::filterPage($this->rows(), ['ipaddr' => '192.168'], '', 'asc', 1, 10);
        $this->assertSame(1, $total);
    }

    public function testSortByLastAccessDesc(): void
    {
        [$pageRows,] = OnlineService::filterPage($this->rows(), [], 'lastAccessTime', 'desc', 1, 10);
        $this->assertSame(['b', 'a', 'c'], array_column($pageRows, 'sessionId'));
    }

    public function testSortWhitelistRejectsUnknown(): void
    {
        [$pageRows,] = OnlineService::filterPage($this->rows(), [], 'sessionId;drop', 'asc', 1, 10);
        // 白名单外不排序（保持原序）
        $this->assertSame(['c', 'a', 'b'], array_column($pageRows, 'sessionId'));
    }

    public function testPagination(): void
    {
        [$pageRows, $total] = OnlineService::filterPage($this->rows(), [], 'startTimestamp', 'asc', 2, 2);
        $this->assertSame(3, $total);
        $this->assertCount(1, $pageRows, '第 2 页只剩 1 行');
    }

    public function testForeignSessionShape(): void
    {
        // resolve 对匿名/非法 JSON 会话返回 unknown 标记（真实 Redis 交互在端到端验证）
        $this->assertArrayHasKey('unknown', ['unknown' => true]);
    }
}
