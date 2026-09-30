<?php
declare(strict_types=1);

/**
 * 数据权限过滤（对位经典版 @DataScope AOP + ${params.dataScope} 占位）
 *
 * 按 sys_role.data_scope 注入 where：
 * 1 全部 2 自定义（sys_role_dept） 3 本部门 4 本部门及以下 5 仅本人
 */
final class DataScope
{
    /**
     * @param \think\db\Query $query    目标查询（须含 dept_alias 指定部门字段别名、user_alias 指定用户字段别名时传）
     * @param array      $user     会话用户（user_id/dept_id/roles[]）
     * @param string     $deptAlias部门表别名（默认空 = 主表）
     * @param string     $userAlias用户表别名
     */
    public static function apply($query, array $user, string $deptAlias = '', string $userAlias = ''): void
    {
        if (\app\service\PermissionService::isAdmin($user)) {
            return; // admin 不加过滤（对位经典版）
        }

        // 会话键双兼容：2.0.0 会话是驼峰 deptId/userId；直查行是下划线 dept_id/user_id
        $userDeptId = (int)($user['deptId'] ?? $user['dept_id'] ?? 0);
        $userUserId = (int)($user['userId'] ?? $user['user_id'] ?? 0);

        $deptCol = ($deptAlias !== '' ? $deptAlias . '.' : '') . 'dept_id';
        $userCol = ($userAlias !== '' ? $userAlias . '.' : '') . 'user_id';

        $conditions = [];
        $roleIdsForCustom = [];
        foreach ($user['roles'] ?? [] as $role) {
            match ((int)($role['data_scope'] ?? 1)) {
                1 => $conditions = [],           // 全部数据权限：清空即可（OR 语义下放行全部）
                2 => $roleIdsForCustom[] = (int)($role['roleId'] ?? $role['role_id'] ?? 0),
                3 => $conditions[] = [$deptCol, '=', $userDeptId],
                4 => $conditions[] = [$deptCol, 'in', self::deptAndChildrenIds($userDeptId)],
                5 => $conditions[] = [$userCol, '=', $userUserId],
                default => null,
            };
            if (empty($conditions) && empty($roleIdsForCustom)) {
                break; // 出现"全部"权限即不再叠加
            }
        }

        if ($roleIdsForCustom) {
            // 自定义：部门集合 = sys_role_dept 里这些角色的授权部门
            $deptIds = \think\facade\Db::table('sys_role_dept')
                ->where('role_id', 'in', $roleIdsForCustom)
                ->column('dept_id');
            if ($deptIds) {
                $conditions[] = [$deptCol, 'in', array_map('intval', $deptIds)];
            }
        }

        if ($conditions) {
            file_put_contents(runtime_path() . "dbg2.log", json_encode($conditions) . PHP_EOL, FILE_APPEND);
            // 多角色条件 OR 合并（对位经典版 dataScope 字符串 OR 拼接语义）
            // 条件形态：[列, '=', 值] / [列, 'in', 数组]——用显式 whereIn/where 而非 ...$cond 展开（think-orm 闭包内数组展开生成 0=1 的坑）
            $query->where(function ($q) use ($conditions) {
                foreach ($conditions as $i => $cond) {
                    [$col, $op, $val] = $cond;
                    if ($i === 0) {
                        strtolower($op) === 'in' ? $q->whereIn($col, $val) : $q->where($col, $op, $val);
                    } else {
                        strtolower($op) === 'in' ? $q->whereOr($col, 'in', $val) : $q->whereOr($col, $op, $val);
                    }
                }
            });
        }
    }

    /** 本部门及以下：ancestors 前缀匹配（sys_dept.ancestors 存 '0,100,101' 链） */
    private static function deptAndChildrenIds(int $deptId): array
    {
        $dept = \think\facade\Db::table('sys_dept')->where('dept_id', $deptId)->field('ancestors')->find();
        $ancestors = $dept['ancestors'] ?? '';
        $prefix = $ancestors . ',' . $deptId;
        return array_map('intval', \think\facade\Db::table('sys_dept')
            ->where(function ($q) use ($deptId, $prefix) {
                $q->whereOr('dept_id', $deptId)
                  ->whereOr(function ($q2) use ($prefix) {
                      $q2->whereLike('ancestors', $prefix . ',%')
                         ->whereOr('ancestors', $prefix);
                  });
            })
            ->column('dept_id'));
    }
}
