<?php
declare(strict_types=1);

namespace app\service;

use app\model\SysDept;
use DataScope;
use think\facade\Db;

/**
 * 部门服务（对位经典版 SysDeptServiceImpl）
 *
 * 数据权限三处：selectDeptList / selectDeptTreeData / checkDeptDataScope。
 * ancestors 级联用前缀锚定替换（spec 特殊行为 3，实现加固非行为差异）。
 */
final class DeptService
{
    private const SELECT_FIELDS = 'dept_id,parent_id,ancestors,dept_name,order_num,leader,phone,email,status,del_flag,create_by,create_time,update_by,update_time';

    /** 列表查询（对位 selectDeptList；@DataScope(deptAlias="d")） */
    public static function selectDeptList(array $filter, array $user): array
    {
        $query = Db::table('sys_dept d')
            ->where('d.del_flag', '0')
            ->field('d.' . str_replace(',', ',d.', self::SELECT_FIELDS));
        if (($filter['deptId'] ?? null) !== null) {
            $query->where('d.dept_id', (int)$filter['deptId']);
        }
        if (($filter['parentId'] ?? null) !== null) {
            $query->where('d.parent_id', (int)$filter['parentId']);
        }
        if (($filter['deptName'] ?? '') !== '') {
            $query->whereLike('d.dept_name', '%' . $filter['deptName'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('d.status', $filter['status']);
        }
        // 对位经典版 ${params.dataScope} 占位：admin 不加条件
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd');
        }
        return $query->order('d.parent_id,d.order_num')->select()->toArray();
    }

    /** 单条（含 parentName 子查询；deptId=100 时 parentName 置「无」——经典版硬编码原样） */
    public static function selectDeptById(int $deptId): ?array
    {
        $row = Db::table('sys_dept d')
            ->leftJoin('sys_dept p', 'p.dept_id = d.parent_id')
            ->where('d.dept_id', $deptId)
            ->field('d.*, p.dept_name as parent_name')
            ->find();
        if ($row && $deptId === 100) {
            $row['parent_name'] = '无';
        }
        return $row ?: null;
    }

    /**
     * Ztree 树数据（对位 selectDeptTreeExcludeChild + buildDeptTreeSelect）：
     * 只含 status='0'；excludeId>0 排除自身及 ancestors 含 excludeId 的后代；带数据权限。
     * 输出 {id, pId, name, title, checked:false, open:false, nocheck:false}——pId 键名精确。
     */
    public static function selectDeptTreeData(int $excludeId, array $user): array
    {
        $query = Db::table('sys_dept d')
            ->where('d.del_flag', '0')
            ->where('d.status', '0')
            ->field('d.dept_id,d.parent_id,d.ancestors,d.dept_name,d.order_num');
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd');
        }
        $rows = $query->order('d.parent_id,d.order_num')->select()->toArray();

        $tree = [];
        foreach ($rows as $row) {
            $deptId = (int)$row['dept_id'];
            if ($excludeId > 0 && ($deptId === $excludeId
                || ($row['ancestors'] !== null && str_contains(',' . $row['ancestors'] . ',', ',' . $excludeId . ',')))) {
                continue;
            }
            $tree[] = [
                'id'      => $deptId,
                'pId'     => (int)$row['parent_id'],
                'name'    => $row['dept_name'],
                'title'   => $row['dept_name'],
                'checked' => false,
                'open'    => false,
                'nocheck' => false,
            ];
        }
        return $tree;
    }

    /** 同父下同名唯一（del_flag='0'；deptId 相同视为自身放行）→ bool 唯一 */
    public static function checkDeptNameUnique(string $deptName, int $parentId, int $deptId = 0): bool
    {
        $row = Db::table('sys_dept')
            ->where('dept_name', $deptName)
            ->where('parent_id', $parentId)
            ->where('del_flag', '0')
            ->find();
        if ($row === null) {
            return true;
        }
        return (int)$row['dept_id'] === $deptId;
    }

    /** 子部门计数（del_flag='0'） */
    public static function selectDeptCount(int $parentId): int
    {
        return Db::table('sys_dept')->where('parent_id', $parentId)->where('del_flag', '0')->count();
    }

    /** 部门下是否存在未删用户（对位 checkDeptExistUser） */
    public static function checkDeptExistUser(int $deptId): bool
    {
        return Db::table('sys_user')->where('dept_id', $deptId)->where('del_flag', '0')->count() > 0;
    }

    /** 是否存在未停用子部门（对位 selectNormalChildrenDeptById） */
    public static function selectNormalChildrenDeptById(int $deptId): int
    {
        return Db::table('sys_dept')
            ->where('status', '0')
            ->where('del_flag', '0')
            ->whereRaw("find_in_set(?, ancestors)", [$deptId])
            ->count();
    }

