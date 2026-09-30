<?php
declare(strict_types=1);

namespace app\service;

use DataScope;
use PasswordService;
use think\facade\Db;

/**
 * 用户服务（对位经典版 SysUserServiceImpl）
 *
 * 数据权限：selectUserList 双别名 DataScope('d','u')（工作区首例，SELF 作用域靠 u.user_id）；
 * checkUserDataScope 八处入口；四处事务（insertUser/updateUser/insertUserAuth/deleteUserByIds）。
 */
final class UserService
{
    /** 列表查询构造器（对位 selectUserList；返回 Query，分页/全量由调用方决定） */
    public static function selectUserList(array $filter, array $user): \think\db\Query
    {
        $query = Db::table('sys_user u')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->where('u.del_flag', '0')
            ->field('u.user_id,u.dept_id,u.login_name,u.user_name,u.user_type,u.email,u.phonenumber,u.sex,u.avatar,u.status,u.del_flag,u.login_ip,u.login_date,u.pwd_update_date,u.create_by,u.create_time,u.remark,d.dept_id as dept__dept_id,d.parent_id as dept__parent_id,d.dept_name as dept__dept_name,d.leader as dept__leader,d.order_num as dept__order_num,d.status as dept__status');
        if (($filter['loginName'] ?? '') !== '') {
            $query->whereLike('u.login_name', '%' . $filter['loginName'] . '%');
        }
        if (($filter['status'] ?? '') !== '') {
            $query->where('u.status', $filter['status']);
        }
        if (($filter['phonenumber'] ?? '') !== '') {
            $query->whereLike('u.phonenumber', '%' . $filter['phonenumber'] . '%');
        }
        // create_time 日粒度区间（对位经典 params[beginTime]/params[endTime] %Y%m%d 比较）
        $begin = $filter['beginTime'] ?? '';
        $end = $filter['endTime'] ?? '';
        if ($begin !== '') {
            $query->whereTime('u.create_time', '>=', date('Y-m-d 00:00:00', strtotime($begin)));
        }
        if ($end !== '') {
            $query->whereTime('u.create_time', '<=', date('Y-m-d 23:59:59', strtotime($end)));
        }
        // 部门树点击：含子部门（对位 mapper 的 OR FIND_IN_SET 子查询）
        if (!empty($filter['deptId'])) {
            $deptId = (int)$filter['deptId'];
            $query->where(function ($q) use ($deptId) {
                $q->whereOr('u.dept_id', $deptId)
                  ->whereOr('u.dept_id', 'in', Db::table('sys_dept')->whereRaw("FIND_IN_SET({$deptId}, ancestors)")->column('dept_id'));
            });
        }
        // 数据权限双别名（工作区首例；SELF 作用域产出 u.user_id 条件）
        if (!PermissionService::isAdmin($user)) {
            DataScope::apply($query, $user, 'd', 'u');
        }
        return $query;
    }

    /** 查询结果行 → 响应行（驼峰 + dept 嵌套对象 + password/salt 剔除，@JsonIgnore 对位） */
    public static function toResponseRow(array $r): array
    {
        $row = [
            'userId'        => (int)$r['user_id'],
            'deptId'        => $r['dept_id'] !== null ? (int)$r['dept_id'] : null,
            'loginName'     => $r['login_name'],
            'userName'      => $r['user_name'],
            'userType'      => $r['user_type'],
            'email'         => $r['email'],
            'phonenumber'   => $r['phonenumber'],
            'sex'           => $r['sex'],
            'avatar'        => $r['avatar'],
            'status'        => $r['status'],
            'loginIp'       => $r['login_ip'],
            'loginDate'     => $r['login_date'],
            'pwdUpdateDate' => $r['pwd_update_date'],
            'createBy'      => $r['create_by'],
            'createTime'    => $r['create_time'],
            'remark'        => $r['remark'],
            'dept'          => [],
            'roles'         => [],
            'roleIds'       => [],
            'postIds'       => [],
            'roleId'        => null,
            'roleKey'       => '',
        ];
        // dept 嵌套（dept__ 前缀列来自 join 别名）
        foreach ($r as $k => $v) {
            if (str_starts_with((string)$k, 'dept__') && $v !== null) {
                $row['dept'][substr((string)$k, 6)] = $v;
            }
        }
        return $row;
    }

