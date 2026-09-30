<?php
declare(strict_types=1);

namespace app\middleware;

use app\attribute\Perm;
use app\service\PermissionService;
use AjaxResult;
use Closure;
use think\Request;
use think\Response;

/**
 * 权限校验中间件（对位 @RequiresPermissions + GlobalExceptionHandler.AuthorizationException）
 *
 * 反射目标控制器方法的 #[Perm] 注解；无权限时：
 * ajax → AjaxResult::error(msg)（code 500）；页面 → 渲染 error/unauth。
 */
class CheckPerm
{
    public function handle(Request $request, Closure $next): Response
    {
        $perm = self::resolvePerm($request);
        if ($perm === null) {
            return $next($request);
        }

        $sessionData = $request->middleware('session') ?? [];
        // 多权限串（如 batchForceLogout 的 batchForceLogout|forceLogout）任一命中即放行
        $granted = false;
        foreach ($perm as $p) {
            if (PermissionService::hasPerm($sessionData, $p)) {
                $granted = true;
                break;
            }
        }
        if (!$granted) {
            return self::unauthorized($request, $perm[0]);
        }
        return $next($request);
    }

    /**
     * 从请求解析当前控制器方法的 #[Perm] 注解值数组（route 管线阶段 controller/action 已就位）。
     * 返回注解 value 列表（多权限串 IS_REPEATABLE 场景返回多个）；无注解返回 null。
     */
    public static function resolvePerm(Request $request): ?array
    {
        $controller = $request->controller();
        $action = $request->action();
        if ($controller === '' || $action === '') {
            return null;
        }
        // controller 形态 = 'system.Dept'（**已含 layer 前缀**，Request::setController 实锤）→ app\controller\system\DeptController
        $class = 'app\\controller\\' . str_replace('.', '\\', $controller) . 'Controller';
        $method = $action;
        if (!class_exists($class) || !method_exists($class, $method)) {
            return null;
        }
        $ref = new \ReflectionMethod($class, $method);
        $attrs = $ref->getAttributes(Perm::class);
        if (!$attrs) {
            return null;
        }
        return array_map(fn($a) => $a->newInstance()->value, $attrs);
    }

    public static function unauthorized(Request $request, string $perm): Response
    {
        if (LoginAuth::isAjax($request)) {
            // 对位经典版 PermissionUtils.getMsg：无权限文案
            return \AjaxResult::error('您没有操作权限，请联系管理员添加权限【' . $perm . '】');
        }
        return \think\Response::create()->code(302)->header(['Location' => '/unauth']);
    }
}
