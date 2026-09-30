<?php
declare(strict_types=1);

namespace app\controller;

use app\service\ConfigService;
use app\service\MenuService;
use app\service\SessionService;
use think\Request;
use think\Response;

/**
 * 主框架控制器（对位经典版 SysIndexController）
 */
class IndexController extends \app\BaseController
{
    /** GET /index（menuStyle 分流 index / index_topnav；移动 UA 强制 index） */
    public function index(Request $request): Response
    {
        $session = $request->middleware('session');
        $user = $session ?: [];

        $menus = MenuService::menusOf($user);
        $menuStyle = ConfigService::get('sys.index.menuStyle', 'default');
        $isMobile = (bool)preg_match('/(android|iphone|ipad|mobile)/i', (string)$request->header('user-agent', ''));
        $indexStyle = $isMobile ? 'index' : $menuStyle;

        $data = [
            'menus'     => $menus,
            'user'      => $user,
            'sideTheme' => ConfigService::get('sys.index.sideTheme', 'theme-dark'),
            'skinName'  => ConfigService::get('sys.index.skinName', 'skin-blue'),
            'footer'    => ConfigService::getBool('sys.index.footer', true),
            'tagsView'  => ConfigService::getBool('sys.index.tagsView', true),
            'version'   => (string)config('tp.version'),
            'isMobile'  => $isMobile,
            'lockscreen' => !empty($user['lockscreen']),
        ];

        $template = strtolower($indexStyle) === 'topnav' ? 'index/index_topnav' : 'index/index';
        return Response::create($template, 'view')->assign($data);
    }

    /** GET /system/main（iframe 内容页；密码策略提醒变量对位经典版 main.html） */
    public function main(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];

        // 对位 SysIndexController.initPasswordIsModify / passwordIsExpiration
        $initPwdModify = ConfigService::get('sys.account.initPasswordModify', '0') === '1'
            && empty($session['pwdUpdateDate']);
        $validateDays = (int)ConfigService::get('sys.account.passwordValidateDays', '0');
        $pwdExpired = false;
        if ($validateDays > 0) {
            if (empty($session['pwdUpdateDate'])) {
                $pwdExpired = true;
            } else {
                $days = (int)((time() - strtotime((string)$session['pwdUpdateDate'])) / 86400);
                $pwdExpired = $days > $validateDays;
            }
        }

        return Response::create('main', 'view')->assign([
            'version'          => (string)config('tp.version'),
            'initPasswordModify' => $initPwdModify,
            'passwordExpired'  => $pwdExpired,
            'user'             => $session,
        ]);
    }

    /** GET /system/switchSkin（皮肤切换页） */
    public function switchSkin(Request $request): Response
    {
        return Response::create('system/skin', 'view')->assign([
            'skinName' => ConfigService::get('sys.index.skinName', 'skin-blue'),
        ]);
    }

    /** GET /system/menuStyle/{style}（写 nav-style cookie，无页面） */
    public function menuStyle(Request $request, string $style): Response
    {
        $allow = ['default', 'topnav'];
        $style = in_array($style, $allow, true) ? $style : 'default';
        return Response::create()->code(302)->header(['Location' => '/index'])
            ->cookie('nav-style', $style, 0);
    }

    /** GET /lockscreen：锁屏页（会话写 lockscreen=true 软锁——不销毁会话，对位经典版 session attribute） */
    public function lockscreen(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $uuid = (string)($request->middleware('session_uuid') ?? '');
        if ($uuid !== '') {
            $session['lockscreen'] = true;
            \app\service\SessionService::write($uuid, $session);
        }
        return Response::create('lock/index', 'view')->assign([
            'lockUserName' => (string)($session['loginName'] ?? ''),
            'lockRealName' => (string)($session['userName'] ?? ''),
            'lockAvatar'   => (string)($session['avatar'] ?? ''),
        ]);
    }

    /** POST /unlockscreen：密码匹配 → 移除 lockscreen 字段（不计 pwd_retry 失败数——经典版 matches 直比实锤） */
    public function unlockscreen(Request $request): Response
    {
        $session = $request->middleware('session') ?? [];
        $uuid = (string)($request->middleware('session_uuid') ?? '');
        if (empty($session)) {
            return \AjaxResult::error('服务器超时，请重新登录');
        }
        $user = \app\service\UserService::selectUserById((int)($session['user_id'] ?? 0));
        $password = (string)$request->post('password', '');
        if ($user !== null && \PasswordService::verify((string)$user['login_name'], $password, (string)($user['salt'] ?? ''), (string)$user['password'])) {
            unset($session['lockscreen']);
            \app\service\SessionService::write($uuid, $session);
            return \AjaxResult::success();
        }
        return \AjaxResult::error('密码不正确，请重新输入。');
    }

    /** GET /unauth（无权限页） */
    public function unauth(): Response
    {
        return Response::create('error/unauth', 'view');
    }
}
