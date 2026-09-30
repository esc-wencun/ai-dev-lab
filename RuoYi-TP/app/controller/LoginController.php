<?php
declare(strict_types=1);

namespace app\controller;

use app\service\CaptchaService;
use app\service\ConfigService;
use app\service\LoginService;
use app\service\SessionService;
use RedisCache;
use TpConstant;
use AjaxResult;
use think\Request;
use think\Response;

/**
 * 登录控制器（对位经典版 SysLoginController + SysCaptchaController）
 */
class LoginController extends \app\BaseController
{
    /** GET /login（对位 login GET；ajax 场景由 LoginAuth 中间件拦截，不达此处） */
    public function index(Request $request): Response
    {
        $uuid = CaptchaService::anonUuid();
        $anon = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid) ?? [];
        if (!is_array($anon)) {
            $anon = [];
        }

        return Response::create('login/index', 'view')->assign([
            'captchaEnabled'   => (bool)config('tp.captcha.enabled'),
            'captchaType'      => (string)config('tp.captcha.type', 'math'),
            'isRemembered'     => (bool)config('tp.rememberMe'),
            'isAllowRegister'  => ConfigService::getBool('sys.account.registerUser', false),
            // 匿名会话 cookie 在渲染页面的响应上落地（captcha 图片与提交都依赖它）
            'anonUuid'         => $uuid,
        ])->cookie(SessionService::COOKIE_NAME, $uuid, 0);
    }

    /** POST /login（对位 ajaxLogin；表单字段与 login.js 一致） */
    public function doLogin(Request $request): Response
    {
        $username = trim((string)$request->post('username', ''));
        $password = (string)$request->post('password', '');
        $validateCode = (string)$request->post('validateCode', '');
        $rememberMe = filter_var($request->post('rememberMe', 'false'), FILTER_VALIDATE_BOOL);

        if ($username === '' || $password === '') {
            return \AjaxResult::error('用户不存在/密码错误');
        }

        $uuid = $request->cookie(SessionService::COOKIE_NAME);
        try {
            LoginService::login($username, $password, $validateCode, $rememberMe, is_string($uuid) ? $uuid : null);
        } catch (\BusinessException $e) {
            return \AjaxResult::error($e->getMessage());
        }
        return \AjaxResult::success('登录成功');
    }

    /** GET /captcha/captchaImage?type=math&s=<rand>（匿名） */
    public function captchaImage(Request $request): Response
    {
        $uuid = CaptchaService::anonUuid();
        $anon = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid);
        if (!is_array($anon)) {
            $anon = [];
        }

        $made = CaptchaService::make($anon);
        SessionService::write($uuid, $anon);

        return Response::create($made['image'])
            ->header([
                'Content-Type'  => 'image/jpeg',
                'Cache-Control' => 'no-store, no-cache, must-revalidate',
                'Pragma'        => 'no-cache',
            ])
            ->cookie(SessionService::COOKIE_NAME, $uuid, 0);
    }

    /** GET /logout（对位 LogoutFilter：清会话 + 跳 /login） */
    public function logout(Request $request): Response
    {
        $uuid = $request->cookie(SessionService::COOKIE_NAME);
        if (is_string($uuid) && $uuid !== '') {
            SessionService::destroy($uuid);
        }
        return Response::create()->code(302)->header(['Location' => '/login'])
            ->cookie(SessionService::COOKIE_NAME, '', 0);
    }
}
