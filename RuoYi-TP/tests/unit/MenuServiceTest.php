<?php
declare(strict_types=1);

namespace app\service;

use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/../../app/service/MenuService.php';

/**
 * MenuService 树构建纯逻辑单测（对位经典版 getChildPerms）
 */
final class MenuServiceTest extends TestCase
{
    private function row(int $id, int $parentId, string $name, int $order = 0): array
    {
        return ['menu_id' => $id, 'parent_id' => $parentId, 'menu_name' => $name, 'order_num' => $order, 'url' => '#', 'children' => []];
    }

    public function testBuildsNestedTree(): void
    {
        $rows = [
            $this->row(1, 0, '系统管理'),
            $this->row(100, 1, '用户管理'),
            $this->row(2, 0, '系统监控'),
        ];
        $tree = MenuService::buildTree($rows, 0);
        $this->assertCount(2, $tree);
        $this->assertSame('系统管理', $tree[0]['menu_name']);
        $this->assertCount(1, $tree[0]['children']);
        $this->assertSame('用户管理', $tree[0]['children'][0]['menu_name']);
        $this->assertSame([], $tree[1]['children']);
    }

    public function testSortsChildrenByOrderNum(): void
    {
        $rows = [
            $this->row(1, 0, '系统管理'),
            $this->row(102, 1, '菜单管理', 3),
            $this->row(100, 1, '用户管理', 1),
            $this->row(101, 1, '角色管理', 2),
        ];
        $tree = MenuService::buildTree($rows, 0);
        $children = $tree[0]['children'];
        $this->assertSame(['用户管理', '角色管理', '菜单管理'], array_column($children, 'menu_name'));
    }

    public function testOrphanRowsIgnored(): void
    {
        // parent_id 指向不存在节点的行不进树（对位经典版 getChildPerms 行为）
        $rows = [
            $this->row(1, 0, '系统管理'),
            $this->row(999, 888, '孤儿菜单'),
        ];
        $tree = MenuService::buildTree($rows, 0);
        $this->assertCount(1, $tree);
        $this->assertSame('系统管理', $tree[0]['menu_name']);
    }

    public function testEmptyRowsReturnEmpty(): void
    {
        $this->assertSame([], MenuService::buildTree([], 0));
    }

    public function testDeepNesting(): void
    {
        $rows = [
            $this->row(1, 0, 'L0'),
            $this->row(2, 1, 'L1'),
            $this->row(3, 2, 'L2'),
        ];
        $tree = MenuService::buildTree($rows, 0);
        $this->assertSame('L2', $tree[0]['children'][0]['children'][0]['menu_name']);
    }
}
