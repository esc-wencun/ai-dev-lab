<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;
use app\service\DeptService;
use app\service\DictService;
use app\service\ExcelExportService;
use app\service\PostService;
use app\service\RoleService;
use app\service\SessionService;
use app\service\UserService;
use AjaxResult;
use PageQuery;
use PasswordService;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 用户管理控制器（对位经典版 SysUserController，21 方法）
 */
class UserController extends \app\BaseController
{
    /** Excel 列定义：导出 11 列（spec Excel 节） */
    private const EXPORT_COLUMNS = [
        ['name' => '用户序号', 'field' => 'userId', 'numeric' => true],
        ['name' => '登录名称', 'field' => 'loginName'],
        ['name' => '用户名称', 'field' => 'userName'],
        ['name' => '用户邮箱', 'field' => 'email'],
        ['name' => '手机号码', 'field' => 'phonenumber'],
        ['name' => '用户性别', 'field' => 'sex', 'convert' => '0=男,1=女,2=未知'],
        ['name' => '账号状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
        ['name' => '最后登录IP', 'field' => 'loginIp'],
        ['name' => '最后登录时间', 'field' => 'loginDate'],
        ['name' => '部门名称', 'field' => 'deptName'],
        ['name' => '部门负责人', 'field' => 'deptLeader'],
    ];

    /** 导入模板 7 列（IMPORT + ALL 型） */
    private const IMPORT_COLUMNS = [
        ['name' => '部门编号'],
        ['name' => '登录名称'],
        ['name' => '用户名称'],
        ['name' => '用户邮箱'],
        ['name' => '手机号码'],
        ['name' => '用户性别'],
        ['name' => '账号状态'],
    ];

    /** GET /system/user：列表页 */
    #[Perm('system:user:view')]
    public function index(Request $request): Response
    {
        return Response::create('user/index', 'view')->assign([
            'datas' => DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/user/list：TableDataInfo + 驼峰行 + dept 嵌套 + password/salt 剔除 */
    #[Perm('system:user:list')]
    public function list(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $pq = PageQuery::from($request->post(), ['login_name', 'create_time']);
        $query = UserService::selectUserList($this->listFilter($request), $user);
        if ($pq->orderBy !== null) {
            $query->order('u.' . $pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([UserService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/user/export：全量导出（同数据权限） */
    #[Perm('system:user:export')]
    #[Log('用户管理', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $pq = PageQuery::from($request->post(), ['login_name', 'create_time']);
        $query = UserService::selectUserList($this->listFilter($request), $user);
        if ($pq->orderBy !== null) {
            $query->order('u.' . $pq->orderBy, $pq->isAsc);
        }
        $rows = $query->select()->toArray();
        $data = array_map(function (array $r): array {
            $row = UserService::toResponseRow($r);
            $row['deptName'] = $row['dept']['dept_name'] ?? '';
            $row['deptLeader'] = $row['dept']['leader'] ?? '';
            return $row;
        }, $rows);
        $fileName = ExcelExportService::export($data, self::EXPORT_COLUMNS, '用户数据');
        return AjaxResult::success($fileName);
    }

    /** POST /system/user/importData：multipart 导入 */
    #[Perm('system:user:import')]
    #[Log('用户管理', Log::IMPORT)]
    public function importData(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $file = $request->file('file');
        if ($file === null || !$file->isValid()) {
            return AjaxResult::error('导入文件不存在');
        }
        $ext = strtolower($file->getOriginalExtension());
        if (!in_array($ext, ['xls', 'xlsx'], true)) {
            return AjaxResult::error('导入文件格式错误，请下载模板并使用 xls 或 xlsx 格式。');
        }
        $updateSupport = $request->post('updateSupport', '') === 'on' || $request->post('updateSupport', '') === 'true';
        $rows = ExcelExportService::parse($file->getRealPath(), [
            '用户性别' => '男=0,女=1,未知=2',
            '账号状态' => '正常=0,停用=1',
        ]);
        $msg = UserService::importUser($rows, $updateSupport, (string)($user['loginName'] ?? ''), $user);
        return AjaxResult::success($msg);
    }

    /** POST /system/user/importTemplate：空模板（权限 view 经典原样） */
    #[Perm('system:user:view')]
    public function importTemplate(Request $request): Response
    {
        $fileName = ExcelExportService::exportTemplate(self::IMPORT_COLUMNS, '用户数据');
        return AjaxResult::success($fileName);
    }

    /** GET /system/user/add：新增标签页 */
    #[Perm('system:user:add')]
    public function add(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $roles = array_values(array_filter(RoleService::selectRoleAll($user), fn($r) => $r['roleId'] !== 1));
        return Response::create('user/add', 'view')->assign([
            'roles'        => $roles,
            'posts'        => PostService::selectPostAll(),
            'initPassword' => \app\service\ConfigService::get('sys.user.initPassword', '123456'),
        ]);
    }

    /** POST /system/user/add */
    #[Perm('system:user:add')]
    #[Log('用户管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->userInput($request);
        $loginName = (string)$input['login_name'];

        // 校验顺序：部门数据权限 → 角色数据权限 → 三唯一
        if (!empty($input['dept_id'])) {
            DeptService::checkDeptDataScope((int)$input['dept_id'], $session);
        }
        RoleService::checkRoleDataScope($this->ids($request->post('roleIds', '')), $session);
        if (!UserService::checkLoginNameUnique($loginName)) {
            return AjaxResult::error("新增用户'{$loginName}'失败，登录账号已存在");
        }
        $phone = (string)($input['phonenumber'] ?? '');
        if ($phone !== '' && !UserService::checkPhoneUnique($phone)) {
            return AjaxResult::error("新增用户'{$loginName}'失败，手机号码已存在");
        }
        $email = (string)($input['email'] ?? '');
        if ($email !== '' && !UserService::checkEmailUnique($email)) {
            return AjaxResult::error("新增用户'{$loginName}'失败，邮箱账号已存在");
        }

        $salt = PasswordService::randomSalt();
        $row = array_merge($input, [
            'salt'            => $salt,
            'password'        => PasswordService::encrypt($loginName, (string)$request->post('password', ''), $salt),
            'user_type'       => '00',
            'create_by'       => (string)($session['loginName'] ?? ''),
            'create_time'     => date('Y-m-d H:i:s'),
            'pwd_update_date' => date('Y-m-d H:i:s'),
        ]);
        UserService::insertUser($row, $this->ids($request->post('roleIds', '')), $this->ids($request->post('postIds', '')));
        return AjaxResult::success();
    }

    /** GET /system/user/edit/{userId} */
    #[Perm('system:user:edit')]
    public function edit(Request $request, int $userId): Response
    {
        $session = $request->middleware('session') ?? [];
        UserService::checkUserDataScope($userId, $session);
        $user = UserService::selectUserById($userId);
        if ($user === null) {
            throw new \BusinessException('用户不存在');
        }
        // 目标用户非 admin 时过滤 admin 角色（经典实锤）
        $excludeAdmin = $userId !== 1;
        return Response::create('user/edit', 'view')->assign([
            'user'         => $user,
            'roles'        => RoleService::selectRolesByUserId($userId, $session, $excludeAdmin),
            'posts'        => PostService::selectPostsByUserId($userId),
            'initPassword' => '',
        ]);
    }

    /** GET /system/user/view/{userId}：右滑详情 */
    #[Perm('system:user:list')]
    public function view(Request $request, int $userId): Response
    {
        $session = $request->middleware('session') ?? [];
        UserService::checkUserDataScope($userId, $session);
        $user = UserService::selectUserById($userId);
        if ($user === null) {
            throw new \BusinessException('用户不存在');
        }
        return Response::create('user/view', 'view')->assign([
            'user'      => $user,
            'roleGroup' => UserService::selectUserRoleGroup($userId),
            'postGroup' => UserService::selectUserPostGroup($userId),
            'datas'     => DictService::listByType('sys_user_sex'),
        ]);
    }

    /** POST /system/user/edit */
    #[Perm('system:user:edit')]
    #[Log('用户管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)$request->post('userId', 0);
        UserService::checkUserAllowed($userId);
        UserService::checkUserDataScope($userId, $session);
        $input = $this->userInput($request);
        $input['user_id'] = $userId;
        $loginName = (string)UserService::selectUserById($userId)['login_name'];

        if (!empty($input['dept_id'])) {
            DeptService::checkDeptDataScope((int)$input['dept_id'], $session);
        }
        RoleService::checkRoleDataScope($this->ids($request->post('roleIds', '')), $session);
        $phone = (string)($input['phonenumber'] ?? '');
        if ($phone !== '' && !UserService::checkPhoneUnique($phone, $userId)) {
            return AjaxResult::error("修改用户'{$loginName}'失败，手机号码已存在");
        }
        $email = (string)($input['email'] ?? '');
        if ($email !== '' && !UserService::checkEmailUnique($email, $userId)) {
            return AjaxResult::error("修改用户'{$loginName}'失败，邮箱账号已存在");
        }

        $input['update_by'] = (string)($session['loginName'] ?? '');
        $input['update_time'] = date('Y-m-d H:i:s');
        UserService::updateUser($input, $this->ids($request->post('roleIds', '')), $this->ids($request->post('postIds', '')));
        return AjaxResult::success();
    }

    /** GET /system/user/resetPwd/{userId}：重置密码弹窗 */
    #[Perm('system:user:resetPwd')]
    public function resetPwd(Request $request, int $userId): Response
    {
        $session = $request->middleware('session') ?? [];
        UserService::checkUserDataScope($userId, $session);
        $user = UserService::selectUserById($userId);
        if ($user === null) {
            throw new \BusinessException('用户不存在');
        }
        return Response::create('user/resetPwd', 'view')->assign([
            'user'         => array_merge($user, ['loginName' => $user['login_name']]),
            'initPassword' => \app\service\ConfigService::get('sys.user.initPassword', '123456'),
        ]);
    }

    /** POST /system/user/resetPwd：重置对象=自己时刷新会话 */
    #[Perm('system:user:resetPwd')]
    #[Log('重置密码', Log::UPDATE)]
    public function resetPwdSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)$request->post('userId', 0);
        UserService::checkUserAllowed($userId);
        UserService::checkUserDataScope($userId, $session);

        $user = UserService::selectUserById($userId);
        $loginName = (string)($user['login_name'] ?? $request->post('loginName', ''));
        $salt = PasswordService::randomSalt();
        $rows = UserService::resetUserPwd($userId, PasswordService::encrypt($loginName, (string)$request->post('password', ''), $salt), $salt);
        if ($rows > 0) {
            // 重置对象=当前登录用户 → 会话写回（对位 setSysUser；本会话为 admin 场景仅 userId 相等时触发）
            if (($session['userId'] ?? 0) === $userId) {
                $fresh = UserService::selectUserById($userId);
                if ($fresh !== null) {
                    // roles 结构对位 LoginService.login：role_id/role_key/role_name/data_scope
                    $roles = \think\facade\Db::table('sys_role r')
                        ->join('sys_user_role ur', 'ur.role_id = r.role_id')
                        ->where('ur.user_id', $userId)
                        ->where('r.status', '0')
                        ->where('r.del_flag', '0')
                        ->field('r.role_id,r.role_key,r.role_name,r.data_scope')
                        ->select()->toArray();
                    SessionService::write((string)$request->middleware('session_uuid'), [
                        'userId'        => (int)$fresh['user_id'],
                        'loginName'     => $fresh['login_name'],
                        'userName'      => $fresh['user_name'],
                        'deptId'        => $fresh['dept_id'] !== null ? (int)$fresh['dept_id'] : null,
                        'avatar'        => (string)$fresh['avatar'],
                        'isAdmin'       => PermissionService::isAdmin(['user_id' => (int)$fresh['user_id']]),
                        'permissions'   => PermissionService::permissionsOf(['user_id' => (int)$fresh['user_id']]),
                        'roles'         => $roles,
                        'pwdUpdateDate' => $fresh['pwd_update_date'],
                        'rememberMe'    => false,
                    ]);
                }
            }
            return AjaxResult::success();
        }
        return AjaxResult::error();
    }

    /** GET /system/user/authRole/{userId}：分配角色标签页 */
    #[Perm('system:user:edit')]
    public function authRole(Request $request, int $userId): Response
    {
        $session = $request->middleware('session') ?? [];
        UserService::checkUserDataScope($userId, $session);
        $user = UserService::selectUserById($userId);
        if ($user === null) {
            throw new \BusinessException('用户不存在');
        }
        return Response::create('user/authRole', 'view')->assign([
            'user'  => $user,
            'roles' => RoleService::selectRolesByUserId($userId, $session, $userId !== 1),
        ]);
    }

    /** POST /system/user/authRole/insertAuthRole：授权（事务，恒 success） */
    #[Perm('system:user:edit')]
    #[Log('用户管理', Log::GRANT)]
    public function insertAuthRole(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)$request->post('userId', 0);
        UserService::checkUserDataScope($userId, $session);
        RoleService::checkRoleDataScope($this->ids($request->post('roleIds', '')), $session);
        UserService::insertUserAuth($userId, $this->ids($request->post('roleIds', '')));
        return AjaxResult::success();
    }

    /** POST /system/user/remove：软删（关联物理删；拦自己 + admin 逐个） */
    #[Perm('system:user:remove')]
    #[Log('用户管理', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $ids = $this->ids($request->post('ids', ''));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        if (in_array((int)($session['userId'] ?? 0), $ids, true)) {
            return AjaxResult::error('当前用户不能删除');
        }
        foreach ($ids as $id) {
            UserService::checkUserAllowed($id);
            UserService::checkUserDataScope($id, $session);
        }
        $rows = UserService::deleteUserByIds($ids);
        return $rows > 0 ? AjaxResult::success() : AjaxResult::error();
    }

    /** POST /system/user/checkLoginNameUnique：裸 boolean（无 #[Perm]） */
    public function checkLoginNameUnique(Request $request): Response
    {
        $unique = UserService::checkLoginNameUnique((string)$request->post('loginName', ''), (int)$request->post('userId', 0));
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/user/checkPhoneUnique */
    public function checkPhoneUnique(Request $request): Response
    {
        $unique = UserService::checkPhoneUnique((string)$request->post('phonenumber', ''), (int)$request->post('userId', 0));
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/user/checkEmailUnique */
    public function checkEmailUnique(Request $request): Response
    {
        $unique = UserService::checkEmailUnique((string)$request->post('email', ''), (int)$request->post('userId', 0));
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/user/changeStatus：admin 行后端拦截（经典 quirk） */
    #[Perm('system:user:edit')]
    #[Log('用户管理', Log::UPDATE)]
    public function changeStatus(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $userId = (int)$request->post('userId', 0);
        UserService::checkUserAllowed($userId);
        UserService::checkUserDataScope($userId, $session);
        UserService::changeStatus($userId, (string)$request->post('status', '0'));
        return AjaxResult::success();
    }

    /** GET /system/user/deptTreeData：Ztree 裸数组（复用 3.0.0） */
    #[Perm('system:user:list')]
    public function deptTreeData(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        return \think\Response::create(DeptService::selectDeptTreeData(0, $session), 'json');
    }

    /** GET /system/user/selectDeptTree/{deptId}：部门树选择弹窗（无 excludeId） */
    #[Perm('system:user:list')]
    public function selectDeptTree(Request $request, int $deptId): Response
    {
        $dept = DeptService::selectDeptById($deptId) ?: [];
        return Response::create('user/deptTree', 'view')->assign(['dept' => $dept]);
    }

    /** 用户表单输入收敛（文案对位经典版 @Validated） */
    private function userInput(Request $request): array
    {
        $loginName = trim((string)$request->post('loginName', ''));
        if ($loginName === '') {
            throw new \BusinessException('登录账号不能为空');
        }
        if (mb_strlen($loginName) > 30) {
            throw new \BusinessException('登录账号长度不能超过30个字符');
        }
        if (preg_match('/<(\S*?)[^>]*>.*?|<.*? \/>/', $loginName)) {
            throw new \BusinessException('登录账号不能包含脚本字符');
        }
        $userName = trim((string)$request->post('userName', ''));
        if (mb_strlen($userName) > 30) {
            throw new \BusinessException('用户昵称长度不能超过30个字符');
        }
        if ($userName !== '' && preg_match('/<(\S*?)[^>]*>.*?|<.*? \/>/', $userName)) {
            throw new \BusinessException('用户昵称不能包含脚本字符');
        }
        $email = (string)$request->post('email', '');
        if ($email !== '') {
            if (!filter_var($email, FILTER_VALIDATE_EMAIL)) {
                throw new \BusinessException('邮箱格式不正确');
            }
            if (mb_strlen($email) > 50) {
                throw new \BusinessException('邮箱长度不能超过50个字符');
            }
        }
        $phone = (string)$request->post('phonenumber', '');
        if (mb_strlen($phone) > 11) {
            throw new \BusinessException('手机号码长度不能超过11个字符');
        }
        return [
            'dept_id'     => (int)$request->post('deptId', 0) ?: null,
            'login_name'  => $loginName,
            'user_name'   => $userName,
            'email'       => $email,
            'phonenumber' => $phone,
            'sex'         => (string)$request->post('sex', '0'),
            'status'      => (string)$request->post('status', '0'),
            'remark'      => (string)$request->post('remark', ''),
        ];
    }

    private function listFilter(Request $request): array
    {
        $params = $request->post('params', []);
        return [
            'loginName'   => trim((string)$request->post('loginName', '')),
            'status'      => (string)$request->post('status', ''),
            'phonenumber' => trim((string)$request->post('phonenumber', '')),
            'beginTime'   => (string)($params['beginTime'] ?? ''),
            'endTime'     => (string)($params['endTime'] ?? ''),
            'deptId'      => $request->post('deptId'),
            'parentId'    => $request->post('parentId'),
        ];
    }

    /** 逗号串 → int 数组（空串过滤） */
    private function ids(string $comma): array
    {
        return array_values(array_filter(array_map('intval', explode(',', $comma)), fn($v) => $v > 0));
    }
}