    /** 单条（selectUserVo 全列 + dept 嵌套；供编辑页/会话刷新） */
    public static function selectUserById(int $userId): ?array
    {
        $row = Db::table('sys_user u')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->where('u.user_id', $userId)
            ->field('u.*,d.dept_name,d.leader')
            ->find();
        if ($row === null) {
            return null;
        }
        $row['dept'] = $row['dept_name'] !== null ? ['dept_name' => $row['dept_name'], 'leader' => $row['leader']] : [];
        unset($row['dept_name'], $row['leader']);
        return $row;
    }

    public static function selectUserByLoginName(string $loginName): ?array
    {
        return Db::table('sys_user')->where('login_name', $loginName)->where('del_flag', '0')->find();
    }

    /** 用户已有角色 id 集 */
    public static function selectUserRoleIds(int $userId): array
    {
        return array_map('intval', Db::table('sys_user_role')->where('user_id', $userId)->column('role_id'));
    }

    /** 三唯一（全局 del_flag='0'，userId 自身放行） */
    public static function checkLoginNameUnique(string $loginName, int $userId = 0): bool
    {
        $row = Db::table('sys_user')->where('login_name', $loginName)->where('del_flag', '0')->find();
        return $row === null || (int)$row['user_id'] === $userId;
    }

    public static function checkPhoneUnique(string $phone, int $userId = 0): bool
    {
        $row = Db::table('sys_user')->where('phonenumber', $phone)->where('del_flag', '0')->find();
        return $row === null || (int)$row['user_id'] === $userId;
    }

    public static function checkEmailUnique(string $email, int $userId = 0): bool
    {
        $row = Db::table('sys_user')->where('email', $email)->where('del_flag', '0')->find();
        return $row === null || (int)$row['user_id'] === $userId;
    }

    /** admin 保护（对位 checkUserAllowed） */
    public static function checkUserAllowed(int $userId): void
    {
        if ($userId === 1) {
            throw new \BusinessException('不允许操作超级管理员用户');
        }
    }

    /** 横向越权防护（对位 checkUserDataScope）：非 admin 用带权限 selectUserList 判空 */
    public static function checkUserDataScope(int $userId, array $user): void
    {
        if (PermissionService::isAdmin($user)) {
            return;
        }
        $rows = self::selectUserList(['userId' => $userId], $user)->select()->toArray();
        // selectUserList 无 userId 条件分支——直接补一条精确查询（避免 filter 数组污染）
        $hit = Db::table('sys_user u')
            ->leftJoin('sys_dept d', 'u.dept_id = d.dept_id')
            ->where('u.user_id', $userId)
            ->where('u.del_flag', '0');
        DataScope::apply($hit, $user, 'd', 'u');
        if ($hit->count() === 0) {
            throw new \BusinessException('没有权限访问用户数据！');
        }
    }

    /** 角色名逗号拼接（对位 selectUserRoleGroup；session roles 优先，库查兜底） */
    public static function selectUserRoleGroup(int $userId): string
    {
        $names = Db::table('sys_role r')
            ->join('sys_user_role ur', 'ur.role_id = r.role_id')
            ->where('ur.user_id', $userId)
            ->where('r.del_flag', '0')
            ->column('r.role_name');
        return implode(',', $names);
    }

    /** 岗位名逗号拼接 */
    public static function selectUserPostGroup(int $userId): string
    {
        $names = Db::table('sys_post p')
            ->join('sys_user_post up', 'up.post_id = p.post_id')
            ->where('up.user_id', $userId)
            ->column('p.post_name');
        return implode(',', $names);
    }

