<?php
declare(strict_types=1);

namespace app\middleware;

use app\service\SessionService;
use Closure;
use think\Request;
use think\Response;

/**
 * 登录态中间件（对位经典版 Shiro 主链 user filter）
 *
 * - 匿名路径放行（对位 filterChainDefinitionMap 的 anon 项）
 * - 未登录分流：ajax → {"code":"1","msg":"未登录或登录超时。请重新登录"}（code 为字符串，经典版特供）；
 *   页面 → 302 /login
 * - 命中会话即续期（30 分钟空闲超时）
 */
class LoginAuth
{
    /** 匿名路径（对位经典版 anon 链——/index /system/main 受 user 保护，不匿名） */
    private const ANON_PATHS = [
        '/login', '/logout', '/register',
        '/captcha/captchaImage',
        '/unauth',
    ];

    public function handle(Request $request, Closure $next): Response
    {
        $path = '/' . ltrim($request->pathinfo(), '/');

        foreach (self::ANON_PATHS as $anon) {
            if ($path === $anon || str_starts_with($path, $anon . '/') || $anon === '/') {
                return $next($request);
            }
        }
        // 静态资源（public 下真实文件由 web server 直接服务，这里兜底非 html 资源后缀）
        if (preg_match('#\.(css|js|png|jpg|jpeg|gif|ico|woff|woff2|ttf|svg|map|html)$#i', $path)) {
            return $next($request);
        }

        $uuid = $request->cookie(SessionService::COOKIE_NAME);
        $session = SessionService::touch($uuid);
        if ($session === null) {
            return self::unauthenticated($request);
        }

        [$uuid, $sessionData] = $session;
        // 在线用户 lastAccessTime 刷新（9.0.0；有 loginName 的会话才写，匿名会话不动）
        if (isset($sessionData['loginName']) && $sessionData['lastAccessTime'] !== date('Y-m-d H:i:s')) {
            $sessionData['lastAccessTime'] = date('Y-m-d H:i:s');
            SessionService::write($uuid, $sessionData);
        }
        $request->withMiddleware([
            'session_uuid' => $uuid,
            'session'      => $sessionData,
        ]);
        return $next($request);
    }

    public static function isAjax(Request $request): bool
    {
        // 对位经典版 ServletUtils.isAjaxRequest
        $accept = (string)$request->header('accept', '');
        if (str_contains($accept, 'application/json')) {
            return true;
        }
        return str_contains((string)$request->header('x-requested-with', ''), 'XMLHttpRequest');
    }

    public static function unauthenticated(Request $request): Response
    {
        if (self::isAjax($request)) {
            // code 是字符串 "1"，对位经典版登录页特供响应，勿改数字
            return \think\Response::create(
                ['code' => '1', 'msg' => '未登录或登录超时。请重新登录'],
                'json'
            );
        }
        return Response::create()->code(302)->header(['Location' => '/login']);
    }
}
