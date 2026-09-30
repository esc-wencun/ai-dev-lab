<?php
declare(strict_types=1);

use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../app/common/PageQuery.php';

/**
 * PageQuery 纯逻辑单测：驼峰转换 / 白名单 / 边界值
 */
final class PageQueryTest extends TestCase
{
    public function testCamelToSnake(): void
    {
        $this->assertSame('create_time', PageQuery::camelToSnake('createTime'));
        $this->assertSame('user_name', PageQuery::camelToSnake('userName'));
        $this->assertSame('login_name', PageQuery::camelToSnake('loginName'));
        $this->assertSame('dept_id', PageQuery::camelToSnake('deptId'));
    }

    public function testCamelToSnakeKeepsSnakeAsIs(): void
    {
        $this->assertSame('create_time', PageQuery::camelToSnake('create_time'));
        $this->assertSame('oper_id', PageQuery::camelToSnake('oper_id'));
    }

    public function testDefaults(): void
    {
        $q = PageQuery::from([]);
        $this->assertSame(1, $q->pageNum);
        $this->assertSame(10, $q->pageSize);
        $this->assertNull($q->orderBy);
        $this->assertNull($q->beginTime);
    }

    public function testWhitelistAcceptsAndConverts(): void
    {
        $q = PageQuery::from([
            'pageNum' => 2, 'pageSize' => 25,
            'orderByColumn' => 'createTime', 'isAsc' => 'desc',
        ], ['create_time', 'login_name']);
        $this->assertSame(2, $q->pageNum);
        $this->assertSame(25, $q->pageSize);
        $this->assertSame('create_time', $q->orderBy);
        $this->assertSame('desc', $q->isAsc);
    }

    public function testWhitelistRejectsUnknownField(): void
    {
        // 白名单外字段静默忽略（orderBy 置 null），杜绝 ORDER BY 注入
        $q = PageQuery::from([
            'orderByColumn' => '(SELECT 1)', 'isAsc' => 'desc',
        ], ['create_time']);
        $this->assertNull($q->orderBy);
    }

    public function testPageSizeClamped(): void
    {
        $q = PageQuery::from(['pageSize' => 999]);
        $this->assertSame(100, $q->pageSize);
        $q2 = PageQuery::from(['pageSize' => 0]);
        $this->assertSame(1, $q2->pageSize);
    }

    public function testTimeRangeParams(): void
    {
        $q = PageQuery::from(['params' => ['beginTime' => '2026-01-01', 'endTime' => '2026-09-29']]);
        $this->assertSame('2026-01-01', $q->beginTime);
        $this->assertSame('2026-09-29', $q->endTime);
        $q2 = PageQuery::from(['params' => ['beginTime' => '', 'endTime' => null]]);
        $this->assertNull($q2->beginTime);
    }
}