    /** 新增（对位 insertDept）：父状态校验 + ancestors 重算 */
    public static function insertDept(array $dept, string $loginName): int
    {
        $parent = Db::table('sys_dept')->where('dept_id', (int)$dept['parent_id'])->find();
        if ($parent === null) {
            throw new \BusinessException('上级部门不存在');
        }
        if ((string)$parent['status'] !== '0') {
            throw new \BusinessException('部门停用，不允许新增');
        }
        $ancestors = $parent['ancestors'] . ',' . $parent['dept_id'];
        $now = date('Y-m-d H:i:s');
        $row = [
            'parent_id'   => (int)$dept['parent_id'],
            'ancestors'   => $ancestors,
            'dept_name'   => $dept['dept_name'],
            'order_num'   => (int)($dept['order_num'] ?? 0),
            'leader'      => $dept['leader'] ?? '',
            'phone'       => $dept['phone'] ?? '',
            'email'       => $dept['email'] ?? '',
            'status'      => $dept['status'] ?? '0',
            'del_flag'    => '0',
            'create_by'   => $loginName,
            'create_time' => $now,
        ];
        Db::table('sys_dept')->insert($row);
        return (int)Db::table('sys_dept')->max('dept_id');
    }

    /** 修改（对位 updateDept，事务）：自身 + 子孙 ancestors 级联 + 启用上级链 */
    public static function updateDept(array $dept, string $loginName): void
    {
        $newParentId = (int)$dept['parent_id'];
        $newParent = Db::table('sys_dept')->where('dept_id', $newParentId)->find();
        $newAncestors = $newParent ? ($newParent['ancestors'] . ',' . $newParentId) : '0';

        Db::startTrans();
        try {
            $now = date('Y-m-d H:i:s');
            Db::table('sys_dept')->where('dept_id', (int)$dept['dept_id'])->update([
                'parent_id'   => $newParentId,
                'ancestors'   => $newAncestors,
                'dept_name'   => $dept['dept_name'],
                'order_num'   => (int)($dept['order_num'] ?? 0),
                'leader'      => $dept['leader'] ?? '',
                'phone'       => $dept['phone'] ?? '',
                'email'       => $dept['email'] ?? '',
                'status'      => $dept['status'] ?? '0',
                'update_by'   => $loginName,
                'update_time' => $now,
            ]);

            // 子孙 ancestors 级联：oldAncestors 前缀锚定替换（spec 加固点）
            $self = Db::table('sys_dept')->where('dept_id', (int)$dept['dept_id'])->field('ancestors')->find();
            $oldAncestors = (string)($dept['ancestors'] ?? '');
            if ($oldAncestors !== '') {
                Db::table('sys_dept')
                    ->whereRaw("find_in_set(?, ancestors)", [(int)$dept['dept_id']])
                    ->where('dept_id', '<>', (int)$dept['dept_id'])
                    ->exp('ancestors', "REPLACE(CONCAT(ancestors, ','), CONCAT('" . addcslashes($oldAncestors, "'\\") . ",', ','), '" . addcslashes($newAncestors, "'\\") . ",') - 1")
                    ->update(['update_time' => $now]);
            }

            // 启用级联：改启用且 ancestors 非空非 "0" → 链上全部置启用（对位经典版 status='0' 分支）
            if (($dept['status'] ?? '0') === '0') {
                $selfRow = Db::table('sys_dept')->where('dept_id', (int)$dept['dept_id'])->field('ancestors')->find();
                $ancestors = (string)($selfRow['ancestors'] ?? '');
                if ($ancestors !== '' && $ancestors !== '0') {
                    Db::table('sys_dept')->whereIn('dept_id', explode(',', $ancestors))->update(['status' => '0', 'update_time' => $now]);
                }
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 软删（对位 deleteDeptById） */
    public static function deleteDeptById(int $deptId): int
    {
        return Db::table('sys_dept')->where('dept_id', $deptId)->update(['del_flag' => '2']);
    }

    /** 批量排序（对位 updateDeptSort，事务；异常 → 「保存排序异常，请联系管理员」） */
    public static function updateDeptSort(array $deptIds, array $orderNums): void
    {
        Db::startTrans();
        try {
            foreach ($deptIds as $i => $id) {
                Db::table('sys_dept')->where('dept_id', (int)$id)->update(['order_num' => (int)$orderNums[$i]]);
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw new \BusinessException('保存排序异常，请联系管理员');
        }
    }

    /** 横向越权防护（对位 checkDeptDataScope）：非 admin 且数据范围外 → 异常 */
    public static function checkDeptDataScope(int $deptId, array $user): void
    {
        if (PermissionService::isAdmin($user)) {
            return;
        }
        $query = Db::table('sys_dept d')->where('d.dept_id', $deptId)->where('d.del_flag', '0');
        DataScope::apply($query, $user, 'd');
        if ($query->count() === 0) {
            throw new \BusinessException('没有权限访问部门数据！');
        }
    }
}
