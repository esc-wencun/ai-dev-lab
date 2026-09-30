<?php
// 应用公共文件

/**
 * 模板函数：按钮级权限（对位经典版 shiro:hasPermission 标签）
 * 用法：{if check_perm('system:user:add')} ... {/if}
 */
function check_perm(string $perm): bool
{
    $request = \think\facade\Request::instance();
    $sessionData = $request->middleware('session');
    if (!is_array($sessionData)) {
        return false; // 未登录渲染场景返回 false，不抛异常
    }
    return \app\service\PermissionService::hasPerm($sessionData, $perm);
}
