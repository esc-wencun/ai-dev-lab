<?php
declare(strict_types=1);

namespace app\service;

/**
 * 权限服务：用户权限集合计算与判定
 *
 * 对位经典版 UserRealm/PermissionService：admin 全量通过；普通用户取角色关联菜单的 perms 并集。
 * 经典版 perms 数据特征：C 型菜单（如 system:user:view）与 F 型按钮（如 system:user:list）均在 sys_menu.perms。
 */
final class PermissionService
{
    /**
     * 计算用户 permissions 集合（登录时调用一次存会话）
     *
     * @param array $user 含 user_id / login_name
     * @return string[] 排序去重后的权限串数组
     */
    public static function permissionsOf(array $user): array
    {
        if (self::isAdmin($user)) {
            $rows = \think\facade\Db::table('sys_menu')
                ->where('menu_type', 'in', ['M', 'C', 'F'])
                ->where('perms', '<>', '')
                ->whereNotNull('perms')
                ->field('perms')
                ->select()->toArray();
        } else {
            $rows = \think\facade\Db::table('sys_menu m')
                ->join('sys_role_menu rm', 'rm.menu_id = m.menu_id')
                ->join('sys_user_role ur', 'ur.role_id = rm.role_id')
                ->join('sys_role r', 'r.role_id = ur.role_id')
                ->where('ur.user_id', $user['user_id'])
                ->where('r.status', '0')
                ->where('r.del_flag', '0')
                ->where('m.menu_type', 'in', ['M', 'C', 'F'])
                ->where('m.perms', '<>', '')
                ->whereNotNull('m.perms')
                ->distinct(true)
                ->field('m.perms')
                ->select()->toArray();
        }
        $perms = array_unique(array_column($rows, 'perms'));
        sort($perms);
        return array_values($perms);
    }

    public static function isAdmin(array $user): bool
    {
        // 经典版：user_id=1 为 admin（SysUser.isAdmin 判据）
        return ($user['user_id'] ?? 0) == 1;
    }

    /** 会话权限判定（对位 PermissionService.hasPermi） */
    public static function hasPerm(array $sessionData, string $perm): bool
    {
        if (self::isAdmin($sessionData)) {
            return true;
        }
        return in_array($perm, $sessionData['permissions'] ?? [], true);
    }
}
