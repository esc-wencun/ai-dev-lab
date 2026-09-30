<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;
use app\service\DeptService;
use app\service\PermissionService;
use AjaxResult;
use TableDataInfo;
use TpConstant;
use think\Request;
use think\Response;

/**
 * 部门管理控制器（对位经典版 SysDeptController，11 方法）
 */
class DeptController extends \app\BaseController
{
    /** GET /system/dept：列表页 */
    #[Perm('system:dept:view')]
    public function index(Request $request): Response
    {
        return Response::create('dept/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_normal_disable'),
        ]);
    }

    /** POST /system/dept/list：**裸数组**（bootstrap-tree-table 特例，禁套信封） */
    #[Perm('system:dept:list')]
    public function list(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $rows = DeptService::selectDeptList([
            'deptName' => trim((string)$request->post('deptName', '')),
            'status'   => (string)$request->post('status', ''),
            'deptId'   => $request->post('deptId'),
            'parentId' => $request->post('parentId'),
        ], $user);
        // 输出驼峰键（对位经典版 SysDept 实体 Jackson 序列化；bootstrap-tree-table columns field: deptName 等按驼峰取值）
        $rows = array_map(static fn(array $r): array => [
            'deptId'      => (int)$r['dept_id'],
            'parentId'    => (int)$r['parent_id'],
            'ancestors'   => $r['ancestors'],
            'deptName'    => $r['dept_name'],
            'orderNum'    => (int)$r['order_num'],
            'leader'      => $r['leader'],
            'phone'       => $r['phone'],
            'email'       => $r['email'],
            'status'      => $r['status'],
            'createTime'  => $r['create_time'],
        ], $rows);
        return \think\Response::create($rows, 'json');
    }

    /** GET /system/dept/add/{parentId}：新增弹窗（非 admin 强制 parentId=本人部门） */
    #[Perm('system:dept:add')]
    public function add(Request $request, int $parentId): Response
    {
        $user = $request->middleware('session') ?? [];
        if (!PermissionService::isAdmin($user)) {
            $parentId = (int)($user['deptId'] ?? $parentId);
        }
        $dept = DeptService::selectDeptById($parentId) ?: [];
        return Response::create('dept/add', 'view')->assign(['dept' => $dept]);
    }

    /** POST /system/dept/add */
    #[Perm('system:dept:add')]
    #[Log('部门管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $input = $this->deptInput($request);
        if (!DeptService::checkDeptNameUnique($input['dept_name'], (int)$input['parent_id'])) {
            return AjaxResult::error("新增部门'{$input['dept_name']}'失败，部门名称已存在");
        }
        $input['ancestors'] = '';
        DeptService::insertDept($input, (string)($user['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/dept/edit/{deptId}：修改弹窗（先数据权限校验；deptId=100 parentName=「无」） */
    #[Perm('system:dept:edit')]
    public function edit(Request $request, int $deptId): Response
    {
        $user = $request->middleware('session') ?? [];
        DeptService::checkDeptDataScope($deptId, $user);
        $dept = DeptService::selectDeptById($deptId);
        if ($dept === null) {
            throw new \BusinessException('部门不存在');
        }
        return Response::create('dept/edit', 'view')->assign(['dept' => $dept]);
    }

    /** POST /system/dept/edit（校验顺序：数据权限由 GET 已查；同父同名 → 上级是自己 → 停用含未停用子部门） */
    #[Perm('system:dept:edit')]
    #[Log('部门管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $user = $request->middleware('session') ?? [];
        $input = $this->deptInput($request);
        $deptId = (int)$input['dept_id'];
        DeptService::checkDeptDataScope($deptId, $user);

        if (!DeptService::checkDeptNameUnique($input['dept_name'], (int)$input['parent_id'], $deptId)) {
            return AjaxResult::error("修改部门'{$input['dept_name']}'失败，部门名称已存在");
        }
        if ((int)$input['parent_id'] === $deptId) {
            return AjaxResult::error("修改部门'{$input['dept_name']}'失败，上级部门不能是自己");
        }
        if (($input['status'] ?? '0') === '1' && DeptService::selectNormalChildrenDeptById($deptId) > 0) {
            return AjaxResult::error('该部门包含未停用的子部门！');
        }
        DeptService::updateDept($input, (string)($user['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /system/dept/updateSort：批量排序（事务） */
    #[Perm('system:dept:edit')]
    #[Log('保存部门排序', Log::UPDATE)]
    public function updateSort(Request $request): Response
    {
        $deptIds = explode(',', (string)$request->post('deptIds', ''));
        $orderNums = explode(',', (string)$request->post('orderNums', ''));
        if (count($deptIds) !== count($orderNums) || $deptIds === ['']) {
            return AjaxResult::error('参数错误');
        }
        DeptService::updateDeptSort($deptIds, $orderNums);
        return AjaxResult::success();
    }

    /** POST /system/dept/remove/{deptId}：软删（两连 warn 301） */
    #[Perm('system:dept:remove')]
    #[Log('部门管理', Log::DELETE)]
    public function remove(Request $request, int $deptId): Response
    {
        $user = $request->middleware('session') ?? [];
        if (DeptService::selectDeptCount($deptId) > 0) {
            return AjaxResult::warn('存在下级部门,不允许删除');
        }
        if (DeptService::checkDeptExistUser($deptId)) {
            return AjaxResult::warn('部门存在用户,不允许删除');
        }
        DeptService::checkDeptDataScope($deptId, $user);
        $rows = DeptService::deleteDeptById($deptId);
        return $rows > 0 ? AjaxResult::success() : AjaxResult::error();
    }

    /** POST /system/dept/checkDeptNameUnique：裸 boolean（无 #[Perm]，仅登录态） */
    public function checkDeptNameUnique(Request $request): Response
    {
        $unique = DeptService::checkDeptNameUnique(
            (string)$request->post('deptName', ''),
            (int)$request->post('parentId', 0),
            (int)$request->post('deptId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** GET /system/dept/selectDeptTree/{deptId}[/{excludeId}]：树选择弹窗 */
    #[Perm('system:dept:list')]
    public function selectDeptTree(Request $request, int $deptId, int $excludeId = 0): Response
    {
        $dept = DeptService::selectDeptById($deptId) ?: [];
        return Response::create('dept/tree', 'view')->assign([
            'dept'      => $dept,
            'excludeId' => $excludeId,
        ]);
    }

    /** GET /system/dept/treeData/{excludeId}：Ztree 裸数组 */
    #[Perm('system:dept:list')]
    public function treeData(Request $request, int $excludeId): Response
    {
        $user = $request->middleware('session') ?? [];
        return \think\Response::create(DeptService::selectDeptTreeData($excludeId, $user), 'json');
    }

    /** 部门表单输入收敛（下划线键；长度上限对位 DB 列） */
    private function deptInput(Request $request): array
    {
        $deptName = trim((string)$request->post('deptName', ''));
        if ($deptName === '') {
            throw new \BusinessException('部门名称不能为空');
        }
        if (mb_strlen($deptName) > 30) {
            throw new \BusinessException('部门名称长度不能超过30个字符');
        }
        if ($request->post('orderNum') === null || $request->post('orderNum') === '') {
            throw new \BusinessException('显示顺序不能为空');
        }
        $phone = (string)$request->post('phone', '');
        if (mb_strlen($phone) > 11) {
            throw new \BusinessException('联系电话长度不能超过11个字符');
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
        return [
            'dept_id'   => (int)$request->post('deptId', 0),
            'parent_id' => (int)$request->post('parentId', 0),
            'dept_name' => $deptName,
            'order_num' => (int)$request->post('orderNum', 0),
            'leader'    => (string)$request->post('leader', ''),
            'phone'     => $phone,
            'email'     => $email,
            'status'    => (string)$request->post('status', '0'),
            'ancestors' => (string)$request->post('ancestors', ''),
        ];
    }
}
