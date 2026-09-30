<?php
declare(strict_types=1);

namespace app\service;

use think\facade\Db;

/**
 * 菜单服务：用户菜单树（对位经典版 SysMenuServiceImpl.selectMenusByUser + getChildPerms）
 *
 * admin 全量（menu_type in M/C 且 visible=0）；普通用户经 角色-菜单 关联去重。
 * 树构建对位 getChildPerms(list, 0)：parent_id 递归、order_num/parent_id 排序。
 */
final class MenuService
{
    /** @return array 嵌套树：每节点含 menu_id/menu_name/url/target/icon/is_refresh/menu_type/children */
    public static function menusOf(array $user): array
    {
        if (PermissionService::isAdmin($user)) {
            $rows = Db::table('sys_menu')
                ->where('menu_type', 'in', ['M', 'C'])
                ->where('visible', '0')
                ->field('menu_id,menu_name,parent_id,order_num,url,target,menu_type,visible,is_refresh,icon')
                ->order('parent_id,order_num')
                ->select()->toArray();
        } else {
            $rows = Db::table('sys_menu m')
                ->join('sys_role_menu rm', 'rm.menu_id = m.menu_id')
                ->join('sys_user_role ur', 'ur.role_id = rm.role_id')
                ->join('sys_role r', 'r.role_id = ur.role_id')
                ->where('ur.user_id', (int)($user['user_id'] ?? 0))
                ->where('r.status', '0')
                ->where('r.del_flag', '0')
                ->where('m.menu_type', 'in', ['M', 'C'])
                ->where('m.visible', '0')
                ->field('m.menu_id,m.menu_name,m.parent_id,m.order_num,m.url,m.target,m.menu_type,m.visible,m.is_refresh,m.icon')
                ->distinct(true)
                ->order('m.parent_id,m.order_num')
                ->select()->toArray();
        }
        return self::buildTree($rows, 0);
    }

    /**
     * 树构建（对位经典版 getChildPerms；PHPUnit 固化纯逻辑）
     *
     * @param array $rows 扁平行（内部按 order_num,menu_id 稳定排序，不依赖输入顺序）
     * @param int   $parentId 根 parent_id
     */
    public static function buildTree(array $rows, int $parentId = 0): array
    {
        usort($rows, fn($a, $b) => [$a['order_num'] ?? 0, $a['menu_id']] <=> [$b['order_num'] ?? 0, $b['menu_id']]);
        $tree = [];
        foreach ($rows as $row) {
            if ((int)$row['parent_id'] === $parentId) {
                $row['children'] = self::buildTree($rows, (int)$row['menu_id']);
                $tree[] = $row;
            }
        }
        return $tree;
    }

    /* ================= 以下为 5.0.0 菜单管理方法（收编本类） ================= */

    private const MENU_ALL_FIELDS = 'menu_id,menu_name,parent_id,order_num,url,target,menu_type,visible,is_refresh,perms,icon,create_by,create_time';

    /** 菜单管理列表（对位 selectMenuList：admin 全量 / 非 admin 按角色关联；不筛 menu_type/visible）；返回驼峰行 */
    public static function selectMenuList(array $filter, array $user): array
    {
        if (PermissionService::isAdmin($user)) {
            $query = Db::table('sys_menu m');
        } else {
            $query = Db::table('sys_menu m')
                ->join('sys_role_menu rm', 'rm.menu_id = m.menu_id')
                ->join('sys_user_role ur', 'ur.role_id = rm.role_id')
                ->join('sys_role r', 'r.role_id = ur.role_id')
                ->where('ur.user_id', (int)($user['userId'] ?? $user['user_id'] ?? 0))
                ->where('r.status', '0')
                ->where('r.del_flag', '0')
                ->distinct(true);
        }
        if (($filter['menuName'] ?? '') !== '') {
            $query->whereLike('m.menu_name', '%' . $filter['menuName'] . '%');
        }
        if (($filter['visible'] ?? '') !== '') {
            $query->where('m.visible', $filter['visible']);
        }
        $rows = $query->field(self::MENU_ALL_FIELDS)->order('m.parent_id,m.order_num')->select()->toArray();
        return array_map([self::class, 'toResponseRow'], $rows);
    }

    /** 菜单行 → 驼峰（13 列，对位 selectMenuVo；perms ifnull '' 字符串） */
    public static function toResponseRow(array $r): array
    {
        return [
            'menuId'     => (int)$r['menu_id'],
            'menuName'   => $r['menu_name'],
            'parentId'   => (int)$r['parent_id'],
            'orderNum'   => (int)$r['order_num'],
            'url'        => $r['url'],
            'target'     => $r['target'],
            'menuType'   => $r['menu_type'],
            'visible'    => $r['visible'],
            'isRefresh'  => $r['is_refresh'],
            'perms'      => $r['perms'] ?? '',
            'icon'       => $r['icon'],
            'createBy'   => $r['create_by'] ?? '',
            'createTime' => $r['create_time'] ?? null,
        ];
    }

