<?php
declare(strict_types=1);

namespace app\controller\system;

use app\attribute\Log;
use app\attribute\Perm;
use app\service\MenuService;
use app\service\PermissionService;
use AjaxResult;
use think\Request;
use think\Response;

/**
 * 菜单管理控制器（对位经典版 SysMenuController 13 方法）
 */
class MenuController extends \app\BaseController
{
    /** GET /system/menu */
    #[Perm('system:menu:view')]
    public function index(Request $request): Response
    {
        return Response::create('menu/index', 'view')->assign([
            'datas' => \app\service\DictService::listByType('sys_show_hide'),
        ]);
    }

    /** POST /system/menu/list：**裸数组**（tree-table 特例）；非 admin 按角色关联过滤 */
    #[Perm('system:menu:list')]
    public function list(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $rows = MenuService::selectMenuList([
            'menuName' => trim((string)$request->post('menuName', '')),
            'visible'  => (string)$request->post('visible', ''),
        ], $session);
        return Response::create($rows, 'json');
    }

    /** POST /system/menu/remove/{menuId}：物理删（双 warn + or parent_id 连带删 quirk） */
    #[Perm('system:menu:remove')]
    #[Log('菜单管理', Log::DELETE)]
    public function remove(Request $request, int $menuId): Response
    {
        if (MenuService::selectCountMenuByParentId($menuId) > 0) {
            return AjaxResult::warn('存在子菜单,不允许删除');
        }
        if (MenuService::selectCountRoleMenuByMenuId($menuId) > 0) {
            return AjaxResult::warn('菜单已分配,不允许删除');
        }
        $rows = MenuService::deleteMenuById($menuId);
        return $rows > 0 ? AjaxResult::success() : AjaxResult::error();
    }

    /** GET /system/menu/add/{parentId}：parentId=0 构造 {menuId:0, menuName:'主目录'} */
    #[Perm('system:menu:add')]
    public function add(Request $request, int $parentId): Response
    {
        $menu = $parentId !== 0 ? MenuService::selectMenuById($parentId) : null;
        if ($menu === null) {
            $menu = ['menu_id' => 0, 'menu_name' => '主目录'];
        }
        return Response::create('menu/add', 'view')->assign([
            'menu' => $menu,
            'datas' => \app\service\DictService::listByType('sys_show_hide'),
        ]);
    }

    /** POST /system/menu/add */
    #[Perm('system:menu:add')]
    #[Log('菜单管理', Log::INSERT)]
    public function addSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->menuInput($request);
        if (!MenuService::checkMenuNameUnique($input['menu_name'], (int)$input['parent_id'])) {
            return AjaxResult::error("新增菜单'{$input['menu_name']}'失败，菜单名称已存在");
        }
        MenuService::insertMenu($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** GET /system/menu/edit/{menuId} */
    #[Perm('system:menu:edit')]
    public function edit(Request $request, int $menuId): Response
    {
        $menu = MenuService::selectMenuById($menuId);
        if ($menu === null) {
            throw new \BusinessException('菜单不存在');
        }
        return Response::create('menu/edit', 'view')->assign([
            'menu' => $menu,
            'datas' => \app\service\DictService::listByType('sys_show_hide'),
        ]);
    }

    /** POST /system/menu/edit */
    #[Perm('system:menu:edit')]
    #[Log('菜单管理', Log::UPDATE)]
    public function editSave(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $input = $this->menuInput($request);
        $input['menu_id'] = (int)$request->post('menuId', 0);
        if (!MenuService::checkMenuNameUnique($input['menu_name'], (int)$input['parent_id'], $input['menu_id'])) {
            return AjaxResult::error("修改菜单'{$input['menu_name']}'失败，菜单名称已存在");
        }
        MenuService::updateMenu($input, (string)($session['loginName'] ?? ''));
        return AjaxResult::success();
    }

    /** POST /system/menu/updateSort：批量排序（事务） */
    #[Perm('system:menu:edit')]
    #[Log('保存菜单排序', Log::UPDATE)]
    public function updateSort(Request $request): Response
    {
        $menuIds = explode(',', (string)$request->post('menuIds', ''));
        $orderNums = explode(',', (string)$request->post('orderNums', ''));
        if (count($menuIds) !== count($orderNums) || $menuIds === ['']) {
            return AjaxResult::error('参数错误');
        }
        MenuService::updateMenuSort($menuIds, $orderNums);
        return AjaxResult::success();
    }

    /** GET /system/menu/icon：图标选择片段（无 #[Perm]，被 add/edit include） */
    public function icon(Request $request): Response
    {
        return Response::create('menu/icon', 'view');
    }

    /** POST /system/menu/checkMenuNameUnique：裸 boolean（同父下唯一） */
    public function checkMenuNameUnique(Request $request): Response
    {
        $unique = MenuService::checkMenuNameUnique(
            (string)$request->post('menuName', ''),
            (int)$request->post('parentId', 0),
            (int)$request->post('menuId', 0)
        );
        return \think\Response::create($unique ? 'true' : 'false');
    }

    /** GET /system/menu/roleMenuTreeData：Ztree（checked=concat(menu_id,perms) 整串；name 带 perms 灰字） */
    public function roleMenuTreeData(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $roleId = $request->get('roleId');
        return \think\Response::create(
            MenuService::roleMenuTreeData($roleId !== null ? (int)$roleId : null, $session),
            'json'
        );
    }

    /** GET /system/menu/menuTreeData：Ztree（无 perms 后缀、全不勾） */
    public function menuTreeData(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        return \think\Response::create(MenuService::menuTreeData($session), 'json');
    }

    /** GET /system/menu/selectMenuTree/{menuId}：菜单树选择弹窗 */
    public function selectMenuTree(Request $request, int $menuId): Response
    {
        $menu = MenuService::selectMenuById($menuId) ?: [];
        return Response::create('menu/tree', 'view')->assign(['menu' => $menu]);
    }

    /** 菜单表单输入收敛（文案对位 @Validated） */
    private function menuInput(Request $request): array
    {
        $menuName = trim((string)$request->post('menuName', ''));
        if ($menuName === '') {
            throw new \BusinessException('菜单名称不能为空');
        }
        if (mb_strlen($menuName) > 50) {
            throw new \BusinessException('菜单名称长度不能超过50个字符');
        }
        if ($request->post('orderNum') === null || $request->post('orderNum') === '') {
            throw new \BusinessException('显示顺序不能为空');
        }
        $url = (string)$request->post('url', '');
        if (mb_strlen($url) > 200) {
            throw new \BusinessException('请求地址不能超过200个字符');
        }
        $menuType = (string)$request->post('menuType', '');
        if ($menuType === '') {
            throw new \BusinessException('菜单类型不能为空');
        }
        $perms = (string)$request->post('perms', '');
        if (mb_strlen($perms) > 100) {
            throw new \BusinessException('权限标识长度不能超过100个字符');
        }
        return [
            'parent_id'  => (int)$request->post('parentId', 0),
            'menu_type'  => $menuType,
            'menu_name'  => $menuName,
            'order_num'  => (int)$request->post('orderNum', 0),
            'url'        => $url,
            'target'     => (string)$request->post('target', ''),
            'perms'      => $perms,
            'visible'    => (string)$request->post('visible', '0'),
            'is_refresh' => (string)$request->post('isRefresh', '1'),
            'icon'       => (string)$request->post('icon', '#'),
        ];
    }
}