    /** 新增（事务）：user + user_post + user_role（空数组跳过关联） */
    public static function insertUser(array $user, array $roleIds, array $postIds): void
    {
        Db::startTrans();
        try {
            Db::table('sys_user')->insert($user);
            $userId = (int)Db::table('sys_user')->max('user_id');
            if ($postIds) {
                $rows = array_map(fn($pid) => ['user_id' => $userId, 'post_id' => (int)$pid], $postIds);
                Db::table('sys_user_post')->insertAll($rows);
            }
            if ($roleIds) {
                $rows = array_map(fn($rid) => ['user_id' => $userId, 'role_id' => (int)$rid], $roleIds);
                Db::table('sys_user_role')->insertAll($rows);
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 修改（事务）：user_role 删→插 → user_post 删→插 → update（非空列语义，login_name 不在 set） */
    public static function updateUser(array $user, array $roleIds, array $postIds): void
    {
        Db::startTrans();
        try {
            $userId = (int)$user['user_id'];
            if ($roleIds) {
                Db::table('sys_user_role')->where('user_id', $userId)->delete();
                Db::table('sys_user_role')->insertAll(array_map(fn($rid) => ['user_id' => $userId, 'role_id' => (int)$rid], $roleIds));
            }
            if ($postIds) {
                Db::table('sys_user_post')->where('user_id', $userId)->delete();
                Db::table('sys_user_post')->insertAll(array_map(fn($pid) => ['user_id' => $userId, 'post_id' => (int)$pid], $postIds));
            }
            $set = array_filter([
                'dept_id'     => $user['dept_id'] ?? null,
                'user_name'   => $user['user_name'] ?? null,
                'email'       => $user['email'] ?? null,
                'phonenumber' => $user['phonenumber'] ?? null,
                'sex'         => $user['sex'] ?? null,
                'status'      => $user['status'] ?? null,
                'remark'      => $user['remark'] ?? null,
                'update_by'   => $user['update_by'] ?? null,
                'update_time' => $user['update_time'] ?? null,
            ], fn($v) => $v !== null && $v !== '');
            if ($set) {
                Db::table('sys_user')->where('user_id', $userId)->update($set);
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 授权（事务）：user_role 删→插（对位 insertUserAuth） */
    public static function insertUserAuth(int $userId, array $roleIds): void
    {
        Db::startTrans();
        try {
            Db::table('sys_user_role')->where('user_id', $userId)->delete();
            if ($roleIds) {
                Db::table('sys_user_role')->insertAll(array_map(fn($rid) => ['user_id' => $userId, 'role_id' => (int)$rid], $roleIds));
            }
            Db::commit();
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 批量删除（事务）：关联物理删 + 软删（调用方已逐个做 admin/越权/自删防护） */
    public static function deleteUserByIds(array $ids): int
    {
        Db::startTrans();
        try {
            Db::table('sys_user_role')->whereIn('user_id', $ids)->delete();
            Db::table('sys_user_post')->whereIn('user_id', $ids)->delete();
            $rows = Db::table('sys_user')->whereIn('user_id', $ids)->update(['del_flag' => '2']);
            Db::commit();
            return (int)$rows;
        } catch (\Throwable $e) {
            Db::rollback();
            throw $e;
        }
    }

    /** 状态切换 */
    public static function changeStatus(int $userId, string $status): int
    {
        return Db::table('sys_user')->where('user_id', $userId)->update(['status' => $status, 'update_time' => date('Y-m-d H:i:s')]);
    }

    /** 重置密码（password+salt+pwd_update_date+update_time） */
    public static function resetUserPwd(int $userId, string $password, string $salt): int
    {
        return Db::table('sys_user')->where('user_id', $userId)->update([
            'password'        => $password,
            'salt'            => $salt,
            'pwd_update_date' => date('Y-m-d H:i:s'),
            'update_time'     => date('Y-m-d H:i:s'),
        ]);
    }

    /** 修改个人资料（userName/email/phonenumber/sex 四字段 + update_time；update_by 不写——mapper 实锤） */
    public static function updateUserInfo(array $data, int $userId): int
    {
        return Db::table('sys_user')->where('user_id', $userId)->update(array_merge([
            'user_name'   => $data['user_name'],
            'email'       => $data['email'],
            'phonenumber' => $data['phonenumber'],
            'sex'         => $data['sex'],
        ], ['update_time' => date('Y-m-d H:i:s')]));
    }

    /** 更新头像 URL（avatar + update_time） */
    public static function updateUserAvatar(string $avatar, int $userId): int
    {
        return Db::table('sys_user')->where('user_id', $userId)->update([
            'avatar'      => $avatar,
            'update_time' => date('Y-m-d H:i:s'),
        ]);
    }

    /**
     * 导入用户（对位 importUser 校验链）。
     * 行键 = 表头名（部门编号/登录名称/用户名称/用户邮箱/手机号码/用户性别/账号状态）。
     * 成功返回成功消息；任一失败整批抛 BusinessException（<br/> 消息拼装）。
     */
    public static function importUser(array $rows, bool $updateSupport, string $operName, array $sessionUser): string
    {
        if (!$rows) {
            throw new \BusinessException('导入用户数据不能为空！');
        }
        $initPassword = ConfigService::get('sys.user.initPassword', '123456');
        $successNum = 0;
        $failureNum = 0;
        $successMsg = [];
        $message = [];

        foreach ($rows as $row) {
            $loginName = trim((string)($row['登录名称'] ?? ''));
            $userName = trim((string)($row['用户名称'] ?? ''));
            try {
                if ($loginName === '') {
                    throw new \BusinessException('登录账号不能为空');
                }
                // 防脚本字符（对位经典 Xss 正则判 HTML）
                foreach (['登录名称' => $loginName, '用户名称' => $userName] as $label => $val) {
                    if ($val !== '' && preg_match('/<(\S*?)[^>]*>.*?|<.*? \/>/', $val)) {
                        throw new \BusinessException($label . '不能包含脚本字符');
                    }
                }
                $email = trim((string)($row['用户邮箱'] ?? ''));
                if ($email !== '') {
                    if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
                        throw new \BusinessException('邮箱格式不正确');
                    }
                    if (mb_strlen($email) > 50) {
                        throw new \BusinessException('邮箱长度不能超过50个字符');
                    }
                }
                $phone = trim((string)($row['手机号码'] ?? ''));
                if (mb_strlen($phone) > 11) {
                    throw new \BusinessException('手机号码长度不能超过11个字符');
                }
                // 性别/状态反向转换（男→0）
                $sex = ExcelExportService::convertByExp((string)($row['用户性别'] ?? ''), '男=0,女=1,未知=2');
                $status = ExcelExportService::convertByExp((string)($row['账号状态'] ?? ''), '正常=0,停用=1');
                $deptId = (int)($row['部门编号'] ?? 0);

                $existing = self::selectUserByLoginName($loginName);
                if ($existing === null) {
                    // 新增：md5(loginName+初始密码) 不写 salt（经典原样 quirk，deviations 候选 1）
                    $password = PasswordService::encrypt($loginName, $initPassword, '');
                    Db::table('sys_user')->insert([
                        'dept_id'     => $deptId ?: null,
                        'login_name'  => $loginName,
                        'user_name'   => $userName,
                        'user_type'   => '00',
                        'email'       => $email,
                        'phonenumber' => $phone,
                        'sex'         => $sex,
                        'password'    => $password,
                        'status'      => $status,
                        'create_by'   => $operName,
                        'create_time' => date('Y-m-d H:i:s'),
                        'remark'      => '',
                    ]);
                    $successNum++;
                    $successMsg[] = "{$successNum}、账号 {$loginName} 导入成功";
                } elseif ($updateSupport) {
                    self::checkUserAllowed((int)$existing['user_id']);
                    self::checkUserDataScope((int)$existing['user_id'], $sessionUser);
                    if ($deptId > 0) {
                        DeptService::checkDeptDataScope($deptId, $sessionUser);
                    }
                    // 更新分支 deptId 用库中现值覆盖（经典原样 quirk：导入的部门编号列不生效）
                    Db::table('sys_user')->where('user_id', (int)$existing['user_id'])->update([
                        'user_name'   => $userName,
                        'email'       => $email,
                        'phonenumber' => $phone,
                        'sex'         => $sex,
                        'status'      => $status,
                        'update_by'   => $operName,
                        'update_time' => date('Y-m-d H:i:s'),
                    ]);
                    $successNum++;
                    $successMsg[] = "{$successNum}、账号 {$loginName} 更新成功";
                } else {
                    $failureNum++;
                    $message[] = "{$failureNum}、账号 {$loginName} 已存在";
                }
            } catch (\BusinessException $e) {
                $failureNum++;
                $msg = htmlspecialchars($loginName, ENT_QUOTES);
                $message[] = "{$failureNum}、账号 {$msg} 导入失败：{$e->getMessage()}";
            } catch (\Throwable $e) {
                $failureNum++;
                $msg = htmlspecialchars($loginName, ENT_QUOTES);
                $message[] = "{$failureNum}、账号 {$msg} 导入失败：{$e->getMessage()}";
            }
        }

        if ($failureNum > 0) {
            throw new \BusinessException("很抱歉，导入失败！共 {$failureNum} 条数据格式不正确，错误如下：<br/>" . implode('<br/>', $message));
        }
        return "恭喜您，数据已全部导入成功！共 {$successNum} 条，数据如下：<br/>" . implode('<br/>', $successMsg);
    }
}
