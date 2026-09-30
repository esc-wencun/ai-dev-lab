<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;
use app\service\ExcelExportService;
use app\service\RoleService;
use app\service\SessionService;
use AjaxResult;
use PageQuery;
use TableDataInfo;
use think\Request;
use think\Response;

/**
 * 角色管理控制器（对位经典版 SysRoleController 23 方法；selectMenuTree 死端点不注册）
 */
class RoleController extends \app\BaseController
{
    private const EXPORT_COLUMNS = [
        ['name' => '角色序号', 'field' => 'roleId', 'numeric' => true],
        ['name' => '角色名称', 'field' => 'roleName'],
        ['name' => '角色权限', 'field' => 'roleKey'],
        ['name' => '角色排序', 'field' => 'roleSort', 'numeric' => true],
        ['name' => '数据范围', 'field' => 'dataScope', 'convert' => '1=所有数据权限,2=自定义数据权限,3=本部门数据权限,4=本部门及以下数据权限,5=仅本人数据权限'],
        ['name' => '角色状态', 'field' => 'status', 'convert' => '0=正常,1=停用'],
    ];

    /** GET /system/role */
    #[Perm('system:role:view')]
    public function index(Request $request): Response
    {
        return Response::create('role/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/role/list */
    #[Perm('system:role:list')]
    public function list(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $pq = PageQuery::from($request->post(), ['role_name', 'role_key', 'role_sort', 'create_time']);
        $query = RoleService::selectRoleList($this->listFilter($request), $session);
        if ($pq->orderBy !== null) {
            $query->order('r.' . $pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([RoleService::class, 'toResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/role/export */
    #[Perm('system:role:export')]
    #[Log('角色管理', Log::EXPORT)]
    public function export(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $pq = PageQuery::from($request->post(), ['role_name', 'role_key', 'role_sort', 'create_time']);
        $query = RoleService::selectRoleList($this->listFilter($request), $session);
        if ($pq->orderBy !== null) {
            $query->order('r.' . $pq->orderBy, $pq->isAsc);
        }
        $rows = array_map([RoleService::class, 'toResponseRow'], $query->select()->toArray());
        $fileName = ExcelExportService::export($rows, self::EXPORT_COLUMNS, '角色数据');
        return AjaxResult::success($fileName);
    }

    /** GET /system/role/add */
    #[Perm('system:role:add')]
    public function add(Request $request): Response
    {
        return Response::create('role/add', 'view');
    }

    /** POST /system/role/add */
    #[Perm('system:role:add')]
    #[Log('角色管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->roleInput($request);
        if (!RoleService::checkRoleNameUnique($input['role_name'])) {
            return AjaxResult::error("新增角色'{$input['role_name']}'失败，角色名称已存在");
        }
        if (!RoleService::checkRoleKeyUnique($input['role_key'])) {
            return AjaxResult::error("新增角色'{$input['role_name']}'失败，角色权限已存在");
        }
        $input['create_by'] = (string)($session['loginName'] ?? '');
        $input['create_time'] = date('Y-m-d H:i:s');
        RoleService::insertRole($input, $this->ids($request->post('menuIds', '')), (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/role/edit/{roleId} */
    #[Perm('system:role:edit')]
    public function edit(Request $request, int $roleId): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([$roleId], $session);
        $role = RoleService::selectRoleById($roleId);
        if ($role === null) {
            throw new \BusinessException('角色不存在');
        }
        return Response::create('role/edit', 'view')->assign(['role' => $role]);
    }

    /** POST /system/role/edit */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        RoleService::checkRoleAllowed($roleId);
        RoleService::checkRoleDataScope([$roleId], $session);
        $input = $this->roleInput($request);
        $input['role_id'] = $roleId;
        if (!RoleService::checkRoleNameUnique($input['role_name'], $roleId)) {
            return AjaxResult::error("修改角色'{$input['role_name']}'失败，角色名称已存在");
        }
        if (!RoleService::checkRoleKeyUnique($input['role_key'], $roleId)) {
            return AjaxResult::error("修改角色'{$input['role_name']}'失败，角色权限已存在");
        }
        $input['update_by'] = (string)($session['loginName'] ?? '');
        $input['update_time'] = date('Y-m-d H:i:s');
        RoleService::updateRole($input, $this->ids($request->post('menuIds', '')), (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/role/authDataScope/{roleId}（无 #[Perm]，经典原样） */
    public function authDataScope(Request $request, int $roleId): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([$roleId], $session);
        $role = RoleService::selectRoleById($roleId);
        if ($role === null) {
            throw new \BusinessException('角色不存在');
        }
        $deptTree = RoleService::deptTreeData($roleId, $session);
        return Response::create('role/dataScope', 'view')->assign([
            'role'     => $role,
            'deptTree' => json_encode($deptTree, JSON_UNESCAPED_UNICODE),
        ]);
    }

    /** POST /system/role/authDataScope：成功后刷本人会话 */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::UPDATE)]
    public function authDataScopeSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        RoleService::checkRoleAllowed($roleId);
        RoleService::checkRoleDataScope([$roleId], $session);
        RoleService::authDataScope(
            $roleId,
            (string)$request->post('dataScope', '1'),
            $this->ids($request->post('deptIds', '')),
            (string)($session['loginName'] ?? '')
        );
        // 刷新当前登录会话的用户数据（对位 setSysUser：本人数据权限立即生效）
        $this->refreshOwnSession($request, $session);
        return AjaxResult::success();
    }

    /** POST /system/role/remove */
    #[Perm('system:role:remove')]
    #[Log('角色管理', Log::DELETE)]
    public function remove(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $ids = $this->ids($request->post('ids', ''));
        if (!$ids) {
            return AjaxResult::error('参数错误');
        }
        foreach ($ids as $id) {
            RoleService::checkRoleAllowed($id);
            RoleService::checkRoleDataScope([$id], $session);
            if (RoleService::countUserRoleByRoleId($id) > 0) {
                $role = RoleService::selectRoleById($id);
                return AjaxResult::error(($role['role_name'] ?? $id) . '已分配,不能删除');
            }
        }
        $rows = RoleService::deleteRoleByIds($ids);
        return $rows > 0 ? AjaxResult::success() : AjaxResult::error();
    }

    /** POST /system/role/checkRoleNameUnique：裸 boolean */
    public function checkRoleNameUnique(Request $request): Response
    {
        $unique = RoleService::checkRoleNameUnique((string)$request->post('roleName', ''), (int)$request->post('roleId', 0));
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/role/checkRoleKeyUnique：裸 boolean */
    public function checkRoleKeyUnique(Request $request): Response
    {
        $unique = RoleService::checkRoleKeyUnique((string)$request->post('roleKey', ''), (int)$request->post('roleId', 0));
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** POST /system/role/changeStatus */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::UPDATE)]
    public function changeStatus(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        RoleService::checkRoleAllowed($roleId);
        RoleService::checkRoleDataScope([$roleId], $session);
        RoleService::changeStatus($roleId, (string)$request->post('status', '0'), (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/role/authUser/{roleId}：分配用户页签 */
    #[Perm('system:role:edit')]
    public function authUser(Request $request, int $roleId): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([$roleId], $session);
        $role = RoleService::selectRoleById($roleId);
        if ($role === null) {
            throw new \BusinessException('角色不存在');
        }
        return Response::create('role/authUser', 'view')->assign([
            'role' => $role,
            'datas' => \app\service\DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/role/authUser/allocatedList */
    #[Perm('system:role:list')]
    public function allocatedList(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        $pq = PageQuery::from($request->post(), ['login_name', 'create_time']);
        $query = RoleService::selectAllocatedList([
            'loginName'   => trim((string)$request->post('loginName', '')),
            'phonenumber' => trim((string)$request->post('phonenumber', '')),
        ], $roleId, $session);
        if ($pq->orderBy !== null) {
            $query->order('u.' . $pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([RoleService::class, 'toUserResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/role/authUser/cancel：物理删单行 */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::GRANT)]
    public function cancel(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([(int)$request->post('roleId', 0)], $session);
        \think\facade\Db::table('sys_user_role')
            ->where('user_id', (int)$request->post('userId', 0))
            ->where('role_id', (int)$request->post('roleId', 0))
            ->delete();
        return AjaxResult::success();
    }

    /** POST /system/role/authUser/cancelAll */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::GRANT)]
    public function cancelAll(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([(int)$request->post('roleId', 0)], $session);
        $userIds = $this->ids($request->post('userIds', ''));
        if ($userIds) {
            \think\facade\Db::table('sys_user_role')
                ->where('role_id', (int)$request->post('roleId', 0))
                ->whereIn('user_id', $userIds)
                ->delete();
        }
        return AjaxResult::success();
    }

    /** GET /system/role/authUser/selectUser/{roleId}：**必须先于 authUser/:roleId 注册** */
    #[Perm('system:role:list')]
    public function selectUser(Request $request, int $roleId): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([$roleId], $session);
        $role = RoleService::selectRoleById($roleId);
        if ($role === null) {
            throw new \BusinessException('角色不存在');
        }
        return Response::create('role/selectUser', 'view')->assign([
            'role' => $role,
            'datas' => \app\service\DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/role/authUser/unallocatedList */
    #[Perm('system:role:list')]
    public function unallocatedList(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        $pq = PageQuery::from($request->post(), ['login_name', 'create_time']);
        $query = RoleService::selectUnallocatedList([
            'loginName'   => trim((string)$request->post('loginName', '')),
            'phonenumber' => trim((string)$request->post('phonenumber', '')),
        ], $roleId, $session);
        if ($pq->orderBy !== null) {
            $query->order('u.' . $pq->orderBy, $pq->isAsc);
        }
        $page = $query->paginate(['list_rows' => $pq->pageSize, 'page' => $pq->pageNum]);
        $rows = array_map([RoleService::class, 'toUserResponseRow'], array_values((array)$page->items()));
        return TableDataInfo::of($rows, (int)$page->total());
    }

    /** POST /system/role/authUser/selectAll：批量授权（无去重，经典原样） */
    #[Perm('system:role:edit')]
    #[Log('角色管理', Log::GRANT)]
    public function selectAll(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = (int)$request->post('roleId', 0);
        RoleService::checkRoleDataScope([$roleId], $session);
        $userIds = $this->ids($request->post('userIds', ''));
        if ($userIds) {
            \think\facade\Db::table('sys_user_role')->insertAll(
                array_map(fn($uid) => ['user_id' => $uid, 'role_id' => $roleId], $userIds)
            );
        }
        return AjaxResult::success();
    }

    /** GET /system/role/deptTreeData：Ztree 裸数组（checked=整串 equals） */
    #[Perm('system:role:edit')]
    public function deptTreeData(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = $request->get('roleId');
        return \think\Response::create(
            RoleService::deptTreeData($roleId !== null ? (int)$roleId : null, $session),
            'json'
        );
    }

    /** GET /system/role/view/{roleId}：详情页 */
    #[Perm('system:role:list')]
    public function view(Request $request, int $roleId): Response
    {
        $session = $request->middleware('session') ?? [];
        RoleService::checkRoleDataScope([$roleId], $session);
        $role = RoleService::selectRoleById($roleId);
        if ($role === null) {
            throw new \BusinessException('角色不存在');
        }
        return Response::create('role/view', 'view')->assign([
            'role'     => $role,
            'menuTree' => json_encode(\app\service\MenuService::roleMenuTreeData($roleId, $session), JSON_UNESCAPED_UNICODE),
            'deptTree' => (string)$role['data_scope'] === '2' ? json_encode(RoleService::deptTreeData($roleId, $session), JSON_UNESCAPED_UNICODE) : '',
            'userCount' => RoleService::countUserRoleByRoleId($roleId),
        ]);
    }

    /** 角色表单输入收敛（文案对位 @Validated） */
    private function roleInput(Request $request): array
    {
        $roleName = trim((string)$request->post('roleName', ''));
        if ($roleName === '') {
            throw new \BusinessException('角色名称不能为空');
        }
        if (mb_strlen($roleName) > 30) {
            throw new \BusinessException('角色名称长度不能超过30个字符');
        }
        $roleKey = trim((string)$request->post('roleKey', ''));
        if ($roleKey === '') {
            throw new \BusinessException('权限字符不能为空');
        }
        if (mb_strlen($roleKey) > 100) {
            throw new \BusinessException('权限字符长度不能超过100个字符');
        }
        if ($request->post('roleSort') === null || $request->post('roleSort') === '') {
            throw new \BusinessException('显示顺序不能为空');
        }
        return [
            'role_name'  => $roleName,
            'role_key'   => $roleKey,
            'role_sort'  => (int)$request->post('roleSort', 0),
            'status'     => (string)$request->post('status', '0'),
            'remark'     => (string)$request->post('remark', ''),
        ];
    }

    private function listFilter(Request $request): array
    {
        $params = $request->post('params', []);
        return [
            'roleName'  => trim((string)$request->post('roleName', '')),
            'roleKey'   => trim((string)$request->post('roleKey', '')),
            'status'    => (string)$request->post('status', ''),
            'beginTime' => (string)($params['beginTime'] ?? ''),
            'endTime'   => (string)($params['endTime'] ?? ''),
        ];
    }

    private function ids(string $comma): array
    {
        return array_values(array_filter(array_map('intval', explode(',', $comma)), fn($v) => $v > 0));
    }

    /** 刷新当前登录会话的用户数据（对位 setSysUser；authDataScopeSave 用） */
    private function refreshOwnSession(Request $request, array $session): void
    {
        $userId = (int)($session['userId'] ?? 0);
        if ($userId <= 0) {
            return;
        }
        $uuid = (string)$request->middleware('session_uuid');
        if ($uuid === '') {
            $uuid = (string)$request->cookie(SessionService::COOKIE_NAME);
        }
        $fresh = \app\service\UserService::selectUserById($userId);
        if ($fresh === null || $uuid === '') {
            return;
        }
        $roles = \think\facade\Db::table('sys_role r')
            ->join('sys_user_role ur', 'ur.role_id = r.role_id')
            ->where('ur.user_id', $userId)
            ->where('r.status', '0')
            ->where('r.del_flag', '0')
            ->field('r.role_id,r.role_key,r.role_name,r.data_scope')
            ->select()->toArray();
        SessionService::write($uuid, array_merge($session, [
            'roles'         => $roles,
            'pwdUpdateDate' => $fresh['pwd_update_date'],
        ]));
    }
}