    /** 全量菜单（对位 selectMenuAll：admin 全量 / 非 admin 角色关联；**含 F 型与隐藏**）；供 Ztree 组装 */
    private static function selectMenuAll(array $user): array
    {
        if (PermissionService::isAdmin($user)) {
            return Db::table('sys_menu')->field(self::MENU_ALL_FIELDS)->order('parent_id,order_num')->select()->toArray();
        }
        return Db::table('sys_menu m')
            ->join('sys_role_menu rm', 'rm.menu_id = m.menu_id')
            ->join('sys_user_role ur', 'ur.role_id = rm.role_id')
            ->join('sys_role r', 'r.role_id = ur.role_id')
            ->where('ur.user_id', (int)($user['userId'] ?? $user['user_id'] ?? 0))
            ->where('r.status', '0')
            ->where('r.del_flag', '0')
            ->field('m.' . str_replace(',', ',m.', self::MENU_ALL_FIELDS))
            ->distinct(true)
            ->order('m.parent_id,m.order_num')
            ->select()->toArray();
    }

    /** 角色的菜单权限 Ztree（对位 roleMenuTreeData：checked = concat(menu_id,perms) 整串相等；name 带 perms 灰字） */
    public static function roleMenuTreeData(?int $roleId, array $user): array
    {
        $menus = self::selectMenuAll($user);
        $roleMenuList = [];
        if ($roleId !== null && $roleId > 0) {
            $rows = Db::table('sys_role_menu rm')
                ->join('sys_menu m', 'm.menu_id = rm.menu_id')
                ->where('rm.role_id', $roleId)
                ->field("CONCAT(rm.menu_id, IFNULL(m.perms, '')) AS combined")
                ->select()->toArray();
            $roleMenuList = array_column($rows, 'combined');
        }
        $tree = [];
        foreach ($menus as $m) {
            $menuId = (int)$m['menu_id'];
            $perms = (string)($m['perms'] ?? '');
            $name = $m['menu_name'] . '<font color="#888">&nbsp;&nbsp;&nbsp;' . $perms . '</font>';
            $tree[] = [
                'id'      => $menuId,
                'pId'     => (int)$m['parent_id'],
                'name'    => $name,
                'title'   => $m['menu_name'],
                'checked' => $roleId !== null && in_array($menuId . $perms, $roleMenuList, true),
                'open'    => false,
                'nocheck' => false,
            ];
        }
        return $tree;
    }

    /** 菜单 Ztree（对位 menuTreeData：无 perms 后缀、全不勾） */
    public static function menuTreeData(array $user): array
    {
        $menus = self::selectMenuAll($user);
        $tree = [];
        foreach ($menus as $m) {
            $tree[] = [
                'id'      => (int)$m['menu_id'],
                'pId'     => (int)$m['parent_id'],
                'name'    => $m['menu_name'],
                'title'   => $m['menu_name'],
                'checked' => false,
                'open'    => false,
                'nocheck' => false,
            ];
        }
        return $tree;
    }

    public static function selectMenuById(int $menuId): ?array
    {
        return Db::table('sys_menu m')
            ->leftJoin('sys_menu p', 'p.menu_id = m.parent_id')
            ->where('m.menu_id', $menuId)
            ->field('m.*, p.menu_name as parent_name')
            ->find();
    }

    /** 同父下同名唯一（sys_menu 无 del_flag，全量口径） */
    public static function checkMenuNameUnique(string $menuName, int $parentId, int $menuId = 0): bool
    {
        $row = Db::table('sys_menu')->where('menu_name', $menuName)->where('parent_id', $parentId)->find();
        return $row === null || (int)$row['menu_id'] === $menuId;
    }

    public static function selectCountMenuByParentId(int $parentId): int
    {
        return Db::table('sys_menu')->where('parent_id', $parentId)->count();
    }

    public static function selectCountRoleMenuByMenuId(int $menuId): int
    {
        return Db::table('sys_role_menu')->where('menu_id', $menuId)->count();
    }

    public static function insertMenu(array $menu, string $loginName): void
    {
        $menu['create_by'] = $loginName;
        $menu['create_time'] = date('Y-m-d H:i:s');
        Db::table('sys_menu')->insert($menu);
    }

    /** 修改（动态 set，对位 mapper if 列表） */
    public static function updateMenu(array $menu, string $loginName): void
    {
        $set = array_filter([
            'menu_name'  => $menu['menu_name'] ?? null,
            'parent_id'  => $menu['parent_id'] ?? null,
            'order_num'  => $menu['order_num'] ?? null,
            'url'        => $menu['url'] ?? null,
            'target'     => $menu['target'] ?? null,
            'menu_type'  => $menu['menu_type'] ?? null,
            'visible'    => $menu['visible'] ?? null,
            'is_refresh' => $menu['is_refresh'] ?? null,
            'perms'      => $menu['perms'] ?? null,
            'icon'       => $menu['icon'] ?? null,
            'update_by'  => $loginName,
            'update_time' => date('Y-m-d H:i:s'),
        ], fn($v) => $v !== null);
        Db::table('sys_menu')->where('menu_id', (int)$menu['menu_id'])->update($set);
    }

    /** 批量排序（事务；异常 → 「保存排序异常，请联系管理员」） */
    public static function updateMenuSort(array $menuIds, array $orderNums): void
    {
        Db::startTrans();
        try {
            foreach ($menuIds as $i => $id) {
                Db::table('sys_menu')->where('menu_id', (int)$id)->update(['order_num' => (int)$orderNums[$i]]);
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw new \BusinessException('保存排序异常，请联系管理员');
        }
    }

    /** 物理删（`or parent_id=?` 连带删一级子菜单为经典原样兜底） */
    public static function deleteMenuById(int $menuId): int
    {
        return Db::table('sys_menu')->where(function ($q) use ($menuId) {
            $q->whereOr('menu_id', $menuId)->whereOr('parent_id', $menuId);
        })->delete();
    }
}
