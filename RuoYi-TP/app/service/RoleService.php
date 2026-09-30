<?php
declare(strict_types=1);

namespace app\service;

use DataScope;
use think\facade\Db;

/**
 * 角色服务（最小只读；对位 SysRoleServiceImpl 的用户模块消费面）
 *
 * 仅提供用户模块三方法：selectRoleAll / selectRolesByUserId / checkRoleDataScope。
 * 完整角色管理（CRUD/数据权限弹窗/分配用户）5.0.0 收编接管。
 */
final class RoleService
{
    /** 全量角色（对位 selectRoleContactVo：join user_role/user/dept 链 + 非 admin DataScope('d')）；输出驼峰 */
    public static function selectRoleAll(array $user): array
    {
        $query = Db::table('sys_role r')
            ->leftJoin('sys_user_role ur', 'ur.role_id = r.role_id')
            ->leftJoin('sys_user u', 'u.user_id = ur.user_id')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->where('r.del_flag', '0')
            ->field('r.role_id,r.role_name,r.role_key,r.role_sort,r.data_scope,r.status,r.create_time')
            ->distinct(true);
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd');
        }
        $rows = $query->order('r.role_sort')->select()->toArray();
        return array_map(static fn(array $r): array => [
            'roleId'     => (int)$r['role_id'],
            'roleName'   => $r['role_name'],
            'roleKey'    => $r['role_key'],
            'roleSort'   => $r['role_sort'],
            'dataScope'  => $r['data_scope'],
            'status'     => $r['status'],
            'createTime' => $r['create_time'],
            'flag'       => false,
        ], $rows);
    }

    /** 全量角色 + 用户已有角色 flag=true 合并（对位 selectRolesByUserId）；$excludeAdmin 过滤 roleId=1（对位目标用户非 admin 场景） */
    public static function selectRolesByUserId(int $userId, array $user, bool $excludeAdmin = false): array
    {
        $roles = self::selectRoleAll($user);
        $owned = Db::table('sys_user_role')->where('user_id', $userId)->column('role_id');
        foreach ($roles as $i => $r) {
            if (in_array($r['roleId'], array_map('intval', $owned), true)) {
                $roles[$i]['flag'] = true;
            }
        }
        if ($excludeAdmin) {
            $roles = array_values(array_filter($roles, fn($r) => $r['roleId'] !== 1));
        }
        return $roles;
    }

    /** 角色数据权限防护（对位 checkRoleDataScope）：非 admin 逐个 roleId 校验，空数组直接放行 */
    public static function checkRoleDataScope(array $roleIds, array $user): void
    {
        if (PermissionService::isAdmin($user) || !$roleIds) {
            return;
        }
        foreach ($roleIds as $roleId) {
            $roleId = (int)$roleId;
            if ($roleId <= 0) {
                continue;
            }
            $query = Db::table('sys_role r')
                ->leftJoin('sys_user_role ur', 'ur.role_id = r.role_id')
                ->leftJoin('sys_user u', 'u.user_id = ur.user_id')
                ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
                ->where('r.role_id', $roleId)
                ->where('r.del_flag', '0');
            DataScope::apply($query, $user, 'd');
            if ($query->count() === 0) {
                throw new \BusinessException('没有权限访问角色数据！');
            }
        }
    }

    /* ================= 以下为 5.0.0 角色管理全量方法（收编本类） ================= */

    /** 分页列表查询构造器（对位 selectRoleContactVo：distinct 四表 join + 非 admin DataScope('d')）；返回 Query */
    public static function selectRoleList(array $filter, array $user): \think\db\Query
    {
        $query = Db::table('sys_role r')
            ->leftJoin('sys_user_role ur', 'ur.role_id = r.role_id')
            ->leftJoin('sys_user u', 'u.user_id = ur.user_id')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->where('r.del_flag', '0')
            ->field('r.role_id,r.role_name,r.role_key,r.role_sort,r.data_scope,r.status,r.del_flag,r.create_time,r.remark')
            ->distinct(true);
        if (($filter['roleName'] ?? '') !== '') {
            $query->whereLike('r.role_name', '%' . $filter['roleName'] . '%');
        }
        if (($filter['roleKey'] ?? '') !== '') {
            $query->whereLike('r.role_key', '%' . $filter['roleKey'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('r.status', $filter['status']);
        }
        $begin = (string)($filter['beginTime'] ?? '');
        $end = (string)($filter['endTime'] ?? '');
        if ($begin !== '') {
            $query->whereTime('r.create_time', '>=', date('Y-m-d 00:00:00', strtotime($begin)));
        }
        if ($end !== '') {
            $query->whereTime('r.create_time', '<=', date('Y-m-d 23:59:59', strtotime($end)));
        }
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd');
        }
        return $query;
    }

    /** 行 → 驼峰响应键（9 列，对位 selectRoleContactVo） */
    public static function toResponseRow(array $r): array
    {
        return [
            'roleId'     => (int)$r['role_id'],
            'roleName'   => $r['role_name'],
            'roleKey'    => $r['role_key'],
            'roleSort'   => $r['role_sort'],
            'dataScope'  => $r['data_scope'],
            'status'     => $r['status'],
            'delFlag'    => $r['del_flag'],
            'createTime' => $r['create_time'],
            'remark'     => $r['remark'] ?? '',
        ];
    }

    public static function selectRoleById(int $roleId): ?array
    {
        return Db::table('sys_role')->where('role_id', $roleId)->where('del_flag', '0')->find();
    }

    public static function checkRoleNameUnique(string $roleName, int $roleId = 0): bool
    {
        $row = Db::table('sys_role')->where('role_name', $roleName)->where('del_flag', '0')->find();
        return $row === null || (int)$row['role_id'] === $roleId;
    }

    public static function checkRoleKeyUnique(string $roleKey, int $roleId = 0): bool
    {
        $row = Db::table('sys_role')->where('role_key', $roleKey)->where('del_flag', '0')->find();
        return $row === null || (int)$row['role_id'] === $roleId;
    }

    /** admin 保护（对位 checkRoleAllowed） */
    public static function checkRoleAllowed(int $roleId): void
    {
        if ($roleId === 1) {
            throw new \BusinessException('不允许操作超级管理员角色');
        }
    }

    /** 新增（事务）：sys_role + role_menu（menuIds 空跳过也算成功） */
    public static function insertRole(array $role, array $menuIds, string $loginName): void
    {
        Db::startTrans();
        try {
            Db::table('sys_role')->insert($role);
            $roleId = (int)Db::table('sys_role')->max('role_id');
            if ($menuIds) {
                Db::table('sys_role_menu')->insertAll(array_map(fn($mid) => ['role_id' => $roleId, 'menu_id' => (int)$mid], $menuIds));
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 修改（事务）：update + role_menu 全量替换 */
    public static function updateRole(array $role, array $menuIds, string $loginName): void
    {
        Db::startTrans();
        try {
            $roleId = (int)$role['role_id'];
            Db::table('sys_role')->where('role_id', $roleId)->update($role);
            if ($menuIds) {
                Db::table('sys_role_menu')->where('role_id', $roleId)->delete();
                Db::table('sys_role_menu')->insertAll(array_map(fn($mid) => ['role_id' => $roleId, 'menu_id' => (int)$mid], $menuIds));
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 数据权限（事务）：update data_scope + role_dept 全量替换；成功后调用方刷会话 */
    public static function authDataScope(int $roleId, string $dataScope, array $deptIds, string $loginName): void
    {
        Db::startTrans();
        try {
            Db::table('sys_role')->where('role_id', $roleId)->update([
                'data_scope'  => $dataScope,
                'update_by'   => $loginName,
                'update_time' => date('Y-m-d H:i:s'),
            ]);
            Db::table('sys_role_dept')->where('role_id', $roleId)->delete();
            if ($deptIds) {
                Db::table('sys_role_dept')->insertAll(array_map(fn($did) => ['role_id' => $roleId, 'dept_id' => (int)$did], $deptIds));
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 已分配用户数（对位 countUserRoleByRoleId） */
    public static function countUserRoleByRoleId(int $roleId): int
    {
        return Db::table('sys_user_role')->where('role_id', $roleId)->count();
    }

    /** 批量删除（事务）：role_menu/role_dept 物理删 + 软删（调用方已逐个防护） */
    public static function deleteRoleByIds(array $ids): int
    {
        Db::startTrans();
        try {
            Db::table('sys_role_menu')->whereIn('role_id', $ids)->delete();
            Db::table('sys_role_dept')->whereIn('role_id', $ids)->delete();
            $rows = Db::table('sys_role')->whereIn('role_id', $ids)->update(['del_flag' => '2']);
            Db::commit();
            return (int)$rows;
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 状态切换（动态 set 仅 status） */
    public static function changeStatus(int $roleId, string $status, string $loginName): int
    {
        return Db::table('sys_role')->where('role_id', $roleId)->update(['status' => $status, 'update_time' => date('Y-m-d H:i:s')]);
    }

    /** 已分配用户列表构造器（对位 selectAllocatedList：join role 限定 + DataScope d,u）；返回 Query */
    public static function selectAllocatedList(array $filter, int $roleId, array $user): \think\db\Query
    {
        $query = Db::table('sys_user u')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->leftJoin('sys_user_role ur', 'ur.user_id = u.user_id')
            ->leftJoin('sys_role r', 'r.role_id = ur.role_id')
            ->where('u.del_flag', '0')
            ->where('r.role_id', $roleId)
            ->field('u.user_id,u.dept_id,u.login_name,u.user_name,u.user_type,u.email,u.avatar,u.phonenumber,u.status,u.create_time')
            ->distinct(true);
        if (($filter['loginName'] ?? '') !== '') {
            $query->whereLike('u.login_name', '%' . $filter['loginName'] . '%');
        }
        if (($filter['phonenumber'] ?? '') !== '') {
            $query->whereLike('u.phonenumber', '%' . $filter['phonenumber'] . '%');
        }
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd', 'u');
        }
        return $query;
    }

    /** 未分配用户列表构造器（对位 selectUnallocatedList）；返回 Query */
    public static function selectUnallocatedList(array $filter, int $roleId, array $user): \think\db\Query
    {
        $query = Db::table('sys_user u')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->leftJoin('sys_user_role ur', 'ur.user_id = u.user_id')
            ->leftJoin('sys_role r', 'r.role_id = ur.role_id')
            ->where('u.del_flag', '0')
            ->where(function ($q) use ($roleId) {
                $q->where('r.role_id', '<>', $roleId)->whereOr('r.role_id', null);
            })
            ->whereNotIn('u.user_id', function ($q) use ($roleId) {
                $q->name('sys_user_role')->where('role_id', $roleId)->field('user_id');
            })
            ->field('u.user_id,u.dept_id,u.login_name,u.user_name,u.user_type,u.email,u.avatar,u.phonenumber,u.status,u.create_time')
            ->distinct(true);
        if (($filter['loginName'] ?? '') !== '') {
            $query->whereLike('u.login_name', '%' . $filter['loginName'] . '%');
        }
        if (($filter['phonenumber'] ?? '') !== '') {
            $query->whereLike('u.phonenumber', '%' . $filter['phonenumber'] . '%');
        }
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd', 'u');
        }
        return $query;
    }

    /** 用户行 → 驼峰（allocated/unallocated 10 列） */
    public static function toUserResponseRow(array $r): array
    {
        return [
            'userId'      => (int)$r['user_id'],
            'deptId'      => $r['dept_id'] !== null ? (int)$r['dept_id'] : null,
            'loginName'   => $r['login_name'],
            'userName'    => $r['user_name'],
            'userType'    => $r['user_type'],
            'email'       => $r['email'],
            'avatar'      => $r['avatar'],
            'phonenumber' => $r['phonenumber'],
            'status'      => $r['status'],
            'createTime'  => $r['create_time'],
        ];
    }

    /** 角色的 deptTreeData（对位 selectDeptList + roleDeptList 整串勾选）；带数据权限 */
    public static function deptTreeData(?int $roleId, array $user): array
    {
        $depts = DeptService::selectDeptList(['status' => '0'], $user);
        $roleDeptList = [];
        if ($roleId !== null && $roleId > 0) {
            $rows = Db::table('sys_role_dept rd')
                ->join('sys_dept d', 'd.dept_id = rd.dept_id')
                ->where('rd.role_id', $roleId)
                ->where('d.del_flag', '0')
                ->field('rd.dept_id,d.dept_name')
                ->select()->toArray();
            $roleDeptList = array_map(fn($r) => $r['dept_id'] . $r['dept_name'], $rows);
        }
        $tree = [];
        foreach ($depts as $d) {
            $deptId = (int)$d['dept_id'];
            $name = $d['dept_name'];
            $tree[] = [
                'id'      => $deptId,
                'pId'     => (int)$d['parent_id'],
                'name'    => $name,
                'title'   => $name,
                'checked' => $roleId !== null && in_array($deptId . $name, $roleDeptList, true),
                'open'    => false,
                'nocheck' => false,
            ];
        }
        return $tree;
    }
}
