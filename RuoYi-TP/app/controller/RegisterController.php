<?php
declare(strict_types=1);

namespace app\controller;

use app\service\CaptchaService;
use app\service\ConfigService;
use app\service\RegisterService;
use app\service\SessionService;
use RedisCache;
use TpConstant;
use AjaxResult;
use think\Request;
use think\Response;

/**
 * 注册控制器（对位经典版 SysRegisterController；GET/POST /register 均匿名——LoginAuth ANON_PATHS 已含）
 */
class RegisterController extends \app\BaseController
{
    /** GET /register：注册页（匿名会话 + 验证码变量；不查注册开关——经典版原样，开关只拦 POST） */
    public function index(Request $request): Response
    {
        $uuid = CaptchaService::anonUuid();
        $anon = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid) ?? [];
        if (!is_array($anon)) {
            $anon = [];
        }

        return Response::create('register/index', 'view')->assign([
            'captchaEnabled' => (bool)config('tp.captcha.enabled'),
            'captchaType'    => (string)config('tp.captcha.type', 'math'),
            'anonUuid'       => $uuid,
        ])->cookie(SessionService::COOKIE_NAME, $uuid, 0);
    }

    /** POST /register：开关 → 验证码 → registerService 校验链 */
    public function doRegister(Request $request): Response
    {
        // 开关检查在验证码之前（对位经典版 controller 先查开关、filter 后校验码的顺序）
        if (ConfigService::get('sys.account.registerUser', 'false') !== 'true') {
            return AjaxResult::error('当前系统没有开启注册功能！');
        }

        $loginName = trim((string)$request->post('loginName', ''));
        $password = (string)$request->post('password', '');
        $validateCode = (string)$request->post('validateCode', '');

        // 验证码（enabled 时；答案用后即删——照 LoginService 模式）
        if (config('tp.captcha.enabled')) {
            $uuid = (string)$request->cookie(SessionService::COOKIE_NAME);
            $anon = RedisCache::get(TpConstant::PREFIX_SESSION . $uuid) ?? [];
            $ok = is_array($anon) && CaptchaService::verify($validateCode, $anon);
            SessionService::write($uuid, $anon);
            if (!$ok) {
                return AjaxResult::error('验证码错误');
            }
        }

        $msg = RegisterService::register($loginName, $password);
        return $msg === '' ? AjaxResult::success() : AjaxResult::error($msg);
    }
}
